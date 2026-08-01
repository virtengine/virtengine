// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package reviewops

import (
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	MinimumSubgroupCell    uint64 = 20
	MaximumSubgroups              = 64
	MaximumCount           uint64 = 1_000_000_000_000
	SnapshotFreshnessLimit        = 5 * time.Minute
	maximumOldestAge              = time.Duration(1<<63-1) - SnapshotFreshnessLimit
)

type QueueKind string
type ControlKind string

type QueueAggregate struct {
	Queue     QueueKind
	Available bool
	Pending   uint64
	OldestAge time.Duration
}

type ControlAggregate struct {
	Queue       QueueKind
	Control     ControlKind
	Available   bool
	Eligible    uint64
	Satisfied   uint64
	Unavailable uint64
}

type SubgroupOverturnAggregate struct {
	Slot       uint8
	Reviewed   uint64
	Overturned uint64
}

type Snapshot struct {
	ObservedAt              time.Time
	Queues                  []QueueAggregate
	Controls                []ControlAggregate
	SubgroupSourceAvailable bool
	SubgroupTaxonomyDigest  string
	Subgroups               []SubgroupOverturnAggregate
	SuppressedSubgroupCount uint64
}

const (
	QueueAppeal       QueueKind = "appeal"
	QueueManualReview QueueKind = "manual_review"

	ControlIndependentReviewer ControlKind = "independent_reviewer"
	ControlRecusal             ControlKind = "recusal"
	ControlReasonDelivery      ControlKind = "reason_delivery"
	ControlRestoration         ControlKind = "restoration"
)

var (
	queueKinds   = [...]QueueKind{QueueAppeal, QueueManualReview}
	controlKinds = [...]ControlKind{
		ControlIndependentReviewer,
		ControlRecusal,
		ControlReasonDelivery,
		ControlRestoration,
	}
)

type OperationalMetrics struct {
	collector *collector
}

func NewOperationalMetrics(registerer prometheus.Registerer) (*OperationalMetrics, error) {
	if registerer == nil {
		return nil, errors.New("review operations metrics registerer is required")
	}
	candidate := newCollector()
	if err := registerer.Register(candidate); err != nil {
		already, ok := err.(prometheus.AlreadyRegisteredError)
		if !ok {
			return nil, err
		}
		existing, ok := already.ExistingCollector.(*collector)
		if !ok {
			return nil, errors.New("registered review operations collector has unexpected type")
		}
		candidate = existing
	}
	return &OperationalMetrics{collector: candidate}, nil
}

func (metrics *OperationalMetrics) SetSnapshot(snapshot Snapshot) error {
	if metrics == nil || metrics.collector == nil {
		return errors.New("review operations metrics are required")
	}
	candidate := cloneSnapshot(snapshot)
	metrics.collector.mu.Lock()
	defer metrics.collector.mu.Unlock()
	if err := candidate.validate(metrics.collector.now()); err != nil {
		return err
	}
	if metrics.collector.initialized {
		if candidate.ObservedAt.Before(metrics.collector.snapshot.ObservedAt) {
			return errors.New("review operations snapshot cannot regress")
		}
		if candidate.ObservedAt.Equal(metrics.collector.snapshot.ObservedAt) {
			if !reflect.DeepEqual(candidate, metrics.collector.snapshot) {
				return errors.New("review operations snapshot conflicts at observation time")
			}
			return nil
		}
	}
	if candidate.SubgroupSourceAvailable {
		if metrics.collector.taxonomyDigest != "" && metrics.collector.taxonomyDigest != candidate.SubgroupTaxonomyDigest {
			return errors.New("subgroup taxonomy digest is immutable for collector lifetime")
		}
		metrics.collector.taxonomyDigest = candidate.SubgroupTaxonomyDigest
	}
	metrics.collector.snapshot = candidate
	metrics.collector.initialized = true
	return nil
}

