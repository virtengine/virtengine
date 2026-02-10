/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

package sla

import "context"

// Remediator handles SLA breach remediation.
type Remediator interface {
	HandleBreach(ctx context.Context, breach *BreachRecord)
	HandleWarning(ctx context.Context, status *SLAStatus, metricType SLAMetricType)
	ApplyCredit(ctx context.Context, breach *BreachRecord)
}

// NoopRemediator is a safe default implementation.
type NoopRemediator struct{}

// HandleBreach implements Remediator.
func (r NoopRemediator) HandleBreach(_ context.Context, _ *BreachRecord) {}

// HandleWarning implements Remediator.
func (r NoopRemediator) HandleWarning(_ context.Context, _ *SLAStatus, _ SLAMetricType) {}

// ApplyCredit implements Remediator.
func (r NoopRemediator) ApplyCredit(_ context.Context, _ *BreachRecord) {}
