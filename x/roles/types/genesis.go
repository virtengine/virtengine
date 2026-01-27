package types

import "fmt"

// GenesisState is the genesis state for the roles module
type GenesisState struct {
	// GenesisAccounts are the accounts with GenesisAccount role
	GenesisAccounts []string `json:"genesis_accounts"`

	// RoleAssignments are the initial role assignments
	RoleAssignments []RoleAssignment `json:"role_assignments"`

	// AccountStates are the initial account states
	AccountStates []AccountStateRecord `json:"account_states"`

	// Params are the module parameters
	Params Params `json:"params"`
}

// Params defines the parameters for the roles module
type Params struct {
	// MaxRolesPerAccount is the maximum number of roles an account can have
	MaxRolesPerAccount uint32 `json:"max_roles_per_account"`

	// AllowSelfRevoke determines if accounts can revoke their own roles
	AllowSelfRevoke bool `json:"allow_self_revoke"`
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		GenesisAccounts: []string{},
		RoleAssignments: []RoleAssignment{},
		AccountStates:   []AccountStateRecord{},
		Params:          DefaultParams(),
	}
}

// DefaultParams returns the default parameters
func DefaultParams() Params {
	return Params{
		MaxRolesPerAccount: 5,
		AllowSelfRevoke:    false,
	}
}

// Validate validates the genesis state
func (gs GenesisState) Validate() error {
	// Validate genesis accounts
	seen := make(map[string]bool)
	for _, addr := range gs.GenesisAccounts {
		if seen[addr] {
			return ErrInvalidAddress.Wrapf("duplicate genesis account: %s", addr)
		}
		seen[addr] = true
	}

	// Validate role assignments
	for _, ra := range gs.RoleAssignments {
		if err := ra.Validate(); err != nil {
			return err
		}
	}

	// Validate account states
	for _, as := range gs.AccountStates {
		if err := as.Validate(); err != nil {
			return err
		}
	}

	// Validate params
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	return nil
}

// Validate validates the params
func (p Params) Validate() error {
	if p.MaxRolesPerAccount == 0 {
		return ErrInvalidRole.Wrap("max_roles_per_account must be greater than 0")
	}
	return nil
}

// ProtoMessage implements proto.Message
func (*GenesisState) ProtoMessage() {}

// Reset implements proto.Message
func (gs *GenesisState) Reset() { *gs = GenesisState{} }

// String implements proto.Message
func (gs *GenesisState) String() string {
	return fmt.Sprintf("%+v", *gs)
}
