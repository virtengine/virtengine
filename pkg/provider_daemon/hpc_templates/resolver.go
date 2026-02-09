// Package hpc_templates provides workload template resolution for HPC jobs.
//
// VE-5F: Template resolver - main entry point for template resolution
package hpc_templates

import (
	"fmt"

	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// DefaultTemplateResolver implements TemplateResolver
type DefaultTemplateResolver struct {
	// SlurmMapper resolves templates to SLURM jobs
	SlurmMapper *SlurmJobMapper

	// ContainerMapper resolves templates to container configs
	ContainerMapper *ContainerRuntimeMapper
}

// NewDefaultTemplateResolver creates a new template resolver
func NewDefaultTemplateResolver(slurmPartition, slurmQoS, modulePath string) *DefaultTemplateResolver {
	return &DefaultTemplateResolver{
		SlurmMapper:     NewSlurmJobMapper(slurmPartition, slurmQoS, modulePath),
		ContainerMapper: NewContainerRuntimeMapper(),
	}
}

// ResolveTemplate resolves a template to a runnable job configuration
func (r *DefaultTemplateResolver) ResolveTemplate(
	template *hpctypes.WorkloadTemplate,
	userParams *UserParameters,
) (*TemplateResolutionResult, error) {
	// Validate template first
	if err := r.ValidateTemplate(template); err != nil {
		return &TemplateResolutionResult{
			Success: false,
			Error:   fmt.Sprintf("template validation failed: %v", err),
		}, nil
	}

	// Check approval status
	if !template.ApprovalStatus.CanBeUsed() {
		return &TemplateResolutionResult{
			Success: false,
			Error:   fmt.Sprintf("template not approved (status: %s)", template.ApprovalStatus),
		}, nil
	}

	// Determine runtime type and resolve accordingly
	var resolvedJob *ResolvedJob
	var err error

	switch template.Runtime.RuntimeType {
	case "singularity", "apptainer", "container":
		resolvedJob, err = r.ContainerMapper.MapToContainerJob(template, userParams)
	case "native", "slurm":
		resolvedJob, err = r.SlurmMapper.MapToSlurmJob(template, userParams)
	default:
		return &TemplateResolutionResult{
			Success: false,
			Error:   fmt.Sprintf("unsupported runtime type: %s", template.Runtime.RuntimeType),
		}, nil
	}

	if err != nil {
		return &TemplateResolutionResult{
			Success: false,
			Error:   fmt.Sprintf("resolution failed: %v", err),
		}, nil
	}

	return &TemplateResolutionResult{
		Success:     true,
		ResolvedJob: resolvedJob,
	}, nil
}

// ValidateTemplate validates a template can be resolved
func (r *DefaultTemplateResolver) ValidateTemplate(template *hpctypes.WorkloadTemplate) error {
	// Validate template structure
	if err := template.Validate(); err != nil {
		return fmt.Errorf("template validation failed: %w", err)
	}

	// Validate runtime type is supported
	supportedRuntimes := map[string]bool{
		"singularity": true,
		"apptainer":   true,
		"container":   true,
		"native":      true,
		"slurm":       true,
	}

	if !supportedRuntimes[template.Runtime.RuntimeType] {
		return fmt.Errorf("unsupported runtime type: %s", template.Runtime.RuntimeType)
	}

	// Validate resources are reasonable
	if template.Resources.MinNodes > template.Resources.MaxNodes {
		return fmt.Errorf("min nodes > max nodes")
	}

	if template.Resources.DefaultNodes < template.Resources.MinNodes ||
		template.Resources.DefaultNodes > template.Resources.MaxNodes {
		return fmt.Errorf("default nodes out of range")
	}

	if template.Resources.MinCPUsPerNode > template.Resources.MaxCPUsPerNode {
		return fmt.Errorf("min CPUs > max CPUs")
	}

	if template.Resources.MinMemoryMBPerNode > template.Resources.MaxMemoryMBPerNode {
		return fmt.Errorf("min memory > max memory")
	}

	if template.Resources.MinRuntimeMinutes > template.Resources.MaxRuntimeMinutes {
		return fmt.Errorf("min runtime > max runtime")
	}

	// Validate security settings
	if err := r.validateSecurity(&template.Security); err != nil {
		return fmt.Errorf("security validation failed: %w", err)
	}

	// Validate entrypoint
	if template.Entrypoint.Command == "" {
		return fmt.Errorf("entrypoint command is required")
	}

	return nil
}

// validateSecurity validates security settings
func (r *DefaultTemplateResolver) validateSecurity(sec *hpctypes.WorkloadSecuritySpec) error {
	// Validate sandbox level
	validSandboxLevels := map[string]bool{
		"none":   true,
		"basic":  true,
		"strict": true,
	}

	if sec.SandboxLevel != "" && !validSandboxLevels[sec.SandboxLevel] {
		return fmt.Errorf("invalid sandbox level: %s", sec.SandboxLevel)
	}

	// Image digest requirement will be validated when template is resolved
	// based on the runtime configuration

	return nil
}

// ResolveWithDefaults resolves a template with default parameters (no user overrides)
func (r *DefaultTemplateResolver) ResolveWithDefaults(template *hpctypes.WorkloadTemplate) (*TemplateResolutionResult, error) {
	return r.ResolveTemplate(template, nil)
}

// GetSupportedRuntimeTypes returns the list of supported runtime types
func (r *DefaultTemplateResolver) GetSupportedRuntimeTypes() []string {
	return []string{"singularity", "apptainer", "container", "native", "slurm"}
}
