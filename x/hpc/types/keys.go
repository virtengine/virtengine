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

	// ClusterTemplatePrefix is the prefix for cluster template storage
	ClusterTemplatePrefix = []byte{0x11}

	// NodeHeartbeatPrefix is the prefix for node heartbeat storage
	NodeHeartbeatPrefix = []byte{0x12}

	// AccountingRecordPrefix is the prefix for HPC accounting records (VE-5A)
	AccountingRecordPrefix = []byte{0x13}

	// UsageSnapshotPrefix is the prefix for usage snapshots (VE-5A)
	UsageSnapshotPrefix = []byte{0x14}

	// ReconciliationPrefix is the prefix for reconciliation records (VE-5A)
	ReconciliationPrefix = []byte{0x15}

	// AuditTrailPrefix is the prefix for audit trail entries (VE-5A)
	AuditTrailPrefix = []byte{0x16}

	// BillingRulesPrefix is the prefix for billing rules (VE-5A)
	BillingRulesPrefix = []byte{0x17}

	// AggregationPrefix is the prefix for accounting aggregations (VE-5A)
	AggregationPrefix = []byte{0x18}

	// RoutingAuditPrefix is the prefix for routing audit records (VE-5B)
	RoutingAuditPrefix = []byte{0x19}

	// RoutingViolationPrefix is the prefix for routing violation records (VE-5B)
	RoutingViolationPrefix = []byte{0x1A}

	// WorkloadTemplatePrefix is the prefix for workload template storage (VE-5F)
	WorkloadTemplatePrefix = []byte{0x1B}

	// WorkloadTemplateVersionPrefix is the prefix for versioned template storage (VE-5F)
	WorkloadTemplateVersionPrefix = []byte{0x1C}

	// WorkloadTemplateByTypePrefix is the prefix for templates indexed by type (VE-5F)
	WorkloadTemplateByTypePrefix = []byte{0x1D}

	// WorkloadProposalPrefix is the prefix for workload governance proposals (VE-5F)
	WorkloadProposalPrefix = []byte{0x1E}

	// WorkloadVotePrefix is the prefix for workload votes (VE-5F)
	WorkloadVotePrefix = []byte{0x1F}

	// SchedulingMetricsPrefix is the prefix for scheduling metrics (VE-78C)
	SchedulingMetricsPrefix = []byte{0x2E}

	// WorkloadGovernanceParamsKey is the key for governance params (VE-5F)
	WorkloadGovernanceParamsKey = []byte{0x20}

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

	// SequenceKeyAccountingRecord is the sequence key for accounting records (VE-5A)
	SequenceKeyAccountingRecord = []byte{0x25}

	// SequenceKeySnapshot is the sequence key for usage snapshots (VE-5A)
	SequenceKeySnapshot = []byte{0x26}

	// SequenceKeyReconciliation is the sequence key for reconciliations (VE-5A)
	SequenceKeyReconciliation = []byte{0x27}

	// SequenceKeyAuditTrail is the sequence key for audit trail entries (VE-5A)
	SequenceKeyAuditTrail = []byte{0x28}

	// SequenceKeyAggregation is the sequence key for aggregations (VE-5A)
	SequenceKeyAggregation = []byte{0x29}

	// SequenceKeyRoutingAudit is the sequence key for routing audit records (VE-5B)
	SequenceKeyRoutingAudit = []byte{0x2A}

	// SequenceKeyRoutingViolation is the sequence key for routing violations (VE-5B)
	SequenceKeyRoutingViolation = []byte{0x2B}

	// SequenceKeyWorkloadTemplate is the sequence key for workload templates (VE-5F)
	SequenceKeyWorkloadTemplate = []byte{0x2C}

	// SequenceKeyWorkloadProposal is the sequence key for workload proposals (VE-5F)
	SequenceKeyWorkloadProposal = []byte{0x2D}
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

// GetClusterTemplateKey returns the key for a cluster template
func GetClusterTemplateKey(templateID string) []byte {
	return append(ClusterTemplatePrefix, []byte(templateID)...)
}

