#!/usr/bin/env node

"use strict";

const assert = require("assert").strict;
const { readFileSync } = require("fs");
const { resolve } = require("path");

const root = resolve(__dirname, "..");
const matrixPath = resolve(root, "_docs/ralph/prototype-integration/required-gate-matrix.json");
const schemaPath = resolve(root, "_docs/ralph/prototype-integration/required-gate-matrix.schema.json");
const planSchemaPath = resolve(root, "_docs/ralph/prototype-integration/required-gate-plan.schema.json");
const resultSchemaPath = resolve(root, "_docs/ralph/prototype-integration/required-gate-results.schema.json");
const categoryCommands = new Map([
  ["go", ["go vet ./...", "go test -count=1 ./..."]],
  ["proto_api", ["docker build --file sdk/generation/Dockerfile --tag virtengine-proto-gen:1.0.0 .", "docker run --rm --volume $PWD:/src --workdir /src/sdk --entrypoint buf virtengine-proto-gen:1.0.0 lint", "bash scripts/verify-proto-generation.sh", "go test -count=1 ./api/... ./client/...", "node scripts/validate-generated-contract-inventory.cjs --require-ready"]],
  ["sdk", ["pnpm --dir sdk/ts install --frozen-lockfile", "pnpm --dir sdk/ts build", "pnpm --dir sdk/ts test"]],
  ["portal", ["pnpm --dir portal install --frozen-lockfile", "pnpm --dir portal lint", "pnpm --dir portal test"]],
  ["mobile", ["pnpm --dir mobile/veid-capture-app install --frozen-lockfile", "pnpm --dir mobile/veid-capture-app typecheck", "pnpm --dir mobile/veid-capture-app test"]],
  ["ml", ["python -m pip install --require-hashes -r ml/requirements-deterministic.txt", "python -m pytest --collect-only -q ml/training/tests", "python -m pytest -q ml/training/tests", "node scripts/validate-ai-production-policy.cjs --enforce", "node scripts/validate-ai-biometric-security-gates.cjs --enforce"]],
  ["deployment", ["docker compose -f docker-compose.yaml config --quiet", "bash scripts/ci/post-deploy-smoke-test.sh"]],
  ["observability", ["docker compose -f docker-compose.observability.yaml config --quiet", "go test -count=1 ./pkg/observability/..."]],
  ["upgrades", ["go test -count=1 ./tests/upgrade/..."]],
  ["security", ["python .github/scripts/validate_security_policies.py", "go test -count=1 -tags=e2e.integration ./tests/integration/security/...", "node scripts/validate-fund-route-inventory.cjs --require-ready"]],
  ["docs_process_boundary_e2e", ["node scripts/validate-prototype-integration.cjs", "node scripts/observe-prototype-intake-tags.test.cjs", "node scripts/plan-prototype-intake-freeze.test.cjs", "node scripts/apply-prototype-intake-freeze.test.cjs", "node scripts/inspect-prototype-intake-freeze-lease.test.cjs", "node scripts/recover-prototype-intake-freeze-lease.test.cjs", "node scripts/preflight-integration-candidate.test.cjs", "node scripts/validate-localnet-integration-launchers.test.cjs", "go test -count=1 -tags=e2e.integration ./tests/e2e/... ./tests/integration/..."]],
]);
const rootKeys = ["schema_version", "task", "control_scope", "status", "completion_claim", "unmatched_path_allowlist", "categories", "blockers"];
const expectedUnmatchedAllowlist = [
  ["*.md", "Root-level policy, legal, and repository narrative metadata has no executable gate."],
  ["*.txt", "Root-level repository text metadata has no executable gate."],
  [".editorconfig", "Repository editor metadata has no executable gate."],
  [".gitattributes", "Repository Git attribute metadata has no executable gate."],
  [".gitignore", "Repository Git ignore metadata has no executable gate."],
  ["codecov.yml", "Repository coverage-reporting metadata has no local executable gate."],
];
const categoryKeys = ["id", "path_selectors", "required_commands", "pinned_tools", "zero_test_policy", "failure_semantics", "dependencies", "status", "blockers"];

