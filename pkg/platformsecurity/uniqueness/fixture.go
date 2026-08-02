package uniqueness

import (
	"context"
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"
)

const fixtureWarning = "fixture_only: deterministic HMAC/Ed25519 fixtures have no OPRF, MPC, TEE, or external security review"

type DeterministicThresholdPRFFixture struct {
	key []byte
}

func NewDeterministicThresholdPRFFixture(key []byte) (*DeterministicThresholdPRFFixture, error) {
	if len(key) < 16 {
		return nil, errors.New("fixture PRF key must contain at least 16 bytes")
	}
	return &DeterministicThresholdPRFFixture{key: slices.Clone(key)}, nil
}

func (f *DeterministicThresholdPRFFixture) FixtureState() string  { return FixtureOnlyState }
func (f *DeterministicThresholdPRFFixture) ProductionReady() bool { return false }
func (f *DeterministicThresholdPRFFixture) ReviewStatus() string  { return fixtureWarning }

func (f *DeterministicThresholdPRFFixture) Evaluate(_ context.Context, input NullifierInput, authorization *VerifiedFinalUniqueAuthorization) (ScopedNullifier, error) {
	canonicalInput, err := input.CanonicalBytes()
	if err != nil {
		return ScopedNullifier{}, err
	}
	if authorization == nil || authorization.inputDigest != digest(canonicalInput) {
		return ScopedNullifier{}, errors.New("verified final-unique authorization does not bind nullifier input")
	}
	mac := hmac.New(sha256.New, f.key)
	_, _ = mac.Write(canonicalInput)
	return ScopedNullifier{
		Version: Version1, Domain: input.Domain, PolicyDigest: input.PolicyDigest,
		ProfileDigest: input.ProfileDigest, KeyEpoch: input.KeyEpoch, Value: fmt.Sprintf("%x", mac.Sum(nil)),
	}, nil
}

type deterministicTemplateArtifact struct {
	commitment  string
	coordinates [4]int64
}

func (a deterministicTemplateArtifact) Commitment() string { return a.commitment }

type DeterministicTemplateFixture struct {
	key []byte
}

func NewDeterministicTemplateFixture(key []byte) (*DeterministicTemplateFixture, error) {
	if len(key) < 16 {
		return nil, errors.New("fixture transform key must contain at least 16 bytes")
	}
	return &DeterministicTemplateFixture{key: slices.Clone(key)}, nil
}

func (f *DeterministicTemplateFixture) FixtureState() string  { return FixtureOnlyState }
func (f *DeterministicTemplateFixture) ProductionReady() bool { return false }
func (f *DeterministicTemplateFixture) ReviewStatus() string  { return fixtureWarning }

func (f *DeterministicTemplateFixture) Transform(_ context.Context, profile CancellableTemplateProfile, opaqueInput []byte) (TemplateArtifact, error) {
	if err := profile.Validate(); err != nil {
		return nil, err
	}
	if len(opaqueInput) == 0 {
		return nil, errors.New("opaque fixture input is required")
	}
	mac := hmac.New(sha256.New, f.key)
	_, _ = mac.Write([]byte("virtengine.uniqueness.template-fixture/v1\x00"))
	_, _ = mac.Write([]byte(profile.ProfileDigest))
	_, _ = mac.Write([]byte{0})
	_, _ = mac.Write(opaqueInput)
	output := mac.Sum(nil)
	artifact := deterministicTemplateArtifact{commitment: digest(output)}
	for index := range artifact.coordinates {
		artifact.coordinates[index] = int64(binary.BigEndian.Uint64(output[index*8:(index+1)*8]) & 0x7fffffffffffffff)
	}
	return artifact, nil
}

type candidateEntry struct {
	recordDigest string
	artifact     deterministicTemplateArtifact
}

type DeterministicCandidateSearcherFixture struct {
	mu      sync.RWMutex
	entries []candidateEntry
}

func (f *DeterministicCandidateSearcherFixture) FixtureState() string  { return FixtureOnlyState }
func (f *DeterministicCandidateSearcherFixture) ProductionReady() bool { return false }
func (f *DeterministicCandidateSearcherFixture) ReviewStatus() string  { return fixtureWarning }

