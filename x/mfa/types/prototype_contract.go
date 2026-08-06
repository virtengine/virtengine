package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

const (
	FactorProfileContractVersion = "virtengine.mfa.factor-profile/v1"
	ChallengeContractVersion     = "virtengine.mfa.challenge/v1"
	MaxChallengeValiditySeconds  = int64(15 * 60)
)

// PrototypeProfileState prevents dependency-incomplete profiles from becoming active.
type PrototypeProfileState uint8

const (
	PrototypeProfileDisabled PrototypeProfileState = iota
	PrototypeProfileFixtureOnly
	PrototypeProfileSandbox
	PrototypeProfileProduction
)

func (s PrototypeProfileState) ValidateFixtureOnly() error {
	if s != PrototypeProfileDisabled && s != PrototypeProfileFixtureOnly {
		return fmt.Errorf("profile state %d exceeds fixture-only boundary", s)
	}
	return nil
}

// FactorProfile identifies the exact verifier policy under which factor proof is accepted.
type FactorProfile struct {
	ContractVersion      string                `json:"contract_version"`
	ProfileID            string                `json:"profile_id"`
	FactorType           FactorType            `json:"factor_type"`
	ChallengeVersion     string                `json:"challenge_version"`
	VerifierID           string                `json:"verifier_id"`
	VerifierKeyEpoch     uint64                `json:"verifier_key_epoch"`
	RootEpoch            uint64                `json:"root_epoch"`
	RevocationPolicy     string                `json:"revocation_policy"`
	FreshnessPolicy      string                `json:"freshness_policy"`
	ProofRequired        bool                  `json:"proof_required"`
	StrongFactorEligible bool                  `json:"strong_factor_eligible"`
	RecoveryEligible     bool                  `json:"recovery_eligible"`
	PriorPolicyRequired  bool                  `json:"prior_policy_required"`
	State                PrototypeProfileState `json:"state"`
	CoreReleaseDigest    string                `json:"core_release_digest,omitempty"`
	ExternalBlocker      string                `json:"external_blocker"`
}

func (p FactorProfile) ValidatePrototype() error {
	if p.ContractVersion != FactorProfileContractVersion {
		return fmt.Errorf("unsupported factor profile contract version %q", p.ContractVersion)
	}
	if p.ProfileID == "" || !p.FactorType.IsValid() {
		return fmt.Errorf("factor profile identity is incomplete")
	}
	if p.ChallengeVersion != ChallengeContractVersion {
		return fmt.Errorf("unsupported challenge version %q", p.ChallengeVersion)
	}
	if p.VerifierID == "" || p.VerifierKeyEpoch == 0 || p.RootEpoch == 0 {
		return fmt.Errorf("verifier identity and epochs are required")
	}
	if p.RevocationPolicy == "" || p.FreshnessPolicy == "" {
		return fmt.Errorf("revocation and freshness policies are required")
	}
	if !p.ProofRequired {
		return fmt.Errorf("factor proof must be required")
	}
	if p.RecoveryEligible && (!p.StrongFactorEligible || !p.PriorPolicyRequired) {
		return fmt.Errorf("recovery eligibility requires a strong factor and prior policy")
	}
	if err := p.State.ValidateFixtureOnly(); err != nil {
		return err
	}
	if p.CoreReleaseDigest != "" {
		return fmt.Errorf("core release digest must remain unset before 88D")
	}
	if p.ExternalBlocker == "" {
		return fmt.Errorf("fixture-only profile must name its external blocker")
	}
	return nil
}

// CanonicalFactorChallenge binds factor proof to one chain, account, action, and verifier epoch.
type CanonicalFactorChallenge struct {
	ContractVersion        string     `json:"contract_version"`
	ChainID                string     `json:"chain_id"`
	AccountAddress         string     `json:"account_address"`
	Action                 string     `json:"action"`
	FactorProfileID        string     `json:"factor_profile_id"`
	FactorType             FactorType `json:"factor_type"`
	FactorID               string     `json:"factor_id"`
	PublicIdentifierDigest [32]byte   `json:"public_identifier_digest"`
	FactorMetadataDigest   [32]byte   `json:"factor_metadata_digest"`
	DeviceBindingDigest    [32]byte   `json:"device_binding_digest"`
	VerifierID             string     `json:"verifier_id"`
	VerifierKeyEpoch       uint64     `json:"verifier_key_epoch"`
	RootEpoch              uint64     `json:"root_epoch"`
	Nonce                  [32]byte   `json:"nonce"`
	Origin                 string     `json:"origin,omitempty"`
	RelyingPartyID         string     `json:"relying_party_id,omitempty"`
	IssuedAt               int64      `json:"issued_at"`
	ExpiresAt              int64      `json:"expires_at"`
}

