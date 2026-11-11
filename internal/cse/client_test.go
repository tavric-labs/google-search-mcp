package cse

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchDecodesSlimHits(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("q") != "mcp runtime" || r.URL.Query().Get("cx") != "cx1" {
			t.Errorf("unexpected query %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"items":[{"title":"MCP Runtime","link":"https://example.com/a","snippet":"A runtime for MCP servers.","displayLink":"example.com"}]}`))
	}))
	defer server.Close()

	hits, err := New("key", server.URL).Search(context.Background(), Params{Query: "mcp runtime", CSEID: "cx1", Num: 3})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 || hits[0].Title != "MCP Runtime" || hits[0].URL != "https://example.com/a" {
		t.Fatalf("hits = %+v", hits)
	}
}

func TestSearchRequiresQuery(t *testing.T) {
	if _, err := New("key", "http://127.0.0.1").Search(context.Background(), Params{CSEID: "cx"}); err == nil {
		t.Fatal("empty query accepted")
	}
}
