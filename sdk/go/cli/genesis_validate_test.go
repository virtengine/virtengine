package cli_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/stretchr/testify/require"

	clitestutil "github.com/virtengine/virtengine/sdk/go/cli/testutil"
)

// An example exported genesis file from a 0.37 chain. Note that evidence
// parameters only contains `max_age`.
var v037Exported = `{
	"app_hash": "",
	"app_state": {},
	"chain_id": "test",
	"consensus_params": {
		"block": {
		"max_bytes": "22020096",
		"max_gas": "-1",
		"time_iota_ms": "1000"
		},
		"evidence": { "max_age": "100000" },
		"validator": { "pub_key_types": ["ed25519"] }
	},
	"genesis_time": "2020-09-29T20:16:29.172362037Z",
	"validators": []
}`

func (s *GenesisCLITestSuite) TestValidateGenesis() {
	testCases := []struct {
		name         string
		genesis      string
		expErrStr    string
		basicManager module.BasicManager
	}{
		{
			"invalid json",
			`{"app_state": {x,}}`,
			"error at offset 16: invalid character",
			module.NewBasicManager(),
		},
		{
			"invalid: missing module config in app_state",
			func() string {
				bz, err := os.ReadFile("./testdata/app_genesis.json")
				require.NoError(s.T(), err)

				return string(bz)
			}(),
			"section is missing in the app_state",
			module.NewBasicManager(mockModule{}),
		},
		{
			"exported 0.37 genesis file",
			v037Exported,
			"make sure that you have correctly migrated all CometBFT consensus params",
			module.NewBasicManager(),
		},
		{
			"valid 0.50 genesis file",
			func() string {
				bz, err := os.ReadFile("./testdata/app_genesis.json")
				require.NoError(s.T(), err)

				return string(bz)
			}(),
			"",
			module.NewBasicManager(),
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			genesisFile := testutil.WriteToNewTempFile(t, tc.genesis)
			_, err := clitestutil.ExecTestCLICmd(context.Background(), client.Context{}, cli.ValidateGenesisCmd(tc.basicManager), genesisFile.Name())
			if tc.expErrStr != "" {
				require.Contains(t, err.Error(), tc.expErrStr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

type mockModule struct {
	module.AppModuleBasic
}

func (m mockModule) Name() string {
	return "mock"
}

func (m mockModule) DefaultGenesis(codec.JSONCodec) json.RawMessage {
	return json.RawMessage(`{"foo": "bar"}`)
}

func (m mockModule) ValidateGenesis(codec.JSONCodec, client.TxEncodingConfig, json.RawMessage) error {
	return fmt.Errorf("mock section is missing: %w", io.EOF)
}
