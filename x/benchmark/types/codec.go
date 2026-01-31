package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/gogoproto/grpc"

	benchmarkv1 "github.com/virtengine/virtengine/sdk/go/node/benchmark/v1"
)

// MsgServer is the server API for Msg service.
type MsgServer = benchmarkv1.MsgServer

// RegisterMsgServer registers the MsgServer implementation with the grpc.Server.
func RegisterMsgServer(s grpc.Server, srv MsgServer) {
	benchmarkv1.RegisterMsgServer(s, srv)
}

// ModuleCdc is the codec for the benchmark module
var ModuleCdc = codec.NewLegacyAmino()

func init() {
	RegisterLegacyAminoCodec(ModuleCdc)
}

// RegisterLegacyAminoCodec registers the benchmark types
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Note: Amino message names must be <= 39 characters
	legacy.RegisterAminoMsg(cdc, &MsgSubmitBenchmarks{}, "ve/bench/MsgSubmitBenchmarks")
	legacy.RegisterAminoMsg(cdc, &MsgRequestChallenge{}, "ve/bench/MsgRequestChallenge")
	legacy.RegisterAminoMsg(cdc, &MsgRespondChallenge{}, "ve/bench/MsgRespondChallenge")
	legacy.RegisterAminoMsg(cdc, &MsgFlagProvider{}, "ve/bench/MsgFlagProvider")
	legacy.RegisterAminoMsg(cdc, &MsgUnflagProvider{}, "ve/bench/MsgUnflagProvider")
	legacy.RegisterAminoMsg(cdc, &MsgResolveAnomalyFlag{}, "ve/bench/MsgResolveAnomaly")
}

// RegisterInterfaces registers the x/benchmark interfaces types with the interface registry
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSubmitBenchmarks{},
		&MsgRequestChallenge{},
		&MsgRespondChallenge{},
		&MsgFlagProvider{},
		&MsgUnflagProvider{},
		&MsgResolveAnomalyFlag{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &benchmarkv1.Msg_serviceDesc)
}
