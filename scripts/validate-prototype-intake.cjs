#!/usr/bin/env node

"use strict";

const assert = require("assert").strict;
const { createHash } = require("crypto");
const { execFileSync } = require("child_process");
const { resolve } = require("path");

const CAMPAIGN = "three-day-prototype";
const FROZEN_BASELINE = "79391a3df86d85522b92e0400c6904971ecbe65d";
const PLANNING_SHA = "1436723bd78980aa0388dbe9fcfa24dda939c54a";
const SHA_PATTERN = /^[a-f0-9]{40}$/;
const REQUIRED_HANDOFF_KEYS = [
  "blockers", "branch", "campaign", "checkpoint", "commits_since_prior_acceptance",
  "external_evidence", "files_changed", "frozen_baseline", "generated_hashes",
  "intake_base_sha", "intake_epoch", "known_failures", "migrations",
  "next_checkpoint", "owned_paths", "payload_head", "planning_sha",
  "prior_accepted_payload", "tests", "thread", "tree_clean",
];
const BRANCHES = {
  T1: "ve/prototype-t1-identity",
  T2: "ve/prototype-t2-product",
  T3: "ve/prototype-t3-reliability",
  T5: "ve/prototype-t5-platform",
};

function git(repo, args, options = {}) {
  try {
    return execFileSync("git", ["-C", repo, ...args], {
      encoding: options.encoding === null ? null : "utf8",
      stdio: ["ignore", "pipe", "pipe"],
    }).toString().trim();
  } catch (error) {
    const detail = error.stderr ? error.stderr.toString().trim() : error.message;
    throw new Error(`git ${args.join(" ")} failed: ${detail}`);
  }
}

function gitSucceeds(repo, args) {
  try {
    execFileSync("git", ["-C", repo, ...args], { stdio: "ignore" });
    return true;
  } catch {
    return false;
  }
}

function gitShow(repo, revision, path, encoding = "utf8") {
  return execFileSync("git", ["-C", repo, "show", `${revision}:${path}`], {
    encoding: encoding === null ? null : encoding,
    stdio: ["ignore", "pipe", "pipe"],
  });
}

function parseJsonDocument(text, label) {
  try {
    return JSON.parse(text);
  } catch {
    throw new Error(`${label} must be JSON or JSON-compatible YAML; dependency-free intake does not accept ambiguous YAML`);
  }
}

function assertExactKeys(value, keys, label) {
  assert.ok(value && typeof value === "object" && !Array.isArray(value), `${label} must be an object`);
  assert.deepEqual(Object.keys(value).sort(), [...keys].sort(), `${label} has missing or unknown keys`);
}

function assertSha(value, label) {
  assert.match(value, SHA_PATTERN, `${label} must be a full lowercase SHA`);
}

function assertPath(value, label) {
  assert.equal(typeof value, "string", `${label} must be a string`);
  assert.ok(value.length > 0 && !value.startsWith("/") && !value.includes("\\"), `${label} must be repository-relative`);
  assert.ok(!value.split("/").includes(".."), `${label} must not traverse parents`);
}

function assertStringArray(value, label, { nonempty = false, sha = false } = {}) {
  assert.ok(Array.isArray(value), `${label} must be an array`);
  if (nonempty) assert.ok(value.length > 0, `${label} must not be empty`);
  assert.equal(new Set(value).size, value.length, `${label} must be unique`);
  for (const entry of value) {
    if (sha) assertSha(entry, label);
    else assertPath(entry, label);
  }
}

function validateSchemaContract(schema) {
  assert.equal(schema.$schema, "https://json-schema.org/draft/2020-12/schema");
  assert.equal(schema.additionalProperties, false);
  assert.deepEqual([...schema.required].sort(), REQUIRED_HANDOFF_KEYS);
  assert.equal(schema.properties.frozen_baseline.const, FROZEN_BASELINE);
  assert.equal(schema.properties.commits_since_prior_acceptance.minItems, 1);
  assert.equal(schema.$defs.testResult.additionalProperties, false);
  assert.deepEqual(schema.$defs.testResult.required, ["command", "exit_code", "result", "tool_versions", "artifact"]);
  assert.equal(schema.$defs.testResult.properties.exit_code.const, 0);
  assert.equal(schema.$defs.testResult.properties.result.const, "passed");
  assert.equal(schema.$defs.testResult.properties.tool_versions.minProperties, 1);
}

