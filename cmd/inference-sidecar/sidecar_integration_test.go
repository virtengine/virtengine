//go:build integration

package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/virtengine/virtengine/pkg/inference"
	inferencepb "github.com/virtengine/virtengine/pkg/inference/proto"
)

type testLogger struct{}

func (testLogger) Debug(string, ...interface{}) {}
func (testLogger) Info(string, ...interface{})  {}
func (testLogger) Warn(string, ...interface{})  {}
func (testLogger) Error(string, ...interface{}) {}

func TestSidecarIntegration(t *testing.T) {
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

	modelDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(modelDir, "stub.bin"), []byte("model"), 0o600); err != nil {
		t.Fatalf("write stub model: %v", err)
	}

	config := inference.InferenceConfig{
		ModelPath:        modelDir,
		ModelVersion:     "v1.0.0",
		Timeout:          2 * time.Second,
		MaxMemoryMB:      512,
		UseSidecar:       false,
		Deterministic:    false,
		ForceCPU:         true,
		RandomSeed:       42,
		ExpectedInputDim: inference.TotalFeatureDim,
	}

	servingConfig := inference.TFServingConfig{
		BaseURL:   tfServer.URL,
		ModelName: "trust_score",
		Timeout:   2 * time.Second,
	}

	server, err := NewInferenceSidecarServer(config, servingConfig, testLogger{})
	if err != nil {
		t.Fatalf("create sidecar server: %v", err)
	}
	t.Cleanup(func() { _ = server.Close() })

	grpcServer := grpc.NewServer()
	inferencepb.RegisterInferenceServiceServer(grpcServer, server)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	go grpcServer.Serve(listener)
	t.Cleanup(grpcServer.Stop)

	conn, err := grpc.Dial(listener.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	client := inferencepb.NewInferenceServiceClient(conn)
	features := make([]float32, inference.TotalFeatureDim)
	features[0] = 0.5

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := client.ComputeScore(ctx, &inferencepb.ComputeScoreRequest{
		Features: features,
		Metadata: &inferencepb.InferenceMetadata{
			AccountAddress: "addr",
			BlockHeight:    1,
			RequestID:      "req-1",
		},
		ReturnContributions: true,
	})
	if err != nil {
		t.Fatalf("ComputeScore failed: %v", err)
	}
	if resp.Score != 55 {
		t.Fatalf("expected score 55, got %d", resp.Score)
	}
	if resp.RawScore != 55.5 {
		t.Fatalf("expected raw score 55.5, got %f", resp.RawScore)
	}
	if resp.InputHash == "" || resp.OutputHash == "" {
		t.Fatalf("expected hashes to be set")
	}
}
