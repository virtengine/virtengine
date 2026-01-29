// Package types provides VEID module types.
//
// This file defines model versioning messages for the VEID module.
//
// Task Reference: VE-3007 - Model Versioning and Governance
package types

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ============================================================================
// Message Types for Model Versioning
// ============================================================================

const (
	// TypeMsgRegisterModel is the type for MsgRegisterModel
	TypeMsgRegisterModel = "register_model"

	// TypeMsgProposeModelUpdate is the type for MsgProposeModelUpdate
	TypeMsgProposeModelUpdate = "propose_model_update"

	// TypeMsgReportModelVersion is the type for MsgReportModelVersion
	TypeMsgReportModelVersion = "report_model_version"

	// TypeMsgActivateModel is the type for MsgActivateModel
	TypeMsgActivateModel = "activate_model"

	// TypeMsgDeprecateModel is the type for MsgDeprecateModel
	TypeMsgDeprecateModel = "deprecate_model"

	// TypeMsgRevokeModel is the type for MsgRevokeModel
	TypeMsgRevokeModel = "revoke_model"
)

// ============================================================================
// MsgRegisterModel - Register a new ML model
// ============================================================================

// MsgRegisterModel registers a new model (authorized only)
type MsgRegisterModel struct {
	// Authority is the address authorized to register models
	Authority string `json:"authority"`

	// ModelInfo contains the model information
	ModelInfo MLModelInfo `json:"model_info"`
}

// NewMsgRegisterModel creates a new MsgRegisterModel
func NewMsgRegisterModel(authority string, info MLModelInfo) *MsgRegisterModel {
	return &MsgRegisterModel{
		Authority: authority,
		ModelInfo: info,
	}
}

// Route implements sdk.Msg
func (m MsgRegisterModel) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgRegisterModel) Type() string {
	return TypeMsgRegisterModel
}

