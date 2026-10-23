import time
import logging
import sys
import ccxt
import pandas as pd
import numpy as np

# ----------------- 1. 日志与基础配置 -----------------
handler = logging.StreamHandler(sys.stdout)
handler.setFormatter(logging.Formatter('%(asctime)s - [%(levelname)s] - %(message)s'))

logger = logging.getLogger("Binance_3M_FVG_Ultra")
logger.setLevel(logging.INFO)
logger.addHandler(handler)

API_KEY = "tj49IrEjBmedtwK3BEJNCqP1lk7fXgNkpmi5Yr2eNUhEXHG35XZzpow5TYu3S69y"
SECRET_KEY = "90ibaaGFYMzE6tE9IDqqnGEQZh0qZ76xReZQFt1AHteCRNEs1T76AGh6MDM3Lkno"

RISK_CONFIG = {
    "LEVERAGE": 10,                 # 10x 杠杆
    "MARGIN_PER_TRADE": 1.5,        # 单币固定 1.5 美金本金
    "MAX_POSITIONS": 2,             # 同时持仓最多 2 个币种
    "TOP_N_SYMBOLS": 30,            # 扫描前 30 个高流动性币种
    "ATR_PERIOD": 14,               # ATR 14
    "RR_RATIO": 2.0,                # 1:2 盈亏比
    "VOL_SURGE_RATIO": 1.8,         # 突破量能达前 5 根 3m 均值的 1.8 倍
    "MAX_RANGE_PCT": 0.008,         # 15 分钟 (5*3m) 盘整波幅上限 0.8%
    "MIN_BODY_RATIO": 0.50,         # K 线实体占比至少 50%
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

def get_active_position_count():
    try:
        positions = exchange.fetch_positions()
        active = [
            p for p in positions 
            if abs(float(p.get('contracts', 0) or p.get('positionAmt', 0))) > 0
        ]
        return len(active), [p['symbol'] for p in active]
    except Exception as e:
        logger.error(f"获取持仓数失败: {e}")
        return 999, []

def fetch_data(symbol):
    """获取 3m 及 1h K线，支持 1h 大趋势过滤"""
    # 3m 主周期数据
    ohlcv_3m = exchange.fetch_ohlcv(symbol, timeframe='3m', limit=35)
    df = pd.DataFrame(ohlcv_3m, columns=['timestamp', 'open', 'high', 'low', 'close', 'volume'])
    
    df['ma5'] = df['close'].rolling(window=5).mean()
    df['ma10'] = df['close'].rolling(window=10).mean()
    
    high_low = df['high'] - df['low']
    high_close = (df['high'] - df['close'].shift()).abs()
    low_close = (df['low'] - df['close'].shift()).abs()
    tr = pd.concat([high_low, high_close, low_close], axis=1).max(axis=1)
    df['atr'] = tr.rolling(window=RISK_CONFIG["ATR_PERIOD"]).mean()

    # 1h 大趋势数据 (EMA50)
    ohlcv_1h = exchange.fetch_ohlcv(symbol, timeframe='1h', limit=60)
    df_1h = pd.DataFrame(ohlcv_1h, columns=['timestamp', 'open', 'high', 'low', 'close', 'volume'])
    ema50_1h = df_1h['close'].ewm(span=50, adjust=False).mean().iloc[-1]
    last_1h_close = df_1h['close'].iloc[-1]

    trend_1h = 'BULL' if last_1h_close > ema50_1h else 'BEAR'

    return df, trend_1h

# ----------------- 3. 策略逻辑模块 -----------------
def check_fvg(df):
    bullish_fvg = df['low'].iloc[-2] > df['high'].iloc[-4]
    bearish_fvg = df['high'].iloc[-2] < df['low'].iloc[-4]
    return bullish_fvg, bearish_fvg

def check_signal(df, trend_1h):
    if len(df) < 20:
        return None

    c = df.iloc[-2]             # 已确立突破完结 K 线
    consolidation = df.iloc[-7:-2] # 突破前 5 根 3m K 线 (15分钟盘整区)

    # 1. 15 分钟盘整波幅过滤 (< 0.8%)
    high_max = consolidation['high'].max()
    low_min = consolidation['low'].min()
    if ((high_max - low_min) / low_min) > RISK_CONFIG["MAX_RANGE_PCT"]:
        return None

    # 2. 突破 K 线实体质量过滤
    k_range = c['high'] - c['low']
    if k_range == 0:
        return None
    body_size = abs(c['close'] - c['open'])
    if (body_size / k_range) < RISK_CONFIG["MIN_BODY_RATIO"]:
        return None

    # 3. 强化量能过滤 (成交量 >= 盘整区均值 1.8 倍)
    if c['volume'] < (consolidation['volume'].mean() * RISK_CONFIG["VOL_SURGE_RATIO"]):
        return None

    # 4. FVG 缺失流动性缺口校验
    bull_fvg, bear_fvg = check_fvg(df)

    # 5. 多头信号（1h 看多 + 3m MA多头排布 + 净实体收在盘整高价上方 + FVG）
    is_ma_bull = (c['ma5'] > c['ma10']) and (c['close'] > c['ma5'])
    no_break_below_ma10 = (consolidation['close'] >= consolidation['ma10']).all()
    clean_break_high = c['close'] > high_max

    if (trend_1h == 'BULL' and is_ma_bull and no_break_below_ma10 and 
        c['close'] > c['open'] and clean_break_high and bull_fvg):
        return 'BUY'

    # 6. 空头信号（1h 看空 + 3m MA空头排布 + 净实体收在盘整低价下方 + FVG）
    is_ma_bear = (c['ma5'] < c['ma10']) and (c['close'] < c['ma10'])
    no_break_above_ma10 = (consolidation['close'] <= consolidation['ma10']).all()
    clean_break_low = c['close'] < low_min

    if (trend_1h == 'BEAR' and is_ma_bear and no_break_above_ma10 and 
        c['close'] < c['open'] and clean_break_low and bear_fvg):
        return 'SELL'

    return None

# ----------------- 4. 下单与风控模块 -----------------
def execute_trade(symbol, side, df):
    try:
        exchange.set_leverage(RISK_CONFIG["LEVERAGE"], symbol)
        
        last_price = df['close'].iloc[-1]
        atr = df['atr'].iloc[-2]
        
        notional = RISK_CONFIG["MARGIN_PER_TRADE"] * RISK_CONFIG["LEVERAGE"]
        
        exchange.load_markets()
        amount = float(exchange.amount_to_precision(symbol, notional / last_price))
        if amount <= 0:
            return

        sl_distance = max(atr * 1.5, last_price * 0.004) # 3m 级最小止损垫 0.4%
        tp_distance = sl_distance * RISK_CONFIG["RR_RATIO"]

        if side == 'BUY':
            sl_price = last_price - sl_distance
            tp_price = last_price + tp_distance
            order_side, exit_side = 'buy', 'sell'
        else:
            sl_price = last_price + sl_distance
            tp_price = last_price - tp_distance
            order_side, exit_side = 'sell', 'buy'

        sl_price = float(exchange.price_to_precision(symbol, sl_price))
        tp_price = float(exchange.price_to_precision(symbol, tp_price))

        logger.info(f"✨ [3m 突破触发] [{symbol}] 方向: {side} | 现价: {last_price} | SL: {sl_price} | TP: {tp_price}")

        order = exchange.create_order(symbol=symbol, type='market', side=order_side, amount=amount)
        logger.info(f"[{symbol}] 市价开仓成功, 单号: {order['id']}")

        # 挂止损单
        try:
            exchange.create_order(symbol, 'STOP_MARKET', exit_side, amount, params={'stopPrice': sl_price, 'reduceOnly': True})
            logger.info(f"[{symbol}] 止损条件单挂单成功 ({sl_price})")
        except Exception as e:
            logger.critical(f"[{symbol}] 止损挂单失败: {e}，触发强平防护！")
            exchange.create_order(symbol, 'market', exit_side, amount, params={'reduceOnly': True})
            return

        # 挂止盈单
        try:
            exchange.create_order(symbol, 'TAKE_PROFIT_MARKET', exit_side, amount, params={'stopPrice': tp_price, 'reduceOnly': True})
            logger.info(f"[{symbol}] 止盈条件单挂单成功 ({tp_price})")
        except Exception as e:
            logger.error(f"[{symbol}] 止盈挂单失败: {e}")

    except Exception as e:
        logger.error(f"[{symbol}] 开仓流程异常: {e}")

# ----------------- 5. 主程序 -----------------
def main():
    logger.info("🚀 币安 3m 高精度突破战法引擎（1h 趋势对齐 + 15m 盘整过滤）启动成功！")
    symbols = get_top_symbols(RISK_CONFIG["TOP_N_SYMBOLS"])
    
    while True:
        try:
            active_count, active_symbols = get_active_position_count()
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

            time.sleep(5)

        except KeyboardInterrupt:
            break
        except Exception as e:
            logger.error(f"轮询异常: {e}")
            time.sleep(5)

if __name__ == '__main__':
    main()
