package test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
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

func TestTerraformContractsAreFailClosed(t *testing.T) {
	t.Parallel()

	globalMain := readRepoFile(t, "infra", "terraform", "global", "main.tf")
	globalVars := readRepoFile(t, "infra", "terraform", "global", "variables.tf")
	scalingMain := readRepoFile(t, "infra", "terraform", "modules", "scaling", "main.tf")

	if strings.Contains(globalMain, "ffffffffffffffffffffffffffffffffffffffff") {
		t.Fatal("global OIDC provider still uses the placeholder thumbprint")
	}
	if !strings.Contains(globalMain, `data "tls_certificate" "github_actions"`) {
		t.Fatal("global OIDC provider no longer derives trust material from the live certificate chain")
	}
	if !strings.Contains(globalVars, "repo:virtengine/virtengine:environment:infra-prod") {
		t.Fatal("allowed GitHub OIDC subjects do not include the production infra environment")
	}
	if strings.Contains(scalingMain, "placeholder.elb.") {
		t.Fatal("scaling module still contains placeholder load balancer aliases")
	}
	if !strings.Contains(scalingMain, "lb_dns_name") || !strings.Contains(scalingMain, "lb_zone_id") {
		t.Fatal("scaling module does not require real regional load balancer aliases")
	}
}

func TestWorkflowContractsUseReviewedPlansAndInfraOwnedAutomation(t *testing.T) {
	t.Parallel()

	infraWorkflow := readRepoFile(t, ".github", "workflows", "infrastructure.yaml")
	multiRegionWorkflow := readRepoFile(t, ".github", "workflows", "multi-region-deploy.yaml")
	drWorkflow := readRepoFile(t, ".github", "workflows", "dr-failover-test.yaml")

	for name, contents := range map[string]string{
		"infrastructure":      infraWorkflow,
		"multi-region-deploy": multiRegionWorkflow,
	} {
		if !strings.Contains(contents, "id-token: write") {
			t.Fatalf("%s workflow is missing OIDC permissions", name)
		}
		if !strings.Contains(contents, "infra/scripts/terraform-run.sh") {
			t.Fatalf("%s workflow does not use the infra-owned terraform wrapper", name)
		}
		if strings.Contains(contents, "terraform apply -auto-approve") || strings.Contains(contents, "terragrunt apply -auto-approve") {
			t.Fatalf("%s workflow still performs direct auto-approve applies", name)
		}
	}

	if !strings.Contains(infraWorkflow, "infra/scripts/check-environment-parity.sh") {
		t.Fatal("infrastructure workflow no longer enforces the environment parity gate")
	}
	if !strings.Contains(drWorkflow, "infra/dr/run-failover-drill.sh") {
		t.Fatal("DR workflow does not use the infra-owned failover drill runner")
	}
	if !strings.Contains(drWorkflow, "failover-drill-evidence.json") {
		t.Fatal("DR workflow no longer publishes structured drill evidence")
	}
}

// The re-pin PR branch must be a STABLE name. With `branch: dr-tools/repin-$SHA`
// every publish mints its own PR, so N builds leave N open re-pins holding
// progressively older digests and merging them out of order rolls the pin
// backwards (seen 2026-10-02: #1092 wanted a digest older than the one #1140
// had already merged, and #1131's manifest blob was byte-identical to develop).
// A stable branch makes create-pull-request update the one open PR instead.
func TestDRToolsRePinBranchIsStable(t *testing.T) {
	t.Parallel()

	wf := readRepoFile(t, ".github", "workflows", "dr-tools-image.yaml")

	if strings.Contains(wf, "branch: dr-tools/repin-${{") {
		t.Fatal("dr-tools-image.yaml re-pins onto a per-commit branch, so each build opens a NEW PR; use a stable branch name")
	}
	if !strings.Contains(wf, "branch: dr-tools/repin\n") {
		t.Fatal("dr-tools-image.yaml no longer opens the re-pin on the stable dr-tools/repin branch")
	}
}
