// Package nitro provides AWS Nitro Enclave integration for VirtEngine TEE.
//
// This file contains tests for the nitro package.
package nitro

import (
	"bytes"
	"crypto/rand"
	"strings"
	"testing"
	"time"
)

// =============================================================================
// Enclave Tests
// =============================================================================

func TestNewNitroEnclave(t *testing.T) {
	ne, err := NewNitroEnclave()
	if err != nil {
		t.Fatalf("NewNitroEnclave failed: %v", err)
	}

	// Should be in simulation mode on non-Nitro hardware
	if !ne.IsSimulated() {
		t.Log("Running on actual Nitro hardware")
	} else {
		t.Log("Running in simulation mode")
	}
}

func TestNitroEnclaveRunSimulated(t *testing.T) {
	ne, err := NewNitroEnclave()
	if err != nil {
		t.Fatalf("NewNitroEnclave failed: %v", err)
	}

	// Run enclave in simulation mode
	info, err := ne.RunEnclave("/tmp/test.eif", 2, 2048)
	if err != nil {
		t.Fatalf("RunEnclave failed: %v", err)
	}

	if info == nil {
		t.Fatal("RunEnclave returned nil info")
	}

	if info.EnclaveID == "" {
		t.Error("EnclaveID is empty")
	}

	if info.EnclaveCID == 0 {
		t.Error("EnclaveCID is zero")
	}

	if info.NumberOfCPUs != 2 {
		t.Errorf("NumberOfCPUs: expected 2, got %d", info.NumberOfCPUs)
	}

	if info.MemoryMiB != 2048 {
		t.Errorf("MemoryMiB: expected 2048, got %d", info.MemoryMiB)
	}

	if info.State != EnclaveStateRunning {
		t.Errorf("State: expected RUNNING, got %s", info.State)
	}

	t.Logf("Simulated enclave: ID=%s, CID=%d", info.EnclaveID, info.EnclaveCID)
}

func TestNitroEnclaveDescribeSimulated(t *testing.T) {
	ne, err := NewNitroEnclave()
	if err != nil {
		t.Fatalf("NewNitroEnclave failed: %v", err)
	}

	// Run enclave
	info, err := ne.RunEnclave("/tmp/test.eif", 2, 2048)
	if err != nil {
		t.Fatalf("RunEnclave failed: %v", err)
	}

	// Describe enclave
	desc, err := ne.DescribeEnclave(info.EnclaveID)
	if err != nil {
		t.Fatalf("DescribeEnclave failed: %v", err)
	}

	if desc.EnclaveID != info.EnclaveID {
		t.Errorf("EnclaveID mismatch: expected %s, got %s", info.EnclaveID, desc.EnclaveID)
	}
}

func TestNitroEnclaveTerminateSimulated(t *testing.T) {
	ne, err := NewNitroEnclave()
	if err != nil {
		t.Fatalf("NewNitroEnclave failed: %v", err)
	}

	// Run enclave
	info, err := ne.RunEnclave("/tmp/test.eif", 2, 2048)
	if err != nil {
		t.Fatalf("RunEnclave failed: %v", err)
	}

	// Terminate enclave
	if err := ne.TerminateEnclave(info.EnclaveID); err != nil {
		t.Fatalf("TerminateEnclave failed: %v", err)
	}

	// Should not be able to describe after termination
	_, err = ne.DescribeEnclave(info.EnclaveID)
	if err != ErrEnclaveNotFound {
		t.Errorf("Expected ErrEnclaveNotFound, got %v", err)
	}
}

func TestNitroEnclaveListSimulated(t *testing.T) {
	ne, err := NewNitroEnclave()
	if err != nil {
		t.Fatalf("NewNitroEnclave failed: %v", err)
	}

	// Run two enclaves
	info1, _ := ne.RunEnclave("/tmp/test1.eif", 2, 2048)
	info2, _ := ne.RunEnclave("/tmp/test2.eif", 4, 4096)

	// List enclaves
	enclaves, err := ne.ListEnclaves()
	if err != nil {
		t.Fatalf("ListEnclaves failed: %v", err)
	}

	if len(enclaves) != 2 {
		t.Errorf("Expected 2 enclaves, got %d", len(enclaves))
	}

	// Clean up
	ne.TerminateEnclave(info1.EnclaveID)
	ne.TerminateEnclave(info2.EnclaveID)
}

