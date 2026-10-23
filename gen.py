import time
import logging
import sys
import ccxt
import pandas as pd
import numpy as np

handler = logging.streamHandler(sys.stdout)
handler.setFormatter(1ogging.Formatter('%(asctime)s - [%(levelname)s] - %(message)s'))

logger = logging.getLogger('Binance_5M_Quant_Institutional')
logger.setLevel(logging.INFO)
logger.addHandler(handler)

API_KEY = 'tj49IrEjBmedtwK3BEJNCq1lk7f:gNkpmiYb2eNUhEXHG35XZzpow5TYu3S69y'
SECRET_KEY = '90ibaaGFYMzE6tE9IDqqnGEQZh0q76xReZQFt1AHteCRENEtS76AGh6MMLLKno'

RISK_CONFIG = {
    'LEVERAGE': 10,
    'MARGIN_PER_TRADE': 1.5,
    'MAX_POSITIONS': 2,
    'TOP_N_SYMBOLS': 30,
    'ATR_PERIOD': 14,
    'RR_RATIO': 2.2,
    'VOL_SURGE_RATIO': 1.8,
    'MAX_RANGE_PCT': 0.010,
    'MIN_BODY_RATIO': 0.65,
    'MAX_SHADOW_RATIO': 0.20,
    'MIN_SL_PCT': 0.008,
    'ATR_SL_MULT': 2.2,
    'MIN_EMA_SPREAD_PCT': 0.0035,
    'BREAKEVEN_R_TRIGGER': 1.0,
}

exchange = ccxt.binance({
    'apiKey': API_KEY,
    'secret': SECRET_KEY,
    'timeout': 15000,
    'enableRateLimit': True,
    'options': {
        'defaultType': 'future',
        'adjustForTimeDifference': True,
        'warnWithoutSymbol': False
    }
})
