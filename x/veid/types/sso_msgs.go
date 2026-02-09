// Package types provides message types for the VEID module.
//
// VE-4B: SSO/OIDC Linkage Message Types
// This file defines the on-chain message types for SSO linkage management.
package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Message type constants for SSO linkage
const (
	TypeMsgCreateSSOLinkage = "create_sso_linkage"
	TypeMsgRevokeSSOLinkage = "revoke_sso_linkage"
	TypeMsgUpdateSSOLinkage = "update_sso_linkage"
)

// ============================================================================
// MsgCreateSSOLinkage
// ============================================================================

// MsgCreateSSOLinkage creates an SSO linkage record on-chain.
type MsgCreateSSOLinkage struct {
	// AccountAddress is the account creating the linkage
	AccountAddress string `json:"account_address"`

	// LinkageID is the unique identifier for this linkage
	LinkageID string `json:"linkage_id"`

	// Provider is the SSO provider type
	Provider SSOProviderType `json:"provider"`

	// Issuer is the OIDC issuer URL
	Issuer string `json:"issuer"`

	// SubjectHash is the SHA256 hash of the SSO subject identifier
	SubjectHash string `json:"subject_hash"`

	// Nonce is the verification nonce used
	Nonce string `json:"nonce"`

	// AttestationData contains the signed attestation (JSON-encoded)
	AttestationData []byte `json:"attestation_data"`

	// AttestationSignature is the signer's signature on the attestation
	AttestationSignature []byte `json:"attestation_signature"`

	// SignerFingerprint is the fingerprint of the signing key
	SignerFingerprint string `json:"signer_fingerprint"`

	// AccountSignature is the account's signature authorizing the linkage
	AccountSignature []byte `json:"account_signature"`

	// ExpiresAt is when the linkage expires (optional)
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// EmailDomainHash is the hash of the email domain (optional)
	EmailDomainHash string `json:"email_domain_hash,omitempty"`

	// TenantIDHash is the hash of the tenant ID (optional)
	TenantIDHash string `json:"tenant_id_hash,omitempty"`

	// EvidenceHash is the hash of the verification evidence payload
	EvidenceHash string `json:"evidence_hash,omitempty"`

	// EvidenceStorageBackend indicates where the encrypted evidence is stored
	EvidenceStorageBackend string `json:"evidence_storage_backend,omitempty"`

	// EvidenceStorageRef is a backend-specific reference to the encrypted evidence
	EvidenceStorageRef string `json:"evidence_storage_ref,omitempty"`

	// EvidenceMetadata contains optional evidence metadata
	EvidenceMetadata map[string]string `json:"evidence_metadata,omitempty"`
}

// Route implements sdk.Msg
func (msg MsgCreateSSOLinkage) Route() string { return ModuleName }

// Type implements sdk.Msg
func (msg MsgCreateSSOLinkage) Type() string { return TypeMsgCreateSSOLinkage }

// GetSigners implements sdk.Msg
func (msg MsgCreateSSOLinkage) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.AccountAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// ValidateBasic implements sdk.Msg
func (msg MsgCreateSSOLinkage) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.AccountAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid account address: %s", err)
	}

	if msg.LinkageID == "" {
		return ErrInvalidSSO.Wrap("linkage_id cannot be empty")
	}

	if !IsValidSSOProviderType(msg.Provider) {
		return ErrInvalidSSO.Wrapf("invalid provider: %s", msg.Provider)
	}

	if msg.Issuer == "" {
		return ErrInvalidSSO.Wrap("issuer cannot be empty")
	}

	if msg.SubjectHash == "" {
		return ErrInvalidSSO.Wrap("subject_hash cannot be empty")
	}

	if len(msg.SubjectHash) != 64 {
		return ErrInvalidSSO.Wrap("subject_hash must be a valid SHA256 hex string")
	}

	if msg.Nonce == "" {
		return ErrInvalidSSO.Wrap("nonce cannot be empty")
	}

	if len(msg.AttestationData) == 0 {
		return ErrInvalidSSO.Wrap("attestation_data cannot be empty")
	}

	if len(msg.AttestationSignature) == 0 {
		return ErrInvalidSSO.Wrap("attestation_signature cannot be empty")
	}

	if msg.SignerFingerprint == "" {
		return ErrInvalidSSO.Wrap("signer_fingerprint cannot be empty")
	}

	if len(msg.AccountSignature) == 0 {
		return ErrInvalidBindingSignature.Wrap("account_signature cannot be empty")
	}

	if err := validateEvidencePointer(msg.EvidenceHash, msg.EvidenceStorageBackend, msg.EvidenceStorageRef, true); err != nil {
		return ErrInvalidSSO.Wrap(err.Error())
	}

	return nil
}

