package cluster

import (
	atypes "github.com/virtengine/virtengine/types"
	mtypes "github.com/virtengine/virtengine/x/market/types"
)

// Reservation interface implements orders and resources
type Reservation interface {
	OrderID() mtypes.OrderID
	Resources() atypes.ResourceGroup
}
