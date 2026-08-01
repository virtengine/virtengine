package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/virtengine/virtengine/pkg/inference"
	inferencepb "github.com/virtengine/virtengine/pkg/inference/proto"
)

type noopLogger struct{}

func (noopLogger) Debug(string, ...interface{}) {}
func (noopLogger) Info(string, ...interface{})  {}
func (noopLogger) Warn(string, ...interface{})  {}
func (noopLogger) Error(string, ...interface{}) {}

func TestVerifyModelBundleMissingManifest(t *testing.T) {
	modelDir := filepath.Join(t.TempDir(), "model")
	if err := os.MkdirAll(modelDir, 0o750); err != nil {
		t.Fatalf("mkdir model dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(modelDir, "saved_model.pb"), []byte("model"), 0o600); err != nil {
		t.Fatalf("write model: %v", err)
	}

	_, err := verifyModelBundle(modelDir, "", "", "")
	if err == nil {
		t.Fatal("expected missing manifest error")
	}

	var verifyErr *verificationError
	if !errors.As(err, &verifyErr) {
		t.Fatalf("expected verificationError, got %T", err)
	}
	if verifyErr.State != verificationStateMissingManifest {
		t.Fatalf("expected missing_manifest, got %s", verifyErr.State)
	}
}

func TestVerifyModelBundleRejectsStaleManifest(t *testing.T) {
	modelDir, manifestPath := createReleaseBundle(t)

	payload := mustReadJSON(t, manifestPath)
	payload["model"].(map[string]any)["runtime_hash"] = stringsOfLength(64, "a")
	writeJSON(t, manifestPath, payload)

	_, err := verifyModelBundle(modelDir, manifestPath, releaseBundleVersion, "")
	if err == nil {
		t.Fatal("expected stale manifest error")
	}

	var verifyErr *verificationError
	if !errors.As(err, &verifyErr) {
		t.Fatalf("expected verificationError, got %T", err)
	}
	if verifyErr.State != verificationStateStaleManifest {
		t.Fatalf("expected stale_manifest, got %s", verifyErr.State)
	}
}

func TestNewInferenceSidecarServerReportsBadManifestState(t *testing.T) {
	modelDir, manifestPath := createReleaseBundle(t)
	writeJSON(t, manifestPath, map[string]any{
		"schema_version": "veid.release.manifest/v1",
		"model": map[string]any{
			"name":         "trust_score",
			"version":      releaseBundleVersion,
			"model_dir":    "model",
			"runtime_hash": "placeholder",
		},
		"artifacts": []any{},
	})

	config := inference.InferenceConfig{
		ModelPath:        modelDir,
		Timeout:          2 * time.Second,
		MaxMemoryMB:      512,
		Deterministic:    true,
		ForceCPU:         true,
		RandomSeed:       42,
		ExpectedInputDim: inference.TotalFeatureDim,
	}

	server, err := NewInferenceSidecarServer(config, inference.TFServingConfig{}, manifestPath, noopLogger{})
	if err != nil {
		t.Fatalf("unexpected server error: %v", err)
	}
	t.Cleanup(func() { _ = server.Close() })

	readiness := server.Readiness()
	if readiness == nil {
		t.Fatal("expected readiness state")
	}
	if readiness.State != verificationStateBadManifest {
		t.Fatalf("expected bad_manifest, got %s", readiness.State)
	}

	resp, err := server.HealthCheck(context.Background(), nil)
	if err != nil {
		t.Fatalf("health check failed: %v", err)
	}
	if !strings.Contains(resp.ErrorMessage, string(verificationStateBadManifest)) {
		t.Fatalf("expected bad_manifest in error message, got %q", resp.ErrorMessage)
	}
	if _, err := server.ComputeScore(context.Background(), &inferencepb.ComputeScoreRequest{
		Features: make([]float32, inference.TotalFeatureDim),
	}); err == nil || !strings.Contains(err.Error(), string(verificationStateBadManifest)) {
		t.Fatalf("expected ComputeScore to fail with bad_manifest, got %v", err)
	}
}

func TestVerifyModelBundleRejectsMissingRequiredArtifactEntry(t *testing.T) {
	modelDir, manifestPath := createReleaseBundle(t)

	payload := mustReadJSON(t, manifestPath)
	artifacts := payload["artifacts"].([]any)
	filtered := make([]any, 0, len(artifacts))
	for _, artifact := range artifacts {
		record := artifact.(map[string]any)
		if record["path"] == "MODEL_HASH.txt" {
			continue
		}
		filtered = append(filtered, record)
	}
	payload["artifacts"] = filtered
	writeJSON(t, manifestPath, payload)

	_, err := verifyModelBundle(modelDir, manifestPath, releaseBundleVersion, "")
	if err == nil {
		t.Fatal("expected bad manifest error")
	}

	var verifyErr *verificationError
	if !errors.As(err, &verifyErr) {
		t.Fatalf("expected verificationError, got %T", err)
	}
	if verifyErr.State != verificationStateBadManifest {
		t.Fatalf("expected bad_manifest, got %s", verifyErr.State)
	}
	if !strings.Contains(verifyErr.Error(), "MODEL_HASH.txt") {
		t.Fatalf("expected missing required artifact in error, got %v", verifyErr)
	}
}

func TestVerifyModelBundleRejectsArtifactPathTraversal(t *testing.T) {
	modelDir, manifestPath := createReleaseBundle(t)
	versionDir := filepath.Dir(manifestPath)
	escapedArtifactPath := filepath.Join(filepath.Dir(versionDir), "outside.txt")
	mustWriteFile(t, escapedArtifactPath, []byte("outside-bundle"))
	escapedArtifactHash := mustComputeHash(t, escapedArtifactPath)
	escapedArtifactInfo, err := os.Stat(escapedArtifactPath)
	if err != nil {
		t.Fatalf("stat escaped artifact: %v", err)
	}

	payload := mustReadJSON(t, manifestPath)
	artifacts := payload["artifacts"].([]any)
	artifacts[0].(map[string]any)["path"] = filepath.ToSlash(filepath.Join("..", filepath.Base(escapedArtifactPath)))
	artifacts[0].(map[string]any)["sha256"] = escapedArtifactHash
	artifacts[0].(map[string]any)["size_bytes"] = escapedArtifactInfo.Size()
	payload["artifacts"] = artifacts
	writeJSON(t, manifestPath, payload)

	_, err = verifyModelBundle(modelDir, manifestPath, releaseBundleVersion, "")
	if err == nil {
		t.Fatal("expected bad manifest error")
	}

	var verifyErr *verificationError
	if !errors.As(err, &verifyErr) {
		t.Fatalf("expected verificationError, got %T", err)
	}
	if verifyErr.State != verificationStateBadManifest {
		t.Fatalf("expected bad_manifest, got %s", verifyErr.State)
	}
	if !strings.Contains(verifyErr.Error(), "outside bundle root") {
		t.Fatalf("expected outside bundle root error, got %v", verifyErr)
	}
}

func TestVerifyModelBundleRejectsMissingProvenanceDeclaration(t *testing.T) {
	modelDir, manifestPath := createReleaseBundle(t)
	payload := mustReadJSON(t, manifestPath)
	delete(payload, "provenance")
	writeJSON(t, manifestPath, payload)

	_, err := verifyModelBundle(modelDir, manifestPath, releaseBundleVersion, "")
	assertVerificationError(t, err, verificationStateBadManifest, "provenance path is required")
}

func TestVerifyModelBundleRejectsInvalidProvenanceDeclaration(t *testing.T) {
	tests := []struct {
		name       string
		update     func(map[string]any)
		wantReason string
	}{
		{
			name: "missing path",
			update: func(provenance map[string]any) {
				delete(provenance, "path")
			},
			wantReason: "provenance path is required",
		},
		{
			name: "missing hash",
			update: func(provenance map[string]any) {
				delete(provenance, "sha256")
			},
			wantReason: "provenance hash must be a valid SHA-256 digest",
		},
		{
			name: "path traversal",
			update: func(provenance map[string]any) {
				provenance["path"] = "../model_provenance.json"
			},
			wantReason: "outside bundle root",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			modelDir, manifestPath := createReleaseBundle(t)
			payload := mustReadJSON(t, manifestPath)
			test.update(payload["provenance"].(map[string]any))
			writeJSON(t, manifestPath, payload)

			_, err := verifyModelBundle(modelDir, manifestPath, releaseBundleVersion, "")
			assertVerificationError(t, err, verificationStateBadManifest, test.wantReason)
		})
	}
}

