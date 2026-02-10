// Package hpc_workload_library provides tests for registry functionality.
//
// VE-5F: Tests for workload registry
package hpc_workload_library

import (
	"context"
	"testing"
	"time"

	_ "github.com/virtengine/virtengine/sdk/go/sdkutil" // Initialize SDK config

	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

func TestNewWorkloadRegistry(t *testing.T) {
	config := DefaultRegistryConfig()
	registry := NewWorkloadRegistry(nil, config)

	if registry == nil {
		t.Fatal("expected registry to be created")
	}

	// Should have built-in templates
	count := registry.Count()
	if count < 5 {
		t.Errorf("expected at least 5 built-in templates, got %d", count)
	}
}

func TestRegistryGet(t *testing.T) {
	config := DefaultRegistryConfig()
	registry := NewWorkloadRegistry(nil, config)
	ctx := context.Background()

	// Should find built-in templates
	template, err := registry.Get(ctx, "mpi-standard")
	if err != nil {
		t.Errorf("failed to get mpi-standard template: %v", err)
	}
	if template == nil {
		t.Fatal("expected template to be found")
	}
	if template.TemplateID != "mpi-standard" {
		t.Errorf("expected template ID 'mpi-standard', got %s", template.TemplateID)
	}

	// Should not find non-existent template
	_, err = registry.Get(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent template")
	}
}

func TestRegistryGetVersion(t *testing.T) {
	config := DefaultRegistryConfig()
	registry := NewWorkloadRegistry(nil, config)
	ctx := context.Background()

	// Should find template by version
	template, err := registry.GetVersion(ctx, "mpi-standard", "1.0.0")
	if err != nil {
		t.Errorf("failed to get template version: %v", err)
	}
	if template == nil {
		t.Fatal("expected template to be found")
	}

	// Should not find non-existent version
	_, err = registry.GetVersion(ctx, "mpi-standard", "999.0.0")
	if err == nil {
		t.Error("expected error for nonexistent version")
	}
}

func TestRegistryList(t *testing.T) {
	config := DefaultRegistryConfig()
	registry := NewWorkloadRegistry(nil, config)
	ctx := context.Background()

	// List all templates
	templates, err := registry.List(ctx, nil)
	if err != nil {
		t.Errorf("failed to list templates: %v", err)
	}
	if len(templates) < 5 {
		t.Errorf("expected at least 5 templates, got %d", len(templates))
	}

	// List with type filter
	templates, err = registry.List(ctx, &TemplateFilter{
		Type: hpctypes.WorkloadTypeMPI,
	})
	if err != nil {
		t.Errorf("failed to list MPI templates: %v", err)
	}
	if len(templates) == 0 {
		t.Error("expected at least one MPI template")
	}
	for _, tmpl := range templates {
		if tmpl.Type != hpctypes.WorkloadTypeMPI {
			t.Errorf("expected MPI type, got %s", tmpl.Type)
		}
	}
}

func TestRegistryListByType(t *testing.T) {
	config := DefaultRegistryConfig()
	registry := NewWorkloadRegistry(nil, config)
	ctx := context.Background()

	tests := []struct {
		workloadType hpctypes.WorkloadType
		expectCount  int
	}{
		{hpctypes.WorkloadTypeMPI, 1},
		{hpctypes.WorkloadTypeGPU, 1},
		{hpctypes.WorkloadTypeBatch, 1},
		{hpctypes.WorkloadTypeDataProcessing, 1},
		{hpctypes.WorkloadTypeInteractive, 1},
	}

	for _, tt := range tests {
		templates, err := registry.ListByType(ctx, tt.workloadType)
		if err != nil {
			t.Errorf("failed to list %s templates: %v", tt.workloadType, err)
			continue
		}
		if len(templates) < tt.expectCount {
			t.Errorf("expected at least %d %s templates, got %d",
				tt.expectCount, tt.workloadType, len(templates))
		}
	}
}

func TestRegistryListByTag(t *testing.T) {
	config := DefaultRegistryConfig()
	registry := NewWorkloadRegistry(nil, config)
	ctx := context.Background()

	// List by tag
	templates, err := registry.ListByTag(ctx, "gpu")
	if err != nil {
		t.Errorf("failed to list templates by tag: %v", err)
	}
	if len(templates) == 0 {
		t.Error("expected at least one template with 'gpu' tag")
	}
}

func TestRegistryListTags(t *testing.T) {
	config := DefaultRegistryConfig()
	registry := NewWorkloadRegistry(nil, config)
	ctx := context.Background()

	tags, err := registry.ListTags(ctx)
	if err != nil {
		t.Errorf("failed to list tags: %v", err)
	}
	if len(tags) == 0 {
		t.Error("expected at least one tag")
	}

	// Check for expected tags
	expectedTags := []string{"mpi", "gpu", "batch", "interactive"}
	for _, expected := range expectedTags {
		if !contains(tags, expected) {
			t.Errorf("expected tag '%s' not found", expected)
		}
	}
}

func TestRegistryListTypes(t *testing.T) {
	config := DefaultRegistryConfig()
	registry := NewWorkloadRegistry(nil, config)
	ctx := context.Background()

	types, err := registry.ListTypes(ctx)
	if err != nil {
		t.Errorf("failed to list types: %v", err)
	}
	if len(types) == 0 {
		t.Error("expected at least one type")
	}

	// Check for expected types
	expectedTypes := []hpctypes.WorkloadType{
		hpctypes.WorkloadTypeMPI,
		hpctypes.WorkloadTypeGPU,
		hpctypes.WorkloadTypeBatch,
	}
	for _, expected := range expectedTypes {
		if _, ok := types[expected]; !ok {
			t.Errorf("expected type '%s' not found", expected)
		}
	}
}

func TestRegistrySearch(t *testing.T) {
	config := DefaultRegistryConfig()
	registry := NewWorkloadRegistry(nil, config)
	ctx := context.Background()

	tests := []struct {
		query       string
		expectMatch bool
	}{
		{"mpi", true},
		{"gpu", true},
		{"batch", true},
		{"parallel", true},
		{"nonexistent-query-xyz", false},
	}

	for _, tt := range tests {
		templates, err := registry.Search(ctx, tt.query, 10)
		if err != nil {
			t.Errorf("search failed for '%s': %v", tt.query, err)
			continue
		}
		if tt.expectMatch && len(templates) == 0 {
			t.Errorf("expected search results for '%s'", tt.query)
		}
		if !tt.expectMatch && len(templates) > 0 {
			t.Errorf("expected no results for '%s', got %d", tt.query, len(templates))
		}
	}
}

func TestRegistryRegister(t *testing.T) {
	config := DefaultRegistryConfig()
	config.RequireApproval = false // Allow unapproved for testing
	registry := NewWorkloadRegistry(nil, config)
	ctx := context.Background()

	template := createValidTemplate()
	template.TemplateID = "custom-template-1"
	template.ApprovalStatus = hpctypes.WorkloadApprovalPending

	err := registry.Register(ctx, template)
	if err != nil {
		t.Errorf("failed to register template: %v", err)
	}

	// Should be findable
	found, err := registry.Get(ctx, "custom-template-1")
	if err != nil {
		t.Errorf("failed to get registered template: %v", err)
	}
	if found == nil {
		t.Error("expected template to be found")
	}

	// Count should increase
	if registry.Count() < 6 {
		t.Error("expected count to increase after registration")
	}
}

func TestRegistryRegisterApprovalRequired(t *testing.T) {
	config := DefaultRegistryConfig()
	config.RequireApproval = true
	registry := NewWorkloadRegistry(nil, config)
	ctx := context.Background()

	template := createValidTemplate()
	template.TemplateID = "custom-template-2"
	template.ApprovalStatus = hpctypes.WorkloadApprovalPending

	err := registry.Register(ctx, template)
	if err == nil {
		t.Error("expected error for unapproved template")
	}

	// Approved template should work
	template.ApprovalStatus = hpctypes.WorkloadApprovalApproved
	err = registry.Register(ctx, template)
	if err != nil {
		t.Errorf("failed to register approved template: %v", err)
	}
}

func TestRegistryUnregister(t *testing.T) {
	config := DefaultRegistryConfig()
	config.RequireApproval = false
	registry := NewWorkloadRegistry(nil, config)
	ctx := context.Background()

	template := createValidTemplate()
	template.TemplateID = "custom-template-3"

	err := registry.Register(ctx, template)
	if err != nil {
		t.Fatalf("failed to register template: %v", err)
	}

	initialCount := registry.Count()

	err = registry.Unregister(ctx, "custom-template-3")
	if err != nil {
		t.Errorf("failed to unregister template: %v", err)
	}

	if registry.Count() >= initialCount {
		t.Error("expected count to decrease after unregister")
	}

	// Should not be findable
	_, err = registry.Get(ctx, "custom-template-3")
	if err == nil {
		t.Error("expected error for unregistered template")
	}
}

func TestRegistryPagination(t *testing.T) {
	config := DefaultRegistryConfig()
	registry := NewWorkloadRegistry(nil, config)
	ctx := context.Background()

	// Get all templates
	all, err := registry.List(ctx, nil)
	if err != nil {
		t.Fatalf("failed to list all templates: %v", err)
	}

	// Get first page
	page1, err := registry.List(ctx, &TemplateFilter{
		Offset: 0,
		Limit:  2,
	})
	if err != nil {
		t.Errorf("failed to get first page: %v", err)
	}
	if len(page1) > 2 {
		t.Errorf("expected max 2 templates, got %d", len(page1))
	}

	// Get second page
	page2, err := registry.List(ctx, &TemplateFilter{
		Offset: 2,
		Limit:  2,
	})
	if err != nil {
		t.Errorf("failed to get second page: %v", err)
	}

	// Pages should not overlap
	for _, p1 := range page1 {
		for _, p2 := range page2 {
			if p1.TemplateID == p2.TemplateID {
				t.Errorf("duplicate template in pages: %s", p1.TemplateID)
			}
		}
	}

	// All pages combined should equal total
	if len(page1)+len(page2) > len(all) {
		t.Error("pagination returned more items than total")
	}
}

func TestRegistryGetDiscoveryInfo(t *testing.T) {
	config := DefaultRegistryConfig()
	registry := NewWorkloadRegistry(nil, config)
	ctx := context.Background()

	info, err := registry.GetDiscoveryInfo(ctx)
	if err != nil {
		t.Errorf("failed to get discovery info: %v", err)
	}

	if info.TotalTemplates < 5 {
		t.Errorf("expected at least 5 templates, got %d", info.TotalTemplates)
	}

	if len(info.TypeCounts) == 0 {
		t.Error("expected type counts")
	}

	if len(info.Tags) == 0 {
		t.Error("expected tags")
	}

	if len(info.FeaturedTemplates) == 0 {
		t.Error("expected featured templates")
	}
}

func TestTemplateFilterMatches(t *testing.T) {
	template := &hpctypes.WorkloadTemplate{
		TemplateID:     "test-template",
		Type:           hpctypes.WorkloadTypeMPI,
		Tags:           []string{"mpi", "parallel"},
		Publisher:      "ve1365yvmc4s7awdyj3n2sav7xfx76adc6dzaf4vr",
		ApprovalStatus: hpctypes.WorkloadApprovalApproved,
	}

	tests := []struct {
		name    string
		filter  *TemplateFilter
		matches bool
	}{
		{
			name:    "nil filter matches all",
			filter:  nil,
			matches: true,
		},
		{
			name:    "empty filter matches all",
			filter:  &TemplateFilter{},
			matches: true,
		},
		{
			name:    "matching type",
			filter:  &TemplateFilter{Type: hpctypes.WorkloadTypeMPI},
			matches: true,
		},
		{
			name:    "non-matching type",
			filter:  &TemplateFilter{Type: hpctypes.WorkloadTypeGPU},
			matches: false,
		},
		{
			name:    "matching tag",
			filter:  &TemplateFilter{Tags: []string{"mpi"}},
			matches: true,
		},
		{
			name:    "matching multiple tags",
			filter:  &TemplateFilter{Tags: []string{"mpi", "parallel"}},
			matches: true,
		},
		{
			name:    "non-matching tag",
			filter:  &TemplateFilter{Tags: []string{"gpu"}},
			matches: false,
		},
		{
			name:    "matching publisher",
			filter:  &TemplateFilter{Publisher: "ve1365yvmc4s7awdyj3n2sav7xfx76adc6dzaf4vr"},
			matches: true,
		},
		{
			name:    "non-matching publisher",
			filter:  &TemplateFilter{Publisher: "ve1other"},
			matches: false,
		},
		{
			name:    "matching approval status",
			filter:  &TemplateFilter{ApprovalStatus: hpctypes.WorkloadApprovalApproved},
			matches: true,
		},
		{
			name:    "non-matching approval status",
			filter:  &TemplateFilter{ApprovalStatus: hpctypes.WorkloadApprovalPending},
			matches: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.filter == nil {
				// nil filter always matches
				return
			}
			result := tt.filter.Matches(template)
			if result != tt.matches {
				t.Errorf("Matches() = %v, want %v", result, tt.matches)
			}
		})
	}
}

