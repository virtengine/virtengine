package security

import (
	"os"
	"path/filepath"
	"testing"
)

func TestContainsTraversalSequence(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		// Should detect traversal
		{"standard parent", "..", true},
		{"parent with slash", "../", true},
		{"parent in path", "foo/../bar", true},
		{"double parent", "foo/../../bar", true},
		{"URL encoded", "%2e%2e", true},
		{"URL encoded uppercase", "%2E%2E", true},
		{"double URL encoded", "%252e%252e", true},
		{"mixed encoding forward", "..%2f", true},
		{"mixed encoding back", "..%5c", true},
		{"null byte", "foo\x00bar", true},
		{"windows traversal", "..\\", true},
		{"extended traversal", "....//", true},

		// Should not detect traversal
		{"normal path", "foo/bar/baz", false},
		{"absolute path", "/usr/local/bin", false},
		{"dots in filename", "foo.bar.baz", false},
		{"single dot", "./foo", false},
		{"extension with dots", "file.tar.gz", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainsTraversalSequence(tt.path)
			if result != tt.expected {
				t.Errorf("ContainsTraversalSequence(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestValidateCLIPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid relative", "config/file.json", false},
		{"valid absolute", "/etc/config.json", false},
		{"traversal attack", "../../../etc/passwd", true},
		{"null byte", "file\x00.json", true},
		{"empty path", "", true},
		{"URL encoded traversal", "%2e%2e/secret", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCLIPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCLIPath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestValidateCLIPathWithExtension(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		extensions []string
		wantErr    bool
	}{
		{"valid json", "config.json", []string{".json"}, false},
		{"valid JSON uppercase", "CONFIG.JSON", []string{".json"}, false},
		{"invalid extension", "config.yaml", []string{".json"}, true},
		{"no extension check", "config.yaml", nil, false},
		{"multiple extensions allowed", "file.json", []string{".yaml", ".json"}, false},
		{"traversal with valid ext", "../secret.json", []string{".json"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCLIPathWithExtension(tt.path, tt.extensions...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCLIPathWithExtension(%q, %v) error = %v, wantErr %v",
					tt.path, tt.extensions, err, tt.wantErr)
			}
		})
	}
}

func TestPathValidator(t *testing.T) {
	// Create a temp directory structure for testing
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	testFile := filepath.Join(subDir, "test.json")
	if err := os.WriteFile(testFile, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewPathValidator(tmpDir,
		WithAllowedExtensions(".json", ".yaml"),
	)

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid file in allowed dir", testFile, false},
		{"traversal attack", filepath.Join(tmpDir, "..", "etc", "passwd"), true},
		{"outside base dir", "/etc/passwd", true},
		{"null byte attack", filepath.Join(tmpDir, "foo\x00bar"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestPathValidatorWithMultipleDirs(t *testing.T) {
	tmpDir1 := t.TempDir()
	tmpDir2 := t.TempDir()

	file1 := filepath.Join(tmpDir1, "file1.txt")
	file2 := filepath.Join(tmpDir2, "file2.txt")
	if err := os.WriteFile(file1, []byte("1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte("2"), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewPathValidator(tmpDir1, WithAllowedDirs(tmpDir2))

	// Both directories should be allowed
	if err := v.ValidatePath(file1); err != nil {
		t.Errorf("file1 should be allowed: %v", err)
	}
	if err := v.ValidatePath(file2); err != nil {
		t.Errorf("file2 should be allowed: %v", err)
	}

	// Other directories should not be allowed
	if err := v.ValidatePath("/etc/passwd"); err == nil {
		t.Error("external file should not be allowed")
	}
}

func TestSafeReadFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := []byte("test content")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatal(err)
	}

	// Valid read
	content, err := SafeReadFile(testFile)
	if err != nil {
		t.Errorf("SafeReadFile failed: %v", err)
	}
	if string(content) != string(testContent) {
		t.Errorf("content mismatch: got %q, want %q", content, testContent)
	}

	// Traversal attack
	_, err = SafeReadFile("../../../etc/passwd")
	if err == nil {
		t.Error("SafeReadFile should reject traversal attack")
	}

	// Null byte attack
	_, err = SafeReadFile("file\x00.txt")
	if err == nil {
		t.Error("SafeReadFile should reject null byte")
	}
}

func TestSafeOpen(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Valid open
	f, err := SafeOpen(testFile)
	if err != nil {
		t.Errorf("SafeOpen failed: %v", err)
	}
	if f != nil {
		f.Close()
	}

	// Traversal attack
	_, err = SafeOpen("../../../etc/passwd")
	if err == nil {
		t.Error("SafeOpen should reject traversal attack")
	}
}

func TestSafeReadFileWithExtension(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "test.json")
	yamlFile := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(jsonFile, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(yamlFile, []byte("key: value"), 0644); err != nil {
		t.Fatal(err)
	}

	// Valid extension
	_, err := SafeReadFileWithExtension(jsonFile, ".json")
	if err != nil {
		t.Errorf("SafeReadFileWithExtension failed for .json: %v", err)
	}

	// Invalid extension
	_, err = SafeReadFileWithExtension(yamlFile, ".json")
	if err == nil {
		t.Error("SafeReadFileWithExtension should reject wrong extension")
	}
}

func TestValidateAndClean(t *testing.T) {
	tmpDir := t.TempDir()
	v := NewPathValidator(tmpDir)

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Valid path
	cleaned, err := v.ValidateAndClean(testFile)
	if err != nil {
		t.Errorf("ValidateAndClean failed: %v", err)
	}
	if cleaned == "" {
		t.Error("ValidateAndClean should return non-empty path")
	}

	// Invalid path
	_, err = v.ValidateAndClean("../../../etc/passwd")
	if err == nil {
		t.Error("ValidateAndClean should reject traversal")
	}
}
