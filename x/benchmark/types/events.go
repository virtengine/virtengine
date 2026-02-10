// Package types contains types for the Benchmark module.
//
// VE-601: Events for the benchmark module
package types

import (
	benchmarkv1 "github.com/virtengine/virtengine/sdk/go/node/benchmark/v1"
)

// Event type aliases to generated protobuf types
type (
	AnomalyDetectedEvent         = benchmarkv1.AnomalyDetectedEvent
	AnomalyResolvedEvent         = benchmarkv1.AnomalyResolvedEvent
	ProviderFlaggedEvent         = benchmarkv1.ProviderFlaggedEvent
	ProviderUnflaggedEvent       = benchmarkv1.ProviderUnflaggedEvent
	ChallengeRequestedEvent      = benchmarkv1.ChallengeRequestedEvent
	ChallengeCompletedEvent      = benchmarkv1.ChallengeCompletedEvent
	ChallengeExpiredEvent        = benchmarkv1.ChallengeExpiredEvent
	BenchmarksSubmittedEvent     = benchmarkv1.BenchmarksSubmittedEvent
	BenchmarksPrunedEvent        = benchmarkv1.BenchmarksPrunedEvent
	ReliabilityScoreUpdatedEvent = benchmarkv1.ReliabilityScoreUpdatedEvent
)

// Event types for the benchmark module
const (
	// EventTypeBenchmarksSubmitted is emitted when benchmarks are submitted
	EventTypeBenchmarksSubmitted = "benchmarks_submitted"

	// EventTypeBenchmarksRejected is emitted when benchmarks are rejected
	EventTypeBenchmarksRejected = "benchmarks_rejected"

	// EventTypeBenchmarksPruned is emitted when old benchmarks are pruned
	EventTypeBenchmarksPruned = "benchmarks_pruned"

	// EventTypeReliabilityScoreUpdated is emitted when a reliability score is updated
	EventTypeReliabilityScoreUpdated = "reliability_score_updated"

	// EventTypeChallengeRequested is emitted when a challenge is requested
	EventTypeChallengeRequested = "benchmark_challenge_requested"

	// EventTypeChallengeCompleted is emitted when a challenge is completed
	EventTypeChallengeCompleted = "benchmark_challenge_completed"

	// EventTypeChallengeExpired is emitted when a challenge expires
	EventTypeChallengeExpired = "benchmark_challenge_expired"

	// EventTypeAnomalyDetected is emitted when an anomaly is detected
	EventTypeAnomalyDetected = "benchmark_anomaly_detected"

	// EventTypeAnomalyResolved is emitted when an anomaly is resolved
	EventTypeAnomalyResolved = "benchmark_anomaly_resolved"

	// EventTypeProviderFlagged is emitted when a provider is flagged
	EventTypeProviderFlagged = "provider_flagged"

	// EventTypeProviderUnflagged is emitted when a provider is unflagged
	EventTypeProviderUnflagged = "provider_unflagged"
)

// Event attribute keys
const (
	AttributeKeyProviderAddress = "provider_address"
	AttributeKeyClusterID       = "cluster_id"
	AttributeKeyReportID        = "report_id"
	AttributeKeyReportCount     = "report_count"
	AttributeKeyChallengeID     = "challenge_id"
	AttributeKeyAnomalyFlagID   = "anomaly_flag_id"
	AttributeKeyAnomalyType     = "anomaly_type"
	AttributeKeySeverity        = "severity"
	AttributeKeyScore           = "score"
	AttributeKeyScoreVersion    = "score_version"
	AttributeKeyReason          = "reason"
	AttributeKeyModerator       = "moderator"
	AttributeKeyPrunedCount     = "pruned_count"
)