func TestNitroEnclaveBuildSimulated(t *testing.T) {
	ne, err := NewNitroEnclave()
	if err != nil {
		t.Fatalf("NewNitroEnclave failed: %v", err)
	}

	// Build in simulation should create measurements
	tmpFile := t.TempDir() + "/test.eif"
	if err := ne.BuildEnclave("test-image:latest", tmpFile); err != nil {
		t.Fatalf("BuildEnclave failed: %v", err)
	}

	measurements := ne.GetMeasurements()
	if measurements == nil {
		t.Fatal("GetMeasurements returned nil")
	}

	if measurements.HashAlgorithm != "SHA384" {
		t.Errorf("HashAlgorithm: expected SHA384, got %s", measurements.HashAlgorithm)
	}

	if measurements.PCR0 == "" {
		t.Error("PCR0 is empty")
	}

	t.Logf("Simulated PCR0: %s", measurements.PCR0)
}

// =============================================================================
// PCR Tests
// =============================================================================

func TestPCRValue(t *testing.T) {
	var pcr PCRValue

	// Zero check
	if !pcr.IsZero() {
		t.Error("Empty PCR should be zero")
	}

	// Set a value
	pcr[0] = 0x01
	if pcr.IsZero() {
		t.Error("Non-zero PCR should not be zero")
	}

	// String representation
	str := pcr.String()
	if len(str) != PCRDigestSize*2 {
		t.Errorf("String length: expected %d, got %d", PCRDigestSize*2, len(str))
	}
}

func TestPCRSet(t *testing.T) {
	var set PCRSet

	// Test Get/Set
	var testPCR PCRValue
	testPCR[0] = 0xAB
	testPCR[47] = 0xCD

	set.Set(0, testPCR)
	retrieved := set.Get(0)

	if !bytes.Equal(retrieved[:], testPCR[:]) {
		t.Error("PCR value mismatch after Get/Set")
	}

	// Test out of bounds
	outOfBounds := set.Get(100)
	if !outOfBounds.IsZero() {
		t.Error("Out of bounds Get should return zero")
	}
}

func TestParsePCRFromHex(t *testing.T) {
	// Valid PCR
	validHex := "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f"
	pcr, err := ParsePCRFromHex(validHex)
	if err != nil {
		t.Fatalf("ParsePCRFromHex failed: %v", err)
	}

	if pcr[0] != 0x00 || pcr[1] != 0x01 || pcr[47] != 0x2f {
		t.Error("Parsed PCR values incorrect")
	}

	// Invalid hex
	_, err = ParsePCRFromHex("xyz")
	if err == nil {
		t.Error("Expected error for invalid hex")
	}

	// Wrong length
	_, err = ParsePCRFromHex("0001020304")
	if err == nil {
		t.Error("Expected error for wrong length")
	}
}

// =============================================================================
// Document Tests
// =============================================================================

func TestDocumentPayloadValidation(t *testing.T) {
	// Valid document
	doc := &AttestationDocument{
		Payload: &DocumentPayload{
			ModuleID:    "test-module",
			Digest:      DigestAlgorithmSHA384,
			Timestamp:   uint64(time.Now().UnixMilli()),
			PCRs:        map[int][]byte{0: make([]byte, PCRDigestSize)},
			Certificate: []byte("test-cert"),
		},
		Signature: []byte("test-signature"),
	}

	// Set PCR0 to non-zero
	doc.Payload.PCRs[0][0] = 0x01

	if err := ValidateDocument(doc); err != nil {
		t.Fatalf("ValidateDocument failed for valid doc: %v", err)
	}

	// Missing module_id
	doc.Payload.ModuleID = ""
	if err := ValidateDocument(doc); err == nil {
		t.Error("Expected error for missing module_id")
	}
	doc.Payload.ModuleID = "test-module"

	// Wrong digest
	doc.Payload.Digest = "SHA256"
	if err := ValidateDocument(doc); err == nil {
		t.Error("Expected error for wrong digest")
	}
	doc.Payload.Digest = DigestAlgorithmSHA384

	// Missing PCR0
	delete(doc.Payload.PCRs, 0)
	if err := ValidateDocument(doc); err == nil {
		t.Error("Expected error for missing PCR0")
	}
}

func TestGetPCRDigest(t *testing.T) {
	pcrs := map[int][]byte{
		0: make([]byte, PCRDigestSize),
		1: make([]byte, PCRDigestSize),
		2: make([]byte, PCRDigestSize),
	}

	// Set some values
	pcrs[0][0] = 0x01
	pcrs[1][0] = 0x02
	pcrs[2][0] = 0x03

	digest := GetPCRDigest(pcrs)
	if len(digest) != PCRDigestSize {
		t.Errorf("Digest length: expected %d, got %d", PCRDigestSize, len(digest))
	}

	// Same input should produce same output
	digest2 := GetPCRDigest(pcrs)
	if !bytes.Equal(digest, digest2) {
		t.Error("GetPCRDigest not deterministic")
	}
}

