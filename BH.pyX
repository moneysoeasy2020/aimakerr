# Termux专用 币安合约全自动交易机器人
# 策略：4H+1H EMA20/50共振定趋势｜3M+1M小周期精准开平仓
# 资金：50USDT本金、5倍杠杆、BTCUSDT+HYPEUSDT双标的
# 风控1：开仓1.5%固定止损｜浮盈2%启动移动止盈（回落0.8%止盈）
# 风控2：防重复下单+信号冷却锁仓
# 风控3：单日最大亏损熔断，亏损超标当日禁止新开仓
# 已修复latin1编码报错，API密钥预填完成
from binance.client import Client
import time
import datetime

# ===================== 预填你的币安API密钥 无需修改 =====================
API_KEY = "tj49IrEjBmedtwK3BEJNCqP1lk7fXgNkpmi5Yr2eNUhEXHG35XZzpow5TYu3S69y"
API_SECRET = "90ibaaGFYMzE6tE9IDqqnGEQZh0qZ76xReZQFt1AHteCRNEs1T76AGh6MDM3Lkno"

# 修复Termux requests UnicodeEncodeError 编码崩溃
import requests
from requests.packages.urllib3.util.connection import create_connection as urllib3_create_conn

def fix_charset(host, port, timeout=None, source_address=None, socket_options=None):
    conn = urllib3_create_conn(host, port, timeout, source_address, socket_options)
    return conn
urllib3_create_conn = fix_charset

# 初始化币安客户端
client = Client(API_KEY, API_SECRET)

# ===================== 核心参数固化，禁止手动修改 =====================
# 交易基础
LEVERAGE = 5
BASE_CAPITAL = 50
FAST_EMA = 20
SLOW_EMA = 50

# 止盈止损风控
FIX_STOP_LOSS_RATE = 0.015    # 固定止损1.5%
TRAIL_START_PROFIT = 0.02     # 盈利2%开启移动止盈
TRAIL_STEP = 0.008            # 回撤0.8%离场

# 防重复下单配置
SIGNAL_COOLDOWN_SEC = 60     # 同方向信号冷却60秒，防止重复开仓

# 单日亏损熔断配置
DAILY_MAX_LOSS_RATE = 0.08   # 单日亏损达到本金8%停止当日所有开仓

# 周期定义
TREND_4H = Client.KLINE_INTERVAL_4HOUR
TREND_1H = Client.KLINE_INTERVAL_1HOUR
ENTRY_3M = Client.KLINE_INTERVAL_3MINUTE
ENTRY_1M = Client.KLINE_INTERVAL_1MINUTE

SYMBOLS = ["BTCUSDT", "HYPEUSDT"]

# 持仓状态结构：开仓价、移动止盈最高价、上一次信号时间（防重下单）
position_info = {
    sym: {
        "long": False,
        "short": False,
        "entry_price": 0.0,
        "trail_high": 0.0,
        "last_long_signal_time": 0,
        "last_short_signal_time": 0
    } for sym in SYMBOLS
}

# 全局单日风控状态
daily_risk_state = {
    "trade_date": datetime.date.today(),
    "daily_realized_pnl": 0.0,
    "trading_halted": False
}

# ===================== 工具函数 =====================
# 批量设置杠杆
def set_leverage(symbol, lev):
    try:
        client.futures_change_leverage(symbol=symbol, leverage=lev)
        print(f"[{symbol}] 杠杆设置 {lev}倍 成功")
    except Exception as e:
        print(f"杠杆设置提示: {str(e)}")

# 标准EMA计算
def calc_ema(close_list, period):
    multi = 2 / (period + 1)
    ema_val = sum(close_list[:period]) / period
    for price in close_list[period:]:
        ema_val = price * multi + ema_val * (1 - multi)
    return ema_val

# 获取当日合约已实现盈亏（判断单日亏损熔断）
def get_today_realized_pnl():
    now = int(time.time() * 1000)
    today_start = int(datetime.datetime.combine(datetime.date.today(), datetime.time.min).timestamp() * 1000)
    # 获取当日所有平仓盈亏记录
    pnl_records = client.futures_income_history(incomeType="REALIZED_PNL", startTime=today_start, endTime=now)
    total_pnl = sum(float(item["income"]) for item in pnl_records)
    return round(total_pnl, 2)

# 每日重置熔断状态（跨天自动刷新）
def refresh_daily_risk_state():
    today = datetime.date.today()
    if daily_risk_state["trade_date"] != today:
        daily_risk_state["trade_date"] = today
        daily_risk_state["daily_realized_pnl"] = 0.0
        daily_risk_state["trading_halted"] = False
        print("【日期更新】新交易日，单日亏损风控已重置")

