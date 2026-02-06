// Package keeper provides the VEID module keeper.
//
// This file implements key rotation mechanisms for VEID approved client keys.
// It provides secure key rotation with overlap periods to prevent service disruption.
//
// Task Reference: 22A - Pre-mainnet security hardening
package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Key Rotation Types
// ============================================================================

// ClientKeyRotationStatus represents the state of a key rotation
type ClientKeyRotationStatus string

const (
	// ClientKeyRotationStatusPending indicates rotation has been initiated
	ClientKeyRotationStatusPending ClientKeyRotationStatus = "pending"

	// ClientKeyRotationStatusActive indicates both old and new keys are valid
	ClientKeyRotationStatusActive ClientKeyRotationStatus = "active"

	// ClientKeyRotationStatusCompleted indicates rotation is complete
	ClientKeyRotationStatusCompleted ClientKeyRotationStatus = "completed"

	// DefaultKeyRotationOverlapBlocks is the default overlap period (1 day at 5s blocks)
	DefaultKeyRotationOverlapBlocks int64 = 17280

	// MaxKeyRotationOverlapBlocks caps the overlap period (7 days)
	MaxKeyRotationOverlapBlocks int64 = 120960
)

// ClientKeyRotation tracks a key rotation for an approved client
type ClientKeyRotation struct {
	// RotationID is a unique identifier for this rotation
	RotationID string `json:"rotation_id"`

	// ClientID is the approved client being rotated
	ClientID string `json:"client_id"`

	// OldKeyHash is the SHA-256 hash of the old public key
	OldKeyHash string `json:"old_key_hash"`

	// NewKeyHash is the SHA-256 hash of the new public key
	NewKeyHash string `json:"new_key_hash"`

	// NewPublicKey is the new public key bytes
	NewPublicKey []byte `json:"new_public_key"`

	// NewAlgorithm is the algorithm for the new key
	NewAlgorithm string `json:"new_algorithm"`

	// Status is the current rotation state
	Status ClientKeyRotationStatus `json:"status"`

	// InitiatedAt is when the rotation was initiated
	InitiatedAt time.Time `json:"initiated_at"`

	// OverlapStartHeight is when both keys become valid
	OverlapStartHeight int64 `json:"overlap_start_height"`

	// OverlapEndHeight is when the old key becomes invalid
	OverlapEndHeight int64 `json:"overlap_end_height"`

	// CompletedAt is when the rotation was finalized
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// InitiatedBy is the authority that initiated this rotation
	InitiatedBy string `json:"initiated_by"`
}

// ============================================================================
// Key Rotation Operations
// ============================================================================

// InitiateClientKeyRotation begins a key rotation for an approved client.
// Only governance authority can initiate rotations.
func (k Keeper) InitiateClientKeyRotation(
	ctx sdk.Context,
	authority string,
	clientID string,
	newPublicKey []byte,
	newAlgorithm string,
	overlapBlocks int64,
) (*ClientKeyRotation, error) {
	// Verify authority
	if authority != k.authority {
		return nil, types.ErrUnauthorized.Wrapf(
			"only governance authority can rotate client keys: expected %s, got %s",
			k.authority, authority,
		)
	}

	// Validate the client exists
	client, found := k.GetApprovedClient(ctx, clientID)
	if !found {
		return nil, types.ErrClientNotApproved.Wrapf("client %s not found", clientID)
	}

	// Validate new key
	if len(newPublicKey) == 0 {
		return nil, types.ErrInvalidKeyRotation.Wrap("new public key cannot be empty")
	}

	// Validate algorithm
	if newAlgorithm != AlgorithmEd25519 && newAlgorithm != AlgorithmSecp256k1 {
		return nil, types.ErrInvalidKeyRotation.Wrapf("unsupported algorithm: %s", newAlgorithm)
	}

	// Validate key size for algorithm
	switch newAlgorithm {
	case AlgorithmEd25519:
		if len(newPublicKey) != Ed25519PublicKeySize {
			return nil, types.ErrInvalidKeyRotation.Wrapf(
				"ed25519 key must be %d bytes, got %d",
				Ed25519PublicKeySize, len(newPublicKey),
			)
		}
	case AlgorithmSecp256k1:
		if len(newPublicKey) != Secp256k1PublicKeySize {
			return nil, types.ErrInvalidKeyRotation.Wrapf(
				"secp256k1 key must be %d bytes, got %d",
				Secp256k1PublicKeySize, len(newPublicKey),
			)
		}
	}

	// Validate overlap period
	if overlapBlocks <= 0 {
		overlapBlocks = DefaultKeyRotationOverlapBlocks
	}
	if overlapBlocks > MaxKeyRotationOverlapBlocks {
		return nil, types.ErrInvalidKeyRotation.Wrapf(
			"overlap period %d blocks exceeds maximum %d",
			overlapBlocks, MaxKeyRotationOverlapBlocks,
		)
	}

	// Check no active rotation exists
	if existing, found := k.GetActiveClientKeyRotation(ctx, clientID); found {
		return nil, types.ErrInvalidKeyRotation.Wrapf(
			"client %s already has an active rotation: %s",
			clientID, existing.RotationID,
		)
	}

	// Compute key hashes
	oldKeyHash := sha256.Sum256(client.PublicKey)
	newKeyHash := sha256.Sum256(newPublicKey)

	// Generate rotation ID
	idData := fmt.Sprintf("%s|%d|%s", clientID, ctx.BlockHeight(), hex.EncodeToString(newKeyHash[:]))
	idHash := sha256.Sum256([]byte(idData))
	rotationID := hex.EncodeToString(idHash[:16])

	currentHeight := ctx.BlockHeight()
	rotation := &ClientKeyRotation{
		RotationID:         rotationID,
		ClientID:           clientID,
		OldKeyHash:         hex.EncodeToString(oldKeyHash[:]),
		NewKeyHash:         hex.EncodeToString(newKeyHash[:]),
		NewPublicKey:       newPublicKey,
		NewAlgorithm:       newAlgorithm,
		Status:             ClientKeyRotationStatusActive,
		InitiatedAt:        ctx.BlockTime(),
		OverlapStartHeight: currentHeight,
		OverlapEndHeight:   currentHeight + overlapBlocks,
		InitiatedBy:        authority,
	}

	if err := k.setClientKeyRotation(ctx, rotation); err != nil {
		return nil, err
	}

	k.Logger(ctx).Info(
		"client key rotation initiated",
		"client_id", clientID,
		"rotation_id", rotationID,
		"overlap_end_height", currentHeight+overlapBlocks,
	)

	return rotation, nil
}

