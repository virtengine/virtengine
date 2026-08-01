package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type verificationState string

const (
	verificationStateVerified        verificationState = "verified"
	verificationStateMissingManifest verificationState = "missing_manifest"
	verificationStateMissingModel    verificationState = "missing_model"
	verificationStateBadManifest     verificationState = "bad_manifest"
	verificationStateStaleManifest   verificationState = "stale_manifest"
)

var placeholderPattern = regexp.MustCompile(`(?i)(placeholder|pending|tbd|not published yet|<path>|sha256:placeholder)`)

var requiredManifestArtifacts = []string{
	"MODEL_HASH.txt",
	"export_metadata.json",
	"manifest.json",
	"model_provenance.json",
	"model_frozen.pb",
	filepath.ToSlash(filepath.Join("model", "saved_model.pb")),
}

type verificationError struct {
	State verificationState
	Path  string
	Err   error
}

func (e *verificationError) Error() string {
	if e.Path == "" {
		return fmt.Sprintf("%s: %v", e.State, e.Err)
	}
	return fmt.Sprintf("%s (%s): %v", e.State, e.Path, e.Err)
}

func (e *verificationError) Unwrap() error {
	return e.Err
}

type releaseManifest struct {
	SchemaVersion string `json:"schema_version"`
	Profile       string `json:"profile"`
	Provenance    struct {
		Path   string `json:"path"`
		SHA256 string `json:"sha256"`
	} `json:"provenance"`
	Model struct {
		Name            string                 `json:"name"`
		Version         string                 `json:"version"`
		ModelDir        string                 `json:"model_dir"`
		RuntimeHash     string                 `json:"runtime_hash"`
		FrozenGraphHash string                 `json:"frozen_graph_hash"`
		SignatureName   string                 `json:"signature_name"`
		InputSignature  map[string]interface{} `json:"input_signature"`
		OutputSignature map[string]interface{} `json:"output_signature"`
	} `json:"model"`
	Artifacts []releaseArtifact `json:"artifacts"`
}

type releaseArtifact struct {
	Path      string `json:"path"`
	SHA256    string `json:"sha256"`
	SizeBytes int64  `json:"size_bytes"`
}

type modelProvenance struct {
	SchemaVersion string `json:"schema_version"`
	Status        string `json:"status"`
}

type verificationResult struct {
	State         verificationState
	ManifestPath  string
	Manifest      *releaseManifest
	ModelHash     string
	Version       string
	FailurePath   string
	FailureReason string
}

func (r *verificationResult) Ready() bool {
	return r != nil && r.State == verificationStateVerified
}

func (r *verificationResult) StatusMessage() string {
	if r == nil {
		return "verification state unavailable"
	}
	if r.State == verificationStateVerified {
		return "model bundle verified"
	}
	if r.FailurePath != "" && r.FailureReason != "" {
		return fmt.Sprintf("%s: %s (%s)", r.State, r.FailureReason, r.FailurePath)
	}
	if r.FailureReason != "" {
		return fmt.Sprintf("%s: %s", r.State, r.FailureReason)
	}
	return string(r.State)
}

func newVerificationFailureResult(manifestPath string, err error) *verificationResult {
	result := &verificationResult{
		State:        verificationStateBadManifest,
		ManifestPath: manifestPath,
	}

	var verifyErr *verificationError
	if errors.As(err, &verifyErr) {
		result.State = verifyErr.State
		result.FailurePath = verifyErr.Path
		if verifyErr.Err != nil {
			result.FailureReason = verifyErr.Err.Error()
		} else {
			result.FailureReason = verifyErr.Error()
		}
		return result
	}

	if err != nil {
		result.FailureReason = err.Error()
	}
	return result
}

func deriveManifestPath(modelPath, explicitManifestPath string) string {
	if strings.TrimSpace(explicitManifestPath) != "" {
		return explicitManifestPath
	}
	return filepath.Join(filepath.Dir(filepath.Clean(modelPath)), "release_manifest.json")
}

