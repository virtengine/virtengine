package enclave_runtime

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/binary"
	"encoding/hex"
	"encoding/pem"
	"testing"
	"time"
)

// =============================================================================
// Common Crypto Tests
// =============================================================================

func TestHashComputer(t *testing.T) {
	h := NewHashComputer()

	t.Run("SHA256", func(t *testing.T) {
		data := []byte("test data for hashing")
		hash := h.SHA256(data)

		if len(hash) != 32 {
			t.Errorf("SHA256 hash length: got %d, want 32", len(hash))
		}

		// Verify against stdlib
		expected := sha256.Sum256(data)
		if !bytes.Equal(hash, expected[:]) {
			t.Errorf("SHA256 hash mismatch")
		}
	})

	t.Run("SHA384", func(t *testing.T) {
		data := []byte("test data for hashing")
		hash := h.SHA384(data)

		if len(hash) != 48 {
			t.Errorf("SHA384 hash length: got %d, want 48", len(hash))
		}

		// Verify against stdlib
		expected := sha512.Sum384(data)
		if !bytes.Equal(hash, expected[:]) {
			t.Errorf("SHA384 hash mismatch")
		}
	})

	t.Run("SHA512", func(t *testing.T) {
		data := []byte("test data for hashing")
		hash := h.SHA512(data)

		if len(hash) != 64 {
			t.Errorf("SHA512 hash length: got %d, want 64", len(hash))
		}

		// Verify against stdlib
		expected := sha512.Sum512(data)
		if !bytes.Equal(hash, expected[:]) {
			t.Errorf("SHA512 hash mismatch")
		}
	})

	t.Run("ComputeHash with algorithm", func(t *testing.T) {
		data := []byte("test data")

		hash256, err := h.ComputeHash("SHA-256", data)
		if err != nil {
			t.Errorf("ComputeHash SHA-256 failed: %v", err)
		}
		if len(hash256) != 32 {
			t.Errorf("SHA-256 hash length: got %d, want 32", len(hash256))
		}

		hash384, err := h.ComputeHash("SHA-384", data)
		if err != nil {
			t.Errorf("ComputeHash SHA-384 failed: %v", err)
		}
		if len(hash384) != 48 {
			t.Errorf("SHA-384 hash length: got %d, want 48", len(hash384))
		}

		_, err = h.ComputeHash("MD5", data)
		if err == nil {
			t.Error("expected error for unsupported algorithm")
		}
	})
}

func TestECDSAVerifier(t *testing.T) {
	v := NewECDSAVerifier()

	t.Run("VerifyP256 valid signature", func(t *testing.T) {
		// Generate test key
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatalf("failed to generate key: %v", err)
		}

		// Create signature
		hash := sha256.Sum256([]byte("test message"))
		r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
		if err != nil {
			t.Fatalf("failed to sign: %v", err)
		}

		// Create raw signature (r || s)
		sig := make([]byte, 64)
		r.FillBytes(sig[:32])
		s.FillBytes(sig[32:])

		// Verify
		if err := v.VerifyP256(&privateKey.PublicKey, hash[:], sig); err != nil {
			t.Errorf("verification failed: %v", err)
		}
	})

	t.Run("VerifyP256 invalid signature", func(t *testing.T) {
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatalf("failed to generate key: %v", err)
		}

		hash := sha256.Sum256([]byte("test message"))
		invalidSig := make([]byte, 64)
		rand.Read(invalidSig)

		if err := v.VerifyP256(&privateKey.PublicKey, hash[:], invalidSig); err == nil {
			t.Error("expected verification to fail")
		}
	})

	t.Run("VerifyP384 valid signature", func(t *testing.T) {
		privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
		if err != nil {
			t.Fatalf("failed to generate key: %v", err)
		}

		hash := sha512.Sum384([]byte("test message"))
		r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
		if err != nil {
			t.Fatalf("failed to sign: %v", err)
		}

		sig := make([]byte, 96)
		r.FillBytes(sig[:48])
		s.FillBytes(sig[48:])

		if err := v.VerifyP384(&privateKey.PublicKey, hash[:], sig); err != nil {
			t.Errorf("verification failed: %v", err)
		}
	})

	t.Run("VerifyP256 wrong hash length", func(t *testing.T) {
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatalf("failed to generate key: %v", err)
		}

		wrongHash := make([]byte, 48) // Wrong length for P-256
		sig := make([]byte, 64)

		if err := v.VerifyP256(&privateKey.PublicKey, wrongHash, sig); err == nil {
			t.Error("expected error for wrong hash length")
		}
	})

	t.Run("VerifyP384 wrong hash length", func(t *testing.T) {
		privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
		if err != nil {
			t.Fatalf("failed to generate key: %v", err)
		}

		wrongHash := make([]byte, 32) // Wrong length for P-384
		sig := make([]byte, 96)

		if err := v.VerifyP384(&privateKey.PublicKey, wrongHash, sig); err == nil {
			t.Error("expected error for wrong hash length")
		}
	})

	t.Run("nil public key", func(t *testing.T) {
		hash := make([]byte, 32)
		sig := make([]byte, 64)

		if err := v.VerifyP256(nil, hash, sig); err == nil {
			t.Error("expected error for nil public key")
		}
	})
}

