package data_vault

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/virtengine/virtengine/pkg/data_vault/keys"
)

// VaultConfig configures the vault service.
type VaultConfig struct {
	Store           *EncryptedBlobStore
	AccessControl   AccessControl
	AuditLogger     *AuditLogger
	AuditOwner      string
	Metrics         *VaultMetrics
	AnomalyDetector *AccessAnomalyDetector

	// KeyRotationOverlap defines the overlap duration for rotations.
	KeyRotationOverlap time.Duration
}

// Vault implements VaultService.
type Vault struct {
	store           *EncryptedBlobStore
	accessControl   AccessControl
	auditLogger     *AuditLogger
	metrics         *VaultMetrics
	anomalyDetector *AccessAnomalyDetector
	rotationOverlap time.Duration
}

// NewVaultService creates a new vault service.
func NewVaultService(cfg VaultConfig) (*Vault, error) {
	if cfg.Store == nil {
		return nil, NewVaultError("NewVault", ErrInvalidRequest, "store required")
	}
	if cfg.AccessControl == nil {
		cfg.AccessControl = OwnerOnlyAccessControl{}
	}
	if cfg.AuditOwner == "" {
		cfg.AuditOwner = "audit-system"
	}
	if cfg.AuditLogger == nil {
		cfg.AuditLogger = NewAuditLogger(DefaultAuditLogConfig(), nil)
		cfg.AuditLogger.RegisterExporter(NewVaultAuditExporter(cfg.Store, cfg.AuditOwner))
	}
	if cfg.Metrics == nil {
		cfg.Metrics = NewVaultMetrics()
	}
	if cfg.KeyRotationOverlap == 0 {
		cfg.KeyRotationOverlap = 24 * time.Hour
	}

	return &Vault{
		store:           cfg.Store,
		accessControl:   cfg.AccessControl,
		auditLogger:     cfg.AuditLogger,
		metrics:         cfg.Metrics,
		anomalyDetector: cfg.AnomalyDetector,
		rotationOverlap: cfg.KeyRotationOverlap,
	}, nil
}

// NewVault creates a new vault service (alias for NewVaultService).
func NewVault(cfg VaultConfig) (*Vault, error) {
	return NewVaultService(cfg)
}

// Upload encrypts and stores a new blob.
func (v *Vault) Upload(ctx context.Context, req *UploadRequest) (*EncryptedBlob, error) {
	if req == nil {
		return nil, NewVaultError("Upload", ErrInvalidRequest, "request required")
	}
	if req.Owner == "" {
		return nil, NewVaultError("Upload", ErrInvalidRequest, "owner required")
	}

	if err := v.accessControl.Authorize(ctx, AccessRequest{
		Action:    AccessActionUpload,
		Requester: req.Owner,
		Owner:     req.Owner,
		Scope:     req.Scope,
		OrgID:     req.OrgID,
	}); err != nil {
		v.recordAccess(AccessActionUpload, req.Scope, req.Owner, false, err)
		v.logAudit(ctx, req.Scope, "", req.Owner, req.OrgID, AccessActionUpload, false, err, nil)
		return nil, err
	}

	blob, err := v.store.Store(ctx, req)
	if err != nil {
		v.recordAccess(AccessActionUpload, req.Scope, req.Owner, false, err)
		v.logAudit(ctx, req.Scope, "", req.Owner, req.OrgID, AccessActionUpload, false, err, nil)
		return nil, err
	}

	v.recordAccess(AccessActionUpload, req.Scope, req.Owner, true, nil)
	v.logAudit(ctx, req.Scope, blob.Metadata.ID, req.Owner, req.OrgID, AccessActionUpload, true, nil, map[string]string{
		"size": fmt.Sprintf("%d", blob.Metadata.Size),
	})

	return blob, nil
}

