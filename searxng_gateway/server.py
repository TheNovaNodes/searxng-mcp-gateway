"""MCP server searxng-mcp-gateway — web search for OpenClaw agents.

Tools:
  - search_web(query, max_results, categories, language, safesearch):
    search via SearXNG, returning clean JSON.
  - searxng_health(): diagnostics of SearXNG availability.
  - deep_research(query, count): orchestrated research + semantic memory.

Raw data only. No LLM synthesis.
Transport: stdio (default) | streamable-http (network deploy).
"""

__all__ = ["search_web", "searxng_health", "deep_research", "mcp"]
__version__ = "1.0.0"
import os
import subprocess
import time
from typing import Any, Dict, Optional

import requests as _requests

try:
    from mcp.server.fastmcp import FastMCP
except ImportError:
    try:
        from fastmcp import FastMCP
    except ImportError:
        from mcp.server.mcpserver import MCPServer as FastMCP

from . import config

# ── Integration: connecting memory_gateway package from mcp-tools monorepo ───────
# deep_research pulls hybrid_search in a single call with web search.
# The memory_gateway package might not be in PYTHONPATH (only searxng-gateway),
# so we add its parent directory to sys.path on startup.
import sys as _sys
_HERE = os.path.dirname(os.path.abspath(__file__))
_MCP_TOOLS = os.path.dirname(os.path.dirname(_HERE))
_MG_PKG = os.path.join(_MCP_TOOLS, "memory-gateway")
for _p in (_MG_PKG, _MCP_TOOLS):
    if _p not in _sys.path:
        _sys.path.insert(0, _p)

try:
    from memory_gateway.search import hybrid_search as _mg_hybrid_search
    _MG_AVAILABLE = True
except Exception:  # noqa: BLE001 — fail silently if package is unavailable
    _mg_hybrid_search = None
    _MG_AVAILABLE = False

try:
    mcp = FastMCP("searxng-mcp-gateway", host=config.HOST, port=config.PORT)
except TypeError:
    mcp = FastMCP("searxng-mcp-gateway")

# ---------------------------------------------------------------------------


def _searxng_search(
    query: str,
    max_results: int = config.DEFAULT_MAX_RESULTS,
    categories: Optional[str] = None,
    language: str = config.DEFAULT_LANGUAGE,
    safesearch: int = config.DEFAULT_SAFESEARCH,
    engines: Optional[str] = None,
    timeout: int = config.DEFAULT_TIMEOUT,
) -> Dict[str, Any]:
    """Call SearXNG API + normalize the response."""
    params: Dict[str, Any] = {
        "q": query,
        "format": "json",
        "language": language,
        "safesearch": safesearch,
    }
    if categories:
        params["categories"] = categories
    if engines:
        params["engines"] = engines

    url = f"{config.SEARXNG_URL}/search"
    t0 = time.time()
    resp = _requests.get(url, params=params, timeout=timeout)
    latency_ms = round((time.time() - t0) * 1000.0, 1)
    resp.raise_for_status()
    raw = resp.json()

    # Normalization — providing only what agents need
    results = []
    for r in raw.get("results", [])[:max_results]:
        item = {
            "title": r.get("title", ""),
            "url": r.get("url", ""),
            "content": r.get("content", ""),
            "engine": r.get("engine", ""),
            "engines": r.get("engines", []),
            "score": r.get("score", 0.0),
            "category": r.get("category", ""),
        }
        pub = r.get("publishedDate")
        if pub:
            item["published_date"] = pub
        results.append(item)

    return {
        "query": query,
        "count": len(results),
        "results": results,
        "answers": raw.get("answers", []),
        "infoboxes": raw.get("infoboxes", []),
        "suggestions": raw.get("suggestions", []),
        "unresponsive_engines": raw.get("unresponsive_engines", []),
        "latency_ms": latency_ms,
    }


# ---------------------------------------------------------------------------


@mcp.tool(name="search_web")
def search_web(
    query: str,
    max_results: int = config.DEFAULT_MAX_RESULTS,
    categories: Optional[str] = None,
    language: str = config.DEFAULT_LANGUAGE,
    safesearch: int = config.DEFAULT_SAFESEARCH,
    engines: Optional[str] = None,
) -> Dict[str, Any]:
    """Web search via SearXNG. Raw results without LLM synthesis.

    Args:
        query: natural language search query.
        max_results: maximum results (1..50, default 10).
        categories: search category — general, images, news, videos, music, files, it, science, social media. None = general.
        language: results language — auto, ru, en, de, ... (default auto).
        safesearch: content filter — 0=off, 1=moderate, 2=strict.
        engines: explicit pool of engines (e.g. "google,bing") or engine category. None = default.

    Returns:
        Clean JSON: {query, count, results[], latency_ms}. Each result:
        {title, url, content, engine, score, category, published_date?}.
    """
    max_results = max(1, min(50, max_results))
    try:
        out = _searxng_search(query, max_results, categories, language, safesearch, engines)
    except Exception as e:  # noqa: BLE001 — tool must not crash the server
        return {
            "query": query,
            "count": 0,
            "results": [],
            "degraded": True,
            "error": f"{type(e).__name__}: {e}",
            "latency_ms": 0,
        }
    return out


