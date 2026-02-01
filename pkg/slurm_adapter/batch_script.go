// Package slurm_adapter implements the SLURM orchestration adapter for VirtEngine.
//
// VE-2020: BatchScriptBuilder for generating SLURM batch scripts
package slurm_adapter

import (
	"fmt"
	"sort"
	"strings"
)

// BatchScriptBuilder constructs SLURM batch scripts with proper #SBATCH directives
type BatchScriptBuilder struct {
	jobName          string
	partition        string
	account          string
	nodes            int32
	ntasks           int32
	cpusPerTask      int32
	memoryMB         int64
	memoryPerCPU     int64
	timeLimitMinutes int64
	gpus             int32
	gpuType          string
	gpusPerNode      int32
	gpusPerTask      int32
	workingDir       string
	outputFile       string
	errorFile        string
	exclusive        bool
	constraints      []string
	nodelist         string
	excludeNodes     string
	dependency       string
	arraySpec        string
	mailType         string
	mailUser         string
	qos              string
	reservation      string

	// Environment and modules
	environment    map[string]string
	modules        []string
	moduleCommands []string

	// Container support
	containerImage   string
	containerOptions []string

	// Script content
	setupCommands   []string
	mainCommands    []string
	cleanupCommands []string

	// Comments
	headerComments []string
}

// NewBatchScriptBuilder creates a new BatchScriptBuilder with default settings
func NewBatchScriptBuilder() *BatchScriptBuilder {
	return &BatchScriptBuilder{
		environment:      make(map[string]string),
		modules:          make([]string, 0),
		moduleCommands:   make([]string, 0),
		constraints:      make([]string, 0),
		setupCommands:    make([]string, 0),
		mainCommands:     make([]string, 0),
		cleanupCommands:  make([]string, 0),
		headerComments:   make([]string, 0),
		containerOptions: make([]string, 0),
	}
}

// FromJobSpec creates a BatchScriptBuilder from a SLURMJobSpec
func FromJobSpec(spec *SLURMJobSpec) *BatchScriptBuilder {
	b := NewBatchScriptBuilder()

	b.SetJobName(spec.JobName)
	b.SetPartition(spec.Partition)
	b.SetNodes(spec.Nodes)
	b.SetCPUsPerTask(spec.CPUsPerNode)
	b.SetMemoryMB(spec.MemoryMB)
	b.SetTimeLimitMinutes(spec.TimeLimit)

	if spec.GPUs > 0 {
		b.SetGPUs(spec.GPUs, spec.GPUType)
	}

	if spec.WorkingDirectory != "" {
		b.SetWorkingDir(spec.WorkingDirectory)
	}

	if spec.OutputDirectory != "" {
		b.SetOutput(fmt.Sprintf("%s/%%j.out", spec.OutputDirectory))
		b.SetError(fmt.Sprintf("%s/%%j.err", spec.OutputDirectory))
	}

	if spec.Exclusive {
		b.SetExclusive(true)
	}

	for _, c := range spec.Constraints {
		b.AddConstraint(c)
	}

	for k, v := range spec.Environment {
		b.SetEnv(k, v)
	}

	if spec.ContainerImage != "" {
		b.SetContainerImage(spec.ContainerImage)
	}

	if spec.Command != "" {
		b.AddCommand(spec.Command, spec.Arguments...)
	}

	return b
}

// SetJobName sets the job name (--job-name)
func (b *BatchScriptBuilder) SetJobName(name string) *BatchScriptBuilder {
	b.jobName = name
	return b
}

// SetPartition sets the partition (--partition)
func (b *BatchScriptBuilder) SetPartition(partition string) *BatchScriptBuilder {
	b.partition = partition
	return b
}

// SetAccount sets the account for billing (--account)
func (b *BatchScriptBuilder) SetAccount(account string) *BatchScriptBuilder {
	b.account = account
	return b
}

// SetNodes sets the number of nodes (--nodes)
func (b *BatchScriptBuilder) SetNodes(nodes int32) *BatchScriptBuilder {
	b.nodes = nodes
	return b
}

