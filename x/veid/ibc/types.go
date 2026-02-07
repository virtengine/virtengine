// Copyright 2024 VirtEngine Contributors
// SPDX-License-Identifier: Apache-2.0

package ibc

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// PortID is the IBC port ID for the VEID module
	PortID = "veid"

	// Version is the IBC version for the VEID attestation protocol
	Version = "veid-1"
)

// VEIDAttestationPacket defines the IBC packet format for cross-chain VEID attestations.
type VEIDAttestationPacket struct {
	// SourceChainID is the chain ID where the VEID was originally attested
	SourceChainID string `json:"source_chain_id"`

	// AccountAddress is the bech32 account address of the VEID holder
	AccountAddress string `json:"account_address"`

	// VEIDHash is the SHA256 hash of the full VEID record
	VEIDHash []byte `json:"veid_hash"`

	// TrustScore is the VEID trust score (0-1000)
	TrustScore uint32 `json:"trust_score"`

	// TierLevel is the VEID tier level (1-5)
	TierLevel uint32 `json:"tier_level"`

	// AttestationTime is the block time when the attestation was created (Unix seconds)
	AttestationTime int64 `json:"attestation_time"`

	// ExpirationTime is the block time when the attestation expires (Unix seconds)
	ExpirationTime int64 `json:"expiration_time"`

	// ValidatorSet is the list of validator addresses who attested this VEID
	ValidatorSet []string `json:"validator_set"`

	// MerkleProof is the Merkle proof of the VEID state in the source chain
	MerkleProof []byte `json:"merkle_proof"`

	// StateRootHash is the state root hash of the source chain at attestation time
	StateRootHash []byte `json:"state_root_hash"`

	// Nonce prevents replay attacks
	Nonce uint64 `json:"nonce"`
}

// Validate performs basic validation of the attestation packet.
func (p VEIDAttestationPacket) Validate() error {
	if p.SourceChainID == "" {
		return fmt.Errorf("source chain ID cannot be empty")
	}
	if p.AccountAddress == "" {
		return fmt.Errorf("account address cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(p.AccountAddress); err != nil {
		return fmt.Errorf("invalid account address: %w", err)
	}
	if len(p.VEIDHash) == 0 {
		return fmt.Errorf("VEID hash cannot be empty")
	}
	if p.TrustScore > 1000 {
		return fmt.Errorf("trust score must be between 0 and 1000, got %d", p.TrustScore)
	}
	if p.TierLevel == 0 || p.TierLevel > 5 {
		return fmt.Errorf("tier level must be between 1 and 5, got %d", p.TierLevel)
	}
	if p.AttestationTime <= 0 {
		return fmt.Errorf("attestation time must be positive")
	}
	if p.ExpirationTime <= p.AttestationTime {
		return fmt.Errorf("expiration time must be after attestation time")
	}
	if len(p.ValidatorSet) == 0 {
		return fmt.Errorf("validator set cannot be empty")
	}
	if p.Nonce == 0 {
		return fmt.Errorf("nonce must be non-zero")
	}
	return nil
}

// CrossChainScorePolicy defines score degradation rules for cross-chain recognition.
type CrossChainScorePolicy struct {
	// TrustedChains maps ChainID to the maximum score allowed from that chain
	TrustedChains map[string]uint32 `json:"trusted_chains"`

	// DegradationFactor is the score multiplier applied to cross-chain scores (e.g., 0.9)
	DegradationFactor sdkmath.LegacyDec `json:"degradation_factor"`

	// MinRecognizedTier is the minimum tier level accepted from remote chains
	MinRecognizedTier uint32 `json:"min_recognized_tier"`

	// RequiredValidators is the minimum number of validators required to attest
	RequiredValidators uint32 `json:"required_validators"`
}

// DefaultCrossChainScorePolicy returns the default cross-chain score policy.
func DefaultCrossChainScorePolicy() CrossChainScorePolicy {
	return CrossChainScorePolicy{
		TrustedChains: map[string]uint32{
			"cosmoshub-4": 900,
			"osmosis-1":   800,
		},
		DegradationFactor:  sdkmath.LegacyNewDecWithPrec(9, 1), // 0.9
		MinRecognizedTier:  2,
		RequiredValidators: 3,
	}
}

// Validate performs basic validation of the cross-chain score policy.
func (p CrossChainScorePolicy) Validate() error {
	if p.DegradationFactor.IsNegative() || p.DegradationFactor.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("degradation factor must be between 0 and 1, got %s", p.DegradationFactor)
	}
	if p.MinRecognizedTier == 0 || p.MinRecognizedTier > 5 {
		return fmt.Errorf("min recognized tier must be between 1 and 5, got %d", p.MinRecognizedTier)
	}
	if p.RequiredValidators == 0 {
		return fmt.Errorf("required validators must be positive")
	}
	for chainID, maxScore := range p.TrustedChains {
		if chainID == "" {
			return fmt.Errorf("trusted chain ID cannot be empty")
		}
		if maxScore > 1000 {
			return fmt.Errorf("max score for chain %s must be <= 1000, got %d", chainID, maxScore)
		}
	}
	return nil
}

