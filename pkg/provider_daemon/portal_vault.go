package provider_daemon

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/virtengine/virtengine/pkg/data_vault"
)

type vaultUploadRequest struct {
	Scope           string            `json:"scope"`
	PayloadBase64   string            `json:"payload_base64"`
	ContentBase64   string            `json:"content_base64"`
	OrgID           string            `json:"org_id,omitempty"`
	RetentionPolicy string            `json:"retention_policy,omitempty"`
	ExpiresAt       *string           `json:"expires_at,omitempty"`
	Tags            map[string]string `json:"tags,omitempty"`
	Recipients      []vaultRecipient  `json:"recipients,omitempty"`
}

type vaultRecipient struct {
	PublicKeyBase64 string `json:"public_key_base64"`
	KeyID           string `json:"key_id,omitempty"`
	KeyVersion      uint32 `json:"key_version,omitempty"`
}

type vaultMetadataResponse struct {
	ID              string            `json:"id"`
	Scope           string            `json:"scope"`
	KeyID           string            `json:"key_id"`
	KeyVersion      uint32            `json:"key_version"`
	ContentHash     string            `json:"content_hash"`
	Size            int64             `json:"size"`
	EncryptedSize   int64             `json:"encrypted_size"`
	Owner           string            `json:"owner"`
	OrgID           string            `json:"org_id,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	ExpiresAt       *time.Time        `json:"expires_at,omitempty"`
	RetentionPolicy string            `json:"retention_policy,omitempty"`
	Tags            map[string]string `json:"tags,omitempty"`
	Backend         string            `json:"backend,omitempty"`
	BackendRef      string            `json:"backend_ref,omitempty"`
}

type vaultUploadResponse struct {
	Metadata vaultMetadataResponse `json:"metadata"`
}

type vaultRetrieveResponse struct {
	Data     string                `json:"data_base64"`
	Metadata vaultMetadataResponse `json:"metadata"`
}

type vaultAuditResponse struct {
	Events []data_vault.AuditEvent `json:"events"`
}

type vaultKeyMetadataResponse struct {
	ID           string     `json:"id"`
	Scope        string     `json:"scope"`
	Version      uint32     `json:"version"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	ActivatedAt  *time.Time `json:"activated_at,omitempty"`
	DeprecatedAt *time.Time `json:"deprecated_at,omitempty"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty"`
}

