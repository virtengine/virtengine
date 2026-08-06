// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"math"
	"sort"
	"strings"
)

const (
	CanonicalReliabilityEnvelopeVersion uint32 = 1
	CanonicalReliabilityDomain                 = "virtengine.benchmark.reliability-envelope.v1"
	CanonicalReliabilityNonceSize              = 32
	CanonicalReliabilityDigestSize             = sha256.Size
	CanonicalReliabilityMaxFieldBytes          = 2048
)

var canonicalReliabilityMagic = [8]byte{'V', 'E', 'R', 'E', 'L', 'I', 'A', 0x01}

var requiredReliabilitySources = []ReliabilitySourceKind{
	ReliabilitySourceAnomalies,
	ReliabilitySourceBenchmark,
	ReliabilitySourceDisputes,
	ReliabilitySourceIncidents,
	ReliabilitySourceProvisioning,
	ReliabilitySourceUptime,
}

type ReliabilitySourceKind string

const (
	ReliabilitySourceAnomalies    ReliabilitySourceKind = "anomalies"
	ReliabilitySourceBenchmark    ReliabilitySourceKind = "benchmark"
	ReliabilitySourceDisputes     ReliabilitySourceKind = "disputes"
	ReliabilitySourceIncidents    ReliabilitySourceKind = "incidents"
	ReliabilitySourceProvisioning ReliabilitySourceKind = "provisioning"
	ReliabilitySourceUptime       ReliabilitySourceKind = "uptime"
)

type ReliabilityVerificationStatus string

const (
	ReliabilityStatusUnverified       ReliabilityVerificationStatus = "unverified"
	ReliabilityStatusVerified         ReliabilityVerificationStatus = "verified"
	ReliabilityStatusMissingSource    ReliabilityVerificationStatus = "missing_source"
	ReliabilityStatusStaleSource      ReliabilityVerificationStatus = "stale_source"
	ReliabilityStatusInvalidSignature ReliabilityVerificationStatus = "invalid_signature"
)

type ReliabilitySourceCommitment struct {
	Kind            ReliabilitySourceKind `json:"kind"`
	Digest          []byte                `json:"digest"`
	WindowStartUnix int64                 `json:"window_start_unix"`
	WindowEndUnix   int64                 `json:"window_end_unix"`
	ObservedThrough int64                 `json:"observed_through_unix"`
	RecordCount     uint64                `json:"record_count"`
}

type CanonicalReliabilityEnvelopeV1 struct {
	Version                uint32                        `json:"version"`
	Domain                 string                        `json:"domain"`
	ChainID                string                        `json:"chain_id"`
	ProviderAddress        string                        `json:"provider_address"`
	ClusterID              string                        `json:"cluster_id"`
	ReportID               string                        `json:"report_id"`
	HardwareClass          string                        `json:"hardware_class"`
	HardwareManifestDigest []byte                        `json:"hardware_manifest_digest"`
	SuiteVersion           string                        `json:"suite_version"`
	SuiteDigest            []byte                        `json:"suite_digest"`
	WorkloadImageDigest    []byte                        `json:"workload_image_digest"`
	ResultDigest           []byte                        `json:"result_digest"`
	Inputs                 ReliabilityScoreInputs        `json:"inputs"`
	RunnerID               string                        `json:"runner_id"`
	ProviderKeyID          string                        `json:"provider_key_id"`
	ProviderKeyEpoch       uint64                        `json:"provider_key_epoch"`
	Sequence               uint64                        `json:"sequence"`
	Nonce                  []byte                        `json:"nonce"`
	ObservationUnix        int64                         `json:"observation_unix"`
	IssuedAtHeight         int64                         `json:"issued_at_height"`
	ExpiresAtHeight        int64                         `json:"expires_at_height"`
	IssuedAtUnix           int64                         `json:"issued_at_unix"`
	ExpiresAtUnix          int64                         `json:"expires_at_unix"`
	FreshnessPolicyVersion uint32                        `json:"freshness_policy_version"`
	Sources                []ReliabilitySourceCommitment `json:"sources"`
	Signature              []byte                        `json:"signature"`
}

