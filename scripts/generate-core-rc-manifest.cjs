#!/usr/bin/env node

"use strict";

const assert = require("assert").strict;
const { createHash } = require("crypto");
const { existsSync, readFileSync, realpathSync, writeFileSync } = require("fs");
const { isAbsolute, relative, resolve } = require("path");
const { spawnSync } = require("child_process");
const { validateEpochSequence } = require("./prototype-intake-epochs.cjs");

const root = resolve(__dirname, "..");
const manifestRelativePath = "_docs/ralph/prototype-integration/core-rc-manifest.json";
const schemaRelativePath = "_docs/ralph/prototype-integration/core-rc-manifest.schema.json";
const generatorRelativePath = "scripts/generate-core-rc-manifest.cjs";
const validatorRelativePath = "scripts/validate-core-rc-manifest.cjs";
const shaPattern = /^[a-f0-9]{40}$/;
const digestPattern = /^[a-f0-9]{64}$/;
const sourceDigestCache = new Map();
const sourceEntriesCache = new Map();
const sourceTextCache = new Map();

const sourceArtifacts = [
  ["migration_inventory", "_docs/ralph/prototype-integration/migration-inventory.json"],
  ["required_gate_matrix", "_docs/ralph/prototype-integration/required-gate-matrix.json"],
  ["slurm_inventory", "_docs/ralph/prototype-integration/slurm-chart-inventory.json"],
  ["slurm_report", "_docs/ralph/prototype-integration/slurm-chart-semantic-report.json"],
  ["model_provenance", "_docs/ralph/prototype-integration/model-provenance.json"],
  ["ai_production_policy", "_docs/ralph/prototype-integration/ai-production-policy.json"],
  ["ai_biometric_security_gates", "_docs/ralph/prototype-integration/ai-biometric-security-gates.json"],
  ["fund_route_inventory", "_docs/ralph/prototype-integration/fund-route-inventory.json"],
  ["generated_contract_inventory", "_docs/ralph/prototype-integration/generated-contract-inventory.json"],
];

