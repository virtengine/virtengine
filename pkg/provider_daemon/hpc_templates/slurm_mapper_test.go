// Package hpc_templates provides workload template resolution for HPC jobs.
//
// VE-5F: Tests for SLURM job mapper
package hpc_templates

import (
	"strings"
	"testing"

	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// TestMapToSlurmJob tests basic SLURM job mapping
func TestMapToSlurmJob(t *testing.T) {
	mapper := NewSlurmJobMapper("gpu", "normal", "/modules")

	template := &hpctypes.WorkloadTemplate{
		TemplateID:  "test-template",
		Name:        "Test Template",
		Version:     "1.0.0",
		Description: "Test workload template",
		Type:        hpctypes.WorkloadTypeMPI,
		Runtime: hpctypes.WorkloadRuntime{
			RuntimeType: "native",
		},
		Resources: hpctypes.WorkloadResourceSpec{
			MinNodes:               1,
			MaxNodes:               10,
			DefaultNodes:           2,
			MinCPUsPerNode:         1,
			MaxCPUsPerNode:         32,
			DefaultCPUsPerNode:     8,
			MinMemoryMBPerNode:     1024,
			MaxMemoryMBPerNode:     65536,
			DefaultMemoryMBPerNode: 8192,
			MinRuntimeMinutes:      1,
			MaxRuntimeMinutes:      1440,
			DefaultRuntimeMinutes:  60,
		},
		Security: hpctypes.WorkloadSecuritySpec{
			SandboxLevel:       "basic",
			AllowNetworkAccess: true,
		},
		Entrypoint: hpctypes.WorkloadEntrypoint{
			Command:     "/usr/bin/myapp",
			DefaultArgs: []string{"--input", "data.txt"},
		},
		Publisher:      "cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a", // Valid bech32 address
		ApprovalStatus: hpctypes.WorkloadApprovalApproved,
	}

	// Test with default parameters
	resolved, err := mapper.MapToSlurmJob(template, nil)
	if err != nil {
		t.Fatalf("MapToSlurmJob failed: %v", err)
	}

	if resolved == nil {
		t.Fatal("resolved job is nil")
	}

	// Check template ID and version
	if resolved.TemplateID != "test-template" {
		t.Errorf("expected template ID test-template, got %s", resolved.TemplateID)
	}

	if resolved.TemplateVersion != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", resolved.TemplateVersion)
	}

	// Check resources match defaults
	if resolved.Resources.Nodes != 2 {
		t.Errorf("expected 2 nodes, got %d", resolved.Resources.Nodes)
	}

	if resolved.Resources.CPUsPerNode != 8 {
		t.Errorf("expected 8 CPUs per node, got %d", resolved.Resources.CPUsPerNode)
	}

	if resolved.Resources.MemoryMBPerNode != 8192 {
		t.Errorf("expected 8192 MB memory per node, got %d", resolved.Resources.MemoryMBPerNode)
	}

	if resolved.Resources.RuntimeMinutes != 60 {
		t.Errorf("expected 60 minute runtime, got %d", resolved.Resources.RuntimeMinutes)
	}

	// Check SLURM script is generated
	if resolved.SlurmScript == "" {
		t.Error("SLURM script is empty")
	}

	// Check script contains expected directives
	script := resolved.SlurmScript
	if !strings.Contains(script, "#SBATCH --nodes=2") {
		t.Error("SLURM script missing nodes directive")
	}

	if !strings.Contains(script, "#SBATCH --ntasks-per-node=8") {
		t.Error("SLURM script missing ntasks-per-node directive")
	}

	if !strings.Contains(script, "#SBATCH --mem=8192M") {
		t.Error("SLURM script missing mem directive")
	}

	if !strings.Contains(script, "#SBATCH --time=60") {
		t.Error("SLURM script missing time directive")
	}

	if !strings.Contains(script, "#SBATCH --partition=gpu") {
		t.Error("SLURM script missing partition directive")
	}

	if !strings.Contains(script, "/usr/bin/myapp --input data.txt") {
		t.Error("SLURM script missing command")
	}
}

