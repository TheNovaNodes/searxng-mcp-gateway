import time
import json
import urllib.request
import urllib.error
from typing import Dict, Any, Optional

from searxng_gateway import vault

class CircuitBreaker:
    def __init__(self, cooldown_seconds: int = 60):
        self.cooldowns: Dict[str, float] = {}
        self.cooldown_seconds = cooldown_seconds
        
    def record_failure(self, key: str):
        self.cooldowns[key] = time.time()
        
    def is_available(self, key: str) -> bool:
        last_failure = self.cooldowns.get(key)
        if not last_failure:
            return True
        if time.time() - last_failure > self.cooldown_seconds:
            del self.cooldowns[key]
            return True
        return False

cb = CircuitBreaker(60)
tavily_idx = 0

def call_tavily(query: str, depth: str = "advanced") -> Dict[str, Any]:
    global tavily_idx
    keys = vault.get_keys("tavily")
    if not keys:
        return {"error": "No Tavily keys found in Vault."}
        
    # Round-robin
    start_idx = tavily_idx
    selected_key = None
    for _ in range(len(keys)):
        k = keys[tavily_idx % len(keys)]
        tavily_idx += 1
        if cb.is_available(k):
            selected_key = k
            break
            
    if not selected_key:
        return {"error": "All Tavily keys are currently on cooldown."}

    url = "https://api.tavily.com/search"
    headers = {"Content-Type": "application/json"}
    payload = json.dumps({
        "api_key": selected_key,
        "query": query,
        "search_depth": depth,
        "include_answer": True,
        "include_images": False,
        "include_raw_content": False
    }).encode("utf-8")
    
    req = urllib.request.Request(url, data=payload, headers=headers, method="POST")
    try:
        with urllib.request.urlopen(req, timeout=15) as response:
            return json.loads(response.read().decode())
    except urllib.error.HTTPError as e:
        if e.code == 429:
            cb.record_failure(selected_key)
        return {"error": f"HTTP {e.code}", "status_code": e.code}
    except Exception as e:
        return {"error": str(e)}

def call_exa(query: str) -> Dict[str, Any]:
    keys = vault.get_keys("exa")
    if not keys:
        return {"error": "No Exa keys found in Vault."}
    
    req = urllib.request.Request(
        "https://api.exa.ai/search", 
        headers={"x-api-key": keys[0], "Content-Type": "application/json"}, 
        data=json.dumps({"query": query, "useAutoprompt": True}).encode("utf-8"), 
        method="POST"
    )
    try:
        with urllib.request.urlopen(req, timeout=15) as res:
            return json.loads(res.read().decode())
    except urllib.error.HTTPError as e:
        return {"error": f"HTTP {e.code}", "status_code": e.code}
    except Exception as e:
        return {"error": str(e)}

def call_firecrawl_scrape(target_url: str) -> Dict[str, Any]:
    keys = vault.get_keys("firecrawl")
    if not keys:
        return {"error": "No Firecrawl keys found in Vault."}
        
    req = urllib.request.Request(
        "https://api.firecrawl.dev/v1/scrape", 
        headers={"Authorization": f"Bearer {keys[0]}", "Content-Type": "application/json"}, 
        data=json.dumps({"url": target_url}).encode("utf-8"), 
        method="POST"
    )
    try:
        with urllib.request.urlopen(req, timeout=20) as res:
            return json.loads(res.read().decode())
    except urllib.error.HTTPError as e:
        return {"error": f"HTTP {e.code}", "status_code": e.code}
    except Exception as e:
        return {"error": str(e)}

def call_olostep_scrape(target_url: str) -> Dict[str, Any]:
    keys = vault.get_keys("olostep")
    if not keys:
        return {"error": "No Olostep keys found in Vault."}
    # Note: Using mock endpoint if exact olostep isn't known, or we return standard
    return {"error": "Olostep client requires precise payload (simulated WAF bypass)."}

def detect_waf(error_dict: Dict[str, Any]) -> bool:
    # If a scrape fails with 403 or 429, it might be a WAF
    if error_dict.get("status_code") in [403, 401, 429]:
        return True
    return False
