import os
import sys
import time
import logging
import traceback
import json
import urllib.request
import urllib.parse
import email
import imaplib
import random
import re
import smtplib
from datetime import datetime
from email.header import decode_header, Header
from email.mime.text import MIMEText
from urllib.parse import urlparse

# ==================== 1. 核心 API 与账户配置 ====================
GEMINI_API_KEY = "AIzaSyBfaiHtDQ_Rmdb4hn46s8QdVKP3hDUc_Nw"
SERPAPI_KEY = "bbef6d8d5efb481459b880b94283d31cb354fddd5528df1e970ceb36daf5ba84"

GMAIL_CONFIG = {
    "email": "wvbiz2013@gmail.com",
    "app_password": "zktz vtav arwc sbzv",
    "daily_limit": 40
}

FIVERR_CONFIG = {
    "affiliate_id": "510726",
    "ai_transformation_deeplink": "https://www.fiverr.com/search/gigs?query=ai+automation+shopify+geo&af=510726"
}

HIGH_INTENT_SEARCH_POOL = [
    'site:myshopify.com "add to cart" "free shipping" "women fashion"',
    'site:myshopify.com "buy now" "new arrival" "women clothing"',
    'site:linkedin.com/jobs "Fashion" "Shopify" "AI Automation"',
    'site:myshopify.com "subscribe & save" "skincare" "beauty"',
    'site:myshopify.com "buy now" "weight loss" "supplements"',
    'site:myshopify.com "exclusive offer" "wellness" "health"',
    'site:myshopify.com "add to cart" "kitchenware" "home decor"',
    'site:myshopify.com "free shipping" "small appliances" "kitchen"',
    'site:greenhouse.io "Home & Kitchen" "Ecommerce"',
    'site:myshopify.com "shop now" "AI toys" "smart toys"',
    'site:myshopify.com "add to cart" "baby supplies" "infant"',
    'site:myshopify.com "buy now" "kids clothing" "children apparel"',
    'site:myshopify.com "subscribe & save" "pet supplies" "dog cat"',
    'site:myshopify.com "add to cart" "pet accessories"',
    'site:myshopify.com "buy now" "yoga wear" "activewear"',
    'site:myshopify.com "free shipping" "fitness equipment" "workout"',
    'site:lever.co "Fitness" "Automation"'
]

DB_FILE = "leads_db.json"

# ==================== 2. 数据库管理与配额控制 ====================
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

# ==================== 3. 原生 HTTP Gemini 安全调用封装 ====================
def safe_gemini_call(prompt, max_retries=3, initial_delay=6):
    url = f"https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key={GEMINI_API_KEY}"
    headers = {'Content-Type': 'application/json'}
    payload = {
        "contents": [{"parts": [{"text": prompt}]}],
        "generationConfig": {"response_mime_type": "application/json"}
    }
    data = json.dumps(payload).encode('utf-8')
    
    delay = initial_delay
    for attempt in range(max_retries):
        try:
            time.sleep(4)
            req = urllib.request.Request(url, data=data, headers=headers, method='POST')
            with urllib.request.urlopen(req, timeout=15) as response:
                res_body = json.loads(response.read().decode('utf-8'))
                text_content = res_body['candidates'][0]['content']['parts'][0]['text']
                return text_content
        except Exception as e:
            if "429" in str(e) or "ResourceExhausted" in str(e):
                time.sleep(delay)
                delay *= 2
            else:
                break
    return None

# ==================== 4. 彻底修复：严格域名清洗与真实邮箱校验引擎 ====================
class DomainCouplingEngine:
    PREFIXES_HR = ["hr", "careers", "talent", "recruiting", "jobs"]
    PREFIXES_D2C = ["marketing", "growth", "support", "hello", "info", "founder", "sales"]

    @staticmethod
    def clean_root_domain(raw_domain):
        """核心修复：彻底剔除 www. 以及协议头，获取纯正的根域名"""
        domain = raw_domain.lower().strip()
        domain = re.sub(r'^(https?://)?(www\.)?', '', domain)
        return domain.split('/')[0]

    @staticmethod
    def is_english_domain(domain):
        non_english_tlds = ['.cn', '.ru', '.kr', '.jp', '.tw', '.hk', '.ir', '.ua', '.by', '.th', '.vn']
        domain_lower = domain.lower()
        return not any(domain_lower.endswith(tld) for tld in non_english_tlds)

    @classmethod
    def score_and_select_best_email(cls, found_emails, raw_domain, lead_type):
        """只对【网页真实抓取到的邮箱】进行清洗和打分，严格禁止虚构邮箱"""
        clean_domain = cls.clean_root_domain(raw_domain)
        scored_emails = []

        for email in found_emails:
            email_lower = email.lower().strip()
            if '@' not in email_lower:
                continue
            
            prefix, email_domain = email_lower.split('@', 1)
            # 彻底清洗邮箱中的 www.
            email_domain = re.sub(r'^www\.', '', email_domain)
            clean_email = f"{prefix}@{email_domain}"

            score = 0
            # 必须包含当前域名后缀
            if clean_domain in email_domain:
                score += 50
            
            # 剔除免费公众邮箱
            if any(free in email_domain for free in ['gmail.com', 'yahoo.com', 'hotmail.com', 'outlook.com']):
                score -= 40

            target_prefixes = cls.PREFIXES_HR if lead_type == "HR_HIRING" else cls.PREFIXES_D2C
            if any(p in prefix for p in target_prefixes):
                score += 30

            scored_emails.append((clean_email, score))

        scored_emails.sort(key=lambda x: x[1], reverse=True)
        # 只有在确有高匹配度真实邮箱时才返回
        return scored_emails[0][0] if scored_emails and scored_emails[0][1] > 0 else None

