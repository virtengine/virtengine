// Package types provides type aliases and extensions for the veid module.
//
// This file imports generated protobuf types from sdk/go/node/veid/v1 and creates
// type aliases for use throughout x/veid. This approach ensures:
// - Generated types are the source of truth for on-chain data structures
// - Additional methods (Validate, GetSigners, etc.) can be added via extension functions
// - Backward compatibility with existing keeper code
package types

import (
	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
)

// ============================================================================
// Proto Type Aliases - Core Types from types.pb.go
// ============================================================================

// ScopeTypePB is the protobuf-generated enum for scope types
// Use the local ScopeType (string-based) for backward compatibility,
// or use these conversion functions to work with the proto enum.
type ScopeTypePB = veidv1.ScopeType

// Proto enum constants for ScopeType
const (
	ScopeTypePBUnspecified       = veidv1.ScopeTypeUnspecified
	ScopeTypePBIDDocument        = veidv1.ScopeTypeIDDocument
	ScopeTypePBSelfie            = veidv1.ScopeTypeSelfie
	ScopeTypePBFaceVideo         = veidv1.ScopeTypeFaceVideo
	ScopeTypePBBiometric         = veidv1.ScopeTypeBiometric
	ScopeTypePBSSOMetadata       = veidv1.ScopeTypeSSOMetadata
	ScopeTypePBEmailProof        = veidv1.ScopeTypeEmailProof
	ScopeTypePBSMSProof          = veidv1.ScopeTypeSMSProof
	ScopeTypePBDomainVerify      = veidv1.ScopeTypeDomainVerify
	ScopeTypePBADSSO             = veidv1.ScopeTypeADSSO
	ScopeTypePBBiometricHardware = veidv1.ScopeTypeBiometricHardware
	ScopeTypePBDeviceAttestation = veidv1.ScopeTypeDeviceAttestation
)

// VerificationStatusPB is the protobuf-generated enum for verification status
type VerificationStatusPB = veidv1.VerificationStatus

// Proto enum constants for VerificationStatus
const (
	VerificationStatusPBUnknown                 = veidv1.VerificationStatusUnknown
	VerificationStatusPBPending                 = veidv1.VerificationStatusPending
	VerificationStatusPBInProgress              = veidv1.VerificationStatusInProgress
	VerificationStatusPBVerified                = veidv1.VerificationStatusVerified
	VerificationStatusPBRejected                = veidv1.VerificationStatusRejected
	VerificationStatusPBExpired                 = veidv1.VerificationStatusExpired
	VerificationStatusPBNeedsAdditionalFactor   = veidv1.VerificationStatusNeedsAdditionalFactor
	VerificationStatusPBAdditionalFactorPending = veidv1.VerificationStatusAdditionalFactorPending
)

// IdentityTierPB is the protobuf-generated enum for identity tier
type IdentityTierPB = veidv1.IdentityTier

// Proto enum constants for IdentityTier
const (
	IdentityTierPBUnverified = veidv1.IdentityTierUnverified
	IdentityTierPBBasic      = veidv1.IdentityTierBasic
	IdentityTierPBStandard   = veidv1.IdentityTierStandard
	IdentityTierPBPremium    = veidv1.IdentityTierPremium
)

// AccountStatusPB is the protobuf-generated enum for account status
type AccountStatusPB = veidv1.AccountStatus

// Proto enum constants for AccountStatus
const (
	AccountStatusPBUnknown               = veidv1.AccountStatusUnknown
	AccountStatusPBPending               = veidv1.AccountStatusPending
	AccountStatusPBInProgress            = veidv1.AccountStatusInProgress
	AccountStatusPBVerified              = veidv1.AccountStatusVerified
	AccountStatusPBRejected              = veidv1.AccountStatusRejected
	AccountStatusPBExpired               = veidv1.AccountStatusExpired
	AccountStatusPBNeedsAdditionalFactor = veidv1.AccountStatusNeedsAdditionalFactor
)

// WalletStatusPB is the protobuf-generated enum for wallet status
type WalletStatusPB = veidv1.WalletStatus

