package cli_test

import (
	"encoding/hex"
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm/ioutils"
	"github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/sdk/go/cli"
	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	testdata "github.com/virtengine/virtengine/sdk/go/cli/testdata/wasm"
)

func TestParseVerificationFlags(t *testing.T) {
	mySender := sdk.MustAccAddressFromBech32("akash1wyqh3n50ecatjg4vww5crmtd0nmyzusnwckw4at4gluc0m5m477qgz4fa7")

	specs := map[string]struct {
		srcPath     string
		args        []string
		expErr      bool
		expSource   string
		expBuilder  string
		expCodeHash string
	}{
		"gov store zipped": {
			srcPath: "./testdata/wasm/hackatom.wasm.gzip",
			args: cli.TestFlags().
				WithFlag(cflags.FlagInstantiateByEverybody, true).
				WithFlag(cflags.FlagCodeHash, testdata.ChecksumHackatom).
				WithFlag(cflags.FlagSource, "https://example.com").
				WithFlag(cflags.FlagBuilder, "cosmwasm/workspace-optimizer:0.12.11"),
			expBuilder:  "cosmwasm/workspace-optimizer:0.12.11",
			expSource:   "https://example.com",
			expCodeHash: testdata.ChecksumHackatom,
		},
		"gov store raw": {
			srcPath: "./testdata/wasm/hackatom.wasm",
			args: cli.TestFlags().
				WithFlag(cflags.FlagInstantiateByEverybody, true).
				WithFlag(cflags.FlagCodeHash, testdata.ChecksumHackatom).
				WithFlag(cflags.FlagSource, "https://example.com").
				WithFlag(cflags.FlagBuilder, "cosmwasm/workspace-optimizer:0.12.11"),
			expBuilder:  "cosmwasm/workspace-optimizer:0.12.11",
			expSource:   "https://example.com",
			expCodeHash: testdata.ChecksumHackatom,
		},
		"gov store checksum mismatch": {
			srcPath: "./testdata/wasm/hackatom.wasm",
			args: cli.TestFlags().
				WithFlag(cflags.FlagInstantiateByEverybody, true).
				WithFlag(cflags.FlagCodeHash, "0000de5e9b93b52e514c74ce87ccddb594b9bcd33b7f1af1bb6da63fc883917b").
				WithFlag(cflags.FlagSource, "https://example.com").
				WithFlag(cflags.FlagBuilder, "cosmwasm/workspace-optimizer:0.12.11"),
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			flagSet := cli.GetTxGovWasmProposalStoreAndInstantiateContractCmd().Flags()
			require.NoError(t, flagSet.Parse(spec.args))

			gotMsg, err := cli.ParseWasmStoreCodeArgs(spec.srcPath, mySender.String(), flagSet)
			require.NoError(t, err)
			require.True(t, ioutils.IsGzip(gotMsg.WASMByteCode))

			gotSource, gotBuilder, gotCodeHash, gotErr := cli.ParseWasmVerificationFlags(gotMsg.WASMByteCode, flagSet)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.expSource, gotSource)
			assert.Equal(t, spec.expBuilder, gotBuilder)
			assert.Equal(t, spec.expCodeHash, hex.EncodeToString(gotCodeHash))
		})
	}
}

func TestParseAccessConfigFlags(t *testing.T) {
	specs := map[string]struct {
		args   []string
		expCfg *types.AccessConfig
		expErr bool
	}{
		"nobody": {
			args:   []string{"--instantiate-nobody=true"},
			expCfg: &types.AccessConfig{Permission: types.AccessTypeNobody},
		},
		"everybody": {
			args:   []string{"--instantiate-everybody=true"},
			expCfg: &types.AccessConfig{Permission: types.AccessTypeEverybody},
		},
		"only address": {
			args:   []string{"--instantiate-only-address=akash1vx8knpllrj7n963p9ttd80w47kpacrhuxtcz8u"},
			expErr: true,
		},
		"only address - invalid": {
			args:   []string{"--instantiate-only-address=foo"},
			expErr: true,
		},
		"any of address": {
			args:   []string{"--instantiate-anyof-addresses=akash1vx8knpllrj7n963p9ttd80w47kpacrhuxtcz8u,akash14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sggdamt"},
			expCfg: &types.AccessConfig{Permission: types.AccessTypeAnyOfAddresses, Addresses: []string{"akash1vx8knpllrj7n963p9ttd80w47kpacrhuxtcz8u", "akash14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sggdamt"}},
		},
		"any of address - invalid": {
			args:   []string{"--instantiate-anyof-addresses=akash1vx8knpllrj7n963p9ttd80w47kpacrhuxtcz8u,foo"},
			expErr: true,
		},
		"not set": {
			args: []string{},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			flags := cli.GetTxWasmStoreCodeCmd().Flags()
			require.NoError(t, flags.Parse(spec.args))
			gotCfg, gotErr := cli.ParseWasmAccessConfigFlags(flags)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.expCfg, gotCfg)
		})
	}
}

