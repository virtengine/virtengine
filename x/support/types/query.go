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

// QuerySupportRequestRequest is the request for querying a support request
type QuerySupportRequestRequest struct {
	TicketID string `json:"ticket_id"`
}

// QuerySupportRequestResponse is the response for querying a support request
type QuerySupportRequestResponse struct {
	Request SupportRequest `json:"request"`
}

// QuerySupportRequestsBySubmitterRequest is the request for submitter queries
type QuerySupportRequestsBySubmitterRequest struct {
	SubmitterAddress string `json:"submitter_address"`
	Status           string `json:"status,omitempty"`
}

// QuerySupportRequestsBySubmitterResponse is the response for submitter queries
type QuerySupportRequestsBySubmitterResponse struct {
	Requests []SupportRequest `json:"requests"`
}

// QuerySupportResponsesByRequestRequest is the request for responses by ticket
type QuerySupportResponsesByRequestRequest struct {
	TicketID string `json:"ticket_id"`
}

// QuerySupportResponsesByRequestResponse is the response for responses by ticket
type QuerySupportResponsesByRequestResponse struct {
	Responses []SupportResponse `json:"responses"`
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

func (*QuerySupportRequestRequest) ProtoMessage()    {}
func (m *QuerySupportRequestRequest) Reset()         { *m = QuerySupportRequestRequest{} }
func (m *QuerySupportRequestRequest) String() string { return fmt.Sprintf("%+v", *m) }

func (*QuerySupportRequestResponse) ProtoMessage()    {}
func (m *QuerySupportRequestResponse) Reset()         { *m = QuerySupportRequestResponse{} }
func (m *QuerySupportRequestResponse) String() string { return fmt.Sprintf("%+v", *m) }

func (*QuerySupportRequestsBySubmitterRequest) ProtoMessage() {}
func (m *QuerySupportRequestsBySubmitterRequest) Reset() {
	*m = QuerySupportRequestsBySubmitterRequest{}
}
func (m *QuerySupportRequestsBySubmitterRequest) String() string { return fmt.Sprintf("%+v", *m) }

func (*QuerySupportRequestsBySubmitterResponse) ProtoMessage() {}
func (m *QuerySupportRequestsBySubmitterResponse) Reset() {
	*m = QuerySupportRequestsBySubmitterResponse{}
}
func (m *QuerySupportRequestsBySubmitterResponse) String() string { return fmt.Sprintf("%+v", *m) }

func (*QuerySupportResponsesByRequestRequest) ProtoMessage()    {}
func (m *QuerySupportResponsesByRequestRequest) Reset()         { *m = QuerySupportResponsesByRequestRequest{} }
func (m *QuerySupportResponsesByRequestRequest) String() string { return fmt.Sprintf("%+v", *m) }

func (*QuerySupportResponsesByRequestResponse) ProtoMessage() {}
func (m *QuerySupportResponsesByRequestResponse) Reset() {
	*m = QuerySupportResponsesByRequestResponse{}
}
func (m *QuerySupportResponsesByRequestResponse) String() string { return fmt.Sprintf("%+v", *m) }
