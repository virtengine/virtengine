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
    kind: command.kind,
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
  ["requires a runtime integration candidate preflight", () => {
    const candidate = cloneMatrix();
    const category = candidate.categories.find((entry) => entry.id === "docs_process_boundary_e2e");
    category.required_commands = category.required_commands.filter((command) => command.id !== "integration-candidate-preflight");
    assert.throws(() => validateRequiredGateMatrix(candidate, { schema }), /required literal commands changed/);
    const runtime = matrix.categories.find((entry) => entry.id === "docs_process_boundary_e2e").required_commands.find((command) => command.id === "integration-candidate-preflight");
    assert.equal(runtime.kind, "policy");
    assert.equal(runtime.command, "node scripts/preflight-integration-candidate.cjs --repo . --candidate origin/stable-virtengine-beta --canonical 016c4145624deff9673c1c5cff3029277b1f41e0");
  }],
  ["requires an explicit blocked live-candidate projection", () => {
    const currentCategory = matrix.categories.find((entry) => entry.id === "docs_process_boundary_e2e");
    assert.ok(currentCategory.blockers.includes("integration-candidate-preflight-blocked"));
    assert.ok(matrix.blockers.some((blocker) => blocker.id === "integration-candidate-preflight-blocked"));
    const missingCategoryBlocker = cloneMatrix();
    const category = missingCategoryBlocker.categories.find((entry) => entry.id === "docs_process_boundary_e2e");
    category.blockers = category.blockers.filter((blocker) => blocker !== "integration-candidate-preflight-blocked");
    assert.throws(() => validateRequiredGateMatrix(missingCategoryBlocker, { schema }), /root blockers must exactly match category blocker usage/);
    const missingRootBlocker = cloneMatrix();
    missingRootBlocker.blockers = missingRootBlocker.blockers.filter((blocker) => blocker.id !== "integration-candidate-preflight-blocked");
    assert.throws(() => validateRequiredGateMatrix(missingRootBlocker, { schema }), /references unknown blocker/);
  }],
  ["rejects a missing SLURM render command", () => {
    const candidate = cloneMatrix();
    const category = candidate.categories.find((entry) => entry.id === "deployment");
    category.required_commands = category.required_commands.filter((command) => command.id !== "slurm-chart-render");
    assert.throws(() => validateRequiredGateMatrix(candidate, { schema }), /required literal commands changed/);
  }],
  ["rejects an unpinned Helm deployment tool", () => {
    const candidate = cloneMatrix();
    const category = candidate.categories.find((entry) => entry.id === "deployment");
    category.pinned_tools = category.pinned_tools.filter((tool) => tool.name !== "helm");
    assert.throws(() => validateRequiredGateMatrix(candidate, { schema }), /deployment must pin helm 3\.18\.6/);
  }],
  ["rejects duplicate tool, dependency, and blocker identities", () => {
    const duplicateTool = cloneMatrix();
    duplicateTool.categories[0].pinned_tools.push({ ...duplicateTool.categories[0].pinned_tools[0], version: "9.9.9" });
    assert.throws(() => validateRequiredGateMatrix(duplicateTool, { schema }), /duplicate pinned tool/);
    const duplicateDependency = cloneMatrix();
    duplicateDependency.categories[0].dependencies.push(structuredClone(duplicateDependency.categories[0].dependencies[0]));
    assert.throws(() => validateRequiredGateMatrix(duplicateDependency, { schema }), /dependency IDs must be unique/);
    const duplicateBlocker = cloneMatrix();
    duplicateBlocker.categories[0].blockers.push(duplicateBlocker.categories[0].blockers[0]);
    assert.throws(() => validateRequiredGateMatrix(duplicateBlocker, { schema }), /blocker IDs must be unique/);
  }],
  ["rejects a schema that permits duplicate pinned tools", () => {
    const weakenedSchema = JSON.parse(JSON.stringify(schema));
    weakenedSchema.$defs.category.properties.pinned_tools.uniqueItems = false;
    assert.throws(() => validateRequiredGateMatrix(cloneMatrix(), { schema: weakenedSchema }), /category pinned_tools must reject duplicates/);
  }],
  ["rejects a plan schema that permits duplicate categories", () => {
    const planSchema = JSON.parse(readFileSync(resolve(root, "_docs/ralph/prototype-integration/required-gate-plan.schema.json"), "utf8"));
    const resultSchema = JSON.parse(readFileSync(resolve(root, "_docs/ralph/prototype-integration/required-gate-results.schema.json"), "utf8"));
    planSchema.properties.categories.uniqueItems = false;
    assert.throws(() => validateRequiredGateMatrix(cloneMatrix(), { schema, planSchema, resultSchema }), /plan categories must reject duplicates/);
  }],
  ["rejects a result schema that permits duplicate records", () => {
    const planSchema = JSON.parse(readFileSync(resolve(root, "_docs/ralph/prototype-integration/required-gate-plan.schema.json"), "utf8"));
    const resultSchema = JSON.parse(readFileSync(resolve(root, "_docs/ralph/prototype-integration/required-gate-results.schema.json"), "utf8"));
    resultSchema.properties.results.uniqueItems = false;
    assert.throws(() => validateRequiredGateMatrix(cloneMatrix(), { schema, planSchema, resultSchema }), /result schema must reject duplicate result records/);
  }],
  ["rejects malformed dependency declarations and missing blockers", () => {
    const invalidStatus = cloneMatrix();
    invalidStatus.categories[0].dependencies[0].status = "unavailble";
    assert.throws(() => validateRequiredGateMatrix(invalidStatus, { schema }), /invalid status/);
    const extraField = cloneMatrix();
    extraField.categories[0].dependencies[0].reason = "informal";
    assert.throws(() => validateRequiredGateMatrix(extraField, { schema }), /unexpected or missing properties/);
    const missingBlocker = cloneMatrix();
    missingBlocker.categories[0].blockers = [];
    assert.throws(() => validateRequiredGateMatrix(missingBlocker, { schema }), /dependency blockers must exactly match unavailable dependencies/);
  }],
  ["rejects stale dependency and ready-category blockers", () => {
    const staleDependency = cloneMatrix();
    staleDependency.categories[0].dependencies[0].status = "available";
    assert.throws(() => validateRequiredGateMatrix(staleDependency, { schema }), /dependency blockers must exactly match unavailable dependencies/);
    const readyBlocked = cloneMatrix();
    const category = readyBlocked.categories.find((entry) => entry.id === "proto_api");
    category.dependencies.forEach((dependency) => { dependency.status = "available"; });
    category.blockers = ["generated-contract-integration-blocked"];
    category.status = "ready";
    assert.throws(() => validateRequiredGateMatrix(readyBlocked, { schema }), /ready category cannot retain blockers/);
  }],
  ["accepts an available dependency transition with blockers cleared", () => {
    const candidate = cloneMatrix();
    candidate.categories[0].dependencies[0].status = "available";
    candidate.categories[0].blockers = [];
    candidate.categories[0].status = "ready";
    assert.doesNotThrow(() => validateRequiredGateMatrix(candidate, { schema }));
  }],
  ["derives ready and complete matrix states exactly", () => {
    const ready = cloneMatrix();
    ready.categories.forEach((category) => {
      category.dependencies.forEach((dependency) => { dependency.status = "available"; });
      category.blockers = [];
      category.status = "ready";
    });
    ready.blockers = [];
    ready.status = "ready";
    ready.completion_claim = false;
    assert.doesNotThrow(() => validateRequiredGateMatrix(ready, { schema }));
    const complete = JSON.parse(JSON.stringify(ready));
    complete.categories.forEach((category) => { category.status = "complete"; });
    complete.status = "complete";
    complete.completion_claim = true;
    assert.doesNotThrow(() => validateRequiredGateMatrix(complete, { schema }));
  }],
  ["rejects stale root blockers and inconsistent matrix status", () => {
    const unused = cloneMatrix();
    unused.blockers.push({ id: "unused-blocker", description: "Not referenced by any category." });
    assert.throws(() => validateRequiredGateMatrix(unused, { schema }), /root blockers must exactly match category blocker usage/);
    const wrongStatus = cloneMatrix();
    wrongStatus.status = "ready";
    assert.throws(() => validateRequiredGateMatrix(wrongStatus, { schema }), /matrix status must be derived from category states/);
    const falseClaim = cloneMatrix();
    falseClaim.completion_claim = true;
    assert.throws(() => validateRequiredGateMatrix(falseClaim, { schema }), /cannot claim Task 88B completion/);
  }],
  ["rejects a missing SLURM validator deployment selector", () => {
    const candidate = cloneMatrix();
    const category = candidate.categories.find((entry) => entry.id === "deployment");
    category.path_selectors = category.path_selectors.filter((selector) => selector !== "scripts/validate_slurm_chart_semantics.py");
    assert.throws(() => validateRequiredGateMatrix(candidate, { schema }), /deployment selectors must include scripts\/validate_slurm_chart_semantics\.py/);
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
  ["rejects incomplete tests and nonzero policy counts", () => {
    const testCategory = cloneMatrix().categories[0];
    const incomplete = validResult(testCategory);
    incomplete.discovered_tests = 3;
    incomplete.executed_tests = 2;
    assert.throws(() => validateGateResult(testCategory, incomplete), /not all discovered tests executed/);
    const policyCategory = cloneMatrix().categories.find((entry) => entry.id === "docs_process_boundary_e2e");
    const policy = policyCategory.required_commands.find((entry) => entry.kind === "policy");
    const result = { ...validResult(policyCategory), command_id: policy.id, kind: policy.kind, command: policy.command, discovered_tests: 1, executed_tests: 1 };
    assert.throws(() => validateGateResult(policyCategory, result), /policy commands must report zero test counts/);
  }],
  ["rejects a mismatched result command kind", () => {
    const category = cloneMatrix().categories[0];
    const result = validResult(category);
    result.kind = "policy";
    assert.throws(() => validateGateResult(category, result), /kind must match/);
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
    assert.throws(() => validateGateResult(category, result), /pinned tool count mismatch/);
  }],
  ["rejects extra, duplicate, or malformed tool evidence", () => {
    const category = cloneMatrix().categories[0];
    const extra = validResult(category);
    extra.tools.push({ name: "extra", version: "1.0.0", available: true });
    assert.throws(() => validateGateResult(category, extra), /pinned tool count mismatch/);
    const twoToolCategory = cloneMatrix().categories.find((entry) => entry.id === "docs_process_boundary_e2e");
    const duplicate = validResult(twoToolCategory);
    duplicate.tools[1] = structuredClone(duplicate.tools[0]);
    assert.throws(() => validateGateResult(twoToolCategory, duplicate), /duplicate result tool/);
    const malformed = validResult(category);
    malformed.tools[0].source = "undeclared";
    assert.throws(() => validateGateResult(category, malformed), /unexpected or missing properties/);
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
    assert.throws(() => validateRequiredGateMatrix(candidate, { schema }), /matrix status must be derived from category states/);
  }],
];

for (const [name, run] of tests) {
  run();
  console.log(`ok - ${name}`);
}