import time
import hashlib
import hmac
import requests
import logging
import re
import traceback
import urllib3

# ===================== 【已填入API密钥 固化参数禁止修改】 =====================
API_KEY = "tj49IrEjBmedtwK3BEJNCqP1lk7fXgNkpmi5Yr2eNUhEXHG35XZzpow5TYu3S69y"
API_SECRET = "90ibaaGFYMzE6tE9IDqqnGEQZh0qZ76xReZQFt1AHteCRNEs1T76AGh6MDM3Lkno"

# 多标的配置：最多3个永续合约标的，可自行替换
SYMBOL_LIST = ["HYPEUSDT", "DOGEUSDT", "SOLUSDT"]
BASE_URL = "https://fapi.binance.com"

# 固定均线参数（原版EMA20/EMA50）全局统一
EMA_FAST = 20
EMA_SLOW = 50

# 止盈止损比例固化（全局统一）
TP_RATE = 0.035
SL_RATE = 0.025

# 单个标的单次开仓占用USDT本金
USDT_CAP = 20

# 均线交叉确认K线根数（连续3根确认，防震荡假信号）
CONFIRM_KLINE = 3

# 主循环轮询间隔 10秒
LOOP_INTERVAL = 10
# =================================================================

# 日志配置
logging.basicConfig(
    format="%(asctime)s | %(levelname)s | %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S",
    level=logging.INFO
)
logger = logging.getLogger(__name__)
urllib3.disable_warnings()

# 多标的精度缓存字典：key=交易对，value=精度信息
symbol_precision = {}


def get_timestamp() -> int:
    return int(time.time() * 1000)


def sign_payload(payload: dict, secret: str) -> str:
    query_str = "&".join([f"{k}={v}" for k, v in payload.items()])
    return hmac.new(secret.encode(), query_str.encode(), hashlib.sha256).hexdigest()


def api_request(method: str, path: str, params: dict = None):
    headers = {"X-MBX-APIKEY": API_KEY}
    url = BASE_URL + path
    params = params or {}
    params["timestamp"] = get_timestamp()
    params["signature"] = sign_payload(params, API_SECRET)
    try:
        resp = requests.request(method, url, headers=headers, params=params, timeout=12, verify=False)
        resp.raise_for_status()
        return resp.json()
    except Exception as e:
        logger.error(f"API请求异常: {str(e)}")
        return None


# 批量加载所有标的交易精度规则
def load_all_symbol_rule():
    global symbol_precision
    info = api_request("GET", "/fapi/v1/exchangeInfo")
    if not info:
        logger.warning("加载交易规则失败，3秒重试")
        time.sleep(3)
        return load_all_symbol_rule()

    for target_sym in SYMBOL_LIST:
        for sym in info["symbols"]:
            if sym["symbol"] == target_sym:
                tick_size = None
                step_size = None
                min_qty = None
                price_decimals = 0
                qty_decimals = 0
                for f in sym["filters"]:
                    if f["filterType"] == "PRICE_FILTER":
                        tick_size = float(f["tickSize"])
                        price_decimals = len(re.sub(r"0+$", "", f["tickSize"]).split(".")[1]) if "." in f["tickSize"] else 0
                    if f["filterType"] == "LOT_SIZE":
                        step_size = float(f["stepSize"])
                        min_qty = float(f["minQty"])
                        qty_decimals = len(re.sub(r"0+$", "", f["stepSize"]).split(".")[1]) if "." in f["stepSize"] else 0
                symbol_precision[target_sym] = {
                    "tick_size": tick_size,
                    "step_size": step_size,
                    "min_qty": min_qty,
                    "price_decimals": price_decimals,
                    "qty_decimals": qty_decimals
                }
                logger.info(f"✅ [{target_sym}] 精度加载完成")
                break


# 价格格式化（适配当前标的精度）
def format_price(symbol: str, price: float) -> float:
    prec = symbol_precision[symbol]
    val = round(price / prec["tick_size"]) * prec["tick_size"]
    return float(f"{val:.{prec['price_decimals']}f}")


