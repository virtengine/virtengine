// Package provider_daemon implements provider-side services for VirtEngine.
//
// VE-25H: Offering publication flow for Waldur-to-chain synchronization with portal management.
// This file implements the publication coordinator that bridges Waldur offerings to the chain
// and provides API endpoints for the portal to manage offerings.
package provider_daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/gorilla/mux"

	"github.com/virtengine/virtengine/pkg/waldur"
	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

// OfferingPublicationConfig configures the offering publication service.
type OfferingPublicationConfig struct {
	// Enabled toggles the publication service.
	Enabled bool

	// ProviderAddress is the provider's on-chain address.
	ProviderAddress string

	// WaldurCustomerUUID is the Waldur customer UUID for offerings.
	WaldurCustomerUUID string

	// WaldurCategoryMap maps VirtEngine categories to Waldur category UUIDs.
	WaldurCategoryMap map[string]string

	// AutoPublish enables automatic publication of new Waldur offerings.
	AutoPublish bool

	// PollIntervalSeconds is how often to poll Waldur for new offerings.
	PollIntervalSeconds int64

	// StateFilePath is the path to persist publication state.
	StateFilePath string

	// OperationTimeout is the timeout for individual operations.
	OperationTimeout time.Duration
}

// DefaultOfferingPublicationConfig returns sensible defaults.
func DefaultOfferingPublicationConfig() OfferingPublicationConfig {
	return OfferingPublicationConfig{
		AutoPublish:         true,
		PollIntervalSeconds: 300,
		StateFilePath:       "data/offering_publication_state.json",
		OperationTimeout:    60 * time.Second,
		WaldurCategoryMap:   make(map[string]string),
	}
}

// OfferingPublicationStatus represents the publication status of an offering.
type OfferingPublicationStatus string

const (
	// PublicationStatusPending indicates the offering is pending publication.
	PublicationStatusPending OfferingPublicationStatus = "pending"

	// PublicationStatusPublished indicates the offering is published on-chain.
	PublicationStatusPublished OfferingPublicationStatus = "published"

	// PublicationStatusFailed indicates publication failed.
	PublicationStatusFailed OfferingPublicationStatus = "failed"

	// PublicationStatusPaused indicates the offering is paused.
	PublicationStatusPaused OfferingPublicationStatus = "paused"

	// PublicationStatusDeprecated indicates the offering is deprecated.
	PublicationStatusDeprecated OfferingPublicationStatus = "deprecated"

	// PublicationStatusDraft indicates the offering is a draft.
	PublicationStatusDraft OfferingPublicationStatus = "draft"
)

// Chain state constants.
const (
	chainStateActive = "active"
	chainStatePaused = "paused"
)

// OfferingPublication represents a managed offering with publication state.
type OfferingPublication struct {
	// WaldurUUID is the Waldur offering UUID.
	WaldurUUID string `json:"waldur_uuid"`

	// ChainOfferingID is the on-chain offering ID (when published).
	ChainOfferingID string `json:"chain_offering_id,omitempty"`

	// Name is the offering name.
	Name string `json:"name"`

	// Description is the offering description.
	Description string `json:"description,omitempty"`

	// Category is the offering category.
	Category string `json:"category"`

	// Status is the publication status.
	Status OfferingPublicationStatus `json:"status"`

	// WaldurState is the current Waldur state.
	WaldurState string `json:"waldur_state"`

	// ChainState is the current chain state (when published).
	ChainState string `json:"chain_state,omitempty"`

	// Pricing contains pricing information.
	Pricing []OfferingPriceComponent `json:"pricing,omitempty"`

	// SyncChecksum is the checksum for drift detection.
	SyncChecksum string `json:"sync_checksum,omitempty"`

	// LastSyncedAt is when the offering was last synced.
	LastSyncedAt *time.Time `json:"last_synced_at,omitempty"`

	// PublishedAt is when the offering was published.
	PublishedAt *time.Time `json:"published_at,omitempty"`

	// LastError is the last error message (if failed).
	LastError string `json:"last_error,omitempty"`

	// CreatedAt is when the publication record was created.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the publication record was last updated.
	UpdatedAt time.Time `json:"updated_at"`

	// ActiveLeases is the number of active leases for this offering.
	ActiveLeases int `json:"active_leases"`

	// TotalOrders is the total number of orders for this offering.
	TotalOrders int `json:"total_orders"`
}

// OfferingPriceComponent represents a pricing component for display.
type OfferingPriceComponent struct {
	ResourceType string `json:"resource_type"`
	Unit         string `json:"unit"`
	Price        string `json:"price"`
	Currency     string `json:"currency"`
}

