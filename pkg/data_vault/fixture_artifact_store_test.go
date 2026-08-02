package data_vault

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/virtengine/virtengine/pkg/artifact_store"
	"github.com/virtengine/virtengine/pkg/data_vault/contracts"
	"github.com/virtengine/virtengine/pkg/data_vault/keys"
)

type fixtureHoldVerifier struct{}

func (fixtureHoldVerifier) VerifyHoldAuthority(contracts.LegalHoldAuthority) error { return nil }

type fixtureErasureVerifier struct {
	authorized map[string]bool
}

func (v *fixtureErasureVerifier) VerifyFixtureErasureAuthorization(_ context.Context, authorization FixtureErasureAuthorization) error {
	if !v.authorized[authorization.Digest] {
		return errors.New("fixture authorization is not approved")
	}
	return nil
}

func (v *fixtureErasureVerifier) allow(authorization FixtureErasureAuthorization) {
	v.authorized[authorization.Digest] = true
}

func fixtureHold(state contracts.HoldState) contracts.LegalHoldAuthority {
	return contracts.LegalHoldAuthority{
		Version: 1, HoldID: "hold-1", State: state, AuthorityType: "x/group",
		PolicyDigest: "policy", EvidenceDigest: "evidence", Approvals: 2, Threshold: 2,
	}
}

func testFixtureSecurity() FixtureSecurityOptions {
	return FixtureSecurityOptions{UnsafeWindowsDevelopment: true}
}

func newTestArtifactStore(root, profile string, anchor RevisionAnchor) (*FixtureFileArtifactStore, error) {
	return NewFixtureFileArtifactStoreWithSecurity(root, profile, testFixtureSecurity(), anchor)
}

func newTestAuditStore(path, profile string, anchor RevisionAnchor) (*FixtureFileAuditStore, error) {
	return NewFixtureFileAuditStoreWithSecurity(path, profile, testFixtureSecurity(), anchor)
}

func newTestKeyPersistence(path string, wrappingKey []byte, profile string, anchor RevisionAnchor) (*keys.FixtureFilePersistence, error) {
	return keys.NewFixtureFilePersistenceWithSecurity(path, wrappingKey, profile, testFixtureSecurity(), anchor)
}

