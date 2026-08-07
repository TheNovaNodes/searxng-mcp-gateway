"""Unit tests for searxng-mcp-gateway configuration."""

import importlib
import os
import unittest
from unittest.mock import patch

from searxng_gateway import config


class TestConfig(unittest.TestCase):
    """Test configuration defaults and environment overrides."""

    def test_default_config_values(self):
        with patch.dict(os.environ, {}, clear=True):
            importlib.reload(config)
            self.assertEqual(config.SEARXNG_URL, "http://127.0.0.1:8081")
            self.assertEqual(config.DEFAULT_MAX_RESULTS, 10)
            self.assertEqual(config.DEFAULT_LANGUAGE, "auto")
            self.assertEqual(config.DEFAULT_SAFESEARCH, 0)
            self.assertEqual(config.DEFAULT_TIMEOUT, 10)
            self.assertEqual(config.HOST, "127.0.0.1")
            self.assertEqual(config.PORT, 8092)
            self.assertEqual(config.DEEP_RESEARCH_TIMEOUT, 240)
            self.assertFalse(config.SEMANTIC_ENABLED)
            self.assertEqual(config.SEMANTIC_TOP_K, 5)
            self.assertTrue(config.SEMANTIC_EXPAND)
            self.assertEqual(config.SEMANTIC_FUSION, "weighted")

    def test_env_overrides(self):
        env_vars = {
            "SEARXNG_URL": "http://192.168.1.100:8081",
            "SEARXNG_DEFAULT_MAX": "25",
            "SEARXNG_DEFAULT_LANG": "en",
            "SEARXNG_SAFESEARCH": "1",
            "SEARXNG_TIMEOUT": "20",
            "SEARXNG_HOST": "0.0.0.0",
            "SEARXNG_PORT": "9090",
            "DEEP_RESEARCH_ORCHESTRATOR": "/custom/orchestrator.sh",
            "DEEP_RESEARCH_TIMEOUT": "120",
            "SEMANTIC_ENABLED": "1",
            "SEMANTIC_TOP_K": "8",
            "SEMANTIC_EXPAND": "0",
            "SEMANTIC_FUSION": "rrf",
        }
        with patch.dict(os.environ, env_vars, clear=True):
            importlib.reload(config)
            self.assertEqual(config.SEARXNG_URL, "http://192.168.1.100:8081")
            self.assertEqual(config.DEFAULT_MAX_RESULTS, 25)
            self.assertEqual(config.DEFAULT_LANGUAGE, "en")
            self.assertEqual(config.DEFAULT_SAFESEARCH, 1)
            self.assertEqual(config.DEFAULT_TIMEOUT, 20)
            self.assertEqual(config.HOST, "0.0.0.0")
            self.assertEqual(config.PORT, 9090)
            self.assertEqual(config.DEEP_RESEARCH_ORCHESTRATOR, "/custom/orchestrator.sh")
            self.assertEqual(config.DEEP_RESEARCH_TIMEOUT, 120)
            self.assertTrue(config.SEMANTIC_ENABLED)
            self.assertEqual(config.SEMANTIC_TOP_K, 8)
            self.assertFalse(config.SEMANTIC_EXPAND)
            self.assertEqual(config.SEMANTIC_FUSION, "rrf")

        # Reload back to clean state
        importlib.reload(config)


if __name__ == "__main__":
    unittest.main()
