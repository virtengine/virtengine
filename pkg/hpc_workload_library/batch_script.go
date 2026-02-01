// Package hpc_workload_library provides batch script generation.
//
// VE-5F: SLURM batch script generation from workload templates
package hpc_workload_library

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// BatchScriptConfig configures batch script generation
type BatchScriptConfig struct {
	// Account is the SLURM account
	Account string

	// Partition is the SLURM partition
	Partition string

	// QOS is the quality of service
	QOS string

	// Cluster is the cluster name
	Cluster string

	// JobName is the job name
	JobName string

	// OutputPath is the output file path
	OutputPath string

	// ErrorPath is the error file path
	ErrorPath string

	// MailUser is the email for notifications
	MailUser string

	// MailType specifies when to send email (BEGIN, END, FAIL, ALL)
	MailType string

	// Reservation is a reservation name
	Reservation string

	// Dependency specifies job dependencies
	Dependency string

	// CustomDirectives are additional SBATCH directives
	CustomDirectives map[string]string
}

// BatchScriptGenerator generates SLURM batch scripts from templates
type BatchScriptGenerator struct {
	config BatchScriptConfig
}

// NewBatchScriptGenerator creates a new batch script generator
func NewBatchScriptGenerator(config BatchScriptConfig) *BatchScriptGenerator {
	return &BatchScriptGenerator{config: config}
}

// GenerateScript generates a SLURM batch script from a template and job parameters
func (g *BatchScriptGenerator) GenerateScript(tmpl *hpctypes.WorkloadTemplate, params *JobParameters) (string, error) {
	if tmpl == nil {
		return "", fmt.Errorf("template is required")
	}
	if params == nil {
		params = &JobParameters{}
	}

	// Merge defaults
	params.applyDefaults(tmpl)

	var buf bytes.Buffer

	// Write shebang
	buf.WriteString("#!/bin/bash\n")
	buf.WriteString("#\n")
	buf.WriteString(fmt.Sprintf("# SLURM batch script generated from template: %s v%s\n", tmpl.TemplateID, tmpl.Version))
	buf.WriteString(fmt.Sprintf("# Template: %s\n", tmpl.Name))
	buf.WriteString("#\n\n")

	// Write SBATCH directives
	g.writeSBATCHDirectives(&buf, tmpl, params)

	// Write module loads
	g.writeModuleLoads(&buf, tmpl)

	// Write environment setup
	g.writeEnvironment(&buf, tmpl, params)

	// Write pre-run script
	g.writePreRunScript(&buf, tmpl)

	// Write main command
	g.writeMainCommand(&buf, tmpl, params)

	// Write post-run script
	g.writePostRunScript(&buf, tmpl)

	return buf.String(), nil
}

