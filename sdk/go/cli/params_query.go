package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
)

// GetQueryParamsCmd returns a root CLI command handler for all x/params query commands.
func GetQueryParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the params module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetQueryParamsSubspaceCmd(),
	)

	return cmd
}

// GetQueryParamsSubspaceCmd returns a CLI command handler for querying subspace
// parameters managed by the x/params module.
func GetQueryParamsSubspaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "subspace [subspace] [key]",
		Short:             "Query for raw parameters by subspace and key",
		Args:              cobra.ExactArgs(2),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			params := proposal.QueryParamsRequest{Subspace: args[0], Key: args[1]}
			res, err := cl.Query().Params().Params(cmd.Context(), &params)
			if err != nil {
				return err
			}

			return cl.PrintMessage(&res.Param)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

