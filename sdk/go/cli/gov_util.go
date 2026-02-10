package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
)

// AddGovPropFlagsToCmd adds flags for defining MsgSubmitProposal fields.
//
// See also ReadGovPropFlags.
func AddGovPropFlagsToCmd(cmd *cobra.Command) {
	cmd.Flags().String(cflags.FlagDeposit, "", "The deposit to include with the governance proposal")
	cmd.Flags().String(cflags.FlagMetadata, "", "The metadata to include with the governance proposal")
	cmd.Flags().String(cflags.FlagTitle, "", "The title to put on the governance proposal")
	cmd.Flags().String(cflags.FlagSummary, "", "The summary to include with the governance proposal")
}

// ReadGovPropFlags parses a MsgSubmitProposal from the provided context and flags.
// Setting the messages is up to the caller.
//
// See also AddGovPropFlagsToCmd.
func ReadGovPropFlags(clientCtx client.Context, flagSet *pflag.FlagSet) (*govv1.MsgSubmitProposal, error) {
	rv := &govv1.MsgSubmitProposal{}

	deposit, err := flagSet.GetString(cflags.FlagDeposit)
	if err != nil {
		return nil, fmt.Errorf("could not read deposit: %w", err)
	}
	if len(deposit) > 0 {
		rv.InitialDeposit, err = sdk.ParseCoinsNormalized(deposit)
		if err != nil {
			return nil, fmt.Errorf("invalid deposit: %w", err)
		}
	}

	rv.Metadata, err = flagSet.GetString(cflags.FlagMetadata)
	if err != nil {
		return nil, fmt.Errorf("could not read metadata: %w", err)
	}

	rv.Title, err = flagSet.GetString(cflags.FlagTitle)
	if err != nil {
		return nil, fmt.Errorf("could not read title: %w", err)
	}

	rv.Summary, err = flagSet.GetString(cflags.FlagSummary)
	if err != nil {
		return nil, fmt.Errorf("could not read summary: %w", err)
	}

	rv.Proposer = clientCtx.GetFromAddress().String()

	return rv, nil
}
