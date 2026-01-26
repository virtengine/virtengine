package veid

import (
	"pkg.akt.dev/node/x/veid/keeper"
	"pkg.akt.dev/node/x/veid/types"
)

// Aliases for types
const (
	ModuleName = types.ModuleName
	StoreKey   = types.StoreKey
	RouterKey  = types.RouterKey
)

// Aliases for scope types
const (
	ScopeTypeIDDocument   = types.ScopeTypeIDDocument
	ScopeTypeSelfie       = types.ScopeTypeSelfie
	ScopeTypeFaceVideo    = types.ScopeTypeFaceVideo
	ScopeTypeBiometric    = types.ScopeTypeBiometric
	ScopeTypeSSOMetadata  = types.ScopeTypeSSOMetadata
	ScopeTypeEmailProof   = types.ScopeTypeEmailProof
	ScopeTypeSMSProof     = types.ScopeTypeSMSProof
	ScopeTypeDomainVerify = types.ScopeTypeDomainVerify
)

// Aliases for verification statuses
const (
	VerificationStatusUnknown    = types.VerificationStatusUnknown
	VerificationStatusPending    = types.VerificationStatusPending
	VerificationStatusInProgress = types.VerificationStatusInProgress
	VerificationStatusVerified   = types.VerificationStatusVerified
	VerificationStatusRejected   = types.VerificationStatusRejected
	VerificationStatusExpired    = types.VerificationStatusExpired
)

// Aliases for identity tiers
const (
	IdentityTierUnverified = types.IdentityTierUnverified
	IdentityTierBasic      = types.IdentityTierBasic
	IdentityTierStandard   = types.IdentityTierStandard
	IdentityTierVerified   = types.IdentityTierVerified
	IdentityTierTrusted    = types.IdentityTierTrusted
)

// Type aliases
type (
	// Keeper
	Keeper = keeper.Keeper

	// Types
	ScopeType          = types.ScopeType
	VerificationStatus = types.VerificationStatus
	IdentityTier       = types.IdentityTier
	IdentityScope      = types.IdentityScope
	IdentityRecord     = types.IdentityRecord
	IdentityWallet     = types.IdentityWallet
	UploadMetadata     = types.UploadMetadata
	ApprovedClient     = types.ApprovedClient
	VerificationEvent  = types.VerificationEvent
	VerificationResult = types.VerificationResult
	ScopeRef           = types.ScopeRef

	// Messages
	MsgUploadScope                  = types.MsgUploadScope
	MsgUploadScopeResponse          = types.MsgUploadScopeResponse
	MsgRevokeScope                  = types.MsgRevokeScope
	MsgRevokeScopeResponse          = types.MsgRevokeScopeResponse
	MsgRequestVerification          = types.MsgRequestVerification
	MsgRequestVerificationResponse  = types.MsgRequestVerificationResponse
	MsgUpdateVerificationStatus     = types.MsgUpdateVerificationStatus
	MsgUpdateVerificationStatusResponse = types.MsgUpdateVerificationStatusResponse
	MsgUpdateScore                  = types.MsgUpdateScore
	MsgUpdateScoreResponse          = types.MsgUpdateScoreResponse

	// Genesis
	GenesisState = types.GenesisState
	Params       = types.Params

	// VE-217: Derived Feature Minimization Types
	EmbeddingType                        = types.EmbeddingType
	EmbeddingEnvelope                    = types.EmbeddingEnvelope
	EmbeddingEnvelopeReference           = types.EmbeddingEnvelopeReference
	RetentionPolicy                      = types.RetentionPolicy
	RetentionType                        = types.RetentionType
	DataLifecycleRules                   = types.DataLifecycleRules
	ArtifactType                         = types.ArtifactType
	ArtifactRetentionRule                = types.ArtifactRetentionRule
	DerivedFeatureVerificationRecord     = types.DerivedFeatureVerificationRecord
	DerivedFeatureReference              = types.DerivedFeatureReference
	FeatureMatchResult                   = types.FeatureMatchResult
	ConsensusVote                        = types.ConsensusVote

	// Events
	EventScopeUploaded         = types.EventScopeUploaded
	EventScopeRevoked          = types.EventScopeRevoked
	EventScopeVerified         = types.EventScopeVerified
	EventScopeRejected         = types.EventScopeRejected
	EventStatusUpdated         = types.EventStatusUpdated
	EventScoreUpdated          = types.EventScoreUpdated
	EventIdentityCreated       = types.EventIdentityCreated
	EventIdentityLocked        = types.EventIdentityLocked
	EventIdentityUnlocked      = types.EventIdentityUnlocked
	EventVerificationRequested = types.EventVerificationRequested
)

