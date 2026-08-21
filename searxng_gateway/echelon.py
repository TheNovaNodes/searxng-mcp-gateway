import time
import json
import urllib.request
import urllib.error
import threading
from typing import Dict, Any, Optional

from searxng_gateway import vault

class CircuitBreaker:
    def __init__(self, cooldown_seconds: int = 60):
        self.cooldowns: Dict[str, float] = {}
        self.cooldown_seconds = cooldown_seconds
        self.lock = threading.Lock()
        
    def record_failure(self, key: str):
        with self.lock:
            self.cooldowns[key] = time.time()
        
    def is_available(self, key: str) -> bool:
        with self.lock:
            last_failure = self.cooldowns.get(key)
            if not last_failure:
                return True
            if time.time() - last_failure > self.cooldown_seconds:
                del self.cooldowns[key]
                return True
            return False

cb = CircuitBreaker(60)

class RoundRobinBalancer:
    def __init__(self):
        self.idx = 0
        self.lock = threading.Lock()

    def get_next(self, keys: list[str]) -> Optional[str]:
        if not keys:
            return None
        with self.lock:
            for _ in range(len(keys)):
                k = keys[self.idx % len(keys)]
                self.idx += 1
                if cb.is_available(k):
                    return k
        return None

balancers = {
    "tavily": RoundRobinBalancer(),
    "exa": RoundRobinBalancer(),
    "firecrawl": RoundRobinBalancer(),
    "olostep": RoundRobinBalancer()
}

def handle_http_error(e: urllib.error.HTTPError, key: str) -> Dict[str, Any]:
    # 401, 403 (Unauthorized/Forbidden) or 429 (Rate Limit) -> Trigger Circuit Breaker
    if e.code in [401, 403, 429]:
        cb.record_failure(key)
    return {"error": f"HTTP {e.code}", "status_code": e.code}


def call_tavily(query: str, depth: str = "advanced") -> Dict[str, Any]:
    keys = vault.get_keys("tavily")
    selected_key = balancers["tavily"].get_next(keys)
    if not selected_key:
        return {"error": "All Tavily keys are currently on cooldown or missing."}

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
        return handle_http_error(e, selected_key)
    except Exception as e:
        return {"error": str(e)}

def call_exa(query: str) -> Dict[str, Any]:
    keys = vault.get_keys("exa")
    selected_key = balancers["exa"].get_next(keys)
    if not selected_key:
        return {"error": "All Exa keys are currently on cooldown or missing."}
    
    req = urllib.request.Request(
        "https://api.exa.ai/search", 
        headers={"x-api-key": selected_key, "Content-Type": "application/json"}, 
        data=json.dumps({"query": query, "useAutoprompt": True}).encode("utf-8"), 
        method="POST"
    )
    try:
        with urllib.request.urlopen(req, timeout=15) as res:
            return json.loads(res.read().decode())
    except urllib.error.HTTPError as e:
        return handle_http_error(e, selected_key)
    except Exception as e:
        return {"error": str(e)}

def call_firecrawl_scrape(target_url: str) -> Dict[str, Any]:
    keys = vault.get_keys("firecrawl")
    selected_key = balancers["firecrawl"].get_next(keys)
    if not selected_key:
        return {"error": "All Firecrawl keys are currently on cooldown or missing."}
        
    req = urllib.request.Request(
        "https://api.firecrawl.dev/v1/scrape", 
        headers={"Authorization": f"Bearer {selected_key}", "Content-Type": "application/json"}, 
        data=json.dumps({"url": target_url}).encode("utf-8"), 
        method="POST"
    )
    try:
        with urllib.request.urlopen(req, timeout=20) as res:
            return json.loads(res.read().decode())
    except urllib.error.HTTPError as e:
        return handle_http_error(e, selected_key)
    except Exception as e:
        return {"error": str(e)}

def call_olostep_scrape(target_url: str) -> Dict[str, Any]:
    keys = vault.get_keys("olostep")
    selected_key = balancers["olostep"].get_next(keys)
    if not selected_key:
        return {"error": "All Olostep keys are currently on cooldown or missing."}
        
    req = urllib.request.Request(
        "https://api.olostep.com/v1/scrape", 
        headers={"Authorization": f"Bearer {selected_key}", "Content-Type": "application/json"}, 
        data=json.dumps({"url": target_url, "bypass_waf": True}).encode("utf-8"), 
        method="POST"
    )
    try:
        with urllib.request.urlopen(req, timeout=25) as res:
            return json.loads(res.read().decode())
    except urllib.error.HTTPError as e:
        return handle_http_error(e, selected_key)
    except Exception as e:
        return {"error": str(e)}


def detect_waf(error_dict: Dict[str, Any]) -> bool:
    if error_dict.get("status_code") in [403, 401]:
        return True
    return False

