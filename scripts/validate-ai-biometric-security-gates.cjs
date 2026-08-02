#!/usr/bin/env node

"use strict";

const assert = require("assert").strict;
const { existsSync, readFileSync, statSync } = require("fs");
const { isAbsolute, resolve } = require("path");

const root = resolve(__dirname, "..");
const gatePath = resolve(root, "_docs/ralph/prototype-integration/ai-biometric-security-gates.json");
const categoryIds = ["template-inversion", "biometric-linkability", "replay", "concurrent-enrollment", "arbitrary-otp", "forged-envelopes", "transferred-proofs", "fund-route-coverage", "deletion-receipts", "client-cleanup"];
const externalBlockerIds = ["privacy-evaluation-absent", "pad-evaluation-absent", "production-model-evaluation-absent"];

function exactKeys(value, keys, label) {
  assert.deepEqual(Object.keys(value).sort(), [...keys].sort(), `${label} has unknown or missing fields`);
}

function validateSecurityGates(gates, options = {}) {
  const rootDir = options.rootDir || root;
  exactKeys(gates, ["blockers", "categories", "checkpoint", "external_blockers", "schema_version", "status"], "security gates");
  assert.equal(gates.schema_version, "virtengine.prototype.ai-biometric-security-gates/v1");
  assert.equal(gates.checkpoint, "T4-15A");
  assert.ok(["blocked", "complete"].includes(gates.status));
  assert.deepEqual(gates.categories.map((category) => category.id).sort(), [...categoryIds].sort());
  assert.ok(gates.external_blockers.every((blocker) => externalBlockerIds.includes(blocker.id)), "unknown external blocker");
  assert.equal(new Set(gates.external_blockers.map((blocker) => blocker.id)).size, gates.external_blockers.length, "duplicate external blocker");
  assert.equal(new Set(gates.blockers).size, gates.blockers.length, "blockers must be unique");

  for (const blocker of gates.external_blockers) {
    exactKeys(blocker, ["description", "id"], `external blocker ${blocker.id || "unknown"}`);
    assert.ok(blocker.description.length > 0);
    assert.ok(gates.blockers.includes(blocker.id), `missing external blocker ${blocker.id}`);
  }
  for (const category of gates.categories) {
    exactKeys(category, ["blockers", "command", "evidence_files", "id", "result", "state"], `category ${category.id || "unknown"}`);
    assert.ok(["covered", "partial", "missing"].includes(category.state));
    assert.ok(Array.isArray(category.evidence_files) && Array.isArray(category.blockers));
    assert.equal(new Set(category.blockers).size, category.blockers.length, `duplicate blocker in ${category.id}`);
    for (const blocker of category.blockers) assert.ok(gates.blockers.includes(blocker), `${category.id} references undeclared blocker ${blocker}`);
    if (category.state === "covered") {
      assert.ok(category.command && category.evidence_files.length > 0, `covered category ${category.id} lacks executable evidence`);
      assert.deepEqual(category.blockers, [], `covered category ${category.id} retains blockers`);
      exactKeys(category.result, ["discovered_tests", "executed_tests", "exit_code", "skipped_tests"], `${category.id} result`);
      assert.equal(category.result.exit_code, 0, `${category.id} result failed`);
      assert.ok(category.result.discovered_tests > 0 && category.result.executed_tests === category.result.discovered_tests, `${category.id} result has zero or incomplete execution`);
      assert.equal(category.result.skipped_tests, 0, `${category.id} result has skipped tests`);
    } else {
      assert.ok(category.blockers.length > 0, `${category.state} category ${category.id} lacks blockers`);
      assert.equal(category.result, null, `${category.state} category ${category.id} must not claim a passing result`);
      if (category.state === "missing") assert.equal(category.command, null, `missing category ${category.id} must not claim a command`);
    }
    if (category.command !== null) assert.ok(category.command.trim().length > 0, `${category.id} has an empty command`);
    for (const path of category.evidence_files) {
      assert.ok(!isAbsolute(path) && !path.split(/[\\/]/).includes(".."), `evidence path escapes repository: ${path}`);
      const absolutePath = resolve(rootDir, path);
      assert.ok(existsSync(absolutePath) && statSync(absolutePath).isFile(), `missing evidence file: ${path}`);
    }
  }

  const referencedBlockers = new Set([...gates.external_blockers.map((blocker) => blocker.id), ...gates.categories.flatMap((category) => category.blockers)]);
  assert.deepEqual([...new Set(gates.blockers)].sort(), [...referencedBlockers].sort(), "global blockers disagree with referenced blockers");

  const complete = gates.categories.every((category) => category.state === "covered") && gates.blockers.length === 0;
  assert.equal(gates.status === "complete", complete, "security gate completion status disagrees with evidence");
  if (options.enforce) assert.equal(complete, true, "AI/biometric security gate enforcement rejected blocked or incomplete evidence");
  return true;
}

module.exports = { validateSecurityGates };

if (require.main === module) {
  const gates = JSON.parse(readFileSync(gatePath, "utf8"));
  validateSecurityGates(gates, { enforce: process.argv.includes("--enforce") });
  console.log(`AI/biometric security gates: ${gates.status}`);
}