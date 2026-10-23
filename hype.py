import time
import hashlib
import hmac
import requests
import logging
import re
import traceback
import urllib3

# ===================== 固定配置（你的API+需求参数） =====================
API_KEY = "tj49IrEjBmedtwK3BEJNCqP1lk7fXgNkpmi5Yr2eNUhEXHG35XZzpow5TYu3S69y"
API_SECRET = "90ibaaGFYMzE6tE9IDqqnGEQZh0qZ76xReZQFt1AHteCRNEs1T76AGh6MDM3Lkno"
BASE_URL = "https://fapi.binance.com"

# 扩列：HYPE同高波动币种 12个
SYMBOLS = [
    "SOLUSDT", "BNBUSDT", "WIFUSDT", "ADAUSDT", "XRPUSDT",
    "ETHUSDT", "PEPEUSDT", "AVAXUSDT", "TRXUSDT",
    "DOGEUSDT", "ZECUSDT", "BCHUSDT", "BTCUSDT"
]
MAX_CONCURRENT_POS = 2       # 最多同时持仓2个
LEVERAGE = 5                 # 固定5倍杠杆
INIT_CAP = 50.0
MAX_DAILY_LOSS = 8.0         # 单日最大亏损8U，超限停止开仓
MAX_HYPE_USAGE = 30.0        # HYPE单笔上限30U，其余20U
USDT_PER_TRADE = 20.0
MIN_PROFIT_RATE = 0.01       # 必须≥1%利润才手动平仓，脚本绝不自动平

EMA_FAST = 20
EMA_SLOW = 50
TREND_INTERVAL = "4h"
SUB_TREND_INTERVAL = "1h"
ENTRY_INTERVAL = "1m"
ENTRY_FILTER_INTERVAL = "3m"

# 防频繁开仓/扫损参数
ORDER_COOLDOWN = 60          # 同方向60秒不能重复下单
MIN_ATR_RATIO = 0.0015       # 窄幅震荡直接禁止开仓
LOOP_INTERVAL = 12           # 12秒扫描一轮所有币种
CONFIRM_KLINE = 3            # 3根K线确认信号，过滤假突破
# ======================================================================

logging.basicConfig(
    format="%(asctime)s | %(levelname)s | %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S",
    level=logging.INFO
)
logger = logging.getLogger(__name__)
urllib3.disable_warnings()

# 全局状态
symbol_rules = {}
daily_total_loss = 0.0
last_order_time = {s: {"LONG":0, "SHORT":0} for s in SYMBOLS}
stop_today = False

def get_ts() -> int:
    return int(time.time() * 1000)

def sign(params: dict, secret: str) -> str:
    qs = "&".join([f"{k}={v}" for k, v in params.items()])
    return hmac.new(secret.encode(), qs.encode(), hashlib.sha256).hexdigest()

def request(method: str, path: str, params=None, retry=3):
    headers = {"X-MBX-APIKEY": API_KEY}
    params = params or {}
    params["timestamp"] = get_ts()
    params["signature"] = sign(params, API_SECRET)
    url = BASE_URL + path

    for i in range(retry):
        try:
            resp = requests.request(method, url, headers=headers, params=params, timeout=15, verify=False)
            resp.raise_for_status()
            return resp.json()
        except Exception as e:
            logger.error(f"API重试{i+1}: {str(e)}")
            time.sleep(2)
    logger.error("API多次请求失败")
    return None

def set_oneway():
    res = request("POST", "/fapi/v1/positionSide/dual", {"dualSidePosition": "false"})
    if res is not None and "code" not in res:
        logger.info("已设置单向持仓模式")

def set_leverage(sym: str, lev: int):
    request("POST", "/fapi/v1/leverage", {"symbol": sym, "leverage": lev})

