// Package types contains types for the HPC module.
package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgRegisterCluster registers a new HPC cluster
type MsgRegisterCluster struct {
	ProviderAddress string          `json:"provider_address"`
	Name            string          `json:"name"`
	Description     string          `json:"description,omitempty"`
	Region          string          `json:"region"`
	Partitions      []Partition     `json:"partitions"`
	TotalNodes      int32           `json:"total_nodes"`
	ClusterMetadata ClusterMetadata `json:"cluster_metadata"`
	SLURMVersion    string          `json:"slurm_version"`
}

// ValidateBasic performs basic validation
func (msg MsgRegisterCluster) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ProviderAddress); err != nil {
		return ErrInvalidCluster.Wrap("invalid provider address")
	}
	if msg.Name == "" {
		return ErrInvalidCluster.Wrap("name cannot be empty")
	}
	if msg.Region == "" {
		return ErrInvalidCluster.Wrap("region cannot be empty")
	}
	if msg.TotalNodes < 1 {
		return ErrInvalidCluster.Wrap("total_nodes must be at least 1")
	}
	return nil
}

// GetSigners returns the signers of the message
func (msg MsgRegisterCluster) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.ProviderAddress)
	return []sdk.AccAddress{addr}
}

// MsgUpdateCluster updates an existing HPC cluster
type MsgUpdateCluster struct {
	ProviderAddress string          `json:"provider_address"`
	ClusterID       string          `json:"cluster_id"`
	Name            string          `json:"name,omitempty"`
	Description     string          `json:"description,omitempty"`
	State           ClusterState    `json:"state,omitempty"`
	Partitions      []Partition     `json:"partitions,omitempty"`
	TotalNodes      int32           `json:"total_nodes,omitempty"`
	AvailableNodes  int32           `json:"available_nodes,omitempty"`
	ClusterMetadata ClusterMetadata `json:"cluster_metadata,omitempty"`
}

// ValidateBasic performs basic validation
func (msg MsgUpdateCluster) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ProviderAddress); err != nil {
		return ErrInvalidCluster.Wrap("invalid provider address")
	}
	if msg.ClusterID == "" {
		return ErrInvalidCluster.Wrap("cluster_id cannot be empty")
	}
	return nil
}

// GetSigners returns the signers of the message
func (msg MsgUpdateCluster) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.ProviderAddress)
	return []sdk.AccAddress{addr}
}

// MsgDeregisterCluster deregisters an HPC cluster
type MsgDeregisterCluster struct {
	ProviderAddress string `json:"provider_address"`
	ClusterID       string `json:"cluster_id"`
}

// ValidateBasic performs basic validation
func (msg MsgDeregisterCluster) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ProviderAddress); err != nil {
		return ErrInvalidCluster.Wrap("invalid provider address")
	}
	if msg.ClusterID == "" {
		return ErrInvalidCluster.Wrap("cluster_id cannot be empty")
	}
	return nil
}

// GetSigners returns the signers of the message
func (msg MsgDeregisterCluster) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.ProviderAddress)
	return []sdk.AccAddress{addr}
}

// MsgCreateOffering creates a new HPC offering
type MsgCreateOffering struct {
	ProviderAddress           string                  `json:"provider_address"`
	ClusterID                 string                  `json:"cluster_id"`
	Name                      string                  `json:"name"`
	Description               string                  `json:"description,omitempty"`
	QueueOptions              []QueueOption           `json:"queue_options"`
	Pricing                   HPCPricing              `json:"pricing"`
	RequiredIdentityThreshold int32                   `json:"required_identity_threshold"`
	MaxRuntimeSeconds         int64                   `json:"max_runtime_seconds"`
	PreconfiguredWorkloads    []PreconfiguredWorkload `json:"preconfigured_workloads,omitempty"`
	SupportsCustomWorkloads   bool                    `json:"supports_custom_workloads"`
}

// ValidateBasic performs basic validation
func (msg MsgCreateOffering) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ProviderAddress); err != nil {
		return ErrInvalidOffering.Wrap("invalid provider address")
	}
	if msg.ClusterID == "" {
		return ErrInvalidOffering.Wrap("cluster_id cannot be empty")
	}
	if msg.Name == "" {
		return ErrInvalidOffering.Wrap("name cannot be empty")
	}
	if len(msg.QueueOptions) == 0 {
		return ErrInvalidOffering.Wrap("at least one queue option is required")
	}
	return nil
}

