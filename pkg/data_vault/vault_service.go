package data_vault

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/virtengine/virtengine/pkg/data_vault/keys"
)

// VaultConfig configures the vault service.
type VaultConfig struct {
	Store           *EncryptedBlobStore
	AccessControl   AccessControl
	ConsentResolver ConsentResolver
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
	consentResolver ConsentResolver
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
	if cfg.AuditLogger == nil {
		return nil, NewVaultError("NewVault", ErrInvalidRequest, "explicit audit logger required")
	}
	if cfg.AccessControl == nil {
		cfg.AccessControl = OwnerOnlyAccessControl{}
	}
	if cfg.ConsentResolver == nil {
		cfg.ConsentResolver = DenyAllConsentResolver{}
	}
	if cfg.AuditOwner == "" {
		cfg.AuditOwner = "audit-system"
	}
	if cfg.Metrics == nil {
		cfg.Metrics = NewVaultMetrics()
	}
	if cfg.KeyRotationOverlap == 0 {
		cfg.KeyRotationOverlap = 24 * time.Hour
	}

	vault := &Vault{
		store:           cfg.Store,
		accessControl:   cfg.AccessControl,
		consentResolver: cfg.ConsentResolver,
		auditLogger:     cfg.AuditLogger,
		metrics:         cfg.Metrics,
		anomalyDetector: cfg.AnomalyDetector,
		rotationOverlap: cfg.KeyRotationOverlap,
	}
	if err := vault.ReconcilePending(ctxWithoutCancellation()); err != nil {
		return nil, fmt.Errorf("reconcile pending vault audits: %w", err)
	}
	return vault, nil
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

	operationID, err := v.beginMutationAudit(ctx, req.Scope, "", req.Owner, req.OrgID, AccessActionUpload)
	if err != nil {
		return nil, NewVaultError("Upload", err, "audit intent append failed")
	}
	mutationRequest := *req
	mutationRequest.AuditOperationID = operationID
	blob, err := v.store.Store(ctx, &mutationRequest)
	if err != nil {
		if errors.Is(err, ErrReconciliationRequired) && blob != nil {
			v.recordAccess(AccessActionUpload, req.Scope, req.Owner, true, nil)
			if auditErr := v.finishMutationAudit(ctx, operationID, req.Scope, blob.Metadata.ID, req.Owner, req.OrgID, AccessActionUpload, true, nil, map[string]string{"storage_reconciliation": "pending"}); auditErr != nil {
				return blob, &ReconciliationRequiredError{OperationID: operationID, Operation: string(AccessActionUpload), Cause: errors.Join(err, auditErr)}
			}
			return blob, err
		}
		v.recordAccess(AccessActionUpload, req.Scope, req.Owner, false, err)
		if auditErr := v.finishMutationAudit(ctx, operationID, req.Scope, "", req.Owner, req.OrgID, AccessActionUpload, false, err, nil); auditErr != nil {
			return nil, &ReconciliationRequiredError{OperationID: operationID, Operation: string(AccessActionUpload), Cause: auditErr}
		}
		return nil, err
	}

	v.recordAccess(AccessActionUpload, req.Scope, req.Owner, true, nil)
	if err := v.finishMutationAudit(ctx, operationID, req.Scope, blob.Metadata.ID, req.Owner, req.OrgID, AccessActionUpload, true, nil, map[string]string{
		"size": fmt.Sprintf("%d", blob.Metadata.Size),
	}); err != nil {
		return blob, &ReconciliationRequiredError{OperationID: operationID, Operation: string(AccessActionUpload), Cause: err}
	}

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

	if v.consentResolver != nil {
		consentOK, err := v.consentResolver.HasConsent(ctx, ConsentRequest{
			Requester: req.Requester,
			Owner:     metadata.Owner,
			Scope:     metadata.Scope,
			OrgID:     req.OrgID,
			Purpose:   req.Purpose,
			Reason:    req.Reason,
			Metadata:  req.Metadata,
		})
		if err != nil {
			v.recordAccess(AccessActionRead, metadata.Scope, req.Requester, false, err)
			v.logAudit(ctx, metadata.Scope, req.ID, req.Requester, req.OrgID, AccessActionRead, false, err, requestMetadata(req))
			return nil, nil, err
		}
		if !consentOK {
			err := NewVaultError("Retrieve", ErrConsentRequired, "consent required")
			v.recordAccess(AccessActionRead, metadata.Scope, req.Requester, false, err)
			v.logAudit(ctx, metadata.Scope, req.ID, req.Requester, req.OrgID, AccessActionRead, false, err, requestMetadata(req))
			return nil, nil, err
		}
	}

	data, meta, err := v.store.Retrieve(ctx, req.ID)
	if err != nil {
		v.recordAccess(AccessActionRead, metadata.Scope, req.Requester, false, err)
		v.logAudit(ctx, metadata.Scope, req.ID, req.Requester, req.OrgID, AccessActionRead, false, err, requestMetadata(req))
		return nil, nil, err
	}

	v.recordAccess(AccessActionRead, metadata.Scope, req.Requester, true, nil)
	if err := v.logAudit(ctx, metadata.Scope, req.ID, req.Requester, req.OrgID, AccessActionRead, true, nil, requestMetadata(req)); err != nil {
		return nil, nil, NewVaultError("Retrieve", err, "audit append failed")
	}

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

	operationID, err := v.beginMutationAudit(ctx, metadata.Scope, id, requester, metadata.OrgID, AccessActionDelete)
	if err != nil {
		return NewVaultError("Delete", err, "audit intent append failed")
	}
	if err := v.store.Delete(ctx, id); err != nil {
		if errors.Is(err, ErrReconciliationRequired) {
			v.recordAccess(AccessActionDelete, metadata.Scope, requester, true, nil)
			if auditErr := v.finishMutationAudit(ctx, operationID, metadata.Scope, id, requester, metadata.OrgID, AccessActionDelete, true, nil, map[string]string{"storage_reconciliation": "pending"}); auditErr != nil {
				return &ReconciliationRequiredError{OperationID: operationID, Operation: string(AccessActionDelete), Cause: errors.Join(err, auditErr)}
			}
			return err
		}
		v.recordAccess(AccessActionDelete, metadata.Scope, requester, false, err)
		if auditErr := v.finishMutationAudit(ctx, operationID, metadata.Scope, id, requester, metadata.OrgID, AccessActionDelete, false, err, nil); auditErr != nil {
			return &ReconciliationRequiredError{OperationID: operationID, Operation: string(AccessActionDelete), Cause: auditErr}
		}
		return err
	}

	v.recordAccess(AccessActionDelete, metadata.Scope, requester, true, nil)
	if err := v.finishMutationAudit(ctx, operationID, metadata.Scope, id, requester, metadata.OrgID, AccessActionDelete, true, nil, nil); err != nil {
		return &ReconciliationRequiredError{OperationID: operationID, Operation: string(AccessActionDelete), Cause: err}
	}

	return nil
}

