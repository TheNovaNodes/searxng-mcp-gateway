package server

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/TheNovaNodes/searxng-mcp-gateway/internal/config"
	"github.com/TheNovaNodes/searxng-mcp-gateway/internal/orchestrator"
	"github.com/TheNovaNodes/searxng-mcp-gateway/internal/searxng"
	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

// Server coordinates the SearXNG MCP gateway tools.
type Server struct {
	mcpServer    *mcpserver.MCPServer
	searxClient  *searxng.Client
	orchestrator *orchestrator.Orchestrator
	cfg          config.Config
}

// NewServer initializes the MCP server with the 3 canonical tools.
func NewServer(searxClient *searxng.Client, orc *orchestrator.Orchestrator, cfg config.Config) *Server {
	mcpSrv := mcpserver.NewMCPServer(
		"searxng-mcp-gateway",
		"2.0.0",
		mcpserver.WithToolCapabilities(true),
	)

	s := &Server{
		mcpServer:    mcpSrv,
		searxClient:  searxClient,
		orchestrator: orc,
		cfg:          cfg,
	}

	s.registerTools()
	return s
}

// MCPServer returns the underlying MCP server.
func (s *Server) MCPServer() *mcpserver.MCPServer {
	return s.mcpServer
}

func (s *Server) registerTools() {
	// 1. search_web
	s.mcpServer.AddTool(
		mcp.NewTool("search_web",
			mcp.WithDescription("Быстрый веб-поиск через локальный SearXNG. Сырые структурированные результаты."),
			mcp.WithString("query", mcp.Required(), mcp.Description("Поисковый запрос на естественном языке")),
			mcp.WithNumber("max_results", mcp.Description("Максимум результатов (1..50, по умолчанию 10)")),
			mcp.WithString("categories", mcp.Description("Категория поиска: general, news, it, science, files (по умолчанию general)")),
			mcp.WithString("language", mcp.Description("Язык результатов (auto, ru, en, de...)")),
			mcp.WithNumber("safesearch", mcp.Description("Фильтр контента: 0=off, 1=moderate, 2=strict")),
			mcp.WithString("engines", mcp.Description("Явный пул движков через запятую (напр. 'google,bing')")),
		),
		s.handleSearchWeb,
	)

	// 2. fetch_page
	s.mcpServer.AddTool(
		mcp.NewTool("fetch_page",
			mcp.WithDescription("Выкачивание и конвертация веб-страницы в Markdown с обходом WAF/Cloudflare."),
			mcp.WithString("url", mcp.Required(), mcp.Description("Полный URL страницы для выкачивания")),
		),
		s.handleFetchPage,
	)

	// 3. deep_research
	s.mcpServer.AddTool(
		mcp.NewTool("deep_research",
			mcp.WithDescription("Глубокий исследовательский поиск: параллельный опрос SearXNG и Exa AI со слиянием через Reciprocal Rank Fusion (RRF)."),
			mcp.WithString("query", mcp.Required(), mcp.Description("Исследовательский поисковый запрос")),
			mcp.WithNumber("max_results", mcp.Description("Максимум результатов (1..50, по умолчанию 10)")),
		),
		s.handleDeepResearch,
	)
}

func (s *Server) handleSearchWeb(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := req.RequireString("query")
	if err != nil || strings.TrimSpace(query) == "" {
		return mcp.NewToolResultError("Argument 'query' is required and cannot be empty"), nil
	}

	maxResults := req.GetInt("max_results", s.cfg.DefaultMaxResults)
	if maxResults <= 0 {
		maxResults = s.cfg.DefaultMaxResults
	}
	if maxResults > s.cfg.MaxAllowedResults {
		maxResults = s.cfg.MaxAllowedResults
	}

	categories := req.GetString("categories", "")
	language := req.GetString("language", s.cfg.DefaultLanguage)
	safeSearch := req.GetInt("safesearch", s.cfg.DefaultSafeSearch)
	engines := req.GetString("engines", "")

	searchCtx, cancel := context.WithTimeout(ctx, s.cfg.SearchTimeout)
	defer cancel()

	res, _ := s.searxClient.Search(searchCtx, searxng.SearchParams{
		Query:      strings.TrimSpace(query),
		MaxResults: maxResults,
		Categories: categories,
		Language:   language,
		SafeSearch: safeSearch,
		Engines:    engines,
	})

	jsonBytes, err := json.Marshal(res)
	if err != nil {
		return mcp.NewToolResultError("Failed to serialize response: " + err.Error()), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

func (s *Server) handleFetchPage(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	targetURL, err := req.RequireString("url")
	if err != nil || strings.TrimSpace(targetURL) == "" {
		return mcp.NewToolResultError("Argument 'url' is required and cannot be empty"), nil
	}
	cleanURL := strings.TrimSpace(targetURL)

	scrapeCtx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()

	res := s.orchestrator.ScrapePage(scrapeCtx, cleanURL)

	jsonBytes, err := json.Marshal(res)
	if err != nil {
		return mcp.NewToolResultError("Failed to serialize response: " + err.Error()), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

func (s *Server) handleDeepResearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := req.RequireString("query")
	if err != nil || strings.TrimSpace(query) == "" {
		return mcp.NewToolResultError("Argument 'query' is required and cannot be empty"), nil
	}

	maxResults := req.GetInt("max_results", s.cfg.DefaultMaxResults)
	if maxResults <= 0 {
		maxResults = s.cfg.DefaultMaxResults
	}
	if maxResults > s.cfg.MaxAllowedResults {
		maxResults = s.cfg.MaxAllowedResults
	}

	researchCtx, cancel := context.WithTimeout(ctx, s.cfg.CascadeTimeout+4*time.Second)
	defer cancel()

	res := s.orchestrator.DeepResearch(researchCtx, strings.TrimSpace(query), maxResults)

	jsonBytes, err := json.Marshal(res)
	if err != nil {
		return mcp.NewToolResultError("Failed to serialize response: " + err.Error()), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}
