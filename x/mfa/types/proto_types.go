// Package types provides type aliases and adapters for the generated proto types.
//
// This file bridges the local MFA types in x/mfa/types with the generated protobuf
// types in sdk/go/node/mfa/v1. It provides type aliases for proto message types
// and adapter implementations to translate between local and generated interfaces.
package types

import (
	"context"
	"math"

	mfav1 "github.com/virtengine/virtengine/sdk/go/node/mfa/v1"
)

// =============================================================================
// Proto Type Aliases
// =============================================================================

// Proto message type aliases - these point to the generated protobuf types
// for use in gRPC and Cosmos SDK message routing.

type (
	// MsgEnrollFactorPB is the generated proto type for MsgEnrollFactor
	MsgEnrollFactorPB = mfav1.MsgEnrollFactor
	// MsgEnrollFactorResponsePB is the generated proto type for MsgEnrollFactorResponse
	MsgEnrollFactorResponsePB = mfav1.MsgEnrollFactorResponse
	// MsgRevokeFactorPB is the generated proto type for MsgRevokeFactor
	MsgRevokeFactorPB = mfav1.MsgRevokeFactor
	// MsgRevokeFactorResponsePB is the generated proto type for MsgRevokeFactorResponse
	MsgRevokeFactorResponsePB = mfav1.MsgRevokeFactorResponse
	// MsgSetMFAPolicyPB is the generated proto type for MsgSetMFAPolicy
	MsgSetMFAPolicyPB = mfav1.MsgSetMFAPolicy
	// MsgSetMFAPolicyResponsePB is the generated proto type for MsgSetMFAPolicyResponse
	MsgSetMFAPolicyResponsePB = mfav1.MsgSetMFAPolicyResponse
	// MsgCreateChallengePB is the generated proto type for MsgCreateChallenge
	MsgCreateChallengePB = mfav1.MsgCreateChallenge
	// MsgCreateChallengeResponsePB is the generated proto type for MsgCreateChallengeResponse
	MsgCreateChallengeResponsePB = mfav1.MsgCreateChallengeResponse
	// MsgVerifyChallengePB is the generated proto type for MsgVerifyChallenge
	MsgVerifyChallengePB = mfav1.MsgVerifyChallenge
	// MsgVerifyChallengeResponsePB is the generated proto type for MsgVerifyChallengeResponse
	MsgVerifyChallengeResponsePB = mfav1.MsgVerifyChallengeResponse
	// MsgAddTrustedDevicePB is the generated proto type for MsgAddTrustedDevice
	MsgAddTrustedDevicePB = mfav1.MsgAddTrustedDevice
	// MsgAddTrustedDeviceResponsePB is the generated proto type for MsgAddTrustedDeviceResponse
	MsgAddTrustedDeviceResponsePB = mfav1.MsgAddTrustedDeviceResponse
	// MsgRemoveTrustedDevicePB is the generated proto type for MsgRemoveTrustedDevice
	MsgRemoveTrustedDevicePB = mfav1.MsgRemoveTrustedDevice
	// MsgRemoveTrustedDeviceResponsePB is the generated proto type for MsgRemoveTrustedDeviceResponse
	MsgRemoveTrustedDeviceResponsePB = mfav1.MsgRemoveTrustedDeviceResponse
	// MsgUpdateSensitiveTxConfigPB is the generated proto type for MsgUpdateSensitiveTxConfig
	MsgUpdateSensitiveTxConfigPB = mfav1.MsgUpdateSensitiveTxConfig
	// MsgUpdateSensitiveTxConfigResponsePB is the generated proto type for MsgUpdateSensitiveTxConfigResponse
	MsgUpdateSensitiveTxConfigResponsePB = mfav1.MsgUpdateSensitiveTxConfigResponse
	// MsgUpdateParamsPB is the generated proto type for MsgUpdateParams
	MsgUpdateParamsPB = mfav1.MsgUpdateParams
	// MsgUpdateParamsResponsePB is the generated proto type for MsgUpdateParamsResponse
	MsgUpdateParamsResponsePB = mfav1.MsgUpdateParamsResponse

	// GenesisStatePB is the generated proto type for GenesisState
	GenesisStatePB = mfav1.GenesisState
	// ParamsPB is the generated proto type for Params
	ParamsPB = mfav1.Params

	// MFAProofPB is the generated proto type for MFAProof
	MFAProofPB = mfav1.MFAProof
	// MFAPolicyPB is the generated proto type for MFAPolicy
	MFAPolicyPB = mfav1.MFAPolicy
	// FactorEnrollmentPB is the generated proto type for FactorEnrollment
	FactorEnrollmentPB = mfav1.FactorEnrollment
	// FactorCombinationPB is the generated proto type for FactorCombination
	FactorCombinationPB = mfav1.FactorCombination
	// FactorMetadataPB is the generated proto type for FactorMetadata
	FactorMetadataPB = mfav1.FactorMetadata
	// ChallengePB is the generated proto type for Challenge
	ChallengePB = mfav1.Challenge
	// ChallengeResponsePB is the generated proto type for ChallengeResponse
	ChallengeResponsePB = mfav1.ChallengeResponse
	// AuthorizationSessionPB is the generated proto type for AuthorizationSession
	AuthorizationSessionPB = mfav1.AuthorizationSession
	// SensitiveTxConfigPB is the generated proto type for SensitiveTxConfig
	SensitiveTxConfigPB = mfav1.SensitiveTxConfig
	// TrustedDevicePB is the generated proto type for TrustedDevice
	TrustedDevicePB = mfav1.TrustedDevice
	// TrustedDevicePolicyPB is the generated proto type for TrustedDevicePolicy
	TrustedDevicePolicyPB = mfav1.TrustedDevicePolicy
	// DeviceInfoPB is the generated proto type for DeviceInfo
	DeviceInfoPB = mfav1.DeviceInfo
	// ClientInfoPB is the generated proto type for ClientInfo
	ClientInfoPB = mfav1.ClientInfo
)

