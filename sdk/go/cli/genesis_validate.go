package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/types/module"
)

const chainUpgradeGuide = "https://github.com/cosmos/cosmos-sdk/blob/main/UPGRADING.md"

// getGenesisValidateCmd takes a genesis file, and makes sure that it is valid.
func getGenesisValidateCmd(mbm module.BasicManager) *cobra.Command {
	return &cobra.Command{
		Use:   "validate [file]",
		Args:  cobra.RangeArgs(0, 1),
		Short: "validates the genesis file at the default location or at the location passed as an arg",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			sctx := server.GetServerContextFromCmd(cmd)
			cctx := client.GetClientContextFromCmd(cmd)

			cdc := cctx.Codec

			// Load default if passed no args, otherwise load passed file
			var genesis string
			if len(args) == 0 {
				genesis = sctx.Config.GenesisFile()
			} else {
				genesis = args[0]
			}

			appGenesis, err := types.AppGenesisFromFile(genesis)
			if err != nil {
				return enrichUnmarshalError(err)
			}

			if err := appGenesis.ValidateAndComplete(); err != nil {
				return fmt.Errorf("make sure that you have correctly migrated all CometBFT consensus params. Refer the UPGRADING.md (%s): %w", chainUpgradeGuide, err)
			}

			var genState map[string]json.RawMessage
			if err = json.Unmarshal(appGenesis.AppState, &genState); err != nil {
				if strings.Contains(err.Error(), "unexpected end of JSON input") {
					return fmt.Errorf("app_state is missing in the genesis file: %s", err.Error())
				}
				return fmt.Errorf("error unmarshalling genesis doc %s: %w", genesis, err)
			}

			if err = mbm.ValidateGenesis(cdc, cctx.TxConfig, genState); err != nil {
				return fmt.Errorf("error validating genesis file %s: %s", genesis, err.Error())
			}

			if err = mbm.ValidateGenesis(cdc, cctx.TxConfig, genState); err != nil {
				errStr := fmt.Sprintf("error validating genesis file %s: %s", genesis, err.Error())
				if errors.Is(err, io.EOF) {
					errStr = fmt.Sprintf("%s: section is missing in the app_state", errStr)
				}
				return fmt.Errorf("%s", errStr)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "File at %s is a valid genesis file\n", genesis)

			return nil
		},
	}
}

func enrichUnmarshalError(err error) error {
	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return fmt.Errorf("error at offset %d: %s", syntaxErr.Offset, syntaxErr.Error())
	}
	return err
}
