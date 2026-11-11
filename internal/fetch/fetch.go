// Package fetch downloads one page and returns a short extract under a char budget.
package fetch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	tagRE   = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>|<style[^>]*>.*?</style>|<[^>]+>`)
	spaceRE = regexp.MustCompile(`\s+`)
)

// Client fetches public HTTPS pages.
type Client struct {
	http *http.Client
}

// New returns a fetch client.
func New() *Client {
	return &Client{http: &http.Client{Timeout: 20 * time.Second}}
}

// Extract gets plain text from url, focused on focus words when possible, capped at maxChars.
func (c *Client) Extract(ctx context.Context, rawURL, focus string, maxChars int) (string, error) {
	if maxChars <= 0 {
		maxChars = 3000
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return "", fmt.Errorf("url must be absolute https")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("User-Agent", "google-search-mcp/0.1")
	response, err := c.http.Do(request)
	if err != nil {
		return "", fmt.Errorf("fetch page: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 1<<16))
		return "", fmt.Errorf("page returned HTTP %d", response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	if err != nil {
		return "", err
	}
	text := tagRE.ReplaceAllString(string(body), " ")
	text = spaceRE.ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)
	if text == "" {
		return "", fmt.Errorf("no extractable text")
	}
	if focus = strings.TrimSpace(focus); focus != "" {
		if clipped := focusWindow(text, focus, maxChars); clipped != "" {
			return clipped, nil
		}
	}
	return truncate(text, maxChars), nil
}

func focusWindow(text, focus string, maxChars int) string {
	lower := strings.ToLower(text)
	idx := strings.Index(lower, strings.ToLower(focus))
	if idx < 0 {
		for _, word := range strings.Fields(focus) {
			if len(word) < 4 {
				continue
			}
			idx = strings.Index(lower, strings.ToLower(word))
			if idx >= 0 {
				break
			}
		}
	}
	if idx < 0 {
		return ""
	}
	start := idx - maxChars/4
	if start < 0 {
		start = 0
	}
	return truncate(text[start:], maxChars)
}

func truncate(text string, maxChars int) string {
	if utf8.RuneCountInString(text) <= maxChars {
		return text
	}
	runes := []rune(text)
	return string(runes[:maxChars]) + "…"
}
