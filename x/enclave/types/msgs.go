package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	v1 "github.com/virtengine/virtengine/sdk/go/node/enclave/v1"
)

// Message types for enclave module
const (
	TypeMsgRegisterEnclaveIdentity = "register_enclave_identity"
	TypeMsgRotateEnclaveIdentity   = "rotate_enclave_identity"
	TypeMsgProposeMeasurement      = "propose_measurement"
	TypeMsgRevokeMeasurement       = "revoke_measurement"
	TypeMsgUpdateParams            = "update_params"
	TypeMsgEnclaveHeartbeat        = "enclave_heartbeat"
)

// ValidateBasicMsgRegisterEnclaveIdentity performs basic validation
func ValidateBasicMsgRegisterEnclaveIdentity(m *v1.MsgRegisterEnclaveIdentity) error {
	if m.ValidatorAddress == "" {
		return ErrInvalidEnclaveIdentity.Wrap("validator address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(m.ValidatorAddress); err != nil {
		return ErrInvalidEnclaveIdentity.Wrapf("invalid validator address: %v", err)
	}

	if !IsValidTEEType(m.TeeType) {
		return ErrInvalidEnclaveIdentity.Wrapf("invalid TEE type: %s", m.TeeType)
	}

	if len(m.MeasurementHash) != 32 {
		return ErrInvalidEnclaveIdentity.Wrapf("measurement hash must be 32 bytes, got %d", len(m.MeasurementHash))
	}

	if len(m.EncryptionPubKey) == 0 {
		return ErrInvalidEnclaveIdentity.Wrap("encryption public key cannot be empty")
	}

	if len(m.SigningPubKey) == 0 {
		return ErrInvalidEnclaveIdentity.Wrap("signing public key cannot be empty")
	}

	if len(m.AttestationQuote) == 0 {
		return ErrInvalidEnclaveIdentity.Wrap("attestation quote cannot be empty")
	}

	return nil
}

// ValidateBasicMsgRotateEnclaveIdentity performs basic validation
func ValidateBasicMsgRotateEnclaveIdentity(m *v1.MsgRotateEnclaveIdentity) error {
	if m.ValidatorAddress == "" {
		return ErrInvalidEnclaveIdentity.Wrap("validator address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(m.ValidatorAddress); err != nil {
		return ErrInvalidEnclaveIdentity.Wrapf("invalid validator address: %v", err)
	}

	if len(m.NewEncryptionPubKey) == 0 {
		return ErrInvalidEnclaveIdentity.Wrap("new encryption public key cannot be empty")
	}

	if len(m.NewSigningPubKey) == 0 {
		return ErrInvalidEnclaveIdentity.Wrap("new signing public key cannot be empty")
	}

	if len(m.NewAttestationQuote) == 0 {
		return ErrInvalidEnclaveIdentity.Wrap("new attestation quote cannot be empty")
	}

	if m.OverlapBlocks <= 0 {
		return ErrInvalidEnclaveIdentity.Wrap("overlap blocks must be positive")
	}

	return nil
}

// MsgProposeMeasurement proposes a new enclave measurement for the allowlist
//
// Deprecated: this legacy message is kept for amino compatibility. Use MsgProposeMeasurement in v1.
type MsgProposeMeasurement struct {
	// Authority is the governance authority address
	Authority string `json:"authority"`

	// MeasurementHash is the enclave measurement hash to add
	MeasurementHash []byte `json:"measurement_hash"`

	// TEEType is the TEE type this measurement is for
	TEEType TEEType `json:"tee_type"`

	// Description is a human-readable description
	Description string `json:"description"`

	// MinISVSVN is the minimum required security version
	MinISVSVN uint16 `json:"min_isv_svn"`

	// ExpiryBlocks is the number of blocks until expiry (0 for no expiry)
	ExpiryBlocks int64 `json:"expiry_blocks,omitempty"`
}

// Route returns the message route
func (m MsgProposeMeasurement) Route() string { return RouterKey }

// Type returns the message type
func (m MsgProposeMeasurement) Type() string { return TypeMsgProposeMeasurement }

// GetSigners returns the signers
func (m MsgProposeMeasurement) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return nil
	}
	return []sdk.AccAddress{addr}
}

// ValidateBasic performs basic validation
func (m MsgProposeMeasurement) ValidateBasic() error {
	if m.Authority == "" {
		return ErrInvalidMeasurement.Wrap("authority cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return ErrInvalidMeasurement.Wrapf("invalid authority address: %v", err)
	}

	if len(m.MeasurementHash) != 32 {
		return ErrInvalidMeasurement.Wrapf("measurement hash must be 32 bytes, got %d", len(m.MeasurementHash))
	}

	if !IsValidTEEType(m.TEEType) {
		return ErrInvalidMeasurement.Wrapf("invalid TEE type: %s", m.TEEType)
	}

	if m.Description == "" {
		return ErrInvalidMeasurement.Wrap("description cannot be empty")
	}

	return nil
}

// ValidateBasicMsgProposeMeasurement performs basic validation
func ValidateBasicMsgProposeMeasurement(m *v1.MsgProposeMeasurement) error {
	if m.Authority == "" {
		return ErrInvalidMeasurement.Wrap("authority cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return ErrInvalidMeasurement.Wrapf("invalid authority address: %v", err)
	}

	if len(m.MeasurementHash) != 32 {
		return ErrInvalidMeasurement.Wrapf("measurement hash must be 32 bytes, got %d", len(m.MeasurementHash))
	}

	if !IsValidTEEType(m.TeeType) {
		return ErrInvalidMeasurement.Wrapf("invalid TEE type: %s", m.TeeType)
	}

	if m.Description == "" {
		return ErrInvalidMeasurement.Wrap("description cannot be empty")
	}

	return nil
}

// MsgRevokeMeasurement revokes an enclave measurement from the allowlist
//
// Deprecated: this legacy message is kept for amino compatibility. Use MsgRevokeMeasurement in v1.
type MsgRevokeMeasurement struct {
	// Authority is the governance authority address
	Authority string `json:"authority"`

	// MeasurementHash is the enclave measurement hash to revoke
	MeasurementHash []byte `json:"measurement_hash"`

	// Reason is the reason for revocation
	Reason string `json:"reason"`
}

// Route returns the message route
func (m MsgRevokeMeasurement) Route() string { return RouterKey }

// Type returns the message type
func (m MsgRevokeMeasurement) Type() string { return TypeMsgRevokeMeasurement }

// GetSigners returns the signers
func (m MsgRevokeMeasurement) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return nil
	}
	return []sdk.AccAddress{addr}
}

