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

	quantities := [...]int{1, 2, 3}
	orderReceipt, err := s.backend.CreateOrder(ctx, owner, quantities[sequence%uint64(len(quantities))], map[string]string{
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
	providerAPrices := [...]int64{120, 121, 122, 123, 124, 125, 126, 127, 128, 129, 130, 131, 132, 133, 134}
	if _, err := s.backend.SubmitBid(ctx, orderID, providerA, providerAPrices[sequence%uint64(len(providerAPrices))]); err != nil {
		return &framework.ExecutionResult{
			Success:  false,
			Duration: time.Since(start),
			Error:    err,
		}, nil
	}
	providerBPrices := [...]int64{100, 101, 102, 103, 104, 105, 106, 107, 108, 109}
	if _, err := s.backend.SubmitBid(ctx, orderID, providerB, providerBPrices[sequence%uint64(len(providerBPrices))]); err != nil {
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