function validateHandoff(handoff, expected) {
  assertExactKeys(handoff, REQUIRED_HANDOFF_KEYS, "handoff");
  assert.equal(handoff.campaign, CAMPAIGN);
  assert.ok(Object.hasOwn(BRANCHES, handoff.thread), "handoff thread is unknown");
  assert.equal(handoff.thread, expected.thread, "tag and handoff thread differ");
  assert.match(handoff.checkpoint, new RegExp(`^${handoff.thread}-[0-9]{2,}[A-Z]?$`));
  assert.equal(handoff.checkpoint.toLowerCase(), expected.checkpoint, "tag and handoff checkpoint differ");
  assert.equal(handoff.branch, BRANCHES[handoff.thread], "branch does not match thread");
  assert.equal(handoff.frozen_baseline, FROZEN_BASELINE, "wrong frozen baseline");
  assertSha(handoff.planning_sha, "planning_sha");
  assert.equal(handoff.planning_sha, PLANNING_SHA, "wrong planning SHA");
  assert.ok(Number.isInteger(handoff.intake_epoch) && handoff.intake_epoch > 0, "intake_epoch must be a positive integer");
  assert.equal(handoff.intake_epoch, expected.epoch, "handoff is for a stale or different epoch");
  assertSha(handoff.intake_base_sha, "intake_base_sha");
  assertSha(handoff.payload_head, "payload_head");
  assert.ok(handoff.prior_accepted_payload === null || SHA_PATTERN.test(handoff.prior_accepted_payload), "prior_accepted_payload must be a SHA or null");
  assert.equal(handoff.tree_clean, true);
  assertStringArray(handoff.commits_since_prior_acceptance, "commits_since_prior_acceptance", { nonempty: true, sha: true });
  assertStringArray(handoff.owned_paths, "owned_paths", { nonempty: true });
  assertStringArray(handoff.files_changed, "files_changed", { nonempty: true });
  assert.ok(Array.isArray(handoff.tests) && handoff.tests.length > 0, "tests must not be empty");
  assert.ok(handoff.generated_hashes && typeof handoff.generated_hashes === "object" && !Array.isArray(handoff.generated_hashes));
  for (const [path, hash] of Object.entries(handoff.generated_hashes)) {
    assertPath(path, "generated hash path");
    assert.match(hash, /^[a-f0-9]{64}$/, "generated hash must be SHA-256");
  }
  for (const field of ["migrations", "external_evidence"]) assertStringArray(handoff[field], field);
  for (const field of ["known_failures", "blockers"]) {
    assert.ok(Array.isArray(handoff[field]));
    assert.equal(new Set(handoff[field]).size, handoff[field].length);
    for (const entry of handoff[field]) assert.ok(typeof entry === "string" && entry.length > 0);
  }
  assert.ok(handoff.next_checkpoint === null || new RegExp(`^${handoff.thread}-[0-9]{2,}[A-Z]?$`).test(handoff.next_checkpoint));

  for (const test of handoff.tests) {
    const required = ["artifact", "command", "exit_code", "result", "tool_versions"];
    const optional = test.test_count === undefined ? [] : ["test_count"];
    assertExactKeys(test, [...required, ...optional], "test result");
    assert.ok(typeof test.command === "string" && test.command.trim() === test.command && test.command.length > 0, "test command must be literal and nonempty");
    assert.equal(test.exit_code, 0);
    assert.equal(test.result, "passed");
    assert.ok(test.tool_versions && typeof test.tool_versions === "object" && !Array.isArray(test.tool_versions) && Object.keys(test.tool_versions).length > 0, "tool_versions must be a nonempty object");
    for (const tool of Object.keys(test.tool_versions)) assert.ok(tool.length > 0, "tool version names must be nonempty");
    for (const version of Object.values(test.tool_versions)) assert.ok(typeof version === "string" && version.length > 0);
    assertPath(test.artifact, "test artifact");
    if (test.test_count !== undefined) assert.ok(Number.isInteger(test.test_count) && test.test_count > 0, "test_count must be positive");
  }
}

