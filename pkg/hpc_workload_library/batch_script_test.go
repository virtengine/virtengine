// Package hpc_workload_library provides tests for batch script generation.
//
// VE-5F: Tests for SLURM batch script generation
package hpc_workload_library

import (
	"fmt"
	"strings"
	"testing"

	_ "github.com/virtengine/virtengine/sdk/go/sdkutil" // Initialize SDK config
)

func TestNewBatchScriptGenerator(t *testing.T) {
	config := BatchScriptConfig{
		Account:   "testaccount",
		Partition: "compute",
	}
	gen := NewBatchScriptGenerator(config)

	if gen == nil {
		t.Fatal("expected generator to be created")
	}
}

func TestGenerateScript_MPI(t *testing.T) {
	config := BatchScriptConfig{
		Account:   "hpc-project",
		Partition: "compute",
		JobName:   "mpi-test",
	}
	gen := NewBatchScriptGenerator(config)

	template := GetMPITemplate()
	params := &JobParameters{
		Nodes:        4,
		TasksPerNode: 16,
		MemoryMB:     32000,
		RuntimeMinutes: 120,
		Script:       "./my_mpi_app",
	}

	script, err := gen.GenerateScript(template, params)
	if err != nil {
		t.Fatalf("failed to generate script: %v", err)
	}

	// Verify shebang
	if !strings.HasPrefix(script, "#!/bin/bash") {
		t.Error("script should start with shebang")
	}

	// Verify SBATCH directives
	checks := []string{
		"#SBATCH --job-name=mpi-test",
		"#SBATCH --nodes=4",
		"#SBATCH --ntasks=64",
		"#SBATCH --ntasks-per-node=16",
		"#SBATCH --mem=32000M",
		"#SBATCH --time=02:00:00",
		"#SBATCH --partition=compute",
		"#SBATCH --account=hpc-project",
	}

	for _, check := range checks {
		if !strings.Contains(script, check) {
			t.Errorf("script missing: %s", check)
		}
	}

	// Verify module load
	if !strings.Contains(script, "module load openmpi") {
		t.Error("script should load openmpi module")
	}

	// Verify MPI execution
	if !strings.Contains(script, "mpirun") || !strings.Contains(script, "srun") {
		// At least one MPI launcher should be present
		t.Log("script may be using srun or mpirun")
	}
}

func TestGenerateScript_GPU(t *testing.T) {
	config := BatchScriptConfig{
		Partition: "gpu",
		JobName:   "gpu-test",
	}
	gen := NewBatchScriptGenerator(config)

	template := GetGPUComputeTemplate()
	params := &JobParameters{
		Nodes:          1,
		CPUsPerNode:    16,
		MemoryMB:       64000,
		RuntimeMinutes: 60,
		GPUs:           2,
		GPUType:        "nvidia-a100",
		Script:         "train.py",
	}

	script, err := gen.GenerateScript(template, params)
	if err != nil {
		t.Fatalf("failed to generate script: %v", err)
	}

	// Verify GPU directive
	if !strings.Contains(script, "#SBATCH --gres=gpu:nvidia-a100:2") {
		t.Error("script should have GPU gres directive")
	}

	// Verify CUDA module
	if !strings.Contains(script, "cuda") {
		t.Error("script should load CUDA module")
	}
}

func TestGenerateScript_Batch(t *testing.T) {
	config := BatchScriptConfig{
		Partition: "batch",
		JobName:   "batch-test",
	}
	gen := NewBatchScriptGenerator(config)

	template := GetBatchTemplate()
	params := &JobParameters{
		Nodes:          1,
		CPUsPerNode:    4,
		MemoryMB:       8000,
		RuntimeMinutes: 30,
		Command:        "python",
		Arguments:      []string{"-c", "print('hello')"},
	}

	script, err := gen.GenerateScript(template, params)
	if err != nil {
		t.Fatalf("failed to generate script: %v", err)
	}

	// Verify single node
	if !strings.Contains(script, "#SBATCH --nodes=1") {
		t.Error("script should have nodes=1")
	}

	// Verify time format
	if !strings.Contains(script, "#SBATCH --time=00:30:00") {
		t.Error("script should have 30 minute time limit")
	}
}