# ==================== 5. 网页抓取与语言校验模块 ====================
HEADERS = {
    "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36",
    "Accept-Language": "en-US,en;q=0.9"
}

def is_text_primarily_english(text):
    non_english_chars = re.findall(r'[\u4e00-\u9fa5\u3040-\u30ff\u3130-\u318f\u0400-\u04FF]', text)
    if len(non_english_chars) > 10:
        return False
    return True

def extract_emails_from_text(text):
    pattern = r'[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}'
    found = re.findall(pattern, text)
    valid_emails = []
    for e in found:
        e_lower = e.lower().strip()
        # 清洗邮箱中的 www.
        if '@www.' in e_lower:
            e_lower = e_lower.replace('@www.', '@')
            
        if not any(e_lower.endswith(ext) for ext in ['.png', '.jpg', '.jpeg', '.gif', '.svg', '.webp', '.css', '.js']):
            valid_emails.append(e_lower)
    return list(set(valid_emails))

def scrape_target_site(target_url):
    clean_domain = DomainCouplingEngine.clean_root_domain(target_url)
    
    if not DomainCouplingEngine.is_english_domain(clean_domain):
        print(f"🚫 [域名过滤] 非英文主流后缀: {clean_domain}，跳过。")
        return False

    found_emails = []
    html_content = ""

    fetch_url = target_url if target_url.startswith("http") else "https://" + target_url

    try:
        req = urllib.request.Request(fetch_url, headers=HEADERS)
        with urllib.request.urlopen(req, timeout=6) as resp:
            html_content = resp.read().decode('utf-8', errors='ignore')
            found_emails.extend(extract_emails_from_text(html_content))

        text_snippet = re.sub(r'<[^>]+>', ' ', html_content)[:1500]
        if not is_text_primarily_english(text_snippet):
            print(f"🚫 [语言过滤] 网页 [{clean_domain}] 含有大量非英文字符，跳过。")
            return False

        found_emails = list(set(found_emails))
        
        # ⚠️ 修复关键：如果网页上压根没有留任何电子邮箱，直接放弃，绝不上报假邮箱！
        if not found_emails:
            print(f"⚠️ [无邮箱放弃] 网页 [{clean_domain}] 未抓取到真实邮箱，放弃入库。")
            return False

        prompt = f"""
        Analyze web content:
        Target URL: {target_url}
        Candidate Raw Emails: {found_emails}
        Raw Text Snippet: {text_snippet}
        
        Task:
        1. Confirm language is English. If NOT English, set is_english to false.
        2. Classify lead_type: 'HR_HIRING' or 'D2C_STORE'.
        3. Identify Brand/Company Name (in English).
        4. Identify Primary Industry/Niche.
        
        Output strictly in JSON:
        {{
            "is_english": true_or_false,
            "lead_type": "HR_HIRING or D2C_STORE",
            "brand_name": "Company Name",
            "niche": "Category Name"
        }}
        """
        resp_json_str = safe_gemini_call(prompt)
        if not resp_json_str:
            return False

        analysis_data = json.loads(resp_json_str)
        if not analysis_data.get("is_english", True):
            return False

        lead_type = analysis_data.get("lead_type", "D2C_STORE")

        # 筛选出最真实的域名匹配邮箱
        best_email = DomainCouplingEngine.score_and_select_best_email(found_emails, clean_domain, lead_type)

        if best_email and "@" in best_email and not "@www." in best_email:
            db = load_db()
            if best_email in db["leads"]:
                return False

            db["leads"][best_email] = {
                "store_name": analysis_data.get("brand_name", clean_domain),
                "lead_type": lead_type,
                "niche": analysis_data.get("niche", "E-commerce"),
                "url": clean_domain,
                "status": "NEW",
                "history": [],
                "created_at": time.strftime("%Y-%m-%d %H:%M:%S")
            }
            save_db(db)
            print(f"✅ [高品质真实线索入库] 域名: {clean_domain} | 真实有效邮箱: {best_email}")
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
    print(f"🎯 [SerpApi 精准抓取] 轮换 #{idx % len(HIGH_INTENT_SEARCH_POOL) + 1}: 【{current_query}】")
    print(f"==================================================")
    
    meta["keyword_index"] = (idx + 1) % len(HIGH_INTENT_SEARCH_POOL)
    save_db(db)
    
    params = urllib.parse.urlencode({
        "engine": "google",
        "q": current_query,
        "api_key": SERPAPI_KEY,
        "gl": "us",
        "hl": "en",
        "lr": "lang_en",
        "cr": "countryUS"
    })
    url = f"https://serpapi.com/search.json?{params}"
    
    try:
        req = urllib.request.Request(url, headers=HEADERS)
        with urllib.request.urlopen(req, timeout=15) as resp:
            data = json.loads(resp.read().decode('utf-8'))
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
            print(f"✨ 本轮完成，新增 {success_count} 条真实有效英文邮箱。")
    except Exception as e:
        print(f"❌ [SerpApi 抓取异常] {e}")

