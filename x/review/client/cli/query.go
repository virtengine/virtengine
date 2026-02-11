package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	reviewv1 "github.com/virtengine/virtengine/sdk/go/node/review/v1"
	"github.com/virtengine/virtengine/x/review/types"
)

// GetQueryCmd returns the root query command for review.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Review query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdReview(),
		CmdReviewsByUser(),
		CmdParams(),
	)

	return cmd
}

// CmdReview queries a review by ID.
func CmdReview() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review [review-id]",
		Short: "Query a review by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := reviewv1.NewQueryClient(clientCtx)
			resp, err := queryClient.Review(cmd.Context(), &reviewv1.QueryReviewRequest{ReviewId: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdReviewsByUser queries reviews by reviewer address.
func CmdReviewsByUser() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reviews-by-user [reviewer]",
		Short: "List reviews submitted by a reviewer",
		Args:  cobra.ExactArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query reviews by reviewer address.

Example:
$ %s query review reviews-by-user ve1reviewer
`, version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := reviewv1.NewQueryClient(clientCtx)
			resp, err := queryClient.ReviewsByUser(cmd.Context(), &reviewv1.QueryReviewsByUserRequest{Reviewer: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdParams queries module parameters.
func CmdParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query review module parameters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := reviewv1.NewQueryClient(clientCtx)
			resp, err := queryClient.Params(cmd.Context(), &reviewv1.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
