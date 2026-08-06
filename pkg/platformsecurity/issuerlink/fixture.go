package issuerlink

import (
	"context"
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"slices"
	"sync"
)

const fixtureWarning = "fixture_only: deterministic adapters are production_unavailable and have no threshold custody or external security review"

type DeterministicThresholdPRFFixture struct{ key []byte }

func NewDeterministicThresholdPRFFixture(key []byte) (*DeterministicThresholdPRFFixture, error) {
	if len(key) < 16 {
		return nil, errors.New("fixture PRF key must contain at least 16 bytes")
	}
	return &DeterministicThresholdPRFFixture{key: slices.Clone(key)}, nil
}

func (f *DeterministicThresholdPRFFixture) FixtureState() string  { return FixtureOnlyState }
func (f *DeterministicThresholdPRFFixture) ProductionReady() bool { return false }
func (f *DeterministicThresholdPRFFixture) ReviewStatus() string  { return fixtureWarning }

func (f *DeterministicThresholdPRFFixture) Evaluate(_ context.Context, authorization VerifiedNullifierAuthorization, input NullifierInput) (string, error) {
	if err := authorization.validate(input); err != nil {
		return "", err
	}
	e := newCanonicalEncoder("virtengine.issuer-link.threshold-prf/v1")
	input.Domain.encode(e)
	e.text(input.IssuerCommitment)
	e.text(input.PolicyDigest)
	e.text(input.ProfileDigest)
	e.u64(input.IssuerKeyEpoch)
	e.text(input.RequestDigest)
	e.bytes(input.stableCommitment)
	mac := hmac.New(sha256.New, f.key)
	_, _ = mac.Write(e.result())
	return fmt.Sprintf("%x", mac.Sum(nil)), nil
}

type IssuerKeyEntry struct {
	Profile   IssuerProfile
	PublicKey ed25519.PublicKey
}

type GovernedIssuerKeyResolverFixture struct {
	Entries map[string]IssuerKeyEntry
}

func IssuerKeyCoordinate(issuerCommitment, keyID string, epoch uint64) string {
	return fmt.Sprintf("%s:%s:%d", issuerCommitment, keyID, epoch)
}

func (f GovernedIssuerKeyResolverFixture) ResolveIssuerKey(_ context.Context, issuerCommitment, keyID string, epoch uint64) (IssuerProfile, ed25519.PublicKey, error) {
	entry, ok := f.Entries[IssuerKeyCoordinate(issuerCommitment, keyID, epoch)]
	if !ok {
		return IssuerProfile{}, nil, errors.New("unknown or stale governed issuer key")
	}
	if err := entry.Profile.Validate(); err != nil {
		return IssuerProfile{}, nil, err
	}
	if entry.Profile.IssuerCommitment != issuerCommitment || entry.Profile.SigningKeyID != keyID || entry.Profile.KeyEpoch != epoch || len(entry.PublicKey) != ed25519.PublicKeySize {
		return IssuerProfile{}, nil, errors.New("governed issuer key entry is inconsistent")
	}
	return entry.Profile, slices.Clone(entry.PublicKey), nil
}

func SignIssuerAttestationFixture(attestation IssuerAttestation, privateKey ed25519.PrivateKey) (IssuerAttestation, error) {
	message, err := attestation.signingBytes()
	if err != nil || len(privateKey) != ed25519.PrivateKeySize {
		return IssuerAttestation{}, errors.New("valid Ed25519 issuer attestation input is required")
	}
	attestation.Signature = ed25519.Sign(privateKey, message)
	return attestation, nil
}

type Ed25519AuthorizerFixture struct {
	Keys                  map[string]ed25519.PublicKey
	AllowedPurposeDigests map[string]bool
}

func (f Ed25519AuthorizerFixture) verify(keyID string, message, signature []byte) error {
	key, ok := f.Keys[keyID]
	if !ok || len(key) != ed25519.PublicKeySize || len(signature) != ed25519.SignatureSize || !ed25519.Verify(key, message, signature) {
		return ErrUnauthorized
	}
	return nil
}

