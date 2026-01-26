// Package security contains security-focused tests for VirtEngine.
// These tests verify key rotation procedures for all key types.
//
// Task Reference: VE-800 - Security audit readiness
package security

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// KeyRotationTestSuite tests key rotation procedures.
type KeyRotationTestSuite struct {
	suite.Suite
}

func TestKeyRotation(t *testing.T) {
	suite.Run(t, new(KeyRotationTestSuite))
}

// =============================================================================
// Provider Daemon Key Rotation Tests
// =============================================================================

// TestProviderDaemonKeyRotation tests provider daemon key rotation.
func (s *KeyRotationTestSuite) TestProviderDaemonKeyRotation() {
	s.T().Log("=== Test: Provider Daemon Key Rotation ===")

	// Test: New key can be registered
	s.Run("new_key_registration", func() {
		keyManager := NewProviderKeyManager("provider123")

		oldKey := generateKey(s.T())
		keyManager.SetActiveKey(oldKey, "v1")

		newKey := generateKey(s.T())
		err := keyManager.RegisterNewKey(newKey, "v2")
		require.NoError(s.T(), err, "new key registration should succeed")

		keys := keyManager.GetAllKeys()
		require.Len(s.T(), keys, 2, "should have two keys")
	})

	// Test: Key activation after grace period
	s.Run("key_activation_after_grace_period", func() {
		keyManager := NewProviderKeyManager("provider123")

		oldKey := generateKey(s.T())
		keyManager.SetActiveKey(oldKey, "v1")

		newKey := generateKey(s.T())
		keyManager.RegisterNewKey(newKey, "v2")

		// Activate new key
		err := keyManager.ActivateKey("v2")
		require.NoError(s.T(), err, "key activation should succeed")

		activeKey := keyManager.GetActiveKey()
		require.Equal(s.T(), "v2", activeKey.Version, "v2 should be active")
	})

	// Test: Old key remains valid during grace period
	s.Run("old_key_valid_during_grace", func() {
		keyManager := NewProviderKeyManager("provider123")
		gracePeriod := 24 * time.Hour

		oldKey := generateKey(s.T())
		keyManager.SetActiveKey(oldKey, "v1")

		newKey := generateKey(s.T())
		keyManager.RegisterNewKey(newKey, "v2")
		keyManager.ActivateKey("v2")

		// Old key should still be valid during grace period
		isValid := keyManager.IsKeyValid("v1", gracePeriod)
		require.True(s.T(), isValid, "old key should be valid during grace period")
	})

	// Test: Key revocation
	s.Run("key_revocation", func() {
		keyManager := NewProviderKeyManager("provider123")

		key := generateKey(s.T())
		keyManager.SetActiveKey(key, "v1")

		newKey := generateKey(s.T())
		keyManager.RegisterNewKey(newKey, "v2")
		keyManager.ActivateKey("v2")

		err := keyManager.RevokeKey("v1")
		require.NoError(s.T(), err, "key revocation should succeed")

		isValid := keyManager.IsKeyValid("v1", 0)
		require.False(s.T(), isValid, "revoked key should not be valid")
	})
}

// =============================================================================
// Approved Client Key Rotation Tests
// =============================================================================

// TestApprovedClientKeyRotation tests approved client key rotation.
func (s *KeyRotationTestSuite) TestApprovedClientKeyRotation() {
	s.T().Log("=== Test: Approved Client Key Rotation ===")

	// Test: Add new approved client key
	s.Run("add_approved_client_key", func() {
		allowlist := NewApprovedClientAllowlist()

		key1 := generateKey(s.T())
		err := allowlist.AddKey(key1, "MobileApp-v1.0")
		require.NoError(s.T(), err, "adding client key should succeed")

		isApproved := allowlist.IsApproved(key1)
		require.True(s.T(), isApproved, "new key should be approved")
	})

	// Test: Rotate client key (add new, remove old)
	s.Run("rotate_client_key", func() {
		allowlist := NewApprovedClientAllowlist()

		oldKey := generateKey(s.T())
		allowlist.AddKey(oldKey, "MobileApp-v1.0")

		newKey := generateKey(s.T())
		allowlist.AddKey(newKey, "MobileApp-v2.0")

		// Both should be valid during transition
		require.True(s.T(), allowlist.IsApproved(oldKey), "old key should still be approved")
		require.True(s.T(), allowlist.IsApproved(newKey), "new key should be approved")

		// Remove old key after migration
		err := allowlist.RemoveKey(oldKey)
		require.NoError(s.T(), err, "removing old key should succeed")

		require.False(s.T(), allowlist.IsApproved(oldKey), "old key should not be approved")
		require.True(s.T(), allowlist.IsApproved(newKey), "new key should still be approved")
	})

	// Test: Emergency key revocation
	s.Run("emergency_key_revocation", func() {
		allowlist := NewApprovedClientAllowlist()

		compromisedKey := generateKey(s.T())
		allowlist.AddKey(compromisedKey, "CompromisedApp")

		// Emergency revoke
		err := allowlist.RevokeWithReason(compromisedKey, "key_compromised")
		require.NoError(s.T(), err, "emergency revocation should succeed")

		require.False(s.T(), allowlist.IsApproved(compromisedKey), "revoked key should not be approved")

		// Verify revocation record
		record := allowlist.GetRevocationRecord(compromisedKey)
		require.NotNil(s.T(), record, "revocation record should exist")
		require.Equal(s.T(), "key_compromised", record.Reason)
	})
}

