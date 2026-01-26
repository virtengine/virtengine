package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Message types for enclave module
const (
	TypeMsgRegisterEnclaveIdentity = "register_enclaVIRTENGINE_identity"
	TypeMsgRotateEnclaveIdentity   = "rotate_enclaVIRTENGINE_identity"
	TypeMsgUpdateParams            = "update_params"
)

// MsgRegisterEnclaveIdentity registers a new enclave identity for a validator
type MsgRegisterEnclaveIdentity struct {
	// ValidatorAddress is the validator operator address (sender must be the validator operator)
	ValidatorAddress string `json:"validator_address"`

	// TEEType is the type of TEE
	TEEType TEEType `json:"tee_type"`

	// MeasurementHash is the enclave measurement hash
	MeasurementHash []byte `json:"measurement_hash"`

	// SignerHash is the signer measurement (MRSIGNER)
	SignerHash []byte `json:"signer_hash,omitempty"`

	// EncryptionPubKey is the enclave's public key for encryption
	EncryptionPubKey []byte `json:"encryption_pub_key"`

	// SigningPubKey is the enclave's public key for signing
	SigningPubKey []byte `json:"signing_pub_key"`

	// AttestationQuote is the raw attestation quote
	AttestationQuote []byte `json:"attestation_quote"`

	// AttestationChain is the certificate chain
	AttestationChain [][]byte `json:"attestation_chain"`

	// ISVProdID is the product ID
	ISVProdID uint16 `json:"isv_prod_id"`

	// ISVSVN is the security version number
	ISVSVN uint16 `json:"isv_svn"`

	// QuoteVersion is the quote format version
	QuoteVersion uint32 `json:"quote_version"`
}

// Route returns the message route
func (m MsgRegisterEnclaveIdentity) Route() string { return RouterKey }

// Type returns the message type
func (m MsgRegisterEnclaveIdentity) Type() string { return TypeMsgRegisterEnclaveIdentity }

// GetSigners returns the signers
func (m MsgRegisterEnclaveIdentity) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.ValidatorAddress)
	if err != nil {
		return nil
	}
	return []sdk.AccAddress{addr}
}

// ValidateBasic performs basic validation
func (m MsgRegisterEnclaveIdentity) ValidateBasic() error {
	if m.ValidatorAddress == "" {
		return ErrInvalidEnclaveIdentity.Wrap("validator address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(m.ValidatorAddress); err != nil {
		return ErrInvalidEnclaveIdentity.Wrapf("invalid validator address: %v", err)
	}

	if !IsValidTEEType(m.TEEType) {
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

// MsgRotateEnclaveIdentity initiates a key rotation for a validator's enclave
type MsgRotateEnclaveIdentity struct {
	// ValidatorAddress is the validator operator address
	ValidatorAddress string `json:"validator_address"`

	// NewEncryptionPubKey is the new enclave encryption public key
	NewEncryptionPubKey []byte `json:"new_encryption_pub_key"`

	// NewSigningPubKey is the new enclave signing public key
	NewSigningPubKey []byte `json:"new_signing_pub_key"`

	// NewAttestationQuote is the new attestation quote
	NewAttestationQuote []byte `json:"new_attestation_quote"`

	// NewAttestationChain is the new certificate chain
	NewAttestationChain [][]byte `json:"new_attestation_chain"`

	// NewMeasurementHash is the new measurement hash (if enclave was upgraded)
	NewMeasurementHash []byte `json:"new_measurement_hash,omitempty"`

	// NewISVSVN is the new security version
	NewISVSVN uint16 `json:"new_isv_svn"`

	// OverlapBlocks is the number of blocks for which both keys are valid
	OverlapBlocks int64 `json:"overlap_blocks"`
}

// Route returns the message route
func (m MsgRotateEnclaveIdentity) Route() string { return RouterKey }

// Type returns the message type
func (m MsgRotateEnclaveIdentity) Type() string { return TypeMsgRotateEnclaveIdentity }

// GetSigners returns the signers
func (m MsgRotateEnclaveIdentity) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.ValidatorAddress)
	if err != nil {
		return nil
	}
	return []sdk.AccAddress{addr}
}

// ValidateBasic performs basic validation
func (m MsgRotateEnclaveIdentity) ValidateBasic() error {
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
func (m MsgProposeMeasurement) Type() string { return "propose_measurement" }

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

// MsgRevokeMeasurement revokes an enclave measurement from the allowlist
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
func (m MsgRevokeMeasurement) Type() string { return "revoke_measurement" }

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

	if len(m.MeasurementHash) != 32 {
		return ErrInvalidMeasurement.Wrapf("measurement hash must be 32 bytes, got %d", len(m.MeasurementHash))
	}

	if m.Reason == "" {
		return ErrInvalidMeasurement.Wrap("reason cannot be empty")
	}

	return nil
}
