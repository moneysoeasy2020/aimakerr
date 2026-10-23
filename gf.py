import os
import sys

# ==================== 1. Termux 依赖自动检测与安装 ====================
def auto_install_dependencies():
    required_packages = ["requests", "beautifulsoup4", "google-generativeai"]
    for package in required_packages:
        try:
            pkg_name = package.replace("-", "_")
            if pkg_name == "beautifulsoup4":
                pkg_name = "bs4"
            __import__(pkg_name)
        except ImportError:
            print(f"[*] [Termux 环境] 检测到缺少依赖 [{package}]，正在自动安装...")
            os.system(f"{sys.executable} -m pip install {package} --quiet")

auto_install_dependencies()

import email
import imaplib
import json
import random
import re
import smtplib
import time
from datetime import datetime
from email.header import decode_header, Header
from email.mime.text import MIMEText
from urllib.parse import urlparse
import requests
from bs4 import BeautifulSoup
import google.generativeai as genai

# ==================== 2. 核心 API 与账户配置 ====================
GEMINI_API_KEY = "AIzaSyBfaiHtDQ_Rmdb4hn46s8QdVKP3hDUc_Nw"
genai.configure(api_key=GEMINI_API_KEY)

# SerpApi Key
SERPAPI_KEY = "bbef6d8d5efb481459b880b94283d31cb354fddd5528df1e970ceb36daf5ba84"

# Gmail 发信配置 (控速 40 封/天)
GMAIL_CONFIG = {
    "email": "wvbiz2013@gmail.com",
    "app_password": "zktz vtav arwc sbzv",
    "daily_limit": 40
}

# Fiverr Affiliate 配置
FIVERR_CONFIG = {
    "affiliate_id": "510726",
    "ai_transformation_deeplink": "https://www.fiverr.com/search/gigs?query=ai+automation+shopify+geo&af=510726"
}

# 高意图搜索语法池（HR 招聘拦截 + D2C 独立站）
HIGH_INTENT_SEARCH_POOL = [
    'site:linkedin.com/jobs "Shopify" "AI Automation"',
    'site:linkedin.com/jobs "E-commerce" "AI Specialist"',
    'site:greenhouse.io "Shopify" "AI"',
    'site:lever.co "E-commerce" "Automation"',
    'site:myshopify.com "add to cart" "free shipping" pet',
    'site:myshopify.com "buy now" "exclusive offer" skincare',
    'site:myshopify.com "subscribe & save" wellness',
    'site:myshopify.com "limited time" "shop now" wig'
]

DB_FILE = "leads_db.json"

# ==================== 3. 数据库管理与配额控制 ====================
def load_db():
    today_str = datetime.now().strftime("%Y-%m-%d")
    if os.path.exists(DB_FILE):
        try:
            with open(DB_FILE, "r", encoding="utf-8") as f:
                data = json.load(f)
                if "leads" not in data:
                    new_data = {
                        "meta": {"last_sent_date": today_str, "today_sent_count": 0, "keyword_index": 0},
                        "leads": data
                    }
                    save_db(new_data)
                    return new_data
                if "keyword_index" not in data.get("meta", {}):
                    data["meta"]["keyword_index"] = 0
                return data
        except Exception:
            pass
            
    return {"meta": {"last_sent_date": today_str, "today_sent_count": 0, "keyword_index": 0}, "leads": {}}

def save_db(db):
    with open(DB_FILE, "w", encoding="utf-8") as f:
        json.dump(db, f, ensure_ascii=False, indent=4)

def check_and_reset_daily_quota(db):
    today_str = datetime.now().strftime("%Y-%m-%d")
    if "meta" not in db:
        db["meta"] = {"last_sent_date": today_str, "today_sent_count": 0, "keyword_index": 0}
    
    if db["meta"].get("last_sent_date") != today_str:
        db["meta"]["last_sent_date"] = today_str
        db["meta"]["today_sent_count"] = 0
        save_db(db)
    return db["meta"]["today_sent_count"]

# ==================== 4. Gemini 安全 API 调用 ====================
def safe_gemini_call(prompt, max_retries=3, initial_delay=6):
    model = genai.GenerativeModel("gemini-2.5-flash")
    delay = initial_delay
    for attempt in range(max_retries):
        try:
            time.sleep(4)
            response = model.generate_content(prompt, generation_config={"response_mime_type": "application/json"})
            return response.text
        except Exception as e:
            if "429" in str(e) or "ResourceExhausted" in str(e):
                time.sleep(delay)
                delay *= 2
            else:
                break
    return None

