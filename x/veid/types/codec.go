package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
)

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)

func init() {
	RegisterLegacyAminoCodec(amino)
}

// RegisterLegacyAminoCodec registers the necessary interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Only register messages that are part of the SDK proto Msg service
	legacy.RegisterAminoMsg(cdc, &veidv1.MsgUploadScope{}, "veid/MsgUploadScope")
	legacy.RegisterAminoMsg(cdc, &veidv1.MsgRevokeScope{}, "veid/MsgRevokeScope")
	legacy.RegisterAminoMsg(cdc, &veidv1.MsgRequestVerification{}, "veid/MsgRequestVerification")
	legacy.RegisterAminoMsg(cdc, &veidv1.MsgUpdateVerificationStatus{}, "veid/MsgUpdateVerificationStatus")
	legacy.RegisterAminoMsg(cdc, &veidv1.MsgUpdateScore{}, "veid/MsgUpdateScore")
	legacy.RegisterAminoMsg(cdc, &veidv1.MsgSubmitSSOVerificationProof{}, "veid/MsgSubmitSSOVerificationProof")
	legacy.RegisterAminoMsg(cdc, &veidv1.MsgSubmitEmailVerificationProof{}, "veid/MsgSubmitEmailVerificationProof")
	legacy.RegisterAminoMsg(cdc, &veidv1.MsgSubmitSMSVerificationProof{}, "veid/MsgSubmitSMSVerificationProof")
	legacy.RegisterAminoMsg(cdc, &veidv1.MsgSubmitSocialMediaScope{}, "veid/MsgSubmitSocialMediaScope")
	// Wallet messages
	legacy.RegisterAminoMsg(cdc, &veidv1.MsgCreateIdentityWallet{}, "veid/MsgCreateIdentityWallet")
	legacy.RegisterAminoMsg(cdc, &veidv1.MsgAddScopeToWallet{}, "veid/MsgAddScopeToWallet")
	legacy.RegisterAminoMsg(cdc, &veidv1.MsgRevokeScopeFromWallet{}, "veid/MsgRevokeScopeFromWallet")
	legacy.RegisterAminoMsg(cdc, &veidv1.MsgUpdateConsentSettings{}, "veid/MsgUpdateConsentSettings")
	legacy.RegisterAminoMsg(cdc, &veidv1.MsgRebindWallet{}, "veid/MsgRebindWallet")
	legacy.RegisterAminoMsg(cdc, &veidv1.MsgUpdateDerivedFeatures{}, "veid/MsgUpdateDerivedFeatures")
	// Borderline fallback messages
	legacy.RegisterAminoMsg(cdc, &veidv1.MsgCompleteBorderlineFallback{}, "veid/MsgCompleteBorderlineFallback")
	legacy.RegisterAminoMsg(cdc, &veidv1.MsgUpdateBorderlineParams{}, "veid/MsgUpdateBorderlineParams")
	// Params message
	legacy.RegisterAminoMsg(cdc, &veidv1.MsgUpdateParams{}, "veid/MsgUpdateParams")
}

// RegisterInterfaces registers the interfaces types with the interface registry.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	// Only register messages that are part of the SDK proto Msg service
	// Appeal, Compliance, and Model messages are in separate proto files
	// and have their own service definitions (not part of the main Msg service)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		// Core messages (in tx.proto Msg service)
		&veidv1.MsgUploadScope{},
		&veidv1.MsgRevokeScope{},
		&veidv1.MsgRequestVerification{},
		&veidv1.MsgUpdateVerificationStatus{},
		&veidv1.MsgUpdateScore{},
		&veidv1.MsgSubmitSSOVerificationProof{},
		&veidv1.MsgSubmitEmailVerificationProof{},
		&veidv1.MsgSubmitSMSVerificationProof{},
		&veidv1.MsgSubmitSocialMediaScope{},
		// Wallet messages (in tx.proto Msg service)
		&veidv1.MsgCreateIdentityWallet{},
		&veidv1.MsgAddScopeToWallet{},
		&veidv1.MsgRevokeScopeFromWallet{},
		&veidv1.MsgUpdateConsentSettings{},
		&veidv1.MsgRebindWallet{},
		&veidv1.MsgUpdateDerivedFeatures{},
		// Borderline fallback messages (in tx.proto Msg service)
		&veidv1.MsgCompleteBorderlineFallback{},
		&veidv1.MsgUpdateBorderlineParams{},
		// Params message (in tx.proto Msg service)
		&veidv1.MsgUpdateParams{},
	)
	// Use the SDK-generated service descriptor which has proper proto metadata
	msgservice.RegisterMsgServiceDesc(registry, &veidv1.Msg_serviceDesc)
}
