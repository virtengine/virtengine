package types

import (
"github.com/cosmos/cosmos-sdk/codec"
"github.com/cosmos/cosmos-sdk/codec/legacy"
cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
sdk "github.com/cosmos/cosmos-sdk/types"
"github.com/cosmos/cosmos-sdk/types/msgservice"
"github.com/cosmos/gogoproto/grpc"

configv1 "github.com/virtengine/virtengine/sdk/go/node/config/v1"
)

// ModuleCdc is the codec for the module
var ModuleCdc = codec.NewLegacyAmino()

func init() {
RegisterLegacyAminoCodec(ModuleCdc)
}

// RegisterLegacyAminoCodec registers the necessary interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
legacy.RegisterAminoMsg(cdc, &MsgRegisterApprovedClient{}, "config/MsgRegisterApprovedClient")
legacy.RegisterAminoMsg(cdc, &MsgUpdateApprovedClient{}, "config/MsgUpdateApprovedClient")
legacy.RegisterAminoMsg(cdc, &MsgSuspendApprovedClient{}, "config/MsgSuspendApprovedClient")
legacy.RegisterAminoMsg(cdc, &MsgRevokeApprovedClient{}, "config/MsgRevokeApprovedClient")
legacy.RegisterAminoMsg(cdc, &MsgReactivateApprovedClient{}, "config/MsgReactivateApprovedClient")
legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "config/MsgUpdateParams")
}

// RegisterInterfaces registers the interfaces types with the interface registry.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
registry.RegisterImplementations((*sdk.Msg)(nil),
&MsgRegisterApprovedClient{},
&MsgUpdateApprovedClient{},
&MsgSuspendApprovedClient{},
&MsgRevokeApprovedClient{},
&MsgReactivateApprovedClient{},
&MsgUpdateParams{},
)

msgservice.RegisterMsgServiceDesc(registry, &configv1.Msg_serviceDesc)
}

// MsgServer is the server interface for config module messages
type MsgServer = configv1.MsgServer

// RegisterMsgServer registers the MsgServer implementation
func RegisterMsgServer(s grpc.Server, srv MsgServer) {
configv1.RegisterMsgServer(s, srv)
}

// QueryServer is the server interface for config module queries
type QueryServer interface {
ApprovedClient(ctx sdk.Context, req *QueryApprovedClientRequest) (*QueryApprovedClientResponse, error)
ApprovedClients(ctx sdk.Context, req *QueryApprovedClientsRequest) (*QueryApprovedClientsResponse, error)
ApprovedClientsByStatus(ctx sdk.Context, req *QueryApprovedClientsByStatusRequest) (*QueryApprovedClientsByStatusResponse, error)
ValidateClientSignature(ctx sdk.Context, req *QueryValidateClientSignatureRequest) (*QueryValidateClientSignatureResponse, error)
ValidateClientVersion(ctx sdk.Context, req *QueryValidateClientVersionRequest) (*QueryValidateClientVersionResponse, error)
Params(ctx sdk.Context, req *QueryParamsRequest) (*QueryParamsResponse, error)
}

// RegisterQueryServer registers the QueryServer implementation
func RegisterQueryServer(s grpc.Server, srv QueryServer) {
// Query server registration is a no-op until proper protobuf generation
_ = s
_ = srv
}

// Query request/response types

// QueryApprovedClientRequest is the request for querying a single approved client
type QueryApprovedClientRequest struct {
ClientID string `json:"client_id"`
}

// QueryApprovedClientResponse is the response for QueryApprovedClientRequest
type QueryApprovedClientResponse struct {
Client ApprovedClient `json:"client"`
}

// QueryApprovedClientsRequest is the request for querying all approved clients
type QueryApprovedClientsRequest struct {
// Pagination options can be added here
}

// QueryApprovedClientsResponse is the response for QueryApprovedClientsRequest
type QueryApprovedClientsResponse struct {
Clients []ApprovedClient `json:"clients"`
}

// QueryApprovedClientsByStatusRequest is the request for querying clients by status
type QueryApprovedClientsByStatusRequest struct {
Status ClientStatus `json:"status"`
}

// QueryApprovedClientsByStatusResponse is the response for QueryApprovedClientsByStatusRequest
type QueryApprovedClientsByStatusResponse struct {
Clients []ApprovedClient `json:"clients"`
}

// QueryValidateClientSignatureRequest is the request for validating a client signature
type QueryValidateClientSignatureRequest struct {
ClientID    string `json:"client_id"`
Signature   []byte `json:"signature"`
PayloadHash []byte `json:"payload_hash"`
}

// QueryValidateClientSignatureResponse is the response for QueryValidateClientSignatureRequest
type QueryValidateClientSignatureResponse struct {
Valid   bool   `json:"valid"`
Message string `json:"message,omitempty"`
}

// QueryValidateClientVersionRequest is the request for validating a client version
type QueryValidateClientVersionRequest struct {
ClientID string `json:"client_id"`
Version  string `json:"version"`
}

// QueryValidateClientVersionResponse is the response for QueryValidateClientVersionRequest
type QueryValidateClientVersionResponse struct {
Valid   bool   `json:"valid"`
Message string `json:"message,omitempty"`
}

// QueryParamsRequest is the request for querying module parameters
type QueryParamsRequest struct{}

// QueryParamsResponse is the response for QueryParamsRequest
type QueryParamsResponse struct {
Params Params `json:"params"`
}