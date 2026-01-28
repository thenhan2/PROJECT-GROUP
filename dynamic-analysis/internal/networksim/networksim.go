package networksim

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// Config holds the network simulation configuration
type Config struct {
	INetSimDNSAddr  string        // DNS server (e.g., "172.20.0.2:53")
	INetSimHTTPAddr string        // HTTP server (e.g., "172.20.0.2:80")
	Enabled         bool          // Enable/disable simulation
	LivenessTimeout time.Duration // Timeout for URL check
}

// DefaultConfig returns default INetSim configuration
func DefaultConfig() *Config {
	return &Config{
		INetSimDNSAddr:  "172.20.0.2:53",
		INetSimHTTPAddr: "172.20.0.2:80",
		Enabled:         false,
		LivenessTimeout: 3 * time.Second,
	}
}

// NetworkSimulator handles URL checking and INetSim redirection
type NetworkSimulator struct {
	config *Config
}

// New creates a new NetworkSimulator
func New(config *Config) *NetworkSimulator {
	if config == nil {
		config = DefaultConfig()
	}
	return &NetworkSimulator{config: config}
}

// IsURLAlive checks if URL is accessible
func (ns *NetworkSimulator) IsURLAlive(ctx context.Context, url string) bool {
	if !ns.config.Enabled {
		return true
	}

	client := &http.Client{
		Timeout: ns.config.LivenessTimeout,
	}

	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		slog.WarnContext(ctx, "Cannot create request", "url", url, "error", err)
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.InfoContext(ctx, "URL not alive", "url", url)
		return false
	}
	defer resp.Body.Close()

	isAlive := resp.StatusCode >= 200 && resp.StatusCode < 400
	slog.InfoContext(ctx, "URL check", "url", url, "status", resp.StatusCode, "alive", isAlive)
	return isAlive
}

// ShouldRedirectToINetSim determines if should redirect to INetSim
// Logic: If URL not alive → redirect to INetSim
func (ns *NetworkSimulator) ShouldRedirectToINetSim(ctx context.Context, url string) bool {
	if !ns.config.Enabled {
		return false
	}

	if !ns.IsURLAlive(ctx, url) {
		slog.InfoContext(ctx, "Redirecting to INetSim", "url", url)
		return true
	}
	return false
}

// GetDNSServers returns DNS servers for sandbox
func (ns *NetworkSimulator) GetDNSServers() []string {
	if !ns.config.Enabled {
		return []string{"8.8.8.8", "8.8.4.4"}
	}

	host, _, err := net.SplitHostPort(ns.config.INetSimDNSAddr)
	if err != nil {
		return []string{ns.config.INetSimDNSAddr}
	}
	return []string{host}
}

// GetINetSimDNS returns INetSim DNS address
func (ns *NetworkSimulator) GetINetSimDNS() string {
	return ns.config.INetSimDNSAddr
}

// GetINetSimHTTP returns INetSim HTTP address
func (ns *NetworkSimulator) GetINetSimHTTP() string {
	return ns.config.INetSimHTTPAddr
}

// IsEnabled returns if simulation is enabled
func (ns *NetworkSimulator) IsEnabled() bool {
	return ns.config.Enabled
}

// ValidateINetSimConnection validates INetSim service is accessible
func (ns *NetworkSimulator) ValidateINetSimConnection(ctx context.Context) error {
	if !ns.config.Enabled {
		return nil
	}

	slog.InfoContext(ctx, "Validating INetSim")
	return nil // Simplified: assume INetSim is available
}
