package backendprofile

import (
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestFixturesAreDeterministicAndFixtureOnly(t *testing.T) {
	constructors := []struct {
		name string
		new  func() (CloudProfile, DesiredResourceGraph)
	}{
		{name: "kubernetes", new: KubernetesFixture},
		{name: "openstack", new: OpenStackFixture},
	}
	for _, constructor := range constructors {
		t.Run(constructor.name, func(t *testing.T) {
			firstProfile, firstGraph := constructor.new()
			secondProfile, secondGraph := constructor.new()
			if !reflect.DeepEqual(firstProfile, secondProfile) || !reflect.DeepEqual(firstGraph, secondGraph) {
				t.Fatal("fixture constructor is not deterministic")
			}
			if firstProfile.Certification != StateFixtureOnly || firstProfile.Certified || firstProfile.ActivationRequested {
				t.Fatal("fixture is not explicitly uncertified and fixture_only")
			}
			if err := Validate(firstProfile, firstGraph); err != nil {
				t.Fatalf("valid fixture rejected: %v", err)
			}
		})
	}
}

func TestProfileRejectsUnsupportedAndUncertifiedActivation(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*CloudProfile)
	}{
		{name: "contract version", mutate: func(profile *CloudProfile) { profile.Version++ }},
		{name: "backend", mutate: func(profile *CloudProfile) { profile.Backend.Backend = "unknown" }},
		{name: "vendor", mutate: func(profile *CloudProfile) { profile.Backend.Vendor = "other" }},
		{name: "API version", mutate: func(profile *CloudProfile) { profile.Backend.APIVersion = "latest" }},
		{name: "capability", mutate: func(profile *CloudProfile) { profile.Capabilities = append(profile.Capabilities, "privileged") }},
		{name: "sandbox state", mutate: func(profile *CloudProfile) { profile.Certification = StateSandbox }},
		{name: "production state", mutate: func(profile *CloudProfile) { profile.Certification = StateProduction }},
		{name: "certified fixture", mutate: func(profile *CloudProfile) { profile.Certified = true }},
		{name: "activation", mutate: func(profile *CloudProfile) { profile.ActivationRequested = true }},
		{name: "dependency", mutate: func(profile *CloudProfile) { profile.DependencyDigest = [32]byte{} }},
		{name: "external blocker", mutate: func(profile *CloudProfile) { profile.ExternalBlocker = "" }},
		{name: "allowlist", mutate: func(profile *CloudProfile) { profile.AllowlistDigests.Images = [32]byte{} }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			profile, graph := KubernetesFixture()
			test.mutate(&profile)
			if err := Validate(profile, graph); err == nil {
				t.Fatal("invalid profile accepted")
			}
		})
	}
}

func TestProfileTransitionPreservesLineageAndCertificationCap(t *testing.T) {
	current, _ := KubernetesFixture()
	current.Certification = StateDisabled
	next := current
	next.Certification = StateFixtureOnly
	if err := ValidateProfileTransition(current, next); err != nil {
		t.Fatalf("disabled to fixture_only transition rejected: %v", err)
	}
	mutated := next
	mutated.Lineage.Digest[0] ^= 0xff
	if err := ValidateProfileTransition(next, mutated); err == nil {
		t.Fatal("mutable profile lineage accepted")
	}
	downgrade := next
	downgrade.Certification = StateDisabled
	if err := ValidateProfileTransition(next, downgrade); err == nil {
		t.Fatal("certification downgrade accepted")
	}
	sandbox := next
	sandbox.Certification = StateSandbox
	if err := ValidateProfileTransition(next, sandbox); err == nil {
		t.Fatal("transition beyond fixture_only accepted")
	}
}

