package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Wallet message type constants
const (
	TypeMsgCreateIdentityWallet  = "create_identity_wallet"
	TypeMsgAddScopeToWallet      = "add_scope_to_wallet"
	TypeMsgRevokeScopeFromWallet = "revoke_scope_from_wallet"
	TypeMsgUpdateConsentSettings = "update_consent_settings"
	TypeMsgRebindWallet          = "rebind_wallet"
	TypeMsgUpdateDerivedFeatures = "update_derived_features"
)

var (
	_ sdk.Msg = &MsgCreateIdentityWallet{}
	_ sdk.Msg = &MsgAddScopeToWallet{}
	_ sdk.Msg = &MsgRevokeScopeFromWallet{}
	_ sdk.Msg = &MsgUpdateConsentSettings{}
	_ sdk.Msg = &MsgRebindWallet{}
	_ sdk.Msg = &MsgUpdateDerivedFeatures{}
)

// ============================================================================
// MsgCreateIdentityWallet
// ============================================================================

// MsgCreateIdentityWallet is the message for creating an identity wallet
type MsgCreateIdentityWallet struct {
	// Sender is the account creating the wallet
	Sender string `json:"sender"`

	// BindingSignature is the signature proving ownership of the account
	// Signs: SHA256("VEID_WALLET_BINDING:" + wallet_id + ":" + sender)
	BindingSignature []byte `json:"binding_signature"`

	// BindingPubKey is the public key used for the binding signature
	BindingPubKey []byte `json:"binding_pub_key"`

	// InitialConsent contains optional initial consent settings
	InitialConsent *ConsentSettings `json:"initial_consent,omitempty"`

	// Metadata contains optional wallet metadata
	Metadata map[string]string `json:"metadata,omitempty"`
}

// NewMsgCreateIdentityWallet creates a new MsgCreateIdentityWallet
func NewMsgCreateIdentityWallet(
	sender string,
	bindingSignature []byte,
	bindingPubKey []byte,
) *MsgCreateIdentityWallet {
	return &MsgCreateIdentityWallet{
		Sender:           sender,
		BindingSignature: bindingSignature,
		BindingPubKey:    bindingPubKey,
		Metadata:         make(map[string]string),
	}
}

// Route returns the route for the message
func (msg MsgCreateIdentityWallet) Route() string { return RouterKey }

// Type returns the type for the message
func (msg MsgCreateIdentityWallet) Type() string { return TypeMsgCreateIdentityWallet }

// ValidateBasic validates the message
func (msg MsgCreateIdentityWallet) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if len(msg.BindingSignature) == 0 {
		return ErrInvalidUserSignature.Wrap("binding_signature cannot be empty")
	}

	if len(msg.BindingPubKey) == 0 {
		return ErrInvalidWallet.Wrap("binding_pub_key cannot be empty")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg MsgCreateIdentityWallet) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{signer}
}

// GetSignBytes returns the sign bytes for the message
func (msg MsgCreateIdentityWallet) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// MsgCreateIdentityWalletResponse is the response for MsgCreateIdentityWallet
type MsgCreateIdentityWalletResponse struct {
	WalletID  string `json:"wallet_id"`
	CreatedAt int64  `json:"created_at"`
}

// ============================================================================
// MsgAddScopeToWallet
// ============================================================================

// MsgAddScopeToWallet is the message for adding a scope reference to a wallet
type MsgAddScopeToWallet struct {
	// Sender is the account that owns the wallet
	Sender string `json:"sender"`

	// ScopeID is the unique identifier of the scope to add
	ScopeID string `json:"scope_id"`

	// ScopeType is the type of scope being added
	ScopeType ScopeType `json:"scope_type"`

	// EnvelopeHash is the hash of the encrypted scope envelope
	EnvelopeHash []byte `json:"envelope_hash"`

	// UserSignature authorizes adding this scope to the wallet
	// Signs: SHA256("VEID_ADD_SCOPE:" + sender + ":" + scope_id)
	UserSignature []byte `json:"user_signature"`

	// GrantConsent indicates if consent should be granted for this scope
	GrantConsent bool `json:"grant_consent"`

	// ConsentPurpose describes the purpose for consent (if granting)
	ConsentPurpose string `json:"consent_purpose,omitempty"`

	// ConsentExpiresAt is when consent expires (optional)
	ConsentExpiresAt *int64 `json:"consent_expires_at,omitempty"`
}

