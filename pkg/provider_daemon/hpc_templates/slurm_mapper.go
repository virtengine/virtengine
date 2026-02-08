// Package hpc_templates provides workload template resolution for HPC jobs.
//
// VE-5F: SLURM job mapper - resolves templates to SLURM batch scripts
package hpc_templates

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"

	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// SlurmJobMapper maps WorkloadTemplate to SLURM batch scripts
type SlurmJobMapper struct {
	// DefaultPartition is the default SLURM partition
	DefaultPartition string

	// DefaultQoS is the default QoS policy
	DefaultQoS string

	// ModulePath is the path to module files
	ModulePath string
}

// NewSlurmJobMapper creates a new SLURM job mapper
func NewSlurmJobMapper(partition, qos, modulePath string) *SlurmJobMapper {
	return &SlurmJobMapper{
		DefaultPartition: partition,
		DefaultQoS:       qos,
		ModulePath:       modulePath,
	}
}

// MapToSlurmJob converts a WorkloadTemplate to a SLURM batch script
func (m *SlurmJobMapper) MapToSlurmJob(
	wt *hpctypes.WorkloadTemplate,
	userParams *UserParameters,
) (*ResolvedJob, error) {
	// Validate template
	if err := wt.Validate(); err != nil {
		return nil, fmt.Errorf("invalid template: %w", err)
	}

	// Resolve resources
	resources, err := m.resolveResources(wt, userParams)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve resources: %w", err)
	}

	// Resolve environment
	environment, err := m.resolveEnvironment(wt, userParams)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve environment: %w", err)
	}

	// Resolve data bindings
	dataBindings, err := m.resolveDataBindings(wt, userParams)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve data bindings: %w", err)
	}

	// Generate SLURM script
	script, err := m.generateSlurmScript(wt, resources, environment, dataBindings, userParams)
	if err != nil {
		return nil, fmt.Errorf("failed to generate SLURM script: %w", err)
	}

	return &ResolvedJob{
		TemplateID:      wt.TemplateID,
		TemplateVersion: wt.Version,
		JobType:         "slurm",
		SlurmScript:     script,
		Resources:       *resources,
		Environment:     environment,
		DataBindings:    dataBindings,
		ResolvedAt:      time.Now().UTC(),
	}, nil
}

// resolveResources resolves resource requirements from template and user overrides
func (m *SlurmJobMapper) resolveResources(
	wt *hpctypes.WorkloadTemplate,
	userParams *UserParameters,
) (*ResolvedResources, error) {
	res := &ResolvedResources{
		Nodes:           wt.Resources.DefaultNodes,
		CPUsPerNode:     wt.Resources.DefaultCPUsPerNode,
		MemoryMBPerNode: wt.Resources.DefaultMemoryMBPerNode,
		GPUsPerNode:     wt.Resources.DefaultGPUsPerNode,
		GPUTypes:        wt.Resources.GPUTypes,
		RuntimeMinutes:  wt.Resources.DefaultRuntimeMinutes,
		StorageGB:       wt.Resources.StorageGBRequired,
		ExclusiveNodes:  wt.Resources.ExclusiveNodes,
		NetworkRequired: wt.Resources.NetworkRequired,
	}

	// Apply user overrides if provided
	if userParams != nil && userParams.Resources != nil {
		if err := m.applyResourceOverrides(res, wt, userParams.Resources); err != nil {
			return nil, err
		}
	}

	// Validate final resources
	if err := m.validateResources(res, wt); err != nil {
		return nil, err
	}

	return res, nil
}

