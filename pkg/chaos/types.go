// Copyright 2024-2026 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0

// Package chaos provides core types and interfaces for chaos engineering
// experiments in VirtEngine. It supports controlled failure injection,
// steady-state hypothesis verification, and automated rollback capabilities.
//
// # Overview
//
// The chaos engineering framework enables controlled experiments to verify
// system resilience. It follows the principles of the Chaos Engineering
// discipline: define steady state, hypothesize impact, inject chaos,
// verify behavior, and rollback if necessary.
//
// # Architecture
//
// The chaos package is structured as follows:
//   - types.go: Core type definitions (ExperimentSpec, Target, Probe, etc.)
//   - controller.go: Experiment orchestration and scheduling
//   - metrics.go: Prometheus metrics for chaos experiments
//   - scenarios/: Specific chaos scenario implementations
//
// # Key Types
//
//   - ExperimentSpec: Complete definition of a chaos experiment
//   - ExperimentState: Current execution state (Pending, Running, etc.)
//   - Target: Resources to affect during experiments
//   - SteadyStateHypothesis: Expected system behavior to verify
//   - ExperimentRunner: Interface for executing experiments
//   - SLOVerifier: Interface for SLO compliance checking
package chaos

import (
	"context"
	"time"
)

// ============================================================================
// Experiment State - Lifecycle states for experiments
// ============================================================================

// ExperimentState represents the current state of an experiment.
type ExperimentState string

const (
	// ExperimentStatePending indicates the experiment is waiting to start.
	ExperimentStatePending ExperimentState = "pending"

	// ExperimentStateRunning indicates the experiment is actively running.
	ExperimentStateRunning ExperimentState = "running"

	// ExperimentStatePaused indicates the experiment is temporarily paused.
	ExperimentStatePaused ExperimentState = "paused"

	// ExperimentStateCompleted indicates the experiment finished successfully.
	ExperimentStateCompleted ExperimentState = "completed"

	// ExperimentStateFailed indicates the experiment failed.
	ExperimentStateFailed ExperimentState = "failed"

	// ExperimentStateAborted indicates the experiment was manually aborted.
	ExperimentStateAborted ExperimentState = "aborted"

	// ExperimentStateCreated indicates the experiment has been created but not started.
	ExperimentStateCreated ExperimentState = "created"

	// ExperimentStateRollingBack indicates the experiment is executing rollback.
	ExperimentStateRollingBack ExperimentState = "rolling_back"

	// ExperimentStateCanceled indicates the experiment was canceled.
	ExperimentStateCanceled ExperimentState = "canceled"
)

// String returns the string representation of the experiment state.
func (s ExperimentState) String() string {
	return string(s)
}

// Valid returns true if the experiment state is valid.
func (s ExperimentState) Valid() bool {
	switch s {
	case ExperimentStatePending, ExperimentStateRunning, ExperimentStatePaused,
		ExperimentStateCompleted, ExperimentStateFailed, ExperimentStateAborted,
		ExperimentStateCreated, ExperimentStateRollingBack, ExperimentStateCanceled:
		return true
	default:
		return false
	}
}

// IsTerminal returns true if the state is a terminal state.
func (s ExperimentState) IsTerminal() bool {
	switch s {
	case ExperimentStateCompleted, ExperimentStateFailed, ExperimentStateAborted, ExperimentStateCanceled:
		return true
	default:
		return false
	}
}

// ============================================================================
// Experiment Type - Categories of chaos experiments
// ============================================================================

// ExperimentType defines the category of chaos experiment.
type ExperimentType string