// OfferingPublicationState holds the persistent state of publications.
type OfferingPublicationState struct {
	mu sync.RWMutex

	// ProviderAddress is the provider's address.
	ProviderAddress string `json:"provider_address"`

	// Publications maps Waldur UUID to publication record.
	Publications map[string]*OfferingPublication `json:"publications"`

	// LastPollAt is when Waldur was last polled.
	LastPollAt time.Time `json:"last_poll_at"`

	// Version is the state file version.
	Version string `json:"version"`
}

// NewOfferingPublicationState creates a new state instance.
func NewOfferingPublicationState(providerAddress string) *OfferingPublicationState {
	return &OfferingPublicationState{
		ProviderAddress: providerAddress,
		Publications:    make(map[string]*OfferingPublication),
		Version:         "1.0.0",
	}
}

// GetPublication returns a publication by Waldur UUID.
func (s *OfferingPublicationState) GetPublication(waldurUUID string) *OfferingPublication {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Publications[waldurUUID]
}

// SetPublication updates or creates a publication.
func (s *OfferingPublicationState) SetPublication(pub *OfferingPublication) {
	s.mu.Lock()
	defer s.mu.Unlock()
	pub.UpdatedAt = time.Now().UTC()
	s.Publications[pub.WaldurUUID] = pub
}

// ListPublications returns all publications sorted by name.
func (s *OfferingPublicationState) ListPublications() []*OfferingPublication {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pubs := make([]*OfferingPublication, 0, len(s.Publications))
	for _, pub := range s.Publications {
		pubs = append(pubs, pub)
	}

	sort.Slice(pubs, func(i, j int) bool {
		return pubs[i].Name < pubs[j].Name
	})

	return pubs
}

// FilterByStatus returns publications matching the given status.
func (s *OfferingPublicationState) FilterByStatus(status OfferingPublicationStatus) []*OfferingPublication {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*OfferingPublication
	for _, pub := range s.Publications {
		if pub.Status == status {
			result = append(result, pub)
		}
	}
	return result
}

// OfferingPublicationService manages offering publication lifecycle.
type OfferingPublicationService struct {
	cfg         OfferingPublicationConfig
	marketplace *waldur.MarketplaceClient
	submitter   OfferingSubmitter
	ingestCfg   marketplace.IngestConfig
	state       *OfferingPublicationState

	mu      sync.RWMutex
	running bool
	stopCh  chan struct{}
	doneCh  chan struct{}
}

// NewOfferingPublicationService creates a new publication service.
func NewOfferingPublicationService(
	cfg OfferingPublicationConfig,
	marketplaceClient *waldur.MarketplaceClient,
	submitter OfferingSubmitter,
) (*OfferingPublicationService, error) {
	if marketplaceClient == nil {
		return nil, fmt.Errorf("marketplace client is required")
	}

	state := NewOfferingPublicationState(cfg.ProviderAddress)

	return &OfferingPublicationService{
		cfg:         cfg,
		marketplace: marketplaceClient,
		submitter:   submitter,
		ingestCfg:   marketplace.DefaultIngestConfig(),
		state:       state,
		stopCh:      make(chan struct{}),
		doneCh:      make(chan struct{}),
	}, nil
}

// Start starts the publication service.
func (s *OfferingPublicationService) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("service already running")
	}
	s.running = true
	s.mu.Unlock()

	log.Printf("[offering-publication] started for provider %s", s.cfg.ProviderAddress)

	// Start polling loop if auto-publish is enabled
	if s.cfg.AutoPublish && s.cfg.PollIntervalSeconds > 0 {
		go s.pollLoop(ctx)
	}

	return nil
}

// Stop stops the publication service.
func (s *OfferingPublicationService) Stop() error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = false
	s.mu.Unlock()

	close(s.stopCh)

	select {
	case <-s.doneCh:
	case <-time.After(10 * time.Second):
		log.Printf("[offering-publication] shutdown timeout")
	}

	log.Printf("[offering-publication] stopped")
	return nil
}

// pollLoop periodically polls Waldur for new offerings.
func (s *OfferingPublicationService) pollLoop(ctx context.Context) {
	defer close(s.doneCh)

	ticker := time.NewTicker(time.Duration(s.cfg.PollIntervalSeconds) * time.Second)
	defer ticker.Stop()

	// Initial poll
	if err := s.PollWaldurOfferings(ctx); err != nil {
		log.Printf("[offering-publication] initial poll failed: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			if err := s.PollWaldurOfferings(ctx); err != nil {
				log.Printf("[offering-publication] poll failed: %v", err)
			}
		}
	}
}

