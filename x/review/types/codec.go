// Package types contains types for the Review module.
//
// VE-911: Provider public reviews - Codec registration
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

	reviewv1 "github.com/virtengine/virtengine/sdk/go/node/review/v1"
)

// ModuleCdc is the codec for the module
var ModuleCdc = codec.NewLegacyAmino()

func init() {
	RegisterLegacyAminoCodec(ModuleCdc)
}

// RegisterLegacyAminoCodec registers the necessary interfaces and concrete types
// on the provided LegacyAmino codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgSubmitReview{}, "virtengine/review/MsgSubmitReview")
	legacy.RegisterAminoMsg(cdc, &MsgDeleteReview{}, "virtengine/review/MsgDeleteReview")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "virtengine/review/MsgUpdateParams")
}

// RegisterInterfaces registers the module interface types
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSubmitReview{},
		&MsgDeleteReview{},
		&MsgUpdateParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &reviewv1.Msg_serviceDesc)
}

// MsgServer is the server API for Msg service.
type MsgServer = reviewv1.MsgServer

// RegisterMsgServer registers the MsgServer implementation with the grpc.Server.
func RegisterMsgServer(s grpc.Server, srv MsgServer) {
	reviewv1.RegisterMsgServer(s, srv)
}

// QueryServer is the server API for Query service.
type QueryServer = reviewv1.QueryServer

// RegisterQueryServer registers the QueryServer implementation with the grpc.Server.
func RegisterQueryServer(s grpc.Server, srv QueryServer) {
	reviewv1.RegisterQueryServer(s, srv)
}

// QueryClient is the client API for Query service.
type QueryClient = reviewv1.QueryClient

// NewQueryClient returns a new QueryClient.
func NewQueryClient(cc grpc.ClientConn) QueryClient {
	return reviewv1.NewQueryClient(cc)
}

// RegisterQueryHandlerClient registers the gRPC gateway routes for the query service.
func RegisterQueryHandlerClient(ctx context.Context, mux *runtime.ServeMux, client QueryClient) error {
	return reviewv1.RegisterQueryHandlerClient(ctx, mux, client)
}