function loadJson(path) {
  return JSON.parse(readFileSync(path, "utf8"));
}

function assertExactKeys(value, expected, label) {
  assert.ok(value && typeof value === "object" && !Array.isArray(value), `${label} must be an object`);
  assert.deepEqual(Object.keys(value).sort(), [...expected].sort(), `${label} has unexpected or missing properties`);
}

function validateSchemaControl(schema) {
  assert.equal(schema.$schema, "https://json-schema.org/draft/2020-12/schema");
  assert.equal(schema.additionalProperties, false, "matrix schema root must reject additional properties");
  assert.deepEqual(schema.required.slice().sort(), rootKeys.slice().sort(), "matrix schema root required fields must be exact");
  assert.equal(schema.properties.categories.minItems, categoryCommands.size);
  assert.equal(schema.properties.categories.maxItems, categoryCommands.size);
  assert.equal(schema.$defs.category.additionalProperties, false);
  assert.deepEqual(schema.$defs.category.required.slice().sort(), categoryKeys.slice().sort(), "category schema required fields must be exact");
  assert.equal(schema.$defs.zeroTestPolicy.properties.minimum_discovered.const, 1);
  assert.equal(schema.$defs.zeroTestPolicy.properties.minimum_executed.const, 1);
  for (const field of ["nonzero_exit", "skipped", "missing_tool", "cancelled"]) {
    assert.equal(schema.$defs.failureSemantics.properties[field].const, "fail");
  }
}

function validateExecutionSchemas(planSchema, resultSchema) {
  assert.equal(planSchema.$schema, "https://json-schema.org/draft/2020-12/schema");
  assert.equal(planSchema.additionalProperties, false, "plan schema root must reject additional properties");
  assert.deepEqual(planSchema.required.slice().sort(), ["schema_version", "base_sha", "head_sha", "matrix_digest", "matrix_status", "changed_paths", "allowlisted_paths", "categories"].sort());
  assert.equal(planSchema.properties.schema_version.const, "virtengine.task-88b.required-gate-plan/v1");
  assert.deepEqual(planSchema.properties.matrix_status.enum, ["dependency_blocked", "ready", "complete"]);
  assert.equal(planSchema.$defs.category.additionalProperties, false);
  assert.equal(planSchema.$defs.command.additionalProperties, false);
  assert.equal(planSchema.$defs.tool.additionalProperties, false);

  assert.equal(resultSchema.$schema, "https://json-schema.org/draft/2020-12/schema");
  assert.equal(resultSchema.additionalProperties, false, "result schema root must reject additional properties");
  assert.deepEqual(resultSchema.required.slice().sort(), ["schema_version", "base_sha", "head_sha", "matrix_digest", "results"].sort());
  assert.equal(resultSchema.properties.schema_version.const, "virtengine.task-88b.required-gate-results/v1");
  assert.equal(resultSchema.$defs.result.additionalProperties, false);
  assert.equal(resultSchema.$defs.result.properties.outcome.const, "passed");
  assert.equal(resultSchema.$defs.result.properties.exit_code.const, 0);
  assert.equal(resultSchema.$defs.result.properties.skipped_tests.const, 0);
  assert.equal(resultSchema.$defs.tool.properties.available.const, true);
}

