// Package keeper provides VEID module keeper implementation.
//
// This file implements the appeal system for contesting verification decisions.
//
// Task Reference: VE-3020 - Appeal and Dispute System
package keeper

import (
	"encoding/binary"
	"encoding/json"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Appeal Parameters
// ============================================================================

// GetAppealParams retrieves the appeal system parameters
func (k Keeper) GetAppealParams(ctx sdk.Context) types.AppealParams {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.AppealParamsKey())
	if bz == nil {
		return types.DefaultAppealParams()
	}

	var params types.AppealParams
	if err := json.Unmarshal(bz, &params); err != nil {
		return types.DefaultAppealParams()
	}
	return params
}

// SetAppealParams sets the appeal system parameters
func (k Keeper) SetAppealParams(ctx sdk.Context, params types.AppealParams) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(params)
	if err != nil {
		return err
	}
	store.Set(types.AppealParamsKey(), bz)
	return nil
}

// ============================================================================
// Authorized Resolvers
// ============================================================================

// IsAuthorizedResolver checks if an address can resolve appeals
func (k Keeper) IsAuthorizedResolver(ctx sdk.Context, address string) bool {
	// The authority (x/gov) is always authorized
	if address == k.authority {
		return true
	}

	// Check if address is a bonded validator
	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return false
	}
	if k.IsValidator(ctx, addr) {
		return true
	}

	// Check explicit authorization
	store := ctx.KVStore(k.skey)
	return store.Has(types.AuthorizedResolverKey(addr.Bytes()))
}

// SetAuthorizedResolver adds or removes an authorized resolver
func (k Keeper) SetAuthorizedResolver(ctx sdk.Context, address sdk.AccAddress, authorized bool) {
	store := ctx.KVStore(k.skey)
	key := types.AuthorizedResolverKey(address.Bytes())
	if authorized {
		store.Set(key, []byte{1})
	} else {
		store.Delete(key)
	}
}

// GetAuthorizedResolvers returns all authorized resolvers
func (k Keeper) GetAuthorizedResolvers(ctx sdk.Context) []string {
	store := ctx.KVStore(k.skey)
	iterator := storetypes.KVStorePrefixIterator(store, types.AuthorizedResolverPrefixKey())
	defer iterator.Close()

	var resolvers []string
	for ; iterator.Valid(); iterator.Next() {
		// Extract address from key
		key := iterator.Key()
		addressBytes := key[len(types.AuthorizedResolverPrefixKey()):]
		resolvers = append(resolvers, sdk.AccAddress(addressBytes).String())
	}
	return resolvers
}

// ============================================================================
// Appeal Storage
// ============================================================================

// SetAppeal stores an appeal record
func (k Keeper) SetAppeal(ctx sdk.Context, appeal *types.AppealRecord) error {
	if err := appeal.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)

	bz, err := json.Marshal(appeal)
	if err != nil {
		return err
	}

	// Store appeal record
	store.Set(types.AppealKey(appeal.AppealID), bz)

	// Update index by account
	address, _ := sdk.AccAddressFromBech32(appeal.AccountAddress)
	store.Set(types.AppealByAccountKey(address.Bytes(), appeal.AppealID), []byte{1})

	// Update index by scope
	store.Set(types.AppealByScopeKey(address.Bytes(), appeal.ScopeID, appeal.AppealNumber), []byte(appeal.AppealID))

	// Add to pending queue if pending
	if appeal.Status == types.AppealStatusPending {
		store.Set(types.PendingAppealKey(appeal.SubmittedAt, appeal.AppealID), []byte{1})
	}

	return nil
}

// GetAppeal retrieves an appeal by ID
func (k Keeper) GetAppeal(ctx sdk.Context, appealID string) (*types.AppealRecord, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.AppealKey(appealID))
	if bz == nil {
		return nil, false
	}

	var appeal types.AppealRecord
	if err := json.Unmarshal(bz, &appeal); err != nil {
		return nil, false
	}
	return &appeal, true
}

// DeleteAppeal removes an appeal from all indexes
func (k Keeper) DeleteAppeal(ctx sdk.Context, appeal *types.AppealRecord) {
	store := ctx.KVStore(k.skey)

	// Remove from pending queue if it was pending
	if appeal.Status == types.AppealStatusPending {
		store.Delete(types.PendingAppealKey(appeal.SubmittedAt, appeal.AppealID))
	}

	// Remove from account index
	address, _ := sdk.AccAddressFromBech32(appeal.AccountAddress)
	store.Delete(types.AppealByAccountKey(address.Bytes(), appeal.AppealID))

	// Remove from scope index
	store.Delete(types.AppealByScopeKey(address.Bytes(), appeal.ScopeID, appeal.AppealNumber))

	// Remove main record
	store.Delete(types.AppealKey(appeal.AppealID))
}

