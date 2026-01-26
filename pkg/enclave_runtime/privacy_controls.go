package enclaVIRTENGINE_runtime

import (
	"regexp"
	"strings"
)

// PrivacyControls provides utilities for preventing plaintext identity data
// from appearing in logs, telemetry, and other observable channels.

// SensitivePatterns defines patterns that indicate sensitive identity data
var SensitivePatterns = []string{
	// Personal identifiers
	`\b\d{3}-\d{2}-\d{4}\b`,                    // SSN pattern
	`\b\d{9}\b`,                                 // Passport/ID numbers
	`\b[A-Z]{1,2}\d{6,9}\b`,                     // Various ID formats
	
	// Biometric indicators
	`face_embedding`,
	`biometric_data`,
	`fingerprint`,
	`iris_scan`,
	
	// Document content
	`document_ocr`,
	`extracted_text`,
	`id_document`,
	
	// Address patterns (simplified)
	`\b\d+\s+\w+\s+(street|st|avenue|ave|road|rd|drive|dr)\b`,
	
	// Email patterns (for identity context)
	`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`,
}

// ForbiddenLogPatterns are patterns that should never appear in logs
// Static analysis should fail builds if these patterns are found in logging calls
var ForbiddenLogPatterns = []string{
	// Direct plaintext logging
	`log\.(Debug|Info|Warn|Error|Fatal)\(.*plaintext`,
	`log\.(Debug|Info|Warn|Error|Fatal)\(.*decrypted`,
	`log\.(Debug|Info|Warn|Error|Fatal)\(.*identity_data`,
	`log\.(Debug|Info|Warn|Error|Fatal)\(.*biometric`,
	`log\.(Debug|Info|Warn|Error|Fatal)\(.*document_content`,
	`log\.(Debug|Info|Warn|Error|Fatal)\(.*face_image`,
	`log\.(Debug|Info|Warn|Error|Fatal)\(.*selfie`,
	
	// Printf-style sensitive data
	`fmt\.(Print|Printf|Println)\(.*plaintext`,
	`fmt\.(Print|Printf|Println)\(.*decrypted`,
	
	// Structured logging with sensitive fields
	`\.With\("plaintext"`,
	`\.With\("decrypted"`,
	`\.With\("identity_data"`,
	`\.With\("document_image"`,
}

// RedactionRules defines how different types of data should be redacted
type RedactionRule struct {
	// Name is the rule identifier
	Name string
	
	// Pattern is the regex pattern to match
	Pattern *regexp.Regexp
	
	// Replacement is what to replace matches with
	Replacement string
	
	// TruncateLength truncates the field to this length (0 for no truncation)
	TruncateLength int
}

// DefaultRedactionRules returns the default set of redaction rules
func DefaultRedactionRules() []RedactionRule {
	return []RedactionRule{
		{
			Name:        "encrypted_blob",
			Pattern:     regexp.MustCompile(`([A-Za-z0-9+/]{64,})`),
			Replacement: "[ENCRYPTED_BLOB_REDACTED]",
		},
		{
			Name:        "base64_data",
			Pattern:     regexp.MustCompile(`data:[^;]+;base64,[A-Za-z0-9+/=]+`),
			Replacement: "[BASE64_DATA_REDACTED]",
		},
		{
			Name:        "hex_data",
			Pattern:     regexp.MustCompile(`0x[A-Fa-f0-9]{64,}`),
			Replacement: "[HEX_DATA_REDACTED]",
		},
		{
			Name:        "request_body",
			Pattern:     regexp.MustCompile(`"body":\s*"[^"]+"`),
			Replacement: `"body":"[REDACTED]"`,
		},
		{
			Name:           "scope_data",
			Pattern:        regexp.MustCompile(`"scope_data":\s*"[^"]+"`),
			Replacement:    `"scope_data":"[REDACTED]"`,
			TruncateLength: 0,
		},
	}
}

// LogRedactor provides log redaction functionality
type LogRedactor struct {
	rules []RedactionRule
}

// NewLogRedactor creates a new log redactor with default rules
func NewLogRedactor() *LogRedactor {
	return &LogRedactor{
		rules: DefaultRedactionRules(),
	}
}

// NewLogRedactorWithRules creates a log redactor with custom rules
func NewLogRedactorWithRules(rules []RedactionRule) *LogRedactor {
	return &LogRedactor{
		rules: rules,
	}
}

// Redact applies redaction rules to the input string
func (r *LogRedactor) Redact(input string) string {
	result := input
	for _, rule := range r.rules {
		result = rule.Pattern.ReplaceAllString(result, rule.Replacement)
	}
	return result
}

// RedactBytes applies redaction rules to byte data
func (r *LogRedactor) RedactBytes(input []byte) []byte {
	return []byte(r.Redact(string(input)))
}

// SanitizeForLogging sanitizes a value for safe logging
// Returns a redacted version of the input
func SanitizeForLogging(input interface{}) string {
	switch v := input.(type) {
	case []byte:
		if len(v) > 32 {
			return "[BINARY_DATA_" + string(rune(len(v))) + "_BYTES]"
		}
		return "[BINARY_DATA]"
	case string:
		redactor := NewLogRedactor()
		return redactor.Redact(v)
	default:
		return "[REDACTED]"
	}
}

