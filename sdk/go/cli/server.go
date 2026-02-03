package cli

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"runtime/pprof"
	"time"

	"github.com/hashicorp/go-metrics"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	cmtserver "github.com/cometbft/cometbft/abci/server"
	cmtcmd "github.com/cometbft/cometbft/cmd/cometbft/commands"
	cmtcfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	pvm "github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cometbft/cometbft/rpc/client/local"

	pruningtypes "cosmossdk.io/store/pruning/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdksrv "github.com/cosmos/cosmos-sdk/server"
	srvapi "github.com/cosmos/cosmos-sdk/server/api"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	servergrpc "github.com/cosmos/cosmos-sdk/server/grpc"
	servercmtlog "github.com/cosmos/cosmos-sdk/server/log"
	"github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	"github.com/cosmos/cosmos-sdk/types/mempool"
	"github.com/cosmos/cosmos-sdk/version"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
)

// StartCmdOptions defines options that can be customized in `StartCmdWithOptions`,
type StartCmdOptions struct {
	// DBOpener can be used to customize db opening, for example, customize db options or support different db backends,
	// default to the builtin db opener.
	DBOpener func(rootDir string, backendType dbm.BackendType) (dbm.DB, error)
	// PostSetup can be used to set up extra services under the same cancellable context;
	// it's not called in stand-alone mode, only for in-process mode.
	PostSetup func(svrCtx *sdksrv.Context, clientCtx client.Context, ctx context.Context, g *errgroup.Group) error
	// PostSetupStandalone can be used to setup extra services under the same cancellable context,
	PostSetupStandalone func(svrCtx *sdksrv.Context, clientCtx client.Context, ctx context.Context, g *errgroup.Group) error
	// AddFlags add custom flags to start cmd
	AddFlags func(cmd *cobra.Command)
	// StartCommandHandler can be used to customize the start command handler
	StartCommandHandler func(svrCtx *sdksrv.Context, clientCtx client.Context, appCreator types.AppCreator, inProcessConsensus bool, opts StartCmdOptions) error
}

// StartCmd runs the service passed in, either stand-alone or in-process with
// CometBFT.
func StartCmd(appCreator types.AppCreator, defaultNodeHome string) *cobra.Command {
	return StartCmdWithOptions(appCreator, defaultNodeHome, StartCmdOptions{})
}

// StartCmdWithOptions runs the service passed in, either stand-alone or in-process with
// CometBFT.
func StartCmdWithOptions(appCreator types.AppCreator, defaultNodeHome string, opts StartCmdOptions) *cobra.Command {
	if opts.DBOpener == nil {
		opts.DBOpener = openDB
	}

	if opts.StartCommandHandler == nil {
		opts.StartCommandHandler = start
	}

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Run the full node",
		Long: `Run the full node application with CometBFT in or out of process. By
default, the application will run with CometBFT in process.

Pruning options can be provided via the '--pruning' flag or alternatively with '--pruning-keep-recent', and
'pruning-interval' together.

For '--pruning' the options are as follows:

default: the last 362880 states are kept, pruning at 10 block intervals
nothing: all historic states will be saved, nothing will be deleted (i.e. archiving node)
everything: 2 latest states will be kept; pruning at 10 block intervals.
custom: allow pruning options to be manually specified through 'pruning-keep-recent', and 'pruning-interval'

Node halting configurations exist in the form of two flags: '--halt-height' and '--halt-time'. During
the ABCI Commit phase, the node will check if the current block height is greater than or equal to
the halt-height or if the current block time is greater than or equal to the halt-time. If so, the
node will attempt to gracefully shutdown and the block will not be committed. In addition, the node
will not be able to commit subsequent blocks.

For profiling and benchmarking purposes, CPU profiling can be enabled via the '--cpu-profile' flag
which accepts a path for the resulting pprof file.

The node may be started in a 'query only' mode where only the gRPC and JSON HTTP
API services are enabled via the 'grpc-only' flag. In this mode, CometBFT is
bypassed and can be used when legacy queries are needed after an on-chain upgrade
is performed. Note, when enabled, gRPC will also be automatically enabled.
`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			sctx := GetServerContextFromCmd(cmd)

			_, err := GetPruningOptionsFromFlags(sctx.Viper)
			if err != nil {
				return err
			}

			cctx, err := GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			withCMT, _ := cmd.Flags().GetBool(cflags.FlagWithComet)
			if !withCMT {
				sctx.Logger.Info("starting ABCI without CometBFT")
			}

			err = wrapCPUProfile(sctx, func() error {
				return opts.StartCommandHandler(sctx, cctx, appCreator, withCMT, opts)
			})

			sctx.Logger.Debug("received quit signal")
			graceDuration, _ := cmd.Flags().GetDuration(cflags.FlagShutdownGrace)
			if graceDuration > 0 {
				sctx.Logger.Info("graceful shutdown start", cflags.FlagShutdownGrace, graceDuration)
				<-time.After(graceDuration)
				sctx.Logger.Info("graceful shutdown complete")
			}

			return err
		},
	}

	cmd.Flags().String(cflags.FlagHome, defaultNodeHome, "The application home directory")
	addStartNodeFlags(cmd, opts)
	return cmd
}

