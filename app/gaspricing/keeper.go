package gaspricing

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	paramsKey = []byte{0x01}
	stateKey  = []byte{0x02}
)

// Keeper manages adaptive gas pricing parameters and state.
type Keeper struct {
	storeKey storetypes.StoreKey
	logger   log.Logger
	defaults Params
}

// NewKeeper constructs a new adaptive gas pricing keeper.
func NewKeeper(storeKey storetypes.StoreKey, logger log.Logger, defaults Params) Keeper {
	return Keeper{
		storeKey: storeKey,
		logger:   logger,
		defaults: defaults,
	}
}

// IsZero reports whether the keeper has been initialized with a store key.
func (k Keeper) IsZero() bool {
	return k.storeKey == nil
}

// GetParams returns configured parameters or defaults.
func (k Keeper) GetParams(ctx sdk.Context) Params {
	store := ctx.KVStore(k.storeKey)
	if !store.Has(paramsKey) {
		return k.defaults
	}
	bz := store.Get(paramsKey)
	var params Params
	if err := json.Unmarshal(bz, &params); err != nil {
		return k.defaults
	}
	if err := params.Validate(); err != nil {
		return k.defaults
	}
	return params
}

// SetParams saves adaptive gas pricing parameters.
func (k Keeper) SetParams(ctx sdk.Context, params Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("marshal gas pricing params: %w", err)
	}
	store.Set(paramsKey, bz)
	return nil
}

// GetState returns the current adaptive gas state or defaults.
func (k Keeper) GetState(ctx sdk.Context) State {
	store := ctx.KVStore(k.storeKey)
	if !store.Has(stateKey) {
		return DefaultState(k.GetParams(ctx))
	}
	bz := store.Get(stateKey)
	var state State
	if err := json.Unmarshal(bz, &state); err != nil {
		return DefaultState(k.GetParams(ctx))
	}
	if state.CurrentMinGasPrices == nil || len(state.CurrentMinGasPrices) == 0 {
		state.CurrentMinGasPrices = k.GetParams(ctx).MinGasPrices
	}
	return state
}

// SetState saves the adaptive gas state.
func (k Keeper) SetState(ctx sdk.Context, state State) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal gas pricing state: %w", err)
	}
	store.Set(stateKey, bz)
	return nil
}

// UpdateMinGasPrices adjusts min gas prices based on congestion metrics.
func (k Keeper) UpdateMinGasPrices(ctx sdk.Context, gasUsed, gasLimit int64) (sdk.DecCoins, State, error) {
	params := k.GetParams(ctx)
	state := k.GetState(ctx)

	if !params.Enabled {
		state.CurrentMinGasPrices = params.MinGasPrices
		state.LastBlockHeight = ctx.BlockHeight()
		if err := k.SetState(ctx, state); err != nil {
			return params.MinGasPrices, state, err
		}
		return params.MinGasPrices, state, nil
	}

	utilBPS := int64(0)
	if gasLimit > 0 && gasUsed > 0 {
		utilBPS = (gasUsed * 10000) / gasLimit
		if utilBPS < 0 {
			utilBPS = 0
		}
		if utilBPS > 10000 {
			utilBPS = 10000
		}
	}

	smoothed := utilBPS
	if params.UtilizationSmoothingWindow > 1 {
		window := int64(params.UtilizationSmoothingWindow)
		smoothed = (state.SmoothedUtilizationBPS*(window-1) + utilBPS) / window
	}

	current := state.CurrentMinGasPrices
	if len(current) == 0 {
		current = params.MinGasPrices
	}

	delta := smoothed - params.TargetBlockUtilizationBPS
	adjustmentBPS := (delta * params.AdjustmentRateBPS) / 10000
	if adjustmentBPS > params.MaxChangeBPS {
		adjustmentBPS = params.MaxChangeBPS
	} else if adjustmentBPS < -params.MaxChangeBPS {
		adjustmentBPS = -params.MaxChangeBPS
	}
	multiplierBPS := int64(10000) + adjustmentBPS
	next := scaleDecCoins(current, multiplierBPS)

	if params.CongestionThresholdBPS > 0 && smoothed >= params.CongestionThresholdBPS {
		next = scaleDecCoins(next, int64(10000)+params.CongestionMultiplierBPS)
	}

	next = clampDecCoins(next, params.MinGasPrices, params.MaxGasPrices)

	state.CurrentMinGasPrices = next
	state.SmoothedUtilizationBPS = smoothed
	state.LastComputedUtilizationBPS = utilBPS
	state.LastBlockHeight = ctx.BlockHeight()

	if err := k.SetState(ctx, state); err != nil {
		return params.MinGasPrices, state, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"adaptive_gas_pricing",
			sdk.NewAttribute("block_height", fmt.Sprintf("%d", ctx.BlockHeight())),
			sdk.NewAttribute("gas_used", fmt.Sprintf("%d", gasUsed)),
			sdk.NewAttribute("gas_limit", fmt.Sprintf("%d", gasLimit)),
			sdk.NewAttribute("utilization_bps", fmt.Sprintf("%d", utilBPS)),
			sdk.NewAttribute("smoothed_utilization_bps", fmt.Sprintf("%d", smoothed)),
			sdk.NewAttribute("min_gas_prices", next.String()),
		),
	)

	k.logger.Debug("adaptive min gas updated",
		"height", ctx.BlockHeight(),
		"util_bps", utilBPS,
		"smoothed_bps", smoothed,
		"min_gas_prices", next.String(),
	)

	return next, state, nil
}
