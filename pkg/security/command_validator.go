package security

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
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
	// AllowedPaths maps a command base name to allowed absolute paths
	AllowedPaths map[string]map[string]struct{}
	// DefaultTimeout is the default timeout for commands
	DefaultTimeout time.Duration
}

// NewCommandValidator creates a new command validator with the given allowed commands
func NewCommandValidator(allowedCommands []string, defaultTimeout time.Duration) *CommandValidator {
	allowed := make(map[string]bool)
	allowedPaths := make(map[string]map[string]struct{})
	for _, cmd := range allowedCommands {
		// Store only the base name for comparison
		clean := filepath.Clean(cmd)
		base := filepath.Base(clean)
		baseNames := []string{base}

		if runtime.GOOS == osWindows {
			ext := strings.ToLower(filepath.Ext(base))
			if ext != "" {
				baseNames = append(baseNames, strings.TrimSuffix(base, ext))
			}
		}

		for _, baseName := range baseNames {
			allowed[baseName] = true
		}

		if isPathLike(cmd) {
			absPath := clean
			if !filepath.IsAbs(absPath) {
				if resolved, err := filepath.Abs(clean); err == nil {
					absPath = resolved
				}
			}
			normalized := normalizeCommandPath(absPath)
			for _, baseName := range baseNames {
				if allowedPaths[baseName] == nil {
					allowedPaths[baseName] = make(map[string]struct{})
				}
				allowedPaths[baseName][normalized] = struct{}{}
			}
		}
	}
	return &CommandValidator{
		AllowedCommands: allowed,
		AllowedPaths:    allowedPaths,
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

	if err := cv.ensureAllowedPath(baseName, name); err != nil {
		return err
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

// SafeCommand creates an exec.Cmd with validation and timeout
func (cv *CommandValidator) SafeCommand(ctx context.Context, name string, args ...string) (*exec.Cmd, error) {
	// Validate command and arguments
	if err := cv.ValidateCommand(name, args...); err != nil {
		return nil, err
	}

	// Create command with context (caller owns cancellation)
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
		if cv.DefaultTimeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(context.Background(), cv.DefaultTimeout)
			defer cancel()
		} else {
			ctx = context.Background()
		}
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

func (cv *CommandValidator) ensureAllowedPath(baseName, name string) error {
	allowedPaths := cv.AllowedPaths[baseName]
	if len(allowedPaths) == 0 {
		return nil
	}

	resolved := name
	if !isPathLike(name) {
		path, err := exec.LookPath(name)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrInvalidCommand, name)
		}
		resolved = path
	}

	absPath := resolved
	if !filepath.IsAbs(absPath) {
		if resolvedAbs, err := filepath.Abs(resolved); err == nil {
			absPath = resolvedAbs
		}
	}

	if _, ok := allowedPaths[normalizeCommandPath(absPath)]; !ok {
		return fmt.Errorf("%w: %s", ErrInvalidCommand, name)
	}

	return nil
}

func isPathLike(value string) bool {
	if filepath.IsAbs(value) {
		return true
	}

	return strings.ContainsAny(value, `/\`)
}

func normalizeCommandPath(path string) string {
	clean := filepath.Clean(path)
	if runtime.GOOS == osWindows {
		return strings.ToLower(clean)
	}
	return clean
}
