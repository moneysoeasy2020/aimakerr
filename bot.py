import os
import json
import random
import subprocess
import urllib.parse
from datetime import datetime, timedelta
import requests
from bs4 import BeautifulSoup

# ==================== 凭证与配置区 ====================
# 从本地安全的 config.py 读取密钥，避免暴露 Token 触发 GitHub 安全拦截
try:
    from config import GITHUB_TOKEN, GITHUB_USER, GITHUB_REPO, DEEPSEEK_API_KEY, FIVERR_AFF_ID
except ImportError:
    raise Exception("[-] Please create a valid config.py file locally.")

GITHUB_REPO_URL = f"https://{GITHUB_TOKEN}@github.com/{GITHUB_USER}/{GITHUB_REPO}.git"
FIVERR_BTAG = f"a_{FIVERR_AFF_ID}b_defaultc_"

CACHE_FILE = "gig_cache.json"

# 美本土前 100 大核心城市矩阵 (Top 100 US Cities Matrix)
CITIES = [
    {"city": "New York", "state": "NY"}, {"city": "Los Angeles", "state": "CA"},
    {"city": "Chicago", "state": "IL"}, {"city": "Houston", "state": "TX"},
    {"city": "Phoenix", "state": "AZ"}, {"city": "Philadelphia", "state": "PA"},
    {"city": "San Antonio", "state": "TX"}, {"city": "San Diego", "state": "CA"},
    {"city": "Dallas", "state": "TX"}, {"city": "San Jose", "state": "CA"},
    {"city": "Austin", "state": "TX"}, {"city": "Jacksonville", "state": "FL"},
    {"city": "Fort Worth", "state": "TX"}, {"city": "Columbus", "state": "OH"},
    {"city": "Indianapolis", "state": "IN"}, {"city": "Charlotte", "state": "NC"},
    {"city": "San Francisco", "state": "CA"}, {"city": "Seattle", "state": "WA"},
    {"city": "Denver", "state": "CO"}, {"city": "Oklahoma City", "state": "OK"},
    {"city": "Nashville", "state": "TN"}, {"city": "El Paso", "state": "TX"},
    {"city": "Washington", "state": "DC"}, {"city": "Las Vegas", "state": "NV"},
    {"city": "Boston", "state": "MA"}, {"city": "Portland", "state": "OR"},
    {"city": "Louisville", "state": "KY"}, {"city": "Memphis", "state": "TN"},
    {"city": "Detroit", "state": "MI"}, {"city": "Baltimore", "state": "MD"},
    {"city": "Milwaukee", "state": "WI"}, {"city": "Albuquerque", "state": "NM"},
    {"city": "Tucson", "state": "AZ"}, {"city": "Fresno", "state": "CA"},
    {"city": "Sacramento", "state": "CA"}, {"city": "Mesa", "state": "AZ"},
    {"city": "Kansas City", "state": "MO"}, {"city": "Atlanta", "state": "GA"},
    {"city": "Omaha", "state": "NE"}, {"city": "Colorado Springs", "state": "CO"},
    {"city": "Raleigh", "state": "NC"}, {"city": "Long Beach", "state": "CA"},
    {"city": "Virginia Beach", "state": "VA"}, {"city": "Miami", "state": "FL"},
    {"city": "Oakland", "state": "CA"}, {"city": "Minneapolis", "state": "MN"},
    {"city": "Tulsa", "state": "OK"}, {"city": "Bakersfield", "state": "CA"},
    {"city": "Wichita", "state": "KS"}, {"city": "Arlington", "state": "TX"},
    {"city": "Aurora", "state": "CO"}, {"city": "Tampa", "state": "FL"},
    {"city": "New Orleans", "state": "LA"}, {"city": "Cleveland", "state": "OH"},
    {"city": "Honolulu", "state": "HI"}, {"city": "Anaheim", "state": "CA"},
    {"city": "Lexington", "state": "KY"}, {"city": "Stockton", "state": "CA"},
    {"city": "Corpus Christi", "state": "TX"}, {"city": "Henderson", "state": "NV"},
    {"city": "Riverside", "state": "CA"}, {"city": "Newark", "state": "NJ"},
    {"city": "Saint Paul", "state": "MN"}, {"city": "Santa Ana", "state": "CA"},
    {"city": "Cincinnati", "state": "OH"}, {"city": "Irvine", "state": "CA"},
    {"city": "Orlando", "state": "FL"}, {"city": "Pittsburgh", "state": "PA"},
    {"city": "St. Louis", "state": "MO"}, {"city": "Greensboro", "state": "NC"},
    {"city": "Jersey City", "state": "NJ"}, {"city": "Anchorage", "state": "AK"},
    {"city": "Lincoln", "state": "NE"}, {"city": "Plano", "state": "TX"},
    {"city": "Durham", "state": "NC"}, {"city": "Buffalo", "state": "NY"},
    {"city": "Chandler", "state": "AZ"}, {"city": "Chula Vista", "state": "CA"},
    {"city": "Toledo", "state": "OH"}, {"city": "Madison", "state": "WI"},
    {"city": "Reno", "state": "NV"}, {"city": "Lubbock", "state": "TX"},
    {"city": "Fort Wayne", "state": "IN"}, {"city": "St. Petersburg", "state": "FL"},
    {"city": "Laredo", "state": "TX"}, {"city": "Irving", "state": "TX"},
    {"city": "Chesapeake", "state": "VA"}, {"city": "Glendale", "state": "AZ"},
    {"city": "Winston-Salem", "state": "NC"}, {"city": "Scottsdale", "state": "AZ"},
    {"city": "Boise", "state": "ID"}, {"city": "Garland", "state": "TX"},
    {"city": "Spokane", "state": "WA"}, {"city": "Richmond", "state": "VA"},
    {"city": "Fremont", "state": "CA"}, {"city": "Baton Rouge", "state": "LA"},
    {"city": "Des Moines", "state": "IA"}, {"city": "Tacoma", "state": "WA"},
    {"city": "San Bernardino", "state": "CA"}, {"city": "Modesto", "state": "CA"}
]