// NewMsgAddScopeToWallet creates a new MsgAddScopeToWallet
func NewMsgAddScopeToWallet(
	sender string,
	scopeID string,
	scopeType ScopeType,
	envelopeHash []byte,
	userSignature []byte,
	grantConsent bool,
) *MsgAddScopeToWallet {
	return &MsgAddScopeToWallet{
		Sender:        sender,
		ScopeID:       scopeID,
		ScopeType:     scopeType,
		EnvelopeHash:  envelopeHash,
		UserSignature: userSignature,
		GrantConsent:  grantConsent,
	}
}

// Route returns the route for the message
func (msg MsgAddScopeToWallet) Route() string { return RouterKey }

// Type returns the type for the message
func (msg MsgAddScopeToWallet) Type() string { return TypeMsgAddScopeToWallet }

// ValidateBasic validates the message
func (msg MsgAddScopeToWallet) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.ScopeID == "" {
		return ErrInvalidScope.Wrap("scope_id cannot be empty")
	}

	if !IsValidScopeType(msg.ScopeType) {
		return ErrInvalidScopeType.Wrapf("invalid scope type: %s", msg.ScopeType)
	}

	if len(msg.EnvelopeHash) != 32 {
		return ErrInvalidPayloadHash.Wrap("envelope_hash must be 32 bytes (SHA256)")
	}

	if len(msg.UserSignature) == 0 {
		return ErrInvalidUserSignature.Wrap("user_signature cannot be empty")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg MsgAddScopeToWallet) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{signer}
}

// GetSignBytes returns the sign bytes for the message
func (msg MsgAddScopeToWallet) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigningMessage returns the message that should be signed
func (msg MsgAddScopeToWallet) GetSigningMessage() []byte {
	return GetAddScopeSigningMessage(msg.Sender, msg.ScopeID)
}

// GetAddScopeSigningMessage returns the canonical signing message for adding a scope
func GetAddScopeSigningMessage(sender, scopeID string) []byte {
	return []byte("VEID_ADD_SCOPE:" + sender + ":" + scopeID)
}

// MsgAddScopeToWalletResponse is the response for MsgAddScopeToWallet
type MsgAddScopeToWalletResponse struct {
	ScopeID string `json:"scope_id"`
	AddedAt int64  `json:"added_at"`
}

// ============================================================================
// MsgRevokeScopeFromWallet
// ============================================================================

// MsgRevokeScopeFromWallet is the message for revoking a scope from a wallet
type MsgRevokeScopeFromWallet struct {
	// Sender is the account that owns the wallet
	Sender string `json:"sender"`

	// ScopeID is the unique identifier of the scope to revoke
	ScopeID string `json:"scope_id"`

	// Reason is the reason for revocation
	Reason string `json:"reason,omitempty"`

	// UserSignature authorizes revoking this scope from the wallet
	// Signs: SHA256("VEID_REVOKE_SCOPE:" + sender + ":" + scope_id)
	UserSignature []byte `json:"user_signature"`
}

// NewMsgRevokeScopeFromWallet creates a new MsgRevokeScopeFromWallet
func NewMsgRevokeScopeFromWallet(
	sender string,
	scopeID string,
	reason string,
	userSignature []byte,
) *MsgRevokeScopeFromWallet {
	return &MsgRevokeScopeFromWallet{
		Sender:        sender,
		ScopeID:       scopeID,
		Reason:        reason,
		UserSignature: userSignature,
	}
}

// Route returns the route for the message
func (msg MsgRevokeScopeFromWallet) Route() string { return RouterKey }

// Type returns the type for the message
func (msg MsgRevokeScopeFromWallet) Type() string { return TypeMsgRevokeScopeFromWallet }

// ValidateBasic validates the message
func (msg MsgRevokeScopeFromWallet) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.ScopeID == "" {
		return ErrInvalidScope.Wrap("scope_id cannot be empty")
	}

	if len(msg.UserSignature) == 0 {
		return ErrInvalidUserSignature.Wrap("user_signature cannot be empty")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg MsgRevokeScopeFromWallet) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{signer}
}

// GetSignBytes returns the sign bytes for the message
func (msg MsgRevokeScopeFromWallet) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigningMessage returns the message that should be signed
func (msg MsgRevokeScopeFromWallet) GetSigningMessage() []byte {
	return GetRevokeScopeSigningMessage(msg.Sender, msg.ScopeID)
}

// GetRevokeScopeSigningMessage returns the canonical signing message for revoking a scope
func GetRevokeScopeSigningMessage(sender, scopeID string) []byte {
	return []byte("VEID_REVOKE_SCOPE:" + sender + ":" + scopeID)
}

// MsgRevokeScopeFromWalletResponse is the response for MsgRevokeScopeFromWallet
type MsgRevokeScopeFromWalletResponse struct {
	ScopeID   string `json:"scope_id"`
	RevokedAt int64  `json:"revoked_at"`
}

