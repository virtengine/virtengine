// Copyright 2024-2026 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0

package slo

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/virtengine/virtengine/pkg/chaos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
)

func TestCheckProbeGRPCServing(t *testing.T) {
	grpcAddr, cleanup := startVerifierGRPCServer(t, verifierGRPCServerConfig{
		statuses: map[string]grpc_health_v1.HealthCheckResponse_ServingStatus{
			"": grpc_health_v1.HealthCheckResponse_SERVING,
		},
	})
	defer cleanup()

	value, err := NewVerifier().CheckProbe(context.Background(), chaos.Probe{
		Name:            "grpc-serving",
		Type:            chaos.ProbeTypeGRPC,
		URL:             "grpc://" + grpcAddr,
		Timeout:         2 * time.Second,
		SuccessCriteria: "status == 1",
	})

	require.NoError(t, err)
	require.Equal(t, float64(grpc_health_v1.HealthCheckResponse_SERVING), value)
}

func TestCheckProbeGRPCSupportsNamedServicesAndMetadata(t *testing.T) {
	var (
		mu       sync.Mutex
		captured metadata.MD
	)

	grpcAddr, cleanup := startVerifierGRPCServer(t, verifierGRPCServerConfig{
		statuses: map[string]grpc_health_v1.HealthCheckResponse_ServingStatus{
			"provider.Health": grpc_health_v1.HealthCheckResponse_SERVING,
		},
		onMetadata: func(md metadata.MD) {
			mu.Lock()
			defer mu.Unlock()
			captured = md.Copy()
		},
	})
	defer cleanup()

	value, err := NewVerifier().CheckProbe(context.Background(), chaos.Probe{
		Name:    "grpc-named-service",
		Type:    chaos.ProbeTypeGRPC,
		URL:     "grpc://" + grpcAddr + "/provider.Health",
		Timeout: 2 * time.Second,
		Headers: map[string]string{
			"authorization": "Bearer probe-token",
		},
	})

	require.NoError(t, err)
	require.Equal(t, float64(grpc_health_v1.HealthCheckResponse_SERVING), value)

	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, []string{"Bearer probe-token"}, captured.Get("authorization"))
}

func TestCheckProbeGRPCReturnsLatencyWhenRequested(t *testing.T) {
	grpcAddr, cleanup := startVerifierGRPCServer(t, verifierGRPCServerConfig{
		statuses: map[string]grpc_health_v1.HealthCheckResponse_ServingStatus{
			"": grpc_health_v1.HealthCheckResponse_SERVING,
		},
		delay: 25 * time.Millisecond,
	})
	defer cleanup()

	value, err := NewVerifier().CheckProbe(context.Background(), chaos.Probe{
		Name:            "grpc-latency",
		Type:            chaos.ProbeTypeGRPC,
		URL:             grpcAddr,
		Timeout:         2 * time.Second,
		SuccessCriteria: "latency < 1",
	})

	require.NoError(t, err)
	require.Greater(t, value, 0.0)
}

func TestCheckProbeGRPCFailsWhenServiceIsNotServing(t *testing.T) {
	grpcAddr, cleanup := startVerifierGRPCServer(t, verifierGRPCServerConfig{
		statuses: map[string]grpc_health_v1.HealthCheckResponse_ServingStatus{
			"": grpc_health_v1.HealthCheckResponse_NOT_SERVING,
		},
	})
	defer cleanup()

	_, err := NewVerifier().CheckProbe(context.Background(), chaos.Probe{
		Name:    "grpc-not-serving",
		Type:    chaos.ProbeTypeGRPC,
		URL:     grpcAddr,
		Timeout: 2 * time.Second,
	})

	require.ErrorContains(t, err, "not serving")
}

type verifierGRPCServerConfig struct {
	statuses   map[string]grpc_health_v1.HealthCheckResponse_ServingStatus
	delay      time.Duration
	onMetadata func(metadata.MD)
}

func startVerifierGRPCServer(t *testing.T, cfg verifierGRPCServerConfig) (string, func()) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	server := grpc.NewServer(grpc.UnaryInterceptor(func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if cfg.onMetadata != nil {
			if md, ok := metadata.FromIncomingContext(ctx); ok {
				cfg.onMetadata(md)
			}
		}

		if cfg.delay > 0 {
			time.Sleep(cfg.delay)
		}

		return handler(ctx, req)
	}))

	healthServer := health.NewServer()
	for service, status := range cfg.statuses {
		healthServer.SetServingStatus(service, status)
	}
	grpc_health_v1.RegisterHealthServer(server, healthServer)

	go func() {
		_ = server.Serve(listener)
	}()

	return listener.Addr().String(), func() {
		server.Stop()
		_ = listener.Close()
	}
}
