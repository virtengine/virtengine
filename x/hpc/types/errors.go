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

	// ErrInvalidSchedulingMetrics is returned when scheduling metrics are invalid
	ErrInvalidSchedulingMetrics = errors.Register(ModuleName, 2154, "invalid scheduling metrics")

	// ErrNoAvailableCluster is returned when no cluster is available for scheduling
	ErrNoAvailableCluster = errors.Register(ModuleName, 2112, "no available cluster for job")

	// ErrTenantQuotaExceeded is returned when tenant quota or burst limits are exceeded
	ErrTenantQuotaExceeded = errors.Register(ModuleName, 2153, "tenant quota exceeded")

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

	// ErrInvalidRoutingAudit is returned when a routing audit record is invalid
	ErrInvalidRoutingAudit = errors.Register(ModuleName, 2131, "invalid routing audit record")

	// ErrInvalidRoutingViolation is returned when a routing violation is invalid
	ErrInvalidRoutingViolation = errors.Register(ModuleName, 2132, "invalid routing violation")

	// ErrRoutingDecisionStale is returned when a scheduling decision is too old
	ErrRoutingDecisionStale = errors.Register(ModuleName, 2133, "scheduling decision is stale")

	// ErrRoutingDecisionNotFound is returned when a scheduling decision is not found
	ErrRoutingDecisionNotFound = errors.Register(ModuleName, 2134, "scheduling decision not found")

	// ErrRoutingClusterMismatch is returned when job is placed on wrong cluster
	ErrRoutingClusterMismatch = errors.Register(ModuleName, 2135, "routing cluster mismatch")

	// ErrRoutingEnforcementFailed is returned when routing enforcement fails
	ErrRoutingEnforcementFailed = errors.Register(ModuleName, 2136, "routing enforcement failed")

	// ErrFallbackNotAuthorized is returned when fallback routing is not authorized
	ErrFallbackNotAuthorized = errors.Register(ModuleName, 2137, "fallback routing not authorized")

	// ErrMissingSchedulingDecision is returned when job has no scheduling decision
	ErrMissingSchedulingDecision = errors.Register(ModuleName, 2138, "missing scheduling decision")

	// ErrInvalidWorkloadTemplate is returned when a workload template is invalid (VE-5F)
	ErrInvalidWorkloadTemplate = errors.Register(ModuleName, 2139, "invalid workload template")

	// ErrWorkloadTemplateNotFound is returned when a workload template is not found (VE-5F)
	ErrWorkloadTemplateNotFound = errors.Register(ModuleName, 2140, "workload template not found")

	// ErrWorkloadTemplateNotApproved is returned when trying to use an unapproved template (VE-5F)
	ErrWorkloadTemplateNotApproved = errors.Register(ModuleName, 2141, "workload template not approved")

	// ErrInvalidWorkloadRuntime is returned when workload runtime config is invalid (VE-5F)
	ErrInvalidWorkloadRuntime = errors.Register(ModuleName, 2142, "invalid workload runtime")

	// ErrInvalidWorkloadResources is returned when workload resources are invalid (VE-5F)
	ErrInvalidWorkloadResources = errors.Register(ModuleName, 2143, "invalid workload resources")

	// ErrInvalidWorkloadSecurity is returned when workload security config is invalid (VE-5F)
	ErrInvalidWorkloadSecurity = errors.Register(ModuleName, 2144, "invalid workload security")

	// ErrInvalidWorkloadEntrypoint is returned when workload entrypoint is invalid (VE-5F)
	ErrInvalidWorkloadEntrypoint = errors.Register(ModuleName, 2145, "invalid workload entrypoint")

	// ErrInvalidWorkloadParameter is returned when a workload parameter is invalid (VE-5F)
	ErrInvalidWorkloadParameter = errors.Register(ModuleName, 2146, "invalid workload parameter")

	// ErrInvalidDataBinding is returned when a data binding is invalid (VE-5F)
	ErrInvalidDataBinding = errors.Register(ModuleName, 2147, "invalid data binding")

	// ErrInvalidWorkloadSignature is returned when workload signature is invalid (VE-5F)
	ErrInvalidWorkloadSignature = errors.Register(ModuleName, 2148, "invalid workload signature")

	// ErrWorkloadValidationFailed is returned when workload validation fails (VE-5F)
	ErrWorkloadValidationFailed = errors.Register(ModuleName, 2149, "workload validation failed")

	// ErrImageNotAllowed is returned when a container image is not allowed (VE-5F)
	ErrImageNotAllowed = errors.Register(ModuleName, 2150, "container image not allowed")

	// ErrResourceLimitExceeded is returned when resource limits are exceeded (VE-5F)
	ErrResourceLimitExceeded = errors.Register(ModuleName, 2151, "resource limit exceeded")

	// ErrWorkloadGovernanceFailed is returned when governance action fails (VE-5F)
	ErrWorkloadGovernanceFailed = errors.Register(ModuleName, 2152, "workload governance action failed")
)
