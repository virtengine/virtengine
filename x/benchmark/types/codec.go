// Package types contains types for the Benchmark module.
package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSubmitBenchmarks{},
		&MsgRequestChallenge{},
		&MsgRespondChallenge{},
		&MsgFlagProvider{},
		&MsgUnflagProvider{},
		&MsgResolveAnomalyFlag{},
	)

	// TODO: Enable when protobuf generation is complete
	// msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
	_ = msgservice.RegisterMsgServiceDesc
}

// _Msg_serviceDesc is a placeholder for the message service descriptor
var _Msg_serviceDesc = struct {
	ServiceName string
}{
	ServiceName: "virtengine.benchmark.v1.Msg",
}
