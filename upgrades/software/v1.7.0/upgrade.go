// Package v1_7_0 activates Task 84D canonical financial cases.
package v1_7_0

import (
	"context"
	"crypto/sha256"
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	apptypes "github.com/virtengine/virtengine/app/types"
	utypes "github.com/virtengine/virtengine/upgrades/types"
	billingtypes "github.com/virtengine/virtengine/x/escrow/types/billing"
	fraudtypes "github.com/virtengine/virtengine/x/fraud/types"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
	settlementkeeper "github.com/virtengine/virtengine/x/settlement/keeper"
	settlementtypes "github.com/virtengine/virtengine/x/settlement/types"
)

const UpgradeName = utypes.CanonicalFinancialCasesUpgradeName

type upgrade struct {
	*apptypes.App
	log log.Logger
}

var _ utypes.IUpgrade = (*upgrade)(nil)

func initUpgrade(logger log.Logger, app *apptypes.App) (utypes.IUpgrade, error) {
	return &upgrade{App: app, log: logger.With("module", "upgrade/"+UpgradeName)}, nil
}
func (*upgrade) StoreLoader() *storetypes.StoreUpgrades { return &storetypes.StoreUpgrades{} }

func (up *upgrade) UpgradeHandler() upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		if financialCaseVersionsCurrent(fromVM) && up.Keepers.VirtEngine.Settlement.IsFinancialCasesActive(sdkCtx) {
			return cloneVersionMap(fromVM), nil
		}
		cacheCtx, write := sdkCtx.CacheContext()
		report, err := up.Keepers.VirtEngine.Settlement.MigrateFinancialCases(cacheCtx)
		if err != nil {
			return nil, err
		}
		billingClaims, fraudClaims, hpcClaims, quarantined, err := up.reconcileAdapters(cacheCtx)
		if err != nil {
			return nil, err
		}
		toVM := cloneVersionMap(fromVM)
		if !financialCaseVersionsCurrent(fromVM) {
			toVM, err = up.MM.RunMigrations(cacheCtx, up.Configurator, fromVM)
			if err != nil {
				return nil, err
			}
		}
		if err := up.Keepers.VirtEngine.Settlement.RebuildFinancialCaseState(cacheCtx); err != nil {
			return nil, err
		}
		up.Keepers.VirtEngine.Settlement.ActivateFinancialCases(cacheCtx)
		if broken := up.Keepers.VirtEngine.Settlement.ValidateFinancialCaseInvariants(cacheCtx); len(broken) != 0 {
			return nil, fmt.Errorf("post-upgrade financial-case invariant: %v", broken)
		}
		write()
		up.log.Info("canonical financial cases activated", "settlement_cases", report.CasesCreated, "settlement_quarantined", report.Quarantined, "billing_claims", billingClaims, "fraud_claims", fraudClaims, "hpc_claims", hpcClaims, "adapter_quarantined", quarantined, "digest", report.Digest)
		return toVM, nil
	}
}

func financialCaseVersionsCurrent(versionMap module.VersionMap) bool {
	return versionMap[settlementtypes.ModuleName] >= 3 &&
		versionMap[fraudtypes.ModuleName] >= 2 &&
		versionMap[hpctypes.ModuleName] >= 3 &&
		versionMap["review"] >= 2 &&
		versionMap["escrow"] >= 4 &&
		versionMap["resources"] >= 3
}

