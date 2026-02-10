// Package hpc provides CLI commands for HPC workload template management.
package hpc

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

const (
	templateOutputTable = "table"
	templateOutputJSON  = "json"
	templateOutputYAML  = "yaml"
	templateOutputText  = "text"
)

// GetTemplateCmd returns the template command group for the hpc CLI.
func GetTemplateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Manage HPC workload templates",
		Long: `Commands for listing, validating, and managing HPC workload templates.

These commands operate on local template specs and built-in templates
used for HPC workload submission workflows.`,
	}

	cmd.AddCommand(
		newTemplateListCmd(),
		newTemplateShowCmd(),
		newTemplateValidateCmd(),
		newTemplateCreateCmd(),
		newTemplateUpdateCmd(),
		newTemplateDeprecateCmd(),
	)

	return cmd
}

// GetTemplatesCmd returns the legacy top-level templates command.
func GetTemplatesCmd() *cobra.Command {
	cmd := GetTemplateCmd()
	cmd.Use = "hpc-templates"
	cmd.Short = "Manage HPC workload templates (legacy)"
	cmd.Long = `Legacy command group for workload templates.

Use "virtengine hpc template" instead.`
	cmd.Deprecated = "use \"virtengine hpc template\""
	return cmd
}

func normalizeTemplateOutput(format string) (string, error) {
	if format == "" {
		return templateOutputTable, nil
	}

	switch strings.ToLower(format) {
	case templateOutputTable, templateOutputJSON, templateOutputYAML, templateOutputText:
		return strings.ToLower(format), nil
	default:
		return "", fmt.Errorf("invalid output format: %s", format)
	}
}

func writeTemplateOutput(w io.Writer, format string, value interface{}) error {
	data, err := marshalTemplateOutput(format, value)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(w, string(data))
	return err
}

func marshalTemplateOutput(format string, value interface{}) ([]byte, error) {
	format, err := normalizeTemplateOutput(format)
	if err != nil {
		return nil, err
	}

	switch format {
	case templateOutputJSON:
		return json.MarshalIndent(value, "", "  ")
	case templateOutputYAML:
		return yaml.Marshal(value)
	default:
		return nil, fmt.Errorf("unsupported output format: %s", format)
	}
}

func renderTemplateTable(w io.Writer, templates []*hpctypes.WorkloadTemplate) error {
	writer := tabwriter.NewWriter(w, 0, 0, 2, byte(32), 0)
	fmt.Fprintln(writer, "ID\tNAME\tVERSION\tTYPE\tSTATUS\tTAGS")
	fmt.Fprintln(writer, "--\t----\t-------\t----\t------\t----")

	for _, t := range templates {
		tags := strings.Join(t.Tags, ", ")
		if len(tags) > 30 {
			tags = tags[:27] + "..."
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			t.TemplateID,
			truncateTemplateText(t.Name, 25),
			t.Version,
			t.Type,
			t.ApprovalStatus,
			tags,
		)
	}

	return writer.Flush()
}

