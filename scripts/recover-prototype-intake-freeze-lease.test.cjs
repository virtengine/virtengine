"use strict";

const assert = require("assert").strict;
const { createHash } = require("crypto");
const { parseArgs, recoverLease } = require("./recover-prototype-intake-freeze-lease.cjs");

const head = "a".repeat(40);
const planDigest = "b".repeat(64);
const expected = { epoch: 1, expectedHead: head, expectedPlanSha256: planDigest };
const content = `${JSON.stringify({ schema_version: "virtengine.prototype.intake-freeze-lease/v1", pid: 42, started_at: "2000-01-01T12:00:00.000Z", epoch: 1, expected_head: head, plan_sha256: planDigest }, null, 2)}\n`;
const leaseDigest = createHash("sha256").update(content).digest("hex");
const inspection = { now: Date.parse("2000-01-02T00:00:00Z"), isProcessAlive: () => false };

function operations(contents = [content, content]) {
  const calls = [];
  return { calls, value: { readFileSync: (path) => { calls.push(["read", path]); return contents.shift(); }, renameSync: (...args) => calls.push(["rename", ...args]), unlinkSync: (...args) => calls.push(["unlink", ...args]) } };
}

const tests = [
  ["claims, revalidates, and removes an exact stale lease", () => { const ops = operations(); const value = recoverLease("freeze.lock", leaseDigest, expected, { ...inspection, operations: ops.value, quarantinePath: "freeze.claimed" }); assert.equal(value.recovery_status, "stale_lease_removed"); assert.deepEqual(ops.calls.map(([name]) => name), ["read", "rename", "read", "unlink"]); }],
  ["rejects lease bytes not matching the reviewed digest", () => { const ops = operations(); assert.throws(() => recoverLease("freeze.lock", "c".repeat(64), expected, { ...inspection, operations: ops.value }), /do not match reviewed/); assert.deepEqual(ops.calls.map(([name]) => name), ["read"]); }],
  ["refuses recovery while the recorded PID is present", () => { const ops = operations(); assert.throws(() => recoverLease("freeze.lock", leaseDigest, expected, { ...inspection, isProcessAlive: () => true, operations: ops.value }), /PID is present/); assert.deepEqual(ops.calls.map(([name]) => name), ["read"]); }],
  ["leaves the original lease when the reviewed checkout boundary fails", () => { const ops = operations(); assert.throws(() => recoverLease("freeze.lock", leaseDigest, expected, { ...inspection, operations: ops.value, validateBoundary: () => { throw new Error("stale checkout"); } }), /stale checkout/); assert.deepEqual(ops.calls.map(([name]) => name), ["read"]); }],
  ["retains quarantine when claimed bytes change", () => { const changed = content.replace("42", "43"); const ops = operations([content, changed]); assert.throws(() => recoverLease("freeze.lock", leaseDigest, expected, { ...inspection, operations: ops.value, quarantinePath: "freeze.claimed" }), /retained at freeze.claimed/); assert.deepEqual(ops.calls.map(([name]) => name), ["read", "rename", "read"]); }],
  ["reports quarantine when the claimed lease cannot be read", () => { const ops = operations([content]); assert.throws(() => recoverLease("freeze.lock", leaseDigest, expected, { ...inspection, operations: ops.value, quarantinePath: "freeze.claimed" }), /retained at freeze.claimed/); assert.deepEqual(ops.calls.map(([name]) => name), ["read", "rename", "read"]); }],
  ["reports quarantine when validated lease removal fails", () => { const ops = operations(); ops.value.unlinkSync = () => { ops.calls.push(["unlink", "freeze.claimed"]); throw new Error("access denied"); }; assert.throws(() => recoverLease("freeze.lock", leaseDigest, expected, { ...inspection, operations: ops.value, quarantinePath: "freeze.claimed" }), /retained at freeze.claimed: access denied/); }],
  ["parses exact compare-and-claim recovery evidence", () => { const value = parseArgs(["--epoch", "1", "--expected-head", head, "--expected-plan-sha256", planDigest, "--expected-lease-sha256", leaseDigest]); assert.equal(value.expectedLeaseSha256, leaseDigest); }],
  ["rejects duplicate or incomplete recovery arguments", () => {
    const required = ["--expected-head", head, "--expected-plan-sha256", planDigest, "--expected-lease-sha256", leaseDigest];
    assert.throws(() => parseArgs(["--epoch", "1", "--epoch", "2", ...required]), /duplicate argument/);
    assert.throws(() => parseArgs(["--epoch", "--expected-head", head, ...required.slice(2)]), /requires a value/);
    assert.throws(() => parseArgs(["--epoch", "1", ...required, "--remote"]), /requires a value/);
  }],
];

for (const [name, run] of tests) { run(); console.log(`ok - ${name}`); }