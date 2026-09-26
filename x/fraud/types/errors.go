// Package types contains types for the Fraud module.
//
// VE-912: Fraud reporting flow - Error definitions
package types

import (
	"cosmossdk.io/errors"

	fraudv1 "github.com/virtengine/virtengine/sdk/go/node/fraud/v1"
)

// Re-export error codes from SDK to avoid duplicate registration
// The SDK registers these errors first, so we alias them here
var (
	// ErrInvalidReporter is returned when the reporter address is invalid
	ErrInvalidReporter = fraudv1.ErrInvalidReporter

	// ErrInvalidReportedParty is returned when the reported party address is invalid
	ErrInvalidReportedParty = fraudv1.ErrInvalidReportedParty

	// ErrUnauthorizedModerator is returned when a non-moderator tries to perform moderator actions
	ErrUnauthorizedModerator = fraudv1.ErrUnauthorizedModerator

	// ErrSelfReport is returned when a party tries to report themselves
	ErrSelfReport = fraudv1.ErrSelfReport

	// ErrInvalidReportID is returned when a report ID is invalid
	ErrInvalidReportID = fraudv1.ErrInvalidReportID

	// ErrInvalidStatus is returned when an invalid status transition is attempted
	ErrInvalidStatus = fraudv1.ErrInvalidStatus

	// ErrInvalidCategory is returned when the fraud category is invalid
	ErrInvalidCategory = fraudv1.ErrInvalidCategory

	// ErrInvalidResolution is returned when the resolution is invalid
	ErrInvalidResolution = fraudv1.ErrInvalidResolution

	// ErrDescriptionTooLong is returned when description exceeds maximum length
	ErrDescriptionTooLong = fraudv1.ErrDescriptionTooLong

	// ErrDescriptionTooShort is returned when description is too short
	ErrDescriptionTooShort = fraudv1.ErrDescriptionTooShort

	// ErrMissingEvidence is returned when evidence is required but not provided
	ErrMissingEvidence = fraudv1.ErrMissingEvidence

	// ErrInvalidResolutionNotes is returned when resolution notes are invalid
	ErrInvalidResolutionNotes = fraudv1.ErrInvalidResolutionNotes
)

// Module-specific errors (not in SDK, unique error codes)
var (
	// ErrInvalidDescription is returned when the fraud description is invalid
	ErrInvalidDescription = errors.Register(ModuleName, 3002, "invalid fraud description")

	// ErrInvalidEvidence is returned when the evidence is invalid or improperly encrypted
	ErrInvalidEvidence = errors.Register(ModuleName, 3003, "invalid evidence: must be properly encrypted")

	// ErrReportNotFound is returned when a fraud report is not found
	ErrReportNotFound = errors.Register(ModuleName, 3004, "fraud report not found")

	// ErrReportAlreadyResolved is returned when trying to modify a resolved report
	ErrReportAlreadyResolved = errors.Register(ModuleName, 3005, "fraud report already resolved")

	// ErrUnauthorizedReporter is returned when a non-provider tries to submit a fraud report
	ErrUnauthorizedReporter = errors.Register(ModuleName, 3007, "unauthorized: only providers can submit fraud reports")

	// ErrReportNotInQueue is returned when a report is not in the moderator queue
	ErrReportNotInQueue = errors.Register(ModuleName, 3011, "report not in moderator queue")

	// ErrAuditLogNotFound is returned when an audit log entry is not found
	ErrAuditLogNotFound = errors.Register(ModuleName, 3012, "audit log entry not found")

	// ErrEvidenceDecryptionFailed is returned when evidence decryption fails
	ErrEvidenceDecryptionFailed = errors.Register(ModuleName, 3019, "evidence decryption failed")

	// ErrInvalidEnvelope is returned when the encrypted evidence envelope is invalid
	ErrInvalidEnvelope = errors.Register(ModuleName, 3020, "invalid encrypted evidence envelope")

	// ErrInvalidEscalationReason is returned when escalation reason is missing or invalid
	ErrInvalidEscalationReason = errors.Register(ModuleName, 3021, "invalid escalation reason")

	// ErrInvalidAuthority is returned when the authority address is invalid
	ErrInvalidAuthority = errors.Register(ModuleName, 3022, "invalid authority address")

	// ErrMissingOrderReference is returned when a non-provider reporter submits a
	// report without any order or resource reference. Every affected participant
	// must anchor their report to the transaction they were party to, so a
	// moderator can verify standing without trusting the reporter's narrative.
	ErrMissingOrderReference = errors.Register(ModuleName, 3008, "report requires an order or resource reference")

	// ErrDuplicateReport is returned when a reporter resubmits a byte-identical
	// report. The original report stays in the moderator queue; the duplicate
	// never creates a second queue entry (spam control).
	ErrDuplicateReport = errors.Register(ModuleName, 3009, "duplicate fraud report")

	// ErrReporterRateLimited is returned when a reporter exceeds the per-window
	// submission limit (spam control).
	ErrReporterRateLimited = errors.Register(ModuleName, 3010, "reporter submission rate limit exceeded")

	// ErrResponseNotFound is returned when a fraud response record is not found
	ErrResponseNotFound = errors.Register(ModuleName, 3013, "fraud response not found")

	// ErrInvalidResponse is returned when a fraud response record is malformed
	ErrInvalidResponse = errors.Register(ModuleName, 3014, "invalid fraud response")

	// ErrUnauthorizedRespondent is returned when a party that is neither the
	// reported party nor the original reporter tries to file a response
	ErrUnauthorizedRespondent = errors.Register(ModuleName, 3015, "unauthorized: only the reported party or the reporter may respond")

	// ErrReportNotPending is returned when a response is filed against a report
	// that is already in a terminal state
	ErrReportNotPending = errors.Register(ModuleName, 3016, "fraud report is not pending")

	// ErrOrderVerificationUnavailable is returned when a non-provider report cites
	// an order but the fraud keeper has no market keeper wired, so reporter
	// standing cannot be verified.
	//
	// This is a node-configuration fault, not a reporter fault, and is kept
	// distinct from ErrUnauthorizedReporter/ErrMissingOrderReference so operators
	// can tell "the reporter has no standing" apart from "this binary forgot to
	// wire SetMarketKeeper". The check fails closed: an unverifiable claim never
	// grants standing.
	ErrOrderVerificationUnavailable = errors.Register(ModuleName, 3023, "order-standing verification unavailable: market keeper is not wired")
)
