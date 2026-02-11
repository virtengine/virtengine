package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

const (
	proofVersionV1             = byte(1)
	pedersenDomainH            = "veid:pedersen:h"
	transcriptDomainCommitment = "veid:commitment"
	transcriptDomainRange      = "veid:rangeproof"
	transcriptDomainSet        = "veid:setproof"
)

// PedersenCommitment represents a commitment point on BN254 G1.
type PedersenCommitment struct {
	Point []byte `json:"point"`
}

// PedersenKnowledgeProof proves knowledge of (value, blind) for a Pedersen commitment.
type PedersenKnowledgeProof struct {
	R  []byte `json:"r"`
	Z1 []byte `json:"z1"`
	Z2 []byte `json:"z2"`
}

// SchnorrProof proves knowledge of a scalar for a single-base commitment.
type SchnorrProof struct {
	R []byte `json:"r"`
	Z []byte `json:"z"`
}

// BitProof proves that a Pedersen commitment contains a bit (0 or 1).
type BitProof struct {
	E0 []byte `json:"e0"`
	E1 []byte `json:"e1"`
	Z0 []byte `json:"z0"`
	Z1 []byte `json:"z1"`
}

// RangeProof proves that a committed value is >= LowerBound and within the bit length range.
type RangeProof struct {
	Commitment       []byte     `json:"commitment"`
	BitCommitments   [][]byte   `json:"bit_commitments"`
	BitProofs        []BitProof `json:"bit_proofs"`
	ConsistencyProof SchnorrProof
	LowerBound       uint64 `json:"lower_bound"`
	BitLength        uint8  `json:"bit_length"`
}

// SetMembershipProof proves that a committed value is in the allowed set.
type SetMembershipProof struct {
	Commitment []byte   `json:"commitment"`
	E          [][]byte `json:"e"`
	Z          [][]byte `json:"z"`
}

// ProofRevocation records a revoked proof.
type ProofRevocation struct {
	ProofID   string `json:"proof_id"`
	RevokedAt int64  `json:"revoked_at"`
	Reason    string `json:"reason"`
}

var (
	pedersenOnce sync.Once
	pedersenH    bn254.G1Affine
	pedersenErr  error
)

func pedersenGenerators() (bn254.G1Affine, error) {
	pedersenOnce.Do(func() {
		var g bn254.G1Affine
		g.ScalarMultiplicationBase(big.NewInt(1))

		h, err := bn254.HashToG1([]byte(pedersenDomainH), []byte(pedersenDomainH))
		if err != nil {
			pedersenErr = err
			return
		}
		if h.Equal(&g) {
			h, err = bn254.HashToG1([]byte(pedersenDomainH+":alt"), []byte(pedersenDomainH))
			if err != nil {
				pedersenErr = err
				return
			}
		}
		pedersenH = h
	})
	return pedersenH, pedersenErr
}

func commitmentToBytes(p *bn254.G1Affine) []byte {
	return p.Marshal()
}

func commitmentFromBytes(b []byte) (bn254.G1Affine, error) {
	var p bn254.G1Affine
	if len(b) == 0 {
		return p, errors.New("empty commitment bytes")
	}
	if err := p.Unmarshal(b); err != nil {
		return p, err
	}
	return p, nil
}

func hashToScalar(label string, parts ...[]byte) fr.Element {
	h := sha256.New()
	h.Write([]byte(label))
	for _, p := range parts {
		if len(p) == 0 {
			continue
		}
		h.Write(p)
	}
	sum := h.Sum(nil)
	var e fr.Element
	e.SetBytes(sum)
	return e
}

// DeriveBlind derives a scalar blinding factor from salt bytes.
func DeriveBlind(label string, salt []byte) fr.Element {
	return hashToScalar(label+":blind", salt)
}

func frToBytes(e fr.Element) []byte {
	b := e.Bytes()
	out := make([]byte, len(b))
	copy(out, b[:])
	return out
}

func frFromBytes(b []byte) (fr.Element, error) {
	if len(b) != fr.Bytes {
		return fr.Element{}, fmt.Errorf("invalid scalar length: %d", len(b))
	}
	var e fr.Element
	e.SetBytes(b)
	return e, nil
}

func scalarBigInt(e fr.Element) *big.Int {
	return e.BigInt(new(big.Int))
}

// CommitScalar creates a Pedersen commitment for a scalar value.
func CommitScalar(value fr.Element, blind fr.Element) ([]byte, error) {
	commitment, err := pedersenCommitScalar(value, blind)
	if err != nil {
		return nil, err
	}
	return commitmentToBytes(&commitment), nil
}

