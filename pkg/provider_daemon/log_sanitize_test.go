// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package provider_daemon

import (
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
)

// TestSanitizeLogValueStripsLineBreaks is the G706 control: a lifecycle
// operation ID arriving from a Waldur callback must not be able to inject a
// second log record.
func TestSanitizeLogValueStripsLineBreaks(t *testing.T) {
	hostile := "op-1\n2026-01-01T00:00:00Z FAKE ADMIN LOGIN SUCCEEDED"

	got := sanitizeLogValue(hostile)

	assert.NotContains(t, got, "\n", "a newline would forge a second log record")
	assert.Contains(t, got, "op-1")
	assert.Contains(t, got, "FAKE ADMIN LOGIN SUCCEEDED",
		"the value should stay visible on one line, not be silently dropped")
	assert.False(t, strings.ContainsAny(got, "\r\n\t"))

	// Every control character must be gone, including NUL and ANSI escapes.
	dirty := "a\x00b\x1b[31mred\x7fc\rd\te"
	clean := sanitizeLogValue(dirty)
	for _, r := range clean {
		assert.False(t, unicode.IsControl(r), "control character %q survived", r)
	}
}

// TestSanitizeLogValueBoundsLength keeps one field from flooding the log.
func TestSanitizeLogValueBoundsLength(t *testing.T) {
	got := sanitizeLogValue(strings.Repeat("A", 10_000))
	assert.LessOrEqual(t, len(got), maxLogValueLen)
}

// TestSanitizeLogValuePreservesPlainValues is the positive control: ordinary
// identifiers are unchanged, so the fix does not alter normal log output.
func TestSanitizeLogValuePreservesPlainValues(t *testing.T) {
	for _, v := range []string{"", "op-1", "alloc/42", "ve1abc", "corr_id-9"} {
		assert.Equal(t, v, sanitizeLogValue(v))
	}
}
