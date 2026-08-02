"use strict";

const assert = require("assert").strict;
const { mkdtempSync, mkdirSync, readFileSync, rmSync, writeFileSync } = require("fs");
const { tmpdir } = require("os");
const { dirname, resolve } = require("path");
const { spawnSync } = require("child_process");
const {
  assertExecutionReady,
  buildExecutionPlan,
  collectChangedPaths,
  validateResultEnvelope,
} = require("./run-required-gates.cjs");

const root = resolve(__dirname, "..");
const matrixPath = resolve(root, "_docs/ralph/prototype-integration/required-gate-matrix.json");

function git(repo, ...args) {
  const result = spawnSync("git", ["-C", repo, ...args], { encoding: "utf8" });
  if (result.error) throw result.error;
  assert.equal(result.status, 0, `git ${args.join(" ")} failed: ${result.stderr}`);
  return result.stdout.trim();
}

function put(repo, path, contents) {
  const target = resolve(repo, path);
  mkdirSync(dirname(target), { recursive: true });
  writeFileSync(target, contents);
}

function commit(repo, message) {
  git(repo, "add", "--all");
  git(repo, "commit", "-m", message);
  return git(repo, "rev-parse", "HEAD");
}

function createMergeRepo() {
  const repo = mkdtempSync(resolve(tmpdir(), "virtengine-required-gates-"));
  git(repo, "init", "-b", "main");
  git(repo, "config", "user.email", "required-gates@example.invalid");
  git(repo, "config", "user.name", "Required Gate Test");
  put(repo, "shared.go", "package shared\n\nconst Value = \"base\"\n");
  put(repo, "README.md", "base\n");
  const base = commit(repo, "base");

  git(repo, "checkout", "-b", "feature");
  put(repo, "shared.go", "package shared\n\nconst Value = \"feature\"\n");
  put(repo, "sdk/feature.ts", "export const feature = true;\n");
  commit(repo, "feature");

  git(repo, "checkout", "main");
  put(repo, "shared.go", "package shared\n\nconst Value = \"main\"\n");
  put(repo, "portal/main.ts", "export const main = true;\n");
  commit(repo, "main");
  const merge = spawnSync("git", ["-C", repo, "merge", "--no-ff", "feature", "-m", "merge"], { encoding: "utf8" });
  assert.notEqual(merge.status, 0, "fixture merge must conflict");
  put(repo, "shared.go", "package shared\n\nconst Value = \"base\"\n");
  const head = commit(repo, "resolve merge");
  return { repo, base, head };
}

function validEnvelope(plan) {
  return {
    schema_version: "virtengine.task-88b.required-gate-results/v1",
    base_sha: plan.base_sha,
    head_sha: plan.head_sha,
    matrix_digest: plan.matrix_digest,
    results: plan.categories.flatMap((category) => category.commands.map((command) => ({
      category_id: category.id,
      command_id: command.id,
      command: command.command,
      outcome: "passed",
      exit_code: 0,
      discovered_tests: command.kind === "test" ? 2 : 0,
      executed_tests: command.kind === "test" ? 2 : 0,
      skipped_tests: 0,
      tools: category.pinned_tools.map((tool) => ({ name: tool.name, version: tool.version, available: true })),
    }))),
  };
}

function clone(value) {
  return JSON.parse(JSON.stringify(value));
}

