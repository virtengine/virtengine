package config

import (
	"github.com/virtengine/virtengine/x/config/keeper"
	"github.com/virtengine/virtengine/x/config/types"
)

// Module aliases
const (
	ModuleName = types.ModuleName
	StoreKey   = types.StoreKey
	RouterKey  = types.RouterKey
)

// Keeper type aliases
type (
	Keeper = keeper.Keeper
)

// Type aliases
type (
	ApprovedClient   = types.ApprovedClient
	ClientStatus     = types.ClientStatus
	KeyType          = types.KeyType
	Params           = types.Params
	GenesisState     = types.GenesisState
	AuditEntry       = types.AuditEntry
	VersionConstraint = types.VersionConstraint

	MsgRegisterApprovedClient         = types.MsgRegisterApprovedClient
	MsgUpdateApprovedClient           = types.MsgUpdateApprovedClient
	MsgSuspendApprovedClient          = types.MsgSuspendApprovedClient
	MsgRevokeApprovedClient           = types.MsgRevokeApprovedClient
	MsgReactivateApprovedClient       = types.MsgReactivateApprovedClient
	MsgRegisterApprovedClientResponse = types.MsgRegisterApprovedClientResponse
	MsgUpdateApprovedClientResponse   = types.MsgUpdateApprovedClientResponse
	MsgSuspendApprovedClientResponse  = types.MsgSuspendApprovedClientResponse
	MsgRevokeApprovedClientResponse   = types.MsgRevokeApprovedClientResponse
	MsgReactivateApprovedClientResponse = types.MsgReactivateApprovedClientResponse

	MsgServer   = types.MsgServer
	QueryServer = types.QueryServer
)

// Status constants
const (
	ClientStatusActive    = types.ClientStatusActive
	ClientStatusSuspended = types.ClientStatusSuspended
	ClientStatusRevoked   = types.ClientStatusRevoked
)

// Key type constants
const (
	KeyTypeEd25519   = types.KeyTypeEd25519
	KeyTypeSecp256k1 = types.KeyTypeSecp256k1
)

// Function aliases
var (
	NewKeeper              = keeper.NewKeeper
	NewMsgServerImpl       = keeper.NewMsgServerImpl
	NewApprovedClient      = types.NewApprovedClient
	NewVersionConstraint   = types.NewVersionConstraint
	DefaultGenesisState    = types.DefaultGenesisState
	DefaultParams          = types.DefaultParams
	RegisterInterfaces     = types.RegisterInterfaces
	RegisterLegacyAminoCodec = types.RegisterLegacyAminoCodec
)

// Error aliases
var (
	ErrInvalidClientID          = types.ErrInvalidClientID
	ErrClientNotFound           = types.ErrClientNotFound
	ErrClientAlreadyExists      = types.ErrClientAlreadyExists
	ErrInvalidStatusTransition  = types.ErrInvalidStatusTransition
	ErrUnauthorized             = types.ErrUnauthorized
	ErrClientNotApproved        = types.ErrClientNotApproved
	ErrClientSuspended          = types.ErrClientSuspended
	ErrClientRevoked            = types.ErrClientRevoked
	ErrInvalidSignature         = types.ErrInvalidSignature
	ErrVersionNotAllowed        = types.ErrVersionNotAllowed
	ErrScopeNotAllowed          = types.ErrScopeNotAllowed
	ErrSignatureVerificationFailed = types.ErrSignatureVerificationFailed
)