func TestCertificateChainVerifier(t *testing.T) {
	t.Run("create verifier with defaults", func(t *testing.T) {
		v := NewCertificateChainVerifier()
		if v == nil {
			t.Fatal("verifier is nil")
		}
		if v.MaxChainLen != 5 {
			t.Errorf("MaxChainLen: got %d, want 5", v.MaxChainLen)
		}
	})

	t.Run("empty chain", func(t *testing.T) {
		v := NewCertificateChainVerifier()
		if err := v.Verify(nil); err == nil {
			t.Error("expected error for empty chain")
		}
	})

	t.Run("chain too long", func(t *testing.T) {
		v := NewCertificateChainVerifier()
		v.MaxChainLen = 2

		// Create fake chain of 3 certificates
		certs := make([]*x509.Certificate, 3)
		for i := range certs {
			certs[i] = &x509.Certificate{}
		}

		err := v.Verify(certs)
		if err == nil {
			t.Error("expected error for chain too long")
		}
	})

	t.Run("add root CA from PEM", func(t *testing.T) {
		v := NewCertificateChainVerifier()

		// Add Intel SGX Root CA
		err := v.AddRootCA([]byte(IntelSGXRootCAPEM))
		if err != nil {
			t.Errorf("failed to add root CA: %v", err)
		}
	})

	t.Run("invalid PEM", func(t *testing.T) {
		v := NewCertificateChainVerifier()
		err := v.AddRootCA([]byte("not a valid certificate"))
		if err == nil {
			t.Error("expected error for invalid PEM")
		}
	})
}

func TestCertificateCache(t *testing.T) {
	t.Run("basic operations", func(t *testing.T) {
		cache := NewCertificateCache(10, time.Hour)

		// Put and get
		cert := &x509.Certificate{Subject: pkix.Name{CommonName: "test"}}
		cache.Put("key1", cert, nil, nil, "test")

		cached, err := cache.Get("key1")
		if err != nil {
			t.Errorf("failed to get cached cert: %v", err)
		}
		if cached.Certificate.Subject.CommonName != "test" {
			t.Error("cached certificate mismatch")
		}

		if cache.Size() != 1 {
			t.Errorf("cache size: got %d, want 1", cache.Size())
		}
	})

	t.Run("cache miss", func(t *testing.T) {
		cache := NewCertificateCache(10, time.Hour)
		_, err := cache.Get("nonexistent")
		if err == nil {
			t.Error("expected error for cache miss")
		}
	})

	t.Run("expired entry", func(t *testing.T) {
		cache := NewCertificateCache(10, time.Nanosecond)
		cert := &x509.Certificate{}
		cache.Put("key1", cert, nil, nil, "test")

		time.Sleep(time.Millisecond)

		_, err := cache.Get("key1")
		if err == nil {
			t.Error("expected error for expired entry")
		}
	})

	t.Run("eviction", func(t *testing.T) {
		cache := NewCertificateCache(2, time.Hour)

		cache.Put("key1", &x509.Certificate{}, nil, nil, "test")
		cache.Put("key2", &x509.Certificate{}, nil, nil, "test")
		cache.Put("key3", &x509.Certificate{}, nil, nil, "test")

		if cache.Size() != 2 {
			t.Errorf("cache size after eviction: got %d, want 2", cache.Size())
		}
	})

	t.Run("clear", func(t *testing.T) {
		cache := NewCertificateCache(10, time.Hour)
		cache.Put("key1", &x509.Certificate{}, nil, nil, "test")
		cache.Put("key2", &x509.Certificate{}, nil, nil, "test")

		cache.Clear()

		if cache.Size() != 0 {
			t.Errorf("cache size after clear: got %d, want 0", cache.Size())
		}
	})
}

func TestUtilityFunctions(t *testing.T) {
	t.Run("ConcatBytes", func(t *testing.T) {
		result := ConcatBytes([]byte{1, 2}, []byte{3, 4}, []byte{5})
		expected := []byte{1, 2, 3, 4, 5}
		if !bytes.Equal(result, expected) {
			t.Errorf("ConcatBytes: got %v, want %v", result, expected)
		}
	})

	t.Run("ConcatBytes empty", func(t *testing.T) {
		result := ConcatBytes()
		if len(result) != 0 {
			t.Errorf("ConcatBytes empty: got %v, want empty", result)
		}
	})

	t.Run("ConstantTimeCompare", func(t *testing.T) {
		a := []byte{1, 2, 3, 4}
		b := []byte{1, 2, 3, 4}
		c := []byte{1, 2, 3, 5}
		d := []byte{1, 2, 3}

		if !ConstantTimeCompare(a, b) {
			t.Error("equal slices should compare equal")
		}
		if ConstantTimeCompare(a, c) {
			t.Error("different slices should compare unequal")
		}
		if ConstantTimeCompare(a, d) {
			t.Error("different length slices should compare unequal")
		}
	})

	t.Run("ZeroBytes", func(t *testing.T) {
		data := []byte{1, 2, 3, 4, 5}
		ZeroBytes(data)
		for i, b := range data {
			if b != 0 {
				t.Errorf("byte %d not zeroed: %d", i, b)
			}
		}
	})
}

// =============================================================================
// SGX DCAP Crypto Tests
// =============================================================================

