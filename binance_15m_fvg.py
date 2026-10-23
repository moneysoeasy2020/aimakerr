import time
import logging
import sys
import ccxt
import pandas as pd
import numpy as np

# ----------------- 1. 日志与基础配置 -----------------
handler = logging.StreamHandler(sys.stdout)
handler.setFormatter(logging.Formatter('%(asctime)s - [%(levelname)s] - %(message)s'))

logger = logging.getLogger("Binance_15M_HighCertainty_Trailing")
logger.setLevel(logging.INFO)
logger.addHandler(handler)

API_KEY = "tj49IrEjBmedtwK3BEJNCqP1lk7fXgNkpmi5Yr2eNUhEXHG35XZzpow5TYu3S69y"
SECRET_KEY = "90ibaaGFYMzE6tE9IDqqnGEQZh0qZ76xReZQFt1AHteCRNEs1T76AGh6MDM3Lkno"

RISK_CONFIG = {
    "LEVERAGE": 10,                 # 10x 杠杆
    "MARGIN_PER_TRADE": 1.5,        # 单币固定 1.5 美金本金
    "MAX_POSITIONS": 2,             # 同时持仓最多 2 个币种
    "TOP_N_SYMBOLS": 30,            # 扫描前 30 个高流动性币种
    "ATR_PERIOD": 14,               # ATR 周期
    "MAX_RR_RATIO": 3.0,            # 移动止盈上限 1:3
    "VOL_SURGE_RATIO": 2.0,         # 突破量能提升至 2.0 倍（确保主力资金注入）
    "MAX_RANGE_PCT": 0.018,         # 75 分钟盘整波幅上限 1.8%（筛选极致压缩）
    "MIN_BODY_RATIO": 0.65,         # K 线实体占比提升至 65%
    "MAX_SHADOW_RATIO": 0.20,       # 突破方向影线占比不超过 20%
    "MIN_SL_PCT": 0.012,            # 15m 绝对最小止损垫 1.2%
    "ATR_SL_MULT": 2.2,             # 初始 ATR 止损倍数 2.2
    
    # 移动止盈追随核心参数
    "STAGE1_TRIGGER_R": 1.0,        # 浮盈达 1.0R 时，触发保本止损 (Breakeven)
    "STAGE2_TRIGGER_R": 1.5,        # 浮盈达 1.5R 时，锁住 0.8R 利润
    "TRAILING_ATR_MULT": 1.2,       # 2.0R 以后，按 1.2x ATR 追踪最高/最低价
}

exchange = ccxt.binance({
    'apiKey': API_KEY,
    'secret': SECRET_KEY,
    'timeout': 15000,
    'enableRateLimit': True,
    'options': {
        'defaultType': 'future',
        'adjustForTimeDifference': True,
    }
})

# 本地持仓元数据追踪（用于移动止盈）
pos_tracker = {}

# ----------------- 2. 辅助与指标计算 -----------------
def get_top_symbols(limit=30):
    try:
        tickers = exchange.fetch_tickers()
        usdt_pairs = []
        for symbol, ticker in tickers.items():
            if symbol.endswith('/USDT:USDT') and ticker.get('quoteVolume'):
                usdt_pairs.append((symbol, ticker['quoteVolume']))
        usdt_pairs.sort(key=lambda x: x[1], reverse=True)
        return [item[0] for item in usdt_pairs[:limit]]
    except Exception as e:
        logger.error(f"获取流动性榜单失败: {e}")
        return []

def get_active_positions():
    try:
        positions = exchange.fetch_positions()
        active = {}
        for p in positions:
            amt = float(p.get('contracts', 0) or p.get('positionAmt', 0))
            if abs(amt) > 0:
                symbol = p['symbol']
                active[symbol] = {
                    'side': 'BUY' if amt > 0 else 'SELL',
                    'amount': abs(amt),
                    'entry_price': float(p['entryPrice']),
                    'unrealized_pnl': float(p.get('unrealizedProfit', 0))
                }
        return active
    except Exception as e:
        logger.error(f"获取持仓状态失败: {e}")
        return {}

def calculate_rsi(series, period=14):
    delta = series.diff()
    gain = (delta.where(delta > 0, 0)).rolling(window=period).mean()
    loss = (-delta.where(delta < 0, 0)).rolling(window=period).mean()
    rs = gain / (loss + 1e-9)
    return 100 - (100 / (1 + rs))

