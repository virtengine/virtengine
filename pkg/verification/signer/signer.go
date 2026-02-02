// Package signer provides the verification attestation signing service.
package signer

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/virtengine/virtengine/pkg/verification/audit"
	"github.com/virtengine/virtengine/pkg/verification/keystorage"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// DefaultSigner implements SignerService with key rotation support.
type DefaultSigner struct {
	config   SignerConfig
	storage  keystorage.KeyStorage
	auditor  audit.AuditLogger
	logger   zerolog.Logger
	registry *veidtypes.SignerRegistryEntry

	// State
	mu              sync.RWMutex
	activeKey       *veidtypes.SignerKeyInfo
	keys            map[string]*veidtypes.SignerKeyInfo
	rotations       map[string]*veidtypes.KeyRotationRecord
	currentRotation *veidtypes.KeyRotationRecord
	sequenceCounter uint64
}

// NewDefaultSigner creates a new DefaultSigner instance.
func NewDefaultSigner(
	ctx context.Context,
	config SignerConfig,
	storage keystorage.KeyStorage,
	auditor audit.AuditLogger,
	logger zerolog.Logger,
) (*DefaultSigner, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	signer := &DefaultSigner{
		config:    config,
		storage:   storage,
		auditor:   auditor,
		logger:    logger.With().Str("component", "signer").Str("signer_id", config.SignerID).Logger(),
		keys:      make(map[string]*veidtypes.SignerKeyInfo),
		rotations: make(map[string]*veidtypes.KeyRotationRecord),
	}

	// Initialize or load existing keys
	if err := signer.initialize(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize signer: %w", err)
	}

	return signer, nil
}

// initialize sets up the signer, loading existing keys or generating new ones.
func (s *DefaultSigner) initialize(ctx context.Context) error {
	// Try to load existing keys from storage
	existingKeys, err := s.storage.ListKeys(ctx, s.config.SignerID)
	if err != nil && err != keystorage.ErrKeyNotFound {
		return fmt.Errorf("failed to list existing keys: %w", err)
	}

	if len(existingKeys) > 0 {
		// Load existing keys
		for _, keyInfo := range existingKeys {
			s.keys[keyInfo.KeyID] = keyInfo
			if keyInfo.SequenceNumber > s.sequenceCounter {
				s.sequenceCounter = keyInfo.SequenceNumber
			}
			if keyInfo.State == veidtypes.SignerKeyStateActive {
				s.activeKey = keyInfo
			}
		}

		// Create registry entry from existing data
		s.registry = &veidtypes.SignerRegistryEntry{
			SignerID:         s.config.SignerID,
			Name:             s.config.SignerName,
			ValidatorAddress: s.config.ValidatorAddress,
			ActiveKeyID:      s.activeKey.KeyID,
			KeyHistory:       make([]string, 0, len(s.keys)),
			Policy:           s.config.KeyPolicy,
			RegisteredAt:     s.activeKey.CreatedAt,
			Active:           true,
		}
		for keyID := range s.keys {
			s.registry.KeyHistory = append(s.registry.KeyHistory, keyID)
		}

		s.logger.Info().
			Int("loaded_keys", len(s.keys)).
			Str("active_key", s.activeKey.KeyID).
			Msg("loaded existing signer keys")
	} else {
		// Generate initial key
		if err := s.generateInitialKey(ctx); err != nil {
			return fmt.Errorf("failed to generate initial key: %w", err)
		}
	}

	return nil
}

