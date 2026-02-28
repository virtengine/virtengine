// Package waldur provides Marketplace operations via Waldur API
//
// VE-2024: Marketplace offerings, orders, and resources management
package waldur

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	client "github.com/waldur/go-client"
)

// MarketplaceClient provides marketplace operations through Waldur
type MarketplaceClient struct {
	client *Client
}

type offeringResponse struct {
	UUID        string    `json:"uuid"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	State       string    `json:"state"`
	Category    string    `json:"category"`
	Shared      bool      `json:"shared"`
	Billable    bool      `json:"billable"`
	CreatedAt   time.Time `json:"created"`
	BackendID   string    `json:"backend_id"`
}

// NewMarketplaceClient creates a new marketplace client
func NewMarketplaceClient(c *Client) *MarketplaceClient {
	return &MarketplaceClient{client: c}
}

// Offering represents a marketplace offering in Waldur
type Offering struct {
	UUID        string    `json:"uuid"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	State       string    `json:"state"`
	Category    string    `json:"category"`
	Shared      bool      `json:"shared"`
	Billable    bool      `json:"billable"`
	CreatedAt   time.Time `json:"created"`
}

// Order represents a marketplace order in Waldur
type Order struct {
	UUID         string    `json:"uuid"`
	State        string    `json:"state"`
	Type         string    `json:"type"`
	ProjectUUID  string    `json:"project_uuid"`
	CreatedAt    time.Time `json:"created"`
	ErrorMessage string    `json:"error_message"`
	ResourceUUID string    `json:"resource_uuid,omitempty"`
}

// Resource represents a marketplace resource in Waldur
type Resource struct {
	UUID         string    `json:"uuid"`
	Name         string    `json:"name"`
	State        string    `json:"state"`
	OfferingUUID string    `json:"offering_uuid"`
	ProjectUUID  string    `json:"project_uuid"`
	ResourceType string    `json:"resource_type"`
	CreatedAt    time.Time `json:"created"`
}

// CreateOrderRequest contains parameters for creating a marketplace order.
type CreateOrderRequest struct {
	OfferingUUID   string
	ProjectUUID    string
	CallbackURL    string
	RequestComment string
	Attributes     map[string]interface{}
	Limits         map[string]int
	Type           string
	Name           string
	Description    string
}

// ListOfferingsParams contains parameters for listing offerings
type ListOfferingsParams struct {
	CustomerUUID string
	CategoryUUID string
	Type         string
	State        string
	Shared       *bool
	Name         string
	Page         int
	PageSize     int
}

// ListOfferings lists marketplace offerings via Waldur
func (m *MarketplaceClient) ListOfferings(ctx context.Context, params ListOfferingsParams) ([]Offering, error) {
	query := url.Values{}
	if params.CustomerUUID != "" {
		query.Set("customer_uuid", params.CustomerUUID)
	}
	if params.CategoryUUID != "" {
		query.Set("category_uuid", params.CategoryUUID)
	}
	if params.Type != "" {
		query.Set("type", params.Type)
	}
	if params.State != "" {
		query.Set("state", params.State)
	}
	if params.Name != "" {
		query.Set("name", params.Name)
	}
	if params.Shared != nil {
		query.Set("shared", fmt.Sprintf("%t", *params.Shared))
	}
	if params.Page > 0 {
		query.Set("page", fmt.Sprintf("%d", params.Page))
	}
	if params.PageSize > 0 {
		query.Set("page_size", fmt.Sprintf("%d", params.PageSize))
	}

	var raw []offeringResponse
	if err := m.client.doWithRetry(ctx, func() error {
		return m.client.getPaginated(ctx, buildQueryPath("/marketplace-public-offerings/", query), &raw)
	}); err != nil {
		return nil, err
	}

	offerings := make([]Offering, 0, len(raw))
	for _, item := range raw {
		offerings = append(offerings, Offering{
			UUID:        item.UUID,
			Name:        item.Name,
			Description: item.Description,
			Type:        item.Type,
			State:       item.State,
			Category:    item.Category,
			Shared:      item.Shared,
			Billable:    item.Billable,
			CreatedAt:   item.CreatedAt,
		})
	}
	return offerings, nil
}

