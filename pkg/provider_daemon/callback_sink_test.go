package provider_daemon

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

// TestFileCallbackSinkRejectsTraversalID drives hostile callback IDs through the
// real Submit entry point and asserts that (a) the call fails and (b) nothing is
// written outside the sink directory. Callback IDs come from Waldur, so a
// payload must never be able to steer the write target.
func TestFileCallbackSinkRejectsTraversalID(t *testing.T) {
	root := t.TempDir()
	sinkDir := filepath.Join(root, "sink")
	require.NoError(t, os.MkdirAll(sinkDir, 0o700))

	// A canary outside the sink directory that must not be touched, plus a
	// sentinel whose contents prove no overwrite happened.
	canary := filepath.Join(root, "canary.txt")
	require.NoError(t, os.WriteFile(canary, []byte("original"), 0o600))

	hostileIDs := []string{
		"../canary.txt",
		"../../etc/passwd",
		`..\canary.txt`,
		"nested/evil",
		`nested\evil`,
		"/etc/passwd",
		"..",
		".",
		"",
		"with\x00nul",
		"with\nnewline",
		"with\ttab",
	}

	for _, id := range hostileIDs {
		t.Run(id, func(t *testing.T) {
			sink := NewFileCallbackSink(sinkDir)
			err := sink.Submit(context.Background(), &marketplace.WaldurCallback{ID: id})
			require.Error(t, err, "hostile callback ID %q must be rejected", id)

			// The canary outside the sink directory must be untouched.
			got, readErr := os.ReadFile(canary)
			require.NoError(t, readErr)
			assert.Equal(t, "original", string(got), "canary was overwritten by ID %q", id)

			// And no file may have been created outside the sink directory.
			entries, readErr := os.ReadDir(root)
			require.NoError(t, readErr)
			for _, e := range entries {
				assert.True(t, e.IsDir() || e.Name() == "canary.txt",
					"unexpected entry %q created outside the sink dir by ID %q", e.Name(), id)
			}
		})
	}
}

// TestFileCallbackSinkAcceptsSafeID is the positive control: a plain Waldur
// callback ID still writes exactly one JSON file inside the sink directory.
func TestFileCallbackSinkAcceptsSafeID(t *testing.T) {
	root := t.TempDir()
	sinkDir := filepath.Join(root, "sink")

	sink := NewFileCallbackSink(sinkDir)
	require.NoError(t, sink.Submit(context.Background(), &marketplace.WaldurCallback{
		ID: "cb-1234-5678",
	}))

	written := filepath.Join(sinkDir, "cb-1234-5678.json")
	data, err := os.ReadFile(written)
	require.NoError(t, err, "the callback should have been written inside the sink dir")
	assert.Contains(t, string(data), "cb-1234-5678")
	assert.NotContains(t, string(data), ".tmp")
}

// TestCallbackFileName pins the accepted and rejected shapes directly.
func TestCallbackFileName(t *testing.T) {
	ok := []string{"cb-1", "a.b", "UPPER_lower-9", "with space"}
	for _, id := range ok {
		name, err := callbackFileName(id)
		require.NoError(t, err, "id %q should be accepted", id)
		assert.Equal(t, id+".json", name)
	}

	bad := []string{"", ".", "..", "a/b", `a\b`, "/abs", `C:\abs`, "a\x00b", "a\nb", "a\x1fb"}
	for _, id := range bad {
		_, err := callbackFileName(id)
		require.Error(t, err, "id %q should be rejected", id)
		assert.ErrorIs(t, err, ErrUnsafeCallbackID)
	}
}

// TestFileCallbackSinkRejectsEmptyDir documents that a sink constructed without
// a directory fails closed rather than writing relative to the process cwd.
func TestFileCallbackSinkRejectsEmptyDir(t *testing.T) {
	sink := NewFileCallbackSink("   ")
	err := sink.Submit(context.Background(), &marketplace.WaldurCallback{ID: "cb-1"})
	require.Error(t, err)
}
