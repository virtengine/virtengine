package cli

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/virtengine/virtengine/x/hpc/types"
)

const (
	flagEndpoint        = "endpoint"
	flagTotalNodes      = "total-nodes"
	flagTotalGpus       = "total-gpus"
	flagActive          = "active"
	flagInactive        = "inactive"
	flagPricePerHour    = "price-per-hour"
	flagJobScript       = "job-script"
	flagJobScriptFile   = "job-script-file"
	flagReason          = "reason"
	flagProgressPercent = "progress-percent"
	flagOutputLocation  = "output-location"
	flagErrorMessage    = "error-message"
	flagGpuModel        = "gpu-model"
	flagGpuMemoryGb     = "gpu-memory-gb"
	flagCpuModel        = "cpu-model"
	flagMemoryGb        = "memory-gb"
	flagEvidence        = "evidence"
	flagRefundAmount    = "refund-amount"
	flagAuthority       = "authority"
	flagConfig          = "config"
)

// GetTxCmd returns the root tx command for the HPC module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "HPC transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		NewCmdRegisterProvider(),
		NewCmdUpdateProvider(),
		NewCmdSetPricing(),
		NewCmdCreateQueue(),
		NewCmdUpdateQueue(),
		NewCmdUpdateParams(),
		NewCmdAddTemplate(),
		NewCmdRegisterCluster(),
		NewCmdUpdateCluster(),
		NewCmdDeregisterCluster(),
		NewCmdCreateOffering(),
		NewCmdUpdateOffering(),
		NewCmdSubmitJob(),
		NewCmdCancelJob(),
		NewCmdExtendJob(),
		NewCmdSubmitFromTemplate(),
		NewCmdReportJobStatus(),
		NewCmdUpdateNodeMetadata(),
		NewCmdFlagDispute(),
		NewCmdResolveDispute(),
	)

	return cmd
}

func parseUintArg(arg, label string) (uint64, error) {
	value, err := strconv.ParseUint(arg, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", label, err)
	}
	return value, nil
}

func readJobScript(cmd *cobra.Command) (string, error) {
	script, err := cmd.Flags().GetString(flagJobScript)
	if err != nil {
		return "", err
	}
	scriptFile, err := cmd.Flags().GetString(flagJobScriptFile)
	if err != nil {
		return "", err
	}
	if scriptFile != "" {
		data, err := os.ReadFile(scriptFile)
		if err != nil {
			return "", fmt.Errorf("read job script file: %w", err)
		}
		return string(data), nil
	}
	if strings.TrimSpace(script) == "" {
		return "", fmt.Errorf("set --%s or --%s", flagJobScript, flagJobScriptFile)
	}
	return script, nil
}

func readActiveFlag(cmd *cobra.Command) (bool, error) {
	active, err := cmd.Flags().GetBool(flagActive)
	if err != nil {
		return false, err
	}
	inactive, err := cmd.Flags().GetBool(flagInactive)
	if err != nil {
		return false, err
	}
	if active && inactive {
		return false, fmt.Errorf("only one of --%s or --%s may be set", flagActive, flagInactive)
	}
	if inactive {
		return false, nil
	}
	return active, nil
}

func uint64ToInt32(value uint64, label string) (int32, error) {
	if value > math.MaxInt32 {
		return 0, fmt.Errorf("%s exceeds int32", label)
	}
	return int32(value), nil
}

func readConfigFlag(cmd *cobra.Command) (string, error) {
	path, err := cmd.Flags().GetString(flagConfig)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(path) == "" {
		return "", nil
	}
	return path, nil
}

func unmarshalConfigFile(path string, out interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("parse json config: %w", err)
		}
		return nil
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, out); err != nil {
			return fmt.Errorf("parse yaml config: %w", err)
		}
		return nil
	default:
		if err := json.Unmarshal(data, out); err == nil {
			return nil
		}
		if err := yaml.Unmarshal(data, out); err == nil {
			return nil
		}
		return fmt.Errorf("unsupported config format for %s", path)
	}
}
