#!/usr/bin/env node

"use strict";

const assert = require("assert").strict;
const { createHash } = require("crypto");
const { readFileSync } = require("fs");
const { resolve } = require("path");
const { spawnSync } = require("child_process");

const leaseKeys = ["epoch", "expected_head", "pid", "plan_sha256", "schema_version", "started_at"];

function processIsAlive(pid) {
  try { process.kill(pid, 0); return true; }
  catch (error) { if (error.code === "ESRCH") return false; if (error.code === "EPERM") return true; throw error; }
}

function inspectLease(lease, expected, options = {}) {
  assert.deepEqual(Object.keys(lease).sort(), leaseKeys, "freeze lease has unknown or missing fields");
  assert.equal(lease.schema_version, "virtengine.prototype.intake-freeze-lease/v1");
  assert.ok(Number.isInteger(lease.pid) && lease.pid > 0, "freeze lease PID is invalid");
  assert.equal(lease.epoch, expected.epoch, "freeze lease epoch mismatch");
  assert.equal(lease.expected_head, expected.expectedHead, "freeze lease reviewed HEAD mismatch");
  assert.equal(lease.plan_sha256, expected.expectedPlanSha256, "freeze lease plan digest mismatch");
  const startedAt = Date.parse(lease.started_at);
  const now = options.now ?? Date.now();
  assert.ok(Number.isFinite(startedAt) && startedAt <= now, "freeze lease start time is invalid");
  const alive = (options.isProcessAlive ?? processIsAlive)(lease.pid);
  return {
    ...lease,
    pid_present: alive,
    recovery_status: alive ? "pid_present_operator_verification_required" : "pid_absent_operator_review_required",
  };
}

function inspectLeaseContent(content, expected, options = {}) {
  const inspected = inspectLease(JSON.parse(content), expected, options);
  return { ...inspected, lease_sha256: createHash("sha256").update(content).digest("hex") };
}

function parseArgs(argv) {
  const options = { epoch: null, expectedHead: null, expectedPlanSha256: null, repo: resolve(__dirname, "..") };
  for (let index = 0; index < argv.length; index += 2) {
    const argument = argv[index];
    const value = argv[index + 1];
    assert.ok(["--epoch", "--expected-head", "--expected-plan-sha256", "--repo"].includes(argument) && value, `invalid argument: ${argument || "missing"}`);
    const key = argument === "--expected-head" ? "expectedHead" : argument === "--expected-plan-sha256" ? "expectedPlanSha256" : argument.slice(2);
    options[key] = value;
  }
  assert.match(options.epoch || "", /^[1-9][0-9]*$/, "--epoch is required");
  assert.match(options.expectedHead || "", /^[a-f0-9]{40}$/, "--expected-head requires an exact commit SHA");
  assert.match(options.expectedPlanSha256 || "", /^[a-f0-9]{64}$/, "--expected-plan-sha256 requires an exact SHA-256 digest");
  options.epoch = Number(options.epoch);
  options.repo = resolve(options.repo);
  return options;
}

function main(argv) {
  const options = parseArgs(argv);
  const lock = spawnSync("git", ["rev-parse", "--git-path", `prototype-intake-freeze-epoch-${options.epoch}.lock`], { cwd: options.repo, encoding: "utf8" });
  assert.equal(lock.status, 0, "unable to resolve intake freeze lock path");
  const content = readFileSync(resolve(options.repo, lock.stdout.trim()), "utf8");
  process.stdout.write(`${JSON.stringify(inspectLeaseContent(content, options), null, 2)}\n`);
}

module.exports = { inspectLease, inspectLeaseContent, parseArgs, processIsAlive };

if (require.main === module) {
  try { main(process.argv.slice(2)); }
  catch (error) { console.error(`prototype intake freeze lease: invalid: ${error.message}`); process.exitCode = 1; }
}