# ==================== 5. 关键词与域名耦合引擎 (Domain Coupling Engine) ====================
class DomainCouplingEngine:
    """域名与关键词耦合匹配逻辑"""
    
    # 常用决策者/部门前缀池（根据域名性质耦合组合）
    PREFIXES_HR = ["hr", "careers", "talent", "recruiting", "jobs"]
    PREFIXES_D2C = ["marketing", "growth", "support", "hello", "info", "founder"]

    @staticmethod
    def extract_domain_keywords(domain):
        """从域名中提取核心语义关键词 (例如: petbeauty-store.com -> ['petbeauty', 'store'])"""
        clean_domain = re.sub(r'^(https?://)?(www\.)?', '', domain.lower())
        base_domain = clean_domain.split('/')[0].split('.')[0]
        keywords = re.split(r'[-_]', base_domain)
        return [kw for kw in keywords if len(kw) > 2]

    @classmethod
    def synthesize_coupled_emails(cls, domain, lead_type):
        """生成符合域名结构的预选耦合邮箱集合"""
        clean_domain = urlparse(domain if domain.startswith("http") else f"https://{domain}").netloc
        if not clean_domain:
            clean_domain = domain.split('/')[0]
            
        prefixes = cls.PREFIXES_HR if lead_type == "HR_HIRING" else cls.PREFIXES_D2C
        coupled = [f"{p}@{clean_domain}" for p in prefixes]
        return coupled

    @classmethod
    def score_and_select_best_email(cls, found_emails, domain, lead_type):
        """对页面抓取的邮箱与域名耦合度进行打分筛选"""
        domain_keywords = cls.extract_domain_keywords(domain)
        scored_emails = []

        for email in found_emails:
            score = 0
            email_lower = email.lower()
            prefix, email_domain = email_lower.split('@') if '@' in email_lower else ("", "")

            # 1. 域名完全匹配加 50 分
            if domain in email_domain:
                score += 50
            
            # 2. 避免公共免费邮箱（gmail, yahoo 等）
            if any(free in email_domain for free in ['gmail.com', 'yahoo.com', 'hotmail.com', 'outlook.com']):
                score -= 30

            # 3. 部门前缀耦合加分
            target_prefixes = cls.PREFIXES_HR if lead_type == "HR_HIRING" else cls.PREFIXES_D2C
            if any(p in prefix for p in target_prefixes):
                score += 30

            # 4. 关键词耦合加分
            if any(kw in prefix for kw in domain_keywords):
                score += 20

            scored_emails.append((email_lower, score))

        scored_emails.sort(key=lambda x: x[1], reverse=True)
        return scored_emails[0][0] if scored_emails and scored_emails[0][1] > 0 else None

# ==================== 6. SerpApi 抓取与深度解析模块 ====================
HEADERS = {
    "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36"
}

def extract_emails_from_text(text):
    pattern = r'[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}'
    found = re.findall(pattern, text)
    valid_emails = []
    for e in found:
        e_lower = e.lower()
        if not any(e_lower.endswith(ext) for ext in ['.png', '.jpg', '.jpeg', '.gif', '.svg', '.webp', '.css', '.js']):
            valid_emails.append(e_lower)
    return list(set(valid_emails))

def scrape_target_site(target_url):
    if not target_url.startswith("http"):
        target_url = "https://" + target_url

    parsed_domain = urlparse(target_url).netloc
    collected_text = ""
    found_emails = []

    try:
        resp = requests.get(target_url, headers=HEADERS, timeout=6)
        if resp.status_code == 200:
            soup = BeautifulSoup(resp.text, 'html.parser')
            collected_text += soup.get_text(separator=' ') + " "
            found_emails.extend(extract_emails_from_text(resp.text))

        found_emails = list(set(found_emails))

        # 调用 AI 智能辨识场景类型与名称
        prompt = f"""
        Analyze the scraped web content:
        Target URL: {target_url}
        Candidate Raw Emails: {found_emails}
        Raw Text Snippet: {collected_text[:1200]}
        
        Task:
        1. Classify lead_type: 'HR_HIRING' (if hiring/jobs) or 'D2C_STORE' (if e-commerce/brand).
        2. Identify Brand/Company Name.
        3. Identify Primary Industry/Niche.
        
        Output strictly in JSON:
        {{
            "lead_type": "HR_HIRING or D2C_STORE",
            "brand_name": "Company or Brand Name",
            "niche": "Industry or Job Role"
        }}
        """
        resp_json_str = safe_gemini_call(prompt)
        if not resp_json_str:
            return False

        analysis_data = json.loads(resp_json_str)
        lead_type = analysis_data.get("lead_type", "D2C_STORE")

        # 🔗 执行【域名耦合引擎】选出最优目标邮箱
        best_email = DomainCouplingEngine.score_and_select_best_email(found_emails, parsed_domain, lead_type)

        # 若网页未直接搜到，自动采用“域名耦合逻辑”生成默认支撑邮箱
        if not best_email:
            synthesized_emails = DomainCouplingEngine.synthesize_coupled_emails(parsed_domain, lead_type)
            best_email = synthesized_emails[0]

        if best_email and "@" in best_email:
            db = load_db()
            if best_email in db["leads"]:
                return False

            db["leads"][best_email] = {
                "store_name": analysis_data.get("brand_name", parsed_domain),
                "lead_type": lead_type,
                "niche": analysis_data.get("niche", "E-commerce"),
                "url": parsed_domain,
                "status": "NEW",
                "history": [],
                "created_at": time.strftime("%Y-%m-%d %H:%M:%S")
            }
            save_db(db)
            print(f"✅ [域名耦合线索入库] 类型: [{lead_type}] | 域名: {parsed_domain} | 耦合邮箱: {best_email}")
            return True
        return False
    except Exception:
        return False