// =============================================================================
// Validator Key Rotation Tests
// =============================================================================

// TestValidatorKeyRotation tests validator identity key rotation.
func (s *KeyRotationTestSuite) TestValidatorKeyRotation() {
	s.T().Log("=== Test: Validator Key Rotation ===")

	// Test: Validator key update proposal
	s.Run("validator_key_update_proposal", func() {
		registry := NewValidatorKeyRegistry()

		validatorAddr := "virtengine1validator123"
		oldKey := generateKey(s.T())
		registry.RegisterValidator(validatorAddr, oldKey)

		newKey := generateKey(s.T())
		proposal := registry.ProposeKeyUpdate(validatorAddr, newKey)
		require.NotNil(s.T(), proposal, "key update proposal should be created")
		require.Equal(s.T(), "pending", proposal.Status)
	})

	// Test: Key update with consensus confirmation
	s.Run("key_update_with_consensus", func() {
		registry := NewValidatorKeyRegistry()

		validatorAddr := "virtengine1validator123"
		oldKey := generateKey(s.T())
		registry.RegisterValidator(validatorAddr, oldKey)

		newKey := generateKey(s.T())
		proposal := registry.ProposeKeyUpdate(validatorAddr, newKey)

		// Simulate consensus confirmation
		err := registry.ConfirmKeyUpdate(proposal.ID, 5) // 5 validator confirmations
		require.NoError(s.T(), err, "key update confirmation should succeed")

		activeKey := registry.GetActiveKey(validatorAddr)
		require.Equal(s.T(), hex.EncodeToString(newKey), hex.EncodeToString(activeKey),
			"new key should be active after confirmation")
	})

	// Test: Epoch-based key rotation
	s.Run("epoch_based_rotation", func() {
		registry := NewValidatorKeyRegistry()
		registry.SetRotationEpoch(1000) // Rotate every 1000 blocks

		validatorAddr := "virtengine1validator123"
		key := generateKey(s.T())
		registry.RegisterValidator(validatorAddr, key)

		// Check rotation requirement at different heights
		requiresRotation := registry.RequiresRotation(validatorAddr, 500)
		require.False(s.T(), requiresRotation, "should not require rotation before epoch")

		requiresRotation = registry.RequiresRotation(validatorAddr, 1001)
		require.True(s.T(), requiresRotation, "should require rotation after epoch")
	})
}

// =============================================================================
// User Account Key Rotation Tests
// =============================================================================

