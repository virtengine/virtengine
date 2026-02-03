// Package hpc_workload_library provides workload validation functionality.
//
// VE-5F: Validation for custom workload uploads (resource limits, images, sandboxing)
package hpc_workload_library

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// ValidationConfig contains configuration for workload validation
type ValidationConfig struct {
	// MaxNodes is the maximum allowed nodes
	MaxNodes int32

	// MaxCPUsPerNode is the maximum CPUs per node
	MaxCPUsPerNode int32

	// MaxMemoryMBPerNode is the maximum memory per node in MB
	MaxMemoryMBPerNode int64

	// MaxGPUsPerNode is the maximum GPUs per node
	MaxGPUsPerNode int32

	// MaxRuntimeMinutes is the maximum runtime in minutes
	MaxRuntimeMinutes int64

	// MaxStorageGB is the maximum storage in GB
	MaxStorageGB int32

	// AllowedRegistries lists allowed container registries
	AllowedRegistries []string

	// BlockedRegistries lists blocked container registries
	BlockedRegistries []string

	// AllowedImages lists allowed image patterns (glob)
	AllowedImages []string

	// BlockedImages lists blocked image patterns (glob)
	BlockedImages []string

	// AllowedHostPaths lists allowed host mount paths
	AllowedHostPaths []string

	// RequireImageDigest requires image digest for verification
	RequireImageDigest bool

	// RequireSignedTemplate requires templates to be signed
	RequireSignedTemplate bool

	// AllowNetworkAccess allows network access by default
	AllowNetworkAccess bool

	// AllowCustomImages allows custom (non-template) images
	AllowCustomImages bool

	// EnforceSandboxing enforces sandboxing requirements
	EnforceSandboxing bool

	// MaxOpenFiles is the maximum allowed open files limit
	MaxOpenFiles int64

	// MaxProcesses is the maximum allowed processes limit
	MaxProcesses int64

	// MaxFileSize is the maximum allowed file size in bytes
	MaxFileSize int64
}

// DefaultValidationConfig returns the default validation configuration
func DefaultValidationConfig() ValidationConfig {
	return ValidationConfig{
		MaxNodes:           128,
		MaxCPUsPerNode:     256,
		MaxMemoryMBPerNode: 2048000, // 2TB
		MaxGPUsPerNode:     8,
		MaxRuntimeMinutes:  10080, // 7 days
		MaxStorageGB:       10240, // 10TB
		AllowedRegistries: []string{
			"library",
			"docker.io",
			"ghcr.io",
			"quay.io",
			"nvcr.io",
			"gcr.io",
		},
		BlockedRegistries: []string{},
		AllowedImages:     []string{"*"},
		BlockedImages: []string{
			"*/*:latest-untrusted",
		},
		AllowedHostPaths: []string{
			"/scratch",
			"/home",
			"/work",
			"/data",
			"/datasets",
			"/models",
		},
		RequireImageDigest:    false,
		RequireSignedTemplate: false,
		AllowNetworkAccess:    true,
		AllowCustomImages:     true,
		EnforceSandboxing:     true,
		MaxOpenFiles:          1048576,
		MaxProcesses:          131072,
		MaxFileSize:           107374182400, // 100GB
	}
}

// ValidationResult contains the result of workload validation
type ValidationResult struct {
	Valid    bool
	Errors   []ValidationError
	Warnings []ValidationWarning
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
	Code    string
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Field   string
	Message string
	Code    string
}

// IsValid returns true if validation passed
func (r *ValidationResult) IsValid() bool {
	return r.Valid && len(r.Errors) == 0
}

// Error returns validation errors as a single error
func (r *ValidationResult) Error() error {
	if r.IsValid() {
		return nil
	}
	msgs := make([]string, 0, len(r.Errors))
	for _, e := range r.Errors {
		msgs = append(msgs, fmt.Sprintf("%s: %s", e.Field, e.Message))
	}
	return fmt.Errorf("validation failed: %s", strings.Join(msgs, "; "))
}

// WorkloadValidator validates HPC workloads
type WorkloadValidator struct {
	config   ValidationConfig
	verifier *TemplateVerifier
}

// NewWorkloadValidator creates a new workload validator
func NewWorkloadValidator(config ValidationConfig) *WorkloadValidator {
	return &WorkloadValidator{
		config:   config,
		verifier: NewTemplateVerifier(),
	}
}

// ValidateTemplate validates a workload template
func (v *WorkloadValidator) ValidateTemplate(ctx context.Context, template *hpctypes.WorkloadTemplate) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Basic schema validation
	if err := template.Validate(); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "template",
			Message: err.Error(),
			Code:    "SCHEMA_INVALID",
		})
		return result
	}

	// Validate resources
	v.validateResources(template, result)

	// Validate runtime
	v.validateRuntime(template, result)

	// Validate security
	v.validateSecurity(template, result)

	// Validate signature if required
	if v.config.RequireSignedTemplate {
		v.validateSignature(template, result)
	}

	return result
}

