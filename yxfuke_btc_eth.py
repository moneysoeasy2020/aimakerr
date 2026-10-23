import time
import logging
import traceback
import ccxt
import numpy as np

# 日志配置
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - [%(levelname)s] - %(message)s'
)
logger = logging.getLogger("BTC_ETH_Trend_Pro")

# ================= 1. API 配置 =================
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
        'recvWindow': 10000,
    }
})

# ================= 2. BTC/ETH 专属策略参数 =================
STRATEGY_CONFIG = {
    "TARGET_SYMBOLS": ["BTC/USDT", "ETH/USDT"], # 写死专属币池
    "SINGLE_COIN_MAX_MARGIN_USDT": 20,           # 单币种最大保证金 (提升至 20U)
    "BASE_LEVERAGE": 15,                         # 主流币基准杠杆 (15x)
    "REDUCED_LEVERAGE": 7,                      # 风控降杠杆 (7x)
    "MAX_DRAWDOWN_TRIGGER": 0.25,               # 账户权益回撤 25% 触发降杠杆
    "MAX_OPEN_POSITIONS": 2,                    # 最多同时持仓 BTC 和 ETH
    "TRADE_COOLDOWN_SECONDS": 21600,            # 正常建仓/加仓冷却 (6小时)
    "STOP_LOSS_COOLDOWN_SECONDS": 14400,        # 止损后强制冷却 (4小时)
    "MAX_SPOT_SL_PCT": 0.025                    # 主流币现货硬止损上限 (2.5%)
}

peak_equity = 0.0
pyramid_tracker = {}
last_trade_time = {}
stop_loss_cooldown = {}

def calculate_ema_np(arr, period):
    if len(arr) < period:
        return np.array([])
    alpha = 2.0 / (period + 1.0)
    ema = np.empty_like(arr)
    ema[0] = arr[0]
    for i in range(1, len(arr)):
        ema[i] = alpha * arr[i] + (1 - alpha) * ema[i - 1]
    return ema

def safe_api_call(func, *args, **kwargs):
    try:
        return func(*args, **kwargs)
    except Exception as e:
        logger.error(f"API 调用异常 [{func.__name__}]: {e}")
        return None

def get_account_equity():
    balance = safe_api_call(exchange.fetch_balance)
    if balance and 'total' in balance and 'USDT' in balance['total']:
        return float(balance['total']['USDT'])
    return 0.0

def get_current_leverage(current_equity):
    global peak_equity
    if current_equity > peak_equity:
        peak_equity = current_equity
    
    if peak_equity > 0:
        drawdown = (peak_equity - current_equity) / peak_equity
        if drawdown >= STRATEGY_CONFIG["MAX_DRAWDOWN_TRIGGER"]:
            logger.warning(f"⚠️ 账户回撤达到 {drawdown*100:.1f}% >= 25%，强风控降杠杆至 7x")
            return STRATEGY_CONFIG["REDUCED_LEVERAGE"]
    
    return STRATEGY_CONFIG["BASE_LEVERAGE"]

def get_active_positions():
    positions = safe_api_call(exchange.fetch_positions)
    if not positions:
        return {}
    
    active = {}
    for pos in positions:
        contracts = float(pos.get('contracts', 0) or pos.get('positionAmt', 0))
        if abs(contracts) > 0 and pos['symbol'] in STRATEGY_CONFIG["TARGET_SYMBOLS"]:
            symbol = pos['symbol']
            side = 'BUY' if contracts > 0 else 'SELL'
            active[symbol] = {
                "contracts": abs(contracts),
                "side": side,
                "entry_price": float(pos.get('entryPrice', 0)),
                "unrealized_pnl": float(pos.get('unrealizedPnl', 0)),
                "initial_margin": float(pos.get('initialMargin', 0))
            }
    return active