// PollWaldurOfferings fetches offerings from Waldur and updates state.
func (s *OfferingPublicationService) PollWaldurOfferings(ctx context.Context) error {
	log.Printf("[offering-publication] polling Waldur for offerings")

	params := waldur.ListOfferingsParams{
		CustomerUUID: s.cfg.WaldurCustomerUUID,
		PageSize:     100,
	}

	offerings, err := s.marketplace.ListOfferings(ctx, params)
	if err != nil {
		return fmt.Errorf("list offerings: %w", err)
	}

	for _, o := range offerings {
		pub := s.state.GetPublication(o.UUID)
		if pub == nil {
			// New offering discovered
			pub = &OfferingPublication{
				WaldurUUID:  o.UUID,
				Name:        o.Name,
				Description: o.Description,
				Category:    mapWaldurTypeToCategory(o.Type),
				Status:      PublicationStatusPending,
				WaldurState: o.State,
				CreatedAt:   time.Now().UTC(),
			}
			s.state.SetPublication(pub)
			log.Printf("[offering-publication] discovered new offering: %s (%s)", o.Name, o.UUID)
		} else {
			// Update existing offering
			pub.Name = o.Name
			pub.Description = o.Description
			pub.WaldurState = o.State
			s.state.SetPublication(pub)
		}
	}

	s.state.LastPollAt = time.Now().UTC()
	log.Printf("[offering-publication] poll complete: %d offerings found", len(offerings))

	return nil
}

// mapWaldurTypeToCategory maps Waldur offering type to VirtEngine category.
func mapWaldurTypeToCategory(waldurType string) string {
	typeMap := map[string]string{
		"VirtEngine.Compute": "compute",
		"VirtEngine.Storage": "storage",
		"VirtEngine.Network": "network",
		"VirtEngine.HPC":     "hpc",
		"VirtEngine.GPU":     "gpu",
		"VirtEngine.ML":      "ml",
	}
	if cat, ok := typeMap[waldurType]; ok {
		return cat
	}
	return "other"
}

// PublishOffering publishes an offering to the chain.
func (s *OfferingPublicationService) PublishOffering(ctx context.Context, waldurUUID string) error {
	pub := s.state.GetPublication(waldurUUID)
	if pub == nil {
		return fmt.Errorf("offering not found: %s", waldurUUID)
	}

	if pub.Status == PublicationStatusPublished {
		return fmt.Errorf("offering already published")
	}

	// Fetch full offering details from Waldur
	offering, err := s.marketplace.GetOffering(ctx, waldurUUID)
	if err != nil {
		pub.Status = PublicationStatusFailed
		pub.LastError = fmt.Sprintf("fetch offering: %v", err)
		s.state.SetPublication(pub)
		return err
	}

	// Convert to chain offering
	importOffering := &marketplace.WaldurOfferingImport{
		UUID:        offering.UUID,
		Name:        offering.Name,
		Description: offering.Description,
		Type:        offering.Type,
		State:       offering.State,
		Shared:      offering.Shared,
		Billable:    offering.Billable,
		Created:     offering.CreatedAt,
		Modified:    offering.CreatedAt,
	}

	// Get next sequence
	sequence, err := s.submitter.GetNextOfferingSequence(ctx, s.cfg.ProviderAddress)
	if err != nil {
		pub.Status = PublicationStatusFailed
		pub.LastError = fmt.Sprintf("get sequence: %v", err)
		s.state.SetPublication(pub)
		return err
	}

	// Create chain offering
	chainOffering := importOffering.ToOffering(s.cfg.ProviderAddress, sequence, s.ingestCfg)
	chainOfferingID, err := s.submitter.CreateOffering(ctx, chainOffering)
	if err != nil {
		pub.Status = PublicationStatusFailed
		pub.LastError = fmt.Sprintf("create offering: %v", err)
		s.state.SetPublication(pub)
		return err
	}

	// Update publication state
	now := time.Now().UTC()
	pub.ChainOfferingID = chainOfferingID
	pub.Status = PublicationStatusPublished
	pub.ChainState = chainStateActive
	pub.PublishedAt = &now
	pub.LastSyncedAt = &now
	pub.LastError = ""
	s.state.SetPublication(pub)

	log.Printf("[offering-publication] published offering %s → %s", waldurUUID, chainOfferingID)
	return nil
}