// applyResourceOverrides applies user resource overrides within template limits
func (m *SlurmJobMapper) applyResourceOverrides(
	res *ResolvedResources,
	wt *hpctypes.WorkloadTemplate,
	overrides *UserResourceOverrides,
) error {
	if overrides.Nodes != nil {
		nodes := *overrides.Nodes
		if nodes < wt.Resources.MinNodes || nodes > wt.Resources.MaxNodes {
			return fmt.Errorf("nodes %d out of range [%d, %d]", nodes, wt.Resources.MinNodes, wt.Resources.MaxNodes)
		}
		res.Nodes = nodes
	}

	if overrides.CPUsPerNode != nil {
		cpus := *overrides.CPUsPerNode
		if cpus < wt.Resources.MinCPUsPerNode || cpus > wt.Resources.MaxCPUsPerNode {
			return fmt.Errorf("CPUs per node %d out of range [%d, %d]", cpus, wt.Resources.MinCPUsPerNode, wt.Resources.MaxCPUsPerNode)
		}
		res.CPUsPerNode = cpus
	}

	if overrides.MemoryMBPerNode != nil {
		mem := *overrides.MemoryMBPerNode
		if mem < wt.Resources.MinMemoryMBPerNode || mem > wt.Resources.MaxMemoryMBPerNode {
			return fmt.Errorf("memory per node %d out of range [%d, %d]", mem, wt.Resources.MinMemoryMBPerNode, wt.Resources.MaxMemoryMBPerNode)
		}
		res.MemoryMBPerNode = mem
	}

	if overrides.GPUsPerNode != nil {
		gpus := *overrides.GPUsPerNode
		if gpus < wt.Resources.MinGPUsPerNode || gpus > wt.Resources.MaxGPUsPerNode {
			return fmt.Errorf("GPUs per node %d out of range [%d, %d]", gpus, wt.Resources.MinGPUsPerNode, wt.Resources.MaxGPUsPerNode)
		}
		res.GPUsPerNode = gpus
	}

	if overrides.RuntimeMinutes != nil {
		runtime := *overrides.RuntimeMinutes
		if runtime < wt.Resources.MinRuntimeMinutes || runtime > wt.Resources.MaxRuntimeMinutes {
			return fmt.Errorf("runtime %d out of range [%d, %d]", runtime, wt.Resources.MinRuntimeMinutes, wt.Resources.MaxRuntimeMinutes)
		}
		res.RuntimeMinutes = runtime
	}

	return nil
}

// validateResources validates final resource requirements
func (m *SlurmJobMapper) validateResources(res *ResolvedResources, _ *hpctypes.WorkloadTemplate) error {
	if res.Nodes < 1 {
		return fmt.Errorf("nodes must be >= 1")
	}
	if res.CPUsPerNode < 1 {
		return fmt.Errorf("CPUs per node must be >= 1")
	}
	if res.MemoryMBPerNode < 1 {
		return fmt.Errorf("memory per node must be >= 1")
	}
	if res.RuntimeMinutes < 1 {
		return fmt.Errorf("runtime must be >= 1 minute")
	}
	return nil
}

// resolveEnvironment resolves environment variables
func (m *SlurmJobMapper) resolveEnvironment(
	wt *hpctypes.WorkloadTemplate,
	userParams *UserParameters,
) (map[string]string, error) {
	env := make(map[string]string)

	// Add template environment variables
	for _, envVar := range wt.Environment {
		value := envVar.Value

		// If value template is provided, substitute parameters
		if envVar.ValueTemplate != "" {
			substituted, err := m.substituteTemplate(envVar.ValueTemplate, userParams)
			if err != nil {
				return nil, fmt.Errorf("failed to substitute env var %s: %w", envVar.Name, err)
			}
			value = substituted
		}

		// Check if required but not provided
		if envVar.Required && value == "" {
			if userParams == nil || userParams.Parameters[envVar.Name] == "" {
				return nil, fmt.Errorf("required environment variable %s not provided", envVar.Name)
			}
			value = userParams.Parameters[envVar.Name]
		}

		env[envVar.Name] = value
	}

	// Add user-provided parameters not in template
	if userParams != nil {
		for key, value := range userParams.Parameters {
			if _, exists := env[key]; !exists {
				env[key] = value
			}
		}
	}

	return env, nil
}

// resolveDataBindings resolves data mount points
func (m *SlurmJobMapper) resolveDataBindings(
	wt *hpctypes.WorkloadTemplate,
	userParams *UserParameters,
) ([]ResolvedDataBinding, error) {
	var bindings []ResolvedDataBinding

	for _, binding := range wt.DataBindings {
		hostPath := binding.HostPath

		// If user provided a mapping, use it
		if userParams != nil && userParams.DataMappings != nil {
			if userPath, ok := userParams.DataMappings[binding.Name]; ok {
				hostPath = userPath
			}
		}

		// Check if required but not provided
		if binding.Required && hostPath == "" {
			return nil, fmt.Errorf("required data binding %s not provided", binding.Name)
		}

		if hostPath != "" {
			bindings = append(bindings, ResolvedDataBinding{
				Name:      binding.Name,
				MountPath: binding.MountPath,
				HostPath:  hostPath,
				DataType:  binding.DataType,
				ReadOnly:  binding.ReadOnly,
			})
		}
	}

	return bindings, nil
}

