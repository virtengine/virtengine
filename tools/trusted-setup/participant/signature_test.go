package participant

import (
	"testing"
)

func TestVerifyContributionSignatureRejectsTampering(t *testing.T) {
	identity := NewDeterministicIdentity("participant-1", "participant-1-seed")
	client := NewClient(identity, "attestation-1")

	input := []byte("phase-input")
	output := []byte("phase-output")
	signature, err := client.SignContribution("phase1", input, output)
	if err != nil {
		t.Fatalf("sign contribution: %v", err)
	}

	if err := VerifyContributionSignature("phase1", identity.ID, identity.PublicKey, client.Attestation, signature, input, output); err != nil {
		t.Fatalf("verify contribution signature: %v", err)
	}
	if err := VerifyContributionSignature("phase1", identity.ID, identity.PublicKey, client.Attestation, signature, input, []byte("tampered")); err == nil {
		t.Fatal("expected tampered contribution verification to fail")
	}
	if err := VerifyContributionSignature("phase1", identity.ID, identity.PublicKey, "tampered-attestation", signature, input, output); err == nil {
		t.Fatal("expected attestation mismatch to fail signature verification")
	}
}
