# searxng-mcp-gateway

[![Python 3.10+](https://img.shields.io/badge/python-3.10%2B-blue)](https://www.python.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![MCP Server](https://img.shields.io/badge/MCP--Server-available-green)](https://modelcontextprotocol.io/)
[![SearXNG](https://img.shields.io/badge/SearXNG-Integrated-orange)]

## About

**searxng-mcp-gateway** is an MCP (Model Context Protocol) server designed for SearXNG metasearch engine integration. It enables AI agents to perform web searches across 90+ search engines, extract content from URLs, and combine web results with semantic memory through a unified interface. The gateway is designed to return raw data for agent processing, without adding any LLM-based text synthesis to the search results.

## Documentation

Comprehensive documentation for developers and operators is located in the `/docs` directory:

- 📊 **[Architecture](docs/architecture.md):** High-level architecture, module logic, and Mermaid.js data flow diagrams.
- 🔀 **[Data Flow & API Reference](docs/data-flow.md):** MCP tools documentation, input parameters, response schemas, and data sanitization logic.
- 🚀 **[Deployment](docs/deployment.md):** Installation guide, step-by-step setup instructions, and environment variable configuration table.

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