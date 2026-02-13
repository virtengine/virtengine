// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/virtengine/virtengine/pkg/security"
)

// ChainHealthChecker checks blockchain health
type ChainHealthChecker struct {
	grpcEndpoint string
}

// NewChainHealthChecker creates a chain health checker
func NewChainHealthChecker(grpcEndpoint string) *ChainHealthChecker {
	return &ChainHealthChecker{
		grpcEndpoint: grpcEndpoint,
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

	// TODO: Implement actual block height check
	// For now, just check gRPC health
	client := grpc_health_v1.NewHealthClient(conn)
	resp, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		return fmt.Errorf("service not serving")
	}

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
}

// NewTransactionSubmitChecker creates a transaction submit checker
func NewTransactionSubmitChecker(grpcEndpoint string) *TransactionSubmitChecker {
	return &TransactionSubmitChecker{
		grpcEndpoint: grpcEndpoint,
	}
}

func (t *TransactionSubmitChecker) Name() string {
	return "tx_submit"
}

func (t *TransactionSubmitChecker) Check(ctx context.Context) error {
	// TODO: Implement actual transaction submission test
	// For now, just check connection
	conn, err := grpc.NewClient(t.grpcEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}
	defer conn.Close()

	return nil
}
