package data_vault

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/virtengine/virtengine/pkg/data_vault/keys"
)

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
	require.NoError(t, vault.RotateKeys(ctx, ScopeSupport))

	_, meta, err := vault.Retrieve(ctx, &RetrieveRequest{
		ID:        blob.Metadata.ID,
		Requester: "owner",
		OrgID:     "org-1",
	})
	require.NoError(t, err)
	require.Greater(t, meta.KeyVersion, originalVersion)
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
