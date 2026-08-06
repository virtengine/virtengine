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
	timeNow func() time.Time
}

// NewInMemoryStore creates a new in-memory notification store.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		records: make(map[string][]Notification),
		timeNow: func() time.Time {
			return time.Now().UTC()
		},
	}
}

// Add stores a notification.
func (s *InMemoryStore) Add(_ context.Context, notif Notification) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cloned := cloneNotification(notif)
	if cloned.CreatedAt.IsZero() {
		cloned.CreatedAt = s.timeNow()
	}
	if cloned.ID != "" {
		for i, existing := range s.records[cloned.UserAddress] {
			if existing.ID != cloned.ID {
				continue
			}
			if cloned.ReadAt == nil {
				cloned.ReadAt = cloneTimePtr(existing.ReadAt)
			}
			if cloned.CreatedAt.IsZero() {
				cloned.CreatedAt = existing.CreatedAt
			}
			s.records[cloned.UserAddress][i] = cloned
			return nil
		}
	}
	s.records[cloned.UserAddress] = append(s.records[cloned.UserAddress], cloned)
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

	filtered := make([]Notification, 0, len(records))
	for _, notif := range records {
		if opts.UnreadOnly && notif.ReadAt != nil {
			continue
		}
		filtered = append(filtered, cloneNotification(notif))
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

	now := s.timeNow()
	for i, notif := range s.records[userAddr] {
		for _, id := range ids {
			if notif.ID == id {
				notif.ReadAt = cloneTimePtr(&now)
				s.records[userAddr][i] = notif
				break
			}
		}
	}

	return nil
}