# 检查单日亏损是否触发熔断
def check_daily_loss_circuit_breaker():
    refresh_daily_risk_state()
    if daily_risk_state["trading_halted"]:
        return True
    # 实时获取当日盈亏
    daily_pnl = get_today_realized_pnl()
    daily_risk_state["daily_realized_pnl"] = daily_pnl
    # 亏损超过阈值，触发熔断
    if daily_pnl < -(BASE_CAPITAL * DAILY_MAX_LOSS_RATE):
        daily_risk_state["trading_halted"] = True
        print(f"⚠️ 单日已实现亏损 {daily_pnl:.2f} USDT，超过{DAILY_MAX_LOSS_RATE*100}%本金阈值，当日停止所有开仓！")
        return True
    return False

# 4H+1H双周期共振判断大趋势
def get_trend_direction(symbol):
    kl_4h = client.futures_klines(symbol=symbol, interval=TREND_4H, limit=100)
    kl_1h = client.futures_klines(symbol=symbol, interval=TREND_1H, limit=100)
    close4h = [float(i[4]) for i in kl_4h]
    close1h = [float(i[4]) for i in kl_1h]

    e20_4h = calc_ema(close4h, FAST_EMA)
    e50_4h = calc_ema(close4h, SLOW_EMA)
    e20_1h = calc_ema(close1h, FAST_EMA)
    e50_1h = calc_ema(close1h, SLOW_EMA)

    if e20_4h > e50_4h and e20_1h > e50_1h:
        return "LONG"
    elif e20_4h < e50_4h and e20_1h < e50_1h:
        return "SHORT"
    else:
        return "SIDEWAY"

# 3M+1M小周期进出场信号
def get_entry_signal(symbol, trend):
    if trend == "SIDEWAY":
        return None
    kl3m = client.futures_klines(symbol=symbol, interval=ENTRY_3M, limit=80)
    kl1m = client.futures_klines(symbol=symbol, interval=ENTRY_1M, limit=80)
    close3m = [float(i[4]) for i in kl3m]
    close1m = [float(i[4]) for i in kl1m]

    e20_3m = calc_ema(close3m, FAST_EMA)
    e50_3m = calc_ema(close3m, SLOW_EMA)
    e20_1m = calc_ema(close1m, FAST_EMA)
    e50_1m = calc_ema(close1m, SLOW_EMA)

    if trend == "LONG":
        if e20_3m > e50_3m and e20_1m > e50_1m:
            return "OPEN_LONG"
        elif e20_3m < e50_3m and e20_1m < e50_1m:
            return "CLOSE_LONG"
    if trend == "SHORT":
        if e20_3m < e50_3m and e20_1m < e50_1m:
            return "OPEN_SHORT"
        elif e20_3m > e50_3m and e20_1m > e50_1m:
            return "CLOSE_SHORT"
    return None

# 50U本金 5倍杠杆计算开仓数量
def calculate_position_size(symbol):
    ticker = client.futures_symbol_ticker(symbol=symbol)
    price = float(ticker["price"])
    usable_margin = BASE_CAPITAL * LEVERAGE
    if "BTC" in symbol:
        qty = round(usable_margin / price, 4)
    else:
        qty = round(usable_margin / price, 2)
    return qty

# 市价下单
def market_order(symbol, side, qty):
    order = client.futures_create_order(
        symbol=symbol,
        side=side,
        type=Client.ORDER_TYPE_MARKET,
        quantity=qty
    )
    return order

# 风控检测：固定止损 + 移动止盈
def risk_control_check(symbol, current_price):
    pos = position_info[symbol]
    entry = pos["entry_price"]

    # 多头风控
    if pos["long"]:
        loss_rate = (current_price - entry) / entry
        # 触发固定止损
        if loss_rate <= -FIX_STOP_LOSS_RATE:
            print(f"【{symbol}】多头触发1.5%固定止损，市价平仓")
            qty = calculate_position_size(symbol)
            market_order(symbol, Client.SIDE_SELL, qty)
            pos["long"] = False
            pos["entry_price"] = 0
            pos["trail_high"] = 0
            return True
        # 浮盈达标开启移动止盈
        if loss_rate >= TRAIL_START_PROFIT:
            if current_price > pos["trail_high"]:
                pos["trail_high"] = current_price
            # 价格回落触发移动止盈
            if current_price <= pos["trail_high"] * (1 - TRAIL_STEP):
                print(f"【{symbol}】多头移动止盈触发，最高价{pos['trail_high']}，现价回落平仓")
                qty = calculate_position_size(symbol)
                market_order(symbol, Client.SIDE_SELL, qty)
                pos["long"] = False
                pos["entry_price"] = 0
                pos["trail_high"] = 0
                return True

    # 空头风控
    if pos["short"]:
        loss_rate = (entry - current_price) / entry
        # 固定止损
        if loss_rate <= -FIX_STOP_LOSS_RATE:
            print(f"【{symbol}】空头触发1.5%固定止损，市价平仓")
            qty = calculate_position_size(symbol)
            market_order(symbol, Client.SIDE_BUY, qty)
            pos["short"] = False
            pos["entry_price"] = 0
            pos["trail_high"] = 0
            return True
        # 浮盈达标开启移动止盈（空头追踪最低点）
        if loss_rate >= TRAIL_START_PROFIT:
            if current_price < pos["trail_high"] or pos["trail_high"] == 0:
                pos["trail_high"] = current_price
            if current_price >= pos["trail_high"] * (1 + TRAIL_STEP):
                print(f"【{symbol}】空头移动止盈触发，最低价{pos['trail_high']}，现价反弹平仓")
                qty = calculate_position_size(symbol)
                market_order(symbol, Client.SIDE_BUY, qty)
                pos["short"] = False
                pos["entry_price"] = 0
                pos["trail_high"] = 0
                return True
    return False