// Retrieve retrieves and decrypts a blob.
func (v *Vault) Retrieve(ctx context.Context, req *RetrieveRequest) ([]byte, *BlobMetadata, error) {
	if req == nil {
		return nil, nil, NewVaultError("Retrieve", ErrInvalidRequest, "request required")
	}
	if req.Requester == "" {
		return nil, nil, NewVaultError("Retrieve", ErrInvalidRequest, "requester required")
	}

	metadata, err := v.store.GetMetadata(req.ID)
	if err != nil {
		v.recordAccess(AccessActionRead, "", req.Requester, false, err)
		v.logAudit(ctx, "", req.ID, req.Requester, req.OrgID, AccessActionRead, false, err, requestMetadata(req))
		return nil, nil, err
	}

	if err := v.accessControl.Authorize(ctx, AccessRequest{
		Action:        AccessActionRead,
		Requester:     req.Requester,
		Owner:         metadata.Owner,
		Scope:         metadata.Scope,
		OrgID:         req.OrgID,
		ResourceOrgID: metadata.OrgID,
	}); err != nil {
		v.recordAccess(AccessActionRead, metadata.Scope, req.Requester, false, err)
		if v.anomalyDetector != nil {
			v.anomalyDetector.RecordFailure(req.Requester)
		}
		v.logAudit(ctx, metadata.Scope, req.ID, req.Requester, req.OrgID, AccessActionRead, false, err, requestMetadata(req))
		return nil, nil, err
	}

	data, meta, err := v.store.Retrieve(ctx, req.ID)
	if err != nil {
		v.recordAccess(AccessActionRead, metadata.Scope, req.Requester, false, err)
		v.logAudit(ctx, metadata.Scope, req.ID, req.Requester, req.OrgID, AccessActionRead, false, err, requestMetadata(req))
		return nil, nil, err
	}

	v.recordAccess(AccessActionRead, metadata.Scope, req.Requester, true, nil)
	v.logAudit(ctx, metadata.Scope, req.ID, req.Requester, req.OrgID, AccessActionRead, true, nil, requestMetadata(req))

	return data, meta, nil
}

// RetrieveStream retrieves and decrypts a blob as a stream.
func (v *Vault) RetrieveStream(ctx context.Context, req *RetrieveRequest) (io.ReadCloser, *BlobMetadata, error) {
	data, meta, err := v.Retrieve(ctx, req)
	if err != nil {
		return nil, nil, err
	}
	return io.NopCloser(bytes.NewReader(data)), meta, nil
}

// GetMetadata retrieves blob metadata without decrypting content.
func (v *Vault) GetMetadata(ctx context.Context, id BlobID, requester string, orgID string) (*BlobMetadata, error) {
	if requester == "" {
		return nil, NewVaultError("GetMetadata", ErrInvalidRequest, "requester required")
	}

	metadata, err := v.store.GetMetadata(id)
	if err != nil {
		v.recordAccess(AccessActionMetadata, "", requester, false, err)
		v.logAudit(ctx, "", id, requester, orgID, AccessActionMetadata, false, err, nil)
		return nil, err
	}

	if err := v.accessControl.Authorize(ctx, AccessRequest{
		Action:        AccessActionMetadata,
		Requester:     requester,
		Owner:         metadata.Owner,
		Scope:         metadata.Scope,
		OrgID:         orgID,
		ResourceOrgID: metadata.OrgID,
	}); err != nil {
		v.recordAccess(AccessActionMetadata, metadata.Scope, requester, false, err)
		v.logAudit(ctx, metadata.Scope, id, requester, orgID, AccessActionMetadata, false, err, nil)
		return nil, err
	}

	v.recordAccess(AccessActionMetadata, metadata.Scope, requester, true, nil)
	v.logAudit(ctx, metadata.Scope, id, requester, orgID, AccessActionMetadata, true, nil, nil)
	return metadata, nil
}

