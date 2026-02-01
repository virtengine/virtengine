// Copyright 2024 VirtEngine Authors
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

// Package scenarios provides chaos engineering scenarios for testing VirtEngine
// resilience against various failure modes including node, pod, and container failures.
//
// This package implements node/pod failure scenarios for chaos testing:
//   - Pod failures (kill, delete, pause)
//   - Node failures (drain, cordon, reboot, shutdown)
//   - Container failures (kill, pause, restart)
//   - Cascade failures (multi-stage failure propagation)
//
// These scenarios are designed to test VirtEngine's resilience in production-like
// environments, particularly for validator nodes and provider daemons.
package scenarios

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Node/pod failure experiment types.
const (
	// ExperimentTypePodFailure represents pod failure experiments.
	ExperimentTypePodFailure ExperimentType = "pod-failure"

	// ExperimentTypeNodeFailure represents node failure experiments.
	ExperimentTypeNodeFailure ExperimentType = "node-failure"

	// ExperimentTypeContainerFailure represents container failure experiments.
	ExperimentTypeContainerFailure ExperimentType = "container-failure"

	// ExperimentTypeCascadeFailure represents cascade failure experiments.
	ExperimentTypeCascadeFailure ExperimentType = "cascade-failure"
)

// Pod failure actions.
const (
	// PodActionKill terminates a pod immediately.
	PodActionKill = "kill"

	// PodActionDelete deletes a pod gracefully.
	PodActionDelete = "delete"

	// PodActionPause pauses a pod (freezes processes).
	PodActionPause = "pause"
)

// Node failure actions.
const (
	// NodeActionDrain drains a node of all pods.
	NodeActionDrain = "drain"

	// NodeActionCordon marks a node as unschedulable.
	NodeActionCordon = "cordon"

	// NodeActionReboot reboots a node.
	NodeActionReboot = "reboot"

	// NodeActionShutdown shuts down a node.
	NodeActionShutdown = "shutdown"
)

// Container failure actions.
const (
	// ContainerActionKill terminates a container.
	ContainerActionKill = "kill"

	// ContainerActionPause pauses a container.
	ContainerActionPause = "pause"

	// ContainerActionRestart restarts a container.
	ContainerActionRestart = "restart"
)

// Default values for scenarios.
const (
	// DefaultGracePeriod is the default grace period for pod deletion.
	DefaultGracePeriod = 30 * time.Second

	// DefaultDuration is the default duration for temporary failures.
	DefaultDuration = 5 * time.Minute

	// DefaultInterval is the default interval between repeated failures.
	DefaultInterval = 1 * time.Minute

	// DefaultEvictGracePeriod is the default eviction grace period for node drains.
	DefaultEvictGracePeriod = 60 * time.Second
)

// Validation errors.
var (
	// ErrEmptyName indicates an empty scenario name.
	ErrEmptyName = errors.New("scenario name cannot be empty")

	// ErrNoTargets indicates no targets specified.
	ErrNoTargets = errors.New("at least one target must be specified")

	// ErrInvalidAction indicates an invalid action.
	ErrInvalidAction = errors.New("invalid action specified")

	// ErrInvalidCount indicates an invalid count.
	ErrInvalidCount = errors.New("count must be greater than zero")

	// ErrInvalidDuration indicates an invalid duration.
	ErrInvalidDuration = errors.New("duration must be positive")

	// ErrEmptyNodeName indicates an empty node name.
	ErrEmptyNodeName = errors.New("node name cannot be empty")

	// ErrEmptyPodName indicates an empty pod name.
	ErrEmptyPodName = errors.New("pod name cannot be empty")

	// ErrEmptyContainerName indicates an empty container name.
	ErrEmptyContainerName = errors.New("container name cannot be empty")

	// ErrEmptyNamespace indicates an empty namespace.
	ErrEmptyNamespace = errors.New("namespace cannot be empty")

	// ErrNoStages indicates no stages specified for cascade failure.
	ErrNoStages = errors.New("at least one stage must be specified")

	// ErrInvalidInterval indicates an invalid interval.
	ErrInvalidInterval = errors.New("interval must be positive")

	// ErrNilScenario indicates a nil scenario.
	ErrNilScenario = errors.New("scenario cannot be nil")

	// ErrCircularDependency indicates a circular dependency in stages.
	ErrCircularDependency = errors.New("circular dependency detected in stages")
)

