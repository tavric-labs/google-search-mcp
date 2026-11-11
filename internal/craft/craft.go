// Package craft turns a natural-language goal into short Google-ready queries
// with operators. It does not call Google; it only designs the query.
package craft

import (
	"fmt"
	"strings"
)

// Request is the craft_query input.
type Request struct {
	Goal        string
	Style       string
	MustInclude []string
	MustExclude []string
	PreferSites []string
	Freshness   string
	MaxVariants int
}

// Result is one or more targeted queries plus CSE filter hints.
type Result struct {
	Primary string            `json:"primary"`
	Alts    []string          `json:"alts,omitempty"`
	Notes   string            `json:"notes"`
	Filters map[string]string `json:"filters,omitempty"`
}

// Build designs targeted queries from a goal. It is deterministic and free.
func Build(req Request) (Result, error) {
	goal := cleanGoal(req.Goal)
	if goal == "" {
		return Result{}, fmt.Errorf("goal is required")
	}
	style := strings.ToLower(strings.TrimSpace(req.Style))
	if style == "" {
		style = "fact"
	}
	max := req.MaxVariants
	if max <= 0 {
		max = 2
	}
	if max > 3 {
		max = 3
	}

	sites := append([]string{}, req.PreferSites...)
	if style == "papers" && len(sites) == 0 {
		sites = []string{"arxiv.org"}
	}

	filters := map[string]string{}
	switch style {
	case "papers":
		filters["filetype"] = "pdf"
		filters["dateRestrict"] = freshness(req.Freshness, "y5")
	case "news":
		filters["dateRestrict"] = freshness(req.Freshness, "w1")
		filters["sort"] = "date"
	case "howto":
		filters["dateRestrict"] = freshness(req.Freshness, "y2")
	default:
		if dr := freshness(req.Freshness, ""); dr != "" {
			filters["dateRestrict"] = dr
		}
	}

	primary := compose(goal, style, req.MustInclude, req.MustExclude, sites)
	alts := variants(goal, style, sites, max-1)

	notes := fmt.Sprintf("style=%s; use these strings with google_search or google_multi_search; do not paste the raw goal as the engine query.", style)
	if filters["dateRestrict"] != "" {
		notes += " apply dateRestrict=" + filters["dateRestrict"] + "."
	}
	if filters["filetype"] != "" {
		notes += " apply filetype=" + filters["filetype"] + "."
	}
	return Result{Primary: primary, Alts: alts, Notes: notes, Filters: omitEmpty(filters)}, nil
}

func compose(goal, style string, include, exclude, sites []string) string {
	parts := []string{goal}
	switch style {
	case "docs":
		parts = append(parts, "(documentation OR docs OR reference)")
	case "howto":
		parts = append(parts, `(howto OR "how to" OR guide OR tutorial)`)
	case "news":
		parts = append(parts, "(news OR announcement OR release)")
	}
	for _, phrase := range include {
		phrase = strings.TrimSpace(phrase)
		if phrase == "" {
			continue
		}
		if strings.Contains(phrase, " ") {
			parts = append(parts, quote(phrase))
		} else {
			parts = append(parts, phrase)
		}
	}
	for _, site := range sites {
		site = strings.TrimPrefix(strings.TrimSpace(site), "site:")
		if site != "" {
			parts = append(parts, "site:"+site)
		}
	}
	for _, ex := range exclude {
		ex = strings.TrimSpace(strings.ReplaceAll(ex, `"`, ""))
		if ex == "" {
			continue
		}
		if strings.Contains(ex, " ") {
			parts = append(parts, "-"+quote(ex))
		} else {
			parts = append(parts, "-"+ex)
		}
	}
	return strings.Join(parts, " ")
}

func quote(s string) string {
	return `"` + s + `"`
}

func cleanGoal(goal string) string {
	goal = strings.TrimSpace(goal)
	lower := strings.ToLower(goal)
	for _, prefix := range []string{"please ", "can you ", "find ", "search for ", "look up ", "i need ", "help me "} {
		if strings.HasPrefix(lower, prefix) {
			goal = strings.TrimSpace(goal[len(prefix):])
			lower = strings.ToLower(goal)
		}
	}
	goal = strings.TrimSpace(strings.TrimSuffix(goal, "?"))
	return strings.Join(strings.Fields(goal), " ")
}

func firstPhrase(goal string) string {
	words := strings.Fields(goal)
	if len(words) == 0 {
		return ""
	}
	n := len(words)
	if n > 6 {
		n = 6
	}
	return strings.Join(words[:n], " ")
}

func freshness(value, fallback string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "day", "d", "d1":
		return "d1"
	case "week", "w", "w1":
		return "w1"
	case "month", "m", "m1":
		return "m1"
	case "year", "y", "y1":
		return "y1"
	case "any", "none", "":
		return fallback
	default:
		return value
	}
}

func variants(goal, style string, sites []string, n int) []string {
	if n <= 0 {
		return nil
	}
	var out []string
	switch style {
	case "papers":
		out = append(out, strings.TrimSpace(goal+" filetype:pdf"))
	case "news":
		out = append(out, strings.TrimSpace(goal+` "breaking" OR update`))
	case "docs":
		out = append(out, strings.TrimSpace(goal+" official documentation"))
	default:
		out = append(out, quote(firstPhrase(goal)))
	}
	if len(sites) > 0 {
		site := strings.TrimPrefix(strings.TrimSpace(sites[0]), "site:")
		if site != "" && !strings.Contains(out[0], "site:") {
			out[0] = strings.TrimSpace(out[0] + " site:" + site)
		}
	}
	if n > 1 {
		out = append(out, strings.TrimSpace(goal+" overview OR explained"))
	}
	if len(out) > n {
		out = out[:n]
	}
	return out
}

func omitEmpty(in map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range in {
		if strings.TrimSpace(v) != "" {
			out[k] = v
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
