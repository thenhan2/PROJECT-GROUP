package networkmode

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TrafficModifier modifies network traffic based on rules
type TrafficModifier struct {
	config *TrafficModifierConfig
	logger *slog.Logger
}

// NewModifier creates a new traffic modifier
func NewModifier(config *TrafficModifierConfig, logger *slog.Logger) *TrafficModifier {
	if logger == nil {
		logger = slog.Default()
	}

	return &TrafficModifier{
		config: config,
		logger: logger,
	}
}

// ModifyRequest modifies a request based on the modifier configuration
func (tm *TrafficModifier) ModifyRequest(ctx context.Context, req *Request, modifier *Modifier) (*Request, error) {
	if !tm.config.Enabled || modifier == nil {
		return req, nil
	}

	modifiedReq := *req // Copy request

	// Strip headers
	if len(modifier.StripHeaders) > 0 {
		for _, header := range modifier.StripHeaders {
			delete(modifiedReq.Headers, header)
		}
		tm.logger.InfoContext(ctx, "Stripped headers from request",
			"headers", modifier.StripHeaders,
			"req_id", req.ID)
	}

	// Inject headers
	if len(modifier.InjectHeaders) > 0 {
		for key, value := range modifier.InjectHeaders {
			modifiedReq.Headers[key] = value
		}
		tm.logger.InfoContext(ctx, "Injected headers into request",
			"headers", modifier.InjectHeaders,
			"req_id", req.ID)
	}

	// Strip auth headers globally
	if tm.config.StripAuthHeaders {
		authHeaders := []string{"Authorization", "Cookie", "X-Auth-Token", "X-API-Key"}
		for _, header := range authHeaders {
			delete(modifiedReq.Headers, header)
		}
	}

	// Inject tracking headers globally
	if tm.config.InjectTrackingHeaders {
		modifiedReq.Headers["X-Pack-A-Mal-Analysis"] = "true"
		modifiedReq.Headers["X-Pack-A-Mal-Request-ID"] = req.ID
		modifiedReq.Headers["X-Pack-A-Mal-Timestamp"] = req.Timestamp.Format(time.RFC3339)
	}

	return &modifiedReq, nil
}

// ModifyResponse modifies a response based on the modifier configuration
func (tm *TrafficModifier) ModifyResponse(ctx context.Context, resp *Response, req *Request, modifier *Modifier) (*Response, error) {
	if !tm.config.Enabled || modifier == nil {
		return resp, nil
	}

	modifiedResp := *resp // Copy response

	// Limit response size
	if tm.config.MaxResponseSize > 0 && resp.ContentLength > tm.config.MaxResponseSize {
		tm.logger.WarnContext(ctx, "Response size exceeds limit, truncating",
			"original_size", resp.ContentLength,
			"max_size", tm.config.MaxResponseSize,
			"req_id", req.ID)
		modifiedResp.Body = resp.Body[:tm.config.MaxResponseSize]
		modifiedResp.ContentLength = tm.config.MaxResponseSize
	}

	// Handle executable downloads
	if modifier.Type == "sandbox_executable" && tm.config.SandboxExecutables {
		return tm.handleExecutableDownload(ctx, resp, req, modifier)
	}

	// Log full content if requested
	if modifier.LogFullContent {
		tm.logger.InfoContext(ctx, "Full response content",
			"req_id", req.ID,
			"content_length", resp.ContentLength,
			"body", string(resp.Body))
	}

	// Strip PII if requested
	if modifier.StripPII {
		modifiedResp.Body = tm.stripPII(resp.Body)
		modifiedResp.ContentLength = int64(len(modifiedResp.Body))
	}

	return &modifiedResp, nil
}

// handleExecutableDownload handles executable file downloads
func (tm *TrafficModifier) handleExecutableDownload(ctx context.Context, resp *Response, req *Request, modifier *Modifier) (*Response, error) {
	tm.logger.WarnContext(ctx, "Executable download detected",
		"domain", req.Domain,
		"path", req.Path,
		"size", resp.ContentLength)

	// Save original executable
	if modifier.SaveOriginal {
		if err := tm.saveExecutable(ctx, req, resp); err != nil {
			tm.logger.ErrorContext(ctx, "Failed to save executable",
				"error", err,
				"req_id", req.ID)
		}
	}

	// Return honeypot or fake response
	honeypotResp := &Response{
		ID:            resp.ID,
		Timestamp:     resp.Timestamp,
		StatusCode:    200,
		Headers:       make(map[string]string),
		Source:        "sandboxed",
		Decision:      resp.Decision,
		ContentLength: 0,
	}

	// Create fake executable (harmless placeholder)
	fakeContent := tm.createFakeExecutable(req)
	honeypotResp.Body = fakeContent
	honeypotResp.ContentLength = int64(len(fakeContent))

	// Copy headers
	for k, v := range(resp.Headers) {
		honeypotResp.Headers[k] = v
	}
	honeypotResp.Headers["X-Pack-A-Mal-Sandboxed"] = "true"
	honeypotResp.Headers["X-Pack-A-Mal-Original-Size"] = fmt.Sprintf("%d", resp.ContentLength)
	honeypotResp.Headers["Content-Length"] = fmt.Sprintf("%d", honeypotResp.ContentLength)

	tm.logger.InfoContext(ctx, "Executable sandboxed",
		"req_id", req.ID,
		"original_size", resp.ContentLength,
		"fake_size", honeypotResp.ContentLength)

	return honeypotResp, nil
}

