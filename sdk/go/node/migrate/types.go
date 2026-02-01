package migrate

import (
	v1 "github.com/virtengine/virtengine/sdk/go/node/types/attributes/v1"
	"github.com/virtengine/virtengine/sdk/go/node/types/resources/v1beta4"
	"github.com/virtengine/virtengine/sdk/go/node/types/v1beta3"
)

func AttributesFromV1Beta3(from v1beta3.Attributes) v1.Attributes {
	res := make(v1.Attributes, 0, len(from))

	for _, attr := range from {
		res = append(res, v1.Attribute{
			Key:   attr.Key,
			Value: attr.Value,
		})
	}

	return res
}

func SignedByFromV1Beta3(from v1beta3.SignedBy) v1.SignedBy {
	return v1.SignedBy{
		AllOf: from.AllOf,
		AnyOf: from.AnyOf,
	}
}

func PlacementRequirementsFromV1Beta3(from v1beta3.PlacementRequirements) v1.PlacementRequirements {
	res := v1.PlacementRequirements{
		SignedBy:   SignedByFromV1Beta3(from.SignedBy),
		Attributes: AttributesFromV1Beta3(from.Attributes),
	}

	return res
}

func ResourceValueFromV1Beta3(from v1beta3.ResourceValue) v1beta4.ResourceValue {
	return v1beta4.NewResourceValue(from.Value())
}

func CPUFromV1Beta3(from *v1beta3.CPU) *v1beta4.CPU {
	if from == nil {
		return nil
	}

	return &v1beta4.CPU{
		Units:      ResourceValueFromV1Beta3(from.Units),
		Attributes: AttributesFromV1Beta3(from.Attributes),
	}
}

func GPUFromV1Beta3(from *v1beta3.GPU) *v1beta4.GPU {
	if from == nil {
		return nil
	}

	return &v1beta4.GPU{
		Units:      ResourceValueFromV1Beta3(from.Units),
		Attributes: AttributesFromV1Beta3(from.Attributes),
	}
}

func MemoryFromV1Beta3(from *v1beta3.Memory) *v1beta4.Memory {
	if from == nil {
		return nil
	}

	return &v1beta4.Memory{
		Quantity:   ResourceValueFromV1Beta3(from.Quantity),
		Attributes: AttributesFromV1Beta3(from.Attributes),
	}
}

func VolumesFromV1Beta3(from v1beta3.Volumes) v1beta4.Volumes {
	res := make(v1beta4.Volumes, 0, len(from))

	for _, storage := range from {
		res = append(res, v1beta4.Storage{
			Name:       "default",
			Quantity:   ResourceValueFromV1Beta3(storage.Quantity),
			Attributes: AttributesFromV1Beta3(storage.Attributes),
		})
	}

	return res
}

func EndpointsFromV1Beta3(from []v1beta3.Endpoint) []v1beta4.Endpoint {
	res := make([]v1beta4.Endpoint, 0, len(from))

	for _, endpoint := range from {
		res = append(res, v1beta4.Endpoint{
			Kind:           v1beta4.Endpoint_Kind(endpoint.Kind),
			SequenceNumber: endpoint.SequenceNumber,
		})
	}

	return res
}

func ResourcesFromV1Beta3(id uint32, from v1beta3.Resources) v1beta4.Resources {
	return v1beta4.Resources{
		ID:        id,
		CPU:       CPUFromV1Beta3(from.CPU),
		GPU:       GPUFromV1Beta3(from.GPU),
		Memory:    MemoryFromV1Beta3(from.Memory),
		Storage:   VolumesFromV1Beta3(from.Storage),
		Endpoints: EndpointsFromV1Beta3(from.Endpoints),
	}
}