// GetAppealsByAccount retrieves all appeals for an account
func (k Keeper) GetAppealsByAccount(ctx sdk.Context, address string) []*types.AppealRecord {
	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil
	}

	store := ctx.KVStore(k.skey)
	iterator := storetypes.KVStorePrefixIterator(store, types.AppealByAccountPrefixKey(addr.Bytes()))
	defer iterator.Close()

	var appeals []*types.AppealRecord
	for ; iterator.Valid(); iterator.Next() {
		// Extract appeal ID from key
		key := iterator.Key()
		prefix := types.AppealByAccountPrefixKey(addr.Bytes())
		appealID := string(key[len(prefix):])

		appeal, found := k.GetAppeal(ctx, appealID)
		if found {
			appeals = append(appeals, appeal)
		}
	}
	return appeals
}

// GetAppealsByScope retrieves all appeals for a specific scope
func (k Keeper) GetAppealsByScope(ctx sdk.Context, address string, scopeID string) []*types.AppealRecord {
	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil
	}

	store := ctx.KVStore(k.skey)
	iterator := storetypes.KVStorePrefixIterator(store, types.AppealByScopePrefixKey(addr.Bytes(), scopeID))
	defer iterator.Close()

	var appeals []*types.AppealRecord
	for ; iterator.Valid(); iterator.Next() {
		appealID := string(iterator.Value())
		appeal, found := k.GetAppeal(ctx, appealID)
		if found {
			appeals = append(appeals, appeal)
		}
	}
	return appeals
}

// GetPendingAppeals retrieves all pending appeals (for reviewers)
func (k Keeper) GetPendingAppeals(ctx sdk.Context) []*types.AppealRecord {
	store := ctx.KVStore(k.skey)
	iterator := storetypes.KVStorePrefixIterator(store, types.PendingAppealPrefixKey())
	defer iterator.Close()

	var appeals []*types.AppealRecord
	for ; iterator.Valid(); iterator.Next() {
		// Extract appeal ID from key
		key := iterator.Key()
		// Key format: prefix (1) + timestamp (8) + '/' (1) + appeal_id
		if len(key) <= 10 {
			continue
		}
		appealID := string(key[10:])

		appeal, found := k.GetAppeal(ctx, appealID)
		if found && appeal.Status == types.AppealStatusPending {
			appeals = append(appeals, appeal)
		}
	}
	return appeals
}

// ============================================================================
// Appeal Count Management
// ============================================================================

// GetAppealCountForScope returns the number of appeals for a scope
func (k Keeper) GetAppealCountForScope(ctx sdk.Context, address sdk.AccAddress, scopeID string) uint32 {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.AppealScopeCountKey(address.Bytes(), scopeID))
	if bz == nil {
		return 0
	}
	return binary.BigEndian.Uint32(bz)
}

// IncrementAppealCountForScope increments the appeal count for a scope
func (k Keeper) IncrementAppealCountForScope(ctx sdk.Context, address sdk.AccAddress, scopeID string) uint32 {
	count := k.GetAppealCountForScope(ctx, address, scopeID) + 1
	store := ctx.KVStore(k.skey)
	bz := make([]byte, 4)
	binary.BigEndian.PutUint32(bz, count)
	store.Set(types.AppealScopeCountKey(address.Bytes(), scopeID), bz)
	return count
}

// ============================================================================
// Appeal Operations
// ============================================================================