// CompleteClientKeyRotation finalizes a key rotation, replacing the old key with the new one
func (k Keeper) CompleteClientKeyRotation(ctx sdk.Context, clientID string) error {
	rotation, found := k.GetActiveClientKeyRotation(ctx, clientID)
	if !found {
		return types.ErrInvalidKeyRotation.Wrapf("no active rotation for client %s", clientID)
	}

	// Verify overlap period has elapsed
	if ctx.BlockHeight() < rotation.OverlapEndHeight {
		return types.ErrInvalidKeyRotation.Wrapf(
			"overlap period not elapsed: current height %d, end height %d",
			ctx.BlockHeight(), rotation.OverlapEndHeight,
		)
	}

	// Update the client's key
	client, found := k.GetApprovedClient(ctx, clientID)
	if !found {
		return types.ErrClientNotApproved.Wrapf("client %s not found", clientID)
	}

	client.PublicKey = rotation.NewPublicKey
	client.Algorithm = rotation.NewAlgorithm

	if err := k.SetApprovedClient(ctx, client); err != nil {
		return fmt.Errorf("failed to update client key: %w", err)
	}

	// Mark rotation as completed
	now := ctx.BlockTime()
	rotation.Status = ClientKeyRotationStatusCompleted
	rotation.CompletedAt = &now

	if err := k.setClientKeyRotation(ctx, rotation); err != nil {
		return err
	}

	k.Logger(ctx).Info(
		"client key rotation completed",
		"client_id", clientID,
		"rotation_id", rotation.RotationID,
	)

	return nil
}

// IsKeyValidDuringRotation checks if a given key is valid during a rotation overlap period.
// During overlap, both old and new keys are accepted.
func (k Keeper) IsKeyValidDuringRotation(ctx sdk.Context, clientID string, keyBytes []byte) bool {
	rotation, found := k.GetActiveClientKeyRotation(ctx, clientID)
	if !found {
		// No active rotation, use normal validation
		return true
	}

	currentHeight := ctx.BlockHeight()
	if currentHeight < rotation.OverlapStartHeight || currentHeight > rotation.OverlapEndHeight {
		return true
	}

	// During overlap, check if key matches either old or new
	keyHash := sha256.Sum256(keyBytes)
	keyHashHex := hex.EncodeToString(keyHash[:])

	return keyHashHex == rotation.OldKeyHash || keyHashHex == rotation.NewKeyHash
}

// ============================================================================
// Key Rotation Store Operations
// ============================================================================

// GetActiveClientKeyRotation retrieves the active key rotation for a client
func (k Keeper) GetActiveClientKeyRotation(ctx sdk.Context, clientID string) (*ClientKeyRotation, bool) {
	store := ctx.KVStore(k.skey)
	key := types.ClientKeyRotationKey(clientID)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var rotation ClientKeyRotation
	if err := json.Unmarshal(bz, &rotation); err != nil {
		return nil, false
	}

	// Only return if actually active
	if rotation.Status != ClientKeyRotationStatusActive {
		return nil, false
	}

	return &rotation, true
}

// setClientKeyRotation stores a key rotation record
func (k Keeper) setClientKeyRotation(ctx sdk.Context, rotation *ClientKeyRotation) error {
	bz, err := json.Marshal(rotation)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	store.Set(types.ClientKeyRotationKey(rotation.ClientID), bz)
	return nil
}

// ProcessExpiredKeyRotations checks for and completes any expired key rotations.
// This should be called in EndBlock.
func (k Keeper) ProcessExpiredKeyRotations(ctx sdk.Context) {
	currentHeight := ctx.BlockHeight()

	k.WithApprovedClients(ctx, func(client types.ApprovedClient) bool {
		rotation, found := k.GetActiveClientKeyRotation(ctx, client.ClientID)
		if !found {
			return false
		}

		// Auto-complete rotation when overlap period expires
		if currentHeight >= rotation.OverlapEndHeight {
			if err := k.CompleteClientKeyRotation(ctx, client.ClientID); err != nil {
				k.Logger(ctx).Error(
					"failed to auto-complete key rotation",
					"client_id", client.ClientID,
					"error", err,
				)
			}
		}

		return false
	})
}
