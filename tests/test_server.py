"""Unit tests for searxng-mcp-gateway server tools."""

import subprocess
import unittest
from unittest.mock import MagicMock, patch

import requests

from searxng_gateway import config, server


class TestServer(unittest.TestCase):
    """Test MCP tools provided by searxng-mcp-gateway."""

    @patch("searxng_gateway.server._requests.get")
    def test_search_web_success(self, mock_get):
        mock_response = MagicMock()
        mock_response.status_code = 200
        mock_response.json.return_value = {
            "query": "python mcp",
            "results": [
                {
                    "title": "Model Context Protocol",
                    "url": "https://modelcontextprotocol.io",
                    "content": "MCP specification and SDKs",
                    "engine": "duckduckgo",
                    "engines": ["duckduckgo", "google"],
                    "score": 1.5,
                    "category": "general",
                    "publishedDate": "2026-01-01",
                },
                {
                    "title": "SearXNG Documentation",
                    "url": "https://docs.searxng.org",
                    "content": "SearXNG metasearch engine docs",
                    "engine": "google",
                    "engines": ["google"],
                    "score": 1.2,
                    "category": "it",
                },
            ],
            "answers": ["Answers list"],
            "infoboxes": [{"infobox": "info"}],
            "suggestions": ["python mcp sdk"],
            "unresponsive_engines": [],
        }
        mock_get.return_value = mock_response

        res = server.search_web(
            query="python mcp",
            max_results=5,
            categories="general",
            language="en",
            safesearch=1,
            engines="duckduckgo,google",
        )

        self.assertEqual(res["query"], "python mcp")
        self.assertEqual(res["count"], 2)
        self.assertEqual(len(res["results"]), 2)
        self.assertEqual(res["results"][0]["title"], "Model Context Protocol")
        self.assertEqual(res["results"][0]["published_date"], "2026-01-01")
        self.assertEqual(res["results"][1]["title"], "SearXNG Documentation")
        self.assertNotIn("published_date", res["results"][1])
        self.assertEqual(res["answers"], ["Answers list"])
        self.assertEqual(res["suggestions"], ["python mcp sdk"])
        self.assertIn("latency_ms", res)

        # Verify call params
        mock_get.assert_called_once()
        args, kwargs = mock_get.call_args
        self.assertTrue(args[0].startswith(config.SEARXNG_URL))
        self.assertEqual(kwargs["params"]["q"], "python mcp")
        self.assertEqual(kwargs["params"]["categories"], "general")
        self.assertEqual(kwargs["params"]["engines"], "duckduckgo,google")
        self.assertEqual(kwargs["params"]["language"], "en")
        self.assertEqual(kwargs["params"]["safesearch"], 1)

    @patch("searxng_gateway.server._requests.get")
    def test_search_web_clamping(self, mock_get):
        mock_response = MagicMock()
        mock_response.status_code = 200
        mock_response.json.return_value = {"results": []}
        mock_get.return_value = mock_response

        # Test max_results < 1 clamped to 1
        res = server.search_web("test", max_results=0)
        self.assertEqual(res["count"], 0)

        # Test max_results > 50 clamped to 50
        res = server.search_web("test", max_results=100)
        self.assertEqual(res["count"], 0)

    @patch("searxng_gateway.server._requests.get")
    def test_search_web_failure_degraded(self, mock_get):
        mock_get.side_effect = requests.RequestException("Connection refused")

        res = server.search_web("error test")
        self.assertEqual(res["query"], "error test")
        self.assertEqual(res["count"], 0)
        self.assertEqual(res["results"], [])
        self.assertTrue(res["degraded"])
        self.assertIn("Connection refused", res["error"])
        self.assertEqual(res["latency_ms"], 0)

    @patch("searxng_gateway.server._requests.get")
    def test_searxng_health_ok(self, mock_get):
        health_resp = MagicMock()
        health_resp.status_code = 200
        health_resp.json.return_value = {
            "results": [{"title": "test"}],
            "unresponsive_engines": [["bing", "timeout"]],
        }

        config_resp = MagicMock()
        config_resp.status_code = 200
        config_resp.json.return_value = {"version": "searxng-2026.1"}

        mock_get.side_effect = [health_resp, config_resp]

        health = server.searxng_health()
        self.assertEqual(health["status"], "ok")
        self.assertTrue(health["reachable"])
        self.assertEqual(health["status_code"], 200)
        self.assertEqual(health["result_count"], 1)
        self.assertEqual(health["unresponsive_engines"], [["bing", "timeout"]])
        self.assertEqual(health["version"], "searxng-2026.1")

    @patch("searxng_gateway.server._requests.get")
    def test_searxng_health_degraded(self, mock_get):
        health_resp = MagicMock()
        health_resp.status_code = 503
        mock_get.return_value = health_resp

        health = server.searxng_health()
        self.assertEqual(health["status"], "degraded")
        self.assertFalse(health["reachable"])
        self.assertEqual(health["status_code"], 503)

    @patch("searxng_gateway.server._requests.get")
    def test_searxng_health_down(self, mock_get):
        mock_get.side_effect = requests.ConnectionError("Connection failed")

        health = server.searxng_health()
        self.assertEqual(health["status"], "down")
        self.assertFalse(health["reachable"])
        self.assertIn("Connection failed", health["error"])

    @patch("os.path.exists")
    def test_deep_research_orchestrator_missing(self, mock_exists):
        mock_exists.return_value = False
        res = server.deep_research("research topic", count=5)
        self.assertEqual(res["query"], "research topic")
        self.assertIsNone(res["answer"])
        self.assertTrue(res["degraded"])
        self.assertIn("DEEP_RESEARCH_ORCHESTRATOR not configured or missing", res["error"])

    @patch("searxng_gateway.server.subprocess.run")
    @patch("os.path.exists")
    def test_deep_research_orchestrator_success(self, mock_exists, mock_run):
        mock_exists.return_value = True
        mock_proc = MagicMock()
        mock_proc.returncode = 0
        mock_proc.stdout = "Comprehensive research synthesis on AI agents."
        mock_proc.stderr = ""
        mock_run.return_value = mock_proc

        res = server.deep_research("AI agents", count=10)
        self.assertEqual(res["query"], "AI agents")
        self.assertEqual(res["answer"], "Comprehensive research synthesis on AI agents.")
        self.assertFalse(res["degraded"])
        mock_run.assert_called_once_with(
            [config.DEEP_RESEARCH_ORCHESTRATOR, "AI agents", "deep_research", "10"],
            capture_output=True,
            text=True,
            timeout=config.DEEP_RESEARCH_TIMEOUT,
        )

    @patch("searxng_gateway.server.subprocess.run")
    @patch("os.path.exists")
    def test_deep_research_orchestrator_nonzero_exit(self, mock_exists, mock_run):
        mock_exists.return_value = True
        mock_proc = MagicMock()
        mock_proc.returncode = 1
        mock_proc.stdout = ""
        mock_proc.stderr = "Orchestrator partial failure warning"
        mock_run.return_value = mock_proc

        res = server.deep_research("AI agents", count=5)
        self.assertEqual(res["answer"], "Orchestrator partial failure warning")
        self.assertTrue(res["degraded"])

    @patch("searxng_gateway.server.subprocess.run")
    @patch("os.path.exists")
    def test_deep_research_orchestrator_timeout(self, mock_exists, mock_run):
        mock_exists.return_value = True
        mock_run.side_effect = subprocess.TimeoutExpired(cmd="orchestrator", timeout=60)

        res = server.deep_research("AI agents", count=5)
        self.assertTrue(res["degraded"])
        self.assertIn("TimeoutExpired", res["error"])

    @patch("searxng_gateway.server._mg_hybrid_search")
    @patch("searxng_gateway.server.subprocess.run")
    @patch("os.path.exists")
    def test_deep_research_with_semantic_memory(self, mock_exists, mock_run, mock_mg):
        mock_exists.return_value = True
        mock_proc = MagicMock()
        mock_proc.returncode = 0
        mock_proc.stdout = "Research findings"
        mock_proc.stderr = ""
        mock_run.return_value = mock_proc

        mock_mg.return_value = {
            "query": "topic",
            "count": 1,
            "results": [{"title": "Lab notes", "score": 0.95}],
            "degraded": False,
        }

        with patch.object(config, "SEMANTIC_ENABLED", True), \
             patch.object(server, "_MG_AVAILABLE", True), \
             patch.object(server, "_mg_hybrid_search", mock_mg):
            res = server.deep_research("topic", count=5)
            self.assertEqual(res["answer"], "Research findings")
            self.assertIsNotNone(res["semantic_memory"])
            self.assertEqual(res["semantic_memory"]["count"], 1)
            self.assertFalse(res["degraded"])

    @patch.object(server.mcp, "run")
    def test_main_entrypoint(self, mock_mcp_run):
        server.main()
        mock_mcp_run.assert_called_once_with(transport="stdio")


if __name__ == "__main__":
    unittest.main()
