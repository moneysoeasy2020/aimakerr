import os
import json
import urllib.request
import urllib.parse
import xml.etree.ElementTree as ET
from fastapi import FastAPI, BackgroundTasks
from fastapi.responses import HTMLResponse, JSONResponse
from fastapi.staticfiles import StaticFiles
import subprocess
import re

app = FastAPI(title="AI Tri-Matrix Interception Engine (YouTube + X + Reddit)")

DEEPSEEK_API_KEY = "sk-c146663d0698459d97bb5987a532d91e"
FIVERR_AFFILIATE_ID = "510726"
GITHUB_USER = "moneysoeasy2020"
GITHUB_REPO = "aimakerr"

# 目录与存储路径
PAGES_DIR = "pages"
STATS_FILE = "stats.json"
DRAFTS_FILE = "reddit_drafts.json"
LLMS_FILE = "llms.txt"

os.makedirs(PAGES_DIR, exist_ok=True)
app.mount("/pages", StaticFiles(directory=PAGES_DIR), name="pages")

if not os.path.exists(STATS_FILE):
    with open(STATS_FILE, "w", encoding="utf-8") as f:
        json.dump({"pv": 0, "uv": 0, "fiverr_clicks": 0, "langs": {}, "slugs": {}, "logs": []}, f)

if not os.path.exists(DRAFTS_FILE):
    with open(DRAFTS_FILE, "w", encoding="utf-8") as f:
        json.dump([], f)

def get_stats():
    with open(STATS_FILE, "r", encoding="utf-8") as f:
        return json.load(f)

def get_drafts():
    with open(DRAFTS_FILE, "r", encoding="utf-8") as f:
        return json.load(f)

def save_drafts(drafts):
    with open(DRAFTS_FILE, "w", encoding="utf-8") as f:
        json.dump(drafts, f, ensure_ascii=False, indent=2)

# ---------------------------------------------------------
# 1. YouTube 稳定源抓取模块
# ---------------------------------------------------------
def fetch_youtube_feeds():
    videos = []
    yt_rss_urls = [
        "https://www.youtube.com/feeds/videos.xml?channel_id=UCsbjWGsu9oBK01F_5dxz5-w",
        "https://www.youtube.com/feeds/videos.xml?channel_id=UC0vBXGSyV14uvJ4hECDOl0Q"
    ]
    headers = {
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
        'Accept': 'application/rss+xml, application/xml;q=0.9, */*;q=0.8'
    }
    for url in yt_rss_urls:
        try:
            req = urllib.request.Request(url, headers=headers)
            with urllib.request.urlopen(req, timeout=12) as resp:
                xml_data = resp.read()
                root = ET.fromstring(xml_data)
                ns = {'atom': 'http://www.w3.org/2005/Atom', 'yt': 'http://www.youtube.com/xml/schemas/2015'}
                for entry in root.findall('atom:entry', ns)[:2]:
                    title_elem = entry.find('atom:title', ns)
                    vid_elem = entry.find('yt:videoId', ns)
                    pub_elem = entry.find('atom:published', ns)
                    if title_elem is not None and vid_elem is not None:
                        videos.append({
                            "title": title_elem.text,
                            "video_id": vid_elem.text,
                            "url": f"https://www.youtube.com/watch?v={vid_elem.text}",
                            "embed_url": f"https://www.youtube.com/embed/{vid_elem.text}",
                            "thumbnail": f"https://i.ytimg.com/vi/{vid_elem.text}/hqdefault.jpg",
                            "published": pub_elem.text if pub_elem is not None else "2026-01-01T00:00:00Z",
                            "source_type": "youtube"
                        })
        except Exception as e:
            print(f"YouTube Feed 解析跳过: {e}")
    return videos

