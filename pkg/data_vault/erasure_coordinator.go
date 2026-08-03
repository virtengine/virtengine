package data_vault

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/virtengine/virtengine/pkg/data_vault/contracts"
)

var (
	ErrErasureHeld       = errors.New("erasure blocked by legal hold")
	ErrBackupFence       = errors.New("governed backup generations remain available")
	ErrErasureConflict   = errors.New("erasure operation request conflict")
	ErrConsentErasure    = errors.New("consent decision does not authorize erasure")
	ErrProductionBackend = errors.New("production erasure backend is not provided")
)

type ErasureClaims struct {
	Version                uint32                      `json:"version"`
	OperationID            string                      `json:"operation_id"`
	RequestDigest          string                      `json:"request_digest"`
	Target                 contracts.EvidenceObjectRef `json:"target"`
	AuthorizationDigest    string                      `json:"authorization_digest"`
	PolicyDigest           string                      `json:"policy_digest"`
	ProfileDigest          string                      `json:"profile_digest"`
	KeyEpoch               uint64                      `json:"key_epoch"`
	ObjectCommitment       string                      `json:"object_commitment"`
	StorageCommitment      string                      `json:"storage_commitment"`
	KeyCommitment          string                      `json:"key_commitment"`
	BackupGenerationDigest string                      `json:"backup_generation_digest"`
	BackupExpiryHeight     int64                       `json:"backup_expiry_height"`
	BackupExpiryUnix       int64                       `json:"backup_expiry_unix"`
	ConsentDecisionDigest  string                      `json:"consent_decision_digest"`
	ConsentPolicyDigest    string                      `json:"consent_policy_digest"`
	ErasureEpoch           uint64                      `json:"erasure_epoch"`
	CurrentHeight          int64                       `json:"current_height"`
	CurrentUnix            int64                       `json:"current_unix"`
}

type ErasureRequest struct {
	ErasureClaims
	StorageOperationID string `json:"storage_operation_id"`
	KMSOperationID     string `json:"kms_operation_id"`
}

type StorageDeletionRequest struct{ ErasureClaims }
type KMSDestructionRequest struct{ ErasureClaims }

type StorageDeletionAdapter interface {
	Delete(context.Context, StorageDeletionRequest) (contracts.DeletionReceipt, error)
}

type KMSDestructionAdapter interface {
	Destroy(context.Context, KMSDestructionRequest) (contracts.DeletionReceipt, error)
}

type HoldReader interface {
	HasActiveHold(context.Context, contracts.EvidenceObjectRef, string) (bool, error)
}

type ConsentDecision struct {
	DecisionDigest    string
	PolicyDigest      string
	RestoreAllowed    bool
	ErasureAuthorized bool
}

type ConsentDecisionReader interface {
	ReadConsentDecision(context.Context, contracts.EvidenceObjectRef) (ConsentDecision, error)
}

func ValidateConsentAccess(decision ConsentDecision, restoreOrPlaintext bool) error {
	if restoreOrPlaintext && !decision.RestoreAllowed {
		return errors.New("consent withdrawal prohibits restore and plaintext access")
	}
	return nil
}

type BackupFenceStatus struct {
	GenerationDigest    string
	ExpiryHeight        int64
	ExpiryUnix          int64
	AllExpiredOrDeleted bool
}

type BackupFence interface {
	CheckBackupFence(context.Context, ErasureClaims) (BackupFenceStatus, error)
}

// ErasureFinalizationFence prevents legal-hold, consent, and backup authority
// state from changing between the final check and the atomic resolution commit.
type ErasureFinalizationFence interface {
	AcquireErasureFinalization(context.Context, ErasureClaims) (func(), error)
}

type DeletionReceiptSigner interface {
	SignDeletionReceipt(context.Context, contracts.DeletionReceipt) (contracts.DeletionReceipt, error)
}

type ErasureOperationState string

const (
	ErasureIntent         ErasureOperationState = "intent"
	ErasureStorageDeleted ErasureOperationState = "storage_deleted"
	ErasureBackupFenced   ErasureOperationState = "backup_fenced"
	ErasureKMSDestroyed   ErasureOperationState = "kms_destroyed"
	ErasureResolved       ErasureOperationState = "resolved"
	ErasureHeld           ErasureOperationState = "held"
	ErasureUnresolved     ErasureOperationState = "unresolved"
)

