package cmd

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/spf13/cobra"
)

// ModuleAddressCmd prints the bech32 address for a named module account.
func ModuleAddressCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "module-address [module-name]",
		Short: "Print the bech32 address for a module account",
		Long: `Print the bech32 account address derived from a module name.
Example:
	virtengine debug module-address distribution
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println(authtypes.NewModuleAddress(args[0]).String())
			return nil
		},
	}
}
