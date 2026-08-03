"use strict";

const assert = require("assert").strict;
const { createHash } = require("crypto");
const { execFileSync } = require("child_process");
const { mkdtempSync, mkdirSync, readFileSync, rmSync, writeFileSync } = require("fs");
const { tmpdir } = require("os");
const { dirname, join } = require("path");
const test = require("node:test");
const { REQUIRED_CI_CHECKS, TAG, parseArgs, preflight, validateCandidateEpochSelection, validateReport, validateReportSchema } = require("./preflight-core-rc-publication.cjs");
const { planDigest } = require("./run-required-gates.cjs");

function fixturePlan(state) {
  return {
    schema_version: "virtengine.task-88b.required-gate-plan/v1",
    base_sha: state.controls.results.base_sha,
    head_sha: state.candidate,
    matrix_digest: state.controls.results.matrix_digest,
    matrix_status: "complete",
    changed_paths: [],
    allowlisted_paths: [],
    categories: state.controls.matrix.categories.map((category) => ({ id: category.id, status: category.status, matched_paths: [], matched_selectors: [], commands: category.required_commands, pinned_tools: category.pinned_tools, dependencies: category.dependencies, blockers: category.blockers })),
  };
}

function git(repo, ...args) {
  return execFileSync("git", ["-C", repo, ...args], { encoding: "utf8", stdio: ["ignore", "pipe", "pipe"] }).trim();
}

function write(repo, path, content) {
  const target = join(repo, ...path.split("/"));
  mkdirSync(dirname(target), { recursive: true });
  writeFileSync(target, content);
}

function fixture(mutate = () => {}) {
  const directory = mkdtempSync(join(tmpdir(), "core-rc-preflight-"));
  const repo = join(directory, "repo");
  const remote = join(directory, "remote.git");
  git(directory, "init", "--bare", remote);
  git(directory, "init", "-b", "ve/prototype-integration", repo);
  git(repo, "config", "user.name", "Preflight Test");
  git(repo, "config", "user.email", "preflight@example.test");
  git(repo, "remote", "add", "fixture", remote);
  write(repo, "source.txt", "source\n");
  git(repo, "add", "--all");
  git(repo, "commit", "-m", "source");
  const source = git(repo, "rev-parse", "HEAD");
  write(repo, "_docs/ralph/prototype-integration/core-rc-manifest.json", "{}\n");
  git(repo, "add", "--all");
  git(repo, "commit", "-m", "candidate evidence");
  const candidate = git(repo, "rev-parse", "HEAD");
  git(repo, "push", "fixture", "ve/prototype-integration");
  const command = { id: "unit", kind: "test", command: "node --test" };
  const category = {
    id: "test", status: "complete", required_commands: [command], pinned_tools: [{ name: "node", version: "20.19.1" }],
    dependencies: [{ id: "88A", status: "available" }], blockers: [],
    zero_test_policy: { minimum_discovered: 1, minimum_executed: 1 },
  };
  const controls = {
    epoch: {
      status: "closed", announcement_cutoff: "2000-01-01T00:00:00Z",
      producers: ["T1", "T2", "T3", "T5"].map((thread) => ({ thread, status: "accepted", tag: `checkpoint/prototype-${thread.toLowerCase()}/${thread.toLowerCase()}-99`, decision: "accepted" })),
    },
    ledger: { generated_hashes: {}, accepted_checkpoints: [] },
    manifest: { schema_version: "virtengine.core-rc.prototype/v1", source: { payload_sha: source }, tooling: { tooling_sha: source }, required_gates: { matrix_path: "_docs/ralph/prototype-integration/required-gate-matrix.json" }, control_artifacts: [], blockers: [], authoritative: false, planned_functionality_complete: false, milestone_m_eligible: false, status: "complete" },
    manifestSchema: { version: "future-ready-v1" }, manifestText: "{}\n", manifestValid: true,
    matrixText: "{}\n",
    matrix: { status: "complete", completion_claim: true, categories: [category] }, matrixValid: true,
    results: {
      schema_version: "virtengine.task-88b.required-gate-results/v1",
      base_sha: source,
      head_sha: candidate,
      matrix_digest: createHash("sha256").update("{}\n").digest("hex"),
      plan_digest: "0".repeat(64),
      results: [{ category_id: "test", command_id: "unit", kind: "test", command: "node --test", outcome: "passed", exit_code: 0, discovered_tests: 1, executed_tests: 1, skipped_tests: 0, tools: [{ name: "node", version: "20.19.1", available: true }] }],
    },
    availability: { rollout: true, rollback: true, sbom: true, releaseProvenance: true, model: true, slurm: true, producers: true },
  };
  const state = { directory, repo, remote, candidate, controls, ci: [] };
  controls.results.plan_digest = planDigest(fixturePlan(state));
  const resultsDigest = createHash("sha256").update(`${JSON.stringify(controls.results, null, 2)}\n`).digest("hex");
  controls.resultsText = `${JSON.stringify(controls.results, null, 2)}\n`;
  controls.ledger.generated_hashes["_docs/ralph/prototype-integration/core-rc-manifest.json"] = createHash("sha256").update(controls.manifestText).digest("hex");
  controls.ledger.generated_hashes["_docs/ralph/prototype-integration/required-gate-results.json"] = resultsDigest;
  controls.manifest.control_artifacts.push({ path: "_docs/ralph/prototype-integration/required-gate-results.json", sha256: resultsDigest });
  controls.ledger.accepted_checkpoints = controls.epoch.producers.map((producer) => ({ thread: producer.thread, checkpoint: producer.tag.split("/").pop().toUpperCase(), tag: producer.tag, tip: candidate, payload_head: source }));
  state.ci = REQUIRED_CI_CHECKS.map(({ workflow, workflowPath, job, expectsResults }, index) => ({ repository: "virtengine/virtengine", workflow, workflowPath, workflowId: index + 1, runId: index + 10, runAttempt: 1, runConclusion: "success", event: "push", branch: "ve/prototype-integration", job, jobConclusion: "success", sha: candidate, resultsDigest: expectsResults ? resultsDigest : undefined }));
  mutate(state);
  return state;
}

