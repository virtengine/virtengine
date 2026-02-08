package provider_daemon

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/virtengine/virtengine/pkg/waldur"
	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

type fakeLifecycleExecutor struct {
	execCount atomic.Int64
	execErr   error
	state     string
}

var (
	lifecycleQueueMetricsOnce sync.Once
	lifecycleQueueMetrics     *LifecycleQueueMetrics
)

func testLifecycleQueueMetrics() *LifecycleQueueMetrics {
	lifecycleQueueMetricsOnce.Do(func() {
		lifecycleQueueMetrics = NewLifecycleQueueMetrics()
	})
	return lifecycleQueueMetrics
}

func (f *fakeLifecycleExecutor) Execute(ctx context.Context, cmd *LifecycleCommand) (*LifecycleCommandExecutionResult, error) {
	f.execCount.Add(1)
	if f.execErr != nil {
		return nil, f.execErr
	}
	return &LifecycleCommandExecutionResult{WaldurOperationID: "op-1", ResourceState: f.state}, nil
}

func (f *fakeLifecycleExecutor) GetResourceState(ctx context.Context, resourceUUID string) (waldur.ResourceState, error) {
	return waldur.ResourceState(f.state), nil
}

func TestLifecycleCommandQueue_PersistsAcrossRestart(t *testing.T) {
	tmp := t.TempDir()
	store, err := OpenBadgerLifecycleCommandStore(tmp, false)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}

	event := marketplace.NewLifecycleActionRequestedEventAt(
		"alloc-1",
		"order-1",
		"provider-1",
		marketplace.LifecycleActionStart,
		"op-1",
		"tester",
		marketplace.AllocationStateActive,
		map[string]interface{}{},
		marketplace.RollbackPolicyManual,
		1,
		1,
		time.Now(),
	)

	cfg := DefaultLifecycleCommandQueueConfig()
	cfg.Path = tmp
	cfg.WorkerCount = 1
	cfg.PollInterval = 20 * time.Millisecond
	cfg.ReconcileInterval = 200 * time.Millisecond
	cfg.StaleAfter = 50 * time.Millisecond

	executor := &fakeLifecycleExecutor{state: "OK"}
	queue, err := NewLifecycleCommandQueue(cfg, store, executor, func(ctx context.Context, allocationID string) (string, error) {
		return "resource-1", nil
	}, nil, testLifecycleQueueMetrics())
	if err != nil {
		t.Fatalf("create queue: %v", err)
	}

	if _, err := queue.EnqueueFromEvent(context.Background(), *event, "resource-1"); err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	queue.Stop()

	store2, err := OpenBadgerLifecycleCommandStore(tmp, false)
	if err != nil {
		t.Fatalf("open store restart: %v", err)
	}
	queue2, err := NewLifecycleCommandQueue(cfg, store2, executor, func(ctx context.Context, allocationID string) (string, error) {
		return "resource-1", nil
	}, nil, testLifecycleQueueMetrics())
	if err != nil {
		t.Fatalf("create queue restart: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := queue2.Start(ctx); err != nil {
		t.Fatalf("start queue: %v", err)
	}

	defer queue2.Stop()

	deadline := time.Now().Add(1 * time.Second)
	for time.Now().Before(deadline) {
		cmd, err := store2.Get(context.Background(), "op-1")
		if err == nil && cmd != nil && cmd.Status == LifecycleCommandStatusSucceeded {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}

	t.Fatalf("command not processed after restart")
}

func TestLifecycleCommandQueue_RequeueStaleExecuting(t *testing.T) {
	tmp := t.TempDir()
	store, err := OpenBadgerLifecycleCommandStore(tmp, true)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}

	now := time.Now().UTC()
	cmd := &LifecycleCommand{
		ID:             "cmd-stale",
		AllocationID:   "alloc-2",
		Action:         marketplace.LifecycleActionStop,
		TargetState:    marketplace.AllocationStateSuspended,
		RequestedBy:    "tester",
		Status:         LifecycleCommandStatusExecuting,
		AttemptCount:   1,
		MaxAttempts:    3,
		LastAttemptAt:  ptrTime(now.Add(-2 * time.Hour)),
		CreatedAt:      now.Add(-2 * time.Hour),
		UpdatedAt:      now.Add(-2 * time.Hour),
		EventTimestamp: now.Add(-2 * time.Hour),
	}
	if _, _, err := store.Enqueue(context.Background(), cmd); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	if err := store.Update(context.Background(), cmd); err != nil {
		t.Fatalf("update: %v", err)
	}

	cfg := DefaultLifecycleCommandQueueConfig()
	cfg.StaleAfter = 30 * time.Minute

	executor := &fakeLifecycleExecutor{state: "Stopped"}
	queue, err := NewLifecycleCommandQueue(cfg, store, executor, nil, nil, testLifecycleQueueMetrics())
	if err != nil {
		t.Fatalf("create queue: %v", err)
	}

	queue.ReconcileOnce(context.Background())

	updated, err := store.Get(context.Background(), "cmd-stale")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if updated.Status != LifecycleCommandStatusPending {
		t.Fatalf("expected pending after requeue, got %s", updated.Status)
	}
	if updated.NextAttemptAt == nil {
		t.Fatalf("expected next attempt timestamp")
	}
}

func TestLifecycleCommandQueue_ReconcileDrift(t *testing.T) {
	tmp := t.TempDir()
	store, err := OpenBadgerLifecycleCommandStore(tmp, true)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}

	desired := &LifecycleDesiredState{
		AllocationID:  "alloc-3",
		ResourceUUID:  "resource-3",
		DesiredState:  marketplace.AllocationStateActive,
		LastAction:    marketplace.LifecycleActionStart,
		LastCommandID: "cmd-prev",
		UpdatedAt:     time.Now().UTC(),
	}
	if err := store.SetDesiredState(context.Background(), desired); err != nil {
		t.Fatalf("set desired: %v", err)
	}

	executor := &fakeLifecycleExecutor{state: "Stopped"}
	queue, err := NewLifecycleCommandQueue(DefaultLifecycleCommandQueueConfig(), store, executor, nil, nil, testLifecycleQueueMetrics())
	if err != nil {
		t.Fatalf("create queue: %v", err)
	}

	queue.ReconcileOnce(context.Background())

	cmds, err := store.ListByAllocation(context.Background(), "alloc-3", LifecycleCommandStatusPending)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(cmds) == 0 {
		t.Fatalf("expected reconcile command to be enqueued")
	}
	if cmds[0].Action != marketplace.LifecycleActionStart {
		t.Fatalf("expected start action, got %s", cmds[0].Action)
	}
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
