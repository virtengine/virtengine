# Task 29F: Enhanced portal_api.go Endpoints

**ID:** 29F  
**Title:** feat(provider): Enhanced portal_api.go endpoints  
**Priority:** P1 (High)  
**Wave:** 2 (Parallel with 29D)  
**Estimated LOC:** ~2000  
**Dependencies:** None  
**Blocking:** 29I (Support ticket flow), 29L (OpenAPI docs)  

---

## Problem Statement

The existing `portal_api.go` provides basic deployment management endpoints, but lacks the full Waldur-like functionality needed for a complete provider portal experience:

1. **No organization management** - Can't manage teams/groups
2. **No support ticket endpoints** - Can't create/view support tickets
3. **No billing endpoints** - Can't view invoices or usage
4. **Limited metrics** - No historical data or aggregations
5. **No provider info** - Can't discover provider capabilities

---

## Acceptance Criteria

### AC-1: Organization Management Endpoints
- [ ] `GET /api/v1/organizations` - List user's organizations
- [ ] `GET /api/v1/organizations/:id` - Get organization details
- [ ] `GET /api/v1/organizations/:id/members` - List members
- [ ] `POST /api/v1/organizations/:id/invite` - Invite member
- [ ] `DELETE /api/v1/organizations/:id/members/:address` - Remove member
- [ ] Query x/group module for organization data

### AC-2: Support Ticket Endpoints
- [ ] `GET /api/v1/tickets` - List tickets for user's deployments
- [ ] `POST /api/v1/tickets` - Create new support ticket
- [ ] `GET /api/v1/tickets/:id` - Get ticket details
- [ ] `POST /api/v1/tickets/:id/comments` - Add comment
- [ ] `PATCH /api/v1/tickets/:id` - Update ticket status
- [ ] Query x/support module for ticket data

### AC-3: Billing Endpoints
- [ ] `GET /api/v1/invoices` - List invoices for user
- [ ] `GET /api/v1/invoices/:id` - Get invoice details
- [ ] `GET /api/v1/usage` - Get current usage metrics
- [ ] `GET /api/v1/usage/history` - Historical usage data
- [ ] Query x/escrow module for billing data

### AC-4: Enhanced Metrics Endpoints
- [ ] `GET /api/v1/deployments/:id/metrics/history` - Historical metrics
- [ ] `GET /api/v1/deployments/:id/events` - Deployment events
- [ ] `GET /api/v1/metrics/aggregate` - Aggregated metrics across deployments
- [ ] Support time range parameters

### AC-5: Provider Info Endpoints
- [ ] `GET /api/v1/provider/info` - Provider capabilities and version
- [ ] `GET /api/v1/provider/pricing` - Current pricing information
- [ ] `GET /api/v1/provider/capacity` - Available capacity
- [ ] `GET /api/v1/provider/attributes` - Provider attributes

### AC-6: OpenAPI Annotations
- [ ] Add OpenAPI/Swagger annotations to all endpoints
- [ ] Generate OpenAPI spec from code
- [ ] Document request/response schemas
- [ ] Document authentication requirements

### AC-7: Pagination Support
- [ ] Implement cursor-based pagination
- [ ] Support `limit` and `cursor` parameters
- [ ] Return `next_cursor` in responses
- [ ] Handle large result sets efficiently

### AC-8: Rate Limiting
- [ ] Implement per-user rate limiting
- [ ] Return `X-RateLimit-*` headers
- [ ] 429 response when limit exceeded
- [ ] Configurable limits per endpoint

---

## Technical Requirements

### Organization Endpoints

