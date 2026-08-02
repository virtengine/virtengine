package data_vault

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/virtengine/virtengine/pkg/data_vault/keys"
)

type instrumentedRotationKeyManager struct {
	keys.VaultKeyManager
	mu             sync.Mutex
	getActiveCalls int
	rotateCalls    int
}

func (m *instrumentedRotationKeyManager) GetActiveKey(scope keys.Scope) (*keys.KeyInfo, error) {
	m.mu.Lock()
	m.getActiveCalls++
	m.mu.Unlock()
	return m.VaultKeyManager.GetActiveKey(scope)
}

func (m *instrumentedRotationKeyManager) RotateKey(scope keys.Scope, overlap time.Duration) error {
	m.mu.Lock()
	m.rotateCalls++
	m.mu.Unlock()
	return m.VaultKeyManager.RotateKey(scope, overlap)
}

func (m *instrumentedRotationKeyManager) reset() {
	m.mu.Lock()
	m.getActiveCalls = 0
	m.rotateCalls = 0
	m.mu.Unlock()
}

func (m *instrumentedRotationKeyManager) calls() (int, int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.getActiveCalls, m.rotateCalls
}

type recordingAccessControl struct {
	mu        sync.Mutex
	requests  []AccessRequest
	authorize func(context.Context, AccessRequest) error
}

func (a *recordingAccessControl) Authorize(ctx context.Context, req AccessRequest) error {
	a.mu.Lock()
	a.requests = append(a.requests, req)
	a.mu.Unlock()
	return a.authorize(ctx, req)
}

func (a *recordingAccessControl) reset() {
	a.mu.Lock()
	a.requests = nil
	a.mu.Unlock()
}

func (a *recordingAccessControl) recorded() []AccessRequest {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]AccessRequest(nil), a.requests...)
}

func artifactPutCount(store *memoryArtifactStore) int {
	store.mu.RLock()
	defer store.mu.RUnlock()
	count := 0
	for _, refs := range store.owners {
		count += len(refs)
	}
	return count
}

func TestVaultService_AccessAndAudit(t *testing.T) {
	ctx := context.Background()

	keyMgr := keys.NewKeyManager()
	if err := keyMgr.Initialize(); err != nil {
		t.Fatalf("init key manager: %v", err)
	}

	store := NewEncryptedBlobStore(newMemoryArtifactStore(), keyMgr)

	access := NewPolicyAccessControl(DefaultAccessPolicy(),
		StaticRoleResolver{Roles: map[string]map[string]bool{
			"admin": {RoleAdministrator: true},
		}},
		StaticOrgResolver{Members: map[string]map[string]bool{
			"org-1": {"admin": true},
		}},
	)

	auditLogger := NewAuditLogger(DefaultAuditLogConfig(), NewMemoryAuditStore())

	vault, err := NewVaultService(VaultConfig{
		Store:           store,
		AccessControl:   access,
		ConsentResolver: AllowAllConsentResolver{},
		AuditLogger:     auditLogger,
	})
	if err != nil {
		t.Fatalf("new vault: %v", err)
	}

	blob, err := vault.Upload(ctx, &UploadRequest{
		Scope:     ScopeSupport,
		Plaintext: []byte("secret"),
		Owner:     "owner1",
		OrgID:     "org-1",
	})
	if err != nil {
		t.Fatalf("upload: %v", err)
	}

	_, _, err = vault.Retrieve(ctx, &RetrieveRequest{
		ID:        blob.Metadata.ID,
		Requester: "user2",
		OrgID:     "org-2",
		Reason:    "test",
	})
	if err == nil {
		t.Fatalf("expected unauthorized access to fail")
	}

	_, _, err = vault.Retrieve(ctx, &RetrieveRequest{
		ID:        blob.Metadata.ID,
		Requester: "admin",
		OrgID:     "org-1",
		Reason:    "support",
	})
	if err != nil {
		t.Fatalf("expected admin access to succeed: %v", err)
	}

	events, err := auditLogger.QueryEvents(ctx, AuditFilter{BlobID: blob.Metadata.ID})
	if err != nil {
		t.Fatalf("query audit: %v", err)
	}
	if len(events) < 2 {
		t.Fatalf("expected audit events for access attempts")
	}
}

