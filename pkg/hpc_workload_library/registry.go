// Package hpc_workload_library provides workload registry functionality.
//
// VE-5F: Registry for workload template discovery and search
package hpc_workload_library

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/virtengine/virtengine/pkg/artifact_store"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// WorkloadRegistry provides discovery and search for workload templates
type WorkloadRegistry struct {
	mu sync.RWMutex

	// In-memory cache of templates
	templates map[string]*hpctypes.WorkloadTemplate

	// Index by type
	byType map[hpctypes.WorkloadType][]string

	// Index by tag
	byTag map[string][]string

	// Artifact store for persistence
	store artifact_store.ArtifactStore

	// Validator for templates
	validator *WorkloadValidator

	// Verifier for signatures
	verifier *TemplateVerifier

	// Config
	config RegistryConfig
}

// RegistryConfig configures the workload registry
type RegistryConfig struct {
	// RequireSignature requires templates to be signed
	RequireSignature bool

	// RequireApproval requires templates to be approved
	RequireApproval bool

	// CacheTTL is the cache TTL
	CacheTTL time.Duration

	// MaxTemplates is the maximum templates in cache
	MaxTemplates int

	// ValidationConfig is the validation configuration
	ValidationConfig ValidationConfig
}

// DefaultRegistryConfig returns the default registry configuration
func DefaultRegistryConfig() RegistryConfig {
	return RegistryConfig{
		RequireSignature: false,
		RequireApproval:  true,
		CacheTTL:         time.Hour,
		MaxTemplates:     1000,
		ValidationConfig: DefaultValidationConfig(),
	}
}

// NewWorkloadRegistry creates a new workload registry
func NewWorkloadRegistry(store artifact_store.ArtifactStore, config RegistryConfig) *WorkloadRegistry {
	r := &WorkloadRegistry{
		templates: make(map[string]*hpctypes.WorkloadTemplate),
		byType:    make(map[hpctypes.WorkloadType][]string),
		byTag:     make(map[string][]string),
		store:     store,
		validator: NewWorkloadValidator(config.ValidationConfig),
		verifier:  NewTemplateVerifier(),
		config:    config,
	}

	// Load built-in templates
	for _, t := range GetBuiltinTemplates() {
		r.addToCache(t)
	}

	return r
}

// Register registers a new workload template
func (r *WorkloadRegistry) Register(ctx context.Context, template *hpctypes.WorkloadTemplate) error {
	// Validate template
	result := r.validator.ValidateTemplate(ctx, template)
	if !result.IsValid() {
		return result.Error()
	}

	// Verify signature if required
	if r.config.RequireSignature {
		if err := r.verifier.Verify(template); err != nil {
			return fmt.Errorf("signature verification failed: %w", err)
		}
	}

	// Check approval status if required
	if r.config.RequireApproval && !template.ApprovalStatus.CanBeUsed() {
		return fmt.Errorf("template must be approved before registration")
	}

	// Store in artifact store if available
	if r.store != nil {
		data, err := json.Marshal(template)
		if err != nil {
			return fmt.Errorf("failed to serialize template: %w", err)
		}

		_, err = r.store.Put(ctx, &artifact_store.PutRequest{
			Data: data,
			EncryptionMetadata: &artifact_store.EncryptionMetadata{
				AlgorithmID: "none",
			},
			Owner:        template.Publisher,
			ArtifactType: "workload_template",
			Metadata: map[string]string{
				"template_id": template.TemplateID,
				"version":     template.Version,
				"type":        string(template.Type),
			},
		})
		if err != nil {
			return fmt.Errorf("failed to store template: %w", err)
		}
	}

	// Add to cache
	r.addToCache(template)

	return nil
}

// Get retrieves a template by ID
func (r *WorkloadRegistry) Get(ctx context.Context, templateID string) (*hpctypes.WorkloadTemplate, error) {
	r.mu.RLock()
	template, ok := r.templates[templateID]
	r.mu.RUnlock()

	if ok {
		return template, nil
	}

	return nil, fmt.Errorf("template not found: %s", templateID)
}

// GetVersion retrieves a specific version of a template
func (r *WorkloadRegistry) GetVersion(ctx context.Context, templateID, version string) (*hpctypes.WorkloadTemplate, error) {
	versionedID := templateID + "@" + version

	r.mu.RLock()
	template, ok := r.templates[versionedID]
	r.mu.RUnlock()

	if ok {
		return template, nil
	}

	// Fall back to latest if version matches
	r.mu.RLock()
	template, ok = r.templates[templateID]
	r.mu.RUnlock()

	if ok && template.Version == version {
		return template, nil
	}

	return nil, fmt.Errorf("template version not found: %s@%s", templateID, version)
}

