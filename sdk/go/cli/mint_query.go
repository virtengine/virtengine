package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/mint/types"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
)

// GetQueryMintCmd returns the cli query commands for the minting module.
func GetQueryMintCmd() *cobra.Command {
	mintingQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the minting module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	mintingQueryCmd.AddCommand(
		GetQueryMintParamsCmd(),
		GetQueryMintInflationCmd(),
		GetQueryMintAnnualProvisionsCmd(),
	)

	return mintingQueryCmd
}

// GetQueryMintParamsCmd implements a command to return the current minting
// parameters.
func GetQueryMintParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "params",
		Short:             "Query the current minting parameters",
		Args:              cobra.NoArgs,
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			params := &types.QueryParamsRequest{}
			res, err := cl.Query().Mint().Params(cmd.Context(), params)
			if err != nil {
				return err
			}

			return cl.PrintMessage(&res.Params)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryMintInflationCmd implements a command to return the current minting
// inflation value.
func GetQueryMintInflationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "inflation",
		Short:             "Query the current minting inflation value",
		Args:              cobra.NoArgs,
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)
			cctx := cl.ClientContext()

			params := &types.QueryInflationRequest{}
			res, err := cl.Query().Mint().Inflation(cmd.Context(), params)
			if err != nil {
				return err
			}

			return cctx.PrintString(fmt.Sprintf("%s\n", res.Inflation))
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryMintAnnualProvisionsCmd implements a command to return the current minting
// annual provisions value.
func GetQueryMintAnnualProvisionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "annual-provisions",
		Short:             "Query the current minting annual provisions value",
		Args:              cobra.NoArgs,
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)
			cctx := cl.ClientContext()

			params := &types.QueryAnnualProvisionsRequest{}
			res, err := cl.Query().Mint().AnnualProvisions(cmd.Context(), params)
			if err != nil {
				return err
			}

			return cctx.PrintString(fmt.Sprintf("%s\n", res.AnnualProvisions))
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

