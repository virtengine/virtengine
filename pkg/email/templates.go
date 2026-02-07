/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

package email

import (
	"bytes"
	"fmt"
	"html/template"
)

type templateBundle struct {
	subject *template.Template
	html    *template.Template
	text    *template.Template
}

// Renderer renders email templates.
type Renderer interface {
	Render(name TemplateName, data any) (RenderedEmail, error)
}

// DefaultRenderer uses Go HTML templates for email rendering.
type DefaultRenderer struct {
	templates map[TemplateName]templateBundle
}

// NewDefaultRenderer creates a renderer with built-in templates.
func NewDefaultRenderer() (*DefaultRenderer, error) {
	templates := map[TemplateName]templateBundle{}

	for name, tpl := range templateSources() {
		subject, err := template.New(string(name) + "_subject").Parse(tpl.subject)
		if err != nil {
			return nil, fmt.Errorf("subject template %s: %w", name, err)
		}
		htmlTpl, err := template.New(string(name) + "_html").Parse(tpl.html)
		if err != nil {
			return nil, fmt.Errorf("html template %s: %w", name, err)
		}
		textTpl, err := template.New(string(name) + "_text").Parse(tpl.text)
		if err != nil {
			return nil, fmt.Errorf("text template %s: %w", name, err)
		}
		templates[name] = templateBundle{
			subject: subject,
			html:    htmlTpl,
			text:    textTpl,
		}
	}

	return &DefaultRenderer{templates: templates}, nil
}

// Render renders a template by name.
func (r *DefaultRenderer) Render(name TemplateName, data any) (RenderedEmail, error) {
	bundle, ok := r.templates[name]
	if !ok {
		return RenderedEmail{}, fmt.Errorf("template %s not found", name)
	}

	var subject bytes.Buffer
	if err := bundle.subject.Execute(&subject, data); err != nil {
		return RenderedEmail{}, err
	}

	var html bytes.Buffer
	if err := bundle.html.Execute(&html, data); err != nil {
		return RenderedEmail{}, err
	}

	var text bytes.Buffer
	if err := bundle.text.Execute(&text, data); err != nil {
		return RenderedEmail{}, err
	}

	return RenderedEmail{
		Subject: subject.String(),
		HTML:    html.String(),
		Text:    text.String(),
	}, nil
}

type templateSource struct {
	subject string
	html    string
	text    string
}

func templateSources() map[TemplateName]templateSource {
	return map[TemplateName]templateSource{
		TemplateOrderConfirmation: {
			subject: "Order {{.OrderID}} confirmed",
			html: `<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif; background:#f6f9fc; padding:24px;">
  <div style="max-width:600px;margin:0 auto;background:#fff;padding:24px;border-radius:12px;">
    <h1 style="color:#1f2937;">Order confirmed</h1>
    <p>Your order <strong>{{.OrderID}}</strong> has been confirmed.</p>
    <p><strong>Provider:</strong> {{.ProviderName}}<br/>
       <strong>Service:</strong> {{.ServiceName}}<br/>
       <strong>Amount:</strong> {{.Amount}} {{.Currency}}</p>
    <p><a href="{{.DashboardURL}}">View order</a></p>
    <p style="color:#6b7280;font-size:12px;">VirtEngine — Decentralized cloud computing.</p>
    <p style="color:#9ca3af;font-size:11px;">Unsubscribe: <a href="{{.Unsubscribe}}">Manage preferences</a></p>
  </div>
</body>
</html>`,
			text: `Your order {{.OrderID}} has been confirmed.
Provider: {{.ProviderName}}
Service: {{.ServiceName}}
Amount: {{.Amount}} {{.Currency}}
View order: {{.DashboardURL}}
Unsubscribe: {{.Unsubscribe}}`,
		},
		TemplateVEIDStatus: {
			subject: "VEID verification {{.Status}}",
			html: `<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif; background:#f6f9fc; padding:24px;">
  <div style="max-width:600px;margin:0 auto;background:#fff;padding:24px;border-radius:12px;">
    <h1 style="color:#1f2937;">VEID status update</h1>
    <p>Your VEID verification status is now <strong>{{.Status}}</strong>.</p>
    <p>{{.Details}}</p>
    <p><a href="{{.DashboardURL}}">View your verification</a></p>
    <p style="color:#9ca3af;font-size:11px;">Unsubscribe: <a href="{{.Unsubscribe}}">Manage preferences</a></p>
  </div>
</body>
</html>`,
			text: `VEID status update: {{.Status}}
{{.Details}}
View: {{.DashboardURL}}
Unsubscribe: {{.Unsubscribe}}`,
		},
		TemplateSecurityAlert: {
			subject: "Security alert: {{.Title}}",
			html: `<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif; background:#fff7ed; padding:24px;">
  <div style="max-width:600px;margin:0 auto;background:#fff;padding:24px;border-radius:12px;border:1px solid #fed7aa;">
    <h1 style="color:#9a3412;">Security alert</h1>
    <p><strong>{{.Title}}</strong></p>
    <p>{{.Description}}</p>
    <p><strong>Time:</strong> {{.Timestamp}}<br/>
       <strong>IP:</strong> {{.IPAddress}}</p>
    <p><a href="{{.DashboardURL}}">Review security activity</a></p>
    <p style="color:#9ca3af;font-size:11px;">Unsubscribe: <a href="{{.Unsubscribe}}">Manage preferences</a></p>
  </div>
</body>
</html>`,
			text: `Security alert: {{.Title}}
{{.Description}}
Time: {{.Timestamp}}
IP: {{.IPAddress}}
Review: {{.DashboardURL}}
Unsubscribe: {{.Unsubscribe}}`,
		},
		TemplateWeeklyDigest: {
			subject: "Your weekly VirtEngine digest",
			html: `<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif; background:#f3f4f6; padding:24px;">
  <div style="max-width:600px;margin:0 auto;background:#fff;padding:24px;border-radius:12px;">
    <h1 style="color:#111827;">Weekly digest</h1>
    <p>{{.Summary}}</p>
    <ul>
      {{range .Items}}
        <li><strong>{{.Title}}</strong> — {{.Body}} <a href="{{.Link}}">View</a></li>
      {{end}}
    </ul>
    <p><a href="{{.DashboardURL}}">Open dashboard</a></p>
    <p style="color:#9ca3af;font-size:11px;">Unsubscribe: <a href="{{.Unsubscribe}}">Manage preferences</a></p>
  </div>
</body>
</html>`,
			text: `Weekly digest
{{.Summary}}
{{range .Items}}- {{.Title}}: {{.Body}} ({{.Link}})
{{end}}
Dashboard: {{.DashboardURL}}
Unsubscribe: {{.Unsubscribe}}`,
		},
	}
}
