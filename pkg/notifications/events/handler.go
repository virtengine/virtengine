/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

package events

import (
	"context"
	"fmt"

	"github.com/virtengine/virtengine/pkg/notifications"
)

// Handler converts chain or service events into notifications.
type Handler struct {
	service notifications.Service
}

// NewHandler creates a new notification event handler.
func NewHandler(service notifications.Service) *Handler {
	return &Handler{service: service}
}

// HandleVEIDStatus emits a VEID status update notification.
func (h *Handler) HandleVEIDStatus(ctx context.Context, userAddr, status string) error {
	return h.service.Send(ctx, notifications.Notification{
		UserAddress: userAddr,
		Type:        notifications.NotificationTypeVEIDStatus,
		Title:       "VEID verification update",
		Body:        fmt.Sprintf("Your VEID verification status is now %s.", status),
		Data: map[string]string{
			"status": status,
		},
	})
}

// HandleOrderUpdate emits an order update notification.
func (h *Handler) HandleOrderUpdate(ctx context.Context, userAddr, orderID, status string) error {
	return h.service.Send(ctx, notifications.Notification{
		UserAddress: userAddr,
		Type:        notifications.NotificationTypeOrderUpdate,
		Title:       "Order update",
		Body:        fmt.Sprintf("Order %s is now %s.", orderID, status),
		Data: map[string]string{
			"order_id": orderID,
			"status":   status,
		},
	})
}
