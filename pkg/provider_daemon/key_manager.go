// Package provider_daemon implements the provider daemon for VirtEngine.
//
// VE-400: Provider Daemon key management and transaction signing
package provider_daemon

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrKeyNotFound is returned when a key is not found
var ErrKeyNotFound = errors.New("key not found")

// ErrKeyExpired is returned when a key has expired
var ErrKeyExpired = errors.New("key expired")

// ErrKeyRevoked is returned when a key has been revoked
var ErrKeyRevoked = errors.New("key revoked")

// ErrKeyStorageLocked is returned when key storage is locked
var ErrKeyStorageLocked = errors.New("key storage is locked")

// ErrInvalidPassphrase is returned when the passphrase is invalid
var ErrInvalidPassphrase = errors.New("invalid passphrase")

const (
	keyStatusActive  = "active"
	keyStatusRotated = "rotated"
	keyStatusRevoked = "revoked"
)

// KeyStorageType represents the type of key storage
type KeyStorageType string

const (
	// KeyStorageTypeFile stores keys in encrypted files
	KeyStorageTypeFile KeyStorageType = "file"

	// KeyStorageTypeHardware uses hardware security modules
	KeyStorageTypeHardware KeyStorageType = "hardware"

	// KeyStorageTypeLedger uses Ledger hardware wallets
	KeyStorageTypeLedger KeyStorageType = "ledger"

	// KeyStorageTypeNonCustodial uses external signing services
	KeyStorageTypeNonCustodial KeyStorageType = "non_custodial"

	// KeyStorageTypeMemory stores keys in memory (for testing)
	KeyStorageTypeMemory KeyStorageType = "memory"
)

// KeyManagerConfig configures the key manager
type KeyManagerConfig struct {
	// StorageType is the type of key storage
	StorageType KeyStorageType `json:"storage_type"`

	// KeyDir is the directory for file-based key storage
	KeyDir string `json:"key_dir,omitempty"`

	// DefaultAlgorithm is the default signing algorithm
	DefaultAlgorithm string `json:"default_algorithm"`

	// KeyRotationDays is the number of days before key rotation is recommended
	KeyRotationDays int `json:"key_rotation_days"`

	// GracePeriodHours is the grace period after rotation
	GracePeriodHours int `json:"grace_period_hours"`

	// HSMConfig is the hardware security module configuration
	HSMConfig *HSMConfig `json:"hsm_config,omitempty"`

	// LedgerConfig is the Ledger device configuration
	LedgerConfig *LedgerConfig `json:"ledger_config,omitempty"`
}

// HSMConfig contains hardware security module configuration
type HSMConfig struct {
	// LibraryPath is the path to the PKCS#11 library
	LibraryPath string `json:"library_path"`

	// SlotID is the HSM slot ID
	SlotID uint `json:"slot_id"`

	// TokenLabel is the HSM token label
	TokenLabel string `json:"token_label"`
}

// LedgerConfig contains Ledger device configuration
type LedgerConfig struct {
	// DerivationPath is the HD derivation path
	DerivationPath string `json:"derivation_path"`

	// RequireConfirmation requires user confirmation for each signature
	RequireConfirmation bool `json:"require_confirmation"`
}

// DefaultKeyManagerConfig returns the default key manager configuration
func DefaultKeyManagerConfig() KeyManagerConfig {
	return KeyManagerConfig{
		StorageType:      KeyStorageTypeFile,
		DefaultAlgorithm: string(HSMKeyTypeEd25519),
		KeyRotationDays:  90,
		GracePeriodHours: 24,
	}
}

// ManagedKey represents a managed signing key
type ManagedKey struct {
	// KeyID is the unique identifier for this key
	KeyID string `json:"key_id"`

	// PublicKey is the public key (hex encoded)
	PublicKey string `json:"public_key"`

	// Algorithm is the key algorithm
	Algorithm string `json:"algorithm"`

	// CreatedAt is when the key was created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when the key expires (zero means no expiry)
	ExpiresAt time.Time `json:"expires_at,omitempty"`

	// Status is the key status (active, rotated, revoked)
	Status string `json:"status"`

	// ProviderAddress is the associated provider address
	ProviderAddress string `json:"provider_address"`

	// privateKey is the private key (never exposed)
	privateKey []byte
}

