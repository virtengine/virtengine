package data_vault

import (
	"context"
	"testing"

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