func (f Ed25519AuthorizerFixture) VerifyLookup(_ context.Context, capability LookupCapability) error {
	message, err := capability.Claim.signingBytes()
	if err != nil {
		return err
	}
	if len(f.AllowedPurposeDigests) != 0 && !f.AllowedPurposeDigests[capability.Claim.PurposeDigest] {
		return ErrUnauthorized
	}
	return f.verify(capability.Claim.AuthorityKeyID, message, capability.Signature)
}

func (f Ed25519AuthorizerFixture) VerifyAppealResolution(_ context.Context, signed SignedAppealResolution) error {
	message, err := signed.Claim.signingBytes()
	if err != nil {
		return err
	}
	return f.verify(signed.Claim.AuthorityKeyID, message, signed.Signature)
}

func (f Ed25519AuthorizerFixture) VerifyRetention(_ context.Context, signed SignedRetentionAuthorization) error {
	message, err := signed.Claim.signingBytes()
	if err != nil {
		return err
	}
	return f.verify(signed.Claim.AuthorityKeyID, message, signed.Signature)
}

func SignLookupCapabilityFixture(claim LookupClaim, key ed25519.PrivateKey) (LookupCapability, error) {
	message, err := claim.signingBytes()
	if err != nil || len(key) != ed25519.PrivateKeySize {
		return LookupCapability{}, errors.New("valid lookup signing input is required")
	}
	return LookupCapability{Claim: claim, Signature: ed25519.Sign(key, message)}, nil
}

func SignAppealResolutionFixture(claim AppealResolutionClaim, key ed25519.PrivateKey) (SignedAppealResolution, error) {
	message, err := claim.signingBytes()
	if err != nil || len(key) != ed25519.PrivateKeySize {
		return SignedAppealResolution{}, errors.New("valid appeal signing input is required")
	}
	return SignedAppealResolution{Claim: claim, Signature: ed25519.Sign(key, message)}, nil
}

func SignRetentionAuthorizationFixture(claim RetentionClaim, key ed25519.PrivateKey) (SignedRetentionAuthorization, error) {
	message, err := claim.signingBytes()
	if err != nil || len(key) != ed25519.PrivateKeySize {
		return SignedRetentionAuthorization{}, errors.New("valid retention signing input is required")
	}
	return SignedRetentionAuthorization{Claim: claim, Signature: ed25519.Sign(key, message)}, nil
}

type idempotencyEntry struct {
	requestDigest      string
	registrationDigest string
	domainKey          string
	recordKey          string
}

type nonceEntry struct {
	requestDigest      string
	registrationDigest string
	domainKey          string
	recordKey          string
}

type authenticationReplayKey struct {
	domainKey       string
	nullifier       string
	challengeDigest string
	coordinate      uint64
}

type FixtureOption func(*MemoryAtomicIssuerLinkFixture)

func WithThresholdPRF(prf ThresholdPRF) FixtureOption {
	return func(store *MemoryAtomicIssuerLinkFixture) { store.prf = prf }
}

func WithLookupAuthorizer(authorizer LookupAuthorizer) FixtureOption {
	return func(store *MemoryAtomicIssuerLinkFixture) { store.lookupAuthorizer = authorizer }
}

func WithAppealAuthorizer(authorizer AppealAuthorizer) FixtureOption {
	return func(store *MemoryAtomicIssuerLinkFixture) { store.appealAuthorizer = authorizer }
}

func WithRetentionAuthorizer(authorizer RetentionAuthorizer) FixtureOption {
	return func(store *MemoryAtomicIssuerLinkFixture) { store.retentionAuthorizer = authorizer }
}

func WithAttestationPolicy(policy AttestationPolicy) FixtureOption {
	return func(store *MemoryAtomicIssuerLinkFixture) { store.attestationPolicy = policy }
}

