package cli

import (
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	"github.com/cosmos/cosmos-sdk/types/query"
)

// GetValidatorSetCmd returns the validator set for a given height
func GetValidatorSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validator-set [height]",
		Short: "Get the full CometBFT validator set at given height",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cctx, err := GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			var height *int64

			// optional height
			if len(args) > 0 {
				val, err := strconv.ParseInt(args[0], 10, 64)
				if err != nil {
					return err
				}

				if val > 0 {
					height = &val
				}
			}

			page, _ := cmd.Flags().GetInt(flags.FlagPage)
			limit, _ := cmd.Flags().GetInt(flags.FlagLimit)

			response, err := cmtservice.ValidatorsOutput(cmd.Context(), cctx, height, page, limit)
			if err != nil {
				return err
			}

			return cctx.PrintProto(response)
		},
	}

	cmd.Flags().String(flags.FlagNode, "tcp://localhost:26657", "<host>:<port> to CometBFT RPC interface for this chain")
	cmd.Flags().StringP(flags.FlagOutput, "o", "text", "Output format (text|json)")
	cmd.Flags().Int(flags.FlagPage, query.DefaultPage, "Query a specific page of paginated results")
	cmd.Flags().Int(flags.FlagLimit, 100, "Query number of results returned per page")

	return cmd
}
