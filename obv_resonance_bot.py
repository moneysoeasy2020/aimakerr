import time
import logging
import ccxt
import numpy as np

# ================= 1. 日志与全局配置 =================
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - [%(levelname)s] - %(message)s'
)
logger = logging.getLogger("OBV_Resonance_Bot")

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

# 策略风控参数
LEVERAGE = 5                        # 5 倍杠杆
RISK_PER_TRADE = 0.005              # 单笔最大亏损 0.5%
MAX_ACCOUNT_MARGIN_RATIO = 0.20     # 总保证金占用不超过 20%
MAX_SINGLE_POS_NOTIONAL_RATIO = 0.10# 单币名义仓位上限 10%
MAX_DAILY_CONSECUTIVE_LOSSES = 3    # 连续止损 3 次触发每日停手

# ================= 指定分类标的池 (按梯队精准匹配) =================
WATCHLIST_CONFIG = [
    # 1–10 头部币种 (1.0 x ATR)
    ("BTC/USDT", 1.0, 1), ("ETH/USDT", 1.0, 2), ("BNB/USDT", 1.0, 3), ("SOL/USDT", 1.0, 4), ("XRP/USDT", 1.0, 5),
    ("ADA/USDT", 1.0, 6), ("DOGE/USDT", 1.0, 7), ("AVAX/USDT", 1.0, 8), ("POL/USDT", 1.0, 9), ("DOT/USDT", 1.0, 10),
    
    # 11–20 中腰部币种 (0.8 x ATR)
    ("LINK/USDT", 0.8, 11), ("UNI/USDT", 0.8, 12), ("ARB/USDT", 0.8, 13), ("OP/USDT", 0.8, 14), ("SUI/USDT", 0.8, 15),
    ("SEI/USDT", 0.8, 16), ("APT/USDT", 0.8, 17), ("LTC/USDT", 0.8, 18), ("BCH/USDT", 0.8, 19), ("FIL/USDT", 0.8, 20),
    
    # 21–30 尾部币种 (0.7 x ATR)
    ("ETC/USDT", 0.7, 21), ("ATOM/USDT", 0.7, 22), ("JUP/USDT", 0.7, 23), ("PYTH/USDT", 0.7, 24), ("RENDER/USDT", 0.7, 25),
    ("TAO/USDT", 0.7, 26), ("WLD/USDT", 0.7, 27), ("FTM/USDT", 0.7, 28), ("NEAR/USDT", 0.7, 29), ("CRV/USDT", 0.7, 30)
]

# 全局持仓管理 (仅允许 1 多 + 1 空)
active_positions = {}
daily_losses_count = 0
blacklisted_today = set()

# ================= 2. 纯 NumPy 技术指标计算 =================
def calculate_ema(arr, period):
    alpha = 2.0 / (period + 1.0)
    ema = np.empty_like(arr)
    ema[0] = arr[0]
    for i in range(1, len(arr)):
        ema[i] = alpha * arr[i] + (1 - alpha) * ema[i - 1]
    return ema

def calculate_macd(closes, fast=12, slow=26, signal=9):
    ema_fast = calculate_ema(closes, fast)
    ema_slow = calculate_ema(closes, slow)
    macd_line = ema_fast - ema_slow
    signal_line = calculate_ema(macd_line, signal)
    hist = macd_line - signal_line
    return macd_line, signal_line, hist

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