@mcp.tool(name="searxng_health")
def searxng_health() -> Dict[str, Any]:
    """Diagnostics of SearXNG: availability, version, number of engines.

    Returns:
        {status, searxng_url, reachable, status_code?, engines?, version?, error?}
    """
    info: Dict[str, Any] = {
        "status": "unknown",
        "searxng_url": config.SEARXNG_URL,
        "reachable": False,
    }
    try:
        # Check /search (quick healthcheck)
        t0 = time.time()
        resp = _requests.get(
            f"{config.SEARXNG_URL}/search",
            params={"q": "healthcheck", "format": "json"},
            timeout=5,
        )
        info["latency_ms"] = round((time.time() - t0) * 1000.0, 1)
        info["status_code"] = resp.status_code
        info["reachable"] = resp.status_code == 200

        if resp.status_code == 200:
            data = resp.json()
            info["result_count"] = len(data.get("results", []))
            info["unresponsive_engines"] = data.get("unresponsive_engines", [])
            info["status"] = "ok"
        else:
            info["status"] = "degraded"

    except Exception as e:  # noqa: BLE001
        info["status"] = "down"
        info["error"] = f"{type(e).__name__}: {e}"

    # Attempt to get version from /config
    try:
        cfg_resp = _requests.get(f"{config.SEARXNG_URL}/config", timeout=3)
        if cfg_resp.status_code == 200:
            cfg = cfg_resp.json()
            info["version"] = cfg.get("version", "unknown")
    except Exception:
        pass

    return info


# ---------------------------------------------------------------------------

@mcp.tool(name="deep_research")
def deep_research(query: str, count: int = 10) -> Dict[str, Any]:
    """Deep research + semantic memory in a SINGLE call.

    Combiner: web (orchestrator /research, fan-out Tavily/Firecrawl/TinyFish/
    SearXNG + merge + synthesis) plus semantic memory
    (memory-gateway.hybrid_search: vector ALM + lexical FTS5). Both layers
    return together; if one is unavailable — degraded, not a crash.

    Args:
        query: research question.
        count: number of results per web layer provider (default 10).

    Returns:
        {query, answer (web synthesis), semantic_memory, degraded?, error?}
    """
    result: Dict[str, Any] = {
        "query": query,
        "answer": None,
        "semantic_memory": None,
        "degraded": False,
    }
    # ── Web layer (orchestrator) ──────────────────────────────────────────
    orchestrator = config.DEEP_RESEARCH_ORCHESTRATOR
    if not orchestrator or not os.path.exists(orchestrator):
        result["degraded"] = True
        result["error"] = "DEEP_RESEARCH_ORCHESTRATOR not configured or missing"
    else:
        try:
            proc = subprocess.run(
                [orchestrator, query, "deep_research", str(count)],
                capture_output=True,
                text=True,
                timeout=config.DEEP_RESEARCH_TIMEOUT,
            )
            out = (proc.stdout or "").strip() or (proc.stderr or "").strip()
            result["answer"] = out or "No research output."
            if proc.returncode != 0:
                result["degraded"] = True
        except Exception as e:  # noqa: BLE001 — tool must not crash the server
            result["degraded"] = True
            result["error"] = f"{type(e).__name__}: {e}"
    # ── Semantic layer (memory) ────────────────────────────────
    if config.SEMANTIC_ENABLED and _MG_AVAILABLE:
        try:
            sem = _mg_hybrid_search(
                query,
                config.SEMANTIC_TOP_K,
                expand_context=config.SEMANTIC_EXPAND,
                fusion=config.SEMANTIC_FUSION,
            )
            result["semantic_memory"] = sem
            if sem.get("degraded"):
                result["degraded"] = result["degraded"] or True
        except Exception as e:  # noqa: BLE001
            result["semantic_memory"] = {
                "query": query,
                "count": 0,
                "results": [],
                "degraded": True,
                "error": f"{type(e).__name__}: {e}",
            }
            result["degraded"] = result["degraded"] or True
    else:
        result["semantic_memory"] = {
            "degraded": True,
            "error": (
                "memory_gateway unavailable"
                if not _MG_AVAILABLE
                else "disabled (SEMANTIC_ENABLED=0)"
            ),
        }
    return result


# ---------------------------------------------------------------------------

def main():
    """Entry point for CLI (pyproject.toml script)."""
    mcp.run(transport="stdio")


if __name__ == "__main__":
    main()
