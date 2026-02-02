package cli

import (
	"github.com/spf13/cobra"

	sdkclient "github.com/cosmos/cosmos-sdk/client"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	types "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
)

// GetQueryVEIDCmd returns the query commands for the VEID module
func GetQueryVEIDCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "VEID identity verification query commands",
		SuggestionsMinimumDistance: 2,
		RunE:                       sdkclient.ValidateCmd,
	}

	cmd.AddCommand(
		GetVEIDIdentityCmd(),
		GetVEIDIdentityRecordCmd(),
		GetVEIDScopeCmd(),
		GetVEIDScopesCmd(),
		GetVEIDIdentityScoreCmd(),
		GetVEIDIdentityStatusCmd(),
		GetVEIDIdentityWalletCmd(),
		GetVEIDWalletScopesCmd(),
		GetVEIDConsentSettingsCmd(),
		GetVEIDParamsCmd(),
	)

	return cmd
}

func GetVEIDIdentityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "identity [address]",
		Short:             "Query identity for an address",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			req := &types.QueryIdentityRequest{
				AccountAddress: args[0],
			}

			res, err := cl.Query().VEID().Identity(ctx, req)
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetVEIDIdentityRecordCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "identity-record [address]",
		Short:             "Query identity record for an address",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			req := &types.QueryIdentityRecordRequest{
				AccountAddress: args[0],
			}

			res, err := cl.Query().VEID().IdentityRecord(ctx, req)
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetVEIDScopeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "scope [address] [scope-id]",
		Short:             "Query a specific scope for an address",
		Args:              cobra.ExactArgs(2),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			req := &types.QueryScopeRequest{
				AccountAddress: args[0],
				ScopeId:        args[1],
			}

			res, err := cl.Query().VEID().Scope(ctx, req)
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetVEIDScopesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "scopes [address]",
		Short:             "Query all scopes for an address",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			req := &types.QueryScopesRequest{
				AccountAddress: args[0],
				Pagination:     pageReq,
			}

			res, err := cl.Query().VEID().Scopes(ctx, req)
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

func GetVEIDIdentityScoreCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "score [address]",
		Short:             "Query identity score for an address",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			req := &types.QueryIdentityScoreRequest{
				AccountAddress: args[0],
			}

			res, err := cl.Query().VEID().IdentityScore(ctx, req)
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetVEIDIdentityStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "status [address]",
		Short:             "Query identity verification status for an address",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			req := &types.QueryIdentityStatusRequest{
				AccountAddress: args[0],
			}

			res, err := cl.Query().VEID().IdentityStatus(ctx, req)
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetVEIDIdentityWalletCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "wallet [wallet-address]",
		Short:             "Query identity wallet by wallet address",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			req := &types.QueryIdentityWalletRequest{
				AccountAddress: args[0],
			}

			res, err := cl.Query().VEID().IdentityWallet(ctx, req)
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetVEIDWalletScopesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "wallet-scopes [wallet-address]",
		Short:             "Query scopes linked to a wallet",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			req := &types.QueryWalletScopesRequest{
				AccountAddress: args[0],
			}

			res, err := cl.Query().VEID().WalletScopes(ctx, req)
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetVEIDConsentSettingsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "consent [address]",
		Short:             "Query consent settings for an address",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			req := &types.QueryConsentSettingsRequest{
				AccountAddress: args[0],
			}

			res, err := cl.Query().VEID().ConsentSettings(ctx, req)
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetVEIDParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "params",
		Short:             "Query VEID module parameters",
		Args:              cobra.NoArgs,
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
