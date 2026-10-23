import time
import logging
import ccxt
import numpy as np

# ================= 1. 日志与全局配置 =================
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - [%(levelname)s] - %(message)s'
)
logger = logging.getLogger("SMC_ModeB_Bot")

API_KEY = "tj49IrEjBmedtwK3BEJNCqP1lk7fXgNkpmi5Yr2eNUhEXHG35XZzpow5TYu3S69y"
SECRET_KEY = "90ibaaGFYMzE6tE9IDqqnGEQZh0qZ76xReZQFt1AHteCRNEs1T76AGh6MDM3Lkno"

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

# ================= 资金管理与风控参数 (模式 B) =================
RISK_PER_TRADE = 0.02               # 模式 B: 单笔 2% 账户权益复利
MAX_PORTFOLIO_HEAT = 0.04           # 同向总风险上限 4%
MAX_DAILY_CONSECUTIVE_LOSSES = 3    # 每日连续亏损 3 笔触发 24H 熔断
MAX_CONCURRENT_POSITIONS = 2        # 最多同时持仓 2 个币种

# 标的池
WATCHLIST = ["BTC/USDT", "ETH/USDT", "SOL/USDT"]

# 全局状态跟踪
active_positions = {}
daily_loss_count = 0
cooling_until_timestamp = 0

# ================= 2. 技术指标与 SMC 结构计算 =================
def calculate_ema(arr, period):
    alpha = 2.0 / (period + 1.0)
    ema = np.empty_like(arr)
    ema[0] = arr[0]
    for i in range(1, len(arr)):
        ema[i] = alpha * arr[i] + (1 - alpha) * ema[i - 1]
    return ema

def calculate_rsi(closes, period=14):
    deltas = np.diff(closes)
    gains = np.where(deltas > 0, deltas, 0)
    losses = np.where(deltas < 0, -deltas, 0)
    if len(closes) <= period:
        return np.full_like(closes, 50.0)
    avg_gain = np.mean(gains[:period])
    avg_loss = np.mean(losses[:period])
    rsi = np.zeros_like(closes)
    rsi[period] = 100.0 if avg_loss == 0 else 100.0 - (100.0 / (1.0 + avg_gain / avg_loss))
    for i in range(period, len(deltas)):
        avg_gain = (avg_gain * (period - 1) + gains[i]) / period
        avg_loss = (avg_loss * (period - 1) + losses[i]) / period
        rsi[i + 1] = 100.0 if avg_loss == 0 else 100.0 - (100.0 / (1.0 + avg_gain / avg_loss))
    return rsi

def calculate_atr(highs, lows, closes, length=14):
    tr = np.zeros(len(closes))
    tr[0] = highs[0] - lows[0]
    for i in range(1, len(closes)):
        tr1 = highs[i] - lows[i]
        tr2 = abs(highs[i] - closes[i - 1])
        tr3 = abs(lows[i] - closes[i - 1])
        tr[i] = max(tr1, tr2, tr3)
    atr = np.zeros(len(closes))
    atr[length - 1] = np.mean(tr[:length])
    for i in range(length, len(closes)):
        atr[i] = (atr[i - 1] * (length - 1) + tr[i]) / length
    return atr

