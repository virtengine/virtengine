// Package types contains types for the HPC module.
//
// VE-500: SLURM cluster lifecycle module
package types

import (
	"cosmossdk.io/errors"
)

// HPC module sentinel errors
var (
	// ErrInvalidCluster is returned when a cluster is invalid
	ErrInvalidCluster = errors.Register(ModuleName, 2100, "invalid cluster")

	// ErrClusterNotFound is returned when a cluster is not found
	ErrClusterNotFound = errors.Register(ModuleName, 2101, "cluster not found")

	// ErrClusterAlreadyExists is returned when a cluster already exists
	ErrClusterAlreadyExists = errors.Register(ModuleName, 2102, "cluster already exists")

	// ErrInvalidOffering is returned when an offering is invalid
	ErrInvalidOffering = errors.Register(ModuleName, 2103, "invalid HPC offering")

	// ErrOfferingNotFound is returned when an offering is not found
	ErrOfferingNotFound = errors.Register(ModuleName, 2104, "HPC offering not found")

	// ErrInvalidJob is returned when a job is invalid
	ErrInvalidJob = errors.Register(ModuleName, 2105, "invalid HPC job")

	// ErrJobNotFound is returned when a job is not found
	ErrJobNotFound = errors.Register(ModuleName, 2106, "HPC job not found")

	// ErrInvalidJobState is returned when a job state transition is invalid
	ErrInvalidJobState = errors.Register(ModuleName, 2107, "invalid job state transition")

	// ErrInsufficientIdentityScore is returned when identity score is too low
	ErrInsufficientIdentityScore = errors.Register(ModuleName, 2108, "insufficient identity score")

	// ErrInvalidNodeMetadata is returned when node metadata is invalid
	ErrInvalidNodeMetadata = errors.Register(ModuleName, 2109, "invalid node metadata")

	// ErrNodeNotFound is returned when a node is not found
	ErrNodeNotFound = errors.Register(ModuleName, 2110, "node not found")

	// ErrInvalidSchedulingDecision is returned when a scheduling decision is invalid
	ErrInvalidSchedulingDecision = errors.Register(ModuleName, 2111, "invalid scheduling decision")

	// ErrNoAvailableCluster is returned when no cluster is available for scheduling
	ErrNoAvailableCluster = errors.Register(ModuleName, 2112, "no available cluster for job")

	// ErrInvalidReward is returned when a reward is invalid
	ErrInvalidReward = errors.Register(ModuleName, 2113, "invalid HPC reward")

	// ErrInvalidDispute is returned when a dispute is invalid
	ErrInvalidDispute = errors.Register(ModuleName, 2114, "invalid dispute")

	// ErrDisputeNotFound is returned when a dispute is not found
	ErrDisputeNotFound = errors.Register(ModuleName, 2115, "dispute not found")

	// ErrJobAccountingNotFound is returned when job accounting is not found
	ErrJobAccountingNotFound = errors.Register(ModuleName, 2116, "job accounting not found")

	// ErrInvalidJobAccounting is returned when job accounting is invalid
	ErrInvalidJobAccounting = errors.Register(ModuleName, 2117, "invalid job accounting")

	// ErrUnauthorized is returned for unauthorized operations
	ErrUnauthorized = errors.Register(ModuleName, 2118, "unauthorized")

	// ErrMaxRuntimeExceeded is returned when job exceeds max runtime
	ErrMaxRuntimeExceeded = errors.Register(ModuleName, 2119, "maximum runtime exceeded")

	// ErrInvalidWorkload is returned when a workload configuration is invalid
	ErrInvalidWorkload = errors.Register(ModuleName, 2120, "invalid workload configuration")

	// ErrInvalidClusterTemplate is returned when a cluster template is invalid
	ErrInvalidClusterTemplate = errors.Register(ModuleName, 2121, "invalid cluster template")

	// ErrInvalidPartition is returned when a partition configuration is invalid
	ErrInvalidPartition = errors.Register(ModuleName, 2122, "invalid partition configuration")

	// ErrInvalidQoSPolicy is returned when a QoS policy is invalid
	ErrInvalidQoSPolicy = errors.Register(ModuleName, 2123, "invalid QoS policy")

	// ErrInvalidHardwareClass is returned when a hardware class is invalid
	ErrInvalidHardwareClass = errors.Register(ModuleName, 2124, "invalid hardware class")

	// ErrInvalidSchedulingPolicy is returned when a scheduling policy is invalid
	ErrInvalidSchedulingPolicy = errors.Register(ModuleName, 2125, "invalid scheduling policy")

	// ErrInvalidHeartbeat is returned when a node heartbeat is invalid
	ErrInvalidHeartbeat = errors.Register(ModuleName, 2126, "invalid node heartbeat")

	// ErrStaleHeartbeat is returned when a heartbeat has an old sequence number
	ErrStaleHeartbeat = errors.Register(ModuleName, 2127, "stale heartbeat sequence")

	// ErrNodeDeregistered is returned when trying to update a deregistered node
	ErrNodeDeregistered = errors.Register(ModuleName, 2128, "node is deregistered")

	// ErrInvalidNodeIdentity is returned when a node identity is invalid
	ErrInvalidNodeIdentity = errors.Register(ModuleName, 2129, "invalid node identity")

	// ErrHeartbeatAuthFailed is returned when heartbeat authentication fails
	ErrHeartbeatAuthFailed = errors.Register(ModuleName, 2130, "heartbeat authentication failed")
)
