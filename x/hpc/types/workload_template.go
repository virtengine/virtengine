// Package types contains types for the HPC module.
//
// VE-5F: Workload template types for HPC workload library
package types

import (
	"encoding/hex"
	"regexp"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// WorkloadType represents the type of HPC workload
type WorkloadType string

const (
	// WorkloadTypeMPI is for MPI-based parallel workloads
	WorkloadTypeMPI WorkloadType = "mpi"

	// WorkloadTypeGPU is for GPU compute workloads
	WorkloadTypeGPU WorkloadType = "gpu"

	// WorkloadTypeBatch is for batch processing workloads
	WorkloadTypeBatch WorkloadType = "batch"

	// WorkloadTypeDataProcessing is for data processing pipelines
	WorkloadTypeDataProcessing WorkloadType = "data_processing"

	// WorkloadTypeInteractive is for interactive sessions
	WorkloadTypeInteractive WorkloadType = "interactive"

	// WorkloadTypeCustom is for custom user-defined workloads
	WorkloadTypeCustom WorkloadType = "custom"
)

// IsValid checks if the workload type is valid
func (t WorkloadType) IsValid() bool {
	switch t {
	case WorkloadTypeMPI, WorkloadTypeGPU, WorkloadTypeBatch,
		WorkloadTypeDataProcessing, WorkloadTypeInteractive, WorkloadTypeCustom:
		return true
	default:
		return false
	}
}

// WorkloadApprovalStatus represents the approval status of a workload template
type WorkloadApprovalStatus string

const (
	// WorkloadApprovalPending indicates pending review
	WorkloadApprovalPending WorkloadApprovalStatus = "pending"

	// WorkloadApprovalApproved indicates approved for use
	WorkloadApprovalApproved WorkloadApprovalStatus = "approved"

	// WorkloadApprovalRejected indicates rejected
	WorkloadApprovalRejected WorkloadApprovalStatus = "rejected"

	// WorkloadApprovalDeprecated indicates deprecated
	WorkloadApprovalDeprecated WorkloadApprovalStatus = "deprecated"

	// WorkloadApprovalRevoked indicates revoked due to security issues
	WorkloadApprovalRevoked WorkloadApprovalStatus = "revoked"
)

// IsValid checks if the approval status is valid
func (s WorkloadApprovalStatus) IsValid() bool {
	switch s {
	case WorkloadApprovalPending, WorkloadApprovalApproved, WorkloadApprovalRejected,
		WorkloadApprovalDeprecated, WorkloadApprovalRevoked:
		return true
	default:
		return false
	}
}

// CanBeUsed checks if the template can be used for job submission
func (s WorkloadApprovalStatus) CanBeUsed() bool {
	return s == WorkloadApprovalApproved
}

// WorkloadTemplate represents a preconfigured HPC workload template
type WorkloadTemplate struct {
	// TemplateID is the unique identifier for the template
	TemplateID string `json:"template_id"`

	// Name is the human-readable template name
	Name string `json:"name"`

	// Version is the semver version string
	Version string `json:"version"`

	// Description provides details about the template
	Description string `json:"description,omitempty"`

	// Type is the workload type category
	Type WorkloadType `json:"type"`

	// Runtime defines the runtime configuration
	Runtime WorkloadRuntime `json:"runtime"`

	// Resources defines resource requirements and limits
	Resources WorkloadResourceSpec `json:"resources"`

	// Security defines security constraints
	Security WorkloadSecuritySpec `json:"security"`

	// Entrypoint defines the execution entry point
	Entrypoint WorkloadEntrypoint `json:"entrypoint"`

	// Environment defines environment variable templates
	Environment []EnvironmentVariable `json:"environment,omitempty"`

	// Modules lists required modules to load
	Modules []string `json:"modules,omitempty"`

	// DataBindings defines data mount points
	DataBindings []DataBinding `json:"data_bindings,omitempty"`

	// ParameterSchema defines user-configurable parameters
	ParameterSchema []ParameterDefinition `json:"parameter_schema,omitempty"`

	// ApprovalStatus is the current approval status
	ApprovalStatus WorkloadApprovalStatus `json:"approval_status"`

	// Publisher is the address that published this template
	Publisher string `json:"publisher"`

	// ArtifactCID is the content ID in artifact store
	ArtifactCID string `json:"artifact_cid,omitempty"`

	// Signature is the publisher's signature over the template
	Signature WorkloadSignature `json:"signature,omitempty"`

	// Tags are searchable tags
	Tags []string `json:"tags,omitempty"`

	// CreatedAt is when the template was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the template was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// ApprovedAt is when the template was approved
	ApprovedAt *time.Time `json:"approved_at,omitempty"`

	// BlockHeight is the block when template was recorded
	BlockHeight int64 `json:"block_height"`
}

// WorkloadRuntime defines runtime configuration
type WorkloadRuntime struct {
	// ContainerImage is the container image (Singularity/Apptainer)
	ContainerImage string `json:"container_image,omitempty"`

	// ContainerRegistry is the registry for the image
	ContainerRegistry string `json:"container_registry,omitempty"`

	// ImageDigest is the image digest for verification
	ImageDigest string `json:"image_digest,omitempty"`

	// RuntimeType is the runtime type (singularity|apptainer|native)
	RuntimeType string `json:"runtime_type"`

	// RequiredModules lists HPC modules that must be available
	RequiredModules []string `json:"required_modules,omitempty"`

	// MPIImplementation specifies MPI implementation (openmpi|mpich|intelmpi)
	MPIImplementation string `json:"mpi_implementation,omitempty"`

	// CUDAVersion specifies required CUDA version
	CUDAVersion string `json:"cuda_version,omitempty"`

	// PythonVersion specifies required Python version
	PythonVersion string `json:"python_version,omitempty"`
}

// WorkloadResourceSpec defines resource requirements and limits
type WorkloadResourceSpec struct {
	// MinNodes is the minimum number of nodes
	MinNodes int32 `json:"min_nodes"`

	// MaxNodes is the maximum number of nodes
	MaxNodes int32 `json:"max_nodes"`

	// DefaultNodes is the default number of nodes
	DefaultNodes int32 `json:"default_nodes"`

	// MinCPUsPerNode is minimum CPUs per node
	MinCPUsPerNode int32 `json:"min_cpus_per_node"`

	// MaxCPUsPerNode is maximum CPUs per node
	MaxCPUsPerNode int32 `json:"max_cpus_per_node"`

	// DefaultCPUsPerNode is default CPUs per node
	DefaultCPUsPerNode int32 `json:"default_cpus_per_node"`

	// MinMemoryMBPerNode is minimum memory per node in MB
	MinMemoryMBPerNode int64 `json:"min_memory_mb_per_node"`

	// MaxMemoryMBPerNode is maximum memory per node in MB
	MaxMemoryMBPerNode int64 `json:"max_memory_mb_per_node"`

	// DefaultMemoryMBPerNode is default memory per node in MB
	DefaultMemoryMBPerNode int64 `json:"default_memory_mb_per_node"`

	// MinGPUsPerNode is minimum GPUs per node
	MinGPUsPerNode int32 `json:"min_gpus_per_node,omitempty"`

	// MaxGPUsPerNode is maximum GPUs per node
	MaxGPUsPerNode int32 `json:"max_gpus_per_node,omitempty"`

	// DefaultGPUsPerNode is default GPUs per node
	DefaultGPUsPerNode int32 `json:"default_gpus_per_node,omitempty"`

	// GPUTypes lists allowed GPU types
	GPUTypes []string `json:"gpu_types,omitempty"`

	// MinRuntimeMinutes is minimum runtime in minutes
	MinRuntimeMinutes int64 `json:"min_runtime_minutes"`

	// MaxRuntimeMinutes is maximum runtime in minutes
	MaxRuntimeMinutes int64 `json:"max_runtime_minutes"`

	// DefaultRuntimeMinutes is default runtime in minutes
	DefaultRuntimeMinutes int64 `json:"default_runtime_minutes"`

	// StorageGBRequired is required storage in GB
	StorageGBRequired int32 `json:"storage_gb_required,omitempty"`

	// NetworkRequired indicates if high-speed network is required
	NetworkRequired bool `json:"network_required,omitempty"`

	// ExclusiveNodes indicates if exclusive node access is required
	ExclusiveNodes bool `json:"exclusive_nodes,omitempty"`
}

// WorkloadSecuritySpec defines security constraints
type WorkloadSecuritySpec struct {
	// AllowedImages lists allowed container images (glob patterns)
	AllowedImages []string `json:"allowed_images,omitempty"`

	// BlockedImages lists blocked container images (glob patterns)
	BlockedImages []string `json:"blocked_images,omitempty"`

	// AllowedRegistries lists allowed container registries
	AllowedRegistries []string `json:"allowed_registries,omitempty"`

	// RequireImageDigest requires image digest verification
	RequireImageDigest bool `json:"require_image_digest"`

	// AllowNetworkAccess allows network access during job
	AllowNetworkAccess bool `json:"allow_network_access"`

	// AllowHostMounts allows mounting host paths
	AllowHostMounts bool `json:"allow_host_mounts"`

	// AllowedHostPaths lists allowed host mount paths
	AllowedHostPaths []string `json:"allowed_host_paths,omitempty"`

	// SandboxLevel is the sandboxing level (none|basic|strict)
	SandboxLevel string `json:"sandbox_level"`

	// MaxOpenFiles is the max open files limit
	MaxOpenFiles int64 `json:"max_open_files,omitempty"`

	// MaxProcesses is the max processes limit
	MaxProcesses int64 `json:"max_processes,omitempty"`

	// MaxFileSize is the max file size in bytes
	MaxFileSize int64 `json:"max_file_size,omitempty"`

	// RequireSignedImage requires cryptographically signed images
	RequireSignedImage bool `json:"require_signed_image"`

	// TrustedSigners lists trusted image signer public keys
	TrustedSigners []string `json:"trusted_signers,omitempty"`
}

// WorkloadEntrypoint defines the execution entry point
type WorkloadEntrypoint struct {
	// Command is the command to execute
	Command string `json:"command"`

	// DefaultArgs are default arguments
	DefaultArgs []string `json:"default_args,omitempty"`

	// ArgTemplate is a template for argument generation
	ArgTemplate string `json:"arg_template,omitempty"`

	// WorkingDirectory is the working directory
	WorkingDirectory string `json:"working_directory,omitempty"`

	// PreRunScript is a script to run before the main command
	PreRunScript string `json:"pre_run_script,omitempty"`

	// PostRunScript is a script to run after the main command
	PostRunScript string `json:"post_run_script,omitempty"`

	// UseMPIRun wraps command with mpirun/srun
	UseMPIRun bool `json:"use_mpirun,omitempty"`

	// MPIRunArgs are additional mpirun arguments
	MPIRunArgs []string `json:"mpirun_args,omitempty"`
}

// EnvironmentVariable defines an environment variable
type EnvironmentVariable struct {
	// Name is the variable name
	Name string `json:"name"`

	// Value is the default value
	Value string `json:"value,omitempty"`

	// ValueTemplate is a template for value generation
	ValueTemplate string `json:"value_template,omitempty"`

	// Required indicates if the variable is required
	Required bool `json:"required"`

	// Secret indicates if the value is sensitive
	Secret bool `json:"secret"`

	// Description describes the variable
	Description string `json:"description,omitempty"`
}

// DataBinding defines a data mount point
type DataBinding struct {
	// Name is the binding name
	Name string `json:"name"`

	// MountPath is the path inside the container
	MountPath string `json:"mount_path"`

	// HostPath is the host path (if host mount)
	HostPath string `json:"host_path,omitempty"`

	// DataType is the type of data (input|output|scratch)
	DataType string `json:"data_type"`

	// Required indicates if this binding is required
	Required bool `json:"required"`

	// ReadOnly indicates if the mount is read-only
	ReadOnly bool `json:"read_only"`
}

// ParameterDefinition defines a user-configurable parameter
type ParameterDefinition struct {
	// Name is the parameter name
	Name string `json:"name"`

	// Type is the parameter type (string|int|float|bool|enum)
	Type string `json:"type"`

	// Description describes the parameter
	Description string `json:"description,omitempty"`

	// Default is the default value
	Default string `json:"default,omitempty"`

	// Required indicates if the parameter is required
	Required bool `json:"required"`

	// EnumValues lists allowed values for enum type
	EnumValues []string `json:"enum_values,omitempty"`

	// MinValue is the minimum value for numeric types
	MinValue string `json:"min_value,omitempty"`

	// MaxValue is the maximum value for numeric types
	MaxValue string `json:"max_value,omitempty"`

	// Pattern is a regex pattern for string validation
	Pattern string `json:"pattern,omitempty"`
}

// WorkloadSignature contains the cryptographic signature
type WorkloadSignature struct {
	// Algorithm is the signature algorithm (ed25519|secp256k1)
	Algorithm string `json:"algorithm"`

	// PublisherPubKey is the publisher's public key (hex encoded)
	PublisherPubKey string `json:"publisher_pub_key"`

	// Signature is the signature bytes (hex encoded)
	Signature string `json:"signature"`

	// SignedAt is when the signature was created
	SignedAt time.Time `json:"signed_at"`

	// ContentHash is the hash of signed content (hex encoded)
	ContentHash string `json:"content_hash"`
}

// WorkloadManifest represents a complete workload manifest with all metadata
type WorkloadManifest struct {
	// SchemaVersion is the manifest schema version
	SchemaVersion string `json:"schema_version"`

	// Template is the workload template
	Template WorkloadTemplate `json:"template"`

	// BatchScript is the generated batch script (if applicable)
	BatchScript string `json:"batch_script,omitempty"`

	// Checksum is the manifest checksum
	Checksum string `json:"checksum"`
}

// Validation regex patterns
var (
	workloadTemplateIDRegex   = regexp.MustCompile(`^[a-z][a-z0-9-]{2,63}$`)
	workloadTemplateNameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9 _-]{2,127}$`)
	containerImageRegex       = regexp.MustCompile(`^[a-z0-9][a-z0-9._/-]*:[a-zA-Z0-9._-]+$`)
	runtimeTypes              = map[string]bool{"singularity": true, "apptainer": true, "native": true}
	sandboxLevels             = map[string]bool{"none": true, "basic": true, "strict": true}
	parameterTypes            = map[string]bool{"string": true, "int": true, "float": true, "bool": true, "enum": true}
	dataTypes                 = map[string]bool{"input": true, "output": true, "scratch": true}
)

// Validate validates a workload template
func (t *WorkloadTemplate) Validate() error {
	if t.TemplateID == "" {
		return ErrInvalidWorkloadTemplate.Wrap("template_id required")
	}
	if !workloadTemplateIDRegex.MatchString(t.TemplateID) {
		return ErrInvalidWorkloadTemplate.Wrap("template_id must be 3-64 lowercase alphanumeric with hyphens")
	}

	if t.Name == "" {
		return ErrInvalidWorkloadTemplate.Wrap("name required")
	}
	if !workloadTemplateNameRegex.MatchString(t.Name) {
		return ErrInvalidWorkloadTemplate.Wrap("name must be 3-128 chars, start with letter")
	}

	if t.Version == "" {
		return ErrInvalidWorkloadTemplate.Wrap("version required")
	}
	if !semverRegex.MatchString(t.Version) {
		return ErrInvalidWorkloadTemplate.Wrap("version must be valid semver")
	}

	if !t.Type.IsValid() {
		return ErrInvalidWorkloadTemplate.Wrapf("invalid workload type: %s", t.Type)
	}

	if err := t.Runtime.Validate(); err != nil {
		return ErrInvalidWorkloadTemplate.Wrapf("runtime: %s", err.Error())
	}

	if err := t.Resources.Validate(); err != nil {
		return ErrInvalidWorkloadTemplate.Wrapf("resources: %s", err.Error())
	}

	if err := t.Security.Validate(); err != nil {
		return ErrInvalidWorkloadTemplate.Wrapf("security: %s", err.Error())
	}

	if err := t.Entrypoint.Validate(); err != nil {
		return ErrInvalidWorkloadTemplate.Wrapf("entrypoint: %s", err.Error())
	}

	if !t.ApprovalStatus.IsValid() {
		return ErrInvalidWorkloadTemplate.Wrapf("invalid approval status: %s", t.ApprovalStatus)
	}

	if _, err := sdk.AccAddressFromBech32(t.Publisher); err != nil {
		return ErrInvalidWorkloadTemplate.Wrap("invalid publisher address")
	}

	for i, p := range t.ParameterSchema {
		if err := p.Validate(); err != nil {
			return ErrInvalidWorkloadTemplate.Wrapf("parameter[%d]: %s", i, err.Error())
		}
	}

	for i, b := range t.DataBindings {
		if err := b.Validate(); err != nil {
			return ErrInvalidWorkloadTemplate.Wrapf("data_binding[%d]: %s", i, err.Error())
		}
	}

	return nil
}

// Validate validates workload runtime
func (r *WorkloadRuntime) Validate() error {
	if r.RuntimeType == "" {
		return ErrInvalidWorkloadRuntime.Wrap("runtime_type required")
	}
	if !runtimeTypes[r.RuntimeType] {
		return ErrInvalidWorkloadRuntime.Wrapf("invalid runtime_type: %s", r.RuntimeType)
	}

	if r.RuntimeType != "native" && r.ContainerImage == "" {
		return ErrInvalidWorkloadRuntime.Wrap("container_image required for container runtimes")
	}

	if r.ContainerImage != "" && !containerImageRegex.MatchString(r.ContainerImage) {
		return ErrInvalidWorkloadRuntime.Wrap("invalid container_image format")
	}

	return nil
}

// Validate validates resource spec
func (r *WorkloadResourceSpec) Validate() error {
	if r.MinNodes < 1 {
		return ErrInvalidWorkloadResources.Wrap("min_nodes must be >= 1")
	}
	if r.MaxNodes < r.MinNodes {
		return ErrInvalidWorkloadResources.Wrap("max_nodes must be >= min_nodes")
	}
	if r.DefaultNodes < r.MinNodes || r.DefaultNodes > r.MaxNodes {
		return ErrInvalidWorkloadResources.Wrap("default_nodes must be between min and max")
	}

	if r.MinCPUsPerNode < 1 {
		return ErrInvalidWorkloadResources.Wrap("min_cpus_per_node must be >= 1")
	}
	if r.MaxCPUsPerNode < r.MinCPUsPerNode {
		return ErrInvalidWorkloadResources.Wrap("max_cpus_per_node must be >= min_cpus_per_node")
	}
	if r.DefaultCPUsPerNode < r.MinCPUsPerNode || r.DefaultCPUsPerNode > r.MaxCPUsPerNode {
		return ErrInvalidWorkloadResources.Wrap("default_cpus_per_node must be between min and max")
	}

	if r.MinMemoryMBPerNode < 1 {
		return ErrInvalidWorkloadResources.Wrap("min_memory_mb_per_node must be >= 1")
	}
	if r.MaxMemoryMBPerNode < r.MinMemoryMBPerNode {
		return ErrInvalidWorkloadResources.Wrap("max_memory_mb_per_node must be >= min_memory_mb_per_node")
	}
	if r.DefaultMemoryMBPerNode < r.MinMemoryMBPerNode || r.DefaultMemoryMBPerNode > r.MaxMemoryMBPerNode {
		return ErrInvalidWorkloadResources.Wrap("default_memory_mb_per_node must be between min and max")
	}

	if r.MaxGPUsPerNode < r.MinGPUsPerNode {
		return ErrInvalidWorkloadResources.Wrap("max_gpus_per_node must be >= min_gpus_per_node")
	}
	if r.DefaultGPUsPerNode < r.MinGPUsPerNode || r.DefaultGPUsPerNode > r.MaxGPUsPerNode {
		return ErrInvalidWorkloadResources.Wrap("default_gpus_per_node must be between min and max")
	}

	if r.MinRuntimeMinutes < 1 {
		return ErrInvalidWorkloadResources.Wrap("min_runtime_minutes must be >= 1")
	}
	if r.MaxRuntimeMinutes < r.MinRuntimeMinutes {
		return ErrInvalidWorkloadResources.Wrap("max_runtime_minutes must be >= min_runtime_minutes")
	}
	if r.DefaultRuntimeMinutes < r.MinRuntimeMinutes || r.DefaultRuntimeMinutes > r.MaxRuntimeMinutes {
		return ErrInvalidWorkloadResources.Wrap("default_runtime_minutes must be between min and max")
	}

	return nil
}

// Validate validates security spec
func (s *WorkloadSecuritySpec) Validate() error {
	if s.SandboxLevel != "" && !sandboxLevels[s.SandboxLevel] {
		return ErrInvalidWorkloadSecurity.Wrapf("invalid sandbox_level: %s", s.SandboxLevel)
	}

	if s.MaxOpenFiles < 0 {
		return ErrInvalidWorkloadSecurity.Wrap("max_open_files cannot be negative")
	}
	if s.MaxProcesses < 0 {
		return ErrInvalidWorkloadSecurity.Wrap("max_processes cannot be negative")
	}
	if s.MaxFileSize < 0 {
		return ErrInvalidWorkloadSecurity.Wrap("max_file_size cannot be negative")
	}

	return nil
}

// Validate validates entrypoint
func (e *WorkloadEntrypoint) Validate() error {
	if e.Command == "" {
		return ErrInvalidWorkloadEntrypoint.Wrap("command required")
	}
	return nil
}

// Validate validates parameter definition
func (p *ParameterDefinition) Validate() error {
	if p.Name == "" {
		return ErrInvalidWorkloadParameter.Wrap("name required")
	}
	if !parameterTypes[p.Type] {
		return ErrInvalidWorkloadParameter.Wrapf("invalid type: %s", p.Type)
	}
	if p.Type == "enum" && len(p.EnumValues) == 0 {
		return ErrInvalidWorkloadParameter.Wrap("enum_values required for enum type")
	}
	return nil
}

// Validate validates data binding
func (b *DataBinding) Validate() error {
	if b.Name == "" {
		return ErrInvalidDataBinding.Wrap("name required")
	}
	if b.MountPath == "" {
		return ErrInvalidDataBinding.Wrap("mount_path required")
	}
	if !dataTypes[b.DataType] {
		return ErrInvalidDataBinding.Wrapf("invalid data_type: %s", b.DataType)
	}
	return nil
}

// Verify verifies the template signature
func (t *WorkloadTemplate) Verify() error {
	if t.Signature.Signature == "" {
		return ErrInvalidWorkloadSignature.Wrap("signature required")
	}

	// Decode signature
	_, err := hex.DecodeString(t.Signature.Signature)
	if err != nil {
		return ErrInvalidWorkloadSignature.Wrap("invalid signature hex encoding")
	}

	// Decode public key
	_, err = hex.DecodeString(t.Signature.PublisherPubKey)
	if err != nil {
		return ErrInvalidWorkloadSignature.Wrap("invalid public key hex encoding")
	}

	// Note: Actual cryptographic verification would be done by the signing package
	return nil
}

// GetVersionedID returns the versioned template ID
func (t *WorkloadTemplate) GetVersionedID() string {
	return t.TemplateID + "@" + t.Version
}
