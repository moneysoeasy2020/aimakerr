import time
import json
import hmac
import hashlib
import requests
from urllib.parse import urlencode

# ==========================================
# 1. API 秘钥配置
# ==========================================
BINANCE_API_KEY = "tj49IrEjBmedtwK3BEJNCqP1lk7fXgNkpmi5Yr2eNUhEXHG35XZzpow5TYu3S69y"
BINANCE_SECRET_KEY = "90ibaaGFYMzE6tE9IDqqnGEQZh0qZ76xReZQFt1AHteCRNEs1T76AGh6MDM3Lkno"

DEEPSEEK_API_KEY = "sk-c146663d0698459d97bb5987a532d91e"
DEEPSEEK_API_URL = "https://api.deepseek.com/v1/chat/completions"

BINANCE_BASE_URL = "https://fapi.binance.com"  # U本位合约 REST Base URL

# ==========================================
# 2. 交易参数与风控规则 (无预设固定本金)
# ==========================================
TRADING_SYMBOLS = {
    "ETHUSDT": {
        "max_margin_ratio": 0.15,  # 最高使用可用余额的 15% 作保证金
        "leverage": 20,            # 20X 杠杆
        "stop_loss_pct": 0.015,    # 盘面 -1.5% 硬止损
        "tp1_pct": 0.012,          # 盘面 +1.2% 第一止盈点
        "tp1_close_ratio": 0.65,
        "quantity_precision": 3    # ETH 下单数量精度
    },
    "ZECUSDT": {
        "max_margin_ratio": 0.05,  # 最高使用可用余额的 5% 作保证金
        "leverage": 10,            # 10X 杠杆
        "stop_loss_pct": 0.025,    # 盘面 -2.5% 硬止损
        "tp1_pct": 0.020,          # 盘面 +2.0% 第一止盈点
        "tp1_close_ratio": 0.65,
        "quantity_precision": 1    # ZEC 下单数量精度
    }
}

# ==========================================
# 3. 币安原生签名与 HTTP 请求底层引擎
# ==========================================
def _send_signed_request(http_method, endpoint, params=None):
    if params is None:
        params = {}
    
    # 增加币安 API 必需的时间戳
    params['timestamp'] = int(time.time() * 1000)
    query_string = urlencode(params)
    
    # 生成 HMAC SHA256 签名
    signature = hmac.new(
        BINANCE_SECRET_KEY.encode('utf-8'),
        query_string.encode('utf-8'),
        hashlib.sha256
    ).hexdigest()
    
    full_url = f"{BINANCE_BASE_URL}{endpoint}?{query_string}&signature={signature}"
    headers = {"X-MBX-APIKEY": BINANCE_API_KEY}
    
    try:
        if http_method.upper() == "GET":
            response = requests.get(full_url, headers=headers, timeout=5)
        elif http_method.upper() == "POST":
            response = requests.post(full_url, headers=headers, timeout=5)
        return response.json()
    except Exception as e:
        print(f"❌ 币安 API 请求异常 ({endpoint}): {e}")
        return None

def get_futures_balance():
    """读取合约账户中可用 USDT 余额"""
    res = _send_signed_request("GET", "/fapi/v2/balance")
    if isinstance(res, list):
        for item in res:
            if item.get('asset') == 'USDT':
                return float(item.get('availableBalance', 0.0))
    print(f"❌ 获取账户余额失败，返回值: {res}")
    return 0.0

def set_leverage(symbol, leverage):
    """设置标的杠杆"""
    params = {"symbol": symbol, "leverage": leverage}
    return _send_signed_request("POST", "/fapi/v1/leverage", params)

