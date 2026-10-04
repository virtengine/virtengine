//go:build windows

package security

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// windowsShortPath returns the 8.3 short form of path, skipping the test when
// short-name generation is unavailable or already identical.
func windowsShortPath(t *testing.T, path string) string {
	t.Helper()
	//nolint:gosec // G204: path is the test's own t.TempDir() value, not external input
	out, err := exec.Command("cmd", "/c", `for %I in (`+path+`) do @echo %~sI`).Output()
	if err != nil {
		t.Skipf("cannot resolve 8.3 short path: %v", err)
	}
	short := strings.TrimSpace(string(out))
	if short == "" || strings.EqualFold(short, path) {
		t.Skip("8.3 short names are not generated on this volume")
	}
	return short
}

// TestPathValidatorAcceptsShortFormBaseDir is the Windows regression test for
// the CI failure "path not in allowed directories" across the whole
// PathValidator/ValidateAndClean/secure-file cluster.
//
// The candidate path is resolved with filepath.EvalSymlinks, which expands 8.3
// short components (C:\Users\RUNNER~1\...) to their long form. The allowed
// directories were previously left un-canonicalized, so a base dir expressed in
// short form compared unequal to every candidate beneath it and rejected valid
// files. Both sides are canonicalized now.
func TestPathValidatorAcceptsShortFormBaseDir(t *testing.T) {
	long := filepath.Join(t.TempDir(), "veLongDirNameRegression")
	subDir := filepath.Join(long, "subdir")
	if err := os.MkdirAll(subDir, 0o750); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(subDir, "test.json")
	if err := os.WriteFile(file, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	short := windowsShortPath(t, long)
	t.Logf("long=%s short=%s", long, short)

	validator := NewPathValidator(short, WithAllowedExtensions(".json"))
	if err := validator.ValidatePath(file); err != nil {
		t.Errorf("ValidatePath(long form) with 8.3 base dir = %v, want nil", err)
	}

	// Only the long form of the candidate is asserted. An 8.3 short form
	// truncates the extension to three characters by Windows design
	// ("test.json" -> "TEST.JSO"), so the extension allowlist legitimately
	// rejects it; real callers pass long-form paths.

	// ValidateAndClean resolves through the same check.
	if _, err := validator.ValidateAndClean(file); err != nil {
		t.Errorf("ValidateAndClean(long form) with 8.3 base dir = %v, want nil", err)
	}

	// A destination that does not exist yet must validate too. This is the
	// shape WriteSecureFile and SafeWriteStateFile use, and it was the second
	// half of the CI failure: the candidate kept its short form because nothing
	// existed to resolve, while the base dir had been expanded.
	notYet := filepath.Join(long, "subdir", "future.json")
	if _, err := os.Lstat(notYet); !os.IsNotExist(err) {
		t.Fatalf("expected %s to not exist", notYet)
	}
	if err := validator.ValidatePath(notYet); err != nil {
		t.Errorf("ValidatePath(not-yet-created) with 8.3 base dir = %v, want nil", err)
	}
}

// TestPathValidatorRejectsSiblingWithSharedPrefix guards the security property
// the old prefix comparison got right by accident: a directory whose name
// starts with the allowed dir's name is NOT inside it.
func TestPathValidatorRejectsSiblingWithSharedPrefix(t *testing.T) {
	base := t.TempDir()
	allowed := filepath.Join(base, "data")
	sibling := filepath.Join(base, "data-secret")
	for _, dir := range []string{allowed, sibling} {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			t.Fatal(err)
		}
	}
	secret := filepath.Join(sibling, "secret.json")
	if err := os.WriteFile(secret, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}

	validator := NewPathValidator(allowed, WithAllowedExtensions(".json"))
	if err := validator.ValidatePath(secret); err == nil {
		t.Error("ValidatePath accepted a path in a sibling directory sharing a name prefix")
	}
}
