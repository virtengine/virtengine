"use strict";

const assert = require("assert").strict;
const { createHash } = require("crypto");
const { readFileSync } = require("fs");
const { resolve } = require("path");
const { parseArgs, validateFreezeEvidence, validateFreezeTransition, validatePlanDigest, validateWorktreeBoundary } = require("./apply-prototype-intake-freeze.cjs");

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

function evidence() {
  const observation = { schema_version: "virtengine.prototype.intake-tag-observation/v1", intake_epoch: 1, announcement_cutoff: "2000-01-02T00:00:00Z", observed_at: "2000-01-01T18:00:00Z", tags: [{ thread: "T1", tag: "checkpoint/prototype-t1/t1-09", target: "c".repeat(40) }] };
  const content = JSON.stringify(observation);
  return { content, manifest: { source: { payload_sha: "d".repeat(40) }, control_artifacts: [{ id: "intake_tag_observation", path: "observation.json", sha256: createHash("sha256").update(content).digest("hex") }] } };
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
  ["parses explicit reviewed HEAD, plan digest, and observation", () => { const value = parseArgs(["--epoch", "1", "--expected-head", "a".repeat(40), "--expected-plan-sha256", "b".repeat(64), "--observation", "observation.json", "--plan", "plan.json"]); assert.equal(value.observation, "observation.json"); }],
  ["rejects a missing expected HEAD", () => assert.throws(() => parseArgs(["--epoch", "1", "--expected-plan-sha256", "b".repeat(64), "--observation", "observation.json", "--plan", "plan.json"]), /expected-head/)],
  ["rejects an abbreviated expected HEAD", () => assert.throws(() => parseArgs(["--epoch", "1", "--expected-head", "abc123", "--expected-plan-sha256", "b".repeat(64), "--observation", "observation.json", "--plan", "plan.json"]), /exact commit SHA/)],
  ["rejects a missing reviewed plan digest", () => assert.throws(() => parseArgs(["--epoch", "1", "--expected-head", "a".repeat(40), "--observation", "observation.json", "--plan", "plan.json"]), /expected-plan-sha256/)],
  ["accepts reviewed plan bytes", () => { const content = "reviewed plan\n"; assert.equal(validatePlanDigest(content, createHash("sha256").update(content).digest("hex")), true); }],
  ["rejects substituted plan bytes", () => { const digest = createHash("sha256").update("reviewed").digest("hex"); assert.throws(() => validatePlanDigest("substituted", digest), /does not match reviewed/); }],
  ["accepts a plan backed by observed tag evidence", () => { const value = evidence(); assert.equal(validateFreezeEvidence(epoch(), frozen(), value.content, "observation.json", value.manifest, { now: afterCutoff, sourceContent: value.content, sourceIsAncestor: true, resolveTag: () => ({ target: "c".repeat(40), tagger_at: "2000-01-01T12:00:00Z" }) }), true); }],
  ["rejects an announced tag absent from observation evidence", () => { const value = evidence(); const plan = frozen(); plan.producers[0].tag = "checkpoint/prototype-t1/t1-10"; assert.throws(() => validateFreezeEvidence(epoch(), plan, value.content, "observation.json", value.manifest, { now: afterCutoff, sourceContent: value.content, sourceIsAncestor: true, resolveTag: () => ({ target: "e".repeat(40), tagger_at: "2000-01-01T12:00:00Z" }) }), /not uniquely observed/); }],
  ["accepts a clean worktree at the reviewed HEAD", () => assert.equal(validateWorktreeBoundary(".", "a".repeat(40), cleanAt("a".repeat(40))), true)],
  ["rejects a clean worktree at a stale HEAD", () => assert.throws(() => validateWorktreeBoundary(".", "a".repeat(40), cleanAt("b".repeat(40))), /does not match reviewed/)],
  ["runbook requires separately reviewed HEAD and plan digest", () => { const runbook = readFileSync(resolve(__dirname, "../_docs/prototype-thread-intake-runbook.md"), "utf8"); assert.match(runbook, /\$reviewedT4 = '<full T4 SHA recorded during separate plan review>'/); assert.match(runbook, /\$reviewedPlanSha256 = '<SHA-256 recorded during separate plan review>'/); assert.doesNotMatch(runbook, /--expected-head \(git rev-parse HEAD\)|\$reviewedPlanSha256\s*=\s*\(Get-FileHash/); }],
];

for (const [name, run] of tests) { run(); console.log(`ok - ${name}`); }