// writeSBATCHDirectives writes SBATCH directives
func (g *BatchScriptGenerator) writeSBATCHDirectives(buf *bytes.Buffer, tmpl *hpctypes.WorkloadTemplate, params *JobParameters) {
	buf.WriteString("# SLURM directives\n")

	// Job name
	jobName := g.config.JobName
	if jobName == "" {
		jobName = tmpl.TemplateID
	}
	buf.WriteString(fmt.Sprintf("#SBATCH --job-name=%s\n", jobName))

	// Nodes
	buf.WriteString(fmt.Sprintf("#SBATCH --nodes=%d\n", params.Nodes))

	// CPUs per node or ntasks
	if tmpl.Type == hpctypes.WorkloadTypeMPI {
		totalTasks := params.Nodes * params.TasksPerNode
		buf.WriteString(fmt.Sprintf("#SBATCH --ntasks=%d\n", totalTasks))
		buf.WriteString(fmt.Sprintf("#SBATCH --ntasks-per-node=%d\n", params.TasksPerNode))
	} else {
		buf.WriteString(fmt.Sprintf("#SBATCH --cpus-per-task=%d\n", params.CPUsPerNode))
	}

	// Memory
	buf.WriteString(fmt.Sprintf("#SBATCH --mem=%dM\n", params.MemoryMB))

	// Time limit
	buf.WriteString(fmt.Sprintf("#SBATCH --time=%s\n", formatTime(params.RuntimeMinutes)))

	// GPUs
	if params.GPUs > 0 {
		if params.GPUType != "" {
			buf.WriteString(fmt.Sprintf("#SBATCH --gres=gpu:%s:%d\n", params.GPUType, params.GPUs))
		} else {
			buf.WriteString(fmt.Sprintf("#SBATCH --gres=gpu:%d\n", params.GPUs))
		}
	}

	// Partition
	if g.config.Partition != "" {
		buf.WriteString(fmt.Sprintf("#SBATCH --partition=%s\n", g.config.Partition))
	}

	// Account
	if g.config.Account != "" {
		buf.WriteString(fmt.Sprintf("#SBATCH --account=%s\n", g.config.Account))
	}

	// QOS
	if g.config.QOS != "" {
		buf.WriteString(fmt.Sprintf("#SBATCH --qos=%s\n", g.config.QOS))
	}

	// Cluster
	if g.config.Cluster != "" {
		buf.WriteString(fmt.Sprintf("#SBATCH --cluster=%s\n", g.config.Cluster))
	}

	// Output/error paths
	outputPath := g.config.OutputPath
	if outputPath == "" {
		outputPath = "%x-%j.out"
	}
	buf.WriteString(fmt.Sprintf("#SBATCH --output=%s\n", outputPath))

	errorPath := g.config.ErrorPath
	if errorPath == "" {
		errorPath = "%x-%j.err"
	}
	buf.WriteString(fmt.Sprintf("#SBATCH --error=%s\n", errorPath))

	// Exclusive nodes
	if tmpl.Resources.ExclusiveNodes || params.Exclusive {
		buf.WriteString("#SBATCH --exclusive\n")
	}

	// Array job
	if params.ArrayStart >= 0 && params.ArrayEnd > params.ArrayStart {
		arraySpec := fmt.Sprintf("%d-%d", params.ArrayStart, params.ArrayEnd)
		if params.ArrayStep > 1 {
			arraySpec += fmt.Sprintf(":%d", params.ArrayStep)
		}
		if params.ArraySimultaneous > 0 {
			arraySpec += fmt.Sprintf("%%%d", params.ArraySimultaneous)
		}
		buf.WriteString(fmt.Sprintf("#SBATCH --array=%s\n", arraySpec))
	}

	// Mail notifications
	if g.config.MailUser != "" {
		buf.WriteString(fmt.Sprintf("#SBATCH --mail-user=%s\n", g.config.MailUser))
		mailType := g.config.MailType
		if mailType == "" {
			mailType = "END,FAIL"
		}
		buf.WriteString(fmt.Sprintf("#SBATCH --mail-type=%s\n", mailType))
	}

	// Reservation
	if g.config.Reservation != "" {
		buf.WriteString(fmt.Sprintf("#SBATCH --reservation=%s\n", g.config.Reservation))
	}

	// Dependency
	if g.config.Dependency != "" {
		buf.WriteString(fmt.Sprintf("#SBATCH --dependency=%s\n", g.config.Dependency))
	}

	// Constraints
	if len(params.Constraints) > 0 {
		buf.WriteString(fmt.Sprintf("#SBATCH --constraint=%s\n", strings.Join(params.Constraints, "&")))
	}

	// Custom directives
	for key, value := range g.config.CustomDirectives {
		buf.WriteString(fmt.Sprintf("#SBATCH --%s=%s\n", key, value))
	}

	buf.WriteString("\n")
}

// writeModuleLoads writes module load commands
func (g *BatchScriptGenerator) writeModuleLoads(buf *bytes.Buffer, tmpl *hpctypes.WorkloadTemplate) {
	modules := append(tmpl.Runtime.RequiredModules, tmpl.Modules...)
	if len(modules) == 0 {
		return
	}

	buf.WriteString("# Load required modules\n")
	buf.WriteString("module purge\n")
	for _, mod := range modules {
		buf.WriteString(fmt.Sprintf("module load %s\n", mod))
	}
	buf.WriteString("\n")
}

// writeEnvironment writes environment variable setup
func (g *BatchScriptGenerator) writeEnvironment(buf *bytes.Buffer, tmpl *hpctypes.WorkloadTemplate, params *JobParameters) {
	if len(tmpl.Environment) == 0 && len(params.Environment) == 0 {
		return
	}

	buf.WriteString("# Environment variables\n")

	// Template environment variables
	for _, env := range tmpl.Environment {
		value := env.Value
		if env.ValueTemplate != "" {
			value = env.ValueTemplate
		}
		if value != "" {
			buf.WriteString(fmt.Sprintf("export %s=\"%s\"\n", env.Name, value))
		}
	}

	// User-provided environment variables
	for key, value := range params.Environment {
		buf.WriteString(fmt.Sprintf("export %s=\"%s\"\n", key, value))
	}

	buf.WriteString("\n")
}

