package keeper

import (
	"context"
	"fmt"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/pkg/cache"
	types "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
)

// ============================================================================
// Provider Cache Layer
// PERF-003: Caching Layer Implementation
// ============================================================================

// ProviderCache provides caching for frequently accessed provider data.
// It wraps the keeper and caches provider information.
type ProviderCache struct {
	keeper *Keeper

	// Provider cache: address -> Provider
	providerCache cache.LoadingCache[string, cachedProvider]

	// Public key cache: address -> public key bytes
	pubKeyCache cache.LoadingCache[string, cachedPubKey]

	// Configuration
	config ProviderCacheConfig

	// Metrics
	metrics *cache.Metrics

	// State
	mu     sync.RWMutex
	closed bool
}

// cachedProvider holds a cached provider with found status.
type cachedProvider struct {
	Provider types.Provider
	Found    bool
}

// cachedPubKey holds a cached public key with found status.
type cachedPubKey struct {
	PublicKey []byte
	Found     bool
}

// ProviderCacheConfig configures the provider cache layer.
type ProviderCacheConfig struct {
	// Enabled determines if caching is enabled.
	Enabled bool

	// ProviderTTL is the TTL for cached provider info.
	ProviderTTL time.Duration

	// PubKeyTTL is the TTL for cached public keys.
	PubKeyTTL time.Duration

	// MaxProviderEntries is the maximum number of cached providers.
	MaxProviderEntries int

	// MaxPubKeyEntries is the maximum number of cached public keys.
	MaxPubKeyEntries int

	// EnableMetrics enables cache metrics collection.
	EnableMetrics bool
}

// DefaultProviderCacheConfig returns the default provider cache configuration.
func DefaultProviderCacheConfig() ProviderCacheConfig {
	return ProviderCacheConfig{
		Enabled:            true,
		ProviderTTL:        10 * time.Minute, // Provider info changes infrequently
		PubKeyTTL:          10 * time.Minute, // Public keys rarely change
		MaxProviderEntries: 5000,
		MaxPubKeyEntries:   5000,
		EnableMetrics:      true,
	}
}

// NewProviderCache creates a new provider cache wrapper around a keeper.
func NewProviderCache(keeper *Keeper, config ProviderCacheConfig) *ProviderCache {
	pc := &ProviderCache{
		keeper: keeper,
		config: config,
	}

	if !config.Enabled {
		return pc
	}

	// Create provider cache
	pc.providerCache = cache.NewMemoryCache(
		cache.WithMaxSize[string, cachedProvider](config.MaxProviderEntries),
		cache.WithDefaultTTL[string, cachedProvider](config.ProviderTTL),
	)

	// Create public key cache
	pc.pubKeyCache = cache.NewMemoryCache(
		cache.WithMaxSize[string, cachedPubKey](config.MaxPubKeyEntries),
		cache.WithDefaultTTL[string, cachedPubKey](config.PubKeyTTL),
	)

	// Setup metrics
	if config.EnableMetrics {
		pc.metrics = cache.NewMetrics("provider_cache", 1*time.Minute)
		pc.metrics.RegisterCache("providers", pc.providerCache.(cache.MetricsProvider))
		pc.metrics.RegisterCache("pubkeys", pc.pubKeyCache.(cache.MetricsProvider))
	}

	return pc
}

// Get returns a provider by address, using cache if available.
func (pc *ProviderCache) Get(ctx sdk.Context, id sdk.Address) (types.Provider, bool) {
	if !pc.config.Enabled || pc.providerCache == nil {
		return pc.keeper.Get(ctx, id)
	}

	key := sdk.AccAddress(id.Bytes()).String()
	bgCtx := context.Background()

	// Try to get from cache with loader
	cached, err := pc.providerCache.GetOrLoad(bgCtx, key, func(_ context.Context, k string) (cachedProvider, error) {
		provider, found := pc.keeper.Get(ctx, id)
		return cachedProvider{Provider: provider, Found: found}, nil
	})

	if err != nil {
		return pc.keeper.Get(ctx, id)
	}

	return cached.Provider, cached.Found
}