func (f *DeterministicCandidateSearcherFixture) Add(recordDigest string, artifact TemplateArtifact) error {
	template, ok := artifact.(deterministicTemplateArtifact)
	if !ok || !validDigest(recordDigest) {
		return errors.New("deterministic fixture artifact and record digest are required")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.entries = append(f.entries, candidateEntry{recordDigest: recordDigest, artifact: template})
	return nil
}

func (f *DeterministicCandidateSearcherFixture) Search(_ context.Context, artifact TemplateArtifact, threshold int64) (CandidateSearchResult, error) {
	template, ok := artifact.(deterministicTemplateArtifact)
	if !ok || threshold < 0 {
		return CandidateSearchResult{}, errors.New("deterministic fixture artifact and non-negative threshold are required")
	}
	f.mu.RLock()
	defer f.mu.RUnlock()
	return candidateResult(searchEntries(f.entries, template, threshold)), nil
}

type MemoryAtomicEnrollmentFixture struct {
	mu          sync.Mutex
	records     map[string]EnrollmentRecord
	entries     []candidateEntry
	failAtStage EnrollmentStage
	now         func() time.Time
}

func NewMemoryAtomicEnrollmentFixture() *MemoryAtomicEnrollmentFixture {
	return &MemoryAtomicEnrollmentFixture{records: make(map[string]EnrollmentRecord), now: time.Now}
}

func (f *MemoryAtomicEnrollmentFixture) FixtureState() string  { return FixtureOnlyState }
func (f *MemoryAtomicEnrollmentFixture) ProductionReady() bool { return false }
func (f *MemoryAtomicEnrollmentFixture) ReviewStatus() string  { return fixtureWarning }

func (f *MemoryAtomicEnrollmentFixture) InjectFailureAfter(stage EnrollmentStage) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.failAtStage = stage
}

func (f *MemoryAtomicEnrollmentFixture) Enroll(ctx context.Context, request EnrollmentRequest, callback func(EnrollmentTransaction) (EnrollmentRecord, error)) (EnrollmentRecord, error) {
	if callback == nil {
		return EnrollmentRecord{}, errors.New("transaction callback is required")
	}
	if err := request.ValidateStructure(); err != nil {
		return EnrollmentRecord{}, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if existing, ok := f.records[request.IdempotencyKey]; ok {
		if err := requireSameRequest(request, existing); err != nil {
			return EnrollmentRecord{}, err
		}
		return existing, nil
	}
	if err := request.Validate(f.now()); err != nil {
		return EnrollmentRecord{}, err
	}
	if err := ctx.Err(); err != nil {
		return EnrollmentRecord{}, err
	}
	records := cloneRecords(f.records)
	entries := slices.Clone(f.entries)
	tx := &memoryEnrollmentTransaction{ctx: ctx, records: records, entries: entries, failAtStage: f.failAtStage}
	if err := request.Validate(f.now()); err != nil {
		return EnrollmentRecord{}, err
	}
	record, err := callback(tx)
	if err != nil {
		return EnrollmentRecord{}, err
	}
	if tx.aborted != nil {
		return EnrollmentRecord{}, fmt.Errorf("transaction was aborted: %w", tx.aborted)
	}
	if err := ctx.Err(); err != nil {
		return EnrollmentRecord{}, err
	}
	if err := request.Validate(f.now()); err != nil {
		return EnrollmentRecord{}, err
	}
	if !tx.inserted || tx.stage != StageInsertion {
		return EnrollmentRecord{}, errors.New("transaction did not atomically complete all enrollment stages")
	}
	if err := requireSameRequest(request, record); err != nil {
		return EnrollmentRecord{}, err
	}
	stored, ok := tx.records[request.IdempotencyKey]
	if !ok || stored != record {
		return EnrollmentRecord{}, errors.New("transaction callback returned a record other than the inserted record")
	}
	f.records, f.entries = tx.records, tx.entries
	f.failAtStage = ""
	return record, nil
}

func (f *MemoryAtomicEnrollmentFixture) Count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.records)
}

type memoryEnrollmentTransaction struct {
	ctx           context.Context
	records       map[string]EnrollmentRecord
	entries       []candidateEntry
	failAtStage   EnrollmentStage
	stage         EnrollmentStage
	searched      deterministicTemplateArtifact
	possibleMatch bool
	inserted      bool
	aborted       error
}

func (tx *memoryEnrollmentTransaction) fail(err error) error {
	if tx.aborted == nil {
		tx.aborted = err
	}
	return err
}

func (tx *memoryEnrollmentTransaction) live() error {
	if tx.aborted != nil {
		return tx.aborted
	}
	if err := tx.ctx.Err(); err != nil {
		return tx.fail(err)
	}
	return nil
}

func (tx *memoryEnrollmentTransaction) Search(ctx context.Context, artifact TemplateArtifact, threshold int64) (CandidateSearchResult, error) {
	if err := tx.live(); err != nil {
		return CandidateSearchResult{}, err
	}
	if tx.stage != "" || ctx.Err() != nil {
		return CandidateSearchResult{}, tx.fail(errors.New("search must be the first live transaction stage"))
	}
	template, ok := artifact.(deterministicTemplateArtifact)
	if !ok || threshold < 0 {
		return CandidateSearchResult{}, tx.fail(errors.New("deterministic fixture artifact and non-negative threshold are required"))
	}
	tx.stage, tx.searched = StageSearch, template
	result := candidateResult(searchEntries(tx.entries, template, threshold))
	tx.possibleMatch = result.PossibleMatch
	if tx.failAtStage == StageSearch {
		return CandidateSearchResult{}, tx.fail(errors.New("injected failure after search"))
	}
	return result, nil
}

