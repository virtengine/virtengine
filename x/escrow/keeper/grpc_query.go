package keeper

import (
	"bytes"
	"context"
	"encoding/json"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkquery "github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	types "github.com/virtengine/virtengine/sdk/go/node/escrow/types/v1"
	"github.com/virtengine/virtengine/sdk/go/node/escrow/v1"

	"github.com/virtengine/virtengine/util/query"
	"github.com/virtengine/virtengine/x/escrow/types/billing"
)

// Querier is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper
type Querier struct {
	*keeper
}

var _ v1.QueryServer = Querier{}

func (k Querier) Accounts(c context.Context, req *v1.QueryAccountsRequest) (*v1.QueryAccountsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Pagination == nil {
		req.Pagination = &sdkquery.PageRequest{}
	}

	if req.Pagination.Limit == 0 {
		req.Pagination.Limit = sdkquery.DefaultLimit
	}

	states := make([]byte, 0, 3)

	var searchPrefix []byte

	// setup for case 3 - cross-index search
	// nolint: gocritic
	if len(req.Pagination.Key) > 0 {
		var key []byte
		var err error
		states, searchPrefix, key, _, err = query.DecodePaginationKey(req.Pagination.Key)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		req.Pagination.Key = key
	} else if req.State != "" {
		stateVal := types.State(types.State_value[req.State])

		if req.State != "" && stateVal == types.StateInvalid {
			return nil, status.Error(codes.InvalidArgument, "invalid state value")
		}

		states = append(states, byte(stateVal))
	} else {
		// request does not have a pagination set. Start from active store
		states = append(states, []byte{byte(types.StateOpen), byte(types.StateClosed), byte(types.StateOverdrawn)}...)
	}

	var accounts types.Accounts
	var pageRes *sdkquery.PageResponse

	total := uint64(0)

	for idx := range states {
		state := types.State(states[idx])

		var err error
		if idx > 0 {
			req.Pagination.Key = nil
		}

		if len(req.Pagination.Key) == 0 {
			req.State = state.String()

			searchPrefix = BuildSearchPrefix(AccountPrefix, req.State, req.XID)
		}

		searchStore := prefix.NewStore(ctx.KVStore(k.skey), searchPrefix)

		count := uint64(0)

		pageRes, err = sdkquery.FilteredPaginate(searchStore, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
			id, _ := ParseAccountKey(append(searchPrefix, key...))
			acc := types.Account{
				ID: id,
			}

			er := k.cdc.Unmarshal(value, &acc.State)
			if er != nil {
				return false, er
			}

			accounts = append(accounts, acc)
			count++

			return false, nil
		})
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		req.Pagination.Limit -= count
		total += count

		if req.Pagination.Limit == 0 {
			if len(pageRes.NextKey) > 0 {
				pageRes.NextKey, err = query.EncodePaginationKey(states[idx:], searchPrefix, pageRes.NextKey, nil)
				if err != nil {
					pageRes.Total = total
					return &v1.QueryAccountsResponse{
						Accounts:   accounts,
						Pagination: pageRes,
					}, status.Error(codes.Internal, err.Error())
				}
			}

			break
		}
	}

	pageRes.Total = total

	return &v1.QueryAccountsResponse{
		Accounts:   accounts,
		Pagination: pageRes,
	}, nil
}

