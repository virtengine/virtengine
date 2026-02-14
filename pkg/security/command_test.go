package security

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestCommandValidator_ValidateCommand(t *testing.T) {
	cv := NewCommandValidator([]string{"ls", "grep", "kubectl"}, 30*time.Second)

	tests := []struct {
		name    string
		cmd     string
		args    []string
		wantErr bool
	}{
		{
			name:    "allowed command",
			cmd:     "ls",
			args:    []string{"-la", "/tmp"},
			wantErr: false,
		},
		{
			name:    "disallowed command",
			cmd:     "rm",
			args:    []string{"-rf", "/"},
			wantErr: true,
		},
		{
			name:    "command with shell metacharacter",
			cmd:     "ls",
			args:    []string{"-la; rm -rf /"},
			wantErr: true,
		},
		{
			name:    "command with pipe",
			cmd:     "grep",
			args:    []string{"pattern | cat"},
			wantErr: true,
		},
		{
			name:    "command with command substitution",
			cmd:     "kubectl",
			args:    []string{"get", "pods", "$(malicious)"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cv.ValidateCommand(tt.cmd, tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommandValidator_ValidateArgument(t *testing.T) {
	cv := NewCommandValidator([]string{}, 30*time.Second)

	tests := []struct {
		name    string
		arg     string
		wantErr bool
	}{
		{"safe argument", "/path/to/file", false},
		{"safe argument with dash", "--flag=value", false},
		{"semicolon", "arg;malicious", true},
		{"pipe", "arg|malicious", true},
		{"ampersand", "arg&malicious", true},
		{"dollar sign", "arg$malicious", true},
		{"backtick", "arg`malicious`", true},
		{"newline", "arg\nmalicious", true},
		{"redirect", "arg>file", true},
		{"subshell", "arg$(cmd)", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cv.ValidateArgument(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateArgument() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommandValidator_SafeCommand(t *testing.T) {
	cv := NewCommandValidator([]string{"echo"}, 30*time.Second)

	t.Run("creates command for allowed", func(t *testing.T) {
		cmd, err := cv.SafeCommand(context.Background(), "echo", "hello")
		if err != nil {
			t.Errorf("SafeCommand() error = %v", err)
		}
		if cmd == nil {
			t.Error("SafeCommand() returned nil cmd")
		}
	})

	t.Run("rejects disallowed command", func(t *testing.T) {
		_, err := cv.SafeCommand(context.Background(), "rm", "-rf", "/")
		if err == nil {
			t.Error("SafeCommand() should reject disallowed command")
		}
	})

	t.Run("rejects unsafe arguments", func(t *testing.T) {
		_, err := cv.SafeCommand(context.Background(), "echo", "test; rm -rf /")
		if err == nil {
			t.Error("SafeCommand() should reject unsafe arguments")
		}
	})
}

func TestCommandValidator_PathAllowlist(t *testing.T) {
	dir := t.TempDir()
	cmdName := "safe-cmd"
	fileName := cmdName
	if runtime.GOOS == "windows" {
		fileName = cmdName + ".exe"
	}

	allowedPath := filepath.Join(dir, fileName)
	if err := os.WriteFile(allowedPath, []byte(""), 0755); err != nil { //nolint:gosec // G306: test file requires exec bit for LookPath
		t.Fatalf("failed to create allowed command: %v", err)
	}

	oldPath := os.Getenv("PATH")
	t.Cleanup(func() {
		_ = os.Setenv("PATH", oldPath)
	})
	if err := os.Setenv("PATH", dir+string(os.PathListSeparator)+oldPath); err != nil {
		t.Fatalf("failed to set PATH: %v", err)
	}

	cv := NewCommandValidator([]string{allowedPath}, 30*time.Second)

	if err := cv.ValidateCommand(cmdName); err != nil {
		t.Fatalf("ValidateCommand() with allowlisted path failed: %v", err)
	}

	otherPath := filepath.Join(t.TempDir(), fileName)
	if err := os.WriteFile(otherPath, []byte(""), 0755); err != nil { //nolint:gosec // G306: test file requires exec bit for LookPath
		t.Fatalf("failed to create other command: %v", err)
	}

	if err := cv.ValidateCommand(otherPath); err == nil {
		t.Fatalf("ValidateCommand() should reject non-allowlisted path")
	}
}
