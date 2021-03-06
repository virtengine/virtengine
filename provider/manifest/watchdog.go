package manifest

import (
	"context"
	"github.com/boz/go-lifecycle"
	"github.com/virtengine/virtengine/provider/session"
	dtypes "github.com/virtengine/virtengine/x/deployment/types"
	"github.com/virtengine/virtengine/x/market/types"
	"github.com/tendermint/tendermint/libs/log"
	"time"
)

type watchdog struct {
	leaseID types.LeaseID
	timeout time.Duration
	lc      lifecycle.Lifecycle
	sess    session.Session
	ctx     context.Context
	log     log.Logger
}

func newWatchdog(sess session.Session, parent <-chan struct{}, done chan<- dtypes.DeploymentID, leaseID types.LeaseID, timeout time.Duration) *watchdog {
	ctx, cancel := context.WithCancel(context.Background())
	result := &watchdog{
		leaseID: leaseID,
		timeout: timeout,
		lc:      lifecycle.New(),
		sess:    sess,
		ctx:     ctx,
		log:     sess.Log().With("leaseID", leaseID),
	}

	go func() {
		result.lc.WatchChannel(parent)
		cancel()
	}()

	go func() {
		<-result.lc.Done()
		done <- leaseID.DeploymentID()
	}()

	go result.run()

	return result
}

func (wd *watchdog) stop() {
	wd.lc.ShutdownAsync(nil)
}

func (wd *watchdog) run() {
	defer wd.lc.ShutdownCompleted()

	wd.log.Debug("watchdog start")
	select {
	case <-time.After(wd.timeout):
		// Close the bid, since if this point is reached then a manifest has not been received
	case err := <-wd.lc.ShutdownRequest():
		wd.lc.ShutdownInitiated(err)
		return // Nothing to do
	}

	wd.log.Info("watchdog closing bid")
	err := wd.sess.Client().Tx().Broadcast(wd.ctx, &types.MsgCloseBid{
		BidID: types.MakeBidID(wd.leaseID.OrderID(), wd.sess.Provider().Address()),
	})
	if err != nil {
		wd.log.Error("failed closing bid", "err", err)
	}
}
