//go:build e2e.integration

package email

import (
	"context"
	"strings"
	"testing"
)

func TestEmailServiceE2EBuildsDeliverableDigest(t *testing.T) {
	renderer, err := NewDefaultRenderer()
	if err != nil {
		t.Fatalf("renderer: %v", err)
	}
	sender := &memorySender{}
	service := NewService(renderer, sender, "noreply@virtengine.com", "")

	if err := service.SendTemplate(context.Background(), "user@example.com", "", TemplateWeeklyDigest, WeeklyDigestData{
		Summary: "Two updates are waiting for review.",
		Items: []DigestItem{
			{Title: "Order update", Body: "Provisioned", Link: "https://virtengine.com/orders/1"},
			{Title: "VEID status", Body: "Approved", Link: "https://virtengine.com/account/veid"},
		},
		DashboardURL: "https://virtengine.com/dashboard",
	}); err != nil {
		t.Fatalf("send template: %v", err)
	}

	if sender.count != 1 {
		t.Fatalf("expected exactly one delivered message, got %d", sender.count)
	}
	if sender.last.Headers != nil && sender.last.Headers["List-Unsubscribe"] != "" {
		t.Fatalf("expected no unsubscribe header when base URL is empty, got %+v", sender.last.Headers)
	}
	if !strings.Contains(sender.last.Text, "Order update") || !strings.Contains(sender.last.HTML, "VEID status") {
		t.Fatalf("expected digest content in both text and HTML outputs, got text=%q html=%q", sender.last.Text, sender.last.HTML)
	}
}
