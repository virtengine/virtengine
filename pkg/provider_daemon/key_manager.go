// Package provider_daemon implements the provider daemon for VirtEngine.
//
// VE-400: Provider Daemon key management and transaction signing
package provider_daemon

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	cosmosed25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	providertypes "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
	"golang.org/x/crypto/argon2"
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

// ErrProviderKeyMismatch is returned when local custody does not match the
// active governed x/provider signing-key epoch.
var ErrProviderKeyMismatch = errors.New("local provider key does not match active on-chain epoch")

const (
	keyStatusActive  = "active"
	keyStatusRotated = "rotated"
	keyStatusRevoked = "revoked"

	fileKeyStoreVersion = 1
	fileKeyStoreName    = "provider-keys.enc.json"
	fileKeyStoreKDF     = "argon2id-v1"
	fileKeyStoreCipher  = "aes-256-gcm"
	fileKeyStoreTime    = uint32(3)
	fileKeyStoreMemory  = uint32(64 * 1024)
	fileKeyStoreThreads = uint8(4)
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
		KeyDir:           ".virtengine/keys",
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
	config              KeyManagerConfig
	keys                map[string]*ManagedKey
	activeID            string
	mu                  sync.RWMutex
	locked              bool
	fileKey             []byte
	fileSalt            []byte
	expectedFingerprint string
}

type persistedManagedKey struct {
	KeyID           string    `json:"key_id"`
	PublicKey       string    `json:"public_key"`
	Algorithm       string    `json:"algorithm"`
	CreatedAt       time.Time `json:"created_at"`
	ExpiresAt       time.Time `json:"expires_at,omitempty"`
	Status          string    `json:"status"`
	ProviderAddress string    `json:"provider_address"`
	PrivateKey      string    `json:"private_key"`
}

type fileKeyStorePayload struct {
	Version  uint32                `json:"version"`
	ActiveID string                `json:"active_id"`
	Keys     []persistedManagedKey `json:"keys"`
}

type fileKeyStoreEnvelope struct {
	Version        uint32 `json:"version"`
	KDF            string `json:"kdf"`
	Cipher         string `json:"cipher"`
	ArgonTime      uint32 `json:"argon_time"`
	ArgonMemoryKiB uint32 `json:"argon_memory_kib"`
	ArgonThreads   uint8  `json:"argon_threads"`
	Salt           string `json:"salt"`
	Nonce          string `json:"nonce"`
	Ciphertext     string `json:"ciphertext"`
	MetadataMAC    string `json:"metadata_mac"`
}

// ManagedKeyAccountAddress returns the Cosmos account address controlled by
// the managed SDK signing key.
func ManagedKeyAccountAddress(key *ManagedKey) (string, error) {
	if key == nil {
		return "", ErrKeyNotFound
	}
	publicKey, err := hex.DecodeString(key.PublicKey)
	if err != nil {
		return "", fmt.Errorf("decode managed public key: %w", err)
	}
	switch key.Algorithm {
	case string(HSMKeyTypeEd25519):
		if len(publicKey) != ed25519.PublicKeySize {
			return "", fmt.Errorf("invalid ed25519 public key size %d", len(publicKey))
		}
		return sdk.AccAddress((&cosmosed25519.PubKey{Key: publicKey}).Address()).String(), nil
	default:
		return "", fmt.Errorf("unsupported SDK signer algorithm %s", key.Algorithm)
	}
}