// saveExecutable saves the executable to disk for analysis
func (tm *TrafficModifier) saveExecutable(ctx context.Context, req *Request, resp *Response) error {
	// Ensure sandbox directory exists
	if err := os.MkdirAll(tm.config.SandboxDir, 0755); err != nil {
		return fmt.Errorf("failed to create sandbox dir: %w", err)
	}

	// Generate filename
	hash := sha256.Sum256(resp.Body)
	hashStr := fmt.Sprintf("%x", hash[:8]) // First 8 bytes of hash
	filename := filepath.Base(req.Path)
	if filename == "" || filename == "." || filename == "/" {
		filename = "download.bin"
	}
	safeFilename := fmt.Sprintf("%s_%s", hashStr, filename)
	fullPath := filepath.Join(tm.config.SandboxDir, safeFilename)

	// Write executable
	if err := os.WriteFile(fullPath, resp.Body, 0644); err != nil {
		return fmt.Errorf("failed to write executable: %w", err)
	}

	// Write metadata
	metadataPath := fullPath + ".metadata.json"
	metadata := fmt.Sprintf(`{
  "request_id": "%s",
  "timestamp": "%s",
  "domain": "%s",
  "path": "%s",
  "size": %d,
  "sha256": "%x",
  "headers": %v
}`, req.ID, req.Timestamp.Format(time.RFC3339), req.Domain, req.Path,
		resp.ContentLength, hash, formatHeaders(req.Headers))

	if err := os.WriteFile(metadataPath, []byte(metadata), 0644); err != nil {
		tm.logger.WarnContext(ctx, "Failed to write metadata",
			"error", err,
			"path", metadataPath)
	}

	tm.logger.InfoContext(ctx, "Executable saved",
		"path", fullPath,
		"size", resp.ContentLength,
		"req_id", req.ID)

	return nil
}

// createFakeExecutable creates a harmless fake executable
func (tm *TrafficModifier) createFakeExecutable(req *Request) []byte {
	// Return a small harmless file with metadata
	content := fmt.Sprintf(`# Pack-A-Mal Sandbox File
# This is a placeholder for security analysis
# Original request: %s
# Domain: %s
# Path: %s
# Timestamp: %s
# Request ID: %s
`,
		req.Method,
		req.Domain,
		req.Path,
		req.Timestamp.Format(time.RFC3339),
		req.ID)

	return []byte(content)
}

// stripPII removes personally identifiable information from data
func (tm *TrafficModifier) stripPII(data []byte) []byte {
	// This is a simplified PII stripper
	// In production, use a proper PII detection/redaction library
	
	content := string(data)
	
	// Redact common PII patterns (basic implementation)
	patterns := map[string]string{
		"password": "[REDACTED_PASSWORD]",
		"passwd":   "[REDACTED_PASSWORD]",
		"pwd":      "[REDACTED_PASSWORD]",
		"token":    "[REDACTED_TOKEN]",
		"key":      "[REDACTED_KEY]",
		"secret":   "[REDACTED_SECRET]",
	}
	
	for pattern, replacement := range patterns {
		content = strings.ReplaceAll(strings.ToLower(content), pattern, replacement)
	}
	
	return []byte(content)
}

// formatHeaders formats headers for JSON output
func formatHeaders(headers map[string]string) string {
	if len(headers) == 0 {
		return "{}"
	}
	
	var builder strings.Builder
	builder.WriteString("{")
	first := true
	for k, v := range headers {
		if !first {
			builder.WriteString(", ")
		}
		builder.WriteString(fmt.Sprintf(`"%s": "%s"`, k, v))
		first = false
	}
	builder.WriteString("}")
	return builder.String()
}

// SandboxFile saves a file to the sandbox directory
func (tm *TrafficModifier) SandboxFile(ctx context.Context, filename string, content []byte, metadata map[string]interface{}) error {
	if err := os.MkdirAll(tm.config.SandboxDir, 0755); err != nil {
		return fmt.Errorf("failed to create sandbox dir: %w", err)
	}

	fullPath := filepath.Join(tm.config.SandboxDir, filename)
	
	if err := os.WriteFile(fullPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	tm.logger.InfoContext(ctx, "File sandboxed",
		"path", fullPath,
		"size", len(content))

	return nil
}

// ReadFile reads a file from disk
func ReadFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}