function parseTag(tag) {
  const match = /^checkpoint\/prototype-(t[1235])\/(t[1235]-[0-9]{2,}[a-z]?)$/.exec(tag);
  assert.ok(match, "tag must use checkpoint/prototype-tN/tN-NN format");
  return { thread: match[1].toUpperCase(), checkpoint: match[2], tag };
}

function fetchAnnotatedTag(repo, remote, tag) {
  const listing = git(repo, ["ls-remote", "--tags", remote, `refs/tags/${tag}`, `refs/tags/${tag}^{}`]);
  const lines = listing.split(/\r?\n/).filter(Boolean);
  assert.equal(lines.length, 2, `remote annotated tag is missing or lightweight: ${tag}`);
  assert.ok(lines.some((line) => line.endsWith(`refs/tags/${tag}`)));
  assert.ok(lines.some((line) => line.endsWith(`refs/tags/${tag}^{}`)));
  git(repo, ["fetch", "--force", "--no-tags", remote, `refs/tags/${tag}:refs/tags/${tag}`]);
  assert.equal(git(repo, ["cat-file", "-t", `refs/tags/${tag}`]), "tag", `${tag} must be annotated`);
  return git(repo, ["rev-parse", `refs/tags/${tag}^{commit}`]);
}

function fetchBranch(repo, remote, branch) {
  const listing = git(repo, ["ls-remote", "--heads", remote, `refs/heads/${branch}`]);
  assert.ok(listing, `remote branch is missing: ${branch}`);
  const remoteRef = `refs/virtengine-intake/${branch.replaceAll("/", "-")}`;
  git(repo, ["fetch", "--force", "--no-tags", remote, `refs/heads/${branch}:${remoteRef}`]);
  return git(repo, ["rev-parse", remoteRef]);
}

function pathMatches(path, declaration) {
  if (declaration.endsWith("/**")) return path.startsWith(declaration.slice(0, -2));
  if (declaration.endsWith("/")) return path.startsWith(declaration);
  if (!declaration.includes("*")) return path === declaration;
  const globstar = "\u0000";
  const escaped = declaration
    .replaceAll("**", globstar)
    .replace(/[.+?^${}()|[\]\\]/g, "\\$&")
    .replaceAll("*", "[^/]*")
    .replaceAll(globstar, ".*");
  return new RegExp(`^${escaped}$`).test(path);
}

function lines(value) {
  return value ? value.split(/\r?\n/).filter(Boolean) : [];
}

function assertSameSet(actual, expected, label) {
  assert.deepEqual([...actual].sort(), [...expected].sort(), label);
}

function objectExistsAt(repo, revision, path) {
  return gitSucceeds(repo, ["cat-file", "-e", `${revision}:${path}`]);
}

function nonemptyBlobExistsAt(repo, revision, path) {
  if (!objectExistsAt(repo, revision, path)) return false;
  if (git(repo, ["cat-file", "-t", `${revision}:${path}`]) !== "blob") return false;
  return Number(git(repo, ["cat-file", "-s", `${revision}:${path}`])) > 0;
}

function completeChangedPaths(repo, range) {
  const changed = new Set();
  for (const commit of lines(git(repo, ["rev-list", "--reverse", "--topo-order", range]))) {
    for (const path of lines(git(repo, ["diff-tree", "--no-commit-id", "--name-only", "-r", "-m", commit]))) {
      changed.add(path);
    }
  }
  return changed;
}

function declaredNodeMajors(repo, revision) {
  const majors = new Set([20]);
  for (const path of [".node-version", ".nvmrc"]) {
    if (!objectExistsAt(repo, revision, path)) continue;
    const match = gitShow(repo, revision, path).trim().match(/^(?:v)?([0-9]+)(?:\.|$)/);
    if (match) majors.add(Number(match[1]));
  }
  if (objectExistsAt(repo, revision, "package.json")) {
    const packageJson = parseJsonDocument(gitShow(repo, revision, "package.json"), "package.json");
    const declaration = packageJson.engines && packageJson.engines.node;
    if (typeof declaration === "string") {
      for (const match of declaration.matchAll(/(?:^|[^0-9])([0-9]+)(?:\.|$)/g)) majors.add(Number(match[1]));
    }
  }
  return majors;
}

