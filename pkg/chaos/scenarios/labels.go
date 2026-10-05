// Copyright 2024-2025 VirtEngine Labs
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scenarios

// Chaos experiment label keys.
//
// These are the literal keys written into Experiment.Labels. They are consumed
// by the Litmus/Chaos Mesh manifests under infra/kubernetes/chaos/, which select
// on the same keys, so the string values here are a wire contract with the
// cluster: changing one changes which experiments a selector picks up.
//
// The `.com` / `.dev` split is PRE-EXISTING and deliberately NOT normalised
// here. The scenario builders in network.go, resource.go and byzantine.go have
// always emitted `chaos.virtengine.com/*`, while the node/pod-failure
// builders in node.go have always emitted `chaos.virtengine.dev/*`. Collapsing
// them to one prefix would silently retarget every live selector that was
// written against the other prefix -- a behaviour change, not a lint fix, and
// not something to smuggle in through a refactor. Unifying them is its own
// change: it has to move the Go builders and the infra/ selectors together.
// See ESTATE.md; label-selector changes are operational, not cosmetic.
const (
	// ChaosLabelType is the experiment type key (`.com` prefix family).
	ChaosLabelType = "chaos.virtengine.com/type"

	// ChaosLabelCategory is the category key (`.com` prefix family).
	ChaosLabelCategory = "chaos.virtengine.com/category"

	// ChaosLabelSeverity is the severity key (`.com` prefix family).
	ChaosLabelSeverity = "chaos.virtengine.com/severity"

	// ChaosLabelNodeCategory is the category key (`.dev` prefix family).
	ChaosLabelNodeCategory = "chaos.virtengine.dev/category"

	// ChaosLabelNodeSeverity is the severity key (`.dev` prefix family).
	ChaosLabelNodeSeverity = "chaos.virtengine.dev/severity"
)

// Category label values. The `.com` family keys are set to the category of the
// experiment they belong to.
const (
	// ChaosCategoryByzantine groups Byzantine-fault experiments.
	ChaosCategoryByzantine = "byzantine"

	// ChaosCategoryNetwork groups network-partition experiments.
	ChaosCategoryNetwork = "network"

	// ChaosCategoryResource groups resource-exhaustion experiments.
	ChaosCategoryResource = "resource"
)

// Category label values for the `.dev` (node/pod-failure) prefix family.
const (
	// ChaosCategoryValidator marks validator-node experiments.
	ChaosCategoryValidator = "validator"

	// ChaosCategoryProvider marks provider-node experiments.
	ChaosCategoryProvider = "provider"

	// ChaosCategoryRandom marks random-node-selection experiments.
	ChaosCategoryRandom = "random"

	// ChaosCategoryRolling marks rolling-restart experiments.
	ChaosCategoryRolling = "rolling"

	// ChaosCategoryNode marks generic node experiments.
	ChaosCategoryNode = "node"

	// ChaosCategoryContainer marks container-level experiments.
	ChaosCategoryContainer = "container"

	// ChaosCategoryCascade marks cascade-failure experiments.
	ChaosCategoryCascade = "cascade"
)

// Severity label values.
const (
	// ChaosSeverityCritical marks experiments that can halt consensus.
	ChaosSeverityCritical = "critical"

	// ChaosSeverityHigh marks experiments with broad blast radius.
	ChaosSeverityHigh = "high"

	// ChaosSeverityMedium marks experiments with bounded blast radius.
	ChaosSeverityMedium = "medium"

	// ChaosSeverityLow marks experiments with negligible blast radius.
	ChaosSeverityLow = "low"
)
