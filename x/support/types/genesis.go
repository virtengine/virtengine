package types

import "fmt"

// GenesisState is the genesis state for the support module
type GenesisState struct {
	// ExternalRefs are the initial external ticket references
	ExternalRefs []ExternalTicketRef `json:"external_refs"`

	// Params are the module parameters
	Params Params `json:"params"`
}

// Params defines the parameters for the support module
type Params struct {
	// AllowedExternalSystems is the list of allowed external systems
	AllowedExternalSystems []string `json:"allowed_external_systems"`

	// AllowedExternalDomains is the list of allowed domains for external URLs
	// Used for validation to prevent phishing
	AllowedExternalDomains []string `json:"allowed_external_domains"`
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		ExternalRefs: []ExternalTicketRef{},
		Params:       DefaultParams(),
	}
}

// DefaultParams returns the default parameters
func DefaultParams() Params {
	return Params{
		AllowedExternalSystems: []string{
			string(ExternalSystemWaldur),
			string(ExternalSystemJira),
		},
		AllowedExternalDomains: []string{}, // Empty = allow all (configure in production)
	}
}

// Validate validates the genesis state
func (gs GenesisState) Validate() error {
	// Validate external refs
	seenRefs := make(map[string]bool)
	for _, ref := range gs.ExternalRefs {
		if err := ref.Validate(); err != nil {
			return err
		}
		key := ref.Key()
		if seenRefs[key] {
			return ErrRefAlreadyExists.Wrapf("duplicate external ref: %s", key)
		}
		seenRefs[key] = true
	}

	// Validate params
	return gs.Params.Validate()
}

// Validate validates the params
func (p Params) Validate() error {
	if len(p.AllowedExternalSystems) == 0 {
		return ErrInvalidParams.Wrap("allowed_external_systems cannot be empty")
	}

	// Validate each system is known
	for _, sys := range p.AllowedExternalSystems {
		if !ExternalSystem(sys).IsValid() {
			return ErrInvalidParams.Wrapf("unknown external system: %s", sys)
		}
	}

	return nil
}

// IsSystemAllowed checks if an external system is allowed
func (p Params) IsSystemAllowed(system ExternalSystem) bool {
	for _, s := range p.AllowedExternalSystems {
		if s == string(system) {
			return true
		}
	}
	return false
}

// Proto message interface stubs for GenesisState
func (*GenesisState) ProtoMessage() {}
func (gs *GenesisState) Reset()     { *gs = GenesisState{} }
func (gs *GenesisState) String() string {
	return fmt.Sprintf("GenesisState{ExternalRefs: %d}", len(gs.ExternalRefs))
}

// Proto message interface stubs for Params
func (*Params) ProtoMessage() {}
func (p *Params) Reset()      { *p = Params{} }
func (p *Params) String() string {
	return fmt.Sprintf("Params{AllowedSystems: %v}", p.AllowedExternalSystems)
}
