package cli

import (
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
		CmdFiatConversion(),
		CmdFiatPayoutPreference(),
	)

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
