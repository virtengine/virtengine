#!/usr/bin/env node

"use strict";

const assert = require("assert").strict;
const { existsSync, readFileSync } = require("fs");
const { resolve } = require("path");

const root = resolve(__dirname, "..");
const inventoryPath = resolve(root, "_docs/ralph/prototype-integration/migration-inventory.json");
const schemaPath = resolve(root, "_docs/ralph/prototype-integration/migration-inventory.schema.json");
const testCasesPath = resolve(root, "tests/upgrade/test-cases.json");
const shaPattern = /^[a-f0-9]{40}$/;
const stateArrays = ["store_impacts", "prefix_impacts", "protobuf_changes", "params_changes", "index_changes", "genesis_changes"];
const expectedProducers = new Map([
  ["T1", ["ve/prototype-t1-identity", "_docs/ralph/handoffs/prototype-t1/HANDOFF.yaml"]],
  ["T2", ["ve/prototype-t2-product", "_docs/ralph/handoffs/prototype-t2/HANDOFF.yaml"]],
  ["T3", ["ve/prototype-t3-reliability", "_docs/ralph/handoffs/prototype-t3/HANDOFF.yaml"]],
  ["T5", ["ve/prototype-t5-platform", "_docs/ralph/handoffs/prototype-t5/HANDOFF.yaml"]],
]);

function loadJson(path) {
  return JSON.parse(readFileSync(path, "utf8"));
}

function transitionKey(upgrade, moduleName, fromVersion, toVersion) {
  return `${upgrade}|${moduleName}|${fromVersion}|${toVersion}`;
}

function assertExactKeys(value, expected, label) {
  assert.deepEqual(Object.keys(value).sort(), [...expected].sort(), `${label} has unexpected or missing properties`);
}

function validateSchemaControl(schema) {
  assert.equal(schema.$schema, "https://json-schema.org/draft/2020-12/schema");
  assert.equal(schema.additionalProperties, false, "inventory schema root must reject additional properties");
  assert.equal(schema.$defs.migration.additionalProperties, false, "migration schema must reject additional properties");
  assert.equal(schema.$defs.producer.additionalProperties, false, "producer schema must reject additional properties");
  for (const field of stateArrays) {
    assert.ok(schema.$defs.migration.required.includes(field), `migration schema must require ${field}`);
    assert.equal(schema.$defs.migration.properties[field].$ref, "#/$defs/stringArray");
  }
}

function validateMigration(migration, rootDir, ids, label) {
  const keys = ["id", "task", "upgrade", "module", "from_version", "to_version", "execution_kind", ...stateArrays, "evidence"];
  assertExactKeys(migration, keys, label);
  assert.equal(typeof migration.id, "string");
  assert.ok(!ids.has(migration.id), `duplicate migration ID: ${migration.id}`);
  ids.add(migration.id);
  assert.equal(typeof migration.task, "string");
  assert.match(migration.upgrade, /^v1\.[5-8]\.0$/);
  assert.equal(typeof migration.module, "string");
  assert.ok(Number.isInteger(migration.from_version), `${migration.id} from_version must be an integer`);
  assert.equal(migration.to_version, migration.from_version + 1, `${migration.id} must increment by one version`);
  assert.equal(migration.execution_kind, "registered_module_migration");
  for (const field of stateArrays) {
    assert.ok(Array.isArray(migration[field]), `${migration.id} must explicitly declare ${field}`);
  }
  assert.ok(Array.isArray(migration.evidence) && migration.evidence.length > 0, `${migration.id} must declare evidence paths`);
  for (const evidence of migration.evidence) {
    assert.ok(existsSync(resolve(rootDir, evidence)), `${migration.id} evidence path does not exist: ${evidence}`);
  }
}

