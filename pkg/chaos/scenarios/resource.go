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

// Package scenarios provides chaos engineering experiment scenarios for VirtEngine.
//
// This file implements resource exhaustion scenarios for testing system resilience
// under constrained conditions. Supported stress scenarios include:
//
//   - CPU stress: saturate CPU with configurable load and workers
//   - Memory stress: consume memory with OOM kill testing support
//   - Disk stress: fill disk, inject I/O latency, or stress I/O subsystem
//   - Process stress: exhaust process table or file descriptors
//   - Clock skew: manipulate system time for temporal testing
//
// All scenarios implement the ResourceScenario interface and can be used with
// the chaos runner to execute experiments against target workloads.
package scenarios

import (
	"errors"
	"fmt"
	"time"
)

// Resource exhaustion experiment types.
const (
	// ExperimentTypeCPUStress indicates a CPU stress experiment.
	ExperimentTypeCPUStress ExperimentType = "cpu-stress"
	// ExperimentTypeMemoryStress indicates a memory stress experiment.
	ExperimentTypeMemoryStress ExperimentType = "memory-stress"
	// ExperimentTypeDiskStress indicates a disk stress experiment.
	ExperimentTypeDiskStress ExperimentType = "disk-stress"
	// ExperimentTypeProcessStress indicates a process stress experiment.
	ExperimentTypeProcessStress ExperimentType = "process-stress"
	// ExperimentTypeClockSkew indicates a clock skew experiment.
	ExperimentTypeClockSkew ExperimentType = "clock-skew"
)

// DiskStressMode defines the mode of disk stress operation.
type DiskStressMode string

const (
	// DiskStressModeFill fills disk space to a specified percentage.
	DiskStressModeFill DiskStressMode = "fill"
	// DiskStressModeIOStress creates I/O stress with concurrent workers.
	DiskStressModeIOStress DiskStressMode = "io-stress"
	// DiskStressModeLatency injects latency into disk I/O operations.
	DiskStressModeLatency DiskStressMode = "latency"
)

// GradualConfig defines configuration for scenarios that ramp up over time.
type GradualConfig struct {
	// StartValue is the initial stress value.
	StartValue int `json:"start_value"`

	// EndValue is the final stress value.
	EndValue int `json:"end_value"`

	// RampDuration is the time to transition from start to end value.
	RampDuration time.Duration `json:"ramp_duration"`

	// RatePerSecond is the rate of change (for leak simulations).
	RatePerSecond int `json:"rate_per_second,omitempty"`
}

// ResourceScenario is the interface that all resource stress scenarios must implement.
type ResourceScenario interface {
	// Name returns a unique identifier for this scenario.
	Name() string

	// Description returns a human-readable description of what this scenario does.
	Description() string

	// Type returns the experiment type category.
	Type() ExperimentType

	// Build constructs the Experiment configuration from the scenario settings.
	Build() (*Experiment, error)

	// Validate checks that the scenario configuration is valid.
	Validate() error
}

// -----------------------------------------------------------------------------
// CPU Stress Spec
// -----------------------------------------------------------------------------

// CPUStressSpec contains the configuration for a CPU stress experiment.
type CPUStressSpec struct {
	// Workers is the number of CPU workers to spawn.
	Workers int `json:"workers"`

	// Load is the target CPU load percentage (0-100).
	Load int `json:"load"`

	// Targets are the workload targets.
	Targets []string `json:"targets"`

	// Namespace is the Kubernetes namespace.
	Namespace string `json:"namespace"`

	// GradualConfig for ramping scenarios.
	GradualConfig *GradualConfig `json:"gradual_config,omitempty"`
}

// MemoryStressSpec contains the configuration for a memory stress experiment.
type MemoryStressSpec struct {
	// MemoryBytes is the absolute amount of memory to consume in bytes.
	MemoryBytes uint64 `json:"memory_bytes"`

	// MemoryPercent is the percentage of container memory limit to consume.
	MemoryPercent int `json:"memory_percent"`

	// OOMKillEnabled, when true, allows the process to be OOM killed.
	OOMKillEnabled bool `json:"oom_kill_enabled"`

	// Targets are the workload targets.
	Targets []string `json:"targets"`

	// Namespace is the Kubernetes namespace.
	Namespace string `json:"namespace"`

	// GradualConfig for leak scenarios.
	GradualConfig *GradualConfig `json:"gradual_config,omitempty"`
}