// NodeScenariodefines the interface for all node/pod failure scenarios.
// All scenario types must implement this interface to be executable
// by the chaos runner.
type NodeScenario interface {
	// Name returns the unique identifier for this scenario.
	Name() string

	// Description returns a human-readable description of what this scenario does.
	Description() string

	// Type returns the experiment type for categorization.
	Type() ExperimentType

	// Build constructs an Experiment from this scenario configuration.
	// Returns an error if the scenario configuration is invalid.
	Build() (*Experiment, error)

	// Validate checks if the scenario configuration is valid.
	// Returns nil if valid, or an error describing the validation failure.
	Validate() error
}

// PodFailureScenario defines a pod failure chaos experiment.
// It supports various failure modes including killing, deleting, and pausing pods.
type PodFailureScenario struct {
	// name is the unique identifier for this scenario.
	name string

	// description describes what this scenario does.
	description string

	// Action specifies the failure action: "kill", "delete", or "pause".
	Action string `json:"action"`

	// Targets specifies the pod names or label selectors to target.
	// Can be pod names like "validator-0" or label selectors like "app=validator".
	Targets []string `json:"targets"`

	// Namespace is the Kubernetes namespace containing the target pods.
	Namespace string `json:"namespace"`

	// Count is the number of pods to affect simultaneously.
	Count int `json:"count"`

	// Duration is the duration of the failure (for pause action).
	Duration time.Duration `json:"duration"`

	// GracePeriod is the grace period for pod deletion.
	GracePeriod time.Duration `json:"gracePeriod"`

	// Interval is the interval between repeated failures (for rolling failures).
	Interval time.Duration `json:"interval,omitempty"`

	// Labels for categorization and filtering.
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations for additional metadata.
	Annotations map[string]string `json:"annotations,omitempty"`
}

// Name returns the scenario name.
func (s *PodFailureScenario) Name() string {
	return s.name
}

// Description returns the scenario description.
func (s *PodFailureScenario) Description() string {
	return s.description
}

// Type returns the experiment type.
func (s *PodFailureScenario) Type() ExperimentType {
	return ExperimentTypePodFailure
}

// Validate checks if the scenario configuration is valid.
func (s *PodFailureScenario) Validate() error {
	if s.name == "" {
		return ErrEmptyName
	}

	if len(s.Targets) == 0 {
		return ErrNoTargets
	}

	switch s.Action {
	case PodActionKill, PodActionDelete, PodActionPause:
		// Valid actions
	default:
		return fmt.Errorf("%w: %s (expected: %s, %s, or %s)",
			ErrInvalidAction, s.Action, PodActionKill, PodActionDelete, PodActionPause)
	}

	if s.Count <= 0 {
		return ErrInvalidCount
	}

	if s.Action == PodActionPause && s.Duration <= 0 {
		return fmt.Errorf("%w: pause action requires positive duration", ErrInvalidDuration)
	}

	if s.Namespace == "" {
		return ErrEmptyNamespace
	}

	return nil
}

// Build constructs an Experiment from this scenario.
func (s *PodFailureScenario) Build() (*Experiment, error) {
	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("invalid pod failure scenario: %w", err)
	}

	return &Experiment{
		Name:        s.name,
		Description: s.description,
		Type:        ExperimentTypePodFailure,
		Duration:    s.Duration,
		Targets:     s.Targets,
		Parameters: map[string]interface{}{
			"action":      s.Action,
			"namespace":   s.Namespace,
			"count":       s.Count,
			"gracePeriod": s.GracePeriod,
			"interval":    s.Interval,
			"labels":      s.Labels,
			"annotations": s.Annotations,
		},
	}, nil
}

// WithLabels adds labels to the scenario.
func (s *PodFailureScenario) WithLabels(labels map[string]string) *PodFailureScenario {
	if s.Labels == nil {
		s.Labels = make(map[string]string)
	}
	for k, v := range labels {
		s.Labels[k] = v
	}
	return s
}

