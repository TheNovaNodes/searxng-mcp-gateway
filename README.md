---
module_type: mcp-server
status: active
protocol: stdio
primary_capability: Web search, scraping, and deep research
requires: searxng
works_with: mcp-clients, Claude Desktop, Cursor, AI agents
last_verified: 2026-09-11
---

# SearXNG MCP Gateway (Go)
*High-performance, privacy-focused Model Context Protocol (MCP) gateway for autonomous web search, scraping, and deep research.*

[![CI](https://github.com/TheNovaNodes/searxng-mcp-gateway/actions/workflows/ci.yml/badge.svg)](https://github.com/TheNovaNodes/searxng-mcp-gateway/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.22%20--%201.25-00ADD8?logo=go)](https://golang.org)
[![Protocol: MCP](https://img.shields.io/badge/protocol-MCP%20JSON--RPC-green.svg)](https://modelcontextprotocol.io/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/TheNovaNodes/searxng-mcp-gateway?logo=github)](https://github.com/TheNovaNodes/searxng-mcp-gateway/releases)

📚 **Documentation Suite:** [Architecture](ARCHITECTURE.md) • [Agent Directives](AGENTS.md) • [Contributing](CONTRIBUTING.md) • [Security Policy](SECURITY.md) • [Code of Conduct](CODE_OF_CONDUCT.md) • [License](LICENSE)

---

## Overview

**SearXNG MCP Gateway** is a production-grade Go server implementing the Model Context Protocol (MCP) over `stdio`. Engineered for autonomous AI agent ecosystems (such as Google Antigravity, Claude Code, Cursor, and Windsurf), it provides agents with high-speed, privacy-first search, resilient web scraping, and multi-engine deep research capabilities.

### Key Capabilities
- 🚀 **Three Canonical Tools:** Zero tool bloat. Exposes exactly 3 focused tools (`search_web`, `fetch_page`, `deep_research`).
- 🔒 **Enterprise-Grade SSRF Protection:** Active blocking of IPv4/IPv6 loopback (`127.0.0.0/8`, `::1`), RFC 1918 private subnets, carrier-grade NAT, and cloud instance metadata services (`169.254.169.254`), with strict redirect validation.
- 🛡️ **Zero-Quota Local First:** Primary web searches execute against local SearXNG instances without incurring paid API fees or rate limits.
- 🌊 **Resilient Scraping Cascade (Echelon):** Web page extraction with automated fallback through Firecrawl API, Olostep API, and native HTTP client with User-Agent rotation and Readability extraction.
- 🧬 **Reciprocal Rank Fusion (RRF, $k=60$):** Multi-source hybrid search combining local SearXNG results with AI discovery engines (Exa AI, Tavily) without hallucinated synthesis.
- ⚡ **Stream & Transport Isolation:** Strict `os.Stdout` ownership by MCP JSON-RPC. All diagnostics and logs route to `os.Stderr` via `log/slog`.

---

## What It Does / Does Not Do

### What It Does
- Exposes a standard MCP `stdio` transport interface compatible with any MCP-compliant client.
- Performs fast local searches across Google, Bing, DuckDuckGo, Wikipedia, and specialized search engines configured in SearXNG.
- Scrapes web pages into clean Markdown, enforcing a strict 35,000-character token guardrail to protect LLM context windows.
- Orchestrates deep research runs concurrently using `golang.org/x/sync/errgroup` with bounded timeouts.
- Dynamically loads provider credentials from RAM-disk vaults (`/dev/shm/agent_vault`) with automatic fallback to standard environment variables.
- Implements thread-safe Circuit Breaker protection against downstream API rate limits and network outages.

### What It Does Not Do
- Does not bundle or run the SearXNG search engine itself (connects via HTTP to an existing SearXNG service).
- Does not synthesize or summarize search results via LLMs (returns raw, objective, structured data to prevent double-generation latency and hallucinations).
- Does not allow access to local network interfaces, loopback IPs, or cloud metadata endpoints.
- Does not leak secrets into logs, stdout, or error payloads.

---

## Quick Start

### Option A: Install via `go install` (Recommended)

```bash
go install github.com/TheNovaNodes/searxng-mcp-gateway@latest
```

### Option B: Build from Source

**Prerequisites:** Go 1.22+ (Go 1.25 recommended).

```bash
# 1. Clone repository
git clone https://github.com/TheNovaNodes/searxng-mcp-gateway.git
cd searxng-mcp-gateway

# 2. Build executable
make build
# Produces a compact static binary in bin/searxng-gateway (~7.7 MB)

# 3. Run test suite
make test
```

### Option C: Docker & Docker Compose

Run SearXNG and the Gateway together:

```bash
docker compose up -d
```

---

## MCP Client Configuration

### Claude Desktop (`claude_desktop_config.json`)

On macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`  
On Windows: `%APPDATA%\Claude\claude_desktop_config.json`  
On Linux: `~/.config/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "searxng-gateway": {
      "command": "/usr/local/bin/searxng-gateway",
      "env": {
        "SEARXNG_URL": "http://127.0.0.1:8889",
        "SEARXNG_DEFAULT_MAX": "10",
        "EXA_API_KEY": "your-exa-key-optional",
        "TAVILY_API_KEY": "your-tavily-key-optional"
      }
    }
  }
}
```

### Cursor (`~/.cursor/mcp.json` or Project `.cursor/mcp.json`)

```json
{
  "mcpServers": {
    "searxng-gateway": {
      "command": "searxng-gateway",
      "env": {
        "SEARXNG_URL": "http://localhost:8889"
      }
    }
  }
}
```

### Ecosystem Router (`mcp-router.yaml`)

```yaml
servers:
  nova-searxng-gateway:
    transport: stdio
    command: /usr/local/bin/searxng-gateway
    env:
      SEARXNG_URL: http://127.0.0.1:8889
      SEARXNG_DEFAULT_MAX: "10"
      CASCADE_TIMEOUT: "6"
    prefix: nova-searxng-gateway__
```

---

## Available MCP Tools

### 1. `search_web`
Fast, privacy-focused search against your SearXNG instance.

| Parameter | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `query` | string | **Yes** | — | Search query string |
| `max_results` | number | No | `10` | Number of results to return (1–50) |
| `categories` | string | No | `general` | SearXNG categories (`general`, `news`, `it`, `science`, `files`) |
| `engines` | string | No | — | Comma-separated list of engines (e.g. `google,duckduckgo`) |
| `language` | string | No | `auto` | Results language code (e.g. `ru`, `en`, `de`) |
| `safesearch` | number | No | `0` | SafeSearch filter (0 = off, 1 = moderate, 2 = strict) |

**Sample Output:**
```json
{
  "query": "Model Context Protocol Go SDK",
  "count": 2,
  "results": [
    {
      "title": "mark3labs/mcp-go: A Go implementation of the Model Context Protocol",
      "url": "https://github.com/mark3labs/mcp-go",
      "content": "Go SDK for building Model Context Protocol servers and clients...",
      "engine": "github",
      "engines": ["github"],
      "score": 1.0,
      "category": "it"
    }
  ],
  "latency_ms": 342.5
}
```

### 2. `fetch_page`
Extract clean, readable Markdown from any public URL with anti-bot bypass and SSRF protection.

| Parameter | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `url` | string | **Yes** | — | Fully-qualified public HTTP/HTTPS URL |

- **Security:** Evaluates URL schemes, hostname DNS resolution, and destination IP ranges. Rejects `127.0.0.1`, `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`, `169.254.169.254`, IPv6 loopback/link-local, and non-HTTP schemes.
- **Guardrail:** Output is truncated at 35,000 characters to safeguard model context.

### 3. `deep_research`
Multi-source research orchestration combining local SearXNG results with specialized deep search providers (Exa AI or Tavily).

| Parameter | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `query` | string | **Yes** | — | Research query string |
| `max_results` | number | No | `10` | Final ranked results count (1–50) |
| `language` | string | No | `auto` | Language code |

- **Ranking:** Merged using Reciprocal Rank Fusion ($Score = \sum \frac{1}{k + rank_i}$, with $k=60$).

---

## Configuration Reference

All settings can be configured via environment variables:

| Variable | Default | Description |
| :--- | :--- | :--- |
| `SEARXNG_URL` | `http://127.0.0.1:8889` | Target SearXNG instance endpoint |
| `SEARXNG_DEFAULT_MAX` | `10` | Default results count for queries |
| `SEARXNG_MAX_ALLOWED` | `50` | Maximum limit for search results |
| `SEARXNG_DEFAULT_LANG` | `auto` | Default search language |
| `SEARXNG_SAFESEARCH` | `0` | Default content filter level |
| `SEARXNG_TIMEOUT` | `10` | Timeout for SearXNG HTTP requests (seconds) |
| `CASCADE_TIMEOUT` | `6` | Timeout for external scraping cascade tiers (seconds) |
| `RRF_K` | `60` | Reciprocal Rank Fusion ranking factor |
| `AGENT_VAULT_DIR` | `/dev/shm/agent_vault` | Path to RAM-disk secret vault |
| `EXA_API_KEY` | — | Fallback API key for Exa AI deep search |
| `TAVILY_API_KEY` | — | Fallback API key for Tavily search |
| `FIRECRAWL_API_KEY` | — | Fallback API key for Firecrawl scraper |
| `OLOSTEP_API_KEY` | — | Fallback API key for Olostep scraper |

---

## Security Architecture

1. **SSRF Hardening:** See [SECURITY.md](SECURITY.md) for full threat model and test matrices. All outbound requests are checked before dispatch and re-checked across HTTP redirects.
2. **Dynamic Credential Handling:** Keys are read on-demand without persistent caching in static structures.
3. **Stream Isolation:** Zero `stdout` pollution ensures zero JSON-RPC framing errors in high-throughput agent swarms.

---

## Contributing

We welcome community contributions! Please review our [Contributing Guide](CONTRIBUTING.md) and [Code of Conduct](CODE_OF_CONDUCT.md) before submitting pull requests.

---

## License

This project is licensed under the [MIT License](LICENSE).
