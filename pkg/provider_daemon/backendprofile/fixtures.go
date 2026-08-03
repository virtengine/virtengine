package backendprofile

import "crypto/sha256"

func KubernetesFixture() (CloudProfile, DesiredResourceGraph) {
	profile := fixtureProfile(
		"fixture-kubernetes",
		BackendBinding{Backend: "kubernetes", Vendor: "kubernetes", APIVersion: "v1.30"},
		[]string{"containers", "namespaces", "networking"},
		"development",
		"cluster-fixture-1",
	)
	resources := []DesiredResource{
		fixtureResource(profile, "namespace-workload", KindKubernetesNamespace, []DesiredSpecField{
			{Key: "name", Value: "fixture-workload"},
		}, nil, ResourceBudget{CostMinorUnits: 100}),
		fixtureResource(profile, "deployment-api", KindKubernetesDeployment, []DesiredSpecField{
			{Key: "image", Value: "registry.invalid/virtengine/api@sha256:fixture"},
			{Key: "replicas", Value: "1"},
		}, []string{"namespace-workload"}, ResourceBudget{VCPUs: 2, MemoryMiB: 2048, StorageGiB: 10, CostMinorUnits: 2400}),
	}
	return profile, fixtureGraph(profile, "graph-kubernetes", resources, []string{"deployment-api", "namespace-workload"})
}

func OpenStackFixture() (CloudProfile, DesiredResourceGraph) {
	profile := fixtureProfile(
		"fixture-openstack",
		BackendBinding{Backend: "openstack", Vendor: "openstack", APIVersion: "2024.1"},
		[]string{"compute", "networks", "volumes"},
		"development",
		"project-fixture-1",
	)
	resources := []DesiredResource{
		fixtureResource(profile, "network-private", KindOpenStackNetwork, []DesiredSpecField{
			{Key: "cidr", Value: "192.0.2.0/24"},
			{Key: "name", Value: "fixture-private"},
		}, nil, ResourceBudget{CostMinorUnits: 500}),
		fixtureResource(profile, "server-api", KindOpenStackServer, []DesiredSpecField{
			{Key: "flavor", Value: "fixture.small"},
			{Key: "image", Value: "fixture-image-digest"},
			{Key: "name", Value: "fixture-api"},
		}, []string{"network-private"}, ResourceBudget{VCPUs: 2, MemoryMiB: 4096, StorageGiB: 40, CostMinorUnits: 4000}),
	}
	return profile, fixtureGraph(profile, "graph-openstack", resources, []string{"server-api", "network-private"})
}

func fixtureProfile(id string, backend BackendBinding, capabilities []string, environment, project string) CloudProfile {
	ownership := OwnershipTags{ProviderID: "provider-fixture", TenantID: "tenant-fixture", ProfileID: id}
	return CloudProfile{
		Version:          Version1,
		ID:               id,
		Backend:          backend,
		Environment:      environment,
		Region:           "region-fixture-1",
		Project:          project,
		Capabilities:     capabilities,
		Quotas:           Quotas{Resources: 8, VCPUs: 16, MemoryMiB: 32768, StorageGiB: 500},
		CostCeiling:      CostCeiling{MinorUnits: 10000, Currency: "USD"},
		AllowlistDigests: AllowlistDigests{Images: digest("images:" + id), Flavors: digest("flavors:" + id), SKUs: digest("skus:" + id)},
		Ownership:        ownership,
		Lineage: LineageTags{
			Source: "virtengine-t5-fixtures", Revision: "t5-05-v1", Digest: digest("lineage:" + id),
		},
		Certification:    StateFixtureOnly,
		Certified:        false,
		DependencyDigest: digest("dependency:90A:88D-unavailable"),
		ExternalBlocker:  "88D integration and backend certification are incomplete",
	}
}

func fixtureResource(profile CloudProfile, id string, kind ResourceKind, spec []DesiredSpecField, dependencies []string, budget ResourceBudget) DesiredResource {
	specDigest, err := DesiredSpecDigest(spec)
	if err != nil {
		panic(err)
	}
	return DesiredResource{
		ID:                           id,
		Kind:                         kind,
		DesiredSpec:                  spec,
		DesiredSpecDigest:            specDigest,
		Dependencies:                 dependencies,
		IdempotencyKey:               "create-" + id,
		Owner:                        profile.Ownership,
		AuthorityCommitment:          digest("authority:" + profile.ID + ":" + id),
		LineageDigest:                profile.Lineage.Digest,
		Budget:                       budget,
		DestructiveCleanupAuthorized: true,
	}
}

func fixtureGraph(profile CloudProfile, id string, resources []DesiredResource, cleanupOrder []string) DesiredResourceGraph {
	edges := make([]CleanupEdge, 0)
	expectations := make([]ResidueExpectation, 0, len(resources))
	for _, resource := range resources {
		for _, dependency := range resource.Dependencies {
			edges = append(edges, CleanupEdge{Before: resource.ID, After: dependency})
		}
		expectations = append(expectations, ResidueExpectation{ResourceID: resource.ID, VerificationRequired: true})
	}
	return DesiredResourceGraph{
		Version:              Version1,
		ID:                   id,
		ProfileID:            profile.ID,
		ProfileLineageDigest: profile.Lineage.Digest,
		Resources:            resources,
		CleanupOrder:         cleanupOrder,
		CleanupEdges:         edges,
		ResidueExpectations:  expectations,
	}
}

func digest(value string) [32]byte {
	return sha256.Sum256([]byte(value))
}
