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
	legacy.RegisterAminoMsg(cdc, &veidv1.MsgUploadScope{}, "veid/MsgUploadScope")
	legacy.RegisterAminoMsg(cdc, &veidv1.MsgRevokeScope{}, "veid/MsgRevokeScope")
	legacy.RegisterAminoMsg(cdc, &veidv1.MsgRequestVerification{}, "veid/MsgRequestVerification")
	legacy.RegisterAminoMsg(cdc, &veidv1.MsgUpdateVerificationStatus{}, "veid/MsgUpdateVerificationStatus")
	legacy.RegisterAminoMsg(cdc, &veidv1.MsgUpdateScore{}, "veid/MsgUpdateScore")
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
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&veidv1.MsgUploadScope{},
		&veidv1.MsgRevokeScope{},
		&veidv1.MsgRequestVerification{},
		&veidv1.MsgUpdateVerificationStatus{},
		&veidv1.MsgUpdateScore{},
		&veidv1.MsgCreateIdentityWallet{},
		&veidv1.MsgAddScopeToWallet{},
		&veidv1.MsgRevokeScopeFromWallet{},
		&veidv1.MsgUpdateConsentSettings{},
		&veidv1.MsgRebindWallet{},
		&veidv1.MsgUpdateDerivedFeatures{},
		&veidv1.MsgCompleteBorderlineFallback{},
		&veidv1.MsgUpdateBorderlineParams{},
		&veidv1.MsgUpdateParams{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &veidv1.Msg_serviceDesc)
}
