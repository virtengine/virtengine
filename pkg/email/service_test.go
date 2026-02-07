package email

import (
	"context"
	"strings"
	"testing"
)

type memorySender struct {
	last EmailMessage
}

func (m *memorySender) Send(_ context.Context, message EmailMessage) error {
	m.last = message
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
		DashboardURL: "https://virtengine.io/dashboard",
		Unsubscribe:  "https://virtengine.io/unsub",
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

func TestSendTemplateIncludesUnsubscribe(t *testing.T) {
	renderer, err := NewDefaultRenderer()
	if err != nil {
		t.Fatalf("renderer: %v", err)
	}
	sender := &memorySender{}
	service := NewService(renderer, sender, "noreply@virtengine.io", "https://virtengine.io/unsubscribe")

	data := VEIDStatusData{
		Status:       "approved",
		Details:      "Your identity is verified.",
		DashboardURL: "https://virtengine.io/account",
		Unsubscribe:  service.BuildUnsubscribeURL("token-123"),
	}
	if err := service.SendTemplate(context.Background(), "user@example.com", "token-123", TemplateVEIDStatus, data); err != nil {
		t.Fatalf("send: %v", err)
	}
	if sender.last.Headers["List-Unsubscribe"] == "" {
		t.Fatalf("expected List-Unsubscribe header")
	}
}