// PauseOffering pauses an offering on the chain.
func (s *OfferingPublicationService) PauseOffering(ctx context.Context, waldurUUID string) error {
	pub := s.state.GetPublication(waldurUUID)
	if pub == nil {
		return fmt.Errorf("offering not found: %s", waldurUUID)
	}

	if pub.Status != PublicationStatusPublished {
		return fmt.Errorf("offering not published")
	}

	if pub.ChainOfferingID == "" {
		return fmt.Errorf("no chain offering ID")
	}

	// Update chain state - for now just update local state
	// In production, this would call the chain to pause the offering
	pub.Status = PublicationStatusPaused
	pub.ChainState = chainStatePaused
	s.state.SetPublication(pub)

	log.Printf("[offering-publication] paused offering %s", waldurUUID)
	return nil
}

// ActivateOffering activates a paused offering.
func (s *OfferingPublicationService) ActivateOffering(ctx context.Context, waldurUUID string) error {
	pub := s.state.GetPublication(waldurUUID)
	if pub == nil {
		return fmt.Errorf("offering not found: %s", waldurUUID)
	}

	if pub.Status != PublicationStatusPaused {
		return fmt.Errorf("offering not paused")
	}

	pub.Status = PublicationStatusPublished
	pub.ChainState = chainStateActive
	s.state.SetPublication(pub)

	log.Printf("[offering-publication] activated offering %s", waldurUUID)
	return nil
}

// DeprecateOffering deprecates an offering.
func (s *OfferingPublicationService) DeprecateOffering(ctx context.Context, waldurUUID string) error {
	pub := s.state.GetPublication(waldurUUID)
	if pub == nil {
		return fmt.Errorf("offering not found: %s", waldurUUID)
	}

	if pub.ChainOfferingID != "" {
		if err := s.submitter.DeprecateOffering(ctx, pub.ChainOfferingID); err != nil {
			pub.LastError = fmt.Sprintf("deprecate: %v", err)
			s.state.SetPublication(pub)
			return err
		}
	}

	pub.Status = PublicationStatusDeprecated
	pub.ChainState = "deprecated"
	s.state.SetPublication(pub)

	log.Printf("[offering-publication] deprecated offering %s", waldurUUID)
	return nil
}

// UpdatePricing updates the pricing for an offering.
func (s *OfferingPublicationService) UpdatePricing(ctx context.Context, waldurUUID string, pricing []OfferingPriceComponent) error {
	pub := s.state.GetPublication(waldurUUID)
	if pub == nil {
		return fmt.Errorf("offering not found: %s", waldurUUID)
	}

	// Update local pricing
	pub.Pricing = pricing
	s.state.SetPublication(pub)

	// If published, update chain offering
	if pub.Status == PublicationStatusPublished && pub.ChainOfferingID != "" {
		// Convert to chain pricing and update
		chainOffering := &marketplace.Offering{
			Prices: convertToChainPricing(pricing),
		}
		if err := s.submitter.UpdateOffering(ctx, pub.ChainOfferingID, chainOffering); err != nil {
			pub.LastError = fmt.Sprintf("update pricing: %v", err)
			s.state.SetPublication(pub)
			return err
		}
		now := time.Now().UTC()
		pub.LastSyncedAt = &now
	}

	log.Printf("[offering-publication] updated pricing for %s", waldurUUID)
	return nil
}

func convertToChainPricing(pricing []OfferingPriceComponent) []marketplace.PriceComponent {
	if len(pricing) == 0 {
		return nil
	}

	result := make([]marketplace.PriceComponent, 0, len(pricing))
	for _, p := range pricing {
		result = append(result, marketplace.PriceComponent{
			ResourceType: marketplace.PriceComponentResourceType(p.ResourceType),
			Unit:         p.Unit,
		})
	}
	return result
}

// ListOfferings returns all managed offerings.
func (s *OfferingPublicationService) ListOfferings() []*OfferingPublication {
	return s.state.ListPublications()
}

// GetOffering returns a specific offering.
func (s *OfferingPublicationService) GetOffering(waldurUUID string) *OfferingPublication {
	return s.state.GetPublication(waldurUUID)
}

// GetOfferingStats returns offering statistics.
func (s *OfferingPublicationService) GetOfferingStats() OfferingPublicationStats {
	pubs := s.state.ListPublications()

	stats := OfferingPublicationStats{}
	for _, pub := range pubs {
		stats.Total++
		switch pub.Status {
		case PublicationStatusPublished:
			stats.Published++
		case PublicationStatusPending:
			stats.Pending++
		case PublicationStatusPaused:
			stats.Paused++
		case PublicationStatusDeprecated:
			stats.Deprecated++
		case PublicationStatusFailed:
			stats.Failed++
		}
		stats.TotalLeases += pub.ActiveLeases
		stats.TotalOrders += pub.TotalOrders
	}

	return stats
}

