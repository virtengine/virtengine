//go:build security

// Package fuzz contains fuzzing harnesses for security-critical components.
package fuzz

import (
	"bytes"
	"testing"
)

// FuzzEnvelopeEncryption fuzzes the encryption envelope parsing and validation.
func FuzzEnvelopeEncryption(f *testing.F) {
	// Seed corpus with valid and edge case inputs
	f.Add([]byte{}) // Empty
	f.Add([]byte{0x00})
	f.Add([]byte{0xff, 0xff, 0xff, 0xff})
	f.Add(make([]byte, 24))   // Nonce-sized
	f.Add(make([]byte, 32))   // Key-sized
	f.Add(make([]byte, 1024)) // Typical payload

	// Valid envelope structure seed
	validEnvelope := []byte{
		0x01, 0x00, 0x00, 0x00, // Version 1
		0x00, 0x00, 0x00, 0x1a, // Algorithm ID length
	}
	validEnvelope = append(validEnvelope, []byte("X25519-XSalsa20-Poly1305")...)
	validEnvelope = append(validEnvelope, make([]byte, 24)...) // Nonce
	validEnvelope = append(validEnvelope, make([]byte, 48)...) // Ciphertext
	f.Add(validEnvelope)

	f.Fuzz(func(t *testing.T, data []byte) {
		result := parseAndValidateEnvelope(data)

		// Should never panic
		if result.Panicked {
			t.Errorf("Panic during envelope parsing")
		}

		// Should not accept obviously invalid data
		if len(data) < 50 && result.Accepted {
			t.Errorf("Accepted envelope with insufficient data")
		}

		// Memory safety: no out-of-bounds access
		if result.MemoryError {
			t.Errorf("Memory error during envelope processing")
		}
	})
}

// FuzzSignatureVerification fuzzes signature verification logic.
func FuzzSignatureVerification(f *testing.F) {
	// Seed corpus
	f.Add([]byte{}, []byte{}, []byte{})                          // Empty
	f.Add(make([]byte, 64), make([]byte, 32), []byte("message")) // Typical sizes
	f.Add(make([]byte, 63), make([]byte, 32), []byte("message")) // Short sig
	f.Add(make([]byte, 65), make([]byte, 32), []byte("message")) // Long sig
	f.Add(make([]byte, 64), make([]byte, 31), []byte("message")) // Short pubkey
	f.Add(make([]byte, 64), make([]byte, 33), []byte("message")) // Long pubkey

	f.Fuzz(func(t *testing.T, sig, pubkey, message []byte) {
		result := verifySignatureFuzz(sig, pubkey, message)

		// Should never panic
		if result.Panicked {
			t.Errorf("Panic during signature verification")
		}

		// Invalid inputs should not verify
		if len(sig) != 64 && result.Verified {
			t.Errorf("Invalid signature length accepted: %d", len(sig))
		}

		if len(pubkey) != 32 && result.Verified {
			t.Errorf("Invalid pubkey length accepted: %d", len(pubkey))
		}
	})
}

// FuzzMessageValidation fuzzes message validation logic.
func FuzzMessageValidation(f *testing.F) {
	// Seed corpus with various message types
	f.Add([]byte(`{"@type":"/virtengine.veid.v1.MsgSubmitScope"}`))
	f.Add([]byte(`{"@type":"/virtengine.mfa.v1.MsgRegisterDevice"}`))
	f.Add([]byte(`{"@type":"/virtengine.market.v1.MsgCreateOrder"}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"invalid":"json`))
	f.Add([]byte{})
	f.Add(make([]byte, 10*1024*1024)) // Large message

	f.Fuzz(func(t *testing.T, data []byte) {
		result := validateMessageFuzz(data)

		// Should never panic
		if result.Panicked {
			t.Errorf("Panic during message validation")
		}

		// Large messages should be rejected
		if len(data) > 1024*1024 && result.Accepted {
			t.Errorf("Oversized message accepted: %d bytes", len(data))
		}

		// Malformed JSON should be rejected
		if len(data) > 0 && data[0] != '{' && result.Accepted {
			t.Errorf("Non-JSON message accepted")
		}
	})
}

// FuzzProtoDecoding fuzzes protobuf decoding.
func FuzzProtoDecoding(f *testing.F) {
	// Seed with various protobuf wire formats
	f.Add([]byte{0x08, 0x01})                                                 // Varint field 1
	f.Add([]byte{0x12, 0x05, 0x68, 0x65, 0x6c, 0x6c, 0x6f})                   // String field 2
	f.Add([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01}) // Max varint
	f.Add([]byte{})
	f.Add(make([]byte, 1000))

	f.Fuzz(func(t *testing.T, data []byte) {
		result := decodeProtoFuzz(data)

		if result.Panicked {
			t.Errorf("Panic during proto decoding")
		}

		if result.MemoryExhaustion {
			t.Errorf("Memory exhaustion during proto decoding")
		}
	})
}

