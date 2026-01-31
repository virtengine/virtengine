// Package provider_daemon implements the provider daemon for VirtEngine.
//
// SECURITY-007: Key Lifecycle Management
// This file provides comprehensive key lifecycle management.
package provider_daemon

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

// ErrKeyLifecycleInvalidTransition is returned for invalid state transitions
var ErrKeyLifecycleInvalidTransition = errors.New("invalid key lifecycle state transition")

// ErrKeyLifecycleNotFound is returned when a key is not found
var ErrKeyLifecycleNotFound = errors.New("key not found in lifecycle manager")

// KeyLifecycleState represents the state of a key in its lifecycle
type KeyLifecycleState string

const (
	// KeyStateCreated indicates key has been created but not activated
	KeyStateCreated KeyLifecycleState = "created"

	// KeyStatePending indicates key is pending activation
	KeyStatePending KeyLifecycleState = "pending"

	// KeyStateActive indicates key is active and can be used
	KeyStateActive KeyLifecycleState = "active"

	// KeyStateRotating indicates key is being rotated
	KeyStateRotating KeyLifecycleState = "rotating"

	// KeyStateSuspended indicates key is temporarily suspended
	KeyStateSuspended KeyLifecycleState = "suspended"

	// KeyStateDeactivated indicates key is deactivated but not destroyed
	KeyStateDeactivated KeyLifecycleState = "deactivated"

	// KeyStateCompromised indicates key has been compromised
	KeyStateCompromised KeyLifecycleState = "compromised"

	// KeyStateExpired indicates key has expired
	KeyStateExpired KeyLifecycleState = "expired"

	// KeyStateArchived indicates key is archived (kept for verification only)
	KeyStateArchived KeyLifecycleState = "archived"

	// KeyStateDestroyed indicates key has been destroyed
	KeyStateDestroyed KeyLifecycleState = "destroyed"
)

// ValidTransitions defines allowed state transitions
var ValidTransitions = map[KeyLifecycleState][]KeyLifecycleState{
	KeyStateCreated:     {KeyStatePending, KeyStateActive, KeyStateDestroyed},
	KeyStatePending:     {KeyStateActive, KeyStateDestroyed},
	KeyStateActive:      {KeyStateRotating, KeyStateSuspended, KeyStateDeactivated, KeyStateCompromised, KeyStateExpired},
	KeyStateRotating:    {KeyStateActive, KeyStateDeactivated},
	KeyStateSuspended:   {KeyStateActive, KeyStateDeactivated, KeyStateCompromised},
	KeyStateDeactivated: {KeyStateArchived, KeyStateDestroyed},
	KeyStateCompromised: {KeyStateDestroyed},
	KeyStateExpired:     {KeyStateArchived, KeyStateDestroyed},
	KeyStateArchived:    {KeyStateDestroyed},
	KeyStateDestroyed:   {}, // Terminal state
}

// KeyLifecyclePolicy defines the lifecycle policy for keys
type KeyLifecyclePolicy struct {
	// Name is the policy name
	Name string `json:"name"`

	// Description is the policy description
	Description string `json:"description,omitempty"`

	// MaxActiveAgeDays is the maximum age before rotation is required
	MaxActiveAgeDays int `json:"max_active_age_days"`

	// RotationGracePeriodDays is the grace period after rotation
	RotationGracePeriodDays int `json:"rotation_grace_period_days"`

	// ExpirationDays is when the key expires
	ExpirationDays int `json:"expiration_days"`

	// ArchiveAfterDeactivationDays is when to archive after deactivation
	ArchiveAfterDeactivationDays int `json:"archive_after_deactivation_days"`

	// DestroyAfterArchiveDays is when to destroy after archiving
	DestroyAfterArchiveDays int `json:"destroy_after_archive_days"`

	// RequireApprovalForActivation requires approval for key activation
	RequireApprovalForActivation bool `json:"require_approval_for_activation"`

	// RequireApprovalForDestruction requires approval for destruction
	RequireApprovalForDestruction bool `json:"require_approval_for_destruction"`

	// AutoRotate automatically rotates keys
	AutoRotate bool `json:"auto_rotate"`

	// NotifyBeforeExpirationDays days before expiration to notify
	NotifyBeforeExpirationDays int `json:"notify_before_expiration_days"`

	// AllowedKeyTypes specifies allowed key types for this policy
	AllowedKeyTypes []string `json:"allowed_key_types,omitempty"`

	// MinimumKeyStrength is the minimum key strength (bits)
	MinimumKeyStrength int `json:"minimum_key_strength,omitempty"`
}

