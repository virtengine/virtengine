package fundauth

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
)

const (
	ValueSourceInventoryVersion = uint32(1)
	RuntimeEnforcementNotWired  = "not_wired_t4_16_required"
)

type ValueSourceInventory struct {
	Version            uint32                  `json:"version"`
	RegistryDigestHex  string                  `json:"registry_digest"`
	RuntimeEnforcement string                  `json:"runtime_enforcement"`
	Sources            []ValueSourceEntry      `json:"sources"`
	Exclusions         []ValueSourceExclusion  `json:"exclusions"`
	Detectors          []SourceDetectorPattern `json:"detectors"`
	ActiveSourceIDs    []string                `json:"active_source_ids"`
	ActiveSourceDigest string                  `json:"active_source_digest"`
}

type ValueSourceEntry struct {
	SourceID      string `json:"source_id"`
	TypeURL       string `json:"type_url"`
	Status        string `json:"status"`
	Effect        string `json:"effect"`
	Phase         string `json:"phase"`
	OwningPackage string `json:"owning_package"`
	File          string `json:"file"`
	Symbol        string `json:"symbol"`
	EvidencePath  string `json:"evidence_path"`
}

type ValueSourceExclusion struct {
	Source       string `json:"source"`
	Reason       string `json:"reason"`
	EvidencePath string `json:"evidence_path"`
}

type SourceDetectorPattern struct {
	ID              string   `json:"id"`
	Kind            string   `json:"kind"`
	Path            string   `json:"path"`
	Markers         []string `json:"markers"`
	SourceIDs       []string `json:"source_ids"`
	ExcludedSources []string `json:"excluded_sources"`
}

func ParseValueSourceInventory(data []byte) (ValueSourceInventory, error) {
	var inventory ValueSourceInventory
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&inventory); err != nil {
		return ValueSourceInventory{}, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			err = errors.New("trailing inventory data")
		}
		return ValueSourceInventory{}, err
	}
	return inventory, nil
}

func CanonicalValueSourceInventory(inventory ValueSourceInventory) ([]byte, Digest, error) {
	if err := ValidateValueSourceInventory(inventory, DefaultRegistry(), ExcludedSources()); err != nil {
		return nil, Digest{}, err
	}
	data, err := json.Marshal(inventory)
	if err != nil {
		return nil, Digest{}, err
	}
	return data, sha256.Sum256(data), nil
}

