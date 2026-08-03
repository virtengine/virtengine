package fundauth

import (
	"encoding/hex"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const inventoryFixturePath = "testdata/value_source_inventory_v1.json"

func loadInventoryFixture(t *testing.T) ValueSourceInventory {
	t.Helper()
	data, err := os.ReadFile(inventoryFixturePath)
	if err != nil {
		t.Fatal(err)
	}
	inventory, err := ParseValueSourceInventory(data)
	if err != nil {
		t.Fatal(err)
	}
	return inventory
}

func TestValueSourceInventoryCoverageAndCanonicalDigest(t *testing.T) {
	inventory := loadInventoryFixture(t)
	if err := ValidateValueSourceInventory(inventory, DefaultRegistry(), ExcludedSources()); err != nil {
		t.Fatal(err)
	}
	if len(inventory.Sources) != 40 || len(inventory.ActiveSourceIDs) != 30 || len(inventory.Exclusions) != 7 {
		t.Fatalf("sources=%d active=%d exclusions=%d", len(inventory.Sources), len(inventory.ActiveSourceIDs), len(inventory.Exclusions))
	}
	canonical, digest, err := CanonicalValueSourceInventory(inventory)
	if err != nil {
		t.Fatal(err)
	}
	second, secondDigest, err := CanonicalValueSourceInventory(inventory)
	if err != nil || string(canonical) != string(second) || digest != secondDigest {
		t.Fatal("inventory canonicalization is not deterministic")
	}
	const wantDigest = "3fea71122bb16bfffcb88309610c92e08e74aabced5743ac27458c572df64f74"
	if got := hex.EncodeToString(digest[:]); got != wantDigest {
		t.Fatalf("inventory digest = %s", got)
	}
}

func TestValueSourceInventoryRejectsCoverageDrift(t *testing.T) {
	base := loadInventoryFixture(t)
	tests := map[string]func(*ValueSourceInventory){
		"omitted":       func(value *ValueSourceInventory) { value.Sources = value.Sources[1:] },
		"unknown":       func(value *ValueSourceInventory) { value.Sources[0].SourceID = "/unknown" },
		"duplicate":     func(value *ValueSourceInventory) { value.Sources[1] = value.Sources[0] },
		"status":        func(value *ValueSourceInventory) { value.Sources[0].Status = "planned" },
		"runtime claim": func(value *ValueSourceInventory) { value.RuntimeEnforcement = "wired" },
		"active exclusion": func(value *ValueSourceInventory) {
			value.Exclusions[0].Source = value.Sources[0].SourceID
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			changed := base
			changed.Sources = append([]ValueSourceEntry(nil), base.Sources...)
			changed.Exclusions = append([]ValueSourceExclusion(nil), base.Exclusions...)
			mutate(&changed)
			if err := ValidateValueSourceInventory(changed, DefaultRegistry(), ExcludedSources()); err == nil {
				t.Fatal("inventory drift accepted")
			}
		})
	}
}

func TestGovernedValueSourceDetectors(t *testing.T) {
	inventory := loadInventoryFixture(t)
	repositoryRoot := filepath.Clean(filepath.Join("..", ".."))
	governed := make(map[string]bool)
	for _, detector := range inventory.Detectors {
		path := filepath.Join(repositoryRoot, filepath.FromSlash(detector.Path))
		switch detector.Kind {
		case "text":
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("detector %s: %v", detector.ID, err)
			}
			for _, marker := range detector.Markers {
				if !strings.Contains(string(data), marker) {
					t.Errorf("detector %s marker %q missing", detector.ID, marker)
				}
			}
		case "ast_functions":
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			if err != nil {
				t.Fatalf("detector %s: %v", detector.ID, err)
			}
			functions := make(map[string]bool)
			for _, declaration := range file.Decls {
				if function, ok := declaration.(*ast.FuncDecl); ok {
					functions[function.Name.Name] = true
				}
			}
			for _, marker := range detector.Markers {
				if !functions[marker] {
					t.Errorf("detector %s function %q missing", detector.ID, marker)
				}
			}
		default:
			t.Fatalf("detector %s has unknown kind %q", detector.ID, detector.Kind)
		}
		for _, sourceID := range detector.SourceIDs {
			governed[sourceID] = true
		}
	}
	for _, descriptor := range DefaultRegistry().Descriptors() {
		if descriptor.CurrentMutation && !governed[descriptor.SourceID] {
			t.Errorf("current value-moving source %q lacks a detector classification", descriptor.SourceID)
		}
	}
}
