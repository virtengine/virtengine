// Package types provides VEID module types.
//
// This file defines type aliases for model versioning messages, using the
// proto-generated types from sdk/go/node/veid/v1 as the source of truth.
//
// Task Reference: VE-3007 - Model Versioning and Governance
package types

import (
	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
)

// ============================================================================
// Message Type Constants for Model Versioning
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
// Model Message Type Aliases - from proto-generated types
// ============================================================================

// MsgRegisterModel registers a new model.
type MsgRegisterModel = veidv1.MsgRegisterModel

// MsgRegisterModelResponse is the response for MsgRegisterModel.
type MsgRegisterModelResponse = veidv1.MsgRegisterModelResponse

// MsgProposeModelUpdate proposes updating active model via governance.
type MsgProposeModelUpdate = veidv1.MsgProposeModelUpdate

// MsgProposeModelUpdateResponse is the response for MsgProposeModelUpdate.
type MsgProposeModelUpdateResponse = veidv1.MsgProposeModelUpdateResponse

// MsgReportModelVersion reports validator's model version.
type MsgReportModelVersion = veidv1.MsgReportModelVersion

// MsgReportModelVersionResponse is the response for MsgReportModelVersion.
type MsgReportModelVersionResponse = veidv1.MsgReportModelVersionResponse

// MsgActivateModel activates a pending model after governance approval.
type MsgActivateModel = veidv1.MsgActivateModel

// MsgActivateModelResponse is the response for MsgActivateModel.
type MsgActivateModelResponse = veidv1.MsgActivateModelResponse

// MsgDeprecateModel deprecates a model.
type MsgDeprecateModel = veidv1.MsgDeprecateModel

// MsgDeprecateModelResponse is the response for MsgDeprecateModel.
type MsgDeprecateModelResponse = veidv1.MsgDeprecateModelResponse

// MsgRevokeModel revokes a model.
type MsgRevokeModel = veidv1.MsgRevokeModel

// MsgRevokeModelResponse is the response for MsgRevokeModel.
type MsgRevokeModelResponse = veidv1.MsgRevokeModelResponse

// ============================================================================
// Proto-Generated Type Aliases for Serialization
// These are used for MsgRegisterModel.ModelInfo field compatibility
// ============================================================================

// ProtoMLModelInfo is the proto-generated MLModelInfo type for serialization.
type ProtoMLModelInfo = veidv1.MLModelInfo

// ProtoModelUpdateProposal is the proto-generated ModelUpdateProposal for serialization.
type ProtoModelUpdateProposal = veidv1.ModelUpdateProposal

// ============================================================================
// Model Constructor Functions
// ============================================================================

// NewMsgRegisterModel creates a new MsgRegisterModel.
func NewMsgRegisterModel(authority string, info *ProtoMLModelInfo) *MsgRegisterModel {
	return &MsgRegisterModel{
		Authority: authority,
		ModelInfo: *info,
	}
}

// NewMsgProposeModelUpdate creates a new MsgProposeModelUpdate.
func NewMsgProposeModelUpdate(proposer string, proposal *ProtoModelUpdateProposal) *MsgProposeModelUpdate {
	return &MsgProposeModelUpdate{
		Proposer: proposer,
		Proposal: *proposal,
	}
}

// NewMsgReportModelVersion creates a new MsgReportModelVersion.
func NewMsgReportModelVersion(validatorAddr string, versions map[string]string) *MsgReportModelVersion {
	return &MsgReportModelVersion{
		ValidatorAddress: validatorAddr,
		ModelVersions:    versions,
	}
}

// NewMsgActivateModel creates a new MsgActivateModel.
func NewMsgActivateModel(authority, modelType, modelID string, govID uint64) *MsgActivateModel {
	return &MsgActivateModel{
		Authority:    authority,
		ModelType:    modelType,
		ModelId:      modelID,
		GovernanceId: govID,
	}
}

// NewMsgDeprecateModel creates a new MsgDeprecateModel.
func NewMsgDeprecateModel(authority, modelID, reason string) *MsgDeprecateModel {
	return &MsgDeprecateModel{
		Authority: authority,
		ModelId:   modelID,
		Reason:    reason,
	}
}

// NewMsgRevokeModel creates a new MsgRevokeModel.
func NewMsgRevokeModel(authority, modelID, reason string) *MsgRevokeModel {
	return &MsgRevokeModel{
		Authority: authority,
		ModelId:   modelID,
		Reason:    reason,
	}
}