func ValidateValueSourceInventory(inventory ValueSourceInventory, registry *Registry, exclusions []Exclusion) error {
	if inventory.Version != ValueSourceInventoryVersion || inventory.RuntimeEnforcement != RuntimeEnforcementNotWired {
		return errors.New("unsupported inventory version or runtime enforcement claim")
	}
	if registry == nil {
		return errors.New("nil registry")
	}
	wantRegistryDigest := registry.Digest()
	if inventory.RegistryDigestHex != hex.EncodeToString(wantRegistryDigest[:]) {
		return errors.New("inventory registry digest mismatch")
	}
	if !sort.SliceIsSorted(inventory.Sources, func(i, j int) bool { return inventory.Sources[i].SourceID < inventory.Sources[j].SourceID }) || len(inventory.Sources) != len(registry.descriptors) {
		return errors.New("inventory sources are not canonical or complete")
	}
	entries := make(map[string]ValueSourceEntry, len(inventory.Sources))
	for _, entry := range inventory.Sources {
		if entry.SourceID == "" || entry.OwningPackage == "" || entry.EvidencePath == "" || entry.File == "" || entry.Symbol == "" {
			return fmt.Errorf("incomplete source evidence %q", entry.SourceID)
		}
		if _, exists := entries[entry.SourceID]; exists {
			return fmt.Errorf("duplicate inventory source %q", entry.SourceID)
		}
		entries[entry.SourceID] = entry
	}
	active := make([]string, 0, len(registry.descriptors))
	for _, descriptor := range registry.descriptors {
		entry, exists := entries[descriptor.SourceID]
		if !exists {
			return fmt.Errorf("omitted registry source %q", descriptor.SourceID)
		}
		if entry.TypeURL != descriptor.TypeURL || entry.Status != sourceStatusName(descriptor.Status) || entry.Effect != effectName(descriptor.Effect) || entry.Phase != phaseName(descriptor.Phase) {
			return fmt.Errorf("inventory descriptor mismatch %q", descriptor.SourceID)
		}
		if descriptor.Status == SourceStatusActive {
			active = append(active, descriptor.SourceID)
		}
	}
	if !equalStrings(inventory.ActiveSourceIDs, active) || inventory.ActiveSourceDigest != sourceListDigest(active) {
		return errors.New("active source list or digest mismatch")
	}

	if !sort.SliceIsSorted(inventory.Exclusions, func(i, j int) bool { return inventory.Exclusions[i].Source < inventory.Exclusions[j].Source }) || len(inventory.Exclusions) != len(exclusions) {
		return errors.New("inventory exclusions are not canonical or complete")
	}
	registered := make(map[string]SourceDescriptor, len(registry.descriptors))
	for _, descriptor := range registry.descriptors {
		registered[descriptor.SourceID] = descriptor
	}
	excluded := make(map[string]struct{}, len(exclusions))
	for index, exclusion := range inventory.Exclusions {
		if exclusion.Source != exclusions[index].Source || exclusion.Reason != exclusions[index].Reason || exclusion.EvidencePath == "" {
			return fmt.Errorf("exclusion mismatch %q", exclusion.Source)
		}
		if descriptor, exists := registered[exclusion.Source]; exists && descriptor.CurrentMutation {
			return fmt.Errorf("active value-moving source excluded %q", exclusion.Source)
		}
		excluded[exclusion.Source] = struct{}{}
	}
	if len(inventory.Detectors) == 0 || !sort.SliceIsSorted(inventory.Detectors, func(i, j int) bool { return inventory.Detectors[i].ID < inventory.Detectors[j].ID }) {
		return errors.New("detector patterns are empty or noncanonical")
	}
	seenDetector := make(map[string]struct{}, len(inventory.Detectors))
	for _, detector := range inventory.Detectors {
		if detector.ID == "" || detector.Path == "" || detector.Kind == "" || len(detector.Markers) == 0 || len(detector.SourceIDs)+len(detector.ExcludedSources) == 0 {
			return fmt.Errorf("incomplete detector %q", detector.ID)
		}
		if _, exists := seenDetector[detector.ID]; exists {
			return fmt.Errorf("duplicate detector %q", detector.ID)
		}
		seenDetector[detector.ID] = struct{}{}
		for _, sourceID := range detector.SourceIDs {
			if _, exists := entries[sourceID]; !exists {
				return fmt.Errorf("detector %q has unknown source %q", detector.ID, sourceID)
			}
		}
		for _, source := range detector.ExcludedSources {
			if _, exists := excluded[source]; !exists {
				return fmt.Errorf("detector %q has unknown exclusion %q", detector.ID, source)
			}
		}
	}
	return nil
}

func sourceListDigest(sourceIDs []string) string {
	hash := sha256.New()
	for _, sourceID := range sourceIDs {
		_, _ = hash.Write([]byte(sourceID))
		_, _ = hash.Write([]byte{0})
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func sourceStatusName(status SourceStatus) string {
	return [...]string{"", "active", "deferred", "planned", "internal"}[status]
}

func phaseName(phase Phase) string {
	return [...]string{"", "immediate", "deferred", "internal", "control"}[phase]
}

func effectName(effect Effect) string {
	return [...]string{"", "issuance_mint", "transfer", "reward", "escrow_lock", "escrow_release", "refund", "settlement", "payout", "withdrawal", "recovery_control", "treasury"}[effect]
}
