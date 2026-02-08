// Package hpc_templates provides workload template resolution for HPC jobs.
//
// VE-5F: Container runtime mapper - resolves templates to container runtime configs
package hpc_templates

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// ContainerRuntimeMapper maps WorkloadTemplate to container runtime configurations
type ContainerRuntimeMapper struct {
	// DefaultRegistry is the default container registry
	DefaultRegistry string

	// AllowedRegistries are registries allowed for image pull
	AllowedRegistries []string
}

// NewContainerRuntimeMapper creates a new container runtime mapper
func NewContainerRuntimeMapper() *ContainerRuntimeMapper {
	return &ContainerRuntimeMapper{
		DefaultRegistry: "docker.io",
		AllowedRegistries: []string{
			"docker.io",
			"ghcr.io",
			"quay.io",
		},
	}
}

// MapToContainerJob converts a WorkloadTemplate to a container runtime configuration
func (m *ContainerRuntimeMapper) MapToContainerJob(
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

	// Build container config
	containerConfig, err := m.buildContainerConfig(wt, environment, dataBindings, userParams)
	if err != nil {
		return nil, fmt.Errorf("failed to build container config: %w", err)
	}

	return &ResolvedJob{
		TemplateID:      wt.TemplateID,
		TemplateVersion: wt.Version,
		JobType:         "container",
		ContainerConfig: containerConfig,
		Resources:       *resources,
		Environment:     environment,
		DataBindings:    dataBindings,
		ResolvedAt:      time.Now().UTC(),
	}, nil
}

// resolveResources resolves resource requirements
func (m *ContainerRuntimeMapper) resolveResources(
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

	return res, nil
}

// applyResourceOverrides applies user resource overrides
func (m *ContainerRuntimeMapper) applyResourceOverrides(
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

	if overrides.RuntimeMinutes != nil {
		runtime := *overrides.RuntimeMinutes
		if runtime < wt.Resources.MinRuntimeMinutes || runtime > wt.Resources.MaxRuntimeMinutes {
			return fmt.Errorf("runtime %d out of range [%d, %d]", runtime, wt.Resources.MinRuntimeMinutes, wt.Resources.MaxRuntimeMinutes)
		}
		res.RuntimeMinutes = runtime
	}

	return nil
}

// resolveEnvironment resolves environment variables
func (m *ContainerRuntimeMapper) resolveEnvironment(
	wt *hpctypes.WorkloadTemplate,
	userParams *UserParameters,
) (map[string]string, error) {
	env := make(map[string]string)

	// Add template environment variables
	for _, envVar := range wt.Environment {
		value := envVar.Value

		// Check if required but not provided
		if envVar.Required && value == "" {
			if userParams == nil || userParams.Parameters[envVar.Name] == "" {
				return nil, fmt.Errorf("required environment variable %s not provided", envVar.Name)
			}
			value = userParams.Parameters[envVar.Name]
		}

		if value != "" {
			env[envVar.Name] = value
		}
	}

	// Add user-provided parameters
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
func (m *ContainerRuntimeMapper) resolveDataBindings(
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

// buildContainerConfig builds the container runtime configuration
func (m *ContainerRuntimeMapper) buildContainerConfig(
	wt *hpctypes.WorkloadTemplate,
	environment map[string]string,
	dataBindings []ResolvedDataBinding,
	_ *UserParameters,
) (*ContainerRuntimeConfig, error) {
	// Resolve container image
	image, digest, err := m.resolveImage(wt)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve image: %w", err)
	}

	// Build mounts
	mounts := make([]ContainerMount, 0, len(dataBindings))
	for _, binding := range dataBindings {
		mounts = append(mounts, ContainerMount{
			Source:   binding.HostPath,
			Target:   binding.MountPath,
			ReadOnly: binding.ReadOnly,
		})
	}

	// Build command and args
	command := wt.Entrypoint.Command
	args := wt.Entrypoint.DefaultArgs

	// Build security options
	securityOpts := &ContainerSecurityOptions{
		AllowNetworkAccess: wt.Security.AllowNetworkAccess,
		AllowHostMounts:    wt.Security.AllowHostMounts,
		SandboxLevel:       wt.Security.SandboxLevel,
		MaxOpenFiles:       wt.Security.MaxOpenFiles,
		MaxProcesses:       wt.Security.MaxProcesses,
		MaxFileSize:        wt.Security.MaxFileSize,
	}

	config := &ContainerRuntimeConfig{
		RuntimeType:     wt.Runtime.RuntimeType,
		Image:           image,
		ImageDigest:     digest,
		Command:         command,
		Args:            args,
		WorkingDir:      wt.Entrypoint.WorkingDirectory,
		Mounts:          mounts,
		Environment:     environment,
		PreRunScript:    wt.Entrypoint.PreRunScript,
		PostRunScript:   wt.Entrypoint.PostRunScript,
		SecurityOptions: securityOpts,
	}

	return config, nil
}

// resolveImage resolves the container image reference
func (m *ContainerRuntimeMapper) resolveImage(wt *hpctypes.WorkloadTemplate) (string, string, error) {
	image := wt.Runtime.ContainerImage
	if image == "" {
		return "", "", fmt.Errorf("container image not specified")
	}

	// Check if image has registry prefix
	if !strings.Contains(image, "/") {
		// No registry prefix, add default
		image = filepath.Join(m.DefaultRegistry, "library", image)
	} else if !strings.Contains(strings.Split(image, "/")[0], ".") {
		// Has path but no registry (e.g., virtengine/workload)
		image = filepath.Join(m.DefaultRegistry, image)
	}

	// Verify registry is allowed
	registry := strings.Split(image, "/")[0]
	if !m.isAllowedRegistry(registry, wt) {
		return "", "", fmt.Errorf("registry %s not allowed", registry)
	}

	// Get image digest if provided
	digest := wt.Runtime.ImageDigest
	if wt.Security.RequireImageDigest && digest == "" {
		return "", "", fmt.Errorf("image digest required but not provided")
	}

	return image, digest, nil
}

// isAllowedRegistry checks if a registry is allowed
func (m *ContainerRuntimeMapper) isAllowedRegistry(registry string, wt *hpctypes.WorkloadTemplate) bool {
	// If template specifies allowed registries, use those
	if len(wt.Security.AllowedRegistries) > 0 {
		for _, allowed := range wt.Security.AllowedRegistries {
			if registry == allowed {
				return true
			}
		}
		return false
	}

	// Otherwise use mapper's allowed registries
	for _, allowed := range m.AllowedRegistries {
		if registry == allowed {
			return true
		}
	}

	return false
}