// WithAnnotations adds annotations to the scenario.
func (s *PodFailureScenario) WithAnnotations(annotations map[string]string) *PodFailureScenario {
	if s.Annotations == nil {
		s.Annotations = make(map[string]string)
	}
	for k, v := range annotations {
		s.Annotations[k] = v
	}
	return s
}

// NewValidatorCrash creates a scenario that crashes a validator pod.
// This simulates a validator node crash to test consensus resilience.
//
// Parameters:
//   - validatorName: The name of the validator pod to crash
//   - duration: How long to wait before allowing recovery
//
// Example:
//
//	scenario := NewValidatorCrash("validator-0", 5*time.Minute)
func NewValidatorCrash(validatorName string, duration time.Duration) *PodFailureScenario {
	return &PodFailureScenario{
		name:        fmt.Sprintf("validator-crash-%s", validatorName),
		description: fmt.Sprintf("Crash validator pod %s to test consensus resilience", validatorName),
		Action:      PodActionKill,
		Targets:     []string{validatorName},
		Namespace:   "virtengine",
		Count:       1,
		Duration:    duration,
		GracePeriod: 0, // Immediate kill for crash simulation
		Labels: map[string]string{
			"chaos.virtengine.dev/category": "validator",
			"chaos.virtengine.dev/severity": "high",
		},
	}
}

// NewProviderDaemonCrash creates a scenario that crashes a provider daemon pod.
// This simulates a provider daemon crash to test provider failover.
//
// Parameters:
//   - providerName: The name of the provider daemon pod to crash
//   - duration: How long to wait before allowing recovery
//
// Example:
//
//	scenario := NewProviderDaemonCrash("provider-daemon-0", 5*time.Minute)
func NewProviderDaemonCrash(providerName string, duration time.Duration) *PodFailureScenario {
	return &PodFailureScenario{
		name:        fmt.Sprintf("provider-daemon-crash-%s", providerName),
		description: fmt.Sprintf("Crash provider daemon %s to test failover", providerName),
		Action:      PodActionKill,
		Targets:     []string{providerName},
		Namespace:   "virtengine",
		Count:       1,
		Duration:    duration,
		GracePeriod: 0, // Immediate kill for crash simulation
		Labels: map[string]string{
			"chaos.virtengine.dev/category": "provider",
			"chaos.virtengine.dev/severity": "medium",
		},
	}
}

// NewRandomPodKill creates a scenario that kills random pods matching a label selector.
// This simulates random pod failures to test overall system resilience.
//
// Parameters:
//   - namespace: The Kubernetes namespace to target
//   - labelSelector: Label selector to filter pods (e.g., "app=validator")
//   - count: Number of pods to kill
//
// Example:
//
//	scenario := NewRandomPodKill("virtengine", "app=validator", 2)
func NewRandomPodKill(namespace, labelSelector string, count int) *PodFailureScenario {
	return &PodFailureScenario{
		name:        fmt.Sprintf("random-pod-kill-%s-%d", sanitizeName(labelSelector), count),
		description: fmt.Sprintf("Kill %d random pods matching %s in %s", count, labelSelector, namespace),
		Action:      PodActionKill,
		Targets:     []string{labelSelector},
		Namespace:   namespace,
		Count:       count,
		Duration:    DefaultDuration,
		GracePeriod: 0, // Immediate kill
		Labels: map[string]string{
			"chaos.virtengine.dev/category": "random",
			"chaos.virtengine.dev/severity": "medium",
		},
	}
}

// NewRollingPodFailure creates a scenario that fails pods in a rolling fashion.
// This simulates rolling failures to test system behavior under sustained stress.
//
// Parameters:
//   - pods: List of pod names to fail in sequence
//   - interval: Time between each pod failure
//   - duration: How long each pod remains failed
//
// Example:
//
//	scenario := NewRollingPodFailure(
//	    []string{"validator-0", "validator-1", "validator-2"},
//	    1*time.Minute,
//	    5*time.Minute,
//	)
func NewRollingPodFailure(pods []string, interval, duration time.Duration) *PodFailureScenario {
	return &PodFailureScenario{
		name:        fmt.Sprintf("rolling-pod-failure-%d-pods", len(pods)),
		description: fmt.Sprintf("Rolling failure of %d pods with %v interval", len(pods), interval),
		Action:      PodActionKill,
		Targets:     pods,
		Namespace:   "virtengine",
		Count:       1, // One at a time for rolling failures
		Duration:    duration,
		GracePeriod: DefaultGracePeriod,
		Interval:    interval,
		Labels: map[string]string{
			"chaos.virtengine.dev/category": "rolling",
			"chaos.virtengine.dev/severity": "high",
		},
	}
}

