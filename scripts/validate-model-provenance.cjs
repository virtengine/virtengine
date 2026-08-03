#!/usr/bin/env node

"use strict";

const assert = require("assert").strict;
const { createHash } = require("crypto");
const { readFileSync, realpathSync, statSync } = require("fs");
const { isAbsolute, resolve, sep } = require("path");

const root = resolve(__dirname, "..");
const manifestPath = resolve(root, "_docs/ralph/prototype-integration/model-provenance.json");
const schemaPath = resolve(root, "_docs/ralph/prototype-integration/model-provenance.schema.json");
const sha1Pattern = /^[a-f0-9]{40}$/;
const sha256Pattern = /^[a-f0-9]{64}$/;
const identifierPattern = /^[a-z0-9.-]+$/;
const knownStages = new Set(["preprocessing", "training", "evaluation", "packaging"]);
const rootKeys = ["schema_version", "manifest_id", "status", "source", "stages", "artifacts", "datasets", "licenses", "redistribution", "bindings", "sbom", "model_card", "evaluation_report", "blockers"];

function loadJson(path) {
  return JSON.parse(readFileSync(path, "utf8"));
}

function assertExactKeys(value, expected, label) {
  assert.ok(value && typeof value === "object" && !Array.isArray(value), `${label} must be an object`);
  assert.deepEqual(Object.keys(value).sort(), [...expected].sort(), `${label} has unexpected or missing properties`);
}

function rejectTrustOnFirstUse(value, label = "manifest") {
  if (!value || typeof value !== "object") return;
  for (const [key, child] of Object.entries(value)) {
    assert.ok(!/^(tofu|trust_on_first_use|learn_on_first_use)$/i.test(key), `${label} contains a trust-on-first-use flag: ${key}`);
    rejectTrustOnFirstUse(child, `${label}.${key}`);
  }
}

function validateSchemaControl(schema) {
  assert.equal(schema.$schema, "https://json-schema.org/draft/2020-12/schema");
  assert.equal(schema.additionalProperties, false, "provenance schema root must reject additional properties");
  assert.deepEqual(schema.required.slice().sort(), rootKeys.slice().sort(), "provenance schema root required fields must be exact");
  assert.equal(schema.properties.schema_version.const, "virtengine.model-provenance/v1");
  assert.deepEqual(schema.properties.status.enum, ["dependency_blocked", "fixture_only", "release_candidate", "production_approved"]);
  assert.equal(schema.$defs.blockedEvidence.additionalProperties, false);
  assert.deepEqual(Object.keys(schema.$defs.blockedEvidence.properties).sort(), ["blocker_id", "id", "state"]);
  assert.equal(schema.$defs.blockedEvidence.properties.state.const, "dependency_blocked");
  assert.deepEqual(schema.$defs.localEvidence.properties.state.enum, ["present", "fixture_only"]);

  function visit(node, location) {
    if (!node || typeof node !== "object") return;
    if (node.type === "object") {
      assert.equal(node.additionalProperties, false, `${location} must set additionalProperties false`);
      assert.ok(node.properties && node.required, `${location} must declare properties and required`);
      assert.deepEqual(node.required.slice().sort(), Object.keys(node.properties).sort(), `${location} must require every property`);
    }
    for (const [key, child] of Object.entries(node)) visit(child, `${location}.${key}`);
  }
  visit(schema, "schema");
}