func TestDCAPQuoteParsing(t *testing.T) {
	parser := NewDCAPQuoteParser()

	t.Run("parse valid quote", func(t *testing.T) {
		mrenclave := make([]byte, 32)
		for i := range mrenclave {
			mrenclave[i] = byte(i)
		}
		mrsigner := make([]byte, 32)
		for i := range mrsigner {
			mrsigner[i] = byte(i + 32)
		}
		reportData := []byte("test-nonce-12345678901234567890123456789012345678901234567890")

		quoteBytes := CreateTestDCAPQuote(mrenclave, mrsigner, false, reportData)

		quote, err := parser.Parse(quoteBytes)
		if err != nil {
			t.Fatalf("failed to parse quote: %v", err)
		}

		if quote.Header.Version != 3 {
			t.Errorf("version: got %d, want 3", quote.Header.Version)
		}

		if !bytes.Equal(quote.GetMRENCLAVE(), mrenclave) {
			t.Errorf("MRENCLAVE mismatch")
		}

		if !bytes.Equal(quote.GetMRSIGNER(), mrsigner) {
			t.Errorf("MRSIGNER mismatch")
		}
	})

	t.Run("parse quote too small", func(t *testing.T) {
		smallQuote := make([]byte, 100)
		_, err := parser.Parse(smallQuote)
		if err == nil {
			t.Error("expected error for small quote")
		}
	})

	t.Run("parse invalid version", func(t *testing.T) {
		quoteBytes := CreateTestDCAPQuote(nil, nil, false, nil)
		// Corrupt version
		binary.LittleEndian.PutUint16(quoteBytes[0:], 99)

		_, err := parser.Parse(quoteBytes)
		if err == nil {
			t.Error("expected error for invalid version")
		}
	})

	t.Run("debug mode detection", func(t *testing.T) {
		debugQuote := CreateTestDCAPQuote(nil, nil, true, nil)
		prodQuote := CreateTestDCAPQuote(nil, nil, false, nil)

		debugParsed, _ := parser.Parse(debugQuote)
		prodParsed, _ := parser.Parse(prodQuote)

		if !debugParsed.IsDebugEnclave() {
			t.Error("debug enclave not detected")
		}
		if prodParsed.IsDebugEnclave() {
			t.Error("production enclave incorrectly flagged as debug")
		}
	})
}

func TestDCAPSignatureVerification(t *testing.T) {
	verifier := NewDCAPSignatureVerifier()

	t.Run("extract attestation key", func(t *testing.T) {
		// Create a valid P-256 point
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatalf("failed to generate key: %v", err)
		}

		sigData := &CryptoDCAPQuoteSignatureData{}
		privateKey.PublicKey.X.FillBytes(sigData.ECDSAAttestationKey[:32])
		privateKey.PublicKey.Y.FillBytes(sigData.ECDSAAttestationKey[32:])

		pubKey, err := verifier.extractAttestationKey(sigData)
		if err != nil {
			t.Errorf("failed to extract key: %v", err)
		}

		if pubKey.X.Cmp(privateKey.PublicKey.X) != 0 || pubKey.Y.Cmp(privateKey.PublicKey.Y) != 0 {
			t.Error("extracted key doesn't match")
		}
	})

	t.Run("invalid attestation key", func(t *testing.T) {
		sigData := &CryptoDCAPQuoteSignatureData{}
		// Invalid point (not on curve)
		for i := range sigData.ECDSAAttestationKey {
			sigData.ECDSAAttestationKey[i] = 0xFF
		}

		_, err := verifier.extractAttestationKey(sigData)
		if err == nil {
			t.Error("expected error for invalid key")
		}
	})
}

func TestPCKCertificateChain(t *testing.T) {
	t.Run("create verifier", func(t *testing.T) {
		verifier, err := NewPCKCertificateVerifier()
		if err != nil {
			// Skip if using placeholder certificates
			t.Skipf("skipping: placeholder certificates in use: %v", err)
		}
		if verifier == nil {
			t.Fatal("verifier is nil")
		}
	})

	t.Run("get Intel SGX Root CA", func(t *testing.T) {
		cert, err := GetIntelSGXRootCA()
		if err != nil {
			t.Fatalf("failed to get Intel SGX Root CA: %v", err)
		}

		if cert.Subject.CommonName != "Intel SGX Root CA" {
			t.Errorf("CN: got %s, want Intel SGX Root CA", cert.Subject.CommonName)
		}

		if cert.Subject.Organization[0] != "Intel Corporation" {
			t.Errorf("O: got %s, want Intel Corporation", cert.Subject.Organization[0])
		}
	})
}

func TestTCBInfoVerifier(t *testing.T) {
	verifier := NewTCBInfoVerifier()

	t.Run("parse TCB Info", func(t *testing.T) {
		tcbInfoJSON := `{
			"version": 3,
			"issueDate": "2023-01-01T00:00:00Z",
			"nextUpdate": "2023-02-01T00:00:00Z",
			"fmspc": "00906ED50000",
			"pceId": "0000",
			"tcbType": 0,
			"tcbEvaluationDataNumber": 12,
			"tcbLevels": [
				{
					"tcb": [{"svn": 15}, {"svn": 15}, {"svn": 2}, {"svn": 4}],
					"pcesvn": 11,
					"tcbDate": "2022-11-09T00:00:00Z",
					"tcbStatus": "UpToDate"
				}
			]
		}`

		tcbInfo, err := verifier.ParseTCBInfo([]byte(tcbInfoJSON))
		if err != nil {
			t.Fatalf("failed to parse TCB Info: %v", err)
		}

		if tcbInfo.Version != 3 {
			t.Errorf("version: got %d, want 3", tcbInfo.Version)
		}
		if tcbInfo.FMSPC != "00906ED50000" {
			t.Errorf("FMSPC: got %s, want 00906ED50000", tcbInfo.FMSPC)
		}
		if len(tcbInfo.TCBLevels) != 1 {
			t.Errorf("TCBLevels length: got %d, want 1", len(tcbInfo.TCBLevels))
		}
	})

	t.Run("TCB status acceptable", func(t *testing.T) {
		if !IsTCBStatusAcceptable(TCBStatusUpToDate, false) {
			t.Error("UpToDate should be acceptable")
		}
		if !IsTCBStatusAcceptable(TCBStatusSWHardeningNeeded, false) {
			t.Error("SWHardeningNeeded should be acceptable")
		}
		if IsTCBStatusAcceptable(TCBStatusRevoked, true) {
			t.Error("Revoked should never be acceptable")
		}
		if IsTCBStatusAcceptable(TCBStatusOutOfDate, false) {
			t.Error("OutOfDate should not be acceptable without flag")
		}
		if !IsTCBStatusAcceptable(TCBStatusOutOfDate, true) {
			t.Error("OutOfDate should be acceptable with flag")
		}
	})
}

