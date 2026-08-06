package keeper

import (
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Block Proposer Hook
// ============================================================================

// ProposerHookConfig holds configuration for the proposer hook
type ProposerHookConfig struct {
	// Enabled controls whether verification processing is enabled
	Enabled bool

	// KeyProviderFactory creates key providers for verification
	// This allows injection of different key sources (file, HSM, etc.)
	KeyProviderFactory func() (ValidatorKeyProvider, error)
}

// DefaultProposerHookConfig returns default proposer hook configuration
func DefaultProposerHookConfig() ProposerHookConfig {
	return ProposerHookConfig{
		Enabled:            true,
		KeyProviderFactory: nil, // Must be set by node configuration
	}
}

// proposerHookConfig holds the current configuration
// This is set during node initialization
var proposerHookConfig ProposerHookConfig

// SetProposerHookConfig sets the proposer hook configuration
// This should be called during node initialization
func SetProposerHookConfig(config ProposerHookConfig) {
	proposerHookConfig = config
}

// GetProposerHookConfig returns the current proposer hook configuration
func GetProposerHookConfig() ProposerHookConfig {
	return proposerHookConfig
}

// ============================================================================
// BeginBlocker Hook
// ============================================================================

// BeginBlocker is called at the beginning of every block
// The block proposer processes pending verification requests here
func (k Keeper) BeginBlocker(ctx sdk.Context) error {
	// Only process if enabled
	if !proposerHookConfig.Enabled {
		return nil
	}

	// Check if we are the block proposer
	if !k.isBlockProposer(ctx) {
		return nil
	}

	k.Logger(ctx).Debug("BeginBlocker verification processing disabled; signed receipt staging occurs in vote extensions",
		"block_height", ctx.BlockHeight(),
	)
	return nil
}

// EndBlocker is called at the end of every block
// Used for cleanup and timeout handling
func (k Keeper) EndBlocker(ctx sdk.Context) error {
	k.handleExpiredRequests(ctx)
	return nil
}

// isBlockProposer checks if the current node is the block proposer
func (k Keeper) isBlockProposer(ctx sdk.Context) bool {
	// Get the proposer address from the block header
	proposerAddr := ctx.BlockHeader().ProposerAddress

	if len(proposerAddr) == 0 {
		// No proposer address in header (shouldn't happen in normal operation)
		return false
	}

	// Check if we are the proposer
	// This requires comparing against the node's validator key
	// In production, this would check against the node's validator consensus key

	// For now, we use a simple approach: check if a proposer key is configured
	if proposerHookConfig.KeyProviderFactory != nil {
		// If we have a key provider, assume we might be the proposer
		// In production, this should verify the proposer address matches our validator
		k.Logger(ctx).Debug("proposer check",
			"proposer_address", hex.EncodeToString(proposerAddr),
			"block_height", ctx.BlockHeight(),
		)
		return true
	}

	return false
}

// handleExpiredRequests handles verification requests that have been pending too long
func (k Keeper) handleExpiredRequests(ctx sdk.Context) {
	config := DefaultVerificationPipelineConfig()

	// Get all pending requests
	requests := k.GetPendingRequests(ctx, config.MaxRequestsPerBlock*2)

	for _, request := range requests {
		// Check if request is too old
		blocksWaiting := ctx.BlockHeight() - request.RequestedBlock
		maxWaitBlocks := int64(config.MaxRetries+1) * config.RetryDelayBlocks * 2

		if blocksWaiting > maxWaitBlocks {
			k.Logger(ctx).Warn("verification request expired",
				"request_id", request.RequestID,
				"blocks_waiting", blocksWaiting,
			)
		}
	}
}

// ============================================================================
// Verification Request Trigger
// ============================================================================

// TriggerVerificationForScopes creates verification requests for accounts with pending scopes
// This is called when new scopes are uploaded
func (k Keeper) TriggerVerificationForScopes(
	ctx sdk.Context,
	accountAddress string,
	scopeIDs []string,
) error {
	// Check if there's already a pending request for these scopes
	if k.hasPendingRequestForScopes(ctx, accountAddress, scopeIDs) {
		return nil
	}

	// Create new verification request
	_, err := k.CreateVerificationRequest(ctx, accountAddress, scopeIDs)
	return err
}

// hasPendingRequestForScopes checks if any of the scope IDs already have a pending request
func (k Keeper) hasPendingRequestForScopes(
	ctx sdk.Context,
	accountAddress string,
	scopeIDs []string,
) bool {
	existingRequests := k.GetVerificationRequestsByAccount(ctx, accountAddress)

	// Build a set of new scope IDs for fast lookup
	newScopeSet := make(map[string]bool, len(scopeIDs))
	for _, id := range scopeIDs {
		newScopeSet[id] = true
	}

	for _, req := range existingRequests {
		if req.Status != types.RequestStatusPending && req.Status != types.RequestStatusInProgress {
			continue
		}
		for _, existingScopeID := range req.ScopeIDs {
			if newScopeSet[existingScopeID] {
				k.Logger(ctx).Debug("verification already pending for scope",
					"scope_id", existingScopeID,
					"request_id", req.RequestID,
				)
				return true
			}
		}
	}
	return false
}
