package utils

import (
	sdkmath "cosmossdk.io/math"

	sdk "github.com/virtengine/virtengine/sdk/go/node/types/sdk"
)

func LeaseCalcBalanceRemain(balance sdkmath.LegacyDec, currBlock, settledAt int64, leasePrice sdk.DecCoin) sdk.DecCoin {
	res, _ := sdk.NewDecFromStr(balance.String())
	diff := sdk.ZeroDec()

	diff = diff.Add(leasePrice.Amount)
	diff = diff.MulInt64(currBlock - settledAt)

	res = res.Sub(diff)

	return sdk.NewDecCoinFromDec(leasePrice.Denom, res)
}

func LeaseCalcBlocksRemain(balance float64, leasePrice sdkmath.LegacyDec) int64 {
	return int64(balance / leasePrice.MustFloat64())
}
