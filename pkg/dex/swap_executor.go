// Package dex provides DEX integration for crypto-to-fiat conversions.
//
// VE-905: DEX integration for crypto-to-fiat conversions
package dex

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"time"

	sdkmath "cosmossdk.io/math"
)

// swapExecutorImpl implements the SwapExecutor interface
type swapExecutorImpl struct {
	cfg     SwapConfig
	service *service
}

// newSwapExecutor creates a new swap executor
func newSwapExecutor(cfg SwapConfig, svc *service) *swapExecutorImpl {
	return &swapExecutorImpl{
		cfg:     cfg,
		service: svc,
	}
}

// GetQuote generates a swap quote with optimal routing
func (e *swapExecutorImpl) GetQuote(ctx context.Context, request SwapRequest) (SwapQuote, error) {
	// Apply default slippage if not specified
	if request.SlippageTolerance == 0 {
		request.SlippageTolerance = e.cfg.DefaultSlippage
	}

	// Validate slippage
	if request.SlippageTolerance > e.cfg.MaxSlippage {
		return SwapQuote{}, fmt.Errorf("slippage tolerance %f exceeds maximum %f", request.SlippageTolerance, e.cfg.MaxSlippage)
	}

	// Find the best route
	route, err := e.FindBestRoute(ctx, request)
	if err != nil {
		return SwapQuote{}, err
	}

	// Calculate amounts from route
	var inputAmount, outputAmount sdkmath.Int
	if request.Type == SwapTypeExactIn {
		inputAmount = request.Amount
		outputAmount = e.calculateRouteOutput(route)
	} else {
		outputAmount = request.Amount
		inputAmount = e.calculateRouteInput(route)
	}

	// Calculate minimum output with slippage
	slippageFactor := sdkmath.LegacyOneDec().Sub(sdkmath.LegacyNewDecWithPrec(int64(request.SlippageTolerance*10000), 4))
	minOutputAmount := slippageFactor.MulInt(outputAmount).TruncateInt()

	// Calculate rate
	rate := sdkmath.LegacyNewDecFromInt(outputAmount).Quo(sdkmath.LegacyNewDecFromInt(inputAmount))

	// Calculate total fee
	totalFee := e.calculateRouteFee(route, inputAmount)

	// Generate quote ID
	quoteID, err := generateQuoteID()
	if err != nil {
		return SwapQuote{}, fmt.Errorf("failed to generate quote ID: %w", err)
	}

	now := time.Now().UTC()

	return SwapQuote{
		ID:              quoteID,
		Request:         request,
		Route:           route,
		InputAmount:     inputAmount,
		OutputAmount:    outputAmount,
		MinOutputAmount: minOutputAmount,
		Rate:            rate,
		PriceImpact:     route.PriceImpact,
		TotalFee:        totalFee,
		GasEstimate:     route.TotalGas,
		ExpiresAt:       now.Add(e.cfg.QuoteValidityPeriod),
		CreatedAt:       now,
	}, nil
}

// ExecuteSwap executes a previously quoted swap
func (e *swapExecutorImpl) ExecuteSwap(ctx context.Context, quote SwapQuote, signedTx []byte) (SwapResult, error) {
	// Validate quote
	if err := e.ValidateQuote(ctx, quote); err != nil {
		return SwapResult{}, err
	}

	// Get the adapter for the first hop
	if len(quote.Route.Hops) == 0 {
		return SwapResult{}, errors.New("no route hops")
	}

	adapterName := quote.Route.Hops[0].DEX
	adapter, err := e.service.GetAdapter(adapterName)
	if err != nil {
		return SwapResult{}, err
	}

	// Execute through adapter
	result, err := adapter.ExecuteSwap(ctx, quote, signedTx)
	if err != nil {
		return SwapResult{}, fmt.Errorf("swap execution failed: %w", err)
	}

	return result, nil
}

