package keeper

import (
	"sort"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/settlement/types"
)

// BuildUsageSummary aggregates usage records based on filters.
func (k Keeper) BuildUsageSummary(
	ctx sdk.Context,
	orderID string,
	provider string,
	periodStart time.Time,
	periodEnd time.Time,
) (types.UsageSummary, error) {
	if provider != "" {
		if _, err := sdk.AccAddressFromBech32(provider); err != nil {
			return types.UsageSummary{}, types.ErrInvalidAddress.Wrap("invalid provider address")
		}
	}

	var (
		totalUnits uint64
		totalCost  = sdk.NewCoins()
		byType     = make(map[string]*types.UsageTypeSummary)
		recordIDs  []string
		minStart   time.Time
		maxEnd     time.Time
	)

	k.WithUsageRecords(ctx, func(record types.UsageRecord) bool {
		if orderID != "" && record.OrderID != orderID {
			return false
		}
		if provider != "" && record.Provider != provider {
			return false
		}
		if !periodStart.IsZero() && record.PeriodEnd.Before(periodStart) {
			return false
		}
		if !periodEnd.IsZero() && record.PeriodStart.After(periodEnd) {
			return false
		}

		totalUnits += record.UsageUnits
		totalCost = totalCost.Add(record.TotalCost...)
		recordIDs = append(recordIDs, record.UsageID)

		if minStart.IsZero() || record.PeriodStart.Before(minStart) {
			minStart = record.PeriodStart
		}
		if maxEnd.IsZero() || record.PeriodEnd.After(maxEnd) {
			maxEnd = record.PeriodEnd
		}

		usageType := record.UsageType
		if usageType == "" {
			usageType = "unknown"
		}
		summary, ok := byType[usageType]
		if !ok {
			summary = &types.UsageTypeSummary{
				UsageType: usageType,
				TotalCost: sdk.NewCoins(),
			}
			byType[usageType] = summary
		}
		summary.UsageUnits += record.UsageUnits
		summary.TotalCost = summary.TotalCost.Add(record.TotalCost...)

		return false
	})

	usageTypes := make([]types.UsageTypeSummary, 0, len(byType))
	for _, summary := range byType {
		usageTypes = append(usageTypes, *summary)
	}
	sort.Slice(usageTypes, func(i, j int) bool {
		return usageTypes[i].UsageType < usageTypes[j].UsageType
	})

	if periodStart.IsZero() && !minStart.IsZero() {
		periodStart = minStart
	}
	if periodEnd.IsZero() && !maxEnd.IsZero() {
		periodEnd = maxEnd
	}

	return types.UsageSummary{
		Provider:       provider,
		OrderID:        orderID,
		PeriodStart:    periodStart,
		PeriodEnd:      periodEnd,
		TotalUsage:     totalUnits,
		TotalCost:      totalCost,
		ByUsageType:    usageTypes,
		GeneratedAt:    ctx.BlockTime(),
		BlockHeight:    ctx.BlockHeight(),
		UsageRecordIDs: recordIDs,
	}, nil
}
