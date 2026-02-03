package v1

// Nodes is a collection of Node.
type Nodes []Node

// ClusterStorage is a collection of Storage for the cluster.
type ClusterStorage []Storage

// CPUInfoS is a collection of CPUInfo.
type CPUInfoS []CPUInfo

// GPUInfoS is a collection of GPUInfo.
type GPUInfoS []GPUInfo

// MemoryInfoS is a collection of MemoryInfo.
type MemoryInfoS []MemoryInfo

// InventoryMetrics represents inventory metrics for a provider cluster.
// It contains the cluster state and aggregated resource information.
type InventoryMetrics struct {
	Cluster     *Cluster `json:"cluster,omitempty"`
	Active      uint64   `json:"active"`
	Pending     uint64   `json:"pending"`
	Available   uint64   `json:"available"`
	Allocatable uint64   `json:"allocatable"`
}