// writePreRunScript writes pre-run setup
func (g *BatchScriptGenerator) writePreRunScript(buf *bytes.Buffer, tmpl *hpctypes.WorkloadTemplate) {
	if tmpl.Entrypoint.PreRunScript == "" {
		return
	}

	buf.WriteString("# Pre-run setup\n")
	buf.WriteString(tmpl.Entrypoint.PreRunScript)
	buf.WriteString("\n\n")
}

// writeMainCommand writes the main execution command
func (g *BatchScriptGenerator) writeMainCommand(buf *bytes.Buffer, tmpl *hpctypes.WorkloadTemplate, params *JobParameters) {
	buf.WriteString("# Main execution\n")

	// Change to working directory
	workDir := tmpl.Entrypoint.WorkingDirectory
	if params.WorkingDirectory != "" {
		workDir = params.WorkingDirectory
	}
	if workDir != "" {
		buf.WriteString(fmt.Sprintf("cd %s || exit 1\n", workDir))
	}

	// Build command
	var cmdParts []string

	// MPI wrapper
	if tmpl.Entrypoint.UseMPIRun {
		mpiCmd := "srun"
		if tmpl.Runtime.MPIImplementation == "openmpi" {
			mpiCmd = "mpirun"
		}
		cmdParts = append(cmdParts, mpiCmd)
		cmdParts = append(cmdParts, tmpl.Entrypoint.MPIRunArgs...)
	}

	// Singularity wrapper
	if tmpl.Runtime.RuntimeType == "singularity" || tmpl.Runtime.RuntimeType == "apptainer" {
		runtime := tmpl.Runtime.RuntimeType
		cmdParts = append(cmdParts, runtime, "exec")

		// GPU flag
		if params.GPUs > 0 {
			cmdParts = append(cmdParts, "--nv")
		}

		// Bind paths
		if len(tmpl.DataBindings) > 0 {
			for _, binding := range tmpl.DataBindings {
				if binding.HostPath != "" {
					bindSpec := fmt.Sprintf("%s:%s", binding.HostPath, binding.MountPath)
					if binding.ReadOnly {
						bindSpec += ":ro"
					}
					cmdParts = append(cmdParts, "-B", bindSpec)
				}
			}
		}

		// Container image
		cmdParts = append(cmdParts, tmpl.Runtime.ContainerImage)
	}

	// Main command
	mainCmd := tmpl.Entrypoint.Command
	if params.Command != "" {
		mainCmd = params.Command
	}
	cmdParts = append(cmdParts, mainCmd)

	// Arguments
	args := tmpl.Entrypoint.DefaultArgs
	if len(params.Arguments) > 0 {
		args = params.Arguments
	}
	cmdParts = append(cmdParts, args...)

	// User script/executable
	if params.Script != "" {
		cmdParts = append(cmdParts, params.Script)
	}

	buf.WriteString(strings.Join(cmdParts, " "))
	buf.WriteString("\n\n")

	// Capture exit code
	buf.WriteString("EXIT_CODE=$?\n\n")
}

// writePostRunScript writes post-run cleanup
func (g *BatchScriptGenerator) writePostRunScript(buf *bytes.Buffer, tmpl *hpctypes.WorkloadTemplate) {
	buf.WriteString("# Post-run cleanup\n")

	if tmpl.Entrypoint.PostRunScript != "" {
		buf.WriteString(tmpl.Entrypoint.PostRunScript)
		buf.WriteString("\n")
	}

	buf.WriteString("\nexit $EXIT_CODE\n")
}

// JobParameters contains job-specific parameters
type JobParameters struct {
	// Nodes is the number of nodes
	Nodes int32

	// CPUsPerNode is CPUs per node
	CPUsPerNode int32

	// TasksPerNode is MPI tasks per node
	TasksPerNode int32

	// MemoryMB is memory in MB
	MemoryMB int64

	// RuntimeMinutes is runtime in minutes
	RuntimeMinutes int64

	// GPUs is number of GPUs
	GPUs int32

	// GPUType is the GPU type
	GPUType string

	// Exclusive requests exclusive nodes
	Exclusive bool

	// Command is the main command
	Command string

	// Arguments are command arguments
	Arguments []string

	// Script is the user script
	Script string

	// WorkingDirectory is the working directory
	WorkingDirectory string

	// Environment contains additional environment variables
	Environment map[string]string

	// ArrayStart is array job start index (-1 for no array)
	ArrayStart int

	// ArrayEnd is array job end index
	ArrayEnd int

	// ArrayStep is array job step
	ArrayStep int

	// ArraySimultaneous is max simultaneous array tasks
	ArraySimultaneous int

	// Constraints are node constraints
	Constraints []string
}

