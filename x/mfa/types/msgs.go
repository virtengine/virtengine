package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgServer is the server API for the MFA module's messages
type MsgServer interface {
	// EnrollFactor enrolls a new MFA factor
	EnrollFactor(ctx context.Context, msg *MsgEnrollFactor) (*MsgEnrollFactorResponse, error)
	// RevokeFactor revokes an enrolled factor
	RevokeFactor(ctx context.Context, msg *MsgRevokeFactor) (*MsgRevokeFactorResponse, error)
	// SetMFAPolicy sets the MFA policy for an account
	SetMFAPolicy(ctx context.Context, msg *MsgSetMFAPolicy) (*MsgSetMFAPolicyResponse, error)
	// CreateChallenge creates an MFA challenge
	CreateChallenge(ctx context.Context, msg *MsgCreateChallenge) (*MsgCreateChallengeResponse, error)
	// VerifyChallenge verifies an MFA challenge response
	VerifyChallenge(ctx context.Context, msg *MsgVerifyChallenge) (*MsgVerifyChallengeResponse, error)
	// AddTrustedDevice adds a trusted device
	AddTrustedDevice(ctx context.Context, msg *MsgAddTrustedDevice) (*MsgAddTrustedDeviceResponse, error)
	// RemoveTrustedDevice removes a trusted device
	RemoveTrustedDevice(ctx context.Context, msg *MsgRemoveTrustedDevice) (*MsgRemoveTrustedDeviceResponse, error)
	// UpdateSensitiveTxConfig updates sensitive transaction configuration
	UpdateSensitiveTxConfig(ctx context.Context, msg *MsgUpdateSensitiveTxConfig) (*MsgUpdateSensitiveTxConfigResponse, error)
}

// MsgEnrollFactor is the message for enrolling a new factor
type MsgEnrollFactor struct {
	// Sender is the account enrolling the factor
	Sender string `json:"sender"`

	// FactorType is the type of factor to enroll
	FactorType FactorType `json:"factor_type"`

	// Label is a user-friendly label for the factor
	Label string `json:"label"`

	// PublicIdentifier is the public component (for FIDO2)
	PublicIdentifier []byte `json:"public_identifier,omitempty"`

	// Metadata contains factor-specific metadata
	Metadata *FactorMetadata `json:"metadata,omitempty"`

	// InitialVerificationProof is proof of factor possession during enrollment
	// For FIDO2: attestation object
	// For OTP factors: initial OTP verification
	InitialVerificationProof []byte `json:"initial_verification_proof,omitempty"`
}

// MsgEnrollFactorResponse is the response for MsgEnrollFactor
type MsgEnrollFactorResponse struct {
	// FactorID is the unique identifier for the enrolled factor
	FactorID string `json:"factor_id"`

	// Status is the enrollment status
	Status FactorEnrollmentStatus `json:"status"`
}

// MsgRevokeFactor is the message for revoking a factor
type MsgRevokeFactor struct {
	// Sender is the account revoking the factor
	Sender string `json:"sender"`

	// FactorType is the type of factor to revoke
	FactorType FactorType `json:"factor_type"`

	// FactorID is the ID of the factor to revoke
	FactorID string `json:"factor_id"`

	// MFAProof is proof of MFA for this operation (required if MFA is enabled)
	MFAProof *MFAProof `json:"mfa_proof,omitempty"`
}

// MsgRevokeFactorResponse is the response for MsgRevokeFactor
type MsgRevokeFactorResponse struct {
	// Success indicates if the revocation was successful
	Success bool `json:"success"`
}

// MsgSetMFAPolicy is the message for setting MFA policy
type MsgSetMFAPolicy struct {
	// Sender is the account setting the policy
	Sender string `json:"sender"`

	// Policy is the new MFA policy
	Policy MFAPolicy `json:"policy"`

	// MFAProof is proof of MFA for this operation (required if MFA is already enabled)
	MFAProof *MFAProof `json:"mfa_proof,omitempty"`
}

