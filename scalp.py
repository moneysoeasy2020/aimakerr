import time
import logging
import os
import ccxt

# ==================== 日志配置 ====================
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - [%(levelname)s] - %(message)s',
    handlers=[
        logging.FileHandler('scalp_strategy.log'),
        logging.StreamHandler()
    ]
)

# ==================== API 密钥注入 (建议后续替换为环境变量) ====================
API_KEY = os.getenv("BINANCE_API_KEY", "tj49IrEjBmedtwK3BEJNCqP1lk7fXgNkpmi5Yr2eNUhEXHG35XZzpow5TYu3S69y")
SECRET_KEY = os.getenv("BINANCE_API_SECRET", "90ibaaGFYMzE6tE9IDqqnGEQZh0qZ76xReZQFt1AHteCRNEs1T76AGh6MDM3Lkno")

# ==================== 策略参数配置 ====================
MAX_CONCURRENT_TRADES = 3      # 最大同时持仓币种数
SINGLE_SYMBOL_MAX_MARGIN = 10  # 单币种最大保证金上限 (USDT)
INITIAL_MARGIN = 3.0           # Phase 1 初始保证金 (USDT)
LEVERAGE = 10                  # 杠杆倍数

# 滚雪球加仓与风控门槛
PYRAMID_PHASE2_THRESHOLD = 0.03 # 浮盈 >= 3% 触发 Phase 2 加仓 (+20% 保证金)
PYRAMID_PHASE3_THRESHOLD = 0.06 # 浮盈 >= 6% 触发 Phase 3 加仓 (+10% 保证金)
BREAKEVEN_THRESHOLD = 0.05      # 浮盈 >= 5% 触发移动保本止损

# ==================== 初始化交易所客户端 ====================
exchange = ccxt.binance({
    'apiKey': API_KEY,
    'secret': SECRET_KEY,
    'timeout': 10000,           # 10秒超时强制中断，防线程卡死
    'enableRateLimit': True,
    'options': {
        'defaultType': 'future'  # 币安 U 本位合约
    }
})

def safe_api_call(func, *args, **kwargs):
    """API 保护壳：网络超时或报错时优雅捕获，防崩溃"""
    try:
        return func(*args, **kwargs)
    except Exception as e:
        logging.error(f"API 调用异常 [{func.__name__}]: {e}")
        return None

def get_active_positions():
    """获取当前所有有仓位的合约信息"""
    positions = safe_api_call(exchange.fetch_positions)
    if not positions:
        return []
    
    active = []
    for pos in positions:
        contracts = float(pos.get('contracts', 0))
        if contracts > 0:
            active.append({
                'symbol': pos['symbol'],
                'entryPrice': float(pos['entryPrice']),
                'markPrice': float(pos['markPrice']),
                'contracts': contracts,
                'notional': float(pos['notional']),
                'initialMargin': float(pos['initialMargin']),
                'unrealizedPnl': float(pos['unrealizedPnl']),
                'percentage': float(pos['percentage']) if pos.get('percentage') else 0.0
            })
    return active

def cleanup_orphan_orders(active_symbols):
    """清理非持仓币种的遗留挂单（防孤儿止损单）"""
    open_orders = safe_api_call(exchange.fetch_open_orders)
    if open_orders is None:
        return
    
    for order in open_orders:
        symbol = order['symbol']
        if symbol not in active_symbols:
            logging.info(f"🧹 清理非持仓标的遗留挂单: {symbol} | OrderID: {order['id']}")
            safe_api_call(exchange.cancel_order, order['id'], symbol)

