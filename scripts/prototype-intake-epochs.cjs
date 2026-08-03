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
  for (let index = 0; index < epochs.length; index += 1) {
    const expected = index + 1;
    const epoch = epochs[index];
    assert.equal(epoch.number, expected, `intake epoch sequence is not contiguous at epoch ${expected}`);
    assert.equal(epoch.file, `epoch-${expected}.json`, `intake epoch filename mismatch at epoch ${expected}`);
    assert.equal(epoch.document.intake_epoch, expected, `intake epoch body mismatch at epoch ${expected}`);
    if (index < epochs.length - 1) assert.equal(epoch.document.status, "closed", `predecessor epoch ${expected} must be closed`);
  }
  return epochs;
}

function currentEpoch(epochs) {
  return validateEpochSequence(epochs).at(-1);
}

module.exports = { currentEpoch, discoverEpochs, validateEpochSequence };