// Function aliases
var (
	NewKeeper              = keeper.NewKeeper
	NewMsgServerImpl       = keeper.NewMsgServerImpl
	NewMsgServerWithContext = keeper.NewMsgServerWithContext

	// Types
	NewIdentityScope        = types.NewIdentityScope
	NewIdentityRecord       = types.NewIdentityRecord
	NewIdentityWallet       = types.NewIdentityWallet
	NewUploadMetadata       = types.NewUploadMetadata
	NewApprovedClient       = types.NewApprovedClient
	NewVerificationEvent    = types.NewVerificationEvent
	NewVerificationResult   = types.NewVerificationResult
	NewScopeRef             = types.NewScopeRef

	// Messages
	NewMsgUploadScope                  = types.NewMsgUploadScope
	NewMsgRevokeScope                  = types.NewMsgRevokeScope
	NewMsgRequestVerification          = types.NewMsgRequestVerification
	NewMsgUpdateVerificationStatus     = types.NewMsgUpdateVerificationStatus
	NewMsgUpdateScore                  = types.NewMsgUpdateScore

	// Validation
	IsValidScopeType          = types.IsValidScopeType
	IsValidVerificationStatus = types.IsValidVerificationStatus
	IsValidIdentityTier       = types.IsValidIdentityTier
	AllScopeTypes             = types.AllScopeTypes
	AllVerificationStatuses   = types.AllVerificationStatuses
	AllIdentityTiers          = types.AllIdentityTiers

	// Utilities
	ComputeTierFromScore  = types.ComputeTierFromScore
	TierMinimumScore      = types.TierMinimumScore
	ScopeTypeWeight       = types.ScopeTypeWeight
	ScopeTypeDescription  = types.ScopeTypeDescription
	ComputeSaltHash       = types.ComputeSaltHash

	// Genesis
	DefaultGenesisState = types.DefaultGenesisState
	DefaultParams       = types.DefaultParams

	// VE-217: Derived Feature Minimization Functions
	NewEmbeddingEnvelope                    = types.NewEmbeddingEnvelope
	NewDerivedFeatureVerificationRecord     = types.NewDerivedFeatureVerificationRecord
	NewDerivedFeatureReference              = types.NewDerivedFeatureReference
	NewRetentionPolicyDuration              = types.NewRetentionPolicyDuration
	NewRetentionPolicyBlockCount            = types.NewRetentionPolicyBlockCount
	NewRetentionPolicyIndefinite            = types.NewRetentionPolicyIndefinite
	NewRetentionPolicyUntilRevoked          = types.NewRetentionPolicyUntilRevoked
	DefaultDataLifecycleRules               = types.DefaultDataLifecycleRules
	ComputeEmbeddingHash                    = types.ComputeEmbeddingHash
	IsValidEmbeddingType                    = types.IsValidEmbeddingType
	IsValidRetentionType                    = types.IsValidRetentionType
	IsValidArtifactType                     = types.IsValidArtifactType
	IsValidFeatureMatchResult               = types.IsValidFeatureMatchResult
	AllEmbeddingTypes                       = types.AllEmbeddingTypes
	AllRetentionTypes                       = types.AllRetentionTypes
	AllArtifactTypes                        = types.AllArtifactTypes
	AllFeatureMatchResults                  = types.AllFeatureMatchResults

	// Codec
	RegisterLegacyAminoCodec = types.RegisterLegacyAminoCodec
	RegisterInterfaces       = types.RegisterInterfaces
)