func TestPCRMapConversions(t *testing.T) {
	original := map[int][]byte{
		0: make([]byte, PCRDigestSize),
		1: make([]byte, PCRDigestSize),
		2: make([]byte, PCRDigestSize),
	}

	// Set values
	original[0][0] = 0xAB
	original[1][0] = 0xCD
	original[2][0] = 0xEF

	// Convert to hex
	hexMap := PCRMapToHex(original)

	// Convert back
	converted, err := PCRMapFromHex(hexMap)
	if err != nil {
		t.Fatalf("PCRMapFromHex failed: %v", err)
	}

	// Compare
	for idx, orig := range original {
		conv, ok := converted[idx]
		if !ok {
			t.Errorf("Missing PCR%d after conversion", idx)
			continue
		}
		if !bytes.Equal(orig, conv) {
			t.Errorf("PCR%d mismatch after round-trip", idx)
		}
	}
}

// =============================================================================
// NSM Tests
// =============================================================================

func TestNewNSMDevice(t *testing.T) {
	device := NewNSMDevice()
	if device == nil {
		t.Fatal("NewNSMDevice returned nil")
	}

	if device.IsOpen() {
		t.Error("Device should not be open initially")
	}
}

func TestNSMDeviceOpenClose(t *testing.T) {
	device := NewNSMDevice()

	// Open (should fall back to simulation)
	if err := device.Open(); err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	if !device.IsOpen() {
		t.Error("Device should be open after Open()")
	}

	// Should be simulated on non-Nitro hardware
	if !device.IsSimulated() {
		t.Log("Running on actual Nitro hardware")
	} else {
		t.Log("Running in simulation mode")
	}

	// Close
	if err := device.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if device.IsOpen() {
		t.Error("Device should not be open after Close()")
	}
}

func TestNSMGetAttestationSimulated(t *testing.T) {
	device := NewNSMDevice()
	if err := device.Open(); err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer device.Close()

	userData := []byte("test user data")
	nonce := make([]byte, 32)
	rand.Read(nonce)
	pubKey := []byte("test public key")

	attestation, err := device.GetAttestation(userData, nonce, pubKey)
	if err != nil {
		t.Fatalf("GetAttestation failed: %v", err)
	}

	if len(attestation) == 0 {
		t.Error("Attestation document is empty")
	}

	t.Logf("Simulated attestation document: %d bytes", len(attestation))
}

func TestNSMDescribePCRSimulated(t *testing.T) {
	device := NewNSMDevice()
	if err := device.Open(); err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer device.Close()

	// Describe PCR0
	desc, err := device.DescribePCR(0)
	if err != nil {
		t.Fatalf("DescribePCR failed: %v", err)
	}

	if desc.Index != 0 {
		t.Errorf("Index: expected 0, got %d", desc.Index)
	}

	if len(desc.Value) != PCRDigestSize {
		t.Errorf("Value length: expected %d, got %d", PCRDigestSize, len(desc.Value))
	}

	// Invalid index
	_, err = device.DescribePCR(100)
	if err != ErrNSMInvalidPCRIndex {
		t.Errorf("Expected ErrNSMInvalidPCRIndex, got %v", err)
	}
}

func TestNSMExtendPCRSimulated(t *testing.T) {
	device := NewNSMDevice()
	if err := device.Open(); err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer device.Close()

	// Get initial PCR value
	before, _ := device.DescribePCR(3)

	// Extend PCR
	data := []byte("test extension data")
	if err := device.ExtendPCR(3, data); err != nil {
		t.Fatalf("ExtendPCR failed: %v", err)
	}

	// Get new value
	after, _ := device.DescribePCR(3)

	// Values should be different
	if bytes.Equal(before.Value, after.Value) {
		t.Error("PCR value should change after extension")
	}
}

func TestNSMLockPCRSimulated(t *testing.T) {
	device := NewNSMDevice()
	if err := device.Open(); err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer device.Close()

	// Lock PCR
	if err := device.LockPCR(5); err != nil {
		t.Fatalf("LockPCR failed: %v", err)
	}

	// Check locked status
	desc, _ := device.DescribePCR(5)
	if !desc.Locked {
		t.Error("PCR should be locked")
	}

	// Extend should fail
	if err := device.ExtendPCR(5, []byte("test")); err != ErrNSMPCRLocked {
		t.Errorf("Expected ErrNSMPCRLocked, got %v", err)
	}
}

