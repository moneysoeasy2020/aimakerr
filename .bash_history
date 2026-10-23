# ================= 1. API 配置 =================
API_KEY = "tj49IrEjBmedtwK3BEJNCqP1lk7fXgNkpmi5Yr2eNUhEXHG35XZzpow5TYu3S69y"
SECRET_KEY = "90ibaaGFYMzE6tE9IDqqnGEQZh0qZ76xReZQFt1AHteCRNEs1T76AGh6MDM3Lkno"

exchange = ccxt.binance({
    'apiKey': API_KEY,
    'secret': SECRET_KEY,
    'timeout': 20000,
    'enableRateLimit': True,
    'options': {
        'defaultType': 'future',
        'adjustForTimeDifference': True,
        'recvWindow': 10000,
    }
})

# ================= 2. 策略参数配置 (4H/1D 专属) =================
STRATEGY_CONFIG = {
    "TARGET_SYMBOLS": ["BTC/USDT", "ETH/USDT"], # 仅交易 BTC 与 ETH
    "LEVERAGE": 10,                            # 10x 杠杆
    "MARGIN_PER_TRADE_RATIO": 0.10,            # 适配 4H ATR：单笔使用总资金 10% 保证金 (~50 USDT)
    "STOP_LOSS_ATR_MULT": 1.5,                 # 4H ATR 动态止损倍数
    "TAKE_PROFIT_ATR_MULT": 3.0,                # 4H ATR 动态止盈倍数
    "BREAKEVEN_TRIGGER_ATR": 1.2,              # 浮盈达到 1.2x 4H ATR 时触发保本止损
}

# 运行时持仓追踪器
pos_tracker = {}

# ================= 3. 技术指标算法 =================
def calculate_ema(data, period):
    alpha = 2.0 / (period + 1.0)
    ema = np.empty_like(data)
    ema[0] = data[0]
    for i in range(1, len(data)):
        ema[i] = alpha * data[i] + (1 - alpha) * ema[i - 1]
    return ema

def calculate_atr(high, low, close, period=14):
    tr = np.maximum(high[1:] - low[1:], 
            np.maximum(np.abs(high[1:] - close[:-1]), 
                       np.abs(low[1:] - close[:-1])))
    tr = np.insert(tr, 0, high[0] - low[0])
    atr = np.zeros_like(tr)
    atr[period-1] = np.mean(tr[:period])
    for i in range(period, len(tr)):
        atr[i] = (atr[i-1] * (period - 1) + tr[i]) / period
    return atr

# ================= 4. 账户与风控模块 =================
def initialize_exchange():
    for symbol in STRATEGY_CONFIG["TARGET_SYMBOLS"]:
        try:
            exchange.set_leverage(STRATEGY_CONFIG["LEVERAGE"], symbol)
            logger.info(f"✅ {symbol} 杠杆已设置为 {STRATEGY_CONFIG['LEVERAGE']}x")
        except Exception as e:
            logger.error(f"设置 {symbol} 杠杆失败: {e}")

def get_account_balance():
    try:
        balance = exchange.fetch_balance()
        return float(balance['total']['USDT']), float(balance['free']['USDT'])
    except Exception as e:
        logger.error(f"获取账户余额失败: {e}")
        return 0.0, 0.0

def get_active_positions():
    try:
        positions = exchange.fetch_positions(STRATEGY_CONFIG["TARGET_SYMBOLS"])
        active = {}
        for pos in positions:
            contracts = float(pos.get('contracts', 0) or pos.get('positionAmt', 0))
            if abs(contracts) > 0:
                symbol = pos['symbol']
                active[symbol] = {
                    "contracts": abs(contracts),
                    "side": 'BUY' if contracts > 0 else 'SELL',
                    "entry_price": float(pos.get('entryPrice', 0)),
                    "unrealized_pnl": float(pos.get('unrealizedPnl', 0)),
                }
        return active
    except Exception as e:
        logger.error(f"获取持仓异常: {e}")
        return {}