```go
// pkg/provider_daemon/handlers/organizations.go
package handlers

import (
    "encoding/json"
    "net/http"
    
    "github.com/go-chi/chi/v5"
)

// Organization represents a group/organization
type Organization struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    Description string   `json:"description,omitempty"`
    Members     []Member `json:"members,omitempty"`
    CreatedAt   string   `json:"created_at"`
}

type Member struct {
    Address string `json:"address"`
    Role    string `json:"role"` // admin, member, viewer
    JoinedAt string `json:"joined_at"`
}

type InviteRequest struct {
    Address string `json:"address"`
    Role    string `json:"role"`
}

// OrganizationHandler handles organization-related endpoints
type OrganizationHandler struct {
    chainQuery ChainQuerier
}

func NewOrganizationHandler(chainQuery ChainQuerier) *OrganizationHandler {
    return &OrganizationHandler{chainQuery: chainQuery}
}

// @Summary List user's organizations
// @Description Returns all organizations the authenticated user belongs to
// @Tags organizations
// @Security WalletAuth
// @Produce json
// @Success 200 {array} Organization
// @Router /api/v1/organizations [get]
func (h *OrganizationHandler) ListOrganizations(w http.ResponseWriter, r *http.Request) {
    address := r.Context().Value("address").(string)
    
    // Query x/group module for user's groups
    groups, err := h.chainQuery.GetGroupsByMember(address)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    orgs := make([]Organization, 0, len(groups))
    for _, g := range groups {
        orgs = append(orgs, Organization{
            ID:          g.ID,
            Name:        g.Metadata.Name,
            Description: g.Metadata.Description,
            CreatedAt:   g.CreatedAt.Format(time.RFC3339),
        })
    }
    
    json.NewEncoder(w).Encode(orgs)
}

// @Summary Get organization members
// @Tags organizations
// @Security WalletAuth
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {array} Member
// @Router /api/v1/organizations/{id}/members [get]
func (h *OrganizationHandler) GetMembers(w http.ResponseWriter, r *http.Request) {
    orgID := chi.URLParam(r, "id")
    
    members, err := h.chainQuery.GetGroupMembers(orgID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(members)
}

// @Summary Invite member to organization
// @Tags organizations
// @Security WalletAuth
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Param invite body InviteRequest true "Invite details"
// @Success 200 {object} Member
// @Router /api/v1/organizations/{id}/invite [post]
func (h *OrganizationHandler) InviteMember(w http.ResponseWriter, r *http.Request) {
    orgID := chi.URLParam(r, "id")
    address := r.Context().Value("address").(string)
    
    // Verify caller is admin of organization
    isAdmin, err := h.chainQuery.IsGroupAdmin(orgID, address)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if !isAdmin {
        http.Error(w, "not authorized to invite members", http.StatusForbidden)
        return
    }
    
    var req InviteRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }
    
    // Submit x/group MsgUpdateGroupMembers transaction
    // This would typically be signed by the provider's key
    // and submitted on behalf of the organization admin
    
    member := Member{
        Address:  req.Address,
        Role:     req.Role,
        JoinedAt: time.Now().Format(time.RFC3339),
    }
    
    json.NewEncoder(w).Encode(member)
}
```

### Support Ticket Endpoints

```go
// pkg/provider_daemon/handlers/tickets.go
package handlers

type Ticket struct {
    ID           string    `json:"id"`
    DeploymentID string    `json:"deployment_id"`
    Subject      string    `json:"subject"`
    Description  string    `json:"description"`
    Status       string    `json:"status"` // open, in_progress, resolved, closed
    Priority     string    `json:"priority"` // low, medium, high, critical
    CreatedBy    string    `json:"created_by"`
    CreatedAt    string    `json:"created_at"`
    UpdatedAt    string    `json:"updated_at"`
    Comments     []Comment `json:"comments,omitempty"`
}

type Comment struct {
    ID        string `json:"id"`
    Author    string `json:"author"`
    Content   string `json:"content"`
    CreatedAt string `json:"created_at"`
}

type CreateTicketRequest struct {
    DeploymentID string `json:"deployment_id"`
    Subject      string `json:"subject"`
    Description  string `json:"description"`
    Priority     string `json:"priority"`
}

type AddCommentRequest struct {
    Content string `json:"content"`
}

type TicketHandler struct {
    chainQuery ChainQuerier
}

// @Summary List support tickets
// @Tags tickets
// @Security WalletAuth
// @Produce json
// @Param status query string false "Filter by status"
// @Param deployment_id query string false "Filter by deployment"
// @Success 200 {array} Ticket
// @Router /api/v1/tickets [get]
func (h *TicketHandler) ListTickets(w http.ResponseWriter, r *http.Request) {
    address := r.Context().Value("address").(string)
    status := r.URL.Query().Get("status")
    deploymentID := r.URL.Query().Get("deployment_id")
    
    // Query x/support module
    tickets, err := h.chainQuery.GetTicketsByUser(address, status, deploymentID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(tickets)
}

// @Summary Create support ticket
// @Tags tickets
// @Security WalletAuth
// @Accept json
// @Produce json
// @Param ticket body CreateTicketRequest true "Ticket details"
// @Success 201 {object} Ticket
// @Router /api/v1/tickets [post]
func (h *TicketHandler) CreateTicket(w http.ResponseWriter, r *http.Request) {
    address := r.Context().Value("address").(string)
    
    var req CreateTicketRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }
    
    // Verify user owns the deployment
    if err := h.verifyDeploymentOwnership(address, req.DeploymentID); err != nil {
        http.Error(w, err.Error(), http.StatusForbidden)
        return
    }
    
    // Submit to x/support module
    ticket, err := h.chainQuery.CreateTicket(req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(ticket)
}

// @Summary Add comment to ticket
// @Tags tickets
// @Security WalletAuth
// @Accept json
// @Produce json
// @Param id path string true "Ticket ID"
// @Param comment body AddCommentRequest true "Comment content"
// @Success 201 {object} Comment
// @Router /api/v1/tickets/{id}/comments [post]
func (h *TicketHandler) AddComment(w http.ResponseWriter, r *http.Request) {
    ticketID := chi.URLParam(r, "id")
    address := r.Context().Value("address").(string)
    
    var req AddCommentRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }
    
    comment, err := h.chainQuery.AddTicketComment(ticketID, address, req.Content)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(comment)
}
```

