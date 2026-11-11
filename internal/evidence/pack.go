// Package evidence builds a short citation pack from selected hits and extracts.
package evidence

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/tavric-labs/google-search-mcp/internal/hits"
)

// Item is one evidence line.
type Item struct {
	HitID   string
	Title   string
	URL     string
	Snippet string
	Extract string
}

// Pack formats items under a character budget, query-aware when question is set.
func Pack(question string, items []Item, maxChars int) string {
	if maxChars <= 0 {
		maxChars = 3200
	}
	var b strings.Builder
	if q := strings.TrimSpace(question); q != "" {
		fmt.Fprintf(&b, "Question: %s\n\n", q)
	}
	used := utf8.RuneCountInString(b.String())
	for i, item := range items {
		block := formatItem(i+1, item)
		need := utf8.RuneCountInString(block) + 1
		if used+need > maxChars {
			remain := maxChars - used - 20
			if remain < 80 {
				break
			}
			block = truncate(block, remain) + "\n"
			b.WriteString(block)
			break
		}
		b.WriteString(block)
		if i+1 < len(items) {
			b.WriteByte('\n')
		}
		used += need
	}
	out := strings.TrimSpace(b.String())
	if out == "" {
		return "No evidence selected."
	}
	return out
}

// FromHits builds items from stored hits.
func FromHits(hs []hits.Hit, extracts map[string]string) []Item {
	out := make([]Item, 0, len(hs))
	for _, h := range hs {
		out = append(out, Item{
			HitID:   h.ID,
			Title:   h.Title,
			URL:     h.URL,
			Snippet: h.Snippet,
			Extract: extracts[h.ID],
		})
	}
	return out
}

func formatItem(n int, item Item) string {
	var b strings.Builder
	fmt.Fprintf(&b, "[%d] %s\n%s\n", n, item.Title, item.URL)
	if item.Extract != "" {
		fmt.Fprintf(&b, "%s\n", item.Extract)
	} else if item.Snippet != "" {
		fmt.Fprintf(&b, "%s\n", item.Snippet)
	}
	if item.HitID != "" {
		fmt.Fprintf(&b, "hit_id=%s\n", item.HitID)
	}
	return b.String()
}

func truncate(text string, maxChars int) string {
	if utf8.RuneCountInString(text) <= maxChars {
		return text
	}
	runes := []rune(text)
	return string(runes[:maxChars]) + "…"
}