// GetProviderPublicKey returns the public key for a provider, using cache if available.
func (pc *ProviderCache) GetProviderPublicKey(ctx sdk.Context, providerAddr sdk.AccAddress) ([]byte, bool) {
	if !pc.config.Enabled || pc.pubKeyCache == nil {
		return pc.keeper.GetProviderPublicKey(ctx, providerAddr)
	}

	key := providerAddr.String()
	bgCtx := context.Background()

	// Try to get from cache with loader
	cached, err := pc.pubKeyCache.GetOrLoad(bgCtx, key, func(_ context.Context, k string) (cachedPubKey, error) {
		pubKey, found := pc.keeper.GetProviderPublicKey(ctx, providerAddr)
		return cachedPubKey{PublicKey: pubKey, Found: found}, nil
	})

	if err != nil {
		return pc.keeper.GetProviderPublicKey(ctx, providerAddr)
	}

	return cached.PublicKey, cached.Found
}

// ProviderExists checks if a provider exists, using cache if available.
func (pc *ProviderCache) ProviderExists(ctx sdk.Context, providerAddr sdk.AccAddress) bool {
	_, found := pc.Get(ctx, providerAddr)
	return found
}

// InvalidateProvider removes a provider from the cache.
func (pc *ProviderCache) InvalidateProvider(id sdk.Address) {
	if !pc.config.Enabled || pc.providerCache == nil {
		return
	}
	key := sdk.AccAddress(id.Bytes()).String()
	_ = pc.providerCache.Delete(context.Background(), key)
}

// InvalidatePublicKey removes a public key from the cache.
func (pc *ProviderCache) InvalidatePublicKey(providerAddr sdk.AccAddress) {
	if !pc.config.Enabled || pc.pubKeyCache == nil {
		return
	}
	_ = pc.pubKeyCache.Delete(context.Background(), providerAddr.String())
}

// InvalidateAll clears all caches.
func (pc *ProviderCache) InvalidateAll() {
	if !pc.config.Enabled {
		return
	}

	bgCtx := context.Background()
	if pc.providerCache != nil {
		_ = pc.providerCache.Clear(bgCtx)
	}
	if pc.pubKeyCache != nil {
		_ = pc.pubKeyCache.Clear(bgCtx)
	}
}

// Stats returns cache statistics.
func (pc *ProviderCache) Stats() ProviderCacheStats {
	stats := ProviderCacheStats{
		Enabled: pc.config.Enabled,
	}

	if !pc.config.Enabled || pc.metrics == nil {
		return stats
	}

	if providerStats, ok := pc.metrics.GetCacheStats("providers"); ok {
		stats.ProviderHits = providerStats.Hits
		stats.ProviderMisses = providerStats.Misses
		stats.ProviderSize = providerStats.Size
		stats.ProviderHitRate = providerStats.HitRate()
	}

	if pubKeyStats, ok := pc.metrics.GetCacheStats("pubkeys"); ok {
		stats.PubKeyHits = pubKeyStats.Hits
		stats.PubKeyMisses = pubKeyStats.Misses
		stats.PubKeySize = pubKeyStats.Size
		stats.PubKeyHitRate = pubKeyStats.HitRate()
	}

	return stats
}

// Close releases cache resources.
func (pc *ProviderCache) Close() error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.closed {
		return nil
	}
	pc.closed = true

	if pc.metrics != nil {
		pc.metrics.Stop()
	}

	var errs []error
	if pc.providerCache != nil {
		if err := pc.providerCache.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if pc.pubKeyCache != nil {
		if err := pc.pubKeyCache.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("provider cache close errors: %v", errs)
	}
	return nil
}

// ProviderCacheStats contains provider cache statistics.
type ProviderCacheStats struct {
	Enabled        bool    `json:"enabled"`
	ProviderHits   uint64  `json:"provider_hits"`
	ProviderMisses uint64  `json:"provider_misses"`
	ProviderSize   int     `json:"provider_size"`
	ProviderHitRate float64 `json:"provider_hit_rate"`
	PubKeyHits     uint64  `json:"pubkey_hits"`
	PubKeyMisses   uint64  `json:"pubkey_misses"`
	PubKeySize     int     `json:"pubkey_size"`
	PubKeyHitRate  float64 `json:"pubkey_hit_rate"`
}

