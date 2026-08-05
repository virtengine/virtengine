package orchestration

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

func TestValidateReceiptRequestBindsEvidence(t *testing.T) {
	now := time.Date(2026, 8, 5, 10, 0, 0, 0, time.UTC)
	evidence := validEvidence(now)
	receipt, privateKey := validReceipt(now, evidence)
	require.NoError(t, receipt.Sign(privateKey))
	require.NoError(t, ValidateReceiptRequest(ReceiptRequest{Receipt: receipt, Evidence: evidence}, now))

	evidence.CaptureEvidenceDigest[0] ^= 0xff
	require.ErrorContains(t, ValidateReceiptRequest(ReceiptRequest{Receipt: receipt, Evidence: evidence}, now), "lineage digest")
}

func TestValidateReceiptRequestFailsClosedOnUnverifiedDevice(t *testing.T) {
	now := time.Date(2026, 8, 5, 10, 0, 0, 0, time.UTC)
	evidence := validEvidence(now)
	receipt, privateKey := validReceipt(now, evidence)
	require.NoError(t, receipt.Sign(privateKey))
	evidence.DeviceAttestation.Verified = false
	require.ErrorContains(t, ValidateReceiptRequest(ReceiptRequest{Receipt: receipt, Evidence: evidence}, now), "supported and verified")
}

func TestVerifyAndPrepareRequiresSignerVerification(t *testing.T) {
	now := time.Date(2026, 8, 5, 10, 0, 0, 0, time.UTC)
	evidence := validEvidence(now)
	receipt, privateKey := validReceipt(now, evidence)
	require.NoError(t, receipt.Sign(privateKey))
	_, err := VerifyAndPrepare(context.Background(), receiptVerifierFunc(func(context.Context, veidtypes.InferenceReceipt) error { return fmt.Errorf("unknown signer") }), ReceiptRequest{Receipt: receipt, Evidence: evidence}, now)
	require.ErrorContains(t, err, "signer verification failed")
}

func TestVerifyAndPrepareReturnsDefensiveReceiptCopy(t *testing.T) {
	now := time.Date(2026, 8, 5, 10, 0, 0, 0, time.UTC)
	evidence := validEvidence(now)
	receipt, privateKey := validReceipt(now, evidence)
	require.NoError(t, receipt.Sign(privateKey))
	prepared, err := VerifyAndPrepare(context.Background(), receiptVerifierFunc(func(context.Context, veidtypes.InferenceReceipt) error { return nil }), ReceiptRequest{Receipt: receipt, Evidence: evidence}, now)
	require.NoError(t, err)
	receipt.InputDigest[0] ^= 0xff
	receipt.ScopeIDs[0] = "mutated"
	require.NotEqual(t, receipt.InputDigest, prepared.InputDigest)
	require.Equal(t, "scope-1", prepared.ScopeIDs[0])
}

type receiptVerifierFunc func(context.Context, veidtypes.InferenceReceipt) error

func (f receiptVerifierFunc) VerifyInferenceReceipt(ctx context.Context, receipt veidtypes.InferenceReceipt) error {
	return f(ctx, receipt)
}

func validEvidence(now time.Time) EvidenceBinding {
	capture := sha256.Sum256([]byte("capture"))
	government := sha256.Sum256([]byte("government"))
	payload := sha256.Sum256([]byte("device-payload"))
	return EvidenceBinding{CaptureEvidenceDigest: capture[:], GovernmentEvidenceDigest: government[:], GovernmentVerificationTime: now.Add(-time.Minute), DeviceAttestation: veidtypes.DeviceAttestationRecord{
		AttestationID: "attestation-1", Platform: veidtypes.DevicePlatformAndroid, Provider: veidtypes.DeviceAttestationProviderPlayIntegrity, Nonce: "capture-nonce-123", AttestedAt: now.Add(-2 * time.Minute), IntegrityLevel: veidtypes.DeviceIntegrityHardwareBacked, DeviceModel: "device", OSVersion: "1", AppVersion: "1", AppID: "app", HardwareBacked: true, Supported: true, Verified: true, PayloadHash: payload[:], VaultRef: "vault://capture/1",
	}}
}

func validReceipt(now time.Time, evidence EvidenceBinding) (veidtypes.InferenceReceipt, ed25519.PrivateKey) {
	pub, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	fingerprint := sha256.Sum256(pub)
	digest := sha256.Sum256([]byte("digest"))
	return veidtypes.InferenceReceipt{Domain: veidtypes.InferenceReceiptDomain, Version: veidtypes.InferenceReceiptVersion, ChainID: "veid-test-1", AccountAddress: "virtengine1account", RequestID: "request-1", ScopeIDs: []string{"scope-1"}, Nonce: "receipt-nonce", InputDigest: digest[:], FeatureDigest: digest[:], SchemaDigest: digest[:], EvidenceLineageDigest: EvidenceLineageDigest(evidence), PipelineVersion: "pipeline-1", ModelManifestDigest: digest[:], ModelDigest: digest[:], RuntimeImageDigest: digest[:], RuntimeDigest: digest[:], ConfigDigest: veidtypes.CanonicalInferenceDeterminismConfigDigest(), DeterminismProfile: veidtypes.CanonicalInferenceDeterminismProfile(), Score: 90, Status: veidtypes.VerificationResultStatusSuccess, ConfidenceMillionths: 900000, ReasonCodes: []veidtypes.ReasonCode{veidtypes.ReasonCodeSuccess}, IssuedHeight: 10, IssuedAt: now, ExpiresHeight: 11, ExpiresAt: now.Add(time.Minute), SignerKeyID: "signer-1", SignerFingerprint: fmt.Sprintf("%x", fingerprint), SignerSequence: 1}, privateKey
}