func TestNSMGetRandomSimulated(t *testing.T) {
	device := NewNSMDevice()
	if err := device.Open(); err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer device.Close()

	// Get random bytes
	random, err := device.GetRandomBytes(32)
	if err != nil {
		t.Fatalf("GetRandomBytes failed: %v", err)
	}

	if len(random) != 32 {
		t.Errorf("Length: expected 32, got %d", len(random))
	}

	// Get another batch - should be different
	random2, _ := device.GetRandomBytes(32)
	if bytes.Equal(random, random2) {
		t.Error("Random bytes should be different each call")
	}

	// Invalid count
	_, err = device.GetRandomBytes(0)
	if err == nil {
		t.Error("Expected error for count 0")
	}

	_, err = device.GetRandomBytes(300)
	if err == nil {
		t.Error("Expected error for count > 256")
	}
}

func TestNSMDescribeNSMSimulated(t *testing.T) {
	device := NewNSMDevice()
	if err := device.Open(); err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer device.Close()

	info, err := device.DescribeNSM()
	if err != nil {
		t.Fatalf("DescribeNSM failed: %v", err)
	}

	if info.ModuleID == "" {
		t.Error("ModuleID is empty")
	}

	if info.MaxPCRs != 16 {
		t.Errorf("MaxPCRs: expected 16, got %d", info.MaxPCRs)
	}

	if info.Digest != DigestAlgorithmSHA384 {
		t.Errorf("Digest: expected SHA384, got %s", info.Digest)
	}

	t.Logf("NSM Info: ModuleID=%s, Version=%s", info.ModuleID, info.Version)
}

func TestNSMSession(t *testing.T) {
	session, err := NewNSMSession()
	if err != nil {
		t.Fatalf("NewNSMSession failed: %v", err)
	}
	defer session.Close()

	if !session.IsSimulated() {
		t.Log("Running on actual Nitro hardware")
	}

	// Test all operations through session
	_, err = session.GetAttestation(nil, nil, nil)
	if err != nil {
		t.Errorf("GetAttestation failed: %v", err)
	}

	_, err = session.GetRandomBytes(16)
	if err != nil {
		t.Errorf("GetRandomBytes failed: %v", err)
	}

	_, err = session.DescribePCR(0)
	if err != nil {
		t.Errorf("DescribePCR failed: %v", err)
	}
}

// =============================================================================
// Verifier Tests
// =============================================================================

func TestNewVerifier(t *testing.T) {
	v := NewVerifier()
	if v == nil {
		t.Fatal("NewVerifier returned nil")
	}
}

func TestVerifierConfig(t *testing.T) {
	config := DefaultVerifierConfig()
	if config.MaxDocumentAge != DefaultMaxDocumentAge {
		t.Errorf("MaxDocumentAge: expected %v, got %v", DefaultMaxDocumentAge, config.MaxDocumentAge)
	}

	if config.MaxClockSkew != DefaultMaxClockSkew {
		t.Errorf("MaxClockSkew: expected %v, got %v", DefaultMaxClockSkew, config.MaxClockSkew)
	}

	if config.RootCA == nil {
		t.Log("Root CA not embedded (expected in test environment)")
	}
}

func TestVerifierSetAllowedPCRs(t *testing.T) {
	v := NewVerifier()

	pcrs := map[int][]byte{
		0: make([]byte, PCRDigestSize),
		1: make([]byte, PCRDigestSize),
	}
	pcrs[0][0] = 0xAB
	pcrs[1][0] = 0xCD

	v.SetAllowedPCRs(pcrs)

	// Test hex version
	hexPCRs := map[int]string{
		0: "ab" + strings.Repeat("00", 47),
		1: "cd" + strings.Repeat("00", 47),
	}

	err := v.SetAllowedPCRsFromHex(hexPCRs)
	if err != nil {
		t.Fatalf("SetAllowedPCRsFromHex failed: %v", err)
	}
}

