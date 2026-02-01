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

import "time"

// ExperimentType defines the category of chaos experiment.
type ExperimentType string

// Experiment represents a complete chaos experiment configuration
// that can be executed by a chaos runner.
type Experiment struct {
	// Name is the unique identifier for this experiment.
	Name string `json:"name"`

	// Description provides human-readable details about the experiment.
	Description string `json:"description"`

	// Type categorizes the experiment.
	Type ExperimentType `json:"type"`

	// Duration specifies how long the experiment should run.
	Duration time.Duration `json:"duration"`

	// Targets lists the nodes or endpoints affected by this experiment.
	Targets []string `json:"targets"`

	// Spec contains the experiment-specific configuration.
	// Deprecated: Use Parameters for new scenarios.
	Spec interface{} `json:"spec,omitempty"`

	// Parameters contains experiment-specific configuration as key-value pairs.
	Parameters map[string]interface{} `json:"parameters,omitempty"`

	// Labels for categorization and filtering.
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations for additional metadata.
	Annotations map[string]string `json:"annotations,omitempty"`
}

