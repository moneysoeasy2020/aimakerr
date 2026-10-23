import os
import time
from datetime import datetime
import requests
from rich.console import Console
from rich.table import Table

console = Console()

# ==================== 策略与凭据配置 ====================
API_KEY = "fapi_f0a61727a53f4c62a5a8cfb744b05c3dc6ab9d4c65f84ffda7ce6633b88592a0"
TG_BOT_TOKEN = "7931068876:AAE4IIG431xCGjTIPoJ0dGKdqmKRtNKWZEA"
TG_CHAT_ID = "7500708914"

TOP_N_TRADERS = 10
MIN_POSITION_VALUE_USD = 1000
CONSENSUS_MIN_TRADERS = 2
CONSENSUS_MIN_TOTAL_USD = 5000

SYMBOL_BLACKLIST = {
    "SOL", "ETH", "BTC", "USDC", "USDT", "WSOL", "WETH", "WBTC"
}

def send_telegram_alert(message: str):
    if not TG_BOT_TOKEN or not TG_CHAT_ID:
        return
    url = f"https://api.telegram.org/bot{TG_BOT_TOKEN}/sendMessage"
    payload = {
        "chat_id": TG_CHAT_ID,
        "text": message,
        "parse_mode": "Markdown"
    }
    try:
        res = requests.post(url, json=payload, timeout=10)
        if res.status_code != 200:
            console.print(f"[dim red]Telegram 推送失败: {res.text}[/dim red]")
    except Exception as e:
        console.print(f"[dim red]Telegram 请求异常: {e}[/dim red]")

class SmartFomoMonitor:
    def __init__(self):
        self.base_url = "https://api.fomoapi.io"
        self.headers = {
            "Accept": "application/json",
            "Authorization": f"Bearer {API_KEY}"
        }

    def _get_with_retry(self, url: str, retries=3, timeout=20):
        for attempt in range(retries):
            try:
                res = requests.get(url, headers=self.headers, timeout=timeout)
                if res.status_code == 200:
                    return res.json()
            except Exception:
                pass
            time.sleep(2)
        return None

    def get_top_traders(self) -> list:
        url = f"{self.base_url}/v2/leaderboard/24h"
        data = self._get_with_retry(url)
        if data:
            return (data if isinstance(data, list) else data.get("data", []))[:TOP_N_TRADERS]
        console.print(f"[bold red]拉取 Leaderboard 失败（已重试）[/bold red]")
        return []

    def get_trader_positions(self, handle: str) -> list:
        url = f"{self.base_url}/v2/users/{handle}/balances"
        data = self._get_with_retry(url)
        if not data:
            return []

        balances = data.get("balances", []) if isinstance(data, dict) else data
        valid_positions = []
        for item in balances:
            token_info = item.get("token", {})
            symbol = token_info.get("symbol") if isinstance(token_info, dict) else item.get("symbol")
            value_usd = item.get("valueUsd", 0)

            if not symbol:
                continue
            symbol = str(symbol).upper()

            if symbol not in SYMBOL_BLACKLIST and value_usd >= MIN_POSITION_VALUE_USD:
                valid_positions.append({
                    "handle": handle,
                    "symbol": symbol,
                    "value_usd": value_usd
                })
        return valid_positions

    def analyze_and_notify(self):
        console.print(f"[dim][{datetime.now().strftime('%H:%M:%S')}] 正在扫描 Top {TOP_N_TRADERS} 交易员持仓...[/dim]")

        traders = self.get_top_traders()
        if not traders:
            return

        all_positions = []
        for rank, trader in enumerate(traders, start=1):
            handle = trader.get("handle") or trader.get("username")
            if not handle:
                continue

            positions = self.get_trader_positions(handle)
            for pos in positions:
                pos["rank"] = rank
                all_positions.append(pos)

        if not all_positions:
            self._render_table([])
            return

        summary_dict = {}
        for pos in all_positions:
            sym = pos["symbol"]
            if sym not in summary_dict:
                summary_dict[sym] = {
                    "symbol": sym,
                    "handles": set(),
                    "ranks": set(),
                    "total_value_usd": 0.0
                }
            summary_dict[sym]["handles"].add(pos["handle"])
            summary_dict[sym]["ranks"].add(pos["rank"])
            summary_dict[sym]["total_value_usd"] += pos["value_usd"]

        smart_signals = []
        for sym, data in summary_dict.items():
            trader_count = len(data["handles"])
            total_usd = data["total_value_usd"]

            if trader_count >= CONSENSUS_MIN_TRADERS and total_usd >= CONSENSUS_MIN_TOTAL_USD:
                sorted_ranks = sorted(list(data["ranks"]))
                smart_signals.append({
                    "symbol": sym,
                    "trader_count": trader_count,
                    "total_value_usd": total_usd,
                    "ranks_str": ", ".join(map(str, sorted_ranks)),
                    "handles_str": ", ".join(data["handles"])
                })

        smart_signals.sort(key=lambda x: x["total_value_usd"], reverse=True)

        self._render_table(smart_signals)

        if smart_signals:
            msg_lines = [f"🔥 *FOMO Top 10 持仓一致性信号* [{datetime.now().strftime('%H:%M')}]"]
            for row in smart_signals:
                msg_lines.append(
                    f"\n• *{row['symbol']}*"
                    f"\n  共识人数: `{row['trader_count']}人` | 总持仓: `${row['total_value_usd']:,.0f}`"
                    f"\n  交易员排名: `Top #{row['ranks_str']}`"
                    f"\n  Handles: `{row['handles_str']}`"
                )
            send_telegram_alert("\n".join(msg_lines))

    def _render_table(self, signals):
        os.system("clear")
        now_str = datetime.now().strftime("%Y-%m-%d %H:%M:%S")

        table = Table(
            title=f"🧠 FOMO Top {TOP_N_TRADERS} 持仓共识信号 [{now_str}]",
            header_style="bold magenta"
        )
        table.add_column("币种 (Symbol)", style="cyan", justify="center")
        table.add_column("共识人数", style="bold yellow", justify="center")
        table.add_column("总持仓价值 (USD)", style="bold green", justify="right")
        table.add_column("重合排名", style="white")
        table.add_column("持仓交易员 Handles", style="dim white")

        if not signals:
            console.print(f"[bold blue][{now_str}][/bold blue] 暂无符合（人数>={CONSENSUS_MIN_TRADERS}，金额>=${CONSENSUS_MIN_TOTAL_USD:,.0f}）的 α 币种信号。")
            return

        for row in signals:
            table.add_row(
                row["symbol"],
                f"{row['trader_count']} 人",
                f"${row['total_value_usd']:,.0f}",
                f"Top #{row['ranks_str']}",
                row["handles_str"]
            )

        console.print(table)

if __name__ == "__main__":
    monitor = SmartFomoMonitor()
    monitor.analyze_and_notify()
    while True:
        time.sleep(10)
        monitor.analyze_and_notify()
