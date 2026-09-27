package types

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/stretchr/testify/require"
)

// Claim keys for the order-independence test below. The country and score
// claims reuse the exported attribute keys so the fixture cannot drift from
// the identifiers the keeper actually records.
const (
	claimKeyAgeOver = "age_over"
	claimKeyZLast   = "z_last"
	claimKeyAFirst  = "a_first"
	claimKeyNested  = "nested"
)

// These tests exercise the ZK proof primitives in proof_types.go: Pedersen
// commitments, Schnorr knowledge proofs, range proofs, set membership proofs
// and their wire encodings.
//
// The properties asserted here are the ones the chain actually depends on:
// determinism (identical inputs must produce identical bytes on every node),
// domain separation (a label must not be interchangeable), binding (a tampered
// proof must not verify), and robustness against malformed wire input (proofs
// arrive from untrusted peers, so a bad payload must return an error rather
// than panic or over-allocate).

func mustFr(t *testing.T, v uint64) fr.Element {
	t.Helper()
	var e fr.Element
	e.SetUint64(v)
	return e
}

func TestDeriveBlindDeterminism(t *testing.T) {
	salt := []byte("veid:test-salt")

	first := DeriveBlind("veid:age", salt)
	for i := 0; i < 8; i++ {
		got := DeriveBlind("veid:age", salt)
		require.True(t, got.Equal(&first),
			"DeriveBlind must be deterministic (iteration %d)", i)
	}

	otherLabel := DeriveBlind("veid:residency", salt)
	require.False(t, otherLabel.Equal(&first),
		"the label must domain-separate the derived blind")

	otherSalt := DeriveBlind("veid:age", []byte("veid:other-salt"))
	require.False(t, otherSalt.Equal(&first),
		"a different salt must produce a different blind")

	require.NotPanics(t, func() { _ = DeriveBlind("veid:age", nil) },
		"a nil salt is a valid input")
}

func TestCommitScalarDeterministicAndBinding(t *testing.T) {
	value := mustFr(t, 42)
	blind := mustFr(t, 7)

	c1, err := CommitScalar(value, blind)
	require.NoError(t, err)
	require.Len(t, c1, 64, "an uncompressed BN254 G1 point is 64 bytes")

	c2, err := CommitScalar(value, blind)
	require.NoError(t, err)
	require.Equal(t, c1, c2, "CommitScalar must be deterministic")

	c3, err := CommitScalar(mustFr(t, 43), blind)
	require.NoError(t, err)
	require.NotEqual(t, c1, c3, "changing the value must change the commitment")

	c4, err := CommitScalar(value, mustFr(t, 8))
	require.NoError(t, err)
	require.NotEqual(t, c1, c4, "changing the blind must change the commitment")

	_, err = CommitScalar(mustFr(t, 0), mustFr(t, 0))
	require.NoError(t, err, "the zero value with a zero blind must be accepted")
}

func TestHashClaimsToScalarOrderIndependence(t *testing.T) {
	nested := map[string]interface{}{"a": 1, "b": 2}
	claims := map[string]interface{}{
		claimKeyAgeOver:     18,
		AttributeKeyCountry: "AU",
		AttributeKeyScore:   720,
		claimKeyZLast:       "x",
		claimKeyAFirst:      true,
		claimKeyNested:      nested,
	}

	baseline := HashClaimsToScalar(claims)
	// Go randomises map iteration order per range statement, so repeated calls
	// on the same map are a genuine probe for order dependence.
	for i := 0; i < 50; i++ {
		got := HashClaimsToScalar(claims)
		require.True(t, got.Equal(&baseline),
			"HashClaimsToScalar must not depend on map iteration order (iteration %d)", i)
	}

	reordered := make(map[string]interface{}, len(claims))
	for _, k := range []string{claimKeyZLast, AttributeKeyScore, claimKeyNested, AttributeKeyCountry, claimKeyAFirst, claimKeyAgeOver} {
		reordered[k] = claims[k]
	}
	fromReordered := HashClaimsToScalar(reordered)
	require.True(t, fromReordered.Equal(&baseline),
		"insertion order must not change the digest")

	changed := map[string]interface{}{
		claimKeyAgeOver:     21,
		AttributeKeyCountry: "AU",
		AttributeKeyScore:   720,
		claimKeyZLast:       "x",
		claimKeyAFirst:      true,
		claimKeyNested:      nested,
	}
	fromChanged := HashClaimsToScalar(changed)
	require.False(t, fromChanged.Equal(&baseline),
		"a changed claim value must change the digest")

	require.NotPanics(t, func() { _ = HashClaimsToScalar(map[string]interface{}{}) },
		"an empty claim set must be well-defined")
}