// List lists all templates with optional filters
func (r *WorkloadRegistry) List(ctx context.Context, filter *TemplateFilter) ([]*hpctypes.WorkloadTemplate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*hpctypes.WorkloadTemplate

	// Get candidate IDs based on filter
	var candidateIDs []string
	if filter != nil && filter.Type != "" {
		candidateIDs = r.byType[filter.Type]
	} else if filter != nil && len(filter.Tags) > 0 {
		// Get intersection of all tag indices
		candidateIDs = r.getIDsByTags(filter.Tags)
	} else {
		// All templates
		for id := range r.templates {
			// Skip versioned IDs to avoid duplicates
			if !strings.Contains(id, "@") {
				candidateIDs = append(candidateIDs, id)
			}
		}
	}

	// Filter and collect results
	for _, id := range candidateIDs {
		template := r.templates[id]
		if template == nil {
			continue
		}

		if filter != nil && !filter.Matches(template) {
			continue
		}

		results = append(results, template)
	}

	// Sort by name
	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	// Apply pagination
	if filter != nil && filter.Limit > 0 {
		start := filter.Offset
		if start >= len(results) {
			return []*hpctypes.WorkloadTemplate{}, nil
		}
		end := start + filter.Limit
		if end > len(results) {
			end = len(results)
		}
		results = results[start:end]
	}

	return results, nil
}

// Search searches templates by query string
func (r *WorkloadRegistry) Search(ctx context.Context, query string, limit int) ([]*hpctypes.WorkloadTemplate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query = strings.ToLower(query)
	var results []*hpctypes.WorkloadTemplate

	for id, template := range r.templates {
		// Skip versioned IDs
		if strings.Contains(id, "@") {
			continue
		}

		// Search in name, description, tags
		if strings.Contains(strings.ToLower(template.Name), query) ||
			strings.Contains(strings.ToLower(template.Description), query) ||
			containsTag(template.Tags, query) {
			results = append(results, template)
		}
	}

	// Sort by relevance (exact matches first)
	sort.Slice(results, func(i, j int) bool {
		iExact := strings.EqualFold(results[i].Name, query)
		jExact := strings.EqualFold(results[j].Name, query)
		if iExact && !jExact {
			return true
		}
		if !iExact && jExact {
			return false
		}
		return results[i].Name < results[j].Name
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// ListByType lists templates by workload type
func (r *WorkloadRegistry) ListByType(ctx context.Context, workloadType hpctypes.WorkloadType) ([]*hpctypes.WorkloadTemplate, error) {
	return r.List(ctx, &TemplateFilter{Type: workloadType})
}

// ListByTag lists templates by tag
func (r *WorkloadRegistry) ListByTag(ctx context.Context, tag string) ([]*hpctypes.WorkloadTemplate, error) {
	return r.List(ctx, &TemplateFilter{Tags: []string{tag}})
}

// ListTags returns all unique tags
func (r *WorkloadRegistry) ListTags(ctx context.Context) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tags := make([]string, 0, len(r.byTag))
	for tag := range r.byTag {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags, nil
}

// ListTypes returns all template types with counts
func (r *WorkloadRegistry) ListTypes(ctx context.Context) (map[hpctypes.WorkloadType]int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	counts := make(map[hpctypes.WorkloadType]int)
	for t, ids := range r.byType {
		counts[t] = len(ids)
	}
	return counts, nil
}

// Count returns the total number of templates
func (r *WorkloadRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for id := range r.templates {
		if !strings.Contains(id, "@") {
			count++
		}
	}
	return count
}

// Unregister removes a template from the registry
func (r *WorkloadRegistry) Unregister(ctx context.Context, templateID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	template, ok := r.templates[templateID]
	if !ok {
		return fmt.Errorf("template not found: %s", templateID)
	}

	// Remove from indices
	r.removeFromIndices(template)

	// Remove from cache
	delete(r.templates, templateID)
	delete(r.templates, template.GetVersionedID())

	return nil
}

// Refresh refreshes the template cache from the artifact store
func (r *WorkloadRegistry) Refresh(ctx context.Context) error {
	if r.store == nil {
		return nil
	}

	// List all workload templates from store
	resp, err := r.store.ListByOwner(ctx, "", &artifact_store.Pagination{
		Limit: uint64(r.config.MaxTemplates),
	})
	if err != nil {
		return fmt.Errorf("failed to list templates: %w", err)
	}

	for _, ref := range resp.References {
		if ref.ArtifactType != "workload_template" {
			continue
		}

		// Get template data
		getResp, err := r.store.Get(ctx, &artifact_store.GetRequest{
			ContentAddress: ref.ContentAddress,
		})
		if err != nil {
			continue
		}

		var template hpctypes.WorkloadTemplate
		if err := json.Unmarshal(getResp.Data, &template); err != nil {
			continue
		}

		// Add to cache
		r.addToCache(&template)
	}

	return nil
}

// addToCache adds a template to the cache and indices
func (r *WorkloadRegistry) addToCache(template *hpctypes.WorkloadTemplate) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Store by ID
	r.templates[template.TemplateID] = template

	// Store by versioned ID
	r.templates[template.GetVersionedID()] = template

	// Update type index
	if _, ok := r.byType[template.Type]; !ok {
		r.byType[template.Type] = []string{}
	}
	if !contains(r.byType[template.Type], template.TemplateID) {
		r.byType[template.Type] = append(r.byType[template.Type], template.TemplateID)
	}

	// Update tag index
	for _, tag := range template.Tags {
		if _, ok := r.byTag[tag]; !ok {
			r.byTag[tag] = []string{}
		}
		if !contains(r.byTag[tag], template.TemplateID) {
			r.byTag[tag] = append(r.byTag[tag], template.TemplateID)
		}
	}
}

// removeFromIndices removes a template from indices
func (r *WorkloadRegistry) removeFromIndices(template *hpctypes.WorkloadTemplate) {
	// Remove from type index
	if ids, ok := r.byType[template.Type]; ok {
		r.byType[template.Type] = removeString(ids, template.TemplateID)
	}

	// Remove from tag index
	for _, tag := range template.Tags {
		if ids, ok := r.byTag[tag]; ok {
			r.byTag[tag] = removeString(ids, template.TemplateID)
		}
	}
}

// getIDsByTags returns template IDs that have all specified tags
func (r *WorkloadRegistry) getIDsByTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}

	// Start with first tag's IDs
	result := make(map[string]bool)
	for _, id := range r.byTag[tags[0]] {
		result[id] = true
	}

	// Intersect with remaining tags
	for _, tag := range tags[1:] {
		ids := r.byTag[tag]
		for id := range result {
			if !contains(ids, id) {
				delete(result, id)
			}
		}
	}

	out := make([]string, 0, len(result))
	for id := range result {
		out = append(out, id)
	}
	return out
}

