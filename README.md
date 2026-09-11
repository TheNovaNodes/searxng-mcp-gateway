---
module_type: gateway
status: active
protocol: mcp
primary_capability: search
requires: searxng
works_with: mcp-clients
last_verified: 2026-09-08
---

# searxng-mcp-gateway

**Высокопроизводительный Go MCP шлюз веб-поиска, глубокого исследования и извлечения контента для агентов экосистемы TheNovaNodes.**

## Статус
Active — Переписан на Go 1.25+ (сентябрь 2026) по стандартам TheNovaNodes.

## Назначение
Предоставляет автономным агентам минималистичный и надежный набор из **3 канонических инструментов** для полноценной работы в сети:
1. `search_web` — быстрый локальный поиск через SearXNG без расхода внешних квот.
2. `fetch_page` — выкачивание веб-страниц в чистый Markdown с обходом WAF/Cloudflare (Firecrawl / Olostep / HTTP fallback).
3. `deep_research` — гибридный исследовательский поиск (SearXNG + Exa AI / Tavily) с дедупликацией и слиянием через Reciprocal Rank Fusion (RRF, $k=60$).

## Архитектура и зависимости
- **Язык:** Go 1.22+ / 1.25
- **MCP SDK:** `github.com/mark3labs/mcp-go v0.58.0`
- **Конкурентность:** Нативные горутины, `golang.org/x/sync/errgroup` с ограниченными контекстными таймаутами.
- **Хранилище секретов:** Динамическое чтение API ключей из RAM-диска (`/dev/shm/agent_vault`).
- **Сетевая защита:** Потокобезопасный `CircuitBreaker` с автоматическим распознаванием таймаутов (`net.Error`) и HTTP кодов 401/403/429/5xx.

## Сборка и тестирование
Сборка бинарника:
```bash
make build
# Создает исполняемый файл bin/searxng-gateway (~7.7 MB)
```

Запуск юнит-тестов с детектором гонок:
```bash
make test
```

Интеграционные тесты против живого SearXNG:
```bash
go test -v -tags=integration ./internal/server
```

## Configuration (Environment Variables)
| Variable | Default Value | Description |
| :--- | :--- | :--- |
| `SEARXNG_URL` | `http://127.0.0.1:8889` | URL of the SearXNG instance |
| `SEARXNG_DEFAULT_MAX` | `10` | Default number of search results |
| `SEARXNG_MAX_ALLOWED` | `50` | Maximum allowed number of search results |
| `SEARXNG_DEFAULT_LANG`| `auto` | Default search language |
| `SEARXNG_SAFESEARCH`  | `0` | Content filter (0=off, 1=moderate, 2=strict) |
| `SEARXNG_TIMEOUT`     | `10` | Timeout for SearXNG requests (seconds) |
| `CASCADE_TIMEOUT`     | `6` | Timeout for external API cascade (seconds) |
| `RRF_K`               | `60` | RRF ranking constant |
| `AGENT_VAULT_DIR`     | `/dev/shm/agent_vault` | Vault directory for keys |

## Доступные инструменты MCP

| Инструмент | Параметры | Описание |
| :--- | :--- | :--- |
| `search_web` | `query` (str, req), `max_results` (int), `categories` (str), `language` (str), `safesearch` (int), `engines` (str) | Быстрый локальный поиск через SearXNG. Сырые структурированные результаты без LLM-галлюцинаций. |
| `fetch_page` | `url` (str, req) | Выкачивание страницы и очистка в Markdown с обходом WAF (Firecrawl $\rightarrow$ Olostep $\rightarrow$ Native). Строгий Token Guardrail урезает результат до 35,000 символов для защиты контекста LLM. |
| `deep_research` | `query` (str, req), `max_results` (int) | Параллельный опрос SearXNG и Exa AI / Tavily со слиянием через RRF ($k=60$) и удалением дубликатов. |

## Подключение к mcp-router
```yaml
  nova-searxng-gateway:
    transport: stdio
    command: /root/projects/TheNovaNodes/searxng-mcp-gateway/bin/searxng-gateway
    env:
      SEARXNG_URL: http://127.0.0.1:8889
    prefix: nova-searxng-gateway__
```

## Лицензия
MIT
