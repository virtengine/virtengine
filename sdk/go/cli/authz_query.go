package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/authz"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
)

// GetQueryAuthzCmd returns the cli query commands for this module
func GetQueryAuthzCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        authz.ModuleName,
		Short:                      "Querying commands for the authz module",
		Long:                       "",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetQueryAuthzGrantsCmd(),
		GetQueryAuthzGranterGrantsCmd(),
		GetQueryAuthzGranteeGrantsCmd(),
	)

	return cmd
}

// GetQueryAuthzGrantsCmd implements the query authorization command.
func GetQueryAuthzGrantsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grants [granter-addr] [grantee-addr] [msg-type-url]?",
		Args:  cobra.RangeArgs(2, 3),
		Short: "query grants for a granter-grantee pair and optionally a msg-type-url",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query authorization grants for a granter-grantee pair. If msg-type-url
is set, it will select grants only for that msg type.
Examples:
$ %[1]s query %[2]s grants ve1skj.. ve1skjwj..
$ %[1]s query %[2]s grants ve1skjw.. ve1skjwj.. %[3]s
`,
				version.AppName, authz.ModuleName, bank.SendAuthorization{}.MsgTypeURL()),
		),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := authz.NewQueryClient(clientCtx)

			granter, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			grantee, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}
			msgAuthorized := ""
			if len(args) >= 3 {
				msgAuthorized = args[2]
			}
			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.Grants(
				cmd.Context(),
				&authz.QueryGrantsRequest{
					Granter:    granter.String(),
					Grantee:    grantee.String(),
					MsgTypeUrl: msgAuthorized,
					Pagination: pageReq,
				},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cflags.AddPaginationFlagsToCmd(cmd, "grants")

	return cmd
}

// GetQueryAuthzGranterGrantsCmd returns cmd to query for all grants for a granter.
func GetQueryAuthzGranterGrantsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grants-by-granter [granter-addr]",
		Args:  cobra.ExactArgs(1),
		Short: "query authorization grants granted by granter",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query authorization grants granted by granter.
Examples:
$ %s q %s grants-by-granter ve1skj..
`,
				version.AppName, authz.ModuleName),
		),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			granter, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := cl.Query().Authz().GranterGrants(
				cmd.Context(),
				&authz.QueryGranterGrantsRequest{
					Granter:    granter.String(),
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
	cflags.AddPaginationFlagsToCmd(cmd, "granter-grants")

	return cmd
}

// GetQueryAuthzGranteeGrantsCmd returns cmd to query for all grants for a grantee.
func GetQueryAuthzGranteeGrantsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grants-by-grantee [grantee-addr]",
		Args:  cobra.ExactArgs(1),
		Short: "query authorization grants granted to a grantee",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query authorization grants granted to a grantee.
Examples:
$ %s q %s grants-by-grantee ve1skj..
`,
				version.AppName, authz.ModuleName),
		),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			grantee, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := cl.Query().Authz().GranteeGrants(
				cmd.Context(),
				&authz.QueryGranteeGrantsRequest{
					Grantee:    grantee.String(),
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
	cflags.AddPaginationFlagsToCmd(cmd, "grantee-grants")

	return cmd
}

