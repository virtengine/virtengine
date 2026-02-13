// Package security provides security utilities for path validation and safe file operations.
package security

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Common error messages for path traversal detection
var (
	ErrPathTraversal    = fmt.Errorf("path traversal detected")
	ErrPathNotAllowed   = fmt.Errorf("path not in allowed directories")
	ErrInvalidPath      = fmt.Errorf("invalid path")
	ErrPathNotAbsolute  = fmt.Errorf("path must be absolute")
	ErrInvalidExtension = fmt.Errorf("invalid file extension")
)

// PathValidator provides path validation to prevent directory traversal attacks.
type PathValidator struct {
	baseDir           string
	allowedDirs       []string
	allowedExtensions []string
	requireAbsolute   bool
}

// PathValidatorOption configures a PathValidator.
type PathValidatorOption func(*PathValidator)

// WithAllowedDirs sets allowed directories for path validation.
func WithAllowedDirs(dirs ...string) PathValidatorOption {
	return func(v *PathValidator) {
		v.allowedDirs = append(v.allowedDirs, dirs...)
	}
}

// WithAllowedExtensions sets allowed file extensions.
func WithAllowedExtensions(exts ...string) PathValidatorOption {
	return func(v *PathValidator) {
		v.allowedExtensions = append(v.allowedExtensions, exts...)
	}
}

// WithRequireAbsolute requires paths to be absolute.
func WithRequireAbsolute(require bool) PathValidatorOption {
	return func(v *PathValidator) {
		v.requireAbsolute = require
	}
}

// NewPathValidator creates a new PathValidator with the specified base directory.
func NewPathValidator(baseDir string, opts ...PathValidatorOption) *PathValidator {
	v := &PathValidator{
		baseDir:     baseDir,
		allowedDirs: []string{baseDir},
	}
	for _, opt := range opts {
		opt(v)
	}
	return v
}

// ValidatePath ensures the path doesn't escape allowed directories and is safe.
// It checks for traversal sequences, resolves symlinks, and validates against allowed dirs.
func (v *PathValidator) ValidatePath(path string) error {
	if path == "" {
		return ErrInvalidPath
	}

	// Clean the path first
	cleanPath := filepath.Clean(path)

	// Check for traversal attempts in the original path
	if ContainsTraversalSequence(path) {
		return fmt.Errorf("%w: %s", ErrPathTraversal, path)
	}

	// Check for traversal in cleaned path (handles cases like "foo/../..")
	if ContainsTraversalSequence(cleanPath) {
		return fmt.Errorf("%w: %s", ErrPathTraversal, path)
	}

	// If require absolute and path is not absolute, error
	if v.requireAbsolute && !filepath.IsAbs(cleanPath) {
		return fmt.Errorf("%w: %s", ErrPathNotAbsolute, path)
	}

	// Resolve to absolute path for directory containment check
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("cannot resolve path: %w", err)
	}

	// Resolve symlinks for existing paths so we can enforce allowed directories.
	if info, err := os.Lstat(absPath); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			linkTarget, err := os.Readlink(absPath)
			if err != nil {
				return fmt.Errorf("%w: %s", ErrInvalidPath, path)
			}
			if !filepath.IsAbs(linkTarget) {
				linkTarget = filepath.Join(filepath.Dir(absPath), linkTarget)
			}
			resolved, err := filepath.Abs(filepath.Clean(linkTarget))
			if err != nil {
				return fmt.Errorf("%w: %s", ErrInvalidPath, path)
			}
			absPath = resolved
		}

		realPath, err := filepath.EvalSymlinks(absPath)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrInvalidPath, path)
		}
		absPath = realPath
	}

	// Check if path is within allowed directories
	if len(v.allowedDirs) > 0 {
		allowed := false
		for _, dir := range v.allowedDirs {
			absDir, err := filepath.Abs(dir)
			if err != nil {
				continue
			}
			// Normalize the directory path with separator
			absDir = filepath.Clean(absDir) + string(filepath.Separator)
			absPathNorm := filepath.Clean(absPath)

			// Check if the path is exactly the allowed dir or starts with it
			if absPathNorm == filepath.Clean(absDir[:len(absDir)-1]) ||
				strings.HasPrefix(absPathNorm+string(filepath.Separator), absDir) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("%w: %s", ErrPathNotAllowed, path)
		}
	}

	// Check file extension if restrictions are set
	if len(v.allowedExtensions) > 0 {
		ext := strings.ToLower(filepath.Ext(cleanPath))
		allowed := false
		for _, allowedExt := range v.allowedExtensions {
			if strings.ToLower(allowedExt) == ext {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("%w: got %s, expected one of %v", ErrInvalidExtension, ext, v.allowedExtensions)
		}
	}

	return nil
}

// ValidateAndClean validates the path and returns the cleaned absolute path.
func (v *PathValidator) ValidateAndClean(path string) (string, error) {
	if err := v.ValidatePath(path); err != nil {
		return "", err
	}
	return filepath.Abs(filepath.Clean(path))
}

