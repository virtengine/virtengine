package cli

import (
	"github.com/spf13/cobra"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	types "github.com/virtengine/virtengine/sdk/go/node/audit/v1"
)

func GetQueryAuditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Audit query commands",
		SuggestionsMinimumDistance: 2,
		RunE:                       sdkclient.ValidateCmd,
	}

	cmd.AddCommand(
		GetAuditProvidersCmd(),
		GetAuditProviderCmd(),
	)

	return cmd
}

func GetAuditProvidersCmd() *cobra.Command {
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

			params := &types.QueryAllProvidersAttributesRequest{
				Pagination: pageReq,
			}

			res, err := cl.Query().Audit().AllProvidersAttributes(ctx, params)
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

func GetAuditProviderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "get [owner address] [auditor address]",
		Short:             "Query provider",
		Args:              cobra.RangeArgs(1, 2),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			owner, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			var res *types.QueryProvidersResponse
			if len(args) == 1 {
				res, err = cl.Query().Audit().ProviderAttributes(
					ctx,
					&types.QueryProviderAttributesRequest{
						Owner: owner.String(),
					},
				)
			} else {
				var auditor sdk.AccAddress
				if auditor, err = sdk.AccAddressFromBech32(args[1]); err != nil {
					return err
				}

				res, err = cl.Query().Audit().ProviderAuditorAttributes(
					ctx,
					&types.QueryProviderAuditorRequest{
						Auditor: auditor.String(),
						Owner:   owner.String(),
					},
				)
			}

			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