func TestFixtureVaultRestartRotationAndCryptoErasure(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	keyPath := filepath.Join(root, "custody", "keys.state")
	artifactPath := filepath.Join(root, "artifacts")
	wrappingKey := []byte("0123456789abcdef0123456789abcdef")
	anchor := NewProcessRevisionAnchor()

	persistence, err := newTestKeyPersistence(keyPath, wrappingKey, "fixture", anchor)
	require.NoError(t, err)
	keyManager, err := keys.NewUninitializedPersistentKeyManager(persistence)
	require.NoError(t, err)
	require.NoError(t, keyManager.Initialize())
	backend, err := newTestArtifactStore(artifactPath, "fixture", anchor)
	require.NoError(t, err)
	store, err := NewEncryptedBlobStoreWithError(backend, keyManager)
	require.NoError(t, err)

	blob, err := store.Store(ctx, &UploadRequest{Scope: ScopeSupport, Plaintext: []byte("durable secret"), Owner: "owner"})
	require.NoError(t, err)
	secondBlob, err := store.Store(ctx, &UploadRequest{Scope: ScopeSupport, Plaintext: []byte("second durable secret"), Owner: "owner"})
	require.NoError(t, err)
	originalKeyID := blob.Metadata.KeyID
	originalBackendRef := blob.Metadata.BackendRef
	require.NoError(t, store.Close())

	reloadedPersistence, err := newTestKeyPersistence(keyPath, wrappingKey, "fixture", anchor)
	require.NoError(t, err)
	reloadedKeys, err := keys.NewPersistentKeyManager(reloadedPersistence)
	require.NoError(t, err)
	reloadedBackend, err := newTestArtifactStore(artifactPath, "fixture", anchor)
	require.NoError(t, err)
	reloadedStore, err := NewEncryptedBlobStoreWithError(reloadedBackend, reloadedKeys)
	require.NoError(t, err)
	plaintext, metadata, err := reloadedStore.Retrieve(ctx, blob.Metadata.ID)
	require.NoError(t, err)
	require.Equal(t, []byte("durable secret"), plaintext)
	require.Equal(t, originalKeyID, metadata.KeyID)
	require.Equal(t, originalBackendRef, metadata.BackendRef)

	vault, err := NewVaultService(VaultConfig{
		Store: reloadedStore, ConsentResolver: AllowAllConsentResolver{},
		AuditLogger: NewAuditLogger(DefaultAuditLogConfig(), NewMemoryAuditStore()),
	})
	require.NoError(t, err)
	require.NoError(t, vault.RotateKeys(ctx, ScopeSupport, "owner", ""))
	require.NoError(t, vault.Close())

	afterRotationPersistence, err := newTestKeyPersistence(keyPath, wrappingKey, "fixture", anchor)
	require.NoError(t, err)
	afterRotationKeys, err := keys.NewPersistentKeyManager(afterRotationPersistence)
	require.NoError(t, err)
	rotation, err := afterRotationKeys.GetRotationStatus(keys.ScopeSupport)
	require.NoError(t, err)
	require.Equal(t, keys.RotationStatusCompleted, rotation.Status)
	afterRotationBackend, err := newTestArtifactStore(artifactPath, "fixture", anchor)
	require.NoError(t, err)
	afterRotationStore, err := NewEncryptedBlobStoreWithError(afterRotationBackend, afterRotationKeys)
	require.NoError(t, err)
	plaintext, rotatedMetadata, err := afterRotationStore.Retrieve(ctx, blob.Metadata.ID)
	require.NoError(t, err)
	require.Equal(t, []byte("durable secret"), plaintext)
	require.NotEqual(t, originalKeyID, rotatedMetadata.KeyID)
	_, secondRotatedMetadata, err := afterRotationStore.Retrieve(ctx, secondBlob.Metadata.ID)
	require.NoError(t, err)
	require.Equal(t, rotatedMetadata.KeyID, secondRotatedMetadata.KeyID)

	_, err = afterRotationKeys.DestroyKeyForFixtureErasure(keys.ScopeSupport, rotatedMetadata.KeyID, "unsafe", nil, fixtureHoldVerifier{})
	require.ErrorContains(t, err, "disabled")
	erasureVerifier := &fixtureErasureVerifier{authorized: make(map[string]bool)}
	coordinator, err := NewFixtureErasureCoordinator(afterRotationBackend, afterRotationKeys, erasureVerifier)
	require.NoError(t, err)
	authorization, err := coordinator.PrepareAuthorization(ScopeSupport, rotatedMetadata.KeyID, "fixture://authorization/1")
	require.NoError(t, err)
	_, err = coordinator.Erase(ctx, authorization)
	require.ErrorContains(t, err, "not approved")
	erasureVerifier.allow(authorization)
	omitted := authorization
	omitted.BlobIDs = omitted.BlobIDs[:1]
	omitted.Digest = fixtureAuthorizationDigest(omitted)
	erasureVerifier.allow(omitted)
	_, err = coordinator.Erase(ctx, omitted)
	require.ErrorContains(t, err, "omitted affected blob")

	address := afterRotationStore.resolveContentAddress(rotatedMetadata, blob.Metadata.ID)
	require.NoError(t, afterRotationBackend.SetLegalHold(address, fixtureHold(contracts.HoldActive), fixtureHoldVerifier{}))
	authorization, err = coordinator.PrepareAuthorization(ScopeSupport, rotatedMetadata.KeyID, "fixture://authorization/1")
	require.NoError(t, err)
	erasureVerifier.allow(authorization)
	_, err = coordinator.Erase(ctx, authorization)
	require.ErrorContains(t, err, "active legal hold")
	require.NoError(t, afterRotationBackend.SetLegalHold(address, fixtureHold(contracts.HoldReleased), fixtureHoldVerifier{}))
	authorization, err = coordinator.PrepareAuthorization(ScopeSupport, rotatedMetadata.KeyID, "fixture://authorization/1")
	require.NoError(t, err)
	erasureVerifier.allow(authorization)
	receipt, err := coordinator.Erase(ctx, authorization)
	require.NoError(t, err)
	require.Equal(t, rotatedMetadata.KeyID, receipt.TargetID)
	require.NotEmpty(t, receipt.Digest)
	require.Equal(t, FixtureErasureComplete, afterRotationBackend.index.ErasureIntents[authorization.Digest].State)
	_, _, err = afterRotationStore.Retrieve(ctx, blob.Metadata.ID)
	require.ErrorIs(t, err, ErrBlobNotFound)
	require.NoError(t, afterRotationStore.Close())

	destroyedPersistence, err := newTestKeyPersistence(keyPath, wrappingKey, "fixture", anchor)
	require.NoError(t, err)
	destroyedKeys, err := keys.NewPersistentKeyManager(destroyedPersistence)
	require.NoError(t, err)
	destroyedBackend, err := newTestArtifactStore(artifactPath, "fixture", anchor)
	require.NoError(t, err)
	destroyedStore, err := NewEncryptedBlobStoreWithError(destroyedBackend, destroyedKeys)
	require.NoError(t, err)
	exists, err := destroyedBackend.Exists(ctx, destroyedStore.resolveContentAddress(rotatedMetadata, blob.Metadata.ID))
	require.NoError(t, err)
	require.False(t, exists)
	_, _, err = destroyedStore.Retrieve(ctx, blob.Metadata.ID)
	require.ErrorIs(t, err, ErrBlobNotFound)
	_, err = destroyedKeys.GetKey(keys.ScopeSupport, rotatedMetadata.KeyID)
	require.Error(t, err)
	require.Equal(t, FixtureErasureComplete, destroyedBackend.index.ErasureIntents[authorization.Digest].State)
	require.NoError(t, destroyedStore.Close())
}

