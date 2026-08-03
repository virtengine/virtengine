// Package backendprofile defines fixture-only cloud backend contracts.
package backendprofile

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
)

const Version1 uint32 = 1

type CertificationState string

const (
	StateDisabled    CertificationState = "disabled"
	StateFixtureOnly CertificationState = "fixture_only"
	StateSandbox     CertificationState = "sandbox"
	StateProduction  CertificationState = "production"
)

type BackendBinding struct {
	Backend    string
	Vendor     string
	APIVersion string
}

type Quotas struct {
	Resources  uint32
	VCPUs      uint32
	MemoryMiB  uint64
	StorageGiB uint64
}

type CostCeiling struct {
	MinorUnits uint64
	Currency   string
}

type AllowlistDigests struct {
	Images  [32]byte
	Flavors [32]byte
	SKUs    [32]byte
}

type OwnershipTags struct {
	ProviderID string
	TenantID   string
	ProfileID  string
}

type LineageTags struct {
	Source   string
	Revision string
	Digest   [32]byte
}

type CloudProfile struct {
	Version             uint32
	ID                  string
	Backend             BackendBinding
	Environment         string
	Region              string
	Project             string
	Capabilities        []string
	Quotas              Quotas
	CostCeiling         CostCeiling
	AllowlistDigests    AllowlistDigests
	Ownership           OwnershipTags
	Lineage             LineageTags
	Certification       CertificationState
	Certified           bool
	ActivationRequested bool
	DependencyDigest    [32]byte
	ExternalBlocker     string
}

type ResourceKind string

const (
	KindKubernetesNamespace  ResourceKind = "kubernetes.namespace"
	KindKubernetesDeployment ResourceKind = "kubernetes.deployment"
	KindOpenStackNetwork     ResourceKind = "openstack.network"
	KindOpenStackServer      ResourceKind = "openstack.server"
)

type ResourceBudget struct {
	VCPUs          uint32
	MemoryMiB      uint64
	StorageGiB     uint64
	CostMinorUnits uint64
}

type DesiredSpecField struct {
	Key   string
	Value string
}

type DesiredResource struct {
	ID                           string
	Kind                         ResourceKind
	DesiredSpec                  []DesiredSpecField
	DesiredSpecDigest            [32]byte
	Dependencies                 []string
	IdempotencyKey               string
	Owner                        OwnershipTags
	AuthorityCommitment          [32]byte
	LineageDigest                [32]byte
	Budget                       ResourceBudget
	External                     bool
	DestructiveCleanupAuthorized bool
}

type CleanupEdge struct {
	Before string
	After  string
}

type ResidueExpectation struct {
	ResourceID           string
	ExpectedResidueCount uint32
	VerificationRequired bool
}

type DesiredResourceGraph struct {
	Version              uint32
	ID                   string
	ProfileID            string
	ProfileLineageDigest [32]byte
	Resources            []DesiredResource
	CleanupOrder         []string
	CleanupEdges         []CleanupEdge
	ResidueExpectations  []ResidueExpectation
}

