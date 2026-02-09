package cli

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/virtengine/virtengine/pkg/security"

	cmtcfg "github.com/cometbft/cometbft/config"
	tmrand "github.com/cometbft/cometbft/libs/rand"
	cmtypes "github.com/cometbft/cometbft/types"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/go-bip39"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
)

type printInfo struct {
	Moniker    string          `json:"moniker" yaml:"moniker"`
	ChainID    string          `json:"chain_id" yaml:"chain_id"`
	NodeID     string          `json:"node_id" yaml:"node_id"`
	GenTxsDir  string          `json:"gentxs_dir" yaml:"gentxs_dir"`
	AppMessage json.RawMessage `json:"app_message" yaml:"app_message"`
}

func newPrintInfo(moniker, chainID, nodeID, genTxsDir string, appMessage json.RawMessage) printInfo {
	return printInfo{
		Moniker:    moniker,
		ChainID:    chainID,
		NodeID:     nodeID,
		GenTxsDir:  genTxsDir,
		AppMessage: appMessage,
	}
}

func displayInfo(info printInfo) error {
	out, err := json.MarshalIndent(info, "", " ")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(os.Stderr, "%s\n", sdk.MustSortJSON(out))

	return err
}

// GetGenesisInitCmd returns a command that initializes all files needed for Tendermint
// and the respective application.
func GetGenesisInitCmd(mbm module.BasicManager, defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [moniker]",
		Short: "Initialize private validator, p2p, genesis, and application configuration files",
		Long:  `Initialize validators's and node's configuration files.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			cdc := clientCtx.Codec

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config
			config.SetRoot(clientCtx.HomeDir)

			dlGenesis := true
			chainID, _ := cmd.Flags().GetString(cflags.FlagChainID)
			switch {
			case chainID != "":
			case clientCtx.ChainID != "":
				chainID = clientCtx.ChainID
			default:
				dlGenesis = false
				chainID = fmt.Sprintf("test-chain-%v", tmrand.Str(6))
			}

			// Get bip39 mnemonic
			var mnemonic string
			if isRecover, _ := cmd.Flags().GetBool(cflags.FlagRecover); isRecover {
				inBuf := bufio.NewReader(cmd.InOrStdin())
				value, err := input.GetString("Enter your bip39 mnemonic", inBuf)
				if err != nil {
					return err
				}

				mnemonic = value
				if !bip39.IsMnemonicValid(mnemonic) {
					return errors.New("invalid mnemonic")
				}
			}

			// Get initial height
			initHeight, _ := cmd.Flags().GetInt64(cflags.FlagInitHeight)
			if initHeight < 1 {
				initHeight = 1
			}

			nodeID, _, err := genutil.InitializeNodeValidatorFilesFromMnemonic(config, mnemonic)
			if err != nil {
				return err
			}

			config.Moniker = args[0]

			genFile := config.GenesisFile()
			overwrite, _ := cmd.Flags().GetBool(cflags.FlagOverwrite)

			// use os.Stat to check if the file exists
			_, err = os.Stat(genFile)
			if !overwrite && !os.IsNotExist(err) {
				return fmt.Errorf("genesis.json file already exists: %v", genFile)
			}

			if dlGenesis {
				// If the chainID is blank or virtengine-2, prep this as a mainnet node

				// Attempt to download the genesis file from the VirtEngine Network GitHub repository
				// Generate a new genesis file if failed
				err = downloadGenesis(config, chainID)
			}

			if !dlGenesis || err != nil {
				// Overwrites the SDK default denom for side effects
				if val, _ := cmd.Flags().GetString(cflags.FlagDefaultBondDenom); val != "" {
					sdk.DefaultBondDenom = val
				}

				appGenState := mbm.DefaultGenesis(cdc)

				appState, err := json.MarshalIndent(appGenState, "", " ")
				if err != nil {
					return errorsmod.Wrap(err, "Failed to marshal default genesis state")
				}

				appGenesis := &types.AppGenesis{}
				if _, err := os.Stat(genFile); err != nil {
					if !os.IsNotExist(err) {
						return err
					}
				} else {
					appGenesis, err = types.AppGenesisFromFile(genFile)
					if err != nil {
						return errorsmod.Wrap(err, "Failed to read genesis doc from file")
					}
				}

				appGenesis.AppName = version.AppName
				appGenesis.AppVersion = version.Version
				appGenesis.ChainID = chainID
				appGenesis.AppState = appState
				appGenesis.InitialHeight = initHeight
				appGenesis.Consensus = &types.ConsensusGenesis{
					Validators: nil,
					Params:     cmtypes.DefaultConsensusParams(),
				}

				consensusKey, err := cmd.Flags().GetString(cflags.FlagConsensusKeyAlgo)
				if err != nil {
					return errorsmod.Wrap(err, "Failed to get consensus key algo")
				}

				appGenesis.Consensus.Params.Validator.PubKeyTypes = []string{consensusKey}

				if err = genutil.ExportGenesisFile(appGenesis, genFile); err != nil {
					return errorsmod.Wrap(err, "Failed to export genesis file")
				}

				toPrint := newPrintInfo(config.Moniker, chainID, nodeID, "", appState)

				cmtcfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)
				return displayInfo(toPrint)
			}

			return nil
		},
	}

	cmd.Flags().String(cflags.FlagHome, defaultNodeHome, "node's home directory")
	cmd.Flags().BoolP(cflags.FlagOverwrite, "o", false, "overwrite the genesis.json file")
	cmd.Flags().Bool(cflags.FlagRecover, false, "provide seed phrase to recover existing key instead of creating")
	cmd.Flags().String(cflags.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	cmd.Flags().String(cflags.FlagDefaultBondDenom, "uve", "genesis file default denomination, if left blank default value is 'uve'")
	cmd.Flags().Int64(cflags.FlagInitHeight, 1, "specify the initial block height at genesis")
	cmd.Flags().String(cflags.FlagConsensusKeyAlgo, ed25519.KeyType, "algorithm to use for the consensus key")

	return cmd
}

// downloadGenesis downloads the genesis file from a predefined URL and writes it to the genesis file path specified in the config.
// It creates an HTTP client to send a GET request to the genesis file URL. If the request is successful, it reads the response body
// and writes it to the destination genesis file path. If any step in this process fails, it generates the default genesis.
//
// Parameters:
// - config: A pointer to a tmcfg.Config object that contains the configuration, including the genesis file path.
//
// Returns:
// - An error if the download or file writing fails, otherwise nil.
func downloadGenesis(config *cmtcfg.Config, chainID string) error {
	// URL of the genesis file to download
	genesisURL := fmt.Sprintf("https://github.com/virtengine/net/raw/main/%s/genesis.json?download", chainID)

	// Determine the destination path for the genesis file
	genFilePath := config.GenesisFile()

	// Create a new HTTP client with a 30-second timeout
	client := security.NewSecureHTTPClient(security.WithTimeout(30 * time.Second))

	// Create a new GET request
	req, err := http.NewRequest("GET", genesisURL, nil)
	if err != nil {
		return errorsmod.Wrap(err, "failed to create HTTP request for genesis file")
	}

	// Send the request
	fmt.Println("Attempting to download genesis file from", genesisURL)
	resp, err := client.Do(req)
	if err != nil {
		return errorsmod.Wrap(err, "failed to download genesis file")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Check if the HTTP request was successful
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download genesis file: HTTP status %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errorsmod.Wrap(err, "failed to read genesis file response body")
	}

	// Write the body to the destination genesis file
	err = os.WriteFile(genFilePath, body, 0644) //nolint: gosec
	if err != nil {
		return errorsmod.Wrap(err, "failed to write genesis file to destination")
	}

	return nil
}
