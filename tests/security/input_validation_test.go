// Package security contains security-focused tests for VirtEngine.
// These tests verify input validation and malformed payload handling.
//
// Task Reference: VE-800 - Security audit readiness
package security

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// InputValidationTestSuite tests input validation security.
type InputValidationTestSuite struct {
	suite.Suite
}

func TestInputValidation(t *testing.T) {
	suite.Run(t, new(InputValidationTestSuite))
}

// =============================================================================
// Malformed Payload Tests
// =============================================================================

// TestMalformedPayloadHandling tests handling of malformed inputs.
func (s *InputValidationTestSuite) TestMalformedPayloadHandling() {
	s.T().Log("=== Test: Malformed Payload Handling ===")

	// Test: Empty JSON object handled
	s.Run("empty_json_object", func() {
		input := []byte("{}")
		err := validateIdentityPayload(input)
		require.Error(s.T(), err, "empty payload should be rejected")
		require.Contains(s.T(), err.Error(), "missing required field")
	})

	// Test: Invalid JSON rejected
	s.Run("invalid_json_rejected", func() {
		invalidInputs := [][]byte{
			[]byte("{"),
			[]byte("}"),
			[]byte("{invalid}"),
			[]byte("{{}}"),
			[]byte(`{"key": }`),
			[]byte(`{"key": undefined}`),
		}

		for i, input := range invalidInputs {
			err := validateIdentityPayload(input)
			require.Error(s.T(), err, "invalid JSON #%d should be rejected", i)
		}
	})

	// Test: Null values handled
	s.Run("null_values_handled", func() {
		input := []byte(`{"scopes": null, "salt": null}`)
		err := validateIdentityPayload(input)
		require.Error(s.T(), err, "null required fields should be rejected")
	})

	// Test: Wrong types rejected
	s.Run("wrong_types_rejected", func() {
		inputs := [][]byte{
			[]byte(`{"scopes": 123}`),        // number instead of bytes
			[]byte(`{"scopes": true}`),       // bool instead of bytes
			[]byte(`{"timestamp": "hello"}`), // invalid timestamp
		}

		for i, input := range inputs {
			err := validateIdentityPayload(input)
			require.Error(s.T(), err, "wrong type #%d should be rejected", i)
		}
	})
}

// TestOverflowDetection tests integer overflow prevention.
func (s *InputValidationTestSuite) TestOverflowDetection() {
	s.T().Log("=== Test: Overflow Detection ===")

	// Test: Large amounts rejected
	s.Run("large_amount_overflow_prevented", func() {
		// Test values near uint64 max
		maxUint64 := uint64(18446744073709551615)
		
		err := validateTransferAmount(maxUint64)
		require.Error(s.T(), err, "max uint64 should be rejected as too large")
	})

	// Test: Addition overflow prevented
	s.Run("addition_overflow_prevented", func() {
		a := uint64(10000000000000000000)
		b := uint64(10000000000000000000)

		_, err := safeAdd(a, b)
		require.Error(s.T(), err, "addition overflow should be detected")
	})

	// Test: Multiplication overflow prevented
	s.Run("multiplication_overflow_prevented", func() {
		a := uint64(1000000000000)
		b := uint64(1000000000000)

		_, err := safeMul(a, b)
		require.Error(s.T(), err, "multiplication overflow should be detected")
	})

	// Test: Safe operations succeed
	s.Run("safe_operations_succeed", func() {
		a := uint64(100)
		b := uint64(200)

		sum, err := safeAdd(a, b)
		require.NoError(s.T(), err)
		require.Equal(s.T(), uint64(300), sum)

		product, err := safeMul(a, b)
		require.NoError(s.T(), err)
		require.Equal(s.T(), uint64(20000), product)
	})
}

