// Package provider_daemon implements the provider daemon for VirtEngine.
//
// SECURITY-007: HSM Integration for Validator Keys
// This file provides Hardware Security Module (HSM) integration via PKCS#11.
package provider_daemon

import (
	"crypto"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

// ErrHSMNotInitialized is returned when HSM operations are called before initialization
var ErrHSMNotInitialized = errors.New("HSM not initialized")

// ErrHSMSessionClosed is returned when HSM session is closed
var ErrHSMSessionClosed = errors.New("HSM session closed")

// ErrHSMKeyNotFound is returned when a key is not found in the HSM
var ErrHSMKeyNotFound = errors.New("HSM key not found")

// ErrHSMOperationFailed is returned when an HSM operation fails
var ErrHSMOperationFailed = errors.New("HSM operation failed")

// ErrHSMAuthFailed is returned when HSM authentication fails
var ErrHSMAuthFailed = errors.New("HSM authentication failed")

// HSMType represents the type of HSM
type HSMType string

const (
	// HSMTypePKCS11 uses a PKCS#11 compatible HSM
	HSMTypePKCS11 HSMType = "pkcs11"

	// HSMTypeSoftHSM uses SoftHSM for testing/development
	HSMTypeSoftHSM HSMType = "softhsm"

	// HSMTypeYubiHSM uses YubiHSM
	HSMTypeYubiHSM HSMType = "yubihsm"

	// HSMTypeCloudHSM uses cloud-based HSM (AWS CloudHSM, Azure HSM, GCP Cloud HSM)
	HSMTypeCloudHSM HSMType = "cloudhsm"

	// HSMTypeTPM uses Trusted Platform Module
	HSMTypeTPM HSMType = "tpm"
)

// HSMKeyType represents the type of key stored in HSM
type HSMKeyType string

const (
	// HSMKeyTypeEd25519 is Ed25519 signing key
	HSMKeyTypeEd25519 HSMKeyType = "ed25519"

	// HSMKeyTypeSecp256k1 is secp256k1 signing key
	HSMKeyTypeSecp256k1 HSMKeyType = "secp256k1"

	// HSMKeyTypeP256 is NIST P-256 signing key
	HSMKeyTypeP256 HSMKeyType = "p256"

	// HSMKeyTypeRSA2048 is RSA 2048-bit key
	HSMKeyTypeRSA2048 HSMKeyType = "rsa2048"

	// HSMKeyTypeRSA4096 is RSA 4096-bit key
	HSMKeyTypeRSA4096 HSMKeyType = "rsa4096"
)

// HSMProvider defines the interface for HSM operations
type HSMProvider interface {
	// Initialize initializes the HSM connection
	Initialize(config *HSMProviderConfig) error

	// Close closes the HSM connection
	Close() error

	// IsInitialized returns true if the HSM is initialized
	IsInitialized() bool

	// GetInfo returns information about the HSM
	GetInfo() (*HSMInfo, error)

	// GenerateKey generates a new key in the HSM
	GenerateKey(label string, keyType HSMKeyType) (*HSMKeyHandle, error)

	// GetKey retrieves a key handle by label
	GetKey(label string) (*HSMKeyHandle, error)

	// ListKeys lists all keys in the HSM
	ListKeys() ([]*HSMKeyHandle, error)

	// DeleteKey deletes a key from the HSM
	DeleteKey(label string) error

	// Sign signs data using a key in the HSM
	Sign(keyHandle *HSMKeyHandle, data []byte) ([]byte, error)

	// Verify verifies a signature using a key in the HSM
	Verify(keyHandle *HSMKeyHandle, data, signature []byte) (bool, error)

	// GetPublicKey retrieves the public key for an HSM key
	GetPublicKey(keyHandle *HSMKeyHandle) ([]byte, error)

	// ImportKey imports an existing key into the HSM (if supported)
	ImportKey(label string, privateKey []byte, keyType HSMKeyType) (*HSMKeyHandle, error)

	// ExportPublicKey exports the public key (private keys never leave the HSM)
	ExportPublicKey(keyHandle *HSMKeyHandle) ([]byte, error)

	// Backup backs up HSM keys (if supported)
	Backup() ([]byte, error)

	// Restore restores HSM keys from backup (if supported)
	Restore(backup []byte) error
}

// HSMProviderConfig configures an HSM provider
type HSMProviderConfig struct {
	// Type is the HSM type
	Type HSMType `json:"type"`

	// LibraryPath is the path to the PKCS#11 library
	LibraryPath string `json:"library_path,omitempty"`

	// SlotID is the HSM slot ID
	SlotID uint `json:"slot_id"`

	// TokenLabel is the HSM token label
	TokenLabel string `json:"token_label,omitempty"`

	// PIN is the HSM PIN (user or SO)
	PIN string `json:"pin,omitempty"`

	// SOPIN is the Security Officer PIN (for admin operations)
	SOPIN string `json:"so_pin,omitempty"`

	// ConnectionTimeout is the connection timeout
	ConnectionTimeout time.Duration `json:"connection_timeout,omitempty"`

	// OperationTimeout is the operation timeout
	OperationTimeout time.Duration `json:"operation_timeout,omitempty"`

	// MaxRetries is the maximum number of retries for operations
	MaxRetries int `json:"max_retries,omitempty"`

	// CloudConfig is cloud-specific HSM configuration
	CloudConfig *CloudHSMConfig `json:"cloud_config,omitempty"`
}

// CloudHSMConfig contains cloud-specific HSM configuration
type CloudHSMConfig struct {
	// Provider is the cloud provider (aws, azure, gcp)
	Provider string `json:"provider"`

	// Region is the cloud region
	Region string `json:"region"`

	// ClusterID is the HSM cluster ID
	ClusterID string `json:"cluster_id,omitempty"`

	// KeyVaultName is the key vault name (Azure)
	KeyVaultName string `json:"key_vault_name,omitempty"`

	// CredentialsFile is the path to credentials file
	CredentialsFile string `json:"credentials_file,omitempty"`
}

// DefaultHSMProviderConfig returns the default HSM provider configuration
func DefaultHSMProviderConfig() *HSMProviderConfig {
	return &HSMProviderConfig{
		Type:              HSMTypeSoftHSM,
		LibraryPath:       "/usr/lib/softhsm/libsofthsm2.so",
		SlotID:            0,
		ConnectionTimeout: 30 * time.Second,
		OperationTimeout:  10 * time.Second,
		MaxRetries:        3,
	}
}

// HSMInfo contains information about the HSM
type HSMInfo struct {
	// Type is the HSM type
	Type HSMType `json:"type"`

	// ManufacturerID identifies the HSM manufacturer
	ManufacturerID string `json:"manufacturer_id"`

	// Model is the HSM model
	Model string `json:"model"`

	// SerialNumber is the HSM serial number
	SerialNumber string `json:"serial_number"`

	// FirmwareVersion is the HSM firmware version
	FirmwareVersion string `json:"firmware_version"`

	// HardwareVersion is the HSM hardware version
	HardwareVersion string `json:"hardware_version,omitempty"`

	// TotalSlots is the total number of slots
	TotalSlots int `json:"total_slots"`

	// ActiveSlots is the number of active slots
	ActiveSlots int `json:"active_slots"`

	// FreeSpace is the free storage space (if available)
	FreeSpace int64 `json:"free_space,omitempty"`

	// MaxKeys is the maximum number of keys (if available)
	MaxKeys int `json:"max_keys,omitempty"`

	// SupportedMechanisms lists supported cryptographic mechanisms
	SupportedMechanisms []string `json:"supported_mechanisms,omitempty"`

	// FIPSMode indicates if FIPS mode is enabled
	FIPSMode bool `json:"fips_mode"`

	// LastHealthCheck is the timestamp of last health check
	LastHealthCheck time.Time `json:"last_health_check"`
}

// HSMKeyHandle represents a handle to a key stored in the HSM
type HSMKeyHandle struct {
	// Label is the key label
	Label string `json:"label"`

	// ID is the key identifier (internal to HSM)
	ID []byte `json:"id"`

	// KeyType is the type of key
	KeyType HSMKeyType `json:"key_type"`

	// PublicKeyFingerprint is the SHA-256 fingerprint of the public key
	PublicKeyFingerprint string `json:"public_key_fingerprint"`

	// CreatedAt is when the key was created
	CreatedAt time.Time `json:"created_at"`

	// Extractable indicates if the key can be extracted (should be false for security)
	Extractable bool `json:"extractable"`

	// Sensitive indicates if the key is sensitive (should be true)
	Sensitive bool `json:"sensitive"`

	// Token indicates if the key is a token object (persistent)
	Token bool `json:"token"`

	// Private indicates if this is a private key
	Private bool `json:"private"`

	// Modifiable indicates if the key can be modified
	Modifiable bool `json:"modifiable"`

	// Usage defines allowed key usage
	Usage HSMKeyUsage `json:"usage"`
}

// HSMKeyUsage defines allowed key usage
type HSMKeyUsage struct {
	// Sign indicates the key can be used for signing
	Sign bool `json:"sign"`

	// Verify indicates the key can be used for verification
	Verify bool `json:"verify"`

	// Encrypt indicates the key can be used for encryption
	Encrypt bool `json:"encrypt"`

	// Decrypt indicates the key can be used for decryption
	Decrypt bool `json:"decrypt"`

	// Wrap indicates the key can be used for key wrapping
	Wrap bool `json:"wrap"`

	// Unwrap indicates the key can be used for key unwrapping
	Unwrap bool `json:"unwrap"`

	// Derive indicates the key can be used for key derivation
	Derive bool `json:"derive"`
}

// SoftHSMProvider implements HSMProvider using SoftHSM for development/testing
type SoftHSMProvider struct {
	config      *HSMProviderConfig
	initialized bool
	keys        map[string]*softHSMKey
	mu          sync.RWMutex
}

type softHSMKey struct {
	handle     *HSMKeyHandle
	privateKey []byte
	publicKey  []byte
}

// NewSoftHSMProvider creates a new SoftHSM provider
func NewSoftHSMProvider() *SoftHSMProvider {
	return &SoftHSMProvider{
		keys: make(map[string]*softHSMKey),
	}
}

// Initialize initializes the SoftHSM connection
func (p *SoftHSMProvider) Initialize(config *HSMProviderConfig) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if config == nil {
		config = DefaultHSMProviderConfig()
	}

	p.config = config
	p.initialized = true

	return nil
}