func TestComputeCommitmentHashBinding(t *testing.T) {
	salt := []byte("veid:commitment-salt")

	h1, err := ComputeCommitmentHash("value1", salt)
	require.NoError(t, err)
	require.Len(t, h1, 64, "an uncompressed BN254 G1 point is 64 bytes")

	for i := 0; i < 5; i++ {
		got, err := ComputeCommitmentHash("value1", salt)
		require.NoError(t, err)
		require.Equal(t, h1, got, "ComputeCommitmentHash must be deterministic")
	}

	h2, err := ComputeCommitmentHash("value2", salt)
	require.NoError(t, err)
	require.NotEqual(t, h1, h2, "a different value must give a different commitment")

	h3, err := ComputeCommitmentHash("value1", []byte("other-salt"))
	require.NoError(t, err)
	require.NotEqual(t, h1, h3, "a different salt must give a different commitment")

	_, err = ComputeCommitmentHash("value1", nil)
	require.NoError(t, err, "a nil salt is a valid input")
}

func TestPedersenKnowledgeProofRoundTrip(t *testing.T) {
	value := mustFr(t, 1234)
	blind := DeriveBlind("veid:pk", []byte("salt"))

	commitment, err := CommitScalar(value, blind)
	require.NoError(t, err)

	proof, err := GeneratePedersenKnowledgeProof(commitment, value, blind, []byte("nonce"), "veid:pk")
	require.NoError(t, err)

	ok, err := VerifyPedersenKnowledgeProof(commitment, proof, "veid:pk")
	require.NoError(t, err)
	require.True(t, ok, "a well-formed knowledge proof must verify")

	ok, err = VerifyPedersenKnowledgeProof(commitment, proof, "veid:other")
	require.NoError(t, err)
	require.False(t, ok, "a different label must not verify")

	tampered := proof
	tampered.Z1 = append([]byte(nil), proof.Z1...)
	tampered.Z1[0] ^= 0xFF
	ok, err = VerifyPedersenKnowledgeProof(commitment, tampered, "veid:pk")
	require.NoError(t, err)
	require.False(t, ok, "a tampered response scalar must not verify")

	otherCommitment, err := CommitScalar(mustFr(t, 999), blind)
	require.NoError(t, err)
	ok, err = VerifyPedersenKnowledgeProof(otherCommitment, proof, "veid:pk")
	require.NoError(t, err)
	require.False(t, ok, "a proof must not verify against another commitment")

	ok, err = VerifyPedersenKnowledgeProof(nil, proof, "veid:pk")
	require.Error(t, err, "empty commitment bytes must be rejected")
	require.False(t, ok)

	ok, err = VerifyPedersenKnowledgeProof(commitment, PedersenKnowledgeProof{}, "veid:pk")
	require.Error(t, err, "an empty proof must be rejected")
	require.False(t, ok)

	raw, err := MarshalPedersenKnowledgeProof(proof)
	require.NoError(t, err)
	parsed, err := UnmarshalPedersenKnowledgeProof(raw)
	require.NoError(t, err)
	require.Equal(t, proof, parsed)
	ok, err = VerifyPedersenKnowledgeProof(commitment, parsed, "veid:pk")
	require.NoError(t, err)
	require.True(t, ok, "an unmarshalled proof must still verify")
}