def fetch_data(symbol):
    """获取 15m 主图及 4h K线，增加动能与波动率指标"""
    ohlcv_15m = exchange.fetch_ohlcv(symbol, timeframe='15m', limit=40)
    df = pd.DataFrame(ohlcv_15m, columns=['timestamp', 'open', 'high', 'low', 'close', 'volume'])
    
    df['ma5'] = df['close'].rolling(window=5).mean()
    df['ma10'] = df['close'].rolling(window=10).mean()
    
    high_low = df['high'] - df['low']
    high_close = (df['high'] - df['close'].shift()).abs()
    low_close = (df['low'] - df['close'].shift()).abs()
    tr = pd.concat([high_low, high_close, low_close], axis=1).max(axis=1)
    df['atr'] = tr.rolling(window=RISK_CONFIG["ATR_PERIOD"]).mean()
    df['rsi'] = calculate_rsi(df['close'], 14)

    # 4h 大趋势数据 (EMA50)
    ohlcv_4h = exchange.fetch_ohlcv(symbol, timeframe='4h', limit=60)
    df_4h = pd.DataFrame(ohlcv_4h, columns=['timestamp', 'open', 'high', 'low', 'close', 'volume'])
    ema50_4h = df_4h['close'].ewm(span=50, adjust=False).mean().iloc[-1]
    last_4h_close = df_4h['close'].iloc[-1]

    trend_4h = 'BULL' if last_4h_close > ema50_4h else 'BEAR'

    return df, trend_4h

# ----------------- 3. 高确定性策略逻辑 -----------------
def check_fvg(df):
    bullish_fvg = df['low'].iloc[-2] > df['high'].iloc[-4]
    bearish_fvg = df['high'].iloc[-2] < df['low'].iloc[-4]
    return bullish_fvg, bearish_fvg

def check_signal(df, trend_4h):
    if len(df) < 25:
        return None

    c = df.iloc[-2]                 # 已确立完结的突破 K 线
    consolidation = df.iloc[-7:-2] # 突破前 5 根 15m K 线 (75分钟盘整区)
    atr = df['atr'].iloc[-2]

    # 1. 极致盘整波幅过滤 (< 1.8%)
    high_max = consolidation['high'].max()
    low_min = consolidation['low'].min()
    if ((high_max - low_min) / low_min) > RISK_CONFIG["MAX_RANGE_PCT"]:
        return None

    # 2. 确定性有效突破过滤：收盘价必须有效超越盘整区边际 0.2 * ATR，防止假穿刺
    breakout_buffer = 0.2 * atr
    
    # 3. K 线实体与影线质量过滤
    k_range = c['high'] - c['low']
    if k_range == 0:
        return None

    body_size = abs(c['close'] - c['open'])
    if (body_size / k_range) < RISK_CONFIG["MIN_BODY_RATIO"]:
        return None

    upper_shadow = c['high'] - max(c['close'], c['open'])
    lower_shadow = min(c['close'], c['open']) - c['low']

    # 4. 主力资金量能爆发 (>= 2.0x 均值)
    if c['volume'] < (consolidation['volume'].mean() * RISK_CONFIG["VOL_SURGE_RATIO"]):
        return None

    rsi_val = c['rsi']
    bull_fvg, bear_fvg = check_fvg(df)

    # 5. 多头信号（4h 强看多 + 有效超越高点 + 无长上影 + RSI 黄金区间 + FVG）
    is_ma_bull = (c['ma5'] > c['ma10']) and (c['close'] > c['ma5'])
    clean_break_high = c['close'] > (high_max + breakout_buffer)
    no_upper_spike = (upper_shadow / k_range) <= RISK_CONFIG["MAX_SHADOW_RATIO"]

    if (trend_4h == 'BULL' and is_ma_bull and clean_break_high and bull_fvg and 
        no_upper_spike and (48 < rsi_val < 68)):
        return 'BUY'

    # 6. 空头信号（4h 强看空 + 有效跌破低点 + 无长下影 + RSI 黄金区间 + FVG）
    is_ma_bear = (c['ma5'] < c['ma10']) and (c['close'] < c['ma10'])
    clean_break_low = c['close'] < (low_min - breakout_buffer)
    no_lower_spike = (lower_shadow / k_range) <= RISK_CONFIG["MAX_SHADOW_RATIO"]

    if (trend_4h == 'BEAR' and is_ma_bear and clean_break_low and bear_fvg and 
        no_lower_spike and (32 < rsi_val < 52)):
        return 'SELL'

    return None