// DefaultKeyLifecyclePolicy returns the default lifecycle policy
func DefaultKeyLifecyclePolicy() *KeyLifecyclePolicy {
	return &KeyLifecyclePolicy{
		Name:                          "default",
		Description:                   "Default key lifecycle policy",
		MaxActiveAgeDays:              90,
		RotationGracePeriodDays:       7,
		ExpirationDays:                365,
		ArchiveAfterDeactivationDays:  30,
		DestroyAfterArchiveDays:       365,
		RequireApprovalForActivation:  false,
		RequireApprovalForDestruction: true,
		AutoRotate:                    true,
		NotifyBeforeExpirationDays:    30,
		AllowedKeyTypes:               []string{"ed25519", "secp256k1", "p256"},
		MinimumKeyStrength:            256,
	}
}

// KeyLifecycleRecord tracks the lifecycle of a key
type KeyLifecycleRecord struct {
	// KeyID is the key identifier
	KeyID string `json:"key_id"`

	// KeyFingerprint is the public key fingerprint
	KeyFingerprint string `json:"key_fingerprint"`

	// KeyType is the type of key
	KeyType string `json:"key_type"`

	// CurrentState is the current lifecycle state
	CurrentState KeyLifecycleState `json:"current_state"`

	// PolicyName is the applied policy name
	PolicyName string `json:"policy_name"`

	// CreatedAt is when the key was created
	CreatedAt time.Time `json:"created_at"`

	// ActivatedAt is when the key was activated
	ActivatedAt *time.Time `json:"activated_at,omitempty"`

	// LastRotatedAt is when the key was last rotated
	LastRotatedAt *time.Time `json:"last_rotated_at,omitempty"`

	// DeactivatedAt is when the key was deactivated
	DeactivatedAt *time.Time `json:"deactivated_at,omitempty"`

	// ExpiresAt is when the key expires
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// DestroyedAt is when the key was destroyed
	DestroyedAt *time.Time `json:"destroyed_at,omitempty"`

	// StateHistory is the history of state transitions
	StateHistory []StateTransition `json:"state_history"`

	// Metadata contains additional metadata
	Metadata map[string]string `json:"metadata,omitempty"`

	// SuccessorKeyID is the ID of the key that replaced this one
	SuccessorKeyID string `json:"successor_key_id,omitempty"`

	// PredecessorKeyID is the ID of the key this replaced
	PredecessorKeyID string `json:"predecessor_key_id,omitempty"`
}

// StateTransition represents a state transition
type StateTransition struct {
	// FromState is the previous state
	FromState KeyLifecycleState `json:"from_state"`

	// ToState is the new state
	ToState KeyLifecycleState `json:"to_state"`

	// TransitionedAt is when the transition occurred
	TransitionedAt time.Time `json:"transitioned_at"`

	// TransitionedBy is who triggered the transition
	TransitionedBy string `json:"transitioned_by"`

	// Reason is the reason for the transition
	Reason string `json:"reason,omitempty"`

	// Approved indicates if the transition was approved
	Approved bool `json:"approved"`

	// ApprovedBy is who approved the transition
	ApprovedBy string `json:"approved_by,omitempty"`
}

// KeyLifecycleManager manages key lifecycles
type KeyLifecycleManager struct {
	policies   map[string]*KeyLifecyclePolicy
	records    map[string]*KeyLifecycleRecord
	keyManager *KeyManager
	mu         sync.RWMutex
}

// NewKeyLifecycleManager creates a new lifecycle manager
func NewKeyLifecycleManager(keyManager *KeyManager) *KeyLifecycleManager {
	lm := &KeyLifecycleManager{
		policies:   make(map[string]*KeyLifecyclePolicy),
		records:    make(map[string]*KeyLifecycleRecord),
		keyManager: keyManager,
	}

	// Register default policy
	lm.policies["default"] = DefaultKeyLifecyclePolicy()

	return lm
}

// RegisterPolicy registers a lifecycle policy
func (lm *KeyLifecycleManager) RegisterPolicy(policy *KeyLifecyclePolicy) error {
	if policy.Name == "" {
		return errors.New("policy name cannot be empty")
	}

	lm.mu.Lock()
	defer lm.mu.Unlock()

	lm.policies[policy.Name] = policy
	return nil
}

