// Package types provides VEID module types.
//
// This file defines model versioning error codes for the VEID module.
// NOTE: ErrModelNotFound is aliased from veidv1 in errors.go - these are additional local errors.
//
// Task Reference: VE-3007 - Model Versioning and Governance
package types

import (
	errorsmod "cosmossdk.io/errors"
)

// Model versioning error codes (range: 6000-6099)
// NOTE: ErrModelNotFound is defined as an alias in errors.go from veidv1
var (
	// ErrModelHashMismatch is returned when a model hash doesn't match expected
	ErrModelHashMismatch = errorsmod.Register(ModuleName, 6002, "model hash mismatch")

	// ErrModelNotActive is returned when trying to use an inactive model
	ErrModelNotActive = errorsmod.Register(ModuleName, 6003, "model not active")

	// ErrInvalidModelType is returned when a model type is invalid
	ErrInvalidModelType = errorsmod.Register(ModuleName, 6004, "invalid model type")

	// ErrModelUpdatePending is returned when a model update is already pending
	ErrModelUpdatePending = errorsmod.Register(ModuleName, 6005, "model update already pending")

	// ErrUnauthorizedModelUpdate is returned when not authorized for model updates
	ErrUnauthorizedModelUpdate = errorsmod.Register(ModuleName, 6006, "not authorized for model updates")

	// ErrModelVersionTooOld is returned when a validator uses an outdated model
	ErrModelVersionTooOld = errorsmod.Register(ModuleName, 6007, "validator using outdated model version")

	// ErrModelActivationPending is returned when model activation is still pending
	ErrModelActivationPending = errorsmod.Register(ModuleName, 6008, "model activation still pending")

	// ErrModelAlreadyExists is returned when trying to register a duplicate model
	ErrModelAlreadyExists = errorsmod.Register(ModuleName, 6009, "model already exists")

	// ErrModelAlreadyActive is returned when trying to activate an already active model
	ErrModelAlreadyActive = errorsmod.Register(ModuleName, 6010, "model is already active")

	// ErrModelDeprecated is returned when trying to use a deprecated model
	ErrModelDeprecated = errorsmod.Register(ModuleName, 6011, "model has been deprecated")

	// ErrModelRevoked is returned when trying to use a revoked model
	ErrModelRevoked = errorsmod.Register(ModuleName, 6012, "model has been revoked")

	// ErrInvalidModelInfo is returned when model info is invalid
	ErrInvalidModelInfo = errorsmod.Register(ModuleName, 6013, "invalid model info")

	// ErrInvalidModelProposal is returned when a model proposal is invalid
	ErrInvalidModelProposal = errorsmod.Register(ModuleName, 6014, "invalid model update proposal")

	// ErrProposalNotFound is returned when a model proposal cannot be found
	ErrProposalNotFound = errorsmod.Register(ModuleName, 6015, "model proposal not found")

	// ErrProposalAlreadyProcessed is returned when a proposal was already processed
	ErrProposalAlreadyProcessed = errorsmod.Register(ModuleName, 6016, "model proposal already processed")

	// ErrActivationHeightNotReached is returned when activation height not yet reached
	ErrActivationHeightNotReached = errorsmod.Register(ModuleName, 6017, "activation height not yet reached")

	// ErrValidatorNotSynced is returned when validator models are not synced
	ErrValidatorNotSynced = errorsmod.Register(ModuleName, 6018, "validator model versions not synced")

	// ErrInvalidModelHash is returned when a model hash format is invalid
	ErrInvalidModelHash = errorsmod.Register(ModuleName, 6019, "invalid model hash format")

	// ErrModelParamsInvalid is returned when model parameters are invalid
	ErrModelParamsInvalid = errorsmod.Register(ModuleName, 6020, "invalid model parameters")

	// ErrGovernanceUpdateDisabled is returned when governance updates are disabled
	ErrGovernanceUpdateDisabled = errorsmod.Register(ModuleName, 6021, "governance model updates are disabled")

	// ErrRegistrarNotAllowed is returned when registrar is not in allowed list
	ErrRegistrarNotAllowed = errorsmod.Register(ModuleName, 6022, "registrar not in allowed list")

	// ErrModelTooOld is returned when a model exceeds max age
	ErrModelTooOld = errorsmod.Register(ModuleName, 6023, "model exceeds maximum age")

	// ErrSyncGracePeriodExpired is returned when validator sync grace period expired
	ErrSyncGracePeriodExpired = errorsmod.Register(ModuleName, 6024, "validator sync grace period expired")

	// ErrNoActiveModel is returned when no active model exists for a type
	ErrNoActiveModel = errorsmod.Register(ModuleName, 6025, "no active model for model type")

	// ErrInvalidHistoryEntry is returned when a history entry is invalid
	ErrInvalidHistoryEntry = errorsmod.Register(ModuleName, 6026, "invalid model version history entry")
)
