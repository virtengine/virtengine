// Package provider_daemon implements the provider daemon for VirtEngine.
//
// VE-402: Provider Daemon manifest parsing and validation
package provider_daemon

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ErrInvalidManifest is returned when a manifest is invalid
var ErrInvalidManifest = errors.New("invalid manifest")

// ErrUnsupportedManifestVersion is returned when manifest version is unsupported
var ErrUnsupportedManifestVersion = errors.New("unsupported manifest version")

// ErrUnsupportedParameter is returned when an unsupported parameter is used
var ErrUnsupportedParameter = errors.New("unsupported parameter")

// ManifestVersion represents the manifest version
type ManifestVersion string

const (
	// ManifestVersionV1 is version 1 of the manifest format
	ManifestVersionV1 ManifestVersion = "v1"

	// ManifestVersionV2Beta is version 2 beta of the manifest format
	ManifestVersionV2Beta ManifestVersion = "v2beta1"
)

// Manifest represents a workload deployment manifest
type Manifest struct {
	// Version is the manifest version
	Version ManifestVersion `json:"version"`

	// Name is the deployment name
	Name string `json:"name"`

	// Services contains the service definitions
	Services []ServiceSpec `json:"services"`

	// Networks contains network definitions
	Networks []NetworkSpec `json:"networks,omitempty"`

	// Volumes contains volume definitions
	Volumes []VolumeSpec `json:"volumes,omitempty"`

	// Constraints contains deployment constraints
	Constraints *DeploymentConstraints `json:"constraints,omitempty"`

	// Lifecycle contains lifecycle hooks
	Lifecycle *LifecycleHooks `json:"lifecycle,omitempty"`

	// Metadata contains optional metadata
	Metadata map[string]string `json:"metadata,omitempty"`
}

// ServiceSpec defines a service/workload
type ServiceSpec struct {
	// Name is the service name
	Name string `json:"name"`

	// Type is the service type (container, vm)
	Type string `json:"type"`

	// Image is the container/VM image
	Image string `json:"image"`

	// Tag is the image tag
	Tag string `json:"tag,omitempty"`

	// Command is the command to run
	Command []string `json:"command,omitempty"`

	// Args are arguments to the command
	Args []string `json:"args,omitempty"`

	// Env contains environment variables
	Env map[string]string `json:"env,omitempty"`

	// Resources specifies resource requirements
	Resources ResourceSpec `json:"resources"`

	// Ports defines exposed ports
	Ports []PortSpec `json:"ports,omitempty"`

	// Volumes defines volume mounts
	Volumes []VolumeMountSpec `json:"volumes,omitempty"`

	// NetworkRefs references networks to attach
	NetworkRefs []string `json:"network_refs,omitempty"`

	// Replicas is the number of replicas
	Replicas int32 `json:"replicas,omitempty"`

	// RestartPolicy defines restart behavior
	RestartPolicy string `json:"restart_policy,omitempty"`

	// HealthCheck defines health check configuration
	HealthCheck *HealthCheckSpec `json:"health_check,omitempty"`
}

// ResourceSpec specifies resource requirements
type ResourceSpec struct {
	// CPU in millicores
	CPU int64 `json:"cpu"`

	// Memory in bytes
	Memory int64 `json:"memory"`

	// Storage in bytes (ephemeral)
	Storage int64 `json:"storage,omitempty"`

	// GPU count
	GPU int64 `json:"gpu,omitempty"`

	// GPUType specifies required GPU type
	GPUType string `json:"gpu_type,omitempty"`
}

// PortSpec defines a port exposure
type PortSpec struct {
	// Name is the port name
	Name string `json:"name,omitempty"`

	// ContainerPort is the container port
	ContainerPort int32 `json:"container_port"`

	// Protocol is the protocol (tcp, udp)
	Protocol string `json:"protocol,omitempty"`

	// Expose indicates if the port should be externally accessible
	Expose bool `json:"expose,omitempty"`

	// ExternalPort is the external port (if different)
	ExternalPort int32 `json:"external_port,omitempty"`
}

// VolumeMountSpec defines a volume mount
type VolumeMountSpec struct {
	// Name references a volume name
	Name string `json:"name"`

	// MountPath is the mount path in the container
	MountPath string `json:"mount_path"`

	// ReadOnly indicates if mount is read-only
	ReadOnly bool `json:"read_only,omitempty"`

	// SubPath is a sub-path within the volume
	SubPath string `json:"sub_path,omitempty"`
}

