package types

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sort"
	"time"
)

// ============================================================================
// Assurance Vector (per-factor evidence state)
// ============================================================================
//
// An AssuranceVector is the per-factor form of an account's authenticated
// evidence state. The composite scorer already computes six factor
// contributions, but only its scalar reduction is persisted
// (keeper/composite_scoring.ComputeAndStoreCompositeScore -> SetScoreWithDetails);
// the per-factor values were flattened into a text summary and discarded.
//
// This type makes those factors first-class on-chain state so a relying party
// can evaluate its own policy ("document > 9500 AND biometric > 9700") instead
// of trusting one opaque number.
//
// PRIVACY: a vector carries scores and provenance references only. It must never
// carry a personal attribute. Date of birth, document number, nationality and
// raw biometric data are forbidden; TestAssuranceVectorHasNoPersonalAttributes
// fails the build if a field name or JSON tag looks like one.
//
// DETERMINISM: every field is integer, string or fixed-width array. There is no
// float and no map iteration in any hashing path; factors are held in a fixed
// order and hashes are computed by an explicit length-prefixed encoder.

// AssuranceVectorVersion is the version of the assurance-vector schema and
// derivation rules. Bump when the field set or the derivation changes in a way
// that alters the vector hash.
const AssuranceVectorVersion = "1.0.0-assurance-vector"

// assuranceVectorDomainSeparator domain-separates the assurance-vector hash from
// any other hash in the module, so a vector commitment can never be replayed as
// some other object's hash.
var assuranceVectorDomainSeparator = []byte("veid/assurance-vector/v1")

// AssuranceFactor identifies one measurable evidence factor.
//
// The values are the spec's composite components, re-expressed in the vocabulary
// a relying party uses. Recency and account age are carried as vector-level
// fields (they describe WHEN the vector was produced, not a single evidence
// source), so they are not factor entries.
type AssuranceFactor string

const (
	// AssuranceFactorDocument is document authenticity: tamper detection,
	// format validity, template match.
	AssuranceFactorDocument AssuranceFactor = "document"

	// AssuranceFactorBiometric is face match: ID photo vs selfie confidence.
	AssuranceFactorBiometric AssuranceFactor = "biometric"

	// AssuranceFactorLiveness is liveness/anti-spoof: blink, motion, challenge.
	AssuranceFactorLiveness AssuranceFactor = "liveness"

	// AssuranceFactorDataConsistency is cross-field consistency of the submitted
	// evidence: field agreement across the documents, age eligibility, and
	// document expiry.
	//
	// This factor is in the set because the composite scorer gives it real weight
	// (15%). Exposing the other five weighted components and hiding this one
	// would leave a relying party unable to re-derive the overall assurance from
	// the vector, and would let a badly-inconsistent document score pass unnoticed
	// as long as the total was high enough.
	AssuranceFactorDataConsistency AssuranceFactor = "data_consistency"

	// AssuranceFactorHistory is historical signals: prior verifications, account age.
	AssuranceFactorHistory AssuranceFactor = "history"

	// AssuranceFactorRisk is risk indicators: fraud patterns, IP reputation,
	// velocity. This is a RISK reading, not an assurance reading — a relying
	// party that wants "low risk" must compare against a low threshold. See
	// AssuranceFactorNotes and IsHigherBetter.
	AssuranceFactorRisk AssuranceFactor = "risk"

	// AssuranceFactorDevice is device trust: fingerprint, integrity, attestation.
	//
	// Device evidence lives inside the composite scorer's risk component rather
	// than in a component of its own, so this factor is derived from that
	// component's device sub-signals instead of being re-measured. When those
	// sub-signals are not supplied the device factor is UNMEASURED, never zero.
	AssuranceFactorDevice AssuranceFactor = "device"
)