def calculate_adx(highs, lows, closes, length=14):
    n = len(closes)
    if n < length + 1:
        return np.zeros(n)
    up_move = highs[1:] - highs[:-1]
    down_move = lows[:-1] - lows[1:]
    plus_dm = np.where((up_move > down_move) & (up_move > 0), up_move, 0.0)
    minus_dm = np.where((down_move > up_move) & (down_move > 0), down_move, 0.0)
    tr = np.maximum(highs[1:] - lows[1:], np.maximum(abs(highs[1:] - closes[:-1]), abs(lows[1:] - closes[:-1])))
    
    atr_smooth = np.zeros(n - 1)
    pdm_smooth = np.zeros(n - 1)
    mdm_smooth = np.zeros(n - 1)
    
    atr_smooth[length - 1] = np.sum(tr[:length])
    pdm_smooth[length - 1] = np.sum(plus_dm[:length])
    mdm_smooth[length - 1] = np.sum(minus_dm[:length])
    
    for i in range(length, n - 1):
        atr_smooth[i] = atr_smooth[i - 1] - (atr_smooth[i - 1] / length) + tr[i]
        pdm_smooth[i] = pdm_smooth[i - 1] - (pdm_smooth[i - 1] / length) + plus_dm[i]
        mdm_smooth[i] = mdm_smooth[i - 1] - (mdm_smooth[i - 1] / length) + minus_dm[i]
        
    p_di = 100 * (pdm_smooth / np.where(atr_smooth == 0, 1e-9, atr_smooth))
    m_di = 100 * (mdm_smooth / np.where(atr_smooth == 0, 1e-9, atr_smooth))
    dx = 100 * abs(p_di - m_di) / np.where((p_di + m_di) == 0, 1e-9, (p_di + m_di))
    
    adx = np.zeros(n)
    adx[2 * length] = np.mean(dx[length:2 * length])
    for i in range(2 * length + 1, n):
        adx[i] = (adx[i - 1] * (length - 1) + dx[i - 1]) / length
    return adx

def calculate_stoch_rsi(closes, length=14, k_period=3, d_period=3):
    deltas = np.diff(closes)
    gains = np.where(deltas > 0, deltas, 0)
    losses = np.where(deltas < 0, -deltas, 0)
    
    rsi = np.zeros_like(closes)
    if len(closes) <= length:
        return np.full_like(closes, 50.0)
        
    avg_gain = np.mean(gains[:length])
    avg_loss = np.mean(losses[:length])
    rsi[length] = 100.0 if avg_loss == 0 else 100.0 - (100.0 / (1.0 + avg_gain / avg_loss))
    
    for i in range(length, len(deltas)):
        avg_gain = (avg_gain * (length - 1) + gains[i]) / length
        avg_loss = (avg_loss * (length - 1) + losses[i]) / length
        rsi[i + 1] = 100.0 if avg_loss == 0 else 100.0 - (100.0 / (1.0 + avg_gain / avg_loss))
        
    stoch_rsi = np.zeros_like(rsi)
    for i in range(length + length, len(rsi)):
        window = rsi[i - length + 1:i + 1]
        min_val = np.min(window)
        max_val = np.max(window)
        stoch_rsi[i] = (rsi[i] - min_val) / (max_val - min_val) * 100.0 if max_val != min_val else 50.0
        
    stoch_k = calculate_ema(stoch_rsi, k_period)
    return stoch_k

def calculate_kdj(highs, lows, closes, n=9, m1=3, m2=3):
    k = np.zeros_like(closes)
    d = np.zeros_like(closes)
    k[0], d[0] = 50.0, 50.0
    
    for i in range(1, len(closes)):
        if i < n:
            low_val = np.min(lows[:i + 1])
            high_val = np.max(highs[:i + 1])
        else:
            low_val = np.min(lows[i - n + 1:i + 1])
            high_val = np.max(highs[i - n + 1:i + 1])
            
        rsv = (closes[i] - low_val) / (high_val - low_val) * 100.0 if high_val != low_val else 50.0
        k[i] = (m1 - 1) / m1 * k[i - 1] + 1 / m1 * rsv
        d[i] = (m2 - 1) / m2 * d[i - 1] + 1 / m2 * k[i]
        
    j = 3 * k - 2 * d
    return k, d, j

def calculate_obv(closes, volumes):
    obv = np.zeros_like(closes)
    obv[0] = volumes[0]
    for i in range(1, len(closes)):
        if closes[i] > closes[i - 1]:
            obv[i] = obv[i - 1] + volumes[i]
        elif closes[i] < closes[i - 1]:
            obv[i] = obv[i - 1] - volumes[i]
        else:
            obv[i] = obv[i - 1]
    return obv

