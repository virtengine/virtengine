/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

package notifications

import (
	"context"
	"errors"
	"sync"
)

var ErrPreferencesNotFound = errors.New("notification preferences not found")

// InMemoryPreferencesStore stores preferences in memory.
type InMemoryPreferencesStore struct {
	mu    sync.RWMutex
	prefs map[string]Preferences
}

// NewInMemoryPreferencesStore creates a new preference store.
func NewInMemoryPreferencesStore() *InMemoryPreferencesStore {
	return &InMemoryPreferencesStore{
		prefs: make(map[string]Preferences),
	}
}

// Get returns preferences for a user.
func (s *InMemoryPreferencesStore) Get(_ context.Context, userAddr string) (Preferences, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	prefs, ok := s.prefs[userAddr]
	if !ok {
		return Preferences{}, ErrPreferencesNotFound
	}

	return prefs, nil
}

// Put stores preferences for a user.
func (s *InMemoryPreferencesStore) Put(_ context.Context, prefs Preferences) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.prefs[prefs.UserAddress] = prefs
	return nil
}
