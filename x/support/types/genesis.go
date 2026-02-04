package types

import (
	"fmt"
	"time"
)

// GenesisState is the genesis state for the support module
type GenesisState struct {
	// ExternalRefs are the initial external ticket references
	ExternalRefs []ExternalTicketRef `json:"external_refs"`

	// SupportRequests are the initial support requests
	SupportRequests []SupportRequest `json:"support_requests"`

	// SupportResponses are the initial support responses
	SupportResponses []SupportResponse `json:"support_responses"`

	// EventSequence is the support event sequence
	EventSequence uint64 `json:"event_sequence"`

	// Params are the module parameters
	Params Params `json:"params"`
}

// Params defines the parameters for the support module
type Params struct {
	// AllowedExternalSystems is the list of allowed external systems
	AllowedExternalSystems []string `json:"allowed_external_systems"`

	// AllowedExternalDomains is the list of allowed domains for external URLs
	// Used for validation to prevent phishing
	AllowedExternalDomains []string `json:"allowed_external_domains"`

	// SupportRecipientKeyIDs is the list of encryption key IDs for support agents
	SupportRecipientKeyIDs []string `json:"support_recipient_key_ids,omitempty"`

	// RequireSupportRecipients enforces support recipient keys inclusion
	RequireSupportRecipients bool `json:"require_support_recipients"`

	// MaxResponsesPerRequest limits responses per ticket
	MaxResponsesPerRequest uint32 `json:"max_responses_per_request"`

	// DefaultRetentionPolicy defines default retention policy for new tickets
	DefaultRetentionPolicy RetentionPolicy `json:"default_retention_policy"`
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		ExternalRefs:     []ExternalTicketRef{},
		SupportRequests:  []SupportRequest{},
		SupportResponses: []SupportResponse{},
		EventSequence:    0,
		Params:           DefaultParams(),
	}
}

// DefaultParams returns the default parameters
func DefaultParams() Params {
	return Params{
		AllowedExternalSystems: []string{
			string(ExternalSystemWaldur),
			string(ExternalSystemJira),
		},
		AllowedExternalDomains:   []string{}, // Empty = allow all (configure in production)
		SupportRecipientKeyIDs:   []string{},
		RequireSupportRecipients: true,
		MaxResponsesPerRequest:   200,
		DefaultRetentionPolicy: RetentionPolicy{
			Version:             RetentionPolicyVersion,
			ArchiveAfterSeconds: int64((90 * 24 * time.Hour).Seconds()),
			PurgeAfterSeconds:   int64((365 * 24 * time.Hour).Seconds()),
		},
	}
}

// Validate validates the genesis state
func (gs GenesisState) Validate() error {
	// Validate external refs
	seenRefs := make(map[string]bool)
	for _, ref := range gs.ExternalRefs {
		if err := ref.Validate(); err != nil {
			return err
		}
		key := ref.Key()
		if seenRefs[key] {
			return ErrRefAlreadyExists.Wrapf("duplicate external ref: %s", key)
		}
		seenRefs[key] = true
	}

	// Validate support requests
	seenRequests := make(map[string]bool)
	for _, req := range gs.SupportRequests {
		if err := req.Validate(); err != nil {
			return err
		}
		id := req.ID.String()
		if seenRequests[id] {
			return ErrInvalidSupportRequest.Wrapf("duplicate support request: %s", id)
		}
		seenRequests[id] = true
	}

	// Validate support responses
	seenResponses := make(map[string]bool)
	for _, resp := range gs.SupportResponses {
		if err := resp.Validate(); err != nil {
			return err
		}
		id := resp.ID.String()
		if seenResponses[id] {
			return ErrInvalidSupportResponse.Wrapf("duplicate support response: %s", id)
		}
		seenResponses[id] = true
	}

	// Validate params
	return gs.Params.Validate()
}

// Validate validates the params
func (p Params) Validate() error {
	if len(p.AllowedExternalSystems) == 0 {
		return ErrInvalidParams.Wrap("allowed_external_systems cannot be empty")
	}

	// Validate each system is known
	for _, sys := range p.AllowedExternalSystems {
		if !ExternalSystem(sys).IsValid() {
			return ErrInvalidParams.Wrapf("unknown external system: %s", sys)
		}
	}

	if p.MaxResponsesPerRequest == 0 {
		return ErrInvalidParams.Wrap("max_responses_per_request must be greater than 0")
	}

	if err := p.DefaultRetentionPolicy.Validate(); err != nil {
		return err
	}

	return nil
}

// IsSystemAllowed checks if an external system is allowed
func (p Params) IsSystemAllowed(system ExternalSystem) bool {
	for _, s := range p.AllowedExternalSystems {
		if s == string(system) {
			return true
		}
	}
	return false
}

// Proto message interface stubs for GenesisState
func (*GenesisState) ProtoMessage() {}
func (gs *GenesisState) Reset()     { *gs = GenesisState{} }
func (gs *GenesisState) String() string {
	return fmt.Sprintf("GenesisState{ExternalRefs: %d, SupportRequests: %d, SupportResponses: %d}",
		len(gs.ExternalRefs), len(gs.SupportRequests), len(gs.SupportResponses))
}

// Proto message interface stubs for Params
func (*Params) ProtoMessage() {}
func (p *Params) Reset()      { *p = Params{} }
func (p *Params) String() string {
	return fmt.Sprintf("Params{AllowedSystems: %v, MaxResponses: %d}", p.AllowedExternalSystems, p.MaxResponsesPerRequest)
}
