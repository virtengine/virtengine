package encryption

import (
	"github.com/virtengine/virtengine/x/encryption/keeper"
	"github.com/virtengine/virtengine/x/encryption/types"
)

// Alias types for external access
type (
	Keeper                   = keeper.Keeper
	EncryptedPayloadEnvelope = types.EncryptedPayloadEnvelope
	RecipientKeyRecord       = types.RecipientKeyRecord
	GenesisState             = types.GenesisState
	Params                   = types.Params
	AlgorithmInfo            = types.AlgorithmInfo
	MsgRegisterRecipientKey  = types.MsgRegisterRecipientKey
	MsgRevokeRecipientKey    = types.MsgRevokeRecipientKey
	MsgUpdateKeyLabel        = types.MsgUpdateKeyLabel
	MsgRotateKey             = types.MsgRotateKey
	EnvelopeMetadata         = types.EnvelopeMetadata
	EnvelopeValidationResult = types.EnvelopeValidationResult
	RecipientStatus          = types.RecipientStatus
)

// Alias constants
const (
	ModuleName   = types.ModuleName
	StoreKey     = types.StoreKey
	RouterKey    = types.RouterKey
	QuerierRoute = types.QuerierRoute

	AlgorithmX25519XSalsa20Poly1305 = types.AlgorithmX25519XSalsa20Poly1305
	AlgorithmAgeX25519              = types.AlgorithmAgeX25519
)

// Alias functions
var (
	NewKeeper                   = keeper.NewKeeper
	NewMsgServerImpl            = keeper.NewMsgServerImpl
	NewEncryptedPayloadEnvelope = types.NewEncryptedPayloadEnvelope
	DefaultParams               = types.DefaultParams
	ComputeKeyFingerprint       = types.ComputeKeyFingerprint
	SupportedAlgorithms         = types.SupportedAlgorithms
	IsAlgorithmSupported        = types.IsAlgorithmSupported
	DefaultAlgorithm            = types.DefaultAlgorithm
	GetAlgorithmInfo            = types.GetAlgorithmInfo
	RegisterInterfaces          = types.RegisterInterfaces
	RegisterLegacyAminoCodec    = types.RegisterLegacyAminoCodec
	ExtractEnvelopeMetadata     = types.ExtractEnvelopeMetadata
	ValidateParams              = types.ValidateParams
)

// Alias error variables
var (
	ErrInvalidAddress        = types.ErrInvalidAddress
	ErrInvalidPublicKey      = types.ErrInvalidPublicKey
	ErrKeyNotFound           = types.ErrKeyNotFound
	ErrKeyAlreadyExists      = types.ErrKeyAlreadyExists
	ErrKeyRevoked            = types.ErrKeyRevoked
	ErrKeyDeprecated         = types.ErrKeyDeprecated
	ErrKeyExpired            = types.ErrKeyExpired
	ErrInvalidEnvelope       = types.ErrInvalidEnvelope
	ErrUnsupportedAlgorithm  = types.ErrUnsupportedAlgorithm
	ErrUnsupportedVersion    = types.ErrUnsupportedVersion
	ErrInvalidSignature      = types.ErrInvalidSignature
	ErrEncryptionFailed      = types.ErrEncryptionFailed
	ErrDecryptionFailed      = types.ErrDecryptionFailed
	ErrInvalidNonce          = types.ErrInvalidNonce
	ErrUnauthorized          = types.ErrUnauthorized
	ErrNotRecipient          = types.ErrNotRecipient
	ErrReencryptionJobFailed = types.ErrReencryptionJobFailed
	ErrUnauthorizedAccess    = types.ErrUnauthorizedAccess
)
