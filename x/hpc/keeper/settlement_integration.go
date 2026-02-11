package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	settlementtypes "github.com/virtengine/virtengine/x/settlement/types"
)

// SettlementKeeper defines the subset of settlement functionality used by HPC.
type SettlementKeeper interface {
	RecordUsage(ctx sdk.Context, record *settlementtypes.UsageRecord) error
	SettleOrder(ctx sdk.Context, orderID string, usageRecordIDs []string, isFinal bool) (*settlementtypes.SettlementRecord, error)
	GetEscrowByOrder(ctx sdk.Context, orderID string) (settlementtypes.EscrowAccount, bool)
	GetEscrow(ctx sdk.Context, escrowID string) (settlementtypes.EscrowAccount, bool)
}

// SetSettlementKeeper configures the settlement integration keeper.
func (k *Keeper) SetSettlementKeeper(settlementKeeper SettlementKeeper) {
	k.settlementKeeper = settlementKeeper
}
