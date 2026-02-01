package slurm_adapter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// VE-2020: BatchScriptBuilder tests

func TestNewBatchScriptBuilder(t *testing.T) {
	b := NewBatchScriptBuilder()
	require.NotNil(t, b)
	assert.NotNil(t, b.environment)
	assert.NotNil(t, b.modules)
	assert.NotNil(t, b.constraints)
}

func TestBatchScriptBuilder_BasicDirectives(t *testing.T) {
	script := NewBatchScriptBuilder().
		SetJobName("test-job").
		SetPartition("compute").
		SetNodes(4).
		SetCPUsPerTask(16).
		SetMemoryMB(65536).
		SetTimeLimitMinutes(120).
		AddCommand("echo", "Hello World").
		Build()

	assert.Contains(t, script, "#!/bin/bash")
	assert.Contains(t, script, "#SBATCH --job-name=test-job")
	assert.Contains(t, script, "#SBATCH --partition=compute")
	assert.Contains(t, script, "#SBATCH --nodes=4")
	assert.Contains(t, script, "#SBATCH --cpus-per-task=16")
	assert.Contains(t, script, "#SBATCH --mem=65536M")
	assert.Contains(t, script, "#SBATCH --time=02:00:00")
	assert.Contains(t, script, "echo")
}

func TestBatchScriptBuilder_GPUConfiguration(t *testing.T) {
	t.Run("GPU with type", func(t *testing.T) {
		script := NewBatchScriptBuilder().
			SetJobName("gpu-job").
			SetNodes(1).
			SetGPUs(4, "a100").
			AddCommand("nvidia-smi").
			Build()

		assert.Contains(t, script, "#SBATCH --gres=gpu:a100:4")
	})

	t.Run("GPU without type", func(t *testing.T) {
		script := NewBatchScriptBuilder().
			SetJobName("gpu-job").
			SetNodes(1).
			SetGPUs(2, "").
			AddCommand("nvidia-smi").
			Build()

		assert.Contains(t, script, "#SBATCH --gres=gpu:2")
	})

	t.Run("GPUs per node", func(t *testing.T) {
		script := NewBatchScriptBuilder().
			SetJobName("gpu-job").
			SetNodes(2).
			SetGPUsPerNode(4).
			AddCommand("python train.py").
			Build()

		assert.Contains(t, script, "#SBATCH --gpus-per-node=4")
	})
}

func TestBatchScriptBuilder_OutputConfiguration(t *testing.T) {
	script := NewBatchScriptBuilder().
		SetJobName("output-test").
		SetNodes(1).
		SetCPUsPerTask(1).
		SetWorkingDir("/home/user/project").
		SetOutput("/logs/%j.out").
		SetError("/logs/%j.err").
		AddCommand("echo test").
		Build()

	assert.Contains(t, script, "#SBATCH --chdir=/home/user/project")
	assert.Contains(t, script, "#SBATCH --output=/logs/%j.out")
	assert.Contains(t, script, "#SBATCH --error=/logs/%j.err")
}

func TestBatchScriptBuilder_Constraints(t *testing.T) {
	script := NewBatchScriptBuilder().
		SetJobName("constraint-test").
		SetNodes(1).
		AddConstraint("nvme").
		AddConstraint("ib").
		SetExclusive(true).
		AddCommand("benchmark").
		Build()

	assert.Contains(t, script, "#SBATCH --constraint=nvme")
	assert.Contains(t, script, "#SBATCH --constraint=ib")
	assert.Contains(t, script, "#SBATCH --exclusive")
}

func TestBatchScriptBuilder_NodeSelection(t *testing.T) {
	script := NewBatchScriptBuilder().
		SetJobName("node-select").
		SetNodelist("node[001-004]").
		SetExcludeNodes("node002").
		AddCommand("hostname").
		Build()

	assert.Contains(t, script, "#SBATCH --nodelist=node[001-004]")
	assert.Contains(t, script, "#SBATCH --exclude=node002")
}

func TestBatchScriptBuilder_Dependencies(t *testing.T) {
	script := NewBatchScriptBuilder().
		SetJobName("dependent-job").
		SetNodes(1).
		SetDependency("afterok:12345:12346").
		AddCommand("process.sh").
		Build()

	assert.Contains(t, script, "#SBATCH --dependency=afterok:12345:12346")
}

func TestBatchScriptBuilder_ArrayJobs(t *testing.T) {
	script := NewBatchScriptBuilder().
		SetJobName("array-job").
		SetNodes(1).
		SetArray("0-99").
		AddRawCommand("python process.py --index=$SLURM_ARRAY_TASK_ID").
		Build()

	assert.Contains(t, script, "#SBATCH --array=0-99")
	assert.Contains(t, script, "$SLURM_ARRAY_TASK_ID")
}

func TestBatchScriptBuilder_EmailNotifications(t *testing.T) {
	script := NewBatchScriptBuilder().
		SetJobName("mail-test").
		SetNodes(1).
		SetMail("BEGIN,END,FAIL", "user@example.com").
		AddCommand("long-running-job.sh").
		Build()

	assert.Contains(t, script, "#SBATCH --mail-type=BEGIN,END,FAIL")
	assert.Contains(t, script, "#SBATCH --mail-user=user@example.com")
}