function validateImmutableSource(source, evidence, label) {
  assertExactKeys(source, ["type", "uri", "revision", "digest"], `${label}.source`);
  assert.ok(["repository_file", "oci"].includes(source.type), `${label} has unsupported source type`);
  assert.equal(typeof source.uri, "string");
  assert.ok(source.uri.length > 0, `${label} source URI must not be empty`);
  assert.doesNotMatch(source.uri, /(?:refs\/heads|\/(?:tree|branch|branches|tags)\/|[?&](?:ref|branch|tag)=|[/:@._-](?:latest|main|master|head)(?:[/?#:@._-]|$))/i, `${label} source URI is mutable`);
  assert.match(source.digest, sha256Pattern, `${label} source digest must be lowercase SHA256`);
  assert.equal(source.digest, evidence.sha256, `${label} source digest must bind the local file digest`);

  if (source.type === "repository_file") {
    assert.match(source.uri, /^git\+https:\/\//, `${label} repository source must use git+https`);
    assert.match(source.revision, sha1Pattern, `${label} repository source must pin a full commit`);
  } else {
    assert.match(source.uri, /@sha256:[a-f0-9]{64}$/, `${label} OCI source must be digest-pinned, not tag-only`);
    assert.match(source.revision, /^sha256:[a-f0-9]{64}$/, `${label} OCI revision must be a digest`);
  }
}

function validateLocalPath(evidence, context, label) {
  assert.equal(typeof evidence.path, "string");
  assert.ok(evidence.path.length > 0, `${label} path must not be empty`);
  assert.ok(!isAbsolute(evidence.path), `${label} path must be repository-relative`);
  assert.ok(!evidence.path.includes("\\"), `${label} path must use forward slashes`);
  const segments = evidence.path.split("/");
  assert.ok(!segments.includes("..") && !segments.includes("."), `${label} path traversal is not allowed`);
  assert.ok(!context.paths.has(evidence.path), `duplicate evidence path: ${evidence.path}`);
  context.paths.add(evidence.path);

  const candidate = resolve(context.rootDir, ...segments);
  const canonicalRoot = realpathSync(context.rootDir);
  const canonicalPath = realpathSync(candidate);
  assert.ok(canonicalPath.startsWith(`${canonicalRoot}${sep}`), `${label} path escapes repository root`);
  const stat = statSync(canonicalPath);
  assert.ok(stat.isFile(), `${label} path must identify a file`);
  assert.ok(Number.isInteger(evidence.size) && evidence.size > 0, `${label} size must be greater than zero`);
  const localBytes = readFileSync(canonicalPath);
  assert.match(evidence.sha256, sha256Pattern, `${label} sha256 must be nonempty lowercase SHA256`);
  const candidates = [localBytes];
  if (evidence.source.type === "repository_file") {
    const lfText = localBytes.toString("utf8").replace(/\r\n/g, "\n");
    candidates.push(Buffer.from(lfText, "utf8"));
    candidates.push(Buffer.from(lfText.replace(/\n/g, "\r\n"), "utf8"));
  }
  const matches = candidates.some((candidate) => candidate.length === evidence.size
    && createHash("sha256").update(candidate).digest("hex") === evidence.sha256);
  assert.ok(matches, `${label} size or sha256 does not match local file`);
}

function validateEvidence(evidence, context, label) {
  assert.ok(evidence && typeof evidence === "object" && !Array.isArray(evidence), `${label} must be an evidence object`);
  assert.match(evidence.id, identifierPattern, `${label} has invalid evidence ID`);
  assert.ok(!context.ids.has(evidence.id), `duplicate evidence ID: ${evidence.id}`);
  context.ids.add(evidence.id);

  if (evidence.state === "dependency_blocked") {
    assertExactKeys(evidence, ["id", "state", "blocker_id"], label);
    assert.match(evidence.blocker_id, identifierPattern, `${label} has invalid blocker ID`);
    assert.ok(context.blockerIds.has(evidence.blocker_id), `${label} references unknown blocker: ${evidence.blocker_id}`);
    context.blocked += 1;
    return;
  }

  assert.ok(["present", "fixture_only"].includes(evidence.state), `${label} has invalid evidence state`);
  assertExactKeys(evidence, ["id", "state", "path", "sha256", "size", "source"], label);
  validateLocalPath(evidence, context, label);
  validateImmutableSource(evidence.source, evidence, label);
  if (evidence.state === "fixture_only") context.fixtures += 1;
}

function validateModelProvenance(manifest, options = {}) {
  const rootDir = options.rootDir || root;
  rejectTrustOnFirstUse(manifest);
  if (options.schema) validateSchemaControl(options.schema);
  assertExactKeys(manifest, rootKeys, "manifest");
  assert.equal(manifest.schema_version, "virtengine.model-provenance/v1");
  assert.match(manifest.manifest_id, identifierPattern, "manifest_id is invalid");
  assert.ok(["dependency_blocked", "fixture_only", "release_candidate", "production_approved"].includes(manifest.status), "invalid manifest status");
  assertExactKeys(manifest.source, ["commit", "tree"], "manifest.source");
  assert.equal(manifest.source.commit, "79391a3df86d85522b92e0400c6904971ecbe65d", "manifest must bind the campaign baseline commit");
  assert.equal(manifest.source.tree, "349820fab7aeeefc20079ac29b98f6ec6736f4fb", "manifest must bind the campaign baseline tree");

  assert.ok(Array.isArray(manifest.blockers), "blockers must be an array");
  const blockerIds = new Set();
  for (const blocker of manifest.blockers) {
    assertExactKeys(blocker, ["id", "description"], `blocker ${blocker.id || "<missing>"}`);
    assert.match(blocker.id, identifierPattern, "invalid blocker ID");
    assert.ok(!blockerIds.has(blocker.id), `duplicate blocker ID: ${blocker.id}`);
    blockerIds.add(blocker.id);
    assert.equal(typeof blocker.description, "string");
    assert.ok(blocker.description.length > 0, `${blocker.id} description must not be empty`);
  }

  const context = { rootDir, blockerIds, ids: new Set(), paths: new Set(), blocked: 0, fixtures: 0 };
  assert.ok(Array.isArray(manifest.stages) && manifest.stages.length > 0, "stages must not be empty");
  const stageIds = new Set();
  for (const stage of manifest.stages) {
    assertExactKeys(stage, ["id", "state", "evidence"], `stage ${stage.id || "<missing>"}`);
    assert.ok(knownStages.has(stage.id), `unknown stage: ${stage.id}`);
    assert.ok(!stageIds.has(stage.id), `duplicate stage: ${stage.id}`);
    stageIds.add(stage.id);
    assert.ok(["present", "fixture_only", "dependency_blocked"].includes(stage.state), `invalid stage state: ${stage.id}`);
    assert.equal(stage.state, stage.evidence.state, `${stage.id} stage state must match its evidence`);
    validateEvidence(stage.evidence, context, `stage ${stage.id} evidence`);
  }

  for (const collection of ["artifacts", "datasets", "licenses"]) {
    assert.ok(Array.isArray(manifest[collection]) && manifest[collection].length > 0, `${collection} must not be empty`);
    manifest[collection].forEach((evidence, index) => validateEvidence(evidence, context, `${collection}[${index}]`));
  }

  assertExactKeys(manifest.redistribution, ["approved", "evidence"], "redistribution");
  assert.equal(typeof manifest.redistribution.approved, "boolean", "redistribution approval must be explicit");
  validateEvidence(manifest.redistribution.evidence, context, "redistribution evidence");
  if (manifest.redistribution.approved) assert.notEqual(manifest.redistribution.evidence.state, "dependency_blocked", "approved redistribution requires evidence");

  assertExactKeys(manifest.bindings, ["preprocessing", "schema", "runtime"], "bindings");
  for (const binding of ["preprocessing", "schema", "runtime"]) validateEvidence(manifest.bindings[binding], context, `${binding} binding`);
  validateEvidence(manifest.sbom, context, "SBOM");
  validateEvidence(manifest.model_card, context, "model card");
  validateEvidence(manifest.evaluation_report, context, "evaluation report");

  if (context.blocked > 0) assert.equal(manifest.status, "dependency_blocked", "blocked evidence requires dependency_blocked status");
  if (manifest.status === "dependency_blocked") {
    assert.ok(context.blocked > 0, "dependency_blocked status requires blocked evidence");
    assert.ok(manifest.blockers.length > 0, "dependency_blocked status requires blockers");
  }
  if (["release_candidate", "production_approved"].includes(manifest.status)) {
    assert.equal(context.blocked, 0, `${manifest.status} cannot contain blocked evidence`);
    assert.equal(manifest.redistribution.approved, true, `${manifest.status} requires approved redistribution`);
  }
  if (manifest.status === "production_approved") assert.equal(context.fixtures, 0, "production_approved cannot rely on fixture evidence");
}

module.exports = { validateModelProvenance };

if (require.main === module) {
  validateModelProvenance(loadJson(manifestPath), { rootDir: root, schema: loadJson(schemaPath) });
  console.log("model provenance manifest: valid");
}