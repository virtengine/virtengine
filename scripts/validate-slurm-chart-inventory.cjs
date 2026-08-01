#!/usr/bin/env node

"use strict";

const assert = require("assert").strict;
const { createHash } = require("crypto");
const { existsSync, readdirSync, readFileSync, statSync } = require("fs");
const { join, relative, resolve } = require("path");

const root = resolve(__dirname, "..");
const inventoryPath = resolve(root, "_docs/ralph/prototype-integration/slurm-chart-inventory.json");
const schemaPath = resolve(root, "_docs/ralph/prototype-integration/slurm-chart-inventory.schema.json");
const invariantIds = ["stable-secrets", "replica-capacity-equality", "immutable-images", "least-privilege", "durable-state"];
const dependencyIds = ["84B", "84C", "85C", "87A", "87D"];

function loadJson(path) {
  return JSON.parse(readFileSync(path, "utf8"));
}

function sha256(content) {
  return createHash("sha256").update(content).digest("hex");
}

function validateSemanticReport(report) {
  assert.deepEqual(Object.keys(report).sort(), ["findings", "invariants", "mode", "passed", "schema_version"]);
  assert.equal(report.schema_version, "virtengine.slurm-semantic-validation/v1");
  assert.equal(report.mode, "diagnostic");
  assert.equal(report.passed, false);
  assert.deepEqual(Object.keys(report.invariants).sort(), [...invariantIds].sort(), "semantic report must contain exactly five invariant statuses");
  for (const status of Object.values(report.invariants)) {
    assert.ok(["failed", "unverified", "passed"].includes(status), `invalid semantic report status: ${status}`);
  }
  assert.ok(Array.isArray(report.findings));
  for (const finding of report.findings) {
    assert.deepEqual(Object.keys(finding).sort(), ["invariant", "location", "message"]);
    assert.ok(invariantIds.includes(finding.invariant), `unknown finding invariant: ${finding.invariant}`);
    assert.ok(finding.location.length > 0 && finding.message.length > 0, "semantic findings require location and message");
  }
  const failedStatuses = Object.entries(report.invariants).filter(([, status]) => status === "failed").map(([id]) => id).sort();
  const findingInvariants = [...new Set(report.findings.map((finding) => finding.invariant))].sort();
  assert.deepEqual(findingInvariants, failedStatuses, "semantic findings must exactly identify failed invariant statuses");
}

function toPosix(path) {
  return path.replaceAll("\\", "/");
}

function discoverSlurmCharts(rootDir, roots) {
  const sources = [];
  for (const searchRoot of roots) {
    const absoluteRoot = resolve(rootDir, searchRoot);
    if (!existsSync(absoluteRoot)) continue;
    const pending = [absoluteRoot];
    while (pending.length > 0) {
      const directory = pending.pop();
      for (const entry of readdirSync(directory)) {
        const path = join(directory, entry);
        if (statSync(path).isDirectory()) {
          pending.push(path);
        } else if (entry === "Chart.yaml") {
          const content = readFileSync(path, "utf8");
          if (/\b(slurm|hpc)\b/i.test(content) || /slurm|hpc/i.test(path)) {
            sources.push(toPosix(relative(rootDir, directory)));
          }
        }
      }
    }
  }
  return [...new Set(sources)].sort();
}

function validateSchemaContract(schema) {
  assert.equal(schema.$schema, "https://json-schema.org/draft/2020-12/schema");
  assert.equal(schema.additionalProperties, false);
  assert.equal(schema.properties.schema_version.const, "virtengine.prototype.slurm-chart-inventory/v2");
  assert.equal(schema.properties.checkpoint.const, "T4-06B");
  assert.deepEqual(schema.properties.research_scope.properties.roots.const, ["deploy", "infra", "config", "charts", "scripts", "_build/helm"]);
  assert.equal(schema.$defs.source.additionalProperties, false);
  assert.equal(schema.$defs.invariant.additionalProperties, false);
}

