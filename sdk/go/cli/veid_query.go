package cli

import (
	"github.com/spf13/cobra"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

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
		GetQueryVEIDIdentityCmd(),
		GetQueryVEIDIdentityRecordCmd(),
		GetQueryVEIDScoreCmd(),
		GetQueryVEIDScopesCmd(),
		GetQueryVEIDScopeCmd(),
		GetQueryVEIDWalletCmd(),
		GetQueryVEIDWalletScopesCmd(),
		GetQueryVEIDIdentityStatusCmd(),
		GetQueryVEIDConsentSettingsCmd(),
		GetQueryVEIDApprovedClientsCmd(),
		GetQueryVEIDParamsCmd(),
		GetQueryVEIDActiveModelsCmd(),
		GetQueryVEIDModelVersionCmd(),
		GetQueryVEIDModelHistoryCmd(),
		GetQueryVEIDValidatorModelSyncCmd(),
		GetQueryVEIDModelParamsCmd(),
	)

	return cmd
}

func GetQueryVEIDIdentityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "identity [address]",
		Short:             "Query identity for an address",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			res, err := cl.Query().VEID().Identity(ctx, &types.QueryIdentityRequest{
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

func GetQueryVEIDIdentityRecordCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "identity-record [address]",
		Short:             "Query identity record for an address",
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

func GetQueryVEIDWalletScopesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "wallet-scopes [wallet-address]",
		Short:             "Query scopes linked to a wallet",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			res, err := cl.Query().VEID().WalletScopes(ctx, &types.QueryWalletScopesRequest{
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

func GetQueryVEIDIdentityStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "status [address]",
		Short:             "Query identity verification status for an address",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			res, err := cl.Query().VEID().IdentityStatus(ctx, &types.QueryIdentityStatusRequest{
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

func GetQueryVEIDConsentSettingsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "consent [address]",
		Short:             "Query consent settings for an address",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			res, err := cl.Query().VEID().ConsentSettings(ctx, &types.QueryConsentSettingsRequest{
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

// GetQueryVEIDActiveModelsCmd returns the command to query all active ML models
func GetQueryVEIDActiveModelsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "model-hashes",
		Short:             "Query all active model hashes",
		Args:              cobra.NoArgs,
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			res, err := cl.Query().VEID().ActiveModels(ctx, &types.QueryActiveModelsRequest{})
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetQueryVEIDModelVersionCmd returns the command to query a specific model version
func GetQueryVEIDModelVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "model-version [model-type]",
		Short:             "Query active model version by type",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			res, err := cl.Query().VEID().ModelVersion(ctx, &types.QueryModelVersionRequest{
				ModelType: args[0],
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

// GetQueryVEIDModelHistoryCmd returns the command to query model version history
func GetQueryVEIDModelHistoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "model-history [model-type]",
		Short:             "Query model version change history",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			res, err := cl.Query().VEID().ModelHistory(ctx, &types.QueryModelHistoryRequest{
				ModelType: args[0],
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

// GetQueryVEIDValidatorModelSyncCmd returns the command to query validator model sync status
func GetQueryVEIDValidatorModelSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "verify-model [validator-address]",
		Short:             "Query validator model sync status",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			res, err := cl.Query().VEID().ValidatorModelSync(ctx, &types.QueryValidatorModelSyncRequest{
				ValidatorAddress: args[0],
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

// GetQueryVEIDModelParamsCmd returns the command to query model management parameters
func GetQueryVEIDModelParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "model-params",
		Short:             "Query model management parameters",
		Args:              cobra.NoArgs,
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			res, err := cl.Query().VEID().ModelParams(ctx, &types.QueryModelParamsRequest{})
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	return cmd
}
