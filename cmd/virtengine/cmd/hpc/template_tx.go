package hpc

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/virtengine/virtengine/pkg/hpc_workload_library"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

func newTemplateCreateCmd() *cobra.Command {
	var (
		manifestPath string
		publisher    string
		outputFormat string
		outputFile   string
	)

	cmd := &cobra.Command{
		Use:   "create [template-file]",
		Short: "Create a workload template spec",
		Long: `Create a workload template spec from a YAML/JSON manifest.

This command validates the manifest and emits a normalized template payload
suitable for governance workflows.`,
		Example: `  # Create a template spec
  virtengine hpc template create ./template.yaml --publisher ve1... --output json`,
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
			if format != templateOutputJSON && format != templateOutputYAML {
				return fmt.Errorf("unsupported output format for create: %s", format)
			}

			path := manifestPath
			if path == "" {
				path = args[0]
			}

			template, err := loadTemplateFromFile(path, publisher)
			if err != nil {
				return err
			}

			if template.Publisher == "" {
				return fmt.Errorf("template publisher missing: set publisher in manifest or pass --publisher")
			}

			if err := validateTemplate(cmd, template); err != nil {
				return err
			}

			data, err := marshalTemplateOutput(format, template)
			if err != nil {
				return err
			}

			return ensureOutputDestination(outputFile, data, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&manifestPath, "file", "", "Load template from a manifest file")
	cmd.Flags().StringVar(&publisher, "publisher", "", "Default publisher address to use if missing from manifest")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", templateOutputJSON, "Output format (json|yaml)")
	cmd.Flags().StringVar(&outputFile, "output-file", "", "Write output to a file instead of stdout")

	return cmd
}

func newTemplateUpdateCmd() *cobra.Command {
	var (
		manifestPath string
		publisher    string
		outputFormat string
		outputFile   string
		newVersion   string
		newStatus    string
	)

	cmd := &cobra.Command{
		Use:   "update [template-file]",
		Short: "Update a workload template spec",
		Long: `Update a workload template manifest and emit a normalized template payload.

Use flags to override version or approval status for governance workflows.`,
		Example: `  # Update a template version
  virtengine hpc template update ./template.yaml --set-version 1.2.0 --output json`,
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
			if format != templateOutputJSON && format != templateOutputYAML {
				return fmt.Errorf("unsupported output format for update: %s", format)
			}

			path := manifestPath
			if path == "" {
				path = args[0]
			}

			template, err := loadTemplateFromFile(path, publisher)
			if err != nil {
				return err
			}

			if newVersion != "" {
				template.Version = newVersion
			}

			if newStatus != "" {
				status := hpctypes.WorkloadApprovalStatus(newStatus)
				if !status.IsValid() {
					return fmt.Errorf("invalid approval status: %s", newStatus)
				}
				template.ApprovalStatus = status
			}

			template.UpdatedAt = time.Now().UTC()

			if template.Publisher == "" {
				return fmt.Errorf("template publisher missing: set publisher in manifest or pass --publisher")
			}

			if err := validateTemplate(cmd, template); err != nil {
				return err
			}

			data, err := marshalTemplateOutput(format, template)
			if err != nil {
				return err
			}

			return ensureOutputDestination(outputFile, data, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&manifestPath, "file", "", "Load template from a manifest file")
	cmd.Flags().StringVar(&publisher, "publisher", "", "Default publisher address to use if missing from manifest")
	cmd.Flags().StringVar(&newVersion, "set-version", "", "Override the template version")
	cmd.Flags().StringVar(&newStatus, "set-status", "", "Override the approval status (pending|approved|rejected|deprecated|revoked)")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", templateOutputJSON, "Output format (json|yaml)")
	cmd.Flags().StringVar(&outputFile, "output-file", "", "Write output to a file instead of stdout")

	return cmd
}

func newTemplateDeprecateCmd() *cobra.Command {
	var (
		manifestPath string
		publisher    string
		outputFormat string
		outputFile   string
	)

	cmd := &cobra.Command{
		Use:   "deprecate [template-id]",
		Short: "Deprecate a workload template",
		Long: `Mark a workload template as deprecated and emit the updated payload.

Use --file to deprecate a local manifest or provide a template ID to
use a built-in template as the base payload.`,
		Example: `  # Deprecate a local manifest
  virtengine hpc template deprecate --file ./template.yaml --output json

  # Deprecate a built-in template payload
  virtengine hpc template deprecate mpi-standard --output json`,
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
			if format != templateOutputJSON && format != templateOutputYAML {
				return fmt.Errorf("unsupported output format for deprecate: %s", format)
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

			template.ApprovalStatus = hpctypes.WorkloadApprovalDeprecated
			template.UpdatedAt = time.Now().UTC()

			if template.Publisher == "" {
				return fmt.Errorf("template publisher missing: set publisher in manifest or pass --publisher")
			}

			if err := validateTemplate(cmd, template); err != nil {
				return err
			}

			data, err := marshalTemplateOutput(format, template)
			if err != nil {
				return err
			}

			return ensureOutputDestination(outputFile, data, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&manifestPath, "file", "", "Load template from a manifest file")
	cmd.Flags().StringVar(&publisher, "publisher", "", "Default publisher address to use if missing from manifest")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", templateOutputJSON, "Output format (json|yaml)")
	cmd.Flags().StringVar(&outputFile, "output-file", "", "Write output to a file instead of stdout")

	return cmd
}

func validateTemplate(cmd *cobra.Command, template *hpctypes.WorkloadTemplate) error {
	config := hpc_workload_library.DefaultValidationConfig()
	validator := hpc_workload_library.NewWorkloadValidator(config)
	result := validator.ValidateTemplate(cmd.Context(), template)
	if result.IsValid() {
		return nil
	}
	return result.Error()
}