func TestVerifySimulatedAttestation(t *testing.T) {
	// Get simulated attestation
	device := NewNSMDevice()
	if err := device.Open(); err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer device.Close()

	attestation, err := device.GetAttestation([]byte("test"), nil, nil)
	if err != nil {
		t.Fatalf("GetAttestation failed: %v", err)
	}

	// Create verifier that allows simulated
	config := DefaultVerifierConfig()
	config.AllowSimulated = true
	config.SkipSignatureVerification = true
	config.SkipCertificateChainVerification = true

	v := NewVerifierWithConfig(config)

	result, err := v.VerifyRaw(attestation)
	if err != nil {
		t.Fatalf("VerifyRaw failed: %v", err)
	}

	if !result.Valid {
		t.Errorf("Verification should pass for simulated attestation: %v", result.Error)
	}

	if result.ModuleID == "" {
		t.Error("ModuleID should be present")
	}

	if len(result.Warnings) == 0 {
		t.Error("Should have warning about simulated document")
	}

	t.Logf("Verification result: ModuleID=%s, Warnings=%v", result.ModuleID, result.Warnings)
}

func TestPCRPolicy(t *testing.T) {
	policy := NewPCRPolicy()

	// Add PCR
	pcr := make([]byte, PCRDigestSize)
	pcr[0] = 0xAB
	if err := policy.AddPCR(0, pcr); err != nil {
		t.Fatalf("AddPCR failed: %v", err)
	}

	// Add from hex
	hexPCR := "cd" + strings.Repeat("00", 47)
	if err := policy.AddPCRHex(1, hexPCR); err != nil {
		t.Fatalf("AddPCRHex failed: %v", err)
	}

	// Get expected
	expected := policy.GetExpected()
	if len(expected) != 2 {
		t.Errorf("Expected 2 PCRs, got %d", len(expected))
	}

	// Validate matching
	actual := map[int][]byte{
		0: make([]byte, PCRDigestSize),
		1: make([]byte, PCRDigestSize),
	}
	actual[0][0] = 0xAB
	actual[1][0] = 0xCD

	if err := policy.Validate(actual); err != nil {
		t.Errorf("Validate failed for matching PCRs: %v", err)
	}

	// Validate non-matching
	actual[0][0] = 0xFF
	if err := policy.Validate(actual); err == nil {
		t.Error("Validate should fail for non-matching PCRs")
	}

	// Invalid index
	if err := policy.AddPCR(100, pcr); err == nil {
		t.Error("AddPCR should fail for invalid index")
	}

	// Invalid size
	if err := policy.AddPCR(0, []byte{0x01, 0x02}); err == nil {
		t.Error("AddPCR should fail for invalid size")
	}
}

// =============================================================================
// CBOR Tests
// =============================================================================

func TestCBORRoundTrip(t *testing.T) {
	// Create a simple payload
	payload := &DocumentPayload{
		ModuleID:  "test-module",
		Digest:    DigestAlgorithmSHA384,
		Timestamp: uint64(time.Now().UnixMilli()),
		PCRs: map[int][]byte{
			0: make([]byte, PCRDigestSize),
			1: make([]byte, PCRDigestSize),
		},
		Certificate: []byte("test-cert"),
		UserData:    []byte("test-user-data"),
		Nonce:       []byte("test-nonce"),
	}
	payload.PCRs[0][0] = 0x01

	// Serialize
	serialized := serializePayload(payload)
	if len(serialized) == 0 {
		t.Fatal("Serialization produced empty output")
	}

	// Parse back
	parsed, err := parsePayload(serialized)
	if err != nil {
		t.Fatalf("parsePayload failed: %v", err)
	}

	// Verify
	if parsed.ModuleID != payload.ModuleID {
		t.Errorf("ModuleID mismatch: expected %s, got %s", payload.ModuleID, parsed.ModuleID)
	}

	if parsed.Digest != payload.Digest {
		t.Errorf("Digest mismatch: expected %s, got %s", payload.Digest, parsed.Digest)
	}

	if parsed.Timestamp != payload.Timestamp {
		t.Errorf("Timestamp mismatch: expected %d, got %d", payload.Timestamp, parsed.Timestamp)
	}

	if !bytes.Equal(parsed.UserData, payload.UserData) {
		t.Error("UserData mismatch")
	}

	if !bytes.Equal(parsed.Nonce, payload.Nonce) {
		t.Error("Nonce mismatch")
	}
}

// =============================================================================
// Vsock Tests
// =============================================================================

