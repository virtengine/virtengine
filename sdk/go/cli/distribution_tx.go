package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/spf13/cobra"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	cclient "github.com/virtengine/virtengine/sdk/go/node/client/v1beta3"
)

// Transaction flags for the x/distribution module
var (
	FlagCommission       = "commission"
	FlagMaxMessagesPerTx = "max-msgs"
)

const (
	MaxMessagesPerTxDefault = 0
)

// getTxDistributionCmd returns a root CLI command handler for all x/distribution transaction commands.
func getTxDistributionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Distribution transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetTxDistributionWithdrawRewardsCmd(),
		GetTxDistributionWithdrawAllRewardsCmd(),
		GetTxDistributionSetWithdrawAddrCmd(),
		GetTxDistributionFundCommunityPoolCmd(),
		GetTxDistributionDepositValidatorRewardsPoolCmd(),
	)

	return cmd
}

type distrGenerateOrBroadcastFunc func(context.Context, []sdk.Msg, ...cclient.BroadcastOption) (interface{}, error)

func newSplitAndApply(
	ctx context.Context,
	genOrBroadcastFn distrGenerateOrBroadcastFunc,
	msgs []sdk.Msg,
	chunkSize int,
	opts ...cclient.BroadcastOption,
) error {
	if chunkSize == 0 {
		if _, err := genOrBroadcastFn(ctx, msgs, opts...); err != nil {
			return err
		}
	}

	// split messages into slices of length chunkSize
	totalMessages := len(msgs)
	for i := 0; i < len(msgs); i += chunkSize {
		sliceEnd := i + chunkSize
		if sliceEnd > totalMessages {
			sliceEnd = totalMessages
		}

		msgChunk := msgs[i:sliceEnd]
		_, err := genOrBroadcastFn(ctx, msgChunk, opts...)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetTxDistributionWithdrawRewardsCmd returns a CLI command handler for creating a MsgWithdrawDelegatorReward transaction.
func GetTxDistributionWithdrawRewardsCmd() *cobra.Command {
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()

	cmd := &cobra.Command{
		Use:   "withdraw-rewards [validator-addr]",
		Short: "Withdraw rewards from a given delegation address, and optionally withdraw validator commission if the delegation address given is a validator operator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Withdraw rewards from a given delegation address,
and optionally withdraw validator commission if the delegation address given is a validator operator.

Example:
$ %s tx distribution withdraw-rewards %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj --from mykey
$ %s tx distribution withdraw-rewards %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj --from mykey --commission
`,
				version.AppName, bech32PrefixValAddr, version.AppName, bech32PrefixValAddr,
			),
		),
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			ac := MustAddressCodecFromContext(ctx)
			vc := MustValidatorCodecFromContext(ctx)
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			delAddr, err := ac.BytesToString(cctx.GetFromAddress())
			if err != nil {
				return err
			}

			_, err = vc.StringToBytes(args[0])
			if err != nil {
				return err
			}

			msgs := []sdk.Msg{types.NewMsgWithdrawDelegatorReward(delAddr, args[0])}

			if commission, _ := cmd.Flags().GetBool(FlagCommission); commission {
				msgs = append(msgs, types.NewMsgWithdrawValidatorCommission(args[0]))
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, msgs)
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cmd.Flags().Bool(FlagCommission, false, "Withdraw the validator's commission in addition to the rewards")
	cflags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetTxDistributionWithdrawAllRewardsCmd returns a CLI command handler for creating a MsgWithdrawDelegatorReward transaction.
func GetTxDistributionWithdrawAllRewardsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-all-rewards",
		Short: "withdraw all delegations rewards for a delegator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Withdraw all rewards for a single delegator.
Note that if you use this command with --%[2]s=%[3]s or --%[2]s=%[4]s, the %[5]s flag will automatically be set to 0.

Example:
$ %[1]s tx distribution withdraw-all-rewards --from mykey
`,
				version.AppName, cflags.FlagBroadcastMode, cflags.BroadcastSync, cflags.BroadcastAsync, FlagMaxMessagesPerTx,
			),
		),
		Args:              cobra.NoArgs,
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			ac := MustAddressCodecFromContext(ctx)
			vc := MustValidatorCodecFromContext(ctx)
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			delAddr, err := ac.BytesToString(cctx.GetFromAddress())
			if err != nil {
				return err
			}

			// The transaction cannot be generated offline since it requires a query
			// to get all the validators.
			if cctx.Offline {
				return fmt.Errorf("cannot generate tx in offline mode")
			}

			delValsRes, err := cl.Query().Distribution().DelegatorValidators(cmd.Context(), &types.QueryDelegatorValidatorsRequest{DelegatorAddress: delAddr})
			if err != nil {
				return err
			}

			validators := delValsRes.Validators
			// build multi-message transaction
			msgs := make([]sdk.Msg, 0, len(validators))
			for _, valAddr := range validators {
				_, err := vc.StringToBytes(valAddr)
				if err != nil {
					return err
				}

				msg := types.NewMsgWithdrawDelegatorReward(delAddr, valAddr)
				msgs = append(msgs, msg)
			}
			chunkSize, _ := cmd.Flags().GetInt(FlagMaxMessagesPerTx)

			return newSplitAndApply(ctx, cl.Tx().BroadcastMsgs, msgs, chunkSize)
		},
	}

	cmd.Flags().Int(FlagMaxMessagesPerTx, MaxMessagesPerTxDefault, "Limit the number of messages per tx (0 for unlimited)")
	cflags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetTxDistributionSetWithdrawAddrCmd returns a CLI command handler for creating a MsgSetWithdrawAddress transaction.
