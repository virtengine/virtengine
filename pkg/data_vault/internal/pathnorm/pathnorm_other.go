//go:build !windows

package pathnorm

// longPath is the identity function off Windows: 8.3 short names do not exist.
func longPath(path string) (string, error) {
	return path, nil
}