### Billing Endpoints

```go
// pkg/provider_daemon/handlers/billing.go
package handlers

type Invoice struct {
    ID           string        `json:"id"`
    LeaseID      string        `json:"lease_id"`
    Period       InvoicePeriod `json:"period"`
    Amount       string        `json:"amount"`
    Currency     string        `json:"currency"`
    Status       string        `json:"status"` // pending, paid, overdue
    DueDate      string        `json:"due_date"`
    PaidAt       string        `json:"paid_at,omitempty"`
    LineItems    []LineItem    `json:"line_items"`
}

type InvoicePeriod struct {
    Start string `json:"start"`
    End   string `json:"end"`
}

type LineItem struct {
    Description string `json:"description"`
    Quantity    string `json:"quantity"`
    UnitPrice   string `json:"unit_price"`
    Total       string `json:"total"`
}

type UsageSummary struct {
    LeaseID   string         `json:"lease_id"`
    Period    InvoicePeriod  `json:"period"`
    Resources ResourceUsage  `json:"resources"`
    Cost      CostBreakdown  `json:"cost"`
}

type ResourceUsage struct {
    CPU        UsageMetric `json:"cpu"`
    Memory     UsageMetric `json:"memory"`
    Storage    UsageMetric `json:"storage"`
    Bandwidth  UsageMetric `json:"bandwidth"`
    GPUSeconds int64       `json:"gpu_seconds,omitempty"`
}

type UsageMetric struct {
    Used    string `json:"used"`
    Limit   string `json:"limit"`
    Percent float64 `json:"percent"`
}

type CostBreakdown struct {
    CPU       string `json:"cpu"`
    Memory    string `json:"memory"`
    Storage   string `json:"storage"`
    Bandwidth string `json:"bandwidth"`
    GPU       string `json:"gpu,omitempty"`
    Total     string `json:"total"`
}

type BillingHandler struct {
    chainQuery ChainQuerier
}

// @Summary List invoices
// @Tags billing
// @Security WalletAuth
// @Produce json
// @Param status query string false "Filter by status"
// @Param limit query int false "Number of results"
// @Param cursor query string false "Pagination cursor"
// @Success 200 {object} InvoiceListResponse
// @Router /api/v1/invoices [get]
func (h *BillingHandler) ListInvoices(w http.ResponseWriter, r *http.Request) {
    address := r.Context().Value("address").(string)
    limit := parseIntParam(r, "limit", 20)
    cursor := r.URL.Query().Get("cursor")
    status := r.URL.Query().Get("status")
    
    // Query x/escrow module for invoices
    invoices, nextCursor, err := h.chainQuery.GetInvoices(address, status, limit, cursor)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    response := map[string]interface{}{
        "invoices":    invoices,
        "next_cursor": nextCursor,
    }
    
    json.NewEncoder(w).Encode(response)
}

// @Summary Get current usage
// @Tags billing
// @Security WalletAuth
// @Produce json
// @Success 200 {array} UsageSummary
// @Router /api/v1/usage [get]
func (h *BillingHandler) GetCurrentUsage(w http.ResponseWriter, r *http.Request) {
    address := r.Context().Value("address").(string)
    
    // Get all active leases for user
    leases, err := h.chainQuery.GetActiveLeases(address)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    summaries := make([]UsageSummary, 0, len(leases))
    for _, lease := range leases {
        usage, err := h.calculateUsage(lease.ID)
        if err != nil {
            continue // Skip on error
        }
        summaries = append(summaries, usage)
    }
    
    json.NewEncoder(w).Encode(summaries)
}

// @Summary Get usage history
// @Tags billing
// @Security WalletAuth
// @Produce json
// @Param start query string true "Start date (RFC3339)"
// @Param end query string true "End date (RFC3339)"
// @Param granularity query string false "Data granularity (hour, day, week)"
// @Success 200 {object} UsageHistoryResponse
// @Router /api/v1/usage/history [get]
func (h *BillingHandler) GetUsageHistory(w http.ResponseWriter, r *http.Request) {
    address := r.Context().Value("address").(string)
    
    start, err := time.Parse(time.RFC3339, r.URL.Query().Get("start"))
    if err != nil {
        http.Error(w, "invalid start date", http.StatusBadRequest)
        return
    }
    
    end, err := time.Parse(time.RFC3339, r.URL.Query().Get("end"))
    if err != nil {
        http.Error(w, "invalid end date", http.StatusBadRequest)
        return
    }
    
    granularity := r.URL.Query().Get("granularity")
    if granularity == "" {
        granularity = "day"
    }
    
    history, err := h.chainQuery.GetUsageHistory(address, start, end, granularity)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(history)
}
```