// generateInitialKey generates the first signing key.
func (s *DefaultSigner) generateInitialKey(ctx context.Context) error {
	s.sequenceCounter = 1
	now := time.Now()

	// Generate key pair based on algorithm
	keyInfo, privateKey, err := s.generateKeyPair(s.config.DefaultAlgorithm, s.sequenceCounter, now)
	if err != nil {
		return err
	}

	// Activate the key
	expiresAt := now.Add(time.Duration(s.config.KeyPolicy.MaxKeyAgeSeconds) * time.Second)
	if err := keyInfo.Activate(now, expiresAt); err != nil {
		return err
	}

	// Store the key
	if err := s.storage.StoreKey(ctx, keyInfo, privateKey); err != nil {
		return fmt.Errorf("failed to store initial key: %w", err)
	}

	s.keys[keyInfo.KeyID] = keyInfo
	s.activeKey = keyInfo

	// Create registry entry
	s.registry = veidtypes.NewSignerRegistryEntry(
		s.config.SignerID,
		s.config.SignerName,
		s.config.ValidatorAddress,
		keyInfo.KeyID,
		now,
	)
	s.registry.Policy = s.config.KeyPolicy

	// Audit log
	if s.auditor != nil {
		s.auditor.Log(ctx, audit.Event{
			Type:      audit.EventTypeKeyGenerated,
			Timestamp: now,
			Actor:     s.config.SignerID,
			Resource:  keyInfo.KeyID,
			Action:    "generate_initial_key",
			Details: map[string]interface{}{
				"algorithm":       keyInfo.Algorithm,
				"fingerprint":     keyInfo.Fingerprint,
				"sequence_number": keyInfo.SequenceNumber,
				"expires_at":      expiresAt,
			},
		})
	}

	s.logger.Info().
		Str("key_id", keyInfo.KeyID).
		Str("fingerprint", keyInfo.Fingerprint[:16]+"...").
		Str("algorithm", string(keyInfo.Algorithm)).
		Msg("generated initial signing key")

	return nil
}

// generateKeyPair generates a new key pair.
func (s *DefaultSigner) generateKeyPair(
	algorithm veidtypes.AttestationProofType,
	sequenceNumber uint64,
	createdAt time.Time,
) (*veidtypes.SignerKeyInfo, []byte, error) {
	switch algorithm {
	case veidtypes.ProofTypeEd25519:
		publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, nil, ErrKeyGenerationFailed.Wrap(err.Error())
		}

		keyInfo := veidtypes.NewSignerKeyInfo(
			s.config.SignerID,
			publicKey,
			algorithm,
			sequenceNumber,
			createdAt,
		)

		return keyInfo, privateKey, nil

	default:
		return nil, nil, ErrInvalidConfig.Wrapf("unsupported algorithm: %s", algorithm)
	}
}

// SignAttestation signs a verification attestation with the active key.
func (s *DefaultSigner) SignAttestation(ctx context.Context, attestation *veidtypes.VerificationAttestation) error {
	s.mu.RLock()
	activeKey := s.activeKey
	s.mu.RUnlock()

	if activeKey == nil {
		return ErrNoActiveKey
	}

	if !activeKey.State.CanSign() {
		return ErrKeyRevoked.Wrapf("key state: %s", activeKey.State)
	}

	// Check key expiration
	if activeKey.IsExpired(time.Now()) {
		return ErrKeyExpired
	}

	// Update issuer information BEFORE getting canonical bytes
	attestation.Issuer = veidtypes.NewAttestationIssuer(activeKey.Fingerprint, s.config.ValidatorAddress)
	attestation.Issuer.KeyID = activeKey.KeyID
	attestation.Issuer.ServiceEndpoint = s.config.ServiceEndpoint

	// Get canonical bytes to sign (now includes correct issuer)
	dataToSign, err := attestation.CanonicalBytes()
	if err != nil {
		return ErrSigningFailed.Wrapf("failed to get canonical bytes: %v", err)
	}

	// Get private key from storage
	privateKey, err := s.storage.GetPrivateKey(ctx, activeKey.KeyID)
	if err != nil {
		return ErrSigningFailed.Wrapf("failed to get private key: %v", err)
	}
	defer clearBytes(privateKey)

	// Sign the data
	var signature []byte
	switch activeKey.Algorithm {
	case veidtypes.ProofTypeEd25519:
		signature = ed25519.Sign(privateKey, dataToSign)
	default:
		return ErrSigningFailed.Wrapf("unsupported algorithm: %s", activeKey.Algorithm)
	}

	// Create and set the proof
	now := time.Now()
	proof := veidtypes.NewAttestationProof(
		activeKey.Algorithm,
		now,
		fmt.Sprintf("%s#%s", attestation.Issuer.ID, activeKey.KeyID),
		signature,
		attestation.Nonce,
	)
	attestation.SetProof(proof)

	// Audit log
	if s.auditor != nil {
		s.auditor.Log(ctx, audit.Event{
			Type:      audit.EventTypeAttestationSigned,
			Timestamp: now,
			Actor:     s.config.SignerID,
			Resource:  attestation.ID,
			Action:    "sign_attestation",
			Details: map[string]interface{}{
				"attestation_type": attestation.Type,
				"subject":          attestation.Subject.AccountAddress,
				"key_id":           activeKey.KeyID,
				"score":            attestation.Score,
			},
		})
	}

	return nil
}