func TestFixtureErasureCoordinatorResumesAfterStorageDeletionRestart(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	keyPath := filepath.Join(root, "keys.state")
	artifactPath := filepath.Join(root, "artifacts")
	wrappingKey := []byte("0123456789abcdef0123456789abcdef")
	anchor := NewProcessRevisionAnchor()
	persistence, err := newTestKeyPersistence(keyPath, wrappingKey, "fixture", anchor)
	require.NoError(t, err)
	keyManager, err := keys.NewUninitializedPersistentKeyManager(persistence)
	require.NoError(t, err)
	require.NoError(t, keyManager.Initialize())
	backend, err := newTestArtifactStore(artifactPath, "fixture", anchor)
	require.NoError(t, err)
	store, err := NewEncryptedBlobStoreWithError(backend, keyManager)
	require.NoError(t, err)
	blob, err := store.Store(ctx, &UploadRequest{Scope: ScopeSupport, Plaintext: []byte("resume secret"), Owner: "owner"})
	require.NoError(t, err)
	erasureVerifier := &fixtureErasureVerifier{authorized: make(map[string]bool)}
	coordinator, err := NewFixtureErasureCoordinator(backend, keyManager, erasureVerifier)
	require.NoError(t, err)
	authorization, err := coordinator.PrepareAuthorization(ScopeSupport, blob.Metadata.KeyID, "fixture://authorization/resume")
	require.NoError(t, err)
	erasureVerifier.allow(authorization)

	backend.mu.Lock()
	targets, err := coordinator.validateTargetsLocked(authorization)
	require.NoError(t, err)
	backend.index.ErasureIntents[authorization.Digest] = &FixtureErasureTombstone{
		ID: authorization.Digest, AuthorizationDigest: authorization.Digest, AuthorizationRef: authorization.ObjectRef,
		Scope: authorization.Scope, KeyID: authorization.KeyID, ArtifactRevision: authorization.ArtifactRevision,
		KeyRevision: authorization.KeyRevision, Targets: targets, State: FixtureErasureIntent, UpdatedAt: time.Now().UTC(),
	}
	require.NoError(t, backend.persist())
	backend.mu.Unlock()
	require.NoError(t, coordinator.resumeStorageDeletion(authorization.Digest))
	require.Equal(t, FixtureErasureStorageDeleted, backend.index.ErasureIntents[authorization.Digest].State)
	require.NoError(t, store.Close())

	restartedPersistence, err := newTestKeyPersistence(keyPath, wrappingKey, "fixture", anchor)
	require.NoError(t, err)
	restartedKeys, err := keys.NewPersistentKeyManager(restartedPersistence)
	require.NoError(t, err)
	restartedBackend, err := newTestArtifactStore(artifactPath, "fixture", anchor)
	require.NoError(t, err)
	restartedCoordinator, err := NewFixtureErasureCoordinator(restartedBackend, restartedKeys, erasureVerifier)
	require.NoError(t, err)
	receipt, err := restartedCoordinator.Erase(ctx, authorization)
	require.NoError(t, err)
	require.Equal(t, authorization.KeyID, receipt.TargetID)
	require.Equal(t, FixtureErasureComplete, restartedBackend.index.ErasureIntents[authorization.Digest].State)
	_, err = restartedKeys.GetKey(keys.ScopeSupport, authorization.KeyID)
	require.Error(t, err)
	require.NoError(t, restartedBackend.Close())
	require.NoError(t, restartedKeys.Close())
}