func TestRangeProofHappyPathAndTampering(t *testing.T) {
	const lower, bits = uint64(18), uint8(8)
	const value = uint64(100) // adjusted = 82, which fits in 8 bits

	proof, err := GenerateRangeProof(value, lower, bits, []byte("range-salt"), []byte("range-nonce"), "veid:age")
	require.NoError(t, err)
	require.Equal(t, lower, proof.LowerBound)
	require.Equal(t, bits, proof.BitLength)
	require.Len(t, proof.BitCommitments, int(bits))
	require.Len(t, proof.BitProofs, int(bits))

	ok, err := VerifyRangeProof(proof, lower, bits, "veid:age")
	require.NoError(t, err)
	require.True(t, ok, "a well-formed range proof must verify")

	ok, err = VerifyRangeProof(proof, lower, 0, "veid:age")
	require.NoError(t, err)
	require.True(t, ok, "an expected bit length of 0 means the check is skipped")

	ok, err = VerifyRangeProof(proof, lower, bits, "veid:other")
	require.NoError(t, err)
	require.False(t, ok, "a different label must not verify")

	// A bit commitment is an uncompressed BN254 G1 point, so flipping its first
	// byte corrupts the field encoding and is rejected at parse time. The
	// verifier must surface that as an error (not a panic, and not a silent pass).
	tamperedBytes := proof
	tamperedBytes.BitCommitments = append([][]byte(nil), proof.BitCommitments...)
	tamperedBytes.BitCommitments[0] = append([]byte(nil), proof.BitCommitments[0]...)
	tamperedBytes.BitCommitments[0][0] ^= 0xFF
	ok, err = VerifyRangeProof(tamperedBytes, lower, bits, "veid:age")
	require.Error(t, err, "a corrupted bit commitment must be rejected")
	require.False(t, ok)

	// Corrupting a bit commitment that still parses must fail verification
	// rather than pass: pick the point's own Proof field instead, which is a
	// well-formed-length scalar, and flip a value bit.
	tamperedProof := proof
	tamperedProof.BitProofs = append([]BitProof(nil), proof.BitProofs...)
	tamperedProof.BitProofs[0].Z1 = append([]byte(nil), proof.BitProofs[0].Z1...)
	tamperedProof.BitProofs[0].Z1[0] ^= 0x01
	ok, err = VerifyRangeProof(tamperedProof, lower, bits, "veid:age")
	if err == nil {
		require.False(t, ok, "a tampered bit proof must not verify")
	} else {
		require.False(t, ok)
	}

	tamperedConsistency := proof
	tamperedConsistency.ConsistencyProof.Z = append([]byte(nil), proof.ConsistencyProof.Z...)
	tamperedConsistency.ConsistencyProof.Z[0] ^= 0xFF
	ok, err = VerifyRangeProof(tamperedConsistency, lower, bits, "veid:age")
	require.NoError(t, err)
	require.False(t, ok, "a tampered consistency proof must not verify")
}

func TestGenerateRangeProofInputValidation(t *testing.T) {
	_, err := GenerateRangeProof(10, 0, 0, nil, nil, "veid:age")
	require.Error(t, err, "a bit length of 0 must be rejected")

	_, err = GenerateRangeProof(5, 10, 8, nil, nil, "veid:age")
	require.Error(t, err, "a value below the lower bound must be rejected")

	_, err = GenerateRangeProof(10, 0, 64, nil, nil, "veid:age")
	require.Error(t, err, "a bit length above 63 must be rejected")
}

