package cluster

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"k8s.io/client-go/tools/remotecommand"
	"math/rand"
	"sync"
	"time"

	eventsv1 "k8s.io/api/events/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/virtengine/virtengine/manifest"
	ctypes "github.com/virtengine/virtengine/provider/cluster/types"
	atypes "github.com/virtengine/virtengine/types"
	"github.com/virtengine/virtengine/types/unit"
	mquery "github.com/virtengine/virtengine/x/market/query"
	mtypes "github.com/virtengine/virtengine/x/market/types"
)

var (
	_ Client = (*nullClient)(nil)
	// Errors types returned by the Exec function on the client interface
	ErrExec                        = errors.New("remote command execute error")
	ErrExecNoServiceWithName       = fmt.Errorf("%w: no such service exists with that name", ErrExec)
	ErrExecServiceNotRunning       = fmt.Errorf("%w: service with that name is not running", ErrExec)
	ErrExecCommandExecutionFailed  = fmt.Errorf("%w: command execution failed", ErrExec)
	ErrExecCommandDoesNotExist     = fmt.Errorf("%w: command could not be executed because it does not exist", ErrExec)
	ErrExecDeploymentNotYetRunning = fmt.Errorf("%w: deployment is not yet active", ErrExec)
	ErrExecMultiplePods            = fmt.Errorf("%w: cannot execute without specifying a pod explicitly", ErrExec)
	ErrExecPodIndexOutOfRange      = fmt.Errorf("%w: pod index out of range", ErrExec)

	errNotImplemented = errors.New("not implemented")
)

type ReadClient interface {
	LeaseStatus(context.Context, mtypes.LeaseID) (*ctypes.LeaseStatus, error)
	LeaseEvents(context.Context, mtypes.LeaseID, string, bool) (ctypes.EventsWatcher, error)
	LeaseLogs(context.Context, mtypes.LeaseID, string, bool, *int64) ([]*ctypes.ServiceLog, error)
	ServiceStatus(context.Context, mtypes.LeaseID, string) (*ctypes.ServiceStatus, error)
}

// Client interface lease and deployment methods
type Client interface {
	ReadClient
	Deploy(context.Context, mtypes.LeaseID, *manifest.Group) error
	TeardownLease(context.Context, mtypes.LeaseID) error
	Deployments(context.Context) ([]ctypes.Deployment, error)
	Inventory(context.Context) ([]ctypes.Node, error)
	Exec(ctx context.Context,
		lID mtypes.LeaseID,
		service string,
		podIndex uint,
		cmd []string,
		stdin io.Reader,
		stdout io.Writer,
		stderr io.Writer,
		tty bool,
		tsq remotecommand.TerminalSizeQueue) (ctypes.ExecResult, error)
}

func ErrorIsOkToSendToClient(err error) bool {
	return errors.Is(err, ErrExec)
}

type node struct {
	id                    string
	availableResources    atypes.ResourceUnits
	allocateableResources atypes.ResourceUnits
}

// NewNode returns new Node instance with provided details
func NewNode(id string, allocateable atypes.ResourceUnits, available atypes.ResourceUnits) ctypes.Node {
	return &node{id: id, allocateableResources: allocateable, availableResources: available}
}

// ID returns id of node
func (n *node) ID() string {
	return n.id
}

func (n *node) Reserve(atypes.ResourceUnits) error {
	return nil
}

// Available returns available units of node
func (n *node) Available() atypes.ResourceUnits {
	return n.availableResources
}

func (n *node) Allocateable() atypes.ResourceUnits {
	return n.allocateableResources
}

const (
	// 5 CPUs, 5Gi memory for null client.
	nullClientCPU     = 5000
	nullClientMemory  = 32 * unit.Gi
	nullClientStorage = 512 * unit.Gi
)

type nullLease struct {
	ctx    context.Context
	cancel func()
	group  *manifest.Group
}

type nullClient struct {
	leases map[string]*nullLease
	mtx    sync.Mutex
}

// NewServiceLog creates and returns a service log with provided details
func NewServiceLog(name string, stream io.ReadCloser) *ctypes.ServiceLog {
	return &ctypes.ServiceLog{
		Name:    name,
		Stream:  stream,
		Scanner: bufio.NewScanner(stream),
	}
}

// NullClient returns nullClient instance
func NullClient() Client {
	return &nullClient{
		leases: make(map[string]*nullLease),
		mtx:    sync.Mutex{},
	}
}

