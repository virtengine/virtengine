#!/usr/bin/env node

"use strict";

const assert = require("assert").strict;
const { existsSync, readFileSync, statSync } = require("fs");
const { isAbsolute, resolve } = require("path");

const root = resolve(__dirname, "..");
const inventoryPath = resolve(root, "_docs/ralph/prototype-integration/generated-contract-inventory.json");
const contractIds = ["stage-decision", "uniqueness-receipt", "eligibility-decision", "claim-presentation", "fund-authorization"];
const targetRoots = ["api/openapi", "sdk/artifacts/proto", "sdk/go/node", "sdk/ts/src/generated"];

function exactKeys(value, keys, label) {
  assert.deepEqual(Object.keys(value).sort(), [...keys].sort(), `${label} has unknown or missing fields`);
}

function sortedUnique(values, label) {
  assert.deepEqual(values, [...values].sort(), `${label} must be sorted`);
  assert.equal(new Set(values).size, values.length, `${label} must be unique`);
}

function insideRoots(path, roots) {
  return roots.some((rootPath) => path === rootPath || path.startsWith(`${rootPath}/`));
}

function isGeneratedTarget(path) {
  if (path.startsWith("api/openapi/")) return /\.(?:json|ya?ml)$/.test(path);
  if (path.startsWith("sdk/artifacts/proto/")) return /(?:\.binpb|\.sha256|\/inventory\.json)$/.test(path);
  if (path.startsWith("sdk/go/node/")) return /\.pb(?:\.gw)?\.go$/.test(path);
  if (path.startsWith("sdk/ts/src/generated/")) return path.endsWith(".ts");
  return false;
}

function isCompatibilityFixture(path) {
  return path.startsWith("tests/") || path.startsWith("testutil/") || path.includes("/testdata/");
}

