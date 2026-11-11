package tools

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/tavric-labs/google-search-mcp/internal/config"
	"github.com/tavric-labs/google-search-mcp/internal/gemini"
	"github.com/tavric-labs/google-search-mcp/internal/hits"
)

type fakeSearch struct {
	hits    []hits.Hit
	lastKey string
}

func (f *fakeSearch) Search(_ context.Context, p gemini.Params) ([]hits.Hit, error) {
	f.lastKey = p.APIKey
	return append([]hits.Hit{}, f.hits...), nil
}

type fakeFetch struct{}

func (fakeFetch) Extract(_ context.Context, rawURL, focus string, _ int) (string, error) {
	return "extract about " + focus + " from " + rawURL, nil
}

func textOf(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if result == nil || len(result.Content) == 0 {
		t.Fatal("empty result")
	}
	content, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("content is %T", result.Content[0])
	}
	return content.Text
}

func withAPIKey(key string) *mcp.CallToolRequest {
	return &mcp.CallToolRequest{Extra: &mcp.RequestExtra{Header: http.Header{config.APIKeyHeader: {key}}}}
}

func TestCraftThenSearchFlow(t *testing.T) {
	store := hits.New(0)
	search := &fakeSearch{hits: []hits.Hit{{Title: "A", URL: "https://example.com/a", Snippet: "enough snippet text for a strong hit here"}}}
	server := &Server{Searcher: search, Fetch: fakeFetch{}, Hits: store, MaxResults: 5, MaxParallel: 3, MaxOpenChars: 1000, MaxPackChars: 2000}

	crafted, _, err := server.CraftQuery(context.Background(), nil, CraftArgs{Goal: "please find MCP gateways", Style: "docs"})
	if err != nil || crafted.IsError {
		t.Fatalf("craft: %+v err=%v", crafted, err)
	}
	if !strings.Contains(textOf(t, crafted), "primary:") {
		t.Fatalf("craft text = %q", textOf(t, crafted))
	}

	result, _, err := server.Search(context.Background(), withAPIKey("user-key"), SearchArgs{Query: "MCP gateways"})
	if err != nil || result.IsError {
		t.Fatalf("search: %+v err=%v", result, err)
	}
	if search.lastKey != "user-key" {
		t.Fatalf("api key = %q", search.lastKey)
	}
	body := textOf(t, result)
	if !strings.Contains(body, "hit_id=h_") || !strings.Contains(body, "https://example.com/a") {
		t.Fatalf("search text = %q", body)
	}
}

func TestSearchRequiresAPIKeyHeader(t *testing.T) {
	server := &Server{Searcher: &fakeSearch{}, Hits: hits.New(0), MaxResults: 5}
	result, _, err := server.Search(context.Background(), nil, SearchArgs{Query: "q"})
	if err != nil || !result.IsError {
		t.Fatalf("expected missing-key error, got %+v err=%v", result, err)
	}
	if !strings.Contains(textOf(t, result), config.APIKeyHeader) {
		t.Fatalf("text = %q", textOf(t, result))
	}
}

func TestPackUsesStoredHits(t *testing.T) {
	store := hits.New(0)
	stored := store.Put([]hits.Hit{{Title: "A", URL: "https://example.com/a", Snippet: "snippet A"}})
	server := &Server{Hits: store, MaxPackChars: 2000}
	result, _, err := server.PackEvidence(context.Background(), nil, PackArgs{HitIDs: []string{stored[0].ID}, Question: "What is A?"})
	if err != nil || result.IsError {
		t.Fatalf("pack: %+v err=%v", result, err)
	}
	if !strings.Contains(textOf(t, result), "[1]") || !strings.Contains(textOf(t, result), "What is A?") {
		t.Fatalf("pack = %q", textOf(t, result))
	}
}

func TestOpenRequiresHitOrURL(t *testing.T) {
	server := &Server{Fetch: fakeFetch{}, Hits: hits.New(0), MaxOpenChars: 1000}
	result, _, err := server.Open(context.Background(), nil, OpenArgs{})
	if err != nil || !result.IsError {
		t.Fatalf("expected error result, got %+v err=%v", result, err)
	}
}
