package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/roles/types"
)

func TestDefaultGenesisState(t *testing.T) {
	gs := types.DefaultGenesisState()
	require.NotNil(t, gs)
	require.Empty(t, gs.GenesisAccounts)
	require.Empty(t, gs.RoleAssignments)
	require.Empty(t, gs.AccountStates)
	require.Equal(t, types.DefaultParams(), gs.Params)
}

func TestGenesisStateValidate(t *testing.T) {
	testCases := []struct {
		name     string
		gs       types.GenesisState
		hasError bool
	}{
		{
			name:     "valid default genesis",
			gs:       *types.DefaultGenesisState(),
			hasError: false,
		},
		{
			name: "duplicate genesis accounts",
			gs: types.GenesisState{
				GenesisAccounts: []string{"addr1", "addr1"},
				Params:          types.DefaultParams(),
			},
			hasError: true,
		},
		{
			name: "invalid role assignment",
			gs: types.GenesisState{
				RoleAssignments: []types.RoleAssignment{
					{Address: "", Role: types.RoleCustomer},
				},
				Params: types.DefaultParams(),
			},
			hasError: true,
		},
		{
			name: "invalid account state",
			gs: types.GenesisState{
				AccountStates: []types.AccountStateRecord{
					{Address: "", State: types.AccountStateActive},
				},
				Params: types.DefaultParams(),
			},
			hasError: true,
		},
		{
			name: "invalid params",
			gs: types.GenesisState{
				Params: types.Params{MaxRolesPerAccount: 0},
			},
			hasError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.gs.Validate()
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDefaultParams(t *testing.T) {
	params := types.DefaultParams()
	require.Equal(t, uint32(5), params.MaxRolesPerAccount)
	require.False(t, params.AllowSelfRevoke)
}

func TestParamsValidate(t *testing.T) {
	testCases := []struct {
		name     string
		params   types.Params
		hasError bool
	}{
		{
			name:     "valid default params",
			params:   types.DefaultParams(),
			hasError: false,
		},
		{
			name: "valid custom params",
			params: types.Params{
				MaxRolesPerAccount: 10,
				AllowSelfRevoke:    true,
			},
			hasError: false,
		},
		{
			name: "invalid max roles",
			params: types.Params{
				MaxRolesPerAccount: 0,
				AllowSelfRevoke:    false,
			},
			hasError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.params.Validate()
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRoleAssignmentValidate(t *testing.T) {
	testCases := []struct {
		name     string
		ra       types.RoleAssignment
		hasError bool
	}{
		{
			name: "valid role assignment",
			ra: types.RoleAssignment{
				Address: "test_address",
				Role:    types.RoleCustomer,
			},
			hasError: false,
		},
		{
			name: "empty address",
			ra: types.RoleAssignment{
				Address: "",
				Role:    types.RoleCustomer,
			},
			hasError: true,
		},
		{
			name: "invalid role",
			ra: types.RoleAssignment{
				Address: "test_address",
				Role:    types.RoleUnspecified,
			},
			hasError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.ra.Validate()
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAccountStateRecordValidate(t *testing.T) {
	testCases := []struct {
		name     string
		record   types.AccountStateRecord
		hasError bool
	}{
		{
			name: "valid record",
			record: types.AccountStateRecord{
				Address: "test_address",
				State:   types.AccountStateActive,
			},
			hasError: false,
		},
		{
			name: "empty address",
			record: types.AccountStateRecord{
				Address: "",
				State:   types.AccountStateActive,
			},
			hasError: true,
		},
		{
			name: "invalid state",
			record: types.AccountStateRecord{
				Address: "test_address",
				State:   types.AccountStateUnspecified,
			},
			hasError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.record.Validate()
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
