// Command google-search-mcp is a token-efficient Google Custom Search MCP server:
// craft targeted queries, search slim SERPs, open a few pages, pack citations.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/tavric-labs/google-search-mcp/internal/config"
	"github.com/tavric-labs/google-search-mcp/internal/cse"
	"github.com/tavric-labs/google-search-mcp/internal/fetch"
	"github.com/tavric-labs/google-search-mcp/internal/hits"
	"github.com/tavric-labs/google-search-mcp/internal/tools"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return err
	}
	handlers := &tools.Server{
		CSE:          cse.New(cfg.APIKey, cfg.APIBase),
		Fetch:        fetch.New(),
		Hits:         hits.New(30 * time.Minute),
		CSEID:        cfg.CSEID,
		PapersCSEID:  cfg.PapersCSEID,
		MaxResults:   cfg.MaxResults,
		MaxParallel:  cfg.MaxParallel,
		MaxOpenChars: cfg.MaxOpenChars,
		MaxPackChars: cfg.MaxPackChars,
	}

	server := mcp.NewServer(&mcp.Implementation{Name: "google-search-mcp", Version: "0.1.0"}, &mcp.ServerOptions{
		Instructions: "Token-efficient Google search: craft a targeted query first, search for slim hits, open at most 1-2 pages, then pack evidence with citations. Prefer google_craft_query before google_search. Default to few results.",
	})
	mcp.AddTool(server, &mcp.Tool{
		Name:        "google_craft_query",
		Description: "Design 1-3 short Google-ready queries with operators from a plain-language goal. Does not call Google. Use before searching.",
	}, handlers.CraftQuery)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "google_search",
		Description: "Run one Google Custom Search and return slim hits (title, url, snippet, hit_id). Prefer a crafted query. Default few results.",
	}, handlers.Search)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "google_multi_search",
		Description: "Run up to 3 short queries in parallel, merge and dedupe hits. For multi-hop questions.",
	}, handlers.MultiSearch)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "google_search_papers",
		Description: "Search with a papers bias (pdf / arxiv or papers CSE). Returns slim hits.",
	}, handlers.SearchPapers)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "google_open",
		Description: "Fetch one prior hit (or https URL) and return a short text extract under a char budget. Use sparingly.",
	}, handlers.Open)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "google_pack_evidence",
		Description: "Build a short citation pack from selected hit_ids for the final answer.",
	}, handlers.PackEvidence)

	mcpHandler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, &mcp.StreamableHTTPOptions{JSONResponse: true})
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"ok":true}`)
	})
	mux.Handle(cfg.MCPPath, mcpHandler)

	httpServer := &http.Server{
		Addr: ":" + cfg.Port, Handler: mux,
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 20 * time.Second,
		WriteTimeout: 60 * time.Second, IdleTimeout: 60 * time.Second,
	}
	stop, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	go func() {
		<-stop.Done()
		shutdown, done := context.WithTimeout(context.Background(), 10*time.Second)
		defer done()
		_ = httpServer.Shutdown(shutdown)
	}()
	log.Printf("google-search-mcp listening on :%s at %s", cfg.Port, cfg.MCPPath)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
