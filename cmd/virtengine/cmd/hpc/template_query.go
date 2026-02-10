package hpc

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/virtengine/virtengine/pkg/hpc_workload_library"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

type templateFilter struct {
	workloadType string
	status       string
	tags         []string
}

type templateValidationReport struct {
	TemplateID string                                   `json:"template_id" yaml:"template_id"`
	Version    string                                   `json:"version" yaml:"version"`
	Valid      bool                                     `json:"valid" yaml:"valid"`
	Errors     []hpc_workload_library.ValidationError   `json:"errors,omitempty" yaml:"errors,omitempty"`
	Warnings   []hpc_workload_library.ValidationWarning `json:"warnings,omitempty" yaml:"warnings,omitempty"`
}

func newTemplateListCmd() *cobra.Command {
	var (
		workloadType string
		status       string
		tags         []string
		outputFormat string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available HPC workload templates",
		Long: `List all available HPC workload templates.

Use filters to narrow results by workload type, approval status, or tags.`,
		Example: `  # List all built-in templates
  virtengine hpc template list

  # Filter by workload type
  virtengine hpc template list --type gpu

  # Output as JSON
  virtengine hpc template list --output json`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			format, err := normalizeTemplateOutput(outputFormat)
			if err != nil {
				return err
			}

			filter := templateFilter{
				workloadType: workloadType,
				status:       status,
				tags:         tags,
			}

			templates, err := filterTemplates(hpc_workload_library.GetBuiltinTemplates(), filter)
			if err != nil {
				return err
			}

			sort.Slice(templates, func(i, j int) bool {
				return templates[i].TemplateID < templates[j].TemplateID
			})

			out := cmd.OutOrStdout()
			switch format {
			case templateOutputTable:
				return renderTemplateTable(out, templates)
			case templateOutputJSON, templateOutputYAML:
				return writeTemplateOutput(out, format, templates)
			default:
				return fmt.Errorf("unsupported output format for list: %s", format)
			}
		},
	}

	cmd.Flags().StringVarP(&workloadType, "type", "t", "", "Filter by workload type (mpi|gpu|batch|data_processing|interactive|custom)")
	cmd.Flags().StringVar(&status, "status", "", "Filter by approval status (pending|approved|rejected|deprecated|revoked)")
	cmd.Flags().StringSliceVar(&tags, "tag", nil, "Filter by tag (repeatable)")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", templateOutputTable, "Output format (table|json|yaml)")

	return cmd
}

func newTemplateShowCmd() *cobra.Command {
	var (
		outputFormat string
		manifestPath string
		publisher    string
	)

	cmd := &cobra.Command{
		Use:   "show [template-id]",
		Short: "Show details of a workload template",
		Long:  "Display detailed information about a specific workload template.",
		Example: `  # Show template details
  virtengine hpc template show mpi-standard

  # Show a local manifest
  virtengine hpc template show --file ./template.yaml

  # Output as JSON
  virtengine hpc template show gpu-compute --output json`,
		Args: func(cmd *cobra.Command, args []string) error {
			if manifestPath == "" && len(args) != 1 {
				return fmt.Errorf("template-id is required unless --file is set")
			}
			if manifestPath != "" && len(args) > 0 {
				return fmt.Errorf("do not provide template-id when using --file")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			format, err := normalizeTemplateOutput(outputFormat)
			if err != nil {
				return err
			}

			var template *hpctypes.WorkloadTemplate
			switch {
			case manifestPath != "":
				template, err = loadTemplateFromFile(manifestPath, publisher)
				if err != nil {
					return err
				}
			default:
				template = hpc_workload_library.GetTemplateByID(args[0])
				if template == nil {
					return fmt.Errorf("template not found: %s", args[0])
				}
			}

			out := cmd.OutOrStdout()
			switch format {
			case templateOutputText:
				renderTemplateDetails(out, template)
				return nil
			case templateOutputJSON, templateOutputYAML:
				return writeTemplateOutput(out, format, template)
			default:
				return fmt.Errorf("unsupported output format for show: %s", format)
			}
		},
	}

	cmd.Flags().StringVar(&manifestPath, "file", "", "Load template from a manifest file")
	cmd.Flags().StringVar(&publisher, "publisher", "", "Default publisher address to use if missing from manifest")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", templateOutputText, "Output format (text|json|yaml)")

	return cmd
}

func newTemplateValidateCmd() *cobra.Command {
	var (
		manifestPath  string
		publisher     string
		outputFormat  string
		requireSigned bool
	)

	cmd := &cobra.Command{
		Use:     "validate [template-file]",
		Aliases: []string{"verify"},
		Short:   "Validate a workload template spec",
		Long: `Validate a workload template manifest against schema and policy rules.

Validation includes schema checks and security/resource policy checks.`,
		Example: `  # Validate a template manifest
  virtengine hpc template validate ./template.yaml

  # Require signature validation
  virtengine hpc template validate ./template.yaml --require-signature`,
		Args: func(cmd *cobra.Command, args []string) error {
			if manifestPath == "" && len(args) != 1 {
				return fmt.Errorf("template file is required unless --file is set")
			}
			if manifestPath != "" && len(args) > 0 {
				return fmt.Errorf("do not provide template file when using --file")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			format, err := normalizeTemplateOutput(outputFormat)
			if err != nil {
				return err
			}

			path := manifestPath
			if path == "" {
				path = args[0]
			}

			template, err := loadTemplateFromFile(path, publisher)
			if err != nil {
				return err
			}

			config := hpc_workload_library.DefaultValidationConfig()
			config.RequireSignedTemplate = requireSigned
			validator := hpc_workload_library.NewWorkloadValidator(config)
			result := validator.ValidateTemplate(cmd.Context(), template)

			report := templateValidationReport{
				TemplateID: template.TemplateID,
				Version:    template.Version,
				Valid:      result.IsValid(),
				Errors:     result.Errors,
				Warnings:   result.Warnings,
			}

			out := cmd.OutOrStdout()
			switch format {
			case templateOutputText:
				renderTemplateValidation(out, template, result)
			case templateOutputJSON, templateOutputYAML:
				if err := writeTemplateOutput(out, format, report); err != nil {
					return err
				}
			default:
				return fmt.Errorf("unsupported output format for validate: %s", format)
			}

			if !result.IsValid() {
				return result.Error()
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&manifestPath, "file", "", "Load template from a manifest file")
	cmd.Flags().StringVar(&publisher, "publisher", "", "Default publisher address to use if missing from manifest")
	cmd.Flags().BoolVar(&requireSigned, "require-signature", false, "Require signature validation")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", templateOutputText, "Output format (text|json|yaml)")

	return cmd
}

func loadTemplateFromFile(path, publisher string) (*hpctypes.WorkloadTemplate, error) {
	loader := hpc_workload_library.NewManifestLoader(publisher)
	return loader.LoadFromFile(path)
}

func renderTemplateValidation(w io.Writer, template *hpctypes.WorkloadTemplate, result *hpc_workload_library.ValidationResult) {
	fmt.Fprintf(w, "Template: %s\n", template.TemplateID)
	fmt.Fprintf(w, "Version: %s\n", template.Version)
	fmt.Fprintf(w, "Valid: %v\n", result.IsValid())

	if len(result.Errors) > 0 {
		fmt.Fprintln(w, "Errors:")
		for _, err := range result.Errors {
			fmt.Fprintf(w, "  - %s: %s\n", err.Field, err.Message)
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Fprintln(w, "Warnings:")
		for _, warn := range result.Warnings {
			fmt.Fprintf(w, "  - %s: %s\n", warn.Field, warn.Message)
		}
	}
}

func filterTemplates(templates []*hpctypes.WorkloadTemplate, filter templateFilter) ([]*hpctypes.WorkloadTemplate, error) {
	filtered := templates
	if filter.workloadType != "" {
		wType := hpctypes.WorkloadType(filter.workloadType)
		if !wType.IsValid() {
			return nil, fmt.Errorf("invalid workload type: %s", filter.workloadType)
		}
		var items []*hpctypes.WorkloadTemplate
		for _, template := range filtered {
			if template.Type == wType {
				items = append(items, template)
			}
		}
		filtered = items
	}

	if filter.status != "" {
		status := hpctypes.WorkloadApprovalStatus(filter.status)
		if !status.IsValid() {
			return nil, fmt.Errorf("invalid approval status: %s", filter.status)
		}
		var items []*hpctypes.WorkloadTemplate
		for _, template := range filtered {
			if template.ApprovalStatus == status {
				items = append(items, template)
			}
		}
		filtered = items
	}

	if len(filter.tags) > 0 {
		var items []*hpctypes.WorkloadTemplate
		for _, template := range filtered {
			if templateHasTag(template, filter.tags) {
				items = append(items, template)
			}
		}
		filtered = items
	}

	return filtered, nil
}

func templateHasTag(template *hpctypes.WorkloadTemplate, tags []string) bool {
	if len(tags) == 0 {
		return true
	}

	for _, target := range tags {
		for _, tag := range template.Tags {
			if strings.EqualFold(tag, target) {
				return true
			}
		}
	}

	return false
}
