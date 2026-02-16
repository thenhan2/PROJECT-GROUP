package networkmode

import (
	"time"
)

// Protocol represents network protocol
type Protocol string

const (
	ProtocolHTTP  Protocol = "HTTP"
	ProtocolHTTPS Protocol = "HTTPS"
	ProtocolDNS   Protocol = "DNS"
	ProtocolSMTP  Protocol = "SMTP"
	ProtocolFTP   Protocol = "FTP"
	ProtocolTCP   Protocol = "TCP"
	ProtocolUDP   Protocol = "UDP"
)

// Request represents a network request
type Request struct {
	// ID - Unique request ID
	ID string `json:"id"`

	// Timestamp - Request timestamp
	Timestamp time.Time `json:"timestamp"`

	// Protocol - Network protocol
	Protocol string `json:"protocol"`

	// Method - HTTP method (for HTTP/HTTPS)
	Method string `json:"method,omitempty"`

	// Domain - Destination domain
	Domain string `json:"domain"`

	// IP - Destination IP address
	IP string `json:"ip,omitempty"`

	// Port - Destination port
	Port int `json:"port"`

	// Path - URL path (for HTTP/HTTPS)
	Path string `json:"path,omitempty"`

	// Query - Query parameters
	Query map[string]string `json:"query,omitempty"`

	// Headers - Request headers
	Headers map[string]string `json:"headers,omitempty"`

	// Body - Request body
	Body []byte `json:"body,omitempty"`

	// ContentLength - Content length
	ContentLength int64 `json:"content_length"`

	// SourceIP - Source IP address
	SourceIP string `json:"source_ip"`

	// SourcePort - Source port
	SourcePort int `json:"source_port"`
}

// Response represents a network response
type Response struct {
	// ID - Unique response ID (matches Request.ID)
	ID string `json:"id"`

	// Timestamp - Response timestamp
	Timestamp time.Time `json:"timestamp"`

	// StatusCode - HTTP status code (for HTTP/HTTPS)
	StatusCode int `json:"status_code,omitempty"`

	// Headers - Response headers
	Headers map[string]string `json:"headers,omitempty"`

	// Body - Response body
	Body []byte `json:"body,omitempty"`

	// ContentLength - Content length
	ContentLength int64 `json:"content_length"`

	// Source - Response source (real, simulated, modified)
	Source string `json:"source"`

	// Decision - Associated decision
	Decision *Decision `json:"decision,omitempty"`

	// Metadata - Additional metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TrafficLog represents a traffic log entry
type TrafficLog struct {
	// Timestamp - Log timestamp
	Timestamp time.Time `json:"timestamp"`

	// Request - The request
	Request *Request `json:"request"`

	// Response - The response
	Response *Response `json:"response,omitempty"`

	// Decision - Decision made
	Decision *Decision `json:"decision"`

	// Action - Action taken
	Action Action `json:"action"`

	// Modifications - Modifications applied
	Modifications []string `json:"modifications,omitempty"`

	// Error - Error if any
	Error string `json:"error,omitempty"`

	// Duration - Request duration
	Duration time.Duration `json:"duration"`
}
