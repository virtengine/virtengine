package cli

import (
	"github.com/spf13/cobra"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	types "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
)

// GetQueryProviderCmds returns the transaction commands for the provider module
func GetQueryProviderCmds() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Provider query commands",
		SuggestionsMinimumDistance: 2,
		RunE:                       sdkclient.ValidateCmd,
	}

	cmd.AddCommand(
		GetQueryProvidersCmd(),
		GetQueryProviderCmd(),
	)

	return cmd
}

func GetQueryProvidersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "list",
		Short:             "Query for all providers",
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			params := &types.QueryProvidersRequest{
				Pagination: pageReq,
			}

			res, err := cl.Query().Provider().Providers(ctx, params)
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cflags.AddPaginationFlagsToCmd(cmd, "providers")

	return cmd
}

func GetQueryProviderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "get [address]",
		Short:             "Query provider",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			owner, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			res, err := cl.Query().Provider().Provider(cmd.Context(), &types.QueryProviderRequest{Owner: owner.String()})
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(&res.Provider)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}