// FindBestRoute finds the optimal swap route
func (e *swapExecutorImpl) FindBestRoute(ctx context.Context, request SwapRequest) (SwapRoute, error) {
	if !e.cfg.EnableRouteFinding {
		// Simple direct route
		return e.findDirectRoute(ctx, request)
	}

	// Try to find multi-hop routes
	routes, err := e.findAllRoutes(ctx, request)
	if err != nil {
		return SwapRoute{}, err
	}

	if len(routes) == 0 {
		return SwapRoute{}, ErrUnsupportedPair
	}

	// Sort by output amount (descending) for exact-in, input amount (ascending) for exact-out
	if request.Type == SwapTypeExactIn {
		sort.Slice(routes, func(i, j int) bool {
			outI := e.calculateRouteOutput(routes[i])
			outJ := e.calculateRouteOutput(routes[j])
			return outI.GT(outJ)
		})
	} else {
		sort.Slice(routes, func(i, j int) bool {
			inI := e.calculateRouteInput(routes[i])
			inJ := e.calculateRouteInput(routes[j])
			return inI.LT(inJ)
		})
	}

	return routes[0], nil
}

// ValidateQuote validates if a quote is still executable
func (e *swapExecutorImpl) ValidateQuote(ctx context.Context, quote SwapQuote) error {
	// Check expiration
	if quote.IsExpired() {
		return ErrQuoteExpired
	}

	// Validate current price hasn't deviated too much
	price, err := e.service.GetPrice(ctx, quote.Request.FromToken.Symbol, quote.Request.ToToken.Symbol)
	if err != nil {
		return fmt.Errorf("failed to get current price: %w", err)
	}

	// Calculate deviation from quote rate
	deviation := quote.Rate.Sub(price.Rate).Abs().Quo(quote.Rate)
	maxDeviation := sdkmath.LegacyNewDecWithPrec(int64(quote.Request.SlippageTolerance*10000), 4)

	if deviation.GT(maxDeviation) {
		return ErrSlippageExceeded
	}

	return nil
}

// GetSlippageEstimate estimates potential slippage
func (e *swapExecutorImpl) GetSlippageEstimate(ctx context.Context, request SwapRequest) (float64, error) {
	route, err := e.FindBestRoute(ctx, request)
	if err != nil {
		return 0, err
	}
	return route.PriceImpact, nil
}

// findDirectRoute finds a direct route through a single pool
func (e *swapExecutorImpl) findDirectRoute(ctx context.Context, request SwapRequest) (SwapRoute, error) {
	adapters := e.service.getHealthyAdapters(ctx)

	for _, adapter := range adapters {
		quote, err := adapter.GetSwapQuote(ctx, request)
		if err == nil {
			return quote.Route, nil
		}
	}

	return SwapRoute{}, ErrUnsupportedPair
}

// findAllRoutes finds all possible routes
//
//nolint:unparam // result 1 (error) reserved for future multi-hop discovery failures
func (e *swapExecutorImpl) findAllRoutes(ctx context.Context, request SwapRequest) ([]SwapRoute, error) {
	adapters := e.service.getHealthyAdapters(ctx)

	var routes []SwapRoute
	for _, adapter := range adapters {
		quote, err := adapter.GetSwapQuote(ctx, request)
		if err == nil {
			routes = append(routes, quote.Route)
		}
	}

	if e.cfg.MaxHops <= 1 {
		return dedupeRoutes(routes), nil
	}

	pools, err := e.service.ListPools(ctx, PoolQuery{})
	if err != nil {
		return dedupeRoutes(routes), nil
	}

	switch request.Type {
	case SwapTypeExactOut:
		reverseVisited := map[string]struct{}{request.ToToken.Symbol: {}}
		routes = append(routes, e.discoverExactOutRoutes(ctx, request, request.ToToken, request.FromToken, request.Amount, pools, reverseVisited, e.cfg.MaxHops)...)
	default:
		forwardVisited := map[string]struct{}{request.FromToken.Symbol: {}}
		routes = append(routes, e.discoverExactInRoutes(ctx, request, request.FromToken, request.ToToken, request.Amount, pools, forwardVisited, e.cfg.MaxHops)...)
	}

	return dedupeRoutes(routes), nil
}

