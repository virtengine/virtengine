package v1beta5

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

// RegisterQueryHandlerClient is a no-op stub when gateway handlers are not generated.
func RegisterQueryHandlerClient(_ context.Context, _ *runtime.ServeMux, _ QueryClient) error {
	return nil
}
