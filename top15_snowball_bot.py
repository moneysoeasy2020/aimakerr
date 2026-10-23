import time
import logging
import ccxt
import numpy as np

# ================= 1. 日志与 API 基础配置 =================
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - [%(levelname)s] - %(message)s'
)
logger = logging.getLogger("Top15_Resonance_Snowball")

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

# ================= 策略与风控参数 =================
FIXED_MARGIN_PER_TRADE = 3.0      # 首仓固定本金 3 USDT
LEVERAGE = 10                     # 固定 10 倍杠杆 (名义价值 30 USDT)
MAX_POSITIONS = 3                 # 全局最多同时持仓 3 个币种

SNOWBALL_TRIGGER_R = 1.5          # 浮盈 1.5R 触发滚雪球加仓
SNOWBALL_ADD_RATIO = 0.5          # 滚雪球加仓幅度为原持仓的 50%

active_positions = {}

# ================= 2. 纯 NumPy 技术指标算法 =================
def calculate_ema(arr, period):
    alpha = 2.0 / (period + 1.0)
    ema = np.empty_like(arr)
    ema[0] = arr[0]
    for i in range(1, len(arr)):
        ema[i] = alpha * arr[i] + (1 - alpha) * ema[i - 1]
    return ema

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

def calculate_stoch_rsi(closes, length=14, k_period=3, d_period=3):
    deltas = np.diff(closes)
    gains = np.where(deltas > 0, deltas, 0)
    losses = np.where(deltas < 0, -deltas, 0)
    
    rsi = np.zeros_like(closes)
    if len(closes) <= length:
        return np.full_like(closes, 50.0), np.full_like(closes, 50.0)
        
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
    stoch_d = calculate_ema(stoch_k, d_period)
    return stoch_k, stoch_d

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
        
    return k, d

# ================= 3. 涨幅榜 TOP 15 抓取 =================
def get_top15_gainers():
    try:
        tickers = exchange.fetch_tickers()
        usdt_futures = []
        for symbol, ticker in tickers.items():
            if symbol.endswith('/USDT:USDT') or symbol.endswith('/USDT'):
                percentage = ticker.get('percentage', 0) or 0
                clean_symbol = symbol.split(':')[0] if ':' in symbol else symbol
                usdt_futures.append((clean_symbol, float(percentage)))
        
        usdt_futures.sort(key=lambda x: x[1], reverse=True)
        top15 = [item[0] for item in usdt_futures[:15]]
        return top15
    except Exception as e:
        logger.error(f"获取涨幅榜 TOP15 失败: {e}")
        return []

# ================= 4. 新核心逻辑：OBV + KDJ + StochRSI 共振 =================
def check_timeframe_signal(symbol, tf):
    ohlcv = exchange.fetch_ohlcv(symbol, timeframe=tf, limit=60)
    data = np.array(ohlcv, dtype=float)
    highs, lows, closes, vols = data[:, 2], data[:, 3], data[:, 4], data[:, 5]
    
    # 1. OBV 拐头向上/向下
    obv = calculate_obv(closes, vols)
    obv_up = obv[-1] > obv[-2] > obv[-3]
    obv_down = obv[-1] < obv[-2] < obv[-3]
    
    # 2. KDJ 金叉/死叉 (或发散状态)
    k_line, d_line = calculate_kdj(highs, lows, closes)
    kdj_gold = (k_line[-1] > d_line[-1]) and (k_line[-2] <= d_line[-2] or k_line[-1] > k_line[-2])
    kdj_dead = (k_line[-1] < d_line[-1]) and (k_line[-2] >= d_line[-2] or k_line[-1] < k_line[-2])
    
    # 3. StochRSI 金叉/死叉
    stoch_k, stoch_d = calculate_stoch_rsi(closes)
    stoch_gold = (stoch_k[-1] > stoch_d[-1]) and (stoch_k[-2] <= stoch_d[-2] or stoch_k[-1] > stoch_k[-2])
    stoch_dead = (stoch_k[-1] < stoch_d[-1]) and (stoch_k[-2] >= stoch_d[-2] or stoch_k[-1] < stoch_k[-2])
    
    atr = calculate_atr(highs, lows, closes, 14)[-1]
    
    # 判定单个周期是否形成完整多/空共振
    is_long = obv_up and kdj_gold and stoch_gold
    is_short = obv_down and kdj_dead and stoch_dead
    
    signal = 1 if is_long else (-1 if is_short else 0)
    return signal, closes[-1], atr

def check_multi_timeframe_resonance(symbol):
    timeframes = ['15m', '1h', '4h', '8h']
    signals = []
    curr_price = 0
    atr_15m = 0
    
    for tf in timeframes:
        sig, price, atr = check_timeframe_signal(symbol, tf)
        signals.append(sig)
        if tf == '15m':
            curr_price = price
            atr_15m = atr

    # 15m/1h/4h/8h 四周期方向必须 100% 高度一致
    if all(s == 1 for s in signals):
        return 1, curr_price, atr_15m  # 全多共振
    elif all(s == -1 for s in signals):
        return -1, curr_price, atr_15m # 全空共振
    return 0, curr_price, atr_15m

