package security

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	// ErrInvalidCommand is returned when a command is not in the allowlist
	ErrInvalidCommand = errors.New("command not in allowlist")
	// ErrInvalidArgument is returned when a command argument contains unsafe characters
	ErrInvalidArgument = errors.New("command argument contains unsafe characters")
	// ErrCommandTimeout is returned when a command exceeds the timeout
	ErrCommandTimeout = errors.New("command execution timeout")
)

// CommandValidator provides safe command execution with validation
type CommandValidator struct {
	// AllowedCommands is the set of allowed executable names
	AllowedCommands map[string]bool
	// DefaultTimeout is the default timeout for commands
	DefaultTimeout time.Duration
}

// NewCommandValidator creates a new command validator with the given allowed commands
func NewCommandValidator(allowedCommands []string, defaultTimeout time.Duration) *CommandValidator {
	allowed := make(map[string]bool)
	for _, cmd := range allowedCommands {
		// Store only the base name for comparison
		allowed[filepath.Base(cmd)] = true
	}
	return &CommandValidator{
		AllowedCommands: allowed,
		DefaultTimeout:  defaultTimeout,
	}
}

// ValidateCommand checks if a command is allowed and its arguments are safe
func (cv *CommandValidator) ValidateCommand(name string, args ...string) error {
	// Check if command is in allowlist
	baseName := filepath.Base(name)
	if !cv.AllowedCommands[baseName] {
		return fmt.Errorf("%w: %s", ErrInvalidCommand, name)
	}

	// Validate each argument
	for i, arg := range args {
		if err := cv.ValidateArgument(arg); err != nil {
			return fmt.Errorf("argument %d: %w", i, err)
		}
	}

	return nil
}

// ValidateArgument checks if a command argument is safe
func (cv *CommandValidator) ValidateArgument(arg string) error {
	// Reject arguments containing shell metacharacters
	unsafeChars := []string{";", "|", "&", "$", "`", "\n", "\r", "<", ">", "(", ")", "{", "}"}
	for _, char := range unsafeChars {
		if strings.Contains(arg, char) {
			return fmt.Errorf("%w: contains '%s'", ErrInvalidArgument, char)
		}
	}
	return nil
}

// SafeCommand creates an exec.Cmd with validation using the provided context.
// If ctx is nil, context.Background() is used and no timeout is applied.
func (cv *CommandValidator) SafeCommand(ctx context.Context, name string, args ...string) (*exec.Cmd, error) {
	// Validate command and arguments
	if err := cv.ValidateCommand(name, args...); err != nil {
		return nil, err
	}

	// Create command with context
	if ctx == nil {
		ctx = context.Background()
	}

	//nolint:gosec // G204: Command and args validated above
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd, nil
}

// Run executes a validated command and returns the output
func (cv *CommandValidator) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), cv.DefaultTimeout)
		defer cancel()
	}

	cmd, err := cv.SafeCommand(ctx, name, args...)
	if err != nil {
		return nil, err
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("command failed: %w", err)
	}

	return output, nil
}