// GetOffering retrieves a specific marketplace offering
func (m *MarketplaceClient) GetOffering(ctx context.Context, offeringUUID string) (*Offering, error) {
	var offering *Offering

	err := m.client.doWithRetry(ctx, func() error {
		uuidType, err := parseUUIDParam("offering UUID", offeringUUID)
		if err != nil {
			return err
		}
		resp, err := m.client.api.MarketplacePublicOfferingsRetrieveWithResponse(ctx, uuidType, nil)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		if resp.JSON200 == nil {
			return ErrInvalidResponse
		}

		o := resp.JSON200
		offering = &Offering{
			Name:        safeString(o.Name),
			Description: safeString(o.Description),
			Type:        safeString(o.Type),
		}
		if o.Uuid != nil {
			offering.UUID = o.Uuid.String()
		}
		if o.State != nil {
			offering.State = string(*o.State)
		}
		if o.Shared != nil {
			offering.Shared = *o.Shared
		}
		if o.Billable != nil {
			offering.Billable = *o.Billable
		}
		if o.Created != nil {
			offering.CreatedAt = *o.Created
		}

		return nil
	})

	return offering, err
}

// ListOrdersParams contains parameters for listing orders
type ListOrdersParams struct {
	ProjectUUID  string
	CustomerUUID string
	OfferingUUID string
	State        string
	Page         int
	PageSize     int
}

// ListOrders lists marketplace orders via Waldur
func (m *MarketplaceClient) ListOrders(ctx context.Context, params ListOrdersParams) ([]Order, error) {
	query := url.Values{}
	if params.ProjectUUID != "" {
		query.Set("project_uuid", params.ProjectUUID)
	}
	if params.CustomerUUID != "" {
		query.Set("customer_uuid", params.CustomerUUID)
	}
	if params.OfferingUUID != "" {
		query.Set("offering_uuid", params.OfferingUUID)
	}
	if params.State != "" {
		query.Set("state", params.State)
	}
	if params.Page > 0 {
		query.Set("page", fmt.Sprintf("%d", params.Page))
	}
	if params.PageSize > 0 {
		query.Set("page_size", fmt.Sprintf("%d", params.PageSize))
	}

	var orders []Order
	if err := m.client.doWithRetry(ctx, func() error {
		return m.client.getPaginated(ctx, buildQueryPath("/marketplace-orders/", query), &orders)
	}); err != nil {
		return nil, err
	}
	return orders, nil
}

// GetOrder retrieves a specific marketplace order
func (m *MarketplaceClient) GetOrder(ctx context.Context, orderUUID string) (*Order, error) {
	var order *Order

	err := m.client.doWithRetry(ctx, func() error {
		uuidType, err := parseUUIDParam("order UUID", orderUUID)
		if err != nil {
			return err
		}
		resp, err := m.client.api.MarketplaceOrdersRetrieveWithResponse(ctx, uuidType, nil)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		if resp.JSON200 == nil {
			return ErrInvalidResponse
		}

		o := resp.JSON200
		order = &Order{
			ErrorMessage: safeString(o.ErrorMessage),
		}
		if o.Uuid != nil {
			order.UUID = o.Uuid.String()
		}
		if o.State != nil {
			order.State = string(*o.State)
		}
		if o.Type != nil {
			order.Type = string(*o.Type)
		}
		if o.ProjectUuid != nil {
			order.ProjectUUID = o.ProjectUuid.String()
		}
		if o.Created != nil {
			order.CreatedAt = *o.Created
		}

		return nil
	})

	return order, err
}