const (
	// Network chaos experiments
	ExperimentTypeNetworkPartition ExperimentType = "network-partition"
	ExperimentTypeNetworkLatency   ExperimentType = "network-latency"
	ExperimentTypePacketLoss       ExperimentType = "packet-loss"
	ExperimentTypeBandwidth        ExperimentType = "bandwidth-limit"

	// Node/Pod chaos experiments
	ExperimentTypePodFailure       ExperimentType = "pod-failure"
	ExperimentTypeNodeFailure      ExperimentType = "node-failure"
	ExperimentTypeContainerFailure ExperimentType = "container-failure"
	ExperimentTypeCascadeFailure   ExperimentType = "cascade-failure"

	// Resource chaos experiments
	ExperimentTypeCPUStress     ExperimentType = "cpu-stress"
	ExperimentTypeMemoryStress  ExperimentType = "memory-stress"
	ExperimentTypeDiskStress    ExperimentType = "disk-stress"
	ExperimentTypeProcessStress ExperimentType = "process-stress"
	ExperimentTypeClockSkew     ExperimentType = "clock-skew"
	ExperimentTypeTimeChaos     ExperimentType = "time-chaos"

	// Byzantine chaos experiments
	ExperimentTypeByzantineDoubleSigning       ExperimentType = "byzantine-double-signing"
	ExperimentTypeByzantineEquivocation        ExperimentType = "byzantine-equivocation"
	ExperimentTypeByzantineInvalidBlock        ExperimentType = "byzantine-invalid-block"
	ExperimentTypeByzantineMessageTampering    ExperimentType = "byzantine-message-tampering"
	ExperimentTypeByzantineSelectiveForwarding ExperimentType = "byzantine-selective-forwarding"
	ExperimentTypeByzantineGeneric             ExperimentType = "byzantine"
)

// String returns the string representation of the experiment type.
func (t ExperimentType) String() string {
	return string(t)
}

// ============================================================================
// Experiment - Lightweight experiment reference
// ============================================================================

// Experiment represents a lightweight chaos experiment reference for metrics
// and status tracking. Use ExperimentSpec for the full experiment definition.
type Experiment struct {
	// ID is a unique identifier for this experiment.
	ID string `json:"id"`

	// Name is a human-readable name for this experiment.
	Name string `json:"name"`

	// Description provides human-readable details about the experiment.
	Description string `json:"description,omitempty"`

	// Type categorizes the experiment.
	Type ExperimentType `json:"type"`

	// State is the current execution state.
	State ExperimentState `json:"state"`

	// Duration specifies how long the experiment should run.
	Duration time.Duration `json:"duration"`

	// Targets lists the nodes or endpoints affected by this experiment.
	Targets []string `json:"targets,omitempty"`

	// StartTime is when the experiment began.
	StartTime time.Time `json:"start_time,omitempty"`

	// EndTime is when the experiment completed.
	EndTime time.Time `json:"end_time,omitempty"`

	// Labels for categorization and filtering.
	Labels map[string]string `json:"labels,omitempty"`

	// Tags for metrics and filtering (first tag often used as experiment type label).
	Tags []string `json:"tags,omitempty"`

	// Namespace is the Kubernetes namespace for the experiment.
	Namespace string `json:"namespace,omitempty"`
}

// ============================================================================
// ExperimentSpec - Complete experiment definition
// ============================================================================

