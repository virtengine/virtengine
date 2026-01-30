// Package waldur provides Marketplace operations via Waldur API
//
// VE-2024: Marketplace offerings, orders, and resources management
package waldur

import (
	verrors "github.com/virtengine/virtengine/pkg/errors"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	client "github.com/waldur/go-client"
)

// MarketplaceClient provides marketplace operations through Waldur
type MarketplaceClient struct {
	client *Client
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
	var offerings []Offering

	err := m.client.doWithRetry(ctx, func() error {
		apiParams := &client.MarketplacePublicOfferingsListParams{}

		if params.CustomerUUID != "" {
			u := uuid.MustParse(params.CustomerUUID)
			apiParams.CustomerUuid = &u
		}
		if params.CategoryUUID != "" {
			u := uuid.MustParse(params.CategoryUUID)
			apiParams.CategoryUuid = &u
		}
		if params.Name != "" {
			apiParams.Name = &params.Name
		}
		if params.Shared != nil {
			apiParams.Shared = params.Shared
		}
		if params.Page > 0 {
			page := client.Page(params.Page)
			apiParams.Page = &page
		}
		if params.PageSize > 0 {
			pageSize := client.PageSize(params.PageSize)
			apiParams.PageSize = &pageSize
		}

		resp, err := m.client.api.MarketplacePublicOfferingsListWithResponse(ctx, apiParams)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		if resp.JSON200 == nil {
			return ErrInvalidResponse
		}

		offerings = make([]Offering, 0, len(*resp.JSON200))
		for _, o := range *resp.JSON200 {
			offering := Offering{
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
			offerings = append(offerings, offering)
		}

		return nil
	})

	return offerings, err
}

// GetOffering retrieves a specific marketplace offering
func (m *MarketplaceClient) GetOffering(ctx context.Context, offeringUUID string) (*Offering, error) {
	var offering *Offering

	err := m.client.doWithRetry(ctx, func() error {
		uuidType := uuid.MustParse(offeringUUID)
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
	var orders []Order

	err := m.client.doWithRetry(ctx, func() error {
		apiParams := &client.MarketplaceOrdersListParams{}

		if params.ProjectUUID != "" {
			u := uuid.MustParse(params.ProjectUUID)
			apiParams.ProjectUuid = &u
		}
		if params.CustomerUUID != "" {
			u := uuid.MustParse(params.CustomerUUID)
			apiParams.CustomerUuid = &u
		}
		if params.OfferingUUID != "" {
			u := uuid.MustParse(params.OfferingUUID)
			apiParams.OfferingUuid = &u
		}
		if params.Page > 0 {
			page := client.Page(params.Page)
			apiParams.Page = &page
		}
		if params.PageSize > 0 {
			pageSize := client.PageSize(params.PageSize)
			apiParams.PageSize = &pageSize
		}

		resp, err := m.client.api.MarketplaceOrdersListWithResponse(ctx, apiParams)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		if resp.JSON200 == nil {
			return ErrInvalidResponse
		}

		orders = make([]Order, 0, len(*resp.JSON200))
		for _, o := range *resp.JSON200 {
			order := Order{
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
			orders = append(orders, order)
		}

		return nil
	})

	return orders, err
}

// GetOrder retrieves a specific marketplace order
func (m *MarketplaceClient) GetOrder(ctx context.Context, orderUUID string) (*Order, error) {
	var order *Order

	err := m.client.doWithRetry(ctx, func() error {
		uuidType := uuid.MustParse(orderUUID)
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
		uuidType := uuid.MustParse(orderUUID)
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
		uuidType := uuid.MustParse(orderUUID)
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
		uuidType := uuid.MustParse(resourceUUID)
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
	var resources []Resource

	err := m.client.doWithRetry(ctx, func() error {
		apiParams := &client.MarketplaceResourcesListParams{}

		if params.ProjectUUID != "" {
			u := uuid.MustParse(params.ProjectUUID)
			apiParams.ProjectUuid = &u
		}
		if params.CustomerUUID != "" {
			u := uuid.MustParse(params.CustomerUUID)
			apiParams.CustomerUuid = &u
		}
		if params.OfferingUUID != "" {
			u := uuid.MustParse(params.OfferingUUID)
			apiParams.OfferingUuid = &[]uuid.UUID{u}
		}
		if params.Page > 0 {
			page := client.Page(params.Page)
			apiParams.Page = &page
		}
		if params.PageSize > 0 {
			pageSize := client.PageSize(params.PageSize)
			apiParams.PageSize = &pageSize
		}

		resp, err := m.client.api.MarketplaceResourcesListWithResponse(ctx, apiParams)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			return mapHTTPError(resp.StatusCode(), resp.Body)
		}

		if resp.JSON200 == nil {
			return ErrInvalidResponse
		}

		resources = make([]Resource, 0, len(*resp.JSON200))
		for _, r := range *resp.JSON200 {
			resource := Resource{
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
			resources = append(resources, resource)
		}

		return nil
	})

	return resources, err
}

// GetResource retrieves a specific marketplace resource
func (m *MarketplaceClient) GetResource(ctx context.Context, resourceUUID string) (*Resource, error) {
	var resource *Resource

	err := m.client.doWithRetry(ctx, func() error {
		uuidType := uuid.MustParse(resourceUUID)
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
			case "done":
				return order, nil
			case "erred":
				return order, fmt.Errorf("order failed: %s", order.ErrorMessage)
			case "canceled", "rejected":
				return order, fmt.Errorf("order %s", order.State)
			}
		}
	}
}
