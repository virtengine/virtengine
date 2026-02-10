package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/roles/types"
)

func TestMsgAssignRoleValidateBasic(t *testing.T) {
	validAddr := sdk.AccAddress([]byte("valid_address_123456")).String()

	testCases := []struct {
		name     string
		msg      types.MsgAssignRole
		hasError bool
	}{
		{
			name: "valid message",
			msg: types.MsgAssignRole{
				Sender:  validAddr,
				Address: validAddr,
				Role:    "customer",
			},
			hasError: false,
		},
		{
			name: "invalid sender",
			msg: types.MsgAssignRole{
				Sender:  "invalid",
				Address: validAddr,
				Role:    "customer",
			},
			hasError: true,
		},
		{
			name: "invalid target",
			msg: types.MsgAssignRole{
				Sender:  validAddr,
				Address: "invalid",
				Role:    "customer",
			},
			hasError: true,
		},
		{
			name: "invalid role",
			msg: types.MsgAssignRole{
				Sender:  validAddr,
				Address: validAddr,
				Role:    "unknown",
			},
			hasError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgRevokeRoleValidateBasic(t *testing.T) {
	validAddr := sdk.AccAddress([]byte("valid_address_123456")).String()

	testCases := []struct {
		name     string
		msg      types.MsgRevokeRole
		hasError bool
	}{
		{
			name: "valid message",
			msg: types.MsgRevokeRole{
				Sender:  validAddr,
				Address: validAddr,
				Role:    "customer",
			},
			hasError: false,
		},
		{
			name: "invalid sender",
			msg: types.MsgRevokeRole{
				Sender:  "invalid",
				Address: validAddr,
				Role:    "customer",
			},
			hasError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgSetAccountStateValidateBasic(t *testing.T) {
	validAddr := sdk.AccAddress([]byte("valid_address_123456")).String()

	testCases := []struct {
		name     string
		msg      types.MsgSetAccountState
		hasError bool
	}{
		{
			name: "valid message",
			msg: types.MsgSetAccountState{
				Sender:  validAddr,
				Address: validAddr,
				State:   "suspended",
				Reason:  "test reason",
			},
			hasError: false,
		},
		{
			name: "invalid state",
			msg: types.MsgSetAccountState{
				Sender:  validAddr,
				Address: validAddr,
				State:   "unknown",
				Reason:  "test reason",
			},
			hasError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgNominateAdminValidateBasic(t *testing.T) {
	validAddr := sdk.AccAddress([]byte("valid_address_123456")).String()

	testCases := []struct {
		name     string
		msg      types.MsgNominateAdmin
		hasError bool
	}{
		{
			name: "valid message",
			msg: types.MsgNominateAdmin{
				Sender:  validAddr,
				Address: validAddr,
			},
			hasError: false,
		},
		{
			name: "invalid sender",
			msg: types.MsgNominateAdmin{
				Sender:  "invalid",
				Address: validAddr,
			},
			hasError: true,
		},
		{
			name: "invalid target",
			msg: types.MsgNominateAdmin{
				Sender:  validAddr,
				Address: "invalid",
			},
			hasError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewMsgAssignRole(t *testing.T) {
	msg := types.NewMsgAssignRole("sender", "address", "customer")
	require.Equal(t, "sender", msg.Sender)
	require.Equal(t, "address", msg.Address)
	require.Equal(t, "customer", msg.Role)
	require.Equal(t, types.RouterKey, msg.Route())
	require.Equal(t, types.TypeMsgAssignRole, msg.Type())
}

func TestNewMsgRevokeRole(t *testing.T) {
	msg := types.NewMsgRevokeRole("sender", "address", "customer")
	require.Equal(t, "sender", msg.Sender)
	require.Equal(t, "address", msg.Address)
	require.Equal(t, "customer", msg.Role)
	require.Equal(t, types.RouterKey, msg.Route())
	require.Equal(t, types.TypeMsgRevokeRole, msg.Type())
}

func TestNewMsgSetAccountState(t *testing.T) {
	msg := types.NewMsgSetAccountState("sender", "address", "suspended", "test reason")
	require.Equal(t, "sender", msg.Sender)
	require.Equal(t, "address", msg.Address)
	require.Equal(t, "suspended", msg.State)
	require.Equal(t, "test reason", msg.Reason)
}

func TestNewMsgNominateAdmin(t *testing.T) {
	msg := types.NewMsgNominateAdmin("sender", "address")
	require.Equal(t, "sender", msg.Sender)
	require.Equal(t, "address", msg.Address)
}
