package networkmode

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// NetworkLogger handles logging for network mode operations
type NetworkLogger struct {
	config     *LoggingConfig
	logger     *slog.Logger
	trafficLog *os.File
	decisionLog *os.File
	mu         sync.Mutex
}

// NewLogger creates a new network logger
func NewLogger(config *LoggingConfig, logger *slog.Logger) (*NetworkLogger, error) {
	if logger == nil {
		logger = slog.Default()
	}

	nl := &NetworkLogger{
		config: config,
		logger: logger,
	}

	// Open traffic log file
	if config.LogAllRequests && config.TrafficLogFile != "" {
		// Create directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(config.TrafficLogFile), 0755); err != nil {
			logger.Warn("Failed to create log directory, disabling traffic log",
				"error", err,
				"path", config.TrafficLogFile)
		} else {
			file, err := os.OpenFile(config.TrafficLogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				logger.Warn("Failed to open traffic log file, disabling traffic log",
					"error", err,
					"path", config.TrafficLogFile)
			} else {
				nl.trafficLog = file
			}
		}
	}

	// Open decisions log file
	if config.LogDecisions && config.DecisionsFile != "" {
		// Create directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(config.DecisionsFile), 0755); err != nil {
			logger.Warn("Failed to create log directory, disabling decisions log",
				"error", err,
				"path", config.DecisionsFile)
		} else {
			file, err := os.OpenFile(config.DecisionsFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				logger.Warn("Failed to open decisions log file, disabling decisions log",
					"error", err,
					"path", config.DecisionsFile)
			} else {
				nl.decisionLog = file
			}
		}
	}

	return nl, nil
}

// LogTraffic logs a traffic event
func (nl *NetworkLogger) LogTraffic(ctx context.Context, entry *TrafficLog) error {
	if !nl.config.LogAllRequests {
		return nil
	}

	nl.mu.Lock()
	defer nl.mu.Unlock()

	// Log to structured logger
	nl.logger.InfoContext(ctx, "Traffic event",
		"req_id", entry.Request.ID,
		"protocol", entry.Request.Protocol,
		"domain", entry.Request.Domain,
		"path", entry.Request.Path,
		"action", entry.Action,
		"decision", entry.Decision.Action,
		"rule", entry.Decision.RuleName,
		"duration", entry.Duration,
		"error", entry.Error)

	// Write to traffic log file
	if nl.trafficLog != nil {
		jsonData, err := json.Marshal(entry)
		if err != nil {
			return fmt.Errorf("failed to marshal traffic log: %w", err)
		}
		if _, err := nl.trafficLog.Write(append(jsonData, '\n')); err != nil {
			return fmt.Errorf("failed to write traffic log: %w", err)
		}
	}

	return nil
}

// LogDecision logs a decision
func (nl *NetworkLogger) LogDecision(ctx context.Context, req *Request, decision *Decision) error {
	if !nl.config.LogDecisions {
		return nil
	}

	nl.mu.Lock()
	defer nl.mu.Unlock()

	logEntry := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"req_id":    req.ID,
		"protocol":  req.Protocol,
		"domain":    req.Domain,
		"path":      req.Path,
		"decision": map[string]interface{}{
			"action":     decision.Action,
			"reason":     decision.Reason,
			"rule_name":  decision.RuleName,
			"confidence": decision.Confidence,
		},
	}

	// Log to structured logger
	nl.logger.InfoContext(ctx, "Decision made",
		"req_id", req.ID,
		"action", decision.Action,
		"rule", decision.RuleName,
		"reason", decision.Reason)

	// Write to decisions log file
	if nl.decisionLog != nil {
		jsonData, err := json.Marshal(logEntry)
		if err != nil {
			return fmt.Errorf("failed to marshal decision log: %w", err)
		}
		if _, err := nl.decisionLog.Write(append(jsonData, '\n')); err != nil {
			return fmt.Errorf("failed to write decision log: %w", err)
		}
	}

	return nil
}

// LogModification logs a traffic modification
func (nl *NetworkLogger) LogModification(ctx context.Context, req *Request, modification string) error {
	if !nl.config.LogModifications {
		return nil
	}

	nl.logger.InfoContext(ctx, "Traffic modification",
		"req_id", req.ID,
		"modification", modification,
		"domain", req.Domain)

	return nil
}

// LogRequest logs a request
func (nl *NetworkLogger) LogRequest(ctx context.Context, req *Request) error {
	if !nl.config.LogAllRequests {
		return nil
	}

	nl.logger.DebugContext(ctx, "Request",
		"req_id", req.ID,
		"protocol", req.Protocol,
		"method", req.Method,
		"domain", req.Domain,
		"path", req.Path,
		"content_length", req.ContentLength)

	return nil
}

// LogResponse logs a response
func (nl *NetworkLogger) LogResponse(ctx context.Context, resp *Response) error {
	if !nl.config.LogResponses {
		return nil
	}

	nl.logger.DebugContext(ctx, "Response",
		"resp_id", resp.ID,
		"status_code", resp.StatusCode,
		"content_length", resp.ContentLength,
		"source", resp.Source)

	return nil
}

// LogError logs an error
func (nl *NetworkLogger) LogError(ctx context.Context, req *Request, err error) {
	nl.logger.ErrorContext(ctx, "Network error",
		"req_id", req.ID,
		"domain", req.Domain,
		"error", err)
}

// Close closes the logger
func (nl *NetworkLogger) Close() error {
	nl.mu.Lock()
	defer nl.mu.Unlock()

	var errs []error

	if nl.trafficLog != nil {
		if err := nl.trafficLog.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if nl.decisionLog != nil {
		if err := nl.decisionLog.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing logger: %v", errs)
	}

	return nil
}
