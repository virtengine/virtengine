// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package provider_daemon

import (
	"strings"
	"unicode"
)

// maxLogValueLen bounds how much of an externally-derived value is written to a
// log line, so one field cannot flood or skew the daemon log.
const maxLogValueLen = 128

// sanitizeLogValue makes an externally-derived value safe to interpolate into a
// log line.
//
// The lifecycle controller and the offering publication service log identifiers
// that arrive from outside this process: lifecycle operation IDs, allocation IDs
// and correlation IDs come from signed Waldur callbacks, and offering identifiers
// come from the Waldur API. A value containing a carriage return or newline
// would otherwise forge additional log records or hide activity (CWE-117 log
// injection, gosec G706). Line breaks are collapsed to spaces so the value stays
// on one line and remains recognisable; every other control character is
// dropped, and the result is length-bounded.
//
// The same helper exists as a package-private function in pkg/govdata, which
// sanitises adapter configuration the same way. It is duplicated rather than
// exported from a shared package to keep this change contained to the daemon;
// unifying the two is a follow-up.
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