// GetSigners returns the signers of the message
func (msg MsgCreateOffering) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.ProviderAddress)
	return []sdk.AccAddress{addr}
}

// MsgUpdateOffering updates an existing HPC offering
type MsgUpdateOffering struct {
	ProviderAddress           string                  `json:"provider_address"`
	OfferingID                string                  `json:"offering_id"`
	Name                      string                  `json:"name,omitempty"`
	Description               string                  `json:"description,omitempty"`
	QueueOptions              []QueueOption           `json:"queue_options,omitempty"`
	Pricing                   HPCPricing              `json:"pricing,omitempty"`
	RequiredIdentityThreshold int32                   `json:"required_identity_threshold,omitempty"`
	MaxRuntimeSeconds         int64                   `json:"max_runtime_seconds,omitempty"`
	PreconfiguredWorkloads    []PreconfiguredWorkload `json:"preconfigured_workloads,omitempty"`
	SupportsCustomWorkloads   *bool                   `json:"supports_custom_workloads,omitempty"`
	Active                    *bool                   `json:"active,omitempty"`
}

// ValidateBasic performs basic validation
func (msg MsgUpdateOffering) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ProviderAddress); err != nil {
		return ErrInvalidOffering.Wrap("invalid provider address")
	}
	if msg.OfferingID == "" {
		return ErrInvalidOffering.Wrap("offering_id cannot be empty")
	}
	return nil
}

// GetSigners returns the signers of the message
func (msg MsgUpdateOffering) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.ProviderAddress)
	return []sdk.AccAddress{addr}
}

// MsgSubmitJob submits a new HPC job
type MsgSubmitJob struct {
	CustomerAddress         string          `json:"customer_address"`
	OfferingID              string          `json:"offering_id"`
	QueueName               string          `json:"queue_name"`
	WorkloadSpec            JobWorkloadSpec `json:"workload_spec"`
	Resources               JobResources    `json:"resources"`
	DataReferences          []DataReference `json:"data_references,omitempty"`
	EncryptedInputsPointer  string          `json:"encrypted_inputs_pointer,omitempty"`
	EncryptedOutputsPointer string          `json:"encrypted_outputs_pointer,omitempty"`
	MaxRuntimeSeconds       int64           `json:"max_runtime_seconds"`
	MaxPrice                sdk.Coins       `json:"max_price"`
}

// ValidateBasic performs basic validation
func (msg MsgSubmitJob) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.CustomerAddress); err != nil {
		return ErrInvalidJob.Wrap("invalid customer address")
	}
	if msg.OfferingID == "" {
		return ErrInvalidJob.Wrap("offering_id cannot be empty")
	}
	if msg.QueueName == "" {
		return ErrInvalidJob.Wrap("queue_name cannot be empty")
	}
	if msg.Resources.Nodes < 1 {
		return ErrInvalidJob.Wrap("nodes must be at least 1")
	}
	return nil
}

// GetSigners returns the signers of the message
func (msg MsgSubmitJob) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.CustomerAddress)
	return []sdk.AccAddress{addr}
}

// MsgCancelJob cancels an HPC job
type MsgCancelJob struct {
	RequesterAddress string `json:"requester_address"`
	JobID            string `json:"job_id"`
	Reason           string `json:"reason,omitempty"`
}

// ValidateBasic performs basic validation
func (msg MsgCancelJob) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.RequesterAddress); err != nil {
		return ErrInvalidJob.Wrap("invalid requester address")
	}
	if msg.JobID == "" {
		return ErrInvalidJob.Wrap("job_id cannot be empty")
	}
	return nil
}

// GetSigners returns the signers of the message
func (msg MsgCancelJob) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.RequesterAddress)
	return []sdk.AccAddress{addr}
}

// MsgReportJobStatus reports job status from the provider daemon
type MsgReportJobStatus struct {
	ProviderAddress  string          `json:"provider_address"`
	JobID            string          `json:"job_id"`
	SLURMJobID       string          `json:"slurm_job_id,omitempty"`
	State            JobState        `json:"state"`
	StatusMessage    string          `json:"status_message,omitempty"`
	ExitCode         int32           `json:"exit_code,omitempty"`
	UsageMetrics     HPCUsageMetrics `json:"usage_metrics,omitempty"`
	Signature        string          `json:"signature"`
	SignedTimestamp  int64           `json:"signed_timestamp"`
}

