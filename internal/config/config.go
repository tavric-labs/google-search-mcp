// Package config reads the server's settings from the environment.
package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// APIKeyHeader is the credential header clients send with their Gemini API key.
const APIKeyHeader = "X-Gemini-Api-Key"

// Config is everything the server needs to start.
// End users supply a Gemini AI Studio API key via APIKeyHeader.
type Config struct {
	Port, MCPPath string
	APIBase       string
	Model         string
	MaxResults    int
	MaxParallel   int
	MaxOpenChars  int
	MaxPackChars  int
}

// Load reads and validates the configuration.
func Load(getenv func(string) string) (Config, error) {
	value := func(name, fallback string) string {
		if v := strings.TrimSpace(getenv(name)); v != "" {
			return v
		}
		return fallback
	}
	intValue := func(name string, fallback int) (int, error) {
		raw := strings.TrimSpace(getenv(name))
		if raw == "" {
			return fallback, nil
		}
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 {
			return 0, fmt.Errorf("%s must be a positive integer", name)
		}
		return n, nil
	}
	maxResults, err := intValue("GOOGLE_SEARCH_MAX_RESULTS", 5)
	if err != nil {
		return Config{}, err
	}
	if maxResults > 10 {
		maxResults = 10
	}
	maxParallel, err := intValue("GOOGLE_SEARCH_MAX_PARALLEL", 3)
	if err != nil {
		return Config{}, err
	}
	if maxParallel > 5 {
		maxParallel = 5
	}
	maxOpen, err := intValue("GOOGLE_SEARCH_MAX_OPEN_CHARS", 3000)
	if err != nil {
		return Config{}, err
	}
	maxPack, err := intValue("GOOGLE_SEARCH_MAX_PACK_CHARS", 3200)
	if err != nil {
		return Config{}, err
	}
	config := Config{
		Port:         value("PORT", "8080"),
		MCPPath:      value("MCP_PATH", "/mcp"),
		APIBase:      strings.TrimRight(value("GEMINI_API_BASE", "https://generativelanguage.googleapis.com/v1beta"), "/"),
		Model:        value("GEMINI_MODEL", "gemini-2.5-flash"),
		MaxResults:   maxResults,
		MaxParallel:  maxParallel,
		MaxOpenChars: maxOpen,
		MaxPackChars: maxPack,
	}
	if !strings.HasPrefix(config.MCPPath, "/") || strings.ContainsAny(config.MCPPath, "?#") || strings.HasSuffix(config.MCPPath, "/") {
		return Config{}, errors.New("MCP_PATH must be an absolute path without trailing slash")
	}
	return config, nil
}
