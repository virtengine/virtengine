package security

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSecureKeyStorage(t *testing.T) {
	tmpDir := t.TempDir()

	storage, err := NewSecureKeyStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewSecureKeyStorage failed: %v", err)
	}

	t.Run("GetKeyFilePath", func(t *testing.T) {
		tests := []struct {
			keyID   string
			wantErr bool
		}{
			{"mykey", false},
			{"key-123", false},
			{"../secret", true},
			{"foo/bar", true},
			{"..", true},
			{".", true},
		}

		for _, tt := range tests {
			t.Run(tt.keyID, func(t *testing.T) {
				path, err := storage.GetKeyFilePath(tt.keyID)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetKeyFilePath(%q) error = %v, wantErr %v", tt.keyID, err, tt.wantErr)
				}
				if err == nil && path == "" {
					t.Error("GetKeyFilePath should return non-empty path on success")
				}
			})
		}
	})

	t.Run("WriteAndReadSecureFile", func(t *testing.T) {
		path := filepath.Join(tmpDir, "test.key.json")
		testData := []byte(`{"key": "value"}`)

		// Write file
		if err := storage.WriteSecureFile(path, testData); err != nil {
			t.Fatalf("WriteSecureFile failed: %v", err)
		}

		// Verify permissions on non-Windows systems
		if runtime.GOOS != "windows" {
			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("Failed to stat file: %v", err)
			}
			if perm := info.Mode().Perm(); perm != KeyFilePerm {
				t.Errorf("File permissions = %o, want %o", perm, KeyFilePerm)
			}
		}

		// Read file
		data, err := storage.ReadSecureFile(path)
		if err != nil {
			t.Fatalf("ReadSecureFile failed: %v", err)
		}
		if string(data) != string(testData) {
			t.Errorf("Data mismatch: got %q, want %q", data, testData)
		}
	})

	t.Run("RejectPathTraversal", func(t *testing.T) {
		// Try to write outside directory
		err := storage.WriteSecureFile("../evil.txt", []byte("malicious"))
		if err == nil {
			t.Error("WriteSecureFile should reject path traversal")
		}

		// Try to read outside directory
		_, err = storage.ReadSecureFile("../etc/passwd")
		if err == nil {
			t.Error("ReadSecureFile should reject path traversal")
		}
	})
}

func TestStateFileValidator(t *testing.T) {
	tmpDir := t.TempDir()

	validator, err := NewStateFileValidator(tmpDir)
	if err != nil {
		t.Fatalf("NewStateFileValidator failed: %v", err)
	}

	t.Run("GetStateFilePath", func(t *testing.T) {
		tests := []struct {
			filename string
			wantErr  bool
		}{
			{"state.json", false},
			{"checkpoint.dat", false},
			{"../secret.json", true},
			{"foo/bar.json", true},
		}

		for _, tt := range tests {
			t.Run(tt.filename, func(t *testing.T) {
				path, err := validator.GetStateFilePath(tt.filename)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetStateFilePath(%q) error = %v, wantErr %v", tt.filename, err, tt.wantErr)
				}
				if err == nil && path == "" {
					t.Error("GetStateFilePath should return non-empty path on success")
				}
			})
		}
	})

	t.Run("SafeWriteAndReadStateFile", func(t *testing.T) {
		path := filepath.Join(tmpDir, "test.state")
		testData := []byte(`{"checkpoint": 12345}`)

		// Write file
		if err := validator.SafeWriteStateFile(path, testData); err != nil {
			t.Fatalf("SafeWriteStateFile failed: %v", err)
		}

		// Read file
		data, err := validator.SafeReadStateFile(path)
		if err != nil {
			t.Fatalf("SafeReadStateFile failed: %v", err)
		}
		if string(data) != string(testData) {
			t.Errorf("Data mismatch: got %q, want %q", data, testData)
		}
	})

	t.Run("AtomicWrite", func(t *testing.T) {
		path := filepath.Join(tmpDir, "atomic.state")

		// Write initial data
		if err := validator.SafeWriteStateFile(path, []byte("v1")); err != nil {
			t.Fatalf("Initial write failed: %v", err)
		}

		// Write new data
		if err := validator.SafeWriteStateFile(path, []byte("v2")); err != nil {
			t.Fatalf("Second write failed: %v", err)
		}

		// Verify latest data
		data, err := validator.SafeReadStateFile(path)
		if err != nil {
			t.Fatalf("Read failed: %v", err)
		}
		if string(data) != "v2" {
			t.Errorf("Data = %q, want %q", data, "v2")
		}

		// Verify no temp files left behind
		matches, _ := filepath.Glob(filepath.Join(tmpDir, ".tmp.*"))
		if len(matches) > 0 {
			t.Errorf("Temp files not cleaned up: %v", matches)
		}
	})

	t.Run("RejectPathTraversal", func(t *testing.T) {
		// Try to write outside directory
		err := validator.SafeWriteStateFile("../evil.state", []byte("bad"))
		if err == nil {
			t.Error("SafeWriteStateFile should reject path traversal")
		}

		// Try to read outside directory
		_, err = validator.SafeReadStateFile("../etc/passwd")
		if err == nil {
			t.Error("SafeReadStateFile should reject path traversal")
		}
	})
}

func TestEnsureSecurePermissions(t *testing.T) {
	// Skip permission tests on Windows
	if runtime.GOOS == "windows" {
		t.Skip("Permission tests not applicable on Windows")
	}

	tmpDir := t.TempDir()
	storage, err := NewSecureKeyStorage(tmpDir)
	if err != nil {
		t.Fatalf("NewSecureKeyStorage failed: %v", err)
	}

	t.Run("FixInsecurePermissions", func(t *testing.T) {
		path := filepath.Join(tmpDir, "insecure.key")
		//nolint:gosec // G306: intentionally insecure for testing permission fix
		if err := os.WriteFile(path, []byte("key"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Verify initially insecure
		info, _ := os.Stat(path)
		if info.Mode().Perm() == KeyFilePerm {
			t.Skip("File already has secure permissions")
		}

		// Fix permissions
		if err := storage.EnsureSecurePermissions(path); err != nil {
			t.Fatalf("EnsureSecurePermissions failed: %v", err)
		}

		// Verify fixed
		info, _ = os.Stat(path)
		if perm := info.Mode().Perm(); perm != KeyFilePerm {
			t.Errorf("Permissions = %o, want %o", perm, KeyFilePerm)
		}
	})
}
