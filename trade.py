import time
import hashlib
import hmac
import requests
import logging
import re
import traceback
import urllib3

# ==================== 【复盘优化后配置区 请勿随意修改】 ====================
API_KEY = "tj49IrEjBmedtwK3BEJNCqP1lk7fXgNkpmi5Yr2eNUhEXHG35XZzpow5TYu3S69y"
API_SECRET = "90ibaaGFYMzE6tE9IDqqnGEQZh0qZ76xReZQFt1AHteCRNEs1T76AGh6MDM3Lkno"
BASE_URL = "https://fapi.binance.com"

SYMBOLS = ["BTCUSDT","ETHUSDT","BNBUSDT","SOLUSDT","AVAXUSDT","XRPUSDT","ADAUSDT","MATICUSDT","TRXUSDT","ZECUSDT","DOGEUSDT","BONKUSDT","PEPEUSDT","HYPEUSDT"]
MAX_CONCURRENT_POS = 2
LEVERAGE = 5
INIT_CAP = 50.0
MAX_DAILY_LOSS = 8.0
MAX_HYPE_USAGE = 30.0
USDT_PER_TRADE = 3.8

EMA_FAST = 20
EMA_SLOW = 50

# 复盘优化核心参数
ATR_TP_MULTI = 3.0       # 折中扩大止盈，解决3.6太远拿不到利润
ATR_SL_MULTI = 1.3
ATR_TRAIL_TRIGGER = 1.6
ATR_TRAIL_MULTI = 1.7
ATR_PERIOD = 14
SL_CONFIRM_COUNT = 4     # 连续4根15m击穿止损才平仓，防插针扫损
STOP_LOSS_COOLDOWN = 3600

CONFIRM_KLINE = 3
ORDER_COOLDOWN = 60
MIN_ATR_RATIO = 0.0028   # 提高波动率门槛，过滤死水震荡
LOOP_INTERVAL = 12

TREND_INTERVAL = "4h"
SUB_TREND_INTERVAL = "1h"
ENTRY_INTERVAL = "15m"
# ==========================================================================

logging.basicConfig(
    format="%(asctime)s | %(levelname)s | %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S",
    level=logging.INFO
)
logger = logging.getLogger(__name__)
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

symbol_rules = {}
daily_total_loss = 0.0
last_order_time = {}
stop_trade_today = False
pos_tp_sl = dict()
sl_cool_down = {s: [] for s in SYMBOLS}
sl_hit_count = {s: 0 for s in SYMBOLS}


def get_timestamp():
    return int(time.time() * 1000)


def sign_params(params: dict):
    query = "&".join([f"{k}={v}" for k, v in params.items()])
    return hmac.new(
        API_SECRET.encode("utf-8"),
        query.encode("utf-8"),
        hashlib.sha256
    ).hexdigest()


def request_api(method: str, path: str, params: dict = None, retry=3):
    headers = {"X-MBX-APIKEY": API_KEY}
    url = BASE_URL + path
    params = params or {}
    params["timestamp"] = get_timestamp()
    params["signature"] = sign_params(params)

    for i in range(retry):
        try:
            if method.upper() == "GET":
                res = requests.get(url, headers=headers, params=params, timeout=15, verify=False)
            else:
                res = requests.post(url, headers=headers, data=params, timeout=15, verify=False)
            res.raise_for_status()
            return res.json()
        except Exception as e:
            logger.warning(f"请求失败 {i+1}/{retry}: {str(e)}")
            time.sleep(2)
    logger.error("API请求多次失败")
    return None


def set_oneway_mode():
    res = request_api("POST", "/fapi/v1/positionSide/dual", {"dualSidePosition": False})
    if res is None or "code" in res:
        logger.info("单向持仓已开启或无需修改")
    else:
        logger.info("已设置全局单向持仓，禁止双向持仓")


def set_leverage(sym: str, lev: int):
    return request_api("POST", "/fapi/v1/leverage", {"symbol": sym, "leverage": lev})


def load_symbol_filters():
    global symbol_rules
    info = request_api("GET", "/fapi/v1/exchangeInfo")
    if not info:
        return
    for s in info["symbols"]:
        if s["symbol"] not in SYMBOLS:
            continue
        min_qty = 0.0
        tick_size = 0.0
        for f in s["filters"]:
            if f["filterType"] == "LOT_SIZE":
                min_qty = float(f["minQty"])
            if f["filterType"] == "PRICE_FILTER":
                tick_size = float(f["tickSize"])
        symbol_rules[s["symbol"]] = {
            "min_qty": min_qty,
            "tick_size": tick_size
        }


def format_price(sym: str, price: float):
    rule = symbol_rules.get(sym)
    if not rule:
        return price
    tick = rule["tick_size"]
    return round(round(price / tick) * tick, 10)


def format_qty(sym: str, qty: float):
    rule = symbol_rules.get(sym)
    if not rule:
        return qty
    min_q = rule["min_qty"]
    q = round(qty / min_q) * min_q
    return max(q, min_q)


