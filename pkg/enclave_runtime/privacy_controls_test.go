package enclave_runtime

import (
	"regexp"
	"strings"
	"testing"
)

func TestLogRedactor_Redact(t *testing.T) {
	redactor := NewLogRedactor()

	tests := []struct {
		name     string
		input    string
		contains string
		notContains string
	}{
		{
			name:        "redacts base64 data",
			input:       `data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUA`,
			notContains: "iVBORw0",
			contains:    "[BASE64_DATA_REDACTED]",
		},
		{
			name:        "redacts hex data",
			input:       `hash: 0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef`,
			notContains: "1234567890abcdef",
			contains:    "[HEX_DATA_REDACTED]",
		},
		{
			name:        "redacts request body",
			input:       `{"method": "post", "body": "sensitive data here"}`,
			notContains: "sensitive data here",
			contains:    "[REDACTED]",
		},
		{
			name:        "preserves safe content",
			input:       `user logged in successfully`,
			contains:    "user logged in successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := redactor.Redact(tt.input)
			
			if tt.notContains != "" && strings.Contains(result, tt.notContains) {
				t.Errorf("result should not contain %q, got %q", tt.notContains, result)
			}
			
			if tt.contains != "" && !strings.Contains(result, tt.contains) {
				t.Errorf("result should contain %q, got %q", tt.contains, result)
			}
		})
	}
}

func TestSanitizeForLogging(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		contains string
	}{
		{
			name:     "binary data",
			input:    make([]byte, 100),
			contains: "BINARY_DATA",
		},
		{
			name:     "string input",
			input:    "normal string",
			contains: "normal string",
		},
		{
			name:     "other type",
			input:    12345,
			contains: "REDACTED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeForLogging(tt.input)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("expected result to contain %q, got %q", tt.contains, result)
			}
		})
	}
}

func TestTruncateForLogging(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		expected  string
	}{
		{
			name:      "short string",
			input:     "hello",
			maxLength: 10,
			expected:  "hello",
		},
		{
			name:      "long string",
			input:     "this is a very long string that needs truncation",
			maxLength: 10,
			expected:  "this is a ...[TRUNCATED]",
		},
		{
			name:      "exact length",
			input:     "exact",
			maxLength: 5,
			expected:  "exact",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateForLogging(tt.input, tt.maxLength)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestIsSensitiveField(t *testing.T) {
	sensitiveFields := []string{
		"plaintext_data",
		"decrypted_content",
		"user_identity",
		"biometric_hash",
		"face_embedding",
		"selfie_image",
		"ssn_number",
		"passport_id",
		"private_key",
	}

	safeFields := []string{
		"user_id",
		"created_at",
		"block_height",
		"validator_address",
		"score",
		"status",
	}

	for _, field := range sensitiveFields {
		if !IsSensitiveField(field) {
			t.Errorf("expected %q to be identified as sensitive", field)
		}
	}

	for _, field := range safeFields {
		if IsSensitiveField(field) {
			t.Errorf("expected %q to NOT be identified as sensitive", field)
		}
	}
}

func TestDefaultMemoryCanaries(t *testing.T) {
	canaries := DefaultMemoryCanaries()
	
	if len(canaries) == 0 {
		t.Error("expected at least one memory canary")
	}

	for _, canary := range canaries {
		if len(canary.Pattern) == 0 {
			t.Error("canary pattern should not be empty")
		}
		if canary.Description == "" {
			t.Error("canary description should not be empty")
		}
	}
}

func TestGetStaticAnalysisChecks(t *testing.T) {
	checks := GetStaticAnalysisChecks()
	
	if len(checks) == 0 {
		t.Error("expected at least one static analysis check")
	}

	for _, check := range checks {
		if check.Name == "" {
			t.Error("check name should not be empty")
		}
		if check.Pattern == nil {
			t.Error("check pattern should not be nil")
		}
		if check.Severity != "error" && check.Severity != "warning" {
			t.Errorf("check severity should be error or warning, got %s", check.Severity)
		}
	}
}

func TestForbiddenLogPatterns(t *testing.T) {
	// Test that forbidden patterns would catch problematic code
	testCases := []struct {
		code        string
		shouldMatch bool
	}{
		{
			code:        `log.Debug("plaintext data: ", data)`,
			shouldMatch: true,
		},
		{
			code:        `log.Info("decrypted content: %v", content)`,
			shouldMatch: true,
		},
		{
			code:        `log.Debug("Processing request")`,
			shouldMatch: false,
		},
		{
			code:        `log.Info("Verification complete")`,
			shouldMatch: false,
		},
	}

	for _, tc := range testCases {
		matched := false
		for _, pattern := range ForbiddenLogPatterns {
			re := regexp.MustCompile(pattern)
			if re.MatchString(tc.code) {
				matched = true
				break
			}
		}

		if matched != tc.shouldMatch {
			t.Errorf("code %q: expected match=%v, got match=%v", tc.code, tc.shouldMatch, matched)
		}
	}
}

func TestIncidentProcedureDocumented(t *testing.T) {
	// Verify incident procedure contains required sections
	requiredSections := []string{
		"DETECTION",
		"CONTAINMENT",
		"ANALYSIS",
		"REMEDIATION",
		"COMMUNICATION",
		"PREVENTION",
	}

	for _, section := range requiredSections {
		if !strings.Contains(IncidentProcedure, section) {
			t.Errorf("incident procedure should contain %s section", section)
		}
	}
}

func TestLogRedactor_RedactBytes(t *testing.T) {
	redactor := NewLogRedactor()
	
	input := []byte(`{"body": "sensitive content"}`)
	result := redactor.RedactBytes(input)
	
	if strings.Contains(string(result), "sensitive content") {
		t.Error("should have redacted sensitive content from bytes")
	}
}

func TestNewLogRedactorWithRules(t *testing.T) {
	customRules := []RedactionRule{
		{
			Name:        "custom_pattern",
			Pattern:     regexp.MustCompile(`SECRET_\d+`),
			Replacement: "[CUSTOM_REDACTED]",
		},
	}

	redactor := NewLogRedactorWithRules(customRules)
	
	input := "Found SECRET_12345 in data"
	result := redactor.Redact(input)
	
	if strings.Contains(result, "SECRET_12345") {
		t.Error("should have applied custom redaction rule")
	}
	if !strings.Contains(result, "[CUSTOM_REDACTED]") {
		t.Error("should contain custom replacement")
	}
}