// AssuranceFactors lists every factor in its canonical, fixed order.
//
// Canonical order is lexicographic by factor name so hashing never depends on
// construction order, and the order is stable across nodes and versions.
// AllAssuranceFactors is the source of truth: never derive it from a map.
var AssuranceFactors = []AssuranceFactor{
	AssuranceFactorBiometric,
	AssuranceFactorDataConsistency,
	AssuranceFactorDevice,
	AssuranceFactorDocument,
	AssuranceFactorHistory,
	AssuranceFactorLiveness,
	AssuranceFactorRisk,
}

// IsValidAssuranceFactor reports whether f is a known factor.
func IsValidAssuranceFactor(f AssuranceFactor) bool {
	for _, known := range AssuranceFactors {
		if known == f {
			return true
		}
	}
	return false
}

// assuranceFactorHigherBetter and assuranceFactorLowerBetter document the
// direction of each factor reading.
const (
	assuranceFactorHigherBetter = "higher_is_better"
	assuranceFactorLowerBetter  = "lower_is_better"
)

// AssuranceFactorNotes documents the direction of each factor so a relying party
// does not have to infer it, and so the semantics survive in one place.
//
// A policy author reads: document/biometric/liveness/history/device/
// data_consistency are "higher is better". risk is "lower is better" and must be
// bounded above.
var AssuranceFactorNotes = map[AssuranceFactor]string{
	AssuranceFactorDocument:        assuranceFactorHigherBetter,
	AssuranceFactorBiometric:       assuranceFactorHigherBetter,
	AssuranceFactorLiveness:        assuranceFactorHigherBetter,
	AssuranceFactorHistory:         assuranceFactorHigherBetter,
	AssuranceFactorDevice:          assuranceFactorHigherBetter,
	AssuranceFactorDataConsistency: assuranceFactorHigherBetter,
	AssuranceFactorRisk:            assuranceFactorLowerBetter,
}

// IsHigherBetter reports whether a larger value for f means more assurance.
// Callers MUST NOT assume it: the risk factor is deliberately inverted.
func IsHigherBetter(f AssuranceFactor) (bool, bool) {
	note, ok := AssuranceFactorNotes[f]
	if !ok {
		return false, false
	}
	return note == "higher_is_better", true
}

// AssuranceFactorEntry is one factor's reading.
//
// Measured is false when the underlying evidence was absent. That is a distinct
// state from a score of zero: an absent factor means "not measured", so a policy
// requiring device trust can require that the factor was actually measured
// rather than silently accepting an unmeasured factor as 0.
type AssuranceFactorEntry struct {
	// Factor is the factor this entry describes.
	Factor AssuranceFactor `json:"factor"`

	// ScoreBps is the factor's reading in basis points (0-10000).
	ScoreBps uint32 `json:"score_bps"`

	// WeightBps is the weight the composite scorer applied to the evidence this
	// factor was derived from, in basis points (0-10000).
	WeightBps uint32 `json:"weight_bps"`

	// WeightedScoreBps is the weighted contribution carried verbatim from the
	// composite contribution, so a verifier can re-derive the overall assurance
	// without re-running the model.
	WeightedScoreBps uint32 `json:"weighted_score_bps"`

	// Measured reports whether the underlying evidence was present.
	Measured bool `json:"measured"`

	// PassedThreshold reports whether the factor cleared the scorer's own
	// minimum for its evidence.
	PassedThreshold bool `json:"passed_threshold"`
}