// MsgSetMFAPolicyResponse is the response for MsgSetMFAPolicy
type MsgSetMFAPolicyResponse struct {
	// Success indicates if the policy update was successful
	Success bool `json:"success"`
}

// MsgCreateChallenge is the message for creating an MFA challenge
type MsgCreateChallenge struct {
	// Sender is the account requesting the challenge
	Sender string `json:"sender"`

	// FactorType is the type of factor to challenge
	FactorType FactorType `json:"factor_type"`

	// FactorID is the specific factor to challenge (optional)
	FactorID string `json:"factor_id,omitempty"`

	// TransactionType is the sensitive transaction this challenge is for
	TransactionType SensitiveTransactionType `json:"transaction_type"`

	// ClientInfo contains client information
	ClientInfo *ClientInfo `json:"client_info,omitempty"`
}

// MsgCreateChallengeResponse is the response for MsgCreateChallenge
type MsgCreateChallengeResponse struct {
	// ChallengeID is the unique identifier for the challenge
	ChallengeID string `json:"challenge_id"`

	// ChallengeData contains factor-specific challenge data
	ChallengeData []byte `json:"challenge_data,omitempty"`

	// ExpiresAt is when the challenge expires
	ExpiresAt int64 `json:"expires_at"`
}

// MsgVerifyChallenge is the message for verifying a challenge response
type MsgVerifyChallenge struct {
	// Sender is the account verifying the challenge
	Sender string `json:"sender"`

	// ChallengeID is the challenge being verified
	ChallengeID string `json:"challenge_id"`

	// Response is the challenge response
	Response *ChallengeResponse `json:"response"`
}

// MsgVerifyChallengeResponse is the response for MsgVerifyChallenge
type MsgVerifyChallengeResponse struct {
	// Verified indicates if verification was successful
	Verified bool `json:"verified"`

	// SessionID is the authorization session ID if verified
	SessionID string `json:"session_id,omitempty"`

	// SessionExpiresAt is when the session expires
	SessionExpiresAt int64 `json:"session_expires_at,omitempty"`

	// RemainingFactors lists factors still needed to satisfy policy
	RemainingFactors []FactorType `json:"remaining_factors,omitempty"`
}

// MsgAddTrustedDevice is the message for adding a trusted device
type MsgAddTrustedDevice struct {
	// Sender is the account adding the device
	Sender string `json:"sender"`

	// DeviceInfo contains the device information
	DeviceInfo DeviceInfo `json:"device_info"`

	// MFAProof is proof of MFA for this operation
	MFAProof *MFAProof `json:"mfa_proof"`
}

// MsgAddTrustedDeviceResponse is the response for MsgAddTrustedDevice
type MsgAddTrustedDeviceResponse struct {
	// Success indicates if the device was added
	Success bool `json:"success"`

	// TrustExpiresAt is when the device trust expires
	TrustExpiresAt int64 `json:"trust_expires_at"`

	// TrustToken is the one-time trust token to be stored by the client
	TrustToken string `json:"trust_token,omitempty"`
}

// MsgRemoveTrustedDevice is the message for removing a trusted device
type MsgRemoveTrustedDevice struct {
	// Sender is the account removing the device
	Sender string `json:"sender"`

	// DeviceFingerprint is the fingerprint of the device to remove
	DeviceFingerprint string `json:"device_fingerprint"`

	// MFAProof is proof of MFA for this operation (optional if removing from trusted device)
	MFAProof *MFAProof `json:"mfa_proof,omitempty"`
}

// MsgRemoveTrustedDeviceResponse is the response for MsgRemoveTrustedDevice
type MsgRemoveTrustedDeviceResponse struct {
	// Success indicates if the device was removed
	Success bool `json:"success"`
}

// MsgUpdateSensitiveTxConfig is the message for updating sensitive tx config (governance)
type MsgUpdateSensitiveTxConfig struct {
	// Authority is the governance authority
	Authority string `json:"authority"`

	// Config is the new configuration
	Config SensitiveTxConfig `json:"config"`
}

