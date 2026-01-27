package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/virtengine/virtengine/x/encryption/types"
)

// Error message constant
const errMsgEmptyRequest = "empty request"

// GRPCQuerier implements the gRPC query interface with proper context handling
type GRPCQuerier struct {
	Keeper
}

var _ types.QueryServer = GRPCQuerier{}

// RecipientKey returns all keys registered for an address
func (q GRPCQuerier) RecipientKey(c context.Context, req *types.QueryRecipientKeyRequest) (*types.QueryRecipientKeyResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	ctx := sdk.UnwrapSDKContext(c)

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(err.Error())
	}

	keys := q.Keeper.GetRecipientKeys(ctx, addr)

	return &types.QueryRecipientKeyResponse{
		Keys: keys,
	}, nil
}

// KeyByFingerprint returns a key by its fingerprint
func (q GRPCQuerier) KeyByFingerprint(c context.Context, req *types.QueryKeyByFingerprintRequest) (*types.QueryKeyByFingerprintResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	ctx := sdk.UnwrapSDKContext(c)

	key, found := q.Keeper.GetRecipientKeyByFingerprint(ctx, req.Fingerprint)
	if !found {
		return nil, types.ErrKeyNotFound.Wrapf("key with fingerprint %s not found", req.Fingerprint)
	}

	return &types.QueryKeyByFingerprintResponse{
		Key: key,
	}, nil
}

// Params returns the module parameters
func (q GRPCQuerier) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	ctx := sdk.UnwrapSDKContext(c)
	params := q.Keeper.GetParams(ctx)

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}

// Algorithms returns the supported algorithms
func (q GRPCQuerier) Algorithms(c context.Context, req *types.QueryAlgorithmsRequest) (*types.QueryAlgorithmsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	var algorithms []types.AlgorithmInfo
	for _, algID := range types.SupportedAlgorithms() {
		info, err := types.GetAlgorithmInfo(algID)
		if err == nil {
			algorithms = append(algorithms, info)
		}
	}

	return &types.QueryAlgorithmsResponse{
		Algorithms: algorithms,
	}, nil
}

// ValidateEnvelope validates an envelope structure
func (q GRPCQuerier) ValidateEnvelope(c context.Context, req *types.QueryValidateEnvelopeRequest) (*types.QueryValidateEnvelopeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, errMsgEmptyRequest)
	}

	ctx := sdk.UnwrapSDKContext(c)

	envelope := &req.Envelope
	response := &types.QueryValidateEnvelopeResponse{
		Valid:          true,
		RecipientCount: len(envelope.RecipientKeyIDs),
		Algorithm:      envelope.AlgorithmID,
	}

	// Validate envelope structure
	if err := q.Keeper.ValidateEnvelope(ctx, envelope); err != nil {
		response.Valid = false
		response.Error = err.Error()
		return response, nil
	}

	// Validate recipients
	missingKeys, err := q.Keeper.ValidateEnvelopeRecipients(ctx, envelope)
	if err != nil {
		response.Valid = false
		response.Error = err.Error()
		return response, nil
	}

	response.MissingKeys = missingKeys
	response.AllKeysRegistered = len(missingKeys) == 0

	// Signature validation
	response.SignatureValid = len(envelope.SenderSignature) > 0

	return response, nil
}