// SubmitAppeal creates a new appeal for a verification decision
func (k Keeper) SubmitAppeal(ctx sdk.Context, msg *types.MsgSubmitAppeal) (*types.AppealRecord, error) {
	params := k.GetAppealParams(ctx)

	// Check if appeal system is enabled
	if !params.Enabled {
		return nil, types.ErrAppealSystemDisabled
	}

	submitter, err := sdk.AccAddressFromBech32(msg.Submitter)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid submitter address")
	}

	// Verify the scope exists and belongs to the submitter
	scope, found := k.GetScope(ctx, submitter, msg.ScopeId)
	if !found {
		return nil, types.ErrScopeNotFound.Wrapf("scope %s not found", msg.ScopeId)
	}

	// Verify the scope was rejected (can only appeal rejections)
	if scope.Status != types.VerificationStatusRejected {
		return nil, types.ErrScopeNotRejected.Wrapf("scope status is %s, not rejected", string(scope.Status))
	}

	// Check appeal window (scope must have been rejected recently)
	// Use the time-based approach to estimate when the rejection occurred
	// Calculate approximate rejection block based on scope upload time and current block time
	var rejectionBlock int64
	if scope.UploadedAt.Before(ctx.BlockTime().Add(-time.Hour * 24 * 30)) {
		// If uploaded more than 30 days ago, rejection was definitely outside window
		rejectionBlock = 0
	} else {
		// Calculate blocks elapsed since upload (rough estimate)
		elapsedSeconds := ctx.BlockTime().Sub(scope.UploadedAt).Seconds()
		blocksElapsed := int64(elapsedSeconds / 6) // Assume ~6 second blocks
		rejectionBlock = ctx.BlockHeight() - blocksElapsed
		if rejectionBlock < 0 {
			rejectionBlock = 0
		}
	}

	if !params.IsWithinAppealWindow(rejectionBlock, ctx.BlockHeight()) {
		return nil, types.ErrAppealWindowExpired.Wrapf(
			"appeal window expired at block %d, current block is %d",
			rejectionBlock+params.AppealWindowBlocks,
			ctx.BlockHeight(),
		)
	}

	// Check max appeals per scope
	currentCount := k.GetAppealCountForScope(ctx, submitter, msg.ScopeId)
	if currentCount >= params.MaxAppealsPerScope {
		return nil, types.ErrMaxAppealsExceeded.Wrapf(
			"scope has %d appeals, maximum is %d",
			currentCount,
			params.MaxAppealsPerScope,
		)
	}

	// Check for existing pending appeal for this scope
	existingAppeals := k.GetAppealsByScope(ctx, msg.Submitter, msg.ScopeId)
	for _, existing := range existingAppeals {
		if existing.Status.IsActive() {
			return nil, types.ErrAppealAlreadyExists.Wrapf(
				"appeal %s is already pending for this scope",
				existing.AppealID,
			)
		}
	}

	// Increment appeal count and get appeal number
	appealNumber := k.IncrementAppealCountForScope(ctx, submitter, msg.ScopeId)

	// Generate appeal ID
	appealID := types.GenerateAppealID(msg.Submitter, msg.ScopeId, ctx.BlockHeight())

	// Get current score for the account
	var originalScore uint32
	score, found := k.GetIdentityScore(ctx, submitter.String())
	if found {
		originalScore = score.Score
	}

	// Create appeal record
	appeal := types.NewAppealRecord(
		appealID,
		msg.Submitter,
		msg.ScopeId,
		string(scope.Status),
		originalScore,
		msg.Reason,
		msg.EvidenceHashes,
		ctx.BlockHeight(),
		ctx.BlockTime(),
		appealNumber,
	)

	// Store the appeal
	if err := k.SetAppeal(ctx, appeal); err != nil {
		return nil, err
	}

	// Emit event
	if err := ctx.EventManager().EmitTypedEvent(&types.EventAppealSubmitted{
		AppealID:       appealID,
		AccountAddress: msg.Submitter,
		ScopeID:        msg.ScopeId,
		OriginalStatus: string(scope.Status),
		AppealNumber:   appealNumber,
		EvidenceCount:  len(msg.EvidenceHashes),
		SubmittedAt:    ctx.BlockHeight(),
	}); err != nil {
		k.Logger(ctx).Error("failed to emit appeal submitted event", "error", err)
	}

	return appeal, nil
}