type denyConsentResolver struct{}

func (denyConsentResolver) HasConsent(_ context.Context, _ ConsentRequest) (bool, error) {
	return false, nil
}

func TestVaultService_ConsentRequired(t *testing.T) {
	ctx := context.Background()

	keyMgr := keys.NewKeyManager()
	if err := keyMgr.Initialize(); err != nil {
		t.Fatalf("init key manager: %v", err)
	}

	store := NewEncryptedBlobStore(newMemoryArtifactStore(), keyMgr)
	access := NewPolicyAccessControl(DefaultAccessPolicy(), nil, StaticOrgResolver{Members: map[string]map[string]bool{
		"org-1": {"owner": true},
	}})
	auditLogger := NewAuditLogger(DefaultAuditLogConfig(), NewMemoryAuditStore())

	vault, err := NewVaultService(VaultConfig{
		Store:           store,
		AccessControl:   access,
		ConsentResolver: denyConsentResolver{},
		AuditLogger:     auditLogger,
	})
	if err != nil {
		t.Fatalf("new vault: %v", err)
	}

	blob, err := vault.Upload(ctx, &UploadRequest{
		Scope:     ScopeSupport,
		Plaintext: []byte("secret"),
		Owner:     "owner",
		OrgID:     "org-1",
	})
	if err != nil {
		t.Fatalf("upload: %v", err)
	}

	_, _, err = vault.Retrieve(ctx, &RetrieveRequest{
		ID:        blob.Metadata.ID,
		Requester: "owner",
		OrgID:     "org-1",
		Purpose:   "audit",
	})
	if err == nil {
		t.Fatalf("expected consent error")
	}
	if !errors.Is(err, ErrConsentRequired) {
		t.Fatalf("expected ErrConsentRequired, got %v", err)
	}
}

func TestVaultService_MissingConsentResolverDenies(t *testing.T) {
	ctx := context.Background()
	keyMgr := keys.NewKeyManager()
	require.NoError(t, keyMgr.Initialize())
	store := NewEncryptedBlobStore(newMemoryArtifactStore(), keyMgr)
	vault, err := NewVaultService(VaultConfig{
		Store:       store,
		AuditLogger: NewAuditLogger(DefaultAuditLogConfig(), NewMemoryAuditStore()),
	})
	require.NoError(t, err)

	blob, err := vault.Upload(ctx, &UploadRequest{
		Scope: ScopeSupport, Plaintext: []byte("secret"), Owner: "owner",
	})
	require.NoError(t, err)

	_, _, err = vault.Retrieve(ctx, &RetrieveRequest{ID: blob.Metadata.ID, Requester: "owner"})
	require.ErrorIs(t, err, ErrConsentRequired)
}

func TestVaultServiceRequiresExplicitAuditLogger(t *testing.T) {
	keyMgr := keys.NewKeyManager()
	require.NoError(t, keyMgr.Initialize())
	store := NewEncryptedBlobStore(newMemoryArtifactStore(), keyMgr)
	_, err := NewVaultService(VaultConfig{Store: store})
	require.ErrorIs(t, err, ErrInvalidRequest)
}

