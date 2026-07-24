# nova-searxng-mcp

[![Python 3.10+](https://img.shields.io/badge/python-3.10%2B-blue)](https://www.python.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![MCP Server](https://img.shields.io/badge/MCP--Server-available-green)](https://modelcontextprotocol.io/)
[![SearXNG](https://img.shields.io/badge/SearXNG-Integrated-orange)]

## About

MCP (Model Context Protocol) server for SearXNG metasearch engine integration. Enables AI agents to perform web searches across 90+ search engines, extract content from URLs, and combine web results with semantic memory through unified interface.

## Features

- **Web Search**: Metasearch across 90+ engines (Google, Bing, DuckDuckGo, etc.)
- **Deep Research**: Multi-step research with source analysis and synthesis
- **Content Extraction**: Clean HTML to markdown from JavaScript-rendered pages
- **Category Filtering**: Search in general, images, news, videos, science modes
- **Language Control**: Multi-language search support with safesearch options
- **7 Tools Exposed**:
  - `search_web` — Standard web search with engine selection
  - `deep_research` — Research orchestration + semantic memory fusion
  - `searxng_health` — Engine diagnostics and connectivity
  - `extract_content` — URL content extraction

## Installation

```bash
# Clone repository
git clone https://github.com/TheNovaNodes/nova-searxng-mcp.git
cd nova-searxng-mcp

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
| `SEARXNG_URL` | SearXNG instance URL | `http://127.0.0.1:8888` |
| `SEARXNG_TIMEOUT` | Request timeout (seconds) | 30 |
| `SEARXNG_MAX_RESULTS` | Max search results | 10 |
| `DEEP_RESEARCH_ORCHESTRATOR` | Orchestrator script path | null |

## MCP Client Integration

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "nova-searxng": {
      "command": "python",
      "args": ["server.py"],
      "cwd": "/path/to/nova-searxng-mcp",
      "env": {
        "SEARXNG_URL": "http://127.0.0.1:8888",
        "PYTHONPATH": "/path/to/nova-searxng-mcp"
      }
    }
  }
}
```

## 💖 Support TheNovaNodes

Если наши MCP-шлюзы экономят вам время и расширяют возможности ваших AI-агентов, вы можете поддержать нашу лабораторию. Все средства идут на поддержку инфраструктуры и развитие новых open-source интеграций.
**USDT (Сеть TRC20): TQvw8MJMdSBFXu5G74JsZm1gzg7cuXBZ2o**

<details>
<summary><b>Показать QR-код для перевода (Bybit)</b></summary>
<br>
<img src=".github/assets/qr-usdt-trc20.png" alt="USDT TRC20 QR Code Bybit" width="300">
</details>

## License

MIT — See LICENSE file.