// ValidateBasic performs basic validation
func (msg MsgReportJobStatus) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ProviderAddress); err != nil {
		return ErrInvalidJob.Wrap("invalid provider address")
	}
	if msg.JobID == "" {
		return ErrInvalidJob.Wrap("job_id cannot be empty")
	}
	if !IsValidJobState(msg.State) {
		return ErrInvalidJob.Wrapf("invalid job state: %s", msg.State)
	}
	if msg.Signature == "" {
		return ErrInvalidJob.Wrap("signature cannot be empty")
	}
	return nil
}

// GetSigners returns the signers of the message
func (msg MsgReportJobStatus) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.ProviderAddress)
	return []sdk.AccAddress{addr}
}

// MsgUpdateNodeMetadata updates node metadata
type MsgUpdateNodeMetadata struct {
	ProviderAddress     string               `json:"provider_address"`
	NodeID              string               `json:"node_id"`
	ClusterID           string               `json:"cluster_id"`
	Region              string               `json:"region,omitempty"`
	Datacenter          string               `json:"datacenter,omitempty"`
	LatencyMeasurements []LatencyMeasurement `json:"latency_measurements,omitempty"`
	NetworkBandwidthMbps int64               `json:"network_bandwidth_mbps,omitempty"`
	Resources           NodeResources        `json:"resources,omitempty"`
	Active              *bool                `json:"active,omitempty"`
}

// ValidateBasic performs basic validation
func (msg MsgUpdateNodeMetadata) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ProviderAddress); err != nil {
		return ErrInvalidNodeMetadata.Wrap("invalid provider address")
	}
	if msg.NodeID == "" {
		return ErrInvalidNodeMetadata.Wrap("node_id cannot be empty")
	}
	if msg.ClusterID == "" {
		return ErrInvalidNodeMetadata.Wrap("cluster_id cannot be empty")
	}
	return nil
}

// GetSigners returns the signers of the message
func (msg MsgUpdateNodeMetadata) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.ProviderAddress)
	return []sdk.AccAddress{addr}
}

// MsgFlagDispute flags a dispute for moderation
type MsgFlagDispute struct {
	DisputerAddress string `json:"disputer_address"`
	JobID           string `json:"job_id"`
	RewardID        string `json:"reward_id,omitempty"`
	DisputeType     string `json:"dispute_type"`
	Reason          string `json:"reason"`
	Evidence        string `json:"evidence,omitempty"`
}

// ValidateBasic performs basic validation
func (msg MsgFlagDispute) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.DisputerAddress); err != nil {
		return ErrInvalidDispute.Wrap("invalid disputer address")
	}
	if msg.JobID == "" {
		return ErrInvalidDispute.Wrap("job_id cannot be empty")
	}
	if msg.DisputeType == "" {
		return ErrInvalidDispute.Wrap("dispute_type cannot be empty")
	}
	if msg.Reason == "" {
		return ErrInvalidDispute.Wrap("reason cannot be empty")
	}
	return nil
}

// GetSigners returns the signers of the message
func (msg MsgFlagDispute) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.DisputerAddress)
	return []sdk.AccAddress{addr}
}

// MsgResolveDispute resolves a dispute (moderator only)
type MsgResolveDispute struct {
	ResolverAddress string        `json:"resolver_address"`
	DisputeID       string        `json:"dispute_id"`
	Status          DisputeStatus `json:"status"`
	Resolution      string        `json:"resolution"`
}

// ValidateBasic performs basic validation
func (msg MsgResolveDispute) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ResolverAddress); err != nil {
		return ErrInvalidDispute.Wrap("invalid resolver address")
	}
	if msg.DisputeID == "" {
		return ErrInvalidDispute.Wrap("dispute_id cannot be empty")
	}
	if msg.Status != DisputeStatusResolved && msg.Status != DisputeStatusRejected {
		return ErrInvalidDispute.Wrap("status must be resolved or rejected")
	}
	if msg.Resolution == "" {
		return ErrInvalidDispute.Wrap("resolution cannot be empty")
	}
	return nil
}

// GetSigners returns the signers of the message
func (msg MsgResolveDispute) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.ResolverAddress)
	return []sdk.AccAddress{addr}
}
