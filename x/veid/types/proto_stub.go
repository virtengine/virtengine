// Package types contains proto.Message stub implementations for the veid module.
//
// These are temporary stub implementations until proper protobuf generation is set up.
// They implement the proto.Message interface required by Cosmos SDK.
package types

import "fmt"

// ============================================================================
// Core VEID Messages (from msgs.go)
// ============================================================================

// Proto.Message interface stubs for MsgUploadScope
func (m *MsgUploadScope) ProtoMessage()  {}
func (m *MsgUploadScope) Reset()         { *m = MsgUploadScope{} }
func (m *MsgUploadScope) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgUploadScopeResponse
func (m *MsgUploadScopeResponse) ProtoMessage()  {}
func (m *MsgUploadScopeResponse) Reset()         { *m = MsgUploadScopeResponse{} }
func (m *MsgUploadScopeResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRevokeScope
func (m *MsgRevokeScope) ProtoMessage()  {}
func (m *MsgRevokeScope) Reset()         { *m = MsgRevokeScope{} }
func (m *MsgRevokeScope) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRevokeScopeResponse
func (m *MsgRevokeScopeResponse) ProtoMessage()  {}
func (m *MsgRevokeScopeResponse) Reset()         { *m = MsgRevokeScopeResponse{} }
func (m *MsgRevokeScopeResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRequestVerification
func (m *MsgRequestVerification) ProtoMessage()  {}
func (m *MsgRequestVerification) Reset()         { *m = MsgRequestVerification{} }
func (m *MsgRequestVerification) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRequestVerificationResponse
func (m *MsgRequestVerificationResponse) ProtoMessage()  {}
func (m *MsgRequestVerificationResponse) Reset()         { *m = MsgRequestVerificationResponse{} }
func (m *MsgRequestVerificationResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgUpdateVerificationStatus
func (m *MsgUpdateVerificationStatus) ProtoMessage()  {}
func (m *MsgUpdateVerificationStatus) Reset()         { *m = MsgUpdateVerificationStatus{} }
func (m *MsgUpdateVerificationStatus) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgUpdateVerificationStatusResponse
func (m *MsgUpdateVerificationStatusResponse) ProtoMessage()  {}
func (m *MsgUpdateVerificationStatusResponse) Reset()         { *m = MsgUpdateVerificationStatusResponse{} }
func (m *MsgUpdateVerificationStatusResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgUpdateScore
func (m *MsgUpdateScore) ProtoMessage()  {}
func (m *MsgUpdateScore) Reset()         { *m = MsgUpdateScore{} }
func (m *MsgUpdateScore) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgUpdateScoreResponse
func (m *MsgUpdateScoreResponse) ProtoMessage()  {}
func (m *MsgUpdateScoreResponse) Reset()         { *m = MsgUpdateScoreResponse{} }
func (m *MsgUpdateScoreResponse) String() string { return fmt.Sprintf("%+v", *m) }

// ============================================================================
// Wallet Messages (from wallet_msgs.go)
// ============================================================================

// Proto.Message interface stubs for MsgCreateIdentityWallet
func (m *MsgCreateIdentityWallet) ProtoMessage()  {}
func (m *MsgCreateIdentityWallet) Reset()         { *m = MsgCreateIdentityWallet{} }
func (m *MsgCreateIdentityWallet) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgCreateIdentityWalletResponse
func (m *MsgCreateIdentityWalletResponse) ProtoMessage()  {}
func (m *MsgCreateIdentityWalletResponse) Reset()         { *m = MsgCreateIdentityWalletResponse{} }
func (m *MsgCreateIdentityWalletResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgAddScopeToWallet
func (m *MsgAddScopeToWallet) ProtoMessage()  {}
func (m *MsgAddScopeToWallet) Reset()         { *m = MsgAddScopeToWallet{} }
func (m *MsgAddScopeToWallet) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgAddScopeToWalletResponse
func (m *MsgAddScopeToWalletResponse) ProtoMessage()  {}
func (m *MsgAddScopeToWalletResponse) Reset()         { *m = MsgAddScopeToWalletResponse{} }
func (m *MsgAddScopeToWalletResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRevokeScopeFromWallet
func (m *MsgRevokeScopeFromWallet) ProtoMessage()  {}
func (m *MsgRevokeScopeFromWallet) Reset()         { *m = MsgRevokeScopeFromWallet{} }
func (m *MsgRevokeScopeFromWallet) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRevokeScopeFromWalletResponse
func (m *MsgRevokeScopeFromWalletResponse) ProtoMessage()  {}
func (m *MsgRevokeScopeFromWalletResponse) Reset()         { *m = MsgRevokeScopeFromWalletResponse{} }
func (m *MsgRevokeScopeFromWalletResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgUpdateConsentSettings
func (m *MsgUpdateConsentSettings) ProtoMessage()  {}
func (m *MsgUpdateConsentSettings) Reset()         { *m = MsgUpdateConsentSettings{} }
func (m *MsgUpdateConsentSettings) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgUpdateConsentSettingsResponse
func (m *MsgUpdateConsentSettingsResponse) ProtoMessage()  {}
func (m *MsgUpdateConsentSettingsResponse) Reset()         { *m = MsgUpdateConsentSettingsResponse{} }
func (m *MsgUpdateConsentSettingsResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRebindWallet
func (m *MsgRebindWallet) ProtoMessage()  {}
func (m *MsgRebindWallet) Reset()         { *m = MsgRebindWallet{} }
func (m *MsgRebindWallet) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRebindWalletResponse
func (m *MsgRebindWalletResponse) ProtoMessage()  {}
func (m *MsgRebindWalletResponse) Reset()         { *m = MsgRebindWalletResponse{} }
func (m *MsgRebindWalletResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgUpdateDerivedFeatures
func (m *MsgUpdateDerivedFeatures) ProtoMessage()  {}
func (m *MsgUpdateDerivedFeatures) Reset()         { *m = MsgUpdateDerivedFeatures{} }
func (m *MsgUpdateDerivedFeatures) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgUpdateDerivedFeaturesResponse
func (m *MsgUpdateDerivedFeaturesResponse) ProtoMessage()  {}
func (m *MsgUpdateDerivedFeaturesResponse) Reset()         { *m = MsgUpdateDerivedFeaturesResponse{} }
func (m *MsgUpdateDerivedFeaturesResponse) String() string { return fmt.Sprintf("%+v", *m) }

