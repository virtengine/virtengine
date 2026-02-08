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
		Store:         store,
		AccessControl: access,
		AuditLogger:   auditLogger,
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
		Store:         store,
		AccessControl: access,
		AuditLogger:   auditLogger,
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
