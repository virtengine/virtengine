package exporter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/virtengine/virtengine/tools/trusted-setup/bundle"
	"github.com/virtengine/virtengine/tools/trusted-setup/coordinator"
	"github.com/virtengine/virtengine/tools/trusted-setup/participant"
	"github.com/virtengine/virtengine/tools/trusted-setup/transcript"
	"github.com/virtengine/virtengine/tools/trusted-setup/verify"
)

const (
	ArtifactManifestFile   = "artifact-manifest.json"
	ArtifactSignatureFile  = "artifact-manifest.sig"
	VerificationReportFile = "verification.json"
)

func ExportArtifacts(state coordinator.State, outDir string, signer *participant.Identity) (*bundle.ArtifactManifest, error) {
	if signer == nil {
		return nil, fmt.Errorf("signer identity is required")
	}
	result, err := verify.Verify(state.BaseDir)
	if err != nil {
		return nil, err
	}
	if !result.Phase1Valid || !result.Phase2Valid {
		return nil, fmt.Errorf("ceremony verification failed")
	}
	if result.Transcript == nil || result.Transcript.Final.VerifyingKeyHash == "" || result.Transcript.Final.ProvingKeyHash == "" {
		return nil, fmt.Errorf("ceremony is not finalized")
	}
	if err := prepareExportDir(state.BaseDir, outDir); err != nil {
		return nil, err
	}

	copyList := []string{
		"config.json",
		"transcript.json",
		filepath.ToSlash(filepath.Join("phase1", "commons.bin")),
		filepath.ToSlash(filepath.Join("phase2", "r1cs.bin")),
		filepath.ToSlash(filepath.Join("phase2", "proving_key.bin")),
		filepath.ToSlash(filepath.Join("phase2", "verifying_key.bin")),
	}
	copyList = append(copyList, relativeFiles(state.BaseDir, result.Phase1Files)...)
	copyList = append(copyList, relativeFiles(state.BaseDir, result.Phase2Files)...)
	sort.Strings(copyList)

	for _, rel := range copyList {
		src := filepath.Join(state.BaseDir, filepath.FromSlash(rel))
		dst := filepath.Join(outDir, filepath.FromSlash(rel))
		if err := copyFile(src, dst); err != nil {
			return nil, err
		}
	}

	reportPath := filepath.Join(outDir, VerificationReportFile)
	reportResult := *result
	reportResult.Phase1Files = relativeFiles(state.BaseDir, result.Phase1Files)
	reportResult.Phase2Files = relativeFiles(state.BaseDir, result.Phase2Files)
	sort.Strings(reportResult.Phase1Files)
	sort.Strings(reportResult.Phase2Files)

	reportBytes, err := json.MarshalIndent(&reportResult, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(reportPath, append(reportBytes, '\n'), 0o600); err != nil {
		return nil, err
	}

	fileRecords := make([]bundle.FileRecord, 0, len(copyList)+1)
	for _, rel := range copyList {
		record, err := bundle.BuildFileRecord(outDir, filepath.Join(outDir, filepath.FromSlash(rel)))
		if err != nil {
			return nil, err
		}
		fileRecords = append(fileRecords, record)
	}
	reportRecord, err := bundle.BuildFileRecord(outDir, reportPath)
	if err != nil {
		return nil, err
	}
	fileRecords = append(fileRecords, reportRecord)
	sort.Slice(fileRecords, func(i, j int) bool {
		return fileRecords[i].Path < fileRecords[j].Path
	})

	contributors := contributorIDs(result.Transcript)
	manifest := &bundle.ArtifactManifest{
		SchemaVersion:        bundle.ArtifactManifestSchema,
		CeremonyID:           result.Transcript.CeremonyID,
		CircuitName:          result.Transcript.CircuitName,
		TranscriptHash:       result.TranscriptHash,
		ParametersVersion:    result.Transcript.Final.ParametersVersion,
		ContributorCount:     len(contributors),
		Contributors:         contributors,
		CoordinatorID:        signer.ID,
		CoordinatorPublicKey: signer.PublicKey,
		SignatureAlgorithm:   "ed25519",
		VerificationReport:   VerificationReportFile,
		Files:                fileRecords,
	}

	manifestPath := filepath.Join(outDir, ArtifactManifestFile)
	manifestBytes, err := bundle.Marshal(manifest)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(manifestPath, manifestBytes, 0o600); err != nil {
		return nil, err
	}
	signature, err := signer.Sign(manifestBytes)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(outDir, ArtifactSignatureFile), []byte(signature+"\n"), 0o600); err != nil {
		return nil, err
	}
	if err := bundle.WriteDigestFile(filepath.Join(outDir, ArtifactManifestFile+".sha256"), manifestPath); err != nil {
		return nil, err
	}
	if err := bundle.WriteDigestFile(filepath.Join(outDir, VerificationReportFile+".sha256"), reportPath); err != nil {
		return nil, err
	}
	if err := bundle.WriteDigestFile(filepath.Join(outDir, ArtifactSignatureFile+".sha256"), filepath.Join(outDir, ArtifactSignatureFile)); err != nil {
		return nil, err
	}

	return manifest, nil
}

func prepareExportDir(baseDir, outDir string) error {
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return err
	}
	absOut, err := filepath.Abs(outDir)
	if err != nil {
		return err
	}
	if rel, err := filepath.Rel(absBase, absOut); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("export directory must be outside ceremony state directory")
	}
	if entries, err := os.ReadDir(outDir); err == nil && len(entries) > 0 {
		return fmt.Errorf("export directory must be empty: %s", outDir)
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}
	return os.MkdirAll(outDir, 0o750)
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o600)
}

func contributorIDs(tr *transcript.Transcript) []string {
	seen := map[string]struct{}{}
	contributors := make([]string, 0, len(tr.Phase1.Contributions)+len(tr.Phase2.Contributions))
	for _, record := range append(append([]transcript.ContributionRecord{}, tr.Phase1.Contributions...), tr.Phase2.Contributions...) {
		if record.ParticipantID == "" {
			continue
		}
		if _, exists := seen[record.ParticipantID]; exists {
			continue
		}
		seen[record.ParticipantID] = struct{}{}
		contributors = append(contributors, record.ParticipantID)
	}
	sort.Strings(contributors)
	return contributors
}

func relativeFiles(baseDir string, paths []string) []string {
	result := make([]string, 0, len(paths))
	for _, path := range paths {
		if rel, err := filepath.Rel(baseDir, path); err == nil {
			result = append(result, filepath.ToSlash(rel))
		}
	}
	return result
}
