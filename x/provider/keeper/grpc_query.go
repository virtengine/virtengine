package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkquery "github.com/cosmos/cosmos-sdk/types/query"

	types "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
)

// Querier is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper
type Querier struct {
	Keeper
}

var _ types.QueryServer = Querier{}

// Providers returns providers list
func (k Querier) Providers(c context.Context, req *types.QueryProvidersRequest) (*types.QueryProvidersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	var providers types.Providers
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.skey)

	pageRes, err := sdkquery.Paginate(store, req.Pagination, func(_ []byte, value []byte) error {
		var provider types.Provider

		err := k.cdc.Unmarshal(value, &provider)
		if err != nil {
			return err
		}

		providers = append(providers, provider)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryProvidersResponse{
		Providers:  providers,
		Pagination: pageRes,
	}, nil
}

// Provider returns provider details based on owner address
func (k Querier) Provider(c context.Context, req *types.QueryProviderRequest) (*types.QueryProviderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if err != nil {
		return nil, types.ErrInvalidAddress
	}

	ctx := sdk.UnwrapSDKContext(c)

	provider, found := k.Get(ctx, owner)
	if !found {
		return nil, types.ErrProviderNotFound
	}

	return &types.QueryProviderResponse{Provider: provider}, nil
}

// ProviderSigningKey returns one exact provider key epoch.
func (k Querier) ProviderSigningKey(c context.Context, req *types.QueryProviderSigningKeyRequest) (*types.QueryProviderSigningKeyResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if err != nil {
		return nil, types.ErrInvalidAddress
	}
	ctx := sdk.UnwrapSDKContext(c)
	record, found := k.GetProviderSigningKey(ctx, owner, req.KeyId, req.Epoch)
	if !found {
		return nil, types.ErrInvalidPublicKey.Wrap("provider signing key epoch not found")
	}
	return &types.QueryProviderSigningKeyResponse{Key: toProviderSigningKeyRecord(record)}, nil
}

// ProviderSigningKeyEpochs returns the immutable key history.
func (k Querier) ProviderSigningKeyEpochs(c context.Context, req *types.QueryProviderSigningKeyEpochsRequest) (*types.QueryProviderSigningKeyEpochsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if err != nil {
		return nil, types.ErrInvalidAddress
	}
	ctx := sdk.UnwrapSDKContext(c)
	records := k.GetProviderSigningKeyEpochs(ctx, owner)
	keys := make([]types.ProviderSigningKeyRecord, 0, len(records))
	for _, record := range records {
		keys = append(keys, toProviderSigningKeyRecord(record))
	}
	return &types.QueryProviderSigningKeyEpochsResponse{Keys: keys}, nil
}

func toProviderSigningKeyRecord(record types.ProviderPublicKeyRecord) types.ProviderSigningKeyRecord {
	return types.ProviderSigningKeyRecord{
		PublicKey:         append([]byte(nil), record.PublicKey...),
		KeyType:           record.KeyType,
		KeyId:             record.KeyID,
		Epoch:             record.Epoch,
		ActivatedAtHeight: record.ActivatedAtHeight,
		ActivatedAtUnix:   record.ActivatedAtUnix,
		ExpiresAtHeight:   record.ExpiresAtHeight,
		ExpiresAtUnix:     record.ExpiresAtUnix,
		RetiredAtHeight:   record.RetiredAtHeight,
		RetiredAtUnix:     record.RetiredAtUnix,
		RevokedAtHeight:   record.RevokedAtHeight,
		RevokedAtUnix:     record.RevokedAtUnix,
		PreviousKeyId:     record.PreviousKeyID,
		RotationCount:     record.RotationCount,
	}
}
