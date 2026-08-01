// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package reviewops

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func TestOperationalMetricsDefaultIsUnavailableAndOmitsValues(t *testing.T) {
	registry := prometheus.NewRegistry()
	_, err := NewOperationalMetrics(registry)
	require.NoError(t, err)
	families, err := registry.Gather()
	require.NoError(t, err)

	requireMetricSeriesCount(t, families, "virtengine_review_operations_queue_source_available", 2)
	requireMetricSeriesCount(t, families, "virtengine_review_operations_control_source_available", 8)
	requireMetricSeriesCount(t, families, "virtengine_review_operations_subgroup_source_available", 1)
	requireMetricAbsent(t, families, "virtengine_review_operations_queue_pending")
	requireMetricAbsent(t, families, "virtengine_review_operations_queue_oldest_age_seconds")
	requireMetricAbsent(t, families, "virtengine_review_operations_control_observations")
	requireMetricAbsent(t, families, "virtengine_review_operations_subgroup_overturn_observations")
}

func TestOperationalMetricsBoundedContract(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := newTestMetrics(t, registry)
	snapshot := validSnapshot()
	require.NoError(t, metrics.SetSnapshot(snapshot))

	families, err := registry.Gather()
	require.NoError(t, err)
	requireMetricLabels(t, families, "virtengine_review_operations_queue_source_available", []string{"queue"})
	requireMetricLabels(t, families, "virtengine_review_operations_queue_pending", []string{"queue"})
	requireMetricLabels(t, families, "virtengine_review_operations_queue_oldest_age_seconds", []string{"queue"})
	requireMetricLabels(t, families, "virtengine_review_operations_control_source_available", []string{"control", "queue"})
	requireMetricLabels(t, families, "virtengine_review_operations_control_observations", []string{"control", "queue", "result"})
	requireMetricLabels(t, families, "virtengine_review_operations_subgroup_overturn_observations", []string{"result", "subgroup"})
	requireMetricSeriesCount(t, families, "virtengine_review_operations_queue_source_available", 2)
	requireMetricSeriesCount(t, families, "virtengine_review_operations_queue_pending", 2)
	requireMetricSeriesCount(t, families, "virtengine_review_operations_queue_oldest_age_seconds", 2)
	requireMetricSeriesCount(t, families, "virtengine_review_operations_control_source_available", 8)
	requireMetricSeriesCount(t, families, "virtengine_review_operations_control_observations", 24)
	requireMetricSeriesCount(t, families, "virtengine_review_operations_subgroup_overturn_observations", 4)
	require.Equal(t, float64(30), metricValue(t, families, "virtengine_review_operations_queue_pending", map[string]string{"queue": "appeal"}))
	require.Equal(t, float64(7_200), metricValue(t, families, "virtengine_review_operations_queue_oldest_age_seconds", map[string]string{"queue": "appeal"}))
	require.Equal(t, float64(2), metricValue(t, families, "virtengine_review_operations_subgroups_suppressed", nil))
}

func TestOperationalMetricsRejectsContradictionsBeforeMutationAndDoesNotLeak(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := newTestMetrics(t, registry)
	require.NoError(t, metrics.SetSnapshot(validSnapshot()))
	before := gatherText(t, registry)

	canary := "alice-appeal-biometric-nullifier-reviewer"
	invalid := validSnapshot()
	invalid.Queues[0].Queue = QueueKind(canary)
	err := metrics.SetSnapshot(invalid)
	require.EqualError(t, err, "review queue snapshot order is invalid")
	require.NotContains(t, err.Error(), canary)
	require.Equal(t, before, gatherText(t, registry))

	invalid = validSnapshot()
	invalid.Controls[0].Satisfied = invalid.Controls[0].Eligible + 1
	require.EqualError(t, metrics.SetSnapshot(invalid), "review control counts are invalid")
	require.Equal(t, before, gatherText(t, registry))

	invalid = validSnapshot()
	invalid.Queues[0] = QueueAggregate{Queue: QueueAppeal, Pending: 1}
	require.EqualError(t, metrics.SetSnapshot(invalid), "unavailable review queue must omit operational values")
	require.Equal(t, before, gatherText(t, registry))
}