// NewKeyManager creates a new key manager with the given configuration
func NewKeyManager(config KeyManagerConfig) (*KeyManager, error) {
	if config.DefaultAlgorithm == "" {
		config.DefaultAlgorithm = string(HSMKeyTypeEd25519)
	}
	switch config.StorageType {
	case KeyStorageTypeFile:
		if strings.TrimSpace(config.KeyDir) == "" {
			return nil, fmt.Errorf("file key directory is required")
		}
	case KeyStorageTypeMemory:
	case KeyStorageTypeHardware, KeyStorageTypeLedger, KeyStorageTypeNonCustodial:
		return nil, fmt.Errorf("key storage backend %q is not implemented; refusing software fallback", config.StorageType)
	default:
		return nil, fmt.Errorf("unsupported key storage backend %q", config.StorageType)
	}
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
	if passphrase == "" {
		return ErrInvalidPassphrase
	}
	keyPath, err := km.fileKeyStorePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(keyPath), 0o700); err != nil {
		return fmt.Errorf("create key store directory: %w", err)
	}
	data, err := os.ReadFile(keyPath) // #nosec G304 -- keyPath is rooted in the configured key directory.
	if errors.Is(err, os.ErrNotExist) {
		km.fileSalt = make([]byte, 32)
		if _, err := io.ReadFull(rand.Reader, km.fileSalt); err != nil {
			return fmt.Errorf("generate key store salt: %w", err)
		}
		km.fileKey = deriveFileKey(passphrase, km.fileSalt)
		km.locked = false
		return nil
	}
	if err != nil {
		return fmt.Errorf("read encrypted key store: %w", err)
	}
	fileKey, payload, err := decryptFileKeyStore(passphrase, data)
	if err != nil {
		return err
	}
	keys, err := managedKeysFromPayload(payload)
	if err != nil {
		scrubBytes(fileKey)
		return err
	}
	km.keys = keys
	km.activeID = payload.ActiveID
	km.fileKey = fileKey
	salt, _ := base64.StdEncoding.DecodeString(envelopeSalt(data))
	km.fileSalt = salt
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
	scrubBytes(km.fileKey)
	scrubBytes(km.fileSalt)
	km.fileKey = nil
	km.fileSalt = nil

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

	previousActiveID := km.activeID
	var key *ManagedKey
	var err error
	switch km.config.DefaultAlgorithm {
	case string(HSMKeyTypeEd25519):
		key, err = km.generateEd25519Key(providerAddress)
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", km.config.DefaultAlgorithm)
	}
	if err != nil {
		return nil, err
	}
	if err := km.persistLocked(); err != nil {
		delete(km.keys, key.KeyID)
		km.activeID = previousActiveID
		scrubBytes(key.privateKey)
		return nil, err
	}
	return key, nil
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
	return providertypes.ComputeProviderKeyID(providertypes.PublicKeyTypeEd25519, pubKey)
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

// ActiveProviderKeyBinding is the non-secret on-chain key state a caller has
// resolved before signing a canonical usage proof.
type ActiveProviderKeyBinding struct {
	ProviderAddress string
	KeyID           string
	Epoch           uint64
	PublicKey       []byte
	Algorithm       string
	BlockHeight     int64
	BlockTime       time.Time
}

