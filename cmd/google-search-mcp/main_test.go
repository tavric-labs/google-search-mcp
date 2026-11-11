package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tavric-labs/google-search-mcp/internal/config"
)

func TestHeaderAuthRequiresKey(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	handler := headerAuth("/mcp", next)

	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestHeaderAuthCheckAndMCP(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	handler := headerAuth("/mcp", next)

	check := httptest.NewRequest(http.MethodGet, "/mcp/auth/check", nil)
	check.Header.Set(config.APIKeyHeader, "key")
	checkRec := httptest.NewRecorder()
	handler.ServeHTTP(checkRec, check)
	if checkRec.Code != http.StatusNoContent || called {
		t.Fatalf("check status=%d called=%v", checkRec.Code, called)
	}

	mcpReq := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	mcpReq.Header.Set(config.APIKeyHeader, "key")
	mcpRec := httptest.NewRecorder()
	handler.ServeHTTP(mcpRec, mcpReq)
	if mcpRec.Code != http.StatusOK || !called {
		t.Fatalf("mcp status=%d called=%v", mcpRec.Code, called)
	}
}
