package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Verification Pipeline Configuration
// ============================================================================

// VerificationPipelineConfig holds configuration for the verification pipeline
type VerificationPipelineConfig struct {
	// MaxVerificationTimePerBlock is the max time for all verifications per block (ms)
	MaxVerificationTimePerBlock int64

	// MaxVerificationTimePerRequest is the max time for a single verification (ms)
	MaxVerificationTimePerRequest int64

	// MaxRequestsPerBlock is the maximum verification requests to process per block
	MaxRequestsPerBlock int

	// MaxRetries is the maximum number of retries for failed verifications
	MaxRetries uint32

	// RetryDelayBlocks is the number of blocks to wait before retrying
	RetryDelayBlocks int64
}

// DefaultVerificationPipelineConfig returns default pipeline configuration
func DefaultVerificationPipelineConfig() VerificationPipelineConfig {
	return VerificationPipelineConfig{
		MaxVerificationTimePerBlock:   2000, // 2 seconds per block
		MaxVerificationTimePerRequest: 500,  // 500ms per request
		MaxRequestsPerBlock:           10,   // Max 10 requests per block
		MaxRetries:                    3,    // 3 retry attempts
		RetryDelayBlocks:              5,    // Wait 5 blocks before retry
	}
}

// ============================================================================
// Verification Request Storage
// ============================================================================

// CreateVerificationRequest creates a new verification request
func (k Keeper) CreateVerificationRequest(
	ctx sdk.Context,
	accountAddress string,
	scopeIDs []string,
) (*types.VerificationRequest, error) {
	// Validate account address
	addr, err := sdk.AccAddressFromBech32(accountAddress)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(err.Error())
	}

	// Verify account has an identity record
	if _, found := k.GetIdentityRecord(ctx, addr); !found {
		return nil, types.ErrIdentityRecordNotFound.Wrapf("no identity record for %s", accountAddress)
	}

	// Validate scope IDs exist
	for _, scopeID := range scopeIDs {
		if _, found := k.GetScope(ctx, addr, scopeID); !found {
			return nil, types.ErrScopeNotFound.Wrapf("scope %s not found", scopeID)
		}
	}

	// Generate request ID
	requestID := k.generateRequestID(ctx, accountAddress, scopeIDs)

	profileSnapshot, err := k.activeInferenceProfileSnapshot(ctx)
	if err != nil {
		return nil, err
	}

	// Create the request
	request := types.NewVerificationRequest(
		requestID,
		accountAddress,
		scopeIDs,
		ctx.BlockTime(),
		ctx.BlockHeight(),
	)
	if err := request.SetInferenceProfileSnapshot(profileSnapshot); err != nil {
		return nil, err
	}

	// Store the request
	if err := k.setVerificationRequest(ctx, request); err != nil {
		return nil, err
	}

	// Add to pending queue
	k.addToPendingQueue(ctx, request)

	k.Logger(ctx).Info("verification request created",
		"request_id", requestID,
		"account", accountAddress,
		"scopes", len(scopeIDs),
	)

	return request, nil
}

// GetVerificationRequest retrieves a verification request by ID
func (k Keeper) GetVerificationRequest(ctx sdk.Context, requestID string) (*types.VerificationRequest, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.VerificationRequestKey(requestID))
	if bz == nil {
		return nil, false
	}

	var request types.VerificationRequest
	if err := json.Unmarshal(bz, &request); err != nil {
		k.Logger(ctx).Error("failed to unmarshal verification request", "error", err)
		return nil, false
	}

	return &request, true
}

// GetPendingRequests returns all pending verification requests up to limit
func (k Keeper) GetPendingRequests(ctx sdk.Context, limit int) []types.VerificationRequest {
	store := ctx.KVStore(k.skey)
	iterator := storetypes.KVStorePrefixIterator(store, types.PendingVerificationRequestPrefixKey())
	defer iterator.Close()

	requests := make([]types.VerificationRequest, 0, limit)
	count := 0

	for ; iterator.Valid() && count < limit; iterator.Next() {
		// Key format: prefix | block_height | "/" | request_id
		key := iterator.Key()
		requestID := extractRequestIDFromQueueKey(key)

		request, found := k.GetVerificationRequest(ctx, requestID)
		if found && request.Status == types.RequestStatusPending {
			requests = append(requests, *request)
			count++
		}
	}

	return requests
}

