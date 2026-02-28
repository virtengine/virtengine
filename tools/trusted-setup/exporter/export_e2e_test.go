//go:build e2e.integration

package exporter

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/virtengine/virtengine/tools/trusted-setup/coordinator"
	"github.com/virtengine/virtengine/tools/trusted-setup/participant"
	"github.com/virtengine/virtengine/tools/trusted-setup/testsupport"
	"github.com/virtengine/virtengine/tools/trusted-setup/transcript"
	"github.com/virtengine/virtengine/tools/trusted-setup/verify"
)

func TestDeterministicAirGappedCeremonyProducesReproducibleArtifacts(t *testing.T) {
	root := t.TempDir()
	runA, err := runDeterministicCeremony(filepath.Join(root, "run-a"))
	if err != nil {
		t.Fatalf("run A: %v", err)
	}
	runB, err := runDeterministicCeremony(filepath.Join(root, "run-b"))
	if err != nil {
		t.Fatalf("run B: %v", err)
	}

	for _, relativePath := range []string{
		"state/transcript.json",
		"export/artifact-manifest.json",
		"export/verification.json",
	} {
		aBytes, err := os.ReadFile(filepath.Join(runA.root, relativePath))
		if err != nil {
			t.Fatalf("read run A %s: %v", relativePath, err)
		}
		bBytes, err := os.ReadFile(filepath.Join(runB.root, relativePath))
		if err != nil {
			t.Fatalf("read run B %s: %v", relativePath, err)
		}
		if string(aBytes) != string(bBytes) {
			t.Fatalf("expected reproducible %s", relativePath)
		}
	}

	if runA.transcriptHash != runB.transcriptHash || runA.verifyingKeyHash != runB.verifyingKeyHash || runA.provingKeyHash != runB.provingKeyHash {
		t.Fatalf("expected reproducible hashes, got A=%#v B=%#v", runA, runB)
	}
}

type deterministicRun struct {
	root             string
	transcriptHash   string
	verifyingKeyHash string
	provingKeyHash   string
}

func runDeterministicCeremony(root string) (*deterministicRun, error) {
	stateDir := filepath.Join(root, "state")
	exportDir := filepath.Join(root, "export")
	clock := testsupport.NewStepClock(time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC), time.Minute)
	state := coordinator.State{
		BaseDir: stateDir,
		NowFunc: clock.Now,
	}
	if _, err := coordinator.InitCeremony(state, "score-range", 2, "deterministic-beacon", []string{"e2e"}); err != nil {
		return nil, err
	}

	alice := participant.NewClient(participant.NewDeterministicIdentity("alice", "alice-seed"), "attestation-alice")
	bob := participant.NewClient(participant.NewDeterministicIdentity("bob", "bob-seed"), "attestation-bob")

	if err := contributeOffline(state, transcript.Phase1, filepath.Join(root, "phase1-alice-request"), filepath.Join(root, "phase1-alice-response"), alice, "phase1-alice"); err != nil {
		return nil, err
	}
	if err := contributeOffline(state, transcript.Phase1, filepath.Join(root, "phase1-bob-request"), filepath.Join(root, "phase1-bob-response"), bob, "phase1-bob"); err != nil {
		return nil, err
	}
	if err := coordinator.StartPhase2(state); err != nil {
		return nil, err
	}
	if err := contributeOffline(state, transcript.Phase2, filepath.Join(root, "phase2-alice-request"), filepath.Join(root, "phase2-alice-response"), alice, "phase2-alice"); err != nil {
		return nil, err
	}
	if err := contributeOffline(state, transcript.Phase2, filepath.Join(root, "phase2-bob-request"), filepath.Join(root, "phase2-bob-response"), bob, "phase2-bob"); err != nil {
		return nil, err
	}

	if _, _, err := coordinator.Finalize(state, "v-e2e"); err != nil {
		return nil, err
	}

	verifyResult, err := verify.Verify(state.BaseDir)
	if err != nil {
		return nil, err
	}
	if !verifyResult.Phase1Valid || !verifyResult.Phase2Valid || !verifyResult.SignaturesValid {
		return nil, fmt.Errorf("unexpected invalid verification result: %#v", verifyResult)
	}

	signer := participant.NewDeterministicIdentity("coordinator", "coordinator-seed")
	if _, err := ExportArtifacts(state, exportDir, signer); err != nil {
		return nil, err
	}
	exportResult, err := verify.VerifyExport(exportDir)
	if err != nil {
		return nil, err
	}
	if !exportResult.SignatureValid || !exportResult.FilesValid {
		return nil, fmt.Errorf("unexpected invalid export verification result: %#v", exportResult)
	}

	return &deterministicRun{
		root:             root,
		transcriptHash:   verifyResult.TranscriptHash,
		verifyingKeyHash: verifyResult.VerifyingKeyHash,
		provingKeyHash:   verifyResult.ProvingKeyHash,
	}, nil
}

func contributeOffline(state coordinator.State, phase, requestDir, responseDir string, client *participant.Client, seed string) error {
	if _, err := coordinator.ExportPhaseBundle(state, phase, requestDir); err != nil {
		return err
	}
	if err := testsupport.WithDeterministicRand(seed, func() error {
		_, err := participant.RespondToPhaseBundle(requestDir, responseDir, client)
		return err
	}); err != nil {
		return err
	}
	return coordinator.AcceptPhaseBundle(state, phase, responseDir)
}
