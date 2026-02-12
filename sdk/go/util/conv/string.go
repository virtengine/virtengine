package conv

import (
	"unsafe"
)

// StrToBytes performs a safe copy from string to []byte.
func StrToBytes(s string) []byte {
	return []byte(s)
}

// UnsafeStrToBytes uses unsafe to convert string into byte array without allocation.
// Returned bytes must be treated as read-only and must not outlive the source string.
// Prefer StrToBytes for safety unless a hot path requires zero-copy conversion.
func UnsafeStrToBytes(s string) []byte {
	//nolint:gosec // G103: zero-copy conversion for hot paths; returned slice is read-only and tied to string lifetime.
	return unsafe.Slice(unsafe.StringData(s), len(s)) // ref https://github.com/golang/go/issues/53003#issuecomment-1140276077
}

// UnsafeBytesToStr is meant to make a zero allocation conversion from []byte -> string.
// Callers must ensure the byte slice is not mutated while the string is in use.
// Prefer BytesToStr for safety unless a hot path requires zero-copy conversion.
func UnsafeBytesToStr(b []byte) string {
	//nolint:gosec // G103: zero-copy conversion; caller must not mutate b while the string is in use.
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// BytesToStr performs a safe copy from []byte to string.
func BytesToStr(b []byte) string {
	return string(b)
}
