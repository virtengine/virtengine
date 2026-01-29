// Package types provides VEID module types.
//
// This file defines compliance-related message types for the VEID module.
//
// Task Reference: VE-3021 - KYC/AML Compliance Interface
package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Message type constants for compliance
const (
	TypeMsgSubmitComplianceCheck  = "submit_compliance_check"
	TypeMsgAttestCompliance       = "attest_compliance"
	TypeMsgUpdateComplianceParams = "update_compliance_params"
	TypeMsgRegisterProvider       = "register_compliance_provider"
	TypeMsgDeactivateProvider     = "deactivate_compliance_provider"
)

var (
	_ sdk.Msg = &MsgSubmitComplianceCheck{}
	_ sdk.Msg = &MsgAttestCompliance{}
	_ sdk.Msg = &MsgUpdateComplianceParams{}
	_ sdk.Msg = &MsgRegisterComplianceProvider{}
	_ sdk.Msg = &MsgDeactivateComplianceProvider{}
)

// ============================================================================
// MsgSubmitComplianceCheck
// ============================================================================

// MsgSubmitComplianceCheck submits external compliance check results
type MsgSubmitComplianceCheck struct {
	// ProviderAddress is the address of the compliance provider
	ProviderAddress string `json:"provider_address"`

	// TargetAddress is the address being checked
	TargetAddress string `json:"target_address"`

	// CheckResults contains the compliance check results
	CheckResults []ComplianceCheckResult `json:"check_results"`

	// ProviderID is the ID of the compliance provider
	ProviderID string `json:"provider_id"`
}

// NewMsgSubmitComplianceCheck creates a new MsgSubmitComplianceCheck
func NewMsgSubmitComplianceCheck(
	providerAddress string,
	targetAddress string,
	checkResults []ComplianceCheckResult,
	providerID string,
) *MsgSubmitComplianceCheck {
	return &MsgSubmitComplianceCheck{
		ProviderAddress: providerAddress,
		TargetAddress:   targetAddress,
		CheckResults:    checkResults,
		ProviderID:      providerID,
	}
}

// Route returns the route for the message
func (msg MsgSubmitComplianceCheck) Route() string { return RouterKey }

// Type returns the message type
func (msg MsgSubmitComplianceCheck) Type() string { return TypeMsgSubmitComplianceCheck }

// GetSigners returns the addresses of required signers
func (msg MsgSubmitComplianceCheck) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.ProviderAddress)
	return []sdk.AccAddress{addr}
}

// GetSignBytes returns the bytes to sign
func (msg MsgSubmitComplianceCheck) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic performs basic validation
func (msg MsgSubmitComplianceCheck) ValidateBasic() error {
	if msg.ProviderAddress == "" {
		return ErrNotComplianceProvider.Wrap("provider address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(msg.ProviderAddress); err != nil {
		return ErrInvalidAddress.Wrap("invalid provider address format")
	}

	if msg.TargetAddress == "" {
		return ErrInvalidAddress.Wrap("target address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(msg.TargetAddress); err != nil {
		return ErrInvalidAddress.Wrap("invalid target address format")
	}

	if len(msg.CheckResults) == 0 {
		return ErrComplianceCheckFailed.Wrap("at least one check result is required")
	}

	for i, result := range msg.CheckResults {
		if err := result.Validate(); err != nil {
			return ErrComplianceCheckFailed.Wrapf("invalid check result at index %d: %v", i, err)
		}
	}

	if msg.ProviderID == "" {
		return ErrNotComplianceProvider.Wrap("provider ID cannot be empty")
	}

	return nil
}

// ============================================================================
// MsgAttestCompliance
// ============================================================================

// MsgAttestCompliance allows validators to attest compliance status
type MsgAttestCompliance struct {
	// ValidatorAddress is the address of the attesting validator
	ValidatorAddress string `json:"validator_address"`

	// TargetAddress is the address being attested
	TargetAddress string `json:"target_address"`

	// AttestationType describes what is being attested
	AttestationType string `json:"attestation_type"`

	// ExpiryBlocks is how long until this attestation expires (in blocks)
	ExpiryBlocks int64 `json:"expiry_blocks,omitempty"`
}

// NewMsgAttestCompliance creates a new MsgAttestCompliance
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

// Route returns the route for the message
func (msg MsgAttestCompliance) Route() string { return RouterKey }

// Type returns the message type
func (msg MsgAttestCompliance) Type() string { return TypeMsgAttestCompliance }

// GetSigners returns the addresses of required signers
func (msg MsgAttestCompliance) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.ValidatorAddress)
	return []sdk.AccAddress{addr}
}

