// Package provider_daemon provides category synchronization for Waldur integration.
//
// VE-25A: Auto-create and sync marketplace categories between VirtEngine and Waldur.
package provider_daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/virtengine/virtengine/pkg/waldur"
)

// VirtEngine default categories that are synced to Waldur.
// These are the only categories that will be synced from provider offerings to Waldur.
var DefaultCategories = []CategoryDefinition{
	{
		Title:       "Compute",
		Description: "Virtual machines, containers, and general-purpose computing resources for running workloads.",
		Icon:        "server",
		Priority:    1,
	},
	{
		Title:       "HPC",
		Description: "High-performance computing resources including MPI clusters, batch processing, and scientific computing environments.",
		Icon:        "cpu",
		Priority:    2,
	},
	{
		Title:       "GPU",
		Description: "GPU-accelerated computing instances for machine learning, deep learning, and graphics processing workloads.",
		Icon:        "gpu",
		Priority:    3,
	},
	{
		Title:       "Storage",
		Description: "Object storage, block storage, and file storage solutions for data persistence and sharing.",
		Icon:        "database",
		Priority:    4,
	},
	{
		Title:       "Network",
		Description: "Networking services including VPNs, load balancers, firewalls, and private networks.",
		Icon:        "network",
		Priority:    5,
	},
	{
		Title:       "TEE",
		Description: "Trusted Execution Environment resources for confidential computing and secure enclaves.",
		Icon:        "shield",
		Priority:    6,
	},
	{
		Title:       "AI/ML",
		Description: "Machine learning platforms, model training infrastructure, and inference services.",
		Icon:        "brain",
		Priority:    7,
	},
}

// CategoryDefinition defines a VirtEngine marketplace category.
type CategoryDefinition struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Icon        string `json:"icon,omitempty"`
	Priority    int    `json:"priority,omitempty"`
}

// CategoryMapping maps a VirtEngine category title to its Waldur UUID.
type CategoryMapping struct {
	Title       string    `json:"title"`
	WaldurUUID  string    `json:"waldur_uuid"`
	Description string    `json:"description,omitempty"`
	SyncedAt    time.Time `json:"synced_at"`
}

// CategorySyncState holds the state of category synchronization.
type CategorySyncState struct {
	Mappings    map[string]*CategoryMapping `json:"mappings"`
	LastSync    time.Time                   `json:"last_sync"`
	SyncVersion string                      `json:"sync_version"`
}

// CategorySyncConfig configures the category sync worker.
type CategorySyncConfig struct {
	// Enabled enables category sync.
	Enabled bool

	// WaldurBaseURL is the Waldur API base URL.
	WaldurBaseURL string

	// WaldurToken is the API authentication token.
	WaldurToken string

	// StateFilePath is the path to persist category sync state.
	StateFilePath string

	// SyncIntervalSeconds is the reconciliation interval.
	SyncIntervalSeconds int64

	// MaxRetries is the maximum retry attempts.
	MaxRetries int

	// OperationTimeout is the timeout for Waldur operations.
	OperationTimeout time.Duration

	// CategoriesFilePath is an optional path to custom categories JSON file.
	CategoriesFilePath string
}

// DefaultCategorySyncConfig returns sensible defaults.
func DefaultCategorySyncConfig() CategorySyncConfig {
	return CategorySyncConfig{
		StateFilePath:       "data/category_sync_state.json",
		SyncIntervalSeconds: 300,
		MaxRetries:          3,
		OperationTimeout:    30 * time.Second,
	}
}

// CategorySyncWorker handles category synchronization with Waldur.
type CategorySyncWorker struct {
	mu     sync.RWMutex
	cfg    CategorySyncConfig
	state  *CategorySyncState
	client *waldur.MarketplaceClient

	stopCh chan struct{}
	doneCh chan struct{}
}

// NewCategorySyncWorker creates a new category sync worker.
func NewCategorySyncWorker(cfg CategorySyncConfig, client *waldur.MarketplaceClient) (*CategorySyncWorker, error) {
	if client == nil {
		return nil, fmt.Errorf("marketplace client is required")
	}

	// Load existing state if available
	state := &CategorySyncState{
		Mappings:    make(map[string]*CategoryMapping),
		SyncVersion: "1.0.0",
	}

	if cfg.StateFilePath != "" {
		if data, err := os.ReadFile(cfg.StateFilePath); err == nil {
			if err := json.Unmarshal(data, state); err != nil {
				log.Printf("[category-sync] warning: failed to parse state file: %v", err)
			}
		}
	}

	return &CategorySyncWorker{
		cfg:    cfg,
		state:  state,
		client: client,
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}, nil
}

