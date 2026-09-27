package govdata

import (
	"strings"
	"testing"
)

func TestSanitizeLogValueStripsLogForgingInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "newline cannot forge a second log record",
			input:    "ACME\n2026-01-01 FAKE ADMIN LOGIN SUCCEEDED",
			expected: "ACME 2026-01-01 FAKE ADMIN LOGIN SUCCEEDED",
		},
		{
			name:     "carriage return is collapsed",
			input:    "ACME\r\nforged",
			expected: "ACME  forged",
		},
		{
			name:     "tab is collapsed",
			input:    "ACME\tSPOOF",
			expected: "ACME SPOOF",
		},
		{
			name:     "other control characters are dropped",
			input:    "AC\x00ME\x1b[31mRED",
			expected: "ACME[31mRED",
		},
		{
			name:     "plain value is unchanged",
			input:    "acme-org-1",
			expected: "acme-org-1",
		},
		{
			name:     "empty value stays empty",
			input:    "",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := sanitizeLogValue(tc.input)
			if got != tc.expected {
				t.Fatalf("sanitizeLogValue(%q) = %q, want %q", tc.input, got, tc.expected)
			}
			if strings.ContainsAny(got, "\r\n") {
				t.Fatalf("sanitized value still contains a line break: %q", got)
			}
		})
	}
}

func TestSanitizeLogValueBoundsLength(t *testing.T) {
	t.Parallel()

	got := sanitizeLogValue(strings.Repeat("a", maxLogValueLen*3))
	if len(got) > maxLogValueLen {
		t.Fatalf("sanitized value length = %d, want <= %d", len(got), maxLogValueLen)
	}
}
