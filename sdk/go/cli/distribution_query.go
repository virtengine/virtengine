package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
)

// GetQueryDistributionCmd returns the cli query commands for this module
func GetQueryDistributionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the distribution module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetQueryDistributionParamsCmd(),
		GetQueryDistributionValidatorDistributionInfoCmd(),
		GetQueryDistributionValidatorOutstandingRewardsCmd(),
		GetQueryDistributionValidatorCommissionCmd(),
		GetQueryDistributionValidatorSlashesCmd(),
		GetQueryDistributionDelegatorRewardsCmd(),
		GetQueryDistributionCommunityPoolCmd(),
	)

	return cmd
}

// GetQueryDistributionParamsCmd implements the query params command.
func GetQueryDistributionParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "params",
		Args:              cobra.NoArgs,
		Short:             "Query distribution params",
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			cl := MustLightClientFromContext(ctx)

			res, err := cl.Query().Distribution().Params(ctx, &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return cl.PrintMessage(&res.Params)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetQueryDistributionValidatorDistributionInfoCmd implements the query validator distribution info command.
func GetQueryDistributionValidatorDistributionInfoCmd() *cobra.Command {
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()

	cmd := &cobra.Command{
		Use:   "validator-distribution-info [validator]",
		Args:  cobra.ExactArgs(1),
		Short: "Query validator distribution info",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query validator distribution info.
Example:
$ %s query distribution validator-distribution-info %s1lwjmdnks33xwnmfayc64ycprww49n33mtm92ne
`,
				version.AppName, bech32PrefixValAddr,
			),
		),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cl := MustLightClientFromContext(ctx)

			validatorAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			res, err := cl.Query().Distribution().ValidatorDistributionInfo(ctx, &types.QueryValidatorDistributionInfoRequest{
				ValidatorAddress: validatorAddr.String(),
			})
			if err != nil {
				return err
			}

			return cl.PrintMessage(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetQueryDistributionValidatorOutstandingRewardsCmd implements the query validator
// outstanding rewards command.
func GetQueryDistributionValidatorOutstandingRewardsCmd() *cobra.Command {
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()

	cmd := &cobra.Command{
		Use:   "validator-outstanding-rewards [validator]",
		Args:  cobra.ExactArgs(1),
		Short: "Query distribution outstanding (un-withdrawn) rewards for a validator and all their delegations",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query distribution outstanding (un-withdrawn) rewards for a validator and all their delegations.

Example:
$ %s query distribution validator-outstanding-rewards %s1lwjmdnks33xwnmfayc64ycprww49n33mtm92ne
`,
				version.AppName, bech32PrefixValAddr,
			),
		),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cl := MustLightClientFromContext(ctx)

			validatorAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			res, err := cl.Query().Distribution().ValidatorOutstandingRewards(
				ctx,
				&types.QueryValidatorOutstandingRewardsRequest{ValidatorAddress: validatorAddr.String()},
			)
			if err != nil {
				return err
			}

			return cl.PrintMessage(&res.Rewards)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetQueryDistributionValidatorCommissionCmd implements the query validator commission command.
func GetQueryDistributionValidatorCommissionCmd() *cobra.Command {
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()

	cmd := &cobra.Command{
		Use:   "commission [validator]",
		Args:  cobra.ExactArgs(1),
		Short: "Query distribution validator commission",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query validator commission rewards from delegators to that validator.

Example:
$ %s query distribution commission %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.AppName, bech32PrefixValAddr,
			),
		),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cl := MustLightClientFromContext(ctx)

			validatorAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			res, err := cl.Query().Distribution().ValidatorCommission(
				ctx,
				&types.QueryValidatorCommissionRequest{ValidatorAddress: validatorAddr.String()},
			)
			if err != nil {
				return err
			}

			return cl.PrintMessage(&res.Commission)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetQueryDistributionValidatorSlashesCmd implements the query validator slashes command.
func GetQueryDistributionValidatorSlashesCmd() *cobra.Command {
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()

	cmd := &cobra.Command{
		Use:   "slashes [validator] [start-height] [end-height]",
		Args:  cobra.ExactArgs(3),
		Short: "Query distribution validator slashes",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query all slashes of a validator for a given block range.

Example:
$ %s query distribution slashes %svaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj 0 100
`,
				version.AppName, bech32PrefixValAddr,
			),
		),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cl := MustLightClientFromContext(ctx)

			validatorAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			startHeight, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("start-height %s not a valid uint, please input a valid start-height", args[1])
			}

			endHeight, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("end-height %s not a valid uint, please input a valid end-height", args[2])
			}

			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := cl.Query().Distribution().ValidatorSlashes(
				ctx,
				&types.QueryValidatorSlashesRequest{
					ValidatorAddress: validatorAddr.String(),
					StartingHeight:   startHeight,
					EndingHeight:     endHeight,
					Pagination:       pageReq,
				},
			)
			if err != nil {
				return err
			}

			return cl.PrintMessage(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	cflags.AddPaginationFlagsToCmd(cmd, "validator slashes")
	return cmd
}

// GetQueryDistributionDelegatorRewardsCmd implements the query delegator rewards command.
func GetQueryDistributionDelegatorRewardsCmd() *cobra.Command {
	bech32PrefixAccAddr := sdk.GetConfig().GetBech32AccountAddrPrefix()
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()

	cmd := &cobra.Command{
		Use:   "rewards [delegator-addr] [validator-addr]",
		Args:  cobra.RangeArgs(1, 2),
		Short: "Query all distribution delegator rewards or rewards from a particular validator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query all rewards earned by a delegator, optionally restrict to rewards from a single validator.

Example:
$ %s query distribution rewards %s1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
$ %s query distribution rewards %s1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.AppName, bech32PrefixAccAddr, version.AppName, bech32PrefixAccAddr, bech32PrefixValAddr,
			),
		),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cl := MustLightClientFromContext(ctx)

			delegatorAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			// query for rewards from a particular delegation
			if len(args) == 2 {
				validatorAddr, err := sdk.ValAddressFromBech32(args[1])
				if err != nil {
					return err
				}

				res, err := cl.Query().Distribution().DelegationRewards(
					ctx,
					&types.QueryDelegationRewardsRequest{DelegatorAddress: delegatorAddr.String(), ValidatorAddress: validatorAddr.String()},
				)
				if err != nil {
					return err
				}

				return cl.PrintMessage(res)
			}

			res, err := cl.Query().Distribution().DelegationTotalRewards(
				ctx,
				&types.QueryDelegationTotalRewardsRequest{DelegatorAddress: delegatorAddr.String()},
			)
			if err != nil {
				return err
			}

			return cl.PrintMessage(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetQueryDistributionCommunityPoolCmd returns the command for fetching community pool info.
func GetQueryDistributionCommunityPoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "community-pool",
		Args:  cobra.NoArgs,
		Short: "Query the amount of coins in the community pool",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query all coins in the community pool which is under Governance control.

Example:
$ %s query distribution community-pool
`,
				version.AppName,
			),
		),
		PersistentPreRunE: QueryPersistentPreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			cl := MustLightClientFromContext(ctx)

			res, err := cl.Query().Distribution().CommunityPool(cmd.Context(), &types.QueryCommunityPoolRequest{})
			if err != nil {
				return err
			}

			return cl.PrintMessage(res)
		},
	}

	cflags.AddQueryFlagsToCmd(cmd)
	return cmd
}
