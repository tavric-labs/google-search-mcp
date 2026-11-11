package config

import "testing"

func TestLoadAppliesDefaultsWithoutAPIKeyOrCSE(t *testing.T) {
	cfg, err := Load(func(string) string { return "" })
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != "8080" || cfg.MCPPath != "/mcp" || cfg.MaxResults != 5 || cfg.MaxParallel != 3 {
		t.Fatalf("defaults = %+v", cfg)
	}
	if cfg.Model != "gemini-2.5-flash" {
		t.Fatalf("model = %q", cfg.Model)
	}
	if !stringsHasSuffix(cfg.APIBase, "/v1beta") {
		t.Fatalf("api base = %q", cfg.APIBase)
	}
}

func TestLoadRejectsBadMCPPath(t *testing.T) {
	if _, err := Load(func(name string) string {
		if name == "MCP_PATH" {
			return "mcp"
		}
		return ""
	}); err == nil {
		t.Fatal("accepted relative MCP_PATH")
	}
}

func stringsHasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
