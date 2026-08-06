"use strict";

const assert = require("assert").strict;
const { createObservation, parseArguments, parseRemoteTagListing, validateTagObservation } = require("./observe-prototype-intake-tags.cjs");

const annotated = [
  `${"a".repeat(40)}\trefs/tags/checkpoint/prototype-t1/t1-09`,
  `${"b".repeat(40)}\trefs/tags/checkpoint/prototype-t1/t1-09^{}`,
  `${"c".repeat(40)}\trefs/tags/checkpoint/prototype-t3/t3-x08`,
  `${"d".repeat(40)}\trefs/tags/checkpoint/prototype-t3/t3-x08^{}`,
  `${"e".repeat(40)}\trefs/tags/checkpoint/prototype-t5/t5-18`,
].join("\n");

const tests = [
  ["rejects duplicate or incomplete observer arguments", () => {
    assert.throws(() => parseArguments(["--epoch", "1", "--epoch", "2"]), /duplicate argument/);
    assert.throws(() => parseArguments(["--epoch", "--repo", "."]), /requires a value/);
    assert.throws(() => parseArguments(["--epoch", "1", "--repo"]), /requires a value/);
  }],
  ["retains only annotated intake-format tags", () => { const tags = parseRemoteTagListing(annotated); assert.deepEqual(tags, [{ thread: "T1", tag: "checkpoint/prototype-t1/t1-09", tag_object: "a".repeat(40), target: "b".repeat(40) }]); }],
  ["creates an in-window observation", () => { const value = createObservation({ intake_epoch: 1, opens_at: "2000-01-01T00:00:00Z", announcement_cutoff: "2000-01-02T00:00:00Z" }, annotated, { now: Date.parse("2000-01-01T12:00:00Z"), remote: "origin" }); assert.equal(value.observed_at, "2000-01-01T12:00:00.000Z"); assert.equal(value.tags.length, 1); }],
  ["rejects pre-open observation", () => assert.throws(() => createObservation({ intake_epoch: 1, opens_at: "2000-01-01T12:00:00Z", announcement_cutoff: "2000-01-02T00:00:00Z" }, annotated, { now: Date.parse("2000-01-01T11:59:59Z") }), /after epoch opens_at/)],
  ["rejects post-cutoff observation", () => assert.throws(() => createObservation({ intake_epoch: 1, opens_at: "2000-01-01T00:00:00Z", announcement_cutoff: "2000-01-02T00:00:00Z" }, annotated, { now: Date.parse("2000-01-02T00:00:01Z") }), /before cutoff/)],
  ["rejects malformed remote SHA", () => assert.throws(() => parseRemoteTagListing(`HEAD\trefs/tags/checkpoint/prototype-t1/t1-09`), /invalid SHA/)],
  ["rejects duplicate observed tags", () => { const epoch = { intake_epoch: 1, opens_at: "2000-01-01T00:00:00Z", announcement_cutoff: "2000-01-02T00:00:00Z" }; const value = createObservation(epoch, annotated, { now: Date.parse("2000-01-01T12:00:00Z"), remote: "origin" }); value.tags.push(structuredClone(value.tags[0])); assert.throws(() => validateTagObservation(value, epoch), /unique/); }],
];

for (const [name, run] of tests) { run(); console.log(`ok - ${name}`); }