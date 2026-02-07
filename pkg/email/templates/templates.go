package templates

import (
	"bytes"
	"html/template"
)

const baseStyle = "font-family: Arial, sans-serif; color: #1f2937;"

// OrderConfirmationData contains template data.
type OrderConfirmationData struct {
	OrderID      string
	ProviderName string
	ServiceName  string
	Amount       string
	Currency     string
	DashboardURL string
}

// VEIDStatusData contains VEID status information.
type VEIDStatusData struct {
	VEID      string
	Status    string
	ActionURL string
}

// SecurityAlertData contains security alert information.
type SecurityAlertData struct {
	Event      string
	Location   string
	OccurredAt string
	ActionURL  string
}

// WeeklyDigestData contains digest content.
type WeeklyDigestData struct {
	Username  string
	Summary   []string
	ActionURL string
}

// RenderOrderConfirmation renders the order confirmation email.
func RenderOrderConfirmation(data OrderConfirmationData) (string, error) {
	return renderTemplate(orderConfirmationTemplate, map[string]any{
		"BaseStyle":    baseStyle,
		"OrderID":      data.OrderID,
		"ProviderName": data.ProviderName,
		"ServiceName":  data.ServiceName,
		"Amount":       data.Amount,
		"Currency":     data.Currency,
		"DashboardURL": data.DashboardURL,
	})
}

// RenderVEIDStatus renders a VEID status email.
func RenderVEIDStatus(data VEIDStatusData) (string, error) {
	return renderTemplate(veidStatusTemplate, map[string]any{
		"BaseStyle": baseStyle,
		"VEID":      data.VEID,
		"Status":    data.Status,
		"ActionURL": data.ActionURL,
	})
}

// RenderSecurityAlert renders a security alert email.
func RenderSecurityAlert(data SecurityAlertData) (string, error) {
	return renderTemplate(securityAlertTemplate, map[string]any{
		"BaseStyle":  baseStyle,
		"Event":      data.Event,
		"Location":   data.Location,
		"OccurredAt": data.OccurredAt,
		"ActionURL":  data.ActionURL,
	})
}

// RenderWeeklyDigest renders a weekly digest email.
func RenderWeeklyDigest(data WeeklyDigestData) (string, error) {
	return renderTemplate(weeklyDigestTemplate, map[string]any{
		"BaseStyle": baseStyle,
		"Username":  data.Username,
		"Summary":   data.Summary,
		"ActionURL": data.ActionURL,
	})
}

func renderTemplate(tmpl string, data any) (string, error) {
	t, err := template.New("email").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

const orderConfirmationTemplate = "<!doctype html>\n<html>\n  <body style=\"{{.BaseStyle}}\">\n    <h2>Order Confirmed</h2>\n    <p>Your order <strong>{{.OrderID}}</strong> has been confirmed.</p>\n    <p><strong>Provider:</strong> {{.ProviderName}}</p>\n    <p><strong>Service:</strong> {{.ServiceName}}</p>\n    <p><strong>Amount:</strong> {{.Amount}} {{.Currency}}</p>\n    <p><a href=\"{{.DashboardURL}}\">View order</a></p>\n  </body>\n</html>\n"

const veidStatusTemplate = "<!doctype html>\n<html>\n  <body style=\"{{.BaseStyle}}\">\n    <h2>VEID Verification Update</h2>\n    <p>Your VEID <strong>{{.VEID}}</strong> status changed to <strong>{{.Status}}</strong>.</p>\n    <p><a href=\"{{.ActionURL}}\">Review verification details</a></p>\n  </body>\n</html>\n"

const securityAlertTemplate = "<!doctype html>\n<html>\n  <body style=\"{{.BaseStyle}}\">\n    <h2>Security Alert</h2>\n    <p>We detected: <strong>{{.Event}}</strong>.</p>\n    <p>Location: {{.Location}}</p>\n    <p>Time: {{.OccurredAt}}</p>\n    <p><a href=\"{{.ActionURL}}\">Review security settings</a></p>\n  </body>\n</html>\n"

const weeklyDigestTemplate = "<!doctype html>\n<html>\n  <body style=\"{{.BaseStyle}}\">\n    <h2>Weekly Digest</h2>\n    <p>Hello {{.Username}}, here is your weekly summary:</p>\n    <ul>\n      {{range .Summary}}<li>{{.}}</li>{{end}}\n    </ul>\n    <p><a href=\"{{.ActionURL}}\">View dashboard</a></p>\n  </body>\n</html>\n"
