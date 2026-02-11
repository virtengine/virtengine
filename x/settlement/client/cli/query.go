package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
	"github.com/virtengine/virtengine/x/settlement/types"
)

// GetQueryCmd returns the root query command for settlement.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Settlement query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdEscrow(),
		CmdEscrows(),
		CmdPayouts(),
		CmdUsageRecords(),
		CmdDisputes(),
		CmdEstimateRewards(),
		CmdFiatConversion(),
		CmdFiatPayoutPreference(),
	)

	return cmd
}

// CmdEscrow queries an escrow by ID.
func CmdEscrow() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "escrow [escrow-id]",
		Short: "Query escrow details by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := settlementv1.NewQueryClient(clientCtx)
			resp, err := queryClient.Escrow(cmd.Context(), &settlementv1.QueryEscrowRequest{EscrowId: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdEscrows queries escrows by order or state.
func CmdEscrows() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "escrows",
		Short: "List escrows with optional filters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			orderID, _ := cmd.Flags().GetString("order-id")
			state, _ := cmd.Flags().GetString("state")
			queryClient := settlementv1.NewQueryClient(clientCtx)
			switch {
			case orderID != "":
				resp, err := queryClient.EscrowsByOrder(cmd.Context(), &settlementv1.QueryEscrowsByOrderRequest{
					OrderId: orderID,
				})
				if err != nil {
					return err
				}
				return clientCtx.PrintProto(resp)
			case state != "":
				resp, err := queryClient.EscrowsByState(cmd.Context(), &settlementv1.QueryEscrowsByStateRequest{
					State: state,
				})
				if err != nil {
					return err
				}
				return clientCtx.PrintProto(resp)
			default:
				return fmt.Errorf("provide --order-id or --state to filter escrows")
			}
		},
	}

	cmd.Flags().String("order-id", "", "Filter escrows by order ID")
	cmd.Flags().String("state", "", "Filter escrows by state (pending, active, disputed, released, refunded, expired)")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdPayouts queries payouts by provider.
func CmdPayouts() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "payouts",
		Short: "List payout records by provider",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			provider, _ := cmd.Flags().GetString("provider")
			if provider == "" {
				return fmt.Errorf("--provider is required")
			}

			queryClient := settlementv1.NewQueryClient(clientCtx)
			resp, err := queryClient.PayoutsByProvider(cmd.Context(), &settlementv1.QueryPayoutsByProviderRequest{
				Provider: provider,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	cmd.Flags().String("provider", "", "Provider address to query payouts for")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdUsageRecords queries usage records by order.
func CmdUsageRecords() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "usage [order-id]",
		Short: "Query usage records by order ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := settlementv1.NewQueryClient(clientCtx)
			resp, err := queryClient.UsageRecordsByOrder(cmd.Context(), &settlementv1.QueryUsageRecordsByOrderRequest{
				OrderId: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdDisputes lists disputed escrows.
func CmdDisputes() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disputes",
		Short: "List disputed escrows",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := settlementv1.NewQueryClient(clientCtx)
			resp, err := queryClient.EscrowsByState(cmd.Context(), &settlementv1.QueryEscrowsByStateRequest{
				State: string(types.EscrowStateDisputed),
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, "disputes")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdEstimateRewards queries claimable rewards for a provider address.
func CmdEstimateRewards() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "estimate-rewards [provider]",
		Short: "Preview claimable rewards for a provider",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := settlementv1.NewQueryClient(clientCtx)
			resp, err := queryClient.ClaimableRewards(cmd.Context(), &settlementv1.QueryClaimableRewardsRequest{
				Address: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdFiatConversion queries a fiat conversion by ID.
func CmdFiatConversion() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fiat-conversion [conversion-id]",
		Short: "Query a fiat conversion record",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := settlementv1.NewQueryClient(clientCtx)
			resp, err := queryClient.FiatConversion(cmd.Context(), &settlementv1.QueryFiatConversionRequest{ConversionId: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdFiatPayoutPreference queries fiat payout preference by provider.
func CmdFiatPayoutPreference() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fiat-preference [provider]",
		Short: "Query fiat payout preference for a provider",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := settlementv1.NewQueryClient(clientCtx)
			resp, err := queryClient.FiatPayoutPreference(cmd.Context(), &settlementv1.QueryFiatPayoutPreferenceRequest{Provider: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
