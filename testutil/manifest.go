package testutil

import (
	"testing"

	"github.com/virtengine/virtengine/manifest"
)

var (
	// DefaultManifestGenerator is the default test manifest generator
	DefaultManifestGenerator = RandManifestGenerator

	// ManifestGenerators is a list of all available manifest generators
	ManifestGenerators = []struct {
		Name      string
		Generator ManifestGenerator
	}{
		{"overflow", OverflowManifestGenerator},
		{"random", RandManifestGenerator},
		{"app", AppManifestGenerator},
	}
)

// ManifestGenerator is an interface for generating test manifests
type ManifestGenerator interface {
	Manifest(t testing.TB) manifest.Manifest
	Group(t testing.TB) manifest.Group
	Service(t testing.TB) manifest.Service
	ServiceExpose(t testing.TB) manifest.ServiceExpose
}
