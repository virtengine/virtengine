// Package hpc contains integration tests for HPC SLURM deployment.
//
//go:build e2e.integration

package hpc

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/pkg/provider_daemon/slurm_k8s"
)

// TestSLURMDeploymentKind tests SLURM deployment on a kind cluster.
// Prerequisites:
// - kind installed and available in PATH
// - kubectl configured for the kind cluster
// - Helm 3 installed
//
// Run with: go test -tags="e2e.integration" -v ./tests/integration/hpc/...
func TestSLURMDeploymentKind(t *testing.T) {
	// Check prerequisites
	if !checkPrerequisites(t) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	clusterName := "slurm-test-" + randomSuffix()
	namespace := "slurm-test"

	// Create kind cluster
	t.Log("Creating kind cluster...")
	if err := createKindCluster(ctx, clusterName); err != nil {
		t.Fatalf("failed to create kind cluster: %v", err)
	}
	defer deleteKindCluster(clusterName)

	// Create namespace
	t.Log("Creating namespace...")
	if err := runKubectl(ctx, "create", "namespace", namespace); err != nil {
		t.Fatalf("failed to create namespace: %v", err)
	}
	for _, secretArgs := range [][]string{
		{"create", "secret", "generic", "slurm-test-munge", "--namespace", namespace, "--from-literal=munge.key=integration-test-munge-key"},
		{"create", "secret", "generic", "slurm-test-database", "--namespace", namespace, "--from-literal=password=integration-test-database-password"},
		{"create", "secret", "generic", "slurm-test-mariadb", "--namespace", namespace, "--from-literal=root-password=integration-test-root-password"},
		{"create", "secret", "generic", "slurm-test-node-agent-tls", "--namespace", namespace, "--from-literal=ca.crt=integration-test-ca", "--from-literal=tls.crt=integration-test-cert", "--from-literal=tls.key=integration-test-key"},
	} {
		if err := runKubectl(ctx, secretArgs...); err != nil {
			t.Fatalf("failed to create test secret: %v", err)
		}
	}

	// Deploy SLURM cluster using Helm
	t.Log("Deploying SLURM cluster...")
	chartPath := "../../../deploy/slurm/slurm-cluster"
	releaseName := "slurm-test"

	helmArgs := []string{
		"install", releaseName, chartPath,
		"--namespace", namespace,
		"--set", "cluster.id=test-cluster",
		"--set", "cluster.name=Test SLURM Cluster",
		"--values", chartPath + "/tests/stable-secrets-values.yaml",
		"--set", "compute.replicas=2",
		"--set", "controller.persistence.size=1Gi",
		"--set", "database.persistence.size=1Gi",
		"--set", "mariadb.persistence.size=1Gi",
		"--wait",
		"--timeout", "10m",
	}

	if err := runHelm(ctx, helmArgs...); err != nil {
		// Get pod status for debugging
		runKubectl(ctx, "get", "pods", "-n", namespace)
		runKubectl(ctx, "describe", "pods", "-n", namespace)
		t.Fatalf("failed to deploy SLURM cluster: %v", err)
	}

	// Verify deployment
	t.Log("Verifying deployment...")
	if err := verifyDeployment(ctx, namespace, releaseName); err != nil {
		t.Fatalf("deployment verification failed: %v", err)
	}

	// Test SLURM functionality
	t.Log("Testing SLURM functionality...")
	if err := testSLURMFunctionality(ctx, namespace, releaseName); err != nil {
		t.Fatalf("SLURM functionality test failed: %v", err)
	}

	// Test scaling
	t.Log("Testing scaling...")
	if err := testScaling(ctx, namespace, releaseName, chartPath); err != nil {
		t.Fatalf("scaling test failed: %v", err)
	}

	// Cleanup
	t.Log("Cleaning up...")
	if err := runHelm(ctx, "uninstall", releaseName, "--namespace", namespace); err != nil {
		t.Errorf("failed to uninstall Helm release: %v", err)
	}

	t.Log("All tests passed!")
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}

