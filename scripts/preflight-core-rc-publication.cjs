#!/usr/bin/env node

"use strict";

const assert = require("assert").strict;
const { createHash } = require("crypto");
const { execFileSync } = require("child_process");
const { mkdtempSync, rmSync } = require("fs");
const { tmpdir } = require("os");
const { join, resolve } = require("path");
const { validateManifest } = require("./generate-core-rc-manifest.cjs");
const { validateSchema: validateManifestSchema } = require("./validate-core-rc-manifest.cjs");
const { validateIntake } = require("./validate-prototype-intake.cjs");
const { validateRequiredGateMatrix } = require("./validate-required-gate-matrix.cjs");
const { buildExecutionPlan, validateResultEnvelope } = require("./run-required-gates.cjs");

const REPORT_VERSION = "virtengine.core-rc-publication-preflight/v1";
const TAG = "checkpoint/prototype-integration/t4-09a";
const SHA_PATTERN = /^[a-f0-9]{40}$/;
const REQUIRED_CI_CHECKS = [
  { workflow: "Core RC Required Gates", workflowPath: ".github/workflows/core-rc-required-gates.yml", job: "core-rc-required-gates", expectsResults: true },
  { workflow: "Prototype Integration Controls", workflowPath: ".github/workflows/prototype-integration-controls.yml", job: "prototype-integration-controls", expectsResults: false },
];
const EXPECTED_REPOSITORY = "virtengine/virtengine";
const EXPECTED_BRANCH = "ve/prototype-integration";
const ALLOWED_CI_EVENTS = new Set(["push", "workflow_dispatch"]);
const REQUIRED_REPORT_CHECK_IDS = [
  "candidate.clean", "candidate.head", "candidate.remote-integration", "ci.available", "ci.core-rc-required-gates", "ci.prototype-integration-controls",
  "controls.available", "epoch.cutoff-elapsed", "epoch.frozen-or-closed", "evidence.model", "evidence.producers", "evidence.release-provenance",
  "evidence.rollback", "evidence.rollout", "evidence.sbom", "evidence.slurm", "gates.matrix-complete", "gates.result-envelope", "manifest.ancestry",
  "manifest.candidate-boundary", "manifest.non-authoritative-contract", "manifest.prototype-success", "manifest.recorded-hash", "manifest.status-ready", "manifest.strict-valid",
  "producers.accepted-ledger-correspondence", "producers.accepted-tags-revalidated", "producers.terminal", "producers.unannounced-frozen-out", "tag.local-absent", "tag.remote-absent",
].sort();
const ALLOWED_REPORT_CHECK_IDS = [...REQUIRED_REPORT_CHECK_IDS, "tag.local-target", "tag.remote-target"].sort();
const RESULT_PATH = "_docs/ralph/prototype-integration/required-gate-results.json";
const ALLOWED_BOUNDARY_PATHS = [
  "_docs/INDEX.md",
  "_docs/prototype-thread-intake-runbook.md",
  "_docs/ralph/handoffs/prototype-integration/",
  "_docs/ralph/prototype-integration/core-rc-manifest.json",
  "_docs/ralph/prototype-integration/core-rc-manifest.schema.json",
  "_docs/ralph/prototype-integration/core-rc-publication-preflight.schema.json",
  "_docs/ralph/prototype-integration/epochs/",
  "_docs/ralph/prototype-integration/evidence/",
  "scripts/generate-core-rc-manifest.cjs",
  "scripts/preflight-core-rc-publication.cjs",
  "scripts/preflight-core-rc-publication.test.cjs",
  "scripts/validate-core-rc-manifest.cjs",
  "scripts/validate-prototype-intake.cjs",
  "scripts/validate-prototype-intake.test.cjs",
  "scripts/validate-prototype-integration.cjs",
  "scripts/validate-prototype-integration.test.cjs",
];

function git(repo, args, allowFailure = false) {
  try {
    return execFileSync("git", ["-C", repo, ...args], { encoding: "utf8", stdio: ["ignore", "pipe", "pipe"] }).trim();
  } catch (error) {
    if (allowFailure) return null;
    const detail = error.stderr ? error.stderr.toString().trim() : error.message;
    throw new Error(`git ${args.join(" ")} failed: ${detail}`);
  }
}