def auto_pipeline_serpapi_scraping():
    db = load_db()
    meta = db.setdefault("meta", {})
    idx = meta.get("keyword_index", 0)
    
    current_query = HIGH_INTENT_SEARCH_POOL[idx % len(HIGH_INTENT_SEARCH_POOL)]
    print(f"\n==================================================")
    print(f"🎯 [SerpApi 抓取] 当前轮换语法: 【{current_query}】")
    print(f"==================================================")
    
    meta["keyword_index"] = (idx + 1) % len(HIGH_INTENT_SEARCH_POOL)
    save_db(db)
    
    url = "https://serpapi.com/search.json"
    params = {
        "engine": "google",
        "q": current_query,
        "api_key": SERPAPI_KEY,
        "gl": "us",
        "hl": "en"
    }
    
    try:
        resp = requests.get(url, params=params, timeout=15)
        if resp.status_code == 200:
            data = resp.json()
            target_urls = []
            
            for ad in data.get("ads", []):
                if ad.get("link"):
                    target_urls.append(ad.get("link"))
                    
            for res in data.get("organic_results", []):
                if res.get("link") and res.get("link") not in target_urls:
                    target_urls.append(res.get("link"))
            
            success_count = 0
            for link in target_urls[:10]:
                if scrape_target_site(link):
                    success_count += 1
                    if success_count >= 3:
                        break
            print(f"✨ 本轮完成，成功为 CRM 新增 {success_count} 条域名耦合线索。")
    except Exception as e:
        print(f"❌ [SerpApi 抓取异常] {e}")

# ==================== 7. AI 破冰信生成与发送 ====================
def generate_ai_email(store_name, store_niche, website_url, lead_type):
    if lead_type == "HR_HIRING":
        prompt = f"""
        You are Alex, an AI talent consultant. Write a cold email to hiring manager/HR at {store_name}.
        Topic: Hiring AI/Automation talent for {store_niche}.
        Angle: Hiring full-time takes 60+ days. Pitch vetted freelance AI specialists (via Fiverr) to deploy prototypes immediately while interviewing.
        Requirements:
        - Subject: Under 6 words.
        - Body: Under 75 words. Casual, sign off as Alex.
        Output strictly in JSON: {{"subject": "...", "body": "..."}}
        """
    else:
        prompt = f"""
        You are Alex, a GEO & AI Growth Consultant. Write a cold email to founder of {store_name} ({website_url}).
        Angle: AI Generative Engine Optimization (GEO). Help them get recommended by ChatGPT / Perplexity to shoppers.
        Requirements:
        - Subject: Under 6 words.
        - Body: Under 70 words. Mention AI search optimization for {store_name}. Casual, sign off as Alex.
        Output strictly in JSON: {{"subject": "...", "body": "..."}}
        """

    res_str = safe_gemini_call(prompt)
    if res_str:
        return json.loads(res_str)
    
    return {
        "subject": f"Quick question regarding {store_name}",
        "body": f"Hi team at {store_name},\n\nNoticed your presence in {store_niche}. We help brands optimize for AI search recommendations (ChatGPT, Perplexity) and automate workflows.\n\nOpen to a quick overview?\n\nBest,\nAlex"
    }

