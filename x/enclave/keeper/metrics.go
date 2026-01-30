package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/virtengine/virtengine/x/enclave/types"
)

var (
	// EnclaveIdentitiesTotal tracks total number of registered enclave identities
	EnclaveIdentitiesTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "virtengine",
			Subsystem: "enclave",
			Name:      "identities_total",
			Help:      "Total number of registered enclave identities",
		},
		[]string{"status"},
	)

	// EnclaveIdentityRegistrations tracks cumulative identity registrations
	EnclaveIdentityRegistrations = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: "virtengine",
			Subsystem: "enclave",
			Name:      "identity_registrations_total",
			Help:      "Total number of enclave identity registrations",
		},
	)

	// EnclaveKeyRotations tracks active key rotations
	EnclaveKeyRotationsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "virtengine",
			Subsystem: "enclave",
			Name:      "key_rotations_active",
			Help:      "Number of active enclave key rotations",
		},
	)

	// EnclaveKeyRotationsCompleted tracks completed key rotations
	EnclaveKeyRotationsCompleted = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: "virtengine",
			Subsystem: "enclave",
			Name:      "key_rotations_completed_total",
			Help:      "Total number of completed enclave key rotations",
		},
	)

	// EnclaveMeasurementsAllowlisted tracks allowlisted measurements
	EnclaveMeasurementsAllowlisted = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "virtengine",
			Subsystem: "enclave",
			Name:      "measurements_allowlisted",
			Help:      "Number of measurements in the allowlist",
		},
		[]string{"tee_type"},
	)

	// EnclaveMeasurementProposals tracks measurement proposals
	EnclaveMeasurementProposals = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "virtengine",
			Subsystem: "enclave",
			Name:      "measurement_proposals_total",
			Help:      "Total number of measurement proposals",
		},
		[]string{"action"}, // "add" or "revoke"
	)

	// EnclaveAttestationVerifications tracks attestation verifications
	EnclaveAttestationVerifications = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "virtengine",
			Subsystem: "enclave",
			Name:      "attestation_verifications_total",
			Help:      "Total number of attestation verifications",
		},
		[]string{"result"}, // "success" or "failure"
	)

	// EnclaveCommitteeSize tracks committee size
	EnclaveCommitteeSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "virtengine",
			Subsystem: "enclave",
			Name:      "committee_size",
			Help:      "Current size of the enclave identity committee",
		},
	)

	// EnclaveAttestationAge tracks age of attestations
	EnclaveAttestationAge = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "virtengine",
			Subsystem: "enclave",
			Name:      "attestation_age_blocks",
			Help:      "Age of attestations in blocks at registration time",
			Buckets:   prometheus.LinearBuckets(0, 1000, 20), // 0-20k blocks
		},
	)

	// EnclaveIdentityExpiries tracks upcoming expiries
	EnclaveIdentityExpiries = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "virtengine",
			Subsystem: "enclave",
			Name:      "identity_expiries_upcoming",
			Help:      "Number of identities expiring in the next N blocks",
		},
		[]string{"time_range"}, // "1h", "24h", "7d"
	)
)

// RecordMetrics updates all enclave module metrics
// This should be called periodically (e.g., in BeginBlocker)
func (k Keeper) RecordMetrics(ctx sdk.Context) {
	// Count identities by status
	statusCounts := make(map[string]float64)
	k.WithEnclaveIdentities(ctx, func(identity types.EnclaveIdentity) bool {
		statusCounts[identity.Status.String()]++
		return false
	})

	// Update identity count metrics
	for status, count := range statusCounts {
		EnclaveIdentitiesTotal.WithLabelValues(status).Set(count)
	}

	// Count active key rotations
	activeRotations := 0
	k.WithEnclaveIdentities(ctx, func(identity types.EnclaveIdentity) bool {
		if identity.Status == types.EnclaveIdentityStatusRotating {
			activeRotations++
		}
		return false
	})
	EnclaveKeyRotationsActive.Set(float64(activeRotations))

	// Count measurements by TEE type
	measurementCounts := make(map[string]float64)
	k.WithMeasurements(ctx, func(measurement types.MeasurementRecord) bool {
		if !measurement.Revoked {
			measurementCounts[measurement.TEEType.String()]++
		}
		return false
	})

	for teeType, count := range measurementCounts {
		EnclaveMeasurementsAllowlisted.WithLabelValues(teeType).Set(count)
	}

	// Update committee size
	params := k.GetParams(ctx)
	if params.EnableCommitteeMode {
		EnclaveCommitteeSize.Set(float64(params.CommitteeSize))
	} else {
		EnclaveCommitteeSize.Set(0)
	}

	// Track upcoming expiries
	k.recordUpcomingExpiries(ctx)
}

// recordUpcomingExpiries tracks identities expiring soon
func (k Keeper) recordUpcomingExpiries(ctx sdk.Context) {
	currentHeight := ctx.BlockHeight()

	// Define time ranges (assuming ~5s block time)
	timeRanges := map[string]int64{
		"1h":  720,   // 1 hour = 720 blocks
		"24h": 17280, // 24 hours = 17,280 blocks
		"7d":  120960, // 7 days = 120,960 blocks
	}

	expiryCounts := make(map[string]float64)

	k.WithEnclaveIdentities(ctx, func(identity types.EnclaveIdentity) bool {
		if identity.Status == types.EnclaveIdentityStatusExpired ||
			identity.Status == types.EnclaveIdentityStatusRevoked {
			return false
		}

		blocksUntilExpiry := identity.ExpiryHeight - currentHeight
		if blocksUntilExpiry < 0 {
			return false
		}

		for label, threshold := range timeRanges {
			if blocksUntilExpiry <= threshold {
				expiryCounts[label]++
			}
		}

		return false
	})

	for label, count := range expiryCounts {
		EnclaveIdentityExpiries.WithLabelValues(label).Set(count)
	}
}

// RecordIdentityRegistration records a new identity registration
func (k Keeper) RecordIdentityRegistration() {
	EnclaveIdentityRegistrations.Inc()
}

// RecordKeyRotationCompleted records a completed key rotation
func (k Keeper) RecordKeyRotationCompleted() {
	EnclaveKeyRotationsCompleted.Inc()
}

// RecordMeasurementProposal records a measurement proposal
func (k Keeper) RecordMeasurementProposal(action string) {
	EnclaveMeasurementProposals.WithLabelValues(action).Inc()
}

// RecordAttestationVerification records an attestation verification result
func (k Keeper) RecordAttestationVerification(success bool) {
	result := "success"
	if !success {
		result = "failure"
	}
	EnclaveAttestationVerifications.WithLabelValues(result).Inc()
}

// RecordAttestationAge records the age of an attestation
func (k Keeper) RecordAttestationAge(ageInBlocks int64) {
	EnclaveAttestationAge.Observe(float64(ageInBlocks))
}