// AssuranceVector is an account's per-factor evidence state for one
// verification epoch.
type AssuranceVector struct {
	// Version is the assurance-vector schema/derivation version.
	Version string `json:"version"`

	// Account is the account address this vector belongs to.
	Account string `json:"account"`

	// Epoch increments on every re-verification that supersedes a prior vector,
	// so a verifier can detect that it is holding a superseded vector.
	Epoch uint64 `json:"epoch"`

	// Factors holds the per-factor readings in canonical order. The slice is
	// fixed-length and ordered; it is never a map.
	Factors []AssuranceFactorEntry `json:"factors"`

	// OverallBps is the composite assurance in basis points (0-10000). It is the
	// same measurement as the scalar score, re-expressed in basis points.
	OverallBps uint32 `json:"overall_bps"`

	// ScalarScore is the score the composite scorer reported on its 0-100 scale.
	// Carried so the vector and the existing scalar query cannot drift apart.
	ScalarScore uint32 `json:"scalar_score"`

	// Passed reports the composite scorer's pass decision.
	Passed bool `json:"passed"`

	// ScoreVersion is the scoring algorithm version that produced the vector,
	// e.g. "2.0.0-composite".
	ScoreVersion string `json:"score_version"`

	// ModelVersion is the ML/policy version whose outputs fed the scorer.
	ModelVersion string `json:"model_version"`

	// AccountAgeBps is the account-age reading in basis points (0-10000).
	AccountAgeBps uint32 `json:"account_age_bps"`

	// VerifiedHeight is the consensus block height at which the verification
	// that produced this vector was recorded.
	VerifiedHeight int64 `json:"verified_height"`

	// VerifiedAtUnix is the consensus block time of that verification, in
	// seconds since the Unix epoch. It comes from block time, never wall clock.
	VerifiedAtUnix int64 `json:"verified_at_unix"`

	// InputHash is the composite scorer's hash of its scoring inputs. A verifier
	// uses it to confirm the vector was derived from the inputs it expects.
	InputHash []byte `json:"input_hash"`

	// Commitment is the domain-separated hash over every field above. A relying
	// party that holds the vector can recompute it and check the stored state
	// against a signed result without re-running the scoring model.
	Commitment []byte `json:"commitment"`
}

// ============================================================================
// Derivation
// ============================================================================

// DeviceEvidence carries the device sub-signals that the composite scorer folds
// into its risk component. They are supplied separately because they live on the
// scorer's INPUTS, not on its result; passing them in is what lets the device
// factor be derived without re-running or forking the scoring algorithm.
//
// Every field is an assurance reading in basis points (higher = more trustworthy
// device). Present=false means "no device evidence", which yields an UNMEASURED
// device factor rather than a zero one.
type DeviceEvidence struct {
	// FingerprintBps is the device fingerprint trust score.
	FingerprintBps uint32 `json:"fingerprint_bps"`

	// IntegrityBps is the device integrity attestation score.
	IntegrityBps uint32 `json:"integrity_bps"`

	// AttestationBps is the hardware attestation score.
	AttestationBps uint32 `json:"attestation_bps"`

	// Present reports whether device evidence was collected at all.
	Present bool `json:"present"`
}

// DerivationInput is everything DeriveAssuranceVector needs that does not
// already live on a CompositeScoreResult.
type DerivationInput struct {
	// Account is the account address the vector belongs to.
	Account string

	// Epoch is the verification epoch. Callers pass the prior epoch + 1; the
	// first vector for an account is epoch 0.
	Epoch uint64

	// ModelVersion is the ML/policy version behind the scorer's inputs.
	ModelVersion string

	// AccountAgeBps is the account-age reading in basis points.
	AccountAgeBps uint32

	// Device carries the device sub-signals for the device factor.
	Device DeviceEvidence

	// VerifiedAt is the CONSENSUS block time of the verification.
	VerifiedAt time.Time
}

// componentToFactor maps an assurance factor to the composite component it is
// derived from. The map is keyed by FACTOR because that is the direction the
// derivation walks; the composite component names it names are the module's
// existing constants.
//
// face_match and liveness_detection become their own factors. Device trust has no
// component of its own (it lives inside risk_indicators), so it is derived
// separately by deriveFactorEntry.
//
// A factor reports the weight of the COMPONENT it was derived from. Because the
// device and risk factors share the risk component, they report the same weight,
// so summing factor weights does NOT reconstruct the total weight. This coupling
// is recorded in the PR so the factor-correlation card orders its independence
// caps against it.
var componentToFactor = map[AssuranceFactor]string{
	AssuranceFactorDocument:        ComponentDocumentAuthenticity,
	AssuranceFactorBiometric:       ComponentFaceMatch,
	AssuranceFactorLiveness:        ComponentLivenessDetection,
	AssuranceFactorHistory:         ComponentHistoricalSignals,
	AssuranceFactorDataConsistency: ComponentDataConsistency,
	AssuranceFactorRisk:            ComponentRiskIndicators,
}