function gitJson(repo, revision, path, optional = false) {
  const text = git(repo, ["show", `${revision}:${path}`], optional);
  if (text === null) return null;
  try {
    return JSON.parse(text);
  } catch {
    throw new Error(`${path} is not valid JSON`);
  }
}

function gitFile(repo, revision, path) {
  return execFileSync("git", ["-C", repo, "show", `${revision}:${path}`], { stdio: ["ignore", "pipe", "pipe"] });
}

function sha256(value) {
  return createHash("sha256").update(value).digest("hex");
}

function exactKeys(value, expected, label) {
  assert.ok(value && typeof value === "object" && !Array.isArray(value), `${label} must be an object`);
  assert.deepEqual(Object.keys(value).sort(), [...expected].sort(), `${label} has unknown or missing fields`);
}

function validateReport(report) {
  exactKeys(report, ["schema_version", "candidate", "epoch", "tag", "passed", "checks", "blockers"], "report");
  assert.equal(report.schema_version, REPORT_VERSION);
  assert.match(report.candidate, SHA_PATTERN);
  assert.ok(Number.isInteger(report.epoch) && report.epoch > 0);
  assert.equal(report.tag, TAG);
  assert.equal(typeof report.passed, "boolean");
  assert.ok(Array.isArray(report.checks));
  assert.ok(Array.isArray(report.blockers));
  assert.equal(new Set(report.checks.map((check) => check.id)).size, report.checks.length, "check IDs must be unique");
  assert.ok(REQUIRED_REPORT_CHECK_IDS.every((id) => report.checks.some((check) => check.id === id)), "report is missing required checks");
  assert.ok(report.checks.every((check) => ALLOWED_REPORT_CHECK_IDS.includes(check.id)), "report contains an unknown check ID");
  assert.deepEqual(report.checks.map((check) => check.id), report.checks.map((check) => check.id).sort());
  assert.deepEqual(report.blockers, [...report.blockers].sort());
  assert.equal(new Set(report.blockers).size, report.blockers.length);
  for (const check of report.checks) {
    exactKeys(check, ["id", "passed", "detail"], `check ${check.id || "<missing>"}`);
    assert.match(check.id, /^[a-z0-9.-]+$/);
    assert.equal(typeof check.passed, "boolean");
    assert.ok(typeof check.detail === "string" && check.detail.length > 0);
  }
  assert.deepEqual(report.blockers, report.checks.filter((check) => !check.passed).map((check) => check.id).sort());
  assert.equal(report.passed, report.blockers.length === 0);
  return true;
}

function validateReportSchema(schema) {
  exactKeys(schema, ["$schema", "$id", "title", "type", "additionalProperties", "required", "properties", "$defs"], "report schema");
  assert.equal(schema.$schema, "https://json-schema.org/draft/2020-12/schema");
  assert.equal(schema.type, "object");
  assert.equal(schema.additionalProperties, false);
  assert.deepEqual([...schema.required].sort(), ["schema_version", "candidate", "epoch", "tag", "passed", "checks", "blockers"].sort());
  assert.equal(schema.properties.schema_version.const, REPORT_VERSION);
  assert.equal(schema.properties.tag.const, TAG);
  assert.equal(schema.properties.epoch.minimum, 1);
  assert.equal(schema.properties.checks.minItems, REQUIRED_REPORT_CHECK_IDS.length);
  assert.equal(schema.properties.checks.maxItems, ALLOWED_REPORT_CHECK_IDS.length);
  assert.equal(schema.properties.blockers.uniqueItems, true);
  assert.deepEqual([...schema.$defs.id.enum].sort(), ALLOWED_REPORT_CHECK_IDS);
  assert.equal(schema.$defs.check.additionalProperties, false);
  assert.deepEqual([...schema.$defs.check.required].sort(), ["id", "passed", "detail"].sort());
  return true;
}