def generate_ai_followup(client_reply_text, lead_type):
    aff_link = FIVERR_CONFIG['ai_transformation_deeplink']
    prompt = f"""
    You are Alex. Client replied: "{client_reply_text}". Lead type: {lead_type}
    Write a short follow-up under 80 words guiding them to check vetted AI experts here: {aff_link}.
    Sign off as Alex.
    Output strictly in JSON: {{"subject": "...", "body": "..."}}
    """
    res_str = safe_gemini_call(prompt)
    return json.loads(res_str) if res_str else {
        "subject": "Re: AI Automation & GEO Setup",
        "body": f"Hi,\n\nThanks for replying! You can check top-rated AI GEO & automation developers directly here: {aff_link}\n\nBest,\nAlex"
    }

def send_gmail_email(recipient_email, subject, body_text):
    sender_email = GMAIL_CONFIG["email"]
    app_password = GMAIL_CONFIG["app_password"]

    msg = MIMEText(body_text, "plain", "utf-8")
    msg["From"] = Header(f"Alex <{sender_email}>", "utf-8")
    msg["To"] = Header(recipient_email, "utf-8")
    msg["Subject"] = Header(subject, "utf-8")

    try:
        server = smtplib.SMTP("smtp.gmail.com", 587)
        server.starttls()
        server.login(sender_email, app_password)
        server.sendmail(sender_email, [recipient_email], msg.as_string())
        server.quit()
        print(f"✅ [Gmail 发送成功] -> {recipient_email}")
        return True
    except Exception as e:
        print(f"❌ [Gmail 发送失败] {e}")
        return False

def reply_email_via_gmail(recipient_email, original_subject, body_text, original_message_id=None):
    sender_email = GMAIL_CONFIG["email"]
    app_password = GMAIL_CONFIG["app_password"]

    msg = MIMEText(body_text, "plain", "utf-8")
    msg["From"] = Header(f"Alex <{sender_email}>", "utf-8")
    msg["To"] = Header(recipient_email, "utf-8")
    
    subject = original_subject if original_subject.startswith("Re:") else f"Re: {original_subject}"
    msg["Subject"] = Header(subject, "utf-8")
    
    if original_message_id:
        msg["In-Reply-To"] = original_message_id
        msg["References"] = original_message_id

    try:
        server = smtplib.SMTP("smtp.gmail.com", 587)
        server.starttls()
        server.login(sender_email, app_password)
        server.sendmail(sender_email, [recipient_email], msg.as_string())
        server.quit()
        print(f"🚀 [自动转化成功] 已回复带 Affiliate 链接的邮件至 [{recipient_email}]")
        return True
    except Exception as e:
        print(f"❌ [跟进回复失败] {e}")
        return False

# ==================== 8. 流程控制与监控看板 ====================
def auto_process_new_outreach():
    db = load_db()
    today_sent = check_and_reset_daily_quota(db)
    remaining_quota = GMAIL_CONFIG["daily_limit"] - today_sent
    print(f"\n📊 [配额看板] 今日已发: {today_sent}/{GMAIL_CONFIG['daily_limit']} 封 | 剩余可用: {remaining_quota} 封")

    if remaining_quota <= 0:
        print("⚠️ 今日发信额度已达上限，暂停本轮发送。")
        return

    new_leads = [email for email, info in db["leads"].items() if info["status"] == "NEW"]
    if not new_leads:
        print("📭 [CRM 调度] 当前无待发新线索。")
        return

    batch_to_send = new_leads[:remaining_quota]
    for client_email in batch_to_send:
        info = db["leads"][client_email]
        lead_type = info.get("lead_type", "D2C_STORE")
        print(f"\n🤖 正在为 [{info['store_name']}] ({lead_type}) 生成破冰邮件...")
        email_data = generate_ai_email(info['store_name'], info['niche'], info['url'], lead_type)
        
        if send_gmail_email(client_email, email_data["subject"], email_data["body"]):
            info["status"] = "OUTREACH_SENT"
            info["history"].append({
                "time": time.strftime("%Y-%m-%d %H:%M:%S"),
                "action": "OUTREACH_SENT",
                "subject": email_data["subject"]
            })
            db["meta"]["today_sent_count"] += 1
            save_db(db)
            
            delay = random.randint(20, 30)
            print(f"💤 平滑延时 {delay} 秒...")
            time.sleep(delay)

