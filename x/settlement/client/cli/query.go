package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/virtengine/virtengine/x/settlement/types"
)

// GetQueryCmd returns the root query command for settlement.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Settlement query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdFiatConversion(),
		CmdFiatPayoutPreference(),
	)

	return cmd
}

// CmdFiatConversion queries a fiat conversion by ID.
func CmdFiatConversion() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fiat-conversion [conversion-id]",
		Short: "Query a fiat conversion record",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			key := types.FiatConversionKey(args[0])
			bz, _, err := clientCtx.QueryStore(key, types.StoreKey)
			if err != nil {
				return err
			}
			if len(bz) == 0 {
				return fmt.Errorf("conversion %s not found", args[0])
			}

			var conversion types.FiatConversionRecord
			if err := json.Unmarshal(bz, &conversion); err != nil {
				return err
			}
			out, err := json.MarshalIndent(conversion, "", "  ")
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), string(out))
			return err
		},
	}
	return cmd
}

// CmdFiatPayoutPreference queries fiat payout preference by provider.
func CmdFiatPayoutPreference() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fiat-preference [provider]",
		Short: "Query fiat payout preference for a provider",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			key := types.FiatPayoutPreferenceKey(args[0])
			bz, _, err := clientCtx.QueryStore(key, types.StoreKey)
			if err != nil {
				return err
			}
			if len(bz) == 0 {
				return fmt.Errorf("preference for %s not found", args[0])
			}

			var pref types.FiatPayoutPreference
			if err := json.Unmarshal(bz, &pref); err != nil {
				return err
			}
			out, err := json.MarshalIndent(pref, "", "  ")
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), string(out))
			return err
		},
	}
	return cmd
}