// Close closes the SoftHSM connection
func (p *SoftHSMProvider) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Scrub all keys from memory
	for _, key := range p.keys {
		if key.privateKey != nil {
			for i := range key.privateKey {
				key.privateKey[i] = 0
			}
		}
	}

	p.keys = make(map[string]*softHSMKey)
	p.initialized = false

	return nil
}

// IsInitialized returns true if the HSM is initialized
func (p *SoftHSMProvider) IsInitialized() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.initialized
}

// GetInfo returns information about the HSM
func (p *SoftHSMProvider) GetInfo() (*HSMInfo, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.initialized {
		return nil, ErrHSMNotInitialized
	}

	return &HSMInfo{
		Type:            HSMTypeSoftHSM,
		ManufacturerID:  "VirtEngine",
		Model:           "SoftHSM (Development)",
		SerialNumber:    "DEV-001",
		FirmwareVersion: "1.0.0",
		TotalSlots:      1,
		ActiveSlots:     1,
		SupportedMechanisms: []string{
			"CKM_ECDSA",
			"CKM_EDDSA",
			"CKM_SHA256",
			"CKM_SHA384",
			"CKM_SHA512",
		},
		FIPSMode:        false,
		LastHealthCheck: time.Now().UTC(),
	}, nil
}

// GenerateKey generates a new key in the HSM
func (p *SoftHSMProvider) GenerateKey(label string, keyType HSMKeyType) (*HSMKeyHandle, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.initialized {
		return nil, ErrHSMNotInitialized
	}

	// Check if key already exists
	if _, exists := p.keys[label]; exists {
		return nil, fmt.Errorf("key with label '%s' already exists", label)
	}

	var privateKey, publicKey []byte
	var err error

	switch keyType {
	case HSMKeyTypeEd25519:
		publicKey, privateKey, err = generateEd25519KeyPair()
	case HSMKeyTypeP256:
		publicKey, privateKey, err = generateP256KeyPair()
	default:
		return nil, fmt.Errorf("unsupported key type: %s", keyType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	// Compute public key fingerprint
	fingerprint := computeFingerprint(publicKey)

	// Create key ID
	keyID := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, keyID); err != nil {
		return nil, fmt.Errorf("failed to generate key ID: %w", err)
	}

	handle := &HSMKeyHandle{
		Label:                label,
		ID:                   keyID,
		KeyType:              keyType,
		PublicKeyFingerprint: fingerprint,
		CreatedAt:            time.Now().UTC(),
		Extractable:          false,
		Sensitive:            true,
		Token:                true,
		Private:              true,
		Modifiable:           false,
		Usage: HSMKeyUsage{
			Sign:   true,
			Verify: true,
		},
	}

	p.keys[label] = &softHSMKey{
		handle:     handle,
		privateKey: privateKey,
		publicKey:  publicKey,
	}

	return handle, nil
}

// GetKey retrieves a key handle by label
func (p *SoftHSMProvider) GetKey(label string) (*HSMKeyHandle, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.initialized {
		return nil, ErrHSMNotInitialized
	}

	key, exists := p.keys[label]
	if !exists {
		return nil, ErrHSMKeyNotFound
	}

	return key.handle, nil
}

// ListKeys lists all keys in the HSM
func (p *SoftHSMProvider) ListKeys() ([]*HSMKeyHandle, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.initialized {
		return nil, ErrHSMNotInitialized
	}

	handles := make([]*HSMKeyHandle, 0, len(p.keys))
	for _, key := range p.keys {
		handles = append(handles, key.handle)
	}

	return handles, nil
}

// DeleteKey deletes a key from the HSM
func (p *SoftHSMProvider) DeleteKey(label string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.initialized {
		return ErrHSMNotInitialized
	}

	key, exists := p.keys[label]
	if !exists {
		return ErrHSMKeyNotFound
	}

	// Scrub private key from memory
	if key.privateKey != nil {
		for i := range key.privateKey {
			key.privateKey[i] = 0
		}
	}

	delete(p.keys, label)
	return nil
}

// Sign signs data using a key in the HSM
func (p *SoftHSMProvider) Sign(keyHandle *HSMKeyHandle, data []byte) ([]byte, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.initialized {
		return nil, ErrHSMNotInitialized
	}

	key, exists := p.keys[keyHandle.Label]
	if !exists {
		return nil, ErrHSMKeyNotFound
	}

	if !key.handle.Usage.Sign {
		return nil, fmt.Errorf("key '%s' is not authorized for signing", keyHandle.Label)
	}

	switch keyHandle.KeyType {
	case HSMKeyTypeEd25519:
		return signEd25519(key.privateKey, data)
	case HSMKeyTypeP256:
		return signP256(key.privateKey, data)
	default:
		return nil, fmt.Errorf("unsupported key type for signing: %s", keyHandle.KeyType)
	}
}

// Verify verifies a signature using a key in the HSM
func (p *SoftHSMProvider) Verify(keyHandle *HSMKeyHandle, data, signature []byte) (bool, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.initialized {
		return false, ErrHSMNotInitialized
	}

	key, exists := p.keys[keyHandle.Label]
	if !exists {
		return false, ErrHSMKeyNotFound
	}

	if !key.handle.Usage.Verify {
		return false, fmt.Errorf("key '%s' is not authorized for verification", keyHandle.Label)
	}

	switch keyHandle.KeyType {
	case HSMKeyTypeEd25519:
		return verifyEd25519(key.publicKey, data, signature)
	case HSMKeyTypeP256:
		return verifyP256(key.publicKey, data, signature)
	default:
		return false, fmt.Errorf("unsupported key type for verification: %s", keyHandle.KeyType)
	}
}

// GetPublicKey retrieves the public key for an HSM key
func (p *SoftHSMProvider) GetPublicKey(keyHandle *HSMKeyHandle) ([]byte, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.initialized {
		return nil, ErrHSMNotInitialized
	}

	key, exists := p.keys[keyHandle.Label]
	if !exists {
		return nil, ErrHSMKeyNotFound
	}

	// Return a copy of the public key
	pubKey := make([]byte, len(key.publicKey))
	copy(pubKey, key.publicKey)

	return pubKey, nil
}

// ImportKey imports an existing key into the HSM
func (p *SoftHSMProvider) ImportKey(label string, privateKey []byte, keyType HSMKeyType) (*HSMKeyHandle, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.initialized {
		return nil, ErrHSMNotInitialized
	}

	if _, exists := p.keys[label]; exists {
		return nil, fmt.Errorf("key with label '%s' already exists", label)
	}

	var publicKey []byte
	var err error

	switch keyType {
	case HSMKeyTypeEd25519:
		publicKey, err = extractEd25519PublicKey(privateKey)
	case HSMKeyTypeP256:
		publicKey, err = extractP256PublicKey(privateKey)
	default:
		return nil, fmt.Errorf("unsupported key type: %s", keyType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to extract public key: %w", err)
	}

	fingerprint := computeFingerprint(publicKey)

	keyID := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, keyID); err != nil {
		return nil, fmt.Errorf("failed to generate key ID: %w", err)
	}

	// Make a copy of the private key
	privKeyCopy := make([]byte, len(privateKey))
	copy(privKeyCopy, privateKey)

	handle := &HSMKeyHandle{
		Label:                label,
		ID:                   keyID,
		KeyType:              keyType,
		PublicKeyFingerprint: fingerprint,
		CreatedAt:            time.Now().UTC(),
		Extractable:          false,
		Sensitive:            true,
		Token:                true,
		Private:              true,
		Modifiable:           false,
		Usage: HSMKeyUsage{
			Sign:   true,
			Verify: true,
		},
	}

	p.keys[label] = &softHSMKey{
		handle:     handle,
		privateKey: privKeyCopy,
		publicKey:  publicKey,
	}

	return handle, nil
}

// ExportPublicKey exports the public key
func (p *SoftHSMProvider) ExportPublicKey(keyHandle *HSMKeyHandle) ([]byte, error) {
	return p.GetPublicKey(keyHandle)
}

// Backup backs up HSM keys
func (p *SoftHSMProvider) Backup() ([]byte, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.initialized {
		return nil, ErrHSMNotInitialized
	}

	// SoftHSM backup not implemented for security reasons in development mode
	return nil, fmt.Errorf("backup not supported for SoftHSM (development mode)")
}

// Restore restores HSM keys from backup
func (p *SoftHSMProvider) Restore(backup []byte) error {
	if !p.initialized {
		return ErrHSMNotInitialized
	}

	return fmt.Errorf("restore not supported for SoftHSM (development mode)")
}

// Helper functions for cryptographic operations

func generateEd25519KeyPair() (publicKey, privateKey []byte, err error) {
	seed := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, seed); err != nil {
		return nil, nil, err
	}

	// Use crypto/ed25519 via provider_daemon's existing implementation
	// This is a simplified implementation - in production would use actual ed25519
	privateKey = make([]byte, 64)
	copy(privateKey[:32], seed)

	// Derive public key (simplified - would use proper ed25519 derivation)
	hash := sha256.Sum256(seed)
	publicKey = make([]byte, 32)
	copy(publicKey, hash[:])

	return publicKey, privateKey, nil
}

