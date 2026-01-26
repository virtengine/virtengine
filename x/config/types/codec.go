package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)

func init() {
	RegisterLegacyAminoCodec(amino)
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

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// _Msg_serviceDesc is the grpc.ServiceDesc for Msg service.
var _Msg_serviceDesc = struct {
	ServiceName string
	HandlerType interface{}
	Methods     []struct {
		MethodName string
		Handler    interface{}
	}
	Streams  []struct{}
	Metadata interface{}
}{
	ServiceName: "virtengine.config.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []struct {
		MethodName string
		Handler    interface{}
	}{
		{MethodName: "RegisterApprovedClient", Handler: nil},
		{MethodName: "UpdateApprovedClient", Handler: nil},
		{MethodName: "SuspendApprovedClient", Handler: nil},
		{MethodName: "RevokeApprovedClient", Handler: nil},
		{MethodName: "ReactivateApprovedClient", Handler: nil},
		{MethodName: "UpdateParams", Handler: nil},
	},
	Streams:  []struct{}{},
	Metadata: "virtengine/config/v1/msg.proto",
}

// MsgServer is the interface for the message server
type MsgServer interface {
	RegisterApprovedClient(ctx sdk.Context, msg *MsgRegisterApprovedClient) (*MsgRegisterApprovedClientResponse, error)
	UpdateApprovedClient(ctx sdk.Context, msg *MsgUpdateApprovedClient) (*MsgUpdateApprovedClientResponse, error)
	SuspendApprovedClient(ctx sdk.Context, msg *MsgSuspendApprovedClient) (*MsgSuspendApprovedClientResponse, error)
	RevokeApprovedClient(ctx sdk.Context, msg *MsgRevokeApprovedClient) (*MsgRevokeApprovedClientResponse, error)
	ReactivateApprovedClient(ctx sdk.Context, msg *MsgReactivateApprovedClient) (*MsgReactivateApprovedClientResponse, error)
	UpdateParams(ctx sdk.Context, msg *MsgUpdateParams) (*MsgUpdateParamsResponse, error)
}

// RegisterMsgServer registers the MsgServer
func RegisterMsgServer(s interface{ RegisterService(desc interface{}, impl interface{}) }, impl MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, impl)
}

// QueryServer is the interface for the query server
type QueryServer interface {
	ApprovedClient(ctx sdk.Context, req *QueryApprovedClientRequest) (*QueryApprovedClientResponse, error)
	ApprovedClients(ctx sdk.Context, req *QueryApprovedClientsRequest) (*QueryApprovedClientsResponse, error)
	ApprovedClientsByStatus(ctx sdk.Context, req *QueryApprovedClientsByStatusRequest) (*QueryApprovedClientsByStatusResponse, error)
	ValidateClientSignature(ctx sdk.Context, req *QueryValidateClientSignatureRequest) (*QueryValidateClientSignatureResponse, error)
	ValidateClientVersion(ctx sdk.Context, req *QueryValidateClientVersionRequest) (*QueryValidateClientVersionResponse, error)
	Params(ctx sdk.Context, req *QueryParamsRequest) (*QueryParamsResponse, error)
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
	ServiceName: "virtengine.config.v1.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []struct {
		MethodName string
		Handler    interface{}
	}{
		{MethodName: "ApprovedClient", Handler: nil},
		{MethodName: "ApprovedClients", Handler: nil},
		{MethodName: "ApprovedClientsByStatus", Handler: nil},
		{MethodName: "ValidateClientSignature", Handler: nil},
		{MethodName: "ValidateClientVersion", Handler: nil},
		{MethodName: "Params", Handler: nil},
	},
	Streams:  []struct{}{},
	Metadata: "virtengine/config/v1/query.proto",
}

// RegisterQueryServer registers the QueryServer
func RegisterQueryServer(s interface{ RegisterService(desc interface{}, impl interface{}) }, impl QueryServer) {
	s.RegisterService(&_Query_serviceDesc, impl)
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