func TestVaultService_RotateKeysReencrypts(t *testing.T) {
	ctx := context.Background()

	keyMgr := keys.NewKeyManager()
	require.NoError(t, keyMgr.Initialize())

	store := NewEncryptedBlobStore(newMemoryArtifactStore(), keyMgr)
	access := NewPolicyAccessControl(DefaultAccessPolicy(), nil, StaticOrgResolver{Members: map[string]map[string]bool{
		"org-1": {"owner": true},
	}})
	auditLogger := NewAuditLogger(DefaultAuditLogConfig(), NewMemoryAuditStore())

	vault, err := NewVaultService(VaultConfig{
		Store:           store,
		AccessControl:   access,
		ConsentResolver: AllowAllConsentResolver{},
		AuditLogger:     auditLogger,
	})
	require.NoError(t, err)

	blob, err := vault.Upload(ctx, &UploadRequest{
		Scope:     ScopeSupport,
		Plaintext: []byte("rotate-me"),
		Owner:     "owner",
		OrgID:     "org-1",
	})
	require.NoError(t, err)

	originalVersion := blob.Metadata.KeyVersion
	require.NoError(t, vault.RotateKeys(ctx, ScopeSupport, "owner", "org-1"))

	_, meta, err := vault.Retrieve(ctx, &RetrieveRequest{
		ID:        blob.Metadata.ID,
		Requester: "owner",
		OrgID:     "org-1",
	})
	require.NoError(t, err)
	require.Greater(t, meta.KeyVersion, originalVersion)
}

func TestVaultService_RotateKeysPreauthorizesAllBlobsBeforeMutation(t *testing.T) {
	ctx := context.Background()
	baseKeyManager := keys.NewKeyManager()
	require.NoError(t, baseKeyManager.Initialize())
	keyManager := &instrumentedRotationKeyManager{VaultKeyManager: baseKeyManager}
	backend := newMemoryArtifactStore()
	policy := NewPolicyAccessControl(DefaultAccessPolicy(), nil, StaticOrgResolver{
		Members: map[string]map[string]bool{"org-1": {"member": true}},
		Roles:   map[string]map[string]string{"org-1": {"member": "member"}},
	})
	access := &recordingAccessControl{}
	access.authorize = func(ctx context.Context, req AccessRequest) error {
		if req.Action == AccessActionRotate {
			getActiveCalls, rotateCalls := keyManager.calls()
			require.Zero(t, getActiveCalls)
			require.Zero(t, rotateCalls)
		}
		return policy.Authorize(ctx, req)
	}
	auditLogger := NewAuditLogger(DefaultAuditLogConfig(), NewMemoryAuditStore())
	vault, err := NewVaultService(VaultConfig{
		Store: NewEncryptedBlobStore(backend, keyManager), AccessControl: access,
		ConsentResolver: AllowAllConsentResolver{}, AuditLogger: auditLogger,
	})
	require.NoError(t, err)
	for _, owner := range []string{"owner-1", "owner-2"} {
		_, err := vault.Upload(ctx, &UploadRequest{
			Scope: ScopeSupport, Plaintext: []byte(owner), Owner: owner, OrgID: "org-1",
		})
		require.NoError(t, err)
	}
	keyManager.reset()
	access.reset()

	require.NoError(t, vault.RotateKeys(ctx, ScopeSupport, "member", "org-1"))
	requests := access.recorded()
	require.Len(t, requests, 2)
	for _, req := range requests {
		require.Equal(t, AccessActionRotate, req.Action)
		require.Equal(t, "member", req.Requester)
		require.Equal(t, ScopeSupport, req.Scope)
		require.Equal(t, "org-1", req.OrgID)
		require.Equal(t, "org-1", req.ResourceOrgID)
		require.NotEmpty(t, req.Owner)
	}
	require.Equal(t, 4, artifactPutCount(backend))

	events, err := auditLogger.QueryEvents(ctx, AuditFilter{Scope: ScopeSupport, Requester: "member"})
	require.NoError(t, err)
	require.Len(t, events, 2)
	for _, event := range events {
		require.Equal(t, string(AccessActionRotate), event.EventType)
		require.Equal(t, "member", event.Requester)
		require.Equal(t, "org-1", event.OrgID)
		require.Equal(t, "org-1", event.Metadata["resource_org_id"])
		require.True(t, event.Success)
	}
}