func (snapshot Snapshot) validate(now time.Time) error {
	if snapshot.ObservedAt.IsZero() || snapshot.ObservedAt.After(now) || now.Sub(snapshot.ObservedAt) > SnapshotFreshnessLimit {
		return errors.New("review operations snapshot freshness is invalid")
	}
	if len(snapshot.Queues) != len(queueKinds) {
		return errors.New("review queue snapshot is incomplete")
	}
	for index, expected := range queueKinds {
		queue := snapshot.Queues[index]
		if queue.Queue != expected {
			return errors.New("review queue snapshot order is invalid")
		}
		if queue.Pending > MaximumCount || queue.OldestAge < 0 || queue.OldestAge > maximumOldestAge {
			return errors.New("review queue values are invalid")
		}
		if !queue.Available && (queue.Pending != 0 || queue.OldestAge != 0) {
			return errors.New("unavailable review queue must omit operational values")
		}
		if queue.Available && queue.Pending == 0 && queue.OldestAge != 0 {
			return errors.New("empty review queue cannot have an oldest age")
		}
	}
	if len(snapshot.Controls) != len(queueKinds)*len(controlKinds) {
		return errors.New("review control snapshot is incomplete")
	}
	index := 0
	for _, queue := range queueKinds {
		for _, control := range controlKinds {
			aggregate := snapshot.Controls[index]
			index++
			if aggregate.Queue != queue || aggregate.Control != control {
				return errors.New("review control snapshot order is invalid")
			}
			if aggregate.Eligible > MaximumCount || aggregate.Satisfied > aggregate.Eligible ||
				aggregate.Unavailable > aggregate.Eligible-aggregate.Satisfied {
				return errors.New("review control counts are invalid")
			}
			if !aggregate.Available && (aggregate.Eligible != 0 || aggregate.Satisfied != 0 || aggregate.Unavailable != 0) {
				return errors.New("unavailable review control must omit operational values")
			}
		}
	}
	return snapshot.validateSubgroups()
}

func (snapshot Snapshot) validateSubgroups() error {
	if !snapshot.SubgroupSourceAvailable {
		if snapshot.SubgroupTaxonomyDigest != "" || len(snapshot.Subgroups) != 0 || snapshot.SuppressedSubgroupCount != 0 {
			return errors.New("unavailable subgroup source must omit operational values")
		}
		return nil
	}
	if !isCanonicalDigest(snapshot.SubgroupTaxonomyDigest) {
		return errors.New("subgroup taxonomy digest is invalid")
	}
	if len(snapshot.Subgroups) > MaximumSubgroups {
		return errors.New("too many subgroup aggregates")
	}
	if snapshot.SuppressedSubgroupCount > uint64(MaximumSubgroups-len(snapshot.Subgroups)) {
		return errors.New("suppressed subgroup count is invalid")
	}
	var prior uint8
	for _, subgroup := range snapshot.Subgroups {
		if subgroup.Slot == 0 || int(subgroup.Slot) > MaximumSubgroups || subgroup.Slot <= prior {
			return errors.New("subgroup slots are invalid")
		}
		if subgroup.Reviewed < MinimumSubgroupCell || subgroup.Reviewed > MaximumCount || subgroup.Overturned > subgroup.Reviewed {
			return errors.New("subgroup overturn counts are invalid")
		}
		prior = subgroup.Slot
	}
	return nil
}

func isCanonicalDigest(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == 32 && hex.EncodeToString(decoded) == value
}

func cloneSnapshot(snapshot Snapshot) Snapshot {
	clone := snapshot
	clone.Queues = append([]QueueAggregate(nil), snapshot.Queues...)
	clone.Controls = append([]ControlAggregate(nil), snapshot.Controls...)
	clone.Subgroups = append([]SubgroupOverturnAggregate(nil), snapshot.Subgroups...)
	return clone
}

type collector struct {
	mu             sync.RWMutex
	initialized    bool
	snapshot       Snapshot
	taxonomyDigest string
	now            func() time.Time

	queueAvailable     *prometheus.Desc
	queuePending       *prometheus.Desc
	queueOldestAge     *prometheus.Desc
	controlAvailable   *prometheus.Desc
	controlCounts      *prometheus.Desc
	subgroupAvailable  *prometheus.Desc
	subgroupTaxonomy   *prometheus.Desc
	subgroupCounts     *prometheus.Desc
	subgroupSuppressed *prometheus.Desc
}

