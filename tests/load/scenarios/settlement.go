// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package scenarios

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/virtengine/virtengine/tests/load/framework"
)

// SettlementScenario executes deterministic order-to-settlement lifecycle probes.
type SettlementScenario struct {
	transport scenarioTransport
	accounts  []string
	backend   scenarioBackend
	sequence  atomic.Uint64
}

// NewSettlementScenario creates a new settlement scenario.
func NewSettlementScenario(grpcEndpoint string, accounts []string) *SettlementScenario {
	return &SettlementScenario{
		transport: scenarioTransport{endpoint: grpcEndpoint},
		accounts:  append([]string(nil), accounts...),
		backend:   newDeterministicBackend(),
	}
}

func (s *SettlementScenario) Name() string {
	return "settlement"
}

func (s *SettlementScenario) Setup(ctx context.Context) error {
	if len(s.accounts) < 3 {
		return fmt.Errorf("settlement requires at least three accounts")
	}
	return s.transport.setup(ctx)
}

func (s *SettlementScenario) Execute(ctx context.Context) (*framework.ExecutionResult, error) {
	start := time.Now()
	sequence := s.sequence.Add(1)

	owner := s.accounts[(sequence-1)%uint64(len(s.accounts))]
	providerA := s.accounts[sequence%uint64(len(s.accounts))]
	providerB := s.accounts[(sequence+1)%uint64(len(s.accounts))]

	orderReceipt, err := s.backend.CreateOrder(ctx, owner, int(sequence%3)+1, map[string]string{
		"region": regionForSequence(sequence),
		"tier":   "settlement",
	})
	if err != nil {
		return &framework.ExecutionResult{
			Success:  false,
			Duration: time.Since(start),
			Error:    err,
		}, nil
	}

	orderID := orderReceipt.ID
	if _, err := s.backend.SubmitBid(ctx, orderID, providerA, 120+int64(sequence%15)); err != nil {
		return &framework.ExecutionResult{
			Success:  false,
			Duration: time.Since(start),
			Error:    err,
		}, nil
	}
	if _, err := s.backend.SubmitBid(ctx, orderID, providerB, 100+int64(sequence%10)); err != nil {
		return &framework.ExecutionResult{
			Success:  false,
			Duration: time.Since(start),
			Error:    err,
		}, nil
	}

	settlementReceipt, err := s.backend.SettleOrder(ctx, orderID)
	if err != nil {
		return &framework.ExecutionResult{
			Success:  false,
			Duration: time.Since(start),
			Error:    err,
		}, nil
	}

	metadata := cloneMetadata(settlementReceipt.Metadata)
	metadata["settlement_id"] = settlementReceipt.ID

	return &framework.ExecutionResult{
		Success:  true,
		Duration: time.Since(start),
		Metadata: metadata,
	}, nil
}

func (s *SettlementScenario) Teardown(ctx context.Context) error {
	_ = ctx
	return s.transport.teardown()
}
