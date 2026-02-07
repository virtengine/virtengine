package provider_daemon

import (
	"context"
	"errors"
	"time"
)

// ErrPortalNotFound indicates the requested resource does not exist.
var ErrPortalNotFound = errors.New("portal resource not found")

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
}

// NoopChainQuery is a default implementation that returns empty results.
type NoopChainQuery struct{}

// ListOrganizations returns no organizations.
func (NoopChainQuery) ListOrganizations(_ context.Context, _ string, _ int, _ string) ([]Organization, string, error) {
	return []Organization{}, "", nil
}

// GetOrganization returns not found.
func (NoopChainQuery) GetOrganization(_ context.Context, _ string) (*OrganizationDetail, error) {
	return nil, ErrPortalNotFound
}

// ListOrganizationMembers returns no members.
func (NoopChainQuery) ListOrganizationMembers(_ context.Context, _ string) ([]OrganizationMember, error) {
	return []OrganizationMember{}, nil
}

// IsOrganizationAdmin returns false.
func (NoopChainQuery) IsOrganizationAdmin(_ context.Context, _ string, _ string) (bool, error) {
	return false, nil
}

// InviteOrganizationMember is unsupported in noop implementation.
func (NoopChainQuery) InviteOrganizationMember(_ context.Context, _ string, _ string, _ string, _ string) (*OrganizationMember, error) {
	return nil, errors.New("organization invites not supported")
}

// RemoveOrganizationMember is unsupported in noop implementation.
func (NoopChainQuery) RemoveOrganizationMember(_ context.Context, _ string, _ string, _ string) error {
	return errors.New("organization member removal not supported")
}

// ListTickets returns no tickets.
func (NoopChainQuery) ListTickets(_ context.Context, _ string, _ string, _ string) ([]Ticket, error) {
	return []Ticket{}, nil
}

// CreateTicket is unsupported in noop implementation.
func (NoopChainQuery) CreateTicket(_ context.Context, _ string, _ CreateTicketRequest) (*Ticket, error) {
	return nil, errors.New("ticket creation not supported")
}

// GetTicket returns not found.
func (NoopChainQuery) GetTicket(_ context.Context, _ string) (*TicketDetail, error) {
	return nil, ErrPortalNotFound
}

// AddTicketComment is unsupported in noop implementation.
func (NoopChainQuery) AddTicketComment(_ context.Context, _ string, _ string, _ string) (*TicketComment, error) {
	return nil, errors.New("ticket comments not supported")
}

// UpdateTicket is unsupported in noop implementation.
func (NoopChainQuery) UpdateTicket(_ context.Context, _ string, _ UpdateTicketRequest) (*Ticket, error) {
	return nil, errors.New("ticket updates not supported")
}

// ListInvoices returns no invoices.
func (NoopChainQuery) ListInvoices(_ context.Context, _ string, _ string, _ int, _ string) ([]Invoice, string, error) {
	return []Invoice{}, "", nil
}

// GetInvoice returns not found.
func (NoopChainQuery) GetInvoice(_ context.Context, _ string) (*Invoice, error) {
	return nil, ErrPortalNotFound
}

// GetUsageSummary returns empty summary.
func (NoopChainQuery) GetUsageSummary(_ context.Context, _ string) (*UsageSummary, error) {
	return &UsageSummary{}, nil
}

// GetUsageHistory returns empty history.
func (NoopChainQuery) GetUsageHistory(_ context.Context, _ string, _ time.Time, _ time.Time, _ time.Duration) (*UsageHistoryResponse, error) {
	return &UsageHistoryResponse{Series: []UsageHistoryPoint{}}, nil
}

// GetDeploymentMetrics returns not found.
func (NoopChainQuery) GetDeploymentMetrics(_ context.Context, _ string) (*PortalResourceMetrics, error) {
	return nil, ErrPortalNotFound
}

// GetDeploymentMetricsHistory returns empty series.
func (NoopChainQuery) GetDeploymentMetricsHistory(_ context.Context, _ string, _ time.Time, _ time.Time, _ time.Duration) (*MetricsSeriesResponse, error) {
	return &MetricsSeriesResponse{Series: []MetricsPoint{}}, nil
}

// GetDeploymentEvents returns empty events.
func (NoopChainQuery) GetDeploymentEvents(_ context.Context, _ string, _ int, _ string) ([]DeploymentEvent, string, error) {
	return []DeploymentEvent{}, "", nil
}

// GetAggregatedMetrics returns empty series.
func (NoopChainQuery) GetAggregatedMetrics(_ context.Context, _ time.Time, _ time.Time, _ time.Duration) (*MetricsSeriesResponse, error) {
	return &MetricsSeriesResponse{Series: []MetricsPoint{}}, nil
}
