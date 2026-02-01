// Package v1 provides additional methods for generated HPC types.
package v1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// sdk.Msg interface methods for MsgRegisterCluster

func (msg *MsgRegisterCluster) ValidateBasic() error {
	if msg.Owner == "" {
		return ErrInvalidAddress.Wrap("owner address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return ErrInvalidAddress.Wrapf("invalid owner address: %v", err)
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
	addr, _ := sdk.AccAddressFromBech32(msg.Owner)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgUpdateCluster

func (msg *MsgUpdateCluster) ValidateBasic() error {
	if msg.Owner == "" {
		return ErrInvalidAddress.Wrap("owner address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return ErrInvalidAddress.Wrapf("invalid owner address: %v", err)
	}

	if msg.ClusterId == "" {
		return ErrInvalidCluster.Wrap("cluster_id cannot be empty")
	}

	return nil
}

func (msg *MsgUpdateCluster) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Owner)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgDeregisterCluster

func (msg *MsgDeregisterCluster) ValidateBasic() error {
	if msg.Owner == "" {
		return ErrInvalidAddress.Wrap("owner address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return ErrInvalidAddress.Wrapf("invalid owner address: %v", err)
	}

	if msg.ClusterId == "" {
		return ErrInvalidCluster.Wrap("cluster_id cannot be empty")
	}

	return nil
}

func (msg *MsgDeregisterCluster) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Owner)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgCreateOffering

func (msg *MsgCreateOffering) ValidateBasic() error {
	if msg.Provider == "" {
		return ErrInvalidAddress.Wrap("provider address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Provider); err != nil {
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
	addr, _ := sdk.AccAddressFromBech32(msg.Provider)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgUpdateOffering

func (msg *MsgUpdateOffering) ValidateBasic() error {
	if msg.Provider == "" {
		return ErrInvalidAddress.Wrap("provider address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Provider); err != nil {
		return ErrInvalidAddress.Wrapf("invalid provider address: %v", err)
	}

	if msg.OfferingId == "" {
		return ErrInvalidOffering.Wrap("offering_id cannot be empty")
	}

	return nil
}

func (msg *MsgUpdateOffering) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Provider)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgSubmitJob

func (msg *MsgSubmitJob) ValidateBasic() error {
	if msg.Submitter == "" {
		return ErrInvalidAddress.Wrap("submitter address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Submitter); err != nil {
		return ErrInvalidAddress.Wrapf("invalid submitter address: %v", err)
	}

	if msg.OfferingId == "" {
		return ErrInvalidJob.Wrap("offering_id cannot be empty")
	}

	if msg.RequestedNodes == 0 {
		return ErrInvalidJob.Wrap("requested_nodes must be greater than zero")
	}

	return nil
}

func (msg *MsgSubmitJob) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Submitter)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgCancelJob

func (msg *MsgCancelJob) ValidateBasic() error {
	if msg.Sender == "" {
		return ErrInvalidAddress.Wrap("sender address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrapf("invalid sender address: %v", err)
	}

	if msg.JobId == "" {
		return ErrInvalidJob.Wrap("job_id cannot be empty")
	}

	return nil
}

func (msg *MsgCancelJob) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgReportJobStatus

func (msg *MsgReportJobStatus) ValidateBasic() error {
	if msg.Reporter == "" {
		return ErrInvalidAddress.Wrap("reporter address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Reporter); err != nil {
		return ErrInvalidAddress.Wrapf("invalid reporter address: %v", err)
	}

	if msg.JobId == "" {
		return ErrInvalidJob.Wrap("job_id cannot be empty")
	}

	if msg.Status == "" {
		return ErrInvalidStatus.Wrap("status cannot be empty")
	}

	return nil
}

func (msg *MsgReportJobStatus) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Reporter)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgUpdateNodeMetadata

func (msg *MsgUpdateNodeMetadata) ValidateBasic() error {
	if msg.Owner == "" {
		return ErrInvalidAddress.Wrap("owner address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return ErrInvalidAddress.Wrapf("invalid owner address: %v", err)
	}

	if msg.ClusterId == "" {
		return ErrInvalidNode.Wrap("cluster_id cannot be empty")
	}

	if msg.NodeId == "" {
		return ErrInvalidNode.Wrap("node_id cannot be empty")
	}

	return nil
}

func (msg *MsgUpdateNodeMetadata) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Owner)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgFlagDispute

func (msg *MsgFlagDispute) ValidateBasic() error {
	if msg.Sender == "" {
		return ErrInvalidAddress.Wrap("sender address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrapf("invalid sender address: %v", err)
	}

	if msg.JobId == "" {
		return ErrInvalidDispute.Wrap("job_id cannot be empty")
	}

	if msg.Reason == "" {
		return ErrInvalidDispute.Wrap("reason cannot be empty")
	}

	return nil
}

func (msg *MsgFlagDispute) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgResolveDispute

func (msg *MsgResolveDispute) ValidateBasic() error {
	if msg.Authority == "" {
		return ErrInvalidAddress.Wrap("authority address is required")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidAddress.Wrapf("invalid authority address: %v", err)
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
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}