// NodeFailureScenario defines a node failure chaos experiment.
// It supports various failure modes including draining, cordoning, rebooting, and shutting down nodes.
type NodeFailureScenario struct {
	// name is the unique identifier for this scenario.
	name string

	// description describes what this scenario does.
	description string

	// Action specifies the failure action: "drain", "cordon", "reboot", or "shutdown".
	Action string `json:"action"`

	// NodeName is the name of the Kubernetes node to target.
	NodeName string `json:"nodeName"`

	// Duration is the duration of the failure (for cordon action).
	Duration time.Duration `json:"duration"`

	// Force indicates whether to force the operation.
	Force bool `json:"force"`

	// EvictGracePeriod is the grace period for pod eviction during drain.
	EvictGracePeriod time.Duration `json:"evictGracePeriod"`

	// Labels for categorization and filtering.
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations for additional metadata.
	Annotations map[string]string `json:"annotations,omitempty"`
}

// Name returns the scenario name.
func (s *NodeFailureScenario) Name() string {
	return s.name
}

// Description returns the scenario description.
func (s *NodeFailureScenario) Description() string {
	return s.description
}

// Type returns the experiment type.
func (s *NodeFailureScenario) Type() ExperimentType {
	return ExperimentTypeNodeFailure
}

// Validate checks if the scenario configuration is valid.
func (s *NodeFailureScenario) Validate() error {
	if s.name == "" {
		return ErrEmptyName
	}

	if s.NodeName == "" {
		return ErrEmptyNodeName
	}

	switch s.Action {
	case NodeActionDrain, NodeActionCordon, NodeActionReboot, NodeActionShutdown:
		// Valid actions
	default:
		return fmt.Errorf("%w: %s (expected: %s, %s, %s, or %s)",
			ErrInvalidAction, s.Action,
			NodeActionDrain, NodeActionCordon, NodeActionReboot, NodeActionShutdown)
	}

	if s.Action == NodeActionCordon && s.Duration <= 0 {
		return fmt.Errorf("%w: cordon action requires positive duration", ErrInvalidDuration)
	}

	return nil
}

// Build constructs an Experiment from this scenario.
func (s *NodeFailureScenario) Build() (*Experiment, error) {
	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("invalid node failure scenario: %w", err)
	}

	return &Experiment{
		Name:        s.name,
		Description: s.description,
		Type:        ExperimentTypeNodeFailure,
		Duration:    s.Duration,
		Targets:     []string{s.NodeName},
		Parameters: map[string]interface{}{
			"action":           s.Action,
			"force":            s.Force,
			"evictGracePeriod": s.EvictGracePeriod,
			"labels":           s.Labels,
			"annotations":      s.Annotations,
		},
	}, nil
}

// WithLabels adds labels to the scenario.
func (s *NodeFailureScenario) WithLabels(labels map[string]string) *NodeFailureScenario {
	if s.Labels == nil {
		s.Labels = make(map[string]string)
	}
	for k, v := range labels {
		s.Labels[k] = v
	}
	return s
}

// WithAnnotations adds annotations to the scenario.
func (s *NodeFailureScenario) WithAnnotations(annotations map[string]string) *NodeFailureScenario {
	if s.Annotations == nil {
		s.Annotations = make(map[string]string)
	}
	for k, v := range annotations {
		s.Annotations[k] = v
	}
	return s
}

// NewNodeDrain creates a scenario that drains a node of all pods.
// This simulates a node maintenance scenario to test pod rescheduling.
//
// Parameters:
//   - nodeName: The name of the node to drain
//   - gracePeriod: Grace period for pod eviction
//
// Example:
//
//	scenario := NewNodeDrain("node-1", 60*time.Second)
func NewNodeDrain(nodeName string, gracePeriod time.Duration) *NodeFailureScenario {
	return &NodeFailureScenario{
		name:             fmt.Sprintf("node-drain-%s", nodeName),
		description:      fmt.Sprintf("Drain node %s to test pod rescheduling", nodeName),
		Action:           NodeActionDrain,
		NodeName:         nodeName,
		Duration:         DefaultDuration,
		Force:            false,
		EvictGracePeriod: gracePeriod,
		Labels: map[string]string{
			"chaos.virtengine.dev/category": "node",
			"chaos.virtengine.dev/severity": "high",
		},
	}
}

