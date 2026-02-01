package cli

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/crisis/types"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
)

// GetTxCrisisCmd returns a root CLI command handler for all x/crisis transaction commands.
func GetTxCrisisCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Crisis transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		GetTxCrisisVerifyInvariantTxCmd(),
	)

	return txCmd
}

// GetTxCrisisVerifyInvariantTxCmd returns a CLI command handler for creating a
// MsgVerifyInvariant transaction.
func GetTxCrisisVerifyInvariantTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "invariant-broken [module-name] [invariant-route] --from mykey",
		Short:             "Submit proof that an invariant broken to halt the chain",
		Args:              cobra.ExactArgs(2),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			moduleName, route := args[0], args[1]
			if moduleName == "" {
				return errors.New("invalid module name")
			}
			if route == "" {
				return errors.New("invalid invariant route")
			}

			senderAddr := cctx.GetFromAddress()

			msg := types.NewMsgVerifyInvariant(senderAddr, moduleName, route)

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

