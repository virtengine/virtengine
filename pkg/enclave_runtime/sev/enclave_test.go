// Package sev provides tests for the AMD SEV-SNP integration package.
package sev

import (
	"bytes"
	"testing"
)

func TestGuestPolicyToUint64(t *testing.T) {
	tests := []struct {
		name   string
		policy GuestPolicy
		want   uint64
	}{
		{
			name: "default production policy",
			policy: GuestPolicy{
				ABIMajor: 1,
				ABIMinor: 0,
				SMT:      true,
				Debug:    false,
			},
			want: 0x0000000000010100, // SMT=1, major=1, minor=0
		},
		{
			name: "debug enabled",
			policy: GuestPolicy{
				ABIMajor: 1,
				ABIMinor: 0,
				SMT:      true,
				Debug:    true,
			},
			want: 0x0000000000090100, // Debug + SMT
		},
		{
			name: "single socket",
			policy: GuestPolicy{
				ABIMajor:     1,
				ABIMinor:     0,
				SingleSocket: true,
			},
			want: 0x0000000000100100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.policy.ToUint64()
			if got != tt.want {
				t.Errorf("ToUint64() = 0x%016x, want 0x%016x", got, tt.want)
			}
		})
	}
}

func TestParseGuestPolicy(t *testing.T) {
	tests := []struct {
		name string
		raw  uint64
		want GuestPolicy
	}{
		{
			name: "production policy",
			raw:  0x0000000000010100,
			want: GuestPolicy{
				ABIMajor: 1,
				ABIMinor: 0,
				SMT:      true,
			},
		},
		{
			name: "debug policy",
			raw:  0x0000000000090100,
			want: GuestPolicy{
				ABIMajor: 1,
				ABIMinor: 0,
				SMT:      true,
				Debug:    true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseGuestPolicy(tt.raw)
			if got.ABIMajor != tt.want.ABIMajor || got.ABIMinor != tt.want.ABIMinor ||
				got.SMT != tt.want.SMT || got.Debug != tt.want.Debug {
				t.Errorf("ParseGuestPolicy() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestGuestPolicyRoundtrip(t *testing.T) {
	policies := []GuestPolicy{
		{ABIMajor: 1, ABIMinor: 0, SMT: true},
		{ABIMajor: 1, ABIMinor: 5, Debug: true},
		{ABIMajor: 2, ABIMinor: 0, SingleSocket: true, Migration: true},
	}

	for _, p := range policies {
		raw := p.ToUint64()
		parsed := ParseGuestPolicy(raw)

		if parsed.ABIMajor != p.ABIMajor || parsed.ABIMinor != p.ABIMinor ||
			parsed.SMT != p.SMT || parsed.Debug != p.Debug ||
			parsed.SingleSocket != p.SingleSocket || parsed.Migration != p.Migration {
			t.Errorf("Roundtrip failed: original=%+v parsed=%+v", p, parsed)
		}
	}
}

func TestTCBVersionToUint64(t *testing.T) {
	tcb := TCBVersion{
		BootLoader: 3,
		TEE:        0,
		SNP:        14,
		Microcode:  209,
	}

	raw := tcb.ToUint64()
	parsed := ParseTCBVersion(raw)

	if parsed.BootLoader != tcb.BootLoader ||
		parsed.TEE != tcb.TEE ||
		parsed.SNP != tcb.SNP ||
		parsed.Microcode != tcb.Microcode {
		t.Errorf("TCBVersion roundtrip failed: original=%+v parsed=%+v", tcb, parsed)
	}
}

func TestTCBVersionCompare(t *testing.T) {
	tests := []struct {
		name string
		a    TCBVersion
		b    TCBVersion
		want int
	}{
		{
			name: "equal",
			a:    TCBVersion{BootLoader: 3, TEE: 0, SNP: 14, Microcode: 209},
			b:    TCBVersion{BootLoader: 3, TEE: 0, SNP: 14, Microcode: 209},
			want: 0,
		},
		{
			name: "a less than b - bootloader",
			a:    TCBVersion{BootLoader: 2, TEE: 0, SNP: 14, Microcode: 209},
			b:    TCBVersion{BootLoader: 3, TEE: 0, SNP: 14, Microcode: 209},
			want: -1,
		},
		{
			name: "a greater than b - microcode",
			a:    TCBVersion{BootLoader: 3, TEE: 0, SNP: 14, Microcode: 210},
			b:    TCBVersion{BootLoader: 3, TEE: 0, SNP: 14, Microcode: 209},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.a.Compare(tt.b)
			if got != tt.want {
				t.Errorf("Compare() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestTCBVersionMeetsMinimum(t *testing.T) {
	current := TCBVersion{BootLoader: 3, TEE: 0, SNP: 14, Microcode: 209}

	tests := []struct {
		name string
		min  TCBVersion
		want bool
	}{
		{
			name: "meets exact",
			min:  TCBVersion{BootLoader: 3, TEE: 0, SNP: 14, Microcode: 209},
			want: true,
		},
		{
			name: "meets higher",
			min:  TCBVersion{BootLoader: 2, TEE: 0, SNP: 10, Microcode: 200},
			want: true,
		},
		{
			name: "fails - bootloader too high",
			min:  TCBVersion{BootLoader: 4, TEE: 0, SNP: 14, Microcode: 209},
			want: false,
		},
		{
			name: "fails - microcode too high",
			min:  TCBVersion{BootLoader: 3, TEE: 0, SNP: 14, Microcode: 210},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := current.MeetsMinimum(tt.min)
			if got != tt.want {
				t.Errorf("MeetsMinimum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSEVGuestInitialize(t *testing.T) {
	// Should initialize in simulation mode since hardware not available
	guest := NewSEVGuest()
	if err := guest.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}
	defer guest.Close()

	if !guest.IsInitialized() {
		t.Error("IsInitialized() = false after Initialize()")
	}

	if !guest.IsSimulated() {
		t.Log("Note: Running in hardware mode")
	}
}

func TestSEVGuestGetPlatformInfo(t *testing.T) {
	guest := NewSEVGuest()
	if err := guest.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}
	defer guest.Close()

	info, err := guest.GetPlatformInfo()
	if err != nil {
		t.Fatalf("GetPlatformInfo() error = %v", err)
	}

	if info == nil {
		t.Fatal("GetPlatformInfo() returned nil")
	}

	if info.APIVersion == 0 {
		t.Error("APIVersion is 0")
	}

	// Check chip ID is non-zero
	var zeroChipID [ChipIDSize]byte
	if bytes.Equal(info.ChipID[:], zeroChipID[:]) {
		t.Error("ChipID is all zeros")
	}
}

func TestSEVGuestGenerateAttestation(t *testing.T) {
	guest := NewSEVGuest()
	if err := guest.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}
	defer guest.Close()

	var userData [ReportDataSize]byte
	copy(userData[:], []byte("test-nonce-12345"))

	attestation, err := guest.GenerateAttestation(userData)
	if err != nil {
		t.Fatalf("GenerateAttestation() error = %v", err)
	}

	if len(attestation) != ReportSize {
		t.Errorf("Attestation size = %d, want %d", len(attestation), ReportSize)
	}

	// Parse and verify the attestation
	report, err := ParseReport(attestation)
	if err != nil {
		t.Fatalf("ParseReport() error = %v", err)
	}

	if report.Version < ReportVersionV2 {
		t.Errorf("Report version = %d, want >= %d", report.Version, ReportVersionV2)
	}

	// Verify userData is in report
	if !bytes.Equal(report.ReportData[:16], []byte("test-nonce-12345")) {
		t.Error("ReportData doesn't contain expected userData")
	}
}

func TestSEVGuestDeriveKey(t *testing.T) {
	guest := NewSEVGuest()
	if err := guest.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}
	defer guest.Close()

	req := &KeyRequest{
		RootKeySelect:    KeyRootVCEK,
		GuestFieldSelect: KeyFieldGuest | KeyFieldTCB,
		VMPL:             VMPL0,
	}

	key, err := guest.DeriveKey(req)
	if err != nil {
		t.Fatalf("DeriveKey() error = %v", err)
	}

	if len(key) != 32 {
		t.Errorf("Key length = %d, want 32", len(key))
	}

	// Derive again with same params - should get same key
	key2, err := guest.DeriveKey(req)
	if err != nil {
		t.Fatalf("DeriveKey() second call error = %v", err)
	}

	if !bytes.Equal(key, key2) {
		t.Error("DeriveKey() not deterministic")
	}

	// Different params should give different key
	req.GuestFieldSelect = KeyFieldGuest
	key3, err := guest.DeriveKey(req)
	if err != nil {
		t.Fatalf("DeriveKey() with different params error = %v", err)
	}

	if bytes.Equal(key, key3) {
		t.Error("DeriveKey() should give different key with different params")
	}
}

func TestKeyRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		req     KeyRequest
		wantErr bool
	}{
		{
			name: "valid VCEK",
			req: KeyRequest{
				RootKeySelect: KeyRootVCEK,
				VMPL:          VMPL0,
			},
			wantErr: false,
		},
		{
			name: "valid VMRK",
			req: KeyRequest{
				RootKeySelect: KeyRootVMRK,
				VMPL:          VMPL3,
			},
			wantErr: false,
		},
		{
			name: "invalid root key",
			req: KeyRequest{
				RootKeySelect: 99,
				VMPL:          VMPL0,
			},
			wantErr: true,
		},
		{
			name: "invalid VMPL",
			req: KeyRequest{
				RootKeySelect: KeyRootVCEK,
				VMPL:          10,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSEVGuestVerifyPolicySecure(t *testing.T) {
	guest := NewSEVGuest()
	if err := guest.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}
	defer guest.Close()

	// Default policy should be secure
	err := guest.VerifyPolicySecure()
	if err != nil {
		t.Errorf("VerifyPolicySecure() error = %v", err)
	}
}
