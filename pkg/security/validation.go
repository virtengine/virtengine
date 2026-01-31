// Package security provides input validation and command injection prevention utilities.
//
// This package implements comprehensive security controls for:
// - CWE-78: OS Command Injection prevention via executable allowlists
// - CWE-190: Integer Overflow prevention via safe conversion functions
// - Input sanitization for shell arguments, hostnames, and paths
//
// All external command execution should use these validation functions before
// invoking exec.Command to prevent command injection attacks.
//
// Task Reference: VE-7A - Command injection prevention and input sanitization
package security

import (
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// =============================================================================
// Error Definitions
// =============================================================================

var (
	// ErrInvalidExecutable is returned when an executable path is not in the allowlist
	ErrInvalidExecutable = errors.New("executable not in trusted allowlist")

	// ErrInvalidHostname is returned when a hostname format is invalid
	ErrInvalidHostname = errors.New("invalid hostname format")

	// ErrInvalidIPAddress is returned when an IP address is invalid
	ErrInvalidIPAddress = errors.New("invalid IP address")

	// ErrInvalidPingTarget is returned when a ping target is neither valid hostname nor IP
	ErrInvalidPingTarget = errors.New("invalid ping target")

	// ErrShellMetacharacters is returned when input contains shell metacharacters
	ErrShellMetacharacters = errors.New("input contains shell metacharacters")

	// ErrEmptyInput is returned when required input is empty
	ErrEmptyInput = errors.New("input cannot be empty")

	// ErrInputTooLong is returned when input exceeds maximum length
	ErrInputTooLong = errors.New("input exceeds maximum length")
)

// =============================================================================
// Trusted Executable Allowlists
// =============================================================================

// TrustedExecutables defines allowlists of trusted executable paths by category.
// Only executables in these lists may be invoked via exec.Command.
var TrustedExecutables = map[string][]string{
	"ansible": {
		// Linux paths
		"/usr/bin/ansible-playbook",
		"/usr/local/bin/ansible-playbook",
		"/opt/ansible/bin/ansible-playbook",
		// Windows paths
		"C:\\Python39\\Scripts\\ansible-playbook.exe",
		"C:\\Python310\\Scripts\\ansible-playbook.exe",
		"C:\\Python311\\Scripts\\ansible-playbook.exe",
		"C:\\Python312\\Scripts\\ansible-playbook.exe",
	},
	"sgx": {
		"/usr/bin/gramine-sgx",
		"/usr/local/bin/gramine-sgx",
		"/opt/intel/sgxsdk/bin/sgx_sign",
		"/opt/intel/sgxsdk/bin64/sgx_sign",
		"/usr/bin/ego",
		"/usr/local/bin/ego",
	},
	"nitro": {
		"/usr/bin/nitro-cli",
		"/usr/local/bin/nitro-cli",
	},
	"slurm": {
		"/usr/bin/squeue",
		"/usr/local/bin/squeue",
		"/opt/slurm/bin/squeue",
		"/usr/bin/sinfo",
		"/usr/local/bin/sinfo",
		"/opt/slurm/bin/sinfo",
		"/usr/bin/sbatch",
		"/usr/local/bin/sbatch",
		"/opt/slurm/bin/sbatch",
		"/usr/bin/scancel",
		"/usr/local/bin/scancel",
		"/opt/slurm/bin/scancel",
		"/usr/bin/slurmd",
		"/usr/local/bin/slurmd",
		"/opt/slurm/bin/slurmd",
	},
	"system": {
		// Network utilities
		"/bin/ping",
		"/usr/bin/ping",
		"/sbin/ping",
		"/usr/sbin/ping",
		// Disk utilities
		"/bin/df",
		"/usr/bin/df",
		// Process utilities
		"/usr/bin/pgrep",
		"/bin/pgrep",
		// Container runtimes
		"/usr/bin/docker",
		"/usr/local/bin/docker",
		"/usr/bin/singularity",
		"/usr/local/bin/singularity",
		// GPU utilities
		"/usr/bin/nvidia-smi",
		"/usr/local/bin/nvidia-smi",
	},
}

// =============================================================================
// Validation Patterns
// =============================================================================

var (
	// hostnameRegex validates RFC 1123 hostnames
	// - Must start with alphanumeric
	// - Can contain alphanumeric and hyphens
	// - Must end with alphanumeric
	// - Max 63 characters per label, 253 total
	hostnameRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

	// shellMetacharacters are characters that could enable command injection
	shellMetacharacters = regexp.MustCompile(`[;|&$\x60\\\"'\n\r\t<>(){}[\]*?#!~]`)

	// pathTraversalPattern detects directory traversal attempts
	pathTraversalPattern = regexp.MustCompile(`(^|[/\\])\.\.([/\\]|$)`)

	// safePathCharacters allows only safe characters in paths
	safePathCharacters = regexp.MustCompile(`^[a-zA-Z0-9._\-/\\:]+$`)

	// nodeIDRegex validates node identifiers
	nodeIDRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,62}$`)

	// ipv4Regex validates IPv4 addresses
	ipv4Regex = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)

	// ipv6Regex validates IPv6 addresses (simplified)
	ipv6Regex = regexp.MustCompile(`^([0-9a-fA-F]{0,4}:){2,7}[0-9a-fA-F]{0,4}$`)
)

// Maximum lengths for various inputs
const (
	MaxHostnameLength = 253
	MaxPathLength     = 4096
	MaxNodeIDLength   = 63
	MaxArgumentLength = 1024
)

// =============================================================================
// Executable Validation
// =============================================================================

// ValidateExecutable checks if an executable path is in the trusted allowlist.
// It performs path normalization and validates against the allowlist for the
// specified category.
//
// Parameters:
//   - category: The category of executable (e.g., "ansible", "sgx", "slurm")
//   - path: The path to the executable
//
// Returns an error if the executable is not trusted.
func ValidateExecutable(category, path string) error {
	if path == "" {
		return fmt.Errorf("%w: empty path", ErrInvalidExecutable)
	}

	// Normalize the path
	cleanPath := filepath.Clean(path)

	// Get the allowlist for this category
	allowlist, exists := TrustedExecutables[category]
	if !exists {
		return fmt.Errorf("%w: unknown category %q", ErrInvalidExecutable, category)
	}

	// Check if the path is in the allowlist
	for _, trusted := range allowlist {
		trustedClean := filepath.Clean(trusted)

		// Case-insensitive comparison on Windows
		if runtime.GOOS == "windows" {
			if strings.EqualFold(cleanPath, trustedClean) {
				return nil
			}
		} else {
			if cleanPath == trustedClean {
				return nil
			}
		}
	}

	return fmt.Errorf("%w: %q not in category %q", ErrInvalidExecutable, path, category)
}

// ResolveAndValidateExecutable resolves an executable name to its full path
// using the system PATH, then validates it against the allowlist.
//
// This is useful when the executable name is provided without a full path
// (e.g., "ansible-playbook" instead of "/usr/bin/ansible-playbook").
func ResolveAndValidateExecutable(category, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("%w: empty name", ErrInvalidExecutable)
	}

	// If it's already an absolute path, validate directly
	if filepath.IsAbs(name) {
		if err := ValidateExecutable(category, name); err != nil {
			return "", err
		}
		return filepath.Clean(name), nil
	}

	// Get the allowlist for this category
	allowlist, exists := TrustedExecutables[category]
	if !exists {
		return "", fmt.Errorf("%w: unknown category %q", ErrInvalidExecutable, category)
	}

	// Look for the executable in the allowlist
	for _, trusted := range allowlist {
		trustedClean := filepath.Clean(trusted)
		baseName := filepath.Base(trustedClean)

		// Match by base name
		if runtime.GOOS == "windows" {
			// Windows: case-insensitive, handle .exe extension
			nameWithExt := name
			if !strings.HasSuffix(strings.ToLower(name), ".exe") {
				nameWithExt = name + ".exe"
			}
			if strings.EqualFold(baseName, nameWithExt) || strings.EqualFold(baseName, name) {
				// Verify the file exists
				if _, err := os.Stat(trustedClean); err == nil {
					return trustedClean, nil
				}
			}
		} else {
			if baseName == name {
				// Verify the file exists
				if _, err := os.Stat(trustedClean); err == nil {
					return trustedClean, nil
				}
			}
		}
	}

	return "", fmt.Errorf("%w: %q not found in category %q", ErrInvalidExecutable, name, category)
}

// =============================================================================
// Input Sanitization
// =============================================================================

// SanitizeShellArg validates that a string does not contain shell metacharacters.
// This should be used for all arguments passed to external commands.
//
// Returns an error if the argument contains potentially dangerous characters.
func SanitizeShellArg(arg string) error {
	if arg == "" {
		return nil // Empty is OK for optional args
	}

	if len(arg) > MaxArgumentLength {
		return fmt.Errorf("%w: length %d exceeds %d", ErrInputTooLong, len(arg), MaxArgumentLength)
	}

	if shellMetacharacters.MatchString(arg) {
		return fmt.Errorf("%w: %q", ErrShellMetacharacters, arg)
	}

	return nil
}

// SanitizePath validates that a path is safe and does not contain traversal.
//
// Returns the cleaned path or an error if the path is invalid.
func SanitizePath(path string) (string, error) {
	if path == "" {
		return "", ErrEmptyInput
	}

	if len(path) > MaxPathLength {
		return "", fmt.Errorf("%w: length %d exceeds %d", ErrInputTooLong, len(path), MaxPathLength)
	}

	// Check for path traversal
	if pathTraversalPattern.MatchString(path) {
		return "", fmt.Errorf("%w: %q", ErrPathTraversal, path)
	}

	// Clean the path
	cleanPath := filepath.Clean(path)

	// Verify no traversal remains after cleaning
	if pathTraversalPattern.MatchString(cleanPath) {
		return "", fmt.Errorf("%w: %q", ErrPathTraversal, path)
	}

	// Verify path contains only safe characters
	if !safePathCharacters.MatchString(cleanPath) {
		return "", fmt.Errorf("%w: contains unsafe characters", ErrInvalidPath)
	}

	return cleanPath, nil
}

// ValidatePlaybookPath validates a playbook path is within allowed directories.
//
// Parameters:
//   - path: The playbook path to validate
//   - allowedDirs: List of allowed base directories (optional, if empty all paths accepted)
//
// Returns the cleaned path or an error.
func ValidatePlaybookPath(path string, allowedDirs []string) (string, error) {
	cleanPath, err := SanitizePath(path)
	if err != nil {
		return "", err
	}

	// Check file extension
	if !strings.HasSuffix(strings.ToLower(cleanPath), ".yml") &&
		!strings.HasSuffix(strings.ToLower(cleanPath), ".yaml") {
		return "", fmt.Errorf("%w: playbook must have .yml or .yaml extension", ErrInvalidPath)
	}

	// If no allowed directories specified, accept all safe paths
	if len(allowedDirs) == 0 {
		return cleanPath, nil
	}

	// Get absolute path
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", fmt.Errorf("%w: failed to get absolute path", ErrInvalidPath)
	}

	// Check if path is under an allowed directory
	for _, dir := range allowedDirs {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			continue
		}

		if strings.HasPrefix(absPath, absDir+string(filepath.Separator)) || absPath == absDir {
			return cleanPath, nil
		}
	}

	return "", fmt.Errorf("%w: not in allowed directories", ErrInvalidPath)
}

// =============================================================================
// Hostname and IP Validation
// =============================================================================

// ValidateHostname validates that a string is a valid hostname per RFC 1123.
//
// Returns an error if the hostname is invalid.
func ValidateHostname(hostname string) error {
	if hostname == "" {
		return ErrEmptyInput
	}

	if len(hostname) > MaxHostnameLength {
		return fmt.Errorf("%w: length %d exceeds %d", ErrInputTooLong, len(hostname), MaxHostnameLength)
	}

	if !hostnameRegex.MatchString(hostname) {
		return fmt.Errorf("%w: %q", ErrInvalidHostname, hostname)
	}

	return nil
}

// ValidateIPAddress validates that a string is a valid IPv4 or IPv6 address.
//
// Returns an error if the IP address is invalid.
func ValidateIPAddress(ip string) error {
	if ip == "" {
		return ErrEmptyInput
	}

	parsed := net.ParseIP(ip)
	if parsed == nil {
		return fmt.Errorf("%w: %q", ErrInvalidIPAddress, ip)
	}

	return nil
}

// ValidatePingTarget validates that a string is a valid target for ping command.
// Accepts either a valid hostname or a valid IP address.
//
// Returns an error if the target is neither a valid hostname nor IP address.
func ValidatePingTarget(target string) error {
	if target == "" {
		return ErrEmptyInput
	}

	// Check for shell metacharacters first
	if err := SanitizeShellArg(target); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidPingTarget, err)
	}

	// Try as IP address first
	if ValidateIPAddress(target) == nil {
		return nil
	}

	// Try as hostname
	if ValidateHostname(target) == nil {
		return nil
	}

	return fmt.Errorf("%w: %q is neither a valid hostname nor IP address", ErrInvalidPingTarget, target)
}

// ValidateNodeID validates a node identifier for HPC operations.
//
// Returns an error if the node ID is invalid.
func ValidateNodeID(nodeID string) error {
	if nodeID == "" {
		return ErrEmptyInput
	}

	if len(nodeID) > MaxNodeIDLength {
		return fmt.Errorf("%w: length %d exceeds %d", ErrInputTooLong, len(nodeID), MaxNodeIDLength)
	}

	if !nodeIDRegex.MatchString(nodeID) {
		return fmt.Errorf("%w: %q contains invalid characters", ErrInvalidHostname, nodeID)
	}

	return nil
}

// =============================================================================
// Safe Integer Conversions
// =============================================================================

// SafeUint32ToInt safely converts uint32 to int, checking for overflow on 32-bit systems.
//
// On 64-bit systems, uint32 always fits in int. On 32-bit systems, values > MaxInt32
// would overflow.
//
// Returns the converted value and an error if overflow would occur.
func SafeUint32ToInt(v uint32) (int, error) {
	// Check if int is 32-bit
	if ^uint(0) == math.MaxUint32 {
		// 32-bit system: check for overflow
		if v > math.MaxInt32 {
			return 0, fmt.Errorf("%w: uint32(%d) exceeds MaxInt32", ErrIntegerOverflow, v)
		}
	}
	// Safe to convert
	return int(v), nil
}

// MustUint32ToInt is like SafeUint32ToInt but returns 0 on overflow instead of error.
// Use this when you need a value and can accept clamping behavior.
func MustUint32ToInt(v uint32) int {
	result, err := SafeUint32ToInt(v)
	if err != nil {
		return math.MaxInt32
	}
	return result
}

// SafeUint16ToInt safely converts uint16 to int.
// This is always safe as uint16 max (65535) fits in int even on 32-bit systems.
func SafeUint16ToInt(v uint16) int {
	return int(v)
}

// SafeInt64ToInt safely converts int64 to int, checking for overflow.
//
// Returns the converted value and an error if overflow would occur.
func SafeInt64ToInt(v int64) (int, error) {
	// Check architecture
	if ^uint(0) == math.MaxUint32 {
		// 32-bit system
		if v > math.MaxInt32 || v < math.MinInt32 {
			return 0, fmt.Errorf("%w: int64(%d) out of int32 range", ErrIntegerOverflow, v)
		}
	}
	return int(v), nil
}

// SafeUint64ToInt safely converts uint64 to int, checking for overflow.
//
// Returns the converted value and an error if overflow would occur.
func SafeUint64ToInt(v uint64) (int, error) {
	maxInt := uint64(^uint(0) >> 1) // MaxInt for current architecture
	if v > maxInt {
		return 0, fmt.Errorf("%w: uint64(%d) exceeds MaxInt", ErrIntegerOverflow, v)
	}
	return int(v), nil
}

// SafeIntToInt32 safely converts int to int32, checking for overflow.
//
// Returns the converted value and an error if overflow would occur.
func SafeIntToInt32(v int) (int32, error) {
	if v > math.MaxInt32 || v < math.MinInt32 {
		return 0, fmt.Errorf("%w: int(%d) out of int32 range", ErrIntegerOverflow, v)
	}
	return int32(v), nil
}

// ClampToInt32 clamps an int value to int32 range without error.
func ClampToInt32(v int) int32 {
	if v > math.MaxInt32 {
		return math.MaxInt32
	}
	if v < math.MinInt32 {
		return math.MinInt32
	}
	return int32(v)
}

// SafeUint64ToUint32 safely converts uint64 to uint32, checking for overflow.
//
// Returns the converted value and an error if overflow would occur.
func SafeUint64ToUint32(v uint64) (uint32, error) {
	if v > math.MaxUint32 {
		return 0, fmt.Errorf("%w: uint64(%d) exceeds MaxUint32", ErrIntegerOverflow, v)
	}
	return uint32(v), nil
}

// SafeFloat32ToUint32 safely converts a float32 to uint32, clamping to valid range.
// Returns the converted value clamped between 0 and 100 for score values.
func SafeFloat32ToUint32(v float32, min, max uint32) uint32 {
	if v < 0 || v != v { // Check for negative or NaN
		return min
	}
	if v > float32(max) {
		return max
	}
	if v < float32(min) {
		return min
	}
	return uint32(v)
}

// =============================================================================
// Command Argument Builders
// =============================================================================

// SLURMSqueueArgs builds validated arguments for the squeue command.
//
// Returns the argument list or an error if any argument is invalid.
func SLURMSqueueArgs(format string, user string, jobID string) ([]string, error) {
	args := []string{"-h"} // No header

	if format != "" {
		if err := SanitizeShellArg(format); err != nil {
			return nil, fmt.Errorf("invalid format: %w", err)
		}
		args = append(args, "-o", format)
	}

	if user != "" {
		if err := ValidateNodeID(user); err != nil {
			return nil, fmt.Errorf("invalid user: %w", err)
		}
		args = append(args, "-u", user)
	}

	if jobID != "" {
		if err := SanitizeShellArg(jobID); err != nil {
			return nil, fmt.Errorf("invalid job ID: %w", err)
		}
		args = append(args, "-j", jobID)
	}

	return args, nil
}

// SLURMSinfoArgs builds validated arguments for the sinfo command.
//
// Returns the argument list or an error if any argument is invalid.
func SLURMSinfoArgs(format string, nodeName string) ([]string, error) {
	args := []string{"-h"} // No header

	if format != "" {
		if err := SanitizeShellArg(format); err != nil {
			return nil, fmt.Errorf("invalid format: %w", err)
		}
		args = append(args, "-o", format)
	}

	if nodeName != "" {
		if err := ValidateHostname(nodeName); err != nil {
			return nil, fmt.Errorf("invalid node name: %w", err)
		}
		args = append(args, "-N", "-n", nodeName)
	}

	return args, nil
}

// PingArgs builds validated arguments for the ping command.
//
// Returns the argument list or an error if any argument is invalid.
func PingArgs(target string, count int) ([]string, error) {
	if err := ValidatePingTarget(target); err != nil {
		return nil, err
	}

	if count <= 0 || count > 100 {
		count = 1 // Safe default
	}

	if runtime.GOOS == "windows" {
		return []string{"-n", fmt.Sprintf("%d", count), target}, nil
	}
	return []string{"-c", fmt.Sprintf("%d", count), "-W", "1", target}, nil
}

// DfArgs builds validated arguments for the df command.
//
// Returns the argument list or an error if any argument is invalid.
func DfArgs(path string) ([]string, error) {
	cleanPath, err := SanitizePath(path)
	if err != nil {
		return nil, err
	}

	return []string{"-B1", cleanPath}, nil
}
