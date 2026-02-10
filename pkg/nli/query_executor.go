package nli

import (
	"context"
)

// ============================================================================
// Query Executor
// ============================================================================

// DefaultQueryExecutor executes blockchain queries based on intent
type DefaultQueryExecutor struct {
	// balanceQuerier queries account balances
	balanceQuerier BalanceQuerier

	// marketQuerier queries marketplace data
	marketQuerier MarketQuerier

	// identityQuerier queries VEID data
	identityQuerier IdentityQuerier

	// orderQuerier queries order/lease data
	orderQuerier OrderQuerier
}

// BalanceQuerier interface for balance queries
type BalanceQuerier interface {
	GetBalances(ctx context.Context, address string) ([]BalanceInfo, error)
}

// MarketQuerier interface for marketplace queries
type MarketQuerier interface {
	SearchOfferings(ctx context.Context, filters map[string]string) ([]OfferingInfo, error)
	GetProviderInfo(ctx context.Context, address string) (map[string]interface{}, error)
}

// IdentityQuerier interface for identity queries
type IdentityQuerier interface {
	GetIdentityScore(ctx context.Context, address string) (float32, error)
	IsVerified(ctx context.Context, address string) (bool, error)
}

// OrderQuerier interface for order queries
type OrderQuerier interface {
	GetOrders(ctx context.Context, address string) ([]OrderInfo, error)
	GetOrder(ctx context.Context, orderID string) (*OrderInfo, error)
}

// NewDefaultQueryExecutor creates a new query executor
func NewDefaultQueryExecutor() *DefaultQueryExecutor {
	return &DefaultQueryExecutor{}
}

// SetBalanceQuerier sets the balance querier
func (e *DefaultQueryExecutor) SetBalanceQuerier(q BalanceQuerier) {
	e.balanceQuerier = q
}

// SetMarketQuerier sets the market querier
func (e *DefaultQueryExecutor) SetMarketQuerier(q MarketQuerier) {
	e.marketQuerier = q
}

// SetIdentityQuerier sets the identity querier
func (e *DefaultQueryExecutor) SetIdentityQuerier(q IdentityQuerier) {
	e.identityQuerier = q
}

// SetOrderQuerier sets the order querier
func (e *DefaultQueryExecutor) SetOrderQuerier(q OrderQuerier) {
	e.orderQuerier = q
}

// Execute executes a query based on intent and entities
func (e *DefaultQueryExecutor) Execute(ctx context.Context, intent Intent, entities map[string]string, userAddress string) (*QueryResult, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	switch intent {
	case IntentQueryBalance:
		return e.executeBalanceQuery(ctx, userAddress)
	case IntentFindOfferings:
		return e.executeOfferingsQuery(ctx, entities)
	case IntentCheckOrder:
		return e.executeOrderQuery(ctx, entities, userAddress)
	case IntentGetProviderInfo:
		return e.executeProviderQuery(ctx, entities)
	case IntentIdentity:
		return e.executeIdentityQuery(ctx, userAddress)
	default:
		// No query execution needed for other intents
		return &QueryResult{
			Success:   true,
			QueryType: intent.String(),
		}, nil
	}
}

// executeBalanceQuery executes a balance query
func (e *DefaultQueryExecutor) executeBalanceQuery(ctx context.Context, address string) (*QueryResult, error) {
	if address == "" {
		return &QueryResult{
			Success:   false,
			Error:     "No wallet address provided",
			QueryType: "balance",
		}, nil
	}

	if e.balanceQuerier == nil {
		// Return mock data when querier is not configured
		return &QueryResult{
			Success:   true,
			QueryType: "balance",
			Data: []BalanceInfo{
				{Denom: "uve", Amount: "1000000", USD: "100.00"},
				{Denom: "uve", Amount: "500000", USD: "50.00"},
			},
		}, nil
	}

	balances, err := e.balanceQuerier.GetBalances(ctx, address)
	if err != nil {
		return &QueryResult{
			Success:   false,
			Error:     err.Error(),
			QueryType: "balance",
		}, nil
	}

	return &QueryResult{
		Success:   true,
		Data:      balances,
		QueryType: "balance",
	}, nil
}

