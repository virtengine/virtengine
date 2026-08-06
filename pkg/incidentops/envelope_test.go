// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package incidentops

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
)

const testDigest = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func TestBuildFixtureEachKind(t *testing.T) {
	for _, kind := range []Kind{KindModelIncident, KindUniquenessServiceOutage, KindMassFalseMatch, KindDeletionBacklog, KindKeyCompromise} {
		t.Run(string(kind), func(t *testing.T) {
			envelope, err := BuildFixture(validInput(kind))
			if err != nil {
				t.Fatal(err)
			}
			if err := envelope.Validate(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestRatioFormulaAndPrivacyFloor(t *testing.T) {
	input := validInput(KindModelIncident)
	for index := range input.Aggregates {
		input.Aggregates[index].Numerator = 1
		input.Aggregates[index].Denominator = 3
	}
	envelope, err := BuildFixture(input)
	if err != nil {
		t.Fatal(err)
	}
	if envelope.Status != StatusFixtureSuppressed || envelope.Aggregates[0].Published != nil || envelope.Aggregates[0].CountBand != CountBandSuppressed1To19 {
		t.Fatal("low-cell ratio was not fully suppressed")
	}
	canonical, err := envelope.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(canonical, []byte(`"numerator":1`)) || bytes.Contains(canonical, []byte(`"denominator":3`)) {
		t.Fatal("low-cell values leaked into serialization")
	}

	for index := range input.Aggregates {
		input.Aggregates[index].Numerator = 7
		input.Aggregates[index].Denominator = 20
	}
	envelope, err = BuildFixture(input)
	if err != nil {
		t.Fatal(err)
	}
	if got := envelope.Aggregates[0].Published.PPM; got != 350_000 {
		t.Fatalf("unexpected PPM: %d", got)
	}
	envelope.Aggregates[0].Published.PPM++
	if envelope.Validate() == nil {
		t.Fatal("tampered formula accepted")
	}
}

func TestCountBandBoundaries(t *testing.T) {
	tests := []struct {
		count uint64
		band  CountBand
	}{{0, CountBandNone}, {1, CountBandSuppressed1To19}, {19, CountBandSuppressed1To19}, {20, CountBand20To99}, {99, CountBand20To99}, {100, CountBand100To999}, {999, CountBand100To999}, {1_000, CountBand1000Plus}, {MaxCount, CountBand1000Plus}}
	for _, test := range tests {
		input := validInput(KindKeyCompromise)
		input.Aggregates[0].Count = test.count
		envelope, err := BuildFixture(input)
		if err != nil {
			t.Fatal(err)
		}
		if envelope.Aggregates[0].CountBand != test.band {
			t.Fatalf("count %d: got %q", test.count, envelope.Aggregates[0].CountBand)
		}
		if envelope.Aggregates[0].Published != nil {
			t.Fatal("count aggregate published an exact value")
		}
	}
}

func TestUnavailableInputsOmitAggregates(t *testing.T) {
	for _, status := range []SourceStatus{SourceStatusUnavailable, SourceStatusStale, SourceStatusInvalid} {
		input := validInput(KindMassFalseMatch)
		input.Sources[0].Status = status
		envelope, err := BuildFixture(input)
		if err != nil {
			t.Fatal(err)
		}
		if envelope.Status != StatusSourceUnavailable || len(envelope.Aggregates) != 0 {
			t.Fatal("nonfixture source did not force unavailable evidence")
		}
	}
	input := validInput(KindModelIncident)
	input.Metadata.Policy = DigestRef{Status: DigestStatusUnavailable}
	envelope, err := BuildFixture(input)
	if err != nil || envelope.Status != StatusSourceUnavailable || len(envelope.Aggregates) != 0 {
		t.Fatal("unavailable digest reference did not force unavailable evidence")
	}
	if _, err := BuildUnavailable(validMetadata(), KindModelIncident, SeverityCritical, validSources(KindModelIncident)); err == nil {
		t.Fatal("healthy sources accepted as unavailable")
	}
}

func TestOrderingDuplicatesAndUnknowns(t *testing.T) {
	input := validInput(KindDeletionBacklog)
	input.Aggregates[0], input.Aggregates[1] = input.Aggregates[1], input.Aggregates[0]
	if _, err := BuildFixture(input); err == nil {
		t.Fatal("unordered aggregates accepted")
	}
	input = validInput(KindDeletionBacklog)
	input.Aggregates[1] = input.Aggregates[0]
	if _, err := BuildFixture(input); err == nil {
		t.Fatal("duplicate aggregates accepted")
	}
	input = validInput(KindDeletionBacklog)
	input.Aggregates[0].Kind = "unknown"
	if _, err := BuildFixture(input); err == nil {
		t.Fatal("unknown aggregate accepted")
	}
	input = validInput(KindModelIncident)
	input.Sources = append(input.Sources, input.Sources[0])
	if _, err := BuildFixture(input); err == nil {
		t.Fatal("duplicate source accepted")
	}
	input = validInput(KindModelIncident)
	input.Sources[0].Name = "unknown"
	if _, err := BuildFixture(input); err == nil {
		t.Fatal("unknown source accepted")
	}
}

func TestMetadataAndClassificationContradictions(t *testing.T) {
	mutations := []func(*Envelope){
		func(envelope *Envelope) { envelope.Domain = "wrong" },
		func(envelope *Envelope) { envelope.Version++ },
		func(envelope *Envelope) { envelope.Class = "production" },
		func(envelope *Envelope) { envelope.Certification = "certified" },
		func(envelope *Envelope) { envelope.Metadata.WindowStartUnixSeconds++ },
		func(envelope *Envelope) { envelope.Metadata.WindowEndUnixSeconds++ },
		func(envelope *Envelope) { envelope.Metadata.Policy.Digest = strings.ToUpper(testDigest) },
		func(envelope *Envelope) {
			envelope.Metadata.Profile = DigestRef{Status: DigestStatusUnavailable, Digest: testDigest}
		},
		func(envelope *Envelope) { envelope.Metadata.FixtureInputDigest = "bad" },
		func(envelope *Envelope) { envelope.Status = StatusFixtureSuppressed },
	}
	for index, mutate := range mutations {
		envelope, err := BuildFixture(validInput(KindModelIncident))
		if err != nil {
			t.Fatal(err)
		}
		mutate(envelope)
		if envelope.Validate() == nil {
			t.Fatalf("contradiction %d accepted", index)
		}
	}
}

func TestCanonicalDigestGoldenAndTamper(t *testing.T) {
	envelope, err := BuildFixture(validInput(KindMassFalseMatch))
	if err != nil {
		t.Fatal(err)
	}
	digest, err := envelope.Digest()
	if err != nil {
		t.Fatal(err)
	}
	const golden = "0c736e194879edbf0f2f4f6e70f31f0522b830371fc9b169ee72c7df052943ee"
	if got := hex.EncodeToString(digest[:]); got != golden {
		t.Fatalf("golden digest mismatch: got %s", got)
	}
	envelope.Severity = SeverityCritical
	tampered, err := envelope.Digest()
	if err != nil {
		t.Fatal(err)
	}
	if tampered == digest {
		t.Fatal("tamper did not change digest")
	}
}

func TestPrivateCanariesAbsentFromSerializationAndErrors(t *testing.T) {
	canaries := []string{"account_id", "subject_id", "biometric_id", "document_id", "nullifier_id", "request_id", "receipt_id", "key_id", "provider_id", "incident_id"}
	envelope, err := BuildFixture(validInput(KindDeletionBacklog))
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := envelope.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	for _, canary := range canaries {
		if bytes.Contains(bytes.ToLower(canonical), []byte(canary)) {
			t.Fatalf("private canary serialized: %s", canary)
		}
	}
	input := validInput(KindModelIncident)
	input.Metadata.FixtureInputDigest = strings.Join(canaries, "-")
	_, err = BuildFixture(input)
	if err == nil {
		t.Fatal("invalid private input accepted")
	}
	for _, canary := range canaries {
		if strings.Contains(strings.ToLower(err.Error()), canary) {
			t.Fatalf("private canary echoed in error: %s", canary)
		}
	}
}

func TestRawLimitsAndLogicalBounds(t *testing.T) {
	input := validInput(KindModelIncident)
	input.Aggregates[0].Numerator = 21
	input.Aggregates[0].Denominator = 20
	if _, err := BuildFixture(input); err == nil {
		t.Fatal("numerator above denominator accepted")
	}
	input = validInput(KindKeyCompromise)
	input.Aggregates[0].Count = MaxCount + 1
	if _, err := BuildFixture(input); err == nil {
		t.Fatal("count above maximum accepted")
	}
	input = validInput(KindMassFalseMatch)
	input.Aggregates[1].Denominator++
	if _, err := BuildFixture(input); err == nil {
		t.Fatal("contradictory ratio aggregate pair accepted")
	}
}

func FuzzEnvelopeValidate(f *testing.F) {
	envelope, err := BuildFixture(validInput(KindModelIncident))
	if err != nil {
		f.Fatal(err)
	}
	seed, err := json.Marshal(envelope)
	if err != nil {
		f.Fatal(err)
	}
	f.Add(seed)
	f.Fuzz(func(t *testing.T, data []byte) {
		var candidate Envelope
		if json.Unmarshal(data, &candidate) == nil {
			_ = candidate.Validate()
			_, _ = candidate.CanonicalBytes()
			_, _ = candidate.Digest()
		}
	})
}

func validMetadata() Metadata {
	return Metadata{
		WindowStartUnixSeconds: 1_728_000,
		WindowEndUnixSeconds:   1_814_400,
		Policy:                 DigestRef{Status: DigestStatusAvailable, Digest: testDigest},
		Profile:                DigestRef{Status: DigestStatusAvailable, Digest: testDigest},
		FixtureInputDigest:     testDigest,
	}
}

func validInput(kind Kind) FixtureInput {
	required, _ := requiredAggregates(kind)
	inputs := make([]AggregateInput, len(required))
	for index, aggregateKind := range required {
		inputs[index] = AggregateInput{Kind: aggregateKind}
		if kind == KindModelIncident || kind == KindUniquenessServiceOutage || kind == KindMassFalseMatch {
			inputs[index].Numerator = 5
			inputs[index].Denominator = 20
		} else {
			inputs[index].Count = 20
		}
	}
	return FixtureInput{Metadata: validMetadata(), Kind: kind, Severity: SeverityWarning, Sources: validSources(kind), Aggregates: inputs}
}

func validSources(kind Kind) []Source {
	required, _ := requiredSources(kind)
	sources := make([]Source, len(required))
	for index, sourceName := range required {
		sources[index] = Source{Name: sourceName, Status: SourceStatusFixture}
	}
	return sources
}