// HashClaimsToScalar hashes claims deterministically and returns a scalar.
func HashClaimsToScalar(claims map[string]interface{}) fr.Element {
	keys := make([]string, 0, len(claims))
	for k := range claims {
		keys = append(keys, k)
	}
	sortStrings(keys)

	h := sha256.New()
	for _, key := range keys {
		h.Write([]byte(key))
		fmt.Fprintf(h, "%v", claims[key])
	}
	var e fr.Element
	e.SetBytes(h.Sum(nil))
	return e
}

// ComputeCommitmentHash computes a Pedersen commitment to the given value.
func ComputeCommitmentHash(value interface{}, salt []byte) ([]byte, error) {
	scalar := hashToScalar(transcriptDomainCommitment, []byte(fmt.Sprintf("%v", value)))
	blind := hashToScalar(transcriptDomainCommitment+":blind", salt)
	commitment, err := pedersenCommitScalar(scalar, blind)
	if err != nil {
		return nil, err
	}
	return commitmentToBytes(&commitment), nil
}

// GeneratePedersenKnowledgeProof proves knowledge of the commitment opening.
func GeneratePedersenKnowledgeProof(commitmentBytes []byte, value fr.Element, blind fr.Element, nonce []byte, label string) (PedersenKnowledgeProof, error) {
	h, err := pedersenGenerators()
	if err != nil {
		return PedersenKnowledgeProof{}, err
	}
	commitment, err := commitmentFromBytes(commitmentBytes)
	if err != nil {
		return PedersenKnowledgeProof{}, err
	}

	k1 := hashToScalar(label+":k1", nonce, frToBytes(value), frToBytes(blind))
	k2 := hashToScalar(label+":k2", nonce, frToBytes(value), frToBytes(blind))

	var r1, r2, r bn254.G1Affine
	r1.ScalarMultiplicationBase(scalarBigInt(k1))
	r2.ScalarMultiplication(&h, scalarBigInt(k2))
	r.Add(&r1, &r2)

	e := hashToScalar(label+":challenge", commitmentToBytes(&commitment), commitmentToBytes(&r))

	var z1, z2 fr.Element
	z1.Mul(&e, &value).Add(&z1, &k1)
	z2.Mul(&e, &blind).Add(&z2, &k2)

	return PedersenKnowledgeProof{
		R:  commitmentToBytes(&r),
		Z1: frToBytes(z1),
		Z2: frToBytes(z2),
	}, nil
}

// VerifyPedersenKnowledgeProof verifies the knowledge proof for a Pedersen commitment.
func VerifyPedersenKnowledgeProof(commitmentBytes []byte, proof PedersenKnowledgeProof, label string) (bool, error) {
	h, err := pedersenGenerators()
	if err != nil {
		return false, err
	}
	commitment, err := commitmentFromBytes(commitmentBytes)
	if err != nil {
		return false, err
	}
	rPoint, err := commitmentFromBytes(proof.R)
	if err != nil {
		return false, err
	}
	z1, err := frFromBytes(proof.Z1)
	if err != nil {
		return false, err
	}
	z2, err := frFromBytes(proof.Z2)
	if err != nil {
		return false, err
	}

	e := hashToScalar(label+":challenge", commitmentToBytes(&commitment), commitmentToBytes(&rPoint))

	var left, right, eg, eh bn254.G1Affine
	left.ScalarMultiplicationBase(scalarBigInt(z1))
	eh.ScalarMultiplication(&h, scalarBigInt(z2))
	left.Add(&left, &eh)

	eg.ScalarMultiplication(&commitment, scalarBigInt(e))
	right.Add(&rPoint, &eg)

	return left.Equal(&right), nil
}

