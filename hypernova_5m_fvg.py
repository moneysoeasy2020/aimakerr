import time, os, math, requests
from eth_account import Account
from eth_account.messages import encode_defunct

SECRET_KEY = os.getenv("HL_SECRET_KEY", "0xd51ab687339ad395fc06f4b65cc6651a20ac3d37351956410eeef65968b25979")
ACCOUNT_ADDRESS = os.getenv("HL_ACCOUNT_ADDRESS", "0xC63708A9E3B674e34E4A983D7c8c929B0b31bc96")

RULES = {
    "INITIAL_BALANCE": 360.12,
    "LEVERAGE": 5,
    "RISK_PER_TRADE_PCT": 0.02, # 调整为 2% 保证满足最小下单名义价值
    "MAX_POSITIONS": 2,
    "TARGET_SYMBOLS": ["BTC", "xyzGOLD", "GOLD", "XAU"],
    "ATR_PERIOD": 14,
    "RR_RATIO": 2.0,
    "VOL_SURGE_RATIO": 1.5,
    "MAX_RANGE_PCT": 0.010,
    "MIN_BODY_RATIO": 0.50
}

s = requests.Session()
s.headers.update({"Content-Type": "application/json"})
agent_wallet = Account.from_key(SECRET_KEY)

def log(msg, level="INFO"):
    print(f"{time.strftime('%Y-%m-%d %H:%M:%S')} - [{level}] - {msg}", flush=True)

def post_req(path, payload):
    try:
        r = s.post(f"https://api.hyperliquid.xyz{path}", json=payload, timeout=12)
        if r.status_code == 200:
            return r.json()
        else:
            log(f"HTTP 错误 {r.status_code}: {r.text}", "ERROR")
            return None
    except Exception as e:
        log(f"网络请求失败: {e}", "ERROR")
        return None

def sign_action(action, nonce):
    msg_str = f"{action}{nonce}"
    signed = agent_wallet.sign_message(encode_defunct(text=msg_str))
    return {"r": hex(signed.r), "s": hex(signed.s), "v": signed.v}

def execute_order(coin, asset_idx, is_buy, sz, limit_px):
    ts = int(time.time() * 1000)
    # 转换为 Hyperliquid 官方标准字符串格式
    px_str = f"{limit_px:.6f}".rstrip('0').rstrip('.')
    sz_str = f"{sz:.6f}".rstrip('0').rstrip('.')
    
    action = {
        "type": "order",
        "orders": [{
            "a": asset_idx,
            "b": is_buy,
            "p": px_str,
            "s": sz_str,
            "r": False,
            "t": {"limit": {"tif": "Gtc"}}
        }],
        "grouping": "na"
    }
    
    # 官方 L1 交互 Payload
    payload = {
        "action": action,
        "nonce": ts,
        "signature": sign_action(action, ts),
        "vaultAddress": None
    }
    return post_req("/exchange", payload)

def get_target_assets_ctx():
    res = post_req("/info", {"type": "metaAndAssetCtxs"})
    if not res or len(res) < 2: return []
    universe, ctxs = res[0]["universe"], res[1]
    
    target_info = []
    for idx, asset in enumerate(universe):
        name = asset["name"]
        if name in RULES["TARGET_SYMBOLS"]:
            target_info.append({
                "symbol": name,
                "sz_decimals": asset["szDecimals"],
                "asset_idx": idx,
                "mark_price": float(ctxs[idx].get("markPx", 0))
            })
    return target_info

def get_account_info():
    res = post_req("/info", {"type": "clearinghouseState", "user": ACCOUNT_ADDRESS})
    val = RULES["INITIAL_BALANCE"]
    active = []
    if res:
        margin_val = float(res.get("marginSummary", {}).get("accountValue", 0))
        if margin_val > 0:
            val = margin_val
        active = [p.get("position", {}).get("coin") for p in res.get("assetPositions", []) if abs(float(p.get("position", {}).get("szi", 0))) > 0]
    return val, active

def get_candles(symbol, interval, count):
    now_ms = int(time.time() * 1000)
    int_ms = 5 * 60 * 1000 if interval == "5m" else 60 * 60 * 1000
    res = post_req("/info", {"type": "candleSnapshot", "req": {"coin": symbol, "interval": interval, "startTime": now_ms - (count * int_ms), "endTime": now_ms}})
    return [{'open': float(c['o']), 'high': float(c['h']), 'low': float(c['l']), 'close': float(c['c']), 'vol': float(c['v'])} for c in res] if res else None