// MsgCreateSSOLinkageResponse is the response for MsgCreateSSOLinkage.
type MsgCreateSSOLinkageResponse struct {
	// LinkageID is the created linkage ID
	LinkageID string `json:"linkage_id"`

	// Status is the linkage status
	Status SSOVerificationStatus `json:"status"`

	// ScoreContribution is the score contribution from this linkage
	ScoreContribution uint32 `json:"score_contribution"`

	// VerifiedAt is when the linkage was verified
	VerifiedAt time.Time `json:"verified_at"`
}

// ============================================================================
// MsgRevokeSSOLinkage
// ============================================================================

// MsgRevokeSSOLinkage revokes an existing SSO linkage.
type MsgRevokeSSOLinkage struct {
	// AccountAddress is the account that owns the linkage
	AccountAddress string `json:"account_address"`

	// LinkageID is the linkage to revoke
	LinkageID string `json:"linkage_id"`

	// Reason is the revocation reason
	Reason string `json:"reason"`

	// Signature is the account's signature authorizing revocation
	Signature []byte `json:"signature"`
}

// Route implements sdk.Msg
func (msg MsgRevokeSSOLinkage) Route() string { return ModuleName }

// Type implements sdk.Msg
func (msg MsgRevokeSSOLinkage) Type() string { return TypeMsgRevokeSSOLinkage }

// GetSigners implements sdk.Msg
func (msg MsgRevokeSSOLinkage) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.AccountAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// ValidateBasic implements sdk.Msg
func (msg MsgRevokeSSOLinkage) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.AccountAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid account address: %s", err)
	}

	if msg.LinkageID == "" {
		return ErrInvalidSSO.Wrap("linkage_id cannot be empty")
	}

	if len(msg.Signature) == 0 {
		return ErrInvalidBindingSignature.Wrap("signature cannot be empty")
	}

	return nil
}

// MsgRevokeSSOLinkageResponse is the response for MsgRevokeSSOLinkage.
type MsgRevokeSSOLinkageResponse struct {
	// LinkageID is the revoked linkage ID
	LinkageID string `json:"linkage_id"`

	// Status is the new status
	Status SSOVerificationStatus `json:"status"`

	// RevokedAt is when the linkage was revoked
	RevokedAt time.Time `json:"revoked_at"`
}

// ============================================================================
// MsgUpdateSSOLinkage
// ============================================================================

// MsgUpdateSSOLinkage updates an existing SSO linkage (e.g., refresh).
type MsgUpdateSSOLinkage struct {
	// AccountAddress is the account that owns the linkage
	AccountAddress string `json:"account_address"`

	// LinkageID is the linkage to update
	LinkageID string `json:"linkage_id"`

	// NewExpiresAt is the new expiration time
	NewExpiresAt *time.Time `json:"new_expires_at,omitempty"`

	// NewAttestationData contains the new signed attestation
	NewAttestationData []byte `json:"new_attestation_data,omitempty"`

	// NewAttestationSignature is the new signature
	NewAttestationSignature []byte `json:"new_attestation_signature,omitempty"`

	// NewSignerFingerprint is the new signer fingerprint
	NewSignerFingerprint string `json:"new_signer_fingerprint,omitempty"`

	// AccountSignature is the account's signature authorizing the update
	AccountSignature []byte `json:"account_signature"`
}

// Route implements sdk.Msg
func (msg MsgUpdateSSOLinkage) Route() string { return ModuleName }

