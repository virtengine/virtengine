package keeper

import (
	"context"
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/enclave/types"
)

var _ types.QueryServer = queryServer{}

type queryServer struct {
	keeper Keeper
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &queryServer{keeper: keeper}
}

// EnclaveIdentity queries an enclave identity by validator address
func (q queryServer) EnclaveIdentity(goCtx context.Context, req *types.QueryEnclaveIdentityRequest) (*types.QueryEnclaveIdentityResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	validatorAddr, err := sdk.AccAddressFromBech32(req.ValidatorAddress)
	if err != nil {
		return nil, types.ErrInvalidEnclaveIdentity.Wrapf("invalid validator address: %v", err)
	}

	identity, exists := q.keeper.GetEnclaveIdentity(ctx, validatorAddr)
	if !exists {
		return &types.QueryEnclaveIdentityResponse{Identity: nil}, nil
	}

	return &types.QueryEnclaveIdentityResponse{Identity: identity}, nil
}

// ActiveValidatorEnclaveKeys queries all active validator enclave keys
func (q queryServer) ActiveValidatorEnclaveKeys(goCtx context.Context, req *types.QueryActiveValidatorEnclaveKeysRequest) (*types.QueryActiveValidatorEnclaveKeysResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	identities := q.keeper.GetActiveValidatorEnclaveKeys(ctx)

	return &types.QueryActiveValidatorEnclaveKeysResponse{Identities: identities}, nil
}

// CommitteeEnclaveKeys queries enclave keys for the identity committee
func (q queryServer) CommitteeEnclaveKeys(goCtx context.Context, req *types.QueryCommitteeEnclaveKeysRequest) (*types.QueryCommitteeEnclaveKeysResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	identities := q.keeper.GetCommitteeEnclaveKeys(ctx, req.CommitteeEpoch)

	return &types.QueryCommitteeEnclaveKeysResponse{Identities: identities}, nil
}

// MeasurementAllowlist queries the measurement allowlist
func (q queryServer) MeasurementAllowlist(goCtx context.Context, req *types.QueryMeasurementAllowlistRequest) (*types.QueryMeasurementAllowlistResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	measurements := q.keeper.GetMeasurementAllowlist(ctx, req.TeeType, req.IncludeRevoked)

	return &types.QueryMeasurementAllowlistResponse{Measurements: measurements}, nil
}

// Measurement queries a specific measurement
func (q queryServer) Measurement(goCtx context.Context, req *types.QueryMeasurementRequest) (*types.QueryMeasurementResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Decode hex-encoded measurement hash
	measurementHash, err := hex.DecodeString(req.MeasurementHash)
	if err != nil {
		return nil, types.ErrInvalidMeasurement.Wrapf("invalid measurement hash: %v", err)
	}

	measurement, exists := q.keeper.GetMeasurement(ctx, measurementHash)
	if !exists {
		return &types.QueryMeasurementResponse{Measurement: nil, IsAllowed: false}, nil
	}

	isAllowed := q.keeper.IsMeasurementAllowed(ctx, measurementHash, ctx.BlockHeight())

	return &types.QueryMeasurementResponse{Measurement: measurement, IsAllowed: isAllowed}, nil
}

// KeyRotation queries the key rotation status for a validator
func (q queryServer) KeyRotation(goCtx context.Context, req *types.QueryKeyRotationRequest) (*types.QueryKeyRotationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	validatorAddr, err := sdk.AccAddressFromBech32(req.ValidatorAddress)
	if err != nil {
		return nil, types.ErrInvalidEnclaveIdentity.Wrapf("invalid validator address: %v", err)
	}

	rotation, exists := q.keeper.GetActiveKeyRotation(ctx, validatorAddr)

	return &types.QueryKeyRotationResponse{
		Rotation:          rotation,
		HasActiveRotation: exists,
	}, nil
}

// ValidKeySet queries the current valid key set
func (q queryServer) ValidKeySet(goCtx context.Context, req *types.QueryValidKeySetRequest) (*types.QueryValidKeySetResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	forHeight := req.ForBlockHeight
	if forHeight == 0 {
		forHeight = ctx.BlockHeight()
	}

	keys := q.keeper.GetValidKeySet(ctx, forHeight)

	return &types.QueryValidKeySetResponse{
		ValidatorKeys: keys,
		TotalCount:    int32(len(keys)),
	}, nil
}

// Params queries the module parameters
func (q queryServer) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params := q.keeper.GetParams(ctx)

	return &types.QueryParamsResponse{Params: params}, nil
}

// AttestedResult queries an attested scoring result
func (q queryServer) AttestedResult(goCtx context.Context, req *types.QueryAttestedResultRequest) (*types.QueryAttestedResultResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	result, _ := q.keeper.GetAttestedResult(ctx, req.BlockHeight, req.ScopeId)

	return &types.QueryAttestedResultResponse{Result: result}, nil
}
