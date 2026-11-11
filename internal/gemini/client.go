// Package gemini calls the Gemini API with Google Search grounding and returns slim hits.
package gemini

import (
	"bytes"
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

// Params are the filters sent into a grounded search prompt.
type Params struct {
	Query        string
	APIKey       string
	Num          int
	Site         string
	FileType     string
	DateRestrict string
	ExactTerms   string
	ExcludeTerms string
	GL, HL       string
	Sort         string
}

// Client talks to the Gemini generateContent endpoint with googleSearch.
type Client struct {
	baseURL string
	model   string
	http    *http.Client
}

// New returns a Gemini search client.
func New(baseURL, model string) *Client {
	if model == "" {
		model = "gemini-2.5-flash"
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		http:    &http.Client{Timeout: 45 * time.Second},
	}
}

// Search runs one grounded Google Search via Gemini and maps citations to hits.
func (c *Client) Search(ctx context.Context, p Params) ([]hits.Hit, error) {
	if strings.TrimSpace(p.Query) == "" {
		return nil, fmt.Errorf("query is required")
	}
	if strings.TrimSpace(p.APIKey) == "" {
		return nil, fmt.Errorf("api key is required")
	}
	if p.Num <= 0 {
		p.Num = 5
	}
	if p.Num > 10 {
		p.Num = 10
	}

	prompt := buildPrompt(p)
	body, err := json.Marshal(map[string]any{
		"contents": []map[string]any{
			{"role": "user", "parts": []map[string]string{{"text": prompt}}},
		},
		"tools": []map[string]any{
			{"googleSearch": map[string]any{}},
		},
	})
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/models/%s:generateContent", c.baseURL, url.PathEscape(c.model))
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("x-goog-api-key", p.APIKey)

	response, err := c.http.Do(request)
	if err != nil {
		return nil, fmt.Errorf("gemini search: %w", err)
	}
	defer response.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(response.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("gemini search returned HTTP %d", response.StatusCode)
	}

	var parsed struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
			GroundingMetadata struct {
				WebSearchQueries []string `json:"webSearchQueries"`
				GroundingChunks  []struct {
					Web struct {
						URI   string `json:"uri"`
						Title string `json:"title"`
					} `json:"web"`
				} `json:"groundingChunks"`
				GroundingSupports []struct {
					Segment struct {
						Text string `json:"text"`
					} `json:"segment"`
					GroundingChunkIndices []int `json:"groundingChunkIndices"`
				} `json:"groundingSupports"`
			} `json:"groundingMetadata"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("decode gemini response: %w", err)
	}
	if len(parsed.Candidates) == 0 {
		return nil, nil
	}
	candidate := parsed.Candidates[0]
	snippets := map[int]string{}
	for _, support := range candidate.GroundingMetadata.GroundingSupports {
		text := strings.TrimSpace(support.Segment.Text)
		if text == "" {
			continue
		}
		for _, idx := range support.GroundingChunkIndices {
			if snippets[idx] == "" {
				snippets[idx] = text
			}
		}
	}
	summary := ""
	for _, part := range candidate.Content.Parts {
		if t := strings.TrimSpace(part.Text); t != "" {
			summary = t
			break
		}
	}

	out := make([]hits.Hit, 0, len(candidate.GroundingMetadata.GroundingChunks))
	seen := map[string]bool{}
	for i, chunk := range candidate.GroundingMetadata.GroundingChunks {
		rawURL := strings.TrimSpace(chunk.Web.URI)
		if rawURL == "" || seen[rawURL] {
			continue
		}
		seen[rawURL] = true
		title := strings.TrimSpace(chunk.Web.Title)
		if title == "" {
			title = rawURL
		}
		snippet := strings.TrimSpace(snippets[i])
		if snippet == "" && summary != "" && len(out) == 0 {
			snippet = truncate(summary, 240)
		}
		display := hostOf(rawURL)
		out = append(out, hits.Hit{
			Title:   title,
			URL:     rawURL,
			Snippet: snippet,
			Display: display,
			Query:   p.Query,
			Weak:    len(snippet) < 40,
		})
		if len(out) >= p.Num {
			break
		}
	}
	return out, nil
}

func buildPrompt(p Params) string {
	var b strings.Builder
	b.WriteString("Search the live web for this query and cite primary sources.\n")
	b.WriteString("Query: ")
	b.WriteString(strings.TrimSpace(p.Query))
	b.WriteByte('\n')
	if p.Site != "" {
		fmt.Fprintf(&b, "Prefer results from site %s.\n", p.Site)
	}
	if p.FileType != "" {
		fmt.Fprintf(&b, "Prefer filetype:%s when relevant.\n", p.FileType)
	}
	if p.DateRestrict != "" {
		fmt.Fprintf(&b, "Prefer fresh results matching date restrict %s.\n", p.DateRestrict)
	}
	if p.ExactTerms != "" {
		fmt.Fprintf(&b, "Must include exact phrase: %s\n", p.ExactTerms)
	}
	if p.ExcludeTerms != "" {
		fmt.Fprintf(&b, "Exclude terms: %s\n", p.ExcludeTerms)
	}
	if p.GL != "" {
		fmt.Fprintf(&b, "Prefer country/region %s.\n", p.GL)
	}
	if p.HL != "" {
		fmt.Fprintf(&b, "Prefer language %s.\n", p.HL)
	}
	if p.Sort == "date" {
		b.WriteString("Prefer the newest sources.\n")
	}
	b.WriteString("Reply with one short factual summary only; grounding will attach the sources.\n")
	return b.String()
}

func hostOf(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return u.Host
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return strings.TrimSpace(s[:n]) + "…"
}
