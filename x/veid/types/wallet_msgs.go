// Package types provides VEID module types.
//
// This file defines type aliases for wallet-related VEID messages, using the
// proto-generated types from sdk/go/node/veid/v1 as the source of truth.
package types

import (
	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
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

// ============================================================================
// Wallet Message Type Aliases - from proto-generated types
// ============================================================================

// MsgCreateIdentityWallet is the message for creating an identity wallet.
type MsgCreateIdentityWallet = veidv1.MsgCreateIdentityWallet

// MsgCreateIdentityWalletResponse is the response for MsgCreateIdentityWallet.
type MsgCreateIdentityWalletResponse = veidv1.MsgCreateIdentityWalletResponse

// MsgAddScopeToWallet is the message for adding a scope reference to a wallet.
type MsgAddScopeToWallet = veidv1.MsgAddScopeToWallet

// MsgAddScopeToWalletResponse is the response for MsgAddScopeToWallet.
type MsgAddScopeToWalletResponse = veidv1.MsgAddScopeToWalletResponse

// MsgRevokeScopeFromWallet is the message for revoking a scope from a wallet.
type MsgRevokeScopeFromWallet = veidv1.MsgRevokeScopeFromWallet

// MsgRevokeScopeFromWalletResponse is the response for MsgRevokeScopeFromWallet.
type MsgRevokeScopeFromWalletResponse = veidv1.MsgRevokeScopeFromWalletResponse

// MsgUpdateConsentSettings is the message for updating consent settings.
type MsgUpdateConsentSettings = veidv1.MsgUpdateConsentSettings

// MsgUpdateConsentSettingsResponse is the response for MsgUpdateConsentSettings.
type MsgUpdateConsentSettingsResponse = veidv1.MsgUpdateConsentSettingsResponse

// MsgRebindWallet is the message for rebinding a wallet during key rotation.
type MsgRebindWallet = veidv1.MsgRebindWallet

// MsgRebindWalletResponse is the response for MsgRebindWallet.
type MsgRebindWalletResponse = veidv1.MsgRebindWalletResponse

// MsgUpdateDerivedFeatures is the message for updating derived features.
type MsgUpdateDerivedFeatures = veidv1.MsgUpdateDerivedFeatures

// MsgUpdateDerivedFeaturesResponse is the response for MsgUpdateDerivedFeatures.
type MsgUpdateDerivedFeaturesResponse = veidv1.MsgUpdateDerivedFeaturesResponse

// ConsentSettings represents consent settings for a wallet.
type ConsentSettings = veidv1.ConsentSettings

// GlobalConsentUpdate represents global consent update settings.
type GlobalConsentUpdate = veidv1.GlobalConsentUpdate

// ============================================================================
// Wallet Constructor Functions
// ============================================================================

// NewMsgCreateIdentityWallet creates a new MsgCreateIdentityWallet.
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

// NewMsgAddScopeToWallet creates a new MsgAddScopeToWallet.
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
		ScopeId:       scopeID,
		ScopeType:     ScopeTypeToProto(scopeType),
		EnvelopeHash:  envelopeHash,
		UserSignature: userSignature,
		GrantConsent:  grantConsent,
	}
}

// NewMsgRevokeScopeFromWallet creates a new MsgRevokeScopeFromWallet.
func NewMsgRevokeScopeFromWallet(
	sender string,
	scopeID string,
	reason string,
	userSignature []byte,
) *MsgRevokeScopeFromWallet {
	return &MsgRevokeScopeFromWallet{
		Sender:        sender,
		ScopeId:       scopeID,
		Reason:        reason,
		UserSignature: userSignature,
	}
}

// NewMsgUpdateConsentSettings creates a new MsgUpdateConsentSettings.
func NewMsgUpdateConsentSettings(
	sender string,
	scopeID string,
	grantConsent bool,
	purpose string,
	userSignature []byte,
) *MsgUpdateConsentSettings {
	return &MsgUpdateConsentSettings{
		Sender:        sender,
		ScopeId:       scopeID,
		GrantConsent:  grantConsent,
		Purpose:       purpose,
		UserSignature: userSignature,
	}
}

// NewMsgRebindWallet creates a new MsgRebindWallet.
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

// NewMsgUpdateDerivedFeatures creates a new MsgUpdateDerivedFeatures.
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

// ============================================================================
// Helper Functions
// ============================================================================

// GetAddScopeSigningMessage returns the canonical signing message for adding a scope.
func GetAddScopeSigningMessage(sender, scopeID string) []byte {
	return []byte("VEID_ADD_SCOPE:" + sender + ":" + scopeID)
}

// GetRevokeScopeSigningMessage returns the canonical signing message for revoking a scope.
func GetRevokeScopeSigningMessage(sender, scopeID string) []byte {
	return []byte("VEID_REVOKE_SCOPE:" + sender + ":" + scopeID)
}
