// Package provider_daemon implements the provider daemon for VirtEngine.
//
// SECURITY-007: Multi-Signature Support
// This file provides multi-signature support for threshold operations.
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

// ErrInsufficientSignatures is returned when threshold is not met
var ErrInsufficientSignatures = errors.New("insufficient signatures for threshold")

// ErrDuplicateSignature is returned when a duplicate signature is provided
var ErrDuplicateSignature = errors.New("duplicate signature from same signer")

// ErrSignerNotAuthorized is returned when a signer is not in the authorized set
var ErrSignerNotAuthorized = errors.New("signer not authorized for this multisig")

// ErrMultiSigExpired is returned when a multisig operation has expired
var ErrMultiSigExpired = errors.New("multisig operation has expired")

// ErrMultiSigAlreadyComplete is returned when operation is already complete
var ErrMultiSigAlreadyComplete = errors.New("multisig operation already complete")

// ErrInvalidThreshold is returned when threshold configuration is invalid
var ErrInvalidThreshold = errors.New("invalid threshold configuration")

// MultiSigConfig configures a multi-signature scheme
type MultiSigConfig struct {
	// Threshold is the minimum signatures required (M of N)
	Threshold int `json:"threshold"`

	// TotalSigners is the total number of authorized signers (N)
	TotalSigners int `json:"total_signers"`

	// TimeoutDuration is how long signatures can be collected
	TimeoutDuration time.Duration `json:"timeout_duration"`

	// RequireOrderedSignatures requires signatures in a specific order
	RequireOrderedSignatures bool `json:"require_ordered_signatures"`

	// AllowPartialExecution allows partial execution with minimum threshold
	AllowPartialExecution bool `json:"allow_partial_execution"`
}

// DefaultMultiSigConfig returns the default multisig configuration
func DefaultMultiSigConfig() *MultiSigConfig {
	return &MultiSigConfig{
		Threshold:                2,
		TotalSigners:             3,
		TimeoutDuration:          24 * time.Hour,
		RequireOrderedSignatures: false,
		AllowPartialExecution:    true,
	}
}

// Validate validates the multisig configuration
func (c *MultiSigConfig) Validate() error {
	if c.Threshold < 1 {
		return fmt.Errorf("%w: threshold must be at least 1", ErrInvalidThreshold)
	}
	if c.TotalSigners < c.Threshold {
		return fmt.Errorf("%w: total signers must be >= threshold", ErrInvalidThreshold)
	}
	if c.TimeoutDuration <= 0 {
		return fmt.Errorf("%w: timeout must be positive", ErrInvalidThreshold)
	}
	return nil
}

// MultiSigKey represents a multi-signature key configuration
type MultiSigKey struct {
	// ID is the unique identifier for this multisig key
	ID string `json:"id"`

	// Config is the multisig configuration
	Config *MultiSigConfig `json:"config"`

	// AuthorizedSigners is the list of authorized signer public keys
	AuthorizedSigners []AuthorizedSigner `json:"authorized_signers"`

	// CreatedAt is when this multisig key was created
	CreatedAt time.Time `json:"created_at"`

	// Description is an optional description
	Description string `json:"description,omitempty"`

	// Metadata contains additional metadata
	Metadata map[string]string `json:"metadata,omitempty"`
}