func (up *upgrade) reconcileAdapters(ctx sdk.Context) (billingClaims, fraudClaims, hpcClaims, quarantined uint64, reconcileErr error) {
	up.Keepers.VirtEngine.Escrow.NewDisputeKeeper().WithDisputes(ctx, func(workflow *billingtypes.DisputeWorkflow) bool {
		if workflow == nil || workflow.Status == billingtypes.DisputeStatusResolved || workflow.Status == billingtypes.DisputeStatusClosed || workflow.Status == billingtypes.DisputeStatusExpired {
			return false
		}
		invoice, err := up.Keepers.VirtEngine.Escrow.NewInvoiceKeeper().GetInvoice(ctx, workflow.InvoiceID)
		if err != nil || invoice == nil {
			quarantined++
			return false
		}
		respondent := invoice.Provider
		if workflow.InitiatedBy == invoice.Provider {
			respondent = invoice.Customer
		}
		hash := sha256.Sum256([]byte(workflow.DisputeID + "\x00" + workflow.InvoiceID))
		idempotencyHash := sha256.Sum256([]byte("migration/billing/v1\x00" + workflow.DisputeID))
		idempotency := idempotencyHash[:]
		financialCase, _, duplicate, openErr := up.Keepers.VirtEngine.Settlement.OpenFinancialCase(ctx, settlementkeeper.FinancialCaseOpenRequest{
			Subject:  settlementtypes.FinancialSubject{Type: settlementtypes.FinancialSubjectTypeInvoice, PrimaryId: workflow.InvoiceID, InvoiceId: workflow.InvoiceID, OrderId: invoice.OrderID, SettlementId: invoice.SettlementID, EscrowId: invoice.EscrowID, LeaseId: invoice.LeaseID},
			Claimant: workflow.InitiatedBy, Respondent: respondent, IdempotencyKey: idempotency, TrustedAdapter: true, Migrated: true,
			Claim: settlementtypes.FinancialClaim{ClaimType: settlementtypes.FinancialClaimTypeMigration, Claimant: workflow.InitiatedBy, SourceModule: "escrow", SourceReference: workflow.DisputeID, EvidenceHash: hash[:], EncryptedReference: "migration://escrow/dispute/" + workflow.DisputeID, IdempotencyKey: idempotency},
		})
		if openErr != nil {
			reconcileErr = openErr
			return true
		}
		if !duplicate {
			billingClaims++
		}
		if financialCase.Quarantined {
			quarantined++
		}
		return false
	})
	if reconcileErr != nil {
		return
	}
	up.Keepers.VirtEngine.Fraud.WithFraudReports(ctx, func(report fraudtypes.FraudReport) bool {
		if len(report.RelatedOrderIDs) == 0 || !report.Status.IsPending() {
			return false
		}
		orderID := report.RelatedOrderIDs[0]
		hash := sha256.Sum256([]byte(report.ContentHash + "\x00" + report.ID))
		idempotencyHash := sha256.Sum256([]byte("migration/fraud/v1\x00" + report.ID + "\x00" + orderID))
		idempotency := idempotencyHash[:]
		financialCase, _, duplicate, err := up.Keepers.VirtEngine.Settlement.OpenFinancialCase(ctx, settlementkeeper.FinancialCaseOpenRequest{
			Subject: settlementtypes.FinancialSubject{Type: settlementtypes.FinancialSubjectTypeOrder, PrimaryId: orderID, OrderId: orderID}, Claimant: report.Reporter, Respondent: report.ReportedParty, IdempotencyKey: idempotency, TrustedAdapter: true, Migrated: true,
			Claim: settlementtypes.FinancialClaim{ClaimType: settlementtypes.FinancialClaimTypeFraud, Claimant: report.Reporter, SourceModule: "fraud", SourceReference: report.ID, EvidenceHash: hash[:], EncryptedReference: "migration://fraud/" + report.ID, IdempotencyKey: idempotency},
		})
		if err != nil {
			reconcileErr = err
			return true
		}
		if !duplicate {
			fraudClaims++
		}
		report.FinancialCaseID, report.FinancialCaseStatus = financialCase.CaseId, financialCase.Status.String()
		if err := up.Keepers.VirtEngine.Fraud.SetFraudReport(ctx, report); err != nil {
			reconcileErr = err
			return true
		}
		return false
	})
	if reconcileErr != nil {
		return
	}
	up.Keepers.VirtEngine.HPC.WithDisputes(ctx, func(disputeType hpctypes.HPCDispute) bool {
		if disputeType.Status != hpctypes.DisputeStatusPending && disputeType.Status != hpctypes.DisputeStatusUnderReview {
			return false
		}
		job, found := up.Keepers.VirtEngine.HPC.GetJob(ctx, disputeType.JobID)
		if !found {
			quarantined++
			return false
		}
		respondent := job.ProviderAddress
		if disputeType.DisputerAddress == job.ProviderAddress {
			respondent = job.CustomerAddress
		}
		hash := sha256.Sum256([]byte(disputeType.DisputeID + "\x00" + disputeType.Evidence))
		idempotencyHash := sha256.Sum256([]byte("migration/hpc/v1\x00" + disputeType.DisputeID))
		idempotency := idempotencyHash[:]
		financialCase, _, duplicate, err := up.Keepers.VirtEngine.Settlement.OpenFinancialCase(ctx, settlementkeeper.FinancialCaseOpenRequest{
			Subject: settlementtypes.FinancialSubject{Type: settlementtypes.FinancialSubjectTypeHPCJob, PrimaryId: job.JobID, HpcJobId: job.JobID, OrderId: job.MarketOrderID, EscrowId: job.EscrowID, ReservationId: job.ReservationID, LeaseId: job.MarketLeaseID}, Claimant: disputeType.DisputerAddress, Respondent: respondent, IdempotencyKey: idempotency, TrustedAdapter: true, Migrated: true,
			Claim: settlementtypes.FinancialClaim{ClaimType: settlementtypes.FinancialClaimTypeHPC, Claimant: disputeType.DisputerAddress, SourceModule: "hpc", SourceReference: disputeType.DisputeID, EvidenceHash: hash[:], EncryptedReference: "migration://hpc/" + disputeType.DisputeID, IdempotencyKey: idempotency},
		})
		if err != nil {
			reconcileErr = err
			return true
		}
		if !duplicate {
			hpcClaims++
		}
		disputeType.FinancialCaseID, disputeType.FinancialCaseStatus = financialCase.CaseId, financialCase.Status.String()
		if err := up.Keepers.VirtEngine.HPC.SetDispute(ctx, disputeType); err != nil {
			reconcileErr = err
			return true
		}
		return false
	})
	return
}

func cloneVersionMap(source module.VersionMap) module.VersionMap {
	result := make(module.VersionMap, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}
