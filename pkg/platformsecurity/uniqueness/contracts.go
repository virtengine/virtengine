// Package uniqueness defines fixture-only contracts for threshold biometric uniqueness.
package uniqueness

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"time"
)

const (
	Version1                   uint32 = 1
	FixtureOnlyState                  = "fixture_only"
	ProductionUnavailable             = "production_unavailable"
	UniquenessPurpose                 = "uniqueness_only"
	ExternalReviewRequired            = "external_review_required"
	ExactProgramDomain                = "virtengine.uniqueness.program/v1"
	PairwiseRelyingPartyDomain        = "virtengine.uniqueness.relying-party/v1"
	MaxCandidateCount          uint32 = 1024
)

var (
	ErrConflict              = errors.New("idempotency conflict")
	ErrProductionUnavailable = errors.New("production uniqueness is unavailable")
	digestPattern            = regexp.MustCompile(`^[0-9a-f]{64}$`)
	opaqueIDPattern          = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,127}$`)
)

type NodeState string

const (
	NodeActive  NodeState = "active"
	NodeRevoked NodeState = "revoked"
)

type CustodyNodeIdentity struct {
	Version                  uint32    `json:"version"`
	NodeID                   string    `json:"node_id"`
	OperatorCommitment       string    `json:"operator_commitment"`
	FailureDomainCommitment  string    `json:"failure_domain_commitment"`
	SigningKeyID             string    `json:"signing_key_id"`
	SigningKeyEpoch          uint64    `json:"signing_key_epoch"`
	SigningPublicKey         []byte    `json:"signing_public_key"`
	EncryptionCommitment     string    `json:"encryption_commitment"`
	ShareCommitment          string    `json:"share_commitment"`
	EndpointCommitment       string    `json:"endpoint_commitment"`
	NodeKeyBindingCommitment string    `json:"node_key_binding_commitment"`
	State                    NodeState `json:"state"`
}

func (n CustodyNodeIdentity) Validate() error {
	if n.Version != Version1 || !validOpaqueID(n.NodeID) || !validOpaqueID(n.SigningKeyID) {
		return errors.New("canonical custody node identifiers are required")
	}
	if !validDigests(n.OperatorCommitment, n.FailureDomainCommitment, n.EncryptionCommitment, n.ShareCommitment, n.EndpointCommitment, n.NodeKeyBindingCommitment) {
		return errors.New("strict custody node commitments are required")
	}
	if n.SigningKeyEpoch == 0 || len(n.SigningPublicKey) != ed25519.PublicKeySize {
		return errors.New("valid signing key epoch and Ed25519 public key are required")
	}
	if n.State != NodeActive && n.State != NodeRevoked {
		return errors.New("unknown custody node state")
	}
	return nil
}
func (n CustodyNodeIdentity) encode(e *canonicalEncoder) {
	e.u32(n.Version)
	e.text(n.NodeID)
	e.text(n.OperatorCommitment)
	e.text(n.FailureDomainCommitment)
	e.text(n.SigningKeyID)
	e.u64(n.SigningKeyEpoch)
	e.bytes(n.SigningPublicKey)
	e.text(n.EncryptionCommitment)
	e.text(n.ShareCommitment)
	e.text(n.EndpointCommitment)
	e.text(n.NodeKeyBindingCommitment)
	e.text(string(n.State))
}

type CustodyNodeSet struct {
	Version   uint32                `json:"version"`
	Threshold uint32                `json:"threshold"`
	Nodes     []CustodyNodeIdentity `json:"nodes"`
}

func (s CustodyNodeSet) Validate() error {
	if s.Version != Version1 || s.Threshold < 2 {
		return errors.New("versioned custody set with threshold of at least two is required")
	}
	active := uint32(0)
	ids, operators, domains := map[string]bool{}, map[string]bool{}, map[string]bool{}
	keyIDs, publicKeys, bindings := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, n := range s.Nodes {
		if err := n.Validate(); err != nil {
			return err
		}
		publicKey := hex.EncodeToString(n.SigningPublicKey)
		if ids[n.NodeID] || keyIDs[n.SigningKeyID] || publicKeys[publicKey] || bindings[n.NodeKeyBindingCommitment] {
			return errors.New("duplicate or aliased custody identity or signing key")
		}
		ids[n.NodeID], keyIDs[n.SigningKeyID], publicKeys[publicKey], bindings[n.NodeKeyBindingCommitment] = true, true, true, true
		if n.State == NodeActive {
			active++
			if operators[n.OperatorCommitment] || domains[n.FailureDomainCommitment] {
				return errors.New("active participants must have independent operators and failure domains")
			}
			operators[n.OperatorCommitment], domains[n.FailureDomainCommitment] = true, true
		}
	}
	if active < s.Threshold || uint32(len(operators)) < s.Threshold || uint32(len(domains)) < s.Threshold {
		return errors.New("insufficient independent active custody nodes")
	}
	return nil
}
func (s CustodyNodeSet) CanonicalBytes() ([]byte, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	nodes := slices.Clone(s.Nodes)
	slices.SortFunc(nodes, func(a, b CustodyNodeIdentity) int { return compareString(a.NodeID, b.NodeID) })
	e := newCanonicalEncoder("virtengine.uniqueness.custody-node-set/v1")
	e.u32(s.Version)
	e.u32(s.Threshold)
	e.u32(uint32(len(nodes)))
	for _, n := range nodes {
		n.encode(e)
	}
	return e.result(), nil
}
func (s CustodyNodeSet) Digest() (string, error) {
	value, err := s.CanonicalBytes()
	if err != nil {
		return "", err
	}
	return digest(value), nil
}

type KeyEpochState string

const (
	KeyEpochPending  KeyEpochState = "pending"
	KeyEpochActive   KeyEpochState = "active"
	KeyEpochRotating KeyEpochState = "rotating"
	KeyEpochRevoked  KeyEpochState = "revoked"
	KeyEpochExpired  KeyEpochState = "expired"
)

type ThresholdKeyEpoch struct {
	Version              uint32        `json:"version"`
	ProfileID            string        `json:"profile_id"`
	Epoch                uint64        `json:"epoch"`
	Threshold            uint32        `json:"threshold"`
	ParticipantSetDigest string        `json:"participant_set_digest"`
	PublicCommitment     string        `json:"public_commitment"`
	TransformDigest      string        `json:"transform_profile_digest"`
	ActivationCoordinate uint64        `json:"activation_coordinate"`
	ExpiryCoordinate     uint64        `json:"expiry_coordinate"`
	State                KeyEpochState `json:"state"`
}

func (e ThresholdKeyEpoch) Validate() error {
	if e.Version != Version1 || e.Epoch == 0 || e.Threshold < 2 || !validOpaqueID(e.ProfileID) || !validDigests(e.ParticipantSetDigest, e.PublicCommitment, e.TransformDigest) {
		return errors.New("complete threshold key epoch is required")
	}
	if e.ExpiryCoordinate <= e.ActivationCoordinate {
		return errors.New("key epoch expiry must follow activation")
	}
	switch e.State {
	case KeyEpochPending, KeyEpochActive, KeyEpochRotating, KeyEpochRevoked, KeyEpochExpired:
		return nil
	}
	return errors.New("unknown key epoch state")
}
func (e ThresholdKeyEpoch) ValidateWithNodeSet(nodes CustodyNodeSet) error {
	if err := e.Validate(); err != nil {
		return err
	}
	setDigest, err := nodes.Digest()
	if err != nil {
		return err
	}
	if e.ParticipantSetDigest != setDigest || e.Threshold != nodes.Threshold {
		return errors.New("key epoch participant set or threshold mismatch")
	}
	for _, n := range nodes.Nodes {
		if n.State == NodeActive && n.SigningKeyEpoch != e.Epoch {
			return errors.New("active node signing-key epoch does not match threshold epoch")
		}
	}
	return nil
}
func ValidateKeyEpochTransition(previous, next ThresholdKeyEpoch) error {
	if err := previous.Validate(); err != nil {
		return err
	}
	if err := next.Validate(); err != nil {
		return err
	}
	if previous.Version != next.Version || previous.ProfileID != next.ProfileID || next.Epoch != previous.Epoch+1 || previous.State != KeyEpochRotating || next.State != KeyEpochPending {
		return errors.New("rotation must advance exactly one pending epoch")
	}
	if next.ActivationCoordinate < previous.ExpiryCoordinate || previous.PublicCommitment == next.PublicCommitment || previous.TransformDigest == next.TransformDigest {
		return errors.New("rotation ranges or commitments are invalid")
	}
	return nil
}
func ValidateKeyEpochStateTransition(previous, next ThresholdKeyEpoch) error {
	if err := previous.Validate(); err != nil {
		return err
	}
	if err := next.Validate(); err != nil {
		return err
	}
	if previous.Version != next.Version || previous.ProfileID != next.ProfileID || previous.Epoch != next.Epoch || previous.Threshold != next.Threshold || previous.ParticipantSetDigest != next.ParticipantSetDigest || previous.PublicCommitment != next.PublicCommitment || previous.TransformDigest != next.TransformDigest || previous.ActivationCoordinate != next.ActivationCoordinate || previous.ExpiryCoordinate != next.ExpiryCoordinate {
		return errors.New("key epoch binding changed during state transition")
	}
	allowed := previous.State == KeyEpochPending && next.State == KeyEpochActive || previous.State == KeyEpochActive && (next.State == KeyEpochRotating || next.State == KeyEpochRevoked || next.State == KeyEpochExpired) || previous.State == KeyEpochRotating && (next.State == KeyEpochRevoked || next.State == KeyEpochExpired)
	if !allowed {
		return errors.New("non-monotonic key epoch state transition")
	}
	return nil
}
func (e ThresholdKeyEpoch) CanIssue(coordinate uint64) error {
	if e.State != KeyEpochActive || coordinate < e.ActivationCoordinate || coordinate >= e.ExpiryCoordinate {
		return errors.New("issuance requires an active in-range key epoch")
	}
	return nil
}

type OutputCommitmentRules struct {
	Version       uint32 `json:"version"`
	Algorithm     string `json:"algorithm"`
	OutputBytes   uint32 `json:"output_bytes"`
	DomainBinding bool   `json:"domain_binding"`
}
type CancellableTemplateProfile struct {
	Version             uint32                `json:"version"`
	ProfileID           string                `json:"profile_id"`
	Purpose             string                `json:"purpose"`
	AlgorithmDigest     string                `json:"algorithm_digest"`
	ProfileDigest       string                `json:"profile_digest"`
	VersionDigest       string                `json:"version_digest"`
	DistanceThreshold   int64                 `json:"distance_threshold_fixed"`
	TransformKeyEpoch   uint64                `json:"transform_key_epoch"`
	OutputRules         OutputCommitmentRules `json:"output_commitment_rules"`
	State               string                `json:"state"`
	ExternalReviewBlock string                `json:"external_review_blocker"`
}

func (p CancellableTemplateProfile) Validate() error {
	if p.Version != Version1 || p.OutputRules.Version != Version1 || !validOpaqueID(p.ProfileID) || !validDigests(p.AlgorithmDigest, p.ProfileDigest, p.VersionDigest) || !validOpaqueID(p.OutputRules.Algorithm) {
		return errors.New("complete cancellable template profile is required")
	}
	if p.Purpose != UniquenessPurpose || p.DistanceThreshold < 0 || p.TransformKeyEpoch == 0 || p.OutputRules.OutputBytes == 0 || !p.OutputRules.DomainBinding {
		return errors.New("uniqueness-only fixed-point profile with domain-bound output rules is required")
	}
	if p.State != FixtureOnlyState || p.ExternalReviewBlock != ExternalReviewRequired {
		return ErrProductionUnavailable
	}
	return nil
}

type EnrollmentOutcome string

const (
	OutcomePending             EnrollmentOutcome = "pending"
	OutcomePossibleMatchReview EnrollmentOutcome = "possible_match_review"
	OutcomeDuplicateConfirmed  EnrollmentOutcome = "duplicate_confirmed"
	OutcomeFinalUnique         EnrollmentOutcome = "final_unique"
	OutcomeRejected            EnrollmentOutcome = "rejected"
	OutcomeUnavailable         EnrollmentOutcome = "unavailable"
	OutcomeCancelled           EnrollmentOutcome = "cancelled"
)

type ReasonCode string

const (
	ReasonPending            ReasonCode = "pending"
	ReasonPossibleMatch      ReasonCode = "possible_match"
	ReasonDuplicate          ReasonCode = "duplicate_confirmed"
	ReasonFinalUnique        ReasonCode = "final_unique"
	ReasonPolicyRejected     ReasonCode = "policy_rejected"
	ReasonServiceUnavailable ReasonCode = "service_unavailable"
	ReasonCancelled          ReasonCode = "cancelled"
)

func (r ReasonCode) valid() bool {
	switch r {
	case ReasonPending, ReasonPossibleMatch, ReasonDuplicate, ReasonFinalUnique, ReasonPolicyRejected, ReasonServiceUnavailable, ReasonCancelled:
		return true
	}
	return false
}

type Freshness struct {
	IssuedAt  int64  `json:"issued_at"`
	ExpiresAt int64  `json:"expires_at"`
	Nonce     string `json:"nonce"`
}

func (f Freshness) ValidateStructure() error {
	if !validOpaqueID(f.Nonce) || f.IssuedAt < 0 || f.ExpiresAt <= f.IssuedAt {
		return errors.New("invalid structural freshness binding")
	}
	return nil
}
func (f Freshness) ValidateAt(now time.Time, lifetime, skew time.Duration) error {
	if err := f.ValidateStructure(); err != nil {
		return err
	}
	if lifetime <= 0 || skew < 0 || time.Duration(f.ExpiresAt-f.IssuedAt)*time.Second > lifetime {
		return errors.New("freshness lifetime exceeds policy")
	}
	if f.IssuedAt > now.Add(skew).Unix() || f.ExpiresAt <= now.Add(-skew).Unix() {
		return errors.New("stale or future freshness binding")
	}
	return nil
}

type EnrollmentRequest struct {
	Version         uint32    `json:"version"`
	RequestDigest   string    `json:"request_digest"`
	EvidenceDigest  string    `json:"evidence_digest"`
	ModelDigest     string    `json:"model_digest"`
	RuntimeDigest   string    `json:"runtime_digest"`
	ProfileDigest   string    `json:"profile_digest"`
	ProgramIDDigest string    `json:"program_id_digest"`
	PolicyDigest    string    `json:"policy_digest"`
	IdempotencyKey  string    `json:"idempotency_key"`
	KeyEpoch        uint64    `json:"key_epoch"`
	Freshness       Freshness `json:"freshness"`
}

func (r EnrollmentRequest) ValidateStructure() error {
	if r.Version != Version1 || r.KeyEpoch == 0 || !validDigests(r.RequestDigest, r.EvidenceDigest, r.ModelDigest, r.RuntimeDigest, r.ProfileDigest, r.ProgramIDDigest, r.PolicyDigest) || !validOpaqueID(r.IdempotencyKey) {
		return errors.New("complete opaque enrollment request is required")
	}
	return r.Freshness.ValidateStructure()
}
func (r EnrollmentRequest) Validate(now time.Time) error {
	if err := r.ValidateStructure(); err != nil {
		return err
	}
	return r.Freshness.ValidateAt(now, 5*time.Minute, 0)
}

type EnrollmentRecord struct {
	Version            uint32            `json:"version"`
	RequestDigest      string            `json:"request_digest"`
	EvidenceDigest     string            `json:"evidence_digest"`
	ModelDigest        string            `json:"model_digest"`
	RuntimeDigest      string            `json:"runtime_digest"`
	ProfileDigest      string            `json:"profile_digest"`
	ProgramIDDigest    string            `json:"program_id_digest"`
	PolicyDigest       string            `json:"policy_digest"`
	IdempotencyKey     string            `json:"idempotency_key"`
	KeyEpoch           uint64            `json:"key_epoch"`
	Freshness          Freshness         `json:"freshness"`
	Outcome            EnrollmentOutcome `json:"outcome"`
	Reason             ReasonCode        `json:"reason"`
	TemplateCommitment string            `json:"template_commitment"`
	ScopedNullifier    string            `json:"scoped_nullifier,omitempty"`
}

func (r EnrollmentRecord) Validate() error {
	if r.Version != Version1 || r.KeyEpoch == 0 || !validDigests(r.RequestDigest, r.EvidenceDigest, r.ModelDigest, r.RuntimeDigest, r.ProfileDigest, r.ProgramIDDigest, r.PolicyDigest, r.TemplateCommitment) || !validOpaqueID(r.IdempotencyKey) || !r.Reason.valid() {
		return errors.New("complete enrollment record commitments are required")
	}
	if err := r.Freshness.ValidateStructure(); err != nil {
		return err
	}
	switch r.Outcome {
	case OutcomePending, OutcomePossibleMatchReview, OutcomeDuplicateConfirmed, OutcomeRejected, OutcomeUnavailable, OutcomeCancelled:
		if r.ScopedNullifier != "" {
			return errors.New("non-final outcome must not contain a nullifier")
		}
	case OutcomeFinalUnique:
		if !validDigest(r.ScopedNullifier) {
			return errors.New("final unique outcome requires a strict nullifier digest")
		}
	default:
		return errors.New("unknown enrollment outcome")
	}
	return nil
}

type NullifierDomain struct {
	Version              uint32 `json:"version"`
	Kind                 string `json:"kind"`
	ProgramIDDigest      string `json:"program_id_digest"`
	RelyingPartyIDDigest string `json:"relying_party_id_digest,omitempty"`
}

func (d NullifierDomain) Validate() error {
	if d.Version != Version1 || !validDigest(d.ProgramIDDigest) {
		return errors.New("versioned program nullifier domain is required")
	}
	switch d.Kind {
	case ExactProgramDomain:
		if d.RelyingPartyIDDigest != "" {
			return errors.New("program domain must not contain relying-party scope")
		}
	case PairwiseRelyingPartyDomain:
		if !validDigest(d.RelyingPartyIDDigest) {
			return errors.New("pairwise domain requires relying-party digest")
		}
	default:
		return errors.New("unknown nullifier domain")
	}
	return nil
}
func (d NullifierDomain) encode(e *canonicalEncoder) {
	e.u32(d.Version)
	e.text(d.Kind)
	e.text(d.ProgramIDDigest)
	e.text(d.RelyingPartyIDDigest)
}

type NullifierInput struct {
	Version       uint32          `json:"version"`
	Domain        NullifierDomain `json:"domain"`
	PolicyDigest  string          `json:"policy_digest"`
	ProfileDigest string          `json:"profile_digest"`
	KeyEpoch      uint64          `json:"key_epoch"`
	StableInput   string          `json:"stable_input_commitment"`
}

func (i NullifierInput) CanonicalBytes() ([]byte, error) {
	if i.Version != Version1 || i.KeyEpoch == 0 || !validDigests(i.PolicyDigest, i.ProfileDigest, i.StableInput) {
		return nil, errors.New("complete nullifier input binding is required")
	}
	if err := i.Domain.Validate(); err != nil {
		return nil, err
	}
	e := newCanonicalEncoder("virtengine.uniqueness.nullifier-input/v1")
	e.u32(i.Version)
	i.Domain.encode(e)
	e.text(i.PolicyDigest)
	e.text(i.ProfileDigest)
	e.u64(i.KeyEpoch)
	e.text(i.StableInput)
	return e.result(), nil
}

type ScopedNullifier struct {
	Version       uint32          `json:"version"`
	Domain        NullifierDomain `json:"domain"`
	PolicyDigest  string          `json:"policy_digest"`
	ProfileDigest string          `json:"profile_digest"`
	KeyEpoch      uint64          `json:"key_epoch"`
	Value         string          `json:"value"`
}

func (n ScopedNullifier) Validate() error {
	if n.Version != Version1 || n.KeyEpoch == 0 || !validDigests(n.PolicyDigest, n.ProfileDigest, n.Value) {
		return errors.New("complete scoped nullifier digests are required")
	}
	return n.Domain.Validate()
}

// KeyEpoch prevents cross-epoch replay. Production continuity requires reviewed
// key migration; deterministic fixture values may rotate between epochs.
type ThresholdPRF interface {
	Evaluate(context.Context, NullifierInput, *VerifiedFinalUniqueAuthorization) (ScopedNullifier, error)
}
type TemplateArtifact interface{ Commitment() string }
type TemplateTransformer interface {
	Transform(context.Context, CancellableTemplateProfile, []byte) (TemplateArtifact, error)
}

type CandidateReviewState string

const (
	CandidateClear          CandidateReviewState = "clear"
	CandidateReviewRequired CandidateReviewState = "review_required"
)

type CandidateSearchResult struct {
	PossibleMatch          bool                 `json:"possible_match"`
	ReviewState            CandidateReviewState `json:"review_state"`
	CandidateCount         uint32               `json:"candidate_count"`
	CandidateSetCommitment string               `json:"candidate_set_commitment"`
	AdjudicationReference  string               `json:"adjudication_reference"`
}

func (r CandidateSearchResult) Validate() error {
	if r.CandidateCount > MaxCandidateCount || !validDigests(r.CandidateSetCommitment, r.AdjudicationReference) || r.PossibleMatch != (r.CandidateCount > 0) {
		return errors.New("invalid opaque candidate result")
	}
	if r.PossibleMatch && r.ReviewState != CandidateReviewRequired || !r.PossibleMatch && r.ReviewState != CandidateClear {
		return errors.New("inconsistent candidate review state")
	}
	return nil
}

type CandidateSearcher interface {
	Search(context.Context, TemplateArtifact, int64) (CandidateSearchResult, error)
}

type EnrollmentStage string

const (
	StageSearch       EnrollmentStage = "search"
	StageAdjudication EnrollmentStage = "adjudication"
	StageNullifier    EnrollmentStage = "nullifier"
	StageInsertion    EnrollmentStage = "insertion"
)

type EnrollmentTransaction interface {
	Search(context.Context, TemplateArtifact, int64) (CandidateSearchResult, error)
	Advance(EnrollmentStage) error
	Insert(EnrollmentRecord, TemplateArtifact) error
}
type AtomicEnrollmentStore interface {
	Enroll(context.Context, EnrollmentRequest, func(EnrollmentTransaction) (EnrollmentRecord, error)) (EnrollmentRecord, error)
}

type CustodyReceiptPayload struct {
	Version              uint32            `json:"version"`
	ProgramDigest        string            `json:"program_digest"`
	PolicyDigest         string            `json:"policy_digest"`
	ProfileDigest        string            `json:"profile_digest"`
	ModelDigest          string            `json:"model_digest"`
	RuntimeDigest        string            `json:"runtime_digest"`
	EvidenceDigest       string            `json:"evidence_digest"`
	RequestDigest        string            `json:"request_digest"`
	Decision             EnrollmentOutcome `json:"decision"`
	Reason               ReasonCode        `json:"reason"`
	KeyEpoch             uint64            `json:"key_epoch"`
	ParticipantSetDigest string            `json:"participant_set_digest"`
	Threshold            uint32            `json:"threshold"`
	Freshness            Freshness         `json:"freshness"`
	AppealReference      string            `json:"appeal_reference"`
	ScopedNullifier      string            `json:"scoped_nullifier,omitempty"`
}

func (p CustodyReceiptPayload) CanonicalBytes() ([]byte, error) {
	if p.Version != Version1 || p.KeyEpoch == 0 || p.Threshold < 2 || !p.Reason.valid() || !validDigests(p.ProgramDigest, p.PolicyDigest, p.ProfileDigest, p.ModelDigest, p.RuntimeDigest, p.EvidenceDigest, p.RequestDigest, p.ParticipantSetDigest, p.AppealReference) {
		return nil, errors.New("complete custody receipt binding is required")
	}
	if err := p.Freshness.ValidateStructure(); err != nil {
		return nil, err
	}
	if p.Decision == OutcomeFinalUnique {
		if !validDigest(p.ScopedNullifier) {
			return nil, errors.New("final receipt requires nullifier digest")
		}
	} else if p.ScopedNullifier != "" {
		return nil, errors.New("non-final receipt contains nullifier")
	}
	e := newCanonicalEncoder("virtengine.uniqueness.custody-receipt/v1")
	e.u32(p.Version)
	e.text(p.ProgramDigest)
	e.text(p.PolicyDigest)
	e.text(p.ProfileDigest)
	e.text(p.ModelDigest)
	e.text(p.RuntimeDigest)
	e.text(p.EvidenceDigest)
	e.text(p.RequestDigest)
	e.text(string(p.Decision))
	e.text(string(p.Reason))
	e.u64(p.KeyEpoch)
	e.text(p.ParticipantSetDigest)
	e.u32(p.Threshold)
	encodeFreshness(e, p.Freshness)
	e.text(p.AppealReference)
	e.text(p.ScopedNullifier)
	return e.result(), nil
}

type NodeSignature struct {
	NodeID          string `json:"node_id"`
	SigningKeyID    string `json:"signing_key_id"`
	SigningKeyEpoch uint64 `json:"signing_key_epoch"`
	Signature       []byte `json:"signature"`
}
type QuorumAttestation struct {
	Version              uint32          `json:"version"`
	PayloadDigest        string          `json:"payload_digest"`
	ParticipantSetDigest string          `json:"participant_set_digest"`
	Threshold            uint32          `json:"threshold"`
	SignerSetDigest      string          `json:"signer_set_digest"`
	Signatures           []NodeSignature `json:"signatures"`
}
type QuorumAttestor interface {
	Attest(context.Context, CustodyReceiptPayload, ThresholdKeyEpoch) (QuorumAttestation, error)
}
type CustodyKeyResolver interface {
	ResolveCustodyNodes(context.Context, uint64) (CustodyNodeSet, error)
}

func VerifyQuorumAttestation(ctx context.Context, payload CustodyReceiptPayload, attestation QuorumAttestation, resolver CustodyKeyResolver, now time.Time, lifetime, skew time.Duration) error {
	if resolver == nil || payload.KeyEpoch == 0 {
		return errors.New("versioned attestation and key resolver are required")
	}
	if err := payload.Freshness.ValidateAt(now, lifetime, skew); err != nil {
		return err
	}
	value, err := payload.CanonicalBytes()
	if err != nil {
		return err
	}
	nodes, err := resolver.ResolveCustodyNodes(ctx, payload.KeyEpoch)
	if err != nil {
		return err
	}
	return verifyAttestation(value, payload.KeyEpoch, payload.ParticipantSetDigest, payload.Threshold, attestation, nodes)
}

type CompromiseRotationState string

const (
	CompromiseFrozen   CompromiseRotationState = "frozen"
	CompromiseApproved CompromiseRotationState = "approved"
	CompromiseComplete CompromiseRotationState = "complete"
	CompromiseRevoked  CompromiseRotationState = "revoked"
)

type CompromiseRotation struct {
	Version              uint32                  `json:"version"`
	OldEpoch             uint64                  `json:"old_epoch"`
	NewEpoch             uint64                  `json:"new_epoch"`
	CompromiseDigest     string                  `json:"compromise_digest"`
	ParticipantSetDigest string                  `json:"participant_set_digest"`
	Threshold            uint32                  `json:"threshold"`
	FreezeCoordinate     uint64                  `json:"freeze_coordinate"`
	ActivationCoordinate uint64                  `json:"activation_coordinate"`
	ReenrollmentRequired bool                    `json:"reenrollment_required"`
	State                CompromiseRotationState `json:"state"`
	Approvals            []NodeSignature         `json:"approvals"`
}

func (r CompromiseRotation) CanonicalBytes() ([]byte, error) {
	if r.Version != Version1 || r.OldEpoch == 0 || r.NewEpoch != r.OldEpoch+1 || r.Threshold < 2 || !validDigests(r.CompromiseDigest, r.ParticipantSetDigest) || r.FreezeCoordinate == 0 || r.ActivationCoordinate < r.FreezeCoordinate || !r.ReenrollmentRequired {
		return nil, errors.New("complete monotonic compromise rotation is required")
	}
	switch r.State {
	case CompromiseFrozen, CompromiseApproved, CompromiseComplete, CompromiseRevoked:
	default:
		return nil, errors.New("unknown compromise rotation state")
	}
	e := newCanonicalEncoder("virtengine.uniqueness.compromise-rotation/v1")
	e.u32(r.Version)
	e.u64(r.OldEpoch)
	e.u64(r.NewEpoch)
	e.text(r.CompromiseDigest)
	e.text(r.ParticipantSetDigest)
	e.u32(r.Threshold)
	e.u64(r.FreezeCoordinate)
	e.u64(r.ActivationCoordinate)
	e.boolean(r.ReenrollmentRequired)
	e.text(string(r.State))
	return e.result(), nil
}
func ValidateCompromiseTransition(previous, next CompromiseRotation) error {
	if _, err := previous.CanonicalBytes(); err != nil {
		return err
	}
	if _, err := next.CanonicalBytes(); err != nil {
		return err
	}
	if previous.Version != next.Version || previous.OldEpoch != next.OldEpoch || previous.NewEpoch != next.NewEpoch || previous.CompromiseDigest != next.CompromiseDigest || previous.ParticipantSetDigest != next.ParticipantSetDigest || previous.Threshold != next.Threshold || previous.FreezeCoordinate != next.FreezeCoordinate || previous.ActivationCoordinate != next.ActivationCoordinate || previous.ReenrollmentRequired != next.ReenrollmentRequired {
		return errors.New("compromise rotation binding changed")
	}
	allowed := previous.State == CompromiseFrozen && next.State == CompromiseApproved || previous.State == CompromiseApproved && next.State == CompromiseComplete || previous.State != CompromiseRevoked && next.State == CompromiseRevoked
	if !allowed {
		return errors.New("non-monotonic compromise rotation transition")
	}
	return nil
}

type VerifiedCompromiseRotation struct{ rotation CompromiseRotation }

func VerifyCompromiseRotation(rotation CompromiseRotation, nodes CustodyNodeSet) (*VerifiedCompromiseRotation, error) {
	value, err := rotation.CanonicalBytes()
	if err != nil {
		return nil, err
	}
	if rotation.State != CompromiseComplete {
		return nil, errors.New("compromise rotation is not complete")
	}
	setDigest, err := nodes.Digest()
	if err != nil {
		return nil, err
	}
	if rotation.ParticipantSetDigest != setDigest || rotation.Threshold != nodes.Threshold {
		return nil, errors.New("rotation does not bind frozen old-epoch authority")
	}
	if err := verifyNodeSignatures(value, rotation.OldEpoch, rotation.Threshold, rotation.Approvals, nodes); err != nil {
		return nil, err
	}
	return &VerifiedCompromiseRotation{rotation: rotation}, nil
}
func (verified *VerifiedCompromiseRotation) CanIssue(epoch, coordinate uint64) error {
	if verified == nil {
		return errors.New("verified compromise rotation is required")
	}
	if epoch == verified.rotation.OldEpoch && coordinate >= verified.rotation.FreezeCoordinate {
		return errors.New("old epoch is frozen")
	}
	if epoch != verified.rotation.NewEpoch || coordinate < verified.rotation.ActivationCoordinate {
		return errors.New("issuance is unavailable during compromise rotation")
	}
	return nil
}

type FinalUniqueAuthorization struct {
	Version               uint32            `json:"version"`
	RequestDigest         string            `json:"request_digest"`
	ProgramDigest         string            `json:"program_digest"`
	PolicyDigest          string            `json:"policy_digest"`
	ProfileDigest         string            `json:"profile_digest"`
	StableInputCommitment string            `json:"stable_input_commitment"`
	Decision              EnrollmentOutcome `json:"decision"`
	KeyEpoch              uint64            `json:"key_epoch"`
	ProfileEpoch          uint64            `json:"profile_epoch"`
	Freshness             Freshness         `json:"freshness"`
	ParticipantSetDigest  string            `json:"participant_set_digest"`
	Threshold             uint32            `json:"threshold"`
	Attestation           QuorumAttestation `json:"attestation"`
}

func (a FinalUniqueAuthorization) CanonicalBytes() ([]byte, error) {
	if a.Version != Version1 || a.Decision != OutcomeFinalUnique || a.KeyEpoch == 0 || a.ProfileEpoch == 0 || a.Threshold < 2 || !validDigests(a.RequestDigest, a.ProgramDigest, a.PolicyDigest, a.ProfileDigest, a.StableInputCommitment, a.ParticipantSetDigest) {
		return nil, errors.New("complete final-unique authorization is required")
	}
	if err := a.Freshness.ValidateStructure(); err != nil {
		return nil, err
	}
	e := newCanonicalEncoder("virtengine.uniqueness.final-unique-authorization/v1")
	e.u32(a.Version)
	e.text(a.RequestDigest)
	e.text(a.ProgramDigest)
	e.text(a.PolicyDigest)
	e.text(a.ProfileDigest)
	e.text(a.StableInputCommitment)
	e.text(string(a.Decision))
	e.u64(a.KeyEpoch)
	e.u64(a.ProfileEpoch)
	encodeFreshness(e, a.Freshness)
	e.text(a.ParticipantSetDigest)
	e.u32(a.Threshold)
	return e.result(), nil
}

type VerifiedFinalUniqueAuthorization struct{ inputDigest string }

func VerifyFinalUniqueAuthorization(a FinalUniqueAuthorization, nodes CustodyNodeSet, now time.Time, lifetime, skew time.Duration) (*VerifiedFinalUniqueAuthorization, error) {
	if err := a.Freshness.ValidateAt(now, lifetime, skew); err != nil {
		return nil, err
	}
	value, err := a.CanonicalBytes()
	if err != nil {
		return nil, err
	}
	if err := verifyAttestation(value, a.KeyEpoch, a.ParticipantSetDigest, a.Threshold, a.Attestation, nodes); err != nil {
		return nil, err
	}
	input := NullifierInput{Version: Version1, Domain: NullifierDomain{Version: Version1, Kind: ExactProgramDomain, ProgramIDDigest: a.ProgramDigest}, PolicyDigest: a.PolicyDigest, ProfileDigest: a.ProfileDigest, KeyEpoch: a.KeyEpoch, StableInput: a.StableInputCommitment}
	inputBytes, err := input.CanonicalBytes()
	if err != nil {
		return nil, err
	}
	return &VerifiedFinalUniqueAuthorization{inputDigest: digest(inputBytes)}, nil
}

func verifyAttestation(payload []byte, epoch uint64, participant string, threshold uint32, attestation QuorumAttestation, nodes CustodyNodeSet) error {
	if attestation.Version != Version1 || !validDigests(attestation.PayloadDigest, attestation.ParticipantSetDigest, attestation.SignerSetDigest) {
		return errors.New("invalid quorum attestation")
	}
	setDigest, err := nodes.Digest()
	if err != nil {
		return err
	}
	if participant != setDigest || attestation.ParticipantSetDigest != setDigest || threshold != nodes.Threshold || attestation.Threshold != threshold {
		return errors.New("attestation participant set or threshold was substituted")
	}
	if attestation.PayloadDigest != digest(payload) {
		return errors.New("attested payload was tampered")
	}
	signerDigest, err := signerSetDigest(attestation.Signatures)
	if err != nil {
		return err
	}
	if signerDigest != attestation.SignerSetDigest {
		return errors.New("quorum signer set was substituted")
	}
	return verifyNodeSignatures(attestationSigningBytes(attestation.PayloadDigest, setDigest, threshold, signerDigest), epoch, threshold, attestation.Signatures, nodes)
}
func verifyNodeSignatures(message []byte, epoch uint64, threshold uint32, signatures []NodeSignature, nodes CustodyNodeSet) error {
	if uint32(len(signatures)) < threshold {
		return errors.New("insufficient quorum signatures")
	}
	byID := map[string]CustodyNodeIdentity{}
	for _, node := range nodes.Nodes {
		byID[node.NodeID] = node
	}
	seen := map[string]bool{}
	for _, signature := range signatures {
		node, ok := byID[signature.NodeID]
		if !validOpaqueID(signature.NodeID) || !validOpaqueID(signature.SigningKeyID) || len(signature.Signature) != ed25519.SignatureSize || !ok || node.State != NodeActive || seen[signature.NodeID] || signature.SigningKeyID != node.SigningKeyID || signature.SigningKeyEpoch != epoch || node.SigningKeyEpoch != epoch {
			return errors.New("unknown, duplicate, stale, or aliased quorum signer")
		}
		seen[signature.NodeID] = true
		if !ed25519.Verify(node.SigningPublicKey, message, signature.Signature) {
			return errors.New("invalid quorum signature")
		}
	}
	return nil
}
func signerSetDigest(signatures []NodeSignature) (string, error) {
	ids := make([]string, 0, len(signatures))
	seen := map[string]bool{}
	for _, signature := range signatures {
		if !validOpaqueID(signature.NodeID) || seen[signature.NodeID] {
			return "", errors.New("duplicate or noncanonical signer")
		}
		seen[signature.NodeID] = true
		ids = append(ids, signature.NodeID)
	}
	slices.Sort(ids)
	e := newCanonicalEncoder("virtengine.uniqueness.signer-set/v1")
	e.u32(uint32(len(ids)))
	for _, id := range ids {
		e.text(id)
	}
	return digest(e.result()), nil
}

type canonicalEncoder struct{ data []byte }

func newCanonicalEncoder(domain string) *canonicalEncoder {
	e := &canonicalEncoder{}
	e.text(domain)
	return e
}
func (e *canonicalEncoder) u32(value uint32) {
	var data [4]byte
	binary.BigEndian.PutUint32(data[:], value)
	e.data = append(e.data, data[:]...)
}
func (e *canonicalEncoder) u64(value uint64) {
	var data [8]byte
	binary.BigEndian.PutUint64(data[:], value)
	e.data = append(e.data, data[:]...)
}
func (e *canonicalEncoder) i64(value int64) { e.u64(uint64(value)) }
func (e *canonicalEncoder) boolean(value bool) {
	if value {
		e.data = append(e.data, 1)
	} else {
		e.data = append(e.data, 0)
	}
}
func (e *canonicalEncoder) text(value string) { e.bytes([]byte(value)) }
func (e *canonicalEncoder) bytes(value []byte) {
	e.u64(uint64(len(value)))
	e.data = append(e.data, value...)
}
func (e *canonicalEncoder) result() []byte { return slices.Clone(e.data) }
func encodeFreshness(e *canonicalEncoder, f Freshness) {
	e.i64(f.IssuedAt)
	e.i64(f.ExpiresAt)
	e.text(f.Nonce)
}
func digest(value []byte) string    { sum := sha256.Sum256(value); return hex.EncodeToString(sum[:]) }
func validDigest(value string) bool { return digestPattern.MatchString(value) }
func validDigests(values ...string) bool {
	for _, value := range values {
		if !validDigest(value) {
			return false
		}
	}
	return true
}
func validOpaqueID(value string) bool { return opaqueIDPattern.MatchString(value) }
func compareString(left, right string) int {
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}
func attestationSigningBytes(payload, participant string, threshold uint32, signers string) []byte {
	e := newCanonicalEncoder("virtengine.uniqueness.quorum-attestation/v1")
	e.text(payload)
	e.text(participant)
	e.u32(threshold)
	e.text(signers)
	return e.result()
}
func requireSameRequest(request EnrollmentRequest, record EnrollmentRecord) error {
	if request.RequestDigest != record.RequestDigest || request.EvidenceDigest != record.EvidenceDigest || request.ModelDigest != record.ModelDigest || request.RuntimeDigest != record.RuntimeDigest || request.ProfileDigest != record.ProfileDigest || request.ProgramIDDigest != record.ProgramIDDigest || request.PolicyDigest != record.PolicyDigest || request.IdempotencyKey != record.IdempotencyKey || request.KeyEpoch != record.KeyEpoch || request.Freshness != record.Freshness {
		return fmt.Errorf("enrollment record does not bind request: %w", ErrConflict)
	}
	return nil
}
