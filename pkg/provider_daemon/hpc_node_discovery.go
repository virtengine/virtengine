package provider_daemon

import (
	"context"

	"github.com/virtengine/virtengine/pkg/slurm_adapter"
)

// HPCSLURMNodeDiscoverer discovers nodes via a SLURM adapter.
type HPCSLURMNodeDiscoverer struct {
	adapter    *slurm_adapter.SLURMAdapter
	clusterID  string
	region     string
	datacenter string
}

// NewHPCSLURMNodeDiscoverer creates a SLURM-backed node discoverer.
func NewHPCSLURMNodeDiscoverer(adapter *slurm_adapter.SLURMAdapter, clusterID, region, datacenter string) *HPCSLURMNodeDiscoverer {
	return &HPCSLURMNodeDiscoverer{
		adapter:    adapter,
		clusterID:  clusterID,
		region:     region,
		datacenter: datacenter,
	}
}

// ListNodes returns nodes discovered from SLURM.
func (d *HPCSLURMNodeDiscoverer) ListNodes(ctx context.Context) ([]HPCDiscoveredNode, error) {
	nodes, err := d.adapter.ListNodes(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]HPCDiscoveredNode, 0, len(nodes))
	for _, node := range nodes {
		memoryGB := node.MemoryMB / 1024
		memoryGB32 := clampInt64ToInt32(memoryGB)

		capacity := &HPCNodeCapacity{
			CPUCoresTotal:     node.CPUs,
			CPUCoresAvailable: node.CPUs,
			MemoryGBTotal:     memoryGB32,
			MemoryGBAvailable: memoryGB32,
			GPUsTotal:         node.GPUs,
			GPUsAvailable:     node.GPUs,
			GPUType:           node.GPUType,
		}

		hardware := &HPCNodeHardware{
			GPUModel: node.GPUType,
			Features: node.Features,
		}

		locality := &HPCNodeLocality{
			Region:     d.region,
			Datacenter: d.datacenter,
		}

		out = append(out, HPCDiscoveredNode{
			NodeID:     node.Name,
			ClusterID:  d.clusterID,
			Region:     d.region,
			Datacenter: d.datacenter,
			Capacity:   capacity,
			Hardware:   hardware,
			Locality:   locality,
		})
	}

	return out, nil
}

func clampInt64ToInt32(value int64) int32 {
	if value < 0 {
		return 0
	}
	if value > int64(^uint32(0)>>1) {
		return int32(^uint32(0) >> 1)
	}
	return int32(value)
}