func (e *swapExecutorImpl) discoverExactInRoutes(
	ctx context.Context,
	baseRequest SwapRequest,
	current Token,
	target Token,
	amountIn sdkmath.Int,
	pools []LiquidityPool,
	visited map[string]struct{},
	remainingHops int,
) []SwapRoute {
	if remainingHops <= 0 {
		return nil
	}

	var routes []SwapRoute
	for _, edge := range findPoolEdges(pools, current.Symbol, visited) {
		hopRequest := baseRequest
		hopRequest.FromToken = current
		hopRequest.ToToken = edge.ToToken
		hopRequest.Amount = amountIn
		hopRequest.Type = SwapTypeExactIn

		segment, ok := e.quoteRouteSegment(ctx, hopRequest, remainingHops)
		if !ok {
			continue
		}

		if edge.ToToken.Symbol == target.Symbol {
			routes = append(routes, segment)
			continue
		}

		nextVisited := cloneVisitedTokens(visited)
		nextVisited[edge.ToToken.Symbol] = struct{}{}
		nextAmount := e.calculateRouteOutput(segment)
		if !nextAmount.IsPositive() {
			continue
		}

		childRoutes := e.discoverExactInRoutes(
			ctx,
			baseRequest,
			edge.ToToken,
			target,
			nextAmount,
			pools,
			nextVisited,
			remainingHops-len(segment.Hops),
		)
		for _, child := range childRoutes {
			routes = append(routes, combineRoutes(segment, child))
		}
	}

	return routes
}

func (e *swapExecutorImpl) discoverExactOutRoutes(
	ctx context.Context,
	baseRequest SwapRequest,
	currentTarget Token,
	source Token,
	requiredOutput sdkmath.Int,
	pools []LiquidityPool,
	visited map[string]struct{},
	remainingHops int,
) []SwapRoute {
	if remainingHops <= 0 {
		return nil
	}

	var routes []SwapRoute
	for _, edge := range findPoolEdges(pools, currentTarget.Symbol, visited) {
		hopRequest := baseRequest
		hopRequest.FromToken = edge.ToToken
		hopRequest.ToToken = currentTarget
		hopRequest.Amount = requiredOutput
		hopRequest.Type = SwapTypeExactOut

		segment, ok := e.quoteRouteSegment(ctx, hopRequest, remainingHops)
		if !ok {
			continue
		}

		requiredInput := e.calculateRouteInput(segment)
		if !requiredInput.IsPositive() {
			continue
		}

		if edge.ToToken.Symbol == source.Symbol {
			routes = append(routes, segment)
			continue
		}

		nextVisited := cloneVisitedTokens(visited)
		nextVisited[edge.ToToken.Symbol] = struct{}{}

		childRoutes := e.discoverExactOutRoutes(
			ctx,
			baseRequest,
			edge.ToToken,
			source,
			requiredInput,
			pools,
			nextVisited,
			remainingHops-len(segment.Hops),
		)
		for _, child := range childRoutes {
			routes = append(routes, combineRoutes(child, segment))
		}
	}

	return routes
}

func (e *swapExecutorImpl) quoteRouteSegment(ctx context.Context, request SwapRequest, remainingHops int) (SwapRoute, bool) {
	adapter, err := e.service.GetAdapter(request.PreferredDEX)
	if request.PreferredDEX != "" && err == nil && adapter != nil && !adapter.IsHealthy(ctx) {
		return SwapRoute{}, false
	}

	if request.PreferredDEX != "" && err == nil && adapter != nil {
		quote, quoteErr := adapter.GetSwapQuote(ctx, request)
		if quoteErr != nil || len(quote.Route.Hops) == 0 || len(quote.Route.Hops) > remainingHops {
			return SwapRoute{}, false
		}
		return quote.Route, true
	}

	adapters := e.service.getHealthyAdapters(ctx)
	for _, adapter := range adapters {
		quote, quoteErr := adapter.GetSwapQuote(ctx, request)
		if quoteErr != nil || len(quote.Route.Hops) == 0 || len(quote.Route.Hops) > remainingHops {
			continue
		}
		return quote.Route, true
	}

	return SwapRoute{}, false
}