// VerifyAttestation verifies an attestation signature.
func (s *DefaultSigner) VerifyAttestation(ctx context.Context, attestation *veidtypes.VerificationAttestation) (bool, error) {
	// Find the key by fingerprint
	keyInfo, err := s.GetKeyByFingerprint(ctx, attestation.Issuer.KeyFingerprint)
	if err != nil {
		return false, err
	}

	if !keyInfo.State.CanVerify() {
		return false, ErrKeyRevoked.Wrapf("key state: %s", keyInfo.State)
	}

	// Get canonical bytes
	dataToVerify, err := attestation.CanonicalBytes()
	if err != nil {
		return false, ErrVerificationFailed.Wrapf("failed to get canonical bytes: %v", err)
	}

	// Get signature bytes
	signatureBytes, err := attestation.GetProofBytes()
	if err != nil {
		return false, ErrVerificationFailed.Wrapf("failed to decode signature: %v", err)
	}

	// Verify based on algorithm
	var valid bool
	switch keyInfo.Algorithm {
	case veidtypes.ProofTypeEd25519:
		valid = ed25519.Verify(keyInfo.PublicKey, dataToVerify, signatureBytes)
	default:
		return false, ErrVerificationFailed.Wrapf("unsupported algorithm: %s", keyInfo.Algorithm)
	}

	// Audit log
	if s.auditor != nil {
		s.auditor.Log(ctx, audit.Event{
			Type:      audit.EventTypeAttestationVerified,
			Timestamp: time.Now(),
			Actor:     s.config.SignerID,
			Resource:  attestation.ID,
			Action:    "verify_attestation",
			Details: map[string]interface{}{
				"attestation_type": attestation.Type,
				"key_id":           keyInfo.KeyID,
				"valid":            valid,
			},
		})
	}

	return valid, nil
}

// GetActiveKey returns the currently active signing key info.
func (s *DefaultSigner) GetActiveKey(ctx context.Context) (*veidtypes.SignerKeyInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.activeKey == nil {
		return nil, ErrNoActiveKey
	}

	// Return a copy
	keyCopy := *s.activeKey
	return &keyCopy, nil
}

// GetKeyByID returns key info by key ID.
func (s *DefaultSigner) GetKeyByID(ctx context.Context, keyID string) (*veidtypes.SignerKeyInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key, ok := s.keys[keyID]
	if !ok {
		return nil, ErrKeyNotFound.Wrapf("key ID: %s", keyID)
	}

	keyCopy := *key
	return &keyCopy, nil
}

// GetKeyByFingerprint returns key info by fingerprint.
func (s *DefaultSigner) GetKeyByFingerprint(ctx context.Context, fingerprint string) (*veidtypes.SignerKeyInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, key := range s.keys {
		if key.Fingerprint == fingerprint {
			keyCopy := *key
			return &keyCopy, nil
		}
	}

	return nil, ErrKeyNotFound.Wrapf("fingerprint: %s...", fingerprint[:16])
}

// ListKeys returns all keys for this signer.
func (s *DefaultSigner) ListKeys(ctx context.Context) ([]*veidtypes.SignerKeyInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*veidtypes.SignerKeyInfo, 0, len(s.keys))
	for _, key := range s.keys {
		keyCopy := *key
		result = append(result, &keyCopy)
	}

	return result, nil
}

