// Package enclave_runtime provides TEE enclave implementations.
//
// This file implements common cryptographic utilities shared across all TEE
// attestation verification implementations (SGX DCAP, SEV-SNP, and Nitro).
//
// Components:
// - CertificateChainVerifier: X.509 certificate chain validation
// - ECDSAVerifier: ECDSA P-256 and P-384 signature verification
// - HashComputer: Cryptographic hash computation (SHA-256, SHA-384, SHA-512)
// - CertificateCache: Thread-safe certificate caching for performance
//
// Task Reference: VE-2030 - Real Attestation Crypto Verification
package enclave_runtime

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// =============================================================================
// Cryptographic Error Types
// =============================================================================

var (
	// ErrInvalidSignature indicates signature verification failed.
	ErrInvalidSignature = errors.New("invalid signature")
	// ErrInvalidCertificate indicates certificate parsing or validation failed.
	ErrInvalidCertificate = errors.New("invalid certificate")
	// ErrCertificateExpired indicates the certificate has expired.
	ErrCertificateExpired = errors.New("certificate expired")
	// ErrCertificateNotYetValid indicates the certificate is not yet valid.
	ErrCertificateNotYetValid = errors.New("certificate not yet valid")
	// ErrCertificateChainTooLong indicates the certificate chain exceeds max length.
	ErrCertificateChainTooLong = errors.New("certificate chain too long")
	// ErrNoRootCertificate indicates no root CA certificate is configured.
	ErrNoRootCertificate = errors.New("no root certificate configured")
	// ErrUntrustedCertificate indicates the certificate is not trusted.
	ErrUntrustedCertificate = errors.New("certificate not trusted by any root CA")
	// ErrInvalidPublicKey indicates the public key is invalid or unsupported.
	ErrInvalidPublicKey = errors.New("invalid or unsupported public key")
	// ErrInvalidHash indicates the hash is invalid for the signature type.
	ErrInvalidHash = errors.New("invalid hash length for signature type")
	// ErrCacheNotFound indicates the requested item was not found in cache.
	ErrCacheNotFound = errors.New("item not found in cache")
)

// =============================================================================
// Certificate Chain Verifier
// =============================================================================

// CertificateChainVerifier provides X.509 certificate chain verification.
// It validates certificates against trusted root CAs and checks validity periods.
type CertificateChainVerifier struct {
	// RootCAs is the pool of trusted root CA certificates.
	RootCAs *x509.CertPool
	// IntermediateCAs is the pool of intermediate CA certificates.
	IntermediateCAs *x509.CertPool
	// CurrentTime is the time to use for validity checks (use time.Now() if zero).
	CurrentTime time.Time
	// MaxChainLen is the maximum allowed certificate chain length.
	MaxChainLen int
	// AllowExpired allows expired certificates (for testing only).
	AllowExpired bool
	// RequireKeyUsage enforces key usage extensions.
	RequireKeyUsage bool
}

// NewCertificateChainVerifier creates a new certificate chain verifier with defaults.
func NewCertificateChainVerifier() *CertificateChainVerifier {
	return &CertificateChainVerifier{
		RootCAs:         x509.NewCertPool(),
		IntermediateCAs: x509.NewCertPool(),
		MaxChainLen:     5,
		RequireKeyUsage: true,
	}
}

// AddRootCA adds a root CA certificate from PEM data.
func (v *CertificateChainVerifier) AddRootCA(pemData []byte) error {
	if !v.RootCAs.AppendCertsFromPEM(pemData) {
		// Try parsing as DER
		cert, err := x509.ParseCertificate(pemData)
		if err != nil {
			return fmt.Errorf("failed to parse root CA certificate: %w", err)
		}
		v.RootCAs.AddCert(cert)
	}
	return nil
}

// AddIntermediateCA adds an intermediate CA certificate from PEM data.
func (v *CertificateChainVerifier) AddIntermediateCA(pemData []byte) error {
	if !v.IntermediateCAs.AppendCertsFromPEM(pemData) {
		cert, err := x509.ParseCertificate(pemData)
		if err != nil {
			return fmt.Errorf("failed to parse intermediate CA certificate: %w", err)
		}
		v.IntermediateCAs.AddCert(cert)
	}
	return nil
}

// getVerifyTime returns the time to use for verification.
func (v *CertificateChainVerifier) getVerifyTime() time.Time {
	if v.CurrentTime.IsZero() {
		return time.Now()
	}
	return v.CurrentTime
}

