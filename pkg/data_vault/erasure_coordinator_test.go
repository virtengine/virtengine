package data_vault

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/virtengine/virtengine/pkg/data_vault/contracts"
)

type fixtureHoldReader struct {
	mu      sync.Mutex
	results []bool
	calls   int
}

func (r *fixtureHoldReader) HasActiveHold(context.Context, contracts.EvidenceObjectRef, string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := false
	if r.calls < len(r.results) {
		result = r.results[r.calls]
	}
	r.calls++
	return result, nil
}

type fixtureConsentReader struct{ decision ConsentDecision }

func (r fixtureConsentReader) ReadConsentDecision(context.Context, contracts.EvidenceObjectRef) (ConsentDecision, error) {
	return r.decision, nil
}

type fixtureBackupFence struct{ status BackupFenceStatus }

func (f fixtureBackupFence) CheckBackupFence(context.Context, ErasureClaims) (BackupFenceStatus, error) {
	return f.status, nil
}

type fixtureFinalizationFence struct {
	holds   HoldReader
	consent ConsentDecisionReader
	backups BackupFence
}

type fixtureRestoreAuthority struct{ manifestDigest string }

func (a fixtureRestoreAuthority) VerifyRestoreInventory(_ context.Context, fence ErasureFence, inventory TombstoneInventory) error {
	if inventory.SnapshotManifestDigest != a.manifestDigest || inventory.BackupGenerationDigest != fence.BackupGenerationDigest {
		return errors.New("fixture restore manifest is not trusted")
	}
	return nil
}

func (f fixtureFinalizationFence) AcquireErasureFinalization(ctx context.Context, claims ErasureClaims) (func(), error) {
	if held, err := f.holds.HasActiveHold(ctx, claims.Target, claims.AuthorizationDigest); err != nil {
		return nil, err
	} else if held {
		return nil, ErrErasureHeld
	}
	decision, err := f.consent.ReadConsentDecision(ctx, claims.Target)
	if err != nil {
		return nil, err
	}
	if decision.DecisionDigest != claims.ConsentDecisionDigest || decision.PolicyDigest != claims.ConsentPolicyDigest || !decision.ErasureAuthorized {
		return nil, ErrConsentErasure
	}
	status, err := f.backups.CheckBackupFence(ctx, claims)
	if err != nil {
		return nil, err
	}
	if status.GenerationDigest != claims.BackupGenerationDigest || status.ExpiryHeight != claims.BackupExpiryHeight || status.ExpiryUnix != claims.BackupExpiryUnix || !status.AllExpiredOrDeleted {
		return nil, ErrBackupFence
	}
	return func() {}, nil
}

type fixtureReceiptResolver map[string]ed25519.PublicKey

func (r fixtureReceiptResolver) ResolveDeletionReceiptKey(kind contracts.DeletionReceiptKind, keyID string, epoch uint64) (ed25519.PublicKey, error) {
	key := r[string(kind)+"/"+keyID]
	if len(key) == 0 || epoch != 1 {
		return nil, errors.New("fixture receipt key not found")
	}
	return key, nil
}

func (r fixtureReceiptResolver) ResolveDeletionReceiptAuthority(kind contracts.DeletionReceiptKind, keyID string, epoch uint64) (string, error) {
	if _, err := r.ResolveDeletionReceiptKey(kind, keyID, epoch); err != nil {
		return "", err
	}
	return string(kind) + "-authority", nil
}

type faultStore struct {
	*MemoryErasureOperationStore
	mu          sync.Mutex
	failEvent   string
	failBegin   bool
	failResolve bool
}

func (s *faultStore) Begin(ctx context.Context, request ErasureRequest) (ErasureOperation, error) {
	s.mu.Lock()
	if s.failBegin {
		s.failBegin = false
		s.mu.Unlock()
		return ErasureOperation{}, errors.New("injected intent persistence failure")
	}
	s.mu.Unlock()
	return s.MemoryErasureOperationStore.Begin(ctx, request)
}

func (s *faultStore) Update(ctx context.Context, operation ErasureOperation) error {
	s.mu.Lock()
	event := operation.Journal[len(operation.Journal)-1].Event
	if s.failEvent == event {
		s.failEvent = ""
		s.mu.Unlock()
		return errors.New("injected operation persistence failure")
	}
	s.mu.Unlock()
	return s.MemoryErasureOperationStore.Update(ctx, operation)
}

