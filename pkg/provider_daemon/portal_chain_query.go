package provider_daemon

import (
	"context"
	"errors"
	"time"
)

// ErrPortalNotFound indicates the requested resource does not exist.
var ErrPortalNotFound = errors.New("portal resource not found")

// ErrPortalFeatureUnavailable identifies portal capabilities without an authoritative backend.
var ErrPortalFeatureUnavailable = errors.New("portal feature unavailable")

// PortalFeatureUnavailableError describes a portal capability that is not available.
type PortalFeatureUnavailableError struct {
	Capability PortalRouteCapability
	Owner      string
}

func (e *PortalFeatureUnavailableError) Error() string {
	return "portal feature unavailable: " + string(e.Capability) + " (owner: " + e.Owner + ")"
}

// Unwrap supports errors.Is checks against ErrPortalFeatureUnavailable.
func (e *PortalFeatureUnavailableError) Unwrap() error {
	return ErrPortalFeatureUnavailable
}

func featureUnavailable(capability PortalRouteCapability, owner string) error {
	return &PortalFeatureUnavailableError{Capability: capability, Owner: owner}
}

// PortalRouteCapability identifies an independently backed portal query surface.
type PortalRouteCapability string

const (
	PortalCapabilityOrganizations PortalRouteCapability = "organizations"
	PortalCapabilityTickets       PortalRouteCapability = "tickets"
	PortalCapabilityBilling       PortalRouteCapability = "billing"
	PortalCapabilityUsage         PortalRouteCapability = "usage"
	PortalCapabilityEvents        PortalRouteCapability = "events"
	PortalCapabilityMetrics       PortalRouteCapability = "metrics"
	PortalCapabilityRoles         PortalRouteCapability = "roles"
	PortalCapabilityConsent       PortalRouteCapability = "consent"
)

var requiredPortalRouteCapabilities = []PortalRouteCapability{
	PortalCapabilityTickets,
	PortalCapabilityBilling,
	PortalCapabilityUsage,
	PortalCapabilityEvents,
	PortalCapabilityMetrics,
}

// PortalChainQueryCapabilities lets a backend declare route-level availability.
type PortalChainQueryCapabilities interface {
	PortalCapability(capability PortalRouteCapability) error
}

func portalQueryCapability(query ChainQuery, capability PortalRouteCapability) error {
	if query == nil {
		return featureUnavailable(capability, "86C")
	}
	capabilities, ok := query.(PortalChainQueryCapabilities)
	if !ok {
		return nil
	}
	return capabilities.PortalCapability(capability)
}

func validatePortalChainQuery(query ChainQuery) error {
	for _, capability := range requiredPortalRouteCapabilities {
		if err := portalQueryCapability(query, capability); err != nil {
			return err
		}
	}
	return nil
}

// ChainQuery provides access to on-chain data for portal endpoints.
type ChainQuery interface {
	// Organization management.
	ListOrganizations(ctx context.Context, address string, limit int, cursor string) ([]Organization, string, error)
	GetOrganization(ctx context.Context, orgID string) (*OrganizationDetail, error)
	ListOrganizationMembers(ctx context.Context, orgID string) ([]OrganizationMember, error)
	IsOrganizationAdmin(ctx context.Context, orgID, address string) (bool, error)
	InviteOrganizationMember(ctx context.Context, orgID, address, role, invitedBy string) (*OrganizationMember, error)
	RemoveOrganizationMember(ctx context.Context, orgID, address, removedBy string) error

	// Support tickets.
	ListTickets(ctx context.Context, address, status, deploymentID string) ([]Ticket, error)
	CreateTicket(ctx context.Context, address string, req CreateTicketRequest) (*Ticket, error)
	GetTicket(ctx context.Context, ticketID string) (*TicketDetail, error)
	AddTicketComment(ctx context.Context, ticketID, author, message string) (*TicketComment, error)
	UpdateTicket(ctx context.Context, ticketID string, req UpdateTicketRequest) (*Ticket, error)

	// Billing and usage.
	ListInvoices(ctx context.Context, address, status string, limit int, cursor string) ([]Invoice, string, error)
	GetInvoice(ctx context.Context, invoiceID string) (*Invoice, error)
	GetUsageSummary(ctx context.Context, address string) (*UsageSummary, error)
	GetUsageHistory(ctx context.Context, address string, start, end time.Time, interval time.Duration) (*UsageHistoryResponse, error)

	// Metrics and events.
	GetDeploymentMetrics(ctx context.Context, deploymentID string) (*PortalResourceMetrics, error)
	GetDeploymentMetricsHistory(ctx context.Context, deploymentID string, start, end time.Time, interval time.Duration) (*MetricsSeriesResponse, error)
	GetDeploymentEvents(ctx context.Context, deploymentID string, limit int, cursor string) ([]DeploymentEvent, string, error)
	GetAggregatedMetrics(ctx context.Context, start, end time.Time, interval time.Duration) (*MetricsSeriesResponse, error)

	// Roles.
	HasRole(ctx context.Context, address, role string) (bool, error)
	// Consent.
	HasConsent(ctx context.Context, address, scopeID string) (bool, error)
}

// NoopChainQuery is an explicitly selected development/test backend that fails closed.
type NoopChainQuery struct{}

// UnavailablePortalChainQuery is installed when production construction has no query backend.
type UnavailablePortalChainQuery struct {
	NoopChainQuery
}

// PortalCapability reports every noop route as unavailable.
func (NoopChainQuery) PortalCapability(capability PortalRouteCapability) error {
	owner := "86C"
	if capability == PortalCapabilityOrganizations {
		owner = "89C"
	}
	return featureUnavailable(capability, owner)
}

