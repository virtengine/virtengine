//go:build windows

package keys

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"unsafe"
)

var (
	shortPathKernel32     = syscall.NewLazyDLL("kernel32.dll")
	procGetShortPathNameW = shortPathKernel32.NewProc("GetShortPathNameW")
)

func windowsShortPath(t *testing.T, long string) string {
	t.Helper()
	utf16Path, err := syscall.UTF16PtrFromString(long)
	if err != nil {
		t.Fatalf("UTF16PtrFromString(%q): %v", long, err)
	}
	buffer := make([]uint16, syscall.MAX_PATH+1)
	written, _, callErr := procGetShortPathNameW.Call(
		uintptr(unsafe.Pointer(utf16Path)),
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(len(buffer)),
	)
	if written == 0 {
		t.Fatalf("GetShortPathNameW(%q): %v", long, callErr)
	}
	return syscall.UTF16ToString(buffer[:written])
}

// TestRejectSymlinkTargetAcceptsShortNameAncestor pins the same regression as the
// data_vault package test: an 8.3 short-name ancestor is a name alias, not a
// reparse point, so key custody paths under the runner's %TEMP%
// (C:\Users\RUNNER~1\AppData\Local\Temp) must be accepted.
func TestRejectSymlinkTargetAcceptsShortNameAncestor(t *testing.T) {
	base, err := os.MkdirTemp("", "data-vault-keys-shortname-")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(base) })

	root := filepath.Join(base, "key-custody-root")
	if err := os.MkdirAll(filepath.Join(root, "nested"), 0o700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	short := windowsShortPath(t, root)
	if filepath.Clean(short) == filepath.Clean(root) {
		t.Skip("volume exposes no 8.3 short names; cannot exercise the short-name regression")
	}
	if _, err := os.Lstat(short); err != nil {
		t.Fatalf("short-name path %q is unusable: %v", short, err)
	}

	if err := rejectSymlinkTarget(short); err != nil {
		t.Fatalf("rejectSymlinkTarget rejected an 8.3 short-name path\n  short: %s\n  long:  %s\n  err:   %v",
			short, root, err)
	}
	if err := rejectSymlinkTarget(filepath.Join(short, "nested")); err != nil {
		t.Fatalf("rejectSymlinkTarget rejected a descendant of an 8.3 short-name path: %v", err)
	}
}
