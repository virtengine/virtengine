package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/spf13/cobra"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
)

func GetQueryModuleNameToAddressCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "module-name-to-address [module-name]",
		Short: "module name to address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			address := authtypes.NewModuleAddress(args[0])
			return clientCtx.PrintString(address.String())
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

