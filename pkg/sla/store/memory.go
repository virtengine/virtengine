/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

package store

import (
	"context"
	"fmt"
	"sync"

	"github.com/virtengine/virtengine/pkg/sla"
)

// MemoryStore provides an in-memory SLA store for development and tests.
type MemoryStore struct {
	mu          sync.RWMutex
	leases      []sla.LeaseInfo
	definitions map[string]*sla.SLADefinition
	breaches    map[string]*sla.BreachRecord
}

// NewMemoryStore creates a new in-memory SLA store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		definitions: make(map[string]*sla.SLADefinition),
		breaches:    make(map[string]*sla.BreachRecord),
	}
}

// SetActiveLeases sets leases for monitoring.
func (s *MemoryStore) SetActiveLeases(leases []sla.LeaseInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.leases = append([]sla.LeaseInfo{}, leases...)
}

// SetSLADefinition associates a provider with an SLA definition.
func (s *MemoryStore) SetSLADefinition(providerID string, definition *sla.SLADefinition) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.definitions[providerID] = definition
}

// GetActiveLeases returns configured leases.
func (s *MemoryStore) GetActiveLeases(_ context.Context) ([]sla.LeaseInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]sla.LeaseInfo{}, s.leases...), nil
}

// GetSLADefinition returns the provider SLA definition.
func (s *MemoryStore) GetSLADefinition(_ context.Context, providerID string) (*sla.SLADefinition, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	def, ok := s.definitions[providerID]
	if !ok {
		return nil, fmt.Errorf("sla definition not found for provider %s", providerID)
	}
	return def, nil
}

// CreateBreach stores an active breach.
func (s *MemoryStore) CreateBreach(_ context.Context, breach *sla.BreachRecord) error {
	if breach == nil {
		return fmt.Errorf("breach is nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.breaches[breachKey(breach.LeaseID, breach.MetricType)] = breach
	return nil
}

// UpdateBreach updates a breach record.
func (s *MemoryStore) UpdateBreach(_ context.Context, breach *sla.BreachRecord) error {
	if breach == nil {
		return fmt.Errorf("breach is nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.breaches[breachKey(breach.LeaseID, breach.MetricType)] = breach
	return nil
}

// GetActiveBreach returns an unresolved breach for a lease/metric.
func (s *MemoryStore) GetActiveBreach(_ context.Context, leaseID string, metricType sla.SLAMetricType) (*sla.BreachRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	breach, ok := s.breaches[breachKey(leaseID, metricType)]
	if !ok {
		return nil, nil
	}
	if breach.ResolvedAt != nil {
		return nil, nil
	}

	return breach, nil
}

func breachKey(leaseID string, metricType sla.SLAMetricType) string {
	return fmt.Sprintf("%s:%s", leaseID, metricType)
}
