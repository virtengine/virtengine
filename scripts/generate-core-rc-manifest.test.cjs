"use strict";

const assert = require("assert").strict;
const { readFileSync } = require("fs");
const { spawnSync } = require("child_process");
const { resolve } = require("path");
const {
  artifactSelections,
  assertUniqueIds,
  buildAiAssurance,
  buildArtifactGroup,
  buildTestEvidence,
  buildTooling,
  listSourceEntries,
  sourceArtifacts,
} = require("./generate-core-rc-manifest.cjs");
const { validateSchema } = require("./validate-core-rc-manifest.cjs");

const root = resolve(__dirname, "..");
const sourceSha = "38afcdae29a66a9952cffe08b616039c371068d6";
const handoffPath = "_docs/ralph/handoffs/prototype-integration/HANDOFF.yaml";
const checkedManifestPath = "_docs/ralph/prototype-integration/core-rc-manifest.json";

function clone(value) {
  return JSON.parse(JSON.stringify(value));
}

function sourceJson(path) {
  const result = spawnSync("git", ["show", `${sourceSha}:${path}`], { cwd: root, encoding: "utf8" });
  assert.equal(result.status, 0, result.stderr);
  return JSON.parse(result.stdout);
}

function runGenerator(args) {
  return spawnSync(process.execPath, [resolve(__dirname, "generate-core-rc-manifest.cjs"), ...args], { cwd: root, encoding: "utf8" });
}

const handoff = sourceJson(handoffPath);
const schema = JSON.parse(readFileSync(resolve(root, "_docs/ralph/prototype-integration/core-rc-manifest.schema.json"), "utf8"));
const modelProvenance = JSON.parse(readFileSync(resolve(root, "_docs/ralph/prototype-integration/model-provenance.json"), "utf8"));
const productionPolicy = JSON.parse(readFileSync(resolve(root, "_docs/ralph/prototype-integration/ai-production-policy.json"), "utf8"));
const featureParity = JSON.parse(readFileSync(resolve(root, "pkg/inference/conformance/testdata/feature_parity_v1.json"), "utf8"));

