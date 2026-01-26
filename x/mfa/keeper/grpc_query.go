package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"pkg.akt.dev/node/x/mfa/types"
)

// GRPCQuerier wraps the keeper with query methods
type GRPCQuerier struct {
	Keeper
}

// GetMFAPolicy returns the MFA policy for an account
func (q GRPCQuerier) GetMFAPolicy(goCtx context.Context, req *types.QueryMFAPolicyRequest) (*types.QueryMFAPolicyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrapf("invalid address: %v", err)
	}

	policy, found := q.Keeper.GetMFAPolicy(ctx, address)
	if !found {
		return &types.QueryMFAPolicyResponse{
			Policy: nil,
		}, nil
	}

	return &types.QueryMFAPolicyResponse{
		Policy: policy,
	}, nil
}

// GetFactorEnrollments returns all factor enrollments for an account
func (q GRPCQuerier) GetFactorEnrollments(goCtx context.Context, req *types.QueryFactorEnrollmentsRequest) (*types.QueryFactorEnrollmentsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrapf("invalid address: %v", err)
	}

	enrollments := q.Keeper.GetFactorEnrollments(ctx, address)

	return &types.QueryFactorEnrollmentsResponse{
		Enrollments: enrollments,
	}, nil
}

// GetChallenge returns a challenge by ID
func (q GRPCQuerier) GetChallenge(goCtx context.Context, req *types.QueryChallengeRequest) (*types.QueryChallengeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	challenge, found := q.Keeper.GetChallenge(ctx, req.ChallengeID)
	if !found {
		return nil, types.ErrChallengeNotFound.Wrapf("challenge %s not found", req.ChallengeID)
	}

	return &types.QueryChallengeResponse{
		Challenge: challenge,
	}, nil
}

// GetPendingChallenges returns pending challenges for an account
func (q GRPCQuerier) GetPendingChallenges(goCtx context.Context, req *types.QueryPendingChallengesRequest) (*types.QueryPendingChallengesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrapf("invalid address: %v", err)
	}

	challenges := q.Keeper.GetPendingChallenges(ctx, address)

	return &types.QueryPendingChallengesResponse{
		Challenges: challenges,
	}, nil
}

// GetTrustedDevices returns trusted devices for an account
func (q GRPCQuerier) GetTrustedDevices(goCtx context.Context, req *types.QueryTrustedDevicesRequest) (*types.QueryTrustedDevicesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrapf("invalid address: %v", err)
	}

	devices := q.Keeper.GetTrustedDevices(ctx, address)

	return &types.QueryTrustedDevicesResponse{
		Devices: devices,
	}, nil
}

// GetSensitiveTxConfig returns the configuration for a sensitive tx type
func (q GRPCQuerier) GetSensitiveTxConfig(goCtx context.Context, req *types.QuerySensitiveTxConfigRequest) (*types.QuerySensitiveTxConfigResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	config, found := q.Keeper.GetSensitiveTxConfig(ctx, req.TransactionType)
	if !found {
		return nil, types.ErrInvalidSensitiveTxType.Wrapf("no config for tx type %s", req.TransactionType.String())
	}

	return &types.QuerySensitiveTxConfigResponse{
		Config: config,
	}, nil
}

// GetParams returns the module parameters
func (q GRPCQuerier) GetParams(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params := q.Keeper.GetParams(ctx)

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}

// IsMFARequired checks if MFA is required for a transaction
func (q GRPCQuerier) IsMFARequired(goCtx context.Context, req *types.QueryMFARequiredRequest) (*types.QueryMFARequiredResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrapf("invalid address: %v", err)
	}

	// Check if there's a sensitive tx config for this type
	config, found := q.Keeper.GetSensitiveTxConfig(ctx, req.TransactionType)
	if !found || !config.Enabled {
		return &types.QueryMFARequiredResponse{
			Required:           false,
			FactorCombinations: nil,
			MinVEIDScore:       0,
		}, nil
	}

	// Check if account has MFA enabled
	policy, policyFound := q.Keeper.GetMFAPolicy(ctx, address)
	if !policyFound || !policy.Enabled {
		// Use default config requirements
		return &types.QueryMFARequiredResponse{
			Required:           true,
			FactorCombinations: config.RequiredFactorCombinations,
			MinVEIDScore:       config.MinVEIDScore,
		}, nil
	}

	// Use account-specific policy
	requiredFactors := policy.GetRequiredFactorsForAction(req.TransactionType)

	return &types.QueryMFARequiredResponse{
		Required:           true,
		FactorCombinations: requiredFactors,
		MinVEIDScore:       policy.VEIDThreshold,
	}, nil
}
