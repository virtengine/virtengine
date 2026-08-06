package keeper

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"

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
	if identity.Status != types.EnclaveIdentityStatusActive && identity.Status != types.EnclaveIdentityStatusRotating {
		return nil, fmt.Errorf("enclave identity is not active: %s", identity.Status)
	}

	// Check for replay attacks using nonce
	if err := k.ValidateHeartbeatNonce(ctx, validatorAddr, msg.Nonce); err != nil {
		return nil, err
	}

	// Validate heartbeat timestamp
	if err := k.ValidateHeartbeatTimestamp(ctx, validatorAddr, msg.Timestamp); err != nil {
		return nil, err
	}

	// Verify heartbeat signature
	if err := k.VerifyHeartbeatSignature(ctx, *identity, msg); err != nil {
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
		if err := k.ProcessHeartbeatAttestation(ctx, *identity, msg.AttestationProof); err != nil {
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
func (k Keeper) ValidateHeartbeatTimestamp(ctx sdk.Context, validatorAddr sdk.AccAddress, timestamp time.Time) error {
	currentTime := ctx.BlockTime()

	// Check if timestamp is too far in the past (more than 5 minutes)
	if currentTime.Sub(timestamp) > 5*time.Minute {
		return types.ErrInvalidHeartbeat.Wrapf("heartbeat timestamp too old: %v", timestamp)
	}

	// Check if timestamp is in the future (allow 1 minute clock drift)
	if timestamp.Sub(currentTime) > 1*time.Minute {
		return types.ErrInvalidHeartbeat.Wrapf("heartbeat timestamp in the future: %v", timestamp)
	}

	health, exists := k.GetEnclaveHealthStatus(ctx, validatorAddr)
	if exists && !health.LastHeartbeat.IsZero() && !timestamp.After(health.LastHeartbeat) {
		return types.ErrInvalidHeartbeat.Wrapf(
			"heartbeat timestamp must be newer than last accepted heartbeat: got %s, last %s",
			timestamp.UTC().Format(time.RFC3339Nano),
			health.LastHeartbeat.UTC().Format(time.RFC3339Nano),
		)
	}

	return nil
}

// ValidateHeartbeatNonce checks if the nonce has been used before
func (k Keeper) ValidateHeartbeatNonce(ctx sdk.Context, validatorAddr sdk.AccAddress, nonce uint64) error {
	store := ctx.KVStore(k.StoreKey())
	nonceKey := types.HeartbeatNonceKey(validatorAddr, nonce)

	if store.Has(nonceKey) {
		return types.ErrHeartbeatReplay.Wrapf("nonce %d already used", nonce)
	}

	lastNonce := store.Get(types.HeartbeatLastNonceKey(validatorAddr))
	if len(lastNonce) == 8 {
		highestAccepted := binary.BigEndian.Uint64(lastNonce)
		if nonce <= highestAccepted {
			return types.ErrHeartbeatReplay.Wrapf("nonce %d is not greater than last accepted nonce %d", nonce, highestAccepted)
		}
	}

	return nil
}

// StoreHeartbeatNonce stores a used nonce
func (k Keeper) StoreHeartbeatNonce(ctx sdk.Context, validatorAddr sdk.AccAddress, nonce uint64) error {
	store := ctx.KVStore(k.StoreKey())
	nonceKey := types.HeartbeatNonceKey(validatorAddr, nonce)

	// Store nonce with expiry timestamp (keep for 24 hours)
	expiryTime := ctx.BlockTime().Add(24 * time.Hour)
	bz, err := json.Marshal(expiryTime)
	if err != nil {
		return err
	}

	store.Set(nonceKey, bz)
	lastNonce := make([]byte, 8)
	binary.BigEndian.PutUint64(lastNonce, nonce)
	store.Set(types.HeartbeatLastNonceKey(validatorAddr), lastNonce)
	return nil
}

// VerifyHeartbeatSignature verifies the signature on a heartbeat message
func (k Keeper) VerifyHeartbeatSignature(ctx sdk.Context, identity types.EnclaveIdentity, msg types.MsgEnclaveHeartbeat) error {
	payload, err := heartbeatSigningPayload(msg)
	if err != nil {
		return types.ErrHeartbeatSignatureInvalid.Wrapf("failed to build heartbeat signing payload: %v", err)
	}

	validatorAddr, err := sdk.AccAddressFromBech32(identity.ValidatorAddress)
	if err != nil {
		return types.ErrHeartbeatSignatureInvalid.Wrapf("invalid validator address: %v", err)
	}

	validKeys, staleKeys := k.heartbeatSigningKeys(ctx, validatorAddr, identity)
	for _, pubKey := range validKeys {
		if err := verifySigningKeySignature(pubKey, payload, msg.Signature); err == nil {
			return nil
		}
	}

	for _, pubKey := range staleKeys {
		if err := verifySigningKeySignature(pubKey, payload, msg.Signature); err == nil {
			return types.ErrHeartbeatSignatureInvalid.Wrap("stale rotation signing key")
		}
	}

	if err := verifySigningKeySignature(identity.SigningPubKey, payload, msg.Signature); err != nil {
		return types.ErrHeartbeatSignatureInvalid.Wrap(err.Error())
	}

	return nil
}

// ProcessHeartbeatAttestation processes an optional attestation in a heartbeat
func (k Keeper) ProcessHeartbeatAttestation(ctx sdk.Context, identity types.EnclaveIdentity, attestationProof []byte) error {
	// Verify attestation format and contents
	if len(attestationProof) == 0 {
		return fmt.Errorf("empty attestation proof")
	}

	// Validate proof size (attestation quotes are typically 1-10KB)
	if len(attestationProof) > 100*1024 {
		return fmt.Errorf("attestation proof too large: %d bytes", len(attestationProof))
	}

	// Parse and verify attestation based on TEE type
	switch identity.TeeType {
	case types.TEETypeSGX:
		return k.verifySGXHeartbeatAttestation(ctx, identity, attestationProof)
	case types.TEETypeSEVSNP:
		return k.verifySEVSNPHeartbeatAttestation(ctx, identity, attestationProof)
	case types.TEETypeNitro:
		return k.verifyNitroHeartbeatAttestation(ctx, identity, attestationProof)
	default:
		return fmt.Errorf("unsupported TEE type for attestation: %s", identity.TeeType)
	}
}

// verifySGXHeartbeatAttestation verifies an SGX DCAP attestation quote for heartbeat
func (k Keeper) verifySGXHeartbeatAttestation(ctx sdk.Context, identity types.EnclaveIdentity, attestationProof []byte) error {
	// Parse the SGX DCAP quote
	quote, err := types.ParseSGXDCAPQuoteV3(attestationProof)
	if err != nil {
		return fmt.Errorf("failed to parse SGX DCAP quote: %w", err)
	}

	// Verify debug mode is disabled in production
	if quote.Report.DebugEnabled() {
		return fmt.Errorf("SGX enclave is running in debug mode")
	}

	// Verify the measurement hash matches the registered identity
	measurementHash := sha256.Sum256(quote.Report.MRENCLAVE[:])
	if len(identity.MeasurementHash) != len(measurementHash) {
		return fmt.Errorf("measurement hash length mismatch")
	}
	for i := range measurementHash {
		if measurementHash[i] != identity.MeasurementHash[i] {
			return fmt.Errorf("SGX MRENCLAVE does not match registered measurement")
		}
	}

	// Verify measurement is in the allowlist
	if !k.IsMeasurementAllowed(ctx, identity.MeasurementHash, ctx.BlockHeight()) {
		return types.ErrMeasurementNotAllowlisted
	}

	return nil
}

// verifySEVSNPHeartbeatAttestation verifies an SEV-SNP attestation report for heartbeat
func (k Keeper) verifySEVSNPHeartbeatAttestation(ctx sdk.Context, identity types.EnclaveIdentity, attestationProof []byte) error {
	// Parse the SEV-SNP report
	report, err := types.ParseSEVSNPReport(attestationProof)
	if err != nil {
		return fmt.Errorf("failed to parse SEV-SNP report: %w", err)
	}

	// Verify debug mode is disabled in production
	if report.DebugEnabled() {
		return fmt.Errorf("SEV-SNP enclave is running in debug mode")
	}

	// Verify the measurement hash matches the registered identity
	measurementHash := types.SEVSNPMeasurementHash(report.Measurement[:])
	if len(identity.MeasurementHash) != len(measurementHash) {
		return fmt.Errorf("measurement hash length mismatch")
	}
	for i := range measurementHash {
		if measurementHash[i] != identity.MeasurementHash[i] {
			return fmt.Errorf("SEV-SNP measurement does not match registered measurement")
		}
	}

	// Verify measurement is in the allowlist
	if !k.IsMeasurementAllowed(ctx, identity.MeasurementHash, ctx.BlockHeight()) {
		return types.ErrMeasurementNotAllowlisted
	}

	return nil
}

// verifyNitroHeartbeatAttestation verifies an AWS Nitro attestation document for heartbeat
func (k Keeper) verifyNitroHeartbeatAttestation(ctx sdk.Context, identity types.EnclaveIdentity, attestationProof []byte) error {
	// Nitro attestation documents are CBOR-encoded COSE Sign1 structures
	// Minimum size check for a valid CBOR structure
	if len(attestationProof) < 100 {
		return fmt.Errorf("nitro attestation document too small")
	}

	// The attestation document contains PCRs (Platform Configuration Registers)
	// PCR0 contains the enclave image file measurement
	// For now, we verify the measurement hash is in the allowlist
	// Full COSE signature verification requires AWS Nitro root CA chain

	// Verify measurement is in the allowlist
	if !k.IsMeasurementAllowed(ctx, identity.MeasurementHash, ctx.BlockHeight()) {
		return types.ErrMeasurementNotAllowlisted
	}

	return nil
}

// CleanupExpiredNonces removes expired heartbeat nonces
func (k Keeper) CleanupExpiredNonces(ctx sdk.Context) {
	store := ctx.KVStore(k.StoreKey())
	prefix := types.HeartbeatNoncePrefixKey()
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
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

type heartbeatSigningData struct {
	ValidatorAddress    string `json:"validator_address"`
	TimestampUnixNano   int64  `json:"timestamp_unix_nano"`
	Nonce               uint64 `json:"nonce"`
	AttestationProofSHA string `json:"attestation_proof_sha256"`
}

func heartbeatSigningPayload(msg types.MsgEnclaveHeartbeat) ([]byte, error) {
	attestationHash := sha256.Sum256(msg.AttestationProof)
	dataBytes, err := json.Marshal(heartbeatSigningData{
		ValidatorAddress:    msg.ValidatorAddress,
		TimestampUnixNano:   msg.Timestamp.UTC().UnixNano(),
		Nonce:               msg.Nonce,
		AttestationProofSHA: hex.EncodeToString(attestationHash[:]),
	})
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(dataBytes)
	return hash[:], nil
}

func (k Keeper) heartbeatSigningKeys(ctx sdk.Context, validatorAddr sdk.AccAddress, identity types.EnclaveIdentity) (validKeys [][]byte, staleKeys [][]byte) {
	addUnique := func(dst [][]byte, candidate []byte) [][]byte {
		if len(candidate) == 0 {
			return dst
		}
		for _, existing := range dst {
			if bytes.Equal(existing, candidate) {
				return dst
			}
		}
		return append(dst, bytes.Clone(candidate))
	}

	validKeys = addUnique(validKeys, identity.SigningPubKey)

	latestRotation, exists := k.getLatestStoredKeyRotation(ctx, validatorAddr)
	if !exists {
		return validKeys, staleKeys
	}

	if latestRotation.Status == types.KeyRotationStatusActive && types.IsInOverlapPeriod(&latestRotation.KeyRotationRecord, ctx.BlockHeight()) {
		validKeys = addUnique(validKeys, latestRotation.OldSigningPubKey)
		validKeys = addUnique(validKeys, latestRotation.NewSigningPubKey)
		return validKeys, staleKeys
	}

	staleKeys = addUnique(staleKeys, latestRotation.OldSigningPubKey)
	return validKeys, staleKeys
}