// ExperimentSpec represents a complete chaos experiment definition and state.
type ExperimentSpec struct {
	// ID is a unique identifier for this experiment.
	ID string `json:"id"`

	// Name is a human-readable name for this experiment.
	Name string `json:"name"`

	// Description provides detailed information about the experiment.
	Description string `json:"description,omitempty"`

	// Type specifies what kind of chaos to inject.
	Type ExperimentType `json:"type"`

	// State is the current execution state.
	State ExperimentState `json:"state"`

	// Targets specifies what resources to affect.
	Targets []Target `json:"targets"`

	// Duration is how long to run the chaos injection.
	Duration time.Duration `json:"duration"`

	// Interval specifies the period between chaos actions.
	Interval time.Duration `json:"interval,omitempty"`

	// Parameters contains type-specific configuration.
	Parameters map[string]interface{} `json:"parameters,omitempty"`

	// SteadyStateHypothesis defines expected system state.
	SteadyStateHypothesis *SteadyStateHypothesis `json:"steady_state_hypothesis,omitempty"`

	// Rollback defines how to restore system state.
	Rollback *RollbackAction `json:"rollback,omitempty"`

	// StartTime is when the experiment began.
	StartTime time.Time `json:"start_time,omitempty"`

	// EndTime is when the experiment completed.
	EndTime time.Time `json:"end_time,omitempty"`

	// Results contains the experiment outcome.
	Results *ExperimentResults `json:"results,omitempty"`

	// Labels for organizing and filtering experiments.
	Labels map[string]string `json:"labels,omitempty"`

	// DryRun indicates whether to simulate without actual chaos injection.
	DryRun bool `json:"dry_run,omitempty"`

	// FailFast indicates whether to stop immediately on first failure.
	FailFast bool `json:"fail_fast,omitempty"`

	// CreatedAt is when the experiment was defined.
	CreatedAt time.Time `json:"created_at,omitempty"`

	// UpdatedAt is when the experiment was last modified.
	UpdatedAt time.Time `json:"updated_at,omitempty"`

	// Namespace is the Kubernetes namespace for the experiment.
	Namespace string `json:"namespace,omitempty"`
}

// ToMetricsExperiment converts an ExperimentSpec to the Experiment type
// used for metrics tracking in metrics.go.
func (e *ExperimentSpec) ToMetricsExperiment() *Experiment {
	return &Experiment{
		ID:        e.ID,
		Name:      e.Name,
		Type:      e.Type,
		Namespace: e.Namespace,
		StartTime: e.StartTime,
		EndTime:   e.EndTime,
		Labels:    e.Labels,
	}
}

// ============================================================================
// ExperimentResults - Outcome of completed experiments
// ============================================================================

// ExperimentResults contains the outcome of a completed experiment.
type ExperimentResults struct {
	// Success indicates whether the experiment met its objectives.
	Success bool `json:"success"`

	// Duration is how long the experiment ran.
	Duration time.Duration `json:"duration"`

	// SteadyStateViolations records all steady-state breaches.
	SteadyStateViolations []SteadyStateViolation `json:"steady_state_violations,omitempty"`

	// Metrics contains collected numerical data.
	Metrics map[string]float64 `json:"metrics,omitempty"`

	// Timeline contains chronological events during the experiment.
	Timeline []Event `json:"timeline,omitempty"`

	// Error contains any error message if the experiment failed.
	Error string `json:"error,omitempty"`

	// Summary provides a human-readable summary of results.
	Summary string `json:"summary,omitempty"`

	// RollbackPerformed indicates if rollback was executed.
	RollbackPerformed bool `json:"rollback_performed,omitempty"`

	// RollbackSuccess indicates if rollback succeeded.
	RollbackSuccess bool `json:"rollback_success,omitempty"`
}

// ============================================================================
// SteadyStateViolation - Records of steady state breaches
// ============================================================================

// SteadyStateViolation records when and how steady state was broken.
type SteadyStateViolation struct {
	// Timestamp is when the violation was detected.
	Timestamp time.Time `json:"timestamp"`

	// ProbeName identifies which probe detected the violation.
	ProbeName string `json:"probe_name"`

	// ExpectedValue is what we expected to see.
	ExpectedValue interface{} `json:"expected_value"`

	// ActualValue is what we actually observed.
	ActualValue interface{} `json:"actual_value"`

	// Deviation is the percentage deviation from expected.
	Deviation float64 `json:"deviation"`

	// Message provides additional context.
	Message string `json:"message,omitempty"`

	// Tolerance is the allowed tolerance threshold.
	Tolerance float64 `json:"tolerance,omitempty"`
}

// ============================================================================
// SteadyStateHypothesis - Expected system state definition
// ============================================================================

// SteadyStateHypothesis defines the expected system state before and after experiments.
type SteadyStateHypothesis struct {
	// Name is a human-readable identifier for the hypothesis.
	Name string `json:"name"`

	// Title provides a descriptive title for the hypothesis.
	Title string `json:"title,omitempty"`

	// Probes are the checks to verify the steady state.
	Probes []Probe `json:"probes"`

	// Tolerance is the acceptable deviation from expected values (0.0-1.0).
	Tolerance float64 `json:"tolerance"`
}