# 开仓数量格式化（适配当前标的精度）
def format_qty(symbol: str, qty: float) -> float:
    prec = symbol_precision[symbol]
    val = int(qty / prec["step_size"]) * prec["step_size"]
    safe_qty = max(val, prec["min_qty"])
    return float(f"{safe_qty:.{prec['qty_decimals']}f}")


# 获取指定标的K线收盘价列表
def get_kline_close(symbol: str, interval="1m", limit=200) -> list:
    klines = api_request("GET", "/fapi/v1/klines", {"symbol": symbol, "interval": interval, "limit": limit})
    if not klines:
        return []
    return [float(item[4]) for item in klines]


# 计算EMA
def calc_ema(price_list: list, period: int) -> float:
    k = 2 / (period + 1)
    sma = sum(price_list[:period]) / period
    ema_val = sma
    for p in price_list[period:]:
        ema_val = p * k + ema_val * (1 - k)
    return ema_val


# 获取标记价格（用于止盈止损计算）
def get_mark_price(symbol: str) -> float:
    res = api_request("GET", "/fapi/v1/premiumIndex", {"symbol": symbol})
    return float(res["markPrice"]) if res else 0.0


# 查询单个标的持仓状态：返回(持仓方向LONG/SHORT/None, 开仓价, 持仓数量)
def get_position(symbol: str) -> tuple:
    pos_data = api_request("GET", "/fapi/v2/positionRisk")
    if not pos_data:
        return None, 0.0, 0.0

    for pos in pos_data:
        if pos["symbol"] == symbol:
            amt = float(pos["positionAmt"])
            entry = float(pos["entryPrice"])
            abs_amt = abs(amt)
            if amt > 1e-9:
                return "LONG", entry, abs_amt
            elif amt < -1e-9:
                return "SHORT", entry, abs_amt
    return None, 0.0, 0.0


# 市价平掉当前标的全部持仓
def market_close_pos(symbol: str, pos_side: str, qty: float):
    close_side = "SELL" if pos_side == "LONG" else "BUY"
    safe_qty = format_qty(symbol, qty)
    params = {
        "symbol": symbol,
        "side": close_side,
        "type": "MARKET",
        "quantity": safe_qty,
        "positionSide": pos_side
    }
    res = api_request("POST", "/fapi/v1/order", params)
    if res:
        logger.info(f"🏁 [{symbol}] 轮询平仓完成 | 订单ID:{res['orderId']}")
    else:
        logger.error(f"❌ [{symbol}] 市价平仓请求失败，请手动检查持仓")


# 市价开仓（单向持仓，不会双向开仓）
def market_open_pos(symbol: str, pos_side: str):
    tick = api_request("GET", "/fapi/v1/ticker/price", {"symbol": symbol})
    if not tick:
        logger.error(f"[{symbol}] 获取市价失败，放弃开仓")
        return None

    current_price = float(tick["price"])
    raw_qty = USDT_CAP / current_price
    safe_qty = format_qty(symbol, raw_qty)
    open_side = "BUY" if pos_side == "LONG" else "SELL"

    params = {
        "symbol": symbol,
        "side": open_side,
        "type": "MARKET",
        "quantity": safe_qty,
        "positionSide": pos_side
    }
    res = api_request("POST", "/fapi/v1/order", params)
    if res:
        logger.info(f"✅ [{symbol}] {pos_side}开仓成功 | 数量:{safe_qty} 订单ID:{res['orderId']}")
    else:
        logger.error(f"❌ [{symbol}] {pos_side}开仓下单失败")