// ContainsTraversalSequence checks if a path contains directory traversal sequences.
// This includes various encoding forms of "../" and "..\\".
func ContainsTraversalSequence(path string) bool {
	// Check for standard traversal patterns
	patterns := []string{
		"..",         // Standard parent directory
		"%2e%2e",     // URL encoded ..
		"%2E%2E",     // URL encoded .. (uppercase)
		"%252e%252e", // Double URL encoded
		"..%2f",      // Mixed encoding
		"..%5c",      // Mixed encoding (Windows)
		"%2e%2e%2f",  // URL encoded ../
		"%2e%2e%5c",  // URL encoded ..\
		"....//",     // Extended traversal
		"....\\",     // Extended traversal (Windows)
		"\x00",       // Null byte injection
	}

	lowerPath := strings.ToLower(path)
	for _, pattern := range patterns {
		if strings.Contains(lowerPath, strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}

// ValidatePathWithinBase ensures targetPath resolves within basePath.
// It rejects traversal sequences and paths that escape the base directory.
func ValidatePathWithinBase(basePath, targetPath string) error {
	if basePath == "" || targetPath == "" {
		return ErrInvalidPath
	}

	if ContainsTraversalSequence(basePath) || ContainsTraversalSequence(targetPath) {
		return fmt.Errorf("%w: %s", ErrPathTraversal, targetPath)
	}

	absBase, err := filepath.Abs(filepath.Clean(basePath))
	if err != nil {
		return fmt.Errorf("invalid base path: %w", err)
	}

	absTarget, err := filepath.Abs(filepath.Clean(targetPath))
	if err != nil {
		return fmt.Errorf("invalid target path: %w", err)
	}

	absBase = filepath.Clean(absBase)
	absTarget = filepath.Clean(absTarget)

	if absTarget == absBase {
		return nil
	}

	relPath, err := filepath.Rel(absBase, absTarget)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrPathTraversal, targetPath)
	}

	if relPath == ".." || strings.HasPrefix(relPath, ".."+string(filepath.Separator)) || filepath.IsAbs(relPath) {
		return fmt.Errorf("%w: %s", ErrPathTraversal, targetPath)
	}

	return nil
}

// CleanPathWithinBase validates and returns a cleaned absolute target path.
func CleanPathWithinBase(basePath, targetPath string) (string, error) {
	if err := ValidatePathWithinBase(basePath, targetPath); err != nil {
		return "", err
	}

	absTarget, err := filepath.Abs(filepath.Clean(targetPath))
	if err != nil {
		return "", fmt.Errorf("invalid target path: %w", err)
	}

	return absTarget, nil
}

// ValidateRelativePath rejects absolute paths and traversal sequences.
func ValidateRelativePath(path string) error {
	if path == "" {
		return ErrInvalidPath
	}

	if filepath.IsAbs(path) {
		return fmt.Errorf("%w: %s", ErrPathNotAllowed, path)
	}

	if ContainsTraversalSequence(path) {
		return fmt.Errorf("%w: %s", ErrPathTraversal, path)
	}

	return nil
}

// ValidateCLIPath validates a path from CLI input without directory restrictions.
// It only checks for traversal sequences and null bytes.
func ValidateCLIPath(path string) error {
	if path == "" {
		return ErrInvalidPath
	}

	if ContainsTraversalSequence(path) {
		return fmt.Errorf("%w: %s", ErrPathTraversal, path)
	}

	// Check for null bytes
	if strings.ContainsRune(path, '\x00') {
		return fmt.Errorf("%w: null byte in path", ErrPathTraversal)
	}

	return nil
}

// ValidateCLIPathWithExtension validates a CLI path and checks for allowed extensions.
func ValidateCLIPathWithExtension(path string, allowedExtensions ...string) error {
	if err := ValidateCLIPath(path); err != nil {
		return err
	}

	if len(allowedExtensions) > 0 {
		ext := strings.ToLower(filepath.Ext(path))
		for _, allowed := range allowedExtensions {
			if strings.ToLower(allowed) == ext {
				return nil
			}
		}
		return fmt.Errorf("%w: got %s, expected one of %v", ErrInvalidExtension, ext, allowedExtensions)
	}

	return nil
}

// SafeReadFile reads a file after validating the path for traversal attacks.
func SafeReadFile(path string) ([]byte, error) {
	if err := ValidateCLIPath(path); err != nil {
		return nil, err
	}
	cleanPath := filepath.Clean(path)
	return os.ReadFile(cleanPath) // #nosec G304 -- path validated above
}

// SafeReadFileWithExtension reads a file after validating path and extension.
func SafeReadFileWithExtension(path string, allowedExtensions ...string) ([]byte, error) {
	if err := ValidateCLIPathWithExtension(path, allowedExtensions...); err != nil {
		return nil, err
	}
	cleanPath := filepath.Clean(path)
	return os.ReadFile(cleanPath) // #nosec G304 -- path validated above
}

// SafeOpen opens a file after validating the path for traversal attacks.
func SafeOpen(path string) (*os.File, error) {
	if err := ValidateCLIPath(path); err != nil {
		return nil, err
	}
	cleanPath := filepath.Clean(path)
	return os.Open(cleanPath) // #nosec G304 -- path validated above
}

// SafeOpenWithExtension opens a file after validating path and extension.
func SafeOpenWithExtension(path string, allowedExtensions ...string) (*os.File, error) {
	if err := ValidateCLIPathWithExtension(path, allowedExtensions...); err != nil {
		return nil, err
	}
	cleanPath := filepath.Clean(path)
	return os.Open(cleanPath) // #nosec G304 -- path validated above
}
