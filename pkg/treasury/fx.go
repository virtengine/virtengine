package treasury

import (
	"context"
	"sort"
	"sync"
	"time"

	sdkmath "cosmossdk.io/math"

	"github.com/virtengine/virtengine/pkg/pricefeed"
)

type FXSnapshot struct {
	BaseAsset  string
	QuoteAsset string
	Rate       sdkmath.LegacyDec
	Timestamp  time.Time
	Source     string
}

type FXRateProvider interface {
	GetPrice(ctx context.Context, baseAsset, quoteAsset string) (pricefeed.PriceData, error)
}

type FXSnapshotStore interface {
	Save(snapshot FXSnapshot)
	GetAt(baseAsset, quoteAsset string, at time.Time) (FXSnapshot, bool)
	List(baseAsset, quoteAsset string) []FXSnapshot
}

// InMemorySnapshotStore keeps snapshots in memory sorted by time.
type InMemorySnapshotStore struct {
	mu        sync.RWMutex
	snapshots map[string][]FXSnapshot
}

func NewInMemorySnapshotStore() *InMemorySnapshotStore {
	return &InMemorySnapshotStore{snapshots: make(map[string][]FXSnapshot)}
}

func (s *InMemorySnapshotStore) Save(snapshot FXSnapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := fxKey(snapshot.BaseAsset, snapshot.QuoteAsset)
	s.snapshots[key] = append(s.snapshots[key], snapshot)
	sort.Slice(s.snapshots[key], func(i, j int) bool {
		return s.snapshots[key][i].Timestamp.Before(s.snapshots[key][j].Timestamp)
	})
}

func (s *InMemorySnapshotStore) GetAt(baseAsset, quoteAsset string, at time.Time) (FXSnapshot, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	key := fxKey(baseAsset, quoteAsset)
	items := s.snapshots[key]
	if len(items) == 0 {
		return FXSnapshot{}, false
	}
	idx := sort.Search(len(items), func(i int) bool {
		return !items[i].Timestamp.Before(at)
	})
	if idx == 0 {
		if items[0].Timestamp.After(at) {
			return FXSnapshot{}, false
		}
	}
	if idx == len(items) || items[idx].Timestamp.After(at) {
		idx--
	}
	if idx < 0 {
		return FXSnapshot{}, false
	}
	return items[idx], true
}

func (s *InMemorySnapshotStore) List(baseAsset, quoteAsset string) []FXSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	key := fxKey(baseAsset, quoteAsset)
	items := s.snapshots[key]
	out := make([]FXSnapshot, len(items))
	copy(out, items)
	return out
}

// FXService fetches FX rates and stores historical snapshots.
type FXService struct {
	provider FXRateProvider
	store    FXSnapshotStore
}

func NewFXService(provider FXRateProvider, store FXSnapshotStore) *FXService {
	return &FXService{provider: provider, store: store}
}

func (s *FXService) GetRate(ctx context.Context, baseAsset, quoteAsset string) (FXSnapshot, error) {
	price, err := s.provider.GetPrice(ctx, baseAsset, quoteAsset)
	if err != nil {
		return FXSnapshot{}, err
	}

	snapshot := FXSnapshot{
		BaseAsset:  baseAsset,
		QuoteAsset: quoteAsset,
		Rate:       price.Price,
		Timestamp:  price.Timestamp,
		Source:     price.Source,
	}

	if s.store != nil {
		s.store.Save(snapshot)
	}

	return snapshot, nil
}

func (s *FXService) GetHistoricalRate(baseAsset, quoteAsset string, at time.Time) (FXSnapshot, bool) {
	if s.store == nil {
		return FXSnapshot{}, false
	}
	return s.store.GetAt(baseAsset, quoteAsset, at)
}

func fxKey(base, quote string) string {
	return base + "/" + quote
}

// AggregatorProvider wraps pricefeed aggregator to satisfy FXRateProvider.
type AggregatorProvider struct {
	aggregator *pricefeed.PriceFeedAggregator
}

func NewAggregatorProvider(aggregator *pricefeed.PriceFeedAggregator) *AggregatorProvider {
	return &AggregatorProvider{aggregator: aggregator}
}

func (a *AggregatorProvider) GetPrice(ctx context.Context, baseAsset, quoteAsset string) (pricefeed.PriceData, error) {
	price, err := a.aggregator.GetPrice(ctx, baseAsset, quoteAsset)
	if err != nil {
		return pricefeed.PriceData{}, err
	}
	return price.PriceData, nil
}
