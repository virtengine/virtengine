// Package contracts defines the fail-closed T5 cross-contract dependency manifest.
package contracts

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
)

const (
	ManifestVersion          = "virtengine.platformsecurity.dependencies/v1"
	T4IntegrationBaselineSHA = "9635513da5e3813e4d30839e0ee22b15f16e3911"
	T4ControlDigest          = "7fade5003634c3310ae16807adfc44674e7ae85bd791392aa9c7adff1b5b8405"
	T4MigrationDigest        = "a78e32dad73854308ca736a7fc390c035902ee8db8e91117693fb416ad9e145f"
)

type DependencyStatus string

const (
	StatusUnavailable DependencyStatus = "unavailable"
	StatusFixtureOnly DependencyStatus = "fixture_only"
	StatusVerified    DependencyStatus = "verified"
)

type ProfileState string

const (
	ProfileDisabled    ProfileState = "disabled"
	ProfileFixtureOnly ProfileState = "fixture_only"
	ProfileSandbox     ProfileState = "sandbox"
	ProfileProduction  ProfileState = "production"
)

type Dependency struct {
	SourceThread string           `json:"source_thread"`
	ContractID   string           `json:"contract_id"`
	Version      string           `json:"version"`
	Digest       string           `json:"digest"`
	Status       DependencyStatus `json:"status"`
	Blocker      string           `json:"blocker"`
	Consumers    []string         `json:"consumers"`
}

type Capability struct {
	TaskID        string       `json:"task_id"`
	ProfileState  ProfileState `json:"profile_state"`
	Registration  bool         `json:"registration"`
	Advertisement bool         `json:"advertisement"`
	Readiness     bool         `json:"readiness"`
	Mutation      bool         `json:"mutation"`
}

type DependencyManifest struct {
	Version                  string       `json:"version"`
	T4IntegrationBaselineSHA string       `json:"t4_integration_baseline_sha"`
	Dependencies             []Dependency `json:"dependencies"`
	Capabilities             []Capability `json:"capabilities"`
}

type dependencySpec struct {
	thread       string
	version      string
	consumers    []string
	pinnedDigest string
}

var allCapabilities = []string{"89A", "89B", "89C", "89D", "90A", "91A"}

var requiredDependencies = map[string]dependencySpec{
	"evidence-envelope":     {"T1", "v1", []string{"89A", "89B", "89C", "89D", "91A"}, ""},
	"runtime-policy":        {"T1", "v1", []string{"89A", "89D", "90A", "91A"}, ""},
	"workload-binding":      {"T1", "v1", []string{"89D", "90A"}, ""},
	"recovery-evidence":     {"T1", "v1", []string{"89A", "89B", "89C", "91A"}, ""},
	"durable-intent":        {"T3", "v1", []string{"89A", "89B", "89C", "89D", "90A", "91A"}, ""},
	"ibc-terminal":          {"T3", "v1", []string{"89C", "89D", "91A"}, ""},
	"reconciliation":        {"T3", "v1", []string{"89B", "89C", "90A", "91A"}, ""},
	"exact-integration-sha": {"T4", "v1", allCapabilities, T4IntegrationBaselineSHA},
	"integration-control":   {"T4", "virtengine.prototype.integration-control/v1", allCapabilities, T4ControlDigest},
	"migration-inventory":   {"T4", "virtengine.task-88a.migration-inventory/v1", allCapabilities, T4MigrationDigest},
}

func (m DependencyManifest) Validate() error {
	if m.Version != ManifestVersion {
		return fmt.Errorf("unknown manifest version %q", m.Version)
	}
	if m.T4IntegrationBaselineSHA != T4IntegrationBaselineSHA {
		return errors.New("T4 integration baseline SHA changed")
	}
	if err := validateDependencies(m.Dependencies); err != nil {
		return err
	}
	return validateCapabilities(m.Capabilities)
}

