# searxng-mcp-gateway

[![Python 3.10+](https://img.shields.io/badge/python-3.10%2B-blue)](https://www.python.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![MCP Server](https://img.shields.io/badge/MCP--Server-available-green)](https://modelcontextprotocol.io/)
[![SearXNG](https://img.shields.io/badge/SearXNG-Integrated-orange)]

## About

**searxng-mcp-gateway** is an MCP (Model Context Protocol) server for SearXNG metasearch engine integration, functioning as a core component of the **Antigravity Agent Ecosystem** managed by **DoctorM & Ai / TheNovaNodes**. It enables AI agents to perform web searches across 90+ search engines, extract content from URLs, and seamlessly combine web results with semantic memory through a unified interface.

## Ecosystem Role

As part of the **Antigravity Agent Ecosystem**, this gateway serves as the primary external knowledge retrieval mechanism for our autonomous agents. It is designed to return raw data without LLM-based synthesis, ensuring agents receive clean, unopinionated information to process in their own context.

## Features

- **Web Search**: Metasearch across 90+ engines (Google, Bing, DuckDuckGo, etc.)
- **Deep Research**: Multi-step research with source analysis and synthesis
- **Content Extraction**: Clean HTML to markdown from JavaScript-rendered pages
- **Category Filtering**: Search in general, images, news, videos, science modes
- **Language Control**: Multi-language search support with safesearch options
- **3 Tools Exposed**:
  - `search_web` — Standard web search with engine selection
  - `searxng_health` — Engine diagnostics and connectivity
  - `deep_research` — Research orchestration + semantic memory fusion (experimental)

## Installation

```bash
# Clone repository
git clone https://github.com/TheNovaNodes/searxng-mcp-gateway.git
cd searxng-mcp-gateway

# Create virtual environment
python3 -m venv venv
source venv/bin/activate

# Install dependencies
pip install -r requirements.txt
```

## Configuration

Creates `.env` from `.env.example`:

| Variable | Description | Default |
|----------|-------------|---------|
| `SEARXNG_URL` | SearXNG instance URL | `http://127.0.0.1:8081` |
| `SEARXNG_TIMEOUT` | Request timeout (seconds) | 30 |
| `SEARXNG_MAX_RESULTS` | Max search results | 10 |
| `DEEP_RESEARCH_ORCHESTRATOR` | Orchestrator script path | `(none)` - requires separate setup |
| `SEMANTIC_ENABLED` | Enable semantic memory fusion | `0` (off by default) |

## MCP Client Integration

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "searxng-gateway": {
      "command": "python",
      "args": ["-m", "searxng_gateway.server"],
      "cwd": "/path/to/searxng-mcp-gateway",
      "env": {
        "SEARXNG_URL": "http://127.0.0.1:8081"
      }
    }
  }
}
```

## 💖 Support TheNovaNodes

If our MCP gateways save you time and expand your AI agents' capabilities, consider supporting our infrastructure and the development of new open-source integrations.

## License

MIT — See LICENSE file.