# ----------------- 4. 阶梯式移动止盈/风控模块 -----------------
def update_trailing_stops(active_positions):
    """根据实时价格推移止损单，实现阶梯保本与动态追踪"""
    for symbol, info in list(active_positions.items()):
        if symbol not in pos_tracker:
            continue

        meta = pos_tracker[symbol]
        df, _ = fetch_data(symbol)
        last_price = df['close'].iloc[-1]
        atr = df['atr'].iloc[-2]

        side = info['side']
        entry_price = info['entry_price']
        init_sl_dist = meta['init_sl_dist']
        current_sl = meta['current_sl']

        trailing_step = atr * RISK_CONFIG["TRAILING_ATR_MULT"]

        if side == 'BUY':
            profit_dist = last_price - entry_price
            
            # 阶段 3：浮盈 >= 2.0R，开启 ATR 动态追踪最高价
            if profit_dist >= (init_sl_dist * 2.0):
                new_sl = max(entry_price + (init_sl_dist * 0.8), last_price - trailing_step)
                new_sl = float(exchange.price_to_precision(symbol, new_sl))
                if new_sl > current_sl:
                    logger.info(f"🔥 [{symbol}] [多单 Stage-3 追踪止盈] 止损抬升: {current_sl} -> {new_sl} (现价: {last_price})")
                    _update_stop_order(symbol, 'sell', info['amount'], new_sl)

            # 阶段 2：1.5R <= 浮盈 < 2.0R，锁定 0.8R 利润
            elif profit_dist >= (init_sl_dist * RISK_CONFIG["STAGE2_TRIGGER_R"]):
                new_sl = float(exchange.price_to_precision(symbol, entry_price + (init_sl_dist * 0.8)))
                if new_sl > current_sl:
                    logger.info(f"🛡️ [{symbol}] [多单 Stage-2 锁利] 止损抬升: {current_sl} -> {new_sl} (锁利 0.8R)")
                    _update_stop_order(symbol, 'sell', info['amount'], new_sl)

            # 阶段 1：1.0R <= 浮盈 < 1.5R，触发保本 (Breakeven)
            elif profit_dist >= (init_sl_dist * RISK_CONFIG["STAGE1_TRIGGER_R"]):
                new_sl = float(exchange.price_to_precision(symbol, entry_price * 1.001)) # 微幅盖过手续费
                if new_sl > current_sl:
                    logger.info(f"🔒 [{symbol}] [多单 Stage-1 保本] 止损抬升: {current_sl} -> {new_sl} (成本保本)")
                    _update_stop_order(symbol, 'sell', info['amount'], new_sl)

        elif side == 'SELL':
            profit_dist = entry_price - last_price
            
            # 阶段 3：浮盈 >= 2.0R，开启 ATR 动态追踪最低价
            if profit_dist >= (init_sl_dist * 2.0):
                new_sl = min(entry_price - (init_sl_dist * 0.8), last_price + trailing_step)
                new_sl = float(exchange.price_to_precision(symbol, new_sl))
                if new_sl < current_sl:
                    logger.info(f"🔥 [{symbol}] [空单 Stage-3 追踪止盈] 止损下移: {current_sl} -> {new_sl} (现价: {last_price})")
                    _update_stop_order(symbol, 'buy', info['amount'], new_sl)

            # 阶段 2：1.5R <= 浮盈 < 2.0R，锁定 0.8R 利润
            elif profit_dist >= (init_sl_dist * RISK_CONFIG["STAGE2_TRIGGER_R"]):
                new_sl = float(exchange.price_to_precision(symbol, entry_price - (init_sl_dist * 0.8)))
                if new_sl < current_sl:
                    logger.info(f"🛡️ [{symbol}] [空单 Stage-2 锁利] 止损下移: {current_sl} -> {new_sl} (锁利 0.8R)")
                    _update_stop_order(symbol, 'buy', info['amount'], new_sl)

            # 阶段 1：1.0R <= 浮盈 < 1.5R，触发保本 (Breakeven)
            elif profit_dist >= (init_sl_dist * RISK_CONFIG["STAGE1_TRIGGER_R"]):
                new_sl = float(exchange.price_to_precision(symbol, entry_price * 0.999))
                if new_sl < current_sl:
                    logger.info(f"🔒 [{symbol}] [空单 Stage-1 保本] 止损下移: {current_sl} -> {new_sl} (成本保本)")
                    _update_stop_order(symbol, 'buy', info['amount'], new_sl)