// Delete marks a blob for deletion.
func (v *Vault) Delete(ctx context.Context, id BlobID, requester string) error {
	if requester == "" {
		return NewVaultError("Delete", ErrInvalidRequest, "requester required")
	}
	metadata, err := v.store.GetMetadata(id)
	if err != nil {
		v.recordAccess(AccessActionDelete, "", requester, false, err)
		v.logAudit(ctx, "", id, requester, "", AccessActionDelete, false, err, nil)
		return err
	}

	if err := v.accessControl.Authorize(ctx, AccessRequest{
		Action:        AccessActionDelete,
		Requester:     requester,
		Owner:         metadata.Owner,
		Scope:         metadata.Scope,
		OrgID:         metadata.OrgID,
		ResourceOrgID: metadata.OrgID,
	}); err != nil {
		v.recordAccess(AccessActionDelete, metadata.Scope, requester, false, err)
		v.logAudit(ctx, metadata.Scope, id, requester, metadata.OrgID, AccessActionDelete, false, err, nil)
		return err
	}

	if err := v.store.Delete(ctx, id); err != nil {
		v.recordAccess(AccessActionDelete, metadata.Scope, requester, false, err)
		v.logAudit(ctx, metadata.Scope, id, requester, metadata.OrgID, AccessActionDelete, false, err, nil)
		return err
	}

	v.recordAccess(AccessActionDelete, metadata.Scope, requester, true, nil)
	v.logAudit(ctx, metadata.Scope, id, requester, metadata.OrgID, AccessActionDelete, true, nil, nil)

	return nil
}

// RotateKeys initiates key rotation for a scope.
func (v *Vault) RotateKeys(ctx context.Context, scope Scope) error {
	_ = ctx
	if v.store == nil || v.store.KeyManager() == nil {
		return NewVaultError("RotateKeys", ErrInvalidRequest, "key manager unavailable")
	}
	if err := v.store.KeyManager().RotateKey(keys.Scope(scope), v.rotationOverlap); err != nil {
		return NewVaultError("RotateKeys", err, "rotation failed")
	}
	return nil
}

// GetAuditEvents retrieves audit events for a scope or blob.
func (v *Vault) GetAuditEvents(ctx context.Context, filter AuditFilter) ([]*AuditEvent, error) {
	if v.auditLogger == nil {
		return nil, NewVaultError("GetAuditEvents", ErrInvalidRequest, "audit logger unavailable")
	}
	return v.auditLogger.QueryEvents(ctx, filter)
}

// Close closes the vault service and releases resources.
func (v *Vault) Close() error {
	if v.store != nil {
		return v.store.Close()
	}
	return nil
}

func (v *Vault) recordAccess(action AccessAction, scope Scope, requester string, success bool, err error) {
	if v.metrics == nil {
		return
	}
	if scope == "" {
		scope = "unknown"
	}
	result := "false"
	if success {
		result = "true"
	}
	v.metrics.AccessTotal.WithLabelValues(string(scope), string(action), result).Inc()
	if !success {
		v.metrics.AccessDeniedTotal.WithLabelValues(string(scope), string(action)).Inc()
	}
	if err != nil && !success && v.anomalyDetector != nil {
		v.anomalyDetector.RecordFailure(requester)
	}
}

func (v *Vault) logAudit(ctx context.Context, scope Scope, blobID BlobID, requester, orgID string, action AccessAction, success bool, err error, metadata map[string]string) {
	if v.auditLogger == nil {
		return
	}
	event := &AuditEvent{
		EventType: string(action),
		BlobID:    blobID,
		Scope:     scope,
		Requester: requester,
		OrgID:     orgID,
		Success:   success,
		Metadata:  metadata,
	}
	if err != nil {
		event.Error = err.Error()
	}
	if logErr := v.auditLogger.LogEvent(ctx, event); logErr != nil && v.metrics != nil {
		v.metrics.AuditFailures.Inc()
	}
}

func requestMetadata(req *RetrieveRequest) map[string]string {
	if req == nil {
		return nil
	}
	metadata := map[string]string{}
	if req.Purpose != "" {
		metadata["purpose"] = req.Purpose
	}
	if req.Reason != "" {
		metadata["reason"] = req.Reason
	}
	for k, v := range req.Metadata {
		metadata[k] = v
	}
	if len(metadata) == 0 {
		return nil
	}
	return metadata
}
