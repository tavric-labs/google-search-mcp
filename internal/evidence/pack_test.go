package evidence

import (
	"strings"
	"testing"
)

func TestPackRespectsBudget(t *testing.T) {
	items := []Item{
		{HitID: "h1", Title: "One", URL: "https://example.com/1", Snippet: strings.Repeat("a ", 200)},
		{HitID: "h2", Title: "Two", URL: "https://example.com/2", Snippet: strings.Repeat("b ", 200)},
	}
	out := Pack("budget test", items, 180)
	if !strings.Contains(out, "[1]") {
		t.Fatalf("missing first item: %q", out)
	}
	if len([]rune(out)) > 200 {
		t.Fatalf("pack too long: %d", len([]rune(out)))
	}
}
