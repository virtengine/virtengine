package provider_daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

	if err := os.MkdirAll(s.dir, 0o700); err != nil {
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

	filename := fmt.Sprintf("%s.json", callback.ID)
	path := filepath.Join(s.dir, filename)
	tmp := path + ".tmp"

	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("write callback tmp: %w", err)
	}

	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename callback: %w", err)
	}

	return nil
}