func newCollector() *collector {
	return &collector{
		now:                time.Now,
		queueAvailable:     prometheus.NewDesc("virtengine_review_operations_queue_source_available", "Whether the review queue projection is authoritative and available.", []string{"queue"}, nil),
		queuePending:       prometheus.NewDesc("virtengine_review_operations_queue_pending", "Pending reviews in an available authoritative queue projection.", []string{"queue"}, nil),
		queueOldestAge:     prometheus.NewDesc("virtengine_review_operations_queue_oldest_age_seconds", "Age in seconds of the oldest item in an available authoritative review queue.", []string{"queue"}, nil),
		controlAvailable:   prometheus.NewDesc("virtengine_review_operations_control_source_available", "Whether aggregate evidence for a review control is authoritative and available.", []string{"queue", "control"}, nil),
		controlCounts:      prometheus.NewDesc("virtengine_review_operations_control_observations", "Aggregate review control observations by bounded result.", []string{"queue", "control", "result"}, nil),
		subgroupAvailable:  prometheus.NewDesc("virtengine_review_operations_subgroup_source_available", "Whether governed subgroup overturn aggregates are available.", nil, nil),
		subgroupTaxonomy:   prometheus.NewDesc("virtengine_review_operations_subgroup_taxonomy_info", "Current governed subgroup taxonomy digest.", []string{"taxonomy_digest"}, nil),
		subgroupCounts:     prometheus.NewDesc("virtengine_review_operations_subgroup_overturn_observations", "Privacy-thresholded subgroup review and restored-overturn observations.", []string{"subgroup", "result"}, nil),
		subgroupSuppressed: prometheus.NewDesc("virtengine_review_operations_subgroups_suppressed", "Number of subgroup cells omitted by privacy suppression.", nil, nil),
	}
}

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range []*prometheus.Desc{
		c.queueAvailable, c.queuePending, c.queueOldestAge, c.controlAvailable, c.controlCounts,
		c.subgroupAvailable, c.subgroupTaxonomy, c.subgroupCounts, c.subgroupSuppressed,
	} {
		ch <- desc
	}
}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	c.mu.RLock()
	initialized := c.initialized
	snapshot := cloneSnapshot(c.snapshot)
	now := c.now()
	c.mu.RUnlock()
	elapsed := now.Sub(snapshot.ObservedAt)
	fresh := initialized && elapsed >= 0 && elapsed <= SnapshotFreshnessLimit

	queues := make(map[QueueKind]QueueAggregate, len(snapshot.Queues))
	controls := make(map[string]ControlAggregate, len(snapshot.Controls))
	if fresh {
		for _, queue := range snapshot.Queues {
			queues[queue.Queue] = queue
		}
		for _, aggregate := range snapshot.Controls {
			controls[controlKey(aggregate.Queue, aggregate.Control)] = aggregate
		}
	}
	for _, queueKind := range queueKinds {
		queue := queues[queueKind]
		queueAvailable := fresh && queue.Available
		available := boolValue(queueAvailable)
		ch <- prometheus.MustNewConstMetric(c.queueAvailable, prometheus.GaugeValue, available, string(queueKind))
		if queueAvailable {
			ch <- prometheus.MustNewConstMetric(c.queuePending, prometheus.GaugeValue, float64(queue.Pending), string(queueKind))
			ch <- prometheus.MustNewConstMetric(c.queueOldestAge, prometheus.GaugeValue, (queue.OldestAge + elapsed).Seconds(), string(queueKind))
		}
		for _, controlKind := range controlKinds {
			aggregate := controls[controlKey(queueKind, controlKind)]
			ch <- prometheus.MustNewConstMetric(c.controlAvailable, prometheus.GaugeValue, boolValue(aggregate.Available), string(queueKind), string(controlKind))
			if !aggregate.Available {
				continue
			}
			for _, result := range []struct {
				name  string
				value uint64
			}{{"eligible", aggregate.Eligible}, {"satisfied", aggregate.Satisfied}, {"unavailable", aggregate.Unavailable}} {
				ch <- prometheus.MustNewConstMetric(c.controlCounts, prometheus.GaugeValue, float64(result.value), string(queueKind), string(controlKind), result.name)
			}
		}
	}
	ch <- prometheus.MustNewConstMetric(c.subgroupAvailable, prometheus.GaugeValue, boolValue(fresh && snapshot.SubgroupSourceAvailable))
	if !fresh || !snapshot.SubgroupSourceAvailable {
		return
	}
	ch <- prometheus.MustNewConstMetric(c.subgroupTaxonomy, prometheus.GaugeValue, 1, snapshot.SubgroupTaxonomyDigest)
	for _, subgroup := range snapshot.Subgroups {
		label := fmt.Sprintf("cohort_%02d", subgroup.Slot)
		ch <- prometheus.MustNewConstMetric(c.subgroupCounts, prometheus.GaugeValue, float64(subgroup.Reviewed), label, "reviewed")
		ch <- prometheus.MustNewConstMetric(c.subgroupCounts, prometheus.GaugeValue, float64(subgroup.Overturned), label, "restored_overturn")
	}
	ch <- prometheus.MustNewConstMetric(c.subgroupSuppressed, prometheus.GaugeValue, float64(snapshot.SuppressedSubgroupCount))
}

func controlKey(queue QueueKind, control ControlKind) string {
	return string(queue) + "\x00" + string(control)
}

func boolValue(value bool) float64 {
	if value {
		return 1
	}
	return 0
}