def load_rules():
    global symbol_rules
    info = request("GET", "/fapi/v1/exchangeInfo")
    if not info:
        time.sleep(3)
        return load_rules()
    for s in info["symbols"]:
        sym = s["symbol"]
        if sym not in SYMBOLS:
            continue
        tick = step = minqty = pd = qd = None
        for f in s["filters"]:
            if f["filterType"] == "PRICE_FILTER":
                tick = float(f["tickSize"])
                pd = len(re.sub(r"0+$", "", f["tickSize"]).split(".")[1]) if "." in f["tickSize"] else 0
            if f["filterType"] == "LOT_SIZE":
                step = float(f["stepSize"])
                minqty = float(f["minQty"])
                qd = len(re.sub(r"0+$", "", f["stepSize"]).split(".")[1]) if "." in f["stepSize"] else 0
        symbol_rules[sym] = {"tick":tick, "step":step, "minqty":minqty, "pd":pd, "qd":qd}
    logger.info("交易规则加载完成 | 共{}个交易币种".format(len(SYMBOLS)))

def fmt_price(sym: str, price: float) -> float:
    r = symbol_rules[sym]
    v = round(price / r["tick"]) * r["tick"]
    return float(f"{v:.{r['pd']}f}")

def fmt_qty(sym: str, qty: float) -> float:
    r = symbol_rules[sym]
    v = int(qty / r["step"]) * r["step"]
    v = max(v, r["minqty"])
    return float(f"{v:.{r['qd']}f}")

def get_klines(sym: str, interval: str, limit=250) -> list:
    data = request("GET", "/fapi/v1/klines", {"symbol": sym, "interval": interval, "limit": limit})
    if not data:
        return []
    return [float(i[4]) for i in data]

def calc_ema(prices: list, period: int) -> float:
    k = 2 / (period + 1)
    sma = sum(prices[:period]) / period
    ema = sma
    for p in prices[period:]:
        ema = p * k + ema * (1 - k)
    return ema

def calc_atr_value(sym: str, interval: str, period=14) -> float:
    kl = request("GET", "/fapi/v1/klines", {"symbol": sym, "interval": interval, "limit": period+2})
    if not kl:
        return 0
    trs = []
    for i in range(1, len(kl)):
        h = float(kl[i][2])
        l = float(kl[i][3])
        pc = float(kl[i-1][4])
        tr = max(h-l, abs(h-pc), abs(l-pc))
        trs.append(tr)
    return sum(trs)/len(trs) if trs else 0

def get_pos(sym: str):
    data = request("GET", "/fapi/v2/positionRisk")
    if not data:
        return None, 0.0, 0.0, 0.0
    for p in data:
        if p["symbol"] == sym:
            amt = float(p["positionAmt"])
            entry = float(p["entryPrice"])
            upnl = float(p["unRealizedProfit"])
            upnl_rate = abs(upnl / (abs(amt)*entry)) if abs(amt)>1e-6 else 0
            a = abs(amt)
            if amt > 1e-9:
                return "LONG", entry, a, upnl_rate
            elif amt < -1e-9:
                return "SHORT", entry, a, upnl_rate
    return None, 0.0, 0.0, 0.0

def get_opened_count() -> int:
    cnt = 0
    for s in SYMBOLS:
        side, _, qty, _ = get_pos(s)
        if side is not None and qty > 1e-6:
            cnt += 1
    return cnt

def open_pos(sym: str, side: str):
    global stop_today, daily_total_loss
    now = time.time()
    if stop_today or daily_total_loss >= MAX_DAILY_LOSS:
        logger.warning(f"[{sym}] 当日亏损超限，禁止开仓")
        return None

    current_open = get_opened_count()
    if current_open >= MAX_CONCURRENT_POS:
        logger.warning(f"持仓已满{MAX_CONCURRENT_POS}个，跳过{sym}开仓")
        return None

    if now - last_order_time[sym][side] < ORDER_COOLDOWN:
        logger.info(f"[{sym}] {side}下单冷却中，跳过")
        return None

    use = USDT_PER_TRADE
    if sym == "HYPEUSDT" and use > MAX_HYPE_USAGE:
        use = MAX_HYPE_USAGE

    tick = request("GET", "/fapi/v1/ticker/price", {"symbol": sym})
    if not tick:
        return None
    px = float(tick["price"])
    qty = fmt_qty(sym, use / px)
    s = "BUY" if side == "LONG" else "SELL"
    params = {"symbol":sym, "side":s, "type":"MARKET", "quantity":qty}
    res = request("POST", "/fapi/v1/order", params)
    if res:
        logger.info(f"✅ [{sym}] {side}自动开仓成功 数量:{qty} | 止盈止损**请手动在APP操作**")
        logger.info(f"💡 提示：浮盈≥{MIN_PROFIT_RATE*100}%再手动平仓，不足1%继续持有")
        last_order_time[sym][side] = now
    else:
        logger.error(f"[{sym}] {side}开仓失败")
    return res