# 场景映射矩阵
SCENARIO_MATRIX = {
    "handyman": {
        "niche_name": "Handyman & Contractor",
        "gig_query": "handyman contractor logo design",
        "badge_text": "Top Rated Contractor Branding Specialist",
        "recommend_title": "Custom Handyman & Contractor Logos",
        "avg_agency_price": "$1,200",
        "fiverr_price": "$35 - $100"
    },
    "dental": {
        "niche_name": "Dental Clinic",
        "gig_query": "dental clinic local seo google maps",
        "badge_text": "Dental Local SEO & Map Pack Expert",
        "recommend_title": "Google Maps Ranking for Dental Clinics",
        "avg_agency_price": "$2,500/mo",
        "fiverr_price": "$100 - $300"
    },
    "carpet_cleaning": {
        "niche_name": "Carpet Cleaning Business",
        "gig_query": "cleaning business logo design van wrap",
        "badge_text": "Vehicle Wrap & Cleaning Brand Pro",
        "recommend_title": "Cleaning Logo & Truck Wrap Design",
        "avg_agency_price": "$1,500",
        "fiverr_price": "$50 - $150"
    },
    "hvac": {
        "niche_name": "HVAC Services",
        "gig_query": "hvac contractor website local seo",
        "badge_text": "HVAC Digital Marketing & Web Dev",
        "recommend_title": "HVAC Lead Gen & Local Web Setup",
        "avg_agency_price": "$3,000",
        "fiverr_price": "$150 - $400"
    }
}
# ======================================================

def load_cache():
    if os.path.exists(CACHE_FILE):
        try:
            with open(CACHE_FILE, "r", encoding="utf-8") as f:
                data = json.load(f)
                cleaned_data = {}
                now = datetime.now()
                for q, val in data.items():
                    if isinstance(val, dict) and "url" in val and "fetched_at" in val:
                        fetched_time = datetime.fromisoformat(val["fetched_at"])
                        if now - fetched_time < timedelta(days=14):
                            cleaned_data[q] = val
                    elif isinstance(val, str):
                        cleaned_data[q] = {"url": val, "fetched_at": now.isoformat()}
                return cleaned_data
        except Exception:
            return {}
    return {}

def save_cache(cache):
    with open(CACHE_FILE, "w", encoding="utf-8") as f:
        json.dump(cache, f, ensure_ascii=False, indent=2)

