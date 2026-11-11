// Package tools implements the Google Search MCP tools.
package tools

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/tavric-labs/google-search-mcp/internal/config"
	"github.com/tavric-labs/google-search-mcp/internal/craft"
	"github.com/tavric-labs/google-search-mcp/internal/cse"
	"github.com/tavric-labs/google-search-mcp/internal/evidence"
	"github.com/tavric-labs/google-search-mcp/internal/hits"
)

// Searcher runs Google Custom Search.
type Searcher interface {
	Search(ctx context.Context, p cse.Params) ([]hits.Hit, error)
}

// Fetcher opens one page.
type Fetcher interface {
	Extract(ctx context.Context, rawURL, focus string, maxChars int) (string, error)
}

// Server holds tool dependencies.
type Server struct {
	CSE          Searcher
	Fetch        Fetcher
	Hits         *hits.Store
	CSEID        string
	MaxResults   int
	MaxParallel  int
	MaxOpenChars int
	MaxPackChars int
	extracts     sync.Map
}

// CraftArgs are inputs for google_craft_query.
type CraftArgs struct {
	Goal        string   `json:"goal" jsonschema:"what you want to find, in plain language"`
	Style       string   `json:"style,omitempty" jsonschema:"fact, news, papers, docs, or howto"`
	MustInclude []string `json:"must_include,omitempty" jsonschema:"phrases that should appear"`
	MustExclude []string `json:"must_exclude,omitempty" jsonschema:"words or phrases to exclude"`
	PreferSites []string `json:"prefer_sites,omitempty" jsonschema:"preferred hostnames, e.g. arxiv.org"`
	Freshness   string   `json:"freshness,omitempty" jsonschema:"day, week, month, year, or any"`
	MaxVariants int      `json:"max_variants,omitempty" jsonschema:"1 to 3 query variants"`
}

// SearchArgs are inputs for google_search.
type SearchArgs struct {
	Query        string `json:"query" jsonschema:"short Google-ready query from google_craft_query when possible"`
	Num          int    `json:"num,omitempty" jsonschema:"results to return, default server max up to 10"`
	Start        int    `json:"start,omitempty" jsonschema:"1-based start index for pagination"`
	Site         string `json:"site,omitempty" jsonschema:"restrict to one site"`
	FileType     string `json:"filetype,omitempty" jsonschema:"e.g. pdf"`
	DateRestrict string `json:"date_restrict,omitempty" jsonschema:"e.g. d7, w1, m1, y1"`
	ExactTerms   string `json:"exact_terms,omitempty"`
	ExcludeTerms string `json:"exclude_terms,omitempty"`
	GL           string `json:"gl,omitempty" jsonschema:"two-letter country code"`
	HL           string `json:"hl,omitempty" jsonschema:"interface language"`
	Sort         string `json:"sort,omitempty" jsonschema:"set date for freshness"`
}

// MultiArgs are inputs for google_multi_search.
type MultiArgs struct {
	Queries      []string `json:"queries" jsonschema:"2 or 3 short Google-ready queries"`
	NumPerQuery  int      `json:"num_per_query,omitempty"`
	Site         string   `json:"site,omitempty"`
	FileType     string   `json:"filetype,omitempty"`
	DateRestrict string   `json:"date_restrict,omitempty"`
	ExcludeTerms string   `json:"exclude_terms,omitempty"`
}

// OpenArgs are inputs for google_open.
type OpenArgs struct {
	HitID string `json:"hit_id,omitempty" jsonschema:"id from a prior search hit"`
	URL   string `json:"url,omitempty" jsonschema:"https URL if hit_id is unavailable"`
	Focus string `json:"focus,omitempty" jsonschema:"what to extract from the page"`
}

// PackArgs are inputs for google_pack_evidence.
type PackArgs struct {
	HitIDs   []string `json:"hit_ids" jsonschema:"hit ids to include"`
	Question string   `json:"question,omitempty" jsonschema:"user question for a tight pack"`
}

// CraftQuery designs targeted Google queries without calling Google.
func (s *Server) CraftQuery(_ context.Context, _ *mcp.CallToolRequest, args CraftArgs) (*mcp.CallToolResult, any, error) {
	result, err := craft.Build(craft.Request{
		Goal: args.Goal, Style: args.Style, MustInclude: args.MustInclude, MustExclude: args.MustExclude,
		PreferSites: args.PreferSites, Freshness: args.Freshness, MaxVariants: args.MaxVariants,
	})
	if err != nil {
		return textResult(err.Error(), true), nil, nil
	}
	var b strings.Builder
	fmt.Fprintf(&b, "primary: %s\n", result.Primary)
	for i, alt := range result.Alts {
		fmt.Fprintf(&b, "alt_%d: %s\n", i+1, alt)
	}
	fmt.Fprintf(&b, "notes: %s\n", result.Notes)
	if len(result.Filters) > 0 {
		b.WriteString("filters:")
		for k, v := range result.Filters {
			fmt.Fprintf(&b, " %s=%s", k, v)
		}
		b.WriteByte('\n')
	}
	return textResult(strings.TrimSpace(b.String()), false), nil, nil
}

// Search runs one Custom Search request.
func (s *Server) Search(ctx context.Context, request *mcp.CallToolRequest, args SearchArgs) (*mcp.CallToolResult, any, error) {
	apiKey, failure := apiKeyFrom(request)
	if failure != nil {
		return failure, nil, nil
	}
	num := args.Num
	if num <= 0 {
		num = s.MaxResults
	}
	found, err := s.CSE.Search(ctx, cse.Params{
		Query: args.Query, APIKey: apiKey, CSEID: s.CSEID, Num: num, Start: args.Start, Site: args.Site,
		FileType: args.FileType, DateRestrict: args.DateRestrict, ExactTerms: args.ExactTerms,
		ExcludeTerms: args.ExcludeTerms, GL: args.GL, HL: args.HL, Sort: args.Sort,
	})
	if err != nil {
		return textResult("Google search failed. Check the query and X-Google-Api-Key.", true), nil, nil
	}
	stored := s.Hits.Put(found)
	return textResult(formatHits(stored), false), nil, nil
}

