// Package issuerlink defines fixture-only contracts for privacy-preserving issuer links.
package issuerlink

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"regexp"
)

const (
	Version1                uint32 = 1
	FixtureOnlyState               = "fixture_only"
	ProductionUnavailable          = "production_unavailable"
	ExternalBlockerRequired        = "external_security_review_required"
	ProgramDomainKind              = "program_scoped"
	PairwiseDomainKind             = "relying_party_pairwise"
)

var (
	ErrConflict              = errors.New("issuer-link idempotency conflict")
	ErrDuplicateLink         = errors.New("duplicate issuer link")
	ErrCrossDomainReplay     = errors.New("cross-domain challenge replay")
	ErrProductionUnavailable = errors.New("production issuer linking is unavailable")
	ErrStale                 = errors.New("stale or expired issuer-link request")
	ErrUnauthorized          = errors.New("issuer-link authorization is invalid")
	digestPattern            = regexp.MustCompile(`^[0-9a-f]{64}$`)
	opaqueIDPattern          = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,127}$`)
)

type NullifierDomain struct {
	Version            uint32 `json:"version"`
	Kind               string `json:"kind"`
	ProgramDigest      string `json:"program_digest"`
	RelyingPartyDigest string `json:"relying_party_digest,omitempty"`
}

func (d NullifierDomain) Validate() error {
	if d.Version != Version1 || !validDigest(d.ProgramDigest) {
		return errors.New("versioned program-scoped nullifier domain is required")
	}
	switch d.Kind {
	case ProgramDomainKind:
		if d.RelyingPartyDigest != "" {
			return errors.New("program domain must not include relying-party scope")
		}
	case PairwiseDomainKind:
		if !validDigest(d.RelyingPartyDigest) {
			return errors.New("pairwise domain requires a relying-party digest")
		}
	default:
		return errors.New("unknown nullifier domain")
	}
	return nil
}

func (d NullifierDomain) encode(e *canonicalEncoder) {
	e.u32(d.Version)
	e.text(d.Kind)
	e.text(d.ProgramDigest)
	e.text(d.RelyingPartyDigest)
}

func (d NullifierDomain) Key() (string, error) {
	if err := d.Validate(); err != nil {
		return "", err
	}
	e := newCanonicalEncoder("virtengine.issuer-link.domain/v1")
	d.encode(e)
	return digest(e.result()), nil
}

// StableInputCommitment keeps confidential adapter input opaque after construction.
type StableInputCommitment struct{ value []byte }

func NewStableInputCommitment(value []byte) (StableInputCommitment, error) {
	if len(value) < 16 {
		return StableInputCommitment{}, errors.New("confidential stable input commitment is required")
	}
	return StableInputCommitment{value: append([]byte(nil), value...)}, nil
}

type NullifierInput struct {
	Domain           NullifierDomain `json:"domain"`
	IssuerCommitment string          `json:"issuer_commitment"`
	PolicyDigest     string          `json:"policy_digest"`
	ProfileDigest    string          `json:"profile_digest"`
	IssuerKeyEpoch   uint64          `json:"issuer_key_epoch"`
	RequestDigest    string          `json:"request_digest"`
	stableCommitment []byte
}

func newNullifierInput(request WalletLinkRequest, epoch uint64, stable StableInputCommitment) (NullifierInput, error) {
	input := NullifierInput{Domain: request.Domain, IssuerCommitment: request.IssuerCommitment, PolicyDigest: request.PolicyDigest, ProfileDigest: request.ProfileDigest, IssuerKeyEpoch: epoch, RequestDigest: request.RequestDigest, stableCommitment: append([]byte(nil), stable.value...)}
	if err := input.validate(); err != nil {
		return NullifierInput{}, err
	}
	return input, nil
}

func (i NullifierInput) validate() error {
	if err := i.Domain.Validate(); err != nil {
		return err
	}
	if !validDigests(i.IssuerCommitment, i.PolicyDigest, i.ProfileDigest, i.RequestDigest) || i.IssuerKeyEpoch == 0 || len(i.stableCommitment) < 16 {
		return errors.New("complete strict nullifier input is required")
	}
	return nil
}

func (i NullifierInput) bindingDigest() (string, error) {
	if err := i.validate(); err != nil {
		return "", err
	}
	e := newCanonicalEncoder("virtengine.issuer-link.nullifier-input/v1")
	i.Domain.encode(e)
	e.text(i.IssuerCommitment)
	e.text(i.PolicyDigest)
	e.text(i.ProfileDigest)
	e.u64(i.IssuerKeyEpoch)
	e.text(i.RequestDigest)
	e.bytes(i.stableCommitment)
	return digest(e.result()), nil
}

// VerifiedNullifierAuthorization is an opaque exact-input capability.
type VerifiedNullifierAuthorization struct {
	binding string
	nonce   string
}

func (a VerifiedNullifierAuthorization) validate(input NullifierInput) error {
	binding, err := input.bindingDigest()
	if err != nil || a.binding == "" || a.binding != binding || !validDigest(a.nonce) {
		return ErrUnauthorized
	}
	return nil
}

type ThresholdPRF interface {
	Evaluate(context.Context, VerifiedNullifierAuthorization, NullifierInput) (string, error)
}

type IssuerVisibility string

const (
	VisibilityProgramScoped IssuerVisibility = ProgramDomainKind
	VisibilityPairwise      IssuerVisibility = PairwiseDomainKind
)

type IssuerState string

const (
	IssuerFixtureOnly IssuerState = FixtureOnlyState
	IssuerRevoked     IssuerState = "revoked"
	IssuerRotating    IssuerState = "rotating"
)

type IssuerProfile struct {
	Version               uint32           `json:"version"`
	ProfileID             string           `json:"profile_id"`
	IssuerCommitment      string           `json:"issuer_commitment"`
	ProgramDigest         string           `json:"program_digest"`
	ProfileDigest         string           `json:"profile_digest"`
	PolicyDigest          string           `json:"policy_digest"`
	SigningKeyID          string           `json:"signing_key_id"`
	KeyEpoch              uint64           `json:"key_epoch"`
	RetentionPolicyDigest string           `json:"retention_policy_digest"`
	DeletionPolicyDigest  string           `json:"deletion_policy_digest"`
	Visibility            IssuerVisibility `json:"visibility"`
	State                 IssuerState      `json:"state"`
	AttestationCapability bool             `json:"attestation_capability"`
	ExternalBlocker       string           `json:"external_blocker"`
}

func (p IssuerProfile) Validate() error {
	if p.Version != Version1 || !validOpaqueID(p.ProfileID) || !validOpaqueID(p.SigningKeyID) || p.KeyEpoch == 0 {
		return errors.New("canonical issuer profile identifiers and key epoch are required")
	}
	if !validDigests(p.IssuerCommitment, p.ProgramDigest, p.ProfileDigest, p.PolicyDigest, p.RetentionPolicyDigest, p.DeletionPolicyDigest) {
		return errors.New("strict issuer profile commitments are required")
	}
	if p.Visibility != VisibilityProgramScoped && p.Visibility != VisibilityPairwise {
		return errors.New("unknown issuer visibility")
	}
	if p.State != IssuerFixtureOnly && p.State != IssuerRevoked && p.State != IssuerRotating {
		return errors.New("unknown issuer state")
	}
	if p.ExternalBlocker != ExternalBlockerRequired {
		return ErrProductionUnavailable
	}
	if p.State == IssuerFixtureOnly && !p.AttestationCapability {
		return errors.New("governed issuer attestation capability is required")
	}
	return nil
}

func (p IssuerProfile) CanAttest(program, profile, policy string, epoch uint64) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if p.State != IssuerFixtureOnly || !p.AttestationCapability {
		return errors.New("issuer is not authorized to attest")
	}
	if p.ProgramDigest != program || p.ProfileDigest != profile || p.PolicyDigest != policy || p.KeyEpoch != epoch {
		return errors.New("issuer authority does not bind request coordinates")
	}
	return nil
}

type WalletLinkRequest struct {
	Version                 uint32          `json:"version"`
	RequestDigest           string          `json:"request_digest"`
	IdempotencyKey          string          `json:"idempotency_key"`
	IssuerCommitment        string          `json:"issuer_commitment"`
	Domain                  NullifierDomain `json:"domain"`
	PolicyDigest            string          `json:"policy_digest"`
	ProfileDigest           string          `json:"profile_digest"`
	OldWalletKeyCommitment  string          `json:"old_wallet_key_commitment,omitempty"`
	NewWalletKeyCommitment  string          `json:"new_wallet_key_commitment"`
	OldWalletKeyEpoch       uint64          `json:"old_wallet_key_epoch,omitempty"`
	NewWalletKeyEpoch       uint64          `json:"new_wallet_key_epoch"`
	ChallengeNonceDigest    string          `json:"challenge_nonce_digest"`
	EvidenceDigest          string          `json:"evidence_digest"`
	CooldownUntilCoordinate uint64          `json:"cooldown_until_coordinate"`
	ExpiresAtCoordinate     uint64          `json:"expires_at_coordinate"`
	CurrentCoordinate       uint64          `json:"current_coordinate"`
}

func (r WalletLinkRequest) IsRotation() bool { return r.OldWalletKeyCommitment != "" }

func (r WalletLinkRequest) ValidateStructure() error {
	if r.Version != Version1 || !validOpaqueID(r.IdempotencyKey) || r.NewWalletKeyEpoch == 0 {
		return errors.New("canonical wallet-link request identifiers are required")
	}
	if err := r.Domain.Validate(); err != nil {
		return err
	}
	if !validDigests(r.RequestDigest, r.IssuerCommitment, r.PolicyDigest, r.ProfileDigest, r.NewWalletKeyCommitment, r.ChallengeNonceDigest, r.EvidenceDigest) {
		return errors.New("strict wallet-link request commitments are required")
	}
	if r.CooldownUntilCoordinate <= r.CurrentCoordinate || r.ExpiresAtCoordinate <= r.CooldownUntilCoordinate {
		return errors.New("cooldown and expiry coordinates must be monotonic")
	}
	if r.IsRotation() {
		if !validDigest(r.OldWalletKeyCommitment) || r.OldWalletKeyEpoch == 0 || r.NewWalletKeyEpoch <= r.OldWalletKeyEpoch || r.OldWalletKeyCommitment == r.NewWalletKeyCommitment {
			return errors.New("rotation requires distinct old and advancing new wallet-key commitments")
		}
	} else if r.OldWalletKeyEpoch != 0 {
		return errors.New("initial link must not include old wallet-key coordinates")
	}
	expected, err := r.computeDigest()
	if err != nil || r.RequestDigest != expected {
		return errors.New("request digest does not match canonical request")
	}
	return nil
}

func (r WalletLinkRequest) ValidateAt(coordinate uint64) error {
	if err := r.ValidateStructure(); err != nil {
		return err
	}
	if coordinate != r.CurrentCoordinate || coordinate >= r.ExpiresAtCoordinate {
		return ErrStale
	}
	return nil
}

func (r WalletLinkRequest) canonicalBytes(includeDigest bool) ([]byte, error) {
	if r.Version != Version1 || !validOpaqueID(r.IdempotencyKey) || r.NewWalletKeyEpoch == 0 || r.Domain.Validate() != nil {
		return nil, errors.New("invalid request structure")
	}
	e := newCanonicalEncoder("virtengine.issuer-link.request/v1")
	e.u32(r.Version)
	if includeDigest {
		e.text(r.RequestDigest)
	}
	e.text(r.IdempotencyKey)
	e.text(r.IssuerCommitment)
	r.Domain.encode(e)
	e.text(r.PolicyDigest)
	e.text(r.ProfileDigest)
	e.text(r.OldWalletKeyCommitment)
	e.text(r.NewWalletKeyCommitment)
	e.u64(r.OldWalletKeyEpoch)
	e.u64(r.NewWalletKeyEpoch)
	e.text(r.ChallengeNonceDigest)
	e.text(r.EvidenceDigest)
	e.u64(r.CooldownUntilCoordinate)
	e.u64(r.ExpiresAtCoordinate)
	e.u64(r.CurrentCoordinate)
	return e.result(), nil
}

func (r WalletLinkRequest) computeDigest() (string, error) {
	value, err := r.canonicalBytes(false)
	if err != nil {
		return "", err
	}
	return digest(value), nil
}

func (r WalletLinkRequest) WithComputedDigest() (WalletLinkRequest, error) {
	r.RequestDigest = ""
	value, err := r.computeDigest()
	if err != nil {
		return WalletLinkRequest{}, err
	}
	r.RequestDigest = value
	return r, nil
}

func (r WalletLinkRequest) SigningBytes() ([]byte, error) {
	if err := r.ValidateStructure(); err != nil {
		return nil, err
	}
	return r.canonicalBytes(true)
}

type WalletAuthorizations struct {
	OldWalletPublicKey []byte `json:"old_wallet_public_key,omitempty"`
	OldWalletSignature []byte `json:"old_wallet_signature,omitempty"`
	NewWalletPublicKey []byte `json:"new_wallet_public_key"`
	NewWalletSignature []byte `json:"new_wallet_signature"`
}

func VerifyWalletAuthorizations(request WalletLinkRequest, auth WalletAuthorizations) error {
	message, err := request.SigningBytes()
	if err != nil {
		return err
	}
	if !publicKeyCommitmentMatches(auth.NewWalletPublicKey, request.NewWalletKeyCommitment) || len(auth.NewWalletSignature) != ed25519.SignatureSize || !ed25519.Verify(auth.NewWalletPublicKey, message, auth.NewWalletSignature) {
		return errors.New("valid new-wallet authorization is required")
	}
	if request.IsRotation() {
		if !publicKeyCommitmentMatches(auth.OldWalletPublicKey, request.OldWalletKeyCommitment) || len(auth.OldWalletSignature) != ed25519.SignatureSize || !ed25519.Verify(auth.OldWalletPublicKey, message, auth.OldWalletSignature) {
			return errors.New("valid old-wallet authorization is required for rotation")
		}
	} else if len(auth.OldWalletPublicKey) != 0 || len(auth.OldWalletSignature) != 0 {
		return errors.New("initial link must not carry old-wallet authorization")
	}
	return nil
}

func WalletKeyCommitment(publicKey ed25519.PublicKey) (string, error) {
	if len(publicKey) != ed25519.PublicKeySize {
		return "", errors.New("Ed25519 wallet public key is required")
	}
	e := newCanonicalEncoder("virtengine.issuer-link.wallet-key/v1")
	e.bytes(publicKey)
	return digest(e.result()), nil
}

func publicKeyCommitmentMatches(publicKey []byte, commitment string) bool {
	value, err := WalletKeyCommitment(ed25519.PublicKey(publicKey))
	return err == nil && value == commitment
}

type IssuerAttestation struct {
	Version             uint32 `json:"version"`
	RequestDigest       string `json:"request_digest"`
	IssuerCommitment    string `json:"issuer_commitment"`
	SigningKeyID        string `json:"signing_key_id"`
	SigningKeyEpoch     uint64 `json:"signing_key_epoch"`
	IssuedAtCoordinate  uint64 `json:"issued_at_coordinate"`
	ExpiresAtCoordinate uint64 `json:"expires_at_coordinate"`
	NonceDigest         string `json:"nonce_digest"`
	Signature           []byte `json:"signature"`
}

func (a IssuerAttestation) signingBytes() ([]byte, error) {
	if a.Version != Version1 || a.SigningKeyEpoch == 0 || !validOpaqueID(a.SigningKeyID) || !validDigests(a.RequestDigest, a.IssuerCommitment, a.NonceDigest) || a.ExpiresAtCoordinate <= a.IssuedAtCoordinate {
		return nil, errors.New("complete issuer attestation binding is required")
	}
	e := newCanonicalEncoder("virtengine.issuer-link.issuer-attestation/v1")
	e.u32(a.Version)
	e.text(a.RequestDigest)
	e.text(a.IssuerCommitment)
	e.text(a.SigningKeyID)
	e.u64(a.SigningKeyEpoch)
	e.u64(a.IssuedAtCoordinate)
	e.u64(a.ExpiresAtCoordinate)
	e.text(a.NonceDigest)
	return e.result(), nil
}

type AttestationPolicy struct {
	MaxLifetimeCoordinates uint64
	MaxFutureSkew          uint64
}

type IssuerKeyResolver interface {
	ResolveIssuerKey(context.Context, string, string, uint64) (IssuerProfile, ed25519.PublicKey, error)
}

func VerifyIssuerAttestation(ctx context.Context, request WalletLinkRequest, attestation IssuerAttestation, resolver IssuerKeyResolver, coordinate uint64, policy AttestationPolicy, stable StableInputCommitment) (VerifiedNullifierAuthorization, NullifierInput, error) {
	if resolver == nil || policy.MaxLifetimeCoordinates == 0 || attestation.RequestDigest != request.RequestDigest || attestation.IssuerCommitment != request.IssuerCommitment || attestation.NonceDigest != request.ChallengeNonceDigest {
		return VerifiedNullifierAuthorization{}, NullifierInput{}, errors.New("issuer attestation does not bind the exact request and challenge")
	}
	if coordinate != request.CurrentCoordinate || attestation.ExpiresAtCoordinate <= attestation.IssuedAtCoordinate || (attestation.IssuedAtCoordinate > coordinate && attestation.IssuedAtCoordinate-coordinate > policy.MaxFutureSkew) || coordinate >= attestation.ExpiresAtCoordinate || attestation.ExpiresAtCoordinate > request.ExpiresAtCoordinate || attestation.ExpiresAtCoordinate-attestation.IssuedAtCoordinate > policy.MaxLifetimeCoordinates {
		return VerifiedNullifierAuthorization{}, NullifierInput{}, ErrStale
	}
	profile, publicKey, err := resolver.ResolveIssuerKey(ctx, attestation.IssuerCommitment, attestation.SigningKeyID, attestation.SigningKeyEpoch)
	if err != nil {
		return VerifiedNullifierAuthorization{}, NullifierInput{}, err
	}
	if profile.IssuerCommitment != request.IssuerCommitment || profile.SigningKeyID != attestation.SigningKeyID {
		return VerifiedNullifierAuthorization{}, NullifierInput{}, errors.New("issuer key resolver returned substituted authority")
	}
	if err := profile.CanAttest(request.Domain.ProgramDigest, request.ProfileDigest, request.PolicyDigest, attestation.SigningKeyEpoch); err != nil {
		return VerifiedNullifierAuthorization{}, NullifierInput{}, err
	}
	if profile.Visibility == VisibilityProgramScoped && request.Domain.Kind != ProgramDomainKind || profile.Visibility == VisibilityPairwise && request.Domain.Kind != PairwiseDomainKind {
		return VerifiedNullifierAuthorization{}, NullifierInput{}, errors.New("request domain violates governed issuer visibility")
	}
	message, err := attestation.signingBytes()
	if err != nil || len(attestation.Signature) != ed25519.SignatureSize || !ed25519.Verify(publicKey, message, attestation.Signature) {
		return VerifiedNullifierAuthorization{}, NullifierInput{}, errors.New("invalid issuer attestation signature")
	}
	input, err := newNullifierInput(request, attestation.SigningKeyEpoch, stable)
	if err != nil {
		return VerifiedNullifierAuthorization{}, NullifierInput{}, err
	}
	binding, _ := input.bindingDigest()
	return VerifiedNullifierAuthorization{binding: binding, nonce: attestation.NonceDigest}, input, nil
}

type LinkStatus string

const (
	StatusPending         LinkStatus = "pending"
	StatusCooldown        LinkStatus = "cooldown"
	StatusActive          LinkStatus = "active"
	StatusSuperseded      LinkStatus = "superseded"
	StatusAppealed        LinkStatus = "appealed"
	StatusRevoked         LinkStatus = "revoked"
	StatusDeletionPending LinkStatus = "deletion_pending"
	StatusDeleted         LinkStatus = "deleted"
)

type RetentionState string

const (
	RetentionRequired  RetentionState = "retained"
	RetentionLegalHold RetentionState = "legal_hold"
	RetentionEligible  RetentionState = "deletion_eligible"
	RetentionDeleted   RetentionState = "deleted"
)

type IssuerLinkRecord struct {
	Version                        uint32          `json:"version"`
	ScopedNullifier                string          `json:"scoped_nullifier"`
	Domain                         NullifierDomain `json:"domain"`
	IssuerCommitment               string          `json:"issuer_commitment"`
	PolicyDigest                   string          `json:"policy_digest"`
	ProfileDigest                  string          `json:"profile_digest"`
	WalletKeyCommitment            string          `json:"wallet_key_commitment"`
	WalletKeyEpoch                 uint64          `json:"wallet_key_epoch"`
	AuthorizedRequestDigest        string          `json:"authorized_request_digest"`
	PredecessorNullifierCommitment string          `json:"predecessor_nullifier_commitment,omitempty"`
	RotationRequestDigest          string          `json:"rotation_request_digest,omitempty"`
	Status                         LinkStatus      `json:"status"`
	CreatedAtUnix                  int64           `json:"created_at_unix"`
	UpdatedAtUnix                  int64           `json:"updated_at_unix"`
	CreatedAtCoordinate            uint64          `json:"created_at_coordinate"`
	UpdatedAtCoordinate            uint64          `json:"updated_at_coordinate"`
	CooldownUntilCoordinate        uint64          `json:"cooldown_until_coordinate"`
	SupersessionCommitment         string          `json:"supersession_commitment,omitempty"`
	NotificationCommitment         string          `json:"notification_commitment,omitempty"`
	AppealCommitment               string          `json:"appeal_commitment,omitempty"`
	Retention                      RetentionState  `json:"retention_state"`
}

func (r IssuerLinkRecord) Validate() error {
	if r.Version != Version1 || r.WalletKeyEpoch == 0 || !validDigests(r.ScopedNullifier, r.IssuerCommitment, r.PolicyDigest, r.ProfileDigest, r.WalletKeyCommitment, r.AuthorizedRequestDigest) {
		return errors.New("complete public issuer-link commitments are required")
	}
	if err := r.Domain.Validate(); err != nil {
		return err
	}
	if r.CreatedAtUnix <= 0 || r.UpdatedAtUnix < r.CreatedAtUnix || r.UpdatedAtCoordinate < r.CreatedAtCoordinate || r.CooldownUntilCoordinate <= r.CreatedAtCoordinate {
		return errors.New("invalid issuer-link record coordinates")
	}
	if (r.PredecessorNullifierCommitment == "") != (r.RotationRequestDigest == "") {
		return errors.New("rotation lineage must be complete")
	}
	if r.PredecessorNullifierCommitment != "" && !validDigests(r.PredecessorNullifierCommitment, r.RotationRequestDigest) {
		return errors.New("strict rotation lineage digests are required")
	}
	switch r.Status {
	case StatusPending, StatusCooldown, StatusActive, StatusRevoked:
		if r.SupersessionCommitment != "" || r.NotificationCommitment != "" || r.AppealCommitment != "" {
			return errors.New("record contains unrelated lifecycle commitments")
		}
	case StatusSuperseded:
		if !validDigests(r.SupersessionCommitment, r.NotificationCommitment) || r.AppealCommitment != "" {
			return errors.New("superseded record requires strict commitments")
		}
	case StatusAppealed:
		if !validDigest(r.AppealCommitment) || r.SupersessionCommitment != "" || r.NotificationCommitment != "" {
			return errors.New("appealed record requires appeal commitment")
		}
	case StatusDeletionPending:
		if r.Retention != RetentionEligible {
			return errors.New("deletion requires eligible retention state")
		}
	default:
		return errors.New("unknown live issuer-link status")
	}
	if r.Retention != RetentionRequired && r.Retention != RetentionLegalHold && r.Retention != RetentionEligible {
		return errors.New("unknown live retention state")
	}
	return nil
}

func (r IssuerLinkRecord) commitment() (string, error) {
	if err := r.Validate(); err != nil {
		return "", err
	}
	e := newCanonicalEncoder("virtengine.issuer-link.record/v1")
	e.u32(r.Version)
	e.text(r.ScopedNullifier)
	r.Domain.encode(e)
	e.text(r.IssuerCommitment)
	e.text(r.PolicyDigest)
	e.text(r.ProfileDigest)
	e.text(r.WalletKeyCommitment)
	e.u64(r.WalletKeyEpoch)
	e.text(r.AuthorizedRequestDigest)
	e.text(r.PredecessorNullifierCommitment)
	e.text(r.RotationRequestDigest)
	e.text(string(r.Status))
	e.u64(uint64(r.CreatedAtUnix))
	e.u64(uint64(r.UpdatedAtUnix))
	e.u64(r.CreatedAtCoordinate)
	e.u64(r.UpdatedAtCoordinate)
	e.u64(r.CooldownUntilCoordinate)
	e.text(r.SupersessionCommitment)
	e.text(r.NotificationCommitment)
	e.text(r.AppealCommitment)
	e.text(string(r.Retention))
	return digest(e.result()), nil
}

func validateStatusTransition(previous, next IssuerLinkRecord) error {
	if err := previous.Validate(); err != nil {
		return err
	}
	if err := next.Validate(); err != nil {
		return err
	}
	if previous.Version != next.Version || previous.ScopedNullifier != next.ScopedNullifier || previous.Domain != next.Domain || previous.IssuerCommitment != next.IssuerCommitment || previous.PolicyDigest != next.PolicyDigest || previous.ProfileDigest != next.ProfileDigest || previous.WalletKeyCommitment != next.WalletKeyCommitment || previous.WalletKeyEpoch != next.WalletKeyEpoch || previous.AuthorizedRequestDigest != next.AuthorizedRequestDigest || previous.PredecessorNullifierCommitment != next.PredecessorNullifierCommitment || previous.RotationRequestDigest != next.RotationRequestDigest || previous.CreatedAtUnix != next.CreatedAtUnix || previous.CreatedAtCoordinate != next.CreatedAtCoordinate || previous.Retention != next.Retention || next.UpdatedAtUnix < previous.UpdatedAtUnix || next.UpdatedAtCoordinate < previous.UpdatedAtCoordinate {
		return errors.New("immutable issuer-link binding changed")
	}
	allowed := map[LinkStatus]map[LinkStatus]bool{
		StatusPending:  {StatusCooldown: true, StatusAppealed: true, StatusRevoked: true},
		StatusCooldown: {StatusActive: true, StatusAppealed: true, StatusRevoked: true},
		StatusActive:   {StatusSuperseded: true, StatusAppealed: true, StatusRevoked: true},
		StatusAppealed: {StatusCooldown: true, StatusRevoked: true},
	}
	if !allowed[previous.Status][next.Status] {
		return errors.New("issuer-link status transition is not allowed")
	}
	return nil
}

type TransitionKind string

const (
	TransitionBeginCooldown TransitionKind = "begin_cooldown"
	TransitionActivate      TransitionKind = "activate"
	TransitionAppeal        TransitionKind = "appeal"
	TransitionResolveAppeal TransitionKind = "resolve_appeal"
	TransitionRevoke        TransitionKind = "revoke"
)

type TransitionAction struct {
	Kind             TransitionKind
	AppealCommitment string
	Resolution       VerifiedAppealResolution
}

type AppealResolutionClaim struct {
	Version                 uint32          `json:"version"`
	Domain                  NullifierDomain `json:"domain"`
	Nullifier               string          `json:"nullifier"`
	RecordCommitment        string          `json:"record_commitment"`
	FreshCooldownCoordinate uint64          `json:"fresh_cooldown_coordinate"`
	CurrentCoordinate       uint64          `json:"current_coordinate"`
	ExpiresAtCoordinate     uint64          `json:"expires_at_coordinate"`
	AuthorityKeyID          string          `json:"authority_key_id"`
}

type SignedAppealResolution struct {
	Claim     AppealResolutionClaim `json:"claim"`
	Signature []byte                `json:"signature"`
}

type AppealAuthorizer interface {
	VerifyAppealResolution(context.Context, SignedAppealResolution) error
}

type VerifiedAppealResolution struct {
	claim  AppealResolutionClaim
	digest string
	scope  *authorizationScope
}

func verifyAppealResolution(ctx context.Context, signed SignedAppealResolution, authorizer AppealAuthorizer, coordinate uint64, scope *authorizationScope) (VerifiedAppealResolution, error) {
	if authorizer == nil || signed.Claim.CurrentCoordinate != coordinate || signed.Claim.ExpiresAtCoordinate <= coordinate || signed.Claim.FreshCooldownCoordinate <= coordinate {
		return VerifiedAppealResolution{}, ErrUnauthorized
	}
	value, err := signed.Claim.signingBytes()
	if err != nil {
		return VerifiedAppealResolution{}, err
	}
	if err := authorizer.VerifyAppealResolution(ctx, signed); err != nil {
		return VerifiedAppealResolution{}, err
	}
	return VerifiedAppealResolution{claim: signed.Claim, digest: digest(value), scope: scope}, nil
}

func (c AppealResolutionClaim) signingBytes() ([]byte, error) {
	if c.Version != Version1 || c.Domain.Validate() != nil || !validDigests(c.Nullifier, c.RecordCommitment) || !validOpaqueID(c.AuthorityKeyID) || c.FreshCooldownCoordinate <= c.CurrentCoordinate || c.ExpiresAtCoordinate <= c.CurrentCoordinate {
		return nil, errors.New("complete appeal resolution claim is required")
	}
	e := newCanonicalEncoder("virtengine.issuer-link.appeal-resolution/v1")
	e.u32(c.Version)
	c.Domain.encode(e)
	e.text(c.Nullifier)
	e.text(c.RecordCommitment)
	e.u64(c.FreshCooldownCoordinate)
	e.u64(c.CurrentCoordinate)
	e.u64(c.ExpiresAtCoordinate)
	e.text(c.AuthorityKeyID)
	return e.result(), nil
}

type LookupClaim struct {
	Version             uint32          `json:"version"`
	Domain              NullifierDomain `json:"domain"`
	Nullifier           string          `json:"nullifier"`
	PurposeDigest       string          `json:"purpose_digest"`
	RequesterDigest     string          `json:"requester_digest"`
	CurrentCoordinate   uint64          `json:"current_coordinate"`
	ExpiresAtCoordinate uint64          `json:"expires_at_coordinate"`
	AuthorityKeyID      string          `json:"authority_key_id"`
	RequireWallet       bool            `json:"require_wallet_commitment"`
}

type LookupCapability struct {
	Claim     LookupClaim `json:"claim"`
	Signature []byte      `json:"signature"`
}

func (c LookupClaim) signingBytes() ([]byte, error) {
	if c.Version != Version1 || c.Domain.Validate() != nil || !validDigests(c.Nullifier, c.PurposeDigest, c.RequesterDigest) || !validOpaqueID(c.AuthorityKeyID) || c.ExpiresAtCoordinate <= c.CurrentCoordinate {
		return nil, errors.New("complete lookup capability claim is required")
	}
	e := newCanonicalEncoder("virtengine.issuer-link.lookup/v1")
	e.u32(c.Version)
	c.Domain.encode(e)
	e.text(c.Nullifier)
	e.text(c.PurposeDigest)
	e.text(c.RequesterDigest)
	e.u64(c.CurrentCoordinate)
	e.u64(c.ExpiresAtCoordinate)
	e.text(c.AuthorityKeyID)
	e.boolean(c.RequireWallet)
	return e.result(), nil
}

type LookupAuthorizer interface {
	VerifyLookup(context.Context, LookupCapability) error
}

type DeletionTombstone struct {
	Version                     uint32 `json:"version"`
	RecordCommitment            string `json:"record_commitment"`
	DeletionAuthorizationDigest string `json:"deletion_authorization_digest"`
	DeletedAtUnix               int64  `json:"deleted_at_unix"`
	DeletedAtCoordinate         uint64 `json:"deleted_at_coordinate"`
}

type IssuerLinkProjection struct {
	Version             uint32             `json:"version"`
	Status              LinkStatus         `json:"status"`
	ScopedNullifier     string             `json:"scoped_nullifier,omitempty"`
	IssuerCommitment    string             `json:"issuer_commitment,omitempty"`
	ProgramDigest       string             `json:"program_digest,omitempty"`
	PolicyDigest        string             `json:"policy_digest,omitempty"`
	ProfileDigest       string             `json:"profile_digest,omitempty"`
	WalletKeyCommitment string             `json:"wallet_key_commitment,omitempty"`
	WalletKeyEpoch      uint64             `json:"wallet_key_epoch,omitempty"`
	Tombstone           *DeletionTombstone `json:"deletion_tombstone,omitempty"`
}

type ConfidentialIssuerLinkRegistry interface {
	Lookup(context.Context, NullifierDomain, string, LookupCapability) (IssuerLinkProjection, error)
}

type RetentionAction string

const (
	RetentionHoldAction     RetentionAction = "hold"
	RetentionReleaseAction  RetentionAction = "release"
	RetentionEligibleAction RetentionAction = "eligible"
	RetentionDeleteAction   RetentionAction = "delete"
)

type RetentionClaim struct {
	Version             uint32          `json:"version"`
	Domain              NullifierDomain `json:"domain"`
	Nullifier           string          `json:"nullifier"`
	RecordCommitment    string          `json:"record_commitment"`
	Action              RetentionAction `json:"action"`
	LegalHold           bool            `json:"legal_hold"`
	CurrentCoordinate   uint64          `json:"current_coordinate"`
	ExpiresAtCoordinate uint64          `json:"expires_at_coordinate"`
	AuthorityKeyID      string          `json:"authority_key_id"`
}

type SignedRetentionAuthorization struct {
	Claim     RetentionClaim `json:"claim"`
	Signature []byte         `json:"signature"`
}

func (c RetentionClaim) signingBytes() ([]byte, error) {
	if c.Version != Version1 || c.Domain.Validate() != nil || !validDigests(c.Nullifier, c.RecordCommitment) || !validOpaqueID(c.AuthorityKeyID) || c.ExpiresAtCoordinate <= c.CurrentCoordinate {
		return nil, errors.New("complete retention authorization claim is required")
	}
	if c.Action != RetentionHoldAction && c.Action != RetentionReleaseAction && c.Action != RetentionEligibleAction && c.Action != RetentionDeleteAction {
		return nil, errors.New("unknown retention action")
	}
	e := newCanonicalEncoder("virtengine.issuer-link.retention/v1")
	e.u32(c.Version)
	c.Domain.encode(e)
	e.text(c.Nullifier)
	e.text(c.RecordCommitment)
	e.text(string(c.Action))
	e.boolean(c.LegalHold)
	e.u64(c.CurrentCoordinate)
	e.u64(c.ExpiresAtCoordinate)
	e.text(c.AuthorityKeyID)
	return e.result(), nil
}

type RetentionAuthorizer interface {
	VerifyRetention(context.Context, SignedRetentionAuthorization) error
}

type VerifiedRetentionAuthorization struct {
	claim  RetentionClaim
	digest string
	scope  *authorizationScope
}

func verifyRetentionAuthorization(ctx context.Context, signed SignedRetentionAuthorization, authorizer RetentionAuthorizer, coordinate uint64, scope *authorizationScope) (VerifiedRetentionAuthorization, error) {
	if authorizer == nil || signed.Claim.CurrentCoordinate != coordinate || signed.Claim.ExpiresAtCoordinate <= coordinate {
		return VerifiedRetentionAuthorization{}, ErrUnauthorized
	}
	value, err := signed.Claim.signingBytes()
	if err != nil {
		return VerifiedRetentionAuthorization{}, err
	}
	if err := authorizer.VerifyRetention(ctx, signed); err != nil {
		return VerifiedRetentionAuthorization{}, err
	}
	e := newCanonicalEncoder("virtengine.issuer-link.retention-authorization/v1")
	e.bytes(value)
	e.bytes(signed.Signature)
	return VerifiedRetentionAuthorization{claim: signed.Claim, digest: digest(e.result()), scope: scope}, nil
}

type authorizationScope struct{}

type LinkRegistration struct {
	Request       WalletLinkRequest
	Authorization WalletAuthorizations
	Attestation   IssuerAttestation
	StableInput   StableInputCommitment
}

type IssuerLinkTransaction interface {
	InsertPending() (IssuerLinkRecord, error)
}

type AtomicIssuerLinkStore interface {
	ConfidentialIssuerLinkRegistry
	Register(context.Context, LinkRegistration, func(IssuerLinkTransaction) (IssuerLinkRecord, error)) (IssuerLinkRecord, error)
	AuthorizeAppealResolution(context.Context, SignedAppealResolution) (VerifiedAppealResolution, error)
	AuthorizeRetention(context.Context, SignedRetentionAuthorization) (VerifiedRetentionAuthorization, error)
	Transition(context.Context, NullifierDomain, string, TransitionAction) error
	ActivateRotation(context.Context, NullifierDomain, string, string, string) error
	ApplyRetention(context.Context, NullifierDomain, string, VerifiedRetentionAuthorization) error
	Authenticate(context.Context, NullifierDomain, string, ed25519.PublicKey, []byte, string) error
}

func AuthenticationSigningBytes(domain NullifierDomain, nullifier, challengeDigest string, coordinate uint64) ([]byte, error) {
	if err := domain.Validate(); err != nil {
		return nil, err
	}
	if !validDigests(nullifier, challengeDigest) {
		return nil, errors.New("strict authentication challenge binding is required")
	}
	e := newCanonicalEncoder("virtengine.issuer-link.authentication/v1")
	domain.encode(e)
	e.text(nullifier)
	e.text(challengeDigest)
	e.u64(coordinate)
	return e.result(), nil
}

func nullifierCommitment(nullifier string) string {
	e := newCanonicalEncoder("virtengine.issuer-link.nullifier-commitment/v1")
	e.text(nullifier)
	return digest(e.result())
}

type canonicalEncoder struct{ data []byte }

func newCanonicalEncoder(domain string) *canonicalEncoder {
	e := &canonicalEncoder{}
	e.text(domain)
	return e
}

func (e *canonicalEncoder) add(value []byte) {
	length := make([]byte, 8)
	binary.BigEndian.PutUint64(length, uint64(len(value)))
	e.data = append(e.data, length...)
	e.data = append(e.data, value...)
}

func (e *canonicalEncoder) text(value string)  { e.add([]byte(value)) }
func (e *canonicalEncoder) bytes(value []byte) { e.add(value) }
func (e *canonicalEncoder) boolean(value bool) {
	if value {
		e.add([]byte{1})
		return
	}
	e.add([]byte{0})
}
func (e *canonicalEncoder) u32(value uint32) {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, value)
	e.add(data)
}
func (e *canonicalEncoder) u64(value uint64) {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, value)
	e.add(data)
}
func (e *canonicalEncoder) result() []byte { return append([]byte(nil), e.data...) }

func digest(value []byte) string {
	sum := sha256.Sum256(value)
	return hex.EncodeToString(sum[:])
}

func validDigest(value string) bool { return digestPattern.MatchString(value) }
func validDigests(values ...string) bool {
	for _, value := range values {
		if !validDigest(value) {
			return false
		}
	}
	return true
}
func validOpaqueID(value string) bool { return opaqueIDPattern.MatchString(value) } // Package issuerlink defines fixture-only contracts for privacy-preserving issuer links.