type ReliabilityVerificationResult struct {
	Status          ReliabilityVerificationStatus `json:"status"`
	Verified        bool                          `json:"verified"`
	EffectiveScore  int64                         `json:"effective_score"`
	ComponentScores ComponentScores               `json:"component_scores"`
	EnvelopeDigest  []byte                        `json:"envelope_digest,omitempty"`
	ReplayKey       []byte                        `json:"replay_key,omitempty"`
}

func CanonicalReliabilitySignBytes(envelope CanonicalReliabilityEnvelopeV1) ([]byte, error) {
	if err := validateCanonicalReliabilityEnvelope(envelope); err != nil {
		return nil, err
	}
	var out bytes.Buffer
	out.Write(canonicalReliabilityMagic[:])
	writeReliabilityUint32(&out, envelope.Version)
	for _, value := range []string{
		envelope.Domain, envelope.ChainID, envelope.ProviderAddress, envelope.ClusterID, envelope.ReportID,
		envelope.HardwareClass, envelope.SuiteVersion, envelope.RunnerID, envelope.ProviderKeyID,
	} {
		writeReliabilityString(&out, value)
	}
	for _, digest := range [][]byte{envelope.HardwareManifestDigest, envelope.SuiteDigest, envelope.WorkloadImageDigest, envelope.ResultDigest} {
		writeReliabilityBytes(&out, digest)
	}
	writeReliabilityInputs(&out, envelope.Inputs)
	writeReliabilityUint64(&out, envelope.ProviderKeyEpoch)
	writeReliabilityUint64(&out, envelope.Sequence)
	writeReliabilityBytes(&out, envelope.Nonce)
	writeReliabilityInt64(&out, envelope.ObservationUnix)
	writeReliabilityInt64(&out, envelope.IssuedAtHeight)
	writeReliabilityInt64(&out, envelope.ExpiresAtHeight)
	writeReliabilityInt64(&out, envelope.IssuedAtUnix)
	writeReliabilityInt64(&out, envelope.ExpiresAtUnix)
	writeReliabilityUint32(&out, envelope.FreshnessPolicyVersion)
	writeReliabilityUint32(&out, uint32(len(envelope.Sources))) //nolint:gosec // bounded by required source set
	for _, source := range envelope.Sources {
		writeReliabilityString(&out, string(source.Kind))
		writeReliabilityBytes(&out, source.Digest)
		writeReliabilityInt64(&out, source.WindowStartUnix)
		writeReliabilityInt64(&out, source.WindowEndUnix)
		writeReliabilityInt64(&out, source.ObservedThrough)
		writeReliabilityUint64(&out, source.RecordCount)
	}
	return out.Bytes(), nil
}

func CanonicalReliabilityDigest(envelope CanonicalReliabilityEnvelopeV1) ([]byte, error) {
	signBytes, err := CanonicalReliabilitySignBytes(envelope)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(signBytes)
	return digest[:], nil
}

func CanonicalReliabilityReplayKey(envelope CanonicalReliabilityEnvelopeV1) ([]byte, error) {
	if err := validateCanonicalReliabilityEnvelope(envelope); err != nil {
		return nil, err
	}
	var out bytes.Buffer
	writeReliabilityString(&out, CanonicalReliabilityDomain+".replay")
	writeReliabilityString(&out, envelope.ChainID)
	writeReliabilityString(&out, envelope.ProviderAddress)
	writeReliabilityString(&out, envelope.ProviderKeyID)
	writeReliabilityUint64(&out, envelope.ProviderKeyEpoch)
	writeReliabilityUint64(&out, envelope.Sequence)
	writeReliabilityBytes(&out, envelope.Nonce)
	digest := sha256.Sum256(out.Bytes())
	return digest[:], nil
}

