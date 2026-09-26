package security

import (
	"os"
	"path/filepath"
	"strings"
)

// resolveExistingPrefix canonicalizes the deepest existing ancestor of path
// with filepath.EvalSymlinks and re-appends the components that do not exist
// yet, so both a to-be-created file and its allowed base directory end up in the
// same canonical form.
//
// This matters on Windows, where EvalSymlinks expands 8.3 short components
// (C:\Users\RUNNER~1\... becomes C:\Users\runneradmin\...). If only the existing
// portion were resolved, a destination validated before it is written would keep
// its short form and fail the containment check against an expanded base dir.
func resolveExistingPrefix(path string) (string, error) {
	remainder := ""
	current := path
	for {
		resolved, err := filepath.EvalSymlinks(current)
		if err == nil {
			if remainder == "" {
				return resolved, nil
			}
			return filepath.Join(resolved, remainder), nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(current)
		if parent == current {
			// Reached the volume root without finding anything that exists.
			return path, nil
		}
		base := filepath.Base(current)
		if remainder == "" {
			remainder = base
		} else {
			remainder = filepath.Join(base, remainder)
		}
		current = parent
	}
}

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