# ================= 5. 滚雪球加仓与移动止盈止损引擎 =================
def manage_snowball_positions():
    global active_positions
    for symbol in list(active_positions.keys()):
        pos = active_positions[symbol]
        try:
            ohlcv = exchange.fetch_ohlcv(symbol, timeframe='15m', limit=5)
            curr_price = float(ohlcv[-1][4])
            
            pnl_r = ((curr_price - pos['entry_price']) / pos['stop_dist']) if pos['side'] == 'long' else ((pos['entry_price'] - curr_price) / pos['stop_dist'])

            # 1. 浮盈 1.5R 移动保本损
            if not pos['be_hit'] and pnl_r >= 1.5:
                pos['sl'] = pos['entry_price']
                pos['be_hit'] = True
                logger.info(f"🛡️ [{symbol}] 浮盈达 1.5R，止损调整至成本价 ({pos['entry_price']})")

            # 2. 浮盈 1.5R 且保本后触发滚雪球加仓
            if pos['be_hit'] and not pos['snowball_added'] and pnl_r >= SNOWBALL_TRIGGER_R:
                add_amount = float(exchange.amount_to_precision(symbol, pos['amount'] * SNOWBALL_ADD_RATIO))
                side = 'buy' if pos['side'] == 'long' else 'sell'
                
                exchange.create_order(symbol, 'market', side, add_amount)
                pos['amount'] += add_amount
                pos['snowball_added'] = True
                logger.info(f"❄️ [{symbol}] 触发【滚雪球加仓】! 加仓数量: {add_amount} | 总持仓量: {pos['amount']}")

            # 3. 浮盈 >= 2.5R 开启 1.0x ATR 移动追踪止损
            if pnl_r >= 2.5:
                atr_15m = pos['atr']
                if pos['side'] == 'long':
                    new_sl = curr_price - 1.0 * atr_15m
                    if new_sl > pos['sl']:
                        pos['sl'] = new_sl
                        logger.info(f"📈 [{symbol}] 移动追踪止损向上提升至: {pos['sl']:.4f}")
                else:
                    new_sl = curr_price + 1.0 * atr_15m
                    if new_sl < pos['sl']:
                        pos['sl'] = new_sl
                        logger.info(f"📉 [{symbol}] 移动追踪止损向下调整至: {pos['sl']:.4f}")

        except Exception as e:
            logger.error(f"持仓管理 [{symbol}] 异常: {e}")

# ================= 6. 首仓开仓（固定 3 USDT 本金 / 10X 杠杆） =================
def execute_trade(symbol, signal, price, atr_15m):
    global active_positions
    try:
        initial_notional = FIXED_MARGIN_PER_TRADE * LEVERAGE  # 30.0 USDT 名义价值
        stop_dist = 1.5 * atr_15m
        sl_price = price - stop_dist if signal == 1 else price + stop_dist

        exchange.set_leverage(LEVERAGE, symbol)
        exchange.load_markets()
        
        position_size = initial_notional / price
        amount = float(exchange.amount_to_precision(symbol, position_size))

        if amount <= 0:
            return

        side = 'buy' if signal == 1 else 'sell'
        close_side = 'sell' if signal == 1 else 'buy'

        exchange.create_order(symbol, 'market', side, amount)
        logger.info(f"🚀 [{symbol}] [OBV+KDJ+StochRSI] 4周期共振开仓 ({'做多' if signal==1 else '做空'}) | 本金: {FIXED_MARGIN_PER_TRADE}U | 杠杆: {LEVERAGE}X | 名义价值: {initial_notional}U | 止损价: {sl_price:.4f}")

        exchange.create_order(symbol, 'STOP_MARKET', close_side, amount, params={'stopPrice': sl_price, 'reduceOnly': True})

        active_positions[symbol] = {
            'side': 'long' if signal == 1 else 'short',
            'amount': amount,
            'entry_price': price,
            'sl': sl_price,
            'stop_dist': stop_dist,
            'atr': atr_15m,
            'be_hit': False,
            'snowball_added': False
        }
    except Exception as e:
        logger.error(f"开仓失败 [{symbol}]: {e}")

# ================= 7. 主循环控制 =================
def run_bot():
    manage_snowball_positions()
    
    # 保证最多同时持仓 3 个币种
    if len(active_positions) >= MAX_POSITIONS:
        return

    top15_symbols = get_top15_gainers()
    if not top15_symbols:
        return

    for symbol in top15_symbols:
        if len(active_positions) >= MAX_POSITIONS:
            break
        if symbol in active_positions:
            continue

        try:
            signal, price, atr_15m = check_multi_timeframe_resonance(symbol)
            if signal != 0:
                logger.info(f"🔥 捕获【OBV+KDJ+StochRSI】四周期全共振标的: [{symbol}] ({'做多' if signal==1 else '做空'})")
                execute_trade(symbol, signal, price, atr_15m)
        except Exception as e:
            logger.error(f"扫描标的 [{symbol}] 异常: {e}")

if __name__ == "__main__":
    logger.info("🚀 启动涨幅榜 TOP15 机器人（OBV+KDJ+StochRSI 四周期共振 / 3U本金 / 10X杠杆 / 限持仓 3 币）")
    while True:
        try:
            run_bot()
            time.sleep(15)
        except KeyboardInterrupt:
            logger.info("手动停止运行")
            break
        except Exception as e:
            logger.error(f"主循环异常重启: {e}")
            time.sleep(10)
