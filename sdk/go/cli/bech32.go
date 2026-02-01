package cli

import (
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/spf13/cobra"

	"github.com/virtengine/virtengine/sdk/go/sdkutil"
)

var flagBech32Prefix = "prefix"

// ConvertBech32Cmd get cmd to convert any bech32 address to a ve prefix.
func ConvertBech32Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bech32-convert [bech32 string]",
		Short: "Convert any bech32 string to the ve prefix",
		Long: `Convert any bech32 string to the ve prefix
Especially useful for converting cosmos addresses to virtengine addresses
Example:
	virtengine bech32-convert ve1ey69r37gfxvxg62sh4r0ktpuc46pzjrmhykvmf
	`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bech32prefix, err := cmd.Flags().GetString(flagBech32Prefix)
			if err != nil {
				return err
			}

			_, bz, err := bech32.DecodeAndConvert(args[0])
			if err != nil {
				return err
			}

			bech32Addr, err := bech32.ConvertAndEncode(bech32prefix, bz)
			if err != nil {
				panic(err)
			}

			cmd.Println(bech32Addr)

			return nil
		},
	}

	cmd.Flags().StringP(flagBech32Prefix, "p", sdkutil.Bech32PrefixAccAddr, "Bech32 Prefix to encode to")

	return cmd
}

