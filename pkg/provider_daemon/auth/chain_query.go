package auth

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	marketv1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"
	marketv1beta5 "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ChainQuerier fetches lease data from chain.
type ChainQuerier interface {
	GetLease(ctx context.Context, leaseID marketv1.LeaseID) (marketv1.Lease, error)
}

type MarketLeaseQuerierConfig struct {
	GRPCEndpoint string
	Timeout      time.Duration
}

type MarketLeaseQuerier struct {
	conn    *grpc.ClientConn
	timeout time.Duration
}

func NewMarketLeaseQuerier(cfg MarketLeaseQuerierConfig) (*MarketLeaseQuerier, error) {
	if cfg.GRPCEndpoint == "" {
		return nil, fmt.Errorf("grpc endpoint is required")
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}
	conn, err := grpc.NewClient(cfg.GRPCEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to grpc endpoint: %w", err)
	}
	return &MarketLeaseQuerier{conn: conn, timeout: cfg.Timeout}, nil
}

func (q *MarketLeaseQuerier) Close() error {
	if q.conn == nil {
		return nil
	}
	return q.conn.Close()
}

func (q *MarketLeaseQuerier) GetLease(ctx context.Context, leaseID marketv1.LeaseID) (marketv1.Lease, error) {
	client := marketv1beta5.NewQueryClient(q.conn)
	req := &marketv1beta5.QueryLeaseRequest{ID: leaseID}
	queryCtx, cancel := context.WithTimeout(ctx, q.timeout)
	defer cancel()
	resp, err := client.Lease(queryCtx, req)
	if err != nil {
		return marketv1.Lease{}, err
	}
	return resp.Lease, nil
}

type leaseOwnerCacheEntry struct {
	owner   string
	expires time.Time
}

// LeaseOwnerCache stores lease owner lookups for a short TTL.
type LeaseOwnerCache struct {
	mu     sync.RWMutex
	ttl    time.Duration
	owners map[string]leaseOwnerCacheEntry
}

func NewLeaseOwnerCache(ttl time.Duration) *LeaseOwnerCache {
	if ttl == 0 {
		ttl = 15 * time.Minute
	}
	return &LeaseOwnerCache{
		ttl:    ttl,
		owners: make(map[string]leaseOwnerCacheEntry),
	}
}

func (c *LeaseOwnerCache) Get(leaseID string) (string, bool) {
	c.mu.RLock()
	entry, ok := c.owners[leaseID]
	c.mu.RUnlock()
	if !ok {
		return "", false
	}
	if time.Now().After(entry.expires) {
		c.mu.Lock()
		delete(c.owners, leaseID)
		c.mu.Unlock()
		return "", false
	}
	return entry.owner, true
}

func (c *LeaseOwnerCache) Set(leaseID, owner string) {
	if leaseID == "" || owner == "" {
		return
	}
	c.mu.Lock()
	c.owners[leaseID] = leaseOwnerCacheEntry{
		owner:   owner,
		expires: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()
}

func parseLeaseID(raw string) (marketv1.LeaseID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return marketv1.LeaseID{}, fmt.Errorf("lease id is required")
	}
	parts := strings.Split(raw, "/")
	if len(parts) > 0 && parts[0] == "lease" {
		parts = parts[1:]
	}
	return marketv1.ParseLeasePath(parts)
}
