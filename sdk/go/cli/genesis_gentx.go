package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"

	"cosmossdk.io/core/address"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
)

// GetGenesisGenTxCmd builds the application's gentx command.
func GetGenesisGenTxCmd(mbm module.BasicManager, txEncCfg client.TxEncodingConfig, genBalIterator types.GenesisBalancesIterator, defaultNodeHome string, valAddressCodec address.Codec) *cobra.Command {
	ipDefault, _ := server.ExternalIP()
	fsCreateValidator, defaultsDesc := CreateValidatorMsgFlagSet(ipDefault)

	cmd := &cobra.Command{
		Use:   "gentx [key_name] [amount]",
		Short: "Generate a genesis tx carrying a self delegation",
		Args:  cobra.ExactArgs(2),
		Long: fmt.Sprintf(`Generate a genesis transaction that creates a validator with a self-delegation,
that is signed by the key in the Keyring referenced by a given name. A node ID and Bech32 consensus
pubkey may optionally be provided. If they are omitted, they will be retrieved from the priv_validator.json
file. The following default parameters are included:
    %s

Example:
$ %s gentx my-key-name 1000000uve --home=/path/to/home/dir --keyring-backend=os --chain-id=test-chain-1 \
    --moniker="myvalidator" \
    --commission-max-change-rate=0.01 \
    --commission-max-rate=1.0 \
    --commission-rate=0.07 \
    --details="..." \
    --security-contact="..." \
    --website="..."
`, defaultsDesc, version.AppName,
		),
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			gas, err := cmd.Flags().GetString(cflags.FlagGas)
			if err != nil {
				return err
			}

			if gas == cflags.GasFlagAuto {
				err = cmd.Flags().Set(cflags.FlagGas, strconv.Itoa(cflags.DefaultGasLimit))
				if err != nil {
					return err
				}
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			stcx := server.GetServerContextFromCmd(cmd)
			cctx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			cdc := cctx.Codec

			config := stcx.Config
			config.SetRoot(cctx.HomeDir)

			nodeID, valPubKey, err := genutil.InitializeNodeValidatorFiles(stcx.Config)
			if err != nil {
				return fmt.Errorf("%w: failed to initialize node validator files", err)
			}

			// read --nodeID, if an empty take it from priv_validator.json
			if nodeIDString, _ := cmd.Flags().GetString(cflags.FlagNodeID); nodeIDString != "" {
				nodeID = nodeIDString
			}

			// read --pubkey, if an empty take it from priv_validator.json
			if pkStr, _ := cmd.Flags().GetString(cflags.FlagPubKey); pkStr != "" {
				if err := cctx.Codec.UnmarshalInterfaceJSON([]byte(pkStr), &valPubKey); err != nil {
					return fmt.Errorf("%w: failed to unmarshal validator public key", err)
				}
			}

			appGenesis, err := types.AppGenesisFromFile(config.GenesisFile())
			if err != nil {
				return fmt.Errorf("%w: failed to read genesis doc file %s", err, config.GenesisFile())
			}

			var genesisState map[string]json.RawMessage
			if err = json.Unmarshal(appGenesis.AppState, &genesisState); err != nil {
				return fmt.Errorf("%w: failed to unmarshal genesis state", err)
			}

			if err = mbm.ValidateGenesis(cdc, txEncCfg, genesisState); err != nil {
				return fmt.Errorf("%w: failed to validate genesis state", err)
			}

			inBuf := bufio.NewReader(cmd.InOrStdin())

			name := args[0]
			key, err := cctx.Keyring.Key(name)
			if err != nil {
				return fmt.Errorf("%w: failed to fetch '%s' from the keyring", err, name)
			}

			moniker := config.Moniker
			if m, _ := cmd.Flags().GetString(cflags.FlagMoniker); m != "" {
				moniker = m
			}

			// set flags for creating a gentx
			createValCfg, err := PrepareConfigForTxCreateValidator(cmd.Flags(), moniker, nodeID, appGenesis.ChainID, valPubKey)
			if err != nil {
				return fmt.Errorf("%w: error creating configuration to create validator msg", err)
			}

			amount := args[1]
			coins, err := sdk.ParseCoinsNormalized(amount)
			if err != nil {
				return fmt.Errorf("%w: failed to parse coins", err)
			}
			addr, err := key.GetAddress()
			if err != nil {
				return err
			}
			err = genutil.ValidateAccountInGenesis(genesisState, genBalIterator, addr, coins, cdc)
			if err != nil {
				return fmt.Errorf("%w: failed to validate account in genesis", err)
			}

			txFactory, err := tx.NewFactoryCLI(cctx, cmd.Flags())
			if err != nil {
				return err
			}

			pub, err := key.GetAddress()
			if err != nil {
				return err
			}
			cctx = cctx.WithInput(inBuf).WithFromAddress(pub)

			// The following line comes from a discrepancy between the `gentx`
			// and `create-validator` commands:
			// - `gentx` expects amount as an arg,
			// - `create-validator` expects amount as a required flag.
			// ref: https://github.com/cosmos/cosmos-sdk/issues/8251
			// Since gentx doesn't set the amount flag (which `create-validator`
			// reads from), we copy the amount arg into the valCfg directly.
			//
			// Ideally, the `create-validator` command should take a validator
			// config file instead of so many flags.
			// ref: https://github.com/cosmos/cosmos-sdk/issues/8177
			createValCfg.Amount = amount

			// create a 'create-validator' message
			txBldr, msg, err := BuildCreateValidatorMsg(cctx, createValCfg, txFactory, true, valAddressCodec)
			if err != nil {
				return fmt.Errorf("%w: failed to build create-validator message", err)
			}

			if key.GetType() == keyring.TypeOffline || key.GetType() == keyring.TypeMulti {
				cmd.PrintErrln("Offline key passed in. Use `tx sign` command to sign.")
				return txBldr.PrintUnsignedTx(cctx, msg)
			}

			// write the unsigned transaction to the buffer
			w := bytes.NewBuffer([]byte{})
			cctx = cctx.WithOutput(w)

			if m, ok := msg.(sdk.HasValidateBasic); ok {
				if err := m.ValidateBasic(); err != nil {
					return err
				}
			}

			if err = txBldr.PrintUnsignedTx(cctx, msg); err != nil {
				return fmt.Errorf("%w: failed to print unsigned std tx", err)
			}

			// read the transaction
			stdTx, err := readUnsignedGenTxFile(cctx, w)
			if err != nil {
				return fmt.Errorf("%w: failed to read unsigned gen tx file", err)
			}

			// sign the transaction and write it to the output file
			txBuilder, err := cctx.TxConfig.WrapTxBuilder(stdTx)
			if err != nil {
				return fmt.Errorf("error creating tx builder: %w", err)
			}

			err = authclient.SignTx(txFactory, cctx, name, txBuilder, true, true)
			if err != nil {
				return fmt.Errorf("%w: failed to sign std tx", err)
			}

			outputDocument, _ := cmd.Flags().GetString(cflags.FlagOutputDocument)
			if outputDocument == "" {
				outputDocument, err = makeOutputFilepath(config.RootDir, nodeID)
				if err != nil {
					return fmt.Errorf("%w: failed to create output file path", err)
				}
			}

			if err := writeSignedGenTx(cctx, outputDocument, stdTx); err != nil {
				return fmt.Errorf("%w: failed to write signed gen tx", err)
			}

			cmd.PrintErrf("Genesis transaction written to %q\n", outputDocument)
			return nil
		},
	}

	cmd.Flags().String(cflags.FlagHome, defaultNodeHome, "The application home directory")
	cmd.Flags().String(cflags.FlagOutputDocument, "", "Write the genesis transaction JSON document to the given file instead of the default location")
	cmd.Flags().AddFlagSet(fsCreateValidator)
	cflags.AddTxFlagsToCmd(cmd)
	_ = cmd.Flags().MarkHidden(cflags.FlagOutput) // signing makes sense to output only json

	return cmd
}

func makeOutputFilepath(rootDir, nodeID string) (string, error) {
	writePath := filepath.Join(rootDir, "config", "gentx")
	if err := os.MkdirAll(writePath, 0o700); err != nil {
		return "", fmt.Errorf("could not create directory %q: %w", writePath, err)
	}

	return filepath.Join(writePath, fmt.Sprintf("gentx-%v.json", nodeID)), nil
}

func readUnsignedGenTxFile(clientCtx client.Context, r io.Reader) (sdk.Tx, error) {
	bz, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	aTx, err := clientCtx.TxConfig.TxJSONDecoder()(bz)
	if err != nil {
		return nil, err
	}

	return aTx, err
}

func writeSignedGenTx(clientCtx client.Context, outputDocument string, tx sdk.Tx) error {
	outputFile, err := os.OpenFile(outputDocument, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644) //nolint: gosec
	if err != nil {
		return err
	}
	defer func() {
		_ = outputFile.Close()
	}()

	txj, err := clientCtx.TxConfig.TxJSONEncoder()(tx)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(outputFile, "%s\n", txj)

	return err
}

