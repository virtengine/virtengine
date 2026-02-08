package keys

import (
	"crypto/rand"
	"fmt"
	"io"
	"sync"
	"time"

	enccrypto "github.com/virtengine/virtengine/x/encryption/crypto"
)

// Scope defines the data classification for vault blobs
type Scope string

const (
	// ScopeVEID is for VEID identity documents and attestations
	ScopeVEID Scope = "veid"

	// ScopeSupport is for support ticket attachments
	ScopeSupport Scope = "support"

	// ScopeMarket is for marketplace deployment artifacts
	ScopeMarket Scope = "market"

	// ScopeAudit is for audit logs and compliance artifacts
	ScopeAudit Scope = "audit"
)

// RotationPolicy defines when and how to rotate keys
type RotationPolicy struct {
	// MaxAge is the maximum key age before rotation (e.g., 90 days)
	MaxAge time.Duration

	// MaxVersions is the maximum number of key versions to keep
	MaxVersions uint32

	// AutoRotate enables automatic scheduled rotation
	AutoRotate bool

	// RotationSchedule is a cron expression for scheduled rotations
	RotationSchedule string
}

var (
	// ErrKeyRotationInProgress indicates a rotation is already active
	ErrKeyRotationInProgress = fmt.Errorf("key rotation in progress")
)

// KeyInfo represents an encryption key with metadata
type KeyInfo struct {
	// ID is the unique key identifier
	ID string

	// Scope is the data scope this key protects
	Scope Scope

	// Version is the key version number
	Version uint32

	// PublicKey is the X25519 public key
	PublicKey [32]byte

	// PrivateKey is the X25519 private key (stored encrypted in production)
	PrivateKey [32]byte

	// CreatedAt is when the key was created
	CreatedAt time.Time

	// ActivatedAt is when the key became active
	ActivatedAt *time.Time

	// DeprecatedAt is when the key was deprecated (can still decrypt)
	DeprecatedAt *time.Time

	// RevokedAt is when the key was revoked
	RevokedAt *time.Time

	// Status is the key status
	Status KeyStatus
}

// KeyStatus represents the lifecycle state of a key
type KeyStatus string

const (
	// KeyStatusPending indicates the key is generated but not yet active
	KeyStatusPending KeyStatus = "pending"

	// KeyStatusActive indicates the key is active for encryption
	KeyStatusActive KeyStatus = "active"

	// KeyStatusDeprecated indicates the key is deprecated but can still decrypt
	KeyStatusDeprecated KeyStatus = "deprecated"

	// KeyStatusRetired indicates the key can no longer be used
	KeyStatusRetired KeyStatus = "retired"

	// KeyStatusRevoked indicates the key has been revoked
	KeyStatusRevoked KeyStatus = "revoked"
)

// KeyManager manages encryption keys for the data vault
// It implements DEK (Data Encryption Key) management with rotation support
type KeyManager struct {
	mu sync.RWMutex

	// keys stores all keys indexed by scope and key ID
	keys map[Scope]map[string]*KeyInfo

	// activeKeys tracks the currently active key for each scope
	activeKeys map[Scope]string

	// rotationPolicies defines rotation policies per scope
	rotationPolicies map[Scope]*RotationPolicy

	// rotationState tracks active rotations
	rotationState map[Scope]*KeyRotation
}

// KeyRotation tracks an active key rotation
type KeyRotation struct {
	// Scope is the scope being rotated
	Scope Scope

	// OldKeyID is the key being replaced
	OldKeyID string

	// NewKeyID is the new active key
	NewKeyID string

	// StartedAt is when rotation started
	StartedAt time.Time

	// OverlapEnd is when the old key should be deprecated
	OverlapEnd time.Time

	// Status is the rotation status
	Status RotationStatus
}

// RotationStatus represents the state of a key rotation
type RotationStatus string

const (
	// RotationStatusPending indicates rotation is scheduled
	RotationStatusPending RotationStatus = "pending"

	// RotationStatusInProgress indicates rotation is active
	RotationStatusInProgress RotationStatus = "in_progress"

	// RotationStatusCompleted indicates rotation finished
	RotationStatusCompleted RotationStatus = "completed"

	// RotationStatusFailed indicates rotation failed
	RotationStatusFailed RotationStatus = "failed"
)

// NewKeyManager creates a new key manager
func NewKeyManager() *KeyManager {
	return &KeyManager{
		keys:             make(map[Scope]map[string]*KeyInfo),
		activeKeys:       make(map[Scope]string),
		rotationPolicies: make(map[Scope]*RotationPolicy),
		rotationState:    make(map[Scope]*KeyRotation),
	}
}

// Initialize initializes keys for all scopes
func (km *KeyManager) Initialize() error {
	scopes := []Scope{ScopeVEID, ScopeSupport, ScopeMarket, ScopeAudit}

	for _, scope := range scopes {
		if err := km.GenerateKey(scope); err != nil {
			return fmt.Errorf("failed to generate key for scope %s: %w", scope, err)
		}
	}

	return nil
}

