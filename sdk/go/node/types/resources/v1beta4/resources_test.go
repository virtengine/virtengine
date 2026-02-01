package v1beta4

import (
	"testing"

	"github.com/stretchr/testify/require"

	attr "github.com/virtengine/virtengine/sdk/go/node/types/attributes/v1"
)

func TestVolumes_Dup(t *testing.T) {
	volumes := Volumes{
		Storage{
			Name:       "default",
			Quantity:   NewResourceValue(100),
			Attributes: attr.Attributes{},
		},
	}

	dVolumes := volumes.Dup()

	require.Equal(t, volumes, dVolumes)
}

func TestVolumes_Equal(t *testing.T) {
	volumes := Volumes{
		Storage{
			Name:       "default",
			Quantity:   NewResourceValue(100),
			Attributes: attr.Attributes{},
		},
	}

	dVolumes := volumes.Dup()

	require.True(t, volumes.Equal(dVolumes))

	dVolumes[0].Name = "class2"
	require.False(t, volumes.Equal(dVolumes))
}

func TestResources_ValidateID(t *testing.T) {
	res := Resources{}

	err := res.Validate()
	require.ErrorContains(t, err, "resources ID must be > 0")
}

func TestResources_ValidateCPU(t *testing.T) {
	res := Resources{
		ID: 1,
	}

	err := res.Validate()
	require.ErrorContains(t, err, "CPU must not be nil")
}

func TestResources_ValidateGPU(t *testing.T) {
	res := Resources{
		ID:  1,
		CPU: &CPU{},
	}

	err := res.Validate()
	require.ErrorContains(t, err, "GPU must not be nil")
}

func TestResources_ValidateMemory(t *testing.T) {
	res := Resources{
		ID:  1,
		CPU: &CPU{},
		GPU: &GPU{},
	}

	err := res.Validate()
	require.ErrorContains(t, err, "memory must not be nil")
}

func TestResources_ValidateStorage(t *testing.T) {
	res := Resources{
		ID:     1,
		CPU:    &CPU{},
		GPU:    &GPU{},
		Memory: &Memory{},
	}

	err := res.Validate()
	require.ErrorContains(t, err, "storage must not be nil")
}

func TestResources_ValidateEndpoints(t *testing.T) {
	res := Resources{
		ID:      1,
		CPU:     &CPU{},
		GPU:     &GPU{},
		Memory:  &Memory{},
		Storage: make(Volumes, 0),
	}

	err := res.Validate()
	require.ErrorContains(t, err, "endpoints must not be nil")
}

func TestResources_Validate(t *testing.T) {
	res := Resources{
		ID:        1,
		CPU:       &CPU{},
		GPU:       &GPU{},
		Memory:    &Memory{},
		Storage:   make(Volumes, 0),
		Endpoints: make(Endpoints, 0),
	}

	err := res.Validate()
	require.NoError(t, err)
}

func TestResources_DupInvalidID(t *testing.T) {
	res := Resources{
		CPU:       &CPU{},
		GPU:       &GPU{},
		Memory:    &Memory{},
		Storage:   make(Volumes, 0),
		Endpoints: make(Endpoints, 0),
	}

	dup := res.Dup()
	err := dup.Validate()
	require.ErrorContains(t, err, "resources ID must be > 0")
}

func TestResources_DupValid(t *testing.T) {
	res := Resources{
		ID:        1,
		CPU:       &CPU{},
		GPU:       &GPU{},
		Memory:    &Memory{},
		Storage:   make(Volumes, 0),
		Endpoints: make(Endpoints, 0),
	}

	dup := res.Dup()
	err := dup.Validate()
	require.NoError(t, err)
}