type poolEdge struct {
	ToToken Token
}

func findPoolEdges(pools []LiquidityPool, currentSymbol string, visited map[string]struct{}) []poolEdge {
	var edges []poolEdge
	seen := make(map[string]struct{})

	for _, pool := range pools {
		if !poolContainsToken(pool, currentSymbol) {
			continue
		}

		for _, token := range pool.Tokens {
			if token.Symbol == currentSymbol {
				continue
			}
			if _, alreadyVisited := visited[token.Symbol]; alreadyVisited {
				continue
			}
			key := currentSymbol + "->" + token.Symbol
			if _, duplicate := seen[key]; duplicate {
				continue
			}
			seen[key] = struct{}{}
			edges = append(edges, poolEdge{ToToken: token})
		}
	}

	return edges
}

func poolContainsToken(pool LiquidityPool, symbol string) bool {
	for _, token := range pool.Tokens {
		if token.Symbol == symbol {
			return true
		}
	}

	return false
}

func cloneVisitedTokens(visited map[string]struct{}) map[string]struct{} {
	next := make(map[string]struct{}, len(visited)+1)
	for key := range visited {
		next[key] = struct{}{}
	}
	return next
}

func combineRoutes(left, right SwapRoute) SwapRoute {
	hops := make([]SwapHop, 0, len(left.Hops)+len(right.Hops))
	hops = append(hops, left.Hops...)
	hops = append(hops, right.Hops...)

	return SwapRoute{
		Hops:        hops,
		TotalGas:    left.TotalGas + right.TotalGas,
		PriceImpact: left.PriceImpact + right.PriceImpact,
	}
}

func dedupeRoutes(routes []SwapRoute) []SwapRoute {
	seen := make(map[string]struct{}, len(routes))
	deduped := make([]SwapRoute, 0, len(routes))

	for _, route := range routes {
		key := routeSignature(route)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		deduped = append(deduped, route)
	}

	return deduped
}

func routeSignature(route SwapRoute) string {
	signature := fmt.Sprintf("gas:%d:impact:%0.6f", route.TotalGas, route.PriceImpact)
	for _, hop := range route.Hops {
		signature += fmt.Sprintf("|%s:%s:%s>%s:%s:%s",
			hop.DEX,
			hop.PoolID,
			hop.FromToken.Symbol,
			hop.ToToken.Symbol,
			hop.AmountIn.String(),
			hop.AmountOut.String(),
		)
	}

	return signature
}

// calculateRouteOutput calculates the total output amount from a route
func (e *swapExecutorImpl) calculateRouteOutput(route SwapRoute) sdkmath.Int {
	if len(route.Hops) == 0 {
		return sdkmath.ZeroInt()
	}
	return route.Hops[len(route.Hops)-1].AmountOut
}

// calculateRouteInput calculates the total input amount for a route
func (e *swapExecutorImpl) calculateRouteInput(route SwapRoute) sdkmath.Int {
	if len(route.Hops) == 0 {
		return sdkmath.ZeroInt()
	}
	return route.Hops[0].AmountIn
}

// calculateRouteFee calculates the total fee for a route
//
//nolint:unparam // inputAmount kept for future slippage-based fee calculation
func (e *swapExecutorImpl) calculateRouteFee(route SwapRoute, _ sdkmath.Int) sdkmath.Int {
	totalFee := sdkmath.ZeroInt()
	for _, hop := range route.Hops {
		hopFee := hop.Fee.MulInt(hop.AmountIn).TruncateInt()
		totalFee = totalFee.Add(hopFee)
	}
	return totalFee
}

// generateQuoteID generates a unique quote ID
func generateQuoteID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "quote_" + hex.EncodeToString(bytes), nil
}
