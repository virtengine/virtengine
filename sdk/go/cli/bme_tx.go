package cli

import (
	"github.com/spf13/cobra"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	types "github.com/virtengine/virtengine/sdk/go/node/bme/v1"
)

// GetTxBMECmd returns the transaction commands for bme module
func GetTxBMECmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "BME transaction subcommands",
		SuggestionsMinimumDistance: 2,
		RunE:                       sdkclient.ValidateCmd,
	}

	cmd.AddCommand(
		GetTxBMEBurnMintCmd(),
		GetTxBMEMintVACTCmd(),
		GetTxBMEBurnVACTCmd(),
	)

	return cmd
}

// GetTxBMEBurnMintCmd returns the command to burn one token and mint another
func GetTxBMEBurnMintCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "burn-mint [coins-to-burn] [denom-to-mint]",
		Short: "Burn tokens to mint another denomination",
		Long: `Burn tokens to mint another denomination.
This allows burning VE to mint vACT, or burning unused vACT back to VE.

Example:
  $ virtengine tx bme burn-mint 1000000uve uvact --from mykey
  $ virtengine tx bme burn-mint 500000uvact uve --from mykey`,
		Args:              cobra.ExactArgs(2),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)

			// Parse the coin to burn
			coinsToBurn, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			// Validate the denom to mint
			denomToMint := args[1]
			if err := sdk.ValidateDenom(denomToMint); err != nil {
				return err
			}

			// Get signer address from client context
			cctx, err := GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			fromAddr := cctx.GetFromAddress().String()

			msg := &types.MsgBurnMint{
				Owner:       fromAddr,
				To:          fromAddr,
				CoinsToBurn: coinsToBurn,
				DenomToMint: denomToMint,
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetTxBMEMintVACTCmd returns the command to burn VE tokens to mint vACT
func GetTxBMEMintVACTCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mint-vact [coins-to-burn]",
		Short: "Mint vACT by burning VE",
		Long: `
Example:
  $ virtengine tx bme mint-vact 500000uve --from mykey`,
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)

			// Parse the coin to burn
			coinsToBurn, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			// Get signer address from client context
			cctx, err := GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			fromAddr := cctx.GetFromAddress().String()

			msg := &types.MsgMintACT{
				Owner:       fromAddr,
				To:          fromAddr,
				CoinsToBurn: coinsToBurn,
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetTxBMEBurnVACTCmd returns the command to burn vACT tokens to mint/remint VE
func GetTxBMEBurnVACTCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "burn-vact [coins-to-burn]",
		Short: "Burn vACT tokens to mint/remint VE",
		Long: `
Example:
  $ virtengine tx bme burn-vact 500000uvact --from mykey`,
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: TxPersistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cl := MustClientFromContext(ctx)

			// Parse the coin to burn
			coinsToBurn, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			// Get signer address from client context
			cctx, err := GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			fromAddr := cctx.GetFromAddress().String()

			msg := &types.MsgBurnACT{
				Owner:       fromAddr,
				To:          fromAddr,
				CoinsToBurn: coinsToBurn,
			}

			resp, err := cl.Tx().BroadcastMsgs(ctx, []sdk.Msg{msg})
			if err != nil {
				return err
			}

			return cl.PrintMessage(resp)
		},
	}

	cflags.AddTxFlagsToCmd(cmd)

	return cmd
}