// GetVerificationRequestsByAccount returns verification requests for an account
func (k Keeper) GetVerificationRequestsByAccount(ctx sdk.Context, accountAddress string) []types.VerificationRequest {
	store := ctx.KVStore(k.skey)
	key := types.VerificationRequestByAccountKey(accountAddress)
	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	requestIDs := make([]string, 0, 1)
	if err := json.Unmarshal(bz, &requestIDs); err != nil {
		return nil
	}

	requests := make([]types.VerificationRequest, 0, len(requestIDs))
	for _, id := range requestIDs {
		if request, found := k.GetVerificationRequest(ctx, id); found {
			requests = append(requests, *request)
		}
	}

	return requests
}

// setVerificationRequest stores a verification request
func (k Keeper) setVerificationRequest(ctx sdk.Context, request *types.VerificationRequest) error {
	store := ctx.KVStore(k.skey)

	bz, err := json.Marshal(request)
	if err != nil {
		return err
	}

	store.Set(types.VerificationRequestKey(request.RequestID), bz)

	// Update account index
	k.updateAccountRequestIndex(ctx, request.AccountAddress, request.RequestID)

	return nil
}

// updateAccountRequestIndex updates the index of requests by account
func (k Keeper) updateAccountRequestIndex(ctx sdk.Context, accountAddress string, requestID string) {
	store := ctx.KVStore(k.skey)
	key := types.VerificationRequestByAccountKey(accountAddress)

	requestIDs := make([]string, 0, 1)
	if bz := store.Get(key); bz != nil {
		_ = json.Unmarshal(bz, &requestIDs)
	}

	// Add if not already present
	for _, id := range requestIDs {
		if id == requestID {
			return
		}
	}
	requestIDs = append(requestIDs, requestID)

	bz, _ := json.Marshal(requestIDs) //nolint:errchkjson // string slice cannot fail to marshal
	store.Set(key, bz)
}

// addToPendingQueue adds a request to the pending verification queue
func (k Keeper) addToPendingQueue(ctx sdk.Context, request *types.VerificationRequest) {
	store := ctx.KVStore(k.skey)
	key := types.PendingVerificationRequestKey(request.RequestedBlock, request.RequestID)
	store.Set(key, []byte{1}) // Value doesn't matter, just needs to exist
}

// removeFromPendingQueue removes a request from the pending queue
func (k Keeper) removeFromPendingQueue(ctx sdk.Context, request *types.VerificationRequest) {
	store := ctx.KVStore(k.skey)
	key := types.PendingVerificationRequestKey(request.RequestedBlock, request.RequestID)
	store.Delete(key)
}

// generateRequestID generates a unique request ID
func (k Keeper) generateRequestID(ctx sdk.Context, accountAddress string, scopeIDs []string) string {
	h := sha256.New()
	h.Write([]byte(accountAddress))
	for _, id := range scopeIDs {
		h.Write([]byte(id))
	}
	fmt.Fprintf(h, "%d", ctx.BlockHeight())
	h.Write([]byte(ctx.BlockTime().String()))
	return hex.EncodeToString(h.Sum(nil)[:16]) // 32 char hex string
}

// extractRequestIDFromQueueKey extracts the request ID from a pending queue key
func extractRequestIDFromQueueKey(key []byte) string {
	// Key format: prefix (1 byte) | block_height (8 bytes) | "/" (1 byte) | request_id
	if len(key) <= 10 {
		return ""
	}
	return string(key[10:])
}

// ============================================================================
// Verification Result Storage
// ============================================================================

