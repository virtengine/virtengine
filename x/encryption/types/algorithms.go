package types

import (
	"fmt"
)

// Algorithm identifiers for supported encryption algorithms
const (
	// AlgorithmX25519XSalsa20Poly1305 is the primary algorithm using NaCl box
	// X25519 for key exchange, XSalsa20 for encryption, Poly1305 for authentication
	AlgorithmX25519XSalsa20Poly1305 = "X25519-XSALSA20-POLY1305"

	// AlgorithmAgeX25519 is an alternative using the age encryption format
	// Reserved for future implementation
	AlgorithmAgeX25519 = "AGE-X25519"

	// Convenience aliases for test code compatibility
	// AlgorithmIDNaClBox is an alias for the NaCl box algorithm
	AlgorithmIDNaClBox = AlgorithmX25519XSalsa20Poly1305

	// AlgorithmIDMLKEM768 is a placeholder for future post-quantum algorithm
	AlgorithmIDMLKEM768 = "ML-KEM-768"

	// AlgorithmIDHybridNaClMLKEM is a placeholder for future hybrid algorithm
	AlgorithmIDHybridNaClMLKEM = "HYBRID-NACL-MLKEM"
)

// NIST security levels for algorithm categorization
const (
	// NISTSecurityLevel128 corresponds to 128-bit classical security
	NISTSecurityLevel128 = 128

	// NISTSecurityLevel192 corresponds to 192-bit classical security
	NISTSecurityLevel192 = 192

	// NISTSecurityLevel256 corresponds to 256-bit classical security
	NISTSecurityLevel256 = 256
)

// AlgorithmSpecVersion is the current version of the algorithm spec format
const AlgorithmSpecVersion uint32 = 1

// Algorithm version for future upgrades
const (
	// AlgorithmVersionV1 is the initial algorithm version
	AlgorithmVersionV1 uint32 = 1
)

// Key sizes and constants for X25519-XSalsa20-Poly1305
const (
	// X25519PublicKeySize is the size of an X25519 public key in bytes
	X25519PublicKeySize = 32

	// X25519PrivateKeySize is the size of an X25519 private key in bytes
	X25519PrivateKeySize = 32

	// XSalsa20NonceSize is the size of the nonce for XSalsa20 in bytes
	XSalsa20NonceSize = 24

	// Poly1305TagSize is the size of the Poly1305 authentication tag
	Poly1305TagSize = 16

	// KeyFingerprintSize is the size of a key fingerprint (SHA256 truncated)
	KeyFingerprintSize = 20
)

// SupportedAlgorithms returns the list of supported algorithm IDs
func SupportedAlgorithms() []string {
	return []string{
		AlgorithmX25519XSalsa20Poly1305,
		// AlgorithmAgeX25519, // Reserved for future
	}
}

// IsAlgorithmSupported checks if an algorithm ID is supported
func IsAlgorithmSupported(algorithmID string) bool {
	for _, alg := range SupportedAlgorithms() {
		if alg == algorithmID {
			return true
		}
	}
	return false
}

// DefaultAlgorithm returns the default (primary) algorithm
func DefaultAlgorithm() string {
	return AlgorithmX25519XSalsa20Poly1305
}

// AlgorithmInfo contains metadata about an encryption algorithm
type AlgorithmInfo struct {
	// ID is the algorithm identifier
	ID string `json:"id"`

	// Version is the algorithm version
	Version uint32 `json:"version"`

	// Description is a human-readable description
	Description string `json:"description"`

	// KeySize is the public key size in bytes
	KeySize int `json:"key_size"`

	// NonceSize is the nonce/IV size in bytes
	NonceSize int `json:"nonce_size"`

	// Deprecated indicates if this algorithm should no longer be used for new encryptions
	Deprecated bool `json:"deprecated"`
}

// GetAlgorithmInfo returns information about an algorithm
func GetAlgorithmInfo(algorithmID string) (AlgorithmInfo, error) {
	switch algorithmID {
	case AlgorithmX25519XSalsa20Poly1305:
		return AlgorithmInfo{
			ID:          AlgorithmX25519XSalsa20Poly1305,
			Version:     AlgorithmVersionV1,
			Description: "X25519 key exchange with XSalsa20-Poly1305 authenticated encryption (NaCl box)",
			KeySize:     X25519PublicKeySize,
			NonceSize:   XSalsa20NonceSize,
			Deprecated:  false,
		}, nil
	case AlgorithmAgeX25519:
		return AlgorithmInfo{
			ID:          AlgorithmAgeX25519,
			Version:     AlgorithmVersionV1,
			Description: "age encryption format with X25519 (reserved for future)",
			KeySize:     X25519PublicKeySize,
			NonceSize:   16, // age uses different nonce handling
			Deprecated:  false,
		}, nil
	default:
		return AlgorithmInfo{}, fmt.Errorf("unknown algorithm: %s", algorithmID)
	}
}

// ValidateAlgorithmParams validates parameters for a specific algorithm
func ValidateAlgorithmParams(algorithmID string, publicKey []byte, nonce []byte) error {
	info, err := GetAlgorithmInfo(algorithmID)
	if err != nil {
		return err
	}

	if len(publicKey) != info.KeySize {
		return fmt.Errorf("invalid public key size for %s: expected %d, got %d",
			algorithmID, info.KeySize, len(publicKey))
	}

	if len(nonce) != info.NonceSize {
		return fmt.Errorf("invalid nonce size for %s: expected %d, got %d",
			algorithmID, info.NonceSize, len(nonce))
	}

	return nil
}