func TestVerifyModelBundleRejectsMissingProvenanceArtifact(t *testing.T) {
	modelDir, manifestPath := createReleaseBundle(t)
	payload := mustReadJSON(t, manifestPath)
	artifacts := payload["artifacts"].([]any)
	filtered := make([]any, 0, len(artifacts)-1)
	for _, artifact := range artifacts {
		record := artifact.(map[string]any)
		if record["path"] != "model_provenance.json" {
			filtered = append(filtered, record)
		}
	}
	payload["artifacts"] = filtered
	writeJSON(t, manifestPath, payload)

	_, err := verifyModelBundle(modelDir, manifestPath, releaseBundleVersion, "")
	assertVerificationError(t, err, verificationStateBadManifest, "model_provenance.json")
}

func TestVerifyModelBundleRejectsProvenanceDigestMismatch(t *testing.T) {
	modelDir, manifestPath := createReleaseBundle(t)
	payload := mustReadJSON(t, manifestPath)
	payload["provenance"].(map[string]any)["sha256"] = stringsOfLength(64, "a")
	writeJSON(t, manifestPath, payload)

	_, err := verifyModelBundle(modelDir, manifestPath, releaseBundleVersion, "")
	assertVerificationError(t, err, verificationStateStaleManifest, "provenance digest mismatch")
}

