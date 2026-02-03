// Copyright 2024-2026 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTFServingClientPredict(t *testing.T) {
	t.Parallel()

	var receivedInputName string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models/trust_score:predict" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		var payload struct {
			Instances []map[string][]float32 `json:"instances"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		for k := range payload.Instances[0] {
			receivedInputName = k
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"predictions": [][]float32{{42.5}},
		})
	}))
	t.Cleanup(server.Close)

	client, err := NewTFServingClient(TFServingConfig{
		BaseURL:   server.URL,
		ModelName: "trust_score",
		InputName: "features",
		Timeout:   2 * time.Second,
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	output, endpoint, _, err := client.Predict(context.Background(), make([]float32, TotalFeatureDim))
	if err != nil {
		t.Fatalf("predict failed: %v", err)
	}
	if endpoint != server.URL {
		t.Fatalf("expected endpoint %s, got %s", server.URL, endpoint)
	}
	if receivedInputName != "features" {
		t.Fatalf("expected input name features, got %s", receivedInputName)
	}
	if len(output) != 1 || output[0] != 42.5 {
		t.Fatalf("unexpected output: %#v", output)
	}
}

func TestTFServingClientPredictionMap(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"predictions": []map[string]any{
				{"trust_score": []float32{18.2}},
			},
		})
	}))
	t.Cleanup(server.Close)

	client, err := NewTFServingClient(TFServingConfig{
		BaseURL:    server.URL,
		ModelName:  "trust_score",
		OutputName: "trust_score",
		Timeout:    2 * time.Second,
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	output, _, _, err := client.Predict(context.Background(), make([]float32, TotalFeatureDim))
	if err != nil {
		t.Fatalf("predict failed: %v", err)
	}
	if len(output) != 1 || output[0] != 18.2 {
		t.Fatalf("unexpected output: %#v", output)
	}
}

func TestTFServingClientFallback(t *testing.T) {
	t.Parallel()

	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(primary.Close)

	fallback := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"predictions": [][]float32{{11.0}},
		})
	}))
	t.Cleanup(fallback.Close)

	client, err := NewTFServingClient(TFServingConfig{
		BaseURL:     primary.URL,
		FallbackURL: fallback.URL,
		ModelName:   "trust_score",
		Timeout:     2 * time.Second,
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	output, endpoint, _, err := client.Predict(context.Background(), make([]float32, TotalFeatureDim))
	if err != nil {
		t.Fatalf("predict failed: %v", err)
	}
	if endpoint != fallback.URL {
		t.Fatalf("expected fallback endpoint %s, got %s", fallback.URL, endpoint)
	}
	if len(output) != 1 || output[0] != 11.0 {
		t.Fatalf("unexpected output: %#v", output)
	}
}

func TestTFServingClientHealth(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"model_version_status": []map[string]any{
				{"state": "AVAILABLE"},
			},
		})
	}))
	t.Cleanup(server.Close)

	client, err := NewTFServingClient(TFServingConfig{
		BaseURL:   server.URL,
		ModelName: "trust_score",
		Timeout:   2 * time.Second,
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	endpoint, err := client.CheckHealth(context.Background())
	if err != nil {
		t.Fatalf("health check failed: %v", err)
	}
	if endpoint != server.URL {
		t.Fatalf("unexpected endpoint: %s", endpoint)
	}
}

func BenchmarkTFServingClientPredict(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"predictions": [][]float32{{42.0}},
		})
	}))
	b.Cleanup(server.Close)

	client, err := NewTFServingClient(TFServingConfig{
		BaseURL:   server.URL,
		ModelName: "trust_score",
		Timeout:   2 * time.Second,
	})
	if err != nil {
		b.Fatalf("create client: %v", err)
	}

	features := make([]float32, TotalFeatureDim)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, _, _, err := client.Predict(ctx, features); err != nil {
			b.Fatalf("predict failed: %v", err)
		}
	}
}
