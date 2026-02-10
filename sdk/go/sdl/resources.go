package sdl

import (
	types "github.com/virtengine/virtengine/sdk/go/node/types/attributes/v1"
	rtypes "github.com/virtengine/virtengine/sdk/go/node/types/resources/v1beta4"
)

type v2ComputeResources struct {
	CPU     *v2ResourceCPU         `yaml:"cpu"`
	GPU     *v2ResourceGPU         `yaml:"gpu"`
	Memory  *v2ResourceMemory      `yaml:"memory"`
	Storage v2ResourceStorageArray `yaml:"storage"`
}

func (sdl *v2ComputeResources) toResources() rtypes.Resources {
	if sdl == nil {
		return rtypes.Resources{}
	}

	units := rtypes.Resources{
		Endpoints: rtypes.Endpoints{},
	}

	if sdl.CPU != nil {
		units.CPU = &rtypes.CPU{
			Units:      rtypes.NewResourceValue(uint64(sdl.CPU.Units)),
			Attributes: types.Attributes(sdl.CPU.Attributes),
		}
	}

	if sdl.GPU != nil {
		units.GPU = &rtypes.GPU{
			Units:      rtypes.NewResourceValue(uint64(sdl.GPU.Units)),
			Attributes: types.Attributes(sdl.GPU.Attributes),
		}
	} else {
		units.GPU = &rtypes.GPU{
			Units: rtypes.NewResourceValue(0),
		}
	}

	if sdl.Memory != nil {
		units.Memory = &rtypes.Memory{
			Quantity:   rtypes.NewResourceValue(uint64(sdl.Memory.Quantity)),
			Attributes: types.Attributes(sdl.Memory.Attributes),
		}
	}

	for _, storage := range sdl.Storage {
		storageEntry := rtypes.Storage{
			Name:       storage.Name,
			Quantity:   rtypes.NewResourceValue(uint64(storage.Quantity)),
			Attributes: types.Attributes(storage.Attributes),
		}

		units.Storage = append(units.Storage, storageEntry)
	}

	return units
}