// NetworkSpec defines a network
type NetworkSpec struct {
	// Name is the network name
	Name string `json:"name"`

	// Type is the network type (private, public)
	Type string `json:"type"`

	// CIDR is the network CIDR (optional)
	CIDR string `json:"cidr,omitempty"`
}

// VolumeSpec defines a volume
type VolumeSpec struct {
	// Name is the volume name
	Name string `json:"name"`

	// Type is the volume type (persistent, ephemeral)
	Type string `json:"type"`

	// Size is the volume size in bytes
	Size int64 `json:"size"`

	// StorageClass is the storage class (optional)
	StorageClass string `json:"storage_class,omitempty"`
}

// DeploymentConstraints specifies deployment constraints
type DeploymentConstraints struct {
	// Region is the required region
	Region string `json:"region,omitempty"`

	// Regions is a list of acceptable regions
	Regions []string `json:"regions,omitempty"`

	// MaxLatencyMs is the maximum acceptable latency in ms
	MaxLatencyMs int64 `json:"max_latency_ms,omitempty"`

	// RequiredTags are tags the provider must have
	RequiredTags []string `json:"required_tags,omitempty"`

	// ExcludedTags are tags the provider must not have
	ExcludedTags []string `json:"excluded_tags,omitempty"`

	// Affinity specifies affinity rules
	Affinity []AffinityRule `json:"affinity,omitempty"`

	// AntiAffinity specifies anti-affinity rules
	AntiAffinity []AffinityRule `json:"anti_affinity,omitempty"`
}

// AffinityRule defines an affinity rule
type AffinityRule struct {
	// Type is the affinity type (host, zone, region)
	Type string `json:"type"`

	// Key is the label key
	Key string `json:"key,omitempty"`

	// Value is the label value
	Value string `json:"value,omitempty"`
}

// LifecycleHooks defines lifecycle hooks
type LifecycleHooks struct {
	// PostStart runs after the workload starts
	PostStart *LifecycleHook `json:"post_start,omitempty"`

	// PreStop runs before the workload stops
	PreStop *LifecycleHook `json:"pre_stop,omitempty"`
}

// LifecycleHook defines a lifecycle hook
type LifecycleHook struct {
	// Exec runs a command
	Exec *ExecAction `json:"exec,omitempty"`

	// HTTP performs an HTTP request
	HTTP *HTTPAction `json:"http,omitempty"`

	// TimeoutSeconds is the timeout
	TimeoutSeconds int32 `json:"timeout_seconds,omitempty"`
}

// ExecAction defines a command execution
type ExecAction struct {
	// Command is the command to run
	Command []string `json:"command"`
}

// HTTPAction defines an HTTP request
type HTTPAction struct {
	// Path is the HTTP path
	Path string `json:"path"`

	// Port is the target port
	Port int32 `json:"port"`

	// Scheme is http or https
	Scheme string `json:"scheme,omitempty"`
}

// HealthCheckSpec defines health check configuration
type HealthCheckSpec struct {
	// Exec runs a command
	Exec *ExecAction `json:"exec,omitempty"`

	// HTTP performs an HTTP request
	HTTP *HTTPAction `json:"http,omitempty"`

	// TCP performs a TCP check
	TCP *TCPAction `json:"tcp,omitempty"`

	// InitialDelaySeconds is the initial delay
	InitialDelaySeconds int32 `json:"initial_delay_seconds,omitempty"`

	// PeriodSeconds is the check period
	PeriodSeconds int32 `json:"period_seconds,omitempty"`

	// TimeoutSeconds is the check timeout
	TimeoutSeconds int32 `json:"timeout_seconds,omitempty"`

	// FailureThreshold is failures before unhealthy
	FailureThreshold int32 `json:"failure_threshold,omitempty"`

	// SuccessThreshold is successes before healthy
	SuccessThreshold int32 `json:"success_threshold,omitempty"`
}

// TCPAction defines a TCP check
type TCPAction struct {
	// Port is the port to check
	Port int32 `json:"port"`
}

