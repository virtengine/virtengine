"use strict";

const assert = require("assert").strict;
const { readFileSync } = require("fs");
const { resolve } = require("path");
const { validateFundRouteInventory } = require("./validate-fund-route-inventory.cjs");

const root = resolve(__dirname, "..");
const inventory = JSON.parse(readFileSync(resolve(root, "_docs/ralph/prototype-integration/fund-route-inventory.json"), "utf8"));
const clone = () => structuredClone(inventory);
const validate = (value, options = {}) => validateFundRouteInventory(value, { rootDir: root, discoveredMoverFiles: options.discoveredMoverFiles || value.mover_files, requireReady: options.requireReady });

const tests = [
  ["accepts the dependency-blocked inventory", () => assert.doesNotThrow(() => validate(clone()))],
  ["rejects an unclassified mover file", () => { const value = clone(); assert.throws(() => validate(value, { discoveredMoverFiles: [...value.mover_files, "x/new/keeper/move.go"].sort() }), /discovery drifted/); }],
  ["rejects a missing mover file", () => { const value = clone(); value.mover_files.pop(); assert.throws(() => validate(value, { discoveredMoverFiles: inventory.mover_files }), /discovery drifted/); }],
  ["rejects duplicate route IDs", () => { const value = clone(); value.routes.push(structuredClone(value.routes[0])); assert.throws(() => validate(value), /duplicate route ID/); }],
  ["rejects an incomplete category set", () => { const value = clone(); value.routes = value.routes.filter((route) => route.category !== "payout"); assert.throws(() => validate(value), /categories are incomplete/); }],
  ["rejects an unassigned discovered mover", () => { const value = clone(); value.routes.forEach((route) => { route.paths = route.paths.filter((path) => path !== "x/bme/keeper/escrow.go"); }); value.routes.find((route) => route.id === "bme-escrow").paths = ["x/bme/module.go"]; assert.throws(() => validate(value), /not fully assigned/); }],
  ["rejects undeclared blockers", () => { const value = clone(); value.routes[0].blockers.push("unknown"); assert.throws(() => validate(value), /undeclared blocker/); }],
  ["rejects path traversal", () => { const value = clone(); value.routes[0].paths = ["../outside.go"]; assert.throws(() => validate(value), /escapes repository/); }],
  ["rejects premature completion", () => { const value = clone(); value.status = "complete"; assert.throws(() => validate(value), /status disagrees/); }],
  ["require-ready rejects the blocked inventory", () => assert.throws(() => validate(clone(), { requireReady: true }), /not ready/)],
  ["accepts a fully wired future transition", () => {
    const value = clone();
    value.status = "complete";
    value.accepted_fund_authorization_checkpoint = "a".repeat(40);
    value.blockers = [];
    value.completion.allowed = true;
    value.routes.forEach((route) => { route.fund_authorization = "wired"; route.atomicity = "verified"; route.blockers = []; });
    assert.doesNotThrow(() => validate(value, { requireReady: true }));
  }],
];

for (const [name, run] of tests) {
  run();
  console.log(`ok - ${name}`);
}