// ============================================================================
// Cached Provider Keeper Wrapper
// ============================================================================

// CachedProviderKeeper wraps a Provider Keeper with caching support.
type CachedProviderKeeper struct {
	Keeper
	cache *ProviderCache
}

// NewCachedProviderKeeper creates a new cached provider keeper wrapper.
func NewCachedProviderKeeper(keeper Keeper, config ProviderCacheConfig) *CachedProviderKeeper {
	k := &keeper
	return &CachedProviderKeeper{
		Keeper: keeper,
		cache:  NewProviderCache(k, config),
	}
}

// Get returns a provider using cache.
func (ck *CachedProviderKeeper) Get(ctx sdk.Context, id sdk.Address) (types.Provider, bool) {
	return ck.cache.Get(ctx, id)
}

// GetProviderPublicKey returns a public key using cache.
func (ck *CachedProviderKeeper) GetProviderPublicKey(ctx sdk.Context, providerAddr sdk.AccAddress) ([]byte, bool) {
	return ck.cache.GetProviderPublicKey(ctx, providerAddr)
}

// ProviderExists checks existence using cache.
func (ck *CachedProviderKeeper) ProviderExists(ctx sdk.Context, providerAddr sdk.AccAddress) bool {
	return ck.cache.ProviderExists(ctx, providerAddr)
}

// Create creates a provider and invalidates cache.
func (ck *CachedProviderKeeper) Create(ctx sdk.Context, provider types.Provider) error {
	err := ck.Keeper.Create(ctx, provider)
	if err == nil {
		owner, _ := sdk.AccAddressFromBech32(provider.Owner)
		ck.cache.InvalidateProvider(owner)
	}
	return err
}

// Update updates a provider and invalidates cache.
func (ck *CachedProviderKeeper) Update(ctx sdk.Context, provider types.Provider) error {
	err := ck.Keeper.Update(ctx, provider)
	if err == nil {
		owner, _ := sdk.AccAddressFromBech32(provider.Owner)
		ck.cache.InvalidateProvider(owner)
	}
	return err
}

// Delete deletes a provider and invalidates cache.
func (ck *CachedProviderKeeper) Delete(ctx sdk.Context, id sdk.Address) {
	ck.Keeper.Delete(ctx, id)
	ck.cache.InvalidateProvider(id)
	ck.cache.InvalidatePublicKey(sdk.AccAddress(id.Bytes()))
}

// SetProviderPublicKey sets a public key and invalidates cache.
func (ck *CachedProviderKeeper) SetProviderPublicKey(ctx sdk.Context, owner sdk.AccAddress, pubKey []byte, keyType string) error {
	err := ck.Keeper.SetProviderPublicKey(ctx, owner, pubKey, keyType)
	if err == nil {
		ck.cache.InvalidatePublicKey(owner)
	}
	return err
}

// RotateProviderPublicKey rotates a public key and invalidates cache.
func (ck *CachedProviderKeeper) RotateProviderPublicKey(ctx sdk.Context, owner sdk.AccAddress, newKey []byte, keyType string, signature []byte) error {
	err := ck.Keeper.RotateProviderPublicKey(ctx, owner, newKey, keyType, signature)
	if err == nil {
		ck.cache.InvalidatePublicKey(owner)
	}
	return err
}

// DeleteProviderPublicKey deletes a public key and invalidates cache.
func (ck *CachedProviderKeeper) DeleteProviderPublicKey(ctx sdk.Context, owner sdk.AccAddress) {
	ck.Keeper.DeleteProviderPublicKey(ctx, owner)
	ck.cache.InvalidatePublicKey(owner)
}

// CacheStats returns cache statistics.
func (ck *CachedProviderKeeper) CacheStats() ProviderCacheStats {
	return ck.cache.Stats()
}

// Close closes the cache resources.
func (ck *CachedProviderKeeper) Close() error {
	return ck.cache.Close()
}
