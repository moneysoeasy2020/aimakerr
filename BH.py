# -*- coding: utf-8 -*-
from binance.client import Client
from binance.exceptions import BinanceAPIException, BinanceOrderException
import time
import urllib3

# 关闭SSL警告
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

# ========== 你的API 直接写在这里 ==========
API_KEY = "tj49IrEjBmedtwK3BEJNCqP1lk7fXgNkpmi5Yr2eNUhEXHG35XZzpow5TYu3S69y"
API_SECRET = "90ibaaGFYMzE6tE9IDqqnGEQZh0qZ76xReZQFt1AHteCRNEs1T76AGh6MDM3Lkno"

client = Client(API_KEY, API_SECRET)

# ===================== 全局配置 =====================
LEVERAGE = 5          # 固定5倍杠杆
BASE_BALANCE = 50.0   # 初始本金50USDT
MAX_DAILY_LOSS = 8.0  # 单日最大亏损8U（可自行改）
MAX_HYPE_POS = 30.0   # HYPE单品种占用不超30U
SYMBOLS = ["BTCUSDT", "HYPEUSDT"]
EMA_FAST = 20
EMA_SLOW = 50

# 风控全局变量
daily_loss = 0.0
last_order_time = 0
ORDER_INTERVAL = 60   # 防重复下单：60秒内不能重复开同方向单
in_position = {s: False for s in SYMBOLS}


def get_ema(symbol, interval, length):
    """获取EMA"""
    klines = client.futures_klines(symbol=symbol, interval=interval, limit=length+5)
    closes = [float(k[4]) for k in klines]
    return sum(closes[-length:]) / length


def get_atr(symbol, interval, period=14):
    """ATR波动率过滤，震荡不开仓"""
    klines = client.futures_klines(symbol=symbol, interval=interval, limit=period+1)
    tr_list = []
    for i in range(1, len(klines)):
        high = float(klines[i][2])
        low = float(klines[i][3])
        prev_close = float(klines[i-1][4])
        tr = max(high - low, abs(high - prev_close), abs(low - prev_close))
        tr_list.append(tr)
    return sum(tr_list) / period


def get_direction(symbol):
    """4h+1h共振定多空"""
    ema4h_f = get_ema(symbol, Client.KLINE_INTERVAL_4HOUR, EMA_FAST)
    ema4h_s = get_ema(symbol, Client.KLINE_INTERVAL_4HOUR, EMA_SLOW)
    ema1h_f = get_ema(symbol, Client.KLINE_INTERVAL_1HOUR, EMA_FAST)
    ema1h_s = get_ema(symbol, Client.KLINE_INTERVAL_1HOUR, EMA_SLOW)

    long_signal = (ema4h_f > ema4h_s) and (ema1h_f > ema1h_s)
    short_signal = (ema4h_f < ema4h_s) and (ema1h_f < ema1h_s)

    if long_signal:
        return "LONG"
    elif short_signal:
        return "SHORT"
    else:
        return "NONE"


def entry_signal(symbol, direction):
    """3m/1m小周期进出场确认"""
    ema3m_f = get_ema(symbol, Client.KLINE_INTERVAL_3MINUTE, EMA_FAST)
    ema3m_s = get_ema(symbol, Client.KLINE_INTERVAL_3MINUTE, EMA_SLOW)
    if direction == "LONG" and ema3m_f > ema3m_s:
        return True
    if direction == "SHORT" and ema3m_f < ema3m_s:
        return True
    return False


def get_usdt_balance():
    bal = client.futures_account_balance()
    for b in bal:
        if b["asset"] == "USDT":
            return float(b["balance"])
    return 0.0


def open_order(symbol, side, usdt_amount):
    global last_order_time, daily_loss
    now = time.time()
    # 防重复下单
    if now - last_order_time < ORDER_INTERVAL:
        print(f"[{symbol}] 禁止重复下单，间隔不足")
        return
    # 单日亏损风控
    if daily_loss >= MAX_DAILY_LOSS:
        print(f"单日亏损已达上限 {daily_loss}U，停止交易")
        return
    # HYPE仓位限制
    if symbol == "HYPEUSDT" and usdt_amount > MAX_HYPE_POS:
        usdt_amount = MAX_HYPE_POS

    price = float(client.futures_mark_price(symbol=symbol)["markPrice"])
    qty = (usdt_amount * LEVERAGE) / price

    try:
        client.futures_change_leverage(symbol=symbol, leverage=LEVERAGE)
        order = client.futures_create_order(
            symbol=symbol,
            side=side,
            type=Client.ORDER_TYPE_MARKET,
            quantity=round(qty, 4)
        )
        last_order_time = now
        in_position[symbol] = True
        print(f"开仓成功 {symbol} {side} 数量:{qty:.4f}")
        return order
    except BinanceAPIException as e:
        print(f"开仓失败: {e}")
        return None


def close_all_pos(symbol):
    pos = client.futures_position_information(symbol=symbol)[0]
    amt = float(pos["positionAmt"])
    if abs(amt) < 0.0001:
        return
    side = Client.SIDE_SELL if amt > 0 else Client.SIDE_BUY
    client.futures_create_order(
        symbol=symbol,
        side=side,
        type=Client.ORDER_TYPE_MARKET,
        quantity=abs(amt)
    )
    in_position[symbol] = False
    print(f"{symbol} 全部平仓")


def main_loop():
    global daily_loss
    print("策略启动：4h/1h定多空，3m/1m进场 | 5倍杠杆 | 单日最大亏损", MAX_DAILY_LOSS)
    while True:
        for sym in SYMBOLS:
            # 波动率过滤：震荡跳过
            atr = get_atr(sym, Client.KLINE_INTERVAL_1HOUR)
            price = float(client.futures_mark_price(sym)["markPrice"])
            atr_ratio = atr / price
            if atr_ratio < 0.0012:
                print(f"{sym} 震荡过窄，禁止开仓")
                continue

            direction = get_direction(sym)
            if direction == "NONE":
                if in_position[sym]:
                    close_all_pos(sym)
                continue

            if entry_signal(sym, direction) and not in_position[sym]:
                open_side = Client.SIDE_BUY if direction == "LONG" else Client.SIDE_SELL
                open_order(sym, open_side, BASE_BALANCE)

        time.sleep(20)


if __name__ == "__main__":
    try:
        main_loop()
    except KeyboardInterrupt:
        print("程序退出，平仓所有持仓")
        for s in SYMBOLS:
            close_all_pos(s)
