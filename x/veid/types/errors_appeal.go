// Package types provides VEID module types.
//
// This file defines appeal-related error codes for the VEID module.
//
// Task Reference: VE-3020 - Appeal and Dispute System
package types

import (
	errorsmod "cosmossdk.io/errors"
)

// Appeal error codes (range: 4000-4099)
var (
	// ErrAppealNotFound is returned when an appeal cannot be found
	ErrAppealNotFound = errorsmod.Register(ModuleName, 4001, "appeal not found")

	// ErrAppealAlreadyExists is returned when an appeal already exists for a scope
	ErrAppealAlreadyExists = errorsmod.Register(ModuleName, 4002, "appeal already exists for this scope")

	// ErrAppealNotPending is returned when trying to modify an appeal that is not pending
	ErrAppealNotPending = errorsmod.Register(ModuleName, 4003, "appeal is not in pending status")

	// ErrNotAppealSubmitter is returned when a non-submitter tries to withdraw an appeal
	ErrNotAppealSubmitter = errorsmod.Register(ModuleName, 4004, "not the appeal submitter")

	// ErrNotAuthorizedResolver is returned when an unauthorized account tries to resolve an appeal
	ErrNotAuthorizedResolver = errorsmod.Register(ModuleName, 4005, "not authorized to resolve appeals")

	// ErrInvalidAppealReason is returned when the appeal reason is too short or invalid
	ErrInvalidAppealReason = errorsmod.Register(ModuleName, 4006, "appeal reason too short or invalid")

	// ErrAppealWindowExpired is returned when the appeal submission window has expired
	ErrAppealWindowExpired = errorsmod.Register(ModuleName, 4007, "appeal submission window has expired")

	// ErrMaxAppealsExceeded is returned when the maximum appeals for a scope is exceeded
	ErrMaxAppealsExceeded = errorsmod.Register(ModuleName, 4008, "maximum appeals for this scope exceeded")

	// ErrInvalidAppealRecord is returned when an appeal record is malformed
	ErrInvalidAppealRecord = errorsmod.Register(ModuleName, 4009, "invalid appeal record")

	// ErrInvalidAppealResolution is returned when an invalid resolution status is provided
	ErrInvalidAppealResolution = errorsmod.Register(ModuleName, 4010, "invalid appeal resolution")

	// ErrAppealSystemDisabled is returned when the appeal system is disabled
	ErrAppealSystemDisabled = errorsmod.Register(ModuleName, 4011, "appeal system is disabled")

	// ErrAppealAlreadyClaimed is returned when trying to claim an already claimed appeal
	ErrAppealAlreadyClaimed = errorsmod.Register(ModuleName, 4012, "appeal already claimed by another reviewer")

	// ErrNotAppealReviewer is returned when a non-reviewer tries to resolve an appeal they didn't claim
	ErrNotAppealReviewer = errorsmod.Register(ModuleName, 4013, "not the assigned appeal reviewer")

	// ErrAppealReviewTimeout is returned when an appeal review has timed out
	ErrAppealReviewTimeout = errorsmod.Register(ModuleName, 4014, "appeal review timeout expired")

	// ErrInvalidEvidenceHash is returned when an evidence hash is malformed
	ErrInvalidEvidenceHash = errorsmod.Register(ModuleName, 4015, "invalid evidence hash format")

	// ErrScopeNotRejected is returned when trying to appeal a scope that was not rejected
	ErrScopeNotRejected = errorsmod.Register(ModuleName, 4016, "scope was not rejected")

	// ErrAppealEscrowRequired is returned when escrow deposit is required but not provided
	ErrAppealEscrowRequired = errorsmod.Register(ModuleName, 4017, "escrow deposit required for appeal")

	// ErrInvalidScoreAdjustment is returned when an invalid score adjustment is provided
	ErrInvalidScoreAdjustment = errorsmod.Register(ModuleName, 4018, "invalid score adjustment value")
)