// ValidateJob validates an HPC job against template and config
func (v *WorkloadValidator) ValidateJob(ctx context.Context, job *hpctypes.HPCJob, template *hpctypes.WorkloadTemplate) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Validate job against schema
	if err := job.Validate(); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "job",
			Message: err.Error(),
			Code:    "JOB_INVALID",
		})
		return result
	}

	// If using preconfigured workload, validate against template
	if job.WorkloadSpec.IsPreconfigured && template != nil {
		v.validateJobAgainstTemplate(job, template, result)
	} else {
		// Custom workload - apply stricter validation
		v.validateCustomWorkload(job, result)
	}

	return result
}

// validateResources validates resource specifications
func (v *WorkloadValidator) validateResources(template *hpctypes.WorkloadTemplate, result *ValidationResult) {
	r := template.Resources

	// Check max nodes
	if r.MaxNodes > v.config.MaxNodes {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "resources.max_nodes",
			Message: fmt.Sprintf("exceeds maximum allowed (%d > %d)", r.MaxNodes, v.config.MaxNodes),
			Code:    "RESOURCE_LIMIT_EXCEEDED",
		})
	}

	// Check max CPUs per node
	if r.MaxCPUsPerNode > v.config.MaxCPUsPerNode {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "resources.max_cpus_per_node",
			Message: fmt.Sprintf("exceeds maximum allowed (%d > %d)", r.MaxCPUsPerNode, v.config.MaxCPUsPerNode),
			Code:    "RESOURCE_LIMIT_EXCEEDED",
		})
	}

	// Check max memory per node
	if r.MaxMemoryMBPerNode > v.config.MaxMemoryMBPerNode {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "resources.max_memory_mb_per_node",
			Message: fmt.Sprintf("exceeds maximum allowed (%d > %d)", r.MaxMemoryMBPerNode, v.config.MaxMemoryMBPerNode),
			Code:    "RESOURCE_LIMIT_EXCEEDED",
		})
	}

	// Check max GPUs per node
	if r.MaxGPUsPerNode > v.config.MaxGPUsPerNode {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "resources.max_gpus_per_node",
			Message: fmt.Sprintf("exceeds maximum allowed (%d > %d)", r.MaxGPUsPerNode, v.config.MaxGPUsPerNode),
			Code:    "RESOURCE_LIMIT_EXCEEDED",
		})
	}

	// Check max runtime
	if r.MaxRuntimeMinutes > v.config.MaxRuntimeMinutes {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "resources.max_runtime_minutes",
			Message: fmt.Sprintf("exceeds maximum allowed (%d > %d)", r.MaxRuntimeMinutes, v.config.MaxRuntimeMinutes),
			Code:    "RESOURCE_LIMIT_EXCEEDED",
		})
	}

	// Check storage
	if int32(r.StorageGBRequired) > v.config.MaxStorageGB {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "resources.storage_gb_required",
			Message: fmt.Sprintf("exceeds maximum allowed (%d > %d)", r.StorageGBRequired, v.config.MaxStorageGB),
			Code:    "RESOURCE_LIMIT_EXCEEDED",
		})
	}
}

// validateRuntime validates runtime configuration
func (v *WorkloadValidator) validateRuntime(template *hpctypes.WorkloadTemplate, result *ValidationResult) {
	r := template.Runtime

	// Validate container image
	if r.ContainerImage != "" {
		// Check registry
		if !v.isRegistryAllowed(r.ContainerImage) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "runtime.container_image",
				Message: fmt.Sprintf("container registry not allowed: %s", getRegistryFromImage(r.ContainerImage)),
				Code:    "REGISTRY_NOT_ALLOWED",
			})
		}

		// Check image against allowed/blocked patterns
		if !v.isImageAllowed(r.ContainerImage) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "runtime.container_image",
				Message: fmt.Sprintf("container image not allowed: %s", r.ContainerImage),
				Code:    "IMAGE_NOT_ALLOWED",
			})
		}

		// Check for digest if required
		if v.config.RequireImageDigest && r.ImageDigest == "" {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "runtime.image_digest",
				Message: "image digest required for verification",
				Code:    "DIGEST_REQUIRED",
			})
		}
	}
}

