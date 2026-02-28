//go:build integration

package coordinator_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/virtengine/virtengine/tools/trusted-setup/coordinator"
	"github.com/virtengine/virtengine/tools/trusted-setup/participant"
	"github.com/virtengine/virtengine/tools/trusted-setup/testsupport"
	"github.com/virtengine/virtengine/tools/trusted-setup/transcript"
	"github.com/virtengine/virtengine/tools/trusted-setup/verify"
)

func TestCoordinatorServerFlowPersistsAcrossRestart(t *testing.T) {
	clock := testsupport.NewStepClock(time.Date(2026, 4, 10, 10, 0, 0, 0, time.UTC), time.Minute)
	state := coordinator.State{
		BaseDir: t.TempDir(),
		NowFunc: clock.Now,
	}
	if _, err := coordinator.InitCeremony(state, "age-range", 2, "integration-beacon", []string{"integration"}); err != nil {
		t.Fatalf("init ceremony: %v", err)
	}

	alice := participant.NewClient(participant.NewDeterministicIdentity("alice", "alice-seed"), "attestation-alice")
	bob := participant.NewClient(participant.NewDeterministicIdentity("bob", "bob-seed"), "attestation-bob")

	server := httptest.NewServer(coordinator.NewServer(state).Handler())
	if err := contributeViaHTTP(server.URL, transcript.Phase1, alice, "phase1-alice"); err != nil {
		t.Fatalf("alice phase1 contribution: %v", err)
	}
	server.Close()

	server = httptest.NewServer(coordinator.NewServer(state).Handler())
	t.Cleanup(server.Close)
	if err := contributeViaHTTP(server.URL, transcript.Phase1, bob, "phase1-bob"); err != nil {
		t.Fatalf("bob phase1 contribution: %v", err)
	}

	if err := coordinator.StartPhase2(state); err != nil {
		t.Fatalf("start phase2: %v", err)
	}
	if err := contributeViaHTTP(server.URL, transcript.Phase2, alice, "phase2-alice"); err != nil {
		t.Fatalf("alice phase2 contribution: %v", err)
	}
	if err := contributeViaHTTP(server.URL, transcript.Phase2, bob, "phase2-bob"); err != nil {
		t.Fatalf("bob phase2 contribution: %v", err)
	}

	if _, _, err := coordinator.Finalize(state, "v-integration"); err != nil {
		t.Fatalf("finalize ceremony: %v", err)
	}

	result, err := verify.Verify(state.BaseDir)
	if err != nil {
		t.Fatalf("verify ceremony: %v", err)
	}
	if !result.Phase1Valid || !result.Phase2Valid || !result.SignaturesValid {
		t.Fatalf("expected valid verification result, got %#v", result)
	}

	status, err := coordinator.StatusSnapshot(state)
	if err != nil {
		t.Fatalf("status snapshot: %v", err)
	}
	if status.Phase != "finalized" {
		t.Fatalf("expected finalized phase, got %s", status.Phase)
	}
}

func contributeViaHTTP(baseURL, phase string, client *participant.Client, randSeed string) error {
	payload, err := fetchCurrent(baseURL, phase)
	if err != nil {
		return err
	}

	var contribution []byte
	var signature string
	err = testsupport.WithDeterministicRand(randSeed, func() error {
		var contributeErr error
		if phase == transcript.Phase1 {
			contribution, signature, contributeErr = client.ContributePhase1(payload)
		} else {
			contribution, signature, contributeErr = client.ContributePhase2(payload)
		}
		return contributeErr
	})
	if err != nil {
		return err
	}

	endpoint := "/api/v1/phase1/contribute"
	if phase == transcript.Phase2 {
		endpoint = "/api/v1/phase2/contribute"
	}
	request, err := http.NewRequest(http.MethodPost, baseURL+endpoint, bytes.NewReader(contribution))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/octet-stream")
	request.Header.Set("X-Participant-Id", client.Identity.ID)
	request.Header.Set("X-Public-Key", client.Identity.PublicKey)
	request.Header.Set("X-Signature", signature)
	request.Header.Set("X-Attestation", client.Attestation)
	request.Header.Set("X-Input-Hash", transcript.HashBytes(payload))
	request.Header.Set("X-Output-Hash", transcript.HashBytes(contribution))

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("coordinator returned %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func fetchCurrent(baseURL, phase string) ([]byte, error) {
	endpoint := "/api/v1/phase1/current"
	if phase == transcript.Phase2 {
		endpoint = "/api/v1/phase2/current"
	}
	resp, err := http.Get(baseURL + endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