func (s *faultStore) Resolve(ctx context.Context, operationID, requestDigest string, storage, kms contracts.DeletionReceiptReplay, apply func(ErasureResolutionTransaction, *ErasureOperation) error) error {
	s.mu.Lock()
	if s.failResolve {
		s.failResolve = false
		s.mu.Unlock()
		return errors.New("injected replay persistence failure")
	}
	s.mu.Unlock()
	return s.MemoryErasureOperationStore.Resolve(ctx, operationID, requestDigest, storage, kms, apply)
}

func TestErasureCoordinatorExactRetryAndConcurrentDuplicate(t *testing.T) {
	request, dependencies := newErasureFixture(t)
	coordinator := dependencies.coordinator(t, dependencies.store, &fixtureHoldReader{}, dependencies.backup, func(context.Context, contracts.EvidenceObjectRef) error { return nil })
	secondCoordinator := dependencies.coordinator(t, dependencies.store, &fixtureHoldReader{}, dependencies.backup, func(context.Context, contracts.EvidenceObjectRef) error { return nil })

	const workers = 12
	var wait sync.WaitGroup
	wait.Add(workers)
	results := make(chan ErasureOperation, workers)
	errorsFound := make(chan error, workers)
	for worker := range workers {
		go func() {
			defer wait.Done()
			active := coordinator
			if worker%2 != 0 {
				active = secondCoordinator
			}
			operation, err := active.Erase(context.Background(), request)
			results <- operation
			errorsFound <- err
		}()
	}
	wait.Wait()
	close(results)
	close(errorsFound)
	for err := range errorsFound {
		if err != nil {
			t.Fatal(err)
		}
	}
	for operation := range results {
		if operation.State != ErasureResolved {
			t.Fatalf("duplicate ended in %s", operation.State)
		}
	}
	if len(dependencies.storage.receipts) != 1 || len(dependencies.kms.receipts) != 1 {
		t.Fatal("concurrent duplicate performed more than one exact adapter operation")
	}
	first, _ := json.Marshal(dependencies.storage.receipts[request.StorageOperationID])
	operation, err := coordinator.Erase(context.Background(), request)
	if err != nil || operation.State != ErasureResolved {
		t.Fatal("terminal exact retry did not return the same result")
	}
	second, _ := json.Marshal(dependencies.storage.receipts[request.StorageOperationID])
	if !bytes.Equal(first, second) {
		t.Fatal("exact retry changed receipt bytes")
	}
	changed := request
	changed.CurrentUnix++
	changed.RequestDigest = ""
	if _, err := coordinator.Erase(context.Background(), changed); !errors.Is(err, ErrErasureConflict) {
		t.Fatalf("changed request reused operation: %v", err)
	}
	storageReplay := contracts.DeletionReceiptReplay{OperationID: request.StorageOperationID, ReceiptDigest: dependencies.store.replays[request.StorageOperationID]}
	mixedKMSReplay := contracts.DeletionReceiptReplay{OperationID: request.KMSOperationID, ReceiptDigest: fixtureDigest("different receipt")}
	if err := dependencies.store.Resolve(context.Background(), request.OperationID, request.RequestDigest, storageReplay, mixedKMSReplay, func(ErasureResolutionTransaction, *ErasureOperation) error { return nil }); err == nil {
		t.Fatal("terminal retry accepted only one matching replay receipt")
	}
}

