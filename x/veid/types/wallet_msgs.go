// Package types provides wallet message types for the VEID module.
//
// This file defines type aliases to the buf-generated protobuf types in
// sdk/go/node/veid/v1 for wallet-related operations.
package types

import (
	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
)

// Wallet message type constants - kept for backwards compatibility
const (
	TypeMsgCreateIdentityWallet  = "create_identity_wallet"
	TypeMsgAddScopeToWallet      = "add_scope_to_wallet"
	TypeMsgRevokeScopeFromWallet = "revoke_scope_from_wallet"
	TypeMsgUpdateConsentSettings = "update_consent_settings"
	TypeMsgRebindWallet          = "rebind_wallet"
	TypeMsgUpdateDerivedFeatures = "update_derived_features"
)

// ============================================================================
// Wallet Message Type Aliases
// These types alias the buf-generated protobuf types which already implement
// proto.Message and sdk.Msg interfaces.
// ============================================================================

// MsgCreateIdentityWallet is the message for creating an identity wallet
type MsgCreateIdentityWallet = veidv1.MsgCreateIdentityWallet

// MsgCreateIdentityWalletResponse is the response for MsgCreateIdentityWallet
type MsgCreateIdentityWalletResponse = veidv1.MsgCreateIdentityWalletResponse

// MsgAddScopeToWallet is the message for adding a scope reference to a wallet
type MsgAddScopeToWallet = veidv1.MsgAddScopeToWallet

// MsgAddScopeToWalletResponse is the response for MsgAddScopeToWallet
type MsgAddScopeToWalletResponse = veidv1.MsgAddScopeToWalletResponse

// MsgRevokeScopeFromWallet is the message for revoking a scope from a wallet
type MsgRevokeScopeFromWallet = veidv1.MsgRevokeScopeFromWallet

// MsgRevokeScopeFromWalletResponse is the response for MsgRevokeScopeFromWallet
type MsgRevokeScopeFromWalletResponse = veidv1.MsgRevokeScopeFromWalletResponse

// MsgUpdateConsentSettings is the message for updating consent settings
type MsgUpdateConsentSettings = veidv1.MsgUpdateConsentSettings

// MsgUpdateConsentSettingsResponse is the response for MsgUpdateConsentSettings
type MsgUpdateConsentSettingsResponse = veidv1.MsgUpdateConsentSettingsResponse

// MsgRebindWallet is the message for rebinding a wallet during key rotation
type MsgRebindWallet = veidv1.MsgRebindWallet

// MsgRebindWalletResponse is the response for MsgRebindWallet
type MsgRebindWalletResponse = veidv1.MsgRebindWalletResponse

// MsgUpdateDerivedFeatures is the message for updating derived features
type MsgUpdateDerivedFeatures = veidv1.MsgUpdateDerivedFeatures

// MsgUpdateDerivedFeaturesResponse is the response for MsgUpdateDerivedFeatures
type MsgUpdateDerivedFeaturesResponse = veidv1.MsgUpdateDerivedFeaturesResponse

// ============================================================================
// Helper Functions
// ============================================================================

// GetAddScopeSigningMessage returns the canonical signing message for adding a scope
func GetAddScopeSigningMessage(sender, scopeID string) []byte {
	return []byte("VEID_ADD_SCOPE:" + sender + ":" + scopeID)
}

// GetRevokeScopeSigningMessage returns the canonical signing message for revoking a scope
func GetRevokeScopeSigningMessage(sender, scopeID string) []byte {
	return []byte("VEID_REVOKE_SCOPE:" + sender + ":" + scopeID)
}