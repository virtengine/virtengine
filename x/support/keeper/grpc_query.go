package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/virtengine/virtengine/x/support/types"
)

// GRPCQuerier implements the gRPC query interface with proper context handling
type GRPCQuerier struct {
	Keeper
}

var _ types.QueryServer = GRPCQuerier{}

// Ticket returns a single ticket by ID
func (q GRPCQuerier) Ticket(c context.Context, req *types.QueryTicketRequest) (*types.QueryTicketResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.TicketID == "" {
		return nil, status.Error(codes.InvalidArgument, "ticket ID is required")
	}

	ctx := sdk.UnwrapSDKContext(c)

	ticket, found := q.Keeper.GetTicket(ctx, req.TicketID)
	if !found {
		return nil, types.ErrTicketNotFound.Wrapf("ticket %s not found", req.TicketID)
	}

	// Note: Access control should be enforced at the client/gateway level
	// or by checking the caller's identity in a more sophisticated setup.
	// For on-chain queries, we return the ticket but encrypted payloads
	// can only be decrypted by authorized recipients.

	return &types.QueryTicketResponse{
		Ticket: ticket,
	}, nil
}

// TicketsByCustomer returns all tickets for a customer
func (q GRPCQuerier) TicketsByCustomer(c context.Context, req *types.QueryTicketsByCustomerRequest) (*types.QueryTicketsByCustomerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.CustomerAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "customer address is required")
	}

	ctx := sdk.UnwrapSDKContext(c)

	customerAddr, err := sdk.AccAddressFromBech32(req.CustomerAddress)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(err.Error())
	}

	tickets := q.Keeper.GetTicketsByCustomer(ctx, customerAddr)

	// Filter by status if provided
	if req.Status != "" {
		statusFilter, err := types.TicketStatusFromString(req.Status)
		if err != nil {
			return nil, types.ErrInvalidTicketStatus.Wrap(err.Error())
		}

		var filteredTickets []types.SupportTicket
		for _, t := range tickets {
			if t.Status == statusFilter {
				filteredTickets = append(filteredTickets, t)
			}
		}
		tickets = filteredTickets
	}

	return &types.QueryTicketsByCustomerResponse{
		Tickets: tickets,
	}, nil
}

// TicketsByProvider returns all tickets related to a provider
func (q GRPCQuerier) TicketsByProvider(c context.Context, req *types.QueryTicketsByProviderRequest) (*types.QueryTicketsByProviderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.ProviderAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "provider address is required")
	}

	ctx := sdk.UnwrapSDKContext(c)

	providerAddr, err := sdk.AccAddressFromBech32(req.ProviderAddress)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(err.Error())
	}

	tickets := q.Keeper.GetTicketsByProvider(ctx, providerAddr)

	// Filter by status if provided
	if req.Status != "" {
		statusFilter, err := types.TicketStatusFromString(req.Status)
		if err != nil {
			return nil, types.ErrInvalidTicketStatus.Wrap(err.Error())
		}

		var filteredTickets []types.SupportTicket
		for _, t := range tickets {
			if t.Status == statusFilter {
				filteredTickets = append(filteredTickets, t)
			}
		}
		tickets = filteredTickets
	}

	return &types.QueryTicketsByProviderResponse{
		Tickets: tickets,
	}, nil
}

// TicketsByAgent returns all tickets assigned to an agent
func (q GRPCQuerier) TicketsByAgent(c context.Context, req *types.QueryTicketsByAgentRequest) (*types.QueryTicketsByAgentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.AgentAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "agent address is required")
	}

	ctx := sdk.UnwrapSDKContext(c)

	agentAddr, err := sdk.AccAddressFromBech32(req.AgentAddress)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(err.Error())
	}

	tickets := q.Keeper.GetTicketsByAgent(ctx, agentAddr)

	// Filter by status if provided
	if req.Status != "" {
		statusFilter, err := types.TicketStatusFromString(req.Status)
		if err != nil {
			return nil, types.ErrInvalidTicketStatus.Wrap(err.Error())
		}

		var filteredTickets []types.SupportTicket
		for _, t := range tickets {
			if t.Status == statusFilter {
				filteredTickets = append(filteredTickets, t)
			}
		}
		tickets = filteredTickets
	}

	return &types.QueryTicketsByAgentResponse{
		Tickets: tickets,
	}, nil
}

// TicketsByStatus returns all tickets with a specific status
func (q GRPCQuerier) TicketsByStatus(c context.Context, req *types.QueryTicketsByStatusRequest) (*types.QueryTicketsByStatusResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Status == "" {
		return nil, status.Error(codes.InvalidArgument, "status is required")
	}

	ctx := sdk.UnwrapSDKContext(c)

	ticketStatus, err := types.TicketStatusFromString(req.Status)
	if err != nil {
		return nil, types.ErrInvalidTicketStatus.Wrap(err.Error())
	}

	tickets := q.Keeper.GetTicketsByStatus(ctx, ticketStatus)

	return &types.QueryTicketsByStatusResponse{
		Tickets: tickets,
	}, nil
}

// TicketResponses returns all responses for a ticket
func (q GRPCQuerier) TicketResponses(c context.Context, req *types.QueryTicketResponsesRequest) (*types.QueryTicketResponsesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.TicketID == "" {
		return nil, status.Error(codes.InvalidArgument, "ticket ID is required")
	}

	ctx := sdk.UnwrapSDKContext(c)

	// Verify ticket exists
	_, found := q.Keeper.GetTicket(ctx, req.TicketID)
	if !found {
		return nil, types.ErrTicketNotFound.Wrapf("ticket %s not found", req.TicketID)
	}

	responses := q.Keeper.GetResponses(ctx, req.TicketID)

	return &types.QueryTicketResponsesResponse{
		Responses: responses,
	}, nil
}

// OpenTickets returns all open tickets (admin/agent access)
func (q GRPCQuerier) OpenTickets(c context.Context, req *types.QueryOpenTicketsRequest) (*types.QueryOpenTicketsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	// Get tickets with Open status
	tickets := q.Keeper.GetTicketsByStatus(ctx, types.TicketStatusOpen)

	// Filter by priority if provided
	if req.Priority != "" {
		priorityFilter, err := types.TicketPriorityFromString(req.Priority)
		if err != nil {
			return nil, types.ErrInvalidTicketPriority.Wrap(err.Error())
		}

		var filteredTickets []types.SupportTicket
		for _, t := range tickets {
			if t.Priority == priorityFilter {
				filteredTickets = append(filteredTickets, t)
			}
		}
		tickets = filteredTickets
	}

	return &types.QueryOpenTicketsResponse{
		Tickets: tickets,
	}, nil
}

// Params returns the module parameters
func (q GRPCQuerier) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	params := q.Keeper.GetParams(ctx)

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}
