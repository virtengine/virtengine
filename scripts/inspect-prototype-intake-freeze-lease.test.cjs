"use strict";

const assert = require("assert").strict;
const { createHash } = require("crypto");
const { inspectLease, inspectLeaseContent, parseArgs } = require("./inspect-prototype-intake-freeze-lease.cjs");

const head = "a".repeat(40);
const digest = "b".repeat(64);
const expected = { epoch: 1, expectedHead: head, expectedPlanSha256: digest };
function lease() { return { schema_version: "virtengine.prototype.intake-freeze-lease/v1", pid: 42, started_at: "2000-01-01T12:00:00.000Z", epoch: 1, expected_head: head, plan_sha256: digest }; }

const tests = [
  ["requires ownership verification when the PID is present", () => { const value = inspectLease(lease(), expected, { now: Date.parse("2000-01-02T00:00:00Z"), isProcessAlive: () => true }); assert.equal(value.recovery_status, "pid_present_operator_verification_required"); assert.equal(value.pid_present, true); }],
  ["requires operator review when the PID is absent", () => { const value = inspectLease(lease(), expected, { now: Date.parse("2000-01-02T00:00:00Z"), isProcessAlive: () => false }); assert.equal(value.recovery_status, "pid_absent_operator_review_required"); }],
  ["never treats PID presence as proof of lease ownership", () => { const value = inspectLease(lease(), expected, { now: Date.parse("2000-01-02T00:00:00Z"), isProcessAlive: () => true }); assert.notEqual(value.recovery_status, "active_process"); }],
  ["reports the exact retained lease byte digest", () => { const content = `${JSON.stringify(lease(), null, 2)}\n`; const value = inspectLeaseContent(content, expected, { now: Date.parse("2000-01-02T00:00:00Z"), isProcessAlive: () => false }); assert.equal(value.lease_sha256, createHash("sha256").update(content).digest("hex")); }],
  ["changes the digest when lease bytes change", () => { const first = `${JSON.stringify(lease())}\n`; const second = `${JSON.stringify(lease(), null, 2)}\n`; const options = { now: Date.parse("2000-01-02T00:00:00Z"), isProcessAlive: () => false }; assert.notEqual(inspectLeaseContent(first, expected, options).lease_sha256, inspectLeaseContent(second, expected, options).lease_sha256); }],
  ["rejects unknown lease fields", () => { const value = lease(); value.auto_remove = true; assert.throws(() => inspectLease(value, expected), /unknown or missing/); }],
  ["rejects a mismatched reviewed HEAD", () => assert.throws(() => inspectLease(lease(), { ...expected, expectedHead: "c".repeat(40) }), /HEAD mismatch/)],
  ["rejects a future lease timestamp", () => assert.throws(() => inspectLease(lease(), expected, { now: Date.parse("1999-01-01T00:00:00Z") }), /start time is invalid/)],
  ["parses exact expected recovery evidence", () => { const value = parseArgs(["--epoch", "1", "--expected-head", head, "--expected-plan-sha256", digest]); assert.equal(value.epoch, 1); }],
];

for (const [name, run] of tests) { run(); console.log(`ok - ${name}`); }