// SetNTasks sets the total number of tasks (--ntasks)
func (b *BatchScriptBuilder) SetNTasks(ntasks int32) *BatchScriptBuilder {
	b.ntasks = ntasks
	return b
}

// SetCPUsPerTask sets CPUs per task (--cpus-per-task)
func (b *BatchScriptBuilder) SetCPUsPerTask(cpus int32) *BatchScriptBuilder {
	b.cpusPerTask = cpus
	return b
}

// SetMemoryMB sets total memory in MB (--mem)
func (b *BatchScriptBuilder) SetMemoryMB(memMB int64) *BatchScriptBuilder {
	b.memoryMB = memMB
	return b
}

// SetMemoryPerCPU sets memory per CPU in MB (--mem-per-cpu)
func (b *BatchScriptBuilder) SetMemoryPerCPU(memMB int64) *BatchScriptBuilder {
	b.memoryPerCPU = memMB
	return b
}

// SetTimeLimitMinutes sets the time limit in minutes (--time)
func (b *BatchScriptBuilder) SetTimeLimitMinutes(minutes int64) *BatchScriptBuilder {
	b.timeLimitMinutes = minutes
	return b
}

// SetTimeLimit sets the time limit in HH:MM:SS format
func (b *BatchScriptBuilder) SetTimeLimit(hours, minutes, seconds int) *BatchScriptBuilder {
	b.timeLimitMinutes = int64(hours*60 + minutes + seconds/60)
	return b
}

// SetGPUs sets GPU resources (--gres=gpu:type:count)
func (b *BatchScriptBuilder) SetGPUs(count int32, gpuType string) *BatchScriptBuilder {
	b.gpus = count
	b.gpuType = gpuType
	return b
}

// SetGPUsPerNode sets GPUs per node (--gpus-per-node)
func (b *BatchScriptBuilder) SetGPUsPerNode(count int32) *BatchScriptBuilder {
	b.gpusPerNode = count
	return b
}

// SetGPUsPerTask sets GPUs per task (--gpus-per-task)
func (b *BatchScriptBuilder) SetGPUsPerTask(count int32) *BatchScriptBuilder {
	b.gpusPerTask = count
	return b
}

// SetWorkingDir sets the working directory (--chdir)
func (b *BatchScriptBuilder) SetWorkingDir(dir string) *BatchScriptBuilder {
	b.workingDir = dir
	return b
}

// SetOutput sets the output file (--output)
func (b *BatchScriptBuilder) SetOutput(path string) *BatchScriptBuilder {
	b.outputFile = path
	return b
}

// SetError sets the error file (--error)
func (b *BatchScriptBuilder) SetError(path string) *BatchScriptBuilder {
	b.errorFile = path
	return b
}

// SetExclusive sets exclusive node access (--exclusive)
func (b *BatchScriptBuilder) SetExclusive(exclusive bool) *BatchScriptBuilder {
	b.exclusive = exclusive
	return b
}

// AddConstraint adds a node constraint (--constraint)
func (b *BatchScriptBuilder) AddConstraint(constraint string) *BatchScriptBuilder {
	b.constraints = append(b.constraints, constraint)
	return b
}

// SetNodelist sets specific nodes to use (--nodelist)
func (b *BatchScriptBuilder) SetNodelist(nodes string) *BatchScriptBuilder {
	b.nodelist = nodes
	return b
}

// SetExcludeNodes sets nodes to exclude (--exclude)
func (b *BatchScriptBuilder) SetExcludeNodes(nodes string) *BatchScriptBuilder {
	b.excludeNodes = nodes
	return b
}

// SetDependency sets job dependencies (--dependency)
// Format: afterok:jobid, afterany:jobid, etc.
func (b *BatchScriptBuilder) SetDependency(dep string) *BatchScriptBuilder {
	b.dependency = dep
	return b
}

// SetArray sets job array specification (--array)
// Format: 0-15, 0-15:4, 0,2,4,6
func (b *BatchScriptBuilder) SetArray(spec string) *BatchScriptBuilder {
	b.arraySpec = spec
	return b
}