func TestOperationalMetricsSubgroupPrivacyAndAvailability(t *testing.T) {
	metrics := newTestMetrics(t, prometheus.NewRegistry())
	for _, mutate := range []func(*Snapshot){
		func(snapshot *Snapshot) { snapshot.Subgroups[0].Reviewed = MinimumSubgroupCell - 1 },
		func(snapshot *Snapshot) { snapshot.Subgroups[0].Overturned = snapshot.Subgroups[0].Reviewed + 1 },
		func(snapshot *Snapshot) { snapshot.Subgroups[1].Slot = snapshot.Subgroups[0].Slot },
		func(snapshot *Snapshot) { snapshot.Subgroups[0].Slot = MaximumSubgroups + 1 },
		func(snapshot *Snapshot) { snapshot.SubgroupTaxonomyDigest = "private-taxonomy-name" },
		func(snapshot *Snapshot) {
			snapshot.SuppressedSubgroupCount = MaximumSubgroups - uint64(len(snapshot.Subgroups)) + 1
		},
	} {
		snapshot := validSnapshot()
		mutate(&snapshot)
		require.Error(t, metrics.SetSnapshot(snapshot))
	}

	snapshot := validSnapshot()
	snapshot.SubgroupSourceAvailable = false
	snapshot.SubgroupTaxonomyDigest = ""
	snapshot.Subgroups = nil
	snapshot.SuppressedSubgroupCount = 0
	require.NoError(t, metrics.SetSnapshot(snapshot))
}

func TestOperationalMetricsDuplicateRegistrationSharesSnapshot(t *testing.T) {
	registry := prometheus.NewRegistry()
	first, err := NewOperationalMetrics(registry)
	require.NoError(t, err)
	first.collector.now = func() time.Time { return fixtureTime }
	second, err := NewOperationalMetrics(registry)
	require.NoError(t, err)
	require.Same(t, first.collector, second.collector)

	snapshot := validSnapshot()
	snapshot.Queues[0].Pending = 41
	require.NoError(t, second.SetSnapshot(snapshot))
	families, err := registry.Gather()
	require.NoError(t, err)
	require.Equal(t, float64(41), metricValue(t, families, "virtengine_review_operations_queue_pending", map[string]string{"queue": "appeal"}))
}

func TestOperationalMetricsRejectsConflictingRegistration(t *testing.T) {
	registry := prometheus.NewRegistry()
	require.NoError(t, registry.Register(prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "virtengine_review_operations_queue_source_available",
		Help: "Conflicting review queue availability.",
	}, []string{"queue"})))
	_, err := NewOperationalMetrics(registry)
	require.Error(t, err)
}

func TestOperationalMetricsClonesInputAndSupportsConcurrentGather(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := newTestMetrics(t, registry)
	snapshot := validSnapshot()
	require.NoError(t, metrics.SetSnapshot(snapshot))
	snapshot.Queues[0].Pending = 999
	families, err := registry.Gather()
	require.NoError(t, err)
	require.Equal(t, float64(30), metricValue(t, families, "virtengine_review_operations_queue_pending", map[string]string{"queue": "appeal"}))

	var wait sync.WaitGroup
	errors := make(chan error, 64)
	for index := 0; index < 32; index++ {
		wait.Add(2)
		go func() {
			defer wait.Done()
			errors <- metrics.SetSnapshot(validSnapshot())
		}()
		go func() {
			defer wait.Done()
			_, err := registry.Gather()
			errors <- err
		}()
	}
	wait.Wait()
	close(errors)
	for err := range errors {
		require.NoError(t, err)
	}
}

func TestOperationalMetricsFailsClosedWhenSnapshotBecomesStale(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := newTestMetrics(t, registry)
	require.NoError(t, metrics.SetSnapshot(validSnapshot()))

	metrics.collector.now = func() time.Time { return fixtureTime.Add(2 * time.Minute) }
	families, err := registry.Gather()
	require.NoError(t, err)
	require.Equal(t, float64(7_320), metricValue(t, families, "virtengine_review_operations_queue_oldest_age_seconds", map[string]string{"queue": "appeal"}))

	metrics.collector.now = func() time.Time { return fixtureTime.Add(SnapshotFreshnessLimit + time.Second) }
	families, err = registry.Gather()
	require.NoError(t, err)
	requireMetricAbsent(t, families, "virtengine_review_operations_queue_pending")
	requireMetricAbsent(t, families, "virtengine_review_operations_control_observations")
	requireMetricAbsent(t, families, "virtengine_review_operations_subgroup_overturn_observations")
	require.Equal(t, float64(0), metricValue(t, families, "virtengine_review_operations_queue_source_available", map[string]string{"queue": "appeal"}))
	require.Equal(t, float64(0), metricValue(t, families, "virtengine_review_operations_subgroup_source_available", nil))
}

func TestOperationalMetricsEmptyQueueAgeRemainsZero(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := newTestMetrics(t, registry)
	snapshot := validSnapshot()
	snapshot.Queues[0].Pending = 0
	snapshot.Queues[0].OldestAge = 0
	require.NoError(t, metrics.SetSnapshot(snapshot))

	metrics.collector.now = func() time.Time { return fixtureTime.Add(2 * time.Minute) }
	families, err := registry.Gather()
	require.NoError(t, err)
	require.Equal(t, float64(0), metricValue(t, families, "virtengine_review_operations_queue_oldest_age_seconds", map[string]string{"queue": "appeal"}))
}

