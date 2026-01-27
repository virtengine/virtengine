package cli

import (
	"github.com/spf13/cobra"

	sdkclient "github.com/cosmos/cosmos-sdk/client"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	types "github.com/virtengine/virtengine/sdk/go/node/oracle/v1"
)

func GetQueryOracleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Oracle query commands",
		SuggestionsMinimumDistance: 2,
		RunE:                       sdkclient.ValidateCmd,
	}

	cmd.AddCommand(
		GetOraclePricesCmd(),
		GetOracleAggregatedPriceCmd(),
		GetOraclePriceFeedConfigCmd(),
		GetOracleParamsCmd(),
	)

	return cmd
}

func GetOraclePricesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "prices",
		Aliases:           []string{"p"},
		Short:             "Query price history for denoms",
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			// Get filter flags
			assetDenom, _ := cmd.Flags().GetString(cflags.FlagAssetDenom)
			baseDenom, _ := cmd.Flags().GetString(cflags.FlagBaseDenom)
			height, _ := cmd.Flags().GetInt64(cflags.FlagHeight)

			// Build request with filters
			req := &types.QueryPricesRequest{
				Filters: types.PricesFilter{
					AssetDenom: assetDenom,
					BaseDenom:  baseDenom,
					Height:     height,
				},
				Pagination: pageReq,
			}

			res, err := cl.Query().Oracle().Prices(ctx, req)
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cflags.AddPaginationFlagsToCmd(cmd, "prices")
	cmd.Flags().String(cflags.FlagAssetDenom, "", "Filter by asset denomination (e.g., uakt)")
	cmd.Flags().String(cflags.FlagBaseDenom, "", "Filter by base denomination (e.g., usd)")

	return cmd
}

func GetOracleAggregatedPriceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "aggregated-price [denom]",
		Aliases:           []string{"ap"},
		Short:             "Query aggregated price for a denom",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			req := &types.QueryAggregatedPriceRequest{
				Denom: args[0],
			}

			res, err := cl.Query().Oracle().AggregatedPrice(ctx, req)
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetOraclePriceFeedConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "price-feed-config [denom]",
		Short:             "Query price feed configuration for a denom",
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			req := &types.QueryPriceFeedConfigRequest{
				Denom: args[0],
			}

			res, err := cl.Query().Oracle().PriceFeedConfig(ctx, req)
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetOracleParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "params",
		Short:             "Query the current oracle parameters",
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cl := MustLightClientFromContext(ctx)

			req := &types.QueryParamsRequest{}

			res, err := cl.Query().Oracle().Params(ctx, req)
			if err != nil {
				return err
			}

			return cl.ClientContext().PrintProto(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}