// SignForProviderKey signs only when local custody exactly matches the active
// governed provider epoch returned by a chain resolver.
func (km *KeyManager) SignForProviderKey(message []byte, binding ActiveProviderKeyBinding) (*Signature, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()
	if km.locked {
		return nil, ErrKeyStorageLocked
	}
	key, err := km.getActiveKeyInternal()
	if err != nil {
		return nil, err
	}
	publicKey, err := hex.DecodeString(key.PublicKey)
	if err != nil {
		return nil, ErrProviderKeyMismatch
	}
	if binding.ProviderAddress == "" || binding.KeyID == "" || binding.Epoch == 0 ||
		key.ProviderAddress != binding.ProviderAddress || key.KeyID != binding.KeyID ||
		key.Algorithm != binding.Algorithm || len(publicKey) != len(binding.PublicKey) ||
		subtle.ConstantTimeCompare(publicKey, binding.PublicKey) != 1 {
		return nil, ErrProviderKeyMismatch
	}
	return km.signWithKey(key, message)
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
	previousActiveID := km.activeID
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
	if err := km.persistLocked(); err != nil {
		delete(km.keys, newKey.KeyID)
		scrubBytes(newKey.privateKey)
		if oldKey != nil {
			oldKey.Status = keyStatusActive
			km.activeID = previousActiveID
		} else {
			km.activeID = ""
		}
		return nil, nil, err
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

	previousStatus := key.Status
	previousActiveID := km.activeID
	key.Status = keyStatusRevoked

	// If this was the active key, clear active
	if km.activeID == keyID {
		km.activeID = ""
	}
	if err := km.persistLocked(); err != nil {
		key.Status = previousStatus
		km.activeID = previousActiveID
		return err
	}
	scrubBytes(key.privateKey)
	key.privateKey = nil
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

	if len(privateKey) == 0 {
		return nil, fmt.Errorf("private key is required")
	}

	privateKeyCopy := make([]byte, len(privateKey))
	copy(privateKeyCopy, privateKey)

	var pubKey []byte

	switch algorithm {
	case string(HSMKeyTypeEd25519):
		if len(privateKeyCopy) != ed25519.PrivateKeySize {
			return nil, fmt.Errorf("invalid ed25519 private key size: %d", len(privateKeyCopy))
		}
		pubKey = make([]byte, ed25519.PublicKeySize)
		copy(pubKey, privateKeyCopy[32:])
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
		privateKey:      privateKeyCopy,
	}

	if km.config.KeyRotationDays > 0 {
		key.ExpiresAt = now.Add(time.Duration(km.config.KeyRotationDays) * 24 * time.Hour)
	}

	previousActiveID := km.activeID
	previousKey, replaced := km.keys[keyID]
	km.keys[keyID] = key
	km.activeID = keyID
	if err := km.persistLocked(); err != nil {
		if replaced {
			km.keys[keyID] = previousKey
		} else {
			delete(km.keys, keyID)
		}
		km.activeID = previousActiveID
		scrubBytes(key.privateKey)
		return nil, err
	}
	return key, nil
}

// ActiveKeyFingerprint returns the stable SHA-256 fingerprint used by startup
// and readiness identity checks. It contains no private material.
func (km *KeyManager) ActiveKeyFingerprint() (string, error) {
	key, err := km.GetActiveKey()
	if err != nil {
		return "", err
	}
	publicKey, err := hex.DecodeString(key.PublicKey)
	if err != nil {
		return "", fmt.Errorf("decode active public key: %w", err)
	}
	digest := sha256.Sum256(publicKey)
	return hex.EncodeToString(digest[:]), nil
}

// SetExpectedActiveKeyFingerprint binds readiness to deployment identity
// metadata. The value is public and compared exactly after normalization.
func (km *KeyManager) SetExpectedActiveKeyFingerprint(fingerprint string) error {
	fingerprint = strings.ToLower(strings.TrimSpace(fingerprint))
	if len(fingerprint) != sha256.Size*2 {
		return fmt.Errorf("expected provider key fingerprint must be a SHA-256 hex digest")
	}
	if _, err := hex.DecodeString(fingerprint); err != nil {
		return fmt.Errorf("expected provider key fingerprint must be hexadecimal: %w", err)
	}
	km.mu.Lock()
	km.expectedFingerprint = fingerprint
	km.mu.Unlock()
	actual, err := km.ActiveKeyFingerprint()
	if err != nil {
		return err
	}
	if actual != fingerprint {
		return ErrProviderKeyMismatch
	}
	return nil
}

// Ready validates that custody is unlocked, an active signing key is loaded,
// and any deployment-bound fingerprint still matches.
func (km *KeyManager) Ready() bool {
	if km == nil || km.IsLocked() {
		return false
	}
	km.mu.RLock()
	expected := km.expectedFingerprint
	km.mu.RUnlock()
	fingerprint, err := km.ActiveKeyFingerprint()
	if err != nil {
		return false
	}
	return expected == "" || fingerprint == expected
}

func (km *KeyManager) fileKeyStorePath() (string, error) {
	absolute, err := filepath.Abs(km.config.KeyDir)
	if err != nil {
		return "", fmt.Errorf("resolve key store directory: %w", err)
	}
	return filepath.Join(filepath.Clean(absolute), fileKeyStoreName), nil
}

func (km *KeyManager) persistLocked() error {
	if km.config.StorageType != KeyStorageTypeFile {
		return nil
	}
	if km.locked || len(km.fileKey) == 0 {
		return ErrKeyStorageLocked
	}
	payload := fileKeyStorePayload{Version: fileKeyStoreVersion, ActiveID: km.activeID, Keys: make([]persistedManagedKey, 0, len(km.keys))}
	ids := make([]string, 0, len(km.keys))
	for id := range km.keys {
		ids = append(ids, id)
	}
	sortStrings(ids)
	for _, id := range ids {
		key := km.keys[id]
		payload.Keys = append(payload.Keys, persistedManagedKey{
			KeyID: key.KeyID, PublicKey: key.PublicKey, Algorithm: key.Algorithm,
			CreatedAt: key.CreatedAt, ExpiresAt: key.ExpiresAt, Status: key.Status,
			ProviderAddress: key.ProviderAddress, PrivateKey: base64.StdEncoding.EncodeToString(key.privateKey),
		})
	}
	data, err := encryptFileKeyStore(km.fileKey, km.fileSalt, payload)
	if err != nil {
		return err
	}
	path, err := km.fileKeyStorePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create key store directory: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".provider-keys-*.tmp")
	if err != nil {
		return fmt.Errorf("create key store temporary file: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write encrypted key store: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := atomicReplaceFile(tmpPath, path); err != nil {
		return fmt.Errorf("replace encrypted key store: %w", err)
	}
	if dir, err := os.Open(filepath.Dir(path)); err == nil { // #nosec G304 -- configured key directory.
		_ = dir.Sync()
		_ = dir.Close()
	}
	return nil
}

func deriveFileKey(passphrase string, salt []byte) []byte {
	return argon2.IDKey([]byte(passphrase), salt, fileKeyStoreTime, fileKeyStoreMemory, fileKeyStoreThreads, 32)
}

func encryptFileKeyStore(key, salt []byte, payload fileKeyStorePayload) ([]byte, error) {
	if len(key) != 32 || len(salt) != 32 {
		return nil, ErrInvalidPassphrase
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate key store nonce: %w", err)
	}
	plaintext, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode key store payload: %w", err)
	}
	defer scrubBytes(plaintext)
	envelope := fileKeyStoreEnvelope{
		Version: fileKeyStoreVersion, KDF: fileKeyStoreKDF, Cipher: fileKeyStoreCipher,
		ArgonTime: fileKeyStoreTime, ArgonMemoryKiB: fileKeyStoreMemory, ArgonThreads: fileKeyStoreThreads,
		Salt: base64.StdEncoding.EncodeToString(salt), Nonce: base64.StdEncoding.EncodeToString(nonce),
	}
	aad := fileKeyStoreAAD(envelope)
	envelope.Ciphertext = base64.StdEncoding.EncodeToString(aead.Seal(nil, nonce, plaintext, aad))
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write(aad)
	envelope.MetadataMAC = hex.EncodeToString(mac.Sum(nil))
	return json.MarshalIndent(envelope, "", "  ")
}

