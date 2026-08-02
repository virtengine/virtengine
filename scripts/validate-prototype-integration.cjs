#!/usr/bin/env node

"use strict";

const assert = require("assert").strict;
const { readFileSync } = require("fs");
const { resolve } = require("path");
const { validateSecurityGates } = require("./validate-ai-biometric-security-gates.cjs");
const { validateProductionPolicy } = require("./validate-ai-production-policy.cjs");
const { validateManifest: validateCoreRcManifest } = require("./generate-core-rc-manifest.cjs");
const { validateSchema: validateCoreRcSchema } = require("./validate-core-rc-manifest.cjs");
const { validateFundRouteInventory } = require("./validate-fund-route-inventory.cjs");
const { validateMigrationInventory } = require("./validate-migration-inventory.cjs");
const { validateModelProvenance } = require("./validate-model-provenance.cjs");
const { validateReportSchema: validatePublicationPreflightSchema } = require("./preflight-core-rc-publication.cjs");
const { validateRequiredGateMatrix } = require("./validate-required-gate-matrix.cjs");
const { validateSlurmChartInventory } = require("./validate-slurm-chart-inventory.cjs");

const root = resolve(__dirname, "..");
const controlPath = resolve(root, "_docs/ralph/prototype-integration/control.json");
const aiBiometricSecurityGatesPath = resolve(root, "_docs/ralph/prototype-integration/ai-biometric-security-gates.json");
const aiProductionPolicyPath = resolve(root, "_docs/ralph/prototype-integration/ai-production-policy.json");
const coreRcManifestPath = resolve(root, "_docs/ralph/prototype-integration/core-rc-manifest.json");
const coreRcSchemaPath = resolve(root, "_docs/ralph/prototype-integration/core-rc-manifest.schema.json");
const schemaPath = resolve(root, "_docs/ralph/prototype-integration/producer-handoff.schema.json");
const epochPath = resolve(root, "_docs/ralph/prototype-integration/epochs/epoch-1.json");
const fundRouteInventoryPath = resolve(root, "_docs/ralph/prototype-integration/fund-route-inventory.json");
const handoffPath = resolve(root, "_docs/ralph/handoffs/prototype-integration/HANDOFF.yaml");
const migrationInventoryPath = resolve(root, "_docs/ralph/prototype-integration/migration-inventory.json");
const migrationSchemaPath = resolve(root, "_docs/ralph/prototype-integration/migration-inventory.schema.json");
const modelProvenancePath = resolve(root, "_docs/ralph/prototype-integration/model-provenance.json");
const modelProvenanceSchemaPath = resolve(root, "_docs/ralph/prototype-integration/model-provenance.schema.json");
const publicationPreflightSchemaPath = resolve(root, "_docs/ralph/prototype-integration/core-rc-publication-preflight.schema.json");
const requiredGateMatrixPath = resolve(root, "_docs/ralph/prototype-integration/required-gate-matrix.json");
const requiredGateSchemaPath = resolve(root, "_docs/ralph/prototype-integration/required-gate-matrix.schema.json");
const requiredGatePlanSchemaPath = resolve(root, "_docs/ralph/prototype-integration/required-gate-plan.schema.json");
const requiredGateResultSchemaPath = resolve(root, "_docs/ralph/prototype-integration/required-gate-results.schema.json");
const slurmChartInventoryPath = resolve(root, "_docs/ralph/prototype-integration/slurm-chart-inventory.json");
const slurmChartSchemaPath = resolve(root, "_docs/ralph/prototype-integration/slurm-chart-inventory.schema.json");
const testCasesPath = resolve(root, "tests/upgrade/test-cases.json");

function loadJson(path) {
  return JSON.parse(readFileSync(path, "utf8"));
}

function validateEpoch(epoch, handoff = null) {
  const expectedKeys = ["announcement_cutoff", "base_sha", "base_tag", "campaign", "intake_epoch", "opens_at", "planning_sha", "producers", "schema_version", "status"];
  assert.deepEqual(Object.keys(epoch).sort(), expectedKeys);
  assert.equal(epoch.schema_version, "virtengine.prototype.intake-epoch/v2");
  assert.equal(epoch.campaign, "three-day-prototype");
  assert.equal(epoch.intake_epoch, 1);
  assert.equal(epoch.base_tag, "checkpoint/prototype-integration/epoch-1-base");
  assert.equal(epoch.base_sha, "5587c384f634552c3a2dd7181ca49cafa4da1984");
  assert.equal(epoch.planning_sha, "1436723bd78980aa0388dbe9fcfa24dda939c54a");
  assert.ok(["open", "frozen", "closed"].includes(epoch.status), "epoch status must be open, frozen, or closed");
  assert.ok(Number.isFinite(Date.parse(epoch.opens_at)), "epoch opens_at must be UTC date-time");
  assert.ok(Number.isFinite(Date.parse(epoch.announcement_cutoff)), "epoch announcement_cutoff must be UTC date-time");
  assert.match(epoch.opens_at, /Z$/);
  assert.match(epoch.announcement_cutoff, /Z$/);
  assert.ok(Date.parse(epoch.opens_at) < Date.parse(epoch.announcement_cutoff), "epoch cutoff must follow opens_at");
  assert.deepEqual(epoch.producers.map((producer) => producer.thread), ["T1", "T2", "T3", "T5"]);
  for (const producer of epoch.producers) {
    assert.deepEqual(Object.keys(producer).sort(), ["decision", "status", "tag", "thread"]);
    assert.ok(["unannounced", "announced", "accepted", "rejected"].includes(producer.status));
    if (producer.status === "unannounced") {
      assert.equal(producer.tag, null);
      assert.ok(producer.decision === null || producer.decision === "frozen-out", "unannounced producer decision must be null or frozen-out");
      if (producer.decision === "frozen-out") assert.ok(["frozen", "closed"].includes(epoch.status), "open epoch cannot freeze out a producer");
    } else {
      assert.match(producer.tag, /^checkpoint\/prototype-t[1235]\/t[1235]-[0-9]{2,}[a-z]?$/);
      if (producer.status === "announced") assert.equal(producer.decision, null);
      if (producer.status === "accepted") assert.equal(producer.decision, "accepted", "accepted producer must record decision=accepted");
      if (producer.status === "rejected") assert.equal(producer.decision, "rejected", "rejected producer must record decision=rejected");
    }
  }
  if (handoff) {
    const accepted = Array.isArray(handoff.accepted_checkpoints) ? handoff.accepted_checkpoints : [];
    for (const producer of epoch.producers.filter((entry) => entry.status === "accepted")) {
      const matches = accepted.filter((entry) => entry.thread === producer.thread && entry.tag === producer.tag);
      assert.equal(matches.length, 1, `accepted epoch producer ${producer.thread} must match exactly one accepted ledger checkpoint/tag`);
      assert.match(matches[0].checkpoint, new RegExp(`^${producer.thread}-[0-9]{2,}[A-Z]?$`));
      assert.match(matches[0].tip, /^[a-f0-9]{40}$/);
      assert.match(matches[0].payload_head, /^[a-f0-9]{40}$/);
    }
  }
}

