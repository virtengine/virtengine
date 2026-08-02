package contracts

import (
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

const (
	EvidenceObjectRefVersion      uint32 = 1
	CommitmentOpeningSize                = 32
	CommitmentDomainIdentityScope        = "identity-scope"
	CommitmentDomainSocialScope          = "social-scope"
	CommitmentDomainDocument             = "document"
	CommitmentDomainBiometric            = "biometric"
)

type RetentionState string

const (
	RetentionActive             RetentionState = "active"
	RetentionDeletionScheduled  RetentionState = "deletion_scheduled"
	RetentionLegalHold          RetentionState = "legal_hold"
	RetentionDeletionPending    RetentionState = "deletion_pending"
	RetentionDeletionUnresolved RetentionState = "deletion_unresolved"
	RetentionDeleted            RetentionState = "deleted"
)

// EvidenceObjectFields are the public, deterministic inputs to an evidence
// commitment. Digests and commitments are lowercase SHA-256 hex strings.
type EvidenceObjectFields struct {
	EvidenceDigest        string `json:"evidence_digest"`
	RetentionPolicyDigest string `json:"retention_policy_digest"`
	PolicyDigest          string `json:"policy_digest"`
	ProfileDigest         string `json:"profile_digest"`
	KeyEpoch              uint64 `json:"key_epoch"`
	CreatedHeight         int64  `json:"created_height"`
	CreatedUnix           int64  `json:"created_unix"`
	ExpiresHeight         int64  `json:"expires_height"`
	ExpiresUnix           int64  `json:"expires_unix"`
}

// EvidenceObjectRef is safe for consensus storage. The commitment opening and
// all storage, encryption, and subject data remain off-chain.
type EvidenceObjectRef struct {
	Version               uint32         `json:"version"`
	CommitmentDomain      string         `json:"commitment_domain"`
	ObjectCommitment      string         `json:"object_commitment"`
	EvidenceDigest        string         `json:"evidence_digest"`
	RetentionPolicyDigest string         `json:"retention_policy_digest"`
	PolicyDigest          string         `json:"policy_digest"`
	ProfileDigest         string         `json:"profile_digest"`
	KeyEpoch              uint64         `json:"key_epoch"`
	State                 RetentionState `json:"state"`
	CreatedHeight         int64          `json:"created_height"`
	CreatedUnix           int64          `json:"created_unix"`
	ExpiresHeight         int64          `json:"expires_height"`
	ExpiresUnix           int64          `json:"expires_unix"`
	UpdatedHeight         int64          `json:"updated_height"`
	UpdatedUnix           int64          `json:"updated_unix"`
}

func CreateEvidenceObjectRef(random io.Reader, domain string, fields EvidenceObjectFields) (EvidenceObjectRef, []byte, error) {
	if random == nil {
		return EvidenceObjectRef{}, nil, errors.New("commitment randomness is required")
	}
	if err := validateEvidenceFields(domain, fields); err != nil {
		return EvidenceObjectRef{}, nil, err
	}
	opening := make([]byte, CommitmentOpeningSize)
	if _, err := io.ReadFull(random, opening); err != nil {
		return EvidenceObjectRef{}, nil, fmt.Errorf("read commitment opening: %w", err)
	}
	if allZero(opening) {
		return EvidenceObjectRef{}, nil, errors.New("commitment opening must not be zero")
	}
	ref := EvidenceObjectRef{
		Version:               EvidenceObjectRefVersion,
		CommitmentDomain:      domain,
		EvidenceDigest:        fields.EvidenceDigest,
		RetentionPolicyDigest: fields.RetentionPolicyDigest,
		PolicyDigest:          fields.PolicyDigest,
		ProfileDigest:         fields.ProfileDigest,
		KeyEpoch:              fields.KeyEpoch,
		State:                 RetentionActive,
		CreatedHeight:         fields.CreatedHeight,
		CreatedUnix:           fields.CreatedUnix,
		ExpiresHeight:         fields.ExpiresHeight,
		ExpiresUnix:           fields.ExpiresUnix,
		UpdatedHeight:         fields.CreatedHeight,
		UpdatedUnix:           fields.CreatedUnix,
	}
	ref.ObjectCommitment = evidenceCommitment(domain, fields, opening)
	return ref, opening, nil
}

func VerifyEvidenceObjectRef(ref EvidenceObjectRef, opening []byte) error {
	if err := ref.Validate(); err != nil {
		return err
	}
	if len(opening) != CommitmentOpeningSize || allZero(opening) {
		return errors.New("commitment opening must be a non-zero 32-byte value")
	}
	expected := evidenceCommitment(ref.CommitmentDomain, ref.Fields(), opening)
	if subtle.ConstantTimeCompare([]byte(expected), []byte(ref.ObjectCommitment)) != 1 {
		return errors.New("evidence object commitment mismatch")
	}
	return nil
}

func (r EvidenceObjectRef) Fields() EvidenceObjectFields {
	return EvidenceObjectFields{
		EvidenceDigest: r.EvidenceDigest, RetentionPolicyDigest: r.RetentionPolicyDigest,
		PolicyDigest: r.PolicyDigest, ProfileDigest: r.ProfileDigest, KeyEpoch: r.KeyEpoch,
		CreatedHeight: r.CreatedHeight, CreatedUnix: r.CreatedUnix,
		ExpiresHeight: r.ExpiresHeight, ExpiresUnix: r.ExpiresUnix,
	}
}

func (r EvidenceObjectRef) Validate() error {
	if r.Version != EvidenceObjectRefVersion {
		return errors.New("unsupported evidence object reference version")
	}
	if err := validateEvidenceFields(r.CommitmentDomain, r.Fields()); err != nil {
		return err
	}
	if err := validateDigest(r.ObjectCommitment, "object commitment"); err != nil {
		return err
	}
	if !validRetentionState(r.State) {
		return errors.New("invalid retention state")
	}
	if r.UpdatedHeight < r.CreatedHeight || r.UpdatedUnix < r.CreatedUnix {
		return errors.New("evidence object update coordinates cannot predate creation")
	}
	return nil
}

func CanTransitionRetention(from, to RetentionState) bool {
	allowed := map[RetentionState]map[RetentionState]bool{
		RetentionActive:             {RetentionDeletionScheduled: true, RetentionLegalHold: true},
		RetentionDeletionScheduled:  {RetentionDeletionPending: true, RetentionLegalHold: true},
		RetentionLegalHold:          {RetentionActive: true, RetentionDeletionScheduled: true},
		RetentionDeletionPending:    {RetentionDeleted: true, RetentionDeletionUnresolved: true, RetentionLegalHold: true},
		RetentionDeletionUnresolved: {RetentionDeletionPending: true, RetentionLegalHold: true},
		RetentionDeleted:            {},
	}
	return allowed[from][to]
}

func TransitionRetention(ref EvidenceObjectRef, state RetentionState, height, unixTime int64) (EvidenceObjectRef, error) {
	if err := ref.Validate(); err != nil {
		return ref, err
	}
	if !CanTransitionRetention(ref.State, state) {
		return ref, fmt.Errorf("invalid retention transition %s -> %s", ref.State, state)
	}
	if height < ref.UpdatedHeight || unixTime < ref.UpdatedUnix {
		return ref, errors.New("retention transition time must be monotonic")
	}
	ref.State, ref.UpdatedHeight, ref.UpdatedUnix = state, height, unixTime
	return ref, nil
}

type DeletionReceiptKind string

const (
	ReceiptStorageDeletion DeletionReceiptKind = "storage_deletion"
	ReceiptKMSDestruction  DeletionReceiptKind = "kms_destruction"
)

type DeletionReceipt struct {
	Version                uint32              `json:"version"`
	Kind                   DeletionReceiptKind `json:"kind"`
	TargetCommitment       string              `json:"target_commitment"`
	RequestDigest          string              `json:"request_digest,omitempty"`
	AuthorizationDigest    string              `json:"authorization_digest"`
	PolicyDigest           string              `json:"policy_digest"`
	ProfileDigest          string              `json:"profile_digest"`
	ConsentDecisionDigest  string              `json:"consent_decision_digest,omitempty"`
	ConsentPolicyDigest    string              `json:"consent_policy_digest,omitempty"`
	OperationID            string              `json:"operation_id"`
	KeyEpoch               uint64              `json:"key_epoch"`
	ObjectCommitment       string              `json:"object_commitment,omitempty"`
	StorageCommitment      string              `json:"storage_commitment,omitempty"`
	KeyCommitment          string              `json:"key_commitment,omitempty"`
	BackupGenerationDigest string              `json:"backup_generation_digest,omitempty"`
	BackupExpiryHeight     int64               `json:"backup_expiry_height,omitempty"`
	BackupExpiryUnix       int64               `json:"backup_expiry_unix,omitempty"`
	ErasureEpoch           uint64              `json:"erasure_epoch,omitempty"`
	CompletedHeight        int64               `json:"completed_height"`
	CompletedUnix          int64               `json:"completed_unix"`
	SignerKeyID            string              `json:"signer_key_id"`
	SignerKeyEpoch         uint64              `json:"signer_key_epoch"`
	Signature              []byte              `json:"signature"`
}

type DeletionReceiptKeyResolver interface {
	ResolveDeletionReceiptKey(kind DeletionReceiptKind, keyID string, keyEpoch uint64) (ed25519.PublicKey, error)
}

// DeletionReceiptAuthorityResolver binds signer keys to independently
// administered authority domains for the v2 erasure lifecycle.
type DeletionReceiptAuthorityResolver interface {
	DeletionReceiptKeyResolver
	ResolveDeletionReceiptAuthority(kind DeletionReceiptKind, keyID string, keyEpoch uint64) (string, error)
}

type DeletionReceiptReplay struct {
	OperationID   string
	ReceiptDigest string
}

type DeletionReceiptReplayConsumer interface {
	ConsumeDeletionReceipts(storage, kms DeletionReceiptReplay, apply func() error) error
}

type DeletionResolutionContext struct {
	AuthorizationDigest    string
	PolicyDigest           string
	ProfileDigest          string
	RequestDigest          string
	ConsentDecisionDigest  string
	ConsentPolicyDigest    string
	ObjectCommitment       string
	StorageCommitment      string
	KeyCommitment          string
	BackupGenerationDigest string
	BackupExpiryHeight     int64
	BackupExpiryUnix       int64
	ErasureEpoch           uint64
	CurrentHeight          int64
	CurrentUnix            int64
	LegalHold              bool
	ReplayConsumer         DeletionReceiptReplayConsumer
	ApplyResolved          func(EvidenceObjectRef) error
}

func (r DeletionReceipt) CanonicalSignBytes() ([]byte, error) {
	if err := r.validate(false); err != nil {
		return nil, err
	}
	if r.ErasureEpoch == 0 {
		return canonicalValues("virtengine/evidence-deletion-receipt/v1", fmt.Sprint(r.Version), string(r.Kind),
			r.TargetCommitment, r.AuthorizationDigest, r.PolicyDigest, r.ProfileDigest, r.OperationID,
			fmt.Sprint(r.KeyEpoch), fmt.Sprint(r.CompletedHeight), fmt.Sprint(r.CompletedUnix),
			r.SignerKeyID, fmt.Sprint(r.SignerKeyEpoch)), nil
	}
	return canonicalValues("virtengine/evidence-deletion-receipt/v2", fmt.Sprint(r.Version), string(r.Kind),
		r.TargetCommitment, r.RequestDigest, r.AuthorizationDigest, r.PolicyDigest, r.ProfileDigest,
		r.ConsentDecisionDigest, r.ConsentPolicyDigest, r.OperationID, fmt.Sprint(r.KeyEpoch),
		r.ObjectCommitment, r.StorageCommitment, r.KeyCommitment, r.BackupGenerationDigest,
		fmt.Sprint(r.BackupExpiryHeight), fmt.Sprint(r.BackupExpiryUnix), fmt.Sprint(r.ErasureEpoch),
		fmt.Sprint(r.CompletedHeight), fmt.Sprint(r.CompletedUnix),
		r.SignerKeyID, fmt.Sprint(r.SignerKeyEpoch)), nil
}

func VerifyDeletionReceipt(receipt DeletionReceipt, resolver DeletionReceiptKeyResolver) error {
	_, err := verifyDeletionReceipt(receipt, resolver)
	return err
}

func verifyDeletionReceipt(receipt DeletionReceipt, resolver DeletionReceiptKeyResolver) (ed25519.PublicKey, error) {
	if resolver == nil {
		return nil, errors.New("deletion receipt key resolver is required")
	}
	if err := receipt.validate(true); err != nil {
		return nil, err
	}
	publicKey, err := resolver.ResolveDeletionReceiptKey(receipt.Kind, receipt.SignerKeyID, receipt.SignerKeyEpoch)
	if err != nil {
		return nil, fmt.Errorf("resolve deletion receipt key: %w", err)
	}
	if len(publicKey) != ed25519.PublicKeySize {
		return nil, errors.New("invalid deletion receipt public key")
	}
	signBytes, _ := receipt.CanonicalSignBytes()
	if !ed25519.Verify(publicKey, signBytes, receipt.Signature) {
		return nil, errors.New("invalid deletion receipt signature")
	}
	return append(ed25519.PublicKey(nil), publicKey...), nil
}

func ResolveDeletion(ref EvidenceObjectRef, resolution DeletionResolutionContext, receipts []DeletionReceipt, resolver DeletionReceiptKeyResolver) (EvidenceObjectRef, error) {
	if err := ref.Validate(); err != nil {
		return ref, err
	}
	if resolution.LegalHold || ref.State == RetentionLegalHold {
		return ref, errors.New("deletion blocked by legal hold")
	}
	if ref.State == RetentionDeleted {
		return ref, errors.New("evidence object is already deleted")
	}
	unresolved := ref
	if ref.State == RetentionDeletionPending {
		unresolved.State = RetentionDeletionUnresolved
	}
	fail := func(err error) (EvidenceObjectRef, error) { return unresolved, err }
	if ref.State != RetentionDeletionPending {
		return fail(errors.New("evidence object is not pending deletion"))
	}
	if len(receipts) != 2 {
		return fail(errors.New("storage and KMS deletion receipts are required"))
	}
	if resolution.ReplayConsumer == nil {
		return fail(errors.New("deletion receipt replay consumer is required"))
	}
	if resolution.ApplyResolved == nil {
		return fail(errors.New("resolved deletion persistence callback is required"))
	}
	for name, digest := range map[string]string{
		"authorization digest": resolution.AuthorizationDigest,
		"policy digest":        resolution.PolicyDigest,
		"profile digest":       resolution.ProfileDigest,
	} {
		if err := validateDigest(digest, name); err != nil {
			return fail(err)
		}
	}
	if resolution.ErasureEpoch != 0 {
		if resolution.ObjectCommitment != ref.ObjectCommitment || resolution.PolicyDigest != ref.PolicyDigest ||
			resolution.ProfileDigest != ref.ProfileDigest {
			return fail(errors.New("erasure resolution claims do not match the evidence object"))
		}
		for name, digest := range map[string]string{
			"request digest": resolution.RequestDigest, "consent decision digest": resolution.ConsentDecisionDigest,
			"consent policy digest": resolution.ConsentPolicyDigest, "object commitment": resolution.ObjectCommitment,
			"storage commitment": resolution.StorageCommitment, "key commitment": resolution.KeyCommitment,
			"backup generation digest": resolution.BackupGenerationDigest,
		} {
			if err := validateDigest(digest, name); err != nil {
				return fail(err)
			}
		}
		if resolution.BackupExpiryHeight <= 0 || resolution.BackupExpiryUnix <= 0 {
			return fail(errors.New("backup expiry coordinates are required"))
		}
	}
	if resolution.CurrentHeight <= 0 || resolution.CurrentUnix <= 0 {
		return fail(errors.New("current deletion resolution coordinates are required"))
	}
	seen := make(map[DeletionReceiptKind]bool, 2)
	operations := make(map[string]bool, 2)
	keys := make(map[DeletionReceiptKind]ed25519.PublicKey, 2)
	authorities := make(map[DeletionReceiptKind]string, 2)
	replays := make(map[DeletionReceiptKind]DeletionReceiptReplay, 2)
	var completedHeight, completedUnix int64
	for _, receipt := range receipts {
		if seen[receipt.Kind] || operations[receipt.OperationID] {
			return fail(errors.New("duplicate deletion receipt kind or operation"))
		}
		publicKey, err := verifyDeletionReceipt(receipt, resolver)
		if err != nil {
			return fail(err)
		}
		if receipt.TargetCommitment != ref.ObjectCommitment || receipt.AuthorizationDigest != resolution.AuthorizationDigest ||
			receipt.PolicyDigest != resolution.PolicyDigest || receipt.ProfileDigest != resolution.ProfileDigest || receipt.KeyEpoch != ref.KeyEpoch {
			return fail(errors.New("deletion receipt claims do not match the evidence object"))
		}
		if resolution.ErasureEpoch != 0 && (receipt.RequestDigest != resolution.RequestDigest ||
			receipt.ConsentDecisionDigest != resolution.ConsentDecisionDigest || receipt.ConsentPolicyDigest != resolution.ConsentPolicyDigest ||
			receipt.ObjectCommitment != resolution.ObjectCommitment || receipt.StorageCommitment != resolution.StorageCommitment ||
			receipt.KeyCommitment != resolution.KeyCommitment || receipt.BackupGenerationDigest != resolution.BackupGenerationDigest ||
			receipt.BackupExpiryHeight != resolution.BackupExpiryHeight || receipt.BackupExpiryUnix != resolution.BackupExpiryUnix ||
			receipt.ErasureEpoch != resolution.ErasureEpoch) {
			return fail(errors.New("deletion receipt erasure fence claims do not match the resolution"))
		}
		if receipt.CompletedHeight < ref.UpdatedHeight || receipt.CompletedUnix < ref.UpdatedUnix {
			return fail(errors.New("deletion receipt predates the pending deletion state"))
		}
		if receipt.CompletedHeight > resolution.CurrentHeight || receipt.CompletedUnix > resolution.CurrentUnix {
			return fail(errors.New("deletion receipt completion is in the future"))
		}
		seen[receipt.Kind] = true
		operations[receipt.OperationID] = true
		keys[receipt.Kind] = publicKey
		if resolution.ErasureEpoch != 0 {
			authorityResolver, ok := resolver.(DeletionReceiptAuthorityResolver)
			if !ok {
				return fail(errors.New("v2 deletion receipts require authority-domain resolution"))
			}
			authority, err := authorityResolver.ResolveDeletionReceiptAuthority(receipt.Kind, receipt.SignerKeyID, receipt.SignerKeyEpoch)
			if err != nil || authority == "" {
				return fail(errors.New("deletion receipt authority domain is unavailable"))
			}
			authorities[receipt.Kind] = authority
		}
		signBytes, _ := receipt.CanonicalSignBytes()
		receiptHash := sha256.New()
		receiptHash.Write(signBytes)
		receiptHash.Write(receipt.Signature)
		replays[receipt.Kind] = DeletionReceiptReplay{OperationID: receipt.OperationID, ReceiptDigest: hex.EncodeToString(receiptHash.Sum(nil))}
		if receipt.CompletedHeight > completedHeight {
			completedHeight = receipt.CompletedHeight
		}
		if receipt.CompletedUnix > completedUnix {
			completedUnix = receipt.CompletedUnix
		}
	}
	if !seen[ReceiptStorageDeletion] || !seen[ReceiptKMSDestruction] {
		return fail(errors.New("independent storage and KMS deletion receipts are required"))
	}
	if subtle.ConstantTimeCompare(keys[ReceiptStorageDeletion], keys[ReceiptKMSDestruction]) == 1 {
		return fail(errors.New("storage and KMS deletion authorities must use distinct public keys"))
	}
	if resolution.ErasureEpoch != 0 && authorities[ReceiptStorageDeletion] == authorities[ReceiptKMSDestruction] {
		return fail(errors.New("storage and KMS deletion authorities must use distinct authority domains"))
	}
	resolved := ref
	err := resolution.ReplayConsumer.ConsumeDeletionReceipts(replays[ReceiptStorageDeletion], replays[ReceiptKMSDestruction], func() error {
		resolved.State, resolved.UpdatedHeight, resolved.UpdatedUnix = RetentionDeleted, completedHeight, completedUnix
		return resolution.ApplyResolved(resolved)
	})
	if err != nil {
		return fail(fmt.Errorf("consume deletion receipts: %w", err))
	}
	return resolved, nil
}

func (r DeletionReceipt) validate(signature bool) error {
	if r.Version != EvidenceObjectRefVersion || (r.Kind != ReceiptStorageDeletion && r.Kind != ReceiptKMSDestruction) {
		return errors.New("invalid deletion receipt version or kind")
	}
	for name, digest := range map[string]string{"target commitment": r.TargetCommitment, "authorization digest": r.AuthorizationDigest, "policy digest": r.PolicyDigest, "profile digest": r.ProfileDigest} {
		if err := validateDigest(digest, name); err != nil {
			return err
		}
	}
	if r.ErasureEpoch != 0 {
		for name, digest := range map[string]string{
			"request digest": r.RequestDigest, "consent decision digest": r.ConsentDecisionDigest,
			"consent policy digest": r.ConsentPolicyDigest, "object commitment": r.ObjectCommitment,
			"storage commitment": r.StorageCommitment, "key commitment": r.KeyCommitment,
			"backup generation digest": r.BackupGenerationDigest,
		} {
			if err := validateDigest(digest, name); err != nil {
				return err
			}
		}
		if r.BackupExpiryHeight <= 0 || r.BackupExpiryUnix <= 0 {
			return errors.New("backup expiry coordinates are required")
		}
	}
	if r.OperationID == "" || r.SignerKeyID == "" || r.KeyEpoch == 0 || r.SignerKeyEpoch == 0 || r.CompletedHeight <= 0 || r.CompletedUnix <= 0 {
		return errors.New("incomplete deletion receipt")
	}
	if signature && len(r.Signature) != ed25519.SignatureSize {
		return errors.New("invalid deletion receipt signature size")
	}
	return nil
}

func evidenceCommitment(domain string, fields EvidenceObjectFields, opening []byte) string {
	values := canonicalValues("virtengine/evidence-object-ref/v1", domain, fields.EvidenceDigest,
		fields.RetentionPolicyDigest, fields.PolicyDigest, fields.ProfileDigest, fmt.Sprint(fields.KeyEpoch),
		fmt.Sprint(fields.CreatedHeight), fmt.Sprint(fields.CreatedUnix), fmt.Sprint(fields.ExpiresHeight), fmt.Sprint(fields.ExpiresUnix))
	hash := sha256.New()
	hash.Write(values)
	hash.Write(opening)
	return hex.EncodeToString(hash.Sum(nil))
}

func canonicalValues(values ...string) []byte {
	result := make([]byte, 0)
	var size [8]byte
	for _, value := range values {
		binary.BigEndian.PutUint64(size[:], uint64(len(value)))
		result = append(result, size[:]...)
		result = append(result, value...)
	}
	return result
}

func validateEvidenceFields(domain string, fields EvidenceObjectFields) error {
	switch domain {
	case CommitmentDomainIdentityScope, CommitmentDomainSocialScope, CommitmentDomainDocument, CommitmentDomainBiometric:
	default:
		return errors.New("unsupported commitment domain")
	}
	for name, digest := range map[string]string{"evidence digest": fields.EvidenceDigest, "retention policy digest": fields.RetentionPolicyDigest, "policy digest": fields.PolicyDigest, "profile digest": fields.ProfileDigest} {
		if err := validateDigest(digest, name); err != nil {
			return err
		}
	}
	if fields.KeyEpoch == 0 || fields.CreatedHeight <= 0 || fields.CreatedUnix <= 0 {
		return errors.New("key epoch and creation coordinates are required")
	}
	if fields.ExpiresHeight < 0 || fields.ExpiresUnix < 0 {
		return errors.New("expiration coordinates cannot be negative")
	}
	return nil
}

func validateDigest(value, name string) error {
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != sha256.Size || hex.EncodeToString(decoded) != value {
		return fmt.Errorf("%s must be lowercase SHA-256 hex", name)
	}
	return nil
}

func validRetentionState(state RetentionState) bool {
	switch state {
	case RetentionActive, RetentionDeletionScheduled, RetentionLegalHold, RetentionDeletionPending, RetentionDeletionUnresolved, RetentionDeleted:
		return true
	default:
		return false
	}
}

func allZero(value []byte) bool {
	var combined byte
	for _, item := range value {
		combined |= item
	}
	return combined == 0
}