// ============================================================================
// RollbackAction - System state restoration configuration
// ============================================================================

// RollbackAction defines how to restore system state after an experiment.
type RollbackAction struct {
	// Name is a human-readable identifier for the rollback action.
	Name string `json:"name,omitempty"`

	// Type specifies the rollback mechanism.
	Type string `json:"type"`

	// Provider specifies the rollback provider.
	Provider string `json:"provider,omitempty"`

	// Parameters are configuration options for the rollback.
	Parameters map[string]interface{} `json:"parameters,omitempty"`

	// Config provides additional configuration for the rollback.
	Config map[string]interface{} `json:"config,omitempty"`

	// Timeout is the maximum time allowed for rollback.
	Timeout time.Duration `json:"timeout"`

	// Force indicates whether to force rollback even on errors.
	Force bool `json:"force,omitempty"`
}

// ============================================================================
// ExperimentRunner Interface - Experiment execution abstraction
// ============================================================================

// ExperimentRunner defines the interface for executing chaos experiments.
type ExperimentRunner interface {
	// Start begins execution of a chaos experiment.
	// Returns an error if the experiment cannot be started.
	Start(ctx context.Context, experiment *ExperimentSpec) error

	// Stop gracefully terminates a running experiment.
	// Returns an error if the experiment cannot be stopped.
	Stop(ctx context.Context, experiment *ExperimentSpec) error

	// Status retrieves the current state of an experiment by ID.
	// Returns the experiment and any error encountered.
	Status(ctx context.Context, experimentID string) (*ExperimentSpec, error)

	// Rollback reverts changes made by an experiment.
	// Returns an error if rollback fails.
	Rollback(ctx context.Context, experiment *ExperimentSpec) error
}

// ============================================================================
// SLOVerifier Interface - SLO compliance checking
// ============================================================================

// SLOVerifier defines the interface for verifying SLO compliance during experiments.
type SLOVerifier interface {
	// Verify checks if the system meets SLO requirements.
	// Returns success status, any violations found, and an error if verification fails.
	Verify(ctx context.Context, experiment *ExperimentSpec) (bool, []Violation, error)

	// VerifySteadyState verifies the steady state hypothesis.
	VerifySteadyState(ctx context.Context, hypothesis *SteadyStateHypothesis) ([]SteadyStateViolation, error)

	// CheckProbe checks a single steady state probe.
	CheckProbe(ctx context.Context, probe Probe) (float64, error)
}

// ============================================================================
// Target Types - What resources to affect during experiments
// ============================================================================

// TargetType defines the type of resource being targeted.
type TargetType string

const (
	// TargetTypePod targets Kubernetes pods.
	TargetTypePod TargetType = "pod"

	// TargetTypeNode targets Kubernetes nodes.
	TargetTypeNode TargetType = "node"

	// TargetTypeNamespace targets entire Kubernetes namespaces.
	TargetTypeNamespace TargetType = "namespace"

	// TargetTypeService targets Kubernetes services.
	TargetTypeService TargetType = "service"

	// TargetTypeNetwork targets network resources.
	TargetTypeNetwork TargetType = "network"

	// TargetTypeDeployment targets Kubernetes deployments.
	TargetTypeDeployment TargetType = "deployment"

	// TargetTypeStatefulSet targets Kubernetes statefulsets.
	TargetTypeStatefulSet TargetType = "statefulset"

	// TargetTypeContainer targets specific containers within pods.
	TargetTypeContainer TargetType = "container"
)

// String returns the string representation of the target type.
func (t TargetType) String() string {
	return string(t)
}

// Valid returns true if the target type is valid.
func (t TargetType) Valid() bool {
	switch t {
	case TargetTypePod, TargetTypeNode, TargetTypeNamespace,
		TargetTypeService, TargetTypeNetwork, TargetTypeDeployment,
		TargetTypeStatefulSet, TargetTypeContainer:
		return true
	default:
		return false
	}
}