// ValidationResult contains the result of manifest validation
type ValidationResult struct {
	// Valid indicates if the manifest is valid
	Valid bool `json:"valid"`

	// Errors contains validation errors
	Errors []ValidationError `json:"errors,omitempty"`

	// Warnings contains validation warnings
	Warnings []string `json:"warnings,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	// Field is the field with the error
	Field string `json:"field"`

	// Message is the error message
	Message string `json:"message"`

	// Code is an error code
	Code string `json:"code"`
}

// ManifestParser parses and validates manifests
type ManifestParser struct {
	// supportedVersions lists supported manifest versions
	supportedVersions map[ManifestVersion]bool

	// supportedServiceTypes lists supported service types
	supportedServiceTypes map[string]bool

	// supportedVolumeTypes lists supported volume types
	supportedVolumeTypes map[string]bool

	// maxServiceCount limits services per manifest
	maxServiceCount int

	// maxReplicaCount limits replicas per service
	maxReplicaCount int32
}

// NewManifestParser creates a new manifest parser with default settings
func NewManifestParser() *ManifestParser {
	return &ManifestParser{
		supportedVersions: map[ManifestVersion]bool{
			ManifestVersionV1: true,
		},
		supportedServiceTypes: map[string]bool{
			"container": true,
			"vm":        true,
		},
		supportedVolumeTypes: map[string]bool{
			"persistent": true,
			"ephemeral":  true,
		},
		maxServiceCount: 100,
		maxReplicaCount: 1000,
	}
}

// Parse parses a manifest from JSON
func (mp *ManifestParser) Parse(data []byte) (*Manifest, error) {
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("%w: failed to parse JSON: %v", ErrInvalidManifest, err)
	}
	return &manifest, nil
}

// ParseYAML parses a manifest from YAML (simplified - just uses JSON for now)
func (mp *ManifestParser) ParseYAML(data []byte) (*Manifest, error) {
	// In a real implementation, this would use a YAML parser
	// For now, we assume YAML is already converted to JSON
	return mp.Parse(data)
}

// Validate validates a manifest and returns detailed results
func (mp *ManifestParser) Validate(manifest *Manifest) ValidationResult {
	result := ValidationResult{
		Valid:    true,
		Errors:   make([]ValidationError, 0),
		Warnings: make([]string, 0),
	}

	// Validate version
	if manifest.Version == "" {
		result.addError("version", "manifest version is required", "REQUIRED_FIELD")
	} else if !mp.supportedVersions[manifest.Version] {
		result.addError("version", fmt.Sprintf("unsupported manifest version: %s", manifest.Version), "UNSUPPORTED_VERSION")
	}

	// Validate name
	if manifest.Name == "" {
		result.addError("name", "manifest name is required", "REQUIRED_FIELD")
	} else if !isValidName(manifest.Name) {
		result.addError("name", "invalid manifest name: must be alphanumeric with dashes", "INVALID_NAME")
	}

	// Validate services
	if len(manifest.Services) == 0 {
		result.addError("services", "at least one service is required", "REQUIRED_FIELD")
	} else if len(manifest.Services) > mp.maxServiceCount {
		result.addError("services", fmt.Sprintf("too many services: max %d", mp.maxServiceCount), "MAX_EXCEEDED")
	}

	serviceNames := make(map[string]bool)
	for i, svc := range manifest.Services {
		mp.validateService(&result, &svc, i, serviceNames)
	}

	// Validate networks
	networkNames := make(map[string]bool)
	for i, net := range manifest.Networks {
		mp.validateNetwork(&result, &net, i, networkNames)
	}

	// Validate volumes
	volumeNames := make(map[string]bool)
	for i, vol := range manifest.Volumes {
		mp.validateVolume(&result, &vol, i, volumeNames)
	}

	// Cross-validate: check volume references exist
	for i, svc := range manifest.Services {
		for j, mount := range svc.Volumes {
			if !volumeNames[mount.Name] {
				result.addError(
					fmt.Sprintf("services[%d].volumes[%d].name", i, j),
					fmt.Sprintf("volume '%s' not defined", mount.Name),
					"UNDEFINED_REFERENCE",
				)
			}
		}
	}

	// Cross-validate: check network references exist
	for i, svc := range manifest.Services {
		for j, netRef := range svc.NetworkRefs {
			if !networkNames[netRef] {
				result.addError(
					fmt.Sprintf("services[%d].network_refs[%d]", i, j),
					fmt.Sprintf("network '%s' not defined", netRef),
					"UNDEFINED_REFERENCE",
				)
			}
		}
	}

	// Validate constraints
	if manifest.Constraints != nil {
		mp.validateConstraints(&result, manifest.Constraints)
	}

	// Validate lifecycle hooks
	if manifest.Lifecycle != nil {
		mp.validateLifecycle(&result, manifest.Lifecycle)
	}

	result.Valid = len(result.Errors) == 0
	return result
}

func (mp *ManifestParser) validateService(result *ValidationResult, svc *ServiceSpec, index int, names map[string]bool) {
	prefix := fmt.Sprintf("services[%d]", index)

	// Name
	if svc.Name == "" {
		result.addError(prefix+".name", "service name is required", "REQUIRED_FIELD")
	} else {
		if names[svc.Name] {
			result.addError(prefix+".name", fmt.Sprintf("duplicate service name: %s", svc.Name), "DUPLICATE_NAME")
		}
		names[svc.Name] = true

		if !isValidName(svc.Name) {
			result.addError(prefix+".name", "invalid service name: must be alphanumeric with dashes", "INVALID_NAME")
		}
	}

	// Type
	if svc.Type == "" {
		result.addError(prefix+".type", "service type is required", "REQUIRED_FIELD")
	} else if !mp.supportedServiceTypes[svc.Type] {
		result.addError(prefix+".type", fmt.Sprintf("unsupported service type: %s", svc.Type), "UNSUPPORTED_TYPE")
	}

	// Image
	if svc.Image == "" {
		result.addError(prefix+".image", "service image is required", "REQUIRED_FIELD")
	}

	// Resources
	if svc.Resources.CPU <= 0 {
		result.addError(prefix+".resources.cpu", "CPU must be positive", "INVALID_VALUE")
	}
	if svc.Resources.Memory <= 0 {
		result.addError(prefix+".resources.memory", "memory must be positive", "INVALID_VALUE")
	}
	if svc.Resources.GPU < 0 {
		result.addError(prefix+".resources.gpu", "GPU cannot be negative", "INVALID_VALUE")
	}

	// Replicas
	if svc.Replicas < 0 {
		result.addError(prefix+".replicas", "replicas cannot be negative", "INVALID_VALUE")
	} else if svc.Replicas > mp.maxReplicaCount {
		result.addError(prefix+".replicas", fmt.Sprintf("replicas exceeds maximum: %d", mp.maxReplicaCount), "MAX_EXCEEDED")
	}

	// Default replicas to 1
	if svc.Replicas == 0 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("%s.replicas defaults to 1", prefix))
	}

	// Ports
	for j, port := range svc.Ports {
		portPrefix := fmt.Sprintf("%s.ports[%d]", prefix, j)
		if port.ContainerPort <= 0 || port.ContainerPort > 65535 {
			result.addError(portPrefix+".container_port", "invalid port number", "INVALID_VALUE")
		}
		if port.Protocol != "" && port.Protocol != "tcp" && port.Protocol != "udp" {
			result.addError(portPrefix+".protocol", "protocol must be tcp or udp", "INVALID_VALUE")
		}
	}

	// Restart policy
	validPolicies := map[string]bool{"always": true, "on-failure": true, "never": true, "": true}
	if !validPolicies[svc.RestartPolicy] {
		result.addError(prefix+".restart_policy", "invalid restart policy", "INVALID_VALUE")
	}

	// Health check
	if svc.HealthCheck != nil {
		mp.validateHealthCheck(result, svc.HealthCheck, prefix+".health_check")
	}
}

func (mp *ManifestParser) validateNetwork(result *ValidationResult, net *NetworkSpec, index int, names map[string]bool) {
	prefix := fmt.Sprintf("networks[%d]", index)

	if net.Name == "" {
		result.addError(prefix+".name", "network name is required", "REQUIRED_FIELD")
	} else {
		if names[net.Name] {
			result.addError(prefix+".name", fmt.Sprintf("duplicate network name: %s", net.Name), "DUPLICATE_NAME")
		}
		names[net.Name] = true
	}

	if net.Type != "private" && net.Type != "public" {
		result.addError(prefix+".type", "network type must be private or public", "INVALID_VALUE")
	}
}

func (mp *ManifestParser) validateVolume(result *ValidationResult, vol *VolumeSpec, index int, names map[string]bool) {
	prefix := fmt.Sprintf("volumes[%d]", index)

	if vol.Name == "" {
		result.addError(prefix+".name", "volume name is required", "REQUIRED_FIELD")
	} else {
		if names[vol.Name] {
			result.addError(prefix+".name", fmt.Sprintf("duplicate volume name: %s", vol.Name), "DUPLICATE_NAME")
		}
		names[vol.Name] = true
	}

	if !mp.supportedVolumeTypes[vol.Type] {
		result.addError(prefix+".type", fmt.Sprintf("unsupported volume type: %s", vol.Type), "UNSUPPORTED_TYPE")
	}

	if vol.Size <= 0 {
		result.addError(prefix+".size", "volume size must be positive", "INVALID_VALUE")
	}
}

func (mp *ManifestParser) validateConstraints(result *ValidationResult, constraints *DeploymentConstraints) {
	if constraints.MaxLatencyMs < 0 {
		result.addError("constraints.max_latency_ms", "max latency cannot be negative", "INVALID_VALUE")
	}

	// Validate affinity rules
	for i, rule := range constraints.Affinity {
		validTypes := map[string]bool{"host": true, "zone": true, "region": true}
		if !validTypes[rule.Type] {
			result.addError(
				fmt.Sprintf("constraints.affinity[%d].type", i),
				"invalid affinity type",
				"INVALID_VALUE",
			)
		}
	}

	for i, rule := range constraints.AntiAffinity {
		validTypes := map[string]bool{"host": true, "zone": true, "region": true}
		if !validTypes[rule.Type] {
			result.addError(
				fmt.Sprintf("constraints.anti_affinity[%d].type", i),
				"invalid anti-affinity type",
				"INVALID_VALUE",
			)
		}
	}
}

func (mp *ManifestParser) validateLifecycle(result *ValidationResult, lifecycle *LifecycleHooks) {
	if lifecycle.PostStart != nil {
		mp.validateLifecycleHook(result, lifecycle.PostStart, "lifecycle.post_start")
	}
	if lifecycle.PreStop != nil {
		mp.validateLifecycleHook(result, lifecycle.PreStop, "lifecycle.pre_stop")
	}
}

func (mp *ManifestParser) validateLifecycleHook(result *ValidationResult, hook *LifecycleHook, prefix string) {
	// At least one action must be specified
	if hook.Exec == nil && hook.HTTP == nil {
		result.addError(prefix, "lifecycle hook must specify exec or http", "REQUIRED_FIELD")
	}

	// Both cannot be specified
	if hook.Exec != nil && hook.HTTP != nil {
		result.addError(prefix, "lifecycle hook cannot specify both exec and http", "INVALID_VALUE")
	}

	if hook.TimeoutSeconds < 0 {
		result.addError(prefix+".timeout_seconds", "timeout cannot be negative", "INVALID_VALUE")
	}

	if hook.Exec != nil && len(hook.Exec.Command) == 0 {
		result.addError(prefix+".exec.command", "exec command is required", "REQUIRED_FIELD")
	}

	if hook.HTTP != nil {
		if hook.HTTP.Port <= 0 || hook.HTTP.Port > 65535 {
			result.addError(prefix+".http.port", "invalid port number", "INVALID_VALUE")
		}
	}
}

func (mp *ManifestParser) validateHealthCheck(result *ValidationResult, check *HealthCheckSpec, prefix string) {
	// At least one probe type must be specified
	count := 0
	if check.Exec != nil {
		count++
	}
	if check.HTTP != nil {
		count++
	}
	if check.TCP != nil {
		count++
	}

	if count == 0 {
		result.addError(prefix, "health check must specify exec, http, or tcp", "REQUIRED_FIELD")
	}
	if count > 1 {
		result.addError(prefix, "health check cannot specify multiple probe types", "INVALID_VALUE")
	}

	if check.InitialDelaySeconds < 0 {
		result.addError(prefix+".initial_delay_seconds", "initial delay cannot be negative", "INVALID_VALUE")
	}
	if check.PeriodSeconds < 0 {
		result.addError(prefix+".period_seconds", "period cannot be negative", "INVALID_VALUE")
	}
	if check.TimeoutSeconds < 0 {
		result.addError(prefix+".timeout_seconds", "timeout cannot be negative", "INVALID_VALUE")
	}
}

func (result *ValidationResult) addError(field, message, code string) {
	result.Errors = append(result.Errors, ValidationError{
		Field:   field,
		Message: message,
		Code:    code,
	})
	result.Valid = false
}

// isValidName checks if a name is valid (alphanumeric with dashes)
func isValidName(name string) bool {
	if len(name) == 0 || len(name) > 253 {
		return false
	}
	match, _ := regexp.MatchString(`^[a-z0-9][a-z0-9-]*[a-z0-9]$|^[a-z0-9]$`, strings.ToLower(name))
	return match
}

// GetVersion returns the manifest version
func (m *Manifest) GetVersion() ManifestVersion {
	return m.Version
}

// TotalResources calculates the total resources required by the manifest
func (m *Manifest) TotalResources() ResourceSpec {
	total := ResourceSpec{}
	for _, svc := range m.Services {
		replicas := svc.Replicas
		if replicas == 0 {
			replicas = 1
		}
		total.CPU += svc.Resources.CPU * int64(replicas)
		total.Memory += svc.Resources.Memory * int64(replicas)
		total.Storage += svc.Resources.Storage * int64(replicas)
		total.GPU += svc.Resources.GPU * int64(replicas)
	}
	return total
}

// ServiceCount returns the number of services
func (m *Manifest) ServiceCount() int {
	return len(m.Services)
}

