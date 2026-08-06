package keeper

import (
	"context"
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/pkg/dex"
	"github.com/virtengine/virtengine/pkg/payments/offramp"
	"github.com/virtengine/virtengine/x/settlement/types"
)

type forbiddenPriceFeed struct{ calls int }

func (f *forbiddenPriceFeed) GetPrice(context.Context, string, string) (types.Price, error) {
	f.calls++
	return types.Price{Rate: math.LegacyOneDec()}, nil
}

func (f *forbiddenPriceFeed) GetPrices(context.Context, []types.CurrencyPair) ([]types.Price, error) {
	f.calls++
	return nil, nil
}

func (f *forbiddenPriceFeed) SubscribePrices(context.Context, []types.CurrencyPair) (<-chan types.PriceUpdate, error) {
	f.calls++
	return nil, nil
}

type forbiddenDexExecutor struct{ calls int }

func (f *forbiddenDexExecutor) GetQuote(context.Context, dex.SwapRequest) (dex.SwapQuote, error) {
	f.calls++
	return dex.SwapQuote{}, nil
}

func (f *forbiddenDexExecutor) ExecuteSwap(context.Context, dex.SwapQuote, []byte) (dex.SwapResult, error) {
	f.calls++
	return dex.SwapResult{}, nil
}

type forbiddenOffRampBridge struct{ calls int }

func (f *forbiddenOffRampBridge) GetQuote(context.Context, offramp.QuoteRequest) (offramp.Quote, error) {
	f.calls++
	return offramp.Quote{}, nil
}

func (f *forbiddenOffRampBridge) InitiatePayout(context.Context, offramp.Quote, string, string, map[string]string) (offramp.PayoutResult, error) {
	f.calls++
	return offramp.PayoutResult{}, nil
}

func (f *forbiddenOffRampBridge) GetStatus(context.Context, string) (offramp.PayoutResult, error) {
	f.calls++
	return offramp.PayoutResult{}, nil
}

func (f *forbiddenOffRampBridge) FindPayoutByMetadata(context.Context, string, map[string]string) (offramp.PayoutResult, error) {
	f.calls++
	return offramp.PayoutResult{}, nil
}

func (f *forbiddenOffRampBridge) Cancel(context.Context, string) error {
	f.calls++
	return nil
}

func TestKeeperDiscardsExternalConsensusAdapters(t *testing.T) {
	t.Parallel()

	var k Keeper
	priceFeed := &forbiddenPriceFeed{}
	dexExecutor := &forbiddenDexExecutor{}
	offRampBridge := &forbiddenOffRampBridge{}

	k.SetPriceFeed(types.OracleSourceTypeBandIBC, priceFeed)
	k.SetDexSwapExecutor(dexExecutor)
	k.SetOffRampBridge(offRampBridge)

	require.Nil(t, k.priceFeedForSource(types.OracleSourceTypeBandIBC))
	require.Nil(t, k.dexSwap)
	require.Nil(t, k.offRampBridge)
	require.ErrorIs(t, ensureNoConsensusExternalIO(), types.ErrExternalIOForbidden)
	require.Zero(t, priceFeed.calls)
	require.Zero(t, dexExecutor.calls)
	require.Zero(t, offRampBridge.calls)
}
