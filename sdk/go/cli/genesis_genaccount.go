package cli

import (
	"bufio"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
)

// GetGenesisAddAccountCmd returns add-genesis-account cobra Command.
// This command is provided as a default, applications are expected to provide their own command if custom genesis accounts are needed.
func GetGenesisAddAccountCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-account [address_or_key_name] [coin][,[coin]]",
		Short: "Add a genesis account to genesis.json",
		Long: `Add a genesis account to genesis.json. The provided account must specify
the account address or key name and a list of initial coins. If a key name is given,
the address will be looked up in the local Keybase. The list of initial tokens must
contain valid denominations. Accounts may optionally be supplied with vesting parameters.
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cctx := client.GetClientContextFromCmd(cmd)
			sctx := server.GetServerContextFromCmd(cmd)

			config := sctx.Config

			config.SetRoot(cctx.HomeDir)

			var kr keyring.Keyring
			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				inBuf := bufio.NewReader(cmd.InOrStdin())
				keyringBackend, _ := cmd.Flags().GetString(cflags.FlagKeyringBackend)

				if keyringBackend != "" && cctx.Keyring == nil {
					var err error
					kr, err = keyring.New(sdk.KeyringServiceName(), keyringBackend, cctx.HomeDir, inBuf, cctx.Codec)
					if err != nil {
						return err
					}
				} else {
					kr = cctx.Keyring
				}

				k, err := kr.Key(args[0])
				if err != nil {
					return fmt.Errorf("failed to get address from Keyring: %w", err)
				}

				addr, err = k.GetAddress()
				if err != nil {
					return err
				}
			}

			appendFlag, _ := cmd.Flags().GetBool(cflags.FlagAppendMode)
			vestingStart, _ := cmd.Flags().GetInt64(cflags.FlagVestingStart)
			vestingEnd, _ := cmd.Flags().GetInt64(cflags.FlagVestingEnd)
			vestingAmtStr, _ := cmd.Flags().GetString(cflags.FlagVestingAmt)
			moduleNameStr, _ := cmd.Flags().GetString(cflags.FlagModuleName)

			return genutil.AddGenesisAccount(cctx.Codec, addr, appendFlag, config.GenesisFile(), args[1], vestingAmtStr, vestingStart, vestingEnd, moduleNameStr)
		},
	}

	cmd.Flags().String(cflags.FlagHome, defaultNodeHome, "The application home directory")
	cmd.Flags().String(cflags.FlagKeyringBackend, cflags.DefaultKeyringBackend, "Select keyring's backend (os|file|kwallet|pass|test)")
	cmd.Flags().String(cflags.FlagVestingAmt, "", "amount of coins for vesting accounts")
	cmd.Flags().Int64(cflags.FlagVestingStart, 0, "schedule start time (unix epoch) for vesting accounts")
	cmd.Flags().Int64(cflags.FlagVestingEnd, 0, "schedule end time (unix epoch) for vesting accounts")
	cmd.Flags().Bool(cflags.FlagAppendMode, false, "append the coins to an account already in the genesis.json file")
	cmd.Flags().String(cflags.FlagModuleName, "", "module account name")

	cflags.AddQueryFlagsToCmd(cmd)

	return cmd
}

