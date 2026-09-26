package bundle

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestSafeJoinRejectsEscapingNames(t *testing.T) {
	t.Parallel()

	base := filepath.Join("base", "dir")

	unsafe := []string{
		"",
		"   ",
		"..",
		".",
		"../secret.bin",
		"../../etc/passwd",
		"nested/input.bin",
		`nested\input.bin`,
		"/etc/passwd",
		`C:\Windows\System32\config\SAM`,
		"input.bin/../../escape",
	}

	for _, name := range unsafe {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := SafeJoin(base, name)
			if err == nil {
				t.Fatalf("SafeJoin(%q, %q) = %q, want error", base, name, got)
			}
			if !errors.Is(err, ErrUnsafeBundlePath) {
				t.Fatalf("SafeJoin(%q, %q) error = %v, want ErrUnsafeBundlePath", base, name, err)
			}
		})
	}
}

func TestSafeJoinAcceptsPlainFileNames(t *testing.T) {
	t.Parallel()

	base := filepath.Join("base", "dir")

	safe := []string{
		"input.bin",
		"contribution.bin",
		"response.json",
		"response.json.sha256",
		"phase1_contrib.bin",
	}

	for _, name := range safe {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := SafeJoin(base, name)
			if err != nil {
				t.Fatalf("SafeJoin(%q, %q) returned unexpected error: %v", base, name, err)
			}
			want := filepath.Join(base, name)
			if got != want {
				t.Fatalf("SafeJoin(%q, %q) = %q, want %q", base, name, got, want)
			}
		})
	}
}