function loadCandidateControls(repo, candidate, epochNumber) {
  const prefix = "_docs/ralph/prototype-integration";
  const epochPaths = git(repo, ["ls-tree", "-r", "--name-only", candidate, `${prefix}/epochs`]).split(/\r?\n/).filter((path) => /\/epoch-[1-9][0-9]*\.json$/.test(path));
  validateCandidateEpochSelection(epochPaths, epochNumber);
  return {
    epoch: gitJson(repo, candidate, `${prefix}/epochs/epoch-${epochNumber}.json`),
    ledger: gitJson(repo, candidate, "_docs/ralph/handoffs/prototype-integration/HANDOFF.yaml"),
    manifest: gitJson(repo, candidate, `${prefix}/core-rc-manifest.json`),
    manifestSchema: gitJson(repo, candidate, `${prefix}/core-rc-manifest.schema.json`),
    matrix: gitJson(repo, candidate, `${prefix}/required-gate-matrix.json`),
    matrixSchema: gitJson(repo, candidate, `${prefix}/required-gate-matrix.schema.json`),
    planSchema: gitJson(repo, candidate, `${prefix}/required-gate-plan.schema.json`),
    resultSchema: gitJson(repo, candidate, `${prefix}/required-gate-results.schema.json`),
    results: gitJson(repo, candidate, RESULT_PATH, true),
    manifestText: gitFile(repo, candidate, `${prefix}/core-rc-manifest.json`),
    matrixText: gitFile(repo, candidate, `${prefix}/required-gate-matrix.json`),
  };
}

function validateCandidateEpochSelection(paths, requestedEpoch) {
  const numbers = paths.map((path) => Number(path.match(/\/epoch-([1-9][0-9]*)\.json$/)?.[1])).sort((left, right) => left - right);
  assert.ok(numbers.length > 0 && numbers.every((number, index) => number === index + 1), "candidate intake epoch history must be contiguous from epoch 1");
  assert.equal(Number(requestedEpoch), numbers.at(-1), `requested epoch ${requestedEpoch} is stale; candidate current epoch is ${numbers.at(-1)}`);
  return true;
}

function defaultCiProvider({ repo, candidate, remote }) {
  const url = git(repo, ["remote", "get-url", remote]);
  const match = /github\.com[/:]([^/]+)\/([^/.]+)(?:\.git)?$/.exec(url);
  assert.ok(match, "remote is not a GitHub repository");
  execFileSync("gh", ["auth", "status"], { stdio: "ignore" });
  const apiRepository = `${match[1]}/${match[2]}`;
  const repository = apiRepository.toLowerCase();
  assert.equal(repository, EXPECTED_REPOSITORY, "remote repository identity is not publication-authorized");
  const workflows = JSON.parse(execFileSync("gh", ["api", "--method", "GET", `repos/${apiRepository}/actions/workflows`, "-f", "per_page=100"], { encoding: "utf8" })).workflows;
  return REQUIRED_CI_CHECKS.flatMap((required) => {
    const workflow = workflows.find((entry) => entry.path === required.workflowPath);
    assert.ok(workflow, `required workflow path is missing: ${required.workflowPath}`);
    const runs = JSON.parse(execFileSync("gh", ["api", "--method", "GET", `repos/${apiRepository}/actions/workflows/${workflow.id}/runs`, "-f", `head_sha=${candidate}`, "-f", `branch=${EXPECTED_BRANCH}`, "-f", "per_page=100"], { encoding: "utf8" })).workflow_runs;
    return runs.flatMap((run) => {
    const jobs = JSON.parse(execFileSync("gh", ["api", "--method", "GET", `repos/${match[1]}/${match[2]}/actions/runs/${run.id}/jobs`, "-f", "per_page=100"], {
      encoding: "utf8",
      stdio: ["ignore", "pipe", "pipe"],
    })).jobs;
      let resultsDigest;
      if (required.expectsResults) {
        const artifacts = JSON.parse(execFileSync("gh", ["api", "--method", "GET", `repos/${apiRepository}/actions/runs/${run.id}/artifacts`, "-f", "per_page=100"], { encoding: "utf8" })).artifacts;
        const artifact = artifacts.find((entry) => entry.name === "required-gate-results" && entry.expired === false);
        resultsDigest = artifact && typeof artifact.digest === "string" ? artifact.digest.replace(/^sha256:/, "") : null;
      }
      return jobs.map((job) => ({ repository, workflow: run.name, workflowId: workflow.id, workflowPath: workflow.path, runId: run.id, runAttempt: run.run_attempt, runConclusion: run.conclusion, event: run.event, branch: run.head_branch, job: job.name, jobConclusion: job.conclusion, sha: run.head_sha, resultsDigest }));
    });
  });
}

