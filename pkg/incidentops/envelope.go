// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package incidentops

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"math/bits"
)

const (
	Domain               = "virtengine/privacy-safe-incident-fixture/v1"
	Version       uint32 = 1
	WindowSeconds uint64 = 86_400
	PrivacyFloor  uint64 = 20
	PPMScale      uint64 = 1_000_000
	MaxCount      uint64 = 1_000_000_000_000
)

type Class string
type Certification string
type Kind string
type Severity string
type Status string
type DigestStatus string
type SourceName string
type SourceStatus string
type AggregateKind string
type CountBand string

const (
	ClassSyntheticFixture     Class         = "synthetic_fixture"
	CertificationNotCertified Certification = "not_certified"

	KindModelIncident           Kind         = "model_incident"
	KindUniquenessServiceOutage Kind         = "uniqueness_service_outage"
	KindMassFalseMatch          Kind         = "mass_false_match"
	KindDeletionBacklog         Kind         = "deletion_backlog"
	KindKeyCompromise           Kind         = "key_compromise"
	SeverityWarning             Severity     = "warning"
	SeverityCritical            Severity     = "critical"
	StatusFixtureActive         Status       = "fixture_active"
	StatusFixtureSuppressed     Status       = "fixture_suppressed"
	StatusSourceUnavailable     Status       = "source_unavailable"
	DigestStatusAvailable       DigestStatus = "available"
	DigestStatusUnavailable     DigestStatus = "unavailable"
	SourceStatusFixture         SourceStatus = "fixture"
	SourceStatusUnavailable     SourceStatus = "unavailable"
	SourceStatusStale           SourceStatus = "stale"
	SourceStatusInvalid         SourceStatus = "invalid"
	CountBandNone               CountBand    = "none"
	CountBandSuppressed1To19    CountBand    = "suppressed_1_19"
	CountBand20To99             CountBand    = "20_99"
	CountBand100To999           CountBand    = "100_999"
	CountBand1000Plus           CountBand    = "1000_plus"

	SourceModelOperations      SourceName = "model_operations"
	SourceUniquenessOperations SourceName = "uniqueness_operations"
	SourceFalseMatchEvaluation SourceName = "false_match_evaluation"
	SourceDeletionOperations   SourceName = "deletion_operations"
	SourceKeyManagement        SourceName = "key_management"

	AggregateModelErrorOperations            AggregateKind = "model_error_operations"
	AggregateModelTotalOperations            AggregateKind = "model_total_operations"
	AggregateUniquenessUnavailableOperations AggregateKind = "uniqueness_unavailable_operations"
	AggregateUniquenessTotalOperations       AggregateKind = "uniqueness_total_operations"
	AggregateConfirmedFalseMatches           AggregateKind = "confirmed_false_matches"
	AggregateFalseMatchEvaluations           AggregateKind = "false_match_evaluations"
	AggregateDeletionPending                 AggregateKind = "deletion_pending"
	AggregateDeletionOverdue                 AggregateKind = "deletion_overdue"
	AggregateDeletionReceiptUnresolved       AggregateKind = "deletion_receipt_unresolved"
	AggregateLegalHoldConflicts              AggregateKind = "legal_hold_conflicts"
	AggregateKeyCompromiseIndicators         AggregateKind = "key_compromise_indicators"
	AggregateKeyDestructionUnresolved        AggregateKind = "key_destruction_unresolved"
)

type DigestRef struct {
	Status DigestStatus `json:"status"`
	Digest string       `json:"digest"`
}

type Metadata struct {
	WindowStartUnixSeconds uint64    `json:"window_start_unix_seconds"`
	WindowEndUnixSeconds   uint64    `json:"window_end_unix_seconds"`
	Policy                 DigestRef `json:"policy"`
	Profile                DigestRef `json:"profile"`
	FixtureInputDigest     string    `json:"fixture_input_digest"`
}

type Source struct {
	Name   SourceName   `json:"name"`
	Status SourceStatus `json:"status"`
}

type AggregateInput struct {
	Kind        AggregateKind
	Numerator   uint64
	Denominator uint64
	Count       uint64
}

type FixtureInput struct {
	Metadata   Metadata
	Kind       Kind
	Severity   Severity
	Sources    []Source
	Aggregates []AggregateInput
}

type Published struct {
	Numerator   uint64 `json:"numerator"`
	Denominator uint64 `json:"denominator"`
	PPM         uint64 `json:"ppm"`
}

type Aggregate struct {
	Kind      AggregateKind `json:"kind"`
	Published *Published    `json:"published,omitempty"`
	CountBand CountBand     `json:"count_band,omitempty"`
}

