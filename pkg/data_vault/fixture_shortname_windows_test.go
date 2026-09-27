//go:build windows

package data_vault

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"unsafe"
)

var (
	shortPathKernel32     = syscall.NewLazyDLL("kernel32.dll")
	procGetShortPathNameW = shortPathKernel32.NewProc("GetShortPathNameW")
)

// windowsShortPath returns the 8.3 short-name spelling of an existing path,
// which is what GitHub's windows-latest runners hand out as %TEMP%
// (C:\Users\RUNNER~1\AppData\Local\Temp).
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

// TestRejectFixtureSymlinkAcceptsShortNameAncestor pins the regression seen in CI
// job "Windows Native Build and Unit Tests" (run 35398529441):
//
//	fixture path ancestor is a reparse point: C:\Users\RUNNER~1\AppData\Local\Temp\...
//
// An 8.3 short-name ancestor is a name alias, not a reparse point, so it must be
// accepted. filepath.EvalSymlinks expands short names as well as resolving
// symlinks/junctions, which is why a naive resolved-vs-original comparison
// misfires on the runner's own %TEMP%.
func TestRejectFixtureSymlinkAcceptsShortNameAncestor(t *testing.T) {
	base, err := os.MkdirTemp("", "data-vault-shortname-")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(base) })

	root := filepath.Join(base, "fixture-artifact-root")
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

	if err := rejectFixtureSymlink(short); err != nil {
		t.Fatalf("rejectFixtureSymlink rejected an 8.3 short-name path\n  short: %s\n  long:  %s\n  err:   %v",
			short, root, err)
	}
	if err := rejectFixtureSymlink(filepath.Join(short, "nested")); err != nil {
		t.Fatalf("rejectFixtureSymlink rejected a descendant of an 8.3 short-name path: %v", err)
	}
}

// TestRejectFixtureSymlinkStillRejectsReparsePointAncestor is the control: the
// short-name fix must not weaken the guard. A junction is a reparse point and its
// ancestor must still be refused.
func TestRejectFixtureSymlinkStillRejectsReparsePointAncestor(t *testing.T) {
	base, err := os.MkdirTemp("", "data-vault-junction-")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(base) })

	target := filepath.Join(base, "real-target")
	if err := os.MkdirAll(filepath.Join(target, "nested"), 0o700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	link := filepath.Join(base, "junction-link")
	if out, err := exec.Command("cmd", "/c", "mklink", "/J", link, target).CombinedOutput(); err != nil {
		t.Skipf("cannot create a junction on this host (%v): %s", err, out)
	}

	err = rejectFixtureSymlink(filepath.Join(link, "nested"))
	if err == nil {
		t.Fatal("rejectFixtureSymlink accepted a descendant of a junction; the reparse-point guard was weakened")
	}
}
