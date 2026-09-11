# TheNovaNodes AGENTS.md Sandwich Manifest

📚 **Documentation Suite:** [README](README.md) • [Architecture](ARCHITECTURE.md) • [Contributing](CONTRIBUTING.md) • [Security Policy](SECURITY.md) • [Code of Conduct](CODE_OF_CONDUCT.md) • [License](LICENSE)

## Part 1: NovaNodes Universal Invariants

1.  **Strict Git Flow (ПРАВИЛА КРОВИ):**
    *   NEVER push directly to `main` or `master`.
    *   All changes MUST be proposed via dedicated feature branches and PRs.
    *   No merge without ЗавЛаб approval.
2.  **Zero Hardcoded Credentials:**
    *   API keys for paid providers (Tavily, Exa, Firecrawl, Olostep, etc.) MUST be dynamically loaded from a shared memory RAM-disk vault located at `/dev/shm/agent_vault` or via environment variables.
    *   Zero plaintext secrets in tests, PRs, or configuration files.
3.  **Deadlock/Timeout Protections:**
    *   All network operations MUST have explicit timeouts.
    *   Concurrency MUST be managed carefully to avoid deadlocks.
4.  **Continuous Native Verification (The Golden Loop):**
    *   Zero test failures are allowed.
    *   All code changes must pass `go vet ./...` without issues.

## Part 2: Repository Specific Architecture & Profile

*   **Runtime Environment:** Go 1.25+
*   **Server Framework:** FastMCP stdio server (`github.com/mark3labs/mcp-go`)
*   **Canonical Tools:** Exactly 3 exposed tools:
    1.  `search_web`: Fast local search via SearXNG.
    2.  `fetch_page`: Web page retrieval with WAF/Cloudflare bypass.
    3.  `deep_research`: Hybrid research search with deduplication.
*   **Security & SSRF Gate:**
    *   All outbound requests in `fetch_page` MUST pass through `internal/echelon.ValidateTargetURL`.
    *   Blocks loopback (`127.0.0.0/8`, `::1`), RFC 1918 subnets, cloud metadata (`169.254.169.254`), and non-HTTP schemes.
    *   All HTTP redirects MUST re-validate the target URL.
*   **Echelon Scraping Cascade:** A resilient 3-tier cascade for fetching web pages:
    *   Tier 1: Firecrawl
    *   Tier 2: Olostep
    *   Tier 3: Native HTTP fallback
*   **Search Fusion (RRF):** Reciprocal Rank Fusion (RRF) with $k=60$ is used to merge and deduplicate search results in `deep_research`.
*   **Resilience:**
    *   **CircuitBreaker:** Protects against external service failures (timeouts, 401/403/429/5xx).
*   **Safety Constraints:**
    *   **Token Guardrail:** Enforces a 35,000 character maximum output ceiling on fetched pages (truncation) to protect LLM context windows.
    *   **No LLM Synthesis:** The gateway returns raw data; do not add LLM-based synthesis to the search results.

## Part 3: Golden Loop & Operational Standards

Before declaring task completion or submitting any changes, you MUST run The Golden Loop:

1.  `go vet ./...`
2.  `go test -v -race ./...`
3.  `make build`

Ensure 100% success for tests and zero vet issues.