// Type implements sdk.Msg
func (msg MsgUpdateSSOLinkage) Type() string { return TypeMsgUpdateSSOLinkage }

// GetSigners implements sdk.Msg
func (msg MsgUpdateSSOLinkage) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.AccountAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// ValidateBasic implements sdk.Msg
func (msg MsgUpdateSSOLinkage) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.AccountAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid account address: %s", err)
	}

	if msg.LinkageID == "" {
		return ErrInvalidSSO.Wrap("linkage_id cannot be empty")
	}

	if len(msg.AccountSignature) == 0 {
		return ErrInvalidBindingSignature.Wrap("account_signature cannot be empty")
	}

	// At least one update field must be provided
	hasUpdate := msg.NewExpiresAt != nil ||
		len(msg.NewAttestationData) > 0 ||
		len(msg.NewAttestationSignature) > 0 ||
		msg.NewSignerFingerprint != ""

	if !hasUpdate {
		return ErrInvalidSSO.Wrap("at least one update field must be provided")
	}

	return nil
}

// MsgUpdateSSOLinkageResponse is the response for MsgUpdateSSOLinkage.
type MsgUpdateSSOLinkageResponse struct {
	// LinkageID is the updated linkage ID
	LinkageID string `json:"linkage_id"`

	// UpdatedAt is when the linkage was updated
	UpdatedAt time.Time `json:"updated_at"`
}

// ============================================================================
// SSO Nonce Tracking for Replay Protection
// ============================================================================

// SSONonceRecord tracks nonce usage for SSO verification replay protection.
type SSONonceRecord struct {
	// NonceHash is the SHA256 hash of the nonce
	NonceHash string `json:"nonce_hash"`

	// AccountAddress is the account that used this nonce
	AccountAddress string `json:"account_address"`

	// Provider is the SSO provider
	Provider SSOProviderType `json:"provider"`

	// Issuer is the OIDC issuer
	Issuer string `json:"issuer"`

	// LinkageID is the resulting linkage ID
	LinkageID string `json:"linkage_id"`

	// UsedAt is when the nonce was used
	UsedAt time.Time `json:"used_at"`

	// BlockHeight is the block height when used
	BlockHeight int64 `json:"block_height"`

	// ExpiresAt is when this record can be pruned
	ExpiresAt time.Time `json:"expires_at"`
}

// NewSSONonceRecord creates a new SSO nonce record.
func NewSSONonceRecord(
	nonceHash string,
	accountAddress string,
	provider SSOProviderType,
	issuer string,
	linkageID string,
	usedAt time.Time,
	blockHeight int64,
	retentionDuration time.Duration,
) *SSONonceRecord {
	return &SSONonceRecord{
		NonceHash:      nonceHash,
		AccountAddress: accountAddress,
		Provider:       provider,
		Issuer:         issuer,
		LinkageID:      linkageID,
		UsedAt:         usedAt,
		BlockHeight:    blockHeight,
		ExpiresAt:      usedAt.Add(retentionDuration),
	}
}

// Validate validates the SSO nonce record.
func (r *SSONonceRecord) Validate() error {
	if r.NonceHash == "" {
		return ErrInvalidNonce.Wrap("nonce_hash cannot be empty")
	}
	if len(r.NonceHash) != 64 {
		return ErrInvalidNonce.Wrap("nonce_hash must be a valid SHA256 hex string")
	}
	if r.AccountAddress == "" {
		return ErrInvalidAddress.Wrap("account_address cannot be empty")
	}
	if !IsValidSSOProviderType(r.Provider) {
		return ErrInvalidSSO.Wrapf("invalid provider: %s", r.Provider)
	}
	if r.Issuer == "" {
		return ErrInvalidSSO.Wrap("issuer cannot be empty")
	}
	if r.UsedAt.IsZero() {
		return ErrInvalidSSO.Wrap("used_at cannot be zero")
	}
	return nil
}

// CanBePruned checks if the record can be pruned.
func (r *SSONonceRecord) CanBePruned(now time.Time) bool {
	return now.After(r.ExpiresAt)
}