function validateToolchains(test, nodeMajors) {
  if (/(?:^|[\s;&|\\/])go(?:1\.25\.8)?(?:\.exe)?(?:\s|$)/i.test(test.command)) {
    assert.ok(Object.values(test.tool_versions).some((version) => /(?:go version )?go?1\.25\.8(?:\s|$)/i.test(version)), "Go commands require literal Go 1.25.8 version evidence");
  }
  if (/(?:^|[\s;&|\\/])(?:node|npm|npx|pnpm)(?:\.exe|\.cmd)?(?:\s|$)/i.test(test.command)) {
    assert.ok(Object.values(test.tool_versions).some((version) => {
      const match = version.match(/(?:^|\D)v?([0-9]+)(?:\.|$)/);
      return match && nodeMajors.has(Number(match[1]));
    }), "Node commands require Node 20 or repository-declared version evidence");
  }
}

function validateEpoch(epoch, requestedEpoch, tagInfo) {
  assertExactKeys(epoch, ["announcement_cutoff", "base_sha", "base_tag", "campaign", "intake_epoch", "opens_at", "planning_sha", "producers", "schema_version", "status"], "epoch manifest");
  assert.equal(epoch.schema_version, "virtengine.prototype.intake-epoch/v2");
  assert.equal(epoch.campaign, CAMPAIGN);
  assert.equal(epoch.intake_epoch, requestedEpoch);
  assert.equal(epoch.planning_sha, PLANNING_SHA, "epoch has wrong planning SHA");
  assert.equal(epoch.status, "frozen", "epoch announcement roster is not frozen");
  assert.equal(epoch.base_tag, `checkpoint/prototype-integration/epoch-${requestedEpoch}-base`);
  assertSha(epoch.base_sha, "epoch base_sha");
  assert.match(epoch.opens_at, /Z$/, "epoch opens_at must be UTC");
  assert.match(epoch.announcement_cutoff, /Z$/, "epoch announcement_cutoff must be UTC");
  assert.ok(Number.isFinite(Date.parse(epoch.opens_at)), "epoch opens_at must be a date-time");
  assert.ok(Number.isFinite(Date.parse(epoch.announcement_cutoff)), "epoch announcement_cutoff must be a date-time");
  assert.ok(Date.parse(epoch.opens_at) < Date.parse(epoch.announcement_cutoff), "epoch time window is invalid");
  assert.ok(Date.now() > Date.parse(epoch.announcement_cutoff), "epoch announcement cutoff has not elapsed");
  assert.ok(Array.isArray(epoch.producers), "epoch producers must be an array");
  assert.deepEqual(epoch.producers.map((entry) => entry.thread), ["T1", "T2", "T3", "T5"], "epoch producer roster is invalid");
  for (const entry of epoch.producers) {
    assertExactKeys(entry, ["decision", "status", "tag", "thread"], `epoch producer ${entry.thread}`);
    assert.ok(["unannounced", "announced", "accepted", "rejected"].includes(entry.status), `invalid epoch producer status for ${entry.thread}`);
    if (entry.status === "unannounced") {
      assert.equal(entry.tag, null);
      assert.equal(entry.decision, null);
    }
  }
  const producer = epoch.producers.find((entry) => entry.thread === tagInfo.thread);
  assert.ok(producer, "producer is unknown to epoch");
  assert.equal(producer.status, "announced", "producer tag was not announced for this epoch");
  assert.equal(producer.tag, tagInfo.tag, "epoch announced a different tag");
  assert.equal(producer.decision, null, "epoch tag already has a decision");
  return producer;
}

