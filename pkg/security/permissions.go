package security

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// File permission constants for secure file storage
const (
	// KeyDirPerm is the permission for directories containing sensitive key material
	KeyDirPerm os.FileMode = 0700
	// KeyFilePerm is the permission for files containing sensitive key material
	KeyFilePerm os.FileMode = 0600
	// ConfigDirPerm is the permission for configuration directories
	ConfigDirPerm os.FileMode = 0750
	// ConfigFilePerm is the permission for configuration files
	ConfigFilePerm os.FileMode = 0640
)

// SecureKeyStorage provides secure key file path validation and permission enforcement.
type SecureKeyStorage struct {
	baseDir   string
	validator *PathValidator
}

// NewSecureKeyStorage creates a new SecureKeyStorage with the specified base directory.
func NewSecureKeyStorage(baseDir string) (*SecureKeyStorage, error) {
	if baseDir == "" {
		return nil, fmt.Errorf("base directory is required")
	}

	// Create base directory with secure permissions if it doesn't exist
	if err := os.MkdirAll(baseDir, KeyDirPerm); err != nil {
		return nil, fmt.Errorf("failed to create key storage directory: %w", err)
	}

	// Verify permissions on the directory (skip on Windows)
	if runtime.GOOS != osWindows {
		info, err := os.Stat(baseDir)
		if err != nil {
			return nil, fmt.Errorf("failed to stat key storage directory: %w", err)
		}
		if perm := info.Mode().Perm(); perm&0077 != 0 {
			return nil, fmt.Errorf("key storage directory %s has insecure permissions %o (expected %o)",
				baseDir, perm, KeyDirPerm)
		}
	}

	return &SecureKeyStorage{
		baseDir: baseDir,
		validator: NewPathValidator(baseDir,
			WithAllowedExtensions(".key", ".json", ".pem"),
		),
	}, nil
}

// ValidateKeyPath validates that a key file path is safe and within the storage directory.
func (s *SecureKeyStorage) ValidateKeyPath(keyPath string) error {
	return s.validator.ValidatePath(keyPath)
}

// GetKeyFilePath returns a safe path for a key file within the storage directory.
func (s *SecureKeyStorage) GetKeyFilePath(keyID string) (string, error) {
	// Sanitize key ID to prevent directory traversal
	safeKeyID := filepath.Base(keyID)
	if safeKeyID != keyID || safeKeyID == "." || safeKeyID == ".." {
		return "", fmt.Errorf("%w: invalid key ID: %s", ErrPathTraversal, keyID)
	}

	path := filepath.Join(s.baseDir, safeKeyID+".key.json")
	return path, nil
}

// EnsureSecurePermissions ensures a file has secure permissions.
func (s *SecureKeyStorage) EnsureSecurePermissions(path string) error {
	// Validate path first
	if err := s.validator.ValidatePath(path); err != nil {
		return err
	}

	// On Windows, file permissions work differently
	if runtime.GOOS == "windows" {
		return nil
	}

	// Check current permissions
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	currentPerm := info.Mode().Perm()

	// Determine expected permission based on file type
	expectedPerm := KeyFilePerm
	if info.IsDir() {
		expectedPerm = KeyDirPerm
	}

	// Check if permissions are too permissive
	if currentPerm&0077 != 0 {
		// Try to fix permissions
		if err := os.Chmod(path, expectedPerm); err != nil {
			return fmt.Errorf("failed to set secure permissions on %s (current: %o, expected: %o): %w",
				path, currentPerm, expectedPerm, err)
		}
	}

	return nil
}

// WriteSecureFile writes data to a file with secure permissions.
func (s *SecureKeyStorage) WriteSecureFile(path string, data []byte) error {
	// Validate path
	if err := s.validator.ValidatePath(path); err != nil {
		return err
	}

	cleanPath := filepath.Clean(path)

	// Write file with secure permissions
	if err := os.WriteFile(cleanPath, data, KeyFilePerm); err != nil { // #nosec G306 -- intentional secure permissions
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ReadSecureFile reads data from a file after validating its path.
func (s *SecureKeyStorage) ReadSecureFile(path string) ([]byte, error) {
	// Validate path
	if err := s.validator.ValidatePath(path); err != nil {
		return nil, err
	}

	cleanPath := filepath.Clean(path)

	data, err := os.ReadFile(cleanPath) // #nosec G304 -- path validated above
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

// BaseDir returns the base directory for key storage.
func (s *SecureKeyStorage) BaseDir() string {
	return s.baseDir
}

// StateFileValidator provides validation for state files in a designated directory.
type StateFileValidator struct {
	dataDir   string
	validator *PathValidator
}

// NewStateFileValidator creates a validator for state files within a data directory.
func NewStateFileValidator(dataDir string) (*StateFileValidator, error) {
	if dataDir == "" {
		return nil, fmt.Errorf("data directory is required")
	}

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, ConfigDirPerm); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return &StateFileValidator{
		dataDir: dataDir,
		validator: NewPathValidator(dataDir,
			WithAllowedExtensions(".json", ".state", ".dat", ".db"),
		),
	}, nil
}

// ValidatePath validates a state file path.
func (v *StateFileValidator) ValidatePath(path string) error {
	return v.validator.ValidatePath(path)
}

// ValidateAndClean validates and returns the cleaned path.
func (v *StateFileValidator) ValidateAndClean(path string) (string, error) {
	return v.validator.ValidateAndClean(path)
}

// GetStateFilePath returns a safe path for a state file within the data directory.
func (v *StateFileValidator) GetStateFilePath(filename string) (string, error) {
	// Sanitize filename
	safeName := filepath.Base(filename)
	if safeName != filename || safeName == "." || safeName == ".." {
		return "", fmt.Errorf("%w: invalid filename: %s", ErrPathTraversal, filename)
	}

	path := filepath.Join(v.dataDir, safeName)
	return path, nil
}

// SafeWriteStateFile writes state data atomically with validation.
func (v *StateFileValidator) SafeWriteStateFile(path string, data []byte) error {
	if err := v.validator.ValidatePath(path); err != nil {
		return err
	}

	cleanPath := filepath.Clean(path)

	// Write to temp file first for atomic operation
	dir := filepath.Dir(cleanPath)
	tmp := filepath.Join(dir, ".tmp."+filepath.Base(cleanPath))

	if err := os.WriteFile(tmp, data, ConfigFilePerm); err != nil { // #nosec G306 -- intentional permissions
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := os.Rename(tmp, cleanPath); err != nil {
		os.Remove(tmp) // Clean up temp file on rename failure
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// SafeReadStateFile reads state data with validation.
func (v *StateFileValidator) SafeReadStateFile(path string) ([]byte, error) {
	if err := v.validator.ValidatePath(path); err != nil {
		return nil, err
	}

	cleanPath := filepath.Clean(path)
	return os.ReadFile(cleanPath) // #nosec G304 -- path validated above
}