// GenerateKey generates a new encryption key for a scope
func (km *KeyManager) GenerateKey(scope Scope) error {
	km.mu.Lock()
	defer km.mu.Unlock()

	// Generate X25519 key pair
	keyPair, err := enccrypto.GenerateKeyPair()
	if err != nil {
		return fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Generate key ID
	var keyIDBytes [16]byte
	if _, err := io.ReadFull(rand.Reader, keyIDBytes[:]); err != nil {
		return fmt.Errorf("failed to generate key ID: %w", err)
	}
	keyID := fmt.Sprintf("%s-%x", scope, keyIDBytes)

	// Determine version
	scopeKeys := km.keys[scope]
	version := uint32(1)
	if scopeKeys != nil {
		if len(scopeKeys) > 0 {
			version = uint32(len(scopeKeys)) + 1 //nolint:gosec // key versions are expected to be reasonable
		}
	}

	// Create key info
	now := time.Now()
	keyInfo := &KeyInfo{
		ID:          keyID,
		Scope:       scope,
		Version:     version,
		PublicKey:   keyPair.PublicKey,
		PrivateKey:  keyPair.PrivateKey,
		CreatedAt:   now,
		ActivatedAt: &now,
		Status:      KeyStatusActive,
	}

	// Store key
	if km.keys[scope] == nil {
		km.keys[scope] = make(map[string]*KeyInfo)
	}
	km.keys[scope][keyID] = keyInfo
	km.activeKeys[scope] = keyID

	return nil
}

// GetActiveKey returns the active encryption key for a scope
func (km *KeyManager) GetActiveKey(scope Scope) (*KeyInfo, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	keyID, exists := km.activeKeys[scope]
	if !exists {
		return nil, fmt.Errorf("no active key for scope %s", scope)
	}

	keyInfo := km.keys[scope][keyID]
	if keyInfo == nil {
		return nil, fmt.Errorf("active key %s not found for scope %s", keyID, scope)
	}
	if keyInfo.Status != KeyStatusActive {
		return nil, fmt.Errorf("active key %s is not active", keyID)
	}

	// Return a copy to prevent mutation
	keyCopy := *keyInfo
	return &keyCopy, nil
}

// GetKey retrieves a key by scope and key ID
func (km *KeyManager) GetKey(scope Scope, keyID string) (*KeyInfo, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	scopeKeys := km.keys[scope]
	if scopeKeys == nil {
		return nil, fmt.Errorf("no keys for scope %s", scope)
	}

	keyInfo := scopeKeys[keyID]
	if keyInfo == nil {
		return nil, fmt.Errorf("key %s not found for scope %s", keyID, scope)
	}

	// Return a copy
	keyCopy := *keyInfo
	return &keyCopy, nil
}

// RotateKey initiates key rotation for a scope
func (km *KeyManager) RotateKey(scope Scope, overlapDuration time.Duration) error {
	km.mu.Lock()
	defer km.mu.Unlock()

	// Check if rotation already in progress
	if rotation, exists := km.rotationState[scope]; exists {
		if rotation.Status == RotationStatusInProgress {
			return ErrKeyRotationInProgress
		}
	}

	// Get current active key
	oldKeyID, exists := km.activeKeys[scope]
	if !exists {
		return fmt.Errorf("no active key to rotate for scope %s", scope)
	}

	// Generate new key pair
	keyPair, err := enccrypto.GenerateKeyPair()
	if err != nil {
		return fmt.Errorf("failed to generate new key pair: %w", err)
	}

	// Generate new key ID
	var keyIDBytes [16]byte
	if _, err := io.ReadFull(rand.Reader, keyIDBytes[:]); err != nil {
		return fmt.Errorf("failed to generate key ID: %w", err)
	}
	newKeyID := fmt.Sprintf("%s-%x", scope, keyIDBytes)

	// Determine new version
	scopeKeys := km.keys[scope]
	newVersion := uint32(len(scopeKeys)) + 1 //nolint:gosec // key versions are expected to be reasonable

	// Create new key info
	now := time.Now()
	newKeyInfo := &KeyInfo{
		ID:          newKeyID,
		Scope:       scope,
		Version:     newVersion,
		PublicKey:   keyPair.PublicKey,
		PrivateKey:  keyPair.PrivateKey,
		CreatedAt:   now,
		ActivatedAt: &now,
		Status:      KeyStatusActive,
	}

	// Store new key
	km.keys[scope][newKeyID] = newKeyInfo

	// Update active key
	km.activeKeys[scope] = newKeyID

	// Deprecate old key (but keep for decryption during overlap)
	oldKeyInfo := km.keys[scope][oldKeyID]
	deprecatedAt := now
	oldKeyInfo.DeprecatedAt = &deprecatedAt
	oldKeyInfo.Status = KeyStatusDeprecated

	// Track rotation
	overlapEnd := now.Add(overlapDuration)
	km.rotationState[scope] = &KeyRotation{
		Scope:      scope,
		OldKeyID:   oldKeyID,
		NewKeyID:   newKeyID,
		StartedAt:  now,
		OverlapEnd: overlapEnd,
		Status:     RotationStatusInProgress,
	}

	return nil
}

// CompleteRotation completes a key rotation and retires the old key
func (km *KeyManager) CompleteRotation(scope Scope) error {
	km.mu.Lock()
	defer km.mu.Unlock()

	rotation, exists := km.rotationState[scope]
	if !exists {
		return fmt.Errorf("no rotation in progress for scope %s", scope)
	}

	if rotation.Status != RotationStatusInProgress {
		return fmt.Errorf("rotation not in progress for scope %s", scope)
	}

	// Mark old key as retired
	oldKeyInfo := km.keys[scope][rotation.OldKeyID]
	if oldKeyInfo != nil {
		oldKeyInfo.Status = KeyStatusRetired
	}

	// Mark rotation as completed
	rotation.Status = RotationStatusCompleted

	return nil
}

// RevokeKey revokes a key and removes it from active use.
func (km *KeyManager) RevokeKey(scope Scope, keyID string) error {
	km.mu.Lock()
	defer km.mu.Unlock()

	scopeKeys := km.keys[scope]
	if scopeKeys == nil {
		return fmt.Errorf("no keys for scope %s", scope)
	}
	keyInfo := scopeKeys[keyID]
	if keyInfo == nil {
		return fmt.Errorf("key %s not found for scope %s", keyID, scope)
	}
	if keyInfo.Status == KeyStatusRevoked {
		return fmt.Errorf("key %s already revoked", keyID)
	}

	now := time.Now()
	keyInfo.RevokedAt = &now
	keyInfo.Status = KeyStatusRevoked

	if km.activeKeys[scope] == keyID {
		delete(km.activeKeys, scope)
	}

	return nil
}

// EmergencyRotateKey revokes the active key and immediately promotes a new key.
func (km *KeyManager) EmergencyRotateKey(scope Scope) (*KeyInfo, error) {
	km.mu.Lock()
	defer km.mu.Unlock()

	oldKeyID, exists := km.activeKeys[scope]
	if !exists {
		return nil, fmt.Errorf("no active key to rotate for scope %s", scope)
	}

	// Revoke old key
	oldKeyInfo := km.keys[scope][oldKeyID]
	if oldKeyInfo != nil {
		now := time.Now()
		oldKeyInfo.RevokedAt = &now
		oldKeyInfo.Status = KeyStatusRevoked
	}

	// Generate new key pair
	keyPair, err := enccrypto.GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate new key pair: %w", err)
	}

	var keyIDBytes [16]byte
	if _, err := io.ReadFull(rand.Reader, keyIDBytes[:]); err != nil {
		return nil, fmt.Errorf("failed to generate key ID: %w", err)
	}
	newKeyID := fmt.Sprintf("%s-%x", scope, keyIDBytes)
	scopeKeys := km.keys[scope]
	newVersion := uint32(len(scopeKeys)) + 1 //nolint:gosec // key versions are expected to be reasonable

	now := time.Now()
	newKeyInfo := &KeyInfo{
		ID:          newKeyID,
		Scope:       scope,
		Version:     newVersion,
		PublicKey:   keyPair.PublicKey,
		PrivateKey:  keyPair.PrivateKey,
		CreatedAt:   now,
		ActivatedAt: &now,
		Status:      KeyStatusActive,
	}

	km.keys[scope][newKeyID] = newKeyInfo
	km.activeKeys[scope] = newKeyID

	return &KeyInfo{
		ID:          newKeyInfo.ID,
		Scope:       newKeyInfo.Scope,
		Version:     newKeyInfo.Version,
		PublicKey:   newKeyInfo.PublicKey,
		PrivateKey:  newKeyInfo.PrivateKey,
		CreatedAt:   newKeyInfo.CreatedAt,
		ActivatedAt: newKeyInfo.ActivatedAt,
		Status:      newKeyInfo.Status,
	}, nil
}

