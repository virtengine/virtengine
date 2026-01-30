package support

import (
	"github.com/virtengine/virtengine/x/support/keeper"
	"github.com/virtengine/virtengine/x/support/types"
)

const (
	// ModuleName is the module name
	ModuleName = types.ModuleName

	// StoreKey is the store key
	StoreKey = types.StoreKey

	// RouterKey is the router key
	RouterKey = types.RouterKey

	// QuerierRoute is the querier route
	QuerierRoute = types.QuerierRoute
)

var (
	// DefaultGenesisState returns the default genesis state
	DefaultGenesisState = types.DefaultGenesisState

	// NewKeeper creates a new keeper
	NewKeeper = keeper.NewKeeper

	// NewMsgServerImpl creates a new message server
	NewMsgServerImpl = keeper.NewMsgServerImpl
)

// Type aliases for types package
type (
	// GenesisState is the genesis state for the support module
	GenesisState = types.GenesisState

	// Params defines the module parameters
	Params = types.Params

	// SupportTicket represents a support ticket
	SupportTicket = types.SupportTicket

	// TicketResponse represents a ticket response
	TicketResponse = types.TicketResponse

	// TicketStatus represents ticket status
	TicketStatus = types.TicketStatus

	// TicketPriority represents ticket priority
	TicketPriority = types.TicketPriority

	// ResourceReference represents a resource reference
	ResourceReference = types.ResourceReference

	// MsgCreateTicket creates a new ticket
	MsgCreateTicket = types.MsgCreateTicket

	// MsgAssignTicket assigns a ticket
	MsgAssignTicket = types.MsgAssignTicket

	// MsgRespondToTicket responds to a ticket
	MsgRespondToTicket = types.MsgRespondToTicket

	// MsgResolveTicket resolves a ticket
	MsgResolveTicket = types.MsgResolveTicket

	// MsgCloseTicket closes a ticket
	MsgCloseTicket = types.MsgCloseTicket

	// MsgReopenTicket reopens a ticket
	MsgReopenTicket = types.MsgReopenTicket

	// MsgUpdateParams updates params
	MsgUpdateParams = types.MsgUpdateParams

	// Keeper is the support keeper
	Keeper = keeper.Keeper
)

// Ticket status constants
const (
	TicketStatusOpen            = types.TicketStatusOpen
	TicketStatusAssigned        = types.TicketStatusAssigned
	TicketStatusInProgress      = types.TicketStatusInProgress
	TicketStatusPendingCustomer = types.TicketStatusPendingCustomer
	TicketStatusResolved        = types.TicketStatusResolved
	TicketStatusClosed          = types.TicketStatusClosed
)

// Ticket priority constants
const (
	TicketPriorityLow    = types.TicketPriorityLow
	TicketPriorityNormal = types.TicketPriorityNormal
	TicketPriorityHigh   = types.TicketPriorityHigh
	TicketPriorityUrgent = types.TicketPriorityUrgent
)
