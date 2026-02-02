package cli_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	abci_server "github.com/cometbft/cometbft/abci/server"

	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/server"
	servercmtlog "github.com/cosmos/cosmos-sdk/server/log"
	"github.com/cosmos/cosmos-sdk/server/mock"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltest "github.com/cosmos/cosmos-sdk/x/genutil/client/testutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/staking"

	vecli "github.com/virtengine/virtengine/sdk/go/cli"
	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
)

// osWindows is the GOOS value for Windows
const osWindows = "windows"

var testMbm = module.NewBasicManager(
	staking.AppModuleBasic{},
	genutil.AppModuleBasic{},
)

func TestInitCmd(t *testing.T) {
	tests := []struct {
		name      string
		flags     func(dir string) []string
		shouldErr bool
		err       error
	}{
		{
			name: "happy path",
			flags: func(_ string) []string {
				return []string{
					"appnode-test",
				}
			},
			shouldErr: false,
			err:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			logger := log.NewNopLogger()
			cfg, err := genutiltest.CreateDefaultCometConfig(home)
			require.NoError(t, err)

			serverCtx := server.NewContext(viper.New(), cfg, logger)
			interfaceRegistry := types.NewInterfaceRegistry()
			marshaler := codec.NewProtoCodec(interfaceRegistry)
			clientCtx := client.Context{}.
				WithCodec(marshaler).
				WithLegacyAmino(makeCodec()).
				WithHomeDir(home)

			ctx := context.Background()
			ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
			ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

			cmd := genutilcli.InitCmd(testMbm, home)
			cmd.SetArgs(
				tt.flags(home),
			)

			if tt.shouldErr {
				err := cmd.ExecuteContext(ctx)
				require.EqualError(t, err, tt.err.Error())
			} else {
				require.NoError(t, cmd.ExecuteContext(ctx))
			}
		})
	}
}

func TestInitRecover(t *testing.T) {
	home := t.TempDir()
	logger := log.NewNopLogger()
	cfg, err := genutiltest.CreateDefaultCometConfig(home)
	require.NoError(t, err)

	serverCtx := server.NewContext(viper.New(), cfg, logger)
	interfaceRegistry := types.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	clientCtx := client.Context{}.
		WithCodec(marshaler).
		WithLegacyAmino(makeCodec()).
		WithHomeDir(home)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
	ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

	cmd := genutilcli.InitCmd(testMbm, home)
	mockIn := testutil.ApplyMockIODiscardOutErr(cmd)

	cmd.SetArgs([]string{
		"appnode-test",
		fmt.Sprintf("--%s=true", genutilcli.FlagRecover),
	})

	// use valid mnemonic and complete recovery key generation successfully
	mockIn.Reset("decide praise business actor peasant farm drastic weather extend front hurt later song give verb rhythm worry fun pond reform school tumble august one\n")
	require.NoError(t, cmd.ExecuteContext(ctx))
}

func TestInitDefaultBondDenom(t *testing.T) {
	home := t.TempDir()
	logger := log.NewNopLogger()
	cfg, err := genutiltest.CreateDefaultCometConfig(home)
	require.NoError(t, err)

	serverCtx := server.NewContext(viper.New(), cfg, logger)
	interfaceRegistry := types.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	clientCtx := client.Context{}.
		WithCodec(marshaler).
		WithLegacyAmino(makeCodec()).
		WithHomeDir(home)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
	ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

	cmd := genutilcli.InitCmd(testMbm, home)

	cmd.SetArgs([]string{
		"appnode-test",
		fmt.Sprintf("--%s=%s", cflags.FlagHome, home),
		fmt.Sprintf("--%s=uve", genutilcli.FlagDefaultBondDenom),
	})
	require.NoError(t, cmd.ExecuteContext(ctx))
}

func TestEmptyState(t *testing.T) {
	// Skip on Windows due to LevelDB file locking issues that prevent
	// TempDir cleanup. This is a known issue with LevelDB on Windows
	// where background goroutines hold file handles longer than expected.
	// See: https://github.com/syndtr/goleveldb/issues/109
	if runtime.GOOS == osWindows {
		t.Skip("Skipping on Windows due to LevelDB file locking in cleanup")
	}

	home := t.TempDir()
	// Prevent Windows file lock cleanup issues with embedded DB handles.
	cfg, err := genutiltest.CreateDefaultCometConfig(home)
	require.NoError(t, err)
	cfg.DBBackend = "memdb"
	logger := log.NewNopLogger()

	serverCtx := server.NewContext(viper.New(), cfg, logger)
	interfaceRegistry := types.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	clientCtx := client.Context{}.
		WithCodec(marshaler).
		WithLegacyAmino(makeCodec()).
		WithHomeDir(home)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
	ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

	cmd := genutilcli.InitCmd(testMbm, home)
	cmd.SetArgs([]string{"appnode-test", fmt.Sprintf("--%s=%s", cflags.FlagHome, home)})

	require.NoError(t, cmd.ExecuteContext(ctx))

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd = vecli.ExportCmd(nil, home)
	cmd.SetArgs([]string{fmt.Sprintf("--%s=%s", cflags.FlagHome, home)})
	require.NoError(t, cmd.ExecuteContext(ctx))

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		require.NoError(t, err)
		outC <- buf.String()
	}()

	require.NoError(t, w.Close())
	os.Stdout = old
	out := <-outC

	// Give time for database to close completely on Windows
	// to avoid file locking issues in TempDir cleanup
	time.Sleep(100 * time.Millisecond)

	require.Contains(t, out, "genesis_time")
	require.Contains(t, out, "chain_id")
	require.Contains(t, out, "consensus")

	require.NoError(t, serverCtx.Config.DB.Close())
	require.Contains(t, out, "app_hash")
	require.Contains(t, out, "app_state")
}

