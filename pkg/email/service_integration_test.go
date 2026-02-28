//go:build integration

package email

import (
	"context"
	"strings"
	"testing"
)

func TestEmailServiceIntegrationProducesSafeMessage(t *testing.T) {
	renderer, err := NewDefaultRenderer()
	if err != nil {
		t.Fatalf("renderer: %v", err)
	}
	sender := &memorySender{}
	service := NewService(renderer, sender, "noreply@virtengine.com", "https://virtengine.com/unsubscribe")

	if err := service.SendTemplate(context.Background(), "user@example.com", "token-abc", TemplateSecurityAlert, SecurityAlertData{
		Title:        "Credential update\r\nBcc:evil@example.com",
		Description:  "<script>alert('x')</script>",
		Timestamp:    "2026-04-11T09:00:00Z",
		IPAddress:    "203.0.113.10",
		DashboardURL: "https://virtengine.com/account",
	}); err != nil {
		t.Fatalf("send template: %v", err)
	}

	if strings.ContainsAny(sender.last.Subject, "\r\n") {
		t.Fatalf("expected sanitized subject, got %q", sender.last.Subject)
	}
	if strings.Contains(sender.last.HTML, "<script>") || !strings.Contains(sender.last.HTML, "&lt;script&gt;") {
		t.Fatalf("expected escaped HTML body, got %q", sender.last.HTML)
	}
	if sender.last.Headers["List-Unsubscribe"] == "" {
		t.Fatalf("expected unsubscribe header")
	}
}
