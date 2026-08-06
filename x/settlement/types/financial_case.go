package types

import settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"

// Type aliases keep the module API tied directly to generated Task 86A contracts.
type (
	FinancialSubject        = settlementv1.FinancialSubject
	FinancialClaim          = settlementv1.FinancialClaim
	FinancialExposure       = settlementv1.FinancialExposure
	TerminalAllocation      = settlementv1.TerminalAllocation
	FinancialAppeal         = settlementv1.FinancialAppeal
	FinancialCaseTransition = settlementv1.FinancialCaseTransition
	FinancialCaseEffect     = settlementv1.FinancialCaseEffect
	FinancialCase           = settlementv1.FinancialCase
	FinancialSubjectType    = settlementv1.FinancialSubjectType
	FinancialClaimType      = settlementv1.FinancialClaimType
	FinancialCaseStatus     = settlementv1.FinancialCaseStatus
	FinancialResolutionType = settlementv1.FinancialResolutionType
	FinancialEffectType     = settlementv1.FinancialEffectType
	FinancialEffectStatus   = settlementv1.FinancialEffectStatus
)

const (
	FinancialSubjectTypeOrder      = settlementv1.FinancialSubjectType_FINANCIAL_SUBJECT_TYPE_ORDER
	FinancialSubjectTypeInvoice    = settlementv1.FinancialSubjectType_FINANCIAL_SUBJECT_TYPE_INVOICE
	FinancialSubjectTypeUsage      = settlementv1.FinancialSubjectType_FINANCIAL_SUBJECT_TYPE_USAGE
	FinancialSubjectTypeHPCJob     = settlementv1.FinancialSubjectType_FINANCIAL_SUBJECT_TYPE_HPC_JOB
	FinancialSubjectTypeSettlement = settlementv1.FinancialSubjectType_FINANCIAL_SUBJECT_TYPE_SETTLEMENT

	FinancialClaimTypeBilling    = settlementv1.FinancialClaimType_FINANCIAL_CLAIM_TYPE_BILLING
	FinancialClaimTypeUsage      = settlementv1.FinancialClaimType_FINANCIAL_CLAIM_TYPE_USAGE
	FinancialClaimTypeService    = settlementv1.FinancialClaimType_FINANCIAL_CLAIM_TYPE_SERVICE
	FinancialClaimTypeFraud      = settlementv1.FinancialClaimType_FINANCIAL_CLAIM_TYPE_FRAUD
	FinancialClaimTypeHPC        = settlementv1.FinancialClaimType_FINANCIAL_CLAIM_TYPE_HPC
	FinancialClaimTypeReview     = settlementv1.FinancialClaimType_FINANCIAL_CLAIM_TYPE_REVIEW
	FinancialClaimTypeModeration = settlementv1.FinancialClaimType_FINANCIAL_CLAIM_TYPE_MODERATION
	FinancialClaimTypeMigration  = settlementv1.FinancialClaimType_FINANCIAL_CLAIM_TYPE_MIGRATION

	FinancialCaseStatusOpen                  = settlementv1.FinancialCaseStatus_FINANCIAL_CASE_STATUS_OPEN
	FinancialCaseStatusEvidence              = settlementv1.FinancialCaseStatus_FINANCIAL_CASE_STATUS_EVIDENCE
	FinancialCaseStatusReview                = settlementv1.FinancialCaseStatus_FINANCIAL_CASE_STATUS_REVIEW
	FinancialCaseStatusEscalated             = settlementv1.FinancialCaseStatus_FINANCIAL_CASE_STATUS_ESCALATED
	FinancialCaseStatusResolvedPendingAppeal = settlementv1.FinancialCaseStatus_FINANCIAL_CASE_STATUS_RESOLVED_PENDING_APPEAL
	FinancialCaseStatusFinal                 = settlementv1.FinancialCaseStatus_FINANCIAL_CASE_STATUS_FINAL
	FinancialCaseStatusRejected              = settlementv1.FinancialCaseStatus_FINANCIAL_CASE_STATUS_REJECTED
	FinancialCaseStatusCancelled             = settlementv1.FinancialCaseStatus_FINANCIAL_CASE_STATUS_CANCELLED
	FinancialCaseStatusExpired               = settlementv1.FinancialCaseStatus_FINANCIAL_CASE_STATUS_EXPIRED
	FinancialCaseStatusQuarantined           = settlementv1.FinancialCaseStatus_FINANCIAL_CASE_STATUS_QUARANTINED

	FinancialResolutionProviderWin         = settlementv1.FinancialResolutionType_FINANCIAL_RESOLUTION_TYPE_PROVIDER_WIN
	FinancialResolutionCustomerWin         = settlementv1.FinancialResolutionType_FINANCIAL_RESOLUTION_TYPE_CUSTOMER_WIN
	FinancialResolutionPartialSplit        = settlementv1.FinancialResolutionType_FINANCIAL_RESOLUTION_TYPE_PARTIAL_SPLIT
	FinancialResolutionMutual              = settlementv1.FinancialResolutionType_FINANCIAL_RESOLUTION_TYPE_MUTUAL
	FinancialResolutionFraudConfirmed      = settlementv1.FinancialResolutionType_FINANCIAL_RESOLUTION_TYPE_FRAUD_CONFIRMED
	FinancialResolutionInconclusiveTimeout = settlementv1.FinancialResolutionType_FINANCIAL_RESOLUTION_TYPE_INCONCLUSIVE_TIMEOUT

	FinancialEffectPayout      = settlementv1.FinancialEffectType_FINANCIAL_EFFECT_TYPE_PAYOUT
	FinancialEffectEscrow      = settlementv1.FinancialEffectType_FINANCIAL_EFFECT_TYPE_ESCROW
	FinancialEffectReward      = settlementv1.FinancialEffectType_FINANCIAL_EFFECT_TYPE_REWARD
	FinancialEffectReservation = settlementv1.FinancialEffectType_FINANCIAL_EFFECT_TYPE_RESERVATION
	FinancialEffectReputation  = settlementv1.FinancialEffectType_FINANCIAL_EFFECT_TYPE_REPUTATION
	FinancialEffectProjection  = settlementv1.FinancialEffectType_FINANCIAL_EFFECT_TYPE_PROJECTION

	FinancialEffectStatusPending = settlementv1.FinancialEffectStatus_FINANCIAL_EFFECT_STATUS_PENDING
	FinancialEffectStatusApplied = settlementv1.FinancialEffectStatus_FINANCIAL_EFFECT_STATUS_APPLIED
	FinancialEffectStatusFailed  = settlementv1.FinancialEffectStatus_FINANCIAL_EFFECT_STATUS_FAILED
)

func IsActiveFinancialCaseStatus(status FinancialCaseStatus) bool {
	switch status {
	case FinancialCaseStatusOpen, FinancialCaseStatusEvidence, FinancialCaseStatusReview,
		FinancialCaseStatusEscalated, FinancialCaseStatusResolvedPendingAppeal, FinancialCaseStatusQuarantined:
		return true
	default:
		return false
	}
}

func IsTerminalFinancialCaseStatus(status FinancialCaseStatus) bool {
	return status == FinancialCaseStatusFinal || status == FinancialCaseStatusRejected ||
		status == FinancialCaseStatusCancelled || status == FinancialCaseStatusExpired
}