// SetMail sets email notifications (--mail-type, --mail-user)
func (b *BatchScriptBuilder) SetMail(mailType, mailUser string) *BatchScriptBuilder {
	b.mailType = mailType
	b.mailUser = mailUser
	return b
}

// SetQOS sets quality of service (--qos)
func (b *BatchScriptBuilder) SetQOS(qos string) *BatchScriptBuilder {
	b.qos = qos
	return b
}

// SetReservation sets the reservation (--reservation)
func (b *BatchScriptBuilder) SetReservation(reservation string) *BatchScriptBuilder {
	b.reservation = reservation
	return b
}

// SetEnv sets an environment variable
func (b *BatchScriptBuilder) SetEnv(key, value string) *BatchScriptBuilder {
	b.environment[key] = value
	return b
}

// SetEnvMap sets multiple environment variables
func (b *BatchScriptBuilder) SetEnvMap(env map[string]string) *BatchScriptBuilder {
	for k, v := range env {
		b.environment[k] = v
	}
	return b
}

// LoadModule adds a module to load
func (b *BatchScriptBuilder) LoadModule(module string) *BatchScriptBuilder {
	b.modules = append(b.modules, module)
	return b
}

// LoadModules adds multiple modules to load
func (b *BatchScriptBuilder) LoadModules(modules ...string) *BatchScriptBuilder {
	b.modules = append(b.modules, modules...)
	return b
}

// AddModuleCommand adds a custom module command (e.g., "module swap gcc/9.3 gcc/11.2")
func (b *BatchScriptBuilder) AddModuleCommand(cmd string) *BatchScriptBuilder {
	b.moduleCommands = append(b.moduleCommands, cmd)
	return b
}

// SetContainerImage sets the container image for Singularity/Apptainer
func (b *BatchScriptBuilder) SetContainerImage(image string) *BatchScriptBuilder {
	b.containerImage = image
	return b
}

// AddContainerOption adds a Singularity/Apptainer option
func (b *BatchScriptBuilder) AddContainerOption(opt string) *BatchScriptBuilder {
	b.containerOptions = append(b.containerOptions, opt)
	return b
}

// AddSetupCommand adds a setup command to run before the main job
func (b *BatchScriptBuilder) AddSetupCommand(cmd string) *BatchScriptBuilder {
	b.setupCommands = append(b.setupCommands, cmd)
	return b
}

// AddCommand adds a main command with arguments
func (b *BatchScriptBuilder) AddCommand(cmd string, args ...string) *BatchScriptBuilder {
	fullCmd := cmd
	for _, arg := range args {
		fullCmd += fmt.Sprintf(" %q", arg)
	}
	b.mainCommands = append(b.mainCommands, fullCmd)
	return b
}

// AddRawCommand adds a raw command line without quoting
func (b *BatchScriptBuilder) AddRawCommand(cmd string) *BatchScriptBuilder {
	b.mainCommands = append(b.mainCommands, cmd)
	return b
}

// AddCleanupCommand adds a cleanup command to run after the main job
func (b *BatchScriptBuilder) AddCleanupCommand(cmd string) *BatchScriptBuilder {
	b.cleanupCommands = append(b.cleanupCommands, cmd)
	return b
}

// AddHeaderComment adds a comment at the top of the script
func (b *BatchScriptBuilder) AddHeaderComment(comment string) *BatchScriptBuilder {
	b.headerComments = append(b.headerComments, comment)
	return b
}

