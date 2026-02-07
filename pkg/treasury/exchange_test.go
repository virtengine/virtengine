package treasury

import (
	"context"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/pkg/dex"
)

func TestExchangeRouter_SelectBestQuote(t *testing.T) {
	dexAdapter, err := dex.NewOsmosisAdapter(dex.AdapterConfig{
		Name:    "osmosis-test",
		Type:    "osmosis",
		Enabled: true,
		ChainID: "osmosis-1",
	})
	require.NoError(t, err)

	router := NewExchangeRouter(DefaultBestExecutionPolicy())
	router.RegisterAdapter(NewDexAdapterWrapper(dexAdapter))
	router.RegisterAdapter(NewTestCEXAdapter("cex-test", map[string]sdkmath.LegacyDec{
		"UVE/USDC": sdkmath.LegacyMustNewDecFromStr("1.10"),
	}, 10))

	req := ExchangeRequest{
		FromAsset:   "UVE",
		ToAsset:     "USDC",
		Amount:      sdkmath.NewInt(1000),
		SlippageBps: 25,
		Deadline:    time.Now().Add(time.Minute),
	}

	quote, err := router.SelectBestQuote(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, AdapterTypeCEX, quote.AdapterType)
}

func TestExchangeRouter_ExecuteDexAndCex(t *testing.T) {
	dexAdapter, err := dex.NewOsmosisAdapter(dex.AdapterConfig{
		Name:    "osmosis-test",
		Type:    "osmosis",
		Enabled: true,
		ChainID: "osmosis-1",
	})
	require.NoError(t, err)

	router := NewExchangeRouter(DefaultBestExecutionPolicy())
	router.RegisterAdapter(NewDexAdapterWrapper(dexAdapter))

	req := ExchangeRequest{
		FromAsset:   "UVE",
		ToAsset:     "USDC",
		Amount:      sdkmath.NewInt(500),
		SlippageBps: 25,
		Deadline:    time.Now().Add(time.Minute),
	}

	dexExec, err := router.ExecuteBestQuote(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, AdapterTypeDEX, dexExec.Quote.AdapterType)

	cexRouter := NewExchangeRouter(DefaultBestExecutionPolicy())
	cexRouter.RegisterAdapter(NewTestCEXAdapter("cex-test", map[string]sdkmath.LegacyDec{
		"UVE/USDC": sdkmath.LegacyMustNewDecFromStr("1.02"),
	}, 15))

	cexReq := req
	cexReq.Amount = sdkmath.NewInt(1200)

	exec, err := cexRouter.ExecuteBestQuote(context.Background(), cexReq)
	require.NoError(t, err)
	require.Equal(t, AdapterTypeCEX, exec.Quote.AdapterType)
	require.NotEmpty(t, exec.TxID)
}