function validateSlurmChartInventory(inventory, options = {}) {
  const rootDir = options.rootDir || root;
  const schema = options.schema || loadJson(schemaPath);
  const discoveredSources = options.discoveredSources || discoverSlurmCharts(rootDir, inventory.research_scope.roots);
  const reportPath = resolve(rootDir, inventory.semantic_report.path);
  const reportContent = options.semanticReportContent || readFileSync(reportPath);
  const semanticReport = options.semanticReport || JSON.parse(reportContent.toString("utf8"));

  validateSchemaContract(schema);
  assert.equal(inventory.schema_version, "virtengine.prototype.slurm-chart-inventory/v2");
  assert.equal(inventory.checkpoint, "T4-06B");
  assert.ok(["blocked", "complete"].includes(inventory.status), "status must be blocked or complete");
  assert.ok(["contract-only", "executable-semantic", "semantic-render"].includes(inventory.validation_mode), "unknown validation mode");
  assert.deepEqual(inventory.research_scope.roots, ["deploy", "infra", "config", "charts", "scripts", "_build/helm"]);
  assert.equal(inventory.canonical_source.path, "deploy/slurm/slurm-cluster");
  assert.equal(inventory.canonical_source.allowed_use, "authoring-and-runtime-after-hardening");

  assert.deepEqual(inventory.semantic_validator, {
    contract: "virtengine.slurm-semantic-validation/v1",
    implementation: "scripts/validate_slurm_chart_semantics.py",
    test: "scripts/validate_slurm_chart_semantics_test.py",
    testdata: "scripts/testdata/slurm-chart-semantics",
    passing_command: "python scripts/validate_slurm_chart_semantics_test.py -v",
    diagnostic_command: "python scripts/validate_slurm_chart_semantics.py --chart deploy/slurm/slurm-cluster --diagnostic --json",
    production_result: "blocked",
  });
  for (const path of [inventory.semantic_validator.implementation, inventory.semantic_validator.test, inventory.semantic_validator.testdata]) {
    assert.ok(existsSync(resolve(rootDir, path)), `semantic validator artifact missing: ${path}`);
  }
  assert.deepEqual(inventory.semantic_report, {
    path: "_docs/ralph/prototype-integration/slurm-chart-semantic-report.json",
    sha256: inventory.semantic_report.sha256,
    source: "deploy/slurm/slurm-cluster",
    generated_by: inventory.semantic_validator.diagnostic_command,
  });
  assert.match(inventory.semantic_report.sha256, /^[a-f0-9]{64}$/);
  assert.equal(sha256(reportContent), inventory.semantic_report.sha256, "semantic report SHA-256 mismatch");
  validateSemanticReport(semanticReport);

  const declaredSources = [inventory.canonical_source, ...inventory.compatibility_import_only_shims, ...inventory.retired_sources];
  const declaredPaths = declaredSources.map((source) => source.path).sort();
  const expectedPresentPaths = declaredSources.filter((source) => source.state !== "absent-forbidden").map((source) => source.path).sort();
  assert.equal(new Set(declaredPaths).size, declaredPaths.length, "chart source paths must be unique");
  const unknownSources = discoveredSources.filter((source) => !declaredPaths.includes(source));
  assert.deepEqual(unknownSources, [], `unknown competing SLURM chart source: discovered ${unknownSources.join(", ")}`);

  for (const shim of inventory.compatibility_import_only_shims) {
    assert.equal(shim.allowed_use, "import-only", `compatibility shim ${shim.path} must be import-only`);
  }
  for (const retired of inventory.retired_sources) {
    assert.equal(retired.allowed_use, "none", `retired source ${retired.path} must forbid all use`);
    const present = discoveredSources.includes(retired.path);
    if (inventory.status === "complete") assert.equal(present, false, `retired source reintroduced: ${retired.path}`);
    assert.equal(retired.state, present ? "present-blocker" : "absent-forbidden", `retired source state mismatch for ${retired.path}`);
  }
  assert.deepEqual(discoveredSources, expectedPresentPaths, `declared SLURM chart source missing: expected ${expectedPresentPaths.join(", ")}`);

  assert.deepEqual(inventory.semantic_invariants.map((item) => item.id).sort(), [...invariantIds].sort());
  for (const invariant of inventory.semantic_invariants) {
    assert.ok(["violated", "unverified", "satisfied"].includes(invariant.status), `invalid status for ${invariant.id}`);
    const expectedReportStatus = { violated: "failed", unverified: "unverified", satisfied: "passed" }[invariant.status];
    assert.equal(semanticReport.invariants[invariant.id], expectedReportStatus, `semantic declaration disagrees with report: ${invariant.id}`);
    if (invariant.status === "satisfied") assert.equal(invariant.blocker, null, `satisfied invariant retains blocker: ${invariant.id}`);
    else assert.ok(invariant.blocker, `unsatisfied invariant lacks blocker: ${invariant.id}`);
  }

  assert.deepEqual(inventory.dependencies.map((item) => item.id).sort(), [...dependencyIds].sort());
  const expectedBlockers = [
    ...inventory.retired_sources.filter((source) => source.state === "present-blocker").map((source) => `retire-source:${source.path}`),
    ...inventory.semantic_invariants.filter((item) => item.status !== "satisfied").map((item) => `invariant:${item.id}`),
    ...inventory.dependencies.filter((item) => item.status !== "ready").map((item) => `dependency:${item.id}`),
  ];
  for (const blocker of expectedBlockers) assert.ok(inventory.blockers.includes(blocker), `missing blocker ${blocker}`);

  const mayComplete = expectedBlockers.length === 0 && inventory.validation_mode === "semantic-render";
  assert.equal(inventory.completion.allowed, mayComplete, "premature complete declaration");
  assert.equal(inventory.status === "complete", mayComplete, "status disagrees with completion readiness");
}

module.exports = { discoverSlurmCharts, validateSemanticReport, validateSlurmChartInventory };

if (require.main === module) {
  validateSlurmChartInventory(loadJson(inventoryPath));
  console.log("SLURM chart inventory: valid (blocked, executable-semantic)");
}