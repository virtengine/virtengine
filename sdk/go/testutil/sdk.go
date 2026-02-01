package testutil

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	cmbtypes "github.com/cometbft/cometbft/abci/types"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
)

func Coin(t testing.TB) sdk.Coin {
	t.Helper()
	return sdk.NewCoin("testcoin", sdkmath.NewInt(int64(RandRangeInt(1, 1000)))) // nolint: gosec
}

func DecCoin(t testing.TB) sdk.DecCoin {
	t.Helper()
	return sdk.NewDecCoin("testcoin", sdkmath.NewInt(int64(RandRangeInt(1, 1000)))) // nolint: gosec
}

func EnsureEvent(t *testing.T, events []cmbtypes.Event, expEvent proto.Message) {
	for _, e := range events {
		iev, err := sdk.ParseTypedEvent(e)
		require.NoError(t, err)
		if reflect.DeepEqual(iev, expEvent) {
			return
		}
	}

	t.Errorf("events don't have required event \"%v\"", expEvent)
}