// TemplateFilter defines filters for template listing
type TemplateFilter struct {
	// Type filters by workload type
	Type hpctypes.WorkloadType

	// Tags filters by tags (AND logic)
	Tags []string

	// Publisher filters by publisher address
	Publisher string

	// ApprovalStatus filters by approval status
	ApprovalStatus hpctypes.WorkloadApprovalStatus

	// Offset for pagination
	Offset int

	// Limit for pagination
	Limit int
}

// Matches checks if a template matches the filter
func (f *TemplateFilter) Matches(t *hpctypes.WorkloadTemplate) bool {
	if f.Type != "" && t.Type != f.Type {
		return false
	}

	if len(f.Tags) > 0 {
		for _, tag := range f.Tags {
			if !containsTag(t.Tags, tag) {
				return false
			}
		}
	}

	if f.Publisher != "" && t.Publisher != f.Publisher {
		return false
	}

	if f.ApprovalStatus != "" && t.ApprovalStatus != f.ApprovalStatus {
		return false
	}

	return true
}

// DiscoveryInfo contains template discovery information
type DiscoveryInfo struct {
	// TotalTemplates is the total template count
	TotalTemplates int `json:"total_templates"`

	// TypeCounts is the count per type
	TypeCounts map[hpctypes.WorkloadType]int `json:"type_counts"`

	// Tags is the list of all tags
	Tags []string `json:"tags"`

	// FeaturedTemplates is a list of featured template IDs
	FeaturedTemplates []string `json:"featured_templates"`
}

// GetDiscoveryInfo returns discovery information about the registry
func (r *WorkloadRegistry) GetDiscoveryInfo(ctx context.Context) (*DiscoveryInfo, error) {
	types, err := r.ListTypes(ctx)
	if err != nil {
		return nil, err
	}

	tags, err := r.ListTags(ctx)
	if err != nil {
		return nil, err
	}

	// Featured templates are built-in ones
	featured := make([]string, 0)
	for _, t := range GetBuiltinTemplates() {
		featured = append(featured, t.TemplateID)
	}

	return &DiscoveryInfo{
		TotalTemplates:    r.Count(),
		TypeCounts:        types,
		Tags:              tags,
		FeaturedTemplates: featured,
	}, nil
}

// Helper functions

func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func containsTag(tags []string, tag string) bool {
	tag = strings.ToLower(tag)
	for _, t := range tags {
		if strings.ToLower(t) == tag {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) []string {
	result := make([]string, 0, len(slice))
	for _, item := range slice {
		if item != s {
			result = append(result, item)
		}
	}
	return result
}