func readRepoFile(t *testing.T, parts ...string) string {
	t.Helper()
	path := filepath.Join(append([]string{repoRoot(t)}, parts...)...)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func verifyOfflineSLURMContracts(t *testing.T) {
	t.Helper()
	chart := readRepoFile(t, "deploy", "slurm", "slurm-cluster", "templates", "compute-nodepools-statefulset.yaml")
	helpers := readRepoFile(t, "deploy", "slurm", "slurm-cluster", "templates", "_helpers.tpl")
	config := readRepoFile(t, "deploy", "slurm", "slurm-cluster", "templates", "configmap.yaml")
	schema := readRepoFile(t, "deploy", "slurm", "slurm-cluster", "values.schema.json")
	values := readRepoFile(t, "deploy", "slurm", "slurm-cluster", "values.yaml")
	adapter := readRepoFile(t, "pkg", "provider_daemon", "slurm_k8s", "adapter.go")

	for label, required := range map[string][]string{
		"chart":   {"StatefulSet", ".Values.nodePools", "slurm-cluster.nodePool.enabled", "virtengine.com/node-pool"},
		"helpers": {"slurm-cluster.dnsName", "sha256sum $raw", "ordinalBudget", "is reserved by an existing chart resource", "slurm-cluster.compute.capacity", "slurm-cluster.partition.capacity", "selects unknown node pool", "selects disabled node pool", "at least one compute replica must be enabled"},
		"config":  {`include "slurm-cluster.compute.capacity" . | fromJson`, `include "slurm-cluster.partition.capacity"`, "Nodes={{ $partitionCapacity.nodes }}", "MaxNodes={{ $partitionCapacity.replicas }}"},
		"schema":  {`"nodePools"`, `"uniqueItems": true`, `"controller"`, `"node-agent"`},
		"values":  {"compute:", "replicas: 2", "nodePools: []"},
		"adapter": {"func (a *SLURMKubernetesAdapter) Scale", "poolName", "waitForScaledStatefulSet"},
	} {
		var contents string
		switch label {
		case "chart":
			contents = chart
		case "helpers":
			contents = helpers
		case "config":
			contents = config
		case "schema":
			contents = schema
		case "values":
			contents = values
		case "adapter":
			contents = adapter
		}
		for _, needle := range required {
			if !strings.Contains(contents, needle) {
				t.Fatalf("offline SLURM contract missing %q in %s", needle, label)
			}
		}
	}
}

func TestReplicaCapacityOfflineContracts(t *testing.T) {
	verifyOfflineSLURMContracts(t)
}

func TestImmutableImagesOfflineContracts(t *testing.T) {
	chartRoot := filepath.Join(repoRoot(t), "deploy", "slurm", "slurm-cluster")
	helpers := readRepoFile(t, "deploy", "slurm", "slurm-cluster", "templates", "_helpers.tpl")
	values := readRepoFile(t, "deploy", "slurm", "slurm-cluster", "values.yaml")
	schema := readRepoFile(t, "deploy", "slurm", "slurm-cluster", "values.schema.json")
	fixture := readRepoFile(t, "deploy", "slurm", "slurm-cluster", "tests", "stable-secrets-values.yaml")

	for _, required := range []string{
		`define "slurm-cluster.immutableImage"`,
		`required (printf`,
		`@sha256:[a-f0-9]{64}$`,
		`slurm-cluster.munge.image`,
		`slurm-cluster.controller.image`,
		`slurm-cluster.database.image`,
		`slurm-cluster.mariadb.image`,
		`slurm-cluster.compute.image`,
		`slurm-cluster.nodeAgent.image`,
		`slurm-cluster.utility.image`,
	} {
		require.Contains(t, helpers, required)
	}
	helperPattern := `^[a-z0-9]+([._-][a-z0-9]+)*(:[0-9]+)?(/[a-z0-9]+([._-][a-z0-9]+)*)*@sha256:[a-f0-9]{64}$`
	_, err := regexp.Compile(helperPattern)
	require.NoError(t, err)
	require.Contains(t, helpers, `regexMatch "`+helperPattern+`"`)
	require.NotContains(t, values, "repository:")
	require.NotContains(t, values, "tag:")
	require.Equal(t, 7, strings.Count(values, `reference: ""`))
	require.Contains(t, schema, `"required": ["reference", "pullPolicy"]`)
	require.Contains(t, schema, `@sha256:[a-f0-9]{64}$`)

	exactReference := regexp.MustCompile(`(?m)^\s+reference:\s+\S+@sha256:[a-f0-9]{64}\s*$`)
	require.Len(t, exactReference.FindAllString(fixture, -1), 7)

	templates, err := filepath.Glob(filepath.Join(chartRoot, "templates", "*.yaml"))
	require.NoError(t, err)
	imageHelper := regexp.MustCompile(`include "slurm-cluster\.(munge|controller|database|mariadb|compute|nodeAgent|utility)\.image"`)
	for _, template := range templates {
		content, err := os.ReadFile(template)
		require.NoError(t, err)
		for _, line := range strings.Split(string(content), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "image:") {
				require.Regexp(t, imageHelper, line, "container image bypasses immutable helper in %s", template)
			}
		}
	}
}

