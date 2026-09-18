//go:build windows

package pathnorm

import (
	"errors"
	"fmt"
	"syscall"
	"unsafe"
)

var (
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	procGetLongPathNameW    = kernel32.NewProc("GetLongPathNameW")
	errLongPathNameTooSmall = errors.New("pathnorm: long path name exceeds buffer")
)

// GetLongPathNameW converts 8.3 short-name components to their long forms. It
// does not resolve symlinks or any other reparse point.
//
// The path may be longer than MAX_PATH, so a short buffer is not an error: the
// call returns the required size (including the terminating NUL) and we retry.
func longPath(path string) (string, error) {
	utf16Path, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return "", fmt.Errorf("pathnorm: convert %q: %w", path, err)
	}

	size := uint32(len(path) + 1)
	if size < syscall.MAX_PATH {
		size = syscall.MAX_PATH
	}
	for attempt := 0; attempt < 8; attempt++ {
		buffer := make([]uint16, size)
		written, _, callErr := procGetLongPathNameW.Call(
			uintptr(unsafe.Pointer(utf16Path)),
			uintptr(unsafe.Pointer(&buffer[0])),
			uintptr(len(buffer)),
		)
		// written counts the characters written, excluding the terminating NUL.
		if written == 0 {
			if callErrno, ok := callErr.(syscall.Errno); ok && callErrno != 0 {
				return "", fmt.Errorf("pathnorm: GetLongPathNameW(%q): %w", path, callErrno)
			}
			return "", fmt.Errorf("pathnorm: GetLongPathNameW(%q): %w", path, callErr)
		}
		if int(written) < len(buffer) {
			return syscall.UTF16ToString(buffer[:written]), nil
		}
		// The buffer was too small; the return value is the required size.
		size = uint32(written) + 1
	}
	return "", fmt.Errorf("pathnorm: GetLongPathNameW(%q): %w", path, errLongPathNameTooSmall)
}
