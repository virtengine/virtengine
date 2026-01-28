package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	rolesv1 "github.com/virtengine/virtengine/sdk/go/node/roles/v1"
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
	legacy.RegisterAminoMsg(cdc, &rolesv1.MsgAssignRole{}, "roles/MsgAssignRole")
	legacy.RegisterAminoMsg(cdc, &rolesv1.MsgRevokeRole{}, "roles/MsgRevokeRole")
	legacy.RegisterAminoMsg(cdc, &rolesv1.MsgSetAccountState{}, "roles/MsgSetAccountState")
	legacy.RegisterAminoMsg(cdc, &rolesv1.MsgNominateAdmin{}, "roles/MsgNominateAdmin")
	legacy.RegisterAminoMsg(cdc, &rolesv1.MsgUpdateParams{}, "roles/MsgUpdateParams")
}

// RegisterInterfaces registers the interfaces types with the interface registry.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&rolesv1.MsgAssignRole{},
		&rolesv1.MsgRevokeRole{},
		&rolesv1.MsgSetAccountState{},
		&rolesv1.MsgNominateAdmin{},
		&rolesv1.MsgUpdateParams{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &rolesv1.Msg_serviceDesc)
}
