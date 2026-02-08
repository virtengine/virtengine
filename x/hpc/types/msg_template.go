// Package types contains types for the HPC module.
//
// VE-5F: Message types for workload template operations
package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Message type constants
const (
	TypeMsgCreateWorkloadTemplate    = "create_workload_template"
	TypeMsgUpdateWorkloadTemplate    = "update_workload_template"
	TypeMsgApproveWorkloadTemplate   = "approve_workload_template"
	TypeMsgRejectWorkloadTemplate    = "reject_workload_template"
	TypeMsgDeprecateWorkloadTemplate = "deprecate_workload_template"
	TypeMsgRevokeWorkloadTemplate    = "revoke_workload_template"
	TypeMsgSubmitJobFromTemplate     = "submit_job_from_template"
)

// NOTE: In production, these messages would be generated from protobuf.
// For now, implementing minimal interface methods for testing.

// MsgCreateWorkloadTemplate creates a new workload template
type MsgCreateWorkloadTemplate struct {
	Creator  string            `json:"creator"`
	Template *WorkloadTemplate `json:"template"`
}

// ProtoMessage implements proto.Message
func (*MsgCreateWorkloadTemplate) ProtoMessage() {}

// Reset implements proto.Message
func (m *MsgCreateWorkloadTemplate) Reset() { *m = MsgCreateWorkloadTemplate{} }

// String implements proto.Message
func (m *MsgCreateWorkloadTemplate) String() string {
	return fmt.Sprintf("MsgCreateWorkloadTemplate{Creator: %s, TemplateID: %s}", m.Creator, m.Template.TemplateID)
}

// NewMsgCreateWorkloadTemplate creates a new create template message
func NewMsgCreateWorkloadTemplate(creator string, template *WorkloadTemplate) *MsgCreateWorkloadTemplate {
	return &MsgCreateWorkloadTemplate{
		Creator:  creator,
		Template: template,
	}
}

// Route returns the route
func (msg *MsgCreateWorkloadTemplate) Route() string { return ModuleName }

// Type returns the message type
func (msg *MsgCreateWorkloadTemplate) Type() string { return TypeMsgCreateWorkloadTemplate }

// GetSigners returns the signers
func (msg *MsgCreateWorkloadTemplate) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes returns the sign bytes
func (msg *MsgCreateWorkloadTemplate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validates the message
func (msg *MsgCreateWorkloadTemplate) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return fmt.Errorf("invalid creator address: %w", err)
	}

	if msg.Template == nil {
		return fmt.Errorf("template is required")
	}

	if err := msg.Template.Validate(); err != nil {
		return ErrInvalidWorkloadTemplate.Wrapf("template validation failed: %s", err)
	}

	return nil
}

// MsgUpdateWorkloadTemplate updates an existing workload template
type MsgUpdateWorkloadTemplate struct {
	Creator  string            `json:"creator"`
	Template *WorkloadTemplate `json:"template"`
}

// ProtoMessage implements proto.Message
func (*MsgUpdateWorkloadTemplate) ProtoMessage() {}

// Reset implements proto.Message
func (m *MsgUpdateWorkloadTemplate) Reset() { *m = MsgUpdateWorkloadTemplate{} }

// String implements proto.Message
func (m *MsgUpdateWorkloadTemplate) String() string {
	return fmt.Sprintf("MsgUpdateWorkloadTemplate{Creator: %s, TemplateID: %s}", m.Creator, m.Template.TemplateID)
}

// NewMsgUpdateWorkloadTemplate creates a new update template message
func NewMsgUpdateWorkloadTemplate(creator string, template *WorkloadTemplate) *MsgUpdateWorkloadTemplate {
	return &MsgUpdateWorkloadTemplate{
		Creator:  creator,
		Template: template,
	}
}

// Route returns the route
func (msg *MsgUpdateWorkloadTemplate) Route() string { return ModuleName }

// Type returns the message type
func (msg *MsgUpdateWorkloadTemplate) Type() string { return TypeMsgUpdateWorkloadTemplate }