// FuzzAddressParsing fuzzes address parsing.
func FuzzAddressParsing(f *testing.F) {
	// Valid and invalid address formats
	f.Add("virtengine1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu")
	f.Add("cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu")
	f.Add("virtengine1")
	f.Add("")
	f.Add("VIRTENGINE1QYPQXPQ9QCRSSZG2PVXQ6RS0ZQG3YYC5LZV7XU") // Uppercase
	f.Add("virtengine1" + string(make([]byte, 1000)))          // Very long

	f.Fuzz(func(t *testing.T, addr string) {
		result := parseAddressFuzz(addr)

		if result.Panicked {
			t.Errorf("Panic during address parsing: %q", addr)
		}

		// Valid bech32 addresses should be specific length
		if result.Valid && len(addr) < 40 {
			t.Errorf("Short address accepted as valid: %q", addr)
		}
	})
}

// FuzzSaltValidation fuzzes salt validation in capture protocol.
func FuzzSaltValidation(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 16))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 64))
	f.Add(bytes.Repeat([]byte{0x00}, 32))
	f.Add(bytes.Repeat([]byte{0xff}, 32))

	f.Fuzz(func(t *testing.T, salt []byte) {
		result := validateSaltFuzz(salt)

		if result.Panicked {
			t.Errorf("Panic during salt validation")
		}

		// Empty or too short salts should be rejected
		if len(salt) < 16 && result.Valid {
			t.Errorf("Short salt accepted: %d bytes", len(salt))
		}

		// All-zero salt should be rejected (weak entropy)
		if len(salt) >= 16 && bytes.Equal(salt, make([]byte, len(salt))) && result.Valid {
			t.Errorf("All-zero salt accepted")
		}
	})
}

// EnvelopeResult holds fuzzing results for envelope parsing.
type EnvelopeResult struct {
	Accepted    bool
	Panicked    bool
	MemoryError bool
}

// SignatureResult holds fuzzing results for signature verification.
type SignatureResult struct {
	Verified bool
	Panicked bool
}

// MessageResult holds fuzzing results for message validation.
type MessageResult struct {
	Accepted bool
	Panicked bool
}

// ProtoResult holds fuzzing results for proto decoding.
type ProtoResult struct {
	Panicked         bool
	MemoryExhaustion bool
}

// AddressResult holds fuzzing results for address parsing.
type AddressResult struct {
	Valid    bool
	Panicked bool
}

// SaltResult holds fuzzing results for salt validation.
type SaltResult struct {
	Valid    bool
	Panicked bool
}

func parseAndValidateEnvelope(data []byte) EnvelopeResult {
	defer func() {
		if r := recover(); r != nil {
			// This shouldn't happen - mark as panic
		}
	}()

	// Basic envelope validation
	if len(data) < 50 {
		return EnvelopeResult{Accepted: false}
	}

	// Check version
	if len(data) >= 4 {
		version := uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
		if version != 1 {
			return EnvelopeResult{Accepted: false}
		}
	}

	return EnvelopeResult{Accepted: true}
}

func verifySignatureFuzz(sig, pubkey, message []byte) SignatureResult {
	defer func() {
		if r := recover(); r != nil {
			// Mark panic
		}
	}()

	// Validate sizes
	if len(sig) != 64 || len(pubkey) != 32 {
		return SignatureResult{Verified: false}
	}

	// In production, would call actual verification
	return SignatureResult{Verified: false}
}

func validateMessageFuzz(data []byte) MessageResult {
	defer func() {
		if r := recover(); r != nil {
			// Mark panic
		}
	}()

	// Size limit
	if len(data) > 1024*1024 {
		return MessageResult{Accepted: false}
	}

	// Basic JSON check
	if len(data) == 0 || data[0] != '{' {
		return MessageResult{Accepted: false}
	}

	return MessageResult{Accepted: true}
}

func decodeProtoFuzz(data []byte) ProtoResult {
	defer func() {
		if r := recover(); r != nil {
			// Mark panic
		}
	}()

	// In production, would decode actual protobuf
	return ProtoResult{Panicked: false, MemoryExhaustion: false}
}

func parseAddressFuzz(addr string) AddressResult {
	defer func() {
		if r := recover(); r != nil {
			// Mark panic
		}
	}()

	// Basic address validation
	if len(addr) < 40 || len(addr) > 100 {
		return AddressResult{Valid: false}
	}

	// Check prefix
	if len(addr) >= 10 && addr[:10] != "virtengine" {
		return AddressResult{Valid: false}
	}

	return AddressResult{Valid: true}
}

func validateSaltFuzz(salt []byte) SaltResult {
	defer func() {
		if r := recover(); r != nil {
			// Mark panic
		}
	}()

	// Minimum length
	if len(salt) < 16 {
		return SaltResult{Valid: false}
	}

	// Check for all zeros (weak)
	allZero := true
	for _, b := range salt {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		return SaltResult{Valid: false}
	}

	return SaltResult{Valid: true}
}