// applyDefaults applies template defaults to parameters
func (p *JobParameters) applyDefaults(tmpl *hpctypes.WorkloadTemplate) {
	if p.Nodes == 0 {
		p.Nodes = tmpl.Resources.DefaultNodes
	}
	if p.CPUsPerNode == 0 {
		p.CPUsPerNode = tmpl.Resources.DefaultCPUsPerNode
	}
	if p.TasksPerNode == 0 {
		p.TasksPerNode = tmpl.Resources.DefaultCPUsPerNode
	}
	if p.MemoryMB == 0 {
		p.MemoryMB = tmpl.Resources.DefaultMemoryMBPerNode
	}
	if p.RuntimeMinutes == 0 {
		p.RuntimeMinutes = tmpl.Resources.DefaultRuntimeMinutes
	}
	if p.GPUs == 0 {
		p.GPUs = tmpl.Resources.DefaultGPUsPerNode
	}
	if p.ArrayStart < 0 {
		p.ArrayStart = -1
	}
	if p.ArrayStep == 0 {
		p.ArrayStep = 1
	}
}

// formatTime formats minutes as HH:MM:SS
func formatTime(minutes int64) string {
	hours := minutes / 60
	mins := minutes % 60
	if hours > 24 {
		days := hours / 24
		hours = hours % 24
		return fmt.Sprintf("%d-%02d:%02d:00", days, hours, mins)
	}
	return fmt.Sprintf("%02d:%02d:00", hours, mins)
}

// BatchScriptTemplateData contains data for script templating
type BatchScriptTemplateData struct {
	Template   *hpctypes.WorkloadTemplate
	Parameters *JobParameters
	Config     *BatchScriptConfig
}

// templateFuncs contains functions for template rendering
var templateFuncs = template.FuncMap{
	"formatTime": formatTime,
}

// scriptTemplate is the Go template for batch scripts (alternative approach)
var scriptTemplate = template.Must(template.New("batchscript").Funcs(templateFuncs).Parse(`#!/bin/bash
#
# SLURM batch script generated from template: {{.Template.TemplateID}} v{{.Template.Version}}
# Template: {{.Template.Name}}
#

# SLURM directives
#SBATCH --job-name={{.Config.JobName}}
#SBATCH --nodes={{.Parameters.Nodes}}
#SBATCH --cpus-per-task={{.Parameters.CPUsPerNode}}
#SBATCH --mem={{.Parameters.MemoryMB}}M
#SBATCH --time={{.Parameters.RuntimeMinutes | formatTime}}
{{if .Config.Partition}}#SBATCH --partition={{.Config.Partition}}{{end}}
{{if .Config.Account}}#SBATCH --account={{.Config.Account}}{{end}}
#SBATCH --output=%x-%j.out
#SBATCH --error=%x-%j.err

{{if .Template.Modules}}
# Load modules
module purge
{{range .Template.Modules}}module load {{.}}
{{end}}{{end}}

# Environment
{{range .Template.Environment}}export {{.Name}}="{{if .Value}}{{.Value}}{{else}}{{.ValueTemplate}}{{end}}"
{{end}}

{{if .Template.Entrypoint.PreRunScript}}
# Pre-run
{{.Template.Entrypoint.PreRunScript}}
{{end}}

# Main execution
{{if .Template.Entrypoint.WorkingDirectory}}cd {{.Template.Entrypoint.WorkingDirectory}} || exit 1{{end}}
{{.Template.Entrypoint.Command}} {{range .Template.Entrypoint.DefaultArgs}}{{.}} {{end}}{{.Parameters.Script}}

EXIT_CODE=$?

{{if .Template.Entrypoint.PostRunScript}}
# Post-run
{{.Template.Entrypoint.PostRunScript}}
{{end}}

exit $EXIT_CODE
`))