// ============================================================================
// MsgUpdateConsentSettings
// ============================================================================

// MsgUpdateConsentSettings is the message for updating consent settings
type MsgUpdateConsentSettings struct {
	// Sender is the account that owns the wallet
	Sender string `json:"sender"`

	// ScopeID is the scope to update consent for (empty for global settings)
	ScopeID string `json:"scope_id,omitempty"`

	// GrantConsent indicates whether to grant or revoke consent
	GrantConsent bool `json:"grant_consent"`

	// Purpose is the purpose for granting consent
	Purpose string `json:"purpose,omitempty"`

	// ExpiresAt is when the consent should expire (Unix timestamp)
	ExpiresAt *int64 `json:"expires_at,omitempty"`

	// GlobalSettings contains global settings updates
	GlobalSettings *GlobalConsentUpdate `json:"global_settings,omitempty"`

	// UserSignature authorizes this consent update
	// Signs: SHA256("VEID_CONSENT_UPDATE:" + sender + ":" + scope_id + ":" + grant_consent)
	UserSignature []byte `json:"user_signature"`
}

// NewMsgUpdateConsentSettings creates a new MsgUpdateConsentSettings
func NewMsgUpdateConsentSettings(
	sender string,
	scopeID string,
	grantConsent bool,
	purpose string,
	userSignature []byte,
) *MsgUpdateConsentSettings {
	return &MsgUpdateConsentSettings{
		Sender:        sender,
		ScopeID:       scopeID,
		GrantConsent:  grantConsent,
		Purpose:       purpose,
		UserSignature: userSignature,
	}
}

// Route returns the route for the message
func (msg MsgUpdateConsentSettings) Route() string { return RouterKey }

// Type returns the type for the message
func (msg MsgUpdateConsentSettings) Type() string { return TypeMsgUpdateConsentSettings }

// ValidateBasic validates the message
func (msg MsgUpdateConsentSettings) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if len(msg.UserSignature) == 0 {
		return ErrInvalidUserSignature.Wrap("user_signature cannot be empty")
	}

	// Must have either scope-specific or global update
	if msg.ScopeID == "" && msg.GlobalSettings == nil {
		return ErrInvalidWallet.Wrap("must specify scope_id or global_settings")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg MsgUpdateConsentSettings) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{signer}
}

// GetSignBytes returns the sign bytes for the message
func (msg MsgUpdateConsentSettings) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigningMessage returns the message that should be signed
func (msg MsgUpdateConsentSettings) GetSigningMessage() []byte {
	grantStr := "revoke"
	if msg.GrantConsent {
		grantStr = "grant"
	}
	return []byte("VEID_CONSENT_UPDATE:" + msg.Sender + ":" + msg.ScopeID + ":" + grantStr)
}

// MsgUpdateConsentSettingsResponse is the response for MsgUpdateConsentSettings
type MsgUpdateConsentSettingsResponse struct {
	UpdatedAt      int64  `json:"updated_at"`
	ConsentVersion uint32 `json:"consent_version"`
}

// ============================================================================
// MsgRebindWallet
// ============================================================================

// MsgRebindWallet is the message for rebinding a wallet during key rotation
type MsgRebindWallet struct {
	// Sender is the account that owns the wallet
	Sender string `json:"sender"`

	// NewBindingSignature is the new binding signature with the new key
	NewBindingSignature []byte `json:"new_binding_signature"`

	// NewBindingPubKey is the new public key
	NewBindingPubKey []byte `json:"new_binding_pub_key"`

	// OldSignature proves ownership of the old key (signs the new pub key)
	OldSignature []byte `json:"old_signature"`
}

// NewMsgRebindWallet creates a new MsgRebindWallet
func NewMsgRebindWallet(
	sender string,
	newBindingSignature []byte,
	newBindingPubKey []byte,
	oldSignature []byte,
) *MsgRebindWallet {
	return &MsgRebindWallet{
		Sender:              sender,
		NewBindingSignature: newBindingSignature,
		NewBindingPubKey:    newBindingPubKey,
		OldSignature:        oldSignature,
	}
}

// Route returns the route for the message
func (msg MsgRebindWallet) Route() string { return RouterKey }

// Type returns the type for the message
func (msg MsgRebindWallet) Type() string { return TypeMsgRebindWallet }

// ValidateBasic validates the message
func (msg MsgRebindWallet) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if len(msg.NewBindingSignature) == 0 {
		return ErrInvalidUserSignature.Wrap("new_binding_signature cannot be empty")
	}

	if len(msg.NewBindingPubKey) == 0 {
		return ErrInvalidWallet.Wrap("new_binding_pub_key cannot be empty")
	}

	if len(msg.OldSignature) == 0 {
		return ErrInvalidUserSignature.Wrap("old_signature cannot be empty")
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg MsgRebindWallet) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{signer}
}