func (c CanonicalFactorChallenge) Validate() error {
	if c.ContractVersion != ChallengeContractVersion {
		return fmt.Errorf("unsupported challenge contract version %q", c.ContractVersion)
	}
	if c.ChainID == "" || c.AccountAddress == "" || c.Action == "" {
		return fmt.Errorf("challenge chain, account, and action are required")
	}
	if c.FactorProfileID == "" || c.FactorID == "" || c.VerifierID == "" {
		return fmt.Errorf("challenge factor and verifier identity are required")
	}
	if !c.FactorType.IsValid() {
		return fmt.Errorf("challenge factor type is invalid")
	}
	if c.VerifierKeyEpoch == 0 || c.RootEpoch == 0 {
		return fmt.Errorf("challenge verifier epochs are required")
	}
	if c.PublicIdentifierDigest == ([32]byte{}) || c.FactorMetadataDigest == ([32]byte{}) ||
		c.DeviceBindingDigest == ([32]byte{}) || c.Nonce == ([32]byte{}) {
		return fmt.Errorf("challenge digests and nonce must be non-zero")
	}
	if c.IssuedAt <= 0 || c.ExpiresAt <= c.IssuedAt || c.ExpiresAt-c.IssuedAt > MaxChallengeValiditySeconds {
		return fmt.Errorf("challenge validity window is invalid")
	}
	return nil
}

func (c CanonicalFactorChallenge) CanonicalBytes() ([]byte, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	var output bytes.Buffer
	// Field order is part of challenge/v1 and must not change.
	for _, value := range []string{
		c.ContractVersion,
		c.ChainID,
		c.AccountAddress,
		c.Action,
		c.FactorProfileID,
	} {
		if err := writeCanonicalString(&output, value); err != nil {
			return nil, err
		}
	}
	output.WriteByte(byte(c.FactorType))
	if err := writeCanonicalString(&output, c.FactorID); err != nil {
		return nil, err
	}
	output.Write(c.PublicIdentifierDigest[:])
	output.Write(c.FactorMetadataDigest[:])
	output.Write(c.DeviceBindingDigest[:])
	if err := writeCanonicalString(&output, c.VerifierID); err != nil {
		return nil, err
	}
	_ = binary.Write(&output, binary.BigEndian, c.VerifierKeyEpoch)
	_ = binary.Write(&output, binary.BigEndian, c.RootEpoch)
	output.Write(c.Nonce[:])
	if err := writeCanonicalString(&output, c.Origin); err != nil {
		return nil, err
	}
	if err := writeCanonicalString(&output, c.RelyingPartyID); err != nil {
		return nil, err
	}
	_ = binary.Write(&output, binary.BigEndian, c.IssuedAt)
	_ = binary.Write(&output, binary.BigEndian, c.ExpiresAt)
	return output.Bytes(), nil
}

func (c CanonicalFactorChallenge) Digest() ([32]byte, error) {
	canonical, err := c.CanonicalBytes()
	if err != nil {
		return [32]byte{}, err
	}
	return sha256.Sum256(canonical), nil
}

// ChallengeReplayConsumer atomically records an accepted proof or rejects a replay.
type ChallengeReplayConsumer interface {
	ConsumeVerifiedChallenge(challengeDigest [32]byte, proofDigest [32]byte, expiresAt int64) error
}

func (c CanonicalFactorChallenge) ValidateAndConsume(
	chainID string,
	accountAddress string,
	action string,
	now int64,
	proofDigest [32]byte,
	replayConsumer ChallengeReplayConsumer,
) error {
	if err := c.Validate(); err != nil {
		return err
	}
	if c.ChainID != chainID {
		return fmt.Errorf("challenge chain mismatch")
	}
	if c.AccountAddress != accountAddress {
		return fmt.Errorf("challenge account mismatch")
	}
	if c.Action != action {
		return fmt.Errorf("challenge action mismatch")
	}
	if now < c.IssuedAt || now >= c.ExpiresAt {
		return fmt.Errorf("challenge is outside its validity window")
	}
	if replayConsumer == nil {
		return fmt.Errorf("durable replay consumer is required")
	}
	if proofDigest == ([32]byte{}) {
		return fmt.Errorf("accepted proof digest is required before replay consumption")
	}
	digest, err := c.Digest()
	if err != nil {
		return err
	}
	return replayConsumer.ConsumeVerifiedChallenge(digest, proofDigest, c.ExpiresAt)
}

func writeCanonicalString(output *bytes.Buffer, value string) error {
	if len(value) > int(^uint32(0)) {
		return fmt.Errorf("canonical string exceeds uint32 length")
	}
	if err := binary.Write(output, binary.BigEndian, uint32(len(value))); err != nil {
		return err
	}
	_, err := output.WriteString(value)
	return err
}