const fixture = createMergeRepo();
try {
  const matrixBytes = readFileSync(matrixPath);
  const matrix = JSON.parse(matrixBytes.toString("utf8"));
  let caseNumber = 0;
  const planFor = (changedPaths) => {
    git(fixture.repo, "reset", "--hard", fixture.head);
    const base = git(fixture.repo, "rev-parse", "HEAD");
    caseNumber += 1;
    for (const path of changedPaths) put(fixture.repo, path, `case ${caseNumber}: ${path}\n`);
    const head = commit(fixture.repo, `plan case ${caseNumber}`);
    return buildExecutionPlan({ repoDir: fixture.repo, base, head, matrixPath });
  };

  const tests = [
    ["collects complete merge history including a final-tree-hidden path", () => {
      const paths = collectChangedPaths(fixture.repo, fixture.base, fixture.head);
      assert.ok(paths.includes("shared.go"));
      assert.ok(paths.includes("sdk/feature.ts"));
      assert.ok(paths.includes("portal/main.ts"));
      const plan = buildExecutionPlan({ repoDir: fixture.repo, base: fixture.base, head: fixture.head, matrixPath });
      assert.deepEqual(plan.categories.map((category) => category.id), ["go", "sdk", "portal"]);
      assert.deepEqual(plan, buildExecutionPlan({ repoDir: fixture.repo, base: fixture.base, head: fixture.head, matrixPath }));
    }],
    ["selects every overlapping category", () => {
      const plan = planFor(["pnpm-lock.yaml"]);
      assert.deepEqual(plan.categories.map((category) => category.id), ["sdk", "portal", "mobile"]);
    }],
    ["selects pinned Helm gates for a SLURM chart change", () => {
      const plan = planFor(["deploy/slurm/slurm-cluster/templates/configmap.yaml"]);
      assert.deepEqual(plan.categories.map((category) => category.id), ["deployment"]);
      assert.deepEqual(plan.categories[0].commands.filter((command) => command.id.startsWith("slurm-chart-")).map((command) => command.id), ["slurm-chart-lint", "slurm-chart-render"]);
      assert.ok(plan.categories[0].pinned_tools.some((tool) => tool.name === "helm" && tool.version === "3.18.6"));
    }],
    ["selects launcher parity for a Windows-only localnet change", () => {
      const plan = planFor(["scripts/localnet.ps1"]);
      assert.deepEqual(plan.categories.map((category) => category.id), ["docs_process_boundary_e2e"]);
      assert.ok(plan.categories[0].commands.some((command) => command.id === "localnet-integration-launchers"));
    }],
    ["selects launcher parity for a shell-only localnet change", () => {
      const plan = planFor(["scripts/localnet.sh"]);
      assert.deepEqual(plan.categories.map((category) => category.id), ["docs_process_boundary_e2e"]);
      assert.ok(plan.categories[0].commands.some((command) => command.id === "localnet-integration-launchers"));
    }],
    ["selects docs/process gates for a root PowerShell script", () => {
      const plan = planFor(["scripts/agent-preflight.ps1"]);
      assert.deepEqual(plan.categories.map((category) => category.id), ["docs_process_boundary_e2e"]);
    }],
    ["selects docs/process gates for a root shell script", () => {
      const plan = planFor(["scripts/verify-modules.sh"]);
      assert.deepEqual(plan.categories.map((category) => category.id), ["docs_process_boundary_e2e"]);
    }],
    ["selects deployment and docs gates for the SLURM validator", () => {
      const plan = planFor(["scripts/validate_slurm_chart_semantics.py", "scripts/validate-agents-docs.mjs"]);
      assert.deepEqual(plan.categories.map((category) => category.id), ["deployment", "docs_process_boundary_e2e"]);
      assert.deepEqual(plan.categories[0].commands.filter((command) => command.id.startsWith("slurm-chart-")).map((command) => command.id), ["slurm-chart-lint", "slurm-chart-render"]);
    }],
    ["selects docs/process gates for root script Markdown and SQL", () => {
      const plan = planFor(["scripts/AGENTS.md", "scripts/archive-vk-data.sql"]);
      assert.deepEqual(plan.categories.map((category) => category.id), ["docs_process_boundary_e2e"]);
    }],
    ["selects docs/process gates for the scripts bosun Gitlink", () => {
      const plan = planFor(["scripts/bosun"]);
      assert.deepEqual(plan.categories.map((category) => category.id), ["docs_process_boundary_e2e"]);
    }],
    ["records an explicit root metadata allowance", () => {
      const plan = planFor(["README.md"]);
      assert.equal(plan.categories.length, 0);
      assert.deepEqual(plan.allowlisted_paths.map((entry) => entry.path), ["README.md"]);
    }],
    ["rejects an unmatched changed path", () => {
      assert.throws(() => planFor(["unowned/component.rs"]), /no required-gate category or metadata allowlist/);
    }],
    ["accepts one exact result per selected command", () => {
      const plan = planFor(["shared.go"]);
      assert.equal(validateResultEnvelope(plan, validEnvelope(plan)), true);
    }],
    ["rejects a missing result", () => {
      const plan = planFor(["shared.go"]);
      const envelope = validEnvelope(plan);
      envelope.results.pop();
      assert.throws(() => validateResultEnvelope(plan, envelope), /missing gate results/);
    }],
    ["rejects duplicate and extra results", () => {
      const plan = planFor(["shared.go"]);
      const duplicate = validEnvelope(plan);
      duplicate.results.push(clone(duplicate.results[0]));
      assert.throws(() => validateResultEnvelope(plan, duplicate), /duplicate gate result/);
      const extra = validEnvelope(plan);
      extra.results.push({ ...clone(extra.results[0]), command_id: "not-selected" });
      assert.throws(() => validateResultEnvelope(plan, extra), /extra gate result/);
    }],
    ["rejects literal command drift and cancellation", () => {
      const plan = planFor(["shared.go"]);
      const wrongCommand = validEnvelope(plan);
      wrongCommand.results[0].command += " --changed";
      assert.throws(() => validateResultEnvelope(plan, wrongCommand), /literal command mismatch/);
      const cancelled = validEnvelope(plan);
      cancelled.results[0].outcome = "cancelled";
      assert.throws(() => validateResultEnvelope(plan, cancelled), /must pass/);
    }],
    ["rejects zero discovered or executed tests and skipped tests", () => {
      const plan = planFor(["shared.go"]);
      const testIndex = validEnvelope(plan).results.findIndex((result) => result.command_id === "go-test");
      const zeroDiscovered = validEnvelope(plan);
      zeroDiscovered.results[testIndex].discovered_tests = 0;
      assert.throws(() => validateResultEnvelope(plan, zeroDiscovered), /zero tests discovered/);
      const zeroExecuted = validEnvelope(plan);
      zeroExecuted.results[testIndex].executed_tests = 0;
      assert.throws(() => validateResultEnvelope(plan, zeroExecuted), /zero tests executed/);
      const skipped = validEnvelope(plan);
      skipped.results[testIndex].skipped_tests = 1;
      assert.throws(() => validateResultEnvelope(plan, skipped), /skipped tests/);
      const incomplete = validEnvelope(plan);
      incomplete.results[testIndex].discovered_tests = 2;
      incomplete.results[testIndex].executed_tests = 1;
      assert.throws(() => validateResultEnvelope(plan, incomplete), /not all discovered tests executed/);
    }],
    ["rejects wrong SHA and matrix digest", () => {
      const plan = planFor(["shared.go"]);
      const wrongSha = validEnvelope(plan);
      wrongSha.head_sha = "0".repeat(40);
      assert.throws(() => validateResultEnvelope(plan, wrongSha), /head SHA mismatch/);
      const wrongDigest = validEnvelope(plan);
      wrongDigest.matrix_digest = "0".repeat(64);
      assert.throws(() => validateResultEnvelope(plan, wrongDigest), /matrix digest mismatch/);
    }],
    ["rejects missing, unavailable, or wrong pinned tools", () => {
      const plan = planFor(["shared.go"]);
      const missing = validEnvelope(plan);
      missing.results[0].tools = [];
      assert.throws(() => validateResultEnvelope(plan, missing), /pinned tool count mismatch/);
      const unavailable = validEnvelope(plan);
      unavailable.results[0].tools[0].available = false;
      assert.throws(() => validateResultEnvelope(plan, unavailable), /missing pinned tool/);
      const wrongVersion = validEnvelope(plan);
      wrongVersion.results[0].tools[0].version = "0.0.0";
      assert.throws(() => validateResultEnvelope(plan, wrongVersion), /pinned tool version mismatch/);
    }],
    ["refuses dependency-blocked execution without an environment bypass", () => {
      const plan = planFor(["shared.go"]);
      process.env.VE_REQUIRED_GATES_BYPASS = "1";
      try {
        assert.throws(() => assertExecutionReady(plan), /execution refused/);
      } finally {
        delete process.env.VE_REQUIRED_GATES_BYPASS;
      }
    }],
    ["refuses matrix-wide blocked execution for allowlist-only changes", () => {
      assert.throws(() => assertExecutionReady(planFor(["README.md"])), /matrix status is dependency_blocked/);
    }],
    ["rejects a reversed base and head range", () => {
      assert.throws(() => buildExecutionPlan({ repoDir: fixture.repo, base: fixture.head, head: fixture.base, matrixPath }), /base must be an ancestor/);
    }],
  ];

  for (const [name, run] of tests) {
    run();
    console.log(`ok - ${name}`);
  }
} finally {
  rmSync(fixture.repo, { recursive: true, force: true });
}