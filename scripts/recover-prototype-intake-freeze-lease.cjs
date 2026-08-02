#!/usr/bin/env node

"use strict";

const assert = require("assert").strict;
const { randomUUID } = require("crypto");
const { readFileSync, renameSync, unlinkSync } = require("fs");
const { resolve } = require("path");
const { spawnSync } = require("child_process");
const { inspectLeaseContent } = require("./inspect-prototype-intake-freeze-lease.cjs");

function recoverLease(lockPath, expectedLeaseSha256, expected, options = {}) {
  const operations = options.operations ?? { readFileSync, renameSync, unlinkSync };
  const inspectionOptions = { now: options.now, isProcessAlive: options.isProcessAlive };
  const content = operations.readFileSync(lockPath, "utf8");
  const inspected = inspectLeaseContent(content, expected, inspectionOptions);
  assert.equal(inspected.lease_sha256, expectedLeaseSha256, "freeze lease bytes do not match reviewed SHA-256");
  assert.equal(inspected.pid_present, false, "refusing recovery while the recorded PID is present");

  const quarantinePath = options.quarantinePath ?? `${lockPath}.recovery-${randomUUID()}`;
  operations.renameSync(lockPath, quarantinePath);
  try {
    const claimedContent = operations.readFileSync(quarantinePath, "utf8");
    const claimed = inspectLeaseContent(claimedContent, expected, inspectionOptions);
    assert.equal(claimed.lease_sha256, expectedLeaseSha256, "claimed freeze lease bytes changed after review");
    assert.equal(claimed.pid_present, false, "recorded PID became present during recovery");
    operations.unlinkSync(quarantinePath);
  } catch (error) {
    throw new Error(`claimed lease retained at ${quarantinePath}: ${error.message}`);
  }
  return { recovery_status: "stale_lease_removed", lease_sha256: expectedLeaseSha256 };
}

function parseArgs(argv) {
  const options = { epoch: null, expectedHead: null, expectedLeaseSha256: null, expectedPlanSha256: null, repo: resolve(__dirname, "..") };
  for (let index = 0; index < argv.length; index += 2) {
    const argument = argv[index];
    const value = argv[index + 1];
    assert.ok(["--epoch", "--expected-head", "--expected-lease-sha256", "--expected-plan-sha256", "--repo"].includes(argument) && value, `invalid argument: ${argument || "missing"}`);
    const key = argument.split("-").slice(2).map((part, partIndex) => partIndex === 0 ? part : `${part[0].toUpperCase()}${part.slice(1)}`).join("");
    options[key] = value;
  }
  assert.match(options.epoch || "", /^[1-9][0-9]*$/, "--epoch is required");
  assert.match(options.expectedHead || "", /^[a-f0-9]{40}$/, "--expected-head requires an exact commit SHA");
  assert.match(options.expectedPlanSha256 || "", /^[a-f0-9]{64}$/, "--expected-plan-sha256 requires an exact SHA-256 digest");
  assert.match(options.expectedLeaseSha256 || "", /^[a-f0-9]{64}$/, "--expected-lease-sha256 requires an exact SHA-256 digest");
  options.epoch = Number(options.epoch);
  options.repo = resolve(options.repo);
  return options;
}

function main(argv) {
  const options = parseArgs(argv);
  const lock = spawnSync("git", ["rev-parse", "--git-path", `prototype-intake-freeze-epoch-${options.epoch}.lock`], { cwd: options.repo, encoding: "utf8" });
  assert.equal(lock.status, 0, "unable to resolve intake freeze lock path");
  const result = recoverLease(resolve(options.repo, lock.stdout.trim()), options.expectedLeaseSha256, options);
  process.stdout.write(`${JSON.stringify(result, null, 2)}\n`);
}

module.exports = { parseArgs, recoverLease };

if (require.main === module) {
  try { main(process.argv.slice(2)); }
  catch (error) { console.error(`prototype intake freeze recovery: invalid: ${error.message}`); process.exitCode = 1; }
}