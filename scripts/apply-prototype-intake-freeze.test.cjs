"use strict";

const assert = require("assert").strict;
const { parseArgs, validateFreezeTransition } = require("./apply-prototype-intake-freeze.cjs");

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
const tests = [
  ["accepts the exact open-to-frozen transition", () => assert.equal(validateFreezeTransition(epoch(), frozen(), afterCutoff), true)],
  ["rejects application before cutoff", () => assert.throws(() => validateFreezeTransition(epoch(), frozen(), Date.parse("2000-01-01T00:00:00Z")), /cutoff has not elapsed/)],
  ["rejects changed epoch metadata", () => { const value = frozen(); value.base_sha = "c".repeat(40); assert.throws(() => validateFreezeTransition(epoch(), value, afterCutoff), /changes base_sha/); }],
  ["rejects an accepted decision during freeze", () => { const value = frozen(); value.producers[0].decision = "accepted"; assert.throws(() => validateFreezeTransition(epoch(), value, afterCutoff), /decision is invalid/); }],
  ["rejects a changed producer roster", () => { const value = frozen(); value.producers[0].thread = "T2"; assert.throws(() => validateFreezeTransition(epoch(), value, afterCutoff), /roster is invalid/); }],
  ["parses an explicit epoch and plan", () => { const value = parseArgs(["--epoch", "1", "--plan", "plan.json"]); assert.equal(value.epoch, "1"); assert.equal(value.plan, "plan.json"); }],
];

for (const [name, run] of tests) { run(); console.log(`ok - ${name}`); }