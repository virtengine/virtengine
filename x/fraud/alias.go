// Package fraud implements the Fraud module for VirtEngine.
//
// VE-912: Fraud reporting flow - Module aliases
package fraud

import (
	"github.com/virtengine/virtengine/x/fraud/keeper"
	"github.com/virtengine/virtengine/x/fraud/types"
)

// Module constants
const (
	ModuleName   = types.ModuleName
	StoreKey     = types.StoreKey
	RouterKey    = types.RouterKey
	QuerierRoute = types.QuerierRoute
)

// Type aliases
type (
	Keeper = keeper.Keeper

	// Genesis types
	GenesisState = types.GenesisState
	Params       = types.Params

	// Fraud report types
	FraudReport       = types.FraudReport
	FraudReportStatus = types.FraudReportStatus
	FraudCategory     = types.FraudCategory
	ResolutionType    = types.ResolutionType
	EncryptedEvidence = types.EncryptedEvidence

	// Audit types
	FraudAuditLog = types.FraudAuditLog
	AuditAction   = types.AuditAction

	// Queue types
	ModeratorQueueEntry = types.ModeratorQueueEntry

	// Message types
	MsgSubmitFraudReport   = types.MsgSubmitFraudReport
	MsgAssignModerator     = types.MsgAssignModerator
	MsgUpdateReportStatus  = types.MsgUpdateReportStatus
	MsgResolveFraudReport  = types.MsgResolveFraudReport
	MsgRejectFraudReport   = types.MsgRejectFraudReport
	MsgEscalateFraudReport = types.MsgEscalateFraudReport
	MsgUpdateParams        = types.MsgUpdateParams
)

// Status constants
const (
	FraudReportStatusUnspecified = types.FraudReportStatusUnspecified
	FraudReportStatusSubmitted   = types.FraudReportStatusSubmitted
	FraudReportStatusReviewing   = types.FraudReportStatusReviewing
	FraudReportStatusResolved    = types.FraudReportStatusResolved
	FraudReportStatusRejected    = types.FraudReportStatusRejected
	FraudReportStatusEscalated   = types.FraudReportStatusEscalated
)

// Category constants
const (
	FraudCategoryUnspecified             = types.FraudCategoryUnspecified
	FraudCategoryFakeIdentity            = types.FraudCategoryFakeIdentity
	FraudCategoryPaymentFraud            = types.FraudCategoryPaymentFraud
	FraudCategoryServiceMisrepresentation = types.FraudCategoryServiceMisrepresentation
	FraudCategoryResourceAbuse           = types.FraudCategoryResourceAbuse
	FraudCategorySybilAttack             = types.FraudCategorySybilAttack
	FraudCategoryMaliciousContent        = types.FraudCategoryMaliciousContent
	FraudCategoryTermsViolation          = types.FraudCategoryTermsViolation
	FraudCategoryOther                   = types.FraudCategoryOther
)

// Resolution constants
const (
	ResolutionTypeUnspecified = types.ResolutionTypeUnspecified
	ResolutionTypeWarning     = types.ResolutionTypeWarning
	ResolutionTypeSuspension  = types.ResolutionTypeSuspension
	ResolutionTypeTermination = types.ResolutionTypeTermination
	ResolutionTypeRefund      = types.ResolutionTypeRefund
	ResolutionTypeNoAction    = types.ResolutionTypeNoAction
)

// Audit action constants
const (
	AuditActionUnspecified    = types.AuditActionUnspecified
	AuditActionSubmitted      = types.AuditActionSubmitted
	AuditActionAssigned       = types.AuditActionAssigned
	AuditActionStatusChanged  = types.AuditActionStatusChanged
	AuditActionEvidenceViewed = types.AuditActionEvidenceViewed
	AuditActionResolved       = types.AuditActionResolved
	AuditActionRejected       = types.AuditActionRejected
	AuditActionEscalated      = types.AuditActionEscalated
	AuditActionCommentAdded   = types.AuditActionCommentAdded
)

// Function aliases
var (
	NewKeeper = keeper.NewKeeper

	// Genesis
	DefaultGenesisState = types.DefaultGenesisState
	DefaultParams       = types.DefaultParams

	// Messages
	NewMsgSubmitFraudReport   = types.NewMsgSubmitFraudReport
	NewMsgAssignModerator     = types.NewMsgAssignModerator
	NewMsgUpdateReportStatus  = types.NewMsgUpdateReportStatus
	NewMsgResolveFraudReport  = types.NewMsgResolveFraudReport
	NewMsgRejectFraudReport   = types.NewMsgRejectFraudReport
	NewMsgEscalateFraudReport = types.NewMsgEscalateFraudReport
	NewMsgUpdateParams        = types.NewMsgUpdateParams

	// Types
	NewFraudReport          = types.NewFraudReport
	NewFraudAuditLog        = types.NewFraudAuditLog
	NewModeratorQueueEntry  = types.NewModeratorQueueEntry
	FraudCategoryFromString = types.FraudCategoryFromString
)

// Error aliases
var (
	ErrInvalidReporter        = types.ErrInvalidReporter
	ErrInvalidReportedParty   = types.ErrInvalidReportedParty
	ErrInvalidDescription     = types.ErrInvalidDescription
	ErrInvalidEvidence        = types.ErrInvalidEvidence
	ErrReportNotFound         = types.ErrReportNotFound
	ErrReportAlreadyResolved  = types.ErrReportAlreadyResolved
	ErrUnauthorizedModerator  = types.ErrUnauthorizedModerator
	ErrUnauthorizedReporter   = types.ErrUnauthorizedReporter
	ErrSelfReport             = types.ErrSelfReport
	ErrInvalidReportID        = types.ErrInvalidReportID
	ErrInvalidStatus          = types.ErrInvalidStatus
	ErrReportNotInQueue       = types.ErrReportNotInQueue
	ErrAuditLogNotFound       = types.ErrAuditLogNotFound
	ErrInvalidCategory        = types.ErrInvalidCategory
	ErrInvalidResolution      = types.ErrInvalidResolution
	ErrDescriptionTooLong     = types.ErrDescriptionTooLong
	ErrDescriptionTooShort    = types.ErrDescriptionTooShort
	ErrMissingEvidence        = types.ErrMissingEvidence
	ErrInvalidResolutionNotes = types.ErrInvalidResolutionNotes
)
