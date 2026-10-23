import requests
import time
import hmac
import hashlib
import logging
import traceback
from urllib.parse import urlencode

# ===================== 全局固化配置（禁止修改密钥） =====================
API_KEY = "tj49IrEjBmedtwK3BEJNCqP1lk7fXgNkpmi5Yr2eNUhEXHG35XZzpow5TYu3S69y"
API_SECRET = "90ibaaGFYMzE6tE9IDqqnGEQZh0qZ76xReZQFt1AHteCRNEs1T76AGh6MDM3Lkno"
BASE_URL = "https://fapi.binance.com"
HEADERS = {"X-MBX-APIKEY": API_KEY}

# 资金&杠杆参数
INIT_BALANCE = 50.0       # 初始本金50USDT
LEVERAGE = 5              # 固定5倍杠杆
POSITION_RATIO = 0.05     # 单笔头寸占用本金5%
MAX_SINGLE_LOSS = 2.8     # 单笔最大亏损2.8USDT
TP_RISK_RATIO = 2.5       # 标准盈亏比2.5:1
TARGET_PROFIT = MAX_SINGLE_LOSS * TP_RISK_RATIO  # 7USDT止盈

# 持仓风控
MAX_HOLD_COINS = 3        # 最多同时持有3个币种
SYMBOL_LIST = ["HYPEUSDT"]# 交易标的池，可自行添加如BTCUSDT、ETHUSDT

# K线周期参数
TREND_INTERVAL = "15m"    # 趋势周期15分钟
SIGNAL_INTERVAL = "1m"    # 信号周期1分钟
EMA_PERIOD = 20
ATR_PERIOD = 20
VOL_MA_PERIOD = 5

# 插针判定阈值
SHADOW_MULTIPLE = 2.2     # 影线≥实体2.2倍判定长引线

# 日志配置
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S"
)
logger = logging.getLogger("BinanceHypeBot")

# 全局缓存
hold_positions = {}  # 存储持仓 {symbol: side}
tick_size_cache = {}
step_size_cache = {}

# ===================== API签名工具 =====================
def get_signature(params):
    query_string = urlencode(params)
    return hmac.new(API_SECRET.encode(), query_string.encode(), hashlib.sha256).hexdigest()

def send_request(method, path, params=None):
    if params is None:
        params = {}
    params["timestamp"] = int(time.time() * 1000)
    params["signature"] = get_signature(params)
    url = f"{BASE_URL}{path}"
    retry = 3
    while retry > 0:
        try:
            if method == "GET":
                res = requests.get(url, headers=HEADERS, params=params, timeout=10)
            else:
                res = requests.post(url, headers=HEADERS, data=params, timeout=10)
            res_json = res.json()
            if res.status_code == 418:
                logger.warning("API限流，休眠2s")
                time.sleep(2)
                retry -= 1
                continue
            if res.status_code != 200:
                logger.error(f"API错误 {res.status_code}: {res_json}")
                return None
            return res_json
        except Exception as e:
            logger.error(f"网络请求异常: {str(e)}")
            retry -= 1
            time.sleep(1)
    logger.error("请求多次失败，跳过本轮")
    return None

# ===================== 合约规格获取（精度适配） =====================
def get_symbol_spec(symbol):
    if symbol in tick_size_cache:
        return tick_size_cache[symbol], step_size_cache[symbol]
    data = send_request("GET", "/fapi/v1/exchangeInfo")
    if not data:
        return 0.01, 0.001
    for s in data["symbols"]:
        if s["symbol"] == symbol:
            tick = float([f for f in s["filters"] if f["filterType"] == "PRICE_FILTER"][0]["tickSize"])
            step = float([f for f in s["filters"] if f["filterType"] == "LOT_SIZE"][0]["stepSize"])
            tick_size_cache[symbol] = tick
            step_size_cache[symbol] = step
            return tick, step
    return 0.01, 0.001

def format_num(num, step):
    dec = len(str(step).split(".")[-1].rstrip("0"))
    return round(num, dec)

# ===================== 账户/持仓查询 =====================
def get_wallet_balance():
    data = send_request("GET", "/fapi/v2/balance")
    if not data:
        return 0.0
    for item in data:
        if item["asset"] == "USDT":
            return float(item["balance"])
    return 0.0

def get_all_positions():
    pos_map = {}
    data = send_request("GET", "/fapi/v2/positionRisk")
    if not data:
        return pos_map
    for pos in data:
        sym = pos["symbol"]
        amt = float(pos["positionAmt"])
        if amt == 0:
            continue
        side = "LONG" if amt > 0 else "SHORT"
        pos_map[sym] = {
            "side": side,
            "amt": abs(amt),
            "entryPrice": float(pos["entryPrice"]),
            "markPrice": float(pos["markPrice"]),
            "unRealizedProfit": float(pos["unRealizedProfit"])
        }
    return pos_map

