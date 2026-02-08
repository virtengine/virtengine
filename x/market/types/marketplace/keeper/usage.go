package keeper

import (
	"fmt"
	"math"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

const (
	usageKeyCPUHoursMilli       = "cpu_hours_milli"
	usageKeyGPUHoursMilli       = "gpu_hours_milli"
	usageKeyRAMGBHoursMilli     = "ram_gb_hours_milli"
	usageKeyStorageGBHoursMilli = "storage_gb_hours_milli"
	usageKeyNetworkGBMilli      = "network_gb_milli"
)

func (k Keeper) processUsageReportCallback(ctx sdk.Context, callback *marketplace.WaldurCallback) error {
	if callback.ChainEntityType != marketplace.SyncTypeAllocation {
		return marketplace.ErrWaldurCallbackInvalid.Wrapf("unexpected entity type: %s", callback.ChainEntityType)
	}

	allocationID, err := marketplace.ParseAllocationID(callback.ChainEntityID)
	if err != nil {
		return marketplace.ErrWaldurCallbackInvalid.Wrap(err.Error())
	}

	allocation, found := k.GetAllocation(ctx, allocationID)
	if !found {
		return marketplace.ErrAllocationNotFound.Wrapf("allocation %s not found", callback.ChainEntityID)
	}

	if allocation.TotalUsage == nil {
		allocation.TotalUsage = make(map[string]uint64)
	}

	applyUsageValue(allocation.TotalUsage, usageKeyCPUHoursMilli, callback.Payload["usage_cpu_hours"])
	applyUsageValue(allocation.TotalUsage, usageKeyGPUHoursMilli, callback.Payload["usage_gpu_hours"])
	applyUsageValue(allocation.TotalUsage, usageKeyRAMGBHoursMilli, callback.Payload["usage_ram_gb_hours"])
	applyUsageValue(allocation.TotalUsage, usageKeyStorageGBHoursMilli, callback.Payload["usage_storage_gb_hours"])
	applyUsageValue(allocation.TotalUsage, usageKeyNetworkGBMilli, callback.Payload["usage_network_gb"])

	now := ctx.BlockTime().UTC()
	allocation.UsageLastReportedAt = &now

	if err := k.UpdateAllocation(ctx, allocation); err != nil {
		return err
	}

	return nil
}

func applyUsageValue(total map[string]uint64, key, raw string) {
	if total == nil || raw == "" {
		return
	}
	value, err := parseUsageFloat(raw)
	if err != nil || value <= 0 {
		return
	}
	// store in milli-units for deterministic integer accounting.
	milli := uint64(math.Round(value * 1000))
	total[key] += milli
}

func parseUsageFloat(raw string) (float64, error) {
	if raw == "" {
		return 0, fmt.Errorf("empty usage value")
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, err
	}
	if value < 0 {
		return 0, fmt.Errorf("negative usage value")
	}
	return value, nil
}