func TestVaultService_RotateKeysLaterDenialHasNoMutation(t *testing.T) {
	ctx := context.Background()
	baseKeyManager := keys.NewKeyManager()
	require.NoError(t, baseKeyManager.Initialize())
	keyManager := &instrumentedRotationKeyManager{VaultKeyManager: baseKeyManager}
	backend := newMemoryArtifactStore()
	access := &recordingAccessControl{}
	rotateAuthorizations := 0
	access.authorize = func(_ context.Context, req AccessRequest) error {
		if req.Action != AccessActionRotate {
			return nil
		}
		rotateAuthorizations++
		if rotateAuthorizations == 2 {
			return NewVaultError("Authorize", ErrUnauthorized, "second blob denied")
		}
		return nil
	}
	auditLogger := NewAuditLogger(DefaultAuditLogConfig(), NewMemoryAuditStore())
	vault, err := NewVaultService(VaultConfig{
		Store: NewEncryptedBlobStore(backend, keyManager), AccessControl: access,
		ConsentResolver: AllowAllConsentResolver{}, AuditLogger: auditLogger,
	})
	require.NoError(t, err)
	blobs := make([]*EncryptedBlob, 0, 2)
	for _, owner := range []string{"owner-1", "owner-2"} {
		blob, err := vault.Upload(ctx, &UploadRequest{
			Scope: ScopeSupport, Plaintext: []byte(owner), Owner: owner, OrgID: "org-1",
		})
		require.NoError(t, err)
		blobs = append(blobs, blob)
	}
	keyManager.reset()
	access.reset()
	putsBefore := artifactPutCount(backend)

	err = vault.RotateKeys(ctx, ScopeSupport, "requester", "org-1")
	require.ErrorIs(t, err, ErrUnauthorized)
	require.Len(t, access.recorded(), 2)
	getActiveCalls, rotateCalls := keyManager.calls()
	require.Zero(t, getActiveCalls)
	require.Zero(t, rotateCalls)
	require.Equal(t, putsBefore, artifactPutCount(backend))
	for _, blob := range blobs {
		metadata, metadataErr := vault.store.GetMetadata(blob.Metadata.ID)
		require.NoError(t, metadataErr)
		require.Equal(t, blob.Metadata.KeyID, metadata.KeyID)
	}

	events, queryErr := auditLogger.QueryEvents(ctx, AuditFilter{Requester: "requester"})
	require.NoError(t, queryErr)
	require.Len(t, events, 1)
	require.Equal(t, string(AccessActionRotate), events[0].EventType)
	require.Equal(t, "org-1", events[0].OrgID)
	require.Equal(t, "org-1", events[0].Metadata["resource_org_id"])
	require.False(t, events[0].Success)
}

func TestVaultService_RotateKeysRejectsMissingRequesterAndOrgMismatchBeforeMutation(t *testing.T) {
	ctx := context.Background()
	baseKeyManager := keys.NewKeyManager()
	require.NoError(t, baseKeyManager.Initialize())
	keyManager := &instrumentedRotationKeyManager{VaultKeyManager: baseKeyManager}
	backend := newMemoryArtifactStore()
	policy := NewPolicyAccessControl(DefaultAccessPolicy(), nil, StaticOrgResolver{
		Members: map[string]map[string]bool{"org-1": {"member": true}},
		Roles:   map[string]map[string]string{"org-1": {"member": "member"}},
	})
	access := &recordingAccessControl{authorize: policy.Authorize}
	vault, err := NewVaultService(VaultConfig{
		Store: NewEncryptedBlobStore(backend, keyManager), AccessControl: access,
		ConsentResolver: AllowAllConsentResolver{},
		AuditLogger:     NewAuditLogger(DefaultAuditLogConfig(), NewMemoryAuditStore()),
	})
	require.NoError(t, err)
	_, err = vault.Upload(ctx, &UploadRequest{
		Scope: ScopeSupport, Plaintext: []byte("secret"), Owner: "owner", OrgID: "org-1",
	})
	require.NoError(t, err)
	keyManager.reset()
	access.reset()
	putsBefore := artifactPutCount(backend)

	err = vault.RotateKeys(ctx, ScopeSupport, "", "org-1")
	require.ErrorIs(t, err, ErrInvalidRequest)
	require.Empty(t, access.recorded())
	err = vault.RotateKeys(ctx, Scope("unknown"), "member", "org-1")
	require.ErrorIs(t, err, ErrInvalidScope)
	require.Empty(t, access.recorded())
	err = vault.RotateKeys(ctx, ScopeSupport, "member", "org-2")
	require.ErrorIs(t, err, ErrUnauthorized)
	require.Len(t, access.recorded(), 1)
	getActiveCalls, rotateCalls := keyManager.calls()
	require.Zero(t, getActiveCalls)
	require.Zero(t, rotateCalls)
	require.Equal(t, putsBefore, artifactPutCount(backend))
}