func start(sctx *sdksrv.Context, cctx client.Context, appCreator types.AppCreator, withCmt bool, opts StartCmdOptions) error {
	scfg, err := getAndValidateConfig(sctx)
	if err != nil {
		return err
	}

	app, appCleanupFn, err := startApp(sctx, appCreator, opts)
	if err != nil {
		return err
	}
	defer appCleanupFn()

	metricsSrv, err := startTelemetry(scfg)
	if err != nil {
		return err
	}

	emitServerInfoMetrics()

	if !withCmt {
		return startStandAlone(sctx, scfg, cctx, app, metricsSrv, opts)
	}

	return startInProcess(sctx, scfg, cctx, app, metricsSrv, opts)
}

func startStandAlone(sctx *sdksrv.Context, svrCfg serverconfig.Config, cctx client.Context, app types.Application, metrics *telemetry.Metrics, opts StartCmdOptions) error {
	addr := sctx.Viper.GetString(cflags.FlagAddress)
	transport := sctx.Viper.GetString(cflags.FlagTransport)

	cmtApp := NewCometABCIWrapper(app)
	svr, err := cmtserver.NewServer(addr, transport, cmtApp)
	if err != nil {
		return fmt.Errorf("error creating listener: %w", err)
	}

	svr.SetLogger(servercmtlog.CometLoggerWrapper{Logger: sctx.Logger.With("module", "abci-server")})

	g, ctx := getCtx(sctx, false)

	// Add the tx service to the gRPC router. We only need to register this
	// service if API or gRPC is enabled, and avoid doing so in the general
	// case, because it spawns a new local CometBFT RPC client.
	if svrCfg.API.Enable || svrCfg.GRPC.Enable {
		// create tendermint client
		// assumes the rpc listen address is where tendermint has its rpc server
		rpcclient, err := rpchttp.New(sctx.Config.RPC.ListenAddress, "/websocket")
		if err != nil {
			return err
		}
		// re-assign for making the client available below
		// do not use := to avoid shadowing clientCtx
		cctx = cctx.WithClient(rpcclient)

		// use the provided clientCtx to register the services
		app.RegisterTxService(cctx)
		app.RegisterTendermintService(cctx)
		app.RegisterNodeService(cctx, svrCfg)
	}

	grpcSrv, cctx, err := startGrpcServer(ctx, g, svrCfg.GRPC, cctx, sctx, app)
	if err != nil {
		return err
	}

	err = startAPIServer(ctx, g, svrCfg, cctx, sctx, app, sctx.Config.RootDir, grpcSrv, metrics)
	if err != nil {
		return err
	}

	if opts.PostSetupStandalone != nil {
		if err := opts.PostSetupStandalone(sctx, cctx, ctx, g); err != nil {
			return err
		}
	}

	g.Go(func() error {
		if err := svr.Start(); err != nil {
			sctx.Logger.Error("failed to start out-of-process ABCI server", "err", err)
			return err
		}

		// Wait for the calling process to be canceled or close the provided context,
		// so we can gracefully stop the ABCI server.
		<-ctx.Done()
		sctx.Logger.Info("stopping the ABCI server...")
		return svr.Stop()
	})

	return g.Wait()
}

