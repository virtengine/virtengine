package servicedesk

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"math"
	"time"

	"cosmossdk.io/log"

	"github.com/virtengine/virtengine/pkg/artifact_store"
	"github.com/virtengine/virtengine/pkg/data_vault"
)

// AttachmentHandler handles attachment synchronization via the artifact store
type AttachmentHandler struct {
	store  artifact_store.ArtifactStore
	vault  data_vault.VaultService
	config AttachmentConfig
	logger log.Logger
}

// AttachmentConfig holds attachment handling configuration
type AttachmentConfig struct {
	// MaxSize is the maximum attachment size in bytes
	MaxSize int64 `json:"max_size"`

	// AllowedTypes are allowed MIME types
	AllowedTypes []string `json:"allowed_types"`

	// AccessTokenTTL is the TTL for temporary access tokens
	AccessTokenTTL time.Duration `json:"access_token_ttl"`

	// EncryptionRequired requires attachments to be encrypted
	EncryptionRequired bool `json:"encryption_required"`

	// VaultService enables storing attachments in the data vault when set
	VaultService data_vault.VaultService `json:"-"`
}

// DefaultAttachmentConfig returns default attachment configuration
func DefaultAttachmentConfig() AttachmentConfig {
	return AttachmentConfig{
		MaxSize: 10 * 1024 * 1024, // 10 MB
		AllowedTypes: []string{
			"image/jpeg",
			"image/png",
			"image/gif",
			"application/pdf",
			"text/plain",
			"application/json",
			"application/zip",
		},
		AccessTokenTTL:     1 * time.Hour,
		EncryptionRequired: true,
	}
}

// NewAttachmentHandler creates a new attachment handler
func NewAttachmentHandler(store artifact_store.ArtifactStore, config AttachmentConfig, logger log.Logger) *AttachmentHandler {
	return &AttachmentHandler{
		store:  store,
		vault:  config.VaultService,
		config: config,
		logger: logger.With("component", "attachments"),
	}
}

// UploadAttachment uploads an attachment to the artifact store
func (h *AttachmentHandler) UploadAttachment(ctx context.Context, req *UploadAttachmentRequest) (*UploadAttachmentResponse, error) {
	if err := h.validateUpload(req); err != nil {
		return nil, err
	}

	// Read the attachment data
	data, err := io.ReadAll(req.Reader)
	if err != nil {
		return nil, ErrAttachmentFailed.Wrapf("failed to read attachment: %v", err)
	}

	// Check size
	if int64(len(data)) > h.config.MaxSize {
		return nil, ErrAttachmentFailed.Wrapf("attachment exceeds max size of %d bytes", h.config.MaxSize)
	}

	if h.vault != nil {
		blob, err := h.vault.Upload(ctx, &data_vault.UploadRequest{
			Scope:     data_vault.ScopeSupport,
			Plaintext: data,
			Owner:     req.Owner,
			OrgID:     req.OrgID,
			Tags: map[string]string{
				"ticket_id":    req.TicketID,
				"file_name":    req.FileName,
				"content_type": req.ContentType,
			},
		})
		if err != nil {
			return nil, ErrAttachmentFailed.Wrapf("failed to store attachment in vault: %v", err)
		}

		accessToken, expiresAt := h.generateAccessToken()

		h.logger.Debug("attachment uploaded via vault",
			"ticket_id", req.TicketID,
			"file_name", req.FileName,
			"size", len(data),
			"blob_id", blob.Metadata.ID,
		)

		return &UploadAttachmentResponse{
			ArtifactAddress: string(blob.Metadata.ID),
			VaultBlobID:     string(blob.Metadata.ID),
			AccessToken:     accessToken,
			ExpiresAt:       expiresAt,
		}, nil
	}

	// Compute content hash
	contentHash := sha256.Sum256(data)

	// Store in artifact store
	putReq := &artifact_store.PutRequest{
		Data:         data,
		ContentHash:  contentHash[:],
		Owner:        req.Owner,
		ArtifactType: "support_attachment",
		Metadata: map[string]string{
			"ticket_id":    req.TicketID,
			"file_name":    req.FileName,
			"content_type": req.ContentType,
		},
		EncryptionMetadata: &artifact_store.EncryptionMetadata{
			AlgorithmID:     "X25519-XSalsa20-Poly1305",
			RecipientKeyIDs: []string{req.EncryptionKeyID},
		},
	}

	resp, err := h.store.Put(ctx, putReq)
	if err != nil {
		return nil, ErrAttachmentFailed.Wrapf("failed to store attachment: %v", err)
	}

	// Generate access token
	accessToken, expiresAt := h.generateAccessToken()

	h.logger.Debug("attachment uploaded",
		"ticket_id", req.TicketID,
		"file_name", req.FileName,
		"size", len(data),
		"content_address", resp.ContentAddress.HashHex(),
	)

	return &UploadAttachmentResponse{
		ArtifactAddress: resp.ContentAddress.HashHex(),
		AccessToken:     accessToken,
		ExpiresAt:       expiresAt,
	}, nil
}

