//go:build e2e.integration

package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/virtengine/virtengine/pkg/inference"
	inferencepb "github.com/virtengine/virtengine/pkg/inference/proto"
)

var _ encoding.Codec = jsonTestCodec{}

type jsonTestCodec struct{}

func (jsonTestCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (jsonTestCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (jsonTestCodec) Name() string {
	return "json"
}

func TestSidecarE2EVerifiedBundleServesTraffic(t *testing.T) {
	tfServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/models/trust_score:predict":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"predictions": [][]float32{{55.5}},
			})
		case "/v1/models/trust_score":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"model_version_status": []map[string]any{
					{"state": "AVAILABLE"},
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(tfServer.Close)

	modelDir, manifestPath := createReleaseBundle(t, "v1.0.0")
	server, err := NewInferenceSidecarServer(
		inference.InferenceConfig{
			ModelPath:        modelDir,
			ModelVersion:     "v1.0.0",
			Timeout:          2 * time.Second,
			MaxMemoryMB:      512,
			Deterministic:    true,
			ForceCPU:         true,
			RandomSeed:       42,
			ExpectedInputDim: inference.TotalFeatureDim,
		},
		inference.TFServingConfig{
			BaseURL:   tfServer.URL,
			ModelName: "trust_score",
			Timeout:   2 * time.Second,
		},
		manifestPath,
		noopLogger{},
	)
	if err != nil {
		t.Fatalf("create sidecar server: %v", err)
	}
	t.Cleanup(func() { _ = server.Close() })

	grpcConn, readinessURL := startSidecarE2EServer(t, server)
	client := inferencepb.NewInferenceServiceClient(grpcConn)
	healthClient := grpc_health_v1.NewHealthClient(grpcConn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	grpcHealth, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: inferencepb.ServiceName})
	if err != nil {
		t.Fatalf("grpc health check failed: %v", err)
	}
	if grpcHealth.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		t.Fatalf("expected grpc health SERVING, got %s", grpcHealth.Status.String())
	}

	healthResp, err := client.HealthCheck(ctx, &inferencepb.HealthCheckRequest{})
	if err != nil {
		t.Fatalf("sidecar health RPC failed: %v", err)
	}
	if healthResp.Status != inferencepb.HealthStatus_HEALTH_STATUS_HEALTHY {
		t.Fatalf("expected healthy sidecar, got %s (%s)", healthResp.Status.String(), healthResp.ErrorMessage)
	}

	httpResp, err := http.Get(readinessURL + "/health")
	if err != nil {
		t.Fatalf("readiness HTTP request failed: %v", err)
	}
	defer httpResp.Body.Close()
	if httpResp.StatusCode != http.StatusOK {
		t.Fatalf("expected readiness 200, got %d", httpResp.StatusCode)
	}
	var readiness readinessStatus
	if err := json.NewDecoder(httpResp.Body).Decode(&readiness); err != nil {
		t.Fatalf("decode readiness payload: %v", err)
	}
	if !readiness.Ready || readiness.State != string(verificationStateVerified) {
		t.Fatalf("expected verified readiness payload, got %#v", readiness)
	}

	features := make([]float32, inference.TotalFeatureDim)
	features[0] = 0.5
	resp, err := client.ComputeScore(ctx, &inferencepb.ComputeScoreRequest{
		Features: features,
		Metadata: &inferencepb.InferenceMetadata{
			AccountAddress: "addr",
			BlockHeight:    1,
			RequestID:      "req-1",
		},
	})
	if err != nil {
		t.Fatalf("ComputeScore failed: %v", err)
	}
	if resp.Score != 55 {
		t.Fatalf("expected score 55, got %d", resp.Score)
	}
	if resp.ModelVersion != "v1.0.0" || resp.ModelHash == "" {
		t.Fatalf("expected verified model metadata in response, got version=%q hash=%q", resp.ModelVersion, resp.ModelHash)
	}
}