// ClaimAppeal allows an arbitrator to claim an appeal for review
func (k Keeper) ClaimAppeal(ctx sdk.Context, msg *types.MsgClaimAppeal) error {
	// Check if resolver is authorized
	if !k.IsAuthorizedResolver(ctx, msg.Reviewer) {
		return types.ErrNotAuthorizedResolver.Wrapf("address %s is not authorized", msg.Reviewer)
	}

	appeal, found := k.GetAppeal(ctx, msg.AppealId)
	if !found {
		return types.ErrAppealNotFound.Wrapf("appeal %s not found", msg.AppealId)
	}

	// Set appeal to reviewing status
	if err := appeal.SetReviewing(msg.Reviewer, ctx.BlockHeight()); err != nil {
		return err
	}

	// Remove from pending queue
	store := ctx.KVStore(k.skey)
	store.Delete(types.PendingAppealKey(appeal.SubmittedAt, appeal.AppealID))

	// Update appeal record
	if err := k.SetAppeal(ctx, appeal); err != nil {
		return err
	}

	// Emit event
	if err := ctx.EventManager().EmitTypedEvent(&types.EventAppealClaimed{
		AppealID:        msg.AppealId,
		ReviewerAddress: msg.Reviewer,
		ClaimedAt:       ctx.BlockHeight(),
	}); err != nil {
		k.Logger(ctx).Error("failed to emit appeal claimed event", "error", err)
	}

	return nil
}

// ResolveAppeal resolves an appeal (approve/reject)
func (k Keeper) ResolveAppeal(ctx sdk.Context, msg *types.MsgResolveAppeal) error {
	// Check if resolver is authorized
	if !k.IsAuthorizedResolver(ctx, msg.Resolver) {
		return types.ErrNotAuthorizedResolver.Wrapf("address %s is not authorized", msg.Resolver)
	}

	appeal, found := k.GetAppeal(ctx, msg.AppealId)
	if !found {
		return types.ErrAppealNotFound.Wrapf("appeal %s not found", msg.AppealId)
	}

	// If appeal is in reviewing status, verify the resolver is the one who claimed it
	if appeal.Status == types.AppealStatusReviewing && appeal.ReviewerAddress != msg.Resolver {
		// Allow x/gov to override
		if msg.Resolver != k.authority {
			return types.ErrNotAppealReviewer.Wrapf(
				"appeal was claimed by %s, not %s",
				appeal.ReviewerAddress,
				msg.Resolver,
			)
		}
	}

	// Resolve the appeal
	if err := appeal.Resolve(
		types.AppealStatusFromProto(msg.Resolution),
		msg.Reason,
		msg.ScoreAdjustment,
		ctx.BlockHeight(),
		ctx.BlockTime(),
	); err != nil {
		return err
	}

	// If not already set, set the reviewer
	if appeal.ReviewerAddress == "" {
		appeal.ReviewerAddress = msg.Resolver
	}

	// Remove from pending queue if it was still pending
	store := ctx.KVStore(k.skey)
	store.Delete(types.PendingAppealKey(appeal.SubmittedAt, appeal.AppealID))

	// Update appeal record
	if err := k.SetAppeal(ctx, appeal); err != nil {
		return err
	}

	// If approved and score adjustment is non-zero, apply score adjustment
	localResolution := types.AppealStatusFromProto(msg.Resolution)
	if localResolution == types.AppealStatusApproved && msg.ScoreAdjustment != 0 {
		if err := k.ApplyAppealScoreAdjustment(ctx, appeal); err != nil {
			k.Logger(ctx).Error("failed to apply score adjustment", "error", err, "appeal_id", msg.AppealId)
		}
	}

	// Emit event
	if err := ctx.EventManager().EmitTypedEvent(&types.EventAppealResolved{
		AppealID:         msg.AppealId,
		AccountAddress:   appeal.AccountAddress,
		ScopeID:          appeal.ScopeID,
		Resolution:       localResolution,
		ResolutionReason: msg.Reason,
		ScoreAdjustment:  msg.ScoreAdjustment,
		ReviewerAddress:  msg.Resolver,
		ResolvedAt:       ctx.BlockHeight(),
	}); err != nil {
		k.Logger(ctx).Error("failed to emit appeal resolved event", "error", err)
	}

	return nil
}

// WithdrawAppeal withdraws a pending appeal
func (k Keeper) WithdrawAppeal(ctx sdk.Context, msg *types.MsgWithdrawAppeal) error {
	appeal, found := k.GetAppeal(ctx, msg.AppealId)
	if !found {
		return types.ErrAppealNotFound.Wrapf("appeal %s not found", msg.AppealId)
	}

	// Verify the submitter owns the appeal
	if appeal.AccountAddress != msg.Submitter {
		return types.ErrNotAppealSubmitter.Wrapf(
			"appeal belongs to %s, not %s",
			appeal.AccountAddress,
			msg.Submitter,
		)
	}

	// Withdraw the appeal
	if err := appeal.Withdraw(); err != nil {
		return err
	}

	// Remove from pending queue
	store := ctx.KVStore(k.skey)
	store.Delete(types.PendingAppealKey(appeal.SubmittedAt, appeal.AppealID))

	// Update appeal record
	if err := k.SetAppeal(ctx, appeal); err != nil {
		return err
	}

	// Emit event
	if err := ctx.EventManager().EmitTypedEvent(&types.EventAppealWithdrawn{
		AppealID:       msg.AppealId,
		AccountAddress: appeal.AccountAddress,
		ScopeID:        appeal.ScopeID,
		WithdrawnAt:    ctx.BlockHeight(),
	}); err != nil {
		k.Logger(ctx).Error("failed to emit appeal withdrawn event", "error", err)
	}

	return nil
}

