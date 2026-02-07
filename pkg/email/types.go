/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

package email

// TemplateName identifies email templates.
type TemplateName string

const (
	TemplateOrderConfirmation TemplateName = "order_confirmation"
	TemplateVEIDStatus        TemplateName = "veid_status"
	TemplateSecurityAlert     TemplateName = "security_alert"
	TemplateWeeklyDigest      TemplateName = "weekly_digest"
)

// EmailMessage represents an outgoing email.
type EmailMessage struct {
	To      string
	From    string
	Subject string
	HTML    string
	Text    string
	Headers map[string]string
}

// RenderedEmail is a rendered template payload.
type RenderedEmail struct {
	Subject string
	HTML    string
	Text    string
}

// OrderConfirmationData provides template data.
type OrderConfirmationData struct {
	OrderID      string
	ProviderName string
	ServiceName  string
	Amount       string
	Currency     string
	DashboardURL string
	Unsubscribe  string
}

// VEIDStatusData provides template data.
type VEIDStatusData struct {
	Status       string
	Details      string
	DashboardURL string
	Unsubscribe  string
}

// SecurityAlertData provides template data.
type SecurityAlertData struct {
	Title        string
	Description  string
	Timestamp    string
	IPAddress    string
	DashboardURL string
	Unsubscribe  string
}

// DigestItem is a weekly digest entry.
type DigestItem struct {
	Title string
	Body  string
	Link  string
}

// WeeklyDigestData provides template data.
type WeeklyDigestData struct {
	Summary      string
	Items        []DigestItem
	DashboardURL string
	Unsubscribe  string
}