func startInProcess(sctx *sdksrv.Context, scfg serverconfig.Config, cctx client.Context, app types.Application,
	metrics *telemetry.Metrics, opts StartCmdOptions,
) error {
	cmtCfg := sctx.Config
	gRPCOnly := sctx.Viper.GetBool(cflags.FlagGRPC)

	g, ctx := getCtx(sctx, true)

	if gRPCOnly {
		// TODO: Generalize logic so that gRPC only is really in startStandAlone
		sctx.Logger.Info("starting node in gRPC only mode; CometBFT is disabled")
		scfg.GRPC.Enable = true
	} else {
		sctx.Logger.Info("starting node with ABCI CometBFT in-process")
		tmNode, cleanupFn, err := startCmtNode(ctx, cmtCfg, app, sctx)
		if err != nil {
			return err
		}
		defer cleanupFn()

		// Add the tx service to the gRPC router. We only need to register this
		// service if API or gRPC is enabled, and avoid doing so in the general
		// case, because it spawns a new local CometBFT RPC client.
		if scfg.API.Enable || scfg.GRPC.Enable {
			// Re-assign for making the client available below do not use := to avoid
			// shadowing the clientCtx variable.
			cctx = cctx.WithClient(local.New(tmNode))

			app.RegisterTxService(cctx)
			app.RegisterTendermintService(cctx)
			app.RegisterNodeService(cctx, scfg)
		}
	}

	grpcSrv, cctx, err := startGrpcServer(ctx, g, scfg.GRPC, cctx, sctx, app)
	if err != nil {
		return err
	}

	err = startAPIServer(ctx, g, scfg, cctx, sctx, app, cmtCfg.RootDir, grpcSrv, metrics)
	if err != nil {
		return err
	}

	if opts.PostSetup != nil {
		if err := opts.PostSetup(sctx, cctx, ctx, g); err != nil {
			return err
		}
	}

	// wait for signal capture and gracefully return
	// we are guaranteed to be waiting for the "ListenForQuitSignals" goroutine.
	return g.Wait()
}

// TODO: Move nodeKey into being created within the function.
func startCmtNode(
	ctx context.Context,
	cfg *cmtcfg.Config,
	app types.Application,
	sctx *sdksrv.Context,
) (tmNode *node.Node, cleanupFn func(), err error) {
	nodeKey, err := p2p.LoadOrGenNodeKey(cfg.NodeKeyFile())
	if err != nil {
		return nil, cleanupFn, err
	}

	cmtApp := NewCometABCIWrapper(app)
	tmNode, err = node.NewNodeWithContext(
		ctx,
		cfg,
		pvm.LoadOrGenFilePV(cfg.PrivValidatorKeyFile(), cfg.PrivValidatorStateFile()),
		nodeKey,
		proxy.NewLocalClientCreator(cmtApp),
		getGenDocProvider(cfg),
		cmtcfg.DefaultDBProvider,
		node.DefaultMetricsProvider(cfg.Instrumentation),
		servercmtlog.CometLoggerWrapper{Logger: sctx.Logger},
	)
	if err != nil {
		return tmNode, cleanupFn, err
	}

	if err := tmNode.Start(); err != nil {
		return tmNode, cleanupFn, err
	}

	cleanupFn = func() {
		if tmNode != nil && tmNode.IsRunning() {
			_ = tmNode.Stop()
		}
	}

	return tmNode, cleanupFn, nil
}

func getAndValidateConfig(sctx *sdksrv.Context) (serverconfig.Config, error) {
	config, err := serverconfig.GetConfig(sctx.Viper)
	if err != nil {
		return config, err
	}

	if err := config.ValidateBasic(); err != nil {
		return config, err
	}
	return config, nil
}