# ================= 3. 跨周期全共振信号识别引擎 =================
def analyze_symbol_signals(symbol, atr_mult, rank_index):
    ohlcv_1h = exchange.fetch_ohlcv(symbol, timeframe='1h', limit=50)
    data_1h = np.array(ohlcv_1h, dtype=float)
    closes_1h, vols_1h = data_1h[:, 4], data_1h[:, 5]
    obv_1h = calculate_obv(closes_1h, vols_1h)
    
    obv_up = obv_1h[-1] > obv_1h[-2] > obv_1h[-3]
    obv_down = obv_1h[-1] < obv_1h[-2] < obv_1h[-3]
    
    if not obv_up and not obv_down:
        return 0, 0, 0, 0, 0

    ohlcv_15m = exchange.fetch_ohlcv(symbol, timeframe='15m', limit=100)
    data_15m = np.array(ohlcv_15m, dtype=float)
    highs, lows, closes = data_15m[:, 2], data_15m[:, 3], data_15m[:, 4]
    
    ema20 = calculate_ema(closes, 20)
    ema50 = calculate_ema(closes, 50)
    adx = calculate_adx(highs, lows, closes, 14)
    _, _, macd_hist = calculate_macd(closes)
    stoch_k = calculate_stoch_rsi(closes)
    k_line, d_line, _ = calculate_kdj(highs, lows, closes)
    atr = calculate_atr(highs, lows, closes, 14)[-1]
    
    curr_close = closes[-1]
    
    if adx[-1] < 20:
        return 0, curr_close, 0, 0, 0

    long_signal = (
        obv_up and 
        (ema20[-1] > ema50[-1]) and 
        (macd_hist[-1] > 0 and macd_hist[-1] > macd_hist[-2]) and 
        (15 <= stoch_k[-1] <= 80 and stoch_k[-1] > stoch_k[-2]) and 
        (k_line[-1] > d_line[-1] and k_line[-2] <= d_line[-2])
    )
    
    short_signal = (
        obv_down and 
        (ema20[-1] < ema50[-1]) and 
        (macd_hist[-1] < 0 and macd_hist[-1] < macd_hist[-2]) and 
        (20 <= stoch_k[-1] <= 85 and stoch_k[-1] < stoch_k[-2]) and 
        (k_line[-1] < d_line[-1] and k_line[-2] >= d_line[-2])
    )
    
    signal = 1 if long_signal else (-1 if short_signal else 0)
    if signal == 0:
        return 0, curr_close, 0, 0, 0

    stop_dist = atr * atr_mult
    if signal == 1:
        sl_price = curr_close - stop_dist
        tp1_price = curr_close + 1.5 * stop_dist
        tp2_price = curr_close + 2.5 * stop_dist
    else:
        sl_price = curr_close + stop_dist
        tp1_price = curr_close - 1.5 * stop_dist
        tp2_price = curr_close - 2.5 * stop_dist

    return signal, curr_close, sl_price, tp1_price, tp2_price

# ================= 4. 持仓管理与风控执行 =================
def manage_active_positions():
    global active_positions
    for symbol in list(active_positions.keys()):
        pos = active_positions[symbol]
        try:
            ohlcv = exchange.fetch_ohlcv(symbol, timeframe='15m', limit=50)
            data = np.array(ohlcv, dtype=float)
            closes = data[:, 4]
            curr_price = closes[-1]
            
            ohlcv_1h = exchange.fetch_ohlcv(symbol, timeframe='1h', limit=10)
            data_1h = np.array(ohlcv_1h, dtype=float)
            obv_1h = calculate_obv(data_1h[:, 4], data_1h[:, 5])
            
            _, _, macd_hist = calculate_macd(closes)
            stoch_k = calculate_stoch_rsi(closes)
            
            indicator_exit = False
            if pos['side'] == 'long':
                if (obv_1h[-1] < obv_1h[-2]) or (macd_hist[-1] < 0) or (stoch_k[-1] > 85):
                    indicator_exit = True
            else:
                if (obv_1h[-1] > obv_1h[-2]) or (macd_hist[-1] > 0) or (stoch_k[-1] < 15):
                    indicator_exit = True

            if indicator_exit:
                logger.info(f"🚨 [{symbol}] 指标失效，强制平仓离场")
                close_side = 'sell' if pos['side'] == 'long' else 'buy'
                exchange.create_order(symbol, 'market', close_side, pos['amount'], params={'reduceOnly': True})
                del active_positions[symbol]
                continue

            if not pos['tp1_hit']:
                if (pos['side'] == 'long' and curr_price >= pos['tp1']) or \
                   (pos['side'] == 'short' and curr_price <= pos['tp1']):
                    logger.info(f"🎯 [{symbol}] 达到 1.5R 止盈点，平仓 50% 并保本")
                    half_amount = float(exchange.amount_to_precision(symbol, pos['amount'] * 0.5))
                    close_side = 'sell' if pos['side'] == 'long' else 'buy'
                    exchange.create_order(symbol, 'market', close_side, half_amount, params={'reduceOnly': True})
                    pos['amount'] -= half_amount
                    pos['sl'] = pos['entry_price']
                    pos['tp1_hit'] = True

            if pos['tp1_hit']:
                if (pos['side'] == 'long' and curr_price >= pos['tp2']) or \
                   (pos['side'] == 'short' and curr_price <= pos['tp2']):
                    logger.info(f"🎉 [{symbol}] 达到 2.5R 止盈点，全部平仓")
                    close_side = 'sell' if pos['side'] == 'long' else 'buy'
                    exchange.create_order(symbol, 'market', close_side, pos['amount'], params={'reduceOnly': True})
                    del active_positions[symbol]
                    continue
        except Exception as e:
            logger.error(f"持仓管理 [{symbol}] 异常: {e}")

