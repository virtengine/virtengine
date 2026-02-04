package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestParseUintArg(t *testing.T) {
	value, err := parseUintArg("42", "nodes")
	require.NoError(t, err)
	require.Equal(t, uint64(42), value)

	_, err = parseUintArg("not-a-number", "nodes")
	require.Error(t, err)
}

func TestReadActiveFlag(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool(flagActive, false, "")
	cmd.Flags().Bool(flagInactive, false, "")

	active, err := readActiveFlag(cmd)
	require.NoError(t, err)
	require.False(t, active)

	require.NoError(t, cmd.Flags().Set(flagActive, "true"))
	active, err = readActiveFlag(cmd)
	require.NoError(t, err)
	require.True(t, active)
}

func TestReadJobScriptFromFlags(t *testing.T) {
	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "job.sh")
	require.NoError(t, os.WriteFile(scriptPath, []byte("echo hi\n"), 0o600))

	cmd := &cobra.Command{}
	cmd.Flags().String(flagJobScript, "", "")
	cmd.Flags().String(flagJobScriptFile, "", "")

	require.NoError(t, cmd.Flags().Set(flagJobScriptFile, scriptPath))
	script, err := readJobScript(cmd)
	require.NoError(t, err)
	require.Equal(t, "echo hi\n", script)
}

func TestReadJobSubmitSpec(t *testing.T) {
	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "job.sh")
	require.NoError(t, os.WriteFile(scriptPath, []byte("echo template\n"), 0o600))

	specPath := filepath.Join(dir, "job.yaml")
	spec := strings.Join([]string{
		"offering_id: OFF-1",
		"requested_nodes: 2",
		"requested_gpus: 4",
		"max_duration: 3600",
		"max_budget: 1000uve",
		"job_script_file: job.sh",
	}, "\n")
	require.NoError(t, os.WriteFile(specPath, []byte(spec), 0o600))

	parsed, err := readJobSubmitSpec(specPath)
	require.NoError(t, err)
	require.Equal(t, "OFF-1", parsed.OfferingID)
	require.Equal(t, uint64(2), parsed.RequestedNodes)
	require.Equal(t, uint64(4), parsed.RequestedGpus)
	require.Equal(t, uint64(3600), parsed.MaxDuration)
	require.Equal(t, "1000uve", parsed.MaxBudget)
	require.Equal(t, "echo template\n", parsed.JobScript)
}

func TestReadTemplateSubmitSpec(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "template.json")
	spec := `{
  "offering_id": "OFF-2",
  "requested_nodes": 4,
  "requested_gpus": 2,
  "max_duration": 7200,
  "max_budget": "2500uve"
}`
	require.NoError(t, os.WriteFile(specPath, []byte(spec), 0o600))

	parsed, err := readTemplateSubmitSpec(specPath)
	require.NoError(t, err)
	require.Equal(t, "OFF-2", parsed.OfferingID)
	require.Equal(t, uint64(4), parsed.RequestedNodes)
	require.Equal(t, uint64(2), parsed.RequestedGpus)
	require.Equal(t, uint64(7200), parsed.MaxDuration)
	require.Equal(t, "2500uve", parsed.MaxBudget)
}

func TestReadProviderRegistrationConfigFromFlags(t *testing.T) {
	cmd := &cobra.Command{}
	addProviderRegistrationFlags(cmd)

	require.NoError(t, cmd.Flags().Set(flagName, "A100-east"))
	require.NoError(t, cmd.Flags().Set(flagClusterType, "slurm"))
	require.NoError(t, cmd.Flags().Set(flagRegion, "us-east-1"))
	require.NoError(t, cmd.Flags().Set(flagEndpoint, "https://hpc.example.com"))
	require.NoError(t, cmd.Flags().Set(flagTotalNodes, "64"))
	require.NoError(t, cmd.Flags().Set(flagTotalGpus, "512"))

	cfg, err := readProviderRegistrationConfig(cmd, nil)
	require.NoError(t, err)
	require.Equal(t, "A100-east", cfg.Name)
	require.Equal(t, "slurm", cfg.ClusterType)
	require.Equal(t, "us-east-1", cfg.Region)
	require.Equal(t, "https://hpc.example.com", cfg.Endpoint)
	require.Equal(t, uint64(64), cfg.TotalNodes)
	require.Equal(t, uint64(512), cfg.TotalGpus)
}

func TestReadQueueConfigFromFlags(t *testing.T) {
	cmd := &cobra.Command{}
	addQueueFlags(cmd)

	require.NoError(t, cmd.Flags().Set(flagClusterID, "HPC-1"))
	require.NoError(t, cmd.Flags().Set(flagName, "A100 on-demand"))
	require.NoError(t, cmd.Flags().Set(flagResource, "gpu"))
	require.NoError(t, cmd.Flags().Set(flagPricePerHour, "12.5uve"))
	require.NoError(t, cmd.Flags().Set(flagMinDuration, "3600"))
	require.NoError(t, cmd.Flags().Set(flagMaxDuration, "86400"))

	cfg, err := readQueueConfig(cmd, nil)
	require.NoError(t, err)
	require.Equal(t, "HPC-1", cfg.ClusterID)
	require.Equal(t, "A100 on-demand", cfg.Name)
	require.Equal(t, "gpu", cfg.ResourceType)
	require.Equal(t, "12.5uve", cfg.PricePerHour)
	require.Equal(t, uint64(3600), cfg.MinDuration)
	require.Equal(t, uint64(86400), cfg.MaxDuration)
}