func (c *nullClient) Deploy(ctx context.Context, lid mtypes.LeaseID, mgroup *manifest.Group) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	ctx, cancel := context.WithCancel(ctx)
	c.leases[mquery.LeasePath(lid)] = &nullLease{
		ctx:    ctx,
		cancel: cancel,
		group:  mgroup,
	}

	return nil
}

func (c *nullClient) LeaseStatus(_ context.Context, lid mtypes.LeaseID) (*ctypes.LeaseStatus, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	lease, ok := c.leases[mquery.LeasePath(lid)]
	if !ok {
		return nil, nil
	}

	resp := &ctypes.LeaseStatus{}
	resp.Services = make(map[string]*ctypes.ServiceStatus)
	for _, svc := range lease.group.Services {
		resp.Services[svc.Name] = &ctypes.ServiceStatus{
			Name:      svc.Name,
			Available: int32(svc.Count),
			Total:     int32(svc.Count),
		}
	}

	return resp, nil
}

func (c *nullClient) LeaseEvents(ctx context.Context, lid mtypes.LeaseID, _ string, follow bool) (ctypes.EventsWatcher, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	lease, ok := c.leases[mquery.LeasePath(lid)]
	if !ok {
		return nil, nil
	}

	if lease.ctx.Err() != nil {
		return nil, nil
	}

	feed := ctypes.NewEventsFeed(ctx)
	go func() {
		defer feed.Shutdown()

		tm := time.NewTicker(7 * time.Second)
		tm.Stop()

		genEvent := func() *eventsv1.Event {
			return &eventsv1.Event{
				EventTime:           v1.NewMicroTime(time.Now()),
				ReportingController: lease.group.Name,
			}
		}

		nfollowCh := make(chan *eventsv1.Event, 1)
		count := 0
		if !follow {
			count = rand.Intn(9) // nolint: gosec
			nfollowCh <- genEvent()
		} else {
			tm.Reset(time.Second)
		}

		for {
			select {
			case <-lease.ctx.Done():
				return
			case evt := <-nfollowCh:
				if !feed.SendEvent(evt) || count == 0 {
					return
				}
				count--
				nfollowCh <- genEvent()
				break
			case <-tm.C:
				tm.Stop()
				if !feed.SendEvent(genEvent()) {
					return
				}
				tm.Reset(time.Duration(rand.Intn(9)+1) * time.Second) // nolint: gosec
				break
			}
		}
	}()

	return feed, nil
}

func (c *nullClient) ServiceStatus(_ context.Context, _ mtypes.LeaseID, _ string) (*ctypes.ServiceStatus, error) {
	return nil, nil
}

func (c *nullClient) LeaseLogs(_ context.Context, _ mtypes.LeaseID, _ string, _ bool, _ *int64) ([]*ctypes.ServiceLog, error) {
	return nil, nil
}

func (c *nullClient) TeardownLease(_ context.Context, lid mtypes.LeaseID) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if lease, ok := c.leases[mquery.LeasePath(lid)]; ok {
		delete(c.leases, mquery.LeasePath(lid))
		lease.cancel()
	}

	return nil
}

func (c *nullClient) Deployments(context.Context) ([]ctypes.Deployment, error) {
	return nil, nil
}

func (c *nullClient) Inventory(context.Context) ([]ctypes.Node, error) {
	return []ctypes.Node{
		NewNode("solo", atypes.ResourceUnits{
			CPU: &atypes.CPU{
				Units: atypes.NewResourceValue(nullClientCPU),
			},
			Memory: &atypes.Memory{
				Quantity: atypes.NewResourceValue(nullClientMemory),
			},
			Storage: &atypes.Storage{
				Quantity: atypes.NewResourceValue(nullClientStorage),
			},
		},
			atypes.ResourceUnits{
				CPU: &atypes.CPU{
					Units: atypes.NewResourceValue(nullClientCPU),
				},
				Memory: &atypes.Memory{
					Quantity: atypes.NewResourceValue(nullClientMemory),
				},
				Storage: &atypes.Storage{
					Quantity: atypes.NewResourceValue(nullClientStorage),
				},
			}),
	}, nil
}

func (c *nullClient) Exec(context.Context, mtypes.LeaseID, string, uint, []string, io.Reader, io.Writer, io.Writer, bool, remotecommand.TerminalSizeQueue) (ctypes.ExecResult, error) {
	return nil, errNotImplemented
}