// GetSignBytes returns the sign bytes for the message
func (msg MsgRebindWallet) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// MsgRebindWalletResponse is the response for MsgRebindWallet
type MsgRebindWalletResponse struct {
	ReboundAt int64 `json:"rebound_at"`
}

// ============================================================================
// MsgUpdateDerivedFeatures
// ============================================================================

// MsgUpdateDerivedFeatures is the message for updating derived features
// This is a validator-only message
type MsgUpdateDerivedFeatures struct {
	// Sender is the validator submitting the update
	Sender string `json:"sender"`

	// AccountAddress is the account to update features for
	AccountAddress string `json:"account_address"`

	// FaceEmbeddingHash is the new face embedding hash (nil to keep existing)
	FaceEmbeddingHash []byte `json:"face_embedding_hash,omitempty"`

	// DocFieldHashes are new document field hashes
	DocFieldHashes map[string][]byte `json:"doc_field_hashes,omitempty"`

	// BiometricHash is the new biometric hash (nil to keep existing)
	BiometricHash []byte `json:"biometric_hash,omitempty"`

	// LivenessProofHash is the new liveness proof hash (nil to keep existing)
	LivenessProofHash []byte `json:"liveness_proof_hash,omitempty"`

	// ModelVersion is the ML model version used
	ModelVersion string `json:"model_version"`
}

// NewMsgUpdateDerivedFeatures creates a new MsgUpdateDerivedFeatures
func NewMsgUpdateDerivedFeatures(
	sender string,
	accountAddress string,
	modelVersion string,
) *MsgUpdateDerivedFeatures {
	return &MsgUpdateDerivedFeatures{
		Sender:         sender,
		AccountAddress: accountAddress,
		ModelVersion:   modelVersion,
		DocFieldHashes: make(map[string][]byte),
	}
}

// Route returns the route for the message
func (msg MsgUpdateDerivedFeatures) Route() string { return RouterKey }

// Type returns the type for the message
func (msg MsgUpdateDerivedFeatures) Type() string { return TypeMsgUpdateDerivedFeatures }

// ValidateBasic validates the message
func (msg MsgUpdateDerivedFeatures) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if _, err := sdk.AccAddressFromBech32(msg.AccountAddress); err != nil {
		return ErrInvalidAddress.Wrap("invalid account address")
	}

	if msg.ModelVersion == "" {
		return ErrInvalidWallet.Wrap("model_version cannot be empty")
	}

	// Validate hash lengths
	if len(msg.FaceEmbeddingHash) > 0 && len(msg.FaceEmbeddingHash) != 32 {
		return ErrInvalidPayloadHash.Wrap("face_embedding_hash must be 32 bytes")
	}

	if len(msg.BiometricHash) > 0 && len(msg.BiometricHash) != 32 {
		return ErrInvalidPayloadHash.Wrap("biometric_hash must be 32 bytes")
	}

	if len(msg.LivenessProofHash) > 0 && len(msg.LivenessProofHash) != 32 {
		return ErrInvalidPayloadHash.Wrap("liveness_proof_hash must be 32 bytes")
	}

	for key, hash := range msg.DocFieldHashes {
		if len(hash) != 32 {
			return ErrInvalidPayloadHash.Wrapf("doc_field_hash[%s] must be 32 bytes", key)
		}
	}

	return nil
}

// GetSigners returns the signers for the message
func (msg MsgUpdateDerivedFeatures) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{signer}
}

// GetSignBytes returns the sign bytes for the message
func (msg MsgUpdateDerivedFeatures) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ToDerivedFeaturesUpdate converts the message to a DerivedFeaturesUpdate
func (msg MsgUpdateDerivedFeatures) ToDerivedFeaturesUpdate() *DerivedFeaturesUpdate {
	return &DerivedFeaturesUpdate{
		AccountAddress:    msg.AccountAddress,
		FaceEmbeddingHash: msg.FaceEmbeddingHash,
		DocFieldHashes:    msg.DocFieldHashes,
		BiometricHash:     msg.BiometricHash,
		LivenessProofHash: msg.LivenessProofHash,
		ModelVersion:      msg.ModelVersion,
		ValidatorAddress:  msg.Sender,
	}
}

// MsgUpdateDerivedFeaturesResponse is the response for MsgUpdateDerivedFeatures
type MsgUpdateDerivedFeaturesResponse struct {
	UpdatedAt int64 `json:"updated_at"`
}
