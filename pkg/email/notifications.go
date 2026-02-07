package email

import (
	"context"
	"fmt"

	"github.com/virtengine/virtengine/pkg/email/templates"
	"github.com/virtengine/virtengine/pkg/notifications"
)

// NotificationProvider renders notification emails and sends them via Sender.
type NotificationProvider struct {
	Sender   Sender
	FromName string
}

// Send renders and sends an email for the notification.
func (p *NotificationProvider) Send(ctx context.Context, notification notifications.Notification) error {
	if p == nil || p.Sender == nil {
		return fmt.Errorf("email sender not configured")
	}

	subject := notification.Title
	if subject == "" {
		subject = defaultSubject(notification.Type)
	}

	htmlBody, err := renderNotification(notification)
	if err != nil {
		return err
	}

	message := Message{
		To:      notification.UserAddress,
		Subject: subject,
		HTML:    htmlBody,
	}

	return p.Sender.Send(ctx, message)
}

func renderNotification(notification notifications.Notification) (string, error) {
	switch notification.Type {
	case notifications.NotificationTypeOrderUpdate:
		return templates.RenderOrderConfirmation(templates.OrderConfirmationData{
			OrderID:      notification.Data["order_id"],
			ProviderName: notification.Data["provider_name"],
			ServiceName:  notification.Data["service_name"],
			Amount:       notification.Data["amount"],
			Currency:     notification.Data["currency"],
			DashboardURL: notification.Data["dashboard_url"],
		})
	case notifications.NotificationTypeVEIDStatus:
		return templates.RenderVEIDStatus(templates.VEIDStatusData{
			VEID:      notification.Data["veid"],
			Status:    notification.Data["status"],
			ActionURL: notification.Data["action_url"],
		})
	case notifications.NotificationTypeSecurityAlert:
		return templates.RenderSecurityAlert(templates.SecurityAlertData{
			Event:      notification.Data["event"],
			Location:   notification.Data["location"],
			OccurredAt: notification.Data["occurred_at"],
			ActionURL:  notification.Data["action_url"],
		})
	default:
		return templates.RenderWeeklyDigest(templates.WeeklyDigestData{
			Username:  notification.UserAddress,
			Summary:   []string{notification.Body},
			ActionURL: notification.Data["action_url"],
		})
	}
}

func defaultSubject(notificationType notifications.NotificationType) string {
	switch notificationType {
	case notifications.NotificationTypeOrderUpdate:
		return "Order Update"
	case notifications.NotificationTypeVEIDStatus:
		return "VEID Status Update"
	case notifications.NotificationTypeSecurityAlert:
		return "Security Alert"
	case notifications.NotificationTypeEscrowDeposit:
		return "Escrow Update"
	case notifications.NotificationTypeProviderAlert:
		return "Provider Alert"
	default:
		return "VirtEngine Notification"
	}
}