async function run(state, overrides = {}) {
  const planProvider = async () => fixturePlan(state);
  return preflight({
    candidate: state.candidate, epoch: 1, tag: TAG, repo: state.repo, remote: "fixture", controls: state.controls,
    ciProvider: async () => state.ci, intakeProvider: async () => true, planProvider,
    manifestValidator: async (manifest, schema) => state.controls.manifestValid !== false && manifest.schema_version === "virtengine.core-rc.prototype/v1" && schema.version === "future-ready-v1",
    ...overrides,
  });
}

function blocked(name, mutate, blocker) {
  test(name, async () => {
    const state = fixture(mutate);
    try {
      const report = await run(state);
      assert.equal(report.passed, false);
      assert.ok(report.blockers.includes(blocker), report.blockers.join(", "));
    } finally {
      rmSync(state.directory, { recursive: true, force: true });
    }
  });
}

test("produces a valid deterministic synthetic pass report without creating a tag", async () => {
  const state = fixture();
  try {
    const first = await run(state);
    const second = await run(state);
    assert.equal(first.passed, true, first.blockers.join(", "));
    assert.deepEqual(first, second);
    assert.doesNotThrow(() => validateReport(first));
    assert.throws(() => git(state.repo, "rev-parse", `refs/tags/${TAG}`));
  } finally {
    rmSync(state.directory, { recursive: true, force: true });
  }
});