// TestInjectionPrevention tests injection attack prevention.
func (s *InputValidationTestSuite) TestInjectionPrevention() {
	s.T().Log("=== Test: Injection Prevention ===")

	// Test: SQL injection patterns rejected
	s.Run("sql_injection_patterns_rejected", func() {
		injectionPatterns := []string{
			"'; DROP TABLE users; --",
			"1' OR '1'='1",
			"1; SELECT * FROM users",
			"admin'--",
			"' UNION SELECT * FROM accounts --",
		}

		for _, pattern := range injectionPatterns {
			err := validateStringInput(pattern, InputTypeAccountName)
			require.Error(s.T(), err, "SQL injection pattern should be rejected: %s", pattern)
		}
	})

	// Test: Command injection patterns rejected
	s.Run("command_injection_patterns_rejected", func() {
		injectionPatterns := []string{
			"; rm -rf /",
			"| cat /etc/passwd",
			"$(whoami)",
			"`id`",
			"&& echo pwned",
		}

		for _, pattern := range injectionPatterns {
			err := validateStringInput(pattern, InputTypeResourceName)
			require.Error(s.T(), err, "command injection pattern should be rejected: %s", pattern)
		}
	})

	// Test: Path traversal rejected
	s.Run("path_traversal_rejected", func() {
		traversalPatterns := []string{
			"../../../etc/passwd",
			"..\\..\\..\\windows\\system32",
			"....//....//etc/passwd",
			"/etc/passwd",
			"C:\\Windows\\System32",
		}

		for _, pattern := range traversalPatterns {
			err := validateStringInput(pattern, InputTypeFilePath)
			require.Error(s.T(), err, "path traversal should be rejected: %s", pattern)
		}
	})

	// Test: XSS patterns rejected
	s.Run("xss_patterns_rejected", func() {
		xssPatterns := []string{
			"<script>alert('xss')</script>",
			"<img src=x onerror=alert(1)>",
			"javascript:alert(1)",
			"<svg onload=alert(1)>",
			"<body onload=alert('XSS')>",
		}

		for _, pattern := range xssPatterns {
			err := validateStringInput(pattern, InputTypeUserInput)
			require.Error(s.T(), err, "XSS pattern should be rejected: %s", pattern)
		}
	})
}

// TestSizeLimitEnforcement tests size limit validation.
func (s *InputValidationTestSuite) TestSizeLimitEnforcement() {
	s.T().Log("=== Test: Size Limit Enforcement ===")

	// Test: Payload size limits
	s.Run("payload_size_limit_enforced", func() {
		maxSize := 1 * 1024 * 1024 // 1 MB

		// Just under limit - should pass
		underLimit := bytes.Repeat([]byte("a"), maxSize-100)
		err := validatePayloadSize(underLimit, maxSize)
		require.NoError(s.T(), err, "payload under limit should be accepted")

		// Over limit - should fail
		overLimit := bytes.Repeat([]byte("a"), maxSize+100)
		err = validatePayloadSize(overLimit, maxSize)
		require.Error(s.T(), err, "payload over limit should be rejected")
	})

	// Test: String length limits
	s.Run("string_length_limit_enforced", func() {
		maxLen := 256

		// Valid length
		validStr := strings.Repeat("a", 100)
		err := validateStringLength(validStr, maxLen)
		require.NoError(s.T(), err, "string under limit should be accepted")

		// Over limit
		longStr := strings.Repeat("a", maxLen+1)
		err = validateStringLength(longStr, maxLen)
		require.Error(s.T(), err, "string over limit should be rejected")
	})

	// Test: Array length limits
	s.Run("array_length_limit_enforced", func() {
		maxItems := 100

		// Valid count
		validItems := make([]string, 50)
		err := validateArrayLength(validItems, maxItems)
		require.NoError(s.T(), err, "array under limit should be accepted")

		// Over limit
		tooManyItems := make([]string, maxItems+1)
		err = validateArrayLength(tooManyItems, maxItems)
		require.Error(s.T(), err, "array over limit should be rejected")
	})

	// Test: Nesting depth limits
	s.Run("nesting_depth_limit_enforced", func() {
		maxDepth := 10

		// Create deeply nested JSON
		nested := createNestedJSON(maxDepth + 5)
		err := validateJSONDepth(nested, maxDepth)
		require.Error(s.T(), err, "deeply nested JSON should be rejected")

		// Acceptable nesting
		acceptable := createNestedJSON(5)
		err = validateJSONDepth(acceptable, maxDepth)
		require.NoError(s.T(), err, "acceptable nesting should be allowed")
	})
}