// DiskStressSpec contains the configuration for a disk stress experiment.
type DiskStressSpec struct {
	// Mode is the disk stress mode (fill, io-stress, latency).
	Mode DiskStressMode `json:"mode"`

	// FillPercent is the percentage of disk to fill (for fill mode).
	FillPercent int `json:"fill_percent,omitempty"`

	// Workers is the number of I/O workers (for io-stress mode).
	Workers int `json:"workers,omitempty"`

	// IOLatencyMs is the latency in milliseconds (for latency mode).
	IOLatencyMs int64 `json:"io_latency_ms,omitempty"`

	// Path is the filesystem path to stress.
	Path string `json:"path"`

	// Targets are the workload targets.
	Targets []string `json:"targets"`

	// Namespace is the Kubernetes namespace.
	Namespace string `json:"namespace"`
}

// ProcessStressSpec contains the configuration for a process stress experiment.
type ProcessStressSpec struct {
	// ProcessCount is the number of processes to spawn.
	ProcessCount int `json:"process_count"`

	// FileDescriptors is the number of file descriptors to consume.
	FileDescriptors int `json:"file_descriptors"`

	// Targets are the workload targets.
	Targets []string `json:"targets"`

	// Namespace is the Kubernetes namespace.
	Namespace string `json:"namespace"`
}

// ClockSkewSpec contains the configuration for a clock skew experiment.
type ClockSkewSpec struct {
	// OffsetMs is the time offset in milliseconds (can be negative).
	OffsetMs int64 `json:"offset_ms"`

	// JitterMs is random variation in milliseconds.
	JitterMs int64 `json:"jitter_ms"`

	// IsJump indicates this is a one-time jump rather than sustained skew.
	IsJump bool `json:"is_jump"`

	// Targets are the workload targets.
	Targets []string `json:"targets"`

	// Namespace is the Kubernetes namespace.
	Namespace string `json:"namespace"`
}

// -----------------------------------------------------------------------------
// CPUStressScenario
// -----------------------------------------------------------------------------

// CPUStressScenario configures a CPU stress experiment.
// It spawns workers that consume CPU cycles to simulate high CPU load conditions.
type CPUStressScenario struct {
	// Workers is the number of CPU workers to spawn.
	// If 0, defaults to the number of CPUs available on target.
	Workers int

	// Load is the target CPU load percentage (0-100).
	Load int

	// Duration is how long to apply the stress.
	Duration time.Duration

	// Targets are the workload targets (pod names, deployment names, etc.).
	Targets []string

	// Namespace is the Kubernetes namespace for the targets.
	Namespace string

	// gradual configuration for ramping scenarios
	gradualStart int
	gradualEnd   int
	isGradual    bool
}

// Name returns the scenario identifier.
func (s *CPUStressScenario) Name() string {
	if s.isGradual {
		return fmt.Sprintf("cpu-gradual-%d-to-%d", s.gradualStart, s.gradualEnd)
	}
	return fmt.Sprintf("cpu-stress-%d-pct", s.Load)
}

// Description returns a human-readable description.
func (s *CPUStressScenario) Description() string {
	if s.isGradual {
		return fmt.Sprintf("Gradually increase CPU load from %d%% to %d%% over %s",
			s.gradualStart, s.gradualEnd, s.Duration)
	}
	return fmt.Sprintf("Apply %d%% CPU load with %d workers for %s",
		s.Load, s.Workers, s.Duration)
}

// Type returns the experiment type.
func (s *CPUStressScenario) Type() ExperimentType {
	return ExperimentTypeCPUStress
}

// Validate checks that the scenario configuration is valid.
func (s *CPUStressScenario) Validate() error {
	if len(s.Targets) == 0 {
		return errors.New("cpu stress scenario: at least one target is required")
	}
	if s.Load < 0 || s.Load > 100 {
		return fmt.Errorf("cpu stress scenario: load must be 0-100, got %d", s.Load)
	}
	if s.Duration <= 0 {
		return errors.New("cpu stress scenario: duration must be positive")
	}
	if s.Workers < 0 {
		return errors.New("cpu stress scenario: workers cannot be negative")
	}
	if s.isGradual {
		if s.gradualStart < 0 || s.gradualStart > 100 {
			return fmt.Errorf("cpu stress scenario: gradual start load must be 0-100, got %d", s.gradualStart)
		}
		if s.gradualEnd < 0 || s.gradualEnd > 100 {
			return fmt.Errorf("cpu stress scenario: gradual end load must be 0-100, got %d", s.gradualEnd)
		}
	}
	return nil
}

