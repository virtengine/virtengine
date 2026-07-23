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
	"sort"
	"strings"
	"sync"
	"time"
)

var ErrProviderMutationNotFound = errors.New("provider mutation not found")

type providerMutationFileState struct {
	SchemaVersion uint32                               `json:"schema_version"`
	Items         map[string]*ProviderMutationEnvelope `json:"items"`
}

// FileProviderMutationStore keeps the existing cross-platform process lock and
// atomic replacement semantics while exposing the pluggable Task 85C store API.
type FileProviderMutationStore struct {
	path  string
	mu    sync.RWMutex
	state providerMutationFileState
	lock  *txSubmissionQueuePathLock
	open  bool
}

func NewFileProviderMutationStore(path string) (*FileProviderMutationStore, error) {
	if path == "" {
		return nil, fmt.Errorf("mutation queue state path is required")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve mutation queue path: %w", err)
	}
	if err := validateStatePath(absolute); err != nil {
		return nil, fmt.Errorf("invalid mutation queue state path: %w", err)
	}
	return &FileProviderMutationStore{path: filepath.Clean(absolute)}, nil
}

func (s *FileProviderMutationStore) Open(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.open {
		return nil
	}
	lock, err := claimTxSubmissionQueuePath(s.path)
	if err != nil {
		return err
	}
	state := providerMutationFileState{SchemaVersion: providerMutationSchemaVersion, Items: make(map[string]*ProviderMutationEnvelope)}
	data, readErr := os.ReadFile(s.path) // #nosec G304 -- constructor validates the state path.
	if readErr == nil {
		if err := json.Unmarshal(data, &state); err != nil {
			lock.release()
			return fmt.Errorf("decode mutation queue: %w", err)
		}
		if state.SchemaVersion != providerMutationSchemaVersion {
			lock.release()
			return fmt.Errorf("unsupported mutation queue schema %d", state.SchemaVersion)
		}
		if state.Items == nil {
			state.Items = make(map[string]*ProviderMutationEnvelope)
		}
	} else if !os.IsNotExist(readErr) {
		lock.release()
		return fmt.Errorf("read mutation queue: %w", readErr)
	}
	s.state = state
	s.lock = lock
	s.open = true
	return nil
}

func (s *FileProviderMutationStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.open {
		return nil
	}
	s.lock.release()
	s.lock = nil
	s.open = false
	return nil
}

func cloneMutationEnvelope(item *ProviderMutationEnvelope) *ProviderMutationEnvelope {
	if item == nil {
		return nil
	}
	data, err := json.Marshal(item)
	if err != nil {
		return nil
	}
	var clone ProviderMutationEnvelope
	if err := json.Unmarshal(data, &clone); err != nil {
		return nil
	}
	return &clone
}

func (s *FileProviderMutationStore) PutIfAbsent(_ context.Context, item *ProviderMutationEnvelope) (*ProviderMutationEnvelope, bool, error) {
	if item == nil || item.ID == "" || item.IdempotencyKey == "" {
		return nil, false, fmt.Errorf("mutation identity is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.open {
		return nil, false, ErrProviderMutationUnavailable
	}
	for _, existing := range s.state.Items {
		if existing.ID == item.ID {
			return cloneMutationEnvelope(existing), true, nil
		}
		if existing.IdempotencyKey != item.IdempotencyKey {
			continue
		}
		if strings.EqualFold(existing.MessageDigest, item.MessageDigest) {
			return cloneMutationEnvelope(existing), true, nil
		}
		if !providerMutationTerminalState(existing.State) {
			return cloneMutationEnvelope(existing), true, nil
		}
	}
	s.state.Items[item.ID] = cloneMutationEnvelope(item)
	if err := s.saveLocked(); err != nil {
		delete(s.state.Items, item.ID)
		return nil, false, err
	}
	return cloneMutationEnvelope(item), false, nil
}

func (s *FileProviderMutationStore) Get(_ context.Context, id string) (*ProviderMutationEnvelope, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.open {
		return nil, ErrProviderMutationUnavailable
	}
	item, ok := s.state.Items[id]
	if !ok {
		return nil, ErrProviderMutationNotFound
	}
	return cloneMutationEnvelope(item), nil
}

func (s *FileProviderMutationStore) Update(_ context.Context, id string, fn func(*ProviderMutationEnvelope) error) (*ProviderMutationEnvelope, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.open {
		return nil, ErrProviderMutationUnavailable
	}
	item, ok := s.state.Items[id]
	if !ok {
		return nil, ErrProviderMutationNotFound
	}
	before := cloneMutationEnvelope(item)
	candidate := cloneMutationEnvelope(item)
	if err := fn(candidate); err != nil {
		return nil, err
	}
	candidate.UpdatedAt = time.Now().UTC()
	candidate.SchemaVersion = providerMutationSchemaVersion
	s.state.Items[id] = candidate
	if err := s.saveLocked(); err != nil {
		s.state.Items[id] = before
		return nil, err
	}
	return cloneMutationEnvelope(candidate), nil
}

func (s *FileProviderMutationStore) List(_ context.Context) ([]*ProviderMutationEnvelope, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.open {
		return nil, ErrProviderMutationUnavailable
	}
	items := make([]*ProviderMutationEnvelope, 0, len(s.state.Items))
	for _, item := range s.state.Items {
		items = append(items, cloneMutationEnvelope(item))
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items, nil
}

func (s *FileProviderMutationStore) saveLocked() error {
	data, err := json.MarshalIndent(s.state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode mutation queue: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".provider-mutation-*.tmp")
	if err != nil {
		return fmt.Errorf("create mutation queue temporary file: %w", err)
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
	if err := atomicReplaceFile(tmpPath, s.path); err != nil {
		return fmt.Errorf("replace mutation queue: %w", err)
	}
	dir, err := os.Open(filepath.Dir(s.path)) // #nosec G304 -- constructor validates the state path.
	if err == nil {
		_ = dir.Sync()
		_ = dir.Close()
	}
	return nil
}

var _ ProviderMutationStore = (*FileProviderMutationStore)(nil)
