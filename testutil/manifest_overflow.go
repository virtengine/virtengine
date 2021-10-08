package testutil

import (
	"math"
	"testing"

	"github.com/virtengine/virtengine/manifest"
	"github.com/virtengine/virtengine/types"
)

// OverflowManifestGenerator generates a manifest maximum integer values
var OverflowManifestGenerator ManifestGenerator = manifestGeneratorOverflow{}

type manifestGeneratorOverflow struct{}

func (mg manifestGeneratorOverflow) Manifest(t testing.TB) manifest.Manifest {
	t.Helper()
	return []manifest.Group{
		mg.Group(t),
	}
}

func (mg manifestGeneratorOverflow) Group(t testing.TB) manifest.Group {
	t.Helper()
	return manifest.Group{
		Name: Name(t, "manifest-group"),
		Services: []manifest.Service{
			mg.Service(t),
		},
	}
}

func (mg manifestGeneratorOverflow) Service(t testing.TB) manifest.Service {
	t.Helper()
	return manifest.Service{
		Name:  "demo",
		Image: "quay.io/virtengine/demo-app",
		Args:  []string{"run"},
		Env:   []string{"VIRTENGINE_TEST_SERVICE=true"},
		Resources: types.ResourceUnits{
			CPU: &types.CPU{
				Units: types.NewResourceValue(math.MaxUint32),
			},
			Memory: &types.Memory{
				Quantity: types.NewResourceValue(math.MaxUint64),
			},
			Storage: &types.Storage{
				Quantity: types.NewResourceValue(math.MaxUint64),
			},
		},
		Count: math.MaxUint32,
		Expose: []manifest.ServiceExpose{
			mg.ServiceExpose(t),
		},
	}
}

func (mg manifestGeneratorOverflow) ServiceExpose(t testing.TB) manifest.ServiceExpose {
	t.Helper()
	return manifest.ServiceExpose{
		Port:         math.MaxUint16,
		ExternalPort: math.MaxUint16,
		Proto:        "TCP",
		Service:      "svc",
		Global:       true,
		Hosts: []string{
			Hostname(t),
		},
	}
}