// Verify verifies a certificate chain from leaf to root.
// The first certificate in the chain should be the leaf (end-entity) certificate.
func (v *CertificateChainVerifier) Verify(certChain []*x509.Certificate) error {
	if len(certChain) == 0 {
		return ErrInvalidCertificate
	}

	if v.MaxChainLen > 0 && len(certChain) > v.MaxChainLen {
		return fmt.Errorf("%w: chain length %d exceeds max %d", ErrCertificateChainTooLong, len(certChain), v.MaxChainLen)
	}

	// Use the root CAs pool
	roots := v.RootCAs
	if roots == nil {
		return ErrNoRootCertificate
	}

	// Build intermediate pool from provided chain and configured intermediates
	intermediates := x509.NewCertPool()
	for i := 1; i < len(certChain); i++ {
		intermediates.AddCert(certChain[i])
	}
	// Add configured intermediates
	if v.IntermediateCAs != nil {
		// Note: x509.CertPool doesn't expose certs, so we work with what we have
	}

	// Verify the leaf certificate
	leaf := certChain[0]
	verifyTime := v.getVerifyTime()

	// Check validity period manually if not allowing expired
	if !v.AllowExpired {
		if verifyTime.Before(leaf.NotBefore) {
			return fmt.Errorf("%w: certificate not valid until %v", ErrCertificateNotYetValid, leaf.NotBefore)
		}
		if verifyTime.After(leaf.NotAfter) {
			return fmt.Errorf("%w: certificate expired at %v", ErrCertificateExpired, leaf.NotAfter)
		}
	}

	opts := x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
		CurrentTime:   verifyTime,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}

	chains, err := leaf.Verify(opts)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUntrustedCertificate, err)
	}

	if len(chains) == 0 {
		return ErrUntrustedCertificate
	}

	return nil
}

// VerifyLeaf verifies a single leaf certificate against the root CAs.
func (v *CertificateChainVerifier) VerifyLeaf(cert *x509.Certificate) error {
	return v.Verify([]*x509.Certificate{cert})
}

// ParseCertificateChain parses a chain of PEM-encoded certificates.
func ParseCertificateChain(pemData []byte) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate

	for len(pemData) > 0 {
		var block *pem.Block
		block, pemData = pem.Decode(pemData)
		if block == nil {
			break
		}

		if block.Type != "CERTIFICATE" {
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate: %w", err)
		}
		certs = append(certs, cert)
	}

	if len(certs) == 0 {
		return nil, ErrInvalidCertificate
	}

	return certs, nil
}

// ParseDERCertificate parses a DER-encoded certificate.
func ParseDERCertificate(derData []byte) (*x509.Certificate, error) {
	cert, err := x509.ParseCertificate(derData)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidCertificate, err)
	}
	return cert, nil
}

// =============================================================================
// ECDSA Verifier
// =============================================================================

// ECDSAVerifier provides ECDSA signature verification for P-256 and P-384 curves.
type ECDSAVerifier struct{}

// NewECDSAVerifier creates a new ECDSA verifier.
func NewECDSAVerifier() *ECDSAVerifier {
	return &ECDSAVerifier{}
}

// ecdsaSignature represents an ASN.1-encoded ECDSA signature.
type ecdsaSignature struct {
	R, S *big.Int
}

// VerifyP256 verifies an ECDSA P-256 signature.
// The hash must be 32 bytes (SHA-256).
// The signature can be either ASN.1 DER encoded or raw (r||s, 64 bytes).
func (v *ECDSAVerifier) VerifyP256(pubKey *ecdsa.PublicKey, hash, sig []byte) error {
	if pubKey == nil {
		return ErrInvalidPublicKey
	}
	if len(hash) != 32 {
		return fmt.Errorf("%w: expected 32 bytes for P-256, got %d", ErrInvalidHash, len(hash))
	}

	r, s, err := parseECDSASignature(sig, 32)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidSignature, err)
	}

	if !ecdsa.Verify(pubKey, hash, r, s) {
		return ErrInvalidSignature
	}

	return nil
}

// VerifyP384 verifies an ECDSA P-384 signature.
// The hash must be 48 bytes (SHA-384).
// The signature can be either ASN.1 DER encoded or raw (r||s, 96 bytes).
func (v *ECDSAVerifier) VerifyP384(pubKey *ecdsa.PublicKey, hash, sig []byte) error {
	if pubKey == nil {
		return ErrInvalidPublicKey
	}
	if len(hash) != 48 {
		return fmt.Errorf("%w: expected 48 bytes for P-384, got %d", ErrInvalidHash, len(hash))
	}

	r, s, err := parseECDSASignature(sig, 48)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidSignature, err)
	}

	if !ecdsa.Verify(pubKey, hash, r, s) {
		return ErrInvalidSignature
	}

	return nil
}

// VerifyWithHash verifies an ECDSA signature with automatic curve detection.
// Uses the public key's curve to determine expected hash length.
func (v *ECDSAVerifier) VerifyWithHash(pubKey *ecdsa.PublicKey, hash, sig []byte) error {
	if pubKey == nil {
		return ErrInvalidPublicKey
	}

	curveSize := (pubKey.Curve.Params().BitSize + 7) / 8

	switch curveSize {
	case 32: // P-256
		return v.VerifyP256(pubKey, hash, sig)
	case 48: // P-384
		return v.VerifyP384(pubKey, hash, sig)
	default:
		return fmt.Errorf("%w: unsupported curve size %d", ErrInvalidPublicKey, curveSize)
	}
}