func TestVerifyRangeProofValidation(t *testing.T) {
	proof, err := GenerateRangeProof(100, 18, 8, []byte("s"), []byte("n"), "veid:age")
	require.NoError(t, err)

	ok, err := VerifyRangeProof(RangeProof{}, 0, 0, "veid:age")
	require.Error(t, err, "an empty range proof must be rejected")
	require.False(t, ok)

	overLong := proof
	overLong.BitLength = 64
	ok, err = VerifyRangeProof(overLong, 18, 0, "veid:age")
	require.Error(t, err, "a bit length above 63 must be rejected")
	require.False(t, ok)

	ok, err = VerifyRangeProof(proof, 18, 16, "veid:age")
	require.Error(t, err, "an expected bit length mismatch must be rejected")
	require.False(t, ok)

	ok, err = VerifyRangeProof(proof, 19, 8, "veid:age")
	require.Error(t, err, "a lower bound mismatch must be rejected")
	require.False(t, ok)

	mismatched := proof
	mismatched.BitProofs = proof.BitProofs[:len(proof.BitProofs)-1]
	ok, err = VerifyRangeProof(mismatched, 18, 8, "veid:age")
	require.Error(t, err, "a bit commitment/proof count mismatch must be rejected")
	require.False(t, ok)
}

func TestSetMembershipProofHappyPathAndTampering(t *testing.T) {
	allowed := []string{"US", "AU", "DE"}

	proof, err := GenerateSetMembershipProof("AU", allowed, []byte("set-salt"), []byte("set-nonce"), "veid:residency")
	require.NoError(t, err)
	require.Len(t, proof.E, len(allowed))
	require.Len(t, proof.Z, len(allowed))

	ok, err := VerifySetMembershipProof(proof, allowed, "veid:residency")
	require.NoError(t, err)
	require.True(t, ok, "membership in the allowed set must verify")

	ok, err = VerifySetMembershipProof(proof, allowed, "veid:other")
	require.NoError(t, err)
	require.False(t, ok, "a different label must not verify")

	tampered := proof
	tampered.Z = append([][]byte(nil), proof.Z...)
	tampered.Z[0] = append([]byte(nil), proof.Z[0]...)
	tampered.Z[0][0] ^= 0xFF
	ok, err = VerifySetMembershipProof(tampered, allowed, "veid:residency")
	require.NoError(t, err)
	require.False(t, ok, "a tampered response scalar must not verify")
}

func TestSetMembershipProofInputValidation(t *testing.T) {
	_, err := GenerateSetMembershipProof("US", nil, nil, nil, "veid:residency")
	require.Error(t, err, "an empty allowed set must be rejected")

	_, err = GenerateSetMembershipProof("FR", []string{"US", "AU"}, nil, nil, "veid:residency")
	require.Error(t, err, "a value outside the allowed set must be rejected")

	allowed := []string{"US", "AU"}
	proof, err := GenerateSetMembershipProof("US", allowed, nil, nil, "veid:residency")
	require.NoError(t, err)

	ok, err := VerifySetMembershipProof(proof, nil, "veid:residency")
	require.Error(t, err, "an empty allowed set must be rejected")
	require.False(t, ok)

	ok, err = VerifySetMembershipProof(proof, []string{"US"}, "veid:residency")
	require.Error(t, err, "a proof/allowed-set length mismatch must be rejected")
	require.False(t, ok)
}