# ==================== 6. AI 破冰信发送 ====================
def generate_ai_email(store_name, store_niche, website_url, lead_type):
    if lead_type == "HR_HIRING":
        prompt = f"""
        You are Alex, an AI talent consultant. Write a cold email in English to hiring manager at {store_name}.
        Industry: {store_niche}
        Angle: Hiring full-time takes 60+ days. Pitch vetted freelance AI specialists (via Fiverr) to deploy workflows immediately while interviewing.
        Requirements:
        - Must be strictly in English.
        - Subject: Under 6 words.
        - Body: Under 75 words. Casual, sign off as Alex.
        Output strictly in JSON: {{"subject": "...", "body": "..."}}
        """
    else:
        prompt = f"""
        You are Alex, a GEO & AI Growth Consultant. Write a cold email in English to founder/marketing lead of {store_name} ({website_url}).
        Category: {store_niche}
        Angle: AI Generative Engine Optimization (GEO). Help their brand get recommended by ChatGPT / Perplexity.
        Requirements:
        - Must be strictly in English.
        - Subject: Under 6 words.
        - Body: Under 70 words. Casual, sign off as Alex.
        Output strictly in JSON: {{"subject": "...", "body": "..."}}
        """

    res_str = safe_gemini_call(prompt)
    if res_str:
        try:
            return json.loads(res_str)
        except:
            pass
    
    return {
        "subject": f"Quick question regarding {store_name}",
        "body": f"Hi team at {store_name},\n\nNoticed your presence in {store_niche}. We help brands optimize for AI search recommendations (ChatGPT, Perplexity) and automate workflows.\n\nOpen to a quick overview?\n\nBest,\nAlex"
    }

def generate_ai_followup(client_reply_text, lead_type):
    aff_link = FIVERR_CONFIG['ai_transformation_deeplink']
    prompt = f"""
    You are Alex. Client replied: "{client_reply_text}". Lead type: {lead_type}
    Write a short follow-up in English under 80 words guiding them to check vetted AI experts here: {aff_link}.
    Sign off as Alex.
    Output strictly in JSON: {{"subject": "...", "body": "..."}}
    """
    res_str = safe_gemini_call(prompt)
    if res_str:
        try:
            return json.loads(res_str)
        except:
            pass
            
    return {
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

# ==================== 7. 流程控制与看板 ====================
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
        print(f"\n🤖 正在为有效目标 [{info['store_name']}] ({info['niche']}) 生成破冰邮件...")
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
        print("📭 [CRM 监控] 无需跟进的客户。")
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

            print(f"🔔 收到真实客户 [{target_email}] 的回复！")
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
    print(f"📈 【Termux SDR 运营看板 (真实有效邮箱防退信版)】 - {time.strftime('%Y-%m-%d %H:%M:%S')}")
    print("==================================================")
    print(f"🔥 今日发信: {today_count} / {GMAIL_CONFIG['daily_limit']} 封 | CRM 数据库总线索: {total_leads} 条")
    print(f"   * 📌 待发新线索 (NEW): {status_counts.get('NEW', 0)}")
    print(f"   * 📤 已发破冰信 (OUTREACH_SENT): {status_counts.get('OUTREACH_SENT', 0)}")
    print(f"   * 💬 已互动跟进 (REPLIED): {status_counts.get('REPLIED', 0)}")
    print(f"💰 Fiverr 追踪 ID: [510726]")
    print("==================================================\n")

# ==================== 8. 主守护进程入口 ====================
if __name__ == "__main__":
    print("==================================================")
    print("🚀 Termux AI SDR Agent (纯真实邮箱零退信引擎)")
    print("==================================================\n")
    
    LOOP_INTERVAL_MINUTES = 60
    
    try:
        while True:
            print(f"⏰ [{time.strftime('%Y-%m-%d %H:%M:%S')}] 唤醒执行任务...")
            
            try:
                auto_pipeline_serpapi_scraping()
            except Exception as e:
                print(f"❌ [抓取流程异常] {e}")

            try:
                auto_process_new_outreach()
            except Exception as e:
                print(f"❌ [发信流程异常] {e}")
                
            try:
                auto_check_and_followup()
            except Exception as e:
                print(f"❌ [回复流程异常] {e}")
                
            try:
                print_hourly_statistics()
            except Exception as e:
                print(f"❌ [看板输出异常] {e}")
                
            print(f"💤 执行完毕，休眠 {LOOP_INTERVAL_MINUTES} 分钟...\n")
            time.sleep(LOOP_INTERVAL_MINUTES * 60)
            
    except KeyboardInterrupt:
        print("\n👋 收到终止指令，进程退出。")