// DeriveAssuranceVector builds a vector deterministically from a composite score
// result.
//
// It does not re-score anything: every number is read off the result the scorer
// already produced, so the vector cannot disagree with the scalar. Passing a nil
// result, an empty account, or an out-of-range basis-point value is an error
// rather than a silently zeroed vector.
func DeriveAssuranceVector(result *CompositeScoreResult, in DerivationInput) (*AssuranceVector, error) {
	if result == nil {
		return nil, ErrInvalidScoringModel.Wrap("nil composite score result")
	}
	if in.Account == "" {
		return nil, ErrInvalidScoringModel.Wrap("empty account for assurance vector")
	}
	if in.AccountAgeBps > uint32(MaxBasisPoints) {
		return nil, ErrInvalidScoringModel.Wrapf("account age %d exceeds basis points", in.AccountAgeBps)
	}

	// Index contributions by component name so the derivation does not depend on
	// the order ComputeCompositeScore happened to append them in. Duplicate
	// components are rejected rather than silently collapsed: a duplicate would
	// mean the scorer forked, and picking either copy would be a consensus fork
	// of its own.
	byComponent := make(map[string]CompositeScoreContribution, len(result.Contributions))
	for _, contrib := range result.Contributions {
		if _, duplicate := byComponent[contrib.ComponentName]; duplicate {
			return nil, ErrInvalidScoringModel.Wrapf("duplicate composite contribution %q", contrib.ComponentName)
		}
		if contrib.RawScore > uint32(MaxBasisPoints) {
			return nil, ErrInvalidScoringModel.Wrapf("component %q raw score %d exceeds basis points", contrib.ComponentName, contrib.RawScore)
		}
		if contrib.Weight > uint32(MaxBasisPoints) {
			return nil, ErrInvalidScoringModel.Wrapf("component %q weight %d exceeds basis points", contrib.ComponentName, contrib.Weight)
		}
		byComponent[contrib.ComponentName] = contrib
	}

	entries := make([]AssuranceFactorEntry, 0, len(AssuranceFactors))
	for _, factor := range AssuranceFactors {
		entries = append(entries, deriveFactorEntry(factor, byComponent, in.Device))
	}

	v := &AssuranceVector{
		Version:        AssuranceVectorVersion,
		Account:        in.Account,
		Epoch:          in.Epoch,
		Factors:        entries,
		OverallBps:     OverallBpsFromScore(result.FinalScore),
		ScalarScore:    result.FinalScore,
		Passed:         result.Passed,
		ScoreVersion:   result.ScoreVersion,
		ModelVersion:   in.ModelVersion,
		AccountAgeBps:  in.AccountAgeBps,
		VerifiedHeight: result.BlockHeight,
		VerifiedAtUnix: in.VerifiedAt.UTC().Unix(),
		InputHash:      append([]byte(nil), result.InputHash...),
	}

	v.Commitment = v.ComputeCommitment()
	return v, nil
}

// deriveFactorEntry builds one factor's entry from its component contribution.
//
// An absent component yields an UNMEASURED entry with zero score and zero weight:
// no assurance is invented for evidence that was never collected.
func deriveFactorEntry(
	factor AssuranceFactor,
	byComponent map[string]CompositeScoreContribution,
	device DeviceEvidence,
) AssuranceFactorEntry {
	entry := AssuranceFactorEntry{Factor: factor}

	if factor == AssuranceFactorDevice {
		score, measured := deriveDeviceScore(device)
		entry.ScoreBps = score
		entry.Measured = measured
		if !measured {
			return entry
		}
		// The device reading rides on the risk component, so it carries the
		// risk component's weight and contribution.
		if risk, ok := byComponent[ComponentRiskIndicators]; ok {
			entry.WeightBps = risk.Weight
			entry.WeightedScoreBps = risk.WeightedScore
			entry.PassedThreshold = risk.PassedThreshold
		}
		return entry
	}

	component, mapped := componentToFactor[factor]
	if !mapped {
		// Unreachable for the canonical factor set. Returning a zeroed entry would
		// emit a factor with no evidence behind it; Validate then rejects the
		// vector because an unmeasured factor must not be paired with a weight.
		return entry
	}

	contrib, present := byComponent[component]
	if !present {
		return entry
	}

	entry.ScoreBps = contrib.RawScore
	entry.WeightBps = contrib.Weight
	entry.WeightedScoreBps = contrib.WeightedScore
	entry.Measured = contributionWasMeasured(contrib)
	entry.PassedThreshold = contrib.PassedThreshold
	return entry
}

