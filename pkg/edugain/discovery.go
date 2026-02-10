// Package edugain provides EduGAIN federation integration.
//
// VE-908: EduGAIN federation support for academic/research institution SSO
package edugain

import (
	"context"
	"strings"
	"sync"
)

// ============================================================================
// Discovery Service Implementation
// ============================================================================

// discoveryService implements the DiscoveryService interface
type discoveryService struct {
	metadata   MetadataService
	usageCache map[string]int64    // entityID -> usage count
	recentUse  map[string][]string // walletAddress -> entityIDs
	mu         sync.RWMutex
}

// newDiscoveryService creates a new discovery service
func newDiscoveryService(metadata MetadataService) DiscoveryService {
	return &discoveryService{
		metadata:   metadata,
		usageCache: make(map[string]int64),
		recentUse:  make(map[string][]string),
	}
}

// Search searches for institutions by query
func (d *discoveryService) Search(ctx context.Context, query InstitutionSearchQuery) (*InstitutionSearchResult, error) {
	return d.metadata.SearchInstitutions(query)
}

// GetByEntityID returns an institution by entity ID
func (d *discoveryService) GetByEntityID(ctx context.Context, entityID string) (*Institution, error) {
	return d.metadata.FindInstitution(entityID)
}

// GetByCountry returns institutions by country code
func (d *discoveryService) GetByCountry(ctx context.Context, countryCode string, limit int) ([]Institution, error) {
	result, err := d.metadata.SearchInstitutions(InstitutionSearchQuery{
		Country: strings.ToUpper(countryCode),
		Limit:   limit,
	})
	if err != nil {
		return nil, err
	}
	return result.Institutions, nil
}

// GetByFederation returns institutions by federation name
func (d *discoveryService) GetByFederation(ctx context.Context, federationName string, limit int) ([]Institution, error) {
	result, err := d.metadata.SearchInstitutions(InstitutionSearchQuery{
		Federation: federationName,
		Limit:      limit,
	})
	if err != nil {
		return nil, err
	}
	return result.Institutions, nil
}

// GetPopular returns popular/frequently used institutions
func (d *discoveryService) GetPopular(ctx context.Context, limit int) ([]Institution, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Get all institutions sorted by usage
	type usageEntry struct {
		entityID string
		count    int64
	}

	entries := make([]usageEntry, 0, len(d.usageCache))
	for entityID, count := range d.usageCache {
		entries = append(entries, usageEntry{entityID, count})
	}

	// Sort by count descending
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].count > entries[i].count {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// Limit results
	if limit > 0 && limit < len(entries) {
		entries = entries[:limit]
	}

	// Look up institutions
	institutions := make([]Institution, 0, len(entries))
	for _, entry := range entries {
		if inst, err := d.metadata.FindInstitution(entry.entityID); err == nil {
			institutions = append(institutions, *inst)
		}
	}

	return institutions, nil
}

// GetRecent returns recently used institutions for a wallet
func (d *discoveryService) GetRecent(ctx context.Context, walletAddress string, limit int) ([]Institution, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	entityIDs, ok := d.recentUse[walletAddress]
	if !ok {
		return nil, nil
	}

	// Limit results
	if limit > 0 && limit < len(entityIDs) {
		entityIDs = entityIDs[:limit]
	}

	// Look up institutions
	institutions := make([]Institution, 0, len(entityIDs))
	for _, entityID := range entityIDs {
		if inst, err := d.metadata.FindInstitution(entityID); err == nil {
			institutions = append(institutions, *inst)
		}
	}

	return institutions, nil
}

// RecordUsage records institution usage for popularity tracking
func (d *discoveryService) RecordUsage(ctx context.Context, entityID string, walletAddress string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Update global usage count
	d.usageCache[entityID]++

	// Update recent use for wallet
	recent := d.recentUse[walletAddress]

	// Remove if already present
	for i, id := range recent {
		if id == entityID {
			recent = append(recent[:i], recent[i+1:]...)
			break
		}
	}

	// Add to front
	recent = append([]string{entityID}, recent...)

	// Limit to 10 recent
	if len(recent) > 10 {
		recent = recent[:10]
	}

	d.recentUse[walletAddress] = recent

	return nil
}

// GetStats returns discovery statistics
func (d *discoveryService) GetStats(ctx context.Context) (*DiscoveryStats, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	stats := &DiscoveryStats{
		TopCountries:    make(map[string]int64),
		TopInstitutions: make([]InstitutionUsage, 0),
	}

	// Count total searches (sum of all usage)
	for _, count := range d.usageCache {
		stats.TotalSearches += count
	}

	// Get top institutions
	type usageEntry struct {
		entityID string
		count    int64
	}

	entries := make([]usageEntry, 0, len(d.usageCache))
	for entityID, count := range d.usageCache {
		entries = append(entries, usageEntry{entityID, count})
	}

	// Sort by count descending
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].count > entries[i].count {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// Get top 10
	limit := 10
	if limit > len(entries) {
		limit = len(entries)
	}

	for _, entry := range entries[:limit] {
		displayName := entry.entityID
		if inst, err := d.metadata.FindInstitution(entry.entityID); err == nil {
			displayName = inst.DisplayName

			// Count by country
			stats.TopCountries[inst.Country] += entry.count
		}

		stats.TopInstitutions = append(stats.TopInstitutions, InstitutionUsage{
			EntityID:    entry.entityID,
			DisplayName: displayName,
			UsageCount:  entry.count,
		})
	}

	return stats, nil
}