// CreateOrder creates a marketplace order via Waldur.
func (m *MarketplaceClient) CreateOrder(ctx context.Context, req CreateOrderRequest) (*Order, error) {
	if req.OfferingUUID == "" || req.ProjectUUID == "" {
		return nil, fmt.Errorf("offering UUID and project UUID are required")
	}

	var order *Order

	err := m.client.doWithRetry(ctx, func() error {
		body := client.MarketplaceOrdersCreateJSONRequestBody{
			Offering: req.OfferingUUID,
			Project:  req.ProjectUUID,
		}

		if req.CallbackURL != "" {
			body.CallbackUrl = &req.CallbackURL
		}
		if req.RequestComment != "" {
			body.RequestComment = &req.RequestComment
		}
		if len(req.Limits) > 0 {
			body.Limits = &req.Limits
		}
		if req.Type != "" {
			orderType := client.RequestTypes(req.Type)
			body.Type = &orderType
		}

		if req.Name != "" || req.Description != "" || len(req.Attributes) > 0 {
			attrs := client.GenericOrderAttributes{
				AdditionalProperties: map[string]interface{}{},
			}
			if req.Name != "" {
				attrs.Name = &req.Name
			}
			if req.Description != "" {
				attrs.Description = &req.Description
			}
			for key, value := range req.Attributes {
				attrs.AdditionalProperties[key] = value
			}
			orderAttrs := client.OrderCreateRequest_Attributes{}
			if err := orderAttrs.FromGenericOrderAttributes(attrs); err != nil {
				return err
			}
			body.Attributes = &orderAttrs
		}

		resp, err := m.client.api.MarketplaceOrdersCreateWithResponse(ctx, body)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusCreated {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		if resp.JSON201 == nil {
			return ErrInvalidResponse
		}

		order = mapOrderDetails(resp.JSON201)
		return nil
	})

	return order, err
}

// ApproveOrderByProvider approves a pending order as provider.
func (m *MarketplaceClient) ApproveOrderByProvider(ctx context.Context, orderUUID string) error {
	if orderUUID == "" {
		return fmt.Errorf("order UUID is required")
	}

	return m.client.doWithRetry(ctx, func() error {
		uuidType, err := parseUUIDParam("order UUID", orderUUID)
		if err != nil {
			return err
		}
		resp, err := m.client.api.MarketplaceOrdersApproveByProviderWithResponse(ctx, uuidType)
		if err != nil {
			return err
		}
		if resp.StatusCode() != http.StatusOK {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}
		return nil
	})
}

// SetOrderBackendID sets backend ID for the order.
func (m *MarketplaceClient) SetOrderBackendID(ctx context.Context, orderUUID string, backendID string) error {
	if orderUUID == "" || backendID == "" {
		return fmt.Errorf("order UUID and backend ID are required")
	}

	return m.client.doWithRetry(ctx, func() error {
		uuidType, err := parseUUIDParam("order UUID", orderUUID)
		if err != nil {
			return err
		}
		body := client.MarketplaceOrdersSetBackendIdJSONRequestBody{
			BackendId: &backendID,
		}
		resp, err := m.client.api.MarketplaceOrdersSetBackendIdWithResponse(ctx, uuidType, body)
		if err != nil {
			return err
		}
		if resp.StatusCode() != http.StatusOK {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}
		return nil
	})
}

// TerminateResource requests termination of a marketplace resource.
func (m *MarketplaceClient) TerminateResource(ctx context.Context, resourceUUID string, attributes map[string]interface{}) error {
	if resourceUUID == "" {
		return fmt.Errorf("resource UUID is required")
	}

	return m.client.doWithRetry(ctx, func() error {
		uuidType, err := parseUUIDParam("resource UUID", resourceUUID)
		if err != nil {
			return err
		}
		body := client.MarketplaceResourcesTerminateJSONRequestBody{}
		if len(attributes) > 0 {
			body.Attributes = attributes
		}

		resp, err := m.client.api.MarketplaceResourcesTerminateWithResponse(ctx, uuidType, body)
		if err != nil {
			return err
		}
		if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusAccepted {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}
		return nil
	})
}

func mapOrderDetails(details *client.OrderDetails) *Order {
	if details == nil {
		return nil
	}

	order := &Order{}
	if details.Uuid != nil {
		order.UUID = details.Uuid.String()
	}
	if details.State != nil {
		order.State = string(*details.State)
	}
	if details.Type != nil {
		order.Type = string(*details.Type)
	}
	if details.ProjectUuid != nil {
		order.ProjectUUID = details.ProjectUuid.String()
	}
	if details.Created != nil {
		order.CreatedAt = *details.Created
	}
	if details.ErrorMessage != nil {
		order.ErrorMessage = *details.ErrorMessage
	}
	if details.MarketplaceResourceUuid != nil {
		order.ResourceUUID = details.MarketplaceResourceUuid.String()
	} else if details.ResourceUuid != nil {
		order.ResourceUUID = details.ResourceUuid.String()
	}
	return order
}

// ListResourcesParams contains parameters for listing resources
type ListResourcesParams struct {
	ProjectUUID  string
	CustomerUUID string
	OfferingUUID string
	State        string
	Page         int
	PageSize     int
}

// ListResources lists marketplace resources via Waldur
func (m *MarketplaceClient) ListResources(ctx context.Context, params ListResourcesParams) ([]Resource, error) {
	query := url.Values{}
	if params.ProjectUUID != "" {
		query.Set("project_uuid", params.ProjectUUID)
	}
	if params.CustomerUUID != "" {
		query.Set("customer_uuid", params.CustomerUUID)
	}
	if params.OfferingUUID != "" {
		query.Set("offering_uuid", params.OfferingUUID)
	}
	if params.State != "" {
		query.Set("state", params.State)
	}
	if params.Page > 0 {
		query.Set("page", fmt.Sprintf("%d", params.Page))
	}
	if params.PageSize > 0 {
		query.Set("page_size", fmt.Sprintf("%d", params.PageSize))
	}

	var resources []Resource
	if err := m.client.doWithRetry(ctx, func() error {
		return m.client.getPaginated(ctx, buildQueryPath("/marketplace-resources/", query), &resources)
	}); err != nil {
		return nil, err
	}
	return resources, nil
}

// GetResource retrieves a specific marketplace resource
func (m *MarketplaceClient) GetResource(ctx context.Context, resourceUUID string) (*Resource, error) {
	var resource *Resource

	err := m.client.doWithRetry(ctx, func() error {
		uuidType, err := parseUUIDParam("resource UUID", resourceUUID)
		if err != nil {
			return err
		}
		resp, err := m.client.api.MarketplaceResourcesRetrieveWithResponse(ctx, uuidType, nil)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		if resp.JSON200 == nil {
			return ErrInvalidResponse
		}

		r := resp.JSON200
		resource = &Resource{
			Name:         safeString(r.Name),
			ResourceType: safeString(r.ResourceType),
		}
		if r.Uuid != nil {
			resource.UUID = r.Uuid.String()
		}
		if r.State != nil {
			resource.State = string(*r.State)
		}
		if r.OfferingUuid != nil {
			resource.OfferingUUID = r.OfferingUuid.String()
		}
		if r.ProjectUuid != nil {
			resource.ProjectUUID = r.ProjectUuid.String()
		}
		if r.Created != nil {
			resource.CreatedAt = *r.Created
		}

		return nil
	})

	return resource, err
}

// WaitForOrderCompletion waits for an order to complete
func (m *MarketplaceClient) WaitForOrderCompletion(ctx context.Context, orderUUID string, pollInterval time.Duration) (*Order, error) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			order, err := m.GetOrder(ctx, orderUUID)
			if err != nil {
				return nil, err
			}

			switch order.State {
			case opStateDone:
				return order, nil
			case "erred":
				return order, fmt.Errorf("order failed: %s", order.ErrorMessage)
			case opStateCanceled, "rejected":
				return order, fmt.Errorf("order %s", order.State)
			}
		}
	}
}

