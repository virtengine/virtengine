package provider_daemon

import "time"

// Organization represents an organization/group record in the portal API.
type Organization struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// OrganizationDetail includes members for a specific organization.
type OrganizationDetail struct {
	Organization
	Members []OrganizationMember `json:"members,omitempty"`
}

// OrganizationMember represents a member in an organization.
type OrganizationMember struct {
	Address  string    `json:"address"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

// OrganizationInviteRequest is the payload for inviting a member.
type OrganizationInviteRequest struct {
	Address string `json:"address"`
	Role    string `json:"role"`
}

// Ticket represents a support ticket summary.
type Ticket struct {
	ID           string    `json:"id"`
	DeploymentID string    `json:"deployment_id"`
	Subject      string    `json:"subject"`
	Description  string    `json:"description"`
	Status       string    `json:"status"`
	Priority     string    `json:"priority"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TicketDetail includes comments for a ticket.
type TicketDetail struct {
	Ticket
	Comments []TicketComment `json:"comments,omitempty"`
}

// TicketComment represents a ticket comment.
type TicketComment struct {
	ID        string    `json:"id"`
	Author    string    `json:"author"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateTicketRequest is the payload for creating a ticket.
type CreateTicketRequest struct {
	DeploymentID string `json:"deployment_id"`
	Subject      string `json:"subject"`
	Description  string `json:"description"`
	Category     string `json:"category,omitempty"`
	Priority     string `json:"priority,omitempty"`
}

// UpdateTicketRequest is the payload for updating a ticket.
type UpdateTicketRequest struct {
	Status   string `json:"status,omitempty"`
	Priority string `json:"priority,omitempty"`
}

// TicketCommentRequest is the payload for adding a comment.
type TicketCommentRequest struct {
	Message string `json:"message"`
}

// Invoice represents a billing invoice.
type Invoice struct {
	ID        string            `json:"id"`
	Number    string            `json:"number"`
	Status    string            `json:"status"`
	Total     string            `json:"total"`
	Currency  string            `json:"currency"`
	DueDate   time.Time         `json:"due_date"`
	IssuedAt  time.Time         `json:"issued_at"`
	LineItems []InvoiceLineItem `json:"line_items,omitempty"`
}

// InvoiceLineItem represents an invoice line item.
type InvoiceLineItem struct {
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   string  `json:"unit_price"`
	Total       string  `json:"total"`
}

// InvoiceListResponse is the paginated invoice response.
type InvoiceListResponse struct {
	Invoices   []Invoice `json:"invoices"`
	NextCursor string    `json:"next_cursor,omitempty"`
}

// UsageSummary represents current usage.
type UsageSummary struct {
	Period    UsagePeriod            `json:"period"`
	TotalCost string                 `json:"total_cost"`
	Currency  string                 `json:"currency"`
	Resources *PortalResourceMetrics `json:"resources,omitempty"`
}

// UsageHistoryResponse is a time series of usage.
type UsageHistoryResponse struct {
	Series []UsageHistoryPoint `json:"series"`
}

// UsageHistoryPoint represents usage at a point in time.
type UsageHistoryPoint struct {
	Period    UsagePeriod            `json:"period"`
	TotalCost string                 `json:"total_cost"`
	Currency  string                 `json:"currency"`
	Resources *PortalResourceMetrics `json:"resources,omitempty"`
}

// UsagePeriod defines start/end for a usage window.
type UsagePeriod struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// UsageMetric represents a resource usage metric.
type UsageMetric struct {
	Usage float64 `json:"usage"`
	Limit float64 `json:"limit"`
	Unit  string  `json:"unit"`
}

// NetworkMetric represents network usage metrics.
type NetworkMetric struct {
	RxBytes float64 `json:"rx_bytes"`
	TxBytes float64 `json:"tx_bytes"`
}

// PortalResourceMetrics is a portal-facing resource metric payload.
type PortalResourceMetrics struct {
	CPU       *UsageMetric   `json:"cpu,omitempty"`
	Memory    *UsageMetric   `json:"memory,omitempty"`
	Storage   *UsageMetric   `json:"storage,omitempty"`
	Network   *NetworkMetric `json:"network,omitempty"`
	Timestamp *time.Time     `json:"timestamp,omitempty"`
}

// MetricsPoint is a single metrics sample.
type MetricsPoint struct {
	Timestamp time.Time              `json:"timestamp"`
	Metrics   *PortalResourceMetrics `json:"metrics,omitempty"`
}

// MetricsSeriesResponse is a time series response.
type MetricsSeriesResponse struct {
	Series []MetricsPoint `json:"series"`
}

// DeploymentEvent represents a deployment event.
type DeploymentEvent struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// DeploymentEventListResponse is a paginated event list.
type DeploymentEventListResponse struct {
	Events     []DeploymentEvent `json:"events"`
	NextCursor string            `json:"next_cursor,omitempty"`
}

// ProviderInfo is the public provider identity.
type ProviderInfo struct {
	Address      string            `json:"address"`
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	ChainID      string            `json:"chain_id"`
	Capabilities []string          `json:"capabilities,omitempty"`
	Region       string            `json:"region"`
	Endpoints    ProviderEndpoints `json:"endpoints"`
}

// ProviderEndpoints lists exposed endpoints.
type ProviderEndpoints struct {
	REST      string `json:"rest"`
	WebSocket string `json:"websocket"`
	GRPC      string `json:"grpc,omitempty"`
}

// ProviderPricing describes provider pricing.
type ProviderPricing struct {
	Currency string        `json:"currency"`
	CPU      ResourcePrice `json:"cpu"`
	Memory   ResourcePrice `json:"memory"`
	Storage  ResourcePrice `json:"storage"`
	GPU      ResourcePrice `json:"gpu"`
}

// ResourcePrice represents pricing for a resource.
type ResourcePrice struct {
	Unit     string `json:"unit"`
	Price    string `json:"price"`
	Interval string `json:"interval"`
}

// ProviderCapacity describes available capacity.
type ProviderCapacity struct {
	CPUCores  float64 `json:"cpu_cores"`
	MemoryGB  float64 `json:"memory_gb"`
	StorageGB float64 `json:"storage_gb"`
	GPUUnits  float64 `json:"gpu_units"`
}

// ProviderAttributes wraps provider attributes.
type ProviderAttributes struct {
	Attributes map[string]interface{} `json:"attributes"`
}