func decryptFileKeyStore(passphrase string, data []byte) ([]byte, fileKeyStorePayload, error) {
	var envelope fileKeyStoreEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fileKeyStorePayload{}, fmt.Errorf("decode encrypted key store: %w", err)
	}
	if envelope.Version != fileKeyStoreVersion || envelope.KDF != fileKeyStoreKDF || envelope.Cipher != fileKeyStoreCipher ||
		envelope.ArgonTime != fileKeyStoreTime || envelope.ArgonMemoryKiB != fileKeyStoreMemory || envelope.ArgonThreads != fileKeyStoreThreads {
		return nil, fileKeyStorePayload{}, fmt.Errorf("unsupported or unsafe encrypted key store profile")
	}
	salt, err := base64.StdEncoding.DecodeString(envelope.Salt)
	if err != nil || len(salt) != 32 {
		return nil, fileKeyStorePayload{}, fmt.Errorf("invalid key store salt")
	}
	key := deriveFileKey(passphrase, salt)
	fail := func() error {
		scrubBytes(key)
		return ErrInvalidPassphrase
	}
	aad := fileKeyStoreAAD(envelope)
	expectedMAC, err := hex.DecodeString(envelope.MetadataMAC)
	if err != nil {
		return nil, fileKeyStorePayload{}, fail()
	}
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write(aad)
	if !hmac.Equal(expectedMAC, mac.Sum(nil)) {
		return nil, fileKeyStorePayload{}, fail()
	}
	nonce, err := base64.StdEncoding.DecodeString(envelope.Nonce)
	if err != nil {
		return nil, fileKeyStorePayload{}, fail()
	}
	ciphertext, err := base64.StdEncoding.DecodeString(envelope.Ciphertext)
	if err != nil {
		return nil, fileKeyStorePayload{}, fail()
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fileKeyStorePayload{}, fail()
	}
	aead, err := cipher.NewGCM(block)
	if err != nil || len(nonce) != aead.NonceSize() {
		return nil, fileKeyStorePayload{}, fail()
	}
	plaintext, err := aead.Open(nil, nonce, ciphertext, aad)
	if err != nil {
		return nil, fileKeyStorePayload{}, fail()
	}
	defer scrubBytes(plaintext)
	var payload fileKeyStorePayload
	if err := json.Unmarshal(plaintext, &payload); err != nil || payload.Version != fileKeyStoreVersion {
		return nil, fileKeyStorePayload{}, fail()
	}
	return key, payload, nil
}

