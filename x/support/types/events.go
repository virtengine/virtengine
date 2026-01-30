package types

import "fmt"

// Event types for the support module
const (
	EventTypeTicketCreated    = "ticket_created"
	EventTypeTicketAssigned   = "ticket_assigned"
	EventTypeTicketResponded  = "ticket_responded"
	EventTypeTicketResolved   = "ticket_resolved"
	EventTypeTicketClosed     = "ticket_closed"
	EventTypeTicketReopened   = "ticket_reopened"
	EventTypeTicketEscalated  = "ticket_escalated"
	EventTypePriorityChanged  = "priority_changed"
)

// Event attribute keys
const (
	AttributeKeyTicketID       = "ticket_id"
	AttributeKeyCustomer       = "customer"
	AttributeKeyProvider       = "provider"
	AttributeKeyAgent          = "assigned_to"
	AttributeKeyStatus         = "status"
	AttributeKeyPreviousStatus = "previous_status"
	AttributeKeyPriority       = "priority"
	AttributeKeyCategory       = "category"
	AttributeKeyResponseIndex  = "response_index"
	AttributeKeyResolution     = "resolution"
	AttributeKeyClosedBy       = "closed_by"
	AttributeKeyAssignedBy     = "assigned_by"
	AttributeKeyBlockHeight    = "block_height"
	AttributeKeyTimestamp      = "timestamp"
)

// EventTicketCreated is emitted when a new support ticket is created
type EventTicketCreated struct {
	TicketID    string `json:"ticket_id"`
	Customer    string `json:"customer"`
	Provider    string `json:"provider,omitempty"`
	Category    string `json:"category"`
	Priority    string `json:"priority"`
	BlockHeight int64  `json:"block_height"`
	Timestamp   int64  `json:"timestamp"`
}

// EventTicketAssigned is emitted when a ticket is assigned to an agent
type EventTicketAssigned struct {
	TicketID    string `json:"ticket_id"`
	AssignedTo  string `json:"assigned_to"`
	AssignedBy  string `json:"assigned_by"`
	BlockHeight int64  `json:"block_height"`
	Timestamp   int64  `json:"timestamp"`
}

// EventTicketResponded is emitted when a response is added to a ticket
type EventTicketResponded struct {
	TicketID      string `json:"ticket_id"`
	Responder     string `json:"responder"`
	ResponseIndex uint32 `json:"response_index"`
	BlockHeight   int64  `json:"block_height"`
	Timestamp     int64  `json:"timestamp"`
}

// EventTicketResolved is emitted when a ticket is resolved
type EventTicketResolved struct {
	TicketID    string `json:"ticket_id"`
	ResolvedBy  string `json:"resolved_by"`
	Resolution  string `json:"resolution"`
	BlockHeight int64  `json:"block_height"`
	Timestamp   int64  `json:"timestamp"`
}

// EventTicketClosed is emitted when a ticket is closed
type EventTicketClosed struct {
	TicketID    string `json:"ticket_id"`
	ClosedBy    string `json:"closed_by"`
	Reason      string `json:"reason,omitempty"`
	BlockHeight int64  `json:"block_height"`
	Timestamp   int64  `json:"timestamp"`
}

// EventTicketReopened is emitted when a ticket is reopened
type EventTicketReopened struct {
	TicketID    string `json:"ticket_id"`
	ReopenedBy  string `json:"reopened_by"`
	Reason      string `json:"reason,omitempty"`
	BlockHeight int64  `json:"block_height"`
	Timestamp   int64  `json:"timestamp"`
}

// EventTicketEscalated is emitted when a ticket priority is escalated
type EventTicketEscalated struct {
	TicketID         string `json:"ticket_id"`
	PreviousPriority string `json:"previous_priority"`
	NewPriority      string `json:"new_priority"`
	EscalatedBy      string `json:"escalated_by"`
	Reason           string `json:"reason,omitempty"`
	BlockHeight      int64  `json:"block_height"`
	Timestamp        int64  `json:"timestamp"`
}

// Proto message interface stubs for Event types

func (*EventTicketCreated) ProtoMessage()    {}
func (m *EventTicketCreated) Reset()         { *m = EventTicketCreated{} }
func (m *EventTicketCreated) String() string { return fmt.Sprintf("%+v", *m) }

func (*EventTicketAssigned) ProtoMessage()    {}
func (m *EventTicketAssigned) Reset()         { *m = EventTicketAssigned{} }
func (m *EventTicketAssigned) String() string { return fmt.Sprintf("%+v", *m) }

func (*EventTicketResponded) ProtoMessage()    {}
func (m *EventTicketResponded) Reset()         { *m = EventTicketResponded{} }
func (m *EventTicketResponded) String() string { return fmt.Sprintf("%+v", *m) }

func (*EventTicketResolved) ProtoMessage()    {}
func (m *EventTicketResolved) Reset()         { *m = EventTicketResolved{} }
func (m *EventTicketResolved) String() string { return fmt.Sprintf("%+v", *m) }

func (*EventTicketClosed) ProtoMessage()    {}
func (m *EventTicketClosed) Reset()         { *m = EventTicketClosed{} }
func (m *EventTicketClosed) String() string { return fmt.Sprintf("%+v", *m) }

func (*EventTicketReopened) ProtoMessage()    {}
func (m *EventTicketReopened) Reset()         { *m = EventTicketReopened{} }
func (m *EventTicketReopened) String() string { return fmt.Sprintf("%+v", *m) }

func (*EventTicketEscalated) ProtoMessage()    {}
func (m *EventTicketEscalated) Reset()         { *m = EventTicketEscalated{} }
func (m *EventTicketEscalated) String() string { return fmt.Sprintf("%+v", *m) }