// GetSigners returns the signers
func (msg *MsgUpdateWorkloadTemplate) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes returns the sign bytes
func (msg *MsgUpdateWorkloadTemplate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validates the message
func (msg *MsgUpdateWorkloadTemplate) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return fmt.Errorf("invalid creator address: %w", err)
	}

	if msg.Template == nil {
		return fmt.Errorf("template is required")
	}

	if err := msg.Template.Validate(); err != nil {
		return ErrInvalidWorkloadTemplate.Wrapf("template validation failed: %s", err)
	}

	return nil
}

// MsgApproveWorkloadTemplate approves a workload template
type MsgApproveWorkloadTemplate struct {
	Authority  string `json:"authority"`
	TemplateID string `json:"template_id"`
	Version    string `json:"version"`
}

// ProtoMessage implements proto.Message
func (*MsgApproveWorkloadTemplate) ProtoMessage() {}

// Reset implements proto.Message
func (m *MsgApproveWorkloadTemplate) Reset() { *m = MsgApproveWorkloadTemplate{} }

// String implements proto.Message
func (m *MsgApproveWorkloadTemplate) String() string {
	return fmt.Sprintf("MsgApproveWorkloadTemplate{Authority: %s, TemplateID: %s, Version: %s}", m.Authority, m.TemplateID, m.Version)
}

// NewMsgApproveWorkloadTemplate creates a new approve template message
func NewMsgApproveWorkloadTemplate(authority, templateID, version string) *MsgApproveWorkloadTemplate {
	return &MsgApproveWorkloadTemplate{
		Authority:  authority,
		TemplateID: templateID,
		Version:    version,
	}
}

// Route returns the route
func (msg *MsgApproveWorkloadTemplate) Route() string { return ModuleName }

// Type returns the message type
func (msg *MsgApproveWorkloadTemplate) Type() string { return TypeMsgApproveWorkloadTemplate }

// GetSigners returns the signers
func (msg *MsgApproveWorkloadTemplate) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{authority}
}

// GetSignBytes returns the sign bytes
func (msg *MsgApproveWorkloadTemplate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validates the message
func (msg *MsgApproveWorkloadTemplate) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return fmt.Errorf("invalid authority address: %w", err)
	}

	if msg.TemplateID == "" {
		return fmt.Errorf("template ID is required")
	}

	if msg.Version == "" {
		return fmt.Errorf("version is required")
	}

	return nil
}

// MsgRejectWorkloadTemplate rejects a workload template
type MsgRejectWorkloadTemplate struct {
	Authority  string `json:"authority"`
	TemplateID string `json:"template_id"`
	Version    string `json:"version"`
	Reason     string `json:"reason"`
}

// ProtoMessage implements proto.Message
func (*MsgRejectWorkloadTemplate) ProtoMessage() {}

// Reset implements proto.Message
func (m *MsgRejectWorkloadTemplate) Reset() { *m = MsgRejectWorkloadTemplate{} }

// String implements proto.Message
func (m *MsgRejectWorkloadTemplate) String() string {
	return fmt.Sprintf("MsgRejectWorkloadTemplate{Authority: %s, TemplateID: %s}", m.Authority, m.TemplateID)
}

// NewMsgRejectWorkloadTemplate creates a new reject template message
func NewMsgRejectWorkloadTemplate(authority, templateID, version, reason string) *MsgRejectWorkloadTemplate {
	return &MsgRejectWorkloadTemplate{
		Authority:  authority,
		TemplateID: templateID,
		Version:    version,
		Reason:     reason,
	}
}

// Route returns the route
func (msg *MsgRejectWorkloadTemplate) Route() string { return ModuleName }

// Type returns the message type
func (msg *MsgRejectWorkloadTemplate) Type() string { return TypeMsgRejectWorkloadTemplate }

// GetSigners returns the signers
func (msg *MsgRejectWorkloadTemplate) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{authority}
}

// GetSignBytes returns the sign bytes
func (msg *MsgRejectWorkloadTemplate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validates the message
func (msg *MsgRejectWorkloadTemplate) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return fmt.Errorf("invalid authority address: %s", err)
	}

	if msg.TemplateID == "" {
		return fmt.Errorf("template ID is required")
	}

	if msg.Version == "" {
		return fmt.Errorf("version is required")
	}

	if msg.Reason == "" {
		return fmt.Errorf("reason is required")
	}

	return nil
}

// MsgDeprecateWorkloadTemplate deprecates a workload template
type MsgDeprecateWorkloadTemplate struct {
	Authority  string `json:"authority"`
	TemplateID string `json:"template_id"`
	Version    string `json:"version"`
	Reason     string `json:"reason"`
}

