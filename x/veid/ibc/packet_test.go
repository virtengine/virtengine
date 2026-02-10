// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package ibc

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestEncodeDecodeVEIDAttestationPacket(t *testing.T) {
	addr := sdk.AccAddress([]byte("packet-test-acct")).String()
	original := VEIDAttestationPacket{
		SourceChainID:   "cosmoshub-4",
		AccountAddress:  addr,
		VEIDHash:        []byte("hash_data_123"),
		TrustScore:      750,
		TierLevel:       3,
		AttestationTime: 1000000,
		ExpirationTime:  2000000,
		ValidatorSet:    []string{"val1", "val2", "val3"},
		MerkleProof:     []byte("merkle_proof_data"),
		StateRootHash:   []byte("state_root_hash"),
		Nonce:           42,
	}

	// Encode
	encoded, err := EncodeVEIDAttestationPacket(original)
	if err != nil {
		t.Fatalf("EncodeVEIDAttestationPacket() error = %v", err)
	}
	if len(encoded) == 0 {
		t.Fatal("EncodeVEIDAttestationPacket() returned empty bytes")
	}

	// Decode
	decoded, err := DecodeVEIDAttestationPacket(encoded)
	if err != nil {
		t.Fatalf("DecodeVEIDAttestationPacket() error = %v", err)
	}

	// Compare
	if decoded.SourceChainID != original.SourceChainID {
		t.Errorf("SourceChainID = %s, want %s", decoded.SourceChainID, original.SourceChainID)
	}
	if decoded.AccountAddress != original.AccountAddress {
		t.Errorf("AccountAddress = %s, want %s", decoded.AccountAddress, original.AccountAddress)
	}
	if decoded.TrustScore != original.TrustScore {
		t.Errorf("TrustScore = %d, want %d", decoded.TrustScore, original.TrustScore)
	}
	if decoded.TierLevel != original.TierLevel {
		t.Errorf("TierLevel = %d, want %d", decoded.TierLevel, original.TierLevel)
	}
	if decoded.Nonce != original.Nonce {
		t.Errorf("Nonce = %d, want %d", decoded.Nonce, original.Nonce)
	}
}

func TestEncodeVEIDAttestationPacket_InvalidPacket(t *testing.T) {
	// Empty source chain should fail validation
	invalid := VEIDAttestationPacket{}
	_, err := EncodeVEIDAttestationPacket(invalid)
	if err == nil {
		t.Error("EncodeVEIDAttestationPacket() should fail for invalid packet")
	}
}

func TestDecodeVEIDAttestationPacket_InvalidData(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"empty data", []byte{}},
		{"invalid json", []byte("not json")},
		{"wrong packet type", []byte(`{"type":"unknown","data":{}}`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeVEIDAttestationPacket(tt.data)
			if err == nil {
				t.Error("DecodeVEIDAttestationPacket() should fail for invalid data")
			}
		})
	}
}

func TestEncodeDecodeAck(t *testing.T) {
	original := VEIDAttestationAck{
		Success:         true,
		RecognizedScore: 720,
	}

	encoded, err := EncodeAck(original)
	if err != nil {
		t.Fatalf("EncodeAck() error = %v", err)
	}

	decoded, err := DecodeAck(encoded)
	if err != nil {
		t.Fatalf("DecodeAck() error = %v", err)
	}

	if decoded.Success != original.Success {
		t.Errorf("Success = %v, want %v", decoded.Success, original.Success)
	}
	if decoded.RecognizedScore != original.RecognizedScore {
		t.Errorf("RecognizedScore = %d, want %d", decoded.RecognizedScore, original.RecognizedScore)
	}
}

func TestEncodeDecodeAck_Error(t *testing.T) {
	original := VEIDAttestationAck{
		Success: false,
		Error:   "chain not trusted",
	}

	encoded, err := EncodeAck(original)
	if err != nil {
		t.Fatalf("EncodeAck() error = %v", err)
	}

	decoded, err := DecodeAck(encoded)
	if err != nil {
		t.Fatalf("DecodeAck() error = %v", err)
	}

	if decoded.Success != false {
		t.Error("expected failure ack")
	}
	if decoded.Error != original.Error {
		t.Errorf("Error = %s, want %s", decoded.Error, original.Error)
	}
}

func TestDecodeAck_InvalidData(t *testing.T) {
	_, err := DecodeAck([]byte("invalid"))
	if err == nil {
		t.Error("DecodeAck() should fail for invalid data")
	}
}
