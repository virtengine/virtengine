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
	ErrInvalidCluster = errors.Register(ModuleName, 2, "invalid cluster")

	// ErrClusterNotFound is returned when a cluster is not found
	ErrClusterNotFound = errors.Register(ModuleName, 3, "cluster not found")

	// ErrClusterAlreadyExists is returned when a cluster already exists
	ErrClusterAlreadyExists = errors.Register(ModuleName, 4, "cluster already exists")

	// ErrInvalidOffering is returned when an offering is invalid
	ErrInvalidOffering = errors.Register(ModuleName, 5, "invalid HPC offering")

	// ErrOfferingNotFound is returned when an offering is not found
	ErrOfferingNotFound = errors.Register(ModuleName, 6, "HPC offering not found")

	// ErrInvalidJob is returned when a job is invalid
	ErrInvalidJob = errors.Register(ModuleName, 7, "invalid HPC job")

	// ErrJobNotFound is returned when a job is not found
	ErrJobNotFound = errors.Register(ModuleName, 8, "HPC job not found")

	// ErrInvalidJobState is returned when a job state transition is invalid
	ErrInvalidJobState = errors.Register(ModuleName, 9, "invalid job state transition")

	// ErrInsufficientIdentityScore is returned when identity score is too low
	ErrInsufficientIdentityScore = errors.Register(ModuleName, 10, "insufficient identity score")

	// ErrInvalidNodeMetadata is returned when node metadata is invalid
	ErrInvalidNodeMetadata = errors.Register(ModuleName, 11, "invalid node metadata")

	// ErrNodeNotFound is returned when a node is not found
	ErrNodeNotFound = errors.Register(ModuleName, 12, "node not found")

	// ErrInvalidSchedulingDecision is returned when a scheduling decision is invalid
	ErrInvalidSchedulingDecision = errors.Register(ModuleName, 13, "invalid scheduling decision")

	// ErrNoAvailableCluster is returned when no cluster is available for scheduling
	ErrNoAvailableCluster = errors.Register(ModuleName, 14, "no available cluster for job")

	// ErrInvalidReward is returned when a reward is invalid
	ErrInvalidReward = errors.Register(ModuleName, 15, "invalid HPC reward")

	// ErrInvalidDispute is returned when a dispute is invalid
	ErrInvalidDispute = errors.Register(ModuleName, 16, "invalid dispute")

	// ErrDisputeNotFound is returned when a dispute is not found
	ErrDisputeNotFound = errors.Register(ModuleName, 17, "dispute not found")

	// ErrJobAccountingNotFound is returned when job accounting is not found
	ErrJobAccountingNotFound = errors.Register(ModuleName, 18, "job accounting not found")

	// ErrInvalidJobAccounting is returned when job accounting is invalid
	ErrInvalidJobAccounting = errors.Register(ModuleName, 19, "invalid job accounting")

	// ErrUnauthorized is returned for unauthorized operations
	ErrUnauthorized = errors.Register(ModuleName, 20, "unauthorized")

	// ErrMaxRuntimeExceeded is returned when job exceeds max runtime
	ErrMaxRuntimeExceeded = errors.Register(ModuleName, 21, "maximum runtime exceeded")

	// ErrInvalidWorkload is returned when a workload configuration is invalid
	ErrInvalidWorkload = errors.Register(ModuleName, 22, "invalid workload configuration")
)
