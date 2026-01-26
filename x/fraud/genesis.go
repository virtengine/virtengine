// Package fraud implements the Fraud module for VirtEngine.
//
// VE-912: Fraud reporting flow - Genesis init/export
package fraud

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"pkg.akt.dev/node/x/fraud/keeper"
	"pkg.akt.dev/node/x/fraud/types"
)

// InitGenesis initializes the fraud module's state from a provided genesis state
func InitGenesis(ctx sdk.Context, k keeper.Keeper, gs *types.GenesisState) {
	// Set module parameters
	if err := k.SetParams(ctx, gs.Params); err != nil {
		panic(err)
	}

	// Set next sequences
	k.SetNextFraudReportSequence(ctx, gs.NextFraudReportSequence)
	k.SetNextAuditLogSequence(ctx, gs.NextAuditLogSequence)

	// Import fraud reports
	for _, report := range gs.FraudReports {
		if err := k.SetFraudReport(ctx, report); err != nil {
			panic(err)
		}
	}

	// Import audit logs
	for _, log := range gs.AuditLogs {
		if err := k.CreateAuditLog(ctx, &log); err != nil {
			panic(err)
		}
	}

	// Import moderator queue
	for _, entry := range gs.ModeratorQueue {
		if err := k.AddToModeratorQueue(ctx, entry); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis exports the fraud module's state to a genesis state
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	var reports []types.FraudReport
	k.WithFraudReports(ctx, func(r types.FraudReport) bool {
		reports = append(reports, r)
		return false
	})

	var logs []types.FraudAuditLog
	k.WithAuditLogs(ctx, func(l types.FraudAuditLog) bool {
		logs = append(logs, l)
		return false
	})

	var queue []types.ModeratorQueueEntry
	k.WithModeratorQueue(ctx, func(e types.ModeratorQueueEntry) bool {
		queue = append(queue, e)
		return false
	})

	return &types.GenesisState{
		Params:                  k.GetParams(ctx),
		FraudReports:            reports,
		AuditLogs:               logs,
		ModeratorQueue:          queue,
		NextFraudReportSequence: k.GetNextFraudReportSequence(ctx),
		NextAuditLogSequence:    k.GetNextAuditLogSequence(ctx),
	}
}