// NewNodeReboot creates a scenario that reboots a node.
// This simulates a node reboot to test cluster recovery.
//
// Parameters:
//   - nodeName: The name of the node to reboot
//
// Example:
//
//	scenario := NewNodeReboot("node-1")
func NewNodeReboot(nodeName string) *NodeFailureScenario {
	return &NodeFailureScenario{
		name:             fmt.Sprintf("node-reboot-%s", nodeName),
		description:      fmt.Sprintf("Reboot node %s to test cluster recovery", nodeName),
		Action:           NodeActionReboot,
		NodeName:         nodeName,
		Duration:         DefaultDuration,
		Force:            false,
		EvictGracePeriod: DefaultEvictGracePeriod,
		Labels: map[string]string{
			"chaos.virtengine.dev/category": "node",
			"chaos.virtengine.dev/severity": "critical",
		},
	}
}

// NewNodeCordon creates a scenario that cordons a node.
// This simulates marking a node as unschedulable to test workload placement.
//
// Parameters:
//   - nodeName: The name of the node to cordon
//   - duration: How long to keep the node cordoned
//
// Example:
//
//	scenario := NewNodeCordon("node-1", 10*time.Minute)
func NewNodeCordon(nodeName string, duration time.Duration) *NodeFailureScenario {
	return &NodeFailureScenario{
		name:             fmt.Sprintf("node-cordon-%s", nodeName),
		description:      fmt.Sprintf("Cordon node %s for %v to test workload placement", nodeName, duration),
		Action:           NodeActionCordon,
		NodeName:         nodeName,
		Duration:         duration,
		Force:            false,
		EvictGracePeriod: DefaultEvictGracePeriod,
		Labels: map[string]string{
			"chaos.virtengine.dev/category": "node",
			"chaos.virtengine.dev/severity": "medium",
		},
	}
}

// ContainerFailureScenario defines a container failure chaos experiment.
// It supports various failure modes including killing, pausing, and restarting containers.
type ContainerFailureScenario struct {
	// name is the unique identifier for this scenario.
	name string

	// description describes what this scenario does.
	description string

	// ContainerName is the name of the container to target.
	ContainerName string `json:"containerName"`

	// PodName is the name of the pod containing the container.
	PodName string `json:"podName"`

	// Namespace is the Kubernetes namespace.
	Namespace string `json:"namespace"`

	// Action specifies the failure action: "kill", "pause", or "restart".
	Action string `json:"action"`

	// Duration is the duration of the failure (for pause action).
	Duration time.Duration `json:"duration"`

	// Labels for categorization and filtering.
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations for additional metadata.
	Annotations map[string]string `json:"annotations,omitempty"`
}

// Name returns the scenario name.
func (s *ContainerFailureScenario) Name() string {
	return s.name
}

// Description returns the scenario description.
func (s *ContainerFailureScenario) Description() string {
	return s.description
}

// Type returns the experiment type.
func (s *ContainerFailureScenario) Type() ExperimentType {
	return ExperimentTypeContainerFailure
}

// Validate checks if the scenario configuration is valid.
func (s *ContainerFailureScenario) Validate() error {
	if s.name == "" {
		return ErrEmptyName
	}

	if s.ContainerName == "" {
		return ErrEmptyContainerName
	}

	if s.PodName == "" {
		return ErrEmptyPodName
	}

	if s.Namespace == "" {
		return ErrEmptyNamespace
	}

	switch s.Action {
	case ContainerActionKill, ContainerActionPause, ContainerActionRestart:
		// Valid actions
	default:
		return fmt.Errorf("%w: %s (expected: %s, %s, or %s)",
			ErrInvalidAction, s.Action,
			ContainerActionKill, ContainerActionPause, ContainerActionRestart)
	}

	if s.Action == ContainerActionPause && s.Duration <= 0 {
		return fmt.Errorf("%w: pause action requires positive duration", ErrInvalidDuration)
	}

	return nil
}

