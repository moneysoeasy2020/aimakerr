
import time
import logging
from datetime import datetime, timezone
import ccxt
import pandas as pd

# ==================== 1. 配置参数 ====================
API_KEY = "tj49IrEjBmedtwK3BEJNCqP1lk7fXgNkpmi5Yr2eNUhEXHG35XZzpow5TYu3S69y"
SECRET_KEY = "90ibaaGFYMzE6tE9IDqqnGEQZh0qZ76xReZQFt1AHteCRNEs1T76AGh6MDM3Lkno"

# 针对截图的高杠杆 Scalping 参数调整
BASE_MARGIN = 10.0            # 单笔基础保证金 (USDT)
LEVERAGE = 10                 # 建议实盘自动化使用 10x (截图为 20x-50x，自动化建议留容错率)
CHANGE_THRESHOLD = 10.0       # 24h 波动幅度阀门 (%)
MIN_QUOTE_VOLUME = 15000000.0 # 成交额过滤 (1500万 USDT)
SL_PADDING_PCT = 0.004        # 极窄硬止损 (0.4%)，适应高杠杆

# 冷却记录 {symbol: "YYYY-MM-DD"}
traded_today = {}

logging.basicConfig(level=logging.INFO, format="%(asctime)s - [%(levelname)s] - %(message)s")

exchange = ccxt.binanceusdm({
    'apiKey': API_KEY,
    'secret': SECRET_KEY,
    'enableRateLimit': True,
    'options': {'defaultType': 'future'}
})

# ==================== 2. 单日交易限制 ====================
def can_trade_today(symbol):
    today_str = datetime.now(timezone.utc).strftime('%Y-%m-%d')
    return traded_today.get(symbol) != today_str

def mark_as_traded(symbol):
    today_str = datetime.now(timezone.utc).strftime('%Y-%m-%d')
    traded_today[symbol] = today_str

# ==================== 3. 图像实盘逻辑：1m/5m 极短线双向信号识别 ====================
def analyze_scalp_signal(symbol):
    """
    基于截图形态的微观结构信号检测:
    1. Short 信号: 1m 冲高扫荡前高流动性后阴线跌破前低 (Sweep & Reject)
    2. Long 信号: 1m 下探扫荡前低流动性后阳线收复支撑 (Sweep & Reclaim)
    """
    try:
        # 读取 1 分钟 K 线（截图均为 1m 级高精细度图表）
        ohlcv = exchange.fetch_ohlcv(symbol, timeframe='1m', limit=30)
        if len(ohlcv) < 20:
            return None, 0, 0
            
        df = pd.DataFrame(ohlcv, columns=['timestamp', 'open', 'high', 'low', 'close', 'volume'])
        
        recent_high = df['high'][:-3].max()
        recent_low = df['low'][:-3].min()
        
        prev_k = df.iloc[-2]  # 刚闭合的 1m K线
        curr_k = df.iloc[-1]  # 当前 1m K线
        
        # --- A. 空头信号（做空高点假突破）---
        short_sweep = prev_k['high'] >= recent_high
        short_break = curr_k['close'] < prev_k['low']
        
        if short_sweep and short_break:
            sl_price = max(prev_k['high'], recent_high) * (1 + SL_PADDING_PCT)
            tp_price = recent_low  # 止盈目标：箱体底部（对应截图 CLOSED 底部线）
            return 'SHORT', sl_price, tp_price

        # --- B. 多头信号（做多低点假跌破 - 对应截图 FHE/AIOT Long）---
        long_sweep = prev_k['low'] <= recent_low
        long_reclaim = curr_k['close'] > prev_k['high']
        
        if long_sweep and long_reclaim:
            sl_price = min(prev_k['low'], recent_low) * (1 - SL_PADDING_PCT)
            tp_price = recent_high # 止盈目标：箱体顶部（对应截图 CLOSED 顶部线）
            return 'LONG', sl_price, tp_price

        return None, 0, 0
    except Exception as e:
        logging.error(f"[{symbol}] 信号分析异常: {e}")
        return None, 0, 0

# ==================== 4. 自动化下单模块 ====================
def execute_scalp_trade(symbol, side, sl_price, tp_price):
    try:
        try:
            exchange.set_margin_mode('ISOLATED', symbol)
        except Exception:
            pass
        exchange.set_leverage(LEVERAGE, symbol)
        
        ticker = exchange.fetch_ticker(symbol)
        curr_price = ticker['last']
        
        position_value = BASE_MARGIN * LEVERAGE
        amount = position_value / curr_price
        amount = float(exchange.amount_to_precision(symbol, amount))
        
        if amount <= 0:
            return

        order_side = 'sell' if side == 'SHORT' else 'buy'
        close_side = 'buy' if side == 'SHORT' else 'sell'

        logging.info(f"==> 触发 {side} 交易 | 标的: {symbol} | 数量: {amount} | 价格: {curr_price}")
        
        # 1. 市价开仓
        exchange.create_order(symbol=symbol, type='market', side=order_side, amount=amount)
        mark_as_traded(symbol)

        # 2. 精确硬止损 (STOP_MARKET)
        sl_price = float(exchange.price_to_precision(symbol, sl_price))
        exchange.create_order(
            symbol=symbol,
            type='STOP_MARKET',
            side=close_side,
            amount=amount,
            params={'stopPrice': sl_price, 'reduceOnly': True}
        )
        
        # 3. 精确箱体止盈 (TAKE_PROFIT_MARKET / LIMIT)
        tp_price = float(exchange.price_to_precision(symbol, tp_price))
        exchange.create_order(
            symbol=symbol,
            type='TAKE_PROFIT_MARKET',
            side=close_side,
            amount=amount,
            params={'stopPrice': tp_price, 'reduceOnly': True}
        )
        
        logging.info(f"[{symbol}] 开仓完成！硬止损价: {sl_price} | 箱体目标止盈价: {tp_price}")

    except Exception as e:
        logging.error(f"执行 {symbol} 交易失败: {e}")

# ==================== 5. 主循环 ====================
def main_loop():
    logging.info("图片实盘复刻版：1m 微观双向 Scalping 机器人启动...")
    while True:
        try:
            tickers = exchange.fetch_tickers()
            for symbol, ticker in tickers.items():
                if not symbol.endswith('/USDT:USDT'):
                    continue
                if not can_trade_today(symbol):
                    continue

                # 筛选高波动标的
                if abs(ticker.get('percentage', 0)) >= CHANGE_THRESHOLD and ticker.get('quoteVolume', 0) >= MIN_QUOTE_VOLUME:
                    side, sl_price, tp_price = analyze_scalp_signal(symbol)
                    if side:
                        execute_scalp_trade(symbol, side, sl_price, tp_price)
                        time.sleep(10)
            time.sleep(15)
        except KeyboardInterrupt:
            break
        except Exception as e:
            logging.error(f"主循环异常: {e}")
            time.sleep(10)

if __name__ == "__main__":
    main_loop()
