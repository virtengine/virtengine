package support

import (
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/support/keeper"
	types "github.com/virtengine/virtengine/x/support/types" //nolint:staticcheck // deprecated types retained for compatibility
)

// InitGenesis initializes the support module's state from a genesis state.
// This simplified module only manages external ticket references.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data *types.GenesisState) {
	// Set module parameters
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	// Initialize external refs
	for _, ref := range data.ExternalRefs {
		refCopy := ref // Create copy to avoid pointer issues
		if err := k.RegisterExternalRef(ctx, &refCopy); err != nil {
			// Skip if ref already exists (may happen during re-init)
			if err != types.ErrRefAlreadyExists {
				panic(err)
			}
		}
	}

	// Initialize support requests
	maxSubmitterSeq := make(map[string]uint64)
	var maxTicketNumber uint64
	for _, req := range data.SupportRequests {
		reqCopy := req
		if err := k.CreateSupportRequest(ctx, &reqCopy); err != nil {
			if err != types.ErrInvalidSupportRequest {
				panic(err)
			}
		}
		if req.ID.Sequence > maxSubmitterSeq[req.ID.SubmitterAddress] {
			maxSubmitterSeq[req.ID.SubmitterAddress] = req.ID.Sequence
		}
		if parsed := parseTicketNumber(req.TicketNumber); parsed > maxTicketNumber {
			maxTicketNumber = parsed
		}
	}

	// Initialize support responses
	maxResponseSeq := make(map[string]uint64)
	for _, resp := range data.SupportResponses {
		respCopy := resp
		if err := k.AddSupportResponse(ctx, &respCopy); err != nil {
			if err != types.ErrInvalidSupportResponse {
				panic(err)
			}
		}
		reqID := respCopy.RequestID.String()
		if respCopy.ID.Sequence > maxResponseSeq[reqID] {
			maxResponseSeq[reqID] = respCopy.ID.Sequence
		}
	}

	// Set sequences
	for submitter, seq := range maxSubmitterSeq {
		addr, err := sdk.AccAddressFromBech32(submitter)
		if err != nil {
			continue
		}
		k.SetSupportRequestSequence(ctx, addr, seq)
	}
	if maxTicketNumber > 0 {
		k.SetTicketNumberSequence(ctx, maxTicketNumber)
	}
	for reqID, seq := range maxResponseSeq {
		k.SetSupportResponseSequence(ctx, reqID, seq)
	}

	k.SetEventSequence(ctx, data.EventSequence)
}

// ExportGenesis exports the support module's state to a genesis state.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	// Get params
	params := k.GetParams(ctx)

	// Get all external refs
	var refs []types.ExternalTicketRef
	k.WithExternalRefs(ctx, func(ref types.ExternalTicketRef) bool {
		refs = append(refs, ref)
		return false
	})

	var requests []types.SupportRequest
	k.WithSupportRequests(ctx, func(req types.SupportRequest) bool {
		requests = append(requests, req)
		return false
	})

	var responses []types.SupportResponse
	for _, req := range requests {
		responses = append(responses, k.GetSupportResponses(ctx, req.ID)...)
	}

	return &types.GenesisState{
		ExternalRefs:     refs,
		SupportRequests:  requests,
		SupportResponses: responses,
		EventSequence:    k.GetEventSequence(ctx),
		Params:           params,
	}
}

// ValidateGenesis validates the genesis state
func ValidateGenesis(data *types.GenesisState) error {
	return data.Validate()
}

func parseTicketNumber(ticketNumber string) uint64 {
	if ticketNumber == "" {
		return 0
	}
	if strings.HasPrefix(ticketNumber, "SUP-") {
		ticketNumber = strings.TrimPrefix(ticketNumber, "SUP-")
	}
	seq, err := strconv.ParseUint(ticketNumber, 10, 64)
	if err != nil {
		return 0
	}
	return seq
}