// Proto enum constants for WalletStatus
const (
	WalletStatusPBUnspecified = veidv1.WalletStatusUnspecified
	WalletStatusPBActive      = veidv1.WalletStatusActive
	WalletStatusPBSuspended   = veidv1.WalletStatusSuspended
	WalletStatusPBRevoked     = veidv1.WalletStatusRevoked
	WalletStatusPBExpired     = veidv1.WalletStatusExpired
)

// ============================================================================
// Proto Type Aliases - Data Types from types.pb.go
// ============================================================================

// EncryptedPayloadEnvelopePB is the generated proto type for encrypted payloads
type EncryptedPayloadEnvelopePB = veidv1.EncryptedPayloadEnvelope

// UploadMetadataPB is the generated proto type for upload metadata
type UploadMetadataPB = veidv1.UploadMetadata

// ScopeRefPB is the generated proto type for scope references
type ScopeRefPB = veidv1.ScopeRef

// IdentityScopePB is the generated proto type for identity scopes
type IdentityScopePB = veidv1.IdentityScope

// IdentityRecordPB is the generated proto type for identity records
type IdentityRecordPB = veidv1.IdentityRecord

// IdentityScorePB is the generated proto type for identity scores
type IdentityScorePB = veidv1.IdentityScore

// ConsentSettingsPB is the generated proto type for consent settings
type ConsentSettingsPB = veidv1.ConsentSettings

// GlobalConsentUpdatePB is the generated proto type for global consent updates
type GlobalConsentUpdatePB = veidv1.GlobalConsentUpdate

// BorderlineParamsPB is the generated proto type for borderline params
type BorderlineParamsPB = veidv1.BorderlineParams

// ApprovedClientPB is the generated proto type for approved clients
type ApprovedClientPB = veidv1.ApprovedClient

// ParamsPB is the generated proto type for module params
type ParamsPB = veidv1.Params

// ============================================================================
// Proto Type Aliases - Message Types from tx.pb.go
// ============================================================================

// MsgUploadScopePB is the generated proto type for scope upload message
type MsgUploadScopePB = veidv1.MsgUploadScope

// MsgUploadScopeResponsePB is the generated proto response type
type MsgUploadScopeResponsePB = veidv1.MsgUploadScopeResponse

// MsgRevokeScopePB is the generated proto type for scope revoke message
type MsgRevokeScopePB = veidv1.MsgRevokeScope

// MsgRevokeScopeResponsePB is the generated proto response type
type MsgRevokeScopeResponsePB = veidv1.MsgRevokeScopeResponse

// MsgRequestVerificationPB is the generated proto type
type MsgRequestVerificationPB = veidv1.MsgRequestVerification

// MsgRequestVerificationResponsePB is the generated proto response type
type MsgRequestVerificationResponsePB = veidv1.MsgRequestVerificationResponse

// MsgUpdateVerificationStatusPB is the generated proto type
type MsgUpdateVerificationStatusPB = veidv1.MsgUpdateVerificationStatus

// MsgUpdateVerificationStatusResponsePB is the generated proto response type
type MsgUpdateVerificationStatusResponsePB = veidv1.MsgUpdateVerificationStatusResponse

// MsgUpdateScorePB is the generated proto type for score update message
type MsgUpdateScorePB = veidv1.MsgUpdateScore

// MsgUpdateScoreResponsePB is the generated proto response type
type MsgUpdateScoreResponsePB = veidv1.MsgUpdateScoreResponse

// MsgCreateIdentityWalletPB is the generated proto type
type MsgCreateIdentityWalletPB = veidv1.MsgCreateIdentityWallet

// MsgCreateIdentityWalletResponsePB is the generated proto response type
type MsgCreateIdentityWalletResponsePB = veidv1.MsgCreateIdentityWalletResponse

// MsgAddScopeToWalletPB is the generated proto type
type MsgAddScopeToWalletPB = veidv1.MsgAddScopeToWallet

// MsgAddScopeToWalletResponsePB is the generated proto response type
type MsgAddScopeToWalletResponsePB = veidv1.MsgAddScopeToWalletResponse

// MsgRevokeScopeFromWalletPB is the generated proto type
type MsgRevokeScopeFromWalletPB = veidv1.MsgRevokeScopeFromWallet

