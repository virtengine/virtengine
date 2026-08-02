#!/usr/bin/env node

"use strict";

const assert = require("assert").strict;
const { createHash, randomUUID } = require("crypto");
const { closeSync, fsyncSync, openSync, readFileSync, renameSync, unlinkSync, writeFileSync } = require("fs");
const { resolve } = require("path");
const { spawnSync } = require("child_process");
const { planFrozenEpoch, resolveAnnotatedTag, validateObservationBinding } = require("./plan-prototype-intake-freeze.cjs");

const threads = ["T1", "T2", "T3", "T5"];
const producerKeys = ["decision", "status", "tag", "thread"];

function createLeaseRecord(options, now = Date.now(), pid = process.pid) {
  assert.ok(Number.isInteger(pid) && pid > 0, "freeze lease PID is invalid");
  assert.ok(Number.isFinite(now), "freeze lease start time is invalid");
  assert.match(options.epoch || "", /^[1-9][0-9]*$/, "freeze lease epoch is invalid");
  assert.match(options.expectedHead || "", /^[a-f0-9]{40}$/, "freeze lease expected HEAD is invalid");
  assert.match(options.expectedPlanSha256 || "", /^[a-f0-9]{64}$/, "freeze lease plan digest is invalid");
  return {
    schema_version: "virtengine.prototype.intake-freeze-lease/v1",
    pid,
    started_at: new Date(now).toISOString(),
    epoch: Number(options.epoch),
    expected_head: options.expectedHead,
    plan_sha256: options.expectedPlanSha256,
  };
}

function atomicWriteFile(path, content, operations = { closeSync, fsyncSync, openSync, renameSync, unlinkSync, writeFileSync }, temporaryPath = `${path}.tmp-${process.pid}-${randomUUID()}`) {
  let temporaryCreated = false;
  try {
    const descriptor = operations.openSync(temporaryPath, "wx");
    temporaryCreated = true;
    try {
      operations.writeFileSync(descriptor, content, "utf8");
      operations.fsyncSync(descriptor);
    } finally {
      operations.closeSync(descriptor);
    }
    operations.renameSync(temporaryPath, path);
    temporaryCreated = false;
  } finally {
    if (temporaryCreated) operations.unlinkSync(temporaryPath);
  }
}

function withExclusiveLease(path, content, operation, operations = { closeSync, fsyncSync, openSync, unlinkSync, writeFileSync }) {
  let acquired = false;
  try {
    const descriptor = operations.openSync(path, "wx");
    acquired = true;
    try {
      operations.writeFileSync(descriptor, content, "utf8");
      operations.fsyncSync(descriptor);
    } finally {
      operations.closeSync(descriptor);
    }
    return operation();
  } finally {
    if (acquired) operations.unlinkSync(path);
  }
}

function validateFreezeTransition(current, proposed, now = Date.now()) {
  assert.deepEqual(Object.keys(proposed).sort(), Object.keys(current).sort(), "freeze plan changes epoch fields");
  assert.equal(current.status, "open", "current epoch must be open");
  assert.equal(proposed.status, "frozen", "freeze plan must set status to frozen");
  assert.ok(now > Date.parse(current.announcement_cutoff), "epoch announcement cutoff has not elapsed");
  for (const key of Object.keys(current).filter((key) => !["status", "producers"].includes(key))) {
    assert.deepEqual(proposed[key], current[key], `freeze plan changes ${key}`);
  }
  assert.deepEqual(current.producers.map((producer) => producer.thread), threads, "current producer roster is invalid");
  assert.deepEqual(proposed.producers.map((producer) => producer.thread), threads, "freeze plan producer roster is invalid");
  const announcedTags = [];
  for (let index = 0; index < threads.length; index += 1) {
    const before = current.producers[index];
    const after = proposed.producers[index];
    assert.deepEqual(Object.keys(after).sort(), producerKeys, `${after.thread} freeze decision fields are invalid`);
    assert.deepEqual(before, { thread: threads[index], status: "unannounced", tag: null, decision: null }, `${before.thread} current state is not open`);
    const announced = after.status === "announced" && typeof after.tag === "string" && new RegExp(`^checkpoint/prototype-${after.thread.toLowerCase()}/${after.thread.toLowerCase()}-[0-9]{2,}[a-z]?$`).test(after.tag) && after.decision === null;
    const frozenOut = after.status === "unannounced" && after.tag === null && after.decision === "frozen-out";
    assert.ok(announced || frozenOut, `${after.thread} freeze decision is invalid`);
    if (announced) announcedTags.push(after.tag);
  }
  assert.equal(new Set(announcedTags).size, announcedTags.length, "announced tags must be unique");
  return true;
}

function parseArgs(argv) {
  const options = { epoch: null, expectedHead: null, expectedPlanSha256: null, manifest: "_docs/ralph/prototype-integration/core-rc-manifest.json", observation: null, plan: null, remote: "origin", repo: resolve(__dirname, "..") };
  for (let index = 0; index < argv.length; index += 2) {
    const argument = argv[index];
    const value = argv[index + 1];
    assert.ok(["--epoch", "--expected-head", "--expected-plan-sha256", "--manifest", "--observation", "--plan", "--remote", "--repo"].includes(argument) && value, `invalid argument: ${argument || "missing"}`);
    const key = argument === "--expected-head" ? "expectedHead" : argument === "--expected-plan-sha256" ? "expectedPlanSha256" : argument.slice(2);
    options[key] = value;
  }
  assert.match(options.epoch || "", /^[1-9][0-9]*$/, "--epoch is required");
  assert.match(options.expectedHead || "", /^[a-f0-9]{40}$/, "--expected-head requires an exact commit SHA");
  assert.match(options.expectedPlanSha256 || "", /^[a-f0-9]{64}$/, "--expected-plan-sha256 requires an exact SHA-256 digest");
  assert.ok(options.observation && !options.observation.startsWith("/") && !options.observation.split(/[\\/]/).includes(".."), "--observation requires a repository-relative path");
  assert.ok(options.plan, "--plan is required");
  options.repo = resolve(options.repo);
  return options;
}

