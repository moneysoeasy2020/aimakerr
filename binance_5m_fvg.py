import time
import logging
import sys
import ccxt
import pandas as pd
import numpy as np

# ----------------- 1. 日志与基础配置 -----------------
handler = logging.StreamHandler(sys.stdout)
handler.setFormatter(logging.Formatter('%(asctime)s - [%(levelname)s] - %(message)s'))

logger = logging.getLogger("Binance_5M_Quant_Institutional")
logger.setLevel(logging.INFO)
logger.addHandler(handler)

API_KEY = "tj49IrEjBmedtwK3BEJNCqP1lk7fXgNkpmi5Yr2eNUhEXHG35XZzpow5TYu3S69y"
SECRET_KEY = "90ibaaGFYMzE6tE9IDqqnGEQZh0qZ76xReZQFt1AHteCRNEs1T76AGh6MDM3Lkno"

RISK_CONFIG = {
    "LEVERAGE": 10,                 # 10x 杠杆
    "MARGIN_PER_TRADE": 1.5,        # 单币 1.5 美金本金
    "MAX_POSITIONS": 2,             # 最多持仓 2 个币种
    "TOP_N_SYMBOLS": 30,            # 扫描前 30 流动性币种
    "ATR_PERIOD": 14,               # ATR 周期
    "RR_RATIO": 2.2,                # 盈亏比提升至 2.2
    "VOL_SURGE_RATIO": 1.8,         # 突破量能提升至前 5 根均值的 1.8 倍
    "MAX_RANGE_PCT": 0.010,         # 盘整波幅进一步收紧至 1.0%
    "MIN_BODY_RATIO": 0.65,         # K 线实体占比提升至 65%
    "MAX_SHADOW_RATIO": 0.20,       # 突破方向影线占比严格限制在 20% 以内
    "MIN_SL_PCT": 0.008,            # 最小止损垫 0.8%
    "ATR_SL_MULT": 2.2,             # ATR 止损倍数
    "MIN_EMA_SPREAD_PCT": 0.0035,   # 均线散度阈值 (低于 0.35% 认定为无方向缠绕)
    "BREAKEVEN_R_TRIGGER": 1.0,     # 浮盈达到 1R 时强制触发保本损
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

trend_1h_cache = {}

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
        active_details = []
        for p in positions:
            amt = abs(float(p.get('contracts', 0) or p.get('positionAmt', 0)))
            if amt > 0:
                active_details.append({
                    'symbol': p['symbol'],
                    'side': 'BUY' if float(p.get('positionAmt', 0) or p.get('contracts', 0)) > 0 else 'SELL',
                    'entryPrice': float(p.get('entryPrice', 0)),
                    'amount': amt
                })
        return len(active_details), active_details
    except Exception as e:
        logger.error(f"获取持仓失败: {e}")
        return 999, []

def clean_orphan_orders(active_symbols):
    """清理无持仓币种的残余挂单"""
    try:
        open_orders = exchange.fetch_open_orders()
        for order in open_orders:
            sym = order['symbol']
            if sym not in active_symbols:
                logger.info(f"🧹 清理未持仓币种 [{sym}] 的残余挂单 ID: {order['id']}")
                exchange.cancel_order(order['id'], sym)
    except Exception as e:
        logger.error(f"清理残余挂单异常: {e}")

def calculate_rsi(series, period=14):
    delta = series.diff()
    gain = (delta.where(delta > 0, 0)).rolling(window=period).mean()
    loss = (-delta.where(delta < 0, 0)).rolling(window=period).mean()
    rs = gain / (loss + 1e-9)
    return 100 - (100 / (1 + rs))

def fetch_trend_1h(symbol):
    now = time.time()
    if symbol in trend_1h_cache and (now - trend_1h_cache[symbol]['time']) < 900:
        return trend_1h_cache[symbol]['trend']

    try:
        ohlcv_1h = exchange.fetch_ohlcv(symbol, timeframe='1h', limit=60)
        df_1h = pd.DataFrame(ohlcv_1h, columns=['timestamp', 'open', 'high', 'low', 'close', 'volume'])
        ema50_1h = df_1h['close'].ewm(span=50, adjust=False).mean().iloc[-1]
        last_1h_close = df_1h['close'].iloc[-1]

        trend = 'BULL' if last_1h_close > ema50_1h else 'BEAR'
        trend_1h_cache[symbol] = {'trend': trend, 'time': now}
        return trend
    except Exception as e:
        logger.error(f"获取 1h 趋势失败 [{symbol}]: {e}")
        return 'UNKNOWN'

def fetch_data(symbol):
    """获取 5m 数据并计算 EMA / ATR / RSI 及均线散度"""
    ohlcv_5m = exchange.fetch_ohlcv(symbol, timeframe='5m', limit=50)
    df = pd.DataFrame(ohlcv_5m, columns=['timestamp', 'open', 'high', 'low', 'close', 'volume'])
    
    df['ema20'] = df['close'].ewm(span=20, adjust=False).mean()
    df['ema50'] = df['close'].ewm(span=50, adjust=False).mean()
    df['ema99'] = df['close'].ewm(span=99, adjust=False).mean()
    
    high_low = df['high'] - df['low']
    high_close = (df['high'] - df['close'].shift()).abs()
    low_close = (df['low'] - df['close'].shift()).abs()
    tr = pd.concat([high_low, high_close, low_close], axis=1).max(axis=1)
    df['atr'] = tr.rolling(window=RISK_CONFIG["ATR_PERIOD"]).mean()
    df['rsi'] = calculate_rsi(df['close'], 14)

    trend_1h = fetch_trend_1h(symbol)
    return df, trend_1h

# ----------------- 3. 高阶信号与防震荡滤网 -----------------
def check_fvg(df):
    bullish_fvg = df['low'].iloc[-2] > df['high'].iloc[-4]
    bearish_fvg = df['high'].iloc[-2] < df['low'].iloc[-4]
    return bullish_fvg, bearish_fvg

def check_signal(df, trend_1h):
    if len(df) < 30 or trend_1h == 'UNKNOWN':
        return None

    c = df.iloc[-2]                 # 已确认收盘的突破 K 线
    consolidation = df.iloc[-7:-2] # 前 25 分钟盘整区

    # 1. 均线散度（EMA Spread）防震荡缠绕
    ema_max = max(c['ema20'], c['ema50'], c['ema99'])
    ema_min = min(c['ema20'], c['ema50'], c['ema99'])
    if ((ema_max - ema_min) / c['close']) < RISK_CONFIG["MIN_EMA_SPREAD_PCT"]:
        return None

    # 2. 盘整区间波幅过滤 (< 1.0%)
    high_max = consolidation['high'].max()
    low_min = consolidation['low'].min()
    if ((high_max - low_min) / low_min) > RISK_CONFIG["MAX_RANGE_PCT"]:
        return None

    # 3. K 线形态与极值冲刺过滤（带 ATR 缓冲区，防极值假插针）
    k_range = c['high'] - c['low']
    if k_range == 0:
        return None

    body_size = abs(c['close'] - c['open'])
    if (body_size / k_range) < RISK_CONFIG["MIN_BODY_RATIO"]:
        return None

    upper_shadow = c['high'] - max(c['close'], c['open'])
    lower_shadow = min(c['close'], c['open']) - c['low']
    atr_val = c['atr']

    # 4. 量能与 RSI 严格校验
    if c['volume'] < (consolidation['volume'].mean() * RISK_CONFIG["VOL_SURGE_RATIO"]):
        return None

    rsi_val = c['rsi']
    bull_fvg, bear_fvg = check_fvg(df)

    # 5. 多头信号（1h 看多 + 5m EMA 多头排列 + 彻底突破前高 + 无上影 + 适中 RSI）
    is_bull_k = c['close'] > c['open']
    is_ema_bull = (c['ema20'] > c['ema50']) and (c['ema50'] > c['ema99'])
    clean_break_high = c['close'] > (high_max + atr_val * 0.15)
    no_upper_spike = (upper_shadow / k_range) <= RISK_CONFIG["MAX_SHADOW_RATIO"]

    if (is_bull_k and trend_1h == 'BULL' and is_ema_bull and clean_break_high and 
        bull_fvg and no_upper_spike and (48 < rsi_val < 68)):
        return 'BUY'

    # 6. 空头信号（1h 看空 + 5m EMA 空头排列 + 彻底突破前低 + 无下影 + 适中 RSI）
    is_bear_k = c['close'] < c['open']
    is_ema_bear = (c['ema20'] < c['ema50']) and (c['ema50'] < c['ema99'])
    clean_break_low = c['close'] < (low_min - atr_val * 0.15)
    no_lower_spike = (lower_shadow / k_range) <= RISK_CONFIG["MAX_SHADOW_RATIO"]

    if (is_bear_k and trend_1h == 'BEAR' and is_ema_bear and clean_break_low and 
        bear_fvg and no_lower_spike and (32 < rsi_val < 52)):
        return 'SELL'

    return None

# ----------------- 4. 动态保本损管理 -----------------
def manage_breakeven(active_positions_details):
    """当浮盈达到 1R 时，将止损移至开仓成本价，彻底规避深套亏损"""
    for pos in active_positions_details:
        symbol = pos['symbol']
        side = pos['side']
        entry_price = pos['entryPrice']
        
        try:
            ticker = exchange.fetch_ticker(symbol)
            current_price = ticker['last']
            open_orders = exchange.fetch_open_orders(symbol)
            
            # 找到当前持仓的止损单
            sl_order = next((o for o in open_orders if o['type'] == 'STOP_MARKET'), None)
            if not sl_order:
                continue

            current_sl = float(sl_order['stopPrice'])
            
            if side == 'BUY':
                risk_dist = entry_price - current_sl
                if risk_dist <= 0: # 已经升至保本或以上
                    continue
                # 浮盈超过 1R 时触发保本
                if (current_price - entry_price) >= (risk_dist * RISK_CONFIG["BREAKEVEN_R_TRIGGER"]):
                    be_price = float(exchange.price_to_precision(symbol, entry_price * 1.001)) # 0.1% 包含手续费
                    exchange.cancel_order(sl_order['id'], symbol)
                    exchange.create_order(symbol, 'STOP_MARKET', 'sell', pos['amount'], params={'stopPrice': be_price, 'reduceOnly': True})
                    logger.info(f"🛡️ [{symbol}] 多头浮盈超 1R，触发动态保本损！止损已升至成本价: {be_price}")
            
            elif side == 'SELL':
                risk_dist = current_sl - entry_price
                if risk_dist <= 0:
                    continue
                # 浮盈超过 1R 时触发保本
                if (entry_price - current_price) >= (risk_dist * RISK_CONFIG["BREAKEVEN_R_TRIGGER"]):
                    be_price = float(exchange.price_to_precision(symbol, entry_price * 0.999))
                    exchange.cancel_order(sl_order['id'], symbol)
                    exchange.create_order(symbol, 'STOP_MARKET', 'buy', pos['amount'], params={'stopPrice': be_price, 'reduceOnly': True})
                    logger.info(f"🛡️ [{symbol}] 空头浮盈超 1R，触发动态保本损！止损已降至成本价: {be_price}")
                    
        except Exception as e:
            logger.error(f"[{symbol}] 管理保本损异常: {e}")

# ----------------- 5. 开仓执行模块 -----------------
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

        if sl_price <= 0:
            return

        sl_price = float(exchange.price_to_precision(symbol, sl_price))
        tp_price = float(exchange.price_to_precision(symbol, tp_price))

        logger.info(f"⚡ [量化精细化信号触发] [{symbol}] 方向: {side} | 现价: {last_price} | SL: {sl_price} | TP: {tp_price}")

        order = exchange.create_order(symbol=symbol, type='market', side=order_side, amount=amount)
        logger.info(f"[{symbol}] 市价开仓成功, 单号: {order['id']}")

        # 挂止损单
        try:
            exchange.create_order(symbol, 'STOP_MARKET', exit_side, amount, params={'stopPrice': sl_price, 'reduceOnly': True})
            logger.info(f"[{symbol}] 止损挂单成功 ({sl_price})")
        except Exception as e:
            logger.critical(f"[{symbol}] 止损挂单失败: {e}，市价平仓保护！")
            exchange.create_order(symbol, 'market', exit_side, amount, params={'reduceOnly': True})
            return

        # 挂止盈单
        try:
            exchange.create_order(symbol, 'TAKE_PROFIT_MARKET', exit_side, amount, params={'stopPrice': tp_price, 'reduceOnly': True})
            logger.info(f"[{symbol}] 止盈挂单成功 ({tp_price})")
        except Exception as e:
            logger.error(f"[{symbol}] 止盈挂单失败: {e}")

    except Exception as e:
        logger.error(f"[{symbol}] 开仓流程异常: {e}")

# ----------------- 6. 主程序 -----------------
def main():
    logger.info("🧠 币安 5m 博士级量化引擎启动（EMA 散度防震荡 + 1R 动态保本 + 流动性插针防御）！")
    symbols = get_top_symbols(RISK_CONFIG["TOP_N_SYMBOLS"])
    
    loop_count = 0
    while True:
        try:
            active_count, active_details = get_active_positions()
            active_symbols = [p['symbol'] for p in active_details]

            # 动态管理已持仓订单的保本损
            if active_count > 0:
                manage_breakeven(active_details)
            
            # 定期清理未持仓残余挂单
            loop_count += 1
            if loop_count % 12 == 0:
                clean_orphan_orders(active_symbols)

            if active_count >= RISK_CONFIG["MAX_POSITIONS"]:
                time.sleep(10)
                continue

            for symbol in symbols:
                if symbol in active_symbols:
                    continue

                df, trend_1h = fetch_data(symbol)
                signal = check_signal(df, trend_1h)

                if signal:
                    execute_trade(symbol, signal, df)
                    break
                
                time.sleep(0.05)

            time.sleep(10)

        except KeyboardInterrupt:
            break
        except Exception as e:
            logger.error(f"轮询主循环异常: {e}")
            time.sleep(5)

if __name__ == '__main__':
    main()