func (k Querier) Payments(c context.Context, req *v1.QueryPaymentsRequest) (*v1.QueryPaymentsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Pagination == nil {
		req.Pagination = &sdkquery.PageRequest{}
	}

	if req.Pagination.Limit == 0 {
		req.Pagination.Limit = sdkquery.DefaultLimit
	}

	states := make([]byte, 0, 3)

	var searchPrefix []byte

	// setup for case 3 - cross-index search
	// nolint: gocritic
	if len(req.Pagination.Key) > 0 {
		var key []byte
		var err error
		states, searchPrefix, key, _, err = query.DecodePaginationKey(req.Pagination.Key)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		req.Pagination.Key = key
	} else if req.State != "" {
		stateVal := types.State(types.State_value[req.State])

		if req.State != "" && stateVal == types.StateInvalid {
			return nil, status.Error(codes.InvalidArgument, "invalid state value")
		}

		states = append(states, byte(stateVal))
	} else {
		// request does not have a pagination set. Start from active store
		states = append(states, []byte{byte(types.StateOpen), byte(types.StateClosed), byte(types.StateOverdrawn)}...)
	}

	var payments types.Payments
	var pageRes *sdkquery.PageResponse

	total := uint64(0)

	for idx := range states {
		state := types.State(states[idx])

		var err error
		if idx > 0 {
			req.Pagination.Key = nil
		}

		if len(req.Pagination.Key) == 0 {
			req.State = state.String()

			searchPrefix = BuildSearchPrefix(PaymentPrefix, req.State, req.XID)
		}

		searchStore := prefix.NewStore(ctx.KVStore(k.skey), searchPrefix)

		count := uint64(0)

		pageRes, err = sdkquery.FilteredPaginate(searchStore, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
			id, _ := ParsePaymentKey(append(searchPrefix, key...))
			pmnt := types.Payment{
				ID: id,
			}

			er := k.cdc.Unmarshal(value, &pmnt.State)
			if er != nil {
				return false, er
			}

			payments = append(payments, pmnt)
			count++

			return false, nil
		})
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		req.Pagination.Limit -= count
		total += count

		if req.Pagination.Limit == 0 {
			if len(pageRes.NextKey) > 0 {
				pageRes.NextKey, err = query.EncodePaginationKey(states[idx:], searchPrefix, pageRes.NextKey, nil)
				if err != nil {
					pageRes.Total = total
					return &v1.QueryPaymentsResponse{
						Payments:   payments,
						Pagination: pageRes,
					}, status.Error(codes.Internal, err.Error())
				}
			}

			break
		}
	}

	pageRes.Total = total

	return &v1.QueryPaymentsResponse{
		Payments:   payments,
		Pagination: pageRes,
	}, nil
}

func (k Querier) Invoice(c context.Context, req *v1.QueryInvoiceRequest) (*v1.QueryInvoiceResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	if req == nil || req.InvoiceId == "" {
		return nil, status.Error(codes.InvalidArgument, "invoice_id cannot be empty")
	}

	record, err := k.NewInvoiceKeeper().GetInvoice(ctx, req.InvoiceId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	bz, err := json.Marshal(record)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &v1.QueryInvoiceResponse{InvoiceJson: string(bz)}, nil
}

func (k Querier) InvoicesByProvider(c context.Context, req *v1.QueryInvoicesByProviderRequest) (*v1.QueryInvoicesByProviderResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	if req == nil || req.Provider == "" {
		return nil, status.Error(codes.InvalidArgument, "provider cannot be empty")
	}

	records, pageRes, err := k.NewInvoiceKeeper().GetInvoicesByProvider(ctx, req.Provider, req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	response := &v1.QueryInvoicesByProviderResponse{
		InvoicesJson: make([]string, 0, len(records)),
		Pagination:   pageRes,
	}

	for _, record := range records {
		bz, err := json.Marshal(record)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		response.InvoicesJson = append(response.InvoicesJson, string(bz))
	}

	return response, nil
}

func (k Querier) InvoicesByCustomer(c context.Context, req *v1.QueryInvoicesByCustomerRequest) (*v1.QueryInvoicesByCustomerResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	if req == nil || req.Customer == "" {
		return nil, status.Error(codes.InvalidArgument, "customer cannot be empty")
	}

	records, pageRes, err := k.NewInvoiceKeeper().GetInvoicesByCustomer(ctx, req.Customer, req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	response := &v1.QueryInvoicesByCustomerResponse{
		InvoicesJson: make([]string, 0, len(records)),
		Pagination:   pageRes,
	}

	for _, record := range records {
		bz, err := json.Marshal(record)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		response.InvoicesJson = append(response.InvoicesJson, string(bz))
	}

	return response, nil
}

func (k Querier) InvoiceLedger(c context.Context, req *v1.QueryInvoiceLedgerRequest) (*v1.QueryInvoiceLedgerResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	if req == nil || req.InvoiceId == "" {
		return nil, status.Error(codes.InvalidArgument, "invoice_id cannot be empty")
	}

	entries, err := k.NewInvoiceKeeper().GetInvoiceLedgerEntries(ctx, req.InvoiceId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	response := &v1.QueryInvoiceLedgerResponse{
		EntriesJson: make([]string, 0, len(entries)),
	}

	for _, entry := range entries {
		bz, err := json.Marshal(entry)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		response.EntriesJson = append(response.EntriesJson, string(bz))
	}

	return response, nil
}

// Ensure billing types are referenced for gogo/proto codegen.
var _ = billing.Invoice{}

func BuildSearchPrefix(prefix []byte, state string, xid string) []byte {
	buf := &bytes.Buffer{}

	buf.Write(prefix)
	if state != "" {
		st := types.State(types.State_value[state])
		buf.Write(stateToPrefix(st))
		if xid != "" {
			buf.WriteRune('/')
			buf.WriteString(xid)
		}
	}

	return buf.Bytes()
}