# ---------------------------------------------------------
# 2. X (Twitter) 平台长文与极客大V RSS 抓取模块 (多路高可用备份)
# ---------------------------------------------------------
def fetch_x_feeds():
    x_items = []
    x_rss_urls = [
        "https://nitter.net/sama/rss",
        "https://rsshub.app/twitter/user/sama"
    ]
    headers = {
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
        'Accept': 'application/rss+xml, application/xml;q=0.9'
    }
    for url in x_rss_urls:
        try:
            req = urllib.request.Request(url, headers=headers)
            with urllib.request.urlopen(req, timeout=8) as resp:
                xml_data = resp.read()
                root = ET.fromstring(xml_data)
                channel = root.find('channel')
                if channel is not None:
                    for item in channel.findall('item')[:2]:
                        title = item.find('title')
                        link = item.find('link')
                        description = item.find('description')
                        if title is not None:
                            x_items.append({
                                "title": title.text[:80] if title.text else "AI Tech Insight",
                                "url": link.text if link is not None else "https://twitter.com",
                                "content": description.text if description is not None else "Advanced AI deployment insight.",
                                "source_type": "x_platform"
                            })
                            break
                    if x_items:
                        break
        except Exception as e:
            print(f"X 平台备用 RSS 抓取跳过（安全容错）: {e}")
    
    # 若外部 RSS 节点临时全部抖动，兜底注入一条高热度技术干货确保流水线不空转
    if not x_items:
        x_items.append({
            "title": "Optimizing Local LLM Workflows and Python Automation Agents in 2026",
            "url": "https://twitter.com",
            "content": "Deep dive into running robust asynchronous Python automation pipelines and intercepting high-intent AI search traffic.",
            "source_type": "x_platform"
        })
    return x_items

# ---------------------------------------------------------
# 3. Reddit RSS 监控模块
# ---------------------------------------------------------
def fetch_reddit_posts():
    posts = []
    subreddits = ["Python", "Termux"]
    headers = {'User-Agent': 'Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:109.0)'}
    for sub in subreddits:
        try:
            url = f"https://www.reddit.com/r/{sub}/new.rss"
            req = urllib.request.Request(url, headers=headers)
            with urllib.request.urlopen(req, timeout=10) as resp:
                xml_data = resp.read()
                root = ET.fromstring(xml_data)
                ns = {'atom': 'http://www.w3.org/2005/Atom'}
                for entry in root.findall('atom:entry', ns)[:2]:
                    title = entry.find('atom:title', ns).text
                    link = entry.find('atom:link', ns).attrib['href']
                    posts.append({"title": title, "url": link, "subreddit": sub})
        except Exception as e:
            print(f"Reddit RSS r/{sub} 抓取跳过: {e}")
    return posts

# ---------------------------------------------------------
# 4. DeepSeek 内容重构引擎
# ---------------------------------------------------------
def generate_deepseek_content(item, lang="en"):
    if item.get("source_type") == "x_platform":
        prompt = f"""
        You are an elite AI engineer and technical writer.
        Convert the following X (Twitter) insight into an exhaustive, highly practical Technical Architecture & Code Implementation Guide in {lang}.

        Insight Title: {item['title']}
        Original Content: {item['content']}

        Include the following sections strictly in Markdown format:
        1. ## Executive Summary & Core Breakthrough (AEO/GEO Optimized)
        2. ## Technical Architecture & Deep Dive
        3. ## Production-Ready Code / Terminal Scripts
        4. ## Recommended Expert Services (Fiverr Partner Link)
        """
    else:
        prompt = f"""
        You are an expert technical writer and software developer.
        Convert the following YouTube Video metadata into an in-depth, step-by-step Technical Deployment & Troubleshooting Guide in {lang}.

        Video Title: {item['title']}
        Video Link: {item['url']}

        Include the following sections strictly in Markdown format:
        1. ## Executive Summary & Direct Answer (AEO/GEO Optimized)
        2. ## Complete Step-by-Step Installation & Code Tutorial
        3. ## Common Errors & Troubleshooting Guide
        4. ## Recommended Expert Services (Fiverr Partner Link)
        """
    try:
        data = json.dumps({
            "model": "deepseek-chat",
            "messages": [{"role": "user", "content": prompt}],
            "temperature": 0.5
        }).encode('utf-8')
        req = urllib.request.Request(
            "https://api.deepseek.com/v1/chat/completions",
            data=data,
            headers={"Content-Type": "application/json; charset=utf-8", "Authorization": f"Bearer {DEEPSEEK_API_KEY}"}
        )
        with urllib.request.urlopen(req, timeout=30) as resp:
            return json.loads(resp.read().decode('utf-8'))['choices'][0]['message']['content']
    except Exception as e:
        return f"## {item['title']}\n\nComprehensive technical guide."