// ===============================================================================
// VE-2D: Offering CRUD Operations for Automatic Sync
// ===============================================================================

// CreateOfferingRequest contains parameters for creating a Waldur offering.
type CreateOfferingRequest struct {
	// Name is the offering name (max 255 chars).
	Name string `json:"name"`

	// Description is the offering description.
	Description string `json:"description,omitempty"`

	// Type is the offering type (e.g., "VirtEngine.Compute").
	Type string `json:"type"`

	// State is the initial state (Active, Paused, Archived).
	State string `json:"state,omitempty"`

	// CategoryUUID is the Waldur category UUID.
	CategoryUUID string `json:"category_uuid,omitempty"`

	// CustomerUUID is the Waldur customer UUID (provider organization).
	CustomerUUID string `json:"customer_uuid"`

	// Shared indicates if the offering is publicly visible.
	Shared bool `json:"shared"`

	// Billable indicates if the offering is billable.
	Billable bool `json:"billable"`

	// BackendID is the on-chain offering ID for cross-reference.
	BackendID string `json:"backend_id,omitempty"`

	// Attributes contains additional offering attributes.
	Attributes map[string]interface{} `json:"attributes,omitempty"`

	// Components contains pricing components.
	Components []PricingComponent `json:"components,omitempty"`
}