const artifactSelections = [
  { id: "lockfiles", prefixes: [], patterns: ["**/go.sum", "**/pnpm-lock.yaml", "**/package-lock.json"], exclusions: [], minimum: 11, expected: 11, matches: (path) => /(^|\/)(go\.sum|pnpm-lock\.yaml|package-lock\.json)$/.test(path) },
  { id: "modules", prefixes: [], patterns: ["**/go.mod", "**/package.json"], exclusions: [], minimum: 17, expected: 17, matches: (path) => /(^|\/)(go\.mod|package\.json)$/.test(path) },
  { id: "proto", prefixes: [], patterns: ["**/*.proto"], exclusions: [], minimum: 249, expected: 249, matches: (path) => path.endsWith(".proto") },
  { id: "openapi", prefixes: ["openapi/"], patterns: ["**/*openapi*.json", "**/*openapi*.yaml", "**/*openapi*.yml", "**/*swagger*.json", "**/*swagger*.yaml", "**/*swagger*.yml"], exclusions: [], minimum: 10, expected: 10, matches: (path) => /(^|\/)(openapi\/|[^/]*(openapi|swagger)[^/]*\.(json|ya?ml)$)/i.test(path) },
  { id: "sdk", prefixes: ["sdk/go/", "sdk/ts/src/generated/"], patterns: [], exclusions: [], minimum: 1508, expected: 1508, matches: (path) => /^sdk\/(go|ts\/src\/generated)\//.test(path) },
  { id: "model", prefixes: ["models/", "ml/training/model/"], patterns: [], exclusions: [], minimum: 13, expected: 13, forcePartial: true, matches: (path) => /^(models\/|ml\/training\/model\/)/.test(path) },
  { id: "chart", prefixes: ["deploy/slurm/slurm-cluster/"], patterns: [], exclusions: ["**/__pycache__/**", "**/.pytest_cache/**", "**/*.pyc"], minimum: 15, expected: 15, matches: (path) => path.startsWith("deploy/slurm/slurm-cluster/") && !/(^|\/)(__pycache__|\.pytest_cache)(\/|$)|\.pyc$/.test(path) },
];

const toolingArtifacts = [
  ["generator", generatorRelativePath],
  ["schema", schemaRelativePath],
  ["validator", validatorRelativePath],
];

const evidencePathPrefixes = [
  "_docs/ralph/handoffs/prototype-integration/",
  "_docs/ralph/prototype-integration/evidence/",
  "_docs/audits/",
];

function runGit(args, cwd = root, allowFailure = false) {
  const result = spawnSync("git", args, { cwd, encoding: "utf8", maxBuffer: 64 * 1024 * 1024 });
  if (!allowFailure && result.status !== 0) {
    throw new Error((result.stderr || result.stdout || `git ${args.join(" ")} failed`).trim());
  }
  return result;
}

function gitText(args, cwd = root) {
  return runGit(args, cwd).stdout.replace(/\r\n/g, "\n").trimEnd();
}

function sourceText(sourceSha, path, cwd = root) {
  const key = `${cwd}\0${sourceSha}\0${path}`;
  if (!sourceTextCache.has(key)) sourceTextCache.set(key, gitText(["show", `${sourceSha}:${path}`], cwd));
  return sourceTextCache.get(key);
}

function sourceJson(sourceSha, path, cwd = root) {
  return JSON.parse(sourceText(sourceSha, path, cwd));
}

function sha256(value) {
  return createHash("sha256").update(value).digest("hex");
}

function compareBytewise(left, right) {
  return Buffer.compare(Buffer.from(left, "utf8"), Buffer.from(right, "utf8"));
}

function sourceDigest(sourceSha, path, cwd = root) {
  const key = `${cwd}\0${sourceSha}\0${path}`;
  if (!sourceDigestCache.has(key)) {
    sourceDigestCache.set(key, sha256(Buffer.from(runGit(["show", `${sourceSha}:${path}`], cwd).stdout, "utf8")));
  }
  return sourceDigestCache.get(key);
}

function assertCommit(sha, label, cwd = root) {
  assert.match(sha, shaPattern, `${label} must be an exact lowercase 40-character commit SHA; mutable refs are forbidden`);
  assert.equal(gitText(["cat-file", "-t", sha], cwd), "commit", `${label} must resolve to a commit`);
}

function isAncestor(ancestor, descendant, cwd = root) {
  return runGit(["merge-base", "--is-ancestor", ancestor, descendant], cwd, true).status === 0;
}

function interveningChangedPaths(ancestor, descendant, cwd = root) {
  const commits = gitText(["log", "--format=%H", `${ancestor}..${descendant}`], cwd).split("\n").filter(Boolean);
  return commits.flatMap((commit) => gitText(["diff-tree", "-m", "--no-commit-id", "--name-only", "-r", commit], cwd).split("\n").filter(Boolean));
}

function sourceBlob(sourceSha, path, cwd = root) {
  const object = gitText(["rev-parse", `${sourceSha}:${path}`], cwd);
  assert.match(object, shaPattern, `invalid Git blob ID for ${path}`);
  assert.equal(gitText(["cat-file", "-t", object], cwd), "blob", `${path} must resolve to a committed blob`);
  return object;
}

function exactKeys(value, keys, label) {
  assert.deepEqual(Object.keys(value).sort(), [...keys].sort(), `${label} has unknown or missing fields`);
}

function blockerIds(manifest) {
  return new Set(manifest.blockers.map((blocker) => blocker.id));
}

function assertBlocker(manifest, blockerId, label) {
  assert.equal(typeof blockerId, "string", `${label} must name a blocker`);
  assert.ok(blockerIds(manifest).has(blockerId), `${label} references missing blocker ${blockerId}`);
}

function assertUniqueIds(entries, label) {
  const ids = entries.map((entry) => entry.id);
  assert.equal(new Set(ids).size, ids.length, `${label} IDs must be unique`);
}

function listSourceEntries(sourceSha, cwd = root) {
  const key = `${cwd}\0${sourceSha}`;
  if (sourceEntriesCache.has(key)) return sourceEntriesCache.get(key);
  const entries = gitText(["ls-tree", "-r", "--full-tree", sourceSha], cwd).split("\n").filter(Boolean).map((line) => {
    const [metadata, path] = line.split("\t", 2);
    const [mode, type, object] = metadata.split(" ");
    assert.ok(["blob", "commit"].includes(type), `unsupported source object type for ${path}`);
    assert.match(object, shaPattern, `invalid Git object ID for ${path}`);
    return { mode, object, path, type };
  }).filter((entry) => entry.type === "blob").map(({ mode, object, path }) => ({ mode, object, path }))
    .sort((left, right) => compareBytewise(left.path, right.path));
  sourceEntriesCache.set(key, entries);
  return entries;
}

function sourceArtifactsFor(sourceSha, cwd = root) {
  const entries = listSourceEntries(sourceSha, cwd);
  const epochs = entries
    .map((entry) => ({ entry, match: entry.path.match(/^_docs\/ralph\/prototype-integration\/epochs\/epoch-([1-9][0-9]*)\.json$/) }))
    .filter((item) => item.match)
    .map((item) => ({ file: `epoch-${item.match[1]}.json`, number: Number(item.match[1]), path: item.entry.path, document: sourceJson(sourceSha, item.entry.path, cwd) }))
    .sort((left, right) => left.number - right.number);
  const current = validateEpochSequence(epochs).at(-1);
  const observationPath = `_docs/ralph/prototype-integration/epochs/epoch-${current.number}-tag-observation.json`;
  const observationPresent = entries.some((entry) => entry.path === observationPath);
  if (current.document.status !== "open") assert.equal(observationPresent, true, `manifest current epoch ${current.number} requires a tag observation`);
  return [
    ...sourceArtifacts,
    ["intake_epoch", current.path],
    ...(observationPresent ? [["intake_tag_observation", observationPath]] : []),
  ];
}

function buildArtifactGroup(entries, contract) {
  const selected = entries.filter((entry) => contract.matches(entry.path));
  const coverageComplete = selected.length >= contract.minimum && selected.length === contract.expected;
  const selection = {
    included_path_prefixes: contract.prefixes,
    included_file_patterns: contract.patterns,
    excluded_path_patterns: contract.exclusions,
    minimum_file_count: contract.minimum,
    expected_file_count: contract.expected,
  };
  if (selected.length === 0) {
    return { id: contract.id, status: "unavailable", selection, file_count: 0, sha256: null, blocker_id: "artifact-selection-incomplete" };
  }
  const records = selected.map((entry) => `${entry.path}\0${entry.mode}\0${entry.object}\n`).join("");
  const partial = contract.forcePartial || !coverageComplete;
  return {
    id: contract.id,
    status: partial ? "partial" : "available",
    selection,
    file_count: selected.length,
    sha256: sha256(records),
    blocker_id: contract.forcePartial ? "production-model-artifacts-unavailable" : (partial ? "artifact-selection-incomplete" : null),
  };
}

function buildTooling(toolingSha, sourceSha, cwd = root) {
  assertCommit(toolingSha, "--tooling-source", cwd);
  assert.ok(isAncestor(sourceSha, toolingSha, cwd), "source SHA must be an ancestor of tooling SHA");
  const tree = gitText(["show", "-s", "--format=%T", toolingSha], cwd);
  return {
    tooling_sha: toolingSha,
    tree,
    artifacts: toolingArtifacts.map(([id, path]) => ({ id, path, git_blob: sourceBlob(toolingSha, path, cwd), sha256: sourceDigest(toolingSha, path, cwd) })),
  };
}

function buildTestEvidence(sourceSha, handoff, handoffPath, cwd = root) {
  assertCommit(handoff.end_head, "handoff end_head", cwd);
  assert.ok(isAncestor(handoff.end_head, sourceSha, cwd), "handoff end_head must be an ancestor of source SHA");
  const unrelated = interveningChangedPaths(handoff.end_head, sourceSha, cwd).filter((path) => !evidencePathPrefixes.some((prefix) => path.startsWith(prefix)));
  assert.deepEqual(unrelated, [], `source commits after handoff end_head change non-evidence paths: ${unrelated.join(", ")}`);
  const ledgerSha = gitText(["log", "-1", "--format=%H", sourceSha, "--", handoffPath], cwd);
  assert.match(ledgerSha, shaPattern, "handoff ledger commit is unavailable");
  const records = handoff.tests.map((test, index) => {
    assert.ok(typeof test.command === "string" && test.command.trim(), `test evidence ${index} is missing command`);
    assert.equal(test.exit_code, 0, `test evidence ${index} has nonzero or missing exit_code`);
    assert.equal(test.result, "passed", `test evidence ${index} did not pass`);
    assert.ok(test.tool_versions && typeof test.tool_versions === "object" && !Array.isArray(test.tool_versions), `test evidence ${index} is missing tool_versions`);
    assert.ok(Object.keys(test.tool_versions).length > 0, `test evidence ${index} is missing tool_versions`);
    for (const version of Object.values(test.tool_versions)) assert.ok(typeof version === "string" && version.trim(), `test evidence ${index} has an empty tool version`);
    assert.ok(test.test_count === undefined || (Number.isInteger(test.test_count) && test.test_count >= 0), `test evidence ${index} has an invalid test_count`);
    return { command: test.command, exit_code: test.exit_code, result: test.result, test_count: test.test_count ?? null, tool_versions: test.tool_versions };
  });
  const counted = records.filter((test) => Number.isInteger(test.test_count));
  return {
    status: counted.length === records.length ? "complete" : "partial",
    ledger_path: handoffPath,
    implementation_sha: handoff.end_head,
    ledger_sha: ledgerSha,
    record_count: records.length,
    declared_test_count: counted.reduce((total, test) => total + test.test_count, 0),
    uncounted_record_count: records.length - counted.length,
    records,
    blocker_id: "test-evidence-partial",
  };
}

function buildToolchains(gateMatrix) {
  const declarations = new Map();
  for (const category of gateMatrix.categories) {
    for (const tool of category.pinned_tools) {
      const key = `${tool.name}\0${tool.version}\0${tool.source}`;
      declarations.set(key, { name: tool.name, version: tool.version, source: tool.source, status: "declared" });
    }
  }
  return [...declarations.values()].sort((left, right) =>
    compareBytewise(`${left.name}\0${left.version}\0${left.source}`, `${right.name}\0${right.version}\0${right.source}`));
}

function provenanceDigest(entry) {
  return {
    id: entry.id,
    state: entry.state,
    sha256: entry.sha256 || null,
    source_blocker_id: entry.blocker_id || null,
  };
}

function buildAiAssurance(modelProvenance, productionPolicy, featureParity) {
  const policyRules = new Set(productionPolicy.known_findings.map((finding) => finding.rule));
  const runtime = modelProvenance.bindings.runtime;
  const schema = modelProvenance.bindings.schema;
  return {
    status: "dependency_blocked",
    provenance_digests: {
      models: modelProvenance.artifacts.map(provenanceDigest),
      runtime: {
        state: runtime.state,
        sha256: runtime.sha256 || null,
        source_blocker_id: runtime.blocker_id || null,
        sbom_sha256: modelProvenance.sbom.sha256 || null,
        sbom_blocker_id: modelProvenance.sbom.blocker_id || null,
      },
      schema: {
        state: schema.state,
        sha256: schema.sha256 || null,
        source_blocker_id: schema.blocker_id || null,
      },
      licenses: modelProvenance.licenses.map(provenanceDigest),
    },
    feature_vector: {
      state: "fixture_only",
      dimension: featureParity.layout.total_dimension,
      ...featureParity.layout.encoding,
      fixture_path: "pkg/inference/conformance/testdata/feature_parity_v1.json",
      test_vector_hashes: featureParity.cases.map((testCase) => ({ id: testCase.name, sha256: testCase.expected_vector_hash })),
    },
    evaluation: {
      status: modelProvenance.evaluation_report.state,
      report_sha256: modelProvenance.evaluation_report.sha256 || null,
      source_blocker_id: modelProvenance.evaluation_report.blocker_id || null,
    },
    uniqueness: {
      status: "blocked",
      implementation_class: policyRules.has("fake-biometric-lsh") ? "truncated_salted_sha256_bucket_equality_not_lsh" : "unverified",
      evidence_path: "x/veid/keeper/biometric_hash.go",
    },
    vault_kms: {
      status: "blocked",
      blob_backend: policyRules.has("memory-vault") ? "process_memory" : "unverified",
      key_custody: policyRules.has("memory-vault") ? "process_memory" : "unverified",
      kms_hsm: "not_configured_or_certified",
    },
    consent_retention: {
      status: "blocked",
      consent: policyRules.has("allow-all-consent") ? "allow_all_default" : "unverified",
      retention: "implementation_present_production_uncertified",
      production_retention_certified: false,
    },
    non_certification: {
      production_certified: false,
      not_certified: ["production_model", "production_runtime", "production_evaluation", "biometric_uniqueness", "durable_vault_or_kms", "production_consent_enforcement", "production_retention_legal_hold_or_erasure"],
      blocker_id: "ai-production-assurance-unavailable",
    },
  };
}

function generateManifest(sourceSha, options = {}) {
  const cwd = options.rootDir || root;
  assertCommit(sourceSha, "--source", cwd);
  assert.ok(options.toolingSha, "--tooling-source <exact-commit-sha> is required");

  const sourceTree = gitText(["show", "-s", "--format=%T", sourceSha], cwd);
  assert.match(sourceTree, shaPattern, "source tree must resolve exactly");
  const control = sourceJson(sourceSha, "_docs/ralph/prototype-integration/control.json", cwd);
  const handoffPath = "_docs/ralph/handoffs/prototype-integration/HANDOFF.yaml";
  const handoff = sourceJson(sourceSha, handoffPath, cwd);
  const migration = sourceJson(sourceSha, sourceArtifacts[0][1], cwd);
  const gates = sourceJson(sourceSha, sourceArtifacts[1][1], cwd);
  const slurmInventory = sourceJson(sourceSha, sourceArtifacts[2][1], cwd);
  const slurmReport = sourceJson(sourceSha, sourceArtifacts[3][1], cwd);
  const modelProvenance = sourceJson(sourceSha, sourceArtifacts[4][1], cwd);
  const productionPolicy = sourceJson(sourceSha, sourceArtifacts[5][1], cwd);
  const featureParity = sourceJson(sourceSha, "pkg/inference/conformance/testdata/feature_parity_v1.json", cwd);
  const entries = listSourceEntries(sourceSha, cwd);
  const controlArtifacts = sourceArtifactsFor(sourceSha, cwd);
  const tooling = buildTooling(options.toolingSha, sourceSha, cwd);
  const testEvidence = buildTestEvidence(sourceSha, handoff, handoffPath, cwd);

  const manifest = {
    schema_version: "virtengine.core-rc.prototype/v0",
    manifest_id: "T4-08A",
    status: "dependency_blocked",
    authoritative: false,
    planned_functionality_complete: false,
    milestone_m_eligible: false,
    source: {
      payload_sha: sourceSha,
      tree: sourceTree,
      branch: control.integration.branch,
      binding: "declared_clean_parent_payload",
    },
    tooling,
    toolchains: buildToolchains(gates),
    artifact_groups: artifactSelections.map((contract) => buildArtifactGroup(entries, contract)),
    control_artifacts: controlArtifacts.map(([id, path]) => {
      const document = sourceJson(sourceSha, path, cwd);
      return { id, path, sha256: sourceDigest(sourceSha, path, cwd), status: document.status || (document.passed ? "passed" : "blocked") };
    }),
    migrations: {
      status: migration.status,
      inventory_path: sourceArtifacts[0][1],
      migration_count: migration.migrations.length,
      unavailable_producer_count: migration.producers.filter((producer) => producer.status !== "accepted").length,
      blocker_id: "producer-migration-handoffs-unavailable",
    },
    required_gates: {
      status: gates.status,
      matrix_path: sourceArtifacts[1][1],
      category_count: gates.categories.length,
      blocked_category_count: gates.categories.filter((category) => category.status !== "passed").length,
      blocker_id: "required-gates-dependency-blocked",
    },
    slurm: {
      status: slurmInventory.status,
      inventory_path: sourceArtifacts[2][1],
      report_path: sourceArtifacts[3][1],
      report_status: slurmReport.passed ? "passed" : "blocked",
      blocker_id: "slurm-production-evidence-unavailable",
    },
    model_provenance: {
      status: modelProvenance.status,
      path: sourceArtifacts[4][1],
      sbom_status: modelProvenance.sbom.state,
      production_weights_status: modelProvenance.artifacts.find((artifact) => artifact.id === "trust-score-model-weights").state,
      blocker_id: "production-model-provenance-unavailable",
    },
    ai_assurance: buildAiAssurance(modelProvenance, productionPolicy, featureParity),
    test_evidence: testEvidence,
    producer_checkpoints: {
      ledger_path: handoffPath,
      ledger_sha256: sourceDigest(sourceSha, handoffPath, cwd),
      accepted: handoff.accepted_checkpoints,
      rejected: handoff.rejected_checkpoints,
      blocker_id: "accepted-producer-checkpoints-unavailable",
    },
    rollout: { status: "not_authorized", evidence: null, blocker_id: "rollout-not-authorized" },
    rollback: { status: "unverified", evidence: null, blocker_id: "rollback-evidence-unavailable" },
    external_dependencies: [
      { id: "producer-checkpoints", status: "unavailable", blocker_id: "accepted-producer-checkpoints-unavailable" },
      { id: "sbom-and-release-provenance", status: "unavailable", blocker_id: "release-sbom-provenance-unavailable" },
      { id: "production-rollout", status: "unavailable", blocker_id: "rollout-not-authorized" },
    ],
    blockers: [
      { id: "accepted-producer-checkpoints-unavailable", description: "No producer checkpoint is accepted by the payload ledger." },
      { id: "artifact-selection-incomplete", description: "One or more artifact selections do not meet their declared coverage contract." },
      { id: "producer-migration-handoffs-unavailable", description: "Committed producer migration handoffs are unavailable." },
      { id: "required-gates-dependency-blocked", description: "The required gate matrix remains dependency-blocked." },
      { id: "slurm-production-evidence-unavailable", description: "SLURM production render, isolation, and live durability evidence remain unavailable." },
      { id: "production-model-artifacts-unavailable", description: "Production model weights and release artifacts are unavailable." },
      { id: "production-model-provenance-unavailable", description: "Production model provenance, approvals, evaluation, and runtime SBOM are unavailable." },
      { id: "ai-production-assurance-unavailable", description: "AI, biometric uniqueness, vault/KMS, consent, retention, and production evaluation assurance remain unavailable." },
      { id: "test-evidence-partial", description: "Some passing handoff test records do not declare test counts." },
      { id: "release-sbom-provenance-unavailable", description: "No release SBOM or signed release provenance is available." },
      { id: "rollout-not-authorized", description: "This non-authoritative prototype manifest does not authorize rollout." },
      { id: "rollback-evidence-unavailable", description: "No production rollback execution evidence is available." },
    ],
  };
  validateManifest(manifest, { rootDir: cwd });
  return manifest;
}

function validateManifest(manifest, options = {}) {
  const cwd = options.rootDir || root;
  exactKeys(manifest, ["schema_version", "manifest_id", "status", "authoritative", "planned_functionality_complete", "milestone_m_eligible", "source", "tooling", "toolchains", "artifact_groups", "control_artifacts", "migrations", "required_gates", "slurm", "model_provenance", "ai_assurance", "test_evidence", "producer_checkpoints", "rollout", "rollback", "external_dependencies", "blockers"], "manifest");
  assert.equal(manifest.schema_version, "virtengine.core-rc.prototype/v0");
  assert.equal(manifest.manifest_id, "T4-08A");
  assert.equal(manifest.status, "dependency_blocked");
  assert.equal(manifest.authoritative, false, "prototype manifest must never be authoritative");
  assert.equal(manifest.planned_functionality_complete, false, "prototype manifest must not claim planned functionality complete");
  assert.equal(manifest.milestone_m_eligible, false, "prototype manifest must not claim milestone M eligibility");
  assert.ok(Array.isArray(manifest.blockers) && manifest.blockers.length > 0, "blocked manifest must include blockers");
  for (const blocker of manifest.blockers) {
    exactKeys(blocker, ["id", "description"], `blocker ${blocker.id || "unknown"}`);
    assert.ok(blocker.id && blocker.description);
  }
  assert.equal(blockerIds(manifest).size, manifest.blockers.length, "blocker IDs must be unique");

  exactKeys(manifest.source, ["payload_sha", "tree", "branch", "binding"], "source");
  assert.match(manifest.source.payload_sha, shaPattern);
  assert.match(manifest.source.tree, shaPattern);
  assert.equal(manifest.source.binding, "declared_clean_parent_payload");
  assert.equal(gitText(["cat-file", "-t", manifest.source.payload_sha], cwd), "commit");
  assert.equal(gitText(["show", "-s", "--format=%T", manifest.source.payload_sha], cwd), manifest.source.tree, "source tree hash mismatch");
  const control = sourceJson(manifest.source.payload_sha, "_docs/ralph/prototype-integration/control.json", cwd);
  assert.equal(manifest.source.branch, control.integration.branch, "source branch binding mismatch");

  exactKeys(manifest.tooling, ["tooling_sha", "tree", "artifacts"], "tooling");
  assert.deepEqual(manifest.tooling, buildTooling(manifest.tooling.tooling_sha, manifest.source.payload_sha, cwd), "tooling provenance mismatch");
  const toolingIds = new Set();
  for (const artifact of manifest.tooling.artifacts) {
    exactKeys(artifact, ["id", "path", "git_blob", "sha256"], `tooling artifact ${artifact.id || "unknown"}`);
    assert.ok(!toolingIds.has(artifact.id), `duplicate tooling artifact ID ${artifact.id}`);
    toolingIds.add(artifact.id);
  }

  const controlArtifacts = sourceArtifactsFor(manifest.source.payload_sha, cwd);
  const gateMatrix = sourceJson(manifest.source.payload_sha, controlArtifacts[1][1], cwd);
  assert.deepEqual(manifest.toolchains, buildToolchains(gateMatrix), "toolchain declarations do not match the source gate matrix");
  const repeat = new Set();
  for (const tool of manifest.toolchains) {
    exactKeys(tool, ["name", "version", "source", "status"], `toolchain ${tool.name || "unknown"}`);
    assert.equal(tool.status, "declared");
    const key = `${tool.name}\0${tool.version}\0${tool.source}`;
    assert.ok(!repeat.has(key), "duplicate toolchain declaration");
    repeat.add(key);
  }

  const sourceEntries = listSourceEntries(manifest.source.payload_sha, cwd);
  const expectedGroups = artifactSelections.map((contract) => buildArtifactGroup(sourceEntries, contract));
  assert.deepEqual(manifest.artifact_groups, expectedGroups, "artifact group count or hash mismatch");
  assertUniqueIds(manifest.artifact_groups, "artifact group");
  for (const group of manifest.artifact_groups) {
    exactKeys(group, ["id", "status", "selection", "file_count", "sha256", "blocker_id"], `artifact group ${group.id || "unknown"}`);
    exactKeys(group.selection, ["included_path_prefixes", "included_file_patterns", "excluded_path_patterns", "minimum_file_count", "expected_file_count"], `artifact group ${group.id} selection`);
    assert.ok(["available", "partial", "unavailable"].includes(group.status));
    if (group.status === "available") {
      assert.equal(group.blocker_id, null);
      assert.equal(group.file_count, group.selection.expected_file_count, `${group.id} available status requires expected coverage`);
      assert.ok(group.file_count >= group.selection.minimum_file_count, `${group.id} available status requires minimum coverage`);
    }
    else assertBlocker(manifest, group.blocker_id, `artifact group ${group.id}`);
    if (group.status === "unavailable") assert.equal(group.sha256, null);
    else assert.match(group.sha256, digestPattern);
  }

  assert.deepEqual(manifest.control_artifacts.map((artifact) => [artifact.id, artifact.path]), controlArtifacts, "control artifact inventory mismatch");
  for (const artifact of manifest.control_artifacts) {
    exactKeys(artifact, ["id", "path", "sha256", "status"], `control artifact ${artifact.id || "unknown"}`);
    assert.match(artifact.sha256, digestPattern);
    assert.equal(sourceDigest(manifest.source.payload_sha, artifact.path, cwd), artifact.sha256, `${artifact.path} hash mismatch`);
    const document = sourceJson(manifest.source.payload_sha, artifact.path, cwd);
    assert.equal(artifact.status, document.status || (document.passed ? "passed" : "blocked"), `${artifact.path} status mismatch`);
  }

  exactKeys(manifest.migrations, ["status", "inventory_path", "migration_count", "unavailable_producer_count", "blocker_id"], "migrations");
  exactKeys(manifest.required_gates, ["status", "matrix_path", "category_count", "blocked_category_count", "blocker_id"], "required_gates");
  exactKeys(manifest.slurm, ["status", "inventory_path", "report_path", "report_status", "blocker_id"], "slurm");
  exactKeys(manifest.model_provenance, ["status", "path", "sbom_status", "production_weights_status", "blocker_id"], "model_provenance");
  assertBlocker(manifest, manifest.ai_assurance.non_certification.blocker_id, "AI assurance");
  exactKeys(manifest.test_evidence, ["status", "ledger_path", "implementation_sha", "ledger_sha", "record_count", "declared_test_count", "uncounted_record_count", "records", "blocker_id"], "test_evidence");
  exactKeys(manifest.producer_checkpoints, ["ledger_path", "ledger_sha256", "accepted", "rejected", "blocker_id"], "producer_checkpoints");
  exactKeys(manifest.rollout, ["status", "evidence", "blocker_id"], "rollout");
  exactKeys(manifest.rollback, ["status", "evidence", "blocker_id"], "rollback");
  const linkedSections = [manifest.migrations, manifest.required_gates, manifest.slurm, manifest.model_provenance, manifest.test_evidence, manifest.producer_checkpoints, manifest.rollout, manifest.rollback];
  linkedSections.forEach((section, index) => assertBlocker(manifest, section.blocker_id, `blocked section ${index}`));
  for (const dependency of manifest.external_dependencies) {
    exactKeys(dependency, ["id", "status", "blocker_id"], `external dependency ${dependency.id || "unknown"}`);
    assert.notEqual(dependency.status, "available");
    assertBlocker(manifest, dependency.blocker_id, `external dependency ${dependency.id}`);
  }
  assertUniqueIds(manifest.external_dependencies, "external dependency");

  const handoff = sourceJson(manifest.source.payload_sha, manifest.producer_checkpoints.ledger_path, cwd);
  const migration = sourceJson(manifest.source.payload_sha, manifest.migrations.inventory_path, cwd);
  assert.equal(manifest.migrations.status, migration.status, "migration status mismatch");
  assert.equal(manifest.migrations.migration_count, migration.migrations.length, "migration count mismatch");
  assert.equal(manifest.migrations.unavailable_producer_count, migration.producers.filter((producer) => producer.status !== "accepted").length, "migration producer count mismatch");
  assert.equal(manifest.required_gates.status, gateMatrix.status, "required gate status mismatch");
  assert.equal(manifest.required_gates.category_count, gateMatrix.categories.length, "required gate category count mismatch");
  assert.equal(manifest.required_gates.blocked_category_count, gateMatrix.categories.filter((category) => category.status !== "passed").length, "blocked gate category count mismatch");
  const slurmInventory = sourceJson(manifest.source.payload_sha, manifest.slurm.inventory_path, cwd);
  const slurmReport = sourceJson(manifest.source.payload_sha, manifest.slurm.report_path, cwd);
  assert.equal(manifest.slurm.status, slurmInventory.status, "SLURM inventory status mismatch");
  assert.equal(manifest.slurm.report_status, slurmReport.passed ? "passed" : "blocked", "SLURM report status mismatch");
  const modelProvenance = sourceJson(manifest.source.payload_sha, manifest.model_provenance.path, cwd);
  assert.equal(manifest.model_provenance.status, modelProvenance.status, "model provenance status mismatch");
  assert.equal(manifest.model_provenance.sbom_status, modelProvenance.sbom.state, "model SBOM status mismatch");
  assert.equal(manifest.model_provenance.production_weights_status, modelProvenance.artifacts.find((artifact) => artifact.id === "trust-score-model-weights").state, "model weights status mismatch");
  const productionPolicy = sourceJson(manifest.source.payload_sha, sourceArtifacts[5][1], cwd);
  const featureParity = sourceJson(manifest.source.payload_sha, "pkg/inference/conformance/testdata/feature_parity_v1.json", cwd);
  assert.deepEqual(manifest.ai_assurance, buildAiAssurance(modelProvenance, productionPolicy, featureParity), "AI assurance projection mismatch");
  const expectedEvidence = buildTestEvidence(manifest.source.payload_sha, handoff, manifest.test_evidence.ledger_path, cwd);
  assert.deepEqual(manifest.test_evidence, expectedEvidence, "test evidence binding mismatch");
  const expectedTestRecords = expectedEvidence.records;
  const countedTests = expectedTestRecords.filter((test) => Number.isInteger(test.test_count));
  assert.deepEqual(manifest.test_evidence.records, expectedTestRecords, "test evidence references mismatch");
  assert.equal(manifest.test_evidence.record_count, expectedTestRecords.length, "test evidence record count mismatch");
  assert.equal(manifest.test_evidence.declared_test_count, countedTests.reduce((total, test) => total + test.test_count, 0), "declared test count mismatch");
  assert.equal(manifest.test_evidence.uncounted_record_count, expectedTestRecords.length - countedTests.length, "uncounted test record count mismatch");
  assert.equal(sourceDigest(manifest.source.payload_sha, manifest.producer_checkpoints.ledger_path, cwd), manifest.producer_checkpoints.ledger_sha256, "producer ledger hash mismatch");
  assert.deepEqual(manifest.producer_checkpoints.accepted, handoff.accepted_checkpoints, "accepted producer ledger binding mismatch");
  assert.deepEqual(manifest.producer_checkpoints.rejected, handoff.rejected_checkpoints, "rejected producer ledger binding mismatch");
  assert.equal(manifest.producer_checkpoints.accepted.length, 0, "payload must not claim integrated producers");
  assert.notEqual(manifest.rollout.status, "complete");
  assert.notEqual(manifest.rollback.status, "verified");
  assert.notEqual(manifest.model_provenance.sbom_status, "available");
  return true;
}

function serialize(manifest) {
  return `${JSON.stringify(manifest, null, 2)}\n`;
}

function pathsReferToSameFile(left, right) {
  const normalize = (path) => process.platform === "win32" ? path.toLowerCase() : path;
  const canonical = (path) => normalize(existsSync(path) ? realpathSync.native(path) : resolve(path));
  return canonical(left) === canonical(right);
}

function parseArgs(argv) {
  const options = { check: false, output: manifestRelativePath, source: null, toolingSource: null };
  const seen = new Set();
  for (let index = 0; index < argv.length; index += 1) {
    const argument = argv[index];
    assert.equal(seen.has(argument), false, `duplicate argument: ${argument}`);
    seen.add(argument);
    if (argument === "--check") options.check = true;
    else if (["--source", "--tooling-source", "--output"].includes(argument)) {
      const value = argv[++index];
      assert.ok(value && !value.startsWith("--"), `${argument} requires a value`);
      const key = { "--source": "source", "--tooling-source": "toolingSource", "--output": "output" }[argument];
      options[key] = value;
    }
    else throw new Error(`unknown argument: ${argument}`);
  }
  assert.ok(options.source, "--source <exact-commit-sha> is required");
  assert.ok(options.toolingSource, "--tooling-source <exact-commit-sha> is required");
  assert.ok(options.output, "--output requires a path");
  return options;
}

function main(argv = process.argv.slice(2)) {
  const options = parseArgs(argv);
  const output = resolve(root, options.output);
  const checkedPath = resolve(root, manifestRelativePath);
  if (options.check || pathsReferToSameFile(output, checkedPath)) {
    assert.equal(gitText(["status", "--porcelain"], root), "", "refusing checked manifest access from a dirty worktree; use a temporary --output with exact source and tooling SHAs");
  }
  const generationOptions = { toolingSha: options.toolingSource };
  const first = serialize(generateManifest(options.source, generationOptions));
  const second = serialize(generateManifest(options.source, generationOptions));
  assert.equal(first, second, "nonreproducible manifest output");
  if (options.check) {
    assert.ok(existsSync(output), `checked manifest does not exist: ${relative(root, output)}`);
    assert.equal(readFileSync(output, "utf8").replace(/\r\n/g, "\n"), first, "checked manifest is not deterministic for the declared source SHA");
    console.log(`core RC manifest check: valid (${options.source})`);
  } else {
    writeFileSync(output, first, "utf8");
    console.log(`core RC manifest generated: ${isAbsolute(options.output) ? output : options.output}`);
  }
}

module.exports = { artifactSelections, assertUniqueIds, buildAiAssurance, buildArtifactGroup, buildTestEvidence, buildTooling, generateManifest, listSourceEntries, main, manifestRelativePath, parseArgs, pathsReferToSameFile, serialize, sourceArtifacts, sourceArtifactsFor, validateManifest };

if (require.main === module) main();