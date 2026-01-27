package types

import (
	"context"

	"github.com/cosmos/gogoproto/grpc"
)

// QueryRecipientKeyRequest is the request for querying a recipient's public key
type QueryRecipientKeyRequest struct {
	Address string `json:"address"`
}

// QueryRecipientKeyResponse is the response for QueryRecipientKeyRequest
type QueryRecipientKeyResponse struct {
	Keys []RecipientKeyRecord `json:"keys"`
}

// QueryKeyByFingerprintRequest is the request for querying a key by fingerprint
type QueryKeyByFingerprintRequest struct {
	Fingerprint string `json:"fingerprint"`
}

// QueryKeyByFingerprintResponse is the response for QueryKeyByFingerprintRequest
type QueryKeyByFingerprintResponse struct {
	Key RecipientKeyRecord `json:"key"`
}

// QueryParamsRequest is the request for querying module parameters
type QueryParamsRequest struct{}

// QueryParamsResponse is the response for QueryParamsRequest
type QueryParamsResponse struct {
	Params Params `json:"params"`
}

// QueryAlgorithmsRequest is the request for querying supported algorithms
type QueryAlgorithmsRequest struct{}

// QueryAlgorithmsResponse is the response for QueryAlgorithmsRequest
type QueryAlgorithmsResponse struct {
	Algorithms []AlgorithmInfo `json:"algorithms"`
}

// QueryValidateEnvelopeRequest is the request for validating an envelope
type QueryValidateEnvelopeRequest struct {
	Envelope EncryptedPayloadEnvelope `json:"envelope"`
}

// QueryValidateEnvelopeResponse is the response for QueryValidateEnvelopeRequest
type QueryValidateEnvelopeResponse struct {
	Valid            bool     `json:"valid"`
	Error            string   `json:"error,omitempty"`
	RecipientCount   int      `json:"recipient_count"`
	Algorithm        string   `json:"algorithm"`
	SignatureValid   bool     `json:"signature_valid"`
	AllKeysRegistered bool    `json:"all_keys_registered"`
	MissingKeys      []string `json:"missing_keys,omitempty"`
}

// QueryServer is the query server interface
type QueryServer interface {
	RecipientKey(ctx context.Context, req *QueryRecipientKeyRequest) (*QueryRecipientKeyResponse, error)
	KeyByFingerprint(ctx context.Context, req *QueryKeyByFingerprintRequest) (*QueryKeyByFingerprintResponse, error)
	Params(ctx context.Context, req *QueryParamsRequest) (*QueryParamsResponse, error)
	Algorithms(ctx context.Context, req *QueryAlgorithmsRequest) (*QueryAlgorithmsResponse, error)
	ValidateEnvelope(ctx context.Context, req *QueryValidateEnvelopeRequest) (*QueryValidateEnvelopeResponse, error)
}

// RegisterQueryServer registers the QueryServer
// This is a stub implementation until proper protobuf generation is set up.
func RegisterQueryServer(s grpc.Server, impl QueryServer) {
	// Registration is a no-op for now since we don't have proper protobuf generated code
	_ = s
	_ = impl
}

// _Query_serviceDesc is the grpc.ServiceDesc for Query service.
var _Query_serviceDesc = struct {
	ServiceName string
	HandlerType interface{}
	Methods     []struct {
		MethodName string
		Handler    interface{}
	}
	Streams  []struct{}
	Metadata interface{}
}{
	ServiceName: "virtengine.encryption.v1.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []struct {
		MethodName string
		Handler    interface{}
	}{
		{MethodName: "RecipientKey", Handler: nil},
		{MethodName: "KeyByFingerprint", Handler: nil},
		{MethodName: "Params", Handler: nil},
		{MethodName: "Algorithms", Handler: nil},
		{MethodName: "ValidateEnvelope", Handler: nil},
	},
	Streams:  []struct{}{},
	Metadata: "virtengine/encryption/v1/query.proto",
}
