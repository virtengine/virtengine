package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/bank/types"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
)

// GetQueryBankCmd returns the parent command for all x/bank CLi query commands. The
// provided cctx should have, at a minimum, a verifier, Tendermint RPC client,
// and marshaler set.
func GetQueryBankCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the bank module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetQueryBankBalancesCmd(),
		GetQueryBankSpendableBalancesCmd(),
		GetQueryBankTotalSupplyCmd(),
		GetQueryBankDenomsMetadataCmd(),
		GetQueryBankSendEnabledCmd(),
	)

	return cmd
}

func GetQueryBankBalancesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balances [address]",
		Short: "Query for account balances by address",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the total balance of an account or of a specific denomination.

Example:
  $ %s query %s balances [address]
  $ %s query %s balances [address] --denom=[denom]
`,
				version.AppName, types.ModuleName, version.AppName, types.ModuleName,
			),
		),
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			denom, err := cmd.Flags().GetString(cflags.FlagDenom)
			if err != nil {
				return err
			}

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			if denom == "" {
				params := types.NewQueryAllBalancesRequest(addr, pageReq, true)

				res, err := cl.Query().Bank().AllBalances(ctx, params)
				if err != nil {
					return err
				}

				return cl.PrintMessage(&res)
			}

			params := types.NewQueryBalanceRequest(addr, denom)

			res, err := cl.Query().Bank().Balance(ctx, params)
			if err != nil {
				return err
			}

			return cl.PrintMessage(&res.Balance)
		},
	}

	cmd.Flags().String(cflags.FlagDenom, "", "The specific balance denomination to query for")
	cflags.AddQueryFlagsToCmd(cmd)
	cflags.AddPaginationFlagsToCmd(cmd, "all balances")

	return cmd
}

func GetQueryBankSpendableBalancesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "spendable-balances [address]",
		Short:             "Query for account spendable balances by address",
		Example:           fmt.Sprintf("$ %s query %s spendable-balances [address]", version.AppName, types.ModuleName),
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			denom, err := cmd.Flags().GetString(cflags.FlagDenom)
			if err != nil {
				return err
			}

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			if denom == "" {
				params := types.NewQuerySpendableBalancesRequest(addr, pageReq)

				res, err := cl.Query().Bank().SpendableBalances(ctx, params)
				if err != nil {
					return err
				}

				return cl.PrintMessage(res)
			}

			params := types.NewQuerySpendableBalanceByDenomRequest(addr, denom)

			res, err := cl.Query().Bank().SpendableBalanceByDenom(ctx, params)
			if err != nil {
				return err
			}

			return cl.PrintMessage(res)
		},
	}

	cmd.Flags().String(cflags.FlagDenom, "", "The specific balance denomination to query for")
	cflags.AddQueryFlagsToCmd(cmd)
	cflags.AddPaginationFlagsToCmd(cmd, "spendable balances")

	return cmd
}

// GetQueryBankDenomsMetadataCmd defines the cobra command to query client denomination metadata.
func GetQueryBankDenomsMetadataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "denom-metadata",
		Short: "Query the client metadata for coin denominations",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the client metadata for all the registered coin denominations

Example:
  To query for the client metadata of all coin denominations use:
  $ %s query %s denom-metadata

To query for the client metadata of a specific coin denomination use:
  $ %s query %s denom-metadata --denom=[denom]
`,
				version.AppName, types.ModuleName, version.AppName, types.ModuleName,
			),
		),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			denom, err := cmd.Flags().GetString(cflags.FlagDenom)
			if err != nil {
				return err
			}

			if denom == "" {
				res, err := cl.Query().Bank().DenomsMetadata(cmd.Context(), &types.QueryDenomsMetadataRequest{})
				if err != nil {
					return err
				}

				return cl.PrintMessage(res)
			}

			res, err := cl.Query().Bank().DenomMetadata(cmd.Context(), &types.QueryDenomMetadataRequest{Denom: denom})
			if err != nil {
				return err
			}

			return cl.PrintMessage(res)
		},
	}

	cmd.Flags().String(cflags.FlagDenom, "", "The specific denomination to query client metadata for")
	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetQueryBankTotalSupplyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "total",
		Short: "Query the total supply of coins of the chain",
		Args:  cobra.NoArgs,
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query total supply of coins that are held by accounts in the chain.

Example:
  $ %s query %s total

To query for the total supply of a specific coin denomination use:
  $ %s query %s total --denom=[denom]
`,
				version.AppName, types.ModuleName, version.AppName, types.ModuleName,
			),
		),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			denom, err := cmd.Flags().GetString(cflags.FlagDenom)
			if err != nil {
				return err
			}

			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			if denom == "" {
				res, err := cl.Query().Bank().TotalSupply(ctx, &types.QueryTotalSupplyRequest{Pagination: pageReq})
				if err != nil {
					return err
				}

				return cl.PrintMessage(res)
			}

			res, err := cl.Query().Bank().SupplyOf(ctx, &types.QuerySupplyOfRequest{Denom: denom})
			if err != nil {
				return err
			}

			return cl.PrintMessage(&res.Amount)
		},
	}

	cmd.Flags().String(cflags.FlagDenom, "", "The specific balance denomination to query for")
	cflags.AddQueryFlagsToCmd(cmd)
	cflags.AddPaginationFlagsToCmd(cmd, "all supply totals")

	return cmd
}

func GetQueryBankSendEnabledCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send-enabled [denom1 ...]",
		Short: "Query for send enabled entries",
		Long: strings.TrimSpace(`Query for send enabled entries that have been specifically set.

To look up one or more specific denoms, supply them as arguments to this command.
To look up all denoms, do not provide any arguments.
`,
		),
		Example: strings.TrimSpace(
			fmt.Sprintf(`Getting one specific entry:
  $ %[1]s query %[2]s send-enabled foocoin

Getting two specific entries:
  $ %[1]s query %[2]s send-enabled foocoin barcoin

Getting all entries:
  $ %[1]s query %[2]s send-enabled
`,
				version.AppName, types.ModuleName,
			),
		),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			reqPag, err := ReadPageRequest(client.MustFlagSetWithPageKeyDecoded(cmd.Flags()))
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			req := &types.QuerySendEnabledRequest{
				Denoms:     args,
				Pagination: reqPag,
			}

			res, err := cl.Query().Bank().SendEnabled(cmd.Context(), req)
			if err != nil {
				return err
			}

			return cl.PrintMessage(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cflags.AddPaginationFlagsToCmd(cmd, "send enabled entries")

	return cmd
}

