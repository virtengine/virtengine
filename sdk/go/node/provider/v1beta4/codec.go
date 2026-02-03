package v1beta4

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

var (
	// ModuleCdc references the global x/provider module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/provider and
	// defined at the application level.
	//
	// Deprecated: ModuleCdc use is deprecated
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)

// RegisterLegacyAminoCodec register concrete types on codec
//
// Deprecated: RegisterLegacyAminoCodec is deprecated
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateProvider{}, ModuleName+"/"+msgTypeCreateProvider, nil)
	cdc.RegisterConcrete(&MsgUpdateProvider{}, ModuleName+"/"+msgTypeUpdateProvider, nil)
	cdc.RegisterConcrete(&MsgDeleteProvider{}, ModuleName+"/"+msgTypeDeleteProvider, nil)
}

// RegisterInterfaces registers the x/provider interfaces types with the interface registry
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateProvider{},
		&MsgUpdateProvider{},
		&MsgDeleteProvider{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