// Build constructs the Experiment from this scenario.
func (s *CPUStressScenario) Build() (*Experiment, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}

	spec := CPUStressSpec{
		Workers:   s.Workers,
		Load:      s.Load,
		Targets:   s.Targets,
		Namespace: s.Namespace,
	}

	if s.isGradual {
		spec.GradualConfig = &GradualConfig{
			StartValue:   s.gradualStart,
			EndValue:     s.gradualEnd,
			RampDuration: s.Duration,
		}
	}

	return &Experiment{
		Name:        s.Name(),
		Description: s.Description(),
		Type:        s.Type(),
		Duration:    s.Duration,
		Targets:     s.Targets,
		Spec:        spec,
		Labels: map[string]string{
			"chaos.virtengine.io/type":     string(ExperimentTypeCPUStress),
			"chaos.virtengine.io/category": "resource",
		},
	}, nil
}

// NewCPUSaturation creates a CPU saturation scenario that applies constant load.
func NewCPUSaturation(targets []string, load int, duration time.Duration) *CPUStressScenario {
	return &CPUStressScenario{
		Workers:   0, // Auto-detect based on available CPUs
		Load:      load,
		Duration:  duration,
		Targets:   targets,
		Namespace: "default",
	}
}

// NewGradualCPUIncrease creates a CPU stress scenario that gradually increases load.
func NewGradualCPUIncrease(targets []string, startLoad, endLoad int, duration time.Duration) *CPUStressScenario {
	return &CPUStressScenario{
		Workers:      0,
		Load:         endLoad, // Final load for display
		Duration:     duration,
		Targets:      targets,
		Namespace:    "default",
		gradualStart: startLoad,
		gradualEnd:   endLoad,
		isGradual:    true,
	}
}

// -----------------------------------------------------------------------------
// MemoryStressScenario
// -----------------------------------------------------------------------------

// MemoryStressScenario configures a memory stress experiment.
// It allocates memory to simulate memory pressure and potential OOM conditions.
type MemoryStressScenario struct {
	// MemoryBytes is the absolute amount of memory to consume in bytes.
	MemoryBytes uint64

	// MemoryPercent is the percentage of container memory limit to consume (0-100).
	MemoryPercent int

	// OOMKillEnabled, when true, allows the process to be OOM killed.
	OOMKillEnabled bool

	// Duration is how long to maintain memory pressure.
	Duration time.Duration

	// Targets are the workload targets (pod names, deployment names, etc.).
	Targets []string

	// Namespace is the Kubernetes namespace for the targets.
	Namespace string

	// gradual configuration for leak scenarios
	leakRateMBPerSec int
	isLeak           bool
}

// Name returns the scenario identifier.
func (s *MemoryStressScenario) Name() string {
	if s.isLeak {
		return fmt.Sprintf("memory-leak-%dMB-per-sec", s.leakRateMBPerSec)
	}
	if s.OOMKillEnabled {
		return "memory-oom-kill-test"
	}
	if s.MemoryBytes > 0 {
		return fmt.Sprintf("memory-stress-%dMB", s.MemoryBytes/(1024*1024))
	}
	return fmt.Sprintf("memory-stress-%d-pct", s.MemoryPercent)
}

// Description returns a human-readable description.
func (s *MemoryStressScenario) Description() string {
	if s.isLeak {
		return fmt.Sprintf("Simulate memory leak at %d MB/sec for %s",
			s.leakRateMBPerSec, s.Duration)
	}
	if s.OOMKillEnabled {
		return "Consume memory until OOM kill is triggered"
	}
	if s.MemoryBytes > 0 {
		return fmt.Sprintf("Allocate %d MB of memory for %s",
			s.MemoryBytes/(1024*1024), s.Duration)
	}
	return fmt.Sprintf("Consume %d%% of container memory limit for %s",
		s.MemoryPercent, s.Duration)
}

// Type returns the experiment type.
func (s *MemoryStressScenario) Type() ExperimentType {
	return ExperimentTypeMemoryStress
}

