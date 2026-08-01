#!/usr/bin/env node

"use strict";

const assert = require("assert").strict;
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
  assert.equal(schema.properties.schema_version.const, "virtengine.prototype.slurm-chart-inventory/v1");
  assert.equal(schema.properties.checkpoint.const, "T4-06A");
  assert.deepEqual(schema.properties.research_scope.properties.roots.const, ["deploy", "infra", "config", "charts", "scripts", "_build/helm"]);
  assert.equal(schema.$defs.source.additionalProperties, false);
  assert.equal(schema.$defs.invariant.additionalProperties, false);
}

function validateSlurmChartInventory(inventory, options = {}) {
  const rootDir = options.rootDir || root;
  const schema = options.schema || loadJson(schemaPath);
  const discoveredSources = options.discoveredSources || discoverSlurmCharts(rootDir, inventory.research_scope.roots);
  const semanticContract = options.semanticContract || Object.fromEntries(inventory.semantic_invariants.map((item) => [item.id, item.status === "satisfied"]));

  validateSchemaContract(schema);
  assert.equal(inventory.schema_version, "virtengine.prototype.slurm-chart-inventory/v1");
  assert.equal(inventory.checkpoint, "T4-06A");
  assert.ok(["blocked", "complete"].includes(inventory.status), "status must be blocked or complete");
  assert.ok(["contract-only", "semantic-render"].includes(inventory.validation_mode), "unknown validation mode");
  assert.deepEqual(inventory.research_scope.roots, ["deploy", "infra", "config", "charts", "scripts", "_build/helm"]);
  assert.equal(inventory.canonical_source.path, "deploy/slurm/slurm-cluster");
  assert.equal(inventory.canonical_source.allowed_use, "authoring-and-runtime-after-hardening");

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
    assert.equal(typeof semanticContract[invariant.id], "boolean", `missing semantic contract for ${invariant.id}`);
    assert.equal(invariant.status === "satisfied", semanticContract[invariant.id], `semantic declaration disagrees with contract: ${invariant.id}`);
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

module.exports = { discoverSlurmCharts, validateSlurmChartInventory };

if (require.main === module) {
  validateSlurmChartInventory(loadJson(inventoryPath));
  console.log("SLURM chart inventory: valid (blocked, contract-only)");
}