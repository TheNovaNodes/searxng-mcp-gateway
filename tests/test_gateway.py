import pytest
from unittest.mock import patch, MagicMock
from searxng_gateway.server import search_web, searxng_health, deep_research

def test_search_web_success():
    with patch("requests.get") as mock_get:
        mock_resp = MagicMock()
        mock_resp.json.return_value = {
            "results": [{"title": "Test", "url": "http://test", "content": "hello"}],
            "answers": []
        }
        mock_resp.raise_for_status = MagicMock()
        mock_get.return_value = mock_resp
        
        res = search_web("test")
        assert res["count"] == 1
        assert "latency_ms" in res
        assert not res.get("degraded", False)

def test_search_web_exception():
    with patch("requests.get") as mock_get:
        mock_get.side_effect = Exception("SearXNG Timeout")
        res = search_web("test")
        assert res["count"] == 0
        assert res["degraded"]
        assert "error" in res
        assert "SearXNG Timeout" in res["error"]

def test_searxng_health():
    with patch("requests.get") as mock_get:
        mock_resp = MagicMock()
        mock_resp.status_code = 200
        mock_resp.json.return_value = {"results": []}
        mock_get.return_value = mock_resp
        
        res = searxng_health()
        assert res["status"] == "ok"
        assert res["reachable"]

def test_deep_research_success():
    with patch("subprocess.run") as mock_run:
        mock_run.return_value = MagicMock(stdout="Research Answer", stderr="", returncode=0)
        res = deep_research("test")
        assert "Research Answer" in res["answer"]
        # Assuming memory_gateway is not available in the test env, it should fallback
        assert res["semantic_memory"]["degraded"]
