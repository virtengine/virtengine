#!/usr/bin/env node

"use strict";

const assert = require("assert").strict;
const { readFileSync } = require("fs");
const { resolve } = require("path");

const root = resolve(__dirname, "..");

function validateLaunchers(shellContent, powershellContent) {
  const required = /go test -count=1 -tags=e2e\.integration -v \.\/tests\/integration\/\.\.\./;
  assert.match(shellContent, required, "localnet.sh test must run the canonical tagged integration package set");
  assert.match(powershellContent, required, "localnet.ps1 test must run the canonical tagged integration package set");
  assert.doesNotMatch(shellContent, /test-runner go test -v \.\/tests\/integration\/\.\.\./, "localnet.sh retains an untagged integration path");
  assert.doesNotMatch(powershellContent, /test-runner go test -v \.\/tests\/integration\/\.\.\./, "localnet.ps1 retains an untagged integration path");
  return true;
}

module.exports = { validateLaunchers };

if (require.main === module) {
  validateLaunchers(readFileSync(resolve(root, "scripts/localnet.sh"), "utf8"), readFileSync(resolve(root, "scripts/localnet.ps1"), "utf8"));
  console.log("localnet integration launchers: valid and tagged");
}