blocked("blocks a dirty tree", (state) => write(state.repo, "dirty.txt", "dirty\n"), "candidate.clean");
blocked("blocks a remote integration mismatch", (state) => {
  write(state.repo, "remote.txt", "new\n"); git(state.repo, "add", "--all"); git(state.repo, "commit", "-m", "remote head"); git(state.repo, "push", "fixture", "ve/prototype-integration"); git(state.repo, "reset", "--hard", state.candidate);
}, "candidate.remote-integration");
blocked("blocks an open epoch", (state) => { state.controls.epoch.status = "open"; }, "epoch.frozen-or-closed");
blocked("blocks missing producer decisions", (state) => { state.controls.epoch.producers[0] = { thread: "T1", status: "unannounced", tag: null, decision: null }; }, "producers.terminal");
blocked("blocks an incomplete gate matrix", (state) => { state.controls.matrix.status = "dependency_blocked"; }, "gates.matrix-complete");
blocked("blocks a manifest that violates the non-authoritative contract", (state) => { state.controls.manifest.authoritative = true; }, "manifest.non-authoritative-contract");
blocked("blocks a dependency-blocked manifest", (state) => { state.controls.manifest.status = "dependency_blocked"; }, "manifest.status-ready");
blocked("blocks unresolved prototype success criteria", (state) => { state.controls.manifest.blockers.push({ id: "missing-evidence" }); }, "manifest.prototype-success");
blocked("blocks a missing result envelope", (state) => { state.controls.results = null; }, "gates.result-envelope");
blocked("blocks unavailable CI", (state) => { state.ci = null; }, "ci.available");
blocked("blocks failed CI", (state) => { state.ci[0].jobConclusion = "failure"; }, `ci.${REQUIRED_CI_CHECKS[0].job}`);
blocked("blocks CI with the wrong result artifact digest", (state) => { state.ci[0].resultsDigest = "0".repeat(64); }, `ci.${REQUIRED_CI_CHECKS[0].job}`);
for (const [field, value] of [["workflowPath", ".github/workflows/spoof.yml"], ["event", "pull_request"], ["branch", "main"], ["repository", "attacker/virtengine"]]) {
  blocked(`blocks CI with the wrong ${field}`, (state) => { state.ci[0][field] = value; }, `ci.${REQUIRED_CI_CHECKS[0].job}`);
}
blocked("blocks an existing local tag without mutating it", (state) => { git(state.repo, "tag", TAG, state.candidate); }, "tag.local-absent");
blocked("reports an existing mismatched local tag", (state) => { git(state.repo, "tag", TAG, state.controls.manifest.source.payload_sha); }, "tag.local-target");
blocked("blocks an existing remote-only tag", (state) => {
  git(state.repo, "tag", TAG, state.candidate);
  git(state.repo, "push", "fixture", `refs/tags/${TAG}`);
  git(state.repo, "tag", "--delete", TAG);
}, "tag.remote-absent");
blocked("blocks a strict manifest validation failure", (state) => { state.controls.manifestValid = false; }, "manifest.strict-valid");
blocked("blocks a missing manifest ledger hash", (state) => { delete state.controls.ledger.generated_hashes["_docs/ralph/prototype-integration/core-rc-manifest.json"]; }, "manifest.recorded-hash");
blocked("blocks fabricated result JSON without a committed hash", (state) => { delete state.controls.ledger.generated_hashes["_docs/ralph/prototype-integration/required-gate-results.json"]; }, "gates.result-envelope");

test("rejects invalid arguments and publication mode", () => {
  assert.throws(() => parseArgs(["--candidate", "HEAD", "--epoch", "1", "--tag", TAG]));
  assert.throws(() => parseArgs(["--publish"]), /unavailable/);
  assert.throws(() => parseArgs(["--candidate", "0".repeat(40), "--epoch", "0", "--tag", TAG]));
  assert.throws(() => parseArgs(["--candidate", "0".repeat(40), "--epoch", "1", "--tag", "wrong"]));
});

test("requires the requested publication epoch to be current and contiguous", () => {
  const paths = [
    "_docs/ralph/prototype-integration/epochs/epoch-1.json",
    "_docs/ralph/prototype-integration/epochs/epoch-2.json",
  ];
  assert.doesNotThrow(() => validateCandidateEpochSelection(paths, 2));
  assert.throws(() => validateCandidateEpochSelection(paths, 1), /candidate current epoch is 2/);
  assert.throws(() => validateCandidateEpochSelection([paths[1]], 2), /contiguous from epoch 1/);
});

test("rejects a forged empty pass report", () => {
  const schema = JSON.parse(readFileSync(join(__dirname, "..", "_docs", "ralph", "prototype-integration", "core-rc-publication-preflight.schema.json"), "utf8"));
  assert.doesNotThrow(() => validateReportSchema(schema));
  assert.ok(schema.properties.checks.minItems > 0);
  assert.throws(() => validateReport({ schema_version: "virtengine.core-rc-publication-preflight/v1", candidate: "0".repeat(40), epoch: 1, tag: TAG, passed: true, checks: [], blockers: [] }), /missing required checks/);
});