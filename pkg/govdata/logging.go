package govdata

import (
	"strings"
	"unicode"
)

// maxLogValueLen bounds how much of an externally-derived value is written to a
// log line, so that one field cannot flood or skew the audit log.
const maxLogValueLen = 128

// sanitizeLogValue makes an externally-derived value safe to interpolate into a
// log line.
//
// Values reaching this helper originate from adapter configuration and
// environment variables (see the individual adapter LoadConfig functions), so
// they are attacker-influenced in the sense that they are not compile-time
// constants. Carriage returns, newlines and other control characters are
// removed before logging so a crafted value cannot forge additional log records
// or hide activity (CWE-117 log injection, gosec G706).
func sanitizeLogValue(value string) string {
	if value == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(value))
	for _, r := range value {
		switch {
		case r == '\r' || r == '\n' || r == '\t':
			// Collapse line breaks into spaces so the value stays on one line
			// while remaining recognisable in the log.
			b.WriteByte(' ')
		case unicode.IsControl(r):
			// Drop every other control character.
		default:
			b.WriteRune(r)
		}
		if b.Len() >= maxLogValueLen {
			break
		}
	}

	return b.String()
}