def get_kline(sym: str, interval: str, limit=200):
    params = {"symbol": sym, "interval": interval, "limit": limit}
    data = request_api("GET", "/fapi/v1/klines", params)
    if not data:
        return []
    closes = [float(item[4]) for item in data]
    highs = [float(item[2]) for item in data]
    lows = [float(item[3]) for item in data]
    return closes, highs, lows


def calc_ema(arr: list, length: int):
    if len(arr) < length:
        return None
    ema = [sum(arr[:length]) / length]
    alpha = 2 / (length + 1)
    for price in arr[length:]:
        ema_val = price * alpha + ema[-1] * (1 - alpha)
        ema.append(ema_val)
    return ema


def calc_atr(highs: list, lows: list, closes: list, length=14):
    tr_list = []
    for i in range(1, len(closes)):
        tr1 = highs[i] - lows[i]
        tr2 = abs(highs[i] - closes[i-1])
        tr3 = abs(lows[i] - closes[i-1])
        tr_list.append(max(tr1, tr2, tr3))
    if len(tr_list) < length:
        return None
    atr = sum(tr_list[:length]) / length
    for tr in tr_list[length:]:
        atr = (atr * (length - 1) + tr) / length
    return atr


def get_current_positions():
    pos_list = request_api("GET", "/fapi/v2/positionRisk")
    hold = {}
    if not pos_list:
        return hold
    for p in pos_list:
        sym = p["symbol"]
        amt = float(p["positionAmt"])
        if abs(amt) < 1e-8:
            continue
        side = "LONG" if amt > 0 else "SHORT"
        hold[sym] = {
            "side": side,
            "qty": abs(amt),
            "entry_price": float(p["entryPrice"])
        }
    return hold


def get_hold_count():
    return len(get_current_positions())


def check_sl_cool_down(sym: str):
    now = time.time()
    arr = sl_cool_down[sym]
    valid = [t for t in arr if now - t < STOP_LOSS_COOLDOWN]
    sl_cool_down[sym] = valid
    return len(valid) < 2


def judge_trend(sym: str):
    # 4h + 1h 大周期定多空
    c4, _, _ = get_kline(sym, TREND_INTERVAL, limit=200)
    c1, _, _ = get_kline(sym, SUB_TREND_INTERVAL, limit=200)
    if not c4 or not c1:
        return None

    e4_f = calc_ema(c4, EMA_FAST)
    e4_s = calc_ema(c4, EMA_SLOW)
    e1_f = calc_ema(c1, EMA_FAST)
    e1_s = calc_ema(c1, EMA_SLOW)
    if not e4_f or not e4_s or not e1_f or not e1_s:
        return None

    long_ok = (e4_f[-1] > e4_s[-1]) and (e1_f[-1] > e1_s[-1])
    short_ok = (e4_f[-1] < e4_s[-1]) and (e1_f[-1] < e1_s[-1])
    if long_ok:
        return "LONG"
    elif short_ok:
        return "SHORT"
    else:
        return None


def check_entry_signal(sym: str, trend_side: str):
    if trend_side is None:
        return False
    closes, highs, lows = get_kline(sym, ENTRY_INTERVAL, limit=100)
    if len(closes) < CONFIRM_KLINE + EMA_FAST:
        return False

    atr = calc_atr(highs, lows, closes, ATR_PERIOD)
    if atr is None:
        return False
    price = closes[-1]
    if atr / price < MIN_ATR_RATIO:
        return False

    ema_f = calc_ema(closes, EMA_FAST)
    if not ema_f:
        return False

    # 连续N根K线同向确认
    if trend_side == "LONG":
        for i in range(1, CONFIRM_KLINE+1):
            idx = -i
            if closes[idx] < ema_f[idx]:
                return False
    else:
        for i in range(1, CONFIRM_KLINE+1):
            idx = -i
            if closes[idx] > ema_f[idx]:
                return False

    return True


def open_order(sym: str, side: str):
    global daily_total_loss, stop_trade_today
    if stop_trade_today:
        logger.info("已触发单日最大亏损，停止开仓")
        return
    if get_hold_count() >= MAX_CONCURRENT_POS:
        logger.info("同时持仓达到上限，不开仓")
        return
    if not check_sl_cool_down(sym):
        logger.info(f"{sym} 止损冷却中，禁止开仓")
        return

    hold = get_current_positions()
    if sym in hold:
        logger.info(f"{sym} 已有持仓，禁止双向加仓")
        return

    closes, highs, lows = get_kline(sym, ENTRY_INTERVAL, limit=200)
    atr = calc_atr(highs, lows, closes, ATR_PERIOD)
    if atr is None:
        return
    mark = closes[-1]

    # 计算下单数量
    use_usdt = MAX_HYPE_USAGE if sym == "HYPEUSDT" else USDT_PER_TRADE
    qty_raw = (use_usdt * LEVERAGE) / mark
    qty = format_qty(sym, qty_raw)
    if qty <= 0:
        return

    side_api = "BUY" if side == "LONG" else "SELL"
    params = {
        "symbol": sym,
        "side": side_api,
        "type": "MARKET",
        "quantity": qty
    }
    res = request_api("POST", "/fapi/v1/order", params)
    if res and "orderId" in res:
        logger.info(f"开仓成功 {sym} {side} 数量:{qty}")
        # 保存止盈止损基准
        if side == "LONG":
            tp = format_price(sym, mark + ATR_TP_MULTI * atr)
            sl = format_price(sym, mark - ATR_SL_MULTI * atr)
        else:
            tp = format_price(sym, mark - ATR_TP_MULTI * atr)
            sl = format_price(sym, mark + ATR_SL_MULTI * atr)
        pos_tp_sl[sym] = {
            "side": side,
            "entry": mark,
            "tp": tp,
            "trail_sl": sl,
            "atr": atr,
            "qty": qty
        }
        sl_hit_count[sym] = 0
    else:
        logger.error(f"{sym} 开仓失败: {res}")


