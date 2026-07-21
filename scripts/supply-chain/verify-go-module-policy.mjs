// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

import { execFileSync } from "node:child_process";
import { readFile } from "node:fs/promises";
import { dirname, relative, resolve, sep } from "node:path";
import { fileURLToPath } from "node:url";

const repositoryRoot = resolve(dirname(fileURLToPath(import.meta.url)), "../..");
const policyPath = resolve(repositoryRoot, process.argv[2] ?? "scripts/supply-chain/go-module-policy.json");
const policy = JSON.parse(await readFile(policyPath, "utf8"));

function normalizePath(path) {
  return path.split(sep).join("/");
}

function replacementText(replacement) {
  const oldVersion = replacement.Old.Version ? ` ${replacement.Old.Version}` : "";
  const newVersion = replacement.New.Version ? ` ${replacement.New.Version}` : "";
  return `${replacement.Old.Path}${oldVersion} => ${replacement.New.Path}${newVersion}`;
}

const failures = [];
if (!Array.isArray(policy.license?.allowedFamilies) || policy.license.allowedFamilies.length === 0) {
  failures.push("license.allowedFamilies must be a non-empty array");
}
if (!Array.isArray(policy.license?.deniedTokens) || policy.license.deniedTokens.length === 0) {
  failures.push("license.deniedTokens must be a non-empty array");
}
for (const [modulePath, allowedEntries] of Object.entries(policy.replaces)) {
  const absoluteModule = resolve(repositoryRoot, modulePath);
  const moduleJSON = JSON.parse(execFileSync("go", ["mod", "edit", "-json"], {
    cwd: dirname(absoluteModule),
    encoding: "utf8",
    env: { ...process.env, GOWORK: "off" },
  }));
  const actual = (moduleJSON.Replace ?? []).map(replacementText).sort();
  const allowed = [...allowedEntries].sort();
  if (JSON.stringify(actual) !== JSON.stringify(allowed)) {
    failures.push(`${modulePath}: replace directives differ from the reviewed allowlist\nexpected: ${allowed.join("\n  ")}\nactual: ${actual.join("\n  ")}`);
  }
}

if (failures.length) {
  throw new Error(failures.join("\n"));
}

process.stdout.write(`Go replace policy passed (${normalizePath(relative(repositoryRoot, policyPath))}); shared license policy defines ${policy.license.allowedFamilies.length} allowed families and ${policy.license.deniedTokens.length} denied tokens\n`);