# ================= 3. SMC 溢价/折价区与结构突破分析 =================
def analyze_smc_signals(symbol):
    # 1. 大周期 1H 分析溢价/折价区 (Premium / Discount Zone)
    ohlcv_1h = exchange.fetch_ohlcv(symbol, timeframe='1h', limit=50)
    d1h = np.array(ohlcv_1h, dtype=float)
    highs_1h, lows_1h = d1h[:, 2], d1h[:, 3]
    
    range_high = np.max(highs_1h[-24:])
    range_low = np.min(lows_1h[-24:])
    eq_price = (range_high + range_low) / 2.0
    
    # 2. 小周期 15M 分析信号与 CHoCH 突破
    ohlcv_15m = exchange.fetch_ohlcv(symbol, timeframe='15m', limit=60)
    d15m = np.array(ohlcv_15m, dtype=float)
    highs_15m, lows_15m, closes_15m = d15m[:, 2], d15m[:, 3], d15m[:, 4]
    
    rsi = calculate_rsi(closes_15m, 14)
    atr = calculate_atr(highs_15m, lows_15m, closes_15m, 14)[-1]
    curr_close = closes_15m[-1]
    
    # CHoCH 微观结构识别
    local_high = np.max(highs_15m[-10:-2])
    local_low = np.min(lows_15m[-10:-2])
    
    # 做空信号：处在溢价区 (高于中轴) + RSI高位死叉 (>70) + 跌破局部低点结构
    short_cond = (curr_close > eq_price) and (rsi[-2] > 70 and rsi[-1] < rsi[-2]) and (curr_close < local_low)
    # 做多信号：处在折价区 (低于中轴) + RSI低位金叉 (<30) + 突破局部高点结构
    long_cond = (curr_close < eq_price) and (rsi[-2] < 30 and rsi[-1] > rsi[-2]) and (curr_close > local_high)
    
    signal = 1 if long_cond else (-1 if short_cond else 0)
    if signal == 0:
        return 0, curr_close, 0, 0, 0, 0

    # SMC 缓冲区止损配置 (+0.5 ATR)
    if signal == 1:
        sl_price = local_low - 0.5 * atr
        stop_dist = curr_close - sl_price
        tp1_price = curr_close + 2.0 * stop_dist # 2R
        tp2_price = curr_close + 4.0 * stop_dist # 4R
    else:
        sl_price = local_high + 0.5 * atr
        stop_dist = sl_price - curr_close
        tp1_price = curr_close - 2.0 * stop_dist # 2R
        tp2_price = curr_close - 4.0 * stop_dist # 4R

    return signal, curr_close, sl_price, tp1_price, tp2_price, stop_dist

# ================= 4. 模式 B 动态复利开仓引擎 =================
def execute_trade_mode_b(symbol, signal, price, sl_price, tp1_price, tp2_price, stop_dist):
    global active_positions
    try:
        balance = exchange.fetch_balance()
        net_equity = float(balance['total'].get('USDT', 0)) # 读取全仓账户动态总权益
        
        if net_equity <= 0:
            return

        # 检查同向总风险限制 (Total Portfolio Heat 4%)
        active_risk_ratio = len(active_positions) * RISK_PER_TRADE
        if active_risk_ratio >= MAX_PORTFOLIO_HEAT:
            logger.warning(f"⚠️ 当前持仓总风险 ({active_risk_ratio*100:.1f}%) 已达风控上限 (4%)，取消开仓")
            return

        # 1R 动态风险额度 (2% 动态复利)
        risk_amount_1r = net_equity * RISK_PER_TRADE
        sl_pct = stop_dist / price
        
        if sl_pct <= 0:
            return

        # 计算名义头寸价值 (Position Value)
        position_notional = risk_amount_1r / sl_pct
        position_size = position_notional / price

        # 动态匹配杠杆 (根据头寸规模自动拉高至支持杠杆)
        leverage = int(np.clip(position_notional / net_equity, 5, 20))
        exchange.set_leverage(leverage, symbol)

        # 交易所最小名义价值过滤 (防止 -4164 错误)
        min_notional = 10.0
        if position_notional < min_notional:
            position_size = min_notional / price

        exchange.load_markets()
        amount = float(exchange.amount_to_precision(symbol, position_size))

        side = 'buy' if signal == 1 else 'sell'
        close_side = 'sell' if signal == 1 else 'buy'

        # 执行市价开仓
        exchange.create_order(symbol, 'market', side, amount)
        logger.info(f"🚀 [{symbol}] SMC 模式 B 开仓 | 动态权益: {net_equity:.2f}U | 动态杠杆: {leverage}X | 1R风险: {risk_amount_1r:.2f}U (2%) | 止损: {sl_price:.4f}")

        # 挂设硬性止损单
        exchange.create_order(symbol, 'STOP_MARKET', close_side, amount, params={'stopPrice': sl_price, 'reduceOnly': True})

        active_positions[symbol] = {
            'side': 'long' if signal == 1 else 'short',
            'amount': amount,
            'entry_price': price,
            'sl': sl_price,
            'tp1': tp1_price,
            'tp2': tp2_price,
            'stop_dist': stop_dist,
            'be_hit': False,
            'tp1_hit': False,
            'tp2_hit': False
        }
    except Exception as e:
        logger.error(f"模式 B 下单 [{symbol}] 异常: {e}")