function validateIntake(options) {
  const repo = resolve(options.repo || resolve(__dirname, ".."));
  const remote = options.remote || "origin";
  const requestedEpoch = Number(options.epoch);
  assert.ok(Number.isInteger(requestedEpoch) && requestedEpoch > 0, "--epoch must be a positive integer");
  const tagInfo = parseTag(options.tag);

  assert.equal(git(repo, ["status", "--porcelain"]), "", "T4 worktree must be clean before intake");
  git(repo, ["fetch", "--prune", remote]);

  const epochPath = `_docs/ralph/prototype-integration/epochs/epoch-${requestedEpoch}.json`;
  assert.ok(objectExistsAt(repo, "HEAD", epochPath), `unknown intake epoch ${requestedEpoch}`);
  const epoch = parseJsonDocument(gitShow(repo, "HEAD", epochPath), "epoch manifest");
  validateEpoch(epoch, requestedEpoch, tagInfo);
  const schema = parseJsonDocument(gitShow(repo, "HEAD", "_docs/ralph/prototype-integration/producer-handoff.schema.json"), "producer schema");
  validateSchemaContract(schema);
  const control = parseJsonDocument(gitShow(repo, "HEAD", "_docs/ralph/prototype-integration/control.json"), "integration control");

  const baseTarget = fetchAnnotatedTag(repo, remote, epoch.base_tag);
  assert.equal(baseTarget, epoch.base_sha, "epoch base tag target does not match manifest");
  const tagTarget = fetchAnnotatedTag(repo, remote, options.tag);

  const handoffPath = `_docs/ralph/handoffs/prototype-${tagInfo.thread.toLowerCase()}/HANDOFF.yaml`;
  assert.ok(objectExistsAt(repo, tagTarget, handoffPath), `committed handoff is missing at tag: ${handoffPath}`);
  const handoff = parseJsonDocument(gitShow(repo, tagTarget, handoffPath), "committed producer handoff");
  validateHandoff(handoff, { ...tagInfo, epoch: requestedEpoch });
  assert.equal(handoff.intake_base_sha, epoch.base_sha, "handoff intake base does not match epoch");

  assert.ok(gitSucceeds(repo, ["cat-file", "-e", `${handoff.payload_head}^{commit}`]), "payload commit is missing");
  assert.ok(gitSucceeds(repo, ["merge-base", "--is-ancestor", epoch.base_sha, handoff.payload_head]), "payload does not descend from epoch base");
  assert.notEqual(tagTarget, handoff.payload_head, "tag must target a handoff commit after payload_head");
  assert.ok(gitSucceeds(repo, ["merge-base", "--is-ancestor", handoff.payload_head, tagTarget]), "tag target does not descend from payload_head");

  const handoffCommits = lines(git(repo, ["rev-list", "--reverse", "--ancestry-path", `${handoff.payload_head}..${tagTarget}`]));
  assert.ok(handoffCommits.length > 0, "handoff commit range is empty");
  let parent = handoff.payload_head;
  for (const commit of handoffCommits) {
    const parents = git(repo, ["show", "-s", "--format=%P", commit]).split(" ");
    assert.deepEqual(parents, [parent], "handoff commits must form a direct non-merge chain from payload_head");
    parent = commit;
  }

  const evidencePrefix = `_docs/ralph/evidence/prototype-${handoff.thread.toLowerCase()}/${handoff.checkpoint.toLowerCase()}/`;
  const retainedEvidence = [...handoff.external_evidence, ...handoff.tests.map((test) => test.artifact)];
  for (const path of retainedEvidence) {
    assert.ok(path.toLowerCase().startsWith(evidencePrefix), `retained evidence is outside checkpoint directory: ${path}`);
  }
  const evidencePaths = new Set([handoffPath, ...retainedEvidence]);
  const tagOnlyPaths = [...completeChangedPaths(repo, `${handoff.payload_head}..${tagTarget}`)];
  assert.ok(tagOnlyPaths.every((path) => evidencePaths.has(path)), "tag range changes undeclared handoff/evidence paths");
  assert.ok(tagOnlyPaths.includes(handoffPath), "tag range must commit the producer handoff");

  const ledger = parseJsonDocument(gitShow(repo, "HEAD", "_docs/ralph/handoffs/prototype-integration/HANDOFF.yaml"), "integration handoff ledger");
  const accepted = Array.isArray(ledger.accepted_checkpoints) ? ledger.accepted_checkpoints : [];
  assert.ok(!accepted.some((entry) => entry.thread === handoff.thread && (entry.checkpoint === handoff.checkpoint || entry.tip === tagTarget || entry.payload_head === handoff.payload_head)), "checkpoint or payload was already accepted");
  const acceptedForThread = accepted.filter((entry) => entry.thread === handoff.thread);
  const latestAccepted = acceptedForThread.length === 0 ? null : acceptedForThread[acceptedForThread.length - 1];
  const expectedPrior = latestAccepted === null ? null : latestAccepted.payload_head;
  assert.equal(handoff.prior_accepted_payload, expectedPrior, "prior_accepted_payload is not the latest accepted payload");

  const rangeBase = handoff.prior_accepted_payload || epoch.base_sha;
  if (handoff.prior_accepted_payload !== null) {
    assert.ok(gitSucceeds(repo, ["merge-base", "--is-ancestor", rangeBase, handoff.payload_head]), "prior accepted payload is not an ancestor");
  }
  const commits = lines(git(repo, ["rev-list", "--reverse", "--topo-order", `${rangeBase}..${handoff.payload_head}`]));
  assert.deepEqual(commits, handoff.commits_since_prior_acceptance, "declared payload commit list is incomplete or out of order");
  const changedFiles = [...completeChangedPaths(repo, `${rangeBase}..${handoff.payload_head}`)];
  assertSameSet(changedFiles, handoff.files_changed, "declared files_changed does not match complete payload range");
  for (const path of changedFiles) {
    assert.ok(handoff.owned_paths.some((owned) => pathMatches(path, owned)), `changed path is outside owned_paths: ${path}`);
    const integrationOwned = control.path_ownership.integration_only.some((owned) => pathMatches(path, owned));
    if (integrationOwned) {
      const lease = control.generated_file_lease;
      const leased = lease.state === "held" &&
        lease.holder === handoff.thread &&
        lease.checkpoint === handoff.checkpoint &&
        lease.base_sha === rangeBase &&
        typeof lease.expires_at === "string" &&
        Number.isFinite(Date.parse(lease.expires_at)) &&
        Date.parse(lease.expires_at) >= Date.parse(epoch.announcement_cutoff) &&
        lease.paths.some((owned) => pathMatches(path, owned));
      assert.ok(leased, `changed path is integration-owned without an active lease: ${path}`);
    }
  }

  const nodeMajors = declaredNodeMajors(repo, tagTarget);
  for (const test of handoff.tests) {
    validateToolchains(test, nodeMajors);
    assert.ok(nonemptyBlobExistsAt(repo, tagTarget, test.artifact), `test artifact is missing or empty at tag: ${test.artifact}`);
  }
  for (const path of handoff.external_evidence) assert.ok(nonemptyBlobExistsAt(repo, tagTarget, path), `external evidence is missing or empty at tag: ${path}`);
  for (const path of handoff.migrations) {
    assert.ok(objectExistsAt(repo, tagTarget, path), `declared migration is missing at tag: ${path}`);
    assert.ok(changedFiles.includes(path), `declared migration is not part of the payload range: ${path}`);
  }
  for (const [path, expectedHash] of Object.entries(handoff.generated_hashes)) {
    const content = gitShow(repo, tagTarget, path, null);
    assert.equal(createHash("sha256").update(content).digest("hex"), expectedHash, `generated hash mismatch: ${path}`);
  }

  const branchHead = fetchBranch(repo, remote, handoff.branch);
  assert.ok(gitSucceeds(repo, ["merge-base", "--is-ancestor", tagTarget, branchHead]), "remote producer branch does not contain checkpoint tag target");

  return { branchHead, checkpoint: handoff.checkpoint, payloadHead: handoff.payload_head, tagTarget, thread: handoff.thread };
}

function parseArgs(argv) {
  const options = {};
  for (let index = 0; index < argv.length; index += 2) {
    const key = argv[index];
    const value = argv[index + 1];
    assert.ok(["--epoch", "--tag", "--repo", "--remote"].includes(key) && value, `unknown or incomplete argument: ${key || "<none>"}`);
    options[key.slice(2)] = value;
  }
  assert.ok(options.epoch, "--epoch is required");
  assert.ok(options.tag, "--tag is required");
  return options;
}

module.exports = { completeChangedPaths, parseJsonDocument, validateHandoff, validateIntake, validateSchemaContract };

if (require.main === module) {
  try {
    const result = validateIntake(parseArgs(process.argv.slice(2)));
    console.log(`prototype intake: valid ${result.thread} ${result.checkpoint} ${result.tagTarget}`);
  } catch (error) {
    console.error(`prototype intake: invalid: ${error.message}`);
    process.exitCode = 1;
  }
}