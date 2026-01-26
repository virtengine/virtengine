// Package types contains types for the HPC module.
//
// VE-500: SLURM cluster lifecycle module
package types

const (
	// ModuleName is the name of the HPC module
	ModuleName = "hpc"

	// StoreKey is the store key for the HPC module
	StoreKey = ModuleName

	// RouterKey is the router key for the HPC module
	RouterKey = ModuleName

	// QuerierRoute is the querier route for the HPC module
	QuerierRoute = ModuleName
)

// Key prefixes for HPC store
var (
	// ClusterPrefix is the prefix for cluster storage
	ClusterPrefix = []byte{0x01}

	// OfferingPrefix is the prefix for HPC offering storage
	OfferingPrefix = []byte{0x02}

	// JobPrefix is the prefix for HPC job storage
	JobPrefix = []byte{0x03}

	// JobAccountingPrefix is the prefix for job accounting storage
	JobAccountingPrefix = []byte{0x04}

	// NodeMetadataPrefix is the prefix for node metadata storage
	NodeMetadataPrefix = []byte{0x05}

	// ClusterMembershipPrefix is the prefix for cluster membership
	ClusterMembershipPrefix = []byte{0x06}

	// SchedulingDecisionPrefix is the prefix for scheduling decision trail
	SchedulingDecisionPrefix = []byte{0x07}

	// HPCRewardPrefix is the prefix for HPC reward records
	HPCRewardPrefix = []byte{0x08}

	// DisputePrefix is the prefix for dispute records
	DisputePrefix = []byte{0x09}

	// ParamsKey is the key for module parameters
	ParamsKey = []byte{0x10}

	// SequenceKeyCluster is the sequence key for clusters
	SequenceKeyCluster = []byte{0x20}

	// SequenceKeyOffering is the sequence key for offerings
	SequenceKeyOffering = []byte{0x21}

	// SequenceKeyJob is the sequence key for jobs
	SequenceKeyJob = []byte{0x22}

	// SequenceKeyDecision is the sequence key for scheduling decisions
	SequenceKeyDecision = []byte{0x23}

	// SequenceKeyDispute is the sequence key for disputes
	SequenceKeyDispute = []byte{0x24}
)

// GetClusterKey returns the key for a cluster
func GetClusterKey(clusterID string) []byte {
	return append(ClusterPrefix, []byte(clusterID)...)
}

// GetOfferingKey returns the key for an offering
func GetOfferingKey(offeringID string) []byte {
	return append(OfferingPrefix, []byte(offeringID)...)
}

// GetJobKey returns the key for a job
func GetJobKey(jobID string) []byte {
	return append(JobPrefix, []byte(jobID)...)
}

// GetJobAccountingKey returns the key for job accounting
func GetJobAccountingKey(jobID string) []byte {
	return append(JobAccountingPrefix, []byte(jobID)...)
}

// GetNodeMetadataKey returns the key for node metadata
func GetNodeMetadataKey(nodeID string) []byte {
	return append(NodeMetadataPrefix, []byte(nodeID)...)
}

// GetSchedulingDecisionKey returns the key for a scheduling decision
func GetSchedulingDecisionKey(decisionID string) []byte {
	return append(SchedulingDecisionPrefix, []byte(decisionID)...)
}

// GetHPCRewardKey returns the key for HPC reward records
func GetHPCRewardKey(rewardID string) []byte {
	return append(HPCRewardPrefix, []byte(rewardID)...)
}

// GetDisputeKey returns the key for a dispute
func GetDisputeKey(disputeID string) []byte {
	return append(DisputePrefix, []byte(disputeID)...)
}
