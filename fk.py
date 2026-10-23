import asyncio
import datetime
import ccxt.pro as ccxtpro
import pandas as pd
import numpy as np
import pandas_ta as ta

# ================= 基础与风控配置 =================
API_KEY = "tj49IrEjBmedtwK3BEJNCqP1lk7fXgNkpmi5Yr2eNUhEXHG35XZzpow5TYu3S69y"
SECRET_KEY = "90ibaaGFYMzE6tE9IDqqnGEQZh0qZ76xReZQFt1AHteCRNEs1T76AGh6MDM3Lkno"

MARGIN_USDT = 2.8         # 单次保证金
LEVERAGE = 10             # 杠杆倍数
NOMINAL_VALUE = MARGIN_USDT * LEVERAGE  # 28 USDT 名义持仓
MAX_POSITIONS = 2         # 最多持仓币种数

MAX_DAILY_DRAWDOWN_PCT = 0.10  # 单日最大回撤熔断比例 (10%)

exchange = ccxtpro.binance({
    'apiKey': API_KEY,
    'secret': SECRET_KEY,
    'options': {'defaultType': 'future'},
    'enableRateLimit': True,
})

# 风控状态全局变量
class RiskManager:
    def __init__(self):
        self.today = datetime.datetime.utcnow().date()
        self.day_start_balance = 0.0
        self.is_circuit_broken = False

risk_mgr = RiskManager()

# ================= 熔断风控检查 =================
async def check_circuit_breaker():
    """实时监控单日回撤，触发 10% 亏损即清仓熔断"""
    current_date = datetime.datetime.utcnow().date()
    balance_data = await exchange.fetch_balance()
    usdt_info = balance_data.get('USDT', {})
    
    total_equity = float(usdt_info.get('total', 0))
    
    # 跨日重置起始资金与熔断状态 (UTC 00:00)
    if current_date != risk_mgr.today or risk_mgr.day_start_balance == 0:
        risk_mgr.today = current_date
        risk_mgr.day_start_balance = total_equity
        risk_mgr.is_circuit_broken = False
        print(f"🌅 [新一日风控基准更新] UTC Date: {current_date} | 日初总权益: ${total_equity:.2f} USDT")

    # 计算今日动态盈亏
    daily_pnl = total_equity - risk_mgr.day_start_balance
    daily_drawdown_pct = -daily_pnl / risk_mgr.day_start_balance if risk_mgr.day_start_balance > 0 else 0

    if daily_drawdown_pct >= MAX_DAILY_DRAWDOWN_PCT and not risk_mgr.is_circuit_broken:
        risk_mgr.is_circuit_broken = True
        print(f"🚨🚨🚨 [风控触发] 今日回撤达到 {daily_drawdown_pct*100:.2f}% (阈值 {MAX_DAILY_DRAWDOWN_PCT*100}%)，执行应急清仓熔断！")
        await emergency_close_all()

    return risk_mgr.is_circuit_broken

async def emergency_close_all():
    """紧急清仓并撤销所有挂单"""
    try:
        # 1. 获取所有持仓并市价平仓
        balance = await exchange.fetch_balance()
        positions = balance.get('info', {}).get('positions', [])
        for pos in positions:
            amt = float(pos.get('positionAmt', 0))
            symbol = pos['symbol']
            if not symbol.endswith(':USDT'):
                symbol = f"{symbol[:-4]}/USDT:USDT" if symbol.endswith('USDT') else symbol
            
            if amt != 0:
                side = 'sell' if amt > 0 else 'buy'
                await exchange.create_market_order(symbol, side, abs(amt), params={'reduceOnly': True})
                print(f"🛡️ [紧急平仓] {symbol} 数量: {abs(amt)}")

        # 2. 撤销所有委托单
        for pos in positions:
            symbol = pos['symbol']
            if not symbol.endswith(':USDT'):
                symbol = f"{symbol[:-4]}/USDT:USDT" if symbol.endswith('USDT') else symbol
            await exchange.cancel_all_orders(symbol)
        print("🛡️ [紧急风控完成] 所有仓位已平，挂单已撤销。今日停止开仓。")
    except Exception as e:
        print(f"❌ 紧急清仓异常: {e}")

# ================= 动态标的与信号逻辑 =================
async def fetch_high_alpha_universe():
    tickers = await exchange.fetch_tickers()
    candidates = []
    for symbol, t in tickers.items():
        if not symbol.endswith('/USDT:USDT') or any(base in symbol for base in ['BTC', 'ETH', 'USDC', 'FDUSD']):
            continue
        quote_vol = t.get('quoteVolume', 0)
        percentage = abs(t.get('percentage', 0))
        if 10_000_000 <= quote_vol <= 100_000_000 and percentage >= 6.0:
            candidates.append((symbol, quote_vol))
    candidates.sort(key=lambda x: x[1], reverse=True)
    return [c[0] for c in candidates[:20]]