// returns a function which returns the genesis doc from the genesis file.
func getGenDocProvider(cfg *cmtcfg.Config) func() (node.ChecksummedGenesisDoc, error) {
	return func() (node.ChecksummedGenesisDoc, error) {
		appGenesis, err := genutiltypes.AppGenesisFromFile(cfg.GenesisFile())
		if err != nil {
			return node.ChecksummedGenesisDoc{}, err
		}

		genDoc, err := appGenesis.ToGenesisDoc()
		if err != nil {
			return node.ChecksummedGenesisDoc{}, err
		}
		return node.ChecksummedGenesisDoc{
			GenesisDoc:     genDoc,
			Sha256Checksum: appGenesis.Sha256Checksum,
		}, nil
	}
}

func setupTraceWriter(svrCtx *sdksrv.Context) (traceWriter io.WriteCloser, cleanup func(), err error) {
	// clean up the traceWriter when the server is shutting down
	cleanup = func() {}

	traceWriterFile := svrCtx.Viper.GetString(cflags.FlagTraceStore)
	traceWriter, err = openTraceWriter(traceWriterFile)
	if err != nil {
		return traceWriter, cleanup, err
	}

	// if flagTraceStore is not used, then traceWriter is nil
	if traceWriter != nil {
		cleanup = func() {
			if err = traceWriter.Close(); err != nil {
				svrCtx.Logger.Error("failed to close trace writer", "err", err)
			}
		}
	}

	return traceWriter, cleanup, nil
}

func startGrpcServer(
	ctx context.Context,
	g *errgroup.Group,
	config serverconfig.GRPCConfig,
	cctx client.Context,
	sctx *sdksrv.Context,
	app types.Application,
) (*grpc.Server, client.Context, error) {
	if !config.Enable {
		// return grpcServer as nil if gRPC is disabled
		return nil, cctx, nil
	}
	_, _, err := net.SplitHostPort(config.Address)
	if err != nil {
		return nil, cctx, err
	}

	maxSendMsgSize := config.MaxSendMsgSize
	if maxSendMsgSize == 0 {
		maxSendMsgSize = serverconfig.DefaultGRPCMaxSendMsgSize
	}

	maxRecvMsgSize := config.MaxRecvMsgSize
	if maxRecvMsgSize == 0 {
		maxRecvMsgSize = serverconfig.DefaultGRPCMaxRecvMsgSize
	}

	// if gRPC is enabled, configure gRPC client for gRPC gateway
	grpcClient, err := grpc.NewClient(
		config.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.ForceCodec(codec.NewProtoCodec(cctx.InterfaceRegistry).GRPCCodec()),
			grpc.MaxCallRecvMsgSize(maxRecvMsgSize),
			grpc.MaxCallSendMsgSize(maxSendMsgSize),
		))
	if err != nil {
		return nil, cctx, err
	}

	cctx = cctx.WithGRPCClient(grpcClient)
	sctx.Logger.Debug("gRPC client assigned to client context", "target", config.Address)

	grpcSrv, err := servergrpc.NewGRPCServer(cctx, app, config)
	if err != nil {
		return nil, cctx, err
	}

	// Start the gRPC server in a goroutine. Note, the provided ctx will ensure
	// that the server is gracefully shut down.
	g.Go(func() error {
		return servergrpc.StartGRPCServer(ctx, sctx.Logger.With("module", "grpc-server"), config, grpcSrv)
	})
	return grpcSrv, cctx, nil
}

func startAPIServer(
	ctx context.Context,
	g *errgroup.Group,
	svrCfg serverconfig.Config,
	cctx client.Context,
	sctx *sdksrv.Context,
	app types.Application,
	home string,
	grpcSrv *grpc.Server,
	metrics *telemetry.Metrics,
) error {
	if !svrCfg.API.Enable {
		return nil
	}

	cctx = cctx.WithHomeDir(home)

	apiSrv := srvapi.New(cctx, sctx.Logger.With("module", "api-server"), grpcSrv)
	app.RegisterAPIRoutes(apiSrv, svrCfg.API)

	if svrCfg.Telemetry.Enabled {
		apiSrv.SetTelemetry(metrics)
	}

	g.Go(func() error {
		return apiSrv.Start(ctx, svrCfg)
	})
	return nil
}

func startTelemetry(cfg serverconfig.Config) (*telemetry.Metrics, error) {
	return telemetry.New(cfg.Telemetry)
}

