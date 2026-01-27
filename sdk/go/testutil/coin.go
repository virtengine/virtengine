package testutil

import (
	"fmt"
	"math/rand"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// VEDenom is the native token denomination for VirtEngine
	VEDenom = "uve"
)

// VECoin creates a sdk.Coin with the VE denomination and specified amount
func VECoin(t testing.TB, amount int64) sdk.Coin {
	t.Helper()
	return sdk.NewCoin(VEDenom, sdkmath.NewInt(amount))
}

// VECoinRandom creates a sdk.Coin with the VE denomination and a random amount
// between min and max (inclusive)
func VECoinRandom(t testing.TB, minAmount, maxAmount int64) sdk.Coin {
	t.Helper()
	if minAmount > maxAmount {
		t.Fatalf("VECoinRandom: minAmount (%d) must be <= maxAmount (%d)", minAmount, maxAmount)
	}
	if minAmount < 0 {
		t.Fatalf("VECoinRandom: minAmount must be non-negative, got %d", minAmount)
	}
	amount := minAmount
	if maxAmount > minAmount {
		amount = minAmount + rand.Int63n(maxAmount-minAmount+1) // nolint: gosec
	}
	return sdk.NewCoin(VEDenom, sdkmath.NewInt(amount))
}

// VEDecCoin creates a sdk.DecCoin with the VE denomination and specified amount
func VEDecCoin(t testing.TB, amount int64) sdk.DecCoin {
	t.Helper()
	return sdk.NewDecCoin(VEDenom, sdkmath.NewInt(amount))
}

// VEDecCoinRandom creates a sdk.DecCoin with the VE denomination and a random amount
// suitable for use as a price in bids/orders
func VEDecCoinRandom(t testing.TB) sdk.DecCoin {
	t.Helper()
	// Generate a random price between 1 and 1000 uve
	amount := 1 + rand.Int63n(1000) // nolint: gosec
	return sdk.NewDecCoin(VEDenom, sdkmath.NewInt(amount))
}

// VEDecCoinAmount creates a sdk.DecCoin with the VE denomination and exact decimal amount
func VEDecCoinAmount(t testing.TB, amount string) sdk.DecCoin {
	t.Helper()
	dec, err := sdkmath.LegacyNewDecFromStr(amount)
	if err != nil {
		t.Fatalf("VEDecCoinAmount: invalid amount %q: %v", amount, err)
	}
	return sdk.NewDecCoinFromDec(VEDenom, dec)
}

// VECoins creates a sdk.Coins with the VE denomination and specified amount
func VECoins(t testing.TB, amount int64) sdk.Coins {
	t.Helper()
	return sdk.NewCoins(VECoin(t, amount))
}

// VECoinsRandom creates a sdk.Coins with the VE denomination and a random amount
func VECoinsRandom(t testing.TB, minAmount, maxAmount int64) sdk.Coins {
	t.Helper()
	return sdk.NewCoins(VECoinRandom(t, minAmount, maxAmount))
}

// MustVEDecCoin creates a sdk.DecCoin from a string amount, panicking on error
// Useful for test data initialization
func MustVEDecCoin(amount string) sdk.DecCoin {
	dec, err := sdkmath.LegacyNewDecFromStr(amount)
	if err != nil {
		panic(fmt.Sprintf("MustVEDecCoin: invalid amount %q: %v", amount, err))
	}
	return sdk.NewDecCoinFromDec(VEDenom, dec)
}
