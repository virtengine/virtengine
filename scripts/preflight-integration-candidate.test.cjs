"use strict";

const assert = require("assert").strict;
const { parseAcceptance, validateCandidatePlan, verifyAcceptedPayload } = require("./preflight-integration-candidate.cjs");

const sha = (character) => character.repeat(40);
function fixture() {
  return {
    canonical_head: sha("a"), candidate_head: sha("b"),
    accepted_payloads: [{ thread: "T1", tag: "checkpoint/prototype-t1/t1-09", payload_sha: sha("c") }],
    contained_producer_commits: [{ thread: "T1", commit: sha("d") }],
  };
}
function options() {
  return { acceptanceCommitted: true, isAncestor: (ancestor, descendant) => [[sha("a"), sha("b")], [sha("c"), sha("b")], [sha("d"), sha("c")]].some(([left, right]) => left === ancestor && right === descendant), verifyAcceptedPayload: () => true };
}
const tests = [
  ["accepts an annotated tag resolving to the payload", () => { const entry = { tag: "checkpoint/prototype-t1/t1-09", payload_sha: sha("c") }; const runGit = (_repo, args) => args[0] === "cat-file" ? { status: 0, stdout: "tag\n" } : { status: 0, stdout: `${entry.payload_sha}\n` }; assert.equal(verifyAcceptedPayload(".", entry, runGit), true); }],
  ["rejects a lightweight tag resolving to the payload", () => { const entry = { tag: "checkpoint/prototype-t1/t1-09", payload_sha: sha("c") }; const runGit = (_repo, args) => args[0] === "cat-file" ? { status: 0, stdout: "commit\n" } : { status: 0, stdout: `${entry.payload_sha}\n` }; assert.equal(verifyAcceptedPayload(".", entry, runGit), false); }],
  ["accepts the strict acceptance artifact schema", () => { const value = { schema_version: "virtengine.prototype.integration-candidate-acceptance/v1", status: "validated", base_sha: sha("a"), candidate_sha: sha("b"), accepted_payloads: [{ thread: "T1", tag: "checkpoint/prototype-t1/t1-09", payload_sha: sha("c") }] }; assert.doesNotThrow(() => parseAcceptance(JSON.stringify(value))); }],
  ["rejects informal acceptance summaries", () => { const value = { schema_version: "virtengine.prototype.live-integration-acceptance/v1", accepted_payloads: [{ thread: "T1", sha: sha("c") }] }; assert.throws(() => parseAcceptance(JSON.stringify(value))); }],
  ["accepts a canonical descendant with covered producer history", () => assert.doesNotThrow(() => validateCandidatePlan(fixture(), options()))],
  ["rejects a divergent candidate", () => { const value = options(); value.isAncestor = () => false; assert.throws(() => validateCandidatePlan(fixture(), value), /does not descend canonical/); }],
  ["rejects an uncommitted acceptance artifact", () => { const value = options(); value.acceptanceCommitted = false; assert.throws(() => validateCandidatePlan(fixture(), value), /not committed/); }],
  ["rejects duplicate accepted threads", () => { const value = fixture(); value.accepted_payloads.push(structuredClone(value.accepted_payloads[0])); assert.throws(() => validateCandidatePlan(value, options()), /threads must be unique/); }],
  ["rejects an unverified payload", () => { const value = options(); value.verifyAcceptedPayload = () => false; assert.throws(() => validateCandidatePlan(fixture(), value), /acceptance verification failed/); }],
  ["rejects uncovered producer history", () => { const value = fixture(); value.contained_producer_commits[0].thread = "T5"; assert.throws(() => validateCandidatePlan(value, options()), /unaccepted T5 commit/); }],
];
for (const [name, run] of tests) { run(); console.log(`ok - ${name}`); }