// Validate checks that the scenario configuration is valid.
func (s *MemoryStressScenario) Validate() error {
	if len(s.Targets) == 0 {
		return errors.New("memory stress scenario: at least one target is required")
	}
	if s.MemoryBytes == 0 && s.MemoryPercent == 0 && !s.OOMKillEnabled && !s.isLeak {
		return errors.New("memory stress scenario: must specify MemoryBytes, MemoryPercent, OOMKillEnabled, or leak rate")
	}
	if s.MemoryPercent < 0 || s.MemoryPercent > 100 {
		return fmt.Errorf("memory stress scenario: MemoryPercent must be 0-100, got %d", s.MemoryPercent)
	}
	if s.Duration <= 0 && !s.OOMKillEnabled {
		return errors.New("memory stress scenario: duration must be positive")
	}
	if s.isLeak && s.leakRateMBPerSec <= 0 {
		return errors.New("memory stress scenario: leak rate must be positive")
	}
	return nil
}

// Build constructs the Experiment from this scenario.
func (s *MemoryStressScenario) Build() (*Experiment, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}

	spec := MemoryStressSpec{
		MemoryBytes:    s.MemoryBytes,
		MemoryPercent:  s.MemoryPercent,
		OOMKillEnabled: s.OOMKillEnabled,
		Targets:        s.Targets,
		Namespace:      s.Namespace,
	}

	if s.isLeak {
		spec.GradualConfig = &GradualConfig{
			StartValue: 0,
			//nolint:gosec // G115: MemoryBytes/MB is bounded memory size
			EndValue:      int(s.MemoryBytes / (1024 * 1024)),
			RampDuration:  s.Duration,
			RatePerSecond: s.leakRateMBPerSec,
		}
	}

	return &Experiment{
		Name:        s.Name(),
		Description: s.Description(),
		Type:        s.Type(),
		Duration:    s.Duration,
		Targets:     s.Targets,
		Spec:        spec,
		Labels: map[string]string{
			"chaos.virtengine.io/type":     string(ExperimentTypeMemoryStress),
			"chaos.virtengine.io/category": "resource",
		},
	}, nil
}

// NewMemoryPressure creates a memory pressure scenario with fixed allocation.
func NewMemoryPressure(targets []string, memoryMB int, duration time.Duration) *MemoryStressScenario {
	return &MemoryStressScenario{
		//nolint:gosec // G115: memoryMB is positive user-provided value
		MemoryBytes:    uint64(memoryMB) * 1024 * 1024,
		OOMKillEnabled: false,
		Duration:       duration,
		Targets:        targets,
		Namespace:      "default",
	}
}

// NewOOMKillTest creates a scenario that triggers OOM kill on target workloads.
func NewOOMKillTest(targets []string) *MemoryStressScenario {
	return &MemoryStressScenario{
		MemoryPercent:  100, // Consume all available memory
		OOMKillEnabled: true,
		Duration:       5 * time.Minute, // Max duration before forced cleanup
		Targets:        targets,
		Namespace:      "default",
	}
}

// NewGradualMemoryLeak creates a scenario simulating a gradual memory leak.
func NewGradualMemoryLeak(targets []string, leakRateMBPerSec int, duration time.Duration) *MemoryStressScenario {
	//nolint:gosec // G115: duration.Seconds() returns small positive value
	totalMB := leakRateMBPerSec * int(duration.Seconds())
	return &MemoryStressScenario{
		//nolint:gosec // G115: totalMB is positive bounded value
		MemoryBytes:      uint64(totalMB) * 1024 * 1024,
		OOMKillEnabled:   false,
		Duration:         duration,
		Targets:          targets,
		Namespace:        "default",
		leakRateMBPerSec: leakRateMBPerSec,
		isLeak:           true,
	}
}

// -----------------------------------------------------------------------------
// DiskStressScenario
// -----------------------------------------------------------------------------

// DiskStressScenario configures a disk stress experiment.
// It can fill disk space, create I/O stress, or inject I/O latency.
type DiskStressScenario struct {
	// Mode is the disk stress mode (fill, io-stress, latency).
	Mode DiskStressMode

	// FillPercent is the percentage of disk to fill (for fill mode).
	FillPercent int

	// Workers is the number of I/O workers (for io-stress mode).
	Workers int

	// IOLatency is the latency to inject into I/O operations (for latency mode).
	IOLatency time.Duration

	// Path is the filesystem path to stress.
	Path string

	// Duration is how long to apply the stress.
	Duration time.Duration

	// Targets are the workload targets (pod names, deployment names, etc.).
	Targets []string

	// Namespace is the Kubernetes namespace for the targets.
	Namespace string
}

