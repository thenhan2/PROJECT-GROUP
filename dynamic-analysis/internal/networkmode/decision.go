package networkmode

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
)

// Action represents the decision action
type Action string

const (
	// ActionForward - Forward request to real destination
	ActionForward Action = "forward"
	// ActionBlock - Block the request
	ActionBlock Action = "block"
	// ActionModify - Modify and forward the request
	ActionModify Action = "modify"
	// ActionSimulate - Simulate response (use Full Mode path)
	ActionSimulate Action = "simulate"
)

// String returns string representation of Action
func (a Action) String() string {
	return string(a)
}

// IsValid checks if action is valid
func (a Action) IsValid() bool {
	return a == ActionForward || a == ActionBlock || a == ActionModify || a == ActionSimulate
}

// Decision represents a decision made by the engine
type Decision struct {
	// Action - The action to take
	Action Action `json:"action"`

	// Reason - Why this decision was made
	Reason string `json:"reason"`

	// RuleName - Which rule triggered this decision
	RuleName string `json:"rule_name,omitempty"`

	// Modifier - Optional modifier configuration
	Modifier *Modifier `json:"modifier,omitempty"`

	// Metadata - Additional metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Confidence - Confidence level (0.0 - 1.0)
	Confidence float64 `json:"confidence"`
}

// Modifier holds modification instructions
type Modifier struct {
	// Type - Type of modification
	Type string `json:"type"`

	// SaveOriginal - Save original content before modification
	SaveOriginal bool `json:"save_original"`

	// ReplaceWith - What to replace with
	ReplaceWith string `json:"replace_with,omitempty"`

	// StripHeaders - Headers to remove
	StripHeaders []string `json:"strip_headers,omitempty"`

	// InjectHeaders - Headers to add
	InjectHeaders map[string]string `json:"inject_headers,omitempty"`

	// LogFullContent - Log complete content
	LogFullContent bool `json:"log_full_content"`

	// StripPII - Strip personally identifiable information
	StripPII bool `json:"strip_pii"`
}