func (tx *memoryEnrollmentTransaction) Advance(stage EnrollmentStage) error {
	if err := tx.live(); err != nil {
		return err
	}
	expected := map[EnrollmentStage]EnrollmentStage{StageSearch: StageAdjudication, StageAdjudication: StageNullifier}[tx.stage]
	if stage != expected {
		return tx.fail(errors.New("enrollment transaction stages must be ordered search, adjudication, nullifier, insertion"))
	}
	tx.stage = stage
	if tx.failAtStage == stage {
		return tx.fail(fmt.Errorf("injected failure after %s", stage))
	}
	return nil
}

func (tx *memoryEnrollmentTransaction) Insert(record EnrollmentRecord, artifact TemplateArtifact) error {
	if err := tx.live(); err != nil {
		return err
	}
	if tx.stage != StageNullifier || tx.inserted {
		return tx.fail(errors.New("insertion requires completed search, adjudication, and nullifier stages"))
	}
	template, ok := artifact.(deterministicTemplateArtifact)
	if !ok || template.commitment != tx.searched.commitment || record.TemplateCommitment != template.commitment {
		return tx.fail(errors.New("inserted template does not match searched commitment"))
	}
	if err := record.Validate(); err != nil {
		return tx.fail(err)
	}
	if record.Outcome == OutcomeFinalUnique && tx.possibleMatch {
		return tx.fail(errors.New("final unique insertion conflicts with a possible-match candidate"))
	}
	if existing, ok := tx.records[record.IdempotencyKey]; ok {
		if existing != record {
			return tx.fail(ErrConflict)
		}
		return nil
	}
	tx.records[record.IdempotencyKey] = record
	tx.entries = append(tx.entries, candidateEntry{recordDigest: record.RequestDigest, artifact: template})
	tx.stage, tx.inserted = StageInsertion, true
	if tx.failAtStage == StageInsertion {
		return tx.fail(errors.New("injected failure after insertion"))
	}
	return nil
}

type internalCandidate struct {
	recordDigest string
	distance     int64
}

func searchEntries(entries []candidateEntry, template deterministicTemplateArtifact, threshold int64) []internalCandidate {
	result := make([]internalCandidate, 0)
	for _, entry := range entries {
		distance := fixedDistance(template, entry.artifact)
		if distance <= threshold {
			result = append(result, internalCandidate{recordDigest: entry.recordDigest, distance: distance})
		}
	}
	slices.SortFunc(result, func(left, right internalCandidate) int {
		if left.distance != right.distance {
			return cmpInt64(left.distance, right.distance)
		}
		if left.recordDigest < right.recordDigest {
			return -1
		}
		if left.recordDigest > right.recordDigest {
			return 1
		}
		return 0
	})
	return result
}

func candidateResult(candidates []internalCandidate) CandidateSearchResult {
	e := newCanonicalEncoder("virtengine.uniqueness.candidate-set/v1")
	count := len(candidates)
	if count > int(MaxCandidateCount) {
		count = int(MaxCandidateCount)
	}
	e.u32(uint32(count))
	for _, candidate := range candidates[:count] {
		e.text(candidate.recordDigest)
		e.i64(candidate.distance)
	}
	commitment := digest(e.result())
	state := CandidateClear
	if count > 0 {
		state = CandidateReviewRequired
	}
	return CandidateSearchResult{PossibleMatch: count > 0, ReviewState: state, CandidateCount: uint32(count), CandidateSetCommitment: commitment, AdjudicationReference: digest([]byte("virtengine.uniqueness.adjudication/v1\x00" + commitment))}
}

func fixedDistance(left, right deterministicTemplateArtifact) int64 {
	var distance uint64
	for index := range left.coordinates {
		a, b := uint64(left.coordinates[index]), uint64(right.coordinates[index])
		if a > b {
			distance += a - b
		} else {
			distance += b - a
		}
		if distance > uint64(^uint64(0)>>1) {
			return int64(^uint64(0) >> 1)
		}
	}
	return int64(distance)
}

func cmpInt64(left, right int64) int {
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}

func cloneRecords(source map[string]EnrollmentRecord) map[string]EnrollmentRecord {
	clone := make(map[string]EnrollmentRecord, len(source))
	for key, record := range source {
		clone[key] = record
	}
	return clone
}

type DeterministicCustodyKeyResolverFixture struct {
	Epochs map[uint64]CustodyNodeSet
}

