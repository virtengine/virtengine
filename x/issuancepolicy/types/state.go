package types

import (
	"fmt"
	"strings"
)

type PolicyStatus string

const (
	PolicyStatusActive     PolicyStatus = "active"
	PolicyStatusPaused     PolicyStatus = "paused"
	PolicyStatusDeprecated PolicyStatus = "deprecated"
)

type ProofMintStatus string

const (
	ProofMintStatusRecorded         ProofMintStatus = "recorded"
	ProofMintStatusPaused           ProofMintStatus = "paused"
	ProofMintStatusCapExceeded      ProofMintStatus = "cap_exceeded"
	ProofMintStatusVerifierMismatch ProofMintStatus = "verifier_scope_mismatch"
	ProofMintStatusDuplicate        ProofMintStatus = "duplicate"
	ProofMintStatusNoActivePolicy   ProofMintStatus = "no_active_policy"
)

type IssuancePolicy struct {
	PolicyID             string `json:"policy_id"`
	Status               string `json:"status"`
	ActiveVerifierScope  string `json:"active_verifier_scope"`
	MintUnitsPerProof    uint64 `json:"mint_units_per_proof"`
	DailyCap             uint64 `json:"daily_cap"`
	EpochCap             uint64 `json:"epoch_cap"`
	MintingPaused        bool   `json:"minting_paused"`
	CreatedAtHeight      int64  `json:"created_at_height"`
	GovernanceProposalID uint64 `json:"governance_proposal_id,omitempty"`
}

func DefaultIssuancePolicy() IssuancePolicy {
	return IssuancePolicy{
		PolicyID:            "default",
		Status:              string(PolicyStatusActive),
		ActiveVerifierScope: "*",
		MintUnitsPerProof:   0,
		DailyCap:            0,
		EpochCap:            0,
		MintingPaused:       false,
	}
}

func (p IssuancePolicy) Validate() error {
	if strings.TrimSpace(p.PolicyID) == "" {
		return fmt.Errorf("policy_id cannot be empty")
	}
	if !IsValidPolicyStatus(p.Status) {
		return fmt.Errorf("invalid policy status: %s", p.Status)
	}
	if p.ActiveVerifierScope == "" {
		return fmt.Errorf("active_verifier_scope cannot be empty")
	}
	if p.CreatedAtHeight < 0 {
		return fmt.Errorf("created_at_height cannot be negative")
	}
	return nil
}

type IssuanceCounters struct {
	DayIndex        uint64 `json:"day_index"`
	MintedToday     uint64 `json:"minted_today"`
	EpochIndex      uint64 `json:"epoch_index"`
	MintedThisEpoch uint64 `json:"minted_this_epoch"`
}

type ProofMintRecord struct {
	ProofID        string `json:"proof_id"`
	AccountAddress string `json:"account_address"`
	VerifierID     string `json:"verifier_id"`
	ModelVersion   string `json:"model_version,omitempty"`
	MintedUnits    uint64 `json:"minted_units"`
	Height         int64  `json:"height"`
	PolicyID       string `json:"policy_id,omitempty"`
	Status         string `json:"status"`
}

func (r ProofMintRecord) Validate() error {
	if strings.TrimSpace(r.ProofID) == "" {
		return fmt.Errorf("proof_id cannot be empty")
	}
	if strings.TrimSpace(r.AccountAddress) == "" {
		return fmt.Errorf("account_address cannot be empty")
	}
	if !IsValidProofMintStatus(r.Status) {
		return fmt.Errorf("invalid proof mint status: %s", r.Status)
	}
	return nil
}

type Params struct {
	EpochLengthBlocks     int64  `json:"epoch_length_blocks"`
	MaxMintUnitsPerProof  uint64 `json:"max_mint_units_per_proof"`
	MaxDailyCap           uint64 `json:"max_daily_cap"`
	MaxEpochCap           uint64 `json:"max_epoch_cap"`
	EmergencyPauseEnabled bool   `json:"emergency_pause_enabled"`
}

func DefaultParams() Params {
	return Params{
		EpochLengthBlocks:     10000,
		MaxMintUnitsPerProof:  1_000_000_000,
		MaxDailyCap:           1_000_000_000,
		MaxEpochCap:           1_000_000_000,
		EmergencyPauseEnabled: true,
	}
}

func (p Params) Validate() error {
	if p.EpochLengthBlocks <= 0 {
		return fmt.Errorf("epoch_length_blocks must be positive")
	}
	return nil
}

type GenesisState struct {
	Params         Params            `json:"params"`
	Policies       []IssuancePolicy  `json:"policies"`
	ActivePolicyID string            `json:"active_policy_id,omitempty"`
	Counters       IssuanceCounters  `json:"counters"`
	ProofRecords   []ProofMintRecord `json:"proof_records"`
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:       DefaultParams(),
		Policies:     []IssuancePolicy{},
		ProofRecords: []ProofMintRecord{},
	}
}

func (g GenesisState) Validate() error {
	if err := g.Params.Validate(); err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(g.Policies))
	for _, policy := range g.Policies {
		if err := policy.Validate(); err != nil {
			return err
		}
		if _, exists := seen[policy.PolicyID]; exists {
			return fmt.Errorf("duplicate policy_id: %s", policy.PolicyID)
		}
		seen[policy.PolicyID] = struct{}{}
	}
	if g.ActivePolicyID != "" {
		if _, exists := seen[g.ActivePolicyID]; !exists {
			return fmt.Errorf("active policy %s not found in genesis policies", g.ActivePolicyID)
		}
	}
	for _, record := range g.ProofRecords {
		if err := record.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func IsValidPolicyStatus(status string) bool {
	switch PolicyStatus(status) {
	case PolicyStatusActive, PolicyStatusPaused, PolicyStatusDeprecated:
		return true
	default:
		return false
	}
}

func IsValidProofMintStatus(status string) bool {
	switch ProofMintStatus(status) {
	case ProofMintStatusRecorded, ProofMintStatusPaused, ProofMintStatusCapExceeded, ProofMintStatusVerifierMismatch, ProofMintStatusDuplicate, ProofMintStatusNoActivePolicy:
		return true
	default:
		return false
	}
}

func CanTransitionPolicyStatus(from, to string) bool {
	if from == "" {
		return PolicyStatus(to) == PolicyStatusPaused || PolicyStatus(to) == PolicyStatusActive
	}

	switch PolicyStatus(from) {
	case PolicyStatusActive:
		return PolicyStatus(to) == PolicyStatusActive ||
			PolicyStatus(to) == PolicyStatusPaused ||
			PolicyStatus(to) == PolicyStatusDeprecated
	case PolicyStatusPaused:
		return PolicyStatus(to) == PolicyStatusPaused ||
			PolicyStatus(to) == PolicyStatusActive ||
			PolicyStatus(to) == PolicyStatusDeprecated
	case PolicyStatusDeprecated:
		return PolicyStatus(to) == PolicyStatusDeprecated
	default:
		return false
	}
}