// Build generates the complete SLURM batch script
func (b *BatchScriptBuilder) Build() string {
	var sb strings.Builder

	// Shebang
	sb.WriteString("#!/bin/bash\n")

	// Header comments
	for _, comment := range b.headerComments {
		sb.WriteString(fmt.Sprintf("# %s\n", comment))
	}
	if len(b.headerComments) > 0 {
		sb.WriteString("\n")
	}

	// SBATCH directives
	b.writeSBATCHDirectives(&sb)

	// Error handling
	sb.WriteString("\n# Exit on error\n")
	sb.WriteString("set -e\n")

	// Module loading
	if len(b.modules) > 0 || len(b.moduleCommands) > 0 {
		sb.WriteString("\n# Module loading\n")
		for _, cmd := range b.moduleCommands {
			sb.WriteString(cmd + "\n")
		}
		for _, module := range b.modules {
			sb.WriteString(fmt.Sprintf("module load %s\n", module))
		}
	}

	// Environment variables
	if len(b.environment) > 0 {
		sb.WriteString("\n# Environment variables\n")
		// Sort keys for deterministic output
		keys := make([]string, 0, len(b.environment))
		for k := range b.environment {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			sb.WriteString(fmt.Sprintf("export %s=%q\n", k, b.environment[k]))
		}
	}

	// Setup commands
	if len(b.setupCommands) > 0 {
		sb.WriteString("\n# Setup\n")
		for _, cmd := range b.setupCommands {
			sb.WriteString(cmd + "\n")
		}
	}

	// Job information
	sb.WriteString("\n# Job information\n")
	sb.WriteString("echo \"Job ID: $SLURM_JOB_ID\"\n")
	sb.WriteString("echo \"Job Name: $SLURM_JOB_NAME\"\n")
	sb.WriteString("echo \"Node List: $SLURM_JOB_NODELIST\"\n")
	sb.WriteString("echo \"Start Time: $(date)\"\n")

	// Main commands
	sb.WriteString("\n# Main execution\n")
	if b.containerImage != "" {
		containerCmd := "singularity exec"
		for _, opt := range b.containerOptions {
			containerCmd += " " + opt
		}
		containerCmd += " " + b.containerImage

		for _, cmd := range b.mainCommands {
			sb.WriteString(fmt.Sprintf("%s %s\n", containerCmd, cmd))
		}
	} else {
		for _, cmd := range b.mainCommands {
			sb.WriteString(cmd + "\n")
		}
	}

	// Capture exit code
	sb.WriteString("\nEXIT_CODE=$?\n")

	// Cleanup commands
	if len(b.cleanupCommands) > 0 {
		sb.WriteString("\n# Cleanup\n")
		for _, cmd := range b.cleanupCommands {
			sb.WriteString(cmd + "\n")
		}
	}

	// Final status
	sb.WriteString("\necho \"End Time: $(date)\"\n")
	sb.WriteString("echo \"Exit Code: $EXIT_CODE\"\n")
	sb.WriteString("exit $EXIT_CODE\n")

	return sb.String()
}

