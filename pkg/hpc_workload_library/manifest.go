// Package hpc_workload_library provides YAML manifest loading.
//
// VE-5F: YAML manifest format for workload templates
package hpc_workload_library

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

//go:embed templates/*.yaml
var embeddedTemplates embed.FS

// ManifestSchemaVersion is the current manifest schema version
const ManifestSchemaVersion = "1.0.0"

// YAMLManifest represents a YAML manifest file
type YAMLManifest struct {
	SchemaVersion string         `yaml:"schema_version"`
	Template      YAMLTemplate   `yaml:"template"`
	Signature     *YAMLSignature `yaml:"signature,omitempty"`
}

// YAMLTemplate represents a template in YAML format
type YAMLTemplate struct {
	TemplateID     string                `yaml:"template_id"`
	Name           string                `yaml:"name"`
	Version        string                `yaml:"version"`
	Description    string                `yaml:"description"`
	Type           string                `yaml:"type"`
	Runtime        YAMLRuntime           `yaml:"runtime"`
	Resources      YAMLResources         `yaml:"resources"`
	Security       YAMLSecurity          `yaml:"security"`
	Entrypoint     YAMLEntrypoint        `yaml:"entrypoint"`
	Environment    []YAMLEnvironment     `yaml:"environment,omitempty"`
	Modules        []string              `yaml:"modules,omitempty"`
	DataBindings   []YAMLDataBinding     `yaml:"data_bindings,omitempty"`
	ParameterSchema []YAMLParameter      `yaml:"parameter_schema,omitempty"`
	ApprovalStatus string                `yaml:"approval_status"`
	Publisher      string                `yaml:"publisher,omitempty"`
	Tags           []string              `yaml:"tags,omitempty"`
}

// YAMLRuntime represents runtime configuration
type YAMLRuntime struct {
	RuntimeType       string   `yaml:"runtime_type"`
	ContainerImage    string   `yaml:"container_image,omitempty"`
	ContainerRegistry string   `yaml:"container_registry,omitempty"`
	ImageDigest       string   `yaml:"image_digest,omitempty"`
	RequiredModules   []string `yaml:"required_modules,omitempty"`
	MPIImplementation string   `yaml:"mpi_implementation,omitempty"`
	CUDAVersion       string   `yaml:"cuda_version,omitempty"`
	PythonVersion     string   `yaml:"python_version,omitempty"`
}

// YAMLResources represents resource configuration
type YAMLResources struct {
	MinNodes               int32    `yaml:"min_nodes"`
	MaxNodes               int32    `yaml:"max_nodes"`
	DefaultNodes           int32    `yaml:"default_nodes"`
	MinCPUsPerNode         int32    `yaml:"min_cpus_per_node"`
	MaxCPUsPerNode         int32    `yaml:"max_cpus_per_node"`
	DefaultCPUsPerNode     int32    `yaml:"default_cpus_per_node"`
	MinMemoryMBPerNode     int64    `yaml:"min_memory_mb_per_node"`
	MaxMemoryMBPerNode     int64    `yaml:"max_memory_mb_per_node"`
	DefaultMemoryMBPerNode int64    `yaml:"default_memory_mb_per_node"`
	MinGPUsPerNode         int32    `yaml:"min_gpus_per_node,omitempty"`
	MaxGPUsPerNode         int32    `yaml:"max_gpus_per_node,omitempty"`
	DefaultGPUsPerNode     int32    `yaml:"default_gpus_per_node,omitempty"`
	GPUTypes               []string `yaml:"gpu_types,omitempty"`
	MinRuntimeMinutes      int64    `yaml:"min_runtime_minutes"`
	MaxRuntimeMinutes      int64    `yaml:"max_runtime_minutes"`
	DefaultRuntimeMinutes  int64    `yaml:"default_runtime_minutes"`
	StorageGBRequired      int32    `yaml:"storage_gb_required,omitempty"`
	NetworkRequired        bool     `yaml:"network_required,omitempty"`
	ExclusiveNodes         bool     `yaml:"exclusive_nodes,omitempty"`
}

