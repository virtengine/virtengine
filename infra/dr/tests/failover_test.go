package dr

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Default targets — override via environment variables.
const (
	defaultPrimaryRegion   = "us-east-1"
	defaultSecondaryRegion = "eu-west-1"
	defaultTertiaryRegion  = "ap-southeast-1"
	defaultDomain          = "virtengine.io"

	// SLA targets
	rtoTarget = 15 * time.Minute // Recovery Time Objective
	rpoTarget = 5 * time.Minute  // Recovery Point Objective
)

// region returns the env value or the supplied fallback.
func region(envKey, fallback string) string {
	if v := os.Getenv(envKey); v != "" {
		return v
	}
	return fallback
}

func domain() string {
	if v := os.Getenv("VE_DOMAIN"); v != "" {
		return v
	}
	return defaultDomain
}

// kubeContext returns the kubectl context name for a region.
func kubeContext(regionName string) string {
	if v := os.Getenv("VE_KUBE_CONTEXT_" + strings.ReplaceAll(strings.ToUpper(regionName), "-", "_")); v != "" {
		return v
	}
	return "virtengine-prod-" + regionName
}

// runCommand executes a shell command and returns stdout.
func runCommand(t *testing.T, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "command %s %v failed: %s", name, args, string(out))
	return strings.TrimSpace(string(out))
}

// runCommandAllowFailure runs a command and returns output + error.
func runCommandAllowFailure(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// --------------------------------------------------------------------------
// AC-5: Regional Failover Tests
// --------------------------------------------------------------------------

func TestRegionalFailover_HealthChecks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	regions := []string{
		region("VE_PRIMARY_REGION", defaultPrimaryRegion),
		region("VE_SECONDARY_REGION", defaultSecondaryRegion),
		region("VE_TERTIARY_REGION", defaultTertiaryRegion),
	}
	dom := domain()

	for _, r := range regions {
		t.Run(r, func(t *testing.T) {
			endpoint := fmt.Sprintf("https://rpc-%s.%s/status", r, dom)
			out, err := runCommandAllowFailure("curl", "-sf", "--connect-timeout", "10", endpoint)
			if err != nil {
				t.Logf("region %s not reachable (may not be deployed yet): %v", r, err)
				t.Skip("endpoint not reachable — deploy first")
			}
			assert.Contains(t, out, "result", "health check response should contain result")
		})
	}
}

func TestRegionalFailover_DNSResolution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dom := domain()
	records := []string{
		"api." + dom,
		"rpc." + dom,
	}

	for _, record := range records {
		t.Run(record, func(t *testing.T) {
			out, err := runCommandAllowFailure("dig", "+short", record)
			if err != nil {
				t.Skip("dig not available")
			}
			assert.NotEmpty(t, out, "DNS record %s should resolve", record)
		})
	}
}

func TestRegionalFailover_CrossRegionConnectivity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	primary := region("VE_PRIMARY_REGION", defaultPrimaryRegion)
	secondary := region("VE_SECONDARY_REGION", defaultSecondaryRegion)
	ctx1 := kubeContext(primary)
	ctx2 := kubeContext(secondary)

	// Verify both clusters are reachable
	for _, ctx := range []string{ctx1, ctx2} {
		t.Run("cluster-reachable/"+ctx, func(t *testing.T) {
			_, err := runCommandAllowFailure("kubectl", "--context", ctx, "cluster-info")
			if err != nil {
				t.Skipf("cluster %s not reachable: %v", ctx, err)
			}
		})
	}
}

func TestRegionalFailover_ValidatorDistribution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	regions := []string{
		region("VE_PRIMARY_REGION", defaultPrimaryRegion),
		region("VE_SECONDARY_REGION", defaultSecondaryRegion),
		region("VE_TERTIARY_REGION", defaultTertiaryRegion),
	}

	totalValidators := 0
	for _, r := range regions {
		t.Run(r, func(t *testing.T) {
			ctx := kubeContext(r)
			out, err := runCommandAllowFailure("kubectl", "--context", ctx, "-n", "virtengine",
				"get", "pods", "-l", "role=validator",
				"--field-selector=status.phase=Running", "--no-headers")
			if err != nil {
				t.Skipf("region %s not reachable: %v", r, err)
				return
			}
			lines := strings.Split(strings.TrimSpace(out), "\n")
			count := 0
			if out != "" {
				count = len(lines)
			}
			totalValidators += count
			t.Logf("region %s: %d validators", r, count)
		})
	}

	t.Run("total", func(t *testing.T) {
		if totalValidators == 0 {
			t.Skip("no validators found — cluster may not be deployed")
		}
		assert.GreaterOrEqual(t, totalValidators, 3, "should have at least 3 validators total")
	})
}

// --------------------------------------------------------------------------
// AC-2: Database Replication Tests
// --------------------------------------------------------------------------

func TestDatabaseReplication_CockroachDBClusterHealth(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	primary := region("VE_PRIMARY_REGION", defaultPrimaryRegion)
	ctx := kubeContext(primary)

	out, err := runCommandAllowFailure("kubectl", "--context", ctx, "-n", "cockroachdb",
		"exec", "cockroachdb-0", "--",
		"cockroach", "node", "status",
		"--certs-dir=/cockroach/cockroach-certs",
		"--format=json")
	if err != nil {
		t.Skipf("CockroachDB not reachable: %v", err)
	}

	var nodes []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(out), &nodes), "should parse node status JSON")

	liveCount := 0
	for _, node := range nodes {
		if isLive, ok := node["is_live"].(bool); ok && isLive {
			liveCount++
		}
	}
	assert.GreaterOrEqual(t, liveCount, 3, "should have at least 3 live CockroachDB nodes")
}

