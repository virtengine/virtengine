package keeper

import (
	"bytes"
	"crypto/sha256"
	"testing"

	enclave "github.com/virtengine/virtengine/pkg/enclave_runtime"
	"github.com/virtengine/virtengine/x/enclave/types"
)

func TestParseAttestationChainPEM(t *testing.T) {
	chain, err := parseAttestationChain([][]byte{[]byte(enclave.IntelSGXRootCAPEM)})
	if err != nil {
		t.Fatalf("parseAttestationChain failed: %v", err)
	}
	if len(chain.certs) == 0 {
		t.Fatalf("expected parsed certs, got none")
	}
}

func TestParseAttestationChainJSON(t *testing.T) {
	payload := []byte(`{"tcbInfo":{"id":"SGX","version":3,"tcbType":0,"tcbEvaluationDataNumber":1,"tcbLevels":[]}}`)
	chain, err := parseAttestationChain([][]byte{payload})
	if err != nil {
		t.Fatalf("parseAttestationChain failed: %v", err)
	}
	if len(chain.json) == 0 {
		t.Fatalf("expected json collateral entries")
	}
}

func TestVerifySEVSNPMeasurement(t *testing.T) {
	measurement := bytes.Repeat([]byte{0xAB}, 48)
	sum := sha256.Sum256(measurement)
	identity := &types.EnclaveIdentity{MeasurementHash: sum[:]}

	if err := verifySEVSNPMeasurement(identity, measurement); err != nil {
		t.Fatalf("verifySEVSNPMeasurement failed: %v", err)
	}
}

func TestVerifyNitroMeasurement(t *testing.T) {
	pcr := bytes.Repeat([]byte{0xCD}, 48)
	sum := sha256.Sum256(pcr)
	identity := &types.EnclaveIdentity{MeasurementHash: sum[:]}

	if err := verifyNitroMeasurement(identity, pcr); err != nil {
		t.Fatalf("verifyNitroMeasurement failed: %v", err)
	}
}
