package cli_test

import (
	"context"
	"fmt"
	"path/filepath"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec/address"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/virtengine/virtengine/sdk/go/cli"
	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	"github.com/virtengine/virtengine/sdk/go/sdkutil"
)

func (s *GenesisCLITestSuite) TestGenTxCmd() {
	amount := sdk.NewCoin("uakt", sdkmath.NewInt(12))

	tests := []struct {
		name         string
		args         []string
		expCmdOutput string
	}{
		{
			name: "invalid commission rate returns error",
			args: []string{
				fmt.Sprintf("--%s=%s", cflags.FlagChainID, s.baseCtx.ChainID),
				fmt.Sprintf("--%s=1", cflags.FlagCommissionRate),
				"node0",
				amount.String(),
			},
			expCmdOutput: fmt.Sprintf("--%s=%s --%s=1 %s %s", cflags.FlagChainID, s.baseCtx.ChainID, cflags.FlagCommissionRate, "node0", amount.String()),
		},
		{
			name: "valid gentx",
			args: []string{
				fmt.Sprintf("--%s=%s", cflags.FlagChainID, s.baseCtx.ChainID),
				"node0",
				amount.String(),
			},
			expCmdOutput: fmt.Sprintf("--%s=%s %s %s", cflags.FlagChainID, s.baseCtx.ChainID, "node0", amount.String()),
		},
		{
			name: "invalid pubkey",
			args: []string{
				fmt.Sprintf("--%s=%s", cflags.FlagChainID, "test-chain-1"),
				fmt.Sprintf("--%s={\"key\":\"BOIkjkFruMpfOFC9oNPhiJGfmY2pHF/gwHdLDLnrnS0=\"}", cflags.FlagPubKey),
				"node0",
				amount.String(),
			},
			expCmdOutput: fmt.Sprintf("--%s=test-chain-1 --%s={\"key\":\"BOIkjkFruMpfOFC9oNPhiJGfmY2pHF/gwHdLDLnrnS0=\"} %s %s ", cflags.FlagChainID, cflags.FlagPubKey, "node0", amount.String()),
		},
		{
			name: "valid pubkey flag",
			args: []string{
				fmt.Sprintf("--%s=%s", cflags.FlagChainID, "test-chain-1"),
				fmt.Sprintf("--%s={\"@type\":\"/cosmos.crypto.ed25519.PubKey\",\"key\":\"BOIkjkFruMpfOFC9oNPhiJGfmY2pHF/gwHdLDLnrnS0=\"}", cflags.FlagPubKey),
				"node0",
				amount.String(),
			},
			expCmdOutput: fmt.Sprintf("--%s=test-chain-1 --%s={\"@type\":\"/cosmos.crypto.ed25519.PubKey\",\"key\":\"BOIkjkFruMpfOFC9oNPhiJGfmY2pHF/gwHdLDLnrnS0=\"} %s %s ", cflags.FlagChainID, cflags.FlagPubKey, "node0", amount.String()),
		},
	}

	for _, tc := range tests {
		dir := s.T().TempDir()
		genTxFile := filepath.Join(dir, "myTx")

		tc.args = append(tc.args, fmt.Sprintf("--%s=%s", cflags.FlagOutputDocument, genTxFile))

		s.Run(tc.name, func() {
			cctx := s.cctx
			ctx := svrcmd.CreateExecuteContext(context.Background())

			cmd := cli.GetGenesisGenTxCmd(
				module.NewBasicManager(),
				cctx.TxConfig,
				banktypes.GenesisBalancesIterator{},
				cctx.HomeDir,
				address.NewBech32Codec(sdkutil.Bech32PrefixValAddr),
			)
			cmd.SetContext(ctx)
			cmd.SetArgs(tc.args)

			s.Require().NoError(client.SetCmdClientContextHandler(cctx, cmd))

			if len(tc.args) != 0 {
				s.Require().Contains(fmt.Sprint(cmd), tc.expCmdOutput)
			}
		})
	}
}
