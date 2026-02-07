/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

package notifications

import (
	"context"
	"sync"
)

// InAppHub broadcasts notifications to subscribers.
type InAppHub struct {
	mu          sync.RWMutex
	subscribers map[string][]chan Notification
}

// NewInAppHub creates a new hub.
func NewInAppHub() *InAppHub {
	return &InAppHub{
		subscribers: make(map[string][]chan Notification),
	}
}

// Publish sends a notification to subscribers.
func (h *InAppHub) Publish(_ context.Context, notif Notification) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, ch := range h.subscribers[notif.UserAddress] {
		select {
		case ch <- notif:
		default:
		}
	}
	return nil
}

// Subscribe registers a subscriber channel for a user.
func (h *InAppHub) Subscribe(userAddr string, buffer int) (<-chan Notification, func()) {
	if buffer <= 0 {
		buffer = 16
	}
	ch := make(chan Notification, buffer)

	h.mu.Lock()
	h.subscribers[userAddr] = append(h.subscribers[userAddr], ch)
	h.mu.Unlock()

	cancel := func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		list := h.subscribers[userAddr]
		for i, sub := range list {
			if sub == ch {
				list = append(list[:i], list[i+1:]...)
				break
			}
		}
		h.subscribers[userAddr] = list
		close(ch)
	}

	return ch, cancel
}