type ErasureJournalEntry struct {
	Sequence uint64                `json:"sequence"`
	State    ErasureOperationState `json:"state"`
	Event    string                `json:"event"`
}

type ErasureOperation struct {
	Request        ErasureRequest               `json:"request"`
	RequestDigest  string                       `json:"request_digest"`
	State          ErasureOperationState        `json:"state"`
	ResumeState    ErasureOperationState        `json:"resume_state,omitempty"`
	Revision       uint64                       `json:"revision"`
	StorageReceipt *contracts.DeletionReceipt   `json:"storage_receipt,omitempty"`
	KMSReceipt     *contracts.DeletionReceipt   `json:"kms_receipt,omitempty"`
	Resolved       *contracts.EvidenceObjectRef `json:"resolved,omitempty"`
	Journal        []ErasureJournalEntry        `json:"journal"`
}

type ErasureOperationStore interface {
	AcquireOperation(context.Context, string, string) (func(), error)
	Begin(context.Context, ErasureRequest) (ErasureOperation, error)
	Load(context.Context, string) (ErasureOperation, error)
	Update(context.Context, ErasureOperation) error
	ListUnresolved(context.Context) ([]ErasureOperation, error)
	Resolve(context.Context, string, string, contracts.DeletionReceiptReplay, contracts.DeletionReceiptReplay, func(ErasureResolutionTransaction, *ErasureOperation) error) error
}

// ErasureResolutionTransaction is supplied by the durable operation store.
// ApplyResolved must stage the domain write in the same transaction as replay
// consumption and operation resolution.
type ErasureResolutionTransaction interface {
	ApplyResolved(contracts.EvidenceObjectRef) error
}

type ErasureCoordinator struct {
	store   ErasureOperationStore
	storage StorageDeletionAdapter
	kms     KMSDestructionAdapter
	holds   HoldReader
	consent ConsentDecisionReader
	backups BackupFence
	finalize ErasureFinalizationFence
	keys    contracts.DeletionReceiptKeyResolver
}

// NewErasureCoordinator wires contracts only. Production remains unavailable
// until external durable storage, KMS/storage signers, backup, hold, and consent
// authorities implement these interfaces and the apply callback joins the store transaction.
func NewErasureCoordinator(store ErasureOperationStore, storage StorageDeletionAdapter, kms KMSDestructionAdapter, holds HoldReader, consent ConsentDecisionReader, backups BackupFence, finalize ErasureFinalizationFence, keys contracts.DeletionReceiptKeyResolver) (*ErasureCoordinator, error) {
	if store == nil || storage == nil || kms == nil || holds == nil || consent == nil || backups == nil || finalize == nil || keys == nil {
		return nil, errors.New("erasure coordinator requires durable store, adapters, and authorities")
	}
	return &ErasureCoordinator{store: store, storage: storage, kms: kms, holds: holds, consent: consent, backups: backups, finalize: finalize, keys: keys}, nil
}

func (c *ErasureCoordinator) Erase(ctx context.Context, request ErasureRequest) (ErasureOperation, error) {
	if err := prepareErasureRequest(&request); err != nil {
		return ErasureOperation{}, err
	}
	release, err := c.store.AcquireOperation(ctx, request.OperationID, request.RequestDigest)
	if err != nil {
		return ErasureOperation{}, err
	}
	defer release()
	op, err := c.store.Begin(ctx, request)
	if err != nil {
		return ErasureOperation{}, err
	}
	return c.resume(ctx, op)
}

func (c *ErasureCoordinator) ResumeUnresolved(ctx context.Context) error {
	operations, err := c.store.ListUnresolved(ctx)
	if err != nil {
		return err
	}
	for _, operation := range operations {
		if _, err := c.Erase(ctx, operation.Request); err != nil && !errors.Is(err, ErrErasureHeld) && !errors.Is(err, ErrBackupFence) {
			return err
		}
	}
	return nil
}

