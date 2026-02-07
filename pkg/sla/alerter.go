/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

package sla

import (
	"context"
	"errors"
	"fmt"

	"github.com/virtengine/virtengine/pkg/notifications"
)

// AlertSeverity controls alert routing.
type AlertSeverity string

const (
	AlertInfo     AlertSeverity = "info"
	AlertWarning  AlertSeverity = "warning"
	AlertCritical AlertSeverity = "critical"
)

// Alert represents a monitoring alert.
type Alert struct {
	Severity AlertSeverity
	Title    string
	Message  string
	Labels   map[string]string
}

// CustomerNotification describes a customer notification payload.
type CustomerNotification struct {
	Type    string
	Title   string
	Message string
	LeaseID string
}

// ProviderNotification describes a provider notification payload.
type ProviderNotification struct {
	Type    string
	Title   string
	Message string
	LeaseID string
	Metric  SLAMetricType
}

// Alerter dispatches alerts and user notifications.
type Alerter interface {
	SendAlert(ctx context.Context, alert Alert) error
	NotifyCustomer(ctx context.Context, customerID string, notif CustomerNotification) error
	NotifyProvider(ctx context.Context, providerID string, notif ProviderNotification) error
}

// PrometheusClient sends alerts to Alertmanager.
type PrometheusClient interface {
	SendAlert(ctx context.Context, alert PrometheusAlert) error
}

// PrometheusAlert is a minimal Alertmanager payload.
type PrometheusAlert struct {
	Labels      map[string]string
	Annotations map[string]string
}

// PagerDutyClient pages on-call.
type PagerDutyClient interface {
	CreateIncident(ctx context.Context, incident PagerDutyIncident) error
}

// PagerDutyIncident represents a PagerDuty incident.
type PagerDutyIncident struct {
	Title    string
	Body     string
	Severity string
	Service  string
}

var errNotificationsUnavailable = errors.New("notification service unavailable")

// AlerterImpl wires Alertmanager, PagerDuty, and notifications.
type AlerterImpl struct {
	prometheus    PrometheusClient
	pagerduty     PagerDutyClient
	notifications notifications.Service
}

// NewAlerter creates an alerter implementation.
func NewAlerter(prometheus PrometheusClient, pagerduty PagerDutyClient, notifService notifications.Service) *AlerterImpl {
	return &AlerterImpl{
		prometheus:    prometheus,
		pagerduty:     pagerduty,
		notifications: notifService,
	}
}

// SendAlert dispatches a monitoring alert.
func (a *AlerterImpl) SendAlert(ctx context.Context, alert Alert) error {
	if a == nil {
		return nil
	}

	if a.prometheus != nil {
		if err := a.prometheus.SendAlert(ctx, PrometheusAlert{
			Labels: map[string]string{
				"alertname": alert.Title,
				"severity":  string(alert.Severity),
			},
			Annotations: map[string]string{
				"summary":     alert.Title,
				"description": alert.Message,
			},
		}); err != nil {
			return err
		}
	}

	if alert.Severity == AlertCritical && a.pagerduty != nil {
		return a.pagerduty.CreateIncident(ctx, PagerDutyIncident{
			Title:    alert.Title,
			Body:     alert.Message,
			Severity: "critical",
			Service:  "virtengine-sla",
		})
	}

	return nil
}

// NotifyCustomer sends notifications to customers.
func (a *AlerterImpl) NotifyCustomer(ctx context.Context, customerID string, notif CustomerNotification) error {
	if a == nil {
		return nil
	}
	if a.notifications == nil {
		return errNotificationsUnavailable
	}

	notifType := notifications.NotificationTypeSLABreach
	if notif.Type == "sla_warning" {
		notifType = notifications.NotificationTypeSLAWarning
	}

	return a.notifications.Send(ctx, notifications.Notification{
		UserAddress: customerID,
		Type:        notifType,
		Title:       notif.Title,
		Body:        notif.Message,
		Data: map[string]string{
			"lease_id": notif.LeaseID,
		},
		Channels: []notifications.Channel{notifications.ChannelPush, notifications.ChannelEmail, notifications.ChannelInApp},
	})
}

// NotifyProvider sends provider alerts.
func (a *AlerterImpl) NotifyProvider(ctx context.Context, providerID string, notif ProviderNotification) error {
	if a == nil {
		return nil
	}
	if a.notifications == nil {
		return errNotificationsUnavailable
	}

	data := map[string]string{
		"lease_id": notif.LeaseID,
		"metric":   string(notif.Metric),
	}

	return a.notifications.Send(ctx, notifications.Notification{
		UserAddress: providerID,
		Type:        notifications.NotificationTypeProviderAlert,
		Title:       fmt.Sprintf("Provider Alert: %s", notif.Title),
		Body:        notif.Message,
		Data:        data,
		Channels:    []notifications.Channel{notifications.ChannelPush, notifications.ChannelEmail, notifications.ChannelInApp},
	})
}
