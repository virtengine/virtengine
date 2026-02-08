package provider_daemon

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/gorilla/mux"

	portalauth "github.com/virtengine/virtengine/pkg/provider_daemon/auth"
	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

type mockLifecycleExecutor struct {
	lastAction marketplace.LifecycleActionType
	lastReq    *LifecycleActionRequest
	result     *LifecycleActionResult
	err        error
}

const (
	testChainID  = "virtengine-1"
	testPathBase = "/api/v1/deployments/lease1/actions"
)

func (m *mockLifecycleExecutor) Start(ctx context.Context, req *LifecycleActionRequest) (*LifecycleActionResult, error) {
	m.lastAction = marketplace.LifecycleActionStart
	m.lastReq = req
	return m.result, m.err
}

func (m *mockLifecycleExecutor) Stop(ctx context.Context, req *LifecycleActionRequest) (*LifecycleActionResult, error) {
	m.lastAction = marketplace.LifecycleActionStop
	m.lastReq = req
	return m.result, m.err
}

func (m *mockLifecycleExecutor) Restart(ctx context.Context, req *LifecycleActionRequest) (*LifecycleActionResult, error) {
	m.lastAction = marketplace.LifecycleActionRestart
	m.lastReq = req
	return m.result, m.err
}

func (m *mockLifecycleExecutor) Resize(ctx context.Context, req *LifecycleActionRequest) (*LifecycleActionResult, error) {
	m.lastAction = marketplace.LifecycleActionResize
	m.lastReq = req
	return m.result, m.err
}

type mockPortalChainQuery struct {
	NoopChainQuery
	hasRole    bool
	hasConsent bool
}

func (m mockPortalChainQuery) HasRole(_ context.Context, _ string, _ string) (bool, error) {
	return m.hasRole, nil
}

func (m mockPortalChainQuery) HasConsent(_ context.Context, _ string, _ string) (bool, error) {
	return m.hasConsent, nil
}

type mockLeaseQuerier struct {
	owner string
}

func (m mockLeaseQuerier) GetLease(_ context.Context, _ string) (*portalauth.Lease, error) {
	return &portalauth.Lease{ID: "lease1", Owner: m.owner}, nil
}

func TestPortalLifecycleActionAuthorized(t *testing.T) {
	priv := secp256k1.GenPrivKey()
	address := sdk.AccAddress(priv.PubKey().Address()).String()

	exec := &mockLifecycleExecutor{result: &LifecycleActionResult{OperationID: "op-1", Success: true, State: marketplace.LifecycleOpStateExecuting}}
	cfg := DefaultPortalAPIServerConfig()
	cfg.AllowInsecure = false
	cfg.LifecycleExecutor = exec
	cfg.LifecycleRequireConsent = true
	cfg.LifecycleAllowedRoles = []string{"customer"}
	cfg.ChainQuery = mockPortalChainQuery{hasRole: true, hasConsent: true}
	cfg.WalletAuthChainID = testChainID
	cfg.WalletAuthChainQuery = mockLeaseQuerier{owner: address}
	server, err := NewPortalAPIServer(cfg)
	require.NoError(t, err)

	body := []byte(`{"action":"start"}`)
	req := buildSignedRequest(t, priv, cfg.WalletAuthChainID, address, body)

	rr := httptest.NewRecorder()
	handler := server.authMiddleware(true)(server.leaseOwnerMiddleware()(http.HandlerFunc(server.handleDeploymentAction)))
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, marketplace.LifecycleActionStart, exec.lastAction)
	require.Equal(t, "lease1", exec.lastReq.AllocationID)
	require.Equal(t, address, exec.lastReq.RequestedBy)
}

