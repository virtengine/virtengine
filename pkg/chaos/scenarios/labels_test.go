// Behaviour-equivalence guard for the chaos label constant refactor.
//
// The refactor replaces repeated string literals in Experiment label maps with
// named constants. That is only legitimate if the constants hold the SAME
// strings and the maps come out byte-identical. This test pins both:
//
//   - the constant values (so nobody can later "tidy" a value and silently
//     retarget a cluster label selector), and
//   - the emitted label maps for a representative builder per family, proving
//     the label-selector wire contract the infra/ manifests select on.
package scenarios

import (
	"reflect"
	"testing"
	"time"
)

func TestChaosLabelKeyConstants(t *testing.T) {
	// These keys are a wire contract with infra/kubernetes/chaos/, whose Litmus
	// and Chaos Mesh manifests select on the same strings. The .com/.dev split
	// is deliberate and pre-existing; normalising it would retarget every live
	// selector, so it is pinned here instead.
	want := map[string]string{
		"type":         ChaosLabelType,
		"category":     ChaosLabelCategory,
		"severity":     ChaosLabelSeverity,
		"nodeCategory": ChaosLabelNodeCategory,
		"nodeSeverity": ChaosLabelNodeSeverity,
	}
	expected := map[string]string{
		"type":         "chaos.virtengine.com/type",
		"category":     "chaos.virtengine.com/category",
		"severity":     "chaos.virtengine.com/severity",
		"nodeCategory": "chaos.virtengine.dev/category",
		"nodeSeverity": "chaos.virtengine.dev/severity",
	}
	if !reflect.DeepEqual(want, expected) {
		t.Fatalf("chaos label keys changed:\n got: %v\nwant: %v", want, expected)
	}
}

func TestChaosCategoryAndSeverityConstants(t *testing.T) {
	pairs := []struct {
		name string
		got  string
		want string
	}{
		{"category/byzantine", ChaosCategoryByzantine, "byzantine"},
		{"category/network", ChaosCategoryNetwork, "network"},
		{"category/resource", ChaosCategoryResource, "resource"},
		{"category/validator", ChaosCategoryValidator, "validator"},
		{"category/provider", ChaosCategoryProvider, "provider"},
		{"category/random", ChaosCategoryRandom, "random"},
		{"category/rolling", ChaosCategoryRolling, "rolling"},
		{"category/node", ChaosCategoryNode, "node"},
		{"category/container", ChaosCategoryContainer, "container"},
		{"category/cascade", ChaosCategoryCascade, "cascade"},
		{"severity/critical", ChaosSeverityCritical, "critical"},
		{"severity/high", ChaosSeverityHigh, "high"},
		{"severity/medium", ChaosSeverityMedium, "medium"},
		{"severity/low", ChaosSeverityLow, "low"},
	}
	for _, p := range pairs {
		if p.got != p.want {
			t.Errorf("%s = %q, want %q", p.name, p.got, p.want)
		}
	}
}

// TestEmittedLabelsUnchanged pins the actual maps the builders emit, per
// prefix family, so a constant cannot drift from what the cluster selects on.
func TestEmittedLabelsUnchanged(t *testing.T) {
	// Duration must be positive (Validate rejects 0), so use a real one.
	byz, err := NewDoubleSigningScenario([]string{"validator-0"}, time.Minute).Build()
	if err != nil {
		t.Fatalf("DoubleSigningScenario.Build: %v", err)
	}
	wantByz := map[string]string{
		"chaos.virtengine.com/type":     "byzantine-double-signing",
		"chaos.virtengine.com/category": "byzantine",
		"chaos.virtengine.com/severity": "critical",
	}
	if !reflect.DeepEqual(byz.Labels, wantByz) {
		t.Errorf("byzantine labels = %v, want %v", byz.Labels, wantByz)
	}

	// Node/pod-failure family: different prefix, same shape.
	node := NewNodeReboot("node-1")
	wantNode := map[string]string{
		"chaos.virtengine.dev/category": "node",
		"chaos.virtengine.dev/severity": "critical",
	}
	if !reflect.DeepEqual(node.Labels, wantNode) {
		t.Errorf("node labels = %v, want %v", node.Labels, wantNode)
	}
}