// GetAttachment retrieves an attachment from the artifact store
func (h *AttachmentHandler) GetAttachment(ctx context.Context, req *GetAttachmentRequest) (*GetAttachmentResponse, error) {
	if h.vault != nil && req.VaultBlobID != "" {
		data, meta, err := h.vault.Retrieve(ctx, &data_vault.RetrieveRequest{
			ID:        data_vault.BlobID(req.VaultBlobID),
			Requester: req.Requester,
			OrgID:     req.OrgID,
			Purpose:   "support_attachment",
			Reason:    "support_ticket_access",
		})
		if err != nil {
			return nil, ErrAttachmentFailed.Wrapf("failed to retrieve attachment from vault: %v", err)
		}

		h.logger.Debug("attachment retrieved via vault",
			"blob_id", req.VaultBlobID,
		)

		return &GetAttachmentResponse{
			Data:        data,
			ContentType: meta.Tags["content_type"],
			VaultBlobID: string(meta.ID),
		}, nil
	}

	// Parse content address from hex string
	hashBytes, err := hex.DecodeString(req.ArtifactAddress)
	if err != nil {
		return nil, ErrAttachmentFailed.Wrapf("invalid artifact address: %v", err)
	}

	// Create content address from hash
	contentAddr := artifact_store.NewContentAddressFromHash(
		hashBytes,
		0, // Size unknown from hash alone
		artifact_store.BackendWaldur,
		req.ArtifactAddress,
	)

	// Verify access token (in production, would validate cryptographically)
	if req.AccessToken == "" {
		return nil, ErrAttachmentFailed.Wrap("access token required")
	}

	// Retrieve from artifact store
	getReq := &artifact_store.GetRequest{
		ContentAddress:    contentAddr,
		RequestingAccount: req.Requester,
		AuthToken:         req.AccessToken,
	}

	resp, err := h.store.Get(ctx, getReq)
	if err != nil {
		return nil, ErrAttachmentFailed.Wrapf("failed to retrieve attachment: %v", err)
	}

	h.logger.Debug("attachment retrieved",
		"artifact_address", req.ArtifactAddress,
	)

	return &GetAttachmentResponse{
		Data:        resp.Data,
		ContentType: "", // Would be stored in metadata
	}, nil
}

// SyncAttachmentToExternal syncs an attachment to an external service desk
func (h *AttachmentHandler) SyncAttachmentToExternal(ctx context.Context, req *AttachmentSyncRequest) (*AttachmentSyncResponse, error) {
	h.logger.Debug("syncing attachment to external",
		"ticket_id", req.TicketID,
		"artifact_address", req.ArtifactAddress,
		"service_desk", req.ServiceDeskType,
	)

	// Generate temporary access token for external system
	accessToken, expiresAt := h.generateAccessToken()

	// In a full implementation, this would:
	// 1. Generate a time-limited signed URL for the attachment
	// 2. Call the external API to upload/link the attachment
	// 3. Track the external attachment ID

	return &AttachmentSyncResponse{
		AttachmentSync: AttachmentSync{
			TicketID:        req.TicketID,
			ArtifactAddress: req.ArtifactAddress,
			FileName:        req.FileName,
			ContentType:     req.ContentType,
			Size:            req.Size,
			AccessToken:     accessToken,
			ExpiresAt:       &expiresAt,
		},
		SyncStatus: SyncStatusPending,
	}, nil
}