### Provider Info Endpoints

```go
// pkg/provider_daemon/handlers/provider.go
package handlers

type ProviderInfo struct {
    Address     string            `json:"address"`
    Name        string            `json:"name"`
    Version     string            `json:"version"`
    Attributes  map[string]string `json:"attributes"`
    Capabilities []string         `json:"capabilities"`
    Region      string            `json:"region"`
    Status      string            `json:"status"`
}

type Pricing struct {
    CPU       ResourcePrice `json:"cpu"`
    Memory    ResourcePrice `json:"memory"`
    Storage   ResourcePrice `json:"storage"`
    Bandwidth ResourcePrice `json:"bandwidth"`
    GPU       *GPUPricing   `json:"gpu,omitempty"`
    Currency  string        `json:"currency"`
    UpdatedAt string        `json:"updated_at"`
}

type ResourcePrice struct {
    Unit     string `json:"unit"`
    Price    string `json:"price"`
    Interval string `json:"interval"` // hourly, daily, monthly
}

type GPUPricing struct {
    Models []GPUModelPrice `json:"models"`
}

type GPUModelPrice struct {
    Model string `json:"model"`
    Price string `json:"price"`
}

type Capacity struct {
    Available CapacityMetrics `json:"available"`
    Total     CapacityMetrics `json:"total"`
    Reserved  CapacityMetrics `json:"reserved"`
}

type CapacityMetrics struct {
    CPU     string `json:"cpu"`
    Memory  string `json:"memory"`
    Storage string `json:"storage"`
    GPUs    int    `json:"gpus"`
}

type ProviderHandler struct {
    provider *Provider
}

// @Summary Get provider info
// @Tags provider
// @Produce json
// @Success 200 {object} ProviderInfo
// @Router /api/v1/provider/info [get]
func (h *ProviderHandler) GetInfo(w http.ResponseWriter, r *http.Request) {
    info := ProviderInfo{
        Address:      h.provider.Address,
        Name:         h.provider.Name,
        Version:      h.provider.Version,
        Attributes:   h.provider.Attributes,
        Capabilities: h.provider.Capabilities,
        Region:       h.provider.Region,
        Status:       h.provider.Status,
    }
    
    json.NewEncoder(w).Encode(info)
}

// @Summary Get provider pricing
// @Tags provider
// @Produce json
// @Success 200 {object} Pricing
// @Router /api/v1/provider/pricing [get]
func (h *ProviderHandler) GetPricing(w http.ResponseWriter, r *http.Request) {
    pricing := h.provider.GetCurrentPricing()
    json.NewEncoder(w).Encode(pricing)
}

// @Summary Get provider capacity
// @Tags provider
// @Produce json
// @Success 200 {object} Capacity
// @Router /api/v1/provider/capacity [get]
func (h *ProviderHandler) GetCapacity(w http.ResponseWriter, r *http.Request) {
    capacity := h.provider.GetCapacity()
    json.NewEncoder(w).Encode(capacity)
}
```