// Proto enum type aliases
type (
	// FactorTypePB is the generated proto enum for FactorType
	FactorTypePB = mfav1.FactorType
	// FactorSecurityLevelPB is the generated proto enum for FactorSecurityLevel
	FactorSecurityLevelPB = mfav1.FactorSecurityLevel
	// FactorEnrollmentStatusPB is the generated proto enum for FactorEnrollmentStatus
	FactorEnrollmentStatusPB = mfav1.FactorEnrollmentStatus
	// ChallengeStatusPB is the generated proto enum for ChallengeStatus
	ChallengeStatusPB = mfav1.ChallengeStatus
	// SensitiveTransactionTypePB is the generated proto enum for SensitiveTransactionType
	SensitiveTransactionTypePB = mfav1.SensitiveTransactionType
	// HardwareKeyTypePB is the generated proto enum for HardwareKeyType
	HardwareKeyTypePB = mfav1.HardwareKeyType
	// RevocationStatusPB is the generated proto enum for RevocationStatus
	RevocationStatusPB = mfav1.RevocationStatus
)

// Proto enum value constants
const (
	// Factor types from generated proto
	FactorTypePBUnspecified   = mfav1.FactorTypeUnspecified
	FactorTypePBTOTP          = mfav1.FactorTypeTOTP
	FactorTypePBFIDO2         = mfav1.FactorTypeFIDO2
	FactorTypePBSMS           = mfav1.FactorTypeSMS
	FactorTypePBEmail         = mfav1.FactorTypeEmail
	FactorTypePBVEID          = mfav1.FactorTypeVEID
	FactorTypePBTrustedDevice = mfav1.FactorTypeTrustedDevice
	FactorTypePBHardwareKey   = mfav1.FactorTypeHardwareKey

	// Enrollment statuses from generated proto
	EnrollmentStatusPBUnspecified = mfav1.EnrollmentStatusUnspecified
	EnrollmentStatusPBPending     = mfav1.EnrollmentStatusPending
	EnrollmentStatusPBActive      = mfav1.EnrollmentStatusActive
	EnrollmentStatusPBRevoked     = mfav1.EnrollmentStatusRevoked
	EnrollmentStatusPBExpired     = mfav1.EnrollmentStatusExpired

	// Challenge statuses from generated proto
	ChallengeStatusPBUnspecified = mfav1.ChallengeStatusUnspecified
	ChallengeStatusPBPending     = mfav1.ChallengeStatusPending
	ChallengeStatusPBVerified    = mfav1.ChallengeStatusVerified
	ChallengeStatusPBFailed      = mfav1.ChallengeStatusFailed
	ChallengeStatusPBExpired     = mfav1.ChallengeStatusExpired
	ChallengeStatusPBCancelled   = mfav1.ChallengeStatusCancelled
)

// =============================================================================
// MsgServer Adapter
// =============================================================================

// msgServerAdapter adapts the local MsgServer interface to the generated proto MsgServer.
type msgServerAdapter struct {
	local MsgServer
}