// DeleteAttachment deletes an attachment from the artifact store
func (h *AttachmentHandler) DeleteAttachment(ctx context.Context, req *DeleteAttachmentRequest) error {
	if h.vault != nil && req.VaultBlobID != "" {
		if err := h.vault.Delete(ctx, data_vault.BlobID(req.VaultBlobID), req.Requester); err != nil {
			return ErrAttachmentFailed.Wrapf("failed to delete vault attachment: %v", err)
		}
		h.logger.Debug("attachment deleted via vault",
			"blob_id", req.VaultBlobID,
		)
		return nil
	}

	// Parse content address from hex string
	hashBytes, err := hex.DecodeString(req.ArtifactAddress)
	if err != nil {
		return ErrAttachmentFailed.Wrapf("invalid artifact address: %v", err)
	}

	// Create content address from hash
	contentAddr := artifact_store.NewContentAddressFromHash(
		hashBytes,
		0, // Size unknown from hash alone
		artifact_store.BackendWaldur,
		req.ArtifactAddress,
	)

	// Delete from artifact store
	delReq := &artifact_store.DeleteRequest{
		ContentAddress:    contentAddr,
		RequestingAccount: req.Requester,
		Force:             req.Force,
	}

	if err := h.store.Delete(ctx, delReq); err != nil {
		return ErrAttachmentFailed.Wrapf("failed to delete attachment: %v", err)
	}

	h.logger.Debug("attachment deleted",
		"artifact_address", req.ArtifactAddress,
	)

	return nil
}

// ListAttachments lists attachments for a ticket
func (h *AttachmentHandler) ListAttachments(ctx context.Context, ticketID string, owner string) ([]*AttachmentInfo, error) {
	// List from artifact store by owner
	listResp, err := h.store.ListByOwner(ctx, owner, &artifact_store.Pagination{
		Limit: 100,
	})
	if err != nil {
		return nil, ErrAttachmentFailed.Wrapf("failed to list attachments: %v", err)
	}

	// Filter by ticket ID
	var attachments []*AttachmentInfo
	for _, ref := range listResp.References {
		if ref.Metadata["ticket_id"] == ticketID {
			attachments = append(attachments, &AttachmentInfo{
				ArtifactAddress: ref.ContentAddress.HashHex(),
				FileName:        ref.Metadata["file_name"],
				ContentType:     ref.Metadata["content_type"],
				Size:            safeInt64FromUint64(ref.ContentAddress.Size),
				CreatedAt:       ref.CreatedAt,
			})
		}
	}

	return attachments, nil
}

// validateUpload validates an upload request
func (h *AttachmentHandler) validateUpload(req *UploadAttachmentRequest) error {
	if req.TicketID == "" {
		return ErrAttachmentFailed.Wrap("ticket_id is required")
	}
	if req.FileName == "" {
		return ErrAttachmentFailed.Wrap("file_name is required")
	}
	if req.Owner == "" {
		return ErrAttachmentFailed.Wrap("owner is required")
	}
	if req.Reader == nil {
		return ErrAttachmentFailed.Wrap("reader is required")
	}

	// Check content type
	if req.ContentType != "" && len(h.config.AllowedTypes) > 0 {
		allowed := false
		for _, t := range h.config.AllowedTypes {
			if t == req.ContentType {
				allowed = true
				break
			}
		}
		if !allowed {
			return ErrAttachmentFailed.Wrapf("content type %s not allowed", req.ContentType)
		}
	}

	// Check encryption requirement
	if h.config.EncryptionRequired && h.vault == nil && req.EncryptionKeyID == "" {
		return ErrAttachmentFailed.Wrap("encryption key required")
	}

	return nil
}

// generateAccessToken generates a temporary access token
func (h *AttachmentHandler) generateAccessToken() (string, time.Time) {
	token := make([]byte, 32)
	_, _ = rand.Read(token)
	expiresAt := time.Now().Add(h.config.AccessTokenTTL)
	return hex.EncodeToString(token), expiresAt
}

func safeInt64FromUint64(value uint64) int64 {
	if value > math.MaxInt64 {
		return math.MaxInt64
	}
	return int64(value)
}

// UploadAttachmentRequest contains parameters for uploading an attachment
type UploadAttachmentRequest struct {
	// TicketID is the ticket to attach to
	TicketID string

	// FileName is the original file name
	FileName string

	// ContentType is the MIME content type
	ContentType string

	// Owner is the account that owns the attachment
	Owner string

	// Requester is the account uploading the attachment
	Requester string

	// OrgID is the organization ID for access control
	OrgID string

	// Reader provides the attachment data
	Reader io.Reader

	// EncryptionKeyID is the key ID for encryption
	EncryptionKeyID string
}

// UploadAttachmentResponse contains the upload result
type UploadAttachmentResponse struct {
	// ArtifactAddress is the content-addressed reference
	ArtifactAddress string

	// VaultBlobID is set when stored in the data vault
	VaultBlobID string `json:"vault_blob_id,omitempty"`

	// AccessToken is a temporary access token
	AccessToken string

	// ExpiresAt is when the access token expires
	ExpiresAt time.Time
}

// GetAttachmentRequest contains parameters for retrieving an attachment
type GetAttachmentRequest struct {
	// ArtifactAddress is the content address (hex-encoded hash)
	ArtifactAddress string

	// VaultBlobID is the data vault blob ID (if stored in vault)
	VaultBlobID string

	// AccessToken is the access token
	AccessToken string

	// Requester is the account requesting access
	Requester string

	// OrgID is the organization ID for access control
	OrgID string
}

// GetAttachmentResponse contains the retrieved attachment
type GetAttachmentResponse struct {
	// Data is the attachment data (encrypted)
	Data []byte

	// ContentType is the MIME content type
	ContentType string

	// VaultBlobID is included when retrieved from vault
	VaultBlobID string `json:"vault_blob_id,omitempty"`
}

// DeleteAttachmentRequest contains parameters for deleting an attachment
type DeleteAttachmentRequest struct {
	// ArtifactAddress is the content address (hex-encoded hash)
	ArtifactAddress string

	// VaultBlobID is the data vault blob ID (if stored in vault)
	VaultBlobID string

	// Requester is the account requesting deletion
	Requester string

	// Force forces deletion even if not expired
	Force bool
}

// AttachmentSyncRequest contains parameters for syncing an attachment
type AttachmentSyncRequest struct {
	// TicketID is the on-chain ticket ID
	TicketID string

	// ArtifactAddress is the artifact store address
	ArtifactAddress string

	// FileName is the file name
	FileName string

	// ContentType is the MIME type
	ContentType string

	// Size is the file size
	Size int64

	// ServiceDeskType is the target service desk
	ServiceDeskType ServiceDeskType

	// ExternalTicketID is the external ticket ID
	ExternalTicketID string
}

// AttachmentSyncResponse contains the sync result
type AttachmentSyncResponse struct {
	// AttachmentSync is the sync record
	AttachmentSync AttachmentSync

	// SyncStatus is the status
	SyncStatus SyncStatus

	// Error is any error
	Error string
}

// AttachmentInfo contains attachment metadata
type AttachmentInfo struct {
	// ArtifactAddress is the content address (hex-encoded hash)
	ArtifactAddress string `json:"artifact_address"`

	// FileName is the original file name
	FileName string `json:"file_name"`

	// ContentType is the MIME type
	ContentType string `json:"content_type"`

	// Size is the file size
	Size int64 `json:"size"`

	// CreatedAt is when the attachment was created
	CreatedAt time.Time `json:"created_at"`

	// ExternalRefs are external attachment references
	ExternalRefs []ExternalAttachmentRef `json:"external_refs,omitempty"`
}

// ExternalAttachmentRef references an external attachment
type ExternalAttachmentRef struct {
	// ServiceDeskType is the service desk type
	ServiceDeskType ServiceDeskType `json:"service_desk_type"`

	// ExternalID is the external attachment ID
	ExternalID string `json:"external_id"`

	// ExternalURL is the URL to the external attachment
	ExternalURL string `json:"external_url,omitempty"`

	// SyncedAt is when the attachment was synced
	SyncedAt time.Time `json:"synced_at"`
}
