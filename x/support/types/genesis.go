package types

import "fmt"

// GenesisState is the genesis state for the support module
type GenesisState struct {
	// Tickets are the initial support tickets
	Tickets []SupportTicket `json:"tickets"`

	// Responses are the initial ticket responses
	Responses []TicketResponse `json:"responses"`

	// Params are the module parameters
	Params Params `json:"params"`

	// TicketSequence is the next ticket ID sequence number
	TicketSequence uint64 `json:"ticket_sequence"`
}

// Params defines the parameters for the support module
type Params struct {
	// MaxTicketsPerCustomerPerDay is the maximum tickets a customer can create per day
	MaxTicketsPerCustomerPerDay uint32 `json:"max_tickets_per_customer_per_day"`

	// MaxResponsesPerTicket is the maximum responses allowed per ticket
	MaxResponsesPerTicket uint32 `json:"max_responses_per_ticket"`

	// TicketCooldownSeconds is the minimum seconds between ticket creation
	TicketCooldownSeconds uint32 `json:"ticket_cooldown_seconds"`

	// AutoCloseAfterDays is the number of days after resolution before auto-close
	AutoCloseAfterDays uint32 `json:"auto_close_after_days"`

	// MaxOpenTicketsPerCustomer is the maximum open tickets per customer
	MaxOpenTicketsPerCustomer uint32 `json:"max_open_tickets_per_customer"`

	// ReopenWindowDays is the number of days after close that ticket can be reopened
	ReopenWindowDays uint32 `json:"reopen_window_days"`

	// AllowedCategories is the list of allowed ticket categories
	AllowedCategories []string `json:"allowed_categories"`
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Tickets:        []SupportTicket{},
		Responses:      []TicketResponse{},
		Params:         DefaultParams(),
		TicketSequence: 1,
	}
}

// DefaultParams returns the default parameters
func DefaultParams() Params {
	return Params{
		MaxTicketsPerCustomerPerDay: 5,
		MaxResponsesPerTicket:       100,
		TicketCooldownSeconds:       60,
		AutoCloseAfterDays:          7,
		MaxOpenTicketsPerCustomer:   10,
		ReopenWindowDays:            30,
		AllowedCategories: []string{
			"billing",
			"technical",
			"account",
			"provider",
			"deployment",
			"security",
			"other",
		},
	}
}

// Validate validates the genesis state
func (gs GenesisState) Validate() error {
	// Validate tickets
	seenTickets := make(map[string]bool)
	for _, ticket := range gs.Tickets {
		if err := ticket.Validate(); err != nil {
			return err
		}
		if seenTickets[ticket.TicketID] {
			return ErrTicketAlreadyExists.Wrapf("duplicate ticket: %s", ticket.TicketID)
		}
		seenTickets[ticket.TicketID] = true
	}

	// Validate responses
	for _, response := range gs.Responses {
		if err := response.Validate(); err != nil {
			return err
		}
		// Ensure parent ticket exists
		if !seenTickets[response.TicketID] {
			return ErrTicketNotFound.Wrapf("response references non-existent ticket: %s", response.TicketID)
		}
	}

	// Validate params
	return gs.Params.Validate()
}

// Validate validates the params
func (p Params) Validate() error {
	if p.MaxTicketsPerCustomerPerDay == 0 {
		return ErrInvalidParams.Wrap("max_tickets_per_customer_per_day must be greater than 0")
	}

	if p.MaxResponsesPerTicket == 0 {
		return ErrInvalidParams.Wrap("max_responses_per_ticket must be greater than 0")
	}

	if p.MaxOpenTicketsPerCustomer == 0 {
		return ErrInvalidParams.Wrap("max_open_tickets_per_customer must be greater than 0")
	}

	if len(p.AllowedCategories) == 0 {
		return ErrInvalidParams.Wrap("allowed_categories cannot be empty")
	}

	return nil
}

// IsCategoryAllowed checks if a category is in the allowed list
func (p Params) IsCategoryAllowed(category string) bool {
	for _, c := range p.AllowedCategories {
		if c == category {
			return true
		}
	}
	return false
}

// Proto message interface stubs for GenesisState
func (*GenesisState) ProtoMessage() {}
func (gs *GenesisState) Reset()     { *gs = GenesisState{} }
func (gs *GenesisState) String() string {
	return fmt.Sprintf("GenesisState{Tickets: %d, Responses: %d, Sequence: %d}",
		len(gs.Tickets), len(gs.Responses), gs.TicketSequence)
}

// Proto message interface stubs for Params
func (*Params) ProtoMessage() {}
func (p *Params) Reset()      { *p = Params{} }
func (p *Params) String() string {
	return fmt.Sprintf("Params{MaxTicketsPerDay: %d, MaxResponses: %d}",
		p.MaxTicketsPerCustomerPerDay, p.MaxResponsesPerTicket)
}
