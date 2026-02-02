package cli

import (
	"github.com/spf13/cobra"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	types "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
)

// GetQueryVEIDCmd returns the query commands for the veid module
func GetQueryVEIDCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "VEID query commands",
		SuggestionsMinimumDistance: 2,
		RunE:                       sdkclient.ValidateCmd,
	}

	cmd.AddCommand(
		GetQueryVEIDIdentityCmd(),
		GetQueryVEIDScoreCmd(),
		GetQueryVEIDScopesCmd(),
		GetQueryVEIDScopeCmd(),
		GetQueryVEIDWalletCmd(),
		GetQueryVEIDApprovedClientsCmd(),
		GetQueryVEIDParamsCmd(),
	)

	return cmd
}

func GetQueryVEIDIdentityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "identity [address]",
		Short:             "Query identity record",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			res, err := cl.Query().VEID().IdentityRecord(ctx, &types.QueryIdentityRecordRequest{
				AccountAddress: addr.String(),
			})
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetQueryVEIDScoreCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "score [address]",
		Short:             "Query VEID score",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			res, err := cl.Query().VEID().IdentityScore(ctx, &types.QueryIdentityScoreRequest{
				AccountAddress: addr.String(),
			})
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetQueryVEIDScopesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "scopes [address]",
		Short:             "List all scopes for an address",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := cl.Query().VEID().Scopes(ctx, &types.QueryScopesRequest{
				AccountAddress: addr.String(),
				Pagination:     pageReq,
			})
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cflags.AddPaginationFlagsToCmd(cmd, "scopes")

	return cmd
}

func GetQueryVEIDScopeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "scope [address] [scope-id]",
		Short:             "Get specific scope",
		Args:              cobra.ExactArgs(2),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			res, err := cl.Query().VEID().Scope(ctx, &types.QueryScopeRequest{
				AccountAddress: addr.String(),
				ScopeId:        args[1],
			})
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetQueryVEIDWalletCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "wallet [address]",
		Short:             "Get identity wallet",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			res, err := cl.Query().VEID().IdentityWallet(ctx, &types.QueryIdentityWalletRequest{
				AccountAddress: addr.String(),
			})
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetQueryVEIDApprovedClientsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "approved-clients",
		Short:             "List approved capture clients",
		Args:              cobra.ExactArgs(0),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := cl.Query().VEID().ApprovedClients(ctx, &types.QueryApprovedClientsRequest{
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cflags.AddPaginationFlagsToCmd(cmd, "approved-clients")

	return cmd
}

func GetQueryVEIDParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "params",
		Short:             "Get module parameters",
		Args:              cobra.ExactArgs(0),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			res, err := cl.Query().VEID().Params(ctx, &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}
