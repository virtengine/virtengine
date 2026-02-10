package types

import (
	"context"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/gogoproto/grpc"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	benchmarkv1 "github.com/virtengine/virtengine/sdk/go/node/benchmark/v1"
)

// MsgServer is the server API for Msg service.
type MsgServer = benchmarkv1.MsgServer

// RegisterMsgServer registers the MsgServer implementation with the grpc.Server.
func RegisterMsgServer(s grpc.Server, srv MsgServer) {
	benchmarkv1.RegisterMsgServer(s, srv)
}

// QueryServer is the server API for Query service.
type QueryServer = benchmarkv1.QueryServer

// RegisterQueryServer registers the QueryServer implementation with the grpc.Server.
func RegisterQueryServer(s grpc.Server, srv QueryServer) {
	benchmarkv1.RegisterQueryServer(s, srv)
}

// QueryClient is the client API for Query service.
type QueryClient = benchmarkv1.QueryClient

// NewQueryClient returns a new QueryClient.
func NewQueryClient(cc grpc.ClientConn) QueryClient {
	return benchmarkv1.NewQueryClient(cc)
}

// RegisterQueryHandlerClient registers the gRPC gateway routes for the query service.
func RegisterQueryHandlerClient(ctx context.Context, mux *runtime.ServeMux, client QueryClient) error {
	return benchmarkv1.RegisterQueryHandlerClient(ctx, mux, client)
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