// MsgRevokeScopeFromWalletResponsePB is the generated proto response type
type MsgRevokeScopeFromWalletResponsePB = veidv1.MsgRevokeScopeFromWalletResponse

// ============================================================================
// Type Conversion Functions - ScopeType
// ============================================================================

// ScopeTypeToProto converts local ScopeType to proto enum
func ScopeTypeToProto(st ScopeType) ScopeTypePB {
	switch st {
	case ScopeTypeIDDocument:
		return ScopeTypePBIDDocument
	case ScopeTypeSelfie:
		return ScopeTypePBSelfie
	case ScopeTypeFaceVideo:
		return ScopeTypePBFaceVideo
	case ScopeTypeBiometric:
		return ScopeTypePBBiometric
	case ScopeTypeBiometricHardware:
		return ScopeTypePBBiometricHardware
	case ScopeTypeDeviceAttestation:
		return ScopeTypePBDeviceAttestation
	case ScopeTypeSSOMetadata:
		return ScopeTypePBSSOMetadata
	case ScopeTypeEmailProof:
		return ScopeTypePBEmailProof
	case ScopeTypeSMSProof:
		return ScopeTypePBSMSProof
	case ScopeTypeDomainVerify:
		return ScopeTypePBDomainVerify
	case ScopeTypeADSSO:
		return ScopeTypePBADSSO
	default:
		return ScopeTypePBUnspecified
	}
}

// ScopeTypeFromProto converts proto enum to local ScopeType
func ScopeTypeFromProto(st ScopeTypePB) ScopeType {
	switch st {
	case ScopeTypePBIDDocument:
		return ScopeTypeIDDocument
	case ScopeTypePBSelfie:
		return ScopeTypeSelfie
	case ScopeTypePBFaceVideo:
		return ScopeTypeFaceVideo
	case ScopeTypePBBiometric:
		return ScopeTypeBiometric
	case ScopeTypePBBiometricHardware:
		return ScopeTypeBiometricHardware
	case ScopeTypePBDeviceAttestation:
		return ScopeTypeDeviceAttestation
	case ScopeTypePBSSOMetadata:
		return ScopeTypeSSOMetadata
	case ScopeTypePBEmailProof:
		return ScopeTypeEmailProof
	case ScopeTypePBSMSProof:
		return ScopeTypeSMSProof
	case ScopeTypePBDomainVerify:
		return ScopeTypeDomainVerify
	case ScopeTypePBADSSO:
		return ScopeTypeADSSO
	default:
		return ""
	}
}

// ============================================================================
// Type Conversion Functions - VerificationStatus
// ============================================================================

// VerificationStatusToProto converts local VerificationStatus to proto enum
func VerificationStatusToProto(vs VerificationStatus) VerificationStatusPB {
	switch vs {
	case VerificationStatusUnknown:
		return VerificationStatusPBUnknown
	case VerificationStatusPending:
		return VerificationStatusPBPending
	case VerificationStatusInProgress:
		return VerificationStatusPBInProgress
	case VerificationStatusVerified:
		return VerificationStatusPBVerified
	case VerificationStatusRejected:
		return VerificationStatusPBRejected
	case VerificationStatusExpired:
		return VerificationStatusPBExpired
	case VerificationStatusNeedsAdditionalFactor:
		return VerificationStatusPBNeedsAdditionalFactor
	case VerificationStatusAdditionalFactorPending:
		return VerificationStatusPBAdditionalFactorPending
	default:
		return VerificationStatusPBUnknown
	}
}

// VerificationStatusFromProto converts proto enum to local VerificationStatus
func VerificationStatusFromProto(vs VerificationStatusPB) VerificationStatus {
	switch vs {
	case VerificationStatusPBUnknown:
		return VerificationStatusUnknown
	case VerificationStatusPBPending:
		return VerificationStatusPending
	case VerificationStatusPBInProgress:
		return VerificationStatusInProgress
	case VerificationStatusPBVerified:
		return VerificationStatusVerified
	case VerificationStatusPBRejected:
		return VerificationStatusRejected
	case VerificationStatusPBExpired:
		return VerificationStatusExpired
	case VerificationStatusPBNeedsAdditionalFactor:
		return VerificationStatusNeedsAdditionalFactor
	case VerificationStatusPBAdditionalFactorPending:
		return VerificationStatusAdditionalFactorPending
	default:
		return VerificationStatusUnknown
	}
}

