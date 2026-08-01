#!/usr/bin/env node

"use strict";

const assert = require("assert").strict;
const { createHash } = require("crypto");
const { readFileSync } = require("fs");
const { resolve } = require("path");
const { spawnSync } = require("child_process");
const { validateRequiredGateMatrix } = require("./validate-required-gate-matrix.cjs");

const root = resolve(__dirname, "..");
const defaultMatrixPath = resolve(root, "_docs/ralph/prototype-integration/required-gate-matrix.json");
const planSchemaVersion = "virtengine.task-88b.required-gate-plan/v1";
const resultSchemaVersion = "virtengine.task-88b.required-gate-results/v1";
const shaPattern = /^[a-f0-9]{40}$/;

function exactKeys(value, expected, label) {
  assert.ok(value && typeof value === "object" && !Array.isArray(value), `${label} must be an object`);
  assert.deepEqual(Object.keys(value).sort(), [...expected].sort(), `${label} has unexpected or missing properties`);
}

function globToRegExp(selector) {
  let pattern = "^";
  for (let index = 0; index < selector.length; index += 1) {
    const character = selector[index];
    if (character === "*" && selector[index + 1] === "*") {
      index += 1;
      if (selector[index + 1] === "/") {
        index += 1;
        pattern += "(?:.*/)?";
      } else {
        pattern += ".*";
      }
    } else if (character === "*") {
      pattern += "[^/]*";
    } else if (character === "?") {
      pattern += "[^/]";
    } else {
      pattern += character.replace(/[\\^$+?.()|{}[\]]/g, "\\$&");
    }
  }
  return new RegExp(`${pattern}$`);
}

function matchesSelector(path, selector) {
  return globToRegExp(selector).test(path);
}

function runGit(repoDir, args, options = {}) {
  const result = spawnSync("git", ["-C", repoDir, ...args], {
    encoding: options.encoding === null ? null : "utf8",
    maxBuffer: 64 * 1024 * 1024,
  });
  if (result.error) throw result.error;
  assert.equal(result.status, 0, `git ${args.join(" ")} failed: ${String(result.stderr).trim()}`);
  return result.stdout;
}

function resolveCommit(repoDir, revision) {
  const resolved = runGit(repoDir, ["rev-parse", "--verify", `${revision}^{commit}`]).trim();
  assert.match(resolved, shaPattern, `revision did not resolve to a full commit SHA: ${revision}`);
  return resolved;
}

  function isAncestor(repoDir, baseSha, headSha) {
    const result = spawnSync("git", ["-C", repoDir, "merge-base", "--is-ancestor", baseSha, headSha]);
    if (result.error) throw result.error;
    return result.status === 0;
  }

function nulPaths(output) {
  return output.toString("utf8").split("\0").filter(Boolean).map((path) => path.replaceAll("\\", "/"));
}

function collectChangedPaths(repoDir, baseSha, headSha) {
  const paths = new Set(nulPaths(runGit(repoDir, ["diff", "--name-only", "-z", baseSha, headSha, "--"], { encoding: null })));
  const commits = runGit(repoDir, ["rev-list", "--reverse", "--topo-order", `${baseSha}..${headSha}`]).trim().split(/\r?\n/).filter(Boolean);
  for (const commit of commits) {
    const parentChanges = runGit(repoDir, ["diff-tree", "--root", "-m", "--no-commit-id", "--name-only", "-r", "-z", commit, "--"], { encoding: null });
    for (const path of nulPaths(parentChanges)) paths.add(path);
  }
  return [...paths].sort();
}

function matrixDigest(matrixBytes) {
  return createHash("sha256").update(matrixBytes).digest("hex");
}

