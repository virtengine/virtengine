package provider_daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	tmbytes "github.com/cometbft/cometbft/libs/bytes"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	providertypes "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
	"github.com/virtengine/virtengine/x/provider/keeper"
)

const testLiteral = "test"

type domainVerificationChainBackend interface {
	QueryDomainVerificationRecord(context.Context, sdk.AccAddress) (*keeper.DomainVerificationRecord, error)
	ConfirmDomainVerification(context.Context, sdk.AccAddress, string) error
}

type rpcDomainVerificationBackend struct {
	storeQuery providerStoreQueryClient
	timeout    time.Duration
}

func (b *rpcDomainVerificationBackend) QueryDomainVerificationRecord(
	ctx context.Context,
	providerAddr sdk.AccAddress,
) (*keeper.DomainVerificationRecord, error) {
	value, err := queryProviderStoreValue(ctx, b.storeQuery, b.timeout, keeper.DomainVerificationKey(providerAddr))
	if err != nil {
		return nil, err
	}
	if len(value) == 0 {
		return nil, nil
	}

	var record keeper.DomainVerificationRecord
	if err := json.Unmarshal(value, &record); err != nil {
		return nil, fmt.Errorf("decode domain verification record: %w", err)
	}

	return &record, nil
}

func (b *rpcDomainVerificationBackend) ConfirmDomainVerification(
	ctx context.Context,
	providerAddr sdk.AccAddress,
	proof string,
) error {
	_ = ctx
	_ = providerAddr
	_ = proof
	return fmt.Errorf("%w: domain confirmation requires generalized mutation submitter", ErrProviderMutationUnavailable)
}

func queryProviderStoreValue(
	ctx context.Context,
	client providerStoreQueryClient,
	timeout time.Duration,
	key []byte,
) ([]byte, error) {
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := client.ABCIQueryWithOptions(
		reqCtx,
		fmt.Sprintf("/store/%s/key", providertypes.StoreKey),
		tmbytes.HexBytes(key),
		rpcclient.ABCIQueryOptions{},
	)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("empty abci query response")
	}
	if result.Response.IsErr() {
		return nil, fmt.Errorf(
			"abci query failed with code %d: %s",
			result.Response.GetCode(),
			result.Response.GetLog(),
		)
	}
	return result.Response.GetValue(), nil
}