type Envelope struct {
	Domain        string        `json:"domain"`
	Version       uint32        `json:"version"`
	Class         Class         `json:"class"`
	Certification Certification `json:"certification"`
	Kind          Kind          `json:"kind"`
	Severity      Severity      `json:"severity"`
	Status        Status        `json:"status"`
	Metadata      Metadata      `json:"metadata"`
	Sources       []Source      `json:"sources"`
	Aggregates    []Aggregate   `json:"aggregates,omitempty"`
}

func BuildFixture(input FixtureInput) (*Envelope, error) {
	envelope := baseEnvelope(input.Metadata, input.Kind, input.Severity, input.Sources)
	available, err := validateAvailability(input.Metadata, input.Kind, input.Sources)
	if err != nil {
		return nil, err
	}
	if !available {
		envelope.Status = StatusSourceUnavailable
		if err := envelope.Validate(); err != nil {
			return nil, err
		}
		return envelope, nil
	}

	aggregates, suppressed, err := buildAggregates(input.Kind, input.Aggregates)
	if err != nil {
		return nil, err
	}
	envelope.Aggregates = aggregates
	envelope.Status = StatusFixtureActive
	if suppressed {
		envelope.Status = StatusFixtureSuppressed
	}
	if err := envelope.Validate(); err != nil {
		return nil, err
	}
	return envelope, nil
}

func BuildUnavailable(metadata Metadata, kind Kind, severity Severity, sources []Source) (*Envelope, error) {
	envelope := baseEnvelope(metadata, kind, severity, sources)
	envelope.Status = StatusSourceUnavailable
	if err := envelope.Validate(); err != nil {
		return nil, err
	}
	return envelope, nil
}

func baseEnvelope(metadata Metadata, kind Kind, severity Severity, sources []Source) *Envelope {
	return &Envelope{
		Domain: Domain, Version: Version, Class: ClassSyntheticFixture,
		Certification: CertificationNotCertified, Kind: kind, Severity: severity,
		Metadata: metadata, Sources: sources,
	}
}

func (envelope *Envelope) Validate() error {
	if envelope == nil {
		return errors.New("incident fixture envelope is required")
	}
	if envelope.Domain != Domain || envelope.Version != Version {
		return errors.New("unsupported incident fixture domain or version")
	}
	if envelope.Class != ClassSyntheticFixture || envelope.Certification != CertificationNotCertified {
		return errors.New("unsupported incident fixture classification")
	}
	if !validKind(envelope.Kind) || !validSeverity(envelope.Severity) {
		return errors.New("incident fixture kind or severity is invalid")
	}
	if err := envelope.Metadata.validate(); err != nil {
		return err
	}
	available, err := validateAvailability(envelope.Metadata, envelope.Kind, envelope.Sources)
	if err != nil {
		return err
	}
	if !available {
		if envelope.Status != StatusSourceUnavailable || len(envelope.Aggregates) != 0 {
			return errors.New("unavailable incident fixture must omit aggregates")
		}
		return nil
	}
	if envelope.Status == StatusSourceUnavailable {
		return errors.New("available sources contradict unavailable status")
	}
	suppressed, err := validateAggregates(envelope.Kind, envelope.Aggregates)
	if err != nil {
		return err
	}
	expected := StatusFixtureActive
	if suppressed {
		expected = StatusFixtureSuppressed
	}
	if envelope.Status != expected {
		return errors.New("incident fixture status is inconsistent")
	}
	return nil
}

func (envelope *Envelope) CanonicalBytes() ([]byte, error) {
	if err := envelope.Validate(); err != nil {
		return nil, err
	}
	payload, err := json.Marshal(envelope)
	if err != nil {
		return nil, errors.New("incident fixture serialization failed")
	}
	canonical := make([]byte, 0, len(Domain)+1+len(payload))
	canonical = append(canonical, Domain...)
	canonical = append(canonical, 0)
	canonical = append(canonical, payload...)
	return canonical, nil
}

func (envelope *Envelope) Digest() ([32]byte, error) {
	canonical, err := envelope.CanonicalBytes()
	if err != nil {
		return [32]byte{}, err
	}
	return sha256.Sum256(canonical), nil
}