func envelopeSalt(data []byte) string {
	var envelope fileKeyStoreEnvelope
	if json.Unmarshal(data, &envelope) != nil {
		return ""
	}
	return envelope.Salt
}

func fileKeyStoreAAD(envelope fileKeyStoreEnvelope) []byte {
	return []byte(fmt.Sprintf("%d|%s|%s|%d|%d|%d|%s|%s", envelope.Version, envelope.KDF, envelope.Cipher,
		envelope.ArgonTime, envelope.ArgonMemoryKiB, envelope.ArgonThreads, envelope.Salt, envelope.Nonce))
}

func managedKeysFromPayload(payload fileKeyStorePayload) (map[string]*ManagedKey, error) {
	keys := make(map[string]*ManagedKey, len(payload.Keys))
	fail := func(err error) (map[string]*ManagedKey, error) {
		for _, key := range keys {
			scrubBytes(key.privateKey)
		}
		return nil, err
	}
	for _, persisted := range payload.Keys {
		privateKey, err := base64.StdEncoding.DecodeString(persisted.PrivateKey)
		if err != nil || len(privateKey) != ed25519.PrivateKeySize {
			return fail(fmt.Errorf("invalid persisted private key for %s", persisted.KeyID))
		}
		publicKey := privateKey[ed25519.PrivateKeySize-ed25519.PublicKeySize:]
		if persisted.Algorithm != string(HSMKeyTypeEd25519) || !strings.EqualFold(hex.EncodeToString(publicKey), persisted.PublicKey) || generateKeyID(publicKey) != persisted.KeyID {
			scrubBytes(privateKey)
			return fail(fmt.Errorf("persisted key metadata integrity check failed for %s", persisted.KeyID))
		}
		if _, exists := keys[persisted.KeyID]; exists {
			scrubBytes(privateKey)
			return fail(fmt.Errorf("duplicate persisted key ID %s", persisted.KeyID))
		}
		keys[persisted.KeyID] = &ManagedKey{KeyID: persisted.KeyID, PublicKey: persisted.PublicKey, Algorithm: persisted.Algorithm,
			CreatedAt: persisted.CreatedAt, ExpiresAt: persisted.ExpiresAt, Status: persisted.Status,
			ProviderAddress: persisted.ProviderAddress, privateKey: privateKey}
	}
	if payload.ActiveID != "" {
		key, ok := keys[payload.ActiveID]
		if !ok || key.Status != keyStatusActive {
			return fail(fmt.Errorf("active key metadata integrity check failed"))
		}
	}
	return keys, nil
}

func sortStrings(values []string) {
	for i := 1; i < len(values); i++ {
		for j := i; j > 0 && values[j] < values[j-1]; j-- {
			values[j], values[j-1] = values[j-1], values[j]
		}
	}
}
