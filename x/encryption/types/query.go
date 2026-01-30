package types

import (
	encryptionv1 "github.com/virtengine/virtengine/sdk/go/node/encryption/v1"
)

// Type aliases for generated query types
type (
	QueryRecipientKeyRequest      = encryptionv1.QueryRecipientKeyRequest
	QueryRecipientKeyResponse     = encryptionv1.QueryRecipientKeyResponse
	QueryKeyByFingerprintRequest  = encryptionv1.QueryKeyByFingerprintRequest
	QueryKeyByFingerprintResponse = encryptionv1.QueryKeyByFingerprintResponse
	QueryParamsRequest            = encryptionv1.QueryParamsRequest
	QueryParamsResponse           = encryptionv1.QueryParamsResponse
	QueryAlgorithmsRequest        = encryptionv1.QueryAlgorithmsRequest
	QueryAlgorithmsResponse       = encryptionv1.QueryAlgorithmsResponse
	QueryValidateEnvelopeRequest  = encryptionv1.QueryValidateEnvelopeRequest
	QueryValidateEnvelopeResponse = encryptionv1.QueryValidateEnvelopeResponse
)

// QueryServer is the interface for the query server - alias to generated type
type QueryServer = encryptionv1.QueryServer

// RegisterQueryServer registers the QueryServer implementation
var RegisterQueryServer = encryptionv1.RegisterQueryServer
