// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package provider_daemon

import (
	"context"
	"fmt"
)

// UsageChainRecorder converts raw metering records into canonical reports and
// delegates durability, proof allocation, signing, and broadcast to the chain
// submitter. It intentionally does not auto-settle on final collection;
// customer acknowledgment and settlement policy remain distinct chain steps.
type UsageChainRecorder struct {
	pipeline *SettlementPipeline
}

func NewUsageChainRecorder(submitter ChainUsageSubmitter) (*UsageChainRecorder, error) {
	if submitter == nil {
		return nil, fmt.Errorf("chain usage submitter is required")
	}
	return &UsageChainRecorder{
		pipeline: NewSettlementPipeline(DefaultSettlementConfig(), nil, nil, NewUsageSnapshotStore(), submitter),
	}, nil
}

func (r *UsageChainRecorder) SubmitUsageRecord(ctx context.Context, record *UsageRecord) error {
	if r == nil || r.pipeline == nil {
		return fmt.Errorf("usage chain recorder is not configured")
	}
	return r.pipeline.SubmitUsageToChain(ctx, record)
}

func (r *UsageChainRecorder) SubmitFinalSettlement(ctx context.Context, record *UsageRecord) error {
	return r.SubmitUsageRecord(ctx, record)
}

var _ ChainRecorder = (*UsageChainRecorder)(nil)