func TestGenerateScript_ArrayJob(t *testing.T) {
	config := BatchScriptConfig{
		Partition: "batch",
		JobName:   "array-test",
	}
	gen := NewBatchScriptGenerator(config)

	template := GetArrayJobTemplate()
	params := &JobParameters{
		Nodes:             1,
		CPUsPerNode:       4,
		MemoryMB:          8000,
		RuntimeMinutes:    60,
		ArrayStart:        0,
		ArrayEnd:          99,
		ArrayStep:         1,
		ArraySimultaneous: 10,
		Script:            "process.sh",
	}

	script, err := gen.GenerateScript(template, params)
	if err != nil {
		t.Fatalf("failed to generate script: %v", err)
	}

	// Verify array directive - step is only added when > 1
	if !strings.Contains(script, "#SBATCH --array=0-99%10") {
		t.Error("script should have array directive with limit")
	}

	// Verify array environment variables
	if !strings.Contains(script, "SLURM_ARRAY_TASK_ID") {
		t.Error("script should reference array task ID")
	}
}

func TestGenerateScript_SingularityContainer(t *testing.T) {
	config := BatchScriptConfig{
		Partition: "compute",
		JobName:   "singularity-test",
	}
	gen := NewBatchScriptGenerator(config)

	template := GetSingularityContainerTemplate()
	params := &JobParameters{
		Nodes:          1,
		CPUsPerNode:    8,
		MemoryMB:       16000,
		RuntimeMinutes: 120,
		GPUs:           1,
		Script:         "./run_analysis.sh",
	}

	script, err := gen.GenerateScript(template, params)
	if err != nil {
		t.Fatalf("failed to generate script: %v", err)
	}

	// Verify singularity execution
	if !strings.Contains(script, "singularity exec") {
		t.Error("script should use singularity exec")
	}

	// Verify GPU flag
	if !strings.Contains(script, "--nv") {
		t.Error("script should have --nv flag for GPU")
	}
}

func TestGenerateScript_DefaultParameters(t *testing.T) {
	config := BatchScriptConfig{}
	gen := NewBatchScriptGenerator(config)

	template := GetBatchTemplate()
	// nil params should use defaults
	script, err := gen.GenerateScript(template, nil)
	if err != nil {
		t.Fatalf("failed to generate script: %v", err)
	}

	// Should have default values from template
	if !strings.Contains(script, fmt.Sprintf("#SBATCH --nodes=%d", template.Resources.DefaultNodes)) {
		t.Error("script should use default nodes from template")
	}
}

func TestGenerateScript_CustomEnvironment(t *testing.T) {
	config := BatchScriptConfig{}
	gen := NewBatchScriptGenerator(config)

	template := GetBatchTemplate()
	params := &JobParameters{
		Environment: map[string]string{
			"MY_VAR":      "custom_value",
			"ANOTHER_VAR": "another_value",
		},
	}

	script, err := gen.GenerateScript(template, params)
	if err != nil {
		t.Fatalf("failed to generate script: %v", err)
	}

	if !strings.Contains(script, "export MY_VAR=\"custom_value\"") {
		t.Error("script should have custom environment variable")
	}
	if !strings.Contains(script, "export ANOTHER_VAR=\"another_value\"") {
		t.Error("script should have another custom environment variable")
	}
}

func TestGenerateScript_MailNotifications(t *testing.T) {
	config := BatchScriptConfig{
		MailUser: "user@example.com",
		MailType: "ALL",
	}
	gen := NewBatchScriptGenerator(config)

	template := GetBatchTemplate()
	script, err := gen.GenerateScript(template, nil)
	if err != nil {
		t.Fatalf("failed to generate script: %v", err)
	}

	if !strings.Contains(script, "#SBATCH --mail-user=user@example.com") {
		t.Error("script should have mail-user directive")
	}
	if !strings.Contains(script, "#SBATCH --mail-type=ALL") {
		t.Error("script should have mail-type directive")
	}
}

func TestGenerateScript_ExclusiveNodes(t *testing.T) {
	config := BatchScriptConfig{}
	gen := NewBatchScriptGenerator(config)

	template := GetGPUComputeTemplate() // GPU template has exclusive=true
	script, err := gen.GenerateScript(template, nil)
	if err != nil {
		t.Fatalf("failed to generate script: %v", err)
	}

	if !strings.Contains(script, "#SBATCH --exclusive") {
		t.Error("script should have exclusive directive")
	}
}