function buildExecutionPlan({ repoDir = root, base, head, matrixPath = defaultMatrixPath }) {
  const bytes = readFileSync(matrixPath);
  const matrix = JSON.parse(bytes.toString("utf8"));
  validateRequiredGateMatrix(matrix);
  const baseSha = resolveCommit(repoDir, base);
  const headSha = resolveCommit(repoDir, head);
  assert.ok(isAncestor(repoDir, baseSha, headSha), "base must be an ancestor of head");
  const paths = collectChangedPaths(repoDir, baseSha, headSha);
  const selectedCategories = [];
  const matchedPaths = new Set();

  for (const category of matrix.categories) {
    const categoryPaths = paths.filter((path) => category.path_selectors.some((selector) => matchesSelector(path, selector)));
    if (categoryPaths.length === 0) continue;
    for (const path of categoryPaths) matchedPaths.add(path);
    selectedCategories.push({
      id: category.id,
      status: category.status,
      matched_paths: categoryPaths,
      matched_selectors: category.path_selectors.filter((selector) => categoryPaths.some((path) => matchesSelector(path, selector))),
      commands: category.required_commands.map((command) => ({ id: command.id, kind: command.kind, command: command.command })),
      pinned_tools: category.pinned_tools.map((tool) => ({ name: tool.name, version: tool.version, source: tool.source })),
    });
  }

  const allowlistedPaths = [];
  const unmatchedPaths = [];
  for (const path of paths.filter((candidate) => !matchedPaths.has(candidate))) {
    const allowance = matrix.unmatched_path_allowlist.find((entry) => matchesSelector(path, entry.selector));
    if (allowance) allowlistedPaths.push({ path, selector: allowance.selector, reason: allowance.reason });
    else unmatchedPaths.push(path);
  }
  assert.deepEqual(unmatchedPaths, [], `changed paths have no required-gate category or metadata allowlist: ${unmatchedPaths.join(", ")}`);

  return {
    schema_version: planSchemaVersion,
    base_sha: baseSha,
    head_sha: headSha,
    matrix_digest: matrixDigest(bytes),
    matrix_status: matrix.status,
    changed_paths: paths,
    allowlisted_paths: allowlistedPaths,
    categories: selectedCategories,
  };
}

function validateResultEnvelope(plan, envelope) {
  exactKeys(envelope, ["schema_version", "base_sha", "head_sha", "matrix_digest", "results"], "result envelope");
  assert.equal(envelope.schema_version, resultSchemaVersion, "result schema version mismatch");
  assert.equal(envelope.base_sha, plan.base_sha, "result base SHA mismatch");
  assert.equal(envelope.head_sha, plan.head_sha, "result head SHA mismatch");
  assert.equal(envelope.matrix_digest, plan.matrix_digest, "result matrix digest mismatch");
  assert.ok(Array.isArray(envelope.results), "results must be an array");

  const expected = new Map();
  for (const category of plan.categories) {
    for (const command of category.commands) expected.set(`${category.id}\0${command.id}`, { category, command });
  }
  const seen = new Set();
  for (const result of envelope.results) {
    exactKeys(result, ["category_id", "command_id", "command", "outcome", "exit_code", "discovered_tests", "executed_tests", "skipped_tests", "tools"], "gate result");
    const key = `${result.category_id}\0${result.command_id}`;
    assert.ok(!seen.has(key), `duplicate gate result: ${result.category_id}.${result.command_id}`);
    seen.add(key);
    const required = expected.get(key);
    assert.ok(required, `extra gate result: ${result.category_id}.${result.command_id}`);
    assert.equal(result.command, required.command.command, `literal command mismatch: ${result.category_id}.${result.command_id}`);
    assert.equal(result.outcome, "passed", `gate result must pass: ${result.category_id}.${result.command_id}`);
    assert.equal(result.exit_code, 0, `gate result exit code must be zero: ${result.category_id}.${result.command_id}`);
    for (const field of ["discovered_tests", "executed_tests", "skipped_tests"]) {
      assert.ok(Number.isSafeInteger(result[field]) && result[field] >= 0, `${field} must be a non-negative integer`);
    }
    assert.ok(Array.isArray(result.tools), "gate result tools must be an array");
    assert.equal(result.tools.length, required.category.pinned_tools.length, `pinned tool count mismatch: ${result.category_id}.${result.command_id}`);
    const actualTools = new Map();
    for (const tool of result.tools) {
      exactKeys(tool, ["name", "version", "available"], "gate result tool");
      assert.ok(!actualTools.has(tool.name), `duplicate result tool: ${tool.name}`);
      actualTools.set(tool.name, tool);
    }
    for (const pinned of required.category.pinned_tools) {
      const actual = actualTools.get(pinned.name);
      assert.ok(actual && actual.available === true, `missing pinned tool: ${pinned.name}`);
      assert.equal(actual.version, pinned.version, `pinned tool version mismatch: ${pinned.name}`);
    }
    if (required.command.kind === "test") {
      assert.ok(result.discovered_tests >= 1, `zero tests discovered: ${result.category_id}.${result.command_id}`);
      assert.ok(result.executed_tests >= 1, `zero tests executed: ${result.category_id}.${result.command_id}`);
      assert.equal(result.skipped_tests, 0, `skipped tests are not allowed: ${result.category_id}.${result.command_id}`);
      assert.equal(result.executed_tests, result.discovered_tests, `not all discovered tests executed: ${result.category_id}.${result.command_id}`);
    }
  }
  const missing = [...expected.keys()].filter((key) => !seen.has(key)).map((key) => key.replace("\0", "."));
  assert.deepEqual(missing, [], `missing gate results: ${missing.join(", ")}`);
  return true;
}

