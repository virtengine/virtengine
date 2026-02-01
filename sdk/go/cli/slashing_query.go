package cli

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
)

// GetQuerySlashingCmd returns the cli query commands for this module
func GetQuerySlashingCmd() *cobra.Command {
	// Group slashing queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the slashing module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetQuerySlashingSigningInfoCmd(),
		GetQuerySlashingParamsCmd(),
		GetQuerySlashingSigningInfosCmd(),
	)

	return cmd
}

// GetQuerySlashingSigningInfoCmd implements the command to query signing info.
func GetQuerySlashingSigningInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "signing-info [validator-conspub]",
		Short: "Query a validator's signing information",
		Long: strings.TrimSpace(`Use a validators' consensus public key to find the signing-info for that validator:

$ <appd> query slashing signing-info '{"@type":"/cosmos.crypto.ed25519.PubKey","key":"OauFcTKbN5Lx3fJL689cikXBqe+hcp6Y+x0rYUdR9Jk="}'
`),
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			cctx := cl.ClientContext()

			var pk cryptotypes.PubKey
			if err := cctx.Codec.UnmarshalInterfaceJSON([]byte(args[0]), &pk); err != nil {
				return err
			}

			consAddr := sdk.ConsAddress(pk.Address())
			params := &types.QuerySigningInfoRequest{ConsAddress: consAddr.String()}
			res, err := cl.Query().Slashing().SigningInfo(cmd.Context(), params)
			if err != nil {
				return err
			}

			return cl.PrintMessage(&res.ValSigningInfo)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQuerySlashingSigningInfosCmd implements the command to query signing infos.
func GetQuerySlashingSigningInfosCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "signing-infos",
		Short: "Query signing information of all validators",
		Long: strings.TrimSpace(`signing infos of validators:

$ <appd> query slashing signing-infos
`),
		Args:              cobra.NoArgs,
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			params := &types.QuerySigningInfosRequest{Pagination: pageReq}
			res, err := cl.Query().Slashing().SigningInfos(cmd.Context(), params)
			if err != nil {
				return err
			}

			return cl.PrintMessage(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cflags.AddPaginationFlagsToCmd(cmd, "signing infos")

	return cmd
}

// GetQuerySlashingParamsCmd implements a command to fetch slashing parameters.
func GetQuerySlashingParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the current slashing parameters",
		Args:  cobra.NoArgs,
		Long: strings.TrimSpace(`Query genesis parameters for the slashing module:

$ <appd> query slashing params
`),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			params := &types.QueryParamsRequest{}
			res, err := cl.Query().Slashing().Params(cmd.Context(), params)
			if err != nil {
				return err
			}

			return cl.PrintMessage(&res.Params)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