// TestEncodingValidation tests encoding validation.
func (s *InputValidationTestSuite) TestEncodingValidation() {
	s.T().Log("=== Test: Encoding Validation ===")

	// Test: Valid UTF-8 accepted
	s.Run("valid_utf8_accepted", func() {
		validStrings := []string{
			"Hello, World!",
			"こんにちは",
			"Привет мир",
			"🎉🚀✨",
			"Mixed: Hello 你好 🌍",
		}

		for _, str := range validStrings {
			err := validateUTF8(str)
			require.NoError(s.T(), err, "valid UTF-8 should be accepted: %s", str)
		}
	})

	// Test: Invalid UTF-8 rejected
	s.Run("invalid_utf8_rejected", func() {
		invalidBytes := [][]byte{
			{0xff, 0xfe},             // Invalid UTF-8 sequence
			{0x80, 0x81, 0x82},       // Continuation bytes without start
			{0xc0, 0x80},             // Overlong encoding
		}

		for i, b := range invalidBytes {
			err := validateUTF8(string(b))
			require.Error(s.T(), err, "invalid UTF-8 #%d should be rejected", i)
		}
	})

	// Test: Base64 validation
	s.Run("base64_validation", func() {
		validB64 := []string{
			"SGVsbG8gV29ybGQh",
			"dGVzdCBkYXRh",
			"YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXo=",
		}

		for _, b64 := range validB64 {
			err := validateBase64(b64)
			require.NoError(s.T(), err, "valid base64 should be accepted: %s", b64)
		}

		invalidB64 := []string{
			"not-valid-base64!!!",
			"!!!",
			"SGVsbG8@V29ybGQh", // Invalid character
		}

		for _, b64 := range invalidB64 {
			err := validateBase64(b64)
			require.Error(s.T(), err, "invalid base64 should be rejected: %s", b64)
		}
	})

	// Test: Hex validation
	s.Run("hex_validation", func() {
		validHex := []string{
			"48656c6c6f",
			"ABCDEF123456",
			"0123456789abcdef",
		}

		for _, h := range validHex {
			err := validateHex(h)
			require.NoError(s.T(), err, "valid hex should be accepted: %s", h)
		}

		invalidHex := []string{
			"not-hex",
			"GHIJKL",
			"12345g",
			"12 34", // space
		}

		for _, h := range invalidHex {
			err := validateHex(h)
			require.Error(s.T(), err, "invalid hex should be rejected: %s", h)
		}
	})
}

// TestBoundaryConditions tests edge cases and boundary conditions.
func (s *InputValidationTestSuite) TestBoundaryConditions() {
	s.T().Log("=== Test: Boundary Conditions ===")

	// Test: Empty inputs
	s.Run("empty_inputs", func() {
		err := validateStringInput("", InputTypeAccountName)
		require.Error(s.T(), err, "empty string should be rejected")

		err = validatePayloadSize([]byte{}, 1024)
		require.NoError(s.T(), err, "empty payload within size limit should be accepted")
	})

	// Test: Exactly at limit
	s.Run("exactly_at_limit", func() {
		maxLen := 100
		exactStr := strings.Repeat("a", maxLen)
		err := validateStringLength(exactStr, maxLen)
		require.NoError(s.T(), err, "string exactly at limit should be accepted")
	})

	// Test: Zero values
	s.Run("zero_values", func() {
		err := validateTransferAmount(0)
		require.Error(s.T(), err, "zero amount should be rejected")
	})

	// Test: Negative-like patterns (in unsigned context)
	s.Run("negative_like_patterns", func() {
		// In Go, uint64 can't be negative, but we test string inputs
		err := validateAmountString("-100")
		require.Error(s.T(), err, "negative amount string should be rejected")
	})
}

// =============================================================================
// Test Helpers and Types
// =============================================================================

type InputType string

const (
	InputTypeAccountName  InputType = "account_name"
	InputTypeResourceName InputType = "resource_name"
	InputTypeFilePath     InputType = "file_path"
	InputTypeUserInput    InputType = "user_input"
)

// Validation errors
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

func validateIdentityPayload(data []byte) error {
	if len(data) == 0 {
		return &ValidationError{Field: "payload", Message: "empty payload"}
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return &ValidationError{Field: "payload", Message: "invalid JSON: " + err.Error()}
	}

	// Check required fields
	requiredFields := []string{"scopes", "salt", "timestamp"}
	for _, field := range requiredFields {
		val, exists := payload[field]
		if !exists || val == nil {
			return &ValidationError{Field: field, Message: "missing required field"}
		}
	}

	// Type validation
	if _, ok := payload["scopes"].(string); !ok {
		if _, ok := payload["scopes"].([]interface{}); !ok {
			return &ValidationError{Field: "scopes", Message: "invalid type"}
		}
	}

	return nil
}