// ============================================================================
// Type Conversion Functions - IdentityTier
// ============================================================================

// IdentityTierToProto converts local IdentityTier to proto enum
func IdentityTierToProto(it IdentityTier) IdentityTierPB {
	switch it {
	case IdentityTierUnverified:
		return IdentityTierPBUnverified
	case IdentityTierBasic:
		return IdentityTierPBBasic
	case IdentityTierStandard:
		return IdentityTierPBStandard
	case IdentityTierPremium:
		return IdentityTierPBPremium
	default:
		return IdentityTierPBUnverified
	}
}

// IdentityTierFromProto converts proto enum to local IdentityTier
func IdentityTierFromProto(it IdentityTierPB) IdentityTier {
	switch it {
	case IdentityTierPBUnverified:
		return IdentityTierUnverified
	case IdentityTierPBBasic:
		return IdentityTierBasic
	case IdentityTierPBStandard:
		return IdentityTierStandard
	case IdentityTierPBPremium:
		return IdentityTierPremium
	default:
		return IdentityTierUnverified
	}
}

// ============================================================================
// Type Conversion Functions - AccountStatus
// ============================================================================

// AccountStatusToProto converts local AccountStatus to proto enum
func AccountStatusToProto(as AccountStatus) AccountStatusPB {
	switch as {
	case AccountStatusUnknown:
		return AccountStatusPBUnknown
	case AccountStatusPending:
		return AccountStatusPBPending
	case AccountStatusInProgress:
		return AccountStatusPBInProgress
	case AccountStatusVerified:
		return AccountStatusPBVerified
	case AccountStatusRejected:
		return AccountStatusPBRejected
	case AccountStatusExpired:
		return AccountStatusPBExpired
	case AccountStatusNeedsAdditionalFactor:
		return AccountStatusPBNeedsAdditionalFactor
	default:
		return AccountStatusPBUnknown
	}
}

// AccountStatusFromProto converts proto enum to local AccountStatus
func AccountStatusFromProto(as AccountStatusPB) AccountStatus {
	switch as {
	case AccountStatusPBUnknown:
		return AccountStatusUnknown
	case AccountStatusPBPending:
		return AccountStatusPending
	case AccountStatusPBInProgress:
		return AccountStatusInProgress
	case AccountStatusPBVerified:
		return AccountStatusVerified
	case AccountStatusPBRejected:
		return AccountStatusRejected
	case AccountStatusPBExpired:
		return AccountStatusExpired
	case AccountStatusPBNeedsAdditionalFactor:
		return AccountStatusNeedsAdditionalFactor
	default:
		return AccountStatusUnknown
	}
}

// ============================================================================
// Type Conversion Functions - WalletStatus
// ============================================================================

// WalletStatusToProto converts local WalletStatus to proto enum
func WalletStatusToProto(ws WalletStatus) WalletStatusPB {
	switch ws {
	case WalletStatusActive:
		return WalletStatusPBActive
	case WalletStatusSuspended:
		return WalletStatusPBSuspended
	case WalletStatusRevoked:
		return WalletStatusPBRevoked
	case WalletStatusExpired:
		return WalletStatusPBExpired
	default:
		return WalletStatusPBUnspecified
	}
}

// WalletStatusFromProto converts proto enum to local WalletStatus
func WalletStatusFromProto(ws WalletStatusPB) WalletStatus {
	switch ws {
	case WalletStatusPBActive:
		return WalletStatusActive
	case WalletStatusPBSuspended:
		return WalletStatusSuspended
	case WalletStatusPBRevoked:
		return WalletStatusRevoked
	case WalletStatusPBExpired:
		return WalletStatusExpired
	default:
		return ""
	}
}

// ============================================================================
// Helper Functions for Proto Types
// ============================================================================

