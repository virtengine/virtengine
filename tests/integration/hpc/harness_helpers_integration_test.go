//go:build e2e.integration

package hpc

import hpcv1 "github.com/virtengine/virtengine/sdk/go/node/hpc/v1"

func realNodeCapacityFixture() *hpcv1.NodeCapacity {
	return &hpcv1.NodeCapacity{
		CpuCoresTotal:      64,
		CpuCoresAvailable:  44,
		CpuCoresAllocated:  20,
		MemoryGbTotal:      512,
		MemoryGbAvailable:  384,
		MemoryGbAllocated:  128,
		GpusTotal:          8,
		GpusAvailable:      6,
		GpusAllocated:      2,
		GpuType:            "NVIDIA H100",
		StorageGbTotal:     4000,
		StorageGbAvailable: 2750,
		StorageGbAllocated: 1250,
	}
}

func realNodeHealthFixture(status hpcv1.HealthStatus, slurmState string) *hpcv1.NodeHealth {
	return &hpcv1.NodeHealth{
		Status:                      status,
		UptimeSeconds:               7200,
		LoadAverage_1M:              "19.250000",
		LoadAverage_5M:              "17.500000",
		LoadAverage_15M:             "16.000000",
		CpuUtilizationPercent:       68,
		MemoryUtilizationPercent:    75,
		GpuUtilizationPercent:       81,
		GpuMemoryUtilizationPercent: 77,
		DiskIoUtilizationPercent:    43,
		NetworkUtilizationPercent:   36,
		GpuTemperatureCelsius:       72,
		SlurmState:                  slurmState,
	}
}

func realNodeHardwareFixture() *hpcv1.NodeHardware {
	return &hpcv1.NodeHardware{
		CpuModel:    "AMD EPYC 9654",
		CpuVendor:   "AMD",
		CpuArch:     "x86_64",
		GpuModel:    "NVIDIA H100",
		GpuMemoryGb: 80,
		StorageType: "NVMe",
		Features:    []string{"infiniband", "rdma", "gpu"},
	}
}

func realNodeLocalityFixture() *hpcv1.NodeLocality {
	return &hpcv1.NodeLocality{
		Region:     "us-east-1",
		Datacenter: "dc1",
		Zone:       "a",
		Rack:       "rack-7",
		Row:        "row-2",
		Position:   "u14",
	}
}