// KeyManager manages provider signing keys
type KeyManager struct {
	config   KeyManagerConfig
	keys     map[string]*ManagedKey
	activeID string
	mu       sync.RWMutex
	locked   bool
}

// NewKeyManager creates a new key manager with the given configuration
func NewKeyManager(config KeyManagerConfig) (*KeyManager, error) {
	km := &KeyManager{
		config: config,
		keys:   make(map[string]*ManagedKey),
		locked: true,
	}

	return km, nil
}

// Unlock unlocks the key manager with a passphrase
// For file-based storage, this decrypts the keys
// For hardware storage, this verifies HSM/Ledger availability
func (km *KeyManager) Unlock(passphrase string) error {
	km.mu.Lock()
	defer km.mu.Unlock()

	if km.config.StorageType == KeyStorageTypeMemory {
		// Memory storage doesn't require a passphrase
		km.locked = false
		return nil
	}

	// For file-based storage, we would decrypt keys here
	// For hardware storage, we would verify device connectivity
	// This is a simplified implementation

	if passphrase == "" {
		return ErrInvalidPassphrase
	}

	km.locked = false
	return nil
}

// Lock locks the key manager, scrubbing keys from memory
func (km *KeyManager) Lock() {
	km.mu.Lock()
	defer km.mu.Unlock()

	// Scrub all private keys from memory
	for _, key := range km.keys {
		if key.privateKey != nil {
			for i := range key.privateKey {
				key.privateKey[i] = 0
			}
			key.privateKey = nil
		}
	}

	km.locked = true
}

// IsLocked returns true if the key manager is locked
func (km *KeyManager) IsLocked() bool {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return km.locked
}

// GenerateKey generates a new signing key
func (km *KeyManager) GenerateKey(providerAddress string) (*ManagedKey, error) {
	km.mu.Lock()
	defer km.mu.Unlock()

	if km.locked {
		return nil, ErrKeyStorageLocked
	}

	switch km.config.DefaultAlgorithm {
	case string(HSMKeyTypeEd25519):
		return km.generateEd25519Key(providerAddress)
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", km.config.DefaultAlgorithm)
	}
}

func (km *KeyManager) generateEd25519Key(providerAddress string) (*ManagedKey, error) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ed25519 key: %w", err)
	}

	keyID := generateKeyID(pubKey)
	now := time.Now().UTC()

	key := &ManagedKey{
		KeyID:           keyID,
		PublicKey:       hex.EncodeToString(pubKey),
		Algorithm:       string(HSMKeyTypeEd25519),
		CreatedAt:       now,
		Status:          keyStatusActive,
		ProviderAddress: providerAddress,
		privateKey:      privKey,
	}

	if km.config.KeyRotationDays > 0 {
		key.ExpiresAt = now.Add(time.Duration(km.config.KeyRotationDays) * 24 * time.Hour)
	}

	km.keys[keyID] = key
	km.activeID = keyID

	return key, nil
}

// generateKeyID generates a unique key ID from the public key
func generateKeyID(pubKey []byte) string {
	hash := sha256.Sum256(pubKey)
	return hex.EncodeToString(hash[:8])
}

// GetActiveKey returns the currently active key
func (km *KeyManager) GetActiveKey() (*ManagedKey, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	if km.locked {
		return nil, ErrKeyStorageLocked
	}

	if km.activeID == "" {
		return nil, ErrKeyNotFound
	}

	key, ok := km.keys[km.activeID]
	if !ok {
		return nil, ErrKeyNotFound
	}

	if key.Status == keyStatusRevoked {
		return nil, ErrKeyRevoked
	}

	if !key.ExpiresAt.IsZero() && time.Now().After(key.ExpiresAt) {
		return nil, ErrKeyExpired
	}

	return key, nil
}