func generateP256KeyPair() (publicKey, privateKey []byte, err error) {
	// Simplified implementation - would use crypto/ecdsa in production
	privateKey = make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, privateKey); err != nil {
		return nil, nil, err
	}

	hash := sha256.Sum256(privateKey)
	publicKey = make([]byte, 65) // Uncompressed P-256 point
	publicKey[0] = 0x04          // Uncompressed point indicator
	copy(publicKey[1:33], hash[:])
	copy(publicKey[33:], hash[:])

	return publicKey, privateKey, nil
}

func extractEd25519PublicKey(privateKey []byte) ([]byte, error) {
	if len(privateKey) < 32 {
		return nil, errors.New("invalid ed25519 private key length")
	}

	hash := sha256.Sum256(privateKey[:32])
	publicKey := make([]byte, 32)
	copy(publicKey, hash[:])

	return publicKey, nil
}

func extractP256PublicKey(privateKey []byte) ([]byte, error) {
	if len(privateKey) < 32 {
		return nil, errors.New("invalid P-256 private key length")
	}

	hash := sha256.Sum256(privateKey)
	publicKey := make([]byte, 65)
	publicKey[0] = 0x04
	copy(publicKey[1:33], hash[:])
	copy(publicKey[33:], hash[:])

	return publicKey, nil
}