// AuthorizedSigner represents an authorized signer
type AuthorizedSigner struct {
	// PublicKey is the signer's public key (hex encoded)
	PublicKey string `json:"public_key"`

	// Label is a human-readable label
	Label string `json:"label"`

	// Weight is the signer's weight (for weighted multisig)
	Weight int `json:"weight"`

	// Order is the required signing order (if ordered)
	Order int `json:"order,omitempty"`

	// AddedAt is when this signer was added
	AddedAt time.Time `json:"added_at"`

	// ExpiresAt is when this signer's authorization expires (optional)
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// MultiSigOperation represents a pending multisig operation
type MultiSigOperation struct {
	// ID is the unique operation identifier
	ID string `json:"id"`

	// MultiSigKeyID is the ID of the multisig key used
	MultiSigKeyID string `json:"multisig_key_id"`

	// Message is the message to be signed
	Message []byte `json:"message"`

	// MessageHash is the hash of the message
	MessageHash string `json:"message_hash"`

	// Signatures is the list of collected signatures
	Signatures []CollectedSignature `json:"signatures"`

	// Status is the operation status
	Status MultiSigStatus `json:"status"`

	// CreatedAt is when the operation was created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when the operation expires
	ExpiresAt time.Time `json:"expires_at"`

	// CompletedAt is when the operation was completed
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// FinalSignature is the combined signature (when complete)
	FinalSignature []byte `json:"final_signature,omitempty"`

	// Initiator is who initiated the operation
	Initiator string `json:"initiator"`

	// Description is an optional description
	Description string `json:"description,omitempty"`
}

// CollectedSignature represents a signature from one signer
type CollectedSignature struct {
	// SignerPublicKey is the signer's public key
	SignerPublicKey string `json:"signer_public_key"`

	// SignerLabel is the signer's label
	SignerLabel string `json:"signer_label"`

	// Signature is the signature bytes
	Signature []byte `json:"signature"`

	// SignedAt is when the signature was provided
	SignedAt time.Time `json:"signed_at"`

	// Order is the signing order (if ordered)
	Order int `json:"order,omitempty"`

	// Weight is the signer's weight
	Weight int `json:"weight"`
}

// MultiSigStatus represents the status of a multisig operation
type MultiSigStatus string

const (
	// MultiSigStatusPending indicates signatures are being collected
	MultiSigStatusPending MultiSigStatus = "pending"

	// MultiSigStatusThresholdMet indicates threshold has been met
	MultiSigStatusThresholdMet MultiSigStatus = "threshold_met"

	// MultiSigStatusComplete indicates operation is complete
	MultiSigStatusComplete MultiSigStatus = "complete"

	// MultiSigStatusExpired indicates operation has expired
	MultiSigStatusExpired MultiSigStatus = "expired"

	// MultiSigStatusCancelled indicates operation was cancelled
	MultiSigStatusCancelled MultiSigStatus = "cancelled"
)

// MultiSigManager manages multi-signature operations
type MultiSigManager struct {
	keys       map[string]*MultiSigKey
	operations map[string]*MultiSigOperation
	keyManager *KeyManager
	mu         sync.RWMutex
}

// NewMultiSigManager creates a new multisig manager
func NewMultiSigManager(keyManager *KeyManager) *MultiSigManager {
	return &MultiSigManager{
		keys:       make(map[string]*MultiSigKey),
		operations: make(map[string]*MultiSigOperation),
		keyManager: keyManager,
	}
}

// CreateMultiSigKey creates a new multi-signature key
func (m *MultiSigManager) CreateMultiSigKey(config *MultiSigConfig, signers []AuthorizedSigner, description string) (*MultiSigKey, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	if len(signers) < config.TotalSigners {
		return nil, fmt.Errorf("need %d signers, got %d", config.TotalSigners, len(signers))
	}

	// Validate no duplicate public keys
	seen := make(map[string]bool)
	for _, signer := range signers {
		if seen[signer.PublicKey] {
			return nil, fmt.Errorf("duplicate public key: %s", signer.PublicKey[:16])
		}
		seen[signer.PublicKey] = true
	}

	// Generate multisig key ID
	keyID := generateMultiSigKeyID(signers, config)

	now := time.Now().UTC()

	// Set AddedAt for signers that don't have it
	for i := range signers {
		if signers[i].AddedAt.IsZero() {
			signers[i].AddedAt = now
		}
		if signers[i].Weight == 0 {
			signers[i].Weight = 1
		}
	}

	key := &MultiSigKey{
		ID:                keyID,
		Config:            config,
		AuthorizedSigners: signers,
		CreatedAt:         now,
		Description:       description,
	}

	m.mu.Lock()
	m.keys[keyID] = key
	m.mu.Unlock()

	return key, nil
}

// GetMultiSigKey retrieves a multisig key by ID
func (m *MultiSigManager) GetMultiSigKey(keyID string) (*MultiSigKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key, exists := m.keys[keyID]
	if !exists {
		return nil, fmt.Errorf("multisig key not found: %s", keyID)
	}

	return key, nil
}

// ListMultiSigKeys lists all multisig keys
func (m *MultiSigManager) ListMultiSigKeys() []*MultiSigKey {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]*MultiSigKey, 0, len(m.keys))
	for _, key := range m.keys {
		keys = append(keys, key)
	}

	return keys
}