func TestDurableStateOfflineContracts(t *testing.T) {
	values := readRepoFile(t, "deploy", "slurm", "slurm-cluster", "values.yaml")
	schema := readRepoFile(t, "deploy", "slurm", "slurm-cluster", "values.schema.json")
	helpers := readRepoFile(t, "deploy", "slurm", "slurm-cluster", "templates", "_helpers.tpl")
	runbook := readRepoFile(t, "_docs", "runbooks", "hpc-slurm", "deployment-runbook.md")
	drill := readRepoFile(t, "scripts", "hpc", "slurm-durable-state-drill.py")

	require.Equal(t, 3, strings.Count(values, `existingClaim: ""`))
	for _, required := range []string{`"durablePersistence"`, `"existingClaimReplicaSafety"`, `"replicas": { "const": 1 }`, `"enabled": { "const": true }`, `"existingClaim"`, `"ReadWriteOncePod"`} {
		require.Contains(t, schema, required)
	}
	require.Contains(t, helpers, `define "slurm-cluster.requireSafePersistenceReplicas"`)
	require.Contains(t, helpers, "HA must use generated per-replica claims")
	for _, required := range []string{"/var/spool/slurm", "/var/lib/mysql", "whenDeleted=Retain", "reclaimPolicy: Retain", "mariadb, slurmdbd, slurmctld", "sha256sum --check"} {
		require.Contains(t, runbook, required)
	}
	for _, required := range []string{`RESTORE_ORDER = ("mariadb", "slurmdbd", "slurmctld")`, "checksum verification failed", "destination.exists()"} {
		require.Contains(t, drill, required)
	}

	contracts := map[string]struct {
		component string
		volume    string
		path      string
	}{
		"controller-statefulset.yaml": {component: "controller", volume: "slurm-spool", path: "/var/spool/slurm"},
		"database-statefulset.yaml":   {component: "database", volume: "slurmdbd-spool", path: "/var/spool/slurm"},
		"mariadb-statefulset.yaml":    {component: "mariadb", volume: "mariadb-data", path: "/var/lib/mysql"},
	}
	for name, contract := range contracts {
		template := readRepoFile(t, "deploy", "slurm", "slurm-cluster", "templates", name)
		for _, required := range []string{
			`.Values.` + contract.component + `.persistence.existingClaim`,
			`include "slurm-cluster.requireSafePersistenceReplicas"`,
			"persistentVolumeClaimRetentionPolicy:",
			"whenDeleted: Retain",
			"whenScaled: Retain",
			"persistentVolumeClaim:",
			"volumeClaimTemplates:",
			"name: " + contract.volume,
			"mountPath: " + contract.path,
		} {
			require.Contains(t, template, required, name)
		}
		require.NotRegexp(t, `(?m)name:\s*`+regexp.QuoteMeta(contract.volume)+`\s*\n\s*emptyDir:`, template, name)
	}
}

func TestLeastPrivilegeOfflineContracts(t *testing.T) {
	chartRoot := filepath.Join(repoRoot(t), "deploy", "slurm", "slurm-cluster")
	helpers := readRepoFile(t, "deploy", "slurm", "slurm-cluster", "templates", "_helpers.tpl")
	values := readRepoFile(t, "deploy", "slurm", "slurm-cluster", "values.yaml")
	schema := readRepoFile(t, "deploy", "slurm", "slurm-cluster", "values.schema.json")
	config := readRepoFile(t, "deploy", "slurm", "slurm-cluster", "templates", "configmap.yaml")

	for _, required := range []string{
		`define "slurm-cluster.podSecurityContext"`,
		`define "slurm-cluster.containerSecurityContext"`,
		"allowPrivilegeEscalation: false",
		"privileged: false",
		"readOnlyRootFilesystem: true",
		"runAsNonRoot: true",
		"$identity.uid",
		"$identity.gid",
		`$identityKey = "slurm"`,
		`define "slurm-cluster.slurmUser"`,
		"type: RuntimeDefault",
	} {
		require.Contains(t, helpers, required)
	}
	require.NotContains(t, values, "podSecurityContext:")
	require.NotContains(t, values, "containerSecurityContext:")
	require.Contains(t, schema, `"hostCgroup": false`)
	require.Contains(t, schema, `"noSecurityOverrides"`)
	require.Contains(t, schema, `"minimum": 1`)
	require.Contains(t, values, "securityIdentities:")
	require.Contains(t, values, "slurm: { username: slurm, uid: 1002, gid: 1002 }")
	require.NotContains(t, values, "controller: { uid:")
	require.NotContains(t, values, "database: { uid:")
	require.NotContains(t, values, "compute: { uid:")
	require.Contains(t, config, `SlurmUser={{ include "slurm-cluster.slurmUser" . }}`)
	require.Contains(t, config, `SlurmdUser={{ include "slurm-cluster.slurmUser" . }}`)
	require.Contains(t, config, "JobAcctGatherType={{ .Values.controller.config.jobAcctGatherType | default \"jobacct_gather/linux\" }}")
	require.Contains(t, config, "ProctrackType={{ .Values.controller.config.proctrackType | default \"proctrack/linuxproc\" }}")
	require.Contains(t, config, "TaskPlugin={{ .Values.controller.config.taskPlugin | default \"task/affinity\" }}")
	require.Contains(t, config, "CgroupAutomount=no")
	require.Contains(t, config, "ConstrainCores=no")
	require.NotContains(t, config, "CgroupPlugin=")
	require.NotContains(t, config, "/sys/fs/cgroup")

	expectedContainerContexts := map[string]int{
		"controller-statefulset.yaml":        3,
		"database-statefulset.yaml":          4,
		"mariadb-statefulset.yaml":           1,
		"compute-statefulset.yaml":           5,
		"compute-nodepools-statefulset.yaml": 5,
	}
	for name, expected := range expectedContainerContexts {
		content := readRepoFile(t, "deploy", "slurm", "slurm-cluster", "templates", name)
		require.Equal(t, 1, strings.Count(content, `include "slurm-cluster.podSecurityContext"`), name)
		require.Equal(t, expected, strings.Count(content, `include "slurm-cluster.containerSecurityContext"`), name)
		for _, forbidden := range []string{"privileged: true", "runAsUser: 0", "runAsNonRoot: false", "hostPath:", "capabilities:\n    add:"} {
			require.NotContains(t, content, forbidden, name)
		}
	}

	_, err := os.Stat(chartRoot)
	require.NoError(t, err)
}

func TestSLURMIdentitySourceContractRejectsMissingUsersOrDivergentIDs(t *testing.T) {
	python := ""
	for _, candidate := range []string{"python", "python3"} {
		if path, err := exec.LookPath(candidate); err == nil {
			python = path
			break
		}
	}
	if python == "" {
		t.Skip("python is required for the SLURM source-contract validator")
	}

	tests := map[string]func(string){
		"missing SlurmUser": func(chartRoot string) {
			replaceChartText(t, chartRoot, "templates/configmap.yaml", `SlurmUser={{ include "slurm-cluster.slurmUser" . }}`, "")
		},
		"missing SlurmdUser": func(chartRoot string) {
			replaceChartText(t, chartRoot, "templates/configmap.yaml", `SlurmdUser={{ include "slurm-cluster.slurmUser" . }}`, "")
		},
		"divergent compute IDs": func(chartRoot string) {
			replaceChartText(t, chartRoot, "values.yaml", "  mariadb: { uid: 1004, gid: 1004 }", "  compute: { uid: 2002, gid: 2002 }\n  mariadb: { uid: 1004, gid: 1004 }")
		},
	}

	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			source := filepath.Join(repoRoot(t), "deploy", "slurm", "slurm-cluster")
			chartRoot := filepath.Join(t.TempDir(), "chart")
			require.NoError(t, os.CopyFS(chartRoot, os.DirFS(source)))
			mutate(chartRoot)

			validator := filepath.Join(repoRoot(t), "scripts", "validate_slurm_chart_semantics.py")
			command := exec.Command(python, validator, "--chart", chartRoot, "--diagnostic", "--json")
			output, err := command.CombinedOutput()
			require.Error(t, err, "diagnostic validation must remain blocking")
			var report struct {
				Findings []struct {
					Invariant string `json:"invariant"`
				} `json:"findings"`
			}
			require.NoError(t, json.Unmarshal(output, &report), string(output))
			require.Contains(t, findingInvariants(report.Findings), "least-privilege", string(output))
		})
	}
}

func replaceChartText(t *testing.T, chartRoot, relative, oldText, newText string) {
	t.Helper()
	path := filepath.Join(chartRoot, filepath.FromSlash(relative))
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	replaced := strings.Replace(string(content), oldText, newText, 1)
	require.NotEqual(t, string(content), replaced, "mutation did not match %s", path)
	require.NoError(t, os.WriteFile(path, []byte(replaced), 0o600))
}

func findingInvariants(findings []struct {
	Invariant string `json:"invariant"`
}) []string {
	result := make([]string, 0, len(findings))
	for _, finding := range findings {
		result = append(result, finding.Invariant)
	}
	return result
}

func checkPrerequisites(t *testing.T) bool {
	t.Helper()

	requiredTools := []string{"kind", "kubectl", "helm"}
	missing := make([]string, 0, len(requiredTools))
	for _, tool := range requiredTools {
		if _, err := exec.LookPath(tool); err != nil {
			missing = append(missing, tool)
		}
	}
	if len(missing) > 0 {
		t.Logf("real Kubernetes deployment harness unavailable; validating source contract guards only (rendered replica-capacity equality remains unverified): missing %v", missing)
		verifyOfflineSLURMContracts(t)
		return false
	}
	return true
}

func createKindCluster(ctx context.Context, name string) error {
	// Create kind cluster config
	config := `
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
  - role: worker
  - role: worker
`
	configFile, err := os.CreateTemp("", "kind-config-*.yaml")
	if err != nil {
		return err
	}
	defer os.Remove(configFile.Name())

	if _, err := configFile.WriteString(config); err != nil {
		return err
	}
	configFile.Close()

	cmd := exec.CommandContext(ctx, "kind", "create", "cluster",
		"--name", name,
		"--config", configFile.Name(),
		"--wait", "5m")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func deleteKindCluster(name string) error {
	cmd := exec.Command("kind", "delete", "cluster", "--name", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runKubectl(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runHelm(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "helm", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func verifyDeployment(ctx context.Context, namespace, releaseName string) error {
	// Wait for controller to be ready
	if err := runKubectl(ctx, "rollout", "status", "statefulset/"+releaseName+"-slurm-cluster-controller",
		"-n", namespace, "--timeout=5m"); err != nil {
		return err
	}

	// Wait for database to be ready
	if err := runKubectl(ctx, "rollout", "status", "statefulset/"+releaseName+"-slurm-cluster-slurmdbd",
		"-n", namespace, "--timeout=5m"); err != nil {
		return err
	}

	// Wait for compute nodes to be ready
	if err := runKubectl(ctx, "rollout", "status", "statefulset/"+releaseName+"-slurm-cluster-compute",
		"-n", namespace, "--timeout=5m"); err != nil {
		return err
	}

	return nil
}

func testSLURMFunctionality(ctx context.Context, namespace, releaseName string) error {
	controllerPod := releaseName + "-slurm-cluster-controller-0"

	// Test scontrol ping
	if err := runKubectl(ctx, "exec", "-n", namespace, controllerPod, "-c", "slurmctld", "--",
		"scontrol", "ping"); err != nil {
		return err
	}

	// Test sinfo
	if err := runKubectl(ctx, "exec", "-n", namespace, controllerPod, "-c", "slurmctld", "--",
		"sinfo"); err != nil {
		return err
	}

	// Test squeue
	if err := runKubectl(ctx, "exec", "-n", namespace, controllerPod, "-c", "slurmctld", "--",
		"squeue"); err != nil {
		return err
	}

	return nil
}

func testScaling(ctx context.Context, namespace, releaseName, chartPath string) error {
	// Scale up to 4 nodes
	if err := runHelm(ctx, "upgrade", releaseName, chartPath,
		"--namespace", namespace,
		"--reuse-values",
		"--set", "compute.replicas=4",
		"--wait",
		"--timeout", "5m"); err != nil {
		return err
	}

	// Verify new nodes are registered
	controllerPod := releaseName + "-slurm-cluster-controller-0"
	if err := runKubectl(ctx, "exec", "-n", namespace, controllerPod, "-c", "slurmctld", "--",
		"sinfo", "-N"); err != nil {
		return err
	}

	// Scale back down
	if err := runHelm(ctx, "upgrade", releaseName, chartPath,
		"--namespace", namespace,
		"--reuse-values",
		"--set", "compute.replicas=2",
		"--wait",
		"--timeout", "5m"); err != nil {
		return err
	}

	return nil
}

func randomSuffix() string {
	return time.Now().Format("150405")
}

// TestAdapterIntegration tests the SLURM Kubernetes adapter with mocks
func TestAdapterIntegration(t *testing.T) {
	k8s := &slurm_k8s.MockK8sChecker{
		StatefulSetStatus: map[string]*slurm_k8s.StatefulSetStatus{
			"test-ns/slurm-integration-test-controller": {ReadyReplicas: 1},
			"test-ns/slurm-integration-test-slurmdbd":   {ReadyReplicas: 1},
			"test-ns/slurm-integration-test-compute":    {ReadyReplicas: 2, Replicas: 2},
		},
		ExecOutput: map[string]string{
			"slurm-integration-test-controller-0:scontrol": "Slurmctld(primary) at slurm-integration-test-controller-0 is UP",
			"slurm-integration-test-controller-0:sinfo":    "compute-0 64 256000 (null) idle\ncompute-1 64 256000 (null) idle",
		},
	}

	helm := &slurm_k8s.MockHelmClient{
		InstallFunc: func(ctx context.Context, releaseName, chartPath, namespace string, values map[string]interface{}) error {
			t.Logf("Mock Helm install: %s in %s", releaseName, namespace)
			return nil
		},
		UpgradeFunc: func(ctx context.Context, releaseName, chartPath, namespace string, values map[string]interface{}) error {
			t.Logf("Mock Helm upgrade: %s", releaseName)
			if rawCompute, ok := values["compute"]; ok {
				if computeValues, ok := rawCompute.(map[string]interface{}); ok {
					if rawReplicas, ok := computeValues["replicas"]; ok {
						replicas := int32(0)
						switch v := rawReplicas.(type) {
						case int:
							replicas = int32(v)
						case int32:
							replicas = v
						case int64:
							replicas = int32(v)
						}
						if replicas > 0 {
							k8s.StatefulSetStatus["test-ns/slurm-integration-test-compute"] = &slurm_k8s.StatefulSetStatus{
								ReadyReplicas: replicas,
								Replicas:      replicas,
							}
						}
					}
				}
			}
			return nil
		},
		UninstallFunc: func(ctx context.Context, releaseName, namespace string) error {
			t.Logf("Mock Helm uninstall: %s", releaseName)
			return nil
		},
	}

	reporter := &slurm_k8s.MockReporter{}

	adapter := slurm_k8s.NewSLURMKubernetesAdapter(slurm_k8s.AdapterConfig{
		Helm:      helm,
		K8s:       k8s,
		Reporter:  reporter,
		ChartPath: "/charts/slurm",
	})

	ctx := context.Background()

	if err := adapter.Start(ctx); err != nil {
		t.Fatalf("failed to start adapter: %v", err)
	}
	defer adapter.Stop()

	// Deploy cluster
	config := slurm_k8s.DeploymentConfig{
		ClusterID:        "integration-test",
		ClusterName:      "Integration Test Cluster",
		Namespace:        "test-ns",
		ProviderAddress:  "virtengine1test",
		ProviderEndpoint: "https://provider.test:8443",
	}

	cluster, err := adapter.Deploy(ctx, config)
	if err != nil {
		t.Fatalf("failed to deploy cluster: %v", err)
	}

	if cluster.State != slurm_k8s.ClusterStateRunning {
		t.Errorf("expected state running, got %s", cluster.State)
	}

	// Get health
	health, err := adapter.GetClusterHealth(ctx, "integration-test")
	if err != nil {
		t.Fatalf("failed to get health: %v", err)
	}

	if !health.ControllerReady {
		t.Error("expected controller to be ready")
	}

	// Scale cluster
	if err := adapter.Scale(ctx, "integration-test", slurm_k8s.ScaleRequest{TargetNodes: 4}); err != nil {
		t.Fatalf("failed to scale cluster: %v", err)
	}

	manager := slurm_k8s.NewLifecycleManager(adapter)
	if err := manager.NodeJoin(ctx, slurm_k8s.JoinRequest{
		ClusterID: "integration-test",
		NodeID:    "compute-2",
		CPUs:      64,
		MemoryGB:  256,
		GPUs:      4,
		GPUType:   "nvidia-h100",
		Features:  []string{"gpu", "infiniband"},
	}); err != nil {
		t.Fatalf("failed to join node: %v", err)
	}
	if err := manager.NodeLeave(ctx, slurm_k8s.LeaveRequest{
		ClusterID: "integration-test",
		NodeID:    "compute-2",
	}); err != nil {
		t.Fatalf("failed to leave node: %v", err)
	}

	// Terminate cluster
	if err := adapter.Terminate(ctx, "integration-test"); err != nil {
		t.Fatalf("failed to terminate cluster: %v", err)
	}

	// Verify reporter was called
	if len(reporter.StatusReports) == 0 {
		t.Error("expected status reports to be submitted")
	}
	if len(reporter.CapacityReports) == 0 {
		t.Error("expected capacity reports to be submitted")
	}
	require.GreaterOrEqual(t, reporter.CapacityReports[len(reporter.CapacityReports)-1].TotalNodes, int32(2))
	require.GreaterOrEqual(t, reporter.CapacityReports[len(reporter.CapacityReports)-1].AvailableNodes, int32(2))
	require.Contains(t, reporter.NodeJoins, "compute-2")
	require.Contains(t, reporter.NodeLeaves, "compute-2")
}
