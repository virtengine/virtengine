package verify

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/virtengine/virtengine/tools/trusted-setup/bundle"
	"github.com/virtengine/virtengine/tools/trusted-setup/participant"
)

func TestVerifyExportRejectsTamperedArtifact(t *testing.T) {
	dir := t.TempDir()
	signer := participant.NewDeterministicIdentity("coordinator", "export-seed")

	artifactPath := filepath.Join(dir, "artifact.bin")
	if err := os.WriteFile(artifactPath, []byte("artifact"), 0o600); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	report := Result{TranscriptHash: "transcript-hash"}
	if err := bundle.WriteJSON(filepath.Join(dir, verificationReportFile), &report); err != nil {
		t.Fatalf("write verification report: %v", err)
	}

	artifactRecord, err := bundle.BuildFileRecord(dir, artifactPath)
	if err != nil {
		t.Fatalf("build artifact record: %v", err)
	}
	reportRecord, err := bundle.BuildFileRecord(dir, filepath.Join(dir, verificationReportFile))
	if err != nil {
		t.Fatalf("build report record: %v", err)
	}

	manifest := &bundle.ArtifactManifest{
		SchemaVersion:        bundle.ArtifactManifestSchema,
		CeremonyID:           "ceremony-1",
		CircuitName:          "age-range",
		TranscriptHash:       "transcript-hash",
		ParametersVersion:    "v-test",
		ContributorCount:     2,
		Contributors:         []string{"alice", "bob"},
		CoordinatorID:        signer.ID,
		CoordinatorPublicKey: signer.PublicKey,
		SignatureAlgorithm:   "ed25519",
		VerificationReport:   verificationReportFile,
		Files:                []bundle.FileRecord{artifactRecord, reportRecord},
	}
	manifestPath := filepath.Join(dir, artifactManifestFile)
	manifestBytes, err := bundle.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	if err := os.WriteFile(manifestPath, manifestBytes, 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	signature, err := signer.Sign(manifestBytes)
	if err != nil {
		t.Fatalf("sign manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, artifactSignatureFile), []byte(signature+"\n"), 0o600); err != nil {
		t.Fatalf("write signature: %v", err)
	}

	if _, err := VerifyExport(dir); err != nil {
		t.Fatalf("verify export: %v", err)
	}

	if err := os.WriteFile(artifactPath, []byte("tampered"), 0o600); err != nil {
		t.Fatalf("tamper artifact: %v", err)
	}
	if _, err := VerifyExport(dir); err == nil {
		t.Fatal("expected tampered artifact export verification to fail")
	}
}