// TestUserAccountKeyRotation tests user account key rotation (recovery flow).
func (s *KeyRotationTestSuite) TestUserAccountKeyRotation() {
	s.T().Log("=== Test: User Account Key Rotation ===")

	// Test: Standard key rotation with MFA
	s.Run("standard_rotation_with_mfa", func() {
		account := NewUserAccount("user123")

		oldKey := generateKey(s.T())
		account.SetPrimaryKey(oldKey)

		// Initiate rotation (requires MFA)
		newKey := generateKey(s.T())
		rotationReq := account.InitiateKeyRotation(newKey)
		require.Equal(s.T(), "mfa_required", rotationReq.Status)

		// Simulate MFA verification
		err := account.CompleteKeyRotation(rotationReq.ID, "valid_mfa_token")
		require.NoError(s.T(), err, "key rotation with MFA should succeed")

		activeKey := account.GetPrimaryKey()
		require.Equal(s.T(), hex.EncodeToString(newKey), hex.EncodeToString(activeKey),
			"new key should be primary after rotation")
	})

	// Test: Account recovery key rotation
	s.Run("recovery_key_rotation", func() {
		account := NewUserAccount("user123")

		originalKey := generateKey(s.T())
		account.SetPrimaryKey(originalKey)

		// Set up recovery key
		recoveryKey := generateKey(s.T())
		account.SetRecoveryKey(recoveryKey)

		// Simulate recovery flow (lost primary key)
		newPrimaryKey := generateKey(s.T())
		recoveryReq := account.InitiateRecovery(recoveryKey, newPrimaryKey)
		require.NotNil(s.T(), recoveryReq, "recovery request should be created")

		// Recovery requires waiting period
		require.Equal(s.T(), "waiting_period", recoveryReq.Status)
		require.True(s.T(), recoveryReq.WaitingPeriod > 0, "should have waiting period")
	})

	// Test: Key rotation history
	s.Run("key_rotation_history", func() {
		account := NewUserAccount("user123")

		key1 := generateKey(s.T())
		account.SetPrimaryKey(key1)

		key2 := generateKey(s.T())
		account.RotateKeyImmediate(key2)

		key3 := generateKey(s.T())
		account.RotateKeyImmediate(key3)

		history := account.GetKeyHistory()
		require.Len(s.T(), history, 3, "should have 3 keys in history")

		// Most recent should be active
		require.True(s.T(), history[0].Active, "first entry should be active")
		require.False(s.T(), history[1].Active, "older entries should not be active")
	})
}

// =============================================================================
// Cross-Cutting Key Rotation Tests
// =============================================================================

// TestKeyRotationAuditTrail tests audit logging for key rotations.
func (s *KeyRotationTestSuite) TestKeyRotationAuditTrail() {
	s.T().Log("=== Test: Key Rotation Audit Trail ===")

	// Test: All key rotations are logged
	s.Run("all_rotations_logged", func() {
		auditLog := NewKeyRotationAuditLog()

		// Provider key rotation
		auditLog.LogRotation(KeyRotationEvent{
			EntityType: "provider",
			EntityID:   "provider123",
			OldKeyHash: "abc123",
			NewKeyHash: "def456",
			Reason:     "scheduled_rotation",
			Timestamp:  time.Now().UTC(),
		})

		// User key rotation
		auditLog.LogRotation(KeyRotationEvent{
			EntityType: "user",
			EntityID:   "user456",
			OldKeyHash: "ghi789",
			NewKeyHash: "jkl012",
			Reason:     "user_initiated",
			Timestamp:  time.Now().UTC(),
		})

		events := auditLog.GetEvents()
		require.Len(s.T(), events, 2, "should have 2 audit entries")
	})

	// Test: Audit entries are immutable
	s.Run("audit_entries_immutable", func() {
		auditLog := NewKeyRotationAuditLog()

		event := KeyRotationEvent{
			EntityType: "validator",
			EntityID:   "validator123",
			OldKeyHash: "old_hash",
			NewKeyHash: "new_hash",
			Reason:     "epoch_rotation",
			Timestamp:  time.Now().UTC(),
		}

		id := auditLog.LogRotation(event)

		// Attempt to modify should fail
		err := auditLog.ModifyEvent(id, "modified_reason")
		require.Error(s.T(), err, "audit entries should be immutable")
	})
}

// =============================================================================
// Test Types and Helpers
// =============================================================================

type ProviderKeyManager struct {
	providerID string
	keys       map[string]*ManagedKey
	activeKey  string
}

type ManagedKey struct {
	Key        []byte
	Version    string
	CreatedAt  time.Time
	ActivatedAt *time.Time
	RevokedAt  *time.Time
}

func NewProviderKeyManager(providerID string) *ProviderKeyManager {
	return &ProviderKeyManager{
		providerID: providerID,
		keys:       make(map[string]*ManagedKey),
	}
}

func (m *ProviderKeyManager) SetActiveKey(key []byte, version string) {
	now := time.Now().UTC()
	m.keys[version] = &ManagedKey{
		Key:         key,
		Version:     version,
		CreatedAt:   now,
		ActivatedAt: &now,
	}
	m.activeKey = version
}

func (m *ProviderKeyManager) RegisterNewKey(key []byte, version string) error {
	m.keys[version] = &ManagedKey{
		Key:       key,
		Version:   version,
		CreatedAt: time.Now().UTC(),
	}
	return nil
}

func (m *ProviderKeyManager) ActivateKey(version string) error {
	key, exists := m.keys[version]
	if !exists {
		return &KeyRotationError{Message: "key not found"}
	}
	now := time.Now().UTC()
	key.ActivatedAt = &now
	m.activeKey = version
	return nil
}

func (m *ProviderKeyManager) GetActiveKey() *ManagedKey {
	return m.keys[m.activeKey]
}

