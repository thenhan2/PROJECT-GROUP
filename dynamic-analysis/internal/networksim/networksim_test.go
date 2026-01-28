package networksim

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	
	if cfg.INetSimDNSAddr != "172.20.0.2:53" {
		t.Errorf("Expected DNS 172.20.0.2:53, got %s", cfg.INetSimDNSAddr)
	}
	
	if cfg.INetSimHTTPAddr != "172.20.0.2:80" {
		t.Errorf("Expected HTTP 172.20.0.2:80, got %s", cfg.INetSimHTTPAddr)
	}
	
	if cfg.Enabled {
		t.Error("Expected disabled by default")
	}
}

func TestIsURLAlive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.LivenessTimeout = 2 * time.Second
	ns := New(cfg)
	ctx := context.Background()

	// Test alive URL
	if !ns.IsURLAlive(ctx, server.URL) {
		t.Error("Alive URL should return true")
	}

	// Test dead URL
	if ns.IsURLAlive(ctx, "http://dead-url-12345.com") {
		t.Error("Dead URL should return false")
	}
}

func TestShouldRedirectToINetSim(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.LivenessTimeout = 2 * time.Second
	ns := New(cfg)
	ctx := context.Background()

	// Alive URL - no redirect
	if ns.ShouldRedirectToINetSim(ctx, server.URL) {
		t.Error("Alive URL should not redirect")
	}

	// Dead URL - should redirect
	if !ns.ShouldRedirectToINetSim(ctx, "http://dead-url.com") {
		t.Error("Dead URL should redirect to INetSim")
	}
}

func TestGetDNSServers(t *testing.T) {
	ns := New(DefaultConfig())
	
	// Disabled - use Google DNS
	dns := ns.GetDNSServers()
	if len(dns) != 2 || dns[0] != "8.8.8.8" {
		t.Errorf("Expected Google DNS when disabled, got %v", dns)
	}

	// Enabled - use INetSim
	ns.config.Enabled = true
	dns = ns.GetDNSServers()
	if len(dns) != 1 || dns[0] != "172.20.0.2" {
		t.Errorf("Expected INetSim DNS when enabled, got %v", dns)
	}
}
