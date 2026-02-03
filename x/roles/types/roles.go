package types

import (
	"fmt"
)

// Role represents the different roles in the VirtEngine system
type Role uint8

const (
	// RoleUnspecified is the default/invalid role
	RoleUnspecified Role = iota

	// RoleGenesisAccount represents the highest privilege role - initial chain authority
	RoleGenesisAccount

	// RoleAdministrator represents platform operations with high trust level
	RoleAdministrator

	// RoleModerator represents content/user moderation with medium-high trust level
	RoleModerator

	// RoleValidator represents consensus participants with high trust level
	RoleValidator

	// RoleServiceProvider represents infrastructure operators with medium trust level
	RoleServiceProvider

	// RoleCustomer represents end users with standard trust level
	RoleCustomer

	// RoleSupportAgent represents customer support with medium trust level
	RoleSupportAgent
)

// String returns the string representation of a role
func (r Role) String() string {
	switch r {
	case RoleUnspecified:
		return "unspecified"
	case RoleGenesisAccount:
		return "genesis_account"
	case RoleAdministrator:
		return "administrator"
	case RoleModerator:
		return "moderator"
	case RoleValidator:
		return "validator"
	case RoleServiceProvider:
		return "service_provider"
	case RoleCustomer:
		return "customer"
	case RoleSupportAgent:
		return "support_agent"
	default:
		return fmt.Sprintf("unknown(%d)", r)
	}
}

// RoleFromString converts a string to a Role
func RoleFromString(s string) (Role, error) {
	switch s {
	case "genesis_account":
		return RoleGenesisAccount, nil
	case "administrator":
		return RoleAdministrator, nil
	case "moderator":
		return RoleModerator, nil
	case "validator":
		return RoleValidator, nil
	case "service_provider":
		return RoleServiceProvider, nil
	case "customer":
		return RoleCustomer, nil
	case "support_agent":
		return RoleSupportAgent, nil
	default:
		return RoleUnspecified, fmt.Errorf("unknown role: %s", s)
	}
}

// IsValid checks if the role is a valid role
func (r Role) IsValid() bool {
	return r >= RoleGenesisAccount && r <= RoleSupportAgent
}

// TrustLevel returns the trust level of a role (higher = more trusted)
func (r Role) TrustLevel() int {
	switch r {
	case RoleGenesisAccount:
		return 100
	case RoleAdministrator:
		return 90
	case RoleValidator:
		return 85
	case RoleModerator:
		return 70
	case RoleServiceProvider:
		return 60
	case RoleSupportAgent:
		return 60
	case RoleCustomer:
		return 50
	default:
		return 0
	}
}

// CanAssignRole checks if this role can assign another role
func (r Role) CanAssignRole(target Role) bool {
	switch r {
	case RoleGenesisAccount:
		// GenesisAccount can assign any role
		return target.IsValid()
	case RoleAdministrator:
		// Administrators can assign roles below their level (except GenesisAccount and Administrator)
		return target.IsValid() && target != RoleGenesisAccount && target != RoleAdministrator
	default:
		return false
	}
}

// CanRevokeRole checks if this role can revoke another role
func (r Role) CanRevokeRole(target Role) bool {
	// Same permissions as assign
	return r.CanAssignRole(target)
}

// CanModifyAccountState checks if this role can modify account states
func (r Role) CanModifyAccountState() bool {
	switch r {
	case RoleGenesisAccount, RoleAdministrator:
		return true
	default:
		return false
	}
}

// AllRoles returns all valid roles
func AllRoles() []Role {
	return []Role{
		RoleGenesisAccount,
		RoleAdministrator,
		RoleModerator,
		RoleValidator,
		RoleServiceProvider,
		RoleCustomer,
		RoleSupportAgent,
	}
}

// RoleAssignment represents a role assigned to an account
type RoleAssignment struct {
	Address    string `json:"address"`
	Role       Role   `json:"role"`
	AssignedBy string `json:"assigned_by"`
	AssignedAt int64  `json:"assigned_at"`
}

// Validate validates the role assignment
func (ra RoleAssignment) Validate() error {
	if ra.Address == "" {
		return ErrInvalidAddress
	}
	if !ra.Role.IsValid() {
		return ErrInvalidRole
	}
	return nil
}