// DecisionRule represents a decision rule
type DecisionRule struct {
	// Name - Rule name
	Name string `json:"name" yaml:"name"`

	// Priority - Higher priority rules are evaluated first
	Priority int `json:"priority" yaml:"priority"`

	// Enabled - Is this rule enabled
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Condition - Rule condition
	Condition *RuleCondition `json:"condition" yaml:"condition"`

	// Action - Action to take if condition matches
	Action Action `json:"action" yaml:"action"`

	// Modifier - Optional modifier configuration
	Modifier *Modifier `json:"modifier,omitempty" yaml:"modifier,omitempty"`

	// Description - Rule description
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// RuleCondition represents a rule condition
type RuleCondition struct {
	// Type - Condition type
	Type ConditionType `json:"type" yaml:"type"`

	// Domain - Domain pattern (for domain_* conditions)
	Domain string `json:"domain,omitempty" yaml:"domain,omitempty"`

	// Domains - Multiple domains
	Domains []string `json:"domains,omitempty" yaml:"domains,omitempty"`

	// DomainPattern - Domain regex pattern
	DomainPattern string `json:"domain_pattern,omitempty" yaml:"domain_pattern,omitempty"`

	// Protocol - Protocol type (HTTP, HTTPS, DNS, etc.)
	Protocol string `json:"protocol,omitempty" yaml:"protocol,omitempty"`

	// Method - HTTP method
	Method string `json:"method,omitempty" yaml:"method,omitempty"`

	// Path - URL path pattern
	Path string `json:"path,omitempty" yaml:"path,omitempty"`

	// ContentType - Content-Type header value
	ContentType string `json:"content_type,omitempty" yaml:"content_type,omitempty"`

	// FileExtension - File extension
	FileExtension string `json:"file_extension,omitempty" yaml:"file_extension,omitempty"`

	// FileExtensions - Multiple file extensions
	FileExtensions []string `json:"file_extensions,omitempty" yaml:"file_extensions,omitempty"`

	// MinSize - Minimum size in bytes
	MinSize int64 `json:"min_size,omitempty" yaml:"min_size,omitempty"`

	// MaxSize - Maximum size in bytes
	MaxSize int64 `json:"max_size,omitempty" yaml:"max_size,omitempty"`

	// SourceFile - File containing list (for blacklist/whitelist)
	SourceFile string `json:"source_file,omitempty" yaml:"source_file,omitempty"`
}

// ConditionType represents the type of condition
type ConditionType string

const (
	ConditionDomainWhitelist ConditionType = "domain_whitelist"
	ConditionDomainBlacklist ConditionType = "domain_blacklist"
	ConditionDomainPattern   ConditionType = "domain_pattern"
	ConditionProtocol        ConditionType = "protocol"
	ConditionFileExtension   ConditionType = "file_extension"
	ConditionContentType     ConditionType = "content_type"
	ConditionMethod          ConditionType = "method"
	ConditionUploadDetection ConditionType = "upload_detection"
	ConditionDefault         ConditionType = "default"
)

// DecisionEngine makes decisions about how to handle traffic
type DecisionEngine struct {
	rules      []DecisionRule
	whitelist  map[string]bool
	blacklist  map[string]bool
	ruleCache  map[string]*Decision
	config     *HalfModeConfig
	logger     *slog.Logger
}

// NewDecisionEngine creates a new decision engine
func NewDecisionEngine(config *HalfModeConfig, logger *slog.Logger) *DecisionEngine {
	if logger == nil {
		logger = slog.Default()
	}

	return &DecisionEngine{
		rules:      []DecisionRule{},
		whitelist:  make(map[string]bool),
		blacklist:  make(map[string]bool),
		ruleCache:  make(map[string]*Decision),
		config:     config,
		logger:     logger,
	}
}

// AddRule adds a decision rule
func (de *DecisionEngine) AddRule(rule DecisionRule) error {
	if !rule.Action.IsValid() {
		return ErrInvalidAction
	}
	de.rules = append(de.rules, rule)
	de.sortRulesByPriority()
	return nil
}

// AddRules adds multiple decision rules
func (de *DecisionEngine) AddRules(rules []DecisionRule) error {
	for _, rule := range rules {
		if err := de.AddRule(rule); err != nil {
			return fmt.Errorf("failed to add rule %s: %w", rule.Name, err)
		}
	}
	return nil
}

// sortRulesByPriority sorts rules by priority (higher first)
func (de *DecisionEngine) sortRulesByPriority() {
	// Simple bubble sort (good enough for small rule sets)
	for i := 0; i < len(de.rules); i++ {
		for j := i + 1; j < len(de.rules); j++ {
			if de.rules[j].Priority > de.rules[i].Priority {
				de.rules[i], de.rules[j] = de.rules[j], de.rules[i]
			}
		}
	}
}

// Decide makes a decision for the given request
func (de *DecisionEngine) Decide(ctx context.Context, req *Request) (*Decision, error) {
	// Check cache first
	cacheKey := de.getCacheKey(req)
	if cached, ok := de.ruleCache[cacheKey]; ok {
		de.logger.DebugContext(ctx, "Using cached decision", "key", cacheKey)
		return cached, nil
	}

	// Evaluate rules in priority order
	for _, rule := range de.rules {
		if !rule.Enabled {
			continue
		}

		matches, err := de.evaluateCondition(ctx, req, rule.Condition)
		if err != nil {
			de.logger.WarnContext(ctx, "Failed to evaluate rule condition",
				"rule", rule.Name,
				"error", err)
			continue
		}

		if matches {
			decision := &Decision{
				Action:     rule.Action,
				Reason:     fmt.Sprintf("Matched rule: %s", rule.Name),
				RuleName:   rule.Name,
				Modifier:   rule.Modifier,
				Confidence: 1.0,
				Metadata: map[string]interface{}{
					"rule_priority": rule.Priority,
				},
			}

			de.logger.InfoContext(ctx, "Decision made",
				"rule", rule.Name,
				"action", decision.Action,
				"domain", req.Domain,
				"protocol", req.Protocol)

			// Cache the decision
			de.ruleCache[cacheKey] = decision

			return decision, nil
		}
	}

	// No rule matched, use default action
	decision := &Decision{
		Action:     de.config.DefaultAction,
		Reason:     "No matching rule, using default action",
		RuleName:   "default",
		Confidence: 0.5,
		Metadata:   map[string]interface{}{},
	}

	de.logger.InfoContext(ctx, "Using default action",
		"action", decision.Action,
		"domain", req.Domain)

	return decision, nil
}

// evaluateCondition evaluates a rule condition
func (de *DecisionEngine) evaluateCondition(ctx context.Context, req *Request, cond *RuleCondition) (bool, error) {
	if cond == nil {
		return false, fmt.Errorf("nil condition")
	}

	switch cond.Type {
	case ConditionDomainWhitelist:
		return de.matchDomainList(req.Domain, cond.Domains, true), nil

	case ConditionDomainBlacklist:
		return de.matchDomainList(req.Domain, cond.Domains, false), nil

	case ConditionDomainPattern:
		return de.matchDomainPattern(req.Domain, cond.DomainPattern)

	case ConditionProtocol:
		return strings.EqualFold(req.Protocol, cond.Protocol), nil

	case ConditionFileExtension:
		return de.matchFileExtension(req.Path, cond.FileExtensions), nil

	case ConditionContentType:
		return de.matchContentType(req.Headers, cond.ContentType), nil

	case ConditionMethod:
		return strings.EqualFold(req.Method, cond.Method), nil

	case ConditionUploadDetection:
		return de.detectUpload(req, cond), nil

	case ConditionDefault:
		return true, nil

	default:
		return false, fmt.Errorf("unknown condition type: %s", cond.Type)
	}
}

// matchDomainList checks if domain matches any in the list
func (de *DecisionEngine) matchDomainList(domain string, domains []string, isWhitelist bool) bool {
	domain = strings.ToLower(domain)
	
	for _, pattern := range domains {
		pattern = strings.ToLower(pattern)
		
		// Exact match
		if domain == pattern {
			return true
		}
		
		// Wildcard match (*.example.com)
		if strings.HasPrefix(pattern, "*.") {
			suffix := pattern[2:] // Remove "*."
			if strings.HasSuffix(domain, suffix) || domain == suffix {
				return true
			}
		}
	}
	
	return false
}

// matchDomainPattern checks if domain matches regex pattern
func (de *DecisionEngine) matchDomainPattern(domain string, pattern string) (bool, error) {
	if pattern == "" {
		return false, nil
	}
	
	matched, err := regexp.MatchString(pattern, domain)
	if err != nil {
		return false, fmt.Errorf("invalid regex pattern: %w", err)
	}
	
	return matched, nil
}

// matchFileExtension checks if path has matching file extension
func (de *DecisionEngine) matchFileExtension(path string, extensions []string) bool {
	path = strings.ToLower(path)
	
	for _, ext := range extensions {
		ext = strings.ToLower(ext)
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	
	return false
}

// matchContentType checks if Content-Type header matches
func (de *DecisionEngine) matchContentType(headers map[string]string, contentType string) bool {
	if contentType == "" {
		return false
	}
	
	for key, value := range headers {
		if strings.EqualFold(key, "content-type") {
			return strings.Contains(strings.ToLower(value), strings.ToLower(contentType))
		}
	}
	
	return false
}

// detectUpload detects if request is an upload
func (de *DecisionEngine) detectUpload(req *Request, cond *RuleCondition) bool {
	// Check method
	if cond.Method != "" {
		if !strings.EqualFold(req.Method, cond.Method) {
			return false
		}
	}
	
	// Check size
	if cond.MinSize > 0 {
		if req.ContentLength < cond.MinSize {
			return false
		}
	}
	
	return true
}

// getCacheKey generates a cache key for the request
func (de *DecisionEngine) getCacheKey(req *Request) string {
	return fmt.Sprintf("%s:%s:%s", req.Protocol, req.Domain, req.Path)
}

// GetRules returns all rules
func (de *DecisionEngine) GetRules() []DecisionRule {
	return de.rules
}

// ClearCache clears the decision cache
func (de *DecisionEngine) ClearCache() {
	de.ruleCache = make(map[string]*Decision)
}

// DefaultRules returns a set of safe default rules
func DefaultRules() []DecisionRule {
	return []DecisionRule{
		{
			Name:     "block_known_c2",
			Priority: 100,
			Enabled:  true,
			Condition: &RuleCondition{
				Type: ConditionDomainBlacklist,
				Domains: []string{
					"*.malware-c2.com",
					"*.evil-domain.net",
					"192.168.1.100",
				},
			},
			Action:      ActionBlock,
			Description: "Block known C2 servers",
		},
		{
			Name:     "allow_legitimate_cdns",
			Priority: 90,
			Enabled:  true,
			Condition: &RuleCondition{
				Type: ConditionDomainWhitelist,
				Domains: []string{
					"*.cloudflare.com",
					"*.akamai.com",
					"*.fastly.com",
				},
			},
			Action:      ActionForward,
			Description: "Allow legitimate CDNs",
		},
		{
			Name:     "intercept_executables",
			Priority: 80,
			Enabled:  true,
			Condition: &RuleCondition{
				Type: ConditionFileExtension,
				FileExtensions: []string{
					".exe", ".dll", ".ps1", ".sh", ".bat", ".cmd",
				},
			},
			Action: ActionModify,
			Modifier: &Modifier{
				Type:         "sandbox_executable",
				SaveOriginal: true,
				ReplaceWith:  "honeypot",
			},
			Description: "Intercept and sandbox executable downloads",
		},
		{
			Name:     "monitor_data_exfiltration",
			Priority: 70,
			Enabled:  true,
			Condition: &RuleCondition{
				Type:    ConditionUploadDetection,
				Method:  "POST",
				MinSize: 1024 * 1024, // 1MB
			},
			Action: ActionModify,
			Modifier: &Modifier{
				Type:           "content_logging",
				LogFullContent: true,
				StripPII:       true,
			},
			Description: "Monitor large POST requests for data exfiltration",
		},
		{
			Name:     "default_simulate",
			Priority: 1,
			Enabled:  true,
			Condition: &RuleCondition{
				Type: ConditionDefault,
			},
			Action:      ActionSimulate,
			Description: "Default action - simulate all unmatched traffic",
		},
	}
}