def update_breakeven_stop(symbol, pos, pnl_pct):
    """完备版移动保本止损：带精度格式化 + 先挂后撤保障"""
    if pnl_pct < BREAKEVEN_THRESHOLD:
        return

    entry_price = pos['entryPrice']
    open_orders = safe_api_call(exchange.fetch_open_orders, symbol)
    if open_orders is None:
        return

    has_breakeven = False
    old_order_ids = []

    for order in open_orders:
        if order.get('type') in ['stop_market', 'STOP_MARKET']:
            stop_price = float(order.get('stopPrice', 0))
            # 若已有止损单价格等于或优于开仓成本价，认定已成功保本
            if stop_price >= entry_price:
                has_breakeven = True
            else:
                old_order_ids.append(order['id'])

    if not has_breakeven:
        # 1. 对止损价格与数量进行交易所精度对齐（防止 Precision Exceeded 拒单）
        formatted_stop_price = exchange.price_to_precision(symbol, entry_price)
        formatted_amount = exchange.amount_to_precision(symbol, pos['contracts'])

        logging.info(f"🛡️ 触发保本止损: {symbol} | 当前浮盈: {pnl_pct*100:.2f}% | 保本止损价: {formatted_stop_price}")
        
        # 2. 先挂设新的保本止损单
        params = {
            'stopPrice': float(formatted_stop_price),
            'reduceOnly': True
        }
        new_order = safe_api_call(
            exchange.create_order,
            symbol=symbol,
            type='STOP_MARKET',
            side='sell',
            amount=float(formatted_amount),
            params=params
        )

        # 3. 确认新单创建成功后，再安全的撤销旧低价止损单
        if new_order:
            logging.info(f"✅ 新保本止损单 (ID: {new_order['id']}) 挂设成功，清理旧止损单...")
            for old_id in old_order_ids:
                safe_api_call(exchange.cancel_order, old_id, symbol)

def check_and_pyramid(symbol, pos, pnl_pct):
    """滚雪球加仓模块：动态对齐最小下下单量"""
    current_margin = pos['initialMargin']
    
    # 单币已达保证金上限，停止加仓
    if current_margin >= (SINGLE_SYMBOL_MAX_MARGIN - 0.5):
        return

    add_margin = 0
    phase_label = ""

    # Phase 2 判定 (+20% 保证金 ≈ 2.0 USDT)
    if pnl_pct >= PYRAMID_PHASE2_THRESHOLD and current_margin < (INITIAL_MARGIN + 1.5):
        add_margin = 2.0
        phase_label = "Phase 2 (加仓 20%)"
    # Phase 3 判定 (+10% 保证金 ≈ 1.0 USDT)
    elif pnl_pct >= PYRAMID_PHASE3_THRESHOLD and current_margin < (INITIAL_MARGIN + 2.5):
        add_margin = 1.0
        phase_label = "Phase 3 (加仓 10%)"

    if add_margin > 0:
        mark_price = pos['markPrice']
        add_notional = add_margin * LEVERAGE
        
        # 币安合约单笔名义价值不得小于 5 USDT，若计算值不足自动对齐至 5.5 USDT
        if add_notional < 5.0:
            add_notional = 5.5

        raw_amount = add_notional / mark_price
        formatted_amount = exchange.amount_to_precision(symbol, raw_amount)

        logging.info(f"📈 [滚雪球 {phase_label}] {symbol} | 浮盈: {pnl_pct*100:.2f}% | 追加名义价值: {add_notional} USDT")
        order = safe_api_call(
            exchange.create_order,
            symbol=symbol,
            type='MARKET',
            side='buy',
            amount=float(formatted_amount)
        )
        if order:
            logging.info(f"🚀 滚雪球加仓成功成交: {symbol}")

# ==================== 主运行 Engine ====================
def main():
    # 自动装载市场的精度规则表 (Market Precision Data)
    exchange.load_markets()
    logging.info(f"⚙️ 滚雪球交易机器人完全体启动 | 精度防护系统已就绪")
    
    while True:
        try:
            # 1. 刷新活跃持仓
            positions = get_active_positions()
            active_symbols = [p['symbol'] for p in positions]

            # 2. 清理平仓遗留挂单
            cleanup_orphan_orders(active_symbols)

            # 3. 轮询持仓：执行【保本止损】与【滚雪球加仓】
            for pos in positions:
                symbol = pos['symbol']
                pnl_pct = pos['percentage'] / 100.0 if pos['percentage'] != 0 else (pos['unrealizedPnl'] / pos['initialMargin'])
                
                # 保本止损巡检
                update_breakeven_stop(symbol, pos, pnl_pct)

                # 滚雪球加仓巡检
                check_and_pyramid(symbol, pos, pnl_pct)

        except Exception as e:
            logging.error(f"❌ 主循环捕获未预期的异常: {e}")

        # 10 秒死循环休眠
        time.sleep(10)

if __name__ == '__main__':
    main()