def _update_stop_order(symbol, exit_side, amount, new_sl):
    try:
        exchange.cancel_all_orders(symbol)
        exchange.create_order(symbol, 'STOP_MARKET', exit_side, amount, params={'stopPrice': new_sl, 'reduceOnly': True})
        pos_tracker[symbol]['current_sl'] = new_sl
    except Exception as e:
        logger.error(f"[{symbol}] 修改止损条件单失败: {e}")

def execute_trade(symbol, side, df):
    try:
        exchange.set_leverage(RISK_CONFIG["LEVERAGE"], symbol)
        
        last_price = df['close'].iloc[-1]
        c_break = df.iloc[-2]
        atr = df['atr'].iloc[-2]
        
        notional = RISK_CONFIG["MARGIN_PER_TRADE"] * RISK_CONFIG["LEVERAGE"]
        exchange.load_markets()
        amount = float(exchange.amount_to_precision(symbol, notional / last_price))
        if amount <= 0:
            return

        min_sl_dist = last_price * RISK_CONFIG["MIN_SL_PCT"]
        atr_sl_dist = atr * RISK_CONFIG["ATR_SL_MULT"]

        if side == 'BUY':
            struct_sl_dist = last_price - c_break['low']
            sl_distance = max(min_sl_dist, atr_sl_dist, struct_sl_dist)
            sl_price = last_price - sl_distance
            tp_price = last_price + (sl_distance * RISK_CONFIG["MAX_RR_RATIO"])
            order_side, exit_side = 'buy', 'sell'
        else:
            struct_sl_dist = c_break['high'] - last_price
            sl_distance = max(min_sl_dist, atr_sl_dist, struct_sl_dist)
            sl_price = last_price + sl_distance
            tp_price = last_price - (sl_distance * RISK_CONFIG["MAX_RR_RATIO"])
            order_side, exit_side = 'sell', 'buy'

        sl_price = float(exchange.price_to_precision(symbol, sl_price))
        tp_price = float(exchange.price_to_precision(symbol, tp_price))

        logger.info(f"✨ [15m 高确定性开仓] [{symbol}] 方向: {side} | 现价: {last_price} | SL: {sl_price} | 移动TP上限: {tp_price}")

        order = exchange.create_order(symbol=symbol, type='market', side=order_side, amount=amount)
        logger.info(f"[{symbol}] 市价开仓成功, 单号: {order['id']}")

        # 挂初始止损单与终极止盈单
        exchange.create_order(symbol, 'STOP_MARKET', exit_side, amount, params={'stopPrice': sl_price, 'reduceOnly': True})
        exchange.create_order(symbol, 'TAKE_PROFIT_MARKET', exit_side, amount, params={'stopPrice': tp_price, 'reduceOnly': True})

        # 写入追踪字典
        pos_tracker[symbol] = {
            'init_sl_dist': sl_distance,
            'current_sl': sl_price,
            'tp_price': tp_price
        }

    except Exception as e:
        logger.error(f"[{symbol}] 开仓流程异常: {e}")

# ----------------- 5. 主程序 -----------------
def main():
    logger.info("🚀 币安 15m 高确定性突破 + 阶梯移动止盈引擎启动完成！")
    symbols = get_top_symbols(RISK_CONFIG["TOP_N_SYMBOLS"])
    
    while True:
        try:
            active_positions = get_active_positions()
            
            # 1. 实时维护移动止盈/保本
            if active_positions:
                update_trailing_stops(active_positions)

            # 2. 清理平仓残余内存
            for tracked_sym in list(pos_tracker.keys()):
                if tracked_sym not in active_positions:
                    del pos_tracker[tracked_sym]

            # 3. 扫描开仓机会
            if len(active_positions) < RISK_CONFIG["MAX_POSITIONS"]:
                for symbol in symbols:
                    if symbol in active_positions:
                        continue

                    df, trend_4h = fetch_data(symbol)
                    signal = check_signal(df, trend_4h)

                    if signal:
                        execute_trade(symbol, signal, df)
                        break

            time.sleep(10)

        except KeyboardInterrupt:
            break
        except Exception as e:
            logger.error(f"轮询异常: {e}")
            time.sleep(5)

if __name__ == '__main__':
    main()
