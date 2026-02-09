package keeper

import (
	"encoding/json"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/settlement/types"
)

const settlementBaseCurrency = "VRT"

var settlementFiatQuotes = []string{"USD", "EUR", "GBP"}

// LockSettlementRates locks oracle rates for a settlement invoice.
func (k Keeper) LockSettlementRates(ctx sdk.Context, settlementID, invoiceID string) (*types.SettlementRateLock, error) {
	pairs := make([]types.CurrencyPair, 0, len(settlementFiatQuotes))
	for _, quote := range settlementFiatQuotes {
		pairs = append(pairs, types.CurrencyPair{Base: settlementBaseCurrency, Quote: quote})
	}

	prices, err := k.AggregatePrices(ctx, pairs)
	if err != nil {
		lock := types.SettlementRateLock{
			SettlementID: settlementID,
			InvoiceID:    invoiceID,
			Status:       types.SettlementRateStatusPending,
			LockedAt:     ctx.BlockTime(),
			Reason:       err.Error(),
		}
		_ = k.SetSettlementRateLock(ctx, lock)
		_ = k.QueueSettlementRateLock(ctx, lock)
		return &lock, types.ErrRateUnavailable.Wrap(err.Error())
	}

	spreadBps := k.fiatConversionSpreadBps(ctx)
	spreadDec := sdkmath.LegacyNewDec(int64(spreadBps)).QuoInt64(10000)
	multiplier := sdkmath.LegacyOneDec().Add(spreadDec)

	locked := make([]types.LockedRate, 0, len(prices))
	for _, price := range prices {
		finalRate := price.Rate.Mul(multiplier)
		locked = append(locked, types.LockedRate{
			Base:      price.Base,
			Quote:     price.Quote,
			Source:    price.Source,
			RawRate:   price.Rate,
			SpreadBps: spreadBps,
			FinalRate: finalRate,
			LockedAt:  ctx.BlockTime(),
		})
	}

	lock := types.SettlementRateLock{
		SettlementID: settlementID,
		InvoiceID:    invoiceID,
		Status:       types.SettlementRateStatusLocked,
		Rates:        locked,
		LockedAt:     ctx.BlockTime(),
	}

	if err := k.SetSettlementRateLock(ctx, lock); err != nil {
		return nil, err
	}
	return &lock, nil
}

// SetSettlementRateLock stores a settlement rate lock.
func (k Keeper) SetSettlementRateLock(ctx sdk.Context, lock types.SettlementRateLock) error {
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(&lock)
	if err != nil {
		return err
	}
	store.Set(types.SettlementRateLockKey(lock.SettlementID), bz)
	return nil
}

// QueueSettlementRateLock stores a settlement rate lock in the pending queue.
func (k Keeper) QueueSettlementRateLock(ctx sdk.Context, lock types.SettlementRateLock) error {
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(&lock)
	if err != nil {
		return err
	}
	store.Set(types.SettlementRateQueueKey(lock.SettlementID), bz)
	return nil
}

// GetSettlementRateLock retrieves a settlement rate lock.
func (k Keeper) GetSettlementRateLock(ctx sdk.Context, settlementID string) (types.SettlementRateLock, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.SettlementRateLockKey(settlementID))
	if bz == nil {
		return types.SettlementRateLock{}, false
	}
	var lock types.SettlementRateLock
	if err := json.Unmarshal(bz, &lock); err != nil {
		return types.SettlementRateLock{}, false
	}
	return lock, true
}

// SettlementRateSummary provides a compact description of locked rates.
func SettlementRateSummary(lock types.SettlementRateLock) string {
	if len(lock.Rates) == 0 {
		return "no rates"
	}
	return fmt.Sprintf("%s:%d", lock.Status, len(lock.Rates))
}
