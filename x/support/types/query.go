package types

import "fmt"

// QueryTicketRequest is the request for querying a single ticket
type QueryTicketRequest struct {
	TicketID string `json:"ticket_id"`
}

// QueryTicketResponse is the response for QueryTicketRequest
type QueryTicketResponse struct {
	Ticket SupportTicket `json:"ticket"`
}

// QueryTicketsByCustomerRequest is the request for querying tickets by customer
type QueryTicketsByCustomerRequest struct {
	CustomerAddress string `json:"customer_address"`
	Status          string `json:"status,omitempty"` // Optional status filter
}

// QueryTicketsByCustomerResponse is the response for QueryTicketsByCustomerRequest
type QueryTicketsByCustomerResponse struct {
	Tickets []SupportTicket `json:"tickets"`
}

// QueryTicketsByProviderRequest is the request for querying tickets by provider
type QueryTicketsByProviderRequest struct {
	ProviderAddress string `json:"provider_address"`
	Status          string `json:"status,omitempty"` // Optional status filter
}

// QueryTicketsByProviderResponse is the response for QueryTicketsByProviderRequest
type QueryTicketsByProviderResponse struct {
	Tickets []SupportTicket `json:"tickets"`
}

// QueryTicketsByAgentRequest is the request for querying tickets by assigned agent
type QueryTicketsByAgentRequest struct {
	AgentAddress string `json:"agent_address"`
	Status       string `json:"status,omitempty"` // Optional status filter
}

// QueryTicketsByAgentResponse is the response for QueryTicketsByAgentRequest
type QueryTicketsByAgentResponse struct {
	Tickets []SupportTicket `json:"tickets"`
}

// QueryTicketsByStatusRequest is the request for querying tickets by status
type QueryTicketsByStatusRequest struct {
	Status string `json:"status"`
}

// QueryTicketsByStatusResponse is the response for QueryTicketsByStatusRequest
type QueryTicketsByStatusResponse struct {
	Tickets []SupportTicket `json:"tickets"`
}

// QueryTicketResponsesRequest is the request for querying ticket responses
type QueryTicketResponsesRequest struct {
	TicketID string `json:"ticket_id"`
}

// QueryTicketResponsesResponse is the response for QueryTicketResponsesRequest
type QueryTicketResponsesResponse struct {
	Responses []TicketResponse `json:"responses"`
}

// QueryParamsRequest is the request for querying module parameters
type QueryParamsRequest struct{}

// QueryParamsResponse is the response for QueryParamsRequest
type QueryParamsResponse struct {
	Params Params `json:"params"`
}

// QueryOpenTicketsRequest is the request for querying all open tickets (admin only)
type QueryOpenTicketsRequest struct {
	Priority string `json:"priority,omitempty"` // Optional priority filter
}

// QueryOpenTicketsResponse is the response for QueryOpenTicketsRequest
type QueryOpenTicketsResponse struct {
	Tickets []SupportTicket `json:"tickets"`
}

// Proto message interface stubs

func (*QueryTicketRequest) ProtoMessage()              {}
func (m *QueryTicketRequest) Reset()                   { *m = QueryTicketRequest{} }
func (m *QueryTicketRequest) String() string           { return fmt.Sprintf("%+v", *m) }

func (*QueryTicketResponse) ProtoMessage()             {}
func (m *QueryTicketResponse) Reset()                  { *m = QueryTicketResponse{} }
func (m *QueryTicketResponse) String() string          { return fmt.Sprintf("%+v", *m) }

func (*QueryTicketsByCustomerRequest) ProtoMessage()   {}
func (m *QueryTicketsByCustomerRequest) Reset()        { *m = QueryTicketsByCustomerRequest{} }
func (m *QueryTicketsByCustomerRequest) String() string { return fmt.Sprintf("%+v", *m) }

func (*QueryTicketsByCustomerResponse) ProtoMessage()  {}
func (m *QueryTicketsByCustomerResponse) Reset()       { *m = QueryTicketsByCustomerResponse{} }
func (m *QueryTicketsByCustomerResponse) String() string { return fmt.Sprintf("%+v", *m) }

func (*QueryTicketsByProviderRequest) ProtoMessage()   {}
func (m *QueryTicketsByProviderRequest) Reset()        { *m = QueryTicketsByProviderRequest{} }
func (m *QueryTicketsByProviderRequest) String() string { return fmt.Sprintf("%+v", *m) }

func (*QueryTicketsByProviderResponse) ProtoMessage()  {}
func (m *QueryTicketsByProviderResponse) Reset()       { *m = QueryTicketsByProviderResponse{} }
func (m *QueryTicketsByProviderResponse) String() string { return fmt.Sprintf("%+v", *m) }

func (*QueryTicketsByAgentRequest) ProtoMessage()      {}
func (m *QueryTicketsByAgentRequest) Reset()           { *m = QueryTicketsByAgentRequest{} }
func (m *QueryTicketsByAgentRequest) String() string   { return fmt.Sprintf("%+v", *m) }

func (*QueryTicketsByAgentResponse) ProtoMessage()     {}
func (m *QueryTicketsByAgentResponse) Reset()          { *m = QueryTicketsByAgentResponse{} }
func (m *QueryTicketsByAgentResponse) String() string  { return fmt.Sprintf("%+v", *m) }

func (*QueryTicketsByStatusRequest) ProtoMessage()     {}
func (m *QueryTicketsByStatusRequest) Reset()          { *m = QueryTicketsByStatusRequest{} }
func (m *QueryTicketsByStatusRequest) String() string  { return fmt.Sprintf("%+v", *m) }

func (*QueryTicketsByStatusResponse) ProtoMessage()    {}
func (m *QueryTicketsByStatusResponse) Reset()         { *m = QueryTicketsByStatusResponse{} }
func (m *QueryTicketsByStatusResponse) String() string { return fmt.Sprintf("%+v", *m) }

func (*QueryTicketResponsesRequest) ProtoMessage()     {}
func (m *QueryTicketResponsesRequest) Reset()          { *m = QueryTicketResponsesRequest{} }
func (m *QueryTicketResponsesRequest) String() string  { return fmt.Sprintf("%+v", *m) }

func (*QueryTicketResponsesResponse) ProtoMessage()    {}
func (m *QueryTicketResponsesResponse) Reset()         { *m = QueryTicketResponsesResponse{} }
func (m *QueryTicketResponsesResponse) String() string { return fmt.Sprintf("%+v", *m) }

func (*QueryParamsRequest) ProtoMessage()              {}
func (m *QueryParamsRequest) Reset()                   { *m = QueryParamsRequest{} }
func (m *QueryParamsRequest) String() string           { return fmt.Sprintf("%+v", *m) }

func (*QueryParamsResponse) ProtoMessage()             {}
func (m *QueryParamsResponse) Reset()                  { *m = QueryParamsResponse{} }
func (m *QueryParamsResponse) String() string          { return fmt.Sprintf("%+v", *m) }

func (*QueryOpenTicketsRequest) ProtoMessage()         {}
func (m *QueryOpenTicketsRequest) Reset()              { *m = QueryOpenTicketsRequest{} }
func (m *QueryOpenTicketsRequest) String() string      { return fmt.Sprintf("%+v", *m) }

func (*QueryOpenTicketsResponse) ProtoMessage()        {}
func (m *QueryOpenTicketsResponse) Reset()             { *m = QueryOpenTicketsResponse{} }
func (m *QueryOpenTicketsResponse) String() string     { return fmt.Sprintf("%+v", *m) }
