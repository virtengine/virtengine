package types

import (
	resourcesv1 "github.com/virtengine/virtengine/sdk/go/node/resources/v1"
)

// Type aliases to generated protobuf types.
type (
	ResourceInventory  = resourcesv1.ResourceInventory
	ResourceRequest    = resourcesv1.ResourceRequest
	ResourceAllocation = resourcesv1.ResourceAllocation
	ResourceCapacity   = resourcesv1.ResourceCapacity
	Locality           = resourcesv1.Locality
	AllocationEvent    = resourcesv1.AllocationEvent
	SlashingEvent      = resourcesv1.SlashingEvent
	ResourceClass      = resourcesv1.ResourceClass
	AllocationState    = resourcesv1.AllocationState
)

// ResourceClass enum constants.
const (
	ResourceClassUnspecified = resourcesv1.ResourceClass_RESOURCE_CLASS_UNSPECIFIED
	ResourceClassCompute     = resourcesv1.ResourceClass_RESOURCE_CLASS_COMPUTE
	ResourceClassStorage     = resourcesv1.ResourceClass_RESOURCE_CLASS_STORAGE
	ResourceClassNetwork     = resourcesv1.ResourceClass_RESOURCE_CLASS_NETWORK
)

// AllocationState enum constants.
const (
	AllocationStateUnspecified = resourcesv1.AllocationState_ALLOCATION_STATE_UNSPECIFIED
	AllocationStatePending     = resourcesv1.AllocationState_ALLOCATION_STATE_PENDING
	AllocationStateActive      = resourcesv1.AllocationState_ALLOCATION_STATE_ACTIVE
	AllocationStateExpired     = resourcesv1.AllocationState_ALLOCATION_STATE_EXPIRED
	AllocationStateReleased    = resourcesv1.AllocationState_ALLOCATION_STATE_RELEASED
)
