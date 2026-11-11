package cse

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchDecodesSlimHits(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("q") != "mcp runtime" || r.URL.Query().Get("cx") != "cx1" || r.URL.Query().Get("key") != "key" {
			t.Errorf("unexpected query %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"items":[{"title":"MCP Runtime","link":"https://example.com/a","snippet":"A runtime for MCP servers.","displayLink":"example.com"}]}`))
	}))
	defer server.Close()

	hits, err := New(server.URL).Search(context.Background(), Params{Query: "mcp runtime", APIKey: "key", CSEID: "cx1", Num: 3})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 || hits[0].Title != "MCP Runtime" || hits[0].URL != "https://example.com/a" {
		t.Fatalf("hits = %+v", hits)
	}
}

func TestSearchRequiresQueryAndAPIKey(t *testing.T) {
	client := New("http://127.0.0.1")
	if _, err := client.Search(context.Background(), Params{CSEID: "cx", APIKey: "key"}); err == nil {
		t.Fatal("empty query accepted")
	}
	if _, err := client.Search(context.Background(), Params{Query: "q", CSEID: "cx"}); err == nil {
		t.Fatal("empty api key accepted")
	}
}
