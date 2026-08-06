package cmd

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/stretchr/testify/require"
	"github.com/virtengine/virtengine/app"
	"github.com/virtengine/virtengine/sdk/go/sdkutil"
)

func TestPrepareGenesisSetsConfiguredCommunityPool(t *testing.T) {
	encodingConfig := sdkutil.MakeEncodingConfig()
	app.ModuleBasics().RegisterInterfaces(encodingConfig.InterfaceRegistry)

	clientCtx := client.Context{}.
		WithCodec(encodingConfig.Codec).
		WithTxConfig(encodingConfig.TxConfig).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry)

	genesisState := app.NewDefaultGenesisState(encodingConfig.Codec)
	appStateJSON, err := json.Marshal(genesisState)
	require.NoError(t, err)

	genDoc := &genutiltypes.AppGenesis{
		ChainID:  "virtengine-1",
		AppState: appStateJSON,
	}

	genesisParams := MainnetGenesisParams()
	appState, _, err := PrepareGenesis(clientCtx, genesisState, genDoc, genesisParams, "virtengine-1")
	require.NoError(t, err)

	var distributionState distributiontypes.GenesisState
	require.NoError(t, encodingConfig.Codec.UnmarshalJSON(appState[distributiontypes.ModuleName], &distributionState))
	require.Equal(t, genesisParams.InitialCommunityPool, distributionState.FeePool.CommunityPool)
	require.Equal(t, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), genDoc.GenesisTime)
}

func TestModuleAddressCmdUsesVirtEngineBech32Prefix(t *testing.T) {
	var output bytes.Buffer

	cmd := ModuleAddressCmd()
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"distribution"})

	require.NoError(t, cmd.Execute())
	require.Equal(t, "ve1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8mzlgxh\n", output.String())
}