func TestVsockConnection(t *testing.T) {
	vc := NewVsockConnection(16, 5000)
	if vc == nil {
		t.Fatal("NewVsockConnection returned nil")
	}

	// Connect (will fall back to simulation)
	if err := vc.Connect(nil); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	if !vc.IsSimulated() {
		t.Log("Running on actual vsock")
	}

	// Send (simulated)
	if err := vc.Send([]byte("test")); err != nil {
		t.Errorf("Send failed: %v", err)
	}

	// Receive (simulated)
	data, err := vc.Receive(1024)
	if err != nil {
		t.Errorf("Receive failed: %v", err)
	}
	if len(data) == 0 {
		t.Log("No data received (expected in simulation)")
	}

	// Close
	if err := vc.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

// =============================================================================
// Validation Tests
// =============================================================================

func TestValidateEnclaveID(t *testing.T) {
	tests := []struct {
		id      string
		wantErr bool
	}{
		{"i-1234567890abcdef", false},
		{"i-sim-abcdef123456", false},
		{"i-a1b2c3d4", false},
		{"", true},        // Empty
		{"invalid", true}, // No i- prefix
		{"i-XYZ", true},   // Invalid characters
		{"i-123", true},   // Too short
	}

	for _, tt := range tests {
		err := validateEnclaveID(tt.id)
		if (err != nil) != tt.wantErr {
			t.Errorf("validateEnclaveID(%q): got err=%v, wantErr=%v", tt.id, err, tt.wantErr)
		}
	}
}

func TestValidateDockerImage(t *testing.T) {
	tests := []struct {
		image   string
		wantErr bool
	}{
		{"nginx", false},
		{"nginx:latest", false},
		{"myrepo/myimage:v1.0", false},
		{"docker.io/library/nginx:latest", false},
		{"", true},         // Empty
		{"@invalid", true}, // Starts with @
		{"-invalid", true}, // Starts with -
	}

	for _, tt := range tests {
		err := validateDockerImage(tt.image)
		if (err != nil) != tt.wantErr {
			t.Errorf("validateDockerImage(%q): got err=%v, wantErr=%v", tt.image, err, tt.wantErr)
		}
	}
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		path    string
		wantErr bool
	}{
		{"/tmp/test.eif", false},
		{"/var/lib/enclave/image.eif", false},
		{"", true},                   // Empty
		{"/tmp/../etc/passwd", true}, // Path traversal
		{"/tmp/test;rm -rf /", true}, // Shell injection
		{"/tmp/test|cat", true},      // Pipe
		{"/tmp/test$(whoami)", true}, // Command substitution
	}

	for _, tt := range tests {
		_, err := sanitizePath(tt.path)
		if (err != nil) != tt.wantErr {
			t.Errorf("sanitizePath(%q): got err=%v, wantErr=%v", tt.path, err, tt.wantErr)
		}
	}
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestFullSimulatedWorkflow(t *testing.T) {
	// 1. Create enclave manager
	ne, err := NewNitroEnclave()
	if err != nil {
		t.Fatalf("NewNitroEnclave failed: %v", err)
	}

	// 2. Build enclave
	tmpDir := t.TempDir()
	eifPath := tmpDir + "/test.eif"
	if err := ne.BuildEnclave("test-image:latest", eifPath); err != nil {
		t.Fatalf("BuildEnclave failed: %v", err)
	}

	// 3. Run enclave
	info, err := ne.RunEnclave(eifPath, 2, 2048)
	if err != nil {
		t.Fatalf("RunEnclave failed: %v", err)
	}

	// 4. Create NSM session (simulated, inside enclave)
	nsm, err := NewNSMSession()
	if err != nil {
		t.Fatalf("NewNSMSession failed: %v", err)
	}
	defer nsm.Close()

	// 5. Get attestation
	attestation, err := nsm.GetAttestation([]byte("challenge"), nil, nil)
	if err != nil {
		t.Fatalf("GetAttestation failed: %v", err)
	}

	// 6. Verify attestation
	config := DefaultVerifierConfig()
	config.AllowSimulated = true
	config.SkipSignatureVerification = true
	config.SkipCertificateChainVerification = true
	verifier := NewVerifierWithConfig(config)

	result, err := verifier.VerifyRaw(attestation)
	if err != nil {
		t.Fatalf("VerifyRaw failed: %v", err)
	}

	if !result.Valid {
		t.Errorf("Verification failed: %v", result.Error)
	}

	// 7. Terminate enclave
	if err := ne.TerminateEnclave(info.EnclaveID); err != nil {
		t.Fatalf("TerminateEnclave failed: %v", err)
	}

	t.Logf("Full workflow completed successfully")
	t.Logf("  Enclave ID: %s", info.EnclaveID)
	t.Logf("  Module ID: %s", result.ModuleID)
	t.Logf("  Attestation size: %d bytes", len(attestation))
}