func TestVaultService_RotateKeysEmptyScopeRequiresAuthorization(t *testing.T) {
	ctx := context.Background()
	baseKeyManager := keys.NewKeyManager()
	require.NoError(t, baseKeyManager.Initialize())
	keyManager := &instrumentedRotationKeyManager{VaultKeyManager: baseKeyManager}
	access := &recordingAccessControl{authorize: func(_ context.Context, _ AccessRequest) error {
		return NewVaultError("Authorize", ErrUnauthorized, "rotation denied")
	}}
	vault, err := NewVaultService(VaultConfig{
		Store: NewEncryptedBlobStore(newMemoryArtifactStore(), keyManager), AccessControl: access,
		ConsentResolver: AllowAllConsentResolver{},
		AuditLogger:     NewAuditLogger(DefaultAuditLogConfig(), NewMemoryAuditStore()),
	})
	require.NoError(t, err)
	keyManager.reset()

	err = vault.RotateKeys(ctx, ScopeSupport, "requester", "org-1")
	require.ErrorIs(t, err, ErrUnauthorized)
	requests := access.recorded()
	require.Len(t, requests, 1)
	require.Equal(t, AccessRequest{
		Action: AccessActionRotate, Requester: "requester", Scope: ScopeSupport,
		OrgID: "org-1", ResourceOrgID: "org-1",
	}, requests[0])
	getActiveCalls, rotateCalls := keyManager.calls()
	require.Zero(t, getActiveCalls)
	require.Zero(t, rotateCalls)
}

func TestVaultServicePropagatesAuditFailures(t *testing.T) {
	ctx := context.Background()
	newVault := func(auditStore *failOnceAuditStore) *Vault {
		keyManager := keys.NewKeyManager()
		require.NoError(t, keyManager.Initialize())
		vault, err := NewVaultService(VaultConfig{
			Store:           NewEncryptedBlobStore(newMemoryArtifactStore(), keyManager),
			ConsentResolver: AllowAllConsentResolver{},
			AuditLogger:     NewAuditLogger(DefaultAuditLogConfig(), auditStore),
		})
		require.NoError(t, err)
		return vault
	}

	uploadAudit := &failOnceAuditStore{fail: true}
	uploadVault := newVault(uploadAudit)
	_, err := uploadVault.Upload(ctx, &UploadRequest{Scope: ScopeSupport, Plaintext: []byte("secret"), Owner: "owner"})
	require.ErrorContains(t, err, "audit intent append failed")

	retrieveAudit := &failOnceAuditStore{}
	retrieveVault := newVault(retrieveAudit)
	retrieveBlob, err := retrieveVault.Upload(ctx, &UploadRequest{Scope: ScopeSupport, Plaintext: []byte("secret"), Owner: "owner"})
	require.NoError(t, err)
	retrieveAudit.fail = true
	_, _, err = retrieveVault.Retrieve(ctx, &RetrieveRequest{ID: retrieveBlob.Metadata.ID, Requester: "owner"})
	require.ErrorContains(t, err, "audit append failed")

	deleteAudit := &failOnceAuditStore{}
	deleteVault := newVault(deleteAudit)
	deleteBlob, err := deleteVault.Upload(ctx, &UploadRequest{Scope: ScopeSupport, Plaintext: []byte("secret"), Owner: "owner"})
	require.NoError(t, err)
	deleteAudit.fail = true
	err = deleteVault.Delete(ctx, deleteBlob.Metadata.ID, "owner")
	require.ErrorContains(t, err, "audit intent append failed")
}

