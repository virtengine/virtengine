package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/gogoproto/grpc"

	mfav1 "github.com/virtengine/virtengine/sdk/go/node/mfa/v1"
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
	legacy.RegisterAminoMsg(cdc, &mfav1.MsgEnrollFactor{}, "mfa/MsgEnrollFactor")
	legacy.RegisterAminoMsg(cdc, &mfav1.MsgRevokeFactor{}, "mfa/MsgRevokeFactor")
	legacy.RegisterAminoMsg(cdc, &mfav1.MsgSetMFAPolicy{}, "mfa/MsgSetMFAPolicy")
	legacy.RegisterAminoMsg(cdc, &mfav1.MsgCreateChallenge{}, "mfa/MsgCreateChallenge")
	legacy.RegisterAminoMsg(cdc, &mfav1.MsgVerifyChallenge{}, "mfa/MsgVerifyChallenge")
	legacy.RegisterAminoMsg(cdc, &mfav1.MsgAddTrustedDevice{}, "mfa/MsgAddTrustedDevice")
	legacy.RegisterAminoMsg(cdc, &mfav1.MsgRemoveTrustedDevice{}, "mfa/MsgRemoveTrustedDevice")
	legacy.RegisterAminoMsg(cdc, &mfav1.MsgUpdateSensitiveTxConfig{}, "mfa/MsgUpdateSensitiveTxConfig")
	legacy.RegisterAminoMsg(cdc, &mfav1.MsgUpdateParams{}, "mfa/MsgUpdateParams")
}

// RegisterInterfaces registers the interfaces types with the interface registry.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	// Register generated proto message types with the interface registry
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&mfav1.MsgEnrollFactor{},
		&mfav1.MsgRevokeFactor{},
		&mfav1.MsgSetMFAPolicy{},
		&mfav1.MsgCreateChallenge{},
		&mfav1.MsgVerifyChallenge{},
		&mfav1.MsgAddTrustedDevice{},
		&mfav1.MsgRemoveTrustedDevice{},
		&mfav1.MsgUpdateSensitiveTxConfig{},
		&mfav1.MsgUpdateParams{},
	)

	// Register the Msg service descriptor for proper gRPC routing
	msgservice.RegisterMsgServiceDesc(registry, &mfav1.Msg_serviceDesc)
}

// RegisterMsgServer registers the MsgServer implementation with the grpc.Server.
// This delegates to the generated proto registration function.
func RegisterMsgServer(s grpc.Server, srv MsgServer) {
	// Wrap the local MsgServer to match the generated interface
	mfav1.RegisterMsgServer(s, NewMsgServerAdapter(srv))
}

// RegisterQueryServer registers the QueryServer implementation with the grpc.Server.
// This delegates to the generated proto registration function.
func RegisterQueryServer(s grpc.Server, srv QueryServer) {
	// Wrap the local QueryServer to match the generated interface
	mfav1.RegisterQueryServer(s, NewQueryServerAdapter(srv))
}