// contributionWasMeasured reports whether a contribution's evidence was present.
//
// ComputeCompositeScore records presence in CompositeScoreResult.ComponentPresence,
// which is not reachable from a contribution alone, so presence is inferred from
// the contribution itself: a missing-evidence component contributes a zero raw
// score and a MISSING_* reason code, while a measured component carries a
// positive raw score.
func contributionWasMeasured(contrib CompositeScoreContribution) bool {
	if contrib.ReasonCode == CompositeReasonMissingDocument || contrib.ReasonCode == CompositeReasonMissingSelfie {
		return false
	}
	return contrib.RawScore > 0
}

// deriveDeviceScore averages the available device sub-signals in basis points.
//
// Unlike the risk component, these sub-signals are already assurance readings
// (higher = more trustworthy device), so no inversion is applied. A truncating
// average over a fixed set of integers is deterministic. Returns measured=false
// when no device evidence was supplied or any sub-score is out of range, so the
// caller records the factor as unmeasured rather than publishing a wrapped value.
func deriveDeviceScore(device DeviceEvidence) (uint32, bool) {
	if !device.Present {
		return 0, false
	}

	var sum uint64
	var count uint32

	for _, sub := range []uint32{device.FingerprintBps, device.IntegrityBps, device.AttestationBps} {
		if sub > uint32(MaxBasisPoints) {
			return 0, false
		}
		sum += uint64(sub)
		count++
	}

	if count == 0 {
		return 0, false
	}

	// sum <= count * 10000 and count >= 1, so the quotient is at most 10000.
	return uint32(sum / uint64(count)), true /* #nosec G115 -- sum is the sum of at most 3 sub-scores each bounded above by MaxBasisPoints (10000), and count is exactly 3 whenever the loop runs, so the quotient is at most 10000 and cannot overflow uint32 */ //nolint:gosec
}

// OverallBpsFromScore re-expresses a 0-100 score in basis points using integer
// math only, so no float ever enters consensus state.
func OverallBpsFromScore(score uint32) uint32 {
	if score > MaxScore {
		score = MaxScore
	}
	// score/100 * 10000 == score * 100, exact in integer math.
	return score * 100
}

// ============================================================================
// Commitment / hashing
// ============================================================================

// assuranceHasher is the subset of hash.Hash the encoders need. sha256.Hash
// satisfies it, and declaring it keeps the encoders testable.
type assuranceHasher interface {
	Write(p []byte) (int, error)
}

