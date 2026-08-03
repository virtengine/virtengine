#!/usr/bin/env node

"use strict";

const assert = require("assert").strict;
const { resolve } = require("path");
const { spawnSync } = require("child_process");
const { discoverEpochs, requireCurrentEpoch } = require("./prototype-intake-epochs.cjs");

const tagPattern = /^refs\/tags\/(checkpoint\/prototype-t([1235])\/(t[1235]-[0-9]{2,}[a-z]?))$/;

function parseRemoteTagListing(listing) {
  const direct = new Map();
  const peeled = new Map();
  for (const line of listing.split(/\r?\n/).filter(Boolean)) {
    const [sha, ref] = line.trim().split(/\s+/, 2);
    assert.match(sha, /^[a-f0-9]{40}$/, "remote tag listing contains invalid SHA");
    if (ref.endsWith("^{}")) peeled.set(ref.slice(0, -3), sha);
    else direct.set(ref, sha);
  }
  const tags = [];
  for (const [ref, tagObject] of direct) {
    const match = ref.match(tagPattern);
    if (!match || !peeled.has(ref)) continue;
    tags.push({ thread: `T${match[2]}`, tag: match[1], tag_object: tagObject, target: peeled.get(ref) });
  }
  return tags.sort((left, right) => left.thread.localeCompare(right.thread) || left.tag.localeCompare(right.tag));
}

function createObservation(epoch, listing, options = {}) {
  const now = options.now ?? Date.now();
  const opensAt = Date.parse(epoch.opens_at);
  const cutoff = Date.parse(epoch.announcement_cutoff);
  assert.ok(Number.isFinite(opensAt), "epoch opens_at is invalid");
  assert.ok(Number.isFinite(cutoff), "epoch cutoff is invalid");
  assert.ok(now >= opensAt, "tag observation must be captured after epoch opens_at");
  assert.ok(now <= cutoff, "tag observation must be captured before cutoff");
  const observation = {
    schema_version: "virtengine.prototype.intake-tag-observation/v1",
    intake_epoch: epoch.intake_epoch,
    announcement_cutoff: epoch.announcement_cutoff,
    observed_at: new Date(now).toISOString(),
    remote: options.remote || "origin",
    tags: parseRemoteTagListing(listing),
  };
  validateTagObservation(observation, epoch);
  return observation;
}

function validateTagObservation(observation, epoch) {
  assert.deepEqual(Object.keys(observation).sort(), ["announcement_cutoff", "intake_epoch", "observed_at", "remote", "schema_version", "tags"]);
  assert.equal(observation.schema_version, "virtengine.prototype.intake-tag-observation/v1");
  assert.equal(observation.intake_epoch, epoch.intake_epoch);
  assert.equal(observation.announcement_cutoff, epoch.announcement_cutoff);
  assert.equal(observation.remote, "origin");
  const observedAt = Date.parse(observation.observed_at);
  assert.ok(observedAt >= Date.parse(epoch.opens_at), "tag observation was captured before epoch opens_at");
  assert.ok(observedAt <= Date.parse(epoch.announcement_cutoff), "tag observation was not captured before cutoff");
  const keys = observation.tags.map((entry) => `${entry.thread}:${entry.tag}`);
  assert.deepEqual(keys, [...keys].sort(), "observed tags must be sorted");
  assert.equal(new Set(keys).size, keys.length, "observed tags must be unique");
  for (const entry of observation.tags) {
    assert.deepEqual(Object.keys(entry).sort(), ["tag", "tag_object", "target", "thread"]);
    const match = `refs/tags/${entry.tag}`.match(tagPattern);
    assert.ok(match && entry.thread === `T${match[2]}`, "observed tag has invalid producer ownership");
    assert.match(entry.tag_object, /^[a-f0-9]{40}$/);
    assert.match(entry.target, /^[a-f0-9]{40}$/);
  }
  return true;
}

function main(argv) {
  const options = { epoch: null, remote: "origin", repo: resolve(__dirname, "..") };
  for (let index = 0; index < argv.length; index += 1) {
    const argument = argv[index];
    assert.ok(["--epoch", "--remote", "--repo"].includes(argument), `unknown argument: ${argument}`);
    const value = argv[++index];
    assert.ok(value, `${argument} requires a value`);
    options[argument.slice(2)] = value;
  }
  assert.match(options.epoch || "", /^[1-9][0-9]*$/, "--epoch is required");
  options.repo = resolve(options.repo);
  const epochDirectory = resolve(options.repo, "_docs/ralph/prototype-integration/epochs");
  const epoch = requireCurrentEpoch(discoverEpochs(epochDirectory), options.epoch);
  assert.equal(epoch.status, "open", "tag observation requires the current epoch to be open");
  const patterns = [1, 2, 3, 5].flatMap((thread) => [`refs/tags/checkpoint/prototype-t${thread}/*`, `refs/tags/checkpoint/prototype-t${thread}/*^{}`]);
  const result = spawnSync("git", ["ls-remote", "--tags", options.remote, ...patterns], { cwd: options.repo, encoding: "utf8" });
  assert.equal(result.status, 0, (result.stderr || "remote tag observation failed").trim());
  process.stdout.write(`${JSON.stringify(createObservation(epoch, result.stdout, options), null, 2)}\n`);
}

module.exports = { createObservation, parseRemoteTagListing, validateTagObservation };

if (require.main === module) {
  try { main(process.argv.slice(2)); }
  catch (error) { console.error(`prototype intake tag observation: invalid: ${error.message}`); process.exitCode = 1; }
}