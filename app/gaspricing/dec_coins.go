package gaspricing

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func scaleDecCoins(coins sdk.DecCoins, multiplierBPS int64) sdk.DecCoins {
	if len(coins) == 0 {
		return sdk.DecCoins{}
	}
	if multiplierBPS <= 0 {
		return sdk.DecCoins{}
	}
	multiplier := sdkmath.LegacyNewDec(multiplierBPS).QuoInt64(10000)
	return coins.MulDec(multiplier).Sort()
}

func clampDecCoins(value, min, max sdk.DecCoins) sdk.DecCoins {
	value = value.Sort()
	min = min.Sort()
	max = max.Sort()

	denoms := make(map[string]struct{})
	for _, coin := range value {
		denoms[coin.Denom] = struct{}{}
	}
	for _, coin := range min {
		denoms[coin.Denom] = struct{}{}
	}
	for _, coin := range max {
		denoms[coin.Denom] = struct{}{}
	}

	out := sdk.DecCoins{}
	for denom := range denoms {
		amt := value.AmountOf(denom)
		minAmt := min.AmountOf(denom)
		maxAmt := max.AmountOf(denom)

		if minAmt.IsPositive() && amt.LT(minAmt) {
			amt = minAmt
		}
		if maxAmt.IsPositive() && amt.GT(maxAmt) {
			amt = maxAmt
		}
		if amt.IsPositive() {
			out = out.Add(sdk.NewDecCoinFromDec(denom, amt))
		}
	}

	return out.Sort()
}

func decCoinsAllGTE(value, min sdk.DecCoins) bool {
	value = value.Sort()
	min = min.Sort()

	denoms := make(map[string]struct{})
	for _, coin := range value {
		denoms[coin.Denom] = struct{}{}
	}
	for _, coin := range min {
		denoms[coin.Denom] = struct{}{}
	}

	for denom := range denoms {
		if value.AmountOf(denom).LT(min.AmountOf(denom)) {
			return false
		}
	}
	return true
}

// DecCoinsAllGTE reports whether all denominations in value are >= min.
func DecCoinsAllGTE(value, min sdk.DecCoins) bool {
	return decCoinsAllGTE(value, min)
}
