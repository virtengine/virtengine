package provider_daemon

import (
	"context"
	"testing"
	"time"
)

func TestDefaultCategories(t *testing.T) {
	// Verify all 7 default categories are defined
	expectedCategories := []string{"Compute", "HPC", "GPU", "Storage", "Network", "TEE", "AI/ML"}

	if len(DefaultCategories) != len(expectedCategories) {
		t.Errorf("expected %d categories, got %d", len(expectedCategories), len(DefaultCategories))
	}

	categoryTitles := make(map[string]bool)
	for _, cat := range DefaultCategories {
		categoryTitles[cat.Title] = true
	}

	for _, expected := range expectedCategories {
		if !categoryTitles[expected] {
			t.Errorf("missing expected category: %s", expected)
		}
	}
}

func TestCategoryDefinitionValidation(t *testing.T) {
	for _, cat := range DefaultCategories {
		if cat.Title == "" {
			t.Error("category has empty title")
		}
		if cat.Description == "" {
			t.Errorf("category %s has empty description", cat.Title)
		}
		if cat.Priority <= 0 {
			t.Errorf("category %s has invalid priority: %d", cat.Title, cat.Priority)
		}
	}
}

func TestCategorySyncStateInitialization(t *testing.T) {
	state := &CategorySyncState{
		Mappings:    make(map[string]*CategoryMapping),
		SyncVersion: "1.0.0",
	}

	if state.Mappings == nil {
		t.Error("mappings should not be nil")
	}

	if state.SyncVersion != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", state.SyncVersion)
	}
}

func TestDefaultCategorySyncConfig(t *testing.T) {
	cfg := DefaultCategorySyncConfig()

	if cfg.StateFilePath == "" {
		t.Error("state file path should have default value")
	}

	if cfg.SyncIntervalSeconds <= 0 {
		t.Error("sync interval should be positive")
	}

	if cfg.MaxRetries <= 0 {
		t.Error("max retries should be positive")
	}

	if cfg.OperationTimeout <= 0 {
		t.Error("operation timeout should be positive")
	}
}

func TestCategorySyncStateCopy(t *testing.T) {
	state := &CategorySyncState{
		Mappings:    make(map[string]*CategoryMapping),
		LastSync:    time.Now().UTC(),
		SyncVersion: "1.0.0",
	}

	state.Mappings["Compute"] = &CategoryMapping{
		Title:      "Compute",
		WaldurUUID: "test-uuid-123",
		SyncedAt:   time.Now().UTC(),
	}

	// Create a copy
	stateCopy := &CategorySyncState{
		Mappings:    make(map[string]*CategoryMapping),
		LastSync:    state.LastSync,
		SyncVersion: state.SyncVersion,
	}
	for k, v := range state.Mappings {
		stateCopy.Mappings[k] = &CategoryMapping{
			Title:       v.Title,
			WaldurUUID:  v.WaldurUUID,
			Description: v.Description,
			SyncedAt:    v.SyncedAt,
		}
	}

	// Modify original
	state.Mappings["Compute"].WaldurUUID = "modified-uuid"

	// Verify copy is not affected
	if stateCopy.Mappings["Compute"].WaldurUUID == "modified-uuid" {
		t.Error("copy should not be affected by modifications to original")
	}
}

func TestCategoryMappingLookup(t *testing.T) {
	state := &CategorySyncState{
		Mappings: map[string]*CategoryMapping{
			"Compute": {
				Title:      "Compute",
				WaldurUUID: "compute-uuid",
			},
			"GPU": {
				Title:      "GPU",
				WaldurUUID: "gpu-uuid",
			},
		},
	}

	// Test existing category
	mapping, ok := state.Mappings["Compute"]
	if !ok {
		t.Error("Compute mapping should exist")
	}
	if mapping.WaldurUUID != "compute-uuid" {
		t.Errorf("expected compute-uuid, got %s", mapping.WaldurUUID)
	}

	// Test non-existent category
	_, ok = state.Mappings["NonExistent"]
	if ok {
		t.Error("NonExistent mapping should not exist")
	}
}

func TestLoadCategoriesFromFile(t *testing.T) {
	// Test loading from the actual fixture file
	categories, err := LoadCategoriesFromFile("../../scripts/waldur/categories.json")
	if err != nil {
		t.Skipf("skipping test: categories file not found: %v", err)
	}

	if len(categories) < 7 {
		t.Errorf("expected at least 7 categories, got %d", len(categories))
	}

	// Verify all expected categories are present
	expectedTitles := map[string]bool{
		"Compute": true,
		"HPC":     true,
		"GPU":     true,
		"Storage": true,
		"Network": true,
		"TEE":     true,
		"AI/ML":   true,
	}

	for _, cat := range categories {
		delete(expectedTitles, cat.Title)
	}

	if len(expectedTitles) > 0 {
		t.Errorf("missing categories: %v", expectedTitles)
	}
}