type MemoryAtomicIssuerLinkFixture struct {
	mu                  sync.Mutex
	resolver            IssuerKeyResolver
	prf                 ThresholdPRF
	lookupAuthorizer    LookupAuthorizer
	appealAuthorizer    AppealAuthorizer
	retentionAuthorizer RetentionAuthorizer
	authorizationScope  *authorizationScope
	attestationPolicy   AttestationPolicy
	coordinate          uint64
	records             map[string]IssuerLinkRecord
	tombstones          map[string]DeletionTombstone
	idempotency         map[string]idempotencyEntry
	nonces              map[string]nonceEntry
	authentications     map[authenticationReplayKey]struct{}
	wallets             map[string]string
	failAfterInsertion  bool
}

func NewMemoryAtomicIssuerLinkFixture(resolver IssuerKeyResolver, options ...FixtureOption) *MemoryAtomicIssuerLinkFixture {
	prf, _ := NewDeterministicThresholdPRFFixture([]byte("issuer-link-explicit-nonproduction-fixture-key"))
	store := &MemoryAtomicIssuerLinkFixture{
		resolver: resolver, prf: prf,
		authorizationScope: &authorizationScope{},
		attestationPolicy:  AttestationPolicy{MaxLifetimeCoordinates: 100, MaxFutureSkew: 0},
		records:            make(map[string]IssuerLinkRecord), tombstones: make(map[string]DeletionTombstone),
		idempotency: make(map[string]idempotencyEntry), nonces: make(map[string]nonceEntry),
		authentications: make(map[authenticationReplayKey]struct{}), wallets: make(map[string]string),
	}
	for _, option := range options {
		option(store)
	}
	return store
}

