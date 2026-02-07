// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package ibc

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// veidTestAddr generates a valid bech32 address from a test string.
func veidTestAddr(name string) string {
	return sdk.AccAddress([]byte(name)).String()
}

func TestVEIDAttestationPacket_Validate(t *testing.T) {
	validPacket := VEIDAttestationPacket{
		SourceChainID:   "cosmoshub-4",
		AccountAddress:  veidTestAddr("valid-account"),
		VEIDHash:        []byte("hash123"),
		TrustScore:      800,
		TierLevel:       3,
		AttestationTime: 1000,
		ExpirationTime:  2000,
		ValidatorSet:    []string{"val1", "val2"},
		MerkleProof:     []byte("proof"),
		StateRootHash:   []byte("state_root"),
		Nonce:           1,
	}

	tests := []struct {
		name    string
		modify  func(*VEIDAttestationPacket)
		wantErr bool
	}{
		{
			name:    "valid packet",
			modify:  func(p *VEIDAttestationPacket) {},
			wantErr: false,
		},
		{
			name:    "empty source chain ID",
			modify:  func(p *VEIDAttestationPacket) { p.SourceChainID = "" },
			wantErr: true,
		},
		{
			name:    "empty account address",
			modify:  func(p *VEIDAttestationPacket) { p.AccountAddress = "" },
			wantErr: true,
		},
		{
			name:    "invalid account address",
			modify:  func(p *VEIDAttestationPacket) { p.AccountAddress = "invalid" },
			wantErr: true,
		},
		{
			name:    "empty VEID hash",
			modify:  func(p *VEIDAttestationPacket) { p.VEIDHash = nil },
			wantErr: true,
		},
		{
			name:    "trust score too high",
			modify:  func(p *VEIDAttestationPacket) { p.TrustScore = 1001 },
			wantErr: true,
		},
		{
			name:    "tier level zero",
			modify:  func(p *VEIDAttestationPacket) { p.TierLevel = 0 },
			wantErr: true,
		},
		{
			name:    "tier level too high",
			modify:  func(p *VEIDAttestationPacket) { p.TierLevel = 6 },
			wantErr: true,
		},
		{
			name:    "zero attestation time",
			modify:  func(p *VEIDAttestationPacket) { p.AttestationTime = 0 },
			wantErr: true,
		},
		{
			name:    "expiration before attestation",
			modify:  func(p *VEIDAttestationPacket) { p.ExpirationTime = 500 },
			wantErr: true,
		},
		{
			name:    "empty validator set",
			modify:  func(p *VEIDAttestationPacket) { p.ValidatorSet = nil },
			wantErr: true,
		},
		{
			name:    "zero nonce",
			modify:  func(p *VEIDAttestationPacket) { p.Nonce = 0 },
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := validPacket
			tt.modify(&p)
			err := p.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCrossChainScorePolicy_Validate(t *testing.T) {
	tests := []struct {
		name    string
		policy  CrossChainScorePolicy
		wantErr bool
	}{
		{
			name:    "default policy is valid",
			policy:  DefaultCrossChainScorePolicy(),
			wantErr: false,
		},
		{
			name: "negative degradation factor",
			policy: CrossChainScorePolicy{
				DegradationFactor:  sdkmath.LegacyNewDec(-1),
				MinRecognizedTier:  2,
				RequiredValidators: 3,
				TrustedChains:      map[string]uint32{"cosmoshub-4": 900},
			},
			wantErr: true,
		},
		{
			name: "degradation factor > 1",
			policy: CrossChainScorePolicy{
				DegradationFactor:  sdkmath.LegacyNewDec(2),
				MinRecognizedTier:  2,
				RequiredValidators: 3,
				TrustedChains:      map[string]uint32{"cosmoshub-4": 900},
			},
			wantErr: true,
		},
		{
			name: "zero min recognized tier",
			policy: CrossChainScorePolicy{
				DegradationFactor:  sdkmath.LegacyNewDecWithPrec(9, 1),
				MinRecognizedTier:  0,
				RequiredValidators: 3,
				TrustedChains:      map[string]uint32{"cosmoshub-4": 900},
			},
			wantErr: true,
		},
		{
			name: "zero required validators",
			policy: CrossChainScorePolicy{
				DegradationFactor:  sdkmath.LegacyNewDecWithPrec(9, 1),
				MinRecognizedTier:  2,
				RequiredValidators: 0,
				TrustedChains:      map[string]uint32{"cosmoshub-4": 900},
			},
			wantErr: true,
		},
		{
			name: "max score too high",
			policy: CrossChainScorePolicy{
				DegradationFactor:  sdkmath.LegacyNewDecWithPrec(9, 1),
				MinRecognizedTier:  2,
				RequiredValidators: 3,
				TrustedChains:      map[string]uint32{"cosmoshub-4": 1001},
			},
			wantErr: true,
		},
		{
			name: "empty chain ID in trusted chains",
			policy: CrossChainScorePolicy{
				DegradationFactor:  sdkmath.LegacyNewDecWithPrec(9, 1),
				MinRecognizedTier:  2,
				RequiredValidators: 3,
				TrustedChains:      map[string]uint32{"": 900},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.policy.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCrossChainScorePolicy_ApplyDegradation(t *testing.T) {
	policy := DefaultCrossChainScorePolicy()

	tests := []struct {
		name      string
		packet    VEIDAttestationPacket
		wantScore uint32
		wantErr   bool
	}{
		{
			name: "normal degradation from cosmoshub",
			packet: VEIDAttestationPacket{
				SourceChainID:   "cosmoshub-4",
				AccountAddress:  veidTestAddr("degradation-normal"),
				VEIDHash:        []byte("hash"),
				TrustScore:      800,
				TierLevel:       3,
				AttestationTime: 1000,
				ExpirationTime:  2000,
				ValidatorSet:    []string{"val1", "val2", "val3"},
				Nonce:           1,
			},
			wantScore: 720, // 800 * 0.9 = 720
			wantErr:   false,
		},
		{
			name: "score capped at chain maximum",
			packet: VEIDAttestationPacket{
				SourceChainID:   "osmosis-1",
				AccountAddress:  veidTestAddr("degradation-capped"),
				VEIDHash:        []byte("hash"),
				TrustScore:      1000,
				TierLevel:       5,
				AttestationTime: 1000,
				ExpirationTime:  2000,
				ValidatorSet:    []string{"val1", "val2", "val3"},
				Nonce:           1,
			},
			wantScore: 800, // 1000 * 0.9 = 900, capped at 800 (osmosis max)
			wantErr:   false,
		},
		{
			name: "untrusted chain",
			packet: VEIDAttestationPacket{
				SourceChainID:   "unknown-chain",
				AccountAddress:  veidTestAddr("degradation-untrust"),
				VEIDHash:        []byte("hash"),
				TrustScore:      800,
				TierLevel:       3,
				AttestationTime: 1000,
				ExpirationTime:  2000,
				ValidatorSet:    []string{"val1", "val2", "val3"},
				Nonce:           1,
			},
			wantErr: true,
		},
		{
			name: "tier below minimum",
			packet: VEIDAttestationPacket{
				SourceChainID:   "cosmoshub-4",
				AccountAddress:  veidTestAddr("degradation-lowtier"),
				VEIDHash:        []byte("hash"),
				TrustScore:      800,
				TierLevel:       1, // Below min tier of 2
				AttestationTime: 1000,
				ExpirationTime:  2000,
				ValidatorSet:    []string{"val1", "val2", "val3"},
				Nonce:           1,
			},
			wantErr: true,
		},
		{
			name: "insufficient validators",
			packet: VEIDAttestationPacket{
				SourceChainID:   "cosmoshub-4",
				AccountAddress:  veidTestAddr("degradation-fewvals"),
				VEIDHash:        []byte("hash"),
				TrustScore:      800,
				TierLevel:       3,
				AttestationTime: 1000,
				ExpirationTime:  2000,
				ValidatorSet:    []string{"val1", "val2"}, // Need 3
				Nonce:           1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, err := policy.ApplyDegradation(tt.packet)
			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyDegradation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && score != tt.wantScore {
				t.Errorf("ApplyDegradation() score = %d, want %d", score, tt.wantScore)
			}
		})
	}
}

func TestDefaultCrossChainScorePolicy(t *testing.T) {
	policy := DefaultCrossChainScorePolicy()

	if err := policy.Validate(); err != nil {
		t.Errorf("DefaultCrossChainScorePolicy() should be valid: %v", err)
	}

	if len(policy.TrustedChains) != 2 {
		t.Errorf("expected 2 trusted chains, got %d", len(policy.TrustedChains))
	}

	if policy.MinRecognizedTier != 2 {
		t.Errorf("expected min tier 2, got %d", policy.MinRecognizedTier)
	}

	if policy.RequiredValidators != 3 {
		t.Errorf("expected 3 required validators, got %d", policy.RequiredValidators)
	}
}
