---
module_type: gateway
status: active
protocol: mcp
primary_capability: search
requires: searxng
works_with: mcp-clients
last_verified: 2026-08-21
---

# searxng-mcp-gateway

**Provides read-only web search and deep research orchestration capabilities via SearXNG.**

## Status
Active - Last verified: 2026-08-21

## What it does / does not do
**Does:**
- Exposes web search capabilities to MCP clients via a local SearXNG instance.
- Diagnoses connectivity and availability of the underlying SearXNG engine and external paid APIs.
- Orchestrates deep research and semantic memory retrieval.
- Performs Quota-Aware Hybrid Search (RRF) cascading through external APIs (Tavily, Exa) and local SearXNG.

**Does not:**
- Mutate SearXNG settings.
- Enable or disable search engines.

## Why an agent would use it
Agents can use this gateway to perform web searches, retrieve context for complex queries, and orchestrate deep research workflows leveraging semantic memory without mutating the underlying search engine state. The new hybrid search provides the highest quality results by gracefully load-balancing paid APIs and falling back to free engines.

## Architecture and dependencies
- Python >= 3.10
- Dependencies: `mcp>=1.0.0`, `requests>=2.28.0`
- Communicates with a SearXNG instance over HTTP.
- Communicates with external Echelon APIs (Tavily, Exa, Firecrawl, Olostep) via Vault secrets.

## Compatibility
Works with standard MCP clients and expects SearXNG.

## Quick start and health check
Start the gateway:
```bash
python -m searxng_mcp_gateway
```
Health check:
```bash
python -m pytest tests/test_health.py
```

## Configuration and environment variables
- `SEARXNG_URL`: The URL of the SearXNG instance (default: `http://127.0.0.1:8081`).
- `HOST`: MCP server host (default: `127.0.0.1`).
- `PORT`: MCP server port (default: `8092`).

## Complete MCP Tool/API table with side effects
| Tool | Description | Side Effects |
|------|-------------|--------------|
| `search_web` | Basic local web search via SearXNG | None |
| `deep_research` | Heavy search orchestrator | None |
| `hybrid_search` | Parallel search (SearXNG + Paid APIs) with Quota-Aware Fallback and RRF fusion | None (consumes API quotas) |
| `ecosystem_health` | Deep diagnostics for SearXNG and commercial API key circuit breakers | None |

## Security model and trust boundaries
- Needs access to a SearXNG instance without authentication. Do not expose SearXNG to the public internet directly.

## Tests and exact commands
```bash
pytest tests/
```

## Operations, logs, backup/restore, rollback
- Stateless. No backups needed.

## Generic MCP-client example
```json
{
  "mcpServers": {
    "searxng": {
      "command": "python",
      "args": ["-m", "searxng_mcp_gateway"],
      "env": {
        "SEARXNG_URL": "http://127.0.0.1:8081"
      }
    }
  }
}
```

## Related TheNovaNodes modules
- searxng-mcp-control

## License
MIT
