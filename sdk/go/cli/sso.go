// Package cli provides CLI commands for the VirtEngine SDK.
//
// VE-4B: SSO/OIDC Admin CLI Commands
// This file provides CLI commands for managing SSO/OIDC issuer configuration.
package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// GetSSOCmd returns the SSO management commands.
func GetSSOCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sso",
		Short: "SSO/OIDC verification management commands",
		Long: `Commands for managing SSO/OIDC verification configuration, including
issuer allowlist management, policy configuration, and linkage operations.`,
	}

	cmd.AddCommand(
		GetSSOIssuerCmd(),
		GetSSOConfigCmd(),
		GetSSOLinkageCmd(),
	)

	return cmd
}

// GetSSOIssuerCmd returns the SSO issuer management commands.
func GetSSOIssuerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issuer",
		Short: "SSO issuer management commands",
		Long:  "Commands for managing the SSO issuer allowlist and policies.",
	}

	cmd.AddCommand(
		getSSOIssuerListCmd(),
		getSSOIssuerAddCmd(),
		getSSOIssuerRemoveCmd(),
		getSSOIssuerShowCmd(),
		getSSOIssuerUpdateCmd(),
	)

	return cmd
}

func getSSOIssuerListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all configured SSO issuers",
		Long:  "List all SSO/OIDC issuers in the allowlist with their policies.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement reading from config file or on-chain query
			fmt.Println("Configured SSO Issuers:")
			fmt.Println("========================")
			fmt.Println("")
			fmt.Println("Well-Known Issuers:")
			for providerType, issuer := range wellKnownIssuers {
				fmt.Printf("  %s: %s\n", providerType, issuer)
			}
			fmt.Println("")
			fmt.Println("Use 'sso issuer show <issuer>' for detailed policy information.")
			return nil
		},
	}
	return cmd
}

func getSSOIssuerAddCmd() *cobra.Command {
	var (
		clientID            string
		providerType        string
		scoreWeight         uint32
		enabled             bool
		requireEmailVerified bool
		allowedEmailDomains []string
		allowedTenants      []string
		policyFile          string
	)

	cmd := &cobra.Command{
		Use:   "add <issuer-url>",
		Short: "Add an SSO issuer to the allowlist",
		Long: `Add a new SSO/OIDC issuer to the allowlist with the specified policy.
The issuer URL should be the OIDC issuer identifier (e.g., https://accounts.google.com).

Example:
  virtengine sso issuer add https://accounts.google.com \
    --client-id="your-client-id" \
    --provider-type=google \
    --score-weight=250 \
    --enabled=true`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			issuerURL := args[0]

			// Validate provider type
			pt := veidtypes.SSOProviderType(providerType)
			if !veidtypes.IsValidSSOProviderType(pt) {
				return fmt.Errorf("invalid provider type: %s (valid: google, microsoft, github, oidc)", providerType)
			}

			policy := IssuerPolicyConfig{
				Issuer:               issuerURL,
				ClientID:             clientID,
				ProviderType:         string(pt),
				ScoreWeight:          scoreWeight,
				Enabled:              enabled,
				RequireEmailVerified: requireEmailVerified,
				AllowedEmailDomains:  allowedEmailDomains,
				AllowedTenants:       allowedTenants,
			}

			if policyFile != "" {
				data, err := os.ReadFile(policyFile)
				if err != nil {
					return fmt.Errorf("failed to read policy file: %w", err)
				}
				if err := json.Unmarshal(data, &policy); err != nil {
					return fmt.Errorf("failed to parse policy file: %w", err)
				}
			}

			// Validate
			if policy.ClientID == "" {
				return fmt.Errorf("client-id is required")
			}

			fmt.Printf("Adding SSO issuer:\n")
			fmt.Printf("  Issuer: %s\n", policy.Issuer)
			fmt.Printf("  Provider Type: %s\n", policy.ProviderType)
			fmt.Printf("  Client ID: %s\n", policy.ClientID)
			fmt.Printf("  Score Weight: %d\n", policy.ScoreWeight)
			fmt.Printf("  Enabled: %t\n", policy.Enabled)

			// TODO: Actually add to config/on-chain
			fmt.Println("\n⚠️  Note: This is a preview. Actual configuration update not yet implemented.")
			fmt.Println("To configure issuers, update the SSO service configuration file.")
			return nil
		},
	}

	cmd.Flags().StringVar(&clientID, "client-id", "", "OAuth client ID (required)")
	cmd.Flags().StringVar(&providerType, "provider-type", "oidc", "Provider type (google, microsoft, github, oidc)")
	cmd.Flags().Uint32Var(&scoreWeight, "score-weight", 150, "Score weight in basis points (0-10000)")
	cmd.Flags().BoolVar(&enabled, "enabled", true, "Enable this issuer")
	cmd.Flags().BoolVar(&requireEmailVerified, "require-email-verified", true, "Require email to be verified")
	cmd.Flags().StringSliceVar(&allowedEmailDomains, "allowed-email-domains", nil, "Allowed email domains")
	cmd.Flags().StringSliceVar(&allowedTenants, "allowed-tenants", nil, "Allowed tenant IDs")
	cmd.Flags().StringVar(&policyFile, "policy-file", "", "JSON policy file")

	_ = cmd.MarkFlagRequired("client-id")

	return cmd
}

func getSSOIssuerRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <issuer-url>",
		Short: "Remove an SSO issuer from the allowlist",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			issuerURL := args[0]
			fmt.Printf("Removing SSO issuer: %s\n", issuerURL)
			fmt.Println("\n⚠️  Note: This is a preview. Actual removal not yet implemented.")
			return nil
		},
	}
	return cmd
}

func getSSOIssuerShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <issuer-url>",
		Short: "Show SSO issuer policy details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			issuerURL := args[0]
			fmt.Printf("SSO Issuer Policy: %s\n", issuerURL)
			fmt.Println("================================")

			// Check if it's a well-known issuer
			for pt, url := range wellKnownIssuers {
				if url == issuerURL {
					fmt.Printf("  Provider Type: %s\n", pt)
					fmt.Printf("  Discovery URL: %s/.well-known/openid-configuration\n", issuerURL)
					weight := veidtypes.GetSSOScoringWeight(pt)
					fmt.Printf("  Default Score Weight: %d (%.2f%%)\n", weight, float64(weight)/100)
					break
				}
			}

			fmt.Println("\n⚠️  Note: Full policy details require configuration file or on-chain query.")
			return nil
		},
	}
	return cmd
}

func getSSOIssuerUpdateCmd() *cobra.Command {
	var (
		scoreWeight          *uint32
		enabled              *bool
		requireEmailVerified *bool
	)

	cmd := &cobra.Command{
		Use:   "update <issuer-url>",
		Short: "Update an SSO issuer policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			issuerURL := args[0]
			fmt.Printf("Updating SSO issuer policy: %s\n", issuerURL)

			if scoreWeight != nil {
				fmt.Printf("  Score Weight: %d\n", *scoreWeight)
			}
			if enabled != nil {
				fmt.Printf("  Enabled: %t\n", *enabled)
			}
			if requireEmailVerified != nil {
				fmt.Printf("  Require Email Verified: %t\n", *requireEmailVerified)
			}

			fmt.Println("\n⚠️  Note: This is a preview. Actual update not yet implemented.")
			return nil
		},
	}

	cmd.Flags().Uint32("score-weight", 0, "Score weight in basis points")
	cmd.Flags().Bool("enabled", true, "Enable/disable issuer")
	cmd.Flags().Bool("require-email-verified", true, "Require email verification")

	return cmd
}

// GetSSOConfigCmd returns the SSO configuration commands.
func GetSSOConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "SSO configuration commands",
		Long:  "Commands for viewing and updating SSO service configuration.",
	}

	cmd.AddCommand(
		getSSOConfigShowCmd(),
		getSSOConfigValidateCmd(),
	)

	return cmd
}

func getSSOConfigShowCmd() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show current SSO configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			config := DefaultSSOConfig()

			if outputFormat == "json" {
				data, err := json.MarshalIndent(config, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
			} else {
				fmt.Println("SSO Service Configuration")
				fmt.Println("=========================")
				fmt.Printf("  Enabled: %t\n", config.Enabled)
				fmt.Printf("  Challenge TTL: %d seconds\n", config.ChallengeTTLSeconds)
				fmt.Printf("  Attestation Validity: %d days\n", config.AttestationValidityDays)
				fmt.Printf("  Max Challenges Per Account: %d\n", config.MaxChallengesPerAccount)
				fmt.Printf("  Enable Rate Limiting: %t\n", config.EnableRateLimiting)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Output format (text, json)")

	return cmd
}

func getSSOConfigValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate <config-file>",
		Short: "Validate SSO configuration file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile := args[0]

			data, err := os.ReadFile(configFile)
			if err != nil {
				return fmt.Errorf("failed to read config file: %w", err)
			}

			var config SSOServiceConfig
			if err := json.Unmarshal(data, &config); err != nil {
				return fmt.Errorf("❌ Invalid JSON: %w", err)
			}

			// Validate
			if err := config.Validate(); err != nil {
				return fmt.Errorf("❌ Validation failed: %w", err)
			}

			fmt.Println("✅ Configuration is valid")
			fmt.Printf("   Issuers configured: %d\n", len(config.IssuerPolicies))
			return nil
		},
	}
	return cmd
}