def check_mtf_signals(symbol):
    # 4h 趋势方向判定
    ohlcv_4h = safe_api_call(exchange.fetch_ohlcv, symbol, timeframe='4h', limit=60)
    if not ohlcv_4h or len(ohlcv_4h) < 55:
        return None, 0, 0
    
    closes_4h = np.array(ohlcv_4h[:-1], dtype=float)[:, 4]
    ema20_4h = calculate_ema_np(closes_4h, 20)
    ema50_4h = calculate_ema_np(closes_4h, 50)
    
    if len(ema20_4h) == 0 or len(ema50_4h) == 0:
        return None, 0, 0

    bull_trend_4h = ema20_4h[-1] > ema50_4h[-1]
    bear_trend_4h = ema20_4h[-1] < ema50_4h[-1]

    # 1h 突破与量能校验
    ohlcv_1h = safe_api_call(exchange.fetch_ohlcv, symbol, timeframe='1h', limit=60)
    if not ohlcv_1h or len(ohlcv_1h) < 30:
        return None, 0, 0
    
    closed_1h = ohlcv_1h[:-1]
    closes_1h = np.array(closed_1h, dtype=float)[:, 4]
    highs_1h = np.array(closed_1h, dtype=float)[:, 2]
    lows_1h = np.array(closed_1h, dtype=float)[:, 3]
    vols_1h = np.array(closed_1h, dtype=float)[:, 5]

    cur_price = closes_1h[-1]
    vol_ma5 = np.mean(vols_1h[-6:-1])
    vol_confirmed = vols_1h[-1] > (vol_ma5 * 1.3) # 主流币放量阈值适度放宽至 1.3 倍

    recent_high_1h = np.max(highs_1h[-21:-1])
    recent_low_1h = np.min(lows_1h[-21:-1])

    breakout_long = (cur_price > recent_high_1h) and vol_confirmed
    breakout_short = (cur_price < recent_low_1h) and vol_confirmed

    # 结构化止损与 2.5% 硬止损结合
    struct_sl_long = np.min(lows_1h[-5:-1])
    struct_sl_short = np.max(highs_1h[-5:-1])

    max_sl_pct = STRATEGY_CONFIG["MAX_SPOT_SL_PCT"]
    sl_long_hard = cur_price * (1.0 - max_sl_pct)
    sl_short_hard = cur_price * (1.0 + max_sl_pct)

    final_sl_long = max(struct_sl_long, sl_long_hard)
    final_sl_short = min(struct_sl_short, sl_short_hard)

    if bull_trend_4h and breakout_long:
        return 'BUY', cur_price, final_sl_long
    elif bear_trend_4h and breakout_short:
        return 'SELL', cur_price, final_sl_short

    return None, cur_price, 0

def is_in_cooldown(symbol):
    now = time.time()
    
    last_sl_time = stop_loss_cooldown.get(symbol, 0)
    sl_cd = STRATEGY_CONFIG["STOP_LOSS_COOLDOWN_SECONDS"]
    if now - last_sl_time < sl_cd:
        rem_hrs = (sl_cd - (now - last_sl_time)) / 3600.0
        return True, f"止损强行冷却中 (还剩 {rem_hrs:.1f} 小时)"

    last_time = last_trade_time.get(symbol, 0)
    trade_cd = STRATEGY_CONFIG["TRADE_COOLDOWN_SECONDS"]
    if now - last_time < trade_cd:
        rem_hrs = (trade_cd - (now - last_time)) / 3600.0
        return True, f"交易冷却中 (还剩 {rem_hrs:.1f} 小时)"

    return False, ""