func TestParseStoreCodeGrants(t *testing.T) {
	specs := map[string]struct {
		src    []string
		exp    []types.CodeGrant
		expErr bool
	}{
		"wildcard : nobody": {
			src: []string{"*:nobody"},
			exp: []types.CodeGrant{{
				CodeHash:              []byte("*"),
				InstantiatePermission: &types.AccessConfig{Permission: types.AccessTypeNobody},
			}},
		},
		"wildcard : wildcard": {
			src: []string{"*:*"},
			exp: []types.CodeGrant{{
				CodeHash: []byte("*"),
			}},
		},
		"wildcard : everybody": {
			src: []string{"*:everybody"},
			exp: []types.CodeGrant{{
				CodeHash:              []byte("*"),
				InstantiatePermission: &types.AccessConfig{Permission: types.AccessTypeEverybody},
			}},
		},
		"wildcard : any of addresses - single": {
			src: []string{"*:akash1vx8knpllrj7n963p9ttd80w47kpacrhuxtcz8u"},
			exp: []types.CodeGrant{
				{
					CodeHash: []byte("*"),
					InstantiatePermission: &types.AccessConfig{
						Permission: types.AccessTypeAnyOfAddresses,
						Addresses:  []string{"akash1vx8knpllrj7n963p9ttd80w47kpacrhuxtcz8u"},
					},
				},
			},
		},
		"wildcard : any of addresses - multiple": {
			src: []string{"*:akash1vx8knpllrj7n963p9ttd80w47kpacrhuxtcz8u,akash14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sggdamt"},
			exp: []types.CodeGrant{
				{
					CodeHash: []byte("*"),
					InstantiatePermission: &types.AccessConfig{
						Permission: types.AccessTypeAnyOfAddresses,
						Addresses:  []string{"akash1vx8knpllrj7n963p9ttd80w47kpacrhuxtcz8u", "akash14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sggdamt"},
					},
				},
			},
		},
		"multiple code hashes with different permissions": {
			src: []string{"any_checksum_1:akash1vx8knpllrj7n963p9ttd80w47kpacrhuxtcz8u,akash14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sggdamt", "any_checksum_2:nobody"},
			exp: []types.CodeGrant{
				{
					CodeHash: []byte("any_checksum_1"),
					InstantiatePermission: &types.AccessConfig{
						Permission: types.AccessTypeAnyOfAddresses,
						Addresses:  []string{"akash1vx8knpllrj7n963p9ttd80w47kpacrhuxtcz8u", "akash14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sggdamt"},
					},
				}, {
					CodeHash: []byte("any_checksum_2"),
					InstantiatePermission: &types.AccessConfig{
						Permission: types.AccessTypeNobody,
					},
				},
			},
		},
		"code hash : wildcard": {
			src: []string{"any_checksum_1:*"},
			exp: []types.CodeGrant{{
				CodeHash: []byte("any_checksum_1"),
			}},
		},
		"code hash : any of addresses - empty list": {
			src:    []string{"any_checksum_1:"},
			expErr: true,
		},
		"code hash : any of addresses - invalid address": {
			src:    []string{"any_checksum_1:foo"},
			expErr: true,
		},
		"code hash : any of addresses - duplicate address": {
			src:    []string{"any_checksum_1:akash1vx8knpllrj7n963p9ttd80w47kpacrhuxtcz8u,akash1vx8knpllrj7n963p9ttd80w47kpacrhuxtcz8u"},
			expErr: true,
		},
		"empty code hash": {
			src:    []string{":everyone"},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got, gotErr := cli.ParseStoreCodeGrants(spec.src)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.exp, got)
		})
	}
}
