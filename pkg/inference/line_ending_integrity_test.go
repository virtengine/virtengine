package inference

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestEmbeddedFeatureSchemaKeepsLineEndingPolicy guards the invariant that the
// canonical feature schema, whose raw bytes feature_parity_test.go digests, is
// byte-stable across platforms.
//
// Regression guard: feature_parity_test.go hashes the schema file as checked out
// on disk and compares it against the fixture's pinned digest of the LF blob.
// With `core.autocrlf=true` a Windows checkout rewrites the artifact to CRLF, the
// raw-byte digest shifts to 7de2340d..., and
// "Windows Native Build and Unit Tests" (ci.yaml) reddens with
// `schema hash: got 7de2340d..., want 010cf22b...`. `.gitattributes` pins
// `/pkg/inference/schema/*.json text eol=lf` precisely because git stores, and
// the go:embed directive carries, the LF form. This mirrors the VEID ZK params
// precedent in x/veid/zk/params/sidecar_integrity_test.go.
func TestEmbeddedFeatureSchemaKeepsLineEndingPolicy(t *testing.T) {
	if len(canonicalFeatureSchemaJSON) == 0 {
		t.Fatal("embedded canonical feature schema is empty")
	}
	if strings.ContainsRune(string(canonicalFeatureSchemaJSON), '\r') {
		t.Fatal("embedded canonical feature schema contains CR bytes; the eol=lf .gitattributes rule must not be removed")
	}

	root := inferenceRepositoryRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, ".gitattributes"))
	if err != nil {
		t.Fatalf("read .gitattributes: %v", err)
	}
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.TrimSpace(line) == "/pkg/inference/schema/*.json text eol=lf" {
			return
		}
	}
	t.Fatal(".gitattributes is missing \"/pkg/inference/schema/*.json text eol=lf\"; " +
		"without it a Windows checkout rewrites the schema to CRLF and the pinned schema digest stops matching " +
		"(Windows Native Build and Unit Tests: schema hash: got 7de2340d..., want 010cf22b...)")
}

// inferenceRepositoryRoot walks up from the package directory to the module root,
// identified by go.mod alongside .git (a directory in a clone, a file in a
// worktree).
func inferenceRepositoryRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		_, goModErr := os.Stat(filepath.Join(dir, "go.mod"))
		_, gitErr := os.Stat(filepath.Join(dir, ".git"))
		if goModErr == nil && gitErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Skipf("repository root not found above %s", dir)
		}
		dir = parent
	}
}
