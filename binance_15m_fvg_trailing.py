import time
import logging
import sys
import ccxt
import pandas as pd
import numpy as np

# ----------------- 1. 日志与基础配置 -----------------
handler = logging.StreamHandler(sys.stdout)
handler.setFormatter(logging.Formatter('%(asctime)s - [%(levelname)s] - %(message)s'))

logger = logging.getLogger("Binance_15M_Trailing_FVG")
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
    "RR_RATIO": 3.0,                # 结合移动止损，放宽硬止盈至 1:3
    "VOL_SURGE_RATIO": 1.8,         # 量能突破 1.8 倍
    "MAX_RANGE_PCT": 0.020,         # 75 分钟盘整波幅上限 2.0%
    "MIN_BODY_RATIO": 0.60,         # K 线实体占比 60%
    "MAX_SHADOW_RATIO": 0.25,       # 突破方向影线占比不超过 25%
    "MIN_SL_PCT": 0.012,            # 15m 绝对最小止损垫 1.2%
    "ATR_SL_MULT": 2.2,             # 初始 ATR 止损倍数
    "TRAILING_TRIGGER_R": 1.0,      # 浮盈达 1.0R 时激活移动止损
    "TRAILING_ATR_MULT": 1.2,       # 移动止损追随距离 (1.2 * ATR)
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

# 本地持仓追踪字典，用于记录移动止损状态
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
        logger.error(f"获取持仓数失败: {e}")
        return {}

def calculate_rsi(series, period=14):
    delta = series.diff()
    gain = (delta.where(delta > 0, 0)).rolling(window=period).mean()
    loss = (-delta.where(delta < 0, 0)).rolling(window=period).mean()
    rs = gain / (loss + 1e-9)
    return 100 - (100 / (1 + rs))

def fetch_data(symbol):
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

    ohlcv_4h = exchange.fetch_ohlcv(symbol, timeframe='4h', limit=60)
    df_4h = pd.DataFrame(ohlcv_4h, columns=['timestamp', 'open', 'high', 'low', 'close', 'volume'])
    ema50_4h = df_4h['close'].ewm(span=50, adjust=False).mean().iloc[-1]
    trend_4h = 'BULL' if df_4h['close'].iloc[-1] > ema50_4h else 'BEAR'

    return df, trend_4h

# ----------------- 3. 策略逻辑模块 -----------------
def check_fvg(df):
    bullish_fvg = df['low'].iloc[-2] > df['high'].iloc[-4]
    bearish_fvg = df['high'].iloc[-2] < df['low'].iloc[-4]
    return bullish_fvg, bearish_fvg

def check_signal(df, trend_4h):
    if len(df) < 25:
        return None

    c = df.iloc[-2]
    consolidation = df.iloc[-7:-2]

    high_max = consolidation['high'].max()
    low_min = consolidation['low'].min()
    if ((high_max - low_min) / low_min) > RISK_CONFIG["MAX_RANGE_PCT"]:
        return None

    k_range = c['high'] - c['low']
    if k_range == 0:
        return None

    body_size = abs(c['close'] - c['open'])
    if (body_size / k_range) < RISK_CONFIG["MIN_BODY_RATIO"]:
        return None

    upper_shadow = c['high'] - max(c['close'], c['open'])
    lower_shadow = min(c['close'], c['open']) - c['low']

    if c['volume'] < (consolidation['volume'].mean() * RISK_CONFIG["VOL_SURGE_RATIO"]):
        return None

    rsi_val = c['rsi']
    bull_fvg, bear_fvg = check_fvg(df)

    is_ma_bull = (c['ma5'] > c['ma10']) and (c['close'] > c['ma5'])
    no_upper_spike = (upper_shadow / k_range) <= RISK_CONFIG["MAX_SHADOW_RATIO"]

    if (trend_4h == 'BULL' and is_ma_bull and c['close'] > high_max and bull_fvg and 
        no_upper_spike and (45 < rsi_val < 70)):
        return 'BUY'

    is_ma_bear = (c['ma5'] < c['ma10']) and (c['close'] < c['ma10'])
    no_lower_spike = (lower_shadow / k_range) <= RISK_CONFIG["MAX_SHADOW_RATIO"]

    if (trend_4h == 'BEAR' and is_ma_bear and c['close'] < low_min and bear_fvg and 
        no_lower_spike and (30 < rsi_val < 55)):
        return 'SELL'

    return None