function validateFreezeEvidence(current, proposed, observationContent, observationPath, manifest, options = {}) {
  validateObservationBinding(observationContent, observationPath, manifest, options);
  const observation = JSON.parse(observationContent);
  const selections = new Map(proposed.producers.filter((producer) => producer.status === "announced").map((producer) => [producer.thread, producer.tag]));
  const recomputed = planFrozenEpoch(current, selections, { now: options.now, observation, resolveTag: options.resolveTag });
  assert.deepEqual(proposed, recomputed, "freeze plan does not match revalidated observation evidence");
  return true;
}

function validatePlanDigest(content, expectedDigest) {
  assert.equal(createHash("sha256").update(content).digest("hex"), expectedDigest, "freeze plan does not match reviewed SHA-256");
  return true;
}

function validateWorktreeBoundary(repo, expectedHead, run = spawnSync) {
  const status = run("git", ["status", "--porcelain"], { cwd: repo, encoding: "utf8" });
  assert.equal(status.status, 0, "unable to inspect T4 worktree");
  assert.equal(status.stdout.trim(), "", "T4 worktree must be clean before freezing intake");
  const head = run("git", ["rev-parse", "HEAD"], { cwd: repo, encoding: "utf8" });
  assert.equal(head.status, 0, "unable to resolve T4 HEAD");
  assert.equal(head.stdout.trim(), expectedHead, "T4 HEAD does not match reviewed freeze SHA");
  return true;
}

function validateRemoteBoundary(repo, expectedHead, remote = "origin", run = spawnSync) {
  const result = run("git", ["ls-remote", "--heads", remote, "refs/heads/ve/prototype-integration"], { cwd: repo, encoding: "utf8" });
  assert.equal(result.status, 0, "unable to resolve remote T4 integration head");
  const lines = result.stdout.trim().split(/\r?\n/).filter(Boolean);
  assert.equal(lines.length, 1, "remote T4 integration head is unavailable or ambiguous");
  assert.equal(lines[0].split(/\s+/)[0], expectedHead, "remote T4 integration head does not match reviewed freeze SHA");
  return true;
}

function withStablePublishedBoundary(repo, expectedHead, remote, operation, run = spawnSync) {
  validateWorktreeBoundary(repo, expectedHead, run);
  validateRemoteBoundary(repo, expectedHead, remote, run);
  const result = operation();
  validateWorktreeBoundary(repo, expectedHead, run);
  validateRemoteBoundary(repo, expectedHead, remote, run);
  return result;
}

function validateAppliedFreeze(path, expectedContent, repo, expectedHead, remote = "origin", options = {}) {
  const read = options.readFileSync ?? readFileSync;
  assert.equal(read(path, "utf8"), expectedContent, "applied epoch bytes do not match reviewed freeze plan");
  validateRemoteBoundary(repo, expectedHead, remote, options.run ?? spawnSync);
  return true;
}

function main(argv) {
  const options = parseArgs(argv);
  const lock = spawnSync("git", ["rev-parse", "--git-path", `prototype-intake-freeze-epoch-${options.epoch}.lock`], { cwd: options.repo, encoding: "utf8" });
  assert.equal(lock.status, 0, "unable to resolve intake freeze lock path");
  const lockPath = resolve(options.repo, lock.stdout.trim());
  const leaseRecord = createLeaseRecord(options);
  withExclusiveLease(lockPath, `${JSON.stringify(leaseRecord, null, 2)}\n`, () => {
    const prepared = withStablePublishedBoundary(options.repo, options.expectedHead, options.remote, () => {
      const epochPath = resolve(options.repo, `_docs/ralph/prototype-integration/epochs/epoch-${options.epoch}.json`);
      const current = JSON.parse(readFileSync(epochPath, "utf8"));
      const planContent = readFileSync(resolve(options.plan), "utf8");
      validatePlanDigest(planContent, options.expectedPlanSha256);
      const proposed = JSON.parse(planContent);
      assert.equal(current.intake_epoch, Number(options.epoch), "epoch number mismatch");
      validateFreezeTransition(current, proposed);
      const observationContent = readFileSync(resolve(options.repo, options.observation), "utf8");
      const manifest = JSON.parse(readFileSync(resolve(options.repo, options.manifest), "utf8"));
      validateFreezeEvidence(current, proposed, observationContent, options.observation, manifest, { repo: options.repo, resolveTag: (tag) => resolveAnnotatedTag(options.repo, options.remote, tag) });
      return { epochPath, serialized: `${JSON.stringify(proposed, null, 2)}\n` };
    });
    atomicWriteFile(prepared.epochPath, prepared.serialized);
    validateAppliedFreeze(prepared.epochPath, prepared.serialized, options.repo, options.expectedHead, options.remote);
  });
  process.stdout.write(`prototype intake epoch ${options.epoch} frozen\n`);
}

module.exports = { atomicWriteFile, createLeaseRecord, parseArgs, validateAppliedFreeze, validateFreezeEvidence, validateFreezeTransition, validatePlanDigest, validateRemoteBoundary, validateWorktreeBoundary, withExclusiveLease, withStablePublishedBoundary };

if (require.main === module) {
  try { main(process.argv.slice(2)); }
  catch (error) { console.error(`prototype intake freeze apply: invalid: ${error.message}`); process.exitCode = 1; }
}