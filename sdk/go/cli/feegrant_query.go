package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"cosmossdk.io/x/feegrant"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
)

// GetQueryFeegrantCmd returns the cli query commands for this module
func GetQueryFeegrantCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        feegrant.ModuleName,
		Short:                      "Querying commands for the feegrant module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetQueryFeeGrantCmd(),
		GetQueryFeeGrantsByGranteeCmd(),
		GetQueryFeeGrantsByGranterCmd(),
	)

	return cmd
}

// GetQueryFeeGrantCmd returns cmd to query for a grant between granter and grantee.
func GetQueryFeeGrantCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grant [granter] [grantee]",
		Args:  cobra.ExactArgs(2),
		Short: "Query details of a single grant",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details for a grant.
You can find the fee-grant of a granter and grantee.

Example:
$ %s query feegrant grant [granter] [grantee]
`, version.AppName),
		),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			granterAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			granteeAddr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			res, err := cl.Query().Feegrant().Allowance(
				cmd.Context(),
				&feegrant.QueryAllowanceRequest{
					Granter: granterAddr.String(),
					Grantee: granteeAddr.String(),
				},
			)
			if err != nil {
				return err
			}

			return cl.PrintMessage(res.Allowance)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryFeeGrantsByGranteeCmd returns cmd to query for all grants for a grantee.
func GetQueryFeeGrantsByGranteeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grants-by-grantee [grantee]",
		Args:  cobra.ExactArgs(1),
		Short: "Query all grants of a grantee",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Queries all the grants for a grantee address.

Example:
$ %s query feegrant grants-by-grantee [grantee]
`, version.AppName),
		),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			granteeAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := cl.Query().Feegrant().Allowances(
				cmd.Context(),
				&feegrant.QueryAllowancesRequest{
					Grantee:    granteeAddr.String(),
					Pagination: pageReq,
				},
			)
			if err != nil {
				return err
			}

			return cl.PrintMessage(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cflags.AddPaginationFlagsToCmd(cmd, "grants")

	return cmd
}

// GetQueryFeeGrantsByGranterCmd returns cmd to query for all grants by a granter.
func GetQueryFeeGrantsByGranterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grants-by-granter [granter]",
		Args:  cobra.ExactArgs(1),
		Short: "Query all grants by a granter",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Queries all the grants issued for a granter address.

Example:
$ %s query feegrant grants-by-granter [granter]
`, version.AppName),
		),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			granterAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := cl.Query().Feegrant().AllowancesByGranter(
				cmd.Context(),
				&feegrant.QueryAllowancesByGranterRequest{
					Granter:    granterAddr.String(),
					Pagination: pageReq,
				},
			)
			if err != nil {
				return err
			}

			return cl.PrintMessage(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cflags.AddPaginationFlagsToCmd(cmd, "grants")

	return cmd
}