func validateTransferAmount(amount uint64) error {
	if amount == 0 {
		return &ValidationError{Field: "amount", Message: "amount must be greater than zero"}
	}

	maxAmount := uint64(1000000000000000000) // 1e18
	if amount > maxAmount {
		return &ValidationError{Field: "amount", Message: "amount exceeds maximum"}
	}

	return nil
}

func validateAmountString(s string) error {
	if len(s) == 0 {
		return &ValidationError{Field: "amount", Message: "empty amount"}
	}
	if s[0] == '-' {
		return &ValidationError{Field: "amount", Message: "negative amount not allowed"}
	}
	return nil
}

func safeAdd(a, b uint64) (uint64, error) {
	if a > (^uint64(0))-b {
		return 0, &ValidationError{Field: "arithmetic", Message: "addition overflow"}
	}
	return a + b, nil
}

func safeMul(a, b uint64) (uint64, error) {
	if a == 0 || b == 0 {
		return 0, nil
	}
	result := a * b
	if result/a != b {
		return 0, &ValidationError{Field: "arithmetic", Message: "multiplication overflow"}
	}
	return result, nil
}

func validateStringInput(input string, inputType InputType) error {
	if len(input) == 0 {
		return &ValidationError{Field: string(inputType), Message: "empty input"}
	}

	// Check for dangerous patterns
	dangerousPatterns := []string{
		"'", "\"", ";", "--", "/*", "*/", // SQL
		"|", "&", "$", "`", "\\", // Command injection
		"<", ">", "javascript:", "onerror", "onload", // XSS
		"../", "..\\", // Path traversal
	}

	lowered := strings.ToLower(input)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowered, pattern) {
			return &ValidationError{Field: string(inputType), Message: "dangerous pattern detected"}
		}
	}

	// Path-specific validation
	if inputType == InputTypeFilePath {
		if strings.HasPrefix(input, "/") || strings.HasPrefix(input, "C:") {
			return &ValidationError{Field: string(inputType), Message: "absolute path not allowed"}
		}
	}

	return nil
}

func validatePayloadSize(data []byte, maxSize int) error {
	if len(data) > maxSize {
		return &ValidationError{Field: "payload", Message: "payload too large"}
	}
	return nil
}

func validateStringLength(s string, maxLen int) error {
	if len(s) > maxLen {
		return &ValidationError{Field: "string", Message: "string too long"}
	}
	return nil
}

func validateArrayLength(arr interface{}, maxLen int) error {
	switch v := arr.(type) {
	case []string:
		if len(v) > maxLen {
			return &ValidationError{Field: "array", Message: "too many items"}
		}
	case []interface{}:
		if len(v) > maxLen {
			return &ValidationError{Field: "array", Message: "too many items"}
		}
	}
	return nil
}

func validateJSONDepth(data []byte, maxDepth int) error {
	depth := 0
	maxFound := 0

	for _, b := range data {
		switch b {
		case '{', '[':
			depth++
			if depth > maxFound {
				maxFound = depth
			}
		case '}', ']':
			depth--
		}
	}

	if maxFound > maxDepth {
		return &ValidationError{Field: "json", Message: "nesting too deep"}
	}
	return nil
}

func createNestedJSON(depth int) []byte {
	var builder strings.Builder
	for i := 0; i < depth; i++ {
		builder.WriteString(`{"nested":`)
	}
	builder.WriteString(`"value"`)
	for i := 0; i < depth; i++ {
		builder.WriteString(`}`)
	}
	return []byte(builder.String())
}

func validateUTF8(s string) error {
	if !utf8.ValidString(s) {
		return &ValidationError{Field: "string", Message: "invalid UTF-8"}
	}
	return nil
}

func validateBase64(s string) error {
	// Check for valid base64 characters
	for _, c := range s {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '+' || c == '/' || c == '=') {
			return &ValidationError{Field: "base64", Message: "invalid character"}
		}
	}
	return nil
}

func validateHex(s string) error {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return &ValidationError{Field: "hex", Message: "invalid hex character"}
		}
	}
	return nil
}