func VerifyCanonicalReliabilityEnvelope(params Params, envelope CanonicalReliabilityEnvelopeV1, publicKey ed25519.PublicKey, currentHeight, currentUnix int64) (ReliabilityVerificationResult, error) {
	result := ReliabilityVerificationResult{Status: ReliabilityStatusUnverified}
	if !params.CanonicalEnvelopeEnabled || !params.VerifiedReliabilityEnabled {
		return result, errors.New("canonical reliability verification is disabled")
	}
	digest, err := CanonicalReliabilityDigest(envelope)
	if err != nil {
		return result, err
	}
	result.EnvelopeDigest = digest
	result.ReplayKey, err = CanonicalReliabilityReplayKey(envelope)
	if err != nil {
		return result, err
	}
	if len(publicKey) != ed25519.PublicKeySize || !ed25519.Verify(publicKey, digest, envelope.Signature) {
		result.Status = ReliabilityStatusInvalidSignature
		return result, nil
	}
	if currentHeight < envelope.IssuedAtHeight || currentHeight > envelope.ExpiresAtHeight ||
		currentUnix < envelope.IssuedAtUnix-params.MaxFutureSkewSeconds || currentUnix > envelope.ExpiresAtUnix ||
		envelope.IssuedAtUnix > currentUnix+params.MaxFutureSkewSeconds || envelope.ObservationUnix > currentUnix+params.MaxFutureSkewSeconds ||
		currentUnix-envelope.ObservationUnix > params.MaxEnvelopeAgeSeconds ||
		envelope.ExpiresAtHeight-envelope.IssuedAtHeight > params.MaxEnvelopeLifetimeBlocks ||
		envelope.ExpiresAtUnix-envelope.IssuedAtUnix > params.MaxEnvelopeLifetimeSeconds {
		result.Status = ReliabilityStatusStaleSource
		return result, nil
	}
	missing, stale := classifyReliabilitySources(envelope.Sources, currentUnix, params.MaxSourceAgeSeconds, params.MaxFutureSkewSeconds)
	if missing {
		result.Status = ReliabilityStatusMissingSource
		return result, nil
	}
	if stale {
		result.Status = ReliabilityStatusStaleSource
		return result, nil
	}
	score, components, err := computeCanonicalReliabilityScore(envelope.Inputs)
	if err != nil {
		return result, err
	}
	result.Status, result.Verified, result.EffectiveScore, result.ComponentScores = ReliabilityStatusVerified, true, score, components
	return result, nil
}

func validateCanonicalReliabilityEnvelope(envelope CanonicalReliabilityEnvelopeV1) error {
	if envelope.Version != CanonicalReliabilityEnvelopeVersion || envelope.Domain != CanonicalReliabilityDomain {
		return errors.New("unsupported canonical reliability envelope")
	}
	for _, value := range []string{
		envelope.ChainID, envelope.ProviderAddress, envelope.ClusterID, envelope.ReportID, envelope.HardwareClass,
		envelope.SuiteVersion, envelope.RunnerID, envelope.ProviderKeyID,
	} {
		if value == "" || len(value) > CanonicalReliabilityMaxFieldBytes || strings.IndexByte(value, 0) >= 0 {
			return errors.New("invalid canonical reliability string field")
		}
	}
	for _, digest := range [][]byte{envelope.HardwareManifestDigest, envelope.SuiteDigest, envelope.WorkloadImageDigest, envelope.ResultDigest} {
		if len(digest) != CanonicalReliabilityDigestSize {
			return errors.New("invalid canonical reliability digest")
		}
	}
	if envelope.ProviderKeyEpoch == 0 || envelope.Sequence == 0 || len(envelope.Nonce) != CanonicalReliabilityNonceSize || envelope.FreshnessPolicyVersion == 0 {
		return errors.New("invalid canonical reliability replay or policy binding")
	}
	if envelope.ObservationUnix <= 0 || envelope.IssuedAtHeight <= 0 || envelope.ExpiresAtHeight <= envelope.IssuedAtHeight ||
		envelope.IssuedAtUnix <= 0 || envelope.ExpiresAtUnix <= envelope.IssuedAtUnix {
		return errors.New("invalid canonical reliability time bounds")
	}
	if err := validateReliabilityInputs(envelope.Inputs); err != nil {
		return err
	}
	previous := ReliabilitySourceKind("")
	seen := make(map[ReliabilitySourceKind]bool)
	for _, source := range envelope.Sources {
		if !isRequiredReliabilitySource(source.Kind) || seen[source.Kind] || (previous != "" && source.Kind <= previous) {
			return errors.New("reliability sources must be unique, known, and sorted")
		}
		if len(source.Digest) != CanonicalReliabilityDigestSize || source.RecordCount == 0 || source.WindowStartUnix <= 0 ||
			source.WindowEndUnix < source.WindowStartUnix || source.ObservedThrough < source.WindowEndUnix {
			return errors.New("invalid reliability source commitment")
		}
		seen[source.Kind], previous = true, source.Kind
	}
	return nil
}

