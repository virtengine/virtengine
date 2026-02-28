// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func TestChainHealthCheckerRequiresAdvancingHeight(t *testing.T) {
	var height atomic.Int64
	height.Store(10)

	statusServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"result": map[string]any{
				"sync_info": map[string]any{
					"latest_block_height": fmt.Sprintf("%d", height.Load()),
				},
			},
		}))
	}))
	defer statusServer.Close()

	grpcAddr, cleanup := startChaosGRPCServer(t, nil, grpc_health_v1.HealthCheckResponse_SERVING)
	defer cleanup()

	checker := NewChainHealthCheckerWithStatusEndpoint(grpcAddr, statusServer.URL)

	require.NoError(t, checker.Check(context.Background()))

	err := checker.Check(context.Background())
	require.ErrorContains(t, err, "block height stalled")

	height.Store(11)
	require.NoError(t, checker.Check(context.Background()))
}

func TestChainHealthCheckerFailsWhenServiceNotServing(t *testing.T) {
	statusServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"result": map[string]any{
				"sync_info": map[string]any{
					"latest_block_height": "1",
				},
			},
		}))
	}))
	defer statusServer.Close()

	grpcAddr, cleanup := startChaosGRPCServer(t, nil, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	defer cleanup()

	checker := NewChainHealthCheckerWithStatusEndpoint(grpcAddr, statusServer.URL)
	err := checker.Check(context.Background())
	require.ErrorContains(t, err, "service not serving")
}

func TestTransactionSubmitCheckerBroadcastsProbeTx(t *testing.T) {
	txServer := &stubTxService{
		resp: &txtypes.BroadcastTxResponse{
			TxResponse: &sdk.TxResponse{
				Code:      4,
				RawLog:    "signature verification failed",
				TxHash:    "ABC123",
				Height:    0,
				Codespace: "sdk",
			},
		},
	}

	grpcAddr, cleanup := startChaosGRPCServer(t, txServer, grpc_health_v1.HealthCheckResponse_SERVING)
	defer cleanup()

	checker := NewTransactionSubmitCheckerWithProbe(grpcAddr, func(context.Context) ([]byte, error) {
		return []byte("probe-payload"), nil
	})

	require.NoError(t, checker.Check(context.Background()))
	require.Equal(t, []byte("probe-payload"), txServer.lastBroadcastBytes())
}

func TestTransactionSubmitCheckerRejectsEmptyBroadcastResponse(t *testing.T) {
	txServer := &stubTxService{
		resp: &txtypes.BroadcastTxResponse{
			TxResponse: &sdk.TxResponse{},
		},
	}

	grpcAddr, cleanup := startChaosGRPCServer(t, txServer, grpc_health_v1.HealthCheckResponse_SERVING)
	defer cleanup()

	checker := NewTransactionSubmitCheckerWithProbe(grpcAddr, func(context.Context) ([]byte, error) {
		return []byte("probe-payload"), nil
	})

	err := checker.Check(context.Background())
	require.ErrorContains(t, err, "missing tx hash and raw log")
}

type stubTxService struct {
	txtypes.UnimplementedServiceServer

	mu   sync.Mutex
	req  *txtypes.BroadcastTxRequest
	resp *txtypes.BroadcastTxResponse
	err  error
}

func (s *stubTxService) BroadcastTx(_ context.Context, req *txtypes.BroadcastTxRequest) (*txtypes.BroadcastTxResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.req = req
	return s.resp, s.err
}

func (s *stubTxService) lastBroadcastBytes() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.req == nil {
		return nil
	}
	return append([]byte(nil), s.req.TxBytes...)
}

func startChaosGRPCServer(t *testing.T, txServer txtypes.ServiceServer, healthStatus grpc_health_v1.HealthCheckResponse_ServingStatus) (string, func()) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	server := grpc.NewServer()

	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthStatus)
	grpc_health_v1.RegisterHealthServer(server, healthServer)

	if txServer != nil {
		txtypes.RegisterServiceServer(server, txServer)
	}

	go func() {
		_ = server.Serve(listener)
	}()

	return listener.Addr().String(), func() {
		server.Stop()
		_ = listener.Close()
	}
}
