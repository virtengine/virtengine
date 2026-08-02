#!/usr/bin/env node

"use strict";

const assert = require("assert").strict;
const { createHash } = require("crypto");
const { readFileSync } = require("fs");
const { isAbsolute, resolve } = require("path");
const { spawnSync } = require("child_process");

const threads = ["T1", "T2", "T3", "T5"];
const tagPattern = /^checkpoint\/prototype-t([1235])\/(t[1235]-[0-9]{2,}[a-z]?)$/;

function git(repo, args) {
  const result = spawnSync("git", args, { cwd: repo, encoding: "utf8" });
  if (result.status !== 0) throw new Error((result.stderr || result.stdout || `git ${args.join(" ")} failed`).trim());
  return result.stdout.trim();
}

function resolveAnnotatedTag(repo, remote, tag) {
  const ref = `refs/tags/${tag}`;
  git(repo, ["fetch", remote, `${ref}:${ref}`]);
  assert.equal(git(repo, ["cat-file", "-t", ref]), "tag", `${tag} is not an annotated tag`);
  return {
    tagger_at: git(repo, ["for-each-ref", "--format=%(taggerdate:iso-strict)", ref]),
    target: git(repo, ["rev-parse", `${ref}^{}`]),
  };
}

function validateObservationBinding(content, observationPath, manifest, options = {}) {
  assert.ok(!observationPath.startsWith("/") && !observationPath.split(/[\\/]/).includes(".."), "observation path must be repository-relative");
  const digest = createHash("sha256").update(content).digest("hex");
  const artifacts = manifest.control_artifacts.filter((artifact) => artifact.id === "intake_tag_observation" && artifact.path === observationPath);
  assert.equal(artifacts.length, 1, "manifest does not bind the intake tag observation");
  assert.equal(artifacts[0].sha256, digest, "observation digest does not match manifest");
  assert.match(manifest.source.payload_sha, /^[a-f0-9]{40}$/);
  const sourceResult = options.sourceContent === undefined
    ? spawnSync("git", ["show", `${manifest.source.payload_sha}:${observationPath}`], { cwd: options.repo, encoding: "buffer" })
    : null;
  if (sourceResult) assert.equal(sourceResult.status, 0, "observation is missing from manifest source commit");
  const sourceContent = options.sourceContent ?? sourceResult.stdout;
  assert.equal(createHash("sha256").update(sourceContent).digest("hex"), digest, "observation bytes do not match manifest source commit");
  const ancestor = options.sourceIsAncestor ?? spawnSync("git", ["merge-base", "--is-ancestor", manifest.source.payload_sha, "HEAD"], { cwd: options.repo }).status === 0;
  assert.equal(ancestor, true, "manifest source is not an ancestor of current T4");
  return true;
}

function planFrozenEpoch(epoch, selections, options = {}) {
  const now = options.now ?? Date.now();
  const cutoff = Date.parse(epoch.announcement_cutoff);
  assert.ok(Number.isFinite(cutoff), "epoch cutoff is invalid");
  assert.ok(now > cutoff, "epoch announcement cutoff has not elapsed");
  assert.equal(epoch.status, "open", "only an open epoch can be frozen");
  assert.deepEqual(epoch.producers.map((producer) => producer.thread), threads, "epoch producer roster is invalid");
  const observation = options.observation;
  assert.ok(observation && observation.schema_version === "virtengine.prototype.intake-tag-observation/v1", "pre-cutoff tag observation is required");
  assert.equal(observation.intake_epoch, epoch.intake_epoch, "tag observation epoch mismatch");
  assert.equal(observation.announcement_cutoff, epoch.announcement_cutoff, "tag observation cutoff mismatch");
  const observedAt = Date.parse(observation.observed_at);
  assert.ok(Number.isFinite(observedAt) && observedAt <= cutoff, "tag observation was not captured before cutoff");
  assert.ok(Array.isArray(observation.tags), "tag observation tags are invalid");
  for (const thread of selections.keys()) assert.ok(threads.includes(thread), `unknown producer selection: ${thread}`);

  const producers = epoch.producers.map((producer) => {
    const tag = selections.get(producer.thread);
    if (!tag) return { thread: producer.thread, status: "unannounced", tag: null, decision: "frozen-out" };
    const match = tag.match(tagPattern);
    assert.ok(match && `T${match[1]}` === producer.thread, `${producer.thread} selected an invalid tag`);
    const resolved = options.resolveTag(tag);
    assert.match(resolved.target, /^[a-f0-9]{40}$/, `${tag} target is not a commit SHA`);
    const observed = observation.tags.filter((entry) => entry.thread === producer.thread && entry.tag === tag && entry.target === resolved.target);
    assert.equal(observed.length, 1, `${tag} was not uniquely observed before cutoff`);
    const taggerTime = Date.parse(resolved.tagger_at);
    assert.ok(Number.isFinite(taggerTime), `${tag} tagger timestamp is invalid`);
    assert.ok(taggerTime <= cutoff, `${tag} was published after the announcement cutoff`);
    return { thread: producer.thread, status: "announced", tag, decision: null };
  });
  return { ...epoch, status: "frozen", producers };
}

function parseArgs(argv) {
  const options = { epoch: null, manifest: "_docs/ralph/prototype-integration/core-rc-manifest.json", observation: null, remote: "origin", repo: resolve(__dirname, ".."), selections: new Map() };
  for (let index = 0; index < argv.length; index += 1) {
    const argument = argv[index];
    if (["--epoch", "--manifest", "--observation", "--remote", "--repo", "--tag"].includes(argument)) {
      const value = argv[++index];
      assert.ok(value, `${argument} requires a value`);
      if (argument === "--tag") {
        const match = value.match(/^(T[1235])=(.+)$/);
        assert.ok(match, "--tag requires THREAD=TAG");
        assert.ok(!options.selections.has(match[1]), `duplicate producer selection: ${match[1]}`);
        options.selections.set(match[1], match[2]);
      } else options[argument.slice(2)] = value;
    } else throw new Error(`unknown argument: ${argument}`);
  }
  assert.match(options.epoch || "", /^[1-9][0-9]*$/, "--epoch is required");
  assert.ok(options.observation, "--observation is required");
  assert.ok(!isAbsolute(options.observation) && !options.observation.split(/[\\/]/).includes(".."), "--observation must be repository-relative");
  options.repo = resolve(options.repo);
  return options;
}

function main(argv) {
  const options = parseArgs(argv);
  const path = resolve(options.repo, `_docs/ralph/prototype-integration/epochs/epoch-${options.epoch}.json`);
  const epoch = JSON.parse(readFileSync(path, "utf8"));
  const observationContent = readFileSync(resolve(options.repo, options.observation), "utf8");
  const observation = JSON.parse(observationContent);
  const manifest = JSON.parse(readFileSync(resolve(options.repo, options.manifest), "utf8"));
  validateObservationBinding(observationContent, options.observation, manifest, { repo: options.repo });
  const plan = planFrozenEpoch(epoch, options.selections, { observation, resolveTag: (tag) => resolveAnnotatedTag(options.repo, options.remote, tag) });
  process.stdout.write(`${JSON.stringify(plan, null, 2)}\n`);
}

module.exports = { parseArgs, planFrozenEpoch, resolveAnnotatedTag, validateObservationBinding };

if (require.main === module) {
  try { main(process.argv.slice(2)); }
  catch (error) { console.error(`prototype intake freeze plan: invalid: ${error.message}`); process.exitCode = 1; }
}