// GenerateRangeProof creates a range proof for value >= lowerBound.
func GenerateRangeProof(value uint64, lowerBound uint64, bitLength uint8, commitmentSalt []byte, nonce []byte, label string) (RangeProof, error) {
	if bitLength == 0 {
		return RangeProof{}, errors.New("bit length must be positive")
	}
	if value < lowerBound {
		return RangeProof{}, fmt.Errorf("value %d below lower bound %d", value, lowerBound)
	}

	h, err := pedersenGenerators()
	if err != nil {
		return RangeProof{}, err
	}
	_ = h

	blind := hashToScalar(label+":blind", commitmentSalt)
	valueScalar := fr.Element{}
	valueScalar.SetUint64(value)

	commitmentPoint, err := pedersenCommitScalar(valueScalar, blind)
	if err != nil {
		return RangeProof{}, err
	}

	adjusted := value - lowerBound

	bitCommitments := make([][]byte, 0, bitLength)
	bitProofs := make([]BitProof, 0, bitLength)
	bitBlinds := make([]fr.Element, 0, bitLength)

	for i := uint8(0); i < bitLength; i++ {
		bit := (adjusted >> i) & 1
		bitBlind := hashToScalar(label+":bitblind", nonce, []byte{byte(i)})
		bitBlinds = append(bitBlinds, bitBlind)

		var bitVal fr.Element
		bitVal.SetUint64(bit)
		bitCommitment, err := pedersenCommitScalar(bitVal, bitBlind)
		if err != nil {
			return RangeProof{}, err
		}
		bitCommitments = append(bitCommitments, commitmentToBytes(&bitCommitment))

		proof, err := generateBitProof(bitCommitment, bit, bitBlind, label)
		if err != nil {
			return RangeProof{}, err
		}
		bitProofs = append(bitProofs, proof)
	}

	cPrime := commitmentPoint
	if lowerBound > 0 {
		var lbPoint bn254.G1Affine
		lbPoint.ScalarMultiplicationBase(new(big.Int).SetUint64(lowerBound))
		cPrime.Sub(&commitmentPoint, &lbPoint)
	}

	var sum bn254.G1Affine
	sum.SetInfinity()
	var blindSum fr.Element
	blindSum.SetZero()

	if bitLength > 63 {
		return RangeProof{}, errors.New("bit length too large")
	}
	for i := 0; i < len(bitCommitments); i++ {
		if i >= 64 {
			return RangeProof{}, errors.New("bit length too large")
		}
		bi, err := commitmentFromBytes(bitCommitments[i])
		if err != nil {
			return RangeProof{}, err
		}
		weight := new(big.Int).SetUint64(uint64(1) << i)
		var weighted bn254.G1Affine
		weighted.ScalarMultiplication(&bi, weight)
		sum.Add(&sum, &weighted)

		var weightFr fr.Element
		weightFr.SetUint64(uint64(1) << i)
		var term fr.Element
		term.Mul(&bitBlinds[i], &weightFr)
		blindSum.Add(&blindSum, &term)
	}

	var c0 bn254.G1Affine
	c0.Sub(&cPrime, &sum)

	var rPrime fr.Element
	rPrime.Sub(&blind, &blindSum)
	consistencyProof, err := generateSchnorrProof(c0, rPrime, label+":consistency", nonce)
	if err != nil {
		return RangeProof{}, err
	}

	return RangeProof{
		Commitment:       commitmentToBytes(&commitmentPoint),
		BitCommitments:   bitCommitments,
		BitProofs:        bitProofs,
		ConsistencyProof: consistencyProof,
		LowerBound:       lowerBound,
		BitLength:        bitLength,
	}, nil
}

// VerifyRangeProof verifies a range proof for the given lower bound and bit length.
func VerifyRangeProof(proof RangeProof, expectedLowerBound uint64, expectedBitLength uint8, label string) (bool, error) {
	if proof.BitLength == 0 || len(proof.BitCommitments) == 0 {
		return false, errors.New("empty range proof")
	}
	if proof.BitLength > 63 {
		return false, errors.New("bit length too large")
	}
	if expectedBitLength != 0 && proof.BitLength != expectedBitLength {
		return false, fmt.Errorf("bit length mismatch: %d", proof.BitLength)
	}
	if expectedLowerBound != proof.LowerBound {
		return false, fmt.Errorf("lower bound mismatch: %d", proof.LowerBound)
	}
	if len(proof.BitCommitments) != len(proof.BitProofs) {
		return false, errors.New("bit commitment/proof length mismatch")
	}

	commitment, err := commitmentFromBytes(proof.Commitment)
	if err != nil {
		return false, err
	}

	cPrime := commitment
	if proof.LowerBound > 0 {
		var lbPoint bn254.G1Affine
		lbPoint.ScalarMultiplicationBase(new(big.Int).SetUint64(proof.LowerBound))
		cPrime.Sub(&commitment, &lbPoint)
	}

	var sum bn254.G1Affine
	sum.SetInfinity()

	for i := 0; i < len(proof.BitCommitments); i++ {
		if i >= 64 {
			return false, errors.New("bit length too large")
		}
		bitCommit, err := commitmentFromBytes(proof.BitCommitments[i])
		if err != nil {
			return false, err
		}
		if ok, err := verifyBitProof(bitCommit, proof.BitProofs[i], label); err != nil || !ok {
			if err != nil {
				return false, err
			}
			return false, nil
		}
		weight := new(big.Int).SetUint64(uint64(1) << i)
		var weighted bn254.G1Affine
		weighted.ScalarMultiplication(&bitCommit, weight)
		sum.Add(&sum, &weighted)
	}

	var c0 bn254.G1Affine
	c0.Sub(&cPrime, &sum)

	if ok, err := verifySchnorrProof(c0, proof.ConsistencyProof, label+":consistency"); err != nil || !ok {
		if err != nil {
			return false, err
		}
		return false, nil
	}

	return true, nil
}

