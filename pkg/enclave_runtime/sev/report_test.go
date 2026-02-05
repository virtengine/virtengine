// Package sev provides tests for the AMD SEV-SNP report handling.
package sev

import (
	"bytes"
	"testing"
)

func TestParseReportRoundtrip(t *testing.T) {
	// Create a report
	original := &AttestationReport{
		Version:          ReportVersionV2,
		GuestSVN:         5,
		Policy:           GuestPolicy{ABIMajor: 1, ABIMinor: 0, SMT: true},
		VMPL:             0,
		SignatureAlgo:    SigAlgoECDSAP384SHA384,
		PlatformInfo:     0x01,
		AuthorKeyEnabled: 0,
		CurrentTCB:       TCBVersion{BootLoader: 3, TEE: 0, SNP: 14, Microcode: 209},
		ReportedTCB:      TCBVersion{BootLoader: 3, TEE: 0, SNP: 14, Microcode: 209},
	}

	// Set some data
	copy(original.FamilyID[:], []byte("test-family-id!"))
	copy(original.ImageID[:], []byte("test-image-id!!"))
	copy(original.ReportData[:], []byte("user-data-in-attestation-report-test"))
	copy(original.LaunchDigest[:], []byte("sha384-launch-measurement-digest!!!!!!!!"))
	copy(original.ReportID[:], []byte("report-identifier-32bytes!!"))
	copy(original.ChipID[:], []byte("chip-identifier-unique-per-cpu-64-bytes!!!!!!!!!!!!!!!!!!!!"))

	// Serialize
	serialized, err := SerializeReport(original)
	if err != nil {
		t.Fatalf("SerializeReport() error = %v", err)
	}

	if len(serialized) != ReportSize {
		t.Errorf("Serialized size = %d, want %d", len(serialized), ReportSize)
	}

	// Parse back
	parsed, err := ParseReport(serialized)
	if err != nil {
		t.Fatalf("ParseReport() error = %v", err)
	}

	// Verify fields
	if parsed.Version != original.Version {
		t.Errorf("Version = %d, want %d", parsed.Version, original.Version)
	}
	if parsed.GuestSVN != original.GuestSVN {
		t.Errorf("GuestSVN = %d, want %d", parsed.GuestSVN, original.GuestSVN)
	}
	if parsed.VMPL != original.VMPL {
		t.Errorf("VMPL = %d, want %d", parsed.VMPL, original.VMPL)
	}
	if parsed.SignatureAlgo != original.SignatureAlgo {
		t.Errorf("SignatureAlgo = %d, want %d", parsed.SignatureAlgo, original.SignatureAlgo)
	}
	if parsed.PlatformInfo != original.PlatformInfo {
		t.Errorf("PlatformInfo = %d, want %d", parsed.PlatformInfo, original.PlatformInfo)
	}

	// Check policy
	if parsed.Policy.ABIMajor != original.Policy.ABIMajor {
		t.Errorf("Policy.ABIMajor = %d, want %d", parsed.Policy.ABIMajor, original.Policy.ABIMajor)
	}
	if parsed.Policy.SMT != original.Policy.SMT {
		t.Errorf("Policy.SMT = %v, want %v", parsed.Policy.SMT, original.Policy.SMT)
	}

	// Check TCB
	if parsed.CurrentTCB.BootLoader != original.CurrentTCB.BootLoader {
		t.Errorf("CurrentTCB.BootLoader = %d, want %d",
			parsed.CurrentTCB.BootLoader, original.CurrentTCB.BootLoader)
	}

	// Check byte arrays
	if !bytes.Equal(parsed.FamilyID[:], original.FamilyID[:]) {
		t.Error("FamilyID mismatch")
	}
	if !bytes.Equal(parsed.ImageID[:], original.ImageID[:]) {
		t.Error("ImageID mismatch")
	}
	if !bytes.Equal(parsed.ReportData[:], original.ReportData[:]) {
		t.Error("ReportData mismatch")
	}
	if !bytes.Equal(parsed.LaunchDigest[:], original.LaunchDigest[:]) {
		t.Error("LaunchDigest mismatch")
	}
	if !bytes.Equal(parsed.ChipID[:], original.ChipID[:]) {
		t.Error("ChipID mismatch")
	}
}

func TestParseReportTooShort(t *testing.T) {
	data := make([]byte, 100) // Too short
	_, err := ParseReport(data)
	if err == nil {
		t.Error("ParseReport() expected error for short data")
	}
}