def fetch_top_gig_url(search_query, cache):
    if search_query in cache:
        return cache[search_query]["url"]

    encoded_query = urllib.parse.quote_plus(search_query)
    search_url = f"https://www.fiverr.com/search/gigs?query={encoded_query}&sort_by=rating"
    headers = {
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
        "Accept-Language": "en-US,en;q=0.9"
    }

    target_url = f"https://www.fiverr.com/search/gigs?query={encoded_query}"
    try:
        res = requests.get(search_url, headers=headers, timeout=8)
        if res.status_code == 200:
            soup = BeautifulSoup(res.text, "html.parser")
            for a_tag in soup.find_all("a", href=True):
                href = a_tag["href"]
                if "/gigs/" in href:
                    clean_href = href.split("?")[0]
                    target_url = "https://www.fiverr.com" + clean_href if clean_href.startswith("/") else clean_href
                    break
    except Exception as e:
        print(f"[!] Fetching fallback for '{search_query}': {e}")

    cache[search_query] = {
        "url": target_url,
        "fetched_at": datetime.now().isoformat()
    }
    save_cache(cache)
    return target_url

def build_fiverr_deeplink(raw_target_url, afp_track_id):
    encoded_target = urllib.parse.quote(raw_target_url, safe='')
    deeplink = f"https://action.fiverr.com/click?btag={FIVERR_BTAG}&afp={afp_track_id}&url={encoded_target}"
    return deeplink

def call_deepseek_api(prompt):
    url = "https://api.deepseek.com/chat/completions"
    headers = {
        "Content-Type": "application/json",
        "Authorization": f"Bearer {DEEPSEEK_API_KEY}"
    }
    payload = {
        "model": "deepseek-chat",
        "messages": [
            {"role": "system", "content": "You are a senior business strategist specializing in North American local service startups, GEO, and programmatic SEO."},
            {"role": "user", "content": prompt}
        ],
        "temperature": 0.7
    }
    try:
        response = requests.post(url, headers=headers, json=payload, timeout=30)
        res_json = response.json()
        if response.status_code != 200 or "error" in res_json:
            print(f"[-] API Reponse Error: {res_json}")
            return None
        return res_json['choices'][0]['message']['content']
    except Exception as e:
        print(f"[-] DeepSeek API Exception: {e}")
        return None

