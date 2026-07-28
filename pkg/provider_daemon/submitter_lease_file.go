// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package provider_daemon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const durableSubmitterLeaseSchemaVersion uint32 = 1

var ErrSubmitterLeaseNotHeld = errors.New("submitter lease is not held")

type durableSubmitterLeaseRecord struct {
	Name      string    `json:"name"`
	OwnerID   string    `json:"owner_id"`
	Token     uint64    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type durableSubmitterLeaseState struct {
	SchemaVersion uint32                                  `json:"schema_version"`
	NextToken     uint64                                  `json:"next_token"`
	Leases        map[string]*durableSubmitterLeaseRecord `json:"leases"`
}

// FileSubmitterLease is a durable cross-process lease and monotonically
// increasing fencing-token allocator. Every operation takes an OS lock, reloads
// current shared state, and atomically replaces it before releasing the lock.
// Its path must be on a filesystem whose file-lock and rename semantics are
// shared by every replica; production startup rejects the process-local lease.
type FileSubmitterLease struct {
	path    string
	ownerID string
	now     func() time.Time
	mu      sync.Mutex
}

func NewFileSubmitterLease(path, ownerID string) (*FileSubmitterLease, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("submitter lease state path is required")
	}
	if strings.TrimSpace(ownerID) == "" {
		return nil, fmt.Errorf("submitter lease owner ID is required")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve submitter lease path: %w", err)
	}
	if err := validateStatePath(absolute); err != nil {
		return nil, fmt.Errorf("invalid submitter lease path: %w", err)
	}
	return &FileSubmitterLease{path: filepath.Clean(absolute), ownerID: ownerID, now: func() time.Time { return time.Now().UTC() }}, nil
}

func (l *FileSubmitterLease) Acquire(ctx context.Context, name string, ttl time.Duration) (uint64, error) {
	if err := validateLeaseArguments(ctx, name, ttl); err != nil {
		return 0, err
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	lock, state, err := l.loadLocked()
	if err != nil {
		return 0, err
	}
	defer lock.release()
	now := l.now()
	if current, ok := state.Leases[name]; ok && now.Before(current.ExpiresAt) {
		return 0, fmt.Errorf("submitter lease %s already held", name)
	}
	if state.NextToken == ^uint64(0) {
		return 0, fmt.Errorf("submitter lease fencing token exhausted")
	}
	state.NextToken++
	if state.NextToken == 0 {
		return 0, fmt.Errorf("submitter lease fencing token exhausted")
	}
	state.Leases[name] = &durableSubmitterLeaseRecord{Name: name, OwnerID: l.ownerID, Token: state.NextToken, ExpiresAt: now.Add(ttl), UpdatedAt: now}
	if err := l.saveLocked(state); err != nil {
		return 0, err
	}
	return state.NextToken, nil
}

func (l *FileSubmitterLease) Renew(ctx context.Context, name string, token uint64, ttl time.Duration) error {
	if err := validateLeaseArguments(ctx, name, ttl); err != nil {
		return err
	}
	if token == 0 {
		return ErrSubmitterLeaseNotHeld
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	lock, state, err := l.loadLocked()
	if err != nil {
		return err
	}
	defer lock.release()
	now := l.now()
	current, ok := state.Leases[name]
	if !ok || current.Token != token || !now.Before(current.ExpiresAt) {
		return fmt.Errorf("%w: %s", ErrSubmitterLeaseNotHeld, name)
	}
	current.OwnerID = l.ownerID
	current.ExpiresAt = now.Add(ttl)
	current.UpdatedAt = now
	return l.saveLocked(state)
}

func (l *FileSubmitterLease) Release(ctx context.Context, name string, token uint64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if strings.TrimSpace(name) == "" || token == 0 {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	lock, state, err := l.loadLocked()
	if err != nil {
		return err
	}
	defer lock.release()
	current, ok := state.Leases[name]
	if !ok || current.OwnerID != l.ownerID || current.Token != token {
		return nil
	}
	delete(state.Leases, name)
	return l.saveLocked(state)
}

func (l *FileSubmitterLease) Held(ctx context.Context, name string, token uint64) bool {
	if ctx.Err() != nil || strings.TrimSpace(name) == "" || token == 0 {
		return false
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	lock, state, err := l.loadLocked()
	if err != nil {
		return false
	}
	defer lock.release()
	current, ok := state.Leases[name]
	return ok && current.Token == token && l.now().Before(current.ExpiresAt)
}

// Snapshot returns the monotonic allocator and a copy of current lease records
// for backup verification and operational readiness evidence.
func (l *FileSubmitterLease) Snapshot(ctx context.Context) (uint64, map[string]uint64, error) {
	if err := ctx.Err(); err != nil {
		return 0, nil, err
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	lock, state, err := l.loadLocked()
	if err != nil {
		return 0, nil, err
	}
	defer lock.release()
	leases := make(map[string]uint64, len(state.Leases))
	for name, record := range state.Leases {
		leases[name] = record.Token
	}
	return state.NextToken, leases, nil
}

func validateLeaseArguments(ctx context.Context, name string, ttl time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("submitter lease name is required")
	}
	if ttl <= 0 {
		return fmt.Errorf("submitter lease TTL must be positive")
	}
	return nil
}

func (l *FileSubmitterLease) loadLocked() (*txSubmissionQueuePathLock, *durableSubmitterLeaseState, error) {
	if err := os.MkdirAll(filepath.Dir(l.path), 0o700); err != nil {
		return nil, nil, fmt.Errorf("create submitter lease directory: %w", err)
	}
	lock, err := claimTxSubmissionQueuePath(l.path)
	if err != nil {
		return nil, nil, err
	}
	state := &durableSubmitterLeaseState{SchemaVersion: durableSubmitterLeaseSchemaVersion, Leases: make(map[string]*durableSubmitterLeaseRecord)}
	data, err := os.ReadFile(l.path) // #nosec G304 -- path validated by constructor.
	if errors.Is(err, os.ErrNotExist) {
		return lock, state, nil
	}
	if err != nil {
		lock.release()
		return nil, nil, fmt.Errorf("read submitter lease state: %w", err)
	}
	if err := json.Unmarshal(data, state); err != nil {
		lock.release()
		return nil, nil, fmt.Errorf("decode submitter lease state: %w", err)
	}
	if state.SchemaVersion != durableSubmitterLeaseSchemaVersion {
		lock.release()
		return nil, nil, fmt.Errorf("unsupported submitter lease schema %d", state.SchemaVersion)
	}
	if state.Leases == nil {
		state.Leases = make(map[string]*durableSubmitterLeaseRecord)
	}
	return lock, state, nil
}

func (l *FileSubmitterLease) saveLocked(state *durableSubmitterLeaseState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode submitter lease state: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(l.path), ".submitter-lease-*.tmp")
	if err != nil {
		return fmt.Errorf("create submitter lease temporary file: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := atomicReplaceFile(tmpPath, l.path); err != nil {
		return fmt.Errorf("replace submitter lease state: %w", err)
	}
	if dir, err := os.Open(filepath.Dir(l.path)); err == nil { // #nosec G304 -- path validated by constructor.
		_ = dir.Sync()
		_ = dir.Close()
	}
	return nil
}

var _ SubmitterLease = (*FileSubmitterLease)(nil)
