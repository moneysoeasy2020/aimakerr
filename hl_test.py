import time, os, requests
from eth_account import Account
from eth_account.messages import encode_defunct

# 配置信息
SECRET_KEY = os.getenv("HL_SECRET_KEY", "0xd51ab687339ad395fc06f4b65cc6651a20ac3d37351956410eeef65968b25979")
ACCOUNT_ADDRESS = os.getenv("HL_ACCOUNT_ADDRESS", "0xC63708A9E3B674e34E4A983D7c8c929B0b31bc96")

s = requests.Session()
s.headers.update({"Content-Type": "application/json"})
agent_wallet = Account.from_key(SECRET_KEY)

def post_req(path, payload):
    try:
        r = s.post(f"https://api.hyperliquid.xyz{path}", json=payload, timeout=10)
        return r.status_code, r.json()
    except Exception as e:
        return 0, str(e)

def sign_action(action, nonce):
    import msgpack
    data = {"action": action, "nonce": nonce}
    packed = msgpack.packb(data, use_bin_type=True)
    msg_hash = encode_defunct(primitive=packed)
    signed = agent_wallet.sign_message(msg_hash)
    return {"r": hex(signed.r), "s": hex(signed.s), "v": signed.v}

def run_test():
    print("========================================")
    print("🔍 开始 Hyperliquid API & 签名诊断测试")
    print("========================================")
    print(f"📌 主账户地址: {ACCOUNT_ADDRESS}")
    print(f"📌 Agent 钱包地址: {agent_wallet.address}")
    print("----------------------------------------")

    # 1. 账户余额与持仓测试
    print("\n[测试 1/3] 正在查询账户余额...")
    code, res = post_req("/info", {"type": "clearinghouseState", "user": ACCOUNT_ADDRESS})
    if code == 200:
        val = res.get("marginSummary", {}).get("accountValue", "0")
        withdrawable = res.get("withdrawable", "0")
        positions = res.get("assetPositions", [])
        print(f" ✅ 成功查到余额！")
        print(f"    - 账户总权益: ${val} USDC")
        print(f"    - 可用/可提现: ${withdrawable} USDC")
        print(f"    - 当前持仓数: {len(positions)}")
    else:
        print(f" ❌ 余额查询失败! HTTP Status: {code}, 返回内容: {res}")
        return

    # 2. 获取 BTC 交易对参数
    print("\n[测试 2/3] 正在读取 BTC 市场实时数据...")
    code, res = post_req("/info", {"type": "metaAndAssetCtxs"})
    btc_idx, btc_price, sz_dec = 0, 0, 4
    if code == 200:
        universe, ctxs = res[0]["universe"], res[1]
        for idx, asset in enumerate(universe):
            if asset["name"] == "BTC":
                btc_idx = idx
                sz_dec = asset["szDecimals"]
                btc_price = float(ctxs[idx].get("markPx", 0))
                break
        print(f" ✅ 成功获取行情！")
        print(f"    - BTC 序号: {btc_idx} | 当前标记价: ${btc_price}")
    else:
        print(f" ❌ 行情获取失败! HTTP Status: {code}, 返回内容: {res}")
        return

    # 3. 签名与测试下单 (挂一个远离现价的买单，测试完立即撤单)
    print("\n[测试 3/3] 正在测试签名与下单（挂深虚买单测试）...")
    # 挂一个低于现价 20% 的买单（例如现价 $76000，挂 $60000），确保不会真的成交
    test_px = round(btc_price * 0.8, 1)
    # 算出一个略大于 $11 最低限制的数量
    test_sz = round(12.0 / test_px, sz_dec)
    
    ts = int(time.time() * 1000)
    action = {
        "type": "order",
        "orders": [{
            "a": btc_idx,
            "b": True, # 买入
            "p": str(test_px),
            "s": str(test_sz),
            "r": False,
            "t": {"limit": {"tif": "Gtc"}}
        }],
        "grouping": "na"
    }

    payload = {
        "action": action,
        "nonce": ts,
        "signature": sign_action(action, ts),
        "vaultAddress": None,
        "agentAddress": agent_wallet.address
    }

    code, order_res = post_req("/exchange", payload)
    print(f" 📩 服务器响应状态: HTTP {code}")
    print(f" 📩 返回 Payload: {order_res}")

    if code == 200 and order_res.get("status") == "ok":
        statuses = order_res.get("response", {}).get("data", {}).get("statuses", [{}])
        if "resting" in statuses[0]:
            oid = statuses[0]["resting"]["oid"]
            print(f" 🎉 下单成功！订单 ID (OID): {oid}")
            
            # 自动撤单清理测试挂单
            print(" 🧹 正在自动撤销测试挂单...")
            cancel_action = {
                "type": "cancel",
                "cancels": [{"a": btc_idx, "o": oid}]
            }
            ts_c = int(time.time() * 1000)
            cancel_payload = {
                "action": cancel_action,
                "nonce": ts_c,
                "signature": sign_action(cancel_action, ts_c),
                "vaultAddress": None,
                "agentAddress": agent_wallet.address
            }
            _, cancel_res = post_req("/exchange", cancel_payload)
            print(f" 🧹 撤单结果: {cancel_res}")
        else:
            print(f" ⚠️ 接口接收成功，但订单被拒绝: {statuses}")
    else:
        print(f" ❌ 下单签名验证失败！请检查 Agent 是否已在网页端完成 Approve/绑定。")

if __name__ == '__main__':
    run_test()