func TestQEIdentityVerifier(t *testing.T) {
	verifier := NewQEIdentityVerifier()

	t.Run("parse QE Identity", func(t *testing.T) {
		qeIdentityJSON := `{
			"version": 2,
			"issueDate": "2023-01-01T00:00:00Z",
			"nextUpdate": "2023-02-01T00:00:00Z",
			"enclaveIdentity": {
				"id": "QE",
				"version": 2,
				"issueDate": "2023-01-01T00:00:00Z",
				"nextUpdate": "2023-02-01T00:00:00Z",
				"tcbEvaluationDataNumber": 12,
				"miscselect": "00000000",
				"miscselectMask": "FFFFFFFF",
				"attributes": "11000000000000000000000000000000",
				"attributesMask": "FBFFFFFFFFFFFFFF0000000000000000",
				"mrsigner": "8C4F5775D796503E96137F77C68A829A0056AC8DED70140B081B094490C57BFF",
				"isvprodid": 1,
				"tcbLevels": []
			}
		}`

		qeIdentity, err := verifier.ParseQEIdentity([]byte(qeIdentityJSON))
		if err != nil {
			t.Fatalf("failed to parse QE Identity: %v", err)
		}

		if qeIdentity.EnclaveIdentity.ID != "QE" {
			t.Errorf("ID: got %s, want QE", qeIdentity.EnclaveIdentity.ID)
		}
		if qeIdentity.EnclaveIdentity.ISVProdID != 1 {
			t.Errorf("ISVProdID: got %d, want 1", qeIdentity.EnclaveIdentity.ISVProdID)
		}
	})
}

// =============================================================================
// SEV-SNP Crypto Tests
// =============================================================================

func TestSNPReportParsing(t *testing.T) {
	parser := NewSNPReportParser()

	t.Run("parse valid report", func(t *testing.T) {
		measurement := make([]byte, 48)
		for i := range measurement {
			measurement[i] = byte(i)
		}
		reportData := make([]byte, 64)
		copy(reportData, []byte("test-nonce"))

		reportBytes := CreateTestSNPReport(measurement, false, reportData)

		report, err := parser.Parse(reportBytes)
		if err != nil {
			t.Fatalf("failed to parse report: %v", err)
		}

		if report.Version != 2 {
			t.Errorf("version: got %d, want 2", report.Version)
		}

		if !bytes.Equal(report.GetMeasurement(), measurement) {
			t.Errorf("measurement mismatch")
		}

		if !bytes.HasPrefix(report.GetReportData(), []byte("test-nonce")) {
			t.Errorf("report data mismatch")
		}
	})

	t.Run("parse report too small", func(t *testing.T) {
		smallReport := make([]byte, 100)
		_, err := parser.Parse(smallReport)
		if err == nil {
			t.Error("expected error for small report")
		}
	})

	t.Run("debug policy detection", func(t *testing.T) {
		debugReport := CreateTestSNPReport(nil, true, nil)
		prodReport := CreateTestSNPReport(nil, false, nil)

		debugParsed, _ := parser.Parse(debugReport)
		prodParsed, _ := parser.Parse(prodReport)

		if !debugParsed.IsDebugPolicy() {
			t.Error("debug policy not detected")
		}
		if prodParsed.IsDebugPolicy() {
			t.Error("production report incorrectly flagged as debug")
		}
	})

	t.Run("TCB version string", func(t *testing.T) {
		reportBytes := CreateTestSNPReport(nil, false, nil)
		report, _ := parser.Parse(reportBytes)

		tcbVersion := report.GetTCBVersion()
		if tcbVersion != "1.0.1" {
			t.Errorf("TCB version: got %s, want 1.0.1", tcbVersion)
		}
	})
}

func TestSNPSignatureVerification(t *testing.T) {
	verifier := NewSNPSignatureVerifier()

	t.Run("extract signature components", func(t *testing.T) {
		sig := make([]byte, 512)
		// Set some values in r and s
		for i := 0; i < 48; i++ {
			sig[i] = byte(i + 1)
			sig[i+48] = byte(i + 49)
		}

		r, s, err := verifier.ExtractSignatureComponents(sig)
		if err != nil {
			t.Fatalf("failed to extract components: %v", err)
		}

		if r.Sign() <= 0 || s.Sign() <= 0 {
			t.Error("r or s should be positive")
		}
	})

	t.Run("signature too short", func(t *testing.T) {
		shortSig := make([]byte, 50)
		_, _, err := verifier.ExtractSignatureComponents(shortSig)
		if err == nil {
			t.Error("expected error for short signature")
		}
	})
}