// ComputeCommitment returns the domain-separated hash over the vector's fields.
//
// The encoder is explicit and length-prefixed: every variable-length value is
// written with its length first, so no two distinct vectors can serialise to the
// same byte stream (no concatenation ambiguity). Factors are written in stored
// order, which Validate guarantees is canonical.
func (v *AssuranceVector) ComputeCommitment() []byte {
	h := sha256.New()
	h.Write(assuranceVectorDomainSeparator) //nolint:errcheck // hash writes never fail

	writeAssuranceBytes(h, []byte(v.Version))
	writeAssuranceBytes(h, []byte(v.Account))
	writeAssuranceUint64(h, v.Epoch)

	writeAssuranceUint64(h, uint64(len(v.Factors)))
	for _, entry := range v.Factors {
		writeAssuranceBytes(h, []byte(entry.Factor))
		writeAssuranceUint32(h, entry.ScoreBps)
		writeAssuranceUint32(h, entry.WeightBps)
		writeAssuranceUint32(h, entry.WeightedScoreBps)
		writeAssuranceBytes(h, []byte{assuranceBoolByte(entry.Measured)})
		writeAssuranceBytes(h, []byte{assuranceBoolByte(entry.PassedThreshold)})
	}

	writeAssuranceUint32(h, v.OverallBps)
	writeAssuranceUint32(h, v.ScalarScore)
	writeAssuranceBytes(h, []byte{assuranceBoolByte(v.Passed)})
	writeAssuranceBytes(h, []byte(v.ScoreVersion))
	writeAssuranceBytes(h, []byte(v.ModelVersion))
	writeAssuranceUint32(h, v.AccountAgeBps)
	writeAssuranceInt64(h, v.VerifiedHeight)
	writeAssuranceInt64(h, v.VerifiedAtUnix)
	writeAssuranceBytes(h, v.InputHash)

	return h.Sum(nil)
}

// VerifyCommitment recomputes the commitment and compares it without an early
// exit. This is the check a relying party performs on a stored vector.
func (v *AssuranceVector) VerifyCommitment() bool {
	return constantTimeEqual(v.Commitment, v.ComputeCommitment())
}

func assuranceBoolByte(b bool) byte {
	if b {
		return 0x01
	}
	return 0x00
}

func writeAssuranceBytes(h assuranceHasher, bz []byte) {
	var lenBuf [8]byte
	binary.BigEndian.PutUint64(lenBuf[:], uint64(len(bz)))
	_, _ = h.Write(lenBuf[:]) // hash.Hash never returns an error per the io.Writer contract
	_, _ = h.Write(bz)        // hash.Hash never returns an error per the io.Writer contract
}

func writeAssuranceUint32(h assuranceHasher, value uint32) {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], value)
	_, _ = h.Write(buf[:]) // hash.Hash never returns an error per the io.Writer contract
}

func writeAssuranceUint64(h assuranceHasher, value uint64) {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], value)
	_, _ = h.Write(buf[:]) // hash.Hash never returns an error per the io.Writer contract
}

func writeAssuranceInt64(h assuranceHasher, value int64) {
	var buf [8]byte
	// The two's-complement reinterpretation is the intent: a negative int64 (a
	// block height cannot be negative in practice, but the encoder is total)
	// must serialise distinctly rather than wrap into a colliding value.
	binary.BigEndian.PutUint64(buf[:], uint64(value)) /* #nosec G115 -- fixed-width two's-complement encoding by design: the int64 is reinterpreted bit-for-bit into 8 bytes, which is lossless by definition and is not an arithmetic truncation */ //nolint:gosec
	_, _ = h.Write(buf[:])                            // hash.Hash never returns an error per the io.Writer contract
}

// constantTimeEqual compares two byte slices without an early exit on the first
// differing byte.
func constantTimeEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte
	for i := range a {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}

// ============================================================================
// Accessors
// ============================================================================

// Lookup returns the entry for a factor.
func (v *AssuranceVector) Lookup(factor AssuranceFactor) (AssuranceFactorEntry, bool) {
	for _, entry := range v.Factors {
		if entry.Factor == factor {
			return entry, true
		}
	}
	return AssuranceFactorEntry{}, false
}

// ScoreBps returns a factor's reading in basis points and whether it was both
// present and measured. A relying party must treat !ok and !Measured as "no
// assurance claim", not as a score of zero.
func (v *AssuranceVector) ScoreBps(factor AssuranceFactor) (uint32, bool) {
	entry, ok := v.Lookup(factor)
	if !ok || !entry.Measured {
		return 0, false
	}
	return entry.ScoreBps, true
}