// UpdateOfferingRequest contains parameters for updating a Waldur offering.
type UpdateOfferingRequest struct {
	// Name is the updated offering name.
	Name string `json:"name,omitempty"`

	// Description is the updated description.
	Description string `json:"description,omitempty"`

	// Attributes contains updated attributes.
	Attributes map[string]interface{} `json:"attributes,omitempty"`

	// Components contains updated pricing components.
	Components []PricingComponent `json:"components,omitempty"`
}

// PricingComponent represents a Waldur pricing component.
type PricingComponent struct {
	// Type is the component type (usage, fixed, one_time).
	Type string `json:"type"`

	// Name is the component name.
	Name string `json:"name"`

	// MeasuredUnit is the unit of measurement.
	MeasuredUnit string `json:"measured_unit,omitempty"`

	// BillingType is how billing is calculated.
	BillingType string `json:"billing_type,omitempty"`

	// Price is the price per unit as a string.
	Price string `json:"price"`

	// MinValue is the minimum value.
	MinValue int64 `json:"min_value,omitempty"`

	// MaxValue is the maximum value.
	MaxValue int64 `json:"max_value,omitempty"`
}

// CreateOffering creates a new marketplace offering in Waldur.
// VE-2D: Automatic offering sync from chain to Waldur.
// Uses raw HTTP requests since provider offerings API may not be in the generated client.
func (m *MarketplaceClient) CreateOffering(ctx context.Context, req CreateOfferingRequest) (*Offering, error) {
	if req.Name == "" || req.CustomerUUID == "" {
		return nil, fmt.Errorf("name and customer UUID are required")
	}

	var offering *Offering

	err := m.client.doWithRetry(ctx, func() error {
		// Build the request body as a simple map
		body := map[string]interface{}{
			"name":     req.Name,
			"customer": req.CustomerUUID,
			"shared":   req.Shared,
			"billable": req.Billable,
		}

		if req.Description != "" {
			body["description"] = req.Description
		}
		if req.Type != "" {
			body["type"] = req.Type
		}
		if req.CategoryUUID != "" {
			body["category"] = req.CategoryUUID
		}
		if req.BackendID != "" {
			body["backend_id"] = req.BackendID
		}
		if len(req.Attributes) > 0 {
			body["attributes"] = req.Attributes
		}
		if len(req.Components) > 0 {
			body["components"] = req.Components
		}

		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}

		respBody, statusCode, err := m.client.doRequest(ctx, http.MethodPost, "/marketplace-provider-offerings/", bodyBytes)
		if err != nil {
			return err
		}

		if statusCode != http.StatusCreated && statusCode != http.StatusOK {
			return mapHTTPError(statusCode, respBody)
		}

		// Parse response
		var respData struct {
			UUID        string `json:"uuid"`
			Name        string `json:"name"`
			Description string `json:"description"`
			Type        string `json:"type"`
			State       string `json:"state"`
			Shared      bool   `json:"shared"`
			Billable    bool   `json:"billable"`
		}
		if err := json.Unmarshal(respBody, &respData); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}

		offering = &Offering{
			UUID:        respData.UUID,
			Name:        respData.Name,
			Description: respData.Description,
			Type:        respData.Type,
			State:       respData.State,
			Shared:      respData.Shared,
			Billable:    respData.Billable,
		}

		return nil
	})

	return offering, err
}

