// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCanonicalReliabilityEnvelopeGolden(t *testing.T) {
	envelope, _, _ := signedReliabilityEnvelope(t)
	signBytes, err := CanonicalReliabilitySignBytes(envelope)
	require.NoError(t, err)
	require.Equal(t, "564552454c494101", hex.EncodeToString(signBytes[:8]))
	digest, err := CanonicalReliabilityDigest(envelope)
	require.NoError(t, err)
	require.Equal(t, "867bcd0756c937b2ffaed2b123d8ab06ac2c20effc4f2d1e525ee19908d00cb2", hex.EncodeToString(digest))
}

func TestCanonicalReliabilityEnvelopeTamperMatrix(t *testing.T) {
	envelope, publicKey, params := signedReliabilityEnvelope(t)
	verify := func(candidate CanonicalReliabilityEnvelopeV1) ReliabilityVerificationResult {
		result, err := VerifyCanonicalReliabilityEnvelope(params, candidate, publicKey, 105, 1_700_000_100)
		require.NoError(t, err)
		return result
	}
	result := verify(envelope)
	require.True(t, result.Verified)
	require.Equal(t, ReliabilityStatusVerified, result.Status)

	mutations := []func(*CanonicalReliabilityEnvelopeV1){
		func(value *CanonicalReliabilityEnvelopeV1) { value.ClusterID = "cluster-b" },
		func(value *CanonicalReliabilityEnvelopeV1) { value.HardwareManifestDigest[0] ^= 1 },
		func(value *CanonicalReliabilityEnvelopeV1) { value.WorkloadImageDigest[0] ^= 1 },
		func(value *CanonicalReliabilityEnvelopeV1) { value.Inputs.TotalUptimeSeconds++ },
		func(value *CanonicalReliabilityEnvelopeV1) { value.RunnerID = "runner-b" },
		func(value *CanonicalReliabilityEnvelopeV1) { value.ProviderKeyEpoch++ },
		func(value *CanonicalReliabilityEnvelopeV1) { value.Nonce[0] ^= 1 },
		func(value *CanonicalReliabilityEnvelopeV1) { value.Sources[0].Digest[0] ^= 1 },
	}
	for _, mutate := range mutations {
		candidate := cloneReliabilityEnvelope(envelope)
		mutate(&candidate)
		result = verify(candidate)
		require.False(t, result.Verified)
		require.Equal(t, ReliabilityStatusInvalidSignature, result.Status)
	}
}

func TestCanonicalReliabilityEnvelopeMissingOrStaleNeverVerified(t *testing.T) {
	envelope, publicKey, params := signedReliabilityEnvelope(t)
	privateKey := deterministicReliabilityKey(t)

	missing := cloneReliabilityEnvelope(envelope)
	missing.Sources = missing.Sources[:len(missing.Sources)-1]
	signReliabilityEnvelope(t, &missing, privateKey)
	result, err := VerifyCanonicalReliabilityEnvelope(params, missing, publicKey, 105, 1_700_000_100)
	require.NoError(t, err)
	require.Equal(t, ReliabilityStatusMissingSource, result.Status)
	require.False(t, result.Verified)
	require.Zero(t, result.EffectiveScore)

	stale := cloneReliabilityEnvelope(envelope)
	stale.Sources[0].ObservedThrough = 1_699_000_000
	stale.Sources[0].WindowEndUnix = 1_699_000_000
	stale.Sources[0].WindowStartUnix = 1_698_999_000
	signReliabilityEnvelope(t, &stale, privateKey)
	result, err = VerifyCanonicalReliabilityEnvelope(params, stale, publicKey, 105, 1_700_000_100)
	require.NoError(t, err)
	require.Equal(t, ReliabilityStatusStaleSource, result.Status)
	require.False(t, result.Verified)
	require.Zero(t, result.EffectiveScore)
}

func TestCanonicalReliabilityEnvelopeReplayKeyBindsNonceAndSequence(t *testing.T) {
	envelope, _, _ := signedReliabilityEnvelope(t)
	first, err := CanonicalReliabilityReplayKey(envelope)
	require.NoError(t, err)
	second, err := CanonicalReliabilityReplayKey(envelope)
	require.NoError(t, err)
	require.Equal(t, first, second)

	changed := cloneReliabilityEnvelope(envelope)
	changed.Sequence++
	third, err := CanonicalReliabilityReplayKey(changed)
	require.NoError(t, err)
	require.NotEqual(t, first, third)
}

func TestCanonicalReliabilityEnvelopeFixedPointGolden(t *testing.T) {
	envelope, _, _ := signedReliabilityEnvelope(t)
	score, components, err := computeCanonicalReliabilityScore(envelope.Inputs)
	require.NoError(t, err)
	require.Equal(t, int64(9380), score)
	require.Equal(t, ComponentScores{PerformanceScore: 9000, UptimeScore: 10000, ProvisioningScore: 10000, TrustScore: 8400}, components)
}

