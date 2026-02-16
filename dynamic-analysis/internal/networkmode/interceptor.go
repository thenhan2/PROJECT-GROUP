package networkmode

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
)

// TrafficInterceptor intercepts and analyzes network traffic
type TrafficInterceptor struct {
	config *Config
	logger *slog.Logger
}

// NewInterceptor creates a new traffic interceptor
func NewInterceptor(config *Config, logger *slog.Logger) *TrafficInterceptor {
	if logger == nil {
		logger = slog.Default()
	}

	return &TrafficInterceptor{
		config: config,
		logger: logger,
	}
}

// InterceptPacket captures and parses a raw packet
func (ti *TrafficInterceptor) InterceptPacket(ctx context.Context, packetData []byte) (*Request, error) {
	packet := gopacket.NewPacket(packetData, layers.LayerTypeEthernet, gopacket.Default)

	req := &Request{
		ID:        generateRequestID(),
		Timestamp: time.Now(),
		Headers:   make(map[string]string),
		Query:     make(map[string]string),
	}

	// Parse network layer
	if netLayer := packet.NetworkLayer(); netLayer != nil {
		if ipLayer, ok := netLayer.(*layers.IPv4); ok {
			req.IP = ipLayer.DstIP.String()
			req.SourceIP = ipLayer.SrcIP.String()
		}
	}

	// Parse transport layer
	if transLayer := packet.TransportLayer(); transLayer != nil {
		switch layer := transLayer.(type) {
		case *layers.TCP:
			req.Port = int(layer.DstPort)
			req.SourcePort = int(layer.SrcPort)
			req.Protocol = string(ProtocolTCP)

		case *layers.UDP:
			req.Port = int(layer.DstPort)
			req.SourcePort = int(layer.SrcPort)
			req.Protocol = string(ProtocolUDP)
		}
	}

	// Parse application layer
	if appLayer := packet.ApplicationLayer(); appLayer != nil {
		req.Body = appLayer.Payload()
		req.ContentLength = int64(len(req.Body))

		// Try to identify protocol
		ti.identifyProtocol(req, appLayer.Payload())
	}

	// Parse DNS layer
	if dnsLayer := packet.Layer(layers.LayerTypeDNS); dnsLayer != nil {
		if dns, ok := dnsLayer.(*layers.DNS); ok {
			req.Protocol = string(ProtocolDNS)
			if len(dns.Questions) > 0 {
				req.Domain = string(dns.Questions[0].Name)
			}
		}
	}

	ti.logger.DebugContext(ctx, "Packet intercepted",
		"protocol", req.Protocol,
		"domain", req.Domain,
		"ip", req.IP,
		"port", req.Port)

	return req, nil
}

// identifyProtocol tries to identify the application protocol
func (ti *TrafficInterceptor) identifyProtocol(req *Request, payload []byte) {
	if len(payload) < 4 {
		return
	}

	// HTTP detection
	httpMethods := []string{"GET ", "POST", "PUT ", "DELE", "HEAD", "OPTI", "PATC"}
	for _, method := range httpMethods {
		if len(payload) >= len(method) && string(payload[:len(method)]) == method {
			if req.Port == 443 {
				req.Protocol = string(ProtocolHTTPS)
			} else {
				req.Protocol = string(ProtocolHTTP)
			}
			ti.parseHTTPRequest(req, payload)
			return
		}
	}

	// SMTP detection
	if req.Port == 25 || (len(payload) >= 4 && (string(payload[:4]) == "MAIL" || string(payload[:4]) == "EHLO")) {
		req.Protocol = string(ProtocolSMTP)
		return
	}

	// FTP detection
	if req.Port == 21 || (len(payload) >= 4 && (string(payload[:4]) == "USER" || string(payload[:4]) == "PASS")) {
		req.Protocol = string(ProtocolFTP)
		return
	}
}

// parseHTTPRequest parses HTTP request from payload
func (ti *TrafficInterceptor) parseHTTPRequest(req *Request, payload []byte) {
	// This is a simplified HTTP parser
	// In production, use a proper HTTP parser library
	lines := splitLines(payload)
	if len(lines) == 0 {
		return
	}

	// Parse request line
	parts := splitSpaces(lines[0])
	if len(parts) >= 2 {
		req.Method = parts[0]
		req.Path = parts[1]
	}

	// Parse headers
	for i := 1; i < len(lines); i++ {
		if len(lines[i]) == 0 {
			break
		}
		if colonIdx := indexOf(lines[i], ':'); colonIdx > 0 {
			key := string(lines[i][:colonIdx])
			value := string(lines[i][colonIdx+1:])
			req.Headers[key] = trimSpace(value)

			// Extract domain from Host header
			if key == "Host" {
				req.Domain = trimSpace(value)
			}
		}
	}
}

// Helper functions
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

func splitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i := 0; i < len(data); i++ {
		if data[i] == '\n' {
			end := i
			if end > 0 && data[end-1] == '\r' {
				end--
			}
			lines = append(lines, data[start:end])
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}

func splitSpaces(data []byte) []string {
	var parts []string
	start := 0
	for i := 0; i <= len(data); i++ {
		if i == len(data) || data[i] == ' ' {
			if i > start {
				parts = append(parts, string(data[start:i]))
			}
			start = i + 1
		}
	}
	return parts
}

func indexOf(data []byte, ch byte) int {
	for i, b := range data {
		if b == ch {
			return i
		}
	}
	return -1
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