// NewScopeRefPB creates a new proto ScopeRef
func NewScopeRefPB(scopeID string, scopeType ScopeTypePB, status VerificationStatusPB, uploadedAt int64) *ScopeRefPB {
	return &ScopeRefPB{
		ScopeId:    scopeID,
		ScopeType:  scopeType,
		Status:     status,
		UploadedAt: uploadedAt,
	}
}

// NewIdentityRecordPB creates a new proto IdentityRecord
func NewIdentityRecordPB(accountAddress string, createdAt int64) *IdentityRecordPB {
	return &IdentityRecordPB{
		AccountAddress: accountAddress,
		ScopeRefs:      make([]ScopeRefPB, 0),
		CurrentScore:   0,
		ScoreVersion:   "",
		CreatedAt:      createdAt,
		UpdatedAt:      createdAt,
		Tier:           IdentityTierPBUnverified,
		Flags:          make([]string, 0),
		Locked:         false,
	}
}

// NewIdentityScorePB creates a new proto IdentityScore
func NewIdentityScorePB(accountAddress string, score uint32, status AccountStatusPB, tier IdentityTierPB, modelVersion string, lastUpdatedAt int64, blockHeight int64) *IdentityScorePB {
	return &IdentityScorePB{
		AccountAddress: accountAddress,
		Score:          score,
		Status:         status,
		Tier:           tier,
		ModelVersion:   modelVersion,
		LastUpdatedAt:  lastUpdatedAt,
		BlockHeight:    blockHeight,
	}
}

// NewConsentSettingsPB creates a new proto ConsentSettings with defaults
func NewConsentSettingsPB() *ConsentSettingsPB {
	return &ConsentSettingsPB{
		ShareWithProviders:         false,
		ShareForVerification:       true,
		AllowReVerification:        true,
		AllowDerivedFeatureSharing: false,
		ConsentVersion:             1,
	}
}

// NewBorderlineParamsPB creates a new proto BorderlineParams with defaults
func NewBorderlineParamsPB() *BorderlineParamsPB {
	return &BorderlineParamsPB{
		LowerThreshold:   DefaultBorderlineLowerThreshold,
		UpperThreshold:   DefaultBorderlineUpperThreshold,
		MfaTimeoutBlocks: DefaultMfaTimeoutBlocks,
		RequiredFactors:  DefaultRequiredFactors,
	}
}

// NewParamsPB creates a new proto Params with defaults
func NewParamsPB() *ParamsPB {
	return &ParamsPB{
		MaxScopesPerAccount:    10,
		MaxScopesPerType:       3,
		SaltMinBytes:           16,
		SaltMaxBytes:           64,
		RequireClientSignature: true,
		RequireUserSignature:   true,
		VerificationExpiryDays: 365,
	}
}

// ============================================================================
// Type Conversion Functions - AppealStatus
// ============================================================================

// AppealStatusPB is the protobuf-generated enum for appeal status
type AppealStatusPB = veidv1.AppealStatus

// Proto enum constants for AppealStatus
const (
	AppealStatusPBUnspecified = veidv1.AppealStatusUnspecified
	AppealStatusPBPending     = veidv1.AppealStatusPending
	AppealStatusPBReviewing   = veidv1.AppealStatusReviewing
	AppealStatusPBApproved    = veidv1.AppealStatusApproved
	AppealStatusPBRejected    = veidv1.AppealStatusRejected
	AppealStatusPBWithdrawn   = veidv1.AppealStatusWithdrawn
	AppealStatusPBExpired     = veidv1.AppealStatusExpired
)

// AppealStatusToProto converts local AppealStatus to proto enum
func AppealStatusToProto(as AppealStatus) AppealStatusPB {
	// Since both enums have the same underlying values, we can cast directly
	return AppealStatusPB(as)
}

// AppealStatusFromProto converts proto enum to local AppealStatus
func AppealStatusFromProto(as AppealStatusPB) AppealStatus {
	// Since both enums have the same underlying values, we can cast directly
	return AppealStatus(as)
}

// ============================================================================
// Re-export Proto Msg Server Interface
// ============================================================================

// MsgServerPB is the proto-generated message server interface
type MsgServerPB = veidv1.MsgServer

// UnimplementedMsgServerPB is the proto-generated unimplemented server
type UnimplementedMsgServerPB = veidv1.UnimplementedMsgServer
