package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/virtengine/virtengine/pkg/observability"
	marketv1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"
	marketv1beta5 "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const defaultLeaseCacheTTL = 15 * time.Minute

// LeaseOwnerQuerier provides lease owner lookup.
type LeaseOwnerQuerier interface {
	LeaseOwner(ctx context.Context, leaseID marketv1.LeaseID) (string, error)
}

// LeaseOwnerCache caches lease owner lookups.
type LeaseOwnerCache struct {
	mu    sync.RWMutex
	ttl   time.Duration
	items map[string]leaseOwnerEntry
}

type leaseOwnerEntry struct {
	owner     string
	expiresAt time.Time
}

// NewLeaseOwnerCache creates a cache with TTL.
func NewLeaseOwnerCache(ttl time.Duration) *LeaseOwnerCache {
	if ttl <= 0 {
		ttl = defaultLeaseCacheTTL
	}
	return &LeaseOwnerCache{
		ttl:   ttl,
		items: make(map[string]leaseOwnerEntry),
	}
}

func (c *LeaseOwnerCache) Get(leaseID string) (string, bool) {
	c.mu.RLock()
	entry, ok := c.items[leaseID]
	c.mu.RUnlock()
	if !ok {
		return "", false
	}
	if time.Now().After(entry.expiresAt) {
		c.mu.Lock()
		delete(c.items, leaseID)
		c.mu.Unlock()
		return "", false
	}
	return entry.owner, true
}

func (c *LeaseOwnerCache) Set(leaseID, owner string) {
	c.mu.Lock()
	c.items[leaseID] = leaseOwnerEntry{
		owner:     owner,
		expiresAt: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()
}

// CachedLeaseOwnerQuerier adds caching to a LeaseOwnerQuerier.
type CachedLeaseOwnerQuerier struct {
	base  LeaseOwnerQuerier
	cache *LeaseOwnerCache
}

// NewCachedLeaseOwnerQuerier wraps a LeaseOwnerQuerier with cache.
func NewCachedLeaseOwnerQuerier(base LeaseOwnerQuerier, cache *LeaseOwnerCache) *CachedLeaseOwnerQuerier {
	if cache == nil {
		cache = NewLeaseOwnerCache(defaultLeaseCacheTTL)
	}
	return &CachedLeaseOwnerQuerier{base: base, cache: cache}
}

// LeaseOwner returns cached lease owner if available.
func (c *CachedLeaseOwnerQuerier) LeaseOwner(ctx context.Context, leaseID marketv1.LeaseID) (string, error) {
	key := leaseID.String()
	if owner, ok := c.cache.Get(key); ok {
		return owner, nil
	}
	owner, err := c.base.LeaseOwner(ctx, leaseID)
	if err != nil {
		return "", err
	}
	c.cache.Set(key, owner)
	return owner, nil
}

// MarketLeaseQuerier queries chain for lease ownership.
type MarketLeaseQuerier struct {
	conn    *grpc.ClientConn
	client  marketv1beta5.QueryClient
	timeout time.Duration
}

// NewMarketLeaseQuerier creates a lease querier using gRPC endpoint.
func NewMarketLeaseQuerier(grpcEndpoint string, timeout time.Duration) (*MarketLeaseQuerier, error) {
	if grpcEndpoint == "" {
		return nil, errors.New("grpc endpoint is required")
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	conn, err := grpc.NewClient(
		grpcEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(observability.GRPCClientStatsHandler()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC endpoint: %w", err)
	}

	return &MarketLeaseQuerier{
		conn:    conn,
		client:  marketv1beta5.NewQueryClient(conn),
		timeout: timeout,
	}, nil
}

// LeaseOwner queries chain for lease owner.
func (q *MarketLeaseQuerier) LeaseOwner(ctx context.Context, leaseID marketv1.LeaseID) (string, error) {
	if q.client == nil {
		return "", errors.New("chain query not configured")
	}
	reqCtx, cancel := context.WithTimeout(ctx, q.timeout)
	defer cancel()

	resp, err := q.client.Lease(reqCtx, &marketv1beta5.QueryLeaseRequest{ID: leaseID})
	if err != nil {
		return "", err
	}
	return resp.Lease.ID.Owner, nil
}

// Close closes the gRPC connection.
func (q *MarketLeaseQuerier) Close() error {
	if q.conn != nil {
		return q.conn.Close()
	}
	return nil
}

// ParseLeaseID parses a string lease identifier into a LeaseID.
func ParseLeaseID(leaseID string) (marketv1.LeaseID, error) {
	trimmed := strings.Trim(leaseID, "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) >= 1 && (parts[0] == "lease" || parts[0] == "leases") {
		parts = parts[1:]
	}
	return marketv1.ParseLeasePath(parts)
}