function assertExecutionReady(plan) {
  assert.ok(["ready", "complete"].includes(plan.matrix_status), `execution refused; matrix status is ${plan.matrix_status}`);
  const blocked = plan.categories.filter((category) => !["ready", "complete"].includes(category.status));
  assert.deepEqual(blocked.map((category) => category.id), [], `execution refused; selected categories are not ready or complete: ${blocked.map((category) => category.id).join(", ")}`);
}

function parseArguments(argv) {
  const options = { repoDir: root, matrixPath: defaultMatrixPath };
  for (let index = 0; index < argv.length; index += 1) {
    const argument = argv[index];
    if (["--base", "--head", "--repo", "--matrix", "--validate-results"].includes(argument)) {
      assert.ok(argv[index + 1], `${argument} requires a value`);
      const key = { "--base": "base", "--head": "head", "--repo": "repoDir", "--matrix": "matrixPath", "--validate-results": "resultsPath" }[argument];
      options[key] = resolve(argv[index + 1]);
      if (["base", "head"].includes(key)) options[key] = argv[index + 1];
      index += 1;
    } else if (argument === "--execute") options.execute = true;
    else throw new Error(`unknown argument: ${argument}`);
  }
  assert.ok(options.base, "--base is required");
  assert.ok(options.head, "--head is required");
  assert.ok(!(options.execute && options.resultsPath), "--execute and --validate-results are mutually exclusive");
  return options;
}

function main(argv) {
  const options = parseArguments(argv);
  const plan = buildExecutionPlan(options);
  if (options.resultsPath) {
    validateResultEnvelope(plan, JSON.parse(readFileSync(options.resultsPath, "utf8")));
    process.stdout.write(`${JSON.stringify({ valid: true, ...plan }, null, 2)}\n`);
    return;
  }
  if (options.execute) {
    assertExecutionReady(plan);
    throw new Error("execution is unavailable until result capture adapters are defined for the ready matrix");
  }
  process.stdout.write(`${JSON.stringify(plan, null, 2)}\n`);
}

module.exports = {
  assertExecutionReady,
  buildExecutionPlan,
  collectChangedPaths,
  globToRegExp,
  matchesSelector,
  validateResultEnvelope,
};

if (require.main === module) {
  try {
    main(process.argv.slice(2));
  } catch (error) {
    process.stderr.write(`required gate control failed: ${error.message}\n`);
    process.exitCode = 1;
  }
}