// Target defines what resources to affect during a chaos experiment.
type Target struct {
	// Type specifies the kind of resource being targeted.
	Type TargetType `json:"type"`

	// Selector is a label selector for matching resources.
	Selector map[string]string `json:"selector,omitempty"`

	// Name is the specific name of the target resource.
	Name string `json:"name,omitempty"`

	// Namespace is the Kubernetes namespace of the target.
	Namespace string `json:"namespace,omitempty"`

	// Percentage of matched targets to affect (0-100).
	Percentage int `json:"percentage,omitempty"`

	// Mode defines how targets are selected (one, all, fixed, percentage).
	Mode string `json:"mode,omitempty"`

	// Value is used with fixed mode to specify number of targets.
	Value int `json:"value,omitempty"`
}

// Validate validates the target configuration.
func (t *Target) Validate() error {
	if !t.Type.Valid() {
		return ErrInvalidTarget
	}
	if t.Name == "" && len(t.Selector) == 0 {
		return ErrTargetNotSpecified
	}
	if t.Percentage < 0 || t.Percentage > 100 {
		return ErrInvalidPercentage
	}
	return nil
}

// ============================================================================
// Probe Types - Steady state verification mechanisms
// ============================================================================

// ProbeType defines the type of probe used for steady-state verification.
type ProbeType string

const (
	// ProbeTypeHTTP uses HTTP requests for probing.
	ProbeTypeHTTP ProbeType = "http"

	// ProbeTypeGRPC uses gRPC health checks for probing.
	ProbeTypeGRPC ProbeType = "grpc"

	// ProbeTypePrometheus uses Prometheus queries for probing.
	ProbeTypePrometheus ProbeType = "prometheus"

	// ProbeTypeCommand uses shell commands for probing.
	ProbeTypeCommand ProbeType = "command"

	// ProbeTypeKubernetes uses Kubernetes API for probing.
	ProbeTypeKubernetes ProbeType = "kubernetes"

	// ProbeTypeCustom allows for custom probe implementations.
	ProbeTypeCustom ProbeType = "custom"
)

// String returns the string representation of the probe type.
func (p ProbeType) String() string {
	return string(p)
}

// Valid returns true if the probe type is valid.
func (p ProbeType) Valid() bool {
	switch p {
	case ProbeTypeHTTP, ProbeTypeGRPC, ProbeTypePrometheus,
		ProbeTypeCommand, ProbeTypeKubernetes, ProbeTypeCustom:
		return true
	default:
		return false
	}
}

// Probe defines a check to verify system state during experiments.
type Probe struct {
	// Type specifies the probe mechanism.
	Type ProbeType `json:"type"`

	// Name is a human-readable identifier for the probe.
	Name string `json:"name"`

	// URL is the endpoint for HTTP/gRPC probes.
	URL string `json:"url,omitempty"`

	// Query is the Prometheus query or similar.
	Query string `json:"query,omitempty"`

	// Command is the shell command for command-type probes.
	Command string `json:"command,omitempty"`

	// Interval specifies how often to run the probe.
	Interval time.Duration `json:"interval"`

	// Timeout specifies the maximum time to wait for probe response.
	Timeout time.Duration `json:"timeout"`

	// SuccessCriteria is an expression defining success (e.g., "status == 200").
	SuccessCriteria string `json:"success_criteria"`

	// ExpectedValue is the expected value for comparison probes.
	ExpectedValue interface{} `json:"expected_value,omitempty"`

	// Headers for HTTP probes.
	Headers map[string]string `json:"headers,omitempty"`

	// Method for HTTP probes (GET, POST, etc.).
	Method string `json:"method,omitempty"`

	// InitialDelay before first probe execution.
	InitialDelay time.Duration `json:"initial_delay,omitempty"`
}