// ApplyAppealScoreAdjustment applies the score adjustment from an approved appeal
func (k Keeper) ApplyAppealScoreAdjustment(ctx sdk.Context, appeal *types.AppealRecord) error {
	if appeal.ScoreAdjustment == 0 {
		return nil
	}

	address, err := sdk.AccAddressFromBech32(appeal.AccountAddress)
	if err != nil {
		return err
	}

	score, found := k.GetIdentityScore(ctx, address.String())
	if !found {
		return nil // No score to adjust
	}

	originalScore := score.Score
	newScore := int64(originalScore) + int64(appeal.ScoreAdjustment)

	// Clamp to valid range
	if newScore < 0 {
		newScore = 0
	}
	if newScore > int64(types.MaxScore) {
		newScore = int64(types.MaxScore)
	}

	// Update the score in the score store (use SetScore, not UpdateScore)
	if err := k.SetScore(ctx, address.String(), uint32(newScore), "appeal_adjustment"); err != nil {
		return err
	}

	// Emit score adjustment event
	if err := ctx.EventManager().EmitTypedEvent(&types.EventAppealScoreAdjusted{
		AppealID:        appeal.AppealID,
		AccountAddress:  appeal.AccountAddress,
		OriginalScore:   originalScore,
		NewScore:        uint32(newScore),
		ScoreAdjustment: appeal.ScoreAdjustment,
		AdjustedAt:      ctx.BlockHeight(),
	}); err != nil {
		k.Logger(ctx).Error("failed to emit score adjusted event", "error", err)
	}

	return nil
}

// ExpireStaleAppeals expires appeals that have been in reviewing status too long
func (k Keeper) ExpireStaleAppeals(ctx sdk.Context) int {
	params := k.GetAppealParams(ctx)
	store := ctx.KVStore(k.skey)

	var expiredCount int

	// Iterate through all pending appeals to check for timeouts
	iterator := storetypes.KVStorePrefixIterator(store, types.AppealPrefixKey())
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var appeal types.AppealRecord
		if err := json.Unmarshal(iterator.Value(), &appeal); err != nil {
			continue
		}

		// Check if appeal in reviewing status has timed out
		if appeal.Status == types.AppealStatusReviewing {
			if ctx.BlockHeight() > appeal.ClaimedAt+params.ReviewTimeoutBlocks {
				// Reset to pending status (release claim)
				appeal.Status = types.AppealStatusPending
				appeal.ReviewerAddress = ""
				appeal.ClaimedAt = 0

				// Re-add to pending queue
				store.Set(types.PendingAppealKey(appeal.SubmittedAt, appeal.AppealID), []byte{1})

				// Update appeal
				if bz, err := json.Marshal(appeal); err == nil {
					store.Set(types.AppealKey(appeal.AppealID), bz)
				}
			}
		}
	}

	return expiredCount
}

// GetAppealSummary returns a summary of appeals for an account
func (k Keeper) GetAppealSummary(ctx sdk.Context, address string) *types.AppealSummary {
	appeals := k.GetAppealsByAccount(ctx, address)
	summary := types.NewAppealSummary()
	for _, appeal := range appeals {
		summary.AddAppeal(appeal.Status)
	}
	return summary
}

// WithAppeals iterates through all appeals
func (k Keeper) WithAppeals(ctx sdk.Context, fn func(*types.AppealRecord) bool) {
	store := ctx.KVStore(k.skey)
	iterator := storetypes.KVStorePrefixIterator(store, types.AppealPrefixKey())
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var appeal types.AppealRecord
		if err := json.Unmarshal(iterator.Value(), &appeal); err != nil {
			continue
		}
		if fn(&appeal) {
			break
		}
	}
}
