import ccxt
import time
import sys

# ==================== 1. 基础配置 ====================
API_KEY = "tj49IrEjBmedtwK3BEJNCqP1lk7fXgNkpmi5Yr2eNUhEXHG35XZzpow5TYu3S69y"
SECRET_KEY = "90ibaaGFYMzE6tE9IDqqnGEQZh0qZ76xReZQFt1AHteCRNEs1T76AGh6MDM3Lkno"

# 监控标的池（全词条对应币安 USDT 永续合约）
SYMBOL_POOL = [
    'LSK/USDT:USDT', 'FIGHT/USDT:USDT', 'QNT/USDT:USDT', 
    'RARE/USDT:USDT', 'BTW/USDT:USDT', 'INX/USDT:USDT'
]

MARGIN_PER_TRADE = 3.8   # 单笔固定本金 3.8 USDT
LEVERAGE = 5             # 固定 5 倍杠杆
MAX_POSITIONS = 2        # 同时最大持仓数 2 个

STOP_LOSS_PCT = 0.015    # 硬止损：1.5%

# 初始化币安合约接口
exchange = ccxt.binance({
    'apiKey': API_KEY,
    'secret': SECRET_KEY,
    'options': {'defaultType': 'future'},
    'enableRateLimit': True,
})

# ==================== 2. 核心逻辑函数 ====================

def init_symbol_config(symbol):
    """设置单币种的杠杆与单向持仓模式"""
    try:
        exchange.set_leverage(LEVERAGE, symbol)
        # 设置为单向持仓模式 (One-way Mode)
        exchange.set_position_mode(hedged=False, symbol=symbol)
    except Exception:
        pass  # 若已设置过则忽略错误

def get_active_positions():
    """读取实盘活动持仓（严格控制持仓总数）"""
    try:
        balance = exchange.fetch_balance()
        positions = balance['info']['positions']
        active_symbols = []
        for p in positions:
            amt = float(p['positionAmt'])
            if amt != 0:
                active_symbols.append(p['symbol'])
        return active_symbols
    except Exception as e:
        print(f"\n[!] 获取持仓失败: {e}")
        return []

def calculate_rsi(closes, period=14):
    """Termux 轻量化原生 RSI 计算（无 pandas 依赖）"""
    if len(closes) < period + 1:
        return 50
    gains, losses = [], []
    for i in range(1, len(closes)):
        change = closes[i] - closes[i - 1]
        if change >= 0:
            gains.append(change)
            losses.append(0)
        else:
            gains.append(abs(change))
            losses.append(change)
    
    avg_gain = sum(gains[-period:]) / period
    avg_loss = sum(losses[-period:]) / period
    
    if avg_loss == 0:
        return 100
    rs = avg_gain / avg_loss
    return 100 - (100 / (1 + rs))

def check_entry_signal(symbol):
    """
    超买高位回调做空信号判定
    - 获取 5 分钟 K 线
    - 条件：上一根 K 线 RSI > 65 且 当前 K 线收阴（出现高位回落）
    """
    try:
        ohlcv = exchange.fetch_ohlcv(symbol, timeframe='5m', limit=30)
        closes = [bar[4] for bar in ohlcv]
        
        # 计算 RSI
        rsi_prev = calculate_rsi(closes[:-1], period=14)
        
        # 当前及前一根 K 线形态
        prev_close = closes[-2]
        prev_open = ohlcv[-2][1]
        last_close = closes[-1]
        last_open = ohlcv[-1][1]
        
        # 触发条件：前一根为高位超买，且当前 K 线发生回落 (收阴)
        if rsi_prev > 65 and (last_close < last_open):
            return True
        return False
    except Exception as e:
        return False

def open_short_position(symbol):
    """执行做空开仓并自动下发 Stop-Loss 挂单"""
    try:
        init_symbol_config(symbol)
        ticker = exchange.fetch_ticker(symbol)
        entry_price = ticker['last']
        
        # 换算下单数量 (3.8 * 5 = 19 USDT 名义价值)
        notional_value = MARGIN_PER_TRADE * LEVERAGE
        amount = notional_value / entry_price
        
        # 交易所精度修正
        amount_str = exchange.amount_to_precision(symbol, amount)
        sl_price = exchange.price_to_precision(symbol, entry_price * (1 + STOP_LOSS_PCT))
        
        print(f"\n[+] 触发做空信号: {symbol} | 当前价: {entry_price}")
        
        # 1. 市价做空
        order = exchange.create_market_sell_order(symbol, float(amount_str))
        print(f"    └─ 开仓成功: {order['id']}")
        
        # 2. 自动挂设硬止损单 (STOP_MARKET)
        sl_order = exchange.create_order(
            symbol=symbol,
            type='STOP_MARKET',
            side='buy',
            amount=float(amount_str),
            params={
                'stopPrice': float(sl_price),
                'reduceOnly': True
            }
        )
        print(f"    └─ 硬止损已设置: {sl_price} USDT")
        
    except Exception as e:
        print(f"\n[-] 开仓失败 [{symbol}]: {e}")

# ==================== 3. Termux 主循环 ====================

def main():
    print("==========================================")
    print("  Termux 高频做空策略机器人已启动")
    print(f"  单笔本金: ${MARGIN_PER_TRADE} | 杠杆: {LEVERAGE}x | 限制持仓: {MAX_POSITIONS}个")
    print("==========================================")
    
    while True:
        try:
            # 1. 获取实盘持仓
            active_positions = get_active_positions()
            pos_count = len(active_positions)
            
            sys.stdout.write(f"\r[监控中] 当前持仓: {pos_count}/{MAX_POSITIONS} | 扫描标的中...")
            sys.stdout.flush()
            
            # 2. 若持仓未满 2 个，继续寻找做空标的
            if pos_count < MAX_POSITIONS:
                for symbol in SYMBOL_POOL:
                    # 规避已持仓币种
                    clean_symbol = symbol.replace('/', '').split(':')[0]
                    if any(clean_symbol in p for p in active_positions):
                        continue
                        
                    if check_entry_signal(symbol):
                        open_short_position(symbol)
                        time.sleep(2)
                        
                        # 开仓后重新刷新持仓
                        active_positions = get_active_positions()
                        if len(active_positions) >= MAX_POSITIONS:
                            break
                            
            time.sleep(8)  # 轮询刷新间隔
            
        except KeyboardInterrupt:
            print("\n用户手动停止机器人。")
            break
        except Exception as e:
            print(f"\n[!] 运行异常: {e}")
            time.sleep(10)

if __name__ == '__main__':
    main()
