// Package types contains types for the Fraud module.
//
// VE-912: Fraud reporting flow - Error definitions
package types

import (
	"cosmossdk.io/errors"
)

// Fraud module error codes
var (
	// ErrInvalidReporter is returned when the reporter address is invalid
	ErrInvalidReporter = errors.Register(ModuleName, 1, "invalid reporter address")

	// ErrInvalidReportedParty is returned when the reported party address is invalid
	ErrInvalidReportedParty = errors.Register(ModuleName, 2, "invalid reported party address")

	// ErrInvalidDescription is returned when the fraud description is invalid
	ErrInvalidDescription = errors.Register(ModuleName, 3, "invalid fraud description")

	// ErrInvalidEvidence is returned when the evidence is invalid or improperly encrypted
	ErrInvalidEvidence = errors.Register(ModuleName, 4, "invalid evidence: must be properly encrypted")

	// ErrReportNotFound is returned when a fraud report is not found
	ErrReportNotFound = errors.Register(ModuleName, 5, "fraud report not found")

	// ErrReportAlreadyResolved is returned when trying to modify a resolved report
	ErrReportAlreadyResolved = errors.Register(ModuleName, 6, "fraud report already resolved")

	// ErrUnauthorizedModerator is returned when a non-moderator tries to perform moderator actions
	ErrUnauthorizedModerator = errors.Register(ModuleName, 7, "unauthorized: moderator role required")

	// ErrUnauthorizedReporter is returned when a non-provider tries to submit a fraud report
	ErrUnauthorizedReporter = errors.Register(ModuleName, 8, "unauthorized: only providers can submit fraud reports")

	// ErrSelfReport is returned when a party tries to report themselves
	ErrSelfReport = errors.Register(ModuleName, 9, "invalid report: cannot report yourself")

	// ErrInvalidReportID is returned when a report ID is invalid
	ErrInvalidReportID = errors.Register(ModuleName, 10, "invalid report ID")

	// ErrInvalidStatus is returned when an invalid status transition is attempted
	ErrInvalidStatus = errors.Register(ModuleName, 11, "invalid status transition")

	// ErrReportNotInQueue is returned when a report is not in the moderator queue
	ErrReportNotInQueue = errors.Register(ModuleName, 12, "report not in moderator queue")

	// ErrAuditLogNotFound is returned when an audit log entry is not found
	ErrAuditLogNotFound = errors.Register(ModuleName, 13, "audit log entry not found")

	// ErrInvalidCategory is returned when the fraud category is invalid
	ErrInvalidCategory = errors.Register(ModuleName, 14, "invalid fraud category")

	// ErrInvalidResolution is returned when the resolution is invalid
	ErrInvalidResolution = errors.Register(ModuleName, 15, "invalid resolution")

	// ErrDescriptionTooLong is returned when description exceeds maximum length
	ErrDescriptionTooLong = errors.Register(ModuleName, 16, "description too long")

	// ErrDescriptionTooShort is returned when description is too short
	ErrDescriptionTooShort = errors.Register(ModuleName, 17, "description too short")

	// ErrMissingEvidence is returned when evidence is required but not provided
	ErrMissingEvidence = errors.Register(ModuleName, 18, "evidence is required for fraud reports")

	// ErrInvalidResolutionNotes is returned when resolution notes are invalid
	ErrInvalidResolutionNotes = errors.Register(ModuleName, 19, "invalid resolution notes")

	// ErrEvidenceDecryptionFailed is returned when evidence decryption fails
	ErrEvidenceDecryptionFailed = errors.Register(ModuleName, 20, "evidence decryption failed")

	// ErrInvalidEnvelope is returned when the encrypted evidence envelope is invalid
	ErrInvalidEnvelope = errors.Register(ModuleName, 21, "invalid encrypted evidence envelope")
)