// MsgUpdateSensitiveTxConfigResponse is the response for MsgUpdateSensitiveTxConfig
type MsgUpdateSensitiveTxConfigResponse struct {
	// Success indicates if the update was successful
	Success bool `json:"success"`
}

// MFAProof represents proof of MFA verification
type MFAProof struct {
	// SessionID is the authorization session ID
	SessionID string `json:"session_id"`

	// VerifiedFactors are the factors that were verified
	VerifiedFactors []FactorType `json:"verified_factors"`

	// Timestamp is when the proof was generated
	Timestamp int64 `json:"timestamp"`

	// Signature is the signature over the proof data
	Signature []byte `json:"signature,omitempty"`

	// DeviceFingerprint identifies the device
	DeviceFingerprint string `json:"device_fingerprint,omitempty"`

	// TrustToken is the trust token for the device (if trusted device)
	TrustToken string `json:"trust_token,omitempty"`
}

// Validate validates the MFA proof
func (p *MFAProof) Validate() error {
	if p.SessionID == "" {
		return ErrInvalidProof.Wrap("session_id cannot be empty")
	}

	if len(p.VerifiedFactors) == 0 {
		return ErrInvalidProof.Wrap("verified_factors cannot be empty")
	}

	if p.Timestamp == 0 {
		return ErrInvalidProof.Wrap("timestamp cannot be zero")
	}

	return nil
}

// Message type URLs
const (
	TypeMsgEnrollFactor            = "mfa/MsgEnrollFactor"
	TypeMsgRevokeFactor            = "mfa/MsgRevokeFactor"
	TypeMsgSetMFAPolicy            = "mfa/MsgSetMFAPolicy"
	TypeMsgCreateChallenge         = "mfa/MsgCreateChallenge"
	TypeMsgVerifyChallenge         = "mfa/MsgVerifyChallenge"
	TypeMsgAddTrustedDevice        = "mfa/MsgAddTrustedDevice"
	TypeMsgRemoveTrustedDevice     = "mfa/MsgRemoveTrustedDevice"
	TypeMsgUpdateSensitiveTxConfig = "mfa/MsgUpdateSensitiveTxConfig"
)

// Implement sdk.Msg interface for all messages

func (m *MsgEnrollFactor) Route() string { return RouterKey }
func (m *MsgEnrollFactor) Type() string  { return TypeMsgEnrollFactor }
func (m *MsgEnrollFactor) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{addr}
}
func (m *MsgEnrollFactor) GetSignBytes() []byte {
	protoMsg := convertMsgEnrollFactorToProto(m)
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(protoMsg))
}
func (m *MsgEnrollFactor) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return ErrInvalidAddress.Wrapf("invalid sender address: %v", err)
	}
	if !m.FactorType.IsValid() {
		return ErrInvalidFactorType.Wrapf("invalid factor type: %d", m.FactorType)
	}
	return nil
}

func (m *MsgRevokeFactor) Route() string { return RouterKey }
func (m *MsgRevokeFactor) Type() string  { return TypeMsgRevokeFactor }
func (m *MsgRevokeFactor) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{addr}
}
func (m *MsgRevokeFactor) GetSignBytes() []byte {
	protoMsg := convertMsgRevokeFactorToProto(m)
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(protoMsg))
}
func (m *MsgRevokeFactor) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return ErrInvalidAddress.Wrapf("invalid sender address: %v", err)
	}
	if !m.FactorType.IsValid() {
		return ErrInvalidFactorType.Wrapf("invalid factor type: %d", m.FactorType)
	}
	if m.FactorID == "" {
		return ErrInvalidEnrollment.Wrap("factor_id cannot be empty")
	}
	return nil
}

func (m *MsgSetMFAPolicy) Route() string { return RouterKey }
func (m *MsgSetMFAPolicy) Type() string  { return TypeMsgSetMFAPolicy }
func (m *MsgSetMFAPolicy) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{addr}
}
func (m *MsgSetMFAPolicy) GetSignBytes() []byte {
	protoMsg := convertMsgSetMFAPolicyToProto(m)
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(protoMsg))
}
func (m *MsgSetMFAPolicy) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return ErrInvalidAddress.Wrapf("invalid sender address: %v", err)
	}
	return m.Policy.Validate()
}