func TestErasureCoordinatorResumesAfterIrreversibleWork(t *testing.T) {
	for _, event := range []string{"storage_receipt_persisted", "backup_fence_satisfied", "kms_receipt_persisted"} {
		t.Run(event, func(t *testing.T) {
			request, dependencies := newErasureFixture(t)
			store := &faultStore{MemoryErasureOperationStore: NewMemoryErasureOperationStore(), failEvent: event}
			coordinator := dependencies.coordinator(t, store, &fixtureHoldReader{}, dependencies.backup, func(context.Context, contracts.EvidenceObjectRef) error { return nil })
			if _, err := coordinator.Erase(context.Background(), request); err == nil {
				t.Fatal("injected crash was not observed")
			}
			storageReceipt, hadStorageReceipt := dependencies.storage.receipts[request.StorageOperationID]
			kmsReceipt, hadKMSReceipt := dependencies.kms.receipts[request.KMSOperationID]
			storageBefore, _ := json.Marshal(storageReceipt)
			kmsBefore, _ := json.Marshal(kmsReceipt)
			operation, err := coordinator.Erase(context.Background(), request)
			if err != nil || operation.State != ErasureResolved {
				t.Fatalf("resume failed: state=%s err=%v", operation.State, err)
			}
			storageAfter, _ := json.Marshal(dependencies.storage.receipts[request.StorageOperationID])
			kmsAfter, _ := json.Marshal(dependencies.kms.receipts[request.KMSOperationID])
			if (hadStorageReceipt && !bytes.Equal(storageBefore, storageAfter)) || (hadKMSReceipt && !bytes.Equal(kmsBefore, kmsAfter)) {
				t.Fatal("resume changed a persisted adapter receipt")
			}
		})
	}
}

func TestErasureCoordinatorIntentAndResolutionRollback(t *testing.T) {
	request, dependencies := newErasureFixture(t)
	store := &faultStore{MemoryErasureOperationStore: NewMemoryErasureOperationStore(), failBegin: true, failResolve: true}
	callbackCalls := 0
	coordinator := dependencies.coordinator(t, store, &fixtureHoldReader{}, dependencies.backup, func(context.Context, contracts.EvidenceObjectRef) error {
		callbackCalls++
		return nil
	})
	if _, err := coordinator.Erase(context.Background(), request); err == nil {
		t.Fatal("intent persistence failure was ignored")
	}
	if _, err := store.Load(context.Background(), request.OperationID); err == nil {
		t.Fatal("failed intent became durable")
	}
	if _, err := coordinator.Erase(context.Background(), request); err == nil {
		t.Fatal("replay-store failure was ignored")
	}
	if callbackCalls != 0 {
		t.Fatal("callback ran outside replay transaction")
	}
	operation, err := coordinator.Erase(context.Background(), request)
	if err != nil || operation.State != ErasureResolved || callbackCalls != 1 {
		t.Fatalf("resolution retry failed: state=%s calls=%d err=%v", operation.State, callbackCalls, err)
	}

	request, dependencies = newErasureFixture(t)
	callbackStore := NewMemoryErasureOperationStore()
	callbackCoordinator := dependencies.coordinator(t, callbackStore, &fixtureHoldReader{}, dependencies.backup, func(context.Context, contracts.EvidenceObjectRef) error {
		return errors.New("injected resolved-state callback failure")
	})
	if _, err := callbackCoordinator.Erase(context.Background(), request); err == nil {
		t.Fatal("callback failure was ignored")
	}
	if len(callbackStore.replays) != 0 {
		t.Fatal("callback failure consumed replay receipts")
	}
}