def execute_trade(symbol, signal, price, sl_price, tp1_price, tp2_price, atr_mult, rank_index):
    global active_positions
    try:
        balance = exchange.fetch_balance()
        usdt_free = float(balance['total'].get('USDT', 0))
        if usdt_free <= 0:
            return

        risk_amount = usdt_free * RISK_PER_TRADE
        stop_dist_abs = abs(price - sl_price)
        if stop_dist_abs == 0:
            return

        position_size = risk_amount / stop_dist_abs
        max_notional = usdt_free * MAX_SINGLE_POS_NOTIONAL_RATIO
        if (position_size * price) > max_notional:
            position_size = max_notional / price

        exchange.set_leverage(LEVERAGE, symbol)
        exchange.load_markets()
        amount = float(exchange.amount_to_precision(symbol, position_size))
        
        if amount <= 0:
            return

        side = 'buy' if signal == 1 else 'sell'
        close_side = 'sell' if signal == 1 else 'buy'

        exchange.create_order(symbol, 'market', side, amount)
        logger.info(f"🚀 [{symbol}] 开仓成功 | TIER #{rank_index} | 杠杆: 5X | 止损倍数: {atr_mult}xATR | 数量: {amount} | 止损价: {sl_price:.4f}")

        exchange.create_order(symbol, 'STOP_MARKET', close_side, amount, params={'stopPrice': sl_price, 'reduceOnly': True})

        active_positions[symbol] = {
            'side': 'long' if signal == 1 else 'short',
            'amount': amount,
            'entry_price': price,
            'sl': sl_price,
            'tp1': tp1_price,
            'tp2': tp2_price,
            'rank': rank_index,
            'tp1_hit': False
        }
    except Exception as e:
        logger.error(f"下单开仓 [{symbol}] 失败: {e}")

# ================= 5. 主逻辑循环 =================
def run_bot():
    if daily_losses_count >= MAX_DAILY_CONSECUTIVE_LOSSES:
        logger.warning("触发连续亏损 3 次熔断，今日暂停交易")
        return

    manage_active_positions()
    
    current_longs = sum(1 for p in active_positions.values() if p['side'] == 'long')
    current_shorts = sum(1 for p in active_positions.values() if p['side'] == 'short')
    
    if len(active_positions) >= 2:
        return

    for symbol, atr_mult, rank_index in WATCHLIST_CONFIG:
        if len(active_positions) >= 2:
            break
        if symbol in active_positions or symbol in blacklisted_today:
            continue
            
        try:
            signal, price, sl_price, tp1_price, tp2_price = analyze_symbol_signals(symbol, atr_mult, rank_index)
            
            if signal == 1 and current_longs > 0:
                continue
            if signal == -1 and current_shorts > 0:
                continue
                
            if signal != 0:
                execute_trade(symbol, signal, price, sl_price, tp1_price, tp2_price, atr_mult, rank_index)
                if signal == 1:
                    current_longs += 1
                else:
                    current_shorts += 1
        except Exception as e:
            logger.error(f"扫描标的 [{symbol}] 异常: {e}")

if __name__ == "__main__":
    logger.info("🔥 启动阶梯化指定 TOP30 标的池短线合约机器人（5X 杠杆）")
    while True:
        try:
            run_bot()
            time.sleep(15)
        except KeyboardInterrupt:
            logger.info("手动退出程序")
            break
        except Exception as e:
            logger.error(f"主程序异常恢复: {e}")
            time.sleep(10)