// GetNodeHeartbeatKey returns the key for a node heartbeat
func GetNodeHeartbeatKey(nodeID string) []byte {
	return append(NodeHeartbeatPrefix, []byte(nodeID)...)
}

// GetAccountingRecordKey returns the key for an accounting record (VE-5A)
func GetAccountingRecordKey(recordID string) []byte {
	return append(AccountingRecordPrefix, []byte(recordID)...)
}

// GetUsageSnapshotKey returns the key for a usage snapshot (VE-5A)
func GetUsageSnapshotKey(snapshotID string) []byte {
	return append(UsageSnapshotPrefix, []byte(snapshotID)...)
}

// GetReconciliationKey returns the key for a reconciliation record (VE-5A)
func GetReconciliationKey(reconciliationID string) []byte {
	return append(ReconciliationPrefix, []byte(reconciliationID)...)
}

// GetAuditTrailKey returns the key for an audit trail entry (VE-5A)
func GetAuditTrailKey(entryID string) []byte {
	return append(AuditTrailPrefix, []byte(entryID)...)
}

// GetBillingRulesKey returns the key for billing rules (VE-5A)
func GetBillingRulesKey(providerAddr string) []byte {
	return append(BillingRulesPrefix, []byte(providerAddr)...)
}

// GetAggregationKey returns the key for an aggregation (VE-5A)
func GetAggregationKey(aggregationID string) []byte {
	return append(AggregationPrefix, []byte(aggregationID)...)
}

// GetJobUsageSnapshotsKey returns prefix for all snapshots of a job (VE-5A)
func GetJobUsageSnapshotsKey(jobID string) []byte {
	return append(UsageSnapshotPrefix, append([]byte(jobID), '/')...)
}

// GetJobAccountingRecordsKey returns prefix for all accounting records of a job (VE-5A)
func GetJobAccountingRecordsKey(jobID string) []byte {
	return append(AccountingRecordPrefix, append([]byte(jobID), '/')...)
}

// GetRoutingAuditKey returns the key for a routing audit record (VE-5B)
func GetRoutingAuditKey(recordID string) []byte {
	return append(RoutingAuditPrefix, []byte(recordID)...)
}

// GetRoutingViolationKey returns the key for a routing violation (VE-5B)
func GetRoutingViolationKey(violationID string) []byte {
	return append(RoutingViolationPrefix, []byte(violationID)...)
}

// GetWorkloadTemplateKey returns the key for a workload template (VE-5F)
func GetWorkloadTemplateKey(templateID string) []byte {
	return append(WorkloadTemplatePrefix, []byte(templateID)...)
}

// GetWorkloadTemplateVersionKey returns the key for a versioned workload template (VE-5F)
func GetWorkloadTemplateVersionKey(templateID, version string) []byte {
	return append(WorkloadTemplateVersionPrefix, []byte(templateID+"/"+version)...)
}

// GetWorkloadTemplateByTypeKey returns the key for templates indexed by type (VE-5F)
func GetWorkloadTemplateByTypeKey(workloadType WorkloadType, templateID string) []byte {
	return append(WorkloadTemplateByTypePrefix, []byte(string(workloadType)+"/"+templateID)...)
}

// GetWorkloadProposalKey returns the key for a governance proposal (VE-5F)
func GetWorkloadProposalKey(proposalID string) []byte {
	return append(WorkloadProposalPrefix, []byte(proposalID)...)
}

// GetWorkloadVoteKey returns the key for a vote (VE-5F)
func GetWorkloadVoteKey(proposalID, voter string) []byte {
	return append(WorkloadVotePrefix, []byte(proposalID+"/"+voter)...)
}

// GetSchedulingMetricsKey returns the key for scheduling metrics.
func GetSchedulingMetricsKey(clusterID, queueName string) []byte {
	return append(SchedulingMetricsPrefix, []byte(clusterID+"/"+queueName)...)
}