// TestMapToSlurmJobWithOverrides tests SLURM job mapping with resource overrides
func TestMapToSlurmJobWithOverrides(t *testing.T) {
	mapper := NewSlurmJobMapper("normal", "standard", "/modules")

	template := &hpctypes.WorkloadTemplate{
		TemplateID: "test-template",
		Name:       "Test Template",
		Version:    "1.0.0",
		Type:       hpctypes.WorkloadTypeBatch,
		Runtime: hpctypes.WorkloadRuntime{
			RuntimeType: "native",
		},
		Resources: hpctypes.WorkloadResourceSpec{
			MinNodes:               1,
			MaxNodes:               10,
			DefaultNodes:           2,
			MinCPUsPerNode:         1,
			MaxCPUsPerNode:         32,
			DefaultCPUsPerNode:     8,
			MinMemoryMBPerNode:     1024,
			MaxMemoryMBPerNode:     65536,
			DefaultMemoryMBPerNode: 8192,
			MinRuntimeMinutes:      1,
			MaxRuntimeMinutes:      1440,
			DefaultRuntimeMinutes:  60,
		},
		Security: hpctypes.WorkloadSecuritySpec{
			SandboxLevel:       "basic",
			AllowNetworkAccess: true,
		},
		Entrypoint: hpctypes.WorkloadEntrypoint{
			Command: "/usr/bin/batch-processor",
		},
		Publisher:      "cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a",
		ApprovalStatus: hpctypes.WorkloadApprovalApproved,
	}

	// User overrides
	nodes := int32(4)
	cpus := int32(16)
	memory := int64(16384)
	runtime := int64(120)

	userParams := &UserParameters{
		Resources: &UserResourceOverrides{
			Nodes:           &nodes,
			CPUsPerNode:     &cpus,
			MemoryMBPerNode: &memory,
			RuntimeMinutes:  &runtime,
		},
	}

	resolved, err := mapper.MapToSlurmJob(template, userParams)
	if err != nil {
		t.Fatalf("MapToSlurmJob failed: %v", err)
	}

	// Check overridden resources
	if resolved.Resources.Nodes != 4 {
		t.Errorf("expected 4 nodes, got %d", resolved.Resources.Nodes)
	}

	if resolved.Resources.CPUsPerNode != 16 {
		t.Errorf("expected 16 CPUs per node, got %d", resolved.Resources.CPUsPerNode)
	}

	if resolved.Resources.MemoryMBPerNode != 16384 {
		t.Errorf("expected 16384 MB memory per node, got %d", resolved.Resources.MemoryMBPerNode)
	}

	if resolved.Resources.RuntimeMinutes != 120 {
		t.Errorf("expected 120 minute runtime, got %d", resolved.Resources.RuntimeMinutes)
	}

	// Check script reflects overrides
	script := resolved.SlurmScript
	if !strings.Contains(script, "#SBATCH --nodes=4") {
		t.Error("SLURM script missing overridden nodes directive")
	}

	if !strings.Contains(script, "#SBATCH --ntasks-per-node=16") {
		t.Error("SLURM script missing overridden ntasks-per-node directive")
	}
}

// TestMapToSlurmJobWithGPU tests SLURM job mapping with GPU resources
func TestMapToSlurmJobWithGPU(t *testing.T) {
	mapper := NewSlurmJobMapper("gpu", "high", "/modules")

	template := &hpctypes.WorkloadTemplate{
		TemplateID: "gpu-template",
		Name:       "GPU Template",
		Version:    "1.0.0",
		Type:       hpctypes.WorkloadTypeGPU,
		Runtime: hpctypes.WorkloadRuntime{
			RuntimeType: "native",
			CUDAVersion: "11.8",
		},
		Resources: hpctypes.WorkloadResourceSpec{
			MinNodes:               1,
			MaxNodes:               4,
			DefaultNodes:           1,
			MinCPUsPerNode:         4,
			MaxCPUsPerNode:         16,
			DefaultCPUsPerNode:     8,
			MinMemoryMBPerNode:     8192,
			MaxMemoryMBPerNode:     65536,
			DefaultMemoryMBPerNode: 16384,
			MinGPUsPerNode:         1,
			MaxGPUsPerNode:         4,
			DefaultGPUsPerNode:     2,
			GPUTypes:               []string{"v100", "a100"},
			MinRuntimeMinutes:      1,
			MaxRuntimeMinutes:      480,
			DefaultRuntimeMinutes:  60,
		},
		Security: hpctypes.WorkloadSecuritySpec{
			SandboxLevel:       "basic",
			AllowNetworkAccess: true,
		},
		Entrypoint: hpctypes.WorkloadEntrypoint{
			Command: "/usr/bin/gpu-app",
		},
		Modules:        []string{"cuda/11.8", "cudnn/8.0"},
		Publisher:      "cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a",
		ApprovalStatus: hpctypes.WorkloadApprovalApproved,
	}

	resolved, err := mapper.MapToSlurmJob(template, nil)
	if err != nil {
		t.Fatalf("MapToSlurmJob failed: %v", err)
	}

	// Check GPU resources
	if resolved.Resources.GPUsPerNode != 2 {
		t.Errorf("expected 2 GPUs per node, got %d", resolved.Resources.GPUsPerNode)
	}

	// Check script contains GPU directive
	script := resolved.SlurmScript
	if !strings.Contains(script, "#SBATCH --gres=gpu:") {
		t.Error("SLURM script missing GPU directive")
	}

	// Check modules are loaded
	if !strings.Contains(script, "module load cuda/11.8") {
		t.Error("SLURM script missing CUDA module load")
	}

	if !strings.Contains(script, "module load cudnn/8.0") {
		t.Error("SLURM script missing cuDNN module load")
	}
}

