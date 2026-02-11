package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"

	reviewv1 "github.com/virtengine/virtengine/sdk/go/node/review/v1"
	"github.com/virtengine/virtengine/x/review/types"
)

// GetTxCmd returns the root tx command for review.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Review transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdSubmitReview(),
		CmdDeleteReview(),
		CmdUpdateParams(),
	)

	return cmd
}

// CmdSubmitReview submits a review.
func CmdSubmitReview() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-review [order-id] [provider] [rating] [comment]",
		Short: "Submit a provider review",
		Args:  cobra.ExactArgs(4),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a review for a completed order.

Example:
$ %s tx review submit-review order-123 ve1provider 5 "Excellent service" --from customer
`, version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			rating, err := strconv.ParseUint(args[2], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid rating: %w", err)
			}

			msg := types.NewMsgSubmitReview(
				clientCtx.GetFromAddress().String(),
				args[0],
				args[1],
				uint32(rating),
				args[3],
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdDeleteReview deletes a review.
func CmdDeleteReview() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-review [review-id] [reason]",
		Short: "Delete a review (governance only)",
		Args:  cobra.ExactArgs(2),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Delete a review using a governance-authorized account.

Example:
$ %s tx review delete-review review-123 "Policy violation" --from gov
`, version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgDeleteReview(
				clientCtx.GetFromAddress().String(),
				args[0],
				args[1],
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdUpdateParams updates review module parameters.
func CmdUpdateParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-params [params-file]",
		Short: "Update review module parameters (governance only)",
		Args:  cobra.ExactArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Update review module parameters from a JSON file.

Example:
$ %s tx review update-params ./review-params.json --from gov
`, version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			payload, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("failed to read params file: %w", err)
			}

			var params reviewv1.Params
			if err := json.Unmarshal(payload, &params); err != nil {
				return fmt.Errorf("failed to decode params: %w", err)
			}

			msg := types.NewMsgUpdateParams(clientCtx.GetFromAddress().String(), params)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
