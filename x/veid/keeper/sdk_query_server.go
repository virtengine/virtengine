package keeper

import (
	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
)

// SDKQueryServer wraps the Keeper to implement the SDK's QueryServer interface.
// It embeds UnimplementedQueryServer to provide forward compatibility and stub
// implementations for methods not yet implemented.
type SDKQueryServer struct {
	veidv1.UnimplementedQueryServer
	keeper Keeper
}

// NewSDKQueryServer creates a new SDKQueryServer
func NewSDKQueryServer(k Keeper) *SDKQueryServer {
	return &SDKQueryServer{keeper: k}
}

var _ veidv1.QueryServer = (*SDKQueryServer)(nil)

// Note: All query methods are provided by UnimplementedQueryServer returning
// "method not implemented" errors. Implement specific methods as needed.
// The existing GRPCQuerier in grpc_query.go implements the internal types.QueryServer
// interface for backward compatibility.
