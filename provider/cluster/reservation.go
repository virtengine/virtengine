package cluster

import (
	atypes "github.com/virtengine/virtengine/types"
	mtypes "github.com/virtengine/virtengine/x/market/types"
)

func newReservation(order mtypes.OrderID, resources atypes.ResourceGroup) *reservation {
	return &reservation{order: order, resources: resources}
}

type reservation struct {
	order     mtypes.OrderID
	resources atypes.ResourceGroup
	allocated bool
}

func (r *reservation) OrderID() mtypes.OrderID {
	return r.order
}

func (r *reservation) Resources() atypes.ResourceGroup {
	return r.resources
}