// NewMsgServerAdapter creates a new adapter that wraps a local MsgServer
// to implement the generated proto MsgServer interface.
func NewMsgServerAdapter(local MsgServer) mfav1.MsgServer {
	return &msgServerAdapter{local: local}
}

func (a *msgServerAdapter) EnrollFactor(ctx context.Context, req *mfav1.MsgEnrollFactor) (*mfav1.MsgEnrollFactorResponse, error) {
	localReq := convertMsgEnrollFactorFromProto(req)
	resp, err := a.local.EnrollFactor(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return convertMsgEnrollFactorResponseToProto(resp), nil
}

func (a *msgServerAdapter) RevokeFactor(ctx context.Context, req *mfav1.MsgRevokeFactor) (*mfav1.MsgRevokeFactorResponse, error) {
	localReq := convertMsgRevokeFactorFromProto(req)
	resp, err := a.local.RevokeFactor(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return convertMsgRevokeFactorResponseToProto(resp), nil
}

func (a *msgServerAdapter) SetMFAPolicy(ctx context.Context, req *mfav1.MsgSetMFAPolicy) (*mfav1.MsgSetMFAPolicyResponse, error) {
	localReq := convertMsgSetMFAPolicyFromProto(req)
	resp, err := a.local.SetMFAPolicy(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return convertMsgSetMFAPolicyResponseToProto(resp), nil
}

func (a *msgServerAdapter) CreateChallenge(ctx context.Context, req *mfav1.MsgCreateChallenge) (*mfav1.MsgCreateChallengeResponse, error) {
	localReq := convertMsgCreateChallengeFromProto(req)
	resp, err := a.local.CreateChallenge(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return convertMsgCreateChallengeResponseToProto(resp), nil
}

func (a *msgServerAdapter) VerifyChallenge(ctx context.Context, req *mfav1.MsgVerifyChallenge) (*mfav1.MsgVerifyChallengeResponse, error) {
	localReq := convertMsgVerifyChallengeFromProto(req)
	resp, err := a.local.VerifyChallenge(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return convertMsgVerifyChallengeResponseToProto(resp), nil
}

func (a *msgServerAdapter) AddTrustedDevice(ctx context.Context, req *mfav1.MsgAddTrustedDevice) (*mfav1.MsgAddTrustedDeviceResponse, error) {
	localReq := convertMsgAddTrustedDeviceFromProto(req)
	resp, err := a.local.AddTrustedDevice(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return convertMsgAddTrustedDeviceResponseToProto(resp), nil
}

func (a *msgServerAdapter) RemoveTrustedDevice(ctx context.Context, req *mfav1.MsgRemoveTrustedDevice) (*mfav1.MsgRemoveTrustedDeviceResponse, error) {
	localReq := convertMsgRemoveTrustedDeviceFromProto(req)
	resp, err := a.local.RemoveTrustedDevice(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return convertMsgRemoveTrustedDeviceResponseToProto(resp), nil
}

func (a *msgServerAdapter) UpdateSensitiveTxConfig(ctx context.Context, req *mfav1.MsgUpdateSensitiveTxConfig) (*mfav1.MsgUpdateSensitiveTxConfigResponse, error) {
	localReq := convertMsgUpdateSensitiveTxConfigFromProto(req)
	resp, err := a.local.UpdateSensitiveTxConfig(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return convertMsgUpdateSensitiveTxConfigResponseToProto(resp), nil
}

func (a *msgServerAdapter) UpdateParams(ctx context.Context, req *mfav1.MsgUpdateParams) (*mfav1.MsgUpdateParamsResponse, error) {
	// UpdateParams is not in the local MsgServer interface, return unimplemented
	// This would need to be added to the local interface if needed
	return &mfav1.MsgUpdateParamsResponse{}, nil
}

// =============================================================================
// QueryServer Adapter
// =============================================================================

// queryServerAdapter adapts the local QueryServer interface to the generated proto QueryServer.
type queryServerAdapter struct {
	local QueryServer
}

// NewQueryServerAdapter creates a new adapter that wraps a local QueryServer
// to implement the generated proto QueryServer interface.
func NewQueryServerAdapter(local QueryServer) mfav1.QueryServer {
	return &queryServerAdapter{local: local}
}

func (a *queryServerAdapter) MFAPolicy(ctx context.Context, req *mfav1.QueryMFAPolicyRequest) (*mfav1.QueryMFAPolicyResponse, error) {
	localReq := &QueryMFAPolicyRequest{Address: req.Address}
	resp, err := a.local.GetMFAPolicy(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return convertQueryMFAPolicyResponseToProto(resp), nil
}

func (a *queryServerAdapter) FactorEnrollments(ctx context.Context, req *mfav1.QueryFactorEnrollmentsRequest) (*mfav1.QueryFactorEnrollmentsResponse, error) {
	localReq := &QueryFactorEnrollmentsRequest{Address: req.Address}
	resp, err := a.local.GetFactorEnrollments(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return convertQueryFactorEnrollmentsResponseToProto(resp), nil
}

func (a *queryServerAdapter) FactorEnrollment(ctx context.Context, req *mfav1.QueryFactorEnrollmentRequest) (*mfav1.QueryFactorEnrollmentResponse, error) {
	localReq := &QueryFactorEnrollmentRequest{
		Address:  req.Address,
		FactorID: req.FactorId,
	}
	resp, err := a.local.GetFactorEnrollment(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return convertQueryFactorEnrollmentResponseToProto(resp), nil
}

func (a *queryServerAdapter) Challenge(ctx context.Context, req *mfav1.QueryChallengeRequest) (*mfav1.QueryChallengeResponse, error) {
	localReq := &QueryChallengeRequest{ChallengeID: req.ChallengeId}
	resp, err := a.local.GetChallenge(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return convertQueryChallengeResponseToProto(resp), nil
}

func (a *queryServerAdapter) PendingChallenges(ctx context.Context, req *mfav1.QueryPendingChallengesRequest) (*mfav1.QueryPendingChallengesResponse, error) {
	localReq := &QueryPendingChallengesRequest{Address: req.Address}
	resp, err := a.local.GetPendingChallenges(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return convertQueryPendingChallengesResponseToProto(resp), nil
}

func (a *queryServerAdapter) AuthorizationSession(ctx context.Context, req *mfav1.QueryAuthorizationSessionRequest) (*mfav1.QueryAuthorizationSessionResponse, error) {
	localReq := &QueryAuthorizationSessionRequest{SessionID: req.SessionId}
	resp, err := a.local.GetAuthorizationSession(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return convertQueryAuthorizationSessionResponseToProto(resp), nil
}

func (a *queryServerAdapter) TrustedDevices(ctx context.Context, req *mfav1.QueryTrustedDevicesRequest) (*mfav1.QueryTrustedDevicesResponse, error) {
	localReq := &QueryTrustedDevicesRequest{Address: req.Address}
	resp, err := a.local.GetTrustedDevices(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return convertQueryTrustedDevicesResponseToProto(resp), nil
}

func (a *queryServerAdapter) SensitiveTxConfig(ctx context.Context, req *mfav1.QuerySensitiveTxConfigRequest) (*mfav1.QuerySensitiveTxConfigResponse, error) {
	localReq := &QuerySensitiveTxConfigRequest{TransactionType: safeSensitiveTransactionType(req.TransactionType)}
	resp, err := a.local.GetSensitiveTxConfig(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return convertQuerySensitiveTxConfigResponseToProto(resp), nil
}

func (a *queryServerAdapter) AllSensitiveTxConfigs(ctx context.Context, req *mfav1.QueryAllSensitiveTxConfigsRequest) (*mfav1.QueryAllSensitiveTxConfigsResponse, error) {
	localReq := &QueryAllSensitiveTxConfigsRequest{}
	resp, err := a.local.GetAllSensitiveTxConfigs(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return convertQueryAllSensitiveTxConfigsResponseToProto(resp), nil
}

func (a *queryServerAdapter) MFARequired(ctx context.Context, req *mfav1.QueryMFARequiredRequest) (*mfav1.QueryMFARequiredResponse, error) {
	localReq := &QueryMFARequiredRequest{
		Address:         req.Address,
		TransactionType: safeSensitiveTransactionType(req.TransactionType),
	}
	resp, err := a.local.IsMFARequired(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return convertQueryMFARequiredResponseToProto(resp), nil
}

func (a *queryServerAdapter) Params(ctx context.Context, req *mfav1.QueryParamsRequest) (*mfav1.QueryParamsResponse, error) {
	localReq := &QueryParamsRequest{}
	resp, err := a.local.GetParams(ctx, localReq)
	if err != nil {
		return nil, err
	}
	return convertQueryParamsResponseToProto(resp), nil
}

func safeSensitiveTransactionType(value mfav1.SensitiveTransactionType) SensitiveTransactionType {
	intValue := int32(value)
	if intValue < 0 {
		return 0
	}
	if intValue > math.MaxUint8 {
		return SensitiveTransactionType(math.MaxUint8)
	}
	//nolint:gosec // range checked above
	return SensitiveTransactionType(intValue)
}