func TestBatchScriptBuilder_QOS(t *testing.T) {
	script := NewBatchScriptBuilder().
		SetJobName("qos-test").
		SetNodes(1).
		SetQOS("high").
		SetReservation("my-reservation").
		AddCommand("echo high priority").
		Build()

	assert.Contains(t, script, "#SBATCH --qos=high")
	assert.Contains(t, script, "#SBATCH --reservation=my-reservation")
}

func TestBatchScriptBuilder_Modules(t *testing.T) {
	script := NewBatchScriptBuilder().
		SetJobName("module-test").
		SetNodes(1).
		AddModuleCommand("module purge").
		LoadModule("python/3.9").
		LoadModule("cuda/11.7").
		LoadModules("gcc/11.2", "openmpi/4.1").
		AddCommand("python --version").
		Build()

	assert.Contains(t, script, "module purge")
	assert.Contains(t, script, "module load python/3.9")
	assert.Contains(t, script, "module load cuda/11.7")
	assert.Contains(t, script, "module load gcc/11.2")
	assert.Contains(t, script, "module load openmpi/4.1")
}

func TestBatchScriptBuilder_Environment(t *testing.T) {
	script := NewBatchScriptBuilder().
		SetJobName("env-test").
		SetNodes(1).
		SetEnv("CUDA_VISIBLE_DEVICES", "0,1,2,3").
		SetEnv("OMP_NUM_THREADS", "16").
		SetEnvMap(map[string]string{
			"MY_VAR":      "value1",
			"ANOTHER_VAR": "value2",
		}).
		AddCommand("env").
		Build()

	assert.Contains(t, script, "export CUDA_VISIBLE_DEVICES=\"0,1,2,3\"")
	assert.Contains(t, script, "export OMP_NUM_THREADS=\"16\"")
	assert.Contains(t, script, "export MY_VAR=")
	assert.Contains(t, script, "export ANOTHER_VAR=")
}

func TestBatchScriptBuilder_ContainerExecution(t *testing.T) {
	script := NewBatchScriptBuilder().
		SetJobName("container-job").
		SetNodes(1).
		SetGPUs(4, "a100").
		SetContainerImage("docker://pytorch/pytorch:2.0-cuda11.7").
		AddContainerOption("--nv").
		AddContainerOption("--bind /data:/data").
		AddCommand("python train.py").
		Build()

	assert.Contains(t, script, "singularity exec --nv --bind /data:/data docker://pytorch/pytorch:2.0-cuda11.7 python")
}

func TestBatchScriptBuilder_SetupAndCleanup(t *testing.T) {
	script := NewBatchScriptBuilder().
		SetJobName("setup-cleanup").
		SetNodes(1).
		AddSetupCommand("mkdir -p /scratch/$SLURM_JOB_ID").
		AddSetupCommand("cp input.dat /scratch/$SLURM_JOB_ID/").
		AddCommand("process.sh /scratch/$SLURM_JOB_ID").
		AddCleanupCommand("cp /scratch/$SLURM_JOB_ID/output.dat ./").
		AddCleanupCommand("rm -rf /scratch/$SLURM_JOB_ID").
		Build()

	// Verify order: setup before main, cleanup after
	setupIdx := strings.Index(script, "mkdir -p /scratch")
	mainIdx := strings.Index(script, "process.sh")
	cleanupIdx := strings.Index(script, "rm -rf /scratch")

	assert.True(t, setupIdx < mainIdx, "setup should come before main")
	assert.True(t, mainIdx < cleanupIdx, "main should come before cleanup")
}

func TestBatchScriptBuilder_Comments(t *testing.T) {
	script := NewBatchScriptBuilder().
		SetJobName("comment-test").
		SetNodes(1).
		AddHeaderComment("Generated by VirtEngine").
		AddHeaderComment("Job ID: ve-12345").
		AddCommand("echo test").
		Build()

	assert.Contains(t, script, "# Generated by VirtEngine")
	assert.Contains(t, script, "# Job ID: ve-12345")
}

func TestBatchScriptBuilder_MemoryOptions(t *testing.T) {
	t.Run("Total memory", func(t *testing.T) {
		script := NewBatchScriptBuilder().
			SetJobName("mem-total").
			SetNodes(1).
			SetMemoryMB(32768).
			AddCommand("echo").
			Build()

		assert.Contains(t, script, "#SBATCH --mem=32768M")
	})

	t.Run("Memory per CPU", func(t *testing.T) {
		script := NewBatchScriptBuilder().
			SetJobName("mem-per-cpu").
			SetNodes(1).
			SetMemoryPerCPU(4096).
			AddCommand("echo").
			Build()

		assert.Contains(t, script, "#SBATCH --mem-per-cpu=4096M")
	})
}

