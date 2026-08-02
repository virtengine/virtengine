#!/usr/bin/env node

"use strict";

const assert = require("assert").strict;
const { readFileSync, writeFileSync } = require("fs");
const { resolve } = require("path");
const { spawnSync } = require("child_process");

const threads = ["T1", "T2", "T3", "T5"];

function validateFreezeTransition(current, proposed, now = Date.now()) {
  assert.deepEqual(Object.keys(proposed).sort(), Object.keys(current).sort(), "freeze plan changes epoch fields");
  assert.equal(current.status, "open", "current epoch must be open");
  assert.equal(proposed.status, "frozen", "freeze plan must set status to frozen");
  assert.ok(now > Date.parse(current.announcement_cutoff), "epoch announcement cutoff has not elapsed");
  for (const key of Object.keys(current).filter((key) => !["status", "producers"].includes(key))) {
    assert.deepEqual(proposed[key], current[key], `freeze plan changes ${key}`);
  }
  assert.deepEqual(current.producers.map((producer) => producer.thread), threads, "current producer roster is invalid");
  assert.deepEqual(proposed.producers.map((producer) => producer.thread), threads, "freeze plan producer roster is invalid");
  for (let index = 0; index < threads.length; index += 1) {
    const before = current.producers[index];
    const after = proposed.producers[index];
    assert.deepEqual(before, { thread: threads[index], status: "unannounced", tag: null, decision: null }, `${before.thread} current state is not open`);
    const announced = after.status === "announced" && typeof after.tag === "string" && after.decision === null;
    const frozenOut = after.status === "unannounced" && after.tag === null && after.decision === "frozen-out";
    assert.ok(announced || frozenOut, `${after.thread} freeze decision is invalid`);
  }
  return true;
}

function parseArgs(argv) {
  const options = { epoch: null, plan: null, repo: resolve(__dirname, "..") };
  for (let index = 0; index < argv.length; index += 2) {
    const argument = argv[index];
    const value = argv[index + 1];
    assert.ok(["--epoch", "--plan", "--repo"].includes(argument) && value, `invalid argument: ${argument || "missing"}`);
    options[argument.slice(2)] = value;
  }
  assert.match(options.epoch || "", /^[1-9][0-9]*$/, "--epoch is required");
  assert.ok(options.plan, "--plan is required");
  options.repo = resolve(options.repo);
  return options;
}

function main(argv) {
  const options = parseArgs(argv);
  const status = spawnSync("git", ["status", "--porcelain"], { cwd: options.repo, encoding: "utf8" });
  assert.equal(status.status, 0, "unable to inspect T4 worktree");
  assert.equal(status.stdout.trim(), "", "T4 worktree must be clean before freezing intake");
  const epochPath = resolve(options.repo, `_docs/ralph/prototype-integration/epochs/epoch-${options.epoch}.json`);
  const current = JSON.parse(readFileSync(epochPath, "utf8"));
  const proposed = JSON.parse(readFileSync(resolve(options.plan), "utf8"));
  assert.equal(current.intake_epoch, Number(options.epoch), "epoch number mismatch");
  validateFreezeTransition(current, proposed);
  writeFileSync(epochPath, `${JSON.stringify(proposed, null, 2)}\n`, "utf8");
  process.stdout.write(`prototype intake epoch ${options.epoch} frozen\n`);
}

module.exports = { parseArgs, validateFreezeTransition };

if (require.main === module) {
  try { main(process.argv.slice(2)); }
  catch (error) { console.error(`prototype intake freeze apply: invalid: ${error.message}`); process.exitCode = 1; }
}