def process_open_and_pyramiding(symbol, active_positions, equity):
    try:
        in_cd, cd_reason = is_in_cooldown(symbol)
        if in_cd:
            return

        signal, cur_price, final_sl = check_mtf_signals(symbol)
        if cur_price <= 0:
            return

        leverage = get_current_leverage(equity)
        safe_api_call(exchange.set_leverage, leverage, symbol)

        max_margin = STRATEGY_CONFIG["SINGLE_COIN_MAX_MARGIN_USDT"]

        # 1. 初始建仓 (Phase 1)
        if symbol not in active_positions:
            if not signal:
                return

            margin_p1 = max_margin * 0.35 # 建仓占比 35%
            notional_p1 = margin_p1 * leverage
            
            final_notional = max(notional_p1, 5.5)
            raw_amount = final_notional / cur_price
            
            amount_str = exchange.amount_to_precision(symbol, raw_amount)
            amount = float(amount_str) if amount_str else 0.0
            if amount <= 0:
                return

            order_side = 'buy' if signal == 'BUY' else 'sell'
            exit_side = 'sell' if signal == 'BUY' else 'buy'

            logger.info(f"⚡ [BTC/ETH 建仓 Phase 1] [{symbol}] 方向: {signal} | 保证金: {margin_p1:.1f} USDT | 入场价: {cur_price} | 止损价: {final_sl:.2f}")
            order = safe_api_call(exchange.create_order, symbol=symbol, type='market', side=order_side, amount=amount)

            if order:
                last_trade_time[symbol] = time.time()
                if final_sl > 0:
                    sl_price = float(exchange.price_to_precision(symbol, final_sl))
                    safe_api_call(
                        exchange.create_order,
                        symbol=symbol, type='STOP_MARKET', side=exit_side, amount=amount,
                        params={'stopPrice': sl_price, 'reduceOnly': True}
                    )

                pyramid_tracker[symbol] = {
                    "phase": 1,
                    "entry_price": cur_price,
                    "side": signal,
                    "total_amount": amount,
                    "breakeven_set": False
                }
            return

        # 2. 滚雪球加仓 (Phase 2 & Phase 3)
        pos_info = active_positions[symbol]
        track_info = pyramid_tracker.get(symbol)
        if not track_info:
            pyramid_tracker[symbol] = {
                "phase": 1,
                "entry_price": pos_info["entry_price"],
                "side": pos_info["side"],
                "total_amount": pos_info["contracts"],
                "breakeven_set": False
            }
            track_info = pyramid_tracker[symbol]

        side = track_info["side"]
        entry_p = track_info["entry_price"]
        phase = track_info["phase"]

        pnl_pct = (cur_price - entry_p) / entry_p if side == 'BUY' else (entry_p - cur_price) / entry_p

        # 主流币优化：加仓门槛调整为 +1.8% 与 +3.5%
        if phase == 1 and pnl_pct >= 0.018:
            margin_p2 = max_margin * 0.35
            notional_p2 = max(margin_p2 * leverage, 5.5)

            add_amount = float(exchange.amount_to_precision(symbol, notional_p2 / cur_price))
            if add_amount > 0:
                order_side = 'buy' if side == 'BUY' else 'sell'
                logger.info(f"📈 [BTC/ETH 加仓 Phase 2 (加仓35%)] [{symbol}] 现货拉升 {pnl_pct*100:.2f}% | 加仓保证金: {margin_p2:.1f} USDT")
                order = safe_api_call(exchange.create_order, symbol=symbol, type='market', side=order_side, amount=add_amount)
                if order:
                    last_trade_time[symbol] = time.time()
                    track_info["phase"] = 2
                    track_info["total_amount"] += add_amount

        elif phase == 2 and pnl_pct >= 0.035:
            margin_p3 = max_margin * 0.30
            notional_p3 = max(margin_p3 * leverage, 5.5)

            add_amount = float(exchange.amount_to_precision(symbol, notional_p3 / cur_price))
            if add_amount > 0:
                order_side = 'buy' if side == 'BUY' else 'sell'
                logger.info(f"🚀 [BTC/ETH 加仓 Phase 3 (加仓30%)] [{symbol}] 现货拉升 {pnl_pct*100:.2f}% | 加仓保证金: {margin_p3:.1f} USDT")
                order = safe_api_call(exchange.create_order, symbol=symbol, type='market', side=order_side, amount=add_amount)
                if order:
                    last_trade_time[symbol] = time.time()
                    track_info["phase"] = 3
                    track_info["total_amount"] += add_amount
    except Exception as e:
        logger.error(f"[{symbol}] 下单/加仓逻辑处理异常: {e}")

