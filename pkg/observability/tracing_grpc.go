package observability

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc/stats"
)

// GRPCServerStatsHandler returns a gRPC stats handler with OpenTelemetry tracing.
func GRPCServerStatsHandler() stats.Handler {
	return otelgrpc.NewServerHandler()
}

// GRPCClientStatsHandler returns a gRPC stats handler with OpenTelemetry tracing.
func GRPCClientStatsHandler() stats.Handler {
	return otelgrpc.NewClientHandler()
}