var stableID = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]{0,62}$`)

var supportedBackends = map[BackendBinding]map[string]struct{}{
	{Backend: "kubernetes", Vendor: "kubernetes", APIVersion: "v1.30"}: {
		"containers": {}, "namespaces": {}, "networking": {},
	},
	{Backend: "openstack", Vendor: "openstack", APIVersion: "2024.1"}: {
		"compute": {}, "networks": {}, "volumes": {},
	},
}

func (profile CloudProfile) Validate() error {
	if profile.Version != Version1 {
		return errors.New("unsupported cloud profile version")
	}
	supportedCapabilities, ok := supportedBackends[profile.Backend]
	if !ok {
		return errors.New("unsupported backend, vendor, or API version")
	}
	if !stableID.MatchString(profile.ID) || profile.Environment == "" || profile.Region == "" || profile.Project == "" {
		return errors.New("complete profile location binding is required")
	}
	if len(profile.Capabilities) == 0 {
		return errors.New("at least one capability is required")
	}
	seenCapabilities := make(map[string]struct{}, len(profile.Capabilities))
	for _, capability := range profile.Capabilities {
		if _, supported := supportedCapabilities[capability]; !supported {
			return fmt.Errorf("unsupported capability %q", capability)
		}
		if _, duplicate := seenCapabilities[capability]; duplicate {
			return fmt.Errorf("duplicate capability %q", capability)
		}
		seenCapabilities[capability] = struct{}{}
	}
	if profile.Quotas.Resources == 0 || profile.Quotas.VCPUs == 0 || profile.Quotas.MemoryMiB == 0 || profile.Quotas.StorageGiB == 0 {
		return errors.New("positive quotas are required")
	}
	if profile.CostCeiling.MinorUnits == 0 || !validCurrency(profile.CostCeiling.Currency) {
		return errors.New("positive cost ceiling and ISO-style currency are required")
	}
	if zeroDigest(profile.AllowlistDigests.Images) || zeroDigest(profile.AllowlistDigests.Flavors) || zeroDigest(profile.AllowlistDigests.SKUs) {
		return errors.New("immutable image, flavor, and SKU allowlist digests are required")
	}
	if !validOwnership(profile.Ownership) || profile.Ownership.ProfileID != profile.ID {
		return errors.New("complete profile ownership tags are required")
	}
	if profile.Lineage.Source == "" || profile.Lineage.Revision == "" || zeroDigest(profile.Lineage.Digest) {
		return errors.New("immutable lineage tags are required")
	}
	if zeroDigest(profile.DependencyDigest) || profile.ExternalBlocker == "" {
		return errors.New("dependency digest and external blocker are required")
	}
	if profile.Certified {
		return errors.New("fixture profile must remain explicitly uncertified")
	}
	switch profile.Certification {
	case StateDisabled, StateFixtureOnly:
	case StateSandbox, StateProduction:
		return errors.New("cloud profile may not exceed fixture_only")
	default:
		return errors.New("unknown certification state")
	}
	if profile.ActivationRequested {
		return errors.New("uncertified cloud profile cannot be activated")
	}
	return nil
}

func ValidateProfileTransition(current, next CloudProfile) error {
	if err := current.Validate(); err != nil {
		return fmt.Errorf("current profile: %w", err)
	}
	if err := next.Validate(); err != nil {
		return fmt.Errorf("next profile: %w", err)
	}
	if current.ID != next.ID || current.Backend != next.Backend || current.Environment != next.Environment || current.Region != next.Region || current.Project != next.Project {
		return errors.New("profile binding is immutable")
	}
	if current.AllowlistDigests != next.AllowlistDigests || current.Ownership != next.Ownership || current.Lineage != next.Lineage || current.DependencyDigest != next.DependencyDigest {
		return errors.New("profile allowlists, ownership, lineage, and dependency are immutable")
	}
	currentRank := certificationRank(current.Certification)
	nextRank := certificationRank(next.Certification)
	if nextRank < currentRank || nextRank-currentRank > 1 {
		return errors.New("invalid certification state transition")
	}
	return nil
}

func Validate(profile CloudProfile, graph DesiredResourceGraph) error {
	if err := profile.Validate(); err != nil {
		return fmt.Errorf("profile: %w", err)
	}
	if graph.Version != Version1 {
		return errors.New("unsupported desired resource graph version")
	}
	if !stableID.MatchString(graph.ID) || graph.ProfileID != profile.ID || graph.ProfileLineageDigest != profile.Lineage.Digest {
		return errors.New("graph profile binding or immutable lineage mismatch")
	}
	if len(graph.Resources) == 0 {
		return errors.New("resource graph is empty")
	}
	resources := make(map[string]DesiredResource, len(graph.Resources))
	idempotencyKeys := make(map[string]string, len(graph.Resources))
	var used ResourceBudget
	for _, resource := range graph.Resources {
		if err := validateResource(profile, resource); err != nil {
			return err
		}
		if _, duplicate := resources[resource.ID]; duplicate {
			return fmt.Errorf("conflicting duplicate resource ID %q", resource.ID)
		}
		if prior, duplicate := idempotencyKeys[resource.IdempotencyKey]; duplicate {
			return fmt.Errorf("conflicting idempotency key %q for %q and %q", resource.IdempotencyKey, prior, resource.ID)
		}
		resources[resource.ID] = resource
		idempotencyKeys[resource.IdempotencyKey] = resource.ID
		if resource.Budget.VCPUs > math.MaxUint32-used.VCPUs || resource.Budget.MemoryMiB > math.MaxUint64-used.MemoryMiB || resource.Budget.StorageGiB > math.MaxUint64-used.StorageGiB || resource.Budget.CostMinorUnits > math.MaxUint64-used.CostMinorUnits {
			return errors.New("resource quota or cost aggregation overflow")
		}
		used.VCPUs += resource.Budget.VCPUs
		used.MemoryMiB += resource.Budget.MemoryMiB
		used.StorageGiB += resource.Budget.StorageGiB
		used.CostMinorUnits += resource.Budget.CostMinorUnits
	}
	if uint32(len(resources)) > profile.Quotas.Resources || used.VCPUs > profile.Quotas.VCPUs || used.MemoryMiB > profile.Quotas.MemoryMiB || used.StorageGiB > profile.Quotas.StorageGiB {
		return errors.New("resource graph exceeds profile quota")
	}
	if used.CostMinorUnits > profile.CostCeiling.MinorUnits {
		return errors.New("resource graph exceeds profile cost ceiling")
	}
	if err := validateDependencies(resources); err != nil {
		return err
	}
	if err := validateCleanup(resources, graph.CleanupOrder, graph.CleanupEdges); err != nil {
		return err
	}
	return validateResidue(resources, graph.ResidueExpectations)
}

func validateResource(profile CloudProfile, resource DesiredResource) error {
	if !stableID.MatchString(resource.ID) || !stableID.MatchString(resource.IdempotencyKey) || zeroDigest(resource.DesiredSpecDigest) || zeroDigest(resource.AuthorityCommitment) || zeroDigest(resource.LineageDigest) {
		return fmt.Errorf("resource %q has incomplete stable identity or commitments", resource.ID)
	}
	digest, err := DesiredSpecDigest(resource.DesiredSpec)
	if err != nil || digest != resource.DesiredSpecDigest {
		return fmt.Errorf("resource %q desired spec digest mismatch", resource.ID)
	}
	if resource.LineageDigest != profile.Lineage.Digest {
		return fmt.Errorf("resource %q has mutable lineage", resource.ID)
	}
	if resource.External || !validOwnership(resource.Owner) || resource.Owner != profile.Ownership {
		return fmt.Errorf("resource %q is external or unowned", resource.ID)
	}
	if !resource.DestructiveCleanupAuthorized {
		return fmt.Errorf("resource %q lacks destructive cleanup authorization", resource.ID)
	}
	if !kindSupported(profile.Backend.Backend, resource.Kind) {
		return fmt.Errorf("resource %q has unsupported kind %q", resource.ID, resource.Kind)
	}
	return nil
}

func validateDependencies(resources map[string]DesiredResource) error {
	visiting := make(map[string]bool, len(resources))
	visited := make(map[string]bool, len(resources))
	var visit func(string) error
	visit = func(id string) error {
		if visiting[id] {
			return errors.New("resource dependency cycle")
		}
		if visited[id] {
			return nil
		}
		visiting[id] = true
		for _, dependency := range resources[id].Dependencies {
			if _, exists := resources[dependency]; !exists {
				return fmt.Errorf("resource %q has missing dependency %q", id, dependency)
			}
			if err := visit(dependency); err != nil {
				return err
			}
		}
		visiting[id] = false
		visited[id] = true
		return nil
	}
	for id := range resources {
		if err := visit(id); err != nil {
			return err
		}
	}
	return nil
}

func validateCleanup(resources map[string]DesiredResource, order []string, edges []CleanupEdge) error {
	if len(order) != len(resources) {
		return errors.New("cleanup order must cover every resource exactly once")
	}
	positions := make(map[string]int, len(order))
	for index, id := range order {
		if _, exists := resources[id]; !exists {
			return fmt.Errorf("cleanup order references unknown resource %q", id)
		}
		if _, duplicate := positions[id]; duplicate {
			return fmt.Errorf("duplicate cleanup order entry %q", id)
		}
		positions[id] = index
	}
	requiredEdges := make(map[CleanupEdge]struct{})
	for id, resource := range resources {
		for _, dependency := range resource.Dependencies {
			if positions[id] >= positions[dependency] {
				return fmt.Errorf("cleanup order does not reverse dependency %q -> %q", id, dependency)
			}
			requiredEdges[CleanupEdge{Before: id, After: dependency}] = struct{}{}
		}
	}
	if len(edges) != len(requiredEdges) {
		return errors.New("cleanup edges do not exactly reverse dependencies")
	}
	seen := make(map[CleanupEdge]struct{}, len(edges))
	for _, edge := range edges {
		if _, required := requiredEdges[edge]; !required {
			return fmt.Errorf("invalid cleanup edge %q -> %q", edge.Before, edge.After)
		}
		if _, duplicate := seen[edge]; duplicate {
			return fmt.Errorf("duplicate cleanup edge %q -> %q", edge.Before, edge.After)
		}
		seen[edge] = struct{}{}
	}
	return nil
}

func validateResidue(resources map[string]DesiredResource, expectations []ResidueExpectation) error {
	if len(expectations) != len(resources) {
		return errors.New("residue expectations must cover every resource exactly once")
	}
	seen := make(map[string]struct{}, len(expectations))
	for _, expectation := range expectations {
		if _, exists := resources[expectation.ResourceID]; !exists {
			return fmt.Errorf("residue expectation references unknown resource %q", expectation.ResourceID)
		}
		if _, duplicate := seen[expectation.ResourceID]; duplicate {
			return fmt.Errorf("duplicate residue expectation for %q", expectation.ResourceID)
		}
		if expectation.ExpectedResidueCount != 0 || !expectation.VerificationRequired {
			return fmt.Errorf("resource %q lacks verified zero-residue expectation", expectation.ResourceID)
		}
		seen[expectation.ResourceID] = struct{}{}
	}
	return nil
}

func kindSupported(backend string, kind ResourceKind) bool {
	switch backend {
	case "kubernetes":
		return kind == KindKubernetesNamespace || kind == KindKubernetesDeployment
	case "openstack":
		return kind == KindOpenStackNetwork || kind == KindOpenStackServer
	default:
		return false
	}
}

func validOwnership(ownership OwnershipTags) bool {
	return stableID.MatchString(ownership.ProviderID) && stableID.MatchString(ownership.TenantID) && stableID.MatchString(ownership.ProfileID)
}

func validCurrency(currency string) bool {
	return len(currency) == 3 && currency[0] >= 'A' && currency[0] <= 'Z' && currency[1] >= 'A' && currency[1] <= 'Z' && currency[2] >= 'A' && currency[2] <= 'Z'
}

func certificationRank(state CertificationState) int {
	switch state {
	case StateDisabled:
		return 0
	case StateFixtureOnly:
		return 1
	case StateSandbox:
		return 2
	case StateProduction:
		return 3
	default:
		return -1
	}
}

func DesiredSpecDigest(spec []DesiredSpecField) ([32]byte, error) {
	if len(spec) == 0 {
		return [32]byte{}, errors.New("desired spec is empty")
	}
	for index, field := range spec {
		if !stableID.MatchString(field.Key) || field.Value == "" {
			return [32]byte{}, errors.New("desired spec contains an invalid field")
		}
		if index > 0 && spec[index-1].Key >= field.Key {
			return [32]byte{}, errors.New("desired spec fields must have unique canonical order")
		}
	}
	encoded, err := json.Marshal(spec)
	if err != nil {
		return [32]byte{}, fmt.Errorf("encode desired spec: %w", err)
	}
	return sha256.Sum256(encoded), nil
}

func zeroDigest(digest [32]byte) bool {
	return digest == [32]byte{}
}
