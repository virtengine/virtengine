package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// ModuleCdc is the codec for the benchmark module
var ModuleCdc = codec.NewLegacyAmino()

func init() {
	RegisterLegacyAminoCodec(ModuleCdc)
}

// RegisterLegacyAminoCodec registers the benchmark types
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgSubmitBenchmarks{}, "benchmark/MsgSubmitBenchmarks", nil)
	cdc.RegisterConcrete(&MsgRequestChallenge{}, "benchmark/MsgRequestChallenge", nil)
	cdc.RegisterConcrete(&MsgRespondChallenge{}, "benchmark/MsgRespondChallenge", nil)
	cdc.RegisterConcrete(&MsgFlagProvider{}, "benchmark/MsgFlagProvider", nil)
	cdc.RegisterConcrete(&MsgUnflagProvider{}, "benchmark/MsgUnflagProvider", nil)
	cdc.RegisterConcrete(&MsgResolveAnomalyFlag{}, "benchmark/MsgResolveAnomalyFlag", nil)
}

// RegisterInterfaces registers the x/benchmark interfaces types with the interface registry
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	// NOTE: These are stub message types without proper protobuf generation.
	// They don't have proper typeURLs (XXX_MessageName() methods), so we cannot
	// register them with RegisterImplementations. This will cause typeURL "/" conflicts.
	//
	// Once proper .proto files are generated with protoc-gen-gogo, this should be:
	//
	// registry.RegisterImplementations((*sdk.Msg)(nil),
	//     &MsgSubmitBenchmarks{},
	//     &MsgRequestChallenge{},
	//     &MsgRespondChallenge{},
	//     &MsgFlagProvider{},
	//     &MsgUnflagProvider{},
	//     &MsgResolveAnomalyFlag{},
	// )
	//
	// msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)

	// Suppress unused import warning
	_ = msgservice.RegisterMsgServiceDesc
}

// _Msg_serviceDesc is a placeholder for the message service descriptor
var _Msg_serviceDesc = struct {
	ServiceName string
}{
	ServiceName: "virtengine.benchmark.v1.Msg",
}