// Validate validates the probe configuration.
func (p *Probe) Validate() error {
	if !p.Type.Valid() {
		return ErrInvalidProbeType
	}
	if p.Name == "" {
		return ErrProbeNameRequired
	}
	if p.Timeout <= 0 {
		return ErrInvalidTimeout
	}
	return nil
}

// ============================================================================
// Event Types - Timeline event tracking
// ============================================================================

// EventType categorizes timeline events during experiment execution.
type EventType string

const (
	// EventTypeExperimentStarted indicates the experiment began.
	EventTypeExperimentStarted EventType = "experiment_started"

	// EventTypeExperimentCompleted indicates the experiment finished.
	EventTypeExperimentCompleted EventType = "experiment_completed"

	// EventTypeExperimentFailed indicates the experiment failed.
	EventTypeExperimentFailed EventType = "experiment_failed"

	// EventTypeExperimentAborted indicates the experiment was aborted.
	EventTypeExperimentAborted EventType = "experiment_aborted"

	// EventTypeChaosInjected indicates chaos was injected into the system.
	EventTypeChaosInjected EventType = "chaos_injected"

	// EventTypeChaosRemoved indicates chaos was removed from the system.
	EventTypeChaosRemoved EventType = "chaos_removed"

	// EventTypeSteadyStateCheck indicates a steady-state probe was run.
	EventTypeSteadyStateCheck EventType = "steady_state_check"

	// EventTypeSteadyStateViolation indicates steady state was violated.
	EventTypeSteadyStateViolation EventType = "steady_state_violation"

	// EventTypeRollbackStarted indicates rollback began.
	EventTypeRollbackStarted EventType = "rollback_started"

	// EventTypeRollbackCompleted indicates rollback finished.
	EventTypeRollbackCompleted EventType = "rollback_completed"

	// EventTypeRollbackFailed indicates rollback failed.
	EventTypeRollbackFailed EventType = "rollback_failed"

	// EventTypeProbeSuccess indicates a probe succeeded.
	EventTypeProbeSuccess EventType = "probe_success"

	// EventTypeProbeFailed indicates a probe failed.
	EventTypeProbeFailed EventType = "probe_failed"

	// EventTypeTargetSelected indicates targets were selected for chaos.
	EventTypeTargetSelected EventType = "target_selected"

	// EventTypeActionStarted indicates an action started execution.
	EventTypeActionStarted EventType = "action_started"

	// EventTypeActionCompleted indicates an action completed.
	EventTypeActionCompleted EventType = "action_completed"
)

// String returns the string representation of the event type.
func (e EventType) String() string {
	return string(e)
}

// Event represents a timeline entry during experiment execution.
type Event struct {
	// Type categorizes this event.
	Type EventType `json:"type"`

	// Timestamp is when this event occurred.
	Timestamp time.Time `json:"timestamp"`

	// Message provides human-readable details.
	Message string `json:"message"`

	// Details contains structured event data.
	Details map[string]interface{} `json:"details,omitempty"`

	// Source identifies what generated this event.
	Source string `json:"source,omitempty"`

	// Severity indicates event importance (info, warning, error).
	Severity string `json:"severity,omitempty"`

	// ExperimentID links this event to a specific experiment.
	ExperimentID string `json:"experiment_id,omitempty"`
}

// NewEvent creates a new event with the current timestamp.
func NewEvent(eventType EventType, message string) *Event {
	return &Event{
		Type:      eventType,
		Timestamp: time.Now().UTC(),
		Message:   message,
		Severity:  "info",
	}
}

// WithDetails adds structured details to the event.
func (e *Event) WithDetails(details map[string]interface{}) *Event {
	e.Details = details
	return e
}

// WithSeverity sets the event severity.
func (e *Event) WithSeverity(severity string) *Event {
	e.Severity = severity
	return e
}

// ============================================================================
// Violation Types - SLO and steady state violations
// ============================================================================