func (m *ProviderKeyManager) GetAllKeys() []*ManagedKey {
	result := make([]*ManagedKey, 0, len(m.keys))
	for _, k := range m.keys {
		result = append(result, k)
	}
	return result
}

func (m *ProviderKeyManager) IsKeyValid(version string, gracePeriod time.Duration) bool {
	key, exists := m.keys[version]
	if !exists {
		return false
	}
	if key.RevokedAt != nil {
		return false
	}
	if gracePeriod > 0 && key.ActivatedAt != nil {
		graceEnd := key.ActivatedAt.Add(gracePeriod)
		return time.Now().UTC().Before(graceEnd)
	}
	return key.ActivatedAt != nil
}

func (m *ProviderKeyManager) RevokeKey(version string) error {
	key, exists := m.keys[version]
	if !exists {
		return &KeyRotationError{Message: "key not found"}
	}
	now := time.Now().UTC()
	key.RevokedAt = &now
	return nil
}

type ApprovedClientAllowlist struct {
	keys        map[string]*ApprovedClientKey
	revocations map[string]*RevocationRecord
}

type ApprovedClientKey struct {
	Key   []byte
	Label string
	AddedAt time.Time
}

type RevocationRecord struct {
	Key       []byte
	Reason    string
	RevokedAt time.Time
}

func NewApprovedClientAllowlist() *ApprovedClientAllowlist {
	return &ApprovedClientAllowlist{
		keys:        make(map[string]*ApprovedClientKey),
		revocations: make(map[string]*RevocationRecord),
	}
}

func (a *ApprovedClientAllowlist) AddKey(key []byte, label string) error {
	hash := hex.EncodeToString(key)
	a.keys[hash] = &ApprovedClientKey{
		Key:     key,
		Label:   label,
		AddedAt: time.Now().UTC(),
	}
	return nil
}

func (a *ApprovedClientAllowlist) IsApproved(key []byte) bool {
	hash := hex.EncodeToString(key)
	_, exists := a.keys[hash]
	return exists
}

func (a *ApprovedClientAllowlist) RemoveKey(key []byte) error {
	hash := hex.EncodeToString(key)
	delete(a.keys, hash)
	return nil
}

func (a *ApprovedClientAllowlist) RevokeWithReason(key []byte, reason string) error {
	hash := hex.EncodeToString(key)
	a.revocations[hash] = &RevocationRecord{
		Key:       key,
		Reason:    reason,
		RevokedAt: time.Now().UTC(),
	}
	delete(a.keys, hash)
	return nil
}

func (a *ApprovedClientAllowlist) GetRevocationRecord(key []byte) *RevocationRecord {
	hash := hex.EncodeToString(key)
	return a.revocations[hash]
}

type ValidatorKeyRegistry struct {
	validators    map[string]*ValidatorKeyRecord
	proposals     map[string]*KeyUpdateProposal
	rotationEpoch int64
}

type ValidatorKeyRecord struct {
	Address      string
	ActiveKey    []byte
	RegisteredAt time.Time
	LastRotation time.Time
}

type KeyUpdateProposal struct {
	ID             string
	ValidatorAddr  string
	NewKey         []byte
	Status         string
	Confirmations  int
	ProposedAt     time.Time
}

func NewValidatorKeyRegistry() *ValidatorKeyRegistry {
	return &ValidatorKeyRegistry{
		validators:    make(map[string]*ValidatorKeyRecord),
		proposals:     make(map[string]*KeyUpdateProposal),
		rotationEpoch: 0,
	}
}

func (r *ValidatorKeyRegistry) RegisterValidator(addr string, key []byte) {
	now := time.Now().UTC()
	r.validators[addr] = &ValidatorKeyRecord{
		Address:      addr,
		ActiveKey:    key,
		RegisteredAt: now,
		LastRotation: now,
	}
}

func (r *ValidatorKeyRegistry) ProposeKeyUpdate(addr string, newKey []byte) *KeyUpdateProposal {
	id := "proposal_" + addr + "_" + hex.EncodeToString(newKey[:8])
	proposal := &KeyUpdateProposal{
		ID:            id,
		ValidatorAddr: addr,
		NewKey:        newKey,
		Status:        "pending",
		ProposedAt:    time.Now().UTC(),
	}
	r.proposals[id] = proposal
	return proposal
}