func TestInitCategoriesResultStructure(t *testing.T) {
	result := &InitCategoriesResult{
		Created:  []string{"Compute", "GPU"},
		Existing: []string{"Storage"},
		Failed:   []string{},
		Mappings: map[string]string{
			"Compute": "uuid-1",
			"GPU":     "uuid-2",
			"Storage": "uuid-3",
		},
	}

	if len(result.Created) != 2 {
		t.Errorf("expected 2 created, got %d", len(result.Created))
	}

	if len(result.Existing) != 1 {
		t.Errorf("expected 1 existing, got %d", len(result.Existing))
	}

	if len(result.Failed) != 0 {
		t.Errorf("expected 0 failed, got %d", len(result.Failed))
	}

	if len(result.Mappings) != 3 {
		t.Errorf("expected 3 mappings, got %d", len(result.Mappings))
	}
}

func TestCategorySyncWorkerCreation(t *testing.T) {
	// Test that worker creation fails without client
	cfg := DefaultCategorySyncConfig()
	_, err := NewCategorySyncWorker(cfg, nil)
	if err == nil {
		t.Error("expected error when creating worker without client")
	}
}

func TestCategorySyncWorkerGetCategoryUUID(t *testing.T) {
	// Create a worker with mock state
	worker := &CategorySyncWorker{
		cfg: DefaultCategorySyncConfig(),
		state: &CategorySyncState{
			Mappings: map[string]*CategoryMapping{
				"Compute": {
					Title:      "Compute",
					WaldurUUID: "test-compute-uuid",
				},
			},
		},
	}

	// Test existing category
	uuid, ok := worker.GetCategoryUUID("Compute")
	if !ok {
		t.Error("Compute should exist")
	}
	if uuid != "test-compute-uuid" {
		t.Errorf("expected test-compute-uuid, got %s", uuid)
	}

	// Test non-existent category
	_, ok = worker.GetCategoryUUID("NonExistent")
	if ok {
		t.Error("NonExistent should not exist")
	}
}

func TestCategorySyncWorkerGetAllMappings(t *testing.T) {
	worker := &CategorySyncWorker{
		cfg: DefaultCategorySyncConfig(),
		state: &CategorySyncState{
			Mappings: map[string]*CategoryMapping{
				"Compute": {Title: "Compute", WaldurUUID: "uuid-1"},
				"GPU":     {Title: "GPU", WaldurUUID: "uuid-2"},
			},
		},
	}

	mappings := worker.GetAllMappings()

	if len(mappings) != 2 {
		t.Errorf("expected 2 mappings, got %d", len(mappings))
	}

	if mappings["Compute"] != "uuid-1" {
		t.Errorf("expected uuid-1 for Compute, got %s", mappings["Compute"])
	}

	if mappings["GPU"] != "uuid-2" {
		t.Errorf("expected uuid-2 for GPU, got %s", mappings["GPU"])
	}
}

func TestCategorySyncWorkerStateCopy(t *testing.T) {
	worker := &CategorySyncWorker{
		cfg: DefaultCategorySyncConfig(),
		state: &CategorySyncState{
			Mappings: map[string]*CategoryMapping{
				"Compute": {Title: "Compute", WaldurUUID: "uuid-1"},
			},
			LastSync:    time.Now().UTC(),
			SyncVersion: "1.0.0",
		},
	}

	// Get state copy
	stateCopy := worker.State()

	// Modify original
	worker.state.Mappings["Compute"].WaldurUUID = "modified"

	// Verify copy is unchanged
	if stateCopy.Mappings["Compute"].WaldurUUID == "modified" {
		t.Error("state copy should not be affected by modifications")
	}
}

func TestCategorySyncWorkerDisabledStart(t *testing.T) {
	worker := &CategorySyncWorker{
		cfg: CategorySyncConfig{Enabled: false},
		state: &CategorySyncState{
			Mappings: make(map[string]*CategoryMapping),
		},
	}

	// Start should return nil when disabled
	err := worker.Start(context.Background())
	if err != nil {
		t.Errorf("expected nil error when disabled, got %v", err)
	}
}