def analyze_market_structure(df):
    df['rsi'] = ta.rsi(df['close'], length=14)
    df['ema7'] = ta.ema(df['close'], length=7)
    df['ema21'] = ta.ema(df['close'], length=21)
    df['atr'] = ta.atr(df['high'], df['low'], df['close'], length=14)
    df['vol_ma20'] = ta.sma(df['vol'], length=20)

    curr, prev = df.iloc[-1], df.iloc[-2]
    recent_high = df['high'].iloc[-6:-1].max()
    recent_low = df['low'].iloc[-6:-1].min()

    signal = None
    volume_spike = curr['vol'] > (curr['vol_ma20'] * 2.5) if curr['vol_ma20'] > 0 else False
    atr_ratio = curr['atr'] / curr['close']
    valid_volatility = 0.004 <= atr_ratio <= 0.025

    if volume_spike and valid_volatility:
        if prev['rsi'] > 75 and prev['high'] >= recent_high:
            if curr['close'] < prev['low'] or (prev['ema7'] >= prev['ema21'] and curr['ema7'] < curr['ema21']):
                signal = 'SHORT'
        elif prev['rsi'] < 25 and prev['low'] <= recent_low:
            if curr['close'] > prev['high'] or (prev['ema7'] <= prev['ema21'] and curr['ema7'] > curr['ema21']):
                signal = 'LONG'

    return signal, curr['atr']

async def get_active_positions():
    balance = await exchange.fetch_balance()
    positions = balance.get('info', {}).get('positions', [])
    active = []
    for pos in positions:
        if float(pos.get('positionAmt', 0)) != 0:
            symbol = pos['symbol']
            if not symbol.endswith(':USDT'):
                symbol = f"{symbol[:-4]}/USDT:USDT" if symbol.endswith('USDT') else symbol
            active.append(symbol)
    return active

async def execute_order(symbol, side, atr):
    try:
        await exchange.set_leverage(LEVERAGE, symbol)
        ticker = await exchange.fetch_ticker(symbol)
        price = ticker['last']
        amount = NOMINAL_VALUE / price

        order_side = 'buy' if side == 'LONG' else 'sell'
        await exchange.create_market_order(symbol, order_side, amount)
        print(f"🚀 [开仓] {symbol} | 方向: {side} | 数量: {amount:.4f} | 价格: {price}")

        sl_distance = max(atr * 1.5, price * 0.008)
        tp_distance = sl_distance * 1.8

        if side == 'LONG':
            tp_price, sl_price, close_side = price + tp_distance, price - sl_distance, 'sell'
        else:
            tp_price, sl_price, close_side = price - tp_distance, price + sl_distance, 'buy'

        await exchange.create_order(symbol, 'TAKE_PROFIT_MARKET', close_side, amount, params={'stopPrice': tp_price, 'reduceOnly': True})
        await exchange.create_order(symbol, 'STOP_MARKET', close_side, amount, params={'stopPrice': sl_price, 'reduceOnly': True})
        print(f"🎯 [挂单] {symbol} | 止盈: {tp_price:.5f} | 止损: {sl_price:.5f}")
    except Exception as e:
        print(f"❌ [下单失败] {symbol}: {e}")

# ================= 主循环 =================
async def main():
    print("⚡ 带有单日熔断保护的高频量能回归脚本启动...")
    while True:
        try:
            # 1. 检查单日熔断状态
            is_broken = await check_circuit_breaker()
            if is_broken:
                print("🛑 当前处于熔断保护状态，暂停开仓检索，等待新一天 UTC00:00 重置...")
                await asyncio.sleep(60)
                continue

            # 2. 正常仓位检索与开仓
            active_positions = await get_active_positions()
            current_count = len(active_positions)

            if current_count < MAX_POSITIONS:
                universe = await fetch_high_alpha_universe()
                for symbol in universe:
                    if symbol in active_positions:
                        continue
                    
                    ohlcv = await exchange.fetch_ohlcv(symbol, timeframe='1m', limit=50)
                    if len(ohlcv) < 50:
                        continue
                        
                    df = pd.DataFrame(ohlcv, columns=['time', 'open', 'high', 'low', 'close', 'vol'])
                    signal, atr = analyze_market_structure(df)

                    if signal:
                        print(f"🔥 捕获量能结构信号: {symbol} -> {signal}")
                        await execute_order(symbol, signal, atr)
                        active_positions = await get_active_positions()
                        if len(active_positions) >= MAX_POSITIONS:
                            break
                            
        except Exception as e:
            print(f"⚠️ 运行异常: {e}")
            
        await asyncio.sleep(8)

if __name__ == '__main__':
    asyncio.run(main())
