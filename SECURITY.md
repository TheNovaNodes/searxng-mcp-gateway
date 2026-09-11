# Security Policy

TheNovaNodes Collective treats autonomous AI agent infrastructure security, credential protection, and operational integrity with the highest priority. **searxng-mcp-gateway** acts as an outbound gateway for AI agents, executing web searches and page retrieval.

---

## 🔒 Supported Versions

| Version | Supported          | Runtime    | Status             |
| ------- | ------------------ | ---------- | ------------------ |
| 2.0.x   | :white_check_mark: | Go 1.22+ / 1.25 | Production Current |
| 1.0.x   | :x:                | Python legacy | Deprecated / EOL   |

---

## 🛡️ Security Architecture & Threat Model

### 1. SSRF Protection & Network Isolation
The gateway fetches content on behalf of LLMs and autonomous agents, making Server-Side Request Forgery (SSRF) a critical threat vector. We implement active defenses at multiple levels:

- **Strict Scheme Allowlist:** Only `http` and `https` protocols are permitted. Schemes such as `file://`, `gopher://`, `ftp://`, or `dict://` are rejected immediately.
- **Pre-Resolution & IP Filtering:** Destination hostnames are resolved before connection. Outbound traffic to the following IP spaces is categorically blocked:
  - IPv4 Loopback (`127.0.0.0/8`)
  - IPv6 Loopback (`::1`)
  - RFC 1918 Private Networks (`10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`)
  - Carrier-Grade NAT (`100.64.0.0/10`)
  - Cloud Instance Metadata Services (e.g., AWS/GCP/Azure link-local `169.254.169.254`)
  - IPv6 Link-Local (`fe80::/10`)
- **HTTP Redirect Re-Validation:** All HTTP redirects are intercepted via `CheckRedirect`. Each subsequent URL in a redirect chain is re-validated against the SSRF policy before the client follows it.

### 2. Dual-Mode Credential Hygiene
- **RAM-Disk Key Vault:** Primary credentials (API keys for Exa, Tavily, Firecrawl, Olostep) can be read dynamically from `/dev/shm/agent_vault`. Keys never persist on disk.
- **Environment Variable Fallback:** For Docker and containerized deployments, credentials may be provided via standard environment variables.
- **Ephemeral Key Lifecycle:** Keys are read per-request and are never stored in global variables or static state caches.
- **Zero-Leak Logging:** Sensitive tokens and authorization headers are never logged to `stdout`, `stderr`, or returned in MCP tool error messages.

### 3. Stdio Transport & IPC Stream Isolation
- **`os.Stdout` Protection:** The standard output stream is exclusively owned by the MCP JSON-RPC protocol transport (`mark3labs/mcp-go`).
- **Log Segregation:** All diagnostic messages, traces, and warnings are strictly directed to `os.Stderr` via `log/slog`. No plain-text logs or debug dumps are permitted on `stdout`, preventing protocol frame corruption and agent parser failures.

### 4. Context Window & Denial-of-Service Safeguards
- **Token Guardrail:** Outbound scraped pages are subject to a hard truncation ceiling of 35,000 characters. This prevents malicious web pages from exhausting LLM context windows or causing token-based Denial-of-Service attacks.
- **Circuit Breaker:** Outbound requests to third-party APIs are wrapped in a thread-safe Circuit Breaker that fails fast upon detecting repeated rate-limits (HTTP 429), timeouts, or upstream outages.

---

## 🚨 Reporting a Vulnerability

If you discover a security vulnerability in `searxng-mcp-gateway`:
1. **Do not** open a public GitHub issue.
2. Report the vulnerability privately using [GitHub Security Advisories](https://github.com/TheNovaNodes/searxng-mcp-gateway/security/advisories/new) or directly to the project maintainers.
3. Include detailed reproduction steps, proof of concept, and environment information.
4. Maintainers will acknowledge receipt within 48 hours and coordinate a patch and responsible disclosure timeline.