func GetTxDistributionSetWithdrawAddrCmd() *cobra.Command {
	bech32PrefixAccAddr := sdk.GetConfig().GetBech32AccountAddrPrefix()

	cmd := &cobra.Command{
		Use:   "set-withdraw-addr [withdraw-addr]",
		Short: "change the default withdraw address for rewards associated with an address",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Set the withdraw address for rewards associated with a delegator address.

Example:
$ %s tx distribution set-withdraw-addr %s1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p --from mykey
`,
				version.AppName, bech32PrefixAccAddr,
			),
		),
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			ac := MustAddressCodecFromContext(ctx)
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			delAddr := cctx.GetFromAddress()
			withdrawAddr, err := ac.StringToBytes(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgSetWithdrawAddress(delAddr, withdrawAddr)

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

// GetTxDistributionFundCommunityPoolCmd returns a CLI command handler for creating a MsgFundCommunityPool transaction.
func GetTxDistributionFundCommunityPoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fund-community-pool [amount]",
		Args:  cobra.ExactArgs(1),
		Short: "Funds the community pool with the specified amount",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Funds the community pool with the specified amount

Example:
$ %s tx distribution fund-community-pool 100uatom --from mykey
`,
				version.AppName,
			),
		),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			ac := MustAddressCodecFromContext(ctx)
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			depositorAddr, err := ac.BytesToString(cctx.GetFromAddress())
			if err != nil {
				return err
			}
			amount, err := sdk.ParseCoinsNormalized(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgFundCommunityPool(amount, depositorAddr)

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

// GetTxDistributionDepositValidatorRewardsPoolCmd returns a CLI command handler for creating
// a MsgDepositValidatorRewardsPool transaction.
func GetTxDistributionDepositValidatorRewardsPoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fund-validator-rewards-pool [val_addr] [amount]",
		Args:  cobra.ExactArgs(2),
		Short: "Fund the validator rewards pool with the specified amount",
		Example: fmt.Sprintf(
			"%s tx distribution fund-validator-rewards-pool cosmosvaloper1x20lytyf6zkcrv5edpkfkn8sz578qg5sqfyqnp 100uatom --from mykey",
			version.AppName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			ac := MustAddressCodecFromContext(ctx)
			vc := MustValidatorCodecFromContext(ctx)
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			depositorAddr, err := ac.BytesToString(cctx.GetFromAddress())
			if err != nil {
				return err
			}

			_, err = vc.StringToBytes(args[0])
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgDepositValidatorRewardsPool(depositorAddr, args[0], amount)
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