// StoreVerificationResult stores a verification result
func (k Keeper) StoreVerificationResult(ctx sdk.Context, result *types.VerificationResult) error {
	if err := result.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)

	bz, err := json.Marshal(result)
	if err != nil {
		return err
	}

	store.Set(types.VerificationResultKey(result.RequestID), bz)

	// Update account result index
	indexKey := types.VerificationResultByAccountKey(result.AccountAddress, result.BlockHeight)
	store.Set(indexKey, []byte(result.RequestID))

	return nil
}

// GetVerificationResult retrieves a verification result by request ID
func (k Keeper) GetVerificationResult(ctx sdk.Context, requestID string) (*types.VerificationResult, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.VerificationResultKey(requestID))
	if bz == nil {
		return nil, false
	}

	var result types.VerificationResult
	if err := json.Unmarshal(bz, &result); err != nil {
		k.Logger(ctx).Error("failed to unmarshal verification result", "error", err)
		return nil, false
	}

	return &result, true
}

// ============================================================================
// Verification Processing
// ============================================================================

// ProcessVerificationRequest processes a single verification request
func (k Keeper) ProcessVerificationRequest(
	ctx sdk.Context,
	request *types.VerificationRequest,
	keyProvider ValidatorKeyProvider,
) *types.VerificationResult {
	if request == nil {
		result := types.NewVerificationResult("", "", ctx.BlockTime(), ctx.BlockHeight())
		result.SetError(types.ReasonCodeMLInferenceError, "verification request is required")
		return result
	}
	result := types.NewVerificationResult(
		request.RequestID,
		request.AccountAddress,
		ctx.BlockTime(),
		ctx.BlockHeight(),
	)
	result.SetError(types.ReasonCodeMLInferenceError, "direct verification processing is disabled; signed inference receipt staging is required")
	return result
}

// applyVerificationResult applies the verification result to on-chain state
func (k Keeper) applyVerificationResult(
	ctx sdk.Context,
	addr sdk.AccAddress,
	request *types.VerificationRequest,
	result *types.VerificationResult,
) error {
	if ctx.ExecMode() != sdk.ExecModeFinalize {
		return types.ErrUnauthorized.Wrap("verification results may only be applied during FinalizeBlock")
	}
	if k.consensusSystemTxAuthorizer == nil || !k.consensusSystemTxAuthorizer(ctx) {
		return types.ErrUnauthorized.Wrap("verification result application requires authorized consensus system transaction")
	}
	// Update identity score
	if result.Status == types.VerificationResultStatusSuccess ||
		result.Status == types.VerificationResultStatusPartial {
		if err := k.enforceVerificationArtifactState(ctx, request, result); err != nil {
			return err
		}

		// Determine account status based on score
		var accountStatus types.AccountStatus
		if result.Score >= types.ThresholdBasic {
			accountStatus = types.AccountStatusVerified
		} else {
			accountStatus = types.AccountStatusNeedsAdditionalFactor
		}

		// Set the score with details
		details := ScoreDetails{
			Status:           accountStatus,
			ModelVersion:     result.ModelVersion,
			VerificationHash: result.InputHash,
		}

		if err := k.SetScoreWithDetails(ctx, addr.String(), result.Score, details); err != nil {
			return err
		}

		// Update scope verification statuses
		for _, sr := range result.ScopeResults {
			if sr.Success {
				_ = k.UpdateVerificationStatus(
					ctx, addr, sr.ScopeID,
					types.VerificationStatusVerified,
					"verified via ML scoring",
					result.ValidatorAddress,
				)
			}
		}

		if k.issuancePolicyKeeper != nil {
			verifierID := result.ModelVersion
			activeVerifier, hasActiveVerifier, err := k.getActiveVerifierInfo(ctx)
			if err != nil {
				return err
			}
			if hasActiveVerifier && activeVerifier.VerifierID != "" {
				verifierID = activeVerifier.VerifierID
			}

			recordedUnits, err := k.issuancePolicyKeeper.RecordVerifiedProof(
				ctx,
				request.RequestID,
				addr.String(),
				verifierID,
				result.ModelVersion,
				result.Score,
			)
			if err != nil {
				k.Logger(ctx).Error("failed to record issuance proof", "request_id", request.RequestID, "error", err)
			} else {
				result.Metadata["issuance_recorded_units"] = fmt.Sprintf("%d", recordedUnits)
			}
		}
	}

	return nil
}

