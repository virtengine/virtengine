// Package signer provides the verification attestation signing service.
package signer

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/pkg/verification/audit"
	"github.com/virtengine/virtengine/pkg/verification/keystorage"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

func TestDefaultSigner_NewSigner(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.Nop()

	storage, err := keystorage.NewMemoryStorage(nil)
	require.NoError(t, err)
	defer storage.Close()

	auditor := audit.NewMemoryLogger(audit.DefaultConfig(), logger)
	defer auditor.Close()

	config := SignerConfig{
		SignerID:         "test-signer-1",
		SignerName:       "Test Signer",
		KeyStorageType:   KeyStorageMemory,
		DefaultAlgorithm: veidtypes.ProofTypeEd25519,
		KeyPolicy:        veidtypes.DefaultSignerKeyPolicy(),
		AuditLogEnabled:  true,
	}

	signer, err := NewDefaultSigner(ctx, config, storage, auditor, logger)
	require.NoError(t, err)
	defer signer.Close()

	// Verify initial key was generated
	activeKey, err := signer.GetActiveKey(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, activeKey.KeyID)
	assert.Equal(t, veidtypes.SignerKeyStateActive, activeKey.State)
	assert.Equal(t, veidtypes.ProofTypeEd25519, activeKey.Algorithm)
}

