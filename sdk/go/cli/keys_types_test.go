package cli_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/sdk/go/cli"
)

func TestConstructors(t *testing.T) {
	require.Equal(t, cli.AddNewKey{
		Name:     "name",
		Password: "password",
		Mnemonic: "mnemonic",
		Account:  1,
		Index:    1,
	}, cli.NewAddNewKey("name", "password", "mnemonic", 1, 1))

	require.Equal(t, cli.RecoverKey{
		Password: "password",
		Mnemonic: "mnemonic",
		Account:  1,
		Index:    1,
	}, cli.NewRecoverKey("password", "mnemonic", 1, 1))

	require.Equal(t, cli.UpdateKeyReq{OldPassword: "old", NewPassword: "new"}, cli.NewUpdateKeyReq("old", "new"))
	require.Equal(t, cli.DeleteKeyReq{Password: "password"}, cli.NewDeleteKeyReq("password"))
}

