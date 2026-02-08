package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	resourcesv1 "github.com/virtengine/virtengine/sdk/go/node/resources/v1"
)

// Type aliases to generated protobuf types.
type (
	MsgProviderHeartbeat          = resourcesv1.MsgProviderHeartbeat
	MsgProviderHeartbeatResponse  = resourcesv1.MsgProviderHeartbeatResponse
	MsgAllocateResources          = resourcesv1.MsgAllocateResources
	MsgAllocateResourcesResponse  = resourcesv1.MsgAllocateResourcesResponse
	MsgActivateAllocation         = resourcesv1.MsgActivateAllocation
	MsgActivateAllocationResponse = resourcesv1.MsgActivateAllocationResponse
	MsgReleaseAllocation          = resourcesv1.MsgReleaseAllocation
	MsgReleaseAllocationResponse  = resourcesv1.MsgReleaseAllocationResponse
	MsgUpdateParams               = resourcesv1.MsgUpdateParams
	MsgUpdateParamsResponse       = resourcesv1.MsgUpdateParamsResponse
)

// Message type constants.
const (
	TypeMsgProviderHeartbeat  = "provider_heartbeat"
	TypeMsgAllocateResources  = "allocate_resources"
	TypeMsgActivateAllocation = "activate_allocation"
	TypeMsgReleaseAllocation  = "release_allocation"
)

var (
	_ sdk.Msg = &MsgProviderHeartbeat{}
	_ sdk.Msg = &MsgAllocateResources{}
	_ sdk.Msg = &MsgActivateAllocation{}
	_ sdk.Msg = &MsgReleaseAllocation{}
	_ sdk.Msg = &MsgUpdateParams{}
)

// NewMsgProviderHeartbeat creates a heartbeat message.
func NewMsgProviderHeartbeat(provider, inventoryID string, class ResourceClass, total, available ResourceCapacity, locality Locality, sequence uint64) *MsgProviderHeartbeat {
	return &MsgProviderHeartbeat{
		ProviderAddress: provider,
		InventoryId:     inventoryID,
		ResourceClass:   class,
		Total:           total,
		Available:       available,
		Locality:        locality,
		Sequence:        sequence,
	}
}

// NewMsgAllocateResources creates an allocation request message.
func NewMsgAllocateResources(requester string, request ResourceRequest) *MsgAllocateResources {
	return &MsgAllocateResources{
		RequesterAddress: requester,
		Request:          request,
	}
}

// NewMsgActivateAllocation creates an allocation activation message.
func NewMsgActivateAllocation(provider, allocationID string) *MsgActivateAllocation {
	return &MsgActivateAllocation{
		ProviderAddress: provider,
		AllocationId:    allocationID,
	}
}

// NewMsgReleaseAllocation creates a release allocation message.
func NewMsgReleaseAllocation(requester, allocationID, reason string) *MsgReleaseAllocation {
	return &MsgReleaseAllocation{
		RequesterAddress: requester,
		AllocationId:     allocationID,
		Reason:           reason,
	}
}

// NewMsgUpdateParams creates an update params message.
func NewMsgUpdateParams(authority string, params Params) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: authority,
		Params:    params.ToProto(),
	}
}
