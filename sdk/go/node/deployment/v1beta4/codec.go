package v1beta4

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	v1 "github.com/virtengine/virtengine/sdk/go/node/deployment/v1"
	"github.com/virtengine/virtengine/sdk/go/sdkutil"
)

var (
	// ModuleCdc references the global x/deployment module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/deployment and
	// defined at the application level.
	//
	// Deprecated: ModuleCdc use is deprecated
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)

func init() {
	sdkutil.RegisterCustomSignerField(&MsgCreateDeployment{}, "id", "owner")
	sdkutil.RegisterCustomSignerField(&MsgUpdateDeployment{}, "id", "owner")
	sdkutil.RegisterCustomSignerField(&MsgCloseDeployment{}, "id", "owner")
	sdkutil.RegisterCustomSignerField(&MsgStartGroup{}, "id", "owner")
	sdkutil.RegisterCustomSignerField(&MsgPauseGroup{}, "id", "owner")
	sdkutil.RegisterCustomSignerField(&MsgCloseGroup{}, "id", "owner")
}

// RegisterLegacyAminoCodec register concrete types on codec
//
// Deprecated: RegisterLegacyAminoCodec is deprecated
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateDeployment{}, "virtengine-sdk/x/"+v1.ModuleName+"/"+(&MsgCreateDeployment{}).Type(), nil)
	cdc.RegisterConcrete(&MsgUpdateDeployment{}, "virtengine-sdk/x/"+v1.ModuleName+"/"+(&MsgUpdateDeployment{}).Type(), nil)
	cdc.RegisterConcrete(&MsgCloseDeployment{}, "virtengine-sdk/x/"+v1.ModuleName+"/"+(&MsgCloseDeployment{}).Type(), nil)
	cdc.RegisterConcrete(&MsgStartGroup{}, "virtengine-sdk/x/"+v1.ModuleName+"/"+(&MsgStartGroup{}).Type(), nil)
	cdc.RegisterConcrete(&MsgPauseGroup{}, "virtengine-sdk/x/"+v1.ModuleName+"/"+(&MsgPauseGroup{}).Type(), nil)
	cdc.RegisterConcrete(&MsgCloseGroup{}, "virtengine-sdk/x/"+v1.ModuleName+"/"+(&MsgCloseGroup{}).Type(), nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "virtengine-sdk/x/"+v1.ModuleName+"/"+(&MsgUpdateParams{}).Type(), nil)
}

// RegisterInterfaces registers the x/deployment interfaces types with the interface registry
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateDeployment{},
		&MsgUpdateDeployment{},
		&MsgCloseDeployment{},
		&MsgStartGroup{},
		&MsgPauseGroup{},
		&MsgCloseGroup{},
		&MsgUpdateParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