// Name returns the scenario identifier.
func (s *DiskStressScenario) Name() string {
	switch s.Mode {
	case DiskStressModeFill:
		return fmt.Sprintf("disk-fill-%d-pct", s.FillPercent)
	case DiskStressModeIOStress:
		return fmt.Sprintf("disk-io-stress-%d-workers", s.Workers)
	case DiskStressModeLatency:
		return fmt.Sprintf("disk-latency-%s", s.IOLatency)
	default:
		return "disk-stress-unknown"
	}
}

// Description returns a human-readable description.
func (s *DiskStressScenario) Description() string {
	switch s.Mode {
	case DiskStressModeFill:
		return fmt.Sprintf("Fill disk at %s to %d%% capacity for %s",
			s.Path, s.FillPercent, s.Duration)
	case DiskStressModeIOStress:
		return fmt.Sprintf("Create disk I/O stress with %d concurrent workers at %s for %s",
			s.Workers, s.Path, s.Duration)
	case DiskStressModeLatency:
		return fmt.Sprintf("Inject %s latency into disk I/O at %s for %s",
			s.IOLatency, s.Path, s.Duration)
	default:
		return "Unknown disk stress mode"
	}
}

// Type returns the experiment type.
func (s *DiskStressScenario) Type() ExperimentType {
	return ExperimentTypeDiskStress
}

// Validate checks that the scenario configuration is valid.
func (s *DiskStressScenario) Validate() error {
	if len(s.Targets) == 0 {
		return errors.New("disk stress scenario: at least one target is required")
	}
	if s.Duration <= 0 {
		return errors.New("disk stress scenario: duration must be positive")
	}
	if s.Path == "" {
		return errors.New("disk stress scenario: path is required")
	}

	switch s.Mode {
	case DiskStressModeFill:
		if s.FillPercent < 1 || s.FillPercent > 100 {
			return fmt.Errorf("disk stress scenario: FillPercent must be 1-100, got %d", s.FillPercent)
		}
	case DiskStressModeIOStress:
		if s.Workers < 1 {
			return fmt.Errorf("disk stress scenario: Workers must be at least 1, got %d", s.Workers)
		}
	case DiskStressModeLatency:
		if s.IOLatency <= 0 {
			return errors.New("disk stress scenario: IOLatency must be positive")
		}
	default:
		return fmt.Errorf("disk stress scenario: unknown mode %q", s.Mode)
	}

	return nil
}

// Build constructs the Experiment from this scenario.
func (s *DiskStressScenario) Build() (*Experiment, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}

	spec := DiskStressSpec{
		Mode:      s.Mode,
		Path:      s.Path,
		Targets:   s.Targets,
		Namespace: s.Namespace,
	}

	switch s.Mode {
	case DiskStressModeFill:
		spec.FillPercent = s.FillPercent
	case DiskStressModeIOStress:
		spec.Workers = s.Workers
	case DiskStressModeLatency:
		spec.IOLatencyMs = s.IOLatency.Milliseconds()
	}

	return &Experiment{
		Name:        s.Name(),
		Description: s.Description(),
		Type:        s.Type(),
		Duration:    s.Duration,
		Targets:     s.Targets,
		Spec:        spec,
		Labels: map[string]string{
			"chaos.virtengine.io/type":     string(ExperimentTypeDiskStress),
			"chaos.virtengine.io/category": "resource",
		},
	}, nil
}

// NewDiskFill creates a disk fill scenario.
func NewDiskFill(targets []string, fillPercent int, duration time.Duration) *DiskStressScenario {
	return &DiskStressScenario{
		Mode:        DiskStressModeFill,
		FillPercent: fillPercent,
		Path:        "/tmp",
		Duration:    duration,
		Targets:     targets,
		Namespace:   "default",
	}
}

// NewDiskIOStress creates a disk I/O stress scenario.
func NewDiskIOStress(targets []string, workers int, duration time.Duration) *DiskStressScenario {
	return &DiskStressScenario{
		Mode:      DiskStressModeIOStress,
		Workers:   workers,
		Path:      "/tmp",
		Duration:  duration,
		Targets:   targets,
		Namespace: "default",
	}
}

// NewDiskLatency creates a disk I/O latency injection scenario.
func NewDiskLatency(targets []string, latency, duration time.Duration) *DiskStressScenario {
	return &DiskStressScenario{
		Mode:      DiskStressModeLatency,
		IOLatency: latency,
		Path:      "/",
		Duration:  duration,
		Targets:   targets,
		Namespace: "default",
	}
}