func TestGraphRejectsIdentityTopologyAndBudgetFailures(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*CloudProfile, *DesiredResourceGraph)
	}{
		{name: "graph version", mutate: func(_ *CloudProfile, graph *DesiredResourceGraph) { graph.Version++ }},
		{name: "duplicate ID", mutate: func(_ *CloudProfile, graph *DesiredResourceGraph) {
			graph.Resources = append(graph.Resources, graph.Resources[0])
		}},
		{name: "duplicate idempotency key", mutate: func(_ *CloudProfile, graph *DesiredResourceGraph) {
			graph.Resources[1].IdempotencyKey = graph.Resources[0].IdempotencyKey
		}},
		{name: "missing dependency", mutate: func(_ *CloudProfile, graph *DesiredResourceGraph) {
			graph.Resources[1].Dependencies = []string{"absent"}
		}},
		{name: "cycle", mutate: func(_ *CloudProfile, graph *DesiredResourceGraph) {
			graph.Resources[0].Dependencies = []string{graph.Resources[1].ID}
		}},
		{name: "resource quota", mutate: func(profile *CloudProfile, _ *DesiredResourceGraph) { profile.Quotas.Resources = 1 }},
		{name: "compute quota", mutate: func(profile *CloudProfile, graph *DesiredResourceGraph) {
			graph.Resources[1].Budget.VCPUs = profile.Quotas.VCPUs + 1
		}},
		{name: "cost ceiling", mutate: func(profile *CloudProfile, graph *DesiredResourceGraph) {
			graph.Resources[1].Budget.CostMinorUnits = profile.CostCeiling.MinorUnits + 1
		}},
		{name: "cost overflow", mutate: func(profile *CloudProfile, graph *DesiredResourceGraph) {
			profile.CostCeiling.MinorUnits = math.MaxUint64
			graph.Resources[0].Budget.CostMinorUnits = math.MaxUint64
			graph.Resources[1].Budget.CostMinorUnits = 1
		}},
		{name: "spec digest", mutate: func(_ *CloudProfile, graph *DesiredResourceGraph) {
			graph.Resources[1].DesiredSpec[0].Value = "mutable"
		}},
		{name: "profile lineage", mutate: func(_ *CloudProfile, graph *DesiredResourceGraph) { graph.ProfileLineageDigest[0] ^= 0xff }},
		{name: "resource lineage", mutate: func(_ *CloudProfile, graph *DesiredResourceGraph) { graph.Resources[0].LineageDigest[0] ^= 0xff }},
		{name: "unsupported kind", mutate: func(_ *CloudProfile, graph *DesiredResourceGraph) { graph.Resources[0].Kind = KindOpenStackNetwork }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			profile, graph := KubernetesFixture()
			test.mutate(&profile, &graph)
			if err := Validate(profile, graph); err == nil {
				t.Fatal("invalid resource graph accepted")
			}
		})
	}
}

func TestGraphRejectsUnsafeCleanupAndResidue(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*DesiredResourceGraph)
	}{
		{name: "external resource", mutate: func(graph *DesiredResourceGraph) { graph.Resources[0].External = true }},
		{name: "unowned resource", mutate: func(graph *DesiredResourceGraph) { graph.Resources[0].Owner.TenantID = "other-tenant" }},
		{name: "missing authority", mutate: func(graph *DesiredResourceGraph) { graph.Resources[0].AuthorityCommitment = [32]byte{} }},
		{name: "unauthorized destructive cleanup", mutate: func(graph *DesiredResourceGraph) { graph.Resources[0].DestructiveCleanupAuthorized = false }},
		{name: "dependency order", mutate: func(graph *DesiredResourceGraph) {
			graph.CleanupOrder[0], graph.CleanupOrder[1] = graph.CleanupOrder[1], graph.CleanupOrder[0]
		}},
		{name: "cleanup edge", mutate: func(graph *DesiredResourceGraph) {
			graph.CleanupEdges[0] = CleanupEdge{Before: "namespace-workload", After: "deployment-api"}
		}},
		{name: "incomplete cleanup", mutate: func(graph *DesiredResourceGraph) { graph.CleanupOrder = graph.CleanupOrder[:1] }},
		{name: "incomplete residue", mutate: func(graph *DesiredResourceGraph) { graph.ResidueExpectations = graph.ResidueExpectations[:1] }},
		{name: "nonzero residue", mutate: func(graph *DesiredResourceGraph) { graph.ResidueExpectations[0].ExpectedResidueCount = 1 }},
		{name: "unverified residue", mutate: func(graph *DesiredResourceGraph) { graph.ResidueExpectations[0].VerificationRequired = false }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			profile, graph := KubernetesFixture()
			test.mutate(&graph)
			if err := Validate(profile, graph); err == nil {
				t.Fatal("ownership-unsafe cleanup or incomplete residue accepted")
			}
		})
	}
}

func TestDesiredSpecRequiresCanonicalImmutableFields(t *testing.T) {
	if _, err := DesiredSpecDigest(nil); err == nil {
		t.Fatal("empty desired spec accepted")
	}
	if _, err := DesiredSpecDigest([]DesiredSpecField{{Key: "z", Value: "1"}, {Key: "a", Value: "2"}}); err == nil {
		t.Fatal("noncanonical desired spec accepted")
	}
	if _, err := DesiredSpecDigest([]DesiredSpecField{{Key: "name", Value: strings.Repeat("x", 2)}}); err != nil {
		t.Fatalf("valid desired spec rejected: %v", err)
	}
}
