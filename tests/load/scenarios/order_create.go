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

// OrderCreateScenario executes deterministic marketplace order creation probes.
type OrderCreateScenario struct {
	transport scenarioTransport
	accounts  []string
	backend   scenarioBackend
	sequence  atomic.Uint64
}

// NewOrderCreateScenario creates a new order creation scenario.
func NewOrderCreateScenario(grpcEndpoint string, accounts []string) *OrderCreateScenario {
	return &OrderCreateScenario{
		transport: scenarioTransport{endpoint: grpcEndpoint},
		accounts:  append([]string(nil), accounts...),
		backend:   newDeterministicBackend(),
	}
}

func (s *OrderCreateScenario) Name() string {
	return "order_create"
}

func (s *OrderCreateScenario) Setup(ctx context.Context) error {
	if len(s.accounts) == 0 {
		return fmt.Errorf("order_create requires at least one account")
	}
	return s.transport.setup(ctx)
}

func (s *OrderCreateScenario) Execute(ctx context.Context) (*framework.ExecutionResult, error) {
	start := time.Now()
	sequence := s.sequence.Add(1)
	owner := s.accounts[(sequence-1)%uint64(len(s.accounts))]
	quantity := int(sequence%4) + 1
	config := map[string]string{
		"region": regionForSequence(sequence),
		"tier":   tierForSequence(sequence),
	}

	receipt, err := s.backend.CreateOrder(ctx, owner, quantity, config)
	if err != nil {
		return &framework.ExecutionResult{
			Success:  false,
			Duration: time.Since(start),
			Error:    err,
		}, nil
	}

	metadata := cloneMetadata(receipt.Metadata)
	metadata["order_id"] = receipt.ID

	return &framework.ExecutionResult{
		Success:  true,
		Duration: time.Since(start),
		Metadata: metadata,
	}, nil
}

func (s *OrderCreateScenario) Teardown(ctx context.Context) error {
	_ = ctx
	return s.transport.teardown()
}

func regionForSequence(sequence uint64) string {
	regions := []string{"ap-southeast", "us-east", "eu-central"}
	return regions[(sequence-1)%uint64(len(regions))]
}

func tierForSequence(sequence uint64) string {
	tiers := []string{"standard", "gpu", "confidential"}
	return tiers[(sequence-1)%uint64(len(tiers))]
}
