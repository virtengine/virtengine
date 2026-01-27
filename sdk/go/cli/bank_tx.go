package cli

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/spf13/cobra"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
)

// GetTxBankCmd returns a root CLI command handler for all x/bank transaction commands.
func GetTxBankCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Bank transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetTxBankSendTxCmd(),
		GetTxBankMultiSendTxCmd(),
	)

	return cmd
}

// GetTxBankSendTxCmd returns a CLI command handler for creating a MsgSend transaction.
func GetTxBankSendTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send <from_key_or_address> [to_address] [amount]",
		Short: "Send funds from one account to another.",
		Long: `Send funds from one account to another.
Note, this command has two way of being executed:
     - sender address|key is specified as a first argument, and command takes 3 arguments. in this case the '--from' flag is ignored as it is implied from [from_key_or_address]
       send [from_key_or_address] [to_address] [amount]
     - sender address|key is taken from --from flag. In this case command takes 2 arguments.
       send [to_address] [amount] --from=address|key"
When using '--dry-run' a key name cannot be used, only a bech32 address.
`,
		Args: cobra.RangeArgs(2, 3),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 3 {
				if err := cmd.Flags().Set(cflags.FlagFrom, args[0]); err != nil {
					return err
				}
			}

			return TxPersistentPreRunE(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			ac := MustAddressCodecFromContext(ctx)
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			if len(args) == 3 {
				args = args[1:]
			}
			toAddr, err := ac.StringToBytes(args[0])
			if err != nil {
				return err
			}

			coins, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}

			if len(coins) == 0 {
				return fmt.Errorf("invalid coins")
			}

			msg := types.NewMsgSend(cctx.GetFromAddress(), toAddr, coins)

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

// GetTxBankMultiSendTxCmd returns a CLI command handler for creating a MsgMultiSend transaction.
// For a better UX this command is limited to send funds from one account to two or more accounts.
func GetTxBankMultiSendTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multi-send [from_key_or_address] [to_address_1 to_address_2 ...] [amount]",
		Short: "Send funds from one account to two or more accounts.",
		Long: `Send funds from one account to two or more accounts.
By default, sends the [amount] to each address of the list.
Using the '--split' flag, the [amount] is split equally between the addresses.
Note, the '--from' flag is ignored as it is implied from [from_key_or_address] and
separate addresses with space.
When using '--dry-run' a key name cannot be used, only a bech32 address.`,
		Example: fmt.Sprintf("%s tx bank multi-send akash1... akash1... akash1... akash1... 10stake", version.AppName),
		Args:    cobra.MinimumNArgs(4),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(cflags.FlagFrom, args[0]); err != nil {
				return err
			}

			return TxPersistentPreRunE(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			ac := MustAddressCodecFromContext(ctx)
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			coins, err := sdk.ParseCoinsNormalized(args[len(args)-1])
			if err != nil {
				return err
			}

			if coins.IsZero() {
				return fmt.Errorf("must send positive amount")
			}

			split, err := cmd.Flags().GetBool(cflags.FlagSplit)
			if err != nil {
				return err
			}

			totalAddrs := sdkmath.NewInt(int64(len(args) - 2))
			// coins to be received by the addresses
			sendCoins := coins
			if split {
				sendCoins = coins.QuoInt(totalAddrs)
			}

			var output []types.Output
			for _, arg := range args[1 : len(args)-1] {
				toAddr, err := ac.StringToBytes(arg)
				if err != nil {
					return err
				}

				output = append(output, types.NewOutput(toAddr, sendCoins))
			}

			// amount to be send from the from address
			var amount sdk.Coins
			if split {
				// user input: 1000uve to send to 3 addresses
				// actual: 333stake to each address (=> 999uve actually sent)
				amount = sendCoins.MulInt(totalAddrs)
			} else {
				amount = coins.MulInt(totalAddrs)
			}

			msg := types.NewMsgMultiSend(types.NewInput(cctx.FromAddress, amount), output)

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cmd.Flags().Bool(cflags.FlagSplit, false, "Send the equally split token amount to each address")
	cflags.AddTxFlagsToCmd(cmd)

	return cmd
}