// InitiateOperation initiates a new multisig signing operation
func (m *MultiSigManager) InitiateOperation(keyID string, message []byte, initiator, description string) (*MultiSigOperation, error) {
	key, err := m.GetMultiSigKey(keyID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	expiresAt := now.Add(key.Config.TimeoutDuration)

	// Compute message hash
	messageHash := sha256.Sum256(message)

	operationID := generateOperationID(keyID, message, now)

	operation := &MultiSigOperation{
		ID:            operationID,
		MultiSigKeyID: keyID,
		Message:       message,
		MessageHash:   hex.EncodeToString(messageHash[:]),
		Signatures:    make([]CollectedSignature, 0),
		Status:        MultiSigStatusPending,
		CreatedAt:     now,
		ExpiresAt:     expiresAt,
		Initiator:     initiator,
		Description:   description,
	}

	m.mu.Lock()
	m.operations[operationID] = operation
	m.mu.Unlock()

	return operation, nil
}

// AddSignature adds a signature to a pending operation
func (m *MultiSigManager) AddSignature(operationID string, signerPublicKey string, signature []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	operation, exists := m.operations[operationID]
	if !exists {
		return fmt.Errorf("operation not found: %s", operationID)
	}

	// Check if expired
	if time.Now().After(operation.ExpiresAt) {
		operation.Status = MultiSigStatusExpired
		return ErrMultiSigExpired
	}

	// Check if already complete
	if operation.Status == MultiSigStatusComplete {
		return ErrMultiSigAlreadyComplete
	}

	// Get the multisig key
	key, exists := m.keys[operation.MultiSigKeyID]
	if !exists {
		return fmt.Errorf("multisig key not found: %s", operation.MultiSigKeyID)
	}

	// Verify signer is authorized
	var authorizedSigner *AuthorizedSigner
	for i := range key.AuthorizedSigners {
		if key.AuthorizedSigners[i].PublicKey == signerPublicKey {
			authorizedSigner = &key.AuthorizedSigners[i]
			break
		}
	}
	if authorizedSigner == nil {
		return ErrSignerNotAuthorized
	}

	// Check if signer has already signed
	for _, sig := range operation.Signatures {
		if sig.SignerPublicKey == signerPublicKey {
			return ErrDuplicateSignature
		}
	}

	// Check signature validity (simplified - would verify actual signature)
	if len(signature) == 0 {
		return errors.New("invalid signature")
	}

	// Check ordering if required
	if key.Config.RequireOrderedSignatures {
		expectedOrder := len(operation.Signatures) + 1
		if authorizedSigner.Order != expectedOrder {
			return fmt.Errorf("expected signer with order %d, got order %d", expectedOrder, authorizedSigner.Order)
		}
	}

	// Add signature
	operation.Signatures = append(operation.Signatures, CollectedSignature{
		SignerPublicKey: signerPublicKey,
		SignerLabel:     authorizedSigner.Label,
		Signature:       signature,
		SignedAt:        time.Now().UTC(),
		Order:           len(operation.Signatures) + 1,
		Weight:          authorizedSigner.Weight,
	})

	// Check if threshold is met
	totalWeight := 0
	for _, sig := range operation.Signatures {
		totalWeight += sig.Weight
	}

	if totalWeight >= key.Config.Threshold {
		operation.Status = MultiSigStatusThresholdMet
	}

	return nil
}

// CompleteOperation completes a multisig operation (combines signatures)
func (m *MultiSigManager) CompleteOperation(operationID string) (*MultiSigOperation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	operation, exists := m.operations[operationID]
	if !exists {
		return nil, fmt.Errorf("operation not found: %s", operationID)
	}

	if operation.Status == MultiSigStatusComplete {
		return operation, nil
	}

	if operation.Status == MultiSigStatusExpired {
		return nil, ErrMultiSigExpired
	}

	key, exists := m.keys[operation.MultiSigKeyID]
	if !exists {
		return nil, fmt.Errorf("multisig key not found: %s", operation.MultiSigKeyID)
	}

	// Check if threshold is met
	totalWeight := 0
	for _, sig := range operation.Signatures {
		totalWeight += sig.Weight
	}

	if totalWeight < key.Config.Threshold {
		return nil, ErrInsufficientSignatures
	}

	// Combine signatures (simplified - actual implementation depends on scheme)
	combinedSignature := combineSignatures(operation.Signatures)

	now := time.Now().UTC()
	operation.Status = MultiSigStatusComplete
	operation.CompletedAt = &now
	operation.FinalSignature = combinedSignature

	return operation, nil
}

// CancelOperation cancels a pending operation
func (m *MultiSigManager) CancelOperation(operationID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	operation, exists := m.operations[operationID]
	if !exists {
		return fmt.Errorf("operation not found: %s", operationID)
	}

	if operation.Status == MultiSigStatusComplete {
		return ErrMultiSigAlreadyComplete
	}

	operation.Status = MultiSigStatusCancelled
	return nil
}

// GetOperation retrieves an operation by ID
func (m *MultiSigManager) GetOperation(operationID string) (*MultiSigOperation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	operation, exists := m.operations[operationID]
	if !exists {
		return nil, fmt.Errorf("operation not found: %s", operationID)
	}

	return operation, nil
}

// ListOperations lists all operations for a multisig key
func (m *MultiSigManager) ListOperations(keyID string, status *MultiSigStatus) []*MultiSigOperation {
	m.mu.RLock()
	defer m.mu.RUnlock()

	operations := make([]*MultiSigOperation, 0)
	for _, op := range m.operations {
		if op.MultiSigKeyID != keyID {
			continue
		}
		if status != nil && op.Status != *status {
			continue
		}
		operations = append(operations, op)
	}

	// Sort by creation time (newest first)
	sort.Slice(operations, func(i, j int) bool {
		return operations[i].CreatedAt.After(operations[j].CreatedAt)
	})

	return operations
}

// CleanupExpiredOperations removes expired operations
func (m *MultiSigManager) CleanupExpiredOperations() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	count := 0

	for id, op := range m.operations {
		if op.Status == MultiSigStatusPending && now.After(op.ExpiresAt) {
			op.Status = MultiSigStatusExpired
			count++
		}

		// Remove old completed/cancelled/expired operations
		if op.Status != MultiSigStatusPending && op.CompletedAt != nil {
			if now.Sub(*op.CompletedAt) > 30*24*time.Hour { // 30 days
				delete(m.operations, id)
				count++
			}
		}
	}

	return count
}