func (c *ErasureCoordinator) resume(ctx context.Context, op ErasureOperation) (ErasureOperation, error) {
	if op.State == ErasureResolved {
		return op, nil
	}
	if op.State == ErasureHeld || op.State == ErasureUnresolved {
		op.State = op.ResumeState
		op.ResumeState = ""
		if err := c.persist(ctx, &op, "resume"); err != nil {
			return op, err
		}
	}
	decision, err := c.consent.ReadConsentDecision(ctx, op.Request.Target)
	if err != nil {
		return c.unresolved(ctx, op, err)
	}
	if decision.DecisionDigest != op.Request.ConsentDecisionDigest || decision.PolicyDigest != op.Request.ConsentPolicyDigest || !decision.ErasureAuthorized {
		return c.unresolved(ctx, op, ErrConsentErasure)
	}
	if op.State == ErasureIntent {
		if held, err := c.holds.HasActiveHold(ctx, op.Request.Target, op.Request.AuthorizationDigest); err != nil {
			return c.unresolved(ctx, op, err)
		} else if held {
			return c.held(ctx, op, ErasureIntent)
		}
		if err := c.persist(ctx, &op, "storage_delete_started"); err != nil {
			return op, err
		}
		receipt, err := c.storage.Delete(ctx, StorageDeletionRequest{ErasureClaims: stageClaims(op.Request, op.Request.StorageOperationID)})
		if err != nil {
			return c.unresolved(ctx, op, err)
		}
		op.StorageReceipt = &receipt
		op.State = ErasureStorageDeleted
		if err := c.persist(ctx, &op, "storage_receipt_persisted"); err != nil {
			return op, err
		}
	}
	if op.State == ErasureStorageDeleted {
		status, err := c.backups.CheckBackupFence(ctx, op.Request.ErasureClaims)
		if err != nil {
			return c.unresolved(ctx, op, err)
		}
		if status.GenerationDigest != op.Request.BackupGenerationDigest || status.ExpiryHeight != op.Request.BackupExpiryHeight || status.ExpiryUnix != op.Request.BackupExpiryUnix || !status.AllExpiredOrDeleted {
			return c.unresolvedWith(ctx, op, ErasureStorageDeleted, ErrBackupFence)
		}
		op.State = ErasureBackupFenced
		if err := c.persist(ctx, &op, "backup_fence_satisfied"); err != nil {
			return op, err
		}
	}
	var releaseFinalization func()
	if op.State == ErasureBackupFenced {
		var err error
		releaseFinalization, op, err = c.acquireFinalization(ctx, op)
		if err != nil {
			return op, err
		}
		defer releaseFinalization()
		if err := c.persist(ctx, &op, "kms_destroy_started"); err != nil {
			return op, err
		}
		receipt, err := c.kms.Destroy(ctx, KMSDestructionRequest{ErasureClaims: stageClaims(op.Request, op.Request.KMSOperationID)})
		if err != nil {
			return c.unresolved(ctx, op, err)
		}
		op.KMSReceipt = &receipt
		op.State = ErasureKMSDestroyed
		if err := c.persist(ctx, &op, "kms_receipt_persisted"); err != nil {
			return op, err
		}
	}
	if op.State != ErasureKMSDestroyed || op.StorageReceipt == nil || op.KMSReceipt == nil {
		return c.unresolved(ctx, op, errors.New("erasure operation has incomplete receipts"))
	}
	if releaseFinalization == nil {
		var err error
		releaseFinalization, op, err = c.acquireFinalization(ctx, op)
		if err != nil {
			return op, err
		}
		defer releaseFinalization()
	}
	var transaction ErasureResolutionTransaction
	resolution := contracts.DeletionResolutionContext{
		AuthorizationDigest: op.Request.AuthorizationDigest, PolicyDigest: op.Request.PolicyDigest,
		ProfileDigest: op.Request.ProfileDigest, RequestDigest: op.Request.RequestDigest,
		ConsentDecisionDigest: op.Request.ConsentDecisionDigest, ConsentPolicyDigest: op.Request.ConsentPolicyDigest,
		ObjectCommitment: op.Request.ObjectCommitment, StorageCommitment: op.Request.StorageCommitment,
		KeyCommitment: op.Request.KeyCommitment, BackupGenerationDigest: op.Request.BackupGenerationDigest,
		BackupExpiryHeight: op.Request.BackupExpiryHeight, BackupExpiryUnix: op.Request.BackupExpiryUnix,
		ErasureEpoch: op.Request.ErasureEpoch, CurrentHeight: op.Request.CurrentHeight, CurrentUnix: op.Request.CurrentUnix,
		ReplayConsumer: operationReplayConsumer{ctx: ctx, store: c.store, operationID: op.Request.OperationID, requestDigest: op.RequestDigest, transaction: &transaction},
		ApplyResolved: func(ref contracts.EvidenceObjectRef) error {
			if transaction == nil {
				return errors.New("erasure resolution transaction is unavailable")
			}
			return transaction.ApplyResolved(ref)
		},
	}
	resolved, err := contracts.ResolveDeletion(op.Request.Target, resolution, []contracts.DeletionReceipt{*op.StorageReceipt, *op.KMSReceipt}, c.keys)
	if err != nil {
		return c.unresolved(ctx, op, err)
	}
	op, err = c.store.Load(ctx, op.Request.OperationID)
	if err == nil && op.Resolved == nil {
		op.Resolved = &resolved
	}
	return op, err
}

