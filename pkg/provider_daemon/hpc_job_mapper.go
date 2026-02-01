// Package provider_daemon implements the VirtEngine provider daemon.
//
// VE-4D: Job spec mapper - maps x/hpc job specs to scheduler-specific specs
package provider_daemon

import (
	"fmt"
	"strings"

	"github.com/virtengine/virtengine/pkg/moab_adapter"
	"github.com/virtengine/virtengine/pkg/ood_adapter"
	"github.com/virtengine/virtengine/pkg/slurm_adapter"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// HPCJobMapper maps x/hpc job specs to scheduler-specific specs
type HPCJobMapper struct {
	schedulerType HPCSchedulerType
	clusterID     string

	// Scheduler-specific defaults
	slurmDefaults slurm_adapter.SLURMConfig
	moabDefaults  moab_adapter.MOABConfig
	oodDefaults   ood_adapter.OODConfig
}

// NewHPCJobMapper creates a new job mapper
func NewHPCJobMapper(schedulerType HPCSchedulerType, clusterID string) *HPCJobMapper {
	return &HPCJobMapper{
		schedulerType: schedulerType,
		clusterID:     clusterID,
		slurmDefaults: slurm_adapter.DefaultSLURMConfig(),
		moabDefaults:  moab_adapter.DefaultMOABConfig(),
		oodDefaults:   ood_adapter.DefaultOODConfig(),
	}
}

// SetSLURMDefaults sets SLURM configuration defaults
func (m *HPCJobMapper) SetSLURMDefaults(config slurm_adapter.SLURMConfig) {
	m.slurmDefaults = config
}

// SetMOABDefaults sets MOAB configuration defaults
func (m *HPCJobMapper) SetMOABDefaults(config moab_adapter.MOABConfig) {
	m.moabDefaults = config
}

// SetOODDefaults sets OOD configuration defaults
func (m *HPCJobMapper) SetOODDefaults(config ood_adapter.OODConfig) {
	m.oodDefaults = config
}

// MapToSLURM maps an x/hpc HPCJob to a SLURM job spec
func (m *HPCJobMapper) MapToSLURM(job *hpctypes.HPCJob) (*slurm_adapter.SLURMJobSpec, error) {
	if job == nil {
		return nil, fmt.Errorf("job cannot be nil")
	}

	if err := job.Validate(); err != nil {
		return nil, fmt.Errorf("invalid job: %w", err)
	}

	// Calculate time limit in minutes
	timeLimitMinutes := job.MaxRuntimeSeconds / 60
	if timeLimitMinutes < 1 {
		timeLimitMinutes = 1
	}

	// Calculate memory in MB
	memoryMB := int64(job.Resources.MemoryGBPerNode) * 1024

	spec := &slurm_adapter.SLURMJobSpec{
		JobName:          m.generateJobName(job),
		Partition:        m.mapQueueToPartition(job.QueueName),
		Nodes:            job.Resources.Nodes,
		CPUsPerNode:      job.Resources.CPUCoresPerNode,
		MemoryMB:         memoryMB,
		GPUs:             job.Resources.GPUsPerNode,
		GPUType:          job.Resources.GPUType,
		TimeLimit:        timeLimitMinutes,
		WorkingDirectory: job.WorkloadSpec.WorkingDirectory,
		ContainerImage:   job.WorkloadSpec.ContainerImage,
		Command:          job.WorkloadSpec.Command,
		Arguments:        job.WorkloadSpec.Arguments,
		Environment:      m.buildEnvironment(job),
		InputFiles:       m.extractInputFiles(job),
		OutputDirectory:  m.generateOutputDirectory(job),
		Exclusive:        m.shouldBeExclusive(job),
		Constraints:      m.buildConstraints(job),
	}

	// Use default partition if not specified
	if spec.Partition == "" {
		spec.Partition = m.slurmDefaults.DefaultPartition
	}

	return spec, nil
}

// MapToMOAB maps an x/hpc HPCJob to a MOAB job spec
func (m *HPCJobMapper) MapToMOAB(job *hpctypes.HPCJob) (*moab_adapter.MOABJobSpec, error) {
	if job == nil {
		return nil, fmt.Errorf("job cannot be nil")
	}

	if err := job.Validate(); err != nil {
		return nil, fmt.Errorf("invalid job: %w", err)
	}

	// Calculate memory in MB
	memoryMB := int64(job.Resources.MemoryGBPerNode) * 1024

	spec := &moab_adapter.MOABJobSpec{
		JobName:          m.generateJobName(job),
		Queue:            m.mapQueueToMOABQueue(job.QueueName),
		Account:          m.moabDefaults.DefaultAccount,
		Nodes:            job.Resources.Nodes,
		ProcsPerNode:     job.Resources.CPUCoresPerNode,
		MemoryMB:         memoryMB,
		GPUs:             job.Resources.GPUsPerNode,
		GPUType:          job.Resources.GPUType,
		WallTimeLimit:    job.MaxRuntimeSeconds,
		WorkingDirectory: job.WorkloadSpec.WorkingDirectory,
		Executable:       m.buildExecutable(job),
		Arguments:        job.WorkloadSpec.Arguments,
		Environment:      m.buildEnvironment(job),
		InputFiles:       m.extractInputFiles(job),
		OutputFile:       m.generateOutputPath(job, "stdout"),
		ErrorFile:        m.generateOutputPath(job, "stderr"),
		Features:         m.buildFeatures(job),
	}

	// Use default queue if not specified
	if spec.Queue == "" {
		spec.Queue = m.moabDefaults.DefaultQueue
	}

	return spec, nil
}

// MapToOOD maps an x/hpc HPCJob to an OOD interactive app spec
// Note: OOD is typically for interactive sessions, so this maps batch jobs
// to job composer submissions
func (m *HPCJobMapper) MapToOOD(job *hpctypes.HPCJob) (*ood_adapter.InteractiveAppSpec, error) {
	if job == nil {
		return nil, fmt.Errorf("job cannot be nil")
	}

	if err := job.Validate(); err != nil {
		return nil, fmt.Errorf("invalid job: %w", err)
	}

	// Calculate hours from seconds
	hours := int32(job.MaxRuntimeSeconds / 3600)
	if hours < 1 {
		hours = 1
	}

	// Determine app type based on workload
	appType := m.determineOODAppType(job)

	spec := &ood_adapter.InteractiveAppSpec{
		AppType:    appType,
		AppVersion: "",
		Resources: &ood_adapter.SessionResources{
			CPUs:      job.Resources.CPUCoresPerNode * job.Resources.Nodes,
			MemoryGB:  job.Resources.MemoryGBPerNode,
			GPUs:      job.Resources.GPUsPerNode,
			GPUType:   job.Resources.GPUType,
			Hours:     hours,
			Partition: m.mapQueueToPartition(job.QueueName),
		},
		Environment:      m.buildEnvironment(job),
		WorkingDirectory: job.WorkloadSpec.WorkingDirectory,
	}

	// Add Jupyter-specific options if applicable
	if appType == ood_adapter.AppTypeJupyter {
		spec.JupyterOptions = &ood_adapter.JupyterOptions{
			EnableGPU: job.Resources.GPUsPerNode > 0,
		}
	}

	return spec, nil
}

// MapSLURMState maps SLURM job state to unified HPC job state
func MapSLURMState(state slurm_adapter.SLURMJobState) HPCJobState {
	switch state {
	case slurm_adapter.SLURMJobStatePending:
		return HPCJobStateQueued
	case slurm_adapter.SLURMJobStateRunning:
		return HPCJobStateRunning
	case slurm_adapter.SLURMJobStateCompleted:
		return HPCJobStateCompleted
	case slurm_adapter.SLURMJobStateFailed:
		return HPCJobStateFailed
	case slurm_adapter.SLURMJobStateCancelled:
		return HPCJobStateCancelled
	case slurm_adapter.SLURMJobStateTimeout:
		return HPCJobStateTimeout
	case slurm_adapter.SLURMJobStateSuspended:
		return HPCJobStateSuspended
	default:
		return HPCJobStatePending
	}
}

// MapMOABState maps MOAB job state to unified HPC job state
func MapMOABState(state moab_adapter.MOABJobState) HPCJobState {
	switch state {
	case moab_adapter.MOABJobStateIdle:
		return HPCJobStateQueued
	case moab_adapter.MOABJobStateStarting:
		return HPCJobStateStarting
	case moab_adapter.MOABJobStateRunning:
		return HPCJobStateRunning
	case moab_adapter.MOABJobStateCompleted:
		return HPCJobStateCompleted
	case moab_adapter.MOABJobStateFailed:
		return HPCJobStateFailed
	case moab_adapter.MOABJobStateCancelled:
		return HPCJobStateCancelled
	case moab_adapter.MOABJobStateHold:
		return HPCJobStateSuspended
	case moab_adapter.MOABJobStateSuspended:
		return HPCJobStateSuspended
	case moab_adapter.MOABJobStateRemoved, moab_adapter.MOABJobStateVacated:
		return HPCJobStateCancelled
	case moab_adapter.MOABJobStateDeferred:
		return HPCJobStateQueued
	default:
		return HPCJobStatePending
	}
}

// MapOODState maps OOD session state to unified HPC job state
func MapOODState(state ood_adapter.SessionState) HPCJobState {
	switch state {
	case ood_adapter.SessionStatePending:
		return HPCJobStatePending
	case ood_adapter.SessionStateStarting:
		return HPCJobStateStarting
	case ood_adapter.SessionStateRunning:
		return HPCJobStateRunning
	case ood_adapter.SessionStateSuspended:
		return HPCJobStateSuspended
	case ood_adapter.SessionStateCompleted:
		return HPCJobStateCompleted
	case ood_adapter.SessionStateFailed:
		return HPCJobStateFailed
	case ood_adapter.SessionStateCancelled:
		return HPCJobStateCancelled
	default:
		return HPCJobStatePending
	}
}

// MapSLURMMetrics maps SLURM usage metrics to unified metrics
func MapSLURMMetrics(metrics *slurm_adapter.SLURMUsageMetrics, nodes int32) *HPCSchedulerMetrics {
	if metrics == nil {
		return nil
	}

	nodeHours := float64(metrics.WallClockSeconds) * float64(nodes) / 3600.0

	return &HPCSchedulerMetrics{
		WallClockSeconds: metrics.WallClockSeconds,
		CPUTimeSeconds:   metrics.CPUTimeSeconds,
		CPUCoreSeconds:   metrics.CPUTimeSeconds, // Approximation
		MemoryBytesMax:   metrics.MaxRSSBytes,
		MemoryGBSeconds:  (metrics.MaxRSSBytes / (1024 * 1024 * 1024)) * metrics.WallClockSeconds,
		GPUSeconds:       metrics.GPUSeconds,
		NodesUsed:        nodes,
		NodeHours:        nodeHours,
		SchedulerSpecific: map[string]interface{}{
			"max_vm_size_bytes": metrics.MaxVMSizeBytes,
		},
	}
}

// MapMOABMetrics maps MOAB usage metrics to unified metrics
func MapMOABMetrics(metrics *moab_adapter.MOABUsageMetrics) *HPCSchedulerMetrics {
	if metrics == nil {
		return nil
	}

	return &HPCSchedulerMetrics{
		WallClockSeconds: metrics.WallClockSeconds,
		CPUTimeSeconds:   metrics.CPUTimeSeconds,
		CPUCoreSeconds:   metrics.CPUTimeSeconds, // Approximation
		MemoryBytesMax:   metrics.MaxRSSBytes,
		MemoryGBSeconds:  (metrics.MaxRSSBytes / (1024 * 1024 * 1024)) * metrics.WallClockSeconds,
		GPUSeconds:       metrics.GPUSeconds,
		NodeHours:        metrics.NodeHours,
		EnergyJoules:     metrics.EnergyJoules,
		SchedulerSpecific: map[string]interface{}{
			"sus_used":          metrics.SUSUsed,
			"max_vm_size_bytes": metrics.MaxVMSizeBytes,
		},
	}
}

// MapOODMetrics maps OOD session metrics to unified metrics
func MapOODMetrics(metrics *ood_adapter.SessionUsageMetrics) *HPCSchedulerMetrics {
	if metrics == nil {
		return nil
	}

	return &HPCSchedulerMetrics{
		WallClockSeconds: metrics.WallClockSeconds,
		CPUTimeSeconds:   metrics.CPUTimeSeconds,
		MemoryBytesMax:   metrics.MemoryBytesAvg, // Using avg as max approximation
		GPUSeconds:       metrics.GPUSeconds,
	}
}

// Helper methods

func (m *HPCJobMapper) generateJobName(job *hpctypes.HPCJob) string {
	// Format: ve-<cluster>-<job_id_prefix>
	jobIDPrefix := job.JobID
	if len(jobIDPrefix) > 8 {
		jobIDPrefix = jobIDPrefix[:8]
	}
	return fmt.Sprintf("ve-%s-%s", m.clusterID[:min(8, len(m.clusterID))], jobIDPrefix)
}

func (m *HPCJobMapper) mapQueueToPartition(queue string) string {
	if queue == "" {
		return ""
	}
	// Queue names map directly to partitions in most cases
	return queue
}

func (m *HPCJobMapper) mapQueueToMOABQueue(queue string) string {
	if queue == "" {
		return ""
	}
	// Queue names map directly to MOAB queues
	return queue
}

func (m *HPCJobMapper) buildEnvironment(job *hpctypes.HPCJob) map[string]string {
	env := make(map[string]string)

	// Copy workload environment
	for k, v := range job.WorkloadSpec.Environment {
		env[k] = v
	}

	// Add VirtEngine-specific environment variables
	env["VIRTENGINE_JOB_ID"] = job.JobID
	env["VIRTENGINE_CLUSTER_ID"] = job.ClusterID
	env["VIRTENGINE_OFFERING_ID"] = job.OfferingID

	return env
}

func (m *HPCJobMapper) extractInputFiles(job *hpctypes.HPCJob) []string {
	var files []string
	for _, ref := range job.DataReferences {
		if !ref.Encrypted {
			files = append(files, ref.URI)
		}
	}
	return files
}

func (m *HPCJobMapper) generateOutputDirectory(job *hpctypes.HPCJob) string {
	return fmt.Sprintf("/scratch/virtengine/%s/%s", m.clusterID, job.JobID)
}

func (m *HPCJobMapper) generateOutputPath(job *hpctypes.HPCJob, suffix string) string {
	return fmt.Sprintf("%s/%s.%s", m.generateOutputDirectory(job), job.JobID, suffix)
}

func (m *HPCJobMapper) shouldBeExclusive(job *hpctypes.HPCJob) bool {
	// Request exclusive access for GPU jobs or large resource requests
	if job.Resources.GPUsPerNode > 0 {
		return true
	}
	if job.Resources.Nodes > 1 {
		return true
	}
	return false
}

func (m *HPCJobMapper) buildConstraints(job *hpctypes.HPCJob) []string {
	var constraints []string

	// Add GPU type constraint if specified
	if job.Resources.GPUType != "" {
		constraints = append(constraints, fmt.Sprintf("gpu:%s", job.Resources.GPUType))
	}

	return constraints
}

func (m *HPCJobMapper) buildFeatures(job *hpctypes.HPCJob) []string {
	var features []string

	// Add GPU type as a feature
	if job.Resources.GPUType != "" {
		features = append(features, job.Resources.GPUType)
	}

	return features
}

func (m *HPCJobMapper) buildExecutable(job *hpctypes.HPCJob) string {
	// For containerized workloads, use singularity/apptainer
	if job.WorkloadSpec.ContainerImage != "" {
		return fmt.Sprintf("singularity exec %s %s",
			job.WorkloadSpec.ContainerImage,
			job.WorkloadSpec.Command)
	}
	return job.WorkloadSpec.Command
}

func (m *HPCJobMapper) determineOODAppType(job *hpctypes.HPCJob) ood_adapter.InteractiveAppType {
	command := strings.ToLower(job.WorkloadSpec.Command)
	image := strings.ToLower(job.WorkloadSpec.ContainerImage)

	// Check for known app types
	if strings.Contains(command, "jupyter") || strings.Contains(image, "jupyter") {
		return ood_adapter.AppTypeJupyter
	}
	if strings.Contains(command, "rstudio") || strings.Contains(image, "rstudio") {
		return ood_adapter.AppTypeRStudio
	}
	if strings.Contains(command, "vscode") || strings.Contains(image, "vscode") {
		return ood_adapter.AppTypeVSCode
	}
	if strings.Contains(command, "matlab") || strings.Contains(image, "matlab") {
		return ood_adapter.AppTypeMatlab
	}

	// Default to custom app type for batch jobs
	return ood_adapter.AppTypeCustom
}

