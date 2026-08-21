import os
import re

VAULT_DIR = "/dev/shm/agent_vault"

POINTERS = {
    "tavily": "gsAYQxw",
    "firecrawl": "x01eFQ",
    "exa": "H0bWdb",
    "olostep": "Ib50He",
    "tinyfish": "u7jMwl"
}

def get_keys(provider: str) -> list[str]:
    """Dynamically reads the keys for a given provider from the SHM Vault."""
    pointer = POINTERS.get(provider)
    if not pointer:
        return []
    
    path = os.path.join(VAULT_DIR, pointer)
    if not os.path.exists(path):
        return []
        
    with open(path, "r", encoding="utf-8") as f:
        content = f.read()
        
    if provider == "tavily":
        return re.findall(r"tvly-dev-[a-zA-Z0-9_-]+", content)
    elif provider == "firecrawl":
        return re.findall(r"fc-[a-zA-Z0-9_-]+", content)
    elif provider == "exa":
        # Exa keys are usually 32+ hex chars or UUIDs
        keys = re.findall(r"(?:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}|[A-Za-z0-9_-]{32,})", content)
        # Filter out "EXA" title or other non-keys
        return [k for k in keys if len(k) >= 32]
    elif provider == "olostep":
        return re.findall(r"olostep_[a-zA-Z0-9_-]+", content)
    elif provider == "tinyfish":
        # Usually starts with Key or just a long string.
        # Format: Key Tiny fish ... API key is often long hex or UUID. Let's grab long words.
        # We can just look for a UUID or long string.
        keys = re.findall(r"[A-Za-z0-9_-]{32,}", content)
        return keys
        
    return []

