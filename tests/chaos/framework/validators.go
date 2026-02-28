// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/virtengine/virtengine/pkg/security"
)

// ChainHealthChecker checks blockchain health
type ChainHealthChecker struct {
	grpcEndpoint   string
	statusEndpoint string
	client         *http.Client

	mu         sync.Mutex
	lastHeight int64
	hasHeight  bool
}

// NewChainHealthChecker creates a chain health checker
func NewChainHealthChecker(grpcEndpoint string) *ChainHealthChecker {
	return NewChainHealthCheckerWithStatusEndpoint(grpcEndpoint, "")
}

// NewChainHealthCheckerWithStatusEndpoint creates a chain health checker with an explicit RPC status endpoint.
func NewChainHealthCheckerWithStatusEndpoint(grpcEndpoint, statusEndpoint string) *ChainHealthChecker {
	if statusEndpoint == "" {
		statusEndpoint = deriveStatusEndpoint(grpcEndpoint)
	}

	return &ChainHealthChecker{
		grpcEndpoint:   grpcEndpoint,
		statusEndpoint: statusEndpoint,
		client:         security.NewSecureHTTPClient(security.WithTimeout(5 * time.Second)),
	}
}

func (c *ChainHealthChecker) Name() string {
	return "chain_health"
}

func (c *ChainHealthChecker) Check(ctx context.Context) error {
	conn, err := grpc.NewClient(c.grpcEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}
	defer conn.Close()

	client := grpc_health_v1.NewHealthClient(conn)
	resp, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		return fmt.Errorf("service not serving")
	}

	height, err := c.fetchLatestHeight(ctx)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.hasHeight && height <= c.lastHeight {
		return fmt.Errorf("block height stalled: current=%d last=%d", height, c.lastHeight)
	}

	c.lastHeight = height
	c.hasHeight = true

	return nil
}

// APIEndpointChecker checks HTTP API availability
type APIEndpointChecker struct {
	endpoint string
	client   *http.Client
}

// NewAPIEndpointChecker creates an API endpoint checker
func NewAPIEndpointChecker(endpoint string) *APIEndpointChecker {
	return &APIEndpointChecker{
		endpoint: endpoint,
		client:   security.NewSecureHTTPClient(security.WithTimeout(5 * time.Second)),
	}
}

func (a *APIEndpointChecker) Name() string {
	return "api_endpoint"
}

func (a *APIEndpointChecker) Check(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", a.endpoint, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// TransactionSubmitChecker checks transaction submission capability
type TransactionSubmitChecker struct {
	grpcEndpoint string
	probeTx      func(context.Context) ([]byte, error)
}

// NewTransactionSubmitChecker creates a transaction submit checker
func NewTransactionSubmitChecker(grpcEndpoint string) *TransactionSubmitChecker {
	return NewTransactionSubmitCheckerWithProbe(grpcEndpoint, nil)
}

// NewTransactionSubmitCheckerWithProbe creates a transaction submit checker with an explicit probe tx payload provider.
func NewTransactionSubmitCheckerWithProbe(grpcEndpoint string, probeTx func(context.Context) ([]byte, error)) *TransactionSubmitChecker {
	if probeTx == nil {
		probeTx = func(context.Context) ([]byte, error) {
			return []byte("virtengine-chaos-submit-probe"), nil
		}
	}

	return &TransactionSubmitChecker{
		grpcEndpoint: grpcEndpoint,
		probeTx:      probeTx,
	}
}

func (t *TransactionSubmitChecker) Name() string {
	return "tx_submit"
}

func (t *TransactionSubmitChecker) Check(ctx context.Context) error {
	conn, err := grpc.NewClient(t.grpcEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}
	defer conn.Close()

	probeTx, err := t.probeTx(ctx)
	if err != nil {
		return fmt.Errorf("build probe tx: %w", err)
	}
	if len(probeTx) == 0 {
		return fmt.Errorf("probe tx bytes are empty")
	}

	resp, err := txtypes.NewServiceClient(conn).BroadcastTx(ctx, &txtypes.BroadcastTxRequest{
		TxBytes: probeTx,
		Mode:    txtypes.BroadcastMode_BROADCAST_MODE_SYNC,
	})
	if err != nil {
		return fmt.Errorf("broadcast tx failed: %w", err)
	}
	if resp == nil || resp.TxResponse == nil {
		return fmt.Errorf("broadcast tx returned empty response")
	}
	if strings.TrimSpace(resp.TxResponse.TxHash) == "" && strings.TrimSpace(resp.TxResponse.RawLog) == "" {
		return fmt.Errorf("broadcast tx response missing tx hash and raw log")
	}

	return nil
}

type rpcStatusResponse struct {
	Result struct {
		SyncInfo rpcSyncInfo `json:"sync_info"`
	} `json:"result"`
	SyncInfo rpcSyncInfo `json:"sync_info"`
}

type rpcSyncInfo struct {
	LatestBlockHeight string `json:"latest_block_height"`
}

func (c *ChainHealthChecker) fetchLatestHeight(ctx context.Context) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.statusEndpoint, nil)
	if err != nil {
		return 0, fmt.Errorf("create status request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("status request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("status endpoint returned %d", resp.StatusCode)
	}

	var status rpcStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return 0, fmt.Errorf("decode status response: %w", err)
	}

	heightText := status.Result.SyncInfo.LatestBlockHeight
	if heightText == "" {
		heightText = status.SyncInfo.LatestBlockHeight
	}
	if heightText == "" {
		return 0, fmt.Errorf("status response missing latest block height")
	}

	height, err := strconv.ParseInt(heightText, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse latest block height %q: %w", heightText, err)
	}
	if height <= 0 {
		return 0, fmt.Errorf("latest block height must be positive, got %d", height)
	}

	return height, nil
}

func deriveStatusEndpoint(grpcEndpoint string) string {
	hostPort := grpcEndpoint
	if strings.Contains(grpcEndpoint, "://") {
		parsed, err := url.Parse(grpcEndpoint)
		if err == nil && parsed.Host != "" {
			hostPort = parsed.Host
		}
	}

	host, _, err := net.SplitHostPort(hostPort)
	if err != nil {
		host = hostPort
	}
	if host == "" {
		host = "127.0.0.1"
	}

	return "http://" + net.JoinHostPort(host, "26657") + "/status"
}