func TestSidecarE2EBadManifestStaysNotReady(t *testing.T) {
	modelDir, manifestPath := createReleaseBundle(t, "v1.0.0")
	payload := mustReadJSON(t, manifestPath)
	payload["model"].(map[string]any)["runtime_hash"] = "placeholder"
	writeJSON(t, manifestPath, payload)

	server, err := NewInferenceSidecarServer(
		inference.InferenceConfig{
			ModelPath:        modelDir,
			ModelVersion:     "v1.0.0",
			Timeout:          2 * time.Second,
			MaxMemoryMB:      512,
			Deterministic:    true,
			ForceCPU:         true,
			RandomSeed:       42,
			ExpectedInputDim: inference.TotalFeatureDim,
		},
		inference.TFServingConfig{},
		manifestPath,
		noopLogger{},
	)
	if err != nil {
		t.Fatalf("create sidecar server: %v", err)
	}
	t.Cleanup(func() { _ = server.Close() })

	grpcConn, readinessURL := startSidecarE2EServer(t, server)
	client := inferencepb.NewInferenceServiceClient(grpcConn)
	healthClient := grpc_health_v1.NewHealthClient(grpcConn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	grpcHealth, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: inferencepb.ServiceName})
	if err != nil {
		t.Fatalf("grpc health check failed: %v", err)
	}
	if grpcHealth.Status != grpc_health_v1.HealthCheckResponse_NOT_SERVING {
		t.Fatalf("expected grpc health NOT_SERVING, got %s", grpcHealth.Status.String())
	}

	healthResp, err := client.HealthCheck(ctx, &inferencepb.HealthCheckRequest{})
	if err != nil {
		t.Fatalf("sidecar health RPC failed: %v", err)
	}
	if healthResp.Status != inferencepb.HealthStatus_HEALTH_STATUS_UNHEALTHY {
		t.Fatalf("expected unhealthy sidecar, got %s (%s)", healthResp.Status.String(), healthResp.ErrorMessage)
	}
	if !strings.Contains(healthResp.ErrorMessage, string(verificationStateBadManifest)) {
		t.Fatalf("expected bad_manifest error, got %q", healthResp.ErrorMessage)
	}

	httpResp, err := http.Get(readinessURL + "/health")
	if err != nil {
		t.Fatalf("readiness HTTP request failed: %v", err)
	}
	defer httpResp.Body.Close()
	if httpResp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected readiness 503, got %d", httpResp.StatusCode)
	}
	var readiness readinessStatus
	if err := json.NewDecoder(httpResp.Body).Decode(&readiness); err != nil {
		t.Fatalf("decode readiness payload: %v", err)
	}
	if readiness.Ready || readiness.State != string(verificationStateBadManifest) {
		t.Fatalf("expected bad_manifest readiness payload, got %#v", readiness)
	}

	_, err = client.ComputeScore(ctx, &inferencepb.ComputeScoreRequest{
		Features: make([]float32, inference.TotalFeatureDim),
	})
	if err == nil || !strings.Contains(err.Error(), string(verificationStateBadManifest)) {
		t.Fatalf("expected ComputeScore to fail with bad_manifest, got %v", err)
	}
}

func startSidecarE2EServer(t *testing.T, server *InferenceSidecarServer) (*grpc.ClientConn, string) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	grpcServer := grpc.NewServer(grpc.ForceServerCodec(jsonTestCodec{}))
	inferencepb.RegisterInferenceServiceServer(grpcServer, server)
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthStatus := grpc_health_v1.HealthCheckResponse_SERVING
	if readiness := server.Readiness(); readiness == nil || !readiness.Ready() {
		healthStatus = grpc_health_v1.HealthCheckResponse_NOT_SERVING
	}
	healthServer.SetServingStatus(inferencepb.ServiceName, healthStatus)

	go func() {
		_ = grpcServer.Serve(listener)
	}()
	t.Cleanup(func() {
		grpcServer.GracefulStop()
		_ = listener.Close()
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		statusCode, payload := readinessHTTPResponse(server.Readiness())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_, _ = w.Write(payload)
	})
	httpServer := httptest.NewServer(mux)
	t.Cleanup(httpServer.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(
		ctx,
		listener.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(jsonTestCodec{})),
		grpc.WithBlock(),
	)
	if err != nil {
		t.Fatalf("dial grpc sidecar: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	return conn, httpServer.URL
}