func TestFixtureErasureCheckpointFailureNeverRestoresLiveMetadata(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	keyPath := filepath.Join(root, "keys.state")
	artifactPath := filepath.Join(root, "artifacts")
	wrappingKey := []byte("0123456789abcdef0123456789abcdef")
	anchor := NewProcessRevisionAnchor()
	persistence, err := newTestKeyPersistence(keyPath, wrappingKey, "fixture", anchor)
	require.NoError(t, err)
	keyManager, err := keys.NewUninitializedPersistentKeyManager(persistence)
	require.NoError(t, err)
	require.NoError(t, keyManager.Initialize())
	backend, err := newTestArtifactStore(artifactPath, "fixture", anchor)
	require.NoError(t, err)
	store, err := NewEncryptedBlobStoreWithError(backend, keyManager)
	require.NoError(t, err)
	blob, err := store.Store(ctx, &UploadRequest{Scope: ScopeSupport, Plaintext: []byte("erase-crash"), Owner: "owner"})
	require.NoError(t, err)
	verifier := &fixtureErasureVerifier{authorized: make(map[string]bool)}
	coordinator, err := NewFixtureErasureCoordinator(backend, keyManager, verifier)
	require.NoError(t, err)
	authorization, err := coordinator.PrepareAuthorization(ScopeSupport, blob.Metadata.KeyID, "fixture://authorization/checkpoint")
	require.NoError(t, err)
	verifier.allow(authorization)
	backend.failPersistAt = backend.persistCalls + 2
	_, err = coordinator.Erase(ctx, authorization)
	require.ErrorIs(t, err, ErrReconciliationRequired)
	_, err = store.GetMetadata(blob.Metadata.ID)
	require.ErrorIs(t, err, ErrBlobNotFound)
	_, err = keyManager.GetKey(keys.ScopeSupport, blob.Metadata.KeyID)
	require.NoError(t, err)
	require.NoError(t, store.Close())

	restartedPersistence, err := newTestKeyPersistence(keyPath, wrappingKey, "fixture", anchor)
	require.NoError(t, err)
	restartedKeys, err := keys.NewPersistentKeyManager(restartedPersistence)
	require.NoError(t, err)
	restartedBackend, err := newTestArtifactStore(artifactPath, "fixture", anchor)
	require.NoError(t, err)
	restartedCoordinator, err := NewFixtureErasureCoordinator(restartedBackend, restartedKeys, verifier)
	require.NoError(t, err)
	require.NoError(t, restartedCoordinator.ResumePending(ctx))
	_, err = restartedKeys.GetKey(keys.ScopeSupport, blob.Metadata.KeyID)
	require.Error(t, err)
	require.Equal(t, FixtureErasureComplete, restartedBackend.index.ErasureIntents[authorization.Digest].State)
	require.NoError(t, restartedBackend.Close())
	require.NoError(t, restartedKeys.Close())
}

