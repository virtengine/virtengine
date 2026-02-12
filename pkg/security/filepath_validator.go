package security

import (
	"fmt"
	"path/filepath"
	"strings"
)

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
