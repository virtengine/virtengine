package encryption

import (
	"pkg.akt.dev/node/x/encryption/keeper"
	"pkg.akt.dev/node/x/encryption/types"
)

// Alias types for external access
type (
	Keeper                     = keeper.Keeper
	EncryptedPayloadEnvelope   = types.EncryptedPayloadEnvelope
	RecipientKeyRecord         = types.RecipientKeyRecord
	GenesisState               = types.GenesisState
	Params                     = types.Params
	AlgorithmInfo              = types.AlgorithmInfo
	MsgRegisterRecipientKey    = types.MsgRegisterRecipientKey
	MsgRevokeRecipientKey      = types.MsgRevokeRecipientKey
	MsgUpdateKeyLabel          = types.MsgUpdateKeyLabel
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
	NewKeeper                    = keeper.NewKeeper
	NewMsgServerImpl             = keeper.NewMsgServerImpl
	NewEncryptedPayloadEnvelope  = types.NewEncryptedPayloadEnvelope
	DefaultGenesisState          = types.DefaultGenesisState
	DefaultParams                = types.DefaultParams
	ComputeKeyFingerprint        = types.ComputeKeyFingerprint
	SupportedAlgorithms          = types.SupportedAlgorithms
	IsAlgorithmSupported         = types.IsAlgorithmSupported
	DefaultAlgorithm             = types.DefaultAlgorithm
	GetAlgorithmInfo             = types.GetAlgorithmInfo
	RegisterInterfaces           = types.RegisterInterfaces
	RegisterLegacyAminoCodec     = types.RegisterLegacyAminoCodec
)

// Alias error variables
var (
	ErrInvalidAddress       = types.ErrInvalidAddress
	ErrInvalidPublicKey     = types.ErrInvalidPublicKey
	ErrKeyNotFound          = types.ErrKeyNotFound
	ErrKeyAlreadyExists     = types.ErrKeyAlreadyExists
	ErrKeyRevoked           = types.ErrKeyRevoked
	ErrInvalidEnvelope      = types.ErrInvalidEnvelope
	ErrUnsupportedAlgorithm = types.ErrUnsupportedAlgorithm
	ErrUnsupportedVersion   = types.ErrUnsupportedVersion
	ErrInvalidSignature     = types.ErrInvalidSignature
	ErrEncryptionFailed     = types.ErrEncryptionFailed
	ErrDecryptionFailed     = types.ErrDecryptionFailed
	ErrInvalidNonce         = types.ErrInvalidNonce
	ErrUnauthorized         = types.ErrUnauthorized
	ErrNotRecipient         = types.ErrNotRecipient
)
