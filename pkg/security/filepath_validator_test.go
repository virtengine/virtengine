package security

import (
	"path/filepath"
	"testing"
)

func TestValidatePathWithinBase(t *testing.T) {
	baseDir := t.TempDir()

	t.Run("accepts path within base", func(t *testing.T) {
		target := filepath.Join(baseDir, "sub", "file.txt")
		if err := ValidatePathWithinBase(baseDir, target); err != nil {
			t.Fatalf("expected valid path, got error: %v", err)
		}
	})

	t.Run("accepts base directory", func(t *testing.T) {
		if err := ValidatePathWithinBase(baseDir, baseDir); err != nil {
			t.Fatalf("expected base path to be valid, got error: %v", err)
		}
	})

	t.Run("rejects path outside base", func(t *testing.T) {
		outside := filepath.Join(filepath.Dir(baseDir), "escape")
		if err := ValidatePathWithinBase(baseDir, outside); err == nil {
			t.Fatal("expected error for path outside base directory")
		}
	})
}

func TestCleanPathWithinBase(t *testing.T) {
	baseDir := t.TempDir()
	target := filepath.Join(baseDir, "sub", "file.txt")

	cleaned, err := CleanPathWithinBase(baseDir, target)
	if err != nil {
		t.Fatalf("expected clean path, got error: %v", err)
	}
	if cleaned == "" || !filepath.IsAbs(cleaned) {
		t.Fatalf("expected absolute cleaned path, got: %q", cleaned)
	}

	outside := filepath.Join(filepath.Dir(baseDir), "escape")
	if _, err := CleanPathWithinBase(baseDir, outside); err == nil {
		t.Fatal("expected error for path outside base directory")
	}
}

func TestValidateRelativePath(t *testing.T) {
	t.Run("accepts relative path", func(t *testing.T) {
		if err := ValidateRelativePath(filepath.Join("sub", "file.txt")); err != nil {
			t.Fatalf("expected valid relative path, got error: %v", err)
		}
	})

	t.Run("rejects absolute path", func(t *testing.T) {
		absPath := filepath.Clean(t.TempDir())
		if err := ValidateRelativePath(absPath); err == nil {
			t.Fatal("expected error for absolute path")
		}
	})

	t.Run("rejects traversal path", func(t *testing.T) {
		traversal := filepath.Join("..", "escape")
		if err := ValidateRelativePath(traversal); err == nil {
			t.Fatal("expected error for traversal path")
		}
	})
}