// Build constructs an Experiment from this scenario.
func (s *ContainerFailureScenario) Build() (*Experiment, error) {
	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("invalid container failure scenario: %w", err)
	}

	return &Experiment{
		Name:        s.name,
		Description: s.description,
		Type:        ExperimentTypeContainerFailure,
		Duration:    s.Duration,
		Targets:     []string{fmt.Sprintf("%s/%s/%s", s.Namespace, s.PodName, s.ContainerName)},
		Parameters: map[string]interface{}{
			"containerName": s.ContainerName,
			"podName":       s.PodName,
			"namespace":     s.Namespace,
			"action":        s.Action,
			"labels":        s.Labels,
			"annotations":   s.Annotations,
		},
	}, nil
}

// WithLabels adds labels to the scenario.
func (s *ContainerFailureScenario) WithLabels(labels map[string]string) *ContainerFailureScenario {
	if s.Labels == nil {
		s.Labels = make(map[string]string)
	}
	for k, v := range labels {
		s.Labels[k] = v
	}
	return s
}

// WithAnnotations adds annotations to the scenario.
func (s *ContainerFailureScenario) WithAnnotations(annotations map[string]string) *ContainerFailureScenario {
	if s.Annotations == nil {
		s.Annotations = make(map[string]string)
	}
	for k, v := range annotations {
		s.Annotations[k] = v
	}
	return s
}

// NewContainerKill creates a scenario that kills a container.
// This simulates a container crash to test container restart behavior.
//
// Parameters:
//   - namespace: The Kubernetes namespace
//   - pod: The name of the pod containing the container
//   - container: The name of the container to kill
//
// Example:
//
//	scenario := NewContainerKill("virtengine", "validator-0", "tendermint")
func NewContainerKill(namespace, pod, container string) *ContainerFailureScenario {
	return &ContainerFailureScenario{
		name:          fmt.Sprintf("container-kill-%s-%s-%s", namespace, pod, container),
		description:   fmt.Sprintf("Kill container %s in pod %s/%s", container, namespace, pod),
		ContainerName: container,
		PodName:       pod,
		Namespace:     namespace,
		Action:        ContainerActionKill,
		Duration:      DefaultDuration,
		Labels: map[string]string{
			"chaos.virtengine.dev/category": "container",
			"chaos.virtengine.dev/severity": "low",
		},
	}
}

// NewContainerPause creates a scenario that pauses a container.
// This simulates a container freeze to test timeout and recovery behavior.
//
// Parameters:
//   - namespace: The Kubernetes namespace
//   - pod: The name of the pod containing the container
//   - container: The name of the container to pause
//   - duration: How long to pause the container
//
// Example:
//
//	scenario := NewContainerPause("virtengine", "validator-0", "tendermint", 30*time.Second)
func NewContainerPause(namespace, pod, container string, duration time.Duration) *ContainerFailureScenario {
	return &ContainerFailureScenario{
		name:          fmt.Sprintf("container-pause-%s-%s-%s", namespace, pod, container),
		description:   fmt.Sprintf("Pause container %s in pod %s/%s for %v", container, namespace, pod, duration),
		ContainerName: container,
		PodName:       pod,
		Namespace:     namespace,
		Action:        ContainerActionPause,
		Duration:      duration,
		Labels: map[string]string{
			"chaos.virtengine.dev/category": "container",
			"chaos.virtengine.dev/severity": "medium",
		},
	}
}

// FailureStage defines a single stage in a cascade failure scenario.
// Each stage contains a scenario to execute and dependency information.
type FailureStage struct {
	// Name is the unique identifier for this stage.
	Name string `json:"name"`

	// Scenario is the chaos scenario to execute in this stage.
	Scenario NodeScenario `json:"scenario"`

	// Delay is the time to wait before executing this stage.
	Delay time.Duration `json:"delay"`

	// DependsOn lists the names of stages that must complete before this stage.
	DependsOn []string `json:"dependsOn,omitempty"`
}

// Validate checks if the stage configuration is valid.
func (s *FailureStage) Validate() error {
	if s.Name == "" {
		return ErrEmptyName
	}

	if s.Scenario == nil {
		return ErrNilScenario
	}

	if err := s.Scenario.Validate(); err != nil {
		return fmt.Errorf("invalid scenario in stage %s: %w", s.Name, err)
	}

	return nil
}