# ================= 5. 4H/1D 右侧信号引擎 =================
def analyze_signal_4h1d(symbol):
    try:
        # 1. 大周期趋势确认：1D 日线 EMA20 与 EMA50 排列
        ohlcv_1d = exchange.fetch_ohlcv(symbol, timeframe='1d', limit=60)
        if len(ohlcv_1d) < 50:
            return None, 0, 0
        closes_1d = np.array(ohlcv_1d, dtype=float)[:, 4]
        ema20_1d = calculate_ema(closes_1d, 20)
        ema50_1d = calculate_ema(closes_1d, 50)
        
        trend_1d = 'BULL' if ema20_1d[-1] > ema50_1d[-1] else 'BEAR'

        # 2. 中观进场触发：4H 周期突破与 4H ATR 波动率
        ohlcv_4h = exchange.fetch_ohlcv(symbol, timeframe='4h', limit=40)
        if len(ohlcv_4h) < 30:
            return None, 0, 0
        
        np_4h = np.array(ohlcv_4h, dtype=float)
        highs_4h, lows_4h, closes_4h = np_4h[:, 2], np_4h[:, 3], np_4h[:, 4]
        
        atr_4h = calculate_atr(highs_4h, lows_4h, closes_4h, 14)[-1]
        current_close = closes_4h[-1]

        # 4H 唐奇安通道（突破前 20 根 4H K 线最高/最低点）
        donchian_high = np.max(highs_4h[-21:-1])
        donchian_low = np.min(lows_4h[-21:-1])

        signal = None
        if trend_1d == 'BULL' and current_close > donchian_high:
            signal = 'BUY'
        elif trend_1d == 'BEAR' and current_close < donchian_low:
            signal = 'SELL'

        return signal, current_close, atr_4h
    except Exception as e:
        logger.error(f"分析 {symbol} 4H/1D 信号时出错: {e}")
        return None, 0, 0

# ================= 6. 交易执行与出场风控 =================
def execute_trade(symbol, signal, close_price, atr, total_equity):
    target_margin = total_equity * STRATEGY_CONFIG["MARGIN_PER_TRADE_RATIO"]
    position_notional = target_margin * STRATEGY_CONFIG["LEVERAGE"]
    amount = position_notional / close_price

    side = 'buy' if signal == 'BUY' else 'sell'
    
    try:
        order = exchange.create_order(
            symbol=symbol,
            type='market',
            side=side,
            amount=amount
        )
        logger.info(f"🚀 【4H/1D 成功开仓】{symbol} | 方向: {signal} | 数量: {amount:.4f} | 当前价: {close_price}")
        
        sl_distance = atr * STRATEGY_CONFIG["STOP_LOSS_ATR_MULT"]
        tp_distance = atr * STRATEGY_CONFIG["TAKE_PROFIT_ATR_MULT"]
        
        pos_tracker[symbol] = {
            "entry_price": close_price,
            "side": signal,
            "sl": close_price - sl_distance if signal == 'BUY' else close_price + sl_distance,
            "tp": close_price + tp_distance if signal == 'BUY' else close_price - tp_distance,
            "atr": atr,
            "is_breakeven": False
        }
    except Exception as e:
        logger.error(f"❌ 开仓失败 {symbol}: {e}")

def manage_open_positions(active_positions):
    for symbol, pos in active_positions.items():
        if symbol not in pos_tracker:
            continue

        tracker = pos_tracker[symbol]
        curr_side = pos['side']
        entry_price = tracker['entry_price']
        
        ticker = exchange.fetch_ticker(symbol)
        curr_price = ticker['last']
        atr = tracker['atr']

        # 1. 动态保本止损 (盈利超 1.2x 4H ATR 时触发)
        if not tracker['is_breakeven']:
            if (curr_side == 'BUY' and (curr_price - entry_price) >= atr * STRATEGY_CONFIG["BREAKEVEN_TRIGGER_ATR"]) or \
               (curr_side == 'SELL' and (entry_price - curr_price) >= atr * STRATEGY_CONFIG["BREAKEVEN_TRIGGER_ATR"]):
                tracker['sl'] = entry_price
                tracker['is_breakeven'] = True
                logger.info(f"🛡️ 【4H/1D 保本锁仓】{symbol} 利润覆盖 1.2x 4H ATR，止损线提升至成本价: {entry_price}")

        # 2. 止损/止盈出场判定
        need_close = False
        reason = ""
        
        if curr_side == 'BUY':
            if curr_price <= tracker['sl']:
                need_close = True
                reason = "触发 4H ATR 止损/保本线"
            elif curr_price >= tracker['tp']:
                need_close = True
                reason = "触发 4H ATR 3.0x 止盈"
        elif curr_side == 'SELL':
            if curr_price >= tracker['sl']:
                need_close = True
                reason = "触发 4H ATR 止损/保本线"
            elif curr_price <= tracker['tp']:
                need_close = True
                reason = "触发 4H ATR 3.0x 止盈"

        if need_close:
            close_side = 'sell' if curr_side == 'BUY' else 'buy'
            try:
                exchange.create_order(
                    symbol=symbol,
                    type='market',
                    side=close_side,
                    amount=pos['contracts']
                )
                logger.info(f"🎯 【4H/1D 平仓成功】{symbol} | 原因: {reason} | 平仓价: {curr_price}")
                del pos_tracker[symbol]
            except Exception as e:
                logger.error(f"平仓失败 {symbol}: {e}")

