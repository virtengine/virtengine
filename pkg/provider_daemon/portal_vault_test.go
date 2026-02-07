package provider_daemon

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/virtengine/virtengine/pkg/artifact_store"
	"github.com/virtengine/virtengine/pkg/data_vault"
	"github.com/virtengine/virtengine/pkg/data_vault/keys"
	portalauth "github.com/virtengine/virtengine/pkg/provider_daemon/auth"
)

func TestPortalVaultUploadInvalidBody(t *testing.T) {
	srv := testPortalServer(t)

	req := httptest.NewRequest(http.MethodPost, "/vault/blobs", bytes.NewBufferString("{invalid}"))
	req = req.WithContext(portalauth.WithAuth(req.Context(), portalauth.AuthContext{Address: "addr1"}))
	rec := httptest.NewRecorder()

	srv.handleVaultUpload(rec, req)

	if rec.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Result().StatusCode)
	}
}

func TestPortalVaultRetrieveNotFound(t *testing.T) {
	srv := testPortalServer(t)

	req := httptest.NewRequest(http.MethodGet, "/vault/blobs/missing", nil)
	req = mux.SetURLVars(req, map[string]string{"blobId": "missing"})
	req = req.WithContext(portalauth.WithAuth(req.Context(), portalauth.AuthContext{Address: "addr1"}))
	rec := httptest.NewRecorder()

	srv.handleVaultRetrieve(rec, req)

	if rec.Result().StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Result().StatusCode)
	}
}

func TestPortalVaultAuditRequiresOrg(t *testing.T) {
	srv := testPortalServer(t)

	req := httptest.NewRequest(http.MethodGet, "/vault/audit?requester=someone", nil)
	req = req.WithContext(portalauth.WithAuth(req.Context(), portalauth.AuthContext{Address: "addr1"}))
	rec := httptest.NewRecorder()

	srv.handleVaultAuditSearch(rec, req)

	if rec.Result().StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Result().StatusCode)
	}
}

func TestPortalVaultUploadAndRetrieve(t *testing.T) {
	srv := testPortalServer(t)

	payload := []byte("secret")
	body := map[string]string{
		"scope":          "support",
		"content_base64": base64.StdEncoding.EncodeToString(payload),
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/vault/blobs", bytes.NewReader(raw))
	req = req.WithContext(portalauth.WithAuth(req.Context(), portalauth.AuthContext{Address: "addr1"}))
	rec := httptest.NewRecorder()

	srv.handleVaultUpload(rec, req)
	if rec.Result().StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Result().StatusCode)
	}

	var uploadResp struct {
		Metadata struct {
			ID string `json:"id"`
		} `json:"metadata"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&uploadResp)
	if uploadResp.Metadata.ID == "" {
		t.Fatalf("expected blob id in response")
	}

	getReq := httptest.NewRequest(http.MethodGet, "/vault/blobs/"+uploadResp.Metadata.ID, nil)
	getReq = mux.SetURLVars(getReq, map[string]string{"blobId": uploadResp.Metadata.ID})
	getReq = getReq.WithContext(portalauth.WithAuth(getReq.Context(), portalauth.AuthContext{Address: "addr1"}))
	getRec := httptest.NewRecorder()

	srv.handleVaultRetrieve(getRec, getReq)
	if getRec.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", getRec.Result().StatusCode)
	}
}

func testPortalServer(t *testing.T) *PortalAPIServer {
	t.Helper()
	vault, err := newTestVaultService()
	if err != nil {
		t.Fatalf("vault service: %v", err)
	}
	srv, err := NewPortalAPIServer(PortalAPIServerConfig{
		VaultService: vault,
	})
	if err != nil {
		t.Fatalf("portal server: %v", err)
	}
	return srv
}

func newTestVaultService() (data_vault.VaultService, error) {
	backend := artifact_store.NewMemoryBackend()
	keyMgr := keys.NewKeyManager()
	if err := keyMgr.Initialize(); err != nil {
		return nil, err
	}
	store := data_vault.NewEncryptedBlobStore(backend, keyMgr)
	auditLogger := data_vault.NewAuditLogger(data_vault.DefaultAuditLogConfig(), nil)
	auditLogger.RegisterExporter(data_vault.NewVaultAuditExporter(store, "audit-system"))
	metrics := data_vault.NewVaultMetrics()
	access := data_vault.NewPolicyAccessControl(data_vault.DefaultAccessPolicy(), nil, data_vault.StaticOrgResolver{})

	return data_vault.NewVaultService(data_vault.VaultConfig{
		Store:              store,
		AccessControl:      access,
		AuditLogger:        auditLogger,
		AuditOwner:         "audit-system",
		Metrics:            metrics,
		KeyRotationOverlap: 24 * time.Hour,
	})
}