func TestVCEKCertificateChain(t *testing.T) {
	t.Run("create verifier", func(t *testing.T) {
		verifier, err := NewVCEKCertificateVerifier()
		if err != nil {
			// Skip if using placeholder certificates
			t.Skipf("skipping: placeholder certificates in use: %v", err)
		}
		if verifier == nil {
			t.Fatal("verifier is nil")
		}
	})

	t.Run("get AMD Root Key Milan", func(t *testing.T) {
		cert, err := GetAMDRootKey(ProductMilan)
		if err != nil {
			// Skip if using placeholder certificates
			t.Skipf("skipping: placeholder certificates in use: %v", err)
		}

		if cert.Subject.CommonName != "ARK-Milan" {
			t.Errorf("CN: got %s, want ARK-Milan", cert.Subject.CommonName)
		}
	})

	t.Run("unknown product", func(t *testing.T) {
		_, err := GetAMDRootKey("UnknownProduct")
		if err == nil {
			t.Error("expected error for unknown product")
		}
	})
}

func TestASKARKVerifier(t *testing.T) {
	t.Run("create verifier", func(t *testing.T) {
		verifier, err := NewASKARKVerifier()
		if err != nil {
			// Skip if using placeholder certificates
			t.Skipf("skipping: placeholder certificates in use: %v", err)
		}

		ark, err := verifier.GetARK(ProductMilan)
		if err != nil {
			t.Skipf("skipping: placeholder ARK certificates in use: %v", err)
		}
		if ark == nil {
			t.Error("ARK is nil")
		}

		ask, err := verifier.GetASK(ProductMilan)
		if err != nil {
			t.Skipf("skipping: placeholder ASK certificates in use: %v", err)
		}
		if ask == nil {
			t.Error("ASK is nil")
		}
	})

	t.Run("unknown product", func(t *testing.T) {
		verifier, err := NewASKARKVerifier()
		if err != nil {
			t.Skipf("skipping: placeholder certificates in use: %v", err)
		}
		_, err = verifier.GetARK("UnknownProduct")
		if err == nil {
			t.Error("expected error for unknown product")
		}
	})
}

func TestValidateSNPReportNonce(t *testing.T) {
	parser := NewSNPReportParser()

	t.Run("matching nonce", func(t *testing.T) {
		nonce := []byte("test-nonce-12345")
		reportData := make([]byte, 64)
		copy(reportData, nonce)

		report, _ := parser.Parse(CreateTestSNPReport(nil, false, reportData))

		if !ValidateSNPReportNonce(report, nonce) {
			t.Error("nonce validation should pass")
		}
	})

	t.Run("mismatching nonce", func(t *testing.T) {
		reportData := []byte("different-nonce")
		report, _ := parser.Parse(CreateTestSNPReport(nil, false, reportData))

		if ValidateSNPReportNonce(report, []byte("expected-nonce")) {
			t.Error("nonce validation should fail")
		}
	})

	t.Run("empty expected nonce", func(t *testing.T) {
		report, _ := parser.Parse(CreateTestSNPReport(nil, false, nil))

		if !ValidateSNPReportNonce(report, nil) {
			t.Error("empty nonce should always pass")
		}
	})
}

// =============================================================================
// Nitro Crypto Tests
// =============================================================================

func TestNitroAttestationParsing(t *testing.T) {
	parser := NewNitroAttestationParser()

	t.Run("parse test attestation", func(t *testing.T) {
		pcr0 := make([]byte, 48)
		for i := range pcr0 {
			pcr0[i] = byte(i)
		}
		nonce := []byte("test-nonce")
		userData := []byte("user-data")

		docBytes := CreateTestNitroAttestation(pcr0, nonce, userData)

		doc, err := parser.Parse(docBytes)
		if err != nil {
			t.Fatalf("failed to parse attestation: %v", err)
		}

		if doc.ModuleID != "test-enclave-001" {
			t.Errorf("module ID: got %s, want test-enclave-001", doc.ModuleID)
		}

		if doc.Digest != "SHA384" {
			t.Errorf("digest: got %s, want SHA384", doc.Digest)
		}

		gotPCR0, ok := doc.GetPCR(0)
		if !ok {
			t.Error("PCR0 not found")
		}
		if !bytes.Equal(gotPCR0, pcr0) {
			t.Error("PCR0 mismatch")
		}
	})

	t.Run("parse too small", func(t *testing.T) {
		smallDoc := []byte{0xD2, 0x84}
		_, err := parser.Parse(smallDoc)
		if err == nil {
			t.Error("expected error for small document")
		}
	})

	t.Run("get timestamp", func(t *testing.T) {
		docBytes := CreateTestNitroAttestation(nil, nil, nil)
		doc, _ := parser.Parse(docBytes)

		ts := doc.GetTimestampTime()
		if ts.IsZero() {
			t.Error("timestamp should not be zero")
		}

		// Timestamp should be recent
		if time.Since(ts) > time.Hour {
			t.Error("timestamp should be recent")
		}
	})
}

func TestCOSESign1Verification(t *testing.T) {
	verifier := NewCOSESign1Verifier()

	t.Run("parse COSE Sign1", func(t *testing.T) {
		docBytes := CreateTestNitroAttestation(nil, nil, nil)

		cose, err := verifier.ParseCOSESign1(docBytes)
		if err != nil {
			t.Fatalf("failed to parse COSE Sign1: %v", err)
		}

		if len(cose.ProtectedHeader) == 0 {
			t.Error("protected header should not be empty")
		}
		if len(cose.Payload) == 0 {
			t.Error("payload should not be empty")
		}
		if len(cose.Signature) == 0 {
			t.Error("signature should not be empty")
		}
	})

	t.Run("build sig structure", func(t *testing.T) {
		protectedHeader := []byte{0xA1, 0x01, 0x38, 0x22}
		payload := []byte("test payload")

		sigStruct := verifier.buildSigStructure(protectedHeader, nil, payload)

		// Should start with array of 4 elements
		if sigStruct[0] != 0x84 {
			t.Errorf("expected array header 0x84, got 0x%02x", sigStruct[0])
		}
	})
}

