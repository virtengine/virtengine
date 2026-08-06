package keeper

import (
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/veid/types"
)

func TestVoteExtensionCarrierSecureBoundary(t *testing.T) {
	t.Parallel()
	require.Equal(t, uint32(1), ActiveVoteExtensionCarrierVersion)
	require.Equal(t, ActiveVoteExtensionCarrierVersion, VoteExtensionVersion)

	var k Keeper
	_, err := k.ExtendVote(sdk.Context{}, nil, nil)
	require.Error(t, err)

	nilRequest, err := k.VerifyVoteExtension(sdk.Context{}, nil, nil)
	require.NoError(t, err)
	require.Equal(t, abci.ResponseVerifyVoteExtension_REJECT, nilRequest.Status)

	badPayloads := [][]byte{
		nil,
		[]byte("not-json"),
		[]byte(`{"version":1,"height":42}`),
		[]byte(`{"version":2,"height":42}`),
	}
	for _, payload := range badPayloads {
		response, verifyErr := k.VerifyVoteExtension(sdk.Context{}, &abci.RequestVerifyVoteExtension{
			Height:           42,
			Hash:             []byte("hash"),
			ValidatorAddress: []byte("validator"),
			VoteExtension:    payload,
		}, nil)
		require.NoError(t, verifyErr)
		require.Equal(t, abci.ResponseVerifyVoteExtension_REJECT, response.Status)
	}
}

func TestLegacyProposalVerificationHooksFailClosed(t *testing.T) {
	t.Parallel()

	var k Keeper
	results, err := k.PrepareProposalVerifications(sdk.Context{}, nil, 1)
	require.Error(t, err)
	require.Empty(t, results)

	require.NoError(t, k.ProcessProposalVerifications(sdk.Context{}, nil, nil))
	err = k.ProcessProposalVerifications(sdk.Context{}, []types.VerificationResult{{RequestID: "request"}}, nil)
	require.Error(t, err)
}
