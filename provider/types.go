package provider

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/provider/bidengine"
	ctypes "github.com/virtengine/virtengine/provider/cluster/types"
	"github.com/virtengine/virtengine/provider/manifest"
)

// Status is the data structure that stores Cluster, Bidengine and Manifest details.
type Status struct {
	Cluster               *ctypes.Status    `json:"cluster"`
	Bidengine             *bidengine.Status `json:"bidengine"`
	Manifest              *manifest.Status  `json:"manifest"`
	ClusterPublicHostname string            `json:"cluster_public_hostname,omitempty"`
}

type ValidateGroupSpecResult struct {
	MinBidPrice sdk.Coin `json:"min_bid_price"`
}