function defaultIntakeProvider({ repo, candidate, epoch, tag, remote }) {
  const directory = mkdtempSync(join(tmpdir(), "core-rc-intake-revalidation-"));
  const clone = join(directory, "repo");
  try {
    const remoteUrl = git(repo, ["remote", "get-url", remote]);
    execFileSync("git", ["clone", "--no-checkout", "--no-local", repo, clone], { stdio: "ignore" });
    git(clone, ["checkout", "--detach", candidate]);
    git(clone, ["remote", "add", "publication-source", remoteUrl]);
    return validateIntake({ epoch, tag, repo: clone, remote: "publication-source", revalidateAccepted: true });
  } finally {
    rmSync(directory, { recursive: true, force: true });
  }
}

function boundaryAllowed(path) {
  return ALLOWED_BOUNDARY_PATHS.some((allowed) => allowed.endsWith("/") ? path.startsWith(allowed) : path === allowed);
}

function addCheck(checks, id, passed, detail) {
  checks.push({ id, passed: Boolean(passed), detail });
}

async function preflight(options) {
  const repo = resolve(options.repo || resolve(__dirname, ".."));
  const remote = options.remote || "origin";
  const candidate = options.candidate;
  const epochNumber = Number(options.epoch);
  const tag = options.tag;
  assert.match(candidate || "", SHA_PATTERN, "--candidate must be an exact lowercase 40-character SHA");
  assert.ok(Number.isInteger(epochNumber) && epochNumber > 0, "--epoch must be a positive integer");
  assert.equal(tag, TAG, `--tag must be exactly ${TAG}`);

  const checks = [];
  const head = git(repo, ["rev-parse", "HEAD"]);
  addCheck(checks, "candidate.clean", git(repo, ["status", "--porcelain"]) === "", "local worktree must be clean");
  addCheck(checks, "candidate.head", head === candidate, "candidate must equal local HEAD");
  const remoteLine = git(repo, ["ls-remote", "--heads", remote, "refs/heads/ve/prototype-integration"], true);
  const remoteHead = remoteLine ? remoteLine.split(/\s+/)[0] : null;
  addCheck(checks, "candidate.remote-integration", remoteHead === candidate, "candidate must equal the remote integration branch head");

  let controls;
  try {
    controls = options.controls || loadCandidateControls(repo, candidate, epochNumber);
    addCheck(checks, "controls.available", true, "candidate control documents are available");
  } catch (error) {
    addCheck(checks, "controls.available", false, error.message);
    controls = {};
  }

  const epoch = controls.epoch;
  addCheck(checks, "epoch.cutoff-elapsed", Boolean(epoch && Number.isFinite(Date.parse(epoch.announcement_cutoff)) && Date.now() > Date.parse(epoch.announcement_cutoff)), "epoch announcement cutoff must have elapsed");
  addCheck(checks, "epoch.frozen-or-closed", Boolean(epoch && ["frozen", "closed"].includes(epoch.status)), "epoch status must be frozen or closed");
  const producers = epoch && Array.isArray(epoch.producers) ? epoch.producers : [];
  const terminal = producers.length > 0 && producers.every((producer) => ["accepted", "rejected"].includes(producer.status) || (producer.status === "unannounced" && producer.decision === "frozen-out"));
  addCheck(checks, "producers.terminal", terminal, "all producers must be accepted, rejected, or explicitly frozen-out");
  const explicitFrozenOut = producers.filter((producer) => producer.status === "unannounced").every((producer) => producer.decision === "frozen-out");
  addCheck(checks, "producers.unannounced-frozen-out", explicitFrozenOut, "every unannounced producer must record decision=frozen-out");
  const accepted = producers.filter((producer) => producer.status === "accepted");
  const acceptedLedger = controls.ledger && Array.isArray(controls.ledger.accepted_checkpoints) ? controls.ledger.accepted_checkpoints : [];
  const acceptedLedgerValid = accepted.every((producer) => acceptedLedger.some((entry) => entry.thread === producer.thread && entry.tag === producer.tag && SHA_PATTERN.test(entry.tip) && SHA_PATTERN.test(entry.payload_head)));
  addCheck(checks, "producers.accepted-ledger-correspondence", acceptedLedgerValid, "every accepted epoch producer must match an accepted ledger checkpoint, tag, tip, and payload");
  let intakeValid = accepted.length > 0;
  for (const producer of accepted) {
    try {
      assert.equal(typeof producer.tag, "string");
      const validator = options.intakeProvider || defaultIntakeProvider;
      await validator({ epoch: epochNumber, tag: producer.tag, repo, remote, candidate });
    } catch {
      intakeValid = false;
    }
  }
  addCheck(checks, "producers.accepted-tags-revalidated", intakeValid, "accepted producer tags must pass the intake validator");

  let manifestValid = false;
  try {
    assert.ok(controls.manifest && controls.manifestSchema);
    if (!options.controls) {
      validateManifestSchema(controls.manifestSchema);
      validateManifest(controls.manifest, { rootDir: repo });
    } else assert.equal(await options.manifestValidator(controls.manifest, controls.manifestSchema), true, "injected manifest validator rejected fixture");
    manifestValid = true;
  } catch {
    manifestValid = false;
  }
  addCheck(checks, "manifest.strict-valid", manifestValid, "core RC manifest and strict schema must validate");
  const manifest = controls.manifest || {};
  let ancestryValid = false;
  try {
    const source = manifest.source.payload_sha;
    const tooling = manifest.tooling.tooling_sha;
    assert.match(source, SHA_PATTERN);
    assert.match(tooling, SHA_PATTERN);
    assert.equal(git(repo, ["merge-base", "--is-ancestor", source, tooling], true), "");
    assert.equal(git(repo, ["merge-base", "--is-ancestor", tooling, candidate], true), "");
    ancestryValid = true;
  } catch {
    ancestryValid = false;
  }
  addCheck(checks, "manifest.ancestry", ancestryValid, "manifest source must be an ancestor of tooling and tooling an ancestor of candidate");
  let boundaryValid = false;
  try {
    const source = manifest.source.payload_sha;
    assert.match(source, SHA_PATTERN);
    assert.equal(git(repo, ["merge-base", "--is-ancestor", source, candidate], true), "");
    const changed = git(repo, ["diff", "--name-only", `${source}..${candidate}`]).split(/\r?\n/).filter(Boolean);
    assert.ok(changed.every(boundaryAllowed));
    boundaryValid = true;
  } catch {
    boundaryValid = false;
  }
  addCheck(checks, "manifest.candidate-boundary", boundaryValid, "candidate delta from manifest source may contain only declared tooling and evidence paths");
  const recordedHash = controls.ledger && controls.ledger.generated_hashes && controls.ledger.generated_hashes["_docs/ralph/prototype-integration/core-rc-manifest.json"];
  addCheck(checks, "manifest.recorded-hash", Boolean(recordedHash && sha256(controls.manifestText || "") === recordedHash), "candidate manifest hash must be present in the ledger and match committed bytes");
  const nonAuthoritativeContract = manifest.authoritative === false && manifest.planned_functionality_complete === false && manifest.milestone_m_eligible === false;
  addCheck(checks, "manifest.non-authoritative-contract", nonAuthoritativeContract, "the current prototype manifest schema requires all publication authority flags to remain false");
  const manifestReady = manifest.status !== "dependency_blocked";
  addCheck(checks, "manifest.status-ready", manifestReady, "publication requires a future manifest version whose status is ready");
  const prototypeSuccess = Array.isArray(manifest.blockers) && manifest.blockers.length === 0;
  addCheck(checks, "manifest.prototype-success", prototypeSuccess, "publication requires a future manifest with no declared prototype success blockers");

  let matrixValid = false;
  try {
    if (!options.controls) validateRequiredGateMatrix(controls.matrix, { schema: controls.matrixSchema, planSchema: controls.planSchema, resultSchema: controls.resultSchema });
    else assert.notEqual(controls.matrixValid, false);
    matrixValid = controls.matrix.status === "complete" && controls.matrix.completion_claim === true && controls.matrix.categories.every((category) => category.status === "complete");
  } catch {
    matrixValid = false;
  }
  addCheck(checks, "gates.matrix-complete", matrixValid, "required gate matrix must be strictly valid and complete");
  let envelopeValid = false;
  let resultsDigest = null;
  try {
    const results = controls.results;
    assert.ok(results, "result envelope is missing");
    const base = manifest.required_gates.base_sha || manifest.source.payload_sha;
    const plan = options.planProvider ? await options.planProvider({ repo, base, candidate, controls }) : buildExecutionPlan({ repoDir: repo, base, head: candidate, matrixPath: resolve(repo, manifest.required_gates.matrix_path) });
    validateResultEnvelope(plan, results);
    const resultsText = controls.resultsText || gitFile(repo, candidate, RESULT_PATH);
    resultsDigest = sha256(resultsText);
    const ledgerResultHash = controls.ledger && controls.ledger.generated_hashes && controls.ledger.generated_hashes[RESULT_PATH];
    const manifestResult = Array.isArray(manifest.control_artifacts) && manifest.control_artifacts.find((entry) => entry.path === RESULT_PATH);
    assert.equal(ledgerResultHash, resultsDigest, "result envelope ledger hash mismatch");
    assert.equal(manifestResult && manifestResult.sha256, resultsDigest, "result envelope manifest hash mismatch");
    envelopeValid = true;
  } catch {
    envelopeValid = false;
  }
  addCheck(checks, "gates.result-envelope", envelopeValid, "computed base-to-candidate plan and committed result envelope must match exactly and be hash-bound in manifest and ledger");

  const availability = options.controls ? controls.availability : {
    rollout: manifest.rollout && manifest.rollout.status === "complete",
    rollback: manifest.rollback && manifest.rollback.status === "verified",
    sbom: manifest.model_provenance && manifest.model_provenance.sbom_status === "available",
    releaseProvenance: Array.isArray(manifest.external_dependencies) && manifest.external_dependencies.some((entry) => entry.id === "sbom-and-release-provenance" && entry.status === "available"),
    model: manifest.model_provenance && manifest.model_provenance.status === "complete" && manifest.model_provenance.production_weights_status === "available",
    slurm: manifest.slurm && manifest.slurm.status === "complete" && manifest.slurm.report_status === "passed",
    producers: manifest.producer_checkpoints && manifest.producer_checkpoints.accepted.length > 0,
  };
  for (const id of ["model", "producers", "releaseProvenance", "rollback", "rollout", "sbom", "slurm"]) {
    addCheck(checks, `evidence.${id.replace(/[A-Z]/g, (letter) => `-${letter.toLowerCase()}`)}`, Boolean(availability && availability[id]), `${id} publication evidence must be available`);
  }

  let ciChecks = null;
  try {
    ciChecks = await (options.ciProvider || defaultCiProvider)({ repo, candidate, remote });
  } catch {
    ciChecks = null;
  }
  addCheck(checks, "ci.available", Array.isArray(ciChecks), "exact-SHA CI evidence provider must be available");
  for (const required of REQUIRED_CI_CHECKS) {
    const evidence = Array.isArray(ciChecks) ? ciChecks.find((check) => check.repository === EXPECTED_REPOSITORY && check.workflow === required.workflow && check.workflowPath === required.workflowPath && Number.isSafeInteger(check.workflowId) && Number.isSafeInteger(check.runId) && Number.isSafeInteger(check.runAttempt) && check.runAttempt > 0 && check.event && ALLOWED_CI_EVENTS.has(check.event) && check.branch === EXPECTED_BRANCH && check.job === required.job && check.sha === candidate) : null;
    const digestMatches = !required.expectsResults || Boolean(evidence && evidence.resultsDigest === resultsDigest);
    addCheck(checks, `ci.${required.job}`, Boolean(evidence && evidence.runConclusion === "success" && evidence.jobConclusion === "success" && digestMatches), `${required.workflowPath} / ${required.job} must conclude success for the exact repository, branch, event, attempt, candidate, and result digest`);
  }

  const localTarget = git(repo, ["rev-parse", `refs/tags/${tag}^{commit}`], true);
  addCheck(checks, "tag.local-absent", localTarget === null, "publication tag must not exist locally");
  if (localTarget !== null) addCheck(checks, "tag.local-target", localTarget === candidate, "existing local tag must target the exact candidate");
  const remoteTagLine = git(repo, ["ls-remote", "--tags", remote, `refs/tags/${tag}`, `refs/tags/${tag}^{}`], true);
  const remoteTargets = remoteTagLine ? remoteTagLine.split(/\r?\n/).filter(Boolean).map((line) => line.split(/\s+/)[0]) : [];
  addCheck(checks, "tag.remote-absent", remoteTargets.length === 0, "publication tag must not exist remotely");
  if (remoteTargets.length > 0) addCheck(checks, "tag.remote-target", remoteTargets.includes(candidate), "existing remote tag must peel to the exact candidate");

  checks.sort((left, right) => left.id < right.id ? -1 : left.id > right.id ? 1 : 0);
  const blockers = checks.filter((check) => !check.passed).map((check) => check.id).sort();
  const report = { schema_version: REPORT_VERSION, candidate, epoch: epochNumber, tag, passed: blockers.length === 0, checks, blockers };
  validateReport(report);
  return report;
}