# ================= 7. 主循环 (轮询 60 秒) =================
def main_loop():
    logger.info("================ 启动 BTC/ETH 4H/1D 波段趋势策略 ================")
    initialize_exchange()

    while True:
        try:
            total_equity, free_margin = get_account_balance()
            active_positions = get_active_positions()

            if active_positions:
                manage_open_positions(active_positions)

            for symbol in STRATEGY_CONFIG["TARGET_SYMBOLS"]:
                if symbol in active_positions:
                    continue

                if free_margin < total_equity * STRATEGY_CONFIG["MARGIN_PER_TRADE_RATIO"]:
                    logger.warning("可用保证金不足，暂停新开仓。")
                    break

                signal, close_price, atr = analyze_signal_4h1d(symbol)
                if signal:
                    execute_trade(symbol, signal, close_price, atr, total_equity)

            time.sleep(60) # 4H 周期轮询设为 60 秒
        except Exception as e:
            logger.error(f"主循环运行异常: {e}")
            time.sleep(15)

if __name__ == "__main__":
    main_loop()
EOF

cat << 'EOF' > /data/data/com.termux/files/home/yxfuke.py
import time
import logging
import ccxt
import numpy as np

# 日志配置
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - [%(levelname)s] - %(message)s'
)
logger = logging.getLogger("BTC_ETH_4H1D_Strategy")

# ================= 1. API 配置 =================
API_KEY = "tj49IrEjBmedtwK3BEJNCqP1lk7fXgNkpmi5Yr2eNUhEXHG35XZzpow5TYu3S69y"
SECRET_KEY = "90ibaaGFYMzE6tE9IDqqnGEQZh0qZ76xReZQFt1AHteCRNEs1T76AGh6MDM3Lkno"

exchange = ccxt.binance({
    'apiKey': API_KEY,
    'secret': SECRET_KEY,
    'timeout': 20000,
    'enableRateLimit': True,
    'options': {
        'defaultType': 'future',
        'adjustForTimeDifference': True,
        'recvWindow': 10000,
    }
})

# ================= 2. 策略参数配置 (4H/1D 专属) =================
STRATEGY_CONFIG = {
    "TARGET_SYMBOLS": ["BTC/USDT", "ETH/USDT"], # 仅交易 BTC 与 ETH
    "LEVERAGE": 10,                            # 10x 杠杆
    "MARGIN_PER_TRADE_RATIO": 0.10,            # 适配 4H ATR：单笔使用总资金 10% 保证金 (~50 USDT)
    "STOP_LOSS_ATR_MULT": 1.5,                 # 4H ATR 动态止损倍数
    "TAKE_PROFIT_ATR_MULT": 3.0,                # 4H ATR 动态止盈倍数
    "BREAKEVEN_TRIGGER_ATR": 1.2,              # 浮盈达到 1.2x 4H ATR 时触发保本止损
}

# 运行时持仓追踪器
pos_tracker = {}

# ================= 3. 技术指标算法 =================
def calculate_ema(data, period):
    alpha = 2.0 / (period + 1.0)
    ema = np.empty_like(data)
    ema[0] = data[0]
    for i in range(1, len(data)):
        ema[i] = alpha * data[i] + (1 - alpha) * ema[i - 1]
    return ema

