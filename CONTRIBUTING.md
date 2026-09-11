# Contributing to SearXNG MCP Gateway

First off, thank you for considering contributing to the **SearXNG MCP Gateway**! This project is part of TheNovaNodes Autonomous AI Agent Ecosystem.

## Code of Conduct
By participating in this project, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md). Please be respectful and professional in all communications, issues, and code reviews.

## How Can I Contribute?

### Reporting Bugs & Vulnerabilities
- For security vulnerabilities, **do not** file a public issue. Follow our [Security Policy](SECURITY.md).
- For non-security bugs, open an issue using a clear descriptive title, exact reproduction steps, environment details (OS, Go version, SearXNG version), and relevant MCP client configuration snippets.

### Feature Suggestions
- Open an issue describing the proposed feature, intended agent workflow, and why existing tools (`search_web`, `fetch_page`, `deep_research`) are insufficient.
- We maintain a strict minimalist philosophy: tool count is kept intentionally bounded to prevent context window bloating in AI agent swarms.

### Pull Requests
1. Create a dedicated branch from `main` (e.g. `feat/your-feature`, `fix/your-bug`, `docs/your-doc`).
2. Follow Conventional Commits format: `feat:`, `fix:`, `refactor:`, `docs:`, `ci:`, `chore:`.
3. Never include secrets, tokens, or personal paths in commits or test fixtures.
4. Ensure all unit and race tests pass before opening a PR.

---

## Development Standards & Guidelines

### The Golden Loop (Required Before Every PR)
Every contribution must pass the Golden Loop locally:

```bash
# 1. Format code
gofmt -s -w .

# 2. Static analysis
go vet ./...

# 3. Unit tests with race detection
go test -v -race ./...

# 4. Binary build verification
make build
```

### Stdio Transport Hygiene
- The standard output stream (`os.Stdout`) is **strictly reserved** for the MCP JSON-RPC protocol transport (`github.com/mark3labs/mcp-go`).
- **Never** print diagnostic text, debug messages, or stack traces to `os.Stdout` (e.g., avoid `fmt.Println` or `println`).
- All logging must be routed to `os.Stderr` using Go's standard structured logger (`log/slog`).

### SSRF Defense & Network Security
- Any code performing outbound HTTP requests (such as scraping or link retrieval) **must** validate destination URLs using `internal/echelon.ValidateTargetURL(targetURL)`.
- HTTP clients must implement `CheckRedirect` functions that re-validate subsequent redirect targets to prevent DNS rebinding or redirect-based SSRF attacks.
- Outbound requests must never touch RFC 1918 private subnets, loopback addresses (`127.0.0.0/8`, `::1`), or cloud metadata endpoints (`169.254.169.254`).

### Credential Hygiene
- Dynamic secret loading must support both RAM-disk vaults (`/dev/shm/agent_vault`) and standard environment variables (`internal/vault`).
- Zero hardcoded keys or passwords in source code, documentation, or commit history.

---

## Local Development Setup

1. Clone your fork:
   ```bash
   git clone https://github.com/TheNovaNodes/searxng-mcp-gateway.git
   cd searxng-mcp-gateway
   ```

2. Copy environment template:
   ```bash
   cp .env.example .env
   ```

3. Start a local SearXNG instance (or use the provided `docker-compose.yml`):
   ```bash
   docker compose up -d searxng
   ```

4. Build and test:
   ```bash
   make test
   make build
   ```

Thank you for helping build resilient infrastructure for autonomous AI agents!