// RotateKey initiates a key rotation to a new key.
func (s *DefaultSigner) RotateKey(ctx context.Context, req *KeyRotationRequest) (*veidtypes.KeyRotationRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentRotation != nil && s.currentRotation.Status == veidtypes.RotationStatusInProgress {
		return nil, ErrRotationInProgress
	}

	if s.activeKey == nil {
		return nil, ErrNoActiveKey
	}

	now := time.Now()
	s.sequenceCounter++

	// Determine algorithm for new key
	algorithm := s.config.DefaultAlgorithm
	if req.NewKeyAlgorithm != "" && s.config.KeyPolicy.IsAlgorithmAllowed(req.NewKeyAlgorithm) {
		algorithm = req.NewKeyAlgorithm
	}

	// Generate new key
	newKeyInfo, privateKey, err := s.generateKeyPair(algorithm, s.sequenceCounter, now)
	if err != nil {
		return nil, err
	}

	// Determine overlap period
	overlapSeconds := s.config.KeyPolicy.RotationOverlapSeconds
	if req.Emergency {
		overlapSeconds = 0
	} else if req.OverrideOverlapSeconds != nil {
		overlapSeconds = *req.OverrideOverlapSeconds
	}

	// Activate the new key
	expiresAt := now.Add(time.Duration(s.config.KeyPolicy.MaxKeyAgeSeconds) * time.Second)
	if err := newKeyInfo.Activate(now, expiresAt); err != nil {
		return nil, err
	}

	// Store new key
	if err := s.storage.StoreKey(ctx, newKeyInfo, privateKey); err != nil {
		return nil, ErrKeyStorageError.Wrapf("failed to store new key: %v", err)
	}
	defer clearBytes(privateKey)

	// Mark old key as rotating
	oldKey := s.activeKey
	if err := oldKey.StartRotation(newKeyInfo.KeyID); err != nil {
		return nil, err
	}
	newKeyInfo.PredecessorKeyID = oldKey.KeyID

	// Create rotation record
	rotationID := uuid.New().String()
	rotation := veidtypes.NewKeyRotationRecord(
		rotationID,
		s.config.SignerID,
		oldKey,
		newKeyInfo,
		now,
		overlapSeconds,
		req.Reason,
		req.InitiatedBy,
		0, // Block height would come from context in on-chain scenarios
	)
	rotation.Notes = req.Notes

	// Update state
	s.keys[newKeyInfo.KeyID] = newKeyInfo
	s.activeKey = newKeyInfo
	s.currentRotation = rotation
	s.rotations[rotationID] = rotation
	s.registry.RotateKey(newKeyInfo.KeyID, now)

	// Audit log
	if s.auditor != nil {
		s.auditor.Log(ctx, audit.Event{
			Type:      audit.EventTypeKeyRotated,
			Timestamp: now,
			Actor:     req.InitiatedBy,
			Resource:  newKeyInfo.KeyID,
			Action:    "rotate_key",
			Details: map[string]interface{}{
				"old_key_id":      oldKey.KeyID,
				"new_key_id":      newKeyInfo.KeyID,
				"reason":          req.Reason,
				"overlap_seconds": overlapSeconds,
				"emergency":       req.Emergency,
			},
		})
	}

	s.logger.Info().
		Str("old_key", oldKey.KeyID).
		Str("new_key", newKeyInfo.KeyID).
		Str("reason", string(req.Reason)).
		Int64("overlap_seconds", overlapSeconds).
		Msg("key rotation initiated")

	return rotation, nil
}

// CompleteRotation completes a key rotation.
func (s *DefaultSigner) CompleteRotation(ctx context.Context, rotationID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	rotation, ok := s.rotations[rotationID]
	if !ok {
		return ErrRotationNotFound.Wrapf("rotation ID: %s", rotationID)
	}

	if rotation.Status != veidtypes.RotationStatusInProgress {
		return fmt.Errorf("rotation not in progress: %s", rotation.Status)
	}

	now := time.Now()

	// Revoke the old key
	oldKey, ok := s.keys[rotation.OldKeyID]
	if ok {
		if err := oldKey.Revoke(now, rotation.Reason); err != nil {
			s.logger.Warn().Err(err).Str("key_id", oldKey.KeyID).Msg("failed to revoke old key")
		}
	}

	// Complete the rotation
	rotation.Complete(now)

	// Clear current rotation if it matches
	if s.currentRotation != nil && s.currentRotation.RotationID == rotationID {
		s.currentRotation = nil
	}

	// Audit log
	if s.auditor != nil {
		s.auditor.Log(ctx, audit.Event{
			Type:      audit.EventTypeKeyRotationCompleted,
			Timestamp: now,
			Actor:     s.config.SignerID,
			Resource:  rotation.NewKeyID,
			Action:    "complete_rotation",
			Details: map[string]interface{}{
				"rotation_id": rotationID,
				"old_key_id":  rotation.OldKeyID,
				"new_key_id":  rotation.NewKeyID,
			},
		})
	}

	s.logger.Info().
		Str("rotation_id", rotationID).
		Str("old_key", rotation.OldKeyID).
		Str("new_key", rotation.NewKeyID).
		Msg("key rotation completed")

	return nil
}

