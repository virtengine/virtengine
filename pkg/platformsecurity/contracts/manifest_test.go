package contracts

import (
	"encoding/hex"
	"os"
	"testing"
)

func loadFixture(t *testing.T) DependencyManifest {
	t.Helper()
	data, err := os.ReadFile("testdata/dependency-manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	return manifest
}

func cloneFixture(t *testing.T) DependencyManifest {
	t.Helper()
	manifest := loadFixture(t)
	manifest.Dependencies = append([]Dependency(nil), manifest.Dependencies...)
	manifest.Capabilities = append([]Capability(nil), manifest.Capabilities...)
	return manifest
}

func TestFixtureCanonicalDigest(t *testing.T) {
	manifest := loadFixture(t)
	digest, err := manifest.CanonicalDigest()
	if err != nil {
		t.Fatal(err)
	}
	const expected = "ace90f4423f4bdeaac92508a5b00e6d2b27c769e00f43b787ad4e8ae5565c527"
	if actual := hex.EncodeToString(digest[:]); actual != expected {
		t.Fatalf("canonical digest = %s, want %s", actual, expected)
	}
}

func TestEveryCapabilityRejectsNegativeEnablement(t *testing.T) {
	mutations := []struct {
		name  string
		apply func(*Capability)
	}{
		{"registration", func(capability *Capability) { capability.Registration = true }},
		{"advertisement", func(capability *Capability) { capability.Advertisement = true }},
		{"readiness", func(capability *Capability) { capability.Readiness = true }},
		{"mutation", func(capability *Capability) { capability.Mutation = true }},
		{"sandbox", func(capability *Capability) { capability.ProfileState = ProfileSandbox }},
		{"production", func(capability *Capability) { capability.ProfileState = ProfileProduction }},
	}
	for capabilityIndex, taskID := range allCapabilities {
		for _, mutation := range mutations {
			t.Run(taskID+"/"+mutation.name, func(t *testing.T) {
				manifest := cloneFixture(t)
				mutation.apply(&manifest.Capabilities[capabilityIndex])
				if err := manifest.Validate(); err == nil {
					t.Fatal("activation accepted")
				}
			})
		}
	}
}

func TestMissingThreadDependencyFailsClosed(t *testing.T) {
	for _, thread := range []string{"T1", "T3", "T4"} {
		t.Run(thread, func(t *testing.T) {
			manifest := cloneFixture(t)
			for index, dependency := range manifest.Dependencies {
				if dependency.SourceThread == thread {
					manifest.Dependencies = append(manifest.Dependencies[:index], manifest.Dependencies[index+1:]...)
					break
				}
			}
			if err := manifest.Validate(); err == nil {
				t.Fatal("missing dependency accepted")
			}
		})
	}
}

func TestChangedPinnedDigestFailsClosed(t *testing.T) {
	manifest := cloneFixture(t)
	for index := range manifest.Dependencies {
		if manifest.Dependencies[index].ContractID == "integration-control" {
			manifest.Dependencies[index].Digest = "0" + manifest.Dependencies[index].Digest[1:]
		}
	}
	if err := manifest.Validate(); err == nil {
		t.Fatal("changed T4 digest accepted")
	}
}

func TestDuplicateAndUnknownVersionsFail(t *testing.T) {
	t.Run("duplicate dependency", func(t *testing.T) {
		manifest := cloneFixture(t)
		manifest.Dependencies = append(manifest.Dependencies, manifest.Dependencies[0])
		if err := manifest.Validate(); err == nil {
			t.Fatal("duplicate dependency accepted")
		}
	})
	t.Run("duplicate capability", func(t *testing.T) {
		manifest := cloneFixture(t)
		manifest.Capabilities = append(manifest.Capabilities, manifest.Capabilities[0])
		if err := manifest.Validate(); err == nil {
			t.Fatal("duplicate capability accepted")
		}
	})
	t.Run("manifest version", func(t *testing.T) {
		manifest := cloneFixture(t)
		manifest.Version = "v2"
		if err := manifest.Validate(); err == nil {
			t.Fatal("unknown manifest version accepted")
		}
	})
	t.Run("dependency version", func(t *testing.T) {
		manifest := cloneFixture(t)
		manifest.Dependencies[0].Version = "v2"
		if err := manifest.Validate(); err == nil {
			t.Fatal("unknown dependency version accepted")
		}
	})
}

func TestMalformedDependencyRowsFail(t *testing.T) {
	tests := []struct {
		name  string
		apply func(*Dependency)
	}{
		{"unknown status", func(dependency *Dependency) { dependency.Status = "ready" }},
		{"absent blocker", func(dependency *Dependency) { dependency.Blocker = "" }},
		{"unavailable digest", func(dependency *Dependency) { dependency.Digest = T4ControlDigest }},
		{"missing consumer", func(dependency *Dependency) { dependency.Consumers = dependency.Consumers[1:] }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manifest := cloneFixture(t)
			test.apply(&manifest.Dependencies[0])
			if err := manifest.Validate(); err == nil {
				t.Fatal("malformed dependency accepted")
			}
		})
	}
}

func TestCapabilityOmissionAndUnknownStateFail(t *testing.T) {
	t.Run("omission", func(t *testing.T) {
		manifest := cloneFixture(t)
		manifest.Capabilities = manifest.Capabilities[1:]
		if err := manifest.Validate(); err == nil {
			t.Fatal("capability omission accepted")
		}
	})
	t.Run("unknown state", func(t *testing.T) {
		manifest := cloneFixture(t)
		manifest.Capabilities[0].ProfileState = "ready"
		if err := manifest.Validate(); err == nil {
			t.Fatal("unknown capability state accepted")
		}
	})
}
