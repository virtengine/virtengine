package verify

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/virtengine/virtengine/tools/trusted-setup/bundle"
	"github.com/virtengine/virtengine/tools/trusted-setup/participant"
)

const (
	artifactManifestFile   = "artifact-manifest.json"
	artifactSignatureFile  = "artifact-manifest.sig"
	verificationReportFile = "verification.json"
)

type ExportResult struct {
	Manifest       *bundle.ArtifactManifest `json:"manifest"`
	ManifestHash   string                   `json:"manifest_hash"`
	SignatureValid bool                     `json:"signature_valid"`
	FilesValid     bool                     `json:"files_valid"`
	Verification   *Result                  `json:"verification,omitempty"`
}

func VerifyExport(dir string) (*ExportResult, error) {
	manifestPath := filepath.Join(dir, artifactManifestFile)
	signaturePath := filepath.Join(dir, artifactSignatureFile)
	reportPath := filepath.Join(dir, verificationReportFile)

	var manifest bundle.ArtifactManifest
	if err := bundle.ReadJSON(manifestPath, &manifest); err != nil {
		return nil, err
	}
	if manifest.SchemaVersion != bundle.ArtifactManifestSchema {
		return nil, fmt.Errorf("unsupported artifact manifest schema %q", manifest.SchemaVersion)
	}

	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}
	signatureBytes, err := os.ReadFile(signaturePath)
	if err != nil {
		return nil, err
	}
	signature := string(signatureBytes)
	signature = string([]byte(signature))
	if len(signature) > 0 && signature[len(signature)-1] == '\n' {
		signature = signature[:len(signature)-1]
	}
	if err := participant.VerifySignature(manifest.CoordinatorPublicKey, manifestBytes, signature); err != nil {
		return nil, fmt.Errorf("verify export signature: %w", err)
	}

	reportData, err := os.ReadFile(reportPath)
	if err != nil {
		return nil, err
	}
	var verification Result
	if err := json.Unmarshal(reportData, &verification); err != nil {
		return nil, err
	}
	if verification.TranscriptHash != manifest.TranscriptHash {
		return nil, fmt.Errorf("verification transcript hash mismatch: %s != %s", verification.TranscriptHash, manifest.TranscriptHash)
	}

	files := append([]bundle.FileRecord(nil), manifest.Files...)
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})
	for _, file := range files {
		target := filepath.Join(dir, filepath.FromSlash(file.Path))
		hash, size, err := bundle.HashFile(target)
		if err != nil {
			return nil, err
		}
		if hash != file.SHA256 {
			return nil, fmt.Errorf("artifact hash mismatch for %s: %s != %s", file.Path, file.SHA256, hash)
		}
		if size != file.SizeBytes {
			return nil, fmt.Errorf("artifact size mismatch for %s: %d != %d", file.Path, file.SizeBytes, size)
		}
	}

	return &ExportResult{
		Manifest:       &manifest,
		ManifestHash:   bundle.HashBytes(manifestBytes),
		SignatureValid: true,
		FilesValid:     true,
		Verification:   &verification,
	}, nil
}