// GetSSOLinkageCmd returns the SSO linkage management commands.
func GetSSOLinkageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "linkage",
		Short: "SSO linkage management commands",
		Long:  "Commands for viewing and managing SSO account linkages.",
	}

	cmd.AddCommand(
		getSSOLinkageListCmd(),
		getSSOLinkageShowCmd(),
		getSSOLinkageRevokeCmd(),
	)

	return cmd
}

func getSSOLinkageListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <account-address>",
		Short: "List SSO linkages for an account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			accountAddress := args[0]
			fmt.Printf("SSO Linkages for: %s\n", accountAddress)
			fmt.Println("===================================")
			fmt.Println("\n⚠️  Note: Query not yet implemented. Use gRPC query for actual linkages.")
			return nil
		},
	}
	return cmd
}

func getSSOLinkageShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <linkage-id>",
		Short: "Show SSO linkage details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			linkageID := args[0]
			fmt.Printf("SSO Linkage: %s\n", linkageID)
			fmt.Println("================")
			fmt.Println("\n⚠️  Note: Query not yet implemented. Use gRPC query for actual linkage details.")
			return nil
		},
	}
	return cmd
}

func getSSOLinkageRevokeCmd() *cobra.Command {
	var reason string

	cmd := &cobra.Command{
		Use:   "revoke <linkage-id>",
		Short: "Revoke an SSO linkage",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			linkageID := args[0]
			fmt.Printf("Revoking SSO linkage: %s\n", linkageID)
			if reason != "" {
				fmt.Printf("Reason: %s\n", reason)
			}
			fmt.Println("\n⚠️  Note: Revocation requires a signed transaction. Use the tx command instead.")
			return nil
		},
	}

	cmd.Flags().StringVar(&reason, "reason", "", "Revocation reason")

	return cmd
}

// ============================================================================
// Helper Types
// ============================================================================

// IssuerPolicyConfig is the CLI representation of an issuer policy.
type IssuerPolicyConfig struct {
	Issuer               string   `json:"issuer"`
	ClientID             string   `json:"client_id"`
	ProviderType         string   `json:"provider_type"`
	ScoreWeight          uint32   `json:"score_weight"`
	Enabled              bool     `json:"enabled"`
	RequireEmailVerified bool     `json:"require_email_verified"`
	AllowedEmailDomains  []string `json:"allowed_email_domains,omitempty"`
	AllowedTenants       []string `json:"allowed_tenants,omitempty"`
}

// SSOServiceConfig is the CLI representation of SSO service config.
type SSOServiceConfig struct {
	Enabled                bool                          `json:"enabled"`
	ChallengeTTLSeconds    int64                         `json:"challenge_ttl_seconds"`
	AttestationValidityDays int                          `json:"attestation_validity_days"`
	MaxChallengesPerAccount int                          `json:"max_challenges_per_account"`
	EnableRateLimiting     bool                          `json:"enable_rate_limiting"`
	IssuerPolicies         map[string]IssuerPolicyConfig `json:"issuer_policies"`
}

// Validate validates the SSO service config.
func (c *SSOServiceConfig) Validate() error {
	if c.ChallengeTTLSeconds <= 0 {
		return fmt.Errorf("challenge_ttl_seconds must be positive")
	}
	if c.AttestationValidityDays <= 0 {
		return fmt.Errorf("attestation_validity_days must be positive")
	}
	return nil
}

// DefaultSSOConfig returns default SSO configuration.
func DefaultSSOConfig() *SSOServiceConfig {
	return &SSOServiceConfig{
		Enabled:                true,
		ChallengeTTLSeconds:    600, // 10 minutes
		AttestationValidityDays: 365,
		MaxChallengesPerAccount: 3,
		EnableRateLimiting:     true,
		IssuerPolicies:         make(map[string]IssuerPolicyConfig),
	}
}

// Well-known OIDC issuers
var wellKnownIssuers = map[veidtypes.SSOProviderType]string{
	veidtypes.SSOProviderGoogle:    "https://accounts.google.com",
	veidtypes.SSOProviderMicrosoft: "https://login.microsoftonline.com/common/v2.0",
	veidtypes.SSOProviderGitHub:    "https://token.actions.githubusercontent.com",
}
