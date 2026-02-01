package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	cclient "github.com/virtengine/virtengine/sdk/go/node/client/v1beta3"
)

func Test_splitAndCall_NoMessages(t *testing.T) {
	err := newSplitAndApply(context.TODO(), nil, nil, 10)
	require.NoError(t, err, "")
}

func Test_splitAndCall_Splitting(t *testing.T) {
	addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// Add five messages
	msgs := []sdk.Msg{
		testdata.NewTestMsg(addr),
		testdata.NewTestMsg(addr),
		testdata.NewTestMsg(addr),
		testdata.NewTestMsg(addr),
		testdata.NewTestMsg(addr),
	}

	// Keep track of number of calls
	const chunkSize = 2

	callCount := 0
	err := newSplitAndApply(
		context.TODO(),
		func(_ context.Context, msgs []sdk.Msg, _ ...cclient.BroadcastOption) (interface{}, error) {
			callCount++

			require.NotNil(t, msgs)

			if callCount < 3 {
				require.Equal(t, len(msgs), 2)
			} else {
				require.Equal(t, len(msgs), 1)
			}

			return nil, nil
		},
		msgs, chunkSize)

	require.NoError(t, err, "")
	require.Equal(t, 3, callCount)
}