# ===================== K线、指标计算 =====================
def get_kline(symbol, interval, limit=50):
    params = {"symbol": symbol, "interval": interval, "limit": limit}
    data = send_request("GET", "/fapi/v1/klines", params)
    if not data:
        return []
    kline_list = []
    for k in data:
        kline_list.append({
            "open": float(k[1]),
            "high": float(k[2]),
            "low": float(k[3]),
            "close": float(k[4]),
            "volume": float(k[5])
        })
    return kline_list

def calc_ema(price_list, period):
    ema = []
    mul = 2 / (period + 1)
    sma = sum(price_list[:period]) / period
    ema.append(sma)
    for p in price_list[period:]:
        new_ema = p * mul + ema[-1] * (1 - mul)
        ema.append(new_ema)
    return ema

def calc_atr(kline_list, period):
    tr_list = []
    for i in range(1, len(kline_list)):
        high = kline_list[i]["high"]
        low = kline_list[i]["low"]
        prev_close = kline_list[i-1]["close"]
        tr = max(high - low, abs(high - prev_close), abs(low - prev_close))
        tr_list.append(tr)
    atr = calc_ema(tr_list, period)
    return atr[-1] if atr else 0

def calc_rsi(price_list, period=14):
    gain = []
    loss = []
    for i in range(1, len(price_list)):
        diff = price_list[i] - price_list[i-1]
        if diff > 0:
            gain.append(diff)
            loss.append(0)
        else:
            gain.append(0)
            loss.append(-diff)
    avg_gain = sum(gain[:period]) / period
    avg_loss = sum(loss[:period]) / period
    rs = avg_gain / avg_loss if avg_loss != 0 else 999
    rsi = 100 - (100 / (1 + rs))
    return rsi

# 判断长引线插针
def check_long_shadow(k):
    o, h, l, c = k["open"], k["high"], k["low"], k["close"]
    body = abs(c - o)
    if body < 1e-8:
        return False, ""
    upper_shadow = h - max(o, c)
    lower_shadow = min(o, c) - l
    # 多头下影插针
    if lower_shadow >= body * SHADOW_MULTIPLE and c > (o + h) / 2:
        return True, "LONG_PIN"
    # 空头上影插针
    if upper_shadow >= body * SHADOW_MULTIPLE and c < (o + l) / 2:
        return True, "SHORT_PIN"
    return False, ""

# 多指标共振校验，需要满足2项
def check_resonance(k_lines, pin_type, atr_val, atr_mean):
    close_prices = [x["close"] for x in k_lines]
    ema20 = calc_ema(close_prices, EMA_PERIOD)[-1]
    rsi = calc_rsi(close_prices)
    vol_avg = sum([x["volume"] for x in k_lines[-VOL_MA_PERIOD:]]) / VOL_MA_PERIOD
    curr_vol = k_lines[-1]["volume"]
    hit = 0

    if pin_type == "LONG_PIN":
        # 多头共振条件
        if close_prices[-1] > ema20: hit +=1
        if rsi < 40: hit +=1
        if atr_val > atr_mean: hit +=1
        if curr_vol > vol_avg: hit +=1
    else:
        # 空头共振条件
        if close_prices[-1] < ema20: hit +=1
        if rsi > 60: hit +=1
        if atr_val > atr_mean: hit +=1
        if curr_vol > vol_avg: hit +=1
    return hit >= 2

# ===================== 趋势判断(15min) =====================
def get_trend_direction(symbol):
    klines = get_kline(symbol, TREND_INTERVAL, limit=EMA_PERIOD+10)
    if len(klines) < EMA_PERIOD:
        return None
    close = [k["close"] for k in klines]
    ema20 = calc_ema(close, EMA_PERIOD)[-1]
    last_price = close[-1]
    if last_price > ema20:
        return "ALLOW_LONG"
    else:
        return "ALLOW_SHORT"

# ===================== 仓位计算 =====================
def calc_position_size(symbol, entry_price, stop_price):
    balance = get_wallet_balance()
    margin = balance * POSITION_RATIO
    notional = margin * LEVERAGE
    price_diff = abs(entry_price - stop_price)
    if price_diff < 1e-8:
        return 0
    qty = MAX_SINGLE_LOSS / price_diff
    max_qty = notional / entry_price
    final_qty = min(qty, max_qty)
    _, step = get_symbol_spec(symbol)
    return format_num(final_qty, step)

# ===================== 开平仓函数 =====================
def set_leverage(symbol):
    params = {"symbol": symbol, "leverage": LEVERAGE}
    send_request("POST", "/fapi/v1/leverage", params)

def open_order(symbol, side, qty):
    set_leverage(symbol)
    side_api = "BUY" if side == "LONG" else "SELL"
    params = {
        "symbol": symbol,
        "side": side_api,
        "type": "MARKET",
        "quantity": qty
    }
    res = send_request("POST", "/fapi/v1/order", params)
    if res:
        logger.info(f"开仓成功 {symbol} {side} 数量:{qty}")
        return True
    return False

