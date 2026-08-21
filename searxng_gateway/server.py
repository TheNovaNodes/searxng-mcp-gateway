"""MCP-сервер searxng-mcp-gateway — веб-поиск для агентов OpenClaw.

Инструменты:
  - search_web(query, max_results, categories, language, safesearch):
    поиск через SearXNG, чистый JSON.
  - searxng_health(): диагностика доступности SearXNG.
  - deep_research(query, count): оркестрованное исследование + семантическая память.

Только сырые данные. Без LLM-синтеза.
Транспорт: stdio (по умолчанию) | streamable-http (сетевой деплой).
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

# ── Spayka: подключаем пакет memory_gateway из монорепо mcp-tools ───────
# deep_research тянет семпамять лабы (hybrid_search) в один вызов с вебом.
# Пакет memory_gateway не установлен в PYTHONPATH (только searxng-gateway),
# поэтому добавляем его родительский каталог в sys.path при старте.
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
except Exception:  # noqa: BLE001 — тихо, если пакет недоступен
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
    """Вызов SearXNG API + нормализация ответа."""
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

    # Нормализация — отдаём только нужное агентам
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
    """Веб-поиск через SearXNG. Сырые результаты без LLM-синтеза.

    Args:
        query: поисковый запрос на естественном языке.
        max_results: максимум результатов (1..50, по умолчанию 10).
        categories: категория поиска — general, images, news, videos, music, files, it, science, social media. None = general.
        language: язык результатов — auto, ru, en, de, ... (по умолчанию auto).
        safesearch: фильтр контента — 0=off, 1=moderate, 2=strict.
        engines: явный пул движков (напр. "google,bing") или категория движков. None = default.

    Returns:
        Чистый JSON: {query, count, results[], latency_ms}. Каждый результат:
        {title, url, content, engine, score, category, published_date?}.
    """
    max_results = max(1, min(50, max_results))
    try:
        out = _searxng_search(query, max_results, categories, language, safesearch, engines)
    except Exception as e:  # noqa: BLE001 — инструмент не должен ронять сервер
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
    """Диагностика SearXNG: доступность, версия, число движков.

    Returns:
        {status, searxng_url, reachable, status_code?, engines?, version?, error?}
    """
    info: Dict[str, Any] = {
        "status": "unknown",
        "searxng_url": config.SEARXNG_URL,
        "reachable": False,
    }
    try:
        # Проверка /search (быстрый healthcheck)
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

    # Попытка достать версию из /config
    try:
        cfg_resp = _requests.get(f"{config.SEARXNG_URL}/config", timeout=3)
        if cfg_resp.status_code == 200:
            cfg = cfg_resp.json()
            info["version"] = cfg.get("version", "unknown")
    except Exception:
        pass

    return info


# ---------------------------------------------------------------------------

from searxng_gateway import echelon

@mcp.tool(name="deep_research")
def deep_research(query_or_url: str, mode: str = "auto") -> Dict[str, Any]:
    """Глубокое исследование (Echelon Routing) с использованием премиальных API.

    Args:
        query_or_url: поисковый запрос или конкретный URL для краулинга.
        mode: режим роутинга - 'auto', 'semantic' (Exa), 'scrape' (Firecrawl/Olostep).

    Returns:
        Синтезированный ответ или выкачанный Markdown.
    """
    result: Dict[str, Any] = {
        "query": query_or_url,
        "mode": mode,
        "provider": None,
        "answer": None,
        "degraded": False,
    }
    
    is_url = query_or_url.startswith("http://") or query_or_url.startswith("https://")
    
    if mode == "semantic" and not is_url:
        # Echelon 2: Exa AI Semantic Search
        exa_res = echelon.call_exa(query_or_url)
        result["provider"] = "exa"
        result["answer"] = exa_res
        if "error" in exa_res:
            result["degraded"] = True
            
    elif is_url or mode == "scrape":
        # Echelon 3: Firecrawl
        fc_res = echelon.call_firecrawl_scrape(query_or_url)
        result["provider"] = "firecrawl"
        
        # WAF Detection & Fallback to Echelon 4 (Olostep)
        if "error" in fc_res and echelon.detect_waf(fc_res):
            result["provider"] = "firecrawl -> olostep (WAF Bypassed)"
            olo_res = echelon.call_olostep_scrape(query_or_url)
            result["answer"] = olo_res
            if "error" in olo_res:
                result["degraded"] = True
        else:
            result["answer"] = fc_res
            if "error" in fc_res:
                result["degraded"] = True
                
    else:
        # Mode auto and is a search query -> Echelon 1: Tavily Unlimited
        tav_res = echelon.call_tavily(query_or_url)
        result["provider"] = "tavily"
        result["answer"] = tav_res
        if "error" in tav_res:
            result["degraded"] = True
            
    return result


# ---------------------------------------------------------------------------

from searxng_gateway import orchestrator
from searxng_gateway import vault
import concurrent.futures

@mcp.tool(name="hybrid_search")
def hybrid_search(query: str, max_results: int = 10) -> Dict[str, Any]:
    """Гибридный поиск: параллельно опрашивает SearXNG и платные API (Quota-Aware Waterfall), склеивая результаты через RRF.
    
    Args:
        query: поисковый запрос
        max_results: максимум результатов (до 50)
    """
    max_results = max(1, min(50, max_results))
    
    with concurrent.futures.ThreadPoolExecutor(max_workers=2) as executor:
        future_searx = executor.submit(_searxng_search, query, max_results)
        future_deep = executor.submit(orchestrator.deep_search_cascade, query, max_results)
        
        done, not_done = concurrent.futures.wait(
            [future_searx, future_deep], 
            timeout=20, 
            return_when=concurrent.futures.ALL_COMPLETED
        )
        
        searx_res = {"results": []}
        deep_res = {"results": []}
        degraded = False
        
        if future_searx in done:
            try:
                searx_res = future_searx.result()
            except Exception:
                degraded = True
        else:
            degraded = True
            
        if future_deep in done:
            try:
                deep_res = future_deep.result()
            except Exception:
                degraded = True
        else:
            degraded = True

    list1 = searx_res.get("results", [])
    list2 = deep_res.get("results", [])
    
    fused = orchestrator.reciprocal_rank_fusion([list1, list2])
    
    return {
        "query": query,
        "provider_deep": deep_res.get("provider", "none"),
        "count": len(fused[:max_results]),
        "results": fused[:max_results],
        "degraded": degraded
    }

@mcp.tool(name="ecosystem_health")
def ecosystem_health() -> Dict[str, Any]:
    """Расширенная диагностика всей поисковой экосистемы (SearXNG + Внешние API)."""
    # Вызываем старый хелсчек для базы
    s_health = searxng_health()
    
    health = {
        "searxng": s_health.get("status", "unknown"),
        "searxng_latency_ms": s_health.get("latency_ms", 0),
        "api_echelons": {}
    }
    
    # Считываем состояние ключей из CircuitBreaker'а
    for provider_name, balancer in echelon.balancers.items():
        keys = vault.get_keys(provider_name)
        total_keys = len(keys)
        if total_keys == 0:
            health["api_echelons"][provider_name] = "🔴 Down (No Keys)"
            continue
            
        available = sum(1 for k in keys if echelon.cb.is_available(k))
        
        if available == total_keys:
            health["api_echelons"][provider_name] = f"🟢 OK ({available}/{total_keys} keys ready)"
        elif available > 0:
            health["api_echelons"][provider_name] = f"🟡 Degraded ({available}/{total_keys} keys ready)"
        else:
            health["api_echelons"][provider_name] = "🔴 Down (All keys on cooldown)"
            
    return health

# ---------------------------------------------------------------------------

def main():
    """Точка входа для CLI (pyproject.toml script)."""
    mcp.run(transport="stdio")


if __name__ == "__main__":
    main()
