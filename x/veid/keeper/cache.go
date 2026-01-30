package keeper

import (
	"context"
	"fmt"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/pkg/cache"
	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// VEID Cache Layer
// PERF-003: Caching Layer Implementation
// ============================================================================

// VEIDCache provides caching for frequently accessed VEID data.
// It wraps the keeper and caches identity records and scores.
type VEIDCache struct {
	keeper *Keeper

	// Score cache: address -> score
	scoreCache cache.LoadingCache[string, cachedScore]

	// Identity record cache: address -> record
	recordCache cache.LoadingCache[string, types.IdentityRecord]

	// Configuration
	config VEIDCacheConfig

	// Metrics
	metrics *cache.Metrics

	// State
	mu     sync.RWMutex
	closed bool
}

// cachedScore holds a cached VEID score with its status.
type cachedScore struct {
	Score  uint32
	Status types.AccountStatus
	Found  bool
}

// VEIDCacheConfig configures the VEID cache layer.
type VEIDCacheConfig struct {
	// Enabled determines if caching is enabled.
	Enabled bool

	// ScoreTTL is the TTL for cached scores.
	ScoreTTL time.Duration

	// RecordTTL is the TTL for cached identity records.
	RecordTTL time.Duration

	// MaxScoreEntries is the maximum number of cached scores.
	MaxScoreEntries int

	// MaxRecordEntries is the maximum number of cached records.
	MaxRecordEntries int

	// EnableMetrics enables cache metrics collection.
	EnableMetrics bool
}

// DefaultVEIDCacheConfig returns the default VEID cache configuration.
func DefaultVEIDCacheConfig() VEIDCacheConfig {
	return VEIDCacheConfig{
		Enabled:          true,
		ScoreTTL:         5 * time.Minute,
		RecordTTL:        5 * time.Minute,
		MaxScoreEntries:  10000,
		MaxRecordEntries: 5000,
		EnableMetrics:    true,
	}
}

// NewVEIDCache creates a new VEID cache wrapper around a keeper.
func NewVEIDCache(keeper *Keeper, config VEIDCacheConfig) *VEIDCache {
	vc := &VEIDCache{
		keeper: keeper,
		config: config,
	}

	if !config.Enabled {
		return vc
	}

	// Create score cache
	vc.scoreCache = cache.NewMemoryCache(
		cache.WithMaxSize[string, cachedScore](config.MaxScoreEntries),
		cache.WithDefaultTTL[string, cachedScore](config.ScoreTTL),
	)

	// Create record cache
	vc.recordCache = cache.NewMemoryCache(
		cache.WithMaxSize[string, types.IdentityRecord](config.MaxRecordEntries),
		cache.WithDefaultTTL[string, types.IdentityRecord](config.RecordTTL),
	)

	// Setup metrics
	if config.EnableMetrics {
		vc.metrics = cache.NewMetrics("veid_cache", 1*time.Minute)
		vc.metrics.RegisterCache("scores", vc.scoreCache.(cache.MetricsProvider))
		vc.metrics.RegisterCache("records", vc.recordCache.(cache.MetricsProvider))
	}

	return vc
}

// GetVEIDScore returns the VEID score for an account, using cache if available.
func (vc *VEIDCache) GetVEIDScore(ctx sdk.Context, address sdk.AccAddress) (uint32, bool) {
	if !vc.config.Enabled || vc.scoreCache == nil {
		return vc.keeper.GetVEIDScore(ctx, address)
	}

	key := address.String()
	bgCtx := context.Background()

	// Try to get from cache with loader
	cached, err := vc.scoreCache.GetOrLoad(bgCtx, key, func(_ context.Context, k string) (cachedScore, error) {
		score, found := vc.keeper.GetVEIDScore(ctx, address)
		return cachedScore{Score: score, Found: found}, nil
	})

	if err != nil {
		// Fallback to direct call on cache error
		return vc.keeper.GetVEIDScore(ctx, address)
	}

	return cached.Score, cached.Found
}

// GetScore returns the score, status, and whether a score was found.
func (vc *VEIDCache) GetScore(ctx sdk.Context, addr string) (uint32, types.AccountStatus, bool) {
	if !vc.config.Enabled || vc.scoreCache == nil {
		return vc.keeper.GetScore(ctx, addr)
	}

	bgCtx := context.Background()

	// Try to get from cache with loader
	cached, err := vc.scoreCache.GetOrLoad(bgCtx, addr, func(_ context.Context, k string) (cachedScore, error) {
		score, status, found := vc.keeper.GetScore(ctx, k)
		return cachedScore{Score: score, Status: status, Found: found}, nil
	})

	if err != nil {
		return vc.keeper.GetScore(ctx, addr)
	}

	return cached.Score, cached.Status, cached.Found
}

// GetIdentityRecord returns an identity record by address, using cache if available.
func (vc *VEIDCache) GetIdentityRecord(ctx sdk.Context, address sdk.AccAddress) (types.IdentityRecord, bool) {
	if !vc.config.Enabled || vc.recordCache == nil {
		return vc.keeper.GetIdentityRecord(ctx, address)
	}

	key := address.String()
	bgCtx := context.Background()

	// Try to get from cache
	record, err := vc.recordCache.Get(bgCtx, key)
	if err == nil {
		return record, true
	}

	// Cache miss - load from keeper
	record, found := vc.keeper.GetIdentityRecord(ctx, address)
	if found {
		// Store in cache
		_ = vc.recordCache.Set(bgCtx, key, record)
	}

	return record, found
}

// InvalidateScore removes a score from the cache.
// Should be called when a score is updated.
func (vc *VEIDCache) InvalidateScore(address sdk.AccAddress) {
	if !vc.config.Enabled || vc.scoreCache == nil {
		return
	}

	key := address.String()
	_ = vc.scoreCache.Delete(context.Background(), key)
}

// InvalidateRecord removes an identity record from the cache.
// Should be called when a record is updated.
func (vc *VEIDCache) InvalidateRecord(address sdk.AccAddress) {
	if !vc.config.Enabled || vc.recordCache == nil {
		return
	}

	key := address.String()
	_ = vc.recordCache.Delete(context.Background(), key)
}

// InvalidateAll clears all caches.
func (vc *VEIDCache) InvalidateAll() {
	if !vc.config.Enabled {
		return
	}

	bgCtx := context.Background()
	if vc.scoreCache != nil {
		_ = vc.scoreCache.Clear(bgCtx)
	}
	if vc.recordCache != nil {
		_ = vc.recordCache.Clear(bgCtx)
	}
}

// Stats returns cache statistics.
func (vc *VEIDCache) Stats() VEIDCacheStats {
	stats := VEIDCacheStats{
		Enabled: vc.config.Enabled,
	}

	if !vc.config.Enabled || vc.metrics == nil {
		return stats
	}

	if scoreStats, ok := vc.metrics.GetCacheStats("scores"); ok {
		stats.ScoreHits = scoreStats.Hits
		stats.ScoreMisses = scoreStats.Misses
		stats.ScoreSize = scoreStats.Size
		stats.ScoreHitRate = scoreStats.HitRate()
	}

	if recordStats, ok := vc.metrics.GetCacheStats("records"); ok {
		stats.RecordHits = recordStats.Hits
		stats.RecordMisses = recordStats.Misses
		stats.RecordSize = recordStats.Size
		stats.RecordHitRate = recordStats.HitRate()
	}

	return stats
}

// Close releases cache resources.
func (vc *VEIDCache) Close() error {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	if vc.closed {
		return nil
	}
	vc.closed = true

	if vc.metrics != nil {
		vc.metrics.Stop()
	}

	var errs []error
	if vc.scoreCache != nil {
		if err := vc.scoreCache.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if vc.recordCache != nil {
		if err := vc.recordCache.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("veid cache close errors: %v", errs)
	}
	return nil
}

// VEIDCacheStats contains VEID cache statistics.
type VEIDCacheStats struct {
	Enabled       bool    `json:"enabled"`
	ScoreHits     uint64  `json:"score_hits"`
	ScoreMisses   uint64  `json:"score_misses"`
	ScoreSize     int     `json:"score_size"`
	ScoreHitRate  float64 `json:"score_hit_rate"`
	RecordHits    uint64  `json:"record_hits"`
	RecordMisses  uint64  `json:"record_misses"`
	RecordSize    int     `json:"record_size"`
	RecordHitRate float64 `json:"record_hit_rate"`
}

// ============================================================================
// Cached Keeper Wrapper
// ============================================================================

// CachedKeeper wraps a Keeper with caching support.
// It implements the same interface as Keeper but uses cache for reads.
type CachedKeeper struct {
	Keeper
	cache *VEIDCache
}

// NewCachedKeeper creates a new cached keeper wrapper.
func NewCachedKeeper(keeper Keeper, config VEIDCacheConfig) *CachedKeeper {
	return &CachedKeeper{
		Keeper: keeper,
		cache:  NewVEIDCache(&keeper, config),
	}
}

// GetVEIDScore returns the VEID score using cache.
func (ck *CachedKeeper) GetVEIDScore(ctx sdk.Context, address sdk.AccAddress) (uint32, bool) {
	return ck.cache.GetVEIDScore(ctx, address)
}

// GetScore returns the score using cache.
func (ck *CachedKeeper) GetScore(ctx sdk.Context, addr string) (uint32, types.AccountStatus, bool) {
	return ck.cache.GetScore(ctx, addr)
}

// GetIdentityRecord returns an identity record using cache.
func (ck *CachedKeeper) GetIdentityRecord(ctx sdk.Context, address sdk.AccAddress) (types.IdentityRecord, bool) {
	return ck.cache.GetIdentityRecord(ctx, address)
}

// SetIdentityRecord stores a record and invalidates cache.
func (ck *CachedKeeper) SetIdentityRecord(ctx sdk.Context, record types.IdentityRecord) error {
	err := ck.Keeper.SetIdentityRecord(ctx, record)
	if err == nil {
		if addr, addrErr := sdk.AccAddressFromBech32(record.AccountAddress); addrErr == nil {
			ck.cache.InvalidateRecord(addr)
			ck.cache.InvalidateScore(addr)
		}
	}
	return err
}

// SetScore stores a score and invalidates cache.
func (ck *CachedKeeper) SetScore(ctx sdk.Context, accountAddr string, score uint32, modelVersion string) error {
	err := ck.Keeper.SetScore(ctx, accountAddr, score, modelVersion)
	if err == nil {
		if addr, addrErr := sdk.AccAddressFromBech32(accountAddr); addrErr == nil {
			ck.cache.InvalidateScore(addr)
		}
	}
	return err
}

// SetScoreWithDetails stores a score with details and invalidates cache.
func (ck *CachedKeeper) SetScoreWithDetails(ctx sdk.Context, accountAddr string, score uint32, details ScoreDetails) error {
	err := ck.Keeper.SetScoreWithDetails(ctx, accountAddr, score, details)
	if err == nil {
		if addr, addrErr := sdk.AccAddressFromBech32(accountAddr); addrErr == nil {
			ck.cache.InvalidateScore(addr)
			ck.cache.InvalidateRecord(addr)
		}
	}
	return err
}

// UpdateScore updates a score and invalidates cache.
func (ck *CachedKeeper) UpdateScore(ctx sdk.Context, address sdk.AccAddress, score uint32, scoreVersion string) error {
	err := ck.Keeper.UpdateScore(ctx, address, score, scoreVersion)
	if err == nil {
		ck.cache.InvalidateScore(address)
		ck.cache.InvalidateRecord(address)
	}
	return err
}

// CacheStats returns cache statistics.
func (ck *CachedKeeper) CacheStats() VEIDCacheStats {
	return ck.cache.Stats()
}

// Close closes the cache resources.
func (ck *CachedKeeper) Close() error {
	return ck.cache.Close()
}