// TruncateForLogging truncates a string to a safe length for logging
func TruncateForLogging(input string, maxLength int) string {
	if len(input) <= maxLength {
		return input
	}
	return input[:maxLength] + "...[TRUNCATED]"
}

// IsSensitiveField checks if a field name indicates sensitive data
func IsSensitiveField(fieldName string) bool {
	sensitiveFields := []string{
		"plaintext",
		"decrypted",
		"identity",
		"biometric",
		"document",
		"face",
		"selfie",
		"embedding",
		"ssn",
		"passport",
		"id_number",
		"address",
		"phone",
		"email",
		"dob",
		"date_of_birth",
		"secret",
		"private_key",
		"seed",
		"mnemonic",
	}
	
	lowerField := strings.ToLower(fieldName)
	for _, sensitive := range sensitiveFields {
		if strings.Contains(lowerField, sensitive) {
			return true
		}
	}
	return false
}

// MemoryCanary is a pattern that should never appear in host memory
// Used for security testing to verify no plaintext leakage
type MemoryCanary struct {
	// Pattern is the byte pattern to search for
	Pattern []byte
	
	// Description describes what this canary represents
	Description string
}

// DefaultMemoryCanaries returns canaries for memory inspection testing
func DefaultMemoryCanaries() []MemoryCanary {
	return []MemoryCanary{
		{
			Pattern:     []byte("IDENTITY_PLAINTEXT_MARKER"),
			Description: "Identity plaintext marker",
		},
		{
			Pattern:     []byte("BIOMETRIC_DATA_MARKER"),
			Description: "Biometric data marker",
		},
		{
			Pattern:     []byte("DOCUMENT_CONTENT_MARKER"),
			Description: "Document content marker",
		},
	}
}

// StaticAnalysisCheck represents a check for static analysis
type StaticAnalysisCheck struct {
	// Name is the check identifier
	Name string
	
	// Description explains what the check looks for
	Description string
	
	// Pattern is the regex pattern that should NOT appear
	Pattern *regexp.Regexp
	
	// Severity is the severity level (error, warning)
	Severity string
	
	// PathPatterns limits the check to files matching these patterns
	PathPatterns []string
}

// GetStaticAnalysisChecks returns checks for CI/CD pipeline
func GetStaticAnalysisChecks() []StaticAnalysisCheck {
	return []StaticAnalysisCheck{
		{
			Name:        "no_plaintext_logging",
			Description: "Prevent logging of plaintext identity data",
			Pattern:     regexp.MustCompile(`log\.[A-Za-z]+\([^)]*plaintext`),
			Severity:    "error",
			PathPatterns: []string{
				"*/veid/*",
				"*/enclave/*",
				"*/identity/*",
			},
		},
		{
			Name:        "no_decrypted_logging",
			Description: "Prevent logging of decrypted data",
			Pattern:     regexp.MustCompile(`log\.[A-Za-z]+\([^)]*decrypted`),
			Severity:    "error",
			PathPatterns: []string{
				"*/veid/*",
				"*/enclave/*",
				"*/encryption/*",
			},
		},
		{
			Name:        "no_fmt_print_sensitive",
			Description: "Prevent fmt.Print of sensitive data",
			Pattern:     regexp.MustCompile(`fmt\.(Print|Printf|Println)\([^)]*(?:plaintext|decrypted|biometric)`),
			Severity:    "error",
			PathPatterns: []string{
				"*",
			},
		},
		{
			Name:        "no_debug_in_production",
			Description: "Prevent debug logging in production paths",
			Pattern:     regexp.MustCompile(`log\.Debug\([^)]*identity`),
			Severity:    "warning",
			PathPatterns: []string{
				"*/veid/*",
			},
		},
	}
}

// LeakageIncidentReport represents a potential leakage incident
type LeakageIncidentReport struct {
	// Timestamp is when the incident was detected
	Timestamp int64
	
	// Source describes where the potential leak was detected
	Source string
	
	// Description describes the incident
	Description string
	
	// Severity is the severity level
	Severity string
	
	// StackTrace contains the stack trace if available
	StackTrace string
	
	// Remediation suggests remediation steps
	Remediation string
}

// IncidentProcedure documents the response procedure for leakage incidents
const IncidentProcedure = `
SUSPECTED PLAINTEXT LEAKAGE INCIDENT RESPONSE PROCEDURE

1. DETECTION
   - Monitor for anomalous patterns in verification flow
   - Check memory inspection results from security tests
   - Review log analysis for redaction failures

2. CONTAINMENT (Immediate - within 15 minutes)
   - Suspend affected validator(s) from consensus participation
   - Disable identity verification for affected scope types
   - Notify security team on-call

3. ANALYSIS (Within 1 hour)
   - Collect attestation quotes from affected validators
   - Gather verification logs (redacted)
   - Identify scope of potential exposure
   - Determine root cause

4. REMEDIATION (Based on severity)
   - Rotate enclave keys for affected validators
   - Update enclave measurements if code changes needed
   - Patch and redeploy affected components
   - Re-attest all affected validators

5. COMMUNICATION (Within 24 hours)
   - Notify affected users per privacy policy
   - File regulatory notifications if required
   - Update incident status page

6. PREVENTION (Post-incident)
   - Update controls based on root cause
   - Add new static analysis checks
   - Enhance memory inspection coverage
   - Document lessons learned
`