// ListOrganizations reports that organization queries require the Task 89C backend.
func (NoopChainQuery) ListOrganizations(_ context.Context, _ string, _ int, _ string) ([]Organization, string, error) {
	return nil, "", featureUnavailable(PortalCapabilityOrganizations, "89C")
}

// GetOrganization reports that organization queries require the Task 89C backend.
func (NoopChainQuery) GetOrganization(_ context.Context, _ string) (*OrganizationDetail, error) {
	return nil, featureUnavailable(PortalCapabilityOrganizations, "89C")
}

// ListOrganizationMembers reports that organization queries require the Task 89C backend.
func (NoopChainQuery) ListOrganizationMembers(_ context.Context, _ string) ([]OrganizationMember, error) {
	return nil, featureUnavailable(PortalCapabilityOrganizations, "89C")
}

// IsOrganizationAdmin reports that organization queries require the Task 89C backend.
func (NoopChainQuery) IsOrganizationAdmin(_ context.Context, _ string, _ string) (bool, error) {
	return false, featureUnavailable(PortalCapabilityOrganizations, "89C")
}

// InviteOrganizationMember reports that organization mutations require the Task 89C backend.
func (NoopChainQuery) InviteOrganizationMember(_ context.Context, _ string, _ string, _ string, _ string) (*OrganizationMember, error) {
	return nil, featureUnavailable(PortalCapabilityOrganizations, "89C")
}

// RemoveOrganizationMember reports that organization mutations require the Task 89C backend.
func (NoopChainQuery) RemoveOrganizationMember(_ context.Context, _ string, _ string, _ string) error {
	return featureUnavailable(PortalCapabilityOrganizations, "89C")
}

// ListTickets reports that ticket queries require an authoritative backend.
func (NoopChainQuery) ListTickets(_ context.Context, _ string, _ string, _ string) ([]Ticket, error) {
	return nil, featureUnavailable(PortalCapabilityTickets, "86C")
}

// CreateTicket reports that ticket mutations require an authoritative backend.
func (NoopChainQuery) CreateTicket(_ context.Context, _ string, _ CreateTicketRequest) (*Ticket, error) {
	return nil, featureUnavailable(PortalCapabilityTickets, "86C")
}

// GetTicket reports that ticket queries require an authoritative backend.
func (NoopChainQuery) GetTicket(_ context.Context, _ string) (*TicketDetail, error) {
	return nil, featureUnavailable(PortalCapabilityTickets, "86C")
}

// AddTicketComment reports that ticket mutations require an authoritative backend.
func (NoopChainQuery) AddTicketComment(_ context.Context, _ string, _ string, _ string) (*TicketComment, error) {
	return nil, featureUnavailable(PortalCapabilityTickets, "86C")
}

// UpdateTicket reports that ticket mutations require an authoritative backend.
func (NoopChainQuery) UpdateTicket(_ context.Context, _ string, _ UpdateTicketRequest) (*Ticket, error) {
	return nil, featureUnavailable(PortalCapabilityTickets, "86C")
}

// ListInvoices reports that billing queries require an authoritative backend.
func (NoopChainQuery) ListInvoices(_ context.Context, _ string, _ string, _ int, _ string) ([]Invoice, string, error) {
	return nil, "", featureUnavailable(PortalCapabilityBilling, "86C")
}

// GetInvoice reports that billing queries require an authoritative backend.
func (NoopChainQuery) GetInvoice(_ context.Context, _ string) (*Invoice, error) {
	return nil, featureUnavailable(PortalCapabilityBilling, "86C")
}

// GetUsageSummary reports that usage queries require an authoritative backend.
func (NoopChainQuery) GetUsageSummary(_ context.Context, _ string) (*UsageSummary, error) {
	return nil, featureUnavailable(PortalCapabilityUsage, "86C")
}

// GetUsageHistory reports that usage queries require an authoritative backend.
func (NoopChainQuery) GetUsageHistory(_ context.Context, _ string, _ time.Time, _ time.Time, _ time.Duration) (*UsageHistoryResponse, error) {
	return nil, featureUnavailable(PortalCapabilityUsage, "86C")
}

// GetDeploymentMetrics reports that metrics queries require an authoritative backend.
func (NoopChainQuery) GetDeploymentMetrics(_ context.Context, _ string) (*PortalResourceMetrics, error) {
	return nil, featureUnavailable(PortalCapabilityMetrics, "86C")
}

// GetDeploymentMetricsHistory reports that metrics queries require an authoritative backend.
func (NoopChainQuery) GetDeploymentMetricsHistory(_ context.Context, _ string, _ time.Time, _ time.Time, _ time.Duration) (*MetricsSeriesResponse, error) {
	return nil, featureUnavailable(PortalCapabilityMetrics, "86C")
}

// GetDeploymentEvents reports that event queries require an authoritative backend.
func (NoopChainQuery) GetDeploymentEvents(_ context.Context, _ string, _ int, _ string) ([]DeploymentEvent, string, error) {
	return nil, "", featureUnavailable(PortalCapabilityEvents, "86C")
}

// GetAggregatedMetrics reports that metrics queries require an authoritative backend.
func (NoopChainQuery) GetAggregatedMetrics(_ context.Context, _ time.Time, _ time.Time, _ time.Duration) (*MetricsSeriesResponse, error) {
	return nil, featureUnavailable(PortalCapabilityMetrics, "86C")
}

// HasRole reports that role queries require an authoritative backend.
func (NoopChainQuery) HasRole(_ context.Context, _ string, _ string) (bool, error) {
	return false, featureUnavailable(PortalCapabilityRoles, "86C")
}

// HasConsent reports that consent queries require an authoritative backend.
func (NoopChainQuery) HasConsent(_ context.Context, _ string, _ string) (bool, error) {
	return false, featureUnavailable(PortalCapabilityConsent, "86C")
}
