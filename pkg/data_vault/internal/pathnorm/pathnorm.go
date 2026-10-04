// Package pathnorm normalises filesystem paths for security comparisons.
package pathnorm

// LongPath returns path with every Windows 8.3 short-name component expanded to
// its long form. On platforms without short names it returns path unchanged.
//
// It exists so path-identity checks do not confuse a short-name spelling with a
// reparse point: filepath.EvalSymlinks expands 8.3 names *and* resolves
// symlinks/junctions, so comparing a path with its EvalSymlinks result treats a
// legitimate ancestor such as C:\Users\RUNNER~1\AppData\Local\Temp as a
// reparse point. Normalising the comparison operand with the platform's
// long-name expansion isolates the short-name difference from the reparse
// difference.
//
// LongPath never resolves symlinks, junctions or any other reparse point.
func LongPath(path string) (string, error) {
	return longPath(path)
}