func (k Keeper) enforceVerificationArtifactState(
	ctx sdk.Context,
	request *types.VerificationRequest,
	result *types.VerificationResult,
) error {
	activeVerifier, hasActiveVerifier, err := k.getActiveVerifierInfo(ctx)
	if err != nil {
		return err
	}
	if hasActiveVerifier {
		result.Metadata["active_verifier_id"] = activeVerifier.VerifierID
		result.Metadata["active_verifier_spec_version"] = activeVerifier.SpecVersion
		if activeVerifier.ActivationHeight > 0 && request.RequestedBlock < activeVerifier.ActivationHeight {
			return types.ErrPipelineVersionMismatch.Wrapf(
				"active verifier %s is not authorized until height %d",
				activeVerifier.VerifierID,
				activeVerifier.ActivationHeight,
			)
		}
		if !versionsMatch(activeVerifier.SpecVersion, result.ModelVersion) {
			return types.ErrPipelineVersionMismatch.Wrapf(
				"active verifier %s expects %s, got %s",
				activeVerifier.VerifierID,
				activeVerifier.SpecVersion,
				result.ModelVersion,
			)
		}
	}

	activePV, err := k.GetActivePipelineVersion(ctx)
	if err != nil {
		if hasActiveVerifier {
			return types.ErrUnauthorized.Wrap("missing active pipeline artifact state for verifier enforcement")
		}
		return err
	}

	manifest, err := k.ensurePipelineVersionUsable(ctx, activePV)
	if err != nil {
		return err
	}

	result.Metadata["pipeline_version"] = activePV.Version
	result.Metadata["pipeline_image_hash"] = activePV.ImageHash
	result.Metadata["pipeline_manifest_hash"] = manifest.ManifestHash

	if !versionsMatch(activePV.Version, result.ModelVersion) {
		return types.ErrPipelineVersionMismatch.Wrapf(
			"active pipeline expects %s, got %s",
			activePV.Version,
			result.ModelVersion,
		)
	}

	if hasActiveVerifier && activeVerifier.WeightsSHA256 != "" &&
		!hashesMatch(activeVerifier.WeightsSHA256, manifest.ManifestHash) &&
		!hashesMatch(activeVerifier.WeightsSHA256, activePV.ImageHash) {
		return types.ErrUnauthorized.Wrapf(
			"active verifier %s expects artifact %s but pipeline publishes manifest %s and image %s",
			activeVerifier.VerifierID,
			activeVerifier.WeightsSHA256,
			manifest.ManifestHash,
			activePV.ImageHash,
		)
	}

	return nil
}

func (k Keeper) rejectVerificationResult(result *types.VerificationResult, err error) {
	result.Metadata["verification_rejection"] = err.Error()

	switch {
	case errorsmod.IsOf(err, types.ErrUnauthorized):
		result.SetFailed(types.ReasonCodeUnauthorizedArtifactState)
	case errorsmod.IsOf(err, types.ErrPipelineVersionMismatch),
		errorsmod.IsOf(err, types.ErrNoPipelineVersionActive),
		errorsmod.IsOf(err, types.ErrModelManifestMismatch):
		result.SetFailed(types.ReasonCodeStaleArtifactState)
	default:
		result.SetError(types.ReasonCodeMLInferenceError, err.Error())
	}
}

func hashesMatch(expected, actual string) bool {
	return normalizeHashString(expected) == normalizeHashString(actual)
}

func normalizeHashString(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	return strings.TrimPrefix(normalized, "sha256:")
}

func resultHasReasonCode(result *types.VerificationResult, code types.ReasonCode) bool {
	for _, reasonCode := range result.ReasonCodes {
		if reasonCode == code {
			return true
		}
	}
	return false
}

