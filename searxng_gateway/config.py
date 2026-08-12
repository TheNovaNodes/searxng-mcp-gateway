"""Configuration for searxng-mcp-gateway."""
import os

SEARXNG_URL = os.getenv("SEARXNG_URL", "http://127.0.0.1:8081")
DEFAULT_MAX_RESULTS = int(os.getenv("SEARXNG_MAX_RESULTS", "10"))
DEFAULT_LANGUAGE = os.getenv("SEARXNG_DEFAULT_LANGUAGE", "auto")
DEFAULT_SAFESEARCH = int(os.getenv("SEARXNG_SAFESEARCH", "0"))  # 0=off, 1=moderate, 2=strict
DEFAULT_TIMEOUT = int(os.getenv("SEARXNG_TIMEOUT", "30"))  # seconds
HOST = os.getenv("SEARXNG_HOST", "127.0.0.1")
PORT = int(os.getenv("SEARXNG_PORT", "8092"))

# ── Deep Research (distilled from lab-research, adapted) ───────────────
# Path to the /research orchestrator (fan-out across providers: Tavily/Firecrawl/
# TinyFish/SearXNG, merge + dedup + freshness + synthesis).
DEEP_RESEARCH_ORCHESTRATOR = os.getenv(
    "DEEP_RESEARCH_ORCHESTRATOR",
    "/path/to/search-orchestrator.sh",
)
# Heavy pipeline - generous timeout, but not infinite.
DEEP_RESEARCH_TIMEOUT = int(os.getenv("DEEP_RESEARCH_TIMEOUT", "240"))  # seconds

# ── Semantic Memory fusion (integration with memory-gateway) ───────────────────
# deep_research pulls both web (orchestrator) and semantic memory
# (memory-gateway.hybrid_search) in a SINGLE call. Graceful degradation: if
# the memory_gateway package is unavailable or disabled, web search still works,
# and semantic memory is marked as degraded.
SEMANTIC_ENABLED = bool(int(os.getenv("SEMANTIC_ENABLED", "0")))  # OFF by default (experimental)
SEMANTIC_TOP_K = int(os.getenv("SEMANTIC_TOP_K", "5"))
SEMANTIC_EXPAND = bool(int(os.getenv("SEMANTIC_EXPAND", "1")))
SEMANTIC_FUSION = os.getenv("SEMANTIC_FUSION", "weighted")  # weighted | rrf