// GenerateSetMembershipProof proves that value is in allowed set.
func GenerateSetMembershipProof(value string, allowed []string, commitmentSalt []byte, nonce []byte, label string) (SetMembershipProof, error) {
	if len(allowed) == 0 {
		return SetMembershipProof{}, errors.New("allowed set cannot be empty")
	}

	h, err := pedersenGenerators()
	if err != nil {
		return SetMembershipProof{}, err
	}

	blind := hashToScalar(label+":blind", commitmentSalt)
	valueScalar := hashToScalar(label+":value", []byte(value))
	commitmentPoint, err := pedersenCommitScalar(valueScalar, blind)
	if err != nil {
		return SetMembershipProof{}, err
	}

	index := -1
	allowedScalars := make([]fr.Element, 0, len(allowed))
	for i, val := range allowed {
		s := hashToScalar(label+":value", []byte(val))
		allowedScalars = append(allowedScalars, s)
		if val == value {
			index = i
		}
	}
	if index < 0 {
		return SetMembershipProof{}, errors.New("value not in allowed set")
	}

	eList := make([]fr.Element, len(allowedScalars))
	zList := make([]fr.Element, len(allowedScalars))

	var sumE fr.Element
	sumE.SetZero()

	for i := range allowedScalars {
		if i == index {
			continue
		}
		eList[i] = hashToScalar(label+":e", nonce, []byte{byte(i)})
		zList[i] = hashToScalar(label+":z", nonce, []byte{byte(i)})
		sumE.Add(&sumE, &eList[i])
	}

	w := hashToScalar(label+":w", nonce, []byte{byte(index)})

	aiBytes := make([][]byte, len(allowedScalars))
	for i, s := range allowedScalars {
		var ci bn254.G1Affine
		if i == index {
			var a bn254.G1Affine
			a.ScalarMultiplication(&h, scalarBigInt(w))
			aiBytes[i] = commitmentToBytes(&a)
			continue
		}
		var mi bn254.G1Affine
		mi.ScalarMultiplicationBase(scalarBigInt(s))
		ci.Sub(&commitmentPoint, &mi)
		var ziH, eiCi bn254.G1Affine
		ziH.ScalarMultiplication(&h, scalarBigInt(zList[i]))
		eiCi.ScalarMultiplication(&ci, scalarBigInt(eList[i]))
		ziH.Sub(&ziH, &eiCi)
		aiBytes[i] = commitmentToBytes(&ziH)
	}

	parts := make([][]byte, 0, len(aiBytes)+1)
	parts = append(parts, commitmentToBytes(&commitmentPoint))
	parts = append(parts, aiBytes...)
	e := hashToScalar(label+":challenge", parts...)

	var eIndex fr.Element
	eIndex.Sub(&e, &sumE)
	eList[index] = eIndex

	var zIndex fr.Element
	zIndex.Mul(&eIndex, &blind).Add(&zIndex, &w)
	zList[index] = zIndex

	proof := SetMembershipProof{
		Commitment: commitmentToBytes(&commitmentPoint),
		E:          make([][]byte, len(eList)),
		Z:          make([][]byte, len(zList)),
	}
	for i := range eList {
		proof.E[i] = frToBytes(eList[i])
		proof.Z[i] = frToBytes(zList[i])
	}

	return proof, nil
}

// VerifySetMembershipProof verifies set membership proof against allowed set.
func VerifySetMembershipProof(proof SetMembershipProof, allowed []string, label string) (bool, error) {
	if len(allowed) == 0 {
		return false, errors.New("allowed set cannot be empty")
	}
	if len(proof.E) != len(allowed) || len(proof.Z) != len(allowed) {
		return false, errors.New("proof length mismatch")
	}

	h, err := pedersenGenerators()
	if err != nil {
		return false, err
	}

	commitmentPoint, err := commitmentFromBytes(proof.Commitment)
	if err != nil {
		return false, err
	}

	allowedScalars := make([]fr.Element, len(allowed))
	for i, val := range allowed {
		allowedScalars[i] = hashToScalar(label+":value", []byte(val))
	}

	eList := make([]fr.Element, len(allowed))
	zList := make([]fr.Element, len(allowed))
	aiBytes := make([][]byte, len(allowed))
	var sumE fr.Element
	sumE.SetZero()

	for i := range allowedScalars {
		e, err := frFromBytes(proof.E[i])
		if err != nil {
			return false, err
		}
		z, err := frFromBytes(proof.Z[i])
		if err != nil {
			return false, err
		}
		eList[i] = e
		zList[i] = z
		sumE.Add(&sumE, &e)

		var mi bn254.G1Affine
		mi.ScalarMultiplicationBase(scalarBigInt(allowedScalars[i]))
		var ci bn254.G1Affine
		ci.Sub(&commitmentPoint, &mi)
		var ziH, eiCi bn254.G1Affine
		ziH.ScalarMultiplication(&h, scalarBigInt(z))
		eiCi.ScalarMultiplication(&ci, scalarBigInt(e))
		ziH.Sub(&ziH, &eiCi)
		aiBytes[i] = commitmentToBytes(&ziH)
	}

	parts := make([][]byte, 0, len(aiBytes)+1)
	parts = append(parts, commitmentToBytes(&commitmentPoint))
	parts = append(parts, aiBytes...)
	e := hashToScalar(label+":challenge", parts...)
	if !sumE.Equal(&e) {
		return false, nil
	}
	return true, nil
}