def get_big_trend(sym: str) -> str:
    c4 = get_klines(sym, TREND_INTERVAL, 200)
    c1 = get_klines(sym, SUB_TREND_INTERVAL, 200)
    if len(c4) < max(EMA_FAST,EMA_SLOW) or len(c1) < max(EMA_FAST,EMA_SLOW):
        return None

    e20_4 = calc_ema(c4, EMA_FAST)
    e50_4 = calc_ema(c4, EMA_SLOW)
    e20_1 = calc_ema(c1, EMA_FAST)
    e50_1 = calc_ema(c1, EMA_SLOW)

    up4 = e20_4 > e50_4
    up1 = e20_1 > e50_1
    if up4 and up1:
        return "LONG"
    if not up4 and not up1:
        return "SHORT"
    return None

def check_entry_signal(sym: str, big_trend: str) -> str:
    if big_trend is None:
        return None
    c1 = get_klines(sym, ENTRY_INTERVAL, 200)
    c3 = get_klines(sym, ENTRY_FILTER_INTERVAL, 200)
    min_len = max(EMA_FAST, EMA_SLOW) + CONFIRM_KLINE
    if len(c1) < min_len or len(c3) < min_len:
        return None

    up_cnt = 0
    dn_cnt = 0
    for i in range(1, CONFIRM_KLINE+1):
        s1 = c1[:-i]
        e20 = calc_ema(s1, EMA_FAST)
        e50 = calc_ema(s1, EMA_SLOW)
        if e20 > e50:
            up_cnt +=1
        elif e20 < e50:
            dn_cnt +=1

    atr = calc_atr_value(sym, ENTRY_INTERVAL, 14)
    price = c1[-1]
    atr_ratio = atr / price
    if atr_ratio < MIN_ATR_RATIO:
        logger.info(f"[{sym}] 波动率过低，禁止开仓")
        return None

    if big_trend == "LONG" and up_cnt == CONFIRM_KLINE:
        return "LONG"
    if big_trend == "SHORT" and dn_cnt == CONFIRM_KLINE:
        return "SHORT"
    return None

def main():
    load_rules()
    set_oneway()
    for s in SYMBOLS:
        set_leverage(s, LEVERAGE)
    logger.info("===== 启动：自动开仓 | 手动止盈止损 | 仅捕捉≥1%短线利润 =====")
    logger.info(f"监控币种：{len(SYMBOLS)}个HYPE同高波动标的 | 最大持仓:{MAX_CONCURRENT_POS}")

    while True:
        try:
            for sym in SYMBOLS:
                pos_side, entry, qty, upnl_rate = get_pos(sym)
                # 仅打印持仓浮盈，绝不自动平仓
                if pos_side and qty > 1e-6:
                    logger.info(f"[{sym}] 持仓:{pos_side} | 浮盈比例:{upnl_rate*100:.2f}% | ≥1%手动平仓")
                    continue
                # 无持仓：判断信号自动开仓
                big_trend = get_big_trend(sym)
                if big_trend is None:
                    continue
                sig = check_entry_signal(sym, big_trend)
                if sig:
                    logger.info(f"[{sym}] 大方向:{big_trend} 进场信号:{sig} → 执行自动开仓")
                    open_pos(sym, sig)
            time.sleep(LOOP_INTERVAL)
        except KeyboardInterrupt:
            logger.info("手动终止脚本")
            break
        except Exception as e:
            logger.error(f"主循环异常: {traceback.format_exc()}")
            time.sleep(LOOP_INTERVAL)

if __name__ == "__main__":
    main()
