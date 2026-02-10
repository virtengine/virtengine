// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package scenarios

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/virtengine/virtengine/tests/load/framework"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// VEIDSubmitScenario simulates VEID scope submission load
type VEIDSubmitScenario struct {
	grpcEndpoint string
	conn         *grpc.ClientConn
	accounts     []string
	scopeSize    int
}

// NewVEIDSubmitScenario creates a new VEID submit scenario
func NewVEIDSubmitScenario(grpcEndpoint string, accounts []string) *VEIDSubmitScenario {
	return &VEIDSubmitScenario{
		grpcEndpoint: grpcEndpoint,
		accounts:     accounts,
		scopeSize:    32 * 1024, // 32KB default
	}
}

// Name returns the scenario name
func (s *VEIDSubmitScenario) Name() string {
	return "veid_submit"
}

// Setup initializes the gRPC connection
func (s *VEIDSubmitScenario) Setup(ctx context.Context) error {
	conn, err := grpc.NewClient(s.grpcEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("dial grpc: %w", err)
	}
	s.conn = conn
	return nil
}

// Execute performs a single VEID submission
func (s *VEIDSubmitScenario) Execute(ctx context.Context) (*framework.ExecutionResult, error) {
	start := time.Now()

	scopeData := make([]byte, s.scopeSize)
	if _, err := rand.Read(scopeData); err != nil {
		return &framework.ExecutionResult{
			Success:  false,
			Duration: time.Since(start),
			Error:    err,
		}, nil
	}

	accountIdx := int(time.Now().UnixNano()) % len(s.accounts)
	account := s.accounts[accountIdx]

	// TODO: Implement actual VEID submission via gRPC
	// For now, simulate execution
	time.Sleep(10 * time.Millisecond)

	return &framework.ExecutionResult{
		Success:  true,
		Duration: time.Since(start),
		Metadata: map[string]interface{}{
			"account":   account,
			"data_size": s.scopeSize,
		},
	}, nil
}

// Teardown closes the gRPC connection
func (s *VEIDSubmitScenario) Teardown(ctx context.Context) error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}