// ProtoMessage implements proto.Message
func (*MsgDeprecateWorkloadTemplate) ProtoMessage() {}

// Reset implements proto.Message
func (m *MsgDeprecateWorkloadTemplate) Reset() { *m = MsgDeprecateWorkloadTemplate{} }

// String implements proto.Message
func (m *MsgDeprecateWorkloadTemplate) String() string {
	return fmt.Sprintf("MsgDeprecateWorkloadTemplate{Authority: %s, TemplateID: %s}", m.Authority, m.TemplateID)
}

// NewMsgDeprecateWorkloadTemplate creates a new deprecate template message
func NewMsgDeprecateWorkloadTemplate(authority, templateID, version, reason string) *MsgDeprecateWorkloadTemplate {
	return &MsgDeprecateWorkloadTemplate{
		Authority:  authority,
		TemplateID: templateID,
		Version:    version,
		Reason:     reason,
	}
}

// Route returns the route
func (msg *MsgDeprecateWorkloadTemplate) Route() string { return ModuleName }

// Type returns the message type
func (msg *MsgDeprecateWorkloadTemplate) Type() string { return TypeMsgDeprecateWorkloadTemplate }

// GetSigners returns the signers
func (msg *MsgDeprecateWorkloadTemplate) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{authority}
}

// GetSignBytes returns the sign bytes
func (msg *MsgDeprecateWorkloadTemplate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validates the message
func (msg *MsgDeprecateWorkloadTemplate) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return fmt.Errorf("invalid authority address: %s", err)
	}

	if msg.TemplateID == "" {
		return fmt.Errorf("template ID is required")
	}

	if msg.Version == "" {
		return fmt.Errorf("version is required")
	}

	if msg.Reason == "" {
		return fmt.Errorf("reason is required")
	}

	return nil
}

// MsgRevokeWorkloadTemplate revokes a workload template
type MsgRevokeWorkloadTemplate struct {
	Authority  string `json:"authority"`
	TemplateID string `json:"template_id"`
	Version    string `json:"version"`
	Reason     string `json:"reason"`
}

// ProtoMessage implements proto.Message
func (*MsgRevokeWorkloadTemplate) ProtoMessage() {}

// Reset implements proto.Message
func (m *MsgRevokeWorkloadTemplate) Reset() { *m = MsgRevokeWorkloadTemplate{} }

// String implements proto.Message
func (m *MsgRevokeWorkloadTemplate) String() string {
	return fmt.Sprintf("MsgRevokeWorkloadTemplate{Authority: %s, TemplateID: %s}", m.Authority, m.TemplateID)
}

// NewMsgRevokeWorkloadTemplate creates a new revoke template message
func NewMsgRevokeWorkloadTemplate(authority, templateID, version, reason string) *MsgRevokeWorkloadTemplate {
	return &MsgRevokeWorkloadTemplate{
		Authority:  authority,
		TemplateID: templateID,
		Version:    version,
		Reason:     reason,
	}
}

// Route returns the route
func (msg *MsgRevokeWorkloadTemplate) Route() string { return ModuleName }

// Type returns the message type
func (msg *MsgRevokeWorkloadTemplate) Type() string { return TypeMsgRevokeWorkloadTemplate }

// GetSigners returns the signers
func (msg *MsgRevokeWorkloadTemplate) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{authority}
}

// GetSignBytes returns the sign bytes
func (msg *MsgRevokeWorkloadTemplate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validates the message
func (msg *MsgRevokeWorkloadTemplate) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return fmt.Errorf("invalid authority address: %s", err)
	}

	if msg.TemplateID == "" {
		return fmt.Errorf("template ID is required")
	}

	if msg.Version == "" {
		return fmt.Errorf("version is required")
	}

	if msg.Reason == "" {
		return fmt.Errorf("reason is required")
	}

	return nil
}

// ResourceOverrides allows users to override template defaults.
type ResourceOverrides struct {
	Nodes           int32 `json:"nodes,omitempty"`
	CpusPerNode     int32 `json:"cpus_per_node,omitempty"`
	MemoryMbPerNode int64 `json:"memory_mb_per_node,omitempty"`
	RuntimeMinutes  int64 `json:"runtime_minutes,omitempty"`
	GpusPerNode     int32 `json:"gpus_per_node,omitempty"`
}