func renderTemplateDetails(w io.Writer, t *hpctypes.WorkloadTemplate) {
	fmt.Fprintf(w, "Template: %s\n", t.TemplateID)
	fmt.Fprintf(w, "Name: %s\n", t.Name)
	fmt.Fprintf(w, "Version: %s\n", t.Version)
	fmt.Fprintf(w, "Type: %s\n", t.Type)
	fmt.Fprintf(w, "Status: %s\n", t.ApprovalStatus)
	fmt.Fprintf(w, "Publisher: %s\n", t.Publisher)
	fmt.Fprintln(w)

	if t.Description != "" {
		fmt.Fprintln(w, "Description:")
		fmt.Fprintf(w, "  %s\n\n", t.Description)
	}

	fmt.Fprintln(w, "Runtime:")
	fmt.Fprintf(w, "  Type: %s\n", t.Runtime.RuntimeType)
	if t.Runtime.ContainerImage != "" {
		fmt.Fprintf(w, "  Container: %s\n", t.Runtime.ContainerImage)
	}
	if t.Runtime.MPIImplementation != "" {
		fmt.Fprintf(w, "  MPI: %s\n", t.Runtime.MPIImplementation)
	}
	if t.Runtime.CUDAVersion != "" {
		fmt.Fprintf(w, "  CUDA: %s\n", t.Runtime.CUDAVersion)
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, "Resources:")
	fmt.Fprintf(w, "  Nodes: %d-%d (default: %d)\n", t.Resources.MinNodes, t.Resources.MaxNodes, t.Resources.DefaultNodes)
	fmt.Fprintf(w, "  CPUs/Node: %d-%d (default: %d)\n", t.Resources.MinCPUsPerNode, t.Resources.MaxCPUsPerNode, t.Resources.DefaultCPUsPerNode)
	fmt.Fprintf(w, "  Memory/Node: %dMB-%dMB (default: %dMB)\n", t.Resources.MinMemoryMBPerNode, t.Resources.MaxMemoryMBPerNode, t.Resources.DefaultMemoryMBPerNode)
	if t.Resources.MaxGPUsPerNode > 0 {
		fmt.Fprintf(w, "  GPUs/Node: %d-%d (default: %d)\n", t.Resources.MinGPUsPerNode, t.Resources.MaxGPUsPerNode, t.Resources.DefaultGPUsPerNode)
		if len(t.Resources.GPUTypes) > 0 {
			fmt.Fprintf(w, "  GPU Types: %s\n", strings.Join(t.Resources.GPUTypes, ", "))
		}
	}
	fmt.Fprintf(w, "  Runtime: %d-%d minutes (default: %d)\n", t.Resources.MinRuntimeMinutes, t.Resources.MaxRuntimeMinutes, t.Resources.DefaultRuntimeMinutes)
	fmt.Fprintln(w)

	fmt.Fprintln(w, "Security:")
	fmt.Fprintf(w, "  Sandbox: %s\n", t.Security.SandboxLevel)
	fmt.Fprintf(w, "  Network Access: %v\n", t.Security.AllowNetworkAccess)
	fmt.Fprintf(w, "  Host Mounts: %v\n", t.Security.AllowHostMounts)
	if t.Security.RequireImageDigest {
		fmt.Fprintf(w, "  Require Digest: %v\n", t.Security.RequireImageDigest)
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, "Entrypoint:")
	fmt.Fprintf(w, "  Command: %s\n", t.Entrypoint.Command)
	if len(t.Entrypoint.DefaultArgs) > 0 {
		fmt.Fprintf(w, "  Default Args: %s\n", strings.Join(t.Entrypoint.DefaultArgs, " "))
	}
	if t.Entrypoint.WorkingDirectory != "" {
		fmt.Fprintf(w, "  Working Dir: %s\n", t.Entrypoint.WorkingDirectory)
	}
	fmt.Fprintln(w)

	if len(t.ParameterSchema) > 0 {
		fmt.Fprintln(w, "Parameters:")
		for _, p := range t.ParameterSchema {
			required := ""
			if p.Required {
				required = " (required)"
			}
			fmt.Fprintf(w, "  %s [%s]%s: %s\n", p.Name, p.Type, required, p.Description)
			if p.Default != "" {
				fmt.Fprintf(w, "    Default: %s\n", p.Default)
			}
		}
		fmt.Fprintln(w)
	}

	if len(t.Tags) > 0 {
		fmt.Fprintf(w, "Tags: %s\n", strings.Join(t.Tags, ", "))
	}
}

func truncateTemplateText(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func writeTemplateOutputToFile(path string, data []byte) error {
	if path == "" {
		return nil
	}
	return os.WriteFile(path, data, 0o600)
}

func ensureOutputDestination(path string, data []byte, w io.Writer) error {
	if path != "" {
		if err := writeTemplateOutputToFile(path, data); err != nil {
			return err
		}
		fmt.Fprintf(w, "Wrote output to %s\n", path)
		return nil
	}

	_, err := fmt.Fprintln(w, string(data))
	return err
}
