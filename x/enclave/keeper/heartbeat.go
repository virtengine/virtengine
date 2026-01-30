package keeper

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	"github.com/virtengine/virtengine/x/enclave/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ProcessHeartbeat processes an enclave heartbeat message
func (k Keeper) ProcessHeartbeat(ctx sdk.Context, msg types.MsgEnclaveHeartbeat) (*types.MsgEnclaveHeartbeatResponse, error) {
	// Parse validator address
	validatorAddr, err := sdk.AccAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid validator address: %w", err)
	}

	// Get enclave identity
	identity, found := k.GetEnclaveIdentity(ctx, validatorAddr)
	if !found {
		return nil, types.ErrEnclaveIdentityNotFound
	}

	// Check if enclave identity is active
	if identity.Status != types.EnclaveIdentityStatusActive {
		return nil, fmt.Errorf("enclave identity is not active: %s", identity.Status)
	}

	// Validate heartbeat timestamp
	if err := k.ValidateHeartbeatTimestamp(ctx, msg.Timestamp); err != nil {
		return nil, err
	}

	// Check for replay attacks using nonce
	if err := k.ValidateHeartbeatNonce(ctx, validatorAddr, msg.Nonce); err != nil {
		return nil, err
	}

	// Verify heartbeat signature
	if err := k.VerifyHeartbeatSignature(ctx, identity, msg); err != nil {
		// Record signature failure
		if recordErr := k.RecordSignatureFailure(ctx, validatorAddr); recordErr != nil {
			ctx.Logger().Error("failed to record signature failure", "error", recordErr)
		}
		return nil, err
	}

	// Record successful signature verification
	if err := k.RecordSignatureSuccess(ctx, validatorAddr); err != nil {
		ctx.Logger().Error("failed to record signature success", "error", err)
	}

	// Process optional attestation proof
	if len(msg.AttestationProof) > 0 {
		if err := k.ProcessHeartbeatAttestation(ctx, identity, msg.AttestationProof); err != nil {
			// Record attestation failure
			if recordErr := k.RecordAttestationFailure(ctx, validatorAddr); recordErr != nil {
				ctx.Logger().Error("failed to record attestation failure", "error", recordErr)
			}
			ctx.Logger().Error("heartbeat attestation verification failed", "error", err, "validator", msg.ValidatorAddress)
		} else {
			// Record successful attestation
			if recordErr := k.RecordAttestationSuccess(ctx, validatorAddr); recordErr != nil {
				ctx.Logger().Error("failed to record attestation success", "error", recordErr)
			}
		}
	}

	// Get or initialize health status
	health, exists := k.GetEnclaveHealthStatus(ctx, validatorAddr)
	if !exists {
		if err := k.InitializeHealthStatus(ctx, validatorAddr); err != nil {
			return nil, fmt.Errorf("failed to initialize health status: %w", err)
		}
		health, _ = k.GetEnclaveHealthStatus(ctx, validatorAddr)
	}

	// Record successful heartbeat
	health.RecordHeartbeat(msg.Timestamp)

	// Save updated health status
	if err := k.SetEnclaveHealthStatus(ctx, health); err != nil {
		return nil, fmt.Errorf("failed to update health status: %w", err)
	}

	// Update overall health status
	if err := k.UpdateHealthStatus(ctx, validatorAddr); err != nil {
		return nil, fmt.Errorf("failed to update health status: %w", err)
	}

	// Store nonce to prevent replay
	if err := k.StoreHeartbeatNonce(ctx, validatorAddr, msg.Nonce); err != nil {
		ctx.Logger().Error("failed to store heartbeat nonce", "error", err)
	}

	// Emit heartbeat received event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeEnclaveHeartbeatReceived,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
			sdk.NewAttribute(types.AttributeKeyLastHeartbeat, msg.Timestamp.Format(time.RFC3339)),
		),
	)

	// Get updated health status for response
	health, _ = k.GetEnclaveHealthStatus(ctx, validatorAddr)

	return &types.MsgEnclaveHeartbeatResponse{
		Success:       true,
		CurrentStatus: health.Status,
		Message:       fmt.Sprintf("Heartbeat processed successfully. Status: %s", health.Status.String()),
	}, nil
}

// ValidateHeartbeatTimestamp validates the heartbeat timestamp
func (k Keeper) ValidateHeartbeatTimestamp(ctx sdk.Context, timestamp time.Time) error {
	currentTime := ctx.BlockTime()

	// Check if timestamp is too far in the past (more than 5 minutes)
	if currentTime.Sub(timestamp) > 5*time.Minute {
		return fmt.Errorf("heartbeat timestamp too old: %v", timestamp)
	}

	// Check if timestamp is in the future (allow 1 minute clock drift)
	if timestamp.Sub(currentTime) > 1*time.Minute {
		return fmt.Errorf("heartbeat timestamp in the future: %v", timestamp)
	}

	return nil
}