// RotateKeys initiates key rotation for a scope.
func (v *Vault) RotateKeys(ctx context.Context, scope Scope, requester, orgID string) error {
	if requester == "" {
		return NewVaultError("RotateKeys", ErrInvalidRequest, "requester required")
	}
	if scope != ScopeVEID && scope != ScopeSupport && scope != ScopeMarket && scope != ScopeAudit {
		return NewVaultError("RotateKeys", ErrInvalidScope, fmt.Sprintf("unknown scope %s", scope))
	}
	if v.store == nil || v.store.KeyManager() == nil {
		return NewVaultError("RotateKeys", ErrInvalidRequest, "key manager unavailable")
	}
	logFailure := func(blobID BlobID, resourceOrgID string, failure error, metadata map[string]string) {
		v.recordAccess(AccessActionRotate, scope, requester, false, failure)
		if metadata == nil {
			metadata = make(map[string]string)
		}
		metadata["resource_org_id"] = resourceOrgID
		v.logAudit(ctx, scope, blobID, requester, orgID, AccessActionRotate, false, failure, metadata)
	}

	metadata, err := v.store.ListByScope(scope)
	if err != nil {
		logFailure("", orgID, err, nil)
		return NewVaultError("RotateKeys", err, "failed to list blobs")
	}
	nonNilMetadata := make([]*BlobMetadata, 0, len(metadata))
	for _, meta := range metadata {
		if meta == nil {
			continue
		}
		nonNilMetadata = append(nonNilMetadata, meta)
		if err := v.accessControl.Authorize(ctx, AccessRequest{
			Action:        AccessActionRotate,
			Requester:     requester,
			Owner:         meta.Owner,
			Scope:         scope,
			OrgID:         orgID,
			ResourceOrgID: meta.OrgID,
		}); err != nil {
			logFailure(meta.ID, meta.OrgID, err, nil)
			return err
		}
	}
	if len(nonNilMetadata) == 0 {
		if err := v.accessControl.Authorize(ctx, AccessRequest{
			Action:        AccessActionRotate,
			Requester:     requester,
			Scope:         scope,
			OrgID:         orgID,
			ResourceOrgID: orgID,
		}); err != nil {
			logFailure("", orgID, err, nil)
			return err
		}
	}

	keyMgr := v.store.KeyManager()
	oldKey, err := keyMgr.GetActiveKey(keys.Scope(scope))
	if err != nil {
		logFailure("", orgID, err, nil)
		return NewVaultError("RotateKeys", err, "active key unavailable")
	}

	if err := keyMgr.RotateKey(keys.Scope(scope), v.rotationOverlap); err != nil {
		logFailure("", orgID, err, map[string]string{"old_key_id": oldKey.ID})
		return NewVaultError("RotateKeys", err, "rotation failed")
	}

	rotation, err := keyMgr.GetRotationStatus(keys.Scope(scope))
	if err != nil {
		logFailure("", orgID, err, map[string]string{"old_key_id": oldKey.ID})
		return NewVaultError("RotateKeys", err, "rotation status unavailable")
	}

	newKey, err := keyMgr.GetKey(keys.Scope(scope), rotation.NewKeyID)
	if err != nil {
		logFailure("", orgID, err, map[string]string{
			"old_key_id": oldKey.ID,
			"new_key_id": rotation.NewKeyID,
		})
		return NewVaultError("RotateKeys", err, "new key unavailable")
	}

	for _, meta := range nonNilMetadata {
		_, reencryptErr := v.store.Reencrypt(ctx, meta.ID, oldKey, newKey)
		if reencryptErr != nil {
			logFailure(meta.ID, meta.OrgID, reencryptErr, map[string]string{
				"old_key_id": oldKey.ID,
				"new_key_id": newKey.ID,
			})
			return NewVaultError("RotateKeys", reencryptErr, "re-encryption failed")
		}
		v.recordAccess(AccessActionRotate, scope, requester, true, nil)
		v.logAudit(ctx, scope, meta.ID, requester, orgID, AccessActionRotate, true, nil, map[string]string{
			"old_key_id":      oldKey.ID,
			"new_key_id":      newKey.ID,
			"resource_org_id": meta.OrgID,
		})
	}

	if err := keyMgr.CompleteRotation(keys.Scope(scope)); err != nil {
		logFailure("", orgID, err, map[string]string{
			"old_key_id": oldKey.ID,
			"new_key_id": newKey.ID,
		})
		return NewVaultError("RotateKeys", err, "complete rotation failed")
	}
	if len(nonNilMetadata) == 0 {
		v.recordAccess(AccessActionRotate, scope, requester, true, nil)
		v.logAudit(ctx, scope, "", requester, orgID, AccessActionRotate, true, nil, map[string]string{
			"old_key_id":      oldKey.ID,
			"new_key_id":      newKey.ID,
			"resource_org_id": orgID,
		})
	}

	return nil
}