func TestPortalLifecycleActionResize(t *testing.T) {
	priv := secp256k1.GenPrivKey()
	address := sdk.AccAddress(priv.PubKey().Address()).String()

	exec := &mockLifecycleExecutor{result: &LifecycleActionResult{OperationID: "op-resize", Success: true, State: marketplace.LifecycleOpStateExecuting}}
	cfg := DefaultPortalAPIServerConfig()
	cfg.AllowInsecure = false
	cfg.LifecycleExecutor = exec
	cfg.LifecycleRequireConsent = false
	cfg.LifecycleAllowedRoles = []string{"customer"}
	cfg.ChainQuery = mockPortalChainQuery{hasRole: true, hasConsent: true}
	cfg.WalletAuthChainID = testChainID
	cfg.WalletAuthChainQuery = mockLeaseQuerier{owner: address}
	server, err := NewPortalAPIServer(cfg)
	require.NoError(t, err)

	body := []byte(`{"action":"resize","parameters":{"cpu_cores":4,"memory_mb":8192,"storage_gb":100}}`)
	req := buildSignedRequest(t, priv, cfg.WalletAuthChainID, address, body)

	rr := httptest.NewRecorder()
	handler := server.authMiddleware(true)(server.leaseOwnerMiddleware()(http.HandlerFunc(server.handleDeploymentAction)))
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, marketplace.LifecycleActionResize, exec.lastAction)
	require.NotNil(t, exec.lastReq.ResizeSpec)
	require.NotNil(t, exec.lastReq.ResizeSpec.CPUCores)
	require.Equal(t, uint32(4), *exec.lastReq.ResizeSpec.CPUCores)
	require.NotNil(t, exec.lastReq.ResizeSpec.MemoryMB)
	require.Equal(t, uint64(8192), *exec.lastReq.ResizeSpec.MemoryMB)
}

func TestPortalLifecycleActionRoleDenied(t *testing.T) {
	priv := secp256k1.GenPrivKey()
	address := sdk.AccAddress(priv.PubKey().Address()).String()

	cfg := DefaultPortalAPIServerConfig()
	cfg.AllowInsecure = false
	cfg.LifecycleExecutor = &mockLifecycleExecutor{result: &LifecycleActionResult{Success: true}}
	cfg.LifecycleRequireConsent = true
	cfg.LifecycleAllowedRoles = []string{"customer"}
	cfg.ChainQuery = mockPortalChainQuery{hasRole: false, hasConsent: true}
	cfg.WalletAuthChainID = testChainID
	cfg.WalletAuthChainQuery = mockLeaseQuerier{owner: address}
	server, err := NewPortalAPIServer(cfg)
	require.NoError(t, err)

	body := []byte(`{"action":"start"}`)
	req := buildSignedRequest(t, priv, cfg.WalletAuthChainID, address, body)

	rr := httptest.NewRecorder()
	handler := server.authMiddleware(true)(server.leaseOwnerMiddleware()(http.HandlerFunc(server.handleDeploymentAction)))
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusForbidden, rr.Code)
}

func TestPortalLifecycleActionConsentDenied(t *testing.T) {
	priv := secp256k1.GenPrivKey()
	address := sdk.AccAddress(priv.PubKey().Address()).String()

	cfg := DefaultPortalAPIServerConfig()
	cfg.AllowInsecure = false
	cfg.LifecycleExecutor = &mockLifecycleExecutor{result: &LifecycleActionResult{Success: true}}
	cfg.LifecycleRequireConsent = true
	cfg.LifecycleAllowedRoles = []string{"customer"}
	cfg.ChainQuery = mockPortalChainQuery{hasRole: true, hasConsent: false}
	cfg.WalletAuthChainID = testChainID
	cfg.WalletAuthChainQuery = mockLeaseQuerier{owner: address}
	server, err := NewPortalAPIServer(cfg)
	require.NoError(t, err)

	body := []byte(`{"action":"start"}`)
	req := buildSignedRequest(t, priv, cfg.WalletAuthChainID, address, body)

	rr := httptest.NewRecorder()
	handler := server.authMiddleware(true)(server.leaseOwnerMiddleware()(http.HandlerFunc(server.handleDeploymentAction)))
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusForbidden, rr.Code)
}

