#!/usr/bin/env node

"use strict";

const assert = require("assert").strict;
const { readdirSync, readFileSync } = require("fs");
const { join } = require("path");

const epochFilePattern = /^epoch-([1-9][0-9]*)\.json$/;

function discoverEpochs(directory) {
  const epochs = readdirSync(directory)
    .map((name) => ({ name, match: name.match(epochFilePattern) }))
    .filter((entry) => entry.match)
    .map((entry) => ({
      file: entry.name,
      number: Number(entry.match[1]),
      document: JSON.parse(readFileSync(join(directory, entry.name), "utf8")),
    }))
    .sort((left, right) => left.number - right.number);
  return validateEpochSequence(epochs);
}

function validateEpochSequence(epochs) {
  assert.ok(Array.isArray(epochs) && epochs.length > 0, "at least one intake epoch is required");
  const planningSha = epochs[0].document.planning_sha;
  for (let index = 0; index < epochs.length; index += 1) {
    const expected = index + 1;
    const epoch = epochs[index];
    assert.equal(epoch.number, expected, `intake epoch sequence is not contiguous at epoch ${expected}`);
    assert.equal(epoch.file, `epoch-${expected}.json`, `intake epoch filename mismatch at epoch ${expected}`);
    assert.equal(epoch.document.intake_epoch, expected, `intake epoch body mismatch at epoch ${expected}`);
    assert.equal(epoch.document.base_tag, `checkpoint/prototype-integration/epoch-${expected}-base`, `intake epoch base tag mismatch at epoch ${expected}`);
    assert.equal(epoch.document.planning_sha, planningSha, `intake epoch planning SHA changed at epoch ${expected}`);
    assert.ok(Number.isFinite(Date.parse(epoch.document.opens_at)), `intake epoch ${expected} opens_at is invalid`);
    assert.ok(Number.isFinite(Date.parse(epoch.document.announcement_cutoff)), `intake epoch ${expected} cutoff is invalid`);
    assert.ok(Date.parse(epoch.document.opens_at) < Date.parse(epoch.document.announcement_cutoff), `intake epoch ${expected} window is invalid`);
    if (index < epochs.length - 1) assert.equal(epoch.document.status, "closed", `predecessor epoch ${expected} must be closed`);
    if (index > 0) {
      assert.ok(Date.parse(epoch.document.opens_at) >= Date.parse(epochs[index - 1].document.announcement_cutoff), `intake epoch ${expected} overlaps its predecessor window`);
    }
  }
  return epochs;
}

function currentEpoch(epochs) {
  return validateEpochSequence(epochs).at(-1);
}

function requireCurrentEpoch(epochs, requestedEpoch) {
  const current = currentEpoch(epochs);
  assert.equal(Number(requestedEpoch), current.number, `requested epoch ${requestedEpoch} is stale; current epoch is ${current.number}`);
  return current.document;
}

function validateEpochBase(epoch, resolvedTag) {
  assert.equal(resolvedTag.type, "tag", `epoch ${epoch.intake_epoch} base tag must be annotated`);
  assert.equal(resolvedTag.declared_name, epoch.base_tag, `epoch ${epoch.intake_epoch} base tag object declares another name`);
  assert.match(resolvedTag.target || "", /^[a-f0-9]{40}$/, `epoch ${epoch.intake_epoch} base tag target is invalid`);
  assert.equal(resolvedTag.target, epoch.base_sha, `epoch ${epoch.intake_epoch} base tag does not target base_sha`);
  return true;
}

function validateAnnotatedTagName(expected, content) {
  const declared = content.match(/^tag (.+)$/m)?.[1];
  assert.equal(declared, expected, `annotated tag object for ${expected} declares ${declared || "no tag name"}`);
  return declared;
}

module.exports = { currentEpoch, discoverEpochs, requireCurrentEpoch, validateAnnotatedTagName, validateEpochBase, validateEpochSequence };