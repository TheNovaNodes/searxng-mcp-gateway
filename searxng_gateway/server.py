"""MCP-сервер searxng-gateway — веб-поиск для агентов OpenClaw.

Инструменты:
  - search_web(query, max_results, categories, language, safesearch):
    поиск через SearXNG, чистый JSON.
  - searxng_health(): диагностика доступности SearXNG.

Только сырые данные. Без LLM-синтеза.
Транспорт: stdio (по умолчанию) | streamable-http (сетевой деплой).
"""
import os
import subprocess
import time
from typing import Any, Dict, Optional

import requests as _requests
from mcp.server.fastmcp import FastMCP

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

mcp = FastMCP("searxng-gateway", host=config.HOST, port=config.PORT)

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

@mcp.tool(name="deep_research")
def deep_research(query: str, count: int = 10) -> Dict[str, Any]:
    """Глубокое исследование + семантическая память лабы в ОДНОМ вызове.

    Комбайн: веб (оркестратор /research, fan-out Tavily/Firecrawl/TinyFish/
    SearXNG + merge + синтез) плюс семантическая память лабы
    (memory-gateway.hybrid_search: vector ALM + lexical FTS5). Оба слоя
    возвращаются вместе; при недоступности одного — degraded, не падение.

    Args:
        query: исследовательский вопрос.
        count: число результатов на провайдера веб-слоя (по умолчанию 10).

    Returns:
        {query, answer (веб-синтез), semantic_memory (лаба), degraded?, error?}
    """
    result: Dict[str, Any] = {
        "query": query,
        "answer": None,
        "semantic_memory": None,
        "degraded": False,
    }
    # ── Веб-слой (оркестратор) ──────────────────────────────────────────
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
        except Exception as e:  # noqa: BLE001 — тул не должен ронять сервер
            result["degraded"] = True
            result["error"] = f"{type(e).__name__}: {e}"
    # ── Семантический слой (память лабы) ────────────────────────────────
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
    """Точка входа для CLI (pyproject.toml script)."""
    mcp.run(transport="stdio")


if __name__ == "__main__":
    main()