function assertImmutableCommand(command, label) {
  assert.equal(typeof command, "string", `${label} command must be a string`);
  assert.doesNotMatch(command, /(?:^|[\s/:@#._-])(?:latest|main|master|head)(?:$|[\s/:@#._-])/i, `${label} contains a mutable command reference`);
  assert.doesNotMatch(command, /(?:npm|pnpm|pip)\s+(?:install|add)\s+(?!.*(?:--frozen-lockfile|--require-hashes))/, `${label} dependency installation is mutable`);
}

function validateGateResult(category, result) {
  assertExactKeys(result, ["command_id", "command", "outcome", "exit_code", "discovered_tests", "executed_tests", "skipped_tests", "tools"], "gate result");
  const command = category.required_commands.find((entry) => entry.id === result.command_id);
  assert.ok(command, `gate result references unknown command: ${result.command_id}`);
  assert.equal(result.command, command.command, "gate result command must match the required literal command");
  assert.equal(result.outcome, "passed", `gate result must pass, not ${result.outcome}`);
  assert.equal(result.exit_code, 0, "gate result exit code must be zero");
  assert.ok(Array.isArray(result.tools), "gate result tools must be an array");
  for (const pinned of category.pinned_tools) {
    const actual = result.tools.find((tool) => tool.name === pinned.name);
    assert.ok(actual && actual.available === true, `missing pinned tool: ${pinned.name}`);
    assert.equal(actual.version, pinned.version, `pinned tool version mismatch: ${pinned.name}`);
  }
  if (command.kind === "test") {
    assert.ok(result.discovered_tests >= category.zero_test_policy.minimum_discovered, "zero tests discovered");
    assert.ok(result.executed_tests >= category.zero_test_policy.minimum_executed, "zero tests executed");
    assert.equal(result.skipped_tests, 0, "skipped tests are not allowed");
  }
}

function validateRequiredGateMatrix(matrix, options = {}) {
  if (options.schema) validateSchemaControl(options.schema);
  if (options.planSchema || options.resultSchema) {
    assert.ok(options.planSchema && options.resultSchema, "plan and result schemas must be validated together");
    validateExecutionSchemas(options.planSchema, options.resultSchema);
  }
  assertExactKeys(matrix, rootKeys, "matrix");
  assert.equal(matrix.schema_version, "virtengine.task-88b.required-gate-matrix/v1");
  assert.equal(matrix.task, "88B");
  assert.equal(matrix.control_scope, "required_gate_matrix_only");
  assert.ok(["dependency_blocked", "ready", "complete"].includes(matrix.status), "invalid matrix status");
  assert.equal(typeof matrix.completion_claim, "boolean");
  assert.ok(Array.isArray(matrix.unmatched_path_allowlist) && matrix.unmatched_path_allowlist.length > 0, "unmatched path allowlist must not be empty");
  const allowanceSelectors = new Set();
  for (const allowance of matrix.unmatched_path_allowlist) {
    assertExactKeys(allowance, ["selector", "reason"], "unmatched path allowance");
    assert.ok(typeof allowance.selector === "string" && allowance.selector.length > 0 && !allowance.selector.includes("\\"), "invalid unmatched path selector");
    assert.ok(!allowanceSelectors.has(allowance.selector), `duplicate unmatched path selector: ${allowance.selector}`);
    allowanceSelectors.add(allowance.selector);
    assert.ok(typeof allowance.reason === "string" && allowance.reason.length > 0, `allowance ${allowance.selector} must have a reason`);
  }
  assert.ok(Array.isArray(matrix.categories), "categories must be an array");
  assert.ok(Array.isArray(matrix.blockers), "blockers must be an array");
  assert.deepEqual(
    matrix.unmatched_path_allowlist.map((entry) => [entry.selector, entry.reason]),
    expectedUnmatchedAllowlist,
    "unmatched path allowlist must use the reviewed root metadata policy",
  );

  const blockerIds = new Set();
  for (const blocker of matrix.blockers) {
    assertExactKeys(blocker, ["id", "description"], `blocker ${blocker.id || "<missing>"}`);
    assert.match(blocker.id, /^[a-z0-9.-]+$/);
    assert.ok(!blockerIds.has(blocker.id), `duplicate blocker: ${blocker.id}`);
    blockerIds.add(blocker.id);
    assert.ok(typeof blocker.description === "string" && blocker.description.length > 0, `${blocker.id} description must not be empty`);
  }

  const remaining = new Map(categoryCommands);
  for (const category of matrix.categories) {
    assertExactKeys(category, categoryKeys, `category ${category.id || "<missing>"}`);
    assert.ok(remaining.has(category.id), `unexpected or duplicate category: ${category.id}`);
    assert.ok(Array.isArray(category.path_selectors) && category.path_selectors.length > 0, `${category.id} path selectors must not be empty`);
    for (const selector of category.path_selectors) {
      assert.ok(typeof selector === "string" && selector.length > 0 && !selector.includes("\\"), `${category.id} has an invalid path selector`);
    }
    assert.ok(Array.isArray(category.required_commands) && category.required_commands.length > 0, `${category.id} required commands must not be empty`);
    const commandIds = new Set();
    for (const command of category.required_commands) {
      assertExactKeys(command, ["id", "kind", "command"], `${category.id} command`);
      assert.ok(!commandIds.has(command.id), `${category.id} has duplicate command ID: ${command.id}`);
      commandIds.add(command.id);
      assertImmutableCommand(command.command, `${category.id}.${command.id}`);
    }
    assert.ok(category.required_commands.some((command) => command.kind === "test"), `${category.id} must define a test command`);
    assert.deepEqual(category.required_commands.map((entry) => entry.command), remaining.get(category.id), `${category.id} required literal commands changed`);
    assert.ok(Array.isArray(category.pinned_tools) && category.pinned_tools.length > 0, `${category.id} pinned tools must not be empty`);
    for (const tool of category.pinned_tools) {
      assertExactKeys(tool, ["name", "version", "source"], `${category.id} pinned tool`);
      assert.match(tool.version, /^[0-9]+\.[0-9]+(?:\.[0-9]+)?$/, `${category.id} tool ${tool.name} must use an exact version`);
    }
    assert.deepEqual(category.zero_test_policy, { applies_to: "test_commands", minimum_discovered: 1, minimum_executed: 1, empty_selection: "fail" });
    assert.deepEqual(category.failure_semantics, { nonzero_exit: "fail", skipped: "fail", missing_tool: "fail", cancelled: "fail" });
    assert.ok(Array.isArray(category.dependencies) && category.dependencies.length > 0, `${category.id} dependencies must not be empty`);
    assert.ok(["dependency_blocked", "ready", "complete"].includes(category.status), `${category.id} has invalid status`);
    assert.ok(Array.isArray(category.blockers), `${category.id} blockers must be an array`);
    const unavailable = category.dependencies.some((dependency) => dependency.status === "unavailable");
    if (unavailable) {
      assert.equal(category.status, "dependency_blocked", `${category.id} must remain dependency_blocked`);
      assert.ok(category.blockers.length > 0, `${category.id} must identify dependency blockers`);
    }
    for (const blocker of category.blockers) assert.ok(blockerIds.has(blocker), `${category.id} references unknown blocker: ${blocker}`);
    if (category.status === "complete") assert.equal(unavailable, false, `${category.id} cannot complete with unavailable dependencies`);
    remaining.delete(category.id);
  }
  assert.equal(remaining.size, 0, `missing category: ${[...remaining.keys()].join(", ")}`);

  const blocked = matrix.categories.some((category) => category.status === "dependency_blocked");
  if (blocked) assert.equal(matrix.status, "dependency_blocked", "matrix must remain dependency_blocked while a category is blocked");
  if (matrix.status === "dependency_blocked") {
    assert.equal(matrix.completion_claim, false, "dependency-blocked matrix cannot claim Task 88B completion");
    assert.ok(matrix.blockers.length > 0, "dependency-blocked matrix must declare blockers");
  }
  if (matrix.status === "complete") {
    assert.equal(matrix.completion_claim, true, "complete matrix must explicitly claim completion");
    assert.ok(matrix.categories.every((category) => category.status === "complete"), "premature complete: every category must be complete");
    assert.equal(matrix.blockers.length, 0, "complete matrix cannot retain blockers");
  }
}

module.exports = { validateExecutionSchemas, validateGateResult, validateRequiredGateMatrix };

if (require.main === module) {
  validateRequiredGateMatrix(loadJson(matrixPath), {
    schema: loadJson(schemaPath),
    planSchema: loadJson(planSchemaPath),
    resultSchema: loadJson(resultSchemaPath),
  });
  console.log("Task 88B required-gate matrix control: valid (dependency_blocked)");
}