// GetKey returns a key by ID
func (km *KeyManager) GetKey(keyID string) (*ManagedKey, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	if km.locked {
		return nil, ErrKeyStorageLocked
	}

	key, ok := km.keys[keyID]
	if !ok {
		return nil, ErrKeyNotFound
	}

	return key, nil
}

// ListKeys returns all keys (without private key data)
func (km *KeyManager) ListKeys() ([]*ManagedKey, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	if km.locked {
		return nil, ErrKeyStorageLocked
	}

	keys := make([]*ManagedKey, 0, len(km.keys))
	for _, key := range km.keys {
		// Return a copy without the private key
		keyCopy := &ManagedKey{
			KeyID:           key.KeyID,
			PublicKey:       key.PublicKey,
			Algorithm:       key.Algorithm,
			CreatedAt:       key.CreatedAt,
			ExpiresAt:       key.ExpiresAt,
			Status:          key.Status,
			ProviderAddress: key.ProviderAddress,
		}
		keys = append(keys, keyCopy)
	}

	return keys, nil
}

// Sign signs a message with the active key
func (km *KeyManager) Sign(message []byte) (*Signature, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	if km.locked {
		return nil, ErrKeyStorageLocked
	}

	key, err := km.getActiveKeyInternal()
	if err != nil {
		return nil, err
	}

	return km.signWithKey(key, message)
}

// SignWithKey signs a message with a specific key
func (km *KeyManager) SignWithKey(keyID string, message []byte) (*Signature, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	if km.locked {
		return nil, ErrKeyStorageLocked
	}

	key, ok := km.keys[keyID]
	if !ok {
		return nil, ErrKeyNotFound
	}

	return km.signWithKey(key, message)
}

func (km *KeyManager) getActiveKeyInternal() (*ManagedKey, error) {
	if km.activeID == "" {
		return nil, ErrKeyNotFound
	}

	key, ok := km.keys[km.activeID]
	if !ok {
		return nil, ErrKeyNotFound
	}

	if key.Status == keyStatusRevoked {
		return nil, ErrKeyRevoked
	}

	if !key.ExpiresAt.IsZero() && time.Now().After(key.ExpiresAt) {
		return nil, ErrKeyExpired
	}

	return key, nil
}

func (km *KeyManager) signWithKey(key *ManagedKey, message []byte) (*Signature, error) {
	if key.privateKey == nil {
		return nil, errors.New("private key not loaded")
	}

	var sigBytes []byte

	switch key.Algorithm {
	case string(HSMKeyTypeEd25519):
		sigBytes = ed25519.Sign(key.privateKey, message)
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", key.Algorithm)
	}

	return &Signature{
		PublicKey: key.PublicKey,
		Signature: hex.EncodeToString(sigBytes),
		Algorithm: key.Algorithm,
		KeyID:     key.KeyID,
		SignedAt:  time.Now().UTC(),
	}, nil
}

// Signature represents a cryptographic signature
type Signature struct {
	// PublicKey is the public key used for signing (hex encoded)
	PublicKey string `json:"public_key"`

	// Signature is the cryptographic signature (hex encoded)
	Signature string `json:"signature"`

	// Algorithm is the signing algorithm used
	Algorithm string `json:"algorithm"`

	// KeyID is the identifier for the key used
	KeyID string `json:"key_id"`

	// SignedAt is when the signature was created
	SignedAt time.Time `json:"signed_at"`
}

// Verify verifies the signature against the provided message
func (s *Signature) Verify(message []byte) error {
	pubKeyBytes, err := hex.DecodeString(s.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to decode public key: %w", err)
	}

	sigBytes, err := hex.DecodeString(s.Signature)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	switch s.Algorithm {
	case string(HSMKeyTypeEd25519):
		if len(pubKeyBytes) != ed25519.PublicKeySize {
			return fmt.Errorf("invalid ed25519 public key size: %d", len(pubKeyBytes))
		}
		if !ed25519.Verify(pubKeyBytes, message, sigBytes) {
			return errors.New("signature verification failed")
		}
		return nil

	default:
		return fmt.Errorf("unsupported algorithm: %s", s.Algorithm)
	}
}

