package coordinator

import (
	"testing"
	"time"

	"github.com/virtengine/virtengine/tools/trusted-setup/participant"
	"github.com/virtengine/virtengine/tools/trusted-setup/testsupport"
	"github.com/virtengine/virtengine/tools/trusted-setup/transcript"
)

func TestInitCeremonyAndAcceptContributionAreIdempotent(t *testing.T) {
	clock := testsupport.NewStepClock(time.Date(2026, 4, 10, 9, 0, 0, 0, time.UTC), time.Minute)
	state := State{
		BaseDir: t.TempDir(),
		NowFunc: clock.Now,
	}

	firstCfg, err := InitCeremony(state, "age-range", 2, "unit-beacon", []string{"unit"})
	if err != nil {
		t.Fatalf("first init: %v", err)
	}
	secondCfg, err := InitCeremony(state, "age-range", 2, "unit-beacon", []string{"unit"})
	if err != nil {
		t.Fatalf("second init: %v", err)
	}
	if firstCfg.CeremonyID != secondCfg.CeremonyID {
		t.Fatalf("expected idempotent init, got %s and %s", firstCfg.CeremonyID, secondCfg.CeremonyID)
	}

	_, initialPayload, err := loadLatestPhase1(state)
	if err != nil {
		t.Fatalf("load latest phase1: %v", err)
	}
	client := participant.NewClient(participant.NewDeterministicIdentity("alice", "alice-seed"), "attestation-alice")
	var contribution []byte
	var signature string
	if err := testsupport.WithDeterministicRand("phase1-alice", func() error {
		var contributeErr error
		contribution, signature, contributeErr = client.ContributePhase1(initialPayload)
		return contributeErr
	}); err != nil {
		t.Fatalf("create contribution: %v", err)
	}
	meta := ContributionMeta{
		ParticipantID: client.Identity.ID,
		PublicKey:     client.Identity.PublicKey,
		Signature:     signature,
		Attestation:   client.Attestation,
		InputHash:     transcript.HashBytes(initialPayload),
		OutputHash:    transcript.HashBytes(contribution),
	}

	if err := AcceptPhase1Contribution(state, contribution, meta); err != nil {
		t.Fatalf("accept phase1 contribution: %v", err)
	}
	if err := AcceptPhase1Contribution(state, contribution, meta); err != nil {
		t.Fatalf("idempotent phase1 accept: %v", err)
	}

	tr, err := loadTranscript(state)
	if err != nil {
		t.Fatalf("load transcript: %v", err)
	}
	if got := len(tr.Phase1.Contributions); got != 1 {
		t.Fatalf("expected one recorded phase1 contribution after duplicate accept, got %d", got)
	}
}
