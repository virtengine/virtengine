package types

import (
	"fmt"
	"strconv"
	"strings"
)

const recipientKeyVersionSeparator = ":v"

// FormatRecipientKeyID returns a versioned key ID string for a fingerprint.
func FormatRecipientKeyID(fingerprint string, version uint32) string {
	if version == 0 {
		return fingerprint
	}
	return fmt.Sprintf("%s%s%d", fingerprint, recipientKeyVersionSeparator, version)
}

// ParseRecipientKeyID extracts the fingerprint and version from a versioned key ID.
// If no version is present, version is 0 and ok is false.
func ParseRecipientKeyID(keyID string) (fingerprint string, version uint32, ok bool) {
	if keyID == "" {
		return "", 0, false
	}
	parts := strings.Split(keyID, recipientKeyVersionSeparator)
	if len(parts) != 2 {
		return keyID, 0, false
	}
	if parts[0] == "" {
		return "", 0, false
	}
	parsed, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return parts[0], 0, false
	}
	return parts[0], uint32(parsed), true
}

// NormalizeRecipientKeyID strips any version suffix to return the fingerprint.
func NormalizeRecipientKeyID(keyID string) string {
	fingerprint, _, _ := ParseRecipientKeyID(keyID)
	return fingerprint
}
