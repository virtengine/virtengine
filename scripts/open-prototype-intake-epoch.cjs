#!/usr/bin/env node

"use strict";

const assert = require("assert").strict;
const { spawnSync } = require("child_process");
const { resolve } = require("path");
const { currentEpoch, discoverEpochs } = require("./prototype-intake-epochs.cjs");

const threads = ["T1", "T2", "T3", "T5"];

function git(repo, args) {
  const result = spawnSync("git", args, { cwd: repo, encoding: "utf8" });
  if (result.status !== 0) throw new Error((result.stderr || result.stdout || `git ${args.join(" ")} failed`).trim());
  return result.stdout.trim();
}

function buildOpenEpoch(previous, options) {
  const epoch = Number(options.epoch);
  assert.equal(previous.document.status, "closed", `predecessor epoch ${previous.number} must be closed`);
  assert.equal(epoch, previous.number + 1, "new intake epoch must immediately follow the closed predecessor");
  assert.match(options.expectedHead || "", /^[a-f0-9]{40}$/, "expected HEAD must be an exact commit SHA");
  const opensAt = Date.parse(options.opensAt);
  const cutoff = Date.parse(options.announcementCutoff);
  assert.ok(Number.isFinite(opensAt) && options.opensAt.endsWith("Z"), "opens-at must be a UTC date-time");
  assert.ok(Number.isFinite(cutoff) && options.announcementCutoff.endsWith("Z"), "announcement-cutoff must be a UTC date-time");
  assert.ok(opensAt < cutoff, "announcement cutoff must follow opens-at");
  assert.ok(opensAt >= Date.parse(previous.document.announcement_cutoff), "new epoch cannot open before the predecessor cutoff");
  return {
    schema_version: "virtengine.prototype.intake-epoch/v2",
    campaign: "three-day-prototype",
    intake_epoch: epoch,
    base_tag: `checkpoint/prototype-integration/epoch-${epoch}-base`,
    base_sha: options.expectedHead,
    planning_sha: previous.document.planning_sha,
    status: "open",
    opens_at: options.opensAt,
    announcement_cutoff: options.announcementCutoff,
    producers: threads.map((thread) => ({ thread, status: "unannounced", tag: null, decision: null })),
  };
}

function parseArgs(argv) {
  const options = { repo: resolve(__dirname, ".."), remote: "origin" };
  const seen = new Set();
  for (let index = 0; index < argv.length; index += 2) {
    const argument = argv[index];
    const value = argv[index + 1];
    assert.ok(["--epoch", "--expected-head", "--opens-at", "--announcement-cutoff", "--repo", "--remote"].includes(argument), `invalid argument: ${argument || "missing"}`);
    assert.equal(seen.has(argument), false, `duplicate argument: ${argument}`);
    seen.add(argument);
    assert.ok(value && !value.startsWith("--"), `${argument} requires a value`);
    const key = argument.slice(2).replace(/-([a-z])/g, (_, letter) => letter.toUpperCase());
    options[key] = value;
  }
  assert.ok(options.epoch && options.expectedHead && options.opensAt && options.announcementCutoff, "epoch, expected-head, opens-at, and announcement-cutoff are required");
  options.repo = resolve(options.repo);
  return options;
}

function validatePublishedBoundary(options, runGit = git) {
  assert.equal(runGit(options.repo, ["status", "--porcelain"]), "", "T4 worktree must be clean before opening an epoch");
  assert.equal(runGit(options.repo, ["rev-parse", "HEAD"]), options.expectedHead, "local HEAD does not match reviewed epoch base");
  assert.equal(runGit(options.repo, ["ls-remote", "--heads", options.remote, "ve/prototype-integration"]).split("\t")[0], options.expectedHead, "remote integration branch does not match reviewed epoch base");
  const tag = `refs/tags/checkpoint/prototype-integration/epoch-${options.epoch}-base`;
  runGit(options.repo, ["fetch", options.remote, `${tag}:${tag}`]);
  assert.equal(runGit(options.repo, ["cat-file", "-t", tag]), "tag", "epoch base tag must be annotated");
  assert.equal(runGit(options.repo, ["rev-parse", `${tag}^{}`]), options.expectedHead, "epoch base tag does not target reviewed HEAD");
}

function main(argv = process.argv.slice(2)) {
  const options = parseArgs(argv);
  const previous = currentEpoch(discoverEpochs(resolve(options.repo, "_docs/ralph/prototype-integration/epochs")));
  validatePublishedBoundary(options);
  process.stdout.write(`${JSON.stringify(buildOpenEpoch(previous, options), null, 2)}\n`);
}

module.exports = { buildOpenEpoch, parseArgs, validatePublishedBoundary };

if (require.main === module) {
  try {
    main();
  } catch (error) {
    console.error(`prototype intake epoch open: invalid: ${error.message}`);
    process.exitCode = 1;
  }
}