// ValidateHeartbeatNonce checks if the nonce has been used before
func (k Keeper) ValidateHeartbeatNonce(ctx sdk.Context, validatorAddr sdk.AccAddress, nonce uint64) error {
	store := ctx.KVStore(k.storeKey)
	nonceKey := k.heartbeatNonceKey(validatorAddr, nonce)

	if store.Has(nonceKey) {
		return types.ErrHeartbeatReplay
	}

	return nil
}

// StoreHeartbeatNonce stores a used nonce
func (k Keeper) StoreHeartbeatNonce(ctx sdk.Context, validatorAddr sdk.AccAddress, nonce uint64) error {
	store := ctx.KVStore(k.storeKey)
	nonceKey := k.heartbeatNonceKey(validatorAddr, nonce)

	// Store nonce with expiry timestamp (keep for 24 hours)
	expiryTime := ctx.BlockTime().Add(24 * time.Hour)
	bz, err := json.Marshal(expiryTime)
	if err != nil {
		return err
	}

	store.Set(nonceKey, bz)
	return nil
}

// heartbeatNonceKey creates a store key for heartbeat nonces
func (k Keeper) heartbeatNonceKey(validatorAddr sdk.AccAddress, nonce uint64) []byte {
	// Use a separate prefix for nonces
	prefix := []byte{0x09} // New prefix for heartbeat nonces
	key := make([]byte, 0, len(prefix)+len(validatorAddr)+8)
	key = append(key, prefix...)
	key = append(key, validatorAddr.Bytes()...)

	// Append nonce as big-endian bytes
	nonceBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBytes, nonce)
	key = append(key, nonceBytes...)

	return key
}

// VerifyHeartbeatSignature verifies the signature on a heartbeat message
func (k Keeper) VerifyHeartbeatSignature(ctx sdk.Context, identity types.EnclaveIdentity, msg types.MsgEnclaveHeartbeat) error {
	// Create the message to verify
	heartbeatData := struct {
		ValidatorAddress string
		Timestamp        time.Time
		Nonce            uint64
	}{
		ValidatorAddress: msg.ValidatorAddress,
		Timestamp:        msg.Timestamp,
		Nonce:            msg.Nonce,
	}

	// Serialize to JSON
	dataBytes, err := json.Marshal(heartbeatData)
	if err != nil {
		return fmt.Errorf("failed to marshal heartbeat data: %w", err)
	}

	// Hash the data
	hash := sha256.Sum256(dataBytes)

	// Verify signature using the enclave's signing public key
	// NOTE: This is a placeholder - actual signature verification would use
	// the appropriate cryptographic library based on the signature scheme
	if err := k.verifySignature(identity.SigningPubKey, hash[:], msg.Signature); err != nil {
		return types.ErrHeartbeatSignatureInvalid
	}

	return nil
}

// verifySignature is a helper function to verify signatures
// NOTE: This is a placeholder implementation
func (k Keeper) verifySignature(pubKey []byte, message []byte, signature []byte) error {
	// TODO: Implement actual signature verification based on the signature scheme
	// This would typically use ed25519, ECDSA, or another appropriate scheme

	// For now, just check that signature is not empty
	if len(signature) == 0 {
		return fmt.Errorf("empty signature")
	}

	// In a real implementation, you would:
	// 1. Parse the public key
	// 2. Verify the signature over the message
	// 3. Return an error if verification fails

	return nil
}

// ProcessHeartbeatAttestation processes an optional attestation in a heartbeat
func (k Keeper) ProcessHeartbeatAttestation(ctx sdk.Context, identity types.EnclaveIdentity, attestationProof []byte) error {
	// Verify attestation format and contents
	if len(attestationProof) == 0 {
		return fmt.Errorf("empty attestation proof")
	}

	// TODO: Implement actual attestation verification
	// This would involve:
	// 1. Parsing the attestation quote
	// 2. Verifying the attestation chain
	// 3. Checking measurement values match the registered identity
	// 4. Verifying the quote signature

	// For now, just validate that the proof is not empty
	// and has a reasonable size (attestation quotes are typically 1-10KB)
	if len(attestationProof) > 100*1024 {
		return fmt.Errorf("attestation proof too large: %d bytes", len(attestationProof))
	}

	return nil
}

// CleanupExpiredNonces removes expired heartbeat nonces
func (k Keeper) CleanupExpiredNonces(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	prefix := []byte{0x09} // Heartbeat nonce prefix
	iterator := sdk.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	currentTime := ctx.BlockTime()
	var keysToDelete [][]byte

	for ; iterator.Valid(); iterator.Next() {
		var expiryTime time.Time
		if err := json.Unmarshal(iterator.Value(), &expiryTime); err != nil {
			ctx.Logger().Error("failed to unmarshal nonce expiry time", "error", err)
			continue
		}

		// If expired, mark for deletion
		if currentTime.After(expiryTime) {
			keysToDelete = append(keysToDelete, iterator.Key())
		}
	}

	// Delete expired nonces
	for _, key := range keysToDelete {
		store.Delete(key)
	}

	if len(keysToDelete) > 0 {
		ctx.Logger().Debug("cleaned up expired heartbeat nonces", "count", len(keysToDelete))
	}
}
