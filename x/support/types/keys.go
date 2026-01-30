package types

import (
	"encoding/binary"
)

const (
	// ModuleName is the module name constant used in many places
	ModuleName = "support"

	// StoreKey is the store key string for support module
	StoreKey = ModuleName

	// RouterKey is the message route for support module
	RouterKey = ModuleName

	// QuerierRoute is the querier route for support module
	QuerierRoute = ModuleName
)

// Store key prefixes
var (
	// PrefixTicket is the prefix for ticket storage
	// Key: PrefixTicket | ticket_id -> SupportTicket
	PrefixTicket = []byte{0x01}

	// PrefixTicketsByCustomer is the prefix for customer ticket index
	// Key: PrefixTicketsByCustomer | customer_address | ticket_id -> bool
	PrefixTicketsByCustomer = []byte{0x02}

	// PrefixTicketsByProvider is the prefix for provider ticket index
	// Key: PrefixTicketsByProvider | provider_address | ticket_id -> bool
	PrefixTicketsByProvider = []byte{0x03}

	// PrefixTicketsByAgent is the prefix for assigned agent ticket index
	// Key: PrefixTicketsByAgent | agent_address | ticket_id -> bool
	PrefixTicketsByAgent = []byte{0x04}

	// PrefixTicketsByStatus is the prefix for status-based ticket index
	// Key: PrefixTicketsByStatus | status | ticket_id -> bool
	PrefixTicketsByStatus = []byte{0x05}

	// PrefixTicketResponse is the prefix for ticket response storage
	// Key: PrefixTicketResponse | ticket_id | response_index -> TicketResponse
	PrefixTicketResponse = []byte{0x06}

	// PrefixParams is the prefix for module parameters
	PrefixParams = []byte{0x07}

	// PrefixRateLimit is the prefix for rate limit tracking
	// Key: PrefixRateLimit | address | day_timestamp -> count
	PrefixRateLimit = []byte{0x08}

	// PrefixTicketSequence is the prefix for ticket ID sequence
	PrefixTicketSequence = []byte{0x09}
)

// TicketKey returns the store key for a ticket
func TicketKey(ticketID string) []byte {
	key := make([]byte, 0, len(PrefixTicket)+len(ticketID))
	key = append(key, PrefixTicket...)
	key = append(key, []byte(ticketID)...)
	return key
}

// TicketsByCustomerKey returns the store key for customer ticket index
func TicketsByCustomerKey(customerAddr []byte, ticketID string) []byte {
	key := make([]byte, 0, len(PrefixTicketsByCustomer)+len(customerAddr)+len(ticketID)+1)
	key = append(key, PrefixTicketsByCustomer...)
	key = append(key, customerAddr...)
	key = append(key, '/')
	key = append(key, []byte(ticketID)...)
	return key
}

// TicketsByCustomerPrefixKey returns the prefix for a customer's tickets
func TicketsByCustomerPrefixKey(customerAddr []byte) []byte {
	key := make([]byte, 0, len(PrefixTicketsByCustomer)+len(customerAddr)+1)
	key = append(key, PrefixTicketsByCustomer...)
	key = append(key, customerAddr...)
	key = append(key, '/')
	return key
}

// TicketsByProviderKey returns the store key for provider ticket index
func TicketsByProviderKey(providerAddr []byte, ticketID string) []byte {
	key := make([]byte, 0, len(PrefixTicketsByProvider)+len(providerAddr)+len(ticketID)+1)
	key = append(key, PrefixTicketsByProvider...)
	key = append(key, providerAddr...)
	key = append(key, '/')
	key = append(key, []byte(ticketID)...)
	return key
}

// TicketsByProviderPrefixKey returns the prefix for a provider's tickets
func TicketsByProviderPrefixKey(providerAddr []byte) []byte {
	key := make([]byte, 0, len(PrefixTicketsByProvider)+len(providerAddr)+1)
	key = append(key, PrefixTicketsByProvider...)
	key = append(key, providerAddr...)
	key = append(key, '/')
	return key
}

// TicketsByAgentKey returns the store key for agent ticket index
func TicketsByAgentKey(agentAddr []byte, ticketID string) []byte {
	key := make([]byte, 0, len(PrefixTicketsByAgent)+len(agentAddr)+len(ticketID)+1)
	key = append(key, PrefixTicketsByAgent...)
	key = append(key, agentAddr...)
	key = append(key, '/')
	key = append(key, []byte(ticketID)...)
	return key
}

// TicketsByAgentPrefixKey returns the prefix for an agent's tickets
func TicketsByAgentPrefixKey(agentAddr []byte) []byte {
	key := make([]byte, 0, len(PrefixTicketsByAgent)+len(agentAddr)+1)
	key = append(key, PrefixTicketsByAgent...)
	key = append(key, agentAddr...)
	key = append(key, '/')
	return key
}

// TicketsByStatusKey returns the store key for status ticket index
func TicketsByStatusKey(status TicketStatus, ticketID string) []byte {
	key := make([]byte, 0, len(PrefixTicketsByStatus)+2+len(ticketID))
	key = append(key, PrefixTicketsByStatus...)
	key = append(key, byte(status))
	key = append(key, '/')
	key = append(key, []byte(ticketID)...)
	return key
}

// TicketsByStatusPrefixKey returns the prefix for tickets with a specific status
func TicketsByStatusPrefixKey(status TicketStatus) []byte {
	key := make([]byte, 0, len(PrefixTicketsByStatus)+2)
	key = append(key, PrefixTicketsByStatus...)
	key = append(key, byte(status))
	key = append(key, '/')
	return key
}

// TicketResponseKey returns the store key for a ticket response
func TicketResponseKey(ticketID string, responseIndex uint32) []byte {
	key := make([]byte, 0, len(PrefixTicketResponse)+len(ticketID)+5)
	key = append(key, PrefixTicketResponse...)
	key = append(key, []byte(ticketID)...)
	key = append(key, '/')
	indexBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(indexBytes, responseIndex)
	key = append(key, indexBytes...)
	return key
}

// TicketResponsePrefixKey returns the prefix for a ticket's responses
func TicketResponsePrefixKey(ticketID string) []byte {
	key := make([]byte, 0, len(PrefixTicketResponse)+len(ticketID)+1)
	key = append(key, PrefixTicketResponse...)
	key = append(key, []byte(ticketID)...)
	key = append(key, '/')
	return key
}

// ParamsKey returns the store key for module parameters
func ParamsKey() []byte {
	return PrefixParams
}

// RateLimitKey returns the store key for rate limit tracking
func RateLimitKey(addr []byte, dayTimestamp int64) []byte {
	key := make([]byte, 0, len(PrefixRateLimit)+len(addr)+9)
	key = append(key, PrefixRateLimit...)
	key = append(key, addr...)
	key = append(key, '/')
	tsBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(tsBytes, uint64(dayTimestamp))
	key = append(key, tsBytes...)
	return key
}

// TicketSequenceKey returns the store key for ticket ID sequence
func TicketSequenceKey() []byte {
	return PrefixTicketSequence
}
