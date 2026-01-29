// Package waldur provides Marketplace operations via Waldur API
//
// VE-2024: Marketplace offerings, orders, and resources management
package waldur

import (
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
	UUID        string    `json:"uuid"`
	State       string    `json:"state"`
	Type        string    `json:"type"`
	ProjectUUID string    `json:"project_uuid"`
	CreatedAt   time.Time `json:"created"`
	ErrorMessage string   `json:"error_message"`
}

// Resource represents a marketplace resource in Waldur
type Resource struct {
	UUID          string    `json:"uuid"`
	Name          string    `json:"name"`
	State         string    `json:"state"`
	OfferingUUID  string    `json:"offering_uuid"`
	ProjectUUID   string    `json:"project_uuid"`
	ResourceType  string    `json:"resource_type"`
	CreatedAt     time.Time `json:"created"`
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
