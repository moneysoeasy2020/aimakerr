#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
币安合约自动交易 - 多维度筛选做空机器人
策略：扫描全市场，综合费率、成交量、价格趋势、持仓变化评分
开仓条件：评分最高且费率 > 0.1%，每币 3.8U，5倍杠杆，最多3个同时持仓
修复：使用 futures_funding_rate 和 futures_mark_price，字段名修正为 fundingRate
"""

import time
import logging
import sys
import math
from datetime import datetime
from binance.client import Client
from binance.exceptions import BinanceAPIException
from dotenv import load_dotenv
import os

load_dotenv()

# ==================== 配置 ====================
API_KEY = os.getenv("BINANCE_API_KEY")
API_SECRET = os.getenv("BINANCE_API_SECRET")

LEVERAGE = 5
POSITION_SIZE_USDT = 3.8
MAX_POSITIONS = 3
FUNDING_RATE_THRESHOLD = 0.001   # 0.1%
MIN_24H_VOLUME = 5_000_000      # 500万 USDT
STOP_LOSS_PCT = 0.03            # 3%
TAKE_PROFIT_PCT = 0.10          # 10% 止盈
SCAN_INTERVAL = 60              # 秒
EXCLUDED_SYMBOLS = ['BTCUSDT', 'ETHUSDT', 'BNBUSDT', 'USDCUSDT']
# =============================================

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[logging.FileHandler('scanner_bot.log'), logging.StreamHandler(sys.stdout)]
)
logger = logging.getLogger(__name__)


class FundingScannerBot:
    def __init__(self, api_key, api_secret):
        self.client = Client(api_key, api_secret)
        self.open_positions = {}  # {symbol: {'entry_price': float, 'entry_time': str, 'qty': float}}

    def set_leverage(self, symbol):
        try:
            self.client.futures_change_leverage(symbol=symbol, leverage=LEVERAGE)
            logger.info(f"✅ {symbol} 杠杆 {LEVERAGE}x")
        except BinanceAPIException as e:
            if "No need to change" not in str(e):
                logger.warning(f"杠杆设置失败 {symbol}: {e}")

    def get_all_candidates(self):
        """获取所有USDT合约的完整数据，用于筛选（兼容新版库）"""
        try:
            # ---------- 获取资金费率 ----------
            funding_rates = self.client.futures_funding_rate()
            # 修正：字段为 'fundingRate' 而非 'lastFundingRate'
            rate_dict = {r['symbol']: float(r['fundingRate']) for r in funding_rates}
            
            # ---------- 获取标记价格 ----------
            mark_prices = self.client.futures_mark_price()
            mark_dict = {m['symbol']: float(m['markPrice']) for m in mark_prices}
            
            # ---------- 获取24h成交量 ----------
            tickers = self.client.futures_ticker()
            ticker_dict = {t['symbol']: t for t in tickers if t['symbol'].endswith('USDT')}

            candidates = []
            for sym, rate in rate_dict.items():
                if not sym.endswith('USDT') or sym in EXCLUDED_SYMBOLS:
                    continue
                if rate <= FUNDING_RATE_THRESHOLD:
                    continue
                mark_price = mark_dict.get(sym, 0)
                if mark_price == 0:
                    continue
                vol = float(ticker_dict.get(sym, {}).get('quoteVolume', 0))
                if vol < MIN_24H_VOLUME:
                    continue
                price_change_pct = float(ticker_dict.get(sym, {}).get('priceChangePercent', 0))
                candidates.append({
                    'symbol': sym,
                    'funding_rate': rate,
                    'mark_price': mark_price,
                    'volume_24h': vol,
                    'price_change_1h': price_change_pct,
                    'price': mark_price
                })
            # 后续逻辑：取费率前20名，获取K线计算均线
            candidates.sort(key=lambda x: x['funding_rate'], reverse=True)
            top = candidates[:20]
            for c in top:
                sym = c['symbol']
                try:
                    klines = self.client.futures_klines(symbol=sym, interval='1h', limit=20)
                    closes = [float(k[4]) for k in klines]
                    ma20 = sum(closes) / len(closes) if closes else c['mark_price']
                    c['ma20'] = ma20
                    c['price_vs_ma'] = (c['mark_price'] - ma20) / ma20
                except Exception as e:
                    logger.warning(f"获取K线失败 {sym}: {e}")
                    c['ma20'] = c['mark_price']
                    c['price_vs_ma'] = 0
            return top
        except BinanceAPIException as e:
            logger.error(f"获取数据失败: {e}")
            return []

    def calculate_score(self, candidate):
        """综合评分（归一化后加权）"""
        # 费率分数（越高越好）
        rate = candidate['funding_rate']
        rate_score = min(rate / 0.005, 1.0)  # 最高封顶0.5%

        # 成交量分数（越高越好，但超过一定值不再加分）
        vol = candidate['volume_24h']
        vol_score = min(vol / 50_000_000, 1.0)  # 5000万以上满分

        # 价格相对均线（负值越大越好，即价格低于均线越多）
        pvm = candidate.get('price_vs_ma', 0)
        if pvm < -0.1:
            pvm_score = 1.0
        elif pvm > 0.05:
            pvm_score = 0.0
        else:
            pvm_score = 1 - (pvm + 0.1) / 0.15

        # 价格变动（负值更好，即下跌）
        change = candidate['price_change_1h']
        if change < -5:
            change_score = 1.0
        elif change > 5:
            change_score = 0.0
        else:
            change_score = 1 - (change + 5) / 10

        # 加权总分
        score = 0.4 * rate_score + 0.2 * vol_score + 0.2 * pvm_score + 0.2 * change_score
        candidate['score'] = score
        return score

    def get_quantity(self, symbol, price):
        """计算下单数量（按名义本金）"""
        try:
            info = self.client.futures_exchange_info()
            for s in info['symbols']:
                if s['symbol'] == symbol:
                    for f in s['filters']:
                        if f['filterType'] == 'LOT_SIZE':
                            step_size = float(f['stepSize'])
                            min_qty = float(f['minQty'])
                            break
                    break
            notional = POSITION_SIZE_USDT * LEVERAGE
            raw_qty = notional / price
            precision = int(round(-math.log10(step_size))) if step_size < 1 else 0
            qty = math.floor(raw_qty / step_size) * step_size
            qty = round(qty, precision)
            if qty < min_qty:
                qty = min_qty
            return qty
        except Exception as e:
            logger.error(f"计算数量失败 {symbol}: {e}")
            return 0

    def open_short(self, symbol, price):
        qty = self.get_quantity(symbol, price)
        if qty <= 0:
            return False
        self.set_leverage(symbol)
        try:
            order = self.client.futures_create_order(
                symbol=symbol, side='SELL', type='MARKET', quantity=qty
            )
            logger.info(f"✅ 开仓 {symbol} 做空 {qty} 张 @ {price:.6f}")
            self.open_positions[symbol] = {
                'entry_price': price,
                'entry_time': datetime.now().isoformat(),
                'qty': qty
            }
            # 设置止损
            stop_price = price * (1 + STOP_LOSS_PCT)
            self.client.futures_create_order(
                symbol=symbol, side='BUY', type='STOP_MARKET',
                stopPrice=round(stop_price, 8), quantity=qty, timeInForce='GTE_GTC'
            )
            logger.info(f"🛑 {symbol} 止损 @ {stop_price:.6f}")
            return True
        except BinanceAPIException as e:
            logger.error(f"开仓失败 {symbol}: {e}")
            return False

    def close_position(self, symbol):
        qty = self.open_positions.get(symbol, {}).get('qty', 0)
        if qty <= 0:
            # 尝试从账户获取
            pos = self.client.futures_position_information(symbol=symbol)
            qty = abs(float(pos[0]['positionAmt'])) if pos else 0
        if qty <= 0:
            logger.warning(f"{symbol} 无持仓可平")
            return False
        try:
            self.client.futures_create_order(
                symbol=symbol, side='BUY', type='MARKET', quantity=qty, reduceOnly=True
            )
            logger.info(f"✅ 平仓 {symbol} {qty} 张")
            if symbol in self.open_positions:
                del self.open_positions[symbol]
            return True
        except Exception as e:
            logger.error(f"平仓失败 {symbol}: {e}")
            return False

    def should_close(self, symbol, current_rate, entry_price, mark_price):
        # 条件1：费率回落至低水平
        if current_rate < 0.0002:  # 0.02%
            logger.info(f"📉 {symbol} 费率回落 {current_rate*100:.3f}%，平仓")
            return True
        # 条件2：止盈
        profit_pct = (entry_price - mark_price) / entry_price * 100
        if profit_pct >= TAKE_PROFIT_PCT * 100:
            logger.info(f"💰 {symbol} 止盈 {profit_pct:.2f}%")
            return True
        return False

    def run(self):
        logger.info("🚀 多维度扫描做空机器人启动")
        logger.info(f"📊 每币本金 {POSITION_SIZE_USDT}U | {LEVERAGE}x | 最大持仓 {MAX_POSITIONS} 个")
        while True:
            try:
                # 1. 获取当前持仓
                positions = self.client.futures_account()['positions']
                pos_symbols = [p['symbol'] for p in positions if float(p['positionAmt']) != 0]
                # 更新本地记录
                for p in positions:
                    sym = p['symbol']
                    amt = float(p['positionAmt'])
                    if amt != 0 and sym not in self.open_positions:
                        self.open_positions[sym] = {
                            'entry_price': float(p['entryPrice']),
                            'entry_time': datetime.now().isoformat(),
                            'qty': abs(amt)
                        }
                logger.info(f"📋 当前持仓 ({len(pos_symbols)}/{MAX_POSITIONS}): {pos_symbols if pos_symbols else '无'}")

                # 2. 检查是否需要平仓
                all_data = self.get_all_candidates()
                rate_dict = {d['symbol']: d for d in all_data}
                for sym in list(self.open_positions.keys()):
                    if sym in rate_dict:
                        rate = rate_dict[sym]['funding_rate']
                        entry = self.open_positions[sym]['entry_price']
                        mark = rate_dict[sym]['mark_price']
                        if self.should_close(sym, rate, entry, mark):
                            self.close_position(sym)
                            time.sleep(2)

                # 3. 如果已满仓，跳过
                if len(self.open_positions) >= MAX_POSITIONS:
                    logger.info("⏳ 持仓已满，等待平仓")
                    time.sleep(SCAN_INTERVAL)
                    continue

                # 4. 获取候选标的并评分
                candidates = self.get_all_candidates()
                if not candidates:
                    logger.info("ℹ️ 无符合条件的标的")
                    time.sleep(SCAN_INTERVAL)
                    continue

                for c in candidates:
                    self.calculate_score(c)
                candidates.sort(key=lambda x: x['score'], reverse=True)

                available = [c for c in candidates if c['symbol'] not in self.open_positions]
                if not available:
                    logger.info("ℹ️ 无新标的可开仓")
                    time.sleep(SCAN_INTERVAL)
                    continue

                # 5. 开仓（最多补到MAX_POSITIONS）
                to_open = available[:MAX_POSITIONS - len(self.open_positions)]
                for c in to_open:
                    logger.info(f"🎯 开仓: {c['symbol']} 费率 {c['funding_rate']*100:.3f}% 评分 {c['score']:.3f}")
                    self.open_short(c['symbol'], c['mark_price'])
                    time.sleep(3)

                time.sleep(SCAN_INTERVAL)

            except KeyboardInterrupt:
                logger.info("🛑 用户中断")
                break
            except Exception as e:
                logger.error(f"⚠️ 错误: {e}")
                time.sleep(SCAN_INTERVAL)


if __name__ == "__main__":
    if not API_KEY or not API_SECRET:
        logger.error("❌ 请在 .env 中设置 BINANCE_API_KEY 和 BINANCE_API_SECRET")
        sys.exit(1)
    bot = FundingScannerBot(API_KEY, API_SECRET)
    bot.run()