// TestMapToSlurmJobWithMPI tests SLURM job mapping with MPI
func TestMapToSlurmJobWithMPI(t *testing.T) {
	mapper := NewSlurmJobMapper("normal", "standard", "/modules")

	template := &hpctypes.WorkloadTemplate{
		TemplateID: "mpi-template",
		Name:       "MPI Template",
		Version:    "1.0.0",
		Type:       hpctypes.WorkloadTypeMPI,
		Runtime: hpctypes.WorkloadRuntime{
			RuntimeType:       "native",
			MPIImplementation: "openmpi",
		},
		Resources: hpctypes.WorkloadResourceSpec{
			MinNodes:               2,
			MaxNodes:               100,
			DefaultNodes:           4,
			MinCPUsPerNode:         1,
			MaxCPUsPerNode:         32,
			DefaultCPUsPerNode:     16,
			MinMemoryMBPerNode:     2048,
			MaxMemoryMBPerNode:     65536,
			DefaultMemoryMBPerNode: 16384,
			MinRuntimeMinutes:      1,
			MaxRuntimeMinutes:      1440,
			DefaultRuntimeMinutes:  120,
		},
		Security: hpctypes.WorkloadSecuritySpec{
			SandboxLevel:       "basic",
			AllowNetworkAccess: true,
		},
		Entrypoint: hpctypes.WorkloadEntrypoint{
			Command:    "/usr/bin/mpi-app",
			UseMPIRun:  true,
			MPIRunArgs: []string{"--bind-to", "core"},
		},
		Modules:        []string{"openmpi/4.1.0"},
		Publisher:      "cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a",
		ApprovalStatus: hpctypes.WorkloadApprovalApproved,
	}

	resolved, err := mapper.MapToSlurmJob(template, nil)
	if err != nil {
		t.Fatalf("MapToSlurmJob failed: %v", err)
	}

	// Check MPI configuration
	script := resolved.SlurmScript
	if !strings.Contains(script, "module load openmpi/4.1.0") {
		t.Error("SLURM script missing OpenMPI module load")
	}

	if !strings.Contains(script, "srun") {
		t.Error("SLURM script missing srun command for MPI")
	}

	if !strings.Contains(script, "--bind-to core") {
		t.Error("SLURM script missing MPI run arguments")
	}
}