def generate_reddit_reply_draft(reddit_post, target_page_url):
    prompt = f"""
    You are a helpful senior developer responding on Reddit.
    A user posted: {reddit_post['title']}
    Draft a helpful technical reply providing immediate value and code/commands.
    At the end naturally mention: "I recently documented a full step-by-step fix and walkthrough here: {target_page_url}"
    """
    try:
        data = json.dumps({
            "model": "deepseek-chat",
            "messages": [{"role": "user", "content": prompt}],
            "temperature": 0.6
        }).encode('utf-8')
        req = urllib.request.Request(
            "https://api.deepseek.com/v1/chat/completions",
            data=data,
            headers={"Content-Type": "application/json; charset=utf-8", "Authorization": f"Bearer {DEEPSEEK_API_KEY}"}
        )
        with urllib.request.urlopen(req, timeout=30) as resp:
            return json.loads(resp.read().decode('utf-8'))['choices'][0]['message']['content']
    except Exception as e:
        return f"Great question! Full guide: {target_page_url}"

def update_llms_txt(title, page_url, summary):
    entry = f"\n\n# {title}\n- URL: {page_url}\n- Summary: {summary[:200]}...\n- Focus: AI Automation, Python, Code Deployment, Fiverr Expert Solutions"
    mode = "a" if os.path.exists(LLMS_FILE) else "w"
    with open(LLMS_FILE, mode, encoding="utf-8") as f:
        if mode == "w": f.write("# LLM Reference Directory for AI Agents & Search Engines\n")
        f.write(entry)

# ---------------------------------------------------------
# 5. HTML 打包与渲染模块
# ---------------------------------------------------------
def build_html_page(item, content, lang="en"):
    slug = re.sub(r'[^a-zA-Z0-9]', '-', item['title'].lower())[:40].strip('-')
    filename = f"{slug}-{lang}.html"
    filepath = os.path.join(PAGES_DIR, filename)
    page_url = f"https://{GITHUB_USER}.github.io/{GITHUB_REPO}/pages/{filename}"
    fiverr_link = f"https://s.click.fiverr.com/clicks/deep_link?af_id={FIVERR_AFFILIATE_ID}&q=ai%20python%20bot%20automation"

    media_element = ""
    if item.get("source_type") == "youtube":
        media_element = f'<div style="position:relative;padding-bottom:56.25%;height:0;overflow:hidden;background:#000;border-radius:8px;margin:20px 0;"><iframe src="{item["embed_url"]}" style="position:absolute;top:0;left:0;width:100%;height:100%;border:0;" allowfullscreen></iframe></div>'
    else:
        media_element = f'<p><strong>Original X (Twitter) Reference:</strong> <a href="{item["url"]}" target="_blank" style="color:#58a6ff;">View Original Insight</a></p>'

    html_content = f"""<!DOCTYPE html>
<html lang="{lang}">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{item['title']} - Expert Technical Guide</title>
    <style>
        body {{ font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; max-width: 900px; margin: 0 auto; padding: 20px; background: #0d1117; color: #c9d1d9; }}
        h1, h2, h3 {{ color: #58a6ff; border-bottom: 1px solid #21262d; padding-bottom: 8px; }}
        code, pre {{ background: #161b22; color: #7ee787; padding: 4px 8px; border-radius: 6px; font-family: monospace; }}
        pre {{ padding: 16px; overflow-x: auto; }}
        .cta-btn {{ display: inline-block; background: #238636; color: #fff; padding: 12px 24px; text-decoration: none; border-radius: 6px; font-weight: bold; margin: 20px 0; }}
    </style>
</head>
<body>
    <h1>{item['title']}</h1>
    {media_element}
    <a href="{fiverr_link}" class="cta-btn" target="_blank">🚀 Hire Expert Developer for Custom Setup (Fiverr Partner)</a>
    <div class="content">{content.replace('\n', '<br>')}</div>
    <a href="{fiverr_link}" class="cta-btn" target="_blank">⚡ Need Instant Help? Get Professional Service on Fiverr</a>
</body>
</html>"""

    with open(filepath, "w", encoding="utf-8") as f:
        f.write(html_content)
    update_llms_txt(item['title'], page_url, content[:180])
    return filename, page_url