func TestDefaultSigner_SignAndVerifyAttestation(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.Nop()

	storage, err := keystorage.NewMemoryStorage(nil)
	require.NoError(t, err)
	defer storage.Close()

	auditor := audit.NewMemoryLogger(audit.DefaultConfig(), logger)
	defer auditor.Close()

	config := SignerConfig{
		SignerID:         "test-signer-sign",
		SignerName:       "Test Signer",
		DefaultAlgorithm: veidtypes.ProofTypeEd25519,
		KeyPolicy:        veidtypes.DefaultSignerKeyPolicy(),
		AuditLogEnabled:  true,
	}

	signer, err := NewDefaultSigner(ctx, config, storage, auditor, logger)
	require.NoError(t, err)
	defer signer.Close()

	// Get active key to use its fingerprint
	activeKey, err := signer.GetActiveKey(ctx)
	require.NoError(t, err)

	// Create a test attestation
	nonce := make([]byte, 32)
	for i := range nonce {
		nonce[i] = byte(i)
	}

	now := time.Now()
	attestation := veidtypes.NewVerificationAttestation(
		veidtypes.AttestationIssuer{
			ID:             "did:virtengine:test",
			KeyFingerprint: activeKey.Fingerprint, // Use actual fingerprint
		},
		veidtypes.NewAttestationSubject("virtengine1test123"),
		veidtypes.AttestationTypeFacialVerification,
		nonce,
		now,
		24*time.Hour,
		85,
		90,
	)

	// Sign the attestation
	err = signer.SignAttestation(ctx, attestation)
	require.NoError(t, err)

	// Verify proof was added
	assert.NotEmpty(t, attestation.Proof.ProofValue)
	assert.Equal(t, veidtypes.ProofTypeEd25519, attestation.Proof.Type)
	assert.Equal(t, "assertionMethod", attestation.Proof.ProofPurpose)

	// Verify the attestation
	valid, err := signer.VerifyAttestation(ctx, attestation)
	require.NoError(t, err)
	assert.True(t, valid)

	// Tamper with attestation and verify fails
	attestation.Score = 99
	valid, err = signer.VerifyAttestation(ctx, attestation)
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestDefaultSigner_KeyRotation(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.Nop()

	storage, err := keystorage.NewMemoryStorage(nil)
	require.NoError(t, err)
	defer storage.Close()

	auditor := audit.NewMemoryLogger(audit.DefaultConfig(), logger)
	defer auditor.Close()

	config := SignerConfig{
		SignerID:         "test-signer-rotate",
		SignerName:       "Test Signer",
		DefaultAlgorithm: veidtypes.ProofTypeEd25519,
		KeyPolicy:        veidtypes.DefaultSignerKeyPolicy(),
		AuditLogEnabled:  true,
	}

	signer, err := NewDefaultSigner(ctx, config, storage, auditor, logger)
	require.NoError(t, err)
	defer signer.Close()

	// Get initial key
	initialKey, err := signer.GetActiveKey(ctx)
	require.NoError(t, err)

	// Rotate key
	rotation, err := signer.RotateKey(ctx, &KeyRotationRequest{
		Reason:      veidtypes.RevocationReasonRotation,
		InitiatedBy: "test-admin",
		Notes:       "Test rotation",
	})
	require.NoError(t, err)
	assert.Equal(t, veidtypes.RotationStatusInProgress, rotation.Status)
	assert.Equal(t, initialKey.KeyID, rotation.OldKeyID)

	// Get new active key
	newKey, err := signer.GetActiveKey(ctx)
	require.NoError(t, err)
	assert.NotEqual(t, initialKey.KeyID, newKey.KeyID)
	assert.Equal(t, veidtypes.SignerKeyStateActive, newKey.State)

	// Old key should be in rotating state
	oldKey, err := signer.GetKeyByID(ctx, initialKey.KeyID)
	require.NoError(t, err)
	assert.Equal(t, veidtypes.SignerKeyStateRotating, oldKey.State)

	// Complete rotation
	err = signer.CompleteRotation(ctx, rotation.RotationID)
	require.NoError(t, err)

	// Old key should now be revoked
	oldKey, err = signer.GetKeyByID(ctx, initialKey.KeyID)
	require.NoError(t, err)
	assert.Equal(t, veidtypes.SignerKeyStateRevoked, oldKey.State)

	// Check rotation status
	rotationStatus, err := signer.GetRotationStatus(ctx, rotation.RotationID)
	require.NoError(t, err)
	assert.Equal(t, veidtypes.RotationStatusCompleted, rotationStatus.Status)
}

func TestDefaultSigner_RevokeKey(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.Nop()

	storage, err := keystorage.NewMemoryStorage(nil)
	require.NoError(t, err)
	defer storage.Close()

	auditor := audit.NewMemoryLogger(audit.DefaultConfig(), logger)
	defer auditor.Close()

	config := SignerConfig{
		SignerID:         "test-signer-revoke",
		SignerName:       "Test Signer",
		DefaultAlgorithm: veidtypes.ProofTypeEd25519,
		KeyPolicy:        veidtypes.DefaultSignerKeyPolicy(),
		AuditLogEnabled:  true,
	}

	signer, err := NewDefaultSigner(ctx, config, storage, auditor, logger)
	require.NoError(t, err)
	defer signer.Close()

	// Get initial key
	initialKey, err := signer.GetActiveKey(ctx)
	require.NoError(t, err)

	// Revoke the active key
	err = signer.RevokeKey(ctx, initialKey.KeyID, veidtypes.RevocationReasonCompromised)
	require.NoError(t, err)

	// Key should be revoked
	revokedKey, err := signer.GetKeyByID(ctx, initialKey.KeyID)
	require.NoError(t, err)
	assert.Equal(t, veidtypes.SignerKeyStateRevoked, revokedKey.State)

	// No active key should be available
	_, err = signer.GetActiveKey(ctx)
	assert.Error(t, err)
}

func TestDefaultSigner_HealthCheck(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.Nop()

	storage, err := keystorage.NewMemoryStorage(nil)
	require.NoError(t, err)
	defer storage.Close()

	auditor := audit.NewMemoryLogger(audit.DefaultConfig(), logger)
	defer auditor.Close()

	config := SignerConfig{
		SignerID:         "test-signer-health",
		SignerName:       "Test Signer",
		DefaultAlgorithm: veidtypes.ProofTypeEd25519,
		KeyPolicy:        veidtypes.DefaultSignerKeyPolicy(),
		AuditLogEnabled:  true,
	}

	signer, err := NewDefaultSigner(ctx, config, storage, auditor, logger)
	require.NoError(t, err)
	defer signer.Close()

	health, err := signer.HealthCheck(ctx)
	require.NoError(t, err)
	assert.True(t, health.Healthy)
	assert.Equal(t, "healthy", health.Status)
	assert.NotEmpty(t, health.ActiveKeyID)
	assert.False(t, health.RotationPending)
}

func TestDefaultSigner_ListKeys(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.Nop()

	storage, err := keystorage.NewMemoryStorage(nil)
	require.NoError(t, err)
	defer storage.Close()

	auditor := audit.NewMemoryLogger(audit.DefaultConfig(), logger)
	defer auditor.Close()

	config := SignerConfig{
		SignerID:         "test-signer-list",
		SignerName:       "Test Signer",
		DefaultAlgorithm: veidtypes.ProofTypeEd25519,
		KeyPolicy:        veidtypes.DefaultSignerKeyPolicy(),
		AuditLogEnabled:  true,
	}

	signer, err := NewDefaultSigner(ctx, config, storage, auditor, logger)
	require.NoError(t, err)
	defer signer.Close()

	// Initially one key
	keys, err := signer.ListKeys(ctx)
	require.NoError(t, err)
	assert.Len(t, keys, 1)

	// Rotate key
	_, err = signer.RotateKey(ctx, &KeyRotationRequest{
		Reason:      veidtypes.RevocationReasonRotation,
		InitiatedBy: "test-admin",
	})
	require.NoError(t, err)

	// Now two keys
	keys, err = signer.ListKeys(ctx)
	require.NoError(t, err)
	assert.Len(t, keys, 2)
}

func TestDefaultSigner_GetSignerInfo(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.Nop()

	storage, err := keystorage.NewMemoryStorage(nil)
	require.NoError(t, err)
	defer storage.Close()

	auditor := audit.NewMemoryLogger(audit.DefaultConfig(), logger)
	defer auditor.Close()

	config := SignerConfig{
		SignerID:         "test-signer-info",
		SignerName:       "Test Signer Info",
		ValidatorAddress: "virtenginevaloper1test",
		DefaultAlgorithm: veidtypes.ProofTypeEd25519,
		KeyPolicy:        veidtypes.DefaultSignerKeyPolicy(),
		AuditLogEnabled:  true,
	}

	signer, err := NewDefaultSigner(ctx, config, storage, auditor, logger)
	require.NoError(t, err)
	defer signer.Close()

	info, err := signer.GetSignerInfo(ctx)
	require.NoError(t, err)
	assert.Equal(t, "test-signer-info", info.SignerID)
	assert.Equal(t, "Test Signer Info", info.Name)
	assert.Equal(t, "virtenginevaloper1test", info.ValidatorAddress)
	assert.True(t, info.Active)
}

func TestDefaultSigner_AuditEvents(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.Nop()

	storage, err := keystorage.NewMemoryStorage(nil)
	require.NoError(t, err)
	defer storage.Close()

	auditor := audit.NewMemoryLogger(audit.DefaultConfig(), logger)
	defer auditor.Close()

	config := SignerConfig{
		SignerID:         "test-signer-audit",
		SignerName:       "Test Signer",
		DefaultAlgorithm: veidtypes.ProofTypeEd25519,
		KeyPolicy:        veidtypes.DefaultSignerKeyPolicy(),
		AuditLogEnabled:  true,
	}

	signer, err := NewDefaultSigner(ctx, config, storage, auditor, logger)
	require.NoError(t, err)
	defer signer.Close()

	// Get active key to use its fingerprint
	activeKey, err := signer.GetActiveKey(ctx)
	require.NoError(t, err)

	// Sign an attestation
	nonce := make([]byte, 32)
	for i := range nonce {
		nonce[i] = byte(i + 1)
	}

	attestation := veidtypes.NewVerificationAttestation(
		veidtypes.AttestationIssuer{
			ID:             "did:virtengine:test",
			KeyFingerprint: activeKey.Fingerprint,
		},
		veidtypes.NewAttestationSubject("virtengine1test456"),
		veidtypes.AttestationTypeEmailVerification,
		nonce,
		time.Now(),
		24*time.Hour,
		100,
		95,
	)

	err = signer.SignAttestation(ctx, attestation)
	require.NoError(t, err)

	// Check audit events
	events := auditor.GetEvents()
	assert.GreaterOrEqual(t, len(events), 2) // key_generated + attestation_signed

	// Find attestation signed event
	var foundSignEvent bool
	for _, e := range events {
		if e.Type == audit.EventTypeAttestationSigned {
			foundSignEvent = true
			assert.Equal(t, attestation.ID, e.Resource)
			break
		}
	}
	assert.True(t, foundSignEvent, "should have attestation_signed audit event")
}