// writeSBATCHDirectives writes all #SBATCH directives
func (b *BatchScriptBuilder) writeSBATCHDirectives(sb *strings.Builder) {
	sb.WriteString("\n#==============================================================================\n")
	sb.WriteString("# SLURM Job Configuration\n")
	sb.WriteString("#==============================================================================\n")

	// Job identification
	if b.jobName != "" {
		sb.WriteString(fmt.Sprintf("#SBATCH --job-name=%s\n", b.jobName))
	}

	// Resource allocation
	if b.partition != "" {
		sb.WriteString(fmt.Sprintf("#SBATCH --partition=%s\n", b.partition))
	}

	if b.account != "" {
		sb.WriteString(fmt.Sprintf("#SBATCH --account=%s\n", b.account))
	}

	if b.qos != "" {
		sb.WriteString(fmt.Sprintf("#SBATCH --qos=%s\n", b.qos))
	}

	if b.reservation != "" {
		sb.WriteString(fmt.Sprintf("#SBATCH --reservation=%s\n", b.reservation))
	}

	// Node/CPU allocation
	if b.nodes > 0 {
		sb.WriteString(fmt.Sprintf("#SBATCH --nodes=%d\n", b.nodes))
	}

	if b.ntasks > 0 {
		sb.WriteString(fmt.Sprintf("#SBATCH --ntasks=%d\n", b.ntasks))
	}

	if b.cpusPerTask > 0 {
		sb.WriteString(fmt.Sprintf("#SBATCH --cpus-per-task=%d\n", b.cpusPerTask))
	}

	// Memory allocation
	if b.memoryMB > 0 {
		sb.WriteString(fmt.Sprintf("#SBATCH --mem=%dM\n", b.memoryMB))
	} else if b.memoryPerCPU > 0 {
		sb.WriteString(fmt.Sprintf("#SBATCH --mem-per-cpu=%dM\n", b.memoryPerCPU))
	}

	// Time limit
	if b.timeLimitMinutes > 0 {
		sb.WriteString(fmt.Sprintf("#SBATCH --time=%s\n", formatTimeLimit(b.timeLimitMinutes)))
	}

	// GPU allocation
	if b.gpus > 0 {
		if b.gpuType != "" {
			sb.WriteString(fmt.Sprintf("#SBATCH --gres=gpu:%s:%d\n", b.gpuType, b.gpus))
		} else {
			sb.WriteString(fmt.Sprintf("#SBATCH --gres=gpu:%d\n", b.gpus))
		}
	}

	if b.gpusPerNode > 0 {
		sb.WriteString(fmt.Sprintf("#SBATCH --gpus-per-node=%d\n", b.gpusPerNode))
	}

	if b.gpusPerTask > 0 {
		sb.WriteString(fmt.Sprintf("#SBATCH --gpus-per-task=%d\n", b.gpusPerTask))
	}

	// Working directory and output
	if b.workingDir != "" {
		sb.WriteString(fmt.Sprintf("#SBATCH --chdir=%s\n", b.workingDir))
	}

	if b.outputFile != "" {
		sb.WriteString(fmt.Sprintf("#SBATCH --output=%s\n", b.outputFile))
	}

	if b.errorFile != "" {
		sb.WriteString(fmt.Sprintf("#SBATCH --error=%s\n", b.errorFile))
	}

	// Node constraints and selection
	if b.exclusive {
		sb.WriteString("#SBATCH --exclusive\n")
	}

	for _, constraint := range b.constraints {
		sb.WriteString(fmt.Sprintf("#SBATCH --constraint=%s\n", constraint))
	}

	if b.nodelist != "" {
		sb.WriteString(fmt.Sprintf("#SBATCH --nodelist=%s\n", b.nodelist))
	}

	if b.excludeNodes != "" {
		sb.WriteString(fmt.Sprintf("#SBATCH --exclude=%s\n", b.excludeNodes))
	}

	// Job dependencies and arrays
	if b.dependency != "" {
		sb.WriteString(fmt.Sprintf("#SBATCH --dependency=%s\n", b.dependency))
	}

	if b.arraySpec != "" {
		sb.WriteString(fmt.Sprintf("#SBATCH --array=%s\n", b.arraySpec))
	}

	// Email notifications
	if b.mailType != "" {
		sb.WriteString(fmt.Sprintf("#SBATCH --mail-type=%s\n", b.mailType))
	}

	if b.mailUser != "" {
		sb.WriteString(fmt.Sprintf("#SBATCH --mail-user=%s\n", b.mailUser))
	}
}

// formatTimeLimit formats minutes into HH:MM:SS or D-HH:MM:SS format
func formatTimeLimit(minutes int64) string {
	hours := minutes / 60
	mins := minutes % 60

	if hours >= 24 {
		days := hours / 24
		hours %= 24
		return fmt.Sprintf("%d-%02d:%02d:00", days, hours, mins)
	}

	return fmt.Sprintf("%02d:%02d:00", hours, mins)
}

// BatchScriptOptions contains options for batch script generation
type BatchScriptOptions struct {
	// UseModulePurge clears all modules before loading new ones
	UseModulePurge bool

	// EnableProfiling adds profiling commands
	EnableProfiling bool

	// EnableTiming adds timing around main commands
	EnableTiming bool

	// StrictMode enables strict bash mode (set -euo pipefail)
	StrictMode bool
}

// DefaultBatchScriptOptions returns default batch script options
func DefaultBatchScriptOptions() BatchScriptOptions {
	return BatchScriptOptions{
		UseModulePurge:  false,
		EnableProfiling: false,
		EnableTiming:    false,
		StrictMode:      true,
	}
}
