package handler

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"

	mkeeper "github.com/virtengine/virtengine/x/market/keeper"
	mfakeeper "github.com/virtengine/virtengine/x/mfa/keeper"
	mfatypes "github.com/virtengine/virtengine/x/mfa/types"
	"github.com/virtengine/virtengine/x/provider/keeper"
	veidkeeper "github.com/virtengine/virtengine/x/veid/keeper"
)

var (
	// ErrInternal defines registered error code for internal error
	ErrInternal = errorsmod.Register(types.ModuleName, 10, "internal error")
	// ErrInsufficientVEIDScore defines error when VEID score is below required threshold
	ErrInsufficientVEIDScore = errorsmod.Register(types.ModuleName, 11, "VEID score below required threshold for provider registration")
	// ErrMFARequired defines error when MFA is required but not provided
	ErrMFARequired = errorsmod.Register(types.ModuleName, 12, "MFA authorization required for provider registration")
)

type msgServer struct {
	provider keeper.IKeeper
	market   mkeeper.IKeeper
	veid     veidkeeper.IKeeper
	mfa      mfakeeper.IKeeper
}

// NewMsgServerImpl returns an implementation of the market MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(k keeper.IKeeper, mk mkeeper.IKeeper, vk veidkeeper.IKeeper, mfak mfakeeper.IKeeper) types.MsgServer {
	return &msgServer{
		provider: k,
		market:   mk,
		veid:     vk,
		mfa:      mfak,
	}
}

var _ types.MsgServer = msgServer{}

func (ms msgServer) CreateProvider(goCtx context.Context, msg *types.MsgCreateProvider) (*types.MsgCreateProviderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	owner, _ := sdk.AccAddressFromBech32(msg.Owner)

	if _, ok := ms.provider.Get(ctx, owner); ok {
		return nil, types.ErrProviderExists.Wrapf("id: %s", msg.Owner)
	}

	// MARKET-VEID-002: Check VEID score requirement (only if VEID keeper is configured)
	if ms.veid != nil {
		score, hasScore := ms.veid.GetVEIDScore(ctx, owner)
		if !hasScore || score < 70 {
			return nil, ErrInsufficientVEIDScore.Wrapf(
				"provider registration requires VEID score ≥70, current score: %d",
				score,
			)
		}
	}

	// MARKET-VEID-002: Verify MFA authorization for provider registration
	// This is a sensitive transaction type that requires MFA approval
	// Only check if MFA keeper is configured
	if ms.mfa == nil {
		// MFA not configured, skip MFA check
		if err := ms.provider.Create(ctx, types.Provider(*msg)); err != nil {
			return nil, ErrInternal.Wrapf("err: %v", err)
		}
		return &types.MsgCreateProviderResponse{}, nil
	}

	txType := mfatypes.SensitiveTxProviderRegistration
	config, found := ms.mfa.GetSensitiveTxConfig(ctx, txType)
	if found && config.Enabled {
		// Check if there's a valid authorization session
		// The MFA gating ante handler should have validated this, but we double-check here
		// as defense in depth for this critical operation
		sessions := ms.mfa.GetAccountSessions(ctx, owner)
		hasValidSession := false
		blockTime := ctx.BlockTime()

		for _, session := range sessions {
			// Session must be valid, not yet used, and for provider registration
			if session.IsValid(blockTime) && session.TransactionType == txType {
				hasValidSession = true
				break
			}
		}

		if !hasValidSession {
			return nil, ErrMFARequired.Wrap(
				"provider registration requires valid MFA authorization session",
			)
		}
	}

	if err := ms.provider.Create(ctx, types.Provider(*msg)); err != nil {
		return nil, ErrInternal.Wrapf("err: %v", err)
	}

	return &types.MsgCreateProviderResponse{}, nil
}

func (ms msgServer) UpdateProvider(goCtx context.Context, msg *types.MsgUpdateProvider) (*types.MsgUpdateProviderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}

	owner, _ := sdk.AccAddressFromBech32(msg.Owner)
	_, found := ms.provider.Get(ctx, owner)
	if !found {
		return nil, types.ErrProviderNotFound.Wrapf("id: %s", msg.Owner)
	}

	if err := ms.provider.Update(ctx, types.Provider(*msg)); err != nil {
		return nil, errorsmod.Wrapf(ErrInternal, "err: %v", err)
	}

	return &types.MsgUpdateProviderResponse{}, nil
}

func (ms msgServer) DeleteProvider(goCtx context.Context, msg *types.MsgDeleteProvider) (*types.MsgDeleteProviderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	// Verify provider exists
	if _, ok := ms.provider.Get(ctx, owner); !ok {
		return nil, types.ErrProviderNotFound
	}

	// Check if provider has active leases - cannot delete if leases exist
	if ms.market.ProviderHasActiveLeases(ctx, owner) {
		return nil, types.ErrProviderHasActiveLeases.Wrapf("provider %s has active leases", msg.Owner)
	}

	// Delete the provider (this also emits EventProviderDeleted)
	ms.provider.Delete(ctx, owner)

	// Also clean up provider public key if exists
	ms.provider.DeleteProviderPublicKey(ctx, owner)

	return &types.MsgDeleteProviderResponse{}, nil
}

func (ms msgServer) GenerateDomainVerificationToken(goCtx context.Context, msg *types.MsgGenerateDomainVerificationToken) (*types.MsgGenerateDomainVerificationTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	owner, _ := sdk.AccAddressFromBech32(msg.Owner)

	// Verify provider exists
	if _, ok := ms.provider.Get(ctx, owner); !ok {
		return nil, types.ErrProviderNotFound.Wrapf("provider not found: %s", msg.Owner)
	}

	// Generate verification token
	record, err := ms.provider.GenerateDomainVerificationToken(ctx, owner, msg.Domain)
	if err != nil {
		return nil, err
	}

	// Emit event
	_ = ctx.EventManager().EmitTypedEvent(
		&types.EventProviderDomainVerificationStarted{
			Owner:  msg.Owner,
			Domain: msg.Domain,
			Token:  record.Token,
		},
	)

	return &types.MsgGenerateDomainVerificationTokenResponse{
		Token:     record.Token,
		ExpiresAt: record.ExpiresAt,
	}, nil
}

func (ms msgServer) VerifyProviderDomain(goCtx context.Context, msg *types.MsgVerifyProviderDomain) (*types.MsgVerifyProviderDomainResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	owner, _ := sdk.AccAddressFromBech32(msg.Owner)

	// Verify provider exists
	if _, ok := ms.provider.Get(ctx, owner); !ok {
		return nil, types.ErrProviderNotFound.Wrapf("provider not found: %s", msg.Owner)
	}

	// Perform domain verification
	err := ms.provider.VerifyProviderDomain(ctx, owner)
	if err != nil {
		return &types.MsgVerifyProviderDomainResponse{
			Verified: false,
		}, err
	}

	return &types.MsgVerifyProviderDomainResponse{
		Verified: true,
	}, nil
}