// GetSignBytes returns the bytes to sign
func (msg MsgAttestCompliance) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic performs basic validation
func (msg MsgAttestCompliance) ValidateBasic() error {
	if msg.ValidatorAddress == "" {
		return ErrInsufficientAttestations.Wrap("validator address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(msg.ValidatorAddress); err != nil {
		return ErrInvalidAddress.Wrap("invalid validator address format")
	}

	if msg.TargetAddress == "" {
		return ErrInvalidAddress.Wrap("target address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(msg.TargetAddress); err != nil {
		return ErrInvalidAddress.Wrap("invalid target address format")
	}

	if msg.AttestationType == "" {
		return ErrInsufficientAttestations.Wrap("attestation type cannot be empty")
	}

	if msg.ExpiryBlocks < 0 {
		return ErrComplianceExpired.Wrap("expiry blocks cannot be negative")
	}

	return nil
}

// ============================================================================
// MsgUpdateComplianceParams
// ============================================================================

// MsgUpdateComplianceParams updates compliance configuration (gov only)
type MsgUpdateComplianceParams struct {
	// Authority is the address that is authorized to update params (x/gov)
	Authority string `json:"authority"`

	// Params are the new compliance parameters
	Params ComplianceParams `json:"params"`
}

// NewMsgUpdateComplianceParams creates a new MsgUpdateComplianceParams
func NewMsgUpdateComplianceParams(authority string, params ComplianceParams) *MsgUpdateComplianceParams {
	return &MsgUpdateComplianceParams{
		Authority: authority,
		Params:    params,
	}
}

// Route returns the route for the message
func (msg MsgUpdateComplianceParams) Route() string { return RouterKey }

// Type returns the message type
func (msg MsgUpdateComplianceParams) Type() string { return TypeMsgUpdateComplianceParams }

// GetSigners returns the addresses of required signers
func (msg MsgUpdateComplianceParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}

// GetSignBytes returns the bytes to sign
func (msg MsgUpdateComplianceParams) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic performs basic validation
func (msg MsgUpdateComplianceParams) ValidateBasic() error {
	if msg.Authority == "" {
		return ErrInvalidComplianceParams.Wrap("authority cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidAddress.Wrap("invalid authority address format")
	}

	if err := msg.Params.Validate(); err != nil {
		return err
	}

	return nil
}

// ============================================================================
// MsgRegisterComplianceProvider
// ============================================================================

// MsgRegisterComplianceProvider registers a new compliance provider
type MsgRegisterComplianceProvider struct {
	// Authority is the address that is authorized to register providers (x/gov)
	Authority string `json:"authority"`

	// Provider is the compliance provider to register
	Provider ComplianceProvider `json:"provider"`
}

// NewMsgRegisterComplianceProvider creates a new MsgRegisterComplianceProvider
func NewMsgRegisterComplianceProvider(authority string, provider ComplianceProvider) *MsgRegisterComplianceProvider {
	return &MsgRegisterComplianceProvider{
		Authority: authority,
		Provider:  provider,
	}
}

// Route returns the route for the message
func (msg MsgRegisterComplianceProvider) Route() string { return RouterKey }

// Type returns the message type
func (msg MsgRegisterComplianceProvider) Type() string { return TypeMsgRegisterProvider }

// GetSigners returns the addresses of required signers
func (msg MsgRegisterComplianceProvider) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}

// GetSignBytes returns the bytes to sign
func (msg MsgRegisterComplianceProvider) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic performs basic validation
func (msg MsgRegisterComplianceProvider) ValidateBasic() error {
	if msg.Authority == "" {
		return ErrNotComplianceProvider.Wrap("authority cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidAddress.Wrap("invalid authority address format")
	}

	if err := msg.Provider.Validate(); err != nil {
		return err
	}

	return nil
}

// ============================================================================
// MsgDeactivateComplianceProvider
// ============================================================================

// MsgDeactivateComplianceProvider deactivates a compliance provider
type MsgDeactivateComplianceProvider struct {
	// Authority is the address that is authorized to deactivate providers (x/gov)
	Authority string `json:"authority"`

	// ProviderID is the ID of the provider to deactivate
	ProviderID string `json:"provider_id"`

	// Reason is the reason for deactivation
	Reason string `json:"reason,omitempty"`
}

// NewMsgDeactivateComplianceProvider creates a new MsgDeactivateComplianceProvider
func NewMsgDeactivateComplianceProvider(authority string, providerID string, reason string) *MsgDeactivateComplianceProvider {
	return &MsgDeactivateComplianceProvider{
		Authority:  authority,
		ProviderID: providerID,
		Reason:     reason,
	}
}

// Route returns the route for the message
func (msg MsgDeactivateComplianceProvider) Route() string { return RouterKey }

// Type returns the message type
func (msg MsgDeactivateComplianceProvider) Type() string { return TypeMsgDeactivateProvider }

// GetSigners returns the addresses of required signers
func (msg MsgDeactivateComplianceProvider) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}

// GetSignBytes returns the bytes to sign
func (msg MsgDeactivateComplianceProvider) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic performs basic validation
func (msg MsgDeactivateComplianceProvider) ValidateBasic() error {
	if msg.Authority == "" {
		return ErrNotComplianceProvider.Wrap("authority cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidAddress.Wrap("invalid authority address format")
	}

	if msg.ProviderID == "" {
		return ErrNotComplianceProvider.Wrap("provider ID cannot be empty")
	}

	return nil
}

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
