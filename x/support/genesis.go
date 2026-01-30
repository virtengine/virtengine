package support

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/support/keeper"
	"github.com/virtengine/virtengine/x/support/types"
)

// InitGenesis initializes the support module's state from a genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data *types.GenesisState) {
	// Set module parameters
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	// Set ticket sequence
	k.SetTicketSequence(ctx, data.TicketSequence)

	// Initialize tickets
	for _, ticket := range data.Tickets {
		ticketCopy := ticket // Create copy to avoid pointer issues
		if err := k.CreateTicket(ctx, &ticketCopy); err != nil {
			// Skip if ticket already exists (may happen during re-init)
			if err != types.ErrTicketAlreadyExists {
				panic(err)
			}
		}
	}

	// Initialize responses
	for _, response := range data.Responses {
		responseCopy := response // Create copy to avoid pointer issues
		if err := k.AddResponse(ctx, response.TicketID, &responseCopy); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis exports the support module's state to a genesis state.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	// Get params
	params := k.GetParams(ctx)

	// Get ticket sequence
	ticketSequence := k.GetTicketSequence(ctx)

	// Get all tickets
	var tickets []types.SupportTicket
	k.WithTickets(ctx, func(ticket types.SupportTicket) bool {
		tickets = append(tickets, ticket)
		return false
	})

	// Get all responses
	var responses []types.TicketResponse
	for _, ticket := range tickets {
		ticketResponses := k.GetResponses(ctx, ticket.TicketID)
		responses = append(responses, ticketResponses...)
	}

	return &types.GenesisState{
		Tickets:        tickets,
		Responses:      responses,
		Params:         params,
		TicketSequence: ticketSequence,
	}
}

// ValidateGenesis validates the genesis state
func ValidateGenesis(data *types.GenesisState) error {
	return data.Validate()
}
