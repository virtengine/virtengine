package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

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

	// Create the request
	request := types.NewVerificationRequest(
		requestID,
		accountAddress,
		scopeIDs,
		ctx.BlockTime(),
		ctx.BlockHeight(),
	)

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

	var requestIDs []string
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

	var requestIDs []string
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

	bz, _ := json.Marshal(requestIDs)
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
	h.Write([]byte(fmt.Sprintf("%d", ctx.BlockHeight())))
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
	startTime := time.Now()

	// Initialize result
	result := types.NewVerificationResult(
		request.RequestID,
		request.AccountAddress,
		ctx.BlockTime(),
		ctx.BlockHeight(),
	)

	// Parse account address
	addr, err := sdk.AccAddressFromBech32(request.AccountAddress)
	if err != nil {
		result.SetError(types.ReasonCodeInvalidPayload, err.Error())
		return result
	}

	// Update request status to in progress
	request.SetInProgress(ctx.BlockTime())
	if err := k.setVerificationRequest(ctx, request); err != nil {
		k.Logger(ctx).Error("failed to update request status", "error", err)
	}

	// Step 1: Decrypt scopes
	decryptedScopes, scopeResults, err := k.DecryptScopesForVerification(
		ctx, addr, request.ScopeIDs, keyProvider)
	if err != nil {
		result.SetError(types.ReasonCodeDecryptError, err.Error())
		k.finalizeRequest(ctx, request, result)
		return result
	}

	// Step 2: Validate decrypted payloads
	validDecrypted := make([]DecryptedScope, 0, len(decryptedScopes))
	for i, ds := range decryptedScopes {
		valid, reason := k.ValidateDecryptedPayload(ctx, ds)
		if valid {
			validDecrypted = append(validDecrypted, ds)
		} else {
			scopeResults[i].SetFailure(types.ReasonCodeInvalidPayload)
			scopeResults[i].Details = reason
		}
	}

	// Step 3: Check if we have enough valid scopes
	if len(validDecrypted) == 0 {
		result.SetFailed(types.ReasonCodeInsufficientScopes)
		for _, sr := range scopeResults {
			result.AddScopeResult(sr)
		}
		k.finalizeRequest(ctx, request, result)
		return result
	}

	// Step 4: Compute identity score using ML
	score, modelVersion, reasonCodes, inputHash, err := k.ComputeIdentityScore(
		ctx, request.AccountAddress, validDecrypted, scopeResults)
	if err != nil {
		result.SetError(types.ReasonCodeMLInferenceError, err.Error())
		k.finalizeRequest(ctx, request, result)
		return result
	}

	// Step 5: Build final result
	result.Score = score
	result.ModelVersion = modelVersion
	result.InputHash = inputHash
	result.ProcessingDuration = time.Since(startTime).Milliseconds()

	// Add scope results
	for _, sr := range scopeResults {
		result.AddScopeResult(sr)
	}

	// Determine overall status
	result.DetermineStatus()

	// Override reason codes with ML reason codes if available
	if len(reasonCodes) > 0 {
		result.ReasonCodes = reasonCodes
	}

	// Step 6: Update on-chain state
	k.applyVerificationResult(ctx, addr, result)

	// Finalize request
	k.finalizeRequest(ctx, request, result)

	k.Logger(ctx).Info("verification request processed",
		"request_id", request.RequestID,
		"account", request.AccountAddress,
		"score", result.Score,
		"status", result.Status,
		"duration_ms", result.ProcessingDuration,
	)

	return result
}

// applyVerificationResult applies the verification result to on-chain state
func (k Keeper) applyVerificationResult(
	ctx sdk.Context,
	addr sdk.AccAddress,
	result *types.VerificationResult,
) {
	// Update identity score
	if result.Status == types.VerificationResultStatusSuccess ||
		result.Status == types.VerificationResultStatusPartial {

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
			k.Logger(ctx).Error("failed to set score", "error", err)
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
	}
}

// finalizeRequest completes the verification request lifecycle
func (k Keeper) finalizeRequest(
	ctx sdk.Context,
	request *types.VerificationRequest,
	result *types.VerificationResult,
) {
	// Update request status based on result
	switch result.Status {
	case types.VerificationResultStatusSuccess, types.VerificationResultStatusPartial:
		request.SetCompleted()
	case types.VerificationResultStatusFailed:
		request.SetFailed(fmt.Sprintf("%v", result.ReasonCodes))
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
	config := DefaultVerificationPipelineConfig()
	startTime := time.Now()
	results := make([]types.VerificationResult, 0)

	// Get pending requests
	requests := k.GetPendingRequests(ctx, config.MaxRequestsPerBlock)
	if len(requests) == 0 {
		return results
	}

	k.Logger(ctx).Info("processing verification requests",
		"count", len(requests),
		"block_height", ctx.BlockHeight(),
	)

	for _, request := range requests {
		// Check if we've exceeded block time budget
		elapsed := time.Since(startTime).Milliseconds()
		if elapsed >= config.MaxVerificationTimePerBlock {
			k.Logger(ctx).Warn("verification time budget exceeded",
				"elapsed_ms", elapsed,
				"processed", len(results),
				"remaining", len(requests)-len(results),
			)
			break
		}

		// Process the request
		result := k.ProcessVerificationRequest(ctx, &request, keyProvider)
		results = append(results, *result)
	}

	return results
}

// HandleVerificationTimeout handles requests that have timed out
func (k Keeper) HandleVerificationTimeout(
	ctx sdk.Context,
	request *types.VerificationRequest,
) *types.VerificationResult {
	config := DefaultVerificationPipelineConfig()

	result := types.NewVerificationResult(
		request.RequestID,
		request.AccountAddress,
		ctx.BlockTime(),
		ctx.BlockHeight(),
	)
	result.SetError(types.ReasonCodeTimeout, "verification timed out")

	// Check if we should retry
	if request.IsRetryable(config.MaxRetries) {
		request.IncrementRetry(ctx.BlockTime())
		request.SetTimeout()
		k.addToPendingQueue(ctx, request)
	} else {
		request.SetFailed("timeout after max retries")
	}

	_ = k.setVerificationRequest(ctx, request)
	_ = k.StoreVerificationResult(ctx, result)

	return result
}