// OfferingPublicationStats contains offering statistics.
type OfferingPublicationStats struct {
	Total       int `json:"total"`
	Published   int `json:"published"`
	Pending     int `json:"pending"`
	Paused      int `json:"paused"`
	Deprecated  int `json:"deprecated"`
	Failed      int `json:"failed"`
	TotalLeases int `json:"total_leases"`
	TotalOrders int `json:"total_orders"`
}

// ===============================================================================
// Portal API Handlers for Offering Management
// ===============================================================================

// RegisterOfferingRoutes registers offering management routes on the router.
func (s *OfferingPublicationService) RegisterOfferingRoutes(router *mux.Router) {
	router.HandleFunc("/offerings", s.handleListOfferings).Methods(http.MethodGet)
	router.HandleFunc("/offerings/stats", s.handleOfferingStats).Methods(http.MethodGet)
	router.HandleFunc("/offerings/{uuid}", s.handleGetOffering).Methods(http.MethodGet)
	router.HandleFunc("/offerings/{uuid}/publish", s.handlePublishOffering).Methods(http.MethodPost)
	router.HandleFunc("/offerings/{uuid}/pause", s.handlePauseOffering).Methods(http.MethodPost)
	router.HandleFunc("/offerings/{uuid}/activate", s.handleActivateOffering).Methods(http.MethodPost)
	router.HandleFunc("/offerings/{uuid}/deprecate", s.handleDeprecateOffering).Methods(http.MethodPost)
	router.HandleFunc("/offerings/{uuid}/pricing", s.handleUpdatePricing).Methods(http.MethodPut)
	router.HandleFunc("/offerings/sync", s.handleSyncOfferings).Methods(http.MethodPost)
}

func (s *OfferingPublicationService) handleListOfferings(w http.ResponseWriter, r *http.Request) {
	statusFilter := r.URL.Query().Get("status")

	var offerings []*OfferingPublication
	if statusFilter != "" {
		offerings = s.state.FilterByStatus(OfferingPublicationStatus(statusFilter))
	} else {
		offerings = s.ListOfferings()
	}

	w.Header().Set("Content-Type", "application/json")
	resp := struct {
		Offerings []*OfferingPublication `json:"offerings"`
		Total     int                    `json:"total"`
	}{
		Offerings: offerings,
		Total:     len(offerings),
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (s *OfferingPublicationService) handleOfferingStats(w http.ResponseWriter, _ *http.Request) {
	stats := s.GetOfferingStats()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (s *OfferingPublicationService) handleGetOffering(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]

	offering := s.GetOffering(uuid)
	if offering == nil {
		http.Error(w, "offering not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(offering); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (s *OfferingPublicationService) handlePublishOffering(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]

	if err := s.PublishOffering(r.Context(), uuid); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	offering := s.GetOffering(uuid)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(offering); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (s *OfferingPublicationService) handlePauseOffering(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]

	if err := s.PauseOffering(r.Context(), uuid); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	offering := s.GetOffering(uuid)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(offering); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (s *OfferingPublicationService) handleActivateOffering(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]

	if err := s.ActivateOffering(r.Context(), uuid); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	offering := s.GetOffering(uuid)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(offering); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (s *OfferingPublicationService) handleDeprecateOffering(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]

	if err := s.DeprecateOffering(r.Context(), uuid); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	offering := s.GetOffering(uuid)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(offering); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (s *OfferingPublicationService) handleUpdatePricing(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]

	var req struct {
		Pricing []OfferingPriceComponent `json:"pricing"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := s.UpdatePricing(r.Context(), uuid, req.Pricing); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	offering := s.GetOffering(uuid)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(offering); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (s *OfferingPublicationService) handleSyncOfferings(w http.ResponseWriter, r *http.Request) {
	if err := s.PollWaldurOfferings(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	offerings := s.ListOfferings()
	w.Header().Set("Content-Type", "application/json")
	resp := struct {
		Offerings []*OfferingPublication `json:"offerings"`
		Total     int                    `json:"total"`
		SyncedAt  time.Time              `json:"synced_at"`
	}{
		Offerings: offerings,
		Total:     len(offerings),
		SyncedAt:  s.state.LastPollAt,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
