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

// longTempRoot returns the long-name spelling of a freshly created temp
// directory, and fails the test if the volume cannot supply one.
//
// Normalising the root first matters: on hosts whose %TEMP% is itself handed
// out in 8.3 form (GitHub's windows-latest runners give
// C:\Users\RUNNER~1\AppData\Local\Temp) the raw MkdirTemp base already carries a
// tilde, and building the "long" operand from it would demand that LongPath
// *preserve* that tilde - which is the opposite of what it is for.
func longTempRoot(t *testing.T, pattern string) string {
	t.Helper()
	base, err := os.MkdirTemp("", pattern)
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(base) })

	long, err := LongPath(base)
	if err != nil {
		t.Fatalf("LongPath(%q): %v", base, err)
	}
	return long
}

func TestLongPathExpandsShortNames(t *testing.T) {
	base := longTempRoot(t, "pathnorm-shortname-")

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

// TestLongPathIsIdempotentOnLongPaths pins the property the guard actually
// relies on: re-normalising an already-normalised path is a no-op. Asserting
// "LongPath(x) == x" instead would be wrong on any host whose temp root is
// itself a short name, because the first call is then *supposed* to change it.
func TestLongPathIsIdempotentOnLongPaths(t *testing.T) {
	base := longTempRoot(t, "pathnorm-idempotent-")

	once, err := LongPath(base)
	if err != nil {
		t.Fatalf("LongPath(%q): %v", base, err)
	}
	twice, err := LongPath(once)
	if err != nil {
		t.Fatalf("LongPath(%q): %v", once, err)
	}
	if !strings.EqualFold(filepath.Clean(once), filepath.Clean(twice)) {
		t.Fatalf("LongPath is not idempotent: LongPath(%q) = %q but LongPath(%q) = %q",
			base, once, once, twice)
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