func TestStartStandAlone(t *testing.T) {
	// Skip on Windows due to LevelDB file locking issues in TempDir cleanup
	if runtime.GOOS == osWindows {
		t.Skip("Skipping on Windows due to LevelDB file locking in cleanup")
	}

	home := t.TempDir()
	logger := log.NewNopLogger()
	interfaceRegistry := types.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	err := genutiltest.ExecInitCmd(testMbm, home, marshaler)
	require.NoError(t, err)

	app, err := mock.NewApp(home, logger)
	require.NoError(t, err)
	if closer, ok := app.(io.Closer); ok {
		t.Cleanup(func() { _ = closer.Close() })
	}

	svrAddr, _, closeFn, err := network.FreeTCPAddr()
	require.NoError(t, err)
	require.NoError(t, closeFn())

	cmtApp := server.NewCometABCIWrapper(app)
	svr, err := abci_server.NewServer(svrAddr, "socket", cmtApp)
	require.NoError(t, err, "error creating listener")

	svr.SetLogger(servercmtlog.CometLoggerWrapper{Logger: logger.With("module", "abci-server")})
	err = svr.Start()
	require.NoError(t, err)

	timer := time.NewTimer(time.Duration(2) * time.Second)
	for range timer.C {
		err = svr.Stop()
		require.NoError(t, err)
		break
	}
}

func TestInitNodeValidatorFiles(t *testing.T) {
	home := t.TempDir()
	cfg, err := genutiltest.CreateDefaultCometConfig(home)
	require.NoError(t, err)

	nodeID, valPubKey, err := genutil.InitializeNodeValidatorFiles(cfg)
	require.NoError(t, err)

	require.NotEqual(t, "", nodeID)
	require.NotEqual(t, 0, len(valPubKey.Bytes()))
}

func TestInitConfig(t *testing.T) {
	// Skip on Windows due to LevelDB file locking issues in TempDir cleanup
	if runtime.GOOS == osWindows {
		t.Skip("Skipping on Windows due to LevelDB file locking in cleanup")
	}

	home := t.TempDir()
	logger := log.NewNopLogger()
	cfg, err := genutiltest.CreateDefaultCometConfig(home)
	require.NoError(t, err)

	serverCtx := server.NewContext(viper.New(), cfg, logger)
	interfaceRegistry := types.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	clientCtx := client.Context{}.
		WithCodec(marshaler).
		WithLegacyAmino(makeCodec()).
		WithChainID("foo"). // add chain-id to clientCtx
		WithHomeDir(home)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
	ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

	cmd := genutilcli.InitCmd(testMbm, home)
	cmd.SetArgs([]string{"testnode"})

	require.NoError(t, cmd.ExecuteContext(ctx))

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd = vecli.ExportCmd(nil, home)
	require.NoError(t, cmd.ExecuteContext(ctx))

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		require.NoError(t, err)
		outC <- buf.String()
	}()

	require.NoError(t, w.Close())
	os.Stdout = old
	out := <-outC

	require.Contains(t, out, "\"chain_id\": \"foo\"")
}

func TestInitWithHeight(t *testing.T) {
	home := t.TempDir()
	logger := log.NewNopLogger()
	cfg, err := genutiltest.CreateDefaultCometConfig(home)
	require.NoError(t, err)

	serverCtx := server.NewContext(viper.New(), cfg, logger)
	interfaceRegistry := types.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	clientCtx := client.Context{}.
		WithCodec(marshaler).
		WithLegacyAmino(makeCodec()).
		WithChainID("foo"). // add chain-id to clientCtx
		WithHomeDir(home)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
	ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

	testInitialHeight := int64(333)

	cmd := genutilcli.InitCmd(testMbm, home)
	cmd.SetArgs([]string{"init-height-test", fmt.Sprintf("--%s=%d", cflags.FlagInitHeight, testInitialHeight)})

	require.NoError(t, cmd.ExecuteContext(ctx))

	appGenesis, importErr := genutiltypes.AppGenesisFromFile(cfg.GenesisFile())
	require.NoError(t, importErr)

	require.Equal(t, testInitialHeight, appGenesis.InitialHeight)
}

func TestInitWithNegativeHeight(t *testing.T) {
	home := t.TempDir()
	logger := log.NewNopLogger()
	cfg, err := genutiltest.CreateDefaultCometConfig(home)
	require.NoError(t, err)

	serverCtx := server.NewContext(viper.New(), cfg, logger)
	interfaceRegistry := types.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	clientCtx := client.Context{}.
		WithCodec(marshaler).
		WithLegacyAmino(makeCodec()).
		WithChainID("foo"). // add chain-id to clientCtx
		WithHomeDir(home)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
	ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

	testInitialHeight := int64(-333)

	cmd := genutilcli.InitCmd(testMbm, home)
	cmd.SetArgs([]string{"init-height-test", fmt.Sprintf("--%s=%d", cflags.FlagInitHeight, testInitialHeight)})

	require.NoError(t, cmd.ExecuteContext(ctx))

	appGenesis, importErr := genutiltypes.AppGenesisFromFile(cfg.GenesisFile())
	require.NoError(t, importErr)

	require.Equal(t, int64(1), appGenesis.InitialHeight)
}

// custom tx codec
func makeCodec() *codec.LegacyAmino {
	cdc := codec.NewLegacyAmino()
	sdk.RegisterLegacyAminoCodec(cdc)
	cryptocodec.RegisterCrypto(cdc)
	return cdc
}