// UpdateOffering updates an existing marketplace offering in Waldur.
// VE-2D: Automatic offering sync from chain to Waldur.
func (m *MarketplaceClient) UpdateOffering(ctx context.Context, offeringUUID string, req UpdateOfferingRequest) (*Offering, error) {
	if offeringUUID == "" {
		return nil, fmt.Errorf("offering UUID is required")
	}

	var offering *Offering

	err := m.client.doWithRetry(ctx, func() error {
		// Build the request body as a simple map - only include non-empty fields
		body := make(map[string]interface{})

		if req.Name != "" {
			body["name"] = req.Name
		}
		if req.Description != "" {
			body["description"] = req.Description
		}
		if len(req.Attributes) > 0 {
			body["attributes"] = req.Attributes
		}
		if len(req.Components) > 0 {
			body["components"] = req.Components
		}

		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}

		path := fmt.Sprintf("/marketplace-provider-offerings/%s/", offeringUUID)
		respBody, statusCode, err := m.client.doRequest(ctx, http.MethodPatch, path, bodyBytes)
		if err != nil {
			return err
		}

		if statusCode != http.StatusOK {
			return mapHTTPError(statusCode, respBody)
		}

		// Parse response
		var respData struct {
			UUID        string `json:"uuid"`
			Name        string `json:"name"`
			Description string `json:"description"`
			Type        string `json:"type"`
			State       string `json:"state"`
			Shared      bool   `json:"shared"`
			Billable    bool   `json:"billable"`
		}
		if err := json.Unmarshal(respBody, &respData); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}

		offering = &Offering{
			UUID:        respData.UUID,
			Name:        respData.Name,
			Description: respData.Description,
			Type:        respData.Type,
			State:       respData.State,
			Shared:      respData.Shared,
			Billable:    respData.Billable,
		}

		return nil
	})

	return offering, err
}

// SetOfferingState changes the state of a Waldur offering.
// Valid actions: "activate", "pause", "archive".
// VE-2D: Automatic offering sync from chain to Waldur.
func (m *MarketplaceClient) SetOfferingState(ctx context.Context, offeringUUID string, action string) error {
	if offeringUUID == "" || action == "" {
		return fmt.Errorf("offering UUID and action are required")
	}

	validActions := map[string]bool{
		"activate": true,
		"pause":    true,
		"archive":  true,
	}
	if !validActions[action] {
		return fmt.Errorf("invalid offering state action: %s", action)
	}

	return m.client.doWithRetry(ctx, func() error {
		path := fmt.Sprintf("/marketplace-provider-offerings/%s/%s/", offeringUUID, action)
		respBody, statusCode, err := m.client.doRequest(ctx, http.MethodPost, path, nil)
		if err != nil {
			return err
		}

		if statusCode != http.StatusOK && statusCode != http.StatusNoContent && statusCode != http.StatusAccepted {
			return mapHTTPError(statusCode, respBody)
		}

		return nil
	})
}