func TestOperationalMetricsTaxonomyDigestIsImmutable(t *testing.T) {
	metrics := newTestMetrics(t, prometheus.NewRegistry())
	require.NoError(t, metrics.SetSnapshot(validSnapshot()))
	metrics.collector.now = func() time.Time { return fixtureTime.Add(time.Second) }
	candidate := validSnapshot()
	candidate.ObservedAt = fixtureTime.Add(time.Second)
	candidate.SubgroupTaxonomyDigest = testDigest("other-taxonomy")
	require.EqualError(t, metrics.SetSnapshot(candidate), "subgroup taxonomy digest is immutable for collector lifetime")
}

func TestOperationalMetricsRejectsRegressingAndConflictingSnapshots(t *testing.T) {
	metrics := newTestMetrics(t, prometheus.NewRegistry())
	snapshot := validSnapshot()
	require.NoError(t, metrics.SetSnapshot(snapshot))
	require.NoError(t, metrics.SetSnapshot(snapshot))

	regressing := validSnapshot()
	regressing.ObservedAt = fixtureTime.Add(-time.Minute)
	require.EqualError(t, metrics.SetSnapshot(regressing), "review operations snapshot cannot regress")

	conflicting := validSnapshot()
	conflicting.Queues[0].Pending++
	require.EqualError(t, metrics.SetSnapshot(conflicting), "review operations snapshot conflicts at observation time")
}

func validSnapshot() Snapshot {
	controls := make([]ControlAggregate, 0, len(queueKinds)*len(controlKinds))
	for _, queue := range queueKinds {
		for _, control := range controlKinds {
			controls = append(controls, ControlAggregate{
				Queue: queue, Control: control, Available: true,
				Eligible: 100, Satisfied: 80, Unavailable: 10,
			})
		}
	}
	return Snapshot{
		ObservedAt: fixtureTime,
		Queues: []QueueAggregate{
			{Queue: QueueAppeal, Available: true, Pending: 30, OldestAge: 2 * time.Hour},
			{Queue: QueueManualReview, Available: true, Pending: 20, OldestAge: time.Hour},
		},
		Controls: controls, SubgroupSourceAvailable: true,
		SubgroupTaxonomyDigest: testDigest("governed-taxonomy"),
		Subgroups: []SubgroupOverturnAggregate{
			{Slot: 1, Reviewed: 30, Overturned: 3},
			{Slot: 2, Reviewed: 40, Overturned: 5},
		},
		SuppressedSubgroupCount: 2,
	}
}

var fixtureTime = time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)

func newTestMetrics(t *testing.T, registry *prometheus.Registry) *OperationalMetrics {
	t.Helper()
	metrics, err := NewOperationalMetrics(registry)
	require.NoError(t, err)
	metrics.collector.now = func() time.Time { return fixtureTime }
	return metrics
}

func testDigest(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}

func gatherText(t *testing.T, registry *prometheus.Registry) string {
	t.Helper()
	families, err := registry.Gather()
	require.NoError(t, err)
	var buffer bytes.Buffer
	for _, family := range families {
		buffer.WriteString(family.String())
	}
	return buffer.String()
}

func requireMetricAbsent(t *testing.T, families []*dto.MetricFamily, name string) {
	t.Helper()
	for _, family := range families {
		require.NotEqual(t, name, family.GetName())
	}
}

func requireMetricSeriesCount(t *testing.T, families []*dto.MetricFamily, name string, expected int) {
	t.Helper()
	for _, family := range families {
		if family.GetName() == name {
			require.Len(t, family.Metric, expected)
			return
		}
	}
	t.Fatalf("metric family %s not found", name)
}

func requireMetricLabels(t *testing.T, families []*dto.MetricFamily, name string, expected []string) {
	t.Helper()
	for _, family := range families {
		if family.GetName() != name {
			continue
		}
		require.NotEmpty(t, family.Metric)
		actual := make([]string, 0, len(family.Metric[0].Label))
		for _, label := range family.Metric[0].Label {
			actual = append(actual, label.GetName())
		}
		require.Equal(t, expected, actual)
		return
	}
	t.Fatalf("metric family %s not found", name)
}

func metricValue(t *testing.T, families []*dto.MetricFamily, name string, labels map[string]string) float64 {
	t.Helper()
	for _, family := range families {
		if family.GetName() != name {
			continue
		}
		for _, metric := range family.Metric {
			matched := true
			for key, expected := range labels {
				found := false
				for _, label := range metric.Label {
					if label.GetName() == key && label.GetValue() == expected {
						found = true
						break
					}
				}
				if !found {
					matched = false
					break
				}
			}
			if matched {
				return metric.GetGauge().GetValue()
			}
		}
	}
	t.Fatalf("metric %s with labels %v not found", name, labels)
	return 0
}