func (c *ErasureCoordinator) acquireFinalization(ctx context.Context, op ErasureOperation) (func(), ErasureOperation, error) {
	release, err := c.finalize.AcquireErasureFinalization(ctx, op.Request.ErasureClaims)
	if err != nil {
		switch {
		case errors.Is(err, ErrErasureHeld):
			held, holdErr := c.held(ctx, op, op.State)
			return nil, held, holdErr
		case errors.Is(err, ErrBackupFence):
			unresolved, fenceErr := c.unresolvedWith(ctx, op, op.State, err)
			return nil, unresolved, fenceErr
		default:
			unresolved, finalizationErr := c.unresolved(ctx, op, err)
			return nil, unresolved, finalizationErr
		}
	}
	if release == nil {
		unresolved, finalizationErr := c.unresolved(ctx, op, errors.New("erasure finalization fence returned no release function"))
		return nil, unresolved, finalizationErr
	}
	return release, op, nil
}

func (c *ErasureCoordinator) persist(ctx context.Context, op *ErasureOperation, event string) error {
	op.Revision++
	op.Journal = append(op.Journal, ErasureJournalEntry{Sequence: uint64(len(op.Journal) + 1), State: op.State, Event: event})
	return c.store.Update(ctx, *op)
}

func (c *ErasureCoordinator) held(ctx context.Context, op ErasureOperation, resume ErasureOperationState) (ErasureOperation, error) {
	op.State, op.ResumeState = ErasureHeld, resume
	return op, joinPersist(c.persist(ctx, &op, "legal_hold_observed"), ErrErasureHeld)
}

func (c *ErasureCoordinator) unresolved(ctx context.Context, op ErasureOperation, cause error) (ErasureOperation, error) {
	return c.unresolvedWith(ctx, op, op.State, cause)
}

func (c *ErasureCoordinator) unresolvedWith(ctx context.Context, op ErasureOperation, resume ErasureOperationState, cause error) (ErasureOperation, error) {
	op.State, op.ResumeState = ErasureUnresolved, resume
	return op, joinPersist(c.persist(ctx, &op, "operation_unresolved"), cause)
}

func joinPersist(persistErr, cause error) error {
	if persistErr != nil {
		return errors.Join(cause, persistErr)
	}
	return cause
}

func stageClaims(request ErasureRequest, operationID string) ErasureClaims {
	claims := request.ErasureClaims
	claims.OperationID = operationID
	return claims
}

func prepareErasureRequest(request *ErasureRequest) error {
	if request.Version != contracts.EvidenceObjectRefVersion || request.OperationID == "" || request.StorageOperationID == "" || request.KMSOperationID == "" || request.StorageOperationID == request.KMSOperationID {
		return errors.New("complete, distinct erasure operation identifiers are required")
	}
	if err := request.Target.Validate(); err != nil {
		return err
	}
	if request.Target.State != contracts.RetentionDeletionPending {
		return errors.New("evidence object is not pending deletion")
	}
	if request.ObjectCommitment != request.Target.ObjectCommitment || request.PolicyDigest != request.Target.PolicyDigest || request.ProfileDigest != request.Target.ProfileDigest || request.KeyEpoch != request.Target.KeyEpoch {
		return errors.New("erasure request does not match target")
	}
	if request.ErasureEpoch == 0 || request.CurrentHeight <= 0 || request.CurrentUnix <= 0 || request.BackupExpiryHeight <= 0 || request.BackupExpiryUnix <= 0 {
		return errors.New("erasure and backup fence coordinates are required")
	}
	for name, value := range map[string]string{
		"authorization": request.AuthorizationDigest, "policy": request.PolicyDigest, "profile": request.ProfileDigest,
		"object": request.ObjectCommitment, "storage": request.StorageCommitment, "key": request.KeyCommitment,
		"backup generation": request.BackupGenerationDigest, "consent decision": request.ConsentDecisionDigest,
		"consent policy": request.ConsentPolicyDigest,
	} {
		if !isSHA256(value) {
			return fmt.Errorf("%s digest must be lowercase SHA-256 hex", name)
		}
	}
	expected, err := erasureRequestDigest(*request)
	if err != nil {
		return err
	}
	if request.RequestDigest == "" {
		request.RequestDigest = expected
	} else if request.RequestDigest != expected {
		return errors.New("immutable erasure request digest mismatch")
	}
	return nil
}