// YAMLSecurity represents security configuration
type YAMLSecurity struct {
	AllowedRegistries  []string `yaml:"allowed_registries,omitempty"`
	BlockedRegistries  []string `yaml:"blocked_registries,omitempty"`
	AllowedImages      []string `yaml:"allowed_images,omitempty"`
	BlockedImages      []string `yaml:"blocked_images,omitempty"`
	RequireImageDigest bool     `yaml:"require_image_digest"`
	AllowNetworkAccess bool     `yaml:"allow_network_access"`
	AllowHostMounts    bool     `yaml:"allow_host_mounts"`
	AllowedHostPaths   []string `yaml:"allowed_host_paths,omitempty"`
	SandboxLevel       string   `yaml:"sandbox_level"`
	MaxOpenFiles       int64    `yaml:"max_open_files,omitempty"`
	MaxProcesses       int64    `yaml:"max_processes,omitempty"`
	MaxFileSize        int64    `yaml:"max_file_size,omitempty"`
}

// YAMLEntrypoint represents entrypoint configuration
type YAMLEntrypoint struct {
	Command          string   `yaml:"command"`
	DefaultArgs      []string `yaml:"default_args,omitempty"`
	ArgTemplate      string   `yaml:"arg_template,omitempty"`
	WorkingDirectory string   `yaml:"working_directory,omitempty"`
	PreRunScript     string   `yaml:"pre_run_script,omitempty"`
	PostRunScript    string   `yaml:"post_run_script,omitempty"`
	UseMPIRun        bool     `yaml:"use_mpirun,omitempty"`
	MPIRunArgs       []string `yaml:"mpirun_args,omitempty"`
}

// YAMLEnvironment represents environment variable configuration
type YAMLEnvironment struct {
	Name          string `yaml:"name"`
	Value         string `yaml:"value,omitempty"`
	ValueTemplate string `yaml:"value_template,omitempty"`
	Required      bool   `yaml:"required,omitempty"`
	Secret        bool   `yaml:"secret,omitempty"`
	Description   string `yaml:"description,omitempty"`
}

// YAMLDataBinding represents data binding configuration
type YAMLDataBinding struct {
	Name      string `yaml:"name"`
	MountPath string `yaml:"mount_path"`
	HostPath  string `yaml:"host_path,omitempty"`
	DataType  string `yaml:"data_type"`
	Required  bool   `yaml:"required"`
	ReadOnly  bool   `yaml:"read_only"`
}

// YAMLParameter represents parameter definition
type YAMLParameter struct {
	Name        string   `yaml:"name"`
	Type        string   `yaml:"type"`
	Description string   `yaml:"description,omitempty"`
	Default     string   `yaml:"default,omitempty"`
	Required    bool     `yaml:"required,omitempty"`
	EnumValues  []string `yaml:"enum_values,omitempty"`
	MinValue    string   `yaml:"min_value,omitempty"`
	MaxValue    string   `yaml:"max_value,omitempty"`
	Pattern     string   `yaml:"pattern,omitempty"`
}

// YAMLSignature represents template signature
type YAMLSignature struct {
	Algorithm       string `yaml:"algorithm"`
	PublisherPubKey string `yaml:"publisher_pub_key"`
	Signature       string `yaml:"signature"`
	SignedAt        string `yaml:"signed_at"`
	ContentHash     string `yaml:"content_hash"`
}

// ManifestLoader loads templates from YAML manifests
type ManifestLoader struct {
	defaultPublisher string
}

// NewManifestLoader creates a new manifest loader
func NewManifestLoader(defaultPublisher string) *ManifestLoader {
	if defaultPublisher == "" {
		defaultPublisher = BuiltinTemplatePublisher
	}
	return &ManifestLoader{
		defaultPublisher: defaultPublisher,
	}
}

// LoadFromBytes loads a template from YAML bytes
func (l *ManifestLoader) LoadFromBytes(data []byte) (*hpctypes.WorkloadTemplate, error) {
	var manifest YAMLManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return l.convertToTemplate(&manifest)
}

// LoadFromFile loads a template from a YAML file
func (l *ManifestLoader) LoadFromFile(path string) (*hpctypes.WorkloadTemplate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return l.LoadFromBytes(data)
}

