package manifest

import (
	"github.com/virtengine/virtengine/manifest"
	dtypes "github.com/virtengine/virtengine/x/deployment/types"
)

// Status is the data structure
type Status struct {
	Deployments uint32 `json:"deployments"`
}

type submitRequest struct {
	Deployment dtypes.DeploymentID `json:"deployment"`
	Manifest   manifest.Manifest   `json:"manifest"`
}