// GetPolicy retrieves a policy by name
func (lm *KeyLifecycleManager) GetPolicy(name string) (*KeyLifecyclePolicy, error) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	policy, exists := lm.policies[name]
	if !exists {
		return nil, fmt.Errorf("policy not found: %s", name)
	}

	return policy, nil
}

// RegisterKey registers a key with the lifecycle manager
func (lm *KeyLifecycleManager) RegisterKey(keyID, keyType, keyFingerprint, policyName string) (*KeyLifecycleRecord, error) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if _, exists := lm.records[keyID]; exists {
		return nil, fmt.Errorf("key already registered: %s", keyID)
	}

	policy, exists := lm.policies[policyName]
	if !exists {
		return nil, fmt.Errorf("policy not found: %s", policyName)
	}

	// Validate key type against policy
	if len(policy.AllowedKeyTypes) > 0 {
		allowed := false
		for _, t := range policy.AllowedKeyTypes {
			if t == keyType {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, fmt.Errorf("key type %s not allowed by policy %s", keyType, policyName)
		}
	}

	now := time.Now().UTC()
	expiresAt := now.AddDate(0, 0, policy.ExpirationDays)

	record := &KeyLifecycleRecord{
		KeyID:          keyID,
		KeyFingerprint: keyFingerprint,
		KeyType:        keyType,
		CurrentState:   KeyStateCreated,
		PolicyName:     policyName,
		CreatedAt:      now,
		ExpiresAt:      &expiresAt,
		StateHistory:   make([]StateTransition, 0),
		Metadata:       make(map[string]string),
	}

	// Add initial state
	record.StateHistory = append(record.StateHistory, StateTransition{
		FromState:      "",
		ToState:        KeyStateCreated,
		TransitionedAt: now,
		TransitionedBy: "lifecycle_manager",
		Reason:         "Key created",
		Approved:       true,
	})

	lm.records[keyID] = record
	return record, nil
}

// TransitionState transitions a key to a new state
func (lm *KeyLifecycleManager) TransitionState(keyID string, newState KeyLifecycleState, transitionedBy, reason string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	record, exists := lm.records[keyID]
	if !exists {
		return ErrKeyLifecycleNotFound
	}

	// Validate transition
	if !isValidTransition(record.CurrentState, newState) {
		return fmt.Errorf("%w: %s -> %s", ErrKeyLifecycleInvalidTransition, record.CurrentState, newState)
	}

	now := time.Now().UTC()

	// Record transition
	transition := StateTransition{
		FromState:      record.CurrentState,
		ToState:        newState,
		TransitionedAt: now,
		TransitionedBy: transitionedBy,
		Reason:         reason,
		Approved:       true,
	}

	record.StateHistory = append(record.StateHistory, transition)
	record.CurrentState = newState

	// Update timestamps based on state
	switch newState {
	case KeyStateActive:
		record.ActivatedAt = &now
	case KeyStateDeactivated:
		record.DeactivatedAt = &now
	case KeyStateDestroyed:
		record.DestroyedAt = &now
		// Actually destroy the key if we have a key manager
		if lm.keyManager != nil {
			lm.keyManager.RevokeKey(keyID)
		}
	}

	return nil
}

// ActivateKey activates a created key
func (lm *KeyLifecycleManager) ActivateKey(keyID, activatedBy string) error {
	return lm.TransitionState(keyID, KeyStateActive, activatedBy, "Key activation")
}

// SuspendKey suspends an active key
func (lm *KeyLifecycleManager) SuspendKey(keyID, suspendedBy, reason string) error {
	return lm.TransitionState(keyID, KeyStateSuspended, suspendedBy, reason)
}

// ReactivateKey reactivates a suspended key
func (lm *KeyLifecycleManager) ReactivateKey(keyID, reactivatedBy string) error {
	return lm.TransitionState(keyID, KeyStateActive, reactivatedBy, "Key reactivation")
}

// RotateKey initiates key rotation
func (lm *KeyLifecycleManager) RotateKey(keyID, successorKeyID, rotatedBy string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	record, exists := lm.records[keyID]
	if !exists {
		return ErrKeyLifecycleNotFound
	}

	now := time.Now().UTC()

	// Set rotating state
	record.CurrentState = KeyStateRotating
	record.LastRotatedAt = &now
	record.SuccessorKeyID = successorKeyID

	record.StateHistory = append(record.StateHistory, StateTransition{
		FromState:      KeyStateActive,
		ToState:        KeyStateRotating,
		TransitionedAt: now,
		TransitionedBy: rotatedBy,
		Reason:         fmt.Sprintf("Rotating to successor key: %s", successorKeyID),
		Approved:       true,
	})

	// Update successor's predecessor reference
	if successorRecord, exists := lm.records[successorKeyID]; exists {
		successorRecord.PredecessorKeyID = keyID
	}

	return nil
}