func (f DeterministicCustodyKeyResolverFixture) ResolveCustodyNodes(_ context.Context, epoch uint64) (CustodyNodeSet, error) {
	nodes, ok := f.Epochs[epoch]
	if !ok {
		return CustodyNodeSet{}, errors.New("unknown custody key epoch")
	}
	return nodes, nil
}

type DeterministicQuorumAttestorFixture struct {
	Nodes       CustodyNodeSet
	PrivateKeys map[string]ed25519.PrivateKey
}

func (f DeterministicQuorumAttestorFixture) FixtureState() string  { return FixtureOnlyState }
func (f DeterministicQuorumAttestorFixture) ProductionReady() bool { return false }
func (f DeterministicQuorumAttestorFixture) ReviewStatus() string  { return fixtureWarning }

func (f DeterministicQuorumAttestorFixture) Attest(_ context.Context, payload CustodyReceiptPayload, epoch ThresholdKeyEpoch) (QuorumAttestation, error) {
	if epoch.Epoch != payload.KeyEpoch || epoch.State != KeyEpochActive {
		return QuorumAttestation{}, errors.New("attestation requires the current active key epoch")
	}
	if err := epoch.ValidateWithNodeSet(f.Nodes); err != nil {
		return QuorumAttestation{}, err
	}
	payloadBytes, err := payload.CanonicalBytes()
	if err != nil {
		return QuorumAttestation{}, err
	}
	return f.attestBytes(payloadBytes, epoch.Epoch, payload.ParticipantSetDigest, payload.Threshold)
}

func (f DeterministicQuorumAttestorFixture) AttestFinalUnique(_ context.Context, authorization FinalUniqueAuthorization) (QuorumAttestation, error) {
	payloadBytes, err := authorization.CanonicalBytes()
	if err != nil {
		return QuorumAttestation{}, err
	}
	return f.attestBytes(payloadBytes, authorization.KeyEpoch, authorization.ParticipantSetDigest, authorization.Threshold)
}

func (f DeterministicQuorumAttestorFixture) attestBytes(payloadBytes []byte, epoch uint64, participantSetDigest string, threshold uint32) (QuorumAttestation, error) {
	setDigest, err := f.Nodes.Digest()
	if err != nil {
		return QuorumAttestation{}, err
	}
	if participantSetDigest != setDigest || threshold != f.Nodes.Threshold {
		return QuorumAttestation{}, errors.New("attestation authority mismatch")
	}
	payloadDigest := digest(payloadBytes)
	nodes := slices.Clone(f.Nodes.Nodes)
	slices.SortFunc(nodes, func(left, right CustodyNodeIdentity) int {
		if left.NodeID < right.NodeID {
			return -1
		}
		if left.NodeID > right.NodeID {
			return 1
		}
		return 0
	})
	signerIDs := make([]string, 0, f.Nodes.Threshold)
	for _, node := range nodes {
		if node.State == NodeActive && uint32(len(signerIDs)) < f.Nodes.Threshold {
			signerIDs = append(signerIDs, node.NodeID)
		}
	}
	unsigned := make([]NodeSignature, 0, len(signerIDs))
	for _, nodeID := range signerIDs {
		unsigned = append(unsigned, NodeSignature{NodeID: nodeID})
	}
	signerSetDigest, err := signerSetDigest(unsigned)
	if err != nil {
		return QuorumAttestation{}, err
	}
	signed := attestationSigningBytes(payloadDigest, setDigest, threshold, signerSetDigest)
	attestation := QuorumAttestation{Version: Version1, PayloadDigest: payloadDigest, ParticipantSetDigest: setDigest, Threshold: threshold, SignerSetDigest: signerSetDigest}
	for _, nodeID := range signerIDs {
		node := nodeByID(nodes, nodeID)
		privateKey := f.PrivateKeys[node.SigningKeyID]
		if len(privateKey) != ed25519.PrivateKeySize || !privateKey.Public().(ed25519.PublicKey).Equal(ed25519.PublicKey(node.SigningPublicKey)) {
			return QuorumAttestation{}, errors.New("fixture signing key is missing or aliased")
		}
		if node.SigningKeyEpoch != epoch {
			return QuorumAttestation{}, errors.New("fixture node key epoch mismatch")
		}
		attestation.Signatures = append(attestation.Signatures, NodeSignature{NodeID: node.NodeID, SigningKeyID: node.SigningKeyID, SigningKeyEpoch: epoch, Signature: ed25519.Sign(privateKey, signed)})
	}
	return attestation, nil
}

func nodeByID(nodes []CustodyNodeIdentity, id string) CustodyNodeIdentity {
	for _, node := range nodes {
		if node.NodeID == id {
			return node
		}
	}
	return CustodyNodeIdentity{}
}
