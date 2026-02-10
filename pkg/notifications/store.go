/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

package notifications

import (
	"context"
	"sort"
	"sync"
	"time"
)

// InMemoryStore is an in-memory notification store for local development.
type InMemoryStore struct {
	mu      sync.RWMutex
	records map[string][]Notification
}

// NewInMemoryStore creates a new in-memory notification store.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		records: make(map[string][]Notification),
	}
}

// Add stores a notification.
func (s *InMemoryStore) Add(_ context.Context, notif Notification) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.records[notif.UserAddress] = append(s.records[notif.UserAddress], notif)
	return nil
}

// List retrieves notifications for a user.
func (s *InMemoryStore) List(_ context.Context, userAddr string, opts ListOptions) ([]Notification, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	records := append([]Notification{}, s.records[userAddr]...)
	sort.Slice(records, func(i, j int) bool {
		return records[i].CreatedAt.After(records[j].CreatedAt)
	})

	filtered := records[:0]
	for _, notif := range records {
		if opts.UnreadOnly && notif.ReadAt != nil {
			continue
		}
		filtered = append(filtered, notif)
	}

	if opts.Limit > 0 && len(filtered) > opts.Limit {
		filtered = filtered[:opts.Limit]
	}

	return filtered, nil
}

// MarkRead marks notifications as read.
func (s *InMemoryStore) MarkRead(_ context.Context, userAddr string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	for i, notif := range s.records[userAddr] {
		for _, id := range ids {
			if notif.ID == id {
				notif.ReadAt = &now
				s.records[userAddr][i] = notif
				break
			}
		}
	}

	return nil
}