func pedersenCommitScalar(value fr.Element, blind fr.Element) (bn254.G1Affine, error) {
	h, err := pedersenGenerators()
	if err != nil {
		return bn254.G1Affine{}, err
	}
	var gMul, hMul, out bn254.G1Affine
	gMul.ScalarMultiplicationBase(scalarBigInt(value))
	hMul.ScalarMultiplication(&h, scalarBigInt(blind))
	out.Add(&gMul, &hMul)
	return out, nil
}

func generateSchnorrProof(public bn254.G1Affine, secret fr.Element, label string, nonce []byte) (SchnorrProof, error) {
	h, err := pedersenGenerators()
	if err != nil {
		return SchnorrProof{}, err
	}

	k := hashToScalar(label+":k", nonce, commitmentToBytes(&public))
	var rPoint bn254.G1Affine
	rPoint.ScalarMultiplication(&h, scalarBigInt(k))
	e := hashToScalar(label+":challenge", commitmentToBytes(&public), commitmentToBytes(&rPoint))

	var z fr.Element
	z.Mul(&e, &secret).Add(&z, &k)

	return SchnorrProof{
		R: commitmentToBytes(&rPoint),
		Z: frToBytes(z),
	}, nil
}

func verifySchnorrProof(public bn254.G1Affine, proof SchnorrProof, label string) (bool, error) {
	h, err := pedersenGenerators()
	if err != nil {
		return false, err
	}
	rPoint, err := commitmentFromBytes(proof.R)
	if err != nil {
		return false, err
	}
	z, err := frFromBytes(proof.Z)
	if err != nil {
		return false, err
	}

	e := hashToScalar(label+":challenge", commitmentToBytes(&public), commitmentToBytes(&rPoint))

	var left, right, eC bn254.G1Affine
	left.ScalarMultiplication(&h, scalarBigInt(z))
	eC.ScalarMultiplication(&public, scalarBigInt(e))
	right.Add(&rPoint, &eC)
	return left.Equal(&right), nil
}

func generateBitProof(commitment bn254.G1Affine, bit uint64, blind fr.Element, label string) (BitProof, error) {
	h, err := pedersenGenerators()
	if err != nil {
		return BitProof{}, err
	}

	c0 := commitment
	var gPoint bn254.G1Affine
	gPoint.ScalarMultiplicationBase(big.NewInt(1))
	var c1 bn254.G1Affine
	c1.Sub(&commitment, &gPoint)

	var e0, e1, z0, z1 fr.Element
	var a0, a1 bn254.G1Affine

	if bit == 0 {
		w := hashToScalar(label+":bit:w0", frToBytes(blind))
		e1 = hashToScalar(label+":bit:e1", frToBytes(blind))
		z1 = hashToScalar(label+":bit:z1", frToBytes(blind))

		var z1h, e1c1 bn254.G1Affine
		z1h.ScalarMultiplication(&h, scalarBigInt(z1))
		e1c1.ScalarMultiplication(&c1, scalarBigInt(e1))
		z1h.Sub(&z1h, &e1c1)
		a1 = z1h

		a0.ScalarMultiplication(&h, scalarBigInt(w))

		e := hashToScalar(label+":bit:challenge", commitmentToBytes(&commitment), commitmentToBytes(&a0), commitmentToBytes(&a1))
		e0.Sub(&e, &e1)
		z0.Mul(&e0, &blind).Add(&z0, &w)
	} else {
		w := hashToScalar(label+":bit:w1", frToBytes(blind))
		e0 = hashToScalar(label+":bit:e0", frToBytes(blind))
		z0 = hashToScalar(label+":bit:z0", frToBytes(blind))

		var z0h, e0c0 bn254.G1Affine
		z0h.ScalarMultiplication(&h, scalarBigInt(z0))
		e0c0.ScalarMultiplication(&c0, scalarBigInt(e0))
		z0h.Sub(&z0h, &e0c0)
		a0 = z0h

		a1.ScalarMultiplication(&h, scalarBigInt(w))

		e := hashToScalar(label+":bit:challenge", commitmentToBytes(&commitment), commitmentToBytes(&a0), commitmentToBytes(&a1))
		e1.Sub(&e, &e0)
		z1.Mul(&e1, &blind).Add(&z1, &w)
	}

	return BitProof{
		E0: frToBytes(e0),
		E1: frToBytes(e1),
		Z0: frToBytes(z0),
		Z1: frToBytes(z1),
	}, nil
}

