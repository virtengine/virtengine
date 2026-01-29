// Package types provides VEID module types.
//
// This file defines type aliases for compliance-related VEID messages, using the
// proto-generated types from sdk/go/node/veid/v1 as the source of truth.
//
// Task Reference: VE-3021 - KYC/AML Compliance Interface
package types

import (
	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
)

// Message type constants for compliance
const (
	TypeMsgSubmitComplianceCheck  = "submit_compliance_check"
	TypeMsgAttestCompliance       = "attest_compliance"
	TypeMsgUpdateComplianceParams = "update_compliance_params"
	TypeMsgRegisterProvider       = "register_compliance_provider"
	TypeMsgDeactivateProvider     = "deactivate_compliance_provider"
)

// ============================================================================
// Compliance Message Type Aliases - from proto-generated types
// ============================================================================

// MsgSubmitComplianceCheck submits external compliance check results.
type MsgSubmitComplianceCheck = veidv1.MsgSubmitComplianceCheck

// MsgSubmitComplianceCheckResponse is the response for MsgSubmitComplianceCheck.
type MsgSubmitComplianceCheckResponse = veidv1.MsgSubmitComplianceCheckResponse

// MsgAttestCompliance allows validators to attest compliance status.
type MsgAttestCompliance = veidv1.MsgAttestCompliance

// MsgAttestComplianceResponse is the response for MsgAttestCompliance.
type MsgAttestComplianceResponse = veidv1.MsgAttestComplianceResponse

// MsgUpdateComplianceParams updates compliance configuration (gov only).
type MsgUpdateComplianceParams = veidv1.MsgUpdateComplianceParams

// MsgUpdateComplianceParamsResponse is the response for MsgUpdateComplianceParams.
type MsgUpdateComplianceParamsResponse = veidv1.MsgUpdateComplianceParamsResponse

// MsgRegisterComplianceProvider registers a new compliance provider.
type MsgRegisterComplianceProvider = veidv1.MsgRegisterComplianceProvider

// MsgRegisterComplianceProviderResponse is the response for MsgRegisterComplianceProvider.
type MsgRegisterComplianceProviderResponse = veidv1.MsgRegisterComplianceProviderResponse

// MsgDeactivateComplianceProvider deactivates a compliance provider.
type MsgDeactivateComplianceProvider = veidv1.MsgDeactivateComplianceProvider

// MsgDeactivateComplianceProviderResponse is the response for MsgDeactivateComplianceProvider.
type MsgDeactivateComplianceProviderResponse = veidv1.MsgDeactivateComplianceProviderResponse

// ============================================================================
// Compliance Types - from proto-generated types
// ============================================================================

// ComplianceStatus represents the compliance status enum.
type ComplianceStatus = veidv1.ComplianceStatus

// ComplianceCheckType represents the type of compliance check.
type ComplianceCheckType = veidv1.ComplianceCheckType

// ComplianceCheckResult contains the result of a compliance check.
type ComplianceCheckResult = veidv1.ComplianceCheckResult

// ComplianceAttestation represents a validator's attestation.
type ComplianceAttestation = veidv1.ComplianceAttestation

// ComplianceRecord stores the compliance status for an address.
type ComplianceRecord = veidv1.ComplianceRecord

// ComplianceParams contains the compliance module parameters.
type ComplianceParams = veidv1.ComplianceParams

// ComplianceProvider represents a registered compliance provider.
type ComplianceProvider = veidv1.ComplianceProvider

// ============================================================================
// Compliance Constructor Functions
// ============================================================================

// NewMsgSubmitComplianceCheck creates a new MsgSubmitComplianceCheck.
func NewMsgSubmitComplianceCheck(
	providerAddress string,
	targetAddress string,
	checkResults []*ComplianceCheckResult,
	providerID string,
) *MsgSubmitComplianceCheck {
	return &MsgSubmitComplianceCheck{
		ProviderAddress: providerAddress,
		TargetAddress:   targetAddress,
		CheckResults:    checkResults,
		ProviderId:      providerID,
	}
}

// NewMsgAttestCompliance creates a new MsgAttestCompliance.
func NewMsgAttestCompliance(
	validatorAddress string,
	targetAddress string,
	attestationType string,
) *MsgAttestCompliance {
	return &MsgAttestCompliance{
		ValidatorAddress: validatorAddress,
		TargetAddress:    targetAddress,
		AttestationType:  attestationType,
	}
}

// NewMsgUpdateComplianceParams creates a new MsgUpdateComplianceParams.
func NewMsgUpdateComplianceParams(authority string, params *ComplianceParams) *MsgUpdateComplianceParams {
	return &MsgUpdateComplianceParams{
		Authority: authority,
		Params:    params,
	}
}

// NewMsgRegisterComplianceProvider creates a new MsgRegisterComplianceProvider.
func NewMsgRegisterComplianceProvider(authority string, provider *ComplianceProvider) *MsgRegisterComplianceProvider {
	return &MsgRegisterComplianceProvider{
		Authority: authority,
		Provider:  provider,
	}
}

// NewMsgDeactivateComplianceProvider creates a new MsgDeactivateComplianceProvider.
func NewMsgDeactivateComplianceProvider(authority string, providerID string, reason string) *MsgDeactivateComplianceProvider {
	return &MsgDeactivateComplianceProvider{
		Authority:  authority,
		ProviderId: providerID,
		Reason:     reason,
	}
}
