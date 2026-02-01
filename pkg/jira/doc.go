// Package jira provides Jira Service Desk integration for VirtEngine.
//
// VE-919: Jira Service Desk using Waldur
//
// This package implements enterprise support ticket management through
// Jira Service Desk integration. It provides:
//
// - Jira REST API client for issue management
// - Ticket bridge for creating Jira issues from VirtEngine support requests
// - SLA tracking for response and resolution times
// - Webhook handling for Jira status change notifications
//
// # Architecture
//
// The package is organized into several components:
//
// Client: Low-level Jira REST API client that handles authentication
// and HTTP communication with Jira Cloud or Server instances.
//
// TicketBridge: Translates VirtEngine support requests into Jira issues,
// handling priority mapping, category-to-component mapping, and
// custom field population.
//
// SLATracker: Monitors SLA compliance for response and resolution times,
// with support for pausing SLA clocks (e.g., waiting for customer).
//
// WebhookHandler: Processes Jira webhooks for status changes, comments,
// and other events to synchronize state back to VirtEngine.
//
// Service: Coordinates all components and provides a unified interface
// for ticket management operations.
//
// # Security Considerations
//
// CRITICAL: This package handles sensitive data. The following rules MUST be followed:
//
// 1. Never log API tokens, bearer tokens, or webhook secrets
// 2. Never log decrypted ticket content
// 3. Truncate wallet addresses in external systems
// 4. Use HTTPS for all Jira API communication
// 5. Verify webhook signatures when RequireSignature is enabled
//
// # Usage
//
// Basic usage:
//
//	cfg := jira.ServiceConfig{
//		ClientConfig: jira.ClientConfig{
//			BaseURL: "https://company.atlassian.net",
//			Auth: jira.AuthConfig{
//				Type:     jira.AuthTypeBasic,
//				Username: "user@company.com",
//				APIToken: os.Getenv("JIRA_API_TOKEN"), // Never hardcode!
//			},
//		},
//		BridgeConfig: jira.DefaultTicketBridgeConfig(),
//		SLAConfig:    jira.DefaultSLAConfig(),
//	}
//
//	service, err := jira.NewService(cfg)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Create ticket from support request
//	req := &jira.VirtEngineSupportRequest{
//		TicketID:     "ve-001",
//		TicketNumber: "VE-001",
//		Subject:      "Need help with deployment",
//		Description:  "...",
//		Category:     "technical",
//		Priority:     "high",
//		CreatedAt:    time.Now(),
//	}
//
//	issue, err := service.CreateTicket(ctx, req)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Printf("Created Jira issue: %s\n", issue.Key)
//
// # Webhook Integration
//
// To receive Jira webhooks:
//
//	handler := service.GetWebhookHandler()
//	http.Handle("/webhooks/jira", handler)
//
// # SLA Configuration
//
// SLA targets can be customized per priority:
//
//	slaConfig := jira.SLAConfig{
//		ResponseTimeTargets: map[string]int64{
//			"urgent": 15,   // 15 minutes
//			"high":   60,   // 1 hour
//			"medium": 240,  // 4 hours
//			"low":    1440, // 24 hours
//		},
//		ResolutionTimeTargets: map[string]int64{
//			"urgent": 240,   // 4 hours
//			"high":   480,   // 8 hours
//			"medium": 2880,  // 2 days
//			"low":    10080, // 7 days
//		},
//		BusinessHours: &jira.BusinessHours{
//			StartHour: 9,
//			EndHour:   17,
//			Timezone:  "America/New_York",
//		},
//		ExcludeWeekends: true,
//	}
package jira

