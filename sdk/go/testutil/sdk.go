package testutil

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	cmbtypes "github.com/cometbft/cometbft/abci/types"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/virtengine/virtengine/sdk/go/sdkutil"
)

func Coin(t testing.TB) sdk.Coin {
	t.Helper()
	return sdk.NewCoin("testcoin", sdkmath.NewInt(int64(RandRangeInt(1, 1000)))) // nolint: gosec
}

func DecCoin(t testing.TB) sdk.DecCoin {
	t.Helper()
	return sdk.NewDecCoin("testcoin", sdkmath.NewInt(int64(RandRangeInt(1, 1000)))) // nolint: gosec
}

// AkashCoinRandom provides simple interface to the Akash sdk.Coin type.
func AkashCoinRandom(t testing.TB) sdk.Coin {
	t.Helper()
	amt := sdkmath.NewInt(int64(RandRangeInt(1, 1000)))
	return sdk.NewCoin(sdkutil.Denomuve, amt)
}

// AkashCoin provides simple interface to the Akash sdk.Coin type.
func AkashCoin(t testing.TB, amount int64) sdk.Coin {
	t.Helper()
	amt := sdkmath.NewInt(amount)
	return sdk.NewCoin(sdkutil.Denomuve, amt)
}

func AkashDecCoin(t testing.TB, amount int64) sdk.DecCoin {
	t.Helper()
	amt := sdkmath.NewInt(amount)
	return sdk.NewDecCoin(sdkutil.Denomuve, amt)
}

func AkashDecCoinRandom(t testing.TB) sdk.DecCoin {
	t.Helper()
	amt := sdkmath.NewInt(int64(RandRangeInt(1, 1000)))
	return sdk.NewDecCoin(sdkutil.Denomuve, amt)
}

// ACTCoinRandom provides simple interface to the ACT sdk.Coin type.
func ACTCoinRandom(t testing.TB) sdk.Coin {
	t.Helper()
	amt := sdkmath.NewInt(int64(RandRangeInt(1, 1000)))
	return sdk.NewCoin(sdkutil.DenomUact, amt)
}

// ACTCoin provides simple interface to the ACT sdk.Coin type.
func ACTCoin(t testing.TB, amount int64) sdk.Coin {
	t.Helper()
	amt := sdkmath.NewInt(amount)
	return sdk.NewCoin(sdkutil.DenomUact, amt)
}

func ACTDecCoin(t testing.TB, amount int64) sdk.DecCoin {
	t.Helper()
	amt := sdkmath.NewInt(amount)
	return sdk.NewDecCoin(sdkutil.DenomUact, amt)
}

func ACTDecCoinRandom(t testing.TB) sdk.DecCoin {
	t.Helper()
	amt := sdkmath.NewInt(int64(RandRangeInt(1, 1000)))
	return sdk.NewDecCoin(sdkutil.DenomUact, amt)
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

