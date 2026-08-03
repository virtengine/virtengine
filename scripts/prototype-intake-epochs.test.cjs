"use strict";

const assert = require("assert").strict;
const { currentEpoch, requireCurrentEpoch, validateEpochSequence } = require("./prototype-intake-epochs.cjs");

function epoch(number, status) {
  return {
    file: `epoch-${number}.json`,
    number,
    document: {
      intake_epoch: number,
      status,
      base_tag: `checkpoint/prototype-integration/epoch-${number}-base`,
      planning_sha: "a".repeat(40),
      opens_at: `2000-01-0${number}T00:00:00Z`,
      announcement_cutoff: `2000-01-0${number}T23:59:59Z`,
    },
  };
}

const tests = [
  ["accepts closed epoch 1 followed by open epoch 2", () => {
    const epochs = validateEpochSequence([epoch(1, "closed"), epoch(2, "open")]);
    assert.equal(currentEpoch(epochs).number, 2);
  }],
  ["accepts one current epoch", () => assert.equal(currentEpoch([epoch(1, "closed")]).number, 1)],
  ["rejects an empty epoch history", () => assert.throws(() => validateEpochSequence([]), /at least one/)],
  ["rejects an epoch sequence gap", () => assert.throws(() => validateEpochSequence([epoch(1, "closed"), epoch(3, "open")]), /not contiguous/)],
  ["rejects a history starting at epoch 2", () => assert.throws(() => validateEpochSequence([epoch(2, "open")]), /not contiguous/)],
  ["rejects a filename and body mismatch", () => {
    const candidate = epoch(1, "open");
    candidate.document.intake_epoch = 2;
    assert.throws(() => validateEpochSequence([candidate]), /body mismatch/);
  }],
  ["rejects a non-closed predecessor", () => assert.throws(() => validateEpochSequence([epoch(1, "frozen"), epoch(2, "open")]), /must be closed/)],
  ["rejects a stale requested epoch", () => assert.throws(() => requireCurrentEpoch([epoch(1, "closed"), epoch(2, "open")], 1), /current epoch is 2/)],
  ["returns the requested current epoch", () => assert.equal(requireCurrentEpoch([epoch(1, "closed"), epoch(2, "open")], 2).intake_epoch, 2)],
  ["rejects an epoch that overlaps its predecessor", () => {
    const epochs = [epoch(1, "closed"), epoch(2, "open")];
    epochs[1].document.opens_at = "2000-01-01T12:00:00Z";
    assert.throws(() => validateEpochSequence(epochs), /overlaps its predecessor/);
  }],
  ["rejects a changed planning SHA", () => {
    const epochs = [epoch(1, "closed"), epoch(2, "open")];
    epochs[1].document.planning_sha = "b".repeat(40);
    assert.throws(() => validateEpochSequence(epochs), /planning SHA changed/);
  }],
  ["rejects a mismatched epoch base tag", () => {
    const epochs = [epoch(1, "closed"), epoch(2, "open")];
    epochs[1].document.base_tag = "checkpoint/prototype-integration/epoch-1-base";
    assert.throws(() => validateEpochSequence(epochs), /base tag mismatch/);
  }],
];

for (const [name, run] of tests) {
  run();
  console.log(`ok - ${name}`);
}