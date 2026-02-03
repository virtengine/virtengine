package cli

import (
	"github.com/spf13/cobra"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	mv1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"
	mtypes "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"
)

// GetTxMarketCmds returns the transaction commands for market module
func GetTxMarketCmds() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        mv1.ModuleName,
		Short:                      "Transaction subcommands",
		SuggestionsMinimumDistance: 2,
		RunE:                       sdkclient.ValidateCmd,
	}
	cmd.AddCommand(
		GetTxMarketBidCmds(),
		GetTxMarketLeaseCmds(),
	)
	return cmd
}

func GetTxMarketBidCmds() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "bid",
		Short:                      "Bid subcommands",
		SuggestionsMinimumDistance: 2,
		RunE:                       sdkclient.ValidateCmd,
	}
	cmd.AddCommand(
		GetTxMarketBidCreateCmd(),
		GetTxMarketBidCloseCmd(),
	)
	return cmd
}

func GetTxMarketBidCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "create",
		Short:             "Create a market bid",
		Args:              cobra.ExactArgs(0),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			price, err := cmd.Flags().GetString("price")
			if err != nil {
				return err
			}

			coin, err := sdk.ParseDecCoin(price)
			if err != nil {
				return err
			}

			id, err := cflags.OrderIDFromFlags(cmd.Flags(), cflags.WithProvider(cctx.FromAddress))
			if err != nil {
				return err
			}

			deposit, err := DetectDeposit(ctx, cmd.Flags(), cl.Query(), DetectBidDeposit)
			if err != nil {
				return err
			}

			msg := &mtypes.MsgCreateBid{
				ID:      mv1.MakeBidID(id, cctx.GetFromAddress()),
				Price:   coin,
				Deposit: deposit,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)
	cflags.AddOrderIDFlags(cmd.Flags())

	cmd.Flags().String("price", "", "Bid Price")
	cflags.AddDepositFlags(cmd.Flags())

	return cmd
}

func GetTxMarketBidCloseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "close",
		Short:             "Close a market bid",
		Args:              cobra.ExactArgs(0),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			id, err := cflags.BidIDFromFlags(cmd.Flags(), cflags.WithProvider(cctx.FromAddress))
			if err != nil {
				return err
			}

			reason, err := cflags.BidClosedReasonFromFlags(cmd.Flags())
			if err != nil {
				return err
			}

			msg := &mtypes.MsgCloseBid{
				ID:     id,
				Reason: reason,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)
	cflags.AddBidIDFlags(cmd.Flags())
	cflags.AddBidClosedReasonFlag(cmd.Flags())

	return cmd
}

func GetTxMarketLeaseCmds() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "lease",
		Short:                      "Lease subcommands",
		SuggestionsMinimumDistance: 2,
		RunE:                       sdkclient.ValidateCmd,
	}

	cmd.AddCommand(
		GetTxMarketLeaseCreateCmd(),
		GetTxMarketLeaseWithdrawCmd(),
		GetTxMarketLeaseCloseCmd(),
	)

	return cmd
}

func GetTxMarketLeaseCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "create",
		Short:             "Create a market lease",
		Args:              cobra.ExactArgs(0),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			id, err := cflags.LeaseIDFromFlags(cmd.Flags(), cflags.WithOwner(cctx.FromAddress))
			if err != nil {
				return err
			}

			msg := &mtypes.MsgCreateLease{
				BidID: id.BidID(),
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)
	cflags.AddLeaseIDFlags(cmd.Flags())
	cflags.MarkReqLeaseIDFlags(cmd, cflags.DeploymentIDOptionNoOwner(true))

	return cmd
}

func GetTxMarketLeaseWithdrawCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "withdraw",
		Short:             "Settle and withdraw available funds from market order escrow account",
		Args:              cobra.ExactArgs(0),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)

			cctx := cl.ClientContext()

			id, err := cflags.LeaseIDFromFlags(cmd.Flags(), cflags.WithOwner(cctx.FromAddress))
			if err != nil {
				return err
			}

			msg := &mtypes.MsgWithdrawLease{
				ID: id,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)
	cflags.AddLeaseIDFlags(cmd.Flags())
	cflags.MarkReqLeaseIDFlags(cmd, cflags.DeploymentIDOptionNoOwner(true))

	return cmd
}

func GetTxMarketLeaseCloseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "close",
		Short:             "Close a market order",
		Args:              cobra.ExactArgs(0),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)
			cctx := cl.ClientContext()

			id, err := cflags.LeaseIDFromFlags(cmd.Flags(), cflags.WithOwner(cctx.FromAddress))
			if err != nil {
				return err
			}

			// for lease closed tx reason is always owner
			msg := &mtypes.MsgCloseLease{
				ID:     id,
				Reason: mv1.LeaseClosedReasonOwner,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)
	cflags.AddLeaseIDFlags(cmd.Flags())
	cflags.MarkReqLeaseIDFlags(cmd, cflags.DeploymentIDOptionNoOwner(true))

	return cmd
}