func erasureRequestDigest(request ErasureRequest) (string, error) {
	request.RequestDigest = ""
	encoded, err := json.Marshal(request)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:]), nil
}

func isSHA256(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == sha256.Size && value == hex.EncodeToString(decoded)
}

type operationReplayConsumer struct {
	ctx           context.Context
	store         ErasureOperationStore
	operationID   string
	requestDigest string
	transaction   *ErasureResolutionTransaction
}

func (c operationReplayConsumer) ConsumeDeletionReceipts(storage, kms contracts.DeletionReceiptReplay, apply func() error) error {
	return c.store.Resolve(c.ctx, c.operationID, c.requestDigest, storage, kms, func(transaction ErasureResolutionTransaction, operation *ErasureOperation) error {
		if c.transaction == nil {
			return errors.New("erasure resolution transaction target is required")
		}
		*c.transaction = transaction
		defer func() { *c.transaction = nil }()
		if err := apply(); err != nil {
			return err
		}
		resolved := operation.Request.Target
		resolved.State = contracts.RetentionDeleted
		resolved.UpdatedHeight = max64(operation.StorageReceipt.CompletedHeight, operation.KMSReceipt.CompletedHeight)
		resolved.UpdatedUnix = max64(operation.StorageReceipt.CompletedUnix, operation.KMSReceipt.CompletedUnix)
		operation.State = ErasureResolved
		operation.Resolved = &resolved
		operation.Revision++
		operation.Journal = append(operation.Journal, ErasureJournalEntry{Sequence: uint64(len(operation.Journal) + 1), State: ErasureResolved, Event: "resolution_and_replay_committed"})
		return nil
	})
}

func max64(left, right int64) int64 {
	if left > right {
		return left
	}
	return right
}

// FixtureEd25519ReceiptSigner is deterministic test infrastructure, not a production signer.
type FixtureEd25519ReceiptSigner struct {
	Kind       contracts.DeletionReceiptKind
	KeyID      string
	KeyEpoch   uint64
	PrivateKey ed25519.PrivateKey
}

func NewFixtureEd25519ReceiptSigner(kind contracts.DeletionReceiptKind, keyID string, seed []byte) (*FixtureEd25519ReceiptSigner, error) {
	if len(seed) != ed25519.SeedSize || keyID == "" {
		return nil, errors.New("fixture signer requires a key id and 32-byte seed")
	}
	return &FixtureEd25519ReceiptSigner{Kind: kind, KeyID: keyID, KeyEpoch: 1, PrivateKey: ed25519.NewKeyFromSeed(seed)}, nil
}

func (s *FixtureEd25519ReceiptSigner) SignDeletionReceipt(_ context.Context, receipt contracts.DeletionReceipt) (contracts.DeletionReceipt, error) {
	if s == nil || receipt.Kind != s.Kind {
		return contracts.DeletionReceipt{}, errors.New("fixture signer authority kind mismatch")
	}
	receipt.SignerKeyID, receipt.SignerKeyEpoch = s.KeyID, s.KeyEpoch
	bytes, err := receipt.CanonicalSignBytes()
	if err != nil {
		return contracts.DeletionReceipt{}, err
	}
	receipt.Signature = ed25519.Sign(s.PrivateKey, bytes)
	return receipt, nil
}

func (s *FixtureEd25519ReceiptSigner) PublicKey() ed25519.PublicKey {
	return append(ed25519.PublicKey(nil), s.PrivateKey.Public().(ed25519.PublicKey)...)
}

type FixtureDeletionAdapter struct {
	Kind     contracts.DeletionReceiptKind
	Signer   DeletionReceiptSigner
	mu       sync.Mutex
	receipts map[string]contracts.DeletionReceipt
	requests map[string]string
}