// GetOfferingByBackendID finds a Waldur offering by its backend ID (on-chain offering ID).
// VE-2D: Used for sync reconciliation.
// Uses raw HTTP since the generated client may not support backend_id filter.
func (m *MarketplaceClient) GetOfferingByBackendID(ctx context.Context, backendID string) (*Offering, error) {
	if backendID == "" {
		return nil, fmt.Errorf("backend ID is required")
	}

	var offering *Offering

	err := m.client.doWithRetry(ctx, func() error {
		query := url.Values{}
		query.Set("backend_id", backendID)
		var offerings []offeringResponse
		if err := m.client.getPaginated(ctx, buildQueryPath("/marketplace-public-offerings/", query), &offerings); err != nil {
			return err
		}

		if len(offerings) == 0 {
			return ErrNotFound
		}
		if len(offerings) > 1 {
			return fmt.Errorf("%w: multiple offerings found for backend ID %s", ErrConflict, backendID)
		}

		o := offerings[0]
		offering = &Offering{
			UUID:        o.UUID,
			Name:        o.Name,
			Description: o.Description,
			Type:        o.Type,
			State:       o.State,
			Category:    o.Category,
			Shared:      o.Shared,
			Billable:    o.Billable,
			CreatedAt:   o.CreatedAt,
		}

		return nil
	})

	return offering, err
}

// ===============================================================================
// VE-25A: Category CRUD Operations for Automatic Sync
// ===============================================================================

// Category represents a marketplace category in Waldur.
type Category struct {
	UUID        string `json:"uuid"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Icon        string `json:"icon,omitempty"`
}

// ListCategoriesParams contains parameters for listing categories.
type ListCategoriesParams struct {
	Title    string
	Page     int
	PageSize int
}

// ListCategories retrieves all marketplace categories from Waldur.
func (m *MarketplaceClient) ListCategories(ctx context.Context, params ListCategoriesParams) ([]Category, error) {
	var categories []Category

	err := m.client.doWithRetry(ctx, func() error {
		query := url.Values{}
		if params.Title != "" {
			query.Set("title", params.Title)
		}
		if params.Page > 0 {
			query.Set("page", fmt.Sprintf("%d", params.Page))
		}
		if params.PageSize > 0 {
			query.Set("page_size", fmt.Sprintf("%d", params.PageSize))
		}
		return m.client.getPaginated(ctx, buildQueryPath("/marketplace-categories/", query), &categories)
	})

	return categories, err
}

// GetCategoryByTitle finds a category by its title.
func (m *MarketplaceClient) GetCategoryByTitle(ctx context.Context, title string) (*Category, error) {
	categories, err := m.ListCategories(ctx, ListCategoriesParams{Title: title})
	if err != nil {
		return nil, err
	}

	for _, cat := range categories {
		if cat.Title == title {
			return &cat, nil
		}
	}

	return nil, ErrNotFound
}

// CreateCategoryRequest contains parameters for creating a category.
type CreateCategoryRequest struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`
}

// CreateCategory creates a new marketplace category in Waldur.
func (m *MarketplaceClient) CreateCategory(ctx context.Context, req CreateCategoryRequest) (*Category, error) {
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	var category *Category

	err := m.client.doWithRetry(ctx, func() error {
		bodyBytes, err := json.Marshal(req)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}

		respBody, statusCode, err := m.client.doRequest(ctx, http.MethodPost, "/marketplace-categories/", bodyBytes)
		if err != nil {
			return err
		}

		if statusCode != http.StatusCreated && statusCode != http.StatusOK {
			return mapHTTPError(statusCode, respBody)
		}

		var respData Category
		if err := json.Unmarshal(respBody, &respData); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}

		category = &respData
		return nil
	})

	return category, err
}

// EnsureCategory creates a category if it doesn't exist, or returns the existing one.
func (m *MarketplaceClient) EnsureCategory(ctx context.Context, req CreateCategoryRequest) (*Category, error) {
	// First check if category exists
	existing, err := m.GetCategoryByTitle(ctx, req.Title)
	if err == nil && existing != nil {
		return existing, nil
	}
	if err != nil && err != ErrNotFound {
		return nil, fmt.Errorf("check existing category: %w", err)
	}

	// Create new category
	return m.CreateCategory(ctx, req)
}