// generateSlurmScript generates the SLURM batch script
func (m *SlurmJobMapper) generateSlurmScript(
	wt *hpctypes.WorkloadTemplate,
	resources *ResolvedResources,
	environment map[string]string,
	_ []ResolvedDataBinding,
	userParams *UserParameters,
) (string, error) {
	var buf bytes.Buffer

	// Write SLURM directives
	buf.WriteString("#!/bin/bash\n")
	buf.WriteString(fmt.Sprintf("#SBATCH --job-name=%s\n", wt.TemplateID))
	buf.WriteString(fmt.Sprintf("#SBATCH --nodes=%d\n", resources.Nodes))
	buf.WriteString(fmt.Sprintf("#SBATCH --ntasks-per-node=%d\n", resources.CPUsPerNode))
	buf.WriteString(fmt.Sprintf("#SBATCH --mem=%dM\n", resources.MemoryMBPerNode))
	buf.WriteString(fmt.Sprintf("#SBATCH --time=%d\n", resources.RuntimeMinutes))

	if resources.GPUsPerNode > 0 {
		gpuSpec := fmt.Sprintf("%d", resources.GPUsPerNode)
		if len(resources.GPUTypes) > 0 {
			gpuSpec = fmt.Sprintf("%s:%d", resources.GPUTypes[0], resources.GPUsPerNode)
		}
		buf.WriteString(fmt.Sprintf("#SBATCH --gres=gpu:%s\n", gpuSpec))
	}

	if m.DefaultPartition != "" {
		buf.WriteString(fmt.Sprintf("#SBATCH --partition=%s\n", m.DefaultPartition))
	}

	if m.DefaultQoS != "" {
		buf.WriteString(fmt.Sprintf("#SBATCH --qos=%s\n", m.DefaultQoS))
	}

	if resources.ExclusiveNodes {
		buf.WriteString("#SBATCH --exclusive\n")
	}

	buf.WriteString("\n")

	// Load required modules
	if len(wt.Modules) > 0 {
		buf.WriteString("# Load required modules\n")
		for _, mod := range wt.Modules {
			buf.WriteString(fmt.Sprintf("module load %s\n", mod))
		}
		buf.WriteString("\n")
	}

	// Set environment variables
	if len(environment) > 0 {
		buf.WriteString("# Set environment variables\n")
		for key, value := range environment {
			// Escape single quotes in value
			escapedValue := strings.ReplaceAll(value, "'", "'\\''")
			buf.WriteString(fmt.Sprintf("export %s='%s'\n", key, escapedValue))
		}
		buf.WriteString("\n")
	}

	// Pre-run script
	if wt.Entrypoint.PreRunScript != "" {
		buf.WriteString("# Pre-run script\n")
		buf.WriteString(wt.Entrypoint.PreRunScript)
		buf.WriteString("\n\n")
	}

	// Change to working directory
	if wt.Entrypoint.WorkingDirectory != "" {
		buf.WriteString(fmt.Sprintf("cd %s\n", wt.Entrypoint.WorkingDirectory))
		buf.WriteString("\n")
	}

	// Generate command
	command, err := m.generateCommand(wt, userParams)
	if err != nil {
		return "", fmt.Errorf("failed to generate command: %w", err)
	}

	buf.WriteString("# Execute workload\n")
	buf.WriteString(command)
	buf.WriteString("\n\n")

	// Post-run script
	if wt.Entrypoint.PostRunScript != "" {
		buf.WriteString("# Post-run script\n")
		buf.WriteString(wt.Entrypoint.PostRunScript)
		buf.WriteString("\n")
	}

	return buf.String(), nil
}

// generateCommand generates the execution command
func (m *SlurmJobMapper) generateCommand(
	wt *hpctypes.WorkloadTemplate,
	userParams *UserParameters,
) (string, error) {
	command := wt.Entrypoint.Command

	// Build arguments
	args := wt.Entrypoint.DefaultArgs

	// If arg template is provided, substitute parameters
	if wt.Entrypoint.ArgTemplate != "" {
		substituted, err := m.substituteTemplate(wt.Entrypoint.ArgTemplate, userParams)
		if err != nil {
			return "", fmt.Errorf("failed to substitute arg template: %w", err)
		}
		args = strings.Fields(substituted)
	}

	// Wrap with MPI if needed
	if wt.Entrypoint.UseMPIRun {
		mpiCmd := "srun"
		if len(wt.Entrypoint.MPIRunArgs) > 0 {
			mpiCmd += " " + strings.Join(wt.Entrypoint.MPIRunArgs, " ")
		}
		command = fmt.Sprintf("%s %s %s", mpiCmd, command, strings.Join(args, " "))
	} else if len(args) > 0 {
		command = fmt.Sprintf("%s %s", command, strings.Join(args, " "))
	}

	return command, nil
}

// substituteTemplate performs template substitution
func (m *SlurmJobMapper) substituteTemplate(templateStr string, userParams *UserParameters) (string, error) {
	if userParams == nil || len(userParams.Parameters) == 0 {
		return templateStr, nil
	}

	tmpl, err := template.New("subst").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, userParams.Parameters); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
