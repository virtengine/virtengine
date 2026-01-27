// Package types contains types for the staking module.
//
// VE-921: Staking rewards errors
package types

import (
	sdkerrors "cosmossdk.io/errors"
)

// Staking module sentinel errors
var (
	// ErrValidatorNotFound is returned when validator is not found
	ErrValidatorNotFound = sdkerrors.Register(ModuleName, 2400, "validator not found")

	// ErrInvalidValidator is returned when validator address is invalid
	ErrInvalidValidator = sdkerrors.Register(ModuleName, 2401, "invalid validator")

	// ErrSlashingAlreadyRecorded is returned when a slashing event was already recorded
	ErrSlashingAlreadyRecorded = sdkerrors.Register(ModuleName, 2402, "slashing already recorded")

	// ErrInvalidSlashReason is returned when slashing reason is invalid
	ErrInvalidSlashReason = sdkerrors.Register(ModuleName, 2403, "invalid slashing reason")

	// ErrInvalidEpoch is returned when epoch is invalid
	ErrInvalidEpoch = sdkerrors.Register(ModuleName, 2404, "invalid epoch")

	// ErrRewardsAlreadyDistributed is returned when rewards were already distributed
	ErrRewardsAlreadyDistributed = sdkerrors.Register(ModuleName, 2405, "rewards already distributed for epoch")

	// ErrInvalidParams is returned when parameters are invalid
	ErrInvalidParams = sdkerrors.Register(ModuleName, 2406, "invalid parameters")

	// ErrValidatorJailed is returned when validator is jailed
	ErrValidatorJailed = sdkerrors.Register(ModuleName, 2407, "validator is jailed")

	// ErrInvalidPerformanceMetric is returned when performance metric is invalid
	ErrInvalidPerformanceMetric = sdkerrors.Register(ModuleName, 2408, "invalid performance metric")

	// ErrInsufficientStake is returned when stake is insufficient
	ErrInsufficientStake = sdkerrors.Register(ModuleName, 2409, "insufficient stake")

	// ErrDoubleSign is returned when double signing is detected
	ErrDoubleSign = sdkerrors.Register(ModuleName, 2410, "double signing detected")

	// ErrInvalidAttestation is returned when VEID attestation is invalid
	ErrInvalidAttestation = sdkerrors.Register(ModuleName, 2411, "invalid VEID attestation")

	// ErrDowntime is returned when validator has excessive downtime
	ErrDowntime = sdkerrors.Register(ModuleName, 2412, "excessive downtime")
)