# ==========================================
# 4. 原生指标计算引擎 (RSI & EMA)
# ==========================================
def get_klines_data(symbol, interval="15m", limit=100):
    """获取 K 线并使用原生算法计算 RSI 与 EMA"""
    url = f"{BINANCE_BASE_URL}/fapi/v1/klines?symbol={symbol}&interval={interval}&limit={limit}"
    try:
        res = requests.get(url, timeout=5).json()
        closes = [float(k[4]) for k in res]
        
        if len(closes) < 50:
            return None
        
        current_price = closes[-1]
        
        # 1. 计算 RSI(14)
        gains, losses = [], []
        for i in range(1, len(closes)):
            change = closes[i] - closes[i-1]
            gains.append(change if change > 0 else 0)
            losses.append(-change if change < 0 else 0)
        
        avg_gain = sum(gains[-14:]) / 14
        avg_loss = sum(losses[-14:]) / 14
        
        if avg_loss == 0:
            rsi = 100
        else:
            rs = avg_gain / avg_loss
            rsi = 100 - (100 / (1 + rs))
            
        # 2. 计算 EMA(20) 与 EMA(50)
        def calc_ema(period, prices):
            k = 2 / (period + 1)
            ema = prices[0]
            for p in prices[1:]:
                ema = (p * k) + (ema * (1 - k))
            return ema

        ema20 = calc_ema(20, closes)
        ema50 = calc_ema(50, closes)
        
        return {
            "price": current_price,
            "rsi": rsi,
            "ema20": ema20,
            "ema50": ema50
        }
    except Exception as e:
        print(f"❌ [{symbol}] 行情拉取失败: {e}")
        return None

# ==========================================
# 5. DeepSeek API 语义风控引擎
# ==========================================
def ask_deepseek_advisor(symbol, current_price, rsi, ema20, ema50, market_trend):
    """调用 DeepSeek 判定入场风险"""
    prompt = f"""
    你是一个精通加密货币量化交易的风控专家。
    标的: {symbol}
    现价: {current_price}
    15m RSI(14): {rsi:.2f}
    EMA20: {ema20:.2f}, EMA50: {ema50:.2f}
    拟定方向: {market_trend}
    
    原则：只做高盈亏比波段，不死扛不摸顶。
    判断当前指标是否适合建立 {market_trend} 仓位。
    输出 JSON 格式（严格禁止包含额外文字）：
    {{"action": "BUY" / "SELL" / "HOLD", "reason": "简短分析原因"}}
    """
    headers = {
        "Authorization": f"Bearer {DEEPSEEK_API_KEY}",
        "Content-Type": "application/json"
    }
    payload = {
        "model": "deepseek-chat",
        "messages": [{"role": "user", "content": prompt}],
        "temperature": 0.1
    }
    try:
        response = requests.post(DEEPSEEK_API_URL, headers=headers, json=payload, timeout=8)
        content = response.json()['choices'][0]['message']['content']
        clean_content = content.replace("```json", "").replace("```", "").strip()
        return json.loads(clean_content)
    except Exception as e:
        print(f"⚠️ DeepSeek API 异常 ({e})，降级为纯技术指标。")
        return {"action": market_trend, "reason": "API Fallback"}

