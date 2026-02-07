package auth

import (
	"context"
	"errors"
)

// ErrLeaseQueryUnavailable indicates lease data cannot be fetched.
var ErrLeaseQueryUnavailable = errors.New("lease query unavailable")

// ErrLeaseNotFound indicates lease does not exist.
var ErrLeaseNotFound = errors.New("lease not found")

// Lease contains ownership info used for authorization checks.
type Lease struct {
	ID    string
	Owner string
}

// ChainQuerier retrieves lease ownership information.
type ChainQuerier interface {
	GetLease(ctx context.Context, leaseID string) (*Lease, error)
}

// NoopChainQuerier returns unavailable errors for lease queries.
type NoopChainQuerier struct{}

// GetLease returns ErrLeaseQueryUnavailable.
func (NoopChainQuerier) GetLease(_ context.Context, _ string) (*Lease, error) {
	return nil, ErrLeaseQueryUnavailable
}
