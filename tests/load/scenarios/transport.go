// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package scenarios

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type scenarioTransport struct {
	endpoint string
	conn     *grpc.ClientConn
}

func (t *scenarioTransport) setup(ctx context.Context) error {
	if t.endpoint == "" || strings.HasPrefix(t.endpoint, "mock://") {
		return nil
	}

	conn, err := grpc.NewClient(t.endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("dial grpc: %w", err)
	}

	resp, err := grpc_health_v1.NewHealthClient(conn).Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("grpc health check: %w", err)
	}
	if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		_ = conn.Close()
		return fmt.Errorf("grpc service not serving: %s", resp.Status.String())
	}

	t.conn = conn
	return nil
}

func (t *scenarioTransport) teardown() error {
	if t.conn != nil {
		return t.conn.Close()
	}
	return nil
}