// GetRotationStatus returns the current rotation status for a scope
func (km *KeyManager) GetRotationStatus(scope Scope) (*KeyRotation, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	rotation, exists := km.rotationState[scope]
	if !exists {
		return nil, fmt.Errorf("no rotation found for scope %s", scope)
	}

	// Return a copy
	rotationCopy := *rotation
	return &rotationCopy, nil
}

// SetRotationPolicy sets the rotation policy for a scope
func (km *KeyManager) SetRotationPolicy(scope Scope, policy *RotationPolicy) {
	km.mu.Lock()
	defer km.mu.Unlock()

	km.rotationPolicies[scope] = policy
}

// GetRotationPolicy returns the rotation policy for a scope
func (km *KeyManager) GetRotationPolicy(scope Scope) *RotationPolicy {
	km.mu.RLock()
	defer km.mu.RUnlock()

	return km.rotationPolicies[scope]
}

// ListKeys lists all keys for a scope
func (km *KeyManager) ListKeys(scope Scope) ([]*KeyInfo, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	scopeKeys := km.keys[scope]
	if scopeKeys == nil {
		return nil, nil
	}

	keys := make([]*KeyInfo, 0, len(scopeKeys))
	for _, keyInfo := range scopeKeys {
		keyCopy := *keyInfo
		keys = append(keys, &keyCopy)
	}

	return keys, nil
}