// ApplyDegradation applies the cross-chain score degradation policy to an attestation.
func (p CrossChainScorePolicy) ApplyDegradation(packet VEIDAttestationPacket) (uint32, error) {
	maxScore, trusted := p.TrustedChains[packet.SourceChainID]
	if !trusted {
		return 0, fmt.Errorf("chain %s is not trusted", packet.SourceChainID)
	}

	if packet.TierLevel < p.MinRecognizedTier {
		return 0, fmt.Errorf("tier level %d is below minimum recognized tier %d", packet.TierLevel, p.MinRecognizedTier)
	}

	if len(packet.ValidatorSet) < int(p.RequiredValidators) {
		return 0, fmt.Errorf("insufficient validators: got %d, need %d", len(packet.ValidatorSet), p.RequiredValidators)
	}

	// Apply degradation factor
	degradedScore := sdkmath.LegacyNewDec(int64(packet.TrustScore)).Mul(p.DegradationFactor).TruncateInt64()
	if degradedScore < 0 {
		degradedScore = 0
	}
	score := uint32(degradedScore) //nolint:gosec // degradedScore is guaranteed non-negative and <= 1000

	// Cap at chain-specific maximum
	if score > maxScore {
		score = maxScore
	}

	return score, nil
}

// VEIDAttestationAck defines the acknowledgement for a VEID attestation packet.
type VEIDAttestationAck struct {
	// Success indicates whether the attestation was accepted
	Success bool `json:"success"`

	// Error contains the error message if the attestation was rejected
	Error string `json:"error,omitempty"`

	// RecognizedScore is the score recognized on the destination chain
	RecognizedScore uint32 `json:"recognized_score,omitempty"`
}

// CrossChainVEIDRecord represents a VEID attestation received from another chain.
type CrossChainVEIDRecord struct {
	// SourceChainID is the originating chain
	SourceChainID string `json:"source_chain_id"`

	// SourceChannel is the IBC channel from which the attestation was received
	SourceChannel string `json:"source_channel"`

	// AccountAddress is the bech32 address of the VEID holder
	AccountAddress string `json:"account_address"`

	// OriginalScore is the score on the source chain
	OriginalScore uint32 `json:"original_score"`

	// RecognizedScore is the degraded score recognized on this chain
	RecognizedScore uint32 `json:"recognized_score"`

	// TierLevel is the original tier level
	TierLevel uint32 `json:"tier_level"`

	// ReceivedAt is the block time when the attestation was received (Unix seconds)
	ReceivedAt int64 `json:"received_at"`

	// ExpiresAt is when this cross-chain recognition expires (Unix seconds)
	ExpiresAt int64 `json:"expires_at"`

	// VEIDHash is the hash of the original VEID record
	VEIDHash []byte `json:"veid_hash"`
}
