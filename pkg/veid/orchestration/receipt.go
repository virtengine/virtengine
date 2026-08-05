// Package orchestration provides the off-chain trust boundary that prepares
// signed VEID inference receipts. It deliberately does not perform inference
// or mutate chain state: callers submit the resulting receipt through the
// existing VEID transaction flow, where signature, freshness, and replay
// checks remain authoritative.
package orchestration

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

const receiptEvidenceBindingDomain = "VEID_RECEIPT_EVIDENCE_BINDING_V1"

// EvidenceBinding holds privacy-preserving hashes of the capture and
// government-verification artifacts used for a receipt. Raw identity data and
// device attestation payloads must remain in the vault and never enter this
// structure or chain state.
type EvidenceBinding struct {
	CaptureEvidenceDigest      []byte
	GovernmentEvidenceDigest   []byte
	DeviceAttestation          veidtypes.DeviceAttestationRecord
	GovernmentVerificationTime time.Time
}

// ReceiptRequest is the fully bound, already signed inference receipt received
// from the inference service. The service must sign only after evaluating the
// exact evidence binding supplied here.
type ReceiptRequest struct {
	Receipt  veidtypes.InferenceReceipt
	Evidence EvidenceBinding
}

// ReceiptVerifier connects this off-chain boundary to the existing VEID
// receipt verification API. Implementations normally resolve the authorized
// signer key from chain state and verify the Ed25519 signature; the interface
// prevents orchestration code from silently accepting an unsigned receipt.
type ReceiptVerifier interface {
	VerifyInferenceReceipt(context.Context, veidtypes.InferenceReceipt) error
}

// VerifyAndPrepare verifies the signer using the supplied chain-aware verifier
// and then validates the capture/government evidence binding. The returned
// receipt remains immutable and is ready to submit through the normal VEID
// receipt transaction path.
func VerifyAndPrepare(ctx context.Context, verifier ReceiptVerifier, request ReceiptRequest, now time.Time) (veidtypes.InferenceReceipt, error) {
	if verifier == nil {
		return veidtypes.InferenceReceipt{}, fmt.Errorf("inference receipt verifier is required")
	}
	prepared := cloneInferenceReceipt(request.Receipt)
	request.Receipt = prepared
	if err := verifier.VerifyInferenceReceipt(ctx, prepared); err != nil {
		return veidtypes.InferenceReceipt{}, fmt.Errorf("inference receipt signer verification failed: %w", err)
	}
	if err := ValidateReceiptRequest(request, now); err != nil {
		return veidtypes.InferenceReceipt{}, err
	}
	return prepared, nil
}

func cloneInferenceReceipt(receipt veidtypes.InferenceReceipt) veidtypes.InferenceReceipt {
	cloned := receipt
	cloned.ScopeIDs = append([]string(nil), receipt.ScopeIDs...)
	cloned.InputDigest = append([]byte(nil), receipt.InputDigest...)
	cloned.FeatureDigest = append([]byte(nil), receipt.FeatureDigest...)
	cloned.SchemaDigest = append([]byte(nil), receipt.SchemaDigest...)
	cloned.EvidenceLineageDigest = append([]byte(nil), receipt.EvidenceLineageDigest...)
	cloned.ModelManifestDigest = append([]byte(nil), receipt.ModelManifestDigest...)
	cloned.ModelDigest = append([]byte(nil), receipt.ModelDigest...)
	cloned.RuntimeImageDigest = append([]byte(nil), receipt.RuntimeImageDigest...)
	cloned.RuntimeDigest = append([]byte(nil), receipt.RuntimeDigest...)
	cloned.ConfigDigest = append([]byte(nil), receipt.ConfigDigest...)
	cloned.ReasonCodes = append([]veidtypes.ReasonCode(nil), receipt.ReasonCodes...)
	cloned.Signature = append([]byte(nil), receipt.Signature...)
	return cloned
}

// ValidateReceiptRequest validates an off-chain receipt submission and ensures
// the receipt's signed evidence lineage digest is bound to capture evidence,
// device attestation, and the government-verification result. It does not make
// an authorization decision and it does not run model inference.
func ValidateReceiptRequest(request ReceiptRequest, now time.Time) error {
	if err := request.Receipt.Validate(); err != nil {
		return fmt.Errorf("invalid signed inference receipt: %w", err)
	}
	if now.IsZero() || !request.Receipt.ExpiresAt.After(now.UTC()) {
		return fmt.Errorf("inference receipt is expired at preparation time")
	}
	if request.Receipt.IssuedAt.After(now.UTC()) {
		return fmt.Errorf("inference receipt issue time is in the future")
	}
	if len(request.Evidence.CaptureEvidenceDigest) != sha256.Size || len(request.Evidence.GovernmentEvidenceDigest) != sha256.Size {
		return fmt.Errorf("capture and government evidence digests must be SHA-256 digests")
	}
	if err := request.Evidence.DeviceAttestation.Validate(); err != nil {
		return fmt.Errorf("invalid device attestation: %w", err)
	}
	if !request.Evidence.DeviceAttestation.Supported || !request.Evidence.DeviceAttestation.Verified {
		return fmt.Errorf("device attestation must be supported and verified")
	}
	if request.Evidence.GovernmentVerificationTime.IsZero() {
		return fmt.Errorf("government verification time is required")
	}
	if request.Evidence.GovernmentVerificationTime.After(now.UTC()) {
		return fmt.Errorf("government verification time is in the future")
	}
	if request.Receipt.IssuedAt.UTC().Before(request.Evidence.GovernmentVerificationTime.UTC()) {
		return fmt.Errorf("receipt predates government verification")
	}
	if !bytes.Equal(request.Receipt.EvidenceLineageDigest, EvidenceLineageDigest(request.Evidence)) {
		return fmt.Errorf("receipt evidence lineage digest does not match bound evidence")
	}
	return nil
}

// EvidenceLineageDigest produces the domain-separated digest that an inference
// service must place in InferenceReceipt.EvidenceLineageDigest. The included
// attestation nonce makes the capture session non-transferable; the receipt's
// own nonce, signer, expiry, and model/runtime/config commitments are covered
// by its existing signature.
func EvidenceLineageDigest(evidence EvidenceBinding) []byte {
	h := sha256.New()
	writeField(h, []byte(receiptEvidenceBindingDomain))
	writeField(h, evidence.CaptureEvidenceDigest)
	writeField(h, evidence.GovernmentEvidenceDigest)
	writeField(h, []byte(evidence.DeviceAttestation.AttestationID))
	writeField(h, []byte(evidence.DeviceAttestation.Nonce))
	writeField(h, []byte(evidence.DeviceAttestation.Provider))
	writeField(h, []byte(evidence.DeviceAttestation.Platform))
	writeField(h, []byte(evidence.DeviceAttestation.IntegrityLevel))
	writeField(h, evidence.DeviceAttestation.PayloadHash)
	writeField(h, []byte(strings.ToLower(evidence.DeviceAttestation.VaultRef)))
	var timestamp [8]byte
	binary.BigEndian.PutUint64(timestamp[:], uint64(evidence.GovernmentVerificationTime.UTC().Unix()))
	writeField(h, timestamp[:])
	return h.Sum(nil)
}

func writeField(h interface{ Write([]byte) (int, error) }, value []byte) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(value)))
	_, _ = h.Write(length[:])
	_, _ = h.Write(value)
}