def manage_exit_and_stops(active_positions):
    for symbol, pos in list(active_positions.items()):
        try:
            contracts = pos["contracts"]
            side = pos["side"]
            entry_price = pos["entry_price"]

            ohlcv_1h = safe_api_call(exchange.fetch_ohlcv, symbol, timeframe='1h', limit=55)
            ohlcv_1d = safe_api_call(exchange.fetch_ohlcv, symbol, timeframe='1d', limit=55)
            if not ohlcv_1h or not ohlcv_1d or len(ohlcv_1h) < 50 or len(ohlcv_1d) < 25:
                continue

            closes_1h = np.array(ohlcv_1h[:-1], dtype=float)[:, 4]
            closes_1d = np.array(ohlcv_1d[:-1], dtype=float)[:, 4]
            cur_price = closes_1h[-1]

            pnl_pct = (cur_price - entry_price) / entry_price if side == 'BUY' else (entry_price - cur_price) / entry_price
            track_info = pyramid_tracker.get(symbol)

            # 1. 主流币浮盈 >= 2.5% 自动移动保本止损
            if pnl_pct >= 0.025 and track_info and not track_info["breakeven_set"]:
                exit_side = 'sell' if side == 'BUY' else 'buy'
                be_price = float(exchange.price_to_precision(symbol, entry_price))
                formatted_amount = float(exchange.amount_to_precision(symbol, contracts))

                params = {'stopPrice': be_price, 'reduceOnly': True}
                new_stop_order = safe_api_call(
                    exchange.create_order,
                    symbol=symbol, type='STOP_MARKET', side=exit_side, amount=formatted_amount, params=params
                )

                if new_stop_order:
                    logger.info(f"🛡️ [{symbol}] 现货浮盈达 {pnl_pct*100:.2f}% >= 2.5%，已自动设置保本止损线: {be_price}")
                    track_info["breakeven_set"] = True
                    
                    open_orders = safe_api_call(exchange.fetch_open_orders, symbol)
                    if open_orders:
                        for order in open_orders:
                            if order['id'] != new_stop_order['id'] and order.get('type') in ['stop_market', 'STOP_MARKET']:
                                safe_api_call(exchange.cancel_order, order['id'], symbol)

            # 2. 1h EMA20/50 交叉 -> 平仓 50%
            ema20_1h = calculate_ema_np(closes_1h, 20)
            ema50_1h = calculate_ema_np(closes_1h, 50)
            
            if len(ema20_1h) >= 2 and len(ema50_1h) >= 2:
                cross_50_close = False
                if side == 'BUY' and (ema20_1h[-2] >= ema50_1h[-2] and ema20_1h[-1] < ema50_1h[-1]):
                    cross_50_close = True
                elif side == 'SELL' and (ema20_1h[-2] <= ema50_1h[-2] and ema20_1h[-1] > ema50_1h[-1]):
                    cross_50_close = True

                if cross_50_close:
                    close_amount = float(exchange.amount_to_precision(symbol, contracts * 0.5))
                    exit_side = 'sell' if side == 'BUY' else 'buy'
                    logger.info(f"⚠️ [{symbol}] 1h 出现 EMA 反向死叉，阶段性锁利平仓 50%: {close_amount}")
                    order = safe_api_call(
                        exchange.create_order,
                        symbol=symbol, type='market', side=exit_side, amount=close_amount,
                        params={'reduceOnly': True}
                    )
                    if order:
                        last_trade_time[symbol] = time.time()

            # 3. 1D 跌破 EMA20 -> 全平出场
            ema20_1d = calculate_ema_np(closes_1d, 20)
            if len(ema20_1d) >= 1:
                d1_break = (side == 'BUY' and closes_1d[-1] < ema20_1d[-1]) or (side == 'SELL' and closes_1d[-1] > ema20_1d[-1])

                if d1_break:
                    exit_side = 'sell' if side == 'BUY' else 'buy'
                    formatted_contracts = float(exchange.amount_to_precision(symbol, contracts))
                    logger.info(f"🚨 [{symbol}] 1D 跌破 EMA20 主趋势线，全平清仓!")
                    
                    order = safe_api_call(
                        exchange.create_order,
                        symbol=symbol, type='market', side=exit_side, amount=formatted_contracts,
                        params={'reduceOnly': True}
                    )
                    
                    if order:
                        last_trade_time[symbol] = time.time()
                        safe_api_call(exchange.cancel_all_orders, symbol)
                        if symbol in pyramid_tracker:
                            del pyramid_tracker[symbol]

        except Exception as e:
            logger.error(f"[{symbol}] 巡检平仓逻辑异常: {e}")

def audit_stop_loss_triggers(active_positions):
    for symbol in list(pyramid_tracker.keys()):
        if symbol not in active_positions:
            logger.warning(f"🛑 [{symbol}] 检测到持仓已平（触发止损或挂单），进入 4 小时冷却时间！")
            stop_loss_cooldown[symbol] = time.time()
            del pyramid_tracker[symbol]

def run_loop():
    logger.info("⚙️ BTC/ETH 专属趋势突破策略 (15x 专属定速巡航) 启动...")
    safe_api_call(exchange.load_markets)
    
    while True:
        try:
            equity = get_account_equity()
            active_positions = get_active_positions()
            
            audit_stop_loss_triggers(active_positions)
            manage_exit_and_stops(active_positions)

            for symbol in STRATEGY_CONFIG["TARGET_SYMBOLS"]:
                process_open_and_pyramiding(symbol, active_positions, equity)

            time.sleep(10)
        except Exception as e:
            logger.error(f"主循环异常拦截: {e}\n{traceback.format_exc()}")
            time.sleep(10)

if __name__ == "__main__":
    run_loop()