// Violation represents an SLO violation during experiment execution.
type Violation struct {
	// SLOName identifies which SLO was violated.
	SLOName string `json:"slo_name"`

	// Timestamp is when the violation occurred.
	Timestamp time.Time `json:"timestamp"`

	// ExpectedValue is the SLO threshold.
	ExpectedValue float64 `json:"expected_value"`

	// ActualValue is the observed value.
	ActualValue float64 `json:"actual_value"`

	// Severity indicates the violation severity (warning, critical).
	Severity string `json:"severity"`

	// Message provides additional context.
	Message string `json:"message,omitempty"`

	// Duration is how long the violation persisted.
	Duration time.Duration `json:"duration,omitempty"`

	// ExperimentID links this violation to a specific experiment.
	ExperimentID string `json:"experiment_id,omitempty"`
}

// IsCritical returns true if this is a critical severity violation.
func (v *Violation) IsCritical() bool {
	return v.Severity == "critical"
}

// ============================================================================
// ChaosInjector Interface - Fault injection abstraction
// ============================================================================

// ChaosInjector defines the interface for injecting specific chaos types.
type ChaosInjector interface {
	// Inject applies chaos to the specified targets.
	Inject(ctx context.Context, targets []Target, params map[string]interface{}) error

	// Remove reverses chaos injection.
	Remove(ctx context.Context, targets []Target) error

	// Type returns the experiment type this injector handles.
	Type() ExperimentType

	// Validate checks if the parameters are valid for this injector.
	Validate(params map[string]interface{}) error

	// Supports returns true if this injector supports the given target type.
	Supports(targetType TargetType) bool
}

// ============================================================================
// ExperimentStore Interface - Persistence abstraction
// ============================================================================

// ExperimentStore defines the interface for persisting experiments.
type ExperimentStore interface {
	// Create stores a new experiment.
	Create(ctx context.Context, experiment *Experiment) error

	// Get retrieves an experiment by ID.
	Get(ctx context.Context, id string) (*Experiment, error)

	// Update modifies an existing experiment.
	Update(ctx context.Context, experiment *Experiment) error

	// Delete removes an experiment.
	Delete(ctx context.Context, id string) error

	// List retrieves experiments matching the given filter.
	List(ctx context.Context, filter ExperimentListFilter) ([]*Experiment, error)

	// GetByName retrieves an experiment by name.
	GetByName(ctx context.Context, name string) (*Experiment, error)
}

// ExperimentListFilter defines criteria for listing experiments.
type ExperimentListFilter struct {
	// States filters by experiment states.
	States []ExperimentState `json:"states,omitempty"`

	// Tags filters by experiment tags.
	Tags []string `json:"tags,omitempty"`

	// Limit restricts the number of results.
	Limit int `json:"limit,omitempty"`

	// Offset skips the first N results.
	Offset int `json:"offset,omitempty"`

	// SortBy specifies the field to sort by.
	SortBy string `json:"sort_by,omitempty"`

	// SortOrder specifies ascending or descending order.
	SortOrder string `json:"sort_order,omitempty"`
}

// ============================================================================
// Additional Type Errors
// ============================================================================

// Error constants for type validation.
var (
	// ErrInvalidTarget is returned when target configuration is invalid.
	ErrInvalidTarget = errorf("invalid target configuration")

	// ErrTargetNotSpecified is returned when no target name or selector is provided.
	ErrTargetNotSpecified = errorf("target name or selector must be specified")

	// ErrInvalidPercentage is returned when percentage is out of range.
	ErrInvalidPercentage = errorf("percentage must be between 0 and 100")

	// ErrInvalidProbeType is returned when probe type is unknown.
	ErrInvalidProbeType = errorf("invalid probe type")

	// ErrProbeNameRequired is returned when probe name is empty.
	ErrProbeNameRequired = errorf("probe name is required")

	// ErrInvalidTimeout is returned when timeout is invalid.
	ErrInvalidTimeout = errorf("timeout must be positive")
)

// errorf creates a simple error with a message.
func errorf(msg string) error {
	return &chaosError{msg: msg}
}

type chaosError struct {
	msg string
}

func (e *chaosError) Error() string {
	return e.msg
}

