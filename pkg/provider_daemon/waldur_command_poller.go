package provider_daemon

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/virtengine/virtengine/pkg/waldur"
	marketplacev1 "github.com/virtengine/virtengine/sdk/go/node/marketplace/v1"
)

// WaldurCommandView is the minimal durable-command surface the poller needs.
type WaldurCommandView struct {
	// ID is the command identifier.
	ID string
	// Kind is the command kind (e.g. create_order).
	Kind string
	// InstanceID identifies the target Waldur instance.
	InstanceID string
	// ChainEntityID is the on-chain entity ID.
	ChainEntityID string
	// WaldurOfferingUUID is the target Waldur offering UUID.
	WaldurOfferingUUID string
	// BackendID is the canonical VirtEngine identifier for reconciliation.
	BackendID string
	// Acked indicates the command was already acknowledged.
	Acked bool
}

// WaldurCommandQuerier lists pending durable commands.
type WaldurCommandQuerier interface {
	// PendingCommands returns unacknowledged commands for the instance.
	PendingCommands(ctx context.Context, instanceID string) ([]WaldurCommandView, error)
}

// ChainWaldurCommandQuerier implements WaldurCommandQuerier against the
// on-chain WaldurCommands query.
type ChainWaldurCommandQuerier struct {
	query marketplacev1.QueryClient
}

// NewChainWaldurCommandQuerier creates a chain-backed command querier.
func NewChainWaldurCommandQuerier(query marketplacev1.QueryClient) (*ChainWaldurCommandQuerier, error) {
	if query == nil {
		return nil, fmt.Errorf("marketplace query client is required")
	}
	return &ChainWaldurCommandQuerier{query: query}, nil
}

// PendingCommands implements WaldurCommandQuerier.
func (q *ChainWaldurCommandQuerier) PendingCommands(ctx context.Context, instanceID string) ([]WaldurCommandView, error) {
	resp, err := q.query.WaldurCommands(ctx, &marketplacev1.QueryWaldurCommandsRequest{
		InstanceId:  instanceID,
		PendingOnly: true,
	})
	if err != nil {
		return nil, err
	}
	commands := make([]WaldurCommandView, 0, len(resp.Commands))
	for _, command := range resp.Commands {
		commands = append(commands, WaldurCommandView{
			ID:                 command.Id,
			Kind:               command.Kind,
			InstanceID:         command.InstanceId,
			ChainEntityID:      command.ChainEntityId,
			WaldurOfferingUUID: command.WaldurOfferingUuid,
			BackendID:          command.BackendId,
			Acked:              command.Acked,
		})
	}
	return commands, nil
}

// WaldurOrderExecutor executes create_order commands against Waldur.
type WaldurOrderExecutor interface {
	// ExecuteCreateOrder creates, approves, and links a Waldur order,
	// returning the Waldur order UUID.
	ExecuteCreateOrder(ctx context.Context, command WaldurCommandView) (string, error)
}

// marketplaceOrderExecutor implements WaldurOrderExecutor with the Waldur
// marketplace client.
type marketplaceOrderExecutor struct {
	marketplace *waldur.MarketplaceClient
	projectUUID string
}

// NewMarketplaceOrderExecutor creates an order executor bound to a Waldur project.
func NewMarketplaceOrderExecutor(marketplace *waldur.MarketplaceClient, projectUUID string) (*marketplaceOrderExecutor, error) {
	if marketplace == nil {
		return nil, fmt.Errorf("waldur marketplace client is required")
	}
	if projectUUID == "" {
		return nil, fmt.Errorf("waldur project UUID is required")
	}
	return &marketplaceOrderExecutor{marketplace: marketplace, projectUUID: projectUUID}, nil
}

// ExecuteCreateOrder implements WaldurOrderExecutor.
func (e *marketplaceOrderExecutor) ExecuteCreateOrder(ctx context.Context, command WaldurCommandView) (string, error) {
	if command.WaldurOfferingUUID == "" {
		return "", fmt.Errorf("command %s has no Waldur offering UUID", command.ID)
	}
	backendID := command.BackendID
	if backendID == "" {
		backendID = command.ChainEntityID
	}
	order, err := e.marketplace.CreateOrder(ctx, waldur.CreateOrderRequest{
		OfferingUUID: command.WaldurOfferingUUID,
		ProjectUUID:  e.projectUUID,
		Name:         fmt.Sprintf("virtengine-%s", backendID),
		Description:  fmt.Sprintf("VirtEngine order %s", command.ChainEntityID),
		Attributes: map[string]interface{}{
			"ve_order_id": backendID,
		},
	})
	if err != nil {
		return "", fmt.Errorf("create waldur order: %w", err)
	}
	if err := e.marketplace.ApproveOrderByProvider(ctx, order.UUID); err != nil {
		return "", fmt.Errorf("approve waldur order: %w", err)
	}
	if err := e.marketplace.SetOrderBackendID(ctx, order.UUID, backendID); err != nil {
		return "", fmt.Errorf("set waldur backend id: %w", err)
	}
	return order.UUID, nil
}

// WaldurCommandAcker acknowledges executed commands on-chain.
type WaldurCommandAcker interface {
	// AckCommand submits MsgAckWaldurCommand for a completed command.
	AckCommand(ctx context.Context, commandID string) error
}

// chainWaldurCommandAcker implements WaldurCommandAcker via the mutation pipeline.
type chainWaldurCommandAcker struct {
	submitter *ProviderMutationSubmitter
	sender    string
}