// wrapCPUProfile starts CPU profiling, if enabled, and executes the provided
// callbackFn in a separate goroutine, then will wait for that callback to
// return.
//
// NOTE: We expect the caller to handle graceful shutdown and signal handling.
func wrapCPUProfile(sctx *sdksrv.Context, callbackFn func() error) error {
	if cpuProfile := sctx.Viper.GetString(cflags.FlagCPUProfile); cpuProfile != "" {
		f, err := os.Create(cpuProfile) //nolint: gosec
		if err != nil {
			return err
		}

		sctx.Logger.Info("starting CPU profiler", "profile", cpuProfile)

		if err := pprof.StartCPUProfile(f); err != nil {
			return err
		}

		defer func() {
			sctx.Logger.Info("stopping CPU profiler", "profile", cpuProfile)
			pprof.StopCPUProfile()

			if err := f.Close(); err != nil {
				sctx.Logger.Info("failed to close cpu-profile file", "profile", cpuProfile, "err", err.Error())
			}
		}()
	}

	return callbackFn()
}

// emitServerInfoMetrics emits server info related metrics using application telemetry.
func emitServerInfoMetrics() {
	var ls []metrics.Label

	versionInfo := version.NewInfo()
	if len(versionInfo.GoVersion) > 0 {
		ls = append(ls, telemetry.NewLabel("go", versionInfo.GoVersion))
	}
	if len(versionInfo.CosmosSdkVersion) > 0 {
		ls = append(ls, telemetry.NewLabel("version", versionInfo.CosmosSdkVersion))
	}

	if len(ls) == 0 {
		return
	}

	telemetry.SetGaugeWithLabels([]string{"server", "info"}, 1, ls)
}

func getCtx(sctx *sdksrv.Context, block bool) (*errgroup.Group, context.Context) {
	ctx, cancelFn := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)
	// listen for quit signals so the calling parent process can gracefully exit
	ListenForQuitSignals(g, block, cancelFn, sctx.Logger)
	return g, ctx
}

func startApp(sctx *sdksrv.Context, appCreator types.AppCreator, opts StartCmdOptions) (app types.Application, cleanupFn func(), err error) {
	traceWriter, traceCleanupFn, err := setupTraceWriter(sctx)
	if err != nil {
		return app, traceCleanupFn, err
	}

	home := sctx.Config.RootDir
	db, err := opts.DBOpener(home, GetAppDBBackend(sctx.Viper))
	if err != nil {
		return app, traceCleanupFn, err
	}

	app = appCreator(sctx.Logger, db, traceWriter, sctx.Viper)

	cleanupFn = func() {
		traceCleanupFn()
		if localErr := app.Close(); localErr != nil {
			sctx.Logger.Error(localErr.Error())
		}
	}
	return app, cleanupFn, nil
}