// MultiSearch runs a few queries in parallel and dedupes by URL.
func (s *Server) MultiSearch(ctx context.Context, request *mcp.CallToolRequest, args MultiArgs) (*mcp.CallToolResult, any, error) {
	apiKey, failure := apiKeyFrom(request)
	if failure != nil {
		return failure, nil, nil
	}
	queries := cleanQueries(args.Queries, s.MaxParallel)
	if len(queries) == 0 {
		return textResult("Provide 1 to 3 short queries.", true), nil, nil
	}
	num := args.NumPerQuery
	if num <= 0 {
		num = 3
	}
	type outcome struct {
		hits []hits.Hit
		err  error
	}
	ch := make(chan outcome, len(queries))
	for _, q := range queries {
		q := q
		go func() {
			h, err := s.CSE.Search(ctx, cse.Params{
				Query: q, APIKey: apiKey, CSEID: s.CSEID, Num: num, Site: args.Site, FileType: args.FileType,
				DateRestrict: args.DateRestrict, ExcludeTerms: args.ExcludeTerms,
			})
			ch <- outcome{hits: h, err: err}
		}()
	}
	seen := map[string]bool{}
	var merged []hits.Hit
	var firstErr error
	for range queries {
		out := <-ch
		if out.err != nil && firstErr == nil {
			firstErr = out.err
			continue
		}
		for _, h := range out.hits {
			if seen[h.URL] {
				continue
			}
			seen[h.URL] = true
			merged = append(merged, h)
		}
	}
	if len(merged) == 0 && firstErr != nil {
		return textResult("Google multi-search failed. Check the queries and X-Google-Api-Key.", true), nil, nil
	}
	stored := s.Hits.Put(merged)
	return textResult(formatHits(stored), false), nil, nil
}

// Open extracts a short window from one prior hit or URL.
func (s *Server) Open(ctx context.Context, _ *mcp.CallToolRequest, args OpenArgs) (*mcp.CallToolResult, any, error) {
	rawURL := strings.TrimSpace(args.URL)
	hitID := strings.TrimSpace(args.HitID)
	if hitID != "" {
		hit, ok := s.Hits.Get(hitID)
		if !ok {
			return textResult("Unknown hit_id. Search again, then open with a fresh id.", true), nil, nil
		}
		rawURL = hit.URL
	}
	if rawURL == "" {
		return textResult("Provide hit_id from a prior search, or an https url.", true), nil, nil
	}
	text, err := s.Fetch.Extract(ctx, rawURL, args.Focus, s.MaxOpenChars)
	if err != nil {
		return textResult("Could not extract useful text from that page.", true), nil, nil
	}
	if hitID != "" {
		s.extracts.Store(hitID, text)
	}
	return textResult(fmt.Sprintf("url: %s\n\n%s", rawURL, text), false), nil, nil
}

// PackEvidence builds a citation pack from selected hits.
func (s *Server) PackEvidence(_ context.Context, _ *mcp.CallToolRequest, args PackArgs) (*mcp.CallToolResult, any, error) {
	if len(args.HitIDs) == 0 {
		return textResult("Provide at least one hit_id.", true), nil, nil
	}
	selected := s.Hits.GetMany(args.HitIDs)
	if len(selected) == 0 {
		return textResult("No matching hits in memory. Search again first.", true), nil, nil
	}
	extracts := map[string]string{}
	for _, id := range args.HitIDs {
		if v, ok := s.extracts.Load(id); ok {
			extracts[id], _ = v.(string)
		}
	}
	pack := evidence.Pack(args.Question, evidence.FromHits(selected, extracts), s.MaxPackChars)
	return textResult(pack, false), nil, nil
}

func apiKeyFrom(request *mcp.CallToolRequest) (string, *mcp.CallToolResult) {
	if request != nil && request.Extra != nil && request.Extra.Header != nil {
		if key := strings.TrimSpace(request.Extra.Header.Get(config.APIKeyHeader)); key != "" {
			return key, nil
		}
	}
	return "", textResult("Missing "+config.APIKeyHeader+". Create a Google API key with Custom Search API enabled and send it in that header.", true)
}

func formatHits(list []hits.Hit) string {
	if len(list) == 0 {
		return "No results. Craft a tighter query (site/filetype/exclude) and search again."
	}
	var b strings.Builder
	for i, h := range list {
		weak := ""
		if h.Weak {
			weak = " weak=true"
		}
		fmt.Fprintf(&b, "[%d] %s%s\n%s\n%s\nhit_id=%s\n", i+1, h.Title, weak, h.URL, h.Snippet, h.ID)
		if i+1 < len(list) {
			b.WriteByte('\n')
		}
	}
	return strings.TrimSpace(b.String())
}

func cleanQueries(in []string, max int) []string {
	if max <= 0 {
		max = 3
	}
	out := make([]string, 0, max)
	seen := map[string]bool{}
	for _, q := range in {
		q = strings.TrimSpace(q)
		if q == "" || seen[q] {
			continue
		}
		seen[q] = true
		out = append(out, q)
		if len(out) >= max {
			break
		}
	}
	return out
}

func textResult(text string, isError bool) *mcp.CallToolResult {
	return &mcp.CallToolResult{IsError: isError, Content: []mcp.Content{&mcp.TextContent{Text: text}}}
}