def calculate_atr(high, low, close, period=14):
    tr = np.maximum(high[1:] - low[1:], 
            np.maximum(np.abs(high[1:] - close[:-1]), 
                       np.abs(low[1:] - close[:-1])))
    tr = np.insert(tr, 0, high[0] - low[0])
    atr = np.zeros_like(tr)
    atr[period-1] = np.mean(tr[:period])
    for i in range(period, len(tr)):
        atr[i] = (atr[i-1] * (period - 1) + tr[i]) / period
    return atr

# ================= 4. 账户与风控模块 =================
def initialize_exchange():
    for symbol in STRATEGY_CONFIG["TARGET_SYMBOLS"]:
        try:
            exchange.set_leverage(STRATEGY_CONFIG["LEVERAGE"], symbol)
            logger.info(f"✅ {symbol} 杠杆已设置为 {STRATEGY_CONFIG['LEVERAGE']}x")
        except Exception as e:
            logger.error(f"设置 {symbol} 杠杆失败: {e}")

def get_account_balance():
    try:
        balance = exchange.fetch_balance()
        return float(balance['total']['USDT']), float(balance['free']['USDT'])
    except Exception as e:
        logger.error(f"获取账户余额失败: {e}")
        return 0.0, 0.0

def get_active_positions():
    try:
        positions = exchange.fetch_positions(STRATEGY_CONFIG["TARGET_SYMBOLS"])
        active = {}
        for pos in positions:
            contracts = float(pos.get('contracts', 0) or pos.get('positionAmt', 0))
            if abs(contracts) > 0:
                symbol = pos['symbol']
                active[symbol] = {
                    "contracts": abs(contracts),
                    "side": 'BUY' if contracts > 0 else 'SELL',
                    "entry_price": float(pos.get('entryPrice', 0)),
                    "unrealized_pnl": float(pos.get('unrealizedPnl', 0)),
                }
        return active
    except Exception as e:
        logger.error(f"获取持仓异常: {e}")
        return {}

# ================= 5. 4H/1D 右侧信号引擎 =================
def analyze_signal_4h1d(symbol):
    try:
        # 1. 大周期趋势确认：1D 日线 EMA20 与 EMA50 排列
        ohlcv_1d = exchange.fetch_ohlcv(symbol, timeframe='1d', limit=60)
        if len(ohlcv_1d) < 50:
            return None, 0, 0
        closes_1d = np.array(ohlcv_1d, dtype=float)[:, 4]
        ema20_1d = calculate_ema(closes_1d, 20)
        ema50_1d = calculate_ema(closes_1d, 50)
        
        trend_1d = 'BULL' if ema20_1d[-1] > ema50_1d[-1] else 'BEAR'

        # 2. 中观进场触发：4H 周期突破与 4H ATR 波动率
        ohlcv_4h = exchange.fetch_ohlcv(symbol, timeframe='4h', limit=40)
        if len(ohlcv_4h) < 30:
            return None, 0, 0
        
        np_4h = np.array(ohlcv_4h, dtype=float)
        highs_4h, lows_4h, closes_4h = np_4h[:, 2], np_4h[:, 3], np_4h[:, 4]
        
        atr_4h = calculate_atr(highs_4h, lows_4h, closes_4h, 14)[-1]
        current_close = closes_4h[-1]

        # 4H 唐奇安通道（突破前 20 根 4H K 线最高/最低点）
        donchian_high = np.max(highs_4h[-21:-1])
        donchian_low = np.min(lows_4h[-21:-1])

        signal = None
        if trend_1d == 'BULL' and current_close > donchian_high:
            signal = 'BUY'
        elif trend_1d == 'BEAR' and current_close < donchian_low:
            signal = 'SELL'

        return signal, current_close, atr_4h
    except Exception as e:
        logger.error(f"分析 {symbol} 4H/1D 信号时出错: {e}")
        return None, 0, 0