func verifyBitProof(commitment bn254.G1Affine, proof BitProof, label string) (bool, error) {
	h, err := pedersenGenerators()
	if err != nil {
		return false, err
	}

	c0 := commitment
	var gPoint bn254.G1Affine
	gPoint.ScalarMultiplicationBase(big.NewInt(1))
	var c1 bn254.G1Affine
	c1.Sub(&commitment, &gPoint)

	e0, err := frFromBytes(proof.E0)
	if err != nil {
		return false, err
	}
	e1, err := frFromBytes(proof.E1)
	if err != nil {
		return false, err
	}
	z0, err := frFromBytes(proof.Z0)
	if err != nil {
		return false, err
	}
	z1, err := frFromBytes(proof.Z1)
	if err != nil {
		return false, err
	}

	var a0, a1 bn254.G1Affine
	var z0h, e0c0 bn254.G1Affine
	z0h.ScalarMultiplication(&h, scalarBigInt(z0))
	e0c0.ScalarMultiplication(&c0, scalarBigInt(e0))
	z0h.Sub(&z0h, &e0c0)
	a0 = z0h

	var z1h, e1c1 bn254.G1Affine
	z1h.ScalarMultiplication(&h, scalarBigInt(z1))
	e1c1.ScalarMultiplication(&c1, scalarBigInt(e1))
	z1h.Sub(&z1h, &e1c1)
	a1 = z1h

	e := hashToScalar(label+":bit:challenge", commitmentToBytes(&commitment), commitmentToBytes(&a0), commitmentToBytes(&a1))

	var sum fr.Element
	sum.Add(&e0, &e1)
	return sum.Equal(&e), nil
}

func sortStrings(values []string) {
	if len(values) < 2 {
		return
	}
	for i := 0; i < len(values)-1; i++ {
		for j := i + 1; j < len(values); j++ {
			if values[j] < values[i] {
				values[i], values[j] = values[j], values[i]
			}
		}
	}
}

// MarshalRangeProof serializes a range proof to bytes.
func MarshalRangeProof(proof RangeProof) ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.WriteByte(proofVersionV1)
	if err := binary.Write(buf, binary.BigEndian, proof.LowerBound); err != nil {
		return nil, err
	}
	buf.WriteByte(proof.BitLength)
	if err := writeBytes(buf, proof.Commitment); err != nil {
		return nil, err
	}
	if err := writeBytesSlice(buf, proof.BitCommitments); err != nil {
		return nil, err
	}
	if err := writeBitProofs(buf, proof.BitProofs); err != nil {
		return nil, err
	}
	if err := writeBytes(buf, proof.ConsistencyProof.R); err != nil {
		return nil, err
	}
	if err := writeBytes(buf, proof.ConsistencyProof.Z); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnmarshalRangeProof parses a range proof from bytes.
func UnmarshalRangeProof(b []byte) (RangeProof, error) {
	reader := bytes.NewReader(b)
	version, err := reader.ReadByte()
	if err != nil {
		return RangeProof{}, err
	}
	if version != proofVersionV1 {
		return RangeProof{}, fmt.Errorf("unsupported proof version: %d", version)
	}

	var lower uint64
	if err := binary.Read(reader, binary.BigEndian, &lower); err != nil {
		return RangeProof{}, err
	}
	bitLengthByte, err := reader.ReadByte()
	if err != nil {
		return RangeProof{}, err
	}

	commitment, err := readBytes(reader)
	if err != nil {
		return RangeProof{}, err
	}
	bitCommitments, err := readBytesSlice(reader)
	if err != nil {
		return RangeProof{}, err
	}
	bitProofs, err := readBitProofs(reader)
	if err != nil {
		return RangeProof{}, err
	}
	consR, err := readBytes(reader)
	if err != nil {
		return RangeProof{}, err
	}
	consZ, err := readBytes(reader)
	if err != nil {
		return RangeProof{}, err
	}

	return RangeProof{
		Commitment:       commitment,
		BitCommitments:   bitCommitments,
		BitProofs:        bitProofs,
		ConsistencyProof: SchnorrProof{R: consR, Z: consZ},
		LowerBound:       lower,
		BitLength:        bitLengthByte,
	}, nil
}

