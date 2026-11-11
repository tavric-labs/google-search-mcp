// Command google-search-mcp is a token-efficient Google web search MCP server
// backed by Gemini Google Search grounding. Clients send X-Gemini-Api-Key
// (a Gemini AI Studio API key). There is no Custom Search / CSE id.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/tavric-labs/google-search-mcp/internal/config"
	"github.com/tavric-labs/google-search-mcp/internal/fetch"
	"github.com/tavric-labs/google-search-mcp/internal/gemini"
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
		Searcher:     gemini.New(cfg.APIBase, cfg.Model),
		Fetch:        fetch.New(),
		Hits:         hits.New(30 * time.Minute),
		MaxResults:   cfg.MaxResults,
		MaxParallel:  cfg.MaxParallel,
		MaxOpenChars: cfg.MaxOpenChars,
		MaxPackChars: cfg.MaxPackChars,
	}

	server := mcp.NewServer(&mcp.Implementation{Name: "google-search-mcp", Version: "0.2.0"}, &mcp.ServerOptions{
		Instructions: "Generic Google web search for any topic via Gemini Search grounding. Send X-Gemini-Api-Key with a Gemini AI Studio API key. Craft a short targeted query first, search for slim hits, open at most 1-2 pages if needed, then pack evidence. Prefer google_craft_query before google_search. Keep result counts small.",
	})
	mcp.AddTool(server, &mcp.Tool{
		Name:        "google_craft_query",
		Description: "Design 1-3 short Google-ready queries from any plain-language goal (news, products, docs, how-tos, people, places, etc.). Does not call Google.",
	}, handlers.CraftQuery)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "google_search",
		Description: "Generic Google web search via Gemini grounding. Returns slim hits (title, url, snippet, hit_id). Prefer a crafted query. Default few results.",
	}, handlers.Search)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "google_multi_search",
		Description: "Run up to 3 short queries in parallel for any multi-part question; merge and dedupe hits.",
	}, handlers.MultiSearch)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "google_open",
		Description: "Fetch one prior hit (or https URL) and return a short text extract under a char budget. Use sparingly.",
	}, handlers.Open)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "google_pack_evidence",
		Description: "Build a short citation pack from selected hit_ids for the final answer.",
	}, handlers.PackEvidence)

	mcpHandler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, &mcp.StreamableHTTPOptions{
		Stateless: true, JSONResponse: true,
	})
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"ok":true}`)
	})
	mux.Handle("/", headerAuth(cfg.MCPPath, mcpHandler))

	httpServer := &http.Server{
		Addr: ":" + cfg.Port, Handler: mux,
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 20 * time.Second,
		WriteTimeout: 90 * time.Second, IdleTimeout: 60 * time.Second,
	}
	stop, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	go func() {
		<-stop.Done()
		shutdown, done := context.WithTimeout(context.Background(), 10*time.Second)
		defer done()
		_ = httpServer.Shutdown(shutdown)
	}()
	log.Printf("google-search-mcp listening on :%s at %s (header %s, model %s)", cfg.Port, cfg.MCPPath, config.APIKeyHeader, cfg.Model)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func headerAuth(mcpPath string, next http.Handler) http.Handler {
	checkPath := mcpPath + "/auth/check"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != mcpPath && r.URL.Path != checkPath {
			http.NotFound(w, r)
			return
		}
		key := strings.TrimSpace(r.Header.Get(config.APIKeyHeader))
		if key == "" {
			w.Header().Set("WWW-Authenticate", config.APIKeyHeader)
			http.Error(w, "missing "+config.APIKeyHeader, http.StatusUnauthorized)
			return
		}
		if r.URL.Path == checkPath {
			if r.Method != http.MethodGet {
				w.Header().Set("Allow", http.MethodGet)
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			w.Header().Set("Cache-Control", "no-store")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