func TestContainsTag(t *testing.T) {
	tags := []string{"mpi", "GPU", "Parallel"}

	tests := []struct {
		tag      string
		expected bool
	}{
		{"mpi", true},
		{"MPI", true}, // Case insensitive
		{"gpu", true},
		{"parallel", true},
		{"nonexistent", false},
	}

	for _, tt := range tests {
		result := containsTag(tags, tt.tag)
		if result != tt.expected {
			t.Errorf("containsTag(%v, %s) = %v, want %v", tags, tt.tag, result, tt.expected)
		}
	}
}

func TestRemoveString(t *testing.T) {
	slice := []string{"a", "b", "c", "d"}

	result := removeString(slice, "b")

	if len(result) != 3 {
		t.Errorf("expected 3 elements, got %d", len(result))
	}

	if contains(result, "b") {
		t.Error("expected 'b' to be removed")
	}

	// Original should be unchanged
	if len(slice) != 4 {
		t.Error("original slice should be unchanged")
	}
}

func TestRegistryConcurrentAccess(t *testing.T) {
	config := DefaultRegistryConfig()
	config.RequireApproval = false
	registry := NewWorkloadRegistry(nil, config)
	ctx := context.Background()

	// Concurrent reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, _ = registry.List(ctx, nil)
			_, _ = registry.Search(ctx, "mpi", 10)
			_, _ = registry.Get(ctx, "mpi-standard")
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for concurrent reads")
		}
	}
}
