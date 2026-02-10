// Package marketplace provides types for the marketplace on-chain module.
//
// VE-36F: Standardized service types and specs for VM/container offerings.
package marketplace

import (
	"fmt"
	"strconv"
	"strings"
)

// ServiceType represents the standardized service type for provisioning.
type ServiceType string

const (
	// ServiceTypeUnknown represents an unspecified service type.
	ServiceTypeUnknown ServiceType = ""
	// ServiceTypeVM represents VM-based provisioning.
	ServiceTypeVM ServiceType = "vm"
	// ServiceTypeContainer represents container-based provisioning.
	ServiceTypeContainer ServiceType = "container"
)

// Specification keys for standardized service specs.
const (
	SpecKeyServiceType = "service_type"

	SpecKeyVMCPU         = "vm.cpu_cores"
	SpecKeyVMMemoryMB    = "vm.memory_mb"
	SpecKeyVMDiskGB      = "vm.disk_gb"
	SpecKeyVMImage       = "vm.image"
	SpecKeyVMSSHKey      = "vm.ssh_key"
	SpecKeyVMNetworkMbps = "vm.network_mbps"
	SpecKeyVMGPUCount    = "vm.gpu_count"
	SpecKeyVMGPUType     = "vm.gpu_type"
	SpecKeyVMFlavor      = "vm.flavor"

	SpecKeyContainerImage      = "container.image"
	SpecKeyContainerCPU        = "container.cpu_cores"
	SpecKeyContainerMemoryMB   = "container.memory_mb"
	SpecKeyContainerCommand    = "container.command"
	SpecKeyContainerArgs       = "container.args"
	SpecKeyContainerEnv        = "container.env"
	SpecKeyContainerPorts      = "container.ports"
	SpecKeyContainerWorkingDir = "container.working_dir"
)

// VMServiceSpec defines standardized VM provisioning specs.
type VMServiceSpec struct {
	CPU         int
	MemoryMB    int
	DiskGB      int
	Image       string
	SSHKey      string
	NetworkMbps int
	GPUCount    int
	GPUType     string
	Flavor      string
}

// ContainerServiceSpec defines standardized container provisioning specs.
type ContainerServiceSpec struct {
	Image      string
	CPU        int
	MemoryMB   int
	Command    []string
	Args       []string
	Env        map[string]string
	Ports      []int
	WorkingDir string
}

// ServiceTypeFromSpecs infers service type from specifications.
func ServiceTypeFromSpecs(specs map[string]string) ServiceType {
	if specs == nil {
		return ServiceTypeUnknown
	}
	if value := strings.TrimSpace(specs[SpecKeyServiceType]); value != "" {
		return ServiceType(strings.ToLower(value))
	}
	if specs[SpecKeyVMImage] != "" || specs[SpecKeyVMCPU] != "" {
		return ServiceTypeVM
	}
	if specs[SpecKeyContainerImage] != "" || specs[SpecKeyContainerCPU] != "" {
		return ServiceTypeContainer
	}
	return ServiceTypeUnknown
}

// ParseVMServiceSpec parses VM specs from the specifications map.
func ParseVMServiceSpec(specs map[string]string) (VMServiceSpec, error) {
	if specs == nil {
		return VMServiceSpec{}, fmt.Errorf("specifications are required")
	}
	spec := VMServiceSpec{
		Image:   strings.TrimSpace(specs[SpecKeyVMImage]),
		SSHKey:  strings.TrimSpace(specs[SpecKeyVMSSHKey]),
		GPUType: strings.TrimSpace(specs[SpecKeyVMGPUType]),
		Flavor:  strings.TrimSpace(specs[SpecKeyVMFlavor]),
	}

	var err error
	if spec.CPU, err = parsePositiveInt(specs[SpecKeyVMCPU], "cpu_cores"); err != nil {
		return VMServiceSpec{}, err
	}
	if spec.MemoryMB, err = parsePositiveInt(specs[SpecKeyVMMemoryMB], "memory_mb"); err != nil {
		return VMServiceSpec{}, err
	}
	if spec.DiskGB, err = parsePositiveInt(specs[SpecKeyVMDiskGB], "disk_gb"); err != nil {
		return VMServiceSpec{}, err
	}
	spec.NetworkMbps = parseOptionalInt(specs[SpecKeyVMNetworkMbps])
	spec.GPUCount = parseOptionalInt(specs[SpecKeyVMGPUCount])

	if spec.Image == "" && spec.Flavor == "" {
		return VMServiceSpec{}, fmt.Errorf("vm.image or vm.flavor is required")
	}

	return spec, nil
}

// ParseContainerServiceSpec parses container specs from the specifications map.
func ParseContainerServiceSpec(specs map[string]string) (ContainerServiceSpec, error) {
	if specs == nil {
		return ContainerServiceSpec{}, fmt.Errorf("specifications are required")
	}
	spec := ContainerServiceSpec{
		Image:      strings.TrimSpace(specs[SpecKeyContainerImage]),
		WorkingDir: strings.TrimSpace(specs[SpecKeyContainerWorkingDir]),
	}
	if spec.Image == "" {
		return ContainerServiceSpec{}, fmt.Errorf("container.image is required")
	}

	var err error
	if spec.CPU, err = parsePositiveInt(specs[SpecKeyContainerCPU], "container.cpu_cores"); err != nil {
		return ContainerServiceSpec{}, err
	}
	if spec.MemoryMB, err = parsePositiveInt(specs[SpecKeyContainerMemoryMB], "container.memory_mb"); err != nil {
		return ContainerServiceSpec{}, err
	}

	spec.Command = splitCSV(specs[SpecKeyContainerCommand])
	spec.Args = splitCSV(specs[SpecKeyContainerArgs])
	spec.Env = parseEnv(specs[SpecKeyContainerEnv])
	spec.Ports = parsePorts(specs[SpecKeyContainerPorts])

	return spec, nil
}

func parsePositiveInt(value string, field string) (int, error) {
	if strings.TrimSpace(value) == "" {
		return 0, fmt.Errorf("%s is required", field)
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", field)
	}
	return parsed, nil
}

func parseOptionalInt(value string) int {
	if strings.TrimSpace(value) == "" {
		return 0
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0
	}
	return parsed
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

func parseEnv(value string) map[string]string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	env := make(map[string]string)
	pairs := splitCSV(value)
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		if key == "" {
			continue
		}
		env[key] = strings.TrimSpace(kv[1])
	}
	if len(env) == 0 {
		return nil
	}
	return env
}

func parsePorts(value string) []int {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := splitCSV(value)
	ports := make([]int, 0, len(parts))
	for _, part := range parts {
		port, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || port <= 0 {
			continue
		}
		ports = append(ports, port)
	}
	return ports
}
