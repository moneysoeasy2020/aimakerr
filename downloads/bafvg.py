#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
==============================================================================
Termux 专用高级量化脚本：bafvg.py (纯标准库无依赖版)
包含：FVG + 订单流 + 多周期共振 + 金字塔推仓 + ATR阶梯动态止损
==============================================================================
[环境说明] 零第三方库依赖 (不用 pandas/numpy/ccxt)，原生 Python 3 直接运行
"""

import os
import sys
import time
import json
import hmac
import hashlib
import signal
import atexit
import logging
from urllib.request import Request, urlopen
from urllib.parse import urlencode
from urllib.error import HTTPError, URLError

# ==========================================
# 1. Termux 环境初始化与代码前缀 (PID 锁与日志)
# ==========================================
SCRIPT_NAME = "bafvg"
PID_FILE = f"/tmp/{SCRIPT_NAME}.pid"
LOG_FILE = f"{SCRIPT_NAME}.log"

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    handlers=[
        logging.FileHandler(LOG_FILE, encoding='utf-8'),
        logging.StreamHandler(sys.stdout)
    ]
)

def cleanup_pid():
    """退出时自动清理 PID 文件"""
    if os.path.exists(PID_FILE):
        os.remove(PID_FILE)
        logging.info("[Termux] 策略已安全退出，PID 锁文件已清理")[span_0](start_span)[span_0](end_span)

def handle_signal(sig, frame):
    """响应 SIGINT / SIGTERM 信号实现优雅退出"""
    logging.info(f"[Termux] 接收到退出信号 ({sig})，正在结束进程...")
    sys.exit(0)

atexit.register(cleanup_pid)
signal.signal(signal.SIGINT, handle_signal)
signal.signal(signal.SIGTERM, handle_signal)

if os.path.exists(PID_FILE):
    with open(PID_FILE, 'r') as f:
        old_pid = f.read().strip()
    logging.warning(f"[Termux] 警告：找到正在运行的 PID 锁 ({old_pid})，程序退出。")[span_1](start_span)[span_1](end_span)
    sys.exit(1)

with open(PID_FILE, 'w') as f:
    f.write(str(os.getpid()))

logging.info(f"[Termux] 策略进程启动成功 | PID: {os.getpid()}")[span_2](start_span)[span_2](end_span)


# ==========================================
# 2. 策略全局参数与 API 密钥
# ==========================================
API_KEY = os.getenv("BINANCE_API_KEY", "YOUR_BINANCE_API_KEY")
SECRET_KEY = os.getenv("BINANCE_SECRET_KEY", "YOUR_BINANCE_SECRET_KEY")

BASE_URL = "https://fapi.binance.com"

MAX_OPEN_POSITIONS = 2         # 最大持仓币种数[span_3](start_span)[span_3](end_span)
LEVERAGE = 10                  # 杠杆倍数[span_4](start_span)[span_4](end_span)
BASE_POSITION_PCT = 0.30       # Phase 1 首仓比例 (30%)[span_5](start_span)[span_5](end_span)
ADD_POS_1_PCT = 0.20           # Phase 2 浮盈 +3% 加仓 (20%)[span_6](start_span)[span_6](end_span)
ADD_POS_2_PCT = 0.10           # Phase 3 浮盈 +6% 加仓 (10%)[span_7](start_span)[span_7](end_span)

COOLDOWN_STOP_LOSS = 6 * 3600  # 止损后冷却 6 小时[span_8](start_span)[span_8](end_span)
COOLDOWN_TRADE = 12 * 3600     # 交易后冷却 12 小时[span_9](start_span)[span_9](end_span)

MIN_OBI = 0.25                 # OBI 深度不平衡度 >= 0.25[span_10](start_span)[span_10](end_span)
MIN_TAKER_BUY_RATIO = 0.55     # Taker 主买占比 >= 55%[span_11](start_span)[span_11](end_span)
MIN_FVG_PCT = 0.003            # FVG 机构缺口 >= 0.3%


# ==========================================
# 3. 原生 Binance API 客户端 (替代 ccxt)
# ==========================================
class BinanceClient:
    def __init__(self, api_key, secret_key):
        self.api_key = api_key
        self.secret_key = secret_key

    def _sign(self, query_string):
        return hmac.new(
            self.secret_key.encode('utf-8'),
            query_string.encode('utf-8'),
            hashlib.sha256
        ).hexdigest()

    def _request(self, method, endpoint, params=None, signed=False):
        if params is None:
            params = {}
        
        if signed:
            params['timestamp'] = int(time.time() * 1000)
            query_string = urlencode(params)
            signature = self._sign(query_string)
            url = f"{BASE_URL}{endpoint}?{query_string}&signature={signature}"
        else:
            query_string = urlencode(params)
            url = f"{BASE_URL}{endpoint}" + (f"?{query_string}" if query_string else "")

        req = Request(url, method=method)
        req.add_header('X-MBX-APIKEY', self.api_key)
        
        try:
            with urlopen(req, timeout=10) as resp:
                return json.loads(resp.read().decode('utf-8'))
        except Exception as e:
            logging.error(f"API 请求失败 [{endpoint}]: {e}")
            return None

    def get_klines(self, symbol, interval, limit=50):
        """获取 K 线数据，返回字典列表"""
        data = self._request('GET', '/fapi/v1/klines', {'symbol': symbol, 'interval': interval, 'limit': limit})
        if not data:
            return []
        klines = []
        for item in data:
            klines.append({
                'open': float(item[1]),
                'high': float(item[2]),
                'low': float(item[3]),
                'close': float(item[4]),
                'volume': float(item[5])
            })
        return klines

    def get_orderbook(self, symbol, limit=10):
        return self._request('GET', '/fapi/v1/depth', {'symbol': symbol, 'limit': limit})

    def get_recent_trades(self, symbol, limit=100):
        return self._request('GET', '/fapi/v1/trades', {'symbol': symbol, 'limit': limit})

    def get_ticker_24hr(self):
        return self._request('GET', '/fapi/v1/ticker/24hr')

    def get_usdt_balance(self):
        data = self._request('GET', '/fapi/v2/balance', signed=True)
        if not data:
            return 0.0
        for asset in data:
            if asset.get('asset') == 'USDT':
                return float(asset.get('availableBalance', 0.0))
        return 0.0

    def set_leverage(self, symbol, leverage):
        return self._request('POST', '/fapi/v1/leverage', {'symbol': symbol, 'leverage': leverage}, signed=True)

    def create_market_order(self, symbol, side, quantity):
        params = {
            'symbol': symbol,
            'side': side,
            'type': 'MARKET',
            'quantity': f"{quantity:.3f}"
        }
        return self._request('POST', '/fapi/v1/order', params, signed=True)


# ==========================================
# 4. 动态移动止损与 FVG 风控模块
# ==========================================
class TrailingStopManager:
    def __init__(self, symbol, entry_price, side="LONG", fvg_bottom=None, initial_atr=0.0):
        self.symbol = symbol
        self.entry_price = entry_price
        self.side = side
        self.highest_price = entry_price
        self.is_breakeven_active = False

        if side == "LONG":
            if fvg_bottom and fvg_bottom < entry_price:
                self.stop_loss = fvg_bottom * 0.999
            else:
                self.stop_loss = entry_price - (initial_atr * 1.5) if initial_atr > 0 else entry_price * 0.985
        else:
            self.stop_loss = entry_price * 1.015

    def update(self, current_price, current_atr):
        if self.side == "LONG":
            if current_price > self.highest_price:
                self.highest_price = current_price

            profit_pct = (self.highest_price - self.entry_price) / self.entry_price

            if profit_pct >= 0.03 and not self.is_breakeven_active:
                self.stop_loss = max(self.stop_loss, self.entry_price * 1.001)
                self.is_breakeven_active = True

            if 0.05 <= profit_pct < 0.08:
                new_stop = self.highest_price - (current_atr * 2.0)
                self.stop_loss = max(self.stop_loss, new_stop)
            elif profit_pct >= 0.08:
                new_stop = self.highest_price - (current_atr * 1.2)
                self.stop_loss = max(self.stop_loss, new_stop)

            if current_price <= self.stop_loss:
                reason = "STOP_LOSS_EXIT" if not self.is_breakeven_active else "TRAILING_TAKE_PROFIT_EXIT"
                return True, reason, self.stop_loss

        return False, "HOLD", self.stop_loss


# ==========================================
# 5. 核心交易引擎 (无 math/pandas 依赖纯算术实现)
# ==========================================
class TradingEngine:
    def __init__(self):
        self.client = BinanceClient(API_KEY, SECRET_KEY)
        self.positions = {}
        self.cooldowns = {}

    def check_cooldown(self, symbol):
        if symbol in self.cooldowns and time.time() < self.cooldowns[symbol]:
            return False
        return True

    def set_cooldown(self, symbol, duration_seconds):
        self.cooldowns[symbol] = time.time() + duration_seconds
        logging.info(f"[{symbol}] 进入冷却状态 {duration_seconds/3600:.1f} 小时")

    def calculate_atr(self, klines, period=14):
        """纯 Python 列表计算 ATR"""
        if len(klines) < period + 1:
            return 0.0
        trs = []
        for i in range(1, len(klines)):
            h = klines[i]['high']
            l = klines[i]['low']
            prev_c = klines[i-1]['close']
            tr = max(h - l, abs(h - prev_c), abs(l - prev_c))
            trs.append(tr)
        return sum(trs[-period:]) / period

    def check_btc_trend_1h(self):
        """原生计算 BTC 1H EMA20"""
        klines = self.client.get_klines('BTCUSDT', '1h', limit=50)
        if not klines:
            return "NEUTRAL"
        
        closes = [k['close'] for k in klines]
        alpha = 2 / (20 + 1)
        ema = closes[0]
        for c in closes[1:]:
            ema = (c * alpha) + (ema * (1 - alpha))
            
        return "BULLISH" if closes[-1] > ema else "BEARISH"

    def fetch_order_flow_metrics(self, symbol):
        """原生计算 OBI 与 Taker 买盘比"""
        try:
            depth = self.client.get_orderbook(symbol, limit=10)
            if not depth:
                return 0.0, 0.5
            
            bid_vol = sum([float(b[1]) for b in depth['bids']])
            ask_vol = sum([float(a[1]) for a in depth['asks']])
            obi = (bid_vol - ask_vol) / (bid_vol + ask_vol) if (bid_vol + ask_vol) > 0 else 0

            trades = self.client.get_recent_trades(symbol, limit=100)
            if not trades:
                return obi, 0.5
                
            taker_buy_vol = sum([float(t['qty']) for t in trades if not t['isBuyerMaker']])
            total_vol = sum([float(t['qty']) for t in trades])
            taker_buy_ratio = taker_buy_vol / total_vol if total_vol > 0 else 0.5

            return obi, taker_buy_ratio
        except Exception as e:
            logging.error(f"[{symbol}] 获取订单流失败: {e}")
            return 0.0, 0.5

    def detect_fvg_15m(self, klines_15m):
        """15m 突破 FVG 条件检测"""
        if len(klines_15m) < 3:
            return False, 0.0, 0.0

        k1 = klines_15m[-3]
        k3 = klines_15m[-1]

        if k3['low'] > k1['high']:
            fvg_gap = (k3['low'] - k1['high']) / k1['high']
            if fvg_gap >= MIN_FVG_PCT:
                return True, k1['high'], k3['low']

        return False, 0.0, 0.0

    def check_entry_signal(self, symbol, btc_trend):
        """四重共振条件计算"""
        if not self.check_cooldown(symbol):
            return None, 0.0

        klines_1h = self.client.get_klines(symbol, '1h', limit=30)
        klines_15m = self.client.get_klines(symbol, '15m', limit=30)
        if not klines_1h or not klines_15m:
            return None, 0.0

        current_price = klines_15m[-1]['close']

        # 1. 1H 高点突破意向
        h1_highs = [k['high'] for k in klines_1h[-21:-1]]
        h1_near_breakout = (current_price >= max(h1_highs) * 0.995) if h1_highs else False

        # 2. 15m 爆发突破
        m15_highs = [k['high'] for k in klines_15m[-21:-1]]
        m15_vols = [k['volume'] for k in klines_15m[-6:-1]]
        m15_vol_ma5 = sum(m15_vols) / len(m15_vols) if m15_vols else 1.0
        
        m15_breakout = (current_price > max(m15_highs)) if m15_highs else False
        m15_volume_spike = (klines_15m[-1]['volume'] > m15_vol_ma5 * 1.5)

        # 3. 订单流
        obi, taker_ratio = self.fetch_order_flow_metrics(symbol)
        flow_pass = (obi >= MIN_OBI) and (taker_ratio >= MIN_TAKER_BUY_RATIO)

        # 4. FVG 确认
        has_fvg, fvg_bottom, fvg_top = self.detect_fvg_15m(klines_15m)

        if (btc_trend == "BULLISH" and h1_near_breakout and 
            m15_breakout and m15_volume_spike and flow_pass and has_fvg):
            logging.info(f"[{symbol}] 捕获四重共振突破信号！")
            return "LONG", fvg_bottom

        return None, 0.0

    def execute_market_order(self, symbol, side, amount_usdt):
        try:
            klines = self.client.get_klines(symbol, '1m', limit=1)
            if not klines:
                return None, 0, 0
            price = klines[-1]['close']
            quantity = amount_usdt / price

            self.client.set_leverage(symbol, LEVERAGE)
            order = self.client.create_market_order(symbol, side, quantity)
            logging.info(f"[{symbol}] 执行 {side} 成功 | 金额: {amount_usdt:.2f} USDT | 价格: {price}")
            return order, price, quantity
        except Exception as e:
            logging.error(f"[{symbol}] 下单失败: {e}")
            return None, 0, 0

    def manage_positions(self):
        """持仓管理与金字塔推仓"""
        for symbol in list(self.positions.keys()):
            pos = self.positions[symbol]
            klines_15m = self.client.get_klines(symbol, '15m', limit=30)
            if not klines_15m:
                continue

            current_price = klines_15m[-1]['close']
            current_atr_15m = self.calculate_atr(klines_15m)
            entry_price = pos['avg_entry_price']
            profit_pct = (current_price - entry_price) / entry_price if pos['side'] == "LONG" else (entry_price - current_price) / entry_price

            ts_manager = pos['ts_manager']
            should_exit, reason, stop_price = ts_manager.update(current_price, current_atr_15m)

            if should_exit:
                logging.info(f"[{symbol}] 平仓触发 [{reason}] | 价格: {current_price}")
                self.execute_market_order(symbol, "SELL" if pos['side'] == "LONG" else "BUY", pos['total_usdt'])
                
                if "STOP_LOSS" in reason:
                    self.set_cooldown(symbol, COOLDOWN_STOP_LOSS)[span_12](start_span)[span_12](end_span)
                else:
                    self.set_cooldown(symbol, COOLDOWN_TRADE)[span_13](start_span)[span_13](end_span)
                    
                del self.positions[symbol]
                continue

            # 金字塔推仓 (Phase 2 & Phase 3)[span_14](start_span)[span_14](end_span)
            if pos['phase'] == 1 and profit_pct >= 0.03:
                add_usdt = pos['initial_usdt'] * (ADD_POS_1_PCT / BASE_POSITION_PCT)
                _, p, q = self.execute_market_order(symbol, pos['side'], add_usdt)
                if p > 0:
                    pos['phase'] = 2
                    pos['total_usdt'] += add_usdt
                    pos['avg_entry_price'] = (entry_price * pos['total_qty'] + p * q) / (pos['total_qty'] + q)
                    pos['total_qty'] += q
                    logging.info(f"[{symbol}] +3% 浮盈达成，触发 Phase 2 加仓！")[span_15](start_span)[span_15](end_span)[span_16](start_span)[span_16](end_span)

            elif pos['phase'] == 2 and profit_pct >= 0.06:
                add_usdt = pos['initial_usdt'] * (ADD_POS_2_PCT / BASE_POSITION_PCT)
                _, p, q = self.execute_market_order(symbol, pos['side'], add_usdt)
                if p > 0:
                    pos['phase'] = 3
                    pos['total_usdt'] += add_usdt
                    pos['avg_entry_price'] = (entry_price * pos['total_qty'] + p * q) / (pos['total_qty'] + q)
                    pos['total_qty'] += q
                    logging.info(f"[{symbol}] +6% 浮盈达成，触发 Phase 3 加仓！")[span_17](start_span)[span_17](end_span)[span_18](start_span)[span_18](end_span)

    def run_loop(self):
        logging.info("=== 策略引擎循环启动 (Termux 无依赖零报错模式) ===")[span_19](start_span)[span_19](end_span)
        
        while True:
            try:
                self.manage_positions()

                if len(self.positions) < MAX_OPEN_POSITIONS:
                    btc_trend = self.check_btc_trend_1h()
                    
                    tickers = self.client.get_ticker_24hr()
                    if tickers and isinstance(tickers, list):
                        # 按 24H 成交量排序挑选前 30 个标的
                        usdt_tickers = [t for t in tickers if t['symbol'].endswith('USDT') and 'BTC' not in t['symbol']]
                        sorted_symbols = sorted(usdt_tickers, key=lambda x: float(x.get('quoteVolume', 0)), reverse=True)[:30]
                        symbol_list = [t['symbol'] for t in sorted_symbols]

                        for symbol in symbol_list:
                            if len(self.positions) >= MAX_OPEN_POSITIONS:
                                break
                            if symbol in self.positions:
                                continue

                            signal_side, fvg_bottom = self.check_entry_signal(symbol, btc_trend)
                            if signal_side == "LONG":
                                balance = self.client.get_usdt_balance()
                                if balance <= 0:
                                    balance = 100.0  # 兜底测试资金大小
                                    
                                initial_usdt = balance * BASE_POSITION_PCT * LEVERAGE
                                
                                order, price, qty = self.execute_market_order(symbol, "BUY", initial_usdt)
                                if order and price > 0:
                                    klines_15m = self.client.get_klines(symbol, '15m', limit=20)
                                    atr = self.calculate_atr(klines_15m) if klines_15m else price * 0.015
                                    
                                    ts = TrailingStopManager(symbol, price, side="LONG", fvg_bottom=fvg_bottom, initial_atr=atr)
                                    
                                    self.positions[symbol] = {
                                        'side': 'LONG',
                                        'phase': 1,
                                        'initial_usdt': initial_usdt,
                                        'total_usdt': initial_usdt,
                                        'total_qty': qty,
                                        'avg_entry_price': price,
                                        'ts_manager': ts
                                    }
                                    self.set_cooldown(symbol, COOLDOWN_TRADE)[span_20](start_span)[span_20](end_span)

                time.sleep(15)

            except Exception as e:
                logging.error(f"[Termux 异常捕捉] 循环错误: {e}")[span_21](start_span)[span_21](end_span)
                time.sleep(10)


# ==========================================
# 6. 代码后缀语法逻辑 (Termux 守护启动)
# ==========================================
if __name__ == '__main__':
    try:
        engine = TradingEngine()
        engine.run_loop()
    except KeyboardInterrupt:
        logging.info("[Termux] 收到键盘中断信号，程序停止运行。")
    except Exception as fatal_error:
        logging.critical(f"[Termux] 致命未捕获错误: {fatal_error}", exc_info=True)[span_22](start_span)[span_22](end_span)
    finally:
        cleanup_pid()[span_23](start_span)[span_23](end_span)