// ListKeyMetadata returns key metadata for a scope.
func (v *Vault) ListKeyMetadata(ctx context.Context, scope Scope, requester, orgID string) ([]KeyMetadata, error) {
	if requester == "" {
		return nil, NewVaultError("ListKeyMetadata", ErrInvalidRequest, "requester required")
	}
	if v.store == nil || v.store.KeyManager() == nil {
		return nil, NewVaultError("ListKeyMetadata", ErrInvalidRequest, "key manager unavailable")
	}
	if err := v.accessControl.Authorize(ctx, AccessRequest{
		Action:        AccessActionKeyInfo,
		Requester:     requester,
		Owner:         requester,
		Scope:         scope,
		OrgID:         orgID,
		ResourceOrgID: orgID,
	}); err != nil {
		return nil, err
	}
	keys, err := v.store.KeyManager().ListKeys(keys.Scope(scope))
	if err != nil {
		return nil, NewVaultError("ListKeyMetadata", err, "list keys failed")
	}
	metadata := make([]KeyMetadata, 0, len(keys))
	for _, key := range keys {
		if key == nil {
			continue
		}
		metadata = append(metadata, KeyMetadata{
			ID:           key.ID,
			Scope:        scope,
			Version:      key.Version,
			Status:       string(key.Status),
			CreatedAt:    key.CreatedAt,
			ActivatedAt:  key.ActivatedAt,
			DeprecatedAt: key.DeprecatedAt,
			RevokedAt:    key.RevokedAt,
		})
	}
	return metadata, nil
}