function validateIntegrationControl(control, schema, handoff, epoch) {
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
  assert.deepEqual(control.dependency_ledger.ready, ["T4-01", "T4-06B"]);
  assert.equal(control.generated_file_lease.state, "available");
  assert.equal(control.generated_file_lease.holder, null);
  assert.deepEqual(control.generated_file_lease.paths, []);

  assert.equal(schema.$schema, "https://json-schema.org/draft/2020-12/schema");
  const intakeFields = ["campaign", "thread", "checkpoint", "branch", "frozen_baseline", "planning_sha", "intake_epoch", "intake_base_sha", "payload_head", "prior_accepted_payload", "tree_clean", "commits_since_prior_acceptance", "owned_paths", "files_changed", "tests", "generated_hashes", "migrations", "external_evidence", "known_failures", "blockers", "next_checkpoint"];
  assert.deepEqual([...schema.required].sort(), [...intakeFields].sort(), "producer schema required fields must exactly match intake-v2");
  assert.equal(schema.additionalProperties, false);
  assert.equal(schema.properties.frozen_baseline.const, control.baseline.sha);
  assert.equal(schema.properties.commits_since_prior_acceptance.minItems, 1);
  assert.equal(schema.properties.tree_clean.const, true);
  assert.equal(schema.$defs.testResult.additionalProperties, false);
  assert.deepEqual(schema.$defs.testResult.required, ["command", "exit_code", "result", "tool_versions", "artifact"]);
  assert.equal(schema.$defs.testResult.properties.exit_code.const, 0);
  assert.equal(schema.$defs.testResult.properties.result.const, "passed");
  assert.equal(schema.$defs.testResult.properties.tool_versions.minProperties, 1);

  assert.equal(handoff.campaign, control.campaign);
  assert.equal(handoff.thread, "T4");
  assert.equal(handoff.branch, control.integration.branch);
  assert.ok(shaPattern.test(handoff.start_head), "handoff start_head must be a full lowercase commit SHA");
  assert.ok(shaPattern.test(handoff.end_head), "handoff end_head must be a full lowercase commit SHA");
  assert.ok(shaPattern.test(handoff.expected_head), "handoff expected_head must be a full lowercase commit SHA");
  assert.equal(handoff.origin_main, control.baseline.sha);
  assert.equal(handoff.tree_clean, true);
  assert.ok(Array.isArray(handoff.accepted_checkpoints));
  assert.ok(Array.isArray(handoff.rejected_checkpoints));
  validateEpoch(epoch, handoff);
}

module.exports = { validateEpoch, validateIntegrationControl };

if (require.main === module) {
  const handoff = loadJson(handoffPath);
  validateIntegrationControl(loadJson(controlPath), loadJson(schemaPath), handoff, loadJson(epochPath));
  validateSecurityGates(loadJson(aiBiometricSecurityGatesPath), { rootDir: root });
  validateProductionPolicy(loadJson(aiProductionPolicyPath), { rootDir: root });
  validateCoreRcSchema(loadJson(coreRcSchemaPath));
  validateCoreRcManifest(loadJson(coreRcManifestPath), { rootDir: root });
  validateFundRouteInventory(loadJson(fundRouteInventoryPath), {
    rootDir: root,
    verifyAcceptedCheckpoint: (checkpoint) => handoff.accepted_checkpoints.some((entry) => entry.thread === "T5" && entry.tag === checkpoint.tag && entry.payload_head === checkpoint.payload_sha),
  });
  validateMigrationInventory(loadJson(migrationInventoryPath), loadJson(testCasesPath), {
    rootDir: root,
    schema: loadJson(migrationSchemaPath),
  });
  validateModelProvenance(loadJson(modelProvenancePath), {
    rootDir: root,
    schema: loadJson(modelProvenanceSchemaPath),
  });
  validatePublicationPreflightSchema(loadJson(publicationPreflightSchemaPath));
  validateRequiredGateMatrix(loadJson(requiredGateMatrixPath), {
    schema: loadJson(requiredGateSchemaPath),
    planSchema: loadJson(requiredGatePlanSchemaPath),
    resultSchema: loadJson(requiredGateResultSchemaPath),
  });
  validateSlurmChartInventory(loadJson(slurmChartInventoryPath), {
    rootDir: root,
    schema: loadJson(slurmChartSchemaPath),
  });
  console.log("prototype integration controls: valid");
}