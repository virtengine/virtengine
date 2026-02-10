package types

// GenesisState is the genesis state for the config module
type GenesisState struct {
	// ApprovedClients are the initially approved clients
	ApprovedClients []ApprovedClient `json:"approved_clients"`

	// Params are the module parameters
	Params Params `json:"params"`
}

// Params defines the parameters for the config module
type Params struct {
	// RequireClientSignature determines if client signatures are mandatory
	RequireClientSignature bool `json:"require_client_signature"`

	// RequireUserSignature determines if user signatures are mandatory
	RequireUserSignature bool `json:"require_user_signature"`

	// RequireSaltBinding determines if salt binding validation is mandatory
	RequireSaltBinding bool `json:"require_salt_binding"`

	// MaxClientsPerRegistrar is the maximum number of clients one account can register
	MaxClientsPerRegistrar uint32 `json:"max_clients_per_registrar"`

	// AllowGovernanceOverride allows governance to override admin decisions
	AllowGovernanceOverride bool `json:"allow_governance_override"`

	// DefaultMinVersion is the default minimum version for new clients
	DefaultMinVersion string `json:"default_min_version"`

	// AdminAddresses are addresses that can manage approved clients
	AdminAddresses []string `json:"admin_addresses"`
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		ApprovedClients: []ApprovedClient{},
		Params:          DefaultParams(),
	}
}

// DefaultParams returns the default parameters
func DefaultParams() Params {
	return Params{
		RequireClientSignature:  true,
		RequireUserSignature:    true,
		RequireSaltBinding:      true,
		MaxClientsPerRegistrar:  10,
		AllowGovernanceOverride: true,
		DefaultMinVersion:       "1.0.0",
		AdminAddresses:          []string{},
	}
}

// Validate validates the genesis state
func (gs GenesisState) Validate() error {
	// Validate params
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	// Validate approved clients
	seenClients := make(map[string]bool)
	for _, client := range gs.ApprovedClients {
		if err := client.Validate(); err != nil {
			return err
		}
		if seenClients[client.ClientID] {
			return ErrClientAlreadyExists.Wrapf("duplicate client: %s", client.ClientID)
		}
		seenClients[client.ClientID] = true
	}

	return nil
}

// Validate validates the parameters
func (p Params) Validate() error {
	if p.MaxClientsPerRegistrar == 0 {
		return ErrInvalidClientID.Wrap("max_clients_per_registrar must be greater than 0")
	}

	if p.DefaultMinVersion != "" && !isValidSemver(p.DefaultMinVersion) {
		return ErrInvalidVersionConstraint.Wrapf("invalid default_min_version: %s", p.DefaultMinVersion)
	}

	return nil
}
