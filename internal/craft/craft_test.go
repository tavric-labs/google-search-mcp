package craft

import (
	"strings"
	"testing"
)

func TestBuildStripsPoliteNoiseAndAddsOperators(t *testing.T) {
	result, err := Build(Request{
		Goal:        "please find query rewriting for RAG agents",
		Style:       "papers",
		MustExclude: []string{"cheatsheet"},
		Freshness:   "year",
		MaxVariants: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.ToLower(result.Primary), "please") {
		t.Fatalf("primary still polite: %q", result.Primary)
	}
	if !strings.Contains(result.Primary, "site:arxiv.org") {
		t.Fatalf("papers style should prefer arxiv: %q", result.Primary)
	}
	if !strings.Contains(result.Primary, "-cheatsheet") {
		t.Fatalf("exclude missing: %q", result.Primary)
	}
	if result.Filters["filetype"] != "pdf" {
		t.Fatalf("filters = %#v", result.Filters)
	}
	if len(result.Alts) == 0 {
		t.Fatal("expected alt variants")
	}
}

func TestBuildRequiresGoal(t *testing.T) {
	if _, err := Build(Request{}); err == nil {
		t.Fatal("empty goal accepted")
	}
}