// -----------------------------------------------------------------------------
// ProcessStressScenario
// -----------------------------------------------------------------------------

// ProcessStressScenario configures a process/resource exhaustion experiment.
// It can exhaust the process table or file descriptors.
type ProcessStressScenario struct {
	// ProcessCount is the number of processes to spawn.
	ProcessCount int

	// FileDescriptors is the number of file descriptors to consume.
	FileDescriptors int

	// Duration is how long to maintain the stress.
	Duration time.Duration

	// Targets are the workload targets (pod names, deployment names, etc.).
	Targets []string

	// Namespace is the Kubernetes namespace for the targets.
	Namespace string
}

// Name returns the scenario identifier.
func (s *ProcessStressScenario) Name() string {
	if s.ProcessCount > 0 {
		return fmt.Sprintf("process-bomb-%d", s.ProcessCount)
	}
	return fmt.Sprintf("fd-exhaustion-%d", s.FileDescriptors)
}

// Description returns a human-readable description.
func (s *ProcessStressScenario) Description() string {
	if s.ProcessCount > 0 {
		return fmt.Sprintf("Spawn %d processes for %s to exhaust process table",
			s.ProcessCount, s.Duration)
	}
	return fmt.Sprintf("Consume %d file descriptors for %s",
		s.FileDescriptors, s.Duration)
}

// Type returns the experiment type.
func (s *ProcessStressScenario) Type() ExperimentType {
	return ExperimentTypeProcessStress
}

// Validate checks that the scenario configuration is valid.
func (s *ProcessStressScenario) Validate() error {
	if len(s.Targets) == 0 {
		return errors.New("process stress scenario: at least one target is required")
	}
	if s.Duration <= 0 {
		return errors.New("process stress scenario: duration must be positive")
	}
	if s.ProcessCount == 0 && s.FileDescriptors == 0 {
		return errors.New("process stress scenario: must specify ProcessCount or FileDescriptors")
	}
	if s.ProcessCount < 0 {
		return errors.New("process stress scenario: ProcessCount cannot be negative")
	}
	if s.FileDescriptors < 0 {
		return errors.New("process stress scenario: FileDescriptors cannot be negative")
	}
	return nil
}

// Build constructs the Experiment from this scenario.
func (s *ProcessStressScenario) Build() (*Experiment, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}

	return &Experiment{
		Name:        s.Name(),
		Description: s.Description(),
		Type:        s.Type(),
		Duration:    s.Duration,
		Spec:        s,
	}, nil
}

// NewProcessBomb creates a process bomb scenario that spawns many processes.
func NewProcessBomb(targets []string, count int, duration time.Duration) *ProcessStressScenario {
	return &ProcessStressScenario{
		ProcessCount: count,
		Duration:     duration,
		Targets:      targets,
		Namespace:    "default",
	}
}

// NewFDExhaustion creates a file descriptor exhaustion scenario.
func NewFDExhaustion(targets []string, fdCount int, duration time.Duration) *ProcessStressScenario {
	return &ProcessStressScenario{
		FileDescriptors: fdCount,
		Duration:        duration,
		Targets:         targets,
		Namespace:       "default",
	}
}

// -----------------------------------------------------------------------------
// ClockSkewScenario
// -----------------------------------------------------------------------------

// ClockSkewScenario configures a clock manipulation experiment.
// It can offset the system clock or add jitter to time queries.
type ClockSkewScenario struct {
	// Offset is the time offset to apply (can be negative for past, positive for future).
	Offset time.Duration

	// Jitter is random variation to add to each time query.
	Jitter time.Duration

	// Duration is how long to maintain the clock skew.
	Duration time.Duration

	// Targets are the workload targets (pod names, deployment names, etc.).
	Targets []string

	// Namespace is the Kubernetes namespace for the targets.
	Namespace string

	// isJump indicates this is a one-time jump rather than sustained skew
	isJump bool
}

// Name returns the scenario identifier.
func (s *ClockSkewScenario) Name() string {
	if s.isJump {
		return fmt.Sprintf("time-jump-%s", formatDuration(s.Offset))
	}
	if s.Jitter > 0 {
		return fmt.Sprintf("clock-jitter-%s", s.Jitter)
	}
	return fmt.Sprintf("clock-skew-%s", formatDuration(s.Offset))
}

