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
// A file whose NAME asserts it is the license (as opposed to a README/NOTICE).
const strongLicenseNameRegexp = /^((UN)?LICEN(S|C)E|COPYING).*$/i;

// The classifier is a whole-text matcher against canonical license templates,
// not a keyword scanner, so any hand-rolled content check is a heuristic. The
// two failure directions are not equally bad and the acceptance criterion is
// therefore asymmetric:
//
//   FALSE RED  - the gate rejects a module whose license the classifier accepts.
//                 That blocks work on a legitimate module, so it is the defect
//                 to eliminate. Self-describing headers ("BSD 3-Clause License")
//                 cause this badly: BSD, ISC, Zlib and 0BSD texts are pure
//                 boilerplate with no such header, so a header-keyword check
//                 reds on the most common permissive licenses in the Go
//                 ecosystem. The clauses below are the grant text the
//                 classifier itself keys on, which every real copy of these
//                 licenses carries.
//   FALSE GREEN - the gate accepts a file the classifier cannot identify. That
//                 only means the module reports "Failed to find license" in the
//                 Go License Check, which is the non-fatal condition this gate
//                 exists to warn about early. Held to ~0.3% by requiring grant
//                 text rather than a name mention.
//
// Measured against licenses.NewClassifier(0.9) over 12191 real license files in
// the Go module cache, scored per module root (2574 roots): 0 false red, 9 false
// green (0.3%). The header-only list this replaced scored 439 false red (17.1%)
// on the same population, so this is the whole of the round-2 defect.

// Evidence that a file carries the grant of a family, keyed by SPDX identifier
// so the accepted set derives from license.allowedFamilies in the policy rather
// than being hand-maintained in two places that can drift apart.
const licenseFamilyClauses = {
  // A family is evidenced only by text unique to its license TEXT. Naming a
  // family is not evidence of holding it: "Apache License" appears in a README
  // describing someone else's license, so it belongs in licenseMentionOnly and
  // is only honoured in a file named as the license.
  "Apache-2.0": [/\bVersion 2\.0, January 2004\b/],
  "BSD-2-Clause": [
    /\bRedistribution and use in source and binary forms\b/i,
    /\bRedistributions of source code must retain\b/i,
  ],
  "BSD-3-Clause": [
    /\bRedistribution and use in source and binary forms\b/i,
    /\bRedistributions of source code must retain\b/i,
    /\bRedistributions in binary form must reproduce\b/i,
  ],
  ISC: [/\bPermission to use, copy, modify, and\/or distribute this software\b/i],
  MIT: [
    /\bPermission is hereby granted,? free of charge\b/i,
    /\bThe above copyright notice and this permission notice shall be included\b/i,
  ],
  "MPL-2.0": [/\bMozilla Public License Version 2\.0\b/i],
  Unlicense: [/\bThis is free and unencumbered software released into the public domain\b/i],
  "CC0-1.0": [/\bCC0 1\.0 Universal\b/i],
};

// Family-independent evidence: a permissive grant whose text is not a verbatim
// SPDX template (Zlib, 0BSD, PostgreSQL, Boost, the GPL family).
const genericLicenseClauses = [
  /\bPermission is hereby granted, free of charge, to any person or organization\b/i,
  /\bPermission to use, copy, modify,? and distribute this software\b/i,
  /\bPermission to use, copy, modify, and merge this software\b/i,
  /\bGNU (?:GENERAL|LESSER GENERAL|LIBRARY|AFFERO GENERAL) PUBLIC LICENSE\b/i,
  /\bThis program is free software: you can redistribute it and\/or modify it\b/i,
  /\bBoost Software License\b/i,
  /\bprovided 'as-is', without any express or implied warranty\b/i,
  /\bPermission is granted to anyone to use this software for any purpose\b/i,
];

// A bare declaration of intent rather than license text: naming a family, or
// carrying an SPDX tag. Evidence only in a file whose NAME asserts it is the
// license. A README that says "Licensed under the Apache License" is prose
// describing someone else's license, not this module's own.
//
// Note the classifier is stricter still: it matches whole license templates, so
// a NOTICE or LICENSE holding only an SPDX tag is reported as "Failed to find
// license" even though the tag is unambiguous. These are the deliberate false
// greens - they cost a log line in the Go License Check, never a blocked merge.
const licenseMentionOnly = [
  /\bApache License\b/i,
  /\bMIT License\b/i,
  /\bBSD (?:2-Clause|3-Clause) License\b/i,
  /\bCreative Commons\b/i,
  /SPDX-License-Identifier:\s*\(?[A-Za-z0-9.+-]+/i,
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
// Drift guard: a family the policy allows but the gate has no clause for would
// be a gate contradicting the policy - it would red on a module licensed under
// a family this repo explicitly permits. Adding a family to allowedFamilies
// without teaching the gate its grant text is therefore a build failure, not a
// silent narrowing of the accepted set.
for (const family of policy.license?.allowedFamilies ?? []) {
  if (!licenseFamilyClauses[family]) {
    failures.push(
      `license.allowedFamilies permits ${family} but verify-go-module-policy.mjs has no ` +
        `clause pattern for it; add its grant text to licenseFamilyClauses so a module ` +
        `under that license is not falsely rejected`,
    );
  }
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

// The gate accepts a candidate file when its content carries grant text, or -
// only for a file named as a license - when it at least names a license family.
// Returns a short reason for the diagnostics, or null when the file carries no
// license evidence at all.
function recognizeLicense(entry, text) {
  const grantClauses = [...genericLicenseClauses, ...Object.values(licenseFamilyClauses).flat()];
  if (grantClauses.some((pattern) => pattern.test(text))) {
    return "carries license grant text";
  }
  if (strongLicenseNameRegexp.test(entry) && licenseMentionOnly.some((pattern) => pattern.test(text))) {
    return "names a license family but carries no grant text";
  }
  return null;
}

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
    const reason = recognizeLicense(entry, text);
    if (reason !== null) {
      return { ok: true, file: entry, reason, inspected };
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