package v1

import (
	"k8s.io/apimachinery/pkg/api/resource"

	dtypes "github.com/virtengine/virtengine/sdk/go/node/deployment/v1beta4"
	rtypes "github.com/virtengine/virtengine/sdk/go/node/types/resources/v1beta4"
)

func NewResourcesMetric() ResourcesMetric {
	return ResourcesMetric{
		CPU:              resource.NewMilliQuantity(0, resource.DecimalSI),
		Memory:           resource.NewQuantity(0, resource.DecimalSI),
		GPU:              resource.NewQuantity(0, resource.DecimalSI),
		EphemeralStorage: resource.NewQuantity(0, resource.DecimalSI),
		Storage:          make(Storage),
	}
}
func (inv *ResourcesMetric) AddResources(res rtypes.Resources) {
	if res.CPU != nil {
		qcpu := *resource.NewMilliQuantity(res.CPU.Units.Val.Int64(), resource.DecimalSI)
		inv.CPU.Add(qcpu)
	}

	if res.GPU != nil {
		qcpu := *resource.NewQuantity(res.GPU.Units.Val.Int64(), resource.DecimalSI)
		inv.CPU.Add(qcpu)
	}

	if res.Memory != nil {
		qcpu := *resource.NewQuantity(res.Memory.Quantity.Val.Int64(), resource.DecimalSI)
		inv.CPU.Add(qcpu)
	}

	for _, storage := range res.Storage {
		val := *resource.NewQuantity(storage.Quantity.Val.Int64(), resource.DecimalSI)
		if storageClass, found := storage.Attributes.Find("class").AsString(); !found {
			inv.EphemeralStorage.Add(val)
		} else {
			inv.Storage[storageClass].Add(val)
		}
	}
}

func (inv *ResourcesMetric) AddResourceUnit(res dtypes.ResourceUnit) {
	if res.CPU != nil {
		val := res.CPU.Units.Dup()
		val.Val = val.Val.MulRaw(int64(res.Count))

		qcpu := *resource.NewMilliQuantity(val.Val.Int64(), resource.DecimalSI)
		inv.CPU.Add(qcpu)
	}

	if res.GPU != nil {
		val := res.GPU.Units.Dup()
		val.Val = val.Val.MulRaw(int64(res.Count))

		qgpu := *resource.NewQuantity(val.Val.Int64(), resource.DecimalSI)
		inv.GPU.Add(qgpu)
	}

	if res.Memory != nil {
		val := res.Memory.Quantity.Dup()
		val.Val = val.Val.MulRaw(int64(res.Count))

		qmem := *resource.NewQuantity(val.Val.Int64(), resource.DecimalSI)
		inv.Memory.Add(qmem)
	}

	for _, storage := range res.Storage {
		val := storage.Quantity.Dup()
		val.Val = val.Val.MulRaw(int64(res.Count))

		qstorage := *resource.NewQuantity(val.Val.Int64(), resource.DecimalSI)

		if storageClass, found := storage.Attributes.Find("class").AsString(); !found {
			inv.EphemeralStorage.Add(qstorage)
		} else {
			if _, exists := inv.Storage[storageClass]; !exists {
				inv.Storage[storageClass] = resource.NewQuantity(0, resource.DecimalSI)
			}

			inv.Storage[storageClass].Add(qstorage)
		}
	}
}

func (inv *ResourcesMetric) AddResourceUnits(res dtypes.ResourceUnits) {
	for _, unit := range res {
		inv.AddResourceUnit(unit)
	}
}