# ================= 6. 交易执行与出场风控 =================
def execute_trade(symbol, signal, close_price, atr, total_equity):
    target_margin = total_equity * STRATEGY_CONFIG["MARGIN_PER_TRADE_RATIO"]
    position_notional = target_margin * STRATEGY_CONFIG["LEVERAGE"]
    amount = position_notional / close_price

    side = 'buy' if signal == 'BUY' else 'sell'
    
    try:
        order = exchange.create_order(
            symbol=symbol,
            type='market',
            side=side,
            amount=amount
        )
        logger.info(f"🚀 【4H/1D 成功开仓】{symbol} | 方向: {signal} | 数量: {amount:.4f} | 当前价: {close_price}")
        
        sl_distance = atr * STRATEGY_CONFIG["STOP_LOSS_ATR_MULT"]
        tp_distance = atr * STRATEGY_CONFIG["TAKE_PROFIT_ATR_MULT"]
        
        pos_tracker[symbol] = {
            "entry_price": close_price,
            "side": signal,
            "sl": close_price - sl_distance if signal == 'BUY' else close_price + sl_distance,
            "tp": close_price + tp_distance if signal == 'BUY' else close_price - tp_distance,
            "atr": atr,
            "is_breakeven": False
        }
    except Exception as e:
        logger.error(f"❌ 开仓失败 {symbol}: {e}")

def manage_open_positions(active_positions):
    for symbol, pos in active_positions.items():
        if symbol not in pos_tracker:
            continue

        tracker = pos_tracker[symbol]
        curr_side = pos['side']
        entry_price = tracker['entry_price']
        
        ticker = exchange.fetch_ticker(symbol)
        curr_price = ticker['last']
        atr = tracker['atr']

        # 1. 动态保本止损 (盈利超 1.2x 4H ATR 时触发)
        if not tracker['is_breakeven']:
            if (curr_side == 'BUY' and (curr_price - entry_price) >= atr * STRATEGY_CONFIG["BREAKEVEN_TRIGGER_ATR"]) or \
               (curr_side == 'SELL' and (entry_price - curr_price) >= atr * STRATEGY_CONFIG["BREAKEVEN_TRIGGER_ATR"]):
                tracker['sl'] = entry_price
                tracker['is_breakeven'] = True
                logger.info(f"🛡️ 【4H/1D 保本锁仓】{symbol} 利润覆盖 1.2x 4H ATR，止损线提升至成本价: {entry_price}")

        # 2. 止损/止盈出场判定
        need_close = False
        reason = ""
        
        if curr_side == 'BUY':
            if curr_price <= tracker['sl']:
                need_close = True
                reason = "触发 4H ATR 止损/保本线"
            elif curr_price >= tracker['tp']:
                need_close = True
                reason = "触发 4H ATR 3.0x 止盈"
        elif curr_side == 'SELL':
            if curr_price >= tracker['sl']:
                need_close = True
                reason = "触发 4H ATR 止损/保本线"
            elif curr_price <= tracker['tp']:
                need_close = True
                reason = "触发 4H ATR 3.0x 止盈"

        if need_close:
            close_side = 'sell' if curr_side == 'BUY' else 'buy'
            try:
                exchange.create_order(
                    symbol=symbol,
                    type='market',
                    side=close_side,
                    amount=pos['contracts']
                )
                logger.info(f"🎯 【4H/1D 平仓成功】{symbol} | 原因: {reason} | 平仓价: {curr_price}")
                del pos_tracker[symbol]
            except Exception as e:
                logger.error(f"平仓失败 {symbol}: {e}")

# ================= 7. 主循环 (轮询 60 秒) =================
def main_loop():
    logger.info("================ 启动 BTC/ETH 4H/1D 波段趋势策略 ================")
    initialize_exchange()

    while True:
        try:
            total_equity, free_margin = get_account_balance()
            active_positions = get_active_positions()

            if active_positions:
                manage_open_positions(active_positions)

            for symbol in STRATEGY_CONFIG["TARGET_SYMBOLS"]:
                if symbol in active_positions:
                    continue

                if free_margin < total_equity * STRATEGY_CONFIG["MARGIN_PER_TRADE_RATIO"]:
                    logger.warning("可用保证金不足，暂停新开仓。")
                    break

                signal, close_price, atr = analyze_signal_4h1d(symbol)
                if signal:
                    execute_trade(symbol, signal, close_price, atr, total_equity)

            time.sleep(60) # 4H 周期轮询设为 60 秒
        except Exception as e:
            logger.error(f"主循环运行异常: {e}")
            time.sleep(15)

if __name__ == "__main__":
    main_loop()
EOF

,0)),
