"use strict";

const assert = require("assert").strict;
const { validateLaunchers } = require("./validate-localnet-integration-launchers.cjs");

const tagged = "test-runner go test -count=1 -tags=e2e.integration -v ./tests/integration/...";
const tests = [
  ["accepts matching tagged shell and PowerShell launchers", () => assert.equal(validateLaunchers(tagged, tagged), true)],
  ["rejects an untagged shell launcher", () => assert.throws(() => validateLaunchers("test-runner go test -v ./tests/integration/...", tagged), /localnet\.sh test/)],
  ["rejects an untagged PowerShell launcher", () => assert.throws(() => validateLaunchers(tagged, "test-runner go test -v ./tests/integration/..."), /localnet\.ps1 test/)],
  ["rejects cached integration execution", () => assert.throws(() => validateLaunchers("test-runner go test -tags=e2e.integration -v ./tests/integration/...", tagged), /localnet\.sh test/)],
];

for (const [name, run] of tests) { run(); console.log(`ok - ${name}`); }