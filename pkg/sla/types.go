/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

package sla

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// SLAMetricType defines the supported SLA metrics.
type SLAMetricType string

const (
	MetricUptime     SLAMetricType = "uptime"
	MetricLatencyP50 SLAMetricType = "latency_p50"
	MetricLatencyP99 SLAMetricType = "latency_p99"
	MetricThroughput SLAMetricType = "throughput"
	MetricErrorRate  SLAMetricType = "error_rate"
)

// SLATier defines standard SLA tiers.
type SLATier string

const (
	TierBronze   SLATier = "bronze"
	TierSilver   SLATier = "silver"
	TierGold     SLATier = "gold"
	TierPlatinum SLATier = "platinum"
)

// ComparisonType describes how SLA targets are evaluated.
type ComparisonType string

const (
	ComparisonGTE ComparisonType = "gte" // >= (for uptime/throughput)
	ComparisonLTE ComparisonType = "lte" // <= (for latency/error rate)
)

// SLADefinition describes the SLA commitment for a provider.
type SLADefinition struct {
	ID            string
	ProviderID    string
	Tier          SLATier
	Metrics       []SLAMetric
	EffectiveFrom time.Time
	EffectiveTo   *time.Time

	CreditPolicy CreditPolicy
}

// SLAMetric is a single SLA metric target and evaluation window.
type SLAMetric struct {
	Type       SLAMetricType
	Target     decimal.Decimal
	Window     time.Duration
	Comparison ComparisonType
}

// CreditPolicy defines breach compensation rules.
type CreditPolicy struct {
	Brackets  []CreditBracket
	MaxCredit decimal.Decimal // Maximum credit percentage
}

// CreditBracket assigns a credit percentage based on deviation.
type CreditBracket struct {
	MinDeviation  decimal.Decimal
	MaxDeviation  decimal.Decimal
	CreditPercent decimal.Decimal
}

// LeaseInfo identifies the lease under SLA monitoring.
type LeaseInfo struct {
	ID         string
	ProviderID string
	CustomerID string
}

// MetricDataPoint is a historical metric sample.
type MetricDataPoint struct {
	Timestamp time.Time
	Value     decimal.Decimal
}

// SLAStatus captures the current SLA health state for a lease.
type SLAStatus struct {
	LeaseID       string
	ProviderID    string
	CustomerID    string
	SLADefinition *SLADefinition

	Metrics       map[SLAMetricType]*MetricStatus
	OverallStatus HealthStatus

	ActiveBreach  *BreachRecord
	BreachHistory []BreachRecord

	LastChecked time.Time
}

// MetricStatus captures status for a single metric.
type MetricStatus struct {
	Type         SLAMetricType
	CurrentValue decimal.Decimal
	Target       decimal.Decimal
	Status       HealthStatus
	LastUpdated  time.Time
}

// HealthStatus represents SLA health categories.
type HealthStatus string

const (
	StatusHealthy  HealthStatus = "healthy"
	StatusWarning  HealthStatus = "warning"
	StatusBreached HealthStatus = "breached"
)

// BreachRecord tracks SLA breaches for a lease.
type BreachRecord struct {
	ID           string
	LeaseID      string
	MetricType   SLAMetricType
	TargetValue  decimal.Decimal
	ActualValue  decimal.Decimal
	StartedAt    time.Time
	ResolvedAt   *time.Time
	Duration     time.Duration
	CreditAmount decimal.Decimal
}

// MetricsCollector supplies aggregated metric values.
type MetricsCollector interface {
	GetMetric(ctx context.Context, leaseID string, metric SLAMetricType, window time.Duration) (decimal.Decimal, error)
	GetHistoricalMetrics(ctx context.Context, leaseID string, metric SLAMetricType, start, end time.Time) ([]MetricDataPoint, error)
}

// SLAStore provides SLA definitions and breach persistence.
type SLAStore interface {
	GetActiveLeases(ctx context.Context) ([]LeaseInfo, error)
	GetSLADefinition(ctx context.Context, providerID string) (*SLADefinition, error)
	CreateBreach(ctx context.Context, breach *BreachRecord) error
	UpdateBreach(ctx context.Context, breach *BreachRecord) error
	GetActiveBreach(ctx context.Context, leaseID string, metricType SLAMetricType) (*BreachRecord, error)
}

// DefaultSLATiers provides baseline SLA tiers.
var DefaultSLATiers = map[SLATier][]SLAMetric{
	TierBronze: {
		{Type: MetricUptime, Target: decimal.NewFromFloat(99.0), Window: 30 * 24 * time.Hour, Comparison: ComparisonGTE},
		{Type: MetricLatencyP99, Target: decimal.NewFromInt(1000), Window: 24 * time.Hour, Comparison: ComparisonLTE},
	},
	TierSilver: {
		{Type: MetricUptime, Target: decimal.NewFromFloat(99.5), Window: 30 * 24 * time.Hour, Comparison: ComparisonGTE},
		{Type: MetricLatencyP99, Target: decimal.NewFromInt(500), Window: 24 * time.Hour, Comparison: ComparisonLTE},
		{Type: MetricErrorRate, Target: decimal.NewFromFloat(1.0), Window: 24 * time.Hour, Comparison: ComparisonLTE},
	},
	TierGold: {
		{Type: MetricUptime, Target: decimal.NewFromFloat(99.9), Window: 30 * 24 * time.Hour, Comparison: ComparisonGTE},
		{Type: MetricLatencyP99, Target: decimal.NewFromInt(200), Window: 24 * time.Hour, Comparison: ComparisonLTE},
		{Type: MetricErrorRate, Target: decimal.NewFromFloat(0.1), Window: 24 * time.Hour, Comparison: ComparisonLTE},
	},
	TierPlatinum: {
		{Type: MetricUptime, Target: decimal.NewFromFloat(99.99), Window: 30 * 24 * time.Hour, Comparison: ComparisonGTE},
		{Type: MetricLatencyP99, Target: decimal.NewFromInt(100), Window: 24 * time.Hour, Comparison: ComparisonLTE},
		{Type: MetricErrorRate, Target: decimal.NewFromFloat(0.01), Window: 24 * time.Hour, Comparison: ComparisonLTE},
	},
}