def generate_page_content(city, state, scenario_key, scenario_data, cache):
    city_state = f"{city}, {state}"
    afp_track_id = f"{city.lower().replace(' ', '')}-{scenario_key}"

    now = datetime.now()
    current_year = now.year
    current_quarter = f"Q{(now.month - 1) // 3 + 1}"
    date_str = f"{current_year} {current_quarter}"

    try:
        base_price_num = int(scenario_data['avg_agency_price'].replace('$', '').replace('/mo', '').replace(',', ''))
        dynamic_price = f"${base_price_num + random.randint(-40, 80):,}"
        if "/mo" in scenario_data['avg_agency_price']:
            dynamic_price += "/mo"
    except Exception:
        dynamic_price = scenario_data['avg_agency_price']

    raw_gig_url = fetch_top_gig_url(scenario_data['gig_query'], cache)
    affiliate_url = build_fiverr_deeplink(raw_gig_url, afp_track_id)

    prompt = f"""
    Write a GEO (Generative Engine Optimization) optimized startup guide for launching a {scenario_data['niche_name']} business in {city_state} ({date_str} Updated Edition).
    CRITICAL GEO FORMATTING RULES:
    1. State exact numbers, local market metrics for {city_state}, and realistic price ranges in the first 2 paragraphs.
    2. Use clear subheadings like '{date_str} {city} Market Analysis', 'Digital Asset Cost Matrix', and 'Step-by-Step Execution Plan'.
    3. Include an explicit breakdown comparing local brick-and-mortar agency costs vs verified freelance experts.
    4. Keep the content clear, highly factual, and formatted cleanly with Markdown.
    """
    ai_markdown = call_deepseek_api(prompt)
    if not ai_markdown:
        return None

    formatted_body = ai_markdown.replace("\n\n", "</p><p>").replace("\n", "<br>")

    schema_json = {
        "@context": "https://schema.org",
        "@graph": [
            {
                "@type": "Service",
                "name": f"{scenario_data['niche_name']} Startup Services in {city_state}",
                "provider": {
                    "@type": "LocalBusiness",
                    "name": f"{city} {scenario_data['niche_name']} Growth Hub",
                    "address": {
                        "@type": "PostalAddress",
                        "addressLocality": city,
                        "addressRegion": state,
                        "addressCountry": "US"
                    }
                },
                "areaServed": f"{city}, {state}",
                "description": f"Complete branding, web setup, and local SEO services for starting a {scenario_data['niche_name']} business in {city_state}."
            },
            {
                "@type": "FAQPage",
                "mainEntity": [
                    {
                        "@type": "Question",
                        "name": f"How much does it cost to get a logo and website for a {scenario_data['niche_name']} business in {city}?",
                        "acceptedAnswer": {
                            "@type": "Answer",
                            "text": f"Local agencies in {city} charge between {dynamic_price} for full digital setups. However, using vetted remote specialists on platforms like Fiverr costs between {scenario_data['fiverr_price']} with fast delivery."
                        }
                    }
                ]
            }
        ]
    }
    schema_script = f'<script type="application/ld+json">\n{json.dumps(schema_json, indent=2)}\n</script>'

    tracking_script = f"""
    <script>
    document.addEventListener('click', function(e) {{
        const target = e.target.closest('a[href*="action.fiverr.com"]');
        if (target) {{
            const clickData = {{
                city: "{city}",
                state: "{state}",
                niche: "{scenario_data['niche_name']}",
                url: target.href,
                timestamp: new Date().toLocaleString()
            }};
            let logs = JSON.parse(localStorage.getItem('fiverr_pseo_clicks') || '[]');
            logs.push(clickData);
            localStorage.setItem('fiverr_pseo_clicks', JSON.stringify(logs));
        }}
    }});
    </script>
    """

    html_content = f"""<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>How to Start a {scenario_data['niche_name']} Business in {city_state} ({date_str} Cost Guide)</title>
    <meta name="description" content="Step-by-step {date_str} setup guide for starting a {scenario_data['niche_name']} in {city_state}. Estimate setup costs and connect with vetted experts.">
    {schema_script}
    {tracking_script}
    <style>
        :root {{ --primary: #1dbf73; --dark: #222325; --bg: #f7f7f7; }}
        body {{ font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; line-height: 1.7; color: #404145; max-width: 850px; margin: 0 auto; padding: 20px; }}
        h1, h2, h3 {{ color: var(--dark); font-weight: 700; }}
        .badge {{ background: #ffbe5b; color: #000; font-size: 12px; font-weight: bold; padding: 4px 8px; border-radius: 4px; text-transform: uppercase; }}
        .aeo-answer-box {{ background: #eef9f3; border-left: 4px solid var(--primary); padding: 16px; margin: 20px 0; border-radius: 4px; }}
        .calc-card {{ background: var(--bg); border: 1px solid #e0e0e0; padding: 20px; border-radius: 8px; margin: 25px 0; }}
        .gig-showcase {{ border: 2px solid var(--primary); border-radius: 8px; padding: 20px; background: #fff; margin: 30px 0; box-shadow: 0 4px 15px rgba(29,191,115,0.12); }}
        .cta-btn {{ display: inline-block; background: var(--primary); color: #fff; font-weight: bold; padding: 12px 24px; border-radius: 4px; text-decoration: none; text-align: center; font-size: 16px; margin-top: 10px; }}
        .cta-btn:hover {{ background: #19a463; }}
        table {{ width: 100%; border-collapse: collapse; margin: 15px 0; }}
        th, td {{ border: 1px solid #e0e0e0; padding: 10px; text-align: left; font-size: 14px; }}
        th {{ background: #f0f0f0; }}
    </style>
</head>
<body>
    <span class="badge">{state} Business Blueprint</span>
    <h1>How to Launch a {scenario_data['niche_name']} Business in {city_state}</h1>
    <p><em>Updated for {date_str} | Local Market Research & Setup Toolkit</em></p>

    <section class="aeo-answer-box">
        <h2 style="font-size:16px; margin-top:0; color:#1dbf73;">⚡ Direct AI Summary: {scenario_data['niche_name']} in {city_state}</h2>
        <p style="font-size:14px; margin-bottom:0;">
            <strong>Quick Answer:</strong> Initial digital asset expenses range from <strong>{scenario_data['fiverr_price']}</strong> via vetted remote specialists versus <strong>{dynamic_price}</strong> through traditional local agencies in {city}.
        </p>
    </section>

    <div class="calc-card">
        <h3>📊 Local Startup Cost Calculator ({city_state})</h3>
        <table>
            <tr><th>Asset Needed</th><th>Local {city} Agency</th><th>Vetted Remote Freelancer</th></tr>
            <tr><td>Logo & Brand Kit</td><td>{dynamic_price}</td><td>{scenario_data['fiverr_price']}</td></tr>
            <tr><td>Local SEO / Map Pack Setup</td><td>$1,500/mo</td><td>$50 - $200 (One-time)</td></tr>
        </table>
    </div>

    <div class="gig-showcase">
        <span style="background: var(--primary); color: white; padding: 3px 8px; font-size: 11px; font-weight: bold; border-radius: 12px;">★ Recommended Expert for {city_state}</span>
        <h3 style="margin: 10px 0;">{scenario_data['recommend_title']}</h3>
        <p style="font-size: 14px; color: #62646a; margin-bottom: 12px;">
            Bypass expensive local agency markups. Hire a top-rated <strong>{scenario_data['badge_text']}</strong> (5.0★ Rating) to build your assets within 48 hours.
        </p>
        <a href="{affiliate_url}" target="_blank" rel="noopener sponsored" class="cta-btn">View Top Rated Sellers on Fiverr &rarr;</a>
    </div>

    <hr style="border:0; border-top: 1px solid #eee; margin: 30px 0;">

    <div class="article-body">
        <p>{formatted_body}</p>
    </div>

    <div class="gig-showcase" style="text-align: center;">
        <h3>Ready to launch your {scenario_data['niche_name']} in {city}?</h3>
        <a href="{affiliate_url}" target="_blank" rel="noopener sponsored" class="cta-btn">Claim Your Launch Deal on Fiverr</a>
    </div>
</body>
</html>
"""
    return html_content