func (m *MsgCreateChallenge) Route() string { return RouterKey }
func (m *MsgCreateChallenge) Type() string  { return TypeMsgCreateChallenge }
func (m *MsgCreateChallenge) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{addr}
}
func (m *MsgCreateChallenge) GetSignBytes() []byte {
	protoMsg := convertMsgCreateChallengeToProto(m)
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(protoMsg))
}
func (m *MsgCreateChallenge) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return ErrInvalidAddress.Wrapf("invalid sender address: %v", err)
	}
	if !m.FactorType.IsValid() {
		return ErrInvalidFactorType.Wrapf("invalid factor type: %d", m.FactorType)
	}
	return nil
}

func (m *MsgVerifyChallenge) Route() string { return RouterKey }
func (m *MsgVerifyChallenge) Type() string  { return TypeMsgVerifyChallenge }
func (m *MsgVerifyChallenge) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{addr}
}
func (m *MsgVerifyChallenge) GetSignBytes() []byte {
	protoMsg := convertMsgVerifyChallengeToProto(m)
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(protoMsg))
}
func (m *MsgVerifyChallenge) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return ErrInvalidAddress.Wrapf("invalid sender address: %v", err)
	}
	if m.ChallengeID == "" {
		return ErrInvalidChallenge.Wrap("challenge_id cannot be empty")
	}
	if m.Response == nil {
		return ErrInvalidChallengeResponse.Wrap("response cannot be nil")
	}
	return m.Response.Validate()
}

func (m *MsgAddTrustedDevice) Route() string { return RouterKey }
func (m *MsgAddTrustedDevice) Type() string  { return TypeMsgAddTrustedDevice }
func (m *MsgAddTrustedDevice) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{addr}
}
func (m *MsgAddTrustedDevice) GetSignBytes() []byte {
	protoMsg := convertMsgAddTrustedDeviceToProto(m)
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(protoMsg))
}
func (m *MsgAddTrustedDevice) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return ErrInvalidAddress.Wrapf("invalid sender address: %v", err)
	}
	if m.DeviceInfo.Fingerprint == "" {
		return ErrInvalidEnrollment.Wrap("device fingerprint cannot be empty")
	}
	if m.MFAProof == nil {
		return ErrMFARequired.Wrap("MFA proof required to add trusted device")
	}
	return m.MFAProof.Validate()
}

func (m *MsgRemoveTrustedDevice) Route() string { return RouterKey }
func (m *MsgRemoveTrustedDevice) Type() string  { return TypeMsgRemoveTrustedDevice }
func (m *MsgRemoveTrustedDevice) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{addr}
}
func (m *MsgRemoveTrustedDevice) GetSignBytes() []byte {
	protoMsg := convertMsgRemoveTrustedDeviceToProto(m)
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(protoMsg))
}
func (m *MsgRemoveTrustedDevice) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return ErrInvalidAddress.Wrapf("invalid sender address: %v", err)
	}
	if m.DeviceFingerprint == "" {
		return ErrInvalidEnrollment.Wrap("device fingerprint cannot be empty")
	}
	return nil
}

func (m *MsgUpdateSensitiveTxConfig) Route() string { return RouterKey }
func (m *MsgUpdateSensitiveTxConfig) Type() string  { return TypeMsgUpdateSensitiveTxConfig }
func (m *MsgUpdateSensitiveTxConfig) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}
func (m *MsgUpdateSensitiveTxConfig) GetSignBytes() []byte {
	protoMsg := convertMsgUpdateSensitiveTxConfigToProto(m)
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(protoMsg))
}
func (m *MsgUpdateSensitiveTxConfig) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return ErrInvalidAddress.Wrapf("invalid authority address: %v", err)
	}
	return m.Config.Validate()
}