// ValidateBasic performs basic validation
func (m MsgRevokeMeasurement) ValidateBasic() error {
	if m.Authority == "" {
		return ErrInvalidMeasurement.Wrap("authority cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return ErrInvalidMeasurement.Wrapf("invalid authority address: %v", err)
	}

	if len(m.MeasurementHash) != 32 {
		return ErrInvalidMeasurement.Wrapf("measurement hash must be 32 bytes, got %d", len(m.MeasurementHash))
	}

	if m.Reason == "" {
		return ErrInvalidMeasurement.Wrap("reason cannot be empty")
	}

	return nil
}

// ValidateBasicMsgRegisterEnclaveIdentity performs basic validation for SDK CLI types.
func ValidateBasicMsgRegisterEnclaveIdentity(m *v1.MsgRegisterEnclaveIdentity) error {
	if m.ValidatorAddress == "" {
		return ErrInvalidEnclaveIdentity.Wrap("validator address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(m.ValidatorAddress); err != nil {
		return ErrInvalidEnclaveIdentity.Wrapf("invalid validator address: %v", err)
	}

	if !IsValidTEEType(TEEType(m.TEEType)) {
		return ErrInvalidEnclaveIdentity.Wrapf("invalid TEE type: %s", m.TEEType)
	}

	if len(m.MeasurementHash) != 32 {
		return ErrInvalidEnclaveIdentity.Wrapf("measurement hash must be 32 bytes, got %d", len(m.MeasurementHash))
	}

	if len(m.EncryptionPubKey) == 0 {
		return ErrInvalidEnclaveIdentity.Wrap("encryption public key cannot be empty")
	}

	if len(m.SigningPubKey) == 0 {
		return ErrInvalidEnclaveIdentity.Wrap("signing public key cannot be empty")
	}

	if len(m.AttestationQuote) == 0 {
		return ErrInvalidEnclaveIdentity.Wrap("attestation quote cannot be empty")
	}

	return nil
}

// ValidateBasicMsgRotateEnclaveIdentity performs basic validation for SDK CLI types.
func ValidateBasicMsgRotateEnclaveIdentity(m *v1.MsgRotateEnclaveIdentity) error {
	if m.ValidatorAddress == "" {
		return ErrInvalidEnclaveIdentity.Wrap("validator address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(m.ValidatorAddress); err != nil {
		return ErrInvalidEnclaveIdentity.Wrapf("invalid validator address: %v", err)
	}

	if len(m.NewEncryptionPubKey) == 0 {
		return ErrInvalidEnclaveIdentity.Wrap("new encryption public key cannot be empty")
	}

	if len(m.NewSigningPubKey) == 0 {
		return ErrInvalidEnclaveIdentity.Wrap("new signing public key cannot be empty")
	}

	if len(m.NewAttestationQuote) == 0 {
		return ErrInvalidEnclaveIdentity.Wrap("new attestation quote cannot be empty")
	}

	if m.OverlapBlocks <= 0 {
		return ErrInvalidEnclaveIdentity.Wrap("overlap blocks must be positive")
	}

	return nil
}

// ValidateBasicMsgProposeMeasurement performs basic validation for SDK CLI types.
func ValidateBasicMsgProposeMeasurement(m *v1.MsgProposeMeasurement) error {
	if m.Authority == "" {
		return ErrInvalidMeasurement.Wrap("authority cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return ErrInvalidMeasurement.Wrapf("invalid authority address: %v", err)
	}

	if len(m.MeasurementHash) != 32 {
		return ErrInvalidMeasurement.Wrapf("measurement hash must be 32 bytes, got %d", len(m.MeasurementHash))
	}

	if !IsValidTEEType(TEEType(m.TEEType)) {
		return ErrInvalidMeasurement.Wrapf("invalid TEE type: %s", m.TEEType)
	}

	if m.Description == "" {
		return ErrInvalidMeasurement.Wrap("description cannot be empty")
	}

	return nil
}

// ValidateBasicMsgRevokeMeasurement performs basic validation for SDK CLI types.
func ValidateBasicMsgRevokeMeasurement(m *v1.MsgRevokeMeasurement) error {
	if m.Authority == "" {
		return ErrInvalidMeasurement.Wrap("authority cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return ErrInvalidMeasurement.Wrapf("invalid authority address: %v", err)
	}

	if len(m.MeasurementHash) != 32 {
		return ErrInvalidMeasurement.Wrapf("measurement hash must be 32 bytes, got %d", len(m.MeasurementHash))
	}

	if m.Reason == "" {
		return ErrInvalidMeasurement.Wrap("reason cannot be empty")
	}

	return nil
}