func TestBatchScriptBuilder_NTasks(t *testing.T) {
	script := NewBatchScriptBuilder().
		SetJobName("mpi-job").
		SetNodes(4).
		SetNTasks(128).
		SetCPUsPerTask(2).
		AddRawCommand("mpirun -np $SLURM_NTASKS ./my_mpi_program").
		Build()

	assert.Contains(t, script, "#SBATCH --nodes=4")
	assert.Contains(t, script, "#SBATCH --ntasks=128")
	assert.Contains(t, script, "#SBATCH --cpus-per-task=2")
}

func TestBatchScriptBuilder_Account(t *testing.T) {
	script := NewBatchScriptBuilder().
		SetJobName("account-test").
		SetAccount("project-abc123").
		SetNodes(1).
		AddCommand("echo").
		Build()

	assert.Contains(t, script, "#SBATCH --account=project-abc123")
}

func TestFromJobSpec(t *testing.T) {
	spec := &SLURMJobSpec{
		JobName:          "from-spec-job",
		Partition:        "gpu",
		Nodes:            2,
		CPUsPerNode:      32,
		MemoryMB:         65536,
		GPUs:             8,
		GPUType:          "a100",
		TimeLimit:        1440,
		WorkingDirectory: "/home/user/project",
		OutputDirectory:  "/home/user/logs",
		Exclusive:        true,
		Constraints:      []string{"nvme"},
		Environment: map[string]string{
			"CUDA_VISIBLE_DEVICES": "all",
		},
		ContainerImage: "docker://nvcr.io/nvidia/pytorch:23.04-py3",
		Command:        "python",
		Arguments:      []string{"train.py", "--epochs=100"},
	}

	builder := FromJobSpec(spec)
	script := builder.Build()

	assert.Contains(t, script, "#SBATCH --job-name=from-spec-job")
	assert.Contains(t, script, "#SBATCH --partition=gpu")
	assert.Contains(t, script, "#SBATCH --nodes=2")
	assert.Contains(t, script, "#SBATCH --cpus-per-task=32")
	assert.Contains(t, script, "#SBATCH --mem=65536M")
	assert.Contains(t, script, "#SBATCH --gres=gpu:a100:8")
	assert.Contains(t, script, "#SBATCH --time=1-00:00:00") // 1440 minutes = 1 day
	assert.Contains(t, script, "#SBATCH --chdir=/home/user/project")
	assert.Contains(t, script, "#SBATCH --output=/home/user/logs/%j.out")
	assert.Contains(t, script, "#SBATCH --exclusive")
	assert.Contains(t, script, "#SBATCH --constraint=nvme")
	assert.Contains(t, script, "export CUDA_VISIBLE_DEVICES=")
	assert.Contains(t, script, "singularity exec")
	assert.Contains(t, script, "python")
}

func TestFormatTimeLimit(t *testing.T) {
	tests := []struct {
		minutes  int64
		expected string
	}{
		{30, "00:30:00"},
		{60, "01:00:00"},
		{90, "01:30:00"},
		{120, "02:00:00"},
		{1440, "1-00:00:00"},  // 24 hours
		{2880, "2-00:00:00"},  // 48 hours
		{4320, "3-00:00:00"},  // 72 hours
		{1500, "1-01:00:00"},  // 25 hours
		{10080, "7-00:00:00"}, // 1 week
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			result := formatTimeLimit(tc.minutes)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestBatchScriptBuilder_JobInformation(t *testing.T) {
	script := NewBatchScriptBuilder().
		SetJobName("info-test").
		SetNodes(1).
		AddCommand("hostname").
		Build()

	// Script should include job information echoes
	assert.Contains(t, script, `echo "Job ID: $SLURM_JOB_ID"`)
	assert.Contains(t, script, `echo "Job Name: $SLURM_JOB_NAME"`)
	assert.Contains(t, script, `echo "Node List: $SLURM_JOB_NODELIST"`)
}

func TestBatchScriptBuilder_ErrorHandling(t *testing.T) {
	script := NewBatchScriptBuilder().
		SetJobName("error-test").
		SetNodes(1).
		AddCommand("might-fail.sh").
		Build()

	// Script should capture exit code
	assert.Contains(t, script, "EXIT_CODE=$?")
	assert.Contains(t, script, "exit $EXIT_CODE")
}

func TestBatchScriptBuilder_SetTimeLimit(t *testing.T) {
	builder := NewBatchScriptBuilder().
		SetJobName("time-test").
		SetNodes(1)

	// Test hours:minutes:seconds format
	builder.SetTimeLimit(2, 30, 0)
	script := builder.AddCommand("echo").Build()

	// 2 hours 30 minutes = 150 minutes
	assert.Contains(t, script, "#SBATCH --time=02:30:00")
}

func TestDefaultBatchScriptOptions(t *testing.T) {
	opts := DefaultBatchScriptOptions()

	assert.False(t, opts.UseModulePurge)
	assert.False(t, opts.EnableProfiling)
	assert.False(t, opts.EnableTiming)
	assert.True(t, opts.StrictMode)
}