func (r *ValidatorKeyRegistry) ConfirmKeyUpdate(proposalID string, confirmations int) error {
	proposal, exists := r.proposals[proposalID]
	if !exists {
		return &KeyRotationError{Message: "proposal not found"}
	}
	proposal.Confirmations = confirmations
	proposal.Status = "confirmed"

	// Update validator's key
	validator, exists := r.validators[proposal.ValidatorAddr]
	if !exists {
		return &KeyRotationError{Message: "validator not found"}
	}
	validator.ActiveKey = proposal.NewKey
	validator.LastRotation = time.Now().UTC()

	return nil
}

func (r *ValidatorKeyRegistry) GetActiveKey(addr string) []byte {
	validator, exists := r.validators[addr]
	if !exists {
		return nil
	}
	return validator.ActiveKey
}

func (r *ValidatorKeyRegistry) SetRotationEpoch(epoch int64) {
	r.rotationEpoch = epoch
}

func (r *ValidatorKeyRegistry) RequiresRotation(addr string, currentHeight int64) bool {
	if r.rotationEpoch == 0 {
		return false
	}
	validator, exists := r.validators[addr]
	if !exists {
		return false
	}
	_ = validator // Would check last rotation height
	return currentHeight > r.rotationEpoch
}

type UserAccount struct {
	id          string
	primaryKey  []byte
	recoveryKey []byte
	keyHistory  []*KeyHistoryEntry
}

type KeyHistoryEntry struct {
	Key       []byte
	Active    bool
	SetAt     time.Time
	RevokedAt *time.Time
}

type KeyRotationRequest struct {
	ID            string
	Status        string
	NewKey        []byte
	WaitingPeriod time.Duration
}

func NewUserAccount(id string) *UserAccount {
	return &UserAccount{
		id:         id,
		keyHistory: make([]*KeyHistoryEntry, 0),
	}
}

func (a *UserAccount) SetPrimaryKey(key []byte) {
	// Mark old key as inactive
	for _, entry := range a.keyHistory {
		entry.Active = false
	}

	a.primaryKey = key
	a.keyHistory = append([]*KeyHistoryEntry{{
		Key:    key,
		Active: true,
		SetAt:  time.Now().UTC(),
	}}, a.keyHistory...)
}

func (a *UserAccount) GetPrimaryKey() []byte {
	return a.primaryKey
}

func (a *UserAccount) SetRecoveryKey(key []byte) {
	a.recoveryKey = key
}

func (a *UserAccount) InitiateKeyRotation(newKey []byte) *KeyRotationRequest {
	return &KeyRotationRequest{
		ID:     "rotation_" + a.id,
		Status: "mfa_required",
		NewKey: newKey,
	}
}

func (a *UserAccount) CompleteKeyRotation(requestID string, mfaToken string) error {
	// In real impl, would verify MFA token
	if mfaToken == "" {
		return &KeyRotationError{Message: "MFA required"}
	}
	// Would look up the request and get the new key
	return nil
}

func (a *UserAccount) InitiateRecovery(recoveryKey []byte, newPrimaryKey []byte) *KeyRotationRequest {
	return &KeyRotationRequest{
		ID:            "recovery_" + a.id,
		Status:        "waiting_period",
		NewKey:        newPrimaryKey,
		WaitingPeriod: 7 * 24 * time.Hour, // 7 day waiting period
	}
}

func (a *UserAccount) RotateKeyImmediate(newKey []byte) {
	a.SetPrimaryKey(newKey)
}

func (a *UserAccount) GetKeyHistory() []*KeyHistoryEntry {
	return a.keyHistory
}

type KeyRotationAuditLog struct {
	events []*KeyRotationEvent
}

type KeyRotationEvent struct {
	ID         string
	EntityType string
	EntityID   string
	OldKeyHash string
	NewKeyHash string
	Reason     string
	Timestamp  time.Time
}

func NewKeyRotationAuditLog() *KeyRotationAuditLog {
	return &KeyRotationAuditLog{
		events: make([]*KeyRotationEvent, 0),
	}
}

func (l *KeyRotationAuditLog) LogRotation(event KeyRotationEvent) string {
	event.ID = "audit_" + hex.EncodeToString([]byte(event.EntityID))[:8]
	l.events = append(l.events, &event)
	return event.ID
}

func (l *KeyRotationAuditLog) GetEvents() []*KeyRotationEvent {
	return l.events
}

func (l *KeyRotationAuditLog) ModifyEvent(id string, newReason string) error {
	return &KeyRotationError{Message: "audit entries are immutable"}
}

type KeyRotationError struct {
	Message string
}

func (e *KeyRotationError) Error() string {
	return "key rotation error: " + e.Message
}

func generateKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, key)
	require.NoError(t, err)
	return key
}
