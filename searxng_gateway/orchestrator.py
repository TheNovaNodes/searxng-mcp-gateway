import concurrent.futures
import time
from typing import Dict, Any, List

from searxng_gateway import echelon

def deep_search_cascade(query: str, max_results: int) -> Dict[str, Any]:
    """Каскадный поиск через платные API (Quota-Aware Waterfall)."""
    # 1. Пробуем Tavily
    tav_res = echelon.call_tavily(query, depth="basic")
    if "error" not in tav_res and tav_res.get("results"):
        results = []
        for r in tav_res.get("results", [])[:max_results]:
            results.append({
                "title": r.get("title", ""),
                "url": r.get("url", ""),
                "content": r.get("content", ""),
                "engine": "tavily",
            })
        return {"provider": "tavily", "results": results}
    
    # 2. Fallback на Exa AI
    exa_res = echelon.call_exa(query)
    if "error" not in exa_res and exa_res.get("results"):
        results = []
        for r in exa_res.get("results", [])[:max_results]:
            results.append({
                "title": r.get("title", ""),
                "url": r.get("url", ""),
                "content": r.get("text", "") or r.get("snippet", ""),
                "engine": "exa",
            })
        return {"provider": "exa", "results": results}
        
    return {"provider": "none", "results": [], "error": "All deep search APIs are on cooldown or returned empty."}


def reciprocal_rank_fusion(lists: List[List[Dict[str, Any]]], k: int = 60) -> List[Dict[str, Any]]:
    """Алгоритм RRF (Reciprocal Rank Fusion) для дедупликации и слияния."""
    scores = {}
    items_by_url = {}
    
    for result_list in lists:
        seen_in_this_list = set()
        for rank, item in enumerate(result_list):
            url = item.get("url")
            if not url or type(url) is not str:
                continue
            
            # Нормализация (убираем фрагменты и trailing slash)
            url_clean = url.split('#')[0].rstrip('/')
            
            if url_clean in seen_in_this_list:
                continue
            seen_in_this_list.add(url_clean)
            
            eng = item.get("engine")
            
            if url_clean not in scores:
                scores[url_clean] = 0.0
                items_by_url[url_clean] = item.copy()
                items_by_url[url_clean]["engines"] = [eng] if eng else []
            else:
                if eng and eng not in items_by_url[url_clean]["engines"]:
                    items_by_url[url_clean]["engines"].append(eng)
            
            scores[url_clean] += 1.0 / (k + rank + 1)
            
    sorted_urls = sorted(scores.keys(), key=lambda u: scores[u], reverse=True)
    
    final_results = []
    for u in sorted_urls:
        final_item = items_by_url[u]
        final_item["rrf_score"] = round(scores[u], 4)
        final_results.append(final_item)
        
    return final_results