func TestErasureCoordinatorHoldBackupAndConsentGates(t *testing.T) {
	t.Run("hold before storage", func(t *testing.T) {
		request, dependencies := newErasureFixture(t)
		coordinator := dependencies.coordinator(t, dependencies.store, &fixtureHoldReader{results: []bool{true}}, dependencies.backup, func(context.Context, contracts.EvidenceObjectRef) error { return nil })
		operation, err := coordinator.Erase(context.Background(), request)
		if !errors.Is(err, ErrErasureHeld) || operation.State != ErasureHeld || len(dependencies.storage.receipts) != 0 {
			t.Fatal("pre-storage legal hold did not stop deletion")
		}
	})
	t.Run("hold between storage and kms", func(t *testing.T) {
		request, dependencies := newErasureFixture(t)
		coordinator := dependencies.coordinator(t, dependencies.store, &fixtureHoldReader{results: []bool{false, true}}, dependencies.backup, func(context.Context, contracts.EvidenceObjectRef) error { return nil })
		operation, err := coordinator.Erase(context.Background(), request)
		if !errors.Is(err, ErrErasureHeld) || operation.State != ErasureHeld || len(dependencies.storage.receipts) != 1 || len(dependencies.kms.receipts) != 0 {
			t.Fatal("between-stage legal hold did not leave resumable storage proof")
		}
	})
	t.Run("finalization hold before kms", func(t *testing.T) {
		request, dependencies := newErasureFixture(t)
		coordinator := dependencies.coordinator(t, dependencies.store, &fixtureHoldReader{results: []bool{false, true}}, dependencies.backup, func(context.Context, contracts.EvidenceObjectRef) error { return nil })
		operation, err := coordinator.Erase(context.Background(), request)
		if !errors.Is(err, ErrErasureHeld) || operation.State != ErasureHeld || len(dependencies.kms.receipts) != 0 {
			t.Fatal("finalization legal hold did not stop KMS destruction")
		}
	})
	t.Run("backup unexpired", func(t *testing.T) {
		request, dependencies := newErasureFixture(t)
		backup := dependencies.backup
		backup.status.AllExpiredOrDeleted = false
		coordinator := dependencies.coordinator(t, dependencies.store, &fixtureHoldReader{}, backup, func(context.Context, contracts.EvidenceObjectRef) error { return nil })
		operation, err := coordinator.Erase(context.Background(), request)
		if !errors.Is(err, ErrBackupFence) || operation.State != ErasureUnresolved || len(dependencies.kms.receipts) != 0 {
			t.Fatal("unexpired backup did not block KMS destruction")
		}
	})
	t.Run("withdrawn restore but authorized erasure", func(t *testing.T) {
		request, dependencies := newErasureFixture(t)
		if err := ValidateConsentAccess(dependencies.consent.decision, true); err == nil {
			t.Fatal("withdrawn consent allowed restore/plaintext access")
		}
		coordinator := dependencies.coordinator(t, dependencies.store, &fixtureHoldReader{}, dependencies.backup, func(context.Context, contracts.EvidenceObjectRef) error { return nil })
		if operation, err := coordinator.Erase(context.Background(), request); err != nil || operation.State != ErasureResolved {
			t.Fatal("policy-authorized erasure was blocked by restore withdrawal")
		}
	})
	t.Run("unauthorized erasure", func(t *testing.T) {
		request, dependencies := newErasureFixture(t)
		dependencies.consent.decision.ErasureAuthorized = false
		coordinator := dependencies.coordinator(t, dependencies.store, &fixtureHoldReader{}, dependencies.backup, func(context.Context, contracts.EvidenceObjectRef) error { return nil })
		if operation, err := coordinator.Erase(context.Background(), request); !errors.Is(err, ErrConsentErasure) || operation.State != ErasureUnresolved {
			t.Fatal("consent policy allowed unauthorized erasure")
		}
	})
}

func TestDeletionReceiptRejectsWrongErasureFenceClaims(t *testing.T) {
	request, dependencies := newErasureFixture(t)
	storage, err := dependencies.storage.Delete(context.Background(), StorageDeletionRequest{ErasureClaims: stageClaims(request, request.StorageOperationID)})
	if err != nil {
		t.Fatal(err)
	}
	kms, err := dependencies.kms.Destroy(context.Background(), KMSDestructionRequest{ErasureClaims: stageClaims(request, request.KMSOperationID)})
	if err != nil {
		t.Fatal(err)
	}
	mutations := map[string]func(*contracts.DeletionReceipt){
		"request digest":    func(receipt *contracts.DeletionReceipt) { receipt.RequestDigest = fixtureDigest("wrong request") },
		"backup generation": func(receipt *contracts.DeletionReceipt) { receipt.BackupGenerationDigest = fixtureDigest("wrong backup") },
		"erasure epoch":     func(receipt *contracts.DeletionReceipt) { receipt.ErasureEpoch++ },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			bad := storage
			mutate(&bad)
			bad, err = dependencies.storage.Signer.SignDeletionReceipt(context.Background(), bad)
			if err != nil {
				t.Fatal(err)
			}
			resolution := fixtureResolution(request, NewMemoryErasureOperationStore())
			resolution.ApplyResolved = func(contracts.EvidenceObjectRef) error { return nil }
			store := NewMemoryErasureOperationStore()
			store.SetFixtureResolvedApply(func(contracts.EvidenceObjectRef) error { return nil })
			_, _ = store.Begin(context.Background(), request)
			var transaction ErasureResolutionTransaction
			resolution.ReplayConsumer = operationReplayConsumer{ctx: context.Background(), store: store, operationID: request.OperationID, requestDigest: request.RequestDigest, transaction: &transaction}
			if result, err := contracts.ResolveDeletion(request.Target, resolution, []contracts.DeletionReceipt{bad, kms}, dependencies.resolver); err == nil || result.State != contracts.RetentionDeletionUnresolved {
				t.Fatalf("wrong %s resolved deletion", name)
			}
		})
	}
	resolution := fixtureResolution(request, NewMemoryErasureOperationStore())
	resolution.ObjectCommitment = fixtureDigest("wrong resolution object")
	if result, err := contracts.ResolveDeletion(request.Target, resolution, []contracts.DeletionReceipt{storage, kms}, dependencies.resolver); err == nil || result.State != contracts.RetentionDeletionUnresolved {
		t.Fatal("resolution context with mismatched target claims resolved deletion")
	}
}