def close_all_pos(symbol, pos_data):
    side_api = "SELL" if pos_data["side"] == "LONG" else "BUY"
    params = {
        "symbol": symbol,
        "side": side_api,
        "type": "MARKET",
        "quantity": pos_data["amt"]
    }
    send_request("POST", "/fapi/v1/order", params)
    logger.info(f"全平 {symbol} {pos_data['side']}")

# ===================== 轮询止盈止损、移动止盈逻辑 =====================
def monitor_tp_sl(pos_map):
    for sym, pos in pos_map.items():
        entry = pos["entryPrice"]
        mark = pos["markPrice"]
        profit = pos["unRealizedProfit"]
        tick, _ = get_symbol_spec(sym)
        # 硬性止损
        if profit <= -MAX_SINGLE_LOSS:
            close_all_pos(sym, pos)
            continue
        # 止盈目标
        if profit >= TARGET_PROFIT:
            close_all_pos(sym, pos)
            continue
        # 移动止盈：浮盈≥2.8U，止损移至开仓价保本
        if profit >= MAX_SINGLE_LOSS:
            if pos["side"] == "LONG" and mark <= entry:
                close_all_pos(sym, pos)
            if pos["side"] == "SHORT" and mark >= entry:
                close_all_pos(sym, pos)
        # 反向插针提前平仓避险
        k1 = get_kline(sym, SIGNAL_INTERVAL, limit=30)
        if len(k1) < 10:
            continue
        has_pin, pin_t = check_long_shadow(k1[-1])
        if has_pin:
            if pos["side"] == "LONG" and pin_t == "SHORT_PIN":
                if check_resonance(k1, pin_t, calc_atr(k1, ATR_PERIOD), calc_atr(k1[:-1], ATR_PERIOD)):
                    logger.info(f"{sym} 反向插针共振，提前平仓避险")
                    close_all_pos(sym, pos)
            if pos["side"] == "SHORT" and pin_t == "LONG_PIN":
                if check_resonance(k1, pin_t, calc_atr(k1, ATR_PERIOD), calc_atr(k1[:-1], ATR_PERIOD)):
                    logger.info(f"{sym} 反向插针共振，提前平仓避险")
                    close_all_pos(sym, pos)

# ===================== 主交易循环 =====================
def main_trade_loop():
    global hold_positions
    logger.info("===== 多周期插针策略机器人启动 =====")
    while True:
        hold_positions = get_all_positions()
        hold_count = len(hold_positions.keys())
        monitor_tp_sl(hold_positions)
        # 持仓已满则不开新仓
        if hold_count >= MAX_HOLD_COINS:
            logger.info(f"当前持仓{hold_count}个，达上限，跳过开仓")
            time.sleep(3)
            continue

        for sym in SYMBOL_LIST:
            if sym in hold_positions:
                continue
            # 1. 获取15min趋势
            trend = get_trend_direction(sym)
            if not trend:
                continue
            # 2. 获取1min K线信号
            k1 = get_kline(sym, SIGNAL_INTERVAL, limit=50)
            if len(k1) < 30:
                continue
            atr_curr = calc_atr(k1, ATR_PERIOD)
            atr_prev = calc_atr(k1[:-1], ATR_PERIOD)
            # 震荡过滤
            if atr_curr < atr_prev * 0.7:
                continue
            # 检测插针
            pin_exist, pin_type = check_long_shadow(k1[-1])
            if not pin_exist:
                continue
            # 趋势与插针方向匹配校验
            if trend == "ALLOW_LONG" and pin_type != "LONG_PIN":
                continue
            if trend == "ALLOW_SHORT" and pin_type != "SHORT_PIN":
                continue
            # 指标共振校验
            if not check_resonance(k1, pin_type, atr_curr, atr_prev):
                continue
            # 计算开仓数量、止损价格
            entry_price = k1[-1]["close"]
            stop_range = MAX_SINGLE_LOSS / ((INIT_BALANCE * POSITION_RATIO) * LEVERAGE / entry_price)
            if pin_type == "LONG_PIN":
                stop_price = entry_price - stop_range
                side = "LONG"
            else:
                stop_price = entry_price + stop_range
                side = "SHORT"
            qty = calc_position_size(sym, entry_price, stop_price)
            if float(qty) <= 0:
                logger.warning(f"{sym} 计算仓位为0，跳过开仓")
                continue
            # 执行开仓
            logger.info(f"触发入场 {sym} {side} 插针共振信号")
            open_order(sym, side, qty)
            time.sleep(1)
        time.sleep(4)

if __name__ == "__main__":
    try:
        main_trade_loop()
    except KeyboardInterrupt:
        logger.info("手动停止机器人")
    except Exception as e:
        logger.error(f"主程序崩溃: {traceback.format_exc()}")
