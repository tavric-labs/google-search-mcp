package config

import "testing"

func TestLoadRequiresGoogleCredentials(t *testing.T) {
	if _, err := Load(func(string) string { return "" }); err == nil {
		t.Fatal("accepted empty config")
	}
}

func TestLoadAppliesDefaults(t *testing.T) {
	cfg, err := Load(func(name string) string {
		switch name {
		case "GOOGLE_API_KEY":
			return "key"
		case "GOOGLE_CSE_ID":
			return "cx"
		default:
			return ""
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != "8080" || cfg.MCPPath != "/mcp" || cfg.MaxResults != 5 || cfg.MaxParallel != 3 {
		t.Fatalf("defaults = %+v", cfg)
	}
}
