#!/usr/bin/env node

"use strict";

const assert = require("assert").strict;
const { readFileSync } = require("fs");
const { resolve } = require("path");

const root = resolve(__dirname, "..");
const controlPath = resolve(root, "_docs/ralph/prototype-integration/control.json");
const schemaPath = resolve(root, "_docs/ralph/prototype-integration/producer-handoff.schema.json");
const handoffPath = resolve(root, "_docs/ralph/handoffs/prototype-integration/HANDOFF.yaml");

function loadJson(path) {
  return JSON.parse(readFileSync(path, "utf8"));
}

function validateIntegrationControl(control, schema, handoff) {
  const shaPattern = /^[a-f0-9]{40}$/;
  const expectedProducers = new Map([
    ["T1", "ve/prototype-t1-identity"],
    ["T2", "ve/prototype-t2-product"],
    ["T3", "ve/prototype-t3-reliability"],
    ["T5", "ve/prototype-t5-platform"],
  ]);

  assert.equal(control.schema_version, "virtengine.prototype.integration-control/v1");
  assert.equal(control.campaign, "three-day-prototype");
  assert.equal(control.baseline.frozen, true);
  assert.ok(shaPattern.test(control.baseline.sha), "baseline SHA must be a full lowercase commit SHA");
  assert.equal(control.integration.thread, "T4");
  assert.equal(control.integration.branch, "ve/prototype-integration");
  assert.equal(control.producers.length, expectedProducers.size);

  for (const producer of control.producers) {
    assert.equal(producer.branch, expectedProducers.get(producer.thread), `unexpected producer registration for ${producer.thread}`);
    expectedProducers.delete(producer.thread);
  }
  assert.equal(expectedProducers.size, 0, "missing producer registration");

  assert.ok(control.path_ownership.integration_only.length > 0, "integration-only paths must not be empty");
  assert.equal(new Set(control.path_ownership.integration_only).size, control.path_ownership.integration_only.length, "integration-only paths must be unique");
  assert.deepEqual(control.dependency_ledger.ready, ["T4-01"]);
  assert.equal(control.generated_file_lease.state, "available");
  assert.equal(control.generated_file_lease.holder, null);
  assert.deepEqual(control.generated_file_lease.paths, []);

  assert.equal(schema.$schema, "https://json-schema.org/draft/2020-12/schema");
  for (const field of ["end_head", "tree_clean", "tests", "migrations", "blockers", "expected_head"]) {
    assert.ok(schema.required.includes(field), `producer schema must require ${field}`);
  }
  assert.equal(schema.properties.tree_clean.const, true);
  assert.equal(schema.$defs.testResult.properties.exit_code.const, 0);
  assert.equal(schema.$defs.testResult.properties.result.const, "passed");

  assert.equal(handoff.campaign, control.campaign);
  assert.equal(handoff.thread, "T4");
  assert.equal(handoff.branch, control.integration.branch);
  assert.equal(handoff.start_head, control.baseline.sha);
  assert.equal(handoff.origin_main, control.baseline.sha);
  assert.ok(Array.isArray(handoff.accepted_checkpoints));
  assert.ok(Array.isArray(handoff.rejected_checkpoints));
}

module.exports = { validateIntegrationControl };

if (require.main === module) {
  validateIntegrationControl(loadJson(controlPath), loadJson(schemaPath), loadJson(handoffPath));
  console.log("prototype integration controls: valid");
}