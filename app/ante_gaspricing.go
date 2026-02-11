package app

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/app/gaspricing"
)

// AdaptiveGasPriceDecorator enforces adaptive min gas prices during CheckTx.
type AdaptiveGasPriceDecorator struct {
	keeper gaspricing.Keeper
}

// NewAdaptiveGasPriceDecorator creates a new decorator.
func NewAdaptiveGasPriceDecorator(keeper gaspricing.Keeper) AdaptiveGasPriceDecorator {
	return AdaptiveGasPriceDecorator{keeper: keeper}
}

// AnteHandle enforces minimum gas prices when enabled.
func (d AdaptiveGasPriceDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if !ctx.IsCheckTx() || simulate {
		return next(ctx, tx, simulate)
	}

	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, fmt.Errorf("tx does not implement FeeTx")
	}

	params := d.keeper.GetParams(ctx)
	if !params.Enabled {
		return next(ctx, tx, simulate)
	}

	minGasPrices := d.keeper.GetState(ctx).CurrentMinGasPrices
	if len(minGasPrices) == 0 {
		minGasPrices = params.MinGasPrices
	}
	if len(minGasPrices) == 0 {
		return next(ctx, tx, simulate)
	}

	gas := feeTx.GetGas()
	if gas == 0 {
		return ctx, fmt.Errorf("gas limit must be positive")
	}

	fees := feeTx.GetFee()
	feeDec := sdk.NewDecCoinsFromCoins(fees...)
	gasDec := sdkmath.LegacyNewDec(int64(gas))
	gasPrices := feeDec.QuoDec(gasDec)

	if !gasPrices.IsAllGTE(minGasPrices) {
		return ctx, fmt.Errorf("insufficient fee: min gas prices %s, got %s", minGasPrices, gasPrices)
	}

	return next(ctx, tx, simulate)
}
