package keeper

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

// applyAllocationStateTransition moves an allocation to the target state using valid transitions.
func applyAllocationStateTransition(allocation *marketplace.Allocation, target marketplace.AllocationState, reason string, now time.Time) error {
	if allocation == nil {
		return fmt.Errorf("allocation is nil")
	}
	if allocation.State == target {
		return nil
	}

	if !isAllocationTransitionAllowed(allocation.State, target) {
		return fmt.Errorf("invalid transition %s -> %s", allocation.State.String(), target.String())
	}

	return allocation.SetStateAt(target, reason, now)
}

func isAllocationTransitionAllowed(from, to marketplace.AllocationState) bool {
	allowed := map[marketplace.AllocationState][]marketplace.AllocationState{
		marketplace.AllocationStatePending: {
			marketplace.AllocationStateAccepted,
			marketplace.AllocationStateProvisioning,
			marketplace.AllocationStateTerminated,
			marketplace.AllocationStateRejected,
			marketplace.AllocationStateFailed,
		},
		marketplace.AllocationStateAccepted: {
			marketplace.AllocationStateProvisioning,
			marketplace.AllocationStateTerminated,
			marketplace.AllocationStateFailed,
		},
		marketplace.AllocationStateProvisioning: {
			marketplace.AllocationStateActive,
			marketplace.AllocationStateFailed,
			marketplace.AllocationStateTerminating,
		},
		marketplace.AllocationStateActive: {
			marketplace.AllocationStateSuspended,
			marketplace.AllocationStateTerminating,
			marketplace.AllocationStateFailed,
		},
		marketplace.AllocationStateSuspended: {
			marketplace.AllocationStateActive,
			marketplace.AllocationStateTerminating,
			marketplace.AllocationStateFailed,
		},
		marketplace.AllocationStateTerminating: {
			marketplace.AllocationStateTerminated,
			marketplace.AllocationStateFailed,
		},
	}

	for _, candidate := range allowed[from] {
		if candidate == to {
			return true
		}
	}
	return false
}

func provisioningPhaseForState(state marketplace.AllocationState) marketplace.ProvisioningPhase {
	switch state {
	case marketplace.AllocationStateProvisioning:
		return marketplace.ProvisioningPhaseProvisioning
	case marketplace.AllocationStateActive:
		return marketplace.ProvisioningPhaseActive
	case marketplace.AllocationStateTerminating, marketplace.AllocationStateTerminated:
		return marketplace.ProvisioningPhaseTerminated
	case marketplace.AllocationStateFailed, marketplace.AllocationStateRejected:
		return marketplace.ProvisioningPhaseFailed
	default:
		return marketplace.ProvisioningPhaseRequested
	}
}

func parseProgress(payload map[string]string, fallback marketplace.AllocationState) uint8 {
	if payload == nil {
		if fallback == marketplace.AllocationStateActive {
			return 100
		}
		return 0
	}
	if value := strings.TrimSpace(payload["progress"]); value != "" {
		parsed, err := strconv.Atoi(value)
		if err == nil {
			if parsed < 0 {
				return 0
			}
			if parsed > 100 {
				parsed = 100
			}
			if parsed > int(^uint8(0)) {
				return 100
			}
			//nolint:gosec // parsed is clamped to a safe 0-100 range above.
			return uint8(parsed)
		}
	}
	if fallback == marketplace.AllocationStateActive {
		return 100
	}
	return 0
}

func (k Keeper) emitAllocationStateChange(ctx sdk.Context, allocation *marketplace.Allocation, oldState marketplace.AllocationState, reason string) error {
	seq := k.IncrementEventSequence(ctx)
	event := marketplace.NewAllocationStateChangedEventAt(allocation, oldState, reason, ctx.BlockHeight(), seq, ctx.BlockTime())
	return k.EmitMarketplaceEvent(ctx, event)
}