// MarshalSetMembershipProof serializes a set membership proof to bytes.
func MarshalSetMembershipProof(proof SetMembershipProof) ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.WriteByte(proofVersionV1)
	if err := writeBytes(buf, proof.Commitment); err != nil {
		return nil, err
	}
	if err := writeBytesSlice(buf, proof.E); err != nil {
		return nil, err
	}
	if err := writeBytesSlice(buf, proof.Z); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnmarshalSetMembershipProof parses a set membership proof.
func UnmarshalSetMembershipProof(b []byte) (SetMembershipProof, error) {
	reader := bytes.NewReader(b)
	version, err := reader.ReadByte()
	if err != nil {
		return SetMembershipProof{}, err
	}
	if version != proofVersionV1 {
		return SetMembershipProof{}, fmt.Errorf("unsupported proof version: %d", version)
	}
	commitment, err := readBytes(reader)
	if err != nil {
		return SetMembershipProof{}, err
	}
	eList, err := readBytesSlice(reader)
	if err != nil {
		return SetMembershipProof{}, err
	}
	zList, err := readBytesSlice(reader)
	if err != nil {
		return SetMembershipProof{}, err
	}

	return SetMembershipProof{
		Commitment: commitment,
		E:          eList,
		Z:          zList,
	}, nil
}

// MarshalPedersenKnowledgeProof serializes a Pedersen knowledge proof.
func MarshalPedersenKnowledgeProof(proof PedersenKnowledgeProof) ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.WriteByte(proofVersionV1)
	if err := writeBytes(buf, proof.R); err != nil {
		return nil, err
	}
	if err := writeBytes(buf, proof.Z1); err != nil {
		return nil, err
	}
	if err := writeBytes(buf, proof.Z2); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnmarshalPedersenKnowledgeProof parses a Pedersen knowledge proof.
func UnmarshalPedersenKnowledgeProof(b []byte) (PedersenKnowledgeProof, error) {
	reader := bytes.NewReader(b)
	version, err := reader.ReadByte()
	if err != nil {
		return PedersenKnowledgeProof{}, err
	}
	if version != proofVersionV1 {
		return PedersenKnowledgeProof{}, fmt.Errorf("unsupported proof version: %d", version)
	}
	rBytes, err := readBytes(reader)
	if err != nil {
		return PedersenKnowledgeProof{}, err
	}
	z1, err := readBytes(reader)
	if err != nil {
		return PedersenKnowledgeProof{}, err
	}
	z2, err := readBytes(reader)
	if err != nil {
		return PedersenKnowledgeProof{}, err
	}

	return PedersenKnowledgeProof{R: rBytes, Z1: z1, Z2: z2}, nil
}

const maxUint32 = int(^uint32(0))

func checkedUint32(n int) (uint32, error) {
	if n < 0 || n > maxUint32 {
		return 0, errors.New("length exceeds uint32")
	}
	return uint32(n), nil
}

func claimTypeUint32(ct ClaimType) (uint32, error) {
	if ct < 0 || int(ct) > maxUint32 {
		return 0, fmt.Errorf("invalid claim type: %d", ct)
	}
	//nolint:gosec // safe: bounds checked above
	return uint32(ct), nil
}

func writeBytes(buf *bytes.Buffer, b []byte) error {
	length, err := checkedUint32(len(b))
	if err != nil {
		return err
	}
	if err := binary.Write(buf, binary.BigEndian, length); err != nil {
		return err
	}
	if len(b) > 0 {
		_, err := buf.Write(b)
		return err
	}
	return nil
}

func writeBytesSlice(buf *bytes.Buffer, items [][]byte) error {
	length, err := checkedUint32(len(items))
	if err != nil {
		return err
	}
	if err := binary.Write(buf, binary.BigEndian, length); err != nil {
		return err
	}
	for _, item := range items {
		if err := writeBytes(buf, item); err != nil {
			return err
		}
	}
	return nil
}

func writeBitProofs(buf *bytes.Buffer, proofs []BitProof) error {
	length, err := checkedUint32(len(proofs))
	if err != nil {
		return err
	}
	if err := binary.Write(buf, binary.BigEndian, length); err != nil {
		return err
	}
	for _, proof := range proofs {
		if err := writeBytes(buf, proof.E0); err != nil {
			return err
		}
		if err := writeBytes(buf, proof.E1); err != nil {
			return err
		}
		if err := writeBytes(buf, proof.Z0); err != nil {
			return err
		}
		if err := writeBytes(buf, proof.Z1); err != nil {
			return err
		}
	}
	return nil
}

