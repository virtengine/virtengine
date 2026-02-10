// Package v1 provides additional methods for generated HPC types.
package v1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// sdk.Msg interface methods for MsgRegisterCluster

func (msg *MsgRegisterCluster) ValidateBasic() error {
	if msg.ProviderAddress == "" {
		return ErrInvalidAddress.Wrap("provider address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.ProviderAddress); err != nil {
		return ErrInvalidAddress.Wrapf("invalid provider address: %v", err)
	}

	if msg.Name == "" {
		return ErrInvalidCluster.Wrap("name cannot be empty")
	}

	if msg.Region == "" {
		return ErrInvalidCluster.Wrap("region cannot be empty")
	}

	if msg.TotalNodes == 0 {
		return ErrInvalidCluster.Wrap("total_nodes must be greater than zero")
	}

	return nil
}

func (msg *MsgRegisterCluster) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.ProviderAddress)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgUpdateCluster

func (msg *MsgUpdateCluster) ValidateBasic() error {
	if msg.ProviderAddress == "" {
		return ErrInvalidAddress.Wrap("provider address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.ProviderAddress); err != nil {
		return ErrInvalidAddress.Wrapf("invalid provider address: %v", err)
	}

	if msg.ClusterId == "" {
		return ErrInvalidCluster.Wrap("cluster_id cannot be empty")
	}

	return nil
}

func (msg *MsgUpdateCluster) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.ProviderAddress)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgDeregisterCluster

func (msg *MsgDeregisterCluster) ValidateBasic() error {
	if msg.ProviderAddress == "" {
		return ErrInvalidAddress.Wrap("provider address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.ProviderAddress); err != nil {
		return ErrInvalidAddress.Wrapf("invalid provider address: %v", err)
	}

	if msg.ClusterId == "" {
		return ErrInvalidCluster.Wrap("cluster_id cannot be empty")
	}

	return nil
}

func (msg *MsgDeregisterCluster) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.ProviderAddress)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgCreateOffering

func (msg *MsgCreateOffering) ValidateBasic() error {
	if msg.ProviderAddress == "" {
		return ErrInvalidAddress.Wrap("provider address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.ProviderAddress); err != nil {
		return ErrInvalidAddress.Wrapf("invalid provider address: %v", err)
	}

	if msg.ClusterId == "" {
		return ErrInvalidOffering.Wrap("cluster_id cannot be empty")
	}

	if msg.Name == "" {
		return ErrInvalidOffering.Wrap("name cannot be empty")
	}

	return nil
}

func (msg *MsgCreateOffering) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.ProviderAddress)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgUpdateOffering

func (msg *MsgUpdateOffering) ValidateBasic() error {
	if msg.ProviderAddress == "" {
		return ErrInvalidAddress.Wrap("provider address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.ProviderAddress); err != nil {
		return ErrInvalidAddress.Wrapf("invalid provider address: %v", err)
	}

	if msg.OfferingId == "" {
		return ErrInvalidOffering.Wrap("offering_id cannot be empty")
	}

	return nil
}

func (msg *MsgUpdateOffering) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.ProviderAddress)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgSubmitJob

func (msg *MsgSubmitJob) ValidateBasic() error {
	if msg.CustomerAddress == "" {
		return ErrInvalidAddress.Wrap("customer address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.CustomerAddress); err != nil {
		return ErrInvalidAddress.Wrapf("invalid customer address: %v", err)
	}

	if msg.OfferingId == "" {
		return ErrInvalidJob.Wrap("offering_id cannot be empty")
	}

	if msg.Resources.Nodes == 0 {
		return ErrInvalidJob.Wrap("resources.nodes must be greater than zero")
	}

	return nil
}

func (msg *MsgSubmitJob) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.CustomerAddress)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgCancelJob

func (msg *MsgCancelJob) ValidateBasic() error {
	if msg.RequesterAddress == "" {
		return ErrInvalidAddress.Wrap("requester address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.RequesterAddress); err != nil {
		return ErrInvalidAddress.Wrapf("invalid requester address: %v", err)
	}

	if msg.JobId == "" {
		return ErrInvalidJob.Wrap("job_id cannot be empty")
	}

	return nil
}

func (msg *MsgCancelJob) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.RequesterAddress)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgReportJobStatus

func (msg *MsgReportJobStatus) ValidateBasic() error {
	if msg.ProviderAddress == "" {
		return ErrInvalidAddress.Wrap("provider address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.ProviderAddress); err != nil {
		return ErrInvalidAddress.Wrapf("invalid provider address: %v", err)
	}

	if msg.JobId == "" {
		return ErrInvalidJob.Wrap("job_id cannot be empty")
	}

	if msg.State == JobStateUnspecified {
		return ErrInvalidStatus.Wrap("state cannot be unspecified")
	}

	return nil
}

func (msg *MsgReportJobStatus) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.ProviderAddress)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgUpdateNodeMetadata

func (msg *MsgUpdateNodeMetadata) ValidateBasic() error {
	if msg.ProviderAddress == "" {
		return ErrInvalidAddress.Wrap("provider address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.ProviderAddress); err != nil {
		return ErrInvalidAddress.Wrapf("invalid provider address: %v", err)
	}

	if msg.NodeId == "" {
		return ErrInvalidNode.Wrap("node_id cannot be empty")
	}

	if msg.ClusterId == "" {
		return ErrInvalidNode.Wrap("cluster_id cannot be empty")
	}

	return nil
}

func (msg *MsgUpdateNodeMetadata) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.ProviderAddress)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgFlagDispute

func (msg *MsgFlagDispute) ValidateBasic() error {
	if msg.DisputerAddress == "" {
		return ErrInvalidAddress.Wrap("disputer address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.DisputerAddress); err != nil {
		return ErrInvalidAddress.Wrapf("invalid disputer address: %v", err)
	}

	if msg.JobId == "" {
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

func (msg *MsgFlagDispute) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.DisputerAddress)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgResolveDispute

func (msg *MsgResolveDispute) ValidateBasic() error {
	if msg.ResolverAddress == "" {
		return ErrInvalidAddress.Wrap("resolver address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.ResolverAddress); err != nil {
		return ErrInvalidAddress.Wrapf("invalid resolver address: %v", err)
	}

	if msg.DisputeId == "" {
		return ErrInvalidDispute.Wrap("dispute_id cannot be empty")
	}

	if msg.Resolution == "" {
		return ErrInvalidDispute.Wrap("resolution cannot be empty")
	}

	return nil
}

func (msg *MsgResolveDispute) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.ResolverAddress)
	return []sdk.AccAddress{addr}
}
