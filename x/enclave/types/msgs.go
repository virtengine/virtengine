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

// ValidateBasicMsgRevokeMeasurement performs basic validation
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
