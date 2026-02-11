package v1beta4

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

// RegisterQueryHandlerClient registers gRPC gateway routes.
// NOTE: Gateway handlers are not generated in this build, so this is a no-op.
func RegisterQueryHandlerClient(_ context.Context, _ *runtime.ServeMux, _ QueryClient) error {
	return nil
}
