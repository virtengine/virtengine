package cli

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	v1 "github.com/virtengine/virtengine/sdk/go/node/enclave/v1"
	"github.com/virtengine/virtengine/pkg/security"
	"github.com/virtengine/virtengine/x/enclave/types"
)

type addMeasurementProposalJSON struct {
	Title           string `json:"title"`
	Description     string `json:"description"`
	MeasurementHash string `json:"measurement_hash"`
	TEEType         string `json:"tee_type"`
	MinISVSVN       uint16 `json:"min_isv_svn"`
	ExpiryBlocks    int64  `json:"expiry_blocks"`
	Deposit         string `json:"deposit"`
}

type revokeMeasurementProposalJSON struct {
	Title           string `json:"title"`
	Description     string `json:"description"`
	MeasurementHash string `json:"measurement_hash"`
	Reason          string `json:"reason"`
	Deposit         string `json:"deposit"`
}

// NewCmdSubmitAddMeasurementProposal submits an add-measurement governance proposal.
func NewCmdSubmitAddMeasurementProposal() *cobra.Command {
	return &cobra.Command{
		Use:   "add-measurement [proposal-file]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit an add measurement proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit an add measurement proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.

Example:
$ %s tx gov submit-legacy-proposal add-measurement <path/to/proposal.json> --from=<key_or_address>

Where proposal.json contains:

{
  "title": "Add SGX measurement v1",
  "description": "Allowlist SGX enclave measurement for production",
  "measurement_hash": "8b0f... (hex string)",
  "tee_type": "SGX",
  "min_isv_svn": 1,
  "expiry_blocks": 0,
  "deposit": "1000uve"
}
`, version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposal, err := parseAddMeasurementProposalJSON(args[0])
			if err != nil {
				return err
			}

			measurementHash, err := decodeMeasurementHash(proposal.MeasurementHash)
			if err != nil {
				return err
			}

			teeType, err := parseTEEType(proposal.TEEType)
			if err != nil {
				return err
			}
			content := types.NewAddMeasurementProposal(
				proposal.Title,
				proposal.Description,
				measurementHash,
				teeType,
				proposal.MinISVSVN,
				proposal.ExpiryBlocks,
			)

			deposit, err := sdk.ParseCoinsNormalized(proposal.Deposit)
			if err != nil {
				return err
			}

			msg, err := govv1beta1.NewMsgSubmitProposal(content, deposit, clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
}

// NewCmdSubmitRevokeMeasurementProposal submits a revoke-measurement governance proposal.
func NewCmdSubmitRevokeMeasurementProposal() *cobra.Command {
	return &cobra.Command{
		Use:   "revoke-measurement [proposal-file]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a revoke measurement proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a revoke measurement proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.

Example:
$ %s tx gov submit-legacy-proposal revoke-measurement <path/to/proposal.json> --from=<key_or_address>

Where proposal.json contains:

{
  "title": "Revoke SGX measurement v1",
  "description": "Revoke compromised measurement",
  "measurement_hash": "8b0f... (hex string)",
  "reason": "CVE-2026-0001",
  "deposit": "1000uve"
}
`, version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposal, err := parseRevokeMeasurementProposalJSON(args[0])
			if err != nil {
				return err
			}

			measurementHash, err := decodeMeasurementHash(proposal.MeasurementHash)
			if err != nil {
				return err
			}

			content := types.NewRevokeMeasurementProposal(
				proposal.Title,
				proposal.Description,
				measurementHash,
				proposal.Reason,
			)

			deposit, err := sdk.ParseCoinsNormalized(proposal.Deposit)
			if err != nil {
				return err
			}

			msg, err := govv1beta1.NewMsgSubmitProposal(content, deposit, clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
}

func parseAddMeasurementProposalJSON(path string) (addMeasurementProposalJSON, error) {
	var proposal addMeasurementProposalJSON

	// Validate path and read with extension check
	raw, err := security.SafeReadFileWithExtension(path, ".json")
	if err != nil {
		return proposal, err
	}

	if err := json.Unmarshal(raw, &proposal); err != nil {
		return proposal, err
	}

	if strings.TrimSpace(proposal.Title) == "" || strings.TrimSpace(proposal.Description) == "" {
		return proposal, fmt.Errorf("title and description are required")
	}
	if strings.TrimSpace(proposal.MeasurementHash) == "" {
		return proposal, fmt.Errorf("measurement_hash is required")
	}
	if strings.TrimSpace(proposal.TEEType) == "" {
		return proposal, fmt.Errorf("tee_type is required")
	}
	if strings.TrimSpace(proposal.Deposit) == "" {
		return proposal, fmt.Errorf("deposit is required")
	}

	return proposal, nil
}

func parseRevokeMeasurementProposalJSON(path string) (revokeMeasurementProposalJSON, error) {
	var proposal revokeMeasurementProposalJSON

	// Validate path and read with extension check
	raw, err := security.SafeReadFileWithExtension(path, ".json")
	if err != nil {
		return proposal, err
	}

	if err := json.Unmarshal(raw, &proposal); err != nil {
		return proposal, err
	}

	if strings.TrimSpace(proposal.Title) == "" || strings.TrimSpace(proposal.Description) == "" {
		return proposal, fmt.Errorf("title and description are required")
	}
	if strings.TrimSpace(proposal.MeasurementHash) == "" {
		return proposal, fmt.Errorf("measurement_hash is required")
	}
	if strings.TrimSpace(proposal.Reason) == "" {
		return proposal, fmt.Errorf("reason is required")
	}
	if strings.TrimSpace(proposal.Deposit) == "" {
		return proposal, fmt.Errorf("deposit is required")
	}

	return proposal, nil
}

func decodeMeasurementHash(input string) ([]byte, error) {
	raw := strings.TrimSpace(input)
	if raw == "" {
		return nil, fmt.Errorf("measurement hash cannot be empty")
	}

	trimmed := strings.TrimPrefix(raw, "0x")
	if len(trimmed) == 64 {
		bz, err := hex.DecodeString(trimmed)
		if err != nil {
			return nil, fmt.Errorf("invalid hex measurement hash: %w", err)
		}
		if len(bz) != 32 {
			return nil, fmt.Errorf("measurement hash must be 32 bytes, got %d", len(bz))
		}
		return bz, nil
	}

	bz, err := base64.StdEncoding.DecodeString(raw)
	if err == nil {
		if len(bz) != 32 {
			return nil, fmt.Errorf("measurement hash must be 32 bytes, got %d", len(bz))
		}
		return bz, nil
	}

	bz, err = base64.RawStdEncoding.DecodeString(raw)
	if err == nil {
		if len(bz) != 32 {
			return nil, fmt.Errorf("measurement hash must be 32 bytes, got %d", len(bz))
		}
		return bz, nil
	}

	return nil, fmt.Errorf("measurement hash must be 32-byte hex or base64")
}

// parseTEEType parses a TEE type string into the corresponding enum value.
// Accepts formats like "SGX", "SEV_SNP", "NITRO", "TRUSTZONE" (case-insensitive).
func parseTEEType(s string) (types.TEEType, error) {
	normalized := strings.ToUpper(strings.TrimSpace(s))
	if normalized == "" {
		return types.TEETypeUnspecified, fmt.Errorf("tee_type cannot be empty")
	}

	// Try direct lookup with TEE_TYPE_ prefix
	if val, ok := v1.TEEType_value["TEE_TYPE_"+normalized]; ok {
		return types.TEEType(val), nil
	}

	// Try exact match (e.g., "TEE_TYPE_SGX")
	if val, ok := v1.TEEType_value[normalized]; ok {
		return types.TEEType(val), nil
	}

	return types.TEETypeUnspecified, fmt.Errorf("invalid tee_type: %s, valid values are: SGX, SEV_SNP, NITRO, TRUSTZONE", s)
}
