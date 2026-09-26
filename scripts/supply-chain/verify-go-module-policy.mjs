// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

import { execFileSync } from "node:child_process";
import { readdir, readFile } from "node:fs/promises";
import { dirname, relative, resolve, sep } from "node:path";
import { fileURLToPath } from "node:url";

const repositoryRoot = resolve(dirname(fileURLToPath(import.meta.url)), "../..");
const policyPath = resolve(repositoryRoot, process.argv[2] ?? "scripts/supply-chain/go-module-policy.json");
const policy = JSON.parse(await readFile(policyPath, "utf8"));

// go-licenses resolves a package's license by walking up from the package
// directory and stopping at the enclosing module root (licenses.Find's rootDir).
// A nested module therefore has to carry its own LICENSE: the repository-root
// and parent-directory LICENSE files are outside that walk, and every package
// in the module is reported as "Failed to find license". Guard that invariant
// here so adding a nested Go module cannot silently break the Go License Check.
const licenseRegexp = /^((UN)?LICEN(S|C)E|COPYING|README|NOTICE).*$/i;
const trackedGoModules = execFileSync("git", ["ls-files", "**/go.mod", "go.mod"], {
  cwd: repositoryRoot,
  encoding: "utf8",
})
  .split("\n")
  .filter(Boolean);

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

// Every Go module root must expose a license file go-licenses can classify.
// A nested module without one makes the whole Go License Check report
// "Failed to find license" for each of its packages.
for (const goModPath of trackedGoModules) {
  const moduleDir = dirname(goModPath);
  const entries = await readdir(resolve(repositoryRoot, moduleDir));
  if (!entries.some((entry) => licenseRegexp.test(entry))) {
    failures.push(
      `${moduleDir}: Go module root has no license file matching ${licenseRegexp} (checked: ${entries.length} entries). ` +
        "go-licenses stops its upward walk at the module root, so a nested module needs its own LICENSE file.",
    );
  }
}

if (failures.length) {
  throw new Error(failures.join("\n"));
}

process.stdout.write(`Go replace policy passed (${normalizePath(relative(repositoryRoot, policyPath))}); shared license policy defines ${policy.license.allowedFamilies.length} allowed families and ${policy.license.deniedTokens.length} denied tokens\n`);