const tests = [
  ["projects blocked AI assurance without certification", () => {
    const assurance = buildAiAssurance(modelProvenance, productionPolicy, featureParity);
    assert.equal(assurance.feature_vector.dimension, 768);
    assert.equal(assurance.feature_vector.test_vector_hashes.length, 4);
    assert.equal(assurance.uniqueness.implementation_class, "truncated_salted_sha256_bucket_equality_not_lsh");
    assert.equal(assurance.vault_kms.kms_hsm, "not_configured_or_certified");
    assert.equal(assurance.consent_retention.production_retention_certified, false);
    assert.equal(assurance.non_certification.production_certified, false);
  }],
  ["binds the AI production policy as a control artifact", () => {
    assert.ok(sourceArtifacts.some(([id, path]) => id === "ai_production_policy" && path === "_docs/ralph/prototype-integration/ai-production-policy.json"));
  }],
  ["binds the AI biometric security gates as a control artifact", () => {
    assert.ok(sourceArtifacts.some(([id, path]) => id === "ai_biometric_security_gates" && path === "_docs/ralph/prototype-integration/ai-biometric-security-gates.json"));
  }],
  ["binds the fund route inventory as a control artifact", () => {
    assert.ok(sourceArtifacts.some(([id, path]) => id === "fund_route_inventory" && path === "_docs/ralph/prototype-integration/fund-route-inventory.json"));
  }],
  ["hashes every tracked canonical chart file", () => {
    const chartContract = artifactSelections.find((contract) => contract.id === "chart");
    const entries = listSourceEntries(sourceSha, root);
    const group = buildArtifactGroup(entries, chartContract);
    assert.equal(group.status, "available");
    assert.equal(group.file_count, 15);
    assert.deepEqual(group.selection.included_path_prefixes, ["deploy/slurm/slurm-cluster/"]);
    const changed = entries.map((entry) => entry.path.endsWith("templates/controller-statefulset.yaml") ? { ...entry, object: "0".repeat(40) } : entry);
    assert.notEqual(buildArtifactGroup(changed, chartContract).sha256, group.sha256, "non-Chart template changes must alter the chart digest");
  }],
  ["does not claim availability below the declared coverage contract", () => {
    const chartContract = artifactSelections.find((contract) => contract.id === "chart");
    const entries = listSourceEntries(sourceSha, root).filter((entry) => chartContract.matches(entry.path)).slice(1);
    const group = buildArtifactGroup(entries, chartContract);
    assert.equal(group.status, "partial");
    assert.equal(group.blocker_id, "artifact-selection-incomplete");
  }],
  ["keeps model artifacts partial despite matching the expected count", () => {
    const modelContract = artifactSelections.find((contract) => contract.id === "model");
    const group = buildArtifactGroup(listSourceEntries(sourceSha, root), modelContract);
    assert.equal(group.file_count, 13);
    assert.equal(group.status, "partial");
    assert.equal(group.blocker_id, "production-model-artifacts-unavailable");
  }],
  ["binds every evidence field to the implementation and ledger commits", () => {
    const evidence = buildTestEvidence(sourceSha, handoff, handoffPath, root);
    assert.equal(evidence.implementation_sha, handoff.end_head);
    assert.equal(evidence.ledger_sha, sourceSha);
    assert.equal(evidence.status, "partial");
    assert.ok(evidence.records.every((record) => Object.hasOwn(record, "exit_code") && Object.hasOwn(record, "test_count") && Object.hasOwn(record, "tool_versions")));
  }],
  ["rejects stale evidence with unrelated intervening changes", () => {
    const candidate = clone(handoff);
    candidate.end_head = "55600687e47c92ce23d6bbfb650f74819b5ec6a3";
    assert.throws(() => buildTestEvidence(sourceSha, candidate, handoffPath, root), /non-evidence paths/);
  }],
  ["rejects evidence from an unrelated implementation commit", () => {
    const candidate = clone(handoff);
    candidate.end_head = "62d8c4b99804b354b36031ea2fd8b609dcac8a87";
    assert.throws(() => buildTestEvidence(sourceSha, candidate, handoffPath, root), /must be an ancestor/);
  }],
  ["rejects nonzero evidence and missing tool versions", () => {
    const nonzero = clone(handoff);
    nonzero.tests[0].exit_code = 1;
    assert.throws(() => buildTestEvidence(sourceSha, nonzero, handoffPath, root), /nonzero or missing exit_code/);
    const versionless = clone(handoff);
    delete versionless.tests[0].tool_versions;
    assert.throws(() => buildTestEvidence(sourceSha, versionless, handoffPath, root), /missing tool_versions/);
  }],
  ["rejects checked-path dirty guard bypasses", () => {
    const common = ["--source", sourceSha, "--tooling-source", sourceSha];
    for (const args of [
      [...common, "--output", checkedManifestPath],
      [...common, "--check", "--output", "core-rc-manifest.tmp.json"],
    ]) {
      const result = runGenerator(args);
      assert.notEqual(result.status, 0);
      assert.match(result.stderr, /dirty worktree/);
    }
    const temporary = runGenerator([...common, "--output", "core-rc-manifest.tmp.json"]);
    assert.notEqual(temporary.status, 0);
    assert.doesNotMatch(temporary.stderr, /dirty worktree/);
    assert.match(temporary.stderr, /(?:does not exist|not) in/);
  }],
  ["blocks tooling provenance when the declared commit lacks tool blobs", () => {
    assert.throws(() => buildTooling(sourceSha, sourceSha, root), /(?:does not exist|not) in/);
  }],
  ["rejects mutated schema structure and duplicate schema IDs", () => {
    assert.doesNotThrow(() => validateSchema(schema));
    const mutated = clone(schema);
    mutated.$defs.testRecord.properties.command.bogus = true;
    assert.throws(() => validateSchema(mutated), /unknown schema keyword/);
    const duplicate = clone(schema);
    duplicate.$defs.artifactGroup.properties.id.enum.push("chart");
    assert.throws(() => validateSchema(duplicate), /must be unique/);
  }],
  ["rejects duplicate manifest entry IDs", () => {
    assert.throws(() => assertUniqueIds([{ id: "chart" }, { id: "chart" }], "artifact group"), /must be unique/);
  }],
];

for (const [name, run] of tests) {
  run();
  console.log(`ok - ${name}`);
}