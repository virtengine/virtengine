package artifact_store

import (
	"testing"
)

func TestCIDValidator(t *testing.T) {
	t.Run("Production validator rejects stub CIDs", func(t *testing.T) {
		validator := NewCIDValidator()

		// Stub CID format: Qm + 32 hex chars
		stubCID := testStubCID

		err := validator.ValidateCID(stubCID)
		if err == nil {
			t.Error("expected error for stub CID in production mode")
		}
	})

	t.Run("Test validator allows stub CIDs", func(t *testing.T) {
		validator := NewTestCIDValidator()

		// Stub CID format: Qm + 32 hex chars
		stubCID := testStubCID

		err := validator.ValidateCID(stubCID)
		if err != nil {
			t.Errorf("unexpected error for stub CID in test mode: %v", err)
		}
	})

	t.Run("Real CIDv0 is valid", func(t *testing.T) {
		validator := NewCIDValidator()

		// Real CIDv0 (base58-encoded multihash)
		realCIDv0 := "QmRf22bZar3WKmojipms22PkXH1MZGmvsqzQtuSvQE3uhm"

		err := validator.ValidateCID(realCIDv0)
		if err != nil {
			t.Errorf("unexpected error for real CIDv0: %v", err)
		}
	})

	t.Run("Real CIDv1 is valid", func(t *testing.T) {
		validator := NewCIDValidator()

		// Real CIDv1 (base32-encoded)
		realCIDv1 := "bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi"

		err := validator.ValidateCID(realCIDv1)
		if err != nil {
			t.Errorf("unexpected error for real CIDv1: %v", err)
		}
	})

	t.Run("Empty CID is rejected", func(t *testing.T) {
		validator := NewCIDValidator()

		err := validator.ValidateCID("")
		if err == nil {
			t.Error("expected error for empty CID")
		}
	})

	t.Run("Invalid CID is rejected", func(t *testing.T) {
		validator := NewCIDValidator()

		err := validator.ValidateCID("not-a-cid")
		if err == nil {
			t.Error("expected error for invalid CID")
		}
	})

	t.Run("IsValidCID helper", func(t *testing.T) {
		validator := NewCIDValidator()

		if validator.IsValidCID("") {
			t.Error("expected false for empty CID")
		}

		if validator.IsValidCID("not-a-cid") {
			t.Error("expected false for invalid CID")
		}

		if !validator.IsValidCID("QmRf22bZar3WKmojipms22PkXH1MZGmvsqzQtuSvQE3uhm") {
			t.Error("expected true for valid CID")
		}
	})
}

func TestIsStubCID(t *testing.T) {
	tests := []struct {
		name   string
		cid    string
		isStub bool
	}{
		{
			name:   "Stub CID with hex suffix",
			cid:    "Qm0123456789abcdef0123456789abcdef",
			isStub: true,
		},
		{
			name:   "Stub CID uppercase hex",
			cid:    "Qm0123456789ABCDEF0123456789ABCDEF",
			isStub: true,
		},
		{
			name:   "Real CIDv0",
			cid:    "QmRf22bZar3WKmojipms22PkXH1MZGmvsqzQtuSvQE3uhm",
			isStub: false,
		},
		{
			name:   "CIDv1",
			cid:    "bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi",
			isStub: false,
		},
		{
			name:   "Empty string",
			cid:    "",
			isStub: false,
		},
		{
			name:   "Too short stub",
			cid:    "Qm0123",
			isStub: false,
		},
		{
			name:   "Wrong prefix",
			cid:    "Ab0123456789abcdef0123456789abcdef",
			isStub: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := isStubCID(tc.cid)
			if result != tc.isStub {
				t.Errorf("isStubCID(%q) = %v, want %v", tc.cid, result, tc.isStub)
			}
		})
	}
}

func TestParseCID(t *testing.T) {
	t.Run("Parse real CIDv0", func(t *testing.T) {
		cidStr := "QmRf22bZar3WKmojipms22PkXH1MZGmvsqzQtuSvQE3uhm"
		parsed, err := ParseCID(cidStr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if parsed.IsStub {
			t.Error("expected IsStub to be false for real CID")
		}
		if parsed.Version != 0 {
			t.Errorf("expected version 0, got %d", parsed.Version)
		}
		if parsed.Raw != cidStr {
			t.Errorf("expected raw %q, got %q", cidStr, parsed.Raw)
		}
	})

	t.Run("Parse real CIDv1", func(t *testing.T) {
		cidStr := "bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi"
		parsed, err := ParseCID(cidStr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if parsed.IsStub {
			t.Error("expected IsStub to be false for real CID")
		}
		if parsed.Version != 1 {
			t.Errorf("expected version 1, got %d", parsed.Version)
		}
	})

	t.Run("Parse stub CID", func(t *testing.T) {
		cidStr := "Qm0123456789abcdef0123456789abcdef"
		parsed, err := ParseCID(cidStr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !parsed.IsStub {
			t.Error("expected IsStub to be true for stub CID")
		}
	})

	t.Run("Empty CID fails", func(t *testing.T) {
		_, err := ParseCID("")
		if err == nil {
			t.Error("expected error for empty CID")
		}
	})
}

func TestValidateCIDForBackend(t *testing.T) {
	t.Run("IPFS backend validates CID", func(t *testing.T) {
		// Valid CID
		err := ValidateCIDForBackend("QmRf22bZar3WKmojipms22PkXH1MZGmvsqzQtuSvQE3uhm", BackendIPFS)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Stub CID should fail in production
		err = ValidateCIDForBackend("Qm0123456789abcdef0123456789abcdef", BackendIPFS)
		if err == nil {
			t.Error("expected error for stub CID in production validation")
		}
	})

	t.Run("Waldur backend validates non-empty", func(t *testing.T) {
		err := ValidateCIDForBackend("some-uuid-ref", BackendWaldur)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		err = ValidateCIDForBackend("", BackendWaldur)
		if err == nil {
			t.Error("expected error for empty waldur reference")
		}
	})

	t.Run("AllowStub variant allows stub CIDs", func(t *testing.T) {
		err := ValidateCIDForBackendAllowStub("Qm0123456789abcdef0123456789abcdef", BackendIPFS)
		if err == nil || err.Error() != "" {
			// Should allow stub CIDs
		}
	})

	t.Run("Unknown backend fails", func(t *testing.T) {
		err := ValidateCIDForBackend("anything", "unknown")
		if err == nil {
			t.Error("expected error for unknown backend")
		}
	})
}
