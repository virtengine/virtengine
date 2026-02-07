package events

import (
	"context"
	"fmt"

	"github.com/virtengine/virtengine/pkg/notifications"
	"github.com/virtengine/virtengine/pubsub"
)

// Handler forwards events to the notification service.
type Handler struct {
	service *notifications.Service
}

// NewHandler creates a new event handler.
func NewHandler(service *notifications.Service) *Handler {
	return &Handler{service: service}
}

// HandleEvent handles a single event.
func (h *Handler) HandleEvent(ctx context.Context, event Event) error {
	if h.service == nil {
		return fmt.Errorf("notifications service not configured")
	}
	notif := notifications.Notification{
		UserAddress:           event.UserAddress,
		Type:                  event.Type,
		Title:                 event.Title,
		Body:                  event.Body,
		Data:                  event.Data,
		Channels:              event.Channels,
		Silent:                event.Silent,
		Topics:                event.Topics,
		AllowDuringQuietHours: event.AllowDuringQuietHours,
	}
	return h.service.Send(ctx, notif)
}

// Listen subscribes to a pubsub channel and forwards events.
func (h *Handler) Listen(ctx context.Context, sub pubsub.Subscriber) error {
	if h.service == nil {
		return fmt.Errorf("notifications service not configured")
	}
	if sub == nil {
		return fmt.Errorf("subscriber required")
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-sub.Done():
			return nil
		case ev, ok := <-sub.Events():
			if !ok {
				return nil
			}
			event, ok := ev.(Event)
			if !ok {
				continue
			}
			if err := h.HandleEvent(ctx, event); err != nil {
				return err
			}
		}
	}
}