// Description returns a human-readable description.
func (s *ClockSkewScenario) Description() string {
	if s.isJump {
		if s.Offset >= 0 {
			return fmt.Sprintf("Jump system clock forward by %s", s.Offset)
		}
		return fmt.Sprintf("Jump system clock backward by %s", -s.Offset)
	}
	if s.Jitter > 0 && s.Offset != 0 {
		return fmt.Sprintf("Skew clock by %s with ±%s jitter for %s",
			formatDuration(s.Offset), s.Jitter, s.Duration)
	}
	if s.Jitter > 0 {
		return fmt.Sprintf("Add ±%s jitter to clock queries for %s",
			s.Jitter, s.Duration)
	}
	return fmt.Sprintf("Skew system clock by %s for %s",
		formatDuration(s.Offset), s.Duration)
}

// Type returns the experiment type.
func (s *ClockSkewScenario) Type() ExperimentType {
	return ExperimentTypeClockSkew
}

// Validate checks that the scenario configuration is valid.
func (s *ClockSkewScenario) Validate() error {
	if len(s.Targets) == 0 {
		return errors.New("clock skew scenario: at least one target is required")
	}
	if s.Offset == 0 && s.Jitter == 0 {
		return errors.New("clock skew scenario: must specify Offset or Jitter")
	}
	if !s.isJump && s.Duration <= 0 {
		return errors.New("clock skew scenario: duration must be positive for sustained skew")
	}
	if s.Jitter < 0 {
		return errors.New("clock skew scenario: Jitter cannot be negative")
	}
	return nil
}

// Build constructs the Experiment from this scenario.
func (s *ClockSkewScenario) Build() (*Experiment, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}

	return &Experiment{
		Name:        s.Name(),
		Description: s.Description(),
		Type:        s.Type(),
		Duration:    s.Duration,
		Spec:        s,
	}, nil
}

// NewClockSkew creates a sustained clock skew scenario.
func NewClockSkew(targets []string, offset, duration time.Duration) *ClockSkewScenario {
	return &ClockSkewScenario{
		Offset:    offset,
		Duration:  duration,
		Targets:   targets,
		Namespace: "default",
	}
}

// NewTimeJump creates a one-time clock jump scenario.
func NewTimeJump(targets []string, jumpAmount time.Duration) *ClockSkewScenario {
	return &ClockSkewScenario{
		Offset:    jumpAmount,
		Duration:  time.Minute, // Cleanup after 1 minute
		Targets:   targets,
		Namespace: "default",
		isJump:    true,
	}
}

// -----------------------------------------------------------------------------
// Default Scenarios
// -----------------------------------------------------------------------------

// DefaultResourceScenarios returns a collection of commonly used resource
// exhaustion scenarios for chaos testing.
func DefaultResourceScenarios() []ResourceScenario {
	defaultTargets := []string{"*"} // Target all pods in namespace
	defaultDuration := 5 * time.Minute

	return []ResourceScenario{
		// CPU Scenarios
		NewCPUSaturation(defaultTargets, 50, defaultDuration),
		NewCPUSaturation(defaultTargets, 80, defaultDuration),
		NewCPUSaturation(defaultTargets, 100, defaultDuration),
		NewGradualCPUIncrease(defaultTargets, 10, 90, defaultDuration),

		// Memory Scenarios
		NewMemoryPressure(defaultTargets, 256, defaultDuration),
		NewMemoryPressure(defaultTargets, 512, defaultDuration),
		NewOOMKillTest(defaultTargets),
		NewGradualMemoryLeak(defaultTargets, 10, defaultDuration),

		// Disk Scenarios
		NewDiskFill(defaultTargets, 50, defaultDuration),
		NewDiskFill(defaultTargets, 90, defaultDuration),
		NewDiskIOStress(defaultTargets, 4, defaultDuration),
		NewDiskIOStress(defaultTargets, 16, defaultDuration),
		NewDiskLatency(defaultTargets, 100*time.Millisecond, defaultDuration),
		NewDiskLatency(defaultTargets, 500*time.Millisecond, defaultDuration),

		// Process Scenarios
		NewProcessBomb(defaultTargets, 100, defaultDuration),
		NewProcessBomb(defaultTargets, 1000, defaultDuration),
		NewFDExhaustion(defaultTargets, 1000, defaultDuration),
		NewFDExhaustion(defaultTargets, 10000, defaultDuration),

		// Clock Scenarios
		NewClockSkew(defaultTargets, 5*time.Second, defaultDuration),
		NewClockSkew(defaultTargets, -5*time.Second, defaultDuration),
		NewClockSkew(defaultTargets, time.Hour, defaultDuration),
		NewTimeJump(defaultTargets, 24*time.Hour),
		NewTimeJump(defaultTargets, -24*time.Hour),
	}
}

