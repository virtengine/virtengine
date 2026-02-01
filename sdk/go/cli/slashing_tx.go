package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
)

// GetTxSlashingCmd returns a root CLI command handler for all x/slashing transaction commands.
func GetTxSlashingCmd() *cobra.Command {
	slashingTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Slashing transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	slashingTxCmd.AddCommand(
		GetTxSlashingUnjailCmd(),
	)

	return slashingTxCmd
}

// GetTxSlashingUnjailCmd returns a CLI command handler for creating a MsgUnjail transaction.
func GetTxSlashingUnjailCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unjail",
		Args:  cobra.NoArgs,
		Short: "unjail validator previously jailed for downtime",
		Long: `unjail a jailed validator:

$ <appd> tx slashing unjail --from mykey
`,
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			operatorAddr := cctx.GetFromAddress()

			msg := types.NewMsgUnjail(sdk.ValAddress(operatorAddr).String())

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)

	return cmd
}