# ----------------- 4. 动态移动止损风控模块 -----------------
def update_trailing_stops(active_positions):
    """根据实时价格推送移动止损，上移/下移止损条件单"""
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
            # 激活条件：浮盈达 1.0R
            if profit_dist >= (init_sl_dist * RISK_CONFIG["TRAILING_TRIGGER_R"]):
                new_sl = max(entry_price, last_price - trailing_step)
                new_sl = float(exchange.price_to_precision(symbol, new_sl))
                
                # 止损位只能向上抬高
                if new_sl > current_sl:
                    logger.info(f"🔄 [{symbol}] 触发移动止盈！原止损: {current_sl} -> 新止损: {new_sl} (当前价: {last_price})")
                    try:
                        exchange.cancel_all_orders(symbol)
                        exchange.create_order(symbol, 'STOP_MARKET', 'sell', info['amount'], params={'stopPrice': new_sl, 'reduceOnly': True})
                        pos_tracker[symbol]['current_sl'] = new_sl
                    except Exception as e:
                        logger.error(f"[{symbol}] 移动止损修改失败: {e}")

        elif side == 'SELL':
            profit_dist = entry_price - last_price
            if profit_dist >= (init_sl_dist * RISK_CONFIG["TRAILING_TRIGGER_R"]):
                new_sl = min(entry_price, last_price + trailing_step)
                new_sl = float(exchange.price_to_precision(symbol, new_sl))
                
                # 止损位只能向下降低
                if new_sl < current_sl:
                    logger.info(f"🔄 [{symbol}] 触发移动止盈！原止损: {current_sl} -> 新止损: {new_sl} (当前价: {last_price})")
                    try:
                        exchange.cancel_all_orders(symbol)
                        exchange.create_order(symbol, 'STOP_MARKET', 'buy', info['amount'], params={'stopPrice': new_sl, 'reduceOnly': True})
                        pos_tracker[symbol]['current_sl'] = new_sl
                    except Exception as e:
                        logger.error(f"[{symbol}] 移动止损修改失败: {e}")

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
            tp_price = last_price + (sl_distance * RISK_CONFIG["RR_RATIO"])
            order_side, exit_side = 'buy', 'sell'
        else:
            struct_sl_dist = c_break['high'] - last_price
            sl_distance = max(min_sl_dist, atr_sl_dist, struct_sl_dist)
            sl_price = last_price + sl_distance
            tp_price = last_price - (sl_distance * RISK_CONFIG["RR_RATIO"])
            order_side, exit_side = 'sell', 'buy'

        sl_price = float(exchange.price_to_precision(symbol, sl_price))
        tp_price = float(exchange.price_to_precision(symbol, tp_price))

        logger.info(f"✨ [15m 开仓] [{symbol}] 方向: {side} | 现价: {last_price} | 初始SL: {sl_price} | TP: {tp_price}")

        order = exchange.create_order(symbol=symbol, type='market', side=order_side, amount=amount)
        logger.info(f"[{symbol}] 市价开仓成功, 单号: {order['id']}")

        # 挂初始止损单与止盈单
        exchange.create_order(symbol, 'STOP_MARKET', exit_side, amount, params={'stopPrice': sl_price, 'reduceOnly': True})
        exchange.create_order(symbol, 'TAKE_PROFIT_MARKET', exit_side, amount, params={'stopPrice': tp_price, 'reduceOnly': True})

        # 记录持仓元数据供移动止损模块识别
        pos_tracker[symbol] = {
            'init_sl_dist': sl_distance,
            'current_sl': sl_price,
            'tp_price': tp_price
        }

    except Exception as e:
        logger.error(f"[{symbol}] 开仓流程异常: {e}")

# ----------------- 5. 主程序 -----------------
def main():
    logger.info("🚀 币安 15m 趋势突破 + 移动止损追随止盈引擎启动完成！")
    symbols = get_top_symbols(RISK_CONFIG["TOP_N_SYMBOLS"])
    
    while True:
        try:
            active_positions = get_active_positions()
            
            # 1. 优先维护现有持仓的移动止损
            if active_positions:
                update_trailing_stops(active_positions)

            # 2. 清理已平仓币种的内存记录
            for tracked_sym in list(pos_tracker.keys()):
                if tracked_sym not in active_positions:
                    del pos_tracker[tracked_sym]

            # 3. 检查是否有新开仓机会
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
