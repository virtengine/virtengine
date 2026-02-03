package types

import "fmt"

// QueryExternalRefRequest is the request for querying a single external ticket reference
type QueryExternalRefRequest struct {
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
}

// QueryExternalRefResponse is the response for QueryExternalRefRequest
type QueryExternalRefResponse struct {
	Ref ExternalTicketRef `json:"ref"`
}

// QueryExternalRefsByOwnerRequest is the request for querying external refs by owner
type QueryExternalRefsByOwnerRequest struct {
	OwnerAddress string `json:"owner_address"`
	ResourceType string `json:"resource_type,omitempty"` // Optional filter by resource type
}

// QueryExternalRefsByOwnerResponse is the response for QueryExternalRefsByOwnerRequest
type QueryExternalRefsByOwnerResponse struct {
	Refs []ExternalTicketRef `json:"refs"`
}

// QueryParamsRequest is the request for querying module parameters
type QueryParamsRequest struct{}

// QueryParamsResponse is the response for QueryParamsRequest
type QueryParamsResponse struct {
	Params Params `json:"params"`
}

// Proto message interface stubs

func (*QueryExternalRefRequest) ProtoMessage()    {}
func (m *QueryExternalRefRequest) Reset()         { *m = QueryExternalRefRequest{} }
func (m *QueryExternalRefRequest) String() string { return fmt.Sprintf("%+v", *m) }

func (*QueryExternalRefResponse) ProtoMessage()    {}
func (m *QueryExternalRefResponse) Reset()         { *m = QueryExternalRefResponse{} }
func (m *QueryExternalRefResponse) String() string { return fmt.Sprintf("%+v", *m) }

func (*QueryExternalRefsByOwnerRequest) ProtoMessage()    {}
func (m *QueryExternalRefsByOwnerRequest) Reset()         { *m = QueryExternalRefsByOwnerRequest{} }
func (m *QueryExternalRefsByOwnerRequest) String() string { return fmt.Sprintf("%+v", *m) }

func (*QueryExternalRefsByOwnerResponse) ProtoMessage()    {}
func (m *QueryExternalRefsByOwnerResponse) Reset()         { *m = QueryExternalRefsByOwnerResponse{} }
func (m *QueryExternalRefsByOwnerResponse) String() string { return fmt.Sprintf("%+v", *m) }

func (*QueryParamsRequest) ProtoMessage()    {}
func (m *QueryParamsRequest) Reset()         { *m = QueryParamsRequest{} }
func (m *QueryParamsRequest) String() string { return fmt.Sprintf("%+v", *m) }

func (*QueryParamsResponse) ProtoMessage()    {}
func (m *QueryParamsResponse) Reset()         { *m = QueryParamsResponse{} }
func (m *QueryParamsResponse) String() string { return fmt.Sprintf("%+v", *m) }