func TestVerifyModelBundleRejectsDependencyBlockedProductionProvenance(t *testing.T) {
	modelDir, manifestPath := createReleaseBundle(t)
	rewriteProvenance(t, manifestPath, map[string]any{
		"schema_version": "virtengine.model-provenance/v1",
		"status":         "dependency_blocked",
	})

	_, err := verifyModelBundle(modelDir, manifestPath, releaseBundleVersion, "")
	assertVerificationError(t, err, verificationStateBadManifest, "production_approved")
}

func TestVerifyModelBundleRejectsWrongProvenanceSchema(t *testing.T) {
	modelDir, manifestPath := createReleaseBundle(t)
	rewriteProvenance(t, manifestPath, map[string]any{
		"schema_version": "virtengine.model-provenance/v2",
		"status":         "production_approved",
	})

	_, err := verifyModelBundle(modelDir, manifestPath, releaseBundleVersion, "")
	assertVerificationError(t, err, verificationStateBadManifest, "unsupported provenance schema_version")
}

func TestVerifyModelBundleRejectsMalformedProvenance(t *testing.T) {
	modelDir, manifestPath := createReleaseBundle(t)
	provenancePath := filepath.Join(filepath.Dir(manifestPath), "model_provenance.json")
	mustWriteFile(t, provenancePath, []byte("{not-json\n"))
	rebindProvenanceArtifact(t, manifestPath)

	_, err := verifyModelBundle(modelDir, manifestPath, releaseBundleVersion, "")
	assertVerificationError(t, err, verificationStateBadManifest, "invalid character")
}

func TestVerifyModelBundleRejectsTamperedProvenanceArtifact(t *testing.T) {
	modelDir, manifestPath := createReleaseBundle(t)
	provenancePath := filepath.Join(filepath.Dir(manifestPath), "model_provenance.json")
	mustWriteFile(t, provenancePath, []byte(`{"schema_version":"virtengine.model-provenance/v1","status":"dependency_blocked"}`))

	_, err := verifyModelBundle(modelDir, manifestPath, releaseBundleVersion, "")
	assertVerificationError(t, err, verificationStateStaleManifest, "artifact hash mismatch")
}

