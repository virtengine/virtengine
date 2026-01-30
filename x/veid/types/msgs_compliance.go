// Package types provides VEID module types.
//
// This file defines compliance-related message types for the VEID module.
// It uses type aliases to the buf-generated protobuf types in sdk/go/node/veid/v1.
//
// Task Reference: VE-3021 - KYC/AML Compliance Interface
package types

import (
	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
)

// Message type constants for compliance - kept for backwards compatibility
const (
	TypeMsgSubmitComplianceCheck  = "submit_compliance_check"
	TypeMsgAttestCompliance       = "attest_compliance"
	TypeMsgUpdateComplianceParams = "update_compliance_params"
	TypeMsgRegisterProvider       = "register_compliance_provider"
	TypeMsgDeactivateProvider     = "deactivate_compliance_provider"
)

// ============================================================================
// Compliance Message Type Aliases
// These types alias the buf-generated protobuf types which already implement
// proto.Message and sdk.Msg interfaces.
// ============================================================================

// MsgSubmitComplianceCheck submits external compliance check results
type MsgSubmitComplianceCheck = veidv1.MsgSubmitComplianceCheck

// MsgSubmitComplianceCheckResponse is the response for MsgSubmitComplianceCheck
type MsgSubmitComplianceCheckResponse = veidv1.MsgSubmitComplianceCheckResponse

// MsgAttestCompliance allows validators to attest compliance status
type MsgAttestCompliance = veidv1.MsgAttestCompliance

// MsgAttestComplianceResponse is the response for MsgAttestCompliance
type MsgAttestComplianceResponse = veidv1.MsgAttestComplianceResponse

// MsgUpdateComplianceParams updates compliance configuration (gov only)
type MsgUpdateComplianceParams = veidv1.MsgUpdateComplianceParams

// MsgUpdateComplianceParamsResponse is the response for MsgUpdateComplianceParams
type MsgUpdateComplianceParamsResponse = veidv1.MsgUpdateComplianceParamsResponse

// MsgRegisterComplianceProvider registers a new compliance provider
type MsgRegisterComplianceProvider = veidv1.MsgRegisterComplianceProvider

// MsgRegisterComplianceProviderResponse is the response for MsgRegisterComplianceProvider
type MsgRegisterComplianceProviderResponse = veidv1.MsgRegisterComplianceProviderResponse

// MsgDeactivateComplianceProvider deactivates a compliance provider
type MsgDeactivateComplianceProvider = veidv1.MsgDeactivateComplianceProvider

// MsgDeactivateComplianceProviderResponse is the response for MsgDeactivateComplianceProvider
type MsgDeactivateComplianceProviderResponse = veidv1.MsgDeactivateComplianceProviderResponse

// ============================================================================
// Query Types
// ============================================================================

// QueryComplianceStatusRequest is the request for QueryComplianceStatus
type QueryComplianceStatusRequest struct {
	Address string `json:"address"`
}

// QueryComplianceStatusResponse is the response for QueryComplianceStatus
type QueryComplianceStatusResponse struct {
	Record *ComplianceRecord `json:"record,omitempty"`
	Found  bool              `json:"found"`
}

// QueryComplianceParamsRequest is the request for QueryComplianceParams
type QueryComplianceParamsRequest struct{}

// QueryComplianceParamsResponse is the response for QueryComplianceParams
type QueryComplianceParamsResponse struct {
	Params ComplianceParams `json:"params"`
}

// QueryComplianceProviderRequest is the request for QueryComplianceProvider
type QueryComplianceProviderRequest struct {
	ProviderID string `json:"provider_id"`
}

// QueryComplianceProviderResponse is the response for QueryComplianceProvider
type QueryComplianceProviderResponse struct {
	Provider *ComplianceProvider `json:"provider,omitempty"`
	Found    bool                `json:"found"`
}

// QueryComplianceProvidersRequest is the request for listing all providers
type QueryComplianceProvidersRequest struct {
	ActiveOnly bool `json:"active_only"`
}

// QueryComplianceProvidersResponse is the response for listing all providers
type QueryComplianceProvidersResponse struct {
	Providers []ComplianceProvider `json:"providers"`
}
