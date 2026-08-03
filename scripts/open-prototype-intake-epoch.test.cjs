"use strict";

const assert = require("assert").strict;
const { buildOpenEpoch, parseArgs, validatePublishedBoundary } = require("./open-prototype-intake-epoch.cjs");

const head = "a".repeat(40);
const previous = {
  file: "epoch-1.json",
  number: 1,
  document: { status: "closed", planning_sha: "b".repeat(40), announcement_cutoff: "2026-08-02T23:59:59Z" },
};
const options = { epoch: "2", expectedHead: head, opensAt: "2026-08-03T00:30:00Z", announcementCutoff: "2026-08-03T23:59:59Z", repo: "repo", remote: "origin" };

const tests = [
  ["builds an open epoch after a closed predecessor", () => {
    const epoch = buildOpenEpoch(previous, options);
    assert.equal(epoch.intake_epoch, 2);
    assert.equal(epoch.base_tag, "checkpoint/prototype-integration/epoch-2-base");
    assert.equal(epoch.base_sha, head);
    assert.ok(epoch.producers.every((producer) => producer.status === "unannounced" && producer.tag === null && producer.decision === null));
  }],
  ["rejects a non-closed predecessor", () => assert.throws(() => buildOpenEpoch({ ...previous, document: { ...previous.document, status: "frozen" } }, options), /must be closed/)],
  ["rejects a non-contiguous epoch", () => assert.throws(() => buildOpenEpoch(previous, { ...options, epoch: "3" }), /immediately follow/)],
  ["rejects a non-exact reviewed SHA", () => assert.throws(() => buildOpenEpoch(previous, { ...options, expectedHead: "HEAD" }), /exact commit SHA/)],
  ["rejects an invalid time window", () => assert.throws(() => buildOpenEpoch(previous, { ...options, announcementCutoff: options.opensAt }), /must follow/)],
  ["parses explicit opening arguments", () => assert.equal(parseArgs(["--epoch", "2", "--expected-head", head, "--opens-at", options.opensAt, "--announcement-cutoff", options.announcementCutoff]).epoch, "2")],
  ["accepts a clean published boundary and annotated exact tag", () => {
    const outputs = new Map([
      ["status --porcelain", ""], ["rev-parse HEAD", head], ["ls-remote --heads origin ve/prototype-integration", `${head}\trefs/heads/ve/prototype-integration`],
      ["fetch origin refs/tags/checkpoint/prototype-integration/epoch-2-base:refs/tags/checkpoint/prototype-integration/epoch-2-base", ""],
      ["cat-file -t refs/tags/checkpoint/prototype-integration/epoch-2-base", "tag"], ["rev-parse refs/tags/checkpoint/prototype-integration/epoch-2-base^{}", head],
    ]);
    assert.doesNotThrow(() => validatePublishedBoundary(options, (_repo, args) => outputs.get(args.join(" "))));
  }],
  ["rejects a lightweight base tag", () => {
    assert.throws(() => validatePublishedBoundary(options, (_repo, args) => args[0] === "cat-file" ? "commit" : args[0] === "status" ? "" : args[0] === "ls-remote" ? `${head}\tref` : head), /must be annotated/);
  }],
];

for (const [name, run] of tests) {
  run();
  console.log(`ok - ${name}`);
}