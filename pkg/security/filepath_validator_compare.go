package security

import (
	"path/filepath"
	"strings"
)

// canonicalDir resolves dir to an absolute, symlink-free, comparison-safe form.
// It returns false when dir cannot be resolved to a usable absolute path.
//
// The candidate path in ValidatePath is resolved with filepath.EvalSymlinks, so
// the allowed directories have to be resolved the same way before they can be
// compared. On Windows that matters concretely: EvalSymlinks expands 8.3 short
// components (C:\Users\RUNNER~1\...) to their long form, so an allowed directory
// left un-canonicalized would compare unequal to every candidate underneath it
// and reject valid paths.
func canonicalDir(dir string) (string, bool) {
	absDir, err := filepath.Abs(filepath.Clean(dir))
	if err != nil {
		return "", false
	}
	if resolved, err := filepath.EvalSymlinks(absDir); err == nil {
		absDir = resolved
	}
	return normalizeForCompare(absDir), true
}

// pathWithinDir reports whether path is dir itself or is nested inside dir.
// Both arguments must already be normalized by normalizeForCompare.
func pathWithinDir(path, dir string) bool {
	if path == dir {
		return true
	}
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	// A relative path that starts with ".." escapes dir; anything else is nested.
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