function validateMigrationInventory(inventory, testCases, options = {}) {
  const rootDir = options.rootDir || root;
  if (options.schema) validateSchemaControl(options.schema);
  assertExactKeys(inventory, ["schema_version", "task", "status", "baseline_commit", "migrations", "producers", "blockers"], "inventory");
  assert.equal(inventory.schema_version, "virtengine.task-88a.migration-inventory/v1");
  assert.equal(inventory.task, "88A");
  assert.ok(["dependency_blocked", "complete"].includes(inventory.status), "invalid inventory status");
  assert.equal(inventory.baseline_commit, "79391a3df86d85522b92e0400c6904971ecbe65d");
  assert.ok(Array.isArray(inventory.migrations));
  assert.ok(Array.isArray(inventory.producers));
  assert.ok(Array.isArray(inventory.blockers));

  const expectedTransitions = [];
  for (const upgrade of ["v1.5.0", "v1.6.0", "v1.7.0", "v1.8.0"]) {
    assert.ok(testCases[upgrade], `missing baseline upgrade fixture: ${upgrade}`);
    for (const [moduleName, transitions] of Object.entries(testCases[upgrade].migrations)) {
      for (const transition of transitions) {
        expectedTransitions.push(transitionKey(upgrade, moduleName, Number(transition.from), Number(transition.to)));
      }
    }
  }

  const ids = new Set();
  for (const migration of inventory.migrations) validateMigration(migration, rootDir, ids, `migration ${migration.id || "<missing>"}`);
  const actualTransitions = inventory.migrations.map((migration) => transitionKey(migration.upgrade, migration.module, migration.from_version, migration.to_version));
  assert.deepEqual(actualTransitions.sort(), expectedTransitions.sort(), "inventory transition set must exactly match v1.5.0 through v1.8.0 fixtures");

  const remainingProducers = new Map(expectedProducers);
  for (const producer of inventory.producers) {
    assertExactKeys(producer, ["thread", "branch", "handoff_path", "status", "sha", "migrations", "blockers"], `producer ${producer.thread || "<missing>"}`);
    assert.ok(remainingProducers.has(producer.thread), `unexpected or duplicate producer: ${producer.thread}`);
    const [branch, handoffPath] = remainingProducers.get(producer.thread);
    assert.equal(producer.branch, branch, `unexpected branch for ${producer.thread}`);
    assert.equal(producer.handoff_path, handoffPath, `unexpected handoff path for ${producer.thread}`);
    assert.ok(["awaiting_committed_handoff", "accepted"].includes(producer.status), `invalid status for ${producer.thread}`);
    assert.ok(Array.isArray(producer.migrations), `${producer.thread} migrations must be explicit`);
    assert.ok(Array.isArray(producer.blockers), `${producer.thread} blockers must be explicit`);
    for (const migration of producer.migrations) {
      validateMigration(migration, rootDir, ids, `${producer.thread} migration ${migration.id || "<missing>"}`);
    }
    if (producer.status === "awaiting_committed_handoff") {
      assert.equal(producer.sha, null, `${producer.thread} awaiting handoff must have null SHA`);
      assert.deepEqual(producer.migrations, [], `${producer.thread} awaiting handoff must not claim migrations`);
      assert.ok(producer.blockers.length > 0, `${producer.thread} awaiting handoff must declare blockers`);
    } else {
      assert.ok(shaPattern.test(producer.sha), `${producer.thread} accepted producer must have a full lowercase SHA`);
    }
    remainingProducers.delete(producer.thread);
  }
  assert.equal(remainingProducers.size, 0, "missing producer declarations");
  if (inventory.producers.some((producer) => producer.status === "awaiting_committed_handoff")) {
    assert.notEqual(inventory.status, "complete", "inventory cannot be complete while a producer awaits committed handoff");
  }
  if (inventory.producers.some((producer) => producer.blockers.length > 0)) {
    assert.notEqual(inventory.status, "complete", "inventory cannot be complete while a producer retains migration blockers");
  }
  if (inventory.status === "dependency_blocked") assert.ok(inventory.blockers.length > 0, "dependency-blocked inventory must declare blockers");
}

module.exports = { validateMigrationInventory };

if (require.main === module) {
  validateMigrationInventory(loadJson(inventoryPath), loadJson(testCasesPath), { rootDir: root, schema: loadJson(schemaPath) });
  console.log("Task 88A migration inventory: valid");
}