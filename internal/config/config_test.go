package config

import "testing"

func TestLoadRequiresCSEID(t *testing.T) {
	if _, err := Load(func(string) string { return "" }); err == nil {
		t.Fatal("accepted empty config")
	}
}

func TestLoadDoesNotRequireAPIKey(t *testing.T) {
	cfg, err := Load(func(name string) string {
		if name == "GOOGLE_CSE_ID" {
			return "cx"
		}
		return ""
	})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != "8080" || cfg.MCPPath != "/mcp" || cfg.MaxResults != 5 || cfg.MaxParallel != 3 {
		t.Fatalf("defaults = %+v", cfg)
	}
	if cfg.CSEID != "cx" {
		t.Fatalf("cseid = %q", cfg.CSEID)
	}
}

func TestLoadRejectsBadMCPPath(t *testing.T) {
	if _, err := Load(func(name string) string {
		switch name {
		case "GOOGLE_CSE_ID":
			return "cx"
		case "MCP_PATH":
			return "mcp"
		default:
			return ""
		}
	}); err == nil {
		t.Fatal("accepted relative MCP_PATH")
	}
}