func NewFixtureDeletionAdapter(kind contracts.DeletionReceiptKind, signer DeletionReceiptSigner) *FixtureDeletionAdapter {
	return &FixtureDeletionAdapter{Kind: kind, Signer: signer, receipts: make(map[string]contracts.DeletionReceipt), requests: make(map[string]string)}
}

func (a *FixtureDeletionAdapter) Delete(ctx context.Context, request StorageDeletionRequest) (contracts.DeletionReceipt, error) {
	return a.execute(ctx, request.ErasureClaims, contracts.ReceiptStorageDeletion)
}

func (a *FixtureDeletionAdapter) Destroy(ctx context.Context, request KMSDestructionRequest) (contracts.DeletionReceipt, error) {
	return a.execute(ctx, request.ErasureClaims, contracts.ReceiptKMSDestruction)
}

func (a *FixtureDeletionAdapter) execute(ctx context.Context, claims ErasureClaims, kind contracts.DeletionReceiptKind) (contracts.DeletionReceipt, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.Kind != kind || a.Signer == nil {
		return contracts.DeletionReceipt{}, errors.New("fixture deletion adapter authority mismatch")
	}
	digest, _ := json.Marshal(claims)
	requestHash := sha256.Sum256(digest)
	requestDigest := hex.EncodeToString(requestHash[:])
	if previous, found := a.requests[claims.OperationID]; found {
		if previous != requestDigest {
			return contracts.DeletionReceipt{}, ErrErasureConflict
		}
		return a.receipts[claims.OperationID], nil
	}
	receipt := receiptFromClaims(kind, claims)
	signed, err := a.Signer.SignDeletionReceipt(ctx, receipt)
	if err != nil {
		return contracts.DeletionReceipt{}, err
	}
	a.requests[claims.OperationID], a.receipts[claims.OperationID] = requestDigest, signed
	return signed, nil
}

func receiptFromClaims(kind contracts.DeletionReceiptKind, claims ErasureClaims) contracts.DeletionReceipt {
	return contracts.DeletionReceipt{
		Version: claims.Version, Kind: kind, TargetCommitment: claims.Target.ObjectCommitment,
		RequestDigest: claims.RequestDigest, AuthorizationDigest: claims.AuthorizationDigest,
		PolicyDigest: claims.PolicyDigest, ProfileDigest: claims.ProfileDigest,
		ConsentDecisionDigest: claims.ConsentDecisionDigest, ConsentPolicyDigest: claims.ConsentPolicyDigest,
		OperationID: claims.OperationID, KeyEpoch: claims.KeyEpoch, ObjectCommitment: claims.ObjectCommitment,
		StorageCommitment: claims.StorageCommitment, KeyCommitment: claims.KeyCommitment,
		BackupGenerationDigest: claims.BackupGenerationDigest, BackupExpiryHeight: claims.BackupExpiryHeight,
		BackupExpiryUnix: claims.BackupExpiryUnix, ErasureEpoch: claims.ErasureEpoch,
		CompletedHeight: claims.CurrentHeight, CompletedUnix: claims.CurrentUnix,
	}
}

// MemoryErasureOperationStore is fixture-only. Production requires an external durable transactional journal.
type MemoryErasureOperationStore struct {
	mu         sync.Mutex
	operations map[string]ErasureOperation
	replays    map[string]string
	leaseMu    sync.Mutex
	leases     map[string]*sync.Mutex
	apply      func(contracts.EvidenceObjectRef) error
}

func NewMemoryErasureOperationStore() *MemoryErasureOperationStore {
	return &MemoryErasureOperationStore{operations: make(map[string]ErasureOperation), replays: make(map[string]string), leases: make(map[string]*sync.Mutex)}
}

// SetFixtureResolvedApply configures the fixture transaction callback.
func (s *MemoryErasureOperationStore) SetFixtureResolvedApply(apply func(contracts.EvidenceObjectRef) error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.apply = apply
}

