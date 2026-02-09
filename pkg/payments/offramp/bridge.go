package offramp

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"
	"time"
)

// bridgeImpl aggregates multiple off-ramp adapters.
type bridgeImpl struct {
	adapters   map[string]Adapter
	adapterMu  sync.RWMutex
	operations map[string]PayoutResult
	opMu       sync.RWMutex
}

// NewBridge creates a new off-ramp bridge.
func NewBridge() *bridgeImpl {
	return &bridgeImpl{
		adapters:   make(map[string]Adapter),
		operations: make(map[string]PayoutResult),
	}
}

// RegisterAdapter registers an adapter with the bridge.
func (b *bridgeImpl) RegisterAdapter(adapter Adapter) error {
	if adapter == nil {
		return ErrAdapterUnavailable
	}
	b.adapterMu.Lock()
	defer b.adapterMu.Unlock()
	b.adapters[adapter.Name()] = adapter
	return nil
}

// GetQuote returns the best quote across adapters.
func (b *bridgeImpl) GetQuote(ctx context.Context, req QuoteRequest) (Quote, error) {
	if req.CryptoAmount.IsNil() || !req.CryptoAmount.IsPositive() {
		return Quote{}, ErrInvalidRequest
	}
	if req.FiatCurrency == "" || req.PaymentMethod == "" {
		return Quote{}, ErrInvalidRequest
	}

	adapter, err := b.selectBestAdapter(ctx, req)
	if err != nil {
		return Quote{}, err
	}

	quote, err := adapter.GetQuote(ctx, req)
	if err != nil {
		return Quote{}, err
	}

	if quote.ExpiresAt.IsZero() {
		quote.ExpiresAt = time.Now().UTC().Add(5 * time.Minute)
	}

	return quote, nil
}

// InitiatePayout executes a payout via the selected adapter.
func (b *bridgeImpl) InitiatePayout(ctx context.Context, quote Quote, cryptoTxRef string, destination string, metadata map[string]string) (PayoutResult, error) {
	if quote.IsExpired(time.Now().UTC()) {
		return PayoutResult{}, ErrQuoteExpired
	}

	b.adapterMu.RLock()
	adapter, ok := b.adapters[quote.Provider]
	b.adapterMu.RUnlock()
	if !ok {
		return PayoutResult{}, ErrAdapterUnavailable
	}

	result, err := adapter.InitiatePayout(ctx, PayoutRequest{
		Quote:       quote,
		CryptoTxRef: cryptoTxRef,
		Destination: destination,
		Metadata:    metadata,
	})
	if err != nil {
		return PayoutResult{}, err
	}

	b.opMu.Lock()
	b.operations[result.ID] = result
	b.opMu.Unlock()

	return result, nil
}

// GetStatus retrieves payout status and refreshes from adapter when available.
func (b *bridgeImpl) GetStatus(ctx context.Context, payoutID string) (PayoutResult, error) {
	b.opMu.RLock()
	result, ok := b.operations[payoutID]
	b.opMu.RUnlock()
	if !ok {
		return PayoutResult{}, fmt.Errorf("payout %s not found", payoutID)
	}

	if result.Status == StatusPending || result.Status == StatusProcessing {
		b.adapterMu.RLock()
		adapter, exists := b.adapters[result.Provider]
		b.adapterMu.RUnlock()
		if exists {
			updated, err := adapter.GetStatus(ctx, payoutID)
			if err == nil {
				b.opMu.Lock()
				b.operations[payoutID] = updated
				b.opMu.Unlock()
				return updated, nil
			}
		}
	}

	return result, nil
}

// Cancel attempts to cancel a payout.
func (b *bridgeImpl) Cancel(ctx context.Context, payoutID string) error {
	b.opMu.RLock()
	result, ok := b.operations[payoutID]
	b.opMu.RUnlock()
	if !ok {
		return fmt.Errorf("payout %s not found", payoutID)
	}

	b.adapterMu.RLock()
	adapter, ok := b.adapters[result.Provider]
	b.adapterMu.RUnlock()
	if !ok {
		return ErrAdapterUnavailable
	}

	if err := adapter.Cancel(ctx, payoutID); err != nil {
		return err
	}

	b.opMu.Lock()
	result.Status = StatusCancelled
	b.operations[payoutID] = result
	b.opMu.Unlock()
	return nil
}

// ListProviders lists registered adapters.
func (b *bridgeImpl) ListProviders() []string {
	b.adapterMu.RLock()
	defer b.adapterMu.RUnlock()

	providers := make([]string, 0, len(b.adapters))
	for name := range b.adapters {
		providers = append(providers, name)
	}
	sort.Strings(providers)
	return providers
}

func (b *bridgeImpl) selectBestAdapter(ctx context.Context, req QuoteRequest) (Adapter, error) {
	b.adapterMu.RLock()
	defer b.adapterMu.RUnlock()

	var best Adapter
	var bestRate float64

	for _, adapter := range b.adapters {
		if !adapter.IsHealthy(ctx) {
			continue
		}
		if !adapter.SupportsCurrency(req.FiatCurrency) || !adapter.SupportsMethod(req.PaymentMethod) {
			continue
		}

		quote, err := adapter.GetQuote(ctx, req)
		if err != nil {
			continue
		}
		rate, _ := quote.ExchangeRate.Float64()
		if best == nil || rate > bestRate {
			best = adapter
			bestRate = rate
		}
	}

	if best == nil {
		return nil, ErrAdapterUnavailable
	}
	return best, nil
}

func generateID(prefix string) (string, error) {
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s", prefix, hex.EncodeToString(buf)), nil
}