// validateSecurity validates security configuration
func (v *WorkloadValidator) validateSecurity(template *hpctypes.WorkloadTemplate, result *ValidationResult) {
	s := template.Security

	// Check sandboxing
	if v.config.EnforceSandboxing && s.SandboxLevel == "none" {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "security.sandbox_level",
			Message: "sandboxing is disabled, this may pose security risks",
			Code:    "SANDBOXING_DISABLED",
		})
	}

	// Check host mount paths
	for _, binding := range template.DataBindings {
		if binding.HostPath != "" {
			if !v.isHostPathAllowed(binding.HostPath) {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Field:   fmt.Sprintf("data_bindings.%s.host_path", binding.Name),
					Message: fmt.Sprintf("host path not allowed: %s", binding.HostPath),
					Code:    "HOST_PATH_NOT_ALLOWED",
				})
			}
		}
	}

	// Check process limits
	if s.MaxOpenFiles > v.config.MaxOpenFiles {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "security.max_open_files",
			Message: fmt.Sprintf("exceeds maximum allowed (%d > %d)", s.MaxOpenFiles, v.config.MaxOpenFiles),
			Code:    "LIMIT_EXCEEDED",
		})
	}

	if s.MaxProcesses > v.config.MaxProcesses {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "security.max_processes",
			Message: fmt.Sprintf("exceeds maximum allowed (%d > %d)", s.MaxProcesses, v.config.MaxProcesses),
			Code:    "LIMIT_EXCEEDED",
		})
	}

	if s.MaxFileSize > v.config.MaxFileSize {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "security.max_file_size",
			Message: fmt.Sprintf("exceeds maximum allowed (%d > %d)", s.MaxFileSize, v.config.MaxFileSize),
			Code:    "LIMIT_EXCEEDED",
		})
	}
}

// validateSignature validates template signature
func (v *WorkloadValidator) validateSignature(template *hpctypes.WorkloadTemplate, result *ValidationResult) {
	if template.Signature.Signature == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "signature",
			Message: "template must be signed",
			Code:    "SIGNATURE_REQUIRED",
		})
		return
	}

	if err := v.verifier.Verify(template); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "signature",
			Message: fmt.Sprintf("signature verification failed: %s", err.Error()),
			Code:    "SIGNATURE_INVALID",
		})
	}
}

// validateJobAgainstTemplate validates a job against its template
func (v *WorkloadValidator) validateJobAgainstTemplate(job *hpctypes.HPCJob, template *hpctypes.WorkloadTemplate, result *ValidationResult) {
	r := template.Resources

	// Check node count
	if job.Resources.Nodes < r.MinNodes || job.Resources.Nodes > r.MaxNodes {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "resources.nodes",
			Message: fmt.Sprintf("nodes must be between %d and %d", r.MinNodes, r.MaxNodes),
			Code:    "RESOURCE_OUT_OF_RANGE",
		})
	}

	// Check CPUs per node
	if job.Resources.CPUCoresPerNode < r.MinCPUsPerNode || job.Resources.CPUCoresPerNode > r.MaxCPUsPerNode {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "resources.cpu_cores_per_node",
			Message: fmt.Sprintf("CPUs per node must be between %d and %d", r.MinCPUsPerNode, r.MaxCPUsPerNode),
			Code:    "RESOURCE_OUT_OF_RANGE",
		})
	}

	// Check memory per node
	memoryMB := int64(job.Resources.MemoryGBPerNode) * 1024
	if memoryMB < r.MinMemoryMBPerNode || memoryMB > r.MaxMemoryMBPerNode {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "resources.memory_gb_per_node",
			Message: fmt.Sprintf("memory per node must be between %dMB and %dMB", r.MinMemoryMBPerNode, r.MaxMemoryMBPerNode),
			Code:    "RESOURCE_OUT_OF_RANGE",
		})
	}

	// Check GPUs per node
	if job.Resources.GPUsPerNode < r.MinGPUsPerNode || job.Resources.GPUsPerNode > r.MaxGPUsPerNode {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "resources.gpus_per_node",
			Message: fmt.Sprintf("GPUs per node must be between %d and %d", r.MinGPUsPerNode, r.MaxGPUsPerNode),
			Code:    "RESOURCE_OUT_OF_RANGE",
		})
	}

	// Check runtime
	runtimeMinutes := job.MaxRuntimeSeconds / 60
	if runtimeMinutes < r.MinRuntimeMinutes || runtimeMinutes > r.MaxRuntimeMinutes {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "max_runtime_seconds",
			Message: fmt.Sprintf("runtime must be between %d and %d minutes", r.MinRuntimeMinutes, r.MaxRuntimeMinutes),
			Code:    "RESOURCE_OUT_OF_RANGE",
		})
	}

	// Validate GPU type if specified
	if job.Resources.GPUType != "" && len(r.GPUTypes) > 0 {
		found := false
		for _, allowed := range r.GPUTypes {
			if job.Resources.GPUType == allowed {
				found = true
				break
			}
		}
		if !found {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "resources.gpu_type",
				Message: fmt.Sprintf("GPU type %s not allowed for this template", job.Resources.GPUType),
				Code:    "GPU_TYPE_NOT_ALLOWED",
			})
		}
	}
}

