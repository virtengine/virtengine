package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	settlementkeeper "github.com/virtengine/virtengine/x/settlement/keeper"
	settlementtypes "github.com/virtengine/virtengine/x/settlement/types"
)

// SettlementKeeper defines the subset of settlement functionality used by HPC.
type SettlementKeeper interface {
	RecordUsage(ctx sdk.Context, record *settlementtypes.UsageRecord) error
	SettleOrder(ctx sdk.Context, orderID string, usageRecordIDs []string, isFinal bool) (*settlementtypes.SettlementRecord, error)
	GetEscrowByOrder(ctx sdk.Context, orderID string) (settlementtypes.EscrowAccount, bool)
	GetEscrow(ctx sdk.Context, escrowID string) (settlementtypes.EscrowAccount, bool)
	GetUsageRecord(ctx sdk.Context, usageID string) (settlementtypes.UsageRecord, bool)
	OpenFinancialCase(ctx sdk.Context, request settlementkeeper.FinancialCaseOpenRequest) (*settlementtypes.FinancialCase, *settlementtypes.FinancialClaim, bool, error)
	AddFinancialClaim(ctx sdk.Context, caseID string, claim settlementtypes.FinancialClaim) (*settlementtypes.FinancialCase, *settlementtypes.FinancialClaim, bool, error)
	EscalateFinancialCase(ctx sdk.Context, caseID, actor string, reasonHash []byte) error
	GetFinancialCase(ctx sdk.Context, caseID string) (settlementtypes.FinancialCase, bool)
	GetFinancialCaseBySubject(ctx sdk.Context, subject settlementtypes.FinancialSubject) (settlementtypes.FinancialCase, bool)
	IsFinancialCasesActive(ctx sdk.Context) bool
}

// SetSettlementKeeper configures the settlement integration keeper.
func (k *Keeper) SetSettlementKeeper(settlementKeeper SettlementKeeper) {
	k.settlementKeeper = settlementKeeper
}
