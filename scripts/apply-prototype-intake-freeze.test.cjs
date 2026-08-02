"use strict";

const assert = require("assert").strict;
const { readFileSync } = require("fs");
const { resolve } = require("path");
const { parseArgs, validateFreezeTransition, validateWorktreeBoundary } = require("./apply-prototype-intake-freeze.cjs");

function epoch() {
  return {
    schema_version: "virtengine.prototype.intake-epoch/v2", campaign: "three-day-prototype", intake_epoch: 1,
    base_tag: "checkpoint/prototype-integration/epoch-1-base", base_sha: "a".repeat(40), planning_sha: "b".repeat(40),
    status: "open", opens_at: "2000-01-01T00:00:00Z", announcement_cutoff: "2000-01-02T00:00:00Z",
    producers: ["T1", "T2", "T3", "T5"].map((thread) => ({ thread, status: "unannounced", tag: null, decision: null })),
  };
}

function frozen() {
  const value = epoch();
  value.status = "frozen";
  value.producers[0] = { thread: "T1", status: "announced", tag: "checkpoint/prototype-t1/t1-09", decision: null };
  for (const producer of value.producers.slice(1)) producer.decision = "frozen-out";
  return value;
}

const afterCutoff = Date.parse("2000-01-03T00:00:00Z");
const cleanAt = (head) => (_command, args) => args[0] === "status" ? { status: 0, stdout: "" } : { status: 0, stdout: `${head}\n` };
const tests = [
  ["accepts the exact open-to-frozen transition", () => assert.equal(validateFreezeTransition(epoch(), frozen(), afterCutoff), true)],
  ["rejects application before cutoff", () => assert.throws(() => validateFreezeTransition(epoch(), frozen(), Date.parse("2000-01-01T00:00:00Z")), /cutoff has not elapsed/)],
  ["rejects changed epoch metadata", () => { const value = frozen(); value.base_sha = "c".repeat(40); assert.throws(() => validateFreezeTransition(epoch(), value, afterCutoff), /changes base_sha/); }],
  ["rejects an accepted decision during freeze", () => { const value = frozen(); value.producers[0].decision = "accepted"; assert.throws(() => validateFreezeTransition(epoch(), value, afterCutoff), /decision is invalid/); }],
  ["rejects a changed producer roster", () => { const value = frozen(); value.producers[0].thread = "T2"; assert.throws(() => validateFreezeTransition(epoch(), value, afterCutoff), /roster is invalid/); }],
  ["rejects a wrong-thread announced tag", () => { const value = frozen(); value.producers[0].tag = "checkpoint/prototype-t3/t3-13a"; assert.throws(() => validateFreezeTransition(epoch(), value, afterCutoff), /decision is invalid/); }],
  ["rejects unknown producer fields", () => { const value = frozen(); value.producers[0].accepted = true; assert.throws(() => validateFreezeTransition(epoch(), value, afterCutoff), /fields are invalid/); }],
  ["parses an explicit epoch, plan, and expected HEAD", () => { const value = parseArgs(["--epoch", "1", "--expected-head", "a".repeat(40), "--plan", "plan.json"]); assert.equal(value.expectedHead, "a".repeat(40)); assert.equal(value.plan, "plan.json"); }],
  ["rejects a missing expected HEAD", () => assert.throws(() => parseArgs(["--epoch", "1", "--plan", "plan.json"]), /expected-head/)],
  ["rejects an abbreviated expected HEAD", () => assert.throws(() => parseArgs(["--epoch", "1", "--expected-head", "abc123", "--plan", "plan.json"]), /exact commit SHA/)],
  ["accepts a clean worktree at the reviewed HEAD", () => assert.equal(validateWorktreeBoundary(".", "a".repeat(40), cleanAt("a".repeat(40))), true)],
  ["rejects a clean worktree at a stale HEAD", () => assert.throws(() => validateWorktreeBoundary(".", "a".repeat(40), cleanAt("b".repeat(40))), /does not match reviewed/)],
  ["runbook requires a separately reviewed HEAD", () => { const runbook = readFileSync(resolve(__dirname, "../_docs/prototype-thread-intake-runbook.md"), "utf8"); assert.match(runbook, /\$reviewedT4 = '<full T4 SHA recorded during separate plan review>'/); assert.doesNotMatch(runbook, /--expected-head \(git rev-parse HEAD\)/); }],
];

for (const [name, run] of tests) { run(); console.log(`ok - ${name}`); }