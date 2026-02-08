package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// UsageTypeSummary aggregates usage by type.
type UsageTypeSummary struct {
	UsageType  string    `json:"usage_type"`
	UsageUnits uint64    `json:"usage_units"`
	TotalCost  sdk.Coins `json:"total_cost"`
}

// UsageSummary provides aggregated usage information.
type UsageSummary struct {
	Provider       string             `json:"provider,omitempty"`
	OrderID        string             `json:"order_id,omitempty"`
	PeriodStart    time.Time          `json:"period_start"`
	PeriodEnd      time.Time          `json:"period_end"`
	TotalUsage     uint64             `json:"total_usage_units"`
	TotalCost      sdk.Coins          `json:"total_cost"`
	ByUsageType    []UsageTypeSummary `json:"by_usage_type"`
	GeneratedAt    time.Time          `json:"generated_at"`
	BlockHeight    int64              `json:"block_height"`
	UsageRecordIDs []string           `json:"usage_record_ids,omitempty"`
}

// RewardHistoryEntry represents a reward distribution entry for an address.
type RewardHistoryEntry struct {
	DistributionID string       `json:"distribution_id"`
	EpochNumber    uint64       `json:"epoch_number"`
	Source         RewardSource `json:"source"`
	Amount         sdk.Coins    `json:"amount"`
	Reason         string       `json:"reason"`
	UsageUnits     uint64       `json:"usage_units,omitempty"`
	ReferenceID    string       `json:"reference_id,omitempty"`
	DistributedAt  time.Time    `json:"distributed_at"`
}