func TestNewInferenceSidecarServerReportsMissingManifestState(t *testing.T) {
	modelDir := filepath.Join(t.TempDir(), "model")
	if err := os.MkdirAll(modelDir, 0o750); err != nil {
		t.Fatalf("mkdir model dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(modelDir, "saved_model.pb"), []byte("model"), 0o600); err != nil {
		t.Fatalf("write model: %v", err)
	}

	config := inference.InferenceConfig{
		ModelPath:        modelDir,
		Timeout:          2 * time.Second,
		MaxMemoryMB:      512,
		Deterministic:    true,
		ForceCPU:         true,
		RandomSeed:       42,
		ExpectedInputDim: inference.TotalFeatureDim,
	}

	server, err := NewInferenceSidecarServer(config, inference.TFServingConfig{}, "", noopLogger{})
	if err != nil {
		t.Fatalf("unexpected server error: %v", err)
	}
	t.Cleanup(func() { _ = server.Close() })

	readiness := server.Readiness()
	if readiness == nil || readiness.State != verificationStateMissingManifest {
		t.Fatalf("expected missing_manifest readiness, got %#v", readiness)
	}

	resp, err := server.HealthCheck(context.Background(), nil)
	if err != nil {
		t.Fatalf("health check failed: %v", err)
	}
	if !strings.Contains(resp.ErrorMessage, string(verificationStateMissingManifest)) {
		t.Fatalf("expected missing_manifest in error message, got %q", resp.ErrorMessage)
	}
	if _, err := server.ComputeScore(context.Background(), &inferencepb.ComputeScoreRequest{
		Features: make([]float32, inference.TotalFeatureDim),
	}); err == nil || !strings.Contains(err.Error(), string(verificationStateMissingManifest)) {
		t.Fatalf("expected ComputeScore to fail with missing_manifest, got %v", err)
	}
}

func TestNewInferenceSidecarServerReportsStaleManifestState(t *testing.T) {
	modelDir, manifestPath := createReleaseBundle(t)

	payload := mustReadJSON(t, manifestPath)
	payload["model"].(map[string]any)["runtime_hash"] = stringsOfLength(64, "a")
	writeJSON(t, manifestPath, payload)

	config := inference.InferenceConfig{
		ModelPath:        modelDir,
		Timeout:          2 * time.Second,
		MaxMemoryMB:      512,
		Deterministic:    true,
		ForceCPU:         true,
		RandomSeed:       42,
		ExpectedInputDim: inference.TotalFeatureDim,
	}

	server, err := NewInferenceSidecarServer(config, inference.TFServingConfig{}, manifestPath, noopLogger{})
	if err != nil {
		t.Fatalf("unexpected server error: %v", err)
	}
	t.Cleanup(func() { _ = server.Close() })

	readiness := server.Readiness()
	if readiness == nil || readiness.State != verificationStateStaleManifest {
		t.Fatalf("expected stale_manifest readiness, got %#v", readiness)
	}

	if _, err := server.ComputeScore(context.Background(), &inferencepb.ComputeScoreRequest{
		Features: make([]float32, inference.TotalFeatureDim),
	}); err == nil || !strings.Contains(err.Error(), string(verificationStateStaleManifest)) {
		t.Fatalf("expected ComputeScore to fail with stale_manifest, got %v", err)
	}
}

func TestNewInferenceSidecarServerReportsMissingModelState(t *testing.T) {
	modelDir, manifestPath := createReleaseBundle(t)
	if err := os.Remove(filepath.Join(filepath.Dir(manifestPath), "model_frozen.pb")); err != nil {
		t.Fatalf("remove artifact: %v", err)
	}

	config := inference.InferenceConfig{
		ModelPath:        modelDir,
		Timeout:          2 * time.Second,
		MaxMemoryMB:      512,
		Deterministic:    true,
		ForceCPU:         true,
		RandomSeed:       42,
		ExpectedInputDim: inference.TotalFeatureDim,
	}

	server, err := NewInferenceSidecarServer(config, inference.TFServingConfig{}, manifestPath, noopLogger{})
	if err != nil {
		t.Fatalf("unexpected server error: %v", err)
	}
	t.Cleanup(func() { _ = server.Close() })

	readiness := server.Readiness()
	if readiness == nil || readiness.State != verificationStateMissingModel {
		t.Fatalf("expected missing_model readiness, got %#v", readiness)
	}

	if _, err := server.ComputeScore(context.Background(), &inferencepb.ComputeScoreRequest{
		Features: make([]float32, inference.TotalFeatureDim),
	}); err == nil || !strings.Contains(err.Error(), string(verificationStateMissingModel)) {
		t.Fatalf("expected ComputeScore to fail with missing_model, got %v", err)
	}
}

func TestReadinessHTTPResponseIncludesFailureState(t *testing.T) {
	statusCode, payload := readinessHTTPResponse(&verificationResult{
		State:         verificationStateStaleManifest,
		ManifestPath:  "release_manifest.json",
		FailurePath:   "model/saved_model.pb",
		FailureReason: "artifact hash mismatch",
	})

	if statusCode != 503 {
		t.Fatalf("expected 503, got %d", statusCode)
	}

	var status readinessStatus
	if err := json.Unmarshal(payload, &status); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if status.State != string(verificationStateStaleManifest) {
		t.Fatalf("expected stale_manifest, got %s", status.State)
	}
	if status.Ready {
		t.Fatal("expected readiness false")
	}
	if !strings.Contains(status.Message, "stale_manifest") {
		t.Fatalf("expected state in message, got %q", status.Message)
	}
}

const releaseBundleVersion = "v1.0.0"

func createReleaseBundle(t *testing.T) (string, string) {
	t.Helper()

	versionDir := filepath.Join(t.TempDir(), releaseBundleVersion)
	modelDir := filepath.Join(versionDir, "model")
	varsDir := filepath.Join(modelDir, "variables")
	for _, dir := range []string{modelDir, varsDir} {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	mustWriteFile(t, filepath.Join(modelDir, "saved_model.pb"), []byte("saved-model"))
	mustWriteFile(t, filepath.Join(varsDir, "variables.data-00000-of-00001"), []byte("weights"))
	mustWriteFile(t, filepath.Join(varsDir, "variables.index"), []byte("index"))
	mustWriteFile(t, filepath.Join(versionDir, "model_frozen.pb"), []byte("frozen-graph"))

	modelHash, err := computeModelDirHash(modelDir)
	if err != nil {
		t.Fatalf("compute model hash: %v", err)
	}

	manifestJSON := map[string]any{
		"model_version": releaseBundleVersion,
		"model_hash":    modelHash,
		"input_signature": map[string]any{
			"name":  "features",
			"shape": []any{nil, 768},
			"dtype": "float32",
		},
		"output_signature": map[string]any{
			"name":  "trust_score",
			"shape": []any{nil, 1},
			"dtype": "float32",
		},
	}
	writeJSON(t, filepath.Join(versionDir, "manifest.json"), manifestJSON)

	exportMetadata := map[string]any{
		"version":        releaseBundleVersion,
		"model_hash":     modelHash,
		"signature_name": "serving_default",
		"input_signature": map[string]any{
			"name":  "features",
			"shape": []any{nil, 768},
			"dtype": "float32",
		},
		"output_signature": map[string]any{
			"name":  "trust_score",
			"shape": []any{nil, 1},
			"dtype": "float32",
		},
	}
	writeJSON(t, filepath.Join(versionDir, "export_metadata.json"), exportMetadata)

	mustWriteFile(t, filepath.Join(versionDir, "MODEL_HASH.txt"), []byte(
		"SHA256="+modelHash+"\nVERSION="+releaseBundleVersion+"\n",
	))
	writeJSON(t, filepath.Join(versionDir, "model_provenance.json"), map[string]any{
		"schema_version": "virtengine.model-provenance/v1",
		"status":         "production_approved",
	})

	artifacts := []string{
		"MODEL_HASH.txt",
		"export_metadata.json",
		"manifest.json",
		"model_provenance.json",
		"model_frozen.pb",
		filepath.ToSlash(filepath.Join("model", "saved_model.pb")),
		filepath.ToSlash(filepath.Join("model", "variables", "variables.data-00000-of-00001")),
		filepath.ToSlash(filepath.Join("model", "variables", "variables.index")),
	}

	artifactRecords := make([]map[string]any, 0, len(artifacts))
	for _, relative := range artifacts {
		absolute := filepath.Join(versionDir, filepath.FromSlash(relative))
		hash, err := computeFileHash(absolute)
		if err != nil {
			t.Fatalf("compute artifact hash for %s: %v", absolute, err)
		}
		info, err := os.Stat(absolute)
		if err != nil {
			t.Fatalf("stat artifact %s: %v", absolute, err)
		}
		artifactRecords = append(artifactRecords, map[string]any{
			"path":       relative,
			"sha256":     hash,
			"size_bytes": info.Size(),
		})
	}

	releaseManifestPath := filepath.Join(versionDir, "release_manifest.json")
	provenanceHash := mustComputeHash(t, filepath.Join(versionDir, "model_provenance.json"))
	writeJSON(t, releaseManifestPath, map[string]any{
		"schema_version": "veid.release.manifest/v1",
		"profile":        "production",
		"provenance": map[string]any{
			"path":   "model_provenance.json",
			"sha256": provenanceHash,
		},
		"model": map[string]any{
			"name":              "trust_score",
			"version":           releaseBundleVersion,
			"model_dir":         "model",
			"runtime_hash":      modelHash,
			"frozen_graph_hash": mustComputeHash(t, filepath.Join(versionDir, "model_frozen.pb")),
			"signature_name":    "serving_default",
			"input_signature":   manifestJSON["input_signature"],
			"output_signature":  manifestJSON["output_signature"],
		},
		"artifacts": artifactRecords,
	})

	return modelDir, releaseManifestPath
}

func rewriteProvenance(t *testing.T, manifestPath string, provenance map[string]any) {
	t.Helper()
	provenancePath := filepath.Join(filepath.Dir(manifestPath), "model_provenance.json")
	writeJSON(t, provenancePath, provenance)
	rebindProvenanceArtifact(t, manifestPath)
}

func rebindProvenanceArtifact(t *testing.T, manifestPath string) {
	t.Helper()
	provenancePath := filepath.Join(filepath.Dir(manifestPath), "model_provenance.json")
	provenanceHash := mustComputeHash(t, provenancePath)
	provenanceInfo, err := os.Stat(provenancePath)
	if err != nil {
		t.Fatalf("stat provenance: %v", err)
	}

	payload := mustReadJSON(t, manifestPath)
	payload["provenance"].(map[string]any)["sha256"] = provenanceHash
	for _, artifact := range payload["artifacts"].([]any) {
		record := artifact.(map[string]any)
		if record["path"] == "model_provenance.json" {
			record["sha256"] = provenanceHash
			record["size_bytes"] = provenanceInfo.Size()
			writeJSON(t, manifestPath, payload)
			return
		}
	}
	t.Fatal("model_provenance.json artifact record not found")
}

func assertVerificationError(t *testing.T, err error, state verificationState, message string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected verification error")
	}
	var verifyErr *verificationError
	if !errors.As(err, &verifyErr) {
		t.Fatalf("expected verificationError, got %T", err)
	}
	if verifyErr.State != state {
		t.Fatalf("expected %s, got %s", state, verifyErr.State)
	}
	if !strings.Contains(verifyErr.Error(), message) {
		t.Fatalf("expected error containing %q, got %v", message, verifyErr)
	}
}

func mustReadJSON(t *testing.T, path string) map[string]any {
	t.Helper()
	var payload map[string]any
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}
	return payload
}

func writeJSON(t *testing.T, path string, payload any) {
	t.Helper()
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		t.Fatalf("marshal %s: %v", path, err)
	}
	mustWriteFile(t, path, append(data, '\n'))
}

func mustWriteFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func mustComputeHash(t *testing.T, path string) string {
	t.Helper()
	hash, err := computeFileHash(path)
	if err != nil {
		t.Fatalf("hash %s: %v", path, err)
	}
	return hash
}

func stringsOfLength(length int, char string) string {
	result := ""
	for len(result) < length {
		result += char
	}
	return result[:length]
}
