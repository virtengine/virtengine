package contracts

import (
	"errors"
	"fmt"
	"sync"
)

var ErrRevisionRollback = errors.New("persisted revision is behind monotonic anchor")

// RevisionAnchor is backend-neutral monotonic storage outside the protected
// payload. Implementations must atomically compare and advance each namespace.
type RevisionAnchor interface {
	Current(namespace string) (uint64, error)
	CompareAndAdvance(namespace string, expected, next uint64) error
	Replayable() bool
	Local() bool
}

// FixtureSecurityOptions contains explicit fixture-only security exceptions.
type FixtureSecurityOptions struct {
	UnsafeWindowsDevelopment bool
}

// ProcessRevisionAnchor is a non-replayable process/test anchor. It detects
// fixture rollback while the anchor instance lives, but is not production-safe.
type ProcessRevisionAnchor struct {
	mu        sync.Mutex
	revisions map[string]uint64
}

func NewProcessRevisionAnchor() *ProcessRevisionAnchor {
	return &ProcessRevisionAnchor{revisions: make(map[string]uint64)}
}

func (a *ProcessRevisionAnchor) Current(namespace string) (uint64, error) {
	if namespace == "" {
		return 0, errors.New("revision anchor namespace is required")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.revisions[namespace], nil
}

func (a *ProcessRevisionAnchor) CompareAndAdvance(namespace string, expected, next uint64) error {
	if namespace == "" || next != expected+1 {
		return errors.New("invalid revision anchor transition")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	current := a.revisions[namespace]
	if current != expected {
		return fmt.Errorf("%w: anchor has %d, expected %d", ErrRevisionRollback, current, expected)
	}
	a.revisions[namespace] = next
	return nil
}

func (*ProcessRevisionAnchor) Replayable() bool { return false }
func (*ProcessRevisionAnchor) Local() bool      { return true }
