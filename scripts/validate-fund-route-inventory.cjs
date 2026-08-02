#!/usr/bin/env node

"use strict";

const assert = require("assert").strict;
const { existsSync, readFileSync, readdirSync, statSync } = require("fs");
const { isAbsolute, join, relative, resolve } = require("path");

const root = resolve(__dirname, "..");
const inventoryPath = resolve(root, "_docs/ralph/prototype-integration/fund-route-inventory.json");
const requiredCategories = ["bank_send", "escrow_refund", "escrow_release", "final_settlement", "issuance_mint", "payout", "privileged_treasury", "recovery", "reward", "withdrawal"];
const primitivePattern = /\.(?:MintCoins|BurnCoins|SendCoinsFromAccountToModule|SendCoinsFromModuleToAccount|SendCoinsFromModuleToModule)\s*\(/;

function exactKeys(value, keys, label) {
  assert.deepEqual(Object.keys(value).sort(), [...keys].sort(), `${label} has unknown or missing fields`);
}

function discoverMoverFiles(rootDir) {
  const files = [];
  const pending = [resolve(rootDir, "x")];
  while (pending.length > 0) {
    const directory = pending.pop();
    for (const entry of readdirSync(directory)) {
      const path = join(directory, entry);
      if (statSync(path).isDirectory()) pending.push(path);
      else if (entry.endsWith(".go") && !entry.endsWith("_test.go") && primitivePattern.test(readFileSync(path, "utf8"))) files.push(relative(rootDir, path).replaceAll("\\", "/"));
    }
  }
  return files.sort();
}

function validateFundRouteInventory(inventory, options = {}) {
  const rootDir = options.rootDir || root;
  const discoveredMoverFiles = options.discoveredMoverFiles || discoverMoverFiles(rootDir);
  exactKeys(inventory, ["accepted_fund_authorization_checkpoint", "blockers", "checkpoint", "completion", "movement_primitives", "mover_files", "routes", "schema_version", "status"], "fund route inventory");
  assert.equal(inventory.schema_version, "virtengine.prototype.fund-route-inventory/v1");
  assert.equal(inventory.checkpoint, "T4-16A");
  assert.ok(["dependency_blocked", "complete"].includes(inventory.status));
  exactKeys(inventory.accepted_fund_authorization_checkpoint, ["payload_sha", "status", "tag", "thread"], "accepted fund authorization checkpoint");
  assert.equal(inventory.accepted_fund_authorization_checkpoint.thread, "T5");
  assert.ok(["accepted", "unavailable"].includes(inventory.accepted_fund_authorization_checkpoint.status));
  if (inventory.accepted_fund_authorization_checkpoint.status === "unavailable") {
    assert.equal(inventory.accepted_fund_authorization_checkpoint.tag, null);
    assert.equal(inventory.accepted_fund_authorization_checkpoint.payload_sha, null);
  } else {
    assert.match(inventory.accepted_fund_authorization_checkpoint.tag, /^checkpoint\/prototype-t5\/[a-z0-9-]+$/);
    assert.match(inventory.accepted_fund_authorization_checkpoint.payload_sha, /^[a-f0-9]{40}$/);
    assert.equal(options.verifyAcceptedCheckpoint?.(inventory.accepted_fund_authorization_checkpoint), true, "accepted T5 checkpoint verification failed");
  }
  assert.deepEqual(inventory.mover_files, [...inventory.mover_files].sort(), "mover files must be sorted");
  assert.deepEqual(inventory.mover_files, discoveredMoverFiles, "production fund mover discovery drifted");
  assert.equal(new Set(inventory.mover_files).size, inventory.mover_files.length, "duplicate mover file");
  assert.equal(new Set(inventory.blockers).size, inventory.blockers.length, "duplicate blocker");
  assert.ok(Array.isArray(inventory.routes) && inventory.routes.length > 0);
  assert.equal(new Set(inventory.routes.map((route) => route.id)).size, inventory.routes.length, "duplicate route ID");
  assert.deepEqual([...new Set(inventory.routes.map((route) => route.category))].sort(), requiredCategories, "required fund route categories are incomplete");

  for (const route of inventory.routes) {
    exactKeys(route, ["atomicity", "blockers", "category", "fund_authorization", "id", "paths", "trigger"], `route ${route.id || "unknown"}`);
    assert.ok(requiredCategories.includes(route.category), `unknown route category ${route.category}`);
    assert.ok(route.id && route.trigger && route.paths.length > 0);
    assert.ok(["missing", "wired"].includes(route.fund_authorization));
    assert.ok(["known_non_atomic", "stub_unwired", "unverified", "verified"].includes(route.atomicity));
    for (const blocker of route.blockers) assert.ok(inventory.blockers.includes(blocker), `${route.id} references undeclared blocker ${blocker}`);
    if (route.fund_authorization === "wired") assert.ok(!route.blockers.includes("route-authorization-unwired"));
    for (const path of route.paths) {
      assert.ok(!isAbsolute(path) && !path.split(/[\\/]/).includes(".."), `route path escapes repository: ${path}`);
      const absolutePath = resolve(rootDir, path);
      assert.ok(existsSync(absolutePath) && statSync(absolutePath).isFile(), `route source missing: ${path}`);
      assert.notEqual(path, "cmd/inference-sidecar/manifest_test2.go", "user-owned sidecar file is outside fund-route evidence");
    }
  }

  const assignedMoverFiles = new Set(inventory.routes.flatMap((route) => route.paths).filter((path) => inventory.mover_files.includes(path)));
  assert.deepEqual([...assignedMoverFiles].sort(), inventory.mover_files, "discovered mover files are not fully assigned to routes");

  const ready = inventory.accepted_fund_authorization_checkpoint.status === "accepted"
    && inventory.routes.every((route) => route.fund_authorization === "wired" && route.atomicity === "verified" && route.blockers.length === 0)
    && inventory.blockers.length === 0;
  assert.equal(inventory.completion.allowed, ready, "fund route completion is premature");
  assert.equal(inventory.status === "complete", ready, "fund route status disagrees with readiness");
  if (options.requireReady) assert.equal(ready, true, "fund route authorization is not ready");
  return true;
}

module.exports = { discoverMoverFiles, validateFundRouteInventory };

if (require.main === module) {
  const inventory = JSON.parse(readFileSync(inventoryPath, "utf8"));
  validateFundRouteInventory(inventory, { requireReady: process.argv.includes("--require-ready") });
  console.log(`fund route inventory: ${inventory.status} (${inventory.routes.length} routes, ${inventory.mover_files.length} mover files)`);
}