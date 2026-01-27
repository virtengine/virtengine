package cli

import (
	"fmt"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	eid "github.com/virtengine/virtengine/sdk/go/node/escrow/id/v1"
	emodule "github.com/virtengine/virtengine/sdk/go/node/escrow/module"
	ev1 "github.com/virtengine/virtengine/sdk/go/node/escrow/v1"
	deposit "github.com/virtengine/virtengine/sdk/go/node/types/deposit/v1"
)

func GetTxEscrowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        emodule.ModuleName,
		Short:                      "Escrow transaction commands",
		SuggestionsMinimumDistance: 2,
		RunE:                       sdkclient.ValidateCmd,
	}
	cmd.AddCommand(
		GetTxEscrowDeposit(),
	)

	return cmd
}

func GetTxEscrowDeposit() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "deposit [deployment] [amount]",
		Short:             "deposit funds to escrow account",
		Args:              cobra.ExactArgs(2),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			var aid eid.Account

			switch args[0] {
			case "deployment":
				id, err := cflags.DeploymentIDFromFlags(cmd.Flags(), cflags.WithOwner(cctx.FromAddress))
				if err != nil {
					return err
				}
				aid = id.ToEscrowAccountID()
			default:
				return fmt.Errorf("invalid account scope. allowed values deployment")
			}

			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			sources, err := DepositSources(cmd.Flags())
			if err != nil {
				return err
			}

			msg := &ev1.MsgAccountDeposit{
				ID:     aid,
				Signer: cctx.FromAddress.String(),
				Deposit: deposit.Deposit{
					Amount:  amount,
					Sources: sources,
				},
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)
	cflags.AddDeploymentIDFlags(cmd.Flags())
	cflags.AddDepositSourcesFlags(cmd.Flags())

	return cmd
}
