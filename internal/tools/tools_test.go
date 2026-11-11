package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/tavric-labs/google-search-mcp/internal/cse"
	"github.com/tavric-labs/google-search-mcp/internal/hits"
)

type fakeSearch struct {
	hits []hits.Hit
}

func (f *fakeSearch) Search(_ context.Context, _ cse.Params) ([]hits.Hit, error) {
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

func TestCraftThenSearchFlow(t *testing.T) {
	store := hits.New(0)
	search := &fakeSearch{hits: []hits.Hit{{Title: "A", URL: "https://example.com/a", Snippet: "enough snippet text for a strong hit here"}}}
	server := &Server{CSE: search, Fetch: fakeFetch{}, Hits: store, CSEID: "cx", MaxResults: 5, MaxParallel: 3, MaxOpenChars: 1000, MaxPackChars: 2000}

	crafted, _, err := server.CraftQuery(context.Background(), nil, CraftArgs{Goal: "please find MCP gateways", Style: "docs"})
	if err != nil || crafted.IsError {
		t.Fatalf("craft: %+v err=%v", crafted, err)
	}
	if !strings.Contains(textOf(t, crafted), "primary:") {
		t.Fatalf("craft text = %q", textOf(t, crafted))
	}

	result, _, err := server.Search(context.Background(), nil, SearchArgs{Query: "MCP gateways"})
	if err != nil || result.IsError {
		t.Fatalf("search: %+v err=%v", result, err)
	}
	body := textOf(t, result)
	if !strings.Contains(body, "hit_id=h_") || !strings.Contains(body, "https://example.com/a") {
		t.Fatalf("search text = %q", body)
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
