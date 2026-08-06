package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/veid/types"
)

func TestIsSSONonceUsedPrunesExpiredRecord(t *testing.T) {
	keeper, ctx := setupPipelineTestKeeper(t)

	now := ctx.BlockTime()
	record := types.NewSSONonceRecord(
		hashNonce("nonce-1"),
		sdk.AccAddress(make([]byte, 20)).String(),
		types.SSOProviderOIDC,
		"https://issuer.example",
		"linkage-1",
		now.Add(-2*time.Hour),
		ctx.BlockHeight(),
		time.Hour,
	)
	keeper.SetSSONonceRecord(ctx, record)

	require.False(t, keeper.IsSSONonceUsed(ctx.WithBlockTime(now), record.NonceHash))
	_, found := keeper.GetSSONonceRecord(ctx, record.NonceHash)
	require.False(t, found)
}
