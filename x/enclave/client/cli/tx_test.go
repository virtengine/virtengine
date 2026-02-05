package cli

import (
	"encoding/base64"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func writeTempJSON(t *testing.T, contents string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "proposal.json")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write temp json: %v", err)
	}
	return path
}

func TestParseAddMeasurementProposalJSON(t *testing.T) {
	valid := `{
  "title": "Add SGX measurement v1",
  "description": "Allowlist SGX enclave measurement",
  "measurement_hash": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
  "tee_type": "SGX",
  "min_isv_svn": 1,
  "expiry_blocks": 0,
  "deposit": "1000uve"
}`

	path := writeTempJSON(t, valid)
	proposal, err := parseAddMeasurementProposalJSON(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if proposal.Title == "" || proposal.Description == "" || proposal.MeasurementHash == "" || proposal.TEEType == "" || proposal.Deposit == "" {
		t.Fatalf("unexpected empty fields in proposal: %+v", proposal)
	}

	missing := `{"title":"","description":"","measurement_hash":"","tee_type":"","deposit":""}`
	path = writeTempJSON(t, missing)
	if _, err := parseAddMeasurementProposalJSON(path); err == nil {
		t.Fatalf("expected error for missing fields")
	}
}

func TestParseRevokeMeasurementProposalJSON(t *testing.T) {
	valid := `{
  "title": "Revoke SGX measurement v1",
  "description": "Revoke compromised measurement",
  "measurement_hash": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
  "reason": "CVE-2026-0001",
  "deposit": "1000uve"
}`

	path := writeTempJSON(t, valid)
	proposal, err := parseRevokeMeasurementProposalJSON(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if proposal.Title == "" || proposal.Description == "" || proposal.MeasurementHash == "" || proposal.Reason == "" || proposal.Deposit == "" {
		t.Fatalf("unexpected empty fields in proposal: %+v", proposal)
	}

	missing := `{"title":"","description":"","measurement_hash":"","reason":"","deposit":""}`
	path = writeTempJSON(t, missing)
	if _, err := parseRevokeMeasurementProposalJSON(path); err == nil {
		t.Fatalf("expected error for missing fields")
	}
}

func TestDecodeMeasurementHash(t *testing.T) {
	hexHash := hex.EncodeToString(make([]byte, 32))
	if _, err := decodeMeasurementHash(hexHash); err != nil {
		t.Fatalf("expected hex hash to decode: %v", err)
	}

	b64 := base64.StdEncoding.EncodeToString(make([]byte, 32))
	if _, err := decodeMeasurementHash(b64); err != nil {
		t.Fatalf("expected base64 hash to decode: %v", err)
	}

	if _, err := decodeMeasurementHash(""); err == nil {
		t.Fatalf("expected error for empty hash")
	}

	if _, err := decodeMeasurementHash("abcd"); err == nil {
		t.Fatalf("expected error for invalid hash")
	}
}

func TestParseTEEType(t *testing.T) {
	if _, err := parseTEEType("sgx"); err != nil {
		t.Fatalf("expected SGX to parse: %v", err)
	}

	if _, err := parseTEEType("TEE_TYPE_SGX"); err != nil {
		t.Fatalf("expected TEE_TYPE_SGX to parse: %v", err)
	}

	if _, err := parseTEEType("sev_snp"); err != nil {
		t.Fatalf("expected sev_snp to parse: %v", err)
	}

	if _, err := parseTEEType(""); err == nil {
		t.Fatalf("expected error for empty tee type")
	}
}
