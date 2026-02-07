package provider_daemon

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/virtengine/virtengine/pkg/data_vault"
)

type stubVault struct{}

func (stubVault) Upload(_ context.Context, req *data_vault.UploadRequest) (*data_vault.EncryptedBlob, error) {
	return &data_vault.EncryptedBlob{
		Metadata: data_vault.BlobMetadata{
			ID:    "blob-1",
			Scope: req.Scope,
			Owner: req.Owner,
			OrgID: req.OrgID,
		},
	}, nil
}

func (stubVault) Retrieve(_ context.Context, req *data_vault.RetrieveRequest) ([]byte, *data_vault.BlobMetadata, error) {
	return []byte("secret"), &data_vault.BlobMetadata{ID: req.ID, Scope: data_vault.ScopeSupport}, nil
}

func (stubVault) RetrieveStream(_ context.Context, req *data_vault.RetrieveRequest) (io.ReadCloser, *data_vault.BlobMetadata, error) {
	return io.NopCloser(bytes.NewReader([]byte("secret"))), &data_vault.BlobMetadata{ID: req.ID}, nil
}

func (stubVault) GetMetadata(_ context.Context, id data_vault.BlobID, _ string, _ string) (*data_vault.BlobMetadata, error) {
	return &data_vault.BlobMetadata{ID: id, Scope: data_vault.ScopeSupport}, nil
}

func (stubVault) Delete(_ context.Context, _ data_vault.BlobID, _ string) error {
	return nil
}

func (stubVault) RotateKeys(_ context.Context, _ data_vault.Scope) error {
	return nil
}

func (stubVault) GetAuditEvents(_ context.Context, _ data_vault.AuditFilter) ([]*data_vault.AuditEvent, error) {
	return []*data_vault.AuditEvent{}, nil
}

func (stubVault) Close() error {
	return nil
}

func TestPortalVaultUploadErrors(t *testing.T) {
	srv, err := NewPortalAPIServer(PortalAPIServerConfig{
		AllowInsecure:        true,
		VaultService:         stubVault{},
		VaultMaxPayloadBytes: 1024,
	})
	if err != nil {
		t.Fatalf("new portal api server: %v", err)
	}

	router := mux.NewRouter()
	srv.setupRoutes(router)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/vault/blobs", bytes.NewBufferString(`{"scope":"support"}`))
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing payload, got %d: %s", resp.Code, resp.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/vault/blobs", bytes.NewBufferString(`{"scope":"support","payload_base64":"bad@@@"}`))
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid base64, got %d: %s", resp.Code, resp.Body.String())
	}
}

func TestPortalVaultUploadSuccess(t *testing.T) {
	srv, err := NewPortalAPIServer(PortalAPIServerConfig{
		AllowInsecure:        true,
		VaultService:         stubVault{},
		VaultMaxPayloadBytes: 1024,
	})
	if err != nil {
		t.Fatalf("new portal api server: %v", err)
	}

	router := mux.NewRouter()
	srv.setupRoutes(router)

	payload := base64.StdEncoding.EncodeToString([]byte("hello"))
	body := `{"scope":"support","payload_base64":"` + payload + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/vault/blobs", bytes.NewBufferString(body))
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", resp.Code, resp.Body.String())
	}
}