func TestProofWireFormatRoundTrip(t *testing.T) {
	rangeProof, err := GenerateRangeProof(100, 18, 8, []byte("s"), []byte("n"), "veid:age")
	require.NoError(t, err)

	raw, err := MarshalRangeProof(rangeProof)
	require.NoError(t, err)
	require.NotEmpty(t, raw)
	require.Equal(t, byte(proofVersionV1), raw[0],
		"the wire format must start with the version tag")

	rawAgain, err := MarshalRangeProof(rangeProof)
	require.NoError(t, err)
	require.Equal(t, raw, rawAgain, "marshalling must be byte-deterministic")

	parsedRange, err := UnmarshalRangeProof(raw)
	require.NoError(t, err)
	require.Equal(t, rangeProof.LowerBound, parsedRange.LowerBound)
	require.Equal(t, rangeProof.BitLength, parsedRange.BitLength)
	require.Equal(t, rangeProof.Commitment, parsedRange.Commitment)
	require.Equal(t, rangeProof.BitCommitments, parsedRange.BitCommitments)
	require.Len(t, parsedRange.BitProofs, len(rangeProof.BitProofs))

	ok, err := VerifyRangeProof(parsedRange, 18, 8, "veid:age")
	require.NoError(t, err)
	require.True(t, ok, "a round-tripped range proof must still verify")

	allowed := []string{"US", "AU", "DE"}
	setProof, err := GenerateSetMembershipProof("AU", allowed, []byte("s"), []byte("n"), "veid:residency")
	require.NoError(t, err)

	setRaw, err := MarshalSetMembershipProof(setProof)
	require.NoError(t, err)
	setRawAgain, err := MarshalSetMembershipProof(setProof)
	require.NoError(t, err)
	require.Equal(t, setRaw, setRawAgain, "marshalling must be byte-deterministic")

	parsedSet, err := UnmarshalSetMembershipProof(setRaw)
	require.NoError(t, err)
	require.Equal(t, setProof, parsedSet)

	ok, err = VerifySetMembershipProof(parsedSet, allowed, "veid:residency")
	require.NoError(t, err)
	require.True(t, ok, "a round-tripped set membership proof must still verify")
}

func TestUnmarshalProofRejectsMalformedInput(t *testing.T) {
	rangeProof, err := GenerateRangeProof(100, 18, 8, []byte("s"), []byte("n"), "veid:age")
	require.NoError(t, err)
	rangeRaw, err := MarshalRangeProof(rangeProof)
	require.NoError(t, err)

	setProof, err := GenerateSetMembershipProof("AU", []string{"US", "AU"}, []byte("s"), []byte("n"), "veid:residency")
	require.NoError(t, err)
	setRaw, err := MarshalSetMembershipProof(setProof)
	require.NoError(t, err)

	pkValue := mustFr(t, 7)
	pkBlind := DeriveBlind("veid:pk-wire", []byte("salt"))
	pkCommitment, err := CommitScalar(pkValue, pkBlind)
	require.NoError(t, err)
	pkProof, err := GeneratePedersenKnowledgeProof(pkCommitment, pkValue, pkBlind, []byte("nonce"), "veid:pk-wire")
	require.NoError(t, err)
	pkRaw, err := MarshalPedersenKnowledgeProof(pkProof)
	require.NoError(t, err)

	type wireCase struct {
		name string
		raw  []byte
		fn   func([]byte) error
	}
	cases := []wireCase{
		{"range", rangeRaw, func(b []byte) error { _, err := UnmarshalRangeProof(b); return err }},
		{"set", setRaw, func(b []byte) error { _, err := UnmarshalSetMembershipProof(b); return err }},
		{"pedersen", pkRaw, func(b []byte) error { _, err := UnmarshalPedersenKnowledgeProof(b); return err }},
	}

	for _, c := range cases {
		require.NotEmpty(t, c.raw, "precondition: %s fixture must marshal", c.name)

		t.Run(c.name+"/empty", func(t *testing.T) {
			require.Error(t, c.fn(nil), "empty input must be rejected")
		})

		t.Run(c.name+"/unknown-version", func(t *testing.T) {
			require.Error(t, c.fn([]byte{0x7F}), "an unknown version tag must be rejected")
		})

		t.Run(c.name+"/version-without-body", func(t *testing.T) {
			require.Error(t, c.fn([]byte{proofVersionV1}), "a bare version tag must be rejected")
		})

		t.Run(c.name+"/truncated-never-panics", func(t *testing.T) {
			for i := 0; i < len(c.raw); i++ {
				require.NotPanics(t, func() { _ = c.fn(c.raw[:i]) },
					"truncation at byte %d must not panic", i)
			}
		})

		t.Run(c.name+"/trailing-garbage-never-panics", func(t *testing.T) {
			withTrailer := append(append([]byte(nil), c.raw...), 0xAB, 0xCD)
			require.NotPanics(t, func() { _ = c.fn(withTrailer) })
		})
	}
}
