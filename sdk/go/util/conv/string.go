package conv

import (
	"unsafe"
)

// Security note: unsafe conversions alias the original memory. Callers must treat
// returned values as read-only and ensure the source outlives the alias.

// StrToBytes performs a safe copy from string to []byte.
func StrToBytes(s string) []byte {
	return []byte(s)
}

// UnsafeStrToBytes uses unsafe to convert string into byte array without allocation.
// Returned bytes must be treated as read-only and must not outlive the source string.
// Prefer StrToBytes for safety unless a hot path requires zero-copy conversion.
func UnsafeStrToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s)) //nolint:gosec // G103: zero-copy conversion; slice aliases string data and must remain read-only for the string's lifetime.
}

// UnsafeBytesToStr is meant to make a zero allocation conversion from []byte -> string.
// Callers must ensure the byte slice is not mutated while the string is in use.
// Prefer BytesToStr for safety unless a hot path requires zero-copy conversion.
func UnsafeBytesToStr(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(unsafe.SliceData(b), len(b)) //nolint:gosec // G103: zero-copy conversion; caller must not mutate b while the string is in use.
}

// BytesToStr performs a safe copy from []byte to string.
func BytesToStr(b []byte) string {
	return string(b)
}