// formatDuration formats a duration in a human-friendly way, handling negative values.
func formatDuration(d time.Duration) string {
	if d < 0 {
		return "-" + (-d).String()
	}
	return d.String()
}

// -----------------------------------------------------------------------------
// Scenario Registration
// -----------------------------------------------------------------------------

// ScenarioRegistry provides a registry for looking up scenarios by name.
type ScenarioRegistry struct {
	scenarios map[string]func() ResourceScenario
}

// NewScenarioRegistry creates a new scenario registry with default scenarios.
func NewScenarioRegistry() *ScenarioRegistry {
	r := &ScenarioRegistry{
		scenarios: make(map[string]func() ResourceScenario),
	}
	r.RegisterDefaults()
	return r
}

// Register adds a scenario factory to the registry.
func (r *ScenarioRegistry) Register(name string, factory func() ResourceScenario) {
	r.scenarios[name] = factory
}

// Get retrieves a scenario by name.
func (r *ScenarioRegistry) Get(name string) (ResourceScenario, bool) {
	factory, ok := r.scenarios[name]
	if !ok {
		return nil, false
	}
	return factory(), true
}

// List returns all registered scenario names.
func (r *ScenarioRegistry) List() []string {
	names := make([]string, 0, len(r.scenarios))
	for name := range r.scenarios {
		names = append(names, name)
	}
	return names
}

// RegisterDefaults registers the default resource scenarios.
func (r *ScenarioRegistry) RegisterDefaults() {
	// CPU scenarios
	r.Register("cpu-saturation-50", func() ResourceScenario {
		return NewCPUSaturation([]string{"*"}, 50, 5*time.Minute)
	})
	r.Register("cpu-saturation-80", func() ResourceScenario {
		return NewCPUSaturation([]string{"*"}, 80, 5*time.Minute)
	})
	r.Register("cpu-saturation-100", func() ResourceScenario {
		return NewCPUSaturation([]string{"*"}, 100, 5*time.Minute)
	})
	r.Register("cpu-gradual", func() ResourceScenario {
		return NewGradualCPUIncrease([]string{"*"}, 10, 90, 5*time.Minute)
	})

	// Memory scenarios
	r.Register("memory-pressure-256MB", func() ResourceScenario {
		return NewMemoryPressure([]string{"*"}, 256, 5*time.Minute)
	})
	r.Register("memory-pressure-512MB", func() ResourceScenario {
		return NewMemoryPressure([]string{"*"}, 512, 5*time.Minute)
	})
	r.Register("memory-oom-kill", func() ResourceScenario {
		return NewOOMKillTest([]string{"*"})
	})
	r.Register("memory-leak", func() ResourceScenario {
		return NewGradualMemoryLeak([]string{"*"}, 10, 5*time.Minute)
	})

	// Disk scenarios
	r.Register("disk-fill-50", func() ResourceScenario {
		return NewDiskFill([]string{"*"}, 50, 5*time.Minute)
	})
	r.Register("disk-fill-90", func() ResourceScenario {
		return NewDiskFill([]string{"*"}, 90, 5*time.Minute)
	})
	r.Register("disk-io-stress", func() ResourceScenario {
		return NewDiskIOStress([]string{"*"}, 8, 5*time.Minute)
	})
	r.Register("disk-latency-100ms", func() ResourceScenario {
		return NewDiskLatency([]string{"*"}, 100*time.Millisecond, 5*time.Minute)
	})

	// Process scenarios
	r.Register("process-bomb-100", func() ResourceScenario {
		return NewProcessBomb([]string{"*"}, 100, 5*time.Minute)
	})
	r.Register("fd-exhaustion-1000", func() ResourceScenario {
		return NewFDExhaustion([]string{"*"}, 1000, 5*time.Minute)
	})

	// Clock scenarios
	r.Register("clock-skew-5s", func() ResourceScenario {
		return NewClockSkew([]string{"*"}, 5*time.Second, 5*time.Minute)
	})
	r.Register("clock-skew-1h", func() ResourceScenario {
		return NewClockSkew([]string{"*"}, time.Hour, 5*time.Minute)
	})
	r.Register("time-jump-24h", func() ResourceScenario {
		return NewTimeJump([]string{"*"}, 24*time.Hour)
	})
}

