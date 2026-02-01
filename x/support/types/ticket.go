// Package types provides the types for the support module.
//
// The support module stores lightweight references to external support tickets
// managed by operator software (Waldur/Jira). It does NOT manage the ticket
// lifecycle on-chain - that is handled by the operator's service desk system.
package types

import (
	"fmt"
	"net/url"
	"time"
)

// ExternalSystem identifies the external service desk system
type ExternalSystem string

const (
	// ExternalSystemWaldur represents Waldur service desk (recommended)
	ExternalSystemWaldur ExternalSystem = "waldur"

	// ExternalSystemJira represents Jira Service Desk (via Waldur integration)
	ExternalSystemJira ExternalSystem = "jira"
)

// IsValid checks if the external system is valid
func (s ExternalSystem) IsValid() bool {
	return s == ExternalSystemWaldur || s == ExternalSystemJira
}

// String returns the string representation
func (s ExternalSystem) String() string {
	return string(s)
}

// ResourceType identifies the type of on-chain resource
type ResourceType string

const (
	ResourceTypeDeployment ResourceType = "deployment"
	ResourceTypeLease      ResourceType = "lease"
	ResourceTypeOrder      ResourceType = "order"
	ResourceTypeProvider   ResourceType = "provider"
)

// IsValid checks if the resource type is valid
func (t ResourceType) IsValid() bool {
	switch t {
	case ResourceTypeDeployment, ResourceTypeLease, ResourceTypeOrder, ResourceTypeProvider:
		return true
	default:
		return false
	}
}

// ExternalTicketRef stores a reference to an external support ticket.
// This is the primary data structure - lightweight reference for traceability.
type ExternalTicketRef struct {
	// ResourceID is the on-chain resource ID (e.g., "owner/dseq/gseq/oseq")
	ResourceID string `json:"resource_id"`

	// ResourceType is the type of resource ("deployment", "lease", "order", "provider")
	ResourceType ResourceType `json:"resource_type"`

	// ExternalSystem is the external service desk system ("waldur" or "jira")
	ExternalSystem ExternalSystem `json:"external_system"`

	// ExternalTicketID is the external ticket ID (e.g., "JIRA-123", waldur UUID)
	ExternalTicketID string `json:"external_ticket_id"`

	// ExternalURL is the URL to the external ticket
	ExternalURL string `json:"external_url"`

	// CreatedAt is when the reference was created
	CreatedAt time.Time `json:"created_at"`

	// CreatedBy is the address that created the reference
	CreatedBy string `json:"created_by"`

	// UpdatedAt is when the reference was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate validates the external ticket reference
func (r *ExternalTicketRef) Validate() error {
	if r.ResourceID == "" {
		return ErrInvalidResourceRef.Wrap("resource_id is required")
	}

	if !r.ResourceType.IsValid() {
		return ErrInvalidResourceRef.Wrapf("invalid resource_type: %s", r.ResourceType)
	}

	if !r.ExternalSystem.IsValid() {
		return ErrInvalidExternalSystem.Wrapf("invalid external_system: %s", r.ExternalSystem)
	}

	if r.ExternalTicketID == "" {
		return ErrInvalidExternalTicketID.Wrap("external_ticket_id is required")
	}

	if r.ExternalURL != "" {
		if _, err := url.Parse(r.ExternalURL); err != nil {
			return ErrInvalidExternalURL.Wrapf("invalid external_url: %v", err)
		}
	}

	if r.CreatedBy == "" {
		return ErrInvalidAddress.Wrap("created_by is required")
	}

	return nil
}

// Key returns the unique key for this reference
func (r *ExternalTicketRef) Key() string {
	return fmt.Sprintf("%s/%s", r.ResourceType, r.ResourceID)
}

// Proto message interface stubs
func (*ExternalTicketRef) ProtoMessage()    {}
func (r *ExternalTicketRef) Reset()         { *r = ExternalTicketRef{} }
func (r *ExternalTicketRef) String() string { return fmt.Sprintf("%+v", *r) }