// addStartNodeFlags should be added to any CLI commands that start the network.
func addStartNodeFlags(cmd *cobra.Command, opts StartCmdOptions) {
	cmd.Flags().Bool(cflags.FlagWithComet, true, "Run abci app embedded in-process with CometBFT")
	cmd.Flags().String(cflags.FlagAddress, "tcp://127.0.0.1:26658", "Listen address")
	cmd.Flags().String(cflags.FlagTransport, "socket", "Transport protocol: socket, grpc")
	cmd.Flags().String(cflags.FlagTraceStore, "", "Enable KVStore tracing to an output file")
	cmd.Flags().String(cflags.FlagMinGasPrices, "", "Minimum gas prices to accept for transactions; Any fee in a tx must meet this minimum (e.g. 0.01photino;0.0001stake)")
	cmd.Flags().Uint64(cflags.FlagQueryGasLimit, 0, "Maximum gas a Rest/Grpc query can consume. Blank and 0 imply unbounded.")
	cmd.Flags().IntSlice(cflags.FlagUnsafeSkipUpgrades, []int{}, "Skip a set of upgrade heights to continue the old binary")
	cmd.Flags().Uint64(cflags.FlagHaltHeight, 0, "Block height at which to gracefully halt the chain and shutdown the node")
	cmd.Flags().Uint64(cflags.FlagHaltTime, 0, "Minimum block time (in Unix seconds) at which to gracefully halt the chain and shutdown the node")
	cmd.Flags().Bool(cflags.FlagInterBlockCache, true, "Enable inter-block caching")
	cmd.Flags().String(cflags.FlagCPUProfile, "", "Enable CPU profiling and write to the provided file")
	cmd.Flags().Bool(cflags.FlagTrace, false, "Provide full stack traces for errors in ABCI Log")
	cmd.Flags().String(cflags.FlagPruning, pruningtypes.PruningOptionDefault, "Pruning strategy (default|nothing|everything|custom)")
	cmd.Flags().Uint64(cflags.FlagPruningKeepRecent, 0, "Number of recent heights to keep on disk (ignored if pruning is not 'custom')")
	cmd.Flags().Uint64(cflags.FlagPruningInterval, 0, "Height interval at which pruned heights are removed from disk (ignored if pruning is not 'custom')")
	cmd.Flags().Uint(cflags.FlagInvCheckPeriod, 0, "Assert registered invariants every N blocks")
	cmd.Flags().Uint64(cflags.FlagMinRetainBlocks, 0, "Minimum block height offset during ABCI commit to prune CometBFT blocks")
	cmd.Flags().Bool(cflags.FlagAPIEnable, false, "Define if the API server should be enabled")
	cmd.Flags().Bool(cflags.FlagAPISwagger, false, "Define if swagger documentation should automatically be registered (Note: the API must also be enabled)")
	cmd.Flags().String(cflags.FlagAPIAddress, serverconfig.DefaultAPIAddress, "the API server address to listen on")
	cmd.Flags().Uint(cflags.FlagAPIMaxOpenConnections, 1000, "Define the number of maximum open connections")
	cmd.Flags().Uint(cflags.FlagRPCReadTimeout, 10, "Define the CometBFT RPC read timeout (in seconds)")
	cmd.Flags().Uint(cflags.FlagRPCWriteTimeout, 0, "Define the CometBFT RPC write timeout (in seconds)")
	cmd.Flags().Uint(cflags.FlagRPCMaxBodyBytes, 1000000, "Define the CometBFT maximum request body (in bytes)")
	cmd.Flags().Bool(cflags.FlagAPIEnableUnsafeCORS, false, "Define if CORS should be enabled (unsafe - use it at your own risk)")
	cmd.Flags().Bool(cflags.FlagGRPCOnly, false, "Start the node in gRPC query only mode (no CometBFT process is started)")
	cmd.Flags().Bool(cflags.FlagGRPCEnable, true, "Define if the gRPC server should be enabled")
	cmd.Flags().String(cflags.FlagGRPCAddress, serverconfig.DefaultGRPCAddress, "the gRPC server address to listen on")
	cmd.Flags().Bool(cflags.FlagGRPCWebEnable, true, "Define if the gRPC-Web server should be enabled. (Note: gRPC must also be enabled)")
	cmd.Flags().Uint64(cflags.FlagStateSyncSnapshotInterval, 0, "State sync snapshot interval")
	cmd.Flags().Uint32(cflags.FlagStateSyncSnapshotKeepRecent, 2, "State sync snapshot to keep")
	cmd.Flags().Bool(cflags.FlagDisableIAVLFastNode, false, "Disable fast node for IAVL tree")
	cmd.Flags().Int(cflags.FlagMempoolMaxTxs, mempool.DefaultMaxTx, "Sets MaxTx value for the app-side mempool")
	cmd.Flags().Duration(cflags.FlagShutdownGrace, 0*time.Second, "On Shutdown, duration to wait for resource clean up")

	// support old flags name for backwards compatibility
	cmd.Flags().SetNormalizeFunc(func(_ *pflag.FlagSet, name string) pflag.NormalizedName {
		if name == cflags.FlagWithTendermint {
			name = cflags.FlagWithComet
		}

		return pflag.NormalizedName(name)
	})

	// add support for all CometBFT-specific command line options
	cmtcmd.AddNodeFlags(cmd)

	if opts.AddFlags != nil {
		opts.AddFlags(cmd)
	}
}
