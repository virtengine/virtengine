package v1beta4

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/virtengine/virtengine/sdk/go/sdkutil"
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

func init() {
	sdkutil.RegisterCustomSignerField(&MsgCreateProvider{}, "owner", "")
	sdkutil.RegisterCustomSignerField(&MsgUpdateProvider{}, "owner", "")
	sdkutil.RegisterCustomSignerField(&MsgDeleteProvider{}, "owner", "")
	sdkutil.RegisterCustomSignerField(&MsgGenerateDomainVerificationToken{}, "owner", "")
	sdkutil.RegisterCustomSignerField(&MsgVerifyProviderDomain{}, "owner", "")
	sdkutil.RegisterCustomSignerField(&MsgRequestDomainVerification{}, "owner", "")
	sdkutil.RegisterCustomSignerField(&MsgConfirmDomainVerification{}, "owner", "")
	sdkutil.RegisterCustomSignerField(&MsgRevokeDomainVerification{}, "owner", "")
}

// RegisterLegacyAminoCodec register concrete types on codec
//
// Deprecated: RegisterLegacyAminoCodec is deprecated
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateProvider{}, ModuleName+"/"+msgTypeCreateProvider, nil)
	cdc.RegisterConcrete(&MsgUpdateProvider{}, ModuleName+"/"+msgTypeUpdateProvider, nil)
	cdc.RegisterConcrete(&MsgDeleteProvider{}, ModuleName+"/"+msgTypeDeleteProvider, nil)
	cdc.RegisterConcrete(&MsgRequestDomainVerification{}, ModuleName+"/"+msgTypeRequestDomainVerification, nil)
	cdc.RegisterConcrete(&MsgConfirmDomainVerification{}, ModuleName+"/"+msgTypeConfirmDomainVerification, nil)
	cdc.RegisterConcrete(&MsgRevokeDomainVerification{}, ModuleName+"/"+msgTypeRevokeDomainVerification, nil)
}

// RegisterInterfaces registers the x/provider interfaces types with the interface registry
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateProvider{},
		&MsgUpdateProvider{},
		&MsgDeleteProvider{},
		&MsgGenerateDomainVerificationToken{},
		&MsgVerifyProviderDomain{},
		&MsgRequestDomainVerification{},
		&MsgConfirmDomainVerification{},
		&MsgRevokeDomainVerification{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