// GetSigners implements sdk.Msg
func (m MsgRegisterModel) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m MsgRegisterModel) GetSignBytes() []byte {
	bz, _ := json.Marshal(m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgRegisterModel) ValidateBasic() error {
	if m.Authority == "" {
		return ErrInvalidAddress.Wrap("authority cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}
	if err := m.ModelInfo.Validate(); err != nil {
		return ErrInvalidModelInfo.Wrap(err.Error())
	}
	return nil
}

// ============================================================================
// MsgProposeModelUpdate - Propose updating active model via governance
// ============================================================================

// MsgProposeModelUpdate proposes updating active model via governance
type MsgProposeModelUpdate struct {
	// Proposer is the address proposing the update
	Proposer string `json:"proposer"`

	// Proposal contains the update proposal details
	Proposal ModelUpdateProposal `json:"proposal"`
}

// NewMsgProposeModelUpdate creates a new MsgProposeModelUpdate
func NewMsgProposeModelUpdate(proposer string, proposal ModelUpdateProposal) *MsgProposeModelUpdate {
	return &MsgProposeModelUpdate{
		Proposer: proposer,
		Proposal: proposal,
	}
}

// Route implements sdk.Msg
func (m MsgProposeModelUpdate) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgProposeModelUpdate) Type() string {
	return TypeMsgProposeModelUpdate
}

// GetSigners implements sdk.Msg
func (m MsgProposeModelUpdate) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Proposer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m MsgProposeModelUpdate) GetSignBytes() []byte {
	bz, _ := json.Marshal(m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgProposeModelUpdate) ValidateBasic() error {
	if m.Proposer == "" {
		return ErrInvalidAddress.Wrap("proposer cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Proposer); err != nil {
		return ErrInvalidAddress.Wrapf("invalid proposer address: %s", err)
	}
	if err := m.Proposal.Validate(); err != nil {
		return ErrInvalidModelProposal.Wrap(err.Error())
	}
	return nil
}

// ============================================================================
// MsgReportModelVersion - Report validator's model versions
// ============================================================================

// MsgReportModelVersion reports validator's model version
type MsgReportModelVersion struct {
	// ValidatorAddress is the validator's operator address
	ValidatorAddress string `json:"validator_address"`

	// ModelVersions maps model type to SHA256 hash
	ModelVersions map[string]string `json:"model_versions"`
}

// NewMsgReportModelVersion creates a new MsgReportModelVersion
func NewMsgReportModelVersion(validatorAddr string, versions map[string]string) *MsgReportModelVersion {
	return &MsgReportModelVersion{
		ValidatorAddress: validatorAddr,
		ModelVersions:    versions,
	}
}

// Route implements sdk.Msg
func (m MsgReportModelVersion) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgReportModelVersion) Type() string {
	return TypeMsgReportModelVersion
}

// GetSigners implements sdk.Msg
func (m MsgReportModelVersion) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.ValidatorAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m MsgReportModelVersion) GetSignBytes() []byte {
	bz, _ := json.Marshal(m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgReportModelVersion) ValidateBasic() error {
	if m.ValidatorAddress == "" {
		return ErrInvalidAddress.Wrap("validator_address cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.ValidatorAddress); err != nil {
		return ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}
	if m.ModelVersions == nil || len(m.ModelVersions) == 0 {
		return ErrInvalidModelInfo.Wrap("model_versions cannot be empty")
	}
	for modelType, hash := range m.ModelVersions {
		if !IsValidModelType(modelType) {
			return ErrInvalidModelType.Wrapf("invalid model type: %s", modelType)
		}
		if len(hash) != 64 {
			return ErrInvalidModelHash.Wrapf("invalid hash length for %s", modelType)
		}
	}
	return nil
}

// ============================================================================
// MsgActivateModel - Activate a pending model
// ============================================================================

// MsgActivateModel activates a pending model after governance approval
type MsgActivateModel struct {
	// Authority is the address authorized to activate models
	Authority string `json:"authority"`

	// ModelType is the type of model to activate
	ModelType string `json:"model_type"`

	// ModelID is the ID of the model to activate
	ModelID string `json:"model_id"`

	// GovernanceID is the governance proposal ID
	GovernanceID uint64 `json:"governance_id"`
}

// NewMsgActivateModel creates a new MsgActivateModel
func NewMsgActivateModel(authority, modelType, modelID string, govID uint64) *MsgActivateModel {
	return &MsgActivateModel{
		Authority:    authority,
		ModelType:    modelType,
		ModelID:      modelID,
		GovernanceID: govID,
	}
}

// Route implements sdk.Msg
func (m MsgActivateModel) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgActivateModel) Type() string {
	return TypeMsgActivateModel
}

// GetSigners implements sdk.Msg
func (m MsgActivateModel) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m MsgActivateModel) GetSignBytes() []byte {
	bz, _ := json.Marshal(m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgActivateModel) ValidateBasic() error {
	if m.Authority == "" {
		return ErrInvalidAddress.Wrap("authority cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}
	if !IsValidModelType(m.ModelType) {
		return ErrInvalidModelType.Wrapf("invalid model type: %s", m.ModelType)
	}
	if m.ModelID == "" {
		return ErrInvalidModelInfo.Wrap("model_id cannot be empty")
	}
	return nil
}

// ============================================================================
// MsgDeprecateModel - Deprecate a model
// ============================================================================

// MsgDeprecateModel deprecates a model
type MsgDeprecateModel struct {
	// Authority is the address authorized to deprecate models
	Authority string `json:"authority"`

	// ModelID is the ID of the model to deprecate
	ModelID string `json:"model_id"`

	// Reason is the reason for deprecation
	Reason string `json:"reason"`
}

// NewMsgDeprecateModel creates a new MsgDeprecateModel
func NewMsgDeprecateModel(authority, modelID, reason string) *MsgDeprecateModel {
	return &MsgDeprecateModel{
		Authority: authority,
		ModelID:   modelID,
		Reason:    reason,
	}
}

// Route implements sdk.Msg
func (m MsgDeprecateModel) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgDeprecateModel) Type() string {
	return TypeMsgDeprecateModel
}

// GetSigners implements sdk.Msg
func (m MsgDeprecateModel) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m MsgDeprecateModel) GetSignBytes() []byte {
	bz, _ := json.Marshal(m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgDeprecateModel) ValidateBasic() error {
	if m.Authority == "" {
		return ErrInvalidAddress.Wrap("authority cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}
	if m.ModelID == "" {
		return ErrInvalidModelInfo.Wrap("model_id cannot be empty")
	}
	return nil
}

// ============================================================================
// MsgRevokeModel - Revoke a model
// ============================================================================

// MsgRevokeModel revokes a model
type MsgRevokeModel struct {
	// Authority is the address authorized to revoke models
	Authority string `json:"authority"`

	// ModelID is the ID of the model to revoke
	ModelID string `json:"model_id"`

	// Reason is the reason for revocation
	Reason string `json:"reason"`
}

// NewMsgRevokeModel creates a new MsgRevokeModel
func NewMsgRevokeModel(authority, modelID, reason string) *MsgRevokeModel {
	return &MsgRevokeModel{
		Authority: authority,
		ModelID:   modelID,
		Reason:    reason,
	}
}

// Route implements sdk.Msg
func (m MsgRevokeModel) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgRevokeModel) Type() string {
	return TypeMsgRevokeModel
}

// GetSigners implements sdk.Msg
func (m MsgRevokeModel) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements sdk.Msg
func (m MsgRevokeModel) GetSignBytes() []byte {
	bz, _ := json.Marshal(m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (m MsgRevokeModel) ValidateBasic() error {
	if m.Authority == "" {
		return ErrInvalidAddress.Wrap("authority cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}
	if m.ModelID == "" {
		return ErrInvalidModelInfo.Wrap("model_id cannot be empty")
	}
	if m.Reason == "" {
		return ErrInvalidModelInfo.Wrap("reason cannot be empty for revocation")
	}
	return nil
}