// validateCustomWorkload validates a custom (non-template) workload
func (v *WorkloadValidator) validateCustomWorkload(job *hpctypes.HPCJob, result *ValidationResult) {
	if !v.config.AllowCustomImages {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "workload_spec",
			Message: "custom workloads are not allowed, use a preconfigured template",
			Code:    "CUSTOM_WORKLOAD_NOT_ALLOWED",
		})
		return
	}

	// Validate container image
	if job.WorkloadSpec.ContainerImage != "" {
		if !v.isRegistryAllowed(job.WorkloadSpec.ContainerImage) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "workload_spec.container_image",
				Message: "container registry not allowed",
				Code:    "REGISTRY_NOT_ALLOWED",
			})
		}

		if !v.isImageAllowed(job.WorkloadSpec.ContainerImage) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "workload_spec.container_image",
				Message: "container image not allowed",
				Code:    "IMAGE_NOT_ALLOWED",
			})
		}
	}

	// Apply global resource limits
	if job.Resources.Nodes > v.config.MaxNodes {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "resources.nodes",
			Message: fmt.Sprintf("exceeds maximum allowed (%d > %d)", job.Resources.Nodes, v.config.MaxNodes),
			Code:    "RESOURCE_LIMIT_EXCEEDED",
		})
	}

	if job.Resources.CPUCoresPerNode > v.config.MaxCPUsPerNode {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "resources.cpu_cores_per_node",
			Message: fmt.Sprintf("exceeds maximum allowed (%d > %d)", job.Resources.CPUCoresPerNode, v.config.MaxCPUsPerNode),
			Code:    "RESOURCE_LIMIT_EXCEEDED",
		})
	}

	if job.Resources.GPUsPerNode > v.config.MaxGPUsPerNode {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "resources.gpus_per_node",
			Message: fmt.Sprintf("exceeds maximum allowed (%d > %d)", job.Resources.GPUsPerNode, v.config.MaxGPUsPerNode),
			Code:    "RESOURCE_LIMIT_EXCEEDED",
		})
	}

	runtimeMinutes := job.MaxRuntimeSeconds / 60
	if runtimeMinutes > v.config.MaxRuntimeMinutes {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "max_runtime_seconds",
			Message: fmt.Sprintf("exceeds maximum allowed (%d > %d minutes)", runtimeMinutes, v.config.MaxRuntimeMinutes),
			Code:    "RESOURCE_LIMIT_EXCEEDED",
		})
	}
}

// isRegistryAllowed checks if a container registry is allowed
func (v *WorkloadValidator) isRegistryAllowed(image string) bool {
	registry := getRegistryFromImage(image)

	// Check blocked registries first
	for _, blocked := range v.config.BlockedRegistries {
		if registry == blocked {
			return false
		}
	}

	// Check allowed registries
	for _, allowed := range v.config.AllowedRegistries {
		if registry == allowed || allowed == "*" {
			return true
		}
	}

	return false
}

// isImageAllowed checks if a container image matches allowed patterns
func (v *WorkloadValidator) isImageAllowed(image string) bool {
	// Check blocked images first
	for _, pattern := range v.config.BlockedImages {
		if matchGlob(pattern, image) {
			return false
		}
	}

	// Check allowed images
	for _, pattern := range v.config.AllowedImages {
		if matchGlob(pattern, image) {
			return true
		}
	}

	return false
}

// isHostPathAllowed checks if a host path is allowed for mounting
func (v *WorkloadValidator) isHostPathAllowed(path string) bool {
	cleanPath := filepath.Clean(path)

	for _, allowed := range v.config.AllowedHostPaths {
		allowedClean := filepath.Clean(allowed)
		if strings.HasPrefix(cleanPath, allowedClean) {
			return true
		}
	}

	return false
}

// getRegistryFromImage extracts the registry from a container image
func getRegistryFromImage(image string) string {
	parts := strings.SplitN(image, "/", 2)
	if len(parts) == 1 {
		return "library"
	}

	// Check if first part looks like a registry (has . or :)
	if strings.Contains(parts[0], ".") || strings.Contains(parts[0], ":") {
		return parts[0]
	}

	// Docker Hub user/image format
	return "docker.io"
}

// matchGlob performs simple glob matching
func matchGlob(pattern, str string) bool {
	// Convert glob to regex
	regexPattern := "^"
	for _, c := range pattern {
		switch c {
		case '*':
			regexPattern += ".*"
		case '?':
			regexPattern += "."
		case '.', '/', '+', '^', '$', '|', '(', ')', '[', ']', '{', '}':
			regexPattern += "\\" + string(c)
		default:
			regexPattern += string(c)
		}
	}
	regexPattern += "$"

	matched, _ := regexp.MatchString(regexPattern, str)
	return matched
}
