package data_vault

import (
	"context"
	"encoding/json"
	"fmt"
)

// VaultAuditExporter persists audit events into the vault's audit scope.
type VaultAuditExporter struct {
	store *EncryptedBlobStore
	owner string
}

// NewVaultAuditExporter creates a vault audit exporter.
func NewVaultAuditExporter(store *EncryptedBlobStore, owner string) *VaultAuditExporter {
	if owner == "" {
		owner = "audit-system"
	}
	return &VaultAuditExporter{
		store: store,
		owner: owner,
	}
}

// Export stores the audit event as an encrypted audit blob.
func (v *VaultAuditExporter) Export(ctx context.Context, event *AuditEvent) error {
	if v == nil || v.store == nil || event == nil {
		return nil
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	tags := map[string]string{
		"event_id":  event.ID,
		"blob_id":   string(event.BlobID),
		"scope":     string(event.Scope),
		"requester": event.Requester,
		"org_id":    event.OrgID,
		"success":   fmt.Sprintf("%t", event.Success),
		"hash":      event.Hash,
		"prev_hash": event.PreviousHash,
		"type":      event.EventType,
	}

	_, err = v.store.Store(ctx, &UploadRequest{
		Scope:     ScopeAudit,
		Plaintext: payload,
		Owner:     v.owner,
		OrgID:     event.OrgID,
		Tags:      tags,
	})
	return err
}
