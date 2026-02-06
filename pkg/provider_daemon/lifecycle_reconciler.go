// Package provider_daemon implements the provider daemon for VirtEngine.
//
// VE-27C: Lifecycle drift reconciler between chain allocation state and Waldur resource state.
package provider_daemon

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/virtengine/virtengine/pkg/waldur"
	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

// LifecycleDriftReconcilerConfig configures lifecycle drift reconciliation.
type LifecycleDriftReconcilerConfig struct {
	// Enabled toggles the reconciler.
	Enabled bool

	// ReconcileInterval defines how often to reconcile.
	ReconcileInterval time.Duration

	// MaxConcurrent limits concurrent reconciliation actions.
	MaxConcurrent int
}

// DefaultLifecycleDriftReconcilerConfig returns defaults.
func DefaultLifecycleDriftReconcilerConfig() LifecycleDriftReconcilerConfig {
	return LifecycleDriftReconcilerConfig{
		Enabled:           true,
		ReconcileInterval: 5 * time.Minute,
		MaxConcurrent:     5,
	}
}

// LifecycleDriftReconciler reconciles chain allocation state with Waldur.
type LifecycleDriftReconciler struct {
	cfg          LifecycleDriftReconcilerConfig
	controller   *LifecycleController
	lifecycleMgr *ResourceLifecycleManager
	lifecycle    *waldur.LifecycleClient
	auditLogger  *AuditLogger

	mu      sync.RWMutex
	running bool
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

// NewLifecycleDriftReconciler creates a new reconciler.
func NewLifecycleDriftReconciler(
	cfg LifecycleDriftReconcilerConfig,
	controller *LifecycleController,
	lifecycleMgr *ResourceLifecycleManager,
	lifecycle *waldur.LifecycleClient,
	auditLogger *AuditLogger,
) *LifecycleDriftReconciler {
	return &LifecycleDriftReconciler{
		cfg:          cfg,
		controller:   controller,
		lifecycleMgr: lifecycleMgr,
		lifecycle:    lifecycle,
		auditLogger:  auditLogger,
		stopCh:       make(chan struct{}),
	}
}

// Start begins the reconciliation loop.
func (r *LifecycleDriftReconciler) Start(ctx context.Context) error {
	if !r.cfg.Enabled {
		return nil
	}

	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return nil
	}
	r.running = true
	r.mu.Unlock()

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		r.run(ctx)
	}()

	log.Printf("[lifecycle-reconciler] started with interval %v", r.cfg.ReconcileInterval)
	return nil
}

// Stop stops the reconciliation loop.
func (r *LifecycleDriftReconciler) Stop() {
	r.mu.Lock()
	if !r.running {
		r.mu.Unlock()
		return
	}
	r.running = false
	r.mu.Unlock()

	close(r.stopCh)
	r.wg.Wait()
	r.stopCh = make(chan struct{})
	log.Printf("[lifecycle-reconciler] stopped")
}

// ReconcileOnce runs a single reconciliation cycle.
func (r *LifecycleDriftReconciler) ReconcileOnce(ctx context.Context) {
	if r.lifecycleMgr == nil || r.lifecycle == nil || r.controller == nil {
		return
	}

	resources := r.lifecycleMgr.GetManagedResources()
	if len(resources) == 0 {
		return
	}

	activeOps := r.lifecycleMgr.GetActiveOperations()
	sem := make(chan struct{}, r.cfg.MaxConcurrent)
	var wg sync.WaitGroup

	for _, resource := range resources {
		if resource == nil || resource.WaldurResourceUUID == "" {
			continue
		}
		if _, busy := activeOps[resource.AllocationID]; busy {
			continue
		}
		if pending := r.controller.GetPendingOperations(resource.AllocationID); len(pending) > 0 {
			continue
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(info *ResourceInfo) {
			defer func() {
				<-sem
				wg.Done()
			}()
			r.reconcileResource(ctx, info)
		}(resource)
	}

	wg.Wait()
}

func (r *LifecycleDriftReconciler) run(ctx context.Context) {
	ticker := time.NewTicker(r.cfg.ReconcileInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-r.stopCh:
			return
		case <-ticker.C:
			r.ReconcileOnce(ctx)
		}
	}
}

func (r *LifecycleDriftReconciler) reconcileResource(ctx context.Context, info *ResourceInfo) {
	waldurState, err := r.lifecycle.GetResourceState(ctx, info.WaldurResourceUUID)
	if err != nil {
		log.Printf("[lifecycle-reconciler] failed to get Waldur state for %s: %v", info.AllocationID, err)
		return
	}

	observedState := mapWaldurStateToAllocationState(string(waldurState))
	if observedState == info.CurrentState {
		return
	}

	action, ok := r.actionForDrift(info.CurrentState, waldurState)
	if !ok {
		log.Printf("[lifecycle-reconciler] drift detected but no safe action for %s (chain=%s waldur=%s)",
			info.AllocationID, info.CurrentState, waldurState)
		return
	}

	params := map[string]string{
		"reconcile":        "true",
		"expected_state":   info.CurrentState.String(),
		"observed_state":   observedState.String(),
		"waldur_state":     string(waldurState),
		"correlation_id":   fmt.Sprintf("reconcile-%s-%s", info.AllocationID, action),
		"reconcile_action": string(action),
	}

	_, err = r.controller.ExecuteLifecycleAction(
		ctx,
		info.AllocationID,
		action,
		observedState,
		info.WaldurResourceUUID,
		"reconciler",
		params,
	)
	if err != nil {
		log.Printf("[lifecycle-reconciler] reconcile action failed for %s: %v", info.AllocationID, err)
		return
	}

	if r.auditLogger != nil {
		_ = r.auditLogger.Log(&AuditEvent{
			Type:      AuditEventLifecycleDriftReconciled,
			Operation: string(action),
			Success:   true,
			Details: map[string]interface{}{
				"allocation_id":  info.AllocationID,
				"expected_state": info.CurrentState.String(),
				"observed_state": observedState.String(),
				"waldur_state":   string(waldurState),
			},
		})
	}
}

func (r *LifecycleDriftReconciler) actionForDrift(chainState marketplace.AllocationState, waldurState waldur.ResourceState) (marketplace.LifecycleActionType, bool) {
	switch chainState {
	case marketplace.AllocationStateActive:
		switch waldurState {
		case waldur.ResourceStateStopped:
			return marketplace.LifecycleActionStart, true
		case waldur.ResourceStatePaused:
			return marketplace.LifecycleActionResume, true
		default:
			return "", false
		}
	case marketplace.AllocationStateSuspended:
		if waldurState == waldur.ResourceStateOK {
			return marketplace.LifecycleActionStop, true
		}
	case marketplace.AllocationStateTerminating, marketplace.AllocationStateTerminated:
		if waldurState != waldur.ResourceStateTerminated && waldurState != waldur.ResourceStateTerminating {
			return marketplace.LifecycleActionTerminate, true
		}
	}
	return "", false
}
