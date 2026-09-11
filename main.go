package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/TheNovaNodes/searxng-mcp-gateway/internal/config"
	"github.com/TheNovaNodes/searxng-mcp-gateway/internal/echelon"
	"github.com/TheNovaNodes/searxng-mcp-gateway/internal/orchestrator"
	"github.com/TheNovaNodes/searxng-mcp-gateway/internal/searxng"
	"github.com/TheNovaNodes/searxng-mcp-gateway/internal/server"
	"github.com/TheNovaNodes/searxng-mcp-gateway/internal/vault"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func main() {
	cfg := config.Load()

	searxClient := searxng.NewClient(cfg.SearXNGURL, cfg.SearchTimeout)
	v := vault.NewVault(cfg.VaultDir)
	cb := echelon.NewCircuitBreaker(60)

	orc := orchestrator.NewOrchestrator(searxClient, v, cb, cfg.RRFK, cfg.AllowPrivateScrape)
	srv := server.NewServer(searxClient, orc, cfg)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Printf("Starting searxng-mcp-gateway (SearXNG: %s, RRF_K: %d)", cfg.SearXNGURL, cfg.RRFK)

	if err := mcpserver.ServeStdio(srv.MCPServer()); err != nil {
		if ctx.Err() != nil {
			log.Println("MCP Server shut down cleanly.")
			return
		}
		fmt.Fprintf(os.Stderr, "FATAL: MCP Server terminated with error: %v\n", err)
		os.Exit(1)
	}
}
