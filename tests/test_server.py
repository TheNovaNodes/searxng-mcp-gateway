import unittest
from unittest.mock import patch, MagicMock
from searxng_gateway import server
from searxng_gateway import echelon

class TestServer(unittest.TestCase):
    
    @patch("searxng_gateway.server.echelon.call_tavily")
    def test_deep_research_auto(self, mock_tavily):
        mock_tavily.return_value = {"answer": "Tavily summary"}
        res = server.deep_research("test query", mode="auto")
        self.assertEqual(res["provider"], "tavily")
        self.assertFalse(res["degraded"])
        self.assertEqual(res["answer"]["answer"], "Tavily summary")

    @patch("searxng_gateway.server.echelon.call_exa")
    def test_deep_research_semantic(self, mock_exa):
        mock_exa.return_value = {"results": ["Exa result"]}
        res = server.deep_research("test query", mode="semantic")
        self.assertEqual(res["provider"], "exa")
        self.assertFalse(res["degraded"])

    @patch("searxng_gateway.server.echelon.call_firecrawl_scrape")
    def test_deep_research_scrape(self, mock_fc):
        mock_fc.return_value = {"markdown": "Hello"}
        res = server.deep_research("http://example.com", mode="auto")
        self.assertEqual(res["provider"], "firecrawl")
        self.assertFalse(res["degraded"])

    @patch("searxng_gateway.server.echelon.call_firecrawl_scrape")
    @patch("searxng_gateway.server.echelon.call_olostep_scrape")
    def test_deep_research_waf_fallback(self, mock_olo, mock_fc):
        mock_fc.return_value = {"error": "Blocked", "status_code": 403}
        mock_olo.return_value = {"markdown": "Bypassed Cloudflare"}
        res = server.deep_research("http://cloudflare.com")
        self.assertEqual(res["provider"], "firecrawl -> olostep (WAF Bypassed)")
        self.assertEqual(res["answer"]["markdown"], "Bypassed Cloudflare")
        self.assertFalse(res["degraded"])

if __name__ == "__main__":
    unittest.main()