func TestFixtureArtifactStoreLegalHoldOwnerIsolationDeleteAndExclusiveLease(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	anchor := NewProcessRevisionAnchor()
	first, err := newTestArtifactStore(root, "fixture", anchor)
	require.NoError(t, err)
	_, err = newTestArtifactStore(root, "fixture", anchor)
	require.ErrorIs(t, err, errFixtureStoreInUse)
	put := func(store *FixtureFileArtifactStore, owner, value string) (*artifact_store.PutResponse, error) {
		return store.Put(ctx, &artifact_store.PutRequest{
			Data: []byte(value), Owner: owner, ArtifactType: "support",
			EncryptionMetadata: &artifact_store.EncryptionMetadata{AlgorithmID: "fixture"},
		})
	}
	stored, err := put(first, "owner-a", "ciphertext-a")
	require.NoError(t, err)
	_, err = put(first, "owner-b", "ciphertext-a")
	require.ErrorContains(t, err, "already owned")

	_, err = first.Get(ctx, &artifact_store.GetRequest{ContentAddress: stored.ContentAddress, RequestingAccount: "owner-b"})
	require.Error(t, err)
	require.NoError(t, first.SetLegalHold(stored.ContentAddress, fixtureHold(contracts.HoldActive), fixtureHoldVerifier{}))
	err = first.Delete(ctx, &artifact_store.DeleteRequest{ContentAddress: stored.ContentAddress, RequestingAccount: "owner-a", Force: true})
	require.ErrorContains(t, err, "active legal hold")

	require.NoError(t, first.SetLegalHold(stored.ContentAddress, fixtureHold(contracts.HoldReleased), fixtureHoldVerifier{}))
	require.NoError(t, first.Delete(ctx, &artifact_store.DeleteRequest{ContentAddress: stored.ContentAddress, RequestingAccount: "owner-a", Force: true}))
	require.NoError(t, first.Close())
	restarted, err := newTestArtifactStore(root, "fixture", anchor)
	require.NoError(t, err)
	exists, err := restarted.Exists(ctx, stored.ContentAddress)
	require.NoError(t, err)
	require.False(t, exists)
	require.NoError(t, restarted.Close())

	_, err = newTestArtifactStore(filepath.Join(t.TempDir(), "prod"), "production", NewProcessRevisionAnchor())
	require.Error(t, err)
}

func TestFixtureArtifactStoreRejectsUnsafeReference(t *testing.T) {
	store, err := newTestArtifactStore(t.TempDir(), "fixture", NewProcessRevisionAnchor())
	require.NoError(t, err)
	defer store.Close()
	_, err = store.Get(context.Background(), &artifact_store.GetRequest{ContentAddress: &artifact_store.ContentAddress{
		Version: 1, Hash: make([]byte, 32), Algorithm: "sha256", Backend: artifact_store.BackendIPFS, BackendRef: "../escape",
	}})
	require.Error(t, err)
	require.False(t, errors.Is(err, artifact_store.ErrArtifactNotFound))
}

func TestFixtureArtifactStoreChecksumValidation(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	store, err := newTestArtifactStore(root, "fixture", NewProcessRevisionAnchor())
	require.NoError(t, err)
	defer store.Close()
	stored, err := store.Put(ctx, &artifact_store.PutRequest{
		Data: []byte("ciphertext"), Owner: "owner", ArtifactType: "support",
		EncryptionMetadata: &artifact_store.EncryptionMetadata{AlgorithmID: "fixture"},
	})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(root, "objects", stored.ContentAddress.BackendRef), []byte("tampered!!"), 0o600))
	_, err = store.Get(ctx, &artifact_store.GetRequest{ContentAddress: stored.ContentAddress, RequestingAccount: "owner"})
	require.ErrorIs(t, err, artifact_store.ErrHashMismatch)
}

func TestFixtureArtifactStoreDeepCopiesInputsAndOutputs(t *testing.T) {
	ctx := context.Background()
	store, err := newTestArtifactStore(t.TempDir(), "fixture", NewProcessRevisionAnchor())
	require.NoError(t, err)
	defer store.Close()
	request := &artifact_store.PutRequest{
		Data: []byte("ciphertext"), Owner: "owner", ArtifactType: "support",
		EncryptionMetadata: &artifact_store.EncryptionMetadata{AlgorithmID: "fixture", RecipientKeyIDs: []string{"key-1"}},
		Metadata:           map[string]string{"classification": "restricted"},
	}
	stored, err := store.Put(ctx, request)
	require.NoError(t, err)
	request.EncryptionMetadata.RecipientKeyIDs[0] = "mutated"
	request.Metadata["classification"] = "public"
	stored.ArtifactReference.Metadata["classification"] = "response-mutated"
	stored.ContentAddress.Hash[0] ^= 0xff

	listed, err := store.ListByOwner(ctx, "owner", nil)
	require.NoError(t, err)
	require.Equal(t, "restricted", listed.References[0].Metadata["classification"])
	require.Equal(t, "key-1", listed.References[0].EncryptionMetadata.RecipientKeyIDs[0])
	listed.References[0].Metadata["classification"] = "list-mutated"
	listedAgain, err := store.ListByOwner(ctx, "owner", nil)
	require.NoError(t, err)
	require.Equal(t, "restricted", listedAgain.References[0].Metadata["classification"])
}