// CascadeFailureScenario defines a multi-stage cascade failure experiment.
// It allows testing cascading failure scenarios where multiple failures occur
// in sequence, potentially with dependencies between them.
type CascadeFailureScenario struct {
	// name is the unique identifier for this scenario.
	name string

	// description describes what this scenario does.
	description string

	// Stages contains the ordered list of failure stages.
	Stages []FailureStage `json:"stages"`

	// StageInterval is the default interval between stages.
	StageInterval time.Duration `json:"stageInterval"`

	// Labels for categorization and filtering.
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations for additional metadata.
	Annotations map[string]string `json:"annotations,omitempty"`
}

// Name returns the scenario name.
func (s *CascadeFailureScenario) Name() string {
	return s.name
}

// Description returns the scenario description.
func (s *CascadeFailureScenario) Description() string {
	return s.description
}

// Type returns the experiment type.
func (s *CascadeFailureScenario) Type() ExperimentType {
	return ExperimentTypeCascadeFailure
}

// Validate checks if the scenario configuration is valid.
func (s *CascadeFailureScenario) Validate() error {
	if s.name == "" {
		return ErrEmptyName
	}

	if len(s.Stages) == 0 {
		return ErrNoStages
	}

	if s.StageInterval < 0 {
		return ErrInvalidInterval
	}

	// Validate each stage
	stageNames := make(map[string]bool)
	for _, stage := range s.Stages {
		if err := stage.Validate(); err != nil {
			return err
		}

		if stageNames[stage.Name] {
			return fmt.Errorf("duplicate stage name: %s", stage.Name)
		}
		stageNames[stage.Name] = true
	}

	// Validate dependencies exist
	for _, stage := range s.Stages {
		for _, dep := range stage.DependsOn {
			if !stageNames[dep] {
				return fmt.Errorf("stage %s depends on non-existent stage %s", stage.Name, dep)
			}
		}
	}

	// Check for circular dependencies
	if hasCircularDependency(s.Stages) {
		return ErrCircularDependency
	}

	return nil
}

// Build constructs an Experiment from this scenario.
func (s *CascadeFailureScenario) Build() (*Experiment, error) {
	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("invalid cascade failure scenario: %w", err)
	}

	// Calculate total duration from all stages
	var totalDuration time.Duration
	var targets []string
	for _, stage := range s.Stages {
		if exp, err := stage.Scenario.Build(); err == nil {
			totalDuration += stage.Delay + exp.Duration
			targets = append(targets, exp.Targets...)
		}
	}
	totalDuration += s.StageInterval * time.Duration(len(s.Stages)-1)

	// Build stage info for parameters
	stageInfo := make([]map[string]interface{}, len(s.Stages))
	for i, stage := range s.Stages {
		stageInfo[i] = map[string]interface{}{
			"name":      stage.Name,
			"delay":     stage.Delay,
			"dependsOn": stage.DependsOn,
		}
	}

	return &Experiment{
		Name:        s.name,
		Description: s.description,
		Type:        ExperimentTypeCascadeFailure,
		Duration:    totalDuration,
		Targets:     targets,
		Parameters: map[string]interface{}{
			"stages":        stageInfo,
			"stageInterval": s.StageInterval,
			"labels":        s.Labels,
			"annotations":   s.Annotations,
		},
	}, nil
}

// WithLabels adds labels to the scenario.
func (s *CascadeFailureScenario) WithLabels(labels map[string]string) *CascadeFailureScenario {
	if s.Labels == nil {
		s.Labels = make(map[string]string)
	}
	for k, v := range labels {
		s.Labels[k] = v
	}
	return s
}

// WithAnnotations adds annotations to the scenario.
func (s *CascadeFailureScenario) WithAnnotations(annotations map[string]string) *CascadeFailureScenario {
	if s.Annotations == nil {
		s.Annotations = make(map[string]string)
	}
	for k, v := range annotations {
		s.Annotations[k] = v
	}
	return s
}