// Start begins the category sync worker.
func (w *CategorySyncWorker) Start(ctx context.Context) error {
	if !w.cfg.Enabled {
		return nil
	}

	// Perform initial sync
	if err := w.SyncCategories(ctx); err != nil {
		log.Printf("[category-sync] initial sync failed (will retry): %v", err)
	}

	// Start background reconciliation loop
	go w.reconcileLoop(ctx)

	log.Printf("[category-sync] worker started")
	return nil
}

// Stop stops the category sync worker.
func (w *CategorySyncWorker) Stop() error {
	close(w.stopCh)
	<-w.doneCh
	return nil
}

// State returns the current sync state.
func (w *CategorySyncWorker) State() *CategorySyncState {
	w.mu.RLock()
	defer w.mu.RUnlock()

	// Return a copy to prevent race conditions
	stateCopy := &CategorySyncState{
		Mappings:    make(map[string]*CategoryMapping),
		LastSync:    w.state.LastSync,
		SyncVersion: w.state.SyncVersion,
	}
	for k, v := range w.state.Mappings {
		stateCopy.Mappings[k] = &CategoryMapping{
			Title:       v.Title,
			WaldurUUID:  v.WaldurUUID,
			Description: v.Description,
			SyncedAt:    v.SyncedAt,
		}
	}
	return stateCopy
}

// GetCategoryUUID returns the Waldur UUID for a category title.
func (w *CategorySyncWorker) GetCategoryUUID(title string) (string, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	mapping, ok := w.state.Mappings[title]
	if !ok || mapping == nil {
		return "", false
	}
	return mapping.WaldurUUID, true
}

// GetAllMappings returns all category mappings.
func (w *CategorySyncWorker) GetAllMappings() map[string]string {
	w.mu.RLock()
	defer w.mu.RUnlock()

	result := make(map[string]string)
	for title, mapping := range w.state.Mappings {
		if mapping != nil {
			result[title] = mapping.WaldurUUID
		}
	}
	return result
}

// SyncCategories synchronizes all default categories to Waldur.
func (w *CategorySyncWorker) SyncCategories(ctx context.Context) error {
	categories := w.getCategoriesToSync()

	log.Printf("[category-sync] syncing %d categories to Waldur", len(categories))

	var syncErrors []error
	for _, cat := range categories {
		if err := w.syncCategory(ctx, cat); err != nil {
			syncErrors = append(syncErrors, fmt.Errorf("sync category %s: %w", cat.Title, err))
			log.Printf("[category-sync] failed to sync category %s: %v", cat.Title, err)
		} else {
			log.Printf("[category-sync] synced category: %s", cat.Title)
		}
	}

	w.mu.Lock()
	w.state.LastSync = time.Now().UTC()
	w.mu.Unlock()

	if err := w.saveState(); err != nil {
		log.Printf("[category-sync] warning: failed to save state: %v", err)
	}

	if len(syncErrors) > 0 {
		return fmt.Errorf("failed to sync %d categories", len(syncErrors))
	}

	return nil
}

// DiscoverCategories discovers existing categories from Waldur and updates mappings.
func (w *CategorySyncWorker) DiscoverCategories(ctx context.Context) error {
	opCtx, cancel := context.WithTimeout(ctx, w.cfg.OperationTimeout)
	defer cancel()

	categories, err := w.client.ListCategories(opCtx, waldur.ListCategoriesParams{PageSize: 100})
	if err != nil {
		return fmt.Errorf("list categories: %w", err)
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	// Update mappings with discovered categories
	for _, cat := range categories {
		if existing, ok := w.state.Mappings[cat.Title]; ok {
			// Update existing mapping if UUID changed
			if existing.WaldurUUID != cat.UUID {
				existing.WaldurUUID = cat.UUID
				existing.SyncedAt = time.Now().UTC()
			}
		} else {
			// Add new mapping for discovered category
			w.state.Mappings[cat.Title] = &CategoryMapping{
				Title:       cat.Title,
				WaldurUUID:  cat.UUID,
				Description: cat.Description,
				SyncedAt:    time.Now().UTC(),
			}
		}
	}

	return nil
}

func (w *CategorySyncWorker) getCategoriesToSync() []CategoryDefinition {
	// If custom categories file is specified, load from there
	if w.cfg.CategoriesFilePath != "" {
		categories, err := LoadCategoriesFromFile(w.cfg.CategoriesFilePath)
		if err != nil {
			log.Printf("[category-sync] warning: failed to load custom categories, using defaults: %v", err)
			return DefaultCategories
		}
		return categories
	}

	return DefaultCategories
}

func (w *CategorySyncWorker) syncCategory(ctx context.Context, cat CategoryDefinition) error {
	opCtx, cancel := context.WithTimeout(ctx, w.cfg.OperationTimeout)
	defer cancel()

	result, err := w.client.EnsureCategory(opCtx, waldur.CreateCategoryRequest{
		Title:       cat.Title,
		Description: cat.Description,
		Icon:        cat.Icon,
	})
	if err != nil {
		return err
	}

	w.mu.Lock()
	w.state.Mappings[cat.Title] = &CategoryMapping{
		Title:       cat.Title,
		WaldurUUID:  result.UUID,
		Description: cat.Description,
		SyncedAt:    time.Now().UTC(),
	}
	w.mu.Unlock()

	return nil
}

func (w *CategorySyncWorker) reconcileLoop(ctx context.Context) {
	defer close(w.doneCh)

	interval := time.Duration(w.cfg.SyncIntervalSeconds) * time.Second
	if interval < time.Minute {
		interval = 5 * time.Minute
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		case <-ticker.C:
			if err := w.SyncCategories(ctx); err != nil {
				log.Printf("[category-sync] reconciliation failed: %v", err)
			}
		}
	}
}

