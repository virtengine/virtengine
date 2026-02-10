package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	pruningtypes "cosmossdk.io/store/pruning/types"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"cosmossdk.io/log"
	dbm "github.com/cosmos/cosmos-db"
	sdksrv "github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/types"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
)

func openDB(rootDir string, backendType dbm.BackendType) (dbm.DB, error) {
	dataDir := filepath.Join(rootDir, "data")
	return dbm.NewDB("application", backendType, dataDir)
}

func openTraceWriter(traceWriterFile string) (w io.WriteCloser, err error) {
	if traceWriterFile == "" {
		return
	}
	return os.OpenFile( //nolint: gosec
		traceWriterFile,
		os.O_WRONLY|os.O_APPEND|os.O_CREATE,
		0o600,
	)
}

// ListenForQuitSignals listens for SIGINT and SIGTERM. When a signal is received,
// the cleanup function is called, indicating the caller can gracefully exit or
// return.
//
// Note, the blocking behavior of this depends on the block argument.
// The caller must ensure the corresponding context derived from the cancelFn is used correctly.
func ListenForQuitSignals(g *errgroup.Group, block bool, cancelFn context.CancelFunc, logger log.Logger) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	f := func() {
		sig := <-sigCh
		cancelFn()

		logger.Info("caught signal", "signal", sig.String())
	}

	if block {
		g.Go(func() error {
			f()
			return nil
		})
	} else {
		go f()
	}
}

// GetAppDBBackend gets the backend type to use for the application DBs.
func GetAppDBBackend(opts types.AppOptions) dbm.BackendType {
	rv := cast.ToString(opts.Get("app-db-backend"))
	if len(rv) == 0 {
		rv = cast.ToString(opts.Get("db_backend"))
	}

	// Cosmos SDK has migrated to cosmos-db which does not support all the backends which tm-db supported
	if rv == "cleveldb" || rv == "badgerdb" || rv == "boltdb" {
		panic(fmt.Sprintf("invalid app-db-backend %q, use %q, %q, %q instead", rv, dbm.GoLevelDBBackend, dbm.PebbleDBBackend, dbm.RocksDBBackend))
	}

	if len(rv) != 0 {
		return dbm.BackendType(rv)
	}

	return dbm.GoLevelDBBackend
}

// GetServerContextFromCmd returns a Context from a command or an empty Context
// if it has not been set.
func GetServerContextFromCmd(cmd *cobra.Command) *sdksrv.Context {
	if v := cmd.Context().Value(ServerContextKey); v != nil {
		serverCtxPtr := v.(*sdksrv.Context)
		return serverCtxPtr
	}

	return sdksrv.NewDefaultContext()
}

// SetCmdServerContext sets a command's Context value to the provided argument.
// If the context has not been set, set the given context as the default.
func SetCmdServerContext(cmd *cobra.Command, serverCtx *sdksrv.Context) error {
	v := cmd.Context().Value(ServerContextKey)
	if v == nil {
		v = serverCtx
	}

	serverCtxPtr := v.(*sdksrv.Context)
	*serverCtxPtr = *serverCtx

	return nil
}

// GetPruningOptionsFromFlags parses command flags and returns the correct
// PruningOptions. If a pruning strategy is provided, that will be parsed and
// returned, otherwise, it is assumed custom pruning options are provided.
func GetPruningOptionsFromFlags(appOpts types.AppOptions) (pruningtypes.PruningOptions, error) {
	strategy := strings.ToLower(cast.ToString(appOpts.Get(cflags.FlagPruning)))

	switch strategy {
	case pruningtypes.PruningOptionDefault, pruningtypes.PruningOptionNothing, pruningtypes.PruningOptionEverything:
		return pruningtypes.NewPruningOptionsFromString(strategy), nil

	case pruningtypes.PruningOptionCustom:
		opts := pruningtypes.NewCustomPruningOptions(
			cast.ToUint64(appOpts.Get(cflags.FlagPruningKeepRecent)),
			cast.ToUint64(appOpts.Get(cflags.FlagPruningInterval)),
		)

		if err := opts.Validate(); err != nil {
			return opts, fmt.Errorf("invalid custom pruning options: %w", err)
		}

		return opts, nil

	default:
		return pruningtypes.PruningOptions{}, fmt.Errorf("unknown pruning strategy %s", strategy)
	}
}