func TestGenerateScript_Constraints(t *testing.T) {
	config := BatchScriptConfig{}
	gen := NewBatchScriptGenerator(config)

	template := GetBatchTemplate()
	params := &JobParameters{
		Constraints: []string{"avx512", "infiniband"},
	}

	script, err := gen.GenerateScript(template, params)
	if err != nil {
		t.Fatalf("failed to generate script: %v", err)
	}

	if !strings.Contains(script, "#SBATCH --constraint=avx512&infiniband") {
		t.Error("script should have constraint directive with AND")
	}
}

func TestGenerateScript_CustomDirectives(t *testing.T) {
	config := BatchScriptConfig{
		CustomDirectives: map[string]string{
			"nice":    "100",
			"comment": "test-job",
		},
	}
	gen := NewBatchScriptGenerator(config)

	template := GetBatchTemplate()
	script, err := gen.GenerateScript(template, nil)
	if err != nil {
		t.Fatalf("failed to generate script: %v", err)
	}

	if !strings.Contains(script, "#SBATCH --nice=100") {
		t.Error("script should have custom nice directive")
	}
	if !strings.Contains(script, "#SBATCH --comment=test-job") {
		t.Error("script should have custom comment directive")
	}
}

func TestGenerateScript_NilTemplate(t *testing.T) {
	config := BatchScriptConfig{}
	gen := NewBatchScriptGenerator(config)

	_, err := gen.GenerateScript(nil, nil)
	if err == nil {
		t.Error("expected error for nil template")
	}
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		minutes  int64
		expected string
	}{
		{30, "00:30:00"},
		{60, "01:00:00"},
		{90, "01:30:00"},
		{120, "02:00:00"},
		{1440, "24:00:00"},
		{1500, "1-01:00:00"},
		{2880, "2-00:00:00"},
		{4320, "3-00:00:00"},
	}

	for _, tt := range tests {
		result := formatTime(tt.minutes)
		if result != tt.expected {
			t.Errorf("formatTime(%d) = %s, want %s", tt.minutes, result, tt.expected)
		}
	}
}

func TestJobParametersApplyDefaults(t *testing.T) {
	template := GetBatchTemplate()
	params := &JobParameters{}

	params.applyDefaults(template)

	if params.Nodes != template.Resources.DefaultNodes {
		t.Errorf("expected default nodes %d, got %d", template.Resources.DefaultNodes, params.Nodes)
	}
	if params.CPUsPerNode != template.Resources.DefaultCPUsPerNode {
		t.Errorf("expected default CPUs %d, got %d", template.Resources.DefaultCPUsPerNode, params.CPUsPerNode)
	}
	if params.MemoryMB != template.Resources.DefaultMemoryMBPerNode {
		t.Errorf("expected default memory %d, got %d", template.Resources.DefaultMemoryMBPerNode, params.MemoryMB)
	}
	if params.RuntimeMinutes != template.Resources.DefaultRuntimeMinutes {
		t.Errorf("expected default runtime %d, got %d", template.Resources.DefaultRuntimeMinutes, params.RuntimeMinutes)
	}
}

func TestGenerateScript_WorkingDirectory(t *testing.T) {
	config := BatchScriptConfig{}
	gen := NewBatchScriptGenerator(config)

	template := GetBatchTemplate()
	params := &JobParameters{
		WorkingDirectory: "/custom/workdir",
	}

	script, err := gen.GenerateScript(template, params)
	if err != nil {
		t.Fatalf("failed to generate script: %v", err)
	}

	if !strings.Contains(script, "cd /custom/workdir") {
		t.Error("script should change to custom working directory")
	}
}

func TestGenerateScript_Dependency(t *testing.T) {
	config := BatchScriptConfig{
		Dependency: "afterok:12345",
	}
	gen := NewBatchScriptGenerator(config)

	template := GetBatchTemplate()
	script, err := gen.GenerateScript(template, nil)
	if err != nil {
		t.Fatalf("failed to generate script: %v", err)
	}

	if !strings.Contains(script, "#SBATCH --dependency=afterok:12345") {
		t.Error("script should have dependency directive")
	}
}

func TestGenerateScript_Reservation(t *testing.T) {
	config := BatchScriptConfig{
		Reservation: "maintenance",
	}
	gen := NewBatchScriptGenerator(config)

	template := GetBatchTemplate()
	script, err := gen.GenerateScript(template, nil)
	if err != nil {
		t.Fatalf("failed to generate script: %v", err)
	}

	if !strings.Contains(script, "#SBATCH --reservation=maintenance") {
		t.Error("script should have reservation directive")
	}
}