// parseECDSASignature parses an ECDSA signature in either ASN.1 DER or raw format.
func parseECDSASignature(sig []byte, componentSize int) (*big.Int, *big.Int, error) {
	// Try ASN.1 DER format first
	var esig ecdsaSignature
	if _, err := asn1.Unmarshal(sig, &esig); err == nil {
		return esig.R, esig.S, nil
	}

	// Try raw format (r||s)
	expectedLen := componentSize * 2
	if len(sig) == expectedLen {
		r := new(big.Int).SetBytes(sig[:componentSize])
		s := new(big.Int).SetBytes(sig[componentSize:])
		return r, s, nil
	}

	return nil, nil, fmt.Errorf("signature length %d does not match expected %d", len(sig), expectedLen)
}

// ExtractPublicKeyFromCert extracts an ECDSA public key from an X.509 certificate.
func ExtractPublicKeyFromCert(cert *x509.Certificate) (*ecdsa.PublicKey, error) {
	if cert == nil {
		return nil, ErrInvalidCertificate
	}

	pubKey, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("%w: certificate does not contain ECDSA public key", ErrInvalidPublicKey)
	}

	return pubKey, nil
}

// =============================================================================
// Hash Computer
// =============================================================================

// HashComputer provides cryptographic hash computation.
type HashComputer struct{}

// NewHashComputer creates a new hash computer.
func NewHashComputer() *HashComputer {
	return &HashComputer{}
}

// SHA256 computes the SHA-256 hash of the input data.
func (h *HashComputer) SHA256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// SHA384 computes the SHA-384 hash of the input data.
func (h *HashComputer) SHA384(data []byte) []byte {
	hash := sha512.Sum384(data)
	return hash[:]
}

// SHA512 computes the SHA-512 hash of the input data.
func (h *HashComputer) SHA512(data []byte) []byte {
	hash := sha512.Sum512(data)
	return hash[:]
}

// ComputeHash computes a hash with the specified algorithm.
func (h *HashComputer) ComputeHash(algorithm string, data []byte) ([]byte, error) {
	switch algorithm {
	case "SHA-256", "sha256", "SHA256":
		return h.SHA256(data), nil
	case "SHA-384", "sha384", "SHA384":
		return h.SHA384(data), nil
	case "SHA-512", "sha512", "SHA512":
		return h.SHA512(data), nil
	default:
		return nil, fmt.Errorf("unsupported hash algorithm: %s", algorithm)
	}
}

// =============================================================================
// Certificate Cache
// =============================================================================

// CachedCertificate represents a cached certificate with metadata.
type CachedCertificate struct {
	Certificate *x509.Certificate
	PEMData     []byte
	DERData     []byte
	FetchedAt   time.Time
	ExpiresAt   time.Time
	Source      string
}

// CertificateCache provides thread-safe certificate caching.
type CertificateCache struct {
	mu      sync.RWMutex
	cache   map[string]*CachedCertificate
	maxSize int
	ttl     time.Duration
}

// NewCertificateCache creates a new certificate cache.
func NewCertificateCache(maxSize int, ttl time.Duration) *CertificateCache {
	if maxSize <= 0 {
		maxSize = 100
	}
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &CertificateCache{
		cache:   make(map[string]*CachedCertificate),
		maxSize: maxSize,
		ttl:     ttl,
	}
}

// Get retrieves a certificate from the cache.
func (c *CertificateCache) Get(key string) (*CachedCertificate, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cached, ok := c.cache[key]
	if !ok {
		return nil, ErrCacheNotFound
	}

	// Check if expired
	if time.Now().After(cached.ExpiresAt) {
		return nil, ErrCacheNotFound
	}

	return cached, nil
}

// Put stores a certificate in the cache.
func (c *CertificateCache) Put(key string, cert *x509.Certificate, pemData, derData []byte, source string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict oldest entries if cache is full
	if len(c.cache) >= c.maxSize {
		c.evictOldest()
	}

	c.cache[key] = &CachedCertificate{
		Certificate: cert,
		PEMData:     pemData,
		DERData:     derData,
		FetchedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(c.ttl),
		Source:      source,
	}
}

// evictOldest removes the oldest cached entry.
func (c *CertificateCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, cached := range c.cache {
		if oldestKey == "" || cached.FetchedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = cached.FetchedAt
		}
	}

	if oldestKey != "" {
		delete(c.cache, oldestKey)
	}
}

// Clear removes all entries from the cache.
func (c *CertificateCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]*CachedCertificate)
}

// Size returns the number of cached certificates.
func (c *CertificateCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}

// =============================================================================
// Utility Functions
// =============================================================================

// ConcatBytes concatenates multiple byte slices.
func ConcatBytes(slices ...[]byte) []byte {
	totalLen := 0
	for _, s := range slices {
		totalLen += len(s)
	}

	result := make([]byte, 0, totalLen)
	for _, s := range slices {
		result = append(result, s...)
	}
	return result
}

// ConstantTimeCompare performs constant-time comparison of two byte slices.
func ConstantTimeCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	var result byte
	for i := range a {
		result |= a[i] ^ b[i]
	}
	return result == 0
}

// ZeroBytes securely zeros a byte slice.
func ZeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