### Router Registration

```go
// pkg/provider_daemon/portal_api.go (updated)

func (s *PortalAPIServer) setupRoutes(r chi.Router) {
    // Existing routes...
    
    // Organization routes
    orgHandler := handlers.NewOrganizationHandler(s.chainQuery)
    r.Route("/api/v1/organizations", func(r chi.Router) {
        r.Use(s.authMiddleware(true))
        r.Get("/", orgHandler.ListOrganizations)
        r.Get("/{id}", orgHandler.GetOrganization)
        r.Get("/{id}/members", orgHandler.GetMembers)
        r.Post("/{id}/invite", orgHandler.InviteMember)
        r.Delete("/{id}/members/{address}", orgHandler.RemoveMember)
    })
    
    // Ticket routes
    ticketHandler := handlers.NewTicketHandler(s.chainQuery)
    r.Route("/api/v1/tickets", func(r chi.Router) {
        r.Use(s.authMiddleware(true))
        r.Get("/", ticketHandler.ListTickets)
        r.Post("/", ticketHandler.CreateTicket)
        r.Get("/{id}", ticketHandler.GetTicket)
        r.Post("/{id}/comments", ticketHandler.AddComment)
        r.Patch("/{id}", ticketHandler.UpdateTicket)
    })
    
    // Billing routes
    billingHandler := handlers.NewBillingHandler(s.chainQuery)
    r.Route("/api/v1/invoices", func(r chi.Router) {
        r.Use(s.authMiddleware(true))
        r.Get("/", billingHandler.ListInvoices)
        r.Get("/{id}", billingHandler.GetInvoice)
    })
    r.Route("/api/v1/usage", func(r chi.Router) {
        r.Use(s.authMiddleware(true))
        r.Get("/", billingHandler.GetCurrentUsage)
        r.Get("/history", billingHandler.GetUsageHistory)
    })
    
    // Provider info routes (public)
    providerHandler := handlers.NewProviderHandler(s.provider)
    r.Route("/api/v1/provider", func(r chi.Router) {
        r.Get("/info", providerHandler.GetInfo)
        r.Get("/pricing", providerHandler.GetPricing)
        r.Get("/capacity", providerHandler.GetCapacity)
        r.Get("/attributes", providerHandler.GetAttributes)
    })
}
```

---

## Files to Create

| Path | Description | Est. Lines |
|------|-------------|------------|
| `pkg/provider_daemon/handlers/organizations.go` | Org endpoints | 300 |
| `pkg/provider_daemon/handlers/tickets.go` | Ticket endpoints | 350 |
| `pkg/provider_daemon/handlers/billing.go` | Billing endpoints | 350 |
| `pkg/provider_daemon/handlers/provider.go` | Provider info | 200 |
| `pkg/provider_daemon/handlers/metrics.go` | Enhanced metrics | 250 |
| `pkg/provider_daemon/middleware/ratelimit.go` | Rate limiting | 150 |
| `pkg/provider_daemon/middleware/pagination.go` | Pagination helpers | 100 |
| `pkg/provider_daemon/chain_query.go` | Chain query interface | 200 |
| `api/openapi/portal_api.yaml` | OpenAPI spec | 500 |

**Total: ~2400 lines**

---

## Validation Checklist

- [ ] All endpoints return valid JSON
- [ ] Authentication works on protected endpoints
- [ ] Public endpoints accessible without auth
- [ ] Pagination works correctly
- [ ] Rate limiting enforced
- [ ] OpenAPI spec generated
- [ ] All tests pass

---

## Vibe-Kanban Task ID

`4a2fad1d-9070-4363-bc05-8fe0a9ef5fe9`