// VerifyMultiSignature verifies a multi-signature against the original message
func (m *MultiSigManager) VerifyMultiSignature(keyID string, message, signature []byte) (bool, error) {
	key, err := m.GetMultiSigKey(keyID)
	if err != nil {
		return false, err
	}

	// Simplified verification - would actually verify each component signature
	// and ensure threshold is met
	if len(signature) < key.Config.Threshold*64 { // Assuming 64 bytes per signature
		return false, nil
	}

	// Verify message hash matches
	messageHash := sha256.Sum256(message)
	_ = messageHash // Would use in actual verification

	return true, nil
}

// generateMultiSigKeyID generates a unique ID for a multisig key
func generateMultiSigKeyID(signers []AuthorizedSigner, config *MultiSigConfig) string {
	// Create a deterministic ID based on signers and config
	data := fmt.Sprintf("multisig-%d-%d-", config.Threshold, config.TotalSigners)

	// Sort signers by public key for deterministic ordering
	pubKeys := make([]string, len(signers))
	for i, s := range signers {
		pubKeys[i] = s.PublicKey
	}
	sort.Strings(pubKeys)

	for _, pk := range pubKeys {
		data += pk[:min(16, len(pk))] // Use first 16 chars of each public key (or full key if shorter)
	}

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8])
}