# ===================== 主交易循环 =====================
if __name__ == "__main__":
    print("==================== 币安合约多周期自动交易【满配风控版】启动 ====================")
    print(f"本金：{BASE_CAPITAL}USDT | 杠杆：{LEVERAGE}倍 | 单日最大亏损熔断：{DAILY_MAX_LOSS_RATE*100}%")
    print(f"防重复下单冷却时间：{SIGNAL_COOLDOWN_SEC}秒")
    # 初始化杠杆
    for sym in SYMBOLS:
        set_leverage(sym, LEVERAGE)

    while True:
        try:
            # 第一步：检测单日亏损熔断，熔断状态直接跳过所有开仓逻辑
            if check_daily_loss_circuit_breaker():
                time.sleep(30)
                continue

            now_ts = time.time()
            for sym in SYMBOLS:
                ticker_price = float(client.futures_symbol_ticker(symbol=sym)["price"])
                pos = position_info[sym]

                # 每轮优先执行止损、移动止盈风控平仓
                risk_control_check(sym, ticker_price)

                trend = get_trend_direction(sym)
                signal = get_entry_signal(sym, trend)
                qty = calculate_position_size(sym)
                print(f"\n{sym} | 现价：{ticker_price:.4f} | 大趋势：{trend} | 交易信号：{signal}")

                # ========== 开多头逻辑（带防重复下单冷却） ==========
                if signal == "OPEN_LONG" and not pos["long"] and not pos["short"]:
                    # 冷却时间判断，60秒内同一信号禁止重复开仓
                    if now_ts - pos["last_long_signal_time"] < SIGNAL_COOLDOWN_SEC:
                        print(f"【{sym}】多头信号冷却中，剩余{int(SIGNAL_COOLDOWN_SEC - (now_ts - pos['last_long_signal_time']))}秒，跳过开仓")
                        continue
                    # 执行开多
                    market_order(sym, Client.SIDE_BUY, qty)
                    pos["long"] = True
                    pos["entry_price"] = ticker_price
                    pos["trail_high"] = ticker_price
                    pos["last_long_signal_time"] = now_ts
                    print(f"【开多仓】入场价：{ticker_price:.4f} 数量：{qty}")

                # 信号反转平多头
                if signal == "CLOSE_LONG" and pos["long"]:
                    market_order(sym, Client.SIDE_SELL, qty)
                    pos["long"] = False
                    pos["entry_price"] = 0
                    pos["trail_high"] = 0
                    print(f"【信号反转平多仓】")

                # ========== 开空头逻辑（带防重复下单冷却） ==========
                if signal == "OPEN_SHORT" and not pos["short"] and not pos["long"]:
                    if now_ts - pos["last_short_signal_time"] < SIGNAL_COOLDOWN_SEC:
                        print(f"【{sym}】空头信号冷却中，剩余{int(SIGNAL_COOLDOWN_SEC - (now_ts - pos['last_short_signal_time']))}秒，跳过开仓")
                        continue
                    # 执行开空
                    market_order(sym, Client.SIDE_SELL, qty)
                    pos["short"] = True
                    pos["entry_price"] = ticker_price
                    pos["trail_high"] = ticker_price
                    pos["last_short_signal_time"] = now_ts
                    print(f"【开空仓】入场价：{ticker_price:.4f} 数量：{qty}")

                # 信号反转平空头
                if signal == "CLOSE_SHORT" and pos["short"]:
                    market_order(sym, Client.SIDE_BUY, qty)
                    pos["short"] = False
                    pos["entry_price"] = 0
                    pos["trail_high"] = 0
                    print(f"【信号反转平空仓】")

        except Exception as e:
            print("主循环异常捕获：", str(e))
        time.sleep(60)