func (f *MemoryAtomicIssuerLinkFixture) AuthorizeAppealResolution(ctx context.Context, signed SignedAppealResolution) (VerifiedAppealResolution, error) {
	if err := ctx.Err(); err != nil {
		return VerifiedAppealResolution{}, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	return verifyAppealResolution(ctx, signed, f.appealAuthorizer, f.coordinate, f.authorizationScope)
}

func (f *MemoryAtomicIssuerLinkFixture) AuthorizeRetention(ctx context.Context, signed SignedRetentionAuthorization) (VerifiedRetentionAuthorization, error) {
	if err := ctx.Err(); err != nil {
		return VerifiedRetentionAuthorization{}, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	return verifyRetentionAuthorization(ctx, signed, f.retentionAuthorizer, f.coordinate, f.authorizationScope)
}

func (f *MemoryAtomicIssuerLinkFixture) FixtureState() string  { return FixtureOnlyState }
func (f *MemoryAtomicIssuerLinkFixture) ProductionReady() bool { return false }
func (f *MemoryAtomicIssuerLinkFixture) ReviewStatus() string  { return fixtureWarning }

func (f *MemoryAtomicIssuerLinkFixture) SetCurrentCoordinate(coordinate uint64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.coordinate = coordinate
}

func (f *MemoryAtomicIssuerLinkFixture) InjectFailureAfterInsertion() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.failAfterInsertion = true
}

func (f *MemoryAtomicIssuerLinkFixture) Register(ctx context.Context, registration LinkRegistration, callback func(IssuerLinkTransaction) (IssuerLinkRecord, error)) (IssuerLinkRecord, error) {
	if callback == nil || f.prf == nil {
		return IssuerLinkRecord{}, errors.New("configured PRF and transaction callback are required")
	}
	request := registration.Request
	if err := request.ValidateStructure(); err != nil {
		return IssuerLinkRecord{}, err
	}
	if err := VerifyWalletAuthorizations(request, registration.Authorization); err != nil {
		return IssuerLinkRecord{}, err
	}
	registrationDigest, err := registration.bindingDigest()
	if err != nil {
		return IssuerLinkRecord{}, err
	}
	if err := ctx.Err(); err != nil {
		return IssuerLinkRecord{}, err
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	domainKey, _ := request.Domain.Key()
	if existing, ok := f.idempotency[request.IdempotencyKey]; ok {
		if existing.requestDigest != request.RequestDigest || existing.registrationDigest != registrationDigest || existing.domainKey != domainKey {
			return IssuerLinkRecord{}, ErrConflict
		}
		return f.records[existing.recordKey], nil
	}
	if used, ok := f.nonces[request.ChallengeNonceDigest]; ok {
		if used.requestDigest != request.RequestDigest || used.domainKey != domainKey {
			return IssuerLinkRecord{}, ErrCrossDomainReplay
		}
		if used.registrationDigest != registrationDigest {
			return IssuerLinkRecord{}, ErrConflict
		}
		return f.records[used.recordKey], nil
	}
	if err := request.ValidateAt(f.coordinate); err != nil {
		return IssuerLinkRecord{}, err
	}
	authorization, input, err := VerifyIssuerAttestation(ctx, request, registration.Attestation, f.resolver, f.coordinate, f.attestationPolicy, registration.StableInput)
	if err != nil {
		return IssuerLinkRecord{}, err
	}
	nullifier, err := f.prf.Evaluate(ctx, authorization, input)
	if err != nil || !validDigest(nullifier) {
		return IssuerLinkRecord{}, errors.New("configured PRF did not return a strict scoped nullifier")
	}
	if err := ctx.Err(); err != nil {
		return IssuerLinkRecord{}, err
	}

	records := cloneRecordMap(f.records)
	wallets := cloneStringMap(f.wallets)
	tx := &memoryIssuerLinkTransaction{registration: registration, domainKey: domainKey, nullifier: nullifier, coordinate: f.coordinate, records: records, wallets: wallets, failAfterInsertion: f.failAfterInsertion}
	f.failAfterInsertion = false
	record, err := callback(tx)
	if err != nil {
		return IssuerLinkRecord{}, err
	}
	if tx.poisoned != nil {
		return IssuerLinkRecord{}, fmt.Errorf("transaction was poisoned: %w", tx.poisoned)
	}
	if err := ctx.Err(); err != nil {
		return IssuerLinkRecord{}, err
	}
	if !tx.inserted {
		return IssuerLinkRecord{}, errors.New("transaction did not insert a pending issuer link")
	}
	recordKey := scopedRecordKey(domainKey, nullifier)
	stored, ok := records[recordKey]
	if !ok || stored != record {
		return IssuerLinkRecord{}, errors.New("callback returned a record other than the inserted record")
	}
	idempotency := cloneIdempotencyMap(f.idempotency)
	nonces := cloneNonceMap(f.nonces)
	idempotency[request.IdempotencyKey] = idempotencyEntry{requestDigest: request.RequestDigest, registrationDigest: registrationDigest, domainKey: domainKey, recordKey: recordKey}
	nonces[request.ChallengeNonceDigest] = nonceEntry{requestDigest: request.RequestDigest, registrationDigest: registrationDigest, domainKey: domainKey, recordKey: recordKey}
	f.records, f.wallets, f.idempotency, f.nonces = records, wallets, idempotency, nonces
	return record, nil
}

func (registration LinkRegistration) bindingDigest() (string, error) {
	requestBytes, err := registration.Request.SigningBytes()
	if err != nil {
		return "", err
	}
	attestationBytes, err := registration.Attestation.signingBytes()
	if err != nil || len(registration.StableInput.value) < 16 {
		return "", errors.New("complete registration binding is required")
	}
	e := newCanonicalEncoder("virtengine.issuer-link.registration/v1")
	e.bytes(requestBytes)
	e.bytes(registration.Authorization.OldWalletPublicKey)
	e.bytes(registration.Authorization.OldWalletSignature)
	e.bytes(registration.Authorization.NewWalletPublicKey)
	e.bytes(registration.Authorization.NewWalletSignature)
	e.bytes(attestationBytes)
	e.bytes(registration.Attestation.Signature)
	e.bytes(registration.StableInput.value)
	return digest(e.result()), nil
}

type memoryIssuerLinkTransaction struct {
	registration       LinkRegistration
	domainKey          string
	nullifier          string
	coordinate         uint64
	records            map[string]IssuerLinkRecord
	wallets            map[string]string
	inserted           bool
	poisoned           error
	failAfterInsertion bool
}

func (tx *memoryIssuerLinkTransaction) fail(err error) error {
	if tx.poisoned == nil {
		tx.poisoned = err
	}
	return err
}

func (tx *memoryIssuerLinkTransaction) InsertPending() (IssuerLinkRecord, error) {
	if tx.poisoned != nil {
		return IssuerLinkRecord{}, tx.poisoned
	}
	if tx.inserted {
		return IssuerLinkRecord{}, tx.fail(errors.New("transaction may insert exactly one record"))
	}
	request := tx.registration.Request
	record := IssuerLinkRecord{
		Version: Version1, ScopedNullifier: tx.nullifier, Domain: request.Domain,
		IssuerCommitment: request.IssuerCommitment, PolicyDigest: request.PolicyDigest, ProfileDigest: request.ProfileDigest,
		WalletKeyCommitment: request.NewWalletKeyCommitment, WalletKeyEpoch: request.NewWalletKeyEpoch,
		AuthorizedRequestDigest: request.RequestDigest, Status: StatusPending,
		CreatedAtUnix: int64(tx.coordinate), UpdatedAtUnix: int64(tx.coordinate),
		CreatedAtCoordinate: tx.coordinate, UpdatedAtCoordinate: tx.coordinate,
		CooldownUntilCoordinate: request.CooldownUntilCoordinate, Retention: RetentionRequired,
	}
	if request.IsRotation() {
		oldWalletKey := scopedWalletKey(tx.domainKey, request.OldWalletKeyCommitment)
		oldRecordKey, ok := tx.wallets[oldWalletKey]
		oldRecord, found := tx.records[oldRecordKey]
		if !ok || !found || oldRecord.Status != StatusActive || oldRecord.Domain != request.Domain || oldRecord.IssuerCommitment != request.IssuerCommitment || oldRecord.PolicyDigest != request.PolicyDigest || oldRecord.ProfileDigest != request.ProfileDigest || oldRecord.WalletKeyCommitment != request.OldWalletKeyCommitment || oldRecord.WalletKeyEpoch != request.OldWalletKeyEpoch {
			return IssuerLinkRecord{}, tx.fail(errors.New("rotation does not exactly bind the active predecessor"))
		}
		record.PredecessorNullifierCommitment = nullifierCommitment(oldRecord.ScopedNullifier)
		record.RotationRequestDigest = request.RequestDigest
	}
	if err := record.Validate(); err != nil {
		return IssuerLinkRecord{}, tx.fail(err)
	}
	recordKey := scopedRecordKey(tx.domainKey, record.ScopedNullifier)
	if _, exists := tx.records[recordKey]; exists {
		return IssuerLinkRecord{}, tx.fail(ErrDuplicateLink)
	}
	walletKey := scopedWalletKey(tx.domainKey, record.WalletKeyCommitment)
	if _, exists := tx.wallets[walletKey]; exists {
		return IssuerLinkRecord{}, tx.fail(ErrDuplicateLink)
	}
	tx.records[recordKey] = record
	tx.wallets[walletKey] = recordKey
	tx.inserted = true
	if tx.failAfterInsertion {
		return IssuerLinkRecord{}, tx.fail(errors.New("injected failure after insertion"))
	}
	return record, nil
}

func (f *MemoryAtomicIssuerLinkFixture) Lookup(ctx context.Context, domain NullifierDomain, nullifier string, capability LookupCapability) (IssuerLinkProjection, error) {
	if err := ctx.Err(); err != nil {
		return IssuerLinkProjection{}, err
	}
	domainKey, err := domain.Key()
	if err != nil || !validDigest(nullifier) || capability.Claim.Domain != domain || capability.Claim.Nullifier != nullifier {
		return IssuerLinkProjection{}, ErrUnauthorized
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.lookupAuthorizer == nil || capability.Claim.CurrentCoordinate != f.coordinate || capability.Claim.ExpiresAtCoordinate <= f.coordinate {
		return IssuerLinkProjection{}, ErrUnauthorized
	}
	if err := f.lookupAuthorizer.VerifyLookup(ctx, capability); err != nil {
		return IssuerLinkProjection{}, err
	}
	recordKey := scopedRecordKey(domainKey, nullifier)
	if tombstone, ok := f.tombstones[recordKey]; ok {
		copy := tombstone
		return IssuerLinkProjection{Version: Version1, Status: StatusDeleted, Tombstone: &copy}, nil
	}
	record, ok := f.records[recordKey]
	if !ok {
		return IssuerLinkProjection{}, errors.New("issuer link not found in exact domain")
	}
	projection := IssuerLinkProjection{Version: Version1, Status: record.Status, ScopedNullifier: record.ScopedNullifier, IssuerCommitment: record.IssuerCommitment, ProgramDigest: record.Domain.ProgramDigest, PolicyDigest: record.PolicyDigest, ProfileDigest: record.ProfileDigest, WalletKeyEpoch: record.WalletKeyEpoch}
	if capability.Claim.RequireWallet {
		projection.WalletKeyCommitment = record.WalletKeyCommitment
	}
	return projection, nil
}

func (f *MemoryAtomicIssuerLinkFixture) Transition(ctx context.Context, domain NullifierDomain, nullifier string, action TransitionAction) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	domainKey, err := domain.Key()
	if err != nil || !validDigest(nullifier) {
		return errors.New("valid exact-domain record coordinate is required")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	recordKey := scopedRecordKey(domainKey, nullifier)
	previous, ok := f.records[recordKey]
	if !ok {
		return errors.New("issuer link not found in exact domain")
	}
	next := previous
	next.UpdatedAtCoordinate, next.UpdatedAtUnix = f.coordinate, int64(f.coordinate)
	switch action.Kind {
	case TransitionBeginCooldown:
		if previous.Status != StatusPending || f.coordinate >= previous.CooldownUntilCoordinate {
			return errors.New("cooldown must begin before its activation coordinate")
		}
		next.Status = StatusCooldown
	case TransitionActivate:
		if previous.Status != StatusCooldown || previous.PredecessorNullifierCommitment != "" || f.coordinate < previous.CooldownUntilCoordinate {
			return errors.New("activation requires elapsed non-rotation cooldown")
		}
		next.Status = StatusActive
	case TransitionAppeal:
		if !validDigest(action.AppealCommitment) {
			return errors.New("strict appeal commitment is required")
		}
		next.Status, next.AppealCommitment = StatusAppealed, action.AppealCommitment
	case TransitionResolveAppeal:
		commitment, _ := previous.commitment()
		claim := action.Resolution.claim
		if previous.Status != StatusAppealed || action.Resolution.digest == "" || action.Resolution.scope != f.authorizationScope || claim.Domain != domain || claim.Nullifier != nullifier || claim.RecordCommitment != commitment || claim.CurrentCoordinate != f.coordinate || claim.ExpiresAtCoordinate <= f.coordinate || claim.FreshCooldownCoordinate <= f.coordinate {
			return ErrUnauthorized
		}
		next.Status, next.AppealCommitment = StatusCooldown, ""
		next.CooldownUntilCoordinate = claim.FreshCooldownCoordinate
	case TransitionRevoke:
		next.Status, next.AppealCommitment = StatusRevoked, ""
	default:
		return errors.New("unknown issuer-link transition action")
	}
	if err := validateStatusTransition(previous, next); err != nil {
		return err
	}
	clone := cloneRecordMap(f.records)
	clone[recordKey] = next
	f.records = clone
	return nil
}

func (f *MemoryAtomicIssuerLinkFixture) ActivateRotation(ctx context.Context, domain NullifierDomain, newNullifier, supersessionCommitment, notificationCommitment string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	domainKey, err := domain.Key()
	if err != nil || !validDigests(newNullifier, supersessionCommitment, notificationCommitment) {
		return errors.New("complete rotation activation commitments are required")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	newKey := scopedRecordKey(domainKey, newNullifier)
	newRecord, ok := f.records[newKey]
	if !ok || newRecord.Status != StatusCooldown || newRecord.PredecessorNullifierCommitment == "" || newRecord.RotationRequestDigest != newRecord.AuthorizedRequestDigest || f.coordinate < newRecord.CooldownUntilCoordinate {
		return errors.New("rotation activation requires authorized lineage and elapsed cooldown")
	}
	var oldKey string
	var oldRecord IssuerLinkRecord
	for candidateKey, candidate := range f.records {
		if candidate.Domain == domain && candidate.Status == StatusActive && nullifierCommitment(candidate.ScopedNullifier) == newRecord.PredecessorNullifierCommitment {
			oldKey, oldRecord = candidateKey, candidate
			break
		}
	}
	if oldKey == "" || oldRecord.IssuerCommitment != newRecord.IssuerCommitment || oldRecord.PolicyDigest != newRecord.PolicyDigest || oldRecord.ProfileDigest != newRecord.ProfileDigest || oldRecord.WalletKeyEpoch >= newRecord.WalletKeyEpoch {
		return errors.New("rotation predecessor lineage mismatch")
	}
	oldNext, newNext := oldRecord, newRecord
	oldNext.Status, oldNext.UpdatedAtCoordinate, oldNext.UpdatedAtUnix = StatusSuperseded, f.coordinate, int64(f.coordinate)
	oldNext.SupersessionCommitment, oldNext.NotificationCommitment = supersessionCommitment, notificationCommitment
	newNext.Status, newNext.UpdatedAtCoordinate, newNext.UpdatedAtUnix = StatusActive, f.coordinate, int64(f.coordinate)
	if err := validateStatusTransition(oldRecord, oldNext); err != nil {
		return err
	}
	if err := validateStatusTransition(newRecord, newNext); err != nil {
		return err
	}
	records, wallets := cloneRecordMap(f.records), cloneStringMap(f.wallets)
	records[oldKey], records[newKey] = oldNext, newNext
	delete(wallets, scopedWalletKey(domainKey, oldRecord.WalletKeyCommitment))
	wallets[scopedWalletKey(domainKey, newRecord.WalletKeyCommitment)] = newKey
	f.records, f.wallets = records, wallets
	return nil
}

func (f *MemoryAtomicIssuerLinkFixture) ApplyRetention(ctx context.Context, domain NullifierDomain, nullifier string, authorization VerifiedRetentionAuthorization) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	domainKey, err := domain.Key()
	if err != nil || !validDigest(nullifier) {
		return errors.New("valid retention target is required")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	recordKey := scopedRecordKey(domainKey, nullifier)
	record, ok := f.records[recordKey]
	if !ok {
		return errors.New("issuer link not found in exact domain")
	}
	commitment, _ := record.commitment()
	claim := authorization.claim
	if authorization.digest == "" || authorization.scope != f.authorizationScope || claim.Domain != domain || claim.Nullifier != nullifier || claim.RecordCommitment != commitment || claim.CurrentCoordinate != f.coordinate || claim.ExpiresAtCoordinate <= f.coordinate || claim.LegalHold != (record.Retention == RetentionLegalHold) {
		return ErrUnauthorized
	}
	record.UpdatedAtCoordinate, record.UpdatedAtUnix = f.coordinate, int64(f.coordinate)
	switch claim.Action {
	case RetentionHoldAction:
		if claim.LegalHold {
			return errors.New("record is already under legal hold")
		}
		record.Retention = RetentionLegalHold
	case RetentionReleaseAction:
		if !claim.LegalHold {
			return errors.New("only governed authorization can release a legal hold")
		}
		record.Retention = RetentionRequired
	case RetentionEligibleAction:
		if claim.LegalHold || (record.Status != StatusRevoked && record.Status != StatusSuperseded) {
			return errors.New("record is not eligible for deletion")
		}
		record.Retention, record.Status = RetentionEligible, StatusDeletionPending
	case RetentionDeleteAction:
		if claim.LegalHold || record.Retention != RetentionEligible || record.Status != StatusDeletionPending {
			return errors.New("deletion is not authorized for current retention state")
		}
		tombstone := DeletionTombstone{Version: Version1, RecordCommitment: commitment, DeletionAuthorizationDigest: authorization.digest, DeletedAtUnix: int64(f.coordinate), DeletedAtCoordinate: f.coordinate}
		records, wallets, tombstones := cloneRecordMap(f.records), cloneStringMap(f.wallets), cloneTombstoneMap(f.tombstones)
		delete(records, recordKey)
		delete(wallets, scopedWalletKey(domainKey, record.WalletKeyCommitment))
		tombstones[recordKey] = tombstone
		f.records, f.wallets, f.tombstones = records, wallets, tombstones
		return nil
	default:
		return errors.New("unknown retention action")
	}
	if err := record.Validate(); err != nil {
		return err
	}
	records := cloneRecordMap(f.records)
	records[recordKey] = record
	f.records = records
	return nil
}

func (f *MemoryAtomicIssuerLinkFixture) Authenticate(ctx context.Context, domain NullifierDomain, nullifier string, publicKey ed25519.PublicKey, signature []byte, challengeDigest string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	domainKey, err := domain.Key()
	if err != nil || !validDigests(nullifier, challengeDigest) {
		return errors.New("valid authentication coordinate is required")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	record, ok := f.records[scopedRecordKey(domainKey, nullifier)]
	if !ok || record.Status != StatusActive || !publicKeyCommitmentMatches(publicKey, record.WalletKeyCommitment) {
		return errors.New("canonical wallet link is not active")
	}
	message, err := AuthenticationSigningBytes(domain, nullifier, challengeDigest, f.coordinate)
	if err != nil || len(signature) != ed25519.SignatureSize || !ed25519.Verify(publicKey, message, signature) {
		return errors.New("invalid canonical wallet authentication proof")
	}
	replayKey := authenticationReplayKey{
		domainKey:       domainKey,
		nullifier:       nullifier,
		challengeDigest: challengeDigest,
		coordinate:      f.coordinate,
	}
	if _, consumed := f.authentications[replayKey]; consumed {
		return ErrAuthenticationReplay
	}
	f.authentications[replayKey] = struct{}{}
	return nil
}

func scopedRecordKey(domainKey, nullifier string) string { return domainKey + ":" + nullifier }
func scopedWalletKey(domainKey, wallet string) string    { return domainKey + ":" + wallet }

func cloneRecordMap(source map[string]IssuerLinkRecord) map[string]IssuerLinkRecord {
	clone := make(map[string]IssuerLinkRecord, len(source))
	for key, value := range source {
		clone[key] = value
	}
	return clone
}

func cloneIdempotencyMap(source map[string]idempotencyEntry) map[string]idempotencyEntry {
	clone := make(map[string]idempotencyEntry, len(source))
	for key, value := range source {
		clone[key] = value
	}
	return clone
}

func cloneNonceMap(source map[string]nonceEntry) map[string]nonceEntry {
	clone := make(map[string]nonceEntry, len(source))
	for key, value := range source {
		clone[key] = value
	}
	return clone
}

func cloneStringMap(source map[string]string) map[string]string {
	clone := make(map[string]string, len(source))
	for key, value := range source {
		clone[key] = value
	}
	return clone
}

func cloneTombstoneMap(source map[string]DeletionTombstone) map[string]DeletionTombstone {
	clone := make(map[string]DeletionTombstone, len(source))
	for key, value := range source {
		clone[key] = value
	}
	return clone
}

var (
	_ ThresholdPRF          = (*DeterministicThresholdPRFFixture)(nil)
	_ IssuerKeyResolver     = GovernedIssuerKeyResolverFixture{}
	_ LookupAuthorizer      = Ed25519AuthorizerFixture{}
	_ AppealAuthorizer      = Ed25519AuthorizerFixture{}
	_ RetentionAuthorizer   = Ed25519AuthorizerFixture{}
	_ AtomicIssuerLinkStore = (*MemoryAtomicIssuerLinkFixture)(nil)
)
