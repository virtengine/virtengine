package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"pkg.akt.dev/node/x/roles/types"
)

func TestRoleIsValid(t *testing.T) {
	testCases := []struct {
		role     types.Role
		expected bool
	}{
		{types.RoleUnspecified, false},
		{types.RoleGenesisAccount, true},
		{types.RoleAdministrator, true},
		{types.RoleModerator, true},
		{types.RoleValidator, true},
		{types.RoleServiceProvider, true},
		{types.RoleCustomer, true},
		{types.RoleSupportAgent, true},
		{types.Role(99), false},
	}

	for _, tc := range testCases {
		t.Run(tc.role.String(), func(t *testing.T) {
			require.Equal(t, tc.expected, tc.role.IsValid())
		})
	}
}

func TestRoleFromString(t *testing.T) {
	testCases := []struct {
		input    string
		expected types.Role
		hasError bool
	}{
		{"genesis_account", types.RoleGenesisAccount, false},
		{"administrator", types.RoleAdministrator, false},
		{"moderator", types.RoleModerator, false},
		{"validator", types.RoleValidator, false},
		{"service_provider", types.RoleServiceProvider, false},
		{"customer", types.RoleCustomer, false},
		{"support_agent", types.RoleSupportAgent, false},
		{"unknown", types.RoleUnspecified, true},
		{"", types.RoleUnspecified, true},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			role, err := types.RoleFromString(tc.input)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, role)
			}
		})
	}
}

func TestRoleString(t *testing.T) {
	testCases := []struct {
		role     types.Role
		expected string
	}{
		{types.RoleUnspecified, "unspecified"},
		{types.RoleGenesisAccount, "genesis_account"},
		{types.RoleAdministrator, "administrator"},
		{types.RoleModerator, "moderator"},
		{types.RoleValidator, "validator"},
		{types.RoleServiceProvider, "service_provider"},
		{types.RoleCustomer, "customer"},
		{types.RoleSupportAgent, "support_agent"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.role.String())
		})
	}
}

func TestRoleTrustLevel(t *testing.T) {
	// GenesisAccount should have highest trust
	require.Equal(t, 100, types.RoleGenesisAccount.TrustLevel())
	require.Equal(t, 90, types.RoleAdministrator.TrustLevel())
	require.Equal(t, 85, types.RoleValidator.TrustLevel())
	require.Equal(t, 70, types.RoleModerator.TrustLevel())
	require.Equal(t, 60, types.RoleServiceProvider.TrustLevel())
	require.Equal(t, 50, types.RoleCustomer.TrustLevel())

	// Higher trust level should be numerically higher
	require.Greater(t, types.RoleGenesisAccount.TrustLevel(), types.RoleAdministrator.TrustLevel())
	require.Greater(t, types.RoleAdministrator.TrustLevel(), types.RoleCustomer.TrustLevel())
}

func TestRoleCanAssignRole(t *testing.T) {
	// GenesisAccount can assign any role
	require.True(t, types.RoleGenesisAccount.CanAssignRole(types.RoleAdministrator))
	require.True(t, types.RoleGenesisAccount.CanAssignRole(types.RoleGenesisAccount))
	require.True(t, types.RoleGenesisAccount.CanAssignRole(types.RoleCustomer))

	// Administrator can assign roles below their level
	require.True(t, types.RoleAdministrator.CanAssignRole(types.RoleCustomer))
	require.True(t, types.RoleAdministrator.CanAssignRole(types.RoleModerator))
	require.True(t, types.RoleAdministrator.CanAssignRole(types.RoleServiceProvider))

	// Administrator cannot assign admin or genesis roles
	require.False(t, types.RoleAdministrator.CanAssignRole(types.RoleAdministrator))
	require.False(t, types.RoleAdministrator.CanAssignRole(types.RoleGenesisAccount))

	// Other roles cannot assign
	require.False(t, types.RoleCustomer.CanAssignRole(types.RoleCustomer))
	require.False(t, types.RoleModerator.CanAssignRole(types.RoleCustomer))
}

func TestAllRoles(t *testing.T) {
	roles := types.AllRoles()
	require.Len(t, roles, 7)

	// Verify all roles are valid
	for _, role := range roles {
		require.True(t, role.IsValid())
	}
}
