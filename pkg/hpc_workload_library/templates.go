// Package hpc_workload_library provides HPC workload templates.
//
// VE-5F: Pre-configured workload templates
package hpc_workload_library

import (
	"time"

	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// BuiltinTemplatePublisher is the publisher address for builtin templates
// This is a well-known address for built-in templates
const BuiltinTemplatePublisher = "akash1365yvmc4s7awdyj3n2sav7xfx76adc6dnmlx63"

// GetBuiltinTemplates returns all built-in workload templates
func GetBuiltinTemplates() []*hpctypes.WorkloadTemplate {
	return []*hpctypes.WorkloadTemplate{
		GetMPITemplate(),
		GetGPUComputeTemplate(),
		GetBatchTemplate(),
		GetDataProcessingTemplate(),
		GetInteractiveTemplate(),
	}
}

// GetMPITemplate returns the MPI workload template
func GetMPITemplate() *hpctypes.WorkloadTemplate {
	return &hpctypes.WorkloadTemplate{
		TemplateID:  "mpi-standard",
		Name:        "MPI Standard Workload",
		Version:     "1.0.0",
		Description: "Standard MPI-based parallel computing workload using OpenMPI. Suitable for distributed memory parallel applications with multi-node communication.",
		Type:        hpctypes.WorkloadTypeMPI,
		Runtime: hpctypes.WorkloadRuntime{
			RuntimeType:       "singularity",
			ContainerImage:    "library/ubuntu-mpi:22.04",
			MPIImplementation: "openmpi",
			RequiredModules:   []string{"openmpi/4.1", "gcc/11"},
		},
		Resources: hpctypes.WorkloadResourceSpec{
			MinNodes:               1,
			MaxNodes:               128,
			DefaultNodes:           4,
			MinCPUsPerNode:         1,
			MaxCPUsPerNode:         128,
			DefaultCPUsPerNode:     16,
			MinMemoryMBPerNode:     1024,
			MaxMemoryMBPerNode:     512000,
			DefaultMemoryMBPerNode: 32000,
			MinRuntimeMinutes:      1,
			MaxRuntimeMinutes:      2880, // 48 hours
			DefaultRuntimeMinutes:  60,
			NetworkRequired:        true,
			ExclusiveNodes:         false,
		},
		Security: hpctypes.WorkloadSecuritySpec{
			AllowedRegistries:  []string{"library", "docker.io", "ghcr.io"},
			RequireImageDigest: false,
			AllowNetworkAccess: true,
			AllowHostMounts:    true,
			AllowedHostPaths:   []string{"/scratch", "/home", "/work"},
			SandboxLevel:       "basic",
			MaxOpenFiles:       65536,
			MaxProcesses:       32768,
		},
		Entrypoint: hpctypes.WorkloadEntrypoint{
			Command:          "mpirun",
			DefaultArgs:      []string{"-np", "${SLURM_NTASKS}"},
			WorkingDirectory: "/scratch/$USER/$SLURM_JOB_ID",
			UseMPIRun:        true,
			PreRunScript:     "mkdir -p /scratch/$USER/$SLURM_JOB_ID",
			PostRunScript:    "echo 'Job completed at $(date)'",
		},
		Environment: []hpctypes.EnvironmentVariable{
			{Name: "OMP_NUM_THREADS", Value: "1", Description: "OpenMP threads per MPI rank"},
			{Name: "MPI_BUFFER_SIZE", Value: "20971520", Description: "MPI buffer size in bytes"},
		},
		Modules: []string{"openmpi/4.1", "gcc/11"},
		DataBindings: []hpctypes.DataBinding{
			{Name: "input", MountPath: "/data/input", DataType: "input", Required: false, ReadOnly: true},
			{Name: "output", MountPath: "/data/output", DataType: "output", Required: false, ReadOnly: false},
			{Name: "scratch", MountPath: "/scratch", DataType: "scratch", Required: true, ReadOnly: false},
		},
		ParameterSchema: []hpctypes.ParameterDefinition{
			{Name: "executable", Type: "string", Description: "Path to MPI executable", Required: true},
			{Name: "args", Type: "string", Description: "Arguments to pass to executable", Required: false},
			{Name: "tasks_per_node", Type: "int", Description: "MPI tasks per node", Default: "16", MinValue: "1", MaxValue: "128"},
		},
		ApprovalStatus: hpctypes.WorkloadApprovalApproved,
		Publisher:      BuiltinTemplatePublisher,
		Tags:           []string{"mpi", "parallel", "distributed", "hpc"},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// GetGPUComputeTemplate returns the GPU compute workload template
func GetGPUComputeTemplate() *hpctypes.WorkloadTemplate {
	return &hpctypes.WorkloadTemplate{
		TemplateID:  "gpu-compute",
		Name:        "GPU Compute Workload",
		Version:     "1.0.0",
		Description: "GPU-accelerated compute workload using CUDA. Suitable for deep learning training, scientific simulations, and GPU-accelerated applications.",
		Type:        hpctypes.WorkloadTypeGPU,
		Runtime: hpctypes.WorkloadRuntime{
			RuntimeType:     "singularity",
			ContainerImage:  "nvcr.io/nvidia/cuda:12.2-runtime-ubuntu22.04",
			CUDAVersion:     "12.2",
			RequiredModules: []string{"cuda/12.2", "cudnn/8.9"},
		},
		Resources: hpctypes.WorkloadResourceSpec{
			MinNodes:               1,
			MaxNodes:               32,
			DefaultNodes:           1,
			MinCPUsPerNode:         4,
			MaxCPUsPerNode:         128,
			DefaultCPUsPerNode:     16,
			MinMemoryMBPerNode:     8192,
			MaxMemoryMBPerNode:     512000,
			DefaultMemoryMBPerNode: 64000,
			MinGPUsPerNode:         1,
			MaxGPUsPerNode:         8,
			DefaultGPUsPerNode:     1,
			GPUTypes:               []string{"nvidia-a100", "nvidia-v100", "nvidia-h100"},
			MinRuntimeMinutes:      1,
			MaxRuntimeMinutes:      4320, // 72 hours
			DefaultRuntimeMinutes:  120,
			NetworkRequired:        true,
			ExclusiveNodes:         true,
		},
		Security: hpctypes.WorkloadSecuritySpec{
			AllowedRegistries:  []string{"nvcr.io", "docker.io", "ghcr.io"},
			RequireImageDigest: true,
			AllowNetworkAccess: true,
			AllowHostMounts:    true,
			AllowedHostPaths:   []string{"/scratch", "/home", "/work", "/datasets"},
			SandboxLevel:       "basic",
			MaxOpenFiles:       262144,
			MaxProcesses:       65536,
		},
		Entrypoint: hpctypes.WorkloadEntrypoint{
			Command:          "/bin/bash",
			DefaultArgs:      []string{"-c"},
			WorkingDirectory: "/workspace",
			PreRunScript:     "nvidia-smi && echo 'CUDA devices: '$CUDA_VISIBLE_DEVICES",
			PostRunScript:    "nvidia-smi --query-gpu=utilization.gpu --format=csv",
		},
		Environment: []hpctypes.EnvironmentVariable{
			{Name: "CUDA_VISIBLE_DEVICES", ValueTemplate: "${SLURM_GPUS_ON_NODE}", Description: "GPU devices visible to CUDA"},
			{Name: "NCCL_DEBUG", Value: "INFO", Description: "NCCL debug level"},
			{Name: "CUDA_CACHE_PATH", Value: "/tmp/cuda_cache", Description: "CUDA compilation cache"},
		},
		Modules: []string{"cuda/12.2", "cudnn/8.9"},
		DataBindings: []hpctypes.DataBinding{
			{Name: "datasets", MountPath: "/datasets", DataType: "input", Required: false, ReadOnly: true},
			{Name: "models", MountPath: "/models", DataType: "output", Required: false, ReadOnly: false},
			{Name: "checkpoints", MountPath: "/checkpoints", DataType: "output", Required: false, ReadOnly: false},
		},
		ParameterSchema: []hpctypes.ParameterDefinition{
			{Name: "script", Type: "string", Description: "Python script to run", Required: true},
			{Name: "gpu_type", Type: "enum", Description: "GPU type to use", EnumValues: []string{"nvidia-a100", "nvidia-v100", "nvidia-h100"}, Default: "nvidia-a100"},
			{Name: "mixed_precision", Type: "bool", Description: "Enable mixed precision training", Default: "true"},
		},
		ApprovalStatus: hpctypes.WorkloadApprovalApproved,
		Publisher:      BuiltinTemplatePublisher,
		Tags:           []string{"gpu", "cuda", "deep-learning", "ai", "ml"},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// GetBatchTemplate returns the batch processing workload template
func GetBatchTemplate() *hpctypes.WorkloadTemplate {
	return &hpctypes.WorkloadTemplate{
		TemplateID:  "batch-standard",
		Name:        "Batch Processing Workload",
		Version:     "1.0.0",
		Description: "Standard batch processing workload for single-node or serial computation tasks. Suitable for parameter sweeps, independent simulations, and data processing.",
		Type:        hpctypes.WorkloadTypeBatch,
		Runtime: hpctypes.WorkloadRuntime{
			RuntimeType:     "singularity",
			ContainerImage:  "library/ubuntu:22.04",
			RequiredModules: []string{},
		},
		Resources: hpctypes.WorkloadResourceSpec{
			MinNodes:               1,
			MaxNodes:               1,
			DefaultNodes:           1,
			MinCPUsPerNode:         1,
			MaxCPUsPerNode:         128,
			DefaultCPUsPerNode:     4,
			MinMemoryMBPerNode:     512,
			MaxMemoryMBPerNode:     256000,
			DefaultMemoryMBPerNode: 8000,
			MinRuntimeMinutes:      1,
			MaxRuntimeMinutes:      1440, // 24 hours
			DefaultRuntimeMinutes:  30,
			ExclusiveNodes:         false,
		},
		Security: hpctypes.WorkloadSecuritySpec{
			AllowedRegistries:  []string{"library", "docker.io", "ghcr.io", "quay.io"},
			RequireImageDigest: false,
			AllowNetworkAccess: false,
			AllowHostMounts:    true,
			AllowedHostPaths:   []string{"/scratch", "/home", "/work"},
			SandboxLevel:       "strict",
			MaxOpenFiles:       16384,
			MaxProcesses:       4096,
			MaxFileSize:        10737418240, // 10GB
		},
		Entrypoint: hpctypes.WorkloadEntrypoint{
			Command:          "/bin/bash",
			DefaultArgs:      []string{"-c"},
			WorkingDirectory: "/work",
		},
		Environment: []hpctypes.EnvironmentVariable{
			{Name: "TMPDIR", Value: "/scratch/$USER/$SLURM_JOB_ID/tmp", Description: "Temporary directory"},
		},
		DataBindings: []hpctypes.DataBinding{
			{Name: "input", MountPath: "/input", DataType: "input", Required: false, ReadOnly: true},
			{Name: "output", MountPath: "/output", DataType: "output", Required: false, ReadOnly: false},
		},
		ParameterSchema: []hpctypes.ParameterDefinition{
			{Name: "command", Type: "string", Description: "Command to execute", Required: true},
			{Name: "array_start", Type: "int", Description: "Array job start index", Default: "0", MinValue: "0"},
			{Name: "array_end", Type: "int", Description: "Array job end index", Default: "0", MinValue: "0"},
		},
		ApprovalStatus: hpctypes.WorkloadApprovalApproved,
		Publisher:      BuiltinTemplatePublisher,
		Tags:           []string{"batch", "serial", "processing"},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// GetDataProcessingTemplate returns the data processing workload template
func GetDataProcessingTemplate() *hpctypes.WorkloadTemplate {
	return &hpctypes.WorkloadTemplate{
		TemplateID:  "data-processing",
		Name:        "Data Processing Pipeline",
		Version:     "1.0.0",
		Description: "Data processing and ETL workload for large-scale data transformation. Includes support for Apache Spark, Dask, and other distributed data frameworks.",
		Type:        hpctypes.WorkloadTypeDataProcessing,
		Runtime: hpctypes.WorkloadRuntime{
			RuntimeType:     "singularity",
			ContainerImage:  "apache/spark-py:3.5.0",
			PythonVersion:   "3.11",
			RequiredModules: []string{"python/3.11"},
		},
		Resources: hpctypes.WorkloadResourceSpec{
			MinNodes:               1,
			MaxNodes:               64,
			DefaultNodes:           4,
			MinCPUsPerNode:         4,
			MaxCPUsPerNode:         128,
			DefaultCPUsPerNode:     32,
			MinMemoryMBPerNode:     8192,
			MaxMemoryMBPerNode:     512000,
			DefaultMemoryMBPerNode: 128000,
			MinRuntimeMinutes:      5,
			MaxRuntimeMinutes:      2880, // 48 hours
			DefaultRuntimeMinutes:  120,
			StorageGBRequired:      100,
			NetworkRequired:        true,
			ExclusiveNodes:         false,
		},
		Security: hpctypes.WorkloadSecuritySpec{
			AllowedRegistries:  []string{"apache", "docker.io", "ghcr.io"},
			RequireImageDigest: false,
			AllowNetworkAccess: true,
			AllowHostMounts:    true,
			AllowedHostPaths:   []string{"/scratch", "/home", "/work", "/data"},
			SandboxLevel:       "basic",
			MaxOpenFiles:       131072,
			MaxProcesses:       32768,
		},
		Entrypoint: hpctypes.WorkloadEntrypoint{
			Command:          "spark-submit",
			WorkingDirectory: "/work",
			PreRunScript:     "export SPARK_LOCAL_DIRS=/scratch/$USER/$SLURM_JOB_ID",
		},
		Environment: []hpctypes.EnvironmentVariable{
			{Name: "SPARK_HOME", Value: "/opt/spark", Description: "Spark installation directory"},
			{Name: "PYSPARK_PYTHON", Value: "/usr/bin/python3", Description: "Python interpreter for PySpark"},
			{Name: "SPARK_WORKER_MEMORY", ValueTemplate: "${SLURM_MEM_PER_NODE}m", Description: "Spark worker memory"},
		},
		Modules: []string{"python/3.11"},
		DataBindings: []hpctypes.DataBinding{
			{Name: "input_data", MountPath: "/data/input", DataType: "input", Required: true, ReadOnly: true},
			{Name: "output_data", MountPath: "/data/output", DataType: "output", Required: true, ReadOnly: false},
			{Name: "spark_scratch", MountPath: "/scratch", DataType: "scratch", Required: true, ReadOnly: false},
		},
		ParameterSchema: []hpctypes.ParameterDefinition{
			{Name: "main_script", Type: "string", Description: "Main Python/Scala script", Required: true},
			{Name: "executor_memory", Type: "string", Description: "Executor memory (e.g., 8g)", Default: "8g"},
			{Name: "executor_cores", Type: "int", Description: "Cores per executor", Default: "4", MinValue: "1", MaxValue: "32"},
			{Name: "num_executors", Type: "int", Description: "Number of executors", Default: "4", MinValue: "1", MaxValue: "256"},
		},
		ApprovalStatus: hpctypes.WorkloadApprovalApproved,
		Publisher:      BuiltinTemplatePublisher,
		Tags:           []string{"data", "spark", "etl", "processing", "analytics"},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// GetInteractiveTemplate returns the interactive session workload template
func GetInteractiveTemplate() *hpctypes.WorkloadTemplate {
	return &hpctypes.WorkloadTemplate{
		TemplateID:  "interactive-session",
		Name:        "Interactive Session",
		Version:     "1.0.0",
		Description: "Interactive computing session with JupyterLab or terminal access. Suitable for development, debugging, and exploratory analysis.",
		Type:        hpctypes.WorkloadTypeInteractive,
		Runtime: hpctypes.WorkloadRuntime{
			RuntimeType:     "singularity",
			ContainerImage:  "jupyter/scipy-notebook:latest",
			PythonVersion:   "3.11",
			RequiredModules: []string{"python/3.11"},
		},
		Resources: hpctypes.WorkloadResourceSpec{
			MinNodes:               1,
			MaxNodes:               1,
			DefaultNodes:           1,
			MinCPUsPerNode:         2,
			MaxCPUsPerNode:         32,
			DefaultCPUsPerNode:     8,
			MinMemoryMBPerNode:     4096,
			MaxMemoryMBPerNode:     128000,
			DefaultMemoryMBPerNode: 16000,
			MinGPUsPerNode:         0,
			MaxGPUsPerNode:         2,
			DefaultGPUsPerNode:     0,
			GPUTypes:               []string{"nvidia-a100", "nvidia-v100"},
			MinRuntimeMinutes:      15,
			MaxRuntimeMinutes:      480, // 8 hours
			DefaultRuntimeMinutes:  120,
			ExclusiveNodes:         false,
		},
		Security: hpctypes.WorkloadSecuritySpec{
			AllowedRegistries:  []string{"jupyter", "docker.io", "ghcr.io"},
			RequireImageDigest: false,
			AllowNetworkAccess: true,
			AllowHostMounts:    true,
			AllowedHostPaths:   []string{"/home", "/work", "/scratch"},
			SandboxLevel:       "basic",
			MaxOpenFiles:       32768,
			MaxProcesses:       8192,
		},
		Entrypoint: hpctypes.WorkloadEntrypoint{
			Command:          "jupyter",
			DefaultArgs:      []string{"lab", "--no-browser", "--ip=0.0.0.0"},
			WorkingDirectory: "/home/jovyan/work",
			PreRunScript:     "echo 'Starting JupyterLab session'",
		},
		Environment: []hpctypes.EnvironmentVariable{
			{Name: "JUPYTER_ENABLE_LAB", Value: "yes", Description: "Enable JupyterLab interface"},
			{Name: "JUPYTER_TOKEN", ValueTemplate: "${SLURM_JOB_ID}", Description: "Jupyter authentication token", Secret: true},
		},
		DataBindings: []hpctypes.DataBinding{
			{Name: "work", MountPath: "/home/jovyan/work", HostPath: "/work/$USER", DataType: "output", Required: true, ReadOnly: false},
			{Name: "data", MountPath: "/home/jovyan/data", DataType: "input", Required: false, ReadOnly: true},
		},
		ParameterSchema: []hpctypes.ParameterDefinition{
			{Name: "port", Type: "int", Description: "JupyterLab port", Default: "8888", MinValue: "8000", MaxValue: "9999"},
			{Name: "gpu_enabled", Type: "bool", Description: "Enable GPU support", Default: "false"},
			{Name: "interface", Type: "enum", Description: "Interface type", EnumValues: []string{"jupyterlab", "notebook", "terminal"}, Default: "jupyterlab"},
		},
		ApprovalStatus: hpctypes.WorkloadApprovalApproved,
		Publisher:      BuiltinTemplatePublisher,
		Tags:           []string{"interactive", "jupyter", "development", "notebook"},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// GetTemplateByID returns a built-in template by ID
func GetTemplateByID(templateID string) *hpctypes.WorkloadTemplate {
	for _, t := range GetBuiltinTemplates() {
		if t.TemplateID == templateID {
			return t
		}
	}
	return nil
}

// GetTemplatesByType returns all templates of a given type
func GetTemplatesByType(workloadType hpctypes.WorkloadType) []*hpctypes.WorkloadTemplate {
	var result []*hpctypes.WorkloadTemplate
	for _, t := range GetBuiltinTemplates() {
		if t.Type == workloadType {
			result = append(result, t)
		}
	}
	return result
}
