package participant

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/virtengine/virtengine/tools/trusted-setup/bundle"
	"github.com/virtengine/virtengine/tools/trusted-setup/transcript"
)

// TestRespondToPhaseBundleRejectsTraversalInputFile proves that a hostile bundle
// cannot use request.json's input_file to read a file outside the bundle
// directory (gosec G304/G703 path traversal).
func TestRespondToPhaseBundleRejectsTraversalInputFile(t *testing.T) {
	t.Parallel()

	traversalNames := []string{
		"../secret.bin",
		"../../secret.bin",
		"nested/input.bin",
		`..\secret.bin`,
		"/etc/passwd",
		"..",
	}

	for _, name := range traversalNames {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			baseDir := t.TempDir()
			bundleDir := filepath.Join(baseDir, "bundle")
			outDir := filepath.Join(baseDir, "out")
			if err := os.MkdirAll(bundleDir, 0o750); err != nil {
				t.Fatalf("create bundle dir: %v", err)
			}

			// A file outside the bundle directory that a hostile bundle must not
			// be able to reach.
			secretPath := filepath.Join(baseDir, "secret.bin")
			if err := os.WriteFile(secretPath, []byte("top secret"), 0o600); err != nil {
				t.Fatalf("write secret: %v", err)
			}

			req := bundle.PhaseRequest{
				SchemaVersion: bundle.PhaseRequestSchema,
				CeremonyID:    "ceremony-1",
				Phase:         transcript.Phase1,
				InputFile:     name,
			}
			if err := bundle.WriteJSON(filepath.Join(bundleDir, "request.json"), req); err != nil {
				t.Fatalf("write request.json: %v", err)
			}

			client := NewClient(&Identity{ID: "participant-1", PublicKey: "pk"}, "")

			_, err := RespondToPhaseBundle(bundleDir, outDir, client)
			if err == nil {
				t.Fatalf("RespondToPhaseBundle accepted traversal input_file %q", name)
			}
			if !errors.Is(err, bundle.ErrUnsafeBundlePath) {
				t.Fatalf("RespondToPhaseBundle(%q) error = %v, want ErrUnsafeBundlePath", name, err)
			}

			// The secret must not have been copied into the output directory.
			if _, statErr := os.Stat(filepath.Join(outDir, "secret.bin")); statErr == nil {
				t.Fatalf("hostile bundle caused secret.bin to be written into the output directory")
			}
			entries, readErr := os.ReadDir(outDir)
			if readErr == nil {
				for _, e := range entries {
					t.Fatalf("unexpected file written to output directory: %s", e.Name())
				}
			}
		})
	}
}

// TestRespondToPhaseBundleAcceptsPlainInputFileName is the positive control: a
// well-formed bundle-relative name passes the path check.
func TestRespondToPhaseBundleAcceptsPlainInputFile(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	bundleDir := filepath.Join(baseDir, "bundle")
	outDir := filepath.Join(baseDir, "out")
	if err := os.MkdirAll(bundleDir, 0o750); err != nil {
		t.Fatalf("create bundle dir: %v", err)
	}

	payload := []byte("phase1-initial-transcript")
	if err := os.WriteFile(filepath.Join(bundleDir, "input.bin"), payload, 0o600); err != nil {
		t.Fatalf("write input: %v", err)
	}

	req := bundle.PhaseRequest{
		SchemaVersion: bundle.PhaseRequestSchema,
		CeremonyID:    "ceremony-1",
		Phase:         transcript.Phase1,
		InputFile:     "input.bin",
		InputHash:     transcript.HashBytes(payload),
	}
	if err := bundle.WriteJSON(filepath.Join(bundleDir, "request.json"), req); err != nil {
		t.Fatalf("write request.json: %v", err)
	}

	client := NewClient(&Identity{ID: "participant-1", PublicKey: "pk"}, "")

	// The transcript payload is not a valid phase1 contribution, so the call may
	// still fail — but it must get past the path validation.
	_, err := RespondToPhaseBundle(bundleDir, outDir, client)
	if errors.Is(err, bundle.ErrUnsafeBundlePath) {
		t.Fatalf("plain file name rejected as unsafe: %v", err)
	}
}