// TestResourceOverrideOutOfRange tests that resource overrides are validated
func TestResourceOverrideOutOfRange(t *testing.T) {
	mapper := NewSlurmJobMapper("normal", "standard", "/modules")

	template := &hpctypes.WorkloadTemplate{
		TemplateID: "test-template",
		Name:       "Test Template",
		Version:    "1.0.0",
		Type:       hpctypes.WorkloadTypeBatch,
		Runtime: hpctypes.WorkloadRuntime{
			RuntimeType: "native",
		},
		Resources: hpctypes.WorkloadResourceSpec{
			MinNodes:               1,
			MaxNodes:               10,
			DefaultNodes:           2,
			MinCPUsPerNode:         1,
			MaxCPUsPerNode:         32,
			DefaultCPUsPerNode:     8,
			MinMemoryMBPerNode:     1024,
			MaxMemoryMBPerNode:     65536,
			DefaultMemoryMBPerNode: 8192,
			MinRuntimeMinutes:      1,
			MaxRuntimeMinutes:      1440,
			DefaultRuntimeMinutes:  60,
		},
		Security: hpctypes.WorkloadSecuritySpec{
			SandboxLevel:       "basic",
			AllowNetworkAccess: true,
		},
		Entrypoint: hpctypes.WorkloadEntrypoint{
			Command: "/usr/bin/test",
		},
		Publisher:      "cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a",
		ApprovalStatus: hpctypes.WorkloadApprovalApproved,
	}

	// Test nodes out of range
	nodesExceeded := int32(20) // Max is 10
	userParams := &UserParameters{
		Resources: &UserResourceOverrides{
			Nodes: &nodesExceeded,
		},
	}

	_, err := mapper.MapToSlurmJob(template, userParams)
	if err == nil {
		t.Error("expected error for nodes out of range, got nil")
	}

	// Test CPUs out of range
	cpusExceeded := int32(64) // Max is 32
	userParams = &UserParameters{
		Resources: &UserResourceOverrides{
			CPUsPerNode: &cpusExceeded,
		},
	}

	_, err = mapper.MapToSlurmJob(template, userParams)
	if err == nil {
		t.Error("expected error for CPUs out of range, got nil")
	}
}

// TestEnvironmentVariables tests environment variable resolution
func TestEnvironmentVariables(t *testing.T) {
	mapper := NewSlurmJobMapper("normal", "standard", "/modules")

	template := &hpctypes.WorkloadTemplate{
		TemplateID: "test-template",
		Name:       "Test Template",
		Version:    "1.0.0",
		Type:       hpctypes.WorkloadTypeBatch,
		Runtime: hpctypes.WorkloadRuntime{
			RuntimeType: "native",
		},
		Resources: hpctypes.WorkloadResourceSpec{
			MinNodes:               1,
			MaxNodes:               10,
			DefaultNodes:           2,
			MinCPUsPerNode:         1,
			MaxCPUsPerNode:         32,
			DefaultCPUsPerNode:     8,
			MinMemoryMBPerNode:     1024,
			MaxMemoryMBPerNode:     65536,
			DefaultMemoryMBPerNode: 8192,
			MinRuntimeMinutes:      1,
			MaxRuntimeMinutes:      1440,
			DefaultRuntimeMinutes:  60,
		},
		Security: hpctypes.WorkloadSecuritySpec{
			SandboxLevel:       "basic",
			AllowNetworkAccess: true,
		},
		Entrypoint: hpctypes.WorkloadEntrypoint{
			Command: "/usr/bin/test",
		},
		Environment: []hpctypes.EnvironmentVariable{
			{
				Name:     "APP_CONFIG",
				Value:    "production",
				Required: false,
			},
			{
				Name:     "USER_PARAM",
				Required: true,
			},
		},
		Publisher:      "cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a",
		ApprovalStatus: hpctypes.WorkloadApprovalApproved,
	}

	userParams := &UserParameters{
		Parameters: map[string]string{
			"USER_PARAM": "test-value",
		},
	}

	resolved, err := mapper.MapToSlurmJob(template, userParams)
	if err != nil {
		t.Fatalf("MapToSlurmJob failed: %v", err)
	}

	// Check environment variables are set
	if resolved.Environment["APP_CONFIG"] != "production" {
		t.Errorf("expected APP_CONFIG=production, got %s", resolved.Environment["APP_CONFIG"])
	}

	if resolved.Environment["USER_PARAM"] != "test-value" {
		t.Errorf("expected USER_PARAM=test-value, got %s", resolved.Environment["USER_PARAM"])
	}

	// Check script contains environment exports
	script := resolved.SlurmScript
	if !strings.Contains(script, "export APP_CONFIG='production'") {
		t.Error("SLURM script missing APP_CONFIG export")
	}

	if !strings.Contains(script, "export USER_PARAM='test-value'") {
		t.Error("SLURM script missing USER_PARAM export")
	}
}