func validateDependencies(dependencies []Dependency) error {
	seen := make(map[string]bool, len(dependencies))
	for _, dependency := range dependencies {
		spec, ok := requiredDependencies[dependency.ContractID]
		if !ok {
			return fmt.Errorf("unknown dependency %q", dependency.ContractID)
		}
		if seen[dependency.ContractID] {
			return fmt.Errorf("duplicate dependency %q", dependency.ContractID)
		}
		seen[dependency.ContractID] = true
		if dependency.SourceThread != spec.thread || dependency.Version != spec.version {
			return fmt.Errorf("dependency %q has unknown source or version", dependency.ContractID)
		}
		if dependency.Blocker == "" {
			return fmt.Errorf("dependency %q has no blocker", dependency.ContractID)
		}
		if !sameStrings(dependency.Consumers, spec.consumers) {
			return fmt.Errorf("dependency %q has incorrect consumers", dependency.ContractID)
		}
		if err := validateDigestStatus(dependency); err != nil {
			return err
		}
		if spec.pinnedDigest != "" && dependency.Digest != spec.pinnedDigest {
			return fmt.Errorf("dependency %q digest does not match the pinned T4 baseline", dependency.ContractID)
		}
	}
	for contractID := range requiredDependencies {
		if !seen[contractID] {
			return fmt.Errorf("mandatory dependency %q is missing", contractID)
		}
	}
	return nil
}

func validateDigestStatus(dependency Dependency) error {
	switch dependency.Status {
	case StatusUnavailable:
		if dependency.Digest != "" {
			return fmt.Errorf("unavailable dependency %q must not claim a digest", dependency.ContractID)
		}
	case StatusFixtureOnly, StatusVerified:
		if !validDigest(dependency.Digest) {
			return fmt.Errorf("dependency %q requires a lowercase SHA digest", dependency.ContractID)
		}
	default:
		return fmt.Errorf("dependency %q has unknown status %q", dependency.ContractID, dependency.Status)
	}
	return nil
}

func validDigest(value string) bool {
	if len(value) != 40 && len(value) != sha256.Size*2 {
		return false
	}
	decoded, err := hex.DecodeString(value)
	return err == nil && hex.EncodeToString(decoded) == value
}

func validateCapabilities(capabilities []Capability) error {
	seen := make(map[string]bool, len(capabilities))
	for _, capability := range capabilities {
		if !slices.Contains(allCapabilities, capability.TaskID) {
			return fmt.Errorf("unknown capability %q", capability.TaskID)
		}
		if seen[capability.TaskID] {
			return fmt.Errorf("duplicate capability %q", capability.TaskID)
		}
		seen[capability.TaskID] = true
		switch capability.ProfileState {
		case ProfileDisabled, ProfileFixtureOnly:
		case ProfileSandbox, ProfileProduction:
			return fmt.Errorf("capability %q may not activate as %s", capability.TaskID, capability.ProfileState)
		default:
			return fmt.Errorf("capability %q has unknown profile state %q", capability.TaskID, capability.ProfileState)
		}
		if capability.Registration || capability.Advertisement || capability.Readiness || capability.Mutation {
			return fmt.Errorf("capability %q may not register, advertise, become ready, or mutate", capability.TaskID)
		}
	}
	for _, taskID := range allCapabilities {
		if !seen[taskID] {
			return fmt.Errorf("capability %q is missing", taskID)
		}
	}
	return nil
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	seen := make(map[string]bool, len(left))
	for _, value := range left {
		if seen[value] || !slices.Contains(right, value) {
			return false
		}
		seen[value] = true
	}
	return true
}

func Parse(data []byte) (DependencyManifest, error) {
	var manifest DependencyManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return DependencyManifest{}, err
	}
	if err := manifest.Validate(); err != nil {
		return DependencyManifest{}, err
	}
	return manifest, nil
}

func (m DependencyManifest) CanonicalDigest() ([sha256.Size]byte, error) {
	if err := m.Validate(); err != nil {
		return [sha256.Size]byte{}, err
	}
	canonical, err := json.Marshal(m)
	if err != nil {
		return [sha256.Size]byte{}, err
	}
	return sha256.Sum256(canonical), nil
}