// RotateKey rotates the active key to a new key
func (km *KeyManager) RotateKey(providerAddress string) (*ManagedKey, *KeyRotation, error) {
	km.mu.Lock()
	defer km.mu.Unlock()

	if km.locked {
		return nil, nil, ErrKeyStorageLocked
	}

	// Get the current active key
	var oldKey *ManagedKey
	if km.activeID != "" {
		oldKey = km.keys[km.activeID]
	}

	// Generate a new key
	newKey, err := km.generateEd25519Key(providerAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate new key: %w", err)
	}

	// Create rotation record
	now := time.Now().UTC()
	rotation := &KeyRotation{
		OldKeyID:       "",
		NewKeyID:       newKey.KeyID,
		RotatedAt:      now,
		GracePeriodEnd: now.Add(time.Duration(km.config.GracePeriodHours) * time.Hour),
	}

	if oldKey != nil {
		rotation.OldKeyID = oldKey.KeyID
		oldKey.Status = keyStatusRotated
	}

	return newKey, rotation, nil
}

// KeyRotation represents a key rotation event
type KeyRotation struct {
	// OldKeyID is the ID of the old key
	OldKeyID string `json:"old_key_id"`

	// NewKeyID is the ID of the new key
	NewKeyID string `json:"new_key_id"`

	// RotatedAt is when the rotation occurred
	RotatedAt time.Time `json:"rotated_at"`

	// GracePeriodEnd is when the old key becomes invalid
	GracePeriodEnd time.Time `json:"grace_period_end"`
}

// RevokeKey revokes a key by ID
func (km *KeyManager) RevokeKey(keyID string) error {
	km.mu.Lock()
	defer km.mu.Unlock()

	if km.locked {
		return ErrKeyStorageLocked
	}

	key, ok := km.keys[keyID]
	if !ok {
		return ErrKeyNotFound
	}

	key.Status = keyStatusRevoked

	// Scrub private key from memory
	if key.privateKey != nil {
		for i := range key.privateKey {
			key.privateKey[i] = 0
		}
		key.privateKey = nil
	}

	// If this was the active key, clear active
	if km.activeID == keyID {
		km.activeID = ""
	}

	return nil
}

// NeedsRotation checks if the active key needs rotation
func (km *KeyManager) NeedsRotation() (bool, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	if km.locked {
		return false, ErrKeyStorageLocked
	}

	if km.activeID == "" {
		return true, nil // No active key, rotation needed
	}

	key, ok := km.keys[km.activeID]
	if !ok {
		return true, nil
	}

	// Check if key is within rotation window (7 days before expiry)
	if !key.ExpiresAt.IsZero() {
		rotationWindow := key.ExpiresAt.Add(-7 * 24 * time.Hour)
		if time.Now().After(rotationWindow) {
			return true, nil
		}
	}

	return false, nil
}

// ImportKey imports an existing key (for testing or migration)
func (km *KeyManager) ImportKey(providerAddress string, privateKey []byte, algorithm string) (*ManagedKey, error) {
	km.mu.Lock()
	defer km.mu.Unlock()

	if km.locked {
		return nil, ErrKeyStorageLocked
	}

	var pubKey []byte

	switch algorithm {
	case string(HSMKeyTypeEd25519):
		if len(privateKey) != ed25519.PrivateKeySize {
			return nil, fmt.Errorf("invalid ed25519 private key size: %d", len(privateKey))
		}
		pubKey = make([]byte, ed25519.PublicKeySize)
		copy(pubKey, privateKey[32:])
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}

	keyID := generateKeyID(pubKey)
	now := time.Now().UTC()

	key := &ManagedKey{
		KeyID:           keyID,
		PublicKey:       hex.EncodeToString(pubKey),
		Algorithm:       algorithm,
		CreatedAt:       now,
		Status:          keyStatusActive,
		ProviderAddress: providerAddress,
		privateKey:      privateKey,
	}

	if km.config.KeyRotationDays > 0 {
		key.ExpiresAt = now.Add(time.Duration(km.config.KeyRotationDays) * 24 * time.Hour)
	}

	km.keys[keyID] = key
	km.activeID = keyID

	return key, nil
}
