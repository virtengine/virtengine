package app

import (
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	veidkeeper "github.com/virtengine/virtengine/x/veid/keeper"
)

func TestVEIDVoteExtensionCarrierVersionOneBoundary(t *testing.T) {
	t.Parallel()
	require.Equal(t, uint32(1), activeVEIDVoteExtensionCarrierVersion)
	require.Panics(t, func() { newVEIDVoteExtensionHandlers(nil) })

	var keeper veidkeeper.Keeper
	extend, verify := newVEIDVoteExtensionHandlers(&keeper)
	_, err := extend(sdk.Context{}, nil)
	require.Error(t, err)

	tests := []struct {
		name   string
		req    *abci.RequestVerifyVoteExtension
		status abci.ResponseVerifyVoteExtension_VerifyStatus
	}{
		{name: "nil request", status: abci.ResponseVerifyVoteExtension_REJECT},
		{name: "empty extension", req: &abci.RequestVerifyVoteExtension{}, status: abci.ResponseVerifyVoteExtension_REJECT},
		{name: "non-empty extension", req: &abci.RequestVerifyVoteExtension{VoteExtension: []byte{1}}, status: abci.ResponseVerifyVoteExtension_REJECT},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			response, verifyErr := verify(sdk.Context{}, tc.req)
			require.NoError(t, verifyErr)
			require.Equal(t, tc.status, response.Status)
		})
	}
}