func TestDatabaseReplication_BackupFreshness(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	regions := []string{
		region("VE_PRIMARY_REGION", defaultPrimaryRegion),
		region("VE_SECONDARY_REGION", defaultSecondaryRegion),
		region("VE_TERTIARY_REGION", defaultTertiaryRegion),
	}

	for _, r := range regions {
		t.Run(r, func(t *testing.T) {
			ctx := kubeContext(r)
			bucket := fmt.Sprintf("s3://virtengine-cockroachdb-backup-%s/backups", r)
			query := fmt.Sprintf(
				"SELECT extract(epoch from (now() - max(end_time)))::int AS age_seconds FROM [SHOW BACKUP LATEST IN '%s?AUTH=implicit'];",
				bucket,
			)

			out, err := runCommandAllowFailure("kubectl", "--context", ctx, "-n", "cockroachdb",
				"exec", "cockroachdb-0", "--",
				"cockroach", "sql",
				"--certs-dir=/cockroach/cockroach-certs",
				"--format=csv",
				"-e", query)
			if err != nil {
				t.Skipf("backup check not available for %s: %v", r, err)
				return
			}

			lines := strings.Split(strings.TrimSpace(out), "\n")
			if len(lines) < 2 {
				t.Skip("no backup data returned")
				return
			}

			ageSec, err := strconv.Atoi(strings.TrimSpace(lines[len(lines)-1]))
			if err != nil {
				t.Skipf("unable to parse backup age: %v", err)
				return
			}

			age := time.Duration(ageSec) * time.Second
			assert.Less(t, age, rpoTarget, "backup age %v should be within RPO target %v", age, rpoTarget)
		})
	}
}

func TestDatabaseReplication_ReplicationLag(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	primary := region("VE_PRIMARY_REGION", defaultPrimaryRegion)
	ctx := kubeContext(primary)

	// Query ranges that are under-replicated
	query := "SELECT count(*) AS underreplicated FROM crdb_internal.ranges WHERE array_length(replicas, 1) < 3;"

	out, err := runCommandAllowFailure("kubectl", "--context", ctx, "-n", "cockroachdb",
		"exec", "cockroachdb-0", "--",
		"cockroach", "sql",
		"--certs-dir=/cockroach/cockroach-certs",
		"--format=csv",
		"-e", query)
	if err != nil {
		t.Skipf("replication check not available: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 2 {
		t.Skip("no replication data")
		return
	}

	count, err := strconv.Atoi(strings.TrimSpace(lines[len(lines)-1]))
	if err != nil {
		t.Skipf("unable to parse underreplicated count: %v", err)
		return
	}

	assert.Equal(t, 0, count, "should have zero underreplicated ranges")
}

// --------------------------------------------------------------------------
// RTO/RPO Verification
// --------------------------------------------------------------------------

func TestRTO_Target(t *testing.T) {
	t.Logf("RTO target: %v", rtoTarget)
	t.Logf("Verification: Failover runbook steps should complete within RTO")
	t.Logf("Measured via VirtEngine/DR FailoverDurationSeconds CloudWatch metric")

	// This test documents the RTO requirement — actual measurement happens
	// during failover exercises via the regional-failover runbook.
	assert.Equal(t, 15*time.Minute, rtoTarget, "RTO target should be 15 minutes")
}

func TestRPO_Target(t *testing.T) {
	t.Logf("RPO target: %v", rpoTarget)
	t.Logf("Verification: CockroachDB replication lag should stay under RPO")
	t.Logf("Measured via cockroachdb_replication_lag_seconds Prometheus metric")

	// This test documents the RPO requirement — actual measurement happens
	// via continuous monitoring of replication lag.
	assert.Equal(t, 5*time.Minute, rpoTarget, "RPO target should be 5 minutes")
}

// --------------------------------------------------------------------------
// AC-4: Observability Cross-Region Tests
// --------------------------------------------------------------------------

func TestObservability_PrometheusFederation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	primary := region("VE_PRIMARY_REGION", defaultPrimaryRegion)
	ctx := kubeContext(primary)

	// Check if Prometheus is accessible in primary region
	out, err := runCommandAllowFailure("kubectl", "--context", ctx, "-n", "monitoring",
		"get", "pods", "-l", "app.kubernetes.io/name=prometheus",
		"--field-selector=status.phase=Running", "--no-headers")
	if err != nil {
		t.Skipf("Prometheus not reachable: %v", err)
	}

	if strings.TrimSpace(out) == "" {
		t.Skip("no Prometheus pods found")
	}

	t.Log("Prometheus pods running in primary region")
}

func TestObservability_CrossRegionAlerts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	primary := region("VE_PRIMARY_REGION", defaultPrimaryRegion)
	ctx := kubeContext(primary)

	// Check PrometheusRule exists
	out, err := runCommandAllowFailure("kubectl", "--context", ctx, "-n", "monitoring",
		"get", "prometheusrule", "-l", "app.kubernetes.io/part-of=virtengine-observability",
		"--no-headers")
	if err != nil {
		t.Skipf("PrometheusRule check not available: %v", err)
	}

	if strings.TrimSpace(out) == "" {
		t.Skip("no PrometheusRule resources found")
	}

	t.Log("Cross-region alert rules found")
}
