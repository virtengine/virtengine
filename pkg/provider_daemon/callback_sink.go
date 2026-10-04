package provider_daemon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

// CallbackSink submits Waldur callbacks to a target system.
type CallbackSink interface {
	Submit(ctx context.Context, callback *marketplace.WaldurCallback) error
}

// FileCallbackSink writes callbacks to disk for external submission.
type FileCallbackSink struct {
	dir string
}

// NewFileCallbackSink creates a file-based callback sink.
func NewFileCallbackSink(dir string) *FileCallbackSink {
	return &FileCallbackSink{dir: dir}
}

// ErrUnsafeCallbackID is returned when a Waldur callback ID cannot be used as a
// file name inside the sink directory.
var ErrUnsafeCallbackID = errors.New("unsafe callback ID")

// callbackFileName converts a Waldur callback ID into the file name used inside
// the sink directory.
//
// Callback IDs arrive from Waldur, i.e. from outside this process, and are
// interpolated into a file path. An ID containing a path separator, "..", a
// drive letter, a NUL byte or a control character would otherwise let a hostile
// or corrupt payload write outside the sink directory (CWE-22). Such IDs are
// rejected rather than sanitised, so a bad payload fails loudly instead of
// silently landing somewhere unexpected.
func callbackFileName(id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("%w: id is empty", ErrUnsafeCallbackID)
	}
	if id == "." || id == ".." {
		return "", fmt.Errorf("%w: %q is a directory reference", ErrUnsafeCallbackID, id)
	}
	if strings.ContainsAny(id, `/\`) {
		return "", fmt.Errorf("%w: %q contains a path separator", ErrUnsafeCallbackID, id)
	}
	if strings.ContainsRune(id, 0) {
		return "", fmt.Errorf("%w: %q contains a NUL byte", ErrUnsafeCallbackID, id)
	}
	if filepath.IsAbs(id) || filepath.VolumeName(id) != "" {
		return "", fmt.Errorf("%w: %q is an absolute path", ErrUnsafeCallbackID, id)
	}
	for _, r := range id {
		if r < 0x20 || r == 0x7f {
			return "", fmt.Errorf("%w: %q contains a control character", ErrUnsafeCallbackID, id)
		}
	}
	return id + ".json", nil
}

// ensureWithinDir re-checks with filepath.Rel that path stays inside dir. The
// components of path are already known to be a bare file name, so this is
// defence in depth: it fails closed if the sink directory itself is odd (a
// relative path, a symlinked base, or an empty string).
func ensureWithinDir(dir, path string) error {
	if strings.TrimSpace(dir) == "" {
		return fmt.Errorf("callback sink directory is required")
	}
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return fmt.Errorf("resolve callback path: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return fmt.Errorf("%w: %q escapes the callback directory", ErrUnsafeCallbackID, path)
	}
	return nil
}

// Submit writes the callback as JSON to a file.
func (s *FileCallbackSink) Submit(ctx context.Context, callback *marketplace.WaldurCallback) error {
	if callback == nil {
		return fmt.Errorf("callback is nil")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	filename, err := callbackFileName(callback.ID)
	if err != nil {
		return err
	}
	path := filepath.Join(s.dir, filename)
	if err := ensureWithinDir(s.dir, path); err != nil {
		return err
	}
	tmp := path + ".tmp"

	if err := os.MkdirAll(s.dir, 0o700); err != nil { // #nosec G703 -- the callback ID is validated by callbackFileName (rejects separators, '..', absolute paths, NUL and control characters) and the result is re-checked against the sink directory with filepath.Rel before this call
		return fmt.Errorf("create callback dir: %w", err)
	}

	payload := struct {
		Callback *marketplace.WaldurCallback `json:"callback"`
		Written  time.Time                   `json:"written_at"`
	}{
		Callback: callback,
		Written:  time.Now().UTC(),
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal callback: %w", err)
	}

	if err := os.WriteFile(tmp, data, 0o600); err != nil { // #nosec G703 -- the callback ID is validated by callbackFileName (rejects separators, '..', absolute paths, NUL and control characters) and the result is re-checked against the sink directory with filepath.Rel before this call
		return fmt.Errorf("write callback tmp: %w", err)
	}

	if err := os.Rename(tmp, path); err != nil { // #nosec G703 -- the callback ID is validated by callbackFileName (rejects separators, '..', absolute paths, NUL and control characters) and the result is re-checked against the sink directory with filepath.Rel before this call
		return fmt.Errorf("rename callback: %w", err)
	}

	return nil
}
