package types

import (
	"fmt"
	"strings"
)

type VerifierStatus string

const (
	VerifierStatusProposed   VerifierStatus = "proposed"
	VerifierStatusApproved   VerifierStatus = "approved"
	VerifierStatusActive     VerifierStatus = "active"
	VerifierStatusDeprecated VerifierStatus = "deprecated"
	VerifierStatusRetired    VerifierStatus = "retired"
	VerifierStatusCancelled  VerifierStatus = "cancelled"
)

type VerifierVersion struct {
	VerifierID           string `json:"verifier_id"`
	SpecVersion          string `json:"spec_version"`
	SpecCID              string `json:"spec_cid,omitempty"`
	SpecSHA256           string `json:"spec_sha256,omitempty"`
	WeightsCID           string `json:"weights_cid,omitempty"`
	WeightsSHA256        string `json:"weights_sha256"`
	TestVectorsCID       string `json:"test_vectors_cid,omitempty"`
	TestVectorsSHA256    string `json:"test_vectors_sha256,omitempty"`
	BuildMetadataCID     string `json:"build_metadata_cid,omitempty"`
	BuildMetadataSHA256  string `json:"build_metadata_sha256,omitempty"`
	ImageHash            string `json:"image_hash,omitempty"`
	ModelManifestHash    string `json:"model_manifest_hash,omitempty"`
	ActivationHeight     int64  `json:"activation_height"`
	Status               string `json:"status"`
	SecurityFix          bool   `json:"security_fix"`
	GovernanceProposalID uint64 `json:"governance_proposal_id,omitempty"`
}

func (v VerifierVersion) Validate() error {
	if strings.TrimSpace(v.VerifierID) == "" {
		return fmt.Errorf("verifier_id cannot be empty")
	}
	if strings.TrimSpace(v.SpecVersion) == "" {
		return fmt.Errorf("spec_version cannot be empty")
	}
	if strings.TrimSpace(v.WeightsSHA256) == "" && strings.TrimSpace(v.ImageHash) == "" && strings.TrimSpace(v.ModelManifestHash) == "" {
		return fmt.Errorf("at least one immutable artifact hash is required")
	}
	if !IsValidVerifierStatus(v.Status) {
		return fmt.Errorf("invalid verifier status: %s", v.Status)
	}
	if v.ActivationHeight < 0 {
		return fmt.Errorf("activation_height cannot be negative")
	}
	return nil
}

type ActiveVerifierPointer struct {
	VerifierID        string `json:"verifier_id"`
	SpecVersion       string `json:"spec_version"`
	ActivatedAtHeight int64  `json:"activated_at_height"`
}

func (p ActiveVerifierPointer) Validate() error {
	if strings.TrimSpace(p.VerifierID) == "" {
		return fmt.Errorf("verifier_id cannot be empty")
	}
	if strings.TrimSpace(p.SpecVersion) == "" {
		return fmt.Errorf("spec_version cannot be empty")
	}
	if p.ActivatedAtHeight < 0 {
		return fmt.Errorf("activated_at_height cannot be negative")
	}
	return nil
}

type ValidatorReadiness struct {
	ValidatorAddress  string `json:"validator_address"`
	VerifierID        string `json:"verifier_id"`
	ConformancePassed bool   `json:"conformance_passed"`
	ImplementationID  string `json:"implementation_id,omitempty"`
	Organization      string `json:"organization,omitempty"`
	ReportedHeight    int64  `json:"reported_height"`
}

func (r ValidatorReadiness) Validate() error {
	if strings.TrimSpace(r.ValidatorAddress) == "" {
		return fmt.Errorf("validator_address cannot be empty")
	}
	if strings.TrimSpace(r.VerifierID) == "" {
		return fmt.Errorf("verifier_id cannot be empty")
	}
	if r.ReportedHeight < 0 {
		return fmt.Errorf("reported_height cannot be negative")
	}
	return nil
}

type Params struct {
	MinimumReadyValidators            uint32 `json:"minimum_ready_validators"`
	MinimumIndependentImplementations uint32 `json:"minimum_independent_implementations"`
	AllowLegacyMirroring              bool   `json:"allow_legacy_mirroring"`
}

func DefaultParams() Params {
	return Params{
		MinimumReadyValidators:            1,
		MinimumIndependentImplementations: 1,
		AllowLegacyMirroring:              true,
	}
}

func (p Params) Validate() error {
	if p.MinimumReadyValidators == 0 {
		return fmt.Errorf("minimum_ready_validators must be at least 1")
	}
	if p.MinimumIndependentImplementations == 0 {
		return fmt.Errorf("minimum_independent_implementations must be at least 1")
	}
	return nil
}

type GenesisState struct {
	Params             Params                 `json:"params"`
	Verifiers          []VerifierVersion      `json:"verifiers"`
	ActiveVerifier     *ActiveVerifierPointer `json:"active_verifier,omitempty"`
	ValidatorReadiness []ValidatorReadiness   `json:"validator_readiness"`
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:             DefaultParams(),
		Verifiers:          []VerifierVersion{},
		ValidatorReadiness: []ValidatorReadiness{},
	}
}

func (g GenesisState) Validate() error {
	if err := g.Params.Validate(); err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(g.Verifiers))
	for _, verifier := range g.Verifiers {
		if err := verifier.Validate(); err != nil {
			return err
		}
		if _, exists := seen[verifier.VerifierID]; exists {
			return fmt.Errorf("duplicate verifier_id: %s", verifier.VerifierID)
		}
		seen[verifier.VerifierID] = struct{}{}
	}
	if g.ActiveVerifier != nil {
		if err := g.ActiveVerifier.Validate(); err != nil {
			return err
		}
		if _, exists := seen[g.ActiveVerifier.VerifierID]; !exists {
			return fmt.Errorf("active verifier %s not present in genesis verifier list", g.ActiveVerifier.VerifierID)
		}
	}
	for _, readiness := range g.ValidatorReadiness {
		if err := readiness.Validate(); err != nil {
			return err
		}
		if _, exists := seen[readiness.VerifierID]; !exists {
			return fmt.Errorf("validator readiness references unknown verifier_id: %s", readiness.VerifierID)
		}
	}
	return nil
}

func IsValidVerifierStatus(status string) bool {
	switch VerifierStatus(status) {
	case VerifierStatusProposed, VerifierStatusApproved, VerifierStatusActive, VerifierStatusDeprecated, VerifierStatusRetired, VerifierStatusCancelled:
		return true
	default:
		return false
	}
}

func CanTransitionVerifierStatus(from, to string) bool {
	if from == "" {
		return VerifierStatus(to) == VerifierStatusProposed
	}

	switch VerifierStatus(from) {
	case VerifierStatusProposed:
		return VerifierStatus(to) == VerifierStatusApproved || VerifierStatus(to) == VerifierStatusCancelled
	case VerifierStatusApproved:
		return VerifierStatus(to) == VerifierStatusApproved ||
			VerifierStatus(to) == VerifierStatusActive ||
			VerifierStatus(to) == VerifierStatusCancelled
	case VerifierStatusActive:
		return VerifierStatus(to) == VerifierStatusActive ||
			VerifierStatus(to) == VerifierStatusDeprecated
	case VerifierStatusDeprecated:
		return VerifierStatus(to) == VerifierStatusDeprecated ||
			VerifierStatus(to) == VerifierStatusRetired
	case VerifierStatusRetired, VerifierStatusCancelled:
		return VerifierStatus(to) == VerifierStatus(from)
	default:
		return false
	}
}