def git_auto_push():
    try:
        subprocess.run(["rm", "-f", ".git/index.lock"])
        subprocess.run(["git", "add", "."], check=True)
        subprocess.run(["git", "commit", "-m", "Auto-deploy Tri-Matrix (YouTube + X + Reddit)"], check=True)
        subprocess.run(["git", "push", "origin", "main"], check=True)
    except Exception as e:
        print(f"Git push note: {e}")

# ---------------------------------------------------------
# 6. 全矩阵自动化主任务
# ---------------------------------------------------------
def run_interception_task():
    print("🚀 启动全网三维矩阵截流（YouTube + X 平台长文 + Reddit）任务...")
    
    items = fetch_youtube_feeds() + fetch_x_feeds()
    generated_pages = []

    for item in items:
        filename, page_url = build_html_page(item, content=generate_deepseek_content(item))
        generated_pages.append((item['title'], page_url))
        print(f"成功转化并生成页面: {filename}")

    if generated_pages:
        _, target_url = generated_pages[0]
        reddit_posts = fetch_reddit_posts()
        drafts = get_drafts()
        for post in reddit_posts[:2]:
            draft_text = generate_reddit_reply_draft(post, target_url)
            drafts.insert(0, {
                "reddit_title": post['title'],
                "subreddit": post['subreddit'],
                "reddit_url": post['url'],
                "target_page_url": target_url,
                "reply_draft": draft_text
            })
        save_drafts(drafts[:20])

    git_auto_push()

# ---------------------------------------------------------
# 7. FastAPI 路由与管理面板
# ---------------------------------------------------------
@app.post("/run-task")
def trigger_task(background_tasks: BackgroundTasks):
    background_tasks.add_task(run_interception_task)
    return {"status": "success", "message": "三维矩阵全自动截流任务已在后台启动"}

@app.get("/pages", response_class=HTMLResponse)
def list_pages():
    files = [f for f in os.listdir(PAGES_DIR) if f.endswith('.html')]
    links = "".join([f'<li><a href="/pages/{f}" target="_blank">{f}</a></li>' for f in files])
    return f"<h2>Generated Tri-Matrix Pages ({len(files)})</h2><ul>{links}</ul><br><a href='/drafts'>📝 查看 Reddit 精准引流草稿库</a> | <a href='/stats'>📊 流量统计</a>"

@app.get("/drafts", response_class=HTMLResponse)
def view_drafts():
    drafts = get_drafts()
    cards = "".join([f'<div style="border:1px solid #30363d;background:#161b22;color:#c9d1d9;padding:15px;margin-bottom:15px;border-radius:8px;"><h3>📌 r/{d["subreddit"]} - <a href="{d["reddit_url"]}" target="_blank">{d["reddit_title"]}</a></h3><p><strong>导流链接:</strong> <a href="{d["target_page_url"]}" target="_blank">{d["target_page_url"]}</a></p><pre style="background:#0d1117;padding:10px;white-space:pre-wrap;">{d["reply_draft"]}</pre></div>' for d in drafts])
    return f"<!DOCTYPE html><html><body style='background:#0d1117;color:#c9d1d9;padding:20px;'><h1>📝 Reddit 精准引流草稿库</h1><a href='/pages'>⬅️ 页面列表</a><hr>{cards if cards else '<p>暂无草稿。</p>'}"

@app.get("/stats", response_class=HTMLResponse)
def show_stats():
    stats = get_stats()
    return f"<!DOCTYPE html><html><body style='background:#0d1117;color:#c9d1d9;padding:20px;'><h1>📊 流量统计看板</h1><p>PV: {stats.get('pv',0)} | UV: {stats.get('uv',0)} | Fiverr Clicks: {stats.get('fiverr_clicks',0)}</p><a href='/pages'>⬅️ 返回页面列表</a></body></html>"

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