func TestValidateRestoreAgainstErasure(t *testing.T) {
	manifest := fixtureDigest("snapshot manifest")
	fence := ErasureFence{TargetCommitment: fixtureDigest("object"), BackupGenerationDigest: fixtureDigest("backup generation"), ErasureEpoch: 9, ErasedHeight: 20, ErasedUnix: 200}
	authority := fixtureRestoreAuthority{manifestDigest: manifest}
	preErasure := TombstoneInventory{TargetCommitment: fence.TargetCommitment, BackupGenerationDigest: fence.BackupGenerationDigest, SnapshotManifestDigest: manifest, SnapshotHeight: 19, SnapshotUnix: 199, ErasureEpoch: 8, Tombstone: true, Undecryptable: true}
	if err := ValidateRestoreAgainstErasure(context.Background(), authority, fence, preErasure); err == nil {
		t.Fatal("pre-erasure restore was accepted")
	}
	valid := TombstoneInventory{TargetCommitment: fence.TargetCommitment, BackupGenerationDigest: fence.BackupGenerationDigest, SnapshotManifestDigest: manifest, SnapshotHeight: 20, SnapshotUnix: 200, ErasureEpoch: 9, Tombstone: true, AuditRecords: true, Undecryptable: true}
	if err := ValidateRestoreAgainstErasure(context.Background(), authority, fence, valid); err != nil {
		t.Fatal(err)
	}
	forged := valid
	forged.SnapshotHeight = 21
	forged.SnapshotManifestDigest = fixtureDigest("forged manifest")
	if err := ValidateRestoreAgainstErasure(context.Background(), authority, fence, forged); err == nil {
		t.Fatal("unauthenticated restore inventory was accepted")
	}
	valid.WrappedKeys = true
	if err := ValidateRestoreAgainstErasure(context.Background(), authority, fence, valid); err == nil {
		t.Fatal("post-erasure restore recreated wrapped keys")
	}
}

type erasureFixture struct {
	store    *MemoryErasureOperationStore
	storage  *FixtureDeletionAdapter
	kms      *FixtureDeletionAdapter
	consent  fixtureConsentReader
	backup   fixtureBackupFence
	resolver fixtureReceiptResolver
}