// RevokeKey revokes a key immediately.
func (s *DefaultSigner) RevokeKey(ctx context.Context, keyID string, reason veidtypes.KeyRevocationReason) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key, ok := s.keys[keyID]
	if !ok {
		return ErrKeyNotFound.Wrapf("key ID: %s", keyID)
	}

	now := time.Now()
	if err := key.Revoke(now, reason); err != nil {
		return err
	}

	// If this was the active key, we have a problem
	if s.activeKey != nil && s.activeKey.KeyID == keyID {
		s.logger.Error().Str("key_id", keyID).Msg("active key revoked - no active key available!")
		s.activeKey = nil
	}

	// Audit log
	if s.auditor != nil {
		s.auditor.Log(ctx, audit.Event{
			Type:      audit.EventTypeKeyRevoked,
			Timestamp: now,
			Actor:     s.config.SignerID,
			Resource:  keyID,
			Action:    "revoke_key",
			Details: map[string]interface{}{
				"key_id": keyID,
				"reason": reason,
			},
		})
	}

	s.logger.Warn().
		Str("key_id", keyID).
		Str("reason", string(reason)).
		Msg("key revoked")

	return nil
}

// GetRotationStatus returns the status of a key rotation.
func (s *DefaultSigner) GetRotationStatus(ctx context.Context, rotationID string) (*veidtypes.KeyRotationRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rotation, ok := s.rotations[rotationID]
	if !ok {
		return nil, ErrRotationNotFound.Wrapf("rotation ID: %s", rotationID)
	}

	rotationCopy := *rotation
	return &rotationCopy, nil
}

// GetSignerInfo returns information about this signer.
func (s *DefaultSigner) GetSignerInfo(ctx context.Context) (*veidtypes.SignerRegistryEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.registry == nil {
		return nil, ErrServiceUnavailable.Wrap("signer not initialized")
	}

	registryCopy := *s.registry
	return &registryCopy, nil
}

// HealthCheck returns the health status of the signer service.
func (s *DefaultSigner) HealthCheck(ctx context.Context) (*HealthStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := &HealthStatus{
		Healthy:   true,
		Status:    "healthy",
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
		Warnings:  make([]string, 0),
	}

	// Check active key
	if s.activeKey == nil {
		status.Healthy = false
		status.Status = "no active key"
		return status, nil
	}

	status.ActiveKeyID = s.activeKey.KeyID

	// Check key expiration
	now := time.Now()
	if s.activeKey.ExpiresAt != nil {
		status.KeyExpiresAt = s.activeKey.ExpiresAt
		if s.activeKey.IsExpired(now) {
			status.Healthy = false
			status.Status = "active key expired"
		} else if s.activeKey.ShouldRotate(now, s.config.KeyPolicy) {
			status.Warnings = append(status.Warnings, "key rotation recommended")
		}
	}

	// Calculate key age
	if s.activeKey.ActivatedAt != nil {
		status.KeyAge = now.Sub(*s.activeKey.ActivatedAt)
	}

	// Check rotation status
	if s.currentRotation != nil && s.currentRotation.Status == veidtypes.RotationStatusInProgress {
		status.RotationPending = true
		status.Details["rotation_id"] = s.currentRotation.RotationID
		status.Details["rotation_started"] = s.currentRotation.InitiatedAt
	}

	// Check key storage health
	if err := s.storage.HealthCheck(ctx); err != nil {
		status.Warnings = append(status.Warnings, fmt.Sprintf("key storage warning: %v", err))
	}

	status.Details["total_keys"] = len(s.keys)
	status.Details["signer_id"] = s.config.SignerID

	return status, nil
}

// Close closes the signer service and releases resources.
func (s *DefaultSigner) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.storage != nil {
		if err := s.storage.Close(); err != nil {
			return err
		}
	}

	s.logger.Info().Msg("signer service closed")
	return nil
}

// clearBytes securely clears sensitive byte slices.
func clearBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

// Helper for base64 encoding (not used but available for external callers)
func encodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func decodeBase64(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}
