// Package security provides cryptographically secure random utilities.
package security

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
)

// =============================================================================
// Secure Random Number Generation
// =============================================================================

// SecureRandomBytes generates cryptographically secure random bytes.
// This uses crypto/rand which is suitable for security-sensitive applications.
//
// Parameters:
//   - n: Number of bytes to generate
//
// Returns the random bytes or an error if random generation fails.
func SecureRandomBytes(n int) ([]byte, error) {
	if n < 0 {
		return nil, fmt.Errorf("invalid byte count: %d", n)
	}
	if n == 0 {
		return []byte{}, nil
	}

	b := make([]byte, n)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return b, nil
}

// MustSecureRandomBytes is like SecureRandomBytes but panics on error.
// Use only in initialization code where failure is unrecoverable.
func MustSecureRandomBytes(n int) []byte {
	b, err := SecureRandomBytes(n)
	if err != nil {
		panic(err)
	}
	return b
}

// SecureRandomHex generates a cryptographically secure random hex string.
//
// Parameters:
//   - nBytes: Number of random bytes (hex string will be 2x this length)
//
// Returns the hex-encoded string or an error.
func SecureRandomHex(nBytes int) (string, error) {
	b, err := SecureRandomBytes(nBytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// MustSecureRandomHex is like SecureRandomHex but panics on error.
func MustSecureRandomHex(nBytes int) string {
	s, err := SecureRandomHex(nBytes)
	if err != nil {
		panic(err)
	}
	return s
}

// SecureRandomBase64 generates a cryptographically secure random base64 string.
//
// Parameters:
//   - nBytes: Number of random bytes (base64 string will be ~4/3x this length)
//
// Returns the base64-encoded string or an error.
func SecureRandomBase64(nBytes int) (string, error) {
	b, err := SecureRandomBytes(nBytes)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// SecureRandomBase64URL generates a URL-safe base64 random string.
//
// Parameters:
//   - nBytes: Number of random bytes
//
// Returns the URL-safe base64-encoded string or an error.
func SecureRandomBase64URL(nBytes int) (string, error) {
	b, err := SecureRandomBytes(nBytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// =============================================================================
// Secure Random Integers
// =============================================================================

// SecureRandomInt generates a cryptographically secure random integer in [0, max).
//
// Parameters:
//   - max: The exclusive upper bound (must be > 0)
//
// Returns a random integer in [0, max) or an error.
func SecureRandomInt(max int64) (int64, error) {
	if max <= 0 {
		return 0, fmt.Errorf("max must be positive, got %d", max)
	}

	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return 0, fmt.Errorf("failed to generate random int: %w", err)
	}
	return n.Int64(), nil
}

// SecureRandomIntRange generates a cryptographically secure random integer in [min, max].
//
// Parameters:
//   - min: The inclusive lower bound
//   - max: The inclusive upper bound
//
// Returns a random integer in [min, max] or an error.
func SecureRandomIntRange(min, max int64) (int64, error) {
	if min > max {
		return 0, fmt.Errorf("min (%d) must be <= max (%d)", min, max)
	}
	if min == max {
		return min, nil
	}

	rangeSize := max - min + 1
	offset, err := SecureRandomInt(rangeSize)
	if err != nil {
		return 0, err
	}
	return min + offset, nil
}

// SecureRandomUint64 generates a cryptographically secure random uint64.
func SecureRandomUint64() (uint64, error) {
	b, err := SecureRandomBytes(8)
	if err != nil {
		return 0, err
	}
	// Big-endian conversion
	return uint64(b[0])<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 | uint64(b[3])<<32 |
		uint64(b[4])<<24 | uint64(b[5])<<16 | uint64(b[6])<<8 | uint64(b[7]), nil
}

// SecureRandomUint32 generates a cryptographically secure random uint32.
func SecureRandomUint32() (uint32, error) {
	b, err := SecureRandomBytes(4)
	if err != nil {
		return 0, err
	}
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3]), nil
}

// =============================================================================
// Secure Random Selection
// =============================================================================

// SecureRandomChoice selects a random element from a slice.
//
// Parameters:
//   - slice: The slice to select from
//
// Returns a random element or an error if the slice is empty.
func SecureRandomChoice[T any](slice []T) (T, error) {
	var zero T
	if len(slice) == 0 {
		return zero, fmt.Errorf("cannot select from empty slice")
	}

	idx, err := SecureRandomInt(int64(len(slice)))
	if err != nil {
		return zero, err
	}
	return slice[idx], nil
}

// SecureRandomShuffle randomly shuffles a slice in-place using crypto/rand.
// This is suitable for security-sensitive shuffling.
//
// Parameters:
//   - slice: The slice to shuffle
//
// Returns an error if random generation fails.
func SecureRandomShuffle[T any](slice []T) error {
	n := len(slice)
	for i := n - 1; i > 0; i-- {
		j, err := SecureRandomInt(int64(i + 1))
		if err != nil {
			return err
		}
		slice[i], slice[j] = slice[j], slice[i]
	}
	return nil
}

// =============================================================================
// Token and Identifier Generation
// =============================================================================

// SecureRandomToken generates a cryptographically secure token suitable for
// API keys, session tokens, CSRF tokens, etc.
//
// Parameters:
//   - nBytes: Number of random bytes (default 32 = 256 bits if 0)
//
// Returns a hex-encoded token string.
func SecureRandomToken(nBytes int) (string, error) {
	if nBytes <= 0 {
		nBytes = 32 // Default to 256 bits
	}
	return SecureRandomHex(nBytes)
}

// SecureRandomID generates a cryptographically secure identifier.
// The identifier is URL-safe and suitable for database IDs.
//
// Parameters:
//   - nBytes: Number of random bytes (default 16 = 128 bits if 0)
//
// Returns a URL-safe base64 encoded ID without padding.
func SecureRandomID(nBytes int) (string, error) {
	if nBytes <= 0 {
		nBytes = 16 // Default to 128 bits
	}

	b, err := SecureRandomBytes(nBytes)
	if err != nil {
		return "", err
	}

	// Use RawURLEncoding to avoid padding
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// =============================================================================
// Jitter and Backoff Utilities
// =============================================================================

// SecureJitterDuration adds cryptographically secure jitter to a duration.
// This is useful for backoff algorithms to prevent thundering herd.
//
// Parameters:
//   - base: The base duration
//   - jitterFraction: The maximum jitter as a fraction of base (0.0 to 1.0)
//
// Returns the duration with added jitter.
func SecureJitterDuration(base int64, jitterFraction float64) (int64, error) {
	if jitterFraction < 0 || jitterFraction > 1 {
		jitterFraction = 0.1 // Default to 10%
	}

	maxJitter := int64(float64(base) * jitterFraction)
	if maxJitter <= 0 {
		return base, nil
	}

	jitter, err := SecureRandomInt(maxJitter * 2) // -jitter to +jitter
	if err != nil {
		return base, err // Return base on error, don't fail
	}

	return base + jitter - maxJitter, nil
}

// SecureExponentialBackoff calculates exponential backoff with secure jitter.
//
// Parameters:
//   - attempt: The current attempt number (0-indexed)
//   - baseDelay: The base delay in milliseconds
//   - maxDelay: The maximum delay in milliseconds
//   - jitterFraction: Jitter fraction (0.0 to 1.0)
//
// Returns the backoff delay in milliseconds.
func SecureExponentialBackoff(attempt int, baseDelay, maxDelay int64, jitterFraction float64) (int64, error) {
	if attempt < 0 {
		attempt = 0
	}

	// Limit attempt to prevent overflow (2^63 is max for int64)
	// With base delay, capping at 62 ensures no overflow
	const maxAttempt = 62
	if attempt > maxAttempt {
		attempt = maxAttempt
	}

	// Calculate exponential delay: base * 2^attempt
	// Safe because attempt is bounded and we check overflow below
	//nolint:gosec // G115: attempt is bounded to maxAttempt (62) above
	delay := baseDelay << uint(attempt)
	if delay > maxDelay || delay < 0 { // Overflow check
		delay = maxDelay
	}

	return SecureJitterDuration(delay, jitterFraction)
}