func (metadata Metadata) validate() error {
	if metadata.WindowStartUnixSeconds%WindowSeconds != 0 ||
		metadata.WindowStartUnixSeconds > math.MaxUint64-WindowSeconds ||
		metadata.WindowEndUnixSeconds != metadata.WindowStartUnixSeconds+WindowSeconds {
		return errors.New("incident fixture window must be one UTC-aligned day")
	}
	if err := metadata.Policy.validate(); err != nil {
		return err
	}
	if err := metadata.Profile.validate(); err != nil {
		return err
	}
	if !canonicalDigest(metadata.FixtureInputDigest) {
		return errors.New("incident fixture digest is malformed")
	}
	return nil
}

func (ref DigestRef) validate() error {
	switch ref.Status {
	case DigestStatusAvailable:
		if !canonicalDigest(ref.Digest) {
			return errors.New("available digest reference is malformed")
		}
	case DigestStatusUnavailable:
		if ref.Digest != "" {
			return errors.New("unavailable digest reference must omit digest")
		}
	default:
		return errors.New("digest reference status is invalid")
	}
	return nil
}

func validateAvailability(metadata Metadata, kind Kind, sources []Source) (bool, error) {
	required, ok := requiredSources(kind)
	if !ok || len(sources) != len(required) {
		return false, errors.New("required incident fixture sources are invalid")
	}
	available := metadata.Policy.Status == DigestStatusAvailable && metadata.Profile.Status == DigestStatusAvailable
	for index, source := range sources {
		if source.Name != required[index] {
			return false, errors.New("incident fixture sources are not in canonical order")
		}
		switch source.Status {
		case SourceStatusFixture:
		case SourceStatusUnavailable, SourceStatusStale, SourceStatusInvalid:
			available = false
		default:
			return false, errors.New("incident fixture source status is invalid")
		}
	}
	return available, nil
}

func buildAggregates(kind Kind, inputs []AggregateInput) ([]Aggregate, bool, error) {
	required, ok := requiredAggregates(kind)
	if !ok || len(inputs) != len(required) {
		return nil, false, errors.New("required incident fixture aggregates are invalid")
	}
	aggregates := make([]Aggregate, len(inputs))
	suppressed := false
	ratioKind := kind == KindModelIncident || kind == KindUniquenessServiceOutage || kind == KindMassFalseMatch
	for index, input := range inputs {
		if input.Kind != required[index] {
			return nil, false, errors.New("incident fixture aggregates are not in canonical order")
		}
		if input.Numerator > MaxCount || input.Denominator > MaxCount || input.Count > MaxCount {
			return nil, false, errors.New("incident fixture aggregate count exceeds limit")
		}
		aggregates[index].Kind = input.Kind
		if ratioKind {
			if input.Count != 0 || input.Numerator > input.Denominator {
				return nil, false, errors.New("incident fixture ratio counts are invalid")
			}
			if input.Denominator < PrivacyFloor {
				aggregates[index].CountBand = CountBandSuppressed1To19
				suppressed = true
				continue
			}
			ppm, err := ratioPPM(input.Numerator, input.Denominator)
			if err != nil {
				return nil, false, err
			}
			aggregates[index].Published = &Published{Numerator: input.Numerator, Denominator: input.Denominator, PPM: ppm}
			continue
		}
		if input.Numerator != 0 || input.Denominator != 0 {
			return nil, false, errors.New("incident fixture count aggregate is invalid")
		}
		aggregates[index].CountBand = bandForCount(input.Count)
		if aggregates[index].CountBand == CountBandSuppressed1To19 {
			suppressed = true
		}
	}
	if ratioKind {
		if err := validateAggregatePairInputs(inputs); err != nil {
			return nil, false, err
		}
	}
	return aggregates, suppressed, nil
}

func validateAggregates(kind Kind, aggregates []Aggregate) (bool, error) {
	required, ok := requiredAggregates(kind)
	if !ok || len(aggregates) != len(required) {
		return false, errors.New("required incident fixture aggregates are invalid")
	}
	suppressed := false
	ratioKind := kind == KindModelIncident || kind == KindUniquenessServiceOutage || kind == KindMassFalseMatch
	for index, aggregate := range aggregates {
		if aggregate.Kind != required[index] {
			return false, errors.New("incident fixture aggregates are not in canonical order")
		}
		if ratioKind {
			if aggregate.Published == nil {
				if aggregate.CountBand != CountBandSuppressed1To19 {
					return false, errors.New("suppressed ratio aggregate is invalid")
				}
				suppressed = true
				continue
			}
			if aggregate.CountBand != "" || aggregate.Published.Numerator > MaxCount ||
				aggregate.Published.Denominator < PrivacyFloor || aggregate.Published.Denominator > MaxCount ||
				aggregate.Published.Numerator > aggregate.Published.Denominator {
				return false, errors.New("published ratio aggregate is invalid")
			}
			ppm, err := ratioPPM(aggregate.Published.Numerator, aggregate.Published.Denominator)
			if err != nil || aggregate.Published.PPM != ppm {
				return false, errors.New("published ratio aggregate formula is invalid")
			}
			continue
		}
		if aggregate.Published != nil || !validCountBand(aggregate.CountBand) {
			return false, errors.New("count band aggregate is invalid")
		}
		if aggregate.CountBand == CountBandSuppressed1To19 {
			suppressed = true
		}
	}
	if ratioKind {
		if err := validateAggregatePair(aggregates); err != nil {
			return false, err
		}
	}
	return suppressed, nil
}