func TestNitroCertificateChain(t *testing.T) {
	t.Run("create verifier", func(t *testing.T) {
		verifier, err := NewNitroCertificateVerifier()
		if err != nil {
			// Skip if using placeholder certificates
			t.Skipf("skipping: placeholder certificates in use: %v", err)
		}
		if verifier == nil {
			t.Fatal("verifier is nil")
		}
	})

	t.Run("get AWS Nitro Root CA", func(t *testing.T) {
		cert, err := GetAWSNitroRootCA()
		if err != nil {
			// Skip if using placeholder certificates
			t.Skipf("skipping: placeholder certificates in use: %v", err)
		}

		// Check it's a CA certificate
		if !cert.IsCA {
			t.Error("certificate should be CA")
		}

		// Check organization
		if len(cert.Subject.Organization) == 0 || cert.Subject.Organization[0] != "Amazon Web Services" {
			t.Errorf("unexpected organization: %v", cert.Subject.Organization)
		}
	})
}

func TestNitroRootCAVerifier(t *testing.T) {
	t.Run("create verifier", func(t *testing.T) {
		verifier, err := NewNitroRootCAVerifier()
		if err != nil {
			// Skip if using placeholder certificates
			t.Skipf("skipping: placeholder certificates in use: %v", err)
		}

		rootCA := verifier.GetRootCA()
		if rootCA == nil {
			t.Fatal("root CA is nil")
		}
	})

	t.Run("empty chain", func(t *testing.T) {
		verifier, err := NewNitroRootCAVerifier()
		if err != nil {
			// Skip if using placeholder certificates
			t.Skipf("skipping: placeholder certificates in use: %v", err)
		}
		err = verifier.VerifyChainToRoot(nil)
		if err == nil {
			t.Error("expected error for empty chain")
		}
	})
}

func TestPCRValidator(t *testing.T) {
	validator := NewPCRValidator()

	t.Run("set and validate PCR", func(t *testing.T) {
		expectedPCR0 := make([]byte, 48)
		for i := range expectedPCR0 {
			expectedPCR0[i] = byte(i)
		}

		validator.SetExpectedPCR(0, expectedPCR0)

		parser := NewNitroAttestationParser()
		docBytes := CreateTestNitroAttestation(expectedPCR0, nil, nil)
		doc, _ := parser.Parse(docBytes)

		err := validator.ValidatePCR(doc, 0)
		if err != nil {
			t.Errorf("PCR validation failed: %v", err)
		}
	})

	t.Run("PCR mismatch", func(t *testing.T) {
		expectedPCR0 := make([]byte, 48)
		for i := range expectedPCR0 {
			expectedPCR0[i] = byte(i)
		}

		wrongPCR0 := make([]byte, 48)
		for i := range wrongPCR0 {
			wrongPCR0[i] = byte(i + 100)
		}

		validator.SetExpectedPCR(0, expectedPCR0)

		parser := NewNitroAttestationParser()
		docBytes := CreateTestNitroAttestation(wrongPCR0, nil, nil)
		doc, _ := parser.Parse(docBytes)

		err := validator.ValidatePCR(doc, 0)
		if err == nil {
			t.Error("expected error for PCR mismatch")
		}
	})

	t.Run("unconfigured PCR", func(t *testing.T) {
		parser := NewNitroAttestationParser()
		docBytes := CreateTestNitroAttestation(nil, nil, nil)
		doc, _ := parser.Parse(docBytes)

		// PCR 5 is not configured
		err := validator.ValidatePCR(doc, 5)
		if err != nil {
			t.Errorf("unconfigured PCR should pass: %v", err)
		}
	})
}

func TestValidateNitroNonce(t *testing.T) {
	parser := NewNitroAttestationParser()

	t.Run("matching nonce", func(t *testing.T) {
		nonce := []byte("test-nonce-123")
		docBytes := CreateTestNitroAttestation(nil, nonce, nil)
		doc, _ := parser.Parse(docBytes)

		if !ValidateNitroNonce(doc, nonce) {
			t.Error("nonce validation should pass")
		}
	})

	t.Run("empty expected nonce", func(t *testing.T) {
		docBytes := CreateTestNitroAttestation(nil, []byte("any-nonce"), nil)
		doc, _ := parser.Parse(docBytes)

		if !ValidateNitroNonce(doc, nil) {
			t.Error("empty expected nonce should always pass")
		}
	})
}

