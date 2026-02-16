package networkmode

import (
	"strings"
	"testing"
)

func TestModifier_StripPII(t *testing.T) {
	config := &TrafficModifierConfig{
		Enabled: true,
	}
	modifier := NewModifier(config, nil)

	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:  "Strip password",
			input: "password=secret123",
			contains: []string{"[redacted_password]"},
		},
		{
			name:  "Strip token",
			input: "Authorization: token abc123",
			contains: []string{"[redacted_token]"},
		},
		{
			name:  "Strip secret",
			input: "secret=mysecret",
			contains: []string{"[redacted_secret]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := modifier.stripPII([]byte(tt.input))
			resultStr := string(result)
			
			for _, expected := range tt.contains {
				// Note: stripPII converts to lowercase and replaces in lowercase
				if !contains(strings.ToLower(resultStr), strings.ToLower(expected)) {
					t.Errorf("stripPII() result should contain %s, got %s", expected, resultStr)
				}
			}
		})
	}
}

func TestModifier_CreateFakeExecutable(t *testing.T) {
	config := &TrafficModifierConfig{
		Enabled: true,
	}
	modifier := NewModifier(config, nil)

	req := &Request{
		ID:     "test-1",
		Domain: "malware.com",
		Path:   "/payload.exe",
		Method: "GET",
	}

	fake := modifier.createFakeExecutable(req)
	
	if len(fake) == 0 {
		t.Error("createFakeExecutable() returned empty content")
	}

	fakeStr := string(fake)
	if !contains(fakeStr, "Pack-A-Mal") {
		t.Error("Fake executable should contain Pack-A-Mal marker")
	}

	if !contains(fakeStr, req.ID) {
		t.Error("Fake executable should contain request ID")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && hasSubstring(s, substr))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