// finalizeRequest completes the verification request lifecycle
func (k Keeper) finalizeRequest(
	ctx sdk.Context,
	request *types.VerificationRequest,
	result *types.VerificationResult,
) {
	if ctx.ExecMode() != sdk.ExecModeFinalize ||
		k.consensusSystemTxAuthorizer == nil ||
		!k.consensusSystemTxAuthorizer(ctx) {
		k.Logger(ctx).Error("refusing to finalize verification outside authorized consensus system transaction")
		return
	}
	// Update request status based on result
	switch result.Status {
	case types.VerificationResultStatusSuccess, types.VerificationResultStatusPartial:
		request.SetCompleted()
	case types.VerificationResultStatusFailed:
		if resultHasReasonCode(result, types.ReasonCodeStaleArtifactState) ||
			resultHasReasonCode(result, types.ReasonCodeUnauthorizedArtifactState) {
			request.SetRejected(fmt.Sprintf("%v", result.ReasonCodes))
		} else {
			request.SetFailed(fmt.Sprintf("%v", result.ReasonCodes))
		}
	case types.VerificationResultStatusError:
		// Check if we should retry
		config := DefaultVerificationPipelineConfig()
		if request.IsRetryable(config.MaxRetries) {
			request.IncrementRetry(ctx.BlockTime())
			request.Status = types.RequestStatusPending
			// Re-add to pending queue for retry
			k.addToPendingQueue(ctx, request)
		} else {
			request.SetFailed("max retries exceeded")
			request.Metadata["final_error"] = fmt.Sprintf("%v", result.ReasonCodes)
		}
	}

	// Store updated request
	if err := k.setVerificationRequest(ctx, request); err != nil {
		k.Logger(ctx).Error("failed to finalize request", "error", err)
	}

	// Remove from pending queue if completed
	if types.IsFinalRequestStatus(request.Status) {
		k.removeFromPendingQueue(ctx, request)
	}

	// Store result
	if err := k.StoreVerificationResult(ctx, result); err != nil {
		k.Logger(ctx).Error("failed to store result", "error", err)
	}

	// Emit events
	_ = ctx.EventManager().EmitTypedEvent(&types.EventVerificationCompleted{
		RequestID:      request.RequestID,
		AccountAddress: request.AccountAddress,
		Score:          result.Score,
		Status:         string(result.Status),
		BlockHeight:    result.BlockHeight,
	})
}

// ProcessPendingVerifications processes all pending verification requests
// This is called during block processing by the proposer
func (k Keeper) ProcessPendingVerifications(
	ctx sdk.Context,
	keyProvider ValidatorKeyProvider,
) []types.VerificationResult {
	results := make([]types.VerificationResult, 0)
	if ctx.ExecMode() != sdk.ExecModeVoteExtension {
		return results
	}

	config := DefaultVerificationPipelineConfig()
	// Get pending requests
	requests := k.GetPendingRequests(ctx, config.MaxRequestsPerBlock)
	if len(requests) == 0 {
		return results
	}

	k.Logger(ctx).Info("processing verification requests",
		"count", len(requests),
		"block_height", ctx.BlockHeight(),
	)

	k.Logger(ctx).Debug("pending verification requests require signed inference receipts for staging",
		"count", len(requests),
		"block_height", ctx.BlockHeight(),
	)
	return results
}

// HandleVerificationTimeout handles requests that have timed out
func (k Keeper) HandleVerificationTimeout(
	ctx sdk.Context,
	request *types.VerificationRequest,
) *types.VerificationResult {
	if request != nil {
		k.Logger(ctx).Warn("verification request timeout left pending for consensus handling",
			"request_id", request.RequestID,
			"block_height", ctx.BlockHeight(),
		)
	}
	if request == nil {
		result := types.NewVerificationResult("invalid-timeout-request", "invalid-timeout-account", ctx.BlockTime(), ctx.BlockHeight())
		result.SetError(types.ReasonCodeTimeout, "verification timeout processing requires a request")
		return result
	}
	result := types.NewVerificationResult(
		request.RequestID,
		request.AccountAddress,
		ctx.BlockTime(),
		ctx.BlockHeight(),
	)
	result.SetError(types.ReasonCodeTimeout, "verification timeout requires consensus-system-transaction agreement")
	return result
}