def auto_check_and_followup():
    db = load_db()
    tracked_emails = [email for email, info in db["leads"].items() if info["status"] in ["OUTREACH_SENT", "REPLIED"]]
    if not tracked_emails:
        print("📭 [CRM 监控] 无需跟进的活跃客户。")
        return

    sender_email = GMAIL_CONFIG["email"]
    app_password = GMAIL_CONFIG["app_password"]

    try:
        imaplib._MAXLINE = 10000000
        mail = imaplib.IMAP4_SSL("imap.gmail.com")
        mail.login(sender_email, app_password)
        mail.select("inbox")

        for target_email in tracked_emails:
            search_query = f'(UNSEEN FROM "{target_email}")'
            status, messages = mail.search(None, search_query)
            if status != "OK":
                continue

            email_ids = messages[0].split()
            if not email_ids:
                continue

            print(f"🔔 收到客户 [{target_email}] 的新回复！")
            for e_id in email_ids:
                res, msg_data = mail.fetch(e_id, "(RFC822)")
                for response_part in msg_data:
                    if isinstance(response_part, tuple):
                        msg = email.message_from_bytes(response_part[1])
                        subject, encoding = decode_header(msg["Subject"])[0]
                        if isinstance(subject, bytes):
                            subject = subject.decode(encoding or "utf-8", errors="ignore")
                        msg_id = msg.get("Message-ID")
                        
                        body = ""
                        if msg.is_multipart():
                            for part in msg.walk():
                                if part.get_content_type() == "text/plain":
                                    body = part.get_payload(decode=True).decode(errors="ignore")
                                    break
                        else:
                            body = msg.get_payload(decode=True).decode(errors="ignore")

                        lead_type = db["leads"][target_email].get("lead_type", "D2C_STORE")
                        followup_data = generate_ai_followup(body, lead_type)
                        
                        if reply_email_via_gmail(target_email, subject, followup_data['body'], msg_id):
                            db["leads"][target_email]["status"] = "REPLIED"
                            db["leads"][target_email]["history"].append({
                                "time": time.strftime("%Y-%m-%d %H:%M:%S"),
                                "action": "AI_FOLLOWUP_SENT",
                                "client_msg": body.strip(),
                                "ai_reply": followup_data['body']
                            })
                            save_db(db)
        mail.logout()
    except Exception as e:
        print(f"❌ [IMAP 监控异常] {e}")

def print_hourly_statistics():
    db = load_db()
    today_count = check_and_reset_daily_quota(db)
    leads = db.get("leads", {})
    
    total_leads = len(leads)
    status_counts = {"NEW": 0, "OUTREACH_SENT": 0, "REPLIED": 0}
    
    for info in leads.values():
        st = info.get("status", "NEW")
        status_counts[st] = status_counts.get(st, 0) + 1

    print("\n==================================================")
    print(f"📈 【Termux SDR 运营看板 (域名耦合增强版)】 - {time.strftime('%Y-%m-%d %H:%M:%S')}")
    print("==================================================")
    print(f"🔥 今日发信: {today_count} / {GMAIL_CONFIG['daily_limit']} 封 | CRM 数据库总线索: {total_leads} 条")
    print(f"   * 📌 待发新线索 (NEW): {status_counts.get('NEW', 0)}")
    print(f"   * 📤 已发破冰信 (OUTREACH_SENT): {status_counts.get('OUTREACH_SENT', 0)}")
    print(f"   * 💬 已互动跟进 (REPLIED): {status_counts.get('REPLIED', 0)}")
    print(f"💰 Fiverr 追踪 ID: [510726]")
    print("==================================================\n")

# ==================== 9. 主守护进程入口 ====================
if __name__ == "__main__":
    print("==================================================")
    print("🚀 Termux AI SDR Agent (域名/关键词耦合 + GEO + HR 双驱动)")
    print("==================================================\n")
    
    LOOP_INTERVAL_MINUTES = 60
    
    try:
        while True:
            print(f"⏰ [{time.strftime('%Y-%m-%d %H:%M:%S')}] 唤醒执行新一轮任务...")
            
            # 1. 执行带有【域名耦合机制】的抓取与过滤
            try:
                auto_pipeline_serpapi_scraping()
            except Exception as e:
                print(f"❌ [抓取流程异常] {e}")

            # 2. 生成定制信并发送
            try:
                auto_process_new_outreach()
            except Exception as e:
                print(f"❌ [发信流程异常] {e}")
                
            # 3. 监控客户未读回复并自动回复包含 Fiverr Affiliate 链接
            try:
                auto_check_and_followup()
            except Exception as e:
                print(f"❌ [回复流程异常] {e}")
                
            # 4. 看板输出
            try:
                print_hourly_statistics()
            except Exception as e:
                print(f"❌ [看板输出异常] {e}")
                
            print(f"💤 执行完毕，进入休眠 {LOOP_INTERVAL_MINUTES} 分钟...\n")
            time.sleep(LOOP_INTERVAL_MINUTES * 60)
            
    except KeyboardInterrupt:
        print("\n👋 收到终止指令，Termux 进程已退出。")