func validateAggregatePairInputs(inputs []AggregateInput) error {
	if len(inputs) != 2 {
		return errors.New("incident fixture ratio aggregate pair is invalid")
	}
	if inputs[0].Numerator != inputs[1].Numerator || inputs[0].Denominator != inputs[1].Denominator {
		return errors.New("incident fixture ratio aggregate pair is inconsistent")
	}
	if inputs[0].Numerator > inputs[0].Denominator {
		return errors.New("incident fixture ratio aggregate pair is invalid")
	}
	return nil
}

func validateAggregatePair(aggregates []Aggregate) error {
	if len(aggregates) != 2 {
		return errors.New("incident fixture ratio aggregate pair is invalid")
	}
	if aggregates[0].Published == nil || aggregates[1].Published == nil {
		if aggregates[0].Published != nil || aggregates[1].Published != nil ||
			aggregates[0].CountBand != CountBandSuppressed1To19 || aggregates[1].CountBand != CountBandSuppressed1To19 {
			return errors.New("incident fixture ratio aggregate pair suppression is inconsistent")
		}
		return nil
	}
	if *aggregates[0].Published != *aggregates[1].Published {
		return errors.New("incident fixture ratio aggregate pair is inconsistent")
	}
	return nil
}

func ratioPPM(numerator, denominator uint64) (uint64, error) {
	high, low := bits.Mul64(numerator, PPMScale)
	if high != 0 || denominator == 0 {
		return 0, errors.New("incident fixture ratio arithmetic is invalid")
	}
	return low / denominator, nil
}

func bandForCount(count uint64) CountBand {
	switch {
	case count == 0:
		return CountBandNone
	case count < PrivacyFloor:
		return CountBandSuppressed1To19
	case count < 100:
		return CountBand20To99
	case count < 1_000:
		return CountBand100To999
	default:
		return CountBand1000Plus
	}
}

func validCountBand(band CountBand) bool {
	switch band {
	case CountBandNone, CountBandSuppressed1To19, CountBand20To99, CountBand100To999, CountBand1000Plus:
		return true
	default:
		return false
	}
}

func validKind(kind Kind) bool {
	_, ok := requiredSources(kind)
	return ok
}

func validSeverity(severity Severity) bool {
	return severity == SeverityWarning || severity == SeverityCritical
}

func requiredSources(kind Kind) ([]SourceName, bool) {
	switch kind {
	case KindModelIncident:
		return []SourceName{SourceModelOperations}, true
	case KindUniquenessServiceOutage:
		return []SourceName{SourceUniquenessOperations}, true
	case KindMassFalseMatch:
		return []SourceName{SourceFalseMatchEvaluation}, true
	case KindDeletionBacklog:
		return []SourceName{SourceDeletionOperations}, true
	case KindKeyCompromise:
		return []SourceName{SourceKeyManagement}, true
	default:
		return nil, false
	}
}

func requiredAggregates(kind Kind) ([]AggregateKind, bool) {
	switch kind {
	case KindModelIncident:
		return []AggregateKind{AggregateModelErrorOperations, AggregateModelTotalOperations}, true
	case KindUniquenessServiceOutage:
		return []AggregateKind{AggregateUniquenessTotalOperations, AggregateUniquenessUnavailableOperations}, true
	case KindMassFalseMatch:
		return []AggregateKind{AggregateConfirmedFalseMatches, AggregateFalseMatchEvaluations}, true
	case KindDeletionBacklog:
		return []AggregateKind{AggregateDeletionOverdue, AggregateDeletionPending, AggregateDeletionReceiptUnresolved, AggregateLegalHoldConflicts}, true
	case KindKeyCompromise:
		return []AggregateKind{AggregateKeyCompromiseIndicators, AggregateKeyDestructionUnresolved}, true
	default:
		return nil, false
	}
}

func canonicalDigest(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	decoded, err := hex.DecodeString(value)
	return err == nil && hex.EncodeToString(decoded) == value
}