# 轮询止盈止损判断逻辑
def check_tp_sl(symbol: str, pos_side: str, entry_price: float, pos_qty: float) -> bool:
    """返回True代表触发平仓，False继续持有"""
    if pos_side is None or entry_price <= 0 or pos_qty <= 0:
        return False

    mark_price = get_mark_price(symbol)
    if mark_price <= 0:
        logger.warning(f"[{symbol}] 获取标记价格失败，暂不判断止盈止损")
        return False

    if pos_side == "LONG":
        profit_rate = (mark_price - entry_price) / entry_price
        # 止损触发
        if profit_rate <= -SL_RATE:
            logger.info(f"🛑 [{symbol}] 多头硬性止损触发 | 入场价:{entry_price:.4f} 现价:{mark_price:.4f} 亏损{abs(profit_rate):.2%}")
            return True
        # 止盈触发
        if profit_rate >= TP_RATE:
            logger.info(f"💰 [{symbol}] 多头止盈触发 | 收益{profit_rate:.2%}")
            return True

    elif pos_side == "SHORT":
        profit_rate = (entry_price - mark_price) / entry_price
        # 止损触发
        if profit_rate <= -SL_RATE:
            logger.info(f"🛑 [{symbol}] 空头硬性止损触发 | 入场价:{entry_price:.4f} 现价:{mark_price:.4f} 亏损{abs(profit_rate):.2%}")
            return True
        # 止盈触发
        if profit_rate >= TP_RATE:
            logger.info(f"💰 [{symbol}] 空头止盈触发 | 收益{profit_rate:.2%}")
            return True

    return False


# 检测连续N根K线金叉/死叉确认信号
def get_trade_signal(symbol: str) -> str:
    """返回 LONG/SHORT/NONE"""
    close_list = get_kline_close(symbol, interval="1m", limit=300)
    if len(close_list) < EMA_SLOW + CONFIRM_KLINE:
        logger.warning(f"[{symbol}] K线数据不足，无法计算交叉信号")
        return "NONE"

    # 逐根计算EMA序列
    ema_fast_list = []
    ema_slow_list = []
    for i in range(EMA_SLOW, len(close_list)):
        slice_close = close_list[:i+1]
        ef = calc_ema(slice_close, EMA_FAST)
        es = calc_ema(slice_close, EMA_SLOW)
        ema_fast_list.append(ef)
        ema_slow_list.append(es)

    # 取最后CONFIRM_KLINE根K线做交叉确认
    confirm_range = CONFIRM_KLINE
    recent_ef = ema_fast_list[-confirm_range:]
    recent_es = ema_slow_list[-confirm_range:]

    # 金叉确认：连续N根 快线上穿慢线
    golden_all = all(recent_ef[i] > recent_es[i] for i in range(confirm_range))
    # 死叉确认：连续N根 快线下穿慢线
    death_all = all(recent_ef[i] < recent_es[i] for i in range(confirm_range))

    if golden_all:
        return "LONG"
    elif death_all:
        return "SHORT"
    else:
        return "NONE"


# 单标的完整一轮逻辑处理
def process_single_symbol(symbol: str):
    logger.info(f"\n========== 开始处理标的：{symbol} ==========")
    pos_side, entry_price, pos_qty = get_position(symbol)

    # 优先级1：持仓状态下只监控止盈止损，禁止新开反向仓
    if pos_side is not None:
        logger.info(f"[{symbol}] 当前持有{pos_side}仓位，仅执行止盈止损监控")
        if check_tp_sl(symbol, pos_side, entry_price, pos_qty):
            market_close_pos(symbol, pos_side, pos_qty)
        return

    # 优先级2：空仓状态，计算开仓信号
    signal = get_trade_signal(symbol)
    if signal == "NONE":
        logger.info(f"[{symbol}] 暂无有效金叉/死叉开仓信号，等待下一轮")
        return

    # 空仓顺势开仓，同一标的永远不会双向持仓
    market_open_pos(symbol, signal)


# 主循环入口
def main():
    logger.info("==================== 多标的合约自动交易机器人启动 ====================")
    logger.info(f"监控标的列表：{SYMBOL_LIST} | 单标的止盈{TP_RATE:.1%} 止损{SL_RATE:.1%}")
    logger.info(f"开仓确认K线数：{CONFIRM_KLINE}根 | 轮询间隔：{LOOP_INTERVAL}秒")
    load_all_symbol_rule()

    while True:
        # 遍历3个标的独立运行
        for sym in SYMBOL_LIST:
            try:
                process_single_symbol(sym)
            except Exception as err:
                logger.error(f"[{sym}] 单次循环异常：{traceback.format_exc()}")

        logger.info(f"\n本轮全部标的遍历完成，休眠 {LOOP_INTERVAL} 秒")
        time.sleep(LOOP_INTERVAL)


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        logger.info("\n程序手动终止，交易机器人退出")