// NewChainWaldurCommandAcker creates an on-chain command acknowledger.
func NewChainWaldurCommandAcker(submitter *ProviderMutationSubmitter, sender string) (*chainWaldurCommandAcker, error) {
	if submitter == nil {
		return nil, fmt.Errorf("mutation submitter is required")
	}
	if sender == "" {
		return nil, fmt.Errorf("sender address is required")
	}
	return &chainWaldurCommandAcker{submitter: submitter, sender: sender}, nil
}

// AckCommand implements WaldurCommandAcker.
func (a *chainWaldurCommandAcker) AckCommand(ctx context.Context, commandID string) error {
	if commandID == "" {
		return fmt.Errorf("command id is required")
	}
	_, err := a.submitter.Submit(ctx, MutationMarketplaceAckCommand, &marketplacev1.MsgAckWaldurCommand{
		Sender:    a.sender,
		CommandId: commandID,
	})
	return err
}

// WaldurCommandPollerConfig configures the command poller.
type WaldurCommandPollerConfig struct {
	// Enabled toggles the poller.
	Enabled bool
	// InstanceID scopes polling to one Waldur instance.
	InstanceID string
	// ProjectUUID is the Waldur project used for order execution.
	ProjectUUID string
	// PollIntervalSeconds is the polling interval.
	PollIntervalSeconds int64
	// OperationTimeout bounds a single poll cycle.
	OperationTimeout time.Duration
}

// DefaultWaldurCommandPollerConfig returns conservative defaults.
func DefaultWaldurCommandPollerConfig() WaldurCommandPollerConfig {
	return WaldurCommandPollerConfig{
		PollIntervalSeconds: 30,
		OperationTimeout:    60 * time.Second,
	}
}

// WaldurCommandPoller executes pending durable Waldur commands.
type WaldurCommandPoller struct {
	cfg      WaldurCommandPollerConfig
	querier  WaldurCommandQuerier
	executor WaldurOrderExecutor
	acker    WaldurCommandAcker

	mu      sync.RWMutex
	running bool
	stopCh  chan struct{}
	doneCh  chan struct{}
}

// NewWaldurCommandPoller creates a command poller.
func NewWaldurCommandPoller(cfg WaldurCommandPollerConfig, querier WaldurCommandQuerier, executor WaldurOrderExecutor, acker WaldurCommandAcker) (*WaldurCommandPoller, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("command poller is disabled")
	}
	if cfg.InstanceID == "" {
		return nil, fmt.Errorf("instance ID is required")
	}
	if querier == nil {
		return nil, fmt.Errorf("command querier is required")
	}
	if executor == nil {
		return nil, fmt.Errorf("order executor is required")
	}
	if acker == nil {
		return nil, fmt.Errorf("command acker is required")
	}
	if cfg.PollIntervalSeconds <= 0 {
		cfg.PollIntervalSeconds = DefaultWaldurCommandPollerConfig().PollIntervalSeconds
	}
	if cfg.OperationTimeout <= 0 {
		cfg.OperationTimeout = DefaultWaldurCommandPollerConfig().OperationTimeout
	}
	return &WaldurCommandPoller{
		cfg:      cfg,
		querier:  querier,
		executor: executor,
		acker:    acker,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}, nil
}

// Start begins polling.
func (p *WaldurCommandPoller) Start(ctx context.Context) error {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return fmt.Errorf("command poller already running")
	}
	p.running = true
	p.mu.Unlock()

	go p.loop(ctx)
	return nil
}

// Stop halts polling.
func (p *WaldurCommandPoller) Stop() {
	p.mu.Lock()
	if !p.running {
		p.mu.Unlock()
		return
	}
	p.running = false
	close(p.stopCh)
	p.mu.Unlock()
	<-p.doneCh
}

func (p *WaldurCommandPoller) loop(ctx context.Context) {
	defer close(p.doneCh)
	ticker := time.NewTicker(time.Duration(p.cfg.PollIntervalSeconds) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopCh:
			return
		case <-ticker.C:
			p.pollOnce(ctx)
		}
	}
}

func (p *WaldurCommandPoller) pollOnce(ctx context.Context) {
	opCtx, cancel := context.WithTimeout(ctx, p.cfg.OperationTimeout)
	defer cancel()

	commands, err := p.querier.PendingCommands(opCtx, p.cfg.InstanceID)
	if err != nil {
		log.Printf("[waldur-commands] poll failed: %v", err)
		return
	}
	for _, command := range commands {
		if err := p.execute(opCtx, command); err != nil {
			log.Printf("[waldur-commands] command %s failed: %v", command.ID, err)
		}
	}
}

// execute runs a single command and acknowledges it. Unknown kinds are left
// pending for a future poller revision.
func (p *WaldurCommandPoller) execute(ctx context.Context, command WaldurCommandView) error {
	switch command.Kind {
	case "create_order":
		waldurUUID, err := p.executor.ExecuteCreateOrder(ctx, command)
		if err != nil {
			return err
		}
		log.Printf("[waldur-commands] create_order %s executed as waldur order %s", command.ID, waldurUUID)
	default:
		log.Printf("[waldur-commands] skipping unsupported command kind %s (%s)", command.Kind, command.ID)
		return nil
	}
	if err := p.acker.AckCommand(ctx, command.ID); err != nil {
		return fmt.Errorf("acknowledge command: %w", err)
	}
	return nil
}