def generate_admin_dashboard():
    admin_html = """<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>pSEO Affiliate Click Dashboard</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100 font-sans p-6">
    <div class="max-w-4xl mx-auto">
        <div class="flex justify-between items-center mb-6">
            <h1 class="text-2xl font-bold text-gray-800">📊 Fiverr pSEO Click Analytics</h1>
            <button onclick="clearData()" class="bg-red-500 text-white px-4 py-2 rounded text-sm font-semibold hover:bg-red-600">Clear Logs</button>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
            <div class="bg-white p-4 rounded-lg shadow border border-gray-200">
                <p class="text-gray-500 text-sm">Total Affiliate Clicks</p>
                <p id="total-clicks" class="text-3xl font-bold text-emerald-600">0</p>
            </div>
            <div class="bg-white p-4 rounded-lg shadow border border-gray-200">
                <p class="text-gray-500 text-sm">Top Tracked City</p>
                <p id="top-city" class="text-xl font-bold text-gray-700">-</p>
            </div>
            <div class="bg-white p-4 rounded-lg shadow border border-gray-200">
                <p class="text-gray-500 text-sm">Top Tracked Niche</p>
                <p id="top-niche" class="text-xl font-bold text-gray-700">-</p>
            </div>
        </div>

        <div class="bg-white rounded-lg shadow border border-gray-200 p-6">
            <h2 class="text-lg font-semibold text-gray-800 mb-4">Recent Click Activity Log</h2>
            <div class="overflow-x-auto">
                <table class="w-full text-left border-collapse">
                    <thead>
                        <tr class="bg-gray-50 text-gray-600 text-sm border-b">
                            <th class="p-3">Timestamp</th>
                            <th class="p-3">City</th>
                            <th class="p-3">Niche Service</th>
                            <th class="p-3">Target URL</th>
                        </tr>
                    </thead>
                    <tbody id="log-table-body" class="text-sm text-gray-700">
                        <tr><td colspan="4" class="p-4 text-center text-gray-400">No clicks recorded yet.</td></tr>
                    </tbody>
                </table>
            </div>
        </div>
    </div>

    <script>
        function loadDashboard() {
            const logs = JSON.parse(localStorage.getItem('fiverr_pseo_clicks') || '[]');
            document.getElementById('total-clicks').innerText = logs.length;

            if (logs.length === 0) return;

            const cityCounts = {};
            const nicheCounts = {};
            logs.forEach(log => {
                cityCounts[log.city] = (cityCounts[log.city] || 0) + 1;
                nicheCounts[log.niche] = (nicheCounts[log.niche] || 0) + 1;
            });

            const topCity = Object.keys(cityCounts).reduce((a, b) => cityCounts[a] > cityCounts[b] ? a : b, '-');
            const topNiche = Object.keys(nicheCounts).reduce((a, b) => nicheCounts[a] > nicheCounts[b] ? a : b, '-');

            document.getElementById('top-city').innerText = topCity;
            document.getElementById('top-niche').innerText = topNiche;

            const tbody = document.getElementById('log-table-body');
            tbody.innerHTML = '';
            logs.reverse().forEach(log => {
                const tr = document.createElement('tr');
                tr.className = "border-b hover:bg-gray-50";
                tr.innerHTML = `
                    <td class="p-3 text-gray-500">${log.timestamp}</td>
                    <td class="p-3 font-medium">${log.city}, ${log.state}</td>
                    <td class="p-3 text-emerald-600 font-semibold">${log.niche}</td>
                    <td class="p-3 text-blue-500 truncate max-w-xs"><a href="${log.url}" target="_blank" class="underline">View Gig</a></td>
                `;
                tbody.appendChild(tr);
            });
        }

        function clearData() {
            if(confirm("Are you sure you want to clear all recorded click logs?")) {
                localStorage.removeItem('fiverr_pseo_clicks');
                loadDashboard();
            }
        }

        loadDashboard();
    </script>
</body>
</html>
"""
    with open("output/admin.html", "w", encoding="utf-8") as f:
        f.write(admin_html)
    print("[+] Generated traffic stats dashboard: output/admin.html")