// MeasuredFactors returns the factors that carry a real measurement, in
// canonical order. A policy needing N factors can require
// len(MeasuredFactors()) >= N; the factor-correlation card further restricts
// which combinations count as independent.
func (v *AssuranceVector) MeasuredFactors() []AssuranceFactor {
	out := make([]AssuranceFactor, 0, len(v.Factors))
	for _, entry := range v.Factors {
		if entry.Measured {
			out = append(out, entry.Factor)
		}
	}
	return out
}

// TotalComponentWeight returns the sum of the component weights the vector was
// built from, counting each COMPONENT once.
//
// Factor weights must not be summed naively: the device and risk factors share
// the risk component and report the same weight, so a naive sum double-counts it
// and exceeds 10000. A fully measured vector totals exactly 10000, which is what
// a verifier should assert.
func (v *AssuranceVector) TotalComponentWeight() uint32 {
	var total uint32
	for _, entry := range v.Factors {
		total += entry.WeightBps
	}
	if device, ok := v.Lookup(AssuranceFactorDevice); ok && device.Measured {
		// The device factor duplicates the risk component's weight.
		total -= device.WeightBps
	}
	return total
}

// ============================================================================
// Validation
// ============================================================================

// Validate checks structural invariants: known factors, canonical order, exact
// factor count, basis-point ranges, absence of assurance on unmeasured factors,
// and a matching commitment.
func (v *AssuranceVector) Validate() error {
	if v == nil {
		return ErrInvalidScoringModel.Wrap("nil assurance vector")
	}
	if v.Version != AssuranceVectorVersion {
		return ErrInvalidScoringModel.Wrapf("assurance vector version %q, want %q", v.Version, AssuranceVectorVersion)
	}
	if v.Account == "" {
		return ErrInvalidScoringModel.Wrap("assurance vector has no account")
	}
	if len(v.Factors) != len(AssuranceFactors) {
		return ErrInvalidScoringModel.Wrapf("assurance vector has %d factors, want %d", len(v.Factors), len(AssuranceFactors))
	}

	// Canonical order, no duplicates, no unknown factors.
	for i, entry := range v.Factors {
		if !IsValidAssuranceFactor(entry.Factor) {
			return ErrInvalidScoringModel.Wrapf("unknown assurance factor %q", entry.Factor)
		}
		if i > 0 && v.Factors[i-1].Factor >= entry.Factor {
			return ErrInvalidScoringModel.Wrapf("assurance factors out of canonical order at %d: %q then %q",
				i, v.Factors[i-1].Factor, entry.Factor)
		}
		if entry.ScoreBps > uint32(MaxBasisPoints) {
			return ErrInvalidScoringModel.Wrapf("factor %q score %d exceeds basis points", entry.Factor, entry.ScoreBps)
		}
		if entry.WeightBps > uint32(MaxBasisPoints) {
			return ErrInvalidScoringModel.Wrapf("factor %q weight %d exceeds basis points", entry.Factor, entry.WeightBps)
		}
		if entry.WeightedScoreBps > uint32(MaxBasisPoints) {
			return ErrInvalidScoringModel.Wrapf("factor %q weighted score %d exceeds basis points", entry.Factor, entry.WeightedScoreBps)
		}
		// An unmeasured factor must not carry assurance. Without this, a caller
		// could publish an unmeasured factor with a full score and a relying
		// party would read it as a real measurement.
		if !entry.Measured && (entry.ScoreBps != 0 || entry.WeightedScoreBps != 0) {
			return ErrInvalidScoringModel.Wrapf("unmeasured factor %q carries a non-zero score", entry.Factor)
		}
	}

	if v.OverallBps > uint32(MaxBasisPoints) {
		return ErrInvalidScoringModel.Wrapf("overall %d exceeds basis points", v.OverallBps)
	}
	if v.ScalarScore > MaxScore {
		return ErrInvalidScoringModel.Wrapf("scalar score %d exceeds max %d", v.ScalarScore, MaxScore)
	}
	if v.AccountAgeBps > uint32(MaxBasisPoints) {
		return ErrInvalidScoringModel.Wrapf("account age %d exceeds basis points", v.AccountAgeBps)
	}
	if len(v.Commitment) == 0 {
		return ErrInvalidScoringModel.Wrap("assurance vector has no commitment")
	}
	if !v.VerifyCommitment() {
		return ErrInvalidScoringModel.Wrap("assurance vector commitment does not match its contents")
	}
	return nil
}

