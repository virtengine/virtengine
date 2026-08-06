package email

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type memorySender struct {
	last  EmailMessage
	count int
}

func (m *memorySender) Send(_ context.Context, message EmailMessage) error {
	m.last = message
	m.count++
	return nil
}

func TestRenderTemplates(t *testing.T) {
	renderer, err := NewDefaultRenderer()
	if err != nil {
		t.Fatalf("renderer: %v", err)
	}

	out, err := renderer.Render(TemplateOrderConfirmation, OrderConfirmationData{
		OrderID:      "order-1",
		ProviderName: "CloudCore",
		ServiceName:  "GPU",
		Amount:       "10",
		Currency:     "USD",
		DashboardURL: "https://virtengine.com/dashboard",
		Unsubscribe:  "https://virtengine.com/unsub",
	})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(out.HTML, "order-1") {
		t.Fatalf("expected rendered order id")
	}
	if !strings.Contains(out.Text, "CloudCore") {
		t.Fatalf("expected rendered provider")
	}
}

func TestRenderTemplatesEscapesHTMLAndSanitizesSubject(t *testing.T) {
	renderer, err := NewDefaultRenderer()
	if err != nil {
		t.Fatalf("renderer: %v", err)
	}

	out, err := renderer.Render(TemplateSecurityAlert, SecurityAlertData{
		Title:        "Credential update\r\nBcc:evil@example.com",
		Description:  "<script>alert('x')</script>",
		Timestamp:    "2026-04-11T08:00:00Z",
		IPAddress:    "203.0.113.20",
		DashboardURL: "https://virtengine.com/account",
		Unsubscribe:  "https://virtengine.com/unsub",
	})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if strings.ContainsAny(out.Subject, "\r\n") {
		t.Fatalf("expected sanitized subject, got %q", out.Subject)
	}
	if strings.Contains(out.HTML, "<script>") || !strings.Contains(out.HTML, "&lt;script&gt;") {
		t.Fatalf("expected HTML output to escape script content, got %q", out.HTML)
	}
}

func TestSendTemplateIncludesUnsubscribe(t *testing.T) {
	renderer, err := NewDefaultRenderer()
	if err != nil {
		t.Fatalf("renderer: %v", err)
	}
	sender := &memorySender{}
	service := NewService(renderer, sender, "noreply@virtengine.com", "https://virtengine.com/unsubscribe")

	data := VEIDStatusData{
		Status:       "approved",
		Details:      "Your identity is verified.",
		DashboardURL: "https://virtengine.com/account",
		Unsubscribe:  service.BuildUnsubscribeURL("token-123"),
	}
	if err := service.SendTemplate(context.Background(), "user@example.com", "token-123", TemplateVEIDStatus, data); err != nil {
		t.Fatalf("send: %v", err)
	}
	if sender.last.Headers["List-Unsubscribe"] == "" {
		t.Fatalf("expected List-Unsubscribe header")
	}
}

func TestSendTemplateOmitsInvalidUnsubscribeHeader(t *testing.T) {
	renderer, err := NewDefaultRenderer()
	if err != nil {
		t.Fatalf("renderer: %v", err)
	}
	sender := &memorySender{}
	service := NewService(renderer, sender, "noreply@virtengine.com", ":bad-url")

	if err := service.SendTemplate(context.Background(), "user@example.com", "token-123", TemplateVEIDStatus, VEIDStatusData{
		Status:       "approved",
		Details:      "Verified",
		DashboardURL: "https://virtengine.com/account",
	}); err != nil {
		t.Fatalf("send: %v", err)
	}
	if sender.last.Headers != nil && sender.last.Headers["List-Unsubscribe"] != "" {
		t.Fatalf("expected invalid unsubscribe base to omit header, got %+v", sender.last.Headers)
	}
}

func TestSendTemplateRejectsUnsafeRecipientAndMissingSender(t *testing.T) {
	renderer, err := NewDefaultRenderer()
	if err != nil {
		t.Fatalf("renderer: %v", err)
	}
	service := NewService(renderer, nil, "noreply@virtengine.com", "https://virtengine.com/unsubscribe")
	if err := service.SendTemplate(context.Background(), "user@example.com", "token-123", TemplateVEIDStatus, VEIDStatusData{
		Status:       "approved",
		Details:      "Verified",
		DashboardURL: "https://virtengine.com/account",
	}); !errors.Is(err, ErrSenderNotConfigured) {
		t.Fatalf("expected sender-not-configured error, got %v", err)
	}

	service = NewService(renderer, &memorySender{}, "noreply@virtengine.com", "https://virtengine.com/unsubscribe")
	if err := service.SendTemplate(context.Background(), "user@example.com\r\nBcc:evil@example.com", "token-123", TemplateVEIDStatus, VEIDStatusData{
		Status:       "approved",
		Details:      "Verified",
		DashboardURL: "https://virtengine.com/account",
	}); err == nil {
		t.Fatalf("expected invalid recipient error")
	}
}