function validateGeneratedContractInventory(inventory, options = {}) {
  const rootDir = options.rootDir || root;
  exactKeys(inventory, ["blockers", "checkpoint", "completion", "contracts", "generation_window", "schema_version", "status"], "generated contract inventory");
  assert.equal(inventory.schema_version, "virtengine.prototype.generated-contract-inventory/v1");
  assert.equal(inventory.checkpoint, "T4-13A");
  assert.ok(["dependency_blocked", "complete"].includes(inventory.status));
  assert.deepEqual(inventory.contracts.map((contract) => contract.id), contractIds);
  assert.equal(new Set(inventory.blockers).size, inventory.blockers.length, "blockers must be unique");

  exactKeys(inventory.generation_window, ["command", "mode", "result", "second_run_command", "source_roots", "state", "tracked_target_roots"], "generation window");
  assert.deepEqual(inventory.generation_window, {
    mode: "all",
    source_roots: ["sdk/proto/node", "sdk/proto/provider"],
    command: "./scripts/proto-generate.sh all",
    second_run_command: "./scripts/verify-proto-generation.sh",
    tracked_target_roots: targetRoots,
    state: inventory.generation_window.state,
    result: inventory.generation_window.result,
  });
  assert.ok(["closed", "completed"].includes(inventory.generation_window.state));
  if (inventory.generation_window.state === "closed") {
    assert.equal(inventory.generation_window.result, null, "closed generation window must not claim results");
  } else {
    assert.ok(inventory.generation_window.result && typeof inventory.generation_window.result === "object", "completed generation window requires results");
    exactKeys(inventory.generation_window.result, ["drift_clean", "evidence_path", "evidence_sha256", "first_run_exit_code", "second_run_exit_code", "source_sha"], "generation result");
    assert.match(inventory.generation_window.result.source_sha, /^[a-f0-9]{40}$/);
    assert.ok(!isAbsolute(inventory.generation_window.result.evidence_path) && !inventory.generation_window.result.evidence_path.split(/[\\/]/).includes(".."), "generation evidence path escapes repository");
    assert.match(inventory.generation_window.result.evidence_sha256, /^[a-f0-9]{64}$/);
    assert.equal(inventory.generation_window.result.first_run_exit_code, 0);
    assert.equal(inventory.generation_window.result.second_run_exit_code, 0);
    assert.equal(inventory.generation_window.result.drift_clean, true);
    assert.equal(options.verifyGenerationResult?.(inventory.generation_window.result), true, "generation result verification failed");
  }

  for (const contract of inventory.contracts) {
    exactKeys(contract, ["blockers", "compatibility_fixtures", "generated_targets", "id", "owner_thread", "producer", "proto_sources", "state"], `contract ${contract.id || "unknown"}`);
    assert.ok(["T1", "T2", "T5"].includes(contract.owner_thread));
    assert.ok(["absent", "generated", "precursor_only"].includes(contract.state));
    exactKeys(contract.producer, ["payload_sha", "status", "tag"], `${contract.id} producer`);
    sortedUnique(contract.proto_sources, `${contract.id} proto sources`);
    sortedUnique(contract.generated_targets, `${contract.id} generated targets`);
    sortedUnique(contract.compatibility_fixtures, `${contract.id} compatibility fixtures`);
    sortedUnique(contract.blockers, `${contract.id} blockers`);
    for (const blocker of contract.blockers) assert.ok(inventory.blockers.includes(blocker), `${contract.id} references undeclared blocker ${blocker}`);

    if (contract.producer.status === "unavailable") {
      assert.equal(contract.producer.tag, null);
      assert.equal(contract.producer.payload_sha, null);
    } else {
      assert.equal(contract.producer.status, "accepted");
      assert.match(contract.producer.tag, new RegExp(`^checkpoint/prototype-${contract.owner_thread.toLowerCase()}/[a-z0-9-]+$`));
      assert.match(contract.producer.payload_sha, /^[a-f0-9]{40}$/);
      assert.equal(options.verifyAcceptedProducer?.(contract), true, `${contract.id} accepted producer verification failed`);
    }

    if (contract.state === "generated") {
      assert.equal(contract.producer.status, "accepted", `${contract.id} generated without accepted producer`);
      assert.ok(contract.proto_sources.length > 0 && contract.generated_targets.length > 0 && contract.compatibility_fixtures.length > 0, `${contract.id} generated evidence is incomplete`);
      assert.deepEqual(contract.blockers, [], `${contract.id} generated contract retains blockers`);
      for (const source of contract.proto_sources) {
        assert.ok(insideRoots(source, inventory.generation_window.source_roots) && source.endsWith(".proto"), `${contract.id} proto source is outside canonical roots`);
        const path = resolve(rootDir, source);
        assert.ok(existsSync(path) && statSync(path).isFile(), `${contract.id} proto source is missing`);
      }
      for (const target of contract.generated_targets) {
        assert.ok(insideRoots(target, targetRoots), `${contract.id} target is outside canonical roots`);
        assert.ok(isGeneratedTarget(target), `${contract.id} target is not a generated artifact type`);
        const path = resolve(rootDir, target);
        assert.ok(existsSync(path) && statSync(path).isFile(), `${contract.id} generated target is missing`);
      }
      for (const fixture of contract.compatibility_fixtures) {
        assert.ok(!isAbsolute(fixture) && !fixture.split(/[\\/]/).includes(".."), `${contract.id} fixture escapes repository`);
        assert.ok(isCompatibilityFixture(fixture), `${contract.id} compatibility fixture is outside test fixture paths`);
        const path = resolve(rootDir, fixture);
        assert.ok(existsSync(path) && statSync(path).isFile(), `${contract.id} compatibility fixture is missing`);
      }
    } else {
      assert.ok(contract.blockers.length > 0, `${contract.id} incomplete contract lacks blockers`);
      assert.deepEqual(contract.proto_sources, [], `${contract.id} incomplete contract claims proto sources`);
      assert.deepEqual(contract.generated_targets, [], `${contract.id} incomplete contract claims generated targets`);
    }
  }

  const referencedBlockers = [...new Set(inventory.contracts.flatMap((contract) => contract.blockers))];
  assert.deepEqual(inventory.blockers, referencedBlockers, "root blockers must exactly match contract blocker usage");

  const complete = inventory.contracts.every((contract) => contract.state === "generated") && inventory.generation_window.state === "completed" && inventory.blockers.length === 0;
  assert.equal(inventory.completion.allowed, complete, "generated contract completion is premature");
  assert.equal(inventory.status === "complete", complete, "generated contract status disagrees with readiness");
  if (options.requireReady) assert.equal(complete, true, "generated contract integration is not ready");
  return true;
}

module.exports = { isCompatibilityFixture, isGeneratedTarget, validateGeneratedContractInventory };

if (require.main === module) {
  const inventory = JSON.parse(readFileSync(inventoryPath, "utf8"));
  validateGeneratedContractInventory(inventory, { requireReady: process.argv.includes("--require-ready") });
  console.log(`generated contract inventory: ${inventory.status}`);
}