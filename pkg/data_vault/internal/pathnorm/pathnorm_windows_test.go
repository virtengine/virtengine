//go:build windows

package pathnorm

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"unsafe"
)

var (
	testKernel32          = syscall.NewLazyDLL("kernel32.dll")
	procGetShortPathNameW = testKernel32.NewProc("GetShortPathNameW")
)

func shortPath(t *testing.T, long string) string {
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

func TestLongPathExpandsShortNames(t *testing.T) {
	base, err := os.MkdirTemp("", "pathnorm-shortname-")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(base) })

	long := filepath.Join(base, "long-component-name", "nested")
	if err := os.MkdirAll(long, 0o700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	short := shortPath(t, long)
	if filepath.Clean(short) == filepath.Clean(long) {
		t.Skip("volume exposes no 8.3 short names")
	}
	if !strings.Contains(short, "~") {
		t.Fatalf("expected a tilde short name, got %q", short)
	}

	got, err := LongPath(short)
	if err != nil {
		t.Fatalf("LongPath(%q): %v", short, err)
	}
	if !strings.EqualFold(filepath.Clean(got), filepath.Clean(long)) {
		t.Fatalf("LongPath(%q) = %q, want %q", short, got, long)
	}
}

func TestLongPathIsIdempotentOnLongPaths(t *testing.T) {
	base, err := os.MkdirTemp("", "pathnorm-idempotent-")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(base) })

	got, err := LongPath(base)
	if err != nil {
		t.Fatalf("LongPath(%q): %v", base, err)
	}
	if !strings.EqualFold(filepath.Clean(got), filepath.Clean(base)) {
		t.Fatalf("LongPath(%q) = %q, want the input unchanged", base, got)
	}
}

// The guard only calls LongPath for ancestors that Lstat proved to exist, so a
// missing path returning an error is expected behaviour worth pinning.
func TestLongPathErrorsOnMissingPath(t *testing.T) {
	missing := filepath.Join(os.TempDir(), "pathnorm-missing-definitely-absent-9f3c1a")
	if _, err := LongPath(missing); err == nil {
		t.Fatalf("LongPath(%q) returned no error for a missing path", missing)
	}
}