// ============================================================================
// Recency
// ============================================================================

// DefaultAssuranceVectorMaxAgeSeconds is the fallback freshness window when no
// explicit one is supplied. Callers holding module params should pass
// Params.VerificationExpiryDays converted to seconds instead.
const DefaultAssuranceVectorMaxAgeSeconds int64 = 365 * 24 * 60 * 60

// AssuranceVectorRecency describes how current a vector's evidence is.
type AssuranceVectorRecency struct {
	// Epoch is the vector's verification epoch.
	Epoch uint64 `json:"epoch"`

	// AgeSeconds is the age of the verification in seconds of CONSENSUS time.
	AgeSeconds int64 `json:"age_seconds"`

	// VerifiedHeight is the block height of the verification.
	VerifiedHeight int64 `json:"verified_height"`

	// VerifiedAtUnix is the consensus time of the verification.
	VerifiedAtUnix int64 `json:"verified_at_unix"`

	// Stale reports whether the evidence is older than the freshness window.
	Stale bool `json:"stale"`

	// MaxAgeSeconds is the freshness window that was applied.
	MaxAgeSeconds int64 `json:"max_age_seconds"`
}

// Recency evaluates the vector's freshness against CONSENSUS time.
//
// nowUnix must come from the block header, never wall clock: a wall-clock read
// would make staleness differ between validators and fork the chain.
//
// Staleness here is purely a freshness statement about the EVIDENCE. It does NOT
// apply the score-decay curve in keeper/score_decay.go, and that is deliberate:
// decay lowers the SCORE, staleness marks the evidence old, and a relying party
// may legitimately want the factor readings of an old vector in order to apply
// its own policy. Callers that want decayed numbers must use the existing decay
// path, so the module keeps exactly one decay implementation.
func (v *AssuranceVector) Recency(nowUnix time.Time, _ int64, maxAgeSeconds int64) AssuranceVectorRecency {
	if maxAgeSeconds <= 0 {
		maxAgeSeconds = DefaultAssuranceVectorMaxAgeSeconds
	}

	age := nowUnix.UTC().Unix() - v.VerifiedAtUnix
	if age < 0 {
		// Defensive: a vector dated after the block being executed cannot be
		// trusted as fresh.
		age = 0
	}

	return AssuranceVectorRecency{
		Epoch:          v.Epoch,
		AgeSeconds:     age,
		VerifiedHeight: v.VerifiedHeight,
		VerifiedAtUnix: v.VerifiedAtUnix,
		Stale:          age > maxAgeSeconds,
		MaxAgeSeconds:  maxAgeSeconds,
	}
}

// IsSuperseded reports whether a vector at currentEpoch replaces this one, i.e.
// whether a re-verification has moved the account forward.
func (v *AssuranceVector) IsSuperseded(currentEpoch uint64) bool {
	return currentEpoch > v.Epoch
}

// NextEpoch returns the epoch a re-verification should write for an account
// whose current vector is at currentEpoch. The first vector is epoch 0.
func NextEpoch(currentEpoch uint64) uint64 {
	return currentEpoch + 1
}

// SortAssuranceFactorEntries orders entries canonically by factor name. Used
// when reading entries from a source that does not guarantee order; the sort is
// over a fixed key set, so it is deterministic.
func SortAssuranceFactorEntries(entries []AssuranceFactorEntry) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Factor < entries[j].Factor
	})
}

// String renders the vector for logs and test failures. Not stable API.
func (v *AssuranceVector) String() string {
	return fmt.Sprintf("AssuranceVector{account:%s epoch:%d overall_bps:%d measured:%v}",
		v.Account, v.Epoch, v.OverallBps, v.MeasuredFactors())
}