def close_order(sym: str, side: str, qty: float, loss_val=0.0):
    global daily_total_loss, stop_trade_today
    side_api = "SELL" if side == "LONG" else "BUY"
    qty_fmt = format_qty(sym, qty)
    params = {
        "symbol": sym,
        "side": side_api,
        "type": "MARKET",
        "quantity": qty_fmt
    }
    res = request_api("POST", "/fapi/v1/order", params)
    if res and "orderId" in res:
        logger.info(f"平仓完成 {sym} {side}")
        if loss_val > 0:
            daily_total_loss += loss_val
            sl_cool_down[sym].append(time.time())
            logger.warning(f"当日累计亏损:{daily_total_loss:.2f}U")
            if daily_total_loss >= MAX_DAILY_LOSS:
                stop_trade_today = True
                logger.warning("触发单日最大亏损，今日停止所有开仓")
        if sym in pos_tp_sl:
            del pos_tp_sl[sym]
        sl_hit_count[sym] = 0


def monitor_positions():
    hold = get_current_positions()
    for sym, info in hold.items():
        if sym not in pos_tp_sl:
            continue
        pos_side = info["side"]
        entry = info["entry_price"]
        qty = info["qty"]
        d = pos_tp_sl[sym]
        tp = d["tp"]
        trail_sl = d["trail_sl"]
        atr = d["atr"]
        entry_p = d["entry"]

        idx_data = request_api("GET", "/fapi/v1/premiumIndex", {"symbol": sym})
        if not idx_data:
            continue
        mark = float(idx_data["markPrice"])
        trigger = False
        hit_sl = False

        # 浮盈ATR倍数
        if pos_side == "LONG":
            profit_atr = (mark - entry_p) / atr
        else:
            profit_atr = (entry_p - mark) / atr

        new_trail = trail_sl
        # 1、浮盈达标先保本
        if profit_atr >= ATR_TRAIL_TRIGGER:
            if pos_side == "LONG":
                new_trail = max(new_trail, entry_p)
            else:
                new_trail = min(new_trail, entry_p)
        # 2、进一步扩大浮盈，动态追踪止损
        if profit_atr >= ATR_TP_MULTI * 0.8:
            if pos_side == "LONG":
                follow = mark - ATR_TRAIL_MULTI * atr
                new_trail = max(new_trail, follow)
            else:
                follow = mark + ATR_TRAIL_MULTI * atr
                new_trail = min(new_trail, follow)
        if abs(new_trail - trail_sl) > 1e-6:
            d["trail_sl"] = format_price(sym, new_trail)
            trail_sl = d["trail_sl"]
            logger.info(f"[{sym}] 追踪止损更新:{trail_sl:.4f}")

        # 判断止盈/击穿止损
        if pos_side == "LONG":
            if mark >= tp:
                trigger = True
            elif mark <= trail_sl:
                hit_sl = True
        else:
            if mark <= tp:
                trigger = True
            elif mark >= trail_sl:
                hit_sl = True

        if hit_sl:
            sl_hit_count[sym] += 1
            logger.info(f"[{sym}] 止损击穿计数:{sl_hit_count[sym]}/{SL_CONFIRM_COUNT}")
            if sl_hit_count[sym] >= SL_CONFIRM_COUNT:
                loss = abs(entry_p - mark) * qty
                trigger = True
        else:
            sl_hit_count[sym] = 0

        if trigger:
            close_order(sym, pos_side, qty, loss_val=loss if hit_sl else 0)


def main_loop():
    set_oneway_mode()
    load_symbol_filters()
    for s in SYMBOLS:
        set_leverage(s, LEVERAGE)
    logger.info("===== 复盘优化版启动 | 4h+1h定方向 15m进场 | 最多2仓 | 单向持仓 =====")
    while True:
        try:
            monitor_positions()
            if not stop_trade_today:
                for sym in SYMBOLS:
                    trend = judge_trend(sym)
                    if trend is None:
                        continue
                    if check_entry_signal(sym, trend):
                        open_order(sym, trend)
            time.sleep(LOOP_INTERVAL)
        except Exception as e:
            logger.error(f"主循环异常:{traceback.format_exc()}")
            time.sleep(5)


if __name__ == "__main__":
    main_loop()