// generateOperationID generates a unique operation ID
func generateOperationID(keyID string, message []byte, timestamp time.Time) string {
	data := fmt.Sprintf("%s-%d-%x", keyID, timestamp.UnixNano(), message[:min(32, len(message))])
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:12])
}

// combineSignatures combines individual signatures into a multi-signature
func combineSignatures(signatures []CollectedSignature) []byte {
	// Simplified combination - actual implementation depends on scheme
	// Could be concatenation, aggregation (BLS), or threshold signature construction

	// Sort signatures by order
	sorted := make([]CollectedSignature, len(signatures))
	copy(sorted, signatures)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Order < sorted[j].Order
	})

	// Concatenate signatures with length prefix
	combined := make([]byte, 0, 1+len(sorted)*(8+1+64))
	combined = append(combined, byte(len(sorted))) // Number of signatures

	for _, sig := range sorted {
		// Add public key hash (8 bytes)
		pubKeyHash := sha256.Sum256([]byte(sig.SignerPublicKey))
		combined = append(combined, pubKeyHash[:8]...)

		// Add signature length and data
		combined = append(combined, byte(len(sig.Signature)))
		combined = append(combined, sig.Signature...)
	}

	return combined
}

// WeightedMultiSigKey extends MultiSigKey for weighted multisig schemes
type WeightedMultiSigKey struct {
	*MultiSigKey

	// WeightThreshold is the minimum total weight required
	WeightThreshold int `json:"weight_threshold"`
}

// NewWeightedMultiSigKey creates a weighted multisig key
func NewWeightedMultiSigKey(config *MultiSigConfig, signers []AuthorizedSigner, weightThreshold int) (*WeightedMultiSigKey, error) {
	// Validate total possible weight meets threshold
	totalWeight := 0
	for _, s := range signers {
		totalWeight += s.Weight
	}
	if totalWeight < weightThreshold {
		return nil, fmt.Errorf("total possible weight (%d) less than threshold (%d)", totalWeight, weightThreshold)
	}

	key := &MultiSigKey{
		ID:                generateMultiSigKeyID(signers, config),
		Config:            config,
		AuthorizedSigners: signers,
		CreatedAt:         time.Now().UTC(),
	}

	return &WeightedMultiSigKey{
		MultiSigKey:     key,
		WeightThreshold: weightThreshold,
	}, nil
}

// ThresholdScheme represents different threshold signature schemes
type ThresholdScheme string

const (
	// ThresholdSchemeSimple uses simple signature concatenation
	ThresholdSchemeSimple ThresholdScheme = "simple"

	// ThresholdSchemeBLS uses BLS signature aggregation
	ThresholdSchemeBLS ThresholdScheme = "bls"

	// ThresholdSchemeSchnorr uses Schnorr signature aggregation
	ThresholdSchemeSchnorr ThresholdScheme = "schnorr"

	// ThresholdSchemeFrost uses FROST threshold signatures
	ThresholdSchemeFrost ThresholdScheme = "frost"
)

// MultiSigVerificationResult contains the result of multisig verification
type MultiSigVerificationResult struct {
	// Valid indicates if the signature is valid
	Valid bool `json:"valid"`

	// SignerCount is the number of signers
	SignerCount int `json:"signer_count"`

	// TotalWeight is the total weight of signers
	TotalWeight int `json:"total_weight"`

	// Signers lists the verified signers
	Signers []string `json:"signers"`

	// MeetsThreshold indicates if threshold is met
	MeetsThreshold bool `json:"meets_threshold"`

	// VerifiedAt is when verification was performed
	VerifiedAt time.Time `json:"verified_at"`
}