// CompleteRotation completes key rotation (old key becomes deactivated)
func (lm *KeyLifecycleManager) CompleteRotation(keyID, completedBy string) error {
	return lm.TransitionState(keyID, KeyStateDeactivated, completedBy, "Rotation complete")
}

// MarkCompromised marks a key as compromised
func (lm *KeyLifecycleManager) MarkCompromised(keyID, reportedBy, reason string) error {
	return lm.TransitionState(keyID, KeyStateCompromised, reportedBy, reason)
}

// ArchiveKey archives a deactivated or expired key
func (lm *KeyLifecycleManager) ArchiveKey(keyID, archivedBy string) error {
	return lm.TransitionState(keyID, KeyStateArchived, archivedBy, "Key archived")
}

// DestroyKey destroys a key
func (lm *KeyLifecycleManager) DestroyKey(keyID, destroyedBy, reason string) error {
	return lm.TransitionState(keyID, KeyStateDestroyed, destroyedBy, reason)
}

// GetRecord retrieves a lifecycle record
func (lm *KeyLifecycleManager) GetRecord(keyID string) (*KeyLifecycleRecord, error) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	record, exists := lm.records[keyID]
	if !exists {
		return nil, ErrKeyLifecycleNotFound
	}

	return record, nil
}

// GetRecordsByState retrieves records by state
func (lm *KeyLifecycleManager) GetRecordsByState(state KeyLifecycleState) []*KeyLifecycleRecord {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	records := make([]*KeyLifecycleRecord, 0)
	for _, record := range lm.records {
		if record.CurrentState == state {
			records = append(records, record)
		}
	}

	return records
}

// GetActiveKeys returns all active keys
func (lm *KeyLifecycleManager) GetActiveKeys() []*KeyLifecycleRecord {
	return lm.GetRecordsByState(KeyStateActive)
}

// GetExpiredKeys returns all expired keys
func (lm *KeyLifecycleManager) GetExpiredKeys() []*KeyLifecycleRecord {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	now := time.Now().UTC()
	records := make([]*KeyLifecycleRecord, 0)

	for _, record := range lm.records {
		if record.ExpiresAt != nil && now.After(*record.ExpiresAt) {
			records = append(records, record)
		}
	}

	return records
}

// GetKeysNeedingRotation returns keys that need rotation
func (lm *KeyLifecycleManager) GetKeysNeedingRotation() []*KeyLifecycleRecord {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	now := time.Now().UTC()
	records := make([]*KeyLifecycleRecord, 0)

	for _, record := range lm.records {
		if record.CurrentState != KeyStateActive {
			continue
		}

		policy, exists := lm.policies[record.PolicyName]
		if !exists {
			continue
		}

		// If MaxActiveAgeDays is 0, key needs rotation immediately
		if policy.MaxActiveAgeDays == 0 {
			records = append(records, record)
			continue
		}

		rotationDue := record.CreatedAt.AddDate(0, 0, policy.MaxActiveAgeDays)
		if record.LastRotatedAt != nil {
			rotationDue = record.LastRotatedAt.AddDate(0, 0, policy.MaxActiveAgeDays)
		}

		if now.After(rotationDue) {
			records = append(records, record)
		}
	}

	return records
}

// GetKeysNearingExpiration returns keys nearing expiration
func (lm *KeyLifecycleManager) GetKeysNearingExpiration() []*KeyLifecycleRecord {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	now := time.Now().UTC()
	records := make([]*KeyLifecycleRecord, 0)

	for _, record := range lm.records {
		if record.CurrentState == KeyStateDestroyed || record.ExpiresAt == nil {
			continue
		}

		policy, exists := lm.policies[record.PolicyName]
		if !exists {
			continue
		}

		notifyDate := record.ExpiresAt.AddDate(0, 0, -policy.NotifyBeforeExpirationDays)
		if now.After(notifyDate) && now.Before(*record.ExpiresAt) {
			records = append(records, record)
		}
	}

	return records
}