# ==========================================
# 6. 原生币安实盘下单与挂单处理
# ==========================================
def execute_live_trade(symbol, side, current_balance, current_price):
    config = TRADING_SYMBOLS[symbol]
    
    # 1. 自动设置杠杆
    set_leverage(symbol, config["leverage"])
    
    # 2. 动态计算开仓量
    margin_usdt = current_balance * config["max_margin_ratio"]
    notional_value = margin_usdt * config["leverage"]
    quantity = round(notional_value / current_price, config["quantity_precision"])
    
    if quantity <= 0:
        print(f"⚠️ [{symbol}] 计算出的下单量过小，取消开仓。")
        return

    print(f"\n🚀 准备提交 [{symbol}] 币安实盘订单:")
    print(f" ├─ 方向: {side}")
    print(f" ├─ 动态保证金: ${margin_usdt:.2f} USDT ({config['leverage']}X)")
    print(f" └─ 下单数量: {quantity} {symbol.replace('USDT','')}")

    # 3. 发送市价开仓单
    entry_params = {
        "symbol": symbol,
        "side": side,
        "type": "MARKET",
        "quantity": quantity
    }
    res_entry = _send_signed_request("POST", "/fapi/v1/order", entry_params)
    
    if res_entry and "orderId" in res_entry:
        print(f"✅ 市价开仓成功！订单 ID: {res_entry['orderId']}")
        
        # 4. 设置硬止损 (STOP_MARKET)
        sl_pct = config["stop_loss_pct"]
        sl_price = round(current_price * (1 - sl_pct) if side == "BUY" else current_price * (1 + sl_pct), 2)
        close_side = "SELL" if side == "BUY" else "BUY"
        
        sl_params = {
            "symbol": symbol,
            "side": close_side,
            "type": "STOP_MARKET",
            "stopPrice": sl_price,
            "closePosition": "true"
        }
        _send_signed_request("POST", "/fapi/v1/order", sl_params)
        print(f"🛡️ 自动硬止损挂单成功，触发价: ${sl_price}")
        
        # 5. 设置分批止盈 (TAKE_PROFIT_MARKET - 平仓 65%)
        tp_pct = config["tp1_pct"]
        tp_price = round(current_price * (1 + tp_pct) if side == "BUY" else current_price * (1 - tp_pct), 2)
        tp_quantity = round(quantity * config["tp1_close_ratio"], config["quantity_precision"])
        
        tp_params = {
            "symbol": symbol,
            "side": close_side,
            "type": "TAKE_PROFIT_MARKET",
            "stopPrice": tp_price,
            "quantity": tp_quantity
        }
        _send_signed_request("POST", "/fapi/v1/order", tp_params)
        print(f"🎯 首波止盈挂单成功 (65%)，触发价: ${tp_price}")
        print("--------------------------------------------------\n")
    else:
        print(f"❌ 开仓失败，币安返回信息: {res_entry}")

# ==========================================
# 7. 主监控循环
# ==========================================
def main():
    print("==========================================================")
    print(" ETH & ZEC 币安实盘量化脚本 (无开源库依赖版) 已启动 ")
    print("==========================================================")
    
    balance = get_futures_balance()
    print(f"💰 实时读取到币安 U本位合约可用余额: ${balance:.2f} USDT\n")
    
    if balance < 10:
        print("⚠️ 合约账户可用余额低于 10 USDT，请先充值或划转资金！")
        return

    while True:
        try:
            current_balance = get_futures_balance()
            
            for symbol in TRADING_SYMBOLS.keys():
                data = get_klines_data(symbol, interval="15m", limit=100)
                if not data:
                    continue
                
                price = data['price']
                rsi = data['rsi']
                ema20 = data['ema20']
                ema50 = data['ema50']
                
                # 技术面初筛
                tech_signal = "HOLD"
                if rsi > 70 and ema20 < ema50:
                    tech_signal = "SELL" # 极度超买，趋势向下
                elif rsi < 30 and ema20 > ema50:
                    tech_signal = "BUY"  # 极度超卖，趋势向上
                    
                print(f"[{time.strftime('%H:%M:%S')}] {symbol} | 现价: {price:.2f} | RSI: {rsi:.1f} | 信号: {tech_signal}")
                
                if tech_signal != "HOLD":
                    print(f"  👉 触发预警，请求 DeepSeek API 决策中...")
                    ai_res = ask_deepseek_advisor(symbol, price, rsi, ema20, ema50, tech_signal)
                    action = ai_res.get("action")
                    reason = ai_res.get("reason")
                    
                    print(f"  🤖 DeepSeek 决策: {action} | 分析: {reason}")
                    
                    if action in ["BUY", "SELL"]:
                        execute_live_trade(symbol, action, current_balance, price)
                        time.sleep(10)
                
            time.sleep(60) # 60 秒轮询一次
            
        except Exception as e:
            print(f"主循环捕获异常: {e}")
            time.sleep(10)

if __name__ == "__main__":
    main()