func TestFixtureBlobPutJournalRecoversAfterBytesBeforeMetadata(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	anchor := NewProcessRevisionAnchor()
	backend, err := newTestArtifactStore(root, "fixture", anchor)
	require.NoError(t, err)
	keyManager := keys.NewKeyManager()
	require.NoError(t, keyManager.Initialize())
	store, err := NewEncryptedBlobStoreWithError(backend, keyManager)
	require.NoError(t, err)
	backend.failPersistAt = backend.persistCalls + 2
	blob, err := store.Store(ctx, &UploadRequest{Scope: ScopeSupport, Plaintext: []byte("put-crash"), Owner: "owner"})
	require.NotNil(t, blob)
	require.ErrorIs(t, err, ErrReconciliationRequired)
	require.NoError(t, backend.Close())

	restarted, err := newTestArtifactStore(root, "fixture", anchor)
	require.NoError(t, err)
	restartedStore, err := NewEncryptedBlobStoreWithError(restarted, keyManager)
	require.NoError(t, err)
	metadata, err := restartedStore.ListByScope(ScopeSupport)
	require.NoError(t, err)
	require.Len(t, metadata, 1)
	plaintext, _, err := restartedStore.Retrieve(ctx, metadata[0].ID)
	require.NoError(t, err)
	require.Equal(t, []byte("put-crash"), plaintext)
	require.NoError(t, restarted.Close())
}

func TestFixtureBlobDeleteJournalRecoversAfterBytesBeforeMetadata(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	anchor := NewProcessRevisionAnchor()
	backend, err := newTestArtifactStore(root, "fixture", anchor)
	require.NoError(t, err)
	keyManager := keys.NewKeyManager()
	require.NoError(t, keyManager.Initialize())
	store, err := NewEncryptedBlobStoreWithError(backend, keyManager)
	require.NoError(t, err)
	blob, err := store.Store(ctx, &UploadRequest{Scope: ScopeSupport, Plaintext: []byte("delete-crash"), Owner: "owner"})
	require.NoError(t, err)
	backend.failPersistAt = backend.persistCalls + 2
	err = store.Delete(ctx, blob.Metadata.ID)
	require.ErrorIs(t, err, ErrReconciliationRequired)
	_, err = store.GetMetadata(blob.Metadata.ID)
	require.ErrorIs(t, err, ErrBlobNotFound)
	require.NoError(t, backend.Close())

	restarted, err := newTestArtifactStore(root, "fixture", anchor)
	require.NoError(t, err)
	restartedStore, err := NewEncryptedBlobStoreWithError(restarted, keyManager)
	require.NoError(t, err)
	_, err = restartedStore.GetMetadata(blob.Metadata.ID)
	require.ErrorIs(t, err, ErrBlobNotFound)
	exists, err := restarted.Exists(ctx, store.resolveContentAddress(&blob.Metadata, blob.Metadata.ID))
	require.NoError(t, err)
	require.False(t, exists)
	require.NoError(t, restarted.Close())
}