// AcquireOperation is a fixture process-local lease. Production stores must
// provide a crash-expiring distributed lease or equivalent serializable claim.
func (s *MemoryErasureOperationStore) AcquireOperation(ctx context.Context, operationID, requestDigest string) (func(), error) {
	if operationID == "" || requestDigest == "" {
		return nil, errors.New("operation id and request digest are required for erasure lease")
	}
	s.leaseMu.Lock()
	lease := s.leases[operationID]
	if lease == nil {
		lease = &sync.Mutex{}
		s.leases[operationID] = lease
	}
	s.leaseMu.Unlock()
	acquired := make(chan struct{})
	go func() {
		lease.Lock()
		close(acquired)
	}()
	select {
	case <-ctx.Done():
		go func() {
			<-acquired
			lease.Unlock()
		}()
		return nil, ctx.Err()
	case <-acquired:
		return lease.Unlock, nil
	}
}

func (s *MemoryErasureOperationStore) Begin(_ context.Context, request ErasureRequest) (ErasureOperation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing, found := s.operations[request.OperationID]; found {
		if existing.RequestDigest != request.RequestDigest {
			return ErasureOperation{}, ErrErasureConflict
		}
		return cloneErasureOperation(existing), nil
	}
	op := ErasureOperation{Request: request, RequestDigest: request.RequestDigest, State: ErasureIntent, Revision: 1,
		Journal: []ErasureJournalEntry{{Sequence: 1, State: ErasureIntent, Event: "intent_persisted"}}}
	s.operations[request.OperationID] = cloneErasureOperation(op)
	return op, nil
}

func (s *MemoryErasureOperationStore) Load(_ context.Context, operationID string) (ErasureOperation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	op, found := s.operations[operationID]
	if !found {
		return ErasureOperation{}, errors.New("erasure operation not found")
	}
	return cloneErasureOperation(op), nil
}

func (s *MemoryErasureOperationStore) Update(_ context.Context, operation ErasureOperation) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, found := s.operations[operation.Request.OperationID]
	if !found || current.RequestDigest != operation.RequestDigest {
		return ErrErasureConflict
	}
	if operation.Revision != current.Revision+1 {
		return errors.New("stale erasure operation revision")
	}
	s.operations[operation.Request.OperationID] = cloneErasureOperation(operation)
	return nil
}

func (s *MemoryErasureOperationStore) ListUnresolved(_ context.Context) ([]ErasureOperation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]ErasureOperation, 0)
	for _, operation := range s.operations {
		if operation.State != ErasureResolved {
			result = append(result, cloneErasureOperation(operation))
		}
	}
	return result, nil
}

func (s *MemoryErasureOperationStore) Resolve(_ context.Context, operationID, requestDigest string, storage, kms contracts.DeletionReceiptReplay, apply func(ErasureResolutionTransaction, *ErasureOperation) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	operation, found := s.operations[operationID]
	if !found || operation.RequestDigest != requestDigest {
		return ErrErasureConflict
	}
	replays := []contracts.DeletionReceiptReplay{storage, kms}
	allExact := operation.State == ErasureResolved
	anyExisting := false
	for _, replay := range replays {
		if digest, exists := s.replays[replay.OperationID]; exists {
			anyExisting = true
			if digest != replay.ReceiptDigest {
				return errors.New("deletion receipt operation replayed")
			}
			continue
		}
		allExact = false
		for _, digest := range s.replays {
			if digest == replay.ReceiptDigest {
				return errors.New("deletion receipt digest replayed")
			}
		}
	}
	if allExact {
		return nil
	}
	if anyExisting {
		return errors.New("partial deletion receipt replay state is inconsistent")
	}
	if operation.State == ErasureResolved {
		return errors.New("resolved deletion receipt pair does not match persisted replay state")
	}
	if s.apply == nil {
		return errors.New("fixture resolved-state transaction callback is not configured")
	}
	candidate := cloneErasureOperation(operation)
	transaction := memoryErasureResolutionTransaction{apply: s.apply}
	if err := apply(transaction, &candidate); err != nil {
		return err
	}
	s.operations[operationID] = candidate
	s.replays[storage.OperationID], s.replays[kms.OperationID] = storage.ReceiptDigest, kms.ReceiptDigest
	return nil
}

type memoryErasureResolutionTransaction struct {
	apply func(contracts.EvidenceObjectRef) error
}

func (t memoryErasureResolutionTransaction) ApplyResolved(ref contracts.EvidenceObjectRef) error {
	return t.apply(ref)
}

func cloneErasureOperation(operation ErasureOperation) ErasureOperation {
	encoded, _ := json.Marshal(operation)
	var clone ErasureOperation
	_ = json.Unmarshal(encoded, &clone)
	return clone
}
