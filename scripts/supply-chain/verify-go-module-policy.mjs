// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

import { execFileSync } from "node:child_process";
import { readdir, readFile, stat } from "node:fs/promises";
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
//
// go-licenses matches candidate files by name first, then requires the file's
// CONTENT to be identifiable by its license classifier (licenses/find.go:
// `if _, _, err := classifier.Identify(path); err != nil { return false }`).
// A prose README.md matches the name regexp but classifies as no license, so a
// name-only check would be a false green: the gate would pass while every
// package in that module still reports "Failed to find license".
const licenseCandidateRegexp = /^((UN)?LICEN(S|C)E|COPYING|README|NOTICE).*$/i;
// Signals that a file carries recognisable license text. This is deliberately a
// conservative subset: the gate must not pass a file go-licenses cannot classify,
// and it must not fail a genuine license whose wording differs from our sample.
const licenseTextPatterns = [
  /SPDX-License-Identifier:\s*\(?[A-Za-z0-9.+-]+/i,
  /\bApache License\b/i,
  /\bMIT License\b/i,
  /\bPermission is hereby granted,? free of charge\b/i,
  /\bGNU (?:GENERAL|LESSER GENERAL|LIBRARY) PUBLIC LICENSE\b/i,
  /\bMozilla Public License\b/i,
  /\bBSD (?:2-Clause|3-Clause) License\b/i,
  /\bBoost Software License\b/i,
];
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

// Every Go module root must expose a license file go-licenses can both find by
// name and classify by content. A nested module without one makes the whole Go
// License Check report "Failed to find license" for each of its packages.
const MAX_LICENSE_SCAN_BYTES = 1024 * 1024;

async function hasClassifiableLicense(moduleDir, entries) {
  const inspected = [];
  for (const entry of entries) {
    if (!licenseCandidateRegexp.test(entry)) {
      continue;
    }
    const absolute = resolve(repositoryRoot, moduleDir, entry);
    const info = await stat(absolute);
    if (!info.isFile() || info.size > MAX_LICENSE_SCAN_BYTES) {
      // A directory or an oversized file cannot be the license text go-licenses
      // would classify; skip rather than load it into memory.
      inspected.push(`${entry} (name matches, not a readable license text file)`);
      continue;
    }
    const text = await readFile(absolute, "utf8");
    const recognizable = licenseTextPatterns.some((pattern) => pattern.test(text));
    if (recognizable) {
      return { ok: true, file: entry, inspected };
    }
    // Candidate by name only: a prose README/NOTICE is exactly the false green
    // this gate must not produce, so report it rather than accept it.
    inspected.push(`${entry} (name matches, no recognisable license text)`);
  }
  return { ok: false, inspected };
}

for (const goModPath of trackedGoModules) {
  const moduleDir = dirname(goModPath);
  const entries = await readdir(resolve(repositoryRoot, moduleDir));
  const result = await hasClassifiableLicense(moduleDir, entries);
  if (!result.ok) {
    failures.push(
      `${moduleDir}: Go module root has no license file go-licenses can identify ` +
        `(name matches ${licenseCandidateRegexp} and content must carry recognisable license text; ` +
        `checked: ${entries.length} entries` +
        `${result.inspected.length > 0 ? `, rejected: ${result.inspected.join("; ")}` : ""}). ` +
        "go-licenses stops its upward walk at the module root, so a nested module needs its own LICENSE file.",
    );
  }
}

if (failures.length) {
  throw new Error(failures.join("\n"));
}

process.stdout.write(`Go replace policy passed (${normalizePath(relative(repositoryRoot, policyPath))}); shared license policy defines ${policy.license.allowedFamilies.length} allowed families and ${policy.license.deniedTokens.length} denied tokens\n`);