func (s *PortalAPIServer) handleVaultUpload(w http.ResponseWriter, r *http.Request) {
	principal := principalFromContext(r.Context())
	if principal == "" {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req vaultUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	scope, err := parseVaultScope(req.Scope)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	payloadB64 := strings.TrimSpace(req.PayloadBase64)
	if payloadB64 == "" {
		payloadB64 = strings.TrimSpace(req.ContentBase64)
	}
	payload, err := base64.StdEncoding.DecodeString(payloadB64)
	if err != nil || len(payload) == 0 {
		writeJSONError(w, http.StatusBadRequest, "invalid payload_base64")
		return
	}
	if max := s.cfg.VaultMaxPayloadBytes; max > 0 && int64(len(payload)) > max {
		writeJSONError(w, http.StatusRequestEntityTooLarge, "payload exceeds size limit")
		return
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil && strings.TrimSpace(*req.ExpiresAt) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(*req.ExpiresAt))
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid expires_at")
			return
		}
		expiresAt = &parsed
	}

	recipients := make([]data_vault.Recipient, 0, len(req.Recipients))
	for _, rec := range req.Recipients {
		if strings.TrimSpace(rec.PublicKeyBase64) == "" {
			continue
		}
		pubKey, err := base64.StdEncoding.DecodeString(strings.TrimSpace(rec.PublicKeyBase64))
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid recipient public key")
			return
		}
		recipients = append(recipients, data_vault.Recipient{
			PublicKey:  pubKey,
			KeyID:      strings.TrimSpace(rec.KeyID),
			KeyVersion: rec.KeyVersion,
		})
	}

	blob, err := s.vault.Upload(r.Context(), &data_vault.UploadRequest{
		Scope:           scope,
		Plaintext:       payload,
		Owner:           principal,
		OrgID:           strings.TrimSpace(req.OrgID),
		RetentionPolicy: strings.TrimSpace(req.RetentionPolicy),
		ExpiresAt:       expiresAt,
		Tags:            req.Tags,
		Recipients:      recipients,
	})
	if err != nil {
		if errors.Is(err, data_vault.ErrUnauthorized) {
			writeJSONError(w, http.StatusForbidden, err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, vaultUploadResponse{Metadata: vaultMetadataFrom(&blob.Metadata)})
}

func (s *PortalAPIServer) handleVaultRetrieve(w http.ResponseWriter, r *http.Request) {
	principal := principalFromContext(r.Context())
	if principal == "" {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	blobID := mux.Vars(r)["blobId"]
	if blobID == "" {
		writeJSONError(w, http.StatusBadRequest, "blob id required")
		return
	}

	orgID := strings.TrimSpace(r.URL.Query().Get("org_id"))
	purpose := strings.TrimSpace(r.URL.Query().Get("purpose"))
	reason := strings.TrimSpace(r.URL.Query().Get("reason"))

	data, metadata, err := s.vault.Retrieve(r.Context(), &data_vault.RetrieveRequest{
		ID:        data_vault.BlobID(blobID),
		Requester: principal,
		OrgID:     orgID,
		Purpose:   purpose,
		Reason:    reason,
	})
	if err != nil {
		if errors.Is(err, data_vault.ErrUnauthorized) {
			writeJSONError(w, http.StatusForbidden, err.Error())
			return
		}
		if errors.Is(err, data_vault.ErrBlobNotFound) {
			writeJSONError(w, http.StatusNotFound, "blob not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := vaultRetrieveResponse{
		Data:     base64.StdEncoding.EncodeToString(data),
		Metadata: vaultMetadataFrom(metadata),
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *PortalAPIServer) handleVaultMetadata(w http.ResponseWriter, r *http.Request) {
	principal := principalFromContext(r.Context())
	if principal == "" {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	blobID := mux.Vars(r)["blobId"]
	if blobID == "" {
		writeJSONError(w, http.StatusBadRequest, "blob id required")
		return
	}

	orgID := strings.TrimSpace(r.URL.Query().Get("org_id"))
	metadata, err := s.vault.GetMetadata(r.Context(), data_vault.BlobID(blobID), principal, orgID)
	if err != nil {
		if errors.Is(err, data_vault.ErrUnauthorized) {
			writeJSONError(w, http.StatusForbidden, err.Error())
			return
		}
		if errors.Is(err, data_vault.ErrBlobNotFound) {
			writeJSONError(w, http.StatusNotFound, "blob not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, vaultMetadataFrom(metadata))
}

func (s *PortalAPIServer) handleVaultDelete(w http.ResponseWriter, r *http.Request) {
	principal := principalFromContext(r.Context())
	if principal == "" {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	blobID := mux.Vars(r)["blobId"]
	if blobID == "" {
		writeJSONError(w, http.StatusBadRequest, "blob id required")
		return
	}

	if err := s.vault.Delete(r.Context(), data_vault.BlobID(blobID), principal); err != nil {
		if errors.Is(err, data_vault.ErrUnauthorized) {
			writeJSONError(w, http.StatusForbidden, err.Error())
			return
		}
		if errors.Is(err, data_vault.ErrBlobNotFound) {
			writeJSONError(w, http.StatusNotFound, "blob not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *PortalAPIServer) handleVaultAuditSearch(w http.ResponseWriter, r *http.Request) {
	principal := principalFromContext(r.Context())
	if principal == "" {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	filter, err := parseVaultAuditFilter(r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := s.authorizeVaultAudit(r, principal, filter); err != nil {
		writeJSONError(w, http.StatusForbidden, err.Error())
		return
	}

	events, err := s.vault.GetAuditEvents(r.Context(), filter)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := vaultAuditResponse{Events: make([]data_vault.AuditEvent, 0, len(events))}
	for _, event := range events {
		if event != nil {
			response.Events = append(response.Events, *event)
		}
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *PortalAPIServer) handleVaultKeyList(w http.ResponseWriter, r *http.Request) {
	principal := principalFromContext(r.Context())
	if principal == "" {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	scope, err := parseVaultScope(r.URL.Query().Get("scope"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	orgID := strings.TrimSpace(r.URL.Query().Get("org_id"))

	keys, err := s.vault.ListKeyMetadata(r.Context(), scope, principal, orgID)
	if err != nil {
		if errors.Is(err, data_vault.ErrUnauthorized) {
			writeJSONError(w, http.StatusForbidden, err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := make([]vaultKeyMetadataResponse, 0, len(keys))
	for _, key := range keys {
		response = append(response, vaultKeyMetadataResponse{
			ID:           key.ID,
			Scope:        string(key.Scope),
			Version:      key.Version,
			Status:       key.Status,
			CreatedAt:    key.CreatedAt,
			ActivatedAt:  key.ActivatedAt,
			DeprecatedAt: key.DeprecatedAt,
			RevokedAt:    key.RevokedAt,
		})
	}

	writeJSON(w, http.StatusOK, response)
}

func (s *PortalAPIServer) handleVaultKeyMetadata(w http.ResponseWriter, r *http.Request) {
	principal := principalFromContext(r.Context())
	if principal == "" {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	scope, err := parseVaultScope(r.URL.Query().Get("scope"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	keyID := strings.TrimSpace(mux.Vars(r)["keyId"])
	if keyID == "" {
		writeJSONError(w, http.StatusBadRequest, "key id required")
		return
	}
	orgID := strings.TrimSpace(r.URL.Query().Get("org_id"))

	key, err := s.vault.GetKeyMetadata(r.Context(), scope, keyID, principal, orgID)
	if err != nil {
		if errors.Is(err, data_vault.ErrUnauthorized) {
			writeJSONError(w, http.StatusForbidden, err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, vaultKeyMetadataResponse{
		ID:           key.ID,
		Scope:        string(key.Scope),
		Version:      key.Version,
		Status:       key.Status,
		CreatedAt:    key.CreatedAt,
		ActivatedAt:  key.ActivatedAt,
		DeprecatedAt: key.DeprecatedAt,
		RevokedAt:    key.RevokedAt,
	})
}

func vaultMetadataFrom(meta *data_vault.BlobMetadata) vaultMetadataResponse {
	if meta == nil {
		return vaultMetadataResponse{}
	}
	hash := ""
	if len(meta.ContentHash) > 0 {
		hash = hex.EncodeToString(meta.ContentHash)
	}
	return vaultMetadataResponse{
		ID:              string(meta.ID),
		Scope:           string(meta.Scope),
		KeyID:           meta.KeyID,
		KeyVersion:      meta.KeyVersion,
		ContentHash:     hash,
		Size:            meta.Size,
		EncryptedSize:   meta.EncryptedSize,
		Owner:           meta.Owner,
		OrgID:           meta.OrgID,
		CreatedAt:       meta.CreatedAt,
		ExpiresAt:       meta.ExpiresAt,
		RetentionPolicy: meta.RetentionPolicy,
		Tags:            meta.Tags,
		Backend:         meta.Backend,
		BackendRef:      meta.BackendRef,
	}
}

func parseVaultScope(scope string) (data_vault.Scope, error) {
	switch strings.ToLower(strings.TrimSpace(scope)) {
	case string(data_vault.ScopeVEID):
		return data_vault.ScopeVEID, nil
	case string(data_vault.ScopeSupport):
		return data_vault.ScopeSupport, nil
	case string(data_vault.ScopeMarket):
		return data_vault.ScopeMarket, nil
	case string(data_vault.ScopeAudit):
		return data_vault.ScopeAudit, nil
	default:
		return "", errors.New("invalid scope")
	}
}

func parseVaultAuditFilter(r *http.Request) (data_vault.AuditFilter, error) {
	query := r.URL.Query()
	filter := data_vault.AuditFilter{
		BlobID:    data_vault.BlobID(strings.TrimSpace(query.Get("blob_id"))),
		Scope:     data_vault.Scope(strings.TrimSpace(query.Get("scope"))),
		Requester: strings.TrimSpace(query.Get("requester")),
		OrgID:     strings.TrimSpace(query.Get("org_id")),
	}
	if raw := strings.TrimSpace(query.Get("start")); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return filter, errors.New("invalid start")
		}
		filter.StartTime = &parsed
	}
	if raw := strings.TrimSpace(query.Get("end")); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return filter, errors.New("invalid end")
		}
		filter.EndTime = &parsed
	}
	if raw := strings.TrimSpace(query.Get("limit")); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil {
			return filter, errors.New("invalid limit")
		}
		filter.Limit = limit
	}
	return filter, nil
}

func (s *PortalAPIServer) authorizeVaultAudit(r *http.Request, principal string, filter data_vault.AuditFilter) error {
	if filter.OrgID == "" && filter.Requester == principal {
		return nil
	}
	if filter.OrgID == "" {
		return errors.New("org_id required")
	}
	return s.authorizeVaultOrg(r, principal, filter.OrgID)
}

func (s *PortalAPIServer) authorizeVaultOrg(r *http.Request, principal, orgID string) error {
	if s.chainQuery == nil {
		return errors.New("organization access unavailable")
	}
	isAdmin, err := s.chainQuery.IsOrganizationAdmin(r.Context(), orgID, principal)
	if err != nil {
		return err
	}
	if !isAdmin {
		return errors.New("not authorized for organization")
	}
	return nil
}
