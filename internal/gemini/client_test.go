package gemini

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSearchMapsGroundingChunks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-goog-api-key") != "key" {
			t.Errorf("missing api key header")
		}
		if !strings.Contains(r.URL.Path, "/models/gemini-2.5-flash:generateContent") {
			t.Errorf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{
			"candidates":[{
				"content":{"parts":[{"text":"MCP Runtime hosts MCP servers."}]},
				"groundingMetadata":{
					"webSearchQueries":["mcp runtime"],
					"groundingChunks":[
						{"web":{"uri":"https://example.com/a","title":"MCP Runtime"}},
						{"web":{"uri":"https://example.com/b","title":"Docs"}}
					],
					"groundingSupports":[
						{"segment":{"text":"A runtime for MCP servers."},"groundingChunkIndices":[0]}
					]
				}
			}]
		}`))
	}))
	defer server.Close()

	found, err := New(server.URL, "gemini-2.5-flash").Search(context.Background(), Params{
		Query: "mcp runtime", APIKey: "key", Num: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 2 || found[0].Title != "MCP Runtime" || found[0].URL != "https://example.com/a" {
		t.Fatalf("hits = %+v", found)
	}
	if found[0].Snippet != "A runtime for MCP servers." {
		t.Fatalf("snippet = %q", found[0].Snippet)
	}
}

func TestSearchRequiresQueryAndAPIKey(t *testing.T) {
	client := New("http://127.0.0.1", "gemini-2.5-flash")
	if _, err := client.Search(context.Background(), Params{APIKey: "key"}); err == nil {
		t.Fatal("empty query accepted")
	}
	if _, err := client.Search(context.Background(), Params{Query: "q"}); err == nil {
		t.Fatal("empty api key accepted")
	}
}