func readBytes(reader *bytes.Reader) ([]byte, error) {
	var length uint32
	if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
		return nil, err
	}
	if length == 0 {
		return nil, nil
	}
	buf := make([]byte, length)
	if _, err := reader.Read(buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func readBytesSlice(reader *bytes.Reader) ([][]byte, error) {
	var length uint32
	if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
		return nil, err
	}
	items := make([][]byte, 0, length)
	for i := uint32(0); i < length; i++ {
		item, err := readBytes(reader)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func readBitProofs(reader *bytes.Reader) ([]BitProof, error) {
	var length uint32
	if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
		return nil, err
	}
	proofs := make([]BitProof, 0, length)
	for i := uint32(0); i < length; i++ {
		e0, err := readBytes(reader)
		if err != nil {
			return nil, err
		}
		e1, err := readBytes(reader)
		if err != nil {
			return nil, err
		}
		z0, err := readBytes(reader)
		if err != nil {
			return nil, err
		}
		z1, err := readBytes(reader)
		if err != nil {
			return nil, err
		}
		proofs = append(proofs, BitProof{E0: e0, E1: e1, Z0: z0, Z1: z1})
	}
	return proofs, nil
}

// ============================================================================
// Selective Disclosure Proof Bundles
// ============================================================================

// ClaimProofKind identifies the proof type used for a claim.
type ClaimProofKind uint8

const (
	ClaimProofKindUnknown ClaimProofKind = iota
	ClaimProofKindRange
	ClaimProofKindSetMembership
	ClaimProofKindPedersenKnowledge
)

// ClaimProofEntry stores the serialized proof for a claim.
type ClaimProofEntry struct {
	ClaimType  ClaimType      `json:"claim_type"`
	ProofKind  ClaimProofKind `json:"proof_kind"`
	Commitment []byte         `json:"commitment"`
	Proof      []byte         `json:"proof"`
}

// SelectiveDisclosureProofBundle aggregates claim proofs in a deterministic order.
type SelectiveDisclosureProofBundle struct {
	Version uint8             `json:"version"`
	Proofs  []ClaimProofEntry `json:"proofs"`
}

const selectiveDisclosureBundleVersion = uint8(1)

// MarshalSelectiveDisclosureProofBundle serializes the bundle to bytes.
func MarshalSelectiveDisclosureProofBundle(bundle SelectiveDisclosureProofBundle) ([]byte, error) {
	buf := &bytes.Buffer{}
	version := bundle.Version
	if version == 0 {
		version = selectiveDisclosureBundleVersion
	}
	buf.WriteByte(version)

	count, err := checkedUint32(len(bundle.Proofs))
	if err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, count); err != nil {
		return nil, err
	}
	for _, entry := range bundle.Proofs {
		ct, err := claimTypeUint32(entry.ClaimType)
		if err != nil {
			return nil, err
		}
		if err := binary.Write(buf, binary.BigEndian, ct); err != nil {
			return nil, err
		}
		buf.WriteByte(byte(entry.ProofKind))
		if err := writeBytes(buf, entry.Commitment); err != nil {
			return nil, err
		}
		if err := writeBytes(buf, entry.Proof); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

// UnmarshalSelectiveDisclosureProofBundle parses a bundle from bytes.
func UnmarshalSelectiveDisclosureProofBundle(b []byte) (SelectiveDisclosureProofBundle, error) {
	reader := bytes.NewReader(b)
	version, err := reader.ReadByte()
	if err != nil {
		return SelectiveDisclosureProofBundle{}, err
	}
	if version != selectiveDisclosureBundleVersion {
		return SelectiveDisclosureProofBundle{}, fmt.Errorf("unsupported bundle version: %d", version)
	}

	var count uint32
	if err := binary.Read(reader, binary.BigEndian, &count); err != nil {
		return SelectiveDisclosureProofBundle{}, err
	}

	proofs := make([]ClaimProofEntry, 0, count)
	for i := uint32(0); i < count; i++ {
		var ct uint32
		if err := binary.Read(reader, binary.BigEndian, &ct); err != nil {
			return SelectiveDisclosureProofBundle{}, err
		}
		kindByte, err := reader.ReadByte()
		if err != nil {
			return SelectiveDisclosureProofBundle{}, err
		}
		commitment, err := readBytes(reader)
		if err != nil {
			return SelectiveDisclosureProofBundle{}, err
		}
		proof, err := readBytes(reader)
		if err != nil {
			return SelectiveDisclosureProofBundle{}, err
		}
		proofs = append(proofs, ClaimProofEntry{
			ClaimType:  ClaimType(ct),
			ProofKind:  ClaimProofKind(kindByte),
			Commitment: commitment,
			Proof:      proof,
		})
	}

	return SelectiveDisclosureProofBundle{
		Version: version,
		Proofs:  proofs,
	}, nil
}