func TestVaultUploadAuditIntentAndTerminalFailureSemantics(t *testing.T) {
	ctx := context.Background()
	newVault := func(auditStore *failOnceAuditStore) *Vault {
		keyManager := keys.NewKeyManager()
		require.NoError(t, keyManager.Initialize())
		vault, err := NewVaultService(VaultConfig{
			Store: NewEncryptedBlobStore(newMemoryArtifactStore(), keyManager), ConsentResolver: AllowAllConsentResolver{},
			AuditLogger: NewAuditLogger(DefaultAuditLogConfig(), auditStore),
		})
		require.NoError(t, err)
		return vault
	}

	intentFailureStore := &failOnceAuditStore{failAt: 1}
	intentFailureVault := newVault(intentFailureStore)
	blob, err := intentFailureVault.Upload(ctx, &UploadRequest{Scope: ScopeSupport, Plaintext: []byte("intent"), Owner: "owner"})
	require.Nil(t, blob)
	require.ErrorContains(t, err, "audit intent append failed")
	metadata, listErr := intentFailureVault.store.ListByScope(ScopeSupport)
	require.NoError(t, listErr)
	require.Empty(t, metadata)

	terminalFailureStore := &failOnceAuditStore{failAt: 2}
	terminalFailureVault := newVault(terminalFailureStore)
	blob, err = terminalFailureVault.Upload(ctx, &UploadRequest{Scope: ScopeSupport, Plaintext: []byte("terminal"), Owner: "owner"})
	require.NotNil(t, blob)
	require.ErrorIs(t, err, ErrReconciliationRequired)
	_, metadataErr := terminalFailureVault.store.GetMetadata(blob.Metadata.ID)
	require.NoError(t, metadataErr)
	require.Len(t, terminalFailureStore.events, 1)
	require.Equal(t, "intent", terminalFailureStore.events[0].Metadata["phase"])
	require.NoError(t, terminalFailureVault.ReconcilePending(ctx))
	require.Len(t, terminalFailureStore.events, 2)
	require.Equal(t, "terminal", terminalFailureStore.events[1].Metadata["phase"])
	require.True(t, terminalFailureStore.events[1].Success)
}

func TestVaultDeleteTerminalAuditFailureLeavesCommittedMutationPending(t *testing.T) {
	ctx := context.Background()
	auditStore := &failOnceAuditStore{}
	keyManager := keys.NewKeyManager()
	require.NoError(t, keyManager.Initialize())
	vault, err := NewVaultService(VaultConfig{
		Store: NewEncryptedBlobStore(newMemoryArtifactStore(), keyManager), ConsentResolver: AllowAllConsentResolver{},
		AuditLogger: NewAuditLogger(DefaultAuditLogConfig(), auditStore),
	})
	require.NoError(t, err)
	blob, err := vault.Upload(ctx, &UploadRequest{Scope: ScopeSupport, Plaintext: []byte("delete"), Owner: "owner"})
	require.NoError(t, err)
	auditStore.failAt = auditStore.appendCalls + 2
	err = vault.Delete(ctx, blob.Metadata.ID, "owner")
	require.ErrorIs(t, err, ErrReconciliationRequired)
	_, metadataErr := vault.store.GetMetadata(blob.Metadata.ID)
	require.ErrorIs(t, metadataErr, ErrBlobNotFound)
	require.NoError(t, vault.ReconcilePending(ctx))
	events, queryErr := auditStore.Query(ctx, AuditFilter{})
	require.NoError(t, queryErr)
	require.Equal(t, "terminal", events[len(events)-1].Metadata["phase"])
	require.True(t, events[len(events)-1].Success)
}