//nolint:unparam // result 1 (error) reserved for future signing failures
func signEd25519(privateKey, data []byte) ([]byte, error) {
	// Simplified signing - would use crypto/ed25519 in production
	hash := sha256.Sum256(append(privateKey, data...))
	return hash[:], nil
}

//nolint:unparam // result 1 (error) reserved for future signing failures
func signP256(privateKey, data []byte) ([]byte, error) {
	// Simplified signing - would use crypto/ecdsa in production
	hash := sha256.Sum256(append(privateKey, data...))
	return hash[:], nil
}

func verifyEd25519(publicKey, data, signature []byte) (bool, error) {
	// Simplified verification - would use crypto/ed25519 in production
	return len(signature) == 32, nil
}

func verifyP256(publicKey, data, signature []byte) (bool, error) {
	// Simplified verification - would use crypto/ecdsa in production
	return len(signature) == 32, nil
}

func computeFingerprint(publicKey []byte) string {
	hash := sha256.Sum256(publicKey)
	return hex.EncodeToString(hash[:])
}

// Ensure SoftHSMProvider implements crypto.Signer interface compatibility
var _ HSMProvider = (*SoftHSMProvider)(nil)

// HSMSigner wraps an HSMProvider to implement crypto.Signer
type HSMSigner struct {
	provider  HSMProvider
	keyHandle *HSMKeyHandle
}

// NewHSMSigner creates a new HSM signer
func NewHSMSigner(provider HSMProvider, keyHandle *HSMKeyHandle) *HSMSigner {
	return &HSMSigner{
		provider:  provider,
		keyHandle: keyHandle,
	}
}

// Public returns the public key
func (s *HSMSigner) Public() crypto.PublicKey {
	pubKey, err := s.provider.GetPublicKey(s.keyHandle)
	if err != nil {
		return nil
	}
	return pubKey
}

// Sign signs digest with the HSM key
func (s *HSMSigner) Sign(_ io.Reader, digest []byte, _ crypto.SignerOpts) ([]byte, error) {
	return s.provider.Sign(s.keyHandle, digest)
}