// MsgSubmitJobFromTemplate submits a job from a template
type MsgSubmitJobFromTemplate struct {
	Creator    string            `json:"creator"`
	TemplateID string            `json:"template_id"`
	Version    string            `json:"version"`
	Parameters map[string]string `json:"parameters,omitempty"`

	// ResourceOverrides provides structured overrides (preferred).
	ResourceOverrides *ResourceOverrides `json:"resource_overrides,omitempty"`

	// Legacy override fields (kept for CLI/backwards compatibility).
	Nodes    int32 `json:"nodes,omitempty"`
	CPUs     int32 `json:"cpus,omitempty"`
	MemoryMB int64 `json:"memory_mb,omitempty"`
	Runtime  int64 `json:"runtime,omitempty"`
}

// ProtoMessage implements proto.Message
func (*MsgSubmitJobFromTemplate) ProtoMessage() {}

// Reset implements proto.Message
func (m *MsgSubmitJobFromTemplate) Reset() { *m = MsgSubmitJobFromTemplate{} }

// String implements proto.Message
func (m *MsgSubmitJobFromTemplate) String() string {
	return fmt.Sprintf("MsgSubmitJobFromTemplate{Creator: %s, TemplateID: %s}", m.Creator, m.TemplateID)
}

// NewMsgSubmitJobFromTemplate creates a new submit job from template message
func NewMsgSubmitJobFromTemplate(creator, templateID, version string, parameters map[string]string) *MsgSubmitJobFromTemplate {
	return &MsgSubmitJobFromTemplate{
		Creator:    creator,
		TemplateID: templateID,
		Version:    version,
		Parameters: parameters,
	}
}

// Route returns the route
func (msg *MsgSubmitJobFromTemplate) Route() string { return ModuleName }

// Type returns the message type
func (msg *MsgSubmitJobFromTemplate) Type() string { return TypeMsgSubmitJobFromTemplate }

// GetSigners returns the signers
func (msg *MsgSubmitJobFromTemplate) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes returns the sign bytes
func (msg *MsgSubmitJobFromTemplate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validates the message
func (msg *MsgSubmitJobFromTemplate) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return fmt.Errorf("invalid creator address: %s", err)
	}

	if msg.TemplateID == "" {
		return fmt.Errorf("template ID is required")
	}

	if msg.Version == "" {
		return fmt.Errorf("version is required")
	}

	if msg.Nodes < 0 {
		return fmt.Errorf("nodes cannot be negative")
	}

	if msg.CPUs < 0 {
		return fmt.Errorf("CPUs cannot be negative")
	}

	if msg.MemoryMB < 0 {
		return fmt.Errorf("memory cannot be negative")
	}

	if msg.Runtime < 0 {
		return fmt.Errorf("runtime cannot be negative")
	}

	if msg.ResourceOverrides != nil {
		if msg.ResourceOverrides.Nodes < 0 {
			return fmt.Errorf("resource_overrides.nodes cannot be negative")
		}
		if msg.ResourceOverrides.CpusPerNode < 0 {
			return fmt.Errorf("resource_overrides.cpus_per_node cannot be negative")
		}
		if msg.ResourceOverrides.MemoryMbPerNode < 0 {
			return fmt.Errorf("resource_overrides.memory_mb_per_node cannot be negative")
		}
		if msg.ResourceOverrides.RuntimeMinutes < 0 {
			return fmt.Errorf("resource_overrides.runtime_minutes cannot be negative")
		}
		if msg.ResourceOverrides.GpusPerNode < 0 {
			return fmt.Errorf("resource_overrides.gpus_per_node cannot be negative")
		}
	}

	return nil
}

// Response types
type MsgCreateWorkloadTemplateResponse struct {
	TemplateID string `json:"template_id"`
}

type MsgUpdateWorkloadTemplateResponse struct{}

type MsgApproveWorkloadTemplateResponse struct{}

type MsgRejectWorkloadTemplateResponse struct{}

type MsgDeprecateWorkloadTemplateResponse struct{}

type MsgRevokeWorkloadTemplateResponse struct{}

type MsgSubmitJobFromTemplateResponse struct {
	JobId string `json:"job_id"`
}