func newErasureFixture(t *testing.T) (ErasureRequest, *erasureFixture) {
	t.Helper()
	ref, _, err := contracts.CreateEvidenceObjectRef(rand.Reader, contracts.CommitmentDomainIdentityScope, contracts.EvidenceObjectFields{
		EvidenceDigest: fixtureDigest("evidence"), RetentionPolicyDigest: fixtureDigest("retention"),
		PolicyDigest: fixtureDigest("policy"), ProfileDigest: fixtureDigest("profile"), KeyEpoch: 3,
		CreatedHeight: 10, CreatedUnix: 100, ExpiresHeight: 100, ExpiresUnix: 1000,
	})
	if err != nil {
		t.Fatal(err)
	}
	ref, _ = contracts.TransitionRetention(ref, contracts.RetentionDeletionScheduled, 11, 101)
	ref, _ = contracts.TransitionRetention(ref, contracts.RetentionDeletionPending, 12, 102)
	request := ErasureRequest{ErasureClaims: ErasureClaims{
		Version: contracts.EvidenceObjectRefVersion, OperationID: "erase-1", Target: ref,
		AuthorizationDigest: fixtureDigest("authorization"), PolicyDigest: ref.PolicyDigest, ProfileDigest: ref.ProfileDigest,
		KeyEpoch: ref.KeyEpoch, ObjectCommitment: ref.ObjectCommitment, StorageCommitment: fixtureDigest("storage"),
		KeyCommitment: fixtureDigest("key"), BackupGenerationDigest: fixtureDigest("backup"),
		BackupExpiryHeight: 13, BackupExpiryUnix: 103, ConsentDecisionDigest: fixtureDigest("consent"),
		ConsentPolicyDigest: fixtureDigest("consent-policy"), ErasureEpoch: 5, CurrentHeight: 15, CurrentUnix: 105,
	}, StorageOperationID: "erase-1/storage", KMSOperationID: "erase-1/kms"}
	if err := prepareErasureRequest(&request); err != nil {
		t.Fatal(err)
	}
	storageSigner, _ := NewFixtureEd25519ReceiptSigner(contracts.ReceiptStorageDeletion, "storage", bytes.Repeat([]byte{1}, ed25519.SeedSize))
	kmsSigner, _ := NewFixtureEd25519ReceiptSigner(contracts.ReceiptKMSDestruction, "kms", bytes.Repeat([]byte{2}, ed25519.SeedSize))
	fixture := &erasureFixture{
		store:   NewMemoryErasureOperationStore(),
		storage: NewFixtureDeletionAdapter(contracts.ReceiptStorageDeletion, storageSigner),
		kms:     NewFixtureDeletionAdapter(contracts.ReceiptKMSDestruction, kmsSigner),
		consent: fixtureConsentReader{decision: ConsentDecision{DecisionDigest: request.ConsentDecisionDigest, PolicyDigest: request.ConsentPolicyDigest, RestoreAllowed: false, ErasureAuthorized: true}},
		backup:  fixtureBackupFence{status: BackupFenceStatus{GenerationDigest: request.BackupGenerationDigest, ExpiryHeight: request.BackupExpiryHeight, ExpiryUnix: request.BackupExpiryUnix, AllExpiredOrDeleted: true}},
		resolver: fixtureReceiptResolver{
			string(contracts.ReceiptStorageDeletion) + "/storage": storageSigner.PublicKey(),
			string(contracts.ReceiptKMSDestruction) + "/kms":      kmsSigner.PublicKey(),
		},
	}
	return request, fixture
}

func (f *erasureFixture) coordinator(t *testing.T, store ErasureOperationStore, holds HoldReader, backups BackupFence, apply func(context.Context, contracts.EvidenceObjectRef) error) *ErasureCoordinator {
	t.Helper()
	fixtureStore, ok := store.(interface {
		SetFixtureResolvedApply(func(contracts.EvidenceObjectRef) error)
	})
	if !ok {
		t.Fatal("fixture erasure store does not expose its transaction callback")
	}
	fixtureStore.SetFixtureResolvedApply(func(ref contracts.EvidenceObjectRef) error { return apply(context.Background(), ref) })
	finalize := fixtureFinalizationFence{holds: holds, consent: f.consent, backups: backups}
	coordinator, err := NewErasureCoordinator(store, f.storage, f.kms, holds, f.consent, backups, finalize, f.resolver)
	if err != nil {
		t.Fatal(err)
	}
	return coordinator
}

func fixtureResolution(request ErasureRequest, store ErasureOperationStore) contracts.DeletionResolutionContext {
	return contracts.DeletionResolutionContext{
		AuthorizationDigest: request.AuthorizationDigest, PolicyDigest: request.PolicyDigest, ProfileDigest: request.ProfileDigest,
		RequestDigest: request.RequestDigest, ConsentDecisionDigest: request.ConsentDecisionDigest, ConsentPolicyDigest: request.ConsentPolicyDigest,
		ObjectCommitment: request.ObjectCommitment, StorageCommitment: request.StorageCommitment, KeyCommitment: request.KeyCommitment,
		BackupGenerationDigest: request.BackupGenerationDigest, BackupExpiryHeight: request.BackupExpiryHeight,
		BackupExpiryUnix: request.BackupExpiryUnix, ErasureEpoch: request.ErasureEpoch,
		CurrentHeight: request.CurrentHeight, CurrentUnix: request.CurrentUnix,
	}
}

func fixtureDigest(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}
