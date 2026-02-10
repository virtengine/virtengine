package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	encryptionv1 "github.com/virtengine/virtengine/sdk/go/node/encryption/v1"
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
	legacy.RegisterAminoMsg(cdc, &MsgRegisterRecipientKey{}, "encryption/MsgRegisterRecipientKey")
	legacy.RegisterAminoMsg(cdc, &MsgRevokeRecipientKey{}, "encryption/MsgRevokeRecipientKey")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateKeyLabel{}, "encryption/MsgUpdateKeyLabel")
	legacy.RegisterAminoMsg(cdc, &MsgRotateKey{}, "encryption/MsgRotateKey")
}

// RegisterInterfaces registers the interfaces types with the interface registry.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRegisterRecipientKey{},
		&MsgRevokeRecipientKey{},
		&MsgUpdateKeyLabel{},
		&MsgRotateKey{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &encryptionv1.Msg_serviceDesc)
}

// MsgServer is the interface for the message server - alias to generated type
type MsgServer = encryptionv1.MsgServer

// RegisterMsgServer registers the MsgServer implementation
var RegisterMsgServer = encryptionv1.RegisterMsgServer