function parseArgs(argv) {
  const options = { json: false };
  const seen = new Set();
  for (let index = 0; index < argv.length; index += 1) {
    const key = argv[index];
    assert.equal(seen.has(key), false, `duplicate argument: ${key}`);
    seen.add(key);
    if (key === "--json") options.json = true;
    else if (key === "--publish") throw new Error("--publish is unavailable; T4-09A is diagnostic-only");
    else if (["--candidate", "--epoch", "--tag", "--repo", "--remote"].includes(key)) {
      const value = argv[++index];
      assert.ok(value && !value.startsWith("--"), `${key} requires a value`);
      options[key.slice(2)] = value;
    } else throw new Error(`unknown argument: ${key || "<none>"}`);
  }
  assert.ok(options.candidate, "--candidate is required");
  assert.ok(options.epoch, "--epoch is required");
  assert.ok(options.tag, "--tag is required");
  assert.match(options.candidate, SHA_PATTERN, "--candidate must be an exact lowercase 40-character SHA");
  assert.ok(Number.isInteger(Number(options.epoch)) && Number(options.epoch) > 0, "--epoch must be a positive integer");
  assert.equal(options.tag, TAG, `--tag must be exactly ${TAG}`);
  return options;
}

async function main(argv = process.argv.slice(2)) {
  try {
    const options = parseArgs(argv);
    const report = await preflight(options);
    if (options.json) console.log(JSON.stringify(report, null, 2));
    else console.log(`core RC publication preflight: ${report.passed ? "passed" : "blocked"}: ${report.blockers.join(", ") || "none"}`);
    if (!report.passed) process.exitCode = 1;
  } catch (error) {
    console.error(`core RC publication preflight: invalid: ${error.message}`);
    process.exitCode = 2;
  }
}

module.exports = { ALLOWED_REPORT_CHECK_IDS, REPORT_VERSION, REQUIRED_CI_CHECKS, REQUIRED_REPORT_CHECK_IDS, TAG, parseArgs, preflight, validateCandidateEpochSelection, validateReport, validateReportSchema };

if (require.main === module) main();