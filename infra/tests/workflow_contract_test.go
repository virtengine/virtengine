package test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
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

// yamlBool renders a workflow `with:` value that YAML may decode as either a
// bool (`delete-branch: true`) or a string (`delete-branch: "true"`).
func yamlBool(v any) (bool, bool) {
	switch typed := v.(type) {
	case bool:
		return typed, true
	case string:
		switch strings.ToLower(strings.TrimSpace(typed)) {
		case "true":
			return true, true
		case "false":
			return false, true
		}
	}
	return false, false
}

// parseCreatePullRequestStep finds the single `with:` block of the
// peter-evans/create-pull-request step in a parsed workflow, matched by
// `actionPrefix`.
//
// Parsing rather than grepping raw bytes is what makes the contract about the
// RESOLVED value: quoting the scalar, re-indenting the step, or letting a YAML
// formatter normalise the document leaves every assertion in
// checkStableRePinBranch true, where a byte-literal strings.Contains would
// instead red-line an unchanged, correct workflow.
//
// Failing closed matters as much as the comparison: if the step is renamed,
// dropped, or duplicated the search returns an error, never a vacuous empty
// `with` map that a caller might read as "nothing configured".
func parseCreatePullRequestStep(workflowFile string, doc []byte, actionPrefix string) (jobName, stepName string, with map[string]any, err error) {
	var workflow struct {
		Jobs map[string]struct {
			Steps []struct {
				Name string         `yaml:"name"`
				Uses string         `yaml:"uses"`
				With map[string]any `yaml:"with"`
			} `yaml:"steps"`
		} `yaml:"jobs"`
	}
	if err := yaml.Unmarshal(doc, &workflow); err != nil {
		return "", "", nil, fmt.Errorf("%s is not valid YAML: %w", workflowFile, err)
	}
	if len(workflow.Jobs) == 0 {
		return "", "", nil, fmt.Errorf("%s parsed to zero jobs; the workflow file is unusable", workflowFile)
	}

	type hit struct{ job, step string }
	var found []hit
	for name, job := range workflow.Jobs {
		for _, step := range job.Steps {
			if strings.HasPrefix(step.Uses, actionPrefix) {
				found = append(found, hit{name, step.Name})
				if with == nil {
					with = step.With
				}
			}
		}
	}

	switch len(found) {
	case 1:
	case 0:
		return "", "", nil, fmt.Errorf("%s has no step using %s*, so its automation contract is unguarded", workflowFile, actionPrefix)
	default:
		names := make([]string, 0, len(found))
		for _, h := range found {
			names = append(names, h.job+"/"+h.step)
		}
		sort.Strings(names)
		return "", "", nil, fmt.Errorf("%s has %d steps using %s* (%s); the automation contract is ambiguous, so it is not enforced",
			workflowFile, len(found), actionPrefix, strings.Join(names, ", "))
	}

	if with == nil {
		with = map[string]any{}
	}
	return found[0].job, found[0].step, with, nil
}

// createPullRequestStep is the *testing.T wrapper around
// parseCreatePullRequestStep for a real workflow file on disk.
func createPullRequestStep(t *testing.T, workflowFile, actionPrefix string) (jobName, stepName string, with map[string]any) {
	t.Helper()
	doc := []byte(readRepoFile(t, strings.Split(workflowFile, "/")...))
	jobName, stepName, with, err := parseCreatePullRequestStep(workflowFile, doc, actionPrefix)
	if err != nil {
		t.Fatal(err)
	}
	return jobName, stepName, with
}

// checkStableRePinBranch holds the re-pin contract itself. It takes already
// parsed inputs and returns an error rather than calling t.Fatal, so the exact
// same assertions can be driven from fixtures as well as from the real workflow
// file — which is what lets the formatting-insensitivity test below exist.
func checkStableRePinBranch(workflowFile, job, step string, with map[string]any) error {
	branch, ok := with["branch"].(string)
	if !ok {
		return fmt.Errorf("%s: %s/%s has no string `branch:` input (got %T %v); the re-pin PR target is unguarded",
			workflowFile, job, step, with["branch"], with["branch"])
	}
	branch = strings.TrimSpace(branch)

	if strings.Contains(branch, "${{") {
		return fmt.Errorf("%s: %s/%s re-pins onto branch %q, which is expanded per run, so every build opens a NEW PR; use the stable dr-tools/repin branch",
			workflowFile, job, step, branch)
	}
	if branch != "dr-tools/repin" {
		return fmt.Errorf("%s: %s/%s re-pins onto branch %q, want the stable dr-tools/repin branch so each publish updates the one open PR",
			workflowFile, job, step, branch)
	}

	base, ok := with["base"].(string)
	if !ok || strings.TrimSpace(base) != "main" {
		return fmt.Errorf("%s: %s/%s re-pins against base %v, want main", workflowFile, job, step, with["base"])
	}

	// delete-branch keeps per-publish leftovers out of the branch list; without
	// it every closed re-pin leaves an orphaned branch behind.
	deleteBranch, ok := yamlBool(with["delete-branch"])
	if !ok || !deleteBranch {
		return fmt.Errorf("%s: %s/%s has delete-branch %v, want true, or closed re-pins leave orphaned branches behind",
			workflowFile, job, step, with["delete-branch"])
	}
	return nil
}

func assertStableRePinBranch(t *testing.T, workflowFile, job, step string, with map[string]any) {
	t.Helper()
	if err := checkStableRePinBranch(workflowFile, job, step, with); err != nil {
		t.Fatal(err)
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

	const workflowFile = ".github/workflows/dr-tools-image.yaml"

	job, step, with := createPullRequestStep(t, workflowFile, "peter-evans/create-pull-request")
	assertStableRePinBranch(t, workflowFile, job, step, with)
}

// TestStableRePinBranchAssertionsAreFormattingInsensitive is the regression
// test for the defect that made this guard a liability: the assertion used to be
// a byte-literal `strings.Contains(wf, "branch: dr-tools/repin\n")`, so merely
// QUOTING the value — semantically identical YAML, the sort of thing a YAML
// formatter or a cautious author does — turned a correct workflow red. Each
// `wantPass` case below is a formatting variation that must stay green, and
// each `wantFail` case is a real regression that must stay red.
func TestStableRePinBranchAssertionsAreFormattingInsensitive(t *testing.T) {
	t.Parallel()

	const document = `name: fixture
jobs:
  pin:
    name: Re-pin
    steps:
      - name: Open re-pin PR
        uses: peter-evans/create-pull-request@v8
        with:
          token: ${{ secrets.GITHUB_TOKEN }}
%s
`

	for _, tc := range []struct {
		name     string
		block    string
		wantPass bool
	}{
		{"unquoted", "          branch: dr-tools/repin\n          base: main\n          delete-branch: true", true},
		{"double quoted", `          branch: "dr-tools/repin"` + "\n" + `          base: "main"` + "\n" + `          delete-branch: "true"`, true},
		{"single quoted", "          branch: 'dr-tools/repin'\n          base: 'main'\n          delete-branch: 'true'", true},
		{"extra inner whitespace", "          branch:    dr-tools/repin\n          base:  main\n          delete-branch:   true", true},
		{"per-commit branch", "          branch: dr-tools/repin-${{ github.sha }}\n          base: main\n          delete-branch: true", false},
		{"per-commit branch quoted", "          branch: 'dr-tools/repin-${{ github.sha }}'\n          base: main\n          delete-branch: true", false},
		{"wrong stable name", "          branch: dr-tools/repin-v2\n          base: main\n          delete-branch: true", false},
		{"wrong base", "          branch: dr-tools/repin\n          base: develop\n          delete-branch: true", false},
		{"delete-branch false", "          branch: dr-tools/repin\n          base: main\n          delete-branch: false", false},
		{"delete-branch absent", "          branch: dr-tools/repin\n          base: main", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			doc := []byte(fmt.Sprintf(document, tc.block))
			job, step, with, err := parseCreatePullRequestStep("fixture.yaml", doc, "peter-evans/create-pull-request")
			if err != nil {
				t.Fatalf("fixture did not yield the step: %v", err)
			}
			if job != "pin" || step != "Open re-pin PR" {
				t.Fatalf("parsed the wrong step: %s/%s", job, step)
			}

			err = checkStableRePinBranch("fixture.yaml", job, step, with)
			if tc.wantPass && err != nil {
				t.Fatalf("formatting variation %q must pass, got: %v", tc.name, err)
			}
			if !tc.wantPass && err == nil {
				t.Fatalf("regression %q must fail, but the contract accepted it", tc.name)
			}
		})
	}
}

// TestCreatePullRequestStepFailsClosed proves the search cannot go vacuous: a
// document with no create-pull-request step, or with two of them, must error
// rather than return an empty `with` map that the assertions above would then
// read as "no branch input configured".
func TestCreatePullRequestStepFailsClosed(t *testing.T) {
	t.Parallel()

	for name, document := range map[string]string{
		"no such step": `jobs:
  pin:
    steps:
      - name: Build
        uses: actions/setup-go@v6
`,
		"two such steps": `jobs:
  pin:
    steps:
      - uses: peter-evans/create-pull-request@v8
        with:
          branch: dr-tools/repin
  pin2:
    steps:
      - uses: peter-evans/create-pull-request@v8
        with:
          branch: dr-tools/repin
`,
		"zero jobs": `name: empty
`,
		"not yaml": "jobs: [unclosed\n",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, _, _, err := parseCreatePullRequestStep("", []byte(document), "peter-evans/create-pull-request"); err == nil {
				t.Fatalf("parse must fail closed for %q, got no error", name)
			}
		})
	}
}

// TestCreatePullRequestStepWithoutWithBlockFailsTheContract covers the
// remaining vacuity path: exactly ONE step resolves, so the parser is happy,
// but the step carries no `with:` block at all. The parser must hand back a
// non-nil empty map (never nil, which would nil-deref at the caller) and the
// contract must reject it — otherwise "branch input missing" reads as a skip.
func TestCreatePullRequestStepWithoutWithBlockFailsTheContract(t *testing.T) {
	t.Parallel()

	const document = `jobs:
  pin:
    steps:
      - uses: peter-evans/create-pull-request@v8
`

	_, _, with, err := parseCreatePullRequestStep("no-with.yaml", []byte(document), "peter-evans/create-pull-request")
	if err != nil {
		t.Fatalf("a single create-pull-request step must parse, got: %v", err)
	}
	if with == nil {
		t.Fatal("parser returned a nil with map; the contract would nil-deref instead of reporting the missing inputs")
	}
	if err := checkStableRePinBranch("no-with.yaml", "pin", "", with); err == nil {
		t.Fatal("a create-pull-request step with no branch/base/delete-branch must fail the contract")
	}
}