func buildSignedRequest(t *testing.T, priv *secp256k1.PrivKey, chainID, address string, body []byte) *http.Request {
	t.Helper()

	bodyHash := hashBodyForTest(body)
	nonce := "nonce-123"
	timestamp := time.Now().UTC()

	requestData := portalauth.RequestData{
		Method:    http.MethodPost,
		Path:      testPathBase,
		Timestamp: timestamp.UnixMilli(),
		Nonce:     nonce,
		BodyHash:  bodyHash,
	}

	dataToSign := serializeRequestData(requestData)
	signDoc := buildSignDoc(chainID, address, base64.StdEncoding.EncodeToString([]byte(dataToSign)))
	signBytes, err := canonicalJSON(signDoc)
	require.NoError(t, err)

	signature, err := priv.Sign([]byte(signBytes))
	require.NoError(t, err)

	pubKey := priv.PubKey().Bytes()

	req := httptest.NewRequest(http.MethodPost, "http://example.com"+testPathBase, bytes.NewReader(body))
	req.Header.Set(portalauth.HeaderAddress, address)
	req.Header.Set(portalauth.HeaderTimestamp, strconv.FormatInt(timestamp.UnixMilli(), 10))
	req.Header.Set(portalauth.HeaderNonce, nonce)
	req.Header.Set(portalauth.HeaderSignature, base64.StdEncoding.EncodeToString(signature))
	req.Header.Set(portalauth.HeaderPubKey, base64.StdEncoding.EncodeToString(pubKey))
	req.Header.Set("Content-Type", "application/json")

	if deploymentID := deploymentIDFromPath(testPathBase); deploymentID != "" {
		req = mux.SetURLVars(req, map[string]string{"deploymentId": deploymentID})
	}

	return req
}

func deploymentIDFromPath(path string) string {
	parts := strings.Split(path, "/")
	for i := 0; i < len(parts)-1; i++ {
		if parts[i] == "deployments" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

func hashBodyForTest(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	if canonical, ok := canonicalizeJSON(body); ok {
		sum := sha256.Sum256([]byte(canonical))
		return hex.EncodeToString(sum[:])
	}
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func canonicalizeJSON(input []byte) (string, bool) {
	decoder := json.NewDecoder(bytes.NewReader(input))
	decoder.UseNumber()

	var payload any
	if err := decoder.Decode(&payload); err != nil {
		return "", false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return "", false
	}
	canonical, err := canonicalJSON(payload)
	if err != nil {
		return "", false
	}
	return canonical, true
}

func serializeRequestData(data portalauth.RequestData) string {
	return fmt.Sprintf(
		"{\"method\":\"%s\",\"path\":\"%s\",\"timestamp\":%d,\"nonce\":\"%s\",\"body_hash\":\"%s\"}",
		data.Method,
		data.Path,
		data.Timestamp,
		data.Nonce,
		data.BodyHash,
	)
}

func buildSignDoc(chainID, address, dataBase64 string) map[string]any {
	return map[string]any{
		"chain_id":       chainID,
		"account_number": "0",
		"sequence":       "0",
		"fee": map[string]any{
			"gas":    "0",
			"amount": []any{},
		},
		"msgs": []any{
			map[string]any{
				"type": "sign/MsgSignData",
				"value": map[string]any{
					"signer": address,
					"data":   dataBase64,
				},
			},
		},
		"memo": "",
	}
}

func canonicalJSON(value any) (string, error) {
	switch typed := value.(type) {
	case nil:
		return "null", nil
	case string:
		return marshalPrimitive(typed)
	case bool:
		return marshalPrimitive(typed)
	case json.Number:
		return typed.String(), nil
	case float64, float32, int, int64, int32, int16, int8, uint, uint64, uint32, uint16, uint8:
		return marshalPrimitive(typed)
	case []any:
		var buf bytes.Buffer
		buf.WriteByte('[')
		for i, item := range typed {
			if i > 0 {
				buf.WriteByte(',')
			}
			encoded, err := canonicalJSON(item)
			if err != nil {
				return "", err
			}
			buf.WriteString(encoded)
		}
		buf.WriteByte(']')
		return buf.String(), nil
	case map[string]any:
		return canonicalMap(typed)
	default:
		return marshalPrimitive(typed)
	}
}

func canonicalMap(value map[string]any) (string, error) {
	keys := make([]string, 0, len(value))
	for key := range value {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, key := range keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		encodedKey, err := marshalPrimitive(key)
		if err != nil {
			return "", err
		}
		buf.WriteString(encodedKey)
		buf.WriteByte(':')
		encodedValue, err := canonicalJSON(value[key])
		if err != nil {
			return "", err
		}
		buf.WriteString(encodedValue)
	}
	buf.WriteByte('}')
	return buf.String(), nil
}

func marshalPrimitive(value any) (string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}
