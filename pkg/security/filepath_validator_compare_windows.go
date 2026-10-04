//go:build windows

package security

import "strings"

// normalizeForCompare folds a path into the form used for containment
// comparisons on Windows, where the filesystem is case-insensitive and
// accepts both slash flavours.
func normalizeForCompare(path string) string {
	return strings.ToLower(strings.ReplaceAll(path, "/", `\`))
}