func (w *CategorySyncWorker) saveState() error {
	if w.cfg.StateFilePath == "" {
		return nil
	}

	w.mu.RLock()
	data, err := json.MarshalIndent(w.state, "", "  ")
	w.mu.RUnlock()

	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	return os.WriteFile(w.cfg.StateFilePath, data, 0o600)
}

// LoadCategoriesFromFile loads category definitions from a JSON file.
func LoadCategoriesFromFile(path string) ([]CategoryDefinition, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var fixture struct {
		Categories []CategoryDefinition `json:"categories"`
	}

	if err := json.Unmarshal(data, &fixture); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return fixture.Categories, nil
}

// SyncDefaultCategories is a convenience function to sync default categories.
// This can be called from the localnet script or CLI.
func SyncDefaultCategories(ctx context.Context, waldurBaseURL, waldurToken string) error {
	cfg := waldur.DefaultConfig()
	cfg.BaseURL = waldurBaseURL
	cfg.Token = waldurToken

	client, err := waldur.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("create waldur client: %w", err)
	}

	marketplace := waldur.NewMarketplaceClient(client)

	workerCfg := DefaultCategorySyncConfig()
	workerCfg.Enabled = true

	worker, err := NewCategorySyncWorker(workerCfg, marketplace)
	if err != nil {
		return fmt.Errorf("create sync worker: %w", err)
	}

	return worker.SyncCategories(ctx)
}

// InitCategoriesResult contains the result of category initialization.
type InitCategoriesResult struct {
	Created  []string          `json:"created"`
	Existing []string          `json:"existing"`
	Failed   []string          `json:"failed"`
	Mappings map[string]string `json:"mappings"`
}

// InitCategories initializes all default categories and returns the result.
func InitCategories(ctx context.Context, waldurBaseURL, waldurToken string) (*InitCategoriesResult, error) {
	cfg := waldur.DefaultConfig()
	cfg.BaseURL = waldurBaseURL
	cfg.Token = waldurToken

	client, err := waldur.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("create waldur client: %w", err)
	}

	marketplace := waldur.NewMarketplaceClient(client)

	result := &InitCategoriesResult{
		Created:  []string{},
		Existing: []string{},
		Failed:   []string{},
		Mappings: make(map[string]string),
	}

	for _, cat := range DefaultCategories {
		opCtx, cancel := context.WithTimeout(ctx, 30*time.Second)

		// Try to get existing category first
		existing, err := marketplace.GetCategoryByTitle(opCtx, cat.Title)
		if err == nil && existing != nil {
			result.Existing = append(result.Existing, cat.Title)
			result.Mappings[cat.Title] = existing.UUID
			cancel()
			continue
		}

		// Create new category
		created, err := marketplace.CreateCategory(opCtx, waldur.CreateCategoryRequest{
			Title:       cat.Title,
			Description: cat.Description,
			Icon:        cat.Icon,
		})
		cancel()

		if err != nil {
			result.Failed = append(result.Failed, cat.Title)
			log.Printf("[init-categories] failed to create category %s: %v", cat.Title, err)
			continue
		}

		result.Created = append(result.Created, cat.Title)
		result.Mappings[cat.Title] = created.UUID
	}

	return result, nil
}