// ProcessExpiredKeys handles expired keys
func (lm *KeyLifecycleManager) ProcessExpiredKeys() (int, error) {
	expiredKeys := lm.GetExpiredKeys()
	count := 0

	for _, record := range expiredKeys {
		if record.CurrentState == KeyStateActive {
			if err := lm.TransitionState(record.KeyID, KeyStateExpired, "lifecycle_manager", "Key expired"); err == nil {
				count++
			}
		}
	}

	return count, nil
}

// ProcessAutoRotation handles automatic key rotation
func (lm *KeyLifecycleManager) ProcessAutoRotation(keyGenerator func(keyType string) (string, error)) (int, error) {
	keysToRotate := lm.GetKeysNeedingRotation()
	count := 0

	for _, record := range keysToRotate {
		policy, err := lm.GetPolicy(record.PolicyName)
		if err != nil || !policy.AutoRotate {
			continue
		}

		// Generate new key
		newKeyID, err := keyGenerator(record.KeyType)
		if err != nil {
			continue
		}

		// Start rotation
		if err := lm.RotateKey(record.KeyID, newKeyID, "lifecycle_manager"); err == nil {
			count++
		}
	}

	return count, nil
}

// GenerateLifecycleReport generates a lifecycle report
func (lm *KeyLifecycleManager) GenerateLifecycleReport() *LifecycleReport {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	report := &LifecycleReport{
		GeneratedAt:  time.Now().UTC(),
		TotalKeys:    len(lm.records),
		ByState:      make(map[string]int),
		ByPolicy:     make(map[string]int),
		ByType:       make(map[string]int),
		RecentEvents: make([]StateTransition, 0),
	}

	// Count by state, policy, type
	for _, record := range lm.records {
		report.ByState[string(record.CurrentState)]++
		report.ByPolicy[record.PolicyName]++
		report.ByType[record.KeyType]++

		// Collect recent transitions
		for _, transition := range record.StateHistory {
			if time.Since(transition.TransitionedAt) <= 24*time.Hour {
				report.RecentEvents = append(report.RecentEvents, transition)
			}
		}
	}

	// Sort recent events by time
	sort.Slice(report.RecentEvents, func(i, j int) bool {
		return report.RecentEvents[i].TransitionedAt.After(report.RecentEvents[j].TransitionedAt)
	})

	// Limit to last 50 events
	if len(report.RecentEvents) > 50 {
		report.RecentEvents = report.RecentEvents[:50]
	}

	// Count important states
	report.ActiveKeys = report.ByState[string(KeyStateActive)]
	report.ExpiredKeys = report.ByState[string(KeyStateExpired)]
	report.CompromisedKeys = report.ByState[string(KeyStateCompromised)]

	return report
}

// LifecycleReport contains lifecycle statistics
type LifecycleReport struct {
	GeneratedAt     time.Time         `json:"generated_at"`
	TotalKeys       int               `json:"total_keys"`
	ActiveKeys      int               `json:"active_keys"`
	ExpiredKeys     int               `json:"expired_keys"`
	CompromisedKeys int               `json:"compromised_keys"`
	ByState         map[string]int    `json:"by_state"`
	ByPolicy        map[string]int    `json:"by_policy"`
	ByType          map[string]int    `json:"by_type"`
	RecentEvents    []StateTransition `json:"recent_events"`
}

// isValidTransition checks if a state transition is valid
func isValidTransition(from, to KeyLifecycleState) bool {
	allowedTransitions, exists := ValidTransitions[from]
	if !exists {
		return false
	}

	for _, allowed := range allowedTransitions {
		if allowed == to {
			return true
		}
	}

	return false
}

// KeyRotationSchedule defines a rotation schedule
type KeyRotationSchedule struct {
	// KeyID is the key to rotate
	KeyID string `json:"key_id"`

	// ScheduledAt is when the rotation is scheduled
	ScheduledAt time.Time `json:"scheduled_at"`

	// CreatedAt is when the schedule was created
	CreatedAt time.Time `json:"created_at"`

	// Reason is the reason for scheduled rotation
	Reason string `json:"reason"`

	// AutoExecute indicates if rotation should execute automatically
	AutoExecute bool `json:"auto_execute"`

	// Status is the schedule status
	Status string `json:"status"` // pending, completed, cancelled
}

// GenerateKeyID generates a unique key ID
func GenerateKeyID(keyType string) string {
	data := fmt.Sprintf("%s-%d", keyType, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16])
}
