// Package hpc provides CLI commands for HPC workload template management.
//
// VE-5F: CLI commands to list and verify workload templates
package hpc

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/virtengine/virtengine/pkg/hpc_workload_library"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// GetTemplatesCmd returns the templates command group
func GetTemplatesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hpc-templates",
		Short: "Manage HPC workload templates",
		Long: `Commands for listing, viewing, and verifying HPC workload templates.
Preferred usage: "virtengine hpc-templates". Legacy usage: "virtengine hpc templates".`,
	}

	cmd.AddCommand(
		getListTemplatesCmd(),
		getShowTemplateCmd(),
		getVerifyTemplateCmd(),
		getTemplateTypesCmd(),
	)

	return cmd
}

func getListTemplatesCmd() *cobra.Command {
	var (
		workloadType string
		outputFormat string
		showAll      bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available HPC workload templates",
		Long: `List all available HPC workload templates. By default, only built-in 
templates are shown. Use --all to include on-chain templates.`,
		Example: `  # List all built-in templates
  virtengine hpc-templates list

  # List templates of a specific type
  virtengine hpc-templates list --type gpu

  # Output as JSON
  virtengine hpc-templates list --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			templates := hpc_workload_library.GetBuiltinTemplates()

			// Filter by type if specified
			if workloadType != "" {
				wlType := hpctypes.WorkloadType(workloadType)
				if !wlType.IsValid() {
					return fmt.Errorf("invalid workload type: %s", workloadType)
				}

				var filtered []*hpctypes.WorkloadTemplate
				for _, t := range templates {
					if t.Type == wlType {
						filtered = append(filtered, t)
					}
				}
				templates = filtered
			}

			// Output
			switch outputFormat {
			case "json":
				return outputJSON(templates)
			case "yaml":
				return outputJSON(templates) // Simplified for now
			default:
				return outputTable(templates)
			}
		},
	}

	cmd.Flags().StringVarP(&workloadType, "type", "t", "", "Filter by workload type (mpi|gpu|batch|data_processing|interactive)")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table|json|yaml)")
	cmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show all templates including on-chain")

	return cmd
}

func getShowTemplateCmd() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:     "show <template-id>",
		Aliases: []string{"get"},
		Short:   "Show details of a workload template",
		Long:    `Display detailed information about a specific workload template.`,
		Example: `  # Show template details
  virtengine hpc-templates show mpi-standard

  # Output as JSON
  virtengine hpc-templates show gpu-compute --output json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateID := args[0]

			template := hpc_workload_library.GetTemplateByID(templateID)
			if template == nil {
				return fmt.Errorf("template not found: %s", templateID)
			}

			switch outputFormat {
			case "json":
				return outputJSON(template)
			default:
				return outputTemplateDetails(template)
			}
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Output format (text|json)")

	return cmd
}

func getVerifyTemplateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "verify <template-id>",
		Aliases: []string{"validate"},
		Short:   "Verify a workload template signature",
		Long: `Verify the cryptographic signature of a workload template to ensure
it has not been tampered with.`,
		Example: `  # Verify a template
  virtengine hpc-templates verify mpi-standard`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateID := args[0]

			template := hpc_workload_library.GetTemplateByID(templateID)
			if template == nil {
				return fmt.Errorf("template not found: %s", templateID)
			}

			// Check if template has a signature
			if template.Signature.Signature == "" {
				fmt.Printf("Template %s: ⚠️  No signature (built-in template)\n", templateID)
				return nil
			}

			// Verify signature
			verifier := hpc_workload_library.NewTemplateVerifier()
			if err := verifier.Verify(template); err != nil {
				fmt.Printf("Template %s: ❌ Signature verification FAILED: %s\n", templateID, err)
				return err
			}

			fmt.Printf("Template %s: ✅ Signature verified\n", templateID)
			fmt.Printf("  Publisher: %s\n", template.Publisher)
			fmt.Printf("  Algorithm: %s\n", template.Signature.Algorithm)
			fmt.Printf("  Signed At: %s\n", template.Signature.SignedAt.Format("2006-01-02 15:04:05 UTC"))
			return nil
		},
	}

	return cmd
}

func getTemplateTypesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "types",
		Short: "List available workload types",
		Long:  `Display all available workload types and their descriptions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			types := []struct {
				Type        hpctypes.WorkloadType
				Description string
			}{
				{hpctypes.WorkloadTypeMPI, "MPI-based parallel workloads for distributed computing"},
				{hpctypes.WorkloadTypeGPU, "GPU-accelerated compute workloads (CUDA, deep learning)"},
				{hpctypes.WorkloadTypeBatch, "Batch processing for single-node or serial tasks"},
				{hpctypes.WorkloadTypeDataProcessing, "Data processing pipelines (Spark, Dask)"},
				{hpctypes.WorkloadTypeInteractive, "Interactive sessions (JupyterLab, terminal)"},
				{hpctypes.WorkloadTypeCustom, "Custom user-defined workloads"},
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "TYPE\tDESCRIPTION")
			fmt.Fprintln(w, "----\t-----------")
			for _, t := range types {
				fmt.Fprintf(w, "%s\t%s\n", t.Type, t.Description)
			}
			return w.Flush()
		},
	}

	return cmd
}

// Output helpers

func outputTable(templates []*hpctypes.WorkloadTemplate) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tVERSION\tTYPE\tSTATUS\tTAGS")
	fmt.Fprintln(w, "--\t----\t-------\t----\t------\t----")

	for _, t := range templates {
		tags := strings.Join(t.Tags, ", ")
		if len(tags) > 30 {
			tags = tags[:27] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			t.TemplateID,
			truncate(t.Name, 25),
			t.Version,
			t.Type,
			t.ApprovalStatus,
			tags,
		)
	}

	return w.Flush()
}

func outputJSON(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func outputTemplateDetails(t *hpctypes.WorkloadTemplate) error {
	fmt.Printf("Template: %s\n", t.TemplateID)
	fmt.Printf("Name: %s\n", t.Name)
	fmt.Printf("Version: %s\n", t.Version)
	fmt.Printf("Type: %s\n", t.Type)
	fmt.Printf("Status: %s\n", t.ApprovalStatus)
	fmt.Printf("Publisher: %s\n", t.Publisher)
	fmt.Println()

	fmt.Println("Description:")
	fmt.Printf("  %s\n", t.Description)
	fmt.Println()

	fmt.Println("Runtime:")
	fmt.Printf("  Type: %s\n", t.Runtime.RuntimeType)
	if t.Runtime.ContainerImage != "" {
		fmt.Printf("  Container: %s\n", t.Runtime.ContainerImage)
	}
	if t.Runtime.MPIImplementation != "" {
		fmt.Printf("  MPI: %s\n", t.Runtime.MPIImplementation)
	}
	if t.Runtime.CUDAVersion != "" {
		fmt.Printf("  CUDA: %s\n", t.Runtime.CUDAVersion)
	}
	fmt.Println()

	fmt.Println("Resources:")
	fmt.Printf("  Nodes: %d-%d (default: %d)\n", t.Resources.MinNodes, t.Resources.MaxNodes, t.Resources.DefaultNodes)
	fmt.Printf("  CPUs/Node: %d-%d (default: %d)\n", t.Resources.MinCPUsPerNode, t.Resources.MaxCPUsPerNode, t.Resources.DefaultCPUsPerNode)
	fmt.Printf("  Memory/Node: %dMB-%dMB (default: %dMB)\n", t.Resources.MinMemoryMBPerNode, t.Resources.MaxMemoryMBPerNode, t.Resources.DefaultMemoryMBPerNode)
	if t.Resources.MaxGPUsPerNode > 0 {
		fmt.Printf("  GPUs/Node: %d-%d (default: %d)\n", t.Resources.MinGPUsPerNode, t.Resources.MaxGPUsPerNode, t.Resources.DefaultGPUsPerNode)
		if len(t.Resources.GPUTypes) > 0 {
			fmt.Printf("  GPU Types: %s\n", strings.Join(t.Resources.GPUTypes, ", "))
		}
	}
	fmt.Printf("  Runtime: %d-%d minutes (default: %d)\n", t.Resources.MinRuntimeMinutes, t.Resources.MaxRuntimeMinutes, t.Resources.DefaultRuntimeMinutes)
	fmt.Println()

	fmt.Println("Security:")
	fmt.Printf("  Sandbox: %s\n", t.Security.SandboxLevel)
	fmt.Printf("  Network Access: %v\n", t.Security.AllowNetworkAccess)
	fmt.Printf("  Host Mounts: %v\n", t.Security.AllowHostMounts)
	if t.Security.RequireImageDigest {
		fmt.Printf("  Require Digest: %v\n", t.Security.RequireImageDigest)
	}
	fmt.Println()

	fmt.Println("Entrypoint:")
	fmt.Printf("  Command: %s\n", t.Entrypoint.Command)
	if len(t.Entrypoint.DefaultArgs) > 0 {
		fmt.Printf("  Default Args: %s\n", strings.Join(t.Entrypoint.DefaultArgs, " "))
	}
	if t.Entrypoint.WorkingDirectory != "" {
		fmt.Printf("  Working Dir: %s\n", t.Entrypoint.WorkingDirectory)
	}
	fmt.Println()

	if len(t.ParameterSchema) > 0 {
		fmt.Println("Parameters:")
		for _, p := range t.ParameterSchema {
			required := ""
			if p.Required {
				required = " (required)"
			}
			fmt.Printf("  %s [%s]%s: %s\n", p.Name, p.Type, required, p.Description)
			if p.Default != "" {
				fmt.Printf("    Default: %s\n", p.Default)
			}
		}
		fmt.Println()
	}

	if len(t.Tags) > 0 {
		fmt.Printf("Tags: %s\n", strings.Join(t.Tags, ", "))
	}

	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