func TestValidateReport(t *testing.T) {
	tests := []struct {
		name    string
		report  *AttestationReport
		wantErr bool
	}{
		{
			name: "valid report",
			report: &AttestationReport{
				Version:       ReportVersionV2,
				SignatureAlgo: SigAlgoECDSAP384SHA384,
				VMPL:          0,
				Policy:        GuestPolicy{ABIMajor: 1, ABIMinor: 0},
				ChipID:        [64]byte{1, 2, 3, 4, 5}, // Non-zero
				LaunchDigest:  [48]byte{1, 2, 3},       // Non-zero
			},
			wantErr: false,
		},
		{
			name:    "nil report",
			report:  nil,
			wantErr: true,
		},
		{
			name: "old version",
			report: &AttestationReport{
				Version:       1, // Old version
				SignatureAlgo: SigAlgoECDSAP384SHA384,
				Policy:        GuestPolicy{ABIMajor: 1},
				ChipID:        [64]byte{1},
				LaunchDigest:  [48]byte{1},
			},
			wantErr: true,
		},
		{
			name: "debug enabled",
			report: &AttestationReport{
				Version:       ReportVersionV2,
				SignatureAlgo: SigAlgoECDSAP384SHA384,
				Policy:        GuestPolicy{ABIMajor: 1, Debug: true},
				ChipID:        [64]byte{1},
				LaunchDigest:  [48]byte{1},
			},
			wantErr: true,
		},
		{
			name: "invalid VMPL",
			report: &AttestationReport{
				Version:       ReportVersionV2,
				SignatureAlgo: SigAlgoECDSAP384SHA384,
				VMPL:          10, // Invalid
				Policy:        GuestPolicy{ABIMajor: 1},
				ChipID:        [64]byte{1},
				LaunchDigest:  [48]byte{1},
			},
			wantErr: true,
		},
		{
			name: "zero chip ID",
			report: &AttestationReport{
				Version:       ReportVersionV2,
				SignatureAlgo: SigAlgoECDSAP384SHA384,
				Policy:        GuestPolicy{ABIMajor: 1},
				ChipID:        [64]byte{}, // All zeros
				LaunchDigest:  [48]byte{1},
			},
			wantErr: true,
		},
		{
			name: "zero launch digest",
			report: &AttestationReport{
				Version:       ReportVersionV2,
				SignatureAlgo: SigAlgoECDSAP384SHA384,
				Policy:        GuestPolicy{ABIMajor: 1},
				ChipID:        [64]byte{1},
				LaunchDigest:  [48]byte{}, // All zeros
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateReport(tt.report)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateReport() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReportRequestToBytes(t *testing.T) {
	req := &ReportRequest{
		VMPL: 2,
	}
	copy(req.UserData[:], []byte("test-user-data"))

	data := req.ToBytes()

	if len(data) != 96 {
		t.Errorf("ToBytes() length = %d, want 96", len(data))
	}

	// Verify user data is at start
	if !bytes.Equal(data[:14], []byte("test-user-data")) {
		t.Error("UserData not at expected position")
	}

	// Verify VMPL is at offset 64
	vmpl := uint32(data[64]) | uint32(data[65])<<8 | uint32(data[66])<<16 | uint32(data[67])<<24
	if vmpl != 2 {
		t.Errorf("VMPL = %d, want 2", vmpl)
	}
}

func TestAttestationReportGetSignatureComponents(t *testing.T) {
	report := &AttestationReport{}

	// Set signature bytes
	for i := 0; i < 48; i++ {
		report.Signature[i] = byte(i)         // R
		report.Signature[48+i] = byte(i + 48) // S
	}

	r := report.GetSignatureR()
	s := report.GetSignatureS()

	if len(r) != 48 {
		t.Errorf("GetSignatureR() length = %d, want 48", len(r))
	}
	if len(s) != 48 {
		t.Errorf("GetSignatureS() length = %d, want 48", len(s))
	}

	// Verify R values
	for i := 0; i < 48; i++ {
		if r[i] != byte(i) {
			t.Errorf("R[%d] = %d, want %d", i, r[i], i)
			break
		}
	}

	// Verify S values
	for i := 0; i < 48; i++ {
		if s[i] != byte(i+48) {
			t.Errorf("S[%d] = %d, want %d", i, s[i], i+48)
			break
		}
	}
}

func TestAttestationReportPlatformFlags(t *testing.T) {
	report := &AttestationReport{}

	// No flags
	report.PlatformInfo = 0x00
	if report.SMTEnabled() {
		t.Error("SMTEnabled() should be false")
	}
	if report.TSMEEnabled() {
		t.Error("TSMEEnabled() should be false")
	}

	// SMT only
	report.PlatformInfo = 0x01
	if !report.SMTEnabled() {
		t.Error("SMTEnabled() should be true")
	}
	if report.TSMEEnabled() {
		t.Error("TSMEEnabled() should be false")
	}

	// Both
	report.PlatformInfo = 0x03
	if !report.SMTEnabled() {
		t.Error("SMTEnabled() should be true")
	}
	if !report.TSMEEnabled() {
		t.Error("TSMEEnabled() should be true")
	}
}

func TestAttestationReportTCBMeetsMinimum(t *testing.T) {
	report := &AttestationReport{
		ReportedTCB: TCBVersion{
			BootLoader: 3,
			TEE:        0,
			SNP:        14,
			Microcode:  209,
		},
	}

	minTCB := TCBVersion{BootLoader: 2, TEE: 0, SNP: 10, Microcode: 200}
	if !report.TCBMeetsMinimum(minTCB) {
		t.Error("TCBMeetsMinimum() should be true")
	}

	highMinTCB := TCBVersion{BootLoader: 4, TEE: 0, SNP: 20, Microcode: 220}
	if report.TCBMeetsMinimum(highMinTCB) {
		t.Error("TCBMeetsMinimum() should be false for high min")
	}
}

func TestAttestationReportGetTCBComponents(t *testing.T) {
	report := &AttestationReport{
		ReportedTCB: TCBVersion{
			BootLoader: 3,
			TEE:        1,
			SNP:        14,
			Microcode:  209,
		},
	}

	bl, tee, snp, ucode := report.GetTCBComponents()

	if bl != 3 {
		t.Errorf("BootLoader = %d, want 3", bl)
	}
	if tee != 1 {
		t.Errorf("TEE = %d, want 1", tee)
	}
	if snp != 14 {
		t.Errorf("SNP = %d, want 14", snp)
	}
	if ucode != 209 {
		t.Errorf("Microcode = %d, want 209", ucode)
	}
}