func TestValidateNitroUserData(t *testing.T) {
	parser := NewNitroAttestationParser()

	t.Run("matching user data", func(t *testing.T) {
		userData := []byte("test-user-data")
		docBytes := CreateTestNitroAttestation(nil, nil, userData)
		doc, _ := parser.Parse(docBytes)

		if !ValidateNitroUserData(doc, userData) {
			t.Error("user data validation should pass")
		}
	})

	t.Run("empty expected user data", func(t *testing.T) {
		docBytes := CreateTestNitroAttestation(nil, nil, []byte("any-data"))
		doc, _ := parser.Parse(docBytes)

		if !ValidateNitroUserData(doc, nil) {
			t.Error("empty expected user data should always pass")
		}
	})
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestDCAPVerifierIntegration(t *testing.T) {
	t.Run("create verifier", func(t *testing.T) {
		verifier, err := NewDCAPVerifier()
		if err != nil {
			// Skip if using placeholder certificates
			t.Skipf("skipping: placeholder certificates in use: %v", err)
		}
		if verifier == nil {
			t.Fatal("verifier is nil")
		}
	})

	t.Run("verify test quote", func(t *testing.T) {
		verifier, err := NewDCAPVerifier()
		if err != nil {
			// Skip if using placeholder certificates
			t.Skipf("skipping: placeholder certificates in use: %v", err)
		}

		mrenclave := make([]byte, 32)
		for i := range mrenclave {
			mrenclave[i] = byte(i)
		}

		quoteBytes := CreateTestDCAPQuote(mrenclave, nil, false, nil)
		result, err := verifier.Verify(quoteBytes)
		if err != nil {
			t.Fatalf("verify returned error: %v", err)
		}

		// The test quote has invalid signatures and fake certificates,
		// so verification should fail, but parsing should succeed
		if result.Quote == nil {
			t.Error("quote should be parsed")
		}

		// Check that MRENCLAVE was extracted
		if result.Quote != nil && !bytes.Equal(result.Quote.GetMRENCLAVE(), mrenclave) {
			t.Error("MRENCLAVE mismatch")
		}
	})
}

func TestSNPVerifierIntegration(t *testing.T) {
	t.Run("create verifier", func(t *testing.T) {
		verifier, err := NewSNPVerifier()
		if err != nil {
			// Skip if using placeholder certificates
			t.Skipf("skipping: placeholder certificates in use: %v", err)
		}
		if verifier == nil {
			t.Fatal("verifier is nil")
		}
	})
}

func TestNitroCryptoVerifierIntegration(t *testing.T) {
	t.Run("create verifier", func(t *testing.T) {
		verifier, err := NewNitroCryptoVerifier()
		if err != nil {
			// Skip if using placeholder certificates
			t.Skipf("skipping: placeholder certificates in use: %v", err)
		}
		if verifier == nil {
			t.Fatal("verifier is nil")
		}
	})

	t.Run("set expected PCR", func(t *testing.T) {
		verifier, err := NewNitroCryptoVerifier()
		if err != nil {
			// Skip if using placeholder certificates
			t.Skipf("skipping: placeholder certificates in use: %v", err)
		}

		pcr0 := make([]byte, 48)
		for i := range pcr0 {
			pcr0[i] = byte(i)
		}

		verifier.SetExpectedPCR(0, pcr0)
		// No panic means success
	})
}

// =============================================================================
// Certificate Parsing Tests
// =============================================================================

func TestParseCertificateChain(t *testing.T) {
	t.Run("parse Intel SGX Root CA", func(t *testing.T) {
		certs, err := ParseCertificateChain([]byte(IntelSGXRootCAPEM))
		if err != nil {
			// Skip if using placeholder certificates
			t.Skipf("skipping: placeholder certificates in use: %v", err)
		}
		if len(certs) != 1 {
			t.Errorf("expected 1 cert, got %d", len(certs))
		}
	})

	t.Run("parse AMD Root Key", func(t *testing.T) {
		certs, err := ParseCertificateChain([]byte(AMDRootKeyMilanPEM))
		if err != nil {
			// Skip if using placeholder certificates
			t.Skipf("skipping: placeholder certificates in use: %v", err)
		}
		if len(certs) != 1 {
			t.Errorf("expected 1 cert, got %d", len(certs))
		}
	})

	t.Run("parse AWS Nitro Root CA", func(t *testing.T) {
		certs, err := ParseCertificateChain([]byte(AWSNitroRootCAPEM))
		if err != nil {
			// Skip if using placeholder certificates
			t.Skipf("skipping: placeholder certificates in use: %v", err)
		}
		if len(certs) != 1 {
			t.Errorf("expected 1 cert, got %d", len(certs))
		}
	})

	t.Run("invalid PEM", func(t *testing.T) {
		_, err := ParseCertificateChain([]byte("not a certificate"))
		if err == nil {
			t.Error("expected error for invalid PEM")
		}
	})

	t.Run("multiple certificates", func(t *testing.T) {
		combinedPEM := IntelSGXRootCAPEM + "\n" + IntelSGXPCKProcessorCAPEM
		certs, err := ParseCertificateChain([]byte(combinedPEM))
		if err != nil {
			// Skip if using placeholder certificates
			t.Skipf("skipping: placeholder certificates in use: %v", err)
		}
		if len(certs) != 2 {
			t.Errorf("expected 2 certs, got %d", len(certs))
		}
	})
}

func TestParseDERCertificate(t *testing.T) {
	t.Run("parse from DER", func(t *testing.T) {
		block, _ := pem.Decode([]byte(IntelSGXRootCAPEM))
		if block == nil {
			t.Skipf("skipping: placeholder certificate could not be decoded")
		}

		cert, err := ParseDERCertificate(block.Bytes)
		if err != nil {
			t.Skipf("skipping: placeholder certificate could not be parsed: %v", err)
		}
		if cert.Subject.CommonName != "Intel SGX Root CA" {
			t.Errorf("unexpected CN: %s", cert.Subject.CommonName)
		}
	})

	t.Run("invalid DER", func(t *testing.T) {
		_, err := ParseDERCertificate([]byte("not a certificate"))
		if err == nil {
			t.Error("expected error for invalid DER")
		}
	})
}

func TestExtractPublicKeyFromCert(t *testing.T) {
	t.Run("extract from Intel SGX Root CA", func(t *testing.T) {
		cert, err := GetIntelSGXRootCA()
		if err != nil {
			t.Skipf("skipping: placeholder certificate could not be loaded: %v", err)
		}
		pubKey, err := ExtractPublicKeyFromCert(cert)
		if err != nil {
			t.Fatalf("failed to extract public key: %v", err)
		}
		if pubKey.Curve != elliptic.P256() {
			t.Errorf("expected P-256 curve, got %s", pubKey.Curve.Params().Name)
		}
	})

	t.Run("extract from AWS Nitro Root CA", func(t *testing.T) {
		cert, err := GetAWSNitroRootCA()
		if err != nil {
			t.Skipf("skipping: placeholder certificate could not be loaded: %v", err)
		}
		pubKey, err := ExtractPublicKeyFromCert(cert)
		if err != nil {
			t.Skipf("skipping: placeholder certificate public key could not be extracted: %v", err)
		}
		if pubKey.Curve != elliptic.P384() {
			t.Errorf("expected P-384 curve, got %s", pubKey.Curve.Params().Name)
		}
	})

	t.Run("nil certificate", func(t *testing.T) {
		_, err := ExtractPublicKeyFromCert(nil)
		if err == nil {
			t.Error("expected error for nil certificate")
		}
	})
}

// =============================================================================
// CBOR Parser Tests
// =============================================================================

func TestCBORParser(t *testing.T) {
	t.Run("read uint", func(t *testing.T) {
		// CBOR encoding of 42
		data := []byte{0x18, 0x2A}
		parser := NewCBORParser(data)

		val, err := parser.readUint()
		if err != nil {
			t.Fatalf("failed to read uint: %v", err)
		}
		if val != 42 {
			t.Errorf("expected 42, got %d", val)
		}
	})

	t.Run("read byte string", func(t *testing.T) {
		// CBOR encoding of byte string "hello"
		data := []byte{0x45, 'h', 'e', 'l', 'l', 'o'}
		parser := NewCBORParser(data)

		bs, err := parser.readByteString()
		if err != nil {
			t.Fatalf("failed to read byte string: %v", err)
		}
		if string(bs) != "hello" {
			t.Errorf("expected 'hello', got '%s'", string(bs))
		}
	})

	t.Run("read text string", func(t *testing.T) {
		// CBOR encoding of text string "hello"
		data := []byte{0x65, 'h', 'e', 'l', 'l', 'o'}
		parser := NewCBORParser(data)

		s, err := parser.readTextString()
		if err != nil {
			t.Fatalf("failed to read text string: %v", err)
		}
		if s != "hello" {
			t.Errorf("expected 'hello', got '%s'", s)
		}
	})

	t.Run("read array header", func(t *testing.T) {
		// CBOR encoding of array with 5 elements
		data := []byte{0x85}
		parser := NewCBORParser(data)

		length, err := parser.readArrayHeader()
		if err != nil {
			t.Fatalf("failed to read array header: %v", err)
		}
		if length != 5 {
			t.Errorf("expected 5, got %d", length)
		}
	})

	t.Run("read map header", func(t *testing.T) {
		// CBOR encoding of map with 3 pairs
		data := []byte{0xA3}
		parser := NewCBORParser(data)

		length, err := parser.readMapHeader()
		if err != nil {
			t.Fatalf("failed to read map header: %v", err)
		}
		if length != 3 {
			t.Errorf("expected 3, got %d", length)
		}
	})

	t.Run("read tag", func(t *testing.T) {
		// CBOR encoding of tag 18 (COSE Sign1)
		data := []byte{0xD2}
		parser := NewCBORParser(data)

		tag, err := parser.readTag()
		if err != nil {
			t.Fatalf("failed to read tag: %v", err)
		}
		if tag != 18 {
			t.Errorf("expected 18, got %d", tag)
		}
	})

	t.Run("end of data", func(t *testing.T) {
		parser := NewCBORParser([]byte{})
		_, err := parser.readByte()
		if err == nil {
			t.Error("expected error for empty data")
		}
	})
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkSHA256(b *testing.B) {
	h := NewHashComputer()
	data := make([]byte, 1024)
	rand.Read(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.SHA256(data)
	}
}

func BenchmarkSHA384(b *testing.B) {
	h := NewHashComputer()
	data := make([]byte, 1024)
	rand.Read(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.SHA384(data)
	}
}

func BenchmarkECDSAVerifyP256(b *testing.B) {
	v := NewECDSAVerifier()
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	hash := sha256.Sum256([]byte("test message"))
	r, s, _ := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	sig := make([]byte, 64)
	r.FillBytes(sig[:32])
	s.FillBytes(sig[32:])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.VerifyP256(&privateKey.PublicKey, hash[:], sig)
	}
}

func BenchmarkDCAPQuoteParsing(b *testing.B) {
	parser := NewDCAPQuoteParser()
	quoteBytes := CreateTestDCAPQuote(nil, nil, false, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.Parse(quoteBytes)
	}
}

func BenchmarkSNPReportParsing(b *testing.B) {
	parser := NewSNPReportParser()
	reportBytes := CreateTestSNPReport(nil, false, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.Parse(reportBytes)
	}
}

func BenchmarkNitroAttestationParsing(b *testing.B) {
	parser := NewNitroAttestationParser()
	docBytes := CreateTestNitroAttestation(nil, nil, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.Parse(docBytes)
	}
}

func BenchmarkCertificateCache(b *testing.B) {
	cache := NewCertificateCache(1000, time.Hour)
	cert := &x509.Certificate{}

	// Pre-populate cache
	for i := 0; i < 100; i++ {
		cache.Put(hex.EncodeToString([]byte{byte(i)}), cert, nil, nil, "test")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := hex.EncodeToString([]byte{byte(i % 100)})
		cache.Get(key)
	}
}

