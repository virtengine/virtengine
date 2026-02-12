package conv

import (
	"unsafe"
)

// StrToBytes performs a safe copy from string to []byte.
func StrToBytes(s string) []byte {
	return []byte(s)
}

// UnsafeStrToBytes uses unsafe to convert string into byte array. Returned bytes
// must not be altered after this function is called as it will cause a segmentation fault.
func UnsafeStrToBytes(s string) []byte {
	//nolint:gosec // G103: zero-copy conversion for hot paths; returned slice is read-only and tied to string lifetime.
	return unsafe.Slice(unsafe.StringData(s), len(s)) // ref https://github.com/golang/go/issues/53003#issuecomment-1140276077
}

// UnsafeBytesToStr is meant to make a zero allocation conversion
// from []byte -> string to speed up operations, it is not meant
// to be used generally, but for a specific pattern to delete keys
// from a map.
func UnsafeBytesToStr(b []byte) string {
	//nolint:gosec // G103: zero-copy conversion; caller must not mutate b while the string is in use.
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// BytesToStr performs a safe copy from []byte to string.
func BytesToStr(b []byte) string {
	return string(b)
}