func classifyReliabilitySources(sources []ReliabilitySourceCommitment, currentUnix, maxAge, maxFutureSkew int64) (bool, bool) {
	seen := make(map[ReliabilitySourceKind]ReliabilitySourceCommitment, len(sources))
	for _, source := range sources {
		seen[source.Kind] = source
	}
	for _, required := range requiredReliabilitySources {
		source, ok := seen[required]
		if !ok {
			return true, false
		}
		if currentUnix-source.ObservedThrough > maxAge || source.ObservedThrough > currentUnix+maxFutureSkew {
			return false, true
		}
	}
	return false, false
}

func computeCanonicalReliabilityScore(inputs ReliabilityScoreInputs) (int64, ComponentScores, error) {
	if err := validateReliabilityInputs(inputs); err != nil {
		return 0, ComponentScores{}, err
	}
	if inputs.ProvisioningAttempts == 0 || inputs.TotalUptimeSeconds+inputs.TotalDowntimeSeconds == 0 {
		return 0, ComponentScores{}, errors.New("required reliability observations are missing")
	}
	score, components := ComputeReliabilityScore(inputs)
	return score, components, nil
}

func validateReliabilityInputs(inputs ReliabilityScoreInputs) error {
	values := []int64{
		inputs.BenchmarkSummary, inputs.ProvisioningSuccessRate, inputs.ProvisioningAttempts, inputs.ProvisioningSuccesses,
		inputs.MeanTimeToProvision, inputs.MeanTimeBetweenFailures, inputs.TotalUptimeSeconds, inputs.TotalDowntimeSeconds,
		inputs.DisputeCount, inputs.DisputesResolved, inputs.DisputesLost, inputs.AnomalyFlagCount,
	}
	for _, value := range values {
		if value < 0 || value > math.MaxInt64/10000 {
			return errors.New("reliability input out of bounds")
		}
	}
	if inputs.BenchmarkSummary > 10000 || inputs.ProvisioningSuccessRate > FixedPointScale ||
		inputs.ProvisioningSuccesses > inputs.ProvisioningAttempts || inputs.DisputesResolved+inputs.DisputesLost > inputs.DisputeCount {
		return errors.New("inconsistent reliability inputs")
	}
	return nil
}

func isRequiredReliabilitySource(kind ReliabilitySourceKind) bool {
	index := sort.Search(len(requiredReliabilitySources), func(i int) bool { return requiredReliabilitySources[i] >= kind })
	return index < len(requiredReliabilitySources) && requiredReliabilitySources[index] == kind
}

func writeReliabilityInputs(out *bytes.Buffer, inputs ReliabilityScoreInputs) {
	for _, value := range []int64{
		inputs.BenchmarkSummary, inputs.ProvisioningSuccessRate, inputs.ProvisioningAttempts, inputs.ProvisioningSuccesses,
		inputs.MeanTimeToProvision, inputs.MeanTimeBetweenFailures, inputs.TotalUptimeSeconds, inputs.TotalDowntimeSeconds,
		inputs.DisputeCount, inputs.DisputesResolved, inputs.DisputesLost, inputs.AnomalyFlagCount,
	} {
		writeReliabilityInt64(out, value)
	}
}

func writeReliabilityString(out *bytes.Buffer, value string) {
	writeReliabilityBytes(out, []byte(value))
}
func writeReliabilityBytes(out *bytes.Buffer, value []byte) {
	writeReliabilityUint32(out, uint32(len(value))) //nolint:gosec // validation bounds all fields
	out.Write(value)
}
func writeReliabilityUint32(out *bytes.Buffer, value uint32) {
	_ = binary.Write(out, binary.BigEndian, value)
}
func writeReliabilityUint64(out *bytes.Buffer, value uint64) {
	_ = binary.Write(out, binary.BigEndian, value)
}
func writeReliabilityInt64(out *bytes.Buffer, value int64) {
	_ = binary.Write(out, binary.BigEndian, value)
}

func ReliabilityDigestHex(value []byte) string { return hex.EncodeToString(value) }
