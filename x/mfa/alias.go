package mfa

import (
	"pkg.akt.dev/node/x/mfa/keeper"
	"pkg.akt.dev/node/x/mfa/types"
)

const (
	// ModuleName is the module name constant used in many places
	ModuleName = types.ModuleName

	// StoreKey is the store key string for mfa module
	StoreKey = types.StoreKey

	// RouterKey is the message route for mfa module
	RouterKey = types.RouterKey

	// QuerierRoute is the querier route for mfa module
	QuerierRoute = types.QuerierRoute
)

var (
	// NewKeeper creates a new keeper
	NewKeeper = keeper.NewKeeper

	// NewMFAGatingHooks creates new MFA gating hooks
	NewMFAGatingHooks = keeper.NewMFAGatingHooks

	// RegisterLegacyAminoCodec registers the legacy amino codec
	RegisterLegacyAminoCodec = types.RegisterLegacyAminoCodec

	// RegisterInterfaces registers the interface registry
	RegisterInterfaces = types.RegisterInterfaces

	// DefaultGenesisState returns the default genesis state
	DefaultGenesisState = types.DefaultGenesisState

	// DefaultParams returns the default parameters
	DefaultParams = types.DefaultParams
)

// Type aliases for types package
type (
	// Keeper is the keeper type alias
	Keeper = keeper.Keeper

	// MFAGatingHooks is the gating hooks type alias
	MFAGatingHooks = keeper.MFAGatingHooks

	// GenesisState is the genesis state type alias
	GenesisState = types.GenesisState

	// Params is the params type alias
	Params = types.Params

	// MFAPolicy is the MFA policy type alias
	MFAPolicy = types.MFAPolicy

	// FactorType is the factor type alias
	FactorType = types.FactorType

	// FactorEnrollment is the factor enrollment type alias
	FactorEnrollment = types.FactorEnrollment

	// FactorCombination is the factor combination type alias
	FactorCombination = types.FactorCombination

	// Challenge is the challenge type alias
	Challenge = types.Challenge

	// ChallengeResponse is the challenge response type alias
	ChallengeResponse = types.ChallengeResponse

	// AuthorizationSession is the authorization session type alias
	AuthorizationSession = types.AuthorizationSession

	// TrustedDevice is the trusted device type alias
	TrustedDevice = types.TrustedDevice

	// SensitiveTransactionType is the sensitive tx type alias
	SensitiveTransactionType = types.SensitiveTransactionType

	// SensitiveTxConfig is the sensitive tx config type alias
	SensitiveTxConfig = types.SensitiveTxConfig

	// MFAProof is the MFA proof type alias
	MFAProof = types.MFAProof

	// MsgEnrollFactor is the message type alias
	MsgEnrollFactor = types.MsgEnrollFactor

	// MsgRevokeFactor is the message type alias
	MsgRevokeFactor = types.MsgRevokeFactor

	// MsgSetMFAPolicy is the message type alias
	MsgSetMFAPolicy = types.MsgSetMFAPolicy

	// MsgCreateChallenge is the message type alias
	MsgCreateChallenge = types.MsgCreateChallenge

	// MsgVerifyChallenge is the message type alias
	MsgVerifyChallenge = types.MsgVerifyChallenge

	// MsgAddTrustedDevice is the message type alias
	MsgAddTrustedDevice = types.MsgAddTrustedDevice

	// MsgRemoveTrustedDevice is the message type alias
	MsgRemoveTrustedDevice = types.MsgRemoveTrustedDevice
)

// Factor type constants
const (
	FactorTypeTOTP          = types.FactorTypeTOTP
	FactorTypeFIDO2         = types.FactorTypeFIDO2
	FactorTypeSMS           = types.FactorTypeSMS
	FactorTypeEmail         = types.FactorTypeEmail
	FactorTypeVEID          = types.FactorTypeVEID
	FactorTypeTrustedDevice = types.FactorTypeTrustedDevice
)

// Sensitive transaction type constants
const (
	SensitiveTxAccountRecovery       = types.SensitiveTxAccountRecovery
	SensitiveTxKeyRotation           = types.SensitiveTxKeyRotation
	SensitiveTxLargeWithdrawal       = types.SensitiveTxLargeWithdrawal
	SensitiveTxProviderRegistration  = types.SensitiveTxProviderRegistration
	SensitiveTxValidatorRegistration = types.SensitiveTxValidatorRegistration
	SensitiveTxHighValueOrder        = types.SensitiveTxHighValueOrder
	SensitiveTxRoleAssignment        = types.SensitiveTxRoleAssignment
	SensitiveTxGovernanceProposal    = types.SensitiveTxGovernanceProposal
)