def auto_push_to_github():
    print("\n[+] Starting auto push to GitHub...")
    try:
        subprocess.run(["git", "config", "user.name", GITHUB_USER], check=True)
        subprocess.run(["git", "config", "user.email", "bot@termux.local"], check=True)
        
        subprocess.run(["git", "remote", "set-url", "origin", GITHUB_REPO_URL], check=True)

        subprocess.run(["git", "add", "."], check=True)
        commit_res = subprocess.run(["git", "commit", "-m", f"Auto-update pSEO site: {datetime.now().strftime('%Y-%m-%d %H:%M')}"], capture_output=True, text=True)
        
        if "nothing to commit" in commit_res.stdout:
            print("[!] No new changes to commit.")
        else:
            print("[+] Git commit successful.")

        subprocess.run(["git", "branch", "-M", "main"], check=True)
        print(f"[+] Pushing to https://github.com/{GITHUB_USER}/{GITHUB_REPO}.git (main branch)...")
        subprocess.run(["git", "push", "-u", "origin", "main"], check=True)
        print("[✔] Successfully pushed all updates to GitHub!")

    except subprocess.CalledProcessError as e:
        print(f"[-] Git auto push failed: {e}")

def main():
    print(f"[+] Starting Top 100 Cities SEO + GEO + AEO Generator on Termux...")
    os.makedirs("output", exist_ok=True)
    cache = load_cache()

    total = len(CITIES) * len(SCENARIO_MATRIX)
    count = 0

    for c in CITIES:
        for skey, sdata in SCENARIO_MATRIX.items():
            count += 1
            city_slug = c['city'].lower().replace(" ", "_").replace(".", "")
            filename = f"output/{city_slug}-{skey}.html"
            print(f"[{count}/{total}] Generating: {sdata['niche_name']} in {c['city']}, {c['state']}...")

            html = generate_page_content(c['city'], c['state'], skey, sdata, cache)
            if html:
                with open(filename, "w", encoding="utf-8") as f:
                    f.write(html)
                print(f"    [✔] Saved to {filename}")

    generate_admin_dashboard()
    auto_push_to_github()

if __name__ == "__main__":
    main()