// executeOfferingsQuery executes a marketplace offerings query
func (e *DefaultQueryExecutor) executeOfferingsQuery(ctx context.Context, entities map[string]string) (*QueryResult, error) {
	if e.marketQuerier == nil {
		// Return mock data when querier is not configured
		return &QueryResult{
			Success:   true,
			QueryType: "offerings",
			Data: []OfferingInfo{
				{
					ID:        "offering-1",
					Provider:  "ve1abc...xyz",
					Type:      "GPU Compute",
					Specs:     map[string]string{"gpu": "NVIDIA A100", "memory": "80GB"},
					Price:     "10 UVE/hour",
					Available: true,
				},
				{
					ID:        "offering-2",
					Provider:  "ve1def...uvw",
					Type:      "CPU Compute",
					Specs:     map[string]string{"cpu": "32 cores", "memory": "128GB"},
					Price:     "2 UVE/hour",
					Available: true,
				},
			},
		}, nil
	}

	offerings, err := e.marketQuerier.SearchOfferings(ctx, entities)
	if err != nil {
		return &QueryResult{
			Success:   false,
			Error:     err.Error(),
			QueryType: "offerings",
		}, nil
	}

	return &QueryResult{
		Success:   true,
		Data:      offerings,
		QueryType: "offerings",
	}, nil
}

// executeOrderQuery executes an order status query
func (e *DefaultQueryExecutor) executeOrderQuery(ctx context.Context, entities map[string]string, address string) (*QueryResult, error) {
	orderID := entities["order_id"]

	if e.orderQuerier == nil {
		// Return mock data when querier is not configured
		if orderID != "" {
			return &QueryResult{
				Success:   true,
				QueryType: "order",
				Data: OrderInfo{
					ID:       orderID,
					Status:   "active",
					Provider: "ve1abc...xyz",
				},
			}, nil
		}
		return &QueryResult{
			Success:   true,
			QueryType: "orders",
			Data: []OrderInfo{
				{ID: "order-1", Status: "active", Provider: "ve1abc...xyz"},
				{ID: "order-2", Status: "matched", Provider: "ve1def...uvw"},
			},
		}, nil
	}

	if orderID != "" {
		order, err := e.orderQuerier.GetOrder(ctx, orderID)
		if err != nil {
			return &QueryResult{
				Success:   false,
				Error:     err.Error(),
				QueryType: "order",
			}, nil
		}
		return &QueryResult{
			Success:   true,
			Data:      order,
			QueryType: "order",
		}, nil
	}

	if address == "" {
		return &QueryResult{
			Success:   false,
			Error:     "No wallet address or order ID provided",
			QueryType: "orders",
		}, nil
	}

	orders, err := e.orderQuerier.GetOrders(ctx, address)
	if err != nil {
		return &QueryResult{
			Success:   false,
			Error:     err.Error(),
			QueryType: "orders",
		}, nil
	}

	return &QueryResult{
		Success:   true,
		Data:      orders,
		QueryType: "orders",
	}, nil
}

// executeProviderQuery executes a provider information query
func (e *DefaultQueryExecutor) executeProviderQuery(ctx context.Context, entities map[string]string) (*QueryResult, error) {
	address := entities["address"]
	if address == "" {
		return &QueryResult{
			Success:   false,
			Error:     "No provider address specified",
			QueryType: "provider",
		}, nil
	}

	if e.marketQuerier == nil {
		// Return mock data when querier is not configured
		return &QueryResult{
			Success:   true,
			QueryType: "provider",
			Data: map[string]interface{}{
				"address":     address,
				"name":        "Example Provider",
				"reputation":  0.95,
				"uptime":      0.999,
				"offerings":   5,
				"totalOrders": 150,
			},
		}, nil
	}

	info, err := e.marketQuerier.GetProviderInfo(ctx, address)
	if err != nil {
		return &QueryResult{
			Success:   false,
			Error:     err.Error(),
			QueryType: "provider",
		}, nil
	}

	return &QueryResult{
		Success:   true,
		Data:      info,
		QueryType: "provider",
	}, nil
}

// executeIdentityQuery executes an identity/VEID query
func (e *DefaultQueryExecutor) executeIdentityQuery(ctx context.Context, address string) (*QueryResult, error) {
	if address == "" {
		return &QueryResult{
			Success:   false,
			Error:     "No wallet address provided",
			QueryType: "identity",
		}, nil
	}

	if e.identityQuerier == nil {
		// Return mock data when querier is not configured
		return &QueryResult{
			Success:   true,
			QueryType: "identity",
			Data: map[string]interface{}{
				"address":    address,
				"verified":   true,
				"score":      0.85,
				"level":      "full",
				"scopes":     []string{"document", "selfie", "liveness"},
				"verifiedAt": "2026-01-15T10:00:00Z",
			},
		}, nil
	}

	verified, err := e.identityQuerier.IsVerified(ctx, address)
	if err != nil {
		return &QueryResult{
			Success:   false,
			Error:     err.Error(),
			QueryType: "identity",
		}, nil
	}

	score, _ := e.identityQuerier.GetIdentityScore(ctx, address)

	return &QueryResult{
		Success:   true,
		QueryType: "identity",
		Data: map[string]interface{}{
			"address":  address,
			"verified": verified,
			"score":    score,
		},
	}, nil
}
