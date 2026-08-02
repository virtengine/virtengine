"use strict";

const assert = require("assert").strict;
const { readFileSync } = require("fs");
const { resolve } = require("path");
const { spawnSync } = require("child_process");
const { validateGateResult, validateRequiredGateMatrix } = require("./validate-required-gate-matrix.cjs");
const { matchesSelector } = require("./run-required-gates.cjs");

const root = resolve(__dirname, "..");
const matrix = JSON.parse(readFileSync(resolve(root, "_docs/ralph/prototype-integration/required-gate-matrix.json"), "utf8"));
const schema = JSON.parse(readFileSync(resolve(root, "_docs/ralph/prototype-integration/required-gate-matrix.schema.json"), "utf8"));

function cloneMatrix() {
  return JSON.parse(JSON.stringify(matrix));
}

function validResult(category) {
  const command = category.required_commands.find((entry) => entry.kind === "test");
  return {
    command_id: command.id,
    command: command.command,
    outcome: "passed",
    exit_code: 0,
    discovered_tests: 2,
    executed_tests: 2,
    skipped_tests: 0,
    tools: category.pinned_tools.map((tool) => ({ name: tool.name, version: tool.version, available: true })),
  };
}

const tests = [
  ["accepts the dependency-blocked matrix control", () => assert.doesNotThrow(() => validateRequiredGateMatrix(cloneMatrix(), { schema }))],
  ["rejects a missing category", () => {
    const candidate = cloneMatrix();
    candidate.categories.pop();
    assert.throws(() => validateRequiredGateMatrix(candidate, { schema }), /missing category/);
  }],
  ["rejects a missing required command", () => {
    const candidate = cloneMatrix();
    candidate.categories[0].required_commands.shift();
    assert.throws(() => validateRequiredGateMatrix(candidate, { schema }), /required literal commands changed/);
  }],
  ["rejects a missing Windows localnet selector", () => {
    const candidate = cloneMatrix();
    const category = candidate.categories.find((entry) => entry.id === "docs_process_boundary_e2e");
    category.path_selectors = category.path_selectors.filter((selector) => selector !== "scripts/localnet.ps1");
    assert.throws(() => validateRequiredGateMatrix(candidate, { schema }), /must include scripts\/localnet\.ps1/);
  }],
  ["rejects a missing shell localnet selector", () => {
    const candidate = cloneMatrix();
    const category = candidate.categories.find((entry) => entry.id === "docs_process_boundary_e2e");
    category.path_selectors = category.path_selectors.filter((selector) => selector !== "scripts/localnet.sh");
    assert.throws(() => validateRequiredGateMatrix(candidate, { schema }), /must include scripts\/localnet\.sh/);
  }],
  ["rejects missing root PowerShell coverage", () => {
    const candidate = cloneMatrix();
    const category = candidate.categories.find((entry) => entry.id === "docs_process_boundary_e2e");
    category.path_selectors = category.path_selectors.filter((selector) => selector !== "scripts/*.ps1");
    assert.throws(() => validateRequiredGateMatrix(candidate, { schema }), /must include scripts\/\*\.ps1/);
  }],
  ["rejects missing root shell coverage", () => {
    const candidate = cloneMatrix();
    const category = candidate.categories.find((entry) => entry.id === "docs_process_boundary_e2e");
    category.path_selectors = category.path_selectors.filter((selector) => selector !== "scripts/*.sh");
    assert.throws(() => validateRequiredGateMatrix(candidate, { schema }), /must include scripts\/\*\.sh/);
  }],
  ["rejects missing root Python coverage", () => {
    const candidate = cloneMatrix();
    const category = candidate.categories.find((entry) => entry.id === "docs_process_boundary_e2e");
    category.path_selectors = category.path_selectors.filter((selector) => selector !== "scripts/*.py");
    assert.throws(() => validateRequiredGateMatrix(candidate, { schema }), /must include scripts\/\*\.py/);
  }],
  ["rejects missing root SQL coverage", () => {
    const candidate = cloneMatrix();
    const category = candidate.categories.find((entry) => entry.id === "docs_process_boundary_e2e");
    category.path_selectors = category.path_selectors.filter((selector) => selector !== "scripts/*.sql");
    assert.throws(() => validateRequiredGateMatrix(candidate, { schema }), /must include scripts\/\*\.sql/);
  }],
  ["covers every tracked root scripts entry", () => {
    const result = spawnSync("git", ["ls-files", "scripts/*"], { cwd: root, encoding: "utf8" });
    assert.equal(result.status, 0);
    const paths = result.stdout.trim().split(/\r?\n/).filter((path) => path && !/^scripts\/[^/]+\//.test(path));
    const selectors = cloneMatrix().categories.flatMap((category) => category.path_selectors);
    assert.deepEqual(paths.filter((path) => !selectors.some((selector) => matchesSelector(path, selector))), []);
  }],
  ["rejects a nonzero exit", () => {
    const category = cloneMatrix().categories[0];
    const result = validResult(category);
    result.exit_code = 1;
    assert.throws(() => validateGateResult(category, result), /exit code must be zero/);
  }],
  ["rejects a skipped result", () => {
    const category = cloneMatrix().categories[0];
    const result = validResult(category);
    result.outcome = "skipped";
    assert.throws(() => validateGateResult(category, result), /must pass/);
  }],
  ["rejects zero discovered tests", () => {
    const category = cloneMatrix().categories[0];
    const result = validResult(category);
    result.discovered_tests = 0;
    assert.throws(() => validateGateResult(category, result), /zero tests discovered/);
  }],
  ["rejects zero executed tests", () => {
    const category = cloneMatrix().categories[0];
    const result = validResult(category);
    result.executed_tests = 0;
    assert.throws(() => validateGateResult(category, result), /zero tests executed/);
  }],
  ["rejects skipped tests", () => {
    const category = cloneMatrix().categories[0];
    const result = validResult(category);
    result.skipped_tests = 1;
    assert.throws(() => validateGateResult(category, result), /skipped tests/);
  }],
  ["rejects a cancelled result", () => {
    const category = cloneMatrix().categories[0];
    const result = validResult(category);
    result.outcome = "cancelled";
    assert.throws(() => validateGateResult(category, result), /must pass/);
  }],
  ["rejects a missing pinned tool", () => {
    const category = cloneMatrix().categories[0];
    const result = validResult(category);
    result.tools = [];
    assert.throws(() => validateGateResult(category, result), /missing pinned tool/);
  }],
  ["rejects a mutable command", () => {
    const candidate = cloneMatrix();
    candidate.categories[0].required_commands[0].command = "docker run example:latest";
    assert.throws(() => validateRequiredGateMatrix(candidate, { schema }), /mutable command reference/);
  }],
  ["rejects a catch-all unmatched path allowance", () => {
    const candidate = cloneMatrix();
    candidate.unmatched_path_allowlist[0].selector = "**";
    assert.throws(() => validateRequiredGateMatrix(candidate, { schema }), /reviewed root metadata policy/);
  }],
  ["rejects premature complete", () => {
    const candidate = cloneMatrix();
    candidate.status = "complete";
    candidate.completion_claim = true;
    assert.throws(() => validateRequiredGateMatrix(candidate, { schema }), /remain dependency_blocked/);
  }],
];

for (const [name, run] of tests) {
  run();
  console.log(`ok - ${name}`);
}