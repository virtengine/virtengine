// Package types provides VEID module types.
//
// This file defines compliance-related error codes for the VEID module.
//
// Task Reference: VE-3021 - KYC/AML Compliance Interface
package types

import (
	errorsmod "cosmossdk.io/errors"
)

// Compliance error codes (range: 5000-5099)
var (
	// ErrComplianceCheckFailed is returned when a compliance check fails
	ErrComplianceCheckFailed = errorsmod.Register(ModuleName, 5001, "compliance check failed")

	// ErrSanctionListMatch is returned when identity matches a sanction list
	ErrSanctionListMatch = errorsmod.Register(ModuleName, 5002, "sanction list match detected")

	// ErrPEPMatch is returned when identity matches a politically exposed person
	ErrPEPMatch = errorsmod.Register(ModuleName, 5003, "politically exposed person match")

	// ErrRestrictedRegion is returned when identity is from a restricted region
	ErrRestrictedRegion = errorsmod.Register(ModuleName, 5004, "identity from restricted region")

	// ErrComplianceExpired is returned when a compliance check has expired
	ErrComplianceExpired = errorsmod.Register(ModuleName, 5005, "compliance check expired")

	// ErrRiskScoreExceeded is returned when risk score exceeds the threshold
	ErrRiskScoreExceeded = errorsmod.Register(ModuleName, 5006, "risk score exceeds threshold")

	// ErrComplianceNotFound is returned when a compliance record is not found
	ErrComplianceNotFound = errorsmod.Register(ModuleName, 5007, "compliance record not found")

	// ErrNotComplianceProvider is returned when sender is not an authorized provider
	ErrNotComplianceProvider = errorsmod.Register(ModuleName, 5008, "not authorized compliance provider")

	// ErrInsufficientAttestations is returned when there are not enough validator attestations
	ErrInsufficientAttestations = errorsmod.Register(ModuleName, 5009, "insufficient compliance attestations")

	// ErrDuplicateAttestation is returned when a validator tries to attest twice
	ErrDuplicateAttestation = errorsmod.Register(ModuleName, 5010, "validator has already attested")

	// ErrComplianceRecordBlocked is returned when trying to modify a blocked record
	ErrComplianceRecordBlocked = errorsmod.Register(ModuleName, 5011, "compliance record is blocked")

	// ErrInvalidComplianceParams is returned when compliance parameters are invalid
	ErrInvalidComplianceParams = errorsmod.Register(ModuleName, 5012, "invalid compliance parameters")

	// ErrProviderAlreadyExists is returned when trying to register an existing provider
	ErrProviderAlreadyExists = errorsmod.Register(ModuleName, 5013, "compliance provider already exists")

	// ErrProviderNotActive is returned when an inactive provider tries to submit checks
	ErrProviderNotActive = errorsmod.Register(ModuleName, 5014, "compliance provider is not active")

	// ErrUnsupportedCheckType is returned when provider doesn't support the check type
	ErrUnsupportedCheckType = errorsmod.Register(ModuleName, 5015, "provider does not support this check type")

	// ErrAdverseMediaMatch is returned when adverse media is found
	ErrAdverseMediaMatch = errorsmod.Register(ModuleName, 5016, "adverse media match found")

	// ErrWatchlistMatch is returned when identity matches a watchlist entry
	ErrWatchlistMatch = errorsmod.Register(ModuleName, 5017, "watchlist match found")

	// ErrDocumentVerificationFailed is returned when document verification fails
	ErrDocumentVerificationFailed = errorsmod.Register(ModuleName, 5018, "document verification failed")

	// ErrAMLRiskHigh is returned when AML risk assessment is too high
	ErrAMLRiskHigh = errorsmod.Register(ModuleName, 5019, "AML risk level is too high")

	// ErrComplianceSystemDisabled is returned when the compliance system is disabled
	ErrComplianceSystemDisabled = errorsmod.Register(ModuleName, 5020, "compliance system is disabled")
)
