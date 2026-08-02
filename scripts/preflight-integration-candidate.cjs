#!/usr/bin/env node

"use strict";

const assert = require("assert").strict;
const { spawnSync } = require("child_process");

function git(repo, args, allowFailure = false) {
  const result = spawnSync("git", args, { cwd: repo, encoding: "utf8" });
  if (!allowFailure && result.status !== 0) throw new Error((result.stderr || result.stdout || `git ${args.join(" ")} failed`).trim());
  return result;
}

function parseAcceptance(content) {
  const acceptance = JSON.parse(content);
  assert.deepEqual(Object.keys(acceptance).sort(), ["accepted_payloads", "base_sha", "candidate_sha", "schema_version", "status"]);
  assert.equal(acceptance.schema_version, "virtengine.prototype.integration-candidate-acceptance/v1");
  assert.equal(acceptance.status, "validated");
  assert.match(acceptance.base_sha, /^[a-f0-9]{40}$/);
  assert.match(acceptance.candidate_sha, /^[a-f0-9]{40}$/);
  assert.ok(Array.isArray(acceptance.accepted_payloads));
  for (const entry of acceptance.accepted_payloads) {
    assert.deepEqual(Object.keys(entry).sort(), ["payload_sha", "tag", "thread"]);
    assert.ok(["T1", "T2", "T3", "T5"].includes(entry.thread));
    assert.match(entry.payload_sha, /^[a-f0-9]{40}$/);
    assert.match(entry.tag, new RegExp(`^checkpoint/prototype-${entry.thread.toLowerCase()}/[a-z0-9-]+$`));
  }
  return acceptance;
}

function validateCandidatePlan(plan, options) {
  assert.match(plan.canonical_head, /^[a-f0-9]{40}$/);
  assert.match(plan.candidate_head, /^[a-f0-9]{40}$/);
  assert.equal(options.isAncestor(plan.canonical_head, plan.candidate_head), true, "candidate does not descend canonical T4");
  assert.equal(options.acceptanceCommitted, true, "candidate acceptance artifact is not committed");
  assert.ok(Array.isArray(plan.accepted_payloads), "accepted payloads must be an array");
  assert.equal(new Set(plan.accepted_payloads.map((entry) => entry.thread)).size, plan.accepted_payloads.length, "accepted payload threads must be unique");
  for (const entry of plan.accepted_payloads) {
    assert.ok(["T1", "T2", "T3", "T5"].includes(entry.thread), `unknown accepted thread ${entry.thread}`);
    assert.match(entry.payload_sha, /^[a-f0-9]{40}$/);
    assert.match(entry.tag, new RegExp(`^checkpoint/prototype-${entry.thread.toLowerCase()}/[a-z0-9-]+$`));
    assert.equal(options.verifyAcceptedPayload(entry), true, `${entry.thread} acceptance verification failed`);
    assert.equal(options.isAncestor(entry.payload_sha, plan.candidate_head), true, `${entry.thread} payload is not in candidate history`);
  }
  for (const contained of plan.contained_producer_commits) {
    const covered = plan.accepted_payloads.some((entry) => entry.thread === contained.thread && options.isAncestor(contained.commit, entry.payload_sha));
    assert.equal(covered, true, `candidate contains unaccepted ${contained.thread} commit ${contained.commit}`);
  }
  return true;
}

function buildCandidatePlan(repo, candidateRef, canonicalRef, acceptancePath, producerBranches) {
  const canonicalHead = git(repo, ["rev-parse", `${canonicalRef}^{commit}`]).stdout.trim();
  const candidateHead = git(repo, ["rev-parse", `${candidateRef}^{commit}`]).stdout.trim();
  const acceptanceObject = git(repo, ["rev-parse", `${candidateRef}:${acceptancePath}`], true);
  assert.equal(acceptanceObject.status, 0, "candidate acceptance artifact is not committed");
  const acceptance = parseAcceptance(git(repo, ["show", `${candidateRef}:${acceptancePath}`]).stdout);
  assert.equal(acceptance.base_sha, canonicalHead, "acceptance base does not match canonical T4");
  assert.equal(acceptance.candidate_sha, candidateHead, "acceptance candidate does not match candidate HEAD");
  const acceptedPayloads = acceptance.accepted_payloads.map((entry) => ({ thread: entry.thread, tag: entry.tag, payload_sha: entry.payload_sha }));
  const base = git(repo, ["merge-base", canonicalHead, candidateHead]).stdout.trim();
  const contained = [];
  for (const [thread, branch] of producerBranches) {
    const commits = git(repo, ["rev-list", "--reverse", `${base}..refs/remotes/origin/${branch}`], true);
    if (commits.status !== 0) throw new Error(`registered producer ref is unavailable: ${branch}`);
    for (const commit of commits.stdout.trim().split(/\r?\n/).filter(Boolean)) {
      if (git(repo, ["merge-base", "--is-ancestor", commit, candidateHead], true).status === 0) contained.push({ thread, commit });
    }
  }
  return { canonical_head: canonicalHead, candidate_head: candidateHead, accepted_payloads: acceptedPayloads, contained_producer_commits: contained };
}

function main(argv) {
  const args = Object.fromEntries(Array.from({ length: argv.length / 2 }, (_, index) => [argv[index * 2], argv[index * 2 + 1]]));
  const repo = args["--repo"];
  const candidate = args["--candidate"];
  const canonical = args["--canonical"] || "origin/ve/prototype-integration";
  const acceptancePath = args["--acceptance"] || "_docs/ralph/prototype-integration/live-integration-acceptance.json";
  assert.ok(repo && candidate, "--repo and --candidate are required");
  const producerBranches = [["T1", "ve/prototype-t1-identity"], ["T2", "ve/prototype-t2-product"], ["T3", "ve/prototype-t3-reliability"], ["T5", "ve/prototype-t5-platform"]];
  const plan = buildCandidatePlan(repo, candidate, canonical, acceptancePath, producerBranches);
  validateCandidatePlan(plan, {
    acceptanceCommitted: true,
    isAncestor: (ancestor, descendant) => git(repo, ["merge-base", "--is-ancestor", ancestor, descendant], true).status === 0,
    verifyAcceptedPayload: (entry) => {
      const tag = git(repo, ["rev-parse", `${entry.tag}^{}`], true);
      return tag.status === 0 && tag.stdout.trim() === entry.payload_sha;
    },
  });
  process.stdout.write(`${JSON.stringify(plan, null, 2)}\n`);
}

module.exports = { buildCandidatePlan, parseAcceptance, validateCandidatePlan };

if (require.main === module) {
  try { main(process.argv.slice(2)); }
  catch (error) { console.error(`integration candidate preflight: invalid: ${error.message}`); process.exitCode = 1; }
}