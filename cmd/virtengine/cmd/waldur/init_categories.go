// Package waldur provides CLI commands for Waldur integration.
//
// VE-25A: CLI command to initialize marketplace categories in Waldur.
package waldur

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/virtengine/virtengine/pkg/provider_daemon"
	"github.com/virtengine/virtengine/pkg/waldur"
)

const (
	flagWaldurURL   = "waldur-url"
	flagWaldurToken = "waldur-token" //nolint:gosec // #nosec G101: CLI flag name, not a credential
	flagOutput      = "output"
	flagTimeout     = "timeout"
)

func getInitCategoriesCmd() *cobra.Command {
	var (
		waldurURL   string
		waldurToken string
		output      string
		timeout     int
	)

	cmd := &cobra.Command{
		Use:   "init-categories",
		Short: "Initialize marketplace categories in Waldur",
		Long: `Initialize the default VirtEngine marketplace categories in Waldur.

This command creates the following categories if they don't already exist:
  - Compute: VMs, containers, general-purpose computing
  - HPC: High-performance computing, MPI clusters
  - GPU: GPU-accelerated instances for ML/DL
  - Storage: Object, block, and file storage
  - Network: VPNs, load balancers, firewalls
  - TEE: Trusted Execution Environments
  - AI/ML: Machine learning platforms

Categories are required before offerings can be created in Waldur.
Run this command after starting Waldur for the first time.`,
		Example: `  # Initialize categories using environment variables
  export WALDUR_URL=https://localhost/api/
  export WALDUR_TOKEN=your-api-token
  virtengine waldur init-categories

  # Initialize categories with explicit flags
  virtengine waldur init-categories --waldur-url https://localhost/api/ --waldur-token your-token

  # Output as JSON
  virtengine waldur init-categories --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get Waldur URL from flag or environment
			if waldurURL == "" {
				waldurURL = os.Getenv("WALDUR_URL")
			}
			if waldurURL == "" {
				return fmt.Errorf("waldur URL is required (set --waldur-url or WALDUR_URL)")
			}

			// Get Waldur token from flag or environment
			if waldurToken == "" {
				waldurToken = os.Getenv("WALDUR_TOKEN")
			}
			if waldurToken == "" {
				return fmt.Errorf("waldur token is required (set --waldur-token or WALDUR_TOKEN)")
			}

			// Create context with timeout
			ctx, cancel := context.WithTimeout(cmd.Context(), time.Duration(timeout)*time.Second)
			defer cancel()

			// Initialize categories
			result, err := provider_daemon.InitCategories(ctx, waldurURL, waldurToken)
			if err != nil {
				return fmt.Errorf("failed to initialize categories: %w", err)
			}

			// Output result
			switch output {
			case "json":
				return outputJSON(result)
			default:
				return outputInitResult(result)
			}
		},
	}

	cmd.Flags().StringVar(&waldurURL, flagWaldurURL, "", "Waldur API URL (or set WALDUR_URL)")
	cmd.Flags().StringVar(&waldurToken, flagWaldurToken, "", "Waldur API token (or set WALDUR_TOKEN)")
	cmd.Flags().StringVarP(&output, flagOutput, "o", "text", "Output format (text|json)")
	cmd.Flags().IntVar(&timeout, flagTimeout, 120, "Timeout in seconds")

	return cmd
}

func getListCategoriesCmd() *cobra.Command {
	var (
		waldurURL   string
		waldurToken string
		output      string
		timeout     int
	)

	cmd := &cobra.Command{
		Use:   "list-categories",
		Short: "List marketplace categories from Waldur",
		Long: `List all marketplace categories configured in Waldur.

This shows the categories that are available for creating offerings.
Use init-categories to create the default VirtEngine categories.`,
		Example: `  # List categories
  virtengine waldur list-categories

  # Output as JSON
  virtengine waldur list-categories --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get Waldur URL from flag or environment
			if waldurURL == "" {
				waldurURL = os.Getenv("WALDUR_URL")
			}
			if waldurURL == "" {
				return fmt.Errorf("waldur URL is required (set --waldur-url or WALDUR_URL)")
			}

			// Get Waldur token from flag or environment
			if waldurToken == "" {
				waldurToken = os.Getenv("WALDUR_TOKEN")
			}
			if waldurToken == "" {
				return fmt.Errorf("waldur token is required (set --waldur-token or WALDUR_TOKEN)")
			}

			// Create context with timeout
			ctx, cancel := context.WithTimeout(cmd.Context(), time.Duration(timeout)*time.Second)
			defer cancel()

			// Create Waldur client
			cfg := waldur.DefaultConfig()
			cfg.BaseURL = waldurURL
			cfg.Token = waldurToken

			client, err := waldur.NewClient(cfg)
			if err != nil {
				return fmt.Errorf("failed to create Waldur client: %w", err)
			}

			marketplace := waldur.NewMarketplaceClient(client)

			// List categories
			categories, err := marketplace.ListCategories(ctx, waldur.ListCategoriesParams{PageSize: 100})
			if err != nil {
				return fmt.Errorf("failed to list categories: %w", err)
			}

			// Output result
			switch output {
			case "json":
				return outputJSON(categories)
			default:
				return outputCategoriesTable(categories)
			}
		},
	}

	cmd.Flags().StringVar(&waldurURL, flagWaldurURL, "", "Waldur API URL (or set WALDUR_URL)")
	cmd.Flags().StringVar(&waldurToken, flagWaldurToken, "", "Waldur API token (or set WALDUR_TOKEN)")
	cmd.Flags().StringVarP(&output, flagOutput, "o", "text", "Output format (text|json)")
	cmd.Flags().IntVar(&timeout, flagTimeout, 60, "Timeout in seconds")

	return cmd
}

func outputInitResult(result *provider_daemon.InitCategoriesResult) error {
	fmt.Println("Category Initialization Result")
	fmt.Println("==============================")
	fmt.Println()

	if len(result.Created) > 0 {
		fmt.Printf("✅ Created (%d):\n", len(result.Created))
		for _, cat := range result.Created {
			uuid := result.Mappings[cat]
			fmt.Printf("   - %s [%s]\n", cat, uuid)
		}
		fmt.Println()
	}

	if len(result.Existing) > 0 {
		fmt.Printf("ℹ️  Already Existed (%d):\n", len(result.Existing))
		for _, cat := range result.Existing {
			uuid := result.Mappings[cat]
			fmt.Printf("   - %s [%s]\n", cat, uuid)
		}
		fmt.Println()
	}

	if len(result.Failed) > 0 {
		fmt.Printf("❌ Failed (%d):\n", len(result.Failed))
		for _, cat := range result.Failed {
			fmt.Printf("   - %s\n", cat)
		}
		fmt.Println()
	}

	total := len(result.Created) + len(result.Existing)
	fmt.Printf("Total: %d categories available\n", total)

	if len(result.Failed) > 0 {
		return fmt.Errorf("%d categories failed to initialize", len(result.Failed))
	}

	return nil
}

func outputCategoriesTable(categories []waldur.Category) error {
	if len(categories) == 0 {
		fmt.Println("No categories found.")
		fmt.Println("Run 'virtengine waldur init-categories' to create default categories.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TITLE\tUUID\tDESCRIPTION")
	fmt.Fprintln(w, "-----\t----\t-----------")

	for _, cat := range categories {
		desc := cat.Description
		if len(desc) > 50 {
			desc = desc[:47] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", cat.Title, cat.UUID, desc)
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