// NewCascadeFailure creates a cascade failure scenario from a list of stages.
// This allows testing complex failure scenarios with dependencies.
//
// Parameters:
//   - stages: The ordered list of failure stages
//
// Example:
//
//	stages := []FailureStage{
//	    {Name: "stage1", Scenario: NewValidatorCrash("validator-0", 5*time.Minute)},
//	    {Name: "stage2", Scenario: NewProviderDaemonCrash("provider-0", 5*time.Minute), DependsOn: []string{"stage1"}},
//	}
//	scenario := NewCascadeFailure(stages)
func NewCascadeFailure(stages []FailureStage) *CascadeFailureScenario {
	stageNames := make([]string, len(stages))
	for i, s := range stages {
		stageNames[i] = s.Name
	}

	return &CascadeFailureScenario{
		name:          fmt.Sprintf("cascade-failure-%d-stages", len(stages)),
		description:   fmt.Sprintf("Cascade failure with stages: %s", strings.Join(stageNames, ", ")),
		Stages:        stages,
		StageInterval: DefaultInterval,
		Labels: map[string]string{
			"chaos.virtengine.dev/category": "cascade",
			"chaos.virtengine.dev/severity": "critical",
		},
	}
}

// DefaultNodeScenarios returns a collection of commonly used node failure scenarios.
// These scenarios cover typical failure modes for VirtEngine clusters.
//
// Example:
//
//	scenarios := DefaultNodeScenarios()
//	for _, scenario := range scenarios {
//	    fmt.Printf("Scenario: %s - %s\n", scenario.Name(), scenario.Description())
//	}
func DefaultNodeScenarios() []NodeScenario {
	return []NodeScenario{
		// Validator failures
		NewValidatorCrash("validator-0", 5*time.Minute),
		NewValidatorCrash("validator-1", 5*time.Minute),

		// Provider daemon failures
		NewProviderDaemonCrash("provider-daemon-0", 5*time.Minute),

		// Random pod kills
		NewRandomPodKill("virtengine", "app=validator", 1),
		NewRandomPodKill("virtengine", "app=provider-daemon", 1),

		// Rolling failures
		NewRollingPodFailure(
			[]string{"validator-0", "validator-1", "validator-2"},
			2*time.Minute,
			5*time.Minute,
		),

		// Node failures
		NewNodeDrain("node-1", 60*time.Second),
		NewNodeCordon("node-1", 10*time.Minute),

		// Container failures
		NewContainerKill("virtengine", "validator-0", "tendermint"),
		NewContainerPause("virtengine", "validator-0", "tendermint", 30*time.Second),

		// Cascade failure: validator crash followed by provider crash
		NewCascadeFailure([]FailureStage{
			{
				Name:     "validator-crash",
				Scenario: NewValidatorCrash("validator-0", 5*time.Minute),
				Delay:    0,
			},
			{
				Name:      "provider-crash",
				Scenario:  NewProviderDaemonCrash("provider-daemon-0", 5*time.Minute),
				Delay:     30 * time.Second,
				DependsOn: []string{"validator-crash"},
			},
		}),
	}
}

// Helper functions

// sanitizeName converts a string to a valid Kubernetes name component.
func sanitizeName(s string) string {
	// Replace common separators with dashes
	s = strings.ReplaceAll(s, "=", "-")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, ":", "-")
	s = strings.ReplaceAll(s, " ", "-")

	// Remove any remaining invalid characters
	var result strings.Builder
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '.' {
			result.WriteRune(c)
		}
	}

	// Convert to lowercase
	return strings.ToLower(result.String())
}

// hasCircularDependency checks if there are circular dependencies in the stages.
func hasCircularDependency(stages []FailureStage) bool {
	// Build adjacency list
	deps := make(map[string][]string)
	for _, stage := range stages {
		deps[stage.Name] = stage.DependsOn
	}

	// DFS to detect cycles
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var hasCycle func(node string) bool
	hasCycle = func(node string) bool {
		visited[node] = true
		recStack[node] = true

		for _, dep := range deps[node] {
			if !visited[dep] {
				if hasCycle(dep) {
					return true
				}
			} else if recStack[dep] {
				return true
			}
		}

		recStack[node] = false
		return false
	}

	for _, stage := range stages {
		if !visited[stage.Name] {
			if hasCycle(stage.Name) {
				return true
			}
		}
	}

	return false
}

