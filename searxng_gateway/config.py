"""Конфигурация searxng-mcp-gateway."""
import os

SEARXNG_URL = os.getenv("SEARXNG_URL", "http://127.0.0.1:8081")
DEFAULT_MAX_RESULTS = int(os.getenv("SEARXNG_DEFAULT_MAX", "10"))
DEFAULT_LANGUAGE = os.getenv("SEARXNG_DEFAULT_LANG", "auto")
DEFAULT_SAFESEARCH = int(os.getenv("SEARXNG_SAFESEARCH", "0"))  # 0=off, 1=moderate, 2=strict
DEFAULT_TIMEOUT = int(os.getenv("SEARXNG_TIMEOUT", "10"))  # seconds
HOST = os.getenv("SEARXNG_HOST", "127.0.0.1")
PORT = int(os.getenv("SEARXNG_PORT", "8092"))

# ── Deep Research (distilled from lab-research, adapted) ───────────────
# Путь к оркестратору /research (fan-out по провайдерам: Tavily/Firecrawl/
# TinyFish/SearXNG, merge + dedup + freshness + synthesis).
DEEP_RESEARCH_ORCHESTRATOR = os.getenv(
    "DEEP_RESEARCH_ORCHESTRATOR",
    "/path/to/search-orchestrator.sh",
)
# Тяжёлый пайплайн — щедрый таймаут, но не бесконечный.
DEEP_RESEARCH_TIMEOUT = int(os.getenv("DEEP_RESEARCH_TIMEOUT", "240"))  # seconds

# ── Semantic Memory fusion (spayka с memory-gateway) ───────────────────
# deep_research тянет в ОДНОМ вызове и веб (оркестратор), и семпамять
# лабы (memory-gateway.hybrid_search). Грациозная деградация: если
# пакет memory_gateway недоступен или выключен — веб работает, semantic
# помечается degraded.
SEMANTIC_ENABLED = bool(int(os.getenv("SEMANTIC_ENABLED", "0")))  # OFF by default (experimental)
SEMANTIC_TOP_K = int(os.getenv("SEMANTIC_TOP_K", "5"))
SEMANTIC_EXPAND = bool(int(os.getenv("SEMANTIC_EXPAND", "1")))
SEMANTIC_FUSION = os.getenv("SEMANTIC_FUSION", "weighted")  # weighted | rrf