def calc_atr(candles, period=14):
    if len(candles) < period + 1: return 0
    trs = [max(c['high'] - c['low'], abs(c['high'] - candles[i-1]['close']), abs(c['low'] - candles[i-1]['close'])) for i, c in enumerate(candles[1:], 1)]
    return sum(trs[-period:]) / period

def check_signal(c5, c1h):
    if not c5 or len(c5) < 20 or not c1h or len(c1h) < 50: return None, 0
    cl1h = [c['close'] for c in c1h]
    ema50 = cl1h[0]
    for c in cl1h[1:]: ema50 = (c * (2/51)) + (ema50 * (1 - (2/51)))
    trend = 'BULL' if cl1h[-1] > ema50 else 'BEAR'
    
    c = c5[-2]
    cons = c5[-7:-2]
    h_max, l_min = max(k['high'] for k in cons), min(k['low'] for k in cons)
    if ((h_max - l_min) / l_min) > RULES["MAX_RANGE_PCT"]: return None, 0
    k_rnd = c['high'] - c['low']
    if k_rnd == 0 or (abs(c['close'] - c['open']) / k_rnd) < RULES["MIN_BODY_RATIO"]: return None, 0
    if c['vol'] < ((sum(k['vol'] for k in cons) / 5) * RULES["VOL_SURGE_RATIO"]): return None, 0

    bull_fvg = c5[-2]['low'] > c5[-4]['high']
    bear_fvg = c5[-2]['high'] < c5[-4]['low']
    cl5 = [k['close'] for k in c5]
    ma5, ma10 = sum(cl5[-6:-1]) / 5, sum(cl5[-11:-1]) / 10
    atr = calc_atr(c5, RULES["ATR_PERIOD"])

    if trend == 'BULL' and ma5 > ma10 and c['close'] > ma5 and c['close'] > h_max and bull_fvg:
        return 'BUY', atr
    elif trend == 'BEAR' and ma5 < ma10 and c['close'] < ma10 and c['close'] < l_min and bear_fvg:
        return 'SELL', atr
    return None, 0

def main():
    log(f"🚀 Hypernova 5m (Agent: {agent_wallet.address}) 修复版挂载...")
    while True:
        try:
            target_assets = get_target_assets_ctx()
            if not target_assets:
                time.sleep(5)
                continue
                
            balance, active_symbols = get_account_info()
            if len(active_symbols) >= RULES["MAX_POSITIONS"]:
                time.sleep(10)
                continue

            for asset in target_assets:
                symbol = asset["symbol"]
                sz_dec = asset["sz_decimals"]
                asset_idx = asset["asset_idx"]
                
                if symbol in active_symbols: 
                    continue

                c5 = get_candles(symbol, "5m", 35)
                c1h = get_candles(symbol, "1h", 60)
                sig, atr = check_signal(c5, c1h)

                if sig and atr > 0:
                    px = c5[-1]['close']
                    sl = max(atr * 1.5, px * 0.005)
                    raw_sz = (balance * RULES["RISK_PER_TRADE_PCT"]) / sl
                    
                    # 格式化小数位数
                    sz = math.floor(raw_sz) if sz_dec == 0 else round(raw_sz, sz_dec)
                    
                    # 确保满足 Hyperliquid 最小 11 美金下单额
                    if sz * px < 11.0:
                        min_sz = 11.0 / px
                        sz = round(min_sz, sz_dec) if sz_dec > 0 else math.ceil(min_sz)

                    is_buy = True if sig == 'BUY' else False
                    log(f"⚡ [触发开仓] {symbol} | 方向: {'买入' if is_buy else '卖出'} | 数量: {sz} | 价格: {px}")
                    res = execute_order(symbol, asset_idx, is_buy, sz, px)
                    log(f"开仓结果: {res}")
                    time.sleep(10) # 触发后休眠，防止频繁对同一信号报单
                    break
                
                time.sleep(0.5)

            time.sleep(10)

        except Exception as e:
            log(f"轮询异常: {e}", "ERROR")
            time.sleep(5)

if __name__ == '__main__':
    main()
