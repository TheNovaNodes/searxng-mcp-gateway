# Data Flow and API Reference

This document details the MCP tool endpoints exposed by the `searxng-mcp-gateway` and outlines the data sanitization and transformation processes.

## MCP Tools Reference

### `search_web`
Performs a standard web search by delegating the query to the underlying SearXNG instance.

**Input Parameters:**

| Parameter | Type | Required | Default | Description |
|---|---|---|---|---|
| `query` | string | Yes | - | Natural language search query. |
| `max_results` | int | No | 10 | Maximum number of results to return (capped at 50). |
| `categories` | string | No | `None` | Search category (e.g., general, images, news, science). |
| `language` | string | No | `auto` | Language for the results. |
| `safesearch` | int | No | 0 | Safesearch level (0=off, 1=moderate, 2=strict). |
| `engines` | string | No | `None` | Comma-separated list of search engines to use. |

**Expected Return Object:**
A JSON dictionary with the following schema:
```json
{
  "query": "string",
  "count": "int",
  "results": [
    {
      "title": "string",
      "url": "string",
      "content": "string",
      "engine": "string",
      "engines": ["string"],
      "score": "float",
      "category": "string",
      "published_date": "string (optional)"
    }
  ],
  "answers": [],
  "infoboxes": [],
  "suggestions": [],
  "unresponsive_engines": [],
  "latency_ms": "float"
}
```

**Error Handling / Fallback Return:**
On exception, it returns a degraded response payload:
```json
{
  "query": "string",
  "count": 0,
  "results": [],
  "degraded": true,
  "error": "ExceptionName: error message",
  "latency_ms": 0
}
```

---

### `searxng_health`
Diagnostic endpoint to check the connectivity and health status of the configured SearXNG service.

**Input Parameters:** None.

**Expected Return Object:**
```json
{
  "status": "ok | degraded | down | unknown",
  "searxng_url": "string",
  "reachable": "boolean",
  "latency_ms": "float",
  "status_code": "int",
  "result_count": "int (optional)",
  "unresponsive_engines": ["string"],
  "version": "string (optional)",
  "error": "string (optional)"
}
```

---

### `deep_research`
Orchestrates deep web research using an external script combined with a local semantic memory retrieval.

**Input Parameters:**

| Parameter | Type | Required | Default | Description |
|---|---|---|---|---|
| `query` | string | Yes | - | The complex research query. |
| `count` | int | No | 10 | Number of results to fetch per provider on the web layer. |

**Expected Return Object:**
```json
{
  "query": "string",
  "answer": "string (Web-synthesized content)",
  "semantic_memory": {
    "query": "string",
    "count": "int",
    "results": [],
    "degraded": "boolean (optional)",
    "error": "string (optional)"
  },
  "degraded": "boolean",
  "error": "string (optional)"
}
```

## Data Transformation and Sanitization

### Source and Extraction
- Requests are dispatched to the downstream SearXNG URL (set via `SEARXNG_URL`).
- Responses are received as raw JSON.

### Sanitization Process
In `server.py` (`_searxng_search`):
- `max_results` is constrained between 1 and 50.
- The raw `results` list returned by SearXNG is truncated using array slicing: `[:max_results]`.
- An iteration loop maps specific keys from the raw JSON (`title`, `url`, `content`, `engine`, `engines`, `score`, `category`, `publishedDate`) into the final structured output items. Unnecessary, bloated, or potentially unsafe metadata returned by SearXNG engines is discarded.
- Fallback structures guarantee that standard expected keys always exist (even as empty strings or lists) to ensure stability in MCP clients.

### Deep Research
- Operates via a subprocess calling a shell or python script. Output is captured as plain text.
- Graceful degradation: The process merges data from the web (subprocess) and memory (local API call). If either layer throws an exception or errors out, its sub-payload sets `degraded: true` and includes the error message, preventing the main server routine from crashing. No LLM-based text synthesis is injected directly by this layer.