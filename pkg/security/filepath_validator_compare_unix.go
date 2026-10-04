//go:build !windows

package security

// normalizeForCompare is a no-op on case-sensitive platforms: POSIX paths
// compare byte-for-byte, so folding case here would widen the allowlist.
func normalizeForCompare(path string) string {
	return path
}