func resolveBundlePath(bundleRoot, relativePath, manifestPath, fieldName string) (string, string, error) {
	absBundleRoot, err := filepath.Abs(filepath.Clean(bundleRoot))
	if err != nil {
		return "", "", &verificationError{State: verificationStateBadManifest, Path: manifestPath, Err: err}
	}

	absTarget, err := filepath.Abs(filepath.Join(absBundleRoot, filepath.Clean(relativePath)))
	if err != nil {
		return "", "", &verificationError{State: verificationStateBadManifest, Path: manifestPath, Err: err}
	}

	relativeTarget, err := filepath.Rel(absBundleRoot, absTarget)
	if err != nil {
		return "", "", &verificationError{State: verificationStateBadManifest, Path: manifestPath, Err: err}
	}
	if relativeTarget == ".." || strings.HasPrefix(relativeTarget, ".."+string(filepath.Separator)) {
		return "", "", &verificationError{
			State: verificationStateBadManifest,
			Path:  manifestPath,
			Err:   fmt.Errorf("%s resolves outside bundle root: %s", fieldName, relativePath),
		}
	}

	return absTarget, filepath.ToSlash(relativeTarget), nil
}

func verifyModelBundle(modelPath, manifestPath, expectedVersion, expectedHash string) (*verificationResult, error) {
	manifestPath = deriveManifestPath(modelPath, manifestPath)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		state := verificationStateBadManifest
		if os.IsNotExist(err) {
			state = verificationStateMissingManifest
		}
		return nil, &verificationError{State: state, Path: manifestPath, Err: err}
	}
	if placeholderPattern.Match(data) {
		return nil, &verificationError{
			State: verificationStateBadManifest,
			Path:  manifestPath,
			Err:   fmt.Errorf("manifest contains placeholder content"),
		}
	}

	var manifest releaseManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, &verificationError{State: verificationStateBadManifest, Path: manifestPath, Err: err}
	}
	if manifest.SchemaVersion != "veid.release.manifest/v1" {
		return nil, &verificationError{
			State: verificationStateBadManifest,
			Path:  manifestPath,
			Err:   fmt.Errorf("unsupported schema_version %q", manifest.SchemaVersion),
		}
	}

	runtimeHash := normalizeHash(manifest.Model.RuntimeHash)
	if !isValidHash(runtimeHash) {
		return nil, &verificationError{
			State: verificationStateBadManifest,
			Path:  manifestPath,
			Err:   fmt.Errorf("manifest runtime_hash must be a valid SHA-256 digest"),
		}
	}

	if strings.TrimSpace(manifest.Model.Version) == "" || strings.TrimSpace(manifest.Model.ModelDir) == "" {
		return nil, &verificationError{
			State: verificationStateBadManifest,
			Path:  manifestPath,
			Err:   fmt.Errorf("manifest model version and model_dir are required"),
		}
	}
	if strings.TrimSpace(manifest.Model.SignatureName) == "" {
		return nil, &verificationError{
			State: verificationStateBadManifest,
			Path:  manifestPath,
			Err:   fmt.Errorf("manifest signature_name is required"),
		}
	}

	if len(manifest.Model.InputSignature) == 0 || len(manifest.Model.OutputSignature) == 0 {
		return nil, &verificationError{
			State: verificationStateBadManifest,
			Path:  manifestPath,
			Err:   fmt.Errorf("manifest input/output signatures are required"),
		}
	}

	manifestDir := filepath.Dir(manifestPath)
	absManifestModelPath, _, err := resolveBundlePath(manifestDir, manifest.Model.ModelDir, manifestPath, "manifest model_dir")
	if err != nil {
		return nil, err
	}
	absModelPath, err := filepath.Abs(filepath.Clean(modelPath))
	if err != nil {
		return nil, &verificationError{State: verificationStateBadManifest, Path: modelPath, Err: err}
	}
	if absManifestModelPath != absModelPath {
		return nil, &verificationError{
			State: verificationStateStaleManifest,
			Path:  manifestPath,
			Err:   fmt.Errorf("manifest model_dir resolves to %s, expected %s", absManifestModelPath, absModelPath),
		}
	}

	if expectedVersion != "" && manifest.Model.Version != expectedVersion {
		return nil, &verificationError{
			State: verificationStateStaleManifest,
			Path:  manifestPath,
			Err:   fmt.Errorf("manifest version mismatch: expected %s, got %s", expectedVersion, manifest.Model.Version),
		}
	}

	if expectedHash != "" {
		normalizedExpectedHash := normalizeHash(expectedHash)
		if !isValidHash(normalizedExpectedHash) {
			return nil, &verificationError{
				State: verificationStateBadManifest,
				Path:  manifestPath,
				Err:   fmt.Errorf("expected hash must be a valid SHA-256 digest"),
			}
		}
		if normalizedExpectedHash != runtimeHash {
			return nil, &verificationError{
				State: verificationStateStaleManifest,
				Path:  manifestPath,
				Err:   fmt.Errorf("manifest runtime hash mismatch: expected %s, got %s", normalizedExpectedHash, runtimeHash),
			}
		}
	}

	computedModelHash, err := computeModelDirHash(absModelPath)
	if err != nil {
		state := verificationStateBadManifest
		if os.IsNotExist(err) {
			state = verificationStateMissingModel
		}
		return nil, &verificationError{State: state, Path: absModelPath, Err: err}
	}
	if computedModelHash != runtimeHash {
		return nil, &verificationError{
			State: verificationStateStaleManifest,
			Path:  absModelPath,
			Err:   fmt.Errorf("model hash mismatch: expected %s, got %s", runtimeHash, computedModelHash),
		}
	}

	if len(manifest.Artifacts) == 0 {
		return nil, &verificationError{
			State: verificationStateBadManifest,
			Path:  manifestPath,
			Err:   fmt.Errorf("manifest artifacts list is empty"),
		}
	}

	verifiedArtifacts := make(map[string]string, len(manifest.Artifacts))
	for _, artifact := range manifest.Artifacts {
		if strings.TrimSpace(artifact.Path) == "" {
			return nil, &verificationError{
				State: verificationStateBadManifest,
				Path:  manifestPath,
				Err:   fmt.Errorf("artifact path cannot be empty"),
			}
		}
		artifactHash := normalizeHash(artifact.SHA256)
		if !isValidHash(artifactHash) {
			return nil, &verificationError{
				State: verificationStateBadManifest,
				Path:  artifact.Path,
				Err:   fmt.Errorf("artifact hash must be a valid SHA-256 digest"),
			}
		}

		if artifact.SizeBytes < 0 {
			return nil, &verificationError{
				State: verificationStateBadManifest,
				Path:  manifestPath,
				Err:   fmt.Errorf("artifact size cannot be negative for %s", artifact.Path),
			}
		}

		artifactPath, normalizedArtifactPath, err := resolveBundlePath(manifestDir, artifact.Path, manifestPath, "artifact path")
		if err != nil {
			return nil, err
		}
		if _, exists := verifiedArtifacts[normalizedArtifactPath]; exists {
			return nil, &verificationError{
				State: verificationStateBadManifest,
				Path:  manifestPath,
				Err:   fmt.Errorf("duplicate artifact entry %s", normalizedArtifactPath),
			}
		}
		verifiedArtifacts[normalizedArtifactPath] = artifactHash

		info, err := os.Stat(artifactPath)
		if err != nil {
			state := verificationStateBadManifest
			if os.IsNotExist(err) {
				state = verificationStateMissingModel
			}
			return nil, &verificationError{State: state, Path: artifactPath, Err: err}
		}
		if info.IsDir() {
			return nil, &verificationError{
				State: verificationStateBadManifest,
				Path:  artifactPath,
				Err:   fmt.Errorf("artifact must reference a file"),
			}
		}

		computedArtifactHash, err := computeFileHash(artifactPath)
		if err != nil {
			return nil, &verificationError{State: verificationStateBadManifest, Path: artifactPath, Err: err}
		}
		if computedArtifactHash != artifactHash {
			return nil, &verificationError{
				State: verificationStateStaleManifest,
				Path:  artifactPath,
				Err:   fmt.Errorf("artifact hash mismatch: expected %s, got %s", artifactHash, computedArtifactHash),
			}
		}

		if info.Size() != artifact.SizeBytes {
			return nil, &verificationError{
				State: verificationStateStaleManifest,
				Path:  artifactPath,
				Err:   fmt.Errorf("artifact size mismatch: expected %d, got %d", artifact.SizeBytes, info.Size()),
			}
		}
	}

	for _, requiredArtifact := range requiredManifestArtifacts {
		if _, exists := verifiedArtifacts[requiredArtifact]; !exists {
			return nil, &verificationError{
				State: verificationStateBadManifest,
				Path:  manifestPath,
				Err:   fmt.Errorf("manifest artifacts missing required entry %s", requiredArtifact),
			}
		}
	}

	provenancePath := strings.TrimSpace(manifest.Provenance.Path)
	if provenancePath == "" {
		return nil, &verificationError{
			State: verificationStateBadManifest,
			Path:  manifestPath,
			Err:   fmt.Errorf("manifest provenance path is required"),
		}
	}
	if filepath.IsAbs(provenancePath) || filepath.VolumeName(provenancePath) != "" {
		return nil, &verificationError{
			State: verificationStateBadManifest,
			Path:  manifestPath,
			Err:   fmt.Errorf("manifest provenance path must be repository-relative: %s", provenancePath),
		}
	}
	provenanceHash := normalizeHash(manifest.Provenance.SHA256)
	if !isValidHash(provenanceHash) {
		return nil, &verificationError{
			State: verificationStateBadManifest,
			Path:  manifestPath,
			Err:   fmt.Errorf("manifest provenance hash must be a valid SHA-256 digest"),
		}
	}

	absProvenancePath, normalizedProvenancePath, err := resolveBundlePath(
		manifestDir,
		provenancePath,
		manifestPath,
		"manifest provenance path",
	)
	if err != nil {
		return nil, err
	}
	if normalizedProvenancePath != "model_provenance.json" {
		return nil, &verificationError{
			State: verificationStateBadManifest,
			Path:  manifestPath,
			Err:   fmt.Errorf("manifest provenance path must reference model_provenance.json"),
		}
	}
	artifactHash, exists := verifiedArtifacts[normalizedProvenancePath]
	if !exists {
		return nil, &verificationError{
			State: verificationStateBadManifest,
			Path:  manifestPath,
			Err:   fmt.Errorf("manifest provenance path is not a verified artifact: %s", normalizedProvenancePath),
		}
	}
	if provenanceHash != artifactHash {
		return nil, &verificationError{
			State: verificationStateStaleManifest,
			Path:  absProvenancePath,
			Err:   fmt.Errorf("provenance digest mismatch: expected %s, got %s", artifactHash, provenanceHash),
		}
	}

	provenanceData, err := os.ReadFile(absProvenancePath)
	if err != nil {
		return nil, &verificationError{State: verificationStateBadManifest, Path: absProvenancePath, Err: err}
	}
	var provenance modelProvenance
	if err := json.Unmarshal(provenanceData, &provenance); err != nil {
		return nil, &verificationError{State: verificationStateBadManifest, Path: absProvenancePath, Err: err}
	}
	if provenance.SchemaVersion != "virtengine.model-provenance/v1" {
		return nil, &verificationError{
			State: verificationStateBadManifest,
			Path:  absProvenancePath,
			Err:   fmt.Errorf("unsupported provenance schema_version %q", provenance.SchemaVersion),
		}
	}
	if manifest.Profile == "production" && provenance.Status != "production_approved" {
		return nil, &verificationError{
			State: verificationStateBadManifest,
			Path:  absProvenancePath,
			Err:   fmt.Errorf("production manifest requires production_approved provenance status, got %q", provenance.Status),
		}
	}

	return &verificationResult{
		State:         verificationStateVerified,
		ManifestPath:  manifestPath,
		Manifest:      &manifest,
		ModelHash:     runtimeHash,
		Version:       manifest.Model.Version,
		FailurePath:   "",
		FailureReason: "",
	}, nil
}

func computeModelDirHash(modelPath string) (string, error) {
	hasher := sha256.New()
	var files []string

	err := filepath.Walk(modelPath, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Base(path) == "export_metadata.json" {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return "", err
	}

	sort.Strings(files)
	for _, path := range files {
		file, err := os.Open(path)
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(hasher, file); err != nil {
			_ = file.Close()
			return "", err
		}
		_ = file.Close()
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func computeFileHash(path string) (string, error) {
	hasher := sha256.New()
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func normalizeHash(value string) string {
	return strings.TrimPrefix(strings.ToLower(strings.TrimSpace(value)), "sha256:")
}

func isValidHash(value string) bool {
	return regexp.MustCompile(`^[a-f0-9]{64}$`).MatchString(value)
}