// GetKeyMetadata returns key metadata for a key ID.
func (v *Vault) GetKeyMetadata(ctx context.Context, scope Scope, keyID string, requester, orgID string) (*KeyMetadata, error) {
	if keyID == "" {
		return nil, NewVaultError("GetKeyMetadata", ErrInvalidRequest, "key id required")
	}
	if v.store == nil || v.store.KeyManager() == nil {
		return nil, NewVaultError("GetKeyMetadata", ErrInvalidRequest, "key manager unavailable")
	}
	if err := v.accessControl.Authorize(ctx, AccessRequest{
		Action:        AccessActionKeyInfo,
		Requester:     requester,
		Owner:         requester,
		Scope:         scope,
		OrgID:         orgID,
		ResourceOrgID: orgID,
	}); err != nil {
		return nil, err
	}
	key, err := v.store.KeyManager().GetKey(keys.Scope(scope), keyID)
	if err != nil {
		return nil, NewVaultError("GetKeyMetadata", err, "key not found")
	}
	return &KeyMetadata{
		ID:           key.ID,
		Scope:        scope,
		Version:      key.Version,
		Status:       string(key.Status),
		CreatedAt:    key.CreatedAt,
		ActivatedAt:  key.ActivatedAt,
		DeprecatedAt: key.DeprecatedAt,
		RevokedAt:    key.RevokedAt,
	}, nil
}

// GetAuditEvents retrieves audit events for a scope or blob.
func (v *Vault) GetAuditEvents(ctx context.Context, filter AuditFilter) ([]*AuditEvent, error) {
	if v.auditLogger == nil {
		return nil, NewVaultError("GetAuditEvents", ErrInvalidRequest, "audit logger unavailable")
	}
	return v.auditLogger.QueryEvents(ctx, filter)
}

// ReconcilePending appends terminal records for durable mutation intents whose
// process stopped after the intent or committed mutation.
func (v *Vault) ReconcilePending(ctx context.Context) error {
	if v.auditLogger == nil {
		return nil
	}
	if v.auditLogger.initErr != nil {
		return v.auditLogger.initErr
	}
	events, err := v.auditLogger.QueryEvents(ctx, AuditFilter{})
	if err != nil {
		return err
	}
	pending := make(map[string]*AuditEvent)
	for _, event := range events {
		operationID := event.Metadata["operation_id"]
		switch event.Metadata["phase"] {
		case "intent":
			pending[operationID] = event
		case "terminal":
			delete(pending, operationID)
		}
	}
	for operationID, intent := range pending {
		success := false
		blobID := intent.BlobID
		switch AccessAction(intent.EventType) {
		case AccessActionUpload:
			metadata, listErr := v.store.ListByScope(intent.Scope)
			if listErr != nil {
				return listErr
			}
			for _, candidate := range metadata {
				if candidate != nil && candidate.AuditOperationID == operationID {
					success, blobID = true, candidate.ID
					break
				}
			}
		case AccessActionDelete:
			_, getErr := v.store.GetMetadata(intent.BlobID)
			success = errors.Is(getErr, ErrBlobNotFound)
		default:
			return fmt.Errorf("unsupported pending audit operation %q", intent.EventType)
		}
		if err := v.finishMutationAudit(ctx, operationID, intent.Scope, blobID, intent.Requester, intent.OrgID, AccessAction(intent.EventType), success, nil, map[string]string{"reconciled": "true"}); err != nil {
			return err
		}
	}
	return nil
}

// Close closes the vault service and releases resources.
func (v *Vault) Close() error {
	var closeErrors []error
	if v.store != nil {
		closeErrors = append(closeErrors, v.store.Close())
	}
	if v.auditLogger != nil {
		closeErrors = append(closeErrors, v.auditLogger.Close())
	}
	return errors.Join(closeErrors...)
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

func (v *Vault) logAudit(ctx context.Context, scope Scope, blobID BlobID, requester, orgID string, action AccessAction, success bool, err error, metadata map[string]string) error {
	if v.auditLogger == nil {
		return nil
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
	if logErr := v.auditLogger.LogEvent(ctx, event); logErr != nil {
		if v.metrics != nil {
			v.metrics.AuditFailures.Inc()
		}
		return logErr
	}
	return nil
}

func (v *Vault) beginMutationAudit(ctx context.Context, scope Scope, blobID BlobID, requester, orgID string, action AccessAction) (string, error) {
	operationID, err := newAuditOperationID()
	if err != nil {
		return "", err
	}
	if err := v.logAudit(ctx, scope, blobID, requester, orgID, action, false, nil, map[string]string{
		"operation_id": operationID,
		"phase":        "intent",
	}); err != nil {
		return "", err
	}
	return operationID, nil
}

func (v *Vault) finishMutationAudit(ctx context.Context, operationID string, scope Scope, blobID BlobID, requester, orgID string, action AccessAction, success bool, operationErr error, metadata map[string]string) error {
	terminal := cloneStringMap(metadata)
	if terminal == nil {
		terminal = make(map[string]string)
	}
	terminal["operation_id"] = operationID
	terminal["phase"] = "terminal"
	return v.logAudit(ctx, scope, blobID, requester, orgID, action, success, operationErr, terminal)
}

func newAuditOperationID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(value[:]), nil
}

func ctxWithoutCancellation() context.Context { return context.Background() }

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