// LoadFromDirectory loads all templates from a directory
func (l *ManifestLoader) LoadFromDirectory(dir string) ([]*hpctypes.WorkloadTemplate, error) {
	var templates []*hpctypes.WorkloadTemplate

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}

		template, err := l.LoadFromFile(path)
		if err != nil {
			return fmt.Errorf("failed to load %s: %w", path, err)
		}

		templates = append(templates, template)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return templates, nil
}

// LoadEmbedded loads templates from embedded files
func (l *ManifestLoader) LoadEmbedded() ([]*hpctypes.WorkloadTemplate, error) {
	var templates []*hpctypes.WorkloadTemplate

	err := fs.WalkDir(embeddedTemplates, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}

		data, err := embeddedTemplates.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}

		template, err := l.LoadFromBytes(data)
		if err != nil {
			return fmt.Errorf("failed to parse embedded file %s: %w", path, err)
		}

		templates = append(templates, template)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return templates, nil
}

// convertToTemplate converts YAML manifest to WorkloadTemplate
func (l *ManifestLoader) convertToTemplate(manifest *YAMLManifest) (*hpctypes.WorkloadTemplate, error) {
	t := manifest.Template

	publisher := t.Publisher
	if publisher == "" {
		publisher = l.defaultPublisher
	}

	workloadType := hpctypes.WorkloadType(t.Type)
	if !workloadType.IsValid() {
		return nil, fmt.Errorf("invalid workload type: %s", t.Type)
	}

	approvalStatus := hpctypes.WorkloadApprovalStatus(t.ApprovalStatus)
	if !approvalStatus.IsValid() {
		approvalStatus = hpctypes.WorkloadApprovalPending
	}

	template := &hpctypes.WorkloadTemplate{
		TemplateID:  t.TemplateID,
		Name:        t.Name,
		Version:     t.Version,
		Description: t.Description,
		Type:        workloadType,
		Runtime: hpctypes.WorkloadRuntime{
			RuntimeType:       t.Runtime.RuntimeType,
			ContainerImage:    t.Runtime.ContainerImage,
			ContainerRegistry: t.Runtime.ContainerRegistry,
			ImageDigest:       t.Runtime.ImageDigest,
			RequiredModules:   t.Runtime.RequiredModules,
			MPIImplementation: t.Runtime.MPIImplementation,
			CUDAVersion:       t.Runtime.CUDAVersion,
			PythonVersion:     t.Runtime.PythonVersion,
		},
		Resources: hpctypes.WorkloadResourceSpec{
			MinNodes:               t.Resources.MinNodes,
			MaxNodes:               t.Resources.MaxNodes,
			DefaultNodes:           t.Resources.DefaultNodes,
			MinCPUsPerNode:         t.Resources.MinCPUsPerNode,
			MaxCPUsPerNode:         t.Resources.MaxCPUsPerNode,
			DefaultCPUsPerNode:     t.Resources.DefaultCPUsPerNode,
			MinMemoryMBPerNode:     t.Resources.MinMemoryMBPerNode,
			MaxMemoryMBPerNode:     t.Resources.MaxMemoryMBPerNode,
			DefaultMemoryMBPerNode: t.Resources.DefaultMemoryMBPerNode,
			MinGPUsPerNode:         t.Resources.MinGPUsPerNode,
			MaxGPUsPerNode:         t.Resources.MaxGPUsPerNode,
			DefaultGPUsPerNode:     t.Resources.DefaultGPUsPerNode,
			GPUTypes:               t.Resources.GPUTypes,
			MinRuntimeMinutes:      t.Resources.MinRuntimeMinutes,
			MaxRuntimeMinutes:      t.Resources.MaxRuntimeMinutes,
			DefaultRuntimeMinutes:  t.Resources.DefaultRuntimeMinutes,
			StorageGBRequired:      t.Resources.StorageGBRequired,
			NetworkRequired:        t.Resources.NetworkRequired,
			ExclusiveNodes:         t.Resources.ExclusiveNodes,
		},
		Security: hpctypes.WorkloadSecuritySpec{
			AllowedRegistries:  t.Security.AllowedRegistries,
			BlockedImages:      t.Security.BlockedImages,
			AllowedImages:      t.Security.AllowedImages,
			RequireImageDigest: t.Security.RequireImageDigest,
			AllowNetworkAccess: t.Security.AllowNetworkAccess,
			AllowHostMounts:    t.Security.AllowHostMounts,
			AllowedHostPaths:   t.Security.AllowedHostPaths,
			SandboxLevel:       t.Security.SandboxLevel,
			MaxOpenFiles:       t.Security.MaxOpenFiles,
			MaxProcesses:       t.Security.MaxProcesses,
			MaxFileSize:        t.Security.MaxFileSize,
		},
		Entrypoint: hpctypes.WorkloadEntrypoint{
			Command:          t.Entrypoint.Command,
			DefaultArgs:      t.Entrypoint.DefaultArgs,
			ArgTemplate:      t.Entrypoint.ArgTemplate,
			WorkingDirectory: t.Entrypoint.WorkingDirectory,
			PreRunScript:     t.Entrypoint.PreRunScript,
			PostRunScript:    t.Entrypoint.PostRunScript,
			UseMPIRun:        t.Entrypoint.UseMPIRun,
			MPIRunArgs:       t.Entrypoint.MPIRunArgs,
		},
		Modules:        t.Modules,
		ApprovalStatus: approvalStatus,
		Publisher:      publisher,
		Tags:           t.Tags,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Convert environment variables
	for _, env := range t.Environment {
		template.Environment = append(template.Environment, hpctypes.EnvironmentVariable{
			Name:          env.Name,
			Value:         env.Value,
			ValueTemplate: env.ValueTemplate,
			Required:      env.Required,
			Secret:        env.Secret,
			Description:   env.Description,
		})
	}

	// Convert data bindings
	for _, binding := range t.DataBindings {
		template.DataBindings = append(template.DataBindings, hpctypes.DataBinding{
			Name:      binding.Name,
			MountPath: binding.MountPath,
			HostPath:  binding.HostPath,
			DataType:  binding.DataType,
			Required:  binding.Required,
			ReadOnly:  binding.ReadOnly,
		})
	}

	// Convert parameter schema
	for _, param := range t.ParameterSchema {
		template.ParameterSchema = append(template.ParameterSchema, hpctypes.ParameterDefinition{
			Name:        param.Name,
			Type:        param.Type,
			Description: param.Description,
			Default:     param.Default,
			Required:    param.Required,
			EnumValues:  param.EnumValues,
			MinValue:    param.MinValue,
			MaxValue:    param.MaxValue,
			Pattern:     param.Pattern,
		})
	}

	// Convert signature if present
	if manifest.Signature != nil {
		signedAt, _ := time.Parse(time.RFC3339, manifest.Signature.SignedAt)
		template.Signature = hpctypes.WorkloadSignature{
			Algorithm:       manifest.Signature.Algorithm,
			PublisherPubKey: manifest.Signature.PublisherPubKey,
			Signature:       manifest.Signature.Signature,
			SignedAt:        signedAt,
			ContentHash:     manifest.Signature.ContentHash,
		}
	}

	return template, nil
}

// ExportToYAML exports a template to YAML format
func ExportToYAML(template *hpctypes.WorkloadTemplate) ([]byte, error) {
	manifest := YAMLManifest{
		SchemaVersion: ManifestSchemaVersion,
		Template: YAMLTemplate{
			TemplateID:     template.TemplateID,
			Name:           template.Name,
			Version:        template.Version,
			Description:    template.Description,
			Type:           string(template.Type),
			ApprovalStatus: string(template.ApprovalStatus),
			Publisher:      template.Publisher,
			Tags:           template.Tags,
			Modules:        template.Modules,
			Runtime: YAMLRuntime{
				RuntimeType:       template.Runtime.RuntimeType,
				ContainerImage:    template.Runtime.ContainerImage,
				ContainerRegistry: template.Runtime.ContainerRegistry,
				ImageDigest:       template.Runtime.ImageDigest,
				RequiredModules:   template.Runtime.RequiredModules,
				MPIImplementation: template.Runtime.MPIImplementation,
				CUDAVersion:       template.Runtime.CUDAVersion,
				PythonVersion:     template.Runtime.PythonVersion,
			},
			Resources: YAMLResources{
				MinNodes:               template.Resources.MinNodes,
				MaxNodes:               template.Resources.MaxNodes,
				DefaultNodes:           template.Resources.DefaultNodes,
				MinCPUsPerNode:         template.Resources.MinCPUsPerNode,
				MaxCPUsPerNode:         template.Resources.MaxCPUsPerNode,
				DefaultCPUsPerNode:     template.Resources.DefaultCPUsPerNode,
				MinMemoryMBPerNode:     template.Resources.MinMemoryMBPerNode,
				MaxMemoryMBPerNode:     template.Resources.MaxMemoryMBPerNode,
				DefaultMemoryMBPerNode: template.Resources.DefaultMemoryMBPerNode,
				MinGPUsPerNode:         template.Resources.MinGPUsPerNode,
				MaxGPUsPerNode:         template.Resources.MaxGPUsPerNode,
				DefaultGPUsPerNode:     template.Resources.DefaultGPUsPerNode,
				GPUTypes:               template.Resources.GPUTypes,
				MinRuntimeMinutes:      template.Resources.MinRuntimeMinutes,
				MaxRuntimeMinutes:      template.Resources.MaxRuntimeMinutes,
				DefaultRuntimeMinutes:  template.Resources.DefaultRuntimeMinutes,
				StorageGBRequired:      template.Resources.StorageGBRequired,
				NetworkRequired:        template.Resources.NetworkRequired,
				ExclusiveNodes:         template.Resources.ExclusiveNodes,
			},
			Security: YAMLSecurity{
				AllowedRegistries:  template.Security.AllowedRegistries,
				AllowedImages:      template.Security.AllowedImages,
				BlockedImages:      template.Security.BlockedImages,
				RequireImageDigest: template.Security.RequireImageDigest,
				AllowNetworkAccess: template.Security.AllowNetworkAccess,
				AllowHostMounts:    template.Security.AllowHostMounts,
				AllowedHostPaths:   template.Security.AllowedHostPaths,
				SandboxLevel:       template.Security.SandboxLevel,
				MaxOpenFiles:       template.Security.MaxOpenFiles,
				MaxProcesses:       template.Security.MaxProcesses,
				MaxFileSize:        template.Security.MaxFileSize,
			},
			Entrypoint: YAMLEntrypoint{
				Command:          template.Entrypoint.Command,
				DefaultArgs:      template.Entrypoint.DefaultArgs,
				ArgTemplate:      template.Entrypoint.ArgTemplate,
				WorkingDirectory: template.Entrypoint.WorkingDirectory,
				PreRunScript:     template.Entrypoint.PreRunScript,
				PostRunScript:    template.Entrypoint.PostRunScript,
				UseMPIRun:        template.Entrypoint.UseMPIRun,
				MPIRunArgs:       template.Entrypoint.MPIRunArgs,
			},
		},
	}

	// Convert environment
	for _, env := range template.Environment {
		manifest.Template.Environment = append(manifest.Template.Environment, YAMLEnvironment{
			Name:          env.Name,
			Value:         env.Value,
			ValueTemplate: env.ValueTemplate,
			Required:      env.Required,
			Secret:        env.Secret,
			Description:   env.Description,
		})
	}

	// Convert data bindings
	for _, binding := range template.DataBindings {
		manifest.Template.DataBindings = append(manifest.Template.DataBindings, YAMLDataBinding{
			Name:      binding.Name,
			MountPath: binding.MountPath,
			HostPath:  binding.HostPath,
			DataType:  binding.DataType,
			Required:  binding.Required,
			ReadOnly:  binding.ReadOnly,
		})
	}

	// Convert parameters
	for _, param := range template.ParameterSchema {
		manifest.Template.ParameterSchema = append(manifest.Template.ParameterSchema, YAMLParameter{
			Name:        param.Name,
			Type:        param.Type,
			Description: param.Description,
			Default:     param.Default,
			Required:    param.Required,
			EnumValues:  param.EnumValues,
			MinValue:    param.MinValue,
			MaxValue:    param.MaxValue,
			Pattern:     param.Pattern,
		})
	}

	// Convert signature
	if template.Signature.Signature != "" {
		manifest.Signature = &YAMLSignature{
			Algorithm:       template.Signature.Algorithm,
			PublisherPubKey: template.Signature.PublisherPubKey,
			Signature:       template.Signature.Signature,
			SignedAt:        template.Signature.SignedAt.Format(time.RFC3339),
			ContentHash:     template.Signature.ContentHash,
		}
	}

	return yaml.Marshal(manifest)
}

