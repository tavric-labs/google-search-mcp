// Package cse calls the Google Custom Search JSON API and returns slim hits.
package cse

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tavric-labs/google-search-mcp/internal/hits"
)

// Params are the filters sent to Custom Search.
type Params struct {
	Query        string
	CSEID        string
	Num          int
	Start        int
	Site         string
	FileType     string
	DateRestrict string
	ExactTerms   string
	ExcludeTerms string
	GL, HL       string
	Sort         string
}

// Client talks to the Custom Search JSON API.
type Client struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

// New returns a CSE client.
func New(apiKey, baseURL string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 20 * time.Second},
	}
}

// Search runs one Custom Search request.
func (c *Client) Search(ctx context.Context, p Params) ([]hits.Hit, error) {
	if strings.TrimSpace(p.Query) == "" {
		return nil, fmt.Errorf("query is required")
	}
	if p.CSEID == "" {
		return nil, fmt.Errorf("cse id is required")
	}
	if p.Num <= 0 {
		p.Num = 5
	}
	if p.Num > 10 {
		p.Num = 10
	}
	if p.Start <= 0 {
		p.Start = 1
	}
	values := url.Values{}
	values.Set("key", c.apiKey)
	values.Set("cx", p.CSEID)
	values.Set("q", p.Query)
	values.Set("num", fmt.Sprintf("%d", p.Num))
	values.Set("start", fmt.Sprintf("%d", p.Start))
	values.Set("filter", "1")
	if p.Site != "" {
		values.Set("siteSearch", p.Site)
		values.Set("siteSearchFilter", "i")
	}
	if p.FileType != "" {
		values.Set("fileType", p.FileType)
	}
	if p.DateRestrict != "" {
		values.Set("dateRestrict", p.DateRestrict)
	}
	if p.ExactTerms != "" {
		values.Set("exactTerms", p.ExactTerms)
	}
	if p.ExcludeTerms != "" {
		values.Set("excludeTerms", p.ExcludeTerms)
	}
	if p.GL != "" {
		values.Set("gl", p.GL)
	}
	if p.HL != "" {
		values.Set("hl", p.HL)
	}
	if p.Sort != "" {
		values.Set("sort", p.Sort)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"?"+values.Encode(), nil)
	if err != nil {
		return nil, err
	}
	response, err := c.http.Do(request)
	if err != nil {
		return nil, fmt.Errorf("google custom search: %w", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	if err != nil {
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("google custom search returned HTTP %d", response.StatusCode)
	}
	var parsed struct {
		Items []struct {
			Title       string `json:"title"`
			Link        string `json:"link"`
			Snippet     string `json:"snippet"`
			DisplayLink string `json:"displayLink"`
		} `json:"items"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("decode custom search response: %w", err)
	}
	out := make([]hits.Hit, 0, len(parsed.Items))
	for _, item := range parsed.Items {
		snippet := strings.TrimSpace(item.Snippet)
		out = append(out, hits.Hit{
			Title:   strings.TrimSpace(item.Title),
			URL:     strings.TrimSpace(item.Link),
			Snippet: snippet,
			Display: strings.TrimSpace(item.DisplayLink),
			Query:   p.Query,
			Weak:    len(snippet) < 40,
		})
	}
	return out, nil
}