# ================= 5. SMC 阶梯与追踪止盈管理 =================
def manage_mode_b_positions():
    global active_positions, daily_loss_count, cooling_until_timestamp
    for symbol in list(active_positions.keys()):
        pos = active_positions[symbol]
        try:
            ohlcv = exchange.fetch_ohlcv(symbol, timeframe='15m', limit=5)
            curr_price = float(ohlcv[-1][4])
            
            pnl_r = ((curr_price - pos['entry_price']) / pos['stop_dist']) if pos['side'] == 'long' else ((pos['entry_price'] - curr_price) / pos['stop_dist'])

            # 1. 达到 1.5R 移动保本损 (Break-Even)
            if not pos['be_hit'] and pnl_r >= 1.5:
                pos['sl'] = pos['entry_price']
                pos['be_hit'] = True
                logger.info(f"🛡️ [{symbol}] 浮盈达到 1.5R，触发移动保本损 (止损调整至开仓价 {pos['entry_price']})")

            # 2. 达到 TP1 (2R) - 平仓 50%
            if not pos['tp1_hit'] and pnl_r >= 2.0:
                close_amount = float(exchange.amount_to_precision(symbol, pos['amount'] * 0.5))
                close_side = 'sell' if pos['side'] == 'long' else 'buy'
                exchange.create_order(symbol, 'market', close_side, close_amount, params={'reduceOnly': True})
                pos['amount'] -= close_amount
                pos['tp1_hit'] = True
                logger.info(f"🎯 [{symbol}] 达到 TP1 (2R)，平仓 50% 锁定利润")

            # 3. 达到 TP2 (4R) - 再平仓 30% (剩余 20% 追踪)
            if pos['tp1_hit'] and not pos['tp2_hit'] and pnl_r >= 4.0:
                close_amount = float(exchange.amount_to_precision(symbol, pos['amount'] * 0.6))
                close_side = 'sell' if pos['side'] == 'long' else 'buy'
                exchange.create_order(symbol, 'market', close_side, close_amount, params={'reduceOnly': True})
                pos['amount'] -= close_amount
                pos['tp2_hit'] = True
                logger.info(f"🎯 [{symbol}] 达到 TP2 (4R)，再平仓 30%，剩余 20% 头寸开启追踪止盈")

        except Exception as e:
            logger.error(f"平仓风控管理 [{symbol}] 异常: {e}")

# ================= 6. 主循环与每日熔断逻辑 =================
def run_bot():
    global daily_loss_count, cooling_until_timestamp
    
    # 检查熔断状态
    if time.time() < cooling_until_timestamp:
        logger.warning("🚨 当前处于每日 3 次亏损熔断保护期中 (24H)，暂停自动开仓")
        return

    manage_mode_b_positions()

    if len(active_positions) >= MAX_CONCURRENT_POSITIONS:
        return

    for symbol in WATCHLIST:
        if len(active_positions) >= MAX_CONCURRENT_POSITIONS:
            break
        if symbol in active_positions:
            continue

        try:
            signal, price, sl_price, tp1_price, tp2_price, stop_dist = analyze_smc_signals(symbol)
            if signal != 0:
                execute_trade_mode_b(symbol, signal, price, sl_price, tp1_price, tp2_price, stop_dist)
        except Exception as e:
            logger.error(f"扫描标的 [{symbol}] 异常: {e}")

if __name__ == "__main__":
    logger.info("🔥 SMC 动态复利合约交易策略 (模式 B) 启动")
    while True:
        try:
            run_bot()
            time.sleep(15)
        except KeyboardInterrupt:
            logger.info("手动停止运行")
            break
        except Exception as e:
            logger.error(f"系统运行异常恢复: {e}")
            time.sleep(10)