func TestFixtureAuditStoreRestartContinuesChain(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.state")
	anchor := NewProcessRevisionAnchor()
	store, err := newTestAuditStore(path, "fixture", anchor)
	require.NoError(t, err)
	_, err = newTestAuditStore(path, "fixture", anchor)
	require.ErrorIs(t, err, errFixtureStoreInUse)
	logger := NewAuditLogger(DefaultAuditLogConfig(), store)
	require.NoError(t, logger.LogEvent(context.Background(), &AuditEvent{EventType: "upload", Requester: "owner"}))
	first, err := store.Query(context.Background(), AuditFilter{})
	require.NoError(t, err)
	require.Len(t, first, 1)
	require.NoError(t, store.Close())

	restored, err := newTestAuditStore(path, "fixture", anchor)
	require.NoError(t, err)
	restoredLogger := NewAuditLogger(DefaultAuditLogConfig(), restored)
	require.NoError(t, restoredLogger.LogEvent(context.Background(), &AuditEvent{EventType: "read", Requester: "owner"}))
	events, err := restored.Query(context.Background(), AuditFilter{})
	require.NoError(t, err)
	require.Len(t, events, 2)
	require.Equal(t, events[0].Hash, events[1].PreviousHash)
	require.NoError(t, restored.Close())
}

func TestFixtureAuditStoreRejectsRechecksummedChainTamper(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.state")
	anchor := NewProcessRevisionAnchor()
	store, err := newTestAuditStore(path, "fixture", anchor)
	require.NoError(t, err)
	logger := NewAuditLogger(DefaultAuditLogConfig(), store)
	require.NoError(t, logger.LogEvent(context.Background(), &AuditEvent{EventType: "upload", Requester: "owner"}))
	require.NoError(t, store.Close())

	encoded, err := os.ReadFile(path)
	require.NoError(t, err)
	var envelope fixtureAuditEnvelope
	require.NoError(t, json.Unmarshal(encoded, &envelope))
	var events []*AuditEvent
	require.NoError(t, json.Unmarshal(envelope.Events, &events))
	events[0].Requester = "attacker"
	envelope.Events, err = json.Marshal(events)
	require.NoError(t, err)
	digest := sha256.Sum256(envelope.Events)
	envelope.Checksum = hex.EncodeToString(digest[:])
	encoded, err = json.Marshal(envelope)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, encoded, 0o600))

	_, err = newTestAuditStore(path, "fixture", anchor)
	require.ErrorContains(t, err, "hash mismatch")
}

func TestFixtureArtifactAndAuditStoresRejectOlderValidReplay(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	anchor := NewProcessRevisionAnchor()
	artifacts, err := newTestArtifactStore(filepath.Join(root, "artifacts"), "fixture", anchor)
	require.NoError(t, err)
	_, err = artifacts.Put(ctx, &artifact_store.PutRequest{
		Data: []byte("first"), Owner: "owner", ArtifactType: "support",
		EncryptionMetadata: &artifact_store.EncryptionMetadata{AlgorithmID: "fixture"},
	})
	require.NoError(t, err)
	olderIndex, err := os.ReadFile(artifacts.indexPath())
	require.NoError(t, err)
	_, err = artifacts.Put(ctx, &artifact_store.PutRequest{
		Data: []byte("second"), Owner: "owner", ArtifactType: "support",
		EncryptionMetadata: &artifact_store.EncryptionMetadata{AlgorithmID: "fixture"},
	})
	require.NoError(t, err)
	require.NoError(t, artifacts.Close())
	require.NoError(t, os.WriteFile(filepath.Join(root, "artifacts", "index.json"), olderIndex, 0o600))
	_, err = newTestArtifactStore(filepath.Join(root, "artifacts"), "fixture", anchor)
	require.ErrorIs(t, err, ErrRevisionRollback)

	auditPath := filepath.Join(root, "audit.state")
	auditAnchor := NewProcessRevisionAnchor()
	audit, err := newTestAuditStore(auditPath, "fixture", auditAnchor)
	require.NoError(t, err)
	logger := NewAuditLogger(DefaultAuditLogConfig(), audit)
	require.NoError(t, logger.LogEvent(ctx, &AuditEvent{EventType: "first", Requester: "owner"}))
	olderAudit, err := os.ReadFile(auditPath)
	require.NoError(t, err)
	require.NoError(t, logger.LogEvent(ctx, &AuditEvent{EventType: "second", Requester: "owner"}))
	require.NoError(t, audit.Close())
	require.NoError(t, os.WriteFile(auditPath, olderAudit, 0o600))
	_, err = newTestAuditStore(auditPath, "fixture", auditAnchor)
	require.ErrorIs(t, err, ErrRevisionRollback)
}
