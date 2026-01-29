package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Authorization Event Emission Helpers
// ============================================================================

// EmitAuthorizationGrantedEvent emits an EventAuthorizationGranted event
// when VEID grants authorization for a sensitive action (typically after MFA)
func (k Keeper) EmitAuthorizationGrantedEvent(
	ctx sdk.Context,
	account string,
	sessionID string,
	action string,
	factorsUsed []string,
	expiresAt int64,
) error {
	return ctx.EventManager().EmitTypedEvent(&types.EventAuthorizationGranted{
		Account:     account,
		SessionID:   sessionID,
		Action:      action,
		FactorsUsed: factorsUsed,
		ExpiresAt:   expiresAt,
		BlockHeight: ctx.BlockHeight(),
		Timestamp:   ctx.BlockTime().Unix(),
	})
}

// EmitAuthorizationConsumedEvent emits an EventAuthorizationConsumed event
// when a previously granted authorization is used
func (k Keeper) EmitAuthorizationConsumedEvent(
	ctx sdk.Context,
	account string,
	sessionID string,
	action string,
) error {
	return ctx.EventManager().EmitTypedEvent(&types.EventAuthorizationConsumed{
		Account:     account,
		SessionID:   sessionID,
		Action:      action,
		BlockHeight: ctx.BlockHeight(),
		Timestamp:   ctx.BlockTime().Unix(),
	})
}

// EmitAuthorizationExpiredEvent emits an EventAuthorizationExpired event
// when an authorization session expires without being used
func (k Keeper) EmitAuthorizationExpiredEvent(
	ctx sdk.Context,
	account string,
	sessionID string,
	action string,
	createdAt int64,
) error {
	return ctx.EventManager().EmitTypedEvent(&types.EventAuthorizationExpired{
		Account:     account,
		SessionID:   sessionID,
		Action:      action,
		CreatedAt:   createdAt,
		ExpiredAt:   ctx.BlockTime().Unix(),
		BlockHeight: ctx.BlockHeight(),
	})
}

// EmitVerificationSubmittedEvent emits an EventVerificationSubmitted event
// when a new verification is submitted
func (k Keeper) EmitVerificationSubmittedEvent(
	ctx sdk.Context,
	account string,
	scopeID string,
	scopeType string,
	requestID string,
) error {
	return ctx.EventManager().EmitTypedEvent(&types.EventVerificationSubmitted{
		Account:     account,
		ScopeID:     scopeID,
		ScopeType:   scopeType,
		RequestID:   requestID,
		BlockHeight: ctx.BlockHeight(),
		Timestamp:   ctx.BlockTime().Unix(),
	})
}

// EmitTierChangedEvent emits an EventTierChanged event
// when an account's tier changes due to score update
func (k Keeper) EmitTierChangedEvent(
	ctx sdk.Context,
	account string,
	oldTier string,
	newTier string,
	score uint32,
) error {
	return ctx.EventManager().EmitTypedEvent(&types.EventTierChanged{
		Account:     account,
		OldTier:     oldTier,
		NewTier:     newTier,
		Score:       score,
		BlockHeight: ctx.BlockHeight(),
		Timestamp:   ctx.BlockTime().Unix(),
	})
}
