package testutil

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func Coin(t testing.TB) sdk.Coin {
	t.Helper()
	return sdk.NewCoin("testcoin", sdk.NewInt(int64(RandRangeInt(1, 1000)))) // nolint: gosec
}

// VirtEngineCoin provides simple interface to the VirtEngine sdk.Coin type.
func VirtEngineCoinRandom(t testing.TB) sdk.Coin {

	t.Helper()
	amt := sdk.NewInt(int64(RandRangeInt(1, 1000)))
	return sdk.NewCoin(CoinDenom, amt)
}

// VirtEngineCoin provides simple interface to the VirtEngine sdk.Coin type.
func VirtEngineCoin(t testing.TB, amount int64) sdk.Coin {
	t.Helper()
	amt := sdk.NewInt(amount)
	return sdk.NewCoin(CoinDenom, amt)
}