func TestCanonicalReliabilityEnvelopeRejectsAmbiguousAndOutOfBounds(t *testing.T) {
	envelope, _, _ := signedReliabilityEnvelope(t)

	unsorted := cloneReliabilityEnvelope(envelope)
	unsorted.Sources[0], unsorted.Sources[1] = unsorted.Sources[1], unsorted.Sources[0]
	_, err := CanonicalReliabilitySignBytes(unsorted)
	require.ErrorContains(t, err, "sorted")

	ambiguous := cloneReliabilityEnvelope(envelope)
	ambiguous.ClusterID = "cluster\x00other"
	_, err = CanonicalReliabilitySignBytes(ambiguous)
	require.ErrorContains(t, err, "string")

	overflow := cloneReliabilityEnvelope(envelope)
	overflow.Inputs.ProvisioningAttempts = int64(^uint64(0) >> 1)
	_, err = CanonicalReliabilitySignBytes(overflow)
	require.ErrorContains(t, err, "bounds")
}

func TestDefaultParamsKeepCanonicalReliabilityInactive(t *testing.T) {
	params := DefaultParams()
	require.False(t, params.CanonicalEnvelopeEnabled)
	require.False(t, params.VerifiedReliabilityEnabled)
	envelope, publicKey, _ := signedReliabilityEnvelope(t)
	result, err := VerifyCanonicalReliabilityEnvelope(params, envelope, publicKey, 105, 1_700_000_100)
	require.ErrorContains(t, err, "disabled")
	require.False(t, result.Verified)
	require.Zero(t, result.EffectiveScore)
}

func signedReliabilityEnvelope(t *testing.T) (CanonicalReliabilityEnvelopeV1, ed25519.PublicKey, Params) {
	t.Helper()
	privateKey := deterministicReliabilityKey(t)
	publicKey := privateKey.Public().(ed25519.PublicKey)
	digest := func(value string) []byte {
		sum := sha256.Sum256([]byte(value))
		return append([]byte(nil), sum[:]...)
	}
	sources := make([]ReliabilitySourceCommitment, 0, len(requiredReliabilitySources))
	for index, kind := range requiredReliabilitySources {
		sources = append(sources, ReliabilitySourceCommitment{
			Kind: kind, Digest: digest(string(kind)), WindowStartUnix: 1_699_999_000,
			WindowEndUnix: 1_700_000_000, ObservedThrough: 1_700_000_000 + int64(index), RecordCount: uint64(index + 1),
		})
	}
	envelope := CanonicalReliabilityEnvelopeV1{
		Version: CanonicalReliabilityEnvelopeVersion, Domain: CanonicalReliabilityDomain,
		ChainID: "virtengine-1", ProviderAddress: "provider-1", ClusterID: "cluster-1", ReportID: "report-1",
		HardwareClass: "gpu-a100", HardwareManifestDigest: digest("hardware"), SuiteVersion: "suite/v1",
		SuiteDigest: digest("suite"), WorkloadImageDigest: digest("image"), ResultDigest: digest("result"),
		Inputs: ReliabilityScoreInputs{
			BenchmarkSummary: 9000, ProvisioningSuccessRate: FixedPointScale, ProvisioningAttempts: 100,
			ProvisioningSuccesses: 100, MeanTimeToProvision: 200, MeanTimeBetweenFailures: 3_600_000,
			TotalUptimeSeconds: 99_000, TotalDowntimeSeconds: 1_000,
			DisputeCount: 2, DisputesResolved: 0, DisputesLost: 2, AnomalyFlagCount: 2,
		},
		RunnerID: "runner-1", ProviderKeyID: "benchmark-key-1", ProviderKeyEpoch: 7, Sequence: 42,
		Nonce: digest("nonce"), ObservationUnix: 1_700_000_000, IssuedAtHeight: 100, ExpiresAtHeight: 120,
		IssuedAtUnix: 1_700_000_000, ExpiresAtUnix: 1_700_000_600, FreshnessPolicyVersion: 1, Sources: sources,
	}
	signReliabilityEnvelope(t, &envelope, privateKey)
	params := DefaultParams()
	params.CanonicalEnvelopeEnabled, params.VerifiedReliabilityEnabled = true, true
	return envelope, publicKey, params
}

func deterministicReliabilityKey(t *testing.T) ed25519.PrivateKey {
	t.Helper()
	seed := sha256.Sum256([]byte("virtengine-t3-06-golden-key"))
	return ed25519.NewKeyFromSeed(seed[:])
}

func signReliabilityEnvelope(t *testing.T, envelope *CanonicalReliabilityEnvelopeV1, privateKey ed25519.PrivateKey) {
	t.Helper()
	digest, err := CanonicalReliabilityDigest(*envelope)
	require.NoError(t, err)
	envelope.Signature = ed25519.Sign(privateKey, digest)
}

func cloneReliabilityEnvelope(envelope CanonicalReliabilityEnvelopeV1) CanonicalReliabilityEnvelopeV1 {
	clone := envelope
	clone.HardwareManifestDigest = append([]byte(nil), envelope.HardwareManifestDigest...)
	clone.SuiteDigest = append([]byte(nil), envelope.SuiteDigest...)
	clone.WorkloadImageDigest = append([]byte(nil), envelope.WorkloadImageDigest...)
	clone.ResultDigest = append([]byte(nil), envelope.ResultDigest...)
	clone.Nonce = append([]byte(nil), envelope.Nonce...)
	clone.Signature = append([]byte(nil), envelope.Signature...)
	clone.Sources = append([]ReliabilitySourceCommitment(nil), envelope.Sources...)
	for index := range clone.Sources {
		clone.Sources[index].Digest = append([]byte(nil), envelope.Sources[index].Digest...)
	}
	return clone
}
