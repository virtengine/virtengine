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

// BidSubmitScenario executes deterministic provider bid probes against seeded orders.
type BidSubmitScenario struct {
	transport scenarioTransport
	accounts  []string
	backend   scenarioBackend
	orderIDs  []string
	sequence  atomic.Uint64
}

// NewBidSubmitScenario creates a new bid submission scenario.
func NewBidSubmitScenario(grpcEndpoint string, accounts []string) *BidSubmitScenario {
	return &BidSubmitScenario{
		transport: scenarioTransport{endpoint: grpcEndpoint},
		accounts:  append([]string(nil), accounts...),
		backend:   newDeterministicBackend(),
	}
}

func (s *BidSubmitScenario) Name() string {
	return "bid_submit"
}

func (s *BidSubmitScenario) Setup(ctx context.Context) error {
	if len(s.accounts) < 2 {
		return fmt.Errorf("bid_submit requires at least two accounts")
	}
	if err := s.transport.setup(ctx); err != nil {
		return err
	}

	s.orderIDs = s.orderIDs[:0]
	for idx, owner := range s.accounts {
		receipt, err := s.backend.CreateOrder(ctx, owner, idx%3+1, map[string]string{
			"region": regionForSequence(uint64(idx + 1)),
			"tier":   tierForSequence(uint64(idx + 1)),
		})
		if err != nil {
			return fmt.Errorf("seed order %d: %w", idx, err)
		}
		s.orderIDs = append(s.orderIDs, receipt.ID)
	}

	return nil
}

func (s *BidSubmitScenario) Execute(ctx context.Context) (*framework.ExecutionResult, error) {
	start := time.Now()
	sequence := s.sequence.Add(1)
	orderID := s.orderIDs[(sequence-1)%uint64(len(s.orderIDs))]
	provider := s.accounts[sequence%uint64(len(s.accounts))]
	price := 90 + int64(sequence%25)

	receipt, err := s.backend.SubmitBid(ctx, orderID, provider, price)
	if err != nil {
		return &framework.ExecutionResult{
			Success:  false,
			Duration: time.Since(start),
			Error:    err,
		}, nil
	}

	metadata := cloneMetadata(receipt.Metadata)
	metadata["bid_id"] = receipt.ID

	return &framework.ExecutionResult{
		Success:  true,
		Duration: time.Since(start),
		Metadata: metadata,
	}, nil
}

func (s *BidSubmitScenario) Teardown(ctx context.Context) error {
	_ = ctx